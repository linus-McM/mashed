# exec-03: `executeCommandNode` + slash-command injection + dispatcher wire-up

**Status:** ready
**Domain:** backend
**Size:** M
**Depends on:** exec-00, exec-01, exec-02
**Phase:** 3

## Description

Assemble the pieces from exec-00/01/02 into the real command-node runner. `executeCommandNode` resolves a reusable session via `resolveCommandSession`, spawns a fresh one when no live parent exists, injects `/<commandName>\n` via `SendInputToTarget` when reusing a session, waits for the idle-prompt state machine to signal completion, captures output, and marks the node complete. The `executeNode` dispatcher (stubbed in skills-cmd-01 with fail-fast, then re-wired in exec-01 for process nodes) now routes `NodeTypeCommand` to this new runner, replacing the sentinel failure.

This story is the user-visible payoff for Phase 3: after it lands, a command node downstream of a process node actually runs the slash command in the parent's live claude session.

## Developer Notes

- **Files to create/modify:**
  - `internal/bmad/executor.go`:
    - Add `executeCommandNode(ctx, state, nodeIndex, nodeID, repoPath, model)` per plan §Phase 3 "Piece 4" full implementation
    - Add `injectSlashCommand(ctx, target, name) error` — thin wrapper around `e.tmuxAdapter.SendInputToTarget(ctx, target, []byte("/"+name+"\n"))`
    - Replace the `NodeTypeCommand` fail-fast branch in `executeNode` dispatcher with a call to `executeCommandNode`
  - `internal/bmad/executor_test.go` (or new `executor_command_test.go`):
    - Unit test for `injectSlashCommand` argv
    - Integration test: `process → command → command` DAG via mock runner, all three share the same `TmuxTarget`
    - Integration test: `command` node with no live parent spawns a fresh session
    - Integration test: missing `commandName` in config fails the node with a clear log line
- **Full `executeCommandNode` flow (from plan §Phase 3 "Piece 4"):**
  1. Mark running, emit `bmad:node:status` with `NodeRunning`
  2. Read `commandName` from `node.Config["commandName"]`; fail-fast if empty
  3. Call `resolveCommandSession` → `(target, reused, err)`
  4. On `reused == false`: call `spawnCommandSession` with the `commandName` payload (spawns a fresh pane whose initial claude invocation IS the slash command, so no injection needed)
  5. Record `node.TmuxTarget = target`, emit another `bmad:node:status` with the target
  6. On `reused == true`: call `injectSlashCommand(ctx, target, commandName)`; fail on error
  7. Call `waitForIdleCompletion(ctx, target, 30*time.Minute)`; fail on error
  8. Best-effort `captureOutput(ctx, target)` → `state.exec.NodeOutputs[nodeID]`
  9. `completeNode(state, idx, nodeID)`
- **Timing observation (plan §Phase 3 "Load-bearing warnings" #3):** when reusing an existing session, the pane is ALREADY at `❯` when we inject. The state machine enters `stagePriming`, captures the post-injection hash as baseline, waits for it to change (claude responds to `/simplify`), then watches for idle. This is correct by construction — no additional delay needed before injection — but write an explicit test that exercises the timing to prevent future regressions.
- **Risks / gotchas:**
  - **Plan §Phase 3 "Load-bearing warning" #3 — injection timing**: the state machine priming step handles this correctly, but only if `injectSlashCommand` runs BEFORE `waitForIdleCompletion` begins. Verify the call order in the test.
  - **Sequential ordering matters**: `resolveCommandSession` → `spawnCommandSession` (if needed) → record target → inject (if reused) → wait → capture → complete. Do NOT reorder.
  - **Always emit TmuxTarget** — the frontend View Terminal button depends on it (plan §Phase 3 appendix "Node status events"). Emit twice if needed: once with running-without-target (instant feedback), once with running-with-target (after resolve/spawn).
  - **Missing `commandName` must fail loud, not silent** — log and `failNode`. Do not crash.
  - `waitForIdleCompletion` is a blocking call; `executeCommandNode` runs inside the executor's per-node goroutine, so blocking is expected and safe.
- **Prerequisites already in place:**
  - exec-00: fixture-verified slash injection + `ErrIdleTimeoutNoStart` sentinel
  - exec-01: `waitForIdleCompletion`, `executeNode` dispatcher, `executeProcessNode` refactor
  - exec-02: `spawnCommandSession`, `SendInputToTarget`, `resolveCommandSession`
  - skills-cmd-01: `NodeTypeCommand` constant, `config["commandName"]` convention

## Acceptance Criteria

**AC-1: `injectSlashCommand` produces exact argv**
- Given `injectSlashCommand(ctx, "bmad-abc:0.0", "simplify")`
- When the adapter observes the call
- Then the runner sees `tmux send-keys -H -t bmad-abc:0.0 2f 73 69 6d 70 6c 69 66 79 0a`

**AC-2: Chained command reuses upstream session**
- Given a DAG `process(A) → command(B) → command(C)` driven through the mock runner
- When the workflow runs
- Then node A spawns a session with target `T1`
- And node B reuses `T1` (no new `tmux new-session` invocation for B)
- And node C reuses `T1` (no new `tmux new-session` invocation for C)
- And both B and C inject their respective `/<commandName>\n` sequences into `T1`
- And all three nodes complete in topological order

**AC-3: Command with no live parent spawns a fresh session**
- Given a DAG with a single command node `X` (no incoming edges)
- When the workflow runs
- Then `spawnCommandSession` is called with `commandName == X.config["commandName"]`
- And the inner claude invocation in the spawned pane contains `/<commandName>`
- And no `injectSlashCommand` is called (the slash is already the initial argument)

**AC-4: Missing `commandName` fails the node with a clear error**
- Given a command node whose config lacks `commandName` (empty string)
- When `executeCommandNode` runs
- Then the node transitions to `failed`
- And the log contains `"command node %s missing commandName in config"`
- And no tmux invocation occurred

**AC-5: Dispatcher routes `NodeTypeCommand` to `executeCommandNode`, not the Phase 2 fail-fast**
- Given the `executeNode` dispatcher
- When a node with `EffectiveType() == NodeTypeCommand` is executed
- Then `executeCommandNode` runs (not `failNode`)
- And the Phase 2 sentinel log line `"command nodes not yet runnable"` does NOT appear

**AC-6: Injection ordering for a reused session is correct**
- Given a reused session with a mock pane
- When `executeCommandNode` invokes `injectSlashCommand` then `waitForIdleCompletion`
- Then injection happens BEFORE the state machine enters `stagePriming`
- And the state machine observes a hash change in `stageWaitingForWork` (claude responded to the slash command)
- And the node completes when the state machine reaches stable idle

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Command node execution

  Scenario: Slash injection argv matches exactly
    Given injectSlashCommand is invoked with target "bmad-abc:0.0" and name "simplify"
    When the mock adapter captures the call
    Then the runner observes: tmux send-keys -H -t bmad-abc:0.0 2f 73 69 6d 70 6c 69 66 79 0a

  Scenario: Process then two chained commands share a session
    Given a DAG process(A) -> command(B: simplify) -> command(C: review)
    When the workflow runs end-to-end through the mock runner
    Then node A spawns target T1
    And node B reuses T1 (no new-session call) and injects "/simplify"
    And node C reuses T1 (no new-session call) and injects "/review"
    And all three nodes complete successfully

  Scenario: Command with no parent spawns fresh
    Given a workflow with a single command node X (commandName "brainstorm") and no incoming edges
    When the workflow runs
    Then spawnCommandSession is called with commandName "brainstorm"
    And injectSlashCommand is never called
    And the spawned pane's initial claude argument contains "/brainstorm"

  Scenario: Missing commandName fails loud
    Given a command node whose config map lacks "commandName"
    When executeCommandNode runs
    Then the node transitions to failed
    And the log output contains "missing commandName"
    And no tmux invocation occurred

  Scenario: Dispatcher replaces Phase 2 fail-fast
    Given a command node and a fully wired Phase 3 executor
    When executeNode runs
    Then executeCommandNode is invoked
    And the Phase 2 sentinel log line "command nodes not yet runnable" does not appear

  Scenario: Reused session injection timing
    Given a command node reusing a parent session whose pane is at idle
    When executeCommandNode injects the slash command then waits for idle
    Then injection happens before stagePriming
    And the state machine observes a hash change in stageWaitingForWork
    And the node completes at the next stable idle
```

## Tasks / Subtasks

- [ ] Task 1 — Implement `injectSlashCommand` (AC-1)
  - [ ] Thin wrapper around `e.tmuxAdapter.SendInputToTarget(ctx, target, []byte("/"+name+"\n"))`
  - [ ] Unit test for exact argv
- [ ] Task 2 — Implement `executeCommandNode` (AC-2, AC-3, AC-4, AC-6)
  - [ ] Full flow per Developer Notes step-by-step
  - [ ] Emit `bmad:node:status` twice if needed (pre-resolve, post-resolve-with-target)
  - [ ] Fail-fast on missing `commandName`
  - [ ] Call sequence: resolve → spawn (if needed) → inject (if reused) → wait → capture → complete
- [ ] Task 3 — Wire the dispatcher (AC-5)
  - [ ] Replace the Phase 2 fail-fast branch in `executeNode` with a call to `executeCommandNode`
  - [ ] Remove (or archive) the `"command nodes not yet runnable"` log line
- [ ] Task 4 — Integration tests (AC-2, AC-3, AC-6)
  - [ ] 3-node DAG `process → command → command` via mock runner; assert shared target + injection sequence
  - [ ] Single-command-node DAG; assert spawn path
  - [ ] Reused-session timing test; assert ordering of inject vs wait
- [ ] Task 5 — Manual smoke test in Wails dev app (documented in commit message)
  - [ ] Drag a command onto a canvas with a process node upstream
  - [ ] Click Run
  - [ ] Confirm the command injects and completes

## Definition of Done

- [ ] All ACs verified by an automated test (Go table-driven + integration; no "manually verified")
- [ ] Coverage ≥ 80% on modified files
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths or magic numbers added
- [ ] Existing tests still pass
