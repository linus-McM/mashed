# Wails CLI Reference

Install: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## wails init

Create a new Wails project.

```bash
wails init -n <name> [-t <template>] [-g] [-l] [-q] [-ide <ide>]
```

| Flag | Description |
|------|-------------|
| `-n <name>` | Project name (required) |
| `-t <template>` | Template: `vue`, `react`, `svelte`, `preact`, `lit`, `vanilla` (+ `-ts` variants) |
| `-g` | Initialize git repository |
| `-l` | List available templates |
| `-q` | Suppress output |
| `-ide <ide>` | Generate IDE project files: `vscode`, `goland` |

Remote templates from GitHub are supported:
```bash
wails init -n myapp -t https://github.com/user/wails-template
```

## wails build

Build production binary.

```bash
wails build [flags]
```

| Flag | Description |
|------|-------------|
| `-platform <os/arch>` | Target platform (e.g., `windows/amd64`, `darwin/universal`) |
| `-o <filename>` | Output filename |
| `-clean` | Clean build directory |
| `-ldflags <flags>` | Go linker flags |
| `-tags <tags>` | Build tags (comma separated) |
| `-race` | Build with race detector |
| `-trimpath` | Remove file system paths from binary |
| `-upx` | Compress binary with UPX |
| `-upxflags <flags>` | Flags for UPX |
| `-nsis` | Generate NSIS installer (Windows) |
| `-s` | Skip frontend build |
| `-f` | Force rebuild |
| `-skipbindings` | Skip bindings generation |
| `-nosyncgomod` | Don't sync go.mod |
| `-nocolour` | Disable colored output |
| `-loglevel <level>` | `Trace`, `Debug`, `Info`, `Warning`, `Error` |
| `-webview2 <strategy>` | WebView2 installer: `download`, `embed`, `browser`, `error` |
| `-obfuscated` | Obfuscate with garble |
| `-garbleargs <args>` | Arguments for garble |

### Platform Values

| Value | Target |
|-------|--------|
| `windows/amd64` | Windows 64-bit |
| `windows/arm64` | Windows ARM |
| `darwin/amd64` | macOS Intel |
| `darwin/arm64` | macOS Apple Silicon |
| `darwin/universal` | macOS Universal Binary |
| `linux/amd64` | Linux 64-bit |
| `linux/arm64` | Linux ARM |

### WebView2 Strategies

| Strategy | Description |
|----------|-------------|
| `download` | Download and install if missing (default) |
| `embed` | Embed installer in binary |
| `browser` | Open download page in browser |
| `error` | Error if not installed |

## wails dev

Run in development mode with live reload.

```bash
wails dev [flags]
```

| Flag | Description |
|------|-------------|
| `-browser` | Open browser on start |
| `-e <extensions>` | File extensions to watch (default: `.go`) |
| `-reloaddirs <dirs>` | Additional directories to watch for changes |
| `-compiler <path>` | Custom Go compiler path |
| `-loglevel <level>` | Log level |
| `-assetdir <dir>` | Serve assets from directory instead of embed.FS |
| `-frontenddevserverurl <url>` | Use external dev server (e.g., `http://localhost:5173`) |
| `-appargs <args>` | Arguments passed to app |
| `-save` | Save flags to wails.json defaults |
| `-race` | Enable race detector |
| `-s` | Skip frontend build |
| `-nosyncgomod` | Don't sync go.mod |
| `-nocolour` | Disable colored output |
| `-noreload` | Disable auto-reload on changes |
| `-debounce <ms>` | Debounce time for file watcher |

## wails doctor

Check system dependencies and environment.

```bash
wails doctor
```

Reports Go version, Wails version, platform, and checks all required dependencies.

## wails generate

### wails generate module

Generate `wailsjs` module (TypeScript bindings for bound Go methods).

```bash
wails generate module
```

### wails generate template

Create a reusable template from an existing project.

```bash
wails generate template -name <name> [-frontend <dir>]
```

## wails update

Update the Wails CLI.

```bash
wails update [-pre]  # -pre for pre-release versions
```

## wails version

Print Wails CLI version.

```bash
wails version
```

## wails show

Show project information.

```bash
wails show
```
