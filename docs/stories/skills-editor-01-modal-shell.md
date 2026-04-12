# skills-editor-01: Skill editor modal shell + sidebar right-click entrypoint

**Status:** ready
**Domain:** frontend
**Size:** S
**Depends on:** none (Phase 1 is shipped)
**Phase:** 4a

## Description

Add a Svelte modal component that opens from a right-click on a mashed-asset row in `ProcessSidebar.svelte`. The modal shell displays the asset's current frontmatter fields as a form (name, description, mashedRole, mashedCompletion, mashedChainable, mashedSessionPinned) and exposes Save / Cancel buttons. Save is wired in skills-editor-02; this story is the UI shell plus the open/close flow.

**This is a UI-facing story — route through ui-architect before implementation.**

## Developer Notes

- **Files to create/modify:**
  - `frontend/src/components/bmad/SkillEditorModal.svelte` — NEW. Form shell with frontmatter fields.
  - `frontend/src/components/bmad/ProcessSidebar.svelte` — add a right-click (contextmenu) handler on mashed-asset rows that emits an `editAsset` event; parent view wires it to open the modal with the asset pre-loaded.
  - `frontend/src/views/WorkflowBuilder.svelte` — host the modal state (`editingAsset | null`), listen for `editAsset`, render `<SkillEditorModal>` when non-null.
- **Field list (from plan §"The schema"):**
  - `name` (text, required)
  - `description` (multiline text, required)
  - `mashedRole` (select: `command | skill | agent`)
  - `mashedCompletion` (select: `idle | exit | marker: <str> | timeout: <duration>`) — this story uses a simple select with three options `idle | exit | custom`; skills-editor-02 handles the custom path
  - `mashedChainable` (select: `single | none | any`)
  - `mashedSessionPinned` (checkbox)
  - `mashedInputs` / `mashedOutputs` (comma-separated text → array)
- **Risks / gotchas:**
  - Right-click handler must call `event.preventDefault()` to suppress the default context menu.
  - The modal must not appear when Phase 1's drag is in progress — add an `isDragging` guard.
  - Cerebrum Do-Not-Repeat 2026-04-10: use design-system tokens only. No hex color literals. Use `color-mix(in srgb, var(--token) N%, transparent)` instead of `rgba()`.
  - Modal close on Escape key and on backdrop click are table-stakes UX.
  - Focus trap inside the modal while open (a11y).
- **Prerequisites already in place:**
  - `MashedAssetInfo` JSON shape from `ListAllMashedAssets` Wails binding.
  - `ProcessSidebar.svelte` already iterates over mashed-asset rows with the drag handlers from Phase 1.

## Acceptance Criteria

**AC-1: Right-click opens the modal with the asset pre-loaded**
- Given a mashed-asset row in the sidebar
- When the user right-clicks the row
- Then the default context menu is suppressed
- And the `SkillEditorModal` opens with every form field pre-populated from the asset's current frontmatter

**AC-2: Escape and backdrop close the modal**
- Given the modal is open
- When the user presses Escape OR clicks the backdrop outside the modal card
- Then the modal closes without calling any save handler

**AC-3: Modal traps focus while open**
- Given the modal is open
- When the user presses Tab repeatedly
- Then focus cycles through the modal's form fields and buttons
- And Tab does not escape to background elements

**AC-4: Design-system tokens only**
- Given `SkillEditorModal.svelte`'s compiled CSS
- When inspected
- Then zero hex literals and zero `rgba(...)` calls appear
- And all colors resolve through `var(--…)`

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Skill editor modal shell

  Scenario: Right-click opens the modal
    Given a sidebar row for a command named "simplify"
    When the user right-clicks the row
    Then the browser context menu does not appear
    And the SkillEditorModal is visible
    And the name field contains "simplify"

  Scenario: Escape closes the modal
    Given the modal is open
    When the user presses Escape
    Then the modal is removed from the DOM
    And no save was attempted

  Scenario: Clicking the backdrop closes the modal
    Given the modal is open
    When the user clicks the backdrop outside the card
    Then the modal is removed from the DOM

  Scenario: Focus cycles inside the modal
    Given the modal is open and focused
    When the user Tabs through every focusable element
    Then focus never escapes the modal

  Scenario: Only design tokens in CSS
    Given the compiled stylesheet
    When inspected for color values
    Then every color uses var(--…)
```

## Tasks / Subtasks

- [ ] Task 1 — Build modal shell (AC-3, AC-4)
  - [ ] Create `SkillEditorModal.svelte` with the full form
  - [ ] Implement focus trap and Escape / backdrop close
  - [ ] Use design-system tokens exclusively
- [ ] Task 2 — Sidebar contextmenu handler (AC-1)
  - [ ] Add `on:contextmenu|preventDefault` to the mashed-asset row
  - [ ] Dispatch an `editAsset` event with the asset payload
- [ ] Task 3 — Host state in `WorkflowBuilder.svelte` (AC-1, AC-2)
  - [ ] Track `editingAsset | null`
  - [ ] Render the modal when non-null
  - [ ] Handle close → `editingAsset = null`
- [ ] Task 4 — Tests (AC-1, AC-2, AC-3, AC-4)
  - [ ] Playwright: right-click opens, Escape closes, backdrop closes, Tab traps focus
  - [ ] Lint/regex check for color literals

## Definition of Done

- [ ] All ACs verified by an automated test (Playwright; no "manually verified")
- [ ] Coverage ≥ 80% on modified files
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths, magic numbers, or color literals added
- [ ] Existing tests still pass

## Design Brief

### Layout composition
Modal card sits on the canonical `var(--overlay-backdrop)` scrim. Card width `520px` (wider than `NameWorkflowModal`'s 420px because this form has seven fields), `max-width: calc(100vw - 2 * var(--sp-xl))`. Vertical rhythm:

1. **Header** — icon + title + close button, `padding: var(--sp-lg) var(--sp-xl) var(--sp-md);` `border-bottom: 1px solid var(--border-subtle);`
2. **Form body** — scrollable region, `padding: var(--sp-lg) var(--sp-xl);` `max-height: calc(100vh - 240px);` `overflow-y: auto;`. Two-column grid for compact fields (`mashedRole`, `mashedCompletion`, `mashedChainable`, `mashedSessionPinned`) via `display: grid; grid-template-columns: 1fr 1fr; gap: var(--sp-md) var(--sp-lg);`. Full-width fields (`name`, `description`, `mashedInputs`, `mashedOutputs`) span both columns via `grid-column: 1 / -1;`.
3. **Footer** — action row, `padding: var(--sp-md) var(--sp-xl) var(--sp-lg);` `border-top: 1px solid var(--border-subtle);` `display: flex; justify-content: flex-end; gap: var(--sp-sm);`.

Each field group stacks: uppercase label → input → optional hint, with `margin-bottom: var(--sp-md)` between groups.

### Typography plan
- **Title**: `font-family: var(--font-ui)`, `font-size: var(--text-section)` (16px), `font-weight: 600`, `color: var(--text-primary)`, `letter-spacing: -0.01em`. Identical to `NameWorkflowModal`'s `h2` for modal consistency.
- **File path subtitle** (e.g. `.claude/skills/simplify/SKILL.md`): `font-family: var(--font-mono)`, `font-size: var(--text-label)` (11px), `color: var(--text-dim)`, `letter-spacing: 0`. Displayed directly under the title as monospace breadcrumb.
- **Field labels**: `font-family: var(--font-ui)`, `font-size: var(--text-label)` (11px), `font-weight: 600`, `text-transform: uppercase`, `letter-spacing: 0.05em`, `color: var(--text-dim)`. Matches `NameWorkflowModal`'s `label`.
- **Input text**: `font-family: var(--font-mono)`, `font-size: var(--text-body)` (13px), `color: var(--text-primary)`. Mono for inputs because this is frontmatter — the user is editing machine-readable fields and should see it as such.
- **Description textarea**: `font-family: var(--font-mono)`, `font-size: var(--text-body)`, `line-height: 1.5`, `min-height: 80px`.
- **Hint / error text**: `font-size: var(--text-label)`, `color: var(--text-muted)` (hint) / `color: var(--accent-red)` (error).
- **Buttons**: `font-family: var(--font-ui)`, `font-size: var(--text-body)`, `font-weight: 500`.

### Color strategy
- **Backdrop**: `var(--overlay-backdrop)` (rgba(0,0,0,0.6)) — the canonical scrim token, never override.
- **Card surface**: `var(--bg-surface)` — matches `NameWorkflowModal`.
- **Card border**: `1px solid var(--border-emphasis)` — emphasized so the card pops off the backdrop.
- **Card shadow**: `0 12px 40px color-mix(in srgb, var(--bg-deepest) 80%, transparent)` — exact reuse of `NameWorkflowModal`'s shadow. No new shadow token needed; matches existing modal precedent.
- **Input background**: `var(--bg-elevated)` — one step up from the card to create a nested surface.
- **Input border**: idle `1px solid var(--border-subtle)` → focus `var(--accent-amber)` (follows `NameWorkflowModal.name-input:focus` precedent — amber for "editable, uncommitted" state — NOT the green accent, which signals "running/active").
- **Select / dropdown chevron**: `color: var(--text-dim)`.
- **Checkbox checked**: `background: var(--accent-green)`, `border: var(--accent-green)` — checked === committed state.
- **Primary Save button**: `background: var(--accent-green); color: var(--bg-deepest); font-weight: 600;`. Green because Save IS the commit action; this modal is data-authoring so it earns brand green. (`NameWorkflowModal` uses amber because saving a name is NOT the main commit in that flow.)
- **Cancel button**: transparent background, `color: var(--text-dim)`, hover → `var(--bg-active)` bg + `var(--text-primary)`. Matches `NameWorkflowModal.btn-cancel`.
- **Error inline banner** (skills-editor-02 wires the content, but reserve the slot): `color-mix(in srgb, var(--accent-red) 12%, transparent)` background.

### Interaction model
- **Open**: fade backdrop + fly card from `y: 8px, opacity: 0` to rest over `var(--duration-medium)` (150ms) with `var(--ease-enter)` (ease-out). Focus jumps to the Name field via `onMount` → `inputEl.focus(); inputEl.select();` (matches `NameWorkflowModal`).
- **Close paths**: `Escape` key (via `svelte:window on:keydown`), backdrop click, Cancel button, close icon in header — all call the same `handleCancel` dispatcher.
- **Focus trap**: query all focusable descendants on mount, intercept `Tab` / `Shift+Tab` to wrap at edges.
- **Tab order**: Name → Description → mashedRole → mashedCompletion → mashedChainable → mashedSessionPinned → mashedInputs → mashedOutputs → Cancel → Save.
- **Enter**: only submits when focus is on the Save button or on a single-line input AND `canSave` is true (multiline textareas allow newline). Mirrors `NameWorkflowModal.handleKeydown` pattern.
- **Hover**: inputs shift `border-color` to `var(--border-emphasis)` over `var(--duration-short) var(--ease-enter)`. Buttons follow `.glow-btn`-style transitions already defined globally.
- **Focus-visible**: the global `button:focus-visible` / `input:focus-visible` rule in `style.css` already paints `2px solid var(--accent-green)` outline — do not override.
- **Drag guard**: host view passes `isDragging` prop; modal returns `null` (does not render) while true.

### Component specs
```
.overlay {
  position: fixed; inset: 0;
  background: var(--overlay-backdrop);
  display: flex; align-items: center; justify-content: center;
  z-index: 400;
}
.modal {
  width: 520px;
  max-width: calc(100vw - 2 * var(--sp-xl));
  background: var(--bg-surface);
  border: 1px solid var(--border-emphasis);
  border-radius: var(--radius-lg);         /* 8px — modal size */
  box-shadow: 0 12px 40px color-mix(in srgb, var(--bg-deepest) 80%, transparent);
  display: flex; flex-direction: column;
  max-height: calc(100vh - 2 * var(--sp-xl));
}
.modal-header { padding: var(--sp-lg) var(--sp-xl) var(--sp-md);
                 border-bottom: 1px solid var(--border-subtle); }
.modal-body   { padding: var(--sp-lg) var(--sp-xl); overflow-y: auto; flex: 1; }
.modal-footer { padding: var(--sp-md) var(--sp-xl) var(--sp-lg);
                 border-top: 1px solid var(--border-subtle);
                 display: flex; justify-content: flex-end; gap: var(--sp-sm); }
.field-grid   { display: grid; grid-template-columns: 1fr 1fr;
                 gap: var(--sp-md) var(--sp-lg); }
.field-full   { grid-column: 1 / -1; }
.input, .textarea, .select {
  width: 100%;
  padding: var(--sp-sm) var(--sp-md);
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-family: var(--font-mono);
  font-size: var(--text-body);
  transition: border-color var(--duration-short) var(--ease-enter);
}
.input:focus, .textarea:focus, .select:focus { border-color: var(--accent-amber); outline: none; }
.btn-save {
  background: var(--accent-green);
  color: var(--bg-deepest);
  font-weight: 600;
  padding: var(--sp-sm) var(--sp-lg);
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  transition: background var(--duration-short) var(--ease-enter);
}
```
No shadow token exists (`--shadow-*` missing) — the hand-rolled shadow above reuses the exact value from `NameWorkflowModal.svelte`. Recommend promoting to `--shadow-modal` once a third modal needs it.

### Signature elements
- **Mono inputs on a clean Geist form** — typing into mono fields on a Geist UI chrome makes the modal feel like "editing a config file in a proper editor", which is the product's core identity. No dev tool modal in this space treats frontmatter as a form.
- **File-path subtitle in mono** under the title is the distinctive mark: the user always sees the truth of what they're editing. This is a "Linear meets Bloomberg" move — path as first-class metadata.
- **Amber focus ring** on inputs — a Warp-influenced choice. Amber says "you're changing this, it's uncommitted". The Save button is green. The color journey amber→green IS the save flow.

### Distinguishing from existing patterns
- **vs `NameWorkflowModal.svelte`**: wider (520 vs 420), scrollable body, two-column field grid, mono file-path subtitle under title, Save button is `var(--accent-green)` not `var(--accent-amber)` (NameWorkflowModal saves a name; this modal commits data). Same shadow, same backdrop, same header-bottom-border grammar — they feel like siblings in the same system.
- **vs `QuestionResponseModal` / `AgentConfigModal`**: those are transient Q&A or configuration surfaces; this is an explicit file editor. The monospace file-path subtitle and mono inputs declare "you are editing a file", not "answer a prompt".
- **Cross-modal rule to codify**: all modals share `var(--overlay-backdrop)` backdrop, `var(--bg-surface)` card, `var(--border-emphasis)` card border, `var(--radius-lg)` radius, `0 12px 40px ...` shadow, and the 150ms ease-out entry. Future modals should inherit this stack.
