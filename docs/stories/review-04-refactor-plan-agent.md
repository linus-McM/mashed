# Story 4: Refactor Plan Agent

**Priority:** P1-high
**Domain:** fullstack
**Estimated Complexity:** M
**Depends On:** review-02, review-03
**Status:** done

## Description

Add the "Create Refactor Plan" feature that spawns a Claude agent to generate a structured refactor plan based on the code review summary and methodology advice. The agent writes its output to `{repoPath}/.claude/plans/refactor-{timestamp}.md`. The modal displays a link to the generated plan file once complete, and the user can click to open it in the editor.

## Developer Notes

### Architecture

- **Modified file:** `app_review.go` -- add `SpawnRefactorPlan` method
- **Modified file:** `frontend/src/views/SummarisationModal.svelte` -- add button + plan link display

### Backend: SpawnRefactorPlan

```go
// SpawnRefactorPlan spawns a Claude agent that produces a refactor plan.
// It writes the plan to {repoPath}/.claude/plans/refactor-{timestamp}.md.
// Returns the path to the generated plan file.
func (a *App) SpawnRefactorPlan(repoPath, advice string) (string, error)
```

**Implementation:**
1. Validate inputs: `repoPath` non-empty, `advice` non-empty
2. Generate output path: `{repoPath}/.claude/plans/refactor-{time.Now().Unix()}.md`
3. Ensure `.claude/plans/` directory exists (`os.MkdirAll`)
4. Get the full diff: `git -C {repoPath} diff HEAD`
5. Construct prompt combining: diff summary + advice text + instruction to write a plan
6. Spawn via `a.spawnSession("refactor", repoPath, cmd, domain.SessionAgent, "claude-opus-4-6")`
7. The Claude agent runs autonomously in a PTY session and writes the plan file
8. Return the plan file path immediately (the agent runs async in background)

**Prompt construction:**

```
You are a senior engineer creating a refactor plan. Based on the following code review
advice and diff, create a detailed refactor plan and write it to: {planPath}

## Advice
{advice text}

## Instructions
- Write a markdown plan to the file path above using Write tool
- Structure: Summary, Priority Items, File-by-File Changes, Testing Strategy
- Be specific: include file paths, function names, line references
- Order by priority (critical first, cosmetic last)
```

**CLI command:**
```
claude --dangerously-skip-permissions --model claude-opus-4-6 -p "{prompt}"
```

### Frontend Changes

In `SummarisationModal.svelte`:
- Add "Create Refactor Plan" button in the footer section
- Button is disabled until advice text is non-empty
- On click: call `SpawnRefactorPlan(repoPath, adviceText)`
- Show loading state: "Creating plan..."
- On success: show the plan file path as a clickable link
- Link click dispatches `open-file` event with the plan path

### Technical Considerations

- **Async agent:** The Claude agent runs in a PTY session. `SpawnRefactorPlan` returns the plan file path immediately -- the file won't exist yet. The frontend should show "Agent spawned -- plan will appear at: {path}"
- **Directory creation:** `os.MkdirAll(filepath.Join(repoPath, ".claude", "plans"), 0755)` must run before spawning
- **Timestamp format:** Use Unix timestamp for uniqueness: `refactor-1712764800.md`
- **Error handling:** If `spawnSession` fails, return the error. The frontend shows it inline.
- **Session type:** Use `domain.SessionAgent` so the spawned session appears in the terminal registry and can be viewed/killed

### Risks & Edge Cases

- **Plan directory doesn't exist:** `MkdirAll` handles this
- **Agent fails to write plan:** The user sees the agent session in the terminal tab and can debug. Plan path was shown preemptively.
- **Very long advice text:** Truncate to 10,000 characters to avoid CLI argument length limits. For very long prompts, consider writing to a temp file and using `--input-file` if available.
- **Concurrent plan creation:** Each call generates a unique timestamp filename, so no conflicts
- **Prompt shell escaping:** Use `fmt.Sprintf` with `%q` for the prompt to handle special characters in advice text

### Reference Files

- `app_git.go` lines 853-876 (`SpawnPRReview`) -- exact pattern for spawning a Claude agent via `spawnSession`
- `app_spawn.go` (`spawnSession`) -- managed PTY session creation
- `frontend/src/views/SummarisationModal.svelte` -- modal being modified (from review-03)

## Acceptance Criteria

AC-1: Spawn refactor plan agent
- Given a repo with uncommitted changes and non-empty advice text
- When `SpawnRefactorPlan(repoPath, adviceText)` is called
- Then a Claude agent session is spawned in a managed PTY
- And the function returns the expected plan file path (`{repoPath}/.claude/plans/refactor-{timestamp}.md`)

AC-2: Plans directory created
- Given `{repoPath}/.claude/plans/` does not exist
- When `SpawnRefactorPlan` is called
- Then the directory is created before the agent is spawned

AC-3: Frontend button state
- Given the summarisation modal is open
- When no advice has been generated yet
- Then the "Create Refactor Plan" button is disabled
- And when advice text is available, the button becomes enabled

AC-4: Plan link display
- Given the user clicks "Create Refactor Plan" and the agent is spawned successfully
- When the function returns the plan path
- Then the modal shows the plan file path as a clickable link
- And clicking the link dispatches an `open-file` event

AC-5: Error handling
- Given an error occurs during agent spawning (e.g., empty repoPath)
- When `SpawnRefactorPlan` is called
- Then the error is returned to the frontend
- And the modal shows the error message inline

## BDD Test Scenarios

### Scenario 1: Plan creation happy path

```gherkin
Feature: Refactor plan creation

  Scenario: Spawn agent and get plan path
    Given repoPath is "/tmp/test-repo" with git initialized
    And advice text is "Focus on SOLID principles..."
    When SpawnRefactorPlan is called with repoPath and advice text
    Then a session named starting with "refactor-" is spawned
    And the returned path matches "{repoPath}/.claude/plans/refactor-*.md"
    And the directory "{repoPath}/.claude/plans/" exists

  Scenario: Plans directory auto-created
    Given "{repoPath}/.claude/plans/" does not exist
    When SpawnRefactorPlan is called
    Then the directory is created with 0755 permissions
    And the agent spawns successfully
```

### Scenario 2: Input validation

```gherkin
Feature: Refactor plan input validation

  Scenario: Empty repo path rejected
    Given repoPath is empty string
    When SpawnRefactorPlan is called
    Then an error "repo path is required" is returned

  Scenario: Empty advice text rejected
    Given advice text is empty string
    When SpawnRefactorPlan is called
    Then an error "advice text is required" is returned
```

### Scenario 3: Frontend interaction

```gherkin
Feature: Refactor plan UI

  Scenario: Button disabled without advice
    Given the summarisation modal is open
    And no advice has been generated
    Then the "Create Refactor Plan" button is disabled

  Scenario: Button enabled with advice
    Given advice streaming has completed with text content
    Then the "Create Refactor Plan" button is enabled

  Scenario: Plan link clickable
    Given SpawnRefactorPlan returned path "/tmp/repo/.claude/plans/refactor-1712764800.md"
    When the path is displayed in the modal
    And the user clicks the path
    Then an "open-file" event is dispatched with that path
```

## Tasks / Subtasks

- [ ] Task 1: Implement SpawnRefactorPlan backend (AC: AC-1, AC-2, AC-5)
  - [ ] Add `SpawnRefactorPlan(repoPath, advice string) (string, error)` to `app_review.go`
  - [ ] Validate repoPath and advice non-empty
  - [ ] Create `.claude/plans/` directory with `os.MkdirAll`
  - [ ] Construct prompt with diff + advice + plan instructions
  - [ ] Spawn via `a.spawnSession` following `SpawnPRReview` pattern
  - [ ] Return the plan file path

- [ ] Task 2: Add frontend button and link (AC: AC-3, AC-4)
  - [ ] Add "Create Refactor Plan" button to SummarisationModal footer
  - [ ] Bind disabled state to `adviceText.length === 0`
  - [ ] On click: call `SpawnRefactorPlan`, show loading, then display plan path
  - [ ] Make plan path clickable, dispatch `open-file` on click

- [ ] Task 3: Write backend tests (AC: AC-1, AC-2, AC-5)
  - [ ] Test plan path generation format
  - [ ] Test directory creation
  - [ ] Test input validation (empty repoPath, empty advice)
  - [ ] Test prompt construction includes advice text

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on SpawnRefactorPlan
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
