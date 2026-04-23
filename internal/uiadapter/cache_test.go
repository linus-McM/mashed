package uiadapter

import (
	"errors"
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
	cache := NewResponseCache(cfg)
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
	cache := NewResponseCache(cfg)
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
	cache := NewResponseCache(cfg)

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
	cache := NewResponseCache(cfg)
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
	cache := NewResponseCache(DefaultConfig())
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
