package backend

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"mashed/internal/uiadapter"
)

// slowBackend — WarmUp sleeps; used to verify parallelism.
type slowBackend struct {
	*StubBackend
	sleep time.Duration
	calls atomic.Int64
}

func (s *slowBackend) WarmUp(ctx context.Context) error {
	s.calls.Add(1)
	time.Sleep(s.sleep)
	return nil
}

// TestLifecycle_WarmUpAllInParallel — AC-17.3. Three backends each
// sleeping 100ms complete in ≈100ms, not 300ms.
func TestLifecycle_WarmUpAllInParallel(t *testing.T) {
	t.Parallel()
	cfg := uiadapter.DefaultConfig()
	cfg.WarmUpTimeoutMs = 2000
	backends := map[string]LLMBackend{}
	for _, n := range []string{"ollama", "claude-api", "claude-cli"} {
		backends[n] = &slowBackend{StubBackend: NewStub(n, Capabilities{}), sleep: 100 * time.Millisecond}
	}
	lc := NewLifecycle(cfg, backends, nil)

	start := time.Now()
	lc.WarmUpAll(context.Background())
	elapsed := time.Since(start)
	assert.Less(t, elapsed, 250*time.Millisecond,
		"three 100ms WarmUps must run in parallel: elapsed=%v", elapsed)
}

// TestLifecycle_WarmUpRecordsResults — every backend has an entry in
// WarmUpState after WarmUpAll.
func TestLifecycle_WarmUpRecordsResults(t *testing.T) {
	t.Parallel()
	backends := map[string]LLMBackend{
		"ollama":     NewStub("ollama", Capabilities{}),
		"claude-api": NewStub("claude-api", Capabilities{}),
	}
	lc := NewLifecycle(uiadapter.DefaultConfig(), backends, nil)
	lc.WarmUpAll(context.Background())
	state := lc.WarmUpState()
	assert.Len(t, state, 2)
}

// TestLifecycle_TickerDisableable — AC-17.2. DisableHealthTicker=true
// makes StartHealthTicker a no-op.
func TestLifecycle_TickerDisableable(t *testing.T) {
	t.Parallel()
	cfg := uiadapter.DefaultConfig()
	cfg.DisableHealthTicker = true
	backends := map[string]LLMBackend{"ollama": NewStub("ollama", Capabilities{})}
	lc := NewLifecycle(cfg, backends, nil)
	stop := lc.StartHealthTicker(10 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	stop()
	// The stub counter is checked via Calls(); Health was not invoked.
	sb := backends["ollama"].(*StubBackend)
	_, _, _, _, health := sb.Calls()
	assert.Zero(t, health)
}

// TestLifecycle_HealthTickerFeedsRouter — Ticker Health results flow
// into the router so Decide reads latest state.
func TestLifecycle_HealthTickerFeedsRouter(t *testing.T) {
	t.Parallel()
	backends := map[string]LLMBackend{
		"ollama":     NewStub("ollama", Capabilities{}),
		"claude-api": NewStub("claude-api", Capabilities{}),
	}
	cfg := uiadapter.DefaultConfig()
	cfg.RouterPolicy = "cost-aware"
	r := NewRouter(cfg, backends)
	lc := NewLifecycle(cfg, backends, r)
	stop := lc.StartHealthTicker(10 * time.Millisecond)
	t.Cleanup(stop)
	// Wait for at least one tick.
	time.Sleep(50 * time.Millisecond)
	// Stub Health returns nil so router sees ollama healthy → cost-aware
	// picks ollama.
	d := r.Decide("x")
	assert.Equal(t, "ollama", d.Primary)
}

// TestLifecycle_WarmUpConcurrentSafe — repeated WarmUpAll calls don't
// double-warm (sync.Once guarantee).
func TestLifecycle_WarmUpConcurrentSafe(t *testing.T) {
	t.Parallel()
	sb := &slowBackend{StubBackend: NewStub("ollama", Capabilities{}), sleep: 10 * time.Millisecond}
	lc := NewLifecycle(uiadapter.DefaultConfig(), map[string]LLMBackend{"ollama": sb}, nil)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); lc.WarmUpAll(context.Background()) }()
	}
	wg.Wait()
	assert.EqualValues(t, 1, sb.calls.Load(), "WarmUp runs exactly once across concurrent callers")
}
