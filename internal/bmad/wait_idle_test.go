package bmad

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newWaitIdleState creates a minimal execState suitable for
// waitForIdleCompletion tests. It provides the maps that pollForIdle
// and pollForQuestionFromCapture need.
func newWaitIdleState(execID, nodeID string) *execState {
	return &execState{
		exec: &WorkflowExecution{
			ID:          execID,
			Status:      ExecRunning,
			Nodes:       []WorkflowNode{{ID: nodeID, Label: nodeID, Status: NodeRunning}},
			NodeOutputs: map[string]string{},
		},
		cancel:           func() {},
		inDegree:         map[string]int{nodeID: 0},
		outEdges:         map[string][]WorkflowEdge{},
		lastQuestionHash: map[string]string{},
		lastOutputHash:   map[string]string{},
		idleEmitted:      map[string]bool{},
	}
}

// TestWaitForIdleCompletion_HappyPath verifies AC-1: the state machine
// transitions through priming → waiting-for-work → watching-for-idle
// and returns nil when the hash stabilises with an idle prompt.
func TestWaitForIdleCompletion_HappyPath(t *testing.T) {
	h := newHarness(t)

	// Hash sequence: H1, H1, H2, H3, H3+idle
	// Tick 1 (priming): captures "output-A" → baseline
	// Tick 2 (waiting): captures "output-A" → same → still waiting
	// Tick 3 (waiting): captures "output-B" → changed → watching
	// Tick 4 (watching): captures "output-C" → changed → update stable
	// Tick 5 (watching): captures idle("output-C") → stable + idle → done
	captureSeq := []string{
		"output-A",
		"output-A",
		"output-B",
		"output-C",
		makeIdleOutput("output-C"),
	}
	runner, captureCount := idleMockRunner(captureSeq, 0)
	h.executor.SetCommandRunner(runner)

	state := newWaitIdleState("exec-happy", "n1")

	err := h.executor.waitForIdleCompletion(context.Background(), state, "n1", "test:0.0", defaultProcessNodeTimeout)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, int(atomic.LoadInt32(captureCount)), 5)
}

// TestWaitForIdleCompletion_NoWork_Timeout verifies AC-2: when the pane hash
// never changes from the baseline, the method returns ErrIdleTimeoutNoStart.
func TestWaitForIdleCompletion_NoWork_Timeout(t *testing.T) {
	h := newHarness(t)

	// Hash never changes — always "static-output".
	captureSeq := []string{"static-output"}
	runner, _ := idleMockRunner(captureSeq, 0)
	h.executor.SetCommandRunner(runner)

	state := newWaitIdleState("exec-timeout", "n1")

	err := h.executor.waitForIdleCompletion(context.Background(), state, "n1", "test:0.0", 150*time.Millisecond)
	assert.ErrorIs(t, err, ErrIdleTimeoutNoStart)
}

// TestWaitForIdleCompletion_PaneDeath verifies AC-3: pane death at any stage
// returns nil (legacy completion).
func TestWaitForIdleCompletion_PaneDeath(t *testing.T) {
	tests := []struct {
		name           string
		captureSeq     []string
		paneDeadAtTick int
		description    string
	}{
		{
			name:           "during_waiting_for_work",
			captureSeq:     []string{"output-A", "output-A"},
			paneDeadAtTick: 3,
			description:    "pane dies on tick 3 while hash is still baseline",
		},
		{
			name: "during_watching_for_idle",
			// Tick 1: priming, tick 2: waiting (hash changes), tick 3-4: watching, tick 5: dead
			captureSeq:     []string{"output-A", "output-B", "output-C", "output-D"},
			paneDeadAtTick: 5,
			description:    "pane dies on tick 5 during watching-for-idle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)
			runner, _ := idleMockRunner(tt.captureSeq, tt.paneDeadAtTick)
			h.executor.SetCommandRunner(runner)

			state := newWaitIdleState("exec-dead-"+tt.name, "n1")

			err := h.executor.waitForIdleCompletion(context.Background(), state, "n1", "test:0.0", defaultProcessNodeTimeout)
			assert.NoError(t, err, tt.description)
		})
	}
}

// TestWaitForIdleCompletion_CtxCancel verifies AC-4: context cancellation
// during any stage returns ctx.Err().
func TestWaitForIdleCompletion_CtxCancel(t *testing.T) {
	tests := []struct {
		name       string
		captureSeq []string
		cancelAt   int // cancel after this many capture-pane calls
	}{
		{
			name:       "during_priming",
			captureSeq: []string{"output-A", "output-A"},
			cancelAt:   0, // cancel before first capture
		},
		{
			name:       "during_waiting_for_work",
			captureSeq: []string{"output-A", "output-A", "output-A"},
			cancelAt:   2,
		},
		{
			name:       "during_watching_for_idle",
			captureSeq: []string{"output-A", "output-B", "output-C", "output-C"},
			cancelAt:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHarness(t)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var captureCount int32
			var cancelled atomic.Bool

			runner := CommandRunner(func(innerCtx context.Context, name string, args ...string) ([]byte, error) {
				if name == "tmux" && len(args) > 0 {
					switch args[0] {
					case "list-panes":
						return []byte("0\n"), nil
					case "capture-pane":
						idx := int(atomic.AddInt32(&captureCount, 1)) - 1
						if !cancelled.Load() && idx >= tt.cancelAt {
							cancel()
							cancelled.Store(true)
						}
						if idx < len(tt.captureSeq) {
							return []byte(tt.captureSeq[idx]), nil
						}
						return []byte(tt.captureSeq[len(tt.captureSeq)-1]), nil
					}
				}
				return []byte("ok"), nil
			})
			h.executor.SetCommandRunner(runner)

			state := newWaitIdleState("exec-cancel-"+tt.name, "n1")

			err := h.executor.waitForIdleCompletion(ctx, state, "n1", "test:0.0", defaultProcessNodeTimeout)
			assert.Error(t, err)
			assert.Equal(t, context.Canceled, err)
		})
	}
}

// TestWaitForIdleCompletion_PaneDeathSessionGone verifies that when tmux
// list-panes returns an error (session completely gone), the method returns nil.
func TestWaitForIdleCompletion_PaneDeathSessionGone(t *testing.T) {
	h := newHarness(t)

	var listPaneCount int32
	runner := CommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "list-panes" {
			tick := int(atomic.AddInt32(&listPaneCount, 1))
			if tick >= 2 {
				return nil, fmt.Errorf("tmux: server not running")
			}
			return []byte("0\n"), nil
		}
		if name == "tmux" && len(args) > 0 && args[0] == "capture-pane" {
			return []byte("some output"), nil
		}
		return []byte("ok"), nil
	})
	h.executor.SetCommandRunner(runner)

	state := newWaitIdleState("exec-gone", "n1")

	err := h.executor.waitForIdleCompletion(context.Background(), state, "n1", "test:0.0", defaultProcessNodeTimeout)
	assert.NoError(t, err)
}

// TestWaitForIdleCompletion_IdlePromptRequiresStableHash verifies that
// detectIdlePrompt alone is not sufficient — the hash must also be stable
// (same across two consecutive polls in the watching-for-idle stage).
func TestWaitForIdleCompletion_IdlePromptRequiresStableHash(t *testing.T) {
	h := newHarness(t)

	// The hash changes on every tick, but idle prompt is present.
	// This should NOT return nil from idle detection — hash is never stable.
	// It should eventually hit pane death.
	seq := []string{
		"output-A",              // priming
		"output-B",              // waiting → watching (hash changed)
		makeIdleOutput("out-C"), // watching: idle prompt but hash changed (C != B)
		makeIdleOutput("out-D"), // watching: idle prompt but hash changed (D != C)
		makeIdleOutput("out-E"), // watching: idle prompt but hash changed (E != D)
	}
	runner, _ := idleMockRunner(seq, 6) // pane dies on tick 6
	h.executor.SetCommandRunner(runner)

	state := newWaitIdleState("exec-unstable", "n1")

	err := h.executor.waitForIdleCompletion(context.Background(), state, "n1", "test:0.0", defaultProcessNodeTimeout)
	// Should complete via pane death, not idle detection.
	assert.NoError(t, err)
}

