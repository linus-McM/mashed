# Story 9: Canvas Editing -- Node/Edge Deletion, Reconnection, Multi-Select

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Add essential canvas editing capabilities to the BMAD workflow builder: deleting nodes and edges via keyboard, reconnecting edges by dragging endpoints, multi-selecting elements with Shift+click or drag rectangle, and providing clear visual feedback for selected/hovered-for-delete states. Without these operations, users cannot customize template workflows or correct wiring mistakes, which blocks the core repo-scoped sprint workflow UX.

## Developer Notes

### Architecture

All changes are frontend-only. No Go backend modifications required.

- **Modify:** `frontend/src/components/bmad/CanvasPane.svelte` -- add xyflow props for deletion key codes, selection key codes, edge reconnection, and new event handlers (`on:nodesdelete`, `on:edgesdelete`, `on:selectionchange`, `on:reconnect`)
- **Modify:** `frontend/src/views/WorkflowBuilder.svelte` -- add handler functions for deletion, reconnection, and selection change; pass them as props to CanvasPane; manage `selectedNode` clearing on deletion
- **Modify:** `frontend/src/components/bmad/ProcessNode.svelte` -- enhance the existing `.selected` CSS class and add a delete-hover visual state

**Data flow for deletion:**
1. User selects node(s)/edge(s) on canvas (click or Shift+click or drag-select)
2. User presses Delete or Backspace
3. xyflow fires `on:nodesdelete` and/or `on:edgesdelete` with the deleted element arrays
4. WorkflowBuilder's handler callbacks update the `nodes` and `edges` writable stores
5. If `selectedNode` was among the deleted nodes, it is cleared to close the NodeConfigPanel

**Data flow for edge reconnection:**
1. User hovers an edge endpoint, cursor changes to grab
2. User drags the endpoint to a different handle
3. xyflow fires `on:reconnect` with the old edge and the new connection params
4. WorkflowBuilder's handler removes the old edge and creates a replacement edge with the new source/target

**Data flow for multi-select:**
1. User Shift+clicks nodes/edges or drags a selection rectangle
2. xyflow fires `on:selectionchange` with `{ nodes: [...], edges: [...] }`
3. WorkflowBuilder tracks the selection set (optional state for future bulk operations)

### CanvasPane.svelte Changes

Add these props to the component:

```svelte
export let onNodesDelete = null;
export let onEdgesDelete = null;
export let onSelectionChange = null;
export let onReconnect = null;
```

Add these xyflow props to `<SvelteFlow>`:

```svelte
<SvelteFlow
  {nodes}
  {edges}
  {nodeTypes}
  {isValidConnection}
  on:connect={handleConnect}
  on:nodeclick={handleNodeClick}
  on:nodesdelete={handleNodesDelete}
  on:edgesdelete={handleEdgesDelete}
  on:selectionchange={handleSelectionChange}
  on:reconnect={handleReconnect}
  deleteKeyCode={['Delete', 'Backspace']}
  selectionKeyCode="Shift"
  multiSelectionKeyCode="Meta"
  edgesReconnectable
  fitView
>
```

Add handler functions that forward to the parent callbacks:

```javascript
function handleNodesDelete(e) {
  if (onNodesDelete) onNodesDelete(e.detail);
}

function handleEdgesDelete(e) {
  if (onEdgesDelete) onEdgesDelete(e.detail);
}

function handleSelectionChange(e) {
  if (onSelectionChange) onSelectionChange(e.detail);
}

function handleReconnect(e) {
  if (onReconnect) onReconnect(e.detail);
}
```

Add CSS for selected edges and reconnectable edge handles:

```css
.flow-wrap :global(.svelte-flow__edge.selected .svelte-flow__edge-path) {
  stroke: var(--accent-green);
  stroke-width: 3;
}

.flow-wrap :global(.svelte-flow__edge-path:hover) {
  stroke: var(--accent-green);
  cursor: pointer;
}

.flow-wrap :global(.svelte-flow__edgeupdater) {
  cursor: grab;
}

/* Selection rectangle */
.flow-wrap :global(.svelte-flow__selection) {
  background: rgba(0, 229, 122, 0.08);
  border: 1px dashed var(--accent-green);
}
```

### WorkflowBuilder.svelte Changes

Add these handler functions:

```javascript
function onNodesDelete(deletedNodes) {
  // xyflow already removes them from the store, but we need to:
  // 1. Clear selectedNode if it was deleted
  if (selectedNode && deletedNodes.some(n => n.id === selectedNode.id)) {
    selectedNode = null;
  }
  // 2. Update progress counts
  updateProgress();
}

function onEdgesDelete(deletedEdges) {
  // xyflow handles store removal; no additional state to clean up
}

function onReconnect(detail) {
  // detail contains { oldEdge, newConnection }
  // Replace the old edge with a new one using the updated connection
  const { oldEdge, newConnection } = detail;
  $edges = $edges.map(e => {
    if (e.id === oldEdge.id) {
      return {
        ...e,
        source: newConnection.source,
        target: newConnection.target,
        sourceHandle: newConnection.sourceHandle,
        targetHandle: newConnection.targetHandle,
      };
    }
    return e;
  });
}

function onSelectionChange(selection) {
  // Track for future bulk operations; update selectedNode for single-select
  if (selection.nodes.length === 1) {
    selectedNode = selection.nodes[0];
  } else if (selection.nodes.length === 0) {
    selectedNode = null;
  }
  // For multi-select, selectedNode stays null (NodeConfigPanel hides)
}
```

Pass the new handlers to CanvasPane:

```svelte
<CanvasPane
  {nodes}
  {edges}
  {nodeTypes}
  {isValidConnection}
  {onConnect}
  {onDropProcess}
  {onNodeClick}
  {onNodesDelete}
  {onEdgesDelete}
  {onSelectionChange}
  {onReconnect}
>
```

### ProcessNode.svelte Changes

The component already has `export let selected = false;` and a `.process-node.selected` CSS class. Enhance the selected styling for better visibility during multi-select and pre-delete:

```css
.process-node.selected {
  border-color: var(--accent-green);
  box-shadow: 0 0 0 2px rgba(0, 229, 122, 0.35);
}
```

This updates the existing `.selected` rule (currently `box-shadow: 0 0 0 1px var(--accent-green)`) to a slightly wider, semi-transparent glow that is more noticeable during multi-select.

### Technical Considerations

- **xyflow store management:** `@xyflow/svelte` internally modifies the writable stores when `deleteKeyCode` triggers deletion. The `on:nodesdelete` and `on:edgesdelete` events fire *after* the store is already updated, so handlers should NOT manually filter the stores -- they should only handle side effects (clearing `selectedNode`, updating progress).
- **Edge reconnection API:** The `on:reconnect` event in `@xyflow/svelte` provides `{ oldEdge, newConnection }`. The handler must update the edge in the store since xyflow does not auto-update reconnections.
- **Execution guard:** During active workflow execution (`executionStatus !== 'idle'`), deletion should be prevented. The simplest approach: do NOT pass `deleteKeyCode` when execution is running, or add an early return in the delete handlers.
- **No backend changes:** All node/edge state is managed client-side in Svelte writable stores. Deletion and reconnection only affect the in-memory canvas state until the user saves.

### Risks & Edge Cases

- **Deleting a running node:** If a user deletes a node that is currently executing in a tmux pane, the canvas removes it but the tmux process continues. The handlers should check for `status === 'running'` and either block deletion or warn. Simplest approach: skip deletion of running nodes.
- **Orphaned edges after node deletion:** xyflow automatically removes edges connected to deleted nodes, so no manual cleanup is needed.
- **Edge reconnection to same node:** `isValidConnection` already prevents self-loops (`connection.source === connection.target`), so reconnecting an edge back to its own node is blocked.
- **Duplicate edge after reconnection:** `isValidConnection` checks for existing edges, but since reconnection replaces the old edge, the check should pass. Verify this in testing.
- **Empty canvas after bulk delete:** If all nodes are deleted, the empty hint should reappear. The existing `{#if $nodes.length === 0 && !currentWorkflow}` conditional handles this partially -- may need to relax the `!currentWorkflow` check.
- **Undo/redo:** Not in scope for this story. Users must re-add deleted nodes manually. A future story can add undo support.

### Reference Files

- `frontend/src/components/bmad/CanvasPane.svelte` -- primary file to modify (xyflow wrapper)
- `frontend/src/views/WorkflowBuilder.svelte` -- add handler functions and pass as props
- `frontend/src/components/bmad/ProcessNode.svelte` -- enhance selection styling
- `docs/stories/sprint2-07-sprint-aware-canvas-nodes.md` -- reference for story format and node modification pattern

## Acceptance Criteria

AC-1: Single node deletion via keyboard
- Given a canvas with three nodes and the user clicks to select one node
- When the user presses the Delete or Backspace key
- Then the selected node is removed from the canvas
- And all edges connected to that node are also removed

AC-2: Single edge deletion via keyboard
- Given a canvas with nodes connected by an edge and the user clicks the edge to select it
- When the user presses the Delete or Backspace key
- Then the selected edge is removed from the canvas
- And the connected nodes remain on the canvas

AC-3: Multi-select and bulk deletion
- Given a canvas with five nodes
- When the user Shift+clicks three nodes (or drags a selection rectangle over them) and presses Delete
- Then all three selected nodes and their connected edges are removed
- And the remaining two nodes and their unaffected edges stay intact

AC-4: Edge reconnection by dragging
- Given a canvas with node A connected to node B via an edge
- When the user drags the target endpoint of that edge from node B to node C
- Then the edge now connects node A to node C
- And node B has no incoming edge from node A

AC-5: Selection visual feedback
- Given a canvas with multiple nodes
- When the user selects a node (single click) or multiple nodes (Shift+click)
- Then each selected node shows a green highlight border with a glow effect
- And selected edges show a thicker green stroke

AC-6: NodeConfigPanel clears on node deletion
- Given a node is selected and the NodeConfigPanel is showing its configuration
- When the user deletes that node
- Then the NodeConfigPanel closes (selectedNode becomes null)

AC-7: Running nodes are protected from deletion
- Given a workflow is executing and node X has status "running"
- When the user selects node X and presses Delete
- Then node X is NOT deleted
- And the node remains on the canvas with its running status

AC-8: Selection rectangle styling
- Given the user clicks on empty canvas space and drags to create a selection rectangle
- When the rectangle is visible
- Then it displays with a dashed green border and a semi-transparent green fill

## BDD Test Scenarios

### Scenario 1: Node and edge deletion

```gherkin
Feature: Canvas element deletion

  Scenario: Delete a single node removes it and connected edges
    Given a canvas with nodes "A", "B", "C" and edges "A->B" and "B->C"
    And node "B" is selected
    When the user presses the Delete key
    Then node "B" is removed from the nodes store
    And edges "A->B" and "B->C" are removed from the edges store
    And nodes "A" and "C" remain on the canvas

  Scenario: Delete a single edge keeps both nodes
    Given a canvas with nodes "A" and "B" and edge "A->B"
    And edge "A->B" is selected
    When the user presses the Backspace key
    Then edge "A->B" is removed from the edges store
    And nodes "A" and "B" remain on the canvas

  Scenario: Delete with nothing selected is a no-op
    Given a canvas with nodes "A" and "B" and edge "A->B"
    And no elements are selected
    When the user presses the Delete key
    Then the nodes store is unchanged
    And the edges store is unchanged
```

### Scenario 2: Multi-select and bulk operations

```gherkin
Feature: Multi-select canvas elements

  Scenario: Shift+click selects multiple nodes
    Given a canvas with nodes "A", "B", "C", "D"
    When the user clicks node "A"
    And Shift+clicks node "C"
    Then nodes "A" and "C" are both in the selected state
    And nodes "B" and "D" are not selected

  Scenario: Bulk delete removes all selected nodes
    Given a canvas with nodes "A", "B", "C" and edges "A->B", "B->C", "A->C"
    And nodes "A" and "B" are selected via Shift+click
    When the user presses Delete
    Then nodes "A" and "B" are removed
    And edges "A->B", "B->C", and "A->C" are removed
    And node "C" remains with no edges

  Scenario: Drag-select rectangle selects enclosed nodes
    Given a canvas with nodes "A" at (100,100), "B" at (200,200), "C" at (500,500)
    When the user drags a selection rectangle from (50,50) to (300,300)
    Then nodes "A" and "B" are selected
    And node "C" is not selected
```

### Scenario 3: Edge reconnection

```gherkin
Feature: Edge reconnection via drag

  Scenario: Reconnect edge target to a different node
    Given a canvas with nodes "A", "B", "C" and edge "A->B"
    When the user drags the target endpoint of edge "A->B" to node "C"
    Then the edge now connects "A" to "C"
    And no edge connects "A" to "B"

  Scenario: Reconnect blocked by self-loop validation
    Given a canvas with nodes "A", "B" and edge "A->B"
    When the user drags the target endpoint of edge "A->B" back to node "A"
    Then the reconnection is rejected by isValidConnection
    And the edge remains connecting "A" to "B"

  Scenario: Reconnect blocked by duplicate edge validation
    Given a canvas with nodes "A", "B", "C" and edges "A->B" and "A->C"
    When the user drags the target endpoint of edge "A->B" to node "C"
    Then the reconnection is rejected because edge "A->C" already exists
    And the edge remains connecting "A" to "B"
```

### Scenario 4: Selection visual feedback and panel interaction

```gherkin
Feature: Selection visual state

  Scenario: Selected node shows green glow
    Given a canvas with node "A" in default state
    When the user clicks node "A"
    Then node "A" has border-color set to the accent-green CSS variable
    And node "A" has a box-shadow glow effect

  Scenario: Selected edge shows thicker green stroke
    Given a canvas with edge "A->B" in default state
    When the user clicks edge "A->B"
    Then the edge path has stroke color set to accent-green
    And the edge path has stroke-width of 3

  Scenario: NodeConfigPanel closes when selected node is deleted
    Given node "A" is selected and the NodeConfigPanel displays its config
    When the user presses Delete to remove node "A"
    Then selectedNode becomes null
    And the NodeConfigPanel is hidden
```

### Scenario 5: Execution guard

```gherkin
Feature: Protect running nodes from deletion

  Scenario: Cannot delete a node with running status
    Given a workflow is executing with executionStatus "running"
    And node "A" has status "running"
    And node "A" is selected
    When the user presses Delete
    Then node "A" is NOT removed from the canvas
    And the edges connected to node "A" remain

  Scenario: Can delete idle nodes during execution
    Given a workflow is executing with executionStatus "running"
    And node "A" has status "pending"
    And node "A" is selected
    When the user presses Delete
    Then node "A" is removed from the canvas
```

## Tasks / Subtasks

- [ ] Task 1: Wire xyflow deletion props and events in CanvasPane (AC: AC-1, AC-2, AC-8)
  - [ ] Subtask 1a: Add `deleteKeyCode={['Delete', 'Backspace']}` prop to `<SvelteFlow>` in `CanvasPane.svelte`
  - [ ] Subtask 1b: Add `on:nodesdelete` and `on:edgesdelete` event handlers that forward to parent callbacks via new `onNodesDelete` and `onEdgesDelete` props
  - [ ] Subtask 1c: Add CSS rules for selection rectangle styling (dashed green border, semi-transparent fill)

- [ ] Task 2: Wire xyflow multi-select props and selection event in CanvasPane (AC: AC-3, AC-5)
  - [ ] Subtask 2a: Add `selectionKeyCode="Shift"` and `multiSelectionKeyCode="Meta"` props to `<SvelteFlow>`
  - [ ] Subtask 2b: Add `on:selectionchange` event handler forwarding to parent `onSelectionChange` callback
  - [ ] Subtask 2c: Add CSS rule for selected edge path (thicker green stroke)

- [ ] Task 3: Wire xyflow edge reconnection in CanvasPane (AC: AC-4)
  - [ ] Subtask 3a: Add `edgesReconnectable` prop to `<SvelteFlow>`
  - [ ] Subtask 3b: Add `on:reconnect` event handler forwarding to parent `onReconnect` callback
  - [ ] Subtask 3c: Add CSS rule for edge updater handle (grab cursor)

- [ ] Task 4: Implement deletion handlers in WorkflowBuilder (AC: AC-1, AC-2, AC-6, AC-7)
  - [ ] Subtask 4a: Add `onNodesDelete(deletedNodes)` function that clears `selectedNode` if it was among deleted nodes and calls `updateProgress()`
  - [ ] Subtask 4b: Add `onEdgesDelete(deletedEdges)` function (currently a no-op placeholder for future logic)
  - [ ] Subtask 4c: Add execution guard -- filter out nodes with `status === 'running'` before allowing deletion (conditionally set `deleteKeyCode` or use `on:beforenodesdelete` if available)
  - [ ] Subtask 4d: Pass `onNodesDelete` and `onEdgesDelete` as props to `<CanvasPane>`

- [ ] Task 5: Implement reconnection and selection handlers in WorkflowBuilder (AC: AC-3, AC-4, AC-5, AC-6)
  - [ ] Subtask 5a: Add `onReconnect(detail)` function that replaces the old edge with updated source/target in the `edges` store
  - [ ] Subtask 5b: Add `onSelectionChange(selection)` function that updates `selectedNode` for single-select and clears it for zero or multi-select
  - [ ] Subtask 5c: Pass `onReconnect` and `onSelectionChange` as props to `<CanvasPane>`

- [ ] Task 6: Enhance ProcessNode selection styling (AC: AC-5)
  - [ ] Subtask 6a: Update `.process-node.selected` CSS in `ProcessNode.svelte` to use wider semi-transparent glow (`box-shadow: 0 0 0 2px rgba(0, 229, 122, 0.35)`)
  - [ ] Subtask 6b: Verify the `selected` prop is already passed by xyflow (it is -- `export let selected = false` exists)

## Definition of Done

- [x] All acceptance criteria pass (AC-1 through AC-8 addressed)
- [ ] All BDD scenarios pass as automated tests (frontend — no test runner)
- [ ] 80%+ code coverage on new/modified files (frontend — no coverage tool)
- [x] `go build ./...` passes (no Go changes)
- [x] `go vet ./...` passes (no Go changes)
- [x] `go test ./... -race` passes (no Go changes)
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
