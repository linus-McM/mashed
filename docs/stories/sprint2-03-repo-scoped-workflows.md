# Story 3: Add RepoPath to WorkflowDef and Storage Filtering

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Make workflows repo-scoped by adding a `RepoPath` field to `WorkflowDef` and updating the storage layer to filter workflows by repo. Currently workflows are globally stored at `~/.mashed/workflows/` with no repo association. After this change, each workflow knows which repo it belongs to, and the API can return only workflows relevant to a given repo.

## Developer Notes

### Architecture
- **Modify:** `internal/bmad/types.go` -- add `RepoPath string` field to `WorkflowDef`
- **Modify:** `internal/bmad/storage.go` -- add `ListWorkflowsByRepo(repoPath string)` method
- **Modify:** `app.go` -- update `ListBmadWorkflows` to accept optional repoPath filter, add `ListBmadWorkflowsByRepo`
- **Modify:** `internal/bmad/templates.go` -- no changes needed (templates have no RepoPath, they are generic)
- Storage location stays at `~/.mashed/workflows/` -- filtering is done in-memory after loading all workflows. This avoids needing directory-per-repo structure and keeps backward compatibility with existing saved workflows.

### Type Changes

```go
// In types.go, modify WorkflowDef:
type WorkflowDef struct {
    ID          string         `json:"id"`
    Name        string         `json:"name"`
    Description string         `json:"description"`
    RepoPath    string         `json:"repoPath,omitempty"` // NEW: repo this workflow belongs to
    Nodes       []WorkflowNode `json:"nodes"`
    Edges       []WorkflowEdge `json:"edges"`
    IsTemplate  bool           `json:"isTemplate"`
    TemplateID  string         `json:"templateId,omitempty"`
    CreatedAt   string         `json:"createdAt"`
    UpdatedAt   string         `json:"updatedAt"`
}
```

### Storage Changes

```go
// In storage.go, add new method:
func (s *Storage) ListWorkflowsByRepo(repoPath string) ([]WorkflowDef, error) {
    all, err := s.ListWorkflows()
    if err != nil {
        return nil, err
    }
    var filtered []WorkflowDef
    for _, wf := range all {
        if wf.RepoPath == repoPath {
            filtered = append(filtered, wf)
        }
    }
    return filtered, nil
}
```

### Technical Considerations
- Backward compatibility: existing workflows on disk have no `repoPath` field -- `json:"repoPath,omitempty"` ensures they deserialize with empty string. The `ListWorkflowsByRepo` filter naturally excludes them (they belong to no repo). The generic `ListWorkflows()` still returns all workflows.
- The `CreateFromTemplate` binding in `app.go` should accept a `repoPath` parameter so that workflows created from templates are immediately repo-scoped. Update signature: `CreateFromTemplate(templateID, repoPath string)`.
- `SaveBmadWorkflow` needs no changes -- it already saves whatever `WorkflowDef` is passed, including the new `RepoPath` field.
- Frontend will need to pass `repoPath` when saving workflows (handled in Story 4).

### Risks & Edge Cases
- Existing workflows with empty RepoPath will not appear in repo-filtered lists -- this is intentional (they are "orphaned" global workflows)
- Path comparison: normalize repoPath by removing trailing slashes before comparison
- Templates should never have RepoPath set (they are global) -- add a guard in `CreateFromTemplate`

### Reference Files
- `internal/bmad/types.go` lines 93-103 -- `WorkflowDef` struct to modify
- `internal/bmad/storage.go` lines 76-102 -- `ListWorkflows` to use as base for filtered version
- `internal/bmad/storage_test.go` -- test patterns for storage operations
- `app.go` lines 1547-1638 -- BMAD workflow CRUD to update

## Acceptance Criteria

AC-1: WorkflowDef has RepoPath field
- Given a WorkflowDef struct
- When RepoPath is set to "/Users/linus/Development/my-project"
- Then serializing to JSON includes `"repoPath": "/Users/linus/Development/my-project"`
- And deserializing back preserves the value

AC-2: ListWorkflowsByRepo filters correctly
- Given 3 saved workflows: 2 with RepoPath "/repo-a" and 1 with RepoPath "/repo-b"
- When `ListWorkflowsByRepo("/repo-a")` is called
- Then exactly 2 workflows are returned
- And both have RepoPath "/repo-a"

AC-3: Backward compatibility with existing workflows
- Given existing workflow JSON files on disk that lack the `repoPath` field
- When `ListWorkflows()` is called
- Then all workflows load successfully with empty RepoPath
- And `ListWorkflowsByRepo("/any-repo")` returns none of them

AC-4: CreateFromTemplate sets RepoPath
- Given a built-in template ID and a repoPath
- When `CreateFromTemplate(templateID, repoPath)` is called
- Then the created workflow has the specified RepoPath set
- And the workflow is saved to disk with the RepoPath

## BDD Test Scenarios

### Scenario 1: Repo-scoped workflow CRUD

```gherkin
Feature: Repo-scoped workflow storage

  Scenario: Save and list workflows filtered by repo
    Given a storage instance with temp directory
    And workflow "wf-1" saved with RepoPath "/repo-a"
    And workflow "wf-2" saved with RepoPath "/repo-a"
    And workflow "wf-3" saved with RepoPath "/repo-b"
    When ListWorkflowsByRepo("/repo-a") is called
    Then 2 workflows are returned
    And their IDs are "wf-1" and "wf-2"

  Scenario: Empty RepoPath workflows excluded from repo filter
    Given a storage instance with a workflow "wf-old" that has empty RepoPath
    When ListWorkflowsByRepo("/any-repo") is called
    Then 0 workflows are returned
    And ListWorkflows returns 1 workflow (the unscoped one)

  Scenario: RepoPath persists through save/load cycle
    Given a workflow with RepoPath "/my/project"
    When saved and then loaded by ID
    Then the loaded workflow has RepoPath "/my/project"

  Scenario: CreateFromTemplate with repoPath
    Given a built-in template "tpl-quick-sprint"
    When CreateFromTemplate is called with repoPath "/my/repo"
    Then the created workflow has RepoPath "/my/repo"
    And IsTemplate is false
```

## Tasks / Subtasks

- [ ] Task 1: Add RepoPath to WorkflowDef (AC: AC-1, AC-3)
  - [ ] Subtask 1a: Add `RepoPath string \`json:"repoPath,omitempty"\`` field to `WorkflowDef` in `internal/bmad/types.go`
  - [ ] Subtask 1b: Verify existing tests still pass with the new field (backward compat)

- [ ] Task 2: Add ListWorkflowsByRepo to Storage (AC: AC-2, AC-3)
  - [ ] Subtask 2a: Implement `ListWorkflowsByRepo(repoPath string) ([]WorkflowDef, error)` in `internal/bmad/storage.go`
  - [ ] Subtask 2b: Normalize repoPath (strip trailing slash) before comparison
  - [ ] Subtask 2c: Write tests for filtered listing with mixed RepoPath values

- [ ] Task 3: Update app.go bindings (AC: AC-1, AC-4)
  - [ ] Subtask 3a: Add `ListBmadWorkflowsByRepo(repoPath string)` method to App in `app.go`
  - [ ] Subtask 3b: Update `CreateFromTemplate` signature to accept `repoPath string` parameter
  - [ ] Subtask 3c: Set `wf.RepoPath = repoPath` in `CreateFromTemplate` before saving

- [ ] Task 4: Write storage tests (AC: AC-2, AC-3)
  - [ ] Subtask 4a: Test ListWorkflowsByRepo with multiple repos
  - [ ] Subtask 4b: Test backward compatibility loading workflows without RepoPath
  - [ ] Subtask 4c: Test round-trip serialization of RepoPath field

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ code coverage on new/modified files (90.8%)
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
