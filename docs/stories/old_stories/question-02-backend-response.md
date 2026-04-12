# Story 2: Backend Response Injection

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** Story 1 (question-01-backend-detection)
**Status:** done

## Description

Add the ability to inject a user's answer back into a running Claude CLI tmux session that is blocked on a question. This introduces a `RespondToQuestion()` method on the Executor that validates the target pane is still alive, escapes the answer for safe tmux transmission, sends it via `tmux send-keys -l`, and clears the question hash. A corresponding Wails binding in `app_bmad.go` exposes this to the frontend.

## Developer Notes

### Architecture

- **Modified file:** `internal/bmad/executor.go` -- add `RespondToQuestion(execID, nodeID, answer string) error` method.
- **Modified file:** `internal/bmad/question.go` -- add `escapeTmuxLiteral(s string) string` helper for safe `send-keys -l` input.
- **Modified file:** `app_bmad.go` -- add `RespondToQuestion(execID, nodeID, answer string) error` Wails binding.
- **Modified file:** `internal/bmad/question_test.go` -- add tests for `escapeTmuxLiteral`.
- **Modified file:** `internal/bmad/executor_test.go` -- add tests for `RespondToQuestion` method.

### Key Method Signatures

```go
// executor.go
func (e *Executor) RespondToQuestion(execID, nodeID, answer string) error

// question.go
func escapeTmuxLiteral(s string) string

// app_bmad.go
func (a *App) RespondToQuestion(execID, nodeID, answer string) error
```

### Technical Considerations

- **Pane liveness check:** Before sending keys, verify the pane is still alive: `tmux list-panes -t {target} -F "#{pane_dead}"` must return `"0"`. If dead, return a descriptive error (e.g., `fmt.Errorf("bmad: node %s pane is dead, cannot respond: %w", nodeID, ErrExecNotRunning)`).
- **tmux send-keys escaping:** `tmux send-keys -l` sends literal characters but single quotes and backslashes need escaping. `escapeTmuxLiteral` should handle: single quotes (replace `'` with `'\''`), and ensure no shell injection. The answer is sent as: `tmux send-keys -l -t {target} {escaped}` followed by `tmux send-keys -t {target} Enter`.
- **Two-command sequence:** The literal text and Enter must be separate `send-keys` calls. The first uses `-l` for literal mode; the second sends `Enter` without `-l`.
- **Hash clearing:** After successful send, clear `lastQuestionHash[nodeID]` so the polling loop can detect a new question if one appears.
- **Accessing `lastQuestionHash`:** This map lives on `execState`. The `RespondToQuestion` method must lock `state.mu` to access it.
- **Error sentinel reuse:** Use existing `ErrExecNotFound` and `ErrExecNotRunning` from `types.go`.

### Risks & Edge Cases

- **Race between response and completion:** The node could complete between the pane-alive check and the send-keys call. This is a TOCTOU race. Mitigation: if `send-keys` fails, return an error but do not panic -- the node is already completing normally.
- **Empty answer:** If the user submits an empty string for a freeform question, still send Enter (selects default). `escapeTmuxLiteral("")` returns `""`, and the Enter send-keys proceeds.
- **Very long answers:** tmux `send-keys -l` handles long strings, but extremely long input (>4096 chars) may be truncated by tmux. Document this limitation; cap at 4096 chars with a descriptive error.
- **Menu option answers:** For menu questions, the frontend sends the option number (e.g., `"1"`), not the option text. This is just a short string.

### Reference Files

- `internal/bmad/executor.go` lines 196-210 -- `GetExecution` pattern for accessing `execState` safely
- `internal/bmad/executor.go` lines 730-815 -- `executeNode` for understanding tmux target format
- `app_bmad.go` -- all existing Wails bindings follow nil-check-then-delegate pattern
- `internal/bmad/executor_test.go` -- `newHarness` and mock `CommandRunner` pattern

## Acceptance Criteria

AC-1: Response is injected into the correct tmux pane
- Given a running workflow node blocked on a question in tmux session `bmad-nodeA-12345:0.0`
- When `RespondToQuestion(execID, "nodeA", "my answer")` is called
- Then `tmux send-keys -l -t bmad-nodeA-12345:0.0 'my answer'` is executed
- And `tmux send-keys -t bmad-nodeA-12345:0.0 Enter` is executed

AC-2: Dead pane returns an error
- Given a workflow node whose tmux pane reports `pane_dead=1`
- When `RespondToQuestion(execID, nodeID, "answer")` is called
- Then an error wrapping `ErrExecNotRunning` is returned
- And no `send-keys` command is executed

AC-3: Special characters in answers are safely escaped
- Given an answer containing single quotes, backslashes, and special shell characters (e.g., `it's a "test" with $vars`)
- When `escapeTmuxLiteral()` is called
- Then the returned string is safe for `tmux send-keys -l`
- And the original meaning of the text is preserved

AC-4: Question hash is cleared after successful response
- Given a node with `lastQuestionHash[nodeID]` set to a non-empty value
- When `RespondToQuestion(execID, nodeID, "answer")` succeeds
- Then `lastQuestionHash[nodeID]` is cleared (set to `""`)

AC-5: Wails binding delegates correctly
- Given `a.bmadExecutor` is initialized
- When `App.RespondToQuestion(execID, nodeID, answer)` is called
- Then it delegates to `a.bmadExecutor.RespondToQuestion(execID, nodeID, answer)`
- And returns the same error (or nil)

AC-6: Wails binding returns error when executor is nil
- Given `a.bmadExecutor` is nil
- When `App.RespondToQuestion(execID, nodeID, answer)` is called
- Then an error "bmad not initialized" is returned

## BDD Test Scenarios

### Scenario 1: Successful Response Injection

```gherkin
Feature: Inject answer into tmux session

  Scenario: Send freeform text answer
    Given a running execution "exec-1" with node "node-A" targeting "bmad-node-A-100:0.0"
    And the mock command runner returns "0" for pane_dead check
    When RespondToQuestion("exec-1", "node-A", "src/main.go") is called
    Then runCmd is called with "tmux", "send-keys", "-l", "-t", "bmad-node-A-100:0.0", "src/main.go"
    And runCmd is called with "tmux", "send-keys", "-t", "bmad-node-A-100:0.0", "Enter"
    And the returned error is nil

  Scenario: Send menu option number
    Given a running execution with node "node-B"
    And the mock command runner returns "0" for pane_dead check
    When RespondToQuestion is called with answer "2"
    Then the literal "2" is sent via send-keys -l
    And Enter is sent
```

### Scenario 2: Dead Pane Rejection

```gherkin
Feature: Reject response to dead pane

  Scenario: Pane is dead
    Given a running execution "exec-1" with node "node-A"
    And the mock command runner returns "1" for pane_dead check
    When RespondToQuestion("exec-1", "node-A", "answer") is called
    Then the returned error wraps ErrExecNotRunning
    And runCmd is NOT called with "send-keys"

  Scenario: Pane check command fails entirely
    Given a running execution "exec-1" with node "node-A"
    And the mock command runner returns an error for the pane_dead check
    When RespondToQuestion("exec-1", "node-A", "answer") is called
    Then the returned error indicates the pane is unreachable
    And runCmd is NOT called with "send-keys"
```

### Scenario 3: Special Character Escaping

```gherkin
Feature: tmux literal escaping

  Scenario: Answer with single quotes
    Given answer "it's fine"
    When escapeTmuxLiteral is called
    Then the result safely wraps single quotes for tmux

  Scenario: Answer with backslashes
    Given answer "path\\to\\file"
    When escapeTmuxLiteral is called
    Then backslashes are preserved literally

  Scenario: Empty answer
    Given answer ""
    When escapeTmuxLiteral is called
    Then the result is ""
    And RespondToQuestion still sends Enter
```

### Scenario 4: Hash Clearing

```gherkin
Feature: Clear question hash after response

  Scenario: Hash cleared on success
    Given node "node-A" has lastQuestionHash["node-A"] = "abc123"
    And the pane is alive
    When RespondToQuestion succeeds
    Then lastQuestionHash["node-A"] is ""

  Scenario: Hash NOT cleared on failure
    Given node "node-A" has lastQuestionHash["node-A"] = "abc123"
    And the pane is dead
    When RespondToQuestion fails
    Then lastQuestionHash["node-A"] remains "abc123"
```

## Tasks / Subtasks

- [ ] Task 1: Implement `escapeTmuxLiteral()` in `question.go` (AC: AC-3)
  - [ ] Subtask 1a: Handle single quotes, backslashes, and special characters
  - [ ] Subtask 1b: Add table-driven tests in `question_test.go`

- [ ] Task 2: Implement `RespondToQuestion()` on Executor (AC: AC-1, AC-2, AC-4)
  - [ ] Subtask 2a: Look up `execState` by `execID`, find node by `nodeID` to get `TmuxTarget`
  - [ ] Subtask 2b: Check pane liveness via `tmux list-panes -t {target} -F "#{pane_dead}"`
  - [ ] Subtask 2c: Send answer via two `runCmd` calls: `send-keys -l` then `send-keys Enter`
  - [ ] Subtask 2d: Clear `lastQuestionHash[nodeID]` under lock on success

- [ ] Task 3: Add Wails binding in `app_bmad.go` (AC: AC-5, AC-6)
  - [ ] Subtask 3a: Add `RespondToQuestion(execID, nodeID, answer string) error` method on `App`
  - [ ] Subtask 3b: Nil-check `a.bmadExecutor`, delegate to executor method

- [ ] Task 4: Write tests (AC: AC-1 through AC-4)
  - [ ] Subtask 4a: Test successful response injection with mock command runner
  - [ ] Subtask 4b: Test dead pane rejection
  - [ ] Subtask 4c: Test hash clearing behavior
  - [ ] Subtask 4d: Test answer length cap (>4096 chars returns error)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified code
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
