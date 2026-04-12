# exec-02: `spawnCommandSession` + `SendInputToTarget` + session-reuse resolver

**Status:** done
**Domain:** backend
**Size:** M
**Depends on:** exec-00, exec-01
**Phase:** 3

## Description

Lay the session-management foundation `executeCommandNode` will consume in exec-03. Three independent pieces:

1. **Factor `spawnCommandSession` out of `executeProcessNode`** so both node types share a single spawn helper. The current `executeProcessNode` hard-codes the tmux spawn inline; exec-03 needs to call this when a command node has no live parent to reuse.
2. **Add `SendInputToTarget(ctx, target, data)` on `*TmuxAdapter`** — a target-string variant of the existing `(*TmuxAttachment).SendInput(data)`, so the executor can inject slash commands into sessions it didn't open itself.
3. **Implement `resolveCommandSession`** — scans incoming edges for a live upstream session; returns `(target, reused, err)` where `reused == true` means "parent's tmux session is alive and we inject", `reused == false` means "no live parent → caller spawns a fresh session". Multi-parent case is a documented PUNT (first-parent-wins = most-recently-started).

`executeCommandNode` itself is NOT in scope for this story — exec-03 assembles these pieces into the full command-node runner.

## Developer Notes

- **Files to create/modify:**
  - `internal/bmad/executor.go`:
    - Extract `spawnCommandSession(ctx, state, nodeID, repoPath, model, commandName) (string, error)` from the current `executeProcessNode` body. The extracted helper handles: building the `bash -c 'claude --dangerously-skip-permissions --model ... "/<commandName>"; exec bash'` argv, running `tmux new-session -d -s <name> ...`, recording the target on the node, and emitting the initial `bmad:node:status` with `TmuxTarget`.
    - `executeProcessNode` then calls `spawnCommandSession` with its existing skill-name payload (via a second helper `spawnProcessSession` or a shared helper with a mode flag — pick whichever reads cleaner; the plan §Phase 3 step 7 says "factor `spawnCommandSession` out" so keep a single helper if possible).
    - Add `resolveCommandSession(ctx, state, nodeIndex, nodeID, repoPath, model) (target string, reused bool, err error)` per the full implementation in plan §Phase 3 "Piece 4".
  - `internal/terminal/tmux_adapter.go`:
    - Add `(*TmuxAdapter).SendInputToTarget(ctx context.Context, target string, data []byte) error`. Thin wrapper that builds the same argv as `(*TmuxAttachment).SendInput(data)` but takes a `target` string directly. Reuse the existing hex-formatting loop at line 524+.
  - `internal/terminal/tmux_adapter_test.go`:
    - Add a unit test matching the existing `TestTmuxAttachment_AC3_SendInputUsesSendKeysHex` style, but for the target-string variant.
  - `internal/bmad/executor_test.go` (or new `executor_session_test.go`):
    - Table tests for `resolveCommandSession`: zero / one / many live parents, parent-TmuxTarget-set-but-session-dead-falls-through-to-spawn.
- **Types/symbols introduced:**
  - `spawnCommandSession(ctx, state, nodeID, repoPath, model, commandName) (string, error)` on `*Executor`
  - `resolveCommandSession(ctx, state, nodeIndex, nodeID, repoPath, model) (string, bool, error)` on `*Executor`
  - `(*TmuxAdapter).SendInputToTarget(ctx, target, data) error` on the adapter
- **Multi-parent merge is a PUNT (plan §Phase 3 "Piece 4" and "Load-bearing warnings"):** when `len(live) > 1`, pick the parent with the most recent `StartedAt` and log `"bmad: command node %s has %d live parents; reusing most recent (%s)"`. RFC3339 string compare is fine for the ordering. Document "first parent wins" in the code comment AND the commit message. Real merge-node semantics are Phase 4+.
- **Risks / gotchas:**
  - `resolveCommandSession` must call `e.runCmd(ctx, "tmux", "list-panes", "-t", n.TmuxTarget, "-F", "#{pane_dead}")` to verify each upstream session is actually alive. A stale `TmuxTarget` field on a dead pane must fall through to the zero-live case.
  - **Plan §Phase 3 "Piece 4" sketch has a bug:** it references `commandName` inside `resolveCommandSession`'s zero-live branch, but that local isn't in scope there. Plumb it through from the caller OR return a sentinel `errNoLiveParent` and have `executeCommandNode` (exec-03) call `spawnCommandSession` itself. Pick the cleaner option — the latter is preferred because it keeps `resolveCommandSession` pure.
  - `SendInputToTarget` must NOT depend on a live `*TmuxAttachment` — the executor works with strings. The hex-formatting loop is pure; lift it into a helper `formatSendKeysHex(data) []string` and call it from both `TmuxAttachment.SendInput` and `SendInputToTarget`.
  - Do not break existing `TmuxAttachment.SendInput` call sites. Tests for the existing method must still pass.
  - Spawn helper must preserve the existing `bash -c '...; exec bash'` wrapper from exec-01.
- **Prerequisites already in place:**
  - exec-01 shipped `executeProcessNode` with the `bash -c 'exec bash'` wrapper and `waitForIdleCompletion` completion signal.
  - `internal/terminal/tmux_adapter.go:524+` has working hex-byte injection for `TmuxAttachment.SendInput`.
  - `WorkflowNode.TmuxTarget` field exists and is populated by the spawn path.
  - `state.exec.WorkflowEdges()` (or inline edge iteration) is available in the executor.

## Acceptance Criteria

**AC-1: `spawnCommandSession` is extracted and usable for both process and command nodes**
- Given a refactor that extracts the spawn logic from `executeProcessNode`
- When `executeProcessNode` runs
- Then it calls the extracted helper and produces the same tmux argv as before the refactor (no regression)
- And the helper can be called directly with a `commandName` payload to spawn a new session whose initial claude invocation is `/<commandName>`

**AC-2: `SendInputToTarget` emits exact `send-keys -H` argv**
- Given `tmuxAdapter.SendInputToTarget(ctx, "bmad-abc:0.0", []byte("/simplify\n"))`
- When the underlying command runner captures the invocation
- Then the argv is `tmux send-keys -H -t bmad-abc:0.0 2f 73 69 6d 70 6c 69 66 79 0a` (verify hex bytes for `/`, `s`, `i`, `m`, `p`, `l`, `i`, `f`, `y`, `\n`)

**AC-3: `resolveCommandSession` — zero live parents returns `reused=false` with a sentinel**
- Given a command node whose only incoming edge points to a parent with empty `TmuxTarget`
- When `resolveCommandSession` runs
- Then it returns `reused=false`
- And the returned `target` is empty string
- And no error is returned (caller will spawn)

**AC-4: `resolveCommandSession` — one live parent returns its target, `reused=true`**
- Given a command node with one incoming edge to a parent whose `TmuxTarget` is live (`pane_dead=0`)
- When `resolveCommandSession` runs
- Then it returns `target == parent.TmuxTarget`, `reused == true`, `err == nil`

**AC-5: `resolveCommandSession` — stale parent target falls through to zero-live case**
- Given a command node with one incoming edge to a parent whose `TmuxTarget` is set but the pane reports `pane_dead=1`
- When `resolveCommandSession` runs
- Then it returns `reused=false` (equivalent to the zero-live case)

**AC-6: `resolveCommandSession` — multi-parent picks most recently started**
- Given a command node with two incoming edges, both parents alive, `parentA.StartedAt < parentB.StartedAt`
- When `resolveCommandSession` runs
- Then it returns `target == parentB.TmuxTarget`, `reused == true`
- And a log line is emitted containing the count of live parents and the chosen parent ID

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Session reuse and spawn helpers

  Scenario: Spawn helper produces bash-wrapped claude argv
    Given an executor refactored to use spawnCommandSession
    When a process node spawns its session
    Then the generated argv is unchanged from the exec-01 baseline
    And the inner claude invocation is wrapped with bash -c '...; exec bash'

  Scenario: Spawn helper accepts a commandName payload
    Given spawnCommandSession is invoked with commandName "simplify"
    When it builds the initial claude invocation
    Then the inner command line contains "/simplify"
    And the pane stays alive after claude exits (bash -c exec bash wrapper)

  Scenario: SendInputToTarget emits exact hex argv
    Given an adapter and target "bmad-abc:0.0"
    When SendInputToTarget is called with []byte("/simplify\n")
    Then the runner observes exactly: tmux send-keys -H -t bmad-abc:0.0 2f 73 69 6d 70 6c 69 66 79 0a

  Scenario: Zero live parents
    Given a command node whose parent has empty TmuxTarget
    When resolveCommandSession runs
    Then it returns reused=false and empty target

  Scenario: One live parent reused
    Given a command node with one parent whose TmuxTarget is alive
    When resolveCommandSession runs
    Then it returns the parent's TmuxTarget with reused=true

  Scenario: Stale parent target falls through
    Given a command node with one parent whose TmuxTarget points to a dead pane
    When resolveCommandSession runs
    Then it returns reused=false

  Scenario: Multi-parent picks most recent
    Given parentA (StartedAt 10:00) and parentB (StartedAt 10:05), both alive
    When resolveCommandSession runs for a node with edges from both
    Then it returns parentB's target with reused=true
    And a log line mentions "2 live parents"
```

## Tasks / Subtasks

- [x] Task 1 — Extract `spawnCommandSession` from `executeProcessNode` (AC-1)
  - [x] Create a single helper that handles the bash-wrapped claude spawn and records TmuxTarget on the node
  - [x] Rewire `executeProcessNode` to call the helper
  - [x] Ensure the helper accepts an optional `commandName` so it can spawn with `/<cmd>` as the initial invocation
- [x] Task 2 — Add `SendInputToTarget` on the adapter (AC-2)
  - [x] Lift the hex-formatting loop out of `TmuxAttachment.SendInput` into a shared `formatSendKeysHex(data) []string` helper
  - [x] Implement `(*TmuxAdapter).SendInputToTarget(ctx, target, data) error`
  - [x] Add a unit test matching `TestTmuxAttachment_AC3_SendInputUsesSendKeysHex` style
- [x] Task 3 — Implement `resolveCommandSession` (AC-3, AC-4, AC-5, AC-6)
  - [x] Collect upstream node pointers from incoming edges
  - [x] Filter by live `TmuxTarget` via `tmux list-panes -F '#{pane_dead}'`
  - [x] Zero / one / many branches per the plan
  - [x] Return `(target, reused, err)`; caller (exec-03) handles the spawn on `reused=false`
- [x] Task 4 — Tests (AC-1 through AC-6)
  - [x] Refactor regression test: process-node spawn argv unchanged
  - [x] SendInputToTarget argv test
  - [x] Table test for resolveCommandSession with four cases

## Definition of Done

- [x] All ACs verified by an automated test (Go table-driven; no "manually verified")
- [x] Coverage ≥ 80% on modified files
- [x] `go build ./... && go vet ./...` clean
- [x] `go test ./... -race -short` clean
- [x] `/simplify` run before sign-off
- [x] No hardcoded paths or magic numbers added
- [x] Existing tests still pass (zero regressions for TmuxAttachment.SendInput users)
