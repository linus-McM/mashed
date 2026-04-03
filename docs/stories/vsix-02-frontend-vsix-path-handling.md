# Story vsix-02: Update frontend path handling for VSIX-encoded theme paths

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** vsix-01
**Status:** ready

## Description

Update the frontend JavaScript to correctly parse VSIX-encoded theme paths (`/path/to/file.vsix::vsix::extension/themes/dark.json`). The `extractExtensionId` and `makeThemeId` functions in `themeInit.js` currently assume filesystem-style paths (splitting on `/`). They must detect the `::vsix::` separator and extract the extension ID from the vsix filename and the theme ID from the internal path. Additionally, `makeThemeId` must be exported so `Settings.svelte` can import it instead of duplicating its logic inline.

## Developer Notes

### Architecture

Two files are modified:

1. **`/Users/linus/Development/mashed/frontend/src/lib/themeInit.js`**
2. **`/Users/linus/Development/mashed/frontend/src/views/Settings.svelte`**

**Changes to `themeInit.js`:**

`extractExtensionId(themePath)` -- add vsix detection at the top:
```javascript
function extractExtensionId(themePath) {
  if (themePath.includes('::vsix::')) {
    // Path format: "/path/to/file.vsix::vsix::extension/themes/dark.json"
    const vsixPart = themePath.split('::vsix::')[0];
    const filename = vsixPart.split('/').pop();        // "file.vsix"
    return filename.replace(/\.vsix$/i, '');            // "file"
  }
  // Existing logic for filesystem paths (backward compat)
  const parts = themePath.split('/');
  if (parts.length >= 3) {
    return parts[parts.length - 3];
  }
  return parts[parts.length - 2] || 'unknown';
}
```

`makeThemeId(themePath, extensionId)` -- add vsix detection at the top:
```javascript
function makeThemeId(themePath, extensionId) {
  if (themePath.includes('::vsix::')) {
    // Path format: "/path/to/file.vsix::vsix::extension/themes/dark.json"
    const internalPath = themePath.split('::vsix::')[1]; // "extension/themes/dark.json"
    const filename = internalPath.split('/').pop().replace('.json', ''); // "dark"
    return 'imported-' + extensionId + '-' + filename;
  }
  // Existing logic for filesystem paths (backward compat)
  const filename = themePath.split('/').pop().replace('.json', '');
  return 'imported-' + extensionId + '-' + filename;
}
```

**Change visibility:** Both functions are currently declared with `function` (private). Change `makeThemeId` to be exported:
```javascript
export function makeThemeId(themePath, extensionId) {
```

Note: `extractExtensionId` can remain private -- it is only used internally by `restoreImportedThemeFromConfig`.

**Changes to `Settings.svelte`:**

Line 6 -- add `makeThemeId` to the import from `themeInit.js`:
```javascript
import { activateImportedTheme, convertedCache, makeThemeId } from '../lib/themeInit.js';
```

Line 171 -- replace inline ID construction:
```svelte
{@const themeId = makeThemeId(entry.themePath, entry.extensionId)}
```

Currently it reads:
```svelte
{@const themeId = 'imported-' + entry.extensionId + '-' + entry.themePath.split('/').pop().replace('.json', '')}
```

### Technical Considerations

- **No new dependencies** -- pure string manipulation changes.
- **Backward compatibility is critical** -- both functions must continue to work with old-style filesystem paths (no `::vsix::` separator). The `if (themePath.includes('::vsix::'))` guard ensures old paths fall through to existing logic.
- **The `convertedCache` uses `themePath` as the key** -- vsix-encoded paths will work as cache keys without any changes since they are stable strings.
- **`activateImportedTheme` and `restoreImportedThemeFromConfig`** call `extractExtensionId` and `makeThemeId` internally -- the changes propagate automatically, no callers need updating beyond Settings.svelte.

### Risks & Edge Cases

- **Theme ID collision:** Two different vsix files could theoretically have theme files with the same base filename (e.g., `dark.json`). The `extensionId` prefix prevents collision because each vsix has a unique filename.
- **Path with `::vsix::` in a non-vsix context:** Extremely unlikely since `::vsix::` is not a valid filesystem path component. No mitigation needed.
- **Settings.svelte `active` class comparison:** The `$currentThemeId === themeId` comparison in Settings.svelte depends on `makeThemeId` returning the exact same value that `activateImportedTheme` produces. Since both now call the same exported function, this is guaranteed.

### Reference Files

- `/Users/linus/Development/mashed/frontend/src/lib/themeInit.js` -- lines 24-44 (extractExtensionId and makeThemeId)
- `/Users/linus/Development/mashed/frontend/src/views/Settings.svelte` -- lines 1-6 (imports) and line 171 (inline ID construction)

## Acceptance Criteria

AC-1: VSIX extension ID extraction
- Given a theme path in the format `"/path/to/dracula.vsix::vsix::extension/themes/dracula.json"`
- When `extractExtensionId` is called
- Then it returns `"dracula"` (vsix filename minus `.vsix`)

AC-2: VSIX theme ID generation
- Given a theme path `"/path/to/monokai.vsix::vsix::extension/themes/dark-plus.json"` and extensionId `"monokai"`
- When `makeThemeId` is called
- Then it returns `"imported-monokai-dark-plus"`

AC-3: Backward compatibility -- filesystem paths still work
- Given a theme path `"/path/to/extensions/dracula-theme/theme/dracula.json"` (no `::vsix::`)
- When `extractExtensionId` is called
- Then it returns `"dracula-theme"` (grandparent directory, existing behavior)
- And when `makeThemeId` is called with that path and extensionId `"dracula-theme"`
- Then it returns `"imported-dracula-theme-dracula"`

AC-4: Settings.svelte uses shared makeThemeId
- Given Settings.svelte renders the theme list
- When a theme entry is rendered
- Then the theme ID is computed via the imported `makeThemeId` function, not inline logic
- And the active state highlight matches correctly for both vsix and filesystem themes

AC-5: makeThemeId is exported
- Given another module imports `makeThemeId` from `themeInit.js`
- When the import is resolved
- Then `makeThemeId` is available as a named export

## BDD Test Scenarios

### Scenario 1: VSIX path parsing

```gherkin
Feature: VSIX path handling in frontend

  Scenario: Extract extension ID from VSIX-encoded path
    Given themePath is "/Users/linus/.vscodium/extensions/dracula.vsix::vsix::extension/themes/dracula.json"
    When extractExtensionId(themePath) is called
    Then it returns "dracula"

  Scenario: Extract extension ID from VSIX with dots in filename
    Given themePath is "/path/to/publisher.theme-name-1.2.3.vsix::vsix::extension/themes/dark.json"
    When extractExtensionId(themePath) is called
    Then it returns "publisher.theme-name-1.2.3"

  Scenario: Generate theme ID from VSIX-encoded path
    Given themePath is "/path/to/monokai.vsix::vsix::extension/themes/dark-plus.json"
    And extensionId is "monokai"
    When makeThemeId(themePath, extensionId) is called
    Then it returns "imported-monokai-dark-plus"

  Scenario: Extract extension ID from regular filesystem path (backward compat)
    Given themePath is "/home/user/.vscodium/extensions/dracula-theme.theme-dracula-2.24.3/theme/dracula.json"
    When extractExtensionId(themePath) is called
    Then it returns "dracula-theme.theme-dracula-2.24.3"

  Scenario: Generate theme ID from regular filesystem path (backward compat)
    Given themePath is "/home/user/.vscodium/extensions/ext-name/themes/ocean.json"
    And extensionId is "ext-name"
    When makeThemeId(themePath, extensionId) is called
    Then it returns "imported-ext-name-ocean"
```

### Scenario 2: Settings.svelte integration

```gherkin
Feature: Settings theme list with VSIX paths

  Scenario: Theme list renders correctly with VSIX paths
    Given vscodiumThemes contains an entry with themePath "/path/to/nord.vsix::vsix::extension/themes/nord.json" and extensionId "nord"
    When the theme list is rendered
    Then the theme ID computed is "imported-nord-nord"
    And the active class is applied when currentThemeId matches

  Scenario: Theme list renders correctly with filesystem paths (backward compat)
    Given vscodiumThemes contains an entry with themePath "/path/to/ext/nord-ext/themes/arctic.json" and extensionId "nord-ext"
    When the theme list is rendered
    Then the theme ID computed is "imported-nord-ext-arctic"
```

## Tasks / Subtasks

- [ ] Task 1: Update `extractExtensionId` for VSIX paths (AC: AC-1, AC-3)
  - [ ] Add `::vsix::` detection guard at top of function
  - [ ] Split on `::vsix::`, take first part, extract filename, strip `.vsix` extension
  - [ ] Verify existing filesystem path logic remains in the else branch

- [ ] Task 2: Update and export `makeThemeId` for VSIX paths (AC: AC-2, AC-3, AC-5)
  - [ ] Add `::vsix::` detection guard at top of function
  - [ ] Split on `::vsix::`, take second part (internal path), extract filename, strip `.json`
  - [ ] Change `function makeThemeId` to `export function makeThemeId`
  - [ ] Verify existing filesystem path logic remains in the else branch

- [ ] Task 3: Update Settings.svelte to use shared `makeThemeId` (AC: AC-4)
  - [ ] Add `makeThemeId` to the import from `themeInit.js` on line 6
  - [ ] Replace line 171 inline logic with `makeThemeId(entry.themePath, entry.extensionId)`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on modified functions in `themeInit.js`
- [ ] `go build ./...` passes (ensures Wails bindings still compile)
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
