// Tests for the synchronous util-file-loader executor branch and for the
// upstream-context augmentation inside buildInteractivePrompt. Both ship
// as a post-sprint bug fix — the original util-file-loader spawned an
// empty-skill claude session that produced no usable output, and
// buildInteractivePrompt dropped incoming edges on the floor so
// downstream interactive nodes saw no upstream data.
package bmad

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// saveFileLoaderWorkflow builds a single util-file-loader node with the
// given filePath config.
func saveFileLoaderWorkflow(t *testing.T, s *Storage, id, filePath string) string {
	t.Helper()
	wf := WorkflowDef{
		ID:   id,
		Name: id,
		Nodes: []WorkflowNode{{
			ID:        "L1",
			Label:     "File Loader",
			ProcessID: "util-file-loader",
			Position:  Position{X: 0, Y: 0},
			Status:    NodePending,
			Config:    map[string]string{"filePath": filePath},
		}},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-20T00:00:00Z",
		UpdatedAt: "2026-04-20T00:00:00Z",
	}
	require.NoError(t, s.SaveWorkflow(wf))
	return wf.ID
}

func TestFileLoader_HappyPath_PopulatesOutputsAndPaths(t *testing.T) {
	h := newHarness(t)
	repoDir := t.TempDir()
	target := filepath.Join(repoDir, "brief.md")
	require.NoError(t, os.WriteFile(target, []byte("hello file loader"), 0o644))

	wfID := saveFileLoaderWorkflow(t, h.storage, "wf-fl-happy", target)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	require.Equal(t, ExecComplete, ex.Status)
	require.Len(t, ex.Nodes, 1)
	assert.Equal(t, NodeComplete, ex.Nodes[0].Status)

	assert.Equal(t, "hello file loader", ex.NodeOutputs["L1"],
		"NodeOutputs must carry file contents so downstream context builders find them")

	require.NotNil(t, ex.Nodes[0].OutputPaths)
	assert.Equal(t, target, ex.Nodes[0].OutputPaths["file-path"],
		"OutputPaths[\"file-path\"] must carry the absolute resolved path")

	events := h.eventsByName("bmad:node:artifacts")
	require.NotEmpty(t, events)
	// Multiple artifact events may fire (one from executeFileLoader, one
	// from completeNode's trailing bookkeeping). Assert at least one
	// carries the file-path entry we care about.
	sawFilePath := false
	for _, ev := range events {
		ae, ok := ev.data.(NodeArtifactEvent)
		if !ok {
			continue
		}
		if ae.NodeID != "L1" {
			continue
		}
		if p, ok := ae.Paths["file-path"]; ok && p == target {
			sawFilePath = true
			break
		}
	}
	assert.True(t, sawFilePath, "expected a bmad:node:artifacts event with file-path=%s; got %+v", target, events)
}

func TestFileLoader_RelativePath_ResolvedAgainstRepo(t *testing.T) {
	h := newHarness(t)
	repoDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, "doc.md"), []byte("relative-ok"), 0o644))

	wfID := saveFileLoaderWorkflow(t, h.storage, "wf-fl-rel", "doc.md")
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	require.Equal(t, ExecComplete, ex.Status)
	assert.Equal(t, "relative-ok", ex.NodeOutputs["L1"])
	assert.Equal(t, filepath.Join(repoDir, "doc.md"), ex.Nodes[0].OutputPaths["file-path"])
}

func TestFileLoader_PathTraversal_Rejected(t *testing.T) {
	h := newHarness(t)
	repoDir := t.TempDir()
	// Try to escape via "../"; the target file could exist outside repoDir
	// but the node must still fail because the relative-path containment
	// check rejects the escape before stat.
	outside := filepath.Join(filepath.Dir(repoDir), "escaped.md")
	require.NoError(t, os.WriteFile(outside, []byte("should not load"), 0o644))
	t.Cleanup(func() { _ = os.Remove(outside) })

	wfID := saveFileLoaderWorkflow(t, h.storage, "wf-fl-traversal", "../escaped.md")
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	require.NotEqual(t, ExecComplete, ex.Status, "path-traversal must not complete")
	require.Len(t, ex.Nodes, 1)
	assert.Equal(t, NodeFailed, ex.Nodes[0].Status)
	assert.Contains(t, ex.NodeOutputs["L1"], "outside repo root")
}

func TestFileLoader_MissingFile_Fails(t *testing.T) {
	h := newHarness(t)
	repoDir := t.TempDir()
	missing := filepath.Join(repoDir, "nope.md")

	wfID := saveFileLoaderWorkflow(t, h.storage, "wf-fl-missing", missing)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	require.NotEqual(t, ExecComplete, ex.Status)
	assert.Equal(t, NodeFailed, ex.Nodes[0].Status)
	assert.Contains(t, ex.NodeOutputs["L1"], "stat")
}

func TestFileLoader_EmptyFilePath_Fails(t *testing.T) {
	h := newHarness(t)
	repoDir := t.TempDir()

	wfID := saveFileLoaderWorkflow(t, h.storage, "wf-fl-empty", "")
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	require.NotEqual(t, ExecComplete, ex.Status)
	assert.Equal(t, NodeFailed, ex.Nodes[0].Status)
	assert.Contains(t, ex.NodeOutputs["L1"], "filePath not configured")
}

func TestFileLoader_Directory_Fails(t *testing.T) {
	h := newHarness(t)
	repoDir := t.TempDir()
	sub := filepath.Join(repoDir, "subdir")
	require.NoError(t, os.Mkdir(sub, 0o755))

	wfID := saveFileLoaderWorkflow(t, h.storage, "wf-fl-dir", sub)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	require.NotEqual(t, ExecComplete, ex.Status)
	assert.Equal(t, NodeFailed, ex.Nodes[0].Status)
	assert.Contains(t, ex.NodeOutputs["L1"], "directory")
}

func TestFileLoader_LargeFile_Truncated(t *testing.T) {
	h := newHarness(t)
	repoDir := t.TempDir()
	target := filepath.Join(repoDir, "big.txt")
	big := make([]byte, fileLoaderMaxBytes+4096)
	for i := range big {
		big[i] = byte('a' + (i % 26))
	}
	require.NoError(t, os.WriteFile(target, big, 0o644))

	wfID := saveFileLoaderWorkflow(t, h.storage, "wf-fl-big", target)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, repoDir, "sonnet")
	require.NoError(t, err)

	ex := waitForTerminal(t, h, exec.ID)
	require.Equal(t, ExecComplete, ex.Status)
	assert.Equal(t, fileLoaderMaxBytes, len(ex.NodeOutputs["L1"]),
		"oversized file must be truncated at fileLoaderMaxBytes")
}

// buildInteractivePrompt tests —— verify that upstream NodeOutputs and
// OutputPaths flow into the prompt when state/nodeID are provided.

func newStateWithUpstream(t *testing.T) *execState {
	t.Helper()
	exec := &WorkflowExecution{
		ID: "e1",
		Nodes: []WorkflowNode{
			{ID: "src", Label: "File Loader", Status: NodeComplete, OutputPaths: map[string]string{"file-path": "/repo/brief.md"}},
			{ID: "dst", Label: "Party Mode", Status: NodePending},
		},
		NodeOutputs: map[string]string{"src": "brief contents here"},
	}
	outEdges := map[string][]WorkflowEdge{
		"src": {{ID: "e", Source: "src", Target: "dst"}},
	}
	return &execState{
		exec:     exec,
		outEdges: outEdges,
		mu:       sync.Mutex{},
	}
}

func TestBuildInteractivePrompt_IncludesUpstreamContext(t *testing.T) {
	proc := ProcessDef{Name: "Party Mode", Description: "chat"}
	resolved := resolvedInputs{}
	state := newStateWithUpstream(t)

	got := buildInteractivePrompt(proc, resolved, state, "dst")
	assert.Contains(t, got, "## Upstream context — File Loader (src)",
		"prompt must surface an upstream block per incoming edge")
	assert.Contains(t, got, "**Path:** /repo/brief.md",
		"OutputPaths[\"file-path\"] must appear as the Path: line")
	assert.Contains(t, got, "brief contents here",
		"NodeOutputs for the upstream must embed as the content block")
}

func TestBuildInteractivePrompt_NoUpstream_SkipsBlock(t *testing.T) {
	proc := ProcessDef{Name: "Party Mode"}
	state := &execState{exec: &WorkflowExecution{ID: "e", Nodes: []WorkflowNode{{ID: "dst"}}}, outEdges: map[string][]WorkflowEdge{}}

	got := buildInteractivePrompt(proc, resolvedInputs{}, state, "dst")
	assert.NotContains(t, got, "Upstream context",
		"without incoming edges the upstream heading must not appear")
}

func TestBuildInteractivePrompt_NilState_Compatible(t *testing.T) {
	proc := ProcessDef{
		Name: "Guided Brief",
		InputSpecs: []InputSpec{{ID: "mode", Source: InputFromUser, Shape: ShapeChoice}},
	}
	got := buildInteractivePrompt(proc, resolvedInputs{"mode": "guided"}, nil, "")
	assert.Contains(t, got, "mode: guided")
	assert.NotContains(t, got, "Upstream context",
		"passing nil state disables upstream injection — useful for the resume path that supplies its own recap")
}

func TestBuildInteractivePrompt_TruncatesLargeUpstream(t *testing.T) {
	huge := strings.Repeat("x", interactivePromptUpstreamCap+4096)
	state := &execState{
		exec: &WorkflowExecution{
			ID: "e",
			Nodes: []WorkflowNode{
				{ID: "src", Label: "Loader"},
				{ID: "dst"},
			},
			NodeOutputs: map[string]string{"src": huge},
		},
		outEdges: map[string][]WorkflowEdge{
			"src": {{ID: "e1", Source: "src", Target: "dst"}},
		},
	}
	got := buildInteractivePrompt(ProcessDef{Name: "P"}, resolvedInputs{}, state, "dst")
	// Full untruncated string never appears verbatim; a substring of length
	// interactivePromptUpstreamCap does.
	assert.NotContains(t, got, huge)
	assert.Contains(t, got, strings.Repeat("x", interactivePromptUpstreamCap))
}

// ── PendingPrompt.LastOutput + extractModalQuestion ────────────────────────────

func TestExtractModalQuestion_Sentinel(t *testing.T) {
	raw := "noise before\n<MASHED_PROMPT>Pick a color: red, blue, green</MASHED_PROMPT>\nnoise after"
	assert.Equal(t, "Pick a color: red, blue, green", extractModalQuestion(raw))
}

func TestExtractModalQuestion_SentinelWins_OverTailFallback(t *testing.T) {
	huge := strings.Repeat("x", modalQuestionCap+500)
	raw := huge + "\n<MASHED_PROMPT>clean question</MASHED_PROMPT>"
	got := extractModalQuestion(raw)
	assert.Equal(t, "clean question", got, "sentinel body must be preferred even when the tail exceeds the cap")
}

func TestExtractModalQuestion_TailFallback_Truncates(t *testing.T) {
	body := strings.Repeat("a", modalQuestionCap+2048)
	got := extractModalQuestion(body)
	assert.LessOrEqual(t, len(got), modalQuestionCap, "tail fallback must truncate to modalQuestionCap")
	assert.NotEqual(t, body, got, "fallback must not return the full oversize input verbatim")
}

func TestExtractModalQuestion_EmptyPassThrough(t *testing.T) {
	assert.Equal(t, "", extractModalQuestion(""))
	assert.Equal(t, "", extractModalQuestion("   \n  "))
}

func TestExtractModalQuestion_UnclosedSentinel_FallsBackToTail(t *testing.T) {
	raw := "content <MASHED_PROMPT>not closed"
	got := extractModalQuestion(raw)
	// No closing tag → tail fallback; full string fits under cap so it
	// returns intact.
	assert.Contains(t, got, "MASHED_PROMPT>not closed")
}
