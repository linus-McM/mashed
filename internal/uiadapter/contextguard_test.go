package uiadapter

import (
	"errors"
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
