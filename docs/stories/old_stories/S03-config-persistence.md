# Story 3: Config Persistence + Edit Menu

**Priority:** P0-critical
**Domain:** fullstack
**Estimated Complexity:** M
**Depends On:** Story 1
**Status:** ready

## Description

Extend the Go backend config system to persist theme selection and VSCodium extension path, expose new Wails-bound methods for the frontend to read and write these settings, and add a hidden native Edit menu to `main.go` so that Cmd+C/V/X work in the frameless window. The frontend loads the persisted theme on startup and applies it before the user sees any content.

## Developer Notes

### Architecture

- **Modified file: `app.go`** -- Extend `conductorConfig` struct (line 43-45), add three new exported methods: `GetConfig()`, `SetTheme()`, `SetVSCodiumExtPath()`.
- **Modified file: `main.go`** -- Add hidden native menu with `menu.EditMenu()` for Cmd+C/V/X support (lines 17-18 area, before `wails.Run`).
- **Modified file: `frontend/src/App.svelte`** -- On mount, call `GetConfig()` to load persisted theme, call `applyTheme()` before rendering content.

### Go Changes (app.go)

Current `conductorConfig` at line 43-45:
```go
type conductorConfig struct {
    DevDir string `json:"devDir"`
}
```

Extend to:
```go
type conductorConfig struct {
    DevDir          string `json:"devDir"`
    Theme           string `json:"theme,omitempty"`
    VSCodiumExtPath string `json:"vscodiumExtPath,omitempty"`
}
```

New exported methods (Wails-bound):

```go
// GetConfig returns the current persisted config for the frontend.
func (a *App) GetConfig() conductorConfig {
    return loadConfig()
}

// SetTheme persists the theme ID to config.
func (a *App) SetTheme(id string) error {
    cfg := loadConfig()
    cfg.Theme = id
    return saveConfig(cfg)
}

// SetVSCodiumExtPath persists the VSCodium extension path to config.
func (a *App) SetVSCodiumExtPath(path string) error {
    cfg := loadConfig()
    cfg.VSCodiumExtPath = path
    return saveConfig(cfg)
}
```

The existing `loadConfig()` (line 55-62) and `saveConfig()` (line 66-76) functions already handle the JSON marshaling. Adding `omitempty` to the new fields ensures backward compatibility -- old config files without these fields will unmarshal cleanly with zero-value strings.

### Go Changes (main.go)

Add import and menu setup. Currently `main.go` does not import the menu package. Add:

```go
import (
    "github.com/wailsapp/wails/v2/pkg/menu"
)
```

Before `wails.Run`, create the menu:
```go
appMenu := menu.NewMenu()
appMenu.Append(menu.EditMenu())
```

In the `options.App` struct, add:
```go
Menu: appMenu,
```

This creates a hidden native macOS Edit menu that provides Cmd+C, Cmd+V, Cmd+X, Cmd+A, Cmd+Z, Cmd+Shift+Z without any visible menu bar (since the window is frameless with `TitleBarHiddenInset`).

### Frontend Changes (App.svelte)

In the `onMount` block (lines 17-29), after checking `GetDevDir()`, also load the theme:

```js
import { GetConfig } from '../wailsjs/go/main/App.js';
import { applyTheme } from './lib/stores/theme.js';

onMount(async () => {
  try {
    const cfg = await GetConfig();
    if (cfg.theme) {
      applyTheme(cfg.theme);
    }
    const dir = await GetDevDir();
    // ... existing logic
  } catch (e) {
    currentView = 'setup';
  }
});
```

Note: `applyTheme` must be called before any themed components mount to avoid a flash of wrong-theme content. Since `onMount` runs before child components mount, this ordering works naturally.

### Technical Considerations

- The `conductorConfig` struct uses `omitempty` on new fields for backward compatibility. Existing `~/.conductor/config.json` files with only `devDir` will parse correctly.
- `GetConfig()` returns the full struct. Wails automatically serializes Go structs to JSON for the JS bridge. The frontend receives `{ devDir: "...", theme: "...", vscodiumExtPath: "..." }`.
- `SetTheme()` and `SetVSCodiumExtPath()` each do a read-modify-write cycle on the config file. There is no concurrency concern because these are only called from UI interactions (single-threaded JS).
- The `menu.EditMenu()` function is from `github.com/wailsapp/wails/v2/pkg/menu`. The `menu` package may already be available -- check `go.mod` to confirm Wails v2 includes it.
- `BackgroundColour` in `main.go` (line 30) remains hardcoded to the dark theme value. This is the window background shown during app startup before the webview loads. Changing it dynamically is not supported by Wails, and the brief flash is acceptable.

### Risks & Edge Cases

- If `config.json` is corrupted or has invalid JSON, `loadConfig()` returns an empty struct (line 60 -- `json.Unmarshal` errors are silently ignored). This is existing behavior and is acceptable.
- If the persisted theme ID doesn't match any theme in `themes.js` (e.g., user hand-edited config), `applyTheme()` silently does nothing and the default dark theme remains.
- The `SetTheme`/`SetVSCodiumExtPath` methods write to disk on every call. For theme switching this is fine since it only happens on user click.
- `menu.EditMenu()` on non-macOS platforms may behave differently. Since the app currently only targets macOS (frameless + TitleBarHiddenInset), this is acceptable.

### Reference Files

- `app.go` -- lines 43-76 (config struct, loadConfig, saveConfig)
- `main.go` -- lines 1-46 (full file, Wails setup)
- `frontend/src/App.svelte` -- lines 17-29 (onMount)
- Wails menu docs: `github.com/wailsapp/wails/v2/pkg/menu`

## Acceptance Criteria

AC-1: Config struct extended with new fields
- Given the `conductorConfig` struct in `app.go`
- When inspecting its fields
- Then it has `Theme string` with json tag `"theme,omitempty"`
- And it has `VSCodiumExtPath string` with json tag `"vscodiumExtPath,omitempty"`

AC-2: GetConfig returns full config to frontend
- Given the config file at `~/.conductor/config.json` contains `{"devDir":"/dev","theme":"conductor-midnight"}`
- When the frontend calls `GetConfig()`
- Then it receives `{ devDir: "/dev", theme: "conductor-midnight", vscodiumExtPath: "" }`

AC-3: SetTheme persists theme ID to disk
- Given the current config has `theme: "conductor-dark"`
- When `SetTheme("conductor-light")` is called
- Then `~/.conductor/config.json` contains `"theme": "conductor-light"`
- And existing `devDir` value is preserved

AC-4: SetVSCodiumExtPath persists path to disk
- Given the current config has no `vscodiumExtPath`
- When `SetVSCodiumExtPath("/Applications/VSCodium.app/Contents/Resources/app/bin/codium")` is called
- Then `~/.conductor/config.json` contains the path under `vscodiumExtPath`
- And existing `devDir` and `theme` values are preserved

AC-5: Persisted theme is loaded on app startup
- Given `~/.conductor/config.json` has `"theme": "conductor-midnight"`
- When the app starts and App.svelte mounts
- Then `applyTheme("conductor-midnight")` is called before content renders
- And the UI displays with midnight theme colors

AC-6: Edit menu enables Cmd+C/V/X in frameless window
- Given the app is running in the frameless window
- When the user presses Cmd+C in a text input or terminal
- Then the copy operation works
- And Cmd+V pastes, Cmd+X cuts, Cmd+Z undoes

AC-7: Backward compatibility with old config files
- Given an existing `config.json` with only `{"devDir":"/dev"}`
- When the app starts and calls `loadConfig()`
- Then `Theme` is empty string and `VSCodiumExtPath` is empty string
- And the app defaults to `conductor-dark` theme

## BDD Test Scenarios

### Scenario 1: Config Struct Serialization

```gherkin
Feature: Config Persistence

  Scenario: Full config round-trip
    Given a conductorConfig with DevDir="/dev", Theme="conductor-light", VSCodiumExtPath="/path"
    When saveConfig is called and then loadConfig is called
    Then the loaded config matches all three fields exactly

  Scenario: Backward compatible deserialization
    Given a config file containing only {"devDir": "/dev"}
    When loadConfig is called
    Then DevDir is "/dev"
    And Theme is ""
    And VSCodiumExtPath is ""

  Scenario: omitempty omits empty fields
    Given a conductorConfig with DevDir="/dev", Theme="", VSCodiumExtPath=""
    When saveConfig is called
    Then the JSON output contains "devDir" but not "theme" or "vscodiumExtPath"
```

### Scenario 2: Wails-Bound Methods

```gherkin
Feature: Config API Methods

  Scenario: SetTheme writes only theme field
    Given config has DevDir="/dev" and Theme="conductor-dark"
    When SetTheme("conductor-midnight") is called
    Then loadConfig returns Theme="conductor-midnight"
    And DevDir is still "/dev"

  Scenario: SetVSCodiumExtPath writes only path field
    Given config has DevDir="/dev" and Theme="conductor-dark"
    When SetVSCodiumExtPath("/usr/local/bin/codium") is called
    Then loadConfig returns VSCodiumExtPath="/usr/local/bin/codium"
    And DevDir and Theme are unchanged

  Scenario: GetConfig returns current state
    Given config file contains all three fields
    When GetConfig is called
    Then the returned struct matches the file contents
```

### Scenario 3: Edit Menu

```gherkin
Feature: Edit Menu for Frameless Window

  Scenario: Hidden edit menu is configured
    Given the Wails app options in main.go
    When the Menu field is inspected
    Then it contains an EditMenu
    And no other menu items are visible in the window chrome

  Scenario: Cmd+C works in text inputs
    Given the app is running with the hidden edit menu
    And the user has selected text in a text input
    When the user presses Cmd+C
    Then the selected text is copied to the clipboard
```

### Scenario 4: Theme Restoration on Startup

```gherkin
Feature: Theme Restoration

  Scenario: Persisted theme applied on mount
    Given config.json has theme "conductor-midnight"
    When App.svelte onMount runs
    Then GetConfig is called
    And applyTheme("conductor-midnight") is called
    And CSS variables on document.documentElement match midnight theme

  Scenario: No persisted theme defaults to dark
    Given config.json has no theme field
    When App.svelte onMount runs
    Then applyTheme is not called
    And the default conductor-dark theme from style.css remains active
```

## Tasks / Subtasks

- [ ] Task 1: Extend conductorConfig and add Go methods (AC: AC-1, AC-2, AC-3, AC-4, AC-7)
  - [ ] Subtask 1a: Add `Theme` and `VSCodiumExtPath` fields with `omitempty` tags to `conductorConfig` struct
  - [ ] Subtask 1b: Add `GetConfig()` method that returns `loadConfig()` result
  - [ ] Subtask 1c: Add `SetTheme(id string) error` method with read-modify-write pattern
  - [ ] Subtask 1d: Add `SetVSCodiumExtPath(path string) error` method with read-modify-write pattern
  - [ ] Subtask 1e: Write Go tests for config round-trip, backward compatibility, and omitempty behavior

- [ ] Task 2: Add hidden Edit menu to main.go (AC: AC-6)
  - [ ] Subtask 2a: Import `github.com/wailsapp/wails/v2/pkg/menu`
  - [ ] Subtask 2b: Create `appMenu := menu.NewMenu()` and append `menu.EditMenu()`
  - [ ] Subtask 2c: Add `Menu: appMenu` to `options.App` struct

- [ ] Task 3: Load persisted theme on App startup (AC: AC-5)
  - [ ] Subtask 3a: Import `GetConfig` from Wails bindings in `App.svelte`
  - [ ] Subtask 3b: Import `applyTheme` from `stores/theme.js`
  - [ ] Subtask 3c: Call `GetConfig()` in `onMount`, apply theme if set, before existing `GetDevDir()` logic

- [ ] Task 4: Write Go unit tests (AC: AC-1, AC-2, AC-3, AC-4, AC-7)
  - [ ] Subtask 4a: Test config serialization round-trip with all fields
  - [ ] Subtask 4b: Test backward compatibility (old config without new fields)
  - [ ] Subtask 4c: Test SetTheme preserves other fields
  - [ ] Subtask 4d: Test SetVSCodiumExtPath preserves other fields

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] Cmd+C/V/X work in text fields and terminal
- [ ] Theme persists across app restart
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
