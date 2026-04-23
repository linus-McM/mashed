# Story 01: Backend Config — MarkdownMenuSettings and Wails bindings

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** none
**Status:** ready
**UI-facing:** no

## Description

Add the `MarkdownMenuSettings` struct, persist it via `mashedConfig.MarkdownMenu`, and expose three Wails-bound methods (`DefaultMarkdownMenuSettings`, `GetMarkdownMenuSettings`, `SetMarkdownMenuSettings`). Regenerate Wails TypeScript bindings so the frontend can consume them. This is the foundation for all subsequent frontend work; no UI changes land until the backend contract is frozen.

## Developer Notes

### Architecture
- File to modify: `/Users/linus/Development/mashed/app.go`
- Add new struct next to existing `EditorSettings` (currently at the `type EditorSettings struct` block). Mirror its shape and placement conventions.
- Extend `mashedConfig` struct — the private config type used by `loadConfig`/`saveConfig`. Add `MarkdownMenu *MarkdownMenuSettings \`json:"markdownMenu,omitempty"\``. Pointer + omitempty so a nil value omits the key entirely and `GetMarkdownMenuSettings` can distinguish "never set" from "explicitly set".
- `GetConfig()` already returns `mashedConfig` — the new field flows through automatically once added to the struct; no additional edits to `GetConfig` needed.
- Wails method signatures (exact):

```go
func (a *App) DefaultMarkdownMenuSettings() MarkdownMenuSettings
func (a *App) GetMarkdownMenuSettings() MarkdownMenuSettings
func (a *App) SetMarkdownMenuSettings(s MarkdownMenuSettings) error
```

- `GetMarkdownMenuSettings` MUST return `a.DefaultMarkdownMenuSettings()` when `cfg.MarkdownMenu == nil`. Do NOT mutate the on-disk config on read.
- `SetMarkdownMenuSettings` uses the same load-modify-save pattern as `SetEditorSettings` (`loadConfig()` → mutate → `saveConfig()`). No validation is needed — all fields are `bool`, so every value is valid.

### Technical Considerations
- Concurrency: `SetMarkdownMenuSettings` must be safe against concurrent writers. Review how `SetEditorSettings` handles this — if it takes a mutex, follow the same pattern. If it relies on the atomic write inside `saveConfig`, do the same.
- Errors: return wrapped errors from `saveConfig` using `fmt.Errorf("save markdown menu settings: %w", err)` style if saveConfig returns an error. No new sentinels required.
- JSON tags MUST be exactly: `bold`, `italic`, `strikethrough`, `code`, `link`, `latex` (lowercase, no abbreviations — the frontend store and Settings panel expect these exact keys).

### Wails Binding Regeneration
After the Go edits, run `wails generate module` from the repo root. This updates:
- `/Users/linus/Development/mashed/frontend/wailsjs/go/main/App.js` — adds three named exports
- `/Users/linus/Development/mashed/frontend/wailsjs/go/main/App.d.ts` — adds TypeScript declarations
- `/Users/linus/Development/mashed/frontend/wailsjs/go/models.ts` — adds `main.MarkdownMenuSettings`

Commit the regenerated files alongside the Go changes. The frontend stories import from these paths.

### Default Values (authoritative — DO NOT change without updating the plan)

| Field          | Default |
| -------------- | ------- |
| Bold           | true    |
| Italic         | true    |
| Strikethrough  | true    |
| Code           | true    |
| Link           | true    |
| Latex          | false   |

### Risks & Edge Cases
- **Config file missing entirely** — `loadConfig()` returns zero value; `cfg.MarkdownMenu == nil`; `GetMarkdownMenuSettings` must return defaults without writing to disk.
- **Partial config file** — existing users have configs without the `markdownMenu` key; unmarshaling leaves `MarkdownMenu` nil — same path as above.
- **Concurrent Set calls** — ensure both saves don't interleave and corrupt the file. Follow whatever pattern `SetEditorSettings` uses (this was already solved).
- **omitempty on pointer** — a nil pointer serialises to absent key, NOT `null`. Test this explicitly.

### Reference Files
- `/Users/linus/Development/mashed/app.go` — existing `EditorSettings`, `DefaultEditorSettings`, `GetEditorSettings`, `SetEditorSettings`, `mashedConfig`, `loadConfig`, `saveConfig` are the exact template.
- `/Users/linus/Development/mashed/editor_settings_test.go` — test-file template for story 09 and the test block in this story's DoD.

## Acceptance Criteria

AC-1: Defaults are exactly the values in the table
- Given no config file exists (or `markdownMenu` key absent)
- When I call `DefaultMarkdownMenuSettings()`
- Then the result is `{Bold:true, Italic:true, Strikethrough:true, Code:true, Link:true, Latex:false}`

AC-2: GetMarkdownMenuSettings returns defaults when unset
- Given `cfg.MarkdownMenu == nil` after `loadConfig()`
- When `GetMarkdownMenuSettings()` is called
- Then it returns `DefaultMarkdownMenuSettings()` byte-for-byte
- And the config file on disk is unchanged (no write triggered by read)

AC-3: Set/Get round-trip persists and reloads identically
- Given `SetMarkdownMenuSettings({Bold:false, Italic:true, Strikethrough:false, Code:true, Link:false, Latex:true})` succeeds
- When a new `App` instance is constructed and `GetMarkdownMenuSettings()` is called
- Then the returned struct equals the value set

AC-4: omitempty — nil MarkdownMenu serialises without the key
- Given a `mashedConfig` with `MarkdownMenu == nil` but `DevDir` set
- When it is marshalled to JSON and the raw bytes inspected
- Then the bytes do NOT contain the substring `"markdownMenu"`
- And unmarshaling those bytes back yields `MarkdownMenu == nil`

AC-5: Set preserves other config fields
- Given a config with `devDir`, `theme`, `monoFont` set
- When `SetMarkdownMenuSettings` is called with any valid value
- Then after reload the original `devDir`, `theme`, `monoFont` are preserved unchanged

AC-6: Wails bindings are regenerated
- Given the three methods exist on `*App` with the signatures above
- When `wails generate module` runs
- Then `frontend/wailsjs/go/main/App.js` exports `DefaultMarkdownMenuSettings`, `GetMarkdownMenuSettings`, `SetMarkdownMenuSettings`
- And `frontend/wailsjs/go/models.ts` contains `main.MarkdownMenuSettings` with the six boolean fields

## BDD Test Scenarios

```gherkin
Feature: Persist markdown menu settings

  Scenario: Default values round-trip
    Given a fresh App with no existing config file
    When DefaultMarkdownMenuSettings is called
    Then the returned struct has Bold true
    And Italic true
    And Strikethrough true
    And Code true
    And Link true
    And Latex false

  Scenario: GetMarkdownMenuSettings falls back to defaults
    Given a config file containing only {"devDir":"/tmp"}
    When GetMarkdownMenuSettings is called
    Then the returned struct equals DefaultMarkdownMenuSettings
    And the config file on disk has not been modified

  Scenario: Toggle LaTeX on, reload
    Given default settings are persisted
    When SetMarkdownMenuSettings is called with Latex=true and all others default
    And a new App reads the config
    Then GetMarkdownMenuSettings returns Latex=true and all other defaults

  Scenario: omitempty keeps JSON clean
    Given a mashedConfig with MarkdownMenu=nil and DevDir="/tmp"
    When the config is marshalled to JSON
    Then the JSON string does not contain "markdownMenu"

  Scenario: Set preserves unrelated fields
    Given a config with devDir="/work" and monoFont="JetBrains Mono"
    When SetMarkdownMenuSettings is called with all defaults
    Then reloading the config preserves devDir="/work" and monoFont="JetBrains Mono"

  Scenario: Concurrent Set calls are safe
    Given 10 goroutines each call SetMarkdownMenuSettings with distinct values
    When they run with -race
    Then the race detector reports no data races
    And the persisted value equals one of the 10 written values (last-writer-wins is acceptable)
```

## Tasks / Subtasks

- [ ] Task 1: Define types (AC-1, AC-4) — backend
  - [ ] Add `MarkdownMenuSettings` struct to `app.go` next to `EditorSettings` with the six bool fields and exact JSON tags.
  - [ ] Add `MarkdownMenu *MarkdownMenuSettings \`json:"markdownMenu,omitempty"\`` to `mashedConfig`.
- [ ] Task 2: Implement methods (AC-1, AC-2, AC-3, AC-5) — backend
  - [ ] `DefaultMarkdownMenuSettings` returns the literal struct from the defaults table.
  - [ ] `GetMarkdownMenuSettings` loads config and returns defaults if `MarkdownMenu == nil`, else dereferences.
  - [ ] `SetMarkdownMenuSettings` load-modify-save, mirroring `SetEditorSettings` concurrency pattern.
- [ ] Task 3: Regenerate Wails bindings (AC-6) — backend
  - [ ] Run `wails generate module` from repo root.
  - [ ] Verify `frontend/wailsjs/go/main/App.js`, `App.d.ts`, and `models.ts` contain the new exports/types.
  - [ ] Commit the regenerated files.
- [ ] Task 4: Unit tests (AC-1, AC-2, AC-3, AC-4, AC-5) — backend
  - [ ] Add `TestDefaultMarkdownMenuSettings` to `app_test.go` (or a new `markdown_menu_test.go` adjacent to `editor_settings_test.go`).
  - [ ] Add `TestGetMarkdownMenuSettings_DefaultsWhenNil`.
  - [ ] Add `TestSetGetMarkdownMenuSettings_RoundTrip`.
  - [ ] Add `TestConfigMarkdownMenuOmitempty`.
  - [ ] Add `TestSetMarkdownMenuSettings_PreservesOtherConfig`.
  - [ ] Add concurrent-access test run under `-race`.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `app.go` lines added by this story (new MarkdownMenu-related code)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `wails generate module` has been run and regenerated files are committed
- [ ] `/simplify` run on all modified Go files
- [ ] Code review: no CRITICAL/HIGH issues
