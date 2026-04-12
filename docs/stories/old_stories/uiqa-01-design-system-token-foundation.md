# Story uiqa-01: Design System Token Foundation

**Status:** done
**Size:** S
**Priority:** P1
**Domain:** frontend
**Depends on:** none

## Description

Add the four missing design tokens (`--accent-cyan`, `--accent-orange`, `--overlay-backdrop`, `--glow-spread`) to `frontend/src/style.css` and document them in `DESIGN.md`. This is a foundational story that unblocks every subsequent color-fix story in this backlog — uiqa-03 (ExecutionBar remap), uiqa-04 (shared `.glow-btn`), uiqa-05 (rgba → color-mix), and uiqa-06 (modal backdrops) all reference these new tokens. No component CSS changes in this story; only token declarations and documentation.

## Developer Notes

### Files to modify

- `frontend/src/style.css` (lines 1–46, `:root` block) — add four new custom properties immediately after `--accent-teal` (line 16) and `--duration-medium` (line 45).
- `DESIGN.md` (repo root) — add the four tokens to the color palette / motion sections so designers have a canonical reference.

### Exact token values (from the review, Section 7)

Add to `:root` in `style.css`:

```css
/* Accent extensions — added for ExecutionBar and signature moments */
--accent-cyan: #22d3ee;      /* ExecutionBar .ctrl-btn.run */
--accent-orange: #fb923c;    /* ExecutionBar .ctrl-btn.stop */

/* Overlay + glow primitives */
--overlay-backdrop: rgba(0, 0, 0, 0.6);   /* modal scrims */
--glow-spread: 0 0 12px;                   /* reusable neon text-shadow base */
```

Rationale for values:
- `#22d3ee` and `#fb923c` match the current ExecutionBar hardcodes at `ExecutionBar.svelte:163,168,197,202` (see review Section 4 "High: Colors outside design system"). Keeping the exact same shade means uiqa-03 is a pure find/replace with zero visual regression.
- `rgba(0, 0, 0, 0.6)` is the value the review flagged as the "~10 instance" black overlay pattern (review Section 4 "Pattern: rgba() with hardcoded RGB", row `rgba(0, 0, 0, ...)`).
- `--glow-spread: 0 0 12px` is the first half of a `text-shadow` declaration — callers append a color (`text-shadow: var(--glow-spread) var(--accent-green)`). This is intentional: it lets uiqa-04 and later stories compose the glow with any accent.

### DESIGN.md updates

- Add `--accent-cyan` and `--accent-orange` under the existing accent color list with the note "used for ExecutionBar run/stop controls; not part of the status palette."
- Add `--overlay-backdrop` under a new "Overlays" subsection.
- Add `--glow-spread` under "Motion & effects" with a usage example:
  `text-shadow: var(--glow-spread) var(--accent-green);`

### Risks

- Token naming collision: none exist in current `style.css` (verified at lines 1–46). The four names are new.
- If `DESIGN.md` already has an "Overlays" section, append to it instead of creating a new one — check before editing.
- Do not rename existing tokens. Do not change existing values. Additive only.

### Out of scope

- Updating any `.svelte` file to consume the new tokens — that's uiqa-03 through uiqa-06.
- Adding `--accent-red-glow`, `--accent-amber-glow`, etc. If those are needed later, they're additions in uiqa-05, not here.
- Removing the old hardcoded values in component files (separate stories handle that).

### Reference skills

`/simplify` on the edited files. No backend work, so no `/golang-*` skills needed.

## Acceptance Criteria

**AC-1: Four new tokens exist in `:root`**
- Given `frontend/src/style.css`
- When the file is parsed
- Then `--accent-cyan`, `--accent-orange`, `--overlay-backdrop`, and `--glow-spread` are all declared inside the `:root` block
- And each has the exact value listed in the Developer Notes

**AC-2: Tokens resolve at runtime**
- Given the running frontend (any view that loads `style.css`)
- When `getComputedStyle(document.documentElement).getPropertyValue('--accent-cyan')` is called
- Then it returns a non-empty string equal to `#22d3ee` (with or without leading whitespace)
- And the same is true for `--accent-orange`, `--overlay-backdrop`, and `--glow-spread`

**AC-3: DESIGN.md documents each token**
- Given `DESIGN.md`
- When the file is read
- Then each of the four new token names appears at least once
- And each entry includes a one-line purpose description

**AC-4: No existing tokens were removed or renamed**
- Given a diff of `frontend/src/style.css` against main
- When inspected
- Then the diff contains only additions inside `:root` (no deletions, no token renames)

**AC-5: Frontend build succeeds**
- Given the modified `style.css`
- When `wails dev` (or equivalent frontend build) runs
- Then the build completes without CSS parse errors
- And no console warnings reference the new tokens

## BDD Test Scenarios

```gherkin
Feature: uiqa-01 token foundation

  Scenario: Computed style exposes --accent-cyan
    Given a test harness that mounts a root element with style.css applied
    When it reads getComputedStyle(document.documentElement).getPropertyValue('--accent-cyan')
    Then the result trimmed equals "#22d3ee"

  Scenario: Computed style exposes --accent-orange
    Given the same harness
    When it reads '--accent-orange'
    Then the result trimmed equals "#fb923c"

  Scenario: Computed style exposes --overlay-backdrop
    Given the same harness
    When it reads '--overlay-backdrop'
    Then the result trimmed equals "rgba(0, 0, 0, 0.6)"

  Scenario: Computed style exposes --glow-spread
    Given the same harness
    When it reads '--glow-spread'
    Then the result trimmed equals "0 0 12px"

  Scenario: DESIGN.md documents all four tokens
    Given DESIGN.md is read
    When searched for the literal strings "--accent-cyan", "--accent-orange", "--overlay-backdrop", "--glow-spread"
    Then each of the four strings is found at least once

  Scenario: No existing tokens removed
    Given the git diff of style.css against main
    When the diff is inspected
    Then every removed line matches an added line with identical content (pure reformat) or the diff has zero deletions
```

## Tasks / Subtasks

- [x] Task 1: Add the four tokens to `frontend/src/style.css` `:root` block (AC-1, AC-4)
  - [x] Subtask 1a: Insert `--accent-cyan: #22d3ee;` and `--accent-orange: #fb923c;` after line 16 (`--accent-teal`)
  - [x] Subtask 1b: Insert `--overlay-backdrop: rgba(0, 0, 0, 0.6);` and `--glow-spread: 0 0 12px;` after line 45 (`--duration-medium`)
  - [x] Subtask 1c: Verify no existing token was moved, renamed, or removed

- [x] Task 2: Document tokens in `DESIGN.md` (AC-3)
  - [x] Subtask 2a: Read existing DESIGN.md structure to locate color palette section
  - [x] Subtask 2b: Append the four tokens with one-line purpose descriptions each
  - [x] Subtask 2c: Include the `text-shadow: var(--glow-spread) var(--accent-green);` usage example

- [x] Task 3: Add a Svelte component test asserting token resolution (AC-2, AC-5)
  - [x] Subtask 3a: Create `frontend/src/components/__tests__/tokens.test.ts` (matches existing frontend test path)
  - [x] Subtask 3b: Mount a DOM node, import style.css, assert each of the four tokens via `getPropertyValue`
  - [x] Subtask 3c: Ensure frontend `npm run build` exits 0

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ coverage on any new test file added
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] Frontend build (`wails build` or `npm run build` inside `frontend/`) passes with zero errors and zero new warnings
- [x] `/simplify` run on modified `style.css` and `DESIGN.md`
- [x] No `#39ff14` introduced anywhere in the diff
- [x] No orphaned raw color or size values introduced in the diff
- [x] Story status updated to `done`
