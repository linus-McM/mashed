# Story breadcrumbs-02: Breadcrumb row on ProcessNode

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** breadcrumbs-01
**Status:** ready

## Description

Phase 1 Task 1.1 (plan lines 36-48). Add a dim, smaller breadcrumb row beneath each existing symbolic `IN:` / `OUT:` label inside `ProcessNode.svelte`. This gives every process node on the canvas two layers of truth: the symbolic artifact name (the contract) and the resolved path basename (the reality).

## Developer Notes

### Architecture

Target file: `frontend/src/components/bmad/ProcessNode.svelte` (corpus line 20011). Existing markup already renders the symbolic `in:` / `out:` rows with `.artifact-label` and `.artifact-list` styles (referenced in plan line 47 — "reuse typography fixed this session").

Use the `getNodePath(node, dir)` helper from breadcrumbs-01. Format function:

```ts
function formatBreadcrumb(absPath: string): string {
  if (!absPath) return '—';
  const base = absPath.split('/').pop() ?? absPath;
  return `.../${base}`;
}
```

Set `title={absPath}` so hovering reveals the full path.

### Technical Considerations

- Styling: 9–10px monospace, `color: var(--text-muted)` (plan line 47).
- A single row per symbolic input/output (multi-input is breadcrumbs-04).
- When `absPath` is empty render the em-dash glyph `—` (U+2014), still dim. Do NOT render an empty div.

### Risks & Edge Cases

- A Windows-style path would break `split('/')` — mashed is macOS/Linux only per spec; treat as not-applicable and document in the helper.
- Very long basenames: rely on `text-overflow: ellipsis` + `white-space: nowrap` on the row.
- A node mid-execution where backend hasn't populated paths yet must render `—`, never stale data.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 36-62 (Task 1.1 + acceptance).
- Corpus: `frontend/src/components/bmad/ProcessNode.svelte`, `frontend/src/lib/bmad/nodePath.ts` (created in breadcrumbs-01).
- Skill: `/playwright-cli` for AC validation of rendered breadcrumbs.

## Acceptance Criteria

AC-1: Breadcrumb row renders for every symbolic in/out on a process node
- Given a process node with symbolic input `brainstorm-notes` and output `prd`
- When the node mounts on the canvas with no resolved paths set
- Then two breadcrumb rows render beneath the symbolic rows, each showing `—`

AC-2: Resolved path renders as `.../<basename>`
- Given `node.data.config.outputPath` is `/Users/x/_bmad-output/planning-artifacts/PRD.md`
- When the node renders
- Then the OUT breadcrumb shows `.../PRD.md`
- And the element's `title` attribute equals the full absolute path

AC-3: Styling matches design tokens
- Given the breadcrumb row renders
- When a tester inspects the computed style
- Then `font-size` is 9–10px, `font-family` is monospace, `color` equals `var(--text-muted)`

AC-4: No regression to symbolic rows
- Given the existing symbolic `IN:` and `OUT:` rows
- When this story lands
- Then those rows still render with their existing labels and styles

## BDD Test Scenarios

```gherkin
Feature: ProcessNode breadcrumb row

  Scenario: Unresolved path renders em-dash
    Given a process node with inputs ["brainstorm-notes"] and empty config
    When the canvas renders the node
    Then the IN breadcrumb row is visible with text "—"

  Scenario: Resolved path renders basename
    Given a process node whose data.config.outputPath is "/abs/_bmad-output/planning-artifacts/PRD.md"
    When the canvas renders the node
    Then the OUT breadcrumb row shows ".../PRD.md"
    And the row's title attribute equals "/abs/_bmad-output/planning-artifacts/PRD.md"

  Scenario: Symbolic rows unchanged
    Given a process node with symbolic inputs ["brainstorm-notes"] and outputs ["prd"]
    When the canvas renders the node
    Then the IN symbolic row still shows "brainstorm-notes"
    And the OUT symbolic row still shows "prd"
```

## Tasks / Subtasks

- [ ] Task 1: Breadcrumb formatter utility (AC-1, AC-2)
  - [ ] Add `formatBreadcrumb(absPath: string): string` in `frontend/src/lib/bmad/nodePath.ts`
  - [ ] Unit tests: empty → `—`; populated → `.../<basename>`
- [ ] Task 2: Markup in ProcessNode.svelte (AC-1, AC-2, AC-4)
  - [ ] Insert breadcrumb `<div>` directly beneath each symbolic `.artifact-list` row
  - [ ] Bind `title={getNodePath(node, dir)}` and text content from `formatBreadcrumb(...)`
  - [ ] Preserve existing symbolic row markup verbatim
- [ ] Task 3: Styling (AC-3)
  - [ ] Add `.breadcrumb-row` class with 9–10px mono, `var(--text-muted)`, ellipsis overflow
- [ ] Task 4: Playwright AC validation (AC-1, AC-2, AC-4)
  - [ ] Add spec under `tests/ac/` that mounts a workflow with one process node and asserts both breadcrumb rows + title attribute

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
Extend `.node-body` in `ProcessNode.svelte` (`frontend/src/components/bmad/ProcessNode.svelte`). Under each existing `.artifacts` row (the symbolic `in:` / `out:` row), insert a sibling `.breadcrumb-row`. Row order within the node: `.node-header` → for each input: `.artifacts` (symbolic) + `.breadcrumb-row` (resolved path) → for each output: `.artifacts` (symbolic) + `.breadcrumb-row` → `.status-row`. The breadcrumb is vertically paired with the symbolic row above; no label column — the preceding `in:`/`out:` owns semantic framing. Indent breadcrumb content by `calc(var(--sp-sm) + 4px)` from the left edge of `.node-body` so it visually hangs under the `artifact-list`, not the `artifact-label`.

### Typography plan
- Font: `var(--font-mono)` (inherited from `.process-node`).
- Size: `9px` (matches existing `.artifacts` scale on this node; stays under `--text-label`/11px).
- Weight: `400` — lighter than the `600` of `.artifact-label` so the breadcrumb recedes.
- Line-height: `1.3` (mirrors `.artifacts`).
- `font-variant-numeric: tabular-nums` on the path so `[0]`/`[1]` indices align in stacked rows.

### Color strategy
- Path text: `color: var(--text-muted)` — explicitly dimmer than `.artifact-list` (which inherits `--text-muted` already) but **dimmer than** the `--text-dim` used for `.artifact-label`. This creates a three-tier hierarchy: label (`--text-dim`) > symbolic (`--text-muted`) > breadcrumb (`--text-muted`).
- Em-dash (unresolved): `color: var(--text-muted)` with `opacity: 0.7` — same token, weaker presence.
- No background, no border — the row is text-only inside the existing `.node-body` surface (`var(--bg-elevated)`).

### Interaction model
- No click. No pointer cursor. `user-select: text` so users can copy the basename.
- Hover on `.breadcrumb-row` exposes `title="{fullPath}"` — the browser native tooltip carries the absolute path. Unresolved rows set `title="unresolved"` to disambiguate from "no tooltip yet".
- Keyboard: the row is not a focus target (node itself remains the focusable unit). No `tabindex`.
- State transition (unresolved → resolved): `transition: color var(--duration-short) var(--ease-enter)` on the text. No fade-in of the element — it was always present as `—`, only its content changes.

### Component specs
- `.breadcrumb-row` — `display: flex`; `gap: var(--sp-xs)`; `padding: 0 0 0 calc(var(--sp-sm) + 4px)`; `margin-bottom: var(--sp-2xs)`; `min-height: 12px` (reserves vertical space so the node doesn't jitter when resolved text arrives).
- Overflow: `overflow: hidden`; `white-space: nowrap`; `text-overflow: ellipsis`; `max-width: 100%` (constrained by parent `.process-node` max-width of 220px).
- The node's existing `max-width: 220px` is the ellipsis boundary. Do **not** widen it — density is the point.

### Signature elements
- The `.../<basename>` prefix is the distinctive move: three-char `.../ ` in `--text-muted` then the filename leans slightly on the visual eye. This is the "Linear-meets-terminal" mark — a path whisper under the symbolic name.
- Em-dash `—` (U+2014, not hyphen, not three dots) as empty-state glyph — a typographic tell that reads as "intentionally empty" rather than "data missing".
- Stacked symbolic + breadcrumb pairing (not side-by-side) preserves the existing `.artifacts` row rhythm; the node grows vertically in 12px increments per artifact.
