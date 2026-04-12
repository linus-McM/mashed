# skills-validate-01: Frontmatter validation badges on sidebar asset rows

**Status:** ready
**Domain:** fullstack
**Size:** M
**Depends on:** none (Phase 1 is shipped; pairs well with skills-watch-02 but not dependent)
**Phase:** 4c

## Description

Surface non-blocking validation warnings inline on each mashed-asset row in the sidebar. Warnings include:

- `description` missing, too short (< 10 chars), or too long (> 200 chars)
- `mashedInputs` / `mashedOutputs` referencing paths that don't exist in the current repo
- `mashedCompletion` value not in the recognised set
- `mashedRole` is `command` but `mashedChainable` is `none` (suspicious — commands usually chain)

Validation is advisory only — assets with warnings remain draggable and runnable. Warnings render as a small badge/icon next to the asset name with a hover tooltip listing the specific issues.

**This is a UI-facing story — route through ui-architect for badge placement and tooltip styling.**

## Developer Notes

- **Files to create/modify:**
  - `internal/bmad/assets.go` (or new `assets_validate.go`):
    - Add `ValidateMashedAsset(asset MashedAssetInfo, repoPath string) []ValidationIssue`
    - `ValidationIssue` struct with `Field string`, `Severity string` (`info | warn`), `Message string`
  - `internal/bmad/types.go` — extend `MashedAssetInfo` with a new `Issues []ValidationIssue` field (populated at load time)
  - `internal/bmad/assets.go:LoadMashedAssetsFromDir` — call `ValidateMashedAsset` after parsing each asset and attach issues
  - `frontend/src/components/bmad/ProcessSidebar.svelte` — render a warning badge next to rows whose `issues.length > 0`
  - `frontend/src/components/bmad/ValidationBadge.svelte` — NEW. Small icon + tooltip listing issues
  - `internal/bmad/assets_validate_test.go` — NEW. Table-driven validation tests
- **Validation rules (v1 — keep simple; expand later):**

| Field | Rule | Severity |
|---|---|---|
| `description` | missing or empty | `warn` |
| `description` | length < 10 | `info` |
| `description` | length > 200 | `info` |
| `mashedInputs` entry | referenced path does not exist (glob match against `repoPath` produces zero files) | `warn` |
| `mashedOutputs` entry | same | `info` (outputs may be produced later) |
| `mashedCompletion` | value not in {`idle`, `exit`, starts with `marker:`, starts with `timeout:`} | `warn` |
| `mashedRole: command` + `mashedChainable: none` | inconsistent | `info` |

- **Frontmatter field parity:** make sure `MashedAssetInfo` already exposes every field the validator reads. If `mashedInputs`/`mashedOutputs` aren't yet on the Go struct, add them as `[]string` slices.
- **Risks / gotchas:**
  - **Validation must NEVER block drag or run.** Phase 1's tolerant parser invariant (plan §"Parser rules") must hold: one bad field does not exclude the asset. Validation runs AFTER parsing and only attaches metadata.
  - **Performance:** `ValidateMashedAsset` is called once per asset per `ListAllMashedAssets` call. Avoid stat'ing the filesystem for every input on every call — the repo-path input existence check should be bounded (O(inputs × single stat)), not recursive.
  - **Empty `repoPath`**: the input-existence check must skip silently when `repoPath == ""` (e.g. no repo selected). Return an empty issue list for that check.
  - **Glob expansion**: `mashedInputs` may contain globs like `**/*.go`. Use `filepath.Glob` for simple cases; for `**` recursive globs use `doublestar` if already in go.mod, otherwise document the limitation as "no ** support in v1".
  - **Tooltip accessibility:** the badge must be keyboard-focusable and announce issues via `aria-describedby`.
  - Design-system tokens only (Do-Not-Repeat 2026-04-10).
- **Prerequisites already in place:**
  - Phase 1 parser tolerant to malformed fields.
  - `MashedAssetInfo` struct and `ListAllMashedAssets` binding.
  - Sidebar row iteration in `ProcessSidebar.svelte`.

## Acceptance Criteria

**AC-1: Missing description produces a `warn` issue**
- Given an asset whose frontmatter lacks a `description` field (or it's empty after body fallback)
- When `ValidateMashedAsset` runs
- Then the returned slice contains one issue with `Field: "description"`, `Severity: "warn"`, `Message` mentioning "missing"

**AC-2: Nonexistent input path produces a `warn` issue**
- Given an asset with `mashedInputs: ["docs/nonexistent.md"]` and a valid `repoPath` where that file does not exist
- When `ValidateMashedAsset` runs with the `repoPath`
- Then the returned slice contains an issue with `Field: "mashedInputs"`, `Severity: "warn"`, `Message` referencing the missing path

**AC-3: Empty `repoPath` skips path-existence checks**
- Given an asset with `mashedInputs: ["docs/any.md"]` and `repoPath: ""`
- When `ValidateMashedAsset` runs
- Then no path-existence issue is returned

**AC-4: Sidebar renders the badge when issues exist**
- Given an asset with at least one issue
- When the sidebar renders
- Then a `ValidationBadge` component appears next to the asset name
- And hovering/focusing the badge shows a tooltip listing each issue's message

**AC-5: Asset with zero issues renders without a badge**
- Given an asset with a valid description, valid completion, no inputs/outputs
- When the sidebar renders
- Then no `ValidationBadge` appears on that row

**AC-6: Validation does not block drag or run**
- Given an asset with multiple `warn` issues
- When the user drags it onto the canvas
- Then the drag is accepted and the node is created normally
- And attempting to run the workflow proceeds without a validation gate

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Frontmatter validation badges

  Scenario: Missing description warns
    Given an asset without a description
    When validation runs
    Then the issues list contains a "description missing" warn

  Scenario: Nonexistent input path warns
    Given an asset with mashedInputs referencing a path not in the repo
    When validation runs with a valid repoPath
    Then the issues list contains a "path does not exist" warn

  Scenario: Empty repoPath skips path checks
    Given an asset with mashedInputs and an empty repoPath
    When validation runs
    Then no path-existence issue is returned

  Scenario: Sidebar shows badge for asset with issues
    Given an asset with one warn issue
    When the sidebar renders
    Then a ValidationBadge is visible next to the asset name
    And focusing the badge shows a tooltip with the issue message

  Scenario: Clean asset has no badge
    Given a fully valid asset
    When the sidebar renders
    Then no badge appears

  Scenario: Warnings do not block drag
    Given an asset with warnings
    When the user drags it onto the canvas
    Then the node is created normally
```

## Tasks / Subtasks

- [ ] Task 1 — `ValidationIssue` type + `ValidateMashedAsset` (AC-1, AC-2, AC-3)
  - [ ] Define the struct in `internal/bmad/types.go`
  - [ ] Implement the validator with each rule from the table above
  - [ ] Skip path checks cleanly when `repoPath == ""`
- [ ] Task 2 — Wire into loader (AC-1 through AC-3)
  - [ ] Call `ValidateMashedAsset` in `LoadMashedAssetsFromDir`
  - [ ] Attach issues to `MashedAssetInfo.Issues`
  - [ ] Ensure `ListAllMashedAssets` returns issues in the payload
- [ ] Task 3 — `ValidationBadge.svelte` (AC-4, AC-5)
  - [ ] Icon + tooltip + keyboard focus
  - [ ] Design-system tokens only
  - [ ] aria-describedby for issue list
- [ ] Task 4 — Sidebar integration (AC-4, AC-5, AC-6)
  - [ ] Render badge only when `issues.length > 0`
  - [ ] Ensure drag handlers still fire on the row itself (badge does not block events)
- [ ] Task 5 — Tests (AC-1 through AC-6)
  - [ ] Go table tests for each validation rule
  - [ ] Playwright: badge visibility, tooltip text, drag still works

## Definition of Done

- [ ] All ACs verified by an automated test
- [ ] Coverage ≥ 80% on modified files
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths, magic numbers, or color literals added
- [ ] Existing tests still pass
- [ ] Phase 1 parser tolerance invariant still holds (verified by running existing `assets_test.go` 25-test suite)

## Design Brief

### Layout composition
Validation badge is a 12px icon slotted inline in the existing `.process-item` row grammar. Existing row layout:
```
[ dot ] [ name ]                           [ module-badge? ]
```
After this story:
```
[ dot ] [ name ] [ badge? ]                [ module-badge? ]
```

Badge lives at `margin-left: var(--sp-xs)` after the name, BEFORE any `flex: 1` spacer that pushes `module-badge` right. The existing row uses `gap: 8px` (which equals `var(--sp-sm)`) — the badge does not add a new gap; it steals `var(--sp-xs)` of tight spacing from the name column.

Tooltip is a floating surface:
- `position: absolute; z-index: 200;`
- Positioned by the badge's `getBoundingClientRect()` + small offset
- Anchored `right` of the badge when there is space, `left` when clipped
- `max-width: 280px` so long issue lists wrap

Tooltip internal layout: each issue is a flex row with icon + message, stacked vertically with `gap: var(--sp-xs)`.

### Typography plan
- **Badge icon**: no text — 12px Lucide `AlertTriangle` (warn) or `Info` (info). Size matches the existing `.cf-icon size={12}` precedent used in control flow rows.
- **Tooltip title** (optional — "Issues"): `font-family: var(--font-ui)`, `font-size: var(--text-label)` (11px), `font-weight: 600`, `text-transform: uppercase`, `letter-spacing: 0.05em`, `color: var(--text-dim)`. Matches the modal field-label pattern.
- **Tooltip issue message**: `font-family: var(--font-ui)`, `font-size: var(--text-label)` (11px), `font-weight: 400`, `color: var(--text-primary)`, `line-height: 1.4`.
- **Tooltip field prefix** (e.g. "description:"): `font-family: var(--font-mono)`, `font-size: 10px` (token missing; reuse the existing `.story-badge-id` 8px / `.artifacts` 9px precedent — use `var(--text-label)` 11px if no smaller token is justified), `color: var(--text-dim)`.

### Color strategy
- **Warn badge** (`Severity: warn`): icon `color: var(--accent-amber)` with `color-mix(in srgb, var(--accent-amber) 15%, transparent)` background circle. Uses the same alpha grammar as `.artifact-icon.missing` — consistency with ProcessNode.
- **Info badge** (`Severity: info`): icon `color: var(--accent-blue)` with `color-mix(in srgb, var(--accent-blue) 15%, transparent)` background circle.
- **Mixed severity** (asset has both warn and info): use warn (amber) — warn wins. Don't try to stack.
- **Tooltip surface**: `background: var(--bg-elevated)`, `border: 1px solid var(--border-emphasis)`, `box-shadow: 0 8px 24px color-mix(in srgb, var(--bg-deepest) 80%, transparent)`. Slightly lighter shadow than the modal because tooltips are secondary, not primary surfaces.
- **Tooltip issue bullet**: tiny 4px dot colored to match severity, `background: var(--accent-amber)` or `var(--accent-blue)`.
- **Row background**: UNCHANGED — validation badges must not tint the row. The design principle is "advisory, not blocking"; recoloring the row would read as "this item is broken, don't use it".

### Interaction model
- **Badge trigger** — hover OR keyboard focus opens the tooltip.
- **Keyboard**: the badge itself is a `<button>` with `tabindex="0"` so it lands in the Tab sequence. Enter/Space toggles the tooltip open state for touch/keyboard users.
- **Click behavior**: clicking the badge toggles the tooltip (persistent open until another click or Escape). This lets the user copy the issue text.
- **Hover**: mouseenter shows tooltip after 200ms delay (avoids jittery flicker as user moves across the row); mouseleave hides after 100ms.
- **Badge does NOT swallow row events**: the parent row's `draggable` and drag handlers still fire. The badge uses `pointer-events: auto` but the row uses event bubbling for drag — confirm in Playwright that dragging the row while over the badge still initiates drag.
- **aria-describedby**: badge has `aria-describedby="asset-{id}-issues"` pointing to the tooltip's id. Screen readers announce the full issue list on focus.
- **Escape**: closes the tooltip without moving focus.
- **Focus ring**: global `button:focus-visible` rule paints `2px solid var(--accent-green)` around the badge on keyboard focus — no override.

### Component specs
```
.validation-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  margin-left: var(--sp-xs);
  padding: 0;
  border: none;
  background: transparent;
  border-radius: 50%;
  cursor: pointer;
  flex-shrink: 0;
  transition: background var(--duration-short) var(--ease-enter);
}
.validation-badge.warn {
  background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
  color: var(--accent-amber);
}
.validation-badge.info {
  background: color-mix(in srgb, var(--accent-blue) 15%, transparent);
  color: var(--accent-blue);
}
.validation-badge:hover.warn {
  background: color-mix(in srgb, var(--accent-amber) 25%, transparent);
}
.validation-badge:hover.info {
  background: color-mix(in srgb, var(--accent-blue) 25%, transparent);
}

.validation-tooltip {
  position: absolute;
  z-index: 200;
  max-width: 280px;
  padding: var(--sp-sm) var(--sp-md);
  background: var(--bg-elevated);
  border: 1px solid var(--border-emphasis);
  border-radius: var(--radius-md);
  box-shadow: 0 8px 24px color-mix(in srgb, var(--bg-deepest) 80%, transparent);
  display: flex;
  flex-direction: column;
  gap: var(--sp-xs);
  font-family: var(--font-ui);
  font-size: var(--text-label);
  color: var(--text-primary);
  line-height: 1.4;
}
.validation-tooltip .issue {
  display: flex;
  align-items: flex-start;
  gap: var(--sp-xs);
}
.validation-tooltip .issue-dot {
  width: 4px; height: 4px;
  border-radius: 50%;
  margin-top: 6px;
  flex-shrink: 0;
}
.validation-tooltip .issue-dot.warn { background: var(--accent-amber); }
.validation-tooltip .issue-dot.info { background: var(--accent-blue); }
.validation-tooltip .field {
  font-family: var(--font-mono);
  color: var(--text-dim);
}
```
Tooltip shadow is lighter than modal shadow — no dedicated `--shadow-tooltip` token exists. Token missing — use the literal value.

### Signature elements
- **14px alpha-tinted circle** next to the name rather than a classic red/yellow exclamation triangle outside the row. The minimal footprint keeps the density bar high (the sidebar already carries counts, module badges, and chevrons).
- **Amber for warn, blue for info** — same two colors the app uses everywhere else for "attention" and "informational". The user never has to learn a new palette for validation.
- **Badge inherits the row's drag semantics** — you can still grab the row and drag it to canvas while hovering the badge, because validation is advisory. This physical affordance IS the design statement: "warnings never stop work".

### Distinguishing from existing patterns
- **vs `.module-badge`** (pinned label on the right of the row): module-badge is a TEXT pill that identifies module provenance; validation-badge is a COLORED CIRCLE ICON between name and module-badge that signals quality. They can coexist on the same row without clutter because one is word-shaped and one is dot-shaped.
- **vs `.artifact-icon`** inside ProcessNode cards (14px circles with `✓` or `!` characters): the validation badge uses a Lucide icon (AlertTriangle / Info), not a character glyph, and it opens a TOOLTIP on interaction. The ProcessNode artifact icons are static status markers; the sidebar validation badge is a progressive disclosure control. Same visual vocabulary (14px tinted circle), different role — the user reads them as "two flavors of the same grammar".
- **vs skills-watch-02 row flash** (transient color wash on the full row): validation is persistent metadata on a single inline element. Never combine them — a row being flashed green because it just saved, with an amber badge because the save left it invalid, is legible: the flash is about WHEN, the badge is about WHAT. They coexist without confusion.
- **Cross-story rule**: the 14px alpha-tinted circle IS the mashed "status chip" primitive. Any future per-item status (deprecated, beta, experimental) should use this component and a new `var(--accent-*)` color — never introduce a new visual form.
