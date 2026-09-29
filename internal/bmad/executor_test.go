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
			// Use autonomous processes (Mode == "") so the test exercises the
			// legacy executeProcessNode dispatch.
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", Position: Position{X: 250, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "C", ProcessID: autonomousProcessFixtureID, Label: "C", Position: Position{X: 500, Y: 0}, Status: NodePending, Config: map[string]string{}},
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

// ── Node Failure Pauses Execution ──

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

// ── Cycle Detection ──

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

// ── Pause/Resume ──

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

// ── Stop ──

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

// ── Error cases ──

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
				ID: "n1", ProcessID: autonomousProcessFixtureID, Label: "A",
				Position: Position{X: 0, Y: 0}, Status: NodePending,
				Config: map[string]string{}, StoryID: "1-2-dashboard",
			},
			wantJSON: `"storyId":"1-2-dashboard"`,
		},
		{
			name: "without storyId (omitempty)",
			node: WorkflowNode{
				ID: "n2", ProcessID: autonomousProcessFixtureID, Label: "B",
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
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}, StoryID: "1-2-dashboard"},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-story", repoDir, "sonnet")
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

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, "/tmp/repo", "sonnet")
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
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}, StoryID: "1-2-dashboard"},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-story-fail", repoDir, "sonnet")
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
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-capture", "/tmp/repo", "sonnet")
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
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-bigcap", "/tmp/repo", "sonnet")
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
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-capfail", "/tmp/repo", "sonnet")
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

	// Return different output per node based on the session-name label
	// parsed from the capture-pane target.
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "capture-pane" {
			if label, ok := sessionLabelFromArgs(args); ok {
				switch label {
				case "a":
					return []byte("output-A"), nil
				case "b":
					return []byte("output-B"), nil
				}
			}
			return []byte("unknown"), nil
		}
		return successRunner()(ctx, name, args...)
	})

	wf := WorkflowDef{
		ID:   "wf-parcap",
		Name: "Parallel Capture",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B", Position: Position{X: 250, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{}, // No edges = parallel.
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-parcap", "/tmp/repo", "sonnet")
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

// runExecuteNodeSessionCase drives a single-node workflow through executeNode
// with a custom mock for `git rev-parse --abbrev-ref HEAD`, capturing the
// tmux session name that was created. Shared by the AC-7 / AC-8 tests.
func runExecuteNodeSessionCase(t *testing.T, gitBranchFn func() ([]byte, error)) (capturedSession string, execStatus WorkflowNodeStatus) {
	t.Helper()
	h := newHarness(t)

	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "git" && len(args) >= 4 && args[2] == "rev-parse" && args[3] == "--abbrev-ref" {
			return gitBranchFn()
		}
		if name == "tmux" && len(args) > 0 && args[0] == "new-session" {
			for i, a := range args {
				if a == "-s" && i+1 < len(args) && capturedSession == "" {
					capturedSession = args[i+1]
				}
			}
			return []byte("ok"), nil
		}
		if name == "tmux" && len(args) > 0 && args[0] == "list-panes" {
			return []byte("1\n"), nil
		}
		return []byte("ok"), nil
	})

	wf := WorkflowDef{
		ID:   "wf-execnode",
		Name: "ExecuteNode Session Case",
		Nodes: []WorkflowNode{
			{ID: "N1", ProcessID: autonomousProcessFixtureID, Label: "Draft PRD", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-10T00:00:00Z",
		UpdatedAt: "2026-04-10T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-execnode", "/tmp/testrepo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	require.NotEmpty(t, capturedSession, "new-session -s argument should be captured")
	return capturedSession, ex.Nodes[0].Status
}

// TestExecuteNode_AC7_UsesDescriptiveName verifies that executeNode builds
// its tmux session name via BuildSessionName and that the resulting name
// round-trips through ParseSessionName with the expected components.
func TestExecuteNode_AC7_UsesDescriptiveName(t *testing.T) {
	capturedSession, nodeStatus := runExecuteNodeSessionCase(t, func() ([]byte, error) {
		return []byte("main\n"), nil
	})

	repo, branch, label, shortHash, ok := ParseSessionName(capturedSession)
	require.True(t, ok, "captured session %q should parse as a BMAD session name", capturedSession)
	assert.Equal(t, "testrepo", repo, "repo component should be the basename of the repoPath")
	assert.Equal(t, "main", branch, "branch component should reflect git rev-parse output")
	assert.Equal(t, "draft-prd", label, "label component should be the slugified node label")
	assert.Len(t, shortHash, 8, "short hash should be 8 hex characters")
	assert.Equal(t, NodeComplete, nodeStatus)
}

// TestExecuteNode_AC8_BranchLookupFailureFallsBackToDetached verifies that
// every way `git rev-parse --abbrev-ref HEAD` can signal "no branch" — an
// error, empty stdout, or the literal "HEAD" string printed by a detached
// HEAD — falls back to DetachedBranch without failing the node.
func TestExecuteNode_AC8_BranchLookupFailureFallsBackToDetached(t *testing.T) {
	cases := []struct {
		name     string
		gitReply func() ([]byte, error)
	}{
		{
			name:     "git_error",
			gitReply: func() ([]byte, error) { return nil, fmt.Errorf("git rev-parse failed: not a repo") },
		},
		{
			name:     "empty_stdout",
			gitReply: func() ([]byte, error) { return []byte("\n"), nil },
		},
		{
			name:     "literal_HEAD_detached",
			gitReply: func() ([]byte, error) { return []byte("HEAD\n"), nil },
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			capturedSession, nodeStatus := runExecuteNodeSessionCase(t, tc.gitReply)

			_, branch, label, _, ok := ParseSessionName(capturedSession)
			require.True(t, ok, "captured session %q should still parse as a BMAD session name", capturedSession)
			assert.Equal(t, DetachedBranch, branch, "branch should fall back to DetachedBranch")
			assert.Equal(t, "draft-prd", label, "label should still be the slugified node label")
			assert.Equal(t, NodeComplete, nodeStatus, "node should complete despite missing branch")
		})
	}
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

// ── GetExecution returns copy ──

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

// ── GetCurrentExecution ────────────────────────────────────────────────
//
// GetCurrentExecution is the restore-on-mount hook used by the frontend
// when the WorkflowBuilder loads. It scans the in-memory executions map
// for the most-recently-started NON-TERMINAL execution (Running or
// Paused) whose RepoPath matches the caller. Completed/failed executions
// are ignored because there is nothing live to restore from them.

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

// ── Dynamic Executor — Condition Branching ──

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
		"process-a": "build version: 3.4.5 deployed",
	}))

	wf := WorkflowDef{
		ID:   "wf-transform-regex",
		Name: "Transform Regex",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "T", NodeType: NodeTypeTransform, Label: "Extract Version", Config: map[string]string{
				"sourceNode":     "A",
				"extractType":    "regex",
				"extractPattern": `version: (\S+)`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Process B", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-transform-regex", "/tmp/repo", "sonnet")
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
		"process-a": "header\nline2\nline3\nline4\nfooter",
	}))

	wf := WorkflowDef{
		ID:   "wf-transform-lines",
		Name: "Transform Lines",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-transform-lines", "/tmp/repo", "sonnet")
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

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-transform-nosrc", "/tmp/repo", "sonnet")
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
		"process-a": "raw output data",
	}))

	wf := WorkflowDef{
		ID:   "wf-transform-pass",
		Name: "Transform Passthrough",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", NodeType: NodeTypeProcess,
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

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-transform-pass", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, "raw output data", ex.NodeOutputs["T"], "unknown extractType should passthrough")
}

// ── buildContextStringV3 Tests ──

func TestBuildContextStringV3_IncludesTransformData(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", Status: NodeComplete, Config: map[string]string{}},
		{ID: "T", NodeType: NodeTypeTransform, Label: "Version Extract", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "3.4.5",
	}

	proc, _ := ProcessByID("bmad-create-prd")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", nil, "B")
	assert.Contains(t, result, "Version Extract")
	assert.Contains(t, result, "3.4.5")
}

func TestBuildContextStringV3_TruncatesLongData(t *testing.T) {
	longData := strings.Repeat("X", 3000)
	nodes := []WorkflowNode{
		{ID: "T", NodeType: NodeTypeTransform, Label: "Big Transform", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": longData,
	}

	proc, _ := ProcessByID("bmad-domain-research")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", nil, "B")
	assert.Contains(t, result, "Big Transform")
	// The data portion should be capped at 2000 chars.
	assert.LessOrEqual(t, len(result), 2100, "result should not contain full 3000-char data")
	assert.NotContains(t, result, longData, "full long data should be truncated")
}

func TestBuildContextStringV3_IncludesArtifactMatching(t *testing.T) {
	// bmad-create-prd has Inputs: ["product-brief"] and bmad-product-brief has Outputs: ["product-brief"].
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-product-brief", Label: "Product Brief", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Create PRD", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-create-prd")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", nil, "B")
	assert.Contains(t, result, "upstream process")
	assert.Contains(t, result, "product-brief")
}

func TestBuildContextStringV3_EmptyTransformData(t *testing.T) {
	// Transform node with empty output should be excluded.
	nodes := []WorkflowNode{
		{ID: "T", NodeType: NodeTypeTransform, Label: "Empty Transform", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "",
	}

	proc, _ := ProcessByID("bmad-domain-research")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", nil, "B")
	assert.NotContains(t, result, "Empty Transform", "empty transform data should not appear")
}

func TestBuildContextStringV3_SkipsNonCompleteTransforms(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "T", NodeType: NodeTypeTransform, Label: "Pending Transform", Status: NodePending, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "some data",
	}

	proc, _ := ProcessByID("bmad-domain-research")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", nil, "B")
	assert.NotContains(t, result, "Pending Transform", "non-complete transform should not appear")
}

func TestBuildContextStringV3_FilePathResolution(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  bool   // whether to create the artifact file on disk
		procID      string // downstream process
		upstreamID  string // upstream process
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:       "file_exists",
			setupFiles: true,
			procID:     "bmad-create-architecture", // Inputs: ["PRD.md"] — concrete contract, not autonomous-fixture
			upstreamID: "bmad-create-prd",          // Outputs: ["PRD.md"]
			wantContain: []string{
				"Read the artifact 'PRD.md' from file",
				"planning-artifacts/PRD.md",
			},
			wantAbsent: []string{
				"was not found",
			},
		},
		{
			name:       "file_missing",
			setupFiles: false,
			procID:     "bmad-create-architecture", // Inputs: ["PRD.md"] — concrete contract
			upstreamID: "bmad-create-prd",          // Outputs: ["PRD.md"]
			wantContain: []string{
				"should have produced 'PRD.md'",
				"but it was not found",
			},
			wantAbsent: []string{
				"Read the artifact",
			},
		},
		{
			name:       "unmapped_artifact",
			setupFiles: false,
			procID:     "bmad-domain-research", // Inputs: [] (empty)
			upstreamID: "bmad-domain-research", // Outputs: ["brainstorm-notes"]
			// brainstorming has no inputs, so no artifact matching at all
			wantContain: []string{},
			wantAbsent:  []string{"upstream process"},
		},
		{
			name:       "no_inputs",
			setupFiles: false,
			procID:     "bmad-domain-research", // Inputs: []
			upstreamID: "",
			wantContain: []string{},
			wantAbsent:  []string{"upstream", "artifact"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.setupFiles {
				// Create the artifact file.
				proc, ok := ProcessByID(tt.upstreamID)
				require.True(t, ok)
				for _, output := range proc.Outputs {
					p := ResolveArtifactPath(output, tmpDir)
					if p == "" {
						continue
					}
					require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
					require.NoError(t, os.WriteFile(p, []byte("test content"), 0o644))
				}
			}

			var nodes []WorkflowNode
			if tt.upstreamID != "" {
				nodes = append(nodes, WorkflowNode{
					ID:        "upstream",
					ProcessID: tt.upstreamID,
					Label:     "Upstream",
					Status:    NodeComplete,
					Config:    map[string]string{},
				})
			}
			nodes = append(nodes, WorkflowNode{
				ID:        "downstream",
				ProcessID: tt.procID,
				Label:     "Downstream",
				Status:    NodePending,
				Config:    map[string]string{},
			})

			nodeIndex := buildNodeIndex(nodes)
			nodeOutputs := map[string]string{}

			proc, ok := ProcessByID(tt.procID)
			require.True(t, ok)

			result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, tmpDir, nil, "downstream")

			for _, want := range tt.wantContain {
				assert.Contains(t, result, want, "expected result to contain %q", want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, result, absent, "expected result NOT to contain %q", absent)
			}
		})
	}
}

func TestBuildContextStringV3_TransformDataPreservedWithRepoPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the product-brief artifact file so we get file-path message.
	prdPath := ResolveArtifactPath("product-brief", tmpDir)
	require.NoError(t, os.MkdirAll(filepath.Dir(prdPath), 0o755))
	require.NoError(t, os.WriteFile(prdPath, []byte("brief content"), 0o644))

	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-product-brief", Label: "Product Brief", Status: NodeComplete, Config: map[string]string{}},
		{ID: "T", NodeType: NodeTypeTransform, Label: "Version Extract", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Create PRD", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "3.4.5",
	}

	proc, _ := ProcessByID("bmad-create-prd")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, tmpDir, nil, "B")

	// Should contain both file-path info AND transform data.
	assert.Contains(t, result, "Read the artifact 'product-brief' from file")
	assert.Contains(t, result, "Version Extract")
	assert.Contains(t, result, "3.4.5")
}

// ── Edge-Based Context Passing ──

func TestBuildContextStringV3_EdgeBasedContext_EmptyInputs(t *testing.T) {
	// Code Review has Inputs: [] but is edge-connected to Dev Story (Outputs: ["code", "tests"]).
	// Edge-based context should pass upstream outputs even without artifact name matching.
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-dev-story", Label: "Develop Story", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-code-review", Label: "Code Review", Status: NodePending, Config: map[string]string{}},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-code-review")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", edges, "B")
	assert.Contains(t, result, "Develop Story")
	assert.Contains(t, result, "code")
	assert.Contains(t, result, "tests")
}

func TestBuildContextStringV3_EdgeBasedContext_WithFileResolution(t *testing.T) {
	// bmad-domain-research outputs "domain-research"; edge-based context
	// mentions both the upstream label and the artifact name.
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-domain-research", Label: "Domain Research", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-product-brief", Label: "Product Brief", Status: NodePending, Config: map[string]string{}},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-product-brief")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", edges, "B")
	assert.Contains(t, result, "Domain Research")
	assert.Contains(t, result, "domain-research")
}

func TestBuildContextStringV3_EdgeBasedContext_NoDuplicates(t *testing.T) {
	// When artifact name matching already covers an output, edge-based context
	// should NOT duplicate it.
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-product-brief", Label: "Product Brief", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Create PRD", Status: NodePending, Config: map[string]string{}},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-create-prd") // Inputs: ["product-brief"]
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", edges, "B")
	// The artifact should be mentioned by exactly one context message, not duplicated
	// by both artifact matching and edge-based context. Count the message prefix
	// (not the raw substring, which also appears in file paths).
	assert.Equal(t, 1, strings.Count(result, "'product-brief'"), "product-brief should be referenced in exactly one context message")
}

func TestBuildContextStringV3_EdgeBasedContext_NonConnectedNodeIgnored(t *testing.T) {
	// Node C is complete but NOT edge-connected to B. Its outputs should NOT
	// appear in edge-based context (only artifact name matching can pick them up).
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-brainstorming", Label: "Brainstorming", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-code-review", Label: "Code Review", Status: NodePending, Config: map[string]string{}},
	}
	// No edge connecting A to B.
	edges := []WorkflowEdge{}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-code-review")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", edges, "B")
	assert.Empty(t, result, "no context without edges or matching inputs")
}

func TestBuildContextStringV3_EdgeBasedContext_SkipsNonCompleteUpstream(t *testing.T) {
	// Upstream node is connected by edge but not yet complete.
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-dev-story", Label: "Develop Story", Status: NodeRunning, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-code-review", Label: "Code Review", Status: NodePending, Config: map[string]string{}},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-code-review")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", edges, "B")
	assert.Empty(t, result, "running upstream should not produce context")
}

// ── Loop / LoopUntil Execution ──

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

// ── Artifact Event Emission ──

func TestCompleteNode_EmitsArtifactEvent_ProcessNode(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())

	// Use bmad-sprint-status (autonomous, outputs "sprint-status.yaml" →
	// implementation-artifacts/sprint-status.yaml). bmad-domain-research no
	// longer fits this test fixture — it is iterative post-rollout-03 and the
	// workflow would block on user input.
	repoDir := t.TempDir()
	artifactDir := filepath.Join(repoDir, "_bmad-output", "implementation-artifacts")
	require.NoError(t, os.MkdirAll(artifactDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(artifactDir, "sprint-status.yaml"), []byte("notes"), 0o644))

	// Save a single-node workflow with a process node.
	wf := WorkflowDef{
		ID:   "wf-artifact-test",
		Name: "Artifact Test",
		Nodes: []WorkflowNode{
			{ID: "N1", ProcessID: autonomousProcessFixtureID, Label: "Sprint Status", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-08T00:00:00Z",
		UpdatedAt: "2026-04-08T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), wf.ID, repoDir, "sonnet")
	require.NoError(t, err)

	// Wait for completion.
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	// Verify artifact event was emitted.
	artifactEvents := h.eventsByName("bmad:node:artifacts")
	require.Len(t, artifactEvents, 1, "should emit exactly one artifact event")

	ae, ok := artifactEvents[0].data.(NodeArtifactEvent)
	require.True(t, ok, "event data should be NodeArtifactEvent")
	assert.Equal(t, exec.ID, ae.ExecID)
	assert.Equal(t, "N1", ae.NodeID)
	assert.Equal(t, []string{"sprint-status.yaml"}, ae.Found)
	assert.Empty(t, ae.Missing, "artifact was created, so nothing should be missing")
}

func TestCompleteNode_ArtifactEvent_MissingArtifact(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())

	// Repo dir WITHOUT the expected artifact file.
	repoDir := t.TempDir()

	wf := WorkflowDef{
		ID:   "wf-artifact-missing",
		Name: "Artifact Missing Test",
		Nodes: []WorkflowNode{
			{ID: "N1", ProcessID: autonomousProcessFixtureID, Label: "PRD", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-08T00:00:00Z",
		UpdatedAt: "2026-04-08T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), wf.ID, repoDir, "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	artifactEvents := h.eventsByName("bmad:node:artifacts")
	require.Len(t, artifactEvents, 1)

	ae, ok := artifactEvents[0].data.(NodeArtifactEvent)
	require.True(t, ok)
	assert.Equal(t, exec.ID, ae.ExecID)
	assert.Equal(t, "N1", ae.NodeID)
	assert.Empty(t, ae.Found, "no artifact files exist on disk")
	assert.Equal(t, []string{"sprint-status.yaml"}, ae.Missing)
}

func TestCompleteNode_NoArtifactEvent_ControlNode(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())

	// A workflow with one condition node that evaluates to "true" and one merge.
	// Neither should produce artifact events.
	wf := WorkflowDef{
		ID:   "wf-control-no-artifact",
		Name: "Control No Artifact",
		Nodes: []WorkflowNode{
			{ID: "P1", ProcessID: autonomousProcessFixtureID, Label: "Brainstorm", Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}, NodeType: NodeTypeProcess},
			{ID: "M1", ProcessID: "", Label: "Merge", Position: Position{X: 250, Y: 0}, Status: NodePending, Config: map[string]string{}, NodeType: NodeTypeMerge},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "P1", Target: "M1"},
		},
		CreatedAt: "2026-04-08T00:00:00Z",
		UpdatedAt: "2026-04-08T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	repoDir := t.TempDir()
	// Create the artifact for P1 so it passes artifact check.
	artifactDir := filepath.Join(repoDir, "_bmad-output", "analysis-artifacts")
	require.NoError(t, os.MkdirAll(artifactDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(artifactDir, "brainstorm-notes.md"), []byte("notes"), 0o644))

	exec, err := h.executor.StartWorkflow(context.Background(), wf.ID, repoDir, "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	// Only process node P1 should have emitted an artifact event, NOT the merge node M1.
	artifactEvents := h.eventsByName("bmad:node:artifacts")
	require.Len(t, artifactEvents, 1, "only process nodes emit artifact events")

	ae, ok := artifactEvents[0].data.(NodeArtifactEvent)
	require.True(t, ok)
	assert.Equal(t, "P1", ae.NodeID, "artifact event should be from process node P1 only")
}

// ── GetArtifactStatus ──

func TestGetArtifactStatus_Exists(t *testing.T) {
	repoDir := t.TempDir()
	artifactDir := filepath.Join(repoDir, "_bmad-output", "planning-artifacts")
	require.NoError(t, os.MkdirAll(artifactDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(artifactDir, "PRD.md"), []byte("# PRD"), 0o644))

	exists, fullPath := GetArtifactStatus(repoDir, "PRD.md")
	assert.True(t, exists)
	assert.Equal(t, filepath.Join(repoDir, "_bmad-output", "planning-artifacts", "PRD.md"), fullPath)
}

func TestGetArtifactStatus_Missing(t *testing.T) {
	repoDir := t.TempDir()

	exists, fullPath := GetArtifactStatus(repoDir, "PRD.md")
	assert.False(t, exists)
	assert.Equal(t, filepath.Join(repoDir, "_bmad-output", "planning-artifacts", "PRD.md"), fullPath)
}

func TestGetArtifactStatus_UnmappedArtifact(t *testing.T) {
	repoDir := t.TempDir()

	exists, fullPath := GetArtifactStatus(repoDir, "code")
	assert.False(t, exists)
	assert.Empty(t, fullPath, "unmapped artifact should return empty path")
}

// ── Loop Items Tests ──

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

// ── Story 2: Backend Response Injection ──
//
// These tests exercise (*Executor).RespondToQuestion in isolation by seeding
// an execState directly in the executor map. They bypass StartWorkflow so the
// test does not race against the dynamic runner goroutine — we control the
// mock CommandRunner deterministically.

// cmdCall captures a single runCmd invocation for later assertions.
type cmdCall struct {
	name string
	args []string
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
