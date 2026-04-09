# Story 5: Scan Loop Dual PID Lookup

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** pty-02, pty-04
**Status:** done

## Description

Update the scan loop in `app_scan.go` to look up running agent processes in the `SessionManager` first (for sessions spawned by mashed), then fall back to tmux pane discovery (for external tmux sessions like BMAD executor sessions). This ensures the process scanner can associate agents with their terminal sessions regardless of how they were spawned.

## Developer Notes

### Architecture
- **Modify file:** `app_scan.go` (219 lines)
- **Change location:** `doScan()` method, lines 134-141 (tmux pane lookup block)
- **Current logic (lines 134-141):**
  ```go
  var tmuxTarget string
  if pane, err := a.panes.FindPaneForPID(s.PID); err == nil && pane != nil {
      tmuxTarget = pane.Target()
  }
  if tmuxTarget == "" {
      continue
  }
  ```
- **New logic:**
  ```go
  var tmuxTarget string
  // First: check if this PID belongs to a session we spawned
  if sess, ok := a.manager.FindByPID(s.PID); ok {
      tmuxTarget = sess.Name() // session name is the WS route
  }
  // Fallback: check tmux panes for externally spawned sessions (BMAD, manual tmux)
  if tmuxTarget == "" {
      if pane, err := a.panes.FindPaneForPID(s.PID); err == nil && pane != nil {
          tmuxTarget = pane.Target()
      }
  }
  if tmuxTarget == "" {
      continue
  }
  ```
- **`ManagedSession.Name()` method:** Needs to be exposed from Story 1. Should return the session's `name` field (the same key used in the manager's map). Verify this getter exists or add it.

### Technical Considerations
- **No behavior change for BMAD sessions:** BMAD executor spawns its own tmux sessions (in `internal/bmad/executor.go`). These are NOT managed by `SessionManager`. The tmux pane fallback ensures they continue to be discovered.
- **Performance:** `FindByPID` iterates all managed sessions (typically <10). This runs every 5 seconds in `scanLoop`. Negligible cost.
- **PID matching accuracy:** `FindByPID` checks `cmd.Process.Pid` which is the direct child PID. The scan finds claude processes which may be grandchildren. However, the existing `FindPaneForPID` already walks the PPID chain, and `FindByPID` should check the direct PID. If the scanned PID is a grandchild, the tmux fallback will catch it. This is acceptable for the transition period.

### Risks & Edge Cases
- **Manager session with same PID as tmux pane:** Shouldn't happen -- managed sessions don't run under tmux. But if it did, manager takes priority (correct behavior since we want to route to the managed PTY).
- **Session dies between FindByPID and doScan completing:** The session name is still valid; the bridge's `handleWS` will return 404 if the session is gone by the time the frontend connects. Frontend already handles `[disconnected]` gracefully.

### Reference Files
- `app_scan.go` (lines 57-218, especially 134-141 for the tmux lookup block)
- `internal/terminal/manager.go` (from Story 2) -- `FindByPID` method
- `internal/terminal/session.go` (from Story 1) -- `Name()` getter
- `internal/bmad/executor.go` -- uses tmux directly (NOT affected by this change)

## Acceptance Criteria

AC-1: Manager-spawned agents are found via FindByPID
- Given a claude process with PID 1234 running inside a managed session "mashed-repo-100"
- When `doScan()` runs
- Then `tmuxTarget` is set to "mashed-repo-100" (from manager)
- And the agent is processed with `HasTmuxPane: true`

AC-2: External tmux agents fall back to pane discovery
- Given a claude process with PID 5678 running in a manually created tmux session
- And that PID is NOT in any managed session
- When `doScan()` runs
- Then `tmuxTarget` is set from `panes.FindPaneForPID(5678)` (tmux fallback)

AC-3: Manager lookup takes priority over tmux pane lookup
- Given a PID that exists in both a managed session and a tmux pane
- When `doScan()` runs
- Then the managed session name is used (not the tmux pane target)

AC-4: Agents without any session are still skipped
- Given a claude process with PID 9999 not in manager or tmux
- When `doScan()` runs
- Then the agent is skipped (continue statement reached)

## BDD Test Scenarios

### Scenario 1: Dual Lookup Priority

```gherkin
Feature: Dual PID lookup in scan

  Scenario: Manager-spawned agent found first
    Given a managed session "mashed-repo-100" with process PID 1234
    And no tmux pane for PID 1234
    When doScan processes PID 1234
    Then tmuxTarget is "mashed-repo-100"

  Scenario: External agent falls back to tmux
    Given no managed session for PID 5678
    And a tmux pane "external-sess:0.0" contains PID 5678
    When doScan processes PID 5678
    Then tmuxTarget is "external-sess:0.0"

  Scenario: Agent with no session is skipped
    Given no managed session for PID 9999
    And no tmux pane for PID 9999
    When doScan processes PID 9999
    Then the agent is not included in notifications
```

### Scenario 2: BMAD Compatibility

```gherkin
Feature: BMAD executor compatibility

  Scenario: BMAD agent in tmux is still discovered
    Given a BMAD executor running claude in tmux session "bmad-workflow-1"
    And the claude process PID is not in SessionManager
    When doScan runs
    Then the BMAD agent is discovered via pane fallback
    And it appears in notifications with tmuxTarget "bmad-workflow-1:0.0"
```

## Tasks / Subtasks

- [x] Task 1: Add `Name()` getter to ManagedSession if not present (AC: AC-1)
  - [x] Subtask 1a: Verify `ManagedSession` has a `Name() string` method; add if missing
  - [x] Subtask 1b: Ensure the name matches the key used in `SessionManager.sessions` map

- [x] Task 2: Update `doScan()` with dual PID lookup (AC: AC-1, AC-2, AC-3, AC-4)
  - [x] Subtask 2a: Add `a.manager.FindByPID(s.PID)` check before tmux fallback
  - [x] Subtask 2b: Use `sess.Name()` as `tmuxTarget` for managed sessions
  - [x] Subtask 2c: Keep existing `a.panes.FindPaneForPID(s.PID)` as fallback
  - [x] Subtask 2d: Keep `continue` for agents without any session

- [x] Task 3: Write tests for dual lookup (AC: AC-1, AC-2, AC-3)
  - [x] Subtask 3a: Test with mock manager that returns a session for known PID
  - [x] Subtask 3b: Test fallback to pane discovery when manager returns nil
  - [x] Subtask 3c: Test skip when neither manager nor panes match

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ code coverage on modified `doScan()` code paths
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `/simplify` run on all modified code
- [x] Code review: no CRITICAL/HIGH issues
