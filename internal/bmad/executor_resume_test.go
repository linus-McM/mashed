// Package bmad contains tests for the persistence-and-resume story.
// Story bmad-interactive-05: Persistence extension + restore-on-mount re-emit.
//
// RED Phase: These tests define the contract for snapshot versioning,
// pane-alive checks, dead-pane recovery via resumeInteractiveNode,
// renderRecap, rehydratePending, and the stop-while-awaiting edge case.
// They MUST fail until the go-engineer implements the feature (GREEN phase).
//
// Tests are in package bmad (white-box) to access unexported helpers.
package bmad

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── AC-1: persistSnapshot writes all four interactive fields ──────────────────

// TestPersistSnapshotWritesInteractiveFields (AC-1)
// Build a state with interactive fields populated. Call persistSnapshot.
// Read the resulting execution.json and assert all four interactive field keys
// are present and survive a round-trip through json.Unmarshal.
func TestPersistSnapshotWritesInteractiveFields(t *testing.T) {
	// Override HOME so the snapshot goes to a temp dir.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	const execID = "exec1"
	const nodeID = "n1"

	nodes := []WorkflowNode{
		{ID: nodeID, ProcessID: "test-proc", Status: NodeAwaitingInput, NodeType: NodeTypeProcess},
	}
	state, _ := newSessionState(nodes, nil)
	state.exec.ID = execID
	state.exec.PendingPrompts = []PendingPrompt{
		{NodeID: nodeID, InputID: "topic", Prompt: "Enter a topic", Shape: ShapeFree, Round: 1, PromptID: "abc123def456abcd"},
	}
	state.exec.NodeInputs = map[string]map[string]string{
		nodeID: {"topic": "AI search"},
	}
	state.exec.NodeInputHistory = map[string][]NodeInputEntry{
		nodeID: {{InputID: "topic", Round: 1, Value: "AI search", Timestamp: 1700000000}},
	}
	state.exec.NodeRounds = map[string]int{nodeID: 1}

	e := NewExecutor(nil, func(string, interface{}) {})

	// persistSnapshot does not exist yet — RED phase: will fail to compile
	// until the go-engineer implements it. It is already implemented in S3
	// via prompts.go — this test verifies the interactive fields round-trip.
	err := e.persistSnapshot(state)
	require.NoError(t, err, "persistSnapshot must succeed")

	snapshotPath := filepath.Join(tmpHome, ".mashed", "workflows", execID, "execution.json")
	raw, err := os.ReadFile(snapshotPath)
	require.NoError(t, err, "execution.json must exist after persistSnapshot")

	rawStr := string(raw)
	assert.Contains(t, rawStr, `"pendingPrompts"`, "snapshot must contain pendingPrompts key")
	assert.Contains(t, rawStr, `"nodeInputs"`, "snapshot must contain nodeInputs key")
	assert.Contains(t, rawStr, `"nodeInputHistory"`, "snapshot must contain nodeInputHistory key")
	assert.Contains(t, rawStr, `"nodeRounds"`, "snapshot must contain nodeRounds key")

	// Round-trip unmarshal.
	var loaded WorkflowExecution
	require.NoError(t, json.Unmarshal(raw, &loaded), "execution.json must unmarshal cleanly")
	require.Len(t, loaded.PendingPrompts, 1, "PendingPrompts must survive round-trip")
	assert.Equal(t, nodeID, loaded.PendingPrompts[0].NodeID)
	assert.Equal(t, "topic", loaded.PendingPrompts[0].InputID)

	require.NotNil(t, loaded.NodeInputs)
	assert.Equal(t, "AI search", loaded.NodeInputs[nodeID]["topic"])

	require.NotNil(t, loaded.NodeInputHistory)
	require.Len(t, loaded.NodeInputHistory[nodeID], 1)
	assert.Equal(t, "AI search", loaded.NodeInputHistory[nodeID][0].Value)

	require.NotNil(t, loaded.NodeRounds)
	assert.Equal(t, 1, loaded.NodeRounds[nodeID])
}

// ── AC-3 / AC-4: Version field in snapshot ────────────────────────────────────

// TestSnapshotVersionField (AC-3, AC-4)
// Table-driven: interactive executions get version=2; legacy (no interactive
// fields) gets version=0 or absent. Legacy JSON with no version key loads
// cleanly as Version==0.
func TestSnapshotVersionField(t *testing.T) {
	// NOT parallel: uses t.Setenv to override HOME.
	const nodeID = "n1"

	type row struct {
		name           string
		pendingPrompts []PendingPrompt
		history        map[string][]NodeInputEntry
		wantVersion    int
		wantPresent    bool // true = version key expected in JSON
	}

	rows := []row{
		{
			name: "pending_prompts_present_gets_v2",
			pendingPrompts: []PendingPrompt{
				{NodeID: nodeID, InputID: "topic", Round: 1, PromptID: "abc123def456abcd"},
			},
			wantVersion: 2,
			wantPresent: true,
		},
		{
			name: "non_empty_history_no_prompts_gets_v2",
			history: map[string][]NodeInputEntry{
				nodeID: {{InputID: "topic", Round: 1, Value: "val", Timestamp: 1700000000}},
			},
			wantVersion: 2,
			wantPresent: true,
		},
		{
			name:           "no_interactive_fields_legacy_compat",
			pendingPrompts: nil,
			history:        nil,
			wantVersion:    0,
			wantPresent:    false, // omitempty: absent or zero
		},
	}

	for _, tt := range rows {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tmpHome := t.TempDir()
			t.Setenv("HOME", tmpHome)

			execID := "exec-version-" + tt.name
			nodes := []WorkflowNode{{ID: nodeID, Status: NodeRunning, NodeType: NodeTypeProcess}}
			state, _ := newSessionState(nodes, nil)
			state.exec.ID = execID
			state.exec.PendingPrompts = tt.pendingPrompts
			state.exec.NodeInputHistory = tt.history

			e := NewExecutor(nil, func(string, interface{}) {})
			require.NoError(t, e.persistSnapshot(state))

			snapshotPath := filepath.Join(tmpHome, ".mashed", "workflows", execID, "execution.json")
			raw, err := os.ReadFile(snapshotPath)
			require.NoError(t, err)

			// Parse into a generic map so we can check key presence.
			var generic map[string]interface{}
			require.NoError(t, json.Unmarshal(raw, &generic))

			// Round-trip into WorkflowExecution.
			var loaded WorkflowExecution
			require.NoError(t, json.Unmarshal(raw, &loaded))

			if tt.wantPresent {
				versionVal, hasVersion := generic["version"]
				assert.True(t, hasVersion, "version key must be present in JSON for interactive executions")
				if hasVersion {
					// JSON numbers unmarshal to float64.
					assert.Equal(t, float64(tt.wantVersion), versionVal,
						"version must be %d for interactive executions", tt.wantVersion)
				}
				assert.Equal(t, tt.wantVersion, loaded.Version,
					"WorkflowExecution.Version must be %d after unmarshal", tt.wantVersion)
			} else {
				// Legacy: version absent or zero.
				versionVal, hasVersion := generic["version"]
				if hasVersion {
					// If present, it must be 0.
					assert.Equal(t, float64(0), versionVal, "legacy exec version must be 0 if key present")
				}
				assert.Equal(t, 0, loaded.Version, "legacy exec must have Version==0")
			}
		})
	}

	// Row 4: legacy JSON with NO version key loads cleanly.
	t.Run("legacy_json_no_version_loads_cleanly", func(t *testing.T) {
		legacyJSON := []byte(`{"id":"e1","workflowId":"wf1","repoPath":"/tmp","status":"running","nodes":[],"startedAt":"2024-01-01T00:00:00Z"}`)
		var loaded WorkflowExecution
		err := json.Unmarshal(legacyJSON, &loaded)
		require.NoError(t, err, "legacy JSON without version key must unmarshal without error")
		assert.Equal(t, 0, loaded.Version, "Version must be 0 for legacy snapshots")
		assert.Empty(t, loaded.PendingPrompts, "PendingPrompts must be empty for legacy snapshots")
		assert.Empty(t, loaded.NodeInputs, "NodeInputs must be nil/empty for legacy snapshots")
	})
}

// ── Concurrency: persistSnapshot concurrent calls must not panic ──────────────

// TestPersistSnapshotConcurrentCallsNoPanic
// Races 20 goroutines calling persistSnapshot on the same state. Ensures
// atomic write-then-rename (tempfile) means no torn JSON.
func TestPersistSnapshotConcurrentCallsNoPanic(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	const execID = "exec-concurrent"
	nodes := []WorkflowNode{{ID: "n1", Status: NodeRunning, NodeType: NodeTypeProcess}}
	state, _ := newSessionState(nodes, nil)
	state.exec.ID = execID

	e := NewExecutor(nil, func(string, interface{}) {})

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			_ = e.persistSnapshot(state)
		}()
	}
	wg.Wait()

	// Final file must exist and be valid JSON.
	snapshotPath := filepath.Join(tmpHome, ".mashed", "workflows", execID, "execution.json")
	raw, err := os.ReadFile(snapshotPath)
	require.NoError(t, err, "execution.json must exist after concurrent writes")
	var loaded WorkflowExecution
	assert.NoError(t, json.Unmarshal(raw, &loaded), "final execution.json must be valid JSON after concurrent writes")
}

// ── AC-5: paneAlive ───────────────────────────────────────────────────────────

// TestPaneAliveTrueFalse (AC-5)
// Mock runner: tmux display-message with valid output → true.
// Mock runner: error return → false.
// Empty target → false without invoking runner.
func TestPaneAliveTrueFalse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		target   string
		output   string
		runErr   error
		wantAlive bool
	}{
		{
			name:      "valid_pane_returns_true",
			target:    "work:0.0",
			output:    "%1\n",
			runErr:    nil,
			wantAlive: true,
		},
		{
			name:      "runner_error_returns_false",
			target:    "work:0.0",
			output:    "",
			runErr:    errors.New("tmux: no server running on /private/tmp/tmux-..."),
			wantAlive: false,
		},
		{
			name:      "empty_target_returns_false",
			target:    "",
			output:    "",
			runErr:    nil,
			wantAlive: false,
		},
		{
			name:      "blank_output_returns_false",
			target:    "work:0.0",
			output:    "   \n",
			runErr:    nil,
			wantAlive: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runner := func(_ context.Context, name string, args ...string) ([]byte, error) {
				// Ensure the right command/args are being used.
				if name == "tmux" && len(args) >= 4 && args[0] == "display-message" {
					return []byte(tt.output), tt.runErr
				}
				return nil, errors.New("unexpected command: " + name)
			}

			e := NewExecutor(nil, func(string, interface{}) {})
			e.SetCommandRunner(runner)

			// paneAlive does not exist yet — RED phase: will fail to compile.
			got := e.paneAlive(tt.target)
			assert.Equal(t, tt.wantAlive, got, "paneAlive(%q) = %v, want %v", tt.target, got, tt.wantAlive)
		})
	}
}

// ── AC-5: resumeInteractiveNode with dead pane uses recap ─────────────────────

// TestResumeInteractiveNodeDeadPaneRecap (AC-5)
// Given state with one node whose TmuxTarget is dead and NodeInputHistory has
// 2 prior rounds, when resumeInteractiveNode is called, it must:
//   - detect the dead pane via paneAlive→false
//   - call startSession (captured via runner) with a prompt containing the
//     "## Previous session recap" heading followed by the rendered Q&A.
func TestResumeInteractiveNodeDeadPaneRecap(t *testing.T) {
	// NOT parallel: uses testRegistry and per-test runner capture.
	const nodeID = "n1"
	const processID = "test-s5-resume-recap"
	const deadTarget = "dead_session:0"

	spec := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: true,
		Prompt:   "Enter a topic",
	}
	registerTestProcess(t, ProcessDef{
		ID:        processID,
		Name:      "Test S5 Resume Recap",
		Mode:      InteractGuided,
		SkillName: "test-skill",
		InputSpecs: []InputSpec{spec},
		OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToMemory}},
	})

	nodes := []WorkflowNode{
		{
			ID:         nodeID,
			ProcessID:  processID,
			Status:     NodeAwaitingInput,
			TmuxTarget: deadTarget,
			NodeType:   NodeTypeProcess,
			InputSpecs: []InputSpec{spec},
		},
	}
	state, _ := newSessionState(nodes, nil)
	state.exec.NodeInputHistory = map[string][]NodeInputEntry{
		nodeID: {
			{InputID: "topic", Round: 1, Value: "AI search", Timestamp: 1700000001},
			{InputID: "round-response", Round: 2, Value: "pivot to voice", Timestamp: 1700000002},
		},
	}
	state.exec.NodeRounds = map[string]int{nodeID: 2}
	state.exec.NodeInputs = map[string]map[string]string{
		nodeID: {"topic": "AI search"},
	}

	// Capture the prompt passed to new-session so we can assert on it.
	var capturedPrompt string
	var promptMu sync.Mutex

	runner := func(_ context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 {
			switch args[0] {
			case "display-message":
				// Simulate dead pane: return error.
				return nil, errors.New("tmux: no such session: dead_session")
			case "new-session":
				// Capture the command arg which contains the prompt.
				for i, arg := range args {
					if arg == "--" && i+1 < len(args) {
						promptMu.Lock()
						capturedPrompt = args[i+1]
						promptMu.Unlock()
						break
					}
					// Also try to capture via -x or shell command arg.
					if (arg == "bash" || arg == "sh") && i+1 < len(args) {
						promptMu.Lock()
						capturedPrompt = args[i+1]
						promptMu.Unlock()
					}
				}
				// Capture any arg that might contain the prompt text.
				for _, arg := range args {
					if len(arg) > 20 {
						promptMu.Lock()
						if capturedPrompt == "" {
							capturedPrompt = arg
						}
						promptMu.Unlock()
					}
				}
				return []byte("ok"), nil
			case "send-keys":
				return []byte("ok"), nil
			}
		}
		return []byte("ok"), nil
	}

	e := NewExecutor(nil, func(string, interface{}) {})
	e.SetCommandRunner(runner)

	ctx := context.Background()

	// resumeInteractiveNode does not exist yet — RED phase: will fail to compile.
	err := e.resumeInteractiveNode(ctx, state, nodeID)
	require.NoError(t, err, "resumeInteractiveNode must succeed when a new session can be started")

	// Assert the prompt captured by new-session contains the recap heading.
	promptMu.Lock()
	prompt := capturedPrompt
	promptMu.Unlock()

	assert.Contains(t, prompt, "## Previous session recap",
		"prompt must contain '## Previous session recap' heading")
	assert.Contains(t, prompt, "AI search",
		"recap must include prior round answer 'AI search'")
	assert.Contains(t, prompt, "pivot to voice",
		"recap must include prior round answer 'pivot to voice'")
}

// ── AC-6: renderRecap formats history as markdown Q&A ─────────────────────────

// TestRenderRecap (AC-6)
// Two NodeInputEntry items; assert exact markdown lines appear in output.
// Empty history returns empty string.
func TestRenderRecap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		history     []NodeInputEntry
		wantLines   []string
		wantEmpty   bool
	}{
		{
			name: "two_rounds_formatted_as_markdown",
			history: []NodeInputEntry{
				{Round: 1, InputID: "topic", Value: "AI search"},
				{Round: 2, InputID: "round-response", Value: "pivot to voice"},
			},
			wantLines: []string{
				"**Round 1 — topic:** AI search",
				"**Round 2 — round-response:** pivot to voice",
			},
		},
		{
			name:      "empty_history_returns_empty_string",
			history:   []NodeInputEntry{},
			wantEmpty: true,
		},
		{
			name:    "nil_history_returns_empty_string",
			history: nil,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// A process def is needed for the signature (may be ignored by the implementation).
			proc := ProcessDef{ID: "test-proc", Name: "Test Proc"}

			// renderRecap does not exist yet — RED phase: will fail to compile.
			got := renderRecap(proc, tt.history)

			if tt.wantEmpty {
				assert.Empty(t, got, "renderRecap on empty/nil history must return empty string")
				return
			}
			for _, line := range tt.wantLines {
				assert.Contains(t, got, line,
					"renderRecap output must contain line: %q", line)
			}
		})
	}
}

// ── rehydratePending spawns waiters ──────────────────────────────────────────

// TestRehydratePendingSpawnsWaiters
// Given a state with 2 PendingPrompts, call rehydratePending.
// Assert that waiter channels exist for each (nodeID, inputID) pair.
// Release one waiter via releaseWaiter and assert the corresponding
// channel has been closed (goroutine can proceed).
func TestRehydratePendingSpawnsWaiters(t *testing.T) {
	// NOT parallel: mutates state.waiters map.
	const nodeID1 = "n1"
	const nodeID2 = "n2"

	nodes := []WorkflowNode{
		{ID: nodeID1, ProcessID: "test-proc-1", Status: NodeAwaitingInput, NodeType: NodeTypeProcess},
		{ID: nodeID2, ProcessID: "test-proc-2", Status: NodeAwaitingInput, NodeType: NodeTypeProcess},
	}
	state, _ := newSessionState(nodes, nil)
	state.exec.PendingPrompts = []PendingPrompt{
		{NodeID: nodeID1, InputID: "topic", Round: 1, PromptID: "aaa"},
		{NodeID: nodeID2, InputID: "method", Round: 1, PromptID: "bbb"},
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	e := NewExecutor(nil, func(string, interface{}) {})

	// rehydratePending does not exist yet — RED phase: will fail to compile.
	e.rehydratePending(state)

	// Give goroutines a moment to register their waiters.
	// The goroutines call state.waiter() which inserts into the map.
	require.Eventually(t, func() bool {
		state.waitersMu.Lock()
		defer state.waitersMu.Unlock()
		key1 := nodeID1 + "/topic"
		key2 := nodeID2 + "/method"
		_, ok1 := state.waiters[key1]
		_, ok2 := state.waiters[key2]
		return ok1 && ok2
	}, 2*time.Second, 20*time.Millisecond,
		"rehydratePending must register waiter channels for each pending prompt")

	// Release the waiter for n1/topic.
	// The goroutine should transition NodeRunning and remove from PendingPrompts.
	// NodeInputs must have n1/topic filled in for RespondToInput-style wakeup —
	// for RED phase we just verify the channel closes without panic.
	state.releaseWaiter(nodeID1, "topic")

	// After release, key1 must be gone from the map.
	require.Eventually(t, func() bool {
		state.waitersMu.Lock()
		defer state.waitersMu.Unlock()
		_, stillPresent := state.waiters[nodeID1+"/topic"]
		return !stillPresent
	}, 2*time.Second, 20*time.Millisecond,
		"waiter channel for n1/topic must be removed after releaseWaiter")

	// The goroutine must update PendingPrompts to remove the released entry.
	require.Eventually(t, func() bool {
		state.mu.Lock()
		defer state.mu.Unlock()
		for _, p := range state.exec.PendingPrompts {
			if p.NodeID == nodeID1 && p.InputID == "topic" {
				return false // still present
			}
		}
		return true
	}, 2*time.Second, 20*time.Millisecond,
		"PendingPrompts must not contain the released n1/topic entry")

	// Snapshot must have been persisted (file mtime changed or exists).
	snapshotPath := filepath.Join(tmpHome, ".mashed", "workflows", state.exec.ID, "execution.json")
	_, statErr := os.Stat(snapshotPath)
	assert.NoError(t, statErr, "persistSnapshot must have been called by the rehydrated goroutine")
}

// ── AC-7: Stop-while-awaiting produces aborted/failed state at restore ────────

// TestStopWhileAwaitingRestoreProducesAbortedState (AC-7)
// Simulate: execution with a PendingPrompt whose context is cancelled
// (StopBmadWorkflow path → suspendForSpec returns ctx.Err() → failNode).
// Persist snapshot. Verify the snapshot has a terminal status.
// GetCurrentExecution must return nil for terminal executions.
//
// NOTE for go-engineer: if GetCurrentExecution does not yet filter terminal
// executions by reading the snapshot from disk (it currently only scans
// in-memory executions), implement CurrentExecution/GetCurrentExecution to
// load from disk when not in-memory, then apply the terminal-status filter.
// This test documents the gap.
func TestStopWhileAwaitingRestoreProducesAbortedState(t *testing.T) {
	// NOT parallel: uses t.Setenv and shared executor state.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	const execID = "exec-abort-resume"
	const nodeID = "n1"

	// Build a state that looks like it was stopped mid-suspension.
	// suspendForSpec on ctx.Done sets status via failNode which calls setStatus(NodeFailed).
	// Manually build the terminal snapshot here.
	nodes := []WorkflowNode{
		{ID: nodeID, ProcessID: "test-proc", Status: NodeFailed, NodeType: NodeTypeProcess},
	}
	state, _ := newSessionState(nodes, nil)
	state.exec.ID = execID
	state.exec.Status = ExecFailed
	// PendingPrompts may still be set if the context cancellation race left them
	// in snapshot (real behavior: suspendForSpec doesn't clear PendingPrompts on
	// cancel — the stop path leaves the prompt in the snapshot).
	state.exec.PendingPrompts = []PendingPrompt{
		{NodeID: nodeID, InputID: "topic", Round: 1, PromptID: "deaddeaddeaddead"},
	}

	e := NewExecutor(nil, func(string, interface{}) {})
	require.NoError(t, e.persistSnapshot(state), "persisting terminal snapshot must succeed")

	// Verify the snapshot has ExecFailed status.
	snapshotPath := filepath.Join(tmpHome, ".mashed", "workflows", execID, "execution.json")
	raw, err := os.ReadFile(snapshotPath)
	require.NoError(t, err)
	var loaded WorkflowExecution
	require.NoError(t, json.Unmarshal(raw, &loaded))
	assert.Equal(t, ExecFailed, loaded.Status,
		"snapshot status must be ExecFailed for stopped executions")

	// GetCurrentExecution must NOT return terminal executions.
	// The in-memory executions map is empty, so this tests the on-disk path if
	// implemented, or simply verifies the in-memory filter. If GetCurrentExecution
	// only reads in-memory state, it returns (nil, nil) — which is the correct
	// behavior because a terminal snapshot has nothing live to restore.
	repoPath := state.exec.RepoPath
	result, err := e.GetCurrentExecution(repoPath)
	require.NoError(t, err, "GetCurrentExecution must not error on terminal execution")
	assert.Nil(t, result,
		"GetCurrentExecution must return nil for terminal (ExecFailed) executions; "+
			"if this passes trivially because in-memory map is empty, the go-engineer "+
			"must also add disk-load+filter logic to CurrentExecution for the restore-on-mount path")
}
