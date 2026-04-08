package bmad

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// ── StoryID JSON serialization ──

func TestWorkflowNode_StoryID_Serialization(t *testing.T) {
	tests := []struct {
		name     string
		node     WorkflowNode
		wantJSON string // substring to check in JSON
		noJSON   string // substring that must NOT appear
	}{
		{
			name: "with storyId",
			node: WorkflowNode{
				ID: "n1", ProcessID: "bmad-brainstorming", Label: "A",
				Position: Position{X: 0, Y: 0}, Status: NodePending,
				Config: map[string]string{}, StoryID: "1-2-dashboard",
			},
			wantJSON: `"storyId":"1-2-dashboard"`,
		},
		{
			name: "without storyId (omitempty)",
			node: WorkflowNode{
				ID: "n2", ProcessID: "bmad-brainstorming", Label: "B",
				Position: Position{X: 0, Y: 0}, Status: NodePending,
				Config: map[string]string{},
			},
			noJSON: `"storyId"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.node)
			require.NoError(t, err)
			jsonStr := string(data)

			if tt.wantJSON != "" {
				assert.Contains(t, jsonStr, tt.wantJSON)
			}
			if tt.noJSON != "" {
				assert.NotContains(t, jsonStr, tt.noJSON)
			}

			// Round-trip
			var decoded WorkflowNode
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)
			assert.Equal(t, tt.node.StoryID, decoded.StoryID)
		})
	}
}

func TestWorkflowNode_StoryID_BackwardCompat(t *testing.T) {
	// JSON without storyId field should deserialize without error.
	jsonStr := `{"id":"n1","processId":"p1","label":"A","position":{"x":0,"y":0},"status":"pending","config":{},"tmuxTarget":""}`
	var node WorkflowNode
	err := json.Unmarshal([]byte(jsonStr), &node)
	require.NoError(t, err)
	assert.Empty(t, node.StoryID)
}

// ── completeNode auto-advances story status ──

func TestCompleteNode_WithStoryID_AdvancesStory(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())

	// Create sprint YAML in a temp repo.
	repoDir := createSprintYAMLForExec(t)

	// Save a workflow with a node that has a StoryID.
	wf := WorkflowDef{
		ID:   "wf-story",
		Name: "Story Workflow",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}, StoryID: "1-2-dashboard"},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-story", repoDir, "sonnet")
	require.NoError(t, err)

	// Wait for completion.
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	// Verify sprint event was emitted.
	sprintEvents := h.eventsByName("bmad:sprint:updated")
	require.NotEmpty(t, sprintEvents, "should emit bmad:sprint:updated event")
	eventData, ok := sprintEvents[0].data.(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "1-2-dashboard", eventData["storyId"])
	assert.Equal(t, "in-progress", eventData["status"])

	// Verify the story status was actually updated in the YAML file.
	result, err := ParseSprintStatus(repoDir)
	require.NoError(t, err)
	require.Len(t, result.Epics, 1)
	// 1-2-dashboard was "backlog", should now be "in-progress"
	found := false
	for _, story := range result.Epics[0].Stories {
		if story.ID == "1-2-dashboard" {
			assert.Equal(t, StoryInProgress, story.Status)
			found = true
		}
	}
	assert.True(t, found, "story 1-2-dashboard should exist in sprint status")
}

func TestCompleteNode_WithoutStoryID_NoSprintEvent(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())
	wfID := saveThreeNodeWorkflow(t, h.storage)

	exec, err := h.executor.StartWorkflow(wfID, "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	// No sprint events should be emitted for nodes without storyID.
	sprintEvents := h.eventsByName("bmad:sprint:updated")
	assert.Empty(t, sprintEvents, "should NOT emit bmad:sprint:updated for nodes without storyID")
}

func TestFailNode_WithStoryID_NoSprintUpdate(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(failRunner())

	// Create sprint YAML.
	repoDir := createSprintYAMLForExec(t)

	wf := WorkflowDef{
		ID:   "wf-story-fail",
		Name: "Story Fail Workflow",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}, StoryID: "1-2-dashboard"},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-story-fail", repoDir, "sonnet")
	require.NoError(t, err)

	// Wait for pause (failure causes pause).
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecPaused
	}, 5*time.Second, 50*time.Millisecond)

	// Verify node failed.
	ex, _ := h.executor.GetExecution(exec.ID)
	assert.Equal(t, NodeFailed, ex.Nodes[0].Status)

	// No sprint events should be emitted on failure.
	sprintEvents := h.eventsByName("bmad:sprint:updated")
	assert.Empty(t, sprintEvents, "should NOT emit bmad:sprint:updated when node fails")

	// Verify story status unchanged in YAML.
	result, err := ParseSprintStatus(repoDir)
	require.NoError(t, err)
	for _, story := range result.Epics[0].Stories {
		if story.ID == "1-2-dashboard" {
			assert.Equal(t, StoryBacklog, story.Status, "story status should remain unchanged on failure")
		}
	}

	// Cleanup: stop the paused execution.
	h.executor.StopWorkflow(exec.ID)
}

// createSprintYAMLForExec creates a sprint-status.yaml in a temp dir for executor tests.
func createSprintYAMLForExec(t *testing.T) string {
	t.Helper()
	content := `generated: "2026-04-07T10:00:00Z"
last_updated: "2026-04-07T12:00:00Z"
project: "mashed"
project_key: "MSHD"
tracking_system: "github"
story_location: "docs/stories"
development_status:
  epic-1: in-progress
  1-1-user-auth: done
  1-2-dashboard: backlog
  1-3-settings: backlog
`
	repoDir := t.TempDir()
	dir := filepath.Join(repoDir, "_bmad-output", "implementation-artifacts")
	require.NoError(t, os.MkdirAll(dir, 0755))
	require.NoError(t, os.WriteFile(
		filepath.Join(dir, "sprint-status.yaml"),
		[]byte(content),
		0644,
	))
	return repoDir
}

// ── Output Capture ──

// captureRunner wraps successRunner with capture-pane handling.
// The outputs map is keyed by tmux target (e.g., "bmad-A-12345:0.0").
func captureRunner(outputs map[string]string) CommandRunner {
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			// Extract target from -t flag.
			target := ""
			for i, a := range args {
				if a == "-t" && i+1 < len(args) {
					target = args[i+1]
				}
			}
			if out, ok := outputs[target]; ok {
				return []byte(out), nil
			}
			return []byte("default output"), nil
		}
		// Delegate to successRunner for everything else.
		return successRunner()(ctx, name, args...)
	}
}

func TestCaptureOutput_StoresOnCompletion(t *testing.T) {
	h := newHarness(t)

	// We need a runner that returns capture-pane output keyed by any target.
	// Since we don't know the exact target name (it includes a timestamp),
	// use a runner that returns captured output for ANY capture-pane call.
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			return []byte("hello from tmux"), nil
		}
		return successRunner()(ctx, name, args...)
	})

	wf := WorkflowDef{
		ID:   "wf-capture",
		Name: "Capture Test",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-capture", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, "hello from tmux", ex.NodeOutputs["A"])
}

func TestCaptureOutput_100KBCap(t *testing.T) {
	h := newHarness(t)

	// Generate a string larger than 102400 bytes.
	bigOutput := strings.Repeat("X", 200000) // 200KB
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			return []byte(bigOutput), nil
		}
		return successRunner()(ctx, name, args...)
	})

	wf := WorkflowDef{
		ID:   "wf-bigcap",
		Name: "Big Capture",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-bigcap", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)

	captured := ex.NodeOutputs["A"]
	assert.LessOrEqual(t, len(captured), 102400, "output should be capped at 100KB")
	// Verify it kept the TAIL (all X's, so the tail is also X's — check length).
	assert.Equal(t, 102400, len(captured))
	// The tail of the original should match the captured output.
	assert.Equal(t, bigOutput[len(bigOutput)-102400:], captured)
}

func TestCaptureOutput_FailureNonFatal(t *testing.T) {
	h := newHarness(t)

	// capture-pane fails, but node should still complete.
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			return nil, fmt.Errorf("capture-pane failed")
		}
		return successRunner()(ctx, name, args...)
	})

	wf := WorkflowDef{
		ID:   "wf-capfail",
		Name: "Capture Fail",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-capfail", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, NodeComplete, ex.Nodes[0].Status, "node should complete even if capture fails")
	// NodeOutputs for A should be empty string (capture failed).
	assert.Empty(t, ex.NodeOutputs["A"])
}

func TestCaptureOutput_ParallelNodes(t *testing.T) {
	h := newHarness(t)

	// Return different output per node based on the target.
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			target := ""
			for i, a := range args {
				if a == "-t" && i+1 < len(args) {
					target = args[i+1]
				}
			}
			// Use target to differentiate outputs; target contains nodeID.
			if strings.Contains(target, "bmad-A-") {
				return []byte("output-A"), nil
			}
			if strings.Contains(target, "bmad-B-") {
				return []byte("output-B"), nil
			}
			return []byte("unknown"), nil
		}
		return successRunner()(ctx, name, args...)
	})

	wf := WorkflowDef{
		ID:   "wf-parcap",
		Name: "Parallel Capture",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", ProcessID: "bmad-domain-research", Label: "B", Position: Position{X: 250, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{}, // No edges = parallel.
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-parcap", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, "output-A", ex.NodeOutputs["A"])
	assert.Equal(t, "output-B", ex.NodeOutputs["B"])
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
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-copycap", "/tmp/repo", "sonnet")
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

// ── Dynamic Executor — Condition Branching ──

// conditionRunner returns a CommandRunner for condition/merge tests.
// It uses the outputs map to return capture-pane content keyed by node ID
// (matches target strings containing "bmad-{nodeID}-").
func conditionRunner(outputs map[string]string) CommandRunner {
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			target := ""
			for i, a := range args {
				if a == "-t" && i+1 < len(args) {
					target = args[i+1]
				}
			}
			for nodeID, out := range outputs {
				if strings.Contains(target, "bmad-"+nodeID+"-") {
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

func TestDynamicExecutor_ConditionBranching_TrueBranch(t *testing.T) {
	// Workflow: A(process) -> B(condition, contains "SUCCESS") -> C(true) and D(false).
	// A outputs "SUCCESS". Verify C completes, D is skipped.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"A": "operation SUCCESS complete",
	}))

	wf := WorkflowDef{
		ID:   "wf-cond-true",
		Name: "Condition True Branch",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", NodeType: NodeTypeCondition, Label: "Check", Config: map[string]string{
				"condition": `{"type":"contains","pattern":"SUCCESS","sourceNode":"A"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "C", ProcessID: "bmad-create-prd", Label: "True Path", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: -100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: "bmad-validate-prd", Label: "False Path", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow("wf-cond-true", "/tmp/repo", "sonnet")
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
		"A": "operation FAILURE complete",
	}))

	wf := WorkflowDef{
		ID:   "wf-cond-false",
		Name: "Condition False Branch",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", NodeType: NodeTypeCondition, Label: "Check", Config: map[string]string{
				"condition": `{"type":"contains","pattern":"SUCCESS","sourceNode":"A"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "C", ProcessID: "bmad-create-prd", Label: "True Path", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: -100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: "bmad-validate-prd", Label: "False Path", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow("wf-cond-false", "/tmp/repo", "sonnet")
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
		"upstream": "has SUCCESS in output",
	}))

	wf := WorkflowDef{
		ID:   "wf-merge",
		Name: "Merge After Condition",
		Nodes: []WorkflowNode{
			{ID: "upstream", ProcessID: "bmad-brainstorming", Label: "Upstream", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "A", NodeType: NodeTypeCondition, Label: "Condition A", Config: map[string]string{
				"condition": `{"type":"contains","pattern":"SUCCESS","sourceNode":"upstream"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: "bmad-create-prd", Label: "True Path", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: -100}, Status: NodePending, Config: map[string]string{}},
			{ID: "C", ProcessID: "bmad-validate-prd", Label: "False Path", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow("wf-merge", "/tmp/repo", "sonnet")
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
		"A": "nothing special here",
	}))

	wf := WorkflowDef{
		ID:   "wf-allskip",
		Name: "All Branches Skipped",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", NodeType: NodeTypeCondition, Label: "Check MAGIC", Config: map[string]string{
				"condition": `{"type":"contains","pattern":"MAGIC","sourceNode":"A"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "C", ProcessID: "bmad-create-prd", Label: "True Only", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow("wf-allskip", "/tmp/repo", "sonnet")
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
		"A": "no match",
	}))

	wf := WorkflowDef{
		ID:   "wf-skipevt",
		Name: "Skipped Events",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", NodeType: NodeTypeCondition, Label: "Check", Config: map[string]string{
				"condition": `{"type":"contains","pattern":"TRIGGER","sourceNode":"A"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "C", ProcessID: "bmad-create-prd", Label: "True", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: -100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: "bmad-validate-prd", Label: "False", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow("wf-skipevt", "/tmp/repo", "sonnet")
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

// ── extractRegex Tests ──

func TestExtractRegex_WithCaptureGroup(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pattern string
		want    string
	}{
		{
			name:    "version capture group",
			input:   "Released version: 1.2.3",
			pattern: `version: (\S+)`,
			want:    "1.2.3",
		},
		{
			name:    "parenthesized group",
			input:   "error code: (42)",
			pattern: `code: \((\d+)\)`,
			want:    "42",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRegex(tt.input, tt.pattern)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExtractRegex_WithoutCaptureGroup(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pattern string
		want    string
	}{
		{
			name:    "semver match",
			input:   "Version 2.5.1",
			pattern: `\d+\.\d+\.\d+`,
			want:    "2.5.1",
		},
		{
			name:    "word match",
			input:   "hello world",
			pattern: `\w+`,
			want:    "hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRegex(tt.input, tt.pattern)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExtractRegex_NoMatch(t *testing.T) {
	result := extractRegex("nothing here", "NOTFOUND")
	assert.Equal(t, "", result)
}

func TestExtractRegex_InvalidRegex(t *testing.T) {
	result := extractRegex("some text", "[invalid")
	assert.Equal(t, "", result)
}

// ── extractLines Tests ──

func TestExtractLines_Range(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5"
	result := extractLines(input, "2-4")
	assert.Equal(t, "line2\nline3\nline4", result)
}

func TestExtractLines_LastN(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5"
	result := extractLines(input, "-3")
	assert.Equal(t, "line3\nline4\nline5", result)
}

func TestExtractLines_Single(t *testing.T) {
	input := "line1\nline2\nline3"
	result := extractLines(input, "1")
	assert.Equal(t, "line1", result)
}

func TestExtractLines_OutOfRange(t *testing.T) {
	input := "line1\nline2"
	result := extractLines(input, "1-100")
	assert.Equal(t, "line1\nline2", result, "should clamp to available lines")
}

func TestExtractLines_InvalidPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{name: "non-numeric", pattern: "abc"},
		{name: "zero start", pattern: "0"},
		{name: "negative start", pattern: "-0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLines("line1\nline2", tt.pattern)
			assert.Equal(t, "", result)
		})
	}
}

func TestExtractLines_StartBeyondLength(t *testing.T) {
	input := "line1\nline2"
	result := extractLines(input, "10")
	assert.Equal(t, "", result, "start beyond line count should return empty")
}

// ── Transform Node Integration Tests ──

func TestTransformNode_RegexExtraction(t *testing.T) {
	// Workflow: A(process) -> T(transform, regex) -> B(process).
	// A produces "version: 3.4.5". T extracts "3.4.5". B should complete.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"A": "build version: 3.4.5 deployed",
	}))

	wf := WorkflowDef{
		ID:   "wf-transform-regex",
		Name: "Transform Regex",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "T", NodeType: NodeTypeTransform, Label: "Extract Version", Config: map[string]string{
				"sourceNode":     "A",
				"extractType":    "regex",
				"extractPattern": `version: (\S+)`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: "bmad-create-prd", Label: "Process B", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "T"},
			{ID: "e2", Source: "T", Target: "B"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-transform-regex", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)

	// Verify transform extracted the version.
	assert.Equal(t, "3.4.5", ex.NodeOutputs["T"])

	// All nodes should complete.
	for _, n := range ex.Nodes {
		assert.Equal(t, NodeComplete, n.Status, "node %s should be complete", n.ID)
	}
}

func TestTransformNode_LinesExtraction(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"A": "header\nline2\nline3\nline4\nfooter",
	}))

	wf := WorkflowDef{
		ID:   "wf-transform-lines",
		Name: "Transform Lines",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "T", NodeType: NodeTypeTransform, Label: "Extract Lines", Config: map[string]string{
				"sourceNode":     "A",
				"extractType":    "lines",
				"extractPattern": "2-4",
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "T"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-transform-lines", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, "line2\nline3\nline4", ex.NodeOutputs["T"])
}

func TestTransformNode_MissingSource(t *testing.T) {
	// Transform with nonexistent sourceNode -> empty output, completes.
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())

	wf := WorkflowDef{
		ID:   "wf-transform-nosrc",
		Name: "Transform Missing Source",
		Nodes: []WorkflowNode{
			{ID: "T", NodeType: NodeTypeTransform, Label: "Orphan Transform", Config: map[string]string{
				"sourceNode":     "nonexistent",
				"extractType":    "regex",
				"extractPattern": `(\d+)`,
			}, Position: Position{X: 0, Y: 0}, Status: NodePending},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-transform-nosrc", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)
	assert.Equal(t, NodeComplete, ex.Nodes[0].Status)
	assert.Equal(t, "", ex.NodeOutputs["T"], "missing source should produce empty output")
}

func TestTransformNode_Passthrough(t *testing.T) {
	// Unknown extractType should passthrough source output.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"A": "raw output data",
	}))

	wf := WorkflowDef{
		ID:   "wf-transform-pass",
		Name: "Transform Passthrough",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "T", NodeType: NodeTypeTransform, Label: "Passthrough", Config: map[string]string{
				"sourceNode":     "A",
				"extractType":    "unknown",
				"extractPattern": "",
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "T"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow("wf-transform-pass", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, "raw output data", ex.NodeOutputs["T"], "unknown extractType should passthrough")
}

// ── buildContextString Tests ──

func TestBuildContextStringV2_IncludesTransformData(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-brainstorming", Label: "Process A", Status: NodeComplete, Config: map[string]string{}},
		{ID: "T", NodeType: NodeTypeTransform, Label: "Version Extract", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-create-prd", Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "3.4.5",
	}

	proc, _ := ProcessByID("bmad-create-prd")
	result := buildContextString(proc, nodes, nodeIndex, nodeOutputs)
	assert.Contains(t, result, "Version Extract")
	assert.Contains(t, result, "3.4.5")
}

func TestBuildContextStringV2_TruncatesLongData(t *testing.T) {
	longData := strings.Repeat("X", 3000)
	nodes := []WorkflowNode{
		{ID: "T", NodeType: NodeTypeTransform, Label: "Big Transform", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-brainstorming", Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": longData,
	}

	proc, _ := ProcessByID("bmad-brainstorming")
	result := buildContextString(proc, nodes, nodeIndex, nodeOutputs)
	assert.Contains(t, result, "Big Transform")
	// The data portion should be capped at 2000 chars.
	assert.LessOrEqual(t, len(result), 2100, "result should not contain full 3000-char data")
	assert.NotContains(t, result, longData, "full long data should be truncated")
}

func TestBuildContextStringV2_IncludesArtifactMatching(t *testing.T) {
	// bmad-create-prd has Inputs: ["product-brief"] and bmad-product-brief has Outputs: ["product-brief"].
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-product-brief", Label: "Product Brief", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-create-prd", Label: "Create PRD", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-create-prd")
	result := buildContextString(proc, nodes, nodeIndex, nodeOutputs)
	assert.Contains(t, result, "upstream process")
	assert.Contains(t, result, "product-brief")
}

func TestBuildContextStringV2_EmptyTransformData(t *testing.T) {
	// Transform node with empty output should be excluded.
	nodes := []WorkflowNode{
		{ID: "T", NodeType: NodeTypeTransform, Label: "Empty Transform", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-brainstorming", Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "",
	}

	proc, _ := ProcessByID("bmad-brainstorming")
	result := buildContextString(proc, nodes, nodeIndex, nodeOutputs)
	assert.NotContains(t, result, "Empty Transform", "empty transform data should not appear")
}

func TestBuildContextStringV2_SkipsNonCompleteTransforms(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "T", NodeType: NodeTypeTransform, Label: "Pending Transform", Status: NodePending, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-brainstorming", Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "some data",
	}

	proc, _ := ProcessByID("bmad-brainstorming")
	result := buildContextString(proc, nodes, nodeIndex, nodeOutputs)
	assert.NotContains(t, result, "Pending Transform", "non-complete transform should not appear")
}

// ── Loop / LoopUntil Execution ──

// loopRunner creates a CommandRunner that returns different capture-pane output
// per iteration. It tracks how many times capture-pane is called for body nodes
// (targets containing "bmad-B-") to determine the current iteration.
// The iterOutputs map is 1-indexed: iterOutputs[1] is the output for the first
// execution of the body node.
func loopRunner(iterOutputs map[int]string) CommandRunner {
	var bodyCaptures int32
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			target := ""
			for i, a := range args {
				if a == "-t" && i+1 < len(args) {
					target = args[i+1]
				}
			}
			// Only count captures for body nodes (containing "bmad-B-").
			if strings.Contains(target, "bmad-B-") {
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
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "L", NodeType: NodeTypeLoop, Label: "Loop", Config: map[string]string{
				"maxIterations": "3",
				"loopBodyNodes": "B",
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: "bmad-brainstorming", Label: "B", NodeType: NodeTypeProcess,
				Position: Position{X: 400, Y: 100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: "bmad-create-prd", Label: "D", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow("wf-loop-fixed", "/tmp/repo", "sonnet")
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
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "LU", NodeType: NodeTypeLoopUntil, Label: "LoopUntil", Config: map[string]string{
				"maxIterations": "10",
				"loopBodyNodes": "B",
				"condition":     `{"type":"contains","pattern":"DONE","sourceNode":"B"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: "bmad-brainstorming", Label: "B", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow("wf-loopuntil-met", "/tmp/repo", "sonnet")
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
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "LU", NodeType: NodeTypeLoopUntil, Label: "LoopUntil", Config: map[string]string{
				"maxIterations": "3",
				"loopBodyNodes": "B",
				"condition":     `{"type":"contains","pattern":"NEVER_MATCH_THIS","sourceNode":"B"}`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: "bmad-brainstorming", Label: "B", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow("wf-loopuntil-max", "/tmp/repo", "sonnet")
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
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "L", NodeType: NodeTypeLoop, Label: "Loop", Config: map[string]string{
				"maxIterations": "5",
				"loopBodyNodes": "B",
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: "bmad-brainstorming", Label: "B", NodeType: NodeTypeProcess,
				Position: Position{X: 400, Y: 100}, Status: NodePending, Config: map[string]string{}},
			{ID: "D", ProcessID: "bmad-create-prd", Label: "D", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow("wf-loop-fail", "/tmp/repo", "sonnet")
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
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "L", NodeType: NodeTypeLoop, Label: "Loop", Config: map[string]string{
				"maxIterations": "5",
				"loopBodyNodes": "",
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "D", ProcessID: "bmad-create-prd", Label: "D", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow("wf-loop-empty", "/tmp/repo", "sonnet")
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
