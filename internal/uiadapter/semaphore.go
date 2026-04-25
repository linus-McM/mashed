package uiadapter

import (
	"context"
	"log/slog"
)

// semaphore is a bounded-concurrency gate backed by a buffered channel. A send
// acquires a slot; a receive releases it. The zero value is not useful — use
// newSemaphore.
//
// The struct carries a *slog.Logger field so future stories can instrument
// acquire/release telemetry without another constructor reshape; Story 2
// only plumbs the field, no log calls land here yet.
type semaphore struct {
	ch     chan struct{}
	logger *slog.Logger
}

// newSemaphore returns a semaphore permitting n concurrent holders. n must be
// > 0; callers clamp. logger may be nil — nilSafeLogger normalises it to a
// discard-backed logger so the field is always usable.
func newSemaphore(n int, logger *slog.Logger) semaphore {
	return semaphore{
		ch:     make(chan struct{}, n),
		logger: nilSafeLogger(logger),
	}
}

// acquire attempts a non-blocking send first so a pre-canceled context never
// races the fast path. Go's select randomises among ready cases; without the
// two-phase attempt, a pre-canceled caller on an empty sem could flip between
// saturated and canceled reasons.
func (s semaphore) acquire(ctx context.Context) bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
	}
	select {
	case s.ch <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

// release frees one slot. Pair with acquire via defer.
func (s semaphore) release() { <-s.ch }
