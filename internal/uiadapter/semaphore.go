package uiadapter

import "context"

// semaphore is a bounded-concurrency gate backed by a buffered channel. A send
// acquires a slot; a receive releases it. The zero value is not useful — use
// newSemaphore.
type semaphore chan struct{}

// newSemaphore returns a semaphore permitting n concurrent holders. n must be
// > 0; callers clamp.
func newSemaphore(n int) semaphore {
	return make(semaphore, n)
}

// acquire attempts a non-blocking send first so a pre-canceled context never
// races the fast path. Go's select randomises among ready cases; without the
// two-phase attempt, a pre-canceled caller on an empty sem could flip between
// saturated and canceled reasons.
func (s semaphore) acquire(ctx context.Context) bool {
	select {
	case s <- struct{}{}:
		return true
	default:
	}
	select {
	case s <- struct{}{}:
		return true
	case <-ctx.Done():
		return false
	}
}

// release frees one slot. Pair with acquire via defer.
func (s semaphore) release() { <-s }
