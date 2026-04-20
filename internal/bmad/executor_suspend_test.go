// Package bmad contains tests for the suspension primitive.
// Story bmad-interactive-03: Suspension primitive + RespondToInput Wails binding.
//
// RED Phase: These tests define the expected contract for suspendForSpec,
// per-waiter channels, hashPendingPrompt, upsertPrompt, removePrompt,
// findPendingPrompt, and findInputSpec. They MUST fail until the go-engineer
// implements the feature (GREEN phase).
//
// Tests use the white-box package (package bmad) to access unexported helpers.
package bmad

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Fixture helpers ────────────────────────────────────────────────────────────

// newSuspendState builds an execState with a single interactive node that has
// one user-input InputSpec. nodeIndex is also returned because it is a local
// variable in runDynamic and passed into executeInteractiveNode.
func newSuspendState(nodeID, processID string, spec InputSpec) (*execState, map[string]int) {
	nodes := []WorkflowNode{
		{
			ID:        nodeID,
			ProcessID: processID,
			Label:     "Suspend Node",
			Status:    NodePending,
			NodeType:  NodeTypeProcess,
			InputSpecs: []InputSpec{spec},
		},
	}
	state, nodeIndex := newSessionState(nodes, nil)
	return state, nodeIndex
}

// hookEvents installs testEventHook and returns a snapshot function and cleanup.
// The snapshot function returns a copy of all captured events.
func hookEvents(t *testing.T) func() []struct{ name string; payload interface{} } {
	t.Helper()
	var mu sync.Mutex
	var captured []struct{ name string; payload interface{} }
	testEventHook = func(event string, payload interface{}) {
		mu.Lock()
		captured = append(captured, struct{ name string; payload interface{} }{event, payload})
		mu.Unlock()
	}
	t.Cleanup(func() { testEventHook = nil })
	return func() []struct{ name string; payload interface{} } {
		mu.Lock()
		defer mu.Unlock()
		out := make([]struct{ name string; payload interface{} }, len(captured))
		copy(out, captured)
		return out
	}
}

// eventsNamed filters a snapshot for entries matching name.
func eventsNamed(snap []struct{ name string; payload interface{} }, name string) []interface{} {
	var out []interface{}
	for _, ev := range snap {
		if ev.name == name {
			out = append(out, ev.payload)
		}
	}
	return out
}

// registerTestProcess adds a ProcessDef to testRegistry and registers a cleanup
// that removes it. Only call from non-parallel tests (testRegistry is global).
func registerTestProcess(t *testing.T, pd ProcessDef) {
	t.Helper()
	testRegistry = append(testRegistry, pd)
	t.Cleanup(func() { testRegistry = testRegistry[:len(testRegistry)-1] })
}

// pollForStatus polls state.exec.Nodes[nodeIndex[nodeID]].Status until it equals
// want or the deadline is reached. Returns true when the status matches.
func pollForStatus(state *execState, nodeIndex map[string]int, nodeID string, want WorkflowNodeStatus, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state.mu.Lock()
		idx := nodeIndex[nodeID]
		status := state.exec.Nodes[idx].Status
		state.mu.Unlock()
		if status == want {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

// ── AC-1: suspendForSpec enters NodeAwaitingInput ─────────────────────────────

// TestSuspendForSpecEntersAwaitingInput (AC-1)
// Given an interactive node with a required InputFromUser spec "topic" ShapeFree.
// When suspendForSpec is called (in a goroutine, since it blocks).
// Then the node status becomes NodeAwaitingInput within 500ms.
// And PendingPrompts has one entry with correct fields including 16-char hex PromptID.
// And a bmad:node:awaiting_input event fires with a PendingPrompt payload.
func TestSuspendForSpecEntersAwaitingInput(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-s3-suspend-awaiting"
	spec := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: true,
		Prompt:   "Enter a topic",
	}
	registerTestProcess(t, ProcessDef{
		ID:         processID,
		Name:       "Test S3 Suspend Awaiting",
		Mode:       InteractGuided,
		SkillName:  "test-skill",
		InputSpecs: []InputSpec{spec},
		OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToMemory}},
	})

	getSnap := hookEvents(t)

	state, nodeIndex := newSuspendState(nodeID, processID, spec)
	state.exec.RepoPath = t.TempDir()

	e := NewExecutor(nil, func(name string, data interface{}) {
		if testEventHook != nil {
			testEventHook(name, data)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		// suspendForSpec does not exist yet — RED phase: this will fail to compile.
		errCh <- e.suspendForSpec(ctx, state, nodeIndex, nodeID, 1, spec)
	}()

	// Assert NodeAwaitingInput within 500ms.
	ok := pollForStatus(state, nodeIndex, nodeID, NodeAwaitingInput, 500*time.Millisecond)
	assert.True(t, ok, "node must reach NodeAwaitingInput within 500ms")

	// Assert PendingPrompts state.
	state.mu.Lock()
	pendingPrompts := make([]PendingPrompt, len(state.exec.PendingPrompts))
	copy(pendingPrompts, state.exec.PendingPrompts)
	state.mu.Unlock()

	require.Len(t, pendingPrompts, 1, "PendingPrompts must have exactly 1 entry")
	pp := pendingPrompts[0]
	assert.Equal(t, nodeID, pp.NodeID, "PendingPrompt.NodeID")
	assert.Equal(t, spec.ID, pp.InputID, "PendingPrompt.InputID")
	assert.Equal(t, 1, pp.Round, "PendingPrompt.Round must be 1")
	assert.Len(t, pp.PromptID, 16, "PromptID must be exactly 16 chars")
	_, hexErr := hex.DecodeString(pp.PromptID)
	assert.NoError(t, hexErr, "PromptID must be valid hex")
	assert.Greater(t, pp.CreatedAt, int64(0), "CreatedAt must be a positive Unix timestamp")

	// Assert bmad:node:awaiting_input event fired.
	snap := getSnap()
	awaitEvts := eventsNamed(snap, EventAwaitingInput)
	require.Len(t, awaitEvts, 1, "bmad:node:awaiting_input must fire exactly once")
	awaitPayload, isPP := awaitEvts[0].(PendingPrompt)
	require.True(t, isPP, "awaiting_input payload must be a PendingPrompt, got %T", awaitEvts[0])
	assert.Equal(t, nodeID, awaitPayload.NodeID)
	assert.Equal(t, spec.ID, awaitPayload.InputID)

	// Release to prevent goroutine leak.
	t.Cleanup(func() {
		cancel()
		select {
		case <-errCh:
		case <-time.After(2 * time.Second):
		}
	})
}

// ── AC-7: context cancellation aborts suspension ──────────────────────────────

// TestSuspendForSpecCtxCancellationAborts (AC-7)
// Given a node suspended on a required user input.
// When the execution context is cancelled.
// Then suspendForSpec returns context.Canceled.
// And bmad:node:aborted is emitted with reason containing "workflow stopped".
func TestSuspendForSpecCtxCancellationAborts(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-s3-suspend-abort"
	spec := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: true,
	}
	registerTestProcess(t, ProcessDef{
		ID:         processID,
		Name:       "Test S3 Suspend Abort",
		Mode:       InteractGuided,
		SkillName:  "test-skill",
		InputSpecs: []InputSpec{spec},
		OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToMemory}},
	})

	getSnap := hookEvents(t)

	state, nodeIndex := newSuspendState(nodeID, processID, spec)
	state.exec.RepoPath = t.TempDir()

	e := NewExecutor(nil, func(name string, data interface{}) {
		if testEventHook != nil {
			testEventHook(name, data)
		}
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- e.suspendForSpec(ctx, state, nodeIndex, nodeID, 1, spec)
	}()

	// Wait for suspension to register.
	pollForStatus(state, nodeIndex, nodeID, NodeAwaitingInput, 500*time.Millisecond)

	// Cancel context — simulates StopBmadWorkflow.
	cancel()

	select {
	case err := <-errCh:
		assert.ErrorIs(t, err, context.Canceled,
			"suspendForSpec must return context.Canceled when ctx is cancelled")
	case <-time.After(2 * time.Second):
		t.Fatal("suspendForSpec did not return within 2s after context cancellation")
	}

	// Assert bmad:node:aborted event fired with reason "workflow stopped".
	snap := getSnap()
	abortedEvts := eventsNamed(snap, EventAborted)
	require.NotEmpty(t, abortedEvts, "bmad:node:aborted must fire on context cancellation")

	// Verify reason field contains "workflow stopped".
	// Accept map[string]string or map[string]interface{} — the go-engineer picks the shape.
	reasonOK := false
	switch p := abortedEvts[0].(type) {
	case map[string]string:
		reasonOK = len(p["reason"]) > 0 && containsSubstr(p["reason"], "workflow stopped")
	case map[string]interface{}:
		reason, _ := p["reason"].(string)
		reasonOK = containsSubstr(reason, "workflow stopped")
	default:
		t.Logf("aborted payload type %T: %+v", abortedEvts[0], abortedEvts[0])
		reasonOK = fmt.Sprintf("%+v", abortedEvts[0]) != ""
	}
	assert.True(t, reasonOK, "aborted payload reason must contain 'workflow stopped'")
}

// containsSubstr is a substring check helper (avoids importing strings in the
// test body where it could shadow other declarations).
func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

// ── hashPendingPrompt ─────────────────────────────────────────────────────────

// TestHashPendingPromptDeterminism
// hashPendingPrompt(nodeID, inputID, round) must be deterministic, 16 hex chars,
// and differ for different round or nodeID inputs.
func TestHashPendingPromptDeterminism(t *testing.T) {
	t.Parallel()

	// hashPendingPrompt does not exist yet — RED phase: will fail to compile.
	h1 := hashPendingPrompt("n1", "topic", 1)
	h2 := hashPendingPrompt("n1", "topic", 1)
	assert.Equal(t, h1, h2, "hashPendingPrompt must be deterministic for same inputs")
	assert.Len(t, h1, 16, "hash must be exactly 16 characters")
	_, err := hex.DecodeString(h1)
	assert.NoError(t, err, "hash must be valid hex")

	hRound2 := hashPendingPrompt("n1", "topic", 2)
	assert.NotEqual(t, h1, hRound2, "different round must produce different hash")

	hNodeB := hashPendingPrompt("n2", "topic", 1)
	assert.NotEqual(t, h1, hNodeB, "different nodeID must produce different hash")
}

// ── upsertPrompt / removePrompt ───────────────────────────────────────────────

// TestUpsertAndRemovePromptHelpers
// Covers the full contract for upsertPrompt and removePrompt slice helpers.
func TestUpsertAndRemovePromptHelpers(t *testing.T) {
	t.Parallel()

	pp1 := PendingPrompt{NodeID: "n1", InputID: "topic", Round: 1, PromptID: "aaa"}
	pp2 := PendingPrompt{NodeID: "n1", InputID: "method", Round: 1, PromptID: "bbb"}

	// Upsert into empty slice → 1 entry.
	// upsertPrompt does not exist yet — RED phase: will fail to compile.
	slice := upsertPrompt(nil, pp1)
	assert.Len(t, slice, 1)
	assert.Equal(t, pp1, slice[0])

	// Upsert same (nodeID, inputID) → replace, length stays 1.
	pp1v2 := PendingPrompt{NodeID: "n1", InputID: "topic", Round: 2, PromptID: "ccc"}
	slice = upsertPrompt(slice, pp1v2)
	assert.Len(t, slice, 1, "upsert same key must replace, not append")
	assert.Equal(t, pp1v2, slice[0], "upserted entry must replace old entry")

	// Upsert different (nodeID, inputID) → length 2.
	slice = upsertPrompt(slice, pp2)
	assert.Len(t, slice, 2, "upsert different key must append")

	// Remove pp1v2 → length 1.
	// removePrompt does not exist yet — RED phase.
	slice = removePrompt(slice, "n1", "topic")
	assert.Len(t, slice, 1, "remove must reduce slice length")
	assert.Equal(t, pp2, slice[0], "remaining entry must be pp2")

	// Remove on not-found → unchanged length.
	before := len(slice)
	slice = removePrompt(slice, "n99", "nope")
	assert.Equal(t, before, len(slice), "remove of absent key must not mutate length")
}

// ── findPendingPrompt ─────────────────────────────────────────────────────────

// TestFindPendingPrompt
// Slice with 3 entries: find existing returns (entry, true), not-found returns (zero, false).
func TestFindPendingPrompt(t *testing.T) {
	t.Parallel()

	prompts := []PendingPrompt{
		{NodeID: "n1", InputID: "alpha", PromptID: "aaa"},
		{NodeID: "n1", InputID: "beta", PromptID: "bbb"},
		{NodeID: "n2", InputID: "alpha", PromptID: "ccc"},
	}

	// findPendingPrompt does not exist yet — RED phase: will fail to compile.
	got, ok := findPendingPrompt(prompts, "n1", "beta")
	require.True(t, ok, "findPendingPrompt must return true for existing (n1, beta)")
	assert.Equal(t, "bbb", got.PromptID)

	got2, ok2 := findPendingPrompt(prompts, "n3", "alpha")
	assert.False(t, ok2, "findPendingPrompt must return false for absent (n3, alpha)")
	assert.Equal(t, PendingPrompt{}, got2, "zero value must be returned when not found")
}

// ── findInputSpec ─────────────────────────────────────────────────────────────

// TestFindInputSpec
// ProcessDef with 3 InputSpecs: find by ID existing, not-found returns zero + false.
func TestFindInputSpec(t *testing.T) {
	t.Parallel()

	proc := ProcessDef{
		ID: "p1",
		InputSpecs: []InputSpec{
			{ID: "topic", Shape: ShapeFree},
			{ID: "method", Shape: ShapeChoice},
			{ID: "approval", Shape: ShapeApproval},
		},
	}

	// findInputSpec does not exist yet — RED phase: will fail to compile.
	got, ok := findInputSpec(proc, "method")
	require.True(t, ok)
	assert.Equal(t, ShapeChoice, got.Shape)

	got2, ok2 := findInputSpec(proc, "not-declared")
	assert.False(t, ok2)
	assert.Equal(t, InputSpec{}, got2)
}
