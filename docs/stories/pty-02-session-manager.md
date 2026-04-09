# Story 2: SessionManager Lifecycle Manager

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** pty-01
**Status:** ready

## Description

Implement the `SessionManager` that owns a map of named `ManagedSession` instances and provides Spawn/Get/Kill/List/FindByPID/Shutdown operations. This is the central registry that replaces tmux session management for all spawned terminals and agents. The bridge, app spawn methods, and scan loop all depend on this manager.

## Developer Notes

### Architecture
- **New file:** `internal/terminal/manager.go`
- **Package:** `terminal` (same package as `session.go`, `bridge.go`, `panes.go`)
- **Key types to create:**
  - `SessionManager` struct: `mu sync.Mutex`, `sessions map[string]*ManagedSession`, `ctx context.Context`
  - Constructor: `NewSessionManager() *SessionManager`
- **Methods:**
  - `Spawn(ctx context.Context, name string, repoPath string, command string) (*ManagedSession, error)` -- creates `exec.Command`, sets working dir, sets `SysProcAttr{Setpgid: true}`, starts via `pty.StartWithSize`, constructs `ManagedSession`, stores in map, starts `readLoop`
  - `Get(name string) (*ManagedSession, bool)` -- thread-safe lookup
  - `Kill(name string) error` -- calls `session.Kill()`, removes from map
  - `List(filter func(*ManagedSession) bool) []*ManagedSession` -- returns filtered snapshot
  - `IsAlive(name string) bool` -- checks if session exists and `session.IsAlive()`
  - `FindByPID(pid int) (*ManagedSession, bool)` -- iterates sessions, checks `session.cmd.Process.Pid`
  - `Shutdown()` -- kills all sessions, clears map
- **Shell detection in Spawn:** When `command` is empty, use `os.Getenv("SHELL")`, fallback to `/bin/zsh`
- **Command parsing:** When `command` is non-empty, split into executable + args. Use `exec.CommandContext(ctx, parts[0], parts[1:]...)`. Set `cmd.Dir = repoPath`, `cmd.Env = append(os.Environ(), "TERM=xterm-256color")`
- **Initial PTY size:** Start with `pty.Winsize{Cols: 80, Rows: 24}` (reasonable default; frontend sends real size immediately)

### Technical Considerations
- **Concurrency:** All map operations protected by `mu`. `Spawn` holds lock only for map insertion, not during PTY start (to avoid blocking other operations during slow process creation).
- **Error handling:** Return `TerminalError` for spawn failures. `ErrSessionDead` if `Kill` called on non-existent session. New sentinel: `ErrSessionExists = errors.New("terminal: session already exists")` for duplicate name.
- **Spawn atomicity:** Check for duplicate name under lock before starting process. If PTY start fails after check, clean up without leaving stale map entry.
- **Shutdown ordering:** Kill sessions in parallel (each `Kill` is independent), use `sync.WaitGroup` to wait for all.

### Risks & Edge Cases
- **Duplicate session names:** `Spawn` must reject duplicate names with `ErrSessionExists`.
- **Spawn with empty repoPath:** Should still work (process starts in current dir). Log a warning.
- **Kill non-existent session:** Return error but do not panic.
- **FindByPID after process exit:** `cmd.Process.Pid` is still set after exit; `IsAlive()` check prevents returning dead sessions.
- **Command parsing edge cases:** Commands with quoted arguments (e.g., `claude --model "opus"`) -- use `strings.Fields` for basic splitting. Document limitation.

### Reference Files
- `internal/terminal/session.go` (from Story 1) -- `ManagedSession`, `newManagedSession`
- `internal/terminal/panes.go` -- `TerminalError` struct, sentinel error patterns
- `internal/terminal/bridge.go` -- `ptyReadSize` constant, `resizeMsg` type
- `app_tmux.go` (lines 19-52) -- current spawn logic to understand naming conventions and session registration

## Acceptance Criteria

AC-1: Spawn creates a PTY session with correct process setup
- Given a `SessionManager` instance
- When `Spawn(ctx, "test-session", "/tmp", "")` is called
- Then a new `ManagedSession` is created running the user's default shell
- And the session's working directory is `/tmp`
- And `TERM=xterm-256color` is in the process environment
- And the process has its own process group (`Setpgid: true`)

AC-2: Spawn rejects duplicate session names
- Given a `SessionManager` with an existing session named "my-session"
- When `Spawn(ctx, "my-session", "/tmp", "")` is called
- Then the call returns `ErrSessionExists`
- And the existing session is not affected

AC-3: Get, Kill, and IsAlive provide correct session lifecycle
- Given a `SessionManager` with a session named "sess1"
- When `Get("sess1")` is called, it returns the session and true
- When `Kill("sess1")` is called, the session process is terminated
- Then `Get("sess1")` returns nil and false
- And `IsAlive("sess1")` returns false

AC-4: FindByPID locates sessions by process ID
- Given a `SessionManager` with a session whose PID is 12345
- When `FindByPID(12345)` is called
- Then the matching session is returned
- And `FindByPID(99999)` returns nil and false

AC-5: Shutdown kills all sessions
- Given a `SessionManager` with 3 active sessions
- When `Shutdown()` is called
- Then all 3 sessions have `IsAlive() == false`
- And the session map is empty

AC-6: Spawn with command string starts the specified process
- Given a `SessionManager`
- When `Spawn(ctx, "agent1", "/tmp", "echo hello world")` is called
- Then the session runs `echo` with args `["hello", "world"]`
- And the session's PTY receives the command output

## BDD Test Scenarios

### Scenario 1: Session Lifecycle

```gherkin
Feature: Session Manager lifecycle

  Scenario: Spawn and retrieve a session
    Given a new SessionManager
    When I spawn a session named "test" in directory "/tmp" with no command
    Then Get("test") returns the session
    And IsAlive("test") returns true
    And the session's process is running

  Scenario: Kill removes session from manager
    Given a SessionManager with session "test"
    When I call Kill("test")
    Then Get("test") returns false
    And IsAlive("test") returns false

  Scenario: Kill non-existent session returns error
    Given a SessionManager with no sessions
    When I call Kill("ghost")
    Then an error is returned
```

### Scenario 2: Duplicate Prevention

```gherkin
Feature: Duplicate session names

  Scenario: Reject duplicate name
    Given a SessionManager with session "dup"
    When I spawn a session named "dup"
    Then ErrSessionExists is returned
    And the original session is still alive
```

### Scenario 3: Shell Detection

```gherkin
Feature: Default shell detection

  Scenario: Empty command uses SHELL env
    Given SHELL environment variable is "/bin/zsh"
    When I spawn a session with empty command
    Then the session runs "/bin/zsh"

  Scenario: Empty command with no SHELL env falls back
    Given SHELL environment variable is unset
    When I spawn a session with empty command
    Then the session runs "/bin/zsh"
```

### Scenario 4: Shutdown

```gherkin
Feature: Manager shutdown

  Scenario: Shutdown kills all sessions
    Given a SessionManager with sessions "a", "b", "c"
    When Shutdown is called
    Then IsAlive returns false for all three
    And List returns empty slice
```

### Scenario 5: FindByPID

```gherkin
Feature: Find session by PID

  Scenario: Find existing session by PID
    Given a SessionManager with a session whose process has PID P
    When FindByPID(P) is called
    Then the session is returned

  Scenario: FindByPID returns nil for unknown PID
    Given a SessionManager with sessions
    When FindByPID(99999) is called
    Then nil and false are returned
```

## Tasks / Subtasks

- [ ] Task 1: Implement `SessionManager` struct and constructor (AC: AC-1)
  - [ ] Subtask 1a: Define `SessionManager` struct with `mu`, `sessions` map
  - [ ] Subtask 1b: Implement `NewSessionManager()` constructor
  - [ ] Subtask 1c: Add sentinel errors `ErrSessionExists`, `ErrSessionNotFound`

- [ ] Task 2: Implement `Spawn` method (AC: AC-1, AC-2, AC-6)
  - [ ] Subtask 2a: Shell detection logic (`$SHELL` env, `/bin/zsh` fallback)
  - [ ] Subtask 2b: Command parsing with `strings.Fields`
  - [ ] Subtask 2c: `exec.CommandContext` setup with `Dir`, `Env`, `SysProcAttr`
  - [ ] Subtask 2d: PTY start via `pty.StartWithSize`, construct `ManagedSession`, store in map, start `readLoop`
  - [ ] Subtask 2e: Duplicate name check under lock; cleanup on PTY start failure

- [ ] Task 3: Implement query methods (AC: AC-3, AC-4)
  - [ ] Subtask 3a: `Get(name)` -- lock, lookup, unlock
  - [ ] Subtask 3b: `IsAlive(name)` -- lock, lookup + `session.IsAlive()`, unlock
  - [ ] Subtask 3c: `FindByPID(pid)` -- lock, iterate, match `cmd.Process.Pid`, unlock
  - [ ] Subtask 3d: `List(filter)` -- lock, iterate with filter, return snapshot

- [ ] Task 4: Implement `Kill` and `Shutdown` (AC: AC-3, AC-5)
  - [ ] Subtask 4a: `Kill(name)` -- lock, lookup, call `session.Kill()`, delete from map
  - [ ] Subtask 4b: `Shutdown()` -- lock, collect all sessions, unlock, kill each, clear map
  - [ ] Subtask 4c: Write tests for concurrent Kill and Shutdown

- [ ] Task 5: Write comprehensive tests (AC: all)
  - [ ] Subtask 5a: Table-driven tests for Spawn (valid, duplicate, empty command, with command)
  - [ ] Subtask 5b: Integration test: spawn real shell, verify PID lookup, kill, verify dead
  - [ ] Subtask 5c: Race condition test: concurrent Spawn + Kill + Get

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/terminal/manager.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
