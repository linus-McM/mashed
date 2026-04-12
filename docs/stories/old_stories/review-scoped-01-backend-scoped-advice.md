# Story 1: StreamScopedAdvice Backend Method

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Add a new `StreamScopedAdvice` method to `app_review.go` that generates methodology-based code review advice scoped to a specific set of file paths rather than the entire git diff. This enables the frontend to request advice on user-selected files only, reducing noise and improving review relevance. The method also accepts an `additionalContext` parameter for incremental context chaining (prior advice, file summaries).

## Developer Notes

### Architecture

- **File to modify:** `/Users/linus/Development/mashed/app_review.go`
- **Test file to modify:** `/Users/linus/Development/mashed/app_review_test.go`
- **New method signature:**
  ```go
  func (a *App) StreamScopedAdvice(repoPath, modeName, model string, filePaths []string, additionalContext string)
  ```
- The method is structurally identical to `StreamAdvice` (line 237) except for diff generation:
  - Instead of `git diff HEAD` (full diff), run `git diff HEAD -- file1.go file2.go ...`
  - For untracked files (no output from `git diff HEAD --`), fall back to `git diff --no-index /dev/null <abs-path>` per the `ReadFileDiff` pattern at `app_git.go:813`
  - Prepend `additionalContext` before the concatenated diff in the stdin payload
- Reuses the same `review:advice:progress` event channel as `StreamAdvice`
- Uses the existing `claudeCommand` helper from `app_claude.go`
- Uses the existing `advice.LoadAdviceBody` for loading the methodology body

### Technical Considerations

- **Diff assembly:** Iterate `filePaths`, call `git diff HEAD -- <path>` for each. If output is empty, try `--no-index /dev/null <abs-path>`. Concatenate all diffs with a `\n---\n` separator.
- **Empty diff guard:** If the concatenated diff is empty (all selected files have no changes), emit a done event with text "No changes found for selected files." and return.
- **Input validation:** `filePaths` must be non-empty. If empty, emit an error event and return early.
- **additionalContext prepend:** If non-empty, prepend it before the diff block with a clear separator: `"## Prior Context\n{additionalContext}\n\n## Code Changes\n{diff}"`.
- **Concurrency:** This method runs in its own goroutine. It does NOT share the `activeReviews` sync.Map guard -- advice streams are independent of summary streams and multiple advice calls are expected (user re-generates with different selections).
- **Existing `StreamAdvice` is unchanged** -- backward compatibility preserved.
- **Wails binding regeneration:** After adding the method, `wails dev` must be restarted to generate `frontend/wailsjs/go/main/App.js` and `App.d.ts` bindings.

### Risks & Edge Cases

- **Path traversal:** Validate that each `filePath` is relative (no `..` or absolute paths) to prevent reading outside the repo.
- **Large file set:** If `filePaths` has many entries, the concatenated diff could be very large. Cap at `maxDiffLines` (500) per file, same as `StreamCodeReviewSummary`.
- **Untracked file with spaces in name:** Ensure paths are properly quoted/escaped when passed to git commands.
- **Git not initialized:** `git diff` will fail -- error is caught by existing error handling and emitted via event.
- **Claude CLI not found:** Same guard as `StreamAdvice` (line 253).

### Reference Files

- `/Users/linus/Development/mashed/app_review.go` -- `StreamAdvice` method to clone and modify (line 237)
- `/Users/linus/Development/mashed/app_git.go` -- `ReadFileDiff` pattern for untracked fallback (line 813)
- `/Users/linus/Development/mashed/app_claude.go` -- `claudeCommand` helper
- `/Users/linus/Development/mashed/app_review_test.go` -- existing test patterns to follow
- `/Users/linus/Development/mashed/internal/advice/loader.go` -- `LoadAdviceBody` function

## Acceptance Criteria

AC-1: Scoped diff generation
- Given a repo with 5 changed files
- When `StreamScopedAdvice` is called with `filePaths` containing 2 of those file paths
- Then the diff sent to Claude contains only changes from those 2 files
- And the remaining 3 files are not included in the stdin payload

AC-2: Untracked file fallback
- Given a repo with an untracked new file "new_feature.go"
- When `StreamScopedAdvice` is called with `filePaths` containing "new_feature.go"
- Then the method falls back to `git diff --no-index /dev/null <abs-path>` for that file
- And the diff content is included in the stdin payload

AC-3: Additional context prepend
- Given a non-empty `additionalContext` string containing prior advice text
- When `StreamScopedAdvice` is called
- Then the stdin payload starts with "## Prior Context" followed by the additionalContext
- And the diff follows under a "## Code Changes" header

AC-4: Empty file list validation
- Given an empty `filePaths` slice
- When `StreamScopedAdvice` is called
- Then a `review:advice:progress` event is emitted with `done: true` and an error message
- And no Claude CLI process is spawned

AC-5: Event compatibility
- Given the frontend is listening on `review:advice:progress`
- When `StreamScopedAdvice` streams advice
- Then events have the same shape as `StreamAdvice` events (`repoPath`, `text`, `done`, `error`)
- And existing event handlers work without modification

## BDD Test Scenarios

### Scenario 1: Scoped diff assembly

```gherkin
Feature: StreamScopedAdvice diff scoping

  Scenario: Only selected files appear in diff
    Given a git repo at "/tmp/test-repo" with tracked changes in "main.go", "util.go", and "config.go"
    When buildScopedDiff is called with filePaths ["main.go", "util.go"]
    Then the returned diff contains content from "main.go"
    And the returned diff contains content from "util.go"
    And the returned diff does not contain content from "config.go"

  Scenario: Untracked file uses no-index fallback
    Given a git repo at "/tmp/test-repo" with an untracked file "new.go"
    When buildScopedDiff is called with filePaths ["new.go"]
    Then the method executes "git diff --no-index /dev/null" for "new.go"
    And the returned diff is non-empty

  Scenario: Mixed tracked and untracked files
    Given a git repo with tracked change "main.go" and untracked file "new.go"
    When buildScopedDiff is called with filePaths ["main.go", "new.go"]
    Then both files appear in the concatenated diff output
```

### Scenario 2: Context prepend

```gherkin
Feature: Additional context in StreamScopedAdvice

  Scenario: Prior advice is prepended to diff
    Given additionalContext is "Previous advice: refactor error handling in main.go"
    And filePaths is ["main.go"]
    When the stdin payload is assembled
    Then it starts with "## Prior Context"
    And contains "Previous advice: refactor error handling in main.go"
    And contains "## Code Changes" before the diff content

  Scenario: Empty additional context is omitted
    Given additionalContext is ""
    And filePaths is ["main.go"]
    When the stdin payload is assembled
    Then it does not contain "## Prior Context"
    And it starts directly with the diff content
```

### Scenario 3: Validation and error handling

```gherkin
Feature: StreamScopedAdvice input validation

  Scenario: Empty filePaths emits error
    Given filePaths is an empty slice
    When StreamScopedAdvice is called
    Then a review:advice:progress event is emitted with done=true
    And the error field contains "no files selected"

  Scenario: All selected files have no changes
    Given filePaths contains "unchanged.go" which has no diff
    When StreamScopedAdvice assembles the diff
    Then it emits a done event with text "No changes found for selected files."
```

## Tasks / Subtasks

- [ ] Task 1: Implement `buildScopedDiff` helper function (AC: 1, 2)
  - [ ] Subtask 1a: Create `buildScopedDiff(ctx context.Context, repoPath string, filePaths []string) (string, error)` that iterates file paths, runs `git diff HEAD -- <path>` per file, falls back to `--no-index` for empty output
  - [ ] Subtask 1b: Add per-file `truncateDiffLines` cap (reuse existing function with `maxDiffLines`)
  - [ ] Subtask 1c: Concatenate diffs with `\n---\n` separator, return empty string if all diffs empty

- [ ] Task 2: Implement `StreamScopedAdvice` method (AC: 1, 3, 4, 5)
  - [ ] Subtask 2a: Clone `StreamAdvice` method structure, replace `git diff HEAD` with call to `buildScopedDiff`
  - [ ] Subtask 2b: Add `filePaths` validation (empty slice emits error event)
  - [ ] Subtask 2c: Implement `additionalContext` prepend logic with `## Prior Context` / `## Code Changes` headers
  - [ ] Subtask 2d: Handle empty assembled diff (all files unchanged)

- [ ] Task 3: Write unit tests (AC: 1, 2, 3, 4)
  - [ ] Subtask 3a: Table-driven tests for `buildScopedDiff` using `t.TempDir()` with real git repos
  - [ ] Subtask 3b: Tests for context prepend assembly (with and without additionalContext)
  - [ ] Subtask 3c: Tests for empty filePaths validation
  - [ ] Subtask 3d: Tests for empty diff handling

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified code in `app_review.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Wails bindings regenerated (`wails dev` restart confirms `StreamScopedAdvice` appears in `App.d.ts`)
