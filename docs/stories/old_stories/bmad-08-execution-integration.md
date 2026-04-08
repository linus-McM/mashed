# Execution Integration: Live Status, Terminal Access & Artifact Passing

**Story ID**: bmad-08
**Status**: ready
**Priority**: P1
**Depends On**: bmad-05, bmad-07

## Description

Wire the frontend execution controls to the backend executor, enabling end-to-end workflow execution. When the user clicks Run, the backend spawns tmux sessions for each node in topological order, emits real-time status events that update the canvas nodes live, and the user can click any running node to open its terminal. Artifact context is passed between sequential nodes so downstream Claude sessions know about upstream outputs. This story makes the workflow builder actually execute BMAD processes.

## Developer Notes

### Architecture

**Modified files:**
- `frontend/src/views/WorkflowBuilder.svelte` -- Wire ExecutionBar events to Wails bindings, handle `bmad:node:status` and `bmad:execution:status` events, manage execution state
- `internal/bmad/executor.go` -- Ensure artifact context passing is implemented (may already be done in bmad-04, but verify integration)

**No new files** -- this story is pure integration of components built in bmad-04 through bmad-07.

### Frontend Event Wiring

In `WorkflowBuilder.svelte`:

```svelte
<script>
  import { EventsOn } from '../../wailsjs/runtime/runtime.js';
  import { StartBmadWorkflow, PauseBmadWorkflow, ResumeBmadWorkflow,
           StopBmadWorkflow, GetBmadExecution } from '../../wailsjs/go/main/App.js';

  let executionId = null;
  let executionStatus = 'idle';  // idle | running | paused | complete | failed
  let nodeProgress = { completed: 0, total: 0 };

  // Handle ExecutionBar events
  async function handleStart(e) {
    const { repoPath, model } = e.detail;
    if (!currentWorkflow?.id) {
      // Auto-save before executing
      await saveWorkflow();
    }
    executionId = await StartBmadWorkflow(currentWorkflow.id, repoPath, model);
    executionStatus = 'running';
  }

  async function handlePause() {
    await PauseBmadWorkflow(executionId);
    executionStatus = 'paused';
  }

  async function handleResume() {
    await ResumeBmadWorkflow(executionId);
    executionStatus = 'running';
  }

  async function handleStop() {
    await StopBmadWorkflow(executionId);
    executionStatus = 'failed';
  }

  // Listen for node status updates
  EventsOn('bmad:node:status', (data) => {
    // data: { execID, nodeID, status, tmuxTarget }
    if (data.execID !== executionId) return;

    $nodes = $nodes.map(n => {
      if (n.id === data.nodeID) {
        return {
          ...n,
          data: {
            ...n.data,
            status: data.status,
            tmuxTarget: data.tmuxTarget,
          }
        };
      }
      return n;
    });

    // Update progress
    const completed = $nodes.filter(n => n.data.status === 'complete').length;
    nodeProgress = { completed, total: $nodes.length };
  });

  // Listen for execution-level status
  EventsOn('bmad:execution:status', (data) => {
    if (data.execID !== executionId) return;
    executionStatus = data.status;
  });
</script>

<ExecutionBar
  {executionStatus}
  {nodeProgress}
  on:start={handleStart}
  on:pause={handlePause}
  on:resume={handleResume}
  on:stop={handleStop}
/>
```

### Terminal Access for Running Nodes

When a node is running and the user clicks "View Terminal" in the NodeConfigPanel:

```svelte
function handleOpenTerminal(e) {
  const { tmuxTarget } = e.detail;
  // Option A: Navigate to AgentDetail view with the tmux target
  // Option B: Show inline terminal in a modal/panel using Terminal.svelte
  // Recommended: Option B for MVP -- show a modal with Terminal.svelte

  terminalTarget = tmuxTarget;
  showTerminalModal = true;
}
```

The existing `Terminal.svelte` component connects via WebSocket to `localhost:{terminalPort}/ws/{tmuxTarget}`. Get the terminal port via `GetTerminalPort()` (existing binding at `app.go` line 701).

### Artifact Context Passing (Backend)

The executor (bmad-04) handles this, but verify the integration:
1. When node A completes, check `ProcessDef.Outputs` for artifact names (e.g., `["PRD.md"]`)
2. For downstream node B that has `ProcessDef.Inputs` containing `"PRD.md"`:
   - Build context: `"The upstream process produced PRD.md at docs/PRD.md. Use this as input."`
   - Append to the Claude CLI command
3. Artifact paths are relative to the repo working directory

### Event Data Shapes

```typescript
// bmad:node:status event data
interface NodeStatusEvent {
  execID: string;
  nodeID: string;
  status: 'pending' | 'running' | 'complete' | 'failed' | 'skipped';
  tmuxTarget: string;  // set when status is 'running'
}

// bmad:execution:status event data
interface ExecutionStatusEvent {
  execID: string;
  status: 'idle' | 'running' | 'paused' | 'complete' | 'failed';
  currentNode: string;  // node ID currently executing
}
```

### Technical Considerations

- **Auto-save before execution**: The workflow must be saved to storage before `StartBmadWorkflow` is called, because the executor loads it by ID from storage.
- **Event scoping**: Multiple workflows could theoretically run simultaneously. Filter events by `execID` to update only the active canvas.
- **Error display**: If `StartBmadWorkflow` returns an error (e.g., tmux not available, cyclic workflow), display it in a toast/banner above the ExecutionBar.
- **Re-execution**: After a workflow completes or fails, the user should be able to click Run again to re-execute. Reset all node statuses to `pending` before starting.
- **Terminal cleanup**: When the WorkflowBuilder unmounts, unsubscribe from events to prevent memory leaks. Use `onDestroy` or the return value of `EventsOn`.

### Risks & Edge Cases

- **tmux not installed**: If the user has no tmux, every node spawn will fail. The executor should report this clearly. The frontend should show the error from `StartBmadWorkflow`.
- **Rapid status updates**: If many nodes run in parallel, many `bmad:node:status` events arrive quickly. SvelteFlow handles reactive updates efficiently via stores, but test with 10+ concurrent nodes.
- **Lost events**: If the frontend misses an event (e.g., during a re-render), `GetBmadExecution` can be polled to reconcile state.
- **Stale terminal**: If a node's tmux session exits but the terminal modal is still open, the terminal will show a dead pane. This is acceptable for MVP.

### Reference Files

- `frontend/src/views/WorkflowBuilder.svelte` (from bmad-06) -- main view to modify
- `frontend/src/components/bmad/ExecutionBar.svelte` (from bmad-07) -- execution controls
- `frontend/src/components/bmad/NodeConfigPanel.svelte` (from bmad-07) -- node config
- `frontend/src/components/Terminal.svelte` -- existing terminal component
- `frontend/src/App.svelte` lines 46-54 -- EventsOn pattern for real-time updates
- `app.go` line 701 -- `GetTerminalPort()` binding

## Acceptance Criteria

- [ ] AC1: Given a workflow on the canvas with a repo and model selected, When the user clicks Run, Then `StartBmadWorkflow` is called and the first node's status changes to `running` on the canvas.
- [ ] AC2: Given a running workflow, When a node completes, Then the node's status badge updates to `complete` (green checkmark) and the next node begins execution.
- [ ] AC3: Given a running workflow, When the user clicks Pause, Then no new nodes start, and the currently running node finishes. Clicking Resume continues execution.
- [ ] AC4: Given a running workflow, When the user clicks Stop, Then the execution is canceled and all remaining nodes show `pending` or `failed`.
- [ ] AC5: Given a running node with a tmuxTarget, When the user clicks the node and then clicks "View Terminal" in the config panel, Then a terminal opens connected to that node's tmux session.
- [ ] AC6: Given a completed node that outputs `"PRD.md"`, When the downstream node starts, Then the executor passes artifact context to the downstream Claude CLI command.
- [ ] AC7: Given a workflow execution that fails to start (e.g., tmux not available), When the error is returned, Then an error message is displayed to the user near the ExecutionBar.

## BDD Test Scenarios

### Scenario 1: End-to-end execution

```gherkin
Feature: Workflow Execution Integration

  Scenario: Run a two-node workflow
    Given the canvas has nodes "Brainstorming" and "Product Brief" connected
    And repo "/tmp/test-repo" is selected
    And model "claude-sonnet-4-20250514" is selected
    When the user clicks Run
    Then StartBmadWorkflow is called with the workflow ID, repo, and model
    And "Brainstorming" node shows status "running"
    And after it completes, "Product Brief" node shows status "running"
    And after it completes, execution status is "complete"

  Scenario: Pause and resume execution
    Given a workflow is running with node A executing
    When Pause is clicked
    Then node A continues to completion
    And node B does not start
    When Resume is clicked
    Then node B starts executing

  Scenario: Stop cancels execution
    Given a workflow is running with node A executing
    When Stop is clicked
    Then execution status becomes "failed"
```

### Scenario 2: Terminal access

```gherkin
Feature: Node Terminal Access

  Scenario: Open terminal for running node
    Given node "Create PRD" is running with tmuxTarget "bmad-prd-123:0.0"
    When the user clicks the node
    And clicks "View Terminal" in the config panel
    Then a terminal modal opens
    And connects to WebSocket at the terminal bridge
```

### Scenario 3: Error handling

```gherkin
Feature: Execution Error Handling

  Scenario: tmux not available
    Given tmux is not installed on the system
    When the user clicks Run
    Then StartBmadWorkflow returns an error
    And an error message is shown near the ExecutionBar
    And no nodes change status

  Scenario: Node execution fails
    Given a workflow is running
    When node A's tmux session exits with non-zero status
    Then node A shows status "failed"
    And execution pauses
    And the user sees the failure in the ExecutionBar
```

## Tasks / Subtasks

- [ ] Task 1: Wire ExecutionBar to Wails bindings (AC: AC1, AC3, AC4)
  - [ ] Subtask 1a: Handle `start` event: auto-save workflow, call `StartBmadWorkflow`
  - [ ] Subtask 1b: Handle `pause` event: call `PauseBmadWorkflow`
  - [ ] Subtask 1c: Handle `resume` event: call `ResumeBmadWorkflow`
  - [ ] Subtask 1d: Handle `stop` event: call `StopBmadWorkflow`
- [ ] Task 2: Listen for real-time status events (AC: AC2)
  - [ ] Subtask 2a: Subscribe to `bmad:node:status` via `EventsOn`
  - [ ] Subtask 2b: Update node data in SvelteFlow store on status change
  - [ ] Subtask 2c: Update `nodeProgress` for ExecutionBar display
  - [ ] Subtask 2d: Subscribe to `bmad:execution:status` for overall state
- [ ] Task 3: Implement terminal access (AC: AC5)
  - [ ] Subtask 3a: Handle `open-terminal` event from NodeConfigPanel
  - [ ] Subtask 3b: Show Terminal.svelte in a modal connected to the node's tmuxTarget
  - [ ] Subtask 3c: Get terminal port via `GetTerminalPort()` for WebSocket URL
- [ ] Task 4: Error handling and edge cases (AC: AC7)
  - [ ] Subtask 4a: Display StartBmadWorkflow errors in UI
  - [ ] Subtask 4b: Handle re-execution (reset node statuses before re-run)
  - [ ] Subtask 4c: Unsubscribe from events on component destroy
- [ ] Task 5: Verify artifact passing (AC: AC6)
  - [ ] Subtask 5a: Test with a two-node workflow where node 1 outputs and node 2 inputs the same artifact
  - [ ] Subtask 5b: Verify the Claude CLI command for node 2 includes artifact context

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] End-to-end execution works with a real 2-node workflow and tmux
- [ ] Node status updates appear in real-time on the canvas
- [ ] Terminal access works for running nodes
- [ ] Error states are displayed to the user
- [ ] `cd frontend && npm run build` passes
- [ ] `wails build` passes
- [ ] /simplify run on all modified code
