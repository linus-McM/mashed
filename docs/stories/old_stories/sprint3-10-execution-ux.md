# Story: sprint3-10 -- Execution UX: Edge Labels, Loop Display, Output Viewer

**Priority:** P2-medium
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** sprint3-08, sprint3-09
**Status:** ready

## Description

Polish the execution experience for control flow workflows. Add "true"/"false" labels on conditional edges so users can see which branch will be taken. Show real-time iteration counts on loop nodes during execution. Add an output viewer modal that displays captured terminal output for any completed node. These features make control flow workflows observable and debuggable.

## Developer Notes

### Architecture

#### 1. Edge labels for conditional edges
- **Modified file:** `frontend/src/components/bmad/DeletableEdge.svelte`
  - Accept new prop: `label` (string, from edge data)
  - Render label text next to the edge midpoint using `EdgeLabelRenderer`
  - Style: small mono text, color-coded (green for "true", red for "false", neutral for "loop-body"/"loop-exit")
- **Modified file:** `frontend/src/views/WorkflowBuilder.svelte`
  - In `onConnect()`, when the source node is a condition type, set `label` on the new edge based on `sourceHandle`:
    ```js
    label: sourceHandle === 'true' ? 'true' : sourceHandle === 'false' ? 'false' : 
           sourceHandle === 'loop-body' ? 'body' : sourceHandle === 'loop-exit' ? 'exit' : '',
    ```
  - In `loadNodesEdges()`, restore edge labels from saved data (or infer from sourceHandle)

#### 2. Loop iteration display
- **Modified file:** `frontend/src/views/WorkflowBuilder.svelte`
  - The `bmad:node:status` event listener already updates node data. Extend it to also update `iterationCount` from `NodeOutputs[nodeID+"_iter"]`
  - Alternative: emit a custom `bmad:node:iteration` event from backend with `{ nodeID, iteration, maxIterations }` -- simpler than polling NodeOutputs
  - The LoopNode/LoopUntilNode components already accept `data.iterationCount` (from sprint3-07)
- **Modified file:** `internal/bmad/executor.go` (minor)
  - Add `Iteration` field to `NodeStatusEvent` (optional, only set for loop nodes)
  - Emit iteration count with each loop iteration status event

#### 3. Output viewer modal
- **New file:** `frontend/src/components/bmad/OutputViewerModal.svelte`
  - Modal overlay with dark background
  - Header: "Output: {node label}" with close button
  - Body: preformatted text (`<pre>`) showing the captured output
  - Scroll: vertical scrollbar for long output
  - Max height: 80vh
  - Trigger: button on completed nodes in the config panel (next to existing "View Terminal" button)
- **Modified file:** `frontend/src/components/bmad/NodeConfigPanel.svelte`
  - Add "View Output" button visible when `node.data.status === 'complete'`
  - On click: call `GetNodeOutput(executionId, node.id)` and show OutputViewerModal
- **Modified file:** `frontend/src/views/WorkflowBuilder.svelte`
  - Add state: `showOutputModal`, `outputModalContent`, `outputModalLabel`
  - Handle `open-output` event from NodeConfigPanel

### Technical Considerations
- Edge labels should not interfere with the delete button on `DeletableEdge`. The label renders at a slight offset from the midpoint (e.g., `labelY - 15`).
- The output viewer loads data on-demand via `GetNodeOutput()` binding (sprint3-09), not from local state. This keeps memory usage low for large outputs.
- Loop iteration display is real-time -- updates on every `bmad:node:status` event with iteration data. The backend emits these events synchronously in the loop iteration.
- ANSI escape codes in terminal output: render as plain text in the output viewer. A future story could add ANSI-to-HTML rendering, but raw text is sufficient for this sprint.

### Reference Files
- `frontend/src/components/bmad/DeletableEdge.svelte` -- current edge component with delete button and EdgeLabelRenderer
- `frontend/src/components/bmad/NodeConfigPanel.svelte` -- existing "View Terminal" button pattern
- `frontend/src/views/WorkflowBuilder.svelte` -- event listeners for `bmad:node:status`
- `internal/bmad/executor.go` -- `NodeStatusEvent` struct at line 22

### Skills
- `/xyflow` for edge label rendering, EdgeLabelRenderer API
- `/wails` for GetNodeOutput binding call from frontend
- `/simplify` mandatory

## Acceptance Criteria

AC-1: Conditional edges display "true"/"false" labels
- Given a workflow with a condition node connected to two downstream nodes
- When the edges are rendered
- Then the "true" edge shows a green "true" label at the midpoint
- And the "false" edge shows a red "false" label

AC-2: Loop edges display "body"/"exit" labels
- Given a workflow with a loop node
- When the edges are rendered
- Then the body edge shows "body" label
- And the exit edge shows "exit" label

AC-3: Edge labels persist through save/load
- Given a workflow with labeled conditional edges
- When saved and reloaded
- Then the edge labels are restored (inferred from sourceHandle)

AC-4: Loop nodes show iteration count during execution
- Given a loop node with maxIterations=5 currently on iteration 3
- When the execution is in progress
- Then the loop node displays "3 / 5"

AC-5: Output viewer modal shows captured output
- Given a completed process node with captured terminal output
- When the user clicks "View Output" in the config panel
- Then a modal appears showing the full captured output in monospaced font
- And the modal has a close button and vertical scroll

AC-6: Output viewer loads data from backend on demand
- Given a completed node is selected
- When "View Output" is clicked
- Then `GetNodeOutput(execID, nodeID)` is called
- And the response is displayed in the modal
- And if the node has no output, the modal shows "No output captured"

AC-7: Edge labels do not interfere with edge delete button
- Given an edge with a label
- When the user hovers over the edge
- Then both the label and the delete button are visible without overlap

## BDD Test Scenarios

### Scenario 1: Conditional edge labels

```gherkin
Feature: Execution UX

  Scenario: Edges from condition nodes show branch labels
    Given a ConditionNode with two outbound edges
    And one edge has sourceHandle "true" and the other "false"
    When the edges render
    Then the first edge displays a green "true" label
    And the second edge displays a red "false" label

  Scenario: Edges from process nodes have no labels
    Given a ProcessNode with an outbound edge
    When the edge renders
    Then no label is displayed on the edge
```

### Scenario 2: Loop iteration real-time display

```gherkin
  Scenario: Loop node shows current iteration
    Given a LoopNode is executing with maxIterations 5
    When a bmad:node:status event arrives with iteration 3
    Then the LoopNode displays "3 / 5"

  Scenario: Loop node shows iteration after completion
    Given a LoopNode completed after 3 iterations
    When the canvas renders
    Then the LoopNode displays "3 / 3" (final count)
```

### Scenario 3: Output viewer modal

```gherkin
  Scenario: View output for completed node
    Given a completed process node is selected
    And the config panel is open
    When the user clicks "View Output"
    Then a modal appears with header "Output: {node label}"
    And the modal body shows the captured terminal output
    And the output is in monospaced font with scroll

  Scenario: No output available
    Given a completed node with no captured output
    When the user clicks "View Output"
    Then the modal shows "No output captured"

  Scenario: Close output modal
    Given the output modal is open
    When the user clicks the close button or presses Escape
    Then the modal closes
```

### Scenario 4: Edge label save/load

```gherkin
  Scenario: Labels restored from sourceHandle on load
    Given a saved workflow with condition edges having sourceHandle "true" and "false"
    When loaded
    Then the true edge has label "true" (green)
    And the false edge has label "false" (red)
```

## Tasks / Subtasks

- [ ] Task 1: Add edge labels to DeletableEdge (AC: AC-1, AC-2, AC-7)
  - [ ] Subtask 1a: Accept `label` and `data` props in DeletableEdge
  - [ ] Subtask 1b: Render label text via EdgeLabelRenderer with color coding
  - [ ] Subtask 1c: Position label offset from delete button to avoid overlap

- [ ] Task 2: Set edge labels on creation and load (AC: AC-1, AC-2, AC-3)
  - [ ] Subtask 2a: In `onConnect()`, set edge label based on source node type and sourceHandle
  - [ ] Subtask 2b: In `loadNodesEdges()`, infer label from sourceHandle on load
  - [ ] Subtask 2c: In `saveWorkflow()`, persist edge label (or infer on load)

- [ ] Task 3: Loop iteration real-time display (AC: AC-4)
  - [ ] Subtask 3a: Add `Iteration int` field to `NodeStatusEvent` in executor.go
  - [ ] Subtask 3b: Emit iteration count in loop execution events
  - [ ] Subtask 3c: Update `bmad:node:status` listener in WorkflowBuilder to pass iteration to node data

- [ ] Task 4: Create OutputViewerModal (AC: AC-5, AC-6)
  - [ ] Subtask 4a: Create `OutputViewerModal.svelte` with modal overlay, header, pre-formatted body, scroll
  - [ ] Subtask 4b: Call `GetNodeOutput()` on open, handle loading and empty states
  - [ ] Subtask 4c: Close on button click and Escape key

- [ ] Task 5: Wire output viewer into config panel (AC: AC-5, AC-6)
  - [ ] Subtask 5a: Add "View Output" button to NodeConfigPanel for completed nodes
  - [ ] Subtask 5b: Emit `open-output` event with node ID
  - [ ] Subtask 5c: Handle event in WorkflowBuilder, show modal

## Dependencies
- Depends on: sprint3-08 (node types registered, config panel type-aware), sprint3-09 (GetNodeOutput binding)
- Blocks: none (final story in sprint)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] Edge labels render correctly for condition and loop edges
- [ ] Loop iteration display updates in real-time during execution
- [ ] Output viewer modal opens, displays content, and closes
- [ ] No visual overlap between edge labels and delete buttons
- [ ] `/simplify` run on all modified components
- [ ] Code review: no CRITICAL/HIGH issues
