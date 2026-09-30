package bmad

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStartWorkflow_Sequential(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())
	wfID := saveThreeNodeWorkflow(t, h.storage)

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, "/tmp/repo", "sonnet")
	require.NoError(t, err)
	assert.Equal(t, ExecRunning, exec.Status)

	// Wait for completion.
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)
	for _, n := range ex.Nodes {
		assert.Equal(t, NodeComplete, n.Status, "node %s should be complete", n.ID)
	}

	// Check events: 3 running + 3 complete = 6 node events.
	nodeEvents := h.eventsByName("bmad:node:status")
	assert.GreaterOrEqual(t, len(nodeEvents), 6) // may have extra tmuxTarget events
}

func TestStartWorkflow_NodeFailure(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(failRunner())
	wfID := saveThreeNodeWorkflow(t, h.storage)

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, "/tmp/repo", "sonnet")
	require.NoError(t, err)

	// Execution should pause after failure.
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecPaused
	}, 5*time.Second, 50*time.Millisecond)

	ex, _ := h.executor.GetExecution(exec.ID)
	assert.Equal(t, NodeFailed, ex.Nodes[0].Status)
	assert.Equal(t, NodePending, ex.Nodes[1].Status)
	assert.Equal(t, NodePending, ex.Nodes[2].Status)
}

func TestStartWorkflow_CyclicWorkflow(t *testing.T) {
	h := newHarness(t)
	wf := WorkflowDef{
		ID:   "wf-cycle",
		Name: "Cyclic",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", Config: map[string]string{}},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "A"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	_, err := h.executor.StartWorkflow(context.Background(), "wf-cycle", "/tmp", "sonnet")
	assert.True(t, errors.Is(err, ErrCyclicWorkflow))
}

func TestStartWorkflow_ParallelBranches(t *testing.T) {
	h := newHarness(t)

	var spawned int32
	var spawnMu sync.Mutex
	spawnTimes := make(map[string]time.Time)

	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "new-session" {
			atomic.AddInt32(&spawned, 1)
			for i, a := range args {
				if a == "-s" && i+1 < len(args) {
					spawnMu.Lock()
					spawnTimes[args[i+1]] = time.Now()
					spawnMu.Unlock()
				}
			}
			return []byte("ok"), nil
		}
		// list-panes: complete immediately
		return []byte("1\n"), nil
	})

	wf := WorkflowDef{
		ID:   "wf-parallel",
		Name: "Parallel",
		Nodes: []WorkflowNode{
			// Autonomous processes only — bmad-domain-research is now interactive.
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", Config: map[string]string{}},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", Config: map[string]string{}},
			{ID: "C", ProcessID: autonomousProcessFixtureID, Label: "C", Config: map[string]string{}},
			{ID: "D", ProcessID: autonomousProcessFixtureID, Label: "D", Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "C"},
			{ID: "e2", Source: "B", Target: "D"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-parallel", "/tmp", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	// All 4 nodes should have been spawned.
	assert.GreaterOrEqual(t, int(atomic.LoadInt32(&spawned)), 4)
}

func TestPauseAndResume(t *testing.T) {
	h := newHarness(t)

	// Use a delay runner so we have time to pause mid-execution.
	h.executor.SetCommandRunner(delayRunner(200 * time.Millisecond))

	wfID := saveThreeNodeWorkflow(t, h.storage)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, "/tmp/repo", "sonnet")
	require.NoError(t, err)

	// Wait for first node to start running.
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		if ex == nil {
			return false
		}
		return ex.Nodes[0].Status == NodeRunning || ex.Nodes[0].Status == NodeComplete
	}, 5*time.Second, 20*time.Millisecond)

	// Pause — only blocks new nodes, current one finishes.
	require.NoError(t, h.executor.PauseWorkflow(exec.ID))

	// Wait for first node to complete.
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Nodes[0].Status == NodeComplete
	}, 5*time.Second, 50*time.Millisecond)

	// Second node should still be pending (paused).
	ex, _ := h.executor.GetExecution(exec.ID)
	assert.Equal(t, ExecPaused, ex.Status)
	assert.Equal(t, NodePending, ex.Nodes[1].Status)

	// Resume.
	require.NoError(t, h.executor.ResumeWorkflow(exec.ID))

	// Should eventually complete.
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 10*time.Second, 50*time.Millisecond)
}

func TestStopWorkflow(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(delayRunner(10 * time.Second)) // long delay

	wfID := saveThreeNodeWorkflow(t, h.storage)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, "/tmp/repo", "sonnet")
	require.NoError(t, err)

	// Wait for first node to start.
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Nodes[0].Status == NodeRunning
	}, 5*time.Second, 20*time.Millisecond)

	require.NoError(t, h.executor.StopWorkflow(exec.ID))

	ex, _ := h.executor.GetExecution(exec.ID)
	assert.Equal(t, ExecFailed, ex.Status)

	// Check event was emitted.
	execEvents := h.eventsByName("bmad:execution:status")
	var foundFailed bool
	for _, ev := range execEvents {
		if se, ok := ev.data.(ExecStatusEvent); ok && se.Status == ExecFailed {
			foundFailed = true
		}
	}
	assert.True(t, foundFailed, "should have emitted ExecFailed event")
}

func TestPauseWorkflow_NotRunning(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())
	wfID := saveThreeNodeWorkflow(t, h.storage)

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, "/tmp", "sonnet")
	require.NoError(t, err)

	// Wait for completion.
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	err = h.executor.PauseWorkflow(exec.ID)
	assert.True(t, errors.Is(err, ErrExecNotRunning))
}

func TestResumeWorkflow_NotPaused(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(delayRunner(5 * time.Second))
	wfID := saveThreeNodeWorkflow(t, h.storage)

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, "/tmp", "sonnet")
	require.NoError(t, err)

	err = h.executor.ResumeWorkflow(exec.ID)
	assert.True(t, errors.Is(err, ErrExecNotPaused))

	// Cleanup.
	h.executor.StopWorkflow(exec.ID)
}

func TestGetExecution_NotFound(t *testing.T) {
	h := newHarness(t)
	_, err := h.executor.GetExecution("nonexistent")
	assert.True(t, errors.Is(err, ErrExecNotFound))
}

func TestStartWorkflow_WorkflowNotFound(t *testing.T) {
	h := newHarness(t)
	_, err := h.executor.StartWorkflow(context.Background(), "nonexistent", "/tmp", "sonnet")
	assert.True(t, errors.Is(err, ErrWorkflowNotFound))
}

func TestGetExecution_CopiesNodeOutputs(t *testing.T) {
	h := newHarness(t)

	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			return []byte("captured text"), nil
		}
		return successRunner()(ctx, name, args...)
	})

	wf := WorkflowDef{
		ID:   "wf-copycap",
		Name: "Copy Capture",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-copycap", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	a, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	b, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)

	// Mutate copy a's NodeOutputs.
	a.NodeOutputs["A"] = "mutated"
	// b should be unaffected.
	assert.Equal(t, "captured text", b.NodeOutputs["A"])
	assert.NotEqual(t, a.NodeOutputs["A"], b.NodeOutputs["A"])
}

func TestGetExecution_ReturnsCopy(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())
	wfID := saveThreeNodeWorkflow(t, h.storage)

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, "/tmp", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	a, _ := h.executor.GetExecution(exec.ID)
	b, _ := h.executor.GetExecution(exec.ID)
	a.Nodes[0].Label = "mutated"
	assert.NotEqual(t, a.Nodes[0].Label, b.Nodes[0].Label)
}

// seedExecState registers a minimal execState directly in the executor,
// bypassing the runner goroutine. Tests use this to drive
// GetCurrentExecution through deterministic scenarios (no wall-clock
// waiting, no background command execution).
func seedExecState(t *testing.T, h *testHarness, execID, repoPath string, status WorkflowExecStatus, startedAt string) {
	t.Helper()
	state := &execState{
		exec: &WorkflowExecution{
			ID:          execID,
			WorkflowID:  "wf-" + execID,
			RepoPath:    repoPath,
			Status:      status,
			StartedAt:   startedAt,
			NodeOutputs: map[string]string{},
			Nodes: []WorkflowNode{
				{ID: "node-A", Label: "A", Status: NodeRunning},
			},
		},
		cancel:           func() {},
		inDegree:         map[string]int{"node-A": 0},
		outEdges:         map[string][]WorkflowEdge{},
		lastQuestionHash: map[string]string{},
		lastOutputHash:   map[string]string{},
		idleEmitted:      map[string]bool{},
	}
	h.executor.mu.Lock()
	h.executor.executions[execID] = state
	h.executor.mu.Unlock()
}

func TestGetCurrentExecution_EmptyRepoPath(t *testing.T) {
	h := newHarness(t)
	seedExecState(t, h, "e1", "/tmp/repoA", ExecRunning, "2026-04-11T10:00:00Z")

	got, err := h.executor.GetCurrentExecution("")
	require.NoError(t, err, "empty repoPath must not be an error")
	assert.Nil(t, got, "empty repoPath must return nil — no ambiguous global match")
}

func TestGetCurrentExecution_NoExecutions(t *testing.T) {
	h := newHarness(t)

	got, err := h.executor.GetCurrentExecution("/tmp/repoA")
	require.NoError(t, err, "missing execution is not an error")
	assert.Nil(t, got, "empty executions map must return nil")
}

func TestGetCurrentExecution_NoMatchingRepo(t *testing.T) {
	h := newHarness(t)
	seedExecState(t, h, "e1", "/tmp/repoA", ExecRunning, "2026-04-11T10:00:00Z")
	seedExecState(t, h, "e2", "/tmp/repoB", ExecRunning, "2026-04-11T10:00:00Z")

	got, err := h.executor.GetCurrentExecution("/tmp/repoC")
	require.NoError(t, err)
	assert.Nil(t, got, "unrelated running executions must not match the query")
}

func TestGetCurrentExecution_SingleRunningMatch(t *testing.T) {
	h := newHarness(t)
	seedExecState(t, h, "e1", "/tmp/repoA", ExecRunning, "2026-04-11T10:00:00Z")

	got, err := h.executor.GetCurrentExecution("/tmp/repoA")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "e1", got.ID)
	assert.Equal(t, ExecRunning, got.Status)
}

func TestGetCurrentExecution_PausedCountsAsNonTerminal(t *testing.T) {
	h := newHarness(t)
	seedExecState(t, h, "e1", "/tmp/repoA", ExecPaused, "2026-04-11T10:00:00Z")

	got, err := h.executor.GetCurrentExecution("/tmp/repoA")
	require.NoError(t, err)
	require.NotNil(t, got, "paused exec is still restorable state")
	assert.Equal(t, ExecPaused, got.Status)
}

func TestGetCurrentExecution_IgnoresTerminalExecutions(t *testing.T) {
	h := newHarness(t)
	seedExecState(t, h, "e-done", "/tmp/repoA", ExecComplete, "2026-04-11T10:00:00Z")
	seedExecState(t, h, "e-failed", "/tmp/repoA", ExecFailed, "2026-04-11T11:00:00Z")

	got, err := h.executor.GetCurrentExecution("/tmp/repoA")
	require.NoError(t, err)
	assert.Nil(t, got,
		"complete and failed executions must never match — nothing live to restore")
}

func TestGetCurrentExecution_PicksLatestByStartedAt(t *testing.T) {
	h := newHarness(t)
	seedExecState(t, h, "e-old", "/tmp/repoA", ExecRunning, "2026-04-11T08:00:00Z")
	seedExecState(t, h, "e-mid", "/tmp/repoA", ExecPaused, "2026-04-11T09:30:00Z")
	seedExecState(t, h, "e-new", "/tmp/repoA", ExecRunning, "2026-04-11T12:00:00Z")

	got, err := h.executor.GetCurrentExecution("/tmp/repoA")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "e-new", got.ID,
		"when multiple non-terminal executions match, pick the one with the latest StartedAt")
}

func TestGetCurrentExecution_ReturnsDeepCopy(t *testing.T) {
	h := newHarness(t)
	seedExecState(t, h, "e1", "/tmp/repoA", ExecRunning, "2026-04-11T10:00:00Z")

	got, err := h.executor.GetCurrentExecution("/tmp/repoA")
	require.NoError(t, err)
	require.NotNil(t, got)
	got.Nodes[0].Label = "mutated-by-caller"

	// Fetch again — the caller's mutation must NOT leak into the stored
	// state. This is the same invariant as TestGetExecution_ReturnsCopy,
	// applied to the new restore path.
	fresh, err := h.executor.GetCurrentExecution("/tmp/repoA")
	require.NoError(t, err)
	require.NotNil(t, fresh)
	assert.NotEqual(t, "mutated-by-caller", fresh.Nodes[0].Label,
		"stored execState.exec.Nodes must not be observably mutated by caller")
}

// responseRunner returns a CommandRunner that records every call and answers
// list-panes pane_dead checks with the given value ("0" alive, "1" dead, "" to
// simulate a tmux error).
func responseRunner(paneDead string) (CommandRunner, *[]cmdCall, *sync.Mutex) {
	var calls []cmdCall
	var mu sync.Mutex
	runner := CommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		mu.Lock()
		calls = append(calls, cmdCall{name: name, args: append([]string(nil), args...)})
		mu.Unlock()
		if len(args) > 0 && args[0] == "list-panes" {
			if paneDead == "" {
				return nil, fmt.Errorf("tmux: server not running")
			}
			return []byte(paneDead + "\n"), nil
		}
		return []byte("ok"), nil
	})
	return runner, &calls, &mu
}

// findCall returns the first captured call matching name and the given leading
// positional args, or nil if not found.
func findCall(calls []cmdCall, name string, leading ...string) *cmdCall {
	for i := range calls {
		c := calls[i]
		if c.name != name || len(c.args) < len(leading) {
			continue
		}
		match := true
		for j, a := range leading {
			if c.args[j] != a {
				match = false
				break
			}
		}
		if match {
			return &c
		}
	}
	return nil
}

// seedResponseState registers an execState with one node under the given
// execID, mirroring what StartWorkflow would do but without triggering the
// runner goroutine. It pre-populates lastQuestionHash so AC-4 can be verified.
func seedResponseState(t *testing.T, h *testHarness, execID, nodeID, tmuxTarget, hash string) *execState {
	t.Helper()
	state := &execState{
		exec: &WorkflowExecution{
			ID:     execID,
			Status: ExecRunning,
			Nodes: []WorkflowNode{
				{ID: nodeID, Label: nodeID, TmuxTarget: tmuxTarget, Status: NodeRunning},
			},
			NodeOutputs: map[string]string{},
		},
		cancel:           func() {},
		inDegree:         map[string]int{nodeID: 0},
		outEdges:         map[string][]WorkflowEdge{},
		lastQuestionHash: map[string]string{nodeID: hash},
		lastOutputHash:   map[string]string{},
		idleEmitted:      map[string]bool{},
	}
	h.executor.mu.Lock()
	h.executor.executions[execID] = state
	h.executor.mu.Unlock()
	return state
}

// AC-1: Response is injected into the correct tmux pane via two send-keys calls.
func TestStory2_AC1_RespondToQuestion_Success(t *testing.T) {
	h := newHarness(t)
	runner, calls, mu := responseRunner("0")
	h.executor.SetCommandRunner(runner)
	seedResponseState(t, h, "exec-1", "node-A", "bmad-node-A-100:0.0", "abc123")

	err := h.executor.RespondToQuestionLegacy("exec-1", "node-A", "src/main.go")
	require.NoError(t, err)

	mu.Lock()
	snapshot := append([]cmdCall(nil), *calls...)
	mu.Unlock()

	// Pane liveness check happened first.
	list := findCall(snapshot, "tmux", "list-panes", "-t", "bmad-node-A-100:0.0")
	require.NotNil(t, list, "expected list-panes call")

	// send-keys -l -t {target} {answer}
	literal := findCall(snapshot, "tmux", "send-keys", "-l", "-t", "bmad-node-A-100:0.0", "src/main.go")
	require.NotNil(t, literal, "expected literal send-keys call: %+v", snapshot)

	// send-keys -t {target} Enter
	enter := findCall(snapshot, "tmux", "send-keys", "-t", "bmad-node-A-100:0.0", "Enter")
	require.NotNil(t, enter, "expected Enter send-keys call: %+v", snapshot)
}

// AC-1 variant: a menu option number is sent literally (not interpreted).
func TestStory2_AC1_RespondToQuestion_MenuOption(t *testing.T) {
	h := newHarness(t)
	runner, calls, mu := responseRunner("0")
	h.executor.SetCommandRunner(runner)
	seedResponseState(t, h, "exec-1", "node-B", "bmad-node-B-200:0.0", "hashB")

	require.NoError(t, h.executor.RespondToQuestionLegacy("exec-1", "node-B", "2"))

	mu.Lock()
	defer mu.Unlock()
	require.NotNil(t, findCall(*calls, "tmux", "send-keys", "-l", "-t", "bmad-node-B-200:0.0", "2"))
	require.NotNil(t, findCall(*calls, "tmux", "send-keys", "-t", "bmad-node-B-200:0.0", "Enter"))
}

// AC-2: a dead pane returns an error wrapping ErrExecNotRunning and does NOT
// issue any send-keys calls.
func TestStory2_AC2_RespondToQuestion_DeadPane(t *testing.T) {
	h := newHarness(t)
	runner, calls, mu := responseRunner("1")
	h.executor.SetCommandRunner(runner)
	seedResponseState(t, h, "exec-1", "node-A", "bmad-node-A-100:0.0", "abc123")

	err := h.executor.RespondToQuestionLegacy("exec-1", "node-A", "answer")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrExecNotRunning),
		"expected error wrapping ErrExecNotRunning, got %v", err)

	mu.Lock()
	defer mu.Unlock()
	assert.Nil(t, findCall(*calls, "tmux", "send-keys"),
		"no send-keys call should be issued when pane is dead")
}

// AC-2 variant: pane check subprocess failure is treated as unreachable.
func TestStory2_AC2_RespondToQuestion_PaneCheckFails(t *testing.T) {
	h := newHarness(t)
	runner, calls, mu := responseRunner("") // empty => runner returns error
	h.executor.SetCommandRunner(runner)
	seedResponseState(t, h, "exec-1", "node-A", "bmad-node-A-100:0.0", "abc123")

	err := h.executor.RespondToQuestionLegacy("exec-1", "node-A", "answer")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrExecNotRunning),
		"expected error wrapping ErrExecNotRunning when tmux list-panes fails, got %v", err)

	mu.Lock()
	defer mu.Unlock()
	assert.Nil(t, findCall(*calls, "tmux", "send-keys"))
}

// AC-4: lastQuestionHash[nodeID] is cleared after a successful response.
func TestStory2_AC4_RespondToQuestion_ClearsHash(t *testing.T) {
	h := newHarness(t)
	runner, _, _ := responseRunner("0")
	h.executor.SetCommandRunner(runner)
	state := seedResponseState(t, h, "exec-1", "node-A", "bmad-node-A-100:0.0", "prev-hash")

	require.NoError(t, h.executor.RespondToQuestionLegacy("exec-1", "node-A", "answer"))

	state.mu.Lock()
	defer state.mu.Unlock()
	assert.Equal(t, "", state.lastQuestionHash["node-A"],
		"hash should be cleared so next poll can detect a new question")
}

// AC-4 variant: hash is NOT cleared when the pane is dead.
func TestStory2_AC4_RespondToQuestion_HashPreservedOnFailure(t *testing.T) {
	h := newHarness(t)
	runner, _, _ := responseRunner("1")
	h.executor.SetCommandRunner(runner)
	state := seedResponseState(t, h, "exec-1", "node-A", "bmad-node-A-100:0.0", "prev-hash")

	require.Error(t, h.executor.RespondToQuestionLegacy("exec-1", "node-A", "answer"))

	state.mu.Lock()
	defer state.mu.Unlock()
	assert.Equal(t, "prev-hash", state.lastQuestionHash["node-A"],
		"hash must survive a failed response so the UI retains the pending question")
}

// Missing execution returns ErrExecNotFound.
func TestStory2_RespondToQuestion_UnknownExec(t *testing.T) {
	h := newHarness(t)
	err := h.executor.RespondToQuestionLegacy("no-such-exec", "node-A", "hi")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrExecNotFound), "expected ErrExecNotFound, got %v", err)
}

// Missing node within a known exec returns an ErrExecNotFound-wrapped error.
func TestStory2_RespondToQuestion_UnknownNode(t *testing.T) {
	h := newHarness(t)
	runner, _, _ := responseRunner("0")
	h.executor.SetCommandRunner(runner)
	seedResponseState(t, h, "exec-1", "node-A", "bmad-node-A-100:0.0", "h")

	err := h.executor.RespondToQuestionLegacy("exec-1", "node-missing", "hi")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrExecNotFound), "expected ErrExecNotFound, got %v", err)
}

// Answers longer than maxAnswerBytes are rejected before touching tmux.
func TestStory2_RespondToQuestion_LongAnswer(t *testing.T) {
	h := newHarness(t)
	runner, calls, mu := responseRunner("0")
	h.executor.SetCommandRunner(runner)
	seedResponseState(t, h, "exec-1", "node-A", "bmad-node-A-100:0.0", "h")

	long := strings.Repeat("x", maxAnswerBytes+1)
	err := h.executor.RespondToQuestionLegacy("exec-1", "node-A", long)
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrAnswerTooLong),
		"expected ErrAnswerTooLong, got %v", err)
	assert.False(t, errors.Is(err, ErrExecNotRunning),
		"length errors must NOT wrap ErrExecNotRunning (misleading classification)")

	mu.Lock()
	defer mu.Unlock()
	assert.Empty(t, *calls, "over-length answer must short-circuit before any tmux call")
}

// A node with no TmuxTarget cannot receive a response.
func TestStory2_RespondToQuestion_NoTmuxTarget(t *testing.T) {
	h := newHarness(t)
	runner, calls, mu := responseRunner("0")
	h.executor.SetCommandRunner(runner)
	// Empty tmuxTarget signals the node was never scheduled.
	seedResponseState(t, h, "exec-1", "node-A", "", "h")

	err := h.executor.RespondToQuestionLegacy("exec-1", "node-A", "hi")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrExecNotRunning), "expected ErrExecNotRunning, got %v", err)

	mu.Lock()
	defer mu.Unlock()
	assert.Empty(t, *calls, "no tmux call should be made when target is empty")
}

// A send-keys -l subprocess failure is surfaced to the caller.
func TestStory2_RespondToQuestion_SendKeysLiteralFails(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "list-panes" {
			return []byte("0\n"), nil
		}
		if len(args) > 0 && args[0] == "send-keys" {
			return nil, fmt.Errorf("tmux: send-keys broke")
		}
		return []byte("ok"), nil
	})
	seedResponseState(t, h, "exec-1", "node-A", "bmad-node-A-100:0.0", "keep-me")

	err := h.executor.RespondToQuestionLegacy("exec-1", "node-A", "hi")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "send-keys")
}

// A send-keys Enter subprocess failure is surfaced to the caller.
func TestStory2_RespondToQuestion_SendKeysEnterFails(t *testing.T) {
	h := newHarness(t)
	var calls int
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "list-panes" {
			return []byte("0\n"), nil
		}
		if len(args) > 0 && args[0] == "send-keys" {
			calls++
			// First send-keys (-l literal) succeeds; second (Enter) fails.
			if calls == 1 {
				return []byte("ok"), nil
			}
			return nil, fmt.Errorf("tmux: enter failed")
		}
		return []byte("ok"), nil
	})
	seedResponseState(t, h, "exec-1", "node-A", "bmad-node-A-100:0.0", "keep-me")

	err := h.executor.RespondToQuestionLegacy("exec-1", "node-A", "hi")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Enter")
}

// Empty answer is allowed: send-keys -l with an empty literal plus Enter.
func TestStory2_RespondToQuestion_EmptyAnswer(t *testing.T) {
	h := newHarness(t)
	runner, calls, mu := responseRunner("0")
	h.executor.SetCommandRunner(runner)
	seedResponseState(t, h, "exec-1", "node-A", "bmad-node-A-100:0.0", "h")

	require.NoError(t, h.executor.RespondToQuestionLegacy("exec-1", "node-A", ""))

	mu.Lock()
	defer mu.Unlock()
	// We expect both a literal call (with empty string) and an Enter call.
	require.NotNil(t, findCall(*calls, "tmux", "send-keys", "-l", "-t", "bmad-node-A-100:0.0", ""))
	require.NotNil(t, findCall(*calls, "tmux", "send-keys", "-t", "bmad-node-A-100:0.0", "Enter"))
}
