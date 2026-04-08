# Story 1: Terminal Session Registry — Domain Type & Backend API

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Add a `TerminalSession` domain type and an in-memory registry on the `App` struct that tracks every tmux session spawned by mashed. Expose two Wails-bound methods — `ListRepoSessions(repoPath)` and `KillTerminalSession(sessionName)` — so the frontend can query and manage sessions per repo. The registry is the foundation for all subsequent session persistence stories.

## Developer Notes

### Architecture

- **New type** `TerminalSession` in `internal/domain/types.go` — add after the existing `AgentSession` struct (line ~122).
- **New field** `terminalSessions map[string]domain.TerminalSession` on the `App` struct in `app.go`. Initialize in `NewApp()`.
- **New file** `app_terminal_registry.go` in the root package — contains `ListRepoSessions`, `KillTerminalSession`, and a private helper `registerSession`.
- The registry is keyed by `SessionName` (e.g., `"term-myrepo-1712600000"`), which is already unique due to the Unix timestamp suffix in `spawnTmuxSession()` (see `app_tmux.go:21`).
- `ListRepoSessions` cross-references the registry with live tmux panes via `a.panes.ListPanes()` to mark `IsAlive` and prune dead entries.

### Key Types

```go
// internal/domain/types.go
type TerminalSession struct {
    SessionName string    `json:"sessionName"` // "term-myrepo-1712600000"
    PaneTarget  string    `json:"paneTarget"`  // "term-myrepo-1712600000:0.0"
    RepoPath    string    `json:"repoPath"`
    RepoName    string    `json:"repoName"`
    SessionType string    `json:"sessionType"` // "terminal" | "agent"
    Model       string    `json:"model"`       // "" for terminals
    SpawnedAt   time.Time `json:"spawnedAt"`
    IsAlive     bool      `json:"isAlive"`     // set at query time from tmux liveness
}
```

### Data Flow

1. `registerSession(session)` — acquires `a.mu`, inserts into `a.terminalSessions`
2. `ListRepoSessions(repoPath)` — locks `a.mu`, filters by `RepoPath`, calls `a.panes.ListPanes()`, builds alive-set from `TmuxPane.SessionName`, marks alive/dead, deletes dead, returns alive sorted by `SpawnedAt`
3. `KillTerminalSession(sessionName)` — runs `tmux kill-session -t {sessionName}`, removes from map, emits `terminal:session:removed` via `runtime.EventsEmit`

### Technical Considerations

- **Concurrency**: `a.mu` (existing `sync.Mutex` on `App`) protects the map. `ListPanes()` has its own internal mutex so call it outside the lock to avoid holding `a.mu` while shelling out to tmux.
- **Error handling**: `KillTerminalSession` should not fail if the session is already dead — log and remove from registry. Use `fmt.Errorf` wrapping per project convention.
- **Wails binding**: Both public methods auto-generate JS bindings on `wails dev`/`wails generate`. Return `[]domain.TerminalSession` (not pointers) for clean JSON serialization.

### Risks & Edge Cases

- **tmux not running**: `ListPanes()` returns `ErrTmuxNotRunning` — catch this and return all entries as `IsAlive: false`, then prune.
- **Empty repoPath**: `ListRepoSessions("")` should return empty slice, not error.
- **Concurrent kill**: Two calls to `KillTerminalSession` for the same session — second should succeed silently (already removed).

### Reference Files

- `app.go` — `App` struct, `NewApp()`, `a.mu` usage pattern
- `app_tmux.go` — `spawnTmuxSession` (session naming convention at line 21)
- `internal/terminal/panes.go` — `PaneDiscovery.ListPanes()`, `TmuxPane.SessionName`
- `internal/domain/types.go` — existing types to place `TerminalSession` near
- `app_bmad.go` — example of a Wails-bound method file pattern

## Acceptance Criteria

AC-1: TerminalSession type exists in domain
- Given the codebase has no `TerminalSession` type
- When `internal/domain/types.go` is compiled
- Then a `TerminalSession` struct with fields `SessionName`, `PaneTarget`, `RepoPath`, `RepoName`, `SessionType`, `Model`, `SpawnedAt`, `IsAlive` is available
- And JSON tags match the plan's naming convention

AC-2: ListRepoSessions returns sessions filtered by repo
- Given the registry contains sessions for repos `/a` and `/b`
- When `ListRepoSessions("/a")` is called
- Then only sessions with `RepoPath == "/a"` are returned
- And results are sorted by `SpawnedAt` ascending

AC-3: ListRepoSessions prunes dead sessions
- Given the registry contains a session whose tmux session no longer exists
- When `ListRepoSessions` is called
- Then the dead session is removed from the registry
- And it is not included in the returned slice

AC-4: KillTerminalSession kills tmux and removes from registry
- Given the registry contains session "term-repo-123"
- When `KillTerminalSession("term-repo-123")` is called
- Then `tmux kill-session -t term-repo-123` is executed
- And the session is removed from `a.terminalSessions`
- And a `terminal:session:removed` Wails event is emitted

AC-5: KillTerminalSession handles already-dead session gracefully
- Given session "term-repo-999" is in the registry but tmux session is already gone
- When `KillTerminalSession("term-repo-999")` is called
- Then the method returns nil (no error)
- And the session is removed from the registry

## BDD Test Scenarios

### Scenario 1: Domain Type

```gherkin
Feature: TerminalSession domain type

  Scenario: TerminalSession struct has correct JSON tags
    Given the TerminalSession type is defined in domain
    When serialized to JSON
    Then the keys are "sessionName", "paneTarget", "repoPath", "repoName", "sessionType", "model", "spawnedAt", "isAlive"
```

### Scenario 2: List Sessions Filtering

```gherkin
Feature: ListRepoSessions filtering and liveness

  Scenario: Filter sessions by repoPath
    Given the registry has session "mashed-foo-100" for "/dev/foo"
    And the registry has session "term-bar-200" for "/dev/bar"
    And both tmux sessions are alive
    When ListRepoSessions("/dev/foo") is called
    Then exactly 1 session is returned
    And its SessionName is "mashed-foo-100"

  Scenario: Empty repoPath returns empty slice
    Given the registry has sessions
    When ListRepoSessions("") is called
    Then an empty slice is returned

  Scenario: Dead sessions are pruned from registry
    Given the registry has session "term-foo-100" for "/dev/foo"
    And the tmux session "term-foo-100" does not exist
    When ListRepoSessions("/dev/foo") is called
    Then an empty slice is returned
    And the registry no longer contains "term-foo-100"

  Scenario: Results sorted by SpawnedAt ascending
    Given the registry has session "term-foo-300" spawned at 300 for "/dev/foo"
    And the registry has session "term-foo-100" spawned at 100 for "/dev/foo"
    And both tmux sessions are alive
    When ListRepoSessions("/dev/foo") is called
    Then the first session has SpawnedAt 100
    And the second session has SpawnedAt 300
```

### Scenario 3: Kill Session

```gherkin
Feature: KillTerminalSession

  Scenario: Kill a live session
    Given the registry has session "term-foo-100"
    And the tmux session exists
    When KillTerminalSession("term-foo-100") is called
    Then tmux kill-session is invoked for "term-foo-100"
    And the registry no longer contains "term-foo-100"
    And a "terminal:session:removed" event is emitted with data "term-foo-100"

  Scenario: Kill an already-dead session
    Given the registry has session "term-foo-100"
    And the tmux session does not exist
    When KillTerminalSession("term-foo-100") is called
    Then no error is returned
    And the registry no longer contains "term-foo-100"

  Scenario: Kill unknown session
    Given the registry does not contain "nonexistent"
    When KillTerminalSession("nonexistent") is called
    Then an error is returned indicating session not found
```

## Tasks / Subtasks

- [ ] Task 1: Add TerminalSession domain type (AC: AC-1)
  - [ ] Add `TerminalSession` struct to `internal/domain/types.go` after `AgentSession`
  - [ ] Write unit test for JSON serialization of `TerminalSession`

- [ ] Task 2: Add registry field to App (AC: AC-2, AC-3, AC-4, AC-5)
  - [ ] Add `terminalSessions map[string]domain.TerminalSession` field to `App` struct in `app.go`
  - [ ] Initialize the map in `NewApp()`

- [ ] Task 3: Implement app_terminal_registry.go (AC: AC-2, AC-3, AC-4, AC-5)
  - [ ] Create `app_terminal_registry.go` with `registerSession`, `ListRepoSessions`, `KillTerminalSession`
  - [ ] Implement liveness cross-reference in `ListRepoSessions` using `a.panes.ListPanes()`
  - [ ] Implement sorting by `SpawnedAt`
  - [ ] Implement `KillTerminalSession` with `tmux kill-session`, registry removal, and event emission
  - [ ] Handle edge cases: tmux not running, empty repoPath, already-dead session

- [ ] Task 4: Write tests for registry (AC: AC-2, AC-3, AC-4, AC-5)
  - [ ] Create `app_terminal_registry_test.go` with table-driven tests
  - [ ] Test ListRepoSessions filtering, sorting, and pruning (mock or inject PaneDiscovery)
  - [ ] Test KillTerminalSession success and already-dead paths

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
