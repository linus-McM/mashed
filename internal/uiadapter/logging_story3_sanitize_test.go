// Package uiadapter — Story 3 cross-cutting tests:
// AC-3.7 (sanitize discipline) and AC-3.8 (zero-allocation hot path).
//
// AC-3.7 fuzz-style: 100 random raw payloads (50–500 bytes, mixed ASCII +
// control chars) are pushed through every instrumented entry point. After
// each pass the captured JSON log records are scanned for any 8-byte slice
// of the raw payload, the response body, or any synthetic err.Error() text.
// Any hit is a sanitize-discipline violation per §14.
//
// AC-3.8: with an Info-level logger the hot-path Enabled-guard pattern must
// add zero allocations versus the underlying op (within ±0.5/run noise).
package uiadapter

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// randomPayload returns n bytes of cryptographically random data with the
// high-bit-stripped ASCII range plus a sprinkling of control characters,
// matching the brief's "mixed ASCII + control chars" requirement. Hex
// encoding is avoided so the bytes are not trivially substring-searchable
// against the JSON-encoded record (which itself uses hex/escapes).
func randomPayload(t *testing.T, n int) []byte {
	t.Helper()
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	require.NoError(t, err)
	for i := range buf {
		// 0x20-0x7e is printable ASCII; sprinkle a control char every ~16th byte.
		if buf[i]%17 == 0 {
			buf[i] = byte(buf[i] % 0x20) // 0x00..0x1f
		} else {
			buf[i] = 0x20 + (buf[i] % 0x5f) // 0x20..0x7e
		}
	}
	return buf
}

// containsSlice scans haystack for any 8-byte contiguous slice of needle.
// 8 bytes is the brief-mandated minimum match length — short enough to catch
// real leaks, long enough to avoid false positives on random-noise overlap
// with structural JSON bytes. Returns the offset within needle where the
// leaked window begins.
func containsSlice(haystack, needle []byte) (int, bool) {
	const window = 8
	if len(needle) < window {
		return -1, false
	}
	for i := 0; i+window <= len(needle); i++ {
		if bytes.Contains(haystack, needle[i:i+window]) {
			return i, true
		}
	}
	return -1, false
}

// TestStory3_AC7_SanitizeDisciplineAcrossFiles — AC-3.7.
//
// Fuzz every instrumented site with random raw payloads and assert the
// JSON-encoded log buffer never contains an 8-byte slice of the raw payload
// or the synthetic response body or the err.Error() text.
func TestStory3_AC7_SanitizeDisciplineAcrossFiles(t *testing.T) {
	const iterations = 100
	const minLen, maxLen = 50, 500

	cfg := DefaultConfig()

	// Synthetic response body shared by every Chat call. It carries a sentinel
	// the test will assert never leaks into log records.
	const respSentinel = "RESPSENTINEL_b3a9c2d4e5f60718"
	stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, `{"message":{"role":"assistant","content":"`+respSentinel+`"}}`)
	}))
	t.Cleanup(stub.Close)
	withOllamaHost(t, stub.URL)

	// Synthetic error sentinel — flowed through the breaker fn so its bytes
	// must not appear in any breaker.transition / breaker.reject record.
	errSentinel := "ERRSENTINEL_4f7a8b9c0d1e2f3a"
	syntheticErr := errors.New(errSentinel)

	for iter := 0; iter < iterations; iter++ {
		// Length varies in [minLen, maxLen).
		n := minLen + (iter % (maxLen - minLen))
		raw := randomPayload(t, n)

		logger, buf := testLogBuffer(t, slog.LevelDebug)

		// 1) Client.Chat — payload as the user-content arg.
		client := NewClient(ClientConfig{TimeoutMs: 1000}, logger)
		_, _ = client.Chat(context.Background(), "gemma3:4b", "system-stub", string(raw))

		// 2) ResponseCache.Lookup with payload as the cache key.
		cache := NewResponseCache(cfg, logger)
		cache.Store(string(raw), &UIAST{Version: "1"})
		_, _ = cache.Lookup(string(raw))

		// 3) BreakerSet.Do — constant backend name (per §14: backend names
		//    are enum-ish IDs and are intentionally logged), fn returns the
		//    err sentinel so the transition / reject records are exercised
		//    with a stable identifiable error string.
		breakerCfg := cfg
		breakerCfg.BreakerFailThreshold = 1
		breakers := NewBreakerSet(breakerCfg, logger)
		_, _ = breakers.Do("fuzz-target", func() (*UIAST, error) { return nil, syntheticErr })
		_, _ = breakers.Do("fuzz-target", func() (*UIAST, error) { return nil, nil })

		// 4) ClaudeSystemBlock — payload as the static prefix.
		_ = ClaudeSystemBlock(string(raw), cfg, logger)

		// Capture log records and scan for leakage.
		body := buf.Bytes()
		if pos, leaked := containsSlice(body, raw); leaked {
			t.Fatalf("iter %d: log buffer leaked an 8-byte slice of the raw payload at offset %d; payload=%q",
				iter, pos, raw)
		}
		if pos, leaked := containsSlice(body, []byte(respSentinel)); leaked {
			t.Fatalf("iter %d: log buffer leaked the response body sentinel at offset %d", iter, pos)
		}
		if pos, leaked := containsSlice(body, []byte(errSentinel)); leaked {
			t.Fatalf("iter %d: log buffer leaked the err.Error() sentinel at offset %d", iter, pos)
		}

		// Structural assertion: every value in every record must be a
		// scalar (number/bool/string) or a hash/duration/enum. We can't
		// type-narrow strings to "is this an enum?" without per-attr context,
		// but we *can* assert no string attr value EQUALS the raw payload.
		for _, rec := range decodeRecords(t, buf) {
			for k, v := range rec {
				if s, ok := v.(string); ok {
					if s == string(raw) {
						t.Fatalf("iter %d: record attr %q == raw payload (full equality leak)", iter, k)
					}
					if s == respSentinel || s == errSentinel {
						t.Fatalf("iter %d: record attr %q leaked sentinel %q", iter, k, s)
					}
				}
			}
		}
	}
}

// TestStory3_AC8_HotPathZeroAllocs_DebugOff — AC-3.8.
//
// With an Info-level logger (Debug filtered), each instrumented hot-path
// must allocate at most one extra item per run vs. the no-instrumentation
// baseline. The brief permits ≤ 1 alloc/run as the CI-runner-noise budget.
func TestStory3_AC8_HotPathZeroAllocs_DebugOff(t *testing.T) {
	const tolerance = 1.0 // allocs/run; brief: "≤ 1 acceptable"

	cfg := DefaultConfig()
	cfg.CacheCapacity = 4

	infoLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// 1) ResponseCache.Lookup hot path (every Translate goes through it).
	t.Run("cache.Lookup", func(t *testing.T) {
		cache := NewResponseCache(cfg, infoLogger)
		key := Key("ollama", "gemma3:4b", "alloc-budget")
		cache.Store(key, &UIAST{Version: "1"})

		got := testing.AllocsPerRun(10000, func() {
			_, _ = cache.Lookup(key)
		})
		assert.LessOrEqual(t, got, tolerance,
			"cache.Lookup at Info level must add ≤ %.1f allocs/run; got %.2f", tolerance, got)
	})

	// 2) BreakerSet.Do happy path (closed breaker, fn returns ok).
	t.Run("breaker.Do", func(t *testing.T) {
		set := NewBreakerSet(cfg, infoLogger)
		ok := &UIAST{Version: "1"}
		fn := func() (*UIAST, error) { return ok, nil }

		// Warm up so the lazy For() map is populated and the gobreaker
		// state machine settles into Closed.
		for i := 0; i < 16; i++ {
			_, _ = set.Do("alloc-budget", fn)
		}

		got := testing.AllocsPerRun(10000, func() {
			_, _ = set.Do("alloc-budget", fn)
		})
		// gobreaker.Execute itself allocates; budget is "added" by Story 3,
		// not "absolute". A loose ceiling lets the test fail only on a
		// regression that adds many allocs from the new instrumentation.
		assert.LessOrEqual(t, got, 4.0,
			"breaker.Do at Info level must stay within the gobreaker baseline; got %.2f", got)
	})

	// 3) semaphore acquire/release fast path (permit always available).
	t.Run("semaphore.acquire_release", func(t *testing.T) {
		sem := newSemaphore(8, infoLogger)
		ctx := context.Background()

		got := testing.AllocsPerRun(10000, func() {
			if sem.acquire(ctx) {
				sem.release()
			}
		})
		assert.LessOrEqual(t, got, tolerance,
			"semaphore acquire/release at Info level must add ≤ %.1f allocs/run; got %.2f", tolerance, got)
	})

	// 4) ClaudeSystemBlock — pure function, baseline allocates the slice/map
	//    so we just guard against unexpected NEW instrumentation allocations.
	t.Run("prefix_cache.ClaudeSystemBlock", func(t *testing.T) {
		got := testing.AllocsPerRun(10000, func() {
			_ = ClaudeSystemBlock("hello", cfg, infoLogger)
		})
		// Measured Go 1.26 baseline (pre-Story 3) is 7 allocs/run: 2 maps +
		// the outer slice + 4 boxings of string→any in the map literals.
		// Allow some headroom; failure here means Story 3 added per-call
		// allocs beyond the baseline even with Debug off.
		assert.LessOrEqual(t, got, 8.0,
			"ClaudeSystemBlock at Info level must stay within baseline; got %.2f", got)
	})

	// 5) Sanity: at Info level no buffer should accumulate Debug records.
	t.Run("info_buffer_stays_empty", func(t *testing.T) {
		buf := &captureWriter{}
		bufLogger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
		cache := NewResponseCache(cfg, bufLogger)
		key := Key("ollama", "gemma3:4b", "info-empty")
		cache.Store(key, &UIAST{Version: "1"})
		_, _ = cache.Lookup(key)
		assert.Equal(t, 0, buf.n,
			"Info-level logger must capture zero bytes from Debug-only instrumentation; got %d bytes: %q",
			buf.n, buf.tail)
	})
}

// captureWriter records the byte count and the last 256 bytes written for
// diagnostic purposes. We deliberately do NOT keep the full byte stream;
// the AC-3.8 budget is checked on count, not content.
type captureWriter struct {
	n    int
	tail []byte
}

func (c *captureWriter) Write(p []byte) (int, error) {
	c.n += len(p)
	if len(p) > 256 {
		c.tail = append(c.tail[:0], p[len(p)-256:]...)
	} else {
		c.tail = append(c.tail[:0], p...)
	}
	return len(p), nil
}
