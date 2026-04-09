# Story 3: Bridge Refactor -- Route WebSockets to Managed Sessions

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** L
**Depends On:** pty-01, pty-02
**Status:** ready

## Description

Refactor `internal/terminal/bridge.go` to route incoming WebSocket connections to `ManagedSession` instances via the `SessionManager`, instead of spawning `tmux attach-session` processes. The bridge becomes a thin HTTP/WS server that extracts session names from URL paths and delegates to `session.AddClient()`. All tmux-specific code (`tmux attach-session`, `tmux set-option`, `tmux capture-pane`, `servePane`) and the `outputBuf` type are removed.

## Developer Notes

### Architecture
- **Modify file:** `internal/terminal/bridge.go` (currently 331 lines)
- **Changes to `Bridge` struct:**
  - Add field: `manager *SessionManager`
  - Change constructor: `NewBridge(manager *SessionManager) *Bridge`
  - Remove: `conns map[*websocket.Conn]context.CancelFunc` (client tracking moves to `ManagedSession`)
- **`handleWS` changes:**
  - Extract session name from URL path `/ws/{name}` (same path pattern, but `name` is now a session name, not a tmux pane target)
  - Call `b.manager.Get(name)` -- if not found, return 404
  - Call `session.AddClient(ws)` -- this blocks until client disconnects or session dies
  - Remove: `connCtx`/`connCancel` per-connection context (session manages client lifecycle)
- **Remove entirely:**
  - `servePane()` method (162-284) -- replaced by `session.AddClient`
  - `outputBuf` type and related functions (292-330) -- replaced by `scrollBuffer` in session.go
  - `trackConn()` / `untrackConn()` methods -- session manages clients
  - tmux exec calls (lines 174, 184-196, 198)
  - `ErrPaneTargetEmpty` sentinel (no longer relevant)
- **Keep:**
  - `Start()` / `Stop()` / `shutdown()` lifecycle methods
  - `GetTerminalPort()` method
  - `wsUpgrader` variable
  - `resizeMsg` type (used by session's clientReader)
  - `sendWSClose` helper
  - `ErrBridgeClosed` sentinel
  - `ptyReadSize` constant (used by session.go)

### Technical Considerations
- **Simplified connection model:** Bridge no longer manages per-connection goroutines or contexts. `handleWS` calls `session.AddClient(ws)` which runs its own clientReader goroutine. The WS connection lifecycle is owned by the session.
- **shutdown() simplification:** Instead of iterating `conns` map and canceling each, `shutdown()` can call `b.manager.Shutdown()` or just close the HTTP server (sessions are killed by the app's `manager.Shutdown()` call).
- **URL path handling:** Keep the `/ws/` prefix pattern. The value after `/ws/` is now a session name (e.g., `mashed-myrepo-1712600000`) instead of a tmux target (e.g., `mashed-myrepo-1712600000:0.0`).
- **Error responses:** Return HTTP 404 for unknown session names, HTTP 400 for empty names.

### Risks & Edge Cases
- **Bridge-session lifecycle mismatch:** If `Shutdown()` is called on the manager while a WS client is still connected, `AddClient`'s clientReader should detect PTY closure and exit cleanly.
- **Concurrent handleWS for same session:** Multiple WS connections to the same session is valid (multi-tab). `AddClient` is thread-safe by design.
- **URL-encoded session names:** Session names contain hyphens and numbers, no special encoding needed. But keep `strings.TrimPrefix` for safety.
- **Compilation after removal:** Ensure `bridge.go` still compiles without the removed types. `resizeMsg` must remain since `session.go` needs it (same package, so it's accessible).

### Reference Files
- `internal/terminal/bridge.go` -- current file (331 lines, full content read above)
- `internal/terminal/session.go` (from Story 1) -- `ManagedSession.AddClient()`
- `internal/terminal/manager.go` (from Story 2) -- `SessionManager.Get()`
- `app.go` line 110 -- `terminal.NewBridge()` call site (must update to `NewBridge(manager)`)

## Acceptance Criteria

AC-1: Bridge constructor accepts a SessionManager
- Given the refactored `Bridge`
- When `NewBridge(manager)` is called with a valid `SessionManager`
- Then the bridge stores the manager reference
- And the bridge compiles and initializes correctly

AC-2: WebSocket connections route to managed sessions
- Given a bridge with a running SessionManager that has a session "test-sess"
- When a WebSocket connects to `/ws/test-sess`
- Then the connection is passed to the session via `AddClient`
- And the client receives PTY output from the session

AC-3: Unknown session names return HTTP 404
- Given a bridge with no session named "ghost"
- When a WebSocket connects to `/ws/ghost`
- Then the server responds with HTTP 404
- And no WebSocket upgrade occurs

AC-4: All tmux code is removed from bridge.go
- Given the refactored `bridge.go` file
- When searched for "tmux" string references
- Then no tmux-related code exists (no exec.Command("tmux", ...), no capture-pane, no attach-session)
- And the `outputBuf` type is removed
- And the `servePane` method is removed

AC-5: Multiple clients can connect to the same session
- Given a session "shared" running in the manager
- When two WebSocket clients connect to `/ws/shared`
- Then both receive PTY output
- And input from either client is written to the PTY

## BDD Test Scenarios

### Scenario 1: Routing

```gherkin
Feature: WebSocket session routing

  Scenario: Connect to existing session
    Given a SessionManager with session "my-session"
    And a Bridge started with that manager
    When a WebSocket connects to "/ws/my-session"
    Then the connection is established successfully
    And the client receives the session's scrollback snapshot

  Scenario: Connect to non-existent session
    Given a SessionManager with no sessions
    And a Bridge started with that manager
    When an HTTP request is made to "/ws/missing"
    Then the response status is 404
```

### Scenario 2: tmux Removal

```gherkin
Feature: tmux code removal

  Scenario: No tmux references in bridge
    Given the refactored bridge.go
    When the file is searched for "tmux" and "attach-session"
    Then zero matches are found

  Scenario: outputBuf type removed
    Given the refactored bridge.go
    When the file is searched for "outputBuf"
    Then zero matches are found
```

### Scenario 3: Multi-Client

```gherkin
Feature: Multiple WebSocket clients

  Scenario: Two clients share a session
    Given a session "shared" running "cat" (echoes input)
    And client A connects to "/ws/shared"
    And client B connects to "/ws/shared"
    When client A sends "hello"
    Then client B receives "hello" in the PTY output
    And client A also receives the echo
```

### Scenario 4: Bridge Lifecycle

```gherkin
Feature: Bridge start and stop

  Scenario: Bridge starts and serves connections
    Given a SessionManager
    When Bridge.Start(ctx) is called
    Then GetTerminalPort returns a non-zero port
    And WebSocket connections are accepted

  Scenario: Bridge stop closes server
    Given a running Bridge
    When Bridge.Stop() is called
    Then the HTTP server stops accepting connections
```

## Tasks / Subtasks

- [ ] Task 1: Update `Bridge` struct and constructor (AC: AC-1)
  - [ ] Subtask 1a: Add `manager *SessionManager` field to `Bridge`
  - [ ] Subtask 1b: Change `NewBridge()` signature to `NewBridge(manager *SessionManager) *Bridge`
  - [ ] Subtask 1c: Remove `conns` map field and related `trackConn`/`untrackConn` methods

- [ ] Task 2: Rewrite `handleWS` to use SessionManager (AC: AC-2, AC-3, AC-5)
  - [ ] Subtask 2a: Extract session name from `/ws/{name}` path
  - [ ] Subtask 2b: Call `manager.Get(name)`, return 404 if not found
  - [ ] Subtask 2c: Upgrade WebSocket, call `session.AddClient(ws)`
  - [ ] Subtask 2d: Remove old pane target extraction and query parameter fallback

- [ ] Task 3: Remove all tmux-specific code (AC: AC-4)
  - [ ] Subtask 3a: Delete `servePane()` method entirely
  - [ ] Subtask 3b: Delete `outputBuf` type, `newOutputBuf`, `Append`, `Drain` methods
  - [ ] Subtask 3c: Delete `ErrPaneTargetEmpty` sentinel
  - [ ] Subtask 3d: Remove tmux-related imports (`bytes`, unused `exec`, `os`)
  - [ ] Subtask 3e: Simplify `shutdown()` -- remove per-connection cleanup (sessions own clients)

- [ ] Task 4: Write tests for refactored bridge (AC: AC-2, AC-3, AC-5)
  - [ ] Subtask 4a: Test: WS connects to valid session, receives data
  - [ ] Subtask 4b: Test: WS to unknown session gets 404
  - [ ] Subtask 4c: Test: Two WS clients to same session both receive output
  - [ ] Subtask 4d: Test: Bridge start/stop lifecycle

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/terminal/bridge.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
