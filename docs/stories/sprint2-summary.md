# Sprint 2 Summary: Repo-Scoped BMAD Workflow Builder

## Vision

Transform the BMAD workflow builder from a **generic, globally-scoped tool** into a **repo-aware sprint execution engine** where `sprint-status.yaml` is the single source of truth. When you open the workflow builder, it locks to a specific repo and shows you exactly what stories need work, what's in progress, and what's done -- all driven by the BMAD method's sprint tracking file.

---

## What Changes

### Before (Current State)
- Workflows stored globally at `~/.mashed/workflows/` with no repo association
- Repo is selected at execution time via a dropdown in the bottom bar
- Sidebar shows generic BMAD processes, templates, and saved workflows
- No awareness of sprint data, epics, or user stories
- Canvas nodes represent abstract BMAD processes with no sprint context

### After (Sprint 2)
- Workflows are **scoped to a repo** via `WorkflowDef.RepoPath`
- Repo is selected **before entering** the builder (picker modal)
- Sidebar gains a **Sprint tab** showing epics/stories from `sprint-status.yaml`
- Stories are **draggable onto the canvas** to create sprint-linked workflow nodes
- Completing a workflow node **auto-advances** the linked story's status
- A **context header** shows repo name, branch, and sprint progress at all times

---

## Architecture Changes

### Go Backend

| File | Change | Story |
|------|--------|-------|
| `internal/bmad/sprint.go` | **NEW** -- Sprint status parser, types (`SprintStatus`, `SprintEpic`, `SprintStory`), `UpdateStoryStatus()` | 1, 2 |
| `internal/bmad/sprint_test.go` | **NEW** -- Table-driven tests for parser and status updates | 1, 2 |
| `internal/bmad/types.go` | Add `RepoPath` to `WorkflowDef`, `StoryID` to `WorkflowNode`, new sentinel errors | 3, 7 |
| `internal/bmad/storage.go` | Add `ListWorkflowsByRepo(repoPath)` method | 3 |
| `internal/bmad/executor.go` | Auto-update story status on node completion, emit `bmad:sprint:updated` events | 7 |
| `app.go` | Add `GetSprintStatus()`, `UpdateStoryStatus()`, `ListBmadWorkflowsByRepo()` bindings; update `CreateFromTemplate()` signature | 2, 3 |

### Svelte Frontend

| File | Change | Story |
|------|--------|-------|
| `frontend/src/App.svelte` | Add repo picker flow before entering workflow builder | 4 |
| `frontend/src/components/bmad/RepoPickerModal.svelte` | **NEW** -- Modal to select target repo | 4 |
| `frontend/src/views/WorkflowBuilder.svelte` | Accept `repoPath` prop, load sprint data, handle story drops, listen for sprint events | 4, 5, 7 |
| `frontend/src/components/bmad/ProcessSidebar.svelte` | Add 4th "Sprint" tab, accept `sprintStatus` prop | 5 |
| `frontend/src/components/bmad/SprintPanel.svelte` | **NEW** -- Epic/story list with status colors, draggable stories | 5 |
| `frontend/src/components/bmad/ExecutionBar.svelte` | Remove repo dropdown, accept `repoPath` prop, show repo label | 6 |
| `frontend/src/components/bmad/ProcessNode.svelte` | Show story status badge when node is linked to a story; selection highlight styling | 7, 9 |
| `frontend/src/components/bmad/NodeConfigPanel.svelte` | Show linked story info in config panel | 7 |
| `frontend/src/components/bmad/CanvasPane.svelte` | Wire xyflow deletion, reconnection, and multi-select props/events | 9 |
| `frontend/src/components/bmad/RepoContextBar.svelte` | **NEW** -- Thin header bar showing repo name, branch, sprint progress | 8 |

### New Types

```go
// Sprint status types (internal/bmad/sprint.go)
type StoryStatus string   // backlog | ready-for-dev | in-progress | review | done
type EpicStatus string    // backlog | in-progress | done

type SprintStory struct {
    ID, EpicID string
    Status     StoryStatus
    Sequence   int
}

type SprintEpic struct {
    ID      string
    Status  EpicStatus
    Stories []SprintStory
}

type SprintStatus struct {
    Generated, LastUpdated, Project, ProjectKey string
    TrackingSystem, StoryLocation               string
    Epics                                       []SprintEpic
}
```

### New Wails Events

| Event | Payload | When |
|-------|---------|------|
| `bmad:sprint:updated` | `{ storyId, status }` | After a workflow node completion auto-advances a story |

---

## Stories at a Glance

### Story 1: Sprint Status YAML Parser (P0, Backend, M)
**What:** Go parser for `_bmad-output/implementation-artifacts/sprint-status.yaml` using `yaml.v3` Node API to preserve key ordering. Distinguishes epics (`epic-*` keys) from stories (`N-N-*` keys) in the flat `development_status` map.
**Key types:** `SprintStatus`, `SprintEpic`, `SprintStory`, `StoryStatus`, `EpicStatus`
**Key function:** `ParseSprintStatus(repoPath string) (SprintStatus, error)`
**Tests:** 5 BDD scenarios covering full parse, missing file, empty status, malformed YAML, status validation.

### Story 2: Wails Bindings for Sprint Status (P0, Backend, S)
**What:** Expose parser to frontend via `app.go` bindings. Add `UpdateStoryStatus()` with yaml.Node round-trip to modify status in-place while preserving formatting.
**Key bindings:** `GetSprintStatus(repoPath)`, `UpdateStoryStatus(repoPath, storyID, newStatus)`
**Tests:** Round-trip read/update, validation errors, nonexistent story handling.

### Story 3: Add RepoPath to WorkflowDef + Storage Filtering (P0, Backend, M)
**What:** Add `RepoPath string` field to `WorkflowDef` and `ListWorkflowsByRepo()` to `Storage`. Existing global workflows gracefully degrade (empty RepoPath excluded from repo-filtered lists). Update `CreateFromTemplate` to accept repoPath.
**Backward compat:** `json:"repoPath,omitempty"` means existing JSON files load fine with empty string.
**Tests:** Filtered listing, backward compatibility, round-trip serialization.

### Story 4: Repo Context Flow (P0, Fullstack, M)
**What:** Plumb repo context from App.svelte to WorkflowBuilder. New `RepoPickerModal` shown when user opens workflow builder -- lists repos from `ListRepoChoices()`, user clicks to select, repoPath flows as prop. All API calls in WorkflowBuilder scoped to this repo.
**UX flow:** Feed -> Click Workflows -> Repo Picker Modal -> Select Repo -> WorkflowBuilder (scoped)
**New component:** `RepoPickerModal.svelte`
**Tests:** Picker display, selection navigation, cancel behavior, filtered workflow list.

### Story 5: Sprint Panel in Sidebar (P1, Frontend, L)
**What:** Add "Sprint" tab (4th tab) to ProcessSidebar showing epics/stories from sprint-status.yaml. Stories are draggable onto canvas with `application/bmad-story` data transfer. Status colors: backlog=gray, ready-for-dev=blue, in-progress=amber, review=purple, done=green. Collapsible epic sections.
**New component:** `SprintPanel.svelte`
**Canvas integration:** Story drops create nodes with `processId: 'bmad-dev-story'` and `storyId` link.
**Tests:** Sprint data display, status colors, story drag-to-canvas, empty state, epic collapse.

### Story 6: ExecutionBar Simplification (P1, Frontend, S)
**What:** Remove the repo dropdown from ExecutionBar (repo is now implicit from WorkflowBuilder prop). Replace with a static repo label showing the directory basename with full-path tooltip. Model selector and play/pause/stop unchanged.
**Removes:** `ListRepoChoices` import, `repoChoices` array, `selectedRepo`, repo `<select>`.
**Adds:** `repoPath` prop, repo label element.
**Tests:** No dropdown present, label shows repo name, tooltip shows full path, run button disabled without repoPath.

### Story 7: Sprint-Aware Canvas Nodes (P2, Fullstack, L)
**What:** Nodes linked to sprint stories display a status badge. On node completion, executor auto-advances the linked story's status in sprint-status.yaml. Node failure does NOT update story status. `StoryID` field added to `WorkflowNode` type. Sprint update events refresh the sidebar.
**Auto-advance logic:** Completion -> "in-progress" (default) or "done" (support phase processes).
**New field:** `WorkflowNode.StoryID string`
**Tests:** Badge rendering, no badge without storyId, completion updates story, failure doesn't, round-trip save/load.

### Story 8: Repo Context Header Bar (P3, Frontend, S)
**What:** Thin (28px) horizontal bar at top of canvas area showing repo name, branch, and sprint progress (`3/8 stories done` with progress bar). Reactive -- updates when sprint data changes.
**New component:** `RepoContextBar.svelte`
**Tests:** Repo name display, sprint progress calculation, null sprint fallback, reactive updates.

### Story 9: Canvas Editing -- Delete, Reconnect, Multi-Select (P1, Frontend, M)
**What:** Add essential canvas editing operations missing from the workflow builder. Wire xyflow's built-in `deleteKeyCode`, `selectionKeyCode`, `edgesReconnectable` props and their event handlers. Users can select nodes/edges and press Delete/Backspace to remove them, drag edge endpoints to reconnect to different nodes, and Shift+click or drag-rectangle to multi-select for bulk operations. Running nodes are protected from deletion.
**Modifies:** `CanvasPane.svelte` (xyflow props/events), `WorkflowBuilder.svelte` (handlers), `ProcessNode.svelte` (selection styling)
**No backend changes.** Pure frontend wiring of existing xyflow capabilities.
**Tests:** Single/multi node deletion, edge deletion, edge reconnection, running node protection, selection visuals.

---

## Execution Plan

```
Wave 1 (parallel, no deps):
  Story 1: Sprint Status YAML Parser     ──┐
  Story 3: RepoPath on WorkflowDef        ──┤
  Story 9: Canvas Editing (Delete/Rewire) ──┤
                                             │
Wave 2 (parallel, deps on Wave 1):          │
  Story 2: Wails Bindings (needs 1)      ───┤
  Story 4: Repo Context Flow (needs 3)   ───┤
                                             │
Wave 3 (parallel, deps on Wave 2):          │
  Story 5: Sprint Panel (needs 2, 4)     ───┤
  Story 6: ExecBar Simplify (needs 4)    ───┤
                                             │
Wave 4:                                      │
  Story 7: Sprint-Aware Nodes (needs 2, 5) ─┤
                                             │
Wave 5:                                      │
  Story 8: Repo Context Header (needs 4, 5) ┘
```

---

## Key Design Decisions

1. **Storage stays at `~/.mashed/workflows/`** -- Workflows are filtered by repo in-memory rather than stored per-repo. This preserves backward compatibility and simplifies cleanup.

2. **Repo selected before entering builder** -- A picker modal ensures repo context is always set, eliminating the confusing "which repo am I targeting?" question during workflow design.

3. **sprint-status.yaml is read-only from the sidebar, writable from executor** -- The Sprint panel displays data; the executor auto-advances story status on node completion. Manual overrides go through `UpdateStoryStatus` binding.

4. **Story drag creates "bmad-dev-story" nodes by default** -- When a story is dragged from the Sprint panel to the canvas, it creates a node pre-configured for the story development process. The user can change the process type afterward.

5. **YAML round-trip preserves formatting** -- `UpdateStoryStatus` uses `yaml.Node` to modify in-place, keeping comments and whitespace intact.

---

## Risk Register

| Risk | Mitigation | Story |
|------|------------|-------|
| YAML key ordering lost in Go maps | Use `yaml.v3` Node API with pairwise iteration | 1 |
| Concurrent status updates race | Per-repo mutex in UpdateStoryStatus | 2 |
| Orphaned global workflows invisible | ListWorkflows still returns all; filtered view is additive | 3 |
| User opens builder with no repos | RepoPickerModal shows empty state, Escape returns to feed | 4 |
| Multiple nodes linked to same story | Last-to-complete wins; documented behavior | 7 |
| sprint-status.yaml missing | Graceful degradation: Sprint tab shows "No sprint data" | 5, 8 |
| Deleting a running node orphans tmux process | Protect running/paused nodes from deletion (AC-7) | 9 |
| Accidental bulk delete | Selection requires explicit Shift+click or drag rectangle | 9 |
