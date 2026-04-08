# Go Backend: Workflow Executor (DAG Walker & tmux Spawner)

**Story ID**: bmad-04
**Status**: ready
**Priority**: P0
**Depends On**: bmad-02, bmad-03

## Description

Implement the workflow execution engine that topologically sorts a workflow DAG, walks it node-by-node, spawns Claude CLI sessions via tmux for each process, monitors completion, and emits real-time status events. This is the core runtime that makes workflows executable rather than just visual. It supports start/pause/resume/stop lifecycle, parallel execution of independent branches, and artifact context passing between sequential nodes.

## Developer Notes

### Architecture

**New files:**
- `internal/bmad/executor.go` -- Executor struct, DAG walking, tmux spawning, event emission
- `internal/bmad/executor_test.go` -- Unit tests with mock emitEvent and mock tmux commands

**Key struct:**
```go
type Executor struct {
    storage    *Storage
    ctx        context.Context
    cancel     context.CancelFunc
    emitEvent  func(string, interface{})  // wraps runtime.EventsEmit
    executions map[string]*WorkflowExecution
    mu         sync.RWMutex
}
```

**Public API:**
```go
func NewExecutor(storage *Storage, emitEvent func(string, interface{})) *Executor
func (e *Executor) StartWorkflow(workflowID, repoPath, model string) (*WorkflowExecution, error)
func (e *Executor) PauseWorkflow(execID string) error
func (e *Executor) ResumeWorkflow(execID string) error
func (e *Executor) StopWorkflow(execID string) error
func (e *Executor) GetExecution(execID string) (*WorkflowExecution, error)
```

**Sentinel errors:**
```go
var (
    ErrExecNotFound    = errors.New("bmad: execution not found")
    ErrExecNotRunning  = errors.New("bmad: execution not running")
    ErrExecNotPaused   = errors.New("bmad: execution not paused")
    ErrCyclicWorkflow  = errors.New("bmad: workflow contains a cycle")
)
```

### DAG Execution Logic

1. **Topological sort**: Kahn's algorithm on the workflow edges. Detect cycles (return `ErrCyclicWorkflow`).
2. **Ready set**: Nodes whose upstream dependencies are all `NodeComplete`. Initially, nodes with no incoming edges.
3. **Execution loop** (runs in a goroutine per execution):
   - While ready set is non-empty and status is `ExecRunning`:
     - For each ready node: spawn in its own goroutine
     - Build command: `claude --dangerously-skip-permissions --model {model} "use {skillName}"`
     - If node has upstream artifacts, append context string
     - Call `SpawnAgentWithCommand` pattern (exec `tmux new-session -d -s {name} -c {repoPath} {command}`)
     - Store tmux target on the node
     - Emit `bmad:node:status` event: `{execID, nodeID, status: "running", tmuxTarget}`
   - Monitor: poll `tmux list-panes -t {target} -F "#{pane_dead}"` every 3 seconds
   - On completion: check exit code, set `NodeComplete` or `NodeFailed`
   - Emit `bmad:node:status` with updated status
   - On failure: set `ExecPaused`, emit `bmad:execution:status` -- user decides to retry/skip
4. **Parallel execution**: Nodes with independent inputs run concurrently. The executor spawns a goroutine per node, uses a `sync.WaitGroup` to track completion within each "tier" of the topological sort.

### tmux Command Construction

Follow the exact pattern from `app.go` `SpawnAgentWithCommand` (lines 792-814):
```go
sessionName := fmt.Sprintf("bmad-%s-%d", nodeID, time.Now().Unix())
tmuxCmd := exec.CommandContext(ctx, "tmux", "new-session", "-d",
    "-s", sessionName,
    "-c", repoPath,
    command,
)
```

### Artifact Context Passing

When a node completes and downstream nodes consume its outputs:
```go
// Build context string for downstream node
contextParts := []string{}
for _, input := range processDef.Inputs {
    // Find upstream node that produces this artifact
    // Append: "The upstream process produced {artifact} -- use it as input."
}
command := fmt.Sprintf(`claude --dangerously-skip-permissions --model %s "use %s%s"`,
    model, processDef.SkillName, contextString)
```

### Technical Considerations

- **Context cancellation**: `StopWorkflow` cancels the execution context, which propagates to all `exec.CommandContext` calls, killing tmux sessions.
- **Goroutine lifecycle**: Each execution runs in its own goroutine. Use `context.WithCancel` per execution. The executor's `mu` protects the `executions` map.
- **Event emission**: The `emitEvent` function wraps `runtime.EventsEmit`. It is called from goroutines, so it must be safe for concurrent use (Wails EventsEmit is goroutine-safe).
- **Polling interval**: 3 seconds for tmux pane monitoring. Use a ticker with context cancellation.
- **Execution ID generation**: Use `fmt.Sprintf("exec-%s-%d", workflowID, time.Now().UnixMilli())` for unique IDs.

### Risks & Edge Cases

- **tmux not running**: If `tmux` is not available, `exec.Command` will fail. Wrap the error: `fmt.Errorf("spawning node %s: %w", nodeID, err)`.
- **Race between pause and node completion**: A node could complete while pause is being processed. The pause should only prevent *new* nodes from starting; running nodes finish naturally.
- **Cyclic workflow**: If a user manually creates edges forming a cycle, `StartWorkflow` must detect this and return `ErrCyclicWorkflow`.
- **Stale executions**: If the app crashes mid-execution, orphaned tmux sessions remain. The executor does not auto-clean these (separate concern).
- **Test isolation**: Tests should not actually spawn tmux. Use a `commandRunner` interface or function parameter to mock `exec.Command` in tests.

### Reference Files

- `app.go` lines 792-814 -- `SpawnAgentWithCommand` pattern (tmux spawning)
- `app.go` lines 800-804 -- `exec.CommandContext` with tmux
- `internal/bmad/types.go` (from bmad-02) -- `WorkflowExecution`, `WorkflowNode`, `WorkflowNodeStatus`
- `internal/bmad/storage.go` (from bmad-03) -- `Storage.LoadWorkflow`

## Acceptance Criteria

- [ ] AC1: Given a valid workflow with 3 sequential nodes, When `StartWorkflow` is called, Then an execution is created with status `ExecRunning`, all nodes set to `NodePending`, and the first node (no incoming edges) begins execution.
- [ ] AC2: Given a running execution, When the first node completes, Then its status becomes `NodeComplete`, the `emitEvent` function is called with `"bmad:node:status"` and the node's updated state, and the next node in topological order begins execution.
- [ ] AC3: Given a running execution, When `PauseWorkflow` is called, Then the execution status becomes `ExecPaused`, no new nodes are spawned, but the currently running node finishes.
- [ ] AC4: Given a paused execution, When `ResumeWorkflow` is called, Then execution resumes from where it left off, spawning the next ready node.
- [ ] AC5: Given a running execution, When `StopWorkflow` is called, Then the execution context is canceled, running tmux sessions are killed, and status becomes `ExecFailed`.
- [ ] AC6: Given a workflow with a cycle in its edges, When `StartWorkflow` is called, Then `ErrCyclicWorkflow` is returned.
- [ ] AC7: Given a workflow with two independent branches (node A -> B and node C -> D), When execution runs, Then nodes A and C start concurrently (parallel execution).

## BDD Test Scenarios

### Scenario 1: Sequential execution

```gherkin
Feature: BMAD Workflow Execution

  Scenario: Three-node sequential workflow executes in order
    Given a workflow with nodes [A, B, C] and edges [A->B, B->C]
    And a mock command runner that succeeds immediately
    When StartWorkflow is called
    Then node A runs first
    And after A completes, node B runs
    And after B completes, node C runs
    And the execution status is ExecComplete
    And 6 bmad:node:status events were emitted (running+complete for each node)

  Scenario: Node failure pauses execution
    Given a workflow with nodes [A, B] and edges [A->B]
    And a mock command runner that fails on node A
    When StartWorkflow is called
    Then node A status becomes NodeFailed
    And execution status becomes ExecPaused
    And node B remains NodePending
```

### Scenario 2: Parallel execution

```gherkin
Feature: BMAD Parallel Execution

  Scenario: Independent branches execute concurrently
    Given a workflow with nodes [A, B, C, D] and edges [A->C, B->D]
    And a mock command runner that takes 100ms per node
    When StartWorkflow is called
    Then nodes A and B start within 50ms of each other
    And nodes C and D start after their respective parents complete
```

### Scenario 3: Lifecycle management

```gherkin
Feature: BMAD Execution Lifecycle

  Scenario: Pause prevents new node spawning
    Given a running execution with node A running and node B pending
    When PauseWorkflow is called
    Then node A continues running
    And after A completes, B does not start
    And execution status is ExecPaused

  Scenario: Resume continues after pause
    Given a paused execution where node A is complete and B is pending
    When ResumeWorkflow is called
    Then node B begins execution
    And execution status is ExecRunning

  Scenario: Stop cancels everything
    Given a running execution with node A running
    When StopWorkflow is called
    Then the execution context is canceled
    And execution status is ExecFailed

  Scenario: Cyclic workflow is rejected
    Given a workflow with edges [A->B, B->A]
    When StartWorkflow is called
    Then ErrCyclicWorkflow is returned
    And no execution is created
```

## Tasks / Subtasks

- [ ] Task 1: Implement topological sort with cycle detection (AC: AC1, AC6)
  - [ ] Subtask 1a: Implement Kahn's algorithm on `[]WorkflowEdge`
  - [ ] Subtask 1b: Return `ErrCyclicWorkflow` when cycle detected
  - [ ] Subtask 1c: Return ordered tiers (groups of nodes that can run in parallel)
- [ ] Task 2: Implement Executor struct and StartWorkflow (AC: AC1, AC2, AC7)
  - [ ] Subtask 2a: Implement `NewExecutor` constructor
  - [ ] Subtask 2b: Implement `StartWorkflow` -- load workflow, topo sort, create execution, start goroutine
  - [ ] Subtask 2c: Implement the execution loop (spawn ready nodes, monitor, advance)
  - [ ] Subtask 2d: Implement tmux command construction following `SpawnAgentWithCommand` pattern
  - [ ] Subtask 2e: Implement artifact context string building for downstream nodes
- [ ] Task 3: Implement pause/resume/stop (AC: AC3, AC4, AC5)
  - [ ] Subtask 3a: Implement `PauseWorkflow` -- set flag, stop spawning new nodes
  - [ ] Subtask 3b: Implement `ResumeWorkflow` -- clear flag, continue execution
  - [ ] Subtask 3c: Implement `StopWorkflow` -- cancel context, kill tmux sessions
- [ ] Task 4: Implement GetExecution and event emission (AC: AC2)
  - [ ] Subtask 4a: Implement `GetExecution` with proper locking
  - [ ] Subtask 4b: Emit `bmad:node:status` events on every node state change
  - [ ] Subtask 4c: Emit `bmad:execution:status` on overall state changes
- [ ] Task 5: Write tests with mock command runner (AC: AC1-AC7)
  - [ ] Subtask 5a: Create a `commandRunner` interface/function for testability
  - [ ] Subtask 5b: Test sequential execution with mock runner
  - [ ] Subtask 5c: Test parallel execution timing
  - [ ] Subtask 5d: Test pause/resume/stop lifecycle
  - [ ] Subtask 5e: Test cycle detection

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `executor.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes (critical for concurrent executor code)
- [ ] /simplify run on all new code
