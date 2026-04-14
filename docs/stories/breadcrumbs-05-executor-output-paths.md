# Story breadcrumbs-05: Executor writes resolved OutputPaths on node complete

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** breadcrumbs-01, breadcrumbs-04
**Status:** done

## Description

Phase 2 Task 2.1 (plan lines 68-85). When a BMAD process node transitions to `complete`, the executor resolves each declared output artifact via `ResolveArtifactPath`, verifies the file exists on disk, and writes the absolute path into `node.data.config.OutputPaths[artifactName]`. The existing `bmad:node:artifacts` event then carries the updated map to the frontend. This is the authoritative side of auto-propagation per architectural decision #1 (plan line 190).

## Developer Notes

### Architecture

Target file: `internal/bmad/executor.go` (corpus line 25304). Relevant types in `internal/bmad/types.go` (corpus line 20748).

Per architectural decision #1 (plan line 190):

```
Artifact path storage shape — resolved.
OutputPaths map[string]string (artifact name → absolute path)
on WorkflowNode.Data.Config. json:",omitempty" so old saved
workflows still parse. Risks to mitigate:
(a) concurrent writes during parallel execution → wrap map
    mutation in the executor's existing node mutex;
(b) stale paths if user moves files between runs → clear on
    next `start`;
(c) drift if artifacts.go map changes → old workflows ignore
    unknown keys, re-populate on next run.
```

Current `WorkflowNode.Config` is `map[string]string` (corpus line 20816). Options:

1. **Introduce a nested typed struct** `WorkflowNodeData` containing `Config map[string]string`, `OutputPaths map[string]string`, `InputPaths map[string]string`. Requires migration.
2. **Stay on flat Config + add `OutputPaths map[string]string` as a sibling field on WorkflowNode**. Keep `Config map[string]string` as-is. Use `json:",omitempty"`.

Go with option 2 — minimal blast radius, matches plan phrasing ("on WorkflowNode.Data.Config" is a pattern-level description, not a literal shape requirement). Add fields:

```go
type WorkflowNode struct {
    // ... existing ...
    OutputPaths map[string]string `json:"outputPaths,omitempty"`
    InputPaths  map[string]string `json:"inputPaths,omitempty"`
}
```

Pseudo-code from plan lines 72-83:

```go
for _, artifactName := range node.Outputs {
    resolved := ResolveArtifactPath(artifactName, repoPath)
    if resolved == "" { continue }            // unmapped
    if !fileExists(resolved) { continue }     // not written
    node.OutputPaths[artifactName] = resolved
}
```

Emit the existing `bmad:node:artifacts` event (plan line 85 — do NOT invent a new event).

### Technical Considerations

- **Concurrency**: executor runs ready nodes in parallel (corpus line 25493). Wrap the map mutation in `state.mu` (the existing per-execution `sync.Mutex`). Do NOT introduce a new mutex.
- **Start-of-run clear**: in `StartWorkflow`, zero `OutputPaths` on every node before the DAG walks begin (plan line 191 rule — default deterministic re-run clears stale breadcrumbs).
- **Unmapped artifacts**: `ResolveArtifactPath` returns `""` for `"code"`, `"tests"`, `"any-doc"`, `"file-path"`. Skip silently (plan line 78).
- **Missing files on disk**: skip silently too (plan line 81). The existing `VerifyArtifacts` already classifies found/missing and feeds `NodeArtifactEvent.Found` / `.Missing` (corpus line 20907).
- **Error handling**: follow the project-memory pattern (sentinels + wrapping). If writes fail (e.g. map init), wrap `fmt.Errorf("bmad: write output path for %s: %w", name, err)` but do NOT fail the node — log and continue.

### Wails binding requirements

No new bindings. `bmad:node:artifacts` already flows through `app.go` / `app_bmad.go`. Payload struct `NodeArtifactEvent` (types.go line 20907) may need a new field `Paths map[string]string` so the frontend (breadcrumbs-06) can read resolved paths without a round-trip.

### Risks & Edge Cases

- Concurrent writes (plan decision #1a): mitigated by existing `state.mu`.
- Stale paths after file move (plan decision #1b): cleared on next `start`.
- Drift between `artifacts.go` map and saved workflows (plan decision #1c): unknown keys ignored; re-populate on next run.
- `fileExists` helper may need to be added (`os.Stat` + `!os.IsNotExist`). Place it as an unexported helper in `executor.go` next to existing file ops.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 68-103, 190.
- Corpus: `internal/bmad/executor.go` (corpus 25304), `internal/bmad/artifacts.go` (corpus 16285), `internal/bmad/types.go` (corpus 20748).
- Skills: `/golang-testing`, `/golang-error-handling`.

## Acceptance Criteria

AC-1: OutputPaths populated on node complete
- Given a BMAD process node with outputs `["PRD.md"]`
- And the underlying artifact file is written to `_bmad-output/planning-artifacts/PRD.md`
- When the executor transitions the node to `complete`
- Then `node.OutputPaths["PRD.md"]` equals the absolute path of that file
- And the `bmad:node:artifacts` event carries the updated map

AC-2: Unmapped artifacts skipped
- Given a node whose outputs include `"code"` (returns `""` from `ResolveArtifactPath`)
- When the node completes
- Then `OutputPaths` does not contain a `"code"` key
- And no error is logged

AC-3: Missing files skipped
- Given a node whose outputs include `"PRD.md"` but the file was never written to disk
- When the node completes
- Then `OutputPaths["PRD.md"]` is not set
- And `NodeArtifactEvent.Missing` contains `"PRD.md"`

AC-4: Start-of-run clears prior paths
- Given a saved workflow whose nodes carry `OutputPaths` from a prior run
- When `StartWorkflow` begins
- Then every node's `OutputPaths` is re-initialized to an empty map before the DAG walks

AC-5: Concurrent-safe writes
- Given two ready nodes A and B complete simultaneously
- When both try to update their own `OutputPaths`
- Then no race is detected under `go test -race`
- And both maps are fully populated on return

AC-6: Legacy workflows round-trip
- Given a pre-existing saved `WorkflowDef` JSON without `outputPaths`
- When the app loads and saves it
- Then JSON round-trips without introducing stale `outputPaths` fields

## BDD Test Scenarios

```gherkin
Feature: Executor writes OutputPaths on complete

  Scenario: Happy path — mapped artifact exists
    Given a process node N with outputs ["PRD.md"]
    And _bmad-output/planning-artifacts/PRD.md exists on disk
    When the executor completes N
    Then N.OutputPaths["PRD.md"] equals the absolute resolved path
    And a bmad:node:artifacts event fires with the updated map

  Scenario: Unmapped artifact
    Given a process node N with outputs ["code"]
    When the executor completes N
    Then N.OutputPaths does not contain "code"
    And no error is returned

  Scenario: Missing on disk
    Given a process node N with outputs ["PRD.md"]
    And _bmad-output/planning-artifacts/PRD.md does NOT exist
    When the executor completes N
    Then N.OutputPaths does not contain "PRD.md"
    And NodeArtifactEvent.Missing includes "PRD.md"

  Scenario: Start-of-run clear
    Given a saved workflow with N.OutputPaths = {"PRD.md": "/stale/path.md"}
    When StartWorkflow begins
    Then N.OutputPaths is empty before the first node runs

  Scenario: Parallel node completion
    Given nodes A and B ready to execute simultaneously
    And each writes its own OutputPaths on complete
    When the executor runs them with -race
    Then no race is detected
    And both maps are fully populated
```

## Tasks / Subtasks

- [x] Task 1: Add OutputPaths + InputPaths fields to WorkflowNode (AC-1, AC-6)
  - [x] Add `OutputPaths map[string]string` in `internal/bmad/types.go`
  - [x] Add `InputPaths map[string]string` for symmetry (used by breadcrumbs-06)
  - [x] Round-trip test with legacy JSON (AC-6)
- [x] Task 2: Use ResolveArtifactPath + os.Stat to skip missing (AC-1, AC-3)
- [x] Task 3: Populate OutputPaths on node complete (AC-1, AC-2, AC-3)
  - [x] `resolveOutputPaths` helper iterates `proc.Outputs`
  - [x] Skip empty/missing silently
  - [x] Write under `state.mu.Lock()` / `Unlock()`
- [x] Task 4: Emit updated event payload (AC-1)
  - [x] `NodeArtifactEvent.Paths map[string]string` added
  - [x] Populated alongside `Found` / `Missing`
- [x] Task 5: Clear on start (AC-4)
  - [x] `StartWorkflow` zeroes `OutputPaths` + `InputPaths` per node
- [x] Task 6: Race test (AC-5)
  - [x] Two parallel `completeNode` goroutines; pass under `-race`

## Definition of Done

- [x] All acceptance criteria pass (7/7 ACs in executor_outputpaths_test.go)
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ code coverage on new/modified files (bmad 91.0%, completeNode 96.2%, resolveOutputPaths 100%)
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `/simplify` run on all modified code
- [x] Code review: no CRITICAL/HIGH issues (coderabbit PASS)
