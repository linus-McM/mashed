# Story bridge-03: Bridge Integration — Route BMAD Sessions Through TmuxAdapter

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** bridge-02
**Status:** ready

## Description

Wire the `TmuxAdapter` from bridge-02 into the WebSocket `Bridge` so that `ws://127.0.0.1:{port}/ws/{name}` falls through to live tmux attach whenever the name has a `bmad-` prefix and is not found in the PTY-managed `SessionManager`. This fixes the `[connection error]` bug on "View Terminal" for running BMAD nodes without disturbing the existing PTY flow for shell/agent sessions. A new `proxyTmuxSession` helper mirrors the semantics of `session.AddClient` for the tmux path.

## Developer Notes

### Architecture

- **Modify file:** `internal/terminal/bridge.go`
  - Add field `tmuxAdapter *TmuxAdapter` to the `Bridge` struct (nil-safe: a nil adapter means "tmux path disabled")
  - Change constructor signature to `NewBridge(manager *SessionManager, tmuxAdapter *TmuxAdapter) *Bridge`
  - Update `handleWS`:
    1. Extract name from `/ws/{name}` as today
    2. Try `b.manager.Get(name)` first — if hit, call `session.AddClient(ws)` (unchanged path)
    3. On miss, check `b.tmuxAdapter != nil && strings.HasPrefix(name, bmad.SessionNamePrefix)`
    4. If yes, upgrade WS and call `b.proxyTmuxSession(ctx, ws, name)`
    5. Otherwise return HTTP 404 as today
  - Add new method `proxyTmuxSession(ctx context.Context, ws *websocket.Conn, sessionName string) error`
    - Constructs `paneTarget := sessionName + ":0.0"` (BMAD always uses window 0, pane 0)
    - Calls `b.tmuxAdapter.Attach(ctx, paneTarget)` — if error, send WS close frame with reason and return
    - Spawns two goroutines:
      - **Reader → WS:** read bytes from `attachment.Read`, write as WS `BinaryMessage` frames, exit on `io.EOF`
      - **WS → input/resize:** read WS messages; text messages matching `resizeMsg` JSON call `attachment.Resize`; binary messages call `attachment.SendInput` (treat raw bytes as literal input)
    - Waits for either goroutine to exit, cancels context, calls `attachment.Close()`, closes the WebSocket
  - **Do not import `internal/bmad`** — this would risk a cycle. Instead, define a local constant `bmadSessionPrefix = "bmad-"` in `bridge.go` and comment that it must match `internal/bmad.SessionNamePrefix`. Add a compile-time test in `bridge_test.go` that imports both packages and asserts the constants match.

- **Modify file:** `internal/terminal/bridge_test.go`
  - Add tests exercising the new tmux path with a mock `TmuxAdapter` (define a small interface so the test doesn't need real tmux)
  - **Consider introducing an interface:** `type tmuxAttacher interface { Attach(ctx context.Context, paneTarget string) (TmuxReadWriteCloser, error) }`. If that conflicts with bridge-02's concrete return type, the simpler path is to keep `*TmuxAdapter` as the field type and use a build-tag or interface wrapper. **Recommendation:** define `TmuxAttachmentIface` interface in bridge-02's adapter file so both packages can share it; this story only consumes it.

- **Modify file:** `app.go`
  - Line 144-146: update `NewBridge` call to `terminal.NewBridge(sm, terminal.NewTmuxAdapter(nil))` (nil runner → DefaultCommandRunner)
  - Check `terminal.IsTmuxAvailable()` before creating the adapter; if false, pass `nil` for the adapter and log a warning
  - On `Stop()`, call `a.bridge.Stop()` as today — the bridge itself will cascade `Close()` to any live attachments via `proxyTmuxSession`'s deferred cleanup

### Technical Considerations

- **Message framing:** The Svelte `Terminal` component sends input as WS binary frames and resize as a JSON text frame like `{"type":"resize","cols":80,"rows":24}` (match the existing `resizeMsg` type in `bridge.go`). `proxyTmuxSession` must recognise both.
- **No input echo loop:** tmux naturally echoes input into pane output, which flows back through the FIFO. Do not echo again in `proxyTmuxSession`.
- **Backpressure:** WebSocket writes are synchronous. If the client is slow, the reader goroutine blocks on `ws.WriteMessage` — acceptable because the FIFO provides upstream buffering (tmux's pipe-pane cat).
- **Context propagation:** Use `ctx, cancel := context.WithCancel(r.Context())` scoped to the WS connection. Cancel on goroutine exit.
- **Error reporting to client:** On attach failure, send `websocket.CloseInternalServerErr` with the error string (truncated to 123 bytes per RFC 6455) before closing.
- **404 semantics stay the same** for non-BMAD unknown sessions — callers depending on 404 for PTY-managed missing sessions still work.

### Risks & Edge Cases

1. **Race: session name hits PTY map AND has bmad- prefix** — unlikely in practice (PTY manager names don't collide with BMAD format) but the PTY lookup wins by ordering. Documented.
2. **Adapter nil** (tmux not installed) — bridge returns 404 for any bmad- session; frontend shows `[session not found]` rather than `[connection error]`. Acceptable degradation.
3. **Attach context mismatch** — use the request context so that WS disconnect cancels the attach. On bridge shutdown, `b.ctx.Done()` cascades.
4. **Double-click View Terminal creating two attachments to same pane** — bridge-02 documented that `pipe-pane -o` replaces the previous pipe. The first WS connection will break when the second arrives. Acceptable; document as known limitation.
5. **WS close mid-stream** — reader goroutine sees a WS write error, exits, cancels context, adapter closes. No leaked goroutines.
6. **Panic safety** — wrap goroutines in `defer recover()` that logs via `log.Printf` and cancels context.

### Reference Files

- `internal/terminal/bridge.go` current 120-line file — the modification target
- `internal/terminal/session.go` lines 216-280 — `AddClient` / `clientReader` goroutine pattern to mirror in `proxyTmuxSession`
- `internal/terminal/bridge_test.go` — existing test patterns for start/stop and WS upgrade
- `app.go` lines 140-170 — current bridge/manager wiring in `OnStartup`
- docs/stories/bridge-02-tmux-adapter.md — the adapter contract this story consumes

Reference skills: `/wails`, `/golang-testing`, `/golang-error-handling`, `/simplify`.

## Acceptance Criteria

**AC-1: Bridge constructor accepts a TmuxAdapter**
- Given a `SessionManager` and a `TmuxAdapter`
- When `NewBridge(manager, adapter)` is called
- Then the bridge stores both references
- And the bridge compiles and starts successfully

**AC-2: PTY-managed sessions continue to route via AddClient**
- Given a session `"shell-abc"` registered in the `SessionManager`
- When a WebSocket connects to `/ws/shell-abc`
- Then `session.AddClient(ws)` is called as before
- And the `TmuxAdapter` is NOT invoked

**AC-3: BMAD-prefixed session names route through the TmuxAdapter**
- Given a `SessionManager` with no session named `"bmad-repo-main-node-deadbeef"`
- And a mock `TmuxAdapter` that returns a successful attachment
- When a WebSocket connects to `/ws/bmad-repo-main-node-deadbeef`
- Then `adapter.Attach(ctx, "bmad-repo-main-node-deadbeef:0.0")` is invoked
- And bytes from the attachment are forwarded to the WS as binary frames

**AC-4: Unknown non-BMAD sessions still return HTTP 404**
- Given no session registered and no BMAD prefix
- When a WebSocket connects to `/ws/randomname`
- Then the server returns HTTP 404
- And the `TmuxAdapter` is NOT invoked

**AC-5: Binary input frames forward to SendInput**
- Given an active BMAD WebSocket connection
- When the client sends a WS binary frame containing `"hello"`
- Then `attachment.SendInput([]byte("hello"))` is invoked

**AC-6: Resize text frames forward to Resize**
- Given an active BMAD WebSocket connection
- When the client sends a WS text frame `{"type":"resize","cols":120,"rows":40}`
- Then `attachment.Resize(120, 40)` is invoked

**AC-7: Attach error sends close frame and disconnects**
- Given a mock adapter that returns `ErrPaneDead` on `Attach`
- When a WebSocket connects to `/ws/bmad-foo:0.0`
- Then the WebSocket receives a close frame containing the error reason
- And the connection is closed

**AC-8: WebSocket close cancels the attachment**
- Given an active BMAD WebSocket connection
- When the client closes the WebSocket
- Then `attachment.Close()` is invoked
- And no goroutines are leaked (verified via `-race` and `goleak` or manual WaitGroup assertion)

**AC-9: App wires the adapter on startup**
- Given `tmux` is on PATH
- When `a.OnStartup` runs
- Then `terminal.NewBridge` is called with a non-nil `TmuxAdapter`
- And `a.bridge.Start(ctx)` succeeds

**AC-10: App tolerates missing tmux**
- Given `tmux` is NOT on PATH
- When `a.OnStartup` runs
- Then `terminal.NewBridge` is called with a nil adapter
- And a warning is logged
- And PTY-managed sessions still work

## BDD Test Scenarios

### Scenario 1: Routing

```gherkin
Feature: Bridge routing precedence

  Scenario: PTY session takes precedence
    Given a SessionManager containing "my-shell"
    And a mock TmuxAdapter
    When a WS connects to "/ws/my-shell"
    Then AddClient is called on the PTY session
    And the TmuxAdapter is not invoked

  Scenario: BMAD prefix on manager miss routes to adapter
    Given a SessionManager with no matching session
    And a mock TmuxAdapter returning a live attachment
    When a WS connects to "/ws/bmad-foo-main-node-deadbeef"
    Then adapter.Attach is called with "bmad-foo-main-node-deadbeef:0.0"

  Scenario: Non-BMAD miss returns 404
    Given a SessionManager with no matching session
    When an HTTP request is made to "/ws/unknown"
    Then the status is 404
    And the adapter is not invoked
```

### Scenario 2: Data flow

```gherkin
Feature: Bidirectional data flow

  Scenario: Pane output streams to client
    Given an active BMAD attachment emitting "line1\nline2\n"
    When the WS client reads frames
    Then the client receives "line1\nline2\n" as binary frames in order

  Scenario: Client input forwards to tmux
    Given an active BMAD connection
    When the client sends binary frame "ls\n"
    Then adapter.SendInput is called with []byte("ls\n")

  Scenario: Resize message forwards
    Given an active BMAD connection
    When the client sends text frame {"type":"resize","cols":120,"rows":40}
    Then adapter.Resize(120, 40) is called
```

### Scenario 3: Error and lifecycle

```gherkin
Feature: Error handling

  Scenario: Attach fails
    Given a mock adapter returning ErrPaneDead
    When a WS connects to "/ws/bmad-dead:0.0"
    Then the WS receives a close frame referencing pane_dead
    And the connection closes

  Scenario: Client disconnects mid-stream
    Given an active BMAD connection
    When the client closes the WS
    Then attachment.Close is called exactly once
    And both proxy goroutines exit

  Scenario: Nil adapter degrades gracefully
    Given a bridge constructed with nil TmuxAdapter
    When a WS connects to "/ws/bmad-anything"
    Then the server returns 404
    And PTY sessions continue to work
```

### Scenario 4: App startup

```gherkin
Feature: App startup wiring

  Scenario: tmux present
    Given IsTmuxAvailable returns true
    When OnStartup runs
    Then NewBridge is called with a non-nil adapter

  Scenario: tmux missing
    Given IsTmuxAvailable returns false
    When OnStartup runs
    Then NewBridge is called with nil adapter
    And a warning is logged
```

## Tasks / Subtasks

- [ ] Task 1: Add `tmuxAdapter` field and update constructor (AC: AC-1, AC-10)
  - [ ] Subtask 1a: Add `tmuxAdapter *TmuxAdapter` field to `Bridge` struct
  - [ ] Subtask 1b: Change `NewBridge` signature to `NewBridge(manager *SessionManager, adapter *TmuxAdapter) *Bridge`
  - [ ] Subtask 1c: Update constructor body; no-op on nil adapter
  - [ ] Subtask 1d: Add local `bmadSessionPrefix = "bmad-"` constant with comment

- [ ] Task 2: Update `handleWS` to try tmux path on PTY miss (AC: AC-2, AC-3, AC-4)
  - [ ] Subtask 2a: Keep existing `manager.Get(name)` happy path
  - [ ] Subtask 2b: On miss, check prefix + non-nil adapter, delegate to `proxyTmuxSession`
  - [ ] Subtask 2c: Otherwise return HTTP 404 as today

- [ ] Task 3: Implement `proxyTmuxSession` (AC: AC-3, AC-5, AC-6, AC-7, AC-8)
  - [ ] Subtask 3a: Build paneTarget by appending `:0.0`
  - [ ] Subtask 3b: Call `Attach`; on error send close frame and return
  - [ ] Subtask 3c: Reader goroutine: `io.Copy`-style loop from `attachment.Read` to `ws.WriteMessage(BinaryMessage, ...)`
  - [ ] Subtask 3d: Writer goroutine: `ws.ReadMessage` loop; dispatch binary→SendInput, text→resize parse
  - [ ] Subtask 3e: Wait on WaitGroup, cancel context, call `attachment.Close()`, close WS

- [ ] Task 4: Update `app.go` wiring (AC: AC-9, AC-10)
  - [ ] Subtask 4a: Call `terminal.IsTmuxAvailable()` during `OnStartup`
  - [ ] Subtask 4b: Create `TmuxAdapter` or pass nil based on availability
  - [ ] Subtask 4c: Pass to `terminal.NewBridge(sm, adapter)`
  - [ ] Subtask 4d: Log warning when adapter is nil

- [ ] Task 5: Write bridge integration tests in `bridge_test.go` (AC: AC-1 through AC-10)
  - [ ] Subtask 5a: Define `mockTmuxAdapter` with recording `Attach`, `SendInput`, `SendKey`, `Resize`, `Close`
  - [ ] Subtask 5b: Test PTY path precedence
  - [ ] Subtask 5c: Test BMAD-prefixed routing to mock adapter
  - [ ] Subtask 5d: Test 404 on non-BMAD miss
  - [ ] Subtask 5e: Test input frame forwarding
  - [ ] Subtask 5f: Test resize frame forwarding
  - [ ] Subtask 5g: Test attach failure close frame
  - [ ] Subtask 5h: Test nil-adapter degradation
  - [ ] Subtask 5i: Compile-time assertion that `bmadSessionPrefix == bmad.SessionNamePrefix`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/terminal/bridge.go` and modified lines of `app.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Story status updated to `done`
