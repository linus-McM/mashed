# Story 8: Repo Context Header Bar

**Priority:** P3-low
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** Story 4, Story 5
**Status:** done

## Description

Add a context header at the top of the WorkflowBuilder that shows which repo is active, its current branch, and sprint progress. This gives the user constant visibility into their repo context without needing to check the sidebar or remember which repo they selected. It also provides a quick sprint health summary (stories done / total).

## Developer Notes

### Architecture
- **New file:** `frontend/src/components/bmad/RepoContextBar.svelte` -- horizontal bar above the toolbar
- **Modify:** `frontend/src/views/WorkflowBuilder.svelte` -- render RepoContextBar above the canvas area
- No backend changes required

### RepoContextBar.svelte Design

```
[GitBranch icon] my-project  |  main  |  Sprint: 3/8 stories done  [progress bar]
```

Layout:
- Horizontal flex bar, full width, thin (28px height)
- Background: `var(--bg-deepest)` with bottom border
- Left: folder/git icon + repo name (bold) + branch name (dim)
- Right: sprint progress summary + thin progress bar
- Font: `var(--font-mono)` at 10px (matches existing chrome)

Props:
```svelte
export let repoPath = '';
export let sprintStatus = null;

$: repoName = repoPath ? repoPath.split('/').pop() : 'No repo';
$: branch = ''; // Could be fetched, or passed from parent
$: totalStories = sprintStatus?.epics?.reduce((sum, e) => sum + (e.stories?.length || 0), 0) || 0;
$: doneStories = sprintStatus?.epics?.reduce((sum, e) => sum + (e.stories?.filter(s => s.status === 'done').length || 0), 0) || 0;
$: progressPct = totalStories > 0 ? (doneStories / totalStories) * 100 : 0;
```

### WorkflowBuilder.svelte Integration

```svelte
<div class="canvas-area">
  <RepoContextBar {repoPath} {sprintStatus} />
  <div class="toolbar">
    <!-- existing toolbar -->
  </div>
  <!-- rest of canvas area -->
</div>
```

### Technical Considerations
- The branch could be extracted from `ListRepoChoices()` data (which includes branch) -- pass it as a separate prop from App.svelte, or have the context bar call the API itself
- Sprint progress should update reactively when `sprintStatus` changes (e.g., after story status updates from Story 7)
- The progress bar reuses the same style as ExecutionBar's progress bar for consistency
- Keep the bar very thin (28px) to avoid eating into canvas space

### Risks & Edge Cases
- Repo name could be very long -- truncate with ellipsis
- Sprint status might be null (no YAML) -- show "No sprint data" instead of progress
- Branch info might not be available -- show repo name only as fallback

### Reference Files
- `frontend/src/components/bmad/ExecutionBar.svelte` -- progress bar styling to reuse
- `frontend/src/views/WorkflowBuilder.svelte` -- canvas-area layout to insert into
- `frontend/src/components/TitleBar.svelte` -- thin bar styling pattern

## Acceptance Criteria

AC-1: Context bar shows repo name
- Given WorkflowBuilder loaded with repoPath "/Users/linus/Development/my-project"
- When the context bar renders
- Then it displays "my-project" as the repo name
- And the full path is available in a tooltip

AC-2: Sprint progress is displayed
- Given sprint data with 8 total stories and 3 done
- When the context bar renders
- Then it shows "3/8 stories done" (or similar)
- And a progress bar shows ~37.5% filled

AC-3: No sprint data shows fallback
- Given sprintStatus is null (no YAML file)
- When the context bar renders
- Then the sprint section shows "No sprint data"
- And no progress bar is shown

AC-4: Context bar is positioned above toolbar
- Given the WorkflowBuilder layout
- When rendered
- Then the context bar appears between the title bar and the toolbar
- And it does not overlap with the canvas or sidebar

## BDD Test Scenarios

### Scenario 1: Repo context bar rendering

```gherkin
Feature: Repo context header bar

  Scenario: Display repo name and sprint progress
    Given RepoContextBar rendered with repoPath "/Users/linus/Dev/my-project"
    And sprintStatus with 2 epics containing 5 total stories, 2 done
    Then the text "my-project" is visible
    And the text "2/5 stories done" is visible
    And the progress bar is approximately 40% filled

  Scenario: No sprint data fallback
    Given RepoContextBar rendered with repoPath "/some/repo" and sprintStatus null
    Then the text "some-repo" or "repo" is visible
    And the text "No sprint data" is visible
    And no progress bar is rendered

  Scenario: Long repo name is truncated
    Given RepoContextBar with repoPath "/Users/linus/Development/my-extremely-long-project-name-that-goes-on"
    Then the repo name is truncated with ellipsis
    And the full path is in the title tooltip

  Scenario: Progress updates reactively
    Given RepoContextBar showing "2/5 stories done"
    When sprintStatus is updated to have 3 done stories
    Then the text changes to "3/5 stories done"
    And the progress bar width increases
```

## Tasks / Subtasks

- [ ] Task 1: Create RepoContextBar component (AC: AC-1, AC-2, AC-3)
  - [ ] Subtask 1a: Create `frontend/src/components/bmad/RepoContextBar.svelte`
  - [ ] Subtask 1b: Implement repo name extraction from path with tooltip
  - [ ] Subtask 1c: Compute sprint progress (done/total) from sprintStatus prop
  - [ ] Subtask 1d: Render progress bar matching ExecutionBar style
  - [ ] Subtask 1e: Handle null sprintStatus with "No sprint data" fallback

- [ ] Task 2: Integrate into WorkflowBuilder layout (AC: AC-4)
  - [ ] Subtask 2a: Import and render RepoContextBar in WorkflowBuilder.svelte
  - [ ] Subtask 2b: Position above the toolbar div within canvas-area
  - [ ] Subtask 2c: Pass repoPath and sprintStatus props

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
