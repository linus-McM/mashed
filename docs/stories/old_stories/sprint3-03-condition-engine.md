# Story: sprint3-03 -- Condition Evaluation Engine

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** sprint3-01
**Status:** ready

## Description

Create a standalone condition evaluation engine that can test captured terminal output against various criteria (contains, not-contains, regex, exit code, file existence). This engine is used by condition nodes, loop-until nodes, and any future control flow that needs to make decisions based on upstream node output. The engine is pure logic with no executor or tmux dependencies, making it highly testable.

## Developer Notes

### Architecture
- **New file:** `internal/bmad/condition.go`
  - `ConditionType` string enum with 6 variants: `CondContains`, `CondNotContains`, `CondRegex`, `CondExitCode`, `CondFileExists`, `CondAlways`
  - `Condition` struct with `Type ConditionType`, `Pattern string`, `SourceNode string`
  - `func (c *Condition) Evaluate(nodeOutputs map[string]string, repoPath string) bool` -- main evaluation method
  - `func ParseCondition(configJSON string) (*Condition, error)` -- deserialize from `node.Config["condition"]`
- **New file:** `internal/bmad/condition_test.go`
  - Table-driven tests for each condition type
  - Edge cases: empty output, invalid regex, missing source node, empty pattern

### Technical Considerations
- `Evaluate()` looks up `nodeOutputs[c.SourceNode]` to get the text to evaluate against. If `SourceNode` is empty or not found in the map, `contains`/`regex`/`exitCode` return `false`, `notContains` returns `true`, `always` returns `true`.
- `CondRegex` uses `regexp.MatchString(c.Pattern, output)`. If the regex is invalid, return `false` (do not panic). Consider pre-compiling in `ParseCondition` and storing the compiled regex, but for simplicity, compile on each eval is acceptable given the low frequency.
- `CondExitCode` checks if the output contains "exit code: N" or similar patterns. Pattern field holds the expected code (e.g., "0"). This is a best-effort heuristic -- tmux output may not have a clean exit code.
- `CondFileExists` uses `os.Stat(filepath.Join(repoPath, c.Pattern))`. The pattern is a relative path from repo root.
- `CondAlways` always returns `true` -- used for unconditional branching or default edges.
- Conditions are stored as JSON in `node.Config["condition"]`. Example: `{"type":"contains","pattern":"SUCCESS","sourceNode":"node-1"}`

### Risks & Edge Cases
- **Invalid regex pattern**: `regexp.Compile` may fail. Return `false`, do not propagate error from `Evaluate()` (it returns bool, not error). Log the invalid pattern.
- **Missing source node output**: Downstream condition node may evaluate before upstream completes (shouldn't happen with ready-set algorithm, but defensive coding). Return `false`.
- **Empty pattern**: `contains("")` is always true in Go's `strings.Contains`. Document this behavior.
- **File path traversal**: `CondFileExists` joins repoPath + pattern. Pattern should not contain `..`. Validate in `ParseCondition`.

### Reference Files
- `internal/bmad/types.go` -- existing type patterns, sentinel errors
- `internal/bmad/executor_test.go` -- table-driven test patterns
- `internal/bmad/sprint.go` -- example of `json.Unmarshal` usage for parsing

### Skills
- `/golang-testing` for table-driven condition tests
- `/golang-error-handling` for parse errors and sentinel errors
- `/simplify` mandatory

## Acceptance Criteria

AC-1: CondContains evaluates correctly
- Given a condition with type "contains" and pattern "SUCCESS"
- When evaluated against node output "Build completed: SUCCESS"
- Then the result is `true`
- And when evaluated against "Build failed: ERROR"
- Then the result is `false`

AC-2: CondNotContains evaluates correctly
- Given a condition with type "notContains" and pattern "ERROR"
- When evaluated against node output "Build completed: SUCCESS"
- Then the result is `true`
- And when evaluated against "Build failed: ERROR"
- Then the result is `false`

AC-3: CondRegex evaluates correctly
- Given a condition with type "regex" and pattern `"test.*passed"`
- When evaluated against "All test cases passed"
- Then the result is `true`
- And when evaluated against an invalid regex pattern `"[invalid"`
- Then the result is `false` (no panic)

AC-4: CondFileExists evaluates correctly
- Given a condition with type "fileExists" and pattern "output/report.md"
- When evaluated with a repoPath where the file exists
- Then the result is `true`
- And when the file does not exist
- Then the result is `false`

AC-5: CondAlways returns true
- Given a condition with type "always"
- When evaluated against any output (including empty)
- Then the result is `true`

AC-6: ParseCondition deserializes from JSON
- Given a JSON string `{"type":"contains","pattern":"SUCCESS","sourceNode":"node-1"}`
- When `ParseCondition()` is called
- Then it returns a `Condition` with `Type: CondContains`, `Pattern: "SUCCESS"`, `SourceNode: "node-1"`
- And when given invalid JSON, it returns an error

AC-7: Missing source node returns safe default
- Given a condition with sourceNode "nonexistent"
- When evaluated against a nodeOutputs map that lacks "nonexistent"
- Then contains/regex/exitCode return `false`
- And notContains returns `true`
- And always returns `true`

## BDD Test Scenarios

### Scenario 1: Contains condition

```gherkin
Feature: Condition evaluation engine

  Scenario Outline: Contains condition
    Given a condition of type "contains" with pattern "<pattern>"
    And nodeOutputs has source node output "<output>"
    When the condition is evaluated
    Then the result is <expected>

    Examples:
      | pattern  | output                    | expected |
      | SUCCESS  | Build completed: SUCCESS  | true     |
      | SUCCESS  | Build failed: ERROR       | false    |
      |          | any output                | true     |
      | needle   |                           | false    |
```

### Scenario 2: Regex condition

```gherkin
  Scenario: Valid regex matches output
    Given a condition of type "regex" with pattern "v\\d+\\.\\d+\\.\\d+"
    And nodeOutputs has source node output "Released v1.2.3 successfully"
    When the condition is evaluated
    Then the result is true

  Scenario: Invalid regex returns false without panic
    Given a condition of type "regex" with pattern "[unclosed"
    And nodeOutputs has source node output "any text"
    When the condition is evaluated
    Then the result is false
    And no panic occurs
```

### Scenario 3: FileExists condition

```gherkin
  Scenario: File exists at repo path
    Given a condition of type "fileExists" with pattern "output/report.md"
    And the file exists at "{repoPath}/output/report.md"
    When the condition is evaluated with repoPath
    Then the result is true

  Scenario: File does not exist
    Given a condition of type "fileExists" with pattern "missing/file.txt"
    And no such file exists
    When the condition is evaluated
    Then the result is false
```

### Scenario 4: Always condition

```gherkin
  Scenario: Always returns true regardless of output
    Given a condition of type "always"
    When the condition is evaluated with empty output
    Then the result is true
    When the condition is evaluated with non-empty output
    Then the result is true
```

### Scenario 5: ParseCondition

```gherkin
  Scenario: Valid JSON parses correctly
    Given JSON string '{"type":"regex","pattern":"ok","sourceNode":"n1"}'
    When ParseCondition is called
    Then a Condition is returned with Type=CondRegex, Pattern="ok", SourceNode="n1"

  Scenario: Invalid JSON returns error
    Given JSON string '{invalid'
    When ParseCondition is called
    Then an error is returned

  Scenario: Path traversal in fileExists is rejected
    Given JSON string '{"type":"fileExists","pattern":"../../etc/passwd","sourceNode":"n1"}'
    When ParseCondition is called
    Then an error is returned
```

## Tasks / Subtasks

- [ ] Task 1: Define condition types and struct (AC: AC-1 through AC-5)
  - [ ] Subtask 1a: Define `ConditionType` string type and 6 constants
  - [ ] Subtask 1b: Define `Condition` struct with `Type`, `Pattern`, `SourceNode` fields and JSON tags
  - [ ] Subtask 1c: Add sentinel error `ErrInvalidCondition` to `types.go`

- [ ] Task 2: Implement Evaluate method (AC: AC-1, AC-2, AC-3, AC-4, AC-5, AC-7)
  - [ ] Subtask 2a: Implement source node output lookup with safe defaults for missing keys
  - [ ] Subtask 2b: Implement `contains` and `notContains` evaluation
  - [ ] Subtask 2c: Implement `regex` evaluation with safe compile-failure handling
  - [ ] Subtask 2d: Implement `exitCode` evaluation (heuristic pattern matching)
  - [ ] Subtask 2e: Implement `fileExists` evaluation with `os.Stat`
  - [ ] Subtask 2f: Implement `always` evaluation (trivial)

- [ ] Task 3: Implement ParseCondition (AC: AC-6)
  - [ ] Subtask 3a: JSON unmarshal into `Condition` struct
  - [ ] Subtask 3b: Validate required fields (Type must be one of the known types)
  - [ ] Subtask 3c: Validate `fileExists` pattern does not contain path traversal (`..`)

- [ ] Task 4: Write comprehensive tests (AC: all)
  - [ ] Subtask 4a: Table-driven tests for each condition type with multiple inputs
  - [ ] Subtask 4b: Edge case tests: empty output, empty pattern, invalid regex, missing source node
  - [ ] Subtask 4c: ParseCondition tests: valid JSON, invalid JSON, path traversal
  - [ ] Subtask 4d: FileExists tests using `t.TempDir()`

## Dependencies
- Depends on: sprint3-01 (NodeType constants for documentation/context, though not directly imported)
- Blocks: sprint3-04, sprint3-05, sprint3-06

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `condition.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
