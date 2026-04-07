# Wails Bindings: BMAD API Surface

**Story ID**: bmad-05
**Status**: ready
**Priority**: P0
**Depends On**: bmad-02, bmad-03, bmad-04

## Description

Wire the entire BMAD backend into the Wails application by adding ~15 new bound methods to the `App` struct in `app.go`. This creates the bridge between the Go backend and the Svelte frontend: workflow CRUD, process registry queries, template listing, execution lifecycle (start/pause/resume/stop), and agent management. After this story, the frontend can call all BMAD operations through the auto-generated `wailsjs/go/main/App.js` bindings.

## Developer Notes

### Architecture

**Modified file:** `app.go`

Add two new fields to the `App` struct:
```go
type App struct {
    // ... existing fields ...
    bmadStorage  *bmad.Storage
    bmadExecutor *bmad.Executor
}
```

Initialize in `startup()` after line 127 (after `cfg` is loaded):
```go
func (a *App) startup(ctx context.Context) {
    // ... existing init ...

    // Initialize BMAD subsystem
    home, _ := os.UserHomeDir()
    bmadDir := filepath.Join(home, ".mashed")
    storage, err := bmad.NewStorage(bmadDir)
    if err != nil {
        log.Printf("bmad storage init failed: %v", err)
    } else {
        a.bmadStorage = storage
        a.bmadExecutor = bmad.NewExecutor(storage, func(event string, data interface{}) {
            runtime.EventsEmit(a.ctx, event, data)
        })
    }
}
```

**New bound methods (15 total):**

Workflow CRUD:
```go
func (a *App) ListBmadWorkflows() ([]bmad.WorkflowDef, error)
func (a *App) GetBmadWorkflow(id string) (bmad.WorkflowDef, error)
func (a *App) SaveBmadWorkflow(wf bmad.WorkflowDef) error
func (a *App) DeleteBmadWorkflow(id string) error
```

Process Registry:
```go
func (a *App) GetBmadProcesses() []bmad.ProcessDef
func (a *App) GetBmadProcessesByPhase(phase string) []bmad.ProcessDef
```

Templates:
```go
func (a *App) ListBmadTemplates() []bmad.WorkflowDef
func (a *App) CreateFromTemplate(templateID string) (bmad.WorkflowDef, error)
```

Execution:
```go
func (a *App) StartBmadWorkflow(workflowID, repoPath, model string) (string, error)
func (a *App) PauseBmadWorkflow(execID string) error
func (a *App) ResumeBmadWorkflow(execID string) error
func (a *App) StopBmadWorkflow(execID string) error
func (a *App) GetBmadExecution(execID string) (*bmad.WorkflowExecution, error)
```

Agent Management:
```go
func (a *App) ListBmadAgents() ([]bmad.BmadAgentConfig, error)
func (a *App) SaveBmadAgent(agent bmad.BmadAgentConfig) error
func (a *App) DeleteBmadAgent(id string) error
```

### CreateFromTemplate Logic

`CreateFromTemplate` deep-copies a template into a user workflow:
```go
func (a *App) CreateFromTemplate(templateID string) (bmad.WorkflowDef, error) {
    templates := bmad.BuiltinTemplates()
    // Find template by ID
    // Deep-copy: new ID (uuid or timestamp-based), new name ("Copy of {name}"),
    //   IsTemplate=false, TemplateID=templateID, timestamps
    // Save via storage
    // Return the new workflow
}
```

### Error Handling Pattern

Follow existing pattern in `app.go` -- methods return `(value, error)` and the Wails framework automatically converts Go errors to JS rejected promises. Example from existing code (line 565):
```go
func (a *App) SetDevDir(dir string) error {
    // ...
    return fmt.Errorf("init scanning: %w", err)
}
```

For nil storage guard (if BMAD init failed):
```go
func (a *App) ListBmadWorkflows() ([]bmad.WorkflowDef, error) {
    if a.bmadStorage == nil {
        return nil, fmt.Errorf("bmad storage not initialized")
    }
    return a.bmadStorage.ListWorkflows()
}
```

### Technical Considerations

- **Import**: Add `"mashed/internal/bmad"` to the import block at the top of `app.go`.
- **Wails binding generation**: After adding the methods, `wails dev` or `wails generate module` will auto-generate `frontend/wailsjs/go/main/App.js` and `App.d.ts`. These files should NOT be manually edited.
- **Phase casting**: `GetBmadProcessesByPhase` receives a `string` from JS. Cast to `bmad.BmadPhase` before calling the registry function.
- **Event naming**: The executor emits events with names like `"bmad:node:status"` and `"bmad:execution:status"`. The frontend will listen for these via `EventsOn`.
- **Nil guards**: Every method must check `a.bmadStorage != nil` and `a.bmadExecutor != nil` before use. Return a descriptive error if nil.

### Risks & Edge Cases

- **Startup failure**: If `bmad.NewStorage` fails (permissions, disk), the storage and executor will be nil. All methods must handle this gracefully.
- **Type serialization**: Wails serializes Go structs to JSON for the JS bridge. All BMAD types have JSON tags (from bmad-02), so this should work automatically. However, verify that `map[string]string` in `WorkflowNode.Config` serializes correctly.
- **Execution ID propagation**: `StartBmadWorkflow` returns the execution ID as a string. The frontend needs this to call `GetBmadExecution` and listen for scoped events.

### Reference Files

- `app.go` lines 29-41 -- App struct definition (add new fields here)
- `app.go` lines 106-128 -- `startup()` method (add BMAD init here)
- `app.go` lines 553-565 -- `PickDirectory`, `SetDevDir` (pattern for methods that return errors)
- `app.go` lines 692-706 -- `GetNotifications`, `GetTerminalPort` (pattern for simple getter methods)
- `app.go` lines 755-814 -- `SpawnAgent`, `SpawnAgentWithCommand` (pattern for tmux methods)

## Acceptance Criteria

- [ ] AC1: Given the `App` struct, When `startup()` completes, Then `a.bmadStorage` and `a.bmadExecutor` are non-nil (assuming `~/.mashed/` is writable).
- [ ] AC2: Given the running app, When `GetBmadProcesses()` is called from JavaScript, Then it returns the full process registry (25+ entries) as a JSON array.
- [ ] AC3: Given the running app, When `ListBmadTemplates()` is called from JavaScript, Then it returns the 6 built-in templates.
- [ ] AC4: Given the running app, When `CreateFromTemplate("quick-sprint")` is called, Then a new workflow is created with `IsTemplate: false`, saved to storage, and returned.
- [ ] AC5: Given a saved workflow, When `SaveBmadWorkflow` then `GetBmadWorkflow` are called sequentially from JavaScript, Then the returned workflow matches what was saved.
- [ ] AC6: Given a valid workflow saved to storage, When `StartBmadWorkflow(id, repoPath, "claude-sonnet-4-20250514")` is called, Then it returns an execution ID string and the execution begins.
- [ ] AC7: Given a nil `bmadStorage` (init failure), When any BMAD method is called, Then it returns a descriptive error rather than panicking.

## BDD Test Scenarios

### Scenario 1: Binding initialization

```gherkin
Feature: BMAD Wails Bindings

  Scenario: BMAD subsystem initializes on startup
    Given the App startup sequence runs
    And ~/.mashed/ is writable
    When startup completes
    Then bmadStorage is non-nil
    And bmadExecutor is non-nil

  Scenario: Graceful handling of init failure
    Given the App startup sequence runs
    And ~/.mashed/ is not writable
    When startup completes
    Then bmadStorage is nil
    And ListBmadWorkflows returns an error containing "not initialized"
```

### Scenario 2: Process and template queries

```gherkin
Feature: BMAD Registry Bindings

  Scenario: GetBmadProcesses returns full catalog
    Given the app is running with BMAD initialized
    When GetBmadProcesses is called
    Then 25+ ProcessDef objects are returned

  Scenario: GetBmadProcessesByPhase filters correctly
    Given the app is running
    When GetBmadProcessesByPhase("analysis") is called
    Then 5 ProcessDef objects are returned
    And all have phase "analysis"

  Scenario: ListBmadTemplates returns built-in templates
    Given the app is running
    When ListBmadTemplates is called
    Then 6 WorkflowDef objects are returned
    And all have isTemplate == true
```

### Scenario 3: Workflow CRUD via bindings

```gherkin
Feature: BMAD Workflow CRUD Bindings

  Scenario: Create from template and retrieve
    Given the app is running
    When CreateFromTemplate("quick-sprint") is called
    Then a WorkflowDef is returned with isTemplate == false
    And templateId == "quick-sprint"
    And the workflow is persisted to storage

  Scenario: Save and load workflow
    Given the app is running
    And a workflow is created via CreateFromTemplate
    When SaveBmadWorkflow is called with modified nodes
    And GetBmadWorkflow is called with the same ID
    Then the returned workflow reflects the modifications
```

## Tasks / Subtasks

- [ ] Task 1: Add BMAD fields to App and initialize in startup (AC: AC1)
  - [ ] Subtask 1a: Add `bmadStorage` and `bmadExecutor` fields to App struct
  - [ ] Subtask 1b: Add `"mashed/internal/bmad"` import
  - [ ] Subtask 1c: Initialize storage and executor in `startup()` with error logging
- [ ] Task 2: Implement registry and template bindings (AC: AC2, AC3)
  - [ ] Subtask 2a: Implement `GetBmadProcesses()` (delegates to `bmad.AllProcesses()`)
  - [ ] Subtask 2b: Implement `GetBmadProcessesByPhase()` with string-to-BmadPhase cast
  - [ ] Subtask 2c: Implement `ListBmadTemplates()` (delegates to `bmad.BuiltinTemplates()`)
- [ ] Task 3: Implement workflow CRUD bindings (AC: AC4, AC5)
  - [ ] Subtask 3a: Implement `ListBmadWorkflows`, `GetBmadWorkflow`, `SaveBmadWorkflow`, `DeleteBmadWorkflow`
  - [ ] Subtask 3b: Implement `CreateFromTemplate` with deep-copy logic
- [ ] Task 4: Implement execution bindings (AC: AC6)
  - [ ] Subtask 4a: Implement `StartBmadWorkflow`, `PauseBmadWorkflow`, `ResumeBmadWorkflow`, `StopBmadWorkflow`
  - [ ] Subtask 4b: Implement `GetBmadExecution`
- [ ] Task 5: Implement agent bindings and nil guards (AC: AC7)
  - [ ] Subtask 5a: Implement `ListBmadAgents`, `SaveBmadAgent`, `DeleteBmadAgent`
  - [ ] Subtask 5b: Add nil guards to all 15 methods

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] `go build ./...` passes (verifies all methods compile and Wails can bind them)
- [ ] `go vet ./...` passes
- [ ] All nil-guard paths have test coverage
- [ ] Wails bindings auto-generate successfully (`wails generate module`)
- [ ] /simplify run on all modified code
