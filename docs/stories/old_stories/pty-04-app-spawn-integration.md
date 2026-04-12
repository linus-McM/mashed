# Story 4: App Spawn and Lifecycle Integration

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** L
**Depends On:** pty-02, pty-03
**Status:** ready

## Description

Wire the `SessionManager` into the application layer by adding a `manager` field to `App`, updating `NewApp()` and `startup()`/`shutdown()`, and rewriting the spawn methods in `app_tmux.go` (renamed to `app_spawn.go`) to use `manager.Spawn()` instead of tmux commands. Also update `app_terminal_registry.go` to use manager-based liveness checks instead of tmux pane discovery. This story connects the new PTY management to the existing Wails bindings.

## Developer Notes

### Architecture
- **Modify file:** `app.go` (374 lines)
  - Add `manager *terminal.SessionManager` field to `App` struct (line 33-51)
  - `NewApp()` (line 108-114): create `terminal.NewSessionManager()`, pass to `terminal.NewBridge(manager)`
  - `startup()` (line 117-155): remove `a.recoverSessions()` call (line 129) -- PTY sessions don't survive restart
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
  - `recoverSessions()` (line 20-60): convert to no-op (PTY sessions don't survive restart). Keep the function signature for now to avoid breaking callers, but gut the body to just `return`.
  - `ListRepoSessions()` (line 140-185): replace `a.panes.ListPanes()` liveness check with `a.manager.IsAlive(name)` for each session in registry
  - `KillTerminalSession()` (line 190-204): replace `exec.Command("tmux", "kill-session", ...)` with `a.manager.Kill(sessionName)`
- **Modify file:** `app_terminal_registry_test.go` -- update tests that reference `:0.0` pane targets

### Technical Considerations
- **Backward compatibility of `PaneTarget` field:** The `domain.TerminalSession.PaneTarget` field is used by the frontend to construct WebSocket URLs. After this change, `PaneTarget` equals `SessionName` (no `:0.0`). The frontend `Terminal.svelte` already passes `paneTarget` directly into the WS URL path, so this change is compatible.
- **`spawnSession` naming:** Rename from `spawnTmuxSession` to `spawnSession` since tmux is no longer involved.
- **KillAgent dual-path:** The `KillAgent` method needs to work for both manager-spawned sessions AND externally running claude processes found by scan. Manager kill is tried first; if the session isn't in the manager (external process), fall back to direct process signal.
- **Event consistency:** `eventSessionAdded` and `eventSessionRemoved` events must continue to fire with the same payload shape so frontend stores work unchanged.

### Risks & Edge Cases
- **File rename `app_tmux.go` -> `app_spawn.go`:** Use `git mv` to preserve history. All functions in this file are methods on `*App`, so the package doesn't change.
- **`recoverSessions` callers:** Only called from `startup()`. Making it a no-op is safe. The function and `recoverSessionsFromOutput` helper can be removed in a follow-up cleanup, but keeping them as no-ops avoids breaking the test file.
- **ListRepoSessions performance:** Currently shells out to `tmux list-sessions`. New version iterates in-memory map -- strictly faster.
- **Test updates:** `app_terminal_registry_test.go` has tests that assert `:0.0` in PaneTarget (lines 46, 72, 294, 300, 406, 433, 458). These need updating.

### Reference Files
- `app.go` (lines 33-51 for App struct, 108-114 for NewApp, 117-155 for startup, 158-165 for shutdown)
- `app_tmux.go` (full file, 159 lines -- spawn + kill logic)
- `app_terminal_registry.go` (full file, 205 lines -- recovery + liveness + kill)
- `app_terminal_registry_test.go` (tests referencing `:0.0` targets)
- `internal/domain/types.go` (lines 123-133 for `TerminalSession` struct)
- `frontend/src/lib/stores/sessions.js` (line 39 -- `makeSession` uses `target.replace(':0.0', '')`)

## Acceptance Criteria

AC-1: App creates SessionManager and passes it to Bridge
- Given the application starts via `NewApp()`
- When `startup()` runs
- Then `a.manager` is a non-nil `*SessionManager`
- And `a.bridge` was created with `NewBridge(a.manager)`

AC-2: SpawnAgent creates a managed PTY session (not tmux)
- Given the app is running
- When `SpawnAgent("/tmp/myrepo", "claude-opus-4-6")` is called
- Then a `ManagedSession` is created in the manager running `claude --dangerously-skip-permissions --model claude-opus-4-6`
- And a `TerminalSession` is registered in `terminalSessions` with no `:0.0` suffix
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
- And the session is deregistered
- And the engine agent is removed

AC-5: ListRepoSessions uses manager-based liveness
- Given the terminal registry has sessions "sess-a" (alive in manager) and "sess-b" (dead)
- When `ListRepoSessions("/tmp/repo")` is called
- Then "sess-a" is returned with `IsAlive: true`
- And "sess-b" is pruned from the registry

AC-6: Shutdown kills all managed sessions before stopping the bridge
- Given the app has 2 active managed sessions
- When `shutdown()` is called
- Then `a.manager.Shutdown()` is called before `a.bridge.Stop()`
- And all sessions are terminated

AC-7: recoverSessions is a no-op
- Given the app starts
- When `startup()` runs
- Then `recoverSessions()` does not shell out to tmux
- And no sessions are recovered from tmux

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
    And "dead-sess" is pruned from the registry
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

## Tasks / Subtasks

- [ ] Task 1: Add `manager` field to App and update constructors (AC: AC-1)
  - [ ] Subtask 1a: Add `manager *terminal.SessionManager` to `App` struct in `app.go`
  - [ ] Subtask 1b: Update `NewApp()` to create `SessionManager` and pass to `NewBridge(manager)`
  - [ ] Subtask 1c: Update `shutdown()` to call `a.manager.Shutdown()` before `a.bridge.Stop()`
  - [ ] Subtask 1d: Remove `a.recoverSessions()` call from `startup()`

- [ ] Task 2: Rename and rewrite spawn methods (AC: AC-2, AC-3)
  - [ ] Subtask 2a: `git mv app_tmux.go app_spawn.go`
  - [ ] Subtask 2b: Rename `spawnTmuxSession` to `spawnSession`, replace tmux exec with `a.manager.Spawn()`
  - [ ] Subtask 2c: Update return value: session name without `:0.0` suffix
  - [ ] Subtask 2d: Update `TerminalSession` registration: `PaneTarget = sessionName` (no `:0.0`)
  - [ ] Subtask 2e: Remove `a.panes.InvalidateCache()` call

- [ ] Task 3: Update KillAgent to use manager (AC: AC-4)
  - [ ] Subtask 3a: Try `a.manager.Kill(sessionName)` first in `KillAgent`
  - [ ] Subtask 3b: Keep process signal fallback for externally spawned processes
  - [ ] Subtask 3c: Keep engine cleanup and notification pruning unchanged

- [ ] Task 4: Update terminal registry for manager-based liveness (AC: AC-5, AC-7)
  - [ ] Subtask 4a: Gut `recoverSessions()` body to no-op
  - [ ] Subtask 4b: Rewrite `ListRepoSessions()` to use `a.manager.IsAlive(name)` instead of `a.panes.ListPanes()`
  - [ ] Subtask 4c: Rewrite `KillTerminalSession()` to use `a.manager.Kill(sessionName)` instead of tmux kill-session

- [ ] Task 5: Update tests (AC: all)
  - [ ] Subtask 5a: Update `app_terminal_registry_test.go` to remove `:0.0` PaneTarget assertions
  - [ ] Subtask 5b: Add tests for `spawnSession` with mock manager
  - [ ] Subtask 5c: Add tests for manager-based `ListRepoSessions`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on modified files (`app_spawn.go`, `app_terminal_registry.go`)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
