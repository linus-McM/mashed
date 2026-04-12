package bmad

// Canonical Mock Helpers for BMAD Executor Tests
//
// This file consolidates the four mock patterns that map to the executor's
// state machine stages. Every executor test should use one of these helpers
// (or a wrapper around them) rather than building ad-hoc CommandRunner
// fixtures with hardcoded tmux responses.
//
// Pattern → Helper mapping:
//
//   SimulatePaneDead       → successRunner()
//   SimulatePaneHashChange → idleMockRunner(varyingSeq, 0)
//   SimulatePaneHashStable → idleMockRunner(stableSeq, 0) with makeIdleOutput
//   SimulateIdleTimeout    → idleMockRunner([]string{"constant"}, 0)
//
// See each function's godoc for detailed usage.

import (
	"context"
	"fmt"
	"sync/atomic"
)

// successRunner returns a CommandRunner that simulates immediate pane death.
//
// State-machine mapping: SimulatePaneDead.
// list-panes returns "1" (pane_dead) on every poll, so waitForIdleCompletion
// short-circuits before any capture-pane call. All other tmux commands
// (new-session, send-keys, etc.) return "ok".
//
// This is the single source of truth for the pane-death mock pattern.
// Callers that need pane-death with additional tracking should wrap this
// helper (see countingNewSessionRunner).
func successRunner() CommandRunner {
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "list-panes" {
			return []byte("1\n"), nil
		}
		return []byte("ok"), nil
	}
}

// failRunner returns a CommandRunner that simulates tmux new-session failure.
//
// The first new-session call returns an error; all other commands (including
// list-panes) return "1" (pane_dead) to prevent hangs if the test somehow
// reaches the poll loop.
func failRunner() CommandRunner {
	return func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if len(args) > 0 && args[0] == "new-session" {
			return nil, fmt.Errorf("tmux failed")
		}
		return []byte("1\n"), nil
	}
}

// idleMockRunner builds a CommandRunner that simulates a sequence of pane
// captures and pane-death transitions for waitForIdleCompletion tests.
//
// State-machine mapping (depends on captureSeq content and paneDeadAtTick):
//
//   SimulatePaneHashChange: pass captureSeq with varying entries, paneDeadAtTick=0.
//     The hash changes on every tick, keeping the state machine in
//     stageWaitingForWork until it transitions to watching.
//
//   SimulatePaneHashStable: pass captureSeq where the last two entries share
//     the same hash and the final entry contains an idle prompt (use
//     makeIdleOutput). The state machine transitions watching → done.
//
//   SimulateIdleTimeout: pass a single constant entry WITHOUT idle prompt,
//     paneDeadAtTick=0. The hash never changes from baseline and no idle
//     prompt appears, so waitForIdleCompletion returns ErrIdleTimeoutNoStart.
//
//   SimulatePaneDead (mid-idle): set paneDeadAtTick > 0 to have list-panes
//     report dead at that tick, triggering legacy pane-death completion
//     mid-state-machine.
//
// Parameters:
//   - captureSeq: successive capture-pane outputs (last entry is cycled if
//     ticks exceed length).
//   - paneDeadAtTick: tick number (1-indexed) at which list-panes reports dead.
//     0 means pane never dies.
//
// Returns the runner and a pointer to the capture-pane call count.
func idleMockRunner(captureSeq []string, paneDeadAtTick int) (CommandRunner, *int32) {
	var captureCount int32
	var listPaneCount int32
	runner := CommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 {
			switch args[0] {
			case "list-panes":
				tick := int(atomic.AddInt32(&listPaneCount, 1))
				if paneDeadAtTick > 0 && tick >= paneDeadAtTick {
					return []byte("1\n"), nil
				}
				return []byte("0\n"), nil
			case "capture-pane":
				idx := int(atomic.AddInt32(&captureCount, 1)) - 1
				if idx < len(captureSeq) {
					return []byte(captureSeq[idx]), nil
				}
				// Cycle the last entry if ticks exceed sequence length.
				return []byte(captureSeq[len(captureSeq)-1]), nil
			}
		}
		return []byte("ok"), nil
	})
	return runner, &captureCount
}

// makeIdleOutput returns a string that detectIdlePrompt will recognise as
// an idle claude CLI prompt (trailing ❯ on the last non-empty line).
//
// Use this to construct the final entry in a captureSeq for
// SimulatePaneHashStable scenarios.
func makeIdleOutput(prefix string) string {
	return fmt.Sprintf("%s\n❯\n", prefix)
}
