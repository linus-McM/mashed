# Go Backend: Templates & File Storage

**Story ID**: bmad-03
**Status**: ready
**Priority**: P0
**Depends On**: bmad-02

## Description

Implement the built-in workflow template catalog and file-based persistence layer for saving/loading user workflows and custom agent configs. Templates provide one-click starting points for common BMAD sequences (Full Lifecycle, Quick Sprint, Architecture Review, etc.). Storage handles CRUD to `~/.mashed/workflows/` and `~/.mashed/bmad-agents/`, giving users persistent workflow definitions that survive app restarts.

## Developer Notes

### Architecture

**New files:**
- `internal/bmad/templates.go` -- Built-in template definitions, `BuiltinTemplates() []WorkflowDef`
- `internal/bmad/storage.go` -- File-based persistence for workflows and agents
- `internal/bmad/templates_test.go` -- Template validation tests
- `internal/bmad/storage_test.go` -- Storage CRUD tests (use `t.TempDir()`)

**Storage directory structure:**
```
~/.mashed/
  workflows/       # WorkflowDef JSON files: {id}.json
  bmad-agents/     # BmadAgentConfig JSON files: {id}.json
```

### Template Definitions

Each template is a `WorkflowDef` with `IsTemplate: true`. Templates have pre-placed nodes with positions and edges. Six templates from the plan:

1. **Full Product Lifecycle** -- 13 core processes in sequence, analysis through implementation
2. **Quick Sprint** -- `bmad-create-story` -> `bmad-dev-story` -> `bmad-code-review` -> `bmad-qa-generate-e2e-tests`
3. **Architecture Review** -- `bmad-create-architecture` -> `bmad-check-implementation-readiness` -> `bmad-review-edge-case-hunter`
4. **Story Development** -- `bmad-create-story` -> `bmad-dev-story` -> `bmad-code-review`
5. **PRD Pipeline** -- `bmad-brainstorming` -> `bmad-product-brief` -> `bmad-create-prd` -> `bmad-validate-prd`
6. **QA & Polish** -- `bmad-code-review` -> `bmad-qa-generate-e2e-tests` -> `bmad-editorial-review-prose`

Each template node must have:
- A unique node ID (e.g., `"tpl-qs-1"`, `"tpl-qs-2"`)
- A `ProcessID` referencing a valid entry from the registry
- A `Position` with x,y coordinates laid out left-to-right with ~250px horizontal spacing
- `Status: NodePending`

Each template edge connects sequential nodes: `Source: "tpl-qs-1", Target: "tpl-qs-2"`.

### Storage Implementation

```go
type Storage struct {
    workflowDir string
    agentDir    string
    mu          sync.RWMutex
}

func NewStorage(baseDir string) (*Storage, error)  // creates dirs if missing
func (s *Storage) SaveWorkflow(wf WorkflowDef) error
func (s *Storage) LoadWorkflow(id string) (WorkflowDef, error)
func (s *Storage) ListWorkflows() ([]WorkflowDef, error)
func (s *Storage) DeleteWorkflow(id string) error
func (s *Storage) SaveAgent(agent BmadAgentConfig) error
func (s *Storage) ListAgents() ([]BmadAgentConfig, error)
func (s *Storage) DeleteAgent(id string) error
```

**Error handling** -- use custom sentinel errors:
```go
var (
    ErrWorkflowNotFound = errors.New("bmad: workflow not found")
    ErrAgentNotFound    = errors.New("bmad: agent not found")
    ErrInvalidID        = errors.New("bmad: invalid ID")
)
```

**Concurrency**: `sync.RWMutex` on all file operations. Reads take `RLock`, writes take full `Lock`.

**ID validation**: IDs must be non-empty and contain only `[a-zA-Z0-9_-]` characters (no path traversal).

### Technical Considerations

- Use `os.MkdirAll` to create storage directories on first use.
- Use `json.MarshalIndent` for human-readable JSON files (consistent with `saveConfig` in `app.go` line 91).
- Use atomic writes: write to temp file, then `os.Rename` to final path (prevents corruption on crash).
- `ListWorkflows` reads the directory, filters for `.json` files, unmarshals each. Skip malformed files with a log warning rather than failing the entire list.
- `NewStorage` should accept a base directory (defaulting to `~/.mashed/`) so tests can use `t.TempDir()`.

### Risks & Edge Cases

- **Path traversal**: Validate workflow/agent IDs before constructing file paths. Reject IDs containing `/`, `..`, or other path separators.
- **Concurrent writes**: Two saves to the same ID could race. The mutex prevents this.
- **Disk full**: `SaveWorkflow` should wrap OS errors with context: `fmt.Errorf("saving workflow %s: %w", id, err)`.
- **Template validation**: Every template node's `ProcessID` must exist in the registry. A test should verify this.
- **Malformed JSON on disk**: `ListWorkflows` must not panic on a corrupted file. Log and skip.

### Reference Files

- `app.go` lines 62-95 -- `configPath()`, `loadConfig()`, `saveConfig()` pattern for file persistence
- `internal/domain/types.go` -- JSON tag conventions
- `internal/bmad/types.go` (from bmad-02) -- types this story persists

## Acceptance Criteria

- [ ] AC1: Given `BuiltinTemplates()` is called, When the result is inspected, Then it returns exactly 6 templates, each with `IsTemplate: true`, valid node ProcessIDs from the registry, and edges forming a connected DAG.
- [ ] AC2: Given a new `Storage` initialized with a temp directory, When `SaveWorkflow(wf)` is called followed by `LoadWorkflow(wf.ID)`, Then the loaded workflow equals the saved workflow in all fields.
- [ ] AC3: Given a `Storage` with 3 saved workflows, When `ListWorkflows()` is called, Then it returns all 3 workflows.
- [ ] AC4: Given a `Storage` with a saved workflow, When `DeleteWorkflow(id)` is called, Then `LoadWorkflow(id)` returns `ErrWorkflowNotFound`.
- [ ] AC5: Given a `Storage`, When `SaveAgent(agent)` is called followed by `ListAgents()`, Then the agent appears in the list.
- [ ] AC6: Given a `Storage`, When `SaveWorkflow` is called with an ID containing `"../"`, Then it returns `ErrInvalidID`.
- [ ] AC7: Given a malformed JSON file in the workflows directory, When `ListWorkflows()` is called, Then it returns all valid workflows and skips the malformed file without panicking.

## BDD Test Scenarios

### Scenario 1: Template validation

```gherkin
Feature: BMAD Workflow Templates

  Scenario: All built-in templates are valid
    Given BuiltinTemplates() is called
    Then exactly 6 templates are returned
    And each has IsTemplate == true
    And each has at least 2 nodes
    And each has at least 1 edge

  Scenario: Template node ProcessIDs exist in registry
    Given BuiltinTemplates() is called
    When each node's ProcessID is looked up via ProcessByID
    Then every lookup returns ok == true

  Scenario: Quick Sprint template has correct structure
    Given BuiltinTemplates() is called
    When the "Quick Sprint" template is found by name
    Then it has 4 nodes
    And 3 edges connecting them sequentially
    And node ProcessIDs are bmad-create-story, bmad-dev-story, bmad-code-review, bmad-qa-generate-e2e-tests
```

### Scenario 2: Workflow storage CRUD

```gherkin
Feature: BMAD Workflow Storage

  Scenario: Save and load round-trip
    Given a Storage with a temp directory
    And a WorkflowDef with ID "test-wf-1" and 2 nodes
    When SaveWorkflow is called
    And LoadWorkflow("test-wf-1") is called
    Then the loaded workflow matches the saved workflow

  Scenario: List returns all saved workflows
    Given a Storage with 3 saved workflows
    When ListWorkflows is called
    Then 3 workflows are returned

  Scenario: Delete removes workflow
    Given a Storage with a saved workflow "test-wf-1"
    When DeleteWorkflow("test-wf-1") is called
    Then LoadWorkflow("test-wf-1") returns ErrWorkflowNotFound

  Scenario: Path traversal is rejected
    Given a Storage
    When SaveWorkflow is called with ID "../../../etc/passwd"
    Then ErrInvalidID is returned

  Scenario: Malformed JSON is skipped during list
    Given a Storage directory containing a valid workflow and a file with invalid JSON
    When ListWorkflows is called
    Then 1 workflow is returned
    And no panic occurs
```

### Scenario 3: Agent storage

```gherkin
Feature: BMAD Agent Storage

  Scenario: Save and list agents
    Given a Storage with a temp directory
    And a BmadAgentConfig with ID "custom-dev-1"
    When SaveAgent is called
    And ListAgents is called
    Then the agent appears in the list

  Scenario: Delete agent
    Given a Storage with a saved agent "custom-dev-1"
    When DeleteAgent("custom-dev-1") is called
    And ListAgents is called
    Then the agent does not appear in the list
```

## Tasks / Subtasks

- [ ] Task 1: Implement templates.go (AC: AC1)
  - [ ] Subtask 1a: Create the 6 template definitions with nodes and edges
  - [ ] Subtask 1b: Assign positions to nodes (left-to-right, 250px spacing)
  - [ ] Subtask 1c: Implement `BuiltinTemplates()` returning all templates
- [ ] Task 2: Implement storage.go (AC: AC2, AC3, AC4, AC5, AC6, AC7)
  - [ ] Subtask 2a: Implement `NewStorage()` with directory creation
  - [ ] Subtask 2b: Implement ID validation function (reject path traversal)
  - [ ] Subtask 2c: Implement `SaveWorkflow` with atomic write (temp file + rename)
  - [ ] Subtask 2d: Implement `LoadWorkflow` with error wrapping
  - [ ] Subtask 2e: Implement `ListWorkflows` with malformed-file resilience
  - [ ] Subtask 2f: Implement `DeleteWorkflow`
  - [ ] Subtask 2g: Implement `SaveAgent`, `ListAgents`, `DeleteAgent`
- [ ] Task 3: Write tests (AC: AC1-AC7)
  - [ ] Subtask 3a: Test all template structures and registry cross-references
  - [ ] Subtask 3b: Test workflow CRUD round-trip using `t.TempDir()`
  - [ ] Subtask 3c: Test path traversal rejection
  - [ ] Subtask 3d: Test malformed JSON resilience
  - [ ] Subtask 3e: Test agent CRUD

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `templates.go` and `storage.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes (verifies mutex correctness)
- [ ] /simplify run on all new code
