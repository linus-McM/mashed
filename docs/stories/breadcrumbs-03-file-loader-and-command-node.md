# Story breadcrumbs-03: Breadcrumb row on File Loader / CommandNode

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** breadcrumbs-02
**Status:** ready

## Description

Phase 1 Task 1.2 (plan lines 49-51). Apply the same breadcrumb treatment from breadcrumbs-02 to every remaining node component. If discovery in breadcrumbs-01 confirmed File Loader renders via `ProcessNode.svelte`, the surface for this story narrows to `CommandNode.svelte` and any other renderers surfaced by discovery — or this story degrades to a no-op verification pass.

## Developer Notes

### Architecture

Reads the discovery output from breadcrumbs-01. Two branches:

1. If File Loader has a dedicated component: add breadcrumb rows per Task 1.1 pattern.
2. If File Loader piggybacks on `ProcessNode.svelte` and breadcrumbs-02 already covered it: scope collapses to `CommandNode.svelte` (if it exists) only. Per plan line 50, "File Loader's output path is its loaded file" — the breadcrumb confirms what was selected in the browse dialog.

`CommandNode.svelte` was not found in the current corpus grep. If absent, the dev MUST:
- Document absence in the story's report-back
- Drop Task 2 (and drop acceptance AC-2 accordingly)
- Close this story with AC-1 + AC-3 only

### Technical Considerations

- Reuse `getNodePath` + `formatBreadcrumb` from breadcrumbs-01/02 — do NOT re-implement.
- Keep the same styling class (`.breadcrumb-row`) for visual consistency across node types.

### Risks & Edge Cases

- Drift: if future node types arrive without breadcrumb rows, the visual contract breaks. Mitigation: add a lint-style Playwright smoke test that asserts every rendered `.process-node`, `.command-node`, `.file-loader-node` (whichever classes exist) has at least one `.breadcrumb-row`.
- A `file-path` symbolic artifact (see `internal/bmad/artifacts.go` — mapped to empty string, plan data) means the File Loader emits unmapped content; breadcrumb reads the user-configured path from `outputPath`, NOT from `ResolveArtifactPath`.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 49-51 (Task 1.2).
- Corpus: `frontend/src/components/bmad/ProcessNode.svelte`, any File Loader / CommandNode component surfaced by breadcrumbs-01.
- Skill: `/playwright-cli` for visual AC.

## Acceptance Criteria

AC-1: All confirmed node renderers show breadcrumb rows
- Given the set of node components surfaced by breadcrumbs-01 discovery
- When each node type renders on the canvas
- Then each renders the same `.breadcrumb-row` markup + styling as `ProcessNode.svelte`

AC-2: File Loader breadcrumb reflects the browse-dialog selection
- Given a File Loader node whose `outputPath` was set via the browse dialog to `/Users/x/notes/brief.md`
- When the node renders on the canvas
- Then the OUT breadcrumb shows `.../brief.md` with full path in the `title` attribute

AC-3: No visual regression on ProcessNode
- Given the existing ProcessNode breadcrumbs from breadcrumbs-02
- When this story lands
- Then those breadcrumbs render identically (snapshot test or Playwright assertion)

## BDD Test Scenarios

```gherkin
Feature: Breadcrumb row parity across node components

  Scenario: File Loader shows its loaded file
    Given a File Loader node with outputPath "/Users/x/notes/brief.md"
    When the canvas renders the node
    Then the OUT breadcrumb shows ".../brief.md"

  Scenario: File Loader unset
    Given a File Loader node with empty outputPath
    When the canvas renders the node
    Then the OUT breadcrumb shows "—"

  Scenario: CommandNode component absent — no-op
    Given breadcrumbs-01 discovery reported CommandNode.svelte does not exist
    When this story executes
    Then no changes are made to CommandNode.svelte
    And the dev logs the decision in the story close-out
```

## Tasks / Subtasks

- [ ] Task 1: Read breadcrumbs-01 discovery findings (AC-1)
  - [ ] Enumerate target components for this story
  - [ ] Record decision (component vs no-op) at top of PR description
- [ ] Task 2 (conditional): Add breadcrumb row to File Loader component if dedicated (AC-1, AC-2)
  - [ ] Insert markup identical to ProcessNode.svelte pattern
  - [ ] Reuse `.breadcrumb-row` CSS
- [ ] Task 3 (conditional): Add breadcrumb row to CommandNode.svelte if it exists (AC-1)
  - [ ] Same pattern as Task 2
- [ ] Task 4: Playwright smoke test (AC-1, AC-3)
  - [ ] Assert every node rendered in a multi-node workflow has `.breadcrumb-row`

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
File Loader is **not** a distinct component — it is `ProcessNode.svelte` rendered with `data.processId === 'util-file-loader'` (confirmed via `NodeConfigPanel.svelte:25831`). Apply the same `.breadcrumb-row` pattern introduced in story -02. For File Loader specifically: the `in:` symbolic row is typically empty (it is the source), so only the `out:` breadcrumb row renders; it displays `.../<basename>` of the user-selected file. If a CommandNode component exists or gets added, mirror the exact DOM fragment — stacked below each `.artifacts` row. No new CSS class — reuse `.breadcrumb-row` defined in story -02; this story adds **no** new styling, only ensures the same shared row renders in every node variant.

### Typography plan
Identical to story -02 to preserve uniformity: `var(--font-mono)`, `9px`, weight `400`, `line-height: 1.3`, `font-variant-numeric: tabular-nums`. No per-node typographic variation — that is a promise to the user that every node speaks the same visual language.

### Color strategy
- Resolved path: `color: var(--text-muted)`.
- Unresolved em-dash: `color: var(--text-muted); opacity: 0.7`.
- File Loader's breadcrumb, once a file is selected via the browse dialog, will be the **only** content that distinguishes two File Loader nodes on a canvas — its prominence matters. Do not add a second color; retain `--text-muted` so the node card's visual weight stays in `.node-label`.
- No status-color tinting (don't recolor on complete/failed) — the status row owns color semantics.

### Interaction model
- `title="{fullPath}"` on `.breadcrumb-row` (identical contract to story -02).
- No click — browse dialog is launched from `NodeConfigPanel.svelte` via the existing gear/config affordance, not the breadcrumb itself.
- State transition when browse dialog returns a path: `transition: color var(--duration-short) var(--ease-enter)` on the text. The em-dash `—` swaps to `.../<basename>` in place; no layout shift, no flash.
- Keyboard: row is not focusable; node card focus remains the single tab stop.

### Component specs
Identical to `.breadcrumb-row` in story -02 — this story is about **application**, not specification. Specifically confirm:
- `padding-left: calc(var(--sp-sm) + 4px)` aligns under `.artifact-list` of the preceding row.
- `max-width: 100%` inherits the node's 220px cap.
- `min-height: 12px` reserves space so File Loader's card doesn't grow vertically when a path resolves.
- For CommandNode (if extended): same row rendered once per declared input + output.

### Signature elements
- The "single breadcrumb pattern, every node" rule is itself the signature — users learn the glyph (`.../filename` dim, `—` dimmer) once and read every node at a glance.
- File Loader becomes self-documenting: the node literally shows which file it loads, no need to open the config panel to check.
- Consistency is a visual product — a CommandNode that styled its breadcrumb differently would immediately feel like a leak. This story's job is to prevent that.
