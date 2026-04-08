# Story 2: Wire Theme into Components

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** Story 1
**Status:** ready

## Description

Make the Terminal and Monaco Editor components reactively subscribe to the theme store so that switching themes instantly updates their visuals without requiring page reload or component recreation. Currently both components use hardcoded color values; this story replaces those with store-driven values and wires up live-update subscriptions.

## Developer Notes

### Architecture

- **Modified file: `frontend/src/components/Terminal.svelte`** -- Import `currentTheme` store, use `$currentTheme.xterm` for initial theme in `new Terminal()` config (line 76), and add a reactive subscription that sets `term.options.theme = $currentTheme.xterm` on theme change.
- **Modified file: `frontend/src/components/MonacoEditor.svelte`** -- Import `currentThemeId` store and `defineAllThemes` from `monacoTheme.js`. Replace `defineTheme(monacoModule)` call (line 406) with `defineAllThemes(monacoModule)`. Replace hardcoded `'conductor-dark'` string in `getEditorOptions()` (line 75) with the current store value. Add a reactive subscription that calls `monaco.editor.setTheme($currentThemeId)` on theme change.

### Terminal.svelte Changes (lines 69-96)

Current code at line 73-90:
```js
term = new Terminal({
  fontFamily: "'JetBrains Mono', monospace",
  fontSize: 13,
  theme: {
    background: '#07080a',
    // ... hardcoded colors
  },
  // ...
});
```

Replace with:
```js
import { currentTheme } from '../lib/stores/theme.js';

// In onMount:
term = new Terminal({
  fontFamily: "'JetBrains Mono', monospace",
  fontSize: 13,
  theme: $currentTheme.xterm,
  // ...
});

// Reactive update (outside onMount, as a reactive statement):
$: if (term && $currentTheme) {
  term.options.theme = $currentTheme.xterm;
}
```

The xterm.js `Terminal` instance supports live theme updates via `term.options.theme = newTheme` -- no need to destroy and recreate the terminal.

### MonacoEditor.svelte Changes

Current `getEditorOptions()` at line 75:
```js
theme: 'conductor-dark',
```

Replace with dynamic store value. Since `getEditorOptions()` is called during editor creation, it should read the current store value:
```js
import { currentThemeId } from '../lib/stores/theme.js';
import { defineAllThemes } from '../lib/monacoTheme.js';
import { get } from 'svelte/store';

// In getEditorOptions():
theme: get(currentThemeId),

// In onMount (line 405-407):
defineAllThemes(monacoModule);

// Reactive update:
$: if (monacoModule && $currentThemeId) {
  monacoModule.editor.setTheme($currentThemeId);
}
```

`monaco.editor.setTheme(id)` is the official API for live theme switching -- it updates all editor instances globally.

### Technical Considerations

- The `$currentTheme` auto-subscription syntax works in Svelte component `<script>` blocks. For derived values in non-reactive contexts (like `getEditorOptions()`), use `get(currentThemeId)` from `svelte/store`.
- xterm.js `term.options.theme` assignment triggers an internal re-render. No need to call `term.refresh()`.
- Monaco `editor.setTheme()` is global -- it changes the theme for ALL Monaco editor instances. This is the correct behavior since we only ever have one active editor.
- The reactive `$:` block for terminal theme update must guard on `term` being initialized (it is null before `onMount`).
- The reactive `$:` block for Monaco theme update must guard on `monacoModule` being loaded (it is null before the dynamic import resolves).

### Risks & Edge Cases

- If the Terminal component is destroyed and recreated (e.g., navigating away from AgentDetail and back), the new instance picks up the current theme from the store -- no stale state.
- If Monaco is not yet loaded when a theme change fires, the reactive block's guard (`if (monacoModule && ...)`) prevents errors.
- The diff editor variant also uses `getEditorOptions()` (line 342-347), so both source and diff editors benefit from the dynamic theme.

### Reference Files

- `frontend/src/components/Terminal.svelte` -- lines 69-96 (Terminal creation with theme)
- `frontend/src/components/MonacoEditor.svelte` -- lines 73-92 (getEditorOptions), 395-415 (onMount with defineTheme)
- `frontend/src/lib/stores/theme.js` -- (from Story 1)
- `frontend/src/lib/monacoTheme.js` -- (modified in Story 1)

## Acceptance Criteria

AC-1: Terminal uses theme store for initial colors
- Given the theme store is set to `'conductor-midnight'`
- When the Terminal component mounts
- Then the xterm.js Terminal instance is created with `$currentTheme.xterm` colors
- And no hardcoded color values remain in Terminal.svelte

AC-2: Terminal updates live when theme changes
- Given the Terminal component is mounted and displaying content
- When `applyTheme('conductor-light')` is called
- Then the terminal background, foreground, and cursor colors update immediately
- And the terminal content remains visible (no flash, no re-creation)

AC-3: Monaco Editor uses theme store for initial theme
- Given the theme store is set to `'conductor-dark'`
- When the MonacoEditor component mounts and creates an editor
- Then the editor uses theme `'conductor-dark'` (from the store, not hardcoded)
- And `defineAllThemes` (not `defineTheme`) is called during setup

AC-4: Monaco Editor updates live when theme changes
- Given a Monaco editor is open showing a file
- When `applyTheme('conductor-light')` is called
- Then the editor background, syntax colors, and line numbers update immediately
- And the editor content (file text) is preserved

AC-5: Diff editor respects theme
- Given the MonacoEditor is in diff mode (`mode='diff'`)
- When the theme changes
- Then the diff editor also updates to the new theme colors

## BDD Test Scenarios

### Scenario 1: Terminal Theme Integration

```gherkin
Feature: Terminal Theme Reactivity

  Scenario: Terminal initializes with current theme
    Given currentThemeId is "conductor-midnight"
    And the Terminal component is rendered
    When checking the xterm Terminal options
    Then term.options.theme.background matches themes["conductor-midnight"].xterm.background
    And term.options.theme.foreground matches themes["conductor-midnight"].xterm.foreground

  Scenario: Terminal updates on theme change
    Given the Terminal component is mounted with "conductor-dark"
    And text content has been written to the terminal
    When applyTheme("conductor-light") is called
    Then term.options.theme equals themes["conductor-light"].xterm
    And the terminal element remains in the DOM (not recreated)

  Scenario: Terminal survives rapid theme switching
    Given the Terminal component is mounted
    When applyTheme is called 5 times in rapid succession
    Then the final term.options.theme matches the last theme applied
    And no errors are thrown
```

### Scenario 2: Monaco Editor Theme Integration

```gherkin
Feature: Monaco Editor Theme Reactivity

  Scenario: Monaco registers all themes on mount
    Given the MonacoEditor component mounts
    When the monaco module loads
    Then defineAllThemes is called (not defineTheme)
    And all three themes are registered with monaco.editor.defineTheme

  Scenario: Monaco editor uses store theme
    Given currentThemeId is "conductor-midnight"
    When a source editor is created via getEditorOptions()
    Then the options.theme is "conductor-midnight"

  Scenario: Monaco updates live on theme change
    Given a Monaco editor is open with "conductor-dark"
    When applyTheme("conductor-light") is called
    Then monaco.editor.setTheme is called with "conductor-light"

  Scenario: Monaco handles theme change before module loads
    Given monacoModule is null (still loading)
    When applyTheme("conductor-light") is called
    Then no error is thrown
    And when monacoModule eventually loads, it uses "conductor-light"
```

## Tasks / Subtasks

- [ ] Task 1: Wire Terminal.svelte to theme store (AC: AC-1, AC-2)
  - [ ] Subtask 1a: Import `currentTheme` from `../lib/stores/theme.js`
  - [ ] Subtask 1b: Replace hardcoded theme object (lines 76-90) with `$currentTheme.xterm`
  - [ ] Subtask 1c: Add reactive `$:` block to update `term.options.theme` on store change
  - [ ] Subtask 1d: Verify terminal still works in both paneTarget (interactive) and log-polling modes

- [ ] Task 2: Wire MonacoEditor.svelte to theme store (AC: AC-3, AC-4, AC-5)
  - [ ] Subtask 2a: Import `currentThemeId` from `../lib/stores/theme.js` and `defineAllThemes` from `monacoTheme.js`
  - [ ] Subtask 2b: Replace `defineTheme(monacoModule)` call in onMount (line 406) with `defineAllThemes(monacoModule)`
  - [ ] Subtask 2c: Replace hardcoded `'conductor-dark'` in `getEditorOptions()` (line 75) with `get(currentThemeId)`
  - [ ] Subtask 2d: Add reactive `$:` block to call `monacoModule.editor.setTheme($currentThemeId)` on theme change
  - [ ] Subtask 2e: Verify both source and diff editor modes pick up theme changes

- [ ] Task 3: Manual integration testing (AC: AC-1 through AC-5)
  - [ ] Subtask 3a: Verify dark theme looks identical to pre-change state
  - [ ] Subtask 3b: Temporarily call `applyTheme('conductor-light')` in App.svelte onMount to test light theme renders
  - [ ] Subtask 3c: Verify no console errors on theme switch while terminal and editor are active

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `wails dev` launches and dark theme looks identical to before
- [ ] Switching themes updates Terminal + Monaco simultaneously with no errors
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
