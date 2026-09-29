package bmad

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadExecutionFromDisk_FiltersGhostTmuxSessions — a snapshot with
// status=running and a TmuxTarget that points at a dead session must be
// demoted to failed and skipped. Prevents the "Your turn" notification
// from a zombie execution that can never be resumed.
func TestLoadExecutionFromDisk_FiltersGhostTmuxSessions(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	workflowsDir := filepath.Join(tmp, ".mashed", "workflows", "exec-ghost")
	require.NoError(t, os.MkdirAll(workflowsDir, 0o755))

	repo := "/tmp/ghost-repo"
	exec := WorkflowExecution{
		ID:       "exec-ghost",
		RepoPath: repo,
		Status:   ExecRunning,
		Nodes: []WorkflowNode{{
			ID:         "n1",
			Status:     NodeAwaitingInput,
			TmuxTarget: "bmad-ghost-session-that-does-not-exist",
		}},
	}
	raw, err := json.Marshal(exec)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(workflowsDir, "execution.json"), raw, 0o644))

	got, err := LoadExecutionFromDisk(repo)
	require.NoError(t, err)
	assert.Nil(t, got, "ghost execution with dead tmux target must not surface")

	// And the on-disk status must be demoted so subsequent loads skip.
	patched, err := os.ReadFile(filepath.Join(workflowsDir, "execution.json"))
	require.NoError(t, err)
	var reloaded WorkflowExecution
	require.NoError(t, json.Unmarshal(patched, &reloaded))
	assert.Equal(t, ExecFailed, reloaded.Status,
		"LoadExecutionFromDisk must stamp the ghost execution as failed on disk")
}

// TestLoadExecutionFromDisk_PreservesExecutionWithoutTmuxTargets — a
// running execution that has no TmuxTarget on any node (e.g. fresh
// load before any process started) must remain loadable so resume
// flows keep working.
func TestLoadExecutionFromDisk_PreservesExecutionWithoutTmuxTargets(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	workflowsDir := filepath.Join(tmp, ".mashed", "workflows", "exec-fresh")
	require.NoError(t, os.MkdirAll(workflowsDir, 0o755))

	repo := "/tmp/fresh-repo"
	exec := WorkflowExecution{
		ID:       "exec-fresh",
		RepoPath: repo,
		Status:   ExecRunning,
		Nodes:    []WorkflowNode{{ID: "n1", Status: NodePending}},
	}
	raw, _ := json.Marshal(exec)
	require.NoError(t, os.WriteFile(filepath.Join(workflowsDir, "execution.json"), raw, 0o644))

	got, err := LoadExecutionFromDisk(repo)
	require.NoError(t, err)
	require.NotNil(t, got, "fresh execution with no tmux targets must surface unchanged")
	assert.Equal(t, ExecRunning, got.Status)
}

// TestTmuxSessionAlive_DeadSessionReturnsFalse — unit test on the probe.
// We can't assert a live one portably (no tmux in CI), but we can assert a
// bogus name fails.
func TestTmuxSessionAlive_DeadSessionReturnsFalse(t *testing.T) {
	t.Parallel()
	assert.False(t, tmuxSessionAlive("bmad-definitely-not-a-real-session-xyz-987654321"))
	assert.False(t, tmuxSessionAlive(""))
}
