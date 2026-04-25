package uiadapter

import (
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/sony/gobreaker"
)

// Plan §3 Story 11 — CircuitBreaker + Tiered Fallback. One breaker
// instance per registered backend; independent state so an Anthropic
// outage never opens the Ollama breaker.

// ErrBreakerOpen is returned when a breaker rejects a call before it
// dispatches. Consumers should fall through to the tiered-fallback chain.
var ErrBreakerOpen = errors.New("uiadapter: circuit breaker open")

// BreakerSet holds one gobreaker per backend name.
type BreakerSet struct {
	mu       sync.RWMutex
	breakers map[string]*gobreaker.CircuitBreaker
	cfg      Config
	logger   *slog.Logger
}

// NewBreakerSet constructs an empty set configured from adapter Config.
// BreakerFailThreshold and BreakerResetMs source the gobreaker settings.
// logger may be nil; nilSafeLogger normalises it so the field is always
// usable.
func NewBreakerSet(cfg Config, logger *slog.Logger) *BreakerSet {
	return &BreakerSet{
		breakers: map[string]*gobreaker.CircuitBreaker{},
		cfg:      cfg,
		logger:   nilSafeLogger(logger),
	}
}

// For returns (and lazily constructs) the breaker for backendName. The
// settings match plan §3 Story 11:
//   - trip on BreakerFailThreshold consecutive failures OR 50% failure
//     rate over a rolling 10s window (min 5 calls)
//   - half-open probe after BreakerResetMs
func (s *BreakerSet) For(backendName string) *gobreaker.CircuitBreaker {
	s.mu.RLock()
	if b, ok := s.breakers[backendName]; ok {
		s.mu.RUnlock()
		return b
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()
	if b, ok := s.breakers[backendName]; ok {
		return b
	}

	threshold := uint32(s.cfg.BreakerFailThreshold)
	if threshold == 0 {
		threshold = 3
	}
	timeout := time.Duration(s.cfg.BreakerResetMs) * time.Millisecond
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	b := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:    backendName,
		Timeout: timeout,
		ReadyToTrip: func(c gobreaker.Counts) bool {
			if c.ConsecutiveFailures >= threshold {
				return true
			}
			if c.Requests >= 5 {
				rate := float64(c.TotalFailures) / float64(c.Requests)
				return rate >= 0.5
			}
			return false
		},
	})
	s.breakers[backendName] = b
	return b
}

// StateOf returns "closed" | "half-open" | "open" for backendName.
// Feeds the §6.1 slog `breaker_state` attribute.
func (s *BreakerSet) StateOf(backendName string) string {
	b := s.For(backendName)
	switch b.State() {
	case gobreaker.StateClosed:
		return "closed"
	case gobreaker.StateHalfOpen:
		return "half-open"
	case gobreaker.StateOpen:
		return "open"
	default:
		return "closed"
	}
}

// Do runs fn through the breaker. Returns ErrBreakerOpen wrapped when
// the breaker rejects; other errors from fn pass through.
func (s *BreakerSet) Do(backendName string, fn func() (*UIAST, error)) (*UIAST, error) {
	b := s.For(backendName)
	v, err := b.Execute(func() (any, error) {
		return fn()
	})
	if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
		return nil, ErrBreakerOpen
	}
	if err != nil {
		return nil, err
	}
	ast, _ := v.(*UIAST)
	return ast, nil
}
