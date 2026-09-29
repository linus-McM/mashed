package bmad

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

// ── Sequential Execution ──

// ── Node Failure Pauses Execution ──

// ── Cycle Detection ──

// ── Parallel Execution ──

// ── Pause/Resume ──

// ── Stop ──

// ── Error cases ──

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

// ── GetExecution returns copy ──

// ── GetCurrentExecution ────────────────────────────────────────────────
//
// GetCurrentExecution is the restore-on-mount hook used by the frontend
// when the WorkflowBuilder loads. It scans the in-memory executions map
// for the most-recently-started NON-TERMINAL execution (Running or
// Paused) whose RepoPath matches the caller. Completed/failed executions
// are ignored because there is nothing live to restore from them.

// ── Dynamic Executor — Condition Branching ──

// ── extractRegex Tests ──

// ── extractLines Tests ──

// ── Transform Node Integration Tests ──

// ── buildContextStringV3 Tests ──

// ── Edge-Based Context Passing ──

// ── Loop / LoopUntil Execution ──

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

// ── Story 2: Backend Response Injection ──
//
// These tests exercise (*Executor).RespondToQuestion in isolation by seeding
// an execState directly in the executor map. They bypass StartWorkflow so the
// test does not race against the dynamic runner goroutine — we control the
// mock CommandRunner deterministically.
