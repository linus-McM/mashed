# `mashed-pty-helper` Implementation Plan

**Date:** 2026-04-09  
**Branch:** feature/xyflow_design → then merge to main  
**Root cause addressed:** Wails/Cocoa callback context contaminates fork/exec regardless of goroutine, OS thread, or channel delegation. Helper binary runs outside that context permanently.

---

## Overview

A separate Go binary (`mashed-pty-helper`) is launched from `main()` before `wails.Run()`. It owns all PTY lifecycle: spawning, resizing, and teardown. The main Wails binary communicates with it via a Unix domain socket using standard JSON messages for control and `SCM_RIGHTS` for passing the PTY master file descriptor back to the main process (so it can do I/O directly without proxying every byte through the helper).

### A few things worth calling out from the plan:

The SCM_RIGHTS proof of concept (Phase 3 step) is the highest-value early test. Write a 30-line standalone Go program that spawns the helper, receives the fd, and does io.Copy(os.Stdout, ptmx) — confirm bytes flow before wiring any of this into Wails.
The waitForHelper approach matters. Polling os.Stat on the socket path is reliable enough, but if you want cleaner signalling, pipe stdout from the helper and read "READY\n" before handing the pipe to os.Stdout. Either works.
Parent death detection in the helper — add a goroutine that polls syscall.Kill(ppid, 0) every 2 seconds. If it returns ESRCH, the main process is gone and the helper should exit. This prevents ghost helper processes after a crash.
Implementation order matters: do the standalone SCM_RIGHTS test before touching SessionManager. The client.go / server.go layer can be developed and tested completely independently of Wails.

---

## Repository Layout

```

mashed/
├── main.go              ← KEEP at root (Wails requirement)
├── cmd/
│   └── pty-helper/      ← NEW
│       └── main.go
├── internal/
│   ├── terminal/
│   │   ├── manager.go   ← strip out direct pty.StartWithSize, use HelperClient
│   │   ├── spawner.go   ← DELETE (channel spawner experiment)
│   │   └── helper/
│   │       ├── client.go     ← NEW: dials socket, sends SpawnRequest, recvs fd
│   │       ├── server.go     ← NEW: listener logic (shared or helper-only)
│   │       └── protocol.go   ← NEW: shared JSON message types
├── build/
│   └── darwin/
│       ├── entitlements.plist   ← KEEP
│       └── build-and-sign.sh    ← REWORK (build both binaries, embed helper)
```

**⚠ Keep `main.go` at project root.** Wails expects `main.go` at the project root — it generates bindings relative to it. Do NOT move it to `cmd/mashed/`. Only add `cmd/pty-helper/main.go` for the helper binary.

---

## Phase 1 — Helper Binary (`cmd/pty-helper/main.go`)

### 1.1 Socket path

```go
func socketPath() string {
    return filepath.Join(os.TempDir(), fmt.Sprintf("mashed-pty-%d.sock", os.Getpid()))
}
```

Use the parent PID (passed as an env var `MASHED_PARENT_PID`) so multiple mashed instances don't collide and cleanup is unambiguous.

### 1.2 Protocol (`internal/terminal/helper/protocol.go`)

```go
type SpawnRequest struct {
    ID      string   `json:"id"`       // UUID, echoed in response
    Shell   string   `json:"shell"`    // e.g. /bin/zsh
    Args    []string `json:"args"`
    Env     []string `json:"env"`
    Cwd     string   `json:"cwd"`
    Cols    uint16   `json:"cols"`
    Rows    uint16   `json:"rows"`
}

type SpawnResponse struct {
    ID    string `json:"id"`
    PID   int    `json:"pid"`
    Error string `json:"error,omitempty"`
    // PTY master fd delivered via SCM_RIGHTS on same conn
}

type KillRequest struct {
    ID     string `json:"id"`
    Signal int    `json:"signal"` // syscall.SIGTERM etc
}

// NOTE: ResizeRequest removed from protocol. Once the main process receives the
// PTY master fd via SCM_RIGHTS, it can call pty.Setsize() directly on that fd.
// No need to round-trip through the helper for resize. This simplifies the
// protocol to just Spawn and Kill.
```

Each message is length-prefixed (4-byte big-endian uint32) followed by JSON. This avoids newline-delimited parsing issues with embedded JSON strings.

### 1.3 Helper main loop

```
1. Read MASHED_PTY_SOCK env var (socket path) — fail fast if missing
2. Listen on Unix socket (unlink first in case of stale file)
3. Write "READY\n" to stdout so parent knows it's listening
4. Accept connections in a loop (one conn per spawn; or multiplex — see note)
5. For each conn:
   a. Read length-prefixed message
   b. Decode as SpawnRequest / ResizeRequest / KillRequest
   c. Dispatch
```

**Multiplexing note:** Single long-lived connection (one `net.UnixConn` per mashed instance) with message routing by session ID is simpler to implement and avoids the overhead of a new connection per spawn. Recommended.

**⚠ SCM_RIGHTS ordering caveat:** If two spawns are in flight concurrently, the fd delivery order must match the response order. The `pending` map routes JSON responses by ID, but `recvFd` is positional — fd #1 arrives for whoever calls `recvFd` first, which may not match the response ID. **Mitigation:** Serialize spawn requests (only one in flight at a time via a mutex on the client side), OR pair fd receipt atomically with the JSON response in the same read call.

### 1.4 doSpawn

```go
func doSpawn(req SpawnRequest, conn *net.UnixConn) {
    cmd := exec.Command(req.Shell, req.Args...)
    cmd.Env = req.Env
    cmd.Dir = req.Cwd
    ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: req.Cols, Rows: req.Rows})
    if err != nil {
        sendResponse(conn, SpawnResponse{ID: req.ID, Error: err.Error()})
        return
    }
    // Send JSON response
    sendResponse(conn, SpawnResponse{ID: req.ID, PID: cmd.Process.Pid})
    // Send PTY master fd via SCM_RIGHTS
    sendFd(conn, int(ptmx.Fd()))
    // Store ptmx for later Resize/Kill
    sessions.Store(req.ID, &Session{cmd: cmd, ptmx: ptmx})
}
```

### 1.5 sendFd / recvFd (SCM_RIGHTS)

```go
// helper side — send fd
func sendFd(conn *net.UnixConn, fd int) error {
    rights := syscall.UnixRights(fd)
    _, _, err := conn.WriteMsgUnix([]byte{0}, rights, nil)
    return err
}

// client side — receive fd
func recvFd(conn *net.UnixConn) (int, error) {
    oob := make([]byte, syscall.CmsgSpace(4))
    _, oobn, _, _, err := conn.ReadMsgUnix(make([]byte, 1), oob)
    if err != nil { return 0, err }
    msgs, err := syscall.ParseSocketControlMessage(oob[:oobn])
    if err != nil { return 0, err }
    fds, err := syscall.ParseUnixRights(&msgs[0])
    if err != nil { return 0, err }
    return fds[0], nil
}
```

The received fd is a full PTY master in the main process's fd table. `os.NewFile(uintptr(fd), "ptmx")` wraps it for use with `io.Copy` etc.

---

## Phase 2 — Helper Client (`internal/terminal/helper/client.go`)

```go
type Client struct {
    conn     *net.UnixConn
    mu       sync.Mutex
    pending  map[string]chan SpawnResponse
}

func Dial(sockPath string) (*Client, error) { ... }

func (c *Client) Spawn(ctx context.Context, req SpawnRequest) (*os.File, int, error) {
    ch := make(chan SpawnResponse, 1)
    c.mu.Lock()
    c.pending[req.ID] = ch
    c.mu.Unlock()

    if err := c.send(req); err != nil { return nil, 0, err }

    select {
    case resp := <-ch:
        if resp.Error != "" { return nil, 0, errors.New(resp.Error) }
        fd, err := recvFd(c.conn)
        if err != nil { return nil, 0, err }
        return os.NewFile(uintptr(fd), "ptmx"), resp.PID, nil
    case <-ctx.Done():
        return nil, 0, ctx.Err()
    }
}

func (c *Client) Kill(id string, sig syscall.Signal) error  { ... }
// NOTE: No Resize method — main process calls pty.Setsize() directly on the
// received fd. No need to proxy resize through the helper.
```

A background goroutine on `Client` reads incoming responses and routes to `pending` channels.

---

## Phase 3 — main.go Integration

```go
func main() {
    // 1. Resolve helper binary path (sibling of this binary)
    exe, _ := os.Executable()
    helperPath := filepath.Join(filepath.Dir(exe), "mashed-pty-helper")

    sockPath := filepath.Join(os.TempDir(),
        fmt.Sprintf("mashed-pty-%d.sock", os.Getpid()))

    // 2. Launch helper BEFORE wails.Run()
    helperCmd := exec.Command(helperPath)
    helperCmd.Env = append(os.Environ(),
        "MASHED_PTY_SOCK="+sockPath,
        fmt.Sprintf("MASHED_PARENT_PID=%d", os.Getpid()),
    )
    helperCmd.Stdout = os.Stdout  // surface helper logs during dev
    helperCmd.Stderr = os.Stderr
    if err := helperCmd.Start(); err != nil {
        log.Fatalf("failed to start pty helper: %v", err)
    }

    // 3. Wait for READY signal (read from helper's stdout pipe instead — see note)
    waitForHelper(sockPath, 3*time.Second)

    // 4. Dial
    client, err := helper.Dial(sockPath)
    if err != nil {
        log.Fatalf("failed to dial pty helper: %v", err)
    }

    // 5. Graceful shutdown
    defer func() {
        client.Close()
        helperCmd.Process.Signal(syscall.SIGTERM)
        helperCmd.Wait()
        os.Remove(sockPath)
    }()

    // 6. Normal Wails startup — pass client (may be nil if helper unavailable)
    app := NewApp(client)
    wails.Run(...)
}
```

### Graceful degradation

If the helper binary doesn't exist or fails to start, the app should still launch with terminal functionality disabled — not hard crash. Pass `nil` for the helper client and have `SessionManager.Spawn()` return a clear error that the frontend can display as a toast:

```go
if m.helperClient == nil {
    return nil, fmt.Errorf("terminal unavailable: PTY helper not running")
}
```

The frontend `spawnTerminalInRepo` catch block should show a visible error instead of silently swallowing it:

```js
catch (e) {
    // TODO: show toast/notification to user
    console.error('Terminal spawn failed:', e);
}
```

**waitForHelper:** Poll `os.Stat(sockPath)` with a 50ms tick and 3s timeout. The socket file appearing means the helper is listening. Alternatively, read a `"READY\n"` line from the helper's stdout pipe before redirecting to `os.Stdout` — either works.

---

## Phase 4 — SessionManager Rework

Remove all direct `pty.StartWithSize` calls. `Spawn()` becomes:

```go
func (m *Manager) Spawn(ctx context.Context, req SpawnParams) (*Session, error) {
    id := uuid.New().String()
    ptmx, pid, err := m.helperClient.Spawn(ctx, helper.SpawnRequest{
        ID:   id,
        Shell: req.Shell,
        Env:   req.Env,
        Cwd:   req.Cwd,
        Cols:  req.Cols,
        Rows:  req.Rows,
    })
    if err != nil {
        return nil, fmt.Errorf("pty helper spawn: %w", err)
    }
    s := &Session{id: id, pid: pid, ptmx: ptmx}
    m.sessions.Store(id, s)
    go s.readLoop(m.outputCh)
    return s, nil
}
```

---

## Phase 5 — Build & Bundle

### 5.1 Build both binaries

In `justfile` (or `Makefile`):

```
build:
    go build -o build/bin/mashed-pty-helper ./cmd/pty-helper
    wails build                                              # builds main binary
    # Copy helper into the app bundle
    cp build/bin/mashed-pty-helper \
       "build/bin/mashed.app/Contents/MacOS/mashed-pty-helper"
```

### 5.2 Sign both binaries

```bash
codesign --force --options runtime \
  --sign "<your signing identity>" \
  --entitlements build/darwin/entitlements.plist \
  "build/bin/mashed.app/Contents/MacOS/mashed-pty-helper"

codesign --force --options runtime \
  --sign "<your signing identity>" \
  --entitlements build/darwin/entitlements.plist \
  "build/bin/mashed.app"
```

Sign the helper **first**, then the bundle — macOS re-verifies inner binaries when you sign the outer `.app`.

### 5.3 `wails dev` mode

In dev mode Wails runs the Go binary directly, not from a `.app` bundle. `os.Executable()` will be in a temp dir. Handle this:

```go
func resolveHelperPath() string {
    exe, _ := os.Executable()
    candidate := filepath.Join(filepath.Dir(exe), "mashed-pty-helper")
    if _, err := os.Stat(candidate); err == nil {
        return candidate
    }
    // Dev fallback: look relative to working directory
    if wd, err := os.Getwd(); err == nil {
        candidate = filepath.Join(wd, "build/bin/mashed-pty-helper")
        if _, err := os.Stat(candidate); err == nil {
            return candidate
        }
    }
    log.Fatal("mashed-pty-helper not found — run 'just build-helper' first")
    return ""
}
```

Add a `build-helper` recipe to `justfile` that only builds the helper binary.

---

## Phase 6 — Cleanup

Remove all experimental code from the investigation:

| File | Action |
|------|--------|
| `internal/terminal/spawner.go` | Delete |
| `main.go` — TestFork debug code | Remove |
| `main.go` — spawner startup | Remove (replaced by helper launch) |
| `app.go` — spawner param in `NewApp()` | Remove |
| `app_spawn.go` — debug logging | Remove |
| `frontend/src/views/NotificationFeed.svelte` — debug logging | Remove |
| `build/darwin/build-and-sign.sh` | Rework per Phase 5 |
| `frontend/wailsjs/runtime/runtime.js` | Keep (valid fix for browser dev mode) |
| `build/darwin/entitlements.plist` | Keep |

---

## Implementation Order

1. **Protocol types** (`protocol.go`) — no dependencies, define these first
2. **Helper binary** (`cmd/pty-helper/main.go` + `server.go`) — can test standalone
3. **SCM_RIGHTS proof of concept** — verify fd passing before wiring into Wails
4. **Client** (`client.go`) — depends on protocol, independent of Wails
5. **SessionManager rework** — swap in client, remove direct PTY calls
6. **main.go integration** — launch helper, dial, pass to app
7. **Build system** — justfile recipes, bundle copy, signing
8. **Cleanup** — remove experimental code
9. **Test** in both `wails dev` and `wails build` modes

---

## Risk Items

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| SCM_RIGHTS fd passing on macOS arm64 has quirks | Low | Test standalone first (Phase 3 step) |
| Helper binary quarantined by Gatekeeper in dev | Medium | Sign with Apple Dev cert from the start |
| `wails dev` can't find helper binary | Medium | `resolveHelperPath()` fallback (Phase 5.3) |
| Helper crashes silently, main blocks on `waitForHelper` | Low | 3s timeout + graceful degradation (app runs, terminals disabled) |
| Stale socket file from crash blocks next launch | Low | `os.Remove(sockPath)` at helper startup before `Listen()` |
| Parent crashes, helper becomes orphan zombie | Medium | Use `kqueue` with `EVFILT_PROC` + `NOTE_EXIT` on parent PID for instant detection (macOS-native, no polling). Fallback: poll `kill(ppid, 0)` and exit when ppid changes to 1 (launchd reparent). |

---

## Testing Checklist

- [ ] Standalone helper binary spawns `/bin/zsh` and returns fd — verify with a tiny Go test client
- [ ] fd received in main process is readable/writable (echo test)
- [ ] `wails dev` — click Terminal button, shell appears
- [ ] `wails build` — signed bundle, click Terminal button, shell appears
- [ ] Multiple concurrent sessions (open 5 terminals simultaneously)
- [ ] Resize propagates to PTY correctly
- [ ] Shell exit detected and session cleaned up
- [ ] mashed quit → helper process exits cleanly (no zombie)
- [ ] mashed crash → helper detects parent gone (poll `kill(ppid, 0)`) and exits
