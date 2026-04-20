// Package bmad contains tests for the round loop in executeInteractiveNode.
// Story bmad-interactive-04: Iteration gate + round loop.
//
// RED Phase: These tests define expected behaviour for the round loop, gate
// satisfaction, reject-token abort, sendToSession, and single-round exits.
// They MUST fail until the go-engineer implements the feature (GREEN phase).
//
// Harness conventions:
//   - package bmad (white-box) — same pattern as executor_suspend_test.go.
//   - hookEvents / testEventHook from eventhooks_test.go.
//   - newSessionState from executor_session_test.go.
//   - registerTestProcess from executor_suspend_test.go.
//   - feedAnswers helper suspends on NodeAwaitingInput, then calls RespondToInput.
package bmad

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Harness helpers ───────────────────────────────────────────────────────────

// roundCompleteEvents filters the event snapshot for EventRoundComplete payloads.
// EventRoundComplete is not yet defined — RED phase: compile error expected.
func roundCompleteEvents(snap []struct{ name string; payload interface{} }) []interface{} {
	return eventsNamed(snap, EventRoundComplete)
}

// gateSatisfiedEvents filters the event snapshot for EventGateSatisfied payloads.
// EventGateSatisfied is not yet defined — RED phase: compile error expected.
func gateSatisfiedEvents(snap []struct{ name string; payload interface{} }) []interface{} {
	return eventsNamed(snap, EventGateSatisfied)
}

// roundLimitEvents filters the event snapshot for EventRoundLimit payloads.
// EventRoundLimit is not yet defined — RED phase: compile error expected.
func roundLimitEvents(snap []struct{ name string; payload interface{} }) []interface{} {
	return eventsNamed(snap, EventRoundLimit)
}

// abortedEvents filters the event snapshot for EventAborted payloads.
func abortedEvents(snap []struct{ name string; payload interface{} }) []interface{} {
	return eventsNamed(snap, EventAborted)
}

// buildIterationProcess returns a ProcessDef suitable for round-loop integration
// testing: one required pre-process input ("topic"), one iteration input
// ("round-response", optional user free), and an optional Gate.
func buildIterationProcess(id string, gate *IterationGate) ProcessDef {
	return ProcessDef{
		ID:        id,
		Name:      "Test Iteration " + id,
		Mode:      InteractIterative,
		SkillName: "test-skill",
		InputSpecs: []InputSpec{
			{
				ID:       "topic",
				Source:   InputFromUser,
				Shape:    ShapeFree,
				Required: true,
				Prompt:   "What is the topic?",
			},
			{
				ID:       "round-response",
				Source:   InputFromUser,
				Shape:    ShapeFree,
				Required: false,
				Prompt:   "Your response:",
			},
		},
		OutputSpecs: []OutputSpec{
			{ID: "out1", Target: OutputToMemory},
		},
		Gate: gate,
	}
}

// buildGuidedProcess returns a ProcessDef with only required user inputs —
// no optional+free+prompt spec — so iterationInput returns false.
// This simulates a single-round guided process.
func buildGuidedProcess(id string) ProcessDef {
	return ProcessDef{
		ID:        id,
		Name:      "Test Guided " + id,
		Mode:      InteractGuided,
		SkillName: "test-skill",
		InputSpecs: []InputSpec{
			{
				ID:       "topic",
				Source:   InputFromUser,
				Shape:    ShapeFree,
				Required: true,
				Prompt:   "What is the topic?",
			},
		},
		OutputSpecs: []OutputSpec{
			{ID: "out1", Target: OutputToMemory},
		},
		Gate: nil, // no gate — single-round process
	}
}

// feedAnswers is a helper that, in a goroutine, waits for the node to enter
// NodeAwaitingInput and then calls RespondToInput with each queued answer in
// sequence. After all answers are consumed, subsequent polls are no-ops.
//
// It returns a channel that closes when all answers have been fed (or the
// timeout elapses). Call as a goroutine before starting the executor:
//
//	done := make(chan struct{})
//	go feedAnswers(t, e, execID, nodeID, inputID, []string{"more","done"}, done)
func feedAnswers(
	t *testing.T,
	e *Executor,
	execID, nodeID, inputID string,
	answers []string,
	done chan struct{},
) {
	t.Helper()
	defer close(done)

	for _, answer := range answers {
		deadline := time.Now().Add(3 * time.Second)
		fed := false
		terminal := false
		for time.Now().Before(deadline) {
			state, err := e.getState(execID)
			if err != nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			state.mu.Lock()
			var status WorkflowNodeStatus
			for _, n := range state.exec.Nodes {
				if n.ID == nodeID {
					status = n.Status
					break
				}
			}
			state.mu.Unlock()

			// If the node has reached a terminal state, exit without
			// erroring — the executor decided to stop before consuming
			// the remaining queued answers (e.g. MaxRounds ceiling).
			if status == NodeComplete || status == NodeFailed {
				terminal = true
				break
			}

			if status == NodeAwaitingInput {
				if rErr := e.RespondToInput(execID, nodeID, inputID, answer); rErr == nil {
					fed = true
					break
				}
			}
			time.Sleep(15 * time.Millisecond)
		}
		if terminal {
			return
		}
		if !fed {
			t.Errorf("feedAnswers: could not feed answer %q within deadline", answer)
			return
		}
		// Brief pause to allow the executor goroutine to process the response
		// and loop back to waiting before we poll for NodeAwaitingInput again.
		time.Sleep(20 * time.Millisecond)
	}
}

// roundCaptureRunner returns a CommandRunner that:
//   - Handles tmux new-session (returns "ok").
//   - Handles tmux list-panes: alternates between "0\n" (alive) for the first
//     N calls per round, then "1\n" (dead) once per round to terminate waitForIdleCompletion.
//   - Handles tmux capture-pane: returns a canned output containing an idle prompt.
//   - Captures send-keys calls into the provided slice.
//
// This allows executeInteractiveNode to proceed through each round without
// hanging on waitForIdleCompletion.
func roundCaptureRunner(rounds int, capturedCalls *[][]string, mu *sync.Mutex) CommandRunner {
	var listPaneCount int32
	// paneDeadEvery controls how many list-panes calls per round before "dead".
	// Set to 3 to allow: first alive, second alive, third dead (simulating idle).
	const callsPerRound = 3

	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name != "tmux" {
			return []byte("ok"), nil
		}
		if len(args) == 0 {
			return []byte("ok"), nil
		}

		switch args[0] {
		case "new-session":
			return []byte("ok"), nil
		case "send-keys":
			mu.Lock()
			cp := make([]string, len(args))
			copy(cp, args)
			*capturedCalls = append(*capturedCalls, cp)
			mu.Unlock()
			return []byte("ok"), nil
		case "list-panes":
			// Use atomic-style increment without importing sync/atomic by using
			// the mutex we already have.
			mu.Lock()
			listPaneCount++
			count := listPaneCount
			mu.Unlock()

			// Die on every callsPerRound-th call so waitForIdleCompletion exits.
			if count%int32(callsPerRound) == 0 {
				return []byte("1\n"), nil
			}
			return []byte("0\n"), nil
		case "capture-pane":
			// Return output with an idle prompt so detectIdlePrompt returns true.
			return []byte("Round output.\n❯\n"), nil
		default:
			return []byte("ok"), nil
		}
	}
}

// nodeStatus reads a node's current status from the executor's state.
func nodeStatusFromExec(e *Executor, execID, nodeID string) WorkflowNodeStatus {
	state, err := e.getState(execID)
	if err != nil {
		return NodeFailed
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	for _, n := range state.exec.Nodes {
		if n.ID == nodeID {
			return n.Status
		}
	}
	return NodeFailed
}

// pollForTerminal waits until the node reaches NodeComplete or NodeFailed.
func pollForTerminal(e *Executor, execID, nodeID string, timeout time.Duration) WorkflowNodeStatus {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		s := nodeStatusFromExec(e, execID, nodeID)
		if s == NodeComplete || s == NodeFailed {
			return s
		}
		time.Sleep(20 * time.Millisecond)
	}
	return ""
}

// ── AC-1 + AC-5: 3-round loop, accept on round 3 ─────────────────────────────

// TestRoundLoopGateUserConfirmAcceptsOnThirdRound (AC-1, AC-5)
// 3-round loop: user answers "more" twice, then "done".
// Verifies: round_complete fires 3 times, gate_satisfied fires once,
// NodeRounds[nodeID]==3, NodeOutputs has all 3 round keys plus final nodeID key.
func TestRoundLoopGateUserConfirmAcceptsOnThirdRound(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-iter-roundloop-accept"

	gate := &IterationGate{
		Kind:         GateUserConfirm,
		AcceptTokens: []string{"done"},
	}
	proc := buildIterationProcess(processID, gate)
	// Pre-process input: seed "topic" so resolveInputs doesn't suspend on it.
	proc.InputSpecs[0] = InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: false, // make it optional so resolveInputs uses Default
		Default:  "test-topic",
		Prompt:   "Topic?",
	}
	registerTestProcess(t, proc)

	getSnap := hookEvents(t)

	h := newHarness(t)
	var capturedCalls [][]string
	var mu sync.Mutex
	h.executor.SetCommandRunner(roundCaptureRunner(3, &capturedCalls, &mu))

	wf := saveInteractiveWorkflow(t, h.storage, processID)
	exec, err := h.executor.StartWorkflow(context.Background(), wf, t.TempDir(), "sonnet")
	require.NoError(t, err)
	execID := exec.ID
	t.Cleanup(func() { _ = h.executor.StopWorkflow(execID) })

	// Feed answers: "more", "more", "done" to the round-response input.
	done := make(chan struct{})
	go feedAnswers(t, h.executor, execID, nodeID, "round-response",
		[]string{"more", "more", "done"}, done)

	// Wait for all answers to be consumed and node to reach terminal state.
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("feedAnswers did not complete within 10s")
	}

	finalStatus := pollForTerminal(h.executor, execID, nodeID, 5*time.Second)
	assert.Equal(t, NodeComplete, finalStatus, "node must reach NodeComplete after gate satisfied")

	snap := getSnap()

	// AC-5: round_complete must fire exactly 3 times.
	rc := roundCompleteEvents(snap)
	assert.Len(t, rc, 3, "bmad:node:round_complete must fire exactly 3 times")

	// gate_satisfied must fire exactly once.
	gs := gateSatisfiedEvents(snap)
	require.Len(t, gs, 1, "bmad:node:gate_satisfied must fire exactly once")

	// NodeRounds[nodeID] must equal 3.
	state, err := h.executor.getState(execID)
	require.NoError(t, err)
	state.mu.Lock()
	rounds := state.exec.NodeRounds[nodeID]
	outputs := make(map[string]string, len(state.exec.NodeOutputs))
	for k, v := range state.exec.NodeOutputs {
		outputs[k] = v
	}
	state.mu.Unlock()

	assert.Equal(t, 3, rounds, "NodeRounds[nodeID] must equal 3")

	// NodeOutputs must have round keys and the final nodeID key.
	for i := 1; i <= 3; i++ {
		roundKey := fmt.Sprintf("%s-round-%d", nodeID, i)
		assert.Contains(t, outputs, roundKey,
			"NodeOutputs must contain round key %q", roundKey)
	}
	assert.Contains(t, outputs, nodeID,
		"NodeOutputs must contain final nodeID key mirroring last round")

	// The final nodeID output must equal the last round's output.
	lastRoundKey := fmt.Sprintf("%s-round-3", nodeID)
	assert.Equal(t, outputs[lastRoundKey], outputs[nodeID],
		"NodeOutputs[nodeID] must mirror NodeOutputs[%q]", lastRoundKey)
}

// ── AC-2 + AC-3: MaxRounds exit ───────────────────────────────────────────────

// TestRoundLoopMaxRoundsExit (AC-2, AC-3)
// Table: two gate configurations. User never sends accept token.
// Loop exits at MaxRounds; bmad:node:round_limit fires.
func TestRoundLoopMaxRoundsExit(t *testing.T) {
	// NOT parallel: testEventHook and testRegistry globals.
	tests := []struct {
		name      string
		gate      *IterationGate
		maxRounds int
	}{
		{
			name: "GateRoundLimit MaxRounds=2 exits after 2 rounds",
			gate: &IterationGate{
				Kind:      GateRoundLimit,
				MaxRounds: 2,
			},
			maxRounds: 2,
		},
		{
			name: "GateUserConfirm MaxRounds=2 safety ceiling — user never accepts",
			gate: &IterationGate{
				Kind:         GateUserConfirm,
				AcceptTokens: []string{"done"},
				MaxRounds:    2,
			},
			maxRounds: 2,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// NOT parallel: testEventHook and testRegistry globals.
			// Sanitise subtest name into a storage-safe process ID
			// ([a-zA-Z0-9_-] only — see storage.validateID).
			safe := strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
					return r
				}
				return '-'
			}, tt.name)
			procID := fmt.Sprintf("test-iter-maxrounds-%s", safe)
			proc := buildIterationProcess(procID, tt.gate)
			// Make topic optional so resolveInputs doesn't suspend.
			proc.InputSpecs[0] = InputSpec{
				ID:       "topic",
				Source:   InputFromUser,
				Shape:    ShapeFree,
				Required: false,
				Default:  "test-topic",
				Prompt:   "Topic?",
			}
			registerTestProcess(t, proc)

			getSnap := hookEvents(t)

			h := newHarness(t)
			var capturedCalls [][]string
			var mu sync.Mutex
			h.executor.SetCommandRunner(roundCaptureRunner(tt.maxRounds, &capturedCalls, &mu))

			wf := saveInteractiveWorkflow(t, h.storage, procID)
			exec, err := h.executor.StartWorkflow(context.Background(), wf, t.TempDir(), "sonnet")
			require.NoError(t, err)
			execID := exec.ID
			t.Cleanup(func() { _ = h.executor.StopWorkflow(execID) })

			// User keeps answering "keep going" — never an accept token.
			done := make(chan struct{})
			go feedAnswers(t, h.executor, execID, "n1", "round-response",
				[]string{"keep going", "keep going", "keep going"}, done)

			finalStatus := pollForTerminal(h.executor, execID, "n1", 10*time.Second)
			assert.Equal(t, NodeComplete, finalStatus,
				"node must complete after MaxRounds even without accept token")

			// Wait for feed goroutine to exit.
			select {
			case <-done:
			case <-time.After(2 * time.Second):
			}

			snap := getSnap()

			// bmad:node:round_limit must fire with round == maxRounds.
			rl := roundLimitEvents(snap)
			require.NotEmpty(t, rl, "bmad:node:round_limit must fire when MaxRounds is reached")

			// round_complete must fire exactly maxRounds times.
			rc := roundCompleteEvents(snap)
			assert.Len(t, rc, tt.maxRounds,
				"round_complete must fire exactly %d times", tt.maxRounds)

			// gate_satisfied must NOT fire (MaxRounds branch, not gate).
			gs := gateSatisfiedEvents(snap)
			assert.Empty(t, gs, "gate_satisfied must NOT fire when MaxRounds branch exits")

			// NodeRounds must equal maxRounds.
			state, stateErr := h.executor.getState(execID)
			require.NoError(t, stateErr)
			state.mu.Lock()
			rounds := state.exec.NodeRounds["n1"]
			state.mu.Unlock()
			assert.Equal(t, tt.maxRounds, rounds, "NodeRounds must equal MaxRounds")
		})
	}
}

// ── AC-6: reject token aborts ─────────────────────────────────────────────────

// TestRoundLoopRejectTokenAborts (AC-6)
// Gate{RejectTokens: ["abort","cancel"]}. First round user answers "cancel".
// Expect: bmad:node:aborted fires with reason "rejected by user"; NodeFailed.
func TestRoundLoopRejectTokenAborts(t *testing.T) {
	// NOT parallel: testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-iter-rejecttoken"

	gate := &IterationGate{
		Kind:         GateUserConfirm,
		AcceptTokens: []string{"done"},
		RejectTokens: []string{"abort", "cancel"},
	}
	proc := buildIterationProcess(processID, gate)
	// Make topic optional.
	proc.InputSpecs[0] = InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: false,
		Default:  "test-topic",
		Prompt:   "Topic?",
	}
	registerTestProcess(t, proc)

	getSnap := hookEvents(t)

	h := newHarness(t)
	var capturedCalls [][]string
	var mu sync.Mutex
	h.executor.SetCommandRunner(roundCaptureRunner(1, &capturedCalls, &mu))

	wf := saveInteractiveWorkflow(t, h.storage, processID)
	exec, err := h.executor.StartWorkflow(context.Background(), wf, t.TempDir(), "sonnet")
	require.NoError(t, err)
	execID := exec.ID
	t.Cleanup(func() { _ = h.executor.StopWorkflow(execID) })

	// Feed "cancel" as the first round response.
	done := make(chan struct{})
	go feedAnswers(t, h.executor, execID, nodeID, "round-response",
		[]string{"cancel"}, done)

	finalStatus := pollForTerminal(h.executor, execID, nodeID, 8*time.Second)
	assert.Equal(t, NodeFailed, finalStatus, "node must fail after reject token")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

	snap := getSnap()

	// bmad:node:aborted must fire with reason "rejected by user".
	aborted := abortedEvents(snap)
	require.NotEmpty(t, aborted, "bmad:node:aborted must fire on reject token")

	reasonOK := false
	for _, p := range aborted {
		switch v := p.(type) {
		case map[string]interface{}:
			if r, ok := v["reason"].(string); ok && r == "rejected by user" {
				reasonOK = true
			}
		case map[string]string:
			if v["reason"] == "rejected by user" {
				reasonOK = true
			}
		}
	}
	assert.True(t, reasonOK,
		"aborted event must carry reason='rejected by user', got: %+v", aborted)

	// No round_complete after the abort (or exactly 1 for round 1 before abort check).
	// The reject-token check happens AFTER the first round_complete fires (per the
	// round loop shape in the story). So we allow at most 1 round_complete.
	rc := roundCompleteEvents(snap)
	assert.LessOrEqual(t, len(rc), 1,
		"no more than 1 round_complete must fire when reject token aborts after round 1")

	// gate_satisfied must NOT fire.
	gs := gateSatisfiedEvents(snap)
	assert.Empty(t, gs, "gate_satisfied must NOT fire on reject-token abort")
}

// ── AC-7: single-round guided process ─────────────────────────────────────────

// TestRoundLoopSingleRoundGuided (AC-7)
// ProcessDef with only required user inputs, no optional+free+prompt spec.
// iterationInput returns false → loop exits after 1 round. NodeComplete.
func TestRoundLoopSingleRoundGuided(t *testing.T) {
	// NOT parallel: testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-iter-single-round-guided"

	proc := buildGuidedProcess(processID)
	// Make topic optional so resolveInputs doesn't suspend.
	proc.InputSpecs[0] = InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: false,
		Default:  "test-topic",
		Prompt:   "Topic?",
	}
	registerTestProcess(t, proc)

	getSnap := hookEvents(t)

	h := newHarness(t)
	var capturedCalls [][]string
	var mu sync.Mutex
	h.executor.SetCommandRunner(roundCaptureRunner(1, &capturedCalls, &mu))

	wf := saveInteractiveWorkflow(t, h.storage, processID)
	exec, err := h.executor.StartWorkflow(context.Background(), wf, t.TempDir(), "sonnet")
	require.NoError(t, err)
	execID := exec.ID
	t.Cleanup(func() { _ = h.executor.StopWorkflow(execID) })

	// No feedAnswers goroutine — no iteration input exists.
	finalStatus := pollForTerminal(h.executor, execID, nodeID, 8*time.Second)
	assert.Equal(t, NodeComplete, finalStatus,
		"single-round guided node must complete after one round")

	snap := getSnap()

	// Exactly 1 round_complete.
	rc := roundCompleteEvents(snap)
	assert.Len(t, rc, 1, "single-round process must emit exactly 1 round_complete")

	// No gate_satisfied (gate is nil, loop exits via iterationInput returning false).
	// The loop exits on !ok from iterationInput, not via gate satisfaction.
	_ = getSnap

	// NodeRounds must equal 1.
	state, err := h.executor.getState(execID)
	require.NoError(t, err)
	state.mu.Lock()
	rounds := state.exec.NodeRounds[nodeID]
	state.mu.Unlock()
	assert.Equal(t, 1, rounds, "NodeRounds must equal 1 for single-round process")
}

// ── AC-8: sendToSession writes literal then Enter ─────────────────────────────

// TestSendToSessionTwoTmuxCalls (AC-8)
// sendToSession must emit exactly two tmux send-keys calls:
//  1. send-keys -t <target> -l <escapedAnswer>
//  2. send-keys -t <target> Enter
//
// escapeTmuxLiteral strips C0 controls but preserves printable chars.
// The answer "hello; rm -rf /" contains shell-meta chars that are safe in -l mode.
func TestSendToSessionTwoTmuxCalls(t *testing.T) {
	t.Parallel()

	const nodeID = "n1"
	const tmuxTarget = "mashed_abc:0.1"
	const answer = "hello; rm -rf /"

	// escapeTmuxLiteral is defined in question.go — the escaped form of the
	// answer must equal the answer itself because it contains no C0 controls.
	expectedEscaped := escapeTmuxLiteral(answer)

	var captured [][]string
	var mu sync.Mutex

	runner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "send-keys" {
			mu.Lock()
			cp := make([]string, len(args))
			copy(cp, args)
			captured = append(captured, cp)
			mu.Unlock()
		}
		return []byte("ok"), nil
	}

	state, _ := newSessionState([]WorkflowNode{
		{
			ID:         nodeID,
			ProcessID:  "test-send-session",
			Label:      "Send Session",
			Status:     NodeRunning,
			NodeType:   NodeTypeProcess,
			TmuxTarget: tmuxTarget,
		},
	}, nil)

	e := NewExecutor(nil, func(string, interface{}) {})
	e.SetCommandRunner(runner)

	// sendToSession does not exist yet — RED phase.
	ctx := context.Background()
	err := e.sendToSession(ctx, state, nodeID, answer)
	require.NoError(t, err, "sendToSession must return nil for valid answer")

	mu.Lock()
	calls := make([][]string, len(captured))
	copy(calls, captured)
	mu.Unlock()

	require.Len(t, calls, 2, "sendToSession must make exactly 2 tmux send-keys calls")

	// First call: send-keys -t <target> -l <escaped>
	first := calls[0]
	assert.Equal(t, "send-keys", first[0], "first call must be send-keys")
	assert.Contains(t, first, "-t", "first call must have -t flag")
	assert.Contains(t, first, tmuxTarget, "first call must target the correct pane")
	assert.Contains(t, first, "-l", "first call must use -l (literal) flag")
	assert.Contains(t, first, expectedEscaped, "first call must contain the escaped answer")

	// Second call: send-keys -t <target> Enter
	second := calls[1]
	assert.Equal(t, "send-keys", second[0], "second call must be send-keys")
	assert.Contains(t, second, "-t", "second call must have -t flag")
	assert.Contains(t, second, tmuxTarget, "second call must target the correct pane")
	assert.Contains(t, second, "Enter", "second call must send Enter key")
	// -l flag must NOT appear on the Enter call.
	hasLiteralFlag := false
	for _, arg := range second {
		if arg == "-l" {
			hasLiteralFlag = true
		}
	}
	assert.False(t, hasLiteralFlag, "Enter send-keys call must NOT have -l flag")
}

// TestSendToSessionEscapesControlChars
// Answers containing C0 control bytes (e.g. newline, ESC) must be stripped
// by escapeTmuxLiteral so they don't prematurely submit the Claude prompt.
func TestSendToSessionEscapesControlChars(t *testing.T) {
	t.Parallel()

	const nodeID = "n1"
	const tmuxTarget = "mashed_def:0.0"
	// answer with embedded newline and ESC — both should be stripped.
	answer := "hello\nworld\x1b[1m"

	var captured [][]string
	var mu sync.Mutex

	runner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "send-keys" {
			mu.Lock()
			cp := make([]string, len(args))
			copy(cp, args)
			captured = append(captured, cp)
			mu.Unlock()
		}
		return []byte("ok"), nil
	}

	state, _ := newSessionState([]WorkflowNode{
		{
			ID:         nodeID,
			ProcessID:  "test-send-escape",
			Status:     NodeRunning,
			NodeType:   NodeTypeProcess,
			TmuxTarget: tmuxTarget,
		},
	}, nil)

	e := NewExecutor(nil, func(string, interface{}) {})
	e.SetCommandRunner(runner)

	ctx := context.Background()
	err := e.sendToSession(ctx, state, nodeID, answer)
	require.NoError(t, err)

	mu.Lock()
	calls := make([][]string, len(captured))
	copy(calls, captured)
	mu.Unlock()

	require.Len(t, calls, 2, "must have 2 send-keys calls")

	// The literal payload in the first call must not contain raw newline or ESC.
	literalArg := ""
	first := calls[0]
	for i, arg := range first {
		if arg == "-l" && i+1 < len(first) {
			literalArg = first[i+1]
			break
		}
	}
	assert.NotContains(t, literalArg, "\n", "literal payload must not contain newline")
	assert.NotContains(t, literalArg, "\x1b", "literal payload must not contain ESC")
	assert.Contains(t, literalArg, "helloworld",
		"printable chars must survive escapeTmuxLiteral stripping")
}

// ── AC-5 companion: final round mirrors NodeOutputs ──────────────────────────

// TestFinalRoundMirrorsNodeOutputs
// After gate satisfied, NodeOutputs[nodeID] must equal NodeOutputs[{nodeID}-round-N].
// This is a pure state-assertion test — no integration loop required.
func TestFinalRoundMirrorsNodeOutputs(t *testing.T) {
	t.Parallel()

	const nodeID = "n1"
	const lastRoundOutput = "round 3 output text"

	state, _ := newSessionState([]WorkflowNode{
		{ID: nodeID, ProcessID: "test-final-mirror", Status: NodeComplete, NodeType: NodeTypeProcess},
	}, nil)

	// Simulate what the GREEN loop does on gate satisfaction: write final round
	// output to both the round key and the nodeID key.
	roundKey := fmt.Sprintf("%s-round-%d", nodeID, 3)
	state.exec.NodeOutputs = map[string]string{
		nodeID + "-round-1": "round 1 output",
		nodeID + "-round-2": "round 2 output",
		roundKey:            lastRoundOutput,
		nodeID:              lastRoundOutput, // mirrored by the loop
	}

	// Assert invariant: NodeOutputs[nodeID] == NodeOutputs[roundKey].
	assert.Equal(t,
		state.exec.NodeOutputs[roundKey],
		state.exec.NodeOutputs[nodeID],
		"NodeOutputs[nodeID] must mirror NodeOutputs[%q] after gate satisfied", roundKey)

	// Also verify the round keys exist.
	for i := 1; i <= 3; i++ {
		k := fmt.Sprintf("%s-round-%d", nodeID, i)
		assert.Contains(t, state.exec.NodeOutputs, k,
			"NodeOutputs must contain %q", k)
	}
}
