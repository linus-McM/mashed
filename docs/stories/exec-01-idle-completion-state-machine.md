# exec-01: `waitForIdleCompletion` state machine + `executeProcessNode` refactor

**Status:** ready
**Domain:** backend
**Size:** L
**Depends on:** exec-00
**Phase:** 3

## Description

Replace the "pane_dead" completion signal that `executeNode` has always relied on with a three-stage idle-prompt state machine. This story introduces `waitForIdleCompletion(ctx, target, timeout)` and refactors the current `executeNode` body into `executeProcessNode`, wrapping the claude invocation in `bash -c '...; exec bash'` so the pane stays alive after claude exits. Pane-death is preserved as an implicit completion signal for backwards compatibility.

**This is the highest-risk story in the sprint.** It touches every workflow, not just command nodes. The entire BMAD executor test suite must continue to pass after the refactor — `CommandRunner` mocks will need updating to simulate the new "hash changes then stabilises with trailing `❯`" signal instead of "pane_dead → 1".

## Developer Notes

- **Files to create/modify:**
  - `internal/bmad/executor.go`:
    - Add `waitForIdleCompletion(ctx context.Context, target string, timeout time.Duration) error` method on `*Executor`
    - Rename current `executeNode` body to `executeProcessNode` (keep same signature)
    - Introduce a thin `executeNode` dispatcher that branches on `EffectiveType()` — this story wires in `NodeTypeProcess` (and the default case); `NodeTypeCommand` is filled in by exec-03
    - Wrap the claude invocation in `bash -c '...; exec bash'` with single-quote escaping
    - Replace the pane-death poll loop (plan §Phase 3 "Current `executeNode` completion path" lines 942-989) with a call to `waitForIdleCompletion`
    - Keep `pollNodeSignals` side-effect wiring so snackbar events still fire
  - `internal/bmad/executor_test.go` — add five table-driven cases covering all stages (priming → work → stable → nil; priming → no-work → ErrIdleTimeoutNoStart; pane-dead during waiting-for-work → nil; pane-dead during watching-for-idle → nil; ctx cancel in each stage → ctx.Err)
  - `internal/bmad/mock_runner.go` (or the existing test helper that holds `CommandRunner` mocks) — add helpers that simulate the new state transitions: `SimulatePaneHashChange(ticks)`, `SimulatePaneHashStable(ticks)`, `SimulatePaneDead()`
  - `internal/bmad/shell_quote_test.go` — NEW. Unit test for the single-quote escape path, including backticks, dollar signs, embedded quotes
- **Types/symbols introduced:**
  - `waitForIdleCompletion(ctx, target, timeout) error`
  - `executeProcessNode(ctx, state, nodeIndex, nodeID, repoPath, model)` (lifted body)
  - `executeNode(ctx, ...)` now a dispatcher
  - Internal stage constants: `stagePriming`, `stageWaitingForWork`, `stageWatchingForIdle` (local `type stage int` inside `waitForIdleCompletion`)
  - `ErrIdleTimeoutNoStart` (declared in exec-00, used here)
- **Full state machine (from plan §Phase 3 "Piece 1"):**
  1. **PRIMING** — first poll captures baseline hash; idle cannot fire
  2. **WAITING_FOR_WORK** — subsequent polls compare to baseline; idle cannot fire until hash changes
  3. **WATCHING_FOR_IDLE** — once hash has changed at least once, watch for hash unchanged across two polls AND `detectIdlePrompt == true` → return nil
  - Pane-death at any stage → return nil (legacy behaviour preserved)
  - Ctx cancel → return ctx.Err()
  - Deadline exceeded in WAITING_FOR_WORK → return `ErrIdleTimeoutNoStart`
- **Shell quoting (plan §Phase 3 "Piece 3" + "Load-bearing warnings"):**
  ```go
  innerCommand := fmt.Sprintf(`claude --dangerously-skip-permissions --model %s "use %s%s"`, model, proc.SkillName, contextStr)
  command := fmt.Sprintf(`bash -c '%s; exec bash'`, strings.ReplaceAll(innerCommand, `'`, `'\''`))
  ```
  Dedicated unit test MUST assert the generated argv for an inner command containing `'`, `` ` ``, and `$`.
- **Risks / gotchas:**
  - **Plan §Phase 3 load-bearing warning #1 — shell quoting**: embedded single quotes are the most common breakage. The `'\''` escape is the portable answer. Test it explicitly.
  - **Plan §Phase 3 load-bearing warning #2 — existing executor tests will break.** The `CommandRunner` mock's expected call sequence (`tmux new-session ; list-panes ; list-panes ; pane_dead=1`) is hardcoded across the suite. Update the shared mock helpers FIRST, then run the suite and fix compilation/assertion failures one at a time.
  - **`pollNodeSignals` MUST still fire.** It emits frontend snackbar events. The simplest path is to have `waitForIdleCompletion` call `pollNodeSignals` internally on each tick with the current capture.
  - **Polling interval** — reuse `e.pollInterval`, do NOT add a new knob.
  - **Deadline vs ticker interaction** — compute `deadline := time.Now().Add(timeout)` once; the ticker fires at `pollInterval`, and the deadline check gates the `stageWaitingForWork` path only. Pane death and idle completion are not subject to the `ErrIdleTimeoutNoStart` deadline — only "claude never produced any output" is.
- **Prerequisites already in place:**
  - `detectIdlePrompt(string) bool` in `internal/bmad/question.go`
  - `hashCapturedOutput(string) string` in `internal/bmad/question.go`
  - `pollForIdle` existing state machine in `internal/bmad/executor.go:1152-1200` — this story does NOT delete it (exec-03 may refactor it); it coexists for now
  - `captureQuestionOutput(ctx, target)` helper for reading pane bytes
  - `e.runCmd` for invoking tmux
  - `pollNodeSignals` side-effect helper
  - `ErrIdleTimeoutNoStart` sentinel (declared in exec-00)

## Acceptance Criteria

**AC-1: `waitForIdleCompletion` happy path — priming → work → stable**
- Given a mock pane whose capture hash is H1 on tick 1, H1 on tick 2 (baseline held), H2 on tick 3 (work), H3 on tick 4, H3 on tick 5 (stable) AND `detectIdlePrompt` returns true on tick 5
- When `waitForIdleCompletion` is called with a 30-minute timeout
- Then it returns `nil`
- And it called `captureQuestionOutput` at least five times

**AC-2: `waitForIdleCompletion` — priming → no work → timeout**
- Given a mock pane whose hash never changes from the baseline
- When `waitForIdleCompletion` is called with a 3-second timeout
- Then it returns `ErrIdleTimeoutNoStart`

**AC-3: `waitForIdleCompletion` — pane death treated as completion**
- Given a mock pane that reports `pane_dead=1` on tick 3 (during WAITING_FOR_WORK)
- When `waitForIdleCompletion` is called
- Then it returns `nil`
- Given a mock pane that reports `pane_dead=1` on tick 6 (during WATCHING_FOR_IDLE)
- When `waitForIdleCompletion` is called
- Then it returns `nil`

**AC-4: `waitForIdleCompletion` — ctx cancel in each stage**
- Given a mock pane and a cancellable ctx
- When the ctx is cancelled during PRIMING / WAITING_FOR_WORK / WATCHING_FOR_IDLE
- Then `waitForIdleCompletion` returns `ctx.Err()` in each case

**AC-5: `executeProcessNode` wraps claude in `bash -c '...; exec bash'`**
- Given a process node with `SkillName: "brainstorm"` and a context string containing a single quote
- When `executeProcessNode` builds the tmux new-session command
- Then the generated argv contains `bash -c '...; exec bash'`
- And the inner single quote is escaped as `'\''`

**AC-6: Existing executor tests keep passing after mock updates**
- Given the current `internal/bmad/executor_test.go` suite
- When the mock helpers are updated to emit the new state-machine signals
- Then every previously-passing test still passes
- And no test is skipped or removed

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Idle-prompt completion state machine

  Scenario: Priming then work then stable
    Given a mock tmux pane with hashes [H1, H1, H2, H3, H3] across 5 ticks
    And detectIdlePrompt returns true on the 5th capture
    When waitForIdleCompletion is called with timeout 30m
    Then it returns nil
    And at least 5 capture calls were made

  Scenario: Claude never produces output
    Given a mock tmux pane whose hash stays at H1 for 20 ticks
    When waitForIdleCompletion is called with timeout 3s
    Then it returns ErrIdleTimeoutNoStart

  Scenario: Pane dies during waiting-for-work
    Given a mock tmux pane whose hash is H1 on ticks 1-2 then list-panes reports pane_dead=1 on tick 3
    When waitForIdleCompletion is called
    Then it returns nil (pane death == completion, legacy)

  Scenario: Pane dies during watching-for-idle
    Given a mock tmux pane that transitions through PRIMING → WAITING_FOR_WORK → WATCHING_FOR_IDLE and then pane_dead=1
    When waitForIdleCompletion is called
    Then it returns nil

  Scenario: Context cancellation during priming
    Given a cancellable context cancelled during PRIMING
    When waitForIdleCompletion is called
    Then the return value equals ctx.Err()

  Scenario: Claude wrapped in bash exec bash
    Given a process node with SkillName "brainstorm" and context 'it''s ready'
    When executeProcessNode builds the tmux new-session command
    Then the inner claude invocation is wrapped with "bash -c '...; exec bash'"
    And the inner single quote is escaped as '\''

  Scenario: Existing executor tests still pass
    Given the full executor_test.go suite after mock helper updates
    When `go test ./internal/bmad/... -race -short` runs
    Then every test passes
```

## Tasks / Subtasks

- [ ] Task 1 — Implement `waitForIdleCompletion` (AC-1, AC-2, AC-3, AC-4)
  - [ ] Add the state-machine method with stages `stagePriming`, `stageWaitingForWork`, `stageWatchingForIdle`
  - [ ] Call `pollNodeSignals` from each tick to preserve snackbar events
  - [ ] Ticker on `e.pollInterval`, deadline check applies only to WAITING_FOR_WORK
- [ ] Task 2 — Lift `executeNode` body into `executeProcessNode` (AC-5, AC-6)
  - [ ] Rename the method and wire the thin dispatcher `executeNode` on `EffectiveType()` (default + process only; command branch remains fail-fast from skills-cmd-01)
  - [ ] Wrap the claude invocation in `bash -c '...; exec bash'` with `'\''` escaping
  - [ ] Replace the pane-death poll loop with a call to `waitForIdleCompletion`
- [ ] Task 3 — Update shared mock helpers (AC-6)
  - [ ] Add `SimulatePaneHashChange`, `SimulatePaneHashStable`, `SimulatePaneDead` test helpers
  - [ ] Update every existing test that relied on the pane-death-only transition
- [ ] Task 4 — Unit tests for the state machine (AC-1, AC-2, AC-3, AC-4)
  - [ ] Five table cases: happy, no-work, pane-dead×2, ctx-cancel×3
- [ ] Task 5 — Shell-quote escape test (AC-5)
  - [ ] Dedicated test asserting argv for inner commands with `'`, `` ` ``, `$`, mixed

## Definition of Done

- [ ] All ACs verified by an automated test (Go table-driven; no "manually verified")
- [ ] Coverage ≥ 80% on modified files
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean — full BMAD suite, not just new tests
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths or magic numbers added
- [ ] Existing tests still pass (zero regressions)
