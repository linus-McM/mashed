package bmad

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── killWorkflowChainTails Tests (exec-04) ──
//
// These tests reuse killSessionTargets from cleanup_test.go and the standard
// testHarness / cmdCall / sampleWorkflow helpers from executor_test.go.

// TestKillWorkflowChainTails_Dedup verifies AC-1: 3 nodes with two distinct
// bare session names produce exactly 2 kill-session calls.
func TestKillWorkflowChainTails_Dedup(t *testing.T) {
	h := newHarness(t)

	var calls []cmdCall
	var mu sync.Mutex
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		mu.Lock()
		calls = append(calls, cmdCall{name: name, args: append([]string(nil), args...)})
		mu.Unlock()
		return nil, nil
	})

	state := &execState{
		exec: &WorkflowExecution{
			Nodes: []WorkflowNode{
				{ID: "n1", TmuxTarget: "bmad-abc:0.0"},
				{ID: "n2", TmuxTarget: "bmad-abc:0.0"},
				{ID: "n3", TmuxTarget: "bmad-def:0.0"},
			},
		},
	}

	h.executor.killWorkflowChainTails(state)

	mu.Lock()
	targets := killSessionTargets(calls)
	mu.Unlock()

	assert.Len(t, targets, 2, "expected exactly 2 kill-session calls (deduped)")
	assert.ElementsMatch(t, []string{"bmad-abc", "bmad-def"}, targets)
}

// TestKillWorkflowChainTails_OnComplete verifies AC-2: kill-session fires
// when a workflow reaches ExecComplete.
func TestKillWorkflowChainTails_OnComplete(t *testing.T) {
	h := newHarness(t)

	wf := sampleWorkflow("cleanup-complete")
	require.NoError(t, h.storage.SaveWorkflow(wf))

	var calls []cmdCall
	var mu sync.Mutex
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		mu.Lock()
		calls = append(calls, cmdCall{name: name, args: append([]string(nil), args...)})
		mu.Unlock()

		if len(args) > 0 && args[0] == "list-panes" {
			return []byte("1\n"), nil
		}
		return []byte("ok"), nil
	})

	exec, err := h.executor.StartWorkflow(context.Background(), wf.ID, t.TempDir(), "")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		e, _ := h.executor.GetExecution(exec.ID)
		return e != nil && e.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	mu.Lock()
	targets := killSessionTargets(calls)
	mu.Unlock()

	assert.NotEmpty(t, targets, "kill-session must be called on ExecComplete")
}

// TestKillWorkflowChainTails_OnFailed verifies AC-3: kill-session fires when a
// workflow reaches ExecFailed, and the failure status is preserved.
func TestKillWorkflowChainTails_OnFailed(t *testing.T) {
	h := newHarness(t)

	wf := sampleWorkflow("cleanup-fail")
	require.NoError(t, h.storage.SaveWorkflow(wf))

	// n1 succeeds (new-session OK, list-panes returns dead immediately).
	// n2 fails (new-session errors). This ensures n1 has a TmuxTarget set
	// and the finalizer has something to kill on ExecFailed.
	var newSessionCount int32
	var calls []cmdCall
	var mu sync.Mutex
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		mu.Lock()
		calls = append(calls, cmdCall{name: name, args: append([]string(nil), args...)})
		mu.Unlock()

		if len(args) > 0 && args[0] == "new-session" {
			n := atomic.AddInt32(&newSessionCount, 1)
			if n >= 2 {
				return nil, assert.AnError
			}
			return []byte("ok"), nil
		}
		if len(args) > 0 && args[0] == "list-panes" {
			return []byte("1\n"), nil
		}
		return []byte("ok"), nil
	})

	exec, err := h.executor.StartWorkflow(context.Background(), wf.ID, t.TempDir(), "")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		e, _ := h.executor.GetExecution(exec.ID)
		return e != nil && e.Status == ExecPaused
	}, 5*time.Second, 50*time.Millisecond)

	require.NoError(t, h.executor.StopWorkflow(exec.ID))

	// Wait for the runner goroutine's defer to fire kill-session.
	require.Eventually(t, func() bool {
		mu.Lock()
		targets := killSessionTargets(calls)
		mu.Unlock()
		return len(targets) > 0
	}, 5*time.Second, 50*time.Millisecond)

	e, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecFailed, e.Status, "status must remain ExecFailed after cleanup")
}

// TestKillWorkflowChainTails_NoTargets verifies AC-4: when no nodes have a
// TmuxTarget, zero kill-session calls are made.
func TestKillWorkflowChainTails_NoTargets(t *testing.T) {
	h := newHarness(t)

	var calls []cmdCall
	var mu sync.Mutex
	h.executor.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		mu.Lock()
		calls = append(calls, cmdCall{name: name, args: append([]string(nil), args...)})
		mu.Unlock()
		return nil, nil
	})

	state := &execState{
		exec: &WorkflowExecution{
			Nodes: []WorkflowNode{
				{ID: "n1", TmuxTarget: ""},
				{ID: "n2", TmuxTarget: ""},
			},
		},
	}

	h.executor.killWorkflowChainTails(state)

	mu.Lock()
	targets := killSessionTargets(calls)
	mu.Unlock()

	assert.Empty(t, targets, "no kill-session calls expected when no targets set")
}
