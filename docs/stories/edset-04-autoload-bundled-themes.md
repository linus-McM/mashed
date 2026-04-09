# Story 4: Auto-Load Bundled Themes from ./themes/

**Priority:** P1-high
**Domain:** fullstack
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Automatically load the 15 bundled `.vsix` theme files from the `./themes/` directory on app startup so they appear in the Settings theme list without requiring users to manually configure a VSCodium extension path, scan, and click-to-import. This eliminates the onboarding friction of having themes available in the repo but invisible in the UI.

## Developer Notes

### Architecture

**Backend changes -- `theme_scanner.go`:**

- Extract the inner loop body of `ListVSCodiumThemes` (lines 231-277) into a new function: `scanVSIXDirectory(dir string) ([]VSCodeThemeEntry, error)`. This is a pure refactor -- `ListVSCodiumThemes` calls `scanVSIXDirectory(extDir)` after its validation.
- Add `bundledThemesDir() string` -- returns the absolute path to the `./themes/` directory relative to the executable. Use `os.Executable()` to find the binary location, then resolve `../themes/` (for dev) or the appropriate bundled path. For dev mode, fall back to `./themes/` relative to CWD.
- Add `ListBundledThemes() ([]VSCodeThemeEntry, error)` method on `*App` -- calls `bundledThemesDir()` then `scanVSIXDirectory()`. Returns empty slice (not error) if the themes directory doesn't exist.
- Add `ReadBundledThemeFile(themePath string) (string, error)` method on `*App` -- reads a theme file from a bundled VSIX. Uses the same `readThemeFromVSIX` logic as `ReadThemeFile` but without requiring VSCodiumExtPath to be configured. The `themePath` is a full VSIX-encoded path (`/abs/path/to/file.vsix::vsix::extension/themes/dark.json`).

**Frontend changes -- `frontend/src/lib/themeInit.js`:**

- Add `loadBundledThemes()` exported async function:
  1. Call `ListBundledThemes()` to get all theme entries
  2. For each entry, compute `themeId` via `makeThemeId(entry.themePath, entry.extensionId)`
  3. Skip if theme ID already exists in the `allThemes` store (user may have imported it manually)
  4. Call `ReadBundledThemeFile(entry.themePath)` to get raw JSON
  5. Parse, convert via `convertVSCodeTheme()`, validate via `validateConvertedTheme()`
  6. Register via `registerImportedTheme(id, theme)` and persist via `SaveTheme(id, JSON.stringify(theme))`
  7. Log and skip individual theme failures (don't abort the entire batch)

**Frontend changes -- `frontend/src/App.svelte`:**

- Import `loadBundledThemes` from `themeInit.js`
- In `onMount`, after `loadSavedThemes()` and theme application: `await loadBundledThemes();`
- This must run after `loadSavedThemes()` so the "already imported" check works

### Technical Considerations

- `bundledThemesDir()` path resolution: In dev mode (`wails dev`), the CWD is the project root, so `./themes/` works. In production (`wails build`), themes may be embedded or in a sibling directory -- use `os.Executable()` resolution with fallback.
- `ReadBundledThemeFile` needs to handle VSIX paths without requiring `VSCodiumExtPath` config. The path is absolute (from `ListBundledThemes`), so no prefix resolution is needed -- just open the VSIX and read the internal path.
- Theme import should be idempotent: running `loadBundledThemes()` on every startup should not duplicate themes that were already saved in `~/.mashed/themes.json`.
- The `SaveTheme` call persists each bundled theme to `~/.mashed/themes.json` so it's available even if the themes directory is later moved. This also means the theme survives app updates.
- Performance: 15 VSIX files, each requiring zip open + JSON parse + conversion. This is fast (< 1s total) but should still be non-blocking (it runs in async onMount).

### Risks & Edge Cases

- `./themes/` directory doesn't exist (clean checkout without submodules, or production build without bundled themes): `ListBundledThemes` returns empty slice, `loadBundledThemes` is a no-op. No error shown to user.
- Corrupt VSIX file in themes directory: skip with log warning, continue with remaining themes
- Theme that was bundled but later removed from `./themes/`: it remains in `~/.mashed/themes.json` from a previous import. This is acceptable behavior (user explicitly removes via Settings if unwanted).
- VSIX contains non-JSON theme files (YAML, tmTheme): skip them (the existing `ListVSCodiumThemes` already filters for `.json` suffix)
- `ReadBundledThemeFile` must handle `include` directives (theme inheritance) within the VSIX -- the existing `readThemeFromVSIX` already resolves includes up to depth 5

### Reference Files

- `theme_scanner.go` lines 205-285 -- `ListVSCodiumThemes` to refactor
- `theme_scanner.go` lines 299-314 -- `ReadThemeFile` pattern for `ReadBundledThemeFile`
- `theme_scanner.go` lines 39-54 -- `readFileFromZip` helper (reused)
- `frontend/src/lib/themeInit.js` lines 61-71 -- `loadSavedThemes` pattern
- `frontend/src/lib/themeInit.js` -- `activateImportedTheme` flow for convert/register/save pattern
- `frontend/src/App.svelte` lines 46-48 -- theme initialization in onMount
- `theme_scanner_test.go` -- existing test patterns for theme scanning

## Acceptance Criteria

AC-1: ListBundledThemes returns themes from ./themes/ directory
- Given the `./themes/` directory contains 15 `.vsix` files with color themes
- When `ListBundledThemes()` is called
- Then it returns a slice of `VSCodeThemeEntry` structs with Label, ExtensionID, ThemePath, and UITheme for each theme found
- And ThemePath uses the VSIX-encoded format (`/path/file.vsix::vsix::extension/themes/dark.json`)

AC-2: ListBundledThemes returns empty slice when themes directory missing
- Given no `./themes/` directory exists
- When `ListBundledThemes()` is called
- Then it returns an empty slice and nil error (not an error)

AC-3: ReadBundledThemeFile reads theme JSON from bundled VSIX
- Given a valid VSIX-encoded theme path from `ListBundledThemes`
- When `ReadBundledThemeFile(themePath)` is called
- Then it returns the theme JSON string with comments stripped and includes resolved
- And it does NOT require VSCodiumExtPath to be configured

AC-4: Bundled themes appear in Settings theme list on startup
- Given the app starts with 15 bundled VSIX files in `./themes/`
- When the user opens Settings
- Then all bundled themes appear in the left theme list alongside built-in themes
- And each can be selected and applied

AC-5: Bundled theme import is idempotent
- Given bundled themes were already imported on a previous startup
- When the app starts again
- Then no duplicate themes appear in the theme list
- And no errors occur

AC-6: scanVSIXDirectory is a clean refactor of ListVSCodiumThemes
- Given the existing `ListVSCodiumThemes` method
- When it is refactored to use `scanVSIXDirectory(dir)`
- Then `ListVSCodiumThemes` produces identical results as before the refactor
- And existing theme scanning tests still pass

## BDD Test Scenarios

### Scenario 1: Bundled theme discovery

```gherkin
Feature: Auto-Load Bundled Themes

  Scenario: Discover themes in bundled directory
    Given a themes directory containing "monokai.theme-monokai-pro-vscode-2.0.13.vsix"
    And the VSIX contains a package.json with contributes.themes entries
    When ListBundledThemes is called
    Then the result includes an entry with ExtensionID "monokai.theme-monokai-pro-vscode-2.0.13"
    And ThemePath contains "::vsix::" separator

  Scenario: Empty themes directory
    Given a themes directory with no .vsix files
    When ListBundledThemes is called
    Then the result is an empty slice
    And no error is returned

  Scenario: Missing themes directory
    Given the themes directory does not exist
    When ListBundledThemes is called
    Then the result is an empty slice
    And no error is returned
```

### Scenario 2: Theme file reading

```gherkin
  Scenario: Read theme JSON from bundled VSIX
    Given a VSIX-encoded theme path from ListBundledThemes
    When ReadBundledThemeFile is called with that path
    Then it returns valid JSON
    And the JSON contains "colors" and "tokenColors" keys

  Scenario: Read theme with include directive
    Given a theme JSON that includes another theme file within the same VSIX
    When ReadBundledThemeFile is called
    Then the include is resolved and colors are merged
```

### Scenario 3: Frontend auto-load

```gherkin
  Scenario: Bundled themes loaded on startup
    Given ListBundledThemes returns 3 theme entries
    And none are in the allThemes store
    When loadBundledThemes is called
    Then all 3 themes are registered in allThemes store
    And all 3 themes are persisted via SaveTheme

  Scenario: Already-imported bundled themes are skipped
    Given the allThemes store already contains "imported-monokai-pro-dark"
    And ListBundledThemes returns an entry that maps to "imported-monokai-pro-dark"
    When loadBundledThemes is called
    Then the theme is not re-imported or duplicated
    And SaveTheme is not called for that theme

  Scenario: Corrupt VSIX is skipped gracefully
    Given ListBundledThemes returns 3 entries
    And ReadBundledThemeFile fails for the 2nd entry
    When loadBundledThemes is called
    Then the 1st and 3rd themes are imported successfully
    And a console warning is logged for the 2nd entry
```

### Scenario 4: Refactor preserves behavior

```gherkin
  Scenario: scanVSIXDirectory produces same results as old ListVSCodiumThemes
    Given a directory with 3 VSIX files containing themes
    When scanVSIXDirectory is called on that directory
    Then the returned entries match what ListVSCodiumThemes would have returned
    And entries are sorted alphabetically by label
```

## Tasks / Subtasks

- [ ] Task 1: Refactor ListVSCodiumThemes to extract scanVSIXDirectory (AC: AC-6)
  - [ ] Extract scanning loop (lines 231-277) from `ListVSCodiumThemes` into `scanVSIXDirectory(dir string) ([]VSCodeThemeEntry, error)`
  - [ ] Update `ListVSCodiumThemes` to call `scanVSIXDirectory(extDir)` after validation
  - [ ] Verify all existing `theme_scanner_test.go` tests still pass
  - [ ] Add unit test for `scanVSIXDirectory` directly

- [ ] Task 2: Add bundledThemesDir, ListBundledThemes, ReadBundledThemeFile (AC: AC-1, AC-2, AC-3)
  - [ ] Implement `bundledThemesDir() string` with executable-relative and CWD fallback resolution
  - [ ] Implement `ListBundledThemes() ([]VSCodeThemeEntry, error)` -- returns empty slice if dir missing
  - [ ] Implement `ReadBundledThemeFile(themePath string) (string, error)` -- reads from absolute VSIX path
  - [ ] Write unit tests: themes found, empty dir, missing dir, corrupt VSIX skip, theme file read

- [ ] Task 3: Add loadBundledThemes to themeInit.js (AC: AC-4, AC-5)
  - [ ] Import `ListBundledThemes` and `ReadBundledThemeFile` from Wails bindings
  - [ ] Implement `loadBundledThemes()`: iterate entries, skip existing, convert, register, save
  - [ ] Handle individual theme failures with console.warn (no abort)
  - [ ] Verify idempotency: second call with same themes does nothing

- [ ] Task 4: Wire loadBundledThemes into App.svelte startup (AC: AC-4)
  - [ ] Import `loadBundledThemes` from themeInit.js
  - [ ] Call `await loadBundledThemes()` in onMount after `loadSavedThemes()` and theme application
  - [ ] Wrap in try/catch (theme loading failure should not block app startup)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
