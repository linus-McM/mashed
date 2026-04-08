# Story: sprint3-06 -- Transform Node Dispatch & Data Passing V2

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** sprint3-04
**Status:** ready

## Description

Implement the Transform node execution and upgrade the context string builder to pass extracted data downstream. Transform nodes read captured output from a source node, apply a regex capture group or line-range extraction, and store the result in their own output for downstream consumption. The upgraded `buildContextStringV2()` appends extracted data from Transform nodes to the context string passed to Claude sessions, turning the workflow from a flat pipeline into a data-passing orchestration engine.

## Developer Notes

### Architecture
- **Modified file:** `internal/bmad/executor.go`
  - Extended `executeControlNode()` to handle `NodeTypeTransform`
  - New method: `func (e *Executor) executeTransformNode(state, nodeIndex, nodeID)` -- reads source output, applies extraction, stores result
  - New function: `func buildContextStringV2(proc ProcessDef, nodes []WorkflowNode, nodeIndex map[string]int, nodeOutputs map[string]string) string` -- replaces `buildContextString()`. Adds extracted Transform outputs to context.
  - Merge node: already handled as no-op in sprint3-04 (completes when in-degree reaches 0)

### Transform Node Config
- `node.Config["sourceNode"]` -- ID of the upstream node whose output to read
- `node.Config["extractType"]` -- `"regex"` or `"lines"`
- `node.Config["extractPattern"]` -- regex with capture group (e.g., `"version: (\\S+)"`) or line range (e.g., `"10-20"`, `"-5"` for last 5 lines)
- Transform reads `state.exec.NodeOutputs[sourceNode]`, applies extraction, stores result in `state.exec.NodeOutputs[nodeID]`
- No tmux session needed -- pure in-memory operation, completes synchronously

### Technical Considerations
- **Regex extraction**: Use `regexp.FindStringSubmatch()`. If the regex has a capture group, return the first captured group. If no capture group, return the full match. If no match, store empty string (not an error -- the transform just produces nothing).
- **Line range extraction**: Split output by `\n`, apply range. `"10-20"` returns lines 10-20 (1-indexed). `"-5"` returns last 5 lines. `"1"` returns first line. Out-of-range is clamped, not an error.
- **buildContextStringV2()**: Extends the current `buildContextString()` logic:
  1. Keep existing artifact-based context (matching inputs/outputs between processes)
  2. NEW: For each completed Transform node upstream, if it has extracted data in `NodeOutputs`, append: `" The data transform '{label}' extracted: {data}"`
  3. Data is truncated to 2000 chars max in the context string to avoid prompt bloat
- **Merge node** is already a no-op from sprint3-04. This story just needs to verify it works in test scenarios where merges follow transforms.

### Risks & Edge Cases
- **Missing source node output**: If `NodeOutputs[sourceNode]` is empty or absent, transform produces empty output. Not an error.
- **Invalid regex in extractPattern**: Return empty string, log the error. Same pattern as `condition.go`.
- **Large extracted data**: Cap at 100KB (same as raw output cap). Context string insertion caps at 2000 chars.
- **Transform chain**: Transform A reads from Process P, Transform B reads from Transform A. This works naturally -- `NodeOutputs` is keyed by node ID, and the ready-set algorithm ensures A completes before B starts.

### Reference Files
- `internal/bmad/executor.go` -- `buildContextString()` at line 366, `executeControlNode()` from sprint3-04
- `internal/bmad/condition.go` -- regex handling pattern to follow
- `internal/bmad/executor_test.go` -- test harness pattern

### Skills
- `/golang-testing` for transform extraction tests
- `/simplify` mandatory

## Acceptance Criteria

AC-1: Transform node extracts data via regex
- Given a transform node with `extractType: "regex"` and `extractPattern: "version: (\\S+)"`
- When the source node output is "Released version: 1.2.3 successfully"
- Then the transform node's output is "1.2.3"
- And the transform node completes successfully

AC-2: Transform node extracts data via line range
- Given a transform node with `extractType: "lines"` and `extractPattern: "2-4"`
- When the source node output has 10 lines
- Then the transform node's output is lines 2, 3, and 4 joined by newlines

AC-3: Transform node handles missing source gracefully
- Given a transform node with `sourceNode: "nonexistent"`
- When executed
- Then the transform node completes with empty output
- And execution continues (no error)

AC-4: buildContextStringV2 includes transform data
- Given a completed transform node with extracted data "extracted_value"
- When `buildContextStringV2()` is called for a downstream process node
- Then the context string includes the extracted data
- And the extracted data is truncated to 2000 chars max

AC-5: buildContextStringV2 preserves existing artifact context
- Given completed upstream process nodes with matching input/output artifacts
- When `buildContextStringV2()` is called
- Then the existing artifact-based context strings are preserved
- And transform data is appended after artifact context

AC-6: Transform chain works (transform reads from transform)
- Given Transform A reads from Process P, Transform B reads from Transform A
- When all nodes complete in order
- Then Transform B's output is based on Transform A's output

## BDD Test Scenarios

### Scenario 1: Regex extraction

```gherkin
Feature: Transform node execution

  Scenario: Regex with capture group extracts value
    Given a transform node with extractType "regex" and extractPattern "status: (\\w+)"
    And sourceNode output is "Job status: PASSED with 0 errors"
    When the transform executes
    Then NodeOutputs[transformID] is "PASSED"

  Scenario: Regex without capture group returns full match
    Given a transform node with extractType "regex" and extractPattern "\\d+\\.\\d+\\.\\d+"
    And sourceNode output is "Version 2.5.1 released"
    When the transform executes
    Then NodeOutputs[transformID] is "2.5.1"

  Scenario: Regex with no match returns empty string
    Given a transform node with extractType "regex" and extractPattern "NOTFOUND"
    And sourceNode output is "nothing matches here"
    When the transform executes
    Then NodeOutputs[transformID] is ""
```

### Scenario 2: Line range extraction

```gherkin
  Scenario: Line range "2-4" extracts lines 2 through 4
    Given a transform node with extractType "lines" and extractPattern "2-4"
    And sourceNode output is "line1\nline2\nline3\nline4\nline5"
    When the transform executes
    Then NodeOutputs[transformID] is "line2\nline3\nline4"

  Scenario: Line range "-3" extracts last 3 lines
    Given sourceNode output is "a\nb\nc\nd\ne"
    And extractPattern is "-3"
    When the transform executes
    Then NodeOutputs[transformID] is "c\nd\ne"

  Scenario: Line range beyond output length is clamped
    Given sourceNode output is "only\ntwo"
    And extractPattern is "1-100"
    When the transform executes
    Then NodeOutputs[transformID] is "only\ntwo"
```

### Scenario 3: Missing source node

```gherkin
  Scenario: Transform with nonexistent source node produces empty output
    Given a transform node with sourceNode "missing-node"
    When the transform executes
    Then NodeOutputs[transformID] is ""
    And the transform node status is "complete"
```

### Scenario 4: Context string v2

```gherkin
  Scenario: buildContextStringV2 includes transform data
    Given a completed transform node "T1" with label "Extract Version" and output "1.2.3"
    And a downstream process node that lists T1 as upstream
    When buildContextStringV2 is called for the downstream node
    Then the context string contains "The data transform 'Extract Version' extracted: 1.2.3"

  Scenario: buildContextStringV2 truncates long transform data
    Given a completed transform node with 5000 chars of output
    When buildContextStringV2 is called
    Then the transform data in context string is at most 2000 chars
```

## Tasks / Subtasks

- [ ] Task 1: Implement transform node execution (AC: AC-1, AC-2, AC-3)
  - [ ] Subtask 1a: Add `executeTransformNode()` method to `Executor`
  - [ ] Subtask 1b: Implement regex extraction with capture group support
  - [ ] Subtask 1c: Implement line range extraction with clamping
  - [ ] Subtask 1d: Handle missing source node gracefully (empty output, complete node)

- [ ] Task 2: Implement buildContextStringV2 (AC: AC-4, AC-5)
  - [ ] Subtask 2a: Copy and extend `buildContextString()` to accept `nodeOutputs` parameter
  - [ ] Subtask 2b: Add transform data inclusion with 2000-char truncation
  - [ ] Subtask 2c: Update `executeNode()` call site to use `buildContextStringV2()` with NodeOutputs
  - [ ] Subtask 2d: Remove old `buildContextString()` function (or keep as wrapper)

- [ ] Task 3: Wire transform dispatch into executeControlNode (AC: AC-1, AC-6)
  - [ ] Subtask 3a: Add `NodeTypeTransform` case to `executeControlNode()` switch
  - [ ] Subtask 3b: Verify transform chain (transform reading from transform) works

- [ ] Task 4: Write tests (AC: all)
  - [ ] Subtask 4a: Table-driven tests for regex extraction (with/without capture group, no match, invalid regex)
  - [ ] Subtask 4b: Table-driven tests for line range extraction (range, last-N, single line, out of bounds)
  - [ ] Subtask 4c: Test missing source node handling
  - [ ] Subtask 4d: Test buildContextStringV2 with transform data inclusion and truncation
  - [ ] Subtask 4e: Test transform chain scenario

## Dependencies
- Depends on: sprint3-04 (dynamic executor, `executeControlNode()` dispatch point)
- Blocks: sprint3-10 (output viewer needs NodeOutputs to display)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on transform execution and buildContextStringV2
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] All existing executor tests still pass
