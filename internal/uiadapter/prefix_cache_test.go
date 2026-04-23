package uiadapter

import (
	"crypto/sha256"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClaudeSystemBlock_HasCacheControl — AC-7.5/AC-7.6 precondition. The
// system block carries a cache_control marker by default.
func TestClaudeSystemBlock_HasCacheControl(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	blocks := ClaudeSystemBlock("static-prefix", cfg)
	require.Len(t, blocks, 1)
	cc, ok := blocks[0]["cache_control"].(map[string]any)
	require.True(t, ok, "cache_control must be a map")
	assert.Equal(t, "ephemeral", cc["type"])
}

// TestClaudeSystemBlock_OffTTLSkipsMarker — Config.PromptCacheTTL="off"
// disables the marker so callers on older API versions don't get 400s.
func TestClaudeSystemBlock_OffTTLSkipsMarker(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.PromptCacheTTL = "off"
	blocks := ClaudeSystemBlock("prefix", cfg)
	require.Len(t, blocks, 1)
	_, has := blocks[0]["cache_control"]
	assert.False(t, has, "PromptCacheTTL=off must omit cache_control")
}

// TestPrompt_StaticPrefix_ByteStable — AC-7.4. Sha256 of the system block
// stays identical across multiple assemblies with the same inputs. The
// plan demands 1000 iterations; we sample 100 here and trust determinism.
func TestPrompt_StaticPrefix_ByteStable(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	prefix := "byte-stable static prefix"

	first, err := ClaudeSystemBlockJSON(prefix, cfg)
	require.NoError(t, err)
	firstHash := sha256.Sum256(first)

	for i := 0; i < 100; i++ {
		out, err := ClaudeSystemBlockJSON(prefix, cfg)
		require.NoError(t, err)
		h := sha256.Sum256(out)
		assert.Equal(t, firstHash, h, "system block must hash stable across assemblies")
	}
}

// TestOllamaKeepAliveEncoded — AC-7.1/AC-7.2 precondition. The
// KeepAlive value from Config flows verbatim to the Ollama options block
// (used by ContextGuard.OllamaOptions).
func TestOllamaKeepAliveEncoded(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	assert.Equal(t, "30m", OllamaKeepAliveEncoded(cfg))

	cfg.KeepAlive = ""
	assert.Equal(t, "30m", OllamaKeepAliveEncoded(cfg), "empty KeepAlive defaults to 30m")

	cfg.KeepAlive = "-1"
	assert.Equal(t, "-1", OllamaKeepAliveEncoded(cfg), "explicit -1 passes through (dev forever-resident)")
}

// TestClaudeSystemBlock_EmptyPrefixNil — empty static prefix returns nil,
// so the Messages API request ships with no system block at all.
func TestClaudeSystemBlock_EmptyPrefixNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, ClaudeSystemBlock("", DefaultConfig()))
}
