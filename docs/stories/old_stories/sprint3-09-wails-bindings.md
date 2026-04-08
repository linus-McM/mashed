# Story: sprint3-09 -- Wails Bindings: GetNodeOutput & GetControlFlowNodes

**Priority:** P2-medium
**Domain:** fullstack
**Estimated Complexity:** S
**Depends On:** sprint3-01, sprint3-02
**Status:** ready

## Description

Expose two new Wails bindings for the control flow feature: `GetNodeOutput()` to retrieve captured terminal output for a specific node, and `GetControlFlowNodes()` to provide the sidebar with the list of available control flow node types. These bindings bridge the backend execution state and type definitions to the frontend.

## Developer Notes

### Architecture
- **Modified file:** `app.go`
  - New method: `func (a *App) GetNodeOutput(execID, nodeID string) (string, error)` -- retrieves `NodeOutputs[nodeID]` from the execution state
  - New method: `func (a *App) GetControlFlowNodes() []ControlFlowNodeDef` -- returns hardcoded list of 5 control flow node definitions
  - New type in `internal/bmad/types.go`: `ControlFlowNodeDef` struct with `Type NodeType`, `Name string`, `Description string`, `Icon string`

### GetNodeOutput Implementation
```go
func (a *App) GetNodeOutput(execID, nodeID string) (string, error) {
    if a.bmadExecutor == nil {
        return "", fmt.Errorf("bmad: executor not initialized")
    }
    exec, err := a.bmadExecutor.GetExecution(execID)
    if err != nil {
        return "", err
    }
    output, ok := exec.NodeOutputs[nodeID]
    if !ok {
        return "", nil  // no output (not an error)
    }
    return output, nil
}
```

### GetControlFlowNodes Implementation
```go
func (a *App) GetControlFlowNodes() []bmad.ControlFlowNodeDef {
    return []bmad.ControlFlowNodeDef{
        {Type: bmad.NodeTypeCondition, Name: "Condition", Description: "If/else branch based on output", Icon: "GitBranch"},
        {Type: bmad.NodeTypeLoop, Name: "Loop", Description: "Repeat N times", Icon: "Repeat"},
        {Type: bmad.NodeTypeLoopUntil, Name: "Loop Until", Description: "Repeat until condition met", Icon: "Target"},
        {Type: bmad.NodeTypeTransform, Name: "Transform", Description: "Extract/transform data", Icon: "Filter"},
        {Type: bmad.NodeTypeMerge, Name: "Merge", Description: "Join branches", Icon: "GitMerge"},
    }
}
```

### Technical Considerations
- `GetNodeOutput` requires the execution to be in memory (`e.executions` map). If the execution is not found, return `ErrExecNotFound` which the frontend handles.
- `GetExecution()` already returns a copy of the execution state including `NodeOutputs`. The binding just looks up the specific node's output.
- `GetControlFlowNodes()` is pure static data -- no storage or executor dependency. Could be called on app startup to populate the sidebar.
- The `ControlFlowNodeDef` type is placed in `types.go` alongside other definition types (`ProcessDef`, `BmadAgentConfig`).
- Follow the existing binding pattern in `app.go`: nil-check on subsystem, delegate to package, return typed result.

### Wails Binding Generation
- After adding the methods, `wails dev` auto-generates TypeScript bindings in `frontend/wailsjs/go/main/App.js`
- The frontend can then import: `import { GetNodeOutput, GetControlFlowNodes } from '../../wailsjs/go/main/App.js'`

### Reference Files
- `app.go` -- existing BMAD bindings starting at line 1641 (pattern to follow)
- `internal/bmad/types.go` -- existing type definitions
- `internal/bmad/executor.go` -- `GetExecution()` at line 164

### Skills
- `/wails` for binding patterns and auto-generation
- `/simplify` mandatory

## Acceptance Criteria

AC-1: GetNodeOutput returns captured output
- Given an active execution where node "A" has completed with captured output "Hello World"
- When `GetNodeOutput(execID, "A")` is called
- Then it returns `"Hello World"` with no error

AC-2: GetNodeOutput returns empty for node without output
- Given an active execution where node "B" has no captured output
- When `GetNodeOutput(execID, "B")` is called
- Then it returns `""` with no error

AC-3: GetNodeOutput returns error for invalid execution
- Given no execution with ID "nonexistent"
- When `GetNodeOutput("nonexistent", "A")` is called
- Then it returns an error wrapping `ErrExecNotFound`

AC-4: GetControlFlowNodes returns 5 node definitions
- Given the app is running
- When `GetControlFlowNodes()` is called
- Then it returns a slice of 5 `ControlFlowNodeDef` items
- And each has a non-empty `Type`, `Name`, `Description`, and `Icon`

AC-5: ControlFlowNodeDef type exists with correct fields
- Given the `bmad` package
- When `ControlFlowNodeDef` is referenced
- Then it has fields: `Type NodeType`, `Name string`, `Description string`, `Icon string`
- And all fields have JSON tags

AC-6: Wails bindings are auto-generated
- Given the new methods exist on `*App`
- When `wails dev` or `wails generate module` runs
- Then `GetNodeOutput` and `GetControlFlowNodes` are available in the frontend TypeScript bindings

## BDD Test Scenarios

### Scenario 1: GetNodeOutput success

```gherkin
Feature: Wails bindings for control flow

  Scenario: Retrieve captured output for completed node
    Given an executor with a running execution "exec-1"
    And node "A" has NodeOutputs entry "Hello from Claude"
    When GetNodeOutput("exec-1", "A") is called via binding
    Then it returns "Hello from Claude"
    And error is nil
```

### Scenario 2: GetNodeOutput missing output

```gherkin
  Scenario: Node exists but has no captured output
    Given an executor with execution "exec-1"
    And node "B" has no entry in NodeOutputs
    When GetNodeOutput("exec-1", "B") is called
    Then it returns ""
    And error is nil
```

### Scenario 3: GetNodeOutput invalid execution

```gherkin
  Scenario: Execution does not exist
    Given no execution with ID "exec-999"
    When GetNodeOutput("exec-999", "A") is called
    Then an error is returned
    And the error message contains "not found"
```

### Scenario 4: GetControlFlowNodes static list

```gherkin
  Scenario: Returns all 5 control flow types
    Given the app is initialized
    When GetControlFlowNodes() is called
    Then 5 items are returned
    And types include "condition", "loop", "loopUntil", "transform", "merge"
    And each has a non-empty Name and Description
```

## Tasks / Subtasks

- [ ] Task 1: Add ControlFlowNodeDef type (AC: AC-5)
  - [ ] Subtask 1a: Add `ControlFlowNodeDef` struct to `internal/bmad/types.go`
  - [ ] Subtask 1b: Add JSON tags for all fields

- [ ] Task 2: Implement GetNodeOutput binding (AC: AC-1, AC-2, AC-3)
  - [ ] Subtask 2a: Add method to `app.go` following existing binding pattern
  - [ ] Subtask 2b: Nil-check executor, delegate to `GetExecution()`, look up NodeOutputs

- [ ] Task 3: Implement GetControlFlowNodes binding (AC: AC-4)
  - [ ] Subtask 3a: Add method to `app.go` returning hardcoded slice
  - [ ] Subtask 3b: Include all 5 node type definitions with icons

- [ ] Task 4: Verify binding generation (AC: AC-6)
  - [ ] Subtask 4a: Run `wails dev` or `wails generate module`
  - [ ] Subtask 4b: Verify TypeScript bindings appear in `frontend/wailsjs/go/main/App.js`

## Dependencies
- Depends on: sprint3-01 (NodeType constants, ControlFlowNodeDef struct), sprint3-02 (NodeOutputs populated by output capture)
- Blocks: sprint3-10 (output viewer modal calls GetNodeOutput)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `/simplify` run on modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] TypeScript bindings generated and importable
