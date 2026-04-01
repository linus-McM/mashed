# Wails Runtime API Reference

Import: `github.com/wailsapp/wails/v2/pkg/runtime`

All Go methods take `context.Context` as the first parameter. Use the context from `OnStartup` or `OnDomReady`.

JS runtime available via `window.runtime` or generated TypeScript in `wailsjs/runtime/`.

**Important**: Runtime calls in `OnStartup` may fail because the window initializes in a separate thread. Use `OnDomReady` for any calls that need the window.

## Application

### Hide
Hides the application (Mac: app-level hide; Windows/Linux: same as WindowHide).

Go: `Hide(ctx context.Context)`
JS: `Hide()`

### Show
Shows the application.

Go: `Show(ctx context.Context)`
JS: `Show()`

### Quit
Quits the application.

Go: `Quit(ctx context.Context)`
JS: `Quit()`

### Environment
Returns environment details.

Go: `Environment(ctx context.Context) EnvironmentInfo`
JS: `Environment(): Promise<EnvironmentInfo>`

```go
type EnvironmentInfo struct {
    BuildType string  // "dev" or "production"
    Platform  string  // "windows", "darwin", "linux"
    Arch      string  // "amd64", "arm64"
}
```

---

## Window

### WindowSetTitle
Go: `WindowSetTitle(ctx context.Context, title string)`
JS: `WindowSetTitle(title: string)`

### WindowFullscreen
Go: `WindowFullscreen(ctx context.Context)`
JS: `WindowFullscreen()`

### WindowUnfullscreen
Go: `WindowUnfullscreen(ctx context.Context)`
JS: `WindowUnfullscreen()`

### WindowIsFullscreen
Go: `WindowIsFullscreen(ctx context.Context) bool`
JS: `WindowIsFullscreen(): Promise<boolean>`

### WindowCenter
Go: `WindowCenter(ctx context.Context)`
JS: `WindowCenter()`

### WindowReload
Go: `WindowReload(ctx context.Context)`
JS: `WindowReload()`

### WindowReloadApp
Go: `WindowReloadApp(ctx context.Context)`
JS: `WindowReloadApp()`

### WindowSetSystemDefaultTheme
Go: `WindowSetSystemDefaultTheme(ctx context.Context)`

### WindowSetDarkTheme
Go: `WindowSetDarkTheme(ctx context.Context)`

### WindowSetLightTheme
Go: `WindowSetLightTheme(ctx context.Context)`

### WindowShow
Go: `WindowShow(ctx context.Context)`
JS: `WindowShow()`

### WindowHide
Go: `WindowHide(ctx context.Context)`
JS: `WindowHide()`

### WindowIsNormal
Returns true if the window is not minimised, maximised or fullscreen.

Go: `WindowIsNormal(ctx context.Context) bool`
JS: `WindowIsNormal(): Promise<boolean>`

### WindowSetSize
Go: `WindowSetSize(ctx context.Context, width int, height int)`
JS: `WindowSetSize(width: number, height: number)`

### WindowGetSize
Go: `WindowGetSize(ctx context.Context) (width int, height int)`
JS: `WindowGetSize(): Promise<Size>`

### WindowSetMinSize
Setting `0,0` disables the constraint.

Go: `WindowSetMinSize(ctx context.Context, width int, height int)`
JS: `WindowSetMinSize(width: number, height: number)`

### WindowSetMaxSize
Setting `0,0` disables the constraint.

Go: `WindowSetMaxSize(ctx context.Context, width int, height int)`
JS: `WindowSetMaxSize(width: number, height: number)`

### WindowSetAlwaysOnTop
Go: `WindowSetAlwaysOnTop(ctx context.Context, b bool)`
JS: `WindowSetAlwaysOnTop(b: boolean)`

### WindowSetPosition
Position relative to the monitor the window is currently on.

Go: `WindowSetPosition(ctx context.Context, x int, y int)`
JS: `WindowSetPosition(x: number, y: number)`

### WindowGetPosition
Go: `WindowGetPosition(ctx context.Context) (x int, y int)`
JS: `WindowGetPosition(): Promise<Position>`

### WindowMaximise
Go: `WindowMaximise(ctx context.Context)`
JS: `WindowMaximise()`

### WindowUnmaximise
Go: `WindowUnmaximise(ctx context.Context)`
JS: `WindowUnmaximise()`

### WindowIsMaximised
Go: `WindowIsMaximised(ctx context.Context) bool`
JS: `WindowIsMaximised(): Promise<boolean>`

### WindowToggleMaximise
Go: `WindowToggleMaximise(ctx context.Context)`
JS: `WindowToggleMaximise()`

### WindowMinimise
Go: `WindowMinimise(ctx context.Context)`
JS: `WindowMinimise()`

### WindowUnminimise
Go: `WindowUnminimise(ctx context.Context)`
JS: `WindowUnminimise()`

### WindowIsMinimised
Go: `WindowIsMinimised(ctx context.Context) bool`
JS: `WindowIsMinimised(): Promise<boolean>`

### WindowExecJS
Runs JS in the browser asynchronously. Errors appear only in browser console.

Go: `WindowExecJS(ctx context.Context, js string)`

---

## Events

Bidirectional event system between Go and JavaScript. Data passed with events is received in local data types.

### EventsOn
Sets up a persistent listener. Returns a cancel function.

Go: `EventsOn(ctx context.Context, eventName string, callback func(optionalData ...interface{})) func()`
JS: `EventsOn(eventName: string, callback: (optionalData?: any) => void): () => void`

### EventsOff
Removes listener(s) for specified event name(s).

Go: `EventsOff(ctx context.Context, eventName string, additionalEventNames ...string)`
JS: `EventsOff(eventName: string, ...additionalEventNames: string[])`

### EventsOffAll
Removes all event listeners.

Go: `EventsOffAll(ctx context.Context)`
JS: `EventsOffAll()`

### EventsOnce
Listens for an event once, then auto-removes. Returns a cancel function.

Go: `EventsOnce(ctx context.Context, eventName string, callback func(optionalData ...interface{})) func()`
JS: `EventsOnce(eventName: string, callback: (optionalData?: any) => void): () => void`

### EventsOnMultiple
Listens for an event up to `counter` times. Returns a cancel function.

Go: `EventsOnMultiple(ctx context.Context, eventName string, callback func(optionalData ...interface{}), counter int) func()`
JS: `EventsOnMultiple(eventName: string, callback: (optionalData?: any) => void, counter: number): () => void`

### EventsEmit
Emits an event with optional data.

Go: `EventsEmit(ctx context.Context, eventName string, optionalData ...interface{})`
JS: `EventsEmit(eventName: string, ...optionalData: any[])`

---

## Dialogs

### OpenDirectoryDialog
Go: `OpenDirectoryDialog(ctx context.Context, dialogOptions OpenDialogOptions) (string, error)`

Returns: selected directory or empty string if cancelled.

### OpenFileDialog
Go: `OpenFileDialog(ctx context.Context, dialogOptions OpenDialogOptions) (string, error)`

Returns: selected file path or empty string if cancelled.

### OpenMultipleFilesDialog
Go: `OpenMultipleFilesDialog(ctx context.Context, dialogOptions OpenDialogOptions) ([]string, error)`

Returns: selected file paths or nil if cancelled.

### SaveFileDialog
Go: `SaveFileDialog(ctx context.Context, dialogOptions SaveDialogOptions) (string, error)`

Returns: selected save path or empty string if cancelled.

### MessageDialog
Go: `MessageDialog(ctx context.Context, dialogOptions MessageDialogOptions) (string, error)`

Returns: text of the selected button.

### OpenDialogOptions
```go
type OpenDialogOptions struct {
    DefaultDirectory           string
    DefaultFilename            string
    Title                      string
    Filters                    []FileFilter
    ShowHiddenFiles            bool
    CanCreateDirectories       bool
    ResolvesAliases            bool
    TreatPackagesAsDirectories bool  // macOS only
}
```

### SaveDialogOptions
```go
type SaveDialogOptions struct {
    DefaultDirectory           string
    DefaultFilename            string
    Title                      string
    Filters                    []FileFilter
    ShowHiddenFiles            bool
    CanCreateDirectories       bool
    TreatPackagesAsDirectories bool  // macOS only
}
```

### MessageDialogOptions
```go
type MessageDialogOptions struct {
    Type          DialogType  // InfoDialog, WarningDialog, ErrorDialog, QuestionDialog
    Title         string
    Message       string
    Buttons       []string
    DefaultButton string
    CancelButton  string
    Icon          []byte      // PNG icon data
}
```

### FileFilter
```go
type FileFilter struct {
    DisplayName string  // e.g., "Images (*.jpg, *.png)"
    Pattern     string  // e.g., "*.jpg;*.png"
}
```

### DialogType Constants
```go
const (
    InfoDialog     DialogType = "info"
    WarningDialog  DialogType = "warning"
    ErrorDialog    DialogType = "error"
    QuestionDialog DialogType = "question"
)
```

---

## Menu

### MenuSetApplicationMenu
Go: `MenuSetApplicationMenu(ctx context.Context, menu *menu.Menu)`

### MenuUpdateApplicationMenu
Call after modifying the menu struct to refresh the UI.

Go: `MenuUpdateApplicationMenu(ctx context.Context)`

### Menu Struct
Import: `github.com/wailsapp/wails/v2/pkg/menu`

```go
type Menu struct {
    Items []*MenuItem
}

func NewMenu() *Menu
```

### MenuItem Types
```go
const (
    TextType      Type = "Text"
    SeparatorType Type = "Separator"
    SubmenuType   Type = "Submenu"
    CheckboxType  Type = "Checkbox"
    RadioType     Type = "Radio"
)
```

### Menu Helper Functions
```go
func Text(label string, accelerator *keys.Accelerator, click Callback) *MenuItem
func Separator() *MenuItem
func Radio(label string, selected bool, accelerator *keys.Accelerator, click Callback) *MenuItem
func Checkbox(label string, checked bool, accelerator *keys.Accelerator, click Callback) *MenuItem
func SubMenu(label string, menu *Menu) *Menu
```

### Menu Builder Methods
```go
menu.AddSubmenu(label string) *Menu
menu.AddText(label string, accelerator *keys.Accelerator, callback Callback) *MenuItem
menu.AddRadio(label string, selected bool, accelerator *keys.Accelerator, callback Callback) *MenuItem
menu.AddCheckbox(label string, checked bool, accelerator *keys.Accelerator, callback Callback) *MenuItem
menu.AddSeparator()
menu.Append(item *MenuItem)
```

### Keyboard Accelerators
Import: `github.com/wailsapp/wails/v2/pkg/menu/keys`

```go
keys.CmdOrCtrl("s")          // Cmd+S on Mac, Ctrl+S on Windows/Linux
keys.OptionOrAlt("f")         // Option+F on Mac, Alt+F on Windows/Linux
keys.Shift("n")               // Shift+N
keys.Combo("k", keys.CmdOrCtrlKey, keys.ShiftKey)  // Cmd+Shift+K
```

### Pre-built Menus
```go
menu.EditMenu()  // Standard Edit menu (Copy, Paste, etc.) — always add on macOS
```

### Callback
```go
type Callback func(callbackData *CallbackData)

type CallbackData struct {
    MenuItem *MenuItem
}
```

---

## Clipboard

### ClipboardGetText
Go: `ClipboardGetText(ctx context.Context) (string, error)`
JS: `ClipboardGetText(): Promise<string>`

### ClipboardSetText
Go: `ClipboardSetText(ctx context.Context, text string) error`
JS: `ClipboardSetText(text: string): Promise<boolean>`

---

## Screen

### ScreenGetAll
Returns all connected screens.

Go: `ScreenGetAll(ctx context.Context) []Screen`
JS: `ScreenGetAll(): Screen[]`

```go
type Screen struct {
    IsCurrent bool
    IsPrimary bool
    Width     int
    Height    int
}
```

---

## Browser

### BrowserOpenURL
Opens URL in system default browser.

Go: `BrowserOpenURL(ctx context.Context, url string)`
JS: `BrowserOpenURL(url: string)`

---

## Log

Import log levels: `github.com/wailsapp/wails/v2/pkg/logger`

### Log Functions
```
LogPrint(ctx, message string)
LogTrace(ctx, message string)     // Level 1
LogDebug(ctx, message string)     // Level 2
LogInfo(ctx, message string)      // Level 3
LogWarning(ctx, message string)   // Level 4
LogError(ctx, message string)     // Level 5
LogFatal(ctx, message string)     // Level 6
```

All available in both Go and JS.

### LogSetLogLevel
Go: `LogSetLogLevel(ctx context.Context, level logger.LogLevel)`
JS: `LogSetLogLevel(level: number)`

### Custom Logger
Implement the `logger.Logger` interface:

```go
type Logger interface {
    Print(message string)
    Trace(message string)
    Debug(message string)
    Info(message string)
    Warning(message string)
    Error(message string)
    Fatal(message string)
}
```

---

## Linux-Specific

### ResetSignalHandlers
On Linux, WebKit may install signal handlers that prevent Go panic recovery. Call this immediately before code that might panic.

Go: `ResetSignalHandlers()`

```go
go func() {
    defer func() {
        if err := recover(); err != nil {
            log.Printf("Recovered: %v", err)
        }
    }()
    runtime.ResetSignalHandlers()
    // potentially dangerous code...
}()
```

Must be called in each goroutine. No-op on macOS and Windows.
