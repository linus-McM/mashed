# PTY fork/exec "operation not permitted" Investigation Report

**Date:** 2026-04-09
**Branch:** feature/xyflow_design
**Platform:** macOS Darwin 25.3.0 (Sequoia), arm64
**Wails:** v2.12.0, Go 1.26.1

---

## Problem Statement

Clicking the **Terminal** or **New Session** button on the home screen repo panels does nothing. The Go backend returns:

```
fork/exec /bin/zsh: operation not permitted
```

The error originates from `pty.StartWithSize()` in `internal/terminal/manager.go:Spawn()`, which calls `creack/pty` → `forkpty()` → `fork/exec`.

---

## Error Chain

```
Frontend: spawnTerminalInRepo(repo)
  → SpawnTerminal(repo.path)           [Wails binding]
  → App.SpawnTerminal()                [app_spawn.go:68]
  → App.spawnSession()                 [app_spawn.go:23]
  → SessionManager.Spawn()            [manager.go:42]
  → pty.StartWithSize(cmd, winsize)    [manager.go:69]
  → fork/exec /bin/zsh: operation not permitted
```

The frontend catch block silently swallows the error with `console.error()`, making the button appear to do nothing.

---

## Environment Details

- Binary: `/Users/linus/Development/mashed/build/bin/mashed.app/Contents/MacOS/mashed`
- Wails builds as a macOS `.app` bundle with embedded WebKit (WKWebView)
- The binary has `com.apple.provenance` xattr (sticky, cannot be removed on Sequoia)
- Wails dev server runs on `:34115` (Go bindings) and `:5173` (Vite frontend)

---

## Approaches Tested

### 1. Browser dev mode proxy (FIXED — separate issue)

**Problem:** Running on `localhost:5173` (Vite only), `window.runtime` and `window.go` are undefined, crashing the app on init.

**Fix:** Added recursive no-op Proxy shim in `frontend/wailsjs/runtime/runtime.js`:
```js
function noopProxy() {
    return new Proxy(() => Promise.resolve(), {
        get: () => noopProxy(),
        apply: () => Promise.resolve()
    });
}
if (!window.runtime) window.runtime = noopProxy();
if (!window.go) window.go = noopProxy();
```

**Result:** ✅ Browser dev mode loads without errors. But this is a separate issue from the PTY fork problem.

### 2. Wails dev bridge URL (`:34115`)

**Test:** Connected playwright-cli to `http://localhost:34115` where Go bindings are available.

**Result:** ❌ Same `fork/exec /bin/zsh: operation not permitted` error. Confirmed the error comes from the Go backend, not the frontend.

### 3. macOS entitlements — ad-hoc signed

**File created:** `build/darwin/entitlements.plist` with:
- `com.apple.security.cs.allow-jit`
- `com.apple.security.cs.allow-unsigned-executable-memory`
- `com.apple.security.cs.disable-library-validation`
- `com.apple.security.inherit`
- `com.apple.security.get-task-allow`

**Command:** `codesign --force --deep --sign - --entitlements entitlements.plist mashed.app`

**Result:** ❌ Ad-hoc signing with entitlements does not grant fork/exec privileges on Sequoia.

### 4. macOS entitlements — with `--options runtime`

**Command:** `codesign --force --deep --options runtime --sign - --entitlements entitlements.plist mashed.app`

**Result:** ❌ Same error. Ad-hoc + runtime hardened still insufficient.

### 5. Apple Developer certificate signing

**Identity:** `Apple Development: linus McManamey (5X8A9U965U)`

**Command:** `codesign --force --deep --options runtime --sign "Apple Development: linus McManamey (5X8A9U965U)" --entitlements entitlements.plist mashed.app`

**Verification:**
```
Authority=Apple Development: linus McManamey (5X8A9U965U)
Authority=Apple Worldwide Developer Relations Certification Authority
Authority=Apple Root CA
CodeDirectory flags=0x10000(runtime)
```

**Result:** ❌ Even with a real Apple Developer certificate, runtime hardening, and entitlements, fork/exec is still blocked.

### 6. Build wrapper script (`-compiler` flag)

**Approach:** Created `build/darwin/build-and-sign.sh` that wraps `go build` and re-signs the output binary with entitlements after compilation.

**Problem:** Wails runs its own `Self-signing application` step AFTER the compiler, overwriting our entitlements.

**Result:** ❌ Entitlements stripped by Wails post-build signing.

### 7. macOS Developer Tools permission

**Approach:** User enabled Developer Tools in System Settings > Privacy & Security.

**Result:** ❌ Developer Tools permission applies to terminal apps, not to the Wails-built binary.

### 8. Removing `com.apple.provenance` xattr

**Commands tried:**
```bash
xattr -dr com.apple.provenance mashed.app
xattr -rc mashed.app
xattr -d com.apple.provenance mashed.app/Contents/MacOS/mashed
```

**Result:** ❌ `com.apple.provenance` is sticky on macOS Sequoia and cannot be removed.

### 9. Running binary outside `.app` bundle

**Test:** Copied binary to `/tmp/mashed-test` and ran directly.

**Result:** ❌ Same error. The restriction is on the binary itself, not the `.app` bundle structure.

### 10. Removing `SysProcAttr{Setpgid: true}`

**Change:** Removed `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}` from `Spawn()`.

**Result:** ❌ No change. `Setpgid` is not the cause.

### 11. Using `exec.Command` instead of `exec.CommandContext`

**Change:** Replaced `exec.CommandContext(ctx, ...)` with `exec.Command(...)` to avoid passing the Wails context.

**Result:** ❌ No change. The Wails context is not the cause.

### 12. Using `context.Background()` for PTY spawning

**Change:** `exec.CommandContext(context.Background(), ...)` instead of the Wails app context.

**Result:** ❌ No change.

### 13. Standalone PTY test binary

**Test:** Built a standalone Go binary that just does `pty.Start(exec.Command("/bin/zsh"))`.

**Result:** ✅ Works perfectly. Standalone Go binaries can fork/exec PTY without issues, even with `com.apple.provenance`.

### 14. `init()` fork test in Wails binary

**Test:** Added `pty.Start()` call in Go `init()` function (runs before `main()`).

**Result:** ✅ Fork succeeds in `init()`. The binary CAN fork before `wails.Run()`.

### 15. `main()` start fork test

**Test:** Added `pty.Start()` call at the very beginning of `main()`, before `wails.Run()`.

**Result:** ✅ Fork succeeds at the start of `main()`.

### 16. Delayed fork test (goroutine started before `wails.Run()`)

**Test:** Started a goroutine at the beginning of `main()` that sleeps N seconds then tries fork/exec.

**Results:**
```
[FORK-TEST] SUCCESS at +5s:  pid=66045
[FORK-TEST] SUCCESS at +10s: pid=66372
[FORK-TEST] SUCCESS at +15s: pid=66847
[FORK-TEST] SUCCESS at +20s: pid=67520
```

**Result:** ✅ Fork succeeds even 20+ seconds after Wails/WebKit has fully initialized. The restriction is NOT caused by WebKit initialization.

### 17. Re-exec child process spawner

**Approach:** Re-exec the binary itself with an env var (`MASHED_SPAWNER_SOCKET`) to run as a spawner child that listens on a Unix socket for PTY spawn requests.

**Result:** ❌ The re-exec'd child inherits the same restrictions. `fork/exec /bin/zsh: operation not permitted` in the child process too.

### 18. In-process Unix socket spawner

**Approach:** Run the spawner as a goroutine in the same process, communicating via Unix socket with SCM_RIGHTS fd passing.

**Result:** ❌ Same error. The spawner goroutine's `pty.StartWithSize()` fails.

### 19. Channel-based spawner (goroutine started before `wails.Run()`)

**Approach:** Wails binding callbacks send spawn requests through a Go channel to a goroutine started before `wails.Run()`. The goroutine performs fork/exec.

**Result:** ❌ Still fails. Even though the goroutine was started pre-Wails, fork/exec fails when triggered via channel from a Wails callback.

### 20. Channel-based spawner with `runtime.LockOSThread()`

**Approach:** Same as #19, but the spawner goroutine calls `runtime.LockOSThread()` to pin it to a dedicated OS thread, preventing migration to Cocoa event loop threads.

**Result:** ❌ Still fails.

---

## Key Findings

### What works:
- Fork/exec from `init()` — before any Go runtime setup
- Fork/exec from the start of `main()` — before `wails.Run()`
- Fork/exec from a goroutine that **autonomously** decides to fork (delayed timer test) — even 20+ seconds after Wails/WebKit init
- Fork/exec from standalone Go binaries (no Wails)
- Fork/exec from `go run` / `go test` contexts

### What fails:
- Fork/exec triggered by a **Wails binding callback** — regardless of:
  - Which goroutine performs the fork
  - Whether `LockOSThread()` is used
  - Whether delegation happens via channel, Unix socket, or direct call
  - Whether the goroutine was created before or after `wails.Run()`
  - Entitlements, codesigning identity, or xattr removal
  - `SysProcAttr`, `exec.CommandContext` vs `exec.Command`, `context.Background()`

### The pattern:
The restriction is **not about timing, threads, or signing**. It is specifically about the **call chain originating from a Wails/WebKit/Cocoa callback**. Something in the macOS runtime marks the execution context when a WebKit callback is being processed, and `fork/exec` is denied within that context — even if the actual fork is delegated to a completely separate goroutine via a channel.

This suggests macOS Sequoia has a **process-level lock or security token** that is set when the Cocoa event loop dispatches a WebKit callback, and `posix_spawn`/`fork` checks this token regardless of which thread performs the call.

---

## How cmux (Claude Code Desktop) Solves This

From analyzing the cmux repository:

1. **Does NOT fork/exec shells directly.** Instead embeds **GhosttyKit** (Zig-based terminal emulator) as a compiled `.xcframework` linked at build time.
2. **Ghostty helper binary** bundled at `app.app/Contents/Resources/bin/ghostty`, built separately and signed with entitlements.
3. **Full Apple Developer signing** with `--options runtime` and notarization.
4. **Built with Xcode** (native Swift app), not a cross-platform framework like Wails.

---

## Proposed Solutions (Not Yet Tested)

### A. Pre-fork pool
Pre-fork a pool of shell PTY sessions at startup (when fork works from autonomous goroutines). Hand out pre-forked sessions when the user clicks Terminal. Requires managing the lifecycle of idle PTY sessions.

### B. Separate helper binary
Build a standalone `mashed-pty-helper` Go binary that runs as a separate process and handles PTY spawning via IPC (Unix socket). The helper binary would NOT be a re-exec of the Wails binary — it would be a completely separate binary built from a separate `main` package.

### C. Embed a terminal emulator library
Similar to cmux's GhosttyKit approach — embed a terminal emulator that handles PTY creation internally, bypassing the Go `fork/exec` path.

### D. Use `nsTask` or `posix_spawn` via CGo
Call macOS-native process spawning APIs directly via CGo, which may have different security checks than Go's `os/exec` → `fork/exec` path.

### E. Defer to tmux
Fall back to tmux for terminal session management (the original architecture before the PTY refactor). tmux runs as an independent process and can fork without restrictions.

---

## Files Modified During Investigation

- `frontend/wailsjs/runtime/runtime.js` — No-op proxy for browser dev mode (KEEP)
- `build/darwin/entitlements.plist` — macOS entitlements (KEEP for future use)
- `build/darwin/build-and-sign.sh` — Build wrapper script (can remove)
- `internal/terminal/spawner.go` — Channel-based spawner (needs rework or removal)
- `internal/terminal/manager.go` — Spawner integration + removed SysProcAttr (needs cleanup)
- `app.go` — `NewApp()` signature changed to accept spawner (needs cleanup)
- `main.go` — Spawner startup + TestFork debug code (needs cleanup)
- `app_spawn.go` — Debug logging in `SpawnTerminal` (remove after fix)
- `frontend/src/views/NotificationFeed.svelte` — Debug logging in `spawnTerminalInRepo` (remove after fix)
- `justfile` — `dev` recipe was modified and reverted

---

## Next Steps

1. Choose a solution approach (A-E above)
2. Clean up debug logging and experimental spawner code
3. Implement the chosen solution
4. Test in both `wails dev` and `wails build` modes
5. Update buglog and cerebrum with findings
