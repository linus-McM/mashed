# Story 4: Settings UI -- Theme Scanning, Activation & Startup Restore

**Priority:** P1-high
**Domain:** fullstack
**Estimated Complexity:** L
**Depends On:** Story 1, Story 2, Story 3
**Status:** ready

## Description

Wire the theme import workflow end-to-end: Settings.svelte gains a new "Imported Themes" section that scans the VSCodium extensions directory, displays discovered themes, and lets users activate them with a single click. Theme activation reads the file from the Go backend, converts it with the frontend converter, registers it in the store, and persists the selection. Critically, this story also implements startup theme restoration in App.svelte so imported themes survive app restarts -- the single most user-visible issue identified in the review (C-5).

## Developer Notes

### Architecture

**Modified files:**
1. `frontend/src/views/Settings.svelte` -- add imported themes section, scanning, activation
2. `frontend/src/App.svelte` -- add imported theme restoration on startup
3. `frontend/src/lib/themeInit.js` (NEW) -- shared `activateImportedTheme` function used by both Settings and App

**New file: `frontend/src/lib/themeInit.js`**

This shared module contains the theme activation logic that both Settings.svelte and App.svelte need:

```js
import { ReadThemeFile, SetImportedTheme, SetTheme } from '../../wailsjs/go/main/App.js';
import { convertVSCodeTheme, validateConvertedTheme } from './themeConverter.js';
import { registerImportedTheme, applyTheme, DEFAULT_THEME } from './stores/theme.js';

// Cache of already-converted themes: { themePath: { id, theme } }
const convertedCache = {};

/**
 * Activate an imported theme by path. Reads, converts, registers, applies, persists.
 * @param {string} themePath - absolute path to the theme JSON file
 * @param {string} extensionId - extension directory name for unique ID generation
 * @returns {Promise<string>} the theme ID
 */
export async function activateImportedTheme(themePath, extensionId) { ... }

/**
 * Restore the imported theme from config on app startup.
 * @param {object} cfg - the config object from GetConfig()
 */
export async function restoreImportedThemeFromConfig(cfg) { ... }
```

**Theme ID generation (M-5 fix):** Include the extension ID to prevent collisions:
```js
const filename = themePath.split('/').pop().replace('.json', '');
const themeId = 'imported-' + extensionId + '-' + filename;
```

**Data flow for activation:**
1. User clicks theme button in Settings
2. `activateImportedTheme(entry.themePath, entry.extensionId)` called
3. `ReadThemeFile(themePath)` --> Go backend reads, strips JSONC, resolves includes
4. `JSON.parse(raw)` --> parse the clean JSON
5. `convertVSCodeTheme(vsTheme, themeId)` --> convert to Conductor format
6. `validateConvertedTheme(converted)` --> ensure minimum CSS vars present
7. `registerImportedTheme(themeId, converted)` --> add to `allThemes` store
8. `applyTheme(themeId)` --> apply CSS vars, update `currentThemeId`
9. `SetTheme(themeId)` + `SetImportedTheme(themePath)` --> persist to config
10. Monaco picks up the change reactively (Story 5)

**Data flow for startup restore (C-5 fix):**
1. App.svelte `onMount` calls `GetConfig()`
2. If `cfg.importedTheme` is set, call `restoreImportedThemeFromConfig(cfg)`
3. This calls `ReadThemeFile` --> convert --> register --> apply
4. If restoration fails, fall back to `cfg.theme` (built-in) or `DEFAULT_THEME`

### Technical Considerations

**Activation guard (M-1 fix):** Prevent double-activation when user clicks rapidly:
```js
let activatingPath = null;

async function activateImportedTheme(themePath, extensionId) {
    if (activatingPath === themePath) return;
    activatingPath = themePath;
    try {
        // ... activation logic
    } finally {
        activatingPath = null;
    }
}
```

**Theme ID collision prevention (M-5 fix):** Two extensions can both have a `dark.json`. Include `extensionId` in the ID:
```js
const themeId = 'imported-' + extensionId + '-' + filename;
// e.g., "imported-github.github-vscode-theme-6.3.5-dark"
```

The `ListVSCodiumThemes` response includes `extensionId` for each entry, so Settings.svelte passes it through.

**Startup restore timing (C-5 fix):** App.svelte must attempt imported theme restoration BEFORE the loading screen clears. The sequence:
1. `GetConfig()` --> get saved config
2. If `cfg.importedTheme`: call `restoreImportedThemeFromConfig(cfg)` (async)
3. If restoration fails OR no imported theme: `applyTheme(cfg.theme || DEFAULT_THEME)`
4. Continue with normal app initialization

**Error recovery:** If theme loading fails at any point, revert to the last known good theme or DEFAULT_THEME. Never leave the app in an unstyled state.

**Settings.svelte new section:** Add an "Imported Themes" section below the existing "VSCodium Extension" section. It shows:
- A loading spinner while scanning
- The list of discovered themes (from `ListVSCodiumThemes`)
- Each theme as a button with a dark/light badge and label
- Active state highlighting for the currently selected imported theme
- Error message if scanning fails

**Selecting a built-in theme should clear the imported theme:**
```js
async function selectBuiltInTheme(id) {
    applyTheme(id);
    await SetTheme(id);
    await SetImportedTheme('');  // Clear imported theme from config
}
```

### Risks & Edge Cases

- **Startup latency:** `ReadThemeFile` + conversion adds ~50-100ms to startup. This is acceptable since it happens during the loading screen.
- **Missing extensions directory on startup:** If the user configured a path but later deleted it, `ReadThemeFile` will fail. The fallback to built-in theme handles this.
- **Theme file moved/deleted:** If the persisted `importedTheme` path no longer exists, restoration fails silently and falls back to the built-in theme.
- **Re-scanning after path change:** When the user changes the VSCodium extensions path via the Browse button, the imported themes list must refresh.
- **Large number of themes:** Some extension packs define 10+ themes. The list should be scrollable.

### Reference Files

- `/Users/linus/Development/mashed/frontend/src/views/Settings.svelte` -- current Settings view (330 lines)
- `/Users/linus/Development/mashed/frontend/src/App.svelte` -- current App.svelte (174 lines), startup logic at lines 19-36
- `/Users/linus/Development/mashed/frontend/src/lib/stores/theme.js` -- store after Story 3 refactor
- `/Users/linus/Development/mashed/frontend/src/lib/themeConverter.js` -- converter from Story 2
- `/Users/linus/Development/mashed/docs/vscodium-theme-loading-plan.md` lines 925-1069 -- Phase 4 Settings UI code

## Acceptance Criteria

AC-1: Theme scanning in Settings
- Given a configured VSCodium extensions path with installed themes
- When the Settings view is opened
- Then the imported themes section displays all discovered themes sorted alphabetically
- And each theme shows a dark/light badge and its label

AC-2: Theme activation
- Given a list of discovered themes in Settings
- When the user clicks on a theme entry
- Then the theme is loaded from the Go backend, converted, registered, and applied
- And CSS custom properties update immediately
- And the theme ID includes the extension ID to prevent collisions

AC-3: Activation guard prevents double-loading
- Given the user rapidly clicks the same theme button twice
- When the second click fires while the first is still loading
- Then only one `ReadThemeFile` call is made
- And the theme is applied exactly once

AC-4: Built-in theme selection clears imported theme
- Given an imported theme is currently active
- When the user selects a built-in theme
- Then the built-in theme is applied
- And `SetImportedTheme('')` is called to clear the persisted imported theme

AC-5: Startup theme restoration (C-5 fix)
- Given the config has `importedTheme` set to a valid theme path
- When the app starts
- Then the imported theme is loaded, converted, registered, and applied in App.svelte `onMount`
- And the user sees the correct theme before any content renders

AC-6: Startup restore fallback
- Given the config has `importedTheme` set to a path that no longer exists
- When the app starts
- Then the restoration fails silently
- And the app falls back to `cfg.theme` (built-in) or `DEFAULT_THEME`
- And the app does not render unstyled

AC-7: Re-scan on path change
- Given themes are displayed from the current extensions path
- When the user changes the VSCodium extensions path via Browse
- Then the theme list refreshes to show themes from the new path

## BDD Test Scenarios

### Scenario 1: End-to-end theme activation

```gherkin
Feature: Import and activate VSCodium theme

  Scenario: Activate a Dracula theme from extensions
    Given VSCodium extensions path contains "dracula-theme.theme-dracula-2.24.3"
    And the extension has theme label "Dracula" at path "./theme/dracula.json"
    When Settings view mounts and scans themes
    Then the imported themes section shows "Dracula" with a dark badge
    When the user clicks the "Dracula" theme button
    Then ReadThemeFile is called with the full path to dracula.json
    And the theme is converted and registered as "imported-dracula-theme.theme-dracula-2.24.3-dracula"
    And CSS custom properties update to match Dracula's colors
    And SetTheme and SetImportedTheme are called to persist the selection
```

### Scenario 2: Startup restore

```gherkin
Feature: Imported theme startup restore

  Scenario: Restore imported theme on app launch
    Given config has importedTheme "/path/to/dracula.json"
    And the file exists and is a valid theme
    When App.svelte mounts
    Then restoreImportedThemeFromConfig is called
    And the imported theme is applied before the loading screen clears

  Scenario: Fallback when imported theme file is missing
    Given config has importedTheme "/path/to/deleted-theme.json"
    And the file no longer exists
    When App.svelte mounts
    Then restoreImportedThemeFromConfig fails
    And applyTheme is called with the built-in theme from cfg.theme
    And no error is shown to the user

  Scenario: No imported theme in config
    Given config has no importedTheme field
    When App.svelte mounts
    Then only the built-in theme from cfg.theme is applied
    And restoreImportedThemeFromConfig is not called
```

### Scenario 3: Activation edge cases

```gherkin
Feature: Theme activation edge cases

  Scenario: Rapid double-click prevention
    Given theme scanning has completed
    When the user clicks "Dracula" theme twice within 100ms
    Then only one ReadThemeFile call is made
    And the theme is applied once

  Scenario: Switch from imported to built-in
    Given "imported-dracula" is the active theme
    When the user clicks the "Dark" built-in theme
    Then applyTheme("conductor-dark") is called
    And SetImportedTheme("") clears the imported theme path
    And the next app restart will use "conductor-dark"

  Scenario: Theme ID collision prevention
    Given two extensions both contain "dark.json":
      | extensionId                         | label       |
      | github.github-vscode-theme-6.3.5    | GitHub Dark |
      | material-theme.material-theme-1.0.0 | Material Dark |
    When both themes are activated
    Then they have different IDs including their extensionId
    And both can be selected independently
```

### Scenario 4: Settings scanning

```gherkin
Feature: Settings theme scanning

  Scenario: Display scan results
    Given VSCodium path is configured with 5 extensions containing 8 total themes
    When Settings view mounts
    Then the imported themes section shows 8 theme buttons
    And they are sorted alphabetically by label

  Scenario: Handle scan error
    Given VSCodium path points to a non-existent directory
    When Settings view mounts and scans
    Then an error message is displayed in the imported themes section
    And the rest of Settings remains functional

  Scenario: Re-scan after path change
    Given old path has 3 themes displayed
    When user clicks Browse and selects a new directory with 5 themes
    Then the imported themes section updates to show 5 themes
```

## Tasks / Subtasks

- [ ] Task 1: Create shared themeInit.js module (AC: AC-2, AC-5, AC-6)
  - [ ] Create `/Users/linus/Development/mashed/frontend/src/lib/themeInit.js`
  - [ ] Implement `activateImportedTheme(themePath, extensionId)` with caching
  - [ ] Implement `restoreImportedThemeFromConfig(cfg)` with fallback logic
  - [ ] Include activation guard variable to prevent double-loading
  - [ ] Use extensionId in theme ID generation (M-5 fix)
  - [ ] Add error recovery -- revert to DEFAULT_THEME on failure

- [ ] Task 2: Update App.svelte for startup restore (AC: AC-5, AC-6)
  - [ ] Import `restoreImportedThemeFromConfig` from `themeInit.js`
  - [ ] In `onMount`, after `GetConfig()`, check for `cfg.importedTheme`
  - [ ] If present, await `restoreImportedThemeFromConfig(cfg)`
  - [ ] If absent or failed, fall back to `applyTheme(cfg.theme)` as before
  - [ ] Ensure theme is applied before `currentView` changes from 'loading'

- [ ] Task 3: Add imported themes section to Settings.svelte (AC: AC-1, AC-4, AC-7)
  - [ ] Add state: `vscodiumThemes`, `loadingThemes`, `themeLoadError`
  - [ ] Add `scanThemes()` function calling `ListVSCodiumThemes`
  - [ ] Call `scanThemes()` in `onMount` when path is configured
  - [ ] Add imported themes UI section with theme buttons, badges, active state
  - [ ] Add `selectBuiltInTheme(id)` that clears imported theme
  - [ ] Update `browseVSCodium()` to re-scan after path change
  - [ ] Add CSS styles for `.theme-list`, `.imported-theme-btn`, `.theme-badge`

- [ ] Task 4: Wire activation in Settings.svelte (AC: AC-2, AC-3)
  - [ ] Import `activateImportedTheme` from `themeInit.js`
  - [ ] Wire click handler on imported theme buttons to call `activateImportedTheme`
  - [ ] Pass both `entry.themePath` and `entry.extensionId` to the function
  - [ ] Show loading state while activation is in progress

- [ ] Task 5: Write tests (AC: all)
  - [ ] Test `activateImportedTheme` with mock ReadThemeFile
  - [ ] Test `restoreImportedThemeFromConfig` happy path and failure fallback
  - [ ] Test activation guard prevents concurrent calls
  - [ ] Test theme ID includes extensionId
  - [ ] Test built-in theme selection clears imported theme

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `themeInit.js`
- [ ] Imported theme survives app restart (verified manually)
- [ ] Startup restore does not add visible flicker
- [ ] `go build ./...` passes (backend Story 1 must be merged first)
- [ ] Code review: no CRITICAL/HIGH issues
