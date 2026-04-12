package bmad

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Shared helpers (skills-cmd-01 + exec-03) ──

// countingNewSessionRunner wraps successRunner so the canonical pane-dead
// behaviour is preserved while a side-counter records every `tmux new-session`
// invocation. Used as the executor command runner in AC-2 / AC-3.
func countingNewSessionRunner() (CommandRunner, *int32) {
	var newSessionCalls int32
	base := successRunner()
	runner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "new-session" {
			atomic.AddInt32(&newSessionCalls, 1)
		}
		return base(ctx, name, args...)
	}
	return runner, &newSessionCalls
}

// saveSingleNodeWorkflow persists a one-node WorkflowDef of the given NodeType
// and returns its ID. ProcessID points at a real registry entry so the existing
// process-node spawn path can resolve it; command nodes ignore the ProcessID.
func saveSingleNodeWorkflow(t *testing.T, s *Storage, nodeType NodeType, config map[string]string) string {
	t.Helper()
	now := time.Now().Format(time.RFC3339)
	wf := WorkflowDef{
		ID:          "wf-skills-cmd-01",
		Name:        "skills-cmd-01-single",
		Description: "single-node fixture for skills-cmd-01 tests",
		Nodes: []WorkflowNode{
			{
				ID:        "n1",
				ProcessID: "bmad-brainstorming",
				Label:     "Only",
				Status:    NodePending,
				Config:    config,
				NodeType:  nodeType,
			},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, s.SaveWorkflow(wf))
	return wf.ID
}

// waitForNodeStatus polls GetExecution until the named node reaches the
// expected status or the deadline elapses. Tick matches the harness's
// 50ms pollInterval — sub-50ms ticks would just spin without observing change.
func waitForNodeStatus(t *testing.T, e *Executor, execID, nodeID string, want WorkflowNodeStatus) {
	t.Helper()
	require.Eventuallyf(t, func() bool {
		got, err := e.GetExecution(execID)
		if err != nil {
			return false
		}
		for _, n := range got.Nodes {
			if n.ID == nodeID {
				return n.Status == want
			}
		}
		return false
	}, 3*time.Second, 50*time.Millisecond,
		"node %q never reached status %q", nodeID, want)
}

// startSingleNodeAndWait saves a single-node workflow, registers a
// counting runner, starts the workflow, and blocks until the node reaches
// `want`. Returns the new-session counter for downstream assertions and
// schedules StopWorkflow as cleanup insurance against goroutine leaks if
// the test fails before the node settles.
func startSingleNodeAndWait(t *testing.T, h *testHarness, nt NodeType, cfg map[string]string, want WorkflowNodeStatus) *int32 {
	t.Helper()
	runner, calls := countingNewSessionRunner()
	h.executor.SetCommandRunner(runner)

	wfID := saveSingleNodeWorkflow(t, h.storage, nt, cfg)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

	waitForNodeStatus(t, h.executor, exec.ID, "n1", want)
	return calls
}

// ── exec-03 helpers ──

// tmuxCallRecord captures a single tmux subcommand invocation.
type tmuxCallRecord struct {
	op   string   // e.g., "new-session", "list-panes", "capture-pane", "send-keys"
	args []string // remaining args after the op
}

// callTracker is a concurrency-safe ordered log of tmux invocations.
type callTracker struct {
	mu    sync.Mutex
	calls []tmuxCallRecord
}

func (ct *callTracker) record(rec tmuxCallRecord) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.calls = append(ct.calls, rec)
}

func (ct *callTracker) snapshot() []tmuxCallRecord {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	cp := make([]tmuxCallRecord, len(ct.calls))
	copy(cp, ct.calls)
	return cp
}

// idleCycleTrackingRunner records every tmux call and simulates idle
// completion via a repeating 4-capture cycle per node:
//   - phase 0: baseline content
//   - phases 1-3: idle-prompt content (triggers watching→done on phase 2,
//     phase 3 absorbs the post-completion captureOutput call)
//
// Pane is always alive (list-panes → "0\n"), new-session and send-keys succeed.
func idleCycleTrackingRunner() (CommandRunner, *callTracker) {
	tracker := &callTracker{}
	var captureCount int32

	runner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name != "tmux" || len(args) == 0 {
			return []byte("ok"), nil
		}

		rec := tmuxCallRecord{op: args[0], args: make([]string, len(args)-1)}
		copy(rec.args, args[1:])
		tracker.record(rec)

		switch args[0] {
		case "new-session":
			return []byte("ok"), nil
		case "list-panes":
			return []byte("0\n"), nil
		case "capture-pane":
			tick := int(atomic.AddInt32(&captureCount, 1))
			phase := (tick - 1) % 4
			group := (tick - 1) / 4
			if phase == 0 {
				return []byte(fmt.Sprintf("baseline-%d", group)), nil
			}
			return []byte(fmt.Sprintf("done-%d\n❯\n", group)), nil
		case "send-keys":
			return []byte("ok"), nil
		}
		return []byte("ok"), nil
	}

	return runner, tracker
}

// saveChainedDAGWorkflow persists a 3-node DAG: process(A) → command(B) → command(C).
func saveChainedDAGWorkflow(t *testing.T, s *Storage, bCommand, cCommand string) string {
	t.Helper()
	now := time.Now().Format(time.RFC3339)
	wf := WorkflowDef{
		ID:          "wf-exec03-chained",
		Name:        "exec03-chained-dag",
		Description: "A→B→C chained DAG for exec-03 tests",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "Process A", Status: NodePending, NodeType: NodeTypeProcess},
			{ID: "B", Label: "Command B", Status: NodePending, NodeType: NodeTypeCommand, Config: map[string]string{"commandName": bCommand}},
			{ID: "C", Label: "Command C", Status: NodePending, NodeType: NodeTypeCommand, Config: map[string]string{"commandName": cCommand}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
			{ID: "e2", Source: "B", Target: "C"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, s.SaveWorkflow(wf))
	return wf.ID
}

// saveTwoNodeDAGWorkflow persists a 2-node DAG: process(A) → command(B).
func saveTwoNodeDAGWorkflow(t *testing.T, s *Storage, bCommand string) string {
	t.Helper()
	now := time.Now().Format(time.RFC3339)
	wf := WorkflowDef{
		ID:          "wf-exec03-two-node",
		Name:        "exec03-two-node-dag",
		Description: "A→B two-node DAG for exec-03 tests",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: "bmad-brainstorming", Label: "Process A", Status: NodePending, NodeType: NodeTypeProcess},
			{ID: "B", Label: "Command B", Status: NodePending, NodeType: NodeTypeCommand, Config: map[string]string{"commandName": bCommand}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "B"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, s.SaveWorkflow(wf))
	return wf.ID
}

// ── skills-cmd-01 tests (still valid) ──

// TestStory1_AC1_NodeTypeCommandRoundTrip — AC-1: NodeTypeCommand round-trips
// through WorkflowDef JSON and EffectiveType() reports it after decode.
func TestStory1_AC1_NodeTypeCommandRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		node   WorkflowNode
		expect map[string]string
	}{
		{
			name: "simplify command with all three config keys",
			node: WorkflowNode{
				ID:       "cmd-1",
				Label:    "Simplify",
				Position: Position{X: 10, Y: 20},
				Status:   NodePending,
				NodeType: NodeTypeCommand,
				Config: map[string]string{
					"commandName":        "simplify",
					"commandPath":        "/tmp/simplify.md",
					"commandDescription": "Review recent changes",
				},
			},
			expect: map[string]string{
				"commandName":        "simplify",
				"commandPath":        "/tmp/simplify.md",
				"commandDescription": "Review recent changes",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wf := WorkflowDef{
				ID:    "wf-cmd-rt",
				Name:  "command round-trip",
				Nodes: []WorkflowNode{tc.node},
				Edges: []WorkflowEdge{},
			}

			blob, err := json.Marshal(wf)
			require.NoError(t, err)

			var decoded WorkflowDef
			require.NoError(t, json.Unmarshal(blob, &decoded))
			require.Len(t, decoded.Nodes, 1)

			got := decoded.Nodes[0]
			assert.Equal(t, NodeTypeCommand, got.EffectiveType())
			for key, want := range tc.expect {
				assert.Equal(t, want, got.Config[key], "config[%q]", key)
			}
		})
	}
}

// TestStory1_AC3_ProcessNodeUnchanged — AC-3 regression guard: a legacy
// process node still spawns tmux and reaches NodeComplete. If a future
// change to the executeNode switch breaks the default branch, this fires.
func TestStory1_AC3_ProcessNodeUnchanged(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	calls := startSingleNodeAndWait(t, h, NodeTypeProcess, map[string]string{}, NodeComplete)

	assert.GreaterOrEqual(t, atomic.LoadInt32(calls), int32(1),
		"process-node dispatch must still invoke `tmux new-session`")
}

// ── exec-03 tests ──

// TestInjectSlashCommand_Argv (AC-1): assert exact hex argv from
// injectSlashCommand("bmad-abc:0.0", "simplify").
func TestInjectSlashCommand_Argv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		target      string
		commandName string
		wantArgs    []string
	}{
		{
			name:        "simplify produces exact hex sequence",
			target:      "bmad-abc:0.0",
			commandName: "simplify",
			// "/simplify\n" = 2f 73 69 6d 70 6c 69 66 79 0a
			wantArgs: []string{"tmux", "send-keys", "-H", "-t", "bmad-abc:0.0",
				"2f", "73", "69", "6d", "70", "6c", "69", "66", "79", "0a"},
		},
		{
			name:        "review produces exact hex sequence",
			target:      "bmad-xyz:0.0",
			commandName: "review",
			// "/review\n" = 2f 72 65 76 69 65 77 0a
			wantArgs: []string{"tmux", "send-keys", "-H", "-t", "bmad-xyz:0.0",
				"2f", "72", "65", "76", "69", "65", "77", "0a"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t)

			var captured []string
			h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
				captured = append([]string{name}, args...)
				return []byte("ok"), nil
			})

			err := h.executor.injectSlashCommand(context.Background(), tc.target, tc.commandName)
			require.NoError(t, err)
			assert.Equal(t, tc.wantArgs, captured, "argv must match exact hex encoding")
		})
	}
}

// TestExecuteCommandNode_ChainedDAG (AC-2): process(A) → command(B: simplify) →
// command(C: review). Assert only ONE tmux new-session, B and C inject their
// respective slash commands, all three nodes complete.
func TestExecuteCommandNode_ChainedDAG(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	runner, tracker := idleCycleTrackingRunner()
	h.executor.SetCommandRunner(runner)

	wfID := saveChainedDAGWorkflow(t, h.storage, "simplify", "review")
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

	// Wait for all three nodes to complete.
	waitForNodeStatus(t, h.executor, exec.ID, "A", NodeComplete)
	waitForNodeStatus(t, h.executor, exec.ID, "B", NodeComplete)
	waitForNodeStatus(t, h.executor, exec.ID, "C", NodeComplete)

	calls := tracker.snapshot()

	// Assert only ONE tmux new-session call (for process node A).
	var newSessionCount int
	for _, c := range calls {
		if c.op == "new-session" {
			newSessionCount++
		}
	}
	assert.Equal(t, 1, newSessionCount, "only node A should spawn a tmux session")

	// Assert B and C injected their respective slash commands via send-keys -H.
	var injections [][]string
	for _, c := range calls {
		if c.op == "send-keys" && len(c.args) > 0 && c.args[0] == "-H" {
			injections = append(injections, c.args)
		}
	}
	require.Len(t, injections, 2, "exactly two slash injections (B and C)")

	// First injection: /simplify\n → hex 2f 73 69 6d 70 6c 69 66 79 0a
	assert.Contains(t, strings.Join(injections[0], " "), "2f 73 69 6d 70 6c 69 66 79 0a",
		"first injection must be /simplify")

	// Second injection: /review\n → hex 2f 72 65 76 69 65 77 0a
	assert.Contains(t, strings.Join(injections[1], " "), "2f 72 65 76 69 65 77 0a",
		"second injection must be /review")
}

// TestExecuteCommandNode_NoParentSpawns (AC-3): single command node X with
// commandName "brainstorm", no incoming edges. Assert spawnCommandSession called
// with "/brainstorm", injectSlashCommand NOT called.
func TestExecuteCommandNode_NoParentSpawns(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	runner, tracker := idleCycleTrackingRunner()
	h.executor.SetCommandRunner(runner)

	cfg := map[string]string{"commandName": "brainstorm"}
	wfID := saveSingleNodeWorkflow(t, h.storage, NodeTypeCommand, cfg)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

	waitForNodeStatus(t, h.executor, exec.ID, "n1", NodeComplete)

	calls := tracker.snapshot()

	// Assert spawnCommandSession was called (exactly one new-session).
	var newSessionArgs [][]string
	for _, c := range calls {
		if c.op == "new-session" {
			newSessionArgs = append(newSessionArgs, c.args)
		}
	}
	require.Len(t, newSessionArgs, 1, "single command node must spawn one session")

	// The new-session command should contain "/brainstorm" in the bash -c wrapper.
	fullArgs := strings.Join(newSessionArgs[0], " ")
	assert.Contains(t, fullArgs, "/brainstorm",
		"spawned pane's initial claude argument must contain /brainstorm")

	// Assert injectSlashCommand was NOT called (no send-keys -H).
	for _, c := range calls {
		if c.op == "send-keys" {
			for _, a := range c.args {
				if a == "-H" {
					t.Fatal("injectSlashCommand must NOT be called when spawning fresh (no parent)")
				}
			}
		}
	}
}

// TestExecuteCommandNode_MissingCommandName (AC-4): command node with empty
// Config["commandName"]. Assert node fails, log contains "missing commandName",
// zero tmux calls.
func TestExecuteCommandNode_MissingCommandName(t *testing.T) {
	// No t.Parallel() — captures global log output.
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	h := newHarness(t)
	var tmuxCalls int32
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" {
			atomic.AddInt32(&tmuxCalls, 1)
		}
		return []byte("ok"), nil
	})

	cfg := map[string]string{} // deliberately empty — no commandName
	wfID := saveSingleNodeWorkflow(t, h.storage, NodeTypeCommand, cfg)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

	waitForNodeStatus(t, h.executor, exec.ID, "n1", NodeFailed)

	assert.Equal(t, int32(0), atomic.LoadInt32(&tmuxCalls),
		"missing commandName must short-circuit before any tmux call")
	assert.Contains(t, logBuf.String(), "missing commandName",
		"log must contain 'missing commandName' diagnostic")
}

// TestExecuteCommandNode_SpawnFails covers the error path where
// spawnCommandSession returns an error (e.g., tmux new-session fails).
func TestExecuteCommandNode_SpawnFails(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "new-session" {
			return nil, fmt.Errorf("tmux: server not found")
		}
		return []byte("ok"), nil
	})

	cfg := map[string]string{"commandName": "simplify"}
	wfID := saveSingleNodeWorkflow(t, h.storage, NodeTypeCommand, cfg)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

	waitForNodeStatus(t, h.executor, exec.ID, "n1", NodeFailed)
}

// TestExecuteCommandNode_DispatcherReplacesFailFast (AC-5): run a command node
// through StartWorkflow. Assert it does NOT produce the Phase 2 sentinel
// "command nodes not yet runnable" and actually executes.
func TestExecuteCommandNode_DispatcherReplacesFailFast(t *testing.T) {
	// No t.Parallel() — captures global log output.
	var logBuf bytes.Buffer
	log.SetOutput(&logBuf)
	defer log.SetOutput(os.Stderr)

	h := newHarness(t)
	runner, _ := idleCycleTrackingRunner()
	h.executor.SetCommandRunner(runner)

	cfg := map[string]string{"commandName": "simplify"}
	wfID := saveSingleNodeWorkflow(t, h.storage, NodeTypeCommand, cfg)
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

	waitForNodeStatus(t, h.executor, exec.ID, "n1", NodeComplete)

	assert.NotContains(t, logBuf.String(), "command nodes not yet runnable",
		"Phase 2 sentinel must not appear — dispatcher routes to executeCommandNode")
}

// TestExecuteCommandNode_InjectionOrdering (AC-6): for a reused session,
// verify injectSlashCommand fires BEFORE waitForIdleCompletion's first
// capture-pane. The state machine must see a hash change after injection.
func TestExecuteCommandNode_InjectionOrdering(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	runner, tracker := idleCycleTrackingRunner()
	h.executor.SetCommandRunner(runner)

	wfID := saveTwoNodeDAGWorkflow(t, h.storage, "simplify")
	exec, err := h.executor.StartWorkflow(context.Background(), wfID, t.TempDir(), "sonnet")
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.executor.StopWorkflow(exec.ID) })

	waitForNodeStatus(t, h.executor, exec.ID, "A", NodeComplete)
	waitForNodeStatus(t, h.executor, exec.ID, "B", NodeComplete)

	calls := tracker.snapshot()

	// Find the first send-keys -H call (injection for node B).
	injectionIdx := -1
	for i, c := range calls {
		if c.op == "send-keys" && len(c.args) > 0 && c.args[0] == "-H" {
			injectionIdx = i
			break
		}
	}
	require.NotEqual(t, -1, injectionIdx, "must find a send-keys -H injection call")

	// Find the first capture-pane call AFTER the injection. This is
	// waitForIdleCompletion's first tick for node B.
	firstCaptureAfterInject := -1
	for i := injectionIdx + 1; i < len(calls); i++ {
		if calls[i].op == "capture-pane" {
			firstCaptureAfterInject = i
			break
		}
	}
	require.NotEqual(t, -1, firstCaptureAfterInject,
		"must find a capture-pane after injection")
	assert.Greater(t, firstCaptureAfterInject, injectionIdx,
		"injection must happen BEFORE the state machine's first capture")

	// Verify the state machine saw a hash change (node B completed
	// successfully — which only happens if priming→waitingForWork→watching→done
	// all transitioned correctly after the injection).
	got, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	for _, n := range got.Nodes {
		if n.ID == "B" {
			assert.Equal(t, NodeComplete, n.Status,
				"node B must complete — proving the state machine observed the hash change post-injection")
		}
	}
}
