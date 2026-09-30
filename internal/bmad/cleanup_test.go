package bmad

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── CleanupStaleSessions Tests (bridge-04 RED) ──
//
// These tests cover bridge-04 AC-1..AC-4 for the (currently undefined)
// `(*Executor).CleanupStaleSessions(ctx)` method. They reuse the existing
// testHarness, cmdCall, findCall, and seedResponseState helpers from
// executor_test.go.

// cleanupRunner returns a CommandRunner tailored for CleanupStaleSessions
// tests. It records every invocation, answers `tmux list-sessions` with
// listOutput / listErr, and returns per-target errors for `tmux kill-session`
// based on killErrs.
func cleanupRunner(listOutput string, listErr error, killErrs map[string]error) (CommandRunner, *[]cmdCall, *sync.Mutex) {
	var calls []cmdCall
	var mu sync.Mutex
	runner := CommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		mu.Lock()
		calls = append(calls, cmdCall{name: name, args: append([]string(nil), args...)})
		mu.Unlock()

		if len(args) == 0 {
			return nil, nil
		}
		switch args[0] {
		case "list-sessions":
			if listErr != nil {
				return []byte(listOutput), listErr
			}
			return []byte(listOutput), nil
		case "kill-session":
			target := ""
			for i, a := range args {
				if a == "-t" && i+1 < len(args) {
					target = args[i+1]
				}
			}
			if e, ok := killErrs[target]; ok && e != nil {
				return nil, e
			}
			return []byte(""), nil
		}
		return []byte("ok"), nil
	})
	return runner, &calls, &mu
}

// killSessionTargets extracts the ordered list of `-t` targets from every
// recorded `tmux kill-session` call.
func killSessionTargets(calls []cmdCall) []string {
	var out []string
	for _, c := range calls {
		if c.name != "tmux" || len(c.args) == 0 || c.args[0] != "kill-session" {
			continue
		}
		for i, a := range c.args {
			if a == "-t" && i+1 < len(c.args) {
				out = append(out, c.args[i+1])
			}
		}
	}
	return out
}

// AC-1: All bmad-prefixed sessions are killed when no executions are tracked;
// non-bmad sessions are left alone.
//
// BDD: "Kill all orphans on startup"
func TestExecutor_CleanupStaleSessions_AC1_KillsOrphanedBmadSessions(t *testing.T) {
	h := newHarness(t)
	runner, calls, mu := cleanupRunner("bmad-a\nbmad-b\nuserfoo\n", nil, nil)
	h.executor.SetCommandRunner(runner)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := h.executor.CleanupStaleSessions(ctx)
	require.NoError(t, err, "AC-1: cleanup should succeed when all kills succeed")

	mu.Lock()
	snapshot := append([]cmdCall(nil), *calls...)
	mu.Unlock()

	require.NotNil(t,
		findCall(snapshot, "tmux", "list-sessions", "-F", "#{session_name}"),
		"AC-1: expected `tmux list-sessions -F #{session_name}` invocation, got: %+v", snapshot,
	)

	targets := killSessionTargets(snapshot)
	assert.ElementsMatch(t, []string{"bmad-a", "bmad-b"}, targets,
		"AC-1: both bmad-prefixed sessions should be killed")
	assert.NotContains(t, targets, "userfoo",
		"AC-1: non-bmad session userfoo must NEVER be killed")
}

// AC-2: A session referenced by a tracked execution's TmuxTarget is preserved;
// only the orphan is killed.
//
// BDD: "Preserve tracked sessions"
func TestExecutor_CleanupStaleSessions_AC2_PreservesTrackedSessions(t *testing.T) {
	h := newHarness(t)
	runner, calls, mu := cleanupRunner("bmad-a\nbmad-b\n", nil, nil)
	h.executor.SetCommandRunner(runner)

	// Seed a live execution whose node owns bmad-a (with the conventional
	// ":0.0" pane suffix). The cleanup must strip the suffix to derive
	// the tracked session name.
	seedResponseState(t, h, "exec-live", "node-1", "bmad-a:0.0", "")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := h.executor.CleanupStaleSessions(ctx)
	require.NoError(t, err, "AC-2: cleanup should succeed when only the orphan is killed")

	mu.Lock()
	snapshot := append([]cmdCall(nil), *calls...)
	mu.Unlock()

	targets := killSessionTargets(snapshot)
	assert.Equal(t, []string{"bmad-b"}, targets,
		"AC-2: only bmad-b should be killed (bmad-a is tracked)")
	assert.NotContains(t, targets, "bmad-a",
		"AC-2: bmad-a is referenced by a live execution and must be preserved")
}

// AC-3: A `no server running` error from list-sessions is swallowed; no kills
// are attempted.
//
// BDD: "No tmux server running"
func TestExecutor_CleanupStaleSessions_AC3_SwallowsNoServerRunning(t *testing.T) {
	h := newHarness(t)
	listErr := errors.New("no server running on /private/tmp/tmux-501/default")
	runner, calls, mu := cleanupRunner("", listErr, nil)
	h.executor.SetCommandRunner(runner)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := h.executor.CleanupStaleSessions(ctx)
	assert.NoError(t, err,
		"AC-3: list-sessions error containing 'no server running' must be swallowed")

	mu.Lock()
	snapshot := append([]cmdCall(nil), *calls...)
	mu.Unlock()

	assert.Empty(t, killSessionTargets(snapshot),
		"AC-3: no kill-session calls should be issued when no tmux server is running")
}

// AC-3 (macOS variant): tmux on macOS reports an absent daemon as
// "error connecting to <socket> (No such file or directory)" rather than
// "no server running". Both phrasings must be swallowed.
func TestExecutor_CleanupStaleSessions_AC3_SwallowsErrorConnectingTo(t *testing.T) {
	h := newHarness(t)
	listErr := errors.New("error connecting to /private/tmp/tmux-501/default (No such file or directory)")
	runner, calls, mu := cleanupRunner("", listErr, nil)
	h.executor.SetCommandRunner(runner)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := h.executor.CleanupStaleSessions(ctx)
	assert.NoError(t, err,
		"AC-3: list-sessions error containing 'error connecting to' must be swallowed (macOS tmux phrasing)")

	mu.Lock()
	snapshot := append([]cmdCall(nil), *calls...)
	mu.Unlock()

	assert.Empty(t, killSessionTargets(snapshot),
		"AC-3: no kill-session calls should be issued when no tmux server is running")
}

// AC-4: When an individual kill fails, the remaining kills are still attempted
// and the returned error wraps the first failure (errors.Is reachable).
//
// BDD: "Partial failure"
func TestExecutor_CleanupStaleSessions_AC4_ContinuesAfterIndividualKillFailure(t *testing.T) {
	h := newHarness(t)
	firstKillErr := errors.New("fake kill error")
	runner, calls, mu := cleanupRunner(
		"bmad-a\nbmad-b\nbmad-c\n",
		nil,
		map[string]error{"bmad-a": firstKillErr},
	)
	h.executor.SetCommandRunner(runner)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := h.executor.CleanupStaleSessions(ctx)
	require.Error(t, err, "AC-4: expected non-nil error when at least one kill fails")
	assert.True(t, errors.Is(err, firstKillErr),
		"AC-4: returned error must wrap the first kill failure (errors.Is), got %v", err)

	mu.Lock()
	snapshot := append([]cmdCall(nil), *calls...)
	mu.Unlock()

	targets := killSessionTargets(snapshot)
	assert.ElementsMatch(t,
		[]string{"bmad-a", "bmad-b", "bmad-c"}, targets,
		"AC-4: all three kill-session attempts must be made even after the first failure")
}
