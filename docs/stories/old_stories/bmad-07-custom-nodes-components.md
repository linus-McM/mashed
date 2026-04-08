# Frontend: Custom ProcessNode & Supporting Components

**Story ID**: bmad-07
**Status**: ready
**Priority**: P1
**Depends On**: bmad-06

## Description

Replace the default SvelteFlow nodes with a rich custom `ProcessNode` component that shows phase color, agent role icon, input/output artifact labels, and live status badges. Also build the `ExecutionBar` (bottom panel with repo/model selectors and run/pause/stop controls), `NodeConfigPanel` (slide-out panel for per-node configuration), and `TemplatePicker` (enhanced template gallery). These components transform the canvas from a basic graph editor into a purpose-built BMAD workflow builder.

## Developer Notes

### Architecture

**New files:**
- `frontend/src/components/bmad/ProcessNode.svelte` -- Custom SvelteFlow node component
- `frontend/src/components/bmad/ExecutionBar.svelte` -- Bottom bar with execution controls
- `frontend/src/components/bmad/NodeConfigPanel.svelte` -- Slide-out config panel for selected node
- `frontend/src/components/bmad/TemplatePicker.svelte` -- Enhanced template gallery (replaces inline template list)

**Modified files:**
- `frontend/src/views/WorkflowBuilder.svelte` -- Register custom node type, add ExecutionBar, wire NodeConfigPanel

### ProcessNode.svelte

Register as a custom node type in SvelteFlow:
```svelte
<!-- WorkflowBuilder.svelte -->
<script>
  import ProcessNode from '../components/bmad/ProcessNode.svelte';
  const nodeTypes = { bmadProcess: ProcessNode };
</script>

<SvelteFlow {nodes} {edges} {nodeTypes} ... />
```

When creating nodes from drag-and-drop, use `type: 'bmadProcess'` instead of `'default'`.

**ProcessNode visual structure:**
```
+----------------------------------+
| [Phase color bar - 4px top]      |
| [AgentIcon]  Process Name        |
| in: PRD.md, architecture.md     |
| out: epics/                      |
| [Status badge: pending]          |
+--[Handle]----------[Handle]------+
```

**Phase colors** (using existing CSS variables where possible):
- analysis = `var(--accent-blue)` / `#3d9eff`
- planning = `var(--accent-green)` / `#00e57a`
- solutioning = `var(--accent-purple)` / `#9d6fff`
- implementation = `var(--accent-amber)` / `#f0a500`
- support = `var(--text-dim)` / `#4a5a6a`

**Agent role icons** (from `lucide-svelte`, already installed):
- analyst = `Search`
- pm = `Briefcase`
- ux-designer = `Palette`
- architect = `Building2`
- developer = `Code`
- tech-writer = `FileText`
- qa = `TestTube`

**Status badge rendering:**
- `pending` -- gray dot
- `running` -- green pulsing dot + pulsing green border on the entire node
- `complete` -- green checkmark
- `failed` -- red X
- `skipped` -- gray dash

**SvelteFlow node handles:**
```svelte
<Handle type="target" position={Position.Left} />
<Handle type="source" position={Position.Right} />
```

### ExecutionBar.svelte

Bottom bar component:
```
+-------------------------------------------------------------------+
| [Repo dropdown] [Model dropdown] | [Run] [Pause] [Stop] | 3/8 done |
+-------------------------------------------------------------------+
```

Props:
- `repoChoices` -- from `ListRepoChoices()` (existing binding)
- `executionStatus` -- `idle | running | paused | complete | failed`
- `nodeProgress` -- `{ completed: number, total: number }`

Events dispatched:
- `start` with `{ repoPath, model }` detail
- `pause`
- `resume`
- `stop`

Model options: `["claude-opus-4-6", "claude-sonnet-4-20250514", "claude-haiku-3.5"]`

Button state logic:
- Idle: Run enabled, Pause/Stop disabled
- Running: Run disabled, Pause/Stop enabled
- Paused: Run (resume) enabled, Stop enabled, Pause disabled
- Complete/Failed: Run (restart) enabled, Pause/Stop disabled

### NodeConfigPanel.svelte

Slide-out panel that appears when a node is clicked:
```
+---------------------+
| Configure: Node Name|
| Model: [dropdown]   |
| Context: [textarea] |
| Agent: [dropdown]   |
| [View Terminal]     |
| [Close]             |
+---------------------+
```

Props:
- `node` -- the selected SvelteFlow node
- `agents` -- list of custom agents from `ListBmadAgents()`

Events:
- `update` with updated node config
- `close`
- `open-terminal` with tmuxTarget

### TemplatePicker.svelte

Grid of template cards:
```
+-------------+  +-------------+
| Quick Sprint|  | PRD Pipeline|
| 4 nodes     |  | 4 nodes     |
| [Use]       |  | [Use]       |
+-------------+  +-------------+
```

Props:
- `templates` -- array of WorkflowDef templates

Events:
- `select` with template ID

### Technical Considerations

- **SvelteFlow custom nodes**: Custom node components receive `data`, `id`, `selected`, `sourcePosition`, `targetPosition` as props. The `data` object is whatever was set when creating the node.
- **Live status updates**: The WorkflowBuilder listens for `bmad:node:status` events via `EventsOn` and updates the node's `data.status` field. SvelteFlow reactively re-renders the ProcessNode.
- **lucide-svelte icons**: Already installed at `0.577.0`. Import as `import { Code, Briefcase, ... } from 'lucide-svelte'`.
- **Existing Terminal.svelte**: The "View Terminal" button in NodeConfigPanel should dispatch an event that the WorkflowBuilder handles by opening the existing Terminal component with the node's tmuxTarget.

### Risks & Edge Cases

- **SvelteFlow custom node prop types**: Ensure `data` passed to the node contains all fields the ProcessNode expects. If `data.process` is missing (e.g., after deserialization), look it up from the registry via `GetBmadProcesses()`.
- **Status updates for nodes not on screen**: SvelteFlow virtualizes rendering. Status updates to off-screen nodes should still update the store; the node will render correctly when scrolled into view.
- **Responsive layout**: The ExecutionBar should not overlap the canvas. Use a fixed-height bottom bar with the canvas taking remaining space.

### Reference Files

- `frontend/src/components/StatusBadge.svelte` -- Existing pattern for status display
- `frontend/src/components/Terminal.svelte` -- Terminal component for node terminal access
- `frontend/src/views/SpawnAgent.svelte` -- Modal pattern with repo/model selectors
- `frontend/src/components/TitleBar.svelte` -- Component layout patterns
- `@xyflow/svelte` custom nodes docs: https://svelteflow.dev/learn/customization/custom-nodes

## Acceptance Criteria

- [ ] AC1: Given a node on the canvas, When it renders, Then it shows the process name, phase color bar, agent role icon, input/output labels, and a status badge.
- [ ] AC2: Given a node with `status: "running"`, When the ProcessNode renders, Then it displays a pulsing green border and a green pulsing status dot.
- [ ] AC3: Given the ExecutionBar is visible, When the user selects a repo and model and clicks Run, Then a `start` event is dispatched with `{ repoPath, model }`.
- [ ] AC4: Given the execution is running, When the Pause button is clicked, Then a `pause` event is dispatched and the Run button becomes a Resume button.
- [ ] AC5: Given a node is clicked on the canvas, When the click is handled, Then the NodeConfigPanel slides out showing the node's configuration options (model override, custom context, agent dropdown).
- [ ] AC6: Given the NodeConfigPanel is open for a running node, When "View Terminal" is clicked, Then the terminal opens for that node's tmuxTarget.
- [ ] AC7: Given the TemplatePicker is displayed, When the user clicks "Use" on a template card, Then a `select` event is dispatched with the template ID.

## BDD Test Scenarios

### Scenario 1: ProcessNode rendering

```gherkin
Feature: BMAD Process Node

  Scenario: Node displays process metadata
    Given a ProcessNode with data containing processId "bmad-create-prd"
    And the process is in phase "planning" with agent role "pm"
    When the node renders
    Then the phase color bar is green (--accent-green)
    And the agent icon is Briefcase
    And the label shows "Create PRD"
    And inputs show "product-brief"
    And outputs show "PRD.md"

  Scenario: Running status shows pulsing animation
    Given a ProcessNode with data.status "running"
    When the node renders
    Then the node has a pulsing green border
    And the status dot is green and animated

  Scenario: Failed status shows red indicator
    Given a ProcessNode with data.status "failed"
    When the node renders
    Then the status badge shows a red X
```

### Scenario 2: ExecutionBar controls

```gherkin
Feature: Execution Bar

  Scenario: Idle state shows Run enabled
    Given execution status is "idle"
    When the ExecutionBar renders
    Then the Run button is enabled
    And Pause and Stop buttons are disabled

  Scenario: Running state enables Pause and Stop
    Given execution status is "running"
    When the ExecutionBar renders
    Then Run is disabled
    And Pause and Stop are enabled
    And progress shows "3 of 8 complete"
```

### Scenario 3: Node configuration

```gherkin
Feature: Node Config Panel

  Scenario: Click node opens config panel
    Given the canvas has a node "Create PRD"
    When the user clicks the node
    Then the NodeConfigPanel slides in from the right
    And shows model override dropdown
    And shows custom context textarea

  Scenario: View terminal for running node
    Given a running node with tmuxTarget "bmad-node1-123:0.0"
    When the user clicks "View Terminal" in NodeConfigPanel
    Then the terminal opens connected to "bmad-node1-123:0.0"
```

## Tasks / Subtasks

- [ ] Task 1: Build ProcessNode.svelte (AC: AC1, AC2)
  - [ ] Subtask 1a: Create component with Handle imports from @xyflow/svelte
  - [ ] Subtask 1b: Implement phase color bar using phase-to-color mapping
  - [ ] Subtask 1c: Implement agent role icon using lucide-svelte
  - [ ] Subtask 1d: Implement input/output label display
  - [ ] Subtask 1e: Implement status badge with animations (pulse for running)
- [ ] Task 2: Build ExecutionBar.svelte (AC: AC3, AC4)
  - [ ] Subtask 2a: Create layout with repo dropdown, model dropdown, control buttons
  - [ ] Subtask 2b: Wire ListRepoChoices for repo options
  - [ ] Subtask 2c: Implement button state logic based on execution status
  - [ ] Subtask 2d: Implement progress display (completed/total)
- [ ] Task 3: Build NodeConfigPanel.svelte (AC: AC5, AC6)
  - [ ] Subtask 3a: Create slide-out panel with transition
  - [ ] Subtask 3b: Add model override dropdown
  - [ ] Subtask 3c: Add custom context textarea
  - [ ] Subtask 3d: Add agent assignment dropdown (from ListBmadAgents)
  - [ ] Subtask 3e: Add "View Terminal" button (dispatches open-terminal event)
- [ ] Task 4: Build TemplatePicker.svelte (AC: AC7)
  - [ ] Subtask 4a: Create grid layout of template cards
  - [ ] Subtask 4b: Show template name, description, node count
  - [ ] Subtask 4c: Wire "Use" button to dispatch select event
- [ ] Task 5: Integrate into WorkflowBuilder.svelte (AC: AC1-AC7)
  - [ ] Subtask 5a: Register `bmadProcess` custom node type
  - [ ] Subtask 5b: Update node creation to use `type: 'bmadProcess'`
  - [ ] Subtask 5c: Add ExecutionBar below the canvas
  - [ ] Subtask 5d: Wire node click to open NodeConfigPanel
  - [ ] Subtask 5e: Listen for `bmad:node:status` events and update node data

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] ProcessNode renders correctly with all metadata
- [ ] ExecutionBar buttons reflect execution state
- [ ] NodeConfigPanel opens/closes smoothly
- [ ] `cd frontend && npm run build` passes
- [ ] `wails build` passes
- [ ] Theme CSS variables are used for all colors
- [ ] /simplify run on all new code
