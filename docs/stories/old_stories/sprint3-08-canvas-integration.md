# Story: sprint3-08 -- Canvas Integration: Register Types, Sidebar, Config Panel, Save/Load

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** L
**Depends On:** sprint3-07
**Status:** ready

## Description

Wire the five new control flow node components into the workflow builder canvas. This includes registering node types with xyflow, adding a "Control Flow" group to the sidebar with drag-and-drop support, making the config panel type-aware (showing different fields for condition vs. loop vs. transform), and updating save/load to persist `nodeType`, `sourceHandle`, and `targetHandle` through the Go backend.

## Developer Notes

### Architecture

#### 1. Register node types
- **Modified file:** `frontend/src/views/WorkflowBuilder.svelte`
  - Import all 5 new components from `../components/bmad/`
  - Expand `nodeTypes` map:
    ```js
    const nodeTypes = {
      bmadProcess: ProcessNode,
      condition: ConditionNode,
      loop: LoopNode,
      loopUntil: LoopUntilNode,
      transform: TransformNode,
      merge: MergeNode,
    };
    ```
  - The xyflow `type` field on each node determines which component renders

#### 2. Sidebar "Control Flow" group
- **Modified file:** `frontend/src/components/bmad/ProcessSidebar.svelte`
  - Add a new collapsible group below the process phases in the "Processes" tab (or as a new section visible in all tabs)
  - Hardcoded list of 5 control flow node types (these are NOT from the process registry):
    ```js
    const controlFlowNodes = [
      { type: 'condition', name: 'Condition', description: 'If/else branch based on output', icon: 'GitBranch' },
      { type: 'loop', name: 'Loop', description: 'Repeat N times', icon: 'Repeat' },
      { type: 'loopUntil', name: 'Loop Until', description: 'Repeat until condition met', icon: 'Target' },
      { type: 'transform', name: 'Transform', description: 'Extract/transform data', icon: 'Filter' },
      { type: 'merge', name: 'Merge', description: 'Join branches', icon: 'GitMerge' },
    ];
    ```
  - Drag data transfer type: `application/bmad-controlflow` (distinct from `application/bmad-process`)
  - Drop handler in WorkflowBuilder creates node with `type: item.type` and `data.nodeType: item.type`

#### 3. Type-aware config panel
- **Modified file:** `frontend/src/components/bmad/NodeConfigPanel.svelte`
  - Switch panel body based on `node.data.nodeType`:
    - `undefined` or `"process"` -- existing model/context/agent fields (unchanged)
    - `"condition"` -- dropdown for condition type (contains, notContains, regex, exitCode, fileExists, always), text input for pattern, dropdown for source node (populated from upstream completed nodes)
    - `"loop"` -- number input for max iterations
    - `"loopUntil"` -- number input for max iterations + condition fields (same as condition)
    - `"transform"` -- dropdown for extract type (regex, lines), text input for extract pattern, dropdown for source node
    - `"merge"` -- label-only (no config needed)
  - Source node dropdown: populated from `node` list, filtered to upstream nodes (those that have a path to this node via edges). Simplification: just list all nodes except self.

#### 4. Save/load updates
- **Modified file:** `frontend/src/views/WorkflowBuilder.svelte`
  - `saveWorkflow()`: Include `nodeType` in node serialization, include `sourceHandle` and `targetHandle` in edge serialization
    ```js
    nodes: $nodes.map(n => ({
      ...existingFields,
      nodeType: n.data.nodeType || '',  // NEW
    })),
    edges: $edges.map(e => ({
      id: e.id,
      source: e.source,
      target: e.target,
      sourceHandle: e.sourceHandle || '',  // NEW
      targetHandle: e.targetHandle || '',  // NEW
    })),
    ```
  - `loadNodesEdges()`: Map `nodeType` to xyflow `type` field:
    ```js
    type: n.nodeType ? n.nodeType : 'bmadProcess',  // '' or 'process' -> 'bmadProcess'
    ```
  - Load edge handles:
    ```js
    sourceHandle: e.sourceHandle || undefined,
    targetHandle: e.targetHandle || undefined,
    ```

#### 5. Drop handler for control flow nodes
- **Modified file:** `frontend/src/views/WorkflowBuilder.svelte`
  - New function: `onDropControlFlow(nodeType, position)`
  - Called from `CanvasPane` drop handler when `dataTransfer` type is `application/bmad-controlflow`
  - Creates node with:
    ```js
    {
      id: `cf-${Date.now()}`,
      type: nodeType,  // 'condition', 'loop', etc.
      position,
      data: { nodeType, label: controlFlowNames[nodeType], status: 'pending', config: {} },
    }
    ```

### Technical Considerations
- Edge handles are already tracked by xyflow internally. Currently `saveWorkflow()` strips them (only saves `id`, `source`, `target`). This story adds `sourceHandle`/`targetHandle` to the saved data.
- The backend `WorkflowEdge` struct now has `SourceHandle` and `TargetHandle` fields (from sprint3-01), so no backend changes needed.
- The config panel `emitUpdate()` function must include all config fields for the current node type, not just model/context/agent.
- Backward compat: existing saved workflows have no `nodeType` on nodes and no handles on edges. `loadNodesEdges()` defaults to `bmadProcess` type and undefined handles.

### Reference Files
- `frontend/src/views/WorkflowBuilder.svelte` -- `saveWorkflow()` at line 323, `loadNodesEdges()` at line 356, `onDropProcess()` at line 157, `nodeTypes` at line 27
- `frontend/src/components/bmad/ProcessSidebar.svelte` -- sidebar tabs and drag handling
- `frontend/src/components/bmad/NodeConfigPanel.svelte` -- config fields and `emitUpdate()`
- `frontend/src/components/bmad/CanvasPane.svelte` -- drop handler, SvelteFlow props

### Skills
- `/xyflow` for node type registration, handle props, drop handling
- `/simplify` mandatory

## Acceptance Criteria

AC-1: All 5 control flow node types render on canvas
- Given a workflow with one of each control flow node type
- When loaded into the canvas
- Then each renders with its correct component (diamond, rounded rect, pill, inverted diamond)

AC-2: Control flow group appears in sidebar
- Given the ProcessSidebar is visible
- When the user views the Processes tab (or a dedicated section)
- Then a "Control Flow" collapsible group shows 5 items: Condition, Loop, Loop Until, Transform, Merge
- And each item is draggable

AC-3: Drag-and-drop creates control flow nodes
- Given the user drags a "Condition" item from the sidebar
- When dropped onto the canvas
- Then a ConditionNode is created with `type: "condition"` and `data.nodeType: "condition"`
- And it renders as a diamond with handles

AC-4: Config panel shows type-specific fields
- Given a condition node is selected
- When the config panel opens
- Then it shows: condition type dropdown, pattern input, source node dropdown
- And does NOT show: model selector, context textarea, agent selector

AC-5: Config panel shows process fields for process nodes
- Given a process node is selected (existing behavior)
- When the config panel opens
- Then it shows model, context, agent fields (unchanged from current)

AC-6: Save persists nodeType and edge handles
- Given a workflow with condition nodes and handle-labeled edges
- When saved via `saveWorkflow()`
- Then the saved JSON includes `nodeType` on nodes and `sourceHandle`/`targetHandle` on edges

AC-7: Load restores nodeType and edge handles
- Given a saved workflow with control flow nodes
- When loaded via `loadNodesEdges()`
- Then nodes render with correct component types
- And edges connect to correct handles

AC-8: Backward compat: old workflows load correctly
- Given a saved workflow from sprint2 (no nodeType, no handles)
- When loaded
- Then all nodes render as `bmadProcess` (ProcessNode)
- And edges connect normally (no handle constraints)

## BDD Test Scenarios

### Scenario 1: Node type registration

```gherkin
Feature: Canvas integration

  Scenario: All node types are registered with xyflow
    Given the WorkflowBuilder is mounted
    When the nodeTypes map is inspected
    Then it contains keys: bmadProcess, condition, loop, loopUntil, transform, merge
    And each maps to its respective Svelte component
```

### Scenario 2: Sidebar drag-and-drop

```gherkin
  Scenario: Control flow node dragged from sidebar
    Given the sidebar shows the "Control Flow" group
    When the user drags "Loop" and drops it on the canvas
    Then a new LoopNode appears at the drop position
    And it has type "loop" and data.nodeType "loop"

  Scenario: Process node drag still works
    Given the sidebar shows the "Processes" tab
    When the user drags a process and drops it on the canvas
    Then a ProcessNode appears (unchanged behavior)
```

### Scenario 3: Type-aware config panel

```gherkin
  Scenario: Condition node shows condition config
    Given a condition node is selected on the canvas
    When the config panel renders
    Then it shows a "Condition Type" dropdown with options: contains, notContains, regex, exitCode, fileExists, always
    And a "Pattern" text input
    And a "Source Node" dropdown listing other nodes

  Scenario: Loop node shows iteration config
    Given a loop node is selected
    When the config panel renders
    Then it shows a "Max Iterations" number input
    And does NOT show model/context/agent fields

  Scenario: Transform node shows extraction config
    Given a transform node is selected
    When the config panel renders
    Then it shows "Extract Type" dropdown (regex, lines)
    And "Extract Pattern" text input
    And "Source Node" dropdown
```

### Scenario 4: Save/load round-trip

```gherkin
  Scenario: Control flow workflow survives save/load
    Given a canvas with Process(A) -> Condition(B) -> Process(C, true) -> Process(D, false)
    When saved and then reloaded
    Then all nodes restore with correct types
    And edges restore with sourceHandle "true" and "false"
    And the condition node renders as diamond
```

### Scenario 5: Backward compatibility

```gherkin
  Scenario: Sprint2 workflow loads without errors
    Given a workflow saved in sprint2 format (no nodeType, no handles)
    When loaded in sprint3 builder
    Then all nodes render as ProcessNode (bmadProcess)
    And no console errors occur
```

## Tasks / Subtasks

- [ ] Task 1: Register node types in WorkflowBuilder (AC: AC-1)
  - [ ] Subtask 1a: Import 5 new components
  - [ ] Subtask 1b: Add to `nodeTypes` map

- [ ] Task 2: Add Control Flow group to sidebar (AC: AC-2, AC-3)
  - [ ] Subtask 2a: Add `controlFlowNodes` array to ProcessSidebar
  - [ ] Subtask 2b: Render collapsible "Control Flow" group with icons
  - [ ] Subtask 2c: Implement drag with `application/bmad-controlflow` data transfer
  - [ ] Subtask 2d: Add drop handler in WorkflowBuilder/CanvasPane

- [ ] Task 3: Make config panel type-aware (AC: AC-4, AC-5)
  - [ ] Subtask 3a: Add `nodeType` detection from `node.data.nodeType`
  - [ ] Subtask 3b: Condition config: type dropdown, pattern input, source node dropdown
  - [ ] Subtask 3c: Loop config: max iterations input
  - [ ] Subtask 3d: LoopUntil config: max iterations + condition fields
  - [ ] Subtask 3e: Transform config: extract type dropdown, pattern input, source node dropdown
  - [ ] Subtask 3f: Merge config: label-only display
  - [ ] Subtask 3g: Update `emitUpdate()` to include type-specific config fields

- [ ] Task 4: Update save/load for nodeType and handles (AC: AC-6, AC-7, AC-8)
  - [ ] Subtask 4a: Add `nodeType` to `saveWorkflow()` node serialization
  - [ ] Subtask 4b: Add `sourceHandle`/`targetHandle` to edge serialization
  - [ ] Subtask 4c: Update `loadNodesEdges()` to map nodeType to xyflow type
  - [ ] Subtask 4d: Load edge handles
  - [ ] Subtask 4e: Test backward compat with old workflow JSON

- [ ] Task 5: Add control flow drop handler (AC: AC-3)
  - [ ] Subtask 5a: Create `onDropControlFlow(type, position)` in WorkflowBuilder
  - [ ] Subtask 5b: Wire into CanvasPane drop handler alongside existing process and story drop handlers

## Dependencies
- Depends on: sprint3-07 (the 5 Svelte components must exist before registration)
- Blocks: sprint3-10 (execution UX needs registered types)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] Control flow nodes can be dragged from sidebar to canvas
- [ ] Config panel shows correct fields for each node type
- [ ] Save/load preserves nodeType and edge handles
- [ ] Old workflows load without errors
- [ ] `/simplify` run on all modified components
- [ ] Code review: no CRITICAL/HIGH issues
