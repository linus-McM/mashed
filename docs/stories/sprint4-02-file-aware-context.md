# Story 2: File-Aware Context Passing in Executor

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** sprint4-01-artifact-path-system
**Status:** done

## Description

Upgrade the executor's context string builder to pass actual file paths to downstream Claude skill sessions instead of vague hints. When an upstream process produces `PRD.md`, the downstream process's context string now says "Read the artifact 'PRD.md' from file '/repo/_bmad-output/planning-artifacts/PRD.md'" -- enabling skills to actually locate and use upstream outputs. This replaces `buildContextString` with `buildContextStringV3` and includes the migration (Phase 6b).

## Developer Notes

### Architecture
- **Modified file:** `internal/bmad/executor.go`
  - Add new function `buildContextStringV3(proc ProcessDef, nodes []WorkflowNode, nodeIndex map[string]int, nodeOutputs map[string]string, repoPath string) string`
  - Update `executeNode()` call site: change `buildContextString(...)` to `buildContextStringV3(...)` (pass `repoPath` as additional arg)
  - Remove the old `buildContextString()` function entirely (it becomes dead code)
- **Modified file:** `internal/bmad/executor_test.go`
  - Update existing `buildContextString` tests to test `buildContextStringV3`
  - Add new tests for file-path context strings (exists vs missing)
- The function signature adds `repoPath string` as a new parameter compared to the old `buildContextString`.

### Technical Considerations
- `buildContextStringV3` logic for each matched upstream artifact:
  1. Resolve artifact name to file path via `ResolveArtifactPath(outputName, repoPath)`.
  2. If path is non-empty, check if file/dir exists on disk (`os.Stat`).
  3. If exists: append `"Read the artifact '{name}' from file '{fullPath}' and use it as input."`
  4. If missing: append `"The upstream process '{procName}' should have produced '{name}' at '{fullPath}' but it was not found. Proceed with best effort."`
  5. If path is empty (unmapped like `"code"`, `"tests"`): fall back to current hint style: `"The upstream process '{procName}' produced '{name}' -- use it as input."`
- Transform data inclusion (from old `buildContextString`) must be preserved in V3 -- the `maxTransformDataLen` logic stays.
- The `executeNode` function currently calls `buildContextString(proc, nodesCopy, nodeIndex, outputsCopy)` at line ~729. Change to `buildContextStringV3(proc, nodesCopy, nodeIndex, outputsCopy, repoPath)`.
- `os.Stat` calls in the context builder are acceptable -- this runs once per node launch, not in a hot loop.

### Risks & Edge Cases
- The `_bmad-output/` directory may not exist at all in a fresh repo. The context string should degrade gracefully to "not found" messages.
- Artifact names with wildcards (`"story-*.md"`) resolve to directories. The "exists" check should use `os.Stat` on the directory path.
- The old `buildContextString` tests in `executor_test.go` will need updating. Review them first to understand current test patterns.
- Concurrency: `buildContextStringV3` is called under a lock-free snapshot (nodesCopy, outputsCopy), so `os.Stat` calls are safe.

### Reference Files
- `internal/bmad/executor.go` -- current `buildContextString` at line ~923, `executeNode` at line ~700
- `internal/bmad/artifacts.go` -- `ResolveArtifactPath` (from Story 1)
- `internal/bmad/executor_test.go` -- existing context string tests (search for `buildContextString`)
- Skill invocation: `/wails` skill for understanding the tmux command construction

## Acceptance Criteria

AC-1: Context string includes file paths for mapped artifacts
- Given a workflow where node A (Create PRD) completed and produced `PRD.md`
- And the file `_bmad-output/planning-artifacts/PRD.md` exists on disk
- When node B (Create Architecture) builds its context string via `buildContextStringV3`
- Then the context string contains `"Read the artifact 'PRD.md' from file '/repo/_bmad-output/planning-artifacts/PRD.md' and use it as input."`

AC-2: Missing artifacts produce best-effort messages
- Given a workflow where node A completed but `PRD.md` does NOT exist on disk
- When node B builds its context string via `buildContextStringV3`
- Then the context string contains `"The upstream process 'Create PRD' should have produced 'PRD.md' at '/repo/_bmad-output/planning-artifacts/PRD.md' but it was not found. Proceed with best effort."`

AC-3: Unmapped artifacts fall back to hint style
- Given a workflow where node A (Develop Story) completed with output `"code"`
- When node B (Code Review) builds its context string
- Then the context string contains `"The upstream process 'Develop Story' produced 'code' -- use it as input."` (hint style, no file path)

AC-4: Transform data is preserved in V3
- Given a completed transform node with extracted data
- When `buildContextStringV3` is called
- Then the context string includes the transform data (capped at 2000 chars) just as the old function did

AC-5: Old buildContextString is removed
- Given the executor.go file after migration
- When searching for `func buildContextString(`
- Then only `buildContextStringV3` exists (the old function is deleted)
- And `executeNode` calls `buildContextStringV3`

## BDD Test Scenarios

### Scenario 1: File-path context for existing artifacts

```gherkin
Feature: File-aware context string building

  Scenario: Upstream artifact exists on disk
    Given a temp repo directory with "_bmad-output/planning-artifacts/PRD.md" containing "# Requirements"
    And a completed upstream node for process "bmad-create-prd" (outputs: ["PRD.md"])
    And a downstream process "bmad-create-architecture" (inputs: ["PRD.md"])
    When buildContextStringV3 is called with the downstream process
    Then the result contains "Read the artifact 'PRD.md' from file"
    And the result contains "_bmad-output/planning-artifacts/PRD.md"
    And the result contains "and use it as input"

  Scenario: Upstream artifact missing from disk
    Given a temp repo directory with NO "_bmad-output/planning-artifacts/PRD.md"
    And a completed upstream node for process "bmad-create-prd" (outputs: ["PRD.md"])
    And a downstream process "bmad-create-architecture" (inputs: ["PRD.md"])
    When buildContextStringV3 is called with the downstream process
    Then the result contains "should have produced 'PRD.md'"
    And the result contains "but it was not found"
    And the result contains "Proceed with best effort"
```

### Scenario 2: Unmapped artifact fallback

```gherkin
Feature: Unmapped artifact hint fallback

  Scenario: Code artifact uses hint style
    Given a completed upstream node for process "bmad-dev-story" (outputs: ["code", "tests"])
    And a downstream process "bmad-code-review" (inputs: ["code"])
    When buildContextStringV3 is called with the downstream process
    Then the result contains "produced 'code' -- use it as input"
    And the result does NOT contain "from file"
```

### Scenario 3: Transform data preserved

```gherkin
Feature: Transform data in V3 context string

  Scenario: Transform node data included
    Given a completed transform node "tx-1" with label "Extract Summary" and output "Key findings: ..."
    And a downstream process that takes inputs
    When buildContextStringV3 is called
    Then the result contains "The data transform 'Extract Summary' extracted: Key findings: ..."

  Scenario: Transform data capped at 2000 chars
    Given a completed transform node with output longer than 2000 characters
    When buildContextStringV3 is called
    Then the transform data in the result is truncated to 2000 characters
```

### Scenario 4: No inputs needed

```gherkin
Feature: Process with no inputs

  Scenario: Root process has empty context
    Given a process "bmad-brainstorming" with no inputs
    When buildContextStringV3 is called
    Then the result is ""
```

## Tasks / Subtasks

- [ ] Task 1: Implement buildContextStringV3 (AC: AC-1, AC-2, AC-3, AC-4)
  - [ ] Subtask 1a: Create `buildContextStringV3` function with `repoPath` parameter added
  - [ ] Subtask 1b: Implement file-path resolution logic using `ResolveArtifactPath` and `os.Stat`
  - [ ] Subtask 1c: Implement fallback hint style for unmapped artifacts (empty path from resolver)
  - [ ] Subtask 1d: Preserve transform data inclusion logic from old function
- [ ] Task 2: Migrate call site and remove old function (AC: AC-5)
  - [ ] Subtask 2a: Update `executeNode()` to call `buildContextStringV3` with `repoPath` argument
  - [ ] Subtask 2b: Delete old `buildContextString` function
- [ ] Task 3: Update and add tests (AC: AC-1, AC-2, AC-3, AC-4, AC-5)
  - [ ] Subtask 3a: Update existing `buildContextString` tests to call `buildContextStringV3`
  - [ ] Subtask 3b: Add table-driven tests for file-exists, file-missing, unmapped, no-inputs scenarios using `t.TempDir()`
  - [ ] Subtask 3c: Add test verifying transform data preservation in V3

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on modified `executor.go` functions
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
