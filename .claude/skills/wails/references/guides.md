# Wails Advanced Guides

## Table of Contents
1. [Project Configuration (wails.json)](#project-configuration)
2. [Frontend Integration](#frontend-integration)
3. [Frameless Windows](#frameless-windows)
4. [Drag & Drop](#drag--drop)
5. [File Associations](#file-associations)
6. [Notifications](#notifications)
7. [Single Instance Lock](#single-instance-lock)
8. [Custom Protocol Schemes](#custom-protocol-schemes)
9. [Routing (SPA)](#routing)
10. [Windows Installer (NSIS)](#windows-installer-nsis)
11. [Code Signing](#code-signing)
12. [Mac App Store](#mac-app-store)
13. [Cross-Platform CI/CD](#cross-platform-cicd)
14. [Obfuscation](#obfuscation)
15. [Manual Builds](#manual-builds)
16. [Troubleshooting](#troubleshooting)

---

## Project Configuration

### wails.json Full Schema

```json
{
  "name": "myapp",
  "outputfilename": "myapp",
  "assetdir": "./frontend/dist",
  "reloaddirs": "",
  "build:dir": "build",
  "build:tags": "",
  "frontend:dir": "frontend",
  "frontend:install": "npm install",
  "frontend:build": "npm run build",
  "frontend:dev": "",
  "frontend:dev:build": "",
  "frontend:dev:install": "",
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto",
  "viteServerTimeout": 10,
  "wailsjsdir": "./frontend",
  "debounceMS": 100,
  "devServer": "localhost:34115",
  "appargs": "",
  "runNonNativeBuildHooks": false,
  "preBuildHooks": {
    "GOOS/GOARCH": "",
    "GOOS/*": "",
    "*/*": ""
  },
  "postBuildHooks": {
    "GOOS/GOARCH": "",
    "GOOS/*": "",
    "*/*": ""
  },
  "nsisType": "",
  "obfuscated": "",
  "garbleargs": "",
  "bindings": {
    "ts_generation": {
      "prefix": "",
      "suffix": "",
      "outputType": "classes"
    }
  },
  "info": {
    "companyName": "My Company",
    "productName": "My App",
    "productVersion": "1.0.0",
    "copyright": "Copyright 2024",
    "comments": "Built with Wails"
  }
}
```

Key notes:
- `frontend:dev:serverUrl: "auto"` — auto-detects Vite dev server URL
- `preBuildHooks` / `postBuildHooks` — `${platform}` is replaced with `GOOS/GOARCH`, `${bin}` with binary path
- `bindings.ts_generation.outputType` — `"classes"` or `"interfaces"`

---

## Frontend Integration

Wails works with any frontend framework. The build process:

1. `frontend:install` runs (e.g., `npm install`)
2. `frontend:build` runs (e.g., `npm run build`)
3. Built assets in `assetdir` are embedded via `//go:embed`

### Calling Go from JavaScript

After binding a struct, auto-generated wrappers appear in `frontend/wailsjs/go/`:

```javascript
// Import generated bindings
import { Greet } from '../wailsjs/go/main/App';

// Call Go method (returns Promise)
const result = await Greet("World");
```

### Calling JavaScript from Go

```go
runtime.WindowExecJS(a.ctx, `document.title = "Updated"`)
```

### Using Vite Dev Server

In `wails.json`:
```json
{
  "frontend:dev:watcher": "npm run dev",
  "frontend:dev:serverUrl": "auto"
}
```

This starts the Vite dev server alongside `wails dev` for HMR.

---

## Frameless Windows

Remove native window chrome:

```go
&options.App{
    Frameless: true,
}
```

### CSS Drag Handles

```css
/* Make the entire body draggable */
body { --wails-draggable: drag; }

/* Exclude interactive elements */
.content { --wails-draggable: no-drag; }
```

Or use HTML attributes: `data-wails-drag` and `data-wails-no-drag`.

### macOS Title Bar Customization

```go
Mac: &mac.Options{
    TitleBar: &mac.TitleBar{
        TitlebarAppearsTransparent: true,
        HideTitle:                  true,
        FullSizeContent:            true,
        UseToolbar:                 true,
        HideToolbarSeparator:       true,
    },
},
```

---

## Drag & Drop

Enable file drop:

```go
&options.App{
    DragAndDrop: options.DragAndDrop{
        EnableFileDrop:     true,
        DisableWebViewDrop: false,
        CSSDropProperty:    "--wails-drop-target",
        CSSDropValue:       "drop",
    },
}
```

HTML:
```html
<div style="--wails-drop-target:drop">Drop files here</div>
```

Listen for drops in JS:
```javascript
runtime.EventsOn("wails:file-drop", (x, y, paths) => {
    console.log("Files dropped:", paths);
});
```

---

## File Associations

Register your app to handle specific file types.

1. Configure in `wails.json`:
```json
{
  "info": {
    "fileAssociations": [
      {
        "ext": "myext",
        "name": "My File Type",
        "description": "My application file",
        "iconName": "myicon",
        "role": "Editor"
      }
    ]
  }
}
```

2. Handle in Go:
```go
OnStartup: func(ctx context.Context) {
    // Check os.Args for file path passed by OS
    if len(os.Args) > 1 {
        filePath := os.Args[1]
        // Open the file
    }
},
```

---

## Notifications

Cross-platform native notifications:

```go
import "github.com/wailsapp/wails/v2/pkg/runtime"

runtime.SendNotification(a.ctx, runtime.NotificationOptions{
    Title:   "Hello",
    Body:    "This is a notification",
})
```

Supports action buttons and text input fields. Currently unsupported in JS runtime.

---

## Single Instance Lock

Prevent multiple instances and receive data from second instance attempts:

```go
&options.App{
    SingleInstanceLock: &options.SingleInstanceLock{
        UniqueId: "unique-uuid-for-myapp",
        OnSecondInstanceLaunch: func(data options.SecondInstanceData) {
            println("Second instance args:", data.Args)
            println("Second instance dir:", data.WorkingDirectory)
            // Bring window to front
            runtime.WindowUnminimise(ctx)
            runtime.Show(ctx)
        },
    },
}
```

Implementation: named mutex on Windows/macOS, D-Bus on Linux.

Treat data from second instance as untrusted — validate args.

---

## Custom Protocol Schemes

Register custom URL schemes (e.g., `myapp://action`):

```go
&options.App{
    // Handle via AssetServer or custom protocol registration
}
```

Platform-specific registration required in manifests/plists.

---

## Routing

### React (react-router)

```javascript
import { HashRouter, Route, Routes } from 'react-router-dom';

function App() {
    return (
        <HashRouter>
            <Routes>
                <Route path="/" element={<Home />} />
                <Route path="/settings" element={<Settings />} />
            </Routes>
        </HashRouter>
    );
}
```

Use `HashRouter` (not `BrowserRouter`) because Wails serves from embedded filesystem.

### Vue (vue-router)

```javascript
import { createRouter, createWebHashHistory } from 'vue-router';

const router = createRouter({
    history: createWebHashHistory(),
    routes: [
        { path: '/', component: Home },
        { path: '/settings', component: Settings },
    ],
});
```

### Svelte (svelte-spa-router)

```svelte
<script>
    import Router from "svelte-spa-router";
</script>

<Router routes={{
    "/": Home,
    "/settings": Settings,
    "*": NotFound,
}} />
```

---

## Windows Installer (NSIS)

### Prerequisites
Install NSIS: https://nsis.sourceforge.io/Download

macOS: `brew install nsis`

### Build
```bash
wails build -nsis
```

Configuration is read from `build/windows/installer/` and `wails.json` `info` section.

### NSIS Types
In `wails.json`:
```json
{
  "nsisType": "multiple"
}
```
- `"multiple"` — one installer per architecture (default)
- `"single"` — universal installer for all architectures

---

## Code Signing

### Windows

Use `signtool.exe` from Windows SDK:
```bash
signtool sign /fd SHA256 /a /f cert.pfx /p password /t http://timestamp.digicert.com myapp.exe
```

### macOS

1. Get a Developer ID certificate from Apple
2. Sign:
```bash
codesign --deep --force --options runtime --sign "Developer ID Application: Your Name" myapp.app
```
3. Notarize:
```bash
xcrun notarytool submit myapp.zip --apple-id your@email.com --team-id TEAMID --password app-specific-password
```

For Wails apps, a standard code signing certificate works (no need for EV/kernel signing).

---

## Mac App Store

1. Create App Store provisioning profile
2. Build with entitlements
3. Sign with `3rd Party Mac Developer Application` certificate
4. Package as `.pkg` with `3rd Party Mac Developer Installer` certificate
5. Submit via Transporter or `xcrun altool`

---

## Cross-Platform CI/CD

### GitHub Actions Example

```yaml
name: Build
on: [push, pull_request]

jobs:
  build:
    strategy:
      matrix:
        platform: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.platform }}
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - uses: actions/setup-node@v3
        with:
          node-version: 18
      - name: Install Wails
        run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
      - name: Install Linux deps
        if: matrix.platform == 'ubuntu-latest'
        run: sudo apt-get update && sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev
      - name: Build
        run: wails build
      - uses: actions/upload-artifact@v3
        with:
          name: build-${{ matrix.platform }}
          path: build/bin/*
```

---

## Obfuscation

Using garble for binary obfuscation:

```bash
wails build -obfuscated -garbleargs "-literals"
```

In `wails.json`:
```json
{
  "obfuscated": true,
  "garbleargs": "-literals"
}
```

---

## Manual Builds

The build process manually:

1. Install frontend deps: `cd frontend && npm install`
2. Build frontend: `cd frontend && npm run build`
3. Generate bindings: `wails generate module`
4. Build Go binary:
```bash
go build -tags desktop,production -ldflags "-w -s" -o build/bin/myapp
```

On Windows, add `-ldflags "-w -s -H windowsgui"` to hide the console window.

On macOS, use `-tags desktop,production,darwin` and wrap in `.app` bundle.

---

## Troubleshooting

### Common Issues

**`wails doctor` shows missing dependencies**
- Linux: `sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev`
- macOS: `xcode-select --install`

**Build fails with "npm not found"**
- Ensure Node.js is installed and in PATH

**Window is blank / white**
- Check `frontend:build` command produces output in `assetdir`
- Check `//go:embed` directive matches the build output path

**Hot reload not working**
- Ensure `frontend:dev:watcher` and `frontend:dev:serverUrl` are set in `wails.json`
- Use `"auto"` for Vite projects

**macOS: "App is damaged and can't be opened"**
- The app needs code signing. For development: `xattr -cr myapp.app`

**Linux: Video tag "ended" event not firing**
WebkitGTK bug. Workaround:
```javascript
videoTag.addEventListener("timeupdate", (event) => {
    if (event.target.duration - event.target.currentTime < 0.2) {
        event.target.dispatchEvent(new Event("ended"));
    }
});
```

**Linux: Panic not recoverable**
Call `runtime.ResetSignalHandlers()` before code that might panic (WebKit overrides Go signal handlers).
