// Story breadcrumbs-07 RED tests: MultiFileLoader executor branch must
// resolve each configured entry, populate OutputPaths under label or
// positional file[N], emit NodeArtifactEvent, and reject invalid config
// (duplicate labels, >64 entries). Types are scaffolded in types.go so
// these tests COMPILE — runtime behaviour is unimplemented.

package bmad

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// saveMultiFileWorkflow builds and persists a single-node workflow whose
// node is a MultiFileLoader with the given entries JSON-encoded into
// Config["entries"].
func saveMultiFileWorkflow(t *testing.T, s *Storage, id string, entries []MultiFileEntry) string {
	t.Helper()
	j, err := json.Marshal(entries)
	require.NoError(t, err)
	wf := WorkflowDef{
		ID:   id,
		Name: id,
		Nodes: []WorkflowNode{{
			ID:       "N1",
			Label:    "Loader",
			Position: Position{X: 0, Y: 0},
			Status:   NodePending,
			NodeType: NodeTypeMultiFileLoader,
			Config:   map[string]string{"entries": string(j)},
		}},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-14T00:00:00Z",
		UpdatedAt: "2026-04-14T00:00:00Z",
	}
	require.NoError(t, s.SaveWorkflow(wf))
	return wf.ID
}

// waitForTerminal returns the final execution once it reaches a terminal state
// (Complete, Paused from node failure, or Failed), or fails the test on timeout.
// Note: the executor sets ExecPaused (not ExecFailed) when a node fails so the
// user can resume. We treat ExecPaused as terminal for validation-failure tests.
func waitForTerminal(t *testing.T, h *testHarness, execID string) *WorkflowExecution {
	t.Helper()
	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(execID)
		return ex != nil && (ex.Status == ExecComplete || ex.Status == ExecFailed || ex.Status == ExecPaused)
	}, 5*time.Second, 50*time.Millisecond)
	ex, err := h.executor.GetExecution(execID)
	require.NoError(t, err)
	return ex
}

// AC-1: NodeTypeMultiFileLoader round-trips through JSON and EffectiveType.
func TestMultiFileLoader_AC1_NodeTypeRoundTrip(t *testing.T) {
	original := WorkflowNode{
		ID:       "N1",
		Label:    "Loader",
		Status:   NodePending,
		NodeType: NodeTypeMultiFileLoader,
		Position: Position{X: 0, Y: 0},
		Config:   map[string]string{},
	}
	b, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Contains(t, string(b), `"nodeType":"multiFileLoader"`)

	var restored WorkflowNode
	require.NoError(t, json.Unmarshal(b, &restored))
	assert.Equal(t, NodeTypeMultiFileLoader, restored.NodeType)
	assert.Equal(t, NodeTypeMultiFileLoader, restored.EffectiveType())
}

// AC-2: Labeled + positional mix — both files present.
func TestMultiFileLoader_AC2_LabeledAndPositional_EmitsOutputPaths(t *testing.T) {
	h := newHarness(t)

	repoDir := t.TempDir()
	briefPath := filepath.Join(repoDir, "brief.md")
	notesPath := filepath.Join(repoDir, "notes.md")
	require.NoError(t, os.WriteFile(briefPath, []byte("brief"), 0o644))
	require.NoError(t, os.WriteFile(notesPath, []byte("notes"), 0o644))

	entries := []MultiFileEntry{
		{Label: "brief", Path: briefPath},
		{Label: "", Path: notesPath},
	}
	wfID := saveMultiFileWorkflow(t, h.storage, "wf-mfl-ac2", entries)

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	require.Equal(t, ExecComplete, ex.Status)
	require.Len(t, ex.Nodes, 1)
	assert.Equal(t, NodeComplete, ex.Nodes[0].Status)

	require.NotNil(t, ex.Nodes[0].OutputPaths)
	assert.Equal(t, briefPath, ex.Nodes[0].OutputPaths["brief"],
		"labeled entry must surface under its label; got %v", ex.Nodes[0].OutputPaths)
	assert.Equal(t, notesPath, ex.Nodes[0].OutputPaths["file[1]"],
		"unlabeled entry at index 1 must surface as file[1]; got %v", ex.Nodes[0].OutputPaths)

	events := h.eventsByName("bmad:node:artifacts")
	require.GreaterOrEqual(t, len(events), 1)
	ae, ok := events[len(events)-1].data.(NodeArtifactEvent)
	require.True(t, ok)
	assert.Equal(t, briefPath, ae.Paths["brief"])
	assert.Equal(t, notesPath, ae.Paths["file[1]"])
}

// AC-3: Missing file on disk is tolerated — skipped, not failed.
func TestMultiFileLoader_AC3_MissingFileTolerated(t *testing.T) {
	h := newHarness(t)

	repoDir := t.TempDir()
	missingPath := filepath.Join(repoDir, "nope.md") // never written

	entries := []MultiFileEntry{{Label: "brief", Path: missingPath}}
	wfID := saveMultiFileWorkflow(t, h.storage, "wf-mfl-ac3", entries)

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	assert.Equal(t, ExecComplete, ex.Status, "missing file must not fail the execution")
	require.Len(t, ex.Nodes, 1)
	assert.Equal(t, NodeComplete, ex.Nodes[0].Status)

	require.NotNil(t, ex.Nodes[0].OutputPaths)
	_, hasBrief := ex.Nodes[0].OutputPaths["brief"]
	assert.False(t, hasBrief, "missing file must be omitted from OutputPaths")

	events := h.eventsByName("bmad:node:artifacts")
	require.GreaterOrEqual(t, len(events), 1)
	ae, ok := events[len(events)-1].data.(NodeArtifactEvent)
	require.True(t, ok)
	assert.Contains(t, ae.Missing, "brief",
		"NodeArtifactEvent.Missing must surface skipped labels; got %v", ae.Missing)
}

// AC-4: Duplicate non-empty labels cause a validation failure wrapping
// ErrDuplicateMultiFileLabel.
func TestMultiFileLoader_AC4_DuplicateLabelsRejected(t *testing.T) {
	h := newHarness(t)

	repoDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, "a.md"), []byte("a"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, "b.md"), []byte("b"), 0o644))

	entries := []MultiFileEntry{
		{Label: "brief", Path: filepath.Join(repoDir, "a.md")},
		{Label: "brief", Path: filepath.Join(repoDir, "b.md")},
	}
	wfID := saveMultiFileWorkflow(t, h.storage, "wf-mfl-ac4", entries)

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	assert.Equal(t, ExecFailed, ex.Status, "duplicate labels must fail the execution")
	require.Len(t, ex.Nodes, 1)
	assert.Equal(t, NodeFailed, ex.Nodes[0].Status)

	// Executor records the failure reason in NodeOutputs[nodeID]. The
	// stored message must wrap ErrDuplicateMultiFileLabel (compared by
	// string since NodeOutputs is map[string]string).
	require.NotNil(t, ex.NodeOutputs)
	reason := ex.NodeOutputs["N1"]
	assert.Contains(t, reason, ErrDuplicateMultiFileLabel.Error(),
		"NodeOutputs[N1] must surface the sentinel error text; got %q", reason)
}

// TooMany: >64 entries is rejected with ErrMultiFileTooMany.
func TestMultiFileLoader_TooManyEntries_Rejected(t *testing.T) {
	h := newHarness(t)
	repoDir := t.TempDir()

	entries := make([]MultiFileEntry, 65)
	for i := range entries {
		entries[i] = MultiFileEntry{Path: filepath.Join(repoDir, "f.md")}
	}
	wfID := saveMultiFileWorkflow(t, h.storage, "wf-mfl-toomany", entries)

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	assert.Equal(t, ExecFailed, ex.Status, "too-many entries must fail the execution")
	require.Len(t, ex.Nodes, 1)
	assert.Equal(t, NodeFailed, ex.Nodes[0].Status)

	require.NotNil(t, ex.NodeOutputs)
	reason := ex.NodeOutputs["N1"]
	assert.True(t, strings.Contains(reason, ErrMultiFileTooMany.Error()),
		"NodeOutputs[N1] must surface ErrMultiFileTooMany; got %q", reason)
}

// Relative path resolution — entry path resolved against repoPath.
func TestMultiFileLoader_RelativePath_ResolvedAgainstRepo(t *testing.T) {
	h := newHarness(t)

	repoDir := t.TempDir()
	subDir := filepath.Join(repoDir, "notes")
	require.NoError(t, os.MkdirAll(subDir, 0o755))
	absFile := filepath.Join(subDir, "brief.md")
	require.NoError(t, os.WriteFile(absFile, []byte("x"), 0o644))

	entries := []MultiFileEntry{{Label: "", Path: "notes/brief.md"}}
	wfID := saveMultiFileWorkflow(t, h.storage, "wf-mfl-rel", entries)

	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	require.Equal(t, ExecComplete, ex.Status)
	require.Len(t, ex.Nodes, 1)
	require.NotNil(t, ex.Nodes[0].OutputPaths)
	got := ex.Nodes[0].OutputPaths["file[0]"]
	assert.Equal(t, absFile, got, "relative path must resolve against repoPath")
}

// AC-5: Registry surfaces a MultiFileLoader entry under PhaseUtilities.
func TestMultiFileLoader_AC5_RegistrySurface(t *testing.T) {
	procs := AllProcesses()
	var found *ProcessDef
	for i := range procs {
		if procs[i].Phase == PhaseUtilities && (procs[i].ID == "util-multi-file-loader" ||
			procs[i].Name == "Multi File Loader") {
			p := procs[i]
			found = &p
			break
		}
	}
	require.NotNil(t, found, "registry must contain a util-multi-file-loader ProcessDef under PhaseUtilities")
	assert.Equal(t, PhaseUtilities, found.Phase)
	assert.Empty(t, found.Inputs, "MultiFileLoader has no declared inputs")
}
