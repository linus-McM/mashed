# Story breadcrumbs-06: Downstream File Loader empty-only auto-fill

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** breadcrumbs-05
**Status:** ready

## Description

Phase 2 Tasks 2.2 + 2.3 (plan lines 87-96). Frontend listens for the `bmad:node:artifacts` event emitted by breadcrumbs-05. For each resolved output path, walks outgoing edges from the source node. For every target whose input path for the matching artifact name is **empty**, copy the resolved path. Never clobber a user-edited path. The Svelte side is the right home for this because it already has the full xyflow graph.

## Developer Notes

### Architecture

Existing event subscription: `frontend/src/components/bmad/CanvasPane.svelte` corpus line 28848:

```
const cancelArtifactListener = EventsOn('bmad:node:artifacts', (event) => { ... })
```

This listener must be extended to:

1. Read `event.Paths` (new field, added in breadcrumbs-05).
2. For each `(artifactName, absPath)`, iterate `edges` where `source === event.NodeID`.
3. For each such edge, look up the target node and inspect `target.data.config.inputPaths?.[matchingName]`.
4. Determine "matching name" from the downstream process's `processDef.inputs` — if any input name equals `artifactName`, that's the slot.
5. **Empty-only fill** (plan line 92): if the target slot is empty/undefined, write `absPath`. Otherwise skip.

Graph traversal stays on the Svelte side (plan line 94 — "Graph traversal for auto-fill runs on the Svelte side"). No new Go binding needed.

### Technical Considerations

- **Reactivity**: xyflow's node store must be updated via its setter so node re-renders pick up the new `inputPaths[name]` value. Use the existing node-update pattern in `CanvasPane.svelte`.
- **Idempotency**: receiving the same event twice must be a no-op (same path writes same value).
- **Persist to backend**: after auto-fill, the frontend calls `SaveBmadWorkflow(workflow)` so the path survives a reload. Do this on a short debounce (e.g. 500ms) to avoid flapping during burst events.
- **Undo visibility**: when auto-fill happens, the breadcrumb row flashes briefly (CSS animation class for ~600ms) so the user sees the change. Cosmetic but pays for itself — ambient auto-fill is otherwise invisible.

### Risks & Edge Cases

- **User-edit race**: user types a path at the exact moment auto-fill runs. Rule is still "empty-only fill" — the moment the field is non-empty, auto-fill is blocked. Simple check against current `inputPaths[name]`; no transactional fencing needed.
- **Multiple downstream targets accept the same artifact**: this is the ambiguity case explicitly deferred to Phase 3 / Ollama (plan lines 138-144). For this story: fill ALL matching empties. Document that this fan-out may become under-Ollama-control later.
- **Edge has `targetHandle` set**: respect it — if handle names indicate a specific artifact slot, match by handle rather than by input list position. This matters for MultiFileLoader in breadcrumbs-08.
- **Target has no processDef (e.g. MultiFileLoader)**: breadcrumbs-08 extends this story's logic; this story's contract is "process-to-process" only.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 87-103.
- Corpus: `frontend/src/components/bmad/CanvasPane.svelte` (line 20278; event handler at 28848).

## Acceptance Criteria

AC-1: Empty downstream slot fills on upstream complete
- Given node A (upstream) with output `PRD.md` connected to node B (downstream) with input `PRD.md`
- And `B.data.config.inputPaths` has no `PRD.md` key (or empty string)
- When A completes and `bmad:node:artifacts` fires with `Paths = {"PRD.md": "/abs/path.md"}`
- Then `B.data.config.inputPaths["PRD.md"]` becomes `"/abs/path.md"`
- And the canvas re-renders the IN breadcrumb on B showing `.../path.md`

AC-2: Non-empty downstream slot is NOT clobbered
- Given B already has `inputPaths["PRD.md"] = "/user/custom.md"`
- When A completes and emits a new path `/auto/PRD.md`
- Then `B.data.config.inputPaths["PRD.md"]` remains `/user/custom.md`
- And a console warn (or notification) records the skipped auto-fill

AC-3: Unmapped artifact produces no auto-fill
- Given A's outputs include `"code"` which is never in `event.Paths`
- When the event fires
- Then no downstream node is touched
- And no `SaveBmadWorkflow` call is made

AC-4: Persistence
- Given auto-fill updates node B
- When the debounce fires
- Then `SaveBmadWorkflow(workflow)` is called exactly once with B's new `inputPaths`

AC-5: Idempotent repeat
- Given the same `bmad:node:artifacts` event fires twice in a row
- When the second event fires
- Then no additional `SaveBmadWorkflow` call is made (path already equals target)

## BDD Test Scenarios

```gherkin
Feature: Downstream auto-fill

  Scenario: Empty slot filled
    Given node A (out: ["PRD.md"]) connected to node B (in: ["PRD.md"])
    And B.inputPaths has no PRD.md key
    When bmad:node:artifacts fires for A with Paths = {"PRD.md": "/a/PRD.md"}
    Then B.inputPaths["PRD.md"] equals "/a/PRD.md"

  Scenario: User-edited slot preserved
    Given B.inputPaths["PRD.md"] = "/user/custom.md"
    When bmad:node:artifacts fires for A with Paths = {"PRD.md": "/auto/PRD.md"}
    Then B.inputPaths["PRD.md"] still equals "/user/custom.md"

  Scenario: Persistence debounce
    Given three rapid bmad:node:artifacts events for A within 200ms
    When auto-fill processes them
    Then SaveBmadWorkflow is called exactly once after the 500ms debounce

  Scenario: Fan-out (documented deferral)
    Given A (out: ["PRD.md"]) connected to both B and C which both accept "PRD.md"
    And both B.inputPaths["PRD.md"] and C.inputPaths["PRD.md"] are empty
    When the event fires
    Then both B and C inputPaths["PRD.md"] are filled
    And a code comment references Phase 3 / Ollama for future disambiguation
```

## Tasks / Subtasks

- [ ] Task 1: Extend CanvasPane listener (AC-1, AC-3)
  - [ ] Parse `event.Paths` in the existing `EventsOn('bmad:node:artifacts', ...)` handler
  - [ ] For each path, walk outgoing edges; identify target nodes
- [ ] Task 2: Empty-only fill rule (AC-1, AC-2)
  - [ ] Helper `shouldFill(target, artifactName): boolean` — true iff `target.data.config.inputPaths?.[artifactName]` is `''` or undefined
  - [ ] Unit tests: empty, undefined, populated, whitespace-only (still counts as populated)
- [ ] Task 3: Update node + persist (AC-1, AC-4)
  - [ ] Use the xyflow node-update API already in CanvasPane
  - [ ] Debounced `SaveBmadWorkflow` call (500ms)
- [ ] Task 4: Idempotency check (AC-5)
  - [ ] If the new value equals the current value, skip both the update and the save
- [ ] Task 5: Playwright AC (AC-1, AC-2)
  - [ ] Fixture workflow with A→B; mock emitting the event; assert B's breadcrumb updates / preserves per case
- [ ] Task 6: Fan-out comment (AC per BDD scenario 4)
  - [ ] Inline code comment in the handler citing plan lines 138-144

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
No new DOM. Auto-fill modifies `node.data.config.inputPath` (or per-input map) on a downstream node via the `bmad:node:artifacts` event handler in `CanvasPane.svelte`. The existing `.breadcrumb-row` (story -02) reactively re-renders when the store updates. Layout is unchanged — the row that previously showed `—` now shows `.../<basename>`. No extra badge, no "auto-filled" indicator — the plan's empty-only rule (line 92) mandates a single source of truth: the resolved path. Visual indistinguishability between auto-filled and user-filled is the correctness guarantee.

### Typography plan
Inherits `.breadcrumb-row` from story -02: `9px var(--font-mono)`, weight `400`, `line-height: 1.3`, `font-variant-numeric: tabular-nums`. No typographic differentiation between auto-filled vs. user-filled values — that would imply a difference in trust, which is exactly what the empty-only rule prevents.

### Color strategy
- Pre-fill (em-dash): `var(--text-muted)` at `opacity: 0.7`.
- Post-fill (resolved path): `var(--text-muted)` at `opacity: 1`.
- Transition: `transition: color var(--duration-short) var(--ease-enter), opacity var(--duration-short) var(--ease-enter)` — a 100ms color/opacity crossfade that signals "state changed" without being animated enough to feel novelty-driven.
- No green/blue tint flash. No "just updated" highlight. If the user is looking at the node, they will see the em-dash become a path; if they are not, re-looking later should read identically to a manually entered value. This is the invariant.

### Interaction model
- Zero UI. The event is consumed, the store updates, the row re-renders.
- User can at any time click the node → open `NodeConfigPanel` → edit the path. The edit persists; subsequent upstream completions will **not** overwrite it (empty-only rule enforced in the handler).
- If the user **clears** a previously auto-filled path back to empty, that constitutes a user edit — the path stays empty on the next upstream completion. The handler reads "empty" from the current store; it does not track provenance. Keep it dumb.
- Hover still shows `title="{fullPath}"` post-fill.
- Accessibility: because the em-dash → path transition is silent, add `aria-live="polite"` to the breadcrumb row only when the node is the currently-selected node, so screen readers on focused nodes announce updates without canvas-wide spam.

### Component specs
- No changes to `.breadcrumb-row` geometry — story -02 specs stand.
- Add (scoped) `transition: color var(--duration-short) var(--ease-enter), opacity var(--duration-short) var(--ease-enter)` to the path text span.
- No layout shift when `—` (1 glyph) swaps for `.../basename` (N glyphs) — the parent row's `min-height: 12px` + `max-width: 100%` + `text-overflow: ellipsis` absorb the change.
- The 220px node max-width cap holds; newly auto-filled long paths ellipsis-clip identically to user-entered long paths.

### Signature elements
- **Silent correctness**: no toast, no flash, no check-mark ghost. An auto-fill looks exactly like a user fill because the downstream node does not care where the path came from — only that it resolves.
- This invisibility is the Mashed signature move — competitors add a "synced" pill; we add a 100ms color transition. Trust in the system is expressed through restraint.
- The em-dash `—` is the load-bearing glyph: it is the only state that visually telegraphs "nothing has been wired here yet". Its replacement, without fanfare, is the event.
