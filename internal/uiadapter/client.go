// Package uiadapter wraps the local Ollama HTTP API for the UI AST layer.
package uiadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"
)

// ollamaHost is the only host this package talks to. Overridable from tests
// only — production code never writes it.
var ollamaHost = "http://localhost:11434"

var (
	sharedTransport = &http.Transport{MaxIdleConns: 2}
	sharedClient    = &http.Client{Transport: sharedTransport}
)

var ErrOllamaUnreachable = errors.New("uiadapter: ollama unreachable")

// HTTPStatusError surfaces a non-2xx HTTP response from Ollama with its
// status code intact so the adapter can map it to "fallback:server:<code>".
type HTTPStatusError struct {
	StatusCode int
	Status     string
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("uiadapter: ollama returned %s", e.Status)
}

// ClientConfig is the Ollama HTTP client's construction knob. Kept distinct
// from the adapter's Config so the two concerns (HTTP timeout vs. adapter
// enablement/concurrency) never collide.
type ClientConfig struct {
	TimeoutMs int
}

type Client struct {
	timeout time.Duration
}

// NewClient adds a 500ms buffer so the caller's context deadline fires
// before http.Client.Timeout — otherwise timeouts surface as transport
// errors instead of deadline errors.
func NewClient(cfg ClientConfig) *Client {
	return &Client{timeout: time.Duration(cfg.TimeoutMs+500) * time.Millisecond}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Format   string        `json:"format"`
	Stream   bool          `json:"stream"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

type tagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// Chat returns message.content verbatim — callers parse the JSON themselves.
func (c *Client) Chat(ctx context.Context, model, system, user string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:  model,
		Format: "json",
		Stream: false,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: user},
		},
	})
	if err != nil {
		return "", fmt.Errorf("uiadapter: marshal chat body: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ollamaHost+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("uiadapter: build chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sharedClient.Do(req)
	if err != nil {
		return "", classifyTransportErr(ctx, "chat", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &HTTPStatusError{StatusCode: resp.StatusCode, Status: resp.Status}
	}

	var decoded chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("uiadapter: decode chat response: %w", err)
	}
	return decoded.Message.Content, nil
}

// ListModels returns locally-pulled model names sorted lexicographically.
// An empty list is not an error.
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ollamaHost+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("uiadapter: build tags request: %w", err)
	}

	resp, err := sharedClient.Do(req)
	if err != nil {
		return nil, classifyTransportErr(ctx, "list_models", err)
	}
	defer resp.Body.Close()

	var decoded tagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("uiadapter: decode tags response: %w", err)
	}

	names := make([]string, 0, len(decoded.Models))
	for _, m := range decoded.Models {
		names = append(names, m.Name)
	}
	sort.Strings(names)
	return names, nil
}

// classifyTransportErr preserves context errors (so callers can distinguish
// timeout from outage) and otherwise wraps as ErrOllamaUnreachable. op names
// the caller operation (e.g. "chat", "list_models") for log diagnostics.
func classifyTransportErr(ctx context.Context, op string, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return fmt.Errorf("uiadapter: %s aborted: %w", op, ctxErr)
	}
	return fmt.Errorf("%w: %v", ErrOllamaUnreachable, err)
}
