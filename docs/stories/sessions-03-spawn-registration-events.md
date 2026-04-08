# Story 3: Register Sessions on Spawn & Emit Wails Events

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** sessions-01
**Status:** ready

## Description

Wire the existing `spawnTmuxSession`, `SpawnAgent`, `SpawnTerminal`, and `KillAgent` methods to register/deregister sessions in the terminal session registry and emit Wails events (`terminal:session:added`, `terminal:session:removed`). This ensures the registry stays current as users create and destroy sessions through normal app usage, not just through recovery.

## Developer Notes

### Architecture

Modify `app_tmux.go` — the existing spawn/kill methods. No new files.

### Changes to `spawnTmuxSession`

Add two new parameters: `sessionType string` and `model string`. After the successful `tmux new-session` call (line 33), register the session:

```go
session := domain.TerminalSession{
    SessionName: sessionName,
    PaneTarget:  target,
    RepoPath:    repoPath,
    RepoName:    repoNameFromDir(repoPath),
    SessionType: sessionType,
    Model:       model,
    SpawnedAt:   time.Now(),
    IsAlive:     true,
}
a.registerSession(session)
runtime.EventsEmit(a.ctx, "terminal:session:added", session)
```

### Caller Updates

- `SpawnAgent`: Pass `"agent"` and `model` to `spawnTmuxSession`.
- `SpawnAgentWithCommand`: Pass `"agent"` and `""` (model unknown from raw command).
- `SpawnTerminal`: Pass `"terminal"` and `""`.

### Changes to `KillAgent`

After the existing `tmux kill-session` call (line 107), also remove from the registry:

```go
a.mu.Lock()
delete(a.terminalSessions, sessionName)
a.mu.Unlock()
runtime.EventsEmit(a.ctx, "terminal:session:removed", sessionName)
```

Note: `KillAgent` already extracts `sessionName` from `tmuxTarget` at line 104. Reuse that value.

### Technical Considerations

- **Signature change**: `spawnTmuxSession` is private, so changing its signature only affects callers in the same package (`SpawnAgent`, `SpawnAgentWithCommand`, `SpawnTerminal`). All are in `app_tmux.go`.
- **Event payload**: `terminal:session:added` sends the full `TerminalSession` struct (serialized as JSON by Wails). `terminal:session:removed` sends just the `sessionName` string.
- **Race condition**: The registry write and event emission happen after the tmux session is confirmed running, so the frontend will never see a session that failed to spawn.
- **Import addition**: `app_tmux.go` needs `"mashed/internal/domain"` — already imported (line 8). Also needs `"github.com/wailsapp/wails/v2/pkg/runtime"` — add to imports.

### Risks & Edge Cases

- **KillAgent called for process-scanned agent without registry entry**: The `delete` on a non-existent key is a no-op in Go. Safe.
- **SpawnAgentWithCommand model inference**: We cannot parse the model from an arbitrary command string. Leave `Model: ""`. The frontend should handle empty model gracefully.

### Reference Files

- `app_tmux.go` — all modifications happen here
- `app_terminal_registry.go` — `registerSession` method from Story 1
- `app.go:19` — Wails runtime import pattern

## Acceptance Criteria

AC-1: SpawnAgent registers session in registry
- Given an agent is spawned via `SpawnAgent("/dev/repo", "claude-opus-4-6")`
- When the spawn succeeds
- Then a `TerminalSession` with `SessionType: "agent"` and `Model: "claude-opus-4-6"` is in the registry
- And a `terminal:session:added` event is emitted with the session data

AC-2: SpawnTerminal registers session in registry
- Given a terminal is spawned via `SpawnTerminal("/dev/repo")`
- When the spawn succeeds
- Then a `TerminalSession` with `SessionType: "terminal"` and `Model: ""` is in the registry
- And a `terminal:session:added` event is emitted

AC-3: KillAgent removes session from registry
- Given an agent session "mashed-repo-100" is in the registry
- When `KillAgent` is called with a tmuxTarget of "mashed-repo-100:0.0"
- Then the session is removed from `a.terminalSessions`
- And a `terminal:session:removed` event is emitted with "mashed-repo-100"

AC-4: Failed spawn does not register
- Given tmux is not available
- When `SpawnAgent` is called
- Then the method returns an error
- And no session is added to the registry
- And no event is emitted

## BDD Test Scenarios

### Scenario 1: Agent Spawn Registration

```gherkin
Feature: Session registration on spawn

  Scenario: SpawnAgent registers and emits event
    Given the app is initialized with an empty registry
    And tmux is available
    When SpawnAgent is called with repoPath "/dev/myrepo" and model "claude-opus-4-6"
    Then the registry contains 1 session
    And the session has SessionType "agent" and Model "claude-opus-4-6"
    And a "terminal:session:added" event was emitted
```

### Scenario 2: Terminal Spawn Registration

```gherkin
Feature: Terminal spawn registration

  Scenario: SpawnTerminal registers with type "terminal"
    Given the app is initialized
    And tmux is available
    When SpawnTerminal is called with repoPath "/dev/myrepo"
    Then the registry contains 1 session
    And the session has SessionType "terminal" and Model ""
    And a "terminal:session:added" event was emitted
```

### Scenario 3: Kill Deregisters

```gherkin
Feature: Kill removes from registry

  Scenario: KillAgent deregisters and emits removal event
    Given the registry has session "mashed-repo-100" with PaneTarget "mashed-repo-100:0.0"
    And the tmux session exists
    When KillAgent is called with tmuxTarget "mashed-repo-100:0.0"
    Then the registry no longer contains "mashed-repo-100"
    And a "terminal:session:removed" event was emitted with "mashed-repo-100"
```

### Scenario 4: Failed Spawn

```gherkin
Feature: Failed spawn does not pollute registry

  Scenario: tmux failure prevents registration
    Given tmux is not available
    When SpawnAgent is called with repoPath "/dev/repo" and model "sonnet"
    Then an error is returned
    And the registry is empty
    And no events were emitted
```

## Tasks / Subtasks

- [ ] Task 1: Update spawnTmuxSession signature (AC: AC-1, AC-2, AC-4)
  - [ ] Add `sessionType string` and `model string` parameters to `spawnTmuxSession`
  - [ ] After successful spawn, call `a.registerSession(session)`
  - [ ] Emit `terminal:session:added` event via `runtime.EventsEmit`
  - [ ] Add `runtime` import to `app_tmux.go`

- [ ] Task 2: Update callers (AC: AC-1, AC-2)
  - [ ] `SpawnAgent`: pass `"agent"`, `model`
  - [ ] `SpawnAgentWithCommand`: pass `"agent"`, `""`
  - [ ] `SpawnTerminal`: pass `"terminal"`, `""`

- [ ] Task 3: Update KillAgent for registry cleanup (AC: AC-3)
  - [ ] After `tmux kill-session`, delete from `a.terminalSessions`
  - [ ] Emit `terminal:session:removed` event with sessionName

- [ ] Task 4: Write tests (AC: AC-1, AC-2, AC-3, AC-4)
  - [ ] Test that registry population happens after spawn
  - [ ] Test that KillAgent removes from registry
  - [ ] Test that failed spawn leaves registry empty

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
