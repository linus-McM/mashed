package backend

import (
	"context"
	"sync"
	"time"

	"mashed/internal/uiadapter"
)

// Plan §3 Story 17 — Per-backend WarmUp + Health. Parallel WarmUp at
// construction, periodic Health ticker with injection-friendly clock.

// Lifecycle owns the WarmUp/Health orchestration for a set of backends.
type Lifecycle struct {
	cfg      uiadapter.Config
	backends map[string]LLMBackend
	router   *Router

	mu        sync.Mutex
	warmed    map[string]error
	stopHealth chan struct{}
	once      sync.Once
}

// NewLifecycle constructs a Lifecycle bound to the router. If router is
// non-nil, Health results flow into router.SetHealth so Decide reads the
// latest state.
func NewLifecycle(cfg uiadapter.Config, backends map[string]LLMBackend, router *Router) *Lifecycle {
	return &Lifecycle{
		cfg:      cfg,
		backends: backends,
		router:   router,
		warmed:   map[string]error{},
	}
}

// WarmUpAll fires every backend's WarmUp in parallel, each under
// Config.WarmUpTimeoutMs. Errors are recorded but non-fatal — the
// adapter keeps running with whatever backends returned nil. Total
// elapsed time ≈ max(per-backend WarmUp), not sum (AC-17.3).
func (l *Lifecycle) WarmUpAll(ctx context.Context) {
	timeout := time.Duration(l.cfg.WarmUpTimeoutMs) * time.Millisecond
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	var wg sync.WaitGroup
	l.once.Do(func() {
		for name, b := range l.backends {
			wg.Add(1)
			go func(name string, b LLMBackend) {
				defer wg.Done()
				wctx, cancel := context.WithTimeout(ctx, timeout)
				defer cancel()
				err := b.WarmUp(wctx)
				l.mu.Lock()
				l.warmed[name] = err
				l.mu.Unlock()
				if l.router != nil {
					l.router.SetHealth(name, err)
				}
			}(name, b)
		}
	})
	wg.Wait()
}

// WarmUpState returns a snapshot of per-backend WarmUp results.
func (l *Lifecycle) WarmUpState() map[string]error {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make(map[string]error, len(l.warmed))
	for k, v := range l.warmed {
		out[k] = v
	}
	return out
}

// StartHealthTicker spawns a goroutine that calls Health on every
// backend at interval. Config.DisableHealthTicker=true makes this a
// no-op (AC-17.2). Returns a stop function.
//
// The stop channel is captured locally so the ticker goroutine + the
// returned stop closure share a single channel reference, no shared
// mutable state on the Lifecycle struct.
func (l *Lifecycle) StartHealthTicker(interval time.Duration) (stop func()) {
	if l.cfg.DisableHealthTicker {
		return func() {}
	}
	if interval <= 0 {
		interval = 30 * time.Second
	}
	stopCh := make(chan struct{})
	l.mu.Lock()
	l.stopHealth = stopCh
	l.mu.Unlock()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				l.healthProbeAll()
			}
		}
	}()
	var once sync.Once
	return func() {
		once.Do(func() { close(stopCh) })
	}
}

func (l *Lifecycle) healthProbeAll() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for name, b := range l.backends {
		err := b.Health(ctx)
		if l.router != nil {
			l.router.SetHealth(name, err)
		}
	}
}
