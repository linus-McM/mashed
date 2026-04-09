# Story 2: Editor Settings Store, Monaco Integration, and Settings UI

**Priority:** P0-critical
**Domain:** fullstack
**Estimated Complexity:** L
**Depends On:** edset-01
**Status:** ready

## Description

Create a frontend reactive store for editor settings, wire it into MonacoEditor.svelte to replace hardcoded options, initialize from the Go backend on app startup, and add a full "Editor" section to Settings.svelte with 13 grouped controls. This delivers user-visible control over Monaco editor behavior -- the core value of the editor settings feature.

## Developer Notes

### Architecture

- **New file:** `frontend/src/lib/stores/editorSettings.js` -- Svelte writable store
  - Pattern: follow `frontend/src/lib/stores/font.js` (writable store + exported helpers)
  - Export: `editorSettings` (writable store), `initEditorSettings(settings)`, `updateEditorSetting(key, value)`
  - `updateEditorSetting` must: update store, then call `SetEditorSettings()` with the full settings object (not individual field setters -- the backend has one atomic setter)
  - Default values in the store must match `DefaultEditorSettings()` from the Go backend

- **Modified file:** `frontend/src/components/MonacoEditor.svelte`
  - Import `editorSettings` store
  - Replace hardcoded values in `getEditorOptions()` with store values
  - Add reactive block: `$: if (editor && $editorSettings) { editor.updateOptions(mapSettingsToMonaco($editorSettings)); }`
  - Fix `lineHeight`: currently hardcoded to `1.5 * 13` -- should use `$currentFontSize` instead of literal 13
  - Map store fields to Monaco option names (most are 1:1 except `minimapEnabled` -> `minimap: { enabled: value }`)

- **Modified file:** `frontend/src/App.svelte`
  - Import `GetEditorSettings` from Wails bindings and `initEditorSettings` from store
  - In `onMount`, after theme/font initialization: `const es = await GetEditorSettings(); initEditorSettings(es);`

- **Modified file:** `frontend/src/views/Settings.svelte`
  - Add new "Editor" section in the right column (`col-settings`), after Font section
  - Import `editorSettings` and `updateEditorSetting` from store
  - Group controls into 4 subsections:
    - **Cursor**: cursorStyle (select: line, line-thin, block, block-outline, underline, underline-thin), cursorBlinking (select: blink, smooth, phase, expand, solid)
    - **Display**: wordWrap (select: off, on, wordWrapColumn, bounded), lineNumbers (select: on, off, relative, interval), renderLineHighlight (select: none, gutter, line, all), renderWhitespace (select: none, boundary, selection, trailing, all), minimapEnabled (toggle)
    - **Editing**: tabSize (number input 2-8), insertSpaces (toggle), bracketPairColorization (toggle)
    - **Behavior**: fontLigatures (toggle), scrollBeyondLastLine (toggle), smoothScrolling (toggle)
  - Each control calls `updateEditorSetting(key, value)` on change
  - Use the existing `settings-section` / `section-title` CSS classes for consistency

### Technical Considerations

- The store must be initialized before any MonacoEditor instance mounts, otherwise the reactive block will fire with undefined values. The App.svelte onMount runs before child components mount, so this ordering is safe.
- `editor.updateOptions()` is idempotent -- calling it with the same values is a no-op, so frequent reactive updates are fine
- Select elements should show the current value from the store using `bind:value` or by setting `value={$editorSettings.fieldName}`
- Toggle controls: use the existing pattern from Settings.svelte (checkbox-style or toggle button)

### Risks & Edge Cases

- If `GetEditorSettings` fails (unlikely but possible on first run), the store should fall back to its built-in defaults
- Monaco option names are case-sensitive -- map exactly: `cursorStyle`, `cursorBlinking`, `wordWrap`, `lineNumbers`, `renderWhitespace`, `renderLineHighlight`, `minimap.enabled`, `tabSize`, `insertSpaces`, `fontLigatures`, `scrollBeyondLastLine`, `smoothScrolling`
- The `bracketPairColorization` Monaco option is actually `bracketPairColorization.enabled` (nested) -- must handle this mapping
- `lineHeight` fix: `1.5 * $currentFontSize` not `1.5 * 13`

### Reference Files

- `frontend/src/lib/stores/font.js` -- store pattern to follow (writable + helpers)
- `frontend/src/components/MonacoEditor.svelte` lines 61-79 -- `getEditorOptions()` to replace
- `frontend/src/views/Settings.svelte` lines 187-242 -- existing Font/Sidebar sections for UI pattern
- `frontend/src/App.svelte` lines 43-56 -- onMount initialization pattern

## Acceptance Criteria

AC-1: Editor settings store is reactive and initialized from backend
- Given the app starts up
- When `GetEditorSettings` returns `{tabSize: 4, cursorStyle: "block", ...}`
- Then the `editorSettings` store contains `{tabSize: 4, cursorStyle: "block", ...}`
- And MonacoEditor instances reflect those settings

AC-2: Changing a setting updates Monaco in real-time
- Given a MonacoEditor is open with a file loaded
- When the user changes tabSize from 2 to 4 in Settings
- Then the editor immediately reflects tab size 4 (no page reload needed)
- And `SetEditorSettings` is called on the Go backend to persist

AC-3: Monaco hardcoded values are replaced by store values
- Given `getEditorOptions()` in MonacoEditor.svelte
- When the function is called
- Then `minimap.enabled`, `scrollBeyondLastLine`, `renderLineHighlight`, `wordWrap`, and `lineNumbers` come from the `editorSettings` store
- And `lineHeight` is computed as `1.5 * $currentFontSize` (not `1.5 * 13`)

AC-4: Settings UI shows all 13 editor controls in 4 groups
- Given the user navigates to Settings
- When the Editor section is visible
- Then it contains Cursor group (cursorStyle select, cursorBlinking select), Display group (wordWrap, lineNumbers, renderLineHighlight, renderWhitespace selects + minimap toggle), Editing group (tabSize number 2-8, insertSpaces toggle, bracketPairColorization toggle), Behavior group (fontLigatures, scrollBeyondLastLine, smoothScrolling toggles)

AC-5: Settings controls show current values from store
- Given editor settings are loaded with tabSize=4 and cursorStyle="block"
- When the user opens Settings
- Then the tabSize control shows 4
- And the cursorStyle select shows "block"

AC-6: Store falls back to defaults on initialization failure
- Given `GetEditorSettings` throws an error
- When the app starts
- Then the `editorSettings` store contains default values (tabSize=2, cursorStyle="line", etc.)
- And MonacoEditor still renders correctly

## BDD Test Scenarios

### Scenario 1: Store initialization

```gherkin
Feature: Editor Settings Store

  Scenario: Store initializes with backend values
    Given GetEditorSettings returns tabSize 4 and cursorStyle "block"
    When initEditorSettings is called with those values
    Then the editorSettings store contains tabSize 4
    And cursorStyle "block"

  Scenario: Store initializes with defaults when backend fails
    Given GetEditorSettings throws an error
    When the app mounts
    Then the editorSettings store contains tabSize 2
    And cursorStyle "line"
    And minimapEnabled false
```

### Scenario 2: Real-time Monaco update

```gherkin
  Scenario: Changing tabSize updates editor immediately
    Given a MonacoEditor instance is mounted
    And editorSettings store has tabSize 2
    When updateEditorSetting("tabSize", 4) is called
    Then editor.updateOptions is invoked with tabSize 4
    And SetEditorSettings is called with the full settings object

  Scenario: Changing minimap toggle updates editor
    Given a MonacoEditor instance is mounted
    And minimapEnabled is false in the store
    When updateEditorSetting("minimapEnabled", true) is called
    Then editor.updateOptions is invoked with minimap.enabled true
```

### Scenario 3: Settings UI controls

```gherkin
  Scenario: Editor section renders all control groups
    Given the Settings view is mounted
    When the Editor section is visible
    Then a select for cursorStyle exists with 6 options
    And a select for cursorBlinking exists with 5 options
    And a select for wordWrap exists with 4 options
    And a number input for tabSize exists with min 2 and max 8
    And toggles exist for minimapEnabled, insertSpaces, bracketPairColorization, fontLigatures, scrollBeyondLastLine, smoothScrolling

  Scenario: Changing a select triggers updateEditorSetting
    Given the Settings view is mounted
    When the user selects "block" in the cursorStyle dropdown
    Then updateEditorSetting("cursorStyle", "block") is called
```

### Scenario 4: lineHeight fix

```gherkin
  Scenario: lineHeight uses dynamic font size
    Given currentFontSize is 16
    When getEditorOptions is called
    Then lineHeight is 24 (1.5 * 16)
    And lineHeight is NOT 19.5 (1.5 * 13)
```

## Tasks / Subtasks

- [ ] Task 1: Create editorSettings store (AC: AC-1, AC-6)
  - [ ] Create `frontend/src/lib/stores/editorSettings.js` with writable store and default values
  - [ ] Implement `initEditorSettings(settings)` that merges backend values into store
  - [ ] Implement `updateEditorSetting(key, value)` that updates store and calls `SetEditorSettings`
  - [ ] Handle missing/undefined settings gracefully (fall back to defaults per field)

- [ ] Task 2: Wire MonacoEditor.svelte to store (AC: AC-2, AC-3)
  - [ ] Import `editorSettings` store in MonacoEditor.svelte
  - [ ] Replace hardcoded values in `getEditorOptions()` with store reads
  - [ ] Fix `lineHeight` to use `$currentFontSize` instead of literal 13
  - [ ] Add reactive block: `$: if (editor && $editorSettings) { editor.updateOptions(...) }`
  - [ ] Map `minimapEnabled` to `minimap: { enabled }` and `bracketPairColorization` to `bracketPairColorization: { enabled }`

- [ ] Task 3: Initialize store in App.svelte (AC: AC-1, AC-6)
  - [ ] Import `GetEditorSettings` from Wails bindings
  - [ ] Import `initEditorSettings` from store
  - [ ] Add initialization call in onMount after theme/font setup, wrapped in try/catch

- [ ] Task 4: Add Editor section to Settings.svelte (AC: AC-4, AC-5)
  - [ ] Import `editorSettings` and `updateEditorSetting` from store
  - [ ] Add "Editor" section with 4 subsection groups (Cursor, Display, Editing, Behavior)
  - [ ] Implement select controls for cursorStyle (6 options), cursorBlinking (5 options), wordWrap (4 options), lineNumbers (4 options), renderLineHighlight (4 options), renderWhitespace (5 options)
  - [ ] Implement number input for tabSize (2-8)
  - [ ] Implement toggle controls for minimapEnabled, insertSpaces, bracketPairColorization, fontLigatures, scrollBeyondLastLine, smoothScrolling
  - [ ] Bind each control to current store value and call `updateEditorSetting` on change

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
