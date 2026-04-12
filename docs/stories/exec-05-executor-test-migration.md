# exec-05: Executor test suite migration + full regression

**Status:** done
**Domain:** backend
**Size:** M
**Depends on:** exec-01, exec-02, exec-03, exec-04
**Phase:** 3

## Description

After exec-01 through exec-04 land, the existing `CommandRunner` mock helpers may still carry hardcoded "pane_dead → 1" sequences from before the state-machine rewrite. Individual stories updated what they touched, but this story is the belt-and-braces full sweep: every remaining test that relied on the old completion signal gets migrated to the new hash-change + idle-prompt signal, and the full BMAD test suite runs clean under `-race -short` with zero skipped or removed tests.

This story also owns the Phase 3 exit criteria: `go test ./... -race` clean with ≥ 80% coverage on `internal/bmad/executor.go`, and a documented regression smoke run against the Wails dev app.

## Developer Notes

- **Files to create/modify:**
  - `internal/bmad/executor_test.go`, `internal/bmad/*_test.go` — sweep for any remaining usage of legacy mock patterns. Expect to find some in tests that spawn ad-hoc fixtures for tier-based runDynamic scenarios.
  - `internal/bmad/mock_runner.go` (or equivalent helper file) — finalise the canonical mock helpers:
    - `SimulatePaneHashChange(target string, ticks int)` — hash varies per tick, `detectIdlePrompt` returns false
    - `SimulatePaneHashStable(target string, ticks int)` — hash constant, `detectIdlePrompt` returns true on the second-to-last tick
    - `SimulatePaneDead(target string)` — `list-panes -F '#{pane_dead}'` returns `"1"`
    - `SimulateIdleTimeout(target string)` — hash constant and `detectIdlePrompt` returns false (→ `ErrIdleTimeoutNoStart`)
  - `internal/bmad/CODEOWNERS` or `internal/bmad/TESTING.md` (if the file exists) — document the canonical mock helpers so future engineers know which to use
- **What to look for in the sweep:**
  - Literal `"#{pane_dead}"` strings in test setup → replace with the helper
  - Literal `"1"` / `"0"` in mock responses for `list-panes` → replace with helper
  - Tests that set up `CommandRunner` with an exact sequence of tmux invocations in a specific order — the new state machine calls `captureQuestionOutput` more often, so call-count assertions may need relaxing to "at least N" instead of "exactly N"
  - Tests that assert a specific completion latency in ticks — relax or remove; the state machine's tick budget is different
- **Risks / gotchas:**
  - **Do NOT skip tests.** Every test must migrate or be deleted with justification in the commit message. A skipped test is a rotting test.
  - **Watch for flaky tests** introduced by the longer completion path. The state machine can take more ticks to complete than the old pane-death signal. Bump per-test timeouts if needed, but document why.
  - **Coverage on `executor.go` is the headline metric.** Run `go test -coverprofile=cover.out ./internal/bmad/... && go tool cover -func=cover.out | grep executor.go` and confirm ≥ 80%. If the new code paths are under-covered, add focused tests rather than dropping coverage.
  - The Phase 3 dev-app smoke test (drag a command onto a canvas, click Run, confirm injection and completion) is a required Definition-of-Done item but NOT a go-test target — document in the commit message with screenshots or a tmux capture excerpt.
- **Prerequisites already in place:**
  - exec-01: `waitForIdleCompletion`, `executeProcessNode`, initial mock helper updates
  - exec-02: `spawnCommandSession`, `SendInputToTarget`, `resolveCommandSession`
  - exec-03: `executeCommandNode`, `injectSlashCommand`, dispatcher wired
  - exec-04: `killWorkflowChainTails` terminal-state cleanup

## Acceptance Criteria

**AC-1: Canonical mock helpers exist and are documented**
- Given the helper file
- When a developer reads it
- Then the four helpers (`SimulatePaneHashChange`, `SimulatePaneHashStable`, `SimulatePaneDead`, `SimulateIdleTimeout`) are present, exported or package-internal as appropriate, and have godoc comments explaining their state-machine mapping
- And at least one test case uses each helper

**AC-2: Full BMAD test suite passes under `-race -short`**
- Given the post-exec-01..04 codebase
- When `go test ./internal/bmad/... -race -short -count=1` runs
- Then every test passes
- And zero tests are skipped
- And zero tests are marked `t.Skip(...)` with a TODO

**AC-3: Coverage ≥ 80% on `executor.go`**
- Given `go test -coverprofile=cover.out ./internal/bmad/...`
- When `go tool cover -func=cover.out | grep executor.go` runs
- Then the reported coverage is ≥ 80%

**AC-4: Legacy pane-death mock patterns are eliminated**
- Given a grep for `"#{pane_dead}"` in `internal/bmad/*_test.go`
- When the grep runs after this story
- Then zero matches appear outside the canonical helper file
- And the helper file is the single source of truth

**AC-5: Manual smoke test documented**
- Given the Wails dev app
- When a command node downstream of a process node is dragged onto the canvas and the workflow is run
- Then the command injects `/<commandName>\n` into the parent's session
- And the command completes cleanly
- And a capture excerpt (tmux capture-pane -p) or screenshot is attached to the commit message

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Executor test suite migration

  Scenario: Canonical helpers cover all four state-machine transitions
    Given the canonical mock helper file
    When a developer reads the godoc
    Then SimulatePaneHashChange maps to stageWaitingForWork
    And SimulatePaneHashStable maps to stageWatchingForIdle with idle prompt true
    And SimulatePaneDead maps to legacy pane-death completion
    And SimulateIdleTimeout maps to ErrIdleTimeoutNoStart

  Scenario: Full BMAD test suite passes with -race -short
    Given the post-migration codebase
    When `go test ./internal/bmad/... -race -short -count=1` runs
    Then every test passes
    And no test is skipped

  Scenario: Coverage threshold met on executor.go
    Given `go test -coverprofile=cover.out ./internal/bmad/...`
    When `go tool cover -func=cover.out` is inspected
    Then the total for executor.go is at least 80%

  Scenario: No legacy pane-death literals remain
    Given the test directory
    When grepped for "#{pane_dead}"
    Then only the canonical helper file contains matches

  Scenario: Manual smoke test is documented
    Given the Phase 3 commit message
    When inspected for a smoke-test section
    Then it references a dev-app run with a command node reusing a process-node session
    And a capture excerpt or screenshot is attached
```

## Tasks / Subtasks

- [x] Task 1 — Canonical helpers (AC-1)
  - [x] Consolidate the four helpers in a single mock file
  - [x] Add godoc comments mapping each to its state-machine stage
- [x] Task 2 — Test sweep (AC-4)
  - [x] Grep for `"#{pane_dead}"` across `internal/bmad/*_test.go`
  - [x] Replace every literal with a canonical helper call
  - [x] Relax exact-count assertions to "at least N" where appropriate
- [x] Task 3 — Full suite run + fixes (AC-2)
  - [x] `go test ./internal/bmad/... -race -short -count=1`
  - [x] Fix any flaky or broken test; NEVER `t.Skip`
- [x] Task 4 — Coverage check (AC-3)
  - [x] `go test -coverprofile=cover.out ./internal/bmad/...`
  - [x] `go tool cover -func=cover.out | grep executor.go`
  - [x] Add targeted tests if < 80%
- [x] Task 5 — Smoke test + commit message (AC-5)
  - [x] Run `wails dev`, drag a command onto a canvas with a process parent, click Run
  - [x] Capture the tmux pane output after completion
  - [x] Attach excerpt/screenshot to commit message

## Definition of Done

- [x] All ACs verified by an automated test or a documented commit-message smoke run
- [x] Coverage ≥ 80% on `internal/bmad/executor.go` specifically
- [x] `go build ./... && go vet ./...` clean
- [x] `go test ./... -race -short -count=1` clean, zero skips
- [x] `/simplify` run before sign-off
- [x] No hardcoded paths, magic numbers, or legacy mock patterns remain
- [x] Phase 3 commit message contains: verification outcomes (from exec-00), smoke-test artefact, coverage delta
- [x] Existing tests still pass
