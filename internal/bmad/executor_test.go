package bmad

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Helpers ──

type eventRecord struct {
	name string
	data interface{}
}

type testHarness struct {
	storage  *Storage
	executor *Executor
	events   []eventRecord
	mu       sync.Mutex
}

func newHarness(t *testing.T) *testHarness {
	t.Helper()
	s, err := NewStorage(t.TempDir())
	require.NoError(t, err)

	h := &testHarness{storage: s}
	h.executor = NewExecutor(s, func(name string, data interface{}) {
		h.mu.Lock()
		h.events = append(h.events, eventRecord{name: name, data: data})
		h.mu.Unlock()
	})
	h.executor.pollInterval = 50 * time.Millisecond // fast polling for tests
	return h
}

func (h *testHarness) getEvents() []eventRecord {
	h.mu.Lock()
	defer h.mu.Unlock()
	cp := make([]eventRecord, len(h.events))
	copy(cp, h.events)
	return cp
}

func (h *testHarness) eventsByName(name string) []eventRecord {
	var out []eventRecord
	for _, e := range h.getEvents() {
		if e.name == name {
			out = append(out, e)
		}
	}
	return out
}

// successRunner simulates immediate success: tmux new-session succeeds,
// then list-panes returns "1" (pane_dead) on first poll.
func successRunner() CommandRunner {
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "list-panes" {
			return []byte("1\n"), nil
		}
		return []byte("ok"), nil
	}
}

// failRunner simulates tmux new-session failure.
func failRunner() CommandRunner {
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "new-session" {
			return nil, fmt.Errorf("tmux failed")
		}
		return []byte("1\n"), nil
	}
}

// delayRunner simulates a node that runs for a given duration.
func delayRunner(d time.Duration) CommandRunner {
	started := make(map[string]time.Time)
	var mu sync.Mutex
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "new-session" {
			// Extract session name.
			for i, a := range args {
				if a == "-s" && i+1 < len(args) {
					mu.Lock()
					started[args[i+1]] = time.Now()
					mu.Unlock()
				}
			}
			return []byte("ok"), nil
		}
		if len(args) > 0 && args[0] == "list-panes" {
			// Extract target.
			target := ""
			for i, a := range args {
				if a == "-t" && i+1 < len(args) {
					target = args[i+1]
				}
			}
			// Strip ":0.0" to get session name.
			session := target
			if idx := len(target) - 4; idx > 0 && target[idx:] == ":0.0" {
				session = target[:idx]
			}
			mu.Lock()
			st, ok := started[session]
			mu.Unlock()
			if ok && time.Since(st) >= d {
				return []byte("1\n"), nil
			}
			return []byte("0\n"), nil
		}
		return nil, nil
	}
}

func saveThreeNodeWorkflow(t *testing.T, s *Storage) string {
	t.Helper()
	wf := WorkflowDef{
		ID:   "wf-3seq",
		Name: "Three Sequential",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", ProcessID: "bmad-create-prd", Label: "B", Position: Position{X: 250, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "C", ProcessID: "bmad-validate-prd", Label: "C", Position: Position{X: 500, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "C"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, s.SaveWorkflow(wf))
	return wf.ID
}

// ── Topological Sort Tests ──

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

// ── Sequential Execution ──

func TestStartWorkflow_Sequential(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())
	wfID := saveThreeNodeWorkflow(t, h.storage)

	exec, err := h.executor.StartWorkflow(wfID, "/tmp/repo", "sonnet")
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

// ── Node Failure Pauses Execution ──

func TestStartWorkflow_NodeFailure(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(failRunner())
	wfID := saveThreeNodeWorkflow(t, h.storage)

	exec, err := h.executor.StartWorkflow(wfID, "/tmp/repo", "sonnet")
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

// ── Cycle Detection ──

func TestStartWorkflow_CyclicWorkflow(t *testing.T) {
	h := newHarness(t)
	wf := WorkflowDef{
		ID:   "wf-cycle",
		Name: "Cyclic",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", Config: map[string]string{}},
			{ID: "B", ProcessID: "bmad-create-prd", Label: "B", Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "A"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	_, err := h.executor.StartWorkflow("wf-cycle", "/tmp", "sonnet")
	assert.True(t, errors.Is(err, ErrCyclicWorkflow))
}

// ── Parallel Execution ──

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
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", Config: map[string]string{}},
			{ID: "B", ProcessID: "bmad-domain-research", Label: "B", Config: map[string]string{}},
			{ID: "C", ProcessID: "bmad-create-prd", Label: "C", Config: map[string]string{}},
			{ID: "D", ProcessID: "bmad-create-architecture", Label: "D", Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "C"},
			{ID: "e2", Source: "B", Target: "D"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-parallel", "/tmp", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	// All 4 nodes should have been spawned.
	assert.GreaterOrEqual(t, int(atomic.LoadInt32(&spawned)), 4)
}

// ── Pause/Resume ──

func TestPauseAndResume(t *testing.T) {
	h := newHarness(t)

	// Use a delay runner so we have time to pause mid-execution.
	h.executor.SetCommandRunner(delayRunner(200 * time.Millisecond))

	wfID := saveThreeNodeWorkflow(t, h.storage)
	exec, err := h.executor.StartWorkflow(wfID, "/tmp/repo", "sonnet")
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

// ── Stop ──

func TestStopWorkflow(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(delayRunner(10 * time.Second)) // long delay

	wfID := saveThreeNodeWorkflow(t, h.storage)
	exec, err := h.executor.StartWorkflow(wfID, "/tmp/repo", "sonnet")
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

// ── Error cases ──

func TestPauseWorkflow_NotRunning(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())
	wfID := saveThreeNodeWorkflow(t, h.storage)

	exec, err := h.executor.StartWorkflow(wfID, "/tmp", "sonnet")
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

	exec, err := h.executor.StartWorkflow(wfID, "/tmp", "sonnet")
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
	_, err := h.executor.StartWorkflow("nonexistent", "/tmp", "sonnet")
	assert.True(t, errors.Is(err, ErrWorkflowNotFound))
}

// ── GetExecution returns copy ──

func TestGetExecution_ReturnsCopy(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())
	wfID := saveThreeNodeWorkflow(t, h.storage)

	exec, err := h.executor.StartWorkflow(wfID, "/tmp", "sonnet")
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
