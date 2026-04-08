# Story 5: Post-Completion Artifact Verification & Frontend Status

**Priority:** P1-high
**Domain:** fullstack
**Estimated Complexity:** L
**Depends On:** sprint4-01-artifact-path-system, sprint4-02-file-aware-context
**Status:** ready

## Description

After a workflow node completes, the executor now verifies which expected output artifacts were actually written to disk and emits a `bmad:node:artifacts` event. The frontend displays artifact status indicators on completed ProcessNodes (green checks for found, amber warnings for missing) and shows artifact details in the NodeConfigPanel. A new Wails binding `GetArtifactStatus` lets the frontend query artifact existence on demand. This closes the feedback loop -- users can see at a glance whether each process step actually produced its outputs.

## Developer Notes

### Architecture

**Backend changes:**
- **Modified file:** `internal/bmad/executor.go`
  - In `completeNode()`: after marking complete, look up the process's expected outputs, call `VerifyArtifacts(repoPath, proc.Outputs)`, and emit a new event `bmad:node:artifacts` with `NodeArtifactEvent` data.
  - Need to pass the `ProcessID` info to `completeNode()`. Currently `completeNode` takes `(state, idx, nodeID)`. Add the process lookup inside the function (read ProcessID from node, look up in registry). Only emit artifact events for `NodeTypeProcess` nodes (control flow nodes don't have process outputs).
- **Modified file:** `internal/bmad/types.go`
  - Add `NodeArtifactEvent` struct:
    ```go
    type NodeArtifactEvent struct {
        ExecID  string   `json:"execId"`
        NodeID  string   `json:"nodeId"`
        Found   []string `json:"found"`
        Missing []string `json:"missing"`
    }
    ```
- **Modified file:** `app_bmad.go`
  - Add `GetArtifactStatus(repoPath, artifactName string) (bool, string, error)` binding.
  - Resolves the artifact path via `bmad.ResolveArtifactPath`, checks existence, returns (exists, fullPath, nil) or error.
- **Modified file:** `internal/bmad/executor_test.go`
  - Add tests for artifact event emission in `completeNode` flow.

**Frontend changes:**
- **Modified file:** `frontend/src/components/bmad/ProcessNode.svelte`
  - Accept new prop `artifactStatus` (object with `found: string[]` and `missing: string[]`).
  - After the status row, if `status === 'complete'` and `artifactStatus` has data, show small indicators: green check icon per found artifact, amber warning icon per missing artifact.
  - Keep it compact -- just icons with tooltips showing artifact names.
- **Modified file:** `frontend/src/components/bmad/NodeConfigPanel.svelte`
  - For completed process nodes with artifact data, show an "Artifacts" section listing each expected output with a found/missing badge.
  - Use the existing field styling pattern (`.field`, `.field-label`).
- **Modified file:** `frontend/src/views/WorkflowBuilder.svelte` (or wherever events are wired)
  - Listen for `bmad:node:artifacts` events from Wails runtime.
  - Store artifact status per node in the execution state.
  - Pass artifact status data to ProcessNode and NodeConfigPanel via node data.

### Technical Considerations
- `completeNode()` is called for ALL node types (process, condition, merge, transform, loop). Only emit artifact events when the node is a process node with a valid ProcessID. For non-process nodes, skip the verification.
- The `VerifyArtifacts` call in `completeNode` needs `repoPath` -- already available via `state.exec.RepoPath`.
- The `ProcessID` is available from `state.exec.Nodes[idx].ProcessID`.
- Event emission order: emit `bmad:node:status` (complete) first, then `bmad:node:artifacts`. This ensures the frontend updates status before processing artifact data.
- Frontend: the `bmad:node:artifacts` event arrives asynchronously. Store it in a reactive map keyed by nodeID. If the node data in the canvas is immutable (xyflow pattern), propagate via a store or by updating node data.
- `GetArtifactStatus` binding is a simple utility for on-demand checks. It wraps `ResolveArtifactPath` + `os.Stat`.

### Risks & Edge Cases
- Process nodes with only unmapped outputs (e.g., `bmad-quick-dev` outputs `["code"]`) will have empty found/missing lists from `VerifyArtifacts`. The frontend should handle this gracefully (show nothing or "no tracked artifacts").
- If `completeNode` is called from `executeControlNode` (merge/condition), the ProcessID might be empty or invalid. Guard against this.
- Race condition: artifact event could arrive before the status event if both are emitted in quick succession. The frontend should handle out-of-order events (only display artifacts for completed nodes).
- `VerifyArtifacts` does I/O (os.Stat). This is fine in `completeNode` since it runs once per node completion, but should not block the critical path. The emission is best-effort.

### Reference Files
- `internal/bmad/executor.go` -- `completeNode()` at line ~787, `executeNode()` at line ~700
- `internal/bmad/artifacts.go` -- `VerifyArtifacts`, `ResolveArtifactPath` (from Story 1)
- `internal/bmad/types.go` -- existing event types `NodeStatusEvent`, `ExecStatusEvent`
- `app_bmad.go` -- existing bindings pattern (nil-check, delegate, return)
- `frontend/src/components/bmad/ProcessNode.svelte` -- current node component
- `frontend/src/components/bmad/NodeConfigPanel.svelte` -- current config panel
- `/wails` skill -- for Wails event listener patterns in Svelte

## Acceptance Criteria

AC-1: Artifact verification runs after process node completion
- Given a workflow with a process node that has expected outputs `["PRD.md"]`
- And the process node completes execution
- When `completeNode` runs
- Then `VerifyArtifacts` is called with the repo path and the process's output list
- And a `bmad:node:artifacts` event is emitted with the found/missing results

AC-2: Non-process nodes do not emit artifact events
- Given a workflow with a condition node or merge node
- When that node completes
- Then no `bmad:node:artifacts` event is emitted

AC-3: GetArtifactStatus binding works
- Given a repo path and artifact name `"PRD.md"`
- When `GetArtifactStatus(repoPath, "PRD.md")` is called
- Then it returns `(true, "/repo/_bmad-output/planning-artifacts/PRD.md", nil)` if the file exists
- And `(false, "/repo/_bmad-output/planning-artifacts/PRD.md", nil)` if the file does not exist

AC-4: ProcessNode shows artifact indicators after completion
- Given a completed process node with artifact data `{found: ["PRD.md"], missing: ["ux-spec.md"]}`
- When the ProcessNode renders
- Then a green indicator appears for "PRD.md"
- And an amber indicator appears for "ux-spec.md"

AC-5: NodeConfigPanel shows artifact details for completed nodes
- Given a completed process node is selected in the config panel
- And it has artifact data `{found: ["PRD.md"], missing: ["ux-spec.md"]}`
- When the panel renders
- Then an "Artifacts" section lists "PRD.md" with a found badge
- And lists "ux-spec.md" with a missing badge

AC-6: Nodes with only unmapped outputs show no artifact indicators
- Given a completed process node for `bmad-quick-dev` (outputs: `["code"]`)
- When artifact verification runs
- Then found and missing are both empty
- And no artifact indicators appear on the node

## BDD Test Scenarios

### Scenario 1: Artifact event emission

```gherkin
Feature: Post-completion artifact verification

  Scenario: Process node emits artifact event on completion
    Given a workflow with one process node "n1" (process: bmad-create-prd, outputs: ["PRD.md"])
    And the repo has "_bmad-output/planning-artifacts/PRD.md" on disk
    And the mock runner completes the tmux session immediately
    When the workflow executes and node "n1" completes
    Then a "bmad:node:artifacts" event is emitted
    And the event contains found: ["PRD.md"] and missing: []

  Scenario: Process node with missing artifact
    Given a workflow with one process node "n1" (process: bmad-create-prd, outputs: ["PRD.md"])
    And the repo does NOT have "_bmad-output/planning-artifacts/PRD.md" on disk
    When the workflow executes and node "n1" completes
    Then a "bmad:node:artifacts" event is emitted
    And the event contains found: [] and missing: ["PRD.md"]

  Scenario: Condition node does not emit artifact event
    Given a workflow with a condition node "c1"
    When "c1" completes
    Then no "bmad:node:artifacts" event is emitted
```

### Scenario 2: GetArtifactStatus binding

```gherkin
Feature: Artifact status binding

  Scenario: Artifact exists on disk
    Given a repo at "/tmp/test-repo" with "_bmad-output/planning-artifacts/PRD.md"
    When GetArtifactStatus("/tmp/test-repo", "PRD.md") is called
    Then it returns (true, "/tmp/test-repo/_bmad-output/planning-artifacts/PRD.md", nil)

  Scenario: Artifact does not exist
    Given a repo at "/tmp/test-repo" without the PRD.md file
    When GetArtifactStatus("/tmp/test-repo", "PRD.md") is called
    Then it returns (false, "/tmp/test-repo/_bmad-output/planning-artifacts/PRD.md", nil)

  Scenario: Unmapped artifact returns empty path
    Given any repo path
    When GetArtifactStatus(repoPath, "code") is called
    Then it returns (false, "", nil)
```

### Scenario 3: Frontend artifact display

```gherkin
Feature: Artifact status visualization

  Scenario: Completed node shows green indicators for found artifacts
    Given a completed ProcessNode with artifactStatus.found = ["PRD.md", "ux-spec.md"]
    When the node renders
    Then 2 green check indicators are visible
    And hovering shows "PRD.md" and "ux-spec.md" as tooltips

  Scenario: Completed node shows amber indicators for missing artifacts
    Given a completed ProcessNode with artifactStatus.missing = ["architecture.md"]
    When the node renders
    Then 1 amber warning indicator is visible

  Scenario: Pending/running nodes show no artifact indicators
    Given a running ProcessNode
    When the node renders
    Then no artifact indicator section is visible

  Scenario: Config panel artifact details
    Given a completed process node is selected
    And artifactStatus is {found: ["PRD.md"], missing: ["ux-spec.md"]}
    When the NodeConfigPanel renders
    Then the "Artifacts" section shows "PRD.md" with a green found badge
    And "ux-spec.md" with an amber missing badge
```

## Tasks / Subtasks

- [ ] Task 1: Add NodeArtifactEvent type to types.go (AC: AC-1)
  - [ ] Subtask 1a: Define `NodeArtifactEvent` struct with ExecID, NodeID, Found, Missing fields and JSON tags
- [ ] Task 2: Add artifact verification to completeNode in executor.go (AC: AC-1, AC-2, AC-6)
  - [ ] Subtask 2a: In `completeNode`, look up ProcessID from node, guard for non-process nodes
  - [ ] Subtask 2b: Call `VerifyArtifacts(state.exec.RepoPath, proc.Outputs)` for process nodes
  - [ ] Subtask 2c: Emit `bmad:node:artifacts` event with `NodeArtifactEvent` data
  - [ ] Subtask 2d: Ensure event is emitted AFTER `bmad:node:status` complete event
- [ ] Task 3: Add GetArtifactStatus binding to app_bmad.go (AC: AC-3)
  - [ ] Subtask 3a: Implement `GetArtifactStatus(repoPath, artifactName string) (bool, string, error)`
  - [ ] Subtask 3b: Use `ResolveArtifactPath` + `os.Stat` for existence check
- [ ] Task 4: Update ProcessNode.svelte with artifact indicators (AC: AC-4, AC-6)
  - [ ] Subtask 4a: Add `artifactStatus` prop (default empty)
  - [ ] Subtask 4b: Add artifact indicator row after status row for completed nodes
  - [ ] Subtask 4c: Style green check and amber warning indicators (compact, with tooltips)
- [ ] Task 5: Update NodeConfigPanel.svelte with artifact details (AC: AC-5)
  - [ ] Subtask 5a: Add "Artifacts" section for completed process nodes
  - [ ] Subtask 5b: List each artifact with found (green) / missing (amber) badge
- [ ] Task 6: Wire artifact events in WorkflowBuilder (AC: AC-4, AC-5)
  - [ ] Subtask 6a: Listen for `bmad:node:artifacts` events from Wails runtime
  - [ ] Subtask 6b: Store artifact status per nodeID in reactive state
  - [ ] Subtask 6c: Pass artifact data through to ProcessNode and NodeConfigPanel
- [ ] Task 7: Backend tests (AC: AC-1, AC-2, AC-3, AC-6)
  - [ ] Subtask 7a: Test artifact event emission for process nodes using test harness
  - [ ] Subtask 7b: Test no artifact event for condition/merge/transform nodes
  - [ ] Subtask 7c: Test GetArtifactStatus binding logic (file exists, missing, unmapped)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on modified Go files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] Frontend renders artifact indicators in `wails dev` manual test
