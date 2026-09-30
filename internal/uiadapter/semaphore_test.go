// Package uiadapter — Story 3: semaphore debug instrumentation
// (uiadapter-logging-3-instrument-cache-and-network).
//
// RED-phase tests for `semaphore.go` log emissions. The file is new in this
// story; prior to Story 2 the semaphore was a bare channel and there was
// nothing to instrument. After Story 3 the struct must emit:
//
//   - `semaphore.wait`      — wait_ms > 0 → permit not immediately available
//   - `semaphore.acquired`  — permit acquired (after a wait, or fast path)
//   - `semaphore.cancelled` — ctx cancelled before a permit landed
//
// Test names embed the AC number per the story's eight acceptance criteria.
package uiadapter

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStory3_AC5_SemaphoreWaitAndAcquired — Story 3, AC-3.5 (wait branch).
//
// Capacity-1 semaphore with the permit held: a second goroutine's acquire
// must block. Releasing the permit ≥ 5ms later must produce a `semaphore.wait`
// record with wait_ms ≥ 5 followed by a `semaphore.acquired` record.
func TestStory3_AC5_SemaphoreWaitAndAcquired(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	sem := newSemaphore(1, logger)

	// Hold the permit on the main goroutine.
	require.True(t, sem.acquire(context.Background()), "fast-path acquire must succeed")

	// Release after a measurable delay so wait_ms > 0.
	const waitFor = 10 * time.Millisecond
	go func() {
		time.Sleep(waitFor)
		sem.release()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	gotPermit := sem.acquire(ctx)
	require.True(t, gotPermit, "second acquire must succeed after release")

	// Give the logger a tick to flush any deferred record.
	sem.release()

	records := decodeRecords(t, buf)
	waits := recordsByMsg(records, "semaphore.wait")
	acquired := recordsByMsg(records, "semaphore.acquired")

	require.NotEmpty(t, waits, "expected at least one semaphore.wait record; got %v", records)
	require.NotEmpty(t, acquired, "expected at least one semaphore.acquired record; got %v", records)

	wait := waits[0]
	assert.Equal(t, "semaphore.acquire", wait["op"], "semaphore.wait must carry op=\"semaphore.acquire\"")
	waitMs, ok := wait["wait_ms"].(float64)
	require.True(t, ok, "semaphore.wait must carry numeric wait_ms; got %v", wait["wait_ms"])
	assert.GreaterOrEqual(t, waitMs, 5.0,
		"semaphore.wait wait_ms (%v) must be ≥ 5 (release happened after ~%v)", waitMs, waitFor)

	acq := acquired[0]
	assert.Equal(t, "semaphore.acquire", acq["op"], "semaphore.acquired must carry op=\"semaphore.acquire\"")
}

// TestStory3_AC5_SemaphoreCancelled — Story 3, AC-3.5 (cancel branch).
//
// Capacity-1 semaphore with the permit held: a second goroutine acquires with
// a ctx that cancels in ~2ms. acquire must return false and emit a
// `semaphore.cancelled` record carrying wait_ms.
func TestStory3_AC5_SemaphoreCancelled(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	sem := newSemaphore(1, logger)

	// Hold the permit; never release within the ctx window.
	require.True(t, sem.acquire(context.Background()))
	t.Cleanup(sem.release)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
	defer cancel()

	var got bool
	var done sync.WaitGroup
	done.Add(1)
	go func() {
		defer done.Done()
		got = sem.acquire(ctx)
	}()
	done.Wait()

	assert.False(t, got, "acquire must return false on ctx cancellation while permit is held")

	records := decodeRecords(t, buf)
	cancelled := recordsByMsg(records, "semaphore.cancelled")
	require.NotEmpty(t, cancelled, "expected at least one semaphore.cancelled record; got %v", records)

	rec := cancelled[0]
	assert.Equal(t, "semaphore.acquire", rec["op"], "semaphore.cancelled must carry op=\"semaphore.acquire\"")
	waitMs, ok := rec["wait_ms"].(float64)
	require.True(t, ok, "semaphore.cancelled must carry numeric wait_ms; got %v", rec["wait_ms"])
	assert.GreaterOrEqual(t, waitMs, 0.0,
		"semaphore.cancelled wait_ms must be ≥ 0; got %v", waitMs)
}
