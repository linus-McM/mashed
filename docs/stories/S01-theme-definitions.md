# Story 1: Theme Definitions + Store

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Create a centralized theme system that defines all visual themes (CSS variables, Monaco editor colors, xterm.js terminal colors) in a single source of truth. Add a Svelte store to hold the active theme and a function to apply it at runtime. This is the foundation for all subsequent theme-switching work -- nothing else can proceed without it.

## Developer Notes

### Architecture

- **New file: `frontend/src/lib/themes.js`** -- Exports a `themes` map keyed by theme ID (`'conductor-dark'`, `'conductor-light'`, `'conductor-midnight'`). Each theme object has three sub-objects: `css`, `monaco`, `xterm`.
- **New file: `frontend/src/lib/stores/theme.js`** -- Exports a `currentThemeId` writable store (default `'conductor-dark'`), a `currentTheme` derived store, and an `applyTheme(id)` function.
- **Modified file: `frontend/src/lib/monacoTheme.js`** -- Replace `defineTheme(monaco)` with `defineAllThemes(monaco)` that iterates over `themes` and registers each. Keep `EDITOR_FONT` export unchanged.
- **New directory: `frontend/src/lib/stores/`** -- Does not currently exist; must be created.

### Data Shapes

The `themes` map structure:

```js
export const themes = {
  'conductor-dark': {
    label: 'Dark',
    css: {
      '--bg-deepest': '#07080a',
      '--bg-surface': '#0d0f12',
      '--bg-elevated': '#12151a',
      '--bg-active': '#181c23',
      '--border-subtle': '#1e2530',
      '--border-emphasis': '#2a3340',
      '--accent-green': '#00e57a',
      '--accent-green-dim': '#006636',
      '--accent-amber': '#f0a500',
      '--accent-red': '#e84545',
      '--accent-blue': '#3d9eff',
      '--accent-purple': '#9d6fff',
      '--accent-teal': '#00c4b3',
      '--text-primary': '#c8d4e0',
      '--text-dim': '#4a5a6a',
      '--text-muted': '#2e3d4d',
    },
    monaco: { /* full IStandaloneThemeData -- extract verbatim from current monacoTheme.js lines 2-38 */ },
    xterm: { /* full ITheme -- extract verbatim from Terminal.svelte lines 76-90 */ },
  },
  'conductor-light': { /* light variant */ },
  'conductor-midnight': { /* midnight-blue variant */ },
};
```

The store file:

```js
import { writable, derived } from 'svelte/store';
import { themes } from '../themes.js';

export const currentThemeId = writable('conductor-dark');
export const currentTheme = derived(currentThemeId, ($id) => themes[$id] || themes['conductor-dark']);

export function applyTheme(id) {
  const theme = themes[id];
  if (!theme) return;
  for (const [prop, value] of Object.entries(theme.css)) {
    document.documentElement.style.setProperty(prop, value);
  }
  currentThemeId.set(id);
}
```

### Technical Considerations

- The 16 CSS custom properties to extract are defined in `frontend/src/style.css` lines 4-19. The `css` sub-object keys must match these property names exactly (including the `--` prefix).
- The Monaco theme data to extract is in `frontend/src/lib/monacoTheme.js` lines 2-38. The `conductor-dark` Monaco sub-object must be byte-for-byte identical to the current definition so the dark theme is visually unchanged.
- The xterm.js theme data to extract is in `frontend/src/components/Terminal.svelte` lines 76-90. Same rule: conductor-dark must match exactly.
- `defineAllThemes(monaco)` must call `monaco.editor.defineTheme(id, themeData)` for each entry in `themes`.
- `monacoTheme.js` should continue to export `EDITOR_FONT` unchanged.
- The light theme CSS should use light backgrounds (#f5f5f5 range) and dark text. Monaco base should be `'vs'` (not `'vs-dark'`).
- The midnight theme should use deep navy backgrounds (#0a0e1a range). Monaco base stays `'vs-dark'`.

### Risks & Edge Cases

- If a theme ID is misspelled or missing from the map, `currentTheme` must fall back to `conductor-dark` (the derived store handles this).
- `applyTheme()` sets CSS properties on `document.documentElement`, which means they override `:root` declarations in `style.css`. This is intentional -- the `style.css` `:root` block serves as the default/fallback.
- Font-related CSS variables (`--font-ui`, `--font-mono`, `--font-terminal`) and spacing/sizing variables should NOT be included in theme definitions -- they are not theme-dependent.

### Reference Files

- `frontend/src/style.css` -- CSS variable names to extract (lines 4-19)
- `frontend/src/lib/monacoTheme.js` -- Monaco theme data to extract (lines 1-41)
- `frontend/src/components/Terminal.svelte` -- xterm theme data to extract (lines 73-96)

## Acceptance Criteria

AC-1: Theme map contains three complete themes
- Given the app source includes `frontend/src/lib/themes.js`
- When a developer imports `themes` from that module
- Then there are exactly three entries: `conductor-dark`, `conductor-light`, `conductor-midnight`
- And each entry has `label`, `css`, `monaco`, and `xterm` sub-objects
- And the `css` object has all 16 color properties from `style.css`

AC-2: Dark theme matches existing visuals exactly
- Given the `conductor-dark` theme is applied
- When comparing its `css` values to `style.css` lines 4-19
- Then every value matches exactly
- And the `monaco` sub-object matches the current `monacoTheme.js` definition
- And the `xterm` sub-object matches the current `Terminal.svelte` theme

AC-3: Theme store provides reactive access to the active theme
- Given `currentThemeId` is set to `'conductor-dark'`
- When a component subscribes to `currentTheme`
- Then it receives the full theme object for `conductor-dark`
- And when `currentThemeId` changes to `'conductor-light'`, subscribers receive the light theme object

AC-4: applyTheme sets CSS custom properties on document root
- Given the app is running with the dark theme
- When `applyTheme('conductor-light')` is called
- Then all 16 CSS custom properties on `document.documentElement` are updated to light theme values
- And `currentThemeId` store value is `'conductor-light'`

AC-5: defineAllThemes registers all themes with Monaco
- Given Monaco editor module is loaded
- When `defineAllThemes(monaco)` is called
- Then `monaco.editor.defineTheme` is called once for each theme in the `themes` map
- And the `EDITOR_FONT` export is still available

AC-6: Invalid theme ID falls back to conductor-dark
- Given `currentThemeId` is set to `'nonexistent-theme'`
- When `currentTheme` is derived
- Then the result is the `conductor-dark` theme object

## BDD Test Scenarios

### Scenario 1: Theme Map Integrity

```gherkin
Feature: Theme Definitions

  Scenario: All themes have required structure
    Given the themes map is imported from themes.js
    When iterating over all theme entries
    Then each theme has a "label" string
    And each theme has a "css" object with 16 properties
    And each theme has a "monaco" object with "base", "inherit", "rules", and "colors" keys
    And each theme has an "xterm" object with "background", "foreground", and "cursor" keys

  Scenario: Dark theme CSS matches style.css defaults
    Given the conductor-dark theme is loaded
    When checking its css property "--bg-deepest"
    Then the value is "#07080a"
    And css property "--accent-green" is "#00e57a"
    And css property "--text-primary" is "#c8d4e0"

  Scenario: Light theme uses light base colors
    Given the conductor-light theme is loaded
    When checking its monaco.base property
    Then the value is "vs"
    And its css "--bg-deepest" is a light color (luminance > 0.7)
    And its css "--text-primary" is a dark color (luminance < 0.3)
```

### Scenario 2: Theme Store Reactivity

```gherkin
Feature: Theme Store

  Scenario: Default theme is conductor-dark
    Given the theme store is freshly imported
    When reading currentThemeId
    Then the value is "conductor-dark"

  Scenario: Changing theme ID updates derived store
    Given currentThemeId is "conductor-dark"
    When applyTheme("conductor-midnight") is called
    Then currentThemeId value is "conductor-midnight"
    And currentTheme value has label "Midnight"

  Scenario: Invalid theme ID falls back gracefully
    Given currentThemeId is set to "does-not-exist"
    When reading currentTheme
    Then the result is the conductor-dark theme object
```

### Scenario 3: defineAllThemes Registration

```gherkin
Feature: Monaco Theme Registration

  Scenario: All themes registered with Monaco
    Given a mock Monaco module with a spy on editor.defineTheme
    When defineAllThemes(mockMonaco) is called
    Then editor.defineTheme was called 3 times
    And the first call used id "conductor-dark"
    And EDITOR_FONT is still exported as "'Geist Mono', 'JetBrains Mono', monospace"
```

## Tasks / Subtasks

- [ ] Task 1: Create themes.js with three theme definitions (AC: AC-1, AC-2)
  - [ ] Subtask 1a: Extract the 16 CSS color variables from `style.css` lines 4-19 into the `conductor-dark` css sub-object
  - [ ] Subtask 1b: Extract the Monaco theme data from `monacoTheme.js` lines 2-38 into the `conductor-dark` monaco sub-object
  - [ ] Subtask 1c: Extract the xterm theme data from `Terminal.svelte` lines 76-90 into the `conductor-dark` xterm sub-object
  - [ ] Subtask 1d: Create `conductor-light` theme with light backgrounds, dark text, `monaco.base: 'vs'`
  - [ ] Subtask 1e: Create `conductor-midnight` theme with deep navy backgrounds, `monaco.base: 'vs-dark'`

- [ ] Task 2: Create theme store (AC: AC-3, AC-4, AC-6)
  - [ ] Subtask 2a: Create `frontend/src/lib/stores/` directory
  - [ ] Subtask 2b: Create `frontend/src/lib/stores/theme.js` with `currentThemeId` writable, `currentTheme` derived, and `applyTheme()` function
  - [ ] Subtask 2c: Implement fallback logic in `currentTheme` derived store for unknown IDs

- [ ] Task 3: Refactor monacoTheme.js to use themes.js (AC: AC-5)
  - [ ] Subtask 3a: Replace `defineTheme(monaco)` export with `defineAllThemes(monaco)` that loops over `themes`
  - [ ] Subtask 3b: Preserve `EDITOR_FONT` export unchanged
  - [ ] Subtask 3c: Remove hardcoded theme data from monacoTheme.js (now sourced from themes.js)

- [ ] Task 4: Write unit tests (AC: AC-1 through AC-6)
  - [ ] Subtask 4a: Test theme map structure and completeness
  - [ ] Subtask 4b: Test store reactivity and fallback behavior
  - [ ] Subtask 4c: Test defineAllThemes with mock Monaco module

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes (no Go changes in this story, but verify no regressions)
- [ ] `wails dev` launches and dark theme looks identical to before
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
