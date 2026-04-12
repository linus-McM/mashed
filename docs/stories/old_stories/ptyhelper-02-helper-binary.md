# Story 2: PTY Helper Binary (Server)

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** L
**Depends On:** Story 1 (ptyhelper-01-protocol-types)
**Status:** done

## Description

Build the `mashed-pty-helper` binary that runs as a separate process, listens on a Unix domain socket, and handles PTY spawn/kill requests. This binary is the core of the solution -- it runs outside the Wails/Cocoa callback context and can freely `fork/exec` without macOS "operation not permitted" errors. It receives spawn requests, creates PTY sessions via `creack/pty`, sends the PTY master fd back to the caller via SCM_RIGHTS, and manages session lifecycle including orphan detection when the parent process dies.

## Developer Notes

### Architecture
- **New file:** `cmd/pty-helper/main.go` -- binary entry point
- **New file:** `internal/terminal/helper/server.go` -- server logic (testable without main)
- Uses protocol types from Story 1 (`internal/terminal/helper/protocol.go`)
- Single long-lived Unix connection (multiplexed by session ID), not one-connection-per-spawn

### Server Lifecycle
1. Read `MASHED_PTY_SOCK` env var for socket path -- fail fast if missing
2. Read `MASHED_PARENT_PID` env var for orphan detection
3. `os.Remove(sockPath)` to clean up stale socket file from previous crash
4. `net.ListenUnix("unix", &net.UnixAddr{Name: sockPath, Net: "unix"})`
5. Write `"READY\n"` to stdout so parent knows it is listening
6. Accept a single connection (one per mashed instance, multiplexed)
7. Read loop: decode `Envelope`, dispatch to `handleSpawn` or `handleKill`
8. On shutdown: close all PTY sessions, remove socket file

### Session Tracking
```go
type helperSession struct {
    cmd  *exec.Cmd
    ptmx *os.File
}

// sessions is keyed by SpawnRequest.ID
var sessions sync.Map // map[string]*helperSession
```

### handleSpawn
```go
func handleSpawn(conn *net.UnixConn, req SpawnRequest) {
    cmd := exec.Command(req.Shell, req.Args...)
    cmd.Env = req.Env
    cmd.Dir = req.Cwd
    cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
    ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: req.Cols, Rows: req.Rows})
    // Send SpawnResponse (with error or PID)
    // If success: SendFd(conn, int(ptmx.Fd()))
    // Store in sessions map
}
```

### handleKill
- Look up session by ID in the `sessions` map
- Send `syscall.Kill(-pid, signal)` to the process group (negative PID for group kill)
- Close the ptmx file
- Remove from sessions map

### Orphan Detection
- Start a goroutine that checks if the parent process is still alive
- **Preferred (macOS):** Use `kqueue` with `EVFILT_PROC` + `NOTE_EXIT` on the parent PID for instant notification (no polling delay)
- **Fallback:** Poll `syscall.Kill(ppid, 0)` every 2 seconds; if it returns `ESRCH`, parent is gone
- On parent death: kill all sessions, remove socket, `os.Exit(0)`
- Parse `MASHED_PARENT_PID` from env var

### Technical Considerations
- `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}` is critical so Kill can terminate the entire process group
- The helper should log to stderr (stdout is reserved for the "READY\n" signal)
- After sending `READY\n`, the helper can optionally close stdout or redirect to stderr
- Socket file permissions: default Unix socket permissions are fine (only local user access)
- The helper binary MUST be buildable independently: `go build -o build/bin/mashed-pty-helper ./cmd/pty-helper`

### Risks & Edge Cases
- If two spawn requests arrive concurrently, fd sending order must match response order. The server processes requests sequentially on a single goroutine (no concurrent handleSpawn), so this is inherently serialized. The CLIENT side must also serialize (Story 4).
- If the shell path doesn't exist (e.g., `/bin/fish` not installed), `pty.StartWithSize` returns a clear error -- pass it through in SpawnResponse.Error
- Stale socket file from a previous crash: `os.Remove` before `Listen` handles this
- Helper must handle SIGTERM gracefully: kill all sessions, remove socket, exit

### Reference Files
- `internal/terminal/helper/protocol.go` -- types from Story 1
- `internal/terminal/manager.go:Spawn()` -- existing spawn logic to mirror (lines 38-91)
- `internal/terminal/session.go:Kill()` -- process group kill pattern (lines 262-273)

## Acceptance Criteria

AC-1: Helper binary starts and listens on Unix socket
- Given `MASHED_PTY_SOCK=/tmp/test-pty.sock` and `MASHED_PARENT_PID={current PID}` are set
- When the helper binary is executed
- Then it writes `"READY\n"` to stdout
- And the socket file exists at the specified path
- And the helper accepts Unix connections on that socket

AC-2: Helper spawns a PTY and returns fd via SCM_RIGHTS
- Given the helper is running and a client connects
- When a SpawnRequest for `/bin/echo hello` is sent
- Then the helper responds with a SpawnResponse containing a valid PID
- And sends the PTY master fd via SCM_RIGHTS on the same connection
- And the received fd is readable (the echo output appears)

AC-3: Helper kills a session by ID
- Given a session "sess-001" was previously spawned
- When a KillRequest with ID "sess-001" and Signal SIGTERM is sent
- Then the session's process group receives SIGTERM
- And the session is removed from the helper's tracking

AC-4: Helper cleans up stale socket on startup
- Given a stale socket file exists at the configured path from a previous crash
- When the helper starts
- Then it removes the stale file and successfully listens on a fresh socket

AC-5: Helper detects parent death and self-terminates
- Given the helper was started with `MASHED_PARENT_PID` set to a specific PID
- When that parent process exits
- Then the helper kills all active sessions
- And removes the socket file
- And exits with code 0

AC-6: Helper handles invalid requests gracefully
- Given the helper is running
- When a malformed message (invalid JSON, unknown type, missing fields) is received
- Then the helper logs the error to stderr
- And continues processing subsequent valid requests (does not crash)

## BDD Test Scenarios

### Scenario 1: Server lifecycle

```gherkin
Feature: PTY helper server lifecycle

  Scenario: Clean startup and READY signal
    Given env MASHED_PTY_SOCK is set to a temporary socket path
    And env MASHED_PARENT_PID is set to the current process PID
    When the server starts via StartServer()
    Then "READY\n" is written to the provided stdout writer
    And the socket file exists on disk
    And a Unix client can connect to the socket

  Scenario: Stale socket cleanup
    Given a file already exists at the configured socket path
    When the server starts
    Then the old file is removed
    And the server successfully listens on a fresh socket
```

### Scenario 2: Spawn and fd delivery

```gherkin
Feature: PTY spawn via helper

  Scenario: Spawn /bin/echo and receive output via fd
    Given a connected client
    When a SpawnRequest is sent with Shell "/bin/echo" Args ["hello"] Cols 80 Rows 24
    Then the SpawnResponse has a non-zero PID and empty Error
    And RecvFd returns a valid fd
    And reading from os.NewFile(fd) yields output containing "hello"

  Scenario: Spawn with invalid shell path
    Given a connected client
    When a SpawnRequest is sent with Shell "/nonexistent/shell"
    Then the SpawnResponse has a non-empty Error field
    And no fd is sent via SCM_RIGHTS
```

### Scenario 3: Kill session

```gherkin
Feature: PTY kill via helper

  Scenario: Kill a running session
    Given a session "sess-kill-test" was spawned with Shell "/bin/cat" (blocks on stdin)
    When a KillRequest is sent with ID "sess-kill-test" and Signal 15
    Then the process is no longer running (kill -0 returns ESRCH)
    And a subsequent SpawnRequest with a new ID succeeds

  Scenario: Kill a non-existent session
    Given no session with ID "ghost-session" exists
    When a KillRequest is sent with ID "ghost-session"
    Then the helper logs a warning but does not crash
```

### Scenario 4: Orphan detection

```gherkin
Feature: Orphan detection

  Scenario: Helper exits when parent dies
    Given the helper is running with MASHED_PARENT_PID set to a subprocess we control
    And a session is active
    When the parent subprocess is killed
    Then the helper detects parent death within 3 seconds
    And all sessions are terminated
    And the socket file is removed
    And the helper process exits
```

## Tasks / Subtasks

- [ ] Task 1: Create helper binary entry point (AC: AC-1, AC-4)
  - [ ] Create `cmd/pty-helper/` directory
  - [ ] Create `cmd/pty-helper/main.go` with env var parsing, signal handling
  - [ ] Call `server.StartServer()` from `internal/terminal/helper/server.go`

- [ ] Task 2: Implement server core (AC: AC-1, AC-2, AC-3, AC-6)
  - [ ] Create `internal/terminal/helper/server.go`
  - [ ] Implement `StartServer(sockPath string, stdout io.Writer) error`
  - [ ] Implement connection accept loop and message dispatch
  - [ ] Implement `handleSpawn` with `pty.StartWithSize` + `SendFd`
  - [ ] Implement `handleKill` with process group signal
  - [ ] Store sessions in `sync.Map`

- [ ] Task 3: Implement orphan detection (AC: AC-5)
  - [ ] Implement `watchParent(ppid int, shutdown func())` goroutine
  - [ ] Try kqueue `EVFILT_PROC`+`NOTE_EXIT` first (instant), fall back to polling `kill(ppid, 0)` every 2s
  - [ ] On parent death: call server shutdown (kill all sessions, remove socket, exit)

- [ ] Task 4: Write tests (AC: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6)
  - [ ] Create `internal/terminal/helper/server_test.go`
  - [ ] Test server startup, READY signal, socket creation
  - [ ] Test spawn + fd round-trip with `/bin/echo`
  - [ ] Test kill with `/bin/cat` (long-running process)
  - [ ] Test stale socket cleanup
  - [ ] Test malformed message handling
  - [ ] Test orphan detection by spawning a child process as fake parent, then killing it

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/terminal/helper/server.go`
- [ ] `go build ./cmd/pty-helper` produces a working binary
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
