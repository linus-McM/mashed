package uiadapter

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestAllowlist_WarnsOnUnvetted — AC-12.1. Model outside the Ollama
// allowlist emits a Warn slog line.
func TestAllowlist_WarnsOnUnvetted(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	cfg := DefaultConfig()
	cfg.Model = "gemma4:latest"
	warned := CheckModelAllowlist(cfg, logger)
	assert.True(t, warned, "unvetted Ollama model warns")
	assert.Contains(t, buf.String(), `"level":"WARN"`)
	assert.Contains(t, buf.String(), `"model":"gemma4:latest"`)
}

// TestAllowlist_AllowsKnownModels — AC-12.2. Allowlisted models stay
// silent.
func TestAllowlist_AllowsKnownModels(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	cfg := DefaultConfig() // Model=gemma3:4b, ClaudeModelPrimary=claude-haiku-4-5
	warned := CheckModelAllowlist(cfg, logger)
	assert.False(t, warned)
	assert.Empty(t, buf.String())
}

// TestAllowlist_OverrideFlag — AC-12.3. AllowUnvettedModels=true bypasses
// the check.
func TestAllowlist_OverrideFlag(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	cfg := DefaultConfig()
	cfg.Model = "gemma4:latest"
	cfg.AllowUnvettedModels = true
	warned := CheckModelAllowlist(cfg, logger)
	assert.False(t, warned, "override flag suppresses the warning")
	assert.Empty(t, buf.String())
}

// TestAllowlist_WarnsOnUnvettedClaude — parallel coverage for Claude.
func TestAllowlist_WarnsOnUnvettedClaude(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn}))
	cfg := DefaultConfig()
	cfg.ClaudeModelPrimary = "claude-foo-9"
	warned := CheckModelAllowlist(cfg, logger)
	assert.True(t, warned)
	assert.Contains(t, buf.String(), `"model":"claude-foo-9"`)
}
