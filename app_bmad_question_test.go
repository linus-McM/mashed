// Package main contains tests for the App.RespondToQuestion Wails binding.
// Story 2: Backend Response Injection (AC-5, AC-6)
//
// RED Phase: App.RespondToQuestion does not yet exist in app_bmad.go. These
// tests define the contract for the go-engineer. They MUST fail until the
// method is implemented and delegates to (*bmad.Executor).RespondToQuestion
// with a nil-check on a.bmadExecutor.

package main

import (
	"context"
	"errors"
	"testing"

	"mashed/internal/bmad"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newBmadTestApp constructs an App wired to a real bmad.Executor backed by a
// temp-dir Storage so the binding can be exercised against the executor's
// in-memory execution map. Frontend event emission is a no-op for tests.
func newBmadTestApp(t *testing.T) *App {
	t.Helper()
	storage, err := bmad.NewStorage(t.TempDir())
	require.NoError(t, err)

	return &App{
		ctx:          context.Background(),
		bmadStorage:  storage,
		bmadExecutor: bmad.NewExecutor(storage, func(string, interface{}) {}),
	}
}

// ── AC-5: Wails binding delegates to the executor ──
//
// We verify delegation indirectly: when the executor is initialised but the
// requested execID does not exist, the underlying executor returns an error
// wrapping bmad.ErrExecNotFound. If the binding correctly delegates, that
// sentinel propagates out of app.RespondToQuestion. If the binding short-
// circuits (returns "bmad not initialized") delegation did NOT happen and
// this test fails.

func TestStory2_AC5_AppRespondToQuestion_Delegates(t *testing.T) {
	app := newBmadTestApp(t)

	err := app.RespondToQuestion("exec-never-existed", "node-A", "answer")
	require.Error(t, err, "binding must propagate executor errors")
	assert.True(t, errors.Is(err, bmad.ErrExecNotFound),
		"AC-5: binding must delegate to executor and propagate ErrExecNotFound, got %v", err)
	assert.NotContains(t, err.Error(), "not initialized",
		"initialised executor must not return the nil-guard error")
}

// ── AC-6: Nil executor returns a descriptive error ──

func TestStory2_AC6_AppRespondToQuestion_NilExecutor(t *testing.T) {
	app := &App{
		ctx:          context.Background(),
		bmadExecutor: nil,
	}

	err := app.RespondToQuestion("any-exec", "any-node", "any-answer")
	require.Error(t, err, "nil bmadExecutor must return an error")
	assert.Contains(t, err.Error(), "bmad",
		"error should mention 'bmad' to aid debugging")
	assert.Contains(t, err.Error(), "not initialized",
		"error should explicitly say the executor is not initialized")
}
