# Story 2: Frontend Menu Event Handlers & Navigation

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** menu-01
**Status:** done

## Description

Wire up the frontend to respond to native menu events emitted by the Go backend. When the user clicks a menu item like "Settings..." or "Feed", the backend emits `menu:navigate` with a route string -- the frontend must listen for this event and switch the active view accordingly. This story also adds the `menu:about` event handler (showing an about modal) and the `screenshot:taken` handler (showing a toast notification), and removes the now-redundant `handleKeydown` Cmd+N shortcut that conflicts with the native menu accelerator.

## Developer Notes

### Architecture
- **Modified file:** `frontend/src/App.svelte` -- add three `EventsOn` listeners in the `<script>` block:
  1. `EventsOn('menu:navigate', (route) => { ... })` -- switches `currentView` based on route string
  2. `EventsOn('menu:about', () => { ... })` -- sets `showAboutModal = true`
  3. `EventsOn('screenshot:taken', (path) => { ... })` -- shows a toast/notification with the path
  4. `EventsOn('menu:open-workspace', (path) => { ... })` -- triggers workspace opening flow
- **New file:** `frontend/src/components/AboutModal.svelte` -- simple modal showing app name, version, and description.
- **Modified file:** `frontend/src/App.svelte` -- remove the Cmd+N handling from `handleKeydown` (lines 127-131 approx), since the native menu now owns Cmd+N via the File > New Agent accelerator.

### Technical Considerations
- **Route mapping** for `menu:navigate`:
  - `"settings"` -> `currentView = 'settings'`
  - `"spawn"` -> `showSpawnModal = true`
  - `"new-repo"` -> `showNewRepoModal = true`
  - `"feed"` -> `currentView = 'feed'`
  - `"workflows"` -> `currentView = 'workflows'` (needs a default repo context -- use empty string or last-used)
- **EventsOn import**: Already imported at line 3 of `App.svelte`.
- **About modal**: Minimal component -- dark background overlay, centered card with app name "mashed", description "Notification-first IDE for multi-agent development", and a Close button. Use existing design system variables (`--bg-surface`, `--text-primary`, `--accent-green`, `--border-subtle`).
- **Screenshot toast**: Use a temporary notification that auto-dismisses after 3 seconds. Show the filename (not full path) and a checkmark icon. Implement as a simple reactive variable `toastMessage` with a `setTimeout` to clear it.
- **Workflow view without repo**: When `menu:navigate` sends `"workflows"` but there's no repo context, navigate to `feed` instead (workflows requires `builderRepoPath`). Alternatively, show the workflow list without a specific repo.

### Risks & Edge Cases
- **Race condition**: Menu events fire from Go's main thread. `EventsOn` callbacks run in the browser's JS event loop. No race condition risk -- Wails serializes events to the frontend.
- **Duplicate Cmd+N**: If the `handleKeydown` Cmd+N block is NOT removed, both the native menu callback AND the JS handler will fire. The native menu callback fires first and emits the event, then the JS handler also triggers `showSpawnModal = true`. This causes the modal to open twice (or flash). Must remove the JS handler.
- **About modal styling**: Must match the existing modal patterns (`NewRepoModal`, `SpawnAgent`). Check those components for overlay and card CSS patterns.
- **View state conflicts**: If the user is in a modal (SpawnAgent, NewRepoModal) and hits Cmd+1 (Feed), the modal should close and navigate. Handle this in the `menu:navigate` handler by setting all modal flags to false.

### Reference Files
- `frontend/src/App.svelte` -- current event listeners, view routing, modal state
- `frontend/src/components/NewRepoModal.svelte` -- modal styling pattern to follow
- `frontend/src/views/SpawnAgent.svelte` -- modal/overlay pattern
- `frontend/src/views/Settings.svelte` -- view pattern for back navigation

## Acceptance Criteria

AC-1: menu:navigate event switches views correctly
- Given the app is displaying the feed view
- When the backend emits `menu:navigate` with `"settings"`
- Then `currentView` changes to `"settings"` and the Settings view is displayed
- And the same works for `"feed"`, `"workflows"`, `"spawn"` (opens spawn modal), `"new-repo"` (opens new repo modal)

AC-2: menu:about event shows the About modal
- Given the app is running in any view
- When the backend emits `menu:about`
- Then an About modal is displayed showing "mashed" and the app description
- And clicking Close or pressing Escape dismisses the modal

AC-3: screenshot:taken event shows a toast notification
- Given the app is running
- When the backend emits `screenshot:taken` with path `/Users/me/Desktop/mashed-screenshot-20260408-143022.png`
- Then a toast notification appears showing "Screenshot saved: mashed-screenshot-20260408-143022.png"
- And the toast auto-dismisses after 3 seconds

AC-4: Cmd+N no longer double-triggers from JavaScript handler
- Given the app is on the feed view
- When the user presses Cmd+N
- Then the spawn modal opens exactly once (via the native menu event, not the JS keydown handler)
- And the `handleKeydown` function no longer contains Cmd+N handling code

AC-5: Navigation closes open modals
- Given the spawn modal is open
- When the backend emits `menu:navigate` with `"settings"`
- Then the spawn modal closes
- And the settings view is displayed

## BDD Test Scenarios

### Scenario 1: View navigation via menu events

```gherkin
Feature: Menu Event Navigation

  Scenario: Navigate to settings via menu
    Given the app is on the "feed" view
    When the "menu:navigate" event fires with "settings"
    Then the current view is "settings"

  Scenario: Navigate to feed via menu
    Given the app is on the "settings" view
    When the "menu:navigate" event fires with "feed"
    Then the current view is "feed"

  Scenario: Open spawn modal via menu
    Given the app is on the "feed" view
    And showSpawnModal is false
    When the "menu:navigate" event fires with "spawn"
    Then showSpawnModal is true

  Scenario: Open new repo modal via menu
    Given the app is on the "feed" view
    And showNewRepoModal is false
    When the "menu:navigate" event fires with "new-repo"
    Then showNewRepoModal is true

  Scenario: Navigate to workflows without repo context
    Given the app is on the "feed" view
    And builderRepoPath is empty
    When the "menu:navigate" event fires with "workflows"
    Then the current view remains "feed"
```

### Scenario 2: About modal

```gherkin
Feature: About Modal

  Scenario: Show about modal
    Given the app is running
    When the "menu:about" event fires
    Then the About modal is visible
    And it displays "mashed"

  Scenario: Close about modal with Escape
    Given the About modal is visible
    When the user presses Escape
    Then the About modal is hidden
```

### Scenario 3: Screenshot toast

```gherkin
Feature: Screenshot Toast

  Scenario: Show screenshot toast
    Given the app is running
    When the "screenshot:taken" event fires with "/Users/me/Desktop/mashed-screenshot-20260408-143022.png"
    Then a toast displays "Screenshot saved: mashed-screenshot-20260408-143022.png"

  Scenario: Toast auto-dismisses
    Given a screenshot toast is visible
    When 3 seconds elapse
    Then the toast is no longer visible
```

### Scenario 4: Cmd+N conflict removal

```gherkin
Feature: Keyboard Shortcut Deduplication

  Scenario: Cmd+N handled by native menu only
    Given the app is on the "feed" view
    When the user presses Cmd+N
    Then the native menu emits "menu:navigate" with "spawn"
    And the handleKeydown function does NOT call showSpawnModal = true directly
```

## Tasks / Subtasks

- [ ] Task 1: Add `menu:navigate` event handler to `App.svelte` (AC: AC-1, AC-5)
  - [ ] Subtask 1a: Add `EventsOn('menu:navigate', ...)` with route-to-view mapping
  - [ ] Subtask 1b: Close all open modals before navigating (`showSpawnModal = false`, `showNewRepoModal = false`, `showAboutModal = false`)
  - [ ] Subtask 1c: Handle `"workflows"` route -- skip if no `builderRepoPath`

- [ ] Task 2: Create `AboutModal.svelte` and wire `menu:about` event (AC: AC-2)
  - [ ] Subtask 2a: Create `frontend/src/components/AboutModal.svelte` with app name, description, close button
  - [ ] Subtask 2b: Add `showAboutModal` state and `EventsOn('menu:about', ...)` in `App.svelte`
  - [ ] Subtask 2c: Render `AboutModal` conditionally and handle Escape key

- [ ] Task 3: Add screenshot toast and `screenshot:taken` handler (AC: AC-3)
  - [ ] Subtask 3a: Add `toastMessage` reactive variable and `EventsOn('screenshot:taken', ...)`
  - [ ] Subtask 3b: Extract filename from path and display toast
  - [ ] Subtask 3c: Auto-dismiss toast after 3 seconds with `setTimeout`

- [ ] Task 4: Remove Cmd+N from `handleKeydown` (AC: AC-4)
  - [ ] Subtask 4a: Remove the `if ((e.ctrlKey || e.metaKey) && e.key === 'n' ...)` block from `handleKeydown`
  - [ ] Subtask 4b: Verify Escape key handling still works for all modal/view states

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
