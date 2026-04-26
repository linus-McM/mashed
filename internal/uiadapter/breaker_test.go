package uiadapter

import (
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBreaker_TripsAfterThree — AC-11.1. Three consecutive failures open
// the breaker; the next call is rejected in <5ms.
func TestBreaker_TripsAfterThree(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.BreakerFailThreshold = 3
	cfg.BreakerResetMs = 30000
	set := NewBreakerSet(cfg, nil)

	boom := errors.New("boom")
	for i := 0; i < 3; i++ {
		_, err := set.Do("flaky", func() (*UIAST, error) { return nil, boom })
		require.Error(t, err)
	}

	start := time.Now()
	_, err := set.Do("flaky", func() (*UIAST, error) {
		t.Fatal("breaker should have rejected before invoking fn")
		return nil, nil
	})
	elapsed := time.Since(start)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBreakerOpen)
	assert.Less(t, elapsed, 5*time.Millisecond, "rejection must be sub-5ms")
	assert.Equal(t, "open", set.StateOf("flaky"))
}

// TestBreaker_PerBackendIsolation — AC-11.4. Tripping one backend does
// not affect another.
func TestBreaker_PerBackendIsolation(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.BreakerFailThreshold = 2
	set := NewBreakerSet(cfg, nil)

	boom := errors.New("boom")
	// Trip "claude-api".
	for i := 0; i < 2; i++ {
		_, _ = set.Do("claude-api", func() (*UIAST, error) { return nil, boom })
	}
	assert.Equal(t, "open", set.StateOf("claude-api"))
	// "ollama" should still be closed.
	assert.Equal(t, "closed", set.StateOf("ollama"))
	got, err := set.Do("ollama", func() (*UIAST, error) {
		return &UIAST{Version: "1"}, nil
	})
	require.NoError(t, err)
	require.NotNil(t, got)
}

// TestBreaker_StateOnIdleIsClosed — an untouched breaker starts closed.
func TestBreaker_StateOnIdleIsClosed(t *testing.T) {
	t.Parallel()
	set := NewBreakerSet(DefaultConfig(), nil)
	assert.Equal(t, "closed", set.StateOf("new-backend"))
}

// TestBreaker_ConcurrentConstruction — the lazy For() map must be
// goroutine-safe.
func TestBreaker_ConcurrentConstruction(t *testing.T) {
	t.Parallel()
	set := NewBreakerSet(DefaultConfig(), nil)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = set.For("ollama")
		}()
	}
	wg.Wait()
	assert.Equal(t, "closed", set.StateOf("ollama"))
}

// -----------------------------------------------------------------------------
// Story 3 — `breaker.go` debug instrumentation (uiadapter-logging-3).
//
// RED-phase: these target log records that do not yet exist in production.
// -----------------------------------------------------------------------------

// TestStory3_AC4_BreakerTransitionAndReject — Story 3, AC-3.4.
//
// Two consecutive failures on backend "ollama" must emit a `breaker.transition`
// record (from="closed", to="open", backend="ollama"). A subsequent request
// while the breaker is open must emit `breaker.reject` with state="open".
func TestStory3_AC4_BreakerTransitionAndReject(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BreakerFailThreshold = 2
	cfg.BreakerResetMs = 30000

	logger, buf := testLogBuffer(t, slog.LevelDebug)
	set := NewBreakerSet(cfg, logger)

	const backend = "ollama"
	boom := errors.New("upstream-failure")

	// Two failures: trip the breaker.
	for i := 0; i < 2; i++ {
		_, err := set.Do(backend, func() (*UIAST, error) { return nil, boom })
		require.Error(t, err)
	}
	require.Equal(t, "open", set.StateOf(backend),
		"breaker must be open after %d consecutive failures", 2)

	// Third request: must be rejected (ErrBreakerOpen) and emit breaker.reject.
	_, rejErr := set.Do(backend, func() (*UIAST, error) {
		t.Fatal("breaker.reject must short-circuit; fn must not run while open")
		return nil, nil
	})
	require.ErrorIs(t, rejErr, ErrBreakerOpen)

	records := decodeRecords(t, buf)

	// breaker.transition (closed → open) — exactly one such record for this run.
	transitions := recordsByMsg(records, "breaker.transition")
	require.NotEmpty(t, transitions, "expected at least one breaker.transition record; got %v", records)

	var sawClosedToOpen bool
	for _, rec := range transitions {
		if rec["from"] == "closed" && rec["to"] == "open" && rec["backend"] == backend {
			sawClosedToOpen = true
			assert.Equal(t, "breaker.transition", rec["op"],
				"breaker.transition must carry op=\"breaker.transition\"")
		}
	}
	assert.True(t, sawClosedToOpen,
		"expected a breaker.transition record with from=\"closed\",to=\"open\",backend=%q; got %v",
		backend, transitions)

	// breaker.reject — at least one record carrying state="open" + backend.
	rejects := recordsByMsg(records, "breaker.reject")
	require.NotEmpty(t, rejects, "expected a breaker.reject record after open; got %v", records)
	rec := rejects[0]
	assert.Equal(t, "breaker.reject", rec["op"], "breaker.reject must carry op=\"breaker.reject\"")
	assert.Equal(t, backend, rec["backend"], "breaker.reject must carry backend=%q", backend)
	assert.Equal(t, "open", rec["state"], "breaker.reject must carry state=\"open\"")
}
