// Story 1: UI Adapter logging infrastructure and boot wiring.
//
// RED-phase test for AC-1.6: `(*App).shutdown` must drain every closer
// registered in the new `App.shutdownHooks []func() error` slice. Errors
// returned by hooks must NOT propagate or panic — they are `log.Printf`'d
// only so shutdown completes promptly.
//
// This test will fail to compile until the go-engineer adds the
// `shutdownHooks` field to App and extends `(*App).shutdown` to drain it.
package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStory1_AC6_AppShutdownDrainsHooks asserts every hook in
// App.shutdownHooks is invoked exactly once when shutdown runs, that errors
// returned by a hook do not block subsequent hooks, and that the call never
// panics.
//
// Story 1, AC-1.6: Shutdown closes the log file cleanly.
func TestStory1_AC6_AppShutdownDrainsHooks(t *testing.T) {
	t.Run("each hook invoked exactly once", func(t *testing.T) {
		var calls int32
		a := &App{
			shutdownHooks: []func() error{
				func() error { atomic.AddInt32(&calls, 1); return nil },
			},
		}
		// Other App fields are nil — existing nil-guards in shutdown
		// (assetWatcher / cancel / manager / bridge) skip safely.
		require.NotPanics(t, func() { a.shutdown(context.Background()) },
			"shutdown must not panic when only shutdownHooks is populated")
		assert.Equal(t, int32(1), atomic.LoadInt32(&calls),
			"hook must be invoked exactly once per shutdown")
	})

	t.Run("hook error does not block subsequent hooks", func(t *testing.T) {
		var calledA, calledB, calledC int32
		a := &App{
			shutdownHooks: []func() error{
				func() error {
					atomic.AddInt32(&calledA, 1)
					return errors.New("boom-a")
				},
				func() error {
					atomic.AddInt32(&calledB, 1)
					return nil
				},
				func() error {
					atomic.AddInt32(&calledC, 1)
					return errors.New("boom-c")
				},
			},
		}
		require.NotPanics(t, func() { a.shutdown(context.Background()) },
			"shutdown must not panic when hooks return errors")

		assert.Equal(t, int32(1), atomic.LoadInt32(&calledA),
			"hook A must run even though it returns an error")
		assert.Equal(t, int32(1), atomic.LoadInt32(&calledB),
			"hook B must run after hook A errored — errors must not short-circuit drain")
		assert.Equal(t, int32(1), atomic.LoadInt32(&calledC),
			"hook C must run even though hook A errored before it")
	})

	t.Run("empty shutdownHooks is a no-op", func(t *testing.T) {
		a := &App{shutdownHooks: nil}
		require.NotPanics(t, func() { a.shutdown(context.Background()) },
			"shutdown must accept a nil shutdownHooks slice without panicking")
	})
}
