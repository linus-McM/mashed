# Story 3: Theme Store Refactor & Consumer Updates

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Refactor the theme store to support dynamically imported themes alongside built-in themes, and update ALL consumer components to use the new store shape. This is the critical integration layer -- the store changes `themeIds` from a plain array to a Svelte derived store and introduces `allThemes` as the single source of truth. Every component that reads themes must be updated simultaneously to prevent the silent TitleBar/Settings breakage identified in the adversarial review (C-2).

## Developer Notes

### Architecture

**Modified files:**
1. `frontend/src/lib/stores/theme.js` -- the store itself
2. `frontend/src/lib/monacoTheme.js` -- add `defineImportedTheme` helper
3. `frontend/src/components/TitleBar.svelte` -- update theme picker to use store subscriptions
4. `frontend/src/views/Settings.svelte` -- update theme grid to use store subscriptions
5. `frontend/src/App.svelte` -- no change needed yet (uses `applyTheme` which is updated in-place)

**Store shape changes:**

Current exports from `stores/theme.js`:
```js
export { DEFAULT_THEME, themes };          // static re-exports from themes.js
export const themeIds = Object.keys(themes); // plain array
export const currentThemeId = writable(DEFAULT_THEME);
export const currentTheme = derived(currentThemeId, ($id) => themes[$id] || themes[DEFAULT_THEME]);
export function applyTheme(id) { ... }
```

New exports:
```js
export { DEFAULT_THEME };
export const allThemes = writable({ ...builtInThemes });  // NEW: reactive map of all themes
export const themeIds = derived(allThemes, ($all) => Object.keys($all)); // CHANGED: now a store
export const builtInThemeIds = Object.keys(builtInThemes);              // NEW: static array
export const currentThemeId = writable(DEFAULT_THEME);                   // unchanged
export const currentTheme = derived(                                     // CHANGED: reads from allThemes
    [currentThemeId, allThemes],
    ([$id, $all]) => $all[$id] || $all[DEFAULT_THEME]
);
export const themes = builtInThemes;       // backward compat (static, DO NOT use for imported)
export function applyTheme(id) { ... }     // CHANGED: reads from allThemes via get()
export function registerImportedTheme(id, theme) { ... }   // NEW
export function unregisterImportedTheme(id) { ... }        // NEW
```

**Critical consumer update (C-2 fix):**

TitleBar.svelte currently uses:
```svelte
import { themes, themeIds, currentThemeId, applyTheme } from '../lib/stores/theme.js';
{#each themeIds as id}
  {@const theme = themes[id]}
```

After the refactor, `themeIds` is a store (not a plain array). Using `themeIds` without `$` will iterate over the store object's properties, producing zero iterations or garbage. `themes` is now only the built-in themes map, so `themes[id]` will return `undefined` for imported themes.

TitleBar.svelte MUST change to:
```svelte
import { allThemes, themeIds, currentThemeId, applyTheme } from '../lib/stores/theme.js';
{#each $themeIds as id}
  {@const theme = $allThemes[id]}
```

Settings.svelte has the same pattern and MUST also change.

### Technical Considerations

**MonacoEditor.svelte compatibility:** MonacoEditor imports `currentThemeId` and uses it reactively. The `currentThemeId` store type does not change (still a `writable`), so MonacoEditor.svelte does NOT need updating in this story. However, `monacoTheme.js` should get a new `defineImportedTheme` helper for Story 4 to use.

**Terminal.svelte compatibility:** Terminal.svelte uses `$currentTheme.xterm`. The `currentTheme` derived store now reads from `[currentThemeId, allThemes]` instead of just `currentThemeId`. This is a transparent change -- Terminal.svelte does not need modification as long as `currentTheme` still resolves to an object with `.xterm`.

**applyTheme update:** The function must use `get(allThemes)` instead of the static `themes` object. Import `get` from `svelte/store`.

```js
import { writable, derived, get } from 'svelte/store';

export function applyTheme(id) {
    const all = get(allThemes);
    const theme = all[id];
    if (!theme) return;
    for (const [prop, value] of Object.entries(theme.css)) {
        document.documentElement.style.setProperty(prop, value);
    }
    currentThemeId.set(id);
}
```

**Backward compatibility:** The `themes` export is kept as a static reference to `builtInThemes` for any code that directly accesses it. But all iteration/lookup should use `$allThemes`.

### Risks & Edge Cases

- **Silent regression risk:** If ANY consumer still uses `themeIds` without `$`, it will break silently (renders nothing). Grep the entire frontend for all imports of `themeIds` and `themes` from the store.
- **Reactive dependency:** TitleBar's `{#each $themeIds as id}` creates a reactive subscription. When a new imported theme is registered via `registerImportedTheme`, the TitleBar popover will automatically re-render to include it. Verify this works.
- **Store initialization timing:** `allThemes` is initialized with `{ ...builtInThemes }` on module load. This happens before any component mounts, so built-in themes are always available.
- **unregisterImportedTheme edge case:** If the currently active theme is removed, it must fall back to `DEFAULT_THEME`. The function should check `get(currentThemeId)` and call `applyTheme(DEFAULT_THEME)` if needed.

### Reference Files

- `/Users/linus/Development/mashed/frontend/src/lib/stores/theme.js` -- the file being modified (17 lines currently)
- `/Users/linus/Development/mashed/frontend/src/lib/themes.js` -- source of built-in themes, exported as `themes` and `themeIds`
- `/Users/linus/Development/mashed/frontend/src/lib/monacoTheme.js` -- add `defineImportedTheme` helper
- `/Users/linus/Development/mashed/frontend/src/components/TitleBar.svelte` -- lines 5, 50-51 must change
- `/Users/linus/Development/mashed/frontend/src/views/Settings.svelte` -- lines 4, 67-68 must change
- `/Users/linus/Development/mashed/frontend/src/components/Terminal.svelte` -- line 77 and 182, uses `$currentTheme.xterm` (should work without changes, but verify)
- `/Users/linus/Development/mashed/frontend/src/components/MonacoEditor.svelte` -- lines 6, 430-432, uses `currentThemeId` (no change needed)
- `/Users/linus/Development/mashed/frontend/src/App.svelte` -- line 12, imports `applyTheme` (no change needed)

## Acceptance Criteria

AC-1: Store exports new shape
- Given the updated `stores/theme.js`
- When a component imports `allThemes`, `themeIds`, `builtInThemeIds`, `currentThemeId`, `currentTheme`
- Then `allThemes` is a writable store containing all built-in themes
- And `themeIds` is a derived store (not a plain array) returning `Object.keys($allThemes)`
- And `builtInThemeIds` is a static array of built-in theme IDs

AC-2: registerImportedTheme works
- Given a converted theme object `{ label: 'Dracula', css: {...}, monaco: {...}, xterm: {...} }`
- When `registerImportedTheme('imported-dracula', theme)` is called
- Then `$allThemes['imported-dracula']` returns the registered theme
- And `$themeIds` includes `'imported-dracula'`

AC-3: unregisterImportedTheme with fallback
- Given `'imported-dracula'` is the current theme and is registered
- When `unregisterImportedTheme('imported-dracula')` is called
- Then `$allThemes` no longer contains `'imported-dracula'`
- And `$currentThemeId` falls back to `DEFAULT_THEME`
- And built-in themes cannot be unregistered

AC-4: TitleBar renders all themes
- Given built-in themes and one registered imported theme
- When the TitleBar theme popover is opened
- Then all themes (built-in + imported) appear in the picker
- And each theme card shows the correct label and color swatches
- And selecting an imported theme applies it correctly

AC-5: Settings renders all themes
- Given built-in themes exist
- When the Settings view is opened
- Then the theme grid renders all built-in themes with correct preview cards
- And no JavaScript errors appear in the console

AC-6: applyTheme works with imported themes
- Given an imported theme is registered in `allThemes`
- When `applyTheme('imported-dracula')` is called
- Then CSS custom properties are updated on `document.documentElement`
- And `$currentThemeId` is set to `'imported-dracula'`

AC-7: Terminal reactivity preserved
- Given Terminal.svelte uses `$currentTheme.xterm`
- When the theme is changed to an imported theme
- Then `$currentTheme` updates to the imported theme
- And Terminal.svelte receives the new xterm colors

## BDD Test Scenarios

### Scenario 1: Store operations

```gherkin
Feature: Theme store with imported themes

  Scenario: Register and apply an imported theme
    Given the store is initialized with 3 built-in themes
    When registerImportedTheme is called with id "imported-dracula" and a valid theme object
    Then allThemes contains 4 entries
    And themeIds derived store contains "imported-dracula"
    When applyTheme("imported-dracula") is called
    Then currentThemeId is "imported-dracula"
    And currentTheme resolves to the Dracula theme object

  Scenario: Unregister the active imported theme
    Given "imported-dracula" is registered and is the current theme
    When unregisterImportedTheme("imported-dracula") is called
    Then allThemes contains only 3 built-in themes
    And currentThemeId reverts to DEFAULT_THEME

  Scenario: Cannot unregister built-in themes
    Given "conductor-dark" is a built-in theme
    When unregisterImportedTheme("conductor-dark") is called
    Then allThemes still contains "conductor-dark"
```

### Scenario 2: TitleBar consumer update

```gherkin
Feature: TitleBar theme picker after store refactor

  Scenario: TitleBar renders all themes including imported
    Given 3 built-in themes and 1 imported theme are registered
    When the TitleBar theme popover is rendered
    Then 4 theme cards are displayed
    And each card shows label and color swatches from $allThemes[id]

  Scenario: TitleBar does not break with store-type themeIds
    Given themeIds is a derived store (not a plain array)
    When TitleBar renders "{#each $themeIds as id}"
    Then it correctly iterates over theme IDs
    And does NOT iterate over store object properties
```

### Scenario 3: Settings consumer update

```gherkin
Feature: Settings theme grid after store refactor

  Scenario: Settings renders built-in themes correctly
    Given 3 built-in themes exist
    When Settings view is mounted
    Then the theme grid shows 3 theme option buttons
    And each button shows the correct theme preview with CSS colors from the theme object

  Scenario: No undefined theme references
    Given themeIds contains ["conductor-dark", "conductor-light", "conductor-midnight"]
    When Settings iterates with "{#each $themeIds as id}"
    And accesses "$allThemes[id]"
    Then every access returns a valid theme object (never undefined)
```

### Scenario 4: Backward compatibility

```gherkin
Feature: Backward compatibility

  Scenario: Static themes export still works
    Given the store exports "themes" as builtInThemes
    When code accesses themes['conductor-dark']
    Then it returns the built-in dark theme object

  Scenario: applyTheme with unknown ID is a no-op
    Given no theme with id "nonexistent" exists
    When applyTheme("nonexistent") is called
    Then currentThemeId does not change
    And no error is thrown
```

## Tasks / Subtasks

- [ ] Task 1: Refactor stores/theme.js (AC: AC-1, AC-2, AC-3, AC-6)
  - [ ] Import `get` from `svelte/store`
  - [ ] Add `allThemes` writable store initialized with `{ ...builtInThemes }`
  - [ ] Change `themeIds` from plain array to `derived(allThemes, ($all) => Object.keys($all))`
  - [ ] Add `builtInThemeIds` as a static array
  - [ ] Update `currentTheme` to derive from `[currentThemeId, allThemes]`
  - [ ] Update `applyTheme` to use `get(allThemes)` for theme lookup
  - [ ] Add `registerImportedTheme(id, theme)` function
  - [ ] Add `unregisterImportedTheme(id)` function with fallback logic

- [ ] Task 2: Update TitleBar.svelte (AC: AC-4)
  - [ ] Change import to `{ allThemes, themeIds, currentThemeId, applyTheme }`
  - [ ] Change `{#each themeIds as id}` to `{#each $themeIds as id}`
  - [ ] Change `{@const theme = themes[id]}` to `{@const theme = $allThemes[id]}`
  - [ ] Verify all `theme.css['--xxx']` and `theme.label` accesses still work

- [ ] Task 3: Update Settings.svelte (AC: AC-5)
  - [ ] Change import to `{ allThemes, themeIds, currentThemeId, applyTheme }`
  - [ ] Change `{#each themeIds as id}` to `{#each $themeIds as id}`
  - [ ] Change `{@const theme = themes[id]}` to `{@const theme = $allThemes[id]}`
  - [ ] Verify theme preview card renders correctly

- [ ] Task 4: Update monacoTheme.js (AC: none directly, enables Story 4/5)
  - [ ] Add `defineImportedTheme(monaco, id, monacoThemeData)` function
  - [ ] Keep existing `defineAllThemes` unchanged (still uses static builtInThemes)
  - [ ] Export `defineImportedTheme`

- [ ] Task 5: Verify other consumers are unaffected (AC: AC-7)
  - [ ] Grep entire frontend for imports from `stores/theme.js` and `themes.js`
  - [ ] Verify MonacoEditor.svelte still works (uses `currentThemeId`, not `themeIds`)
  - [ ] Verify Terminal.svelte still works (`$currentTheme.xterm` resolves correctly)
  - [ ] Verify App.svelte still works (`applyTheme` signature unchanged)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `stores/theme.js`
- [ ] No console errors in TitleBar or Settings views
- [ ] `grep -r "each themeIds" frontend/src/` returns zero results (all use `$themeIds`)
- [ ] `grep -r "themes\[id\]" frontend/src/` returns zero results (all use `$allThemes[id]`)
- [ ] Code review: no CRITICAL/HIGH issues
