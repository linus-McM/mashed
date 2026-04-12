# Story 5: SessionManager Rework -- Use Helper Client

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** Story 4 (ptyhelper-04-client)
**Status:** done

## Description

Rework `SessionManager.Spawn()` to use the PTY helper client instead of calling `pty.StartWithSize` directly. This is the key architectural change: the main Wails process no longer fork/execs anything -- it delegates PTY creation to the helper binary and receives the fd back via SCM_RIGHTS. Resize continues to work locally because the main process owns the PTY master fd. The session, bridge, scrollback buffer, and WebSocket fan-out remain unchanged.

## Developer Notes

### Architecture
- **Modified file:** `internal/terminal/manager.go` -- replace direct `pty.StartWithSize` with `helperClient.Spawn`
- **Modified file:** `app.go` -- `NewApp()` accepts a `*helper.Client` parameter (or nil for graceful degradation)
- The `ManagedSession` struct in `session.go` stays the same -- it already takes a `*os.File` ptmx and an `*exec.Cmd`. The cmd will now be nil (process is owned by helper), so we need a new constructor variant.
- The `Bridge` and `session.go` readLoop/clientReader are UNCHANGED -- they work on `*os.File` regardless of how it was obtained.

### Key Changes to manager.go

**Before (current):**
```go
cmd := exec.CommandContext(ctx, parts[0], parts[1:]...)
ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{...})
ms := newManagedSession(name, cmd, ptmx)
```

**After:**
```go
if sm.helperClient == nil {
    return nil, ErrHelperNotRunning
}
ptmx, pid, err := sm.helperClient.Spawn(ctx, helper.SpawnRequest{
    ID: name, Shell: parts[0], Args: parts[1:],
    Env: append(os.Environ(), "TERM=xterm-256color"),
    Cwd: repoPath, Cols: 80, Rows: 24,
})
ms := newRemoteManagedSession(name, pid, ptmx)
```

### New Sentinel Error
```go
var ErrHelperNotRunning = errors.New("terminal: PTY helper not running")
```

### New Constructor: newRemoteManagedSession
The existing `newManagedSession` expects `*exec.Cmd` (for PID, SysProcAttr check, cmd.Wait). For helper-spawned sessions:
- PID is known (from SpawnResponse) but there's no local `*exec.Cmd`
- Kill goes through the helper client (`client.Kill(id, SIGTERM)`)
- We need a variant that stores only the PID and ptmx, and delegates kill to the helper

```go
func newRemoteManagedSession(name string, pid int, ptmx *os.File) *ManagedSession {
    ms := &ManagedSession{
        name:   name,
        ptmx:   ptmx,
        scroll: newScrollBuffer(scrollBufferDefaultCap),
        clients: make(map[*websocket.Conn]*sync.Mutex),
        done:   make(chan struct{}),
        pid:    pid,
        // cmd is nil for remote sessions
    }
    go ms.readLoop()
    return ms
}
```

### SessionManager Changes
```go
type SessionManager struct {
    mu           sync.Mutex
    sessions     map[string]*ManagedSession
    helperClient *helper.Client  // nil = helper unavailable
}

func NewSessionManager(client *helper.Client) *SessionManager {
    return &SessionManager{
        sessions:     make(map[string]*ManagedSession),
        helperClient: client,
    }
}
```

### Kill Changes
`ManagedSession.Kill()` currently sends `syscall.Kill(-pid, SIGTERM)` locally. For remote sessions (cmd == nil), it should call `sm.helperClient.Kill(name, SIGTERM)` instead. Two approaches:
1. Store a kill function `killFn func()` on `ManagedSession` (set during construction)
2. Check `ms.cmd == nil` and delegate to helper client

Option 1 is cleaner -- the session knows how to kill itself regardless of backend.

### Resize
No changes needed. `session.go:clientReader` already calls `pty.Setsize(ms.ptmx, ...)` directly on the fd. Since the main process owns the PTY master fd (received via SCM_RIGHTS), this works as-is.

### Graceful Degradation
When `helperClient` is nil:
- `Spawn()` returns `ErrHelperNotRunning`
- `Kill()` still works for existing sessions (if any were somehow created)
- The frontend receives the error and shows it to the user

### Technical Considerations
- The `sessionManager` interface in `app.go` stays the same -- `Spawn`, `Kill`, `IsAlive`, `FindByPID`, `Shutdown`
- `FindByPID` still works because we store the PID from SpawnResponse
- `Shutdown` should call `helperClient.Kill` for each session, then `helperClient.Close`
- The `cmd` field on ManagedSession becomes optional (nil for remote sessions)
- Update `Kill()` to handle nil cmd gracefully (don't call `cmd.Wait()` if cmd is nil)

### Risks & Edge Cases
- `readLoop` calls `ms.cmd.Wait()` on EOF -- must be guarded with nil check for remote sessions
- If the helper crashes, the PTY fd becomes invalid and readLoop gets EOF -- this triggers normal session cleanup
- The existing `newManagedSession` constructor has a log warning for missing `SysProcAttr.Setpgid` -- skip this for remote sessions

### Reference Files
- `internal/terminal/manager.go` -- the file being modified
- `internal/terminal/session.go` -- `newManagedSession`, `Kill`, `readLoop` -- modified for remote sessions
- `internal/terminal/helper/client.go` -- the client being used (Story 4)
- `app.go:NewApp()` -- modified to accept helper client
- `app.go:sessionManager` interface -- verify compatibility

## Acceptance Criteria

AC-1: SessionManager.Spawn delegates to helper client
- Given a SessionManager initialized with a valid helper Client
- When Spawn is called with a session name and repo path
- Then the helper client's Spawn method is invoked
- And the returned ManagedSession has a valid PTY fd and PID
- And no local fork/exec occurs in the main process

AC-2: ManagedSession works with remote PTY fd
- Given a ManagedSession created via the helper
- When a WebSocket client connects
- Then the scrollback buffer works (readLoop reads from PTY fd)
- And client input is written to the PTY fd
- And resize works via pty.Setsize on the local fd

AC-3: Graceful degradation when helper is unavailable
- Given a SessionManager initialized with nil helper client
- When Spawn is called
- Then it returns `ErrHelperNotRunning`
- And the error message is "terminal: PTY helper not running"

AC-4: Kill works for remote sessions
- Given a remote ManagedSession exists
- When Kill is called
- Then the helper's Kill method is invoked with the correct session ID
- And the PTY fd is closed locally
- And the session is removed from the manager

AC-5: Shutdown terminates all sessions and closes client
- Given multiple remote sessions exist
- When Shutdown is called
- Then all sessions are killed via the helper client
- And the helper client connection is closed

## BDD Test Scenarios

### Scenario 1: Remote spawn

```gherkin
Feature: SessionManager spawn via helper

  Scenario: Spawn creates remote session
    Given a SessionManager with a mock helper client
    When Spawn is called with name "test-session" and repoPath "/tmp/test"
    Then the helper client receives a SpawnRequest with ID "test-session"
    And the returned ManagedSession has a PID > 0
    And the ManagedSession.Name() returns "test-session"
    And the ManagedSession is stored in the sessions map

  Scenario: Duplicate session name rejected
    Given a session "dup-test" already exists
    When Spawn is called with name "dup-test"
    Then it returns ErrSessionExists
    And no SpawnRequest is sent to the helper
```

### Scenario 2: Graceful degradation

```gherkin
Feature: Graceful degradation without helper

  Scenario: Spawn fails gracefully when helper is nil
    Given a SessionManager initialized with nil helper client
    When Spawn is called
    Then it returns ErrHelperNotRunning
    And no panic occurs

  Scenario: Other operations work without helper
    Given a SessionManager initialized with nil helper client
    When List is called
    Then it returns an empty list
    When IsAlive is called with any name
    Then it returns false
```

### Scenario 3: Kill and shutdown

```gherkin
Feature: Session kill and shutdown

  Scenario: Kill remote session
    Given a remote session "kill-me" exists
    When Kill("kill-me") is called
    Then the helper client receives a KillRequest with ID "kill-me"
    And the session is removed from the manager
    And IsAlive("kill-me") returns false

  Scenario: Shutdown kills all sessions
    Given 3 remote sessions exist
    When Shutdown is called
    Then the helper client receives 3 KillRequests
    And the sessions map is empty
```

## Tasks / Subtasks

- [ ] Task 1: Add helper client to SessionManager (AC: AC-1, AC-3)
  - [ ] Add `helperClient *helper.Client` field to `SessionManager` struct
  - [ ] Update `NewSessionManager(client *helper.Client)` constructor
  - [ ] Add `ErrHelperNotRunning` sentinel error
  - [ ] Guard `Spawn` with nil check on `helperClient`

- [ ] Task 2: Implement remote session construction (AC: AC-1, AC-2)
  - [ ] Add `newRemoteManagedSession(name string, pid int, ptmx *os.File) *ManagedSession`
  - [ ] Add `killFn func()` field to ManagedSession or guard `Kill`/`readLoop` for nil cmd
  - [ ] Rewrite `Spawn` to use `helperClient.Spawn` + `newRemoteManagedSession`
  - [ ] Ensure readLoop, scrollBuffer, WebSocket fan-out all work with remote fd

- [ ] Task 3: Update Kill and Shutdown for remote sessions (AC: AC-4, AC-5)
  - [ ] Update `ManagedSession.Kill()` to handle nil cmd (no cmd.Wait for remote sessions)
  - [ ] Update `SessionManager.Kill()` to delegate to helper client
  - [ ] Update `SessionManager.Shutdown()` to kill all sessions and close client

- [ ] Task 4: Update app.go integration points (AC: AC-1, AC-3)
  - [ ] Update `NewApp()` to accept `*helper.Client` parameter
  - [ ] Pass client through to `NewSessionManager`
  - [ ] Update `sessionManager` interface if needed (it should NOT need changes)

- [ ] Task 5: Write tests (AC: AC-1, AC-2, AC-3, AC-4, AC-5)
  - [ ] Create mock helper client for unit tests
  - [ ] Test Spawn with mock client (verify request sent, session created)
  - [ ] Test graceful degradation (nil client)
  - [ ] Test Kill for remote sessions
  - [ ] Test Shutdown
  - [ ] Verify existing manager_test.go tests still pass (or update them)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on modified `manager.go` and `session.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Existing `bridge_test.go` and `session_test.go` still pass
