# Story 5: Justfile Targets for Testing and Hook Management

**Priority:** P1-high
**Domain:** fullstack
**Estimated Complexity:** S
**Depends On:** lefthook-01, lefthook-03, lefthook-04
**Status:** done

## Description

Add new targets to the `justfile` for running the full test suite (Go + frontend), coverage enforcement, Go linting, hook installation, and manual pre-commit checking. These targets provide convenient developer shortcuts that mirror what the lefthook hooks do, without requiring a git commit/push to trigger them. Developers can run `just test-all` during development to catch issues before committing.

## Developer Notes

### Architecture
- **Modified file:** `justfile` (repo root)
- **No new files** -- purely additive targets to the existing justfile
- **Targets to add:**
  - `test-all` -- runs Go tests then frontend tests
  - `test-cover` -- runs the coverage enforcement script from Story 1
  - `lint` -- runs `go vet ./...`
  - `hooks-install` -- runs `lefthook install`
  - `pre-check` -- runs `lefthook run pre-commit` manually

### Technical Considerations
- **`test-all` depends on existing `test` target:** Use just's dependency syntax: `test-all: test` to run Go tests first, then frontend tests as a separate step.
- **Frontend test command:** `cd frontend && npx vitest run` -- must run from the `frontend/` directory since vitest needs access to `node_modules` and `vitest.config.js`.
- **`test-cover` command:** `bash scripts/check-coverage.sh` -- simple delegation to the script from Story 1.
- **`lint` command:** `go vet ./...` for now. The plan says "go vet for now" suggesting future expansion to golangci-lint or similar.
- **`hooks-install` command:** `lefthook install` -- installs lefthook's git hooks shim. Requires lefthook to be installed on the system.
- **`pre-check` command:** `lefthook run pre-commit` -- lets developers manually trigger what pre-commit would do without actually committing. Useful for CI or manual verification.
- **Justfile syntax:** Just uses a Makefile-like syntax but with some differences. Recipe names use hyphens, not underscores. Dependencies use `: dependency-name` syntax. Multi-line shell blocks need `#!/usr/bin/env bash` shebang or use `&&` chaining.
- **Existing targets to preserve:** `test`, `build`, `dev`, `run`, `sessions`, `attach-all`, `sonnet`, `opus`, `haiku`, `c_session`, `g_session`, `update`, `upgrade`, `repomixer`. Do NOT modify any existing target.

### Risks & Edge Cases
- **`test-all` failure modes:** If Go tests fail (the `test` dependency), the frontend tests should NOT run. Just's dependency behavior handles this -- if the dependency recipe fails, the parent recipe doesn't execute.
- **lefthook not installed:** `just hooks-install` will fail with "command not found" if lefthook is not on PATH. The error message from the shell is clear enough. No need to add a check.
- **`npx vitest` without node_modules:** If `frontend/node_modules` doesn't exist, `npx vitest run` will fail. The error message is descriptive. The existing `build` target runs `cd frontend && npm install` first -- `test-all` does not do this because tests should assume dependencies are installed.
- **Target naming conventions:** The existing justfile uses underscores for some targets (`c_session`, `g_session`) and hyphens for multi-word targets (`attach-all`). New targets use hyphens consistently: `test-all`, `test-cover`, `hooks-install`, `pre-check`.

### Reference Files
- `justfile` -- current targets (lines 1-112), shows recipe syntax patterns
- `scripts/check-coverage.sh` -- from Story 1, invoked by `test-cover`
- `lefthook.yml` -- from Story 4, invoked by `pre-check` and `hooks-install`

## Acceptance Criteria

AC-1: `just test-all` runs Go and frontend tests
- Given the justfile has a `test-all` target
- When a developer runs `just test-all`
- Then Go tests run first via the existing `test` target dependency
- And frontend tests run via `cd frontend && npx vitest run`
- And the command fails if either test suite fails

AC-2: `just test-cover` runs coverage enforcement
- Given the justfile has a `test-cover` target
- When a developer runs `just test-cover`
- Then `bash scripts/check-coverage.sh` executes
- And per-package coverage results are printed
- And the exit code reflects pass/fail status

AC-3: `just lint` runs Go static analysis
- Given the justfile has a `lint` target
- When a developer runs `just lint`
- Then `go vet ./...` executes
- And any vet warnings are printed to stderr

AC-4: `just hooks-install` installs lefthook
- Given the justfile has a `hooks-install` target
- When a developer runs `just hooks-install`
- Then `lefthook install` executes
- And `.git/hooks/pre-commit` is replaced with lefthook's shim

AC-5: `just pre-check` runs pre-commit manually
- Given the justfile has a `pre-check` target
- When a developer runs `just pre-check`
- Then `lefthook run pre-commit` executes
- And all pre-commit commands run as defined in `lefthook.yml`

AC-6: Existing justfile targets are unchanged
- Given the justfile has existing targets (test, build, dev, run, etc.)
- When new targets are added
- Then all existing targets continue to work as before
- And no existing target definitions are modified

## BDD Test Scenarios

### Scenario 1: Developer runs full test suite
```gherkin
Feature: Justfile testing targets

  Scenario: Run all tests
    Given the developer has Go and Node.js dependencies installed
    When they run "just test-all"
    Then Go tests execute via "go test ./internal/... -count=1"
    And frontend tests execute via "cd frontend && npx vitest run"
    And both results are visible in the terminal output
```

### Scenario 2: Coverage check
```gherkin
  Scenario: Run coverage enforcement
    Given scripts/check-coverage.sh exists and is executable
    When the developer runs "just test-cover"
    Then per-package coverage percentages are printed
    And the exit code is 0 when all thresholds are met
```

### Scenario 3: Lint check
```gherkin
  Scenario: Run Go linting
    Given the Go codebase compiles cleanly
    When the developer runs "just lint"
    Then "go vet ./..." runs and reports any issues
    And the exit code is 0 when no issues are found
```

### Scenario 4: Existing targets unaffected
```gherkin
  Scenario: Existing targets still work
    Given new targets have been added to the justfile
    When the developer runs "just test"
    Then the original test target runs unchanged
    And "go test ./internal/... -count=1" executes
```

## Tasks / Subtasks

- [ ] Task 1: Add test and coverage targets (AC: AC-1, AC-2, AC-3, AC-6)
  - [ ] Subtask 1a: Add `test-all: test` target with `cd frontend && npx vitest run` body
  - [ ] Subtask 1b: Add `test-cover` target with `bash scripts/check-coverage.sh` body
  - [ ] Subtask 1c: Add `lint` target with `go vet ./...` body
  - [ ] Subtask 1d: Add comments above each new target explaining its purpose

- [ ] Task 2: Add hook management targets (AC: AC-4, AC-5, AC-6)
  - [ ] Subtask 2a: Add `hooks-install` target with `lefthook install` body
  - [ ] Subtask 2b: Add `pre-check` target with `lefthook run pre-commit` body

- [ ] Task 3: Verify all targets (AC: AC-1 through AC-6)
  - [ ] Subtask 3a: Run `just test-all` and confirm Go + frontend tests execute
  - [ ] Subtask 3b: Run `just test-cover` and confirm coverage output
  - [ ] Subtask 3c: Run `just lint` and confirm vet output
  - [ ] Subtask 3d: Run existing `just test` and confirm it's unchanged

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] `just test-all` runs Go and frontend tests
- [ ] `just test-cover` runs coverage enforcement
- [ ] `just lint` runs go vet
- [ ] `just hooks-install` installs lefthook
- [ ] `just pre-check` runs pre-commit hooks
- [ ] All existing targets remain functional
- [ ] `/simplify` run on modified justfile
