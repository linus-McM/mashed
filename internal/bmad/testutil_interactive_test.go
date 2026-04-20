package bmad

// Interactive smoke-test harness used by executor_interactive_smoke_test.go
// (story bmad-interactive-07). The harness wraps newHarness with helpers that
// drive a single interactive node end-to-end via a mock CommandRunner.
//
// Production processes spawn tmux + claude; this harness substitutes
// roundCaptureRunner so tests exercise executeInteractiveNode without any
// real subprocess. Event capture uses the same testEventHook atomic.Value
// pattern as executor_iteration_test.go.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// saveInteractiveWorkflowWithOverrides persists a single-node workflow where
// the node's InputSpecs field optionally overrides the registry-derived specs.
// When overrides is nil, behaves like saveInteractiveWorkflow.
func saveInteractiveWorkflowWithOverrides(t *testing.T, s *Storage, processID string, overrides []InputSpec) string {
	t.Helper()
	now := time.Now().Format(time.RFC3339)
	wf := WorkflowDef{
		ID:          fmt.Sprintf("wf-smoke-%s", processID),
		Name:        "smoke-interactive-fixture",
		Description: "single-node smoke workflow",
		Nodes: []WorkflowNode{
			{
				ID:         "n1",
				ProcessID:  processID,
				Label:      "Interactive Node",
				Status:     NodePending,
				NodeType:   NodeTypeProcess,
				InputSpecs: overrides,
			},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	require.NoError(t, s.SaveWorkflow(wf))
	return wf.ID
}

// interactiveHarness bundles executor + event capture for a single
// single-node interactive workflow. Callers drive the node via respond() and
// assert via expectAwaitingInput / waitForRoundComplete / assertNodeComplete.
type interactiveHarness struct {
	t        *testing.T
	h        *testHarness
	execID   string
	nodeID   string
	repoPath string
	getSnap  func() []struct {
		name    string
		payload interface{}
	}
}

// newInteractiveHarness installs the event hook and returns a harness with a
// fresh testHarness. Callers invoke startSingleNode next.
func newInteractiveHarness(t *testing.T) *interactiveHarness {
	t.Helper()
	h := newHarness(t)
	return &interactiveHarness{
		t:       t,
		h:       h,
		getSnap: hookEvents(t),
	}
}

// startSingleNode persists a one-node workflow using procID, wires the mock
// runner, starts execution, and returns (execID, nodeID). The default nodeID
// is "n1".
func (h *interactiveHarness) startSingleNode(procID string) (string, string) {
	h.t.Helper()
	return h.startSingleNodeWithOverrides(procID, nil)
}

// startSingleNodeWithOverrides is like startSingleNode but lets the caller
// override the node's InputSpecs (used by advanced-elicitation where the
// target-content upstream dependency is stubbed to non-required for the smoke
// test — no real upstream node exists in the single-node workflow).
func (h *interactiveHarness) startSingleNodeWithOverrides(procID string, inputSpecOverrides []InputSpec) (string, string) {
	h.t.Helper()
	h.repoPath = h.t.TempDir()

	h.h.executor.SetCommandRunner(h.makeRunner())

	wfID := saveInteractiveWorkflowWithOverrides(h.t, h.h.storage, procID, inputSpecOverrides)
	exec, err := h.h.executor.StartWorkflow(context.Background(), wfID, h.repoPath, "sonnet")
	require.NoError(h.t, err)
	h.execID = exec.ID
	h.nodeID = "n1"
	h.t.Cleanup(func() { _ = h.h.executor.StopWorkflow(h.execID) })
	return h.execID, h.nodeID
}

// makeRunner returns a CommandRunner identical in spirit to roundCaptureRunner:
// every 3rd list-panes call reports pane-dead so waitForIdleCompletion exits,
// and capture-pane returns an idle prompt so detectIdlePrompt succeeds. tmux
// new-session and send-keys are ack'd with "ok" — send-keys output is not
// inspected by the smoke tests.
func (h *interactiveHarness) makeRunner() CommandRunner {
	const callsPerRound = 3
	var listPaneCount int
	var mu sync.Mutex

	return func(_ context.Context, name string, args ...string) ([]byte, error) {
		if name != "tmux" || len(args) == 0 {
			return []byte("ok"), nil
		}
		switch args[0] {
		case "list-panes":
			mu.Lock()
			listPaneCount++
			count := listPaneCount
			mu.Unlock()
			if count%callsPerRound == 0 {
				return []byte("1\n"), nil
			}
			return []byte("0\n"), nil
		case "capture-pane":
			return []byte("Round output.\n❯\n"), nil
		}
		return []byte("ok"), nil
	}
}

// respond records a user answer for a suspended node input. Fails the test
// on error — callers must ensure the node is in NodeAwaitingInput first.
func (h *interactiveHarness) respond(execID, nodeID, inputID, value string) {
	h.t.Helper()
	require.NoError(h.t, h.h.executor.RespondToInput(execID, nodeID, inputID, value),
		"respond(%s, %s, %s, %q) failed", execID, nodeID, inputID, value)
}

// expectAwaitingInput polls up to 2s for a PendingPrompt matching (nodeID,
// inputID). Fails if the prompt never appears.
func (h *interactiveHarness) expectAwaitingInput(nodeID, inputID string) {
	h.t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		state, err := h.h.executor.getState(h.execID)
		if err == nil {
			state.mu.Lock()
			for _, p := range state.exec.PendingPrompts {
				if p.NodeID == nodeID && p.InputID == inputID {
					state.mu.Unlock()
					return
				}
			}
			state.mu.Unlock()
		}
		time.Sleep(15 * time.Millisecond)
	}
	h.t.Fatalf("expectAwaitingInput: timed out waiting for pending prompt (%s/%s)", nodeID, inputID)
}

// waitForRoundComplete blocks until at least `round` round_complete events
// have fired for nodeID, or the 3s deadline elapses.
func (h *interactiveHarness) waitForRoundComplete(nodeID string, round int) {
	h.t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if h.countRoundComplete(nodeID) >= round {
			return
		}
		time.Sleep(15 * time.Millisecond)
	}
	h.t.Fatalf("waitForRoundComplete: timed out waiting for round %d on %s (saw %d)",
		round, nodeID, h.countRoundComplete(nodeID))
}

// countRoundComplete counts round_complete events targeting nodeID.
func (h *interactiveHarness) countRoundComplete(nodeID string) int {
	snap := h.getSnap()
	count := 0
	for _, ev := range snap {
		if ev.name != EventRoundComplete {
			continue
		}
		if m, ok := ev.payload.(map[string]any); ok {
			if m["nodeId"] == nodeID {
				count++
			}
		}
	}
	return count
}

// assertGateSatisfied asserts a gate_satisfied event fired for nodeID with the
// given round. If reasonContains is non-empty, the event's reason must contain
// that substring.
func (h *interactiveHarness) assertGateSatisfied(nodeID string, round int, reasonContains string) {
	h.t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if h.matchGateSatisfied(nodeID, round, reasonContains) {
			return
		}
		time.Sleep(15 * time.Millisecond)
	}
	h.t.Fatalf("assertGateSatisfied: no gate_satisfied event for node=%s round=%d reason~=%q",
		nodeID, round, reasonContains)
}

func (h *interactiveHarness) matchGateSatisfied(nodeID string, round int, reasonContains string) bool {
	for _, ev := range h.getSnap() {
		if ev.name != EventGateSatisfied {
			continue
		}
		m, ok := ev.payload.(map[string]any)
		if !ok {
			continue
		}
		if m["nodeId"] != nodeID {
			continue
		}
		if r, rok := m["round"].(int); rok && r != round {
			continue
		}
		if reasonContains != "" {
			reason, _ := m["reason"].(string)
			if !strings.Contains(reason, reasonContains) {
				continue
			}
		}
		return true
	}
	return false
}

// assertNodeComplete blocks until nodeID reaches NodeComplete or fails the
// test after 5s.
func (h *interactiveHarness) assertNodeComplete(nodeID string) {
	h.t.Helper()
	status := pollForTerminal(h.h.executor, h.execID, nodeID, 5*time.Second)
	require.Equal(h.t, NodeComplete, status,
		"assertNodeComplete: node %s did not complete (got %s)", nodeID, status)
}

// assertNodeFailed blocks until nodeID reaches NodeFailed or fails the test
// after 5s.
func (h *interactiveHarness) assertNodeFailed(nodeID string) {
	h.t.Helper()
	status := pollForTerminal(h.h.executor, h.execID, nodeID, 5*time.Second)
	require.Equal(h.t, NodeFailed, status,
		"assertNodeFailed: node %s did not fail (got %s)", nodeID, status)
}

// writeArtifact creates the _bmad-output file for an artifact name so
// verifyOutputs succeeds. Uses the canonical artifactPaths map.
func (h *interactiveHarness) writeArtifact(artifactName, body string) {
	h.t.Helper()
	path := ResolveArtifactPath(artifactName, h.repoPath)
	require.NotEmpty(h.t, path, "writeArtifact: artifact %q is unmapped", artifactName)
	require.NoError(h.t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(h.t, os.WriteFile(path, []byte(body), 0o644))
}

// abortedEventsFor returns aborted-event payloads for nodeID.
func (h *interactiveHarness) abortedEventsFor(nodeID string) []map[string]any {
	var out []map[string]any
	for _, ev := range h.getSnap() {
		if ev.name != EventAborted {
			continue
		}
		if m, ok := ev.payload.(map[string]any); ok && m["nodeId"] == nodeID {
			out = append(out, m)
		}
	}
	return out
}

