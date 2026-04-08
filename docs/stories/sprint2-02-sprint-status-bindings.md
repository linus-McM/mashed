# Story 2: Wails Bindings for Sprint Status

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** Story 1
**Status:** ready

## Description

Expose the sprint status parser to the frontend by adding Wails-bound methods to `app.go`. This creates the bridge that lets the Svelte UI read sprint data from any repo and update story statuses. The frontend cannot display or interact with sprint data without these bindings.

## Developer Notes

### Architecture
- **Modify:** `app.go` -- add 3 new methods in a new `// -- BMAD Sprint Status --` section after the existing BMAD Modules section (~line 1718)
- All methods follow the existing pattern: nil-check on dependencies, delegate to `bmad` package functions, return typed results
- No new struct fields on `App` needed -- `ParseSprintStatus` and `UpdateStoryStatus` are stateless functions that operate on the filesystem
- Wails auto-generates TypeScript bindings in `frontend/wailsjs/go/main/App.js` on next `wails dev` run

### Methods to Add

```go
// GetSprintStatus parses and returns the sprint-status.yaml for a repo.
func (a *App) GetSprintStatus(repoPath string) (bmad.SprintStatus, error) {
    return bmad.ParseSprintStatus(repoPath)
}

// UpdateStoryStatus modifies a story's status in sprint-status.yaml.
func (a *App) UpdateStoryStatus(repoPath, storyID, newStatus string) error {
    return bmad.UpdateStoryStatus(repoPath, storyID, newStatus)
}

// WatchSprintStatus starts watching sprint-status.yaml for changes and emits events.
// Returns a cancel function ID for cleanup.
func (a *App) WatchSprintStatus(repoPath string) (string, error) { ... }
```

### Technical Considerations
- `UpdateStoryStatus` must be implemented in `internal/bmad/sprint.go` as well -- it reads the YAML, modifies the target story's status in-place, and writes it back atomically (same pattern as `atomicWriteJSON` in `storage.go` but for YAML)
- For `UpdateStoryStatus`, preserve the original YAML formatting as much as possible. Use `yaml.Node` round-trip: unmarshal to Node, modify the value node, marshal back.
- `WatchSprintStatus` uses `fsnotify` (already in go.mod) to watch the YAML file. On change, it re-parses and emits a Wails event `bmad:sprint:updated` with the new `SprintStatus`. Store watchers in a map on the App struct with cancel functions.
- Input validation: `repoPath` must be non-empty and an absolute path. `storyID` must match the `\d+-\d+-.*` pattern. `newStatus` must pass `ValidateStoryStatus`.

### Risks & Edge Cases
- Race condition on concurrent `UpdateStoryStatus` calls -- use a mutex in the update function (or per-repo mutex)
- File watcher may fire multiple events for a single save -- debounce with 500ms window
- YAML file could be deleted while being watched -- handle gracefully, emit error event

### Reference Files
- `app.go` lines 1547-1718 -- existing BMAD binding patterns (nil checks, error wrapping)
- `internal/bmad/sprint.go` -- parser from Story 1
- `internal/bmad/storage.go` -- `atomicWriteJSON` pattern for atomic writes

## Acceptance Criteria

AC-1: GetSprintStatus binding works end-to-end
- Given a repo path with a valid sprint-status.yaml
- When `GetSprintStatus(repoPath)` is called from the frontend
- Then it returns the parsed `SprintStatus` as a JSON-serializable object

AC-2: UpdateStoryStatus modifies YAML on disk
- Given a repo path with sprint-status.yaml containing story "1-2-account-mgmt" with status "backlog"
- When `UpdateStoryStatus(repoPath, "1-2-account-mgmt", "in-progress")` is called
- Then the YAML file on disk reflects the updated status
- And re-parsing the file returns the story with status "in-progress"

AC-3: Input validation rejects invalid parameters
- Given an empty repoPath or invalid storyID format or unrecognized status
- When `UpdateStoryStatus` is called
- Then it returns a descriptive validation error without modifying the file

AC-4: GetSprintStatus returns error for missing file
- Given a repo path with no sprint-status.yaml
- When `GetSprintStatus(repoPath)` is called from the frontend
- Then it returns an error that the frontend can display

## BDD Test Scenarios

### Scenario 1: Sprint status round-trip

```gherkin
Feature: Sprint status Wails bindings

  Scenario: Read sprint status through binding
    Given a repo at "/tmp/test-repo" with a sprint-status.yaml containing 2 epics
    When GetSprintStatus("/tmp/test-repo") is called
    Then the returned SprintStatus has Project set and 2 epics with correct stories

  Scenario: Update story status and re-read
    Given a repo with sprint-status.yaml containing story "1-2-acct" at "backlog"
    When UpdateStoryStatus(repoPath, "1-2-acct", "ready-for-dev") is called
    Then the call succeeds with no error
    And GetSprintStatus returns story "1-2-acct" with status "ready-for-dev"

  Scenario: Update with invalid status string
    Given a repo with a valid sprint-status.yaml
    When UpdateStoryStatus(repoPath, "1-1-auth", "invalid-status") is called
    Then the error contains "invalid status"
    And the YAML file is unchanged

  Scenario: Update nonexistent story
    Given a repo with a valid sprint-status.yaml that has no story "99-99-nope"
    When UpdateStoryStatus(repoPath, "99-99-nope", "done") is called
    Then the error contains "story not found"
```

## Tasks / Subtasks

- [ ] Task 1: Implement UpdateStoryStatus in bmad package (AC: AC-2, AC-3)
  - [ ] Subtask 1a: Add `UpdateStoryStatus(repoPath, storyID, newStatus string) error` to `internal/bmad/sprint.go`
  - [ ] Subtask 1b: Use yaml.Node round-trip to preserve formatting during update
  - [ ] Subtask 1c: Add file-level mutex for concurrent write safety
  - [ ] Subtask 1d: Add validation for storyID format and status value

- [ ] Task 2: Add Wails bindings to app.go (AC: AC-1, AC-4)
  - [ ] Subtask 2a: Add `GetSprintStatus(repoPath string)` method to App
  - [ ] Subtask 2b: Add `UpdateStoryStatus(repoPath, storyID, newStatus string)` method to App
  - [ ] Subtask 2c: Add section comment `// -- BMAD Sprint Status --` for organization

- [ ] Task 3: Write tests for UpdateStoryStatus (AC: AC-2, AC-3)
  - [ ] Subtask 3a: Test successful status update with round-trip verification
  - [ ] Subtask 3b: Test validation errors (bad status, bad story ID, missing file)
  - [ ] Subtask 3c: Test that original YAML structure is preserved after update

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
