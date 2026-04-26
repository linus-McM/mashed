package uiadapter

import (
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContextGuard_TruncatesLongCapture_Ollama — AC-3.1. 20K-char input
// truncated to ≤8K tokens on Ollama path with sentinel in place.
func TestContextGuard_TruncatesLongCapture_Ollama(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.NumCtx = 8192
	g := NewContextGuard(cfg, nil)

	raw := strings.Repeat("long narrative line with several words in it.\n", 800) // ≈ 40K bytes
	fitted, truncated := g.ApplyOllama(raw)
	assert.True(t, truncated)
	// Token budget: 8192-1024 reserve = 7168 tokens ≈ 28672 bytes.
	assert.LessOrEqual(t, len(fitted), 7168*4+200, "fitted ≤ token-budget-bytes")
	assert.Contains(t, fitted, "elided", "visible sentinel replaces dropped middle")
}

// TestContextGuard_NoTruncationUnderBudget — short captures pass through.
func TestContextGuard_NoTruncationUnderBudget(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	g := NewContextGuard(cfg, nil)
	raw := "short prompt"
	fitted, truncated := g.ApplyOllama(raw)
	assert.False(t, truncated)
	assert.Equal(t, raw, fitted)
}

// TestContextGuard_OllamaOptions — AC-3.2. `num_ctx` + `keep_alive` appear
// in the options block.
func TestContextGuard_OllamaOptions(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	g := NewContextGuard(cfg, nil)
	opts := g.OllamaOptions()
	assert.Equal(t, 8192, opts["num_ctx"], "num_ctx from Config")
	assert.Equal(t, "30m", opts["keep_alive"], "keep_alive from Config")
}

// TestContextGuard_RefusesLongCapture_Claude — AC-3.3. A 200K-token raw on
// the Claude path returns ErrContextOverflow.
func TestContextGuard_RefusesLongCapture_Claude(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	g := NewContextGuard(cfg, nil)

	// 200K tokens = 800K bytes approximately; model context is 200K tokens.
	// ClaudeMaxTokens=2048 reserve → headroom ≈ 197952; exceed it.
	raw := strings.Repeat("x", 1_000_000) // 250K tokens
	_, err := g.ApplyClaude(raw, 200_000)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrContextOverflow),
		"must wrap ErrContextOverflow so fallback chain (Story v3-11) can branch")
}

// TestContextGuard_AcceptsShortClaude — short captures pass.
func TestContextGuard_AcceptsShortClaude(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	g := NewContextGuard(cfg, nil)
	raw := "a short prompt"
	out, err := g.ApplyClaude(raw, 200_000)
	require.NoError(t, err)
	assert.Equal(t, raw, out)
}

// TestContextGuard_ZeroConfigDefaults — zero-valued NumCtx still produces a
// working guard (Story B defaults applied).
func TestContextGuard_ZeroConfigDefaults(t *testing.T) {
	t.Parallel()
	g := NewContextGuard(Config{}, nil)
	raw := strings.Repeat("x", 100_000) // ~25K tokens
	fitted, truncated := g.ApplyOllama(raw)
	assert.True(t, truncated)
	assert.NotEmpty(t, fitted)
}

// TestStory5_AC3_ApplyOllamaTruncated — Story 5 AC-5.3.
// Over-budget input emits contextguard.ollama.start at entry and a
// contextguard.ollama.truncated record carrying truncated=true,
// original_tokens > fitted_tokens, reserve_tokens >= 0. No raw payload bytes
// leak into the records.
func TestStory5_AC3_ApplyOllamaTruncated(t *testing.T) {
	t.Parallel()
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	cfg := DefaultConfig()
	cfg.NumCtx = 1024
	g := NewContextGuard(cfg, logger)

	const sentinel = "OVERBUDGETPAYLOAD"
	raw := strings.Repeat(sentinel, 600) // ~10K bytes, well above 1024 budget

	fitted, truncated := g.ApplyOllama(raw)
	require.True(t, truncated)
	require.NotEmpty(t, fitted)

	records := decodeRecords(t, buf)
	starts := recordsByMsg(records, "contextguard.ollama.start")
	require.Len(t, starts, 1)
	assert.Equal(t, "contextguard.ollama", starts[0]["op"])
	assert.EqualValues(t, len(raw), starts[0]["bytes_in"])

	truncs := recordsByMsg(records, "contextguard.ollama.truncated")
	require.Len(t, truncs, 1)
	assert.Equal(t, "contextguard.ollama", truncs[0]["op"])
	assert.EqualValues(t, true, truncs[0]["truncated"])

	orig, ok := truncs[0]["original_tokens"].(float64)
	require.True(t, ok, "original_tokens must be a JSON number")
	fitted2, ok := truncs[0]["fitted_tokens"].(float64)
	require.True(t, ok, "fitted_tokens must be a JSON number")
	reserve, ok := truncs[0]["reserve_tokens"].(float64)
	require.True(t, ok, "reserve_tokens must be a JSON number")
	assert.Greater(t, orig, fitted2, "original_tokens must exceed fitted_tokens on truncation")
	assert.GreaterOrEqual(t, reserve, 0.0, "reserve_tokens must be non-negative")

	// §14 — no raw payload substring leaks.
	assert.NotContains(t, buf.String(), sentinel)
}

// TestStory5_AC3_ApplyOllamaUnderBudget — Story 5 AC-5.3 happy path.
// Under-budget input still emits start + truncated records, with truncated=false
// and original_tokens == fitted_tokens.
func TestStory5_AC3_ApplyOllamaUnderBudget(t *testing.T) {
	t.Parallel()
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	cfg := DefaultConfig()
	g := NewContextGuard(cfg, logger)

	raw := "short prompt"
	fitted, truncated := g.ApplyOllama(raw)
	require.False(t, truncated)
	require.Equal(t, raw, fitted)

	records := decodeRecords(t, buf)
	require.Len(t, recordsByMsg(records, "contextguard.ollama.start"), 1)

	truncs := recordsByMsg(records, "contextguard.ollama.truncated")
	require.Len(t, truncs, 1)
	assert.EqualValues(t, false, truncs[0]["truncated"])
	orig, _ := truncs[0]["original_tokens"].(float64)
	fitted2, _ := truncs[0]["fitted_tokens"].(float64)
	assert.Equal(t, orig, fitted2,
		"under-budget: original_tokens must equal fitted_tokens (no truncation occurred)")

	assert.NotContains(t, buf.String(), raw, "raw must never appear verbatim in records")
}

// TestStory5_AC3_ApplyClaudeTruncated — Story 5 AC-5.3 Claude branch.
// Over-budget input on the Claude path emits contextguard.claude.start and
// contextguard.claude.truncated records analogous to the Ollama branch.
func TestStory5_AC3_ApplyClaudeTruncated(t *testing.T) {
	t.Parallel()
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	cfg := DefaultConfig()
	g := NewContextGuard(cfg, logger)

	const sentinel = "BUDGETBUSTERPAYLOAD"
	raw := strings.Repeat(sentinel, 60_000) // exceeds the Claude headroom
	_, err := g.ApplyClaude(raw, 200_000)
	require.Error(t, err, "over-budget Claude must return ErrContextOverflow")
	assert.True(t, errors.Is(err, ErrContextOverflow))

	records := decodeRecords(t, buf)
	starts := recordsByMsg(records, "contextguard.claude.start")
	require.Len(t, starts, 1)
	assert.Equal(t, "contextguard.claude", starts[0]["op"])
	assert.EqualValues(t, len(raw), starts[0]["bytes_in"])
	// max_model_context attr provides operator visibility into the limit applied.
	assert.EqualValues(t, 200_000, starts[0]["max_model_context"])

	truncs := recordsByMsg(records, "contextguard.claude.truncated")
	require.Len(t, truncs, 1)
	assert.Equal(t, "contextguard.claude", truncs[0]["op"])
	assert.EqualValues(t, true, truncs[0]["truncated"])
	orig, _ := truncs[0]["original_tokens"].(float64)
	fitted, _ := truncs[0]["fitted_tokens"].(float64)
	assert.Greater(t, orig, fitted)

	// §14 sanitize: no raw byte slice in the buffer.
	assert.NotContains(t, buf.String(), sentinel)
}
