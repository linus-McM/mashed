package bmad

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Story skills-cmd-01 — RED phase ──
//
// Until the GREEN phase declares NodeTypeCommand on internal/bmad/types.go,
// this file fails to compile. That compile failure IS the RED-state signal
// for the TDD lifecycle.

// countingNewSessionRunner wraps successRunner so the canonical pane-dead
// behaviour is preserved while a side-counter records every `tmux new-session`
// invocation. Used as the executor command runner in AC-2 / AC-3.
func countingNewSessionRunner() (CommandRunner, *int32) {
	var newSessionCalls int32
	base := successRunner()
	runner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "new-session" {
			atomic.AddInt32(&newSessionCalls, 1)
		}
		return base(ctx, name, args...)
	}
	return runner, &newSessionCalls
}

// saveSingleNodeWorkflow persists a one-node WorkflowDef of the given NodeType
// and returns its ID. ProcessID points at a real registry entry so the existing
// process-node spawn path can resolve it; command nodes ignore the ProcessID.
func saveSingleNodeWorkflow(t *testing.T, s *Storage, nodeType NodeType, config map[string]string) string {
	t.Helper()
	now := time.Now().Format(time.RFC3339)
	wf := WorkflowDef{
		ID:          "wf-skills-cmd-01",
		Name:        "skills-cmd-01-single",
		Description: "single-node fixture for skills-cmd-01 tests",
		Nodes: []WorkflowNode{
			{
				ID:        "n1",
				ProcessID: "bmad-brainstorming",
				Label:     "Only",
				Status:    NodePending,
				Config:    config,
				NodeType:  nodeType,
			},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, s.SaveWorkflow(wf))
	return wf.ID
}

// waitForNodeStatus polls GetExecution until the named node reaches the
// expected status or the deadline elapses. Tick matches the harness's
// 50ms pollInterval — sub-50ms ticks would just spin without observing change.
func waitForNodeStatus(t *testing.T, e *Executor, execID, nodeID string, want WorkflowNodeStatus) {
	t.Helper()
	require.Eventuallyf(t, func() bool {
		got, err := e.GetExecution(execID)
		if err != nil {
			return false
		}
		for _, n := range got.Nodes {
			if n.ID == nodeID {
				return n.Status == want
			}
		}
		return false
	}, 3*time.Second, 50*time.Millisecond,
		"node %q never reached status %q", nodeID, want)
}

// startSingleNodeAndWait saves a single-node workflow, registers a
// counting runner, starts the workflow, and blocks until the node reaches
// `want`. Returns the new-session counter for downstream assertions and
// schedules StopWorkflow as cleanup insurance against goroutine leaks if
// the test fails before the node settles.
func startSingleNodeAndWait(t *testing.T, h *testHarness, nt NodeType, cfg map[string]string, want WorkflowNodeStatus) *int32 {
	t.Helper()
	runner, calls := countingNewSessionRunner()
	h.executor.SetCommandRunner(runner)

	wfID := saveSingleNodeWorkflow(t, h.storage, nt, cfg)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

	waitForNodeStatus(t, h.executor, exec.ID, "n1", want)
	return calls
}

// TestStory1_AC1_NodeTypeCommandRoundTrip — AC-1: NodeTypeCommand round-trips
// through WorkflowDef JSON and EffectiveType() reports it after decode.
func TestStory1_AC1_NodeTypeCommandRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		node   WorkflowNode
		expect map[string]string
	}{
		{
			name: "simplify command with all three config keys",
			node: WorkflowNode{
				ID:       "cmd-1",
				Label:    "Simplify",
				Position: Position{X: 10, Y: 20},
				Status:   NodePending,
				NodeType: NodeTypeCommand,
				Config: map[string]string{
					"commandName":        "simplify",
					"commandPath":        "/tmp/simplify.md",
					"commandDescription": "Review recent changes",
				},
			},
			expect: map[string]string{
				"commandName":        "simplify",
				"commandPath":        "/tmp/simplify.md",
				"commandDescription": "Review recent changes",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wf := WorkflowDef{
				ID:    "wf-cmd-rt",
				Name:  "command round-trip",
				Nodes: []WorkflowNode{tc.node},
				Edges: []WorkflowEdge{},
			}

			blob, err := json.Marshal(wf)
			require.NoError(t, err)

			var decoded WorkflowDef
			require.NoError(t, json.Unmarshal(blob, &decoded))
			require.Len(t, decoded.Nodes, 1)

			got := decoded.Nodes[0]
			assert.Equal(t, NodeTypeCommand, got.EffectiveType())
			for key, want := range tc.expect {
				assert.Equal(t, want, got.Config[key], "config[%q]", key)
			}
		})
	}
}

// TestStory1_AC2_CommandNodeFailFast — AC-2: a command node transitions to
// NodeFailed with zero `tmux new-session` invocations. The status transition
// is the externally observable signal that failNode was invoked from the
// command-node fail-fast branch.
func TestStory1_AC2_CommandNodeFailFast(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config map[string]string
	}{
		{
			name: "simplify command fails fast with no tmux spawn",
			config: map[string]string{
				"commandName":        "simplify",
				"commandPath":        "/tmp/simplify.md",
				"commandDescription": "Review recent changes",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)
			calls := startSingleNodeAndWait(t, h, NodeTypeCommand, tc.config, NodeFailed)

			assert.Equal(t, int32(0), atomic.LoadInt32(calls),
				"command fail-fast must NOT invoke `tmux new-session`")
		})
	}
}

// TestStory1_AC3_ProcessNodeUnchanged — AC-3 regression guard: a legacy
// process node still spawns tmux and reaches NodeComplete. If a future
// change to the executeNode switch breaks the default branch, this fires.
func TestStory1_AC3_ProcessNodeUnchanged(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	calls := startSingleNodeAndWait(t, h, NodeTypeProcess, map[string]string{}, NodeComplete)

	assert.GreaterOrEqual(t, atomic.LoadInt32(calls), int32(1),
		"process-node dispatch must still invoke `tmux new-session`")
}
