# Story 5: Screenshot-to-Claude-Code -- Backend Changes

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Modify the `TakeScreenshot` method to save screenshots to `{repoPath}/.screenshots/` (gitignored) and emit a scoped `screenshot:inject` Wails event with the path and paneTarget. Add a `SetActiveContext` method so the frontend can push the active repo and terminal context to Go, enabling the menu callback to route screenshots to the correct terminal session.

## Developer Notes

### Architecture
- **Modified files:**
  - `app.go` -- add `activeRepoPath` and `activePaneTarget` fields to App struct (lines 33-51), add `SetActiveContext` method
  - `app.go` -- modify `TakeScreenshot` signature: `TakeScreenshot(repoPath, paneTarget string)` -> saves to repo `.screenshots/`, emits scoped event (lines 181-209)
  - `main.go` -- update menu callback at line 53 to use cached active context instead of calling `TakeScreenshot()` with no args
- **Data flow:**
  1. Frontend calls `SetActiveContext(repoPath, paneTarget)` whenever active view/session changes
  2. User presses `Cmd+Shift+S` -> menu callback reads cached `activeRepoPath`/`activePaneTarget`
  3. `TakeScreenshot(repoPath, paneTarget)` saves to `{repoPath}/.screenshots/screenshot-{timestamp}.png`
  4. Go emits `screenshot:inject` event with `{ path, paneTarget }`
  5. Go also emits `screenshot:taken` event with path (existing toast behavior preserved)

### Technical Considerations
- **Concurrency:** `activeRepoPath` and `activePaneTarget` are written by Wails binding calls (from JS main thread) and read by the menu callback (from macOS main thread). Protect with `a.mu` (existing `sync.Mutex` on App struct).
- **Directory creation:** `os.MkdirAll(filepath.Join(repoPath, ".screenshots"), 0755)` on each screenshot
- **Gitignore auto-append:** Read `.gitignore`, check if `.screenshots/` already present, append if not. Use `os.OpenFile` with `O_APPEND|O_CREATE|O_WRONLY`.
- **User cancellation:** `screencapture -i` returns exit code 1 when user cancels (presses Escape). Continue returning `("", nil)` for cancellation.
- **Empty context:** If `activeRepoPath` is empty when menu fires, log warning and skip. Do not fall back to Desktop save.
- **Event payload:** Emit `screenshot:inject` as a Go `map[string]string` with keys `"path"` and `"paneTarget"`. Wails serializes this as a JSON object to JavaScript.

### Risks & Edge Cases
- **No active repo:** User presses Cmd+Shift+S on the notification feed (no agent detail view). Menu callback should log and return silently.
- **Repo path gone:** User navigates away right before screenshot completes. The save path may still be valid (repo directory exists). If not, `os.MkdirAll` will fail and that error propagates.
- **Gitignore race:** Two rapid screenshots could both try to append to `.gitignore`. The mutex protects context reads but not gitignore writes. Use `strings.Contains` check before each append -- worst case is a duplicate line, which is harmless.
- **Existing screenshot tests:** `screenshot_test.go` has tests for the current `TakeScreenshot()`. These must be updated for the new signature.

### Reference Files
- `app.go:181-209` -- current `TakeScreenshot` implementation
- `main.go:53-57` -- current menu callback for screenshot
- `screenshot_test.go` -- existing tests to update
- `docs/feasibility-multi-editor.md` -- Sections 8.4, 8.5, 8.6

## Acceptance Criteria

AC-1: Screenshot saves to repo directory
- Given `activeRepoPath` is set to `/Users/dev/myrepo`
- When the user takes a screenshot via Cmd+Shift+S
- Then the screenshot is saved to `/Users/dev/myrepo/.screenshots/screenshot-{timestamp}.png`
- And the `.screenshots/` directory is created if it does not exist

AC-2: Scoped event emission
- Given `activePaneTarget` is set to `"myrepo:agent-1"`
- When a screenshot is successfully saved
- Then a `screenshot:inject` Wails event is emitted with `{ path: "/full/path.png", paneTarget: "myrepo:agent-1" }`
- And a `screenshot:taken` Wails event is emitted with the path (preserving existing toast behavior)

AC-3: SetActiveContext method
- Given the App is running
- When `SetActiveContext("myrepo", "myrepo:agent-1")` is called from the frontend
- Then `a.activeRepoPath` is set to `"myrepo"`
- And `a.activePaneTarget` is set to `"myrepo:agent-1"`
- And the values are thread-safe (mutex-protected)

AC-4: Clear context
- Given `activeRepoPath` is set to a repo
- When `SetActiveContext("", "")` is called
- Then both fields are cleared
- And subsequent screenshot attempts log a warning and return without saving

AC-5: Gitignore auto-append
- Given the repo has a `.gitignore` that does not contain `.screenshots/`
- When the first screenshot is taken for that repo
- Then `.screenshots/` is appended to `.gitignore` with a comment header
- And subsequent screenshots do not duplicate the entry

AC-6: User cancellation
- Given the user presses Cmd+Shift+S
- When the user presses Escape during area selection
- Then `TakeScreenshot` returns `("", nil)`
- And no event is emitted

AC-7: Menu callback uses cached context
- Given `activeRepoPath` and `activePaneTarget` are cached in the App struct
- When the menu callback fires for "Take Screenshot"
- Then it reads the cached values under mutex
- And passes them to `TakeScreenshot(repoPath, paneTarget)`

## BDD Test Scenarios

### Scenario 1: Screenshot saving
```gherkin
Feature: Screenshot saves to repo .screenshots directory

  Scenario: Save screenshot to active repo
    Given activeRepoPath is "/Users/dev/myrepo"
    And activePaneTarget is "myrepo:agent-1"
    When TakeScreenshot is called with repoPath="/Users/dev/myrepo" and paneTarget="myrepo:agent-1"
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
    Then a.activeRepoPath equals "myrepo"
    And a.activePaneTarget equals "myrepo:agent-1"

  Scenario: Clear active context
    Given activeRepoPath is "myrepo"
    When SetActiveContext("", "") is called
    Then a.activeRepoPath equals ""
    And a.activePaneTarget equals ""

  Scenario: Concurrent access is safe
    Given two goroutines call SetActiveContext and read activeRepoPath simultaneously
    When run with -race flag
    Then no race condition is detected
```

### Scenario 3: Gitignore management
```gherkin
Feature: Auto-append .screenshots/ to .gitignore

  Scenario: Gitignore does not exist
    Given no .gitignore file exists in the repo
    When TakeScreenshot creates the .screenshots/ directory
    Then a new .gitignore is created containing ".screenshots/"

  Scenario: Gitignore exists without .screenshots/
    Given .gitignore contains "node_modules/" but not ".screenshots/"
    When TakeScreenshot runs
    Then ".screenshots/" is appended to .gitignore

  Scenario: Gitignore already contains .screenshots/
    Given .gitignore already contains ".screenshots/"
    When TakeScreenshot runs
    Then .gitignore is not modified
```

### Scenario 4: Event emission
```gherkin
Feature: Wails event emission for screenshot inject

  Scenario: Emit scoped inject event
    Given a screenshot was saved to "/repo/.screenshots/screenshot-20260409-143000.png"
    And paneTarget is "myrepo:agent-1"
    When TakeScreenshot completes successfully
    Then "screenshot:inject" event is emitted with {path: "/repo/.screenshots/screenshot-20260409-143000.png", paneTarget: "myrepo:agent-1"}
    And "screenshot:taken" event is emitted with the path string

  Scenario: No event on empty context
    Given activeRepoPath is ""
    When the menu callback fires
    Then no TakeScreenshot call is made
    And no events are emitted
```

## Tasks / Subtasks

- [ ] Task 1: Add active context fields and method (AC: 3, 4)
  - [ ] Subtask 1a: Add `activeRepoPath string` and `activePaneTarget string` fields to App struct in `app.go`
  - [ ] Subtask 1b: Implement `SetActiveContext(repoPath, paneTarget string)` with mutex protection
  - [ ] Subtask 1c: Write tests for SetActiveContext (set, clear, concurrent access with -race)

- [ ] Task 2: Modify TakeScreenshot (AC: 1, 2, 5, 6)
  - [ ] Subtask 2a: Change signature to `TakeScreenshot(repoPath, paneTarget string) (string, error)`
  - [ ] Subtask 2b: Replace `~/Desktop` save path with `{repoPath}/.screenshots/screenshot-{timestamp}.png`
  - [ ] Subtask 2c: Add `os.MkdirAll` for `.screenshots/` directory
  - [ ] Subtask 2d: Add gitignore auto-append logic (check for `.screenshots/`, append if missing)
  - [ ] Subtask 2e: Emit `screenshot:inject` event with `map[string]string{"path": path, "paneTarget": paneTarget}`
  - [ ] Subtask 2f: Preserve existing `screenshot:taken` event emission

- [ ] Task 3: Update menu callback (AC: 7)
  - [ ] Subtask 3a: Modify menu callback in `main.go:53` to read cached context under mutex
  - [ ] Subtask 3b: Add guard: if `activeRepoPath == ""`, log and return without calling TakeScreenshot
  - [ ] Subtask 3c: Pass cached values to `TakeScreenshot(rp, pt)`

- [ ] Task 4: Update existing tests (AC: 1, 2, 6)
  - [ ] Subtask 4a: Update `screenshot_test.go` for new `TakeScreenshot` signature
  - [ ] Subtask 4b: Add test for gitignore append behavior
  - [ ] Subtask 4c: Add test for event emission (may require mocking runtime.EventsEmit)
  - [ ] Subtask 4d: Verify user cancellation still returns `("", nil)`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] Existing screenshot tests updated and passing
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
