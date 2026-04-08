# Story: sprint3-05 -- Loop and LoopUntil Execution

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** L
**Depends On:** sprint3-04
**Status:** ready

## Description

Add loop and loop-until execution semantics to the dynamic executor. Loop nodes repeat their body nodes a fixed number of times. LoopUntil nodes repeat until a condition is met or a max iteration count is reached. Both node types reset their body nodes to `pending` after each iteration, re-execute them, and track the current iteration count in `NodeOutputs`. This is the most complex control flow feature -- it introduces cycles into execution without violating DAG validation.

## Developer Notes

### Architecture
- **Modified file:** `internal/bmad/executor.go`
  - New method: `func (e *Executor) executeLoopNode(ctx, state, nodeIndex, nodeID, repoPath, model)` -- loop dispatch
  - New method: `func (e *Executor) resetLoopBody(state *execState, bodyNodeIDs []string)` -- reset body nodes to `NodePending` and restore their in-degrees
  - Extended `executeControlNode()` (from sprint3-04) to handle `NodeTypeLoop` and `NodeTypeLoopUntil`
  - Loop body node IDs stored in `node.Config["loopBodyNodes"]` as comma-separated list
  - Iteration count tracked in `state.exec.NodeOutputs[nodeID+"_iter"]` (e.g., "3")
  - Max iterations from `node.Config["maxIterations"]` (string, parsed to int, default 10)

### Technical Considerations
- **Loop mechanics**: A loop node has two outbound edge types:
  - `sourceHandle: "loop-body"` -- points to the first node of the loop body
  - `sourceHandle: "loop-exit"` -- points to the node after the loop
  - Each iteration: reset body nodes, decrement body entry node's in-degree to 0, let ready-set pick it up
  - After max iterations (or condition met for loopUntil): activate exit edge, skip body

- **resetLoopBody()**: Sets each body node back to `NodePending`, restores their in-degrees from the initial snapshot. This is safe because `wg.Wait()` guarantees all parallel body nodes finished before the loop node checks the next iteration.

- **LoopUntil condition**: Same `Condition` struct from sprint3-03. Stored in `node.Config["condition"]`. Evaluated after each iteration against the body's last node's output. If true, exit the loop.

- **In-degree snapshot**: At execution start, take a snapshot of initial in-degrees. `resetLoopBody()` restores from this snapshot, not from zero. This handles body nodes that have inbound edges from other body nodes.

- **Iteration event**: Emit `bmad:node:status` with a custom field for iteration count so the frontend can display "2 / 5".

- **Concurrency**: Loop body executes sequentially within the loop (one iteration at a time). Multiple loops in parallel are fine -- they operate on disjoint body nodes.

- **topoSort and cycle detection**: The workflow graph with loops technically has cycles (loop node -> body -> ... -> loop node implicit). BUT we do NOT add back-edges to the graph. The loop body is declared in config, not via edges. `topoSort()` sees the loop node with outbound edges to body-entry and exit, which is acyclic. The loop iteration is managed by `resetLoopBody()` outside the DAG structure.

### Risks & Edge Cases
- **Infinite loop protection**: Max iterations (default 10, max 100) prevents runaway loops. If max is exceeded, activate exit edge and log warning.
- **Loop body failure**: If any body node fails during an iteration, pause execution (same as normal failure handling). Do not continue iterating.
- **Nested loops**: Not supported in this story. Validation: a loop's body nodes cannot include another loop node. Fail at `StartWorkflow()` with a descriptive error.
- **Empty body**: If `loopBodyNodes` is empty, the loop completes immediately (0 iterations, activate exit edge).
- **Loop body nodes must be a connected subgraph**: They should form a path/DAG within the loop. Disconnected body nodes could execute out of order. Document this constraint.

### Reference Files
- `internal/bmad/executor.go` -- `executeControlNode()` from sprint3-04
- `internal/bmad/condition.go` -- `Condition.Evaluate()` from sprint3-03
- `internal/bmad/executor_test.go` -- test helpers, `delayRunner()`

### Skills
- `/golang-testing` for complex concurrent test scenarios
- `/golang-error-handling` for loop configuration errors
- `/simplify` mandatory

## Acceptance Criteria

AC-1: Loop node executes body N times
- Given a workflow with a loop node configured with `maxIterations: 3` and body nodes B, C
- When executed
- Then B and C execute 3 times each
- And the loop node's `NodeOutputs[loopID+"_iter"]` is "3" after completion
- And the exit edge is activated after iteration 3

AC-2: LoopUntil exits when condition met
- Given a workflow with a loopUntil node, condition `contains "DONE"`, and body node B
- When B produces output "DONE" on iteration 2
- Then the loop exits after iteration 2
- And `NodeOutputs[loopID+"_iter"]` is "2"
- And the exit edge is activated

AC-3: LoopUntil respects max iterations
- Given a loopUntil node with maxIterations 5 and condition that never matches
- When executed
- Then the loop exits after 5 iterations
- And the exit edge is activated

AC-4: Loop body reset restores node states
- Given loop body nodes B and C that completed in iteration 1
- When iteration 2 begins
- Then B and C are reset to `NodePending`
- And their in-degrees are restored to initial values
- And they execute again

AC-5: Loop body failure pauses execution
- Given a loop with body node B that fails on iteration 2
- When the failure occurs
- Then execution pauses
- And the loop does not continue to iteration 3

AC-6: Iteration count emitted as event
- Given a loop node executing
- When each iteration starts
- Then a `bmad:node:status` event is emitted with the current iteration count

AC-7: Empty loop body completes immediately
- Given a loop node with empty `loopBodyNodes`
- When executed
- Then the loop completes with 0 iterations
- And the exit edge is activated

## BDD Test Scenarios

### Scenario 1: Fixed-count loop

```gherkin
Feature: Loop execution

  Scenario: Loop executes body 3 times
    Given a workflow with Process(A) -> Loop(L, maxIterations=3, body=[B,C]) -> Process(D)
    And B and C are process nodes using successRunner
    When the workflow executes
    Then A completes
    And B executes 3 times (3 running events, 3 complete events)
    And C executes 3 times
    And D completes after the loop
    And NodeOutputs["L_iter"] equals "3"
```

### Scenario 2: LoopUntil early exit

```gherkin
  Scenario: LoopUntil exits when condition met
    Given a workflow with LoopUntil(L, condition=contains "DONE", maxIterations=10, body=[B])
    And B produces "working..." on iteration 1 and "DONE" on iteration 2
    When the workflow executes
    Then B executes exactly 2 times
    And NodeOutputs["L_iter"] equals "2"
    And the exit edge node completes
```

### Scenario 3: LoopUntil max iterations

```gherkin
  Scenario: LoopUntil stops at max iterations when condition never met
    Given a workflow with LoopUntil(L, condition=contains "NEVER", maxIterations=3, body=[B])
    And B always produces "still going"
    When the workflow executes
    Then B executes exactly 3 times
    And the exit edge node completes
```

### Scenario 4: Loop body failure

```gherkin
  Scenario: Body node failure pauses execution
    Given a loop with body node B that fails on iteration 2
    When the workflow executes
    Then B completes on iteration 1
    And B fails on iteration 2
    And execution status is "paused"
    And no iteration 3 occurs
```

### Scenario 5: Empty loop body

```gherkin
  Scenario: Empty loop completes immediately
    Given a loop node with loopBodyNodes="" and maxIterations=5
    When the loop node executes
    Then it completes with 0 iterations
    And the exit edge is activated
```

## Tasks / Subtasks

- [ ] Task 1: Implement resetLoopBody (AC: AC-4)
  - [ ] Subtask 1a: Parse `loopBodyNodes` from comma-separated config string
  - [ ] Subtask 1b: Reset each body node to `NodePending` status
  - [ ] Subtask 1c: Restore body node in-degrees from initial snapshot

- [ ] Task 2: Implement Loop node execution (AC: AC-1, AC-6, AC-7)
  - [ ] Subtask 2a: Parse `maxIterations` from config (default 10, cap at 100)
  - [ ] Subtask 2b: Implement iteration loop: reset body, execute ready-set, check completion
  - [ ] Subtask 2c: Track iteration count in `NodeOutputs[nodeID+"_iter"]`
  - [ ] Subtask 2d: Emit iteration count in status events
  - [ ] Subtask 2e: Activate exit edge after final iteration

- [ ] Task 3: Implement LoopUntil node execution (AC: AC-2, AC-3)
  - [ ] Subtask 3a: Parse condition from `node.Config["condition"]` using `ParseCondition()`
  - [ ] Subtask 3b: After each iteration, evaluate condition against last body node's output
  - [ ] Subtask 3c: Exit when condition is true OR max iterations reached

- [ ] Task 4: Handle edge cases (AC: AC-5, AC-7)
  - [ ] Subtask 4a: Pause on body node failure (integrate with existing failure handling)
  - [ ] Subtask 4b: Handle empty body (immediate completion)
  - [ ] Subtask 4c: Validate no nested loops at `StartWorkflow()` time

- [ ] Task 5: Write tests (AC: all)
  - [ ] Subtask 5a: Test fixed-count loop with successRunner
  - [ ] Subtask 5b: Test loopUntil with dynamic output (iteration-aware captureRunner)
  - [ ] Subtask 5c: Test max iterations cap
  - [ ] Subtask 5d: Test body failure pauses execution
  - [ ] Subtask 5e: Test empty loop body
  - [ ] Subtask 5f: Run with `-race` flag

## Dependencies
- Depends on: sprint3-04 (dynamic executor, `executeControlNode()` dispatch)
- Blocks: sprint3-10 (loop iteration display)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new loop execution code
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] All existing executor tests still pass
