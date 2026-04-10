# Story 2: File Selection UI in Summarisation Modal

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Add click-to-select functionality to file cards in the Summarisation Modal, allowing users to pick individual files for scoped code review advice. Selected files get a visual accent border and background tint. "Select All" and "Select None" controls appear near the totals bar. The advice and plan buttons become gated on having a selection, preventing users from running advice on zero files.

## Developer Notes

### Architecture

- **File to modify:** `/Users/linus/Development/mashed/frontend/src/views/SummarisationModal.svelte`
- **No backend changes** -- this story is pure frontend state and UI.

### State Changes

Add to the `<script>` section (after line 20, alongside existing state variables):

```typescript
// Selection state
let selectedFiles = new Set<string>();

// Computed
$: hasSelection = selectedFiles.size > 0;
```

Modify the existing computed properties (line 45-46):
```typescript
$: canGetAdvice = selectedMode && !adviceLoading && hasSelection;
$: canCreatePlan = adviceText && !planLoading && hasSelection;
```

### UI Changes

**File card selection (around line 202-226):**
- Add `on:click` handler to `.file-card` div that toggles the file path in `selectedFiles`
- Add conditional class: `class:selected={selectedFiles.has(file.path)}`
- The existing `file-path` button (opens file) must call `stopPropagation` to prevent toggle on click-to-open

**Select All / Select None (near line 230, inside or adjacent to `.totals-bar`):**
- Add a row of text links: "Select All" and "Select None"
- `selectAll()` sets `selectedFiles = new Set(files.map(f => f.path))`
- `selectNone()` sets `selectedFiles = new Set()`
- These links appear only when `hasFiles` is true

### CSS Changes

Add to the `<style>` section:
```css
.file-card.selected {
  border-color: var(--accent-green);
  background: color-mix(in srgb, var(--accent-green) 4%, var(--bg-elevated));
}

.file-card {
  cursor: pointer;  /* Add to existing .file-card rule */
}
```

**Important:** Use `color-mix()` not `rgba()` for the selected background tint (per cerebrum Do-Not-Repeat rule about hardcoded rgba).

### Technical Considerations

- **Svelte reactivity with Set:** Svelte does not detect mutations on Set objects. After `.add()` or `.delete()`, reassign: `selectedFiles = new Set(selectedFiles)` (or `selectedFiles = selectedFiles` after mutation won't work -- must create new Set).
- **Selection during streaming:** Files arrive one at a time via `review:summary:progress`. The selection Set is independent -- selecting a file while more are streaming in works naturally because Set membership is path-based.
- **Deselect after advice:** If user deselects all files after advice has been generated, `canCreatePlan` becomes false (plan button disables), but `adviceText` remains visible.

### Risks & Edge Cases

- **Double-click on file-path button:** The `stopPropagation` on the inner `file-path` button prevents the card click handler from toggling selection when the user clicks the filename link.
- **No files after summary:** If the repo has no code changes (only markdown/config), `hasFiles` is false and the selection UI never appears -- this is correct behavior.
- **Large file count:** With 50 files (maxReviewFiles), selecting all is tedious without "Select All" -- hence the bulk controls.

### Reference Files

- `/Users/linus/Development/mashed/frontend/src/views/SummarisationModal.svelte` -- the only file to modify
- `/Users/linus/Development/mashed/.wolf/cerebrum.md` -- Do-Not-Repeat: use `color-mix()` not `rgba()`, use `var(--accent-green)` not `#39ff14`

## Acceptance Criteria

AC-1: Click-to-select file cards
- Given the Summarise modal is open with file summaries visible
- When the user clicks on a file card
- Then the card is visually highlighted with an accent-green border and subtle background tint
- And clicking the same card again deselects it (toggle behavior)

AC-2: Select All / Select None controls
- Given the Summarise modal has file summaries loaded
- When the user clicks "Select All"
- Then all file cards are selected (all paths in the selection Set)
- And when the user clicks "Select None"
- Then all file cards are deselected (empty selection Set)

AC-3: Advice button gated on selection
- Given no files are selected
- When the user looks at the advice controls
- Then the "Get Advice" button is disabled (grayed out)
- And when the user selects at least one file
- Then the "Get Advice" button becomes enabled (assuming a methodology is also selected)

AC-4: Plan button gated on selection
- Given advice text has been generated but all files are deselected
- When the user looks at the plan row
- Then the "Create Refactor Plan" button is disabled
- And the advice text remains visible (not cleared)

AC-5: Selection independent of streaming
- Given file summaries are actively streaming in (not yet complete)
- When the user selects a file card that has already appeared
- Then the selection succeeds immediately
- And newly arriving file cards appear unselected

AC-6: File open click does not toggle selection
- Given a file card is unselected
- When the user clicks the filename link (the blue path text)
- Then the `open-file` event dispatches
- And the file card does NOT become selected (stopPropagation)

## BDD Test Scenarios

### Scenario 1: Toggle selection

```gherkin
Feature: File card selection in Summarise modal

  Scenario: Select a single file card
    Given the modal displays 3 file cards: "main.go", "util.go", "config.go"
    When the user clicks the "main.go" file card body
    Then the "main.go" card has the CSS class "selected"
    And the "util.go" and "config.go" cards do not have the "selected" class

  Scenario: Deselect a selected file card
    Given "main.go" file card is selected
    When the user clicks the "main.go" file card body again
    Then the "main.go" card no longer has the "selected" class

  Scenario: Multiple selections
    Given no files are selected
    When the user clicks "main.go" and then "util.go" file cards
    Then both "main.go" and "util.go" have the "selected" class
    And "config.go" does not have the "selected" class
```

### Scenario 2: Bulk selection controls

```gherkin
Feature: Select All / Select None

  Scenario: Select All selects every file
    Given the modal displays 5 file cards with none selected
    When the user clicks "Select All"
    Then all 5 file cards have the "selected" class

  Scenario: Select None clears all selections
    Given all 5 file cards are selected
    When the user clicks "Select None"
    Then no file cards have the "selected" class

  Scenario: Select All then deselect one
    Given the user has clicked "Select All" on 5 file cards
    When the user clicks the "config.go" card
    Then 4 file cards are selected and "config.go" is not
```

### Scenario 3: Button gating

```gherkin
Feature: Advice and plan buttons require selection

  Scenario: Advice button disabled without selection
    Given no files are selected and a methodology is chosen
    When the user views the advice controls
    Then the "Get Advice" button is disabled

  Scenario: Advice button enabled with selection
    Given "main.go" is selected and a methodology is chosen
    When the user views the advice controls
    Then the "Get Advice" button is enabled

  Scenario: Plan button disabled after deselecting all
    Given advice text "Some advice" exists and no files are selected
    When the user views the plan row
    Then the "Create Refactor Plan" button is disabled
    And the advice text "Some advice" is still visible
```

### Scenario 4: File link isolation

```gherkin
Feature: Filename link does not toggle selection

  Scenario: Clicking filename opens file without selecting card
    Given "main.go" file card is not selected
    When the user clicks the blue filename link text inside the card
    Then an "open-file" event is dispatched with path "main.go"
    And the "main.go" card remains unselected
```

## Tasks / Subtasks

- [ ] Task 1: Add selection state and computed properties (AC: 3, 4)
  - [ ] Subtask 1a: Add `selectedFiles = new Set()` and `hasSelection` reactive declaration
  - [ ] Subtask 1b: Update `canGetAdvice` to require `hasSelection`
  - [ ] Subtask 1c: Update `canCreatePlan` to require `hasSelection`

- [ ] Task 2: Add click-to-select on file cards (AC: 1, 5, 6)
  - [ ] Subtask 2a: Add `toggleFile(path)` function that toggles Set membership with reassignment
  - [ ] Subtask 2b: Add `on:click={() => toggleFile(file.path)}` and `class:selected` to `.file-card` div
  - [ ] Subtask 2c: Add `on:click|stopPropagation` to the inner `.file-path` button
  - [ ] Subtask 2d: Add `cursor: pointer` to `.file-card` CSS rule

- [ ] Task 3: Add Select All / Select None controls (AC: 2)
  - [ ] Subtask 3a: Add `selectAll()` and `selectNone()` functions
  - [ ] Subtask 3b: Add link-style buttons near the totals bar, visible only when `hasFiles`
  - [ ] Subtask 3c: Style the links consistently with existing UI (text-dim, hover to text-primary)

- [ ] Task 4: Add selected card CSS (AC: 1)
  - [ ] Subtask 4a: Add `.file-card.selected` rule with `border-color: var(--accent-green)` and `color-mix` background
  - [ ] Subtask 4b: Verify against design system (no hardcoded colors)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass via manual browser test or Playwright
- [ ] Visual: selected cards have visible accent-green border + tinted background
- [ ] Visual: unselected cards look unchanged from current design
- [ ] `go build ./...` passes (no backend changes, but verify no breakage)
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues

## Design Brief

### 1. Layout Composition

The selection controls ("Select All" / "Select None") integrate into the **existing `.totals-bar`** element rather than creating a new row. They sit on the **left side**, replacing the static "Overall" label with a richer inline layout:

```
[ Select All | Select None ]    [ 3 of 8 selected ]    [ +142  -37 ]
```

- The totals bar already uses `display: flex; justify-content: space-between; padding: var(--sp-sm) var(--sp-xl)`.
- Left group: selection links separated by a `var(--text-muted)` pipe character at `var(--sp-xs)` horizontal margin.
- Center: selection count badge (only visible when `hasSelection`), positioned via `margin-left: auto; margin-right: auto` or a flex spacer.
- Right group: existing `.totals-stats` (added/removed counts) unchanged.
- The file card list retains its existing `gap: var(--sp-sm)` between cards. No spacing changes to `.file-list`.

When `summaryLoading` is still true, the totals bar (and thus selection controls) remains hidden -- the existing `{#if !summaryLoading && hasFiles}` guard is sufficient.

### 2. Typography Plan

| Element | Token | Weight | Transform | Letter-spacing |
|---|---|---|---|---|
| "Select All" / "Select None" links | `font-size: var(--text-label)` (11px) | 500 | none | normal |
| Pipe separator | `font-size: var(--text-label)` | 400 | none | normal |
| Selection count ("3 of 8 selected") | `font-size: var(--text-label)` | 600 | uppercase | `0.05em` |
| Existing "Overall" label | removed -- replaced by the selection links |
| Existing totals stats | unchanged: `font-size: var(--text-body)`, `font-family: var(--font-mono)`, `font-variant-numeric: tabular-nums` |

All selection control text uses `font-family: var(--font-ui)` (Geist). The selection count uses monospace numerals via `font-variant-numeric: tabular-nums` so the count does not shift layout as digits change.

### 3. Color Strategy

**Selected card:**
- Border: `border-color: var(--accent-green)` (full `#00e57a`, not dimmed -- this is the primary selection signal)
- Background: `background: color-mix(in srgb, var(--accent-green) 4%, var(--bg-elevated))` -- a barely-perceptible green wash that reads as "active" without overwhelming the summary text
- The `.file-path` link color remains `var(--accent-blue)` when selected -- no color override on inner elements

**Unselected card (unchanged):**
- Border: `border-color: var(--border-subtle)` (existing)
- Background: `var(--bg-elevated)` (existing)

**Hover states (pre-click feedback):**
- Unselected hover: `border-color: var(--border-emphasis)` (existing rule, unchanged)
- Selected hover: `border-color: var(--accent-green); background: color-mix(in srgb, var(--accent-green) 7%, var(--bg-elevated))` -- slightly intensified tint to signal "you can click to deselect"

**Selection control links:**
- Default: `color: var(--text-dim)`
- Hover: `color: var(--accent-green)`
- Active link for current state (e.g., "Select All" when all are already selected): `color: var(--text-muted)` with `pointer-events: none` to indicate no-op

**Selection count badge:**
- Text: `color: var(--accent-green)`
- No background or border -- pure inline text, not a pill

**Disabled button styling (existing `.action-btn:disabled`):**
- Already uses `opacity: 0.4; cursor: not-allowed` -- this is correct and sufficient
- No color change needed; the reduced opacity communicates disabled state clearly

### 4. Interaction Model

**Click-to-select cards:**
- Click anywhere on `.file-card` div toggles selection (add/remove from `selectedFiles` Set)
- The inner `.file-path` button uses `on:click|stopPropagation` to prevent toggle when clicking the filename link
- `cursor: pointer` added to `.file-card` to signal clickability
- No drag-to-select -- click only (drag conflicts with scroll in the `.file-list` viewport)

**Keyboard accessibility:**
- Each `.file-card` gets `role="checkbox"`, `aria-checked={selectedFiles.has(file.path)}`, and `tabindex="0"`
- `on:keydown` handler on `.file-card`: Space or Enter toggles selection (calls `toggleFile(file.path)`)
- Tab order: file cards are naturally in DOM order within `.file-list`, so Tab moves through them sequentially
- Focus ring: `.file-card:focus-visible` gets `outline: 1px solid var(--accent-blue); outline-offset: var(--sp-2xs)`
- "Select All" and "Select None" are `<button>` elements (not `<a>` tags) for correct keyboard semantics

**Hover feedback before click:**
- Unselected cards: existing `border-color: var(--border-emphasis)` hover is sufficient to signal interactivity
- Selected cards: background tint intensifies from 4% to 7% on hover (see Color Strategy)
- Cards get `cursor: pointer` to reinforce clickability

**Transition timing:**
- Selection state change (border-color + background): `var(--duration-medium)` (150ms), `var(--ease-enter)` (ease-out)
- This matches the existing `.file-card` transition: `transition: border-color var(--duration-short) var(--ease-enter)` but extends duration slightly for the richer two-property change

### 5. Component Specs

**`.file-card` (modified existing rule):**
- Add: `cursor: pointer`
- Extend existing transition: `transition: border-color var(--duration-medium) var(--ease-enter), background var(--duration-medium) var(--ease-enter)`
- Existing values unchanged: `background: var(--bg-elevated)`, `border: 1px solid var(--border-subtle)`, `border-radius: var(--radius-md)`, `padding: var(--sp-md) var(--sp-lg)`

**`.file-card.selected` (new rule):**
- `border-color: var(--accent-green)`
- `background: color-mix(in srgb, var(--accent-green) 4%, var(--bg-elevated))`

**`.file-card.selected:hover` (new rule):**
- `background: color-mix(in srgb, var(--accent-green) 7%, var(--bg-elevated))`

**`.file-card:focus-visible` (new rule):**
- `outline: 1px solid var(--accent-blue)`
- `outline-offset: var(--sp-2xs)` (2px)
- `border-radius: var(--radius-md)`

**`.totals-bar` (modified):**
- Layout changes to accommodate three groups: selection links (left), count (center), stats (right)
- `gap: var(--sp-sm)` between child elements

**`.selection-controls` (new element inside totals-bar):**
- `display: flex; align-items: center; gap: var(--sp-xs)`

**Selection link buttons (`.select-link`):**
- `background: none; border: none`
- `padding: var(--sp-2xs) var(--sp-xs)` (2px 4px) -- minimal hit target padding
- `border-radius: var(--radius-sm)` (2px)
- `font-size: var(--text-label)`
- `font-weight: 500`
- `color: var(--text-dim)`
- `cursor: pointer`
- `transition: color var(--duration-short) var(--ease-enter)`

**Selection count (`.selection-count`):**
- `font-size: var(--text-label)`
- `font-weight: 600`
- `color: var(--accent-green)`
- `text-transform: uppercase`
- `letter-spacing: 0.05em`
- `font-variant-numeric: tabular-nums`

**Pipe separator:**
- `color: var(--text-muted)`
- Inline text character `|`, no separate element needed

### 6. Motion Plan

**Card selection state change:**
- Properties: `border-color`, `background`
- Duration: `var(--duration-medium)` (150ms) -- fast enough to feel instant, slow enough to register the color shift
- Easing: `var(--ease-enter)` (ease-out) -- snaps into the selected state, matching existing card hover behavior
- Both properties on the same transition line for synchronized change

**Card hover (selected state):**
- The background tint shift from 4% to 7% uses the same `var(--duration-medium)` transition already declared on the element -- no additional transition rule needed

**Selection count appearance:**
- When `hasSelection` becomes true, the count text appears. Use a Svelte `transition:fade={{ duration: 100 }}` for a micro-fade rather than a pop-in.

**No motion on Select All / Select None:**
- These are instant state operations affecting many cards simultaneously. The per-card `border-color` + `background` transitions provide the visual feedback. Adding animation to the links themselves would compete for attention.

**Totals bar entry (existing):**
- Already animated with `in:fly={{ y: 8, duration: 200, easing: cubicOut }}` -- this naturally includes the new selection controls, so no additional entry animation is needed.

### 7. Signature Elements

What makes this feel distinctly "mashed" rather than generic checkbox selection:

**No checkboxes.** The entire card is the selection target. The green border wrapping the full card creates a bold, confident selection state that reads instantly in a list of 20+ files. This matches mashed's terminal-monitor aesthetic: you select items by highlighting them, not by ticking boxes.

**Green border as brand signal.** `var(--accent-green)` (#00e57a) is mashed's signature color. Using it as the selection indicator ties file selection to the same visual language as the app's status indicators, cursor blink, and text selection highlight (defined in `::selection` in style.css). The 4% `color-mix` background tint is barely visible but subliminally reinforces the green identity.

**Information-dense totals bar.** Rather than adding a separate toolbar, the selection controls are woven into the existing totals bar. "3 of 8 selected" in green alongside the +/- stats creates a dense, dashboard-style status line that feels like a terminal footer -- compact, data-rich, no wasted space.

**Instant tactile feedback.** The 150ms transition on border-color + background is calibrated to feel like a physical click -- fast enough that there is no perceived delay, but with just enough easing to avoid a harsh binary snap. Combined with the existing `transform: scale(0.98)` on `:active` (inherited from `.action-btn` -- note: this should NOT be added to file cards; cards should not scale on press), the interaction feels responsive without being bouncy.

**Keyboard-native.** `role="checkbox"` with Space/Enter toggle and visible focus rings means power users can Tab through the file list and select files without touching the mouse. This is a desktop-app convention that web apps often skip -- mashed does not.
