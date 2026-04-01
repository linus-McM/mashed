# Wails Application Options Reference

Complete reference for the `options.App` struct used in `wails.Run()`.

Import: `github.com/wailsapp/wails/v2/pkg/options`

## options.App Fields

### Window Properties

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Title` | `string` | `""` | Window title |
| `Width` | `int` | `1024` | Initial window width |
| `Height` | `int` | `768` | Initial window height |
| `MinWidth` | `int` | `0` | Minimum window width (0 = no limit) |
| `MinHeight` | `int` | `0` | Minimum window height (0 = no limit) |
| `MaxWidth` | `int` | `0` | Maximum window width (0 = no limit) |
| `MaxHeight` | `int` | `0` | Maximum window height (0 = no limit) |
| `DisableResize` | `bool` | `false` | Prevent window resizing |
| `Fullscreen` | `bool` | `false` | Start in fullscreen mode |
| `Frameless` | `bool` | `false` | Remove window chrome (title bar, borders) |
| `AlwaysOnTop` | `bool` | `false` | Keep window above others |
| `StartHidden` | `bool` | `false` | Start with window hidden |
| `HideWindowOnClose` | `bool` | `false` | Hide instead of quit on close |
| `WindowStartState` | `WindowStartState` | `Normal` | Initial state: `Normal`, `Maximised`, `Minimised`, `Fullscreen` |
| `BackgroundColour` | `*RGBA` | `nil` | Window background color (e.g., `&options.RGBA{R: 27, G: 38, B: 54, A: 255}`) |

### Lifecycle Hooks

| Field | Type | Description |
|-------|------|-------------|
| `OnStartup` | `func(ctx context.Context)` | Called when app starts, resources allocated. Save ctx for runtime calls. |
| `OnDomReady` | `func(ctx context.Context)` | Called when frontend DOM is loaded. Safe for all runtime calls. |
| `OnShutdown` | `func(ctx context.Context)` | Called when app is closing. Cleanup resources. |
| `OnBeforeClose` | `func(ctx context.Context) bool` | Called before close. Return `true` to prevent close. |

### Binding

| Field | Type | Description |
|-------|------|-------------|
| `Bind` | `[]interface{}` | Struct instances whose exported methods become callable from JS |
| `EnumBind` | `[]interface{}` | Enum types to expose to frontend |

### Asset Server

| Field | Type | Description |
|-------|------|-------------|
| `AssetServer` | `*assetserver.Options` | Asset serving configuration (see below) |

#### assetserver.Options

Import: `github.com/wailsapp/wails/v2/pkg/options/assetserver`

| Field | Type | Description |
|-------|------|-------------|
| `Assets` | `fs.FS` | Frontend assets (typically `embed.FS`) |
| `Handler` | `http.Handler` | Fallback handler for assets not found in `Assets` |
| `Middleware` | `func(http.Handler) http.Handler` | HTTP middleware for all asset requests |

### Menu

| Field | Type | Description |
|-------|------|-------------|
| `Menu` | `*menu.Menu` | Application menu (see menus reference) |

### Logging

| Field | Type | Description |
|-------|------|-------------|
| `LogLevel` | `logger.LogLevel` | Dev log level: `TRACE`, `DEBUG`, `INFO`, `WARNING`, `ERROR` |
| `LogLevelProduction` | `logger.LogLevel` | Production log level |
| `Logger` | `logger.Logger` | Custom logger implementation |

### Context Menu

| Field | Type | Description |
|-------|------|-------------|
| `EnableDefaultContextMenu` | `bool` | Enable browser right-click context menu |

### Drag & Drop

| Field | Type | Description |
|-------|------|-------------|
| `DragAndDrop` | `options.DragAndDrop` | Drag and drop configuration |

#### DragAndDrop struct:
```go
type DragAndDrop struct {
    EnableFileDrop     bool   // Enable file drop support
    DisableWebViewDrop bool   // Disable webview's native drop handling
    CSSDropProperty    string // CSS property for drop targets (default: "--wails-drop-target")
    CSSDropValue       string // CSS value to mark drop zone (default: "drop")
}
```

### Frameless Drag

| Field | Type | Description |
|-------|------|-------------|
| `CSSDragProperty` | `string` | CSS property for drag handles (default: `--wails-draggable`) |
| `CSSDragValue` | `string` | CSS value to enable drag (default: `drag`) |

### Single Instance Lock

```go
SingleInstanceLock: &options.SingleInstanceLock{
    UniqueId:               "unique-app-id-uuid",
    OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
        // data.Args - command line args from second instance
        // data.WorkingDirectory - working dir of second instance
    },
},
```

### Error Formatting

| Field | Type | Description |
|-------|------|-------------|
| `ErrorFormatter` | `func(error) any` | Custom error formatting for frontend |

## Platform-Specific Options

### Windows Options

Import: `github.com/wailsapp/wails/v2/pkg/options/windows`

```go
Windows: &windows.Options{
    WebviewIsTransparent:              false,
    WindowIsTranslucent:               false,
    DisableWindowIcon:                 false,
    DisableFramelessWindowDecorations: false,
    WebviewUserDataPath:               "",
    WebviewBrowserPath:                "",
    Theme:                             windows.SystemDefault, // SystemDefault, Dark, Light
    CustomTheme: &windows.ThemeSettings{
        DarkModeTitleBar:   windows.RGB(20, 20, 20),
        DarkModeTitleText:  windows.RGB(200, 200, 200),
        DarkModeBorder:     windows.RGB(20, 0, 20),
        LightModeTitleBar:  windows.RGB(200, 200, 200),
        LightModeTitleText: windows.RGB(20, 20, 20),
        LightModeBorder:    windows.RGB(200, 200, 200),
    },
    Messages:              nil,        // Custom WebView2 installer messages
    OnSuspend:             func() {},
    OnResume:              func() {},
    WebviewGpuIsDisabled:  false,
    EnableSwipeGestures:   false,
    DisablePinchZoom:      false,
    IsZoomControlEnabled:  false,
    ZoomFactor:            1.0,
    WindowClassname:       "",
    OpenInspectorOnStartup: false,
},
```

### Mac Options

Import: `github.com/wailsapp/wails/v2/pkg/options/mac`

```go
Mac: &mac.Options{
    TitleBar: &mac.TitleBar{
        TitlebarAppearsTransparent: false,
        HideTitle:                  false,
        HideTitleBar:               false,
        FullSizeContent:            false,
        UseToolbar:                 false,
        HideToolbarSeparator:       true,
        ToolbarStyle:               mac.ToolbarStyleDefault,
    },
    Appearance:           mac.DefaultAppearance, // or NSAppearanceNameDarkAqua, etc.
    WebviewIsTransparent: false,
    WindowIsTranslucent:  false,
    About: &mac.AboutInfo{
        Title:   "My App",
        Message: "Version 1.0.0",
        Icon:    icon,  // []byte of PNG
    },
    Preferences: &mac.Preferences{
        TabFocusesLinks:        mac.Enabled,
        TextInteractionEnabled: mac.Enabled,
        FullscreenEnabled:      mac.Enabled,
    },
},
```

### Linux Options

Import: `github.com/wailsapp/wails/v2/pkg/options/linux`

```go
Linux: &linux.Options{
    Icon:                icon,       // []byte of window icon
    WindowIsTranslucent: false,
    WebviewGpuPolicy:    linux.WebviewGpuPolicyAlways, // Always, OnDemand, Never
    ProgramName:         "myapp",
},
```

## RGBA Struct

```go
type RGBA struct {
    R, G, B, A uint8
}
```

## WindowStartState Constants

```go
const (
    Normal     WindowStartState = 0
    Maximised  WindowStartState = 1
    Minimised  WindowStartState = 2
    Fullscreen WindowStartState = 3
)
```
