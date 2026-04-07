# Go Backend: BMAD Domain Types & Process Registry

**Story ID**: bmad-02
**Status**: ready
**Priority**: P0
**Depends On**: none

## Description

Create the `internal/bmad/` package with all core domain types (`ProcessDef`, `WorkflowDef`, `WorkflowNode`, `WorkflowEdge`, `BmadAgentConfig`, execution status types) and a hardcoded process registry containing all 25+ BMAD processes. This is the foundational data layer that every other backend story depends on. The registry provides query functions for listing, filtering by phase, and looking up processes by ID.

## Developer Notes

### Architecture

**New package:** `internal/bmad/`

**New files:**
- `internal/bmad/types.go` -- All domain types, status enums, phase/role constants
- `internal/bmad/registry.go` -- Hardcoded process catalog with query functions
- `internal/bmad/types_test.go` -- Type validation tests
- `internal/bmad/registry_test.go` -- Registry query tests

**Follow the pattern from** `internal/domain/types.go`:
- Use typed string constants for enums (like `AgentStatus`, `LogKind`, `EventType` in domain/types.go)
- Use sentinel errors with `errors.New()` for error cases (e.g., `ErrProcessNotFound`)
- All public types get JSON tags
- Keep types in a single `types.go` file

### Type Definitions

The types are specified exactly in the plan. Key types to implement:

```go
package bmad

// Enums: BmadPhase, BmadAgentRole, WorkflowNodeStatus, WorkflowExecStatus
// Structs: ProcessDef, WorkflowNode, Position, WorkflowEdge, WorkflowDef,
//          WorkflowExecution, BmadAgentConfig
```

See the plan (Phase 1.1) for the complete type definitions with all fields and JSON tags.

### Registry Functions

```go
// Sentinel errors
var (
    ErrProcessNotFound = errors.New("bmad: process not found")
    ErrModuleNotFound  = errors.New("bmad: module not found")
)

func AllProcesses() []ProcessDef
func ProcessesByPhase(phase BmadPhase) []ProcessDef
func ProcessByID(id string) (ProcessDef, bool)
func ProcessesByModule(moduleID string) []ProcessDef
```

The registry must contain all 25 processes from the plan (Phase 1.2 table):
- 5 analysis processes (brainstorming, product-brief, domain-research, market-research, technical-research)
- 4 planning processes (create-prd, edit-prd, validate-prd, create-ux-design)
- 4 solutioning processes (create-architecture, check-implementation-readiness, create-epics-and-stories, generate-project-context)
- 8 implementation processes (create-story, dev-story, quick-dev, sprint-planning, sprint-status, code-review, qa-generate-e2e-tests, retrospective)
- 4 support processes (editorial-review-prose, editorial-review-structure, advanced-elicitation, review-edge-case-hunter)

All processes in the initial registry have `ModuleID: "core"` and `Version: "1.0.0"`.

### Technical Considerations

- The registry is an in-memory `[]ProcessDef` initialized at package load via `init()` or a `sync.Once` pattern.
- All functions are pure and safe for concurrent use (the registry slice is never mutated after init).
- The `ProcessDef.Inputs` and `Outputs` fields use artifact names, not file paths. These are logical names like `"PRD.md"`, `"architecture.md"`, `"epics/"`.
- Each `ProcessDef.SkillName` matches the CLI invocation: `use bmad-{skillName}` becomes `use bmad-brainstorming`, etc.

### Risks & Edge Cases

- Typos in process IDs would cause silent failures downstream. Tests must verify all IDs are unique and follow the `bmad-` prefix convention.
- The `Inputs`/`Outputs` fields define the implicit DAG structure. If a process lists an input that no other process outputs, it is a "start node" (requires user-provided artifacts or is optional). Tests should verify that all non-empty inputs reference an output from at least one other process.

### Reference Files

- `internal/domain/types.go` -- Pattern for typed string constants, struct definitions, JSON tags
- `app.go` lines 29-41 -- App struct pattern (for seeing how packages are imported and used)

## Acceptance Criteria

- [ ] AC1: Given the `internal/bmad/` package, When `AllProcesses()` is called, Then it returns a slice containing at least 25 `ProcessDef` entries, each with a non-empty ID, Name, Phase, AgentRole, and SkillName.
- [ ] AC2: Given the registry, When `ProcessesByPhase(PhaseAnalysis)` is called, Then it returns exactly 5 processes all with `Phase == PhaseAnalysis`.
- [ ] AC3: Given the registry, When `ProcessByID("bmad-create-prd")` is called, Then it returns the matching ProcessDef with `ok == true`. When called with `"nonexistent"`, Then it returns `ok == false`.
- [ ] AC4: Given the registry, When `ProcessesByModule("core")` is called, Then it returns all 25 processes since all initial processes belong to the core module.
- [ ] AC5: Given all ProcessDef entries, When their IDs are collected, Then every ID is unique and starts with `"bmad-"`.
- [ ] AC6: Given all type definitions, When they are marshaled to JSON and unmarshaled back, Then all fields round-trip correctly with no data loss.

## BDD Test Scenarios

### Scenario 1: Registry returns all processes

```gherkin
Feature: BMAD Process Registry

  Scenario: AllProcesses returns the full catalog
    Given the bmad package is imported
    When AllProcesses() is called
    Then the returned slice has length >= 25
    And every entry has a non-empty ID, Name, Phase, AgentRole, and SkillName

  Scenario: ProcessesByPhase filters correctly
    Given the full process registry
    When ProcessesByPhase(PhaseAnalysis) is called
    Then exactly 5 processes are returned
    And all have Phase == PhaseAnalysis

  Scenario: ProcessesByPhase for all phases covers all processes
    Given the full process registry
    When ProcessesByPhase is called for each of the 5 phases
    Then the union of all results equals AllProcesses()

  Scenario: ProcessByID finds existing process
    Given the full process registry
    When ProcessByID("bmad-create-prd") is called
    Then the returned ProcessDef has Name == "Create PRD"
    And ok == true

  Scenario: ProcessByID returns false for missing process
    Given the full process registry
    When ProcessByID("nonexistent") is called
    Then ok == false

  Scenario: All process IDs are unique
    Given the full process registry
    When all IDs are collected into a set
    Then the set size equals the slice length

  Scenario: All process IDs follow naming convention
    Given the full process registry
    When each ID is checked
    Then every ID starts with "bmad-"
```

### Scenario 2: Type serialization round-trips

```gherkin
Feature: BMAD Type Serialization

  Scenario: WorkflowDef JSON round-trip
    Given a WorkflowDef with 3 nodes and 2 edges
    When it is marshaled to JSON and unmarshaled back
    Then all fields match the original values
    And node positions preserve float64 precision

  Scenario: ProcessDef JSON round-trip
    Given a ProcessDef with inputs ["PRD.md"] and outputs ["architecture.md"]
    When it is marshaled to JSON and unmarshaled back
    Then Inputs and Outputs slices match exactly

  Scenario: Empty optional fields serialize correctly
    Given a WorkflowDef with IsTemplate=false and empty TemplateID
    When it is marshaled to JSON
    Then the "templateId" field is omitted from the output
```

## Tasks / Subtasks

- [ ] Task 1: Create types.go with all domain types (AC: AC6)
  - [ ] Subtask 1a: Define BmadPhase, BmadAgentRole enums with typed string constants
  - [ ] Subtask 1b: Define ProcessDef struct with all fields and JSON tags
  - [ ] Subtask 1c: Define WorkflowNodeStatus, WorkflowExecStatus enums
  - [ ] Subtask 1d: Define WorkflowNode, Position, WorkflowEdge, WorkflowDef structs
  - [ ] Subtask 1e: Define WorkflowExecution and BmadAgentConfig structs
  - [ ] Subtask 1f: Define sentinel errors (ErrProcessNotFound, ErrModuleNotFound)
- [ ] Task 2: Create registry.go with process catalog (AC: AC1, AC2, AC3, AC4, AC5)
  - [ ] Subtask 2a: Define the internal registry slice with all 25 processes
  - [ ] Subtask 2b: Implement AllProcesses()
  - [ ] Subtask 2c: Implement ProcessesByPhase()
  - [ ] Subtask 2d: Implement ProcessByID()
  - [ ] Subtask 2e: Implement ProcessesByModule()
- [ ] Task 3: Write comprehensive tests (AC: AC1-AC6)
  - [ ] Subtask 3a: Test AllProcesses count, field completeness, ID uniqueness
  - [ ] Subtask 3b: Test ProcessesByPhase for each phase
  - [ ] Subtask 3c: Test ProcessByID for existing and missing IDs
  - [ ] Subtask 3d: Test JSON round-trip for all major types
  - [ ] Subtask 3e: Test that all inputs reference at least one other process's outputs

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/bmad/types.go` and `internal/bmad/registry.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] /simplify run on all new code
