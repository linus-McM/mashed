# Story 1: ManagedSession with Scrollback Ring Buffer

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Implement the `ManagedSession` type that owns a PTY file descriptor, an `exec.Cmd`, a 1MB scrollback ring buffer, and a set of connected WebSocket clients. This is the foundational building block: every subsequent story depends on `ManagedSession` for PTY lifecycle management. The session reads PTY output in a goroutine, appends to scrollback, and fans data out to all connected WebSocket clients with per-client write mutexes.

## Developer Notes

### Architecture
- **New file:** `internal/terminal/session.go`
- **Package:** `terminal` (same package as `bridge.go` and `panes.go`)
- **Key types to create:**
  - `scrollBuffer` -- circular byte buffer with `Write([]byte)` and `Snapshot() []byte`, 1MB max (`1 << 20`)
  - `ManagedSession` -- holds: `cmd *exec.Cmd`, `ptmx *os.File`, `scroll *scrollBuffer`, `clients map[*websocket.Conn]*sync.Mutex`, `mu sync.Mutex`, `done chan struct{}`, `name string`, `pid int`
  - `wsClient` -- optional thin wrapper: `conn *websocket.Conn`, `mu sync.Mutex`
- **Data flow:** `readLoop()` goroutine reads PTY in 4KB chunks -> `scroll.Write(chunk)` -> iterates `clients` map, calls `client.mu.Lock(); client.conn.WriteMessage(BinaryMessage, chunk); client.mu.Unlock()` -- failure on a client removes that client only
- **`AddClient(ws *websocket.Conn)`**: acquires session `mu`, sends `scroll.Snapshot()` to `ws`, registers `ws` in `clients` map, releases `mu`, then starts a `clientReader` goroutine that reads WS input and writes to `ptmx` (or handles resize JSON)
- **`RemoveClient(ws)`**: acquires `mu`, deletes from `clients`, closes `ws`
- **`Kill()`**: sends `syscall.Kill(-pid, syscall.SIGTERM)` for process group kill, closes `ptmx`, closes `done` channel
- **`IsAlive()`**: checks if `done` channel is closed (non-blocking select)
- **`Wait()`**: blocks on `done` channel

### Technical Considerations
- **Concurrency:** `readLoop` holds session `mu` only for the duration of iterating clients + writing scrollback. Per-client `sync.Mutex` on `WriteMessage` prevents interleaving. Snapshot+register in `AddClient` is atomic under session `mu` to prevent a gap where data arrives between snapshot send and client registration.
- **Error handling:** Use `TerminalError` from `panes.go` for structured errors (it already exists in the package). Add sentinel `ErrSessionDead = errors.New("terminal: session is not alive")`.
- **Process group kill:** Set `cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}` before starting so `Kill(-pid, SIGTERM)` catches child processes.
- **Slow client:** Use `ws.SetWriteDeadline(time.Now().Add(5*time.Second))` before each write. If write fails or times out, remove client but do NOT kill the session.
- **PTY read size:** 4096 bytes per read (same constant `ptyReadSize` already defined in `bridge.go`).

### Risks & Edge Cases
- **Race on AddClient during readLoop fan-out:** Must hold `mu` during both snapshot send and client insertion to avoid missing data.
- **PTY EOF:** When process exits, `ptmx.Read` returns EOF. `readLoop` should close `done` channel and clean up.
- **Kill after process already exited:** `syscall.Kill(-pid, SIGTERM)` returns ESRCH -- handle gracefully.
- **Multiple Kill() calls:** Use `sync.Once` for `done` channel close to avoid panic.
- **Client disconnect during AddClient snapshot send:** If `WriteMessage` fails during snapshot, do not register the client.

### Reference Files
- `internal/terminal/bridge.go` -- existing `outputBuf` type (lines 292-330), `resizeMsg` type (line 40-44), `ptyReadSize` constant, `TerminalError` usage
- `internal/terminal/panes.go` -- `TerminalError` struct definition (lines 22-30), sentinel error patterns
- `go.mod` -- confirms `github.com/creack/pty v1.1.24` and `github.com/gorilla/websocket` are available

## Acceptance Criteria

AC-1: Scrollback ring buffer stores and replays data correctly
- Given a `scrollBuffer` with 1MB capacity
- When 2MB of data is written in sequential chunks
- Then `Snapshot()` returns exactly the most recent 1MB of data
- And the returned slice is a copy (mutating it does not affect the buffer)

AC-2: ManagedSession fans PTY output to multiple connected clients
- Given a `ManagedSession` with an active PTY process
- When two WebSocket clients are connected via `AddClient`
- Then both clients receive all PTY output written after their connection
- And each client receives output independently (one slow client does not block the other)

AC-3: AddClient sends scrollback snapshot atomically before streaming
- Given a `ManagedSession` that has already produced 10KB of output
- When a new client connects via `AddClient`
- Then the client first receives the 10KB scrollback snapshot
- And subsequent PTY output follows with no gap or duplication

AC-4: Kill terminates the process group and cleans up resources
- Given a `ManagedSession` running `/bin/sh -c "sleep 300"`
- When `Kill()` is called
- Then the shell process and its children receive SIGTERM
- And `IsAlive()` returns false
- And the PTY file descriptor is closed

AC-5: Client failure does not affect other clients or kill the session
- Given a `ManagedSession` with two connected clients
- When one client's WebSocket connection is forcibly closed
- Then the other client continues receiving output
- And `IsAlive()` still returns true

## BDD Test Scenarios

### Scenario 1: Ring Buffer Capacity

```gherkin
Feature: Scrollback ring buffer

  Scenario: Buffer wraps when capacity exceeded
    Given a scrollBuffer with capacity 1024 bytes
    When I write 1500 bytes of sequential data
    Then Snapshot returns the last 1024 bytes
    And the returned slice length is 1024

  Scenario: Buffer returns all data when under capacity
    Given a scrollBuffer with capacity 1024 bytes
    When I write 500 bytes of data
    Then Snapshot returns exactly those 500 bytes

  Scenario: Snapshot returns independent copy
    Given a scrollBuffer containing "hello"
    When I call Snapshot and modify the returned slice
    Then a second Snapshot call returns the original "hello"
```

### Scenario 2: Multi-Client Fan-Out

```gherkin
Feature: PTY output fan-out

  Scenario: Two clients receive same output
    Given a ManagedSession running "echo hello"
    And client A is connected via AddClient
    And client B is connected via AddClient
    When the PTY produces output "hello\n"
    Then client A receives "hello\n"
    And client B receives "hello\n"

  Scenario: Failed client is removed without killing session
    Given a ManagedSession with two clients
    When client A's connection returns a write error
    Then client A is removed from the session
    And client B continues to receive output
    And IsAlive returns true
```

### Scenario 3: Atomic Scrollback Replay

```gherkin
Feature: Scrollback replay on connect

  Scenario: New client gets scrollback then live data
    Given a ManagedSession that has produced "line1\nline2\n"
    When a new client connects via AddClient
    Then the client receives "line1\nline2\n" as the first message
    And subsequent PTY output is received in order after the snapshot
```

### Scenario 4: Process Group Kill

```gherkin
Feature: Session kill

  Scenario: Kill terminates process group
    Given a ManagedSession running "/bin/sh -c 'sleep 300'"
    When Kill is called
    Then IsAlive returns false
    And the shell process is no longer running
    And calling Kill again does not panic

  Scenario: Session cleanup after process exits naturally
    Given a ManagedSession running "echo done"
    When the process exits on its own
    Then IsAlive returns false after readLoop detects EOF
```

## Tasks / Subtasks

- [x] Task 1: Implement `scrollBuffer` type (AC: AC-1)
  - [x] Subtask 1a: Define `scrollBuffer` struct with `buf []byte`, `size int`, `cap int` fields and `New(capacity int) *scrollBuffer` constructor
  - [x] Subtask 1b: Implement `Write(p []byte)` that appends and trims oldest bytes when over capacity
  - [x] Subtask 1c: Implement `Snapshot() []byte` that returns a copy of buffered data
  - [x] Subtask 1d: Write table-driven tests for Write/Snapshot with edge cases (empty, exact capacity, over capacity, multiple writes)

- [x] Task 2: Implement `ManagedSession` core struct and lifecycle (AC: AC-2, AC-4, AC-5)
  - [x] Subtask 2a: Define `ManagedSession` struct with all fields; constructor `newManagedSession(name string, cmd *exec.Cmd, ptmx *os.File) *ManagedSession`
  - [x] Subtask 2b: Implement `readLoop()` goroutine: read PTY -> write scrollback -> fan out to clients -> handle EOF by closing `done`
  - [x] Subtask 2c: Implement `Kill()` with `sync.Once`, process group SIGTERM, PTY close, done channel close
  - [x] Subtask 2d: Implement `IsAlive()` via non-blocking select on `done` channel

- [x] Task 3: Implement `AddClient` / `RemoveClient` with atomic scrollback replay (AC: AC-3, AC-5)
  - [x] Subtask 3a: Implement `AddClient(ws *websocket.Conn)` -- lock, send snapshot, register, unlock, start clientReader goroutine
  - [x] Subtask 3b: Implement `RemoveClient(ws)` -- lock, delete from map, close ws
  - [x] Subtask 3c: Implement `clientReader(ws)` -- read WS messages, handle binary (write to PTY) and text (resize JSON), remove client on error
  - [x] Subtask 3d: Write integration tests using `creack/pty` with a real shell process

- [x] Task 4: Add sentinel errors and integrate with existing error types (AC: AC-4)
  - [x] Subtask 4a: Add `ErrSessionDead` sentinel to `session.go`
  - [x] Subtask 4b: Use `TerminalError` wrapper for operation failures (reuse existing type from `panes.go`)

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ code coverage on `internal/terminal/session.go`
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes (critical for concurrency correctness)
- [x] `/simplify` run on all modified code
- [x] Code review: no CRITICAL/HIGH issues
