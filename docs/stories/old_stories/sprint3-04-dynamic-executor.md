# Story: sprint3-04 -- Dynamic Ready-Set Executor Algorithm

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** L
**Depends On:** sprint3-01, sprint3-02, sprint3-03
**Status:** ready

## Description

Replace the static topological-tier execution algorithm with a dynamic ready-set algorithm that supports branching and selective edge activation. The current executor runs `topoSort()` to produce fixed tiers, then executes tier-by-tier. The new algorithm dynamically computes which nodes are ready after each completion, enabling conditional branching (only activate edges matching a condition result) and future loop support. All 13 existing executor tests must continue to pass.

## Developer Notes

### Architecture
- **Modified file:** `internal/bmad/executor.go`
  - New type: `execState` gains fields `inDegree map[string]int`, `outEdges map[string][]WorkflowEdge`, `allEdges []WorkflowEdge`
  - Replace `run(ctx, state, tiers, repoPath, model)` with `runDynamic(ctx, state, repoPath, model)`
  - `StartWorkflow()` still calls `topoSort()` for cycle detection, but passes edges to `runDynamic()` instead of tiers
  - New method: `computeInDegrees(nodes, edges)` -- builds initial in-degree map (same as topoSort but without consuming)
  - New method: `readyNodes(state)` -- returns all node IDs with `inDegree == 0` and status `NodePending`
  - New method: `activeOutEdges(node WorkflowNode, result string) []WorkflowEdge` -- for condition nodes, filters by `edge.SourceHandle == result`; for all others, returns all outbound edges
  - New method: `executeControlNode(ctx, state, nodeIndex, nodeID, repoPath, model)` -- dispatches condition/merge nodes. Loop dispatch is sprint3-05.
  - Modified `run()` algorithm:
    1. Build in-degree map and outEdges map from `state.allEdges`
    2. Loop: find ready set (inDegree 0 + pending)
    3. Execute all ready nodes in parallel (process nodes via `executeNode()`, control nodes via `executeControlNode()`)
    4. On each completion: call `activeOutEdges()`, decrement in-degrees of targets
    5. If ready set is empty and any node still pending, execution is stuck (should not happen with valid DAGs)
    6. When all nodes are complete/failed/skipped, mark execution complete

### Technical Considerations
- **Concurrency model**: Same as current -- `sync.WaitGroup` per ready-set batch. All nodes in a ready set execute in parallel, then we compute the next ready set. This is safe because in-degree modification happens only after `wg.Wait()`.
- **Condition node dispatch**: When a condition node completes, `activeOutEdges()` returns only edges where `SourceHandle` matches the condition result (`"true"` or `"false"`). Edges NOT matching have their targets' in-degrees left unchanged, effectively skipping those branches.
- **Merge node dispatch**: A merge node is a no-op -- it completes immediately when all inbound edges are satisfied (in-degree reaches 0). No tmux session, no command execution.
- **Transform node**: Handled in sprint3-06. For now, `executeControlNode()` can log a warning for unhandled types.
- **Skipping unreachable nodes**: After execution finishes, any nodes still in `NodePending` should be marked `NodeSkipped`. This happens when a condition branch is not taken.
- **Backward compat**: Process nodes (including those with empty `NodeType`) are routed to `executeNode()` unchanged. The `activeOutEdges()` method returns all outbound edges for process nodes.
- **Existing `topoSort()`**: Keep it -- used for cycle detection in `StartWorkflow()`. Remove tier-based execution.

### Risks & Edge Cases
- **Race condition on in-degree map**: In-degrees are modified only after `wg.Wait()` (all parallel nodes finished). No concurrent writes.
- **All nodes skipped**: If a condition skips ALL downstream branches, the execution should still complete (not hang).
- **Diamond with condition**: Node A (condition) -> B (true) and C (false), both -> D (merge). If A evaluates true, B runs, C is skipped. D should wait for B only. In-degree of D is 2 initially, but C's edge never decrements it. Solution: when a node is skipped (not activated), immediately decrement all its downstream in-degrees recursively.
- **Pause/resume**: Must work identically to current behavior. The ready-set loop checks `waitWhilePaused()` before each batch.
- **Failure handling**: Same as current -- a node failure pauses execution. The ready-set loop checks for failures after each batch.

### Reference Files
- `internal/bmad/executor.go` -- current `run()` at line 190, `topoSort()` at line 403
- `internal/bmad/executor_test.go` -- all 13 existing tests that must pass
- `internal/bmad/types.go` -- `WorkflowNode`, `WorkflowEdge` with new fields from sprint3-01

### Skills
- `/golang-testing` for executor concurrency tests
- `/golang-error-handling` for execution state errors
- `/simplify` mandatory

## Acceptance Criteria

AC-1: Ready-set algorithm produces same results as tier-based for linear DAGs
- Given a sequential workflow A -> B -> C (no conditions)
- When executed with the new algorithm
- Then nodes execute in order A, B, C
- And all nodes complete
- And execution status is "complete"

AC-2: Ready-set algorithm handles parallel branches
- Given a diamond workflow A -> {B, C} -> D
- When executed with the new algorithm
- Then B and C execute in parallel after A
- And D executes after both B and C complete
- And execution status is "complete"

AC-3: Condition node activates correct branch
- Given a workflow: A (process) -> B (condition) -> C (sourceHandle="true") and D (sourceHandle="false")
- When B evaluates to "true"
- Then C is executed and D is skipped
- And skipped nodes have status `NodeSkipped`

AC-4: Merge node completes when all active inbound edges are satisfied
- Given a diamond: A (condition, true) -> B, A (false) -> C, B -> D (merge), C -> D (merge)
- When A evaluates true (B runs, C skipped)
- Then D executes after B completes (does not wait for skipped C)

AC-5: Unreachable nodes marked as skipped
- Given a condition workflow where one branch is not taken
- When execution completes
- Then nodes on the inactive branch have status `NodeSkipped`

AC-6: Pause/resume works with dynamic executor
- Given a multi-node workflow executing with the new algorithm
- When paused mid-execution
- Then running nodes complete but no new nodes start
- And when resumed, remaining nodes execute

AC-7: Node failure pauses execution
- Given a workflow where a node fails
- When the failure is detected
- Then execution pauses (same behavior as current)
- And downstream nodes remain pending

AC-8: All 13 existing executor tests pass
- Given the refactored executor
- When `go test ./internal/bmad/... -race` is run
- Then all existing tests pass without modification

## BDD Test Scenarios

### Scenario 1: Linear DAG equivalence

```gherkin
Feature: Dynamic ready-set executor

  Scenario: Sequential workflow executes identically to tier-based
    Given a workflow with nodes A -> B -> C using process nodes
    And all nodes use successRunner
    When the workflow executes
    Then nodes complete in order A, B, C
    And execution status is "complete"
```

### Scenario 2: Condition branching

```gherkin
  Scenario: Condition node takes true branch
    Given a workflow: Process(A) -> Condition(B) -> Process(C, sourceHandle=true) and Process(D, sourceHandle=false)
    And node A captures output containing "SUCCESS"
    And node B has condition type "contains" with pattern "SUCCESS" and sourceNode "A"
    When the workflow executes
    Then node A completes
    And node B completes with output "true"
    And node C completes
    And node D has status "skipped"

  Scenario: Condition node takes false branch
    Given the same workflow
    And node A captures output containing "FAILURE"
    When the workflow executes
    Then node C has status "skipped"
    And node D completes
```

### Scenario 3: Merge after conditional

```gherkin
  Scenario: Merge node fires after active branch completes
    Given a workflow: Condition(A) -> Process(B, true) -> Merge(D), Condition(A) -> Process(C, false) -> Merge(D)
    And node A evaluates to true
    When the workflow executes
    Then B completes, C is skipped
    And D completes (does not wait for skipped C)
```

### Scenario 4: All branches skipped

```gherkin
  Scenario: Execution completes even when condition skips all nodes
    Given a workflow: Process(A) -> Condition(B) -> Process(C, sourceHandle=true)
    And B evaluates to false (no false branch exists)
    When the workflow executes
    Then A completes, B completes, C is skipped
    And execution status is "complete"
```

### Scenario 5: Pause/resume with dynamic executor

```gherkin
  Scenario: Pause prevents new ready-set from executing
    Given a sequential workflow A -> B -> C using delayRunner(200ms)
    When node A starts running
    And the workflow is paused
    Then node A completes but B remains pending
    When the workflow is resumed
    Then B and C execute and complete
```

## Tasks / Subtasks

- [ ] Task 1: Add execution state fields for dynamic algorithm (AC: AC-1, AC-2)
  - [ ] Subtask 1a: Add `inDegree map[string]int`, `outEdges map[string][]WorkflowEdge`, `allEdges []WorkflowEdge` to `execState`
  - [ ] Subtask 1b: Implement `computeInDegrees(nodes, edges)` utility
  - [ ] Subtask 1c: Implement `readyNodes(state)` that returns pending nodes with inDegree 0

- [ ] Task 2: Implement `runDynamic()` main loop (AC: AC-1, AC-2, AC-5, AC-6, AC-7)
  - [ ] Subtask 2a: Replace `run()` body with ready-set loop: compute ready set, execute in parallel, decrement in-degrees
  - [ ] Subtask 2b: After each batch, mark unreachable pending nodes as `NodeSkipped` via recursive propagation
  - [ ] Subtask 2c: Integrate pause/resume checks (reuse `waitWhilePaused()`)
  - [ ] Subtask 2d: Integrate failure handling (pause on failure, same as current)

- [ ] Task 3: Implement `activeOutEdges()` for conditional edge routing (AC: AC-3, AC-4)
  - [ ] Subtask 3a: For condition nodes, filter edges by `SourceHandle == result`
  - [ ] Subtask 3b: For all other node types, return all outbound edges
  - [ ] Subtask 3c: Implement `skipBranch()` to recursively mark unreachable nodes and decrement their downstream in-degrees

- [ ] Task 4: Implement `executeControlNode()` for condition and merge (AC: AC-3, AC-4)
  - [ ] Subtask 4a: Condition dispatch: parse condition from `node.Config["condition"]`, evaluate, store result in `NodeOutputs`
  - [ ] Subtask 4b: Merge dispatch: immediately complete (no-op, handled by in-degree reaching 0)
  - [ ] Subtask 4c: Route to `executeControlNode()` when `node.EffectiveType() != NodeTypeProcess`

- [ ] Task 5: Update StartWorkflow to use dynamic executor (AC: AC-8)
  - [ ] Subtask 5a: Keep `topoSort()` call for cycle detection
  - [ ] Subtask 5b: Pass edges to `runDynamic()` instead of tiers
  - [ ] Subtask 5c: Store edges on `execState`

- [ ] Task 6: Verify backward compatibility (AC: AC-8)
  - [ ] Subtask 6a: Run all 13 existing executor tests
  - [ ] Subtask 6b: Add new tests for condition branching, merge, and skip scenarios
  - [ ] Subtask 6c: Run with `-race` flag

## Dependencies
- Depends on: sprint3-01 (NodeType, edge handles), sprint3-02 (NodeOutputs populated), sprint3-03 (condition evaluation)
- Blocks: sprint3-05, sprint3-06, sprint3-09, sprint3-10

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on modified executor code
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] All 13 existing executor tests pass unchanged
