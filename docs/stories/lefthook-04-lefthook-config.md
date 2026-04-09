# Story 4: Lefthook Pre-commit and Pre-push Configuration

**Priority:** P0-critical
**Domain:** fullstack
**Estimated Complexity:** M
**Depends On:** lefthook-01, lefthook-02, lefthook-03
**Status:** done

## Description

Install and configure lefthook as the git hooks manager, defining pre-commit hooks (go vet, go build, changed-file Go tests, frontend build, desloppify) and pre-push hooks (full Go test suite, coverage thresholds, frontend tests). This replaces the existing `.git/hooks/pre-commit` desloppify script with a structured, parallel hook system. The desloppify scan is preserved as one command within the lefthook configuration.

## Developer Notes

### Architecture
- **New file:** `lefthook.yml` -- at repo root, defines all hook commands
- **Modified state:** `.git/hooks/pre-commit` -- replaced by lefthook's shim after `lefthook install`
- **No Go or frontend code changes** -- this story is purely hook infrastructure
- **Dependencies:** Story 1 (coverage script) must exist for `pre-push:go-coverage`, Story 2 (PTY skip guards) must exist for `-short` tests to work, Story 3 (vitest) must exist for `pre-push:frontend-test`

### Technical Considerations
- **lefthook installation:** The tool itself needs to be installed. Check if it's available via `brew install lefthook` or `go install github.com/evilmartians/lefthook@latest`. Document the prerequisite.
- **Pre-commit hooks (parallel: true):**
  - `go-vet`: `go vet ./...` -- fast static analysis, glob-triggered on `*.go`
  - `go-build`: `go build ./...` -- ensures compilation, glob-triggered on `*.go`
  - `go-test-changed`: Runs `-short` tests only on packages with staged Go files. Command:
    ```
    go test -short -count=1 $(git diff --cached --name-only -- '*.go' | xargs -I{} dirname {} | sort -u | grep -v './internal/terminal' | sed 's|^|./|' | paste -sd' ')
    ```
    The `grep -v './internal/terminal'` exclusion prevents running PTY tests even in `-short` mode to avoid any PTY-related setup overhead. The `-short` flag is a safety net.
  - `frontend-build`: `npx vite build --mode development 2>&1 | tail -5` -- verifies frontend compiles, runs from `frontend/` root
  - `desloppify`: Conditionally runs `$REPO_ROOT/.venv-desloppify/bin/desloppify scan --skip-slow` if the binary exists. Non-blocking if desloppify is not installed.
- **Pre-push hooks (parallel: true):**
  - `go-test-all`: Full `go test -short -count=1 ./internal/...` -- comprehensive but still skips PTY
  - `go-coverage`: `bash scripts/check-coverage.sh` -- runs coverage threshold enforcement from Story 1
  - `frontend-test`: `npx vitest run` -- runs all frontend tests from Story 3, from `frontend/` root
- **Glob patterns:** lefthook uses glob patterns to determine if hooks should run. If no staged files match the glob, the command is skipped entirely (good for performance).
- **`root` directive:** For frontend commands, use `root: frontend/` so the command runs with `frontend/` as cwd.
- **Existing desloppify hook:** The current `.git/hooks/pre-commit` is a shell script that runs desloppify. After `lefthook install`, lefthook replaces this file with its own shim. The desloppify logic moves into `lefthook.yml` as a command. This is a migration, not a loss of functionality.
- **`fail_text`:** Custom error messages for each command help developers understand what failed.

### Risks & Edge Cases
- **lefthook not installed:** If a developer doesn't have lefthook installed, `lefthook install` fails. The `hooks-install` just target (Story 5) handles this, but document the prerequisite.
- **Empty git diff:** When `git diff --cached --name-only -- '*.go'` returns empty (no Go files staged), the `go test` command receives no packages and may behave unexpectedly. Wrap in a conditional: `packages=$(... | paste -sd' '); [ -n "$packages" ] && go test -short -count=1 $packages || echo "No Go packages to test"`.
- **desloppify not installed:** The desloppify command checks `[ -x "$DESLOPPIFY" ]` before running. If not found, it silently passes. This is intentional.
- **Frontend node_modules missing:** If `frontend/node_modules/` doesn't exist, `npx vite build` and `npx vitest run` will fail. Developers should run `cd frontend && npm install` after cloning. This is a known project setup step.
- **Large commits:** Pre-commit runs on every commit. The parallel execution keeps total time reasonable. The `go-test-changed` command only tests affected packages, not the full suite.
- **macOS paste command:** The `paste -sd' '` command is POSIX-compliant and works on macOS. No GNU coreutils dependency.
- **Replacing .git/hooks/pre-commit:** Back up the existing hook before running `lefthook install` in case the developer needs to revert. Document this in the justfile target (Story 5).

### Reference Files
- `justfile` -- existing targets (test, build, dev) for command patterns
- `.git/hooks/pre-commit` -- current desloppify hook (will be replaced)
- `scripts/check-coverage.sh` -- from Story 1, used by `pre-push:go-coverage`
- `frontend/package.json` -- from Story 3, has `test` script for Vitest

## Acceptance Criteria

AC-1: lefthook.yml defines pre-commit hooks
- Given `lefthook.yml` exists at the repo root
- When reviewing its pre-commit section
- Then it defines commands: `go-vet`, `go-build`, `go-test-changed`, `frontend-build`, `desloppify`
- And `parallel: true` is set for the pre-commit group
- And each Go command has `glob: "*.go"` trigger
- And `frontend-build` has `glob: "frontend/**/*.{svelte,ts,js,css}"` and `root: frontend/`

AC-2: lefthook.yml defines pre-push hooks
- Given `lefthook.yml` exists at the repo root
- When reviewing its pre-push section
- Then it defines commands: `go-test-all`, `go-coverage`, `frontend-test`
- And `parallel: true` is set for the pre-push group
- And `go-coverage` runs `bash scripts/check-coverage.sh`
- And `frontend-test` runs from `root: frontend/`

AC-3: lefthook install replaces the existing pre-commit hook
- Given `.git/hooks/pre-commit` currently runs desloppify directly
- When a developer runs `lefthook install`
- Then `.git/hooks/pre-commit` is replaced with lefthook's shim
- And desloppify runs as a lefthook command within the pre-commit phase

AC-4: Pre-commit hooks run on `git commit`
- Given lefthook is installed and hooks are active
- When a developer stages Go files and runs `git commit`
- Then `go vet`, `go build`, and `go-test-changed` execute in parallel
- And the commit proceeds only if all commands pass

AC-5: go-test-changed handles empty staged Go files gracefully
- Given no Go files are staged (only frontend files changed)
- When the pre-commit hook runs
- Then `go-test-changed` is skipped (glob doesn't match)
- And `frontend-build` still runs

AC-6: Desloppify is optional and non-blocking when absent
- Given desloppify is not installed (`.venv-desloppify/bin/desloppify` does not exist)
- When the pre-commit hook runs
- Then the desloppify command passes silently
- And other hooks still execute

## BDD Test Scenarios

### Scenario 1: Pre-commit with Go changes
```gherkin
Feature: Lefthook pre-commit hooks

  Scenario: Developer commits Go changes
    Given lefthook is installed
    And the developer has staged changes to internal/bmad/storage.go
    When they run "git commit -m 'test'"
    Then lefthook runs go-vet, go-build, go-test-changed, and desloppify in parallel
    And go-test-changed runs tests for ./internal/bmad/ only
    And the commit succeeds if all pass
```

### Scenario 2: Pre-commit with frontend changes
```gherkin
  Scenario: Developer commits frontend changes only
    Given lefthook is installed
    And the developer has staged changes to frontend/src/App.svelte
    And no Go files are staged
    When they run "git commit -m 'ui update'"
    Then lefthook runs frontend-build and desloppify
    And go-vet, go-build, go-test-changed are skipped (no Go glob match)
    And the commit succeeds if frontend-build passes
```

### Scenario 3: Pre-push runs full suite
```gherkin
  Scenario: Developer pushes to remote
    Given lefthook is installed
    And the local branch has commits ready to push
    When they run "git push"
    Then lefthook runs go-test-all, go-coverage, and frontend-test in parallel
    And go-test-all runs "go test -short -count=1 ./internal/..."
    And go-coverage runs "bash scripts/check-coverage.sh"
    And frontend-test runs "npx vitest run" from frontend/
    And the push proceeds only if all pass
```

### Scenario 4: Failed pre-commit blocks commit
```gherkin
  Scenario: go vet finds issues
    Given lefthook is installed
    And staged Go code has a vet warning
    When the developer runs "git commit -m 'broken'"
    Then go-vet exits non-zero
    And the commit is blocked
    And the error output explains the vet failure
```

### Scenario 5: Desloppify not installed
```gherkin
  Scenario: Desloppify binary missing
    Given lefthook is installed
    And .venv-desloppify/bin/desloppify does not exist
    When the developer runs "git commit -m 'no desloppify'"
    Then the desloppify command exits 0 (silent pass)
    And other pre-commit commands still run
```

## Tasks / Subtasks

- [ ] Task 1: Create lefthook.yml with pre-commit hooks (AC: AC-1, AC-5, AC-6)
  - [ ] Subtask 1a: Create `lefthook.yml` at repo root with `pre-commit:` section
  - [ ] Subtask 1b: Define `go-vet` command with `glob: "*.go"` and `run: go vet ./...`
  - [ ] Subtask 1c: Define `go-build` command with `glob: "*.go"` and `run: go build ./...`
  - [ ] Subtask 1d: Define `go-test-changed` command with the `git diff --cached` pipeline, including empty-package guard
  - [ ] Subtask 1e: Define `frontend-build` command with `root: frontend/` and `glob: "frontend/**/*.{svelte,ts,js,css}"`
  - [ ] Subtask 1f: Define `desloppify` command with conditional execution and `fail_text`

- [ ] Task 2: Add pre-push hooks to lefthook.yml (AC: AC-2)
  - [ ] Subtask 2a: Add `pre-push:` section with `parallel: true`
  - [ ] Subtask 2b: Define `go-test-all`, `go-coverage`, and `frontend-test` commands
  - [ ] Subtask 2c: Set appropriate glob patterns and root directives

- [ ] Task 3: Install lefthook and verify (AC: AC-3, AC-4)
  - [ ] Subtask 3a: Run `lefthook install` to replace `.git/hooks/pre-commit`
  - [ ] Subtask 3b: Run `lefthook run pre-commit` to verify hooks execute correctly
  - [ ] Subtask 3c: Run `lefthook run pre-push` to verify push hooks execute

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] `lefthook.yml` exists at repo root with both pre-commit and pre-push sections
- [ ] `lefthook run pre-commit` executes all pre-commit commands successfully
- [ ] `lefthook run pre-push` executes all pre-push commands successfully
- [ ] Desloppify scan is preserved as a lefthook command
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `/simplify` run on `lefthook.yml`
