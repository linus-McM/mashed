# Story breadcrumbs-04: Multi-input breadcrumb rendering

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** breadcrumbs-02
**Status:** ready

## Description

Phase 1 Task 1.3 (plan lines 53-55). When a node accepts multiple inputs (e.g. `["prd", "sprint-status"]`), render one breadcrumb row per symbolic name so a user can tell at a glance which upstream file each input is wired to. Labels must truncate with ellipsis; the full resolved path stays visible on hover via the `title` attribute.

## Developer Notes

### Architecture

The UI already loops over `node.processDef.inputs` / `.outputs` for symbolic rows (confirmed via corpus line 20011 component). Extend each loop to render a sibling breadcrumb row using a per-artifact-name path lookup, not a single node-wide `inputPath`.

This means the data model in breadcrumbs-01 (flat `inputPath` string) is insufficient for multi-input — we need a per-artifact map. Two paths forward:

1. **Use an object-shaped config value**: read `node.data.config.inputPaths?.[artifactName]` when present, fall back to legacy single `inputPath` when only one input exists. Mirror for `outputPaths`.
2. Defer proper structured shape to breadcrumbs-05 (which introduces `OutputPaths map[string]string` per architectural decision #1, plan line 190). This story writes the reader; the writer comes in 05.

Recommended: take path 1 now — add the map-valued reader, accept both forms (`inputPath`: singleton, `inputPaths`: keyed by artifact name). Breadcrumbs-05 then writes `inputPaths` / `outputPaths`. This preserves a narrow contract and avoids re-touching ProcessNode when Phase 2 lands.

### Technical Considerations

- Update `getNodePath(node, dir, artifactName?)` helper signature to optionally take the symbolic artifact name. When omitted, fall back to the single-value legacy key.
- Ellipsis truncation requires `.breadcrumb-row { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }`.
- `title` attribute must always be the absolute full path (not `.../basename`).

### Risks & Edge Cases

- Process has 0 inputs (pure source): render no IN breadcrumb rows at all. Loop body handles this naturally.
- Process has 1 input, only legacy `inputPath` set: helper returns the legacy value even when called with the artifact name.
- Process has N inputs but only some mapped: unmapped ones render `—`.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 53-62.
- Corpus: `frontend/src/components/bmad/ProcessNode.svelte`, `frontend/src/lib/bmad/nodePath.ts` (from breadcrumbs-01/02).

## Acceptance Criteria

AC-1: One breadcrumb row per symbolic input
- Given a process node with inputs `["prd", "sprint-status"]`
- When the node renders
- Then two IN breadcrumb rows appear, one per symbolic name, in the same order as `processDef.inputs`

AC-2: Per-artifact path lookup
- Given `node.data.config.inputPaths = { prd: "/a/PRD.md", "sprint-status": "/a/sprint.yaml" }`
- When the node renders
- Then the prd row shows `.../PRD.md` and the sprint-status row shows `.../sprint.yaml`

AC-3: Legacy singleton fallback
- Given a process node with one input and only `node.data.config.inputPath` set (no `inputPaths` map)
- When the node renders
- Then the breadcrumb correctly uses the legacy value

AC-4: Ellipsis + hover full path
- Given a basename longer than the node's visual width
- When the user hovers the breadcrumb
- Then the row displays with ellipsis and the `title` attribute contains the full absolute path

## BDD Test Scenarios

```gherkin
Feature: Multi-input breadcrumb rendering

  Scenario: Two inputs render two rows
    Given a process node with inputs ["prd", "sprint-status"]
    And data.config.inputPaths = {prd: "/a/PRD.md", "sprint-status": "/a/sprint.yaml"}
    When the node renders
    Then the IN breadcrumb list contains ".../PRD.md" and ".../sprint.yaml" in that order

  Scenario: Partial mapping
    Given a process node with inputs ["prd", "sprint-status"]
    And data.config.inputPaths = {prd: "/a/PRD.md"}
    When the node renders
    Then the prd breadcrumb shows ".../PRD.md"
    And the sprint-status breadcrumb shows "—"

  Scenario: Legacy fallback
    Given a process node with inputs ["brainstorm-notes"]
    And data.config.inputPath = "/a/brain.md" (no inputPaths map)
    When the node renders
    Then the brainstorm-notes breadcrumb shows ".../brain.md"
```

## Tasks / Subtasks

- [ ] Task 1: Extend helper signature (AC-2, AC-3)
  - [ ] `getNodePath(node, dir, artifactName?)` reads `inputPaths[artifactName]` first
  - [ ] Falls back to legacy `inputPath` when map missing or key absent AND only one input exists
  - [ ] Unit tests for multi-input, partial map, legacy fallback, missing key
- [ ] Task 2: Per-input loop in ProcessNode.svelte (AC-1)
  - [ ] Inside the existing `#each processDef.inputs as name` loop, render one `.breadcrumb-row` per name
  - [ ] Mirror for outputs
- [ ] Task 3: Ellipsis styling (AC-4)
  - [ ] Confirm `.breadcrumb-row` CSS includes overflow/ellipsis rules
- [ ] Task 4: Playwright AC (AC-1, AC-2, AC-4)
  - [ ] Fixture workflow with 2-input node; assert count + text + `title` attribute

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
Replace the single current `{#if inputs.length > 0}` block (ProcessNode.svelte:20060-20065, which renders `inputs.join(', ')` into one row) with an `{#each inputs as name}` loop. Each iteration emits a pair: `.artifacts` (symbolic name for this specific input) + `.breadcrumb-row` (resolved path for this input). Same treatment for outputs. Row order top-to-bottom: header → (per input: symbolic row, breadcrumb row) → (per output: symbolic row, breadcrumb row) → status row. Gap between pairs: none — they read as bonded. Gap between the last input pair and the first output pair: `margin-top: var(--sp-2xs)` on the first output `.artifacts` row to create a micro-separation without a divider line.

### Typography plan
- Symbolic row: unchanged — `9px var(--font-mono)`, weight `600` for the label, weight `400` for the name.
- Breadcrumb row: `9px var(--font-mono)`, weight `400`, `line-height: 1.3`, `font-variant-numeric: tabular-nums`.
- When multiple inputs stack, tabular-nums guarantees `[0]`/`[1]`/`[2]` indices (if present in basename) align vertically across rows — visual precision at 9px depends on this.

### Color strategy
- Per-input symbolic label: `var(--text-dim)` (unchanged from current `.artifact-label`).
- Per-input symbolic name: `var(--text-muted)` (unchanged `.artifact-list`).
- Per-input breadcrumb path: `var(--text-muted)`.
- Unresolved em-dash: `var(--text-muted)` with `opacity: 0.7`.
- No per-input accent coloring — a three-input node must stay calm. Color variance here would make the card read as "noisy".

### Interaction model
- `title="{fullPath}"` on each individual `.breadcrumb-row` — each input exposes its own path independently.
- Symbolic row gets `title="{symbolicName}"` only if the name is long enough to ellipsis-truncate (detectable via `scrollWidth > clientWidth` at render time — if expensive, always set it; redundant tooltips are cheap).
- No keyboard focus on individual rows; the node remains one focus target. No per-row click handlers.
- Stagger: when the node first mounts with N inputs, do **not** stagger. All rows appear simultaneously — the node is a single unit, not a list.

### Component specs
- Per-row `min-height: 12px` (symbolic) and `12px` (breadcrumb) — predictable vertical growth: each input consumes exactly `24px + var(--sp-2xs)` = `28px`.
- Node `max-width: 220px` remains the constraint; each row independently ellipsis-clips.
- Symbolic row: `display: flex; gap: var(--sp-xs)` — `flex-shrink: 0` on label, `overflow: hidden; text-overflow: ellipsis` on the name span.
- Breadcrumb row: `padding-left: calc(var(--sp-sm) + 4px)`; `max-width: 100%`; `overflow: hidden; text-overflow: ellipsis; white-space: nowrap`.
- A three-input node consumes roughly 84px more height than its single-input sibling — acceptable given the density payoff; canvas zoom compensates for outliers.

### Signature elements
- Pair-stacked rows (symbolic above, breadcrumb below) form a two-line "unit" per artifact — this is the distinct pattern competitors collapse into a single comma-joined line. Mashed preserves the vertical rhythm so the eye scans down symbolic names and down paths as two parallel columns.
- Per-input `title` tooltips make every row independently inspectable — the node becomes a tactile information object, not a summary glyph.
- The quiet micro-gap (`--sp-2xs`) between the last input and the first output is the only visual divider between in- and out-groups — no horizontal rule, no label change. Restraint is the signature.
