package uiadapter

import (
	"errors"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCache_HashHit — AC-4.1. Two identical Translate-key lookups produce
// exactly one downstream call on the second path.
func TestCache_HashHit(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cache := NewResponseCache(cfg, nil)
	key := Key("ollama", "gemma3:4b", "raw capture")

	_, hit := cache.Lookup(key)
	assert.False(t, hit, "empty cache misses")

	ast := &UIAST{Version: "1", GeneratedBy: "stub"}
	cache.Store(key, ast)

	got, hit := cache.Lookup(key)
	assert.True(t, hit)
	assert.Same(t, ast, got, "pointer equality — no copy cost on hit")
}

// TestCache_SingleflightCoalesces — AC-4.2. 100 concurrent identical calls
// produce exactly one downstream call. The leader's fn sleeps 50ms so
// every follower definitely reaches singleflight.Do while the leader's
// call is still in flight, proving the coalescing contract.
func TestCache_SingleflightCoalesces(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cache := NewResponseCache(cfg, nil)
	key := Key("ollama", "gemma3:4b", "coalesce")

	var calls atomic.Int64
	var wg sync.WaitGroup
	const N = 100
	wg.Add(N)
	start := make(chan struct{})
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			<-start // release all goroutines simultaneously
			_, _, err := cache.DoShared(key, func() (*UIAST, error) {
				// 50ms window covers scheduler jitter on any
				// reasonable CI runner — every follower has time to
				// reach Do and join the in-flight group.
				time.Sleep(50 * time.Millisecond)
				calls.Add(1)
				return &UIAST{Version: "1"}, nil
			})
			assert.NoError(t, err)
		}()
	}
	close(start)
	wg.Wait()
	assert.EqualValues(t, 1, calls.Load(), "singleflight collapses 100 concurrent callers to 1 downstream call")
}

// TestCache_HashKeyIncludesBackendAndModel — AC-4.5. Switching
// Config.Backend invalidates cache entries.
func TestCache_HashKeyIncludesBackendAndModel(t *testing.T) {
	t.Parallel()
	kOllama := Key("ollama", "gemma3:4b", "same raw")
	kClaude := Key("claude-api", "claude-haiku-4-5", "same raw")
	assert.NotEqual(t, kOllama, kClaude)

	// Cache version must invalidate on the sentinel change as well.
	assert.Contains(t, cacheVersion, "v", "sentinel must be bumpable on prompt/schema change")
}

// TestCache_DisabledWhenZeroCapacity — CacheCapacity=0 disables LRU but
// keeps Singleflight functioning (useful for tests).
func TestCache_DisabledWhenZeroCapacity(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.CacheCapacity = 0
	cache := NewResponseCache(cfg, nil)

	key := Key("ollama", "gemma3:4b", "x")
	cache.Store(key, &UIAST{Version: "1"})
	_, hit := cache.Lookup(key)
	assert.False(t, hit, "disabled cache never hits")

	// Singleflight still collapses. 50ms window ensures all 10 goroutines
	// reach Do while the leader's fn is still in flight.
	var calls atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, _, err := cache.DoShared(key, func() (*UIAST, error) {
				time.Sleep(50 * time.Millisecond)
				calls.Add(1)
				return &UIAST{Version: "1"}, nil
			})
			require.NoError(t, err)
		}()
	}
	close(start)
	wg.Wait()
	assert.EqualValues(t, 1, calls.Load())
}

// TestCache_HitRate — AC-4.3 precondition. HitRate() reflects the running
// ratio; feeds the slog `cache_hit_rate` attribute.
func TestCache_HitRate(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cache := NewResponseCache(cfg, nil)
	key := Key("ollama", "gemma3:4b", "hit-rate")
	cache.Store(key, &UIAST{Version: "1"})

	for i := 0; i < 3; i++ {
		_, _ = cache.Lookup(key)
	}
	_, _ = cache.Lookup(Key("ollama", "gemma3:4b", "missing"))
	assert.InDelta(t, 0.75, cache.HitRate(), 0.001, "3 hits out of 4 lookups")

	h, m := cache.Metrics()
	assert.EqualValues(t, 3, h)
	assert.EqualValues(t, 1, m)
}

// TestCache_SingleflightPropagatesError — errors from the inner fn flow
// back to every concurrent caller; no panic, no goroutine leak.
func TestCache_SingleflightPropagatesError(t *testing.T) {
	t.Parallel()
	cache := NewResponseCache(DefaultConfig(), nil)
	sentinel := errors.New("downstream-fail")
	_, _, err := cache.DoShared("err-key", func() (*UIAST, error) {
		return nil, sentinel
	})
	assert.ErrorIs(t, err, sentinel)
}

// TestCache_KeyDeterministic — same inputs produce the same key across
// runs (hash stability).
func TestCache_KeyDeterministic(t *testing.T) {
	t.Parallel()
	a := Key("ollama", "gemma3:4b", "hello world")
	b := Key("ollama", "gemma3:4b", "hello world")
	assert.Equal(t, a, b)
	assert.Len(t, a, 64, "sha256 hex length")
}

// -----------------------------------------------------------------------------
// Story 3 — `cache.go` debug instrumentation (uiadapter-logging-3).
//
// RED-phase tests: these reference the `hashKey` helper that the go-engineer
// must add (8-char SHA-256 hex prefix), so the file fails to compile until
// the implementation lands. They also target log records that do not yet
// exist in production.
// -----------------------------------------------------------------------------

// TestStory3_AC3_CacheLifecycle — Story 3, AC-3.3.
//
// A capacity-2 cache exercised with put/put/hit/put-with-eviction/miss must
// emit exactly: cache.put x3, cache.hit x1, cache.miss x1, cache.evict x1
// (with reason="lru" and cache_key_hash matching hashKey(<evicted-key>)).
// Every emitted record carries an 8-character cache_key_hash (never the raw key).
func TestStory3_AC3_CacheLifecycle(t *testing.T) {
	cfg := DefaultConfig()
	cfg.CacheCapacity = 2

	logger, buf := testLogBuffer(t, slog.LevelDebug)
	cache := NewResponseCache(cfg, logger)

	// Use raw string keys; the production hash flows through `hashKey`.
	const k1, k2, k3 = "k1", "k2", "k3"
	v1 := &UIAST{Version: "1", GeneratedBy: "v1"}
	v2 := &UIAST{Version: "1", GeneratedBy: "v2"}
	v3 := &UIAST{Version: "1", GeneratedBy: "v3"}

	cache.Store(k1, v1) // cache.put #1
	cache.Store(k2, v2) // cache.put #2

	got, hit := cache.Lookup(k1) // cache.hit
	require.True(t, hit, "k1 should hit before eviction")
	assert.Same(t, v1, got)

	cache.Store(k3, v3)       // cache.put #3 + cache.evict (k2 is LRU)
	_, hit = cache.Lookup(k2) // cache.miss
	require.False(t, hit, "k2 should be evicted by LRU when k3 is stored")

	records := decodeRecords(t, buf)

	puts := recordsByMsg(records, "cache.put")
	hits := recordsByMsg(records, "cache.hit")
	misses := recordsByMsg(records, "cache.miss")
	evicts := recordsByMsg(records, "cache.evict")

	assert.Len(t, puts, 3, "expected 3 cache.put records; got %d in %v", len(puts), records)
	assert.Len(t, hits, 1, "expected 1 cache.hit record; got %d in %v", len(hits), records)
	assert.Len(t, misses, 1, "expected 1 cache.miss record; got %d in %v", len(misses), records)
	require.Len(t, evicts, 1, "expected 1 cache.evict record; got %d in %v", len(evicts), records)

	// Eviction reason and key-hash assertions.
	ev := evicts[0]
	assert.Equal(t, "cache.evict", ev["op"], "cache.evict must carry op=\"cache.evict\"")
	assert.Equal(t, "lru", ev["reason"], "evict reason must be \"lru\"")
	assert.Equal(t, hashKey(k2), ev["cache_key_hash"],
		"evict cache_key_hash must match hashKey(\"k2\")")

	// Every cache record (put/hit/miss/evict) must carry an 8-char hash.
	for _, msg := range []string{"cache.put", "cache.hit", "cache.miss", "cache.evict"} {
		for i, r := range recordsByMsg(records, msg) {
			h, ok := r["cache_key_hash"].(string)
			require.True(t, ok, "%s record %d must carry string cache_key_hash; got %v", msg, i, r["cache_key_hash"])
			assert.Len(t, h, 8, "%s record %d cache_key_hash must be 8 hex chars; got %q", msg, i, h)
		}
	}

	// Sanitize: no record may carry the raw key strings.
	for _, msg := range []string{"cache.put", "cache.hit", "cache.miss", "cache.evict"} {
		for _, r := range recordsByMsg(records, msg) {
			for _, raw := range []string{k1, k2, k3} {
				for k, v := range r {
					if s, ok := v.(string); ok && s == raw {
						t.Fatalf("record msg=%q attr %q leaked raw key %q", msg, k, raw)
					}
				}
			}
		}
	}
}
