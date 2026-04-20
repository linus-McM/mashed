// Package main contains tests for the App.GetBmadCurrentExecution restore-on-mount
// re-emit path.
// Story bmad-interactive-05: Persistence extension + restore-on-mount re-emit.
//
// RED Phase: These tests define the contract for the pending-prompt re-emit
// behaviour that GetBmadCurrentExecution must exhibit after S5 is implemented.
// They MUST fail (or succeed trivially) until:
//   - WorkflowExecution gains a Version field.
//   - GetBmadCurrentExecution emits bmad:node:awaiting_input for each PendingPrompt.
//   - The emit is deferred to a goroutine so it fires AFTER the method returns.
//
// Wails runtime.EventsEmit is not mockable in unit tests because it requires
// a live Wails context. Instead, we use the App-level emitHook indirection:
//
//   var appEmitHook func(eventName string, data ...any)
//
// This hook is declared below for test-only use. The go-engineer wires the
// production GetBmadCurrentExecution emit call through appEmitHook when non-nil,
// falling back to runtime.EventsEmit when nil. Tests set the hook before calling
// GetBmadCurrentExecution.
//
// Seeding snapshots:
// GetBmadCurrentExecution calls a.bmadExecutor.GetCurrentExecution which scans
// the in-memory executions map. To test the restore-from-disk path, we need
// either:
//   (a) An executor method that loads from disk (future), or
//   (b) Register a pre-built execState into the executor's map directly (white-box — not
//       possible from package main), or
//   (c) Call (*Executor).StartWorkflow then forcibly write a snapshot with PendingPrompts
//       (complex), or
//   (d) Use a fake executor interface.
//
// Decision: we use approach (d) — introduce a BmadExecutor interface in app.go (or
// a thin shim in the test file) so tests can inject a fake that returns a pre-built
// WorkflowExecution. This is the simplest path that does not require white-box access.
//
// If the go-engineer chooses not to introduce an interface, the alternative is to
// use package-level var appEmitHook and seed the executor's in-memory map via an
// exported helper. Either approach is acceptable — the tests document the contract.
package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"mashed/internal/bmad"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// appEmitHook is declared canonically in app_bmad.go. Tests set it via
// hookAppEmit and t.Cleanup restores the nil default.

// capturedAppEvent holds one intercepted event.
type capturedAppEvent struct {
	name    string
	payload any
}

// hookAppEmit sets appEmitHook and returns a snapshot function + cleanup.
// Events are appended in the order received (including from background goroutines).
func hookAppEmit(t *testing.T) func() []capturedAppEvent {
	t.Helper()
	var mu sync.Mutex
	var events []capturedAppEvent
	appEmitHook = func(name string, data ...any) {
		mu.Lock()
		defer mu.Unlock()
		var payload any
		if len(data) > 0 {
			payload = data[0]
		}
		events = append(events, capturedAppEvent{name: name, payload: payload})
	}
	t.Cleanup(func() { appEmitHook = nil })
	return func() []capturedAppEvent {
		mu.Lock()
		defer mu.Unlock()
		out := make([]capturedAppEvent, len(events))
		copy(out, events)
		return out
	}
}

// newBmadTestAppWithExec constructs an App with a real executor and optionally
// seeds a snapshot on disk so GetBmadCurrentExecution can find it.
// HOME is overridden to tmpDir so snapshots land in isolation.
func newBmadTestAppWithExec(t *testing.T, tmpHome string) (*App, *bmad.Executor) {
	t.Helper()
	storage, err := bmad.NewStorage(t.TempDir())
	require.NoError(t, err)
	exec := bmad.NewExecutor(storage, func(string, interface{}) {})
	app := &App{
		ctx:          context.Background(),
		bmadStorage:  storage,
		bmadExecutor: exec,
	}
	return app, exec
}

// seedSnapshotOnDisk writes an execution.json into the ~/.mashed/workflows/{execID}/
// directory so GetCurrentExecution's on-disk load path picks it up.
// If GetCurrentExecution only reads in-memory state, the on-disk seed is a
// no-op and the test correctly fails (documents the gap for the go-engineer).
func seedSnapshotOnDisk(t *testing.T, tmpHome string, exec *bmad.WorkflowExecution) {
	t.Helper()
	raw, err := json.MarshalIndent(exec, "", "  ")
	require.NoError(t, err)
	dir := filepath.Join(tmpHome, ".mashed", "workflows", exec.ID)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	snapshotPath := filepath.Join(dir, "execution.json")
	require.NoError(t, os.WriteFile(snapshotPath, raw, 0o644))
}

// ── AC-2: GetBmadCurrentExecution re-emits bmad:node:awaiting_input ──────────

// TestGetBmadCurrentExecutionReEmitsAwaitingInput (AC-2)
// Given a snapshot on disk with PendingPrompts=[{NodeID:"n1", InputID:"topic"}]
// When (*App).GetBmadCurrentExecution(repoPath) is called
// Then the returned execution has the pending prompt
// And within 500ms a bmad:node:awaiting_input event fires with matching payload.
//
// RED phase: this test fails because GetBmadCurrentExecution does not yet
// re-emit pending prompts and/or GetCurrentExecution does not read from disk.
func TestGetBmadCurrentExecutionReEmitsAwaitingInput(t *testing.T) {
	// NOT parallel: sets appEmitHook global and HOME env var.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	getSnap := hookAppEmit(t)

	app, executor := newBmadTestAppWithExec(t, tmpHome)

	const execID = "exec-reemit-test"
	const repoPath = "/tmp/test-repo"
	pendingPrompt := bmad.PendingPrompt{
		NodeID:    "n1",
		InputID:   "topic",
		Prompt:    "Enter a topic",
		Shape:     bmad.ShapeFree,
		Round:     1,
		CreatedAt: time.Now().Unix(),
		PromptID:  "abc123def456abcd",
	}
	seedExec := &bmad.WorkflowExecution{
		ID:         execID,
		WorkflowID: "wf-test",
		RepoPath:   repoPath,
		Status:     bmad.ExecRunning,
		Nodes: []bmad.WorkflowNode{
			{ID: "n1", ProcessID: "test-proc", Status: bmad.NodeAwaitingInput},
		},
		StartedAt:      time.Now().UTC().Format(time.RFC3339),
		PendingPrompts: []bmad.PendingPrompt{pendingPrompt},
		Version:        2,
	}
	seedSnapshotOnDisk(t, tmpHome, seedExec)

	// Also register the execution in the executor's in-memory map so
	// GetCurrentExecution finds it (until the disk-load path is implemented).
	// If InjectExecution doesn't exist yet, this is a compile error — RED phase.
	// The go-engineer adds: func (e *Executor) InjectExecution(state *WorkflowExecution)
	// for test seeding.
	_ = executor // suppress "declared but not used" — seeding happens via disk above.

	result, err := app.GetBmadCurrentExecution(repoPath)
	require.NoError(t, err, "GetBmadCurrentExecution must not error when executor is initialized")

	// If on-disk loading is not yet implemented, result may be nil.
	// Document both cases clearly.
	if result == nil {
		t.Log("RED phase gap: GetCurrentExecution does not load from disk yet; " +
			"result is nil. The go-engineer must implement disk-load in CurrentExecution.")
		t.Log("Checking that no awaiting_input event was erroneously fired for nil result...")
		// Give 500ms for any erroneous events.
		time.Sleep(200 * time.Millisecond)
		snap := getSnap()
		for _, ev := range snap {
			assert.NotEqual(t, "bmad:node:awaiting_input", ev.name,
				"must not emit awaiting_input when exec is nil")
		}
		// Fail the test so the go-engineer knows the disk-load path is needed.
		t.Fatal("RED: GetBmadCurrentExecution returned nil — disk-load path not implemented")
	}

	// Returned execution must contain the pending prompt.
	require.NotEmpty(t, result.PendingPrompts,
		"returned execution must contain PendingPrompts from snapshot")
	assert.Equal(t, "n1", result.PendingPrompts[0].NodeID)
	assert.Equal(t, "topic", result.PendingPrompts[0].InputID)

	// Within 500ms, a bmad:node:awaiting_input event must be emitted.
	deadline := time.Now().Add(500 * time.Millisecond)
	var gotAwaitingEvent bool
	for time.Now().Before(deadline) {
		snap := getSnap()
		for _, ev := range snap {
			if ev.name == "bmad:node:awaiting_input" {
				gotAwaitingEvent = true
				// Verify payload matches the pending prompt.
				if pp, ok := ev.payload.(bmad.PendingPrompt); ok {
					assert.Equal(t, "n1", pp.NodeID, "awaiting_input payload NodeID must be 'n1'")
					assert.Equal(t, "topic", pp.InputID, "awaiting_input payload InputID must be 'topic'")
				}
				break
			}
		}
		if gotAwaitingEvent {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	assert.True(t, gotAwaitingEvent,
		"bmad:node:awaiting_input must be emitted within 500ms of GetBmadCurrentExecution returning")
}

// ── AC-4: No emit for executions with empty PendingPrompts ────────────────────

// TestGetBmadCurrentExecutionNoEmitForEmptyPending (AC-4)
// Given a snapshot with no PendingPrompts.
// When GetBmadCurrentExecution is called.
// Then no bmad:node:awaiting_input event fires within 500ms.
func TestGetBmadCurrentExecutionNoEmitForEmptyPending(t *testing.T) {
	// NOT parallel: sets appEmitHook and HOME.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	getSnap := hookAppEmit(t)

	app, _ := newBmadTestAppWithExec(t, tmpHome)

	const execID = "exec-no-pending"
	const repoPath = "/tmp/test-repo-no-pending"
	seedExec := &bmad.WorkflowExecution{
		ID:             execID,
		WorkflowID:     "wf-test",
		RepoPath:       repoPath,
		Status:         bmad.ExecRunning,
		Nodes:          []bmad.WorkflowNode{{ID: "n1", ProcessID: "test-proc", Status: bmad.NodeRunning}},
		StartedAt:      time.Now().UTC().Format(time.RFC3339),
		PendingPrompts: nil, // no pending prompts
	}
	seedSnapshotOnDisk(t, tmpHome, seedExec)

	_, err := app.GetBmadCurrentExecution(repoPath)
	require.NoError(t, err)

	// Wait 500ms and verify no awaiting_input event fired.
	time.Sleep(500 * time.Millisecond)
	snap := getSnap()
	for _, ev := range snap {
		assert.NotEqual(t, "bmad:node:awaiting_input", ev.name,
			"must not emit awaiting_input when PendingPrompts is empty")
	}
}

// ── AC-7: Terminal executions return nil and fire no events ───────────────────

// TestGetBmadCurrentExecutionTerminalExecutionReturnsNil (AC-7)
// Given a snapshot with status=ExecFailed.
// When GetBmadCurrentExecution is called.
// Then nil is returned and no awaiting_input event fires.
func TestGetBmadCurrentExecutionTerminalExecutionReturnsNil(t *testing.T) {
	// NOT parallel: sets appEmitHook and HOME.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	getSnap := hookAppEmit(t)

	app, _ := newBmadTestAppWithExec(t, tmpHome)

	const execID = "exec-terminal"
	const repoPath = "/tmp/test-repo-terminal"
	seedExec := &bmad.WorkflowExecution{
		ID:         execID,
		WorkflowID: "wf-test",
		RepoPath:   repoPath,
		Status:     bmad.ExecFailed, // terminal
		Nodes: []bmad.WorkflowNode{
			{ID: "n1", ProcessID: "test-proc", Status: bmad.NodeFailed},
		},
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		// PendingPrompts set before abort — should NOT be re-emitted.
		PendingPrompts: []bmad.PendingPrompt{
			{NodeID: "n1", InputID: "topic", Round: 1, PromptID: "deaddeaddeaddead"},
		},
		Version: 2,
	}
	seedSnapshotOnDisk(t, tmpHome, seedExec)

	result, err := app.GetBmadCurrentExecution(repoPath)
	require.NoError(t, err, "GetBmadCurrentExecution must not error on terminal execution")

	// Terminal executions must return nil (existing filter behavior).
	// If result is non-nil the go-engineer's filter is missing.
	assert.Nil(t, result,
		"GetBmadCurrentExecution must return nil for terminal (ExecFailed) executions")

	// Allow 500ms for any erroneous events.
	time.Sleep(500 * time.Millisecond)
	snap := getSnap()
	for _, ev := range snap {
		assert.NotEqual(t, "bmad:node:awaiting_input", ev.name,
			"must not emit awaiting_input for terminal executions")
	}
}
