package uiadapter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"sync/atomic"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/singleflight"
)

// hashKey returns the first 8 hex chars of sha256(s). Used as the
// `cache_key_hash` log attribute so observability never sees the raw key.
// Standard attrs cross-reference: see logging.go.
func hashKey(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:4])
}

// cacheVersion bumps whenever prompts / schemas change in a way that
// invalidates cached UIASTs. Plan §3 Story 4 "simplest correct
// invalidation". Touch this when editing any prompt file or schema.
const cacheVersion = "v3"

// ResponseCache is the in-memory LRU + singleflight pair that fronts every
// LLM call. Keys are sha256(backend + model + cacheVersion + sanitized
// raw) so outputs never collide across backends and switching
// Config.Backend silently invalidates prior keys (AC-4.5).
//
// The singleflight group collapses concurrent identical calls to a single
// downstream execution (AC-4.2). Capacity is Config.CacheCapacity (Story
// B); 0 disables the LRU but keeps singleflight — useful for tests.
type ResponseCache struct {
	lru    *lru.Cache[string, *UIAST]
	sf     singleflight.Group
	hits   atomic.Int64
	miss   atomic.Int64
	logger *slog.Logger
}

// NewResponseCache constructs the cache from the adapter Config. Capacity
// ≤0 returns a cache with no LRU (singleflight still works). logger may be
// nil; nilSafeLogger normalises it so the field is always usable.
func NewResponseCache(cfg Config, logger *slog.Logger) *ResponseCache {
	cap := cfg.CacheCapacity
	safeLogger := nilSafeLogger(logger)
	if cap <= 0 {
		return &ResponseCache{logger: safeLogger}
	}
	c := &ResponseCache{logger: safeLogger}
	l, err := lru.NewWithEvict[string, *UIAST](cap, c.onEvicted)
	if err != nil {
		// lru.New only errors on size ≤ 0; we guarded above, so this is
		// a programmer error per §6.5 "Never panic in library code.
		// Panic is for programmer bugs only."
		panic("uiadapter: ResponseCache lru.New: " + err.Error())
	}
	c.lru = l
	return c
}

// onEvicted is the lru eviction callback. Hot-path-guarded so a non-Debug
// logger pays only the Enabled() check.
func (c *ResponseCache) onEvicted(key string, _ *UIAST) {
	ctx := context.Background()
	if c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.LogAttrs(ctx, slog.LevelDebug, "cache.evict",
			slog.String("op", "cache.evict"),
			slog.String("cache_key_hash", hashKey(key)),
			slog.String("reason", "lru"),
		)
	}
}

// Key returns the cache key for a (backend, model, sanitizedRaw) triple.
// Exported so callers can precompute the key once across cache Lookup
// and Singleflight.Do invocations.
func Key(backendName, model, sanitizedRaw string) string {
	h := sha256.New()
	h.Write([]byte(backendName))
	h.Write([]byte{0x00})
	h.Write([]byte(model))
	h.Write([]byte{0x00})
	h.Write([]byte(cacheVersion))
	h.Write([]byte{0x00})
	h.Write([]byte(sanitizedRaw))
	return hex.EncodeToString(h.Sum(nil))
}

// Lookup returns the cached UIAST and hit=true on a hit, or (nil, false)
// on a miss. Disabled caches always miss. Counters are atomic.
func (c *ResponseCache) Lookup(key string) (*UIAST, bool) {
	if c == nil {
		return nil, false
	}
	if c.lru != nil {
		if v, ok := c.lru.Get(key); ok {
			c.hits.Add(1)
			c.logCacheGet(key, true)
			return v, true
		}
	}
	c.miss.Add(1)
	c.logCacheGet(key, false)
	return nil, false
}

// Store populates the cache. No-op on disabled caches.
func (c *ResponseCache) Store(key string, ast *UIAST) {
	if c == nil || c.lru == nil || ast == nil {
		return
	}
	c.lru.Add(key, ast)
	ctx := context.Background()
	if c.logger.Enabled(ctx, slog.LevelDebug) {
		c.logger.LogAttrs(ctx, slog.LevelDebug, "cache.put",
			slog.String("op", "cache.put"),
			slog.String("cache_key_hash", hashKey(key)),
		)
	}
}

// logCacheGet emits cache.hit or cache.miss with the standard attrs. The
// hot-path Enabled guard keeps Info-level callers allocation-free.
func (c *ResponseCache) logCacheGet(key string, hit bool) {
	ctx := context.Background()
	if !c.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	msg := "cache.miss"
	if hit {
		msg = "cache.hit"
	}
	c.logger.LogAttrs(ctx, slog.LevelDebug, msg,
		slog.String("op", "cache.get"),
		slog.String("cache_key_hash", hashKey(key)),
	)
}

// HitRate is hits / (hits + misses). Feeds the slog `cache_hit_rate`
// attribute on every Translate line (AC-4.3).
func (c *ResponseCache) HitRate() float64 {
	if c == nil {
		return 0
	}
	h := c.hits.Load()
	m := c.miss.Load()
	total := h + m
	if total == 0 {
		return 0
	}
	return float64(h) / float64(total)
}

// DoShared coalesces concurrent identical calls. `shared=true` means this
// goroutine received a result produced by another goroutine's invocation
// — AC-4.3's `singleflight_shared` attribute surfaces it on the log line.
// The fn argument is only invoked at most once per key across concurrent
// callers.
func (c *ResponseCache) DoShared(key string, fn func() (*UIAST, error)) (*UIAST, bool, error) {
	v, err, shared := c.sf.Do(key, func() (any, error) {
		return fn()
	})
	if err != nil {
		return nil, shared, err
	}
	ast, _ := v.(*UIAST)
	return ast, shared, nil
}

// Metrics returns a snapshot (hits, misses). Exposed for tests.
func (c *ResponseCache) Metrics() (hits, misses int64) {
	if c == nil {
		return 0, 0
	}
	return c.hits.Load(), c.miss.Load()
}
