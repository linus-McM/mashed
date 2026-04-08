# Story 1: Sprint Status YAML Parser

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Create a Go package to parse `sprint-status.yaml` files from any repo path. This is the foundational data layer that all subsequent stories depend on -- it provides the `SprintStatus`, `SprintEpic`, and `SprintStory` types and a parser that reads the YAML format used by the BMAD method. Without this, neither the frontend sprint panel nor the workflow-to-story linking can function.

## Developer Notes

### Architecture
- **New file:** `internal/bmad/sprint.go` -- types and parser
- **New file:** `internal/bmad/sprint_test.go` -- table-driven tests
- The parser lives in the existing `bmad` package alongside `types.go`, `storage.go`, etc.
- YAML path convention: `{repoPath}/_bmad-output/implementation-artifacts/sprint-status.yaml`
- `gopkg.in/yaml.v3` is already in `go.mod` (indirect dep) -- import it directly

### Key Types to Create

```go
// StoryStatus represents the lifecycle state of a sprint story.
type StoryStatus string

const (
    StoryBacklog     StoryStatus = "backlog"
    StoryReadyForDev StoryStatus = "ready-for-dev"
    StoryInProgress  StoryStatus = "in-progress"
    StoryReview      StoryStatus = "review"
    StoryDone        StoryStatus = "done"
)

// EpicStatus represents the lifecycle state of an epic.
type EpicStatus string

const (
    EpicBacklog    EpicStatus = "backlog"
    EpicInProgress EpicStatus = "in-progress"
    EpicDone       EpicStatus = "done"
)

// SprintStory is a single user story within an epic.
type SprintStory struct {
    ID       string      `json:"id"`       // e.g. "1-2-account-management"
    EpicID   string      `json:"epicId"`   // e.g. "epic-1"
    Status   StoryStatus `json:"status"`
    Sequence int         `json:"sequence"` // ordering within epic
}

// SprintEpic groups stories under a named epic.
type SprintEpic struct {
    ID      string     `json:"id"`      // e.g. "epic-1"
    Status  EpicStatus `json:"status"`
    Stories []SprintStory `json:"stories"`
}

// SprintStatus is the parsed representation of sprint-status.yaml.
type SprintStatus struct {
    Generated     string       `json:"generated"`
    LastUpdated   string       `json:"lastUpdated"`
    Project       string       `json:"project"`
    ProjectKey    string       `json:"projectKey"`
    TrackingSystem string     `json:"trackingSystem"`
    StoryLocation string       `json:"storyLocation"`
    Epics         []SprintEpic `json:"epics"`
}
```

### Technical Considerations
- The YAML `development_status` section is a flat ordered map, not nested. Epic entries like `epic-1: backlog` are interspersed with story entries like `1-1-user-authentication: done`. The parser must distinguish epics from stories by key pattern: keys starting with `epic-` are epics; keys matching `\d+-\d+-.*` are stories belonging to the most recently seen epic. Keys ending in `-retrospective` are optional metadata entries (treat as stories with the parent epic).
- Use `yaml.v3` Node API for ordered map parsing since Go maps lose key order. The `development_status` field must be parsed as a `yaml.Node` of kind `MappingNode`, then iterated pairwise over `Content` to preserve insertion order.
- Error handling: return sentinel errors `ErrSprintFileNotFound` and `ErrSprintFileMalformed` (add to `types.go`). Wrap with `fmt.Errorf("parsing sprint status: %w", err)`.
- The `ParseSprintStatus(repoPath string) (SprintStatus, error)` function derives the full path internally.
- Add a `ValidateStoryStatus(s string) (StoryStatus, bool)` helper for safe casting from raw YAML strings.

### Risks & Edge Cases
- YAML file might not exist yet (repo has not run BMAD init) -- return `ErrSprintFileNotFound`
- Keys with unexpected formats (e.g. comments, extra fields) -- skip gracefully, log warning
- Empty `development_status` section -- return valid `SprintStatus` with empty `Epics` slice
- Multiple epics with stories interleaved -- parser must track "current epic" as it iterates

### Reference Files
- `internal/bmad/types.go` -- existing type patterns, sentinel error style
- `internal/bmad/storage.go` -- file I/O patterns (os.ReadFile, error wrapping)
- `internal/bmad/storage_test.go` -- test patterns (testify, t.TempDir(), table-driven)
- `go.mod` line 43: `gopkg.in/yaml.v3 v3.0.1` already available

## Acceptance Criteria

AC-1: Parse valid sprint-status.yaml
- Given a repo path containing a valid `_bmad-output/implementation-artifacts/sprint-status.yaml`
- When `ParseSprintStatus(repoPath)` is called
- Then it returns a `SprintStatus` with correct project name, epics, and stories
- And stories are correctly associated with their parent epic

AC-2: Handle missing YAML file
- Given a repo path where `_bmad-output/implementation-artifacts/sprint-status.yaml` does not exist
- When `ParseSprintStatus(repoPath)` is called
- Then it returns `ErrSprintFileNotFound`

AC-3: Preserve story ordering within epics
- Given a sprint-status.yaml with multiple stories under an epic
- When parsed
- Then `SprintStory.Sequence` values reflect the original YAML ordering (0-indexed)
- And stories appear in the `SprintEpic.Stories` slice in YAML document order

AC-4: Parse all valid status values
- Given stories with statuses `backlog`, `ready-for-dev`, `in-progress`, `review`, `done`
- When parsed
- Then each `SprintStory.Status` maps to the correct `StoryStatus` constant

AC-5: Handle empty or minimal YAML
- Given a sprint-status.yaml with header fields but an empty `development_status` section
- When parsed
- Then it returns a valid `SprintStatus` with an empty `Epics` slice and no error

## BDD Test Scenarios

### Scenario 1: Full parse of multi-epic YAML

```gherkin
Feature: Sprint status YAML parsing

  Scenario: Parse a complete sprint-status.yaml with two epics
    Given a temporary repo directory with sprint-status.yaml containing:
      """
      generated: 05-06-2025 21:30
      last_updated: 05-06-2025 21:30
      project: Test Project
      project_key: TEST
      tracking_system: file-system
      story_location: "_bmad-output/stories"

      development_status:
        epic-1: in-progress
        1-1-user-auth: done
        1-2-account-mgmt: ready-for-dev
        1-3-data-model: backlog
        epic-1-retrospective: optional

        epic-2: backlog
        2-1-personality: backlog
        2-2-chat-ui: backlog
      """
    When ParseSprintStatus is called with the repo path
    Then the result contains 2 epics
    And epic-1 has status "in-progress" and 4 stories
    And epic-2 has status "backlog" and 2 stories
    And story "1-1-user-auth" has status "done" and sequence 0
    And story "1-2-account-mgmt" has status "ready-for-dev" and sequence 1

  Scenario: Missing sprint-status.yaml file
    Given a temporary repo directory with no _bmad-output directory
    When ParseSprintStatus is called with the repo path
    Then the error wraps ErrSprintFileNotFound

  Scenario: Empty development_status section
    Given a temporary repo directory with sprint-status.yaml containing only header fields
    When ParseSprintStatus is called
    Then SprintStatus.Epics is an empty slice
    And SprintStatus.Project equals the header project name

  Scenario: Malformed YAML content
    Given a temporary repo directory with sprint-status.yaml containing invalid YAML
    When ParseSprintStatus is called
    Then the error wraps ErrSprintFileMalformed

  Scenario: Story status validation
    Given raw status strings "backlog", "ready-for-dev", "in-progress", "review", "done", "unknown"
    When ValidateStoryStatus is called for each
    Then the first 5 return the correct StoryStatus and true
    And "unknown" returns empty StoryStatus and false
```

## Tasks / Subtasks

- [ ] Task 1: Define sprint types and sentinel errors (AC: AC-1, AC-4)
  - [ ] Subtask 1a: Add `StoryStatus`, `EpicStatus` enums with constants to `internal/bmad/sprint.go`
  - [ ] Subtask 1b: Add `SprintStory`, `SprintEpic`, `SprintStatus` structs to `internal/bmad/sprint.go`
  - [ ] Subtask 1c: Add `ErrSprintFileNotFound` and `ErrSprintFileMalformed` sentinels to `internal/bmad/types.go`
  - [ ] Subtask 1d: Add `ValidateStoryStatus` and `ValidateEpicStatus` helper functions

- [ ] Task 2: Implement YAML parser with ordered map handling (AC: AC-1, AC-3, AC-5)
  - [ ] Subtask 2a: Implement `ParseSprintStatus(repoPath string) (SprintStatus, error)` with path derivation
  - [ ] Subtask 2b: Implement `parseDevelopmentStatus(node *yaml.Node) ([]SprintEpic, error)` using yaml.Node pairwise iteration
  - [ ] Subtask 2c: Handle epic detection (key starts with `epic-`) vs story detection (key matches `\d+-\d+-.*`)
  - [ ] Subtask 2d: Assign `Sequence` values to stories based on their order within each epic

- [ ] Task 3: Write comprehensive tests (AC: AC-1, AC-2, AC-3, AC-4, AC-5)
  - [ ] Subtask 3a: Write test helpers to create temp repo dirs with sprint-status.yaml content
  - [ ] Subtask 3b: Write table-driven test for `ParseSprintStatus` covering all scenarios
  - [ ] Subtask 3c: Write table-driven test for `ValidateStoryStatus` and `ValidateEpicStatus`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
