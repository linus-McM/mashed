# Story 1: Go Coverage Enforcement Script

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** none
**Status:** done

## Description

Create a shell script that runs `go test -coverprofile` on each Go package and enforces per-package minimum coverage thresholds. This prevents coverage regressions by exiting non-zero when any package drops below its configured minimum. The script is a prerequisite for the lefthook pre-push hook (Story 4) and the `just test-cover` target (Story 5).

## Developer Notes

### Architecture
- **New file:** `scripts/check-coverage.sh`
- **No Go code changes** -- pure shell script
- **Data flow:** For each package in the threshold table, the script runs `go test -coverprofile=/tmp/cover-{pkg}.out -count=1 ./{pkg}/...`, parses `go tool cover -func=/tmp/cover-{pkg}.out` for the `total:` line, extracts the percentage, and compares against the threshold.

### Technical Considerations
- **Threshold table** (hardcoded associative array or case statement):
  - `internal/bmad` = 85%
  - `internal/agent` = 55%
  - `internal/scanner` = 30%
  - `internal/git` = 30%
- **Excluded packages:** `internal/domain` (types-only, no logic), `internal/explain` (0% currently, no tests yet), `internal/terminal` (PTY tests require `-short` skip, handled by Story 2)
- **Parsing:** The `go tool cover -func` output ends with a line like `total:	(statements)	90.6%`. Use `grep '^total:' | awk '{print $NF}' | tr -d '%'` to extract the float.
- **Comparison:** Use `awk` or `bc` for float comparison (bash cannot do float math natively). Example: `echo "$actual < $threshold" | bc -l` returns `1` if below.
- **Exit behavior:** Accumulate all failures, print a summary, then `exit 1` if any failed. Do not fail-fast on the first package.
- **`-short` flag:** The script should pass `-short` to `go test` so PTY-dependent tests are skipped during coverage runs in constrained environments.
- **Temp files:** Use a temp directory (`mktemp -d`) for cover profiles, clean up on exit via `trap`.

### Risks & Edge Cases
- **Package with no test files:** `go test -coverprofile` exits 0 but prints `?	package [no test files]`. The `go tool cover -func` will fail on the empty profile. Detect this case and report 0% or skip gracefully.
- **bc not available:** macOS ships with `bc`, but if missing, fall back to `awk` comparison: `awk "BEGIN {exit ($actual < $threshold) ? 0 : 1}"` (inverted logic -- exit 0 means true).
- **Permission bits:** Script must be `chmod +x`.
- **Relative paths:** Script should work when run from repo root (lefthook runs from repo root by default).

### Reference Files
- `justfile` -- existing `test` target (line 39-40) for the `go test` invocation pattern
- `cover.out` -- existing coverage file in repo root (currently tracked but should be gitignored)

## Acceptance Criteria

AC-1: Script runs coverage for all configured packages
- Given the script `scripts/check-coverage.sh` is executable
- When a developer runs `bash scripts/check-coverage.sh` from the repo root
- Then `go test -coverprofile` is executed for each of `internal/bmad`, `internal/agent`, `internal/scanner`, `internal/git`
- And the coverage percentage for each package is printed to stdout

AC-2: Script exits zero when all packages meet thresholds
- Given all configured packages have coverage at or above their thresholds
- When the script runs
- Then it prints a PASS summary for each package showing actual vs. threshold
- And exits with code 0

AC-3: Script exits non-zero when any package is below threshold
- Given one or more packages have coverage below their threshold
- When the script runs
- Then it prints a FAIL line for each failing package showing actual vs. threshold
- And exits with code 1 after checking all packages (no fail-fast)

AC-4: Script handles packages with no test files
- Given a package listed in the threshold table has no `_test.go` files
- When the script runs
- Then it reports 0% coverage for that package (or a clear "no tests" message)
- And compares against the threshold normally

AC-5: Script is idempotent and cleans up temp files
- Given the script creates temporary coverage profile files
- When the script completes (success or failure)
- Then all temporary files are removed via a trap handler
- And no leftover `/tmp/cover-*.out` files remain

## BDD Test Scenarios

### Scenario 1: All packages pass thresholds
```gherkin
Feature: Coverage threshold enforcement

  Scenario: All packages above threshold
    Given the Go packages have current coverage levels
    And internal/bmad has 90.6% coverage (threshold 85%)
    And internal/agent has 61.3% coverage (threshold 55%)
    And internal/scanner has 35.2% coverage (threshold 30%)
    And internal/git has 32.8% coverage (threshold 30%)
    When the developer runs "bash scripts/check-coverage.sh"
    Then stdout contains "PASS" for each of the 4 packages
    And the exit code is 0
```

### Scenario 2: One package below threshold
```gherkin
  Scenario: One package fails threshold
    Given internal/bmad coverage has dropped to 80% (threshold 85%)
    And all other packages are above their thresholds
    When the developer runs "bash scripts/check-coverage.sh"
    Then stdout contains "FAIL" for internal/bmad showing "80.0% < 85%"
    And stdout contains "PASS" for the other 3 packages
    And the exit code is 1
```

### Scenario 3: Package with no test files
```gherkin
  Scenario: Package has no test files
    Given internal/explain has no _test.go files
    And internal/explain is NOT in the threshold table
    When the developer runs "bash scripts/check-coverage.sh"
    Then internal/explain is not checked
    And the exit code reflects only configured packages
```

### Scenario 4: Script cleanup on failure
```gherkin
  Scenario: Temporary files are cleaned up
    Given the script creates temporary coverage profiles in a temp directory
    When the script finishes (regardless of pass/fail)
    Then the temp directory and all profile files are deleted
```

## Tasks / Subtasks

- [ ] Task 1: Create coverage enforcement script (AC: AC-1, AC-2, AC-3, AC-4, AC-5)
  - [ ] Subtask 1a: Create `scripts/check-coverage.sh` with shebang, threshold table (associative array), and temp directory setup with trap cleanup
  - [ ] Subtask 1b: Implement per-package coverage collection loop: `go test -short -coverprofile -count=1`, parse `go tool cover -func` total line, handle no-test-files case
  - [ ] Subtask 1c: Implement threshold comparison using `awk` or `bc`, accumulate pass/fail results, print summary, exit appropriately
  - [ ] Subtask 1d: Make script executable (`chmod +x`), test manually from repo root

- [ ] Task 2: Verify script behavior (AC: AC-2, AC-3)
  - [ ] Subtask 2a: Run script and confirm all current packages pass their conservative thresholds
  - [ ] Subtask 2b: Temporarily lower a threshold to verify FAIL output format and non-zero exit

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] Script is executable and runs from repo root
- [ ] `bash scripts/check-coverage.sh` exits 0 on current codebase
- [ ] Temporary files are cleaned up after execution
- [ ] `/simplify` run on the script
- [ ] Script handles edge cases (no test files, missing packages)
