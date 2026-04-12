# Story 3: Summarisation Modal Frontend

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** L
**Depends On:** review-02
**Status:** done

## Description

Create the `SummarisationModal.svelte` component -- the main UI surface for the code review summarisation feature. The modal displays per-file change summaries (streamed in real-time), a methodology advice dropdown, a streaming advice panel, and a "Create Refactor Plan" button. This is the user-facing experience that ties together all backend review capabilities.

## Developer Notes

### Architecture

- **New file:** `frontend/src/views/SummarisationModal.svelte`
- **Modified file:** `frontend/src/components/bmad/GitPanel.svelte` -- add Summarise button and modal mount

The modal follows the existing modal pattern used by `MergeModal.svelte`, `BranchModal.svelte`, etc.:
- Props: `repoPath`, event dispatchers for `close`
- Overlay + centered card layout
- Uses `EventsOn`/`EventsOff` from Wails runtime for event subscriptions
- Calls Go bindings from `wailsjs/go/main/App.js`

### Component Structure

```
SummarisationModal.svelte
  |-- Header: "Code Review Summary" + close button
  |-- File Cards Section: scrollable list of FileSummary cards
  |    |-- Each card: file path (clickable), +/- counts, summary text
  |    |-- Loading spinner while streaming
  |-- Advice Section:
  |    |-- Dropdown: select advice mode (populated from ListAdviceModes)
  |    |-- "Get Advice" button
  |    |-- Streaming advice text panel (monospace, scrolls to bottom)
  |-- Footer:
  |    |-- "Create Refactor Plan" button (disabled until advice is loaded)
  |    |-- Close button
```

### Data Flow

1. **onMount:** Call `StreamCodeReviewSummary(repoPath)` to start per-file summarisation
2. **EventsOn("review:summary:progress"):** Push each `FileSummary` to the reactive `files` array
3. **EventsOn("review:summary:done"):** Mark summarisation complete, show totals
4. **Dropdown populated:** Call `ListAdviceModes(repoPath)` on mount, populate `<select>`
5. **User clicks "Get Advice":** Call `StreamAdvice(repoPath, selectedMode)`
6. **EventsOn("review:advice:progress"):** Append text to advice panel, auto-scroll
7. **User clicks "Create Refactor Plan":** Call `SpawnRefactorPlan(repoPath, adviceText)` (story review-04)

### Technical Considerations

- **Imports from Wails bindings:** `StreamCodeReviewSummary`, `ListAdviceModes`, `StreamAdvice`, `SpawnRefactorPlan` from `../../../wailsjs/go/main/App.js`
- **Event subscription cleanup:** Must call `EventsOff` for all three event names in `onDestroy`
- **File card click:** Dispatch event `open-file` with `{ path: file.path }` -- the parent `ProcessSidebar` or `WorkflowBuilder` handles opening in the editor
- **CSS variables:** Use existing design system variables: `--bg-elevated`, `--bg-surface`, `--border-subtle`, `--accent-green`, `--text-primary`, `--text-dim`, `--text-muted`, `--font-mono`, `--radius-md`, `--sp-md`, `--sp-sm`
- **Advice panel:** Use `<pre>` or `<div style="white-space: pre-wrap">` for the streaming advice text. Auto-scroll via `scrollTop = scrollHeight` after each append.
- **Progress indicator:** Show "Summarising file N of M..." while streaming, with a progress bar or counter
- **Icons:** Use `lucide-svelte` -- `FileText`, `Plus`, `Minus`, `ChevronDown`, `Sparkles`, `FileCode` (or similar)
- **Overlay pattern:** Same as `MergeModal` -- fixed overlay with `z-index` above content, click-outside-to-close

### Risks & Edge Cases

- **No changes:** If `review:summary:done` comes with empty files array, show "No changes to review" message
- **Claude CLI unavailable:** If error event comes, show error banner with "Claude CLI not found" message
- **Long file paths:** Truncate with ellipsis in the card, full path in tooltip
- **Large number of files:** Scrollable container with max-height, virtual scrolling not needed (50 file cap in backend)
- **Advice streaming interrupted:** Handle `done: true` with error -- show error inline in advice panel
- **Modal escape key:** Add keydown handler for Escape to close modal

### Reference Files

- `frontend/src/views/MergeModal.svelte` -- modal overlay pattern, props, event dispatching
- `frontend/src/views/BranchModal.svelte` -- modal with dynamic content loading on mount
- `frontend/src/components/bmad/GitPanel.svelte` -- existing button patterns, event subscription pattern
- `frontend/src/components/bmad/OutputViewerModal.svelte` -- modal for viewing streaming output

## Acceptance Criteria

AC-1: Modal opens and starts summarisation
- Given the user clicks the "Summarise" button in the Git panel
- When the SummarisationModal mounts
- Then `StreamCodeReviewSummary` is called with the current repo path
- And a loading state is shown ("Summarising...")

AC-2: File cards display with streaming updates
- Given summarisation is in progress
- When a `review:summary:progress` event is received
- Then a new file card appears showing the file path, +/- line counts, and the AI-generated summary
- And the progress counter updates (e.g., "3 of 7 files")

AC-3: Advice mode dropdown populated
- Given the modal is mounted
- When `ListAdviceModes` returns the available modes
- Then the advice dropdown contains all modes sorted by order
- And each option shows the mode's `displayName` and `icon`

AC-4: Advice streaming display
- Given the user selects an advice mode and clicks "Get Advice"
- When `review:advice:progress` events are received
- Then the advice panel shows streaming text appended in real-time
- And the panel auto-scrolls to the bottom

AC-5: File path click emits open-file event
- Given a file card is displayed for path "internal/advice/loader.go"
- When the user clicks the file path
- Then an `open-file` event is dispatched with the file path

AC-6: Error display
- Given Claude CLI is not available
- When the summary streaming returns an error
- Then an error banner is shown with the error message
- And the modal remains usable (can be closed)

AC-7: Modal cleanup on close
- Given the modal is open with active event subscriptions
- When the user closes the modal (close button, overlay click, or Escape key)
- Then all `EventsOff` calls are made for subscribed events
- And the modal is unmounted

## BDD Test Scenarios

### Scenario 1: Full summarisation flow

```gherkin
Feature: Summarisation modal

  Scenario: Open modal and see streamed file summaries
    Given the GitPanel has repoPath "/Users/dev/myproject"
    And the repo has 3 changed files
    When the user clicks the "Summarise" button
    Then the SummarisationModal opens
    And shows "Summarising..." loading state
    And after 3 progress events, 3 file cards are displayed
    And each card shows file path, green +N / red -N counts, and summary text

  Scenario: No changes to review
    Given the repo has no uncommitted changes
    When the modal opens and receives summary:done with empty files
    Then the modal shows "No changes to review" message
    And the advice section is hidden
```

### Scenario 2: Advice interaction

```gherkin
Feature: Methodology advice

  Scenario: Select and stream advice
    Given the modal has loaded 5 advice modes in the dropdown
    And summarisation is complete
    When the user selects "Clean Code" from the dropdown
    And clicks "Get Advice"
    Then StreamAdvice is called with repoPath and "clean-code"
    And the advice panel shows streaming text as events arrive
    And the panel scrolls to bottom automatically

  Scenario: Switch advice mode mid-stream
    Given advice is currently streaming for "solid"
    When the user selects "security-first" and clicks "Get Advice"
    Then the advice panel is cleared
    And new streaming starts for "security-first"
```

### Scenario 3: Modal lifecycle

```gherkin
Feature: Modal lifecycle

  Scenario: Close via Escape key
    Given the summarisation modal is open
    When the user presses Escape
    Then the modal closes
    And EventsOff is called for "review:summary:progress", "review:summary:done", "review:advice:progress"

  Scenario: Close via overlay click
    Given the summarisation modal is open
    When the user clicks the overlay outside the modal card
    Then the modal closes

  Scenario: Error state display
    Given Claude CLI is not installed
    When StreamCodeReviewSummary emits an error event
    Then an error banner shows "Claude CLI not found"
    And the close button remains functional
```

## Tasks / Subtasks

- [ ] Task 1: Create modal shell and layout (AC: AC-1, AC-7)
  - [ ] Create `frontend/src/views/SummarisationModal.svelte` with overlay + card layout
  - [ ] Add props: `repoPath`, event dispatcher for `close`
  - [ ] Implement close handlers: button, overlay click, Escape key
  - [ ] Wire up `onDestroy` with `EventsOff` cleanup

- [ ] Task 2: File cards with streaming (AC: AC-2, AC-5, AC-6)
  - [ ] Call `StreamCodeReviewSummary(repoPath)` in `onMount`
  - [ ] Subscribe to `review:summary:progress` -- push to reactive `files` array
  - [ ] Subscribe to `review:summary:done` -- set complete state, show totals
  - [ ] Render file cards with path (clickable), +/- counts, summary
  - [ ] Implement `open-file` event dispatch on path click
  - [ ] Handle error events with error banner
  - [ ] Handle empty files with "No changes" message

- [ ] Task 3: Advice dropdown and streaming panel (AC: AC-3, AC-4)
  - [ ] Call `ListAdviceModes(repoPath)` in `onMount`, populate dropdown
  - [ ] Add "Get Advice" button with selected mode binding
  - [ ] Subscribe to `review:advice:progress` -- append to advice text
  - [ ] Implement auto-scroll on advice panel
  - [ ] Handle clearing advice when switching modes

- [ ] Task 4: Styling (AC: AC-1 through AC-7)
  - [ ] Style modal overlay and card following existing modal patterns
  - [ ] Style file cards with monospace font, colored +/- counts
  - [ ] Style advice panel with pre-wrap text, scrollable container
  - [ ] Style dropdown and buttons following action-btn pattern
  - [ ] Add loading spinner/animation during streaming
  - [ ] Ensure responsive layout within the sidebar width

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] Modal matches existing design system (CSS variables, spacing, typography)
- [ ] `go build ./...` passes (Wails binding generation)
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
