# Story 05: Markdown Editor Panel in Settings

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** Story 02 (store), Story 04 (panel grid layout)
**Status:** done
**UI-facing:** YES — flag for ui-architect design brief

## Description

Add the "Markdown Editor" settings panel to Column 2 of `Settings.svelte`. The panel lists six toggle rows (Bold, Italic, Strikethrough, Code, Link, LaTeX) bound to the `markdownMenuSettings` store; clicking a toggle calls `updateMarkdownMenuItem`. Clicking Back or pressing Esc calls `clearMarkdownMenuDirty` before dispatching the `back` event — this is the handoff that lets `MarkdownEditor.svelte` (story 06) know it's safe to re-init.

## Developer Notes

### Architecture
- File to modify: `/Users/linus/Development/mashed/frontend/src/views/Settings.svelte`
- Import additions in `<script>`:
  ```ts
  import { markdownMenuSettings, updateMarkdownMenuItem, clearMarkdownMenuDirty }
    from '../lib/stores/markdownMenuSettings';
  ```
- Constant array at top of `<script>`:
  ```ts
  const TOOLBAR_ITEMS = [
    { key: 'bold',          label: 'Bold' },
    { key: 'italic',        label: 'Italic' },
    { key: 'strikethrough', label: 'Strikethrough' },
    { key: 'code',          label: 'Code' },
    { key: 'link',          label: 'Link' },
    { key: 'latex',         label: 'LaTeX' },
  ] as const;
  ```
- Panel markup (belongs in column 2, after Theme Extensions — see Story 04 for layout):

```svelte
<div class="settings-panel">
  <section class="settings-section" data-testid="markdown-menu-section">
    <h2 class="section-title">Markdown Editor</h2>
    <p class="section-desc">
      Selection toolbar items. Changes apply when you close Settings.
    </p>

    {#each TOOLBAR_ITEMS as item}
      <div class="setting-row">
        <span class="setting-label">{item.label}</span>
        <button
          class="setting-toggle"
          class:active={$markdownMenuSettings[item.key]}
          data-testid={`toolbar-toggle-${item.key}`}
          on:click={() => updateMarkdownMenuItem(item.key, !$markdownMenuSettings[item.key])}
        >
          {$markdownMenuSettings[item.key] ? 'On' : 'Off'}
        </button>
      </div>
    {/each}
  </section>
</div>
```

- Back button + Esc key paths BOTH must call `clearMarkdownMenuDirty()` before `dispatch('back')`:

```svelte
<button class="back-btn" on:click={() => { clearMarkdownMenuDirty(); dispatch('back'); }}>
  <ArrowLeft size={14} /> Back
</button>
```

```ts
function handleKeydown(e) {
  if (e.key === 'Escape') {
    clearMarkdownMenuDirty();
    dispatch('back');
  }
}
```

### Behavior contract
- Clicking a toggle calls `updateMarkdownMenuItem(key, !current)` — that function already updates the store, flips `markdownMenuDirty` to true, and persists via Wails.
- While the user is inside Settings with a dirty toggle, the underlying markdown editor MUST NOT re-init. That guarantee is enforced in story 06 via the `markdownMenuDirty` store subscription. This story's job is to ensure `clearMarkdownMenuDirty` fires before the Settings view unmounts.

### Copy (label text, final)
The six labels are exactly: `Bold`, `Italic`, `Strikethrough`, `Code`, `Link`, `LaTeX`. The section title is `Markdown Editor`. The section description is `Selection toolbar items. Changes apply when you close Settings.` — this wording is user-facing; do not change without ui-architect sign-off.

### Styling
No new CSS. Reuse existing `.setting-row`, `.setting-label`, `.setting-toggle.active` styles from Settings.svelte. These are the same classes the Editor panel's toggles use — visual parity is intentional.

### Risks & Edge Cases
- **User toggles then closes the app without clicking Back** — `clearMarkdownMenuDirty` doesn't fire. On next launch the dirty flag is re-initialised to false by the store module, so no issue. Settings are already persisted in `updateMarkdownMenuItem`.
- **Rapid toggle (double-click)** — both clicks call `updateMarkdownMenuItem`. Second click toggles back the other way; net behavior is whatever the final state is. Wails persist calls are serialised server-side.
- **Focus management on keyboard navigation** — six buttons in a row need tab order to go top-to-bottom. Default DOM order handles this; no custom tabindex required.
- **Accessibility** — each toggle is a `<button>`, which is focusable and has implicit role `button`. The `On`/`Off` text inside is accessible. Optionally add `aria-pressed={$markdownMenuSettings[item.key]}` for screen readers. Not strictly required for MVP but recommended.

### Reference Files
- `/Users/linus/Development/mashed/frontend/src/views/Settings.svelte` — the file to modify (Editor panel has similar toggle rows to mirror).
- `/Users/linus/Development/mashed/frontend/src/lib/stores/markdownMenuSettings.ts` — the store imports.

## Acceptance Criteria

AC-1: Panel renders in column 2 with six toggle rows
- Given the Settings view is opened after Story 04
- When I look at column 2
- Then a `.settings-panel` with `data-testid="markdown-menu-section"` is present
- And it contains exactly six `setting-row` entries with labels Bold, Italic, Strikethrough, Code, Link, LaTeX in that order

AC-2: Toggle reflects current store value
- Given `markdownMenuSettings` has `bold:true, latex:false`
- When the panel renders
- Then the `toolbar-toggle-bold` button has the `active` class and displays "On"
- And the `toolbar-toggle-latex` button does NOT have `active` and displays "Off"

AC-3: Clicking a toggle calls updateMarkdownMenuItem
- Given `markdownMenuSettings.bold === true`
- When I click `toolbar-toggle-bold`
- Then `updateMarkdownMenuItem('bold', false)` is called exactly once
- And the store subsequently reflects `bold:false`
- And the button label flips to "Off"

AC-4: Back button clears dirty flag
- Given `markdownMenuDirty === true`
- When I click the Back button
- Then `clearMarkdownMenuDirty()` is called BEFORE the `back` event is dispatched
- And `markdownMenuDirty` reads `false` after the click

AC-5: Escape key clears dirty flag
- Given `markdownMenuDirty === true`
- When I press Escape anywhere in Settings
- Then `clearMarkdownMenuDirty()` is called BEFORE the `back` event fires
- And `markdownMenuDirty` reads `false`

AC-6: Section copy matches spec
- Given the panel is rendered
- When the text is read
- Then the heading is "Markdown Editor"
- And the description is "Selection toolbar items. Changes apply when you close Settings."

## BDD Test Scenarios

```gherkin
Feature: Markdown Editor settings panel

  Scenario: Panel renders six toggle rows
    Given Settings is open with default store values
    When the Markdown Editor panel is rendered
    Then there are 6 rows with labels Bold, Italic, Strikethrough, Code, Link, LaTeX
    And Bold, Italic, Strikethrough, Code, Link buttons show "On"
    And LaTeX button shows "Off"

  Scenario: Toggle off persists via store
    Given the store has bold=true
    When I click the Bold toggle button
    Then updateMarkdownMenuItem is invoked with ("bold", false)
    And after the update the button displays "Off" and lacks the active class

  Scenario: Toggle on flips back
    Given I just toggled Italic off
    When I click Italic again
    Then updateMarkdownMenuItem is called with ("italic", true)
    And the button displays "On"

  Scenario: Back button clears dirty
    Given markdownMenuDirty is true (because I just toggled something)
    When I click the Back button
    Then clearMarkdownMenuDirty fires first
    And a "back" event is dispatched second
    And markdownMenuDirty is now false

  Scenario: Escape clears dirty
    Given markdownMenuDirty is true
    When Escape is pressed
    Then clearMarkdownMenuDirty fires
    And a "back" event is dispatched

  Scenario: Section copy is exact
    Given the panel is rendered
    Then the section title reads "Markdown Editor"
    And the description reads "Selection toolbar items. Changes apply when you close Settings."
```

## Tasks / Subtasks

- [ ] Task 1: Script imports and constants (AC-1, AC-2) — frontend
  - [ ] Add store imports to `<script>` block in Settings.svelte.
  - [ ] Declare `TOOLBAR_ITEMS` const.
- [ ] Task 2: Panel markup (AC-1, AC-2, AC-3, AC-6) — frontend
  - [ ] Insert the `.settings-panel` block in column 2 after Theme Extensions.
  - [ ] Use exact heading and description copy.
  - [ ] Ensure `data-testid="markdown-menu-section"` and per-toggle `data-testid` attributes are present.
- [ ] Task 3: Dirty-clear on Back and Escape (AC-4, AC-5) — frontend
  - [ ] Wrap the Back button's on:click to call `clearMarkdownMenuDirty` before `dispatch('back')`.
  - [ ] Update `handleKeydown` for Escape to call `clearMarkdownMenuDirty` before `dispatch('back')`.
- [ ] Task 4: Tests (all ACs) — frontend
  - [ ] Component render test with testing-library-svelte: mount Settings, assert six rows, assert labels and On/Off states.
  - [ ] Click test: click bold toggle, assert `updateMarkdownMenuItem` called with ('bold', false) (mock the store).
  - [ ] Back-click test: set dirty=true via direct store write, click Back, assert `clearMarkdownMenuDirty` runs first and a 'back' event is dispatched.
  - [ ] Escape test: analogous to Back-click but via keyboard event.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ line coverage on the new panel-specific code (markup is hard to cover; test the behavioral handlers)
- [ ] `svelte-check` passes with 0 warnings, 0 errors
- [ ] `/simplify` run on Settings.svelte
- [ ] ui-architect design brief reviewed (labels, density, card alignment with other panels)
- [ ] Code review: no CRITICAL/HIGH issues

## Design Brief

### Intent
The Markdown Editor panel is a six-switch control surface — the simplest possible expression of "which chrome do you want on the selection toolbar?" It must feel identical in weight and rhythm to the Editor panel's toggles above it (Minimap, Insert Spaces, Bracket Colors, etc.). Visual parity with the Editor panel is the whole point: the user already knows how those toggles work, so these just extend the vocabulary. Zero bespoke styling.

### 1. Layout composition
- **Host container:** the story-04 `.settings-panel` card in `.settings-col-2`, below Theme Extensions. No grid; a vertical flex stack inside the panel.
- **Panel internal rhythm:**
  - `.section-title` (`Markdown Editor`) — mono, `var(--text-data)`, `margin-bottom: var(--sp-md)` (existing rule).
  - `.section-desc` (`Selection toolbar items. Changes apply when you close Settings.`) — `margin-bottom: var(--sp-md)` (existing rule).
  - Six `.setting-row` blocks, stacked. Each row: `display: flex`, `justify-content: space-between`, `align-items: center`, `padding: var(--sp-xs) 0`, `gap: var(--sp-md)` (all existing — unchanged).
- **Label ↔ toggle balance:**
  - Label hugs left edge of the panel padding; toggle hugs right edge. `justify-content: space-between` does the work — no fixed column widths, labels of any length right-align the toggle cleanly.
  - Toggles have `min-width: 48px` (existing rule). Six stacked 48px-wide buttons create a clean vertical rail on the right edge of the panel — this IS the signature visual of the panel.
- **Row count:** exactly 6 rows, fixed order: Bold → Italic → Strikethrough → Code → Link → LaTeX. No collapse, no scroll, no "more" button — the full set is always visible.
- **Estimated panel height:** `var(--sp-lg)` top padding + title (14px + 12px margin) + desc (13px line + 12px margin) + 6 × (4px + ~24px + 4px) ≈ ~240–260px. Balances reasonably against Theme Extensions above it without overflow.

### 2. Typography plan
- **Title `Markdown Editor`** — `.section-title` existing rules: `var(--font-mono)`, `var(--text-data)` (14px), weight 600, `color: var(--text-primary)`. Unchanged.
- **Description** — `.section-desc`: `var(--text-body)` (13px), `color: var(--text-dim)`. Unchanged.
- **Row label `Bold`/`Italic`/etc.** — `.setting-label`: `var(--font-mono)`, `var(--text-body)` (13px), `color: var(--text-dim)`, `white-space: nowrap`. Unchanged.
- **Toggle text `On`/`Off`** — `.setting-toggle`: `var(--font-mono)`, `font-size: 11px` (literal, consistent with sibling toggles), center-aligned. Unchanged.
- **Typography rule:** use the exact classes, no new font-size values, no new font-family declarations. Parity with the Editor panel's Minimap row is the acceptance criterion.

### 3. Color strategy
Reuse existing `.setting-toggle` states verbatim — no delta.
- **Inactive (Off)** — `background: var(--bg-surface)` / `border: 1px solid var(--border-subtle)` / `color: var(--text-muted)` (`#2e3d4d`, heavily dimmed — "this switch is resting").
- **Hover (Off)** — `border-color: var(--border-emphasis)` / `color: var(--text-dim)` (lifts slightly from muted to dim — subtle acknowledgement).
- **Active (On)** — `background: var(--bg-active)` (`#181c23`) / `border-color: var(--accent-green)` / `color: var(--accent-green)`. The neon edge is where the brand color earns its keep — six stacked "On" buttons should read as a green rail of active state.
- **Focus-visible** — inherits global rule (`outline: 2px solid var(--accent-green); outline-offset: 2px`). No panel-local override needed.
- **No new tokens.** No custom colors, no opacity variants, no color-mix. Every state already exists.

### 4. Interaction model
- **Tab order** — six buttons follow DOM order. Tab from Theme Extensions' last control → Bold → Italic → Strikethrough → Code → Link → LaTeX → Back button. Escape anywhere triggers the clean-dirty + back dispatch. No `tabindex` manipulation.
- **Activation** — `Space` and `Enter` both fire the native `<button>` click (Svelte `on:click`). No manual keydown handler on the toggle itself.
- **Hover** — existing `.setting-toggle:hover` rule covers it. Mouse users see the border lift on approach.
- **Transition** — existing `transition: all 100ms ease` on `.setting-toggle`. `100ms` matches DESIGN.md's `--duration-short` — instant but legible. Do NOT upgrade to 150ms; that would break parity with Minimap/Insert Spaces toggles.
- **No double-click protection needed** — rapid toggling just flips state back; the store is idempotent.
- **No loading/saving indicator** — `updateMarkdownMenuItem` persists via Wails in the background; UX contract is "optimistic immediate flip," same as Editor panel toggles.
- **Recommended a11y add** — `aria-pressed={$markdownMenuSettings[item.key]}` on each toggle. The story marks this optional; the design brief upgrades it to recommended because the only textual signal (`On`/`Off`) is not screen-reader-friendly without it. This is a 1-attribute addition, not visual.
- **No extra confirmation** — toggling directly mutates the store. The `markdownMenuDirty` flag is for the editor re-init deferral (story 06), not for user-facing "unsaved changes" UX.

### 5. Component specs
**Zero new CSS.** Entire delta is markup. The panel reuses the `.settings-panel` wrapper from story 04 and the existing toggle styles:

```svelte
<div class="settings-panel">
  <section class="settings-section" data-testid="markdown-menu-section">
    <h2 class="section-title">Markdown Editor</h2>
    <p class="section-desc">
      Selection toolbar items. Changes apply when you close Settings.
    </p>

    {#each TOOLBAR_ITEMS as item}
      <div class="setting-row">
        <span class="setting-label">{item.label}</span>
        <button
          class="setting-toggle"
          class:active={$markdownMenuSettings[item.key]}
          aria-pressed={$markdownMenuSettings[item.key]}
          data-testid={`toolbar-toggle-${item.key}`}
          on:click={() => updateMarkdownMenuItem(item.key, !$markdownMenuSettings[item.key])}
        >
          {$markdownMenuSettings[item.key] ? 'On' : 'Off'}
        </button>
      </div>
    {/each}
  </section>
</div>
```

**Token reference table (all existing, no new values):**

| Element        | Property         | Token                   | Resolved |
| -------------- | ---------------- | ----------------------- | -------- |
| Panel          | `background`     | `var(--bg-surface)`     | `#0d0f12` |
| Panel          | `border`         | `1px solid var(--border-subtle)` | `#1e2530` |
| Panel          | `border-radius` | `var(--radius-md)`      | `4px` |
| Panel          | `padding`        | `var(--sp-lg)`          | `16px` |
| Row            | `padding`        | `var(--sp-xs) 0`        | `4px 0` |
| Row            | `gap`            | `var(--sp-md)`          | `12px` |
| Label          | `font-size`      | `var(--text-body)`      | `13px` |
| Label          | `color`          | `var(--text-dim)`       | `#4a5a6a` |
| Toggle         | `padding`        | `4px 12px`              | (existing literal) |
| Toggle         | `border-radius` | `var(--radius-sm)`      | `2px` |
| Toggle         | `min-width`      | `48px`                  | (existing literal) |
| Toggle on hover| `border-color`   | `var(--border-emphasis)` | `#2a3340` |
| Toggle active  | `background`     | `var(--bg-active)`      | `#181c23` |
| Toggle active  | `border-color`   | `var(--accent-green)`   | `#00e57a` |
| Toggle active  | `color`          | `var(--accent-green)`   | `#00e57a` |
| Transition     | `duration`       | `100ms ease`            | matches `--duration-short` |

**Delta from existing `.setting-toggle`: zero.** If you're writing any new CSS selector for this panel, stop and re-read this row of the brief.

### 6. Signature elements
- **Visual twin of the Editor panel above.** The six toggles should be indistinguishable in style from Minimap / Insert Spaces / Bracket Colors. That parity is the design — it communicates "same kind of setting, just for markdown." A user scanning the column shouldn't notice a panel boundary in terms of control language.
- **Green rail of active toggles.** Defaults have Bold/Italic/Strike/Code/Link ON, LaTeX OFF. That gives five green-bordered buttons stacked on the panel's right edge — a vertical green bar is the panel's passive signature. Remove one, and the gap is immediately readable. This is density-as-affordance, DESIGN.md style.
- **Mono-face labels left, mono-face On/Off right.** Same typeface on both sides makes the row read as a terminal `key=value` line. The label is `--text-dim`; the value (when active) is `--accent-green`. Instantly scannable.
- **No icons in labels.** The Editor panel's toggle rows use text labels, not icons. Keep that discipline here — "Bold"/"Italic" as text, not B/I glyphs. Icons would violate parity and add noise.
- **4-space 2px radius on the toggle itself** (`--radius-sm`) — a sharper corner than the panel's own `--radius-md` (4px). Nested radii stepping down is the Bloomberg-terminal idiom; it reads as precise.

### 7. Anti-patterns to avoid
- **Do not write new CSS.** If you find yourself adding a `.markdown-toggle` or `.toolbar-toggle-row` class, stop. The story explicitly states "No new CSS. Reuse existing `.setting-row`, `.setting-label`, `.setting-toggle.active` styles."
- **Do not introduce icons.** No `<Bold size={14} />` from lucide-svelte for the labels. Text labels keep parity with Editor panel.
- **Do not swap On/Off for a physical switch UI** (sliding knob). That's a different design language — this panel is text toggles, matching siblings.
- **Do not group rows into subsections** (e.g., "Inline formatting" / "Links & math"). Six items is below the threshold for subheads; grouping would out-weigh the content.
- **Do not animate the toggle flip with anything > 100ms.** Keep parity. No spring curves, no slide-knob animation.
- **Do not reorder the six items.** Order is Bold → Italic → Strikethrough → Code → Link → LaTeX, full stop. Order is AC-1.
- **Do not localise copy in this story.** Heading text `Markdown Editor` and description `Selection toolbar items. Changes apply when you close Settings.` are locked to the story spec.
- **Do not emit a "Save" button.** Toggles persist on click. A save button would contradict the deferred-re-init design of story 06.
- **Do not introduce a confirmation modal for toggling off.** There's nothing destructive here — each toggle is reversible.
- **Do not place the panel in column 1.** It belongs in `.settings-col-2` per story 04's layout contract.
