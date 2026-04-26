// Package claudecli wraps `claude -p` as a backend. Story v3-15 adds the
// subprocess lifecycle, session reuse, and fenced-JSON parsing on top of
// the Story C stub registration.
package claudecli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"sync"

	"mashed/internal/uiadapter"
	"mashed/internal/uiadapter/backend"
)

// Client wraps the claude CLI in the uiadapter interface.
type Client struct {
	cfg       uiadapter.Config
	binary    string
	extraArgs []string

	mu        sync.Mutex
	sessionID string
}

// NewClient constructs the CLI backend from Config. Does not verify the
// binary at construction — WarmUp surfaces missing-binary errors.
func NewClient(cfg uiadapter.Config) *Client {
	bin := cfg.ClaudeCLIBinary
	if bin == "" {
		bin = "claude"
	}
	return &Client{cfg: cfg, binary: bin, extraArgs: cfg.ClaudeCLIExtraFlags}
}

// Name implements backend.LLMBackend.
func (c *Client) Name() string { return "claude-cli" }

// Capabilities implements backend.LLMBackend.
func (c *Client) Capabilities() backend.Capabilities {
	return backend.Capabilities{
		Provider:            "claude-cli",
		Model:               c.cfg.ClaudeModelPrimary,
		MaxContextTokens:    200_000,
		SupportsSingleShot:  true,
		SupportsPromptCache: true, // via session resume
		SupportsSeed:        false,
		IsLocal:             false,
	}
}

// WarmUp verifies the CLI is on PATH. Missing binary is the most common
// failure mode for this backend.
func (c *Client) WarmUp(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.binary, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: claude binary %q not executable: %v",
			backend.ErrBackendUnreachable, c.binary, err)
	}
	return nil
}

// Health re-uses WarmUp. The CLI has no cheap health endpoint.
func (c *Client) Health(ctx context.Context) error { return c.WarmUp(ctx) }

// Classify sends the stage-1 prompt through `claude -p`; parses the
// fenced JSON at end-of-turn.
func (c *Client) Classify(ctx context.Context, raw string) (backend.Kind, error) {
	user := uiadapter.AssembleStage1(raw, nil)
	out, err := c.runOneShot(ctx, "Return a single JSON object matching {\"kind\": <one of yn|menu|form|text>}.", user)
	if err != nil {
		return "", err
	}
	k := struct {
		Kind string `json:"kind"`
	}{}
	if err := json.Unmarshal([]byte(out), &k); err != nil {
		return "", fmt.Errorf("claude-cli: classify decode: %w", err)
	}
	parsed, perr := uiadapter.ParseStageKind(k.Kind, nil)
	if perr != nil {
		return "", perr
	}
	return backend.Kind(parsed), nil
}

// Generate runs stage-2 via `claude -p`; fenced-JSON contract (softer
// than Anthropic API tool_use — hence the lower parse-rate target).
func (c *Client) Generate(ctx context.Context, raw string, kind backend.Kind) (*uiadapter.UIAST, error) {
	stageKind := uiadapter.StageKind(kind)
	stage2, err := uiadapter.AssembleStage2(stageKind, raw, nil)
	if err != nil {
		return nil, err
	}
	systemPrompt, err := uiadapter.GeneratePromptFor(stageKind)
	if err != nil {
		return nil, err
	}
	out, err := c.runOneShot(ctx, systemPrompt, stage2)
	if err != nil {
		return nil, err
	}
	ast := &uiadapter.UIAST{}
	if err := json.Unmarshal([]byte(out), ast); err != nil {
		return nil, fmt.Errorf("claude-cli: generate decode: %w", err)
	}
	return ast, nil
}

// GenerateSingleShot runs without explicit classification — the single-
// shot prompt carries all four per-kind schemas.
func (c *Client) GenerateSingleShot(ctx context.Context, raw string) (*uiadapter.UIAST, error) {
	return c.Generate(ctx, raw, backend.Kind(uiadapter.StageKindText))
}

// runOneShot spawns `claude -p` with the given system + user and returns
// the parsed fenced-JSON block.
func (c *Client) runOneShot(ctx context.Context, systemPrompt, user string) (string, error) {
	args := []string{
		"-p",
		"--output-format", "stream-json",
		"--append-system-prompt", systemPrompt,
		"--disallowedTools", "*",
		"--permission-mode", "bypassPermissions",
		"--input-format", "text",
	}
	c.mu.Lock()
	if c.sessionID != "" {
		args = append(args, "--resume", c.sessionID)
	}
	c.mu.Unlock()
	args = append(args, c.extraArgs...)

	cmd := exec.CommandContext(ctx, c.binary, args...)
	cmd.Cancel = func() error {
		// Best-effort SIGTERM before CommandContext's Kill.
		if cmd.Process != nil {
			_ = cmd.Process.Signal(termSignal)
		}
		return nil
	}
	cmd.Stdin = strings.NewReader(user)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("%w: %v", backend.ErrBackendUnreachable, err)
	}

	// Parse stream-json lines; extract session_id from the init event
	// and concatenate assistant text deltas. The fenced JSON at the end
	// of the final assistant message is the output contract.
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1*1024*1024)

	var assistantText strings.Builder
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var evt map[string]any
		if err := json.Unmarshal(line, &evt); err != nil {
			continue
		}
		if sid, ok := evt["session_id"].(string); ok && sid != "" {
			c.mu.Lock()
			if c.sessionID == "" {
				c.sessionID = sid
			}
			c.mu.Unlock()
		}
		if t, _ := evt["type"].(string); t == "assistant" {
			if msg, ok := evt["message"].(map[string]any); ok {
				if content, ok := msg["content"].(string); ok {
					assistantText.WriteString(content)
				}
			}
		}
	}
	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("claude-cli: process failed: %w", err)
	}

	return extractFencedJSON(assistantText.String())
}

// extractFencedJSON pulls the largest ```json ... ``` (or ``` ... ```)
// block from the assistant output. Returns the JSON body or an error if
// no block was found.
func extractFencedJSON(s string) (string, error) {
	// Strip optional "json" language tag.
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return "", errors.New("claude-cli: empty assistant output")
	}
	// Find the last fenced block; model sometimes emits commentary first.
	idx := strings.LastIndex(trimmed, "```")
	if idx < 0 {
		// Might be bare JSON already.
		if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
			return trimmed, nil
		}
		return "", errors.New("claude-cli: no fenced JSON block found")
	}
	// Find the preceding opening fence.
	open := strings.LastIndex(trimmed[:idx], "```")
	if open < 0 {
		return "", errors.New("claude-cli: unmatched fence")
	}
	body := trimmed[open+3 : idx]
	body = strings.TrimPrefix(body, "json")
	return strings.TrimSpace(body), nil
}

func init() {
	// Story C already registered a stub. No re-registration from this
	// package — the real client is returned via the router's SetClient
	// path in v3-16.
	_ = NewClient
}
