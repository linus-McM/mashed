# Story 04: Settings Layout Refactor — Two-Column Panel Grid

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** none (pure layout refactor; orthogonal to stories 01-03)
**Status:** done
**UI-facing:** YES — flag for ui-architect design brief

## Description

Refactor `Settings.svelte`'s right-side `.col-settings` from a single vertical stack into a 2-column grid of visually separated panels. Keep the left theme list and all existing controls untouched; only their containers change. Each existing `<section class="settings-section">` gets wrapped in a `.settings-panel` card, and the five current panels (Font, Sidebar Width, Editor, Theme Extensions, UI AST adapter) are distributed into two columns. This lands before story 05 so the Markdown Editor panel has a home.

## Developer Notes

### Why this is its own story
The layout refactor touches every existing panel's container markup. Bundling it with the Markdown Editor panel addition (story 05) would make the diff hard to review and risks breaking unrelated settings. Shipping the grid first, with NO new content, lets QA verify nothing regressed for existing controls before the new toggle section shows up.

### Architecture
- File to modify: `/Users/dev/Development/mashed/frontend/src/views/Settings.svelte`
- Within `.settings-body` the current structure is:
  ```
  .col-themes          (unchanged, 280px)
  .col-settings        (flex:1, currently vertical stack via default block flow)
    └── 5x <section class="settings-section">
  ```
- New structure:
  ```
  .col-themes          (unchanged)
  .col-settings        (flex:1, now display:grid)
    ├── .settings-col.settings-col-1
    │   ├── .settings-panel > Font
    │   ├── .settings-panel > Sidebar Width
    │   └── .settings-panel > Editor
    └── .settings-col.settings-col-2
        ├── .settings-panel > Theme Extensions
        └── .settings-panel > UI AST adapter
  ```
- Note: the current `Settings.svelte` (per the repo corpus) has Font / Sidebar Width / Editor / Theme Extensions — no visible "UI AST adapter" section. If UI AST adapter does NOT yet exist in `Settings.svelte`, place the four existing panels as Font + Sidebar Width in column 1, Editor alone in column 1 or 2 depending on height balance, Theme Extensions in column 2. The Markdown Editor panel (story 05) will occupy the Column 2 bottom slot.

### Exact CSS additions (from plan)

```css
.col-settings {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--sp-lg);
  align-items: start;
  /* keep existing padding and overflow-y: auto */
}

.settings-col {
  display: flex;
  flex-direction: column;
  gap: var(--sp-lg);
  min-width: 0;
}

.settings-panel {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--sp-lg);
}

.settings-panel .settings-section {
  margin-bottom: 0;
}

@media (max-width: 1100px) {
  .col-settings { grid-template-columns: 1fr; }
}
```

### Design system compliance (DESIGN.md)
- `--bg-surface`, `--border-subtle`, `--radius-md`, `--sp-lg` are existing design tokens — use them, do not introduce new ones.
- `border-radius: var(--radius-md)` == 4px per DESIGN.md — consistent with cards/inputs.
- Compact density is the rule — `--sp-lg` (16px) is tight but still breathing room for cards.

### Panel wrapper convention
Do NOT extract a `Panel.svelte` component for this story. The plan notes "or small Panel.svelte if preferred" but a minimal inline wrapper avoids coupling and keeps the diff reviewable:

```svelte
<div class="settings-panel">
  <section class="settings-section">
    <!-- existing content unchanged -->
  </section>
</div>
```

Existing `.section-title`, `.section-desc`, `.subsection-title`, `.setting-row`, `.setting-toggle`, `.setting-select`, `.setting-number` classes stay unchanged. `.settings-section` keeps its internal structure; only its outer spacing (`margin-bottom: var(--sp-2xl)`) is zeroed out by the new rule because the panel gap provides spacing.

### Responsive behavior
At viewport width ≤ 1100px the grid collapses to a single column. Panels remain as cards (still have borders). This preserves usability on narrower windows without needing a separate mobile layout.

### Risks & Edge Cases
- **Existing `.settings-section { margin-bottom: var(--sp-2xl); }` rule** — with the panel wrapper's padding + grid gap, the old margin would cause double-spacing. The new rule `.settings-panel .settings-section { margin-bottom: 0 }` cancels it.
- **`.col-settings` overflow** — existing `overflow-y: auto` on `.col-settings` must be preserved; the grid replacement still needs to scroll when content is tall.
- **VSCodium theme panel height imbalance** — Theme Extensions panel can get tall (long import list). If Column 2 dominates, Editor panel in Column 1 may look isolated. This is acceptable — `align-items: start` lets columns be independent heights.
- **Drag-reorder of panels not supported** — the plan rules this out explicitly; column assignments are hardcoded.

### Reference Files
- `/Users/dev/Development/mashed/frontend/src/views/Settings.svelte` — the file to modify.
- `/Users/dev/Development/mashed/DESIGN.md` — tokens and conventions.
- `/Users/dev/Development/mashed/frontend/src/style.css` — CSS variable definitions.

## Acceptance Criteria

AC-1: Two-column grid renders at ≥ 1100px
- Given the app runs in a viewport ≥ 1100px wide
- When the Settings view is opened
- Then `.col-settings` is a CSS grid with two columns of equal width
- And Font, Sidebar Width, Editor panels are in column 1
- And Theme Extensions (and any additional existing panel) is in column 2

AC-2: Each panel is a visually separated card
- Given the Settings view is open
- When each `.settings-panel` is inspected
- Then it has `background: var(--bg-surface)`, a 1px border `var(--border-subtle)`, `border-radius: 4px`, and `padding: var(--sp-lg)`

AC-3: Narrow viewport collapses to a single column
- Given the viewport is resized to 1000px wide
- When Settings is rendered
- Then `.col-settings` uses a single-column grid
- And panels still render as bordered cards

AC-4: All existing controls continue to function
- Given the Settings view after the refactor
- When I change font size, toggle minimap, enter a VSCodium path, etc.
- Then each control persists its value exactly as before the refactor
- And keyboard focus order is preserved (tab order within a column goes top-to-bottom)

AC-5: No double spacing between panel content and card edge
- Given a `.settings-panel` containing a `.settings-section`
- When computed styles are inspected
- Then the `.settings-section`'s `margin-bottom` is `0px`
- And there is no visible dead space below the last control in each panel

AC-6: Design tokens only — no hardcoded colors or lengths
- Given the diff for this story
- When CSS changes are audited
- Then every color uses a `var(--...)` token
- And every spacing/radius uses a `var(--sp-*)` or `var(--radius-*)` token

## BDD Test Scenarios

```gherkin
Feature: Settings page 2-column panel layout

  Scenario: Desktop viewport renders two columns
    Given the window is 1400px wide
    When I open Settings
    Then the right column contains a two-column grid
    And column 1 contains Font, Sidebar Width, Editor panels
    And column 2 contains Theme Extensions panel

  Scenario: Panels are visually bordered cards
    Given Settings is open
    When I inspect any .settings-panel
    Then its background is var(--bg-surface)
    And it has a 1px border using var(--border-subtle)
    And its border-radius resolves to 4px

  Scenario: Narrow window collapses to single column
    Given the window is 1000px wide
    When I open Settings
    Then the right column is a single-column grid
    And panels stack vertically
    And each panel retains its card appearance

  Scenario: Existing controls still work after refactor
    Given Settings is open
    When I increase the font size from 13 to 15
    Then the preview updates immediately
    And GetConfig() eventually returns fontSize: 15

  Scenario: No hardcoded CSS values
    Given the Settings.svelte <style> block after refactor
    When scanned for color or spacing literals
    Then there are no hex colors, rgb() or rgba() calls
    And no px values outside of --sp-* or --radius-* tokens (except where already present before the refactor)
```

## Tasks / Subtasks

- [ ] Task 1: Markup restructure (AC-1, AC-4) — frontend
  - [ ] Wrap each existing `<section class="settings-section">` in a `<div class="settings-panel">`.
  - [ ] Group the panels into two `<div class="settings-col settings-col-1">` / `settings-col-2` containers inside `.col-settings`.
  - [ ] Assign panels: Font, Sidebar Width, Editor → col 1; Theme Extensions (+ UI AST adapter if present) → col 2.
- [ ] Task 2: CSS additions (AC-1, AC-2, AC-3, AC-5, AC-6) — frontend
  - [ ] Replace `.col-settings` single-column rules with the grid rules above.
  - [ ] Add `.settings-col` and `.settings-panel` rules.
  - [ ] Add `.settings-panel .settings-section { margin-bottom: 0 }` override.
  - [ ] Add `@media (max-width: 1100px)` single-column fallback.
- [ ] Task 3: Component tests (AC-1, AC-3) — frontend
  - [ ] Add a render test that mounts Settings at a wide viewport (mock or set viewport) and asserts grid column count by querying computed styles or the number of `.settings-col` children.
  - [ ] Add a narrow-viewport test (mock `matchMedia` or similar) — assert fallback.
- [ ] Task 4: Visual verification (AC-2, AC-4, AC-6) — frontend / ui-architect
  - [ ] Run the app in `wails dev` and open Settings. Confirm cards, borders, spacing match DESIGN.md.
  - [ ] Resize window below 1100px, confirm collapse.
  - [ ] Run `/simplify` on the modified Settings.svelte.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests (or are verified manually and documented when a scenario cannot be unit-tested)
- [ ] `svelte-check` passes with 0 warnings, 0 errors
- [ ] `/simplify` run on Settings.svelte
- [ ] Design review by ui-architect (screenshot comparison pre/post refactor) — no regressions on existing controls
- [ ] Code review: no CRITICAL/HIGH issues

## Design Brief

### Intent
Settings is the command center's configuration surface. The current single-column stack scrolls forever and wastes horizontal space on wide displays. A 2-column panel grid converts it from "long document" to "dashboard": every panel earns a framed slot, density stays Bloomberg-tight, and the eye can scan two orthogonal stacks in parallel. This is Linear's settings-as-cards language, not VS Code's flat-list language.

### 1. Layout composition
- **Outer structure preserved:** `.settings-body` stays `display: flex` with `.col-themes` (fixed `280px`, border-right `1px solid var(--border-subtle)`) + `.col-settings` (flex: 1, `overflow-y: auto`, existing padding `var(--sp-lg) var(--sp-xl)` retained).
- **New grid on `.col-settings`:**
  ```css
  .col-settings {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--sp-lg);        /* 16px — matches panel internal padding for visual rhythm */
    align-items: start;        /* columns grow independently; no forced equal height */
    /* preserve existing: flex:1, overflow-y:auto, padding, min-width:0 */
  }
  ```
- **Column wrapper** — each `.settings-col` is a vertical flex stack:
  ```css
  .settings-col {
    display: flex;
    flex-direction: column;
    gap: var(--sp-lg);         /* 16px between stacked panels — same token as grid gap */
    min-width: 0;              /* allow shrink inside grid */
  }
  ```
- **Panel distribution (visual balance — shorter content in col-2 to match col-1's 3-panel height):**
  - Column 1: Font (tallest — live preview + slider + font list) / Sidebar Width (compact) / Editor (tall — many rows)
  - Column 2: Theme Extensions (medium — path + import list) / [Markdown Editor panel from story 05 slots below Theme Extensions] / UI AST adapter (if present)
- **Row alignment within each panel** — existing `.setting-row` with `justify-content: space-between` is preserved; label left-aligned, control right-aligned. No changes.
- **Responsive collapse:**
  ```css
  @media (max-width: 1100px) {
    .col-settings { grid-template-columns: 1fr; }
  }
  ```
  Breakpoint reasoning: left theme column is `280px`, minimum useful right column is ~800px (two panels at ~380px + `var(--sp-lg)` gap). Below `1100px` the grid stops earning its keep; collapse rather than cramp.

### 2. Typography plan
No type changes — panels inherit the existing scale. For reference:
- **Panel title (`.section-title`)** — `var(--font-mono)`, `var(--text-data)` (14px), weight 600, `color: var(--text-primary)`, `margin-bottom: var(--sp-md)`. Mono face is the signature — it's what makes Settings read as "command console" not "marketing page."
- **Description (`.section-desc`)** — `var(--text-body)` (13px), `color: var(--text-dim)`.
- **Subsection (`.subsection-title`)** — existing 11–13px mono — unchanged.
- **No new headings.** The panel card border does the hierarchy work; adding display typography would double-signal.

### 3. Color strategy
- **Panel surface:** `background: var(--bg-surface)` (`#0d0f12`) — one step above `--bg-deepest` page bg. This creates the "panel lifts off page" effect without shadow (DESIGN.md: Minimal decoration).
- **Panel border:** `1px solid var(--border-subtle)` (`#1e2530`) — whisper-thin at rest.
- **Panel corner:** `border-radius: var(--radius-md)` (4px) — matches inputs/cards per DESIGN.md.
- **No accent color on panels.** Panels are passive frames. Accent (`--accent-green`) is reserved for active toggles, focus rings, and the Back button — keeping the neon punch meaningful.
- **No box-shadow.** DESIGN.md specifies minimal decoration; layered background tokens are the depth language.
- **No hover state on the panel itself.** The panel is scenery; the controls inside are the interactive surface.

### 4. Interaction model
- **Panel is static** — no hover, no active, no focus. It is a container.
- **Controls inside** keep their existing focus/hover behaviour (the global `:focus-visible` rule + each control's own transitions).
- **Keyboard navigation:** DOM order determines tab order. With `display: grid` the visual order equals document order (col-1 panels read top-to-bottom, then col-2 panels read top-to-bottom) — which is the expected tab flow. No `tabindex` overrides required.
- **Escape key** — existing `handleKeydown` already dispatches `back`; unchanged by this refactor.
- **Resize behaviour:** the `@media` collapse at 1100px is instant (no transition). Snap-layout is correct here — an animated column collapse would feel gimmicky on a settings page and slow the dev QA loop.
- **No transitions on grid itself** — a settings refactor should feel deterministic. The panels appear framed on mount; no entry animation.

### 5. Component specs
Exact `.settings-panel` rule (the only new primitive):
```css
.settings-panel {
  background: var(--bg-surface);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  padding: var(--sp-lg);
}

/* Neutralise existing .settings-section margin so panel padding owns spacing */
.settings-panel .settings-section {
  margin-bottom: 0;
}
```
All four values are tokens — zero magic numbers.

**Inside-panel spacing contract:**
- Panel padding: `var(--sp-lg)` (16px) on all four sides.
- Section title's existing `margin-bottom: var(--sp-md)` (12px) still applies between title and first row.
- `.setting-row` existing `padding: var(--sp-xs) 0` (4px vertical) = the row-to-row rhythm.
- Last `.setting-row` sits `var(--sp-lg)` from the panel's bottom edge via panel padding (not via section margin).

**Column gap contract:**
- Both grid `gap` and within-column `gap` are `var(--sp-lg)` — this uniformity is intentional. It creates a single background rhythm: every gap you see between frames is 16px, so the eye locks onto the grid instead of parsing mixed spacings.

### 6. Signature elements
What makes this unmistakably Mashed rather than generic settings:
- **Token-only spatial system.** Every dimension traces to `--sp-*` / `--radius-*` — no 5px, no 10px, nothing outside the ladder. The discipline is visible.
- **Dark-on-dark elevation via tokens, not shadows.** `--bg-deepest` page → `--bg-surface` panel is the depth cue. DESIGN.md rule: "typography, color, and density do all the work."
- **Mono-face section titles.** `Geist Mono` headings are the terminal-native signal. Keep them.
- **Whisper borders.** `--border-subtle` (`#1e2530`) is nearly black — panels read as framed, not boxed-in. If you need to see them, you see them; otherwise they recede.
- **Breakpoint at 1100px, not 768px.** This is a desktop app for Mac windows, not a responsive website. The breakpoint exists only so narrow window users don't get cramped — it's utility, not mobile design.
- **Compact 4px corner radius.** DESIGN.md's `--radius-md` — assertively square compared to the 8–12px radii of consumer apps. Reads as "tool," not "product page."

### 7. Anti-patterns to avoid
- **No hardcoded widths** — `grid-template-columns: 1fr 1fr`, never `400px 400px`. Column widths must flex with window size.
- **No magic pixel values** anywhere in the new rules. Every `16px` is `var(--sp-lg)`. Every `4px` corner is `var(--radius-md)`.
- **No inline hex colors.** Every color is a `var(--...)` reference. Inline hex on panels was called out as a bug class in AC-6.
- **No `box-shadow` on panels.** DESIGN.md forbids it; depth comes from the `--bg-deepest` → `--bg-surface` step.
- **No `margin` on `.settings-panel`.** Use the parent `gap` exclusively. Mixing gap+margin creates the double-spacing bug AC-5 exists to prevent.
- **No `Panel.svelte` component abstraction** — story explicitly rules this out. An inline `<div class="settings-panel">` keeps the diff reviewable.
- **No transition on the grid collapse** — instant at breakpoint; animated column reflows feel janky.
- **Do not touch `.col-themes`** — left column is explicitly untouched. Any cross-column visual change is scope creep.
- **No new accent usage on panels.** Green is reserved. Bordering a panel in `--accent-green` would steal the signal from active toggles and focus rings.
- **Do not remove `align-items: start`.** Without it, the taller column forces the shorter column to stretch, creating dead space at the bottom of col-2.
