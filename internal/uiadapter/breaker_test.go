package uiadapter

import (
	"errors"
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
	set := NewBreakerSet(cfg)

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
	set := NewBreakerSet(cfg)

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
	set := NewBreakerSet(DefaultConfig())
	assert.Equal(t, "closed", set.StateOf("new-backend"))
}

// TestBreaker_ConcurrentConstruction — the lazy For() map must be
// goroutine-safe.
func TestBreaker_ConcurrentConstruction(t *testing.T) {
	t.Parallel()
	set := NewBreakerSet(DefaultConfig())
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
