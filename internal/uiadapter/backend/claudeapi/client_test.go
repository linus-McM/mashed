package claudeapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/uiadapter"
	"mashed/internal/uiadapter/backend"
)

// TestClaudeAPI_ToolUseRoundtrip — AC-14.1. A stubbed httptest server
// returns a tool_use block containing a valid UIAST; Client.Generate
// decodes it directly — no intermediate JSON parsing.
func TestClaudeAPI_ToolUseRoundtrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r
		_, _ = io.ReadAll(r.Body)
		resp := map[string]any{
			"stop_reason": "tool_use",
			"content": []map[string]any{{
				"type":  "tool_use",
				"name":  "emit_uiast_text",
				"input": map[string]any{"version": "1", "turn_summary": "stub", "nodes": []any{}, "fallback_answer_shape": "free"},
			}},
			"usage": map[string]any{"input_tokens": 10, "output_tokens": 5},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := uiadapter.DefaultConfig()
	cfg.AnthropicAPIKeyEnv = "CLAUDE_TEST_KEY"
	t.Setenv("CLAUDE_TEST_KEY", "sk-test")
	c := NewClient(cfg)
	c.endpoint = srv.URL

	ast, err := c.Generate(context.Background(), "anything", backend.Kind(uiadapter.StageKindText))
	require.NoError(t, err)
	require.NotNil(t, ast)
	assert.Equal(t, "1", ast.Version)
}

// TestClaudeAPI_RetryAfter429 — AC-14.2. A 429 with Retry-After: 1 is
// respected; second-attempt success returns a UIAST.
func TestClaudeAPI_RetryAfter429(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		resp := map[string]any{
			"stop_reason": "tool_use",
			"content": []map[string]any{{
				"type": "tool_use", "name": "emit_uiast_text",
				"input": map[string]any{"version": "1", "nodes": []any{}, "fallback_answer_shape": "free"},
			}},
			"usage": map[string]any{"input_tokens": 1, "output_tokens": 1},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	cfg := uiadapter.DefaultConfig()
	cfg.AnthropicAPIKeyEnv = "CLAUDE_TEST_KEY"
	t.Setenv("CLAUDE_TEST_KEY", "sk-test")
	c := NewClient(cfg)
	c.endpoint = srv.URL
	c.httpClient.Timeout = 10 * 1_000_000_000 // 10s for sleep

	ast, err := c.Generate(context.Background(), "anything", backend.Kind(uiadapter.StageKindText))
	require.NoError(t, err)
	require.NotNil(t, ast)
	assert.Equal(t, 2, hits, "retry after first 429")
}

// TestClaudeAPI_KeyNotLogged — AC-14.5. The sealed apiKey's Stringer
// returns "<redacted>" so fmt/log can never leak the secret.
func TestClaudeAPI_KeyNotLogged(t *testing.T) {
	t.Parallel()
	k := apiKey("sk-super-secret")
	s := k.String()
	assert.Equal(t, "<redacted>", s)
	assert.NotContains(t, s, "secret")
}

// TestClaudeAPI_WarmUpMissingKeyErrors — construction with an empty env
// makes WarmUp report ErrBackendUnreachable.
func TestClaudeAPI_WarmUpMissingKeyErrors(t *testing.T) {
	t.Parallel()
	cfg := uiadapter.DefaultConfig()
	cfg.AnthropicAPIKeyEnv = "NONEXISTENT_KEY_FOR_TEST"
	c := NewClient(cfg)
	err := c.WarmUp(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NONEXISTENT_KEY_FOR_TEST")
}

// TestClaudeAPI_PromptCacheMarkers — AC-14.4 partial. The generate wire
// payload carries cache_control on the system block.
func TestClaudeAPI_PromptCacheMarkers(t *testing.T) {
	t.Parallel()
	cfg := uiadapter.DefaultConfig()
	req := generateRequest(cfg, "user content", nil, "emit_uiast_text")
	body, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Contains(t, string(body), "cache_control")
	assert.Contains(t, string(body), "ephemeral")
}

// TestClaudeAPI_AnthropicVersionHeader — pinned version header.
func TestClaudeAPI_AnthropicVersionHeader(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("anthropic-version")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"stop_reason":"tool_use","content":[{"type":"tool_use","name":"emit_uiast_text","input":{"version":"1","nodes":[],"fallback_answer_shape":"free"}}],"usage":{}}`))
	}))
	defer srv.Close()

	cfg := uiadapter.DefaultConfig()
	cfg.AnthropicAPIKeyEnv = "CLAUDE_TEST_KEY"
	t.Setenv("CLAUDE_TEST_KEY", "sk-test")
	c := NewClient(cfg)
	c.endpoint = srv.URL
	_, _ = c.Generate(context.Background(), "anything", backend.Kind(uiadapter.StageKindText))
	assert.Equal(t, "2023-06-01", got)
}

// TestClaudeAPI_StopReasonNotToolUseErrors — guard against an accidental
// text response; the contract is tool_use always.
func TestClaudeAPI_StopReasonNotToolUseErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"stop_reason":"end_turn","content":[]}`))
	}))
	defer srv.Close()
	cfg := uiadapter.DefaultConfig()
	cfg.AnthropicAPIKeyEnv = "CLAUDE_TEST_KEY"
	t.Setenv("CLAUDE_TEST_KEY", "sk-test")
	c := NewClient(cfg)
	c.endpoint = srv.URL
	_, err := c.Generate(context.Background(), "x", backend.Kind(uiadapter.StageKindText))
	require.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "stop_reason"))
}
