# Story breadcrumbs-08: MultiFileLoader node — frontend

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** breadcrumbs-07, breadcrumbs-06
**Status:** ready

## Description

Phase 2b Task 2b.2 (plan lines 117-121). Ship the Svelte side of MultiFileLoader: a dedicated node component showing one breadcrumb per configured path, a list-editor config panel reusing `ArrayEditorModal.svelte`, and sidebar visibility under Utilities.

## Developer Notes

### Architecture

Three surfaces:

1. **`frontend/src/components/bmad/MultiFileLoaderNode.svelte`** — new component. Renders header + N breadcrumbs, one per configured entry (`label` or positional `file[N]`). Reuse `.breadcrumb-row` styling from breadcrumbs-02.
2. **`frontend/src/components/bmad/NodeConfigPanel.svelte`** (corpus line 25766) — extend to recognize `nodeType === 'multiFileLoader'` and render a list editor. Reuse `ArrayEditorModal.svelte` (corpus line 9680) for add/edit/reorder.
3. **`frontend/src/components/bmad/ProcessSidebar.svelte`** (corpus line 22655) — surface MultiFileLoader in the Utilities group.

Entry shape matches breadcrumbs-07:

```ts
interface MultiFileEntry { label: string; path: string; }
```

Stored as a JSON-encoded string in `node.data.config.entries` (mirrors backend). Parse on render, stringify on save.

### Technical Considerations

- **Browse dialog per entry**: reuse the same OS-file-picker binding already used by File Loader (grep `BrowseFile` or similar in `frontend/wailsjs/go/main/App.d.ts`).
- **Reordering**: `ArrayEditorModal.svelte` already handles reorder — confirm by reading the component; if it lacks reorder, scope creeps. Flag in implementation.
- **Per-entry breadcrumb**: breadcrumb label is the entry's `label` (or `file[N]` fallback) and breadcrumb body is `formatBreadcrumb(entry.path)`.
- **Downstream wiring**: edges from this node carry `sourceHandle = <label>` or `file[N]` so breadcrumbs-06 auto-fill matches correctly by handle name.

### Risks & Edge Cases

- Duplicate label visual warning: the backend rejects at execute time; the UI should warn earlier — red outline on conflicting rows in the modal. Pure UX polish; include.
- Long lists: modal must scroll. Use the existing modal pattern.
- Empty list: node renders but has no outputs; downstream sees `—` breadcrumbs. Acceptable.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 117-127.
- Corpus: `frontend/src/components/bmad/ArrayEditorModal.svelte` (9680), `NodeConfigPanel.svelte` (25766), `ProcessSidebar.svelte` (22655), `ProcessNode.svelte` (20011 — breadcrumb pattern).
- Skill: `/xyflow` for custom node registration, `/playwright-cli` for AC validation.

## Acceptance Criteria

AC-1: Dragging from sidebar onto canvas creates a MultiFileLoader node
- Given the Utilities group in the sidebar
- When the user drags the MultiFileLoader item onto the canvas
- Then a new node appears with `nodeType = multiFileLoader` and an empty entries list

AC-2: Config panel lists entries via ArrayEditorModal
- Given a MultiFileLoader node selected on canvas
- When the user opens the config panel
- Then an Entries editor renders with add / edit / reorder actions
- And each row exposes `label` + `path` inputs and a browse-file button

AC-3: Node renders one breadcrumb per entry
- Given entries `[{label:"brief", path:"/x/a.md"}, {path:"/x/b.md"}]`
- When the canvas renders the node
- Then two breadcrumb rows appear: `brief` → `.../a.md` and `file[1]` → `.../b.md`
- And hovering each row shows the full absolute path via `title`

AC-4: Saves round-trip entries as JSON
- Given the user adds two entries and saves
- When the workflow is reloaded
- Then both entries parse back with identical labels and paths

AC-5: Downstream handle naming
- Given a downstream node with input `brief`
- When the edge connects from MultiFileLoader's `brief` output handle to the downstream input
- Then running MultiFileLoader causes auto-fill of the downstream `brief` slot per breadcrumbs-06

## BDD Test Scenarios

```gherkin
Feature: MultiFileLoader frontend

  Scenario: Add from sidebar
    Given the Utilities group shows a MultiFileLoader item
    When the user drags it onto the canvas
    Then a new MultiFileLoader node exists with zero entries

  Scenario: Two entries render two breadcrumbs
    Given a MultiFileLoader with entries [{label:"brief", path:"/x/a.md"}, {path:"/x/b.md"}]
    When the canvas renders the node
    Then two breadcrumb rows appear with basenames a.md and b.md
    And their prefix labels are "brief" and "file[1]"

  Scenario: Round-trip persistence
    Given the user adds entry {label:"plan", path:"/y/plan.md"} via the config panel
    When SaveBmadWorkflow fires and the workflow is reloaded
    Then the entry parses back identically

  Scenario: Downstream auto-fill with handle
    Given an edge from MultiFileLoader handle "brief" to a downstream process input "brief"
    And the downstream slot is empty
    When the MultiFileLoader node completes
    Then the downstream inputPaths["brief"] fills with the entry's absolute path
```

## Tasks / Subtasks

- [ ] Task 1: Create `MultiFileLoaderNode.svelte` (AC-3)
  - [ ] Header with node label
  - [ ] Loop over parsed entries; render `.breadcrumb-row` per entry using `formatBreadcrumb`
  - [ ] Source handles per entry (xyflow `<Handle>` with `id=<label or file[N]>`)
- [ ] Task 2: Register component with xyflow (AC-1)
  - [ ] Add to the `nodeTypes` map used by `CanvasPane.svelte`
- [ ] Task 3: Sidebar entry (AC-1)
  - [ ] Add MultiFileLoader item under Utilities in `ProcessSidebar.svelte`
- [ ] Task 4: Config panel integration (AC-2, AC-4)
  - [ ] Detect `nodeType === 'multiFileLoader'` in `NodeConfigPanel.svelte`
  - [ ] Render `ArrayEditorModal.svelte` over parsed entries
  - [ ] Stringify back to `node.data.config.entries` on save
  - [ ] Red outline on duplicate-label rows
- [ ] Task 5: Playwright AC (AC-1, AC-2, AC-3, AC-4)
  - [ ] Drag-create the node, add two entries, save, reload, assert breadcrumbs

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues

## Design Brief

### Layout composition
New component: `frontend/src/components/bmad/MultiFileLoaderNode.svelte`. Reuses `.process-node` / `.node-body` / `.phase-bar` / `.node-header` / `.status-row` structure from `ProcessNode.svelte` so visual kinship is automatic. Body structure top-to-bottom:
1. `.phase-bar` — uses `support` phase color (`var(--text-dim)`) since this is a utility node, matching File Loader.
2. `.node-header` — `FolderTree` or `Files` icon from lucide-svelte (colored `var(--text-dim)`) + label ("Multi-File Loader" or user-set alias).
3. `.path-list` — scrollable container holding N `.breadcrumb-item` rows (one per configured path). Each row: `file[i]` positional label (or user-supplied alias if provided via ArrayEditorModal) in `.artifacts` style + resolved `.breadcrumb-row` below it. Same stacked-pair rhythm as story -04.
4. Divider: `border-top: 1px solid var(--border-subtle); margin-top: var(--sp-xs); padding-top: var(--sp-xs);` — the one structural departure from ProcessNode, separating the variable-length path list from the fixed footer.
5. `.node-footer` — `count badge` (`N files`) in `--text-muted` + `.status-row` reused verbatim.

Handles: only a `source` handle (`Position.Right`). No target handle — MultiFileLoader has no upstream inputs; paths come from config.

### Typography plan
- Header label: `11px` Geist UI, weight `600`, `var(--text-primary)` — identical to `ProcessNode .node-label`.
- Positional label (`file[0]`, `file[1]`): `9px var(--font-mono)`, weight `600`, `var(--text-dim)` — identical to `.artifact-label`.
- Breadcrumb path: `9px var(--font-mono)`, weight `400`, `var(--text-muted)`, `line-height: 1.3`, `font-variant-numeric: tabular-nums` — identical to story -02 `.breadcrumb-row`.
- Count badge ("N files"): `9px var(--font-mono)`, weight `400`, `var(--text-muted)`, `letter-spacing: 0.3px` — matches `.status-text` existing rule.

### Color strategy
- Surface: `background: var(--bg-elevated)` (inherit `.process-node`).
- Border: `border: 1px solid var(--border-subtle)`; selected → `var(--accent-green)`.
- Phase bar: `var(--text-dim)` (utility nodes are color-neutral — they do no BMAD work).
- Path-list divider: `border-top: 1px solid var(--border-subtle)` — uses the *subtle* border, not emphasis, so the divider whispers.
- Positional labels: `var(--text-dim)`.
- Paths: `var(--text-muted)`.
- Unresolved em-dash: `var(--text-muted)` at `opacity: 0.7`.
- Count badge: `var(--text-muted)`.
- No accent color on this node — File Loader doesn't use one either; accent color belongs to BMAD-phase process nodes.

### Interaction model
- Click on node body → open `NodeConfigPanel`, which hosts the existing `ArrayEditorModal.svelte` for add/edit/reorder/delete of paths.
- Hover on any `.breadcrumb-row` → native `title="{fullPath}"` tooltip (story -02 contract).
- Hover on any positional label → `title="output: file[i]"` so users learn the symbolic name to wire downstream.
- Drag the node: unchanged (xyflow default).
- Keyboard: node is a single focus target (no per-row focus).
- Empty state (zero paths configured): render the `.path-list` with a single ghost row reading `—  (no files)` in `var(--text-muted)` at `opacity: 0.5` so the node never appears broken.
- Path-list max height: `var(--sp-2xl) * 3` = 96px with `overflow-y: auto` so a node with 10 paths does not swallow the canvas. Scrollbar uses global `::-webkit-scrollbar` styling (6px, subtle).

### Component specs
- Node `max-width: 240px` (20px wider than standard `.process-node` to accommodate `file[NN]` label + ellipsized path on the same inferred row width — still tight).
- Node `min-width: 180px`.
- `.path-list` — `display: flex; flex-direction: column; gap: 0; margin-top: var(--sp-2xs); max-height: 96px; overflow-y: auto`.
- `.breadcrumb-item` pair — positional label row `min-height: 12px`; breadcrumb row `min-height: 12px`; inter-pair gap `2px`.
- Divider: `margin: var(--sp-xs) calc(-1 * 10px) 0; padding-top: var(--sp-xs); border-top: 1px solid var(--border-subtle)` — negative horizontal margin extends the rule edge-to-edge of `.node-body` for visual weight; padding-top reintroduces internal breathing room for the footer.
- `.node-footer` — `display: flex; justify-content: space-between; align-items: center` — count badge on the left, status on the right.
- `ArrayEditorModal` integration: pass entries as `{path: string, alias?: string}` objects; when `alias` set, breadcrumb row's positional label renders the alias (e.g. `config:`); when unset, renders `file[i]` with `i` zero-based.

### Signature elements
- The **divider between path list and footer** is the single structural move that distinguishes MultiFileLoader from every other node on the canvas — users recognize it by shape alone, even at 40% zoom.
- **Stacked pairs inside a scrollable viewport** — the node is the only one with internal scroll, reinforcing "this is a list-valued source".
- **Positional `file[i]` symbols** are the contract to downstream nodes: they read as array-indexed, making the wiring grammar obvious without documentation.
- Restraint again: no count pill with color, no "multi-" badge graphic — a single divider line and a right-aligned `N files` count carry the identity.
