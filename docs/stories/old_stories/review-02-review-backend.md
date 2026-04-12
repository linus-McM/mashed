# Story 2: Code Review Summary & Advice Streaming Backend

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** L
**Depends On:** review-01
**Status:** done

## Description

Implement the `app_review.go` backend that orchestrates the code review summarisation workflow: streaming per-file diff summaries via Claude CLI, streaming methodology advice using advice mode prompts, and listing available advice modes. All operations emit Wails events for real-time frontend consumption. This is the core engine that powers the summarisation modal.

## Developer Notes

### Architecture

- **New file:** `app_review.go` -- all review-related Wails-bound methods on `*App`
- **New file:** `app_review_test.go` -- tests (unit tests for helpers, integration-style tests for streaming with mocked CLI)

- **Types to define in `app_review.go`:**

```go
// FileSummary is a per-file AI-generated change summary.
type FileSummary struct {
    Path     string `json:"path"`
    Added    int    `json:"added"`
    Removed  int    `json:"removed"`
    Summary  string `json:"summary"`
    IsBinary bool   `json:"isBinary"`
}

// ReviewSummary is the complete review result.
type ReviewSummary struct {
    Files      []FileSummary `json:"files"`
    TotalAdded int           `json:"totalAdded"`
    TotalRemoved int         `json:"totalRemoved"`
}
```

- **Wails-bound methods on `*App`:**

```go
// ListAdviceModes returns available methodology advice modes for a repo.
func (a *App) ListAdviceModes(repoPath string) ([]advice.AdviceMode, error)

// StreamCodeReviewSummary starts async per-file summarisation.
// Emits "review:summary:progress" events per file, "review:summary:done" when complete.
func (a *App) StreamCodeReviewSummary(repoPath string)

// StreamAdvice starts async methodology advice streaming.
// Emits "review:advice:progress" events with streaming text chunks.
func (a *App) StreamAdvice(repoPath, modeName string)
```

### Data Flow

**StreamCodeReviewSummary:**
1. Call `git.ScopedDiff(repoPath)` to get `[]DiffFileStat`
2. For each file, call `ReadFileDiff(repoPath, file.Path)` to get the diff text
3. For each file, spawn `claude -p "Summarise this diff..."` with the diff as input
4. Emit `review:summary:progress` event with `FileSummary` after each file completes
5. Emit `review:summary:done` event with the full `ReviewSummary`

**StreamAdvice:**
1. Call `advice.LoadAdviceBody(repoPath, modeName)` to get the system prompt
2. Build a combined prompt: advice body as system context + full diff + file summaries
3. Spawn `claude -p` with the combined prompt, capture output
4. Emit `review:advice:progress` events with streaming text chunks (read stdout line-by-line)
5. Final event includes `done: true`

### Technical Considerations

- **Goroutines:** Both `StreamCodeReviewSummary` and `StreamAdvice` run in goroutines (same pattern as `GitCommitStreaming`). Use `go func() { ... }()` immediately.
- **Event emission pattern:** Follow `GitCommitStreaming` exactly -- define an `emit` closure that calls `runtime.EventsEmit(a.ctx, eventName, map[string]interface{}{...})`.
- **Event payloads:**
  - `review:summary:progress`: `{ repoPath, file: FileSummary, index: int, total: int, error: string }`
  - `review:summary:done`: `{ repoPath, summary: ReviewSummary, error: string }`
  - `review:advice:progress`: `{ repoPath, text: string, done: bool, error: string }`
- **Claude CLI invocation:** Use `exec.CommandContext(a.ctx, "claude", "-p", prompt)` with `cmd.Dir = repoPath`. For streaming advice, use `cmd.StdoutPipe()` + `bufio.Scanner` to read line-by-line.
- **Rate limiting:** Process files sequentially (not in parallel) to avoid overloading Claude CLI. One file at a time.
- **Diff truncation:** Truncate individual file diffs to 500 lines before sending to Claude (reuse `truncateHunk` pattern from explain package, or inline).
- **Binary files:** Skip binary files (already flagged in `DiffFileStat.IsBinary`), emit a summary of "Binary file changed".
- **Context cancellation:** Check `a.ctx.Done()` between files to allow cancellation.

### Risks & Edge Cases

- **No changes in repo:** If `ScopedDiff` returns empty file list, emit `review:summary:done` immediately with empty summary
- **Claude CLI not available:** Check `exec.LookPath("claude")` at start, emit error event if missing
- **Very large diffs:** Truncate to 500 lines per file; if total files > 50, only summarise first 50
- **Claude CLI timeout:** Use `context.WithTimeout` (60s per file) to prevent hanging
- **Concurrent calls:** Only one review can run at a time per repo. Use a `sync.Map` keyed by repoPath to track active reviews; reject if one is already running.

### Reference Files

- `app_git.go` lines 466-513 (`GitCommitStreaming`) -- streaming goroutine + event emission pattern
- `app_explain.go` -- Claude CLI invocation pattern
- `internal/explain/explain.go` -- prompt construction, `exec.CommandContext`, output parsing
- `internal/git/diff.go` -- `ScopedDiff` function
- `internal/domain/types.go` lines 203-214 -- `ScopedDiff`, `DiffFileStat` types

## Acceptance Criteria

AC-1: List advice modes
- Given the advice backend (story review-01) is implemented with at least one bundled default
- When `ListAdviceModes(repoPath)` is called
- Then it returns a non-empty slice of `AdviceMode` structs with `Name`, `DisplayName`, `Icon`, and `Order`

AC-2: Stream per-file summaries
- Given a repo with 3 changed files (2 modified, 1 new)
- When `StreamCodeReviewSummary(repoPath)` is called
- Then 3 `review:summary:progress` events are emitted, each containing a `FileSummary` with `Path`, `Added`, `Removed`, and a non-empty `Summary`
- And 1 `review:summary:done` event is emitted with the complete `ReviewSummary`

AC-3: Stream methodology advice
- Given a valid advice mode name "clean-code" and a repo with changes
- When `StreamAdvice(repoPath, "clean-code")` is called
- Then `review:advice:progress` events are emitted with streaming text
- And the final event has `done: true`

AC-4: Error handling -- no Claude CLI
- Given the `claude` CLI is not on PATH
- When `StreamCodeReviewSummary(repoPath)` is called
- Then a `review:summary:done` event is emitted with an error message indicating Claude CLI is not available

AC-5: Error handling -- no changes
- Given a repo with a clean working tree (no modifications)
- When `StreamCodeReviewSummary(repoPath)` is called
- Then a `review:summary:done` event is emitted with an empty `ReviewSummary` and no error

AC-6: Binary file handling
- Given a repo with a modified binary file (e.g., an image)
- When `StreamCodeReviewSummary` processes that file
- Then the emitted `FileSummary` has `IsBinary: true` and `Summary` set to "Binary file changed"
- And no Claude CLI call is made for that file

## BDD Test Scenarios

### Scenario 1: Streaming summary happy path

```gherkin
Feature: Code review summary streaming

  Scenario: Summarise 2 modified files
    Given a repo at "/tmp/test-repo" with git initialized
    And file "main.go" has staged changes adding 10 lines and removing 3 lines
    And file "util.go" has staged changes adding 5 lines
    When StreamCodeReviewSummary is called with "/tmp/test-repo"
    Then 2 "review:summary:progress" events are emitted
    And event 1 has index 0 and total 2
    And event 2 has index 1 and total 2
    And 1 "review:summary:done" event is emitted
    And the done event has totalAdded 15 and totalRemoved 3
```

### Scenario 2: Advice streaming

```gherkin
Feature: Methodology advice streaming

  Scenario: Stream advice with clean-code perspective
    Given advice mode "clean-code" exists with a markdown body
    And the repo has uncommitted changes
    When StreamAdvice is called with repoPath and "clean-code"
    Then at least 1 "review:advice:progress" event is emitted with non-empty text
    And the final event has done set to true

  Scenario: Invalid advice mode name
    Given no advice mode named "nonexistent" exists
    When StreamAdvice is called with repoPath and "nonexistent"
    Then a "review:advice:progress" event is emitted with done true and an error message
```

### Scenario 3: Edge cases

```gherkin
Feature: Review edge cases

  Scenario: Empty diff
    Given a repo with no uncommitted changes
    When StreamCodeReviewSummary is called
    Then "review:summary:done" is emitted with files as empty array and no error

  Scenario: Binary file skipped
    Given a repo with one binary file change and one text file change
    When StreamCodeReviewSummary processes the binary file
    Then the FileSummary for the binary file has summary "Binary file changed"
    And no Claude subprocess is spawned for the binary file

  Scenario: Concurrent review rejected
    Given a review is already running for repo "/tmp/repo"
    When StreamCodeReviewSummary is called again for "/tmp/repo"
    Then "review:summary:done" is emitted with error "review already in progress"
```

## Tasks / Subtasks

- [ ] Task 1: Create `app_review.go` with types and ListAdviceModes (AC: AC-1)
  - [ ] Define `FileSummary` and `ReviewSummary` structs
  - [ ] Implement `ListAdviceModes` delegating to `advice.LoadAdviceModes`
  - [ ] Add `activeReviews sync.Map` field to `App` struct in `app.go` (or local to `app_review.go` as package-level var)

- [ ] Task 2: Implement StreamCodeReviewSummary (AC: AC-2, AC-4, AC-5, AC-6)
  - [ ] Implement the streaming goroutine with event emission
  - [ ] Add Claude CLI availability check at start
  - [ ] Implement per-file diff retrieval and truncation
  - [ ] Handle binary files by emitting "Binary file changed"
  - [ ] Handle empty diff (no files changed)
  - [ ] Add concurrent review guard via `activeReviews`
  - [ ] Add context cancellation checks between files

- [ ] Task 3: Implement StreamAdvice (AC: AC-3)
  - [ ] Load advice body via `advice.LoadAdviceBody`
  - [ ] Construct combined prompt (advice body + diff + summaries)
  - [ ] Spawn Claude CLI with `StdoutPipe` for line-by-line streaming
  - [ ] Emit `review:advice:progress` events per line
  - [ ] Handle invalid mode name with error event

- [ ] Task 4: Write tests (AC: AC-1 through AC-6)
  - [ ] Unit tests for `FileSummary`/`ReviewSummary` construction
  - [ ] Test `ListAdviceModes` delegates correctly
  - [ ] Test empty diff produces immediate done event
  - [ ] Test binary file handling
  - [ ] Test concurrent review guard
  - [ ] Test advice mode not found error path

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `app_review.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
