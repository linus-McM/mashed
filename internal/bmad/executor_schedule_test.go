package bmad

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopoSort_Sequential(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "A"}, {ID: "B"}, {ID: "C"},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
		{ID: "e2", Source: "B", Target: "C"},
	}
	tiers, err := topoSort(nodes, edges)
	require.NoError(t, err)
	require.Len(t, tiers, 3)
	assert.Equal(t, []string{"A"}, tiers[0])
	assert.Equal(t, []string{"B"}, tiers[1])
	assert.Equal(t, []string{"C"}, tiers[2])
}

func TestTopoSort_Parallel(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "C"},
		{ID: "e2", Source: "B", Target: "D"},
	}
	tiers, err := topoSort(nodes, edges)
	require.NoError(t, err)
	require.Len(t, tiers, 2)
	assert.ElementsMatch(t, []string{"A", "B"}, tiers[0])
	assert.ElementsMatch(t, []string{"C", "D"}, tiers[1])
}

func TestTopoSort_Cycle(t *testing.T) {
	nodes := []WorkflowNode{{ID: "A"}, {ID: "B"}}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
		{ID: "e2", Source: "B", Target: "A"},
	}
	_, err := topoSort(nodes, edges)
	assert.True(t, errors.Is(err, ErrCyclicWorkflow))
}

func TestTopoSort_Diamond(t *testing.T) {
	//   A
	//  / \
	// B   C
	//  \ /
	//   D
	nodes := []WorkflowNode{{ID: "A"}, {ID: "B"}, {ID: "C"}, {ID: "D"}}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
		{ID: "e2", Source: "A", Target: "C"},
		{ID: "e3", Source: "B", Target: "D"},
		{ID: "e4", Source: "C", Target: "D"},
	}
	tiers, err := topoSort(nodes, edges)
	require.NoError(t, err)
	require.Len(t, tiers, 3)
	assert.Equal(t, []string{"A"}, tiers[0])
	assert.ElementsMatch(t, []string{"B", "C"}, tiers[1])
	assert.Equal(t, []string{"D"}, tiers[2])
}

// conditionRunner returns a CommandRunner for condition/merge tests. The
// outputs map is keyed by the slugified node label that appears in
// BuildSessionName (e.g. "Process A" → "process-a"); the runner parses
// each capture-pane target via ParseSessionName and returns the matching
// output.
func conditionRunner(outputs map[string]string) CommandRunner {
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			if label, ok := sessionLabelFromArgs(args); ok {
				if out, found := outputs[label]; found {
					return []byte(out), nil
				}
			}
			return []byte("default output"), nil
		}
		if len(args) > 0 && args[0] == "new-session" {
			return []byte("ok"), nil
		}
		if len(args) > 0 && args[0] == "list-panes" {
			return []byte("1\n"), nil
		}
		return []byte("ok"), nil
	}
}

// sessionLabelFromArgs scans a tmux argv list for a "-t {session}:0.0"
// target and returns the label field recovered by ParseSessionName.
func sessionLabelFromArgs(args []string) (string, bool) {
	target := ""
	for i, a := range args {
		if a == "-t" && i+1 < len(args) {
			target = args[i+1]
		}
	}
	session := strings.TrimSuffix(target, ":0.0")
	_, _, label, _, ok := ParseSessionName(session)
	return label, ok
}

func TestDynamicExecutor_ConditionBranching_TrueBranch(t *testing.T) {
	// Workflow: A(process) -> B(condition, contains "SUCCESS") -> C(true) and D(false).
	// A outputs "SUCCESS". Verify C completes, D is skipped.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"process-a": "operation SUCCESS complete",
	}))

	wf := WorkflowDef{
		ID:   "wf-cond-true",
		Name: "Condition True Branch",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", NodeType: NodeTypeCondition, Label: "Check", Config: map[string]string{
				"condition": `{"type":"contains","pattern":"SUCCESS","sourceNode":"A"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "C", ProcessID: autonomousProcessFixtureID, Label: "True Path", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: -100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: autonomousProcessFixtureID, Label: "False Path", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: 100}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "C", SourceHandle: "true"},
			{ID: "e3", Source: "B", Target: "D", SourceHandle: "false"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-cond-true", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)

	nodeStatus := make(map[string]WorkflowNodeStatus)
	for _, n := range ex.Nodes {
		nodeStatus[n.ID] = n.Status
	}
	assert.Equal(t, NodeComplete, nodeStatus["A"], "A should complete")
	assert.Equal(t, NodeComplete, nodeStatus["B"], "B (condition) should complete")
	assert.Equal(t, NodeComplete, nodeStatus["C"], "C (true branch) should complete")
	assert.Equal(t, NodeSkipped, nodeStatus["D"], "D (false branch) should be skipped")

	// Condition node should store its result in NodeOutputs.
	assert.Equal(t, "true", ex.NodeOutputs["B"])
}

func TestDynamicExecutor_ConditionBranching_FalseBranch(t *testing.T) {
	// Same topology but A outputs "FAILURE" — condition evaluates false.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"process-a": "operation FAILURE complete",
	}))

	wf := WorkflowDef{
		ID:   "wf-cond-false",
		Name: "Condition False Branch",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", NodeType: NodeTypeCondition, Label: "Check", Config: map[string]string{
				"condition": `{"type":"contains","pattern":"SUCCESS","sourceNode":"A"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "C", ProcessID: autonomousProcessFixtureID, Label: "True Path", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: -100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: autonomousProcessFixtureID, Label: "False Path", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: 100}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "C", SourceHandle: "true"},
			{ID: "e3", Source: "B", Target: "D", SourceHandle: "false"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-cond-false", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)

	nodeStatus := make(map[string]WorkflowNodeStatus)
	for _, n := range ex.Nodes {
		nodeStatus[n.ID] = n.Status
	}
	assert.Equal(t, NodeComplete, nodeStatus["A"], "A should complete")
	assert.Equal(t, NodeComplete, nodeStatus["B"], "B (condition) should complete")
	assert.Equal(t, NodeSkipped, nodeStatus["C"], "C (true branch) should be skipped")
	assert.Equal(t, NodeComplete, nodeStatus["D"], "D (false branch) should complete")

	assert.Equal(t, "false", ex.NodeOutputs["B"])
}

func TestDynamicExecutor_MergeAfterCondition(t *testing.T) {
	// Diamond: A(condition,always true) -> B(true), A -> C(false), B -> D(merge), C -> D(merge).
	// Verify B runs, C skipped, D (merge) runs.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"upstream": "has SUCCESS in output", // label "Upstream" → slug "upstream"
	}))

	wf := WorkflowDef{
		ID:   "wf-merge",
		Name: "Merge After Condition",
		Nodes: []WorkflowNode{
			{ID: "upstream", ProcessID: autonomousProcessFixtureID, Label: "Upstream", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "A", NodeType: NodeTypeCondition, Label: "Condition A", Config: map[string]string{
				"condition": `{"type":"contains","pattern":"SUCCESS","sourceNode":"upstream"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "True Path", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: -100}, Status: NodePending, Config: map[string]string{}},
			{ID: "C", ProcessID: autonomousProcessFixtureID, Label: "False Path", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: 100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", NodeType: NodeTypeMerge, Label: "Merge", Config: map[string]string{},
				Position: Position{X: 750, Y: 0}, Status: NodePending},
		},
		Edges: []WorkflowEdge{
			{ID: "e0", Source: "upstream", Target: "A"},
			{ID: "e1", Source: "A", Target: "B", SourceHandle: "true"},
			{ID: "e2", Source: "A", Target: "C", SourceHandle: "false"},
			{ID: "e3", Source: "B", Target: "D"},
			{ID: "e4", Source: "C", Target: "D"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-merge", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)

	nodeStatus := make(map[string]WorkflowNodeStatus)
	for _, n := range ex.Nodes {
		nodeStatus[n.ID] = n.Status
	}
	assert.Equal(t, NodeComplete, nodeStatus["upstream"], "upstream should complete")
	assert.Equal(t, NodeComplete, nodeStatus["A"], "A (condition) should complete")
	assert.Equal(t, NodeComplete, nodeStatus["B"], "B (true path) should complete")
	assert.Equal(t, NodeSkipped, nodeStatus["C"], "C (false path) should be skipped")
	assert.Equal(t, NodeComplete, nodeStatus["D"], "D (merge) should complete")
}

func TestDynamicExecutor_AllBranchesSkipped(t *testing.T) {
	// A(process) -> B(condition, contains "MAGIC") -> C(true only).
	// A outputs "nothing special". B evaluates false. C should be skipped.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"process-a": "nothing special here",
	}))

	wf := WorkflowDef{
		ID:   "wf-allskip",
		Name: "All Branches Skipped",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", NodeType: NodeTypeCondition, Label: "Check MAGIC", Config: map[string]string{
				"condition": `{"type":"contains","pattern":"MAGIC","sourceNode":"A"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "C", ProcessID: autonomousProcessFixtureID, Label: "True Only", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "C", SourceHandle: "true"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-allskip", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)

	nodeStatus := make(map[string]WorkflowNodeStatus)
	for _, n := range ex.Nodes {
		nodeStatus[n.ID] = n.Status
	}
	assert.Equal(t, NodeComplete, nodeStatus["A"])
	assert.Equal(t, NodeComplete, nodeStatus["B"])
	assert.Equal(t, NodeSkipped, nodeStatus["C"], "C should be skipped when condition is false and no false branch")
}

func TestDynamicExecutor_SkippedStatus(t *testing.T) {
	// Verify that skipped nodes emit NodeSkipped status events.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"process-a": "no match",
	}))

	wf := WorkflowDef{
		ID:   "wf-skipevt",
		Name: "Skipped Events",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", NodeType: NodeTypeCondition, Label: "Check", Config: map[string]string{
				"condition": `{"type":"contains","pattern":"TRIGGER","sourceNode":"A"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "C", ProcessID: autonomousProcessFixtureID, Label: "True", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: -100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: autonomousProcessFixtureID, Label: "False", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: 100}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "C", SourceHandle: "true"},
			{ID: "e3", Source: "B", Target: "D", SourceHandle: "false"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-skipevt", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	// Verify NodeSkipped events were emitted.
	nodeEvents := h.eventsByName("bmad:node:status")
	var skippedNodeIDs []string
	for _, ev := range nodeEvents {
		if se, ok := ev.data.(NodeStatusEvent); ok && se.Status == NodeSkipped {
			skippedNodeIDs = append(skippedNodeIDs, se.NodeID)
		}
	}
	// Condition evaluates false: C (true branch) should be skipped.
	// D (false branch) should complete.
	assert.Contains(t, skippedNodeIDs, "C", "C should have NodeSkipped event")
	assert.NotContains(t, skippedNodeIDs, "D", "D should NOT have NodeSkipped event")
}

// loopRunner creates a CommandRunner that returns different capture-pane output
// per iteration. It tracks how many times capture-pane is called for the body
// node (identified by the slugified label "b" parsed from the session name).
// The iterOutputs map is 1-indexed: iterOutputs[1] is the output for the first
// execution of the body node.
func loopRunner(iterOutputs map[int]string) CommandRunner {
	var bodyCaptures int32
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			label, ok := sessionLabelFromArgs(args)
			if ok && label == "b" {
				iter := int(atomic.AddInt32(&bodyCaptures, 1))
				if out, ok := iterOutputs[iter]; ok {
					return []byte(out), nil
				}
				return []byte(fmt.Sprintf("iteration-%d", iter)), nil
			}
			return []byte("non-body-output"), nil
		}
		if len(args) > 0 && args[0] == "new-session" {
			return []byte("ok"), nil
		}
		if len(args) > 0 && args[0] == "list-panes" {
			return []byte("1\n"), nil
		}
		return []byte("ok"), nil
	}
}

func TestLoopNode_FixedCount(t *testing.T) {
	// Workflow: A(process) -> L(loop, maxIterations=3, body=[B]) -> D(process).
	// Verify: B executes 3 times, D executes once after loop, NodeOutputs has iter count "3".
	h := newHarness(t)
	h.executor.SetCommandRunner(loopRunner(map[int]string{
		1: "iter-1-output",
		2: "iter-2-output",
		3: "iter-3-output",
	}))

	wf := WorkflowDef{
		ID:   "wf-loop-fixed",
		Name: "Loop Fixed Count",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "L", NodeType: NodeTypeLoop, Label: "Loop", Config: map[string]string{
				"maxIterations": "3",
				"loopBodyNodes": "B",
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", NodeType: NodeTypeProcess,
				Position: Position{X: 400, Y: 100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: autonomousProcessFixtureID, Label: "D", NodeType: NodeTypeProcess,
				Position: Position{X: 650, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "L"},
			{ID: "e2", Source: "L", Target: "B", SourceHandle: "loop-body"},
			{ID: "e3", Source: "L", Target: "D", SourceHandle: "loop-exit"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-loop-fixed", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 10*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)

	nodeStatus := make(map[string]WorkflowNodeStatus)
	for _, n := range ex.Nodes {
		nodeStatus[n.ID] = n.Status
	}
	assert.Equal(t, NodeComplete, nodeStatus["A"], "A should complete")
	assert.Equal(t, NodeComplete, nodeStatus["L"], "L (loop) should complete")
	assert.Equal(t, NodeComplete, nodeStatus["B"], "B (body) should complete (last iteration)")
	assert.Equal(t, NodeComplete, nodeStatus["D"], "D (exit) should complete after loop")

	// Verify iteration count stored in NodeOutputs.
	assert.Equal(t, "3", ex.NodeOutputs["L_iter"], "should store final iteration count")
}

func TestLoopUntil_ConditionMet(t *testing.T) {
	// LoopUntil with condition checking for "DONE" in body node output.
	// Body outputs "working" on iter 1, "DONE" on iter 2.
	// Verify: exits after 2 iterations, not 10.
	h := newHarness(t)
	h.executor.SetCommandRunner(loopRunner(map[int]string{
		1: "still working",
		2: "task DONE successfully",
	}))

	wf := WorkflowDef{
		ID:   "wf-loopuntil-met",
		Name: "LoopUntil Condition Met",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "LU", NodeType: NodeTypeLoopUntil, Label: "LoopUntil", Config: map[string]string{
				"maxIterations": "10",
				"loopBodyNodes": "B",
				"condition":     `{"type":"contains","pattern":"DONE","sourceNode":"B"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", NodeType: NodeTypeProcess,
				Position: Position{X: 400, Y: 100}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "LU"},
			{ID: "e2", Source: "LU", Target: "B", SourceHandle: "loop-body"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-loopuntil-met", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 10*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)

	// Verify exited after exactly 2 iterations.
	assert.Equal(t, "2", ex.NodeOutputs["LU_iter"], "should exit after 2 iterations when condition met")
}

func TestLoopUntil_MaxIterations(t *testing.T) {
	// LoopUntil with condition that never matches. maxIterations=3.
	// Verify: exits after 3 iterations.
	h := newHarness(t)
	h.executor.SetCommandRunner(loopRunner(map[int]string{
		1: "nope",
		2: "still nope",
		3: "never matches",
	}))

	wf := WorkflowDef{
		ID:   "wf-loopuntil-max",
		Name: "LoopUntil Max Iterations",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "LU", NodeType: NodeTypeLoopUntil, Label: "LoopUntil", Config: map[string]string{
				"maxIterations": "3",
				"loopBodyNodes": "B",
				"condition":     `{"type":"contains","pattern":"NEVER_MATCH_THIS","sourceNode":"B"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", NodeType: NodeTypeProcess,
				Position: Position{X: 400, Y: 100}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "LU"},
			{ID: "e2", Source: "LU", Target: "B", SourceHandle: "loop-body"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-loopuntil-max", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 10*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)

	// Should have hit max iterations.
	assert.Equal(t, "3", ex.NodeOutputs["LU_iter"], "should exhaust all 3 iterations")
}

func TestLoop_BodyFailure(t *testing.T) {
	// Loop body node fails on iteration 2. Verify loop node fails.
	h := newHarness(t)

	var newSessionCount int32
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "new-session" {
			count := int(atomic.AddInt32(&newSessionCount, 1))
			// First new-session is A. Second is B iter 1. Third is B iter 2 — fail it.
			if count == 3 {
				return nil, fmt.Errorf("tmux failed on iteration 2")
			}
			return []byte("ok"), nil
		}
		if len(args) > 0 && args[0] == "capture-pane" {
			return []byte("body output"), nil
		}
		if len(args) > 0 && args[0] == "list-panes" {
			return []byte("1\n"), nil
		}
		return []byte("ok"), nil
	})

	wf := WorkflowDef{
		ID:   "wf-loop-fail",
		Name: "Loop Body Failure",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "L", NodeType: NodeTypeLoop, Label: "Loop", Config: map[string]string{
				"maxIterations": "5",
				"loopBodyNodes": "B",
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", NodeType: NodeTypeProcess,
				Position: Position{X: 400, Y: 100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: autonomousProcessFixtureID, Label: "D", NodeType: NodeTypeProcess,
				Position: Position{X: 650, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "L"},
			{ID: "e2", Source: "L", Target: "B", SourceHandle: "loop-body"},
			{ID: "e3", Source: "L", Target: "D", SourceHandle: "loop-exit"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-loop-fail", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	// Execution should pause after failure (loop node fails -> pause).
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecPaused
	}, 10*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)

	nodeStatus := make(map[string]WorkflowNodeStatus)
	for _, n := range ex.Nodes {
		nodeStatus[n.ID] = n.Status
	}
	assert.Equal(t, NodeComplete, nodeStatus["A"], "A should complete before loop")
	assert.Equal(t, NodeFailed, nodeStatus["L"], "L (loop) should fail when body fails")
	assert.Equal(t, NodePending, nodeStatus["D"], "D should remain pending when loop fails")

	// Cleanup.
	h.executor.StopWorkflow(exec.ID)
}

func TestLoop_EmptyBody(t *testing.T) {
	// Loop with empty loopBodyNodes. Should complete immediately with 0 iterations.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{}))

	wf := WorkflowDef{
		ID:   "wf-loop-empty",
		Name: "Loop Empty Body",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "L", NodeType: NodeTypeLoop, Label: "Loop", Config: map[string]string{
				"maxIterations": "5",
				"loopBodyNodes": "",
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "D", ProcessID: autonomousProcessFixtureID, Label: "D", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "L"},
			{ID: "e2", Source: "L", Target: "D", SourceHandle: "loop-exit"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-loop-empty", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)

	nodeStatus := make(map[string]WorkflowNodeStatus)
	for _, n := range ex.Nodes {
		nodeStatus[n.ID] = n.Status
	}
	assert.Equal(t, NodeComplete, nodeStatus["A"], "A should complete")
	assert.Equal(t, NodeComplete, nodeStatus["L"], "L should complete with empty body")
	assert.Equal(t, NodeComplete, nodeStatus["D"], "D should complete after empty loop")

	// Should store 0 iterations.
	assert.Equal(t, "0", ex.NodeOutputs["L_iter"], "empty body should report 0 iterations")
}

func TestLoopNode_IteratesOverItems(t *testing.T) {
	// Loop with 3 items and maxIterations=10. Should iterate exactly 3 times.
	h := newHarness(t)
	h.executor.SetCommandRunner(loopRunner(map[int]string{
		1: "output-a",
		2: "output-b",
		3: "output-c",
	}))

	wf := WorkflowDef{
		ID:   "wf-loop-items",
		Name: "Loop Items",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "L", NodeType: NodeTypeLoop, Label: "Loop", Config: map[string]string{
				"maxIterations": "10",
				"loopBodyNodes": "B",
				"items":         `["alpha","beta","gamma"]`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", NodeType: NodeTypeProcess,
				Position: Position{X: 400, Y: 100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: autonomousProcessFixtureID, Label: "D", NodeType: NodeTypeProcess,
				Position: Position{X: 650, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "L"},
			{ID: "e2", Source: "L", Target: "B", SourceHandle: "loop-body"},
			{ID: "e3", Source: "L", Target: "D", SourceHandle: "loop-exit"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-loop-items", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 10*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)

	// Should iterate 3 times (items length), not 10.
	assert.Equal(t, "3", ex.NodeOutputs["L_iter"])
	// Last item should be "gamma".
	assert.Equal(t, "gamma", ex.NodeOutputs["L_item"])
	// Full items array stored as node output.
	assert.Equal(t, `["alpha","beta","gamma"]`, ex.NodeOutputs["L"])
}

func TestLoopNode_ItemsCappedByMaxIterations(t *testing.T) {
	// 5 items but maxIterations=2. Should only iterate 2 times.
	h := newHarness(t)
	h.executor.SetCommandRunner(loopRunner(map[int]string{
		1: "out-1",
		2: "out-2",
	}))

	wf := WorkflowDef{
		ID:   "wf-loop-items-capped",
		Name: "Loop Items Capped",
		Nodes: []WorkflowNode{
			{ID: "L", NodeType: NodeTypeLoop, Label: "Loop", Config: map[string]string{
				"maxIterations": "2",
				"loopBodyNodes": "B",
				"items":         `["a","b","c","d","e"]`,
			}, Position: Position{X: 0, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", NodeType: NodeTypeProcess,
				Position: Position{X: 200, Y: 100}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "L", Target: "B", SourceHandle: "loop-body"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-loop-items-capped", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 10*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)

	// Only 2 iterations despite 5 items.
	assert.Equal(t, "2", ex.NodeOutputs["L_iter"])
	assert.Equal(t, "b", ex.NodeOutputs["L_item"])
}

func TestLoopNode_InvalidItemsJSON(t *testing.T) {
	// Malformed items JSON should fall back to counter-based iteration.
	h := newHarness(t)
	h.executor.SetCommandRunner(loopRunner(map[int]string{
		1: "out-1",
		2: "out-2",
		3: "out-3",
	}))

	wf := WorkflowDef{
		ID:   "wf-loop-bad-items",
		Name: "Loop Bad Items",
		Nodes: []WorkflowNode{
			{ID: "L", NodeType: NodeTypeLoop, Label: "Loop", Config: map[string]string{
				"maxIterations": "3",
				"loopBodyNodes": "B",
				"items":         `not valid json`,
			}, Position: Position{X: 0, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", NodeType: NodeTypeProcess,
				Position: Position{X: 200, Y: 100}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "L", Target: "B", SourceHandle: "loop-body"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-loop-bad-items", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 10*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)

	// Should fall back to counter: 3 iterations, no _item output.
	assert.Equal(t, "3", ex.NodeOutputs["L_iter"])
	assert.Empty(t, ex.NodeOutputs["L_item"], "no _item when items is invalid JSON")
}

func TestLoopNode_EmptyItems(t *testing.T) {
	// Empty items array "[]" should iterate 0 times (maxIter stays 10 but items caps to 0).
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())

	wf := WorkflowDef{
		ID:   "wf-loop-empty-items",
		Name: "Loop Empty Items",
		Nodes: []WorkflowNode{
			{ID: "L", NodeType: NodeTypeLoop, Label: "Loop", Config: map[string]string{
				"maxIterations": "10",
				"loopBodyNodes": "B",
				"items":         `[]`,
			}, Position: Position{X: 0, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", NodeType: NodeTypeProcess,
				Position: Position{X: 200, Y: 100}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "L", Target: "B", SourceHandle: "loop-body"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-loop-empty-items", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 10*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)

	// Empty items: iterateItems is false (len == 0), so falls back to counter.
	// With 10 iterations and a body, it should run 10 times.
	assert.Equal(t, "10", ex.NodeOutputs["L_iter"])
	assert.Empty(t, ex.NodeOutputs["L_item"])
}

// cmdCall captures a single runCmd invocation for later assertions.
type cmdCall struct {
	name string
	args []string
}
