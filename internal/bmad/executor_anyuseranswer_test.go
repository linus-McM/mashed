//go:build testing

// Tests for §5.3.2 composite-key accept/reject walk (AC-7, AC-8) and §5.3.4
// anyUserAnswerMatches last-round window (AC-12). Behind `-tags testing`
// because AC-7 + AC-8 wire uiadapter.NewMock, which lives under that tag.
package bmad

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/uiadapter"
)

// eventHasReason reports whether any event payload carries a `reason` field
// matching want. Handles both map[string]interface{} and map[string]string
// shapes since emit() serialises payloads through both paths.
func eventHasReason(events []interface{}, want string) bool {
	for _, p := range events {
		switch v := p.(type) {
		case map[string]interface{}:
			if r, _ := v["reason"].(string); r == want {
				return true
			}
		case map[string]string:
			if v["reason"] == want {
				return true
			}
		}
	}
	return false
}

// ── AC-12: anyUserAnswerMatches walks the last-round window only ─────────────

// TestGate_AnyUserAnswerMatches_LastRoundWindow verifies §5.3.4 AC-12.
// Legacy entries (Round==0) collapse to a full walk because all pre-U4
// entries share the same Round.
func TestGate_AnyUserAnswerMatches_LastRoundWindow(t *testing.T) {
	t.Parallel()

	const nodeID = "n1"

	tests := []struct {
		name    string
		history []NodeInputEntry
		tokens  []string
		want    bool
		why     string
	}{
		{
			name: "round-2 entry matches — walk scoped to last round returns true",
			history: []NodeInputEntry{
				{InputID: "iter", Round: 1, Value: "keep going"},
				{InputID: "iter", Round: 2, Value: "done", Key: "iter:confirm"},
			},
			tokens: []string{"done"},
			want:   true,
			why:    "last-round tail match is the GateUserConfirm happy path",
		},
		{
			name: "round-1 matches but tail is round-2 without match — ignored (stale)",
			history: []NodeInputEntry{
				{InputID: "iter", Round: 1, Value: "done", Key: "iter:confirm"},
				{InputID: "iter", Round: 2, Value: "keep going", Key: "iter:confirm"},
			},
			tokens: []string{"done"},
			want:   false,
			why:    "stale accept-token from a prior round must NOT satisfy the gate",
		},
		{
			name: "multi-sub-answer last round, matching NOT last in slice — still found",
			history: []NodeInputEntry{
				{InputID: "iter", Round: 1, Value: "seed"},
				{InputID: "iter", Round: 2, Value: "done", Key: "iter:confirm"},
				{InputID: "iter", Round: 2, Value: "noise", Key: "iter:stub"},
			},
			tokens: []string{"done"},
			want:   true,
			why:    "anyUserAnswerMatches walks every last-round entry, not just the tail",
		},
		{
			name: "multi-sub-answer last round, no matches — returns false",
			history: []NodeInputEntry{
				{InputID: "iter", Round: 2, Value: "pending", Key: "iter:confirm"},
				{InputID: "iter", Round: 2, Value: "idle", Key: "iter:stub"},
			},
			tokens: []string{"done"},
			want:   false,
			why:    "sanity: no entry in the last-round window matches the token set",
		},
		{
			name: "all legacy entries Round==0 with match — walk covers every entry",
			history: []NodeInputEntry{
				{InputID: "iter", Round: 0, Value: "first"},
				{InputID: "iter", Round: 0, Value: "done"},
				{InputID: "iter", Round: 0, Value: "last"},
			},
			tokens: []string{"done"},
			want:   true,
			why:    "pre-U4 snapshots share Round==0 — walk must reach every entry",
		},
		{
			name: "all legacy entries Round==0 without match — returns false",
			history: []NodeInputEntry{
				{InputID: "iter", Round: 0, Value: "first"},
				{InputID: "iter", Round: 0, Value: "middle"},
				{InputID: "iter", Round: 0, Value: "last"},
			},
			tokens: []string{"done"},
			want:   false,
			why:    "legacy walk covers everything but still respects containsToken semantics",
		},
		{
			name:    "empty history returns false",
			history: nil,
			tokens:  []string{"done"},
			want:    false,
			why:     "no history means no user answer — gate cannot satisfy",
		},
		{
			name: "empty tokens returns false even with matching-looking history",
			history: []NodeInputEntry{
				{InputID: "iter", Round: 2, Value: "done"},
			},
			tokens: nil,
			want:   false,
			why:    "containsToken contract: empty token set is never a match",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state, _ := newSessionState([]WorkflowNode{
				{ID: nodeID, ProcessID: "test-any-user-answer", Status: NodeRunning, NodeType: NodeTypeProcess},
			}, nil)
			state.exec.NodeInputHistory = map[string][]NodeInputEntry{nodeID: tt.history}

			got := anyUserAnswerMatches(state, nodeID, tt.tokens)
			assert.Equal(t, tt.want, got, "anyUserAnswerMatches mismatch: %s", tt.why)
		})
	}
}

// ── AC-7: reject-token matches a sub-answer and aborts ───────────────────────

// TestPartyMode_JSONSubmission_RejectTokenInFreeWidget verifies §5.3.2 AC-7.
// Pre-T5-GREEN, `containsToken(RejectTokens, rawJSONBlob)` is an exact match
// — the raw JSON `{"confirm":"done","stub":"cancel"}` never equals "cancel",
// abort never fires, test FAILS. T5-GREEN's sub-answer walk flips it to PASS.
func TestPartyMode_JSONSubmission_RejectTokenInFreeWidget(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-u4-reject-subanswer"
	const iterID = "round-response"
	const rawAnswer = `{"confirm":"done","stub":"cancel"}`

	iterSpec := InputSpec{
		ID:       iterID,
		Source:   InputFromUser,
		Shape:    ShapeJSON,
		Required: false,
		Prompt:   "Your response:",
	}
	topicSpec := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: false,
		Default:  "reject-topic",
		Prompt:   "Topic?",
	}
	proc := ProcessDef{
		ID:               processID,
		Name:             "U4 Reject Sub-answer",
		Mode:             InteractIterative,
		SkillName:        "test-skill",
		InputSpecs:       []InputSpec{topicSpec, iterSpec},
		OutputSpecs:      []OutputSpec{{ID: "out1", Target: OutputToMemory}},
		EnableAstAdapter: true,
		Gate: &IterationGate{
			Kind:         GateUserConfirm,
			AcceptTokens: []string{"accept-never-sent"},
			RejectTokens: []string{"cancel"},
			MaxRounds:    3, // safety ceiling — abort should land first
		},
	}
	registerTestProcess(t, proc)

	getSnap := hookEvents(t)

	h := newHarness(t)
	fixedAST := &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "mock:u4-reject",
		Nodes:       []uiadapter.UINode{{Type: "markdown", Content: "choose"}},
	}
	h.executor.adapter = uiadapter.NewMock(fixedAST, nil)

	var captured [][]string
	var mu sync.Mutex
	h.executor.SetCommandRunner(roundCaptureRunner(3, &captured, &mu))

	wf := saveInteractiveWorkflow(t, h.storage, processID)
	exec, err := h.executor.StartWorkflow(context.Background(), wf, t.TempDir(), "sonnet")
	require.NoError(t, err)
	execID := exec.ID
	t.Cleanup(func() { _ = h.executor.StopWorkflow(execID) })

	done := make(chan struct{})
	go feedAnswers(t, h.executor, execID, nodeID, iterID, []string{rawAnswer}, done)

	finalStatus := pollForTerminal(h.executor, execID, nodeID, 8*time.Second)
	assert.Equal(t, NodeFailed, finalStatus,
		"node must reach NodeFailed after reject-token sub-answer aborts the iteration")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}

	snap := getSnap()

	aborted := abortedEvents(snap)
	require.NotEmpty(t, aborted,
		"bmad:node:aborted must fire when a sub-answer matches RejectTokens")
	assert.True(t, eventHasReason(aborted, "rejected by user"),
		"aborted event must carry reason='rejected by user', got: %+v", aborted)

	gs := gateSatisfiedEvents(snap)
	assert.Empty(t, gs,
		"gate_satisfied must NOT fire on reject-token abort")
}

// ── AC-8: accept-token over any sub-answer ───────────────────────────────────

// TestPartyMode_JSONSubmission_AcceptTokenOnApprovalWidget verifies §5.3.2
// AC-8 as a behavioural smoke test for the T5-GREEN lastUserAnswer →
// anyUserAnswerMatches swap.
//
// RED caveat: with a single-key JSON submission, post-T4-GREEN flatten
// leaves exactly one composite entry whose Value=="done"; legacy
// lastUserAnswer also matches, so this test passes pre-T5-GREEN. AC-12's
// unit test is the strong RED signal.
func TestPartyMode_JSONSubmission_AcceptTokenOnApprovalWidget(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-u4-accept-approval"
	const iterID = "round-response"
	const rawAnswer = `{"confirm":"done"}`

	iterSpec := InputSpec{
		ID:       iterID,
		Source:   InputFromUser,
		Shape:    ShapeJSON,
		Required: false,
		Prompt:   "Approve?",
	}
	topicSpec := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: false,
		Default:  "approval-topic",
		Prompt:   "Topic?",
	}
	proc := ProcessDef{
		ID:               processID,
		Name:             "U4 Accept Approval Widget",
		Mode:             InteractIterative,
		SkillName:        "test-skill",
		InputSpecs:       []InputSpec{topicSpec, iterSpec},
		OutputSpecs:      []OutputSpec{{ID: "out1", Target: OutputToMemory}},
		EnableAstAdapter: true,
		Gate: &IterationGate{
			Kind:         GateUserConfirm,
			AcceptTokens: []string{"done"},
			MaxRounds:    3, // safety ceiling — accept-match should land first
		},
	}
	registerTestProcess(t, proc)

	getSnap := hookEvents(t)

	h := newHarness(t)
	// Decision-group + approval widget; response_key="confirm" binds the
	// widget output into the JSON submission's "confirm" key. The gate walks
	// Value, not Key, but the AST shape is what the story spec calls out.
	fixedAST := &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "mock:u4-approval",
		TurnSummary: "approve or revise",
		Nodes: []uiadapter.UINode{
			{
				Type:        "decision_group",
				ResponseKey: "confirm",
				Prompt:      "Approve?",
				Widget: &uiadapter.WidgetNode{
					Type:     "approval",
					YesLabel: "done",
					NoLabel:  "revise",
				},
			},
		},
	}
	h.executor.adapter = uiadapter.NewMock(fixedAST, nil)

	var captured [][]string
	var mu sync.Mutex
	h.executor.SetCommandRunner(roundCaptureRunner(3, &captured, &mu))

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
		"node must reach NodeComplete after GateUserConfirm accept-token match")

	snap := getSnap()

	gs := gateSatisfiedEvents(snap)
	require.Len(t, gs, 1,
		"gate_satisfied must fire exactly once on accept-token match")
	assert.True(t, eventHasReason(gs, "accept-token matched"),
		"gate_satisfied event must carry reason='accept-token matched', got: %+v", gs)

	aborted := abortedEvents(snap)
	assert.Empty(t, aborted,
		"aborted must NOT fire on accept-token match")
}
