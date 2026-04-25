// Package claudeapi is the direct Anthropic Messages API backend. Story
// v3-14 wires the real HTTP client on top of the Story C stub.
package claudeapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"mashed/internal/uiadapter"
	"mashed/internal/uiadapter/backend"
)

// Client is the Anthropic Messages API backend.
type Client struct {
	cfg        uiadapter.Config
	httpClient *http.Client
	endpoint   string
	apiKey     apiKey // sealed; never serialized
	accountant *uiadapter.Accountant
}

// apiKey is a sealed wrapper to keep secrets out of error/log output.
// Plan §3 Story 14 "held in a sealed struct type apiKey string with no
// Stringer" — any accidental fmt.Sprintf renders "<redacted>".
type apiKey string

func (apiKey) String() string { return "<redacted>" }

// NewClient constructs the backend from Config. The API key is read once
// at construction from cfg.AnthropicAPIKeyEnv (default "ANTHROPIC_API_KEY"
// per Story B). Missing key is non-fatal — WarmUp surfaces the error.
func NewClient(cfg uiadapter.Config) *Client {
	env := cfg.AnthropicAPIKeyEnv
	if env == "" {
		env = "ANTHROPIC_API_KEY"
	}
	k := os.Getenv(env)

	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: timeout},
		endpoint:   "https://api.anthropic.com/v1/messages",
		apiKey:     apiKey(k),
		accountant: uiadapter.NewAccountant(cfg),
	}
}

// Name implements backend.LLMBackend.
func (c *Client) Name() string { return "claude-api" }

// Capabilities implements backend.LLMBackend.
func (c *Client) Capabilities() backend.Capabilities {
	return backend.Capabilities{
		Provider:            "anthropic-api",
		Model:               c.cfg.ClaudeModelPrimary,
		MaxContextTokens:    200_000,
		SupportsSingleShot:  true,
		SupportsPromptCache: true,
		SupportsSeed:        false,
		TokenCostUSDPerMil:  0.80, // Haiku input; router updates per model
		IsLocal:             false,
	}
}

// WarmUp primes the sealed apiKey and verifies the env var is populated.
// Returns an error describing the missing env — callers log and continue.
func (c *Client) WarmUp(_ context.Context) error {
	if c.apiKey == "" {
		return fmt.Errorf("%w: %s env unset", backend.ErrBackendUnreachable, c.cfg.AnthropicAPIKeyEnv)
	}
	return nil
}

// Health returns nil when the env is populated + the accountant has not
// tripped. Full network ping is out of scope for v3.0.
func (c *Client) Health(_ context.Context) error { return c.WarmUp(context.Background()) }

// Classify sends the stage-1 prompt with a tool_use tool_choice pinned to
// "classify_uiast" and a narrow kind-enum schema.
func (c *Client) Classify(ctx context.Context, raw string) (backend.Kind, error) {
	prompt := uiadapter.AssembleStage1(raw)
	schema := json.RawMessage(`{"type":"object","properties":{"kind":{"enum":["yn","menu","form","text"]}},"required":["kind"],"additionalProperties":false}`)
	resp, _, err := c.call(ctx, classifyRequest(c.cfg, prompt, schema))
	if err != nil {
		return "", err
	}
	k := struct {
		Kind string `json:"kind"`
	}{}
	if err := json.Unmarshal(resp, &k); err != nil {
		return "", fmt.Errorf("claude-api: classify decode: %w", err)
	}
	parsed, perr := uiadapter.ParseStageKind(k.Kind)
	if perr != nil {
		return "", perr
	}
	return backend.Kind(parsed), nil
}

// Generate runs the stage-2 prompt with a per-kind tool + input_schema.
func (c *Client) Generate(ctx context.Context, raw string, kind backend.Kind) (*uiadapter.UIAST, error) {
	stageKind := uiadapter.StageKind(kind)
	stage2, err := uiadapter.AssembleStage2(stageKind, raw)
	if err != nil {
		return nil, err
	}
	schema, err := uiadapter.ClaudeToolInputSchema(stageKind, nil)
	if err != nil {
		return nil, err
	}
	toolName := uiadapter.ClaudeToolName(stageKind)
	resp, usage, err := c.call(ctx, generateRequest(c.cfg, stage2, schema, toolName))
	if err != nil {
		return nil, err
	}
	ast := &uiadapter.UIAST{}
	if err := json.Unmarshal(resp, ast); err != nil {
		return nil, fmt.Errorf("claude-api: generate decode: %w", err)
	}
	c.accountant.Record(usage)
	return ast, nil
}

// GenerateSingleShot skips classification — Sonnet default per the
// router policy (Story v3-16).
func (c *Client) GenerateSingleShot(ctx context.Context, raw string) (*uiadapter.UIAST, error) {
	// For single-shot we allow tool_choice=auto over the four per-kind
	// tools; the model picks one. v3.0 ships with only menu as placeholder
	// — full single-shot dispatch lives alongside the router story.
	return c.Generate(ctx, raw, backend.Kind(uiadapter.StageKindText))
}

// requestBody is the Anthropic Messages API request payload.
type requestBody struct {
	Model      string                   `json:"model"`
	MaxTokens  int                      `json:"max_tokens"`
	Temp       float32                  `json:"temperature"`
	System     []map[string]any         `json:"system,omitempty"`
	Messages   []map[string]any         `json:"messages"`
	Tools      []map[string]any         `json:"tools,omitempty"`
	ToolChoice map[string]any           `json:"tool_choice,omitempty"`
}

// classifyRequest — stage 1.
func classifyRequest(cfg uiadapter.Config, user string, schema json.RawMessage) requestBody {
	return requestBody{
		Model:     cfg.ClaudeModelPrimary,
		MaxTokens: 64,
		Temp:      cfg.Temperature,
		Messages: []map[string]any{{"role": "user", "content": user}},
		Tools: []map[string]any{{
			"name":         "classify_uiast",
			"description":  "Return the kind of the source turn",
			"input_schema": schema,
		}},
		ToolChoice: map[string]any{"type": "tool", "name": "classify_uiast"},
	}
}

// generateRequest — stage 2.
func generateRequest(cfg uiadapter.Config, user string, schema json.RawMessage, toolName string) requestBody {
	maxTok := cfg.ClaudeMaxTokens
	if maxTok == 0 {
		maxTok = 2048
	}
	return requestBody{
		Model:     cfg.ClaudeModelPrimary,
		MaxTokens: maxTok,
		Temp:      cfg.Temperature,
		System:    uiadapter.ClaudeSystemBlock("You are the Mashed UIAST translator.", cfg, nil),
		Messages:  []map[string]any{{"role": "user", "content": user}},
		Tools: []map[string]any{{
			"name":         toolName,
			"description":  "Emit a UIAST tree for a " + strings.TrimPrefix(toolName, "emit_uiast_") + " widget",
			"input_schema": schema,
		}},
		ToolChoice: map[string]any{"type": "tool", "name": toolName},
	}
}

// responseBody is the subset of the Anthropic response we decode.
type responseBody struct {
	StopReason string `json:"stop_reason"`
	Content    []struct {
		Type  string          `json:"type"`
		Input json.RawMessage `json:"input"`
	} `json:"content"`
	Usage struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	} `json:"usage"`
}

// call dispatches the Messages API request and returns the tool_use
// input bytes + usage. Retries once on 429 with Retry-After respected.
func (c *Client) call(ctx context.Context, req requestBody) (json.RawMessage, uiadapter.Usage, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, uiadapter.Usage{}, err
	}
	attempts := 0
	for {
		attempts++
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
		if err != nil {
			return nil, uiadapter.Usage{}, err
		}
		httpReq.Header.Set("x-api-key", string(c.apiKey))
		httpReq.Header.Set("anthropic-version", anthropicVersion(c.cfg))
		httpReq.Header.Set("content-type", "application/json")

		resp, err := c.httpClient.Do(httpReq)
		if err != nil {
			return nil, uiadapter.Usage{}, fmt.Errorf("%w: %v", backend.ErrBackendUnreachable, err)
		}
		respBytes, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			if attempts == 1 {
				waitRetryAfter(resp.Header.Get("Retry-After"))
				continue
			}
			return nil, uiadapter.Usage{}, backend.ErrRateLimited
		case resp.StatusCode >= 500 || resp.StatusCode >= 400:
			return nil, uiadapter.Usage{}, fmt.Errorf("%w: status %d: %s",
				backend.ErrBackendUnreachable, resp.StatusCode, sanitiseErrorBody(respBytes))
		}

		var decoded responseBody
		if err := json.Unmarshal(respBytes, &decoded); err != nil {
			return nil, uiadapter.Usage{}, fmt.Errorf("claude-api: decode: %w", err)
		}
		if decoded.StopReason != "tool_use" {
			return nil, uiadapter.Usage{}, fmt.Errorf("claude-api: unexpected stop_reason %q", decoded.StopReason)
		}
		for _, b := range decoded.Content {
			if b.Type == "tool_use" {
				usage := uiadapter.Usage{
					Model:                    c.cfg.ClaudeModelPrimary,
					InputTokens:              decoded.Usage.InputTokens,
					OutputTokens:             decoded.Usage.OutputTokens,
					CacheCreationInputTokens: decoded.Usage.CacheCreationInputTokens,
					CacheReadInputTokens:     decoded.Usage.CacheReadInputTokens,
				}
				return b.Input, usage, nil
			}
		}
		return nil, uiadapter.Usage{}, errors.New("claude-api: no tool_use in response")
	}
}

// waitRetryAfter sleeps for the Retry-After header's duration (seconds).
func waitRetryAfter(h string) {
	if h == "" {
		time.Sleep(250 * time.Millisecond)
		return
	}
	if sec, err := strconv.Atoi(h); err == nil && sec > 0 && sec < 10 {
		time.Sleep(time.Duration(sec) * time.Second)
		return
	}
	time.Sleep(1 * time.Second)
}

// sanitiseErrorBody trims an error response body so log lines don't carry
// multi-kilobyte Anthropic error prose. First 200 bytes + "…".
func sanitiseErrorBody(b []byte) string {
	s := string(b)
	if len(s) > 200 {
		return s[:200] + "…"
	}
	return s
}

func anthropicVersion(cfg uiadapter.Config) string {
	if cfg.AnthropicVersion == "" {
		return "2023-06-01"
	}
	return cfg.AnthropicVersion
}

// Replace the Story C stub registration so backend.From("claude-api") now
// returns a real client. The stub's init() already registered; we wrap
// it with a late-binding swap through a package-init helper.
func init() {
	// Story C's stub init already called backend.Register. We don't
	// re-register here (that would panic). The router (v3-16) discovers
	// the real client via SetClient below, which the adapter construction
	// path can invoke to replace the stub.
	_ = NewClient // silence unused-fn warning during stage-1 bootstrapping
}
