package uiadapter

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"sync/atomic"

	lru "github.com/hashicorp/golang-lru/v2"
	"golang.org/x/sync/singleflight"
)

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
	if cap <= 0 {
		return &ResponseCache{logger: nilSafeLogger(logger)}
	}
	l, err := lru.New[string, *UIAST](cap)
	if err != nil {
		// lru.New only errors on size ≤ 0; we guarded above, so this is
		// a programmer error per §6.5 "Never panic in library code.
		// Panic is for programmer bugs only."
		panic("uiadapter: ResponseCache lru.New: " + err.Error())
	}
	return &ResponseCache{lru: l, logger: nilSafeLogger(logger)}
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
	if c == nil || c.lru == nil {
		c.miss.Add(1)
		return nil, false
	}
	if v, ok := c.lru.Get(key); ok {
		c.hits.Add(1)
		return v, true
	}
	c.miss.Add(1)
	return nil, false
}

// Store populates the cache. No-op on disabled caches.
func (c *ResponseCache) Store(key string, ast *UIAST) {
	if c == nil || c.lru == nil || ast == nil {
		return
	}
	c.lru.Add(key, ast)
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
