---
name: wails
description: >
  Comprehensive guide for building desktop applications with the Wails framework (Go backend + web frontend).
  Use this skill whenever the user is building, configuring, debugging, or deploying a Wails application.
  Triggers include: creating a new Wails project, configuring wails.json, using the Wails runtime API (window,
  dialog, events, menu, clipboard, screen), binding Go methods to the frontend, building for production,
  frameless windows, drag-and-drop, file associations, notifications, code signing, cross-platform builds,
  NSIS installers, Mac App Store submission, single-instance locks, custom asset servers, or any question
  about the wails CLI (init, build, dev, doctor, generate). Also use when you see imports from
  "github.com/wailsapp/wails/v2" or the user mentions "wails" in the context of desktop apps.
---

# Wails Development Guide

Wails lets you build desktop applications using Go for the backend and any web technology (React, Vue, Svelte, etc.) for the frontend. The Go backend and web frontend communicate through bindings and a unified event system. The result is a native desktop app with a webview-based UI.

## Table of Contents

1. [Installation & Setup](#installation--setup)
2. [Project Structure](#project-structure)
3. [Application Lifecycle](#application-lifecycle)
4. [Binding Go Methods](#binding-go-methods)
5. [Runtime API Overview](#runtime-api-overview)
6. [Events System](#events-system)
7. [Window Management](#window-management)
8. [Dialogs](#dialogs)
9. [Menus](#menus)
10. [Asset Server & Dynamic Assets](#asset-server--dynamic-assets)
11. [Frameless Windows & Drag](#frameless-windows--drag)
12. [CLI Reference](#cli-reference)
13. [Project Configuration (wails.json)](#project-configuration-wailsjson)
14. [Application Options](#application-options)
15. [Building & Distribution](#building--distribution)
16. [Platform-Specific Guides](#platform-specific-guides)

For detailed API reference, see the files in `references/`:
- `references/options.md` — Full `options.App` struct reference
- `references/runtime-api.md` — Complete runtime API (window, dialog, events, menu, clipboard, screen, log, browser)
- `references/cli.md` — CLI commands and flags
- `references/guides.md` — Platform guides, signing, installers, advanced patterns

---

## Installation & Setup

### Prerequisites

- Go 1.18+
- Node.js 15+ (for frontend tooling)
- Platform-specific dependencies:
  - **macOS**: Xcode command line tools (`xcode-select --install`)
  - **Linux**: `gcc`, `libgtk-3-dev`, `libwebkit2gtk-4.0-dev` (or `4.1` on newer distros)
  - **Windows**: WebView2 runtime (usually pre-installed on Windows 10/11)

### Install CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify installation:
```bash
wails doctor
```

### Create a New Project

```bash
wails init -n myapp -t vue       # Options: vue, react, svelte, preact, lit, vanilla
wails init -n myapp -t vue-ts    # TypeScript variants available
```

Remote templates from GitHub are also supported:
```bash
wails init -n myapp -t https://github.com/user/wails-template
```

---

## Project Structure

A standard Wails project:

```
myapp/
├── build/                    # Build assets (icons, manifests, installer configs)
│   ├── appicon.png          # Application icon
│   ├── darwin/              # macOS-specific (Info.plist)
│   └── windows/             # Windows-specific (manifest, NSIS installer)
├── frontend/                 # Web frontend (React/Vue/Svelte/etc.)
│   ├── src/
│   ├── package.json
│   ├── index.html
│   └── wailsjs/             # Auto-generated JS bindings (created by wails dev/build)
│       ├── go/              # Generated Go method wrappers
│       └── runtime/         # Runtime TypeScript declarations
├── app.go                    # Application logic struct + methods
├── main.go                   # Entry point — calls wails.Run()
├── wails.json                # Project configuration
└── go.mod
```

The `wailsjs/` directory is auto-generated. It contains:
- `go/main/App.js` — JavaScript wrappers for bound Go methods
- `runtime/runtime.d.ts` — TypeScript declarations for the Wails runtime

---

## Application Lifecycle

### main.go — Entry Point

```go
package main

import (
    "embed"
    "github.com/wailsapp/wails/v2"
    "github.com/wailsapp/wails/v2/pkg/options"
    "github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
    app := NewApp()

    err := wails.Run(&options.App{
        Title:  "My App",
        Width:  1024,
        Height: 768,
        AssetServer: &assetserver.Options{
            Assets: assets,
        },
        OnStartup:  app.startup,
        OnShutdown: app.shutdown,
        Bind: []interface{}{
            app,
        },
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

### app.go — Application Logic

```go
package main

import "context"

type App struct {
    ctx context.Context
}

func NewApp() *App {
    return &App{}
}

// startup is called when the app starts. Save the context for runtime calls.
func (a *App) startup(ctx context.Context) {
    a.ctx = ctx
}

// shutdown is called when the app is closing.
func (a *App) shutdown(ctx context.Context) {
    // cleanup resources
}

// Greet is exposed to the frontend via binding
func (a *App) Greet(name string) string {
    return "Hello " + name + "!"
}
```

### Lifecycle Hooks

| Hook | When Called | Use For |
|------|-----------|---------|
| `OnStartup` | App starts, resources allocated | Initialize services, save context |
| `OnDomReady` | Frontend DOM loaded | Runtime calls that need the window |
| `OnShutdown` | App closing | Cleanup resources |
| `OnBeforeClose` | Before window closes | Confirm exit dialogs (return `true` to prevent close) |

**Important**: The runtime may not be fully available in `OnStartup` because the window initializes in a separate thread. Use `OnDomReady` for runtime calls that need the window (like `WindowSetTitle`).

---

## Binding Go Methods

Any **exported method** on a struct passed to `Bind` becomes callable from JavaScript.

### Go Side

```go
// In options.App:
Bind: []interface{}{
    app,           // all exported methods on *App become available
    &myService,    // can bind multiple structs
},
```

### Frontend Side

```javascript
// Auto-generated in frontend/wailsjs/go/main/App.js
import { Greet } from '../wailsjs/go/main/App';

const result = await Greet("World");
// result: "Hello World!"
```

- Methods must be **exported** (capitalized)
- Supported parameter types: basic types, structs, slices, maps
- Structs are passed as JSON objects
- Methods can return `(value, error)` — errors become rejected promises in JS
- Bound methods appear at `window.go.main.App.MethodName()` (or via the generated imports)

---

## Runtime API Overview

The Go runtime is imported from `github.com/wailsapp/wails/v2/pkg/runtime`. Every method takes `context.Context` as the first parameter (use the context saved in `OnStartup`).

The JS runtime is available via `window.runtime` or the generated TypeScript declarations in `wailsjs/runtime/`.

Key API areas:
- **Window** — size, position, title, minimize, maximize, fullscreen, show/hide
- **Dialog** — open file, save file, message dialogs
- **Events** — emit/listen for events between Go and JS
- **Menu** — application menus with shortcuts
- **Clipboard** — get/set text
- **Screen** — monitor information
- **Log** — structured logging (Trace, Debug, Info, Warning, Error, Fatal)
- **Browser** — open URLs in system browser

See `references/runtime-api.md` for the complete API reference.

---

## Events System

Events flow bidirectionally between Go and JavaScript.

### Go Side

```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

// Listen for an event
runtime.EventsOn(a.ctx, "myevent", func(data ...interface{}) {
    fmt.Println("Received:", data)
})

// Emit an event
runtime.EventsEmit(a.ctx, "backend-update", "some data")

// Listen once
runtime.EventsOnce(a.ctx, "init", func(data ...interface{}) { /* ... */ })

// Listen N times
runtime.EventsOnMultiple(a.ctx, "tick", func(data ...interface{}) { /* ... */ }, 5)

// Remove listeners
runtime.EventsOff(a.ctx, "myevent")
runtime.EventsOffAll(a.ctx)
```

### JavaScript Side

```javascript
import { EventsOn, EventsEmit, EventsOff } from '../wailsjs/runtime/runtime';

// Listen
const cancel = EventsOn("backend-update", (data) => {
    console.log("Got update:", data);
});

// Emit (received by Go listeners)
EventsEmit("myevent", { key: "value" });

// Cleanup
cancel();  // or EventsOff("backend-update")
```

---

## Window Management

```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

runtime.WindowSetTitle(a.ctx, "New Title")
runtime.WindowSetSize(a.ctx, 1200, 800)
runtime.WindowSetMinSize(a.ctx, 400, 300)
runtime.WindowSetMaxSize(a.ctx, 1920, 1080)
runtime.WindowSetPosition(a.ctx, 100, 100)
runtime.WindowCenter(a.ctx)
runtime.WindowMaximise(a.ctx)
runtime.WindowMinimise(a.ctx)
runtime.WindowFullscreen(a.ctx)
runtime.WindowUnfullscreen(a.ctx)
runtime.WindowShow(a.ctx)
runtime.WindowHide(a.ctx)
runtime.WindowSetAlwaysOnTop(a.ctx, true)
runtime.WindowSetDarkTheme(a.ctx)
runtime.WindowSetLightTheme(a.ctx)
runtime.WindowSetSystemDefaultTheme(a.ctx)
runtime.WindowReload(a.ctx)
runtime.WindowReloadApp(a.ctx)

// Query state
w, h := runtime.WindowGetSize(a.ctx)
x, y := runtime.WindowGetPosition(a.ctx)
isMax := runtime.WindowIsMaximised(a.ctx)
isMin := runtime.WindowIsMinimised(a.ctx)
isFull := runtime.WindowIsFullscreen(a.ctx)
isNormal := runtime.WindowIsNormal(a.ctx)
```

All window methods are available in JS too via `window.runtime.*` or the runtime imports.

---

## Dialogs

```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

// Open file
file, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
    Title:            "Select a file",
    DefaultDirectory: "/home/user",
    Filters: []runtime.FileFilter{
        {DisplayName: "Images", Pattern: "*.png;*.jpg;*.jpeg"},
        {DisplayName: "All Files", Pattern: "*.*"},
    },
})

// Open multiple files
files, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{...})

// Open directory
dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
    Title: "Select folder",
})

// Save file
path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
    Title:           "Save as...",
    DefaultFilename: "output.txt",
})

// Message dialog
result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
    Type:    runtime.QuestionDialog,   // InfoDialog, WarningDialog, ErrorDialog
    Title:   "Confirm",
    Message: "Are you sure?",
    Buttons: []string{"Yes", "No"},
    DefaultButton: "No",
})
```

---

## Menus

```go
import (
    "github.com/wailsapp/wails/v2/pkg/menu"
    "github.com/wailsapp/wails/v2/pkg/menu/keys"
    rt "github.com/wailsapp/wails/v2/pkg/runtime"
)

AppMenu := menu.NewMenu()
FileMenu := AppMenu.AddSubmenu("File")
FileMenu.AddText("Open", keys.CmdOrCtrl("o"), func(cb *menu.CallbackData) {
    // handle open
})
FileMenu.AddSeparator()
FileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(cb *menu.CallbackData) {
    rt.Quit(app.ctx)
})

// On macOS, add EditMenu for Cmd+C/V/Z shortcuts
if runtime.GOOS == "darwin" {
    AppMenu.Append(menu.EditMenu())
}

// Set in options
err := wails.Run(&options.App{
    Menu: AppMenu,
    // ...
})

// Or set dynamically
rt.MenuSetApplicationMenu(a.ctx, newMenu)
rt.MenuUpdateApplicationMenu(a.ctx)
```

Menu item types: `Text`, `Separator`, `Checkbox`, `Radio`, `Submenu`.

---

## Asset Server & Dynamic Assets

### Embedded Assets (Default)

```go
//go:embed all:frontend/dist
var assets embed.FS

// In options:
AssetServer: &assetserver.Options{
    Assets: assets,
},
```

### Custom Asset Handler

For dynamic assets not in the embedded filesystem, implement `http.Handler`:

```go
type FileLoader struct {
    http.Handler
}

func NewFileLoader() *FileLoader {
    return &FileLoader{}
}

func (h *FileLoader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    requestedFile := r.URL.Path
    fileData, err := os.ReadFile(requestedFile)
    if err != nil {
        w.WriteHeader(http.StatusNotFound)
        return
    }
    w.Write(fileData)
}

// In options:
AssetServer: &assetserver.Options{
    Assets:  assets,
    Handler: NewFileLoader(),  // Fallback when asset not found in embed.FS
},
```

### Middleware

```go
AssetServer: &assetserver.Options{
    Assets:     assets,
    Middleware: MyMiddleware,  // func(http.Handler) http.Handler
},
```

---

## Frameless Windows & Drag

Enable frameless mode:
```go
err := wails.Run(&options.App{
    Frameless: true,
    // ...
})
```

### CSS-based Drag Regions

Use the CSS property `--wails-draggable` to define drag handles:

```html
<body style="--wails-draggable:drag">
    <div id="toolbar" style="--wails-draggable:drag">Drag me</div>
    <div id="content" style="--wails-draggable:no-drag">
        <input type="text" />  <!-- inputs won't drag -->
    </div>
</body>
```

Alternative: use the `data-wails-drag` and `data-wails-no-drag` HTML attributes.

### Drag & Drop Files

Use the CSS property `--wails-drop-target` to define drop zones:

```html
<div style="--wails-drop-target:drop">Drop files here</div>
<div style="--wails-drop-target:no-drop">No dropping here</div>
```

---

## CLI Reference

| Command | Description |
|---------|------------|
| `wails init -n <name> -t <template>` | Create new project |
| `wails build` | Build production binary |
| `wails dev` | Run in development mode (hot reload) |
| `wails doctor` | Check system dependencies |
| `wails generate module` | Generate `wailsjs` bindings |
| `wails generate template` | Create a template from project |
| `wails update` | Update Wails CLI |
| `wails version` | Print version |

### Key Build Flags

| Flag | Description |
|------|------------|
| `-platform <os/arch>` | Target platform (e.g., `windows/amd64`, `darwin/universal`) |
| `-ldflags <flags>` | Linker flags |
| `-o <filename>` | Output filename |
| `-clean` | Clean build directory first |
| `-nsis` | Generate NSIS installer (Windows) |
| `-webview2 <strategy>` | WebView2 install strategy: `download`, `embed`, `browser`, `error` |
| `-upx` | Compress binary with UPX |
| `-race` | Build with race detector |
| `-s` | Skip frontend build |
| `-tags <tags>` | Build tags |
| `-trimpath` | Remove file system paths from binary |
| `-skipbindings` | Skip bindings generation |

### Key Dev Flags

| Flag | Description |
|------|------------|
| `-browser` | Open browser on start |
| `-e <extensions>` | File extensions to watch (default: `.go`) |
| `-reloaddirs <dirs>` | Additional directories to watch |
| `-compiler <path>` | Custom Go compiler |
| `-loglevel <level>` | Log level: `Trace`, `Debug`, `Info`, `Warning`, `Error` |
| `-assetdir <dir>` | Serve assets from directory instead of embed.FS |
| `-frontenddevserverurl <url>` | Use external dev server (e.g., Vite) |

See `references/cli.md` for the complete flag reference.

---

## Project Configuration (wails.json)

```json
{
  "name": "myapp",
  "outputfilename": "myapp",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "wailsjsdir": "./frontend",
  "assetdir": "./frontend/dist",
  "debounceMS": 100,
  "devServer": "localhost:34115",
  "info": {
    "companyName": "My Company",
    "productName": "My App",
    "productVersion": "1.0.0",
    "copyright": "Copyright 2024",
    "comments": "Built with Wails"
  }
}
```

Key fields:
- `frontend:install` — Command to install frontend dependencies (e.g., `npm install`)
- `frontend:build` — Command to build frontend assets (e.g., `npm run build`)
- `frontend:dev:watcher` — Command for frontend dev server in `wails dev`
- `frontend:dev:serverUrl` — URL of frontend dev server (`auto` for Vite auto-detection)
- `wailsjsdir` — Where to put generated JS bindings
- `assetdir` — Directory of compiled frontend assets
- `author` — Author information

See `references/guides.md` for the full schema.

---

## Application Options

The `options.App` struct configures everything. Key fields:

```go
&options.App{
    Title:             "My App",
    Width:             1024,
    Height:            768,
    MinWidth:          400,
    MinHeight:         300,
    MaxWidth:          1920,
    MaxHeight:         1080,
    Frameless:         false,
    AlwaysOnTop:       false,
    DisableResize:     false,
    Fullscreen:        false,
    StartHidden:       false,
    HideWindowOnClose: false,
    BackgroundColour:  &options.RGBA{R: 255, G: 255, B: 255, A: 255},
    WindowStartState:  options.Normal,  // Normal, Maximised, Minimised, Fullscreen
    
    // Asset serving
    AssetServer: &assetserver.Options{
        Assets:     assets,
        Handler:    customHandler,
        Middleware: middleware,
    },
    
    // Lifecycle hooks
    OnStartup:     app.startup,
    OnDomReady:    app.domReady,
    OnShutdown:    app.shutdown,
    OnBeforeClose: app.beforeClose,
    
    // Binding
    Bind: []interface{}{app},
    
    // Menu
    Menu: appMenu,
    
    // Logging
    LogLevel:           logger.DEBUG,
    LogLevelProduction: logger.ERROR,
    Logger:             customLogger,
    
    // Drag & drop
    EnableDefaultContextMenu: true,
    DragAndDrop: options.DragAndDrop{
        EnableFileDrop:     true,
        DisableWebViewDrop: false,
        CSSDropProperty:    "--wails-drop-target",
        CSSDropValue:       "drop",
    },
    
    // Single instance
    SingleInstanceLock: &options.SingleInstanceLock{
        UniqueId:               "unique-app-id",
        OnSecondInstanceLaunch: app.onSecondInstance,
    },
    
    // Platform-specific options
    Windows: &windows.Options{...},
    Mac:     &mac.Options{...},
    Linux:   &linux.Options{...},
}
```

See `references/options.md` for the complete struct reference.

---

## Building & Distribution

### Production Build

```bash
wails build                          # Build for current platform
wails build -platform windows/amd64  # Cross-compile
wails build -platform darwin/universal  # macOS universal binary
wails build -nsis                    # Windows NSIS installer
wails build -upx                    # Compress with UPX
wails build -ldflags "-s -w"        # Strip debug info
wails build -trimpath               # Remove file paths
wails build -obfuscated -garbleargs "-literals"  # Obfuscation
```

### Supported Platforms

| Target | Value |
|--------|-------|
| macOS (Intel) | `darwin/amd64` |
| macOS (Apple Silicon) | `darwin/arm64` |
| macOS (Universal) | `darwin/universal` |
| Windows (64-bit) | `windows/amd64` |
| Windows (ARM) | `windows/arm64` |
| Linux (64-bit) | `linux/amd64` |
| Linux (ARM) | `linux/arm64` |

### Cross-Platform CI Build

Use GitHub Actions matrix strategy. See `references/guides.md` for complete CI/CD examples.

---

## Platform-Specific Guides

Read `references/guides.md` for detailed coverage of:

- **Windows**: NSIS installers, WebView2 strategies, manifest configuration
- **macOS**: Code signing, notarization, Mac App Store submission, universal binaries
- **Linux**: Dependencies per distro, WebkitGTK versions, video tag workarounds
- **Frameless**: CSS drag handles, title bar customization
- **File Associations**: Register your app to open specific file types
- **Notifications**: Cross-platform native notifications with action buttons
- **Single Instance Lock**: Prevent multiple instances, pass data between them
- **Custom Protocol Schemes**: Register `myapp://` URL schemes
- **Routing**: SPA routing with React Router, Vue Router, Svelte SPA Router
- **Obfuscation**: Using garble for binary obfuscation
- **Code Signing**: Windows and macOS signing workflows
