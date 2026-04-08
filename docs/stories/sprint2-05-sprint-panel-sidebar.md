# Story 5: Sprint Panel in Sidebar

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** L
**Depends On:** Story 2, Story 4
**Status:** ready

## Description

Add a "Sprint" tab to the ProcessSidebar that displays epics and stories from sprint-status.yaml, color-coded by status. Stories are draggable onto the canvas to create workflow nodes linked to specific sprint stories. This gives the user direct visibility into their sprint backlog while building workflows, making story-to-workflow mapping natural and intuitive.

## Developer Notes

### Architecture
- **Modify:** `frontend/src/components/bmad/ProcessSidebar.svelte` -- add 4th tab "Sprint", accept new props
- **New file:** `frontend/src/components/bmad/SprintPanel.svelte` -- the sprint tab content (extracted for readability)
- **Modify:** `frontend/src/views/WorkflowBuilder.svelte` -- load sprint data, pass to sidebar, handle story drops
- Uses `GetSprintStatus(repoPath)` binding from Story 2

### ProcessSidebar Changes

Add a 4th tab:
```svelte
export let sprintStatus = null;  // NEW prop
export let repoPath = '';        // NEW prop

<!-- In tabs section -->
<button class="tab" class:active={activeTab === 'sprint'} on:click={() => activeTab = 'sprint'}>Sprint</button>

<!-- In tab-content -->
{:else if activeTab === 'sprint'}
  <SprintPanel {sprintStatus} on:drag-story />
{/if}
```

### SprintPanel.svelte Design

Layout per epic:
```
[Epic-1 header] (status pill) (story count)
  [v] 1-1 User Auth           [DONE]       (green)
  [_] 1-2 Account Mgmt        [READY]      (blue)
  [_] 1-3 Data Model           [BACKLOG]    (gray)
  [_] epic-1-retrospective    [OPTIONAL]   (dim)

[Epic-2 header] (status pill) (story count)
  [_] 2-1 Personality          [BACKLOG]    (gray)
  [_] 2-2 Chat UI              [BACKLOG]    (gray)
```

Status colors (matching existing theme CSS vars):
- `backlog` -> `var(--text-muted)` (gray)
- `ready-for-dev` -> `var(--accent-blue, #3d9eff)` (blue)
- `in-progress` -> `var(--accent-amber, #f0a500)` (amber)
- `review` -> `var(--accent-purple, #9d6fff)` (purple)
- `done` -> `var(--accent-green, #00e57a)` (green)

Each story item is draggable. On drag start, set:
```javascript
e.dataTransfer.setData('application/bmad-story', JSON.stringify({
  storyId: story.id,
  epicId: story.epicId,
  status: story.status,
}));
```

### WorkflowBuilder.svelte Changes

Add sprint data loading in onMount:
```javascript
let sprintStatus = null;

onMount(async () => {
  // ...existing loads...
  if (repoPath) {
    try {
      sprintStatus = await GetSprintStatus(repoPath);
    } catch (e) {
      console.warn('No sprint status for repo:', e);
    }
  }
});
```

Handle story drops on canvas (modify `onDropProcess` or add `onDropStory`):
```javascript
function onDropStory(storyData, position) {
  // Create a node that references a story instead of (or in addition to) a process
  const newNode = {
    id: `story-node-${Date.now()}`,
    type: 'bmadProcess',
    position,
    data: {
      label: storyData.storyId,
      processId: 'bmad-dev-story',  // default process for stories
      storyId: storyData.storyId,   // NEW: link to sprint story
      storyStatus: storyData.status,
      status: 'pending',
      config: { storyId: storyData.storyId },
    },
  };
  $nodes = [...$nodes, newNode];
}
```

### Technical Considerations
- The CanvasPane's drop handler must differentiate between process drags (`application/bmad-process`) and story drags (`application/bmad-story`)
- Sprint data should be refreshed when the user returns to the Sprint tab or on a timer (prepare for fsnotify events from Story 2's WatchSprintStatus)
- Epic sections should be collapsible (reuse the existing `expandedPhases` pattern from ProcessSidebar)
- Story count badge per epic: `{doneCount}/{totalCount}` format
- If sprint-status.yaml is missing, show a helpful message: "No sprint-status.yaml found in this repo"

### Risks & Edge Cases
- Large sprints with many epics/stories could overflow the sidebar -- ensure scrolling works within the Sprint tab
- Repos without BMAD init have no sprint-status.yaml -- degrade gracefully with empty state
- Story IDs might contain characters that need escaping in data transfer -- use JSON serialization

### Reference Files
- `frontend/src/components/bmad/ProcessSidebar.svelte` -- tab structure, phase grouping pattern, drag handling
- `frontend/src/views/WorkflowBuilder.svelte` -- `onDropProcess` function for canvas drop handling
- `frontend/src/components/bmad/CanvasPane.svelte` -- drop event handling on the canvas

## Acceptance Criteria

AC-1: Sprint tab appears in sidebar
- Given the WorkflowBuilder is loaded with a repoPath
- When the user looks at the sidebar tabs
- Then there are 4 tabs: Processes, Templates, Saved, Sprint

AC-2: Epics and stories displayed from YAML
- Given the repo has a sprint-status.yaml with 2 epics and 5 total stories
- When the user clicks the Sprint tab
- Then both epics are shown with their stories listed underneath
- And each story shows its ID and status

AC-3: Status colors are correct
- Given stories with various statuses in the Sprint panel
- When rendered
- Then "done" stories have green status indicators
- And "ready-for-dev" stories have blue indicators
- And "backlog" stories have gray indicators
- And "in-progress" stories have amber indicators

AC-4: Stories are draggable to canvas
- Given a story in the Sprint panel
- When the user drags it onto the canvas
- Then a new workflow node is created
- And the node's data includes the storyId reference
- And the node uses "bmad-dev-story" as the default process

AC-5: Empty state when no sprint data
- Given a repo without sprint-status.yaml
- When the user clicks the Sprint tab
- Then a message "No sprint data found" is displayed
- And no errors are thrown

AC-6: Epic sections are collapsible
- Given epics displayed in the Sprint panel
- When the user clicks an epic header
- Then the stories under that epic collapse/expand
- And the chevron icon toggles direction

## BDD Test Scenarios

### Scenario 1: Sprint panel rendering

```gherkin
Feature: Sprint panel in sidebar

  Scenario: Display sprint data from YAML
    Given WorkflowBuilder loaded with repoPath pointing to a repo with sprint-status.yaml
    And the YAML contains epic-1 with 3 stories and epic-2 with 2 stories
    When the user clicks the Sprint tab
    Then 2 epic sections are visible
    And epic-1 shows 3 story items
    And epic-2 shows 2 story items

  Scenario: Status colors match story status
    Given sprint data with stories: "1-1" (done), "1-2" (ready-for-dev), "1-3" (backlog)
    When rendered in the Sprint panel
    Then "1-1" has a green status badge
    And "1-2" has a blue status badge
    And "1-3" has a gray status badge

  Scenario: Drag story onto canvas creates node
    Given the Sprint panel shows story "1-2-account-mgmt" with status "ready-for-dev"
    When the user drags "1-2-account-mgmt" onto the canvas at position (300, 200)
    Then a new node appears at approximately (300, 200)
    And the node label contains "1-2-account-mgmt"
    And the node data has storyId "1-2-account-mgmt"

  Scenario: Empty sprint state
    Given WorkflowBuilder loaded with repoPath pointing to a repo without sprint-status.yaml
    When the user clicks the Sprint tab
    Then a message "No sprint data found" is displayed
    And the panel is otherwise empty

  Scenario: Collapse and expand epic section
    Given the Sprint panel shows epic-1 with 3 stories expanded
    When the user clicks the epic-1 header
    Then the 3 stories are hidden
    And clicking the header again shows them
```

## Tasks / Subtasks

- [ ] Task 1: Create SprintPanel component (AC: AC-2, AC-3, AC-5, AC-6)
  - [ ] Subtask 1a: Create `frontend/src/components/bmad/SprintPanel.svelte` with epic/story rendering
  - [ ] Subtask 1b: Implement status-to-color mapping using CSS variables
  - [ ] Subtask 1c: Add collapsible epic sections with chevron toggle (reuse ProcessSidebar pattern)
  - [ ] Subtask 1d: Add empty state rendering when sprintStatus is null or has no epics
  - [ ] Subtask 1e: Add story count badge per epic (done/total format)

- [ ] Task 2: Add drag support to SprintPanel (AC: AC-4)
  - [ ] Subtask 2a: Make story items draggable with `application/bmad-story` data transfer type
  - [ ] Subtask 2b: Serialize story data (storyId, epicId, status) as JSON in drag data

- [ ] Task 3: Add Sprint tab to ProcessSidebar (AC: AC-1)
  - [ ] Subtask 3a: Add `sprintStatus` prop to ProcessSidebar
  - [ ] Subtask 3b: Add 4th "Sprint" tab button and conditional rendering of SprintPanel
  - [ ] Subtask 3c: Forward `drag-story` events from SprintPanel

- [ ] Task 4: Wire sprint data in WorkflowBuilder (AC: AC-2, AC-4)
  - [ ] Subtask 4a: Import and call `GetSprintStatus(repoPath)` in onMount
  - [ ] Subtask 4b: Pass `sprintStatus` to ProcessSidebar
  - [ ] Subtask 4c: Handle `application/bmad-story` drops in the canvas drop handler
  - [ ] Subtask 4d: Create story-linked nodes with default "bmad-dev-story" process

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
