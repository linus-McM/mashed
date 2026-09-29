// Package bmad contains tests for the RespondToInput executor method.
// Story bmad-interactive-03: Suspension primitive + RespondToInput Wails binding.
//
// RED Phase: These tests define the expected contract for (*Executor).RespondToInput,
// ErrInvalidInput, ErrPathOutsideRepo, ErrNoPendingPrompt, ErrUnknownInput,
// ErrStalePrompt, and ErrInvalidRegistryRef sentinels. They MUST fail until the
// go-engineer implements the feature (GREEN phase).
package bmad

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Harness for respond tests ─────────────────────────────────────────────────

// suspendState holds the infrastructure needed to set up a suspension and then
// call RespondToInput on it.
type suspendHarness struct {
	e         *Executor
	state     *execState
	nodeIndex map[string]int
	execID    string
	snap      func() []struct{ name string; payload interface{} }
	cancel    context.CancelFunc
	errCh     chan error
	nodeID    string
	spec      InputSpec
	processID string
}

// setupSuspension creates an executor, registers a test process, puts the
// execution into the executor's map, and starts suspendForSpec in a goroutine.
// The caller must call s.waitForAwaiting() before calling RespondToInput.
func setupSuspension(t *testing.T, nodeID, processID string, spec InputSpec) *suspendHarness {
	t.Helper()

	registerTestProcess(t, ProcessDef{
		ID:         processID,
		Name:       "Test S3 Respond " + processID,
		Mode:       InteractGuided,
		SkillName:  "test-skill",
		InputSpecs: []InputSpec{spec},
		OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToMemory}},
	})

	getSnap := hookEvents(t)

	state, nodeIndex := newSuspendState(nodeID, processID, spec)
	state.exec.RepoPath = t.TempDir()
	execID := state.exec.ID

	e := NewExecutor(nil, func(name string, data interface{}) {})

	// Register the state in the executor's executions map so RespondToInput
	// can find it via getState(execID).
	e.mu.Lock()
	e.executions[execID] = state
	e.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- e.suspendForSpec(ctx, state, nodeIndex, nodeID, 1, spec, "")
	}()

	h := &suspendHarness{
		e:         e,
		state:     state,
		nodeIndex: nodeIndex,
		execID:    execID,
		snap:      getSnap,
		cancel:    cancel,
		errCh:     errCh,
		nodeID:    nodeID,
		spec:      spec,
		processID: processID,
	}
	t.Cleanup(func() { h.cancel(); select { case <-h.errCh: case <-time.After(2 * time.Second): } })
	return h
}

// waitForAwaiting polls until NodeAwaitingInput (or times out after 500ms).
func (h *suspendHarness) waitForAwaiting(t *testing.T) {
	t.Helper()
	ok := pollForStatus(h.state, h.nodeIndex, h.nodeID, NodeAwaitingInput, 500*time.Millisecond)
	require.True(t, ok, "node must reach NodeAwaitingInput before calling RespondToInput")
}

// nodeStatus reads the current node status under state.mu.
func (h *suspendHarness) nodeStatus() WorkflowNodeStatus {
	h.state.mu.Lock()
	defer h.state.mu.Unlock()
	idx := h.nodeIndex[h.nodeID]
	return h.state.exec.Nodes[idx].Status
}

// nodeInputs reads NodeInputs[nodeID] under state.mu.
func (h *suspendHarness) nodeInputs() map[string]string {
	h.state.mu.Lock()
	defer h.state.mu.Unlock()
	if h.state.exec.NodeInputs == nil {
		return nil
	}
	src := h.state.exec.NodeInputs[h.nodeID]
	out := make(map[string]string, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

// nodeInputHistory reads NodeInputHistory[nodeID] under state.mu.
func (h *suspendHarness) nodeInputHistory() []NodeInputEntry {
	h.state.mu.Lock()
	defer h.state.mu.Unlock()
	if h.state.exec.NodeInputHistory == nil {
		return nil
	}
	src := h.state.exec.NodeInputHistory[h.nodeID]
	out := make([]NodeInputEntry, len(src))
	copy(out, src)
	return out
}

// pendingPrompts reads PendingPrompts under state.mu.
func (h *suspendHarness) pendingPrompts() []PendingPrompt {
	h.state.mu.Lock()
	defer h.state.mu.Unlock()
	out := make([]PendingPrompt, len(h.state.exec.PendingPrompts))
	copy(out, h.state.exec.PendingPrompts)
	return out
}

// ── AC-2: RespondToInput happy path ──────────────────────────────────────────

// TestRespondToInputHappyPath (AC-2)
// Given a node suspended on input "topic" ShapeFree.
// When RespondToInput is called with a valid value.
// Then returns nil; NodeInputs updated; NodeInputHistory has entry with Round=1;
// node status returns to NodeRunning; PendingPrompts empty;
// bmad:node:input_resolved fires with valueHash (no raw value).
func TestRespondToInputHappyPath(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-s3-respond-happy"
	spec := InputSpec{ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true}
	value := "my idea"

	h := setupSuspension(t, nodeID, processID, spec)
	h.waitForAwaiting(t)

	// RespondToInput does not exist yet — RED phase: will fail to compile.
	err := h.e.RespondToInput(h.execID, nodeID, "topic", value)
	require.NoError(t, err, "RespondToInput must return nil for valid ShapeFree value")

	// Wait for node to return to NodeRunning.
	ok := pollForStatus(h.state, h.nodeIndex, nodeID, NodeRunning, 500*time.Millisecond)
	assert.True(t, ok, "node must return to NodeRunning after RespondToInput")

	// NodeInputs updated.
	inputs := h.nodeInputs()
	assert.Equal(t, value, inputs["topic"], "NodeInputs[nodeID][topic] must equal submitted value")

	// NodeInputHistory has entry with Round=1.
	history := h.nodeInputHistory()
	require.Len(t, history, 1, "NodeInputHistory must have exactly one entry")
	assert.Equal(t, "topic", history[0].InputID)
	assert.Equal(t, 1, history[0].Round)
	assert.Equal(t, value, history[0].Value)

	// PendingPrompts must be empty.
	pp := h.pendingPrompts()
	assert.Empty(t, pp, "PendingPrompts must be empty after successful RespondToInput")

	// bmad:node:input_resolved event must fire with valueHash.
	snap := h.snap()
	resolvedEvts := eventsNamed(snap, EventInputResolved)
	require.Len(t, resolvedEvts, 1, "bmad:node:input_resolved must fire exactly once")

	// Marshal payload to JSON to verify no raw value present.
	payloadJSON, err2 := json.Marshal(resolvedEvts[0])
	require.NoError(t, err2)
	payloadStr := string(payloadJSON)

	assert.NotContains(t, payloadStr, value,
		"bmad:node:input_resolved payload must NOT contain raw value")

	// valueHash must be SHA-256[:16] of the value.
	wantHash := sha256hex16(value)
	assert.Contains(t, payloadStr, wantHash,
		"bmad:node:input_resolved payload must contain correct valueHash")

	// Drain suspendForSpec goroutine.
	select {
	case <-h.errCh:
	case <-time.After(2 * time.Second):
		t.Error("suspendForSpec goroutine did not exit within 2s")
	}
}

// ── AC-3: Invalid choice stays awaiting ──────────────────────────────────────

// TestRespondToInputInvalidChoiceKeepsAwaiting (AC-3)
// Given a node suspended on ShapeChoice Options=["a","b","c"].
// When RespondToInput is called with "z" (not in options).
// Then returns error wrapping ErrInvalidInput; status still NodeAwaitingInput;
// PendingPrompts still has entry; bmad:node:input_invalid fired with reason.
func TestRespondToInputInvalidChoiceKeepsAwaiting(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-s3-respond-choice-invalid"
	spec := InputSpec{
		ID:       "approach",
		Source:   InputFromUser,
		Shape:    ShapeChoice,
		Required: true,
		Options:  []string{"a", "b", "c"},
	}

	h := setupSuspension(t, nodeID, processID, spec)
	h.waitForAwaiting(t)

	err := h.e.RespondToInput(h.execID, nodeID, "approach", "z")
	require.Error(t, err, "RespondToInput must return error for invalid choice")
	// ErrInvalidInput does not exist yet — RED phase.
	assert.True(t, errors.Is(err, ErrInvalidInput),
		"error must wrap ErrInvalidInput, got: %v", err)

	// Node must still be NodeAwaitingInput.
	assert.Equal(t, NodeAwaitingInput, h.nodeStatus(),
		"node must remain NodeAwaitingInput after invalid input")

	// PendingPrompts still has entry.
	pp := h.pendingPrompts()
	require.Len(t, pp, 1, "PendingPrompts must still have 1 entry after invalid input")

	// bmad:node:input_invalid must fire with reason mentioning allowed options.
	snap := h.snap()
	invalidEvts := eventsNamed(snap, EventInputInvalid)
	require.NotEmpty(t, invalidEvts, "bmad:node:input_invalid must fire after invalid input")
	payloadJSON, _ := json.Marshal(invalidEvts[0])
	payloadStr := string(payloadJSON)
	// Reason should mention at least one allowed option.
	assert.Contains(t, payloadStr, "a",
		"bmad:node:input_invalid payload must mention allowed options")
}

// ── AC-4: ShapeFile path traversal rejection ──────────────────────────────────

// TestRespondToInputShapeFilePathTraversalRejected (AC-4)
// Given a node suspended on ShapeFile.
// When RespondToInput is called with "/etc/passwd".
// Then returns error wrapping ErrPathOutsideRepo; status still NodeAwaitingInput;
// bmad:node:input_invalid fired with reason containing "path outside repository root".
// Happy path: value inside repoPath is accepted.
func TestRespondToInputShapeFilePathTraversalRejected(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-s3-respond-file-traversal"
	spec := InputSpec{
		ID:       "targetfile",
		Source:   InputFromUser,
		Shape:    ShapeFile,
		Required: true,
	}

	h := setupSuspension(t, nodeID, processID, spec)
	h.waitForAwaiting(t)

	// Attempt path traversal.
	err := h.e.RespondToInput(h.execID, nodeID, "targetfile", "/etc/passwd")
	require.Error(t, err, "RespondToInput must reject path outside repo root")
	// ErrPathOutsideRepo does not exist yet — RED phase.
	assert.True(t, errors.Is(err, ErrPathOutsideRepo),
		"error must wrap ErrPathOutsideRepo, got: %v", err)
	assert.Equal(t, NodeAwaitingInput, h.nodeStatus(),
		"node must remain NodeAwaitingInput after path traversal rejection")

	snap := h.snap()
	invalidEvts := eventsNamed(snap, EventInputInvalid)
	require.NotEmpty(t, invalidEvts, "bmad:node:input_invalid must fire for path traversal")
	payloadJSON, _ := json.Marshal(invalidEvts[0])
	assert.Contains(t, string(payloadJSON), "path outside repository root",
		"input_invalid reason must mention 'path outside repository root'")

	// Happy path: a path inside repoPath should be accepted.
	// Note: The state was constructed with t.TempDir() as RepoPath.
	repoPath := h.state.exec.RepoPath
	validPath := repoPath + "/somefile.md"
	err2 := h.e.RespondToInput(h.execID, nodeID, "targetfile", validPath)
	assert.NoError(t, err2,
		"RespondToInput must accept a path inside the repo root, got: %v", err2)
}

// ── AC-5: input_resolved payload has valueHash, never raw value ───────────────

// TestRespondToInputValueHashOnlyInEvent (AC-5)
// Captures bmad:node:input_resolved and verifies:
//   - Payload JSON does not contain the raw secret string.
//   - valueHash equals sha256(value)[:16].
func TestRespondToInputValueHashOnlyInEvent(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-s3-respond-hash-only"
	spec := InputSpec{ID: "secret", Source: InputFromUser, Shape: ShapeFree, Required: true}
	secret := "SECRET-TOKEN-ZZZ-987"

	h := setupSuspension(t, nodeID, processID, spec)
	h.waitForAwaiting(t)

	err := h.e.RespondToInput(h.execID, nodeID, "secret", secret)
	require.NoError(t, err)

	// Wait for goroutine to settle.
	select {
	case <-h.errCh:
	case <-time.After(2 * time.Second):
		t.Error("suspendForSpec goroutine did not exit within 2s")
	}

	snap := h.snap()
	resolvedEvts := eventsNamed(snap, EventInputResolved)
	require.Len(t, resolvedEvts, 1)

	payloadJSON, err2 := json.Marshal(resolvedEvts[0])
	require.NoError(t, err2)
	payloadStr := string(payloadJSON)

	assert.NotContains(t, payloadStr, secret,
		"bmad:node:input_resolved payload must NEVER contain the raw secret value")

	wantHash := sha256hex16(secret)
	assert.Contains(t, payloadStr, wantHash,
		"bmad:node:input_resolved payload must contain sha256[:16] as valueHash")
}

// sha256hex16 returns the first 16 hex chars of SHA-256(s).
func sha256hex16(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:16]
}

// ── No pending prompt ─────────────────────────────────────────────────────────

// TestRespondToInputNoPendingPrompt
// Node exists but no pending prompt → error wrapping ErrNoPendingPrompt.
func TestRespondToInputNoPendingPrompt(t *testing.T) {
	// NOT parallel: uses testRegistry global.
	const nodeID = "n1"
	const processID = "test-s3-respond-no-prompt"
	spec := InputSpec{ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true}
	registerTestProcess(t, ProcessDef{
		ID:         processID,
		Name:       "Test S3 No Prompt",
		Mode:       InteractGuided,
		SkillName:  "test-skill",
		InputSpecs: []InputSpec{spec},
		OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToMemory}},
	})
	hookEvents(t)

	state, _ := newSuspendState(nodeID, processID, spec)
	execID := state.exec.ID
	e := NewExecutor(nil, func(name string, data interface{}) {})
	e.mu.Lock()
	e.executions[execID] = state
	e.mu.Unlock()

	// No suspendForSpec called → PendingPrompts is empty.
	err := e.RespondToInput(execID, nodeID, "topic", "somevalue")
	require.Error(t, err)
	// ErrNoPendingPrompt does not exist yet — RED phase.
	assert.True(t, errors.Is(err, ErrNoPendingPrompt),
		"error must wrap ErrNoPendingPrompt, got: %v", err)
}

// ── Double-response: second fails ─────────────────────────────────────────────

// TestRespondToInputDoubleResponseSecondFails (§13.1 edge)
// Two sequential RespondToInput calls with valid values.
// First succeeds; second returns ErrNoPendingPrompt.
func TestRespondToInputDoubleResponseSecondFails(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-s3-respond-double"
	spec := InputSpec{ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true}

	h := setupSuspension(t, nodeID, processID, spec)
	h.waitForAwaiting(t)

	// First response — must succeed.
	err1 := h.e.RespondToInput(h.execID, nodeID, "topic", "first value")
	require.NoError(t, err1, "first RespondToInput must succeed")

	// Drain suspendForSpec goroutine.
	select {
	case <-h.errCh:
	case <-time.After(2 * time.Second):
		t.Error("suspendForSpec goroutine did not exit within 2s")
	}

	// Second response — must return ErrNoPendingPrompt.
	err2 := h.e.RespondToInput(h.execID, nodeID, "topic", "second value")
	require.Error(t, err2, "second RespondToInput must fail")
	assert.True(t, errors.Is(err2, ErrNoPendingPrompt),
		"second response must wrap ErrNoPendingPrompt, got: %v", err2)
}

// ── Unknown inputID ───────────────────────────────────────────────────────────

// TestRespondToInputUnknownInputID
// Suspend on "topic"; call with inputID "not-declared" → error wrapping ErrUnknownInput.
func TestRespondToInputUnknownInputID(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-s3-respond-unknown-input"
	spec := InputSpec{ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true}

	h := setupSuspension(t, nodeID, processID, spec)
	h.waitForAwaiting(t)

	// Inject a pending prompt with a fake inputID to bypass the NoPendingPrompt guard.
	// We need to test the ErrUnknownInput path specifically.
	// Artificially add a prompt for "not-declared" so the prompt-lookup passes
	// but findInputSpec fails.
	h.state.mu.Lock()
	h.state.exec.PendingPrompts = upsertPrompt(h.state.exec.PendingPrompts, PendingPrompt{
		NodeID:  nodeID,
		InputID: "not-declared",
		Round:   1,
		PromptID: hashPendingPrompt(nodeID, "not-declared", 1),
	})
	h.state.mu.Unlock()

	err := h.e.RespondToInput(h.execID, nodeID, "not-declared", "somevalue")
	require.Error(t, err)
	// ErrUnknownInput does not exist yet — RED phase.
	assert.True(t, errors.Is(err, ErrUnknownInput),
		"error must wrap ErrUnknownInput, got: %v", err)
}

// ── Bad execID ────────────────────────────────────────────────────────────────

// TestRespondToInputExecNotFound
// Bad execID → error wrapping ErrExecNotFound.
func TestRespondToInputExecNotFound(t *testing.T) {
	t.Parallel()
	hookEvents(t)
	e := NewExecutor(nil, func(string, interface{}) {})
	err := e.RespondToInput("exec-does-not-exist", "n1", "topic", "value")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrExecNotFound),
		"bad execID must return ErrExecNotFound, got: %v", err)
}

// ── AC-6: registryLookup rejects non-registry schemes (supplementary rows) ────

// TestRegistryLookupRejectsSchemes (AC-6)
// Marks t.Skip if TestRegistryLookupRejectsNonRegistryScheme already covers
// file:/http:/mcp:. Otherwise adds ftp: as an additional rejected scheme.
func TestRegistryLookupRejectsSchemes(t *testing.T) {
	t.Parallel()

	// Additional scheme not tested in executor_interactive_test.go.
	schemes := []struct {
		name string
		ref  string
	}{
		{"file: scheme rejected", "file:../secrets"},
		{"http: scheme rejected", "http://evil.example.com"},
		{"mcp: scheme rejected", "mcp:x#y"},
		{"ftp: scheme rejected", "ftp://example.com/path"},
		{"bare path rejected", "/etc/passwd"},
		{"relative path rejected", "../config.yaml"},
	}

	for _, tt := range schemes {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// registryLookup already tested in executor_interactive_test.go.
			// This file adds ftp: and bare-path rows for completeness.
			result, err := registryLookup(tt.ref)
			require.Error(t, err,
				"registryLookup must reject %q", tt.ref)
			assert.Empty(t, result)
			assert.True(t, errors.Is(err, ErrInvalidRegistryRef),
				"error must wrap ErrInvalidRegistryRef for %q, got: %v", tt.ref, err)
		})
	}
}

// ── Concurrency safety smoke ──────────────────────────────────────────────────

// TestRespondToInputConcurrencyNoPanic verifies that concurrent RespondToInput
// calls for different (execID, nodeID) pairs do not panic or deadlock.
// (run with -race to detect actual data races)
func TestRespondToInputConcurrencyNoPanic(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	// We set up multiple suspension states on the same executor and respond
	// concurrently; each responds to its own node.
	hookEvents(t)

	const processID = "test-s3-concurrency"
	spec := InputSpec{ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true}
	registerTestProcess(t, ProcessDef{
		ID:         processID,
		Name:       "Test S3 Concurrency",
		Mode:       InteractGuided,
		SkillName:  "test-skill",
		InputSpecs: []InputSpec{spec},
		OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToMemory}},
	})

	const N = 3
	type entry struct {
		execID string
		nodeID string
		cancel context.CancelFunc
		errCh  chan error
	}
	entries := make([]entry, N)
	e := NewExecutor(nil, func(name string, data interface{}) {})

	for i := 0; i < N; i++ {
		nodeID := fmt.Sprintf("n%d", i+1)
		state, nodeIndex := newSuspendState(nodeID, processID, spec)
		state.exec.ID = fmt.Sprintf("exec-concurrency-%d", i)
		state.exec.RepoPath = t.TempDir()
		execID := state.exec.ID

		e.mu.Lock()
		e.executions[execID] = state
		e.mu.Unlock()

		ctx, cancel := context.WithCancel(context.Background())
		errCh := make(chan error, 1)
		go func(s *execState, ni map[string]int, nid string) {
			errCh <- e.suspendForSpec(ctx, s, ni, nid, 1, spec, "")
		}(state, nodeIndex, nodeID)

		entries[i] = entry{execID: execID, nodeID: nodeID, cancel: cancel, errCh: errCh}

		// Wait for suspension.
		ok := pollForStatus(state, nodeIndex, nodeID, NodeAwaitingInput, 500*time.Millisecond)
		require.True(t, ok, "node %s must reach NodeAwaitingInput", nodeID)
	}

	// Respond concurrently.
	var wg sync.WaitGroup
	errs := make([]error, N)
	for i, en := range entries {
		i, en := i, en
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = e.RespondToInput(en.execID, en.nodeID, "topic", "value-"+en.nodeID)
		}()
	}
	wg.Wait()

	for i, err := range errs {
		assert.NoError(t, err, "concurrent RespondToInput for entry %d must not error", i)
	}

	// Drain goroutines.
	for _, en := range entries {
		select {
		case <-en.errCh:
		case <-time.After(2 * time.Second):
		}
		en.cancel()
	}
}
