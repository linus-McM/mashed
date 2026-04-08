# Story: sprint3-01 -- Node Type System & Edge Handles

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Extend the BMAD type system to support control flow nodes alongside existing process nodes. This adds a `NodeType` discriminator to `WorkflowNode`, handle fields to `WorkflowEdge`, and a `NodeOutputs` map to `WorkflowExecution`. Backward compatibility is critical -- existing workflow JSON files without these fields must deserialize without error, and the executor must treat empty `NodeType` as `"process"`.

## Developer Notes

### Architecture
- **Modified file:** `internal/bmad/types.go`
  - Add `NodeType` type alias and constants (`NodeTypeProcess`, `NodeTypeCondition`, `NodeTypeLoop`, `NodeTypeLoopUntil`, `NodeTypeTransform`, `NodeTypeMerge`)
  - Add `NodeType NodeType` field to `WorkflowNode` with `json:"nodeType,omitempty"` tag
  - Add `EffectiveType() NodeType` method on `WorkflowNode` -- returns `NodeTypeProcess` when `NodeType` is empty string
  - Add `SourceHandle string` and `TargetHandle string` fields to `WorkflowEdge` with `json:",omitempty"` tags
  - Add `NodeOutputs map[string]string` field to `WorkflowExecution` with `json:"nodeOutputs,omitempty"` tag
- **Modified file:** `internal/bmad/executor_test.go` -- add backward compat tests
- **New file:** `internal/bmad/types_nodetype_test.go` -- focused unit tests for the type additions

### Technical Considerations
- `omitempty` on `NodeType` means existing JSON files that lack the field deserialize to `""` (Go zero value for string). The `EffectiveType()` helper normalizes this to `NodeTypeProcess`. All executor code should call `EffectiveType()`, never read `NodeType` directly.
- `SourceHandle`/`TargetHandle` on edges are already tracked by xyflow on the frontend. The frontend currently strips them during `saveWorkflow()` -- this story makes the backend ready to persist them. The frontend wiring is in sprint3-08.
- `NodeOutputs` is initialized lazily -- the executor will allocate the map when first writing to it (sprint3-02). Here we just define the field.
- No executor logic changes in this story. This is pure data model expansion.

### Risks & Edge Cases
- Existing `executor_test.go` helpers (`saveThreeNodeWorkflow`, `newHarness`) create `WorkflowNode` structs without `NodeType`. These must continue to compile and pass -- the zero value `""` is intentional.
- JSON round-trip: a workflow saved with `nodeType: "condition"` then loaded must preserve the value. A workflow saved without `nodeType` must load with empty string.
- The `Config map[string]string` field on `WorkflowNode` will later store condition configs, loop configs, etc. No changes needed here -- the map is already generic.

### Reference Files
- `internal/bmad/types.go` -- existing type definitions (lines 1-142)
- `internal/bmad/executor_test.go` -- existing test patterns, backward compat tests for StoryID (lines 471-527)
- `docs/stories/sprint2-07-sprint-aware-canvas-nodes.md` -- prior story that added `StoryID` with same backward compat approach

### Skills
- `/golang-testing` for table-driven backward compat tests
- `/golang-error-handling` for any new sentinel errors
- `/simplify` mandatory in Definition of Done

## Acceptance Criteria

AC-1: NodeType enum and constants defined
- Given the `internal/bmad/types.go` file
- When a developer imports the `bmad` package
- Then the constants `NodeTypeProcess`, `NodeTypeCondition`, `NodeTypeLoop`, `NodeTypeLoopUntil`, `NodeTypeTransform`, `NodeTypeMerge` are available
- And each constant has the expected string value (`"process"`, `"condition"`, `"loop"`, `"loopUntil"`, `"transform"`, `"merge"`)

AC-2: WorkflowNode has NodeType field with backward compat
- Given an existing JSON workflow file without a `nodeType` field
- When the JSON is unmarshaled into `WorkflowNode`
- Then `node.NodeType` is `""` (empty string)
- And `node.EffectiveType()` returns `NodeTypeProcess`

AC-3: WorkflowNode with NodeType serializes correctly
- Given a `WorkflowNode` with `NodeType: NodeTypeCondition`
- When marshaled to JSON
- Then the output contains `"nodeType":"condition"`
- And when round-tripped (marshal then unmarshal), the value is preserved

AC-4: WorkflowEdge has handle fields
- Given a `WorkflowEdge` with `SourceHandle: "true"` and `TargetHandle: "left"`
- When marshaled to JSON and unmarshaled back
- Then both handle values are preserved
- And an edge without handles produces JSON without `sourceHandle`/`targetHandle` keys (omitempty)

AC-5: WorkflowExecution has NodeOutputs field
- Given a `WorkflowExecution` with `NodeOutputs: map[string]string{"node-1": "output data"}`
- When marshaled to JSON
- Then the output contains the `nodeOutputs` key with the map
- And an execution without NodeOutputs produces JSON without `nodeOutputs` key (omitempty)

AC-6: All existing tests pass unchanged
- Given the current test suite at `internal/bmad/`
- When `go test ./internal/bmad/... -race` is run
- Then all existing tests pass without modification

## BDD Test Scenarios

### Scenario 1: NodeType Constants

```gherkin
Feature: Node type system

  Scenario: All node type constants have correct string values
    Given the NodeType type is defined
    When I check each constant value
    Then NodeTypeProcess equals "process"
    And NodeTypeCondition equals "condition"
    And NodeTypeLoop equals "loop"
    And NodeTypeLoopUntil equals "loopUntil"
    And NodeTypeTransform equals "transform"
    And NodeTypeMerge equals "merge"
```

### Scenario 2: EffectiveType normalization

```gherkin
  Scenario: EffectiveType returns process for empty NodeType
    Given a WorkflowNode with NodeType ""
    When I call EffectiveType()
    Then the result is NodeTypeProcess

  Scenario: EffectiveType returns the actual type when set
    Given a WorkflowNode with NodeType "condition"
    When I call EffectiveType()
    Then the result is NodeTypeCondition
```

### Scenario 3: Backward compatible JSON deserialization

```gherkin
  Scenario: Legacy JSON without nodeType deserializes cleanly
    Given JSON string: {"id":"n1","processId":"p1","label":"A","position":{"x":0,"y":0},"status":"pending","config":{},"tmuxTarget":""}
    When unmarshaled into WorkflowNode
    Then NodeType is ""
    And EffectiveType() returns NodeTypeProcess
    And no error is returned

  Scenario: JSON with nodeType deserializes correctly
    Given JSON string with "nodeType":"condition"
    When unmarshaled into WorkflowNode
    Then NodeType is NodeTypeCondition
```

### Scenario 4: Edge handle round-trip

```gherkin
  Scenario: Edge handles survive JSON round-trip
    Given a WorkflowEdge with SourceHandle "true" and TargetHandle "left"
    When marshaled to JSON and unmarshaled back
    Then SourceHandle is "true"
    And TargetHandle is "left"

  Scenario: Edge without handles omits fields in JSON
    Given a WorkflowEdge with empty SourceHandle and TargetHandle
    When marshaled to JSON
    Then the JSON does not contain "sourceHandle"
    And the JSON does not contain "targetHandle"
```

### Scenario 5: NodeOutputs on WorkflowExecution

```gherkin
  Scenario: NodeOutputs round-trip
    Given a WorkflowExecution with NodeOutputs {"node-1": "hello world"}
    When marshaled to JSON and unmarshaled back
    Then NodeOutputs["node-1"] equals "hello world"

  Scenario: Empty NodeOutputs omitted from JSON
    Given a WorkflowExecution with nil NodeOutputs
    When marshaled to JSON
    Then the JSON does not contain "nodeOutputs"
```

## Tasks / Subtasks

- [ ] Task 1: Add NodeType enum and constants (AC: AC-1)
  - [ ] Subtask 1a: Define `type NodeType string` and 6 constants in `types.go`
  - [ ] Subtask 1b: Add `NodeType NodeType json:"nodeType,omitempty"` field to `WorkflowNode`
  - [ ] Subtask 1c: Implement `func (n WorkflowNode) EffectiveType() NodeType` method

- [ ] Task 2: Extend WorkflowEdge with handle fields (AC: AC-4)
  - [ ] Subtask 2a: Add `SourceHandle string json:"sourceHandle,omitempty"` to `WorkflowEdge`
  - [ ] Subtask 2b: Add `TargetHandle string json:"targetHandle,omitempty"` to `WorkflowEdge`

- [ ] Task 3: Add NodeOutputs to WorkflowExecution (AC: AC-5)
  - [ ] Subtask 3a: Add `NodeOutputs map[string]string json:"nodeOutputs,omitempty"` to `WorkflowExecution`

- [ ] Task 4: Write comprehensive tests (AC: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6)
  - [ ] Subtask 4a: Create `internal/bmad/types_nodetype_test.go` with table-driven tests for NodeType constants, EffectiveType, JSON round-trips
  - [ ] Subtask 4b: Add edge handle serialization tests
  - [ ] Subtask 4c: Add NodeOutputs serialization tests
  - [ ] Subtask 4d: Run `go test ./internal/bmad/... -race` and verify all existing tests pass

## Dependencies
- Depends on: none
- Blocks: sprint3-02, sprint3-03, sprint3-04, sprint3-05, sprint3-06, sprint3-07, sprint3-08

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Existing 13 tests in `executor_test.go` pass unchanged
