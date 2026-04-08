# Story 6: ExecutionBar Simplification -- Remove Repo Dropdown

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** Story 4
**Status:** done

## Description

Remove the repo dropdown from ExecutionBar since the repo is now implicit from the WorkflowBuilder's repoPath prop. This simplifies the execution UI by eliminating a redundant control -- the user already selected a repo when entering the builder. The model selector and play/pause/stop controls remain unchanged.

## Developer Notes

### Architecture
- **Modify:** `frontend/src/components/bmad/ExecutionBar.svelte` -- remove repo dropdown, accept repoPath as prop
- **Modify:** `frontend/src/views/WorkflowBuilder.svelte` -- pass repoPath to ExecutionBar, update event handling
- No backend changes required

### ExecutionBar.svelte Changes

Remove:
- The `ListRepoChoices` import
- The `repoChoices` array and `selectedRepo` state
- The `onMount` that calls `ListRepoChoices()`
- The repo `<select>` element from the template
- The repo path from the `start` event dispatch

Add:
```svelte
export let repoPath = '';  // NEW: passed from parent

// In handleRun, use repoPath prop directly:
function handleRun() {
  if (isPaused) {
    dispatch('resume');
  } else {
    dispatch('start', { model: selectedModel });  // repoPath removed from event
  }
}
```

Display the repo context as a label instead of a dropdown:
```svelte
<div class="repo-label" title={repoPath}>
  {repoPath.split('/').pop()}
</div>
```

### WorkflowBuilder.svelte Changes

```svelte
<!-- Pass repoPath to ExecutionBar -->
<ExecutionBar
  {executionStatus}
  {nodeProgress}
  repoPath={repoPath}
  on:start={handleExecStart}
  on:pause={handleExecPause}
  on:resume={handleExecResume}
  on:stop={handleExecStop}
/>

<!-- Update handleExecStart to use prop repoPath -->
async function handleExecStart(e) {
  const { model } = e.detail;  // No more repoPath from event
  // ...
  executionId = await StartBmadWorkflow(currentWorkflow.id, repoPath, model);
  terminalRepoPath = repoPath;
}
```

### Technical Considerations
- The `ListRepoChoices` import can be fully removed from ExecutionBar (it will still be used in RepoPickerModal from Story 4)
- The repo label should show just the directory name (last path segment) with full path in a title tooltip
- The repo label should use the same styling as the model selector for visual consistency
- Ensure the execution flow still works: `handleExecStart` now reads repoPath from the component prop, not from the event detail

### Risks & Edge Cases
- If repoPath prop is empty (should not happen after Story 4), the Start button should be disabled
- Existing keyboard shortcut behavior is unaffected

### Reference Files
- `frontend/src/components/bmad/ExecutionBar.svelte` -- the file to simplify
- `frontend/src/views/WorkflowBuilder.svelte` -- parent that dispatches/handles events

## Acceptance Criteria

AC-1: Repo dropdown is removed
- Given the ExecutionBar is rendered
- When the user looks at the execution controls
- Then there is no repo dropdown/select element
- And a repo name label is shown instead

AC-2: Repo label shows current repo
- Given WorkflowBuilder has repoPath "/Users/linus/Development/my-project"
- When ExecutionBar renders
- Then it shows "my-project" as the repo label
- And the full path is available as a tooltip

AC-3: Execution uses prop repoPath
- Given ExecutionBar has repoPath="/Users/linus/Development/my-project"
- When the user clicks Run
- Then the start event is dispatched with the selected model only
- And WorkflowBuilder uses its repoPath prop for the StartBmadWorkflow call

AC-4: Start disabled without repoPath
- Given ExecutionBar has an empty repoPath prop
- When rendered
- Then the Run button is disabled

## BDD Test Scenarios

### Scenario 1: Simplified execution bar

```gherkin
Feature: ExecutionBar without repo dropdown

  Scenario: Repo displayed as label not dropdown
    Given ExecutionBar rendered with repoPath "/Users/linus/Development/my-app"
    Then a text label "my-app" is visible in the bar
    And no <select> element for repos exists
    And the model dropdown still exists

  Scenario: Run dispatches without repoPath in event
    Given ExecutionBar with repoPath set and model "claude-opus-4-6" selected
    When the user clicks Run
    Then the "start" event detail contains model "claude-opus-4-6"
    And the "start" event detail does not contain repoPath

  Scenario: Run disabled when repoPath is empty
    Given ExecutionBar with repoPath=""
    Then the Run button is disabled
    And the model dropdown is still enabled

  Scenario: Tooltip shows full path
    Given ExecutionBar with repoPath "/Users/linus/Development/my-long-project-name"
    When the user hovers over the repo label
    Then the tooltip shows "/Users/linus/Development/my-long-project-name"
```

## Tasks / Subtasks

- [ ] Task 1: Remove repo dropdown from ExecutionBar (AC: AC-1, AC-2, AC-4)
  - [ ] Subtask 1a: Remove `ListRepoChoices` import, `repoChoices`, `selectedRepo`, and the onMount fetch
  - [ ] Subtask 1b: Add `export let repoPath = '';` prop
  - [ ] Subtask 1c: Replace the repo `<select>` with a styled repo label showing the directory basename
  - [ ] Subtask 1d: Add `title={repoPath}` for full-path tooltip on the label
  - [ ] Subtask 1e: Disable Run button when `repoPath` is empty

- [ ] Task 2: Update event dispatch and parent wiring (AC: AC-3)
  - [ ] Subtask 2a: Modify `handleRun` to dispatch `start` event without repoPath (model only)
  - [ ] Subtask 2b: Update WorkflowBuilder's `handleExecStart` to read repoPath from prop, not event detail
  - [ ] Subtask 2c: Pass `repoPath` prop from WorkflowBuilder to ExecutionBar

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
