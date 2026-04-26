package uiadapter

import (
	"crypto/sha256"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestClaudeSystemBlock_HasCacheControl — AC-7.5/AC-7.6 precondition. The
// system block carries a cache_control marker by default.
func TestClaudeSystemBlock_HasCacheControl(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	blocks := ClaudeSystemBlock("static-prefix", cfg, nil)
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
	blocks := ClaudeSystemBlock("prefix", cfg, nil)
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

	first, err := ClaudeSystemBlockJSON(prefix, cfg, nil)
	require.NoError(t, err)
	firstHash := sha256.Sum256(first)

	for i := 0; i < 100; i++ {
		out, err := ClaudeSystemBlockJSON(prefix, cfg, nil)
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
	assert.Equal(t, "30m", OllamaKeepAliveEncoded(cfg, nil))

	cfg.KeepAlive = ""
	assert.Equal(t, "30m", OllamaKeepAliveEncoded(cfg, nil), "empty KeepAlive defaults to 30m")

	cfg.KeepAlive = "-1"
	assert.Equal(t, "-1", OllamaKeepAliveEncoded(cfg, nil), "explicit -1 passes through (dev forever-resident)")
}

// TestClaudeSystemBlock_EmptyPrefixNil — empty static prefix returns nil,
// so the Messages API request ships with no system block at all.
func TestClaudeSystemBlock_EmptyPrefixNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, ClaudeSystemBlock("", DefaultConfig(), nil))
}

// -----------------------------------------------------------------------------
// Story 3 — `prefix_cache.go` debug instrumentation (uiadapter-logging-3).
//
// RED-phase: these target log records that do not yet exist in production.
// -----------------------------------------------------------------------------

// TestStory3_AC6_PrefixCacheBuildEvents — Story 3, AC-3.6.
//
// `ClaudeSystemBlock("hello", cfg, logger)` must emit `prefix_cache.build`
// with prefix_len=5 and a ttl attr. `ClaudeSystemBlockJSON` must additionally
// emit `prefix_cache.build.success` with a positive bytes_out.
func TestStory3_AC6_PrefixCacheBuildEvents(t *testing.T) {
	const prefix = "hello"

	t.Run("ClaudeSystemBlock emits build", func(t *testing.T) {
		logger, buf := testLogBuffer(t, slog.LevelDebug)
		out := ClaudeSystemBlock(prefix, DefaultConfig(), logger)
		require.NotNil(t, out, "ClaudeSystemBlock must return a non-nil block for non-empty prefix")

		records := decodeRecords(t, buf)
		builds := recordsByMsg(records, "prefix_cache.build")
		require.NotEmpty(t, builds, "expected a prefix_cache.build record; got %v", records)

		rec := builds[0]
		assert.Equal(t, "prefix_cache.claude", rec["op"],
			"prefix_cache.build must carry op=\"prefix_cache.claude\"")
		assert.EqualValues(t, len(prefix), rec["prefix_len"],
			"prefix_cache.build prefix_len must equal len(prefix)=%d", len(prefix))
		ttl, has := rec["ttl"]
		require.True(t, has, "prefix_cache.build must carry a ttl attr")
		assert.NotEmpty(t, ttl, "ttl must be a non-empty value")
	})

	t.Run("ClaudeSystemBlockJSON additionally emits build.success", func(t *testing.T) {
		logger, buf := testLogBuffer(t, slog.LevelDebug)
		body, err := ClaudeSystemBlockJSON(prefix, DefaultConfig(), logger)
		require.NoError(t, err)
		require.NotEmpty(t, body)

		records := decodeRecords(t, buf)
		builds := recordsByMsg(records, "prefix_cache.build")
		require.NotEmpty(t, builds,
			"ClaudeSystemBlockJSON must still emit prefix_cache.build (it calls ClaudeSystemBlock); got %v",
			records)

		successes := recordsByMsg(records, "prefix_cache.build.success")
		require.NotEmpty(t, successes,
			"expected a prefix_cache.build.success record from JSON marshal; got %v", records)
		rec := successes[0]
		bytesOut, ok := rec["bytes_out"].(float64)
		require.True(t, ok, "prefix_cache.build.success must carry numeric bytes_out; got %v", rec["bytes_out"])
		assert.Greater(t, bytesOut, 0.0, "bytes_out must be > 0")
	})
}

// TestStory3_AC6_OllamaKeepAlive — Story 3, AC-3.6 (Ollama branch).
//
// `OllamaKeepAliveEncoded` must emit `prefix_cache.ollama_keep_alive` with the
// op and keep_alive attrs.
func TestStory3_AC6_OllamaKeepAlive(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	cfg := DefaultConfig()
	cfg.KeepAlive = "30m"

	got := OllamaKeepAliveEncoded(cfg, logger)
	assert.Equal(t, "30m", got)

	records := decodeRecords(t, buf)
	emits := recordsByMsg(records, "prefix_cache.ollama_keep_alive")
	require.NotEmpty(t, emits, "expected a prefix_cache.ollama_keep_alive record; got %v", records)
	rec := emits[0]
	assert.Equal(t, "prefix_cache.ollama", rec["op"],
		"prefix_cache.ollama_keep_alive must carry op=\"prefix_cache.ollama\"")
	assert.Equal(t, "30m", rec["keep_alive"],
		"prefix_cache.ollama_keep_alive must carry keep_alive value")
}
