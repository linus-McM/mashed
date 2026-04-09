# Story 6: main.go Integration -- Launch Helper and Wire Into App

**Priority:** P0-critical
**Domain:** fullstack
**Estimated Complexity:** M
**Depends On:** Story 4 (ptyhelper-04-client), Story 5 (ptyhelper-05-session-manager-rework)
**Status:** ready

## Description

Wire the PTY helper lifecycle into the main Wails application. This story modifies `main.go` to launch the helper binary before `wails.Run()`, wait for its READY signal, dial the Unix socket, create the helper Client, and pass it through to `NewApp`. It also implements graceful shutdown (SIGTERM to helper on exit), graceful degradation (app launches even if helper unavailable), and `wails dev` mode support (resolving the helper binary path in both bundle and development layouts). The frontend error handling for terminal spawn failures is also improved.

## Developer Notes

### Architecture
- **Modified file:** `main.go` -- add helper launch, wait, dial before `wails.Run()`
- **Modified file:** `app.go` -- `NewApp` accepts `*helper.Client` (from Story 5)
- **Modified file:** `frontend/src/views/NotificationFeed.svelte` -- show error toast on spawn failure
- **Modified file:** `frontend/src/views/AgentDetail.svelte` -- show error toast on spawn failure

### main.go Changes

The flow becomes:
```
1. resolveHelperPath() -> find the helper binary
2. Generate socket path: /tmp/mashed-pty-{pid}.sock
3. Launch helper subprocess with MASHED_PTY_SOCK and MASHED_PARENT_PID env vars
4. waitForHelper() -> poll for socket file or read READY from stdout pipe
5. helper.Dial(sockPath) -> create Client
6. NewApp(client) -> pass Client into App
7. defer: cleanup (close client, SIGTERM helper, wait, remove socket)
8. wails.Run(...)
```

### resolveHelperPath
```go
func resolveHelperPath() (string, error) {
    exe, err := os.Executable()
    if err != nil {
        return "", fmt.Errorf("resolve helper: executable path: %w", err)
    }
    // Production: helper is sibling to the main binary inside .app bundle
    candidate := filepath.Join(filepath.Dir(exe), "mashed-pty-helper")
    if _, err := os.Stat(candidate); err == nil {
        return candidate, nil
    }
    // Dev fallback: helper in build/bin/ relative to working directory
    if wd, err := os.Getwd(); err == nil {
        candidate = filepath.Join(wd, "build", "bin", "mashed-pty-helper")
        if _, err := os.Stat(candidate); err == nil {
            return candidate, nil
        }
    }
    return "", fmt.Errorf("resolve helper: mashed-pty-helper not found")
}
```

### waitForHelper
```go
func waitForHelper(sockPath string, timeout time.Duration) error {
    deadline := time.After(timeout)
    tick := time.NewTicker(50 * time.Millisecond)
    defer tick.Stop()
    for {
        select {
        case <-deadline:
            return fmt.Errorf("helper not ready after %v", timeout)
        case <-tick.C:
            if _, err := os.Stat(sockPath); err == nil {
                return nil
            }
        }
    }
}
```

### Graceful Degradation
If any step fails (binary not found, launch fails, dial fails), log the error and proceed with `NewApp(nil)`:
```go
client, err := launchHelper()
if err != nil {
    log.Printf("WARNING: PTY helper unavailable: %v — terminals will be disabled", err)
    // client remains nil
}
app := NewApp(client) // nil client = terminals disabled
```

The frontend already calls `SpawnTerminal` and `SpawnAgent` which return errors. The catch blocks currently swallow errors. Improve them to show a visible notification.

### Frontend Changes

**NotificationFeed.svelte** -- `spawnTerminalInRepo`:
```javascript
async function spawnTerminalInRepo(repo) {
    if (spawningTerminal) return;
    spawningTerminal = repo.path;
    try {
        const target = await SpawnTerminal(repo.path);
        // ... existing success handling
    } catch (e) {
        // NEW: show visible error instead of just console.error
        console.error('Terminal spawn failed:', e);
        // Emit or display an error notification
    } finally {
        spawningTerminal = null;
    }
}
```

The exact UI treatment (toast, inline error, etc.) should use whatever notification pattern the app already has. The key requirement is that the error is **visible to the user**, not silently swallowed.

### Shutdown Sequence
```go
defer func() {
    if client != nil {
        client.Close()
    }
    if helperCmd != nil {
        helperCmd.Process.Signal(syscall.SIGTERM)
        done := make(chan error, 1)
        go func() { done <- helperCmd.Wait() }()
        select {
        case <-done:
        case <-time.After(5 * time.Second):
            helperCmd.Process.Kill()
        }
    }
    os.Remove(sockPath)
}()
```

### Technical Considerations
- `main.go` must remain at the project root (Wails requirement)
- The helper's stdout must be piped to read "READY\n", then can be redirected to `os.Stdout` for logging
- `os.Getpid()` gives the main process PID, passed as `MASHED_PARENT_PID` to helper
- Socket path format: `/tmp/mashed-pty-{pid}.sock` (includes PID for multi-instance support)
- The helper binary path resolution must handle both `.app` bundle layout and `wails dev` layout

### Risks & Edge Cases
- `wails dev` runs the binary from a temp directory -- `resolveHelperPath` falls back to cwd-relative path
- If the helper binary doesn't exist (first build without `just build-helper`), graceful degradation kicks in
- Socket path in `/tmp` could collide if PID is reused quickly -- extremely unlikely
- If helper hangs during shutdown, the 5-second timeout + Kill prevents zombie
- macOS app bundle signing: the helper binary inside the bundle needs to be signed (Story 7)

### Reference Files
- `main.go` -- current main function (lines 92-124)
- `app.go:NewApp()` -- constructor to modify (line 121)
- `app.go:shutdown()` -- existing shutdown hook (line 172)
- `frontend/src/views/NotificationFeed.svelte:spawnTerminalInRepo` -- error handling to improve (line 340)
- `frontend/src/views/AgentDetail.svelte` -- terminal spawn (line 117)
- `build/darwin/build-and-sign.sh` -- existing build script (reference for Story 7)

## Acceptance Criteria

AC-1: Helper binary is launched before wails.Run
- Given the helper binary exists at the expected path
- When the app starts
- Then the helper process is running before `wails.Run()` is called
- And the helper's socket is available for connections

AC-2: App starts with graceful degradation when helper unavailable
- Given the helper binary does NOT exist
- When the app starts
- Then a warning is logged: "PTY helper unavailable"
- And the app launches normally (Wails window appears)
- And `NewApp` receives nil for the helper client

AC-3: Terminal spawn failure shows visible error in UI
- Given the helper is unavailable (nil client)
- When the user clicks the Terminal button on a repo panel
- Then the button shows an error state (not just silent failure)
- And the error message indicates terminals are unavailable

AC-4: Helper binary resolves in both dev and production layouts
- Given the app is running in `wails dev` mode
- When `resolveHelperPath()` is called
- Then it finds the binary at `{cwd}/build/bin/mashed-pty-helper`
- Given the app is running from a `.app` bundle
- When `resolveHelperPath()` is called
- Then it finds the binary at `{bundle}/Contents/MacOS/mashed-pty-helper`

AC-5: Clean shutdown terminates helper
- Given the helper is running with active sessions
- When the app is closed (Wails shutdown)
- Then the helper receives SIGTERM
- And the helper exits within 5 seconds
- And the socket file is removed

## BDD Test Scenarios

### Scenario 1: Helper launch and readiness

```gherkin
Feature: Helper binary launch at startup

  Scenario: Successful helper launch
    Given the helper binary exists at build/bin/mashed-pty-helper
    And no stale socket file exists
    When main() calls launchHelper()
    Then the helper process is started
    And waitForHelper returns within 3 seconds
    And Dial succeeds

  Scenario: Helper binary not found
    Given the helper binary does not exist at any expected path
    When main() calls launchHelper()
    Then it returns a non-nil error
    And the error message contains "not found"
    And the app proceeds with nil client
```

### Scenario 2: Graceful degradation frontend

```gherkin
Feature: Frontend error display for unavailable terminals

  Scenario: Terminal button shows error when helper unavailable
    Given the app started without a helper (nil client)
    When the user clicks the Terminal button on a repo panel
    Then the SpawnTerminal call returns an error
    And the UI displays a visible error message
    And the button is re-enabled after the error

  Scenario: New Session button shows error when helper unavailable
    Given the app started without a helper (nil client)
    When the user clicks New Session and submits
    Then the spawn call returns an error
    And the UI displays a visible error message
```

### Scenario 3: Clean shutdown

```gherkin
Feature: App shutdown cleans up helper

  Scenario: Normal shutdown
    Given the helper is running with 2 active sessions
    When the app's shutdown callback fires
    Then SIGTERM is sent to the helper process
    And the helper process exits
    And the socket file at /tmp/mashed-pty-{pid}.sock is removed

  Scenario: Helper hangs on shutdown
    Given the helper is running but unresponsive
    When the app's shutdown callback fires
    Then SIGTERM is sent
    And after 5 seconds, SIGKILL is sent
    And the socket file is removed
```

## Tasks / Subtasks

- [ ] Task 1: Implement resolveHelperPath (AC: AC-4)
  - [ ] Add `resolveHelperPath() (string, error)` to `main.go`
  - [ ] Check sibling path (production bundle layout)
  - [ ] Check `{cwd}/build/bin/` (dev layout)
  - [ ] Return clear error if not found

- [ ] Task 2: Implement helper launch and wait (AC: AC-1, AC-2)
  - [ ] Add `launchHelper() (*helper.Client, *exec.Cmd, string, error)` to `main.go`
  - [ ] Launch subprocess with env vars
  - [ ] Implement `waitForHelper` (poll socket file, 3s timeout)
  - [ ] Dial and return client
  - [ ] On any failure, return nil client (graceful degradation)

- [ ] Task 3: Wire into main() (AC: AC-1, AC-2, AC-5)
  - [ ] Modify `main()` to call `launchHelper()` before `wails.Run()`
  - [ ] Pass client to `NewApp(client)`
  - [ ] Add defer block for shutdown (SIGTERM, wait, cleanup)
  - [ ] Update `NewApp` signature in `app.go`

- [ ] Task 4: Improve frontend error handling (AC: AC-3)
  - [ ] Update `spawnTerminalInRepo` in NotificationFeed.svelte to show visible error
  - [ ] Update terminal spawn in AgentDetail.svelte to show visible error
  - [ ] Use existing notification/toast pattern or add inline error display

- [ ] Task 5: Write tests (AC: AC-1, AC-2, AC-4, AC-5)
  - [ ] Test `resolveHelperPath` with mocked filesystem (or test both paths)
  - [ ] Test `waitForHelper` with timeout
  - [ ] Test shutdown sequence (helper receives SIGTERM and exits)
  - [ ] Verify `main_test.go` still passes

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified code in main.go
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Manual verification: `wails dev` with pre-built helper shows terminal working
