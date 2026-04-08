# Story 7: Sprint-Aware Canvas Nodes

**Priority:** P2-medium
**Domain:** fullstack
**Estimated Complexity:** L
**Depends On:** Story 2, Story 5
**Status:** done

## Description

Enhance workflow canvas nodes to display and interact with sprint story status. When a node is linked to a sprint story (via storyId from Story 5), the node shows the story's current status with a colored badge. Completing a workflow node can optionally update the linked story's status in sprint-status.yaml. This closes the loop between workflow execution and sprint tracking.

## Developer Notes

### Architecture
- **Modify:** `frontend/src/components/bmad/ProcessNode.svelte` -- show story status badge when storyId is present
- **Modify:** `frontend/src/components/bmad/NodeConfigPanel.svelte` -- add story link display and status override control
- **Modify:** `internal/bmad/types.go` -- add `StoryID` field to `WorkflowNode`
- **Modify:** `internal/bmad/executor.go` -- after node completion, optionally update story status
- **Modify:** `app.go` -- pipe story status updates from executor

### WorkflowNode Type Change

```go
// In types.go, add to WorkflowNode:
type WorkflowNode struct {
    ID         string             `json:"id"`
    ProcessID  string             `json:"processId"`
    Label      string             `json:"label"`
    Position   Position           `json:"position"`
    Status     WorkflowNodeStatus `json:"status"`
    Config     map[string]string  `json:"config"`
    TmuxTarget string             `json:"tmuxTarget"`
    StoryID    string             `json:"storyId,omitempty"` // NEW: linked sprint story
}
```

### ProcessNode.svelte Enhancement

Add a story status badge below the node label:
```svelte
$: storyId = data.storyId || '';
$: storyStatus = data.storyStatus || '';

{#if storyId}
  <div class="story-badge" style="color: {statusColor(storyStatus)}">
    {storyId} [{storyStatus}]
  </div>
{/if}
```

Status color mapping (reuse from SprintPanel):
```javascript
const statusColors = {
  'backlog': 'var(--text-muted)',
  'ready-for-dev': 'var(--accent-blue, #3d9eff)',
  'in-progress': 'var(--accent-amber, #f0a500)',
  'review': 'var(--accent-purple, #9d6fff)',
  'done': 'var(--accent-green, #00e57a)',
};
```

### Executor Changes

In `executor.go`, after `completeNode`:
```go
func (e *Executor) completeNode(state *execState, idx int, nodeID string) {
    state.mu.Lock()
    state.exec.Nodes[idx].Status = NodeComplete
    storyID := state.exec.Nodes[idx].StoryID  // NEW
    repoPath := state.exec.RepoPath            // NEW
    state.mu.Unlock()
    e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeComplete})

    // Auto-advance story status on node completion
    if storyID != "" && repoPath != "" {
        // Determine target status based on process type
        targetStatus := "in-progress" // default
        proc, ok := ProcessByID(state.exec.Nodes[idx].ProcessID)
        if ok && proc.Phase == PhaseSupport {
            targetStatus = "done"
        }
        if err := UpdateStoryStatus(repoPath, storyID, targetStatus); err != nil {
            // Log but don't fail the node
            log.Printf("bmad: failed to update story %s status: %v", storyID, err)
        }
        e.emitEvent("bmad:sprint:updated", map[string]string{"storyId": storyID, "status": targetStatus})
    }
}
```

### NodeConfigPanel.svelte Enhancement

Show linked story info in the config panel when a node is selected:
```svelte
{#if node?.data?.storyId}
  <div class="config-section">
    <h4>Linked Story</h4>
    <div class="story-link">
      <span class="story-id">{node.data.storyId}</span>
      <span class="story-status" style="color: {statusColor(node.data.storyStatus)}">{node.data.storyStatus}</span>
    </div>
  </div>
{/if}
```

### Technical Considerations
- The `storyId` in node data is set during story drop (Story 5) and persisted when the workflow is saved
- Story status on nodes should be refreshed when `bmad:sprint:updated` events fire
- The auto-advance logic is intentionally simple: node completion moves story to "in-progress" or "done" depending on the process phase. Users can override via `UpdateStoryStatus` binding.
- Frontend must handle the `bmad:sprint:updated` Wails event to refresh sprint panel data
- Nodes without storyId are unaffected -- the ProcessNode renders identically to before

### Risks & Edge Cases
- Multiple nodes linked to the same story -- last-to-complete wins for status update
- Node failure should NOT update story status (only completion triggers update)
- Story might be manually updated while workflow is running -- the node badge may go stale until refresh
- If `UpdateStoryStatus` fails (YAML file locked, missing), the node still completes successfully

### Reference Files
- `frontend/src/components/bmad/ProcessNode.svelte` -- node rendering to extend
- `frontend/src/components/bmad/NodeConfigPanel.svelte` -- config panel to extend
- `internal/bmad/executor.go` -- `completeNode` and `failNode` functions
- `internal/bmad/types.go` -- `WorkflowNode` struct

## Acceptance Criteria

AC-1: Node displays story badge when linked
- Given a canvas node with storyId "1-2-account-mgmt" and storyStatus "ready-for-dev"
- When rendered on the canvas
- Then the node shows "1-2-account-mgmt" with a blue status indicator below the label

AC-2: Node without story shows no badge
- Given a canvas node with no storyId
- When rendered on the canvas
- Then no story badge is visible
- And the node looks identical to the current design

AC-3: Node completion updates story status
- Given a running workflow with a node linked to story "1-2-acct" (current status "ready-for-dev")
- When the node completes execution
- Then the story status in sprint-status.yaml is updated to "in-progress"
- And a `bmad:sprint:updated` event is emitted

AC-4: Node failure does not update story
- Given a running workflow with a node linked to story "1-2-acct"
- When the node fails execution
- Then the story status in sprint-status.yaml is NOT modified

AC-5: StoryID persists in workflow save
- Given a canvas with nodes that have storyId set
- When the workflow is saved
- Then the saved WorkflowDef JSON includes storyId for each linked node
- And loading the workflow restores the storyId values

AC-6: Config panel shows story link
- Given a node linked to story "2-1-personality" with status "backlog"
- When the user selects the node
- Then the NodeConfigPanel shows "Linked Story: 2-1-personality" with status "backlog"

## BDD Test Scenarios

### Scenario 1: Sprint-aware node rendering and execution

```gherkin
Feature: Sprint-aware canvas nodes

  Scenario: Node with linked story shows badge
    Given a ProcessNode rendered with data containing storyId "1-2-acct" and storyStatus "ready-for-dev"
    Then the node displays text "1-2-acct"
    And the status indicator uses the blue color variable

  Scenario: Node without story has no badge
    Given a ProcessNode rendered with data containing no storyId
    Then no story badge element exists in the node

  Scenario: Completing a node updates the linked story
    Given an executor running a workflow with node "n1" linked to story "1-2-acct" in repo "/test/repo"
    And the sprint-status.yaml has story "1-2-acct" at "ready-for-dev"
    When node "n1" completes successfully
    Then UpdateStoryStatus is called with storyId "1-2-acct" and status "in-progress"
    And a bmad:sprint:updated event is emitted

  Scenario: Failed node does not update story status
    Given an executor running a workflow with node "n1" linked to story "1-2-acct"
    When node "n1" fails
    Then UpdateStoryStatus is NOT called
    And sprint-status.yaml is unchanged

  Scenario: StoryID round-trips through save/load
    Given a workflow with node "n1" having storyId "2-1-personality"
    When the workflow is saved and loaded back
    Then node "n1" still has storyId "2-1-personality"
```

## Tasks / Subtasks

- [ ] Task 1: Add StoryID to WorkflowNode type (AC: AC-5)
  - [ ] Subtask 1a: Add `StoryID string \`json:"storyId,omitempty"\`` to `WorkflowNode` in `internal/bmad/types.go`
  - [ ] Subtask 1b: Verify existing tests pass with new field (backward compat)

- [ ] Task 2: Enhance ProcessNode to show story badge (AC: AC-1, AC-2)
  - [ ] Subtask 2a: Add story badge rendering in `ProcessNode.svelte` (conditional on storyId)
  - [ ] Subtask 2b: Implement `statusColor` function with the 5 status-to-color mappings
  - [ ] Subtask 2c: Style the badge to be compact and not overflow the node width

- [ ] Task 3: Update executor to auto-advance story status (AC: AC-3, AC-4)
  - [ ] Subtask 3a: Modify `completeNode` in `executor.go` to read StoryID and RepoPath
  - [ ] Subtask 3b: Call `UpdateStoryStatus` on completion (not failure) with appropriate target status
  - [ ] Subtask 3c: Emit `bmad:sprint:updated` event after successful status update
  - [ ] Subtask 3d: Write executor test verifying story status update on completion and no-op on failure

- [ ] Task 4: Enhance NodeConfigPanel with story info (AC: AC-6)
  - [ ] Subtask 4a: Add "Linked Story" section to `NodeConfigPanel.svelte` showing storyId and status
  - [ ] Subtask 4b: Style to match existing config panel sections

- [ ] Task 5: Wire sprint update events in WorkflowBuilder (AC: AC-3)
  - [ ] Subtask 5a: Listen for `bmad:sprint:updated` events in WorkflowBuilder
  - [ ] Subtask 5b: Refresh sprintStatus by re-calling `GetSprintStatus(repoPath)` on event
  - [ ] Subtask 5c: Update node storyStatus in canvas nodes when sprint data refreshes

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
