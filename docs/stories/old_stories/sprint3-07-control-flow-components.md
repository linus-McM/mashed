# Story: sprint3-07 -- Control Flow Svelte Node Components

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** L
**Depends On:** sprint3-01
**Status:** ready

## Description

Create five new Svelte components for the control flow node types: ConditionNode (diamond), LoopNode (rounded rect with loop icon), LoopUntilNode (rounded rect with target icon), TransformNode (small pill), and MergeNode (inverted diamond). Each component uses xyflow handles for typed connections and displays status, configuration summary, and execution state. These components are the visual representation of control flow on the canvas.

## Developer Notes

### Architecture
- **New files** in `frontend/src/components/bmad/`:
  - `ConditionNode.svelte` -- diamond shape, 1 input handle (left), 2 output handles (right-top "true" green, right-bottom "false" red)
  - `LoopNode.svelte` -- rounded rect with repeat icon, 1 input handle (left), 2 output handles (right-top "loop-body", right-bottom "loop-exit")
  - `LoopUntilNode.svelte` -- rounded rect with target icon, 1 input handle (left), 2 output handles (right-top "loop-body", right-bottom "loop-exit")
  - `TransformNode.svelte` -- small pill shape, 1 input handle (left), 1 output handle (right)
  - `MergeNode.svelte` -- inverted diamond, N input handles (left, dynamically positioned), 1 output handle (right)

### Component Structure (each follows ProcessNode.svelte pattern)
```svelte
<script>
  import { Handle, Position } from '@xyflow/svelte';
  export let data = {};
  export let id = '';
  export let selected = false;
  // Component-specific logic
</script>

<div class="node-wrapper" class:selected>
  <!-- Shape container -->
  <!-- Status display -->
  <Handle type="target" position={Position.Left} />
  <Handle type="source" position={Position.Right} id="handleId" />
</div>
```

### Design System (consistent with ProcessNode.svelte)
- Font: `var(--font-mono)`
- Colors: use CSS variables from DESIGN.md
- Selection: green border + box-shadow (same as ProcessNode `.selected` class)
- Status: reuse same status dot/check/x patterns from ProcessNode
- Node colors:
  - Condition: `var(--accent-purple, #bc8cff)` accent bar
  - Loop/LoopUntil: `var(--accent-amber, #d29922)` accent bar
  - Transform: `var(--accent-blue, #58a6ff)` accent bar
  - Merge: `var(--text-dim, #4a5a6a)` accent bar

### Handle IDs (critical for edge routing)
- ConditionNode outputs: `id="true"` and `id="false"`
- LoopNode/LoopUntilNode outputs: `id="loop-body"` and `id="loop-exit"`
- TransformNode output: default (no ID needed)
- MergeNode: single target handle, single source handle
- These IDs become the `sourceHandle` values on edges, which the backend uses for branch routing

### xyflow Handle Positioning
- Multiple output handles on the same side need vertical offset via `style` prop:
  - Top handle: `style="top: 35%"` with `id="true"` or `id="loop-body"`
  - Bottom handle: `style="top: 65%"` with `id="false"` or `id="loop-exit"`
- Labels next to handles (small text, e.g., "T" / "F" for condition, "body" / "exit" for loop)

### Data Props
Each component receives `data` from xyflow node:
- `data.status` -- `"pending"`, `"running"`, `"complete"`, `"failed"`, `"skipped"`
- `data.nodeType` -- the control flow type
- `data.label` -- user-defined label
- `data.config` -- type-specific config (condition type, pattern, max iterations, etc.)
- `data.iterationCount` -- (loop nodes) current iteration during execution

### Reference Files
- `frontend/src/components/bmad/ProcessNode.svelte` -- follow this exact pattern for component structure, styling, and status display
- `frontend/src/components/bmad/DeletableEdge.svelte` -- handle/edge patterns
- `DESIGN.md` -- CSS variable names and design tokens

### Skills
- `/xyflow` for Handle positioning, custom node registration
- `/simplify` mandatory

## Acceptance Criteria

AC-1: ConditionNode renders as diamond with two labeled output handles
- Given a ConditionNode with data `{nodeType: "condition", label: "Check Output", config: {conditionType: "contains", pattern: "SUCCESS"}}`
- When rendered on the canvas
- Then it displays as a diamond shape with purple accent
- And has one input handle on the left
- And has two output handles labeled "T" (green) and "F" (red) on the right
- And the label "Check Output" is visible
- And the condition summary (e.g., "contains: SUCCESS") is displayed

AC-2: LoopNode renders with loop icon and iteration display
- Given a LoopNode with `{nodeType: "loop", label: "Repeat 3x", config: {maxIterations: "3"}}`
- When rendered on the canvas
- Then it displays as a rounded rectangle with amber accent and repeat/loop icon
- And has handles: 1 input (left), "body" output (right-top), "exit" output (right-bottom)
- And shows "max: 3" text
- And during execution with `data.iterationCount = 2`, shows "2 / 3"

AC-3: LoopUntilNode renders with condition and iteration info
- Given a LoopUntilNode with condition config and maxIterations
- When rendered
- Then it shows both the condition summary and max iterations
- And has the same handle layout as LoopNode (body/exit)

AC-4: TransformNode renders as compact pill
- Given a TransformNode with `{nodeType: "transform", label: "Extract Version", config: {extractType: "regex", extractPattern: "v\\d+"}}`
- When rendered
- Then it displays as a small pill/capsule shape with blue accent
- And has 1 input and 1 output handle
- And shows the extract type summary

AC-5: MergeNode renders as inverted diamond
- Given a MergeNode with `{nodeType: "merge", label: "Join"}`
- When rendered
- Then it displays as an inverted diamond (point up, point down) with muted accent
- And has 1 target handle on the left and 1 source handle on the right

AC-6: All nodes display execution status
- Given any control flow node with `data.status = "running"`
- When rendered
- Then it shows the running status indicator (pulsing dot, same as ProcessNode)
- And `skipped` status shows a dash with "skipped" text

AC-7: Selection styling matches ProcessNode
- Given any control flow node that is selected
- When rendered
- Then it has green border and box-shadow (same as ProcessNode `.selected`)

## BDD Test Scenarios

### Scenario 1: ConditionNode rendering

```gherkin
Feature: Control flow node components

  Scenario: ConditionNode displays condition summary
    Given a ConditionNode on the canvas
    And its data has conditionType "contains" and pattern "SUCCESS"
    When the node renders
    Then the node shows text "contains: SUCCESS"
    And two output handles are visible with labels "T" and "F"

  Scenario: ConditionNode handles are correctly identified
    Given a ConditionNode
    When an edge is drawn from the top-right handle
    Then the edge sourceHandle is "true"
    When an edge is drawn from the bottom-right handle
    Then the edge sourceHandle is "false"
```

### Scenario 2: LoopNode iteration display

```gherkin
  Scenario: LoopNode shows max iterations
    Given a LoopNode with maxIterations "5"
    When rendered in pending state
    Then the node shows "max: 5"

  Scenario: LoopNode shows current iteration during execution
    Given a LoopNode with maxIterations "5" and iterationCount 3
    When rendered in running state
    Then the node shows "3 / 5"
```

### Scenario 3: TransformNode compact display

```gherkin
  Scenario: TransformNode shows extract type
    Given a TransformNode with extractType "regex"
    When rendered
    Then the node shows a compact pill shape
    And displays "regex" extract type indicator
```

### Scenario 4: Status display across types

```gherkin
  Scenario Outline: Status indicators work for all control flow nodes
    Given a <nodeType> node with status "<status>"
    When rendered
    Then the status indicator shows "<indicator>"

    Examples:
      | nodeType  | status   | indicator     |
      | condition | pending  | gray dot      |
      | condition | running  | pulsing dot   |
      | condition | complete | green check   |
      | loop      | failed   | red X         |
      | transform | skipped  | dash          |
      | merge     | complete | green check   |
```

### Scenario 5: MergeNode minimal display

```gherkin
  Scenario: MergeNode renders with minimal UI
    Given a MergeNode with label "Join Branches"
    When rendered
    Then the node shows an inverted diamond shape
    And displays "Join Branches"
    And has exactly 1 target and 1 source handle
```

## Tasks / Subtasks

- [ ] Task 1: Create ConditionNode.svelte (AC: AC-1, AC-6, AC-7)
  - [ ] Subtask 1a: Diamond shape CSS (rotated square with overflow hidden)
  - [ ] Subtask 1b: Two source handles with IDs "true" and "false", vertically offset
  - [ ] Subtask 1c: Condition summary display from data.config
  - [ ] Subtask 1d: Status indicators (reuse ProcessNode pattern)

- [ ] Task 2: Create LoopNode.svelte (AC: AC-2, AC-6, AC-7)
  - [ ] Subtask 2a: Rounded rect shape with repeat icon (use lucide-svelte `Repeat` or `RefreshCw`)
  - [ ] Subtask 2b: Two source handles with IDs "loop-body" and "loop-exit"
  - [ ] Subtask 2c: Max iterations display and iteration count during execution
  - [ ] Subtask 2d: Status indicators

- [ ] Task 3: Create LoopUntilNode.svelte (AC: AC-3, AC-6, AC-7)
  - [ ] Subtask 3a: Similar to LoopNode but with target icon (lucide-svelte `Target`)
  - [ ] Subtask 3b: Display both condition summary and max iterations
  - [ ] Subtask 3c: Same handle layout as LoopNode

- [ ] Task 4: Create TransformNode.svelte (AC: AC-4, AC-6, AC-7)
  - [ ] Subtask 4a: Compact pill/capsule shape CSS
  - [ ] Subtask 4b: Single input and output handles
  - [ ] Subtask 4c: Extract type display

- [ ] Task 5: Create MergeNode.svelte (AC: AC-5, AC-6, AC-7)
  - [ ] Subtask 5a: Inverted diamond shape CSS
  - [ ] Subtask 5b: Single target and source handles
  - [ ] Subtask 5c: Minimal label display

## Dependencies
- Depends on: sprint3-01 (NodeType constants define which components exist)
- Blocks: sprint3-08 (canvas registration and sidebar need these components)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All 5 components render correctly in the browser
- [ ] Handle IDs match backend expectations ("true", "false", "loop-body", "loop-exit")
- [ ] Status indicators work for all 5 status values
- [ ] Selection styling matches ProcessNode
- [ ] CSS uses design system variables (no hardcoded colors)
- [ ] `/simplify` run on all new components
- [ ] Code review: no CRITICAL/HIGH issues
