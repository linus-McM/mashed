# Story: sprint3-02 -- Terminal Output Capture on Node Completion

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** sprint3-01
**Status:** ready

## Description

Capture tmux terminal output when a workflow node completes execution. This is the foundational data that all control flow nodes depend on -- conditions evaluate captured output, transforms extract data from it, and loops check it for termination criteria. Without output capture, the executor is blind to what Claude actually produced.

## Developer Notes

### Architecture
- **Modified file:** `internal/bmad/executor.go`
  - New method: `func (e *Executor) captureOutput(ctx context.Context, target string) (string, error)` -- calls `tmux capture-pane -t {target} -p -S -5000 -E -` to grab scrollback buffer
  - Modified method: `completeNode()` -- call `captureOutput()` before marking complete, store result in `execution.NodeOutputs[nodeID]`
  - Initialize `NodeOutputs` map in `StartWorkflow()` if nil: `execution.NodeOutputs = make(map[string]string)`
  - Cap output at 100KB (102400 bytes) to prevent memory bloat from long-running sessions
- **Modified file:** `internal/bmad/executor_test.go`
  - Add `captureRunner()` test helper that returns configurable output when `capture-pane` is called
  - Add tests for output capture, size capping, and capture failure graceful handling

### Technical Considerations
- `captureOutput()` is called AFTER `pane_dead == "1"` check but BEFORE `completeNode()` status update. The pane is dead but the session still exists, so `capture-pane` can read the scrollback.
- If `capture-pane` fails (e.g., tmux session already killed), log the error but do NOT fail the node. Output capture is best-effort -- the node still completed successfully.
- The `tmux capture-pane` command with `-p` flag prints to stdout. `-S -5000` starts capture from 5000 lines back. `-E -` captures to the end.
- Thread safety: `NodeOutputs` is written under `state.mu.Lock()` in `completeNode()`. Multiple concurrent nodes can safely write to different keys.
- The `CommandRunner` interface (`func(ctx, name, args...) ([]byte, error)`) already supports this -- the mock just needs to handle `capture-pane` as a new subcommand.

### Risks & Edge Cases
- **tmux session gone before capture**: The executor polls `list-panes` and sees `pane_dead == "1"`, but by the time `capture-pane` runs, the session may have been garbage-collected. Handle this gracefully -- empty output, no error.
- **Very large output**: A long-running Claude session could produce megabytes of scrollback. The 100KB cap prevents OOM. Consider truncating from the beginning (keep the tail, which has the final output).
- **Binary/non-UTF8 output**: tmux panes can contain ANSI escape codes. We store raw output -- downstream consumers (conditions, transforms) handle parsing.
- **Existing test runners**: `successRunner()` and `failRunner()` don't handle `capture-pane`. The new `captureRunner()` wraps `successRunner()` and adds capture handling. Existing tests should still pass because `capture-pane` failure is non-fatal.

### Reference Files
- `internal/bmad/executor.go` -- `executeNode()` at line 272 (where `pane_dead` check happens), `completeNode()` at line 340
- `internal/bmad/executor_test.go` -- `successRunner()` at line 68, `delayRunner()` at line 88

### Skills
- `/golang-testing` for mock command runner patterns
- `/golang-error-handling` for non-fatal capture errors
- `/simplify` mandatory

## Acceptance Criteria

AC-1: Output captured on node completion
- Given a workflow with one process node executing in tmux
- When the node completes (pane_dead == "1")
- Then `captureOutput()` is called with the node's tmux target
- And the captured output is stored in `execution.NodeOutputs[nodeID]`

AC-2: NodeOutputs map initialized
- Given a new workflow execution is started
- When `StartWorkflow()` creates the `WorkflowExecution`
- Then `NodeOutputs` is initialized as an empty `map[string]string` (not nil)

AC-3: Output capped at 100KB
- Given a tmux pane with output exceeding 100KB
- When `captureOutput()` is called
- Then the stored output is at most 102400 bytes
- And the output is truncated from the beginning (tail preserved)

AC-4: Capture failure does not fail the node
- Given a tmux session where `capture-pane` returns an error
- When the node completes
- Then the node status is still `NodeComplete`
- And `NodeOutputs[nodeID]` is empty string
- And execution continues to downstream nodes

AC-5: Multiple concurrent nodes write to NodeOutputs safely
- Given a workflow with 2 parallel nodes (same tier)
- When both nodes complete concurrently
- Then both outputs are stored in `NodeOutputs` without race conditions
- And `go test -race` passes

## BDD Test Scenarios

### Scenario 1: Successful output capture

```gherkin
Feature: Terminal output capture

  Scenario: Node output captured on completion
    Given a workflow with node "A" using process "bmad-brainstorming"
    And the command runner returns "pane_dead=1" on list-panes
    And the command runner returns "Hello from Claude\nTask complete" on capture-pane
    When the workflow executes
    Then node "A" status is "complete"
    And NodeOutputs["A"] contains "Hello from Claude"
    And NodeOutputs["A"] contains "Task complete"
```

### Scenario 2: Output size cap

```gherkin
  Scenario: Large output truncated to 100KB
    Given a workflow with node "A"
    And the command runner returns 200KB of output on capture-pane
    When the workflow executes
    Then NodeOutputs["A"] is at most 102400 bytes long
    And NodeOutputs["A"] ends with the tail of the original output
```

### Scenario 3: Capture failure graceful handling

```gherkin
  Scenario: capture-pane failure does not fail the node
    Given a workflow with node "A"
    And the command runner returns an error on capture-pane
    And the command runner returns "pane_dead=1" on list-panes
    When the workflow executes
    Then node "A" status is "complete"
    And NodeOutputs["A"] is ""
    And execution status is "complete"
```

### Scenario 4: Parallel nodes both capture output

```gherkin
  Scenario: Concurrent nodes write outputs without race
    Given a workflow with parallel nodes "A" and "B" (no edges between them)
    And the command runner returns "output-A" for node A's capture-pane
    And the command runner returns "output-B" for node B's capture-pane
    When the workflow executes with -race flag
    Then NodeOutputs["A"] is "output-A"
    And NodeOutputs["B"] is "output-B"
    And no race condition is detected
```

### Scenario 5: Backward compatibility

```gherkin
  Scenario: Existing workflows without output capture still complete
    Given a workflow with the successRunner (no capture-pane handling)
    When the workflow executes
    Then all nodes complete successfully
    And execution status is "complete"
```

## Tasks / Subtasks

- [ ] Task 1: Initialize NodeOutputs in StartWorkflow (AC: AC-2)
  - [ ] Subtask 1a: Add `NodeOutputs: make(map[string]string)` to `WorkflowExecution` creation in `StartWorkflow()`
  - [ ] Subtask 1b: Verify `GetExecution()` copy includes NodeOutputs

- [ ] Task 2: Implement captureOutput method (AC: AC-1, AC-3)
  - [ ] Subtask 2a: Add `captureOutput(ctx, target string) (string, error)` method to `Executor`
  - [ ] Subtask 2b: Call `tmux capture-pane -t {target} -p -S -5000 -E -` via `e.runCmd`
  - [ ] Subtask 2c: Implement 100KB cap with tail preservation (truncate from beginning)

- [ ] Task 3: Integrate capture into executeNode (AC: AC-1, AC-4)
  - [ ] Subtask 3a: Call `captureOutput()` in `executeNode()` after `pane_dead == "1"` and before `completeNode()`
  - [ ] Subtask 3b: Store result in `state.exec.NodeOutputs[nodeID]` under `state.mu.Lock()`
  - [ ] Subtask 3c: Log and swallow capture errors (non-fatal)

- [ ] Task 4: Write tests (AC: AC-1, AC-2, AC-3, AC-4, AC-5)
  - [ ] Subtask 4a: Create `captureRunner()` helper that wraps `successRunner()` with capture-pane handling
  - [ ] Subtask 4b: Test output capture stores correct data
  - [ ] Subtask 4c: Test 100KB cap with large output
  - [ ] Subtask 4d: Test capture failure is non-fatal
  - [ ] Subtask 4e: Test parallel nodes with `-race` flag

## Dependencies
- Depends on: sprint3-01 (NodeOutputs field on WorkflowExecution)
- Blocks: sprint3-03, sprint3-04, sprint3-05, sprint3-06, sprint3-10

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] All 13 existing executor tests still pass
