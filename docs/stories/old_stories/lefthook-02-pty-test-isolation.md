# Story 2: PTY Test Isolation with testing.Short() Guards

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** none
**Status:** done

## Description

Add `testing.Short()` skip guards to all PTY-dependent tests in `internal/terminal/` so they are skipped when `go test -short` is used. These tests use `creack/pty` which requires `fork/exec` and fails in sandboxed environments (macOS App Sandbox, CI runners, lefthook subshells). The skip guards keep `just test` working locally while pre-commit hooks (Story 4) can use `-short` to avoid PTY failures.

## Developer Notes

### Architecture
- **Modified files:**
  - `internal/terminal/session_test.go` -- add skip guard to every test function or at the top of each `Test*` function
  - `internal/terminal/manager_test.go` -- same treatment
- **Package:** `terminal` (existing)
- **No new types or interfaces** -- this is purely test infrastructure

### Technical Considerations
- **Skip pattern:** At the top of each `func Test*(t *testing.T)` that uses PTY:
  ```go
  if testing.Short() {
      t.Skip("requires PTY (fork/exec)")
  }
  ```
- **Scope:** Every test in `session_test.go` and `manager_test.go` uses PTY. Add the guard to ALL test functions in these files.
- **bridge_test.go:** Check if `internal/terminal/bridge_test.go` also has PTY-dependent tests. If so, add guards there too.
- **Subtests:** If tests use `t.Run()` for subtests, the skip should be at the parent `Test*` level, not inside each subtest. This avoids the overhead of setting up PTY fixtures before hitting the skip.
- **Do NOT modify non-PTY tests:** Any test that only tests pure logic (e.g., data structures, parsing) should NOT get a skip guard.

### Risks & Edge Cases
- **Over-skipping:** If a test function mixes PTY and non-PTY subtests, extract the non-PTY subtests into a separate `Test*` function first, then add the skip only to the PTY function. However, in the current codebase all terminal tests are PTY-dependent.
- **Forgotten tests:** New tests added later must also include the skip guard. Add a comment at the top of each test file: `// NOTE: All tests in this file require a real PTY. Add testing.Short() skip to new tests.`
- **`bridge_test.go` assessment:** Read the file before modifying. It may have WebSocket-only tests that don't need PTY skips.

### Reference Files
- `internal/terminal/session_test.go` -- current test file, all tests use `creack/pty` fork/exec
- `internal/terminal/manager_test.go` -- current test file, uses `SessionManager.Spawn()` which creates PTY
- `internal/terminal/bridge_test.go` -- may have PTY-dependent tests, check before modifying

## Acceptance Criteria

AC-1: PTY tests are skipped with `-short` flag
- Given the test files in `internal/terminal/` contain PTY-dependent tests
- When a developer runs `go test -short ./internal/terminal/...`
- Then all PTY-dependent tests are skipped with message "requires PTY (fork/exec)"
- And the test run exits with code 0 (no failures)

AC-2: PTY tests still run without `-short` flag
- Given a developer is on a machine with PTY support (local macOS terminal)
- When they run `go test ./internal/terminal/...` (without `-short`)
- Then all PTY tests execute normally
- And no tests are skipped (unless they have other skip conditions)

AC-3: Skip guard is applied consistently to all PTY test functions
- Given `session_test.go` and `manager_test.go` contain test functions
- When reviewing the modified files
- Then every `func Test*(t *testing.T)` that uses PTY has `testing.Short()` skip as its first statement
- And a file-level comment documents the PTY requirement for future contributors

AC-4: Full test suite passes with `-short` flag
- Given the terminal tests are skipped
- When a developer runs `go test -short -count=1 ./internal/...`
- Then the entire internal package suite passes
- And no terminal test failures occur

## BDD Test Scenarios

### Scenario 1: Short mode skips PTY tests
```gherkin
Feature: PTY test isolation

  Scenario: Developer runs tests in short mode
    Given the terminal package has PTY-dependent tests
    And testing.Short() skip guards are in place
    When the developer runs "go test -short ./internal/terminal/..."
    Then the output shows "SKIP" for each PTY test
    And the skip message contains "requires PTY (fork/exec)"
    And the exit code is 0

  Scenario: Developer runs tests in full mode
    Given the developer has a local terminal with PTY support
    When the developer runs "go test ./internal/terminal/..."
    Then all PTY tests execute and pass
    And no tests show "SKIP"
```

### Scenario 2: Integration with broader test suite
```gherkin
  Scenario: Short mode works for entire internal package tree
    Given terminal tests have skip guards
    And other packages (bmad, agent, scanner, git) have no PTY dependencies
    When the developer runs "go test -short -count=1 ./internal/..."
    Then terminal tests are skipped
    And all other package tests run normally
    And the exit code is 0
```

### Scenario 3: New test reminder
```gherkin
  Scenario: File-level comment guides future contributors
    Given a developer opens session_test.go or manager_test.go
    When they read the file header
    Then they see a comment stating all tests require PTY
    And the comment instructs them to add testing.Short() skip to new tests
```

## Tasks / Subtasks

- [ ] Task 1: Add skip guards to session_test.go (AC: AC-1, AC-2, AC-3)
  - [ ] Subtask 1a: Read `internal/terminal/session_test.go` and identify all `func Test*` functions
  - [ ] Subtask 1b: Add `if testing.Short() { t.Skip("requires PTY (fork/exec)") }` as the first statement in each test function
  - [ ] Subtask 1c: Add a file-level comment at the top noting PTY requirement

- [ ] Task 2: Add skip guards to manager_test.go (AC: AC-1, AC-2, AC-3)
  - [ ] Subtask 2a: Read `internal/terminal/manager_test.go` and identify all `func Test*` functions
  - [ ] Subtask 2b: Add skip guards to each test function
  - [ ] Subtask 2c: Add file-level comment

- [ ] Task 3: Assess and optionally guard bridge_test.go (AC: AC-1, AC-3)
  - [ ] Subtask 3a: Read `internal/terminal/bridge_test.go` and determine if tests use PTY
  - [ ] Subtask 3b: If PTY-dependent, add skip guards; if not, leave unchanged

- [ ] Task 4: Verify full suite passes with -short (AC: AC-4)
  - [ ] Subtask 4a: Run `go test -short -count=1 ./internal/...` and confirm zero failures
  - [ ] Subtask 4b: Run `go test -short ./internal/terminal/...` and confirm all tests are skipped

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] `go test -short -count=1 ./internal/...` passes with 0 failures
- [ ] `go test -short ./internal/terminal/...` shows all tests skipped
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `/simplify` run on all modified test files
- [ ] File-level comments added to both test files
