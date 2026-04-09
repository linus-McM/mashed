# Story 2: Startup Session Recovery from tmux

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** sessions-01
**Status:** done

## Description

When the mashed app starts (or restarts after a crash), reconstruct the terminal session registry by scanning existing tmux sessions that match the `term-*` and `mashed-*` naming prefixes. This ensures users who close and reopen mashed can see all their still-running terminal and agent sessions without losing track of them.

## Developer Notes

### Architecture

- **New private method** `recoverSessions()` on `App` in `app_terminal_registry.go` (same file created in Story 1).
- Called from `startup()` in `app.go`, after `a.bridge.Start()` and before the scanning goroutine begins.
- Queries tmux for all session names, filters by prefix, then queries each session's working directory to populate `RepoPath`.

### Implementation Steps

1. Run `tmux list-sessions -F '#{session_name}'` via `exec.CommandContext(a.ctx, ...)`.
2. Filter lines matching `term-*` or `mashed-*` prefixes using `strings.HasPrefix`.
3. For each matching session, run `tmux display-message -t {name} -p '#{pane_current_path}'` to get the working directory.
4. Infer `SessionType` from prefix: `"term-"` -> `"terminal"`, `"mashed-"` -> `"agent"`.
5. Build `domain.TerminalSession` with `PaneTarget: name + ":0.0"`, `RepoName` from `filepath.Base(repoPath)`, `SpawnedAt: time.Now()` (approximate — tmux doesn't expose creation time easily), `IsAlive: true`.
6. Register via `a.registerSession(session)`.

### Technical Considerations

- **tmux not running**: If `tmux list-sessions` fails (no server), log a message and return early — no sessions to recover.
- **Model inference**: For `mashed-*` sessions, we cannot determine the model from tmux alone. Set `Model: ""` — the frontend can display "unknown" or infer from context.
- **SpawnedAt accuracy**: Using `time.Now()` for recovered sessions means ordering within a recovery batch is arbitrary. This is acceptable — the user sees them as "recovered" sessions, and new sessions spawned after startup will have accurate timestamps.
- **Duplicate protection**: `registerSession` uses the session name as map key, so calling `recoverSessions()` multiple times is idempotent.

### Risks & Edge Cases

- **Non-mashed tmux sessions**: Sessions named `term-*` or `mashed-*` that were created outside mashed will be picked up. This is intentional — better to show an extra session than miss a real one.
- **Stale working directory**: If the user `cd`'d inside the tmux session, `pane_current_path` reflects the current (not original) directory. This is acceptable and often more useful.
- **Large number of sessions**: Unlikely in practice (< 20), but the implementation shells out once per session for `display-message`. For 20 sessions this is ~200ms total, which is fine during startup.

### Reference Files

- `app.go:104-139` — `startup()` method where `recoverSessions()` will be called
- `app_terminal_registry.go` — file created in Story 1, add `recoverSessions()` here
- `app_tmux.go:21` — session naming pattern (`prefix-repoName-timestamp`)
- `internal/terminal/panes.go` — `discoverPanes()` for tmux command patterns

## Acceptance Criteria

AC-1: Recovery scans tmux on startup
- Given tmux has sessions "term-myrepo-1000" and "mashed-other-2000" running
- When the app starts and `recoverSessions()` executes
- Then both sessions are registered in `a.terminalSessions`
- And their `RepoPath` fields are populated from `pane_current_path`

AC-2: Non-matching sessions are ignored
- Given tmux has sessions "term-myrepo-1000" and "my-personal-session"
- When `recoverSessions()` executes
- Then only "term-myrepo-1000" is registered
- And "my-personal-session" is not in the registry

AC-3: Recovery handles tmux not running
- Given tmux server is not running
- When `recoverSessions()` executes
- Then no error is raised
- And the registry remains empty
- And a log message is emitted

AC-4: Recovery is called during app startup
- Given the app is starting
- When `startup()` runs
- Then `recoverSessions()` is called before the first scan cycle

## BDD Test Scenarios

### Scenario 1: Successful Recovery

```gherkin
Feature: Startup session recovery

  Scenario: Recover terminal and agent sessions
    Given tmux is running with sessions "term-repo1-1000" and "mashed-repo2-2000"
    And "term-repo1-1000" has pane_current_path "/dev/repo1"
    And "mashed-repo2-2000" has pane_current_path "/dev/repo2"
    When recoverSessions() is called
    Then the registry contains 2 sessions
    And session "term-repo1-1000" has RepoPath "/dev/repo1" and SessionType "terminal"
    And session "mashed-repo2-2000" has RepoPath "/dev/repo2" and SessionType "agent"
```

### Scenario 2: Filter Non-Matching

```gherkin
Feature: Recovery prefix filtering

  Scenario: Only recover sessions with known prefixes
    Given tmux is running with sessions "term-foo-1", "mashed-bar-2", "irssi", "dev-session"
    When recoverSessions() is called
    Then the registry contains exactly 2 sessions
    And session names are "term-foo-1" and "mashed-bar-2"
```

### Scenario 3: tmux Not Available

```gherkin
Feature: Recovery when tmux is not running

  Scenario: Graceful handling when tmux server is down
    Given tmux server is not running
    When recoverSessions() is called
    Then the registry is empty
    And no panic occurs
    And a warning is logged
```

### Scenario 4: Idempotent Recovery

```gherkin
Feature: Recovery is idempotent

  Scenario: Calling recoverSessions twice does not duplicate entries
    Given tmux has session "term-foo-100"
    When recoverSessions() is called twice
    Then the registry contains exactly 1 session
```

## Tasks / Subtasks

- [ ] Task 1: Implement recoverSessions (AC: AC-1, AC-2, AC-3)
  - [ ] Add `recoverSessions()` method to `app_terminal_registry.go`
  - [ ] Shell out to `tmux list-sessions -F '#{session_name}'`
  - [ ] Filter by `term-` and `mashed-` prefixes
  - [ ] Query `tmux display-message -t {name} -p '#{pane_current_path}'` for each
  - [ ] Build and register `TerminalSession` entries
  - [ ] Handle `tmux list-sessions` failure gracefully (log + return)

- [ ] Task 2: Wire into startup (AC: AC-4)
  - [ ] Call `a.recoverSessions()` in `app.go` `startup()` after bridge start

- [ ] Task 3: Write tests (AC: AC-1, AC-2, AC-3)
  - [ ] Add test cases to `app_terminal_registry_test.go` for recovery
  - [ ] Test prefix filtering logic (unit test the filter, not the tmux command)
  - [ ] Test idempotency of registration

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
