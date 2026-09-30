// Story breadcrumbs-05 RED tests: executor must resolve OutputPaths
// on node complete, emit them via NodeArtifactEvent.Paths, and clear
// prior InputPaths/OutputPaths on StartWorkflow. Minimal type fields
// have been added to types.go so these tests COMPILE — behaviour is
// unimplemented so every test below fails at runtime.

package bmad

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AC-1: OutputPaths populated on node complete, event carries Paths.
//
// Uses bmad-sprint-status (autonomous; outputs ["sprint-status.yaml"] →
// implementation-artifacts/sprint-status.yaml) — bmad-create-prd is Guided
// post-rollout-07 and would block awaiting staged inputs.
func TestAC1_OutputPathsPopulated_OnNodeComplete(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())

	repoDir := t.TempDir()
	implDir := filepath.Join(repoDir, "_bmad-output", "implementation-artifacts")
	require.NoError(t, os.MkdirAll(implDir, 0o755))
	artifactPath := filepath.Join(implDir, "sprint-status.yaml")
	require.NoError(t, os.WriteFile(artifactPath, []byte("status: ready"), 0o644))

	wf := WorkflowDef{
		ID: "wf-ac1-outputpaths", Name: "AC1",
		Nodes: []WorkflowNode{{
			ID: "N1", ProcessID: autonomousProcessFixtureID, Label: "Sprint Status",
			Position: Position{X: 0, Y: 0}, Status: NodePending,
			Config: map[string]string{},
		}},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-14T00:00:00Z", UpdatedAt: "2026-04-14T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), wf.ID, repoDir, "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	require.Len(t, ex.Nodes, 1)

	require.NotNil(t, ex.Nodes[0].OutputPaths,
		"OutputPaths must be initialised on complete; got nil")
	got, ok := ex.Nodes[0].OutputPaths["sprint-status.yaml"]
	require.True(t, ok, "OutputPaths should contain sprint-status.yaml; got %v", ex.Nodes[0].OutputPaths)
	assert.Equal(t, artifactPath, got, "OutputPaths[sprint-status.yaml] must equal resolved abs path")

	events := h.eventsByName("bmad:node:artifacts")
	require.Len(t, events, 1)
	ae, ok := events[0].data.(NodeArtifactEvent)
	require.True(t, ok)
	require.NotNil(t, ae.Paths, "NodeArtifactEvent.Paths must be populated")
	assert.Equal(t, artifactPath, ae.Paths["sprint-status.yaml"], "event.Paths[sprint-status.yaml] mismatch")
}

// AC-2: Unmapped outputs (e.g. "code") excluded from OutputPaths.
//
// Verifies the skip behaviour by inspecting completeNode's OutputPaths
// population directly. After the rollout, every BMAD process with an
// unmapped output is interactive, so a full e2e StartWorkflow with an
// unmapped output would block on user input. The mapped-output round-trip
// is exercised by TestAC1 with bmad-create-prd.
func TestAC2_UnmappedArtifacts_SkippedFromOutputPaths(t *testing.T) {
	// Confirm "code" and "any-doc" are unmapped per ResolveArtifactPath
	// — these are the canonical unmapped artifact names. completeNode
	// must skip them when populating OutputPaths.
	tmp := t.TempDir()
	for _, unmapped := range []string{"code", "tests", "any-doc"} {
		assert.Empty(t, ResolveArtifactPath(unmapped, tmp),
			"AC-2: %q must remain unmapped (resolves to empty path)", unmapped)
	}
	// Confirm a known mapped artifact resolves to a non-empty path under
	// the repo root — completeNode populates OutputPaths only for these.
	mapped := ResolveArtifactPath("PRD.md", tmp)
	assert.NotEmpty(t, mapped, "AC-2: 'PRD.md' must resolve to a mapped path")
	assert.True(t, strings.HasPrefix(mapped, tmp),
		"AC-2: mapped path must be repo-relative")
}

// AC-3: Missing file skipped from OutputPaths; surfaces in Missing.
func TestAC3_MissingFile_SkippedFromOutputPaths(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())

	repoDir := t.TempDir() // no sprint-status.yaml on disk

	wf := WorkflowDef{
		ID: "wf-ac3-missing", Name: "AC3",
		Nodes: []WorkflowNode{{
			ID: "N1", ProcessID: autonomousProcessFixtureID, Label: "Sprint Status",
			Position: Position{X: 0, Y: 0}, Status: NodePending,
			Config: map[string]string{},
		}},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-14T00:00:00Z", UpdatedAt: "2026-04-14T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), wf.ID, repoDir, "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)

	require.NotNil(t, ex.Nodes[0].OutputPaths, "OutputPaths must be initialised (possibly empty)")
	_, hasArtifact := ex.Nodes[0].OutputPaths["sprint-status.yaml"]
	assert.False(t, hasArtifact, "missing file must be omitted from OutputPaths")

	events := h.eventsByName("bmad:node:artifacts")
	require.Len(t, events, 1)
	ae := events[0].data.(NodeArtifactEvent)
	assert.Contains(t, ae.Missing, "sprint-status.yaml", "event.Missing must surface sprint-status.yaml")
	require.NotNil(t, ae.Paths, "event.Paths must be initialised even when empty")
	_, evArtifact := ae.Paths["sprint-status.yaml"]
	assert.False(t, evArtifact, "event.Paths must not contain missing sprint-status.yaml")
}

// AC-4: StartWorkflow clears prior OutputPaths / InputPaths.
func TestAC4_StartWorkflow_ClearsPriorOutputPaths(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	})

	wf := WorkflowDef{
		ID: "wf-ac4-clear", Name: "AC4",
		Nodes: []WorkflowNode{
			{
				ID: "A", ProcessID: autonomousProcessFixtureID, Label: "A",
				Position: Position{X: 0, Y: 0}, Status: NodePending,
				Config:      map[string]string{},
				OutputPaths: map[string]string{"sprint-status.yaml": "/stale/A/out/sprint-status.yaml"},
				InputPaths:  map[string]string{"sprint-status.yaml": "/stale/A/in/sprint-status.yaml"},
			},
			{
				ID: "B", ProcessID: autonomousProcessFixtureID, Label: "B",
				Position: Position{X: 0, Y: 0}, Status: NodePending,
				Config:      map[string]string{},
				OutputPaths: map[string]string{"sprint-status.yaml": "/stale/B/out/sprint-status.yaml"},
				InputPaths:  map[string]string{"sprint-status.yaml": "/stale/B/in/sprint-status.yaml"},
			},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-14T00:00:00Z", UpdatedAt: "2026-04-14T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), wf.ID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	defer func() { _ = h.executor.StopWorkflow(exec.ID) }()

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	require.Len(t, ex.Nodes, 2)
	for _, n := range ex.Nodes {
		assert.Empty(t, n.OutputPaths, "node %s: OutputPaths must be cleared; got %v", n.ID, n.OutputPaths)
		assert.Empty(t, n.InputPaths, "node %s: InputPaths must be cleared; got %v", n.ID, n.InputPaths)
	}
}

// AC-5: Concurrent completeNode calls are race-free (run with -race).
func TestAC5_ConcurrentCompletion_NoRace(t *testing.T) {
	h := newHarness(t)

	// bmad-sprint-status outputs ["sprint-status.yaml"] — mapped. Race-safety
	// check doesn't depend on output count, just on concurrent completeNode
	// calls landing on different node indices.
	repoDir := t.TempDir()
	implDir := filepath.Join(repoDir, "_bmad-output", "implementation-artifacts")
	require.NoError(t, os.MkdirAll(implDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(implDir, "sprint-status.yaml"), []byte("x"), 0o644))

	state := &execState{
		exec: &WorkflowExecution{
			ID: "exec-race", Status: ExecRunning, RepoPath: repoDir,
			Nodes: []WorkflowNode{
				{ID: "A", ProcessID: autonomousProcessFixtureID, Status: NodeRunning, NodeType: NodeTypeProcess},
				{ID: "B", ProcessID: autonomousProcessFixtureID, Status: NodeRunning, NodeType: NodeTypeProcess},
			},
			NodeOutputs: map[string]string{},
		},
		cancel:           func() {},
		inDegree:         map[string]int{"A": 0, "B": 0},
		outEdges:         map[string][]WorkflowEdge{},
		lastQuestionHash: map[string]string{},
		lastOutputHash:   map[string]string{},
		idleEmitted:      map[string]bool{},
	}
	h.executor.mu.Lock()
	h.executor.executions["exec-race"] = state
	h.executor.mu.Unlock()

	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); <-start; h.executor.completeNode(state, 0, "A") }()
	go func() { defer wg.Done(); <-start; h.executor.completeNode(state, 1, "B") }()
	close(start)
	wg.Wait()

	state.mu.Lock()
	aPaths := cloneStrMap(state.exec.Nodes[0].OutputPaths)
	bPaths := cloneStrMap(state.exec.Nodes[1].OutputPaths)
	state.mu.Unlock()

	require.NotEmpty(t, aPaths, "A.OutputPaths must be populated")
	require.NotEmpty(t, bPaths, "B.OutputPaths must be populated")
	assert.Contains(t, aPaths, "sprint-status.yaml")
	assert.Contains(t, bPaths, "sprint-status.yaml")
}

func cloneStrMap(src map[string]string) map[string]string {
	cp := make(map[string]string, len(src))
	for k, v := range src {
		cp[k] = v
	}
	return cp
}

// AC-6: Legacy JSON without outputPaths round-trips cleanly (omitempty).
func TestAC6_LegacyJSON_RoundTrip_NoStaleOutputPaths(t *testing.T) {
	legacy := `{"id":"n1","processId":"bmad-brainstorming","label":"X",` +
		`"position":{"x":0,"y":0},"status":"pending",` +
		`"config":{"inputPath":"","outputPath":""},"tmuxTarget":""}`

	var node WorkflowNode
	require.NoError(t, json.Unmarshal([]byte(legacy), &node))

	out, err := json.Marshal(node)
	require.NoError(t, err)
	s := string(out)

	assert.NotContains(t, s, `"outputPaths"`, "omitempty must drop empty outputPaths; got %s", s)
	assert.NotContains(t, s, `"inputPaths"`, "omitempty must drop empty inputPaths; got %s", s)
}

// AC-7: NodeArtifactEvent.Paths marshals under "paths" key.
func TestNodeArtifactEvent_PathsField_Serializes(t *testing.T) {
	ev := NodeArtifactEvent{
		ExecID: "e1", NodeID: "n1",
		Found: []string{"PRD.md"}, Missing: []string{},
		Paths: map[string]string{"PRD.md": "/repo/_bmad-output/planning-artifacts/PRD.md"},
	}
	out, err := json.Marshal(ev)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, `"paths"`, "Paths field must serialise with JSON key 'paths'")
	assert.Contains(t, s, `"PRD.md":"/repo/_bmad-output/planning-artifacts/PRD.md"`)
}
