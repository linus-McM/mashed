# Story 1: Native macOS Menu Bar Construction

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Replace the minimal Edit-only native menu in `main.go` with a fully constructed 5-menu menu bar (mashed, File, Edit, View, Help). Each menu item either calls a Wails runtime function directly (Quit, Minimize, Toggle Fullscreen, BrowserOpenURL) or emits a Wails event (`menu:navigate`, `menu:about`) for the frontend to handle. This also adds the `TakeScreenshot()` method to `app.go` for the File > Take Screenshot action.

## Developer Notes

### Architecture
- **Modified file:** `main.go` -- extract menu construction into a `buildMenu(app *App) *menu.Menu` function. Replace the current 2-line `appMenu` block with a call to `buildMenu`.
- **Modified file:** `app.go` -- add two new methods:
  - `TakeScreenshot() (string, error)` -- runs `screencapture -i -x <path>`, saves to `~/Desktop/mashed-screenshot-<timestamp>.png`, emits `screenshot:taken` event with the path, returns the path.
  - `OpenGitHub()` -- calls `runtime.BrowserOpenURL(a.ctx, "https://github.com/user/mashed")`.
- The `buildMenu` function receives `*App` so menu callbacks can access `app.ctx` for `runtime.Quit`, `runtime.EventsEmit`, `runtime.WindowMinimise`, `runtime.WindowToggleMaximise`, `runtime.BrowserOpenURL`, and `app.TakeScreenshot()`.

### Technical Considerations
- **Wails menu API** (from `github.com/wailsapp/wails/v2/pkg/menu`):
  - `menu.NewMenu()` creates the root menu bar.
  - `m.AddSubmenu("label")` returns a `*menu.Menu` for the submenu items.
  - `sub.AddText("label", accelerator, callback)` adds a clickable item.
  - `sub.AddSeparator()` adds a visual separator.
  - `menu.AppMenu()` inserts the macOS-standard app role menu (appears before custom items in the mashed submenu).
  - `menu.EditMenu()` inserts the standard Edit role menu (Undo, Redo, Cut, Copy, Paste, Select All).
- **Accelerators** (from `github.com/wailsapp/wails/v2/pkg/menu/keys`):
  - `keys.CmdOrCtrl("q")` for Cmd+Q
  - `keys.CmdOrCtrl(",")` for Cmd+,
  - `keys.CmdOrCtrl("n")` for Cmd+N
  - `keys.Combo("n", keys.CmdOrCtrlKey, keys.ShiftKey)` for Cmd+Shift+N
  - `keys.CmdOrCtrl("o")` for Cmd+O
  - `keys.Combo("s", keys.CmdOrCtrlKey, keys.ShiftKey)` for Cmd+Shift+S
  - `keys.CmdOrCtrl("1")` for Cmd+1, `keys.CmdOrCtrl("2")` for Cmd+2
  - `keys.CmdOrCtrl("m")` for Cmd+M
  - `keys.Combo("f", keys.ControlKey, keys.CmdOrCtrlKey)` for Ctrl+Cmd+F
- **Callback signature**: `func(cd *menu.CallbackData)` -- the callback type is `menu.Callback` which is `func(*CallbackData)`.
- **Event pattern**: Use `runtime.EventsEmit(app.ctx, "menu:navigate", "settings")` to pass a route string. The frontend listens with `EventsOn('menu:navigate', (route) => { ... })`.
- **Screenshot implementation**: Use `exec.Command("screencapture", "-i", "-x", filepath)` (interactive selection, no sound). Filename: `mashed-screenshot-YYYYMMDD-HHMMSS.png` on the Desktop. After successful capture, emit `screenshot:taken` event with the filepath.
- **Import additions to `main.go`**: Add `"github.com/wailsapp/wails/v2/pkg/menu/keys"` and `"github.com/wailsapp/wails/v2/pkg/runtime"`.
- **Import additions to `app.go`**: Add `"os/exec"` and `"time"` (if not already present).

### Menu Structure (exact code pattern)
```go
func buildMenu(app *App) *menu.Menu {
    appMenu := menu.NewMenu()

    // Menu 1: "mashed" (App Menu)
    mashedMenu := appMenu.AddSubmenu("mashed")
    mashedMenu.AddText("About mashed", nil, func(cd *menu.CallbackData) {
        runtime.EventsEmit(app.ctx, "menu:about")
    })
    mashedMenu.AddSeparator()
    mashedMenu.AddText("Settings...", keys.CmdOrCtrl(","), func(cd *menu.CallbackData) {
        runtime.EventsEmit(app.ctx, "menu:navigate", "settings")
    })
    mashedMenu.AddSeparator()
    mashedMenu.AddText("Quit mashed", keys.CmdOrCtrl("q"), func(cd *menu.CallbackData) {
        runtime.Quit(app.ctx)
    })

    // Menu 2: "File"
    fileMenu := appMenu.AddSubmenu("File")
    // ... items ...

    // Menu 3: "Edit" (preserved)
    appMenu.Append(menu.EditMenu())

    // Menu 4: "View"
    viewMenu := appMenu.AddSubmenu("View")
    // ... items ...

    // Menu 5: "Help"
    helpMenu := appMenu.AddSubmenu("Help")
    // ... items ...

    return appMenu
}
```

### Risks & Edge Cases
- **Cmd+N conflict**: The frontend currently handles Cmd+N in `handleKeydown`. The native menu accelerator takes priority over JavaScript keydown handlers, so the `handleKeydown` Cmd+N block in `App.svelte` should be removed (it will be handled by the menu event in Story 2). However, for this story, the menu callback emits the event -- the frontend handler is Story 2's concern.
- **screencapture blocking**: `screencapture -i` is interactive (user selects area). If the user presses Escape, the command returns exit code 1 and no file is created. Handle this gracefully -- return empty string, no error (user cancelled).
- **screencapture permissions**: macOS requires Screen Recording permission. If denied, `screencapture` fails silently or returns an error. Wrap the error with context.
- **Menu callback context**: Menu callbacks run on the main Go thread. `runtime.EventsEmit` is safe to call from menu callbacks. `runtime.Quit` is also safe.
- **PickDirectory reuse**: File > Open Workspace calls the existing `app.PickDirectory()` method followed by emitting `menu:navigate` with `"feed"` (or a new workspace event). The exact flow: call `PickDirectory()`, if non-empty emit `menu:open-workspace` with the path.

### Reference Files
- `main.go` -- current menu setup (will be refactored)
- `app.go` -- App struct, existing PickDirectory pattern, exec.Command patterns
- `app_git.go` -- exec.Command usage patterns with `exec.CommandContext`
- Wails menu API: `github.com/wailsapp/wails/v2/pkg/menu/` (menu.go, menuitem.go, keys/keys.go, menuroles.go)

## Acceptance Criteria

AC-1: Five menus appear in the macOS menu bar
- Given the application is launched
- When the user looks at the native macOS menu bar
- Then five menus are visible: "mashed", "File", "Edit", "View", "Help"
- And the Edit menu contains standard items (Undo, Redo, Cut, Copy, Paste, Select All)

AC-2: Keyboard shortcuts are registered and functional
- Given the application is running
- When the user presses Cmd+Q
- Then the application quits
- And Cmd+, emits `menu:navigate` with `"settings"`
- And Cmd+N emits `menu:navigate` with `"spawn"`
- And Cmd+Shift+N emits `menu:navigate` with `"new-repo"`
- And Cmd+1 emits `menu:navigate` with `"feed"`
- And Cmd+2 emits `menu:navigate` with `"workflows"`

AC-3: Take Screenshot captures and saves an image
- Given the application is running
- When the user selects File > Take Screenshot (Cmd+Shift+S)
- Then macOS `screencapture` is invoked in interactive mode
- And the screenshot is saved to `~/Desktop/mashed-screenshot-<timestamp>.png`
- And a `screenshot:taken` event is emitted with the file path

AC-4: Window management menu items work
- Given the application is running
- When the user selects View > Minimize (Cmd+M)
- Then the window is minimized via `runtime.WindowMinimise`
- And when the user selects View > Toggle Fullscreen (Ctrl+Cmd+F)
- Then the window toggles fullscreen via `runtime.WindowToggleMaximise`

AC-5: Help menu opens GitHub in default browser
- Given the application is running
- When the user selects Help > mashed on GitHub
- Then the default browser opens the GitHub repository URL via `runtime.BrowserOpenURL`

AC-6: Screenshot cancellation is handled gracefully
- Given the application is running
- When the user triggers Take Screenshot and then presses Escape to cancel
- Then no file is created
- And no error is surfaced to the user
- And the method returns an empty string with nil error

## BDD Test Scenarios

### Scenario 1: Menu construction

```gherkin
Feature: Native Menu Bar

  Scenario: buildMenu returns a complete menu bar
    Given a new App instance
    When buildMenu(app) is called
    Then the returned menu has 5 top-level items
    And item 0 label is "mashed"
    And item 1 label is "File"
    And item 2 has Role EditMenuRole
    And item 3 label is "View"
    And item 4 label is "Help"

  Scenario: mashed submenu has correct items
    Given a new App instance
    When buildMenu(app) is called
    Then the "mashed" submenu has 5 items
    And item 0 is "About mashed" with no accelerator
    And item 1 is a separator
    And item 2 is "Settings..." with CmdOrCtrl+","
    And item 3 is a separator
    And item 4 is "Quit mashed" with CmdOrCtrl+"q"

  Scenario: File submenu has correct items
    Given a new App instance
    When buildMenu(app) is called
    Then the "File" submenu has 7 items
    And item 0 is "New Agent" with CmdOrCtrl+"n"
    And item 1 is "New Repository..." with CmdOrCtrl+Shift+"n"
    And item 2 is a separator
    And item 3 is "Open Workspace..." with CmdOrCtrl+"o"
    And item 4 is a separator
    And item 5 is "Take Screenshot" with CmdOrCtrl+Shift+"s"

  Scenario: View submenu has correct items
    Given a new App instance
    When buildMenu(app) is called
    Then the "View" submenu has 5 items
    And item 0 is "Feed" with CmdOrCtrl+"1"
    And item 1 is "Workflows" with CmdOrCtrl+"2"
    And item 2 is a separator
    And item 3 is "Minimize" with CmdOrCtrl+"m"
    And item 4 is "Toggle Fullscreen" with Ctrl+CmdOrCtrl+"f"

  Scenario: Help submenu has correct items
    Given a new App instance
    When buildMenu(app) is called
    Then the "Help" submenu has 1 item
    And item 0 is "mashed on GitHub" with no accelerator
```

### Scenario 2: TakeScreenshot method

```gherkin
Feature: Screenshot Capture

  Scenario: Successful screenshot capture
    Given the App has a valid context
    And the screencapture command succeeds
    When TakeScreenshot() is called
    Then a file path matching "~/Desktop/mashed-screenshot-*.png" is returned
    And the error is nil

  Scenario: User cancels screenshot
    Given the App has a valid context
    And the screencapture command exits with code 1 (user cancelled)
    When TakeScreenshot() is called
    Then an empty string is returned
    And the error is nil

  Scenario: screencapture binary not found
    Given the App has a valid context
    And the screencapture binary is not available
    When TakeScreenshot() is called
    Then an empty string is returned
    And the error wraps the underlying exec error
    And the error message contains "screencapture"
```

## Tasks / Subtasks

- [ ] Task 1: Extract `buildMenu` function in `main.go` (AC: AC-1, AC-2, AC-4, AC-5)
  - [ ] Subtask 1a: Create `buildMenu(app *App) *menu.Menu` function with all 5 submenus
  - [ ] Subtask 1b: Wire accelerators using `keys.CmdOrCtrl`, `keys.Combo`, etc.
  - [ ] Subtask 1c: Implement callbacks: `runtime.Quit`, `runtime.EventsEmit`, `runtime.WindowMinimise`, `runtime.WindowToggleMaximise`, `runtime.BrowserOpenURL`
  - [ ] Subtask 1d: Replace existing `appMenu` construction in `main()` with `buildMenu(app)` call

- [ ] Task 2: Add `TakeScreenshot()` method to `app.go` (AC: AC-3, AC-6)
  - [ ] Subtask 2a: Implement `TakeScreenshot() (string, error)` using `exec.CommandContext`
  - [ ] Subtask 2b: Generate timestamped filename on Desktop
  - [ ] Subtask 2c: Handle user cancellation (exit code 1) as non-error
  - [ ] Subtask 2d: Emit `screenshot:taken` event on success

- [ ] Task 3: Write unit tests for `buildMenu` in `main_test.go` (AC: AC-1, AC-2, AC-4, AC-5)
  - [ ] Subtask 3a: Test menu structure (5 top-level items, correct labels)
  - [ ] Subtask 3b: Test each submenu has correct items, labels, and accelerators
  - [ ] Subtask 3c: Test separator placement

- [ ] Task 4: Write unit tests for `TakeScreenshot` in `app_test.go` (AC: AC-3, AC-6)
  - [ ] Subtask 4a: Test successful capture path generation (timestamp format)
  - [ ] Subtask 4b: Test cancellation handling (exit code 1 returns empty string, nil error)
  - [ ] Subtask 4c: Test error wrapping when screencapture fails

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
