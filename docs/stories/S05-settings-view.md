# Story 5: Settings View

**Priority:** P2-medium
**Domain:** fullstack
**Estimated Complexity:** L
**Depends On:** Story 1, Story 3, Story 4
**Status:** ready

## Description

Create a full-page Settings view accessible via the gear icon in the TitleBar. The view includes a theme picker (larger cards with preview), a VSCodium extension path input with Browse button, and an About section. It integrates with the existing view routing system in App.svelte and supports navigation back to the feed via a back button or Escape key.

## Developer Notes

### Architecture

- **New file: `frontend/src/views/Settings.svelte`** -- Full-page settings view component. Uses `createEventDispatcher` to dispatch `'back'` event, same pattern as `AgentDetail.svelte` (line 2: `import { createEventDispatcher }`).
- **Modified file: `frontend/src/App.svelte`** -- Add `'settings'` to the `currentView` state machine (line 12), handle `on:open-settings` from TitleBar, render `<Settings on:back={goBack} />` in the template.

### View Routing in App.svelte

Current `currentView` states (line 12): `'loading' | 'setup' | 'feed' | 'detail'`

Add `'settings'`:
```js
let currentView = 'loading'; // 'loading' | 'setup' | 'feed' | 'detail' | 'settings'
```

TitleBar event handler:
```html
<TitleBar on:open-settings={() => currentView = 'settings'} />
```

Template addition (after the `AgentDetail` else block, around line 121):
```html
{:else if currentView === 'settings'}
  <Settings on:back={goBack} />
```

Escape key handling (line 86-99) -- add settings view:
```js
function handleKeydown(e) {
  if (e.key === 'Escape') {
    if (showSpawnModal) {
      showSpawnModal = false;
    } else if (currentView === 'detail' || currentView === 'settings') {
      goBack();
    }
  }
  // ...
}
```

### Settings.svelte Structure

```html
<script>
  import { createEventDispatcher, onMount } from 'svelte';
  import { ArrowLeft } from 'lucide-svelte';
  import { themes } from '../lib/themes.js';
  import { currentThemeId, applyTheme } from '../lib/stores/theme.js';
  import { GetConfig, SetTheme, SetVSCodiumExtPath, PickDirectory } from '../../wailsjs/go/main/App.js';

  const dispatch = createEventDispatcher();

  let vscodiumPath = '';
  let saveStatus = '';

  onMount(async () => {
    try {
      const cfg = await GetConfig();
      vscodiumPath = cfg.vscodiumExtPath || '';
    } catch (e) {
      console.error('Failed to load config:', e);
    }
  });

  async function selectTheme(id) {
    applyTheme(id);
    try {
      await SetTheme(id);
    } catch (e) {
      console.error('Failed to persist theme:', e);
    }
  }

  async function browseVSCodium() {
    try {
      const dir = await PickDirectory();
      if (dir) {
        vscodiumPath = dir;
        await SetVSCodiumExtPath(dir);
        saveStatus = 'saved';
        setTimeout(() => saveStatus = '', 2000);
      }
    } catch (e) {
      console.error('Failed to pick directory:', e);
    }
  }

  async function saveVSCodiumPath() {
    try {
      await SetVSCodiumExtPath(vscodiumPath);
      saveStatus = 'saved';
      setTimeout(() => saveStatus = '', 2000);
    } catch (e) {
      saveStatus = 'error';
    }
  }
</script>
```

### Settings.svelte Layout

Three sections:

1. **Header** -- Back button (ArrowLeft icon) + "Settings" title. Same pattern as `AgentDetail.svelte` header.

2. **Theme section** -- Grid of theme cards, larger than the popover version. Each card shows:
   - Theme label
   - 6-8 color swatches in a grid
   - A small preview rectangle showing bg + text + accent colors
   - Active indicator (green border)

3. **VSCodium Extension section** -- Text input for path + Browse button + Save button.
   - The Browse button reuses `PickDirectory()` from `app.go` (line 499) -- this is the same native OS dialog used in the Setup view.
   - Manual path entry is also supported with a Save button.

4. **About section** -- App name, version, links. Static content.

### Styling Guidelines

Follow the patterns in `AgentDetail.svelte` for the page layout:
- Full-height scrollable container
- Header with back button on the left
- Sections with headings using `var(--text-section)` (16px) size
- Labels using `var(--text-label)` (11px) in `var(--text-dim)`
- Input fields: `var(--bg-elevated)` background, `var(--border-subtle)` border, `var(--font-mono)` for paths
- Buttons: match the existing mode-toggle button styling from `MonacoEditor.svelte` lines 520-538
- Section spacing: `var(--sp-xl)` (24px) between sections
- Theme cards: wider than popover cards, arranged in a flex row with `gap: var(--sp-md)` (12px)

### Technical Considerations

- `PickDirectory()` returns a string or empty string if the user cancels. It's already Wails-bound (line 499 of `app.go`).
- The VSCodium path input should debounce saves or use an explicit Save button (explicit Save is cleaner for file paths).
- The Settings view should load the current config on mount to populate the VSCodium path field.
- The `on:back` event is dispatched and caught by App.svelte's `goBack()` function, which sets `currentView = 'feed'` (line 62-64).
- Theme cards in Settings should use the same `selectTheme` logic as the popover (Story 4), sharing the same store and backend calls.

### Risks & Edge Cases

- If `GetConfig()` fails on mount (e.g., disk read error), the Settings view should still render with empty/default values and not crash.
- If `PickDirectory()` is cancelled (returns empty string), don't overwrite the current path.
- If the user navigates to Settings from the detail view and presses Back, they should return to feed (not detail). The existing `goBack()` function already does this.
- Deep linking: if the user is on Settings and an agent notification arrives, they should still receive it (the notification system runs globally via `EventsOn` in App.svelte, unaffected by view state).

### Reference Files

- `frontend/src/views/AgentDetail.svelte` -- View structure pattern, back button, event dispatching
- `frontend/src/views/Setup.svelte` -- Another view that uses `PickDirectory`
- `frontend/src/App.svelte` -- View routing, handleKeydown, goBack
- `frontend/src/components/MonacoEditor.svelte` -- Button and input styling patterns
- `frontend/src/style.css` -- Design system variables

## Acceptance Criteria

AC-1: Settings view renders with three sections
- Given the user navigates to Settings
- When the Settings view loads
- Then it displays three sections: Theme, VSCodium Extension, and About
- And each section has a heading label

AC-2: Settings view is accessible from TitleBar gear icon
- Given the app is on the feed view
- When the user clicks the gear icon in the TitleBar
- Then `currentView` changes to `'settings'`
- And the Settings view replaces the feed content

AC-3: Back button returns to feed
- Given the user is on the Settings view
- When the user clicks the back arrow button
- Then the app returns to the feed view
- And `currentView` is `'feed'`

AC-4: Escape key returns to feed from Settings
- Given the user is on the Settings view
- When the user presses the Escape key
- Then the app returns to the feed view

AC-5: Theme picker in Settings applies and persists theme
- Given the Settings view is open with the Theme section visible
- When the user clicks a theme card (e.g., "Light")
- Then the app UI switches to that theme immediately
- And `SetTheme` is called to persist the selection

AC-6: VSCodium path can be entered manually and saved
- Given the Settings view is open
- When the user types a path in the VSCodium Extension input and clicks Save
- Then `SetVSCodiumExtPath` is called with the entered path
- And a "Saved" confirmation appears briefly

AC-7: VSCodium path can be set via Browse button
- Given the Settings view is open
- When the user clicks the Browse button
- Then the native OS directory picker opens (via `PickDirectory()`)
- And when a directory is selected, the input updates and the path is persisted

AC-8: VSCodium path loads from config on mount
- Given `~/.conductor/config.json` has `"vscodiumExtPath": "/Applications/VSCodium.app"`
- When the Settings view mounts
- Then the VSCodium Extension input is pre-filled with that path

AC-9: Settings view handles config load failure gracefully
- Given `GetConfig()` will throw an error
- When the Settings view mounts
- Then the view still renders with empty defaults
- And no unhandled error is thrown

## BDD Test Scenarios

### Scenario 1: Settings View Navigation

```gherkin
Feature: Settings View Navigation

  Scenario: Navigate to Settings from gear icon
    Given currentView is "feed"
    When the TitleBar dispatches "open-settings"
    Then currentView becomes "settings"
    And the Settings component is rendered

  Scenario: Back button returns to feed
    Given currentView is "settings"
    When the back button in Settings is clicked
    Then the Settings component dispatches "back"
    And currentView becomes "feed"

  Scenario: Escape returns to feed from Settings
    Given currentView is "settings"
    When the user presses Escape
    Then currentView becomes "feed"

  Scenario: Escape closes spawn modal before navigating
    Given currentView is "feed" and showSpawnModal is true
    When the user presses Escape
    Then showSpawnModal becomes false
    And currentView remains "feed"
```

### Scenario 2: Theme Section

```gherkin
Feature: Settings Theme Picker

  Scenario: Theme cards display all themes
    Given the Settings view is mounted
    When inspecting the Theme section
    Then there are 3 theme cards
    And each shows a label and color swatches

  Scenario: Clicking a theme card applies it
    Given the Settings view shows the Theme section
    And the current theme is "conductor-dark"
    When the user clicks the "conductor-light" card
    Then applyTheme("conductor-light") is called
    And SetTheme("conductor-light") is called
    And the "conductor-light" card now has the active indicator
```

### Scenario 3: VSCodium Extension Path

```gherkin
Feature: VSCodium Extension Path Setting

  Scenario: Path loaded from config
    Given config has vscodiumExtPath "/Applications/VSCodium.app"
    When the Settings view mounts
    Then the text input contains "/Applications/VSCodium.app"

  Scenario: Manual path entry with Save
    Given the Settings view is open
    When the user types "/usr/local/bin/codium" in the input
    And clicks Save
    Then SetVSCodiumExtPath("/usr/local/bin/codium") is called
    And a "Saved" indicator appears for 2 seconds

  Scenario: Browse button opens directory picker
    Given the Settings view is open
    When the user clicks Browse
    And selects a directory in the native picker
    Then the input updates to the selected path
    And SetVSCodiumExtPath is called with the selected path

  Scenario: Browse cancelled does not change path
    Given the Settings input shows "/existing/path"
    When the user clicks Browse and cancels the dialog
    Then the input still shows "/existing/path"
    And SetVSCodiumExtPath is not called

  Scenario: Config load failure shows empty defaults
    Given GetConfig will throw an error
    When the Settings view mounts
    Then the VSCodium input is empty
    And no error is displayed to the user
```

## Tasks / Subtasks

- [ ] Task 1: Create Settings.svelte view component (AC: AC-1, AC-3, AC-9)
  - [ ] Subtask 1a: Create `frontend/src/views/Settings.svelte` with header (back button + title) and three section containers
  - [ ] Subtask 1b: Implement back button dispatching `'back'` event
  - [ ] Subtask 1c: Load config on mount with error handling
  - [ ] Subtask 1d: Style following AgentDetail patterns: scrollable, section headings, spacing

- [ ] Task 2: Implement Theme section (AC: AC-5)
  - [ ] Subtask 2a: Render theme cards from `themes` map with label and swatches
  - [ ] Subtask 2b: Highlight active theme with `var(--accent-green)` border
  - [ ] Subtask 2c: Wire card click to `applyTheme(id)` + `SetTheme(id)`
  - [ ] Subtask 2d: Add preview area showing bg + text + accent color interplay

- [ ] Task 3: Implement VSCodium Extension section (AC: AC-6, AC-7, AC-8)
  - [ ] Subtask 3a: Add text input bound to `vscodiumPath` variable, pre-filled from config
  - [ ] Subtask 3b: Add Browse button calling `PickDirectory()` -- update input + persist on selection
  - [ ] Subtask 3c: Add Save button calling `SetVSCodiumExtPath(vscodiumPath)` with status feedback
  - [ ] Subtask 3d: Handle empty PickDirectory result (user cancelled)

- [ ] Task 4: Implement About section (AC: AC-1)
  - [ ] Subtask 4a: Add static About section with app name ("Conductor") and version
  - [ ] Subtask 4b: Style consistently with other sections

- [ ] Task 5: Wire Settings view into App.svelte routing (AC: AC-2, AC-4)
  - [ ] Subtask 5a: Add `'settings'` to `currentView` type comment (line 12)
  - [ ] Subtask 5b: Handle `on:open-settings` event on `<TitleBar>` component (line 105)
  - [ ] Subtask 5c: Add `{:else if currentView === 'settings'}` template block rendering `<Settings on:back={goBack} />`
  - [ ] Subtask 5d: Update `handleKeydown` to also return from settings on Escape (line 87-93)
  - [ ] Subtask 5e: Import `Settings` from `./views/Settings.svelte`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `wails dev` launches; gear icon opens Settings view
- [ ] Theme picker in Settings works identically to popover
- [ ] VSCodium path persists and loads on re-open
- [ ] Escape and back button both return to feed
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
