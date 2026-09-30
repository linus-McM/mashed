package uiadapter

import (
	"context"
	"log/slog"
	"time"
)

// semaphore is a bounded-concurrency gate backed by a buffered channel. A send
// acquires a slot; a receive releases it. The zero value is not useful — use
// newSemaphore.
//
// Story 3 (uiadapter-logging-3) instruments acquire with debug logs:
// semaphore.wait / semaphore.acquired / semaphore.cancelled. All sites are
// hot-path-guarded by logger.Enabled so the Info-level path stays
// allocation-free.
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
	start := time.Now()
	select {
	case s.ch <- struct{}{}:
		s.logAcquired(ctx, 0)
		return true
	default:
	}
	select {
	case s.ch <- struct{}{}:
		waitMs := time.Since(start).Milliseconds()
		s.logWaitIfPositive(ctx, waitMs)
		s.logAcquired(ctx, waitMs)
		return true
	case <-ctx.Done():
		s.logCancelled(ctx, time.Since(start).Milliseconds())
		return false
	}
}

// logWaitIfPositive emits semaphore.wait when the caller blocked for at
// least 1ms. Hot-path-guarded so a non-Debug logger pays only the Enabled
// check.
func (s semaphore) logWaitIfPositive(ctx context.Context, waitMs int64) {
	if waitMs <= 0 || !s.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	s.logger.LogAttrs(ctx, slog.LevelDebug, "semaphore.wait",
		slog.String("op", "semaphore.acquire"),
		slog.Int64("wait_ms", waitMs),
		slog.Int("in_flight", cap(s.ch)-len(s.ch)),
	)
}

// logAcquired emits semaphore.acquired (always, fast or slow path).
func (s semaphore) logAcquired(ctx context.Context, waitMs int64) {
	if !s.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	s.logger.LogAttrs(ctx, slog.LevelDebug, "semaphore.acquired",
		slog.String("op", "semaphore.acquire"),
		slog.Int64("wait_ms", waitMs),
		slog.Int("in_flight", cap(s.ch)-len(s.ch)),
	)
}

// logCancelled emits semaphore.cancelled when ctx fires before a permit lands.
func (s semaphore) logCancelled(ctx context.Context, waitMs int64) {
	if !s.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	s.logger.LogAttrs(ctx, slog.LevelDebug, "semaphore.cancelled",
		slog.String("op", "semaphore.acquire"),
		slog.Int64("wait_ms", waitMs),
	)
}

// release frees one slot. Pair with acquire via defer.
func (s semaphore) release() { <-s.ch }
