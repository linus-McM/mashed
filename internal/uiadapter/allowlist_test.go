package uiadapter

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// TestStory5_AC6_AllowlistOK — Story 5 AC-5.6.
// A vetted Ollama + Claude config emits exactly one allowlist.ok Debug record
// carrying op="allowlist.check" and model=cfg.Model.
func TestStory5_AC6_AllowlistOK(t *testing.T) {
	t.Parallel()
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	cfg := DefaultConfig()

	warned := CheckModelAllowlist(cfg, logger)
	require.False(t, warned)

	records := decodeRecords(t, buf)
	recs := recordsByMsg(records, "allowlist.ok")
	require.GreaterOrEqual(t, len(recs), 1, "vetted-OK path must emit allowlist.ok")
	assert.Equal(t, "allowlist.check", recs[0]["op"])
	assert.Equal(t, cfg.Model, recs[0]["model"])
	// Debug-level guard.
	level, _ := recs[0]["level"].(string)
	assert.Equal(t, "DEBUG", level, "allowlist.ok must be Debug level")
}

// TestStory5_AC6_AllowlistDefaultRemoved — Story 5 AC-5.6.
// Compile-time grep: the source of allowlist.go must contain zero
// `slog.Default()` references after Story 5 lands. This is the static
// linting bar specified in the AC.
func TestStory5_AC6_AllowlistDefaultRemoved(t *testing.T) {
	t.Parallel()

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller must succeed for path discovery")
	dir := filepath.Dir(thisFile)
	src, err := os.ReadFile(filepath.Join(dir, "allowlist.go"))
	require.NoError(t, err, "allowlist.go must be readable")

	count := strings.Count(string(src), "slog.Default()")
	assert.Equal(t, 0, count,
		"allowlist.go must contain zero slog.Default() references after Story 5 (found %d)", count)
}

// TestStory5_AC6_AllowlistNilLoggerStillSafe — Story 5 AC-5.6.
// Calling CheckModelAllowlist with a nil logger must still work — the
// fallback path is `nilSafeLogger` (Story 1/2), no longer slog.Default().
func TestStory5_AC6_AllowlistNilLoggerStillSafe(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()

	require.NotPanics(t, func() {
		_ = CheckModelAllowlist(cfg, nil)
	}, "nil logger must be safe via nilSafeLogger")
}
