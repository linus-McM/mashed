# Story 6: GitPanel Integration & Wiring

**Priority:** P1-high
**Domain:** fullstack
**Estimated Complexity:** S
**Depends On:** review-03
**Status:** done

## Description

Wire the SummarisationModal into the existing GitPanel by adding a "Summarise" button in the actions section and mounting the modal conditionally. This is the final integration step that makes the entire code review summarisation feature accessible to users through the existing Git workflow.

## Developer Notes

### Architecture

- **Modified file:** `frontend/src/components/bmad/GitPanel.svelte`
  - Add import for `SummarisationModal`
  - Add `summariseModalOpen` boolean state
  - Add "Summarise" button in `.actions` div
  - Add conditional modal mount at bottom of template

- **Modified file:** `frontend/src/views/SummarisationModal.svelte` (if needed)
  - Ensure `close` event dispatch is correct for GitPanel consumption

### Frontend Changes

**GitPanel.svelte additions:**

1. **Import:**
```javascript
import SummarisationModal from '../../views/SummarisationModal.svelte';
import { FileSearch } from 'lucide-svelte';
```

2. **State:**
```javascript
let summariseModalOpen = false;
```

3. **Button (add after the Review button in `.actions`):**
```svelte
<button
  class="action-btn"
  class:hot={status.dirty}
  disabled={!!actionState.action}
  on:click={() => summariseModalOpen = true}
  title="AI-powered code review summary"
>
  <FileSearch size={14} />
  <span>Summarise</span>
</button>
```

4. **Modal mount (add after the ForcePushModal block):**
```svelte
{#if summariseModalOpen}
  <SummarisationModal
    {repoPath}
    on:close={() => summariseModalOpen = false}
    on:open-file
  />
{/if}
```

### Technical Considerations

- **Event forwarding:** The `on:open-file` without a handler forwards the event to GitPanel's parent (ProcessSidebar or WorkflowBuilder), which handles file opening in the editor
- **Button placement:** Add the Summarise button between the "PR" button and the "Review" button, or after "Review" -- it's a review-adjacent action
- **Hot state:** The button glows green (`class:hot`) when the repo is dirty, since that's when a review is most useful
- **Disabled state:** Follows existing pattern -- disabled when any other action is in progress
- **Icon choice:** `FileSearch` from lucide -- represents searching/analysing files. Alternatives: `ScanSearch`, `FileText`, `ClipboardList`
- **Wails binding regeneration:** After adding `SpawnRefactorPlan`, `ListAdviceModes`, `StreamCodeReviewSummary`, and `StreamAdvice` to `app_review.go`, run `wails dev` or `wails generate module` to regenerate `frontend/wailsjs/go/main/App.js` and `.d.ts` files

### Risks & Edge Cases

- **Binding generation:** The new Go methods won't appear in `App.js` until Wails regenerates bindings. Engineers must run `wails dev` after adding backend methods.
- **Event bubbling:** `on:open-file` must bubble up from SummarisationModal through GitPanel to the parent component that handles file opening
- **Modal z-index:** Ensure the SummarisationModal's overlay z-index is high enough to appear above the GitPanel sidebar content

### Reference Files

- `frontend/src/components/bmad/GitPanel.svelte` -- the file being modified (current button pattern)
- `frontend/src/views/SummarisationModal.svelte` -- the modal being mounted (from review-03)
- `frontend/src/views/MergeModal.svelte` -- pattern for modal mounting from GitPanel
- `frontend/wailsjs/go/main/App.js` -- binding file that will need regeneration

## Acceptance Criteria

AC-1: Summarise button visible
- Given the GitPanel is rendered for a repository
- When the user views the actions section
- Then a "Summarise" button is visible with a search/file icon
- And it is positioned among the other git action buttons

AC-2: Button opens modal
- Given the GitPanel shows the Summarise button
- When the user clicks the button
- Then the SummarisationModal opens with the current `repoPath`

AC-3: Button hot state on dirty repo
- Given the repository has uncommitted changes (dirty state)
- When the GitPanel renders
- Then the Summarise button has the `hot` CSS class (green glow)

AC-4: Button disabled during actions
- Given another git action is in progress (e.g., commit, push)
- When the action is running
- Then the Summarise button is disabled

AC-5: Modal close resets state
- Given the SummarisationModal is open
- When the user closes the modal
- Then `summariseModalOpen` is set to false
- And the modal is unmounted from the DOM

AC-6: File open event bubbles
- Given the user clicks a file path in the SummarisationModal
- When the `open-file` event is dispatched
- Then it bubbles through GitPanel to the parent component

## BDD Test Scenarios

### Scenario 1: Button rendering

```gherkin
Feature: GitPanel Summarise button

  Scenario: Button appears in actions
    Given a GitPanel with repoPath "/Users/dev/project"
    When the panel renders
    Then a button with text "Summarise" is visible in the .actions section
    And the button has a FileSearch icon

  Scenario: Button hot when dirty
    Given the repo status has dirty: true
    When the GitPanel renders
    Then the Summarise button has class "hot"

  Scenario: Button normal when clean
    Given the repo status has dirty: false
    When the GitPanel renders
    Then the Summarise button does not have class "hot"
```

### Scenario 2: Modal lifecycle

```gherkin
Feature: Modal open/close from GitPanel

  Scenario: Open modal on click
    Given summariseModalOpen is false
    When the user clicks the Summarise button
    Then summariseModalOpen is set to true
    And SummarisationModal component is rendered in the DOM

  Scenario: Close modal via event
    Given summariseModalOpen is true and the modal is rendered
    When the modal dispatches a "close" event
    Then summariseModalOpen is set to false
    And SummarisationModal is removed from the DOM

  Scenario: Button disabled during commit
    Given actionState.action is "commit"
    Then the Summarise button has disabled attribute
```

### Scenario 3: Event forwarding

```gherkin
Feature: Event forwarding

  Scenario: Open-file event bubbles up
    Given the SummarisationModal dispatches open-file with path "src/main.go"
    Then the GitPanel forwards the event to its parent
    And the parent can handle file opening
```

## Tasks / Subtasks

- [ ] Task 1: Add button to GitPanel (AC: AC-1, AC-3, AC-4)
  - [ ] Import `SummarisationModal` and `FileSearch` icon
  - [ ] Add `summariseModalOpen` state variable
  - [ ] Add button markup in `.actions` section following existing pattern
  - [ ] Apply `class:hot={status.dirty}` and `disabled={!!actionState.action}`

- [ ] Task 2: Mount modal conditionally (AC: AC-2, AC-5, AC-6)
  - [ ] Add `{#if summariseModalOpen}` block after other modals
  - [ ] Pass `repoPath` prop and bind `on:close` handler
  - [ ] Forward `on:open-file` event to parent

- [ ] Task 3: Verify Wails bindings (AC: AC-2)
  - [ ] Run `wails dev` to regenerate bindings after backend changes
  - [ ] Verify `StreamCodeReviewSummary`, `ListAdviceModes`, `StreamAdvice`, `SpawnRefactorPlan` appear in `App.js`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] `go build ./...` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Wails bindings regenerated and committed
