# Story 4: App Spawn Integration & Frontend Terminal Cleanup

**Priority:** P0-critical
**Domain:** full-stack (backend + frontend)
**Estimated Complexity:** L
**Depends On:** pty-02, pty-03
**Status:** ready

---

## Description

Wire the `SessionManager` into the application layer and clean up all tmux remnants from the frontend in a single story. This combines the backend integration (App struct, spawn methods, kill, liveness) with the frontend cleanup (remove `stripControlSequences`, update session naming) to avoid back-and-forth on the same files — particularly `AgentDetail.svelte` and `sessions.js` which both need session naming changes.

**Merged from:** `pty-04-app-spawn-integration.md` + `pty-06-frontend-cleanup.md`

---

## Developer Notes

### Backend: App Struct & Spawn Integration

- **Modify file:** `app.go` (374 lines)
  - Add `manager *terminal.SessionManager` field to `App` struct (line 33-51)
  - `NewApp()` (line 108-114): create `terminal.NewSessionManager()`, pass to `terminal.NewBridge(manager)`
  - `startup()` (line 117-155): remove `a.recoverSessions()` call (line 129) — PTY sessions don't survive restart
  - `shutdown()` (line 158-165): call `a.manager.Shutdown()` before `a.bridge.Stop()`

- **Rename file:** `app_tmux.go` -> `app_spawn.go` (159 lines)
  - Rewrite `spawnTmuxSession` -> `spawnSession(prefix, repoPath, command string, sessionType domain.SessionType, model string) (string, error)`
    - Uses `a.manager.Spawn(a.ctx, sessionName, repoPath, command)` instead of `exec.Command("tmux", ...)`
    - Returns `sessionName` directly (no `:0.0` suffix)
    - Remove `a.panes.InvalidateCache()` call
    - Keep: session registration, event emission, logging
    - Update `TerminalSession.PaneTarget` to be same as `SessionName` (no `:0.0`)
  - Update `SpawnAgent`, `SpawnAgentWithCommand`, `SpawnTerminal` to call `spawnSession`
  - Update `KillAgent` (line 106-158):
    - Try `a.manager.Kill(sessionName)` first
    - Keep process signal fallback for externally spawned agents
    - Keep engine cleanup and notification pruning

- **Modify file:** `app_terminal_registry.go` (205 lines)
  - `recoverSessions()` (line 20-60): convert to no-op (PTY sessions don't survive restart)
  - `ListRepoSessions()` (line 140-185): replace `a.panes.ListPanes()` liveness check with `a.manager.IsAlive(name)`
  - `KillTerminalSession()` (line 190-204): replace `exec.Command("tmux", "kill-session", ...)` with `a.manager.Kill(sessionName)`

- **Modify file:** `app_terminal_registry_test.go` — update tests that reference `:0.0` pane targets

### Frontend: Terminal & Session Cleanup

- **Modify file:** `frontend/src/components/Terminal.svelte` (278 lines)
  - **Remove:** `stripRe` regex definition (lines 123-128) and `stripControlSequences` function (lines 129-131)
  - **Remove:** Call to `stripControlSequences(raw)` in `ws.onmessage` handler (line 185). Pass `raw` directly to `term.write(raw)`
  - **Remove:** Comment block about stripping escape sequences (lines 117-122)
  - **Change:** Line 166: `'Connecting to tmux session...'` -> `'Connecting...'`

- **Modify file:** `frontend/src/lib/stores/sessions.js` (48 lines)
  - **Change:** Line 39: `sessionName: target.replace(':0.0', '')` -> `sessionName: target`
  - The `makeSession` function is called from AgentDetail.svelte after spawn methods return. Since spawn methods now return session names without `:0.0`, the `.replace()` is no longer needed.

### Technical Considerations
- **Backward compatibility of `PaneTarget` field:** The `domain.TerminalSession.PaneTarget` field is used by the frontend to construct WebSocket URLs. After this change, `PaneTarget` equals `SessionName` (no `:0.0`). Terminal.svelte already passes `paneTarget` directly into the WS URL path.
- **KillAgent dual-path:** Must work for both manager-spawned sessions AND externally running claude processes found by scan. Manager kill first, process signal fallback.
- **Event consistency:** `eventSessionAdded` and `eventSessionRemoved` must continue to fire with the same payload shape.
- **xterm.js improvement:** Without `stripControlSequences`, xterm.js receives raw PTY output directly. Since managed PTY runs shell/claude directly (not through tmux), no alternate screen sequences to strip.
- **File rename:** Use `git mv app_tmux.go app_spawn.go` to preserve history.

### Risks & Edge Cases
- **ListRepoSessions performance:** Replaces tmux shell-out with in-memory map — strictly faster.
- **Programs using alternate screen (vim, less):** Escape sequences flow directly to xterm.js which handles them natively. Correct behavior.
- **Bracketed paste mode:** Without tmux the shell itself sends these, xterm.js handles correctly.
- **Test updates:** `app_terminal_registry_test.go` has tests asserting `:0.0` in PaneTarget (lines 46, 72, 294, 300, 406, 433, 458).

### Reference Files
- `app.go` (lines 33-51 App struct, 108-114 NewApp, 117-155 startup, 158-165 shutdown)
- `app_tmux.go` (full file, 159 lines — spawn + kill logic)
- `app_terminal_registry.go` (full file, 205 lines)
- `app_terminal_registry_test.go` (tests referencing `:0.0`)
- `internal/domain/types.go` (lines 123-133 `TerminalSession` struct)
- `frontend/src/components/Terminal.svelte` (full file, 278 lines)
- `frontend/src/lib/stores/sessions.js` (full file, 48 lines)

---

## Acceptance Criteria

### Backend

AC-1: App creates SessionManager and passes it to Bridge
- Given the application starts via `NewApp()`
- When `startup()` runs
- Then `a.manager` is a non-nil `*SessionManager`
- And `a.bridge` was created with `NewBridge(a.manager)`

AC-2: SpawnAgent creates a managed PTY session (not tmux)
- Given the app is running
- When `SpawnAgent("/tmp/myrepo", "claude-opus-4-6")` is called
- Then a `ManagedSession` is created running `claude --dangerously-skip-permissions --model claude-opus-4-6`
- And a `TerminalSession` is registered with no `:0.0` suffix
- And an `eventSessionAdded` Wails event is emitted

AC-3: SpawnTerminal creates a managed shell PTY session
- Given the app is running
- When `SpawnTerminal("/tmp/myrepo")` is called
- Then a `ManagedSession` is created running the user's default shell
- And its working directory is `/tmp/myrepo`

AC-4: KillAgent terminates managed sessions
- Given a managed session "mashed-myrepo-123" exists
- When `KillAgent("pid-456", 456, "mashed-myrepo-123")` is called
- Then `a.manager.Kill("mashed-myrepo-123")` is called
- And the session is deregistered and engine agent removed

AC-5: ListRepoSessions uses manager-based liveness
- Given the terminal registry has "sess-a" (alive) and "sess-b" (dead)
- When `ListRepoSessions("/tmp/repo")` is called
- Then "sess-a" is returned with `IsAlive: true`
- And "sess-b" is pruned

AC-6: Shutdown kills all managed sessions before stopping the bridge
- Given 2 active managed sessions
- When `shutdown()` is called
- Then `a.manager.Shutdown()` completes before `a.bridge.Stop()`

AC-7: recoverSessions is a no-op
- Given the app starts
- When `startup()` runs
- Then `recoverSessions()` does not shell out to tmux

### Frontend

AC-8: stripControlSequences is completely removed
- Given the updated Terminal.svelte
- When searched for "stripControlSequences" or "stripRe"
- Then zero matches are found
- And ws.onmessage writes raw data directly to `term.write()`

AC-9: Connection message no longer mentions tmux
- Given a terminal connecting to a session
- Then the user sees "Connecting..." (not "Connecting to tmux session...")

AC-10: makeSession no longer strips :0.0 suffix
- Given the sessions store
- When `makeSession("mashed-repo-123", ...)` is called
- Then `sessionName` is `"mashed-repo-123"` (unchanged, no `.replace()`)

AC-11: Terminal works with direct PTY output
- Given a Terminal connected to a managed session
- When the session produces ANSI color codes and cursor movement
- Then xterm.js renders them correctly without stripping

AC-12: Clipboard integration unchanged
- Given a Terminal with an active session
- When user presses Cmd+C/Cmd+V
- Then clipboard operations work via `ClipboardSetText`/`ClipboardGetText`

---

## BDD Test Scenarios

### Scenario 1: Spawn Integration
```gherkin
Feature: PTY spawn integration

  Scenario: SpawnAgent creates managed session
    Given a running App with SessionManager
    When SpawnAgent is called with repo "/tmp/repo" and model "claude-opus-4-6"
    Then manager.Get returns a session with the expected name pattern
    And the session command contains "claude --dangerously-skip-permissions"
    And terminalSessions registry has the session with no ":0.0" suffix

  Scenario: SpawnTerminal creates shell session
    Given a running App with SessionManager
    When SpawnTerminal is called with repo "/tmp/repo"
    Then manager.Get returns a session running the default shell
    And the session name starts with "term-"
```

### Scenario 2: Kill Integration
```gherkin
Feature: Kill agent integration

  Scenario: KillAgent kills managed session
    Given a managed session "mashed-repo-100" in the manager
    When KillAgent is called with tmuxTarget "mashed-repo-100"
    Then the session is killed in the manager
    And deregisterSession is called
    And engine.RemoveAgent is called

  Scenario: KillAgent falls back to process signal for external agents
    Given no managed session for the target
    And a process with PID 123 is running
    When KillAgent is called with pid 123
    Then the process receives SIGINT
```

### Scenario 3: Liveness Check
```gherkin
Feature: Manager-based liveness

  Scenario: ListRepoSessions uses manager.IsAlive
    Given terminalSessions has "alive-sess" and "dead-sess" for "/tmp/repo"
    And manager.IsAlive("alive-sess") returns true
    And manager.IsAlive("dead-sess") returns false
    When ListRepoSessions("/tmp/repo") is called
    Then only "alive-sess" is returned
    And "dead-sess" is pruned
```

### Scenario 4: Startup and Shutdown
```gherkin
Feature: App lifecycle

  Scenario: Startup does not recover tmux sessions
    Given a fresh App instance
    When startup is called
    Then no tmux commands are executed
    And terminalSessions map is empty

  Scenario: Shutdown order
    Given an App with active sessions
    When shutdown is called
    Then manager.Shutdown() completes before bridge.Stop()
```

### Scenario 5: Raw Output Passthrough
```gherkin
Feature: Terminal raw output

  Scenario: PTY output rendered without stripping
    Given a Terminal component connected to a managed session
    When the session sends "\x1b[32mgreen text\x1b[0m"
    Then xterm.js renders "green text" in green
    And no escape sequences are stripped

  Scenario: Alternate screen sequences pass through
    Given a Terminal component connected to a managed session
    When the session runs a program that uses alternate screen
    Then xterm.js enters alternate screen mode natively
```

### Scenario 6: Session Name Handling
```gherkin
Feature: Session name in store

  Scenario: makeSession preserves session name as-is
    Given a spawn method returns "mashed-repo-123"
    When makeSession("mashed-repo-123", "/tmp/repo", "repo", "agent", "opus") is called
    Then sessionName is "mashed-repo-123"
    And paneTarget is "mashed-repo-123"
```

---

## Tasks / Subtasks

### Phase A: Backend Integration

- [ ] Task 1: Add `manager` field to App and update constructors (AC: 1, 6, 7)
  - [ ] Subtask 1a: Add `manager *terminal.SessionManager` to `App` struct in `app.go`
  - [ ] Subtask 1b: Update `NewApp()` to create `SessionManager` and pass to `NewBridge(manager)`
  - [ ] Subtask 1c: Update `shutdown()` to call `a.manager.Shutdown()` before `a.bridge.Stop()`
  - [ ] Subtask 1d: Remove `a.recoverSessions()` call from `startup()`

- [ ] Task 2: Rename and rewrite spawn methods (AC: 2, 3)
  - [ ] Subtask 2a: `git mv app_tmux.go app_spawn.go`
  - [ ] Subtask 2b: Rename `spawnTmuxSession` to `spawnSession`, replace tmux exec with `a.manager.Spawn()`
  - [ ] Subtask 2c: Update return value: session name without `:0.0` suffix
  - [ ] Subtask 2d: Update `TerminalSession` registration: `PaneTarget = sessionName` (no `:0.0`)
  - [ ] Subtask 2e: Remove `a.panes.InvalidateCache()` call

- [ ] Task 3: Update KillAgent to use manager (AC: 4)
  - [ ] Subtask 3a: Try `a.manager.Kill(sessionName)` first in `KillAgent`
  - [ ] Subtask 3b: Keep process signal fallback for externally spawned processes
  - [ ] Subtask 3c: Keep engine cleanup and notification pruning unchanged

- [ ] Task 4: Update terminal registry for manager-based liveness (AC: 5, 7)
  - [ ] Subtask 4a: Gut `recoverSessions()` body to no-op
  - [ ] Subtask 4b: Rewrite `ListRepoSessions()` to use `a.manager.IsAlive(name)`
  - [ ] Subtask 4c: Rewrite `KillTerminalSession()` to use `a.manager.Kill(sessionName)`

- [ ] Task 5: Update backend tests (AC: 1-7)
  - [ ] Subtask 5a: Update `app_terminal_registry_test.go` to remove `:0.0` PaneTarget assertions
  - [ ] Subtask 5b: Add tests for `spawnSession` with mock manager
  - [ ] Subtask 5c: Add tests for manager-based `ListRepoSessions`

### Phase B: Frontend Cleanup

- [ ] Task 6: Remove stripControlSequences from Terminal.svelte (AC: 8, 9, 11)
  - [ ] Subtask 6a: Delete `stripRe` regex definition (lines 123-128)
  - [ ] Subtask 6b: Delete `stripControlSequences` function (lines 129-131)
  - [ ] Subtask 6c: Delete comment block about stripping (lines 117-122)
  - [ ] Subtask 6d: Update `ws.onmessage` to pass `raw` directly to `term.write(raw)`
  - [ ] Subtask 6e: Change "Connecting to tmux session..." to "Connecting..." on line 166

- [ ] Task 7: Update sessions store (AC: 10)
  - [ ] Subtask 7a: Change `sessionName: target.replace(':0.0', '')` to `sessionName: target` in `makeSession`

- [ ] Task 8: Manual verification (AC: 11, 12)
  - [ ] Subtask 8a: Verify xterm.js renders colored output correctly with direct PTY
  - [ ] Subtask 8b: Verify mouse scroll, text selection, Cmd+C/Cmd+V clipboard operations

---

## Files Modified

| File | Action |
|------|--------|
| `app.go` | Add manager field, update NewApp/startup/shutdown |
| `app_tmux.go` -> `app_spawn.go` | Rename + rewrite spawn/kill methods |
| `app_terminal_registry.go` | Replace tmux liveness with manager-based |
| `app_terminal_registry_test.go` | Update `:0.0` assertions |
| `frontend/src/components/Terminal.svelte` | Remove stripControlSequences, update connection msg |
| `frontend/src/lib/stores/sessions.js` | Remove `:0.0` suffix stripping |

---

## Definition of Done

- [ ] All 12 acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on modified files (`app_spawn.go`, `app_terminal_registry.go`)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] No "tmux" references remain in Terminal.svelte (except historical comments)
- [ ] No "stripControlSequences" references remain in `frontend/src/`
- [ ] `wails dev` compiles the frontend without errors
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
