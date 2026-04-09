# Story 1: Backend Editor Settings Config

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** none
**Status:** ready

## Description

Add an `EditorSettings` struct to the Go backend with persistence via `mashedConfig`. This provides the data model and Wails-bound methods (`DefaultEditorSettings`, `GetEditorSettings`, `SetEditorSettings`) that the frontend will consume to control Monaco editor behavior. Without this foundation, no editor customization can be persisted across sessions.

## Developer Notes

### Architecture

- **File:** `app.go` -- add `EditorSettings` struct and `*EditorSettings` pointer field on `mashedConfig`
- The pointer + `omitempty` pattern ensures backward compatibility with existing `~/.mashed/config.json` files that lack the field
- Three new methods on `*App`:
  - `DefaultEditorSettings() EditorSettings` -- returns sensible defaults (no receiver state needed, but bound to App for Wails)
  - `GetEditorSettings() EditorSettings` -- loads config, returns `EditorSettings` or defaults if nil
  - `SetEditorSettings(settings EditorSettings) error` -- validates, merges into config, persists

### Technical Considerations

- Follow the exact pattern of existing setters (`SetTheme`, `SetMonoFont`, etc.): lock `a.mu`, call `loadConfig()`, mutate, call `saveConfig()`
- `EditorSettings` struct fields and their JSON tags must match the plan exactly -- the frontend store will mirror these field names
- Validation in `SetEditorSettings`: `TabSize` must be clamped to 2-8, `WordWrap` must be one of `"off"`, `"on"`, `"wordWrapColumn"`, `"bounded"`, `CursorStyle` must be one of Monaco's 6 options, etc. Return `fmt.Errorf(...)` for invalid values
- `DefaultEditorSettings()` returns: `MinimapEnabled: false`, `WordWrap: "off"`, `LineNumbers: "on"`, `RenderWhitespace: "none"`, `TabSize: 2`, `InsertSpaces: true`, `CursorStyle: "line"`, `CursorBlinking: "blink"`, `BracketPairColorization: true`, `RenderLineHighlight: "line"`, `FontLigatures: false`, `ScrollBeyondLastLine: false`, `SmoothScrolling: false`

### Risks & Edge Cases

- Existing config files have no `editorSettings` key -- `GetEditorSettings` must handle `nil` pointer gracefully by returning `DefaultEditorSettings()`
- Concurrent access to config: the `a.mu` mutex already protects load/save; follow the same pattern
- Invalid enum values from a hand-edited config file should not crash -- validate on read as well as write, or fall back to defaults

### Reference Files

- `app.go` lines 77-85 (`mashedConfig` struct) -- add field here
- `app.go` lines 323-365 (`SetTheme`, `SetMonoFont`, etc.) -- follow this setter pattern
- `app.go` lines 99-121 (`loadConfig`, `saveConfig`) -- config persistence
- `theme_scanner_test.go` -- test patterns (testify, table-driven)

## Acceptance Criteria

AC-1: EditorSettings struct exists with all 13 fields
- Given the Go codebase
- When `app.go` is compiled
- Then `EditorSettings` has fields: MinimapEnabled (bool), WordWrap (string), LineNumbers (string), RenderWhitespace (string), TabSize (int), InsertSpaces (bool), CursorStyle (string), CursorBlinking (string), BracketPairColorization (bool), RenderLineHighlight (string), FontLigatures (bool), ScrollBeyondLastLine (bool), SmoothScrolling (bool)
- And all fields have correct `json:"..."` tags matching camelCase

AC-2: GetEditorSettings returns defaults when config has no editorSettings
- Given a `~/.mashed/config.json` that does not contain an `editorSettings` key
- When `GetEditorSettings()` is called
- Then it returns an `EditorSettings` with all default values (e.g., TabSize=2, CursorStyle="line", MinimapEnabled=false)

AC-3: SetEditorSettings persists and validates
- Given valid editor settings with TabSize=4 and CursorStyle="block"
- When `SetEditorSettings(settings)` is called
- Then the config file at `~/.mashed/config.json` contains `"editorSettings"` with the provided values
- And a subsequent `GetEditorSettings()` returns the same values

AC-4: SetEditorSettings rejects invalid values
- Given editor settings with TabSize=0 or CursorStyle="invalid"
- When `SetEditorSettings(settings)` is called
- Then it returns a non-nil error describing the invalid field
- And the config file is not modified

AC-5: Backward compatibility with existing configs
- Given a config file containing only `{"devDir":"/Users/x/dev","theme":"mashed-dark"}`
- When `loadConfig()` parses the file
- Then no error occurs
- And `cfg.EditorSettings` is nil (not a zero-value struct)

## BDD Test Scenarios

### Scenario 1: Default settings

```gherkin
Feature: Editor Settings Backend

  Scenario: Fresh install returns defaults
    Given a config file with no editorSettings key
    When GetEditorSettings is called
    Then it returns EditorSettings with TabSize 2
    And CursorStyle "line"
    And MinimapEnabled false
    And WordWrap "off"
    And BracketPairColorization true

  Scenario: DefaultEditorSettings returns consistent values
    Given no prior configuration
    When DefaultEditorSettings is called
    Then all fields match the documented defaults
```

### Scenario 2: Persistence round-trip

```gherkin
  Scenario: Settings are persisted and retrieved
    Given a temporary config directory
    When SetEditorSettings is called with TabSize 4 and CursorStyle "block"
    Then GetEditorSettings returns TabSize 4 and CursorStyle "block"
    And the config file contains "editorSettings" JSON key

  Scenario: Partial update preserves other config fields
    Given a config file with devDir "/Users/x/dev" and theme "mashed-dark"
    When SetEditorSettings is called with TabSize 4
    Then the config file still contains devDir "/Users/x/dev"
    And the config file still contains theme "mashed-dark"
```

### Scenario 3: Validation rejects bad values

```gherkin
  Scenario: TabSize out of range
    Given editor settings with TabSize 0
    When SetEditorSettings is called
    Then an error is returned containing "tabSize"
    And the config file is unchanged

  Scenario: Invalid CursorStyle
    Given editor settings with CursorStyle "blinky"
    When SetEditorSettings is called
    Then an error is returned containing "cursorStyle"

  Scenario: Invalid WordWrap value
    Given editor settings with WordWrap "maybe"
    When SetEditorSettings is called
    Then an error is returned containing "wordWrap"
```

### Scenario 4: Concurrent access

```gherkin
  Scenario: Concurrent SetEditorSettings calls
    Given two goroutines calling SetEditorSettings simultaneously
    When both complete
    Then no data race is detected (go test -race)
    And the config file is valid JSON
```

## Tasks / Subtasks

- [ ] Task 1: Define EditorSettings struct and mashedConfig field (AC: AC-1, AC-5)
  - [ ] Add `EditorSettings` struct with 13 fields and JSON tags to `app.go`
  - [ ] Add `EditorSettings *EditorSettings \`json:"editorSettings,omitempty"\`` to `mashedConfig`
  - [ ] Verify `go build ./...` passes

- [ ] Task 2: Implement DefaultEditorSettings and GetEditorSettings (AC: AC-2)
  - [ ] Add `DefaultEditorSettings() EditorSettings` method on `*App`
  - [ ] Add `GetEditorSettings() EditorSettings` method that returns defaults when `cfg.EditorSettings == nil`
  - [ ] Write table-driven tests for default values

- [ ] Task 3: Implement SetEditorSettings with validation (AC: AC-3, AC-4)
  - [ ] Add `validateEditorSettings(es EditorSettings) error` helper with enum/range checks
  - [ ] Add `SetEditorSettings(settings EditorSettings) error` method following existing mutex pattern
  - [ ] Write tests: valid round-trip, invalid TabSize, invalid CursorStyle, invalid WordWrap, invalid LineNumbers, invalid RenderWhitespace, invalid CursorBlinking, invalid RenderLineHighlight
  - [ ] Write concurrent access test with `-race`

- [ ] Task 4: Backward compatibility test (AC: AC-5)
  - [ ] Write test: parse legacy config JSON without editorSettings key
  - [ ] Write test: config with editorSettings is parsed correctly
  - [ ] Verify `EditorSettings` field is nil (not zero-struct) when absent

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
