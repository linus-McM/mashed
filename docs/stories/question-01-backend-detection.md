# Story 1: Backend Question Detection

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Add the ability for the BMAD executor to detect when a Claude CLI process running in a tmux pane asks the user a question (via `AskUserQuestion`). This introduces a `QuestionEvent` struct, ANSI stripping, regex-based question pattern detection, and integrates question scanning into the existing `executeNode()` polling loop with content-hash deduplication. When a question is detected, the executor emits a `bmad:node:question` event; when the node completes or fails, it emits `bmad:node:question:dismissed` to clear stale notifications.

## Developer Notes

### Architecture

- **New file:** `internal/bmad/question.go` -- contains `QuestionEvent` struct, `detectQuestion()`, `stripANSI()`, `hashQuestion()`, and `captureQuestionOutput()` (a lighter capture for question scanning).
- **New file:** `internal/bmad/question_test.go` -- table-driven tests for all detection functions.
- **Modified file:** `internal/bmad/executor.go` -- add `lastQuestionHash` field to `execState` struct; add question polling every 3rd tick inside the `executeNode()` loop when `pane_dead == "0"`; emit `bmad:node:question:dismissed` in `completeNode()` and `failNode()`.
- **Modified file:** `internal/bmad/types.go` -- no changes needed; `QuestionEvent` lives in `question.go` within the same package.

### Key Types

```go
// question.go
type QuestionEvent struct {
    ExecID     string   `json:"execId"`
    NodeID     string   `json:"nodeId"`
    RepoPath   string   `json:"repoPath"`
    RepoName   string   `json:"repoName"`
    Question   string   `json:"question"`
    Options    []string `json:"options"`    // menu options; empty for freeform
    TmuxTarget string   `json:"tmuxTarget"`
    Timestamp  int64    `json:"timestamp"`  // unix millis
    QuestionID string   `json:"questionId"` // content hash for dedup
}

func detectQuestion(output string) (question string, options []string, found bool)
func stripANSI(s string) string
func hashQuestion(q string) string
```

### Technical Considerations

- **ANSI stripping:** Claude CLI output contains ANSI escape sequences (colors, cursor movement). Use `regexp.MustCompile("\\x1b\\[[0-9;]*[a-zA-Z]")` compiled once at package level.
- **Detection heuristics:** Claude CLI `AskUserQuestion` renders a bordered box with `?` prefix or `>` prompt. Scan the last 50 lines of captured output for these patterns. Also detect numbered menu options (e.g., `1. Option A`, `2. Option B`).
- **`captureQuestionOutput()`:** Uses `tmux capture-pane -t {target} -p -S -200` (last 200 lines) -- lighter than the full `-S -5000` capture used for output collection.
- **Dedup via hash:** `hashQuestion()` uses `crypto/sha256` on the question text. Store in `execState.lastQuestionHash[nodeID]`. Only emit a new event when hash changes.
- **Polling frequency:** Every 3rd tick of the existing poll loop (so ~9 seconds at the default 3s interval). Add a `questionPollCounter` local variable inside `executeNode()`.
- **Concurrency:** `lastQuestionHash` is accessed only within `executeNode()` (single goroutine per node), so no additional synchronization needed beyond the existing `state.mu` for emitting events.
- **`completeNode()` and `failNode()` changes:** After existing logic, emit `bmad:node:question:dismissed` with `{ execId, nodeId }` to clear any stale snackbar on the frontend.

### Risks & Edge Cases

- Claude CLI may change its question rendering format -- keep regex patterns flexible and document them.
- A question may span multiple lines; `detectQuestion` should join lines between the `?` prefix and the next prompt/border.
- If the tmux session is killed externally, `captureQuestionOutput` will fail silently (err is logged, not propagated).
- Multiple rapid questions from the same node: the hash will change, triggering a new event and effectively replacing the old one.

### Reference Files

- `internal/bmad/executor.go` lines 718-815 -- existing `captureOutput` and `executeNode` polling loop
- `internal/bmad/executor_test.go` -- test harness pattern (`newHarness`, `eventRecord`, `eventsByName`)
- `internal/bmad/condition.go` -- example of a well-structured utility in the bmad package

## Acceptance Criteria

AC-1: ANSI escape sequences are stripped from captured output
- Given a string containing ANSI color codes, cursor movement, and formatting escapes
- When `stripANSI()` is called
- Then the returned string contains only printable text with no escape sequences

AC-2: Freeform questions are detected from Claude CLI output
- Given tmux captured output containing a Claude CLI `AskUserQuestion` prompt (bordered box with `?` prefix)
- When `detectQuestion()` is called with the output
- Then `found` is true, `question` contains the extracted question text, and `options` is empty

AC-3: Menu-style questions with numbered options are detected
- Given tmux captured output containing a Claude CLI question with numbered options (e.g., `1. Option A`)
- When `detectQuestion()` is called with the output
- Then `found` is true, `question` contains the question text, and `options` contains each menu option string

AC-4: Question events are emitted during node execution polling
- Given a running workflow node in a tmux session where Claude asks a question
- When the executor's polling loop runs every 3rd tick
- Then a `bmad:node:question` event is emitted with a valid `QuestionEvent`
- And the event includes `execId`, `nodeId`, `repoPath`, `repoName`, `question`, `tmuxTarget`, `timestamp`, and `questionId`

AC-5: Duplicate questions are not re-emitted
- Given a question has already been detected and emitted for a node
- When the next question poll captures the same question text (same hash)
- Then no new `bmad:node:question` event is emitted

AC-6: Stale question notifications are dismissed on node completion
- Given a question event was previously emitted for a running node
- When the node completes (success or failure)
- Then a `bmad:node:question:dismissed` event is emitted with `{ execId, nodeId }`

AC-7: Output without questions produces no false positives
- Given tmux captured output that contains normal Claude CLI working output (no question)
- When `detectQuestion()` is called
- Then `found` is false

## BDD Test Scenarios

### Scenario 1: ANSI Stripping

```gherkin
Feature: ANSI escape code removal

  Scenario: Strip color codes from terminal output
    Given a string "\x1b[32mHello\x1b[0m \x1b[1;31mWorld\x1b[0m"
    When stripANSI is called
    Then the result is "Hello World"

  Scenario: Strip cursor movement sequences
    Given a string "\x1b[2J\x1b[H? What should I do?"
    When stripANSI is called
    Then the result is "? What should I do?"

  Scenario: No-op on clean strings
    Given a string "Just plain text"
    When stripANSI is called
    Then the result is "Just plain text"
```

### Scenario 2: Question Detection

```gherkin
Feature: Claude CLI question detection

  Scenario: Detect freeform question with ? prefix
    Given captured output containing lines:
      """
      Working on the implementation...
      
      ? What directory should I create the files in?
      
      > 
      """
    When detectQuestion is called
    Then found is true
    And question is "What directory should I create the files in?"
    And options is empty

  Scenario: Detect menu question with numbered options
    Given captured output containing lines:
      """
      ? How would you like to proceed?
      1. Create new file
      2. Modify existing file
      3. Skip this step
      """
    When detectQuestion is called
    Then found is true
    And question is "How would you like to proceed?"
    And options contains ["Create new file", "Modify existing file", "Skip this step"]

  Scenario: No question in normal output
    Given captured output containing lines:
      """
      Reading file src/main.go...
      Analyzing code structure...
      Found 3 functions to modify.
      """
    When detectQuestion is called
    Then found is false

  Scenario: Only last 50 lines are scanned
    Given captured output with 200 lines where a question appears at line 10
    When detectQuestion is called
    Then found is false
    And the question at line 10 is outside the scan window
```

### Scenario 3: Deduplication and Event Emission

```gherkin
Feature: Question event deduplication

  Scenario: First question detection emits event
    Given a running workflow node with no prior question hash
    When captureQuestionOutput returns output containing a question
    And detectQuestion returns found=true with question "What file?"
    Then a bmad:node:question event is emitted
    And lastQuestionHash[nodeID] is set to hashQuestion("What file?")

  Scenario: Same question on next poll is suppressed
    Given lastQuestionHash[nodeID] equals hashQuestion("What file?")
    When captureQuestionOutput returns the same question "What file?"
    Then no bmad:node:question event is emitted

  Scenario: Different question replaces hash and emits
    Given lastQuestionHash[nodeID] equals hashQuestion("What file?")
    When captureQuestionOutput returns a new question "Which branch?"
    Then a bmad:node:question event is emitted with question "Which branch?"
    And lastQuestionHash[nodeID] is updated to hashQuestion("Which branch?")
```

### Scenario 4: Dismissal on Completion

```gherkin
Feature: Question dismissal on node lifecycle events

  Scenario: Node completion dismisses question
    Given a workflow node that previously emitted a question event
    When the node's tmux pane reports pane_dead=1
    Then a bmad:node:question:dismissed event is emitted with execId and nodeId
    And the node is marked complete

  Scenario: Node failure dismisses question
    Given a workflow node that previously emitted a question event
    When the node's execution context is cancelled
    Then a bmad:node:question:dismissed event is emitted with execId and nodeId
    And the node is marked failed
```

## Tasks / Subtasks

- [ ] Task 1: Implement `question.go` utility functions (AC: AC-1, AC-2, AC-3, AC-7)
  - [ ] Subtask 1a: Implement `stripANSI()` with compiled regex at package level
  - [ ] Subtask 1b: Implement `detectQuestion()` -- strip ANSI, take last 50 lines, scan for `?` prefix and numbered options
  - [ ] Subtask 1c: Implement `hashQuestion()` using `crypto/sha256`
  - [ ] Subtask 1d: Define `QuestionEvent` struct with JSON tags

- [ ] Task 2: Implement `captureQuestionOutput()` on Executor (AC: AC-4)
  - [ ] Subtask 2a: Add `captureQuestionOutput(ctx, target)` method using `tmux capture-pane -p -S -200`
  - [ ] Subtask 2b: Cap output at `maxCaptureBytes` (reuse existing constant)

- [ ] Task 3: Integrate question polling into `executeNode()` (AC: AC-4, AC-5)
  - [ ] Subtask 3a: Add `lastQuestionHash map[string]string` field to `execState` struct, initialize in `StartWorkflow`
  - [ ] Subtask 3b: Add `questionPollCounter` variable and every-3rd-tick logic in the polling loop
  - [ ] Subtask 3c: Call `captureQuestionOutput`, `detectQuestion`, hash-compare, and emit `bmad:node:question` event

- [ ] Task 4: Emit dismissal events on node lifecycle (AC: AC-6)
  - [ ] Subtask 4a: Add `bmad:node:question:dismissed` emission to `completeNode()`
  - [ ] Subtask 4b: Add `bmad:node:question:dismissed` emission to `failNode()`

- [x] Task 5: Write comprehensive tests (AC: AC-1 through AC-7)
  - [x] Subtask 5a: Table-driven tests for `stripANSI` (clean, colors, cursor, mixed)
  - [x] Subtask 5b: Table-driven tests for `detectQuestion` (freeform, menu, no question, edge cases)
  - [x] Subtask 5c: Integration test using `testHarness` for question event emission during mock node execution
  - [x] Subtask 5d: Integration test verifying dismissal events on node complete/fail

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `question.go` and modified sections of `executor.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
