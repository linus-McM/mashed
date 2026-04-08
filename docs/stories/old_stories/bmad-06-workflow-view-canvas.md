# Frontend: WorkflowBuilder View & SvelteFlow Canvas

**Story ID**: bmad-06
**Status**: ready
**Priority**: P1
**Depends On**: bmad-01, bmad-05

## Description

Create the main WorkflowBuilder view with a three-panel layout: a left sidebar with process palette and template/workflow lists, a center SvelteFlow canvas for visual workflow editing, and the navigation integration to reach it from the app. This is the core frontend surface that lets users drag BMAD processes onto a canvas, connect them with edges, save/load workflows, and select templates. This story does NOT include custom node rendering (uses default nodes) or the execution bar -- those come in subsequent stories.

## Developer Notes

### Architecture

**New files:**
- `frontend/src/views/WorkflowBuilder.svelte` -- Main view with three-panel layout
- `frontend/src/components/bmad/ProcessSidebar.svelte` -- Left panel: process palette, templates, saved workflows

**Modified files:**
- `frontend/src/App.svelte` -- Add `'workflows'` to the view state machine, import WorkflowBuilder, add navigation button

### App.svelte Changes

The current `currentView` state machine is: `'loading' | 'setup' | 'feed' | 'detail' | 'settings'`

Add `'workflows'` to this set. In the template section, add:
```svelte
{:else if currentView === 'workflows'}
  <WorkflowBuilder on:back={goBack} />
```

Add a navigation trigger. Options:
1. Add a button in the TitleBar (preferred -- matches existing settings gear pattern)
2. Add a keyboard shortcut (Ctrl/Cmd+W)

Import at the top:
```svelte
import WorkflowBuilder from './views/WorkflowBuilder.svelte';
```

### WorkflowBuilder.svelte Structure

```svelte
<script>
  import { SvelteFlow, Controls, MiniMap, Background } from '@xyflow/svelte';
  import '@xyflow/svelte/dist/style.css';
  import { writable } from 'svelte/store';
  import { createEventDispatcher } from 'svelte';
  import { GetBmadProcesses, ListBmadTemplates, ListBmadWorkflows,
           SaveBmadWorkflow, GetBmadWorkflow, CreateFromTemplate,
           DeleteBmadWorkflow } from '../../wailsjs/go/main/App.js';
  import ProcessSidebar from '../components/bmad/ProcessSidebar.svelte';

  const dispatch = createEventDispatcher();

  // SvelteFlow reactive stores
  const nodes = writable([]);
  const edges = writable([]);

  let processes = [];
  let templates = [];
  let savedWorkflows = [];
  let currentWorkflow = null;  // WorkflowDef being edited
  let workflowName = 'Untitled Workflow';
</script>
```

**Three-panel layout:**
```
+-------------------+-------------------------------+
| ProcessSidebar    | SvelteFlow Canvas             |
| 240px fixed width | flex: 1                       |
|                   |                               |
+-------------------+-------------------------------+
```

### ProcessSidebar.svelte

Accordion sections grouped by BMAD phase:
- Analysis (blue accent)
- Planning (green accent)
- Solutioning (purple accent)
- Implementation (amber accent)
- Support (gray accent)

Each process item is draggable:
```svelte
<div
  draggable="true"
  on:dragstart={(e) => {
    e.dataTransfer.setData('application/bmad-process', process.id);
    e.dataTransfer.effectAllowed = 'move';
  }}
>
  {process.name}
</div>
```

Tabs at the top of the sidebar:
1. **Processes** -- The phase-grouped palette
2. **Templates** -- Grid of template cards (name, node count, "Use" button)
3. **Saved** -- List of user-saved workflows (click to load, delete button)

### Canvas Drop Handler

On the `SvelteFlow` wrapper div:
```svelte
<div
  on:dragover|preventDefault
  on:drop={(e) => {
    const processId = e.dataTransfer.getData('application/bmad-process');
    if (!processId) return;
    const position = screenToFlowPosition({ x: e.clientX, y: e.clientY });
    // Create a new node from the process definition
    const process = processes.find(p => p.id === processId);
    const newNode = {
      id: `node-${Date.now()}`,
      type: 'default',  // will become 'bmadProcess' in bmad-07
      position,
      data: { label: process.name, processId: process.id, process }
    };
    $nodes = [...$nodes, newNode];
  }}
>
```

### Connection Validation

Use SvelteFlow's `isValidConnection` prop to validate edges:
```svelte
function isValidConnection(connection) {
  // Prevent self-connections
  if (connection.source === connection.target) return false;
  // Prevent duplicate edges
  const exists = $edges.some(e =>
    e.source === connection.source && e.target === connection.target
  );
  return !exists;
}
```

### Save/Load

Save button in a toolbar above the canvas:
```javascript
async function saveWorkflow() {
  const wf = {
    id: currentWorkflow?.id || `wf-${Date.now()}`,
    name: workflowName,
    nodes: $nodes.map(n => ({
      id: n.id,
      processId: n.data.processId,
      label: n.data.label,
      position: n.position,
      status: 'pending',
      config: {},
    })),
    edges: $edges.map(e => ({ id: e.id, source: e.source, target: e.target })),
    isTemplate: false,
  };
  await SaveBmadWorkflow(wf);
  savedWorkflows = await ListBmadWorkflows();
}
```

### Technical Considerations

- **SvelteFlow stores**: `@xyflow/svelte` uses Svelte writable stores for nodes and edges. Use `bind:nodes` and `bind:edges` or the store-based approach.
- **CSS**: Import `@xyflow/svelte/dist/style.css` for default SvelteFlow styling. Override with CSS variables from the app's theme system (use `--bg-surface`, `--border-subtle`, etc.).
- **SvelteFlow version**: Use `@xyflow/svelte@^1.0.0`. The API uses `<SvelteFlow>`, not the older `<ReactFlow>` or `<SvelteFlowProvider>`.
- **screenToFlowPosition**: This function is available from the `useSvelteFlow` hook in `@xyflow/svelte`. It converts screen coordinates to flow coordinates for accurate drop placement.
- **Theme integration**: The SvelteFlow canvas background and minimap should use the app's CSS variables for consistent theming.

### Risks & Edge Cases

- **SvelteFlow + Svelte 4 compatibility**: `@xyflow/svelte@1.x` requires Svelte 4+. Story bmad-01 ensures this.
- **Large workflows**: A workflow with 20+ nodes may need zoom/pan controls. SvelteFlow's `<Controls>` component provides this.
- **Empty state**: When no workflow is loaded, show an empty canvas with a hint: "Drag processes from the sidebar or select a template."
- **Unsaved changes**: If the user navigates away with unsaved changes, no blocking dialog is needed for MVP, but the workflow state should persist in component state.

### Reference Files

- `frontend/src/App.svelte` -- View state machine pattern (lines 16, 131-155)
- `frontend/src/views/Settings.svelte` -- Back button pattern (`dispatch('back')`)
- `frontend/src/views/NotificationFeed.svelte` -- View layout patterns
- `frontend/src/components/TitleBar.svelte` -- Navigation button pattern
- `@xyflow/svelte` docs at https://svelteflow.dev/

## Acceptance Criteria

- [ ] AC1: Given the running app on the feed view, When the user clicks the Workflows navigation button (or presses the keyboard shortcut), Then the view transitions to WorkflowBuilder.
- [ ] AC2: Given the WorkflowBuilder view, When the sidebar loads, Then it displays all BMAD processes grouped by phase (Analysis, Planning, Solutioning, Implementation, Support) from `GetBmadProcesses()`.
- [ ] AC3: Given the WorkflowBuilder view, When the user drags a process from the sidebar and drops it on the canvas, Then a new node appears at the drop position with the process name as its label.
- [ ] AC4: Given two nodes on the canvas, When the user drags from one node's output handle to another's input handle, Then an edge is created connecting them.
- [ ] AC5: Given a workflow with nodes and edges on the canvas, When the user clicks "Save" and enters a name, Then `SaveBmadWorkflow` is called and the workflow appears in the Saved tab.
- [ ] AC6: Given a saved workflow in the Saved tab, When the user clicks it, Then the canvas loads with that workflow's nodes and edges.
- [ ] AC7: Given the Templates tab in the sidebar, When the user clicks "Use" on a template, Then `CreateFromTemplate` is called and the canvas populates with the template's nodes and edges.
- [ ] AC8: Given the WorkflowBuilder view, When the user clicks the Back button, Then the view returns to the feed.

## BDD Test Scenarios

### Scenario 1: Navigation

```gherkin
Feature: Workflow Builder Navigation

  Scenario: Navigate to WorkflowBuilder from feed
    Given the app is showing the feed view
    When the user clicks the Workflows button in the TitleBar
    Then the WorkflowBuilder view is displayed
    And the sidebar shows process categories

  Scenario: Navigate back to feed
    Given the WorkflowBuilder view is displayed
    When the user clicks the Back button
    Then the feed view is displayed
```

### Scenario 2: Drag and drop

```gherkin
Feature: Workflow Canvas Drag and Drop

  Scenario: Drag process to canvas creates node
    Given the WorkflowBuilder is displayed with processes loaded
    When the user drags "Create PRD" from the sidebar
    And drops it on the canvas at position (400, 300)
    Then a node appears at approximately (400, 300)
    And the node label is "Create PRD"
    And the node data contains processId "bmad-create-prd"

  Scenario: Connect two nodes with an edge
    Given two nodes exist on the canvas
    When the user drags from node A's source handle to node B's target handle
    Then an edge appears connecting node A to node B

  Scenario: Self-connection is prevented
    Given a node exists on the canvas
    When the user tries to connect the node to itself
    Then no edge is created
```

### Scenario 3: Save and load

```gherkin
Feature: Workflow Save and Load

  Scenario: Save workflow
    Given the canvas has 3 nodes and 2 edges
    And the workflow name is "My Sprint Workflow"
    When the user clicks Save
    Then SaveBmadWorkflow is called with the correct node and edge data
    And the workflow appears in the Saved tab

  Scenario: Load saved workflow
    Given a workflow "My Sprint Workflow" exists in storage
    When the user clicks it in the Saved tab
    Then the canvas displays the workflow's nodes and edges

  Scenario: Load template
    Given the Templates tab is active
    When the user clicks "Use" on "Quick Sprint"
    Then CreateFromTemplate is called with "quick-sprint"
    And the canvas populates with 4 nodes and 3 edges
```

## Tasks / Subtasks

- [ ] Task 1: Add navigation to App.svelte (AC: AC1, AC8)
  - [ ] Subtask 1a: Add `'workflows'` to currentView state machine
  - [ ] Subtask 1b: Import WorkflowBuilder and add conditional render block
  - [ ] Subtask 1c: Add navigation button to TitleBar (or keyboard shortcut)
  - [ ] Subtask 1d: Wire `on:back` to `goBack()`
- [ ] Task 2: Create WorkflowBuilder.svelte with layout (AC: AC2)
  - [ ] Subtask 2a: Create the three-panel layout (sidebar + canvas)
  - [ ] Subtask 2b: Load processes, templates, saved workflows on mount
  - [ ] Subtask 2c: Initialize SvelteFlow with nodes/edges stores, Controls, MiniMap, Background
- [ ] Task 3: Create ProcessSidebar.svelte (AC: AC2, AC7)
  - [ ] Subtask 3a: Implement phase-grouped accordion with draggable process items
  - [ ] Subtask 3b: Implement Templates tab with "Use" button calling CreateFromTemplate
  - [ ] Subtask 3c: Implement Saved tab with workflow list and click-to-load
- [ ] Task 4: Implement drag-and-drop and connections (AC: AC3, AC4)
  - [ ] Subtask 4a: Implement drop handler on canvas wrapper
  - [ ] Subtask 4b: Implement `isValidConnection` for edge validation
  - [ ] Subtask 4c: Handle `screenToFlowPosition` for accurate placement
- [ ] Task 5: Implement save/load (AC: AC5, AC6)
  - [ ] Subtask 5a: Implement save toolbar with name input and save button
  - [ ] Subtask 5b: Implement save function mapping SvelteFlow nodes/edges to WorkflowDef
  - [ ] Subtask 5c: Implement load function mapping WorkflowDef to SvelteFlow nodes/edges

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] WorkflowBuilder view renders without errors
- [ ] SvelteFlow canvas is interactive (pan, zoom, drag nodes)
- [ ] `cd frontend && npm run build` passes
- [ ] `wails build` passes
- [ ] Theme CSS variables are applied to the canvas
- [ ] /simplify run on all new code
