# Story 5: Screenshot-to-Claude-Code — Full Stack

**Priority:** P1-high
**Domain:** full-stack (backend + frontend)
**Estimated Complexity:** M
**Depends On:** none (but should land AFTER pty-04 to avoid Terminal.svelte merge conflicts)
**Status:** done

---

## Description

Implement the complete screenshot-to-Claude-Code injection pipeline in a single story. Modify `TakeScreenshot` to save screenshots to `{repoPath}/.screenshots/` (gitignored), emit a scoped `screenshot:inject` Wails event, add `SetActiveContext` for the frontend to push active repo/terminal context to Go, and wire `Terminal.svelte` to receive the event and inject the path via WebSocket bracketed paste.

**Merged from:** `meditor-05-screenshot-backend.md` + `meditor-06-screenshot-frontend.md`

**Landing note:** This story should land AFTER `pty-04-app-integration-and-frontend-cleanup` because both modify `Terminal.svelte` and `AgentDetail.svelte`. Landing PTY first means this story works on the already-cleaned-up versions of those files (no `stripControlSequences`, no `:0.0` suffixes).

---

## Developer Notes

### Backend Changes

- **Modify file:** `app.go`
  - Add `activeRepoPath string` and `activePaneTarget string` fields to App struct (protected by existing `a.mu` mutex)
  - Add `SetActiveContext(repoPath, paneTarget string)` method — mutex-protected setter
  - Modify `TakeScreenshot` signature: `TakeScreenshot(repoPath, paneTarget string) (string, error)`
    - Save to `{repoPath}/.screenshots/screenshot-{timestamp}.png`
    - Emit `screenshot:inject` event with `map[string]string{"path": path, "paneTarget": paneTarget}`
    - Preserve existing `screenshot:taken` event for toast

- **Modify file:** `main.go`
  - Update menu callback at line 53 to read cached `activeRepoPath`/`activePaneTarget` under mutex
  - Guard: if `activeRepoPath == ""`, log and return without calling TakeScreenshot

- **Gitignore auto-append:** Read `.gitignore`, check if `.screenshots/` present, append if not. Use `os.OpenFile` with `O_APPEND|O_CREATE|O_WRONLY`.

- **User cancellation:** `screencapture -i` returns exit code 1 when user cancels. Return `("", nil)`.

### Frontend Changes

- **Modify file:** `frontend/src/components/Terminal.svelte`
  - Add scoped `screenshot:inject` event listener inside `onMount` (after WebSocket opens)
  - Guard: `if (data.paneTarget !== paneTarget) return` — only matching terminal responds
  - Guard: `if (!ws || ws.readyState !== WebSocket.OPEN) return` — skip if WS not ready
  - Send path as bracketed paste + Enter via `ws.send(encoder.encode(...))`
  - Store unsubscribe function, call in `onDestroy`

- **Modify file:** `frontend/src/views/AgentDetail.svelte`
  - Import `SetActiveContext` from Wails bindings
  - Add reactive statement: `$: if (agent?.repoPath && activeSession?.paneTarget) SetActiveContext(agent.repoPath, activeSession.paneTarget)`

- **Modify file:** `frontend/src/App.svelte`
  - Add `SetActiveContext('', '')` call in `goBack()` function

### Technical Considerations
- **Concurrency:** `activeRepoPath`/`activePaneTarget` written by Wails JS thread, read by macOS menu thread. Protect with `a.mu`.
- **Event listener lifecycle:** Subscribe in `onMount`, unsubscribe in `onDestroy`. Use Wails `EventsOn` return value.
- **Bracketed paste format:** Same as existing `pasteToTerminal`: `\x1b[200~` + path + `\x1b[201~` then `\r`
- **Multiple terminals:** paneTarget matching ensures only the active terminal responds.
- **PTY dependency note:** After pty-04 lands, Terminal.svelte no longer has `stripControlSequences`. The screenshot listener is purely additive — a new event subscription alongside existing WebSocket setup.

### Risks & Edge Cases
- **No active repo:** User presses Cmd+Shift+S on the feed. Menu callback logs and returns silently.
- **Claude Code not at prompt:** screencapture takes focus, so Claude is unlikely to be streaming.
- **WebSocket not yet open:** Guard with readyState check, log warning.
- **Gitignore race:** Two rapid screenshots could both append. `strings.Contains` check makes duplicate harmless.
- **Existing tests:** `screenshot_test.go` must be updated for new `TakeScreenshot` signature.

### Reference Files
- `app.go:181-209` — current `TakeScreenshot` implementation
- `main.go:53-57` — current menu callback
- `screenshot_test.go` — existing tests to update
- `frontend/src/components/Terminal.svelte:109-115` — existing `pasteToTerminal` pattern
- `frontend/src/components/Terminal.svelte:164-229` — WebSocket setup
- `frontend/src/views/AgentDetail.svelte:85-102` — session tab state
- `frontend/src/App.svelte:92-96` — existing `screenshot:taken` listener
- `docs/feasibility-multi-editor.md` — Sections 8.4-8.6

---

## Acceptance Criteria

### Backend

AC-1: Screenshot saves to repo directory
- Given `activeRepoPath` is set to `/Users/dev/myrepo`
- When the user takes a screenshot via Cmd+Shift+S
- Then the screenshot is saved to `/Users/dev/myrepo/.screenshots/screenshot-{timestamp}.png`
- And the `.screenshots/` directory is created if it does not exist

AC-2: Scoped event emission
- Given `activePaneTarget` is set to `"myrepo:agent-1"`
- When a screenshot is successfully saved
- Then `screenshot:inject` is emitted with `{ path, paneTarget }`
- And `screenshot:taken` is emitted with path (preserving toast)

AC-3: SetActiveContext method
- Given the App is running
- When `SetActiveContext("myrepo", "myrepo:agent-1")` is called
- Then both fields are set, mutex-protected

AC-4: Clear context
- Given context is set
- When `SetActiveContext("", "")` is called
- Then both fields are cleared
- And subsequent screenshots log a warning and skip

AC-5: Gitignore auto-append
- Given the repo `.gitignore` does not contain `.screenshots/`
- When the first screenshot is taken
- Then `.screenshots/` is appended with a comment header
- And subsequent screenshots do not duplicate the entry

AC-6: User cancellation
- Given the user presses Escape during area selection
- Then `TakeScreenshot` returns `("", nil)` and no event is emitted

AC-7: Menu callback uses cached context
- Given cached context values
- When the menu callback fires
- Then it reads under mutex and passes to `TakeScreenshot(repoPath, paneTarget)`

### Frontend

AC-8: Screenshot path injection
- Given Terminal.svelte with matching paneTarget and open WebSocket
- When `screenshot:inject` fires with matching paneTarget
- Then the path is sent as bracketed paste + Enter via WebSocket

AC-9: Scoped event matching
- Given two Terminal instances with different paneTargets
- When `screenshot:inject` fires for one
- Then only the matching terminal sends the path

AC-10: SetActiveContext on mount/session change
- Given AgentDetail mounted for repo "myrepo"
- When active session tab changes
- Then `SetActiveContext` is called with updated values

AC-11: Clear context on navigation
- Given user viewing AgentDetail
- When user navigates back to feed
- Then `SetActiveContext("", "")` is called

AC-12: Event listener cleanup
- Given Terminal.svelte has subscribed to `screenshot:inject`
- When the component is destroyed
- Then the listener is unsubscribed

AC-13: WebSocket not ready guard
- Given WebSocket is not yet open
- When `screenshot:inject` fires with matching paneTarget
- Then no data is sent and a warning is logged

---

## BDD Test Scenarios

### Scenario 1: Screenshot saving
```gherkin
Feature: Screenshot saves to repo .screenshots directory

  Scenario: Save screenshot to active repo
    Given activeRepoPath is "/Users/dev/myrepo"
    And activePaneTarget is "myrepo:agent-1"
    When TakeScreenshot is called with repoPath and paneTarget
    Then os.MkdirAll is called for "/Users/dev/myrepo/.screenshots/"
    And screencapture -i -x is called with a path under .screenshots/
    And the filename matches "screenshot-YYYYMMDD-HHMMSS.png"

  Scenario: User cancels screenshot
    Given screencapture returns exit code 1
    When TakeScreenshot is called
    Then the return value is ("", nil)
    And no Wails events are emitted
```

### Scenario 2: SetActiveContext
```gherkin
Feature: Active context tracking

  Scenario: Set active context
    Given App is initialized
    When SetActiveContext("myrepo", "myrepo:agent-1") is called
    Then a.activeRepoPath equals "myrepo" and a.activePaneTarget equals "myrepo:agent-1"

  Scenario: Clear active context
    When SetActiveContext("", "") is called
    Then both fields are empty

  Scenario: Concurrent access is safe
    Given two goroutines call SetActiveContext and read simultaneously
    When run with -race flag
    Then no race condition is detected
```

### Scenario 3: Gitignore management
```gherkin
Feature: Auto-append .screenshots/ to .gitignore

  Scenario: Gitignore does not exist
    Given no .gitignore file exists
    When TakeScreenshot creates the .screenshots/ directory
    Then a new .gitignore is created containing ".screenshots/"

  Scenario: Gitignore exists without .screenshots/
    Given .gitignore contains "node_modules/" but not ".screenshots/"
    When TakeScreenshot runs
    Then ".screenshots/" is appended

  Scenario: Gitignore already contains .screenshots/
    Given .gitignore already contains ".screenshots/"
    When TakeScreenshot runs
    Then .gitignore is not modified
```

### Scenario 4: Event emission
```gherkin
Feature: Wails event emission for screenshot inject

  Scenario: Emit scoped inject event
    Given a screenshot was saved and paneTarget is "myrepo:agent-1"
    When TakeScreenshot completes successfully
    Then "screenshot:inject" event is emitted with {path, paneTarget}
    And "screenshot:taken" event is emitted with the path string

  Scenario: No event on empty context
    Given activeRepoPath is ""
    When the menu callback fires
    Then no TakeScreenshot call is made and no events are emitted
```

### Scenario 5: Frontend injection
```gherkin
Feature: Screenshot path injection into terminal

  Scenario: Inject screenshot path into matching terminal
    Given Terminal.svelte with paneTarget="myrepo:agent-1" and open WebSocket
    When "screenshot:inject" event fires with matching paneTarget
    Then ws.send is called with bracketed paste bytes for the path
    And ws.send is called with carriage return bytes

  Scenario: Ignore non-matching paneTarget
    Given Terminal.svelte with paneTarget="myrepo:agent-1"
    When "screenshot:inject" fires with paneTarget="myrepo:agent-2"
    Then ws.send is not called

  Scenario: Skip injection when WebSocket not ready
    Given WebSocket is in CONNECTING state
    When "screenshot:inject" fires with matching paneTarget
    Then ws.send is not called and a console.warn is logged
```

### Scenario 6: Context tracking from frontend
```gherkin
Feature: Frontend pushes active context to Go

  Scenario: Set context on AgentDetail mount
    Given user navigates to AgentDetail for repo "myrepo"
    And active session has paneTarget "myrepo:agent-1"
    When AgentDetail mounts
    Then SetActiveContext("myrepo", "myrepo:agent-1") is called

  Scenario: Update context on session tab switch
    Given AgentDetail is mounted
    When user clicks the tab for "myrepo:agent-2"
    Then SetActiveContext is called with updated paneTarget

  Scenario: Clear context on goBack
    Given user is viewing AgentDetail
    When user clicks back to the feed
    Then SetActiveContext("", "") is called
```

### Scenario 7: Cleanup
```gherkin
Feature: Event listener cleanup on destroy

  Scenario: Unsubscribe on destroy
    Given Terminal.svelte has subscribed to "screenshot:inject"
    When the component is destroyed
    Then the EventsOn unsubscribe function is called
```

---

## Tasks / Subtasks

### Phase A: Backend

- [ ] Task 1: Add active context fields and method (AC: 3, 4)
  - [ ] Subtask 1a: Add `activeRepoPath string` and `activePaneTarget string` fields to App struct
  - [ ] Subtask 1b: Implement `SetActiveContext(repoPath, paneTarget string)` with mutex protection
  - [ ] Subtask 1c: Write tests for SetActiveContext (set, clear, concurrent -race)

- [ ] Task 2: Modify TakeScreenshot (AC: 1, 2, 5, 6)
  - [ ] Subtask 2a: Change signature to `TakeScreenshot(repoPath, paneTarget string) (string, error)`
  - [ ] Subtask 2b: Replace `~/Desktop` save path with `{repoPath}/.screenshots/screenshot-{timestamp}.png`
  - [ ] Subtask 2c: Add `os.MkdirAll` for `.screenshots/` directory
  - [ ] Subtask 2d: Add gitignore auto-append logic
  - [ ] Subtask 2e: Emit `screenshot:inject` event with path and paneTarget
  - [ ] Subtask 2f: Preserve existing `screenshot:taken` event

- [ ] Task 3: Update menu callback (AC: 7)
  - [ ] Subtask 3a: Modify menu callback in `main.go` to read cached context under mutex
  - [ ] Subtask 3b: Guard: if `activeRepoPath == ""`, log and return
  - [ ] Subtask 3c: Pass cached values to `TakeScreenshot(rp, pt)`

- [ ] Task 4: Update existing tests (AC: 1, 2, 6)
  - [ ] Subtask 4a: Update `screenshot_test.go` for new signature
  - [ ] Subtask 4b: Add test for gitignore append behavior
  - [ ] Subtask 4c: Verify user cancellation still returns `("", nil)`

### Phase B: Frontend

- [ ] Task 5: Add screenshot:inject listener to Terminal.svelte (AC: 8, 9, 12, 13)
  - [ ] Subtask 5a: Subscribe to `screenshot:inject` inside `onMount`, after WebSocket setup
  - [ ] Subtask 5b: Implement paneTarget matching guard
  - [ ] Subtask 5c: Implement WebSocket readiness guard
  - [ ] Subtask 5d: Send path as bracketed paste + Enter via `ws.send()`
  - [ ] Subtask 5e: Store unsubscribe function and call in `onDestroy`

- [ ] Task 6: Add SetActiveContext calls to AgentDetail.svelte (AC: 10)
  - [ ] Subtask 6a: Import `SetActiveContext` from Wails bindings
  - [ ] Subtask 6b: Add reactive statement for context tracking

- [ ] Task 7: Clear context on navigation in App.svelte (AC: 11)
  - [ ] Subtask 7a: Import `SetActiveContext` and call `('', '')` in `goBack()`

---

## Files Modified

| File | Action |
|------|--------|
| `app.go` | Add context fields, SetActiveContext, modify TakeScreenshot |
| `main.go` | Update menu callback with cached context |
| `screenshot_test.go` | Update for new TakeScreenshot signature |
| `frontend/src/components/Terminal.svelte` | Add screenshot:inject event listener |
| `frontend/src/views/AgentDetail.svelte` | Add SetActiveContext reactive call |
| `frontend/src/App.svelte` | Clear context on goBack() |

---

## Definition of Done

- [ ] All 13 acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] Screenshot inject works end-to-end: Cmd+Shift+S -> path appears in terminal
- [ ] Only matching terminal receives the injected path
- [ ] Event listener properly cleaned up on component destroy
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
