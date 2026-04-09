# Story 3: Terminal Cursor Sync from Editor Settings

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** edset-02
**Status:** ready

## Description

Synchronize the Terminal.svelte xterm.js cursor appearance with the editor settings store so that cursor style and blink behavior are consistent across Monaco editors and embedded terminals. When a user changes cursor style or blinking in Settings, both Monaco and xterm.js update to match.

## Developer Notes

### Architecture

- **Modified file:** `frontend/src/components/Terminal.svelte`
  - Import `editorSettings` store from `frontend/src/lib/stores/editorSettings.js`
  - Map Monaco cursor styles to xterm equivalents on Terminal initialization and reactively on store changes
  - Mapping table:
    - Monaco `"line"` / `"line-thin"` -> xterm `"bar"`
    - Monaco `"underline"` / `"underline-thin"` -> xterm `"underline"`
    - Monaco `"block"` / `"block-outline"` -> xterm `"block"`
  - Map `cursorBlinking`:
    - Monaco `"blink"` / `"smooth"` / `"phase"` / `"expand"` -> xterm `cursorBlink: true`
    - Monaco `"solid"` -> xterm `cursorBlink: false`
  - Apply via `term.options.cursorStyle = ...` and `term.options.cursorBlink = ...` in a reactive block

### Technical Considerations

- xterm.js only supports 3 cursor styles: `"block"`, `"underline"`, `"bar"`. Monaco has 6, so we must map.
- xterm.js `cursorBlink` is a boolean, while Monaco has 5 blink animation types. We collapse to on/off.
- The Terminal component creates `term` in `onMount` (async import). The reactive block must guard against `term` being null during SSR/initial render.
- `cursorInactiveStyle` is currently hardcoded to `"outline"` -- keep this unchanged (it only affects unfocused terminals and has no Monaco equivalent).
- Current hardcoded values: `cursorBlink: true`, `cursorStyle: 'block'`. After this story, these come from the store.

### Risks & Edge Cases

- Terminal mounts before editor settings are loaded: the store has defaults (cursorStyle: "line", cursorBlinking: "blink"), which maps to xterm `"bar"` + `cursorBlink: true`. This is a reasonable fallback.
- If `editorSettings` store is undefined/empty, the mapping functions must return safe defaults (bar + blink).
- Multiple Terminal instances (e.g., multiple tabs) must all react to the same store. Svelte's reactive `$:` handles this automatically since all instances subscribe to the same store.
- Log-only terminals (no paneTarget, `disableStdin: true`) still show a cursor sometimes -- the cursor settings should apply uniformly.

### Reference Files

- `frontend/src/components/Terminal.svelte` lines 77-87 -- current Terminal initialization with hardcoded cursor
- `frontend/src/lib/stores/editorSettings.js` -- the store to import (created in edset-02)
- `frontend/src/lib/stores/font.js` -- pattern for importing stores in components

## Acceptance Criteria

AC-1: Terminal cursor style matches editor settings
- Given editor settings have cursorStyle "block"
- When a Terminal component mounts
- Then xterm.js `cursorStyle` is set to "block"

AC-2: Terminal cursor style updates reactively
- Given a Terminal component is mounted with cursorStyle "bar" (from Monaco "line")
- When the user changes cursorStyle to "underline" in Settings
- Then the terminal cursor changes to "underline" without remounting

AC-3: Monaco cursor styles map correctly to xterm styles
- Given the mapping function receives Monaco style "line"
- When the function is called
- Then it returns xterm style "bar"
- And "line-thin" also returns "bar"
- And "block" returns "block"
- And "block-outline" returns "block"
- And "underline" returns "underline"
- And "underline-thin" returns "underline"

AC-4: Cursor blinking maps correctly
- Given editor settings have cursorBlinking "solid"
- When a Terminal component mounts
- Then xterm.js `cursorBlink` is `false`
- And for cursorBlinking "blink", xterm.js `cursorBlink` is `true`
- And for cursorBlinking "smooth", xterm.js `cursorBlink` is `true`

## BDD Test Scenarios

### Scenario 1: Initial cursor from store

```gherkin
Feature: Terminal Cursor Sync

  Scenario: Terminal initializes with store cursor style
    Given editorSettings store has cursorStyle "block" and cursorBlinking "blink"
    When a Terminal component mounts
    Then xterm Terminal is created with cursorStyle "block"
    And cursorBlink true

  Scenario: Terminal initializes with line cursor
    Given editorSettings store has cursorStyle "line" and cursorBlinking "solid"
    When a Terminal component mounts
    Then xterm Terminal is created with cursorStyle "bar"
    And cursorBlink false
```

### Scenario 2: Reactive cursor update

```gherkin
  Scenario: Cursor style changes after mount
    Given a Terminal is mounted with cursorStyle "bar"
    When editorSettings.cursorStyle changes to "underline"
    Then term.options.cursorStyle is set to "underline"

  Scenario: Cursor blink changes after mount
    Given a Terminal is mounted with cursorBlink true
    When editorSettings.cursorBlinking changes to "solid"
    Then term.options.cursorBlink is set to false
```

### Scenario 3: Mapping edge cases

```gherkin
  Scenario: All 6 Monaco styles map to valid xterm styles
    Given the cursor mapping function
    When called with each of "line", "line-thin", "block", "block-outline", "underline", "underline-thin"
    Then results are "bar", "bar", "block", "block", "underline", "underline" respectively
    And no undefined or null values are returned

  Scenario: Unknown cursor style falls back to "bar"
    Given the cursor mapping function
    When called with an unknown style "custom"
    Then it returns "bar" as the default
```

### Scenario 4: Multiple terminals sync

```gherkin
  Scenario: Two terminals update simultaneously
    Given two Terminal components are mounted
    When editorSettings.cursorStyle changes to "block"
    Then both terminals show block cursors
```

## Tasks / Subtasks

- [ ] Task 1: Add cursor mapping utilities (AC: AC-3)
  - [ ] Create `mapMonacoCursorToXterm(monacoStyle)` function in Terminal.svelte (or extract to a shared util)
  - [ ] Create `mapCursorBlinkToXterm(monacoBlinking)` function
  - [ ] Handle unknown values with safe defaults ("bar" and true)

- [ ] Task 2: Wire store into Terminal initialization (AC: AC-1, AC-4)
  - [ ] Import `editorSettings` store in Terminal.svelte
  - [ ] Replace hardcoded `cursorBlink: true` with `cursorBlink: mapCursorBlinkToXterm($editorSettings.cursorBlinking)`
  - [ ] Replace hardcoded `cursorStyle: 'block'` with `cursorStyle: mapMonacoCursorToXterm($editorSettings.cursorStyle)`

- [ ] Task 3: Add reactive update block (AC: AC-2)
  - [ ] Add `$: if (term && $editorSettings) { ... }` reactive block
  - [ ] Set `term.options.cursorStyle` and `term.options.cursorBlink` from mapped store values
  - [ ] Guard against term being null

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
