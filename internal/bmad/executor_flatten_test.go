//go:build testing

// Tests for §5.3.1 flatten-on-receipt (AC-6) and §5.3.3 legacy ShapeFree
// regression (AC-9). Behind `-tags testing` because AC-6 wires
// uiadapter.NewMock, which lives under that tag.
package bmad

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/uiadapter"
)

// snapshotNodeState returns a copy of NodeInputs[nodeID] and
// NodeInputHistory[nodeID] taken under state.mu. Collapses the read-copy
// block shared by both flatten tests.
func snapshotNodeState(t *testing.T, e *Executor, execID, nodeID string) (map[string]string, []NodeInputEntry) {
	t.Helper()
	state, err := e.getState(execID)
	require.NoError(t, err)
	state.mu.Lock()
	defer state.mu.Unlock()
	inputs := make(map[string]string, len(state.exec.NodeInputs[nodeID]))
	for k, v := range state.exec.NodeInputs[nodeID] {
		inputs[k] = v
	}
	history := make([]NodeInputEntry, len(state.exec.NodeInputHistory[nodeID]))
	copy(history, state.exec.NodeInputHistory[nodeID])
	return inputs, history
}

// ── AC-6: §5.3.1 flatten-on-receipt populates composite + bare keys ──────────

// TestFlatten_CompositeAndBareKeys verifies AC-6. Gate is GateRoundLimit
// MaxRounds=2 so the loop exits deterministically after flatten runs on the
// round-2 resume, independent of T5-GREEN's accept-token walk. The adapter
// is wired so §5.3.1's `astStructuredInUse` predicate fires via the normal
// T3-GREEN path (EnableAstAdapter + non-empty PendingPrompt.Structured).
func TestFlatten_CompositeAndBareKeys(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-u4-flatten-composite"
	const iterID = "round-response"
	const rawAnswer = `{"confirm":"done","stub":"pytest"}`

	iterSpec := InputSpec{
		ID:       iterID,
		Source:   InputFromUser,
		Shape:    ShapeJSON,
		Required: false,
		Prompt:   "Your response:",
		HelpText: "Submit a JSON object of decisions.",
	}
	// Default + non-required so resolveInputs doesn't suspend on topic.
	topicSpec := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: false,
		Default:  "flatten-topic",
		Prompt:   "Topic?",
	}
	proc := ProcessDef{
		ID:               processID,
		Name:             "U4 Flatten Composite",
		Mode:             InteractIterative,
		SkillName:        "test-skill",
		InputSpecs:       []InputSpec{topicSpec, iterSpec},
		OutputSpecs:      []OutputSpec{{ID: "out1", Target: OutputToMemory}},
		EnableAstAdapter: true,
		Gate:             &IterationGate{Kind: GateRoundLimit, MaxRounds: 2},
	}
	registerTestProcess(t, proc)

	h := newHarness(t)
	// White-box adapter wiring: newHarness does not expose Option plumbing.
	fixedAST := &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "mock:u4-flatten",
		TurnSummary: "round 2 seed",
		Nodes:       []uiadapter.UINode{{Type: "markdown", Content: "choose"}},
	}
	h.executor.adapter = uiadapter.NewMock(fixedAST, nil)

	var captured [][]string
	var mu sync.Mutex
	h.executor.SetCommandRunner(roundCaptureRunner(2, &captured, &mu))

	wf := saveInteractiveWorkflow(t, h.storage, processID)
	exec, err := h.executor.StartWorkflow(context.Background(), wf, t.TempDir(), "sonnet")
	require.NoError(t, err)
	execID := exec.ID
	t.Cleanup(func() { _ = h.executor.StopWorkflow(execID) })

	done := make(chan struct{})
	go feedAnswers(t, h.executor, execID, nodeID, iterID, []string{rawAnswer}, done)

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("feedAnswers did not complete within 10s")
	}

	finalStatus := pollForTerminal(h.executor, execID, nodeID, 5*time.Second)
	require.Equal(t, NodeComplete, finalStatus,
		"node must reach NodeComplete after MaxRounds=2 ceiling")

	inputs, history := snapshotNodeState(t, h.executor, execID, nodeID)

	// Composite keys: the core §5.3.1 contract.
	assert.Equal(t, "done", inputs[iterID+":confirm"],
		"NodeInputs[%s][%s:confirm] must equal \"done\" (flatten composite)", nodeID, iterID)
	assert.Equal(t, "pytest", inputs[iterID+":stub"],
		"NodeInputs[%s][%s:stub] must equal \"pytest\" (flatten composite)", nodeID, iterID)

	// Bare specID preserves the raw JSON blob for downstream sendToSession
	// + upstream context readers (§5.3.1 last paragraph).
	assert.Equal(t, rawAnswer, inputs[iterID],
		"NodeInputs[%s][%s] must still carry the raw JSON blob under the bare spec ID", nodeID, iterID)

	// History: one entry per sub-answer, each tagged with composite Key and
	// Round=2 (the suspension that received the JSON was round+1 where round=1).
	findHistEntry := func(key string) (NodeInputEntry, bool) {
		for _, e := range history {
			if e.Key == key {
				return e, true
			}
		}
		return NodeInputEntry{}, false
	}

	confirmEntry, confirmOK := findHistEntry(iterID + ":confirm")
	require.True(t, confirmOK,
		"NodeInputHistory must contain an entry with Key=%q (one-per-sub-answer §5.3.1)", iterID+":confirm")
	assert.Equal(t, "done", confirmEntry.Value,
		"history entry for %s:confirm must carry Value=\"done\"", iterID)
	assert.Equal(t, 2, confirmEntry.Round,
		"history entry for %s:confirm must be tagged Round=2 (round+1 suspension)", iterID)

	stubEntry, stubOK := findHistEntry(iterID + ":stub")
	require.True(t, stubOK,
		"NodeInputHistory must contain an entry with Key=%q (one-per-sub-answer §5.3.1)", iterID+":stub")
	assert.Equal(t, "pytest", stubEntry.Value,
		"history entry for %s:stub must carry Value=\"pytest\"", iterID)
	assert.Equal(t, 2, stubEntry.Round,
		"history entry for %s:stub must be tagged Round=2 (round+1 suspension)", iterID)
}

// ── AC-9: legacy ShapeFree regression guard ──────────────────────────────────

// TestPartyMode_LegacyShapeFreeUnchanged is the regression guard against
// T4-GREEN (or T3-GREEN) leaking U4 behaviour into non-migrated processes.
// Passes in RED phase (flatten does not yet exist) and must keep passing
// after T4-GREEN lands.
func TestPartyMode_LegacyShapeFreeUnchanged(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-u4-legacy-shapefree"
	const iterID = "round-response"

	iterSpec := InputSpec{
		ID:       iterID,
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: false,
		Prompt:   "Your response:",
	}
	topicSpec := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: false,
		Default:  "legacy-topic",
		Prompt:   "Topic?",
	}
	proc := ProcessDef{
		ID:          processID,
		Name:        "U4 Legacy ShapeFree",
		Mode:        InteractIterative,
		SkillName:   "test-skill",
		InputSpecs:  []InputSpec{topicSpec, iterSpec},
		OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToMemory}},
		// EnableAstAdapter intentionally false — this is the regression surface.
		Gate: &IterationGate{Kind: GateUserConfirm, AcceptTokens: []string{"done"}},
	}
	registerTestProcess(t, proc)

	getSnap := hookEvents(t)

	h := newHarness(t)
	// Belt-and-braces: a bugged T3-GREEN that ignored EnableAstAdapter could
	// still populate Structured from a wired adapter, so assert none is wired.
	require.Nil(t, h.executor.adapter,
		"harness must start with nil adapter — test asserts the non-migrated path")

	var captured [][]string
	var mu sync.Mutex
	h.executor.SetCommandRunner(roundCaptureRunner(2, &captured, &mu))

	wf := saveInteractiveWorkflow(t, h.storage, processID)
	exec, err := h.executor.StartWorkflow(context.Background(), wf, t.TempDir(), "sonnet")
	require.NoError(t, err)
	execID := exec.ID
	t.Cleanup(func() { _ = h.executor.StopWorkflow(execID) })

	done := make(chan struct{})
	go feedAnswers(t, h.executor, execID, nodeID, iterID, []string{"more", "done"}, done)

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("feedAnswers did not complete within 10s")
	}

	finalStatus := pollForTerminal(h.executor, execID, nodeID, 5*time.Second)
	require.Equal(t, NodeComplete, finalStatus,
		"legacy node must reach NodeComplete on gate_satisfied")

	snap := getSnap()

	awaitingEvts := eventsNamed(snap, EventAwaitingInput)
	require.NotEmpty(t, awaitingEvts,
		"awaiting_input must fire at least once (gate_satisfied path requires 2 user turns)")
	for i, p := range awaitingEvts {
		pp, ok := p.(PendingPrompt)
		require.True(t, ok,
			"awaiting_input payload[%d] must be a PendingPrompt, got %T", i, p)
		assert.Empty(t, pp.Structured,
			"PendingPrompt.Structured must stay empty on legacy ShapeFree path (prompt %s/%s, round %d)",
			pp.NodeID, pp.InputID, pp.Round)
	}

	inputs, _ := snapshotNodeState(t, h.executor, execID, nodeID)
	for k := range inputs {
		assert.False(t, strings.Contains(k, ":"),
			"legacy NodeInputs must NOT contain composite keys (got %q)", k)
	}

	gs := gateSatisfiedEvents(snap)
	require.Len(t, gs, 1,
		"gate_satisfied must fire exactly once on accept-token match")
	rc := roundCompleteEvents(snap)
	assert.GreaterOrEqual(t, len(rc), 1,
		"at least one round_complete must fire before gate_satisfied")
	resolved := eventsNamed(snap, EventInputResolved)
	assert.GreaterOrEqual(t, len(resolved), 2,
		"input_resolved must fire at least twice (2 user turns: \"more\", \"done\")")
}
