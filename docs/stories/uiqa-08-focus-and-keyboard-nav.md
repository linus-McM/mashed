# Story uiqa-08: Focus States + CanvasPane Keyboard Nav

**Status:** ready
**Size:** M
**Priority:** P3
**Domain:** frontend
**Depends on:** uiqa-01, uiqa-04

## Description

Add `:focus-visible` outline rings to every interactive element that currently lacks one (agent rows, action buttons, modal inputs, repo group headers) and implement full keyboard navigation for the CanvasPane context menu in the BMAD workflow builder: Escape to close, arrow keys to cycle items, Enter to activate. This is the accessibility polish that takes the app from "mouse-only power tool" to "keyboard-first power tool" — critical for the audience that actually uses mashed.

## Developer Notes

### Scope 1: `:focus-visible` rings (FIX-16)

The review notes "currently only some elements have focus styles." Audit and fix:

**Target elements:**
- Agent rows in `NotificationFeed.svelte` (the `.agent-row` element)
- Action buttons in NotificationFeed (`.action-*` buttons — including the new `.glow-btn` from uiqa-04, which should already be covered)
- Modal inputs in NewSessionModal.svelte, SpawnAgent.svelte (all `<input>`, `<select>`, `<textarea>`, `<button>` elements)
- Repo group headers in NotificationFeed (the clickable `.repo-header`)
- Tab close buttons in AgentDetail
- Theme thumbnails in Settings (currently may use `:hover` only)
- ExecutionBar control buttons
- CanvasPane sidebar template items

**Standard focus ring style (add to `style.css` as a utility):**
```css
/* Focus visible — keyboard navigation ring */
.focus-ring:focus-visible,
button:focus-visible,
[role="button"]:focus-visible,
input:focus-visible,
select:focus-visible,
textarea:focus-visible,
.agent-row:focus-visible,
.repo-header:focus-visible {
  outline: 2px solid var(--accent-green);
  outline-offset: 2px;
  border-radius: var(--radius-md);
}
```

**Decision:** add as a global selector in `style.css` rather than per-component. This catches every element with no per-component overrides. Scoped component rules can opt out by declaring their own `:focus-visible`.

### Scope 2: CanvasPane context menu keyboard nav (FIX-17 / WB-04)

**File:** `frontend/src/components/bmad/CanvasPane.svelte`

**Requirements:**
- Context menu opens via right-click (existing).
- When menu is open:
  - `Escape` → close menu, return focus to previously focused element
  - `ArrowDown` → move selection to next menu item (wrap from last to first)
  - `ArrowUp` → previous item (wrap)
  - `Home` → first item
  - `End` → last item
  - `Enter` / `Space` → activate selected item, close menu
  - Typing a letter → jump to first item starting with that letter (nice-to-have; optional)
- Menu items should have `tabindex="-1"` and be focused programmatically (not via tab order).
- `role="menu"` on container, `role="menuitem"` on items.
- `aria-activedescendant` pattern or `.focus()` on items — pick the simpler one (direct `.focus()` is easier).

**Implementation sketch:**
```svelte
<script>
  let menuRef;
  let itemRefs = [];
  let activeIdx = 0;

  function handleKeydown(e) {
    if (!menuOpen) return;
    switch (e.key) {
      case 'Escape':
        closeMenu();
        e.preventDefault();
        break;
      case 'ArrowDown':
        activeIdx = (activeIdx + 1) % items.length;
        itemRefs[activeIdx]?.focus();
        e.preventDefault();
        break;
      case 'ArrowUp':
        activeIdx = (activeIdx - 1 + items.length) % items.length;
        itemRefs[activeIdx]?.focus();
        e.preventDefault();
        break;
      case 'Home':
        activeIdx = 0;
        itemRefs[0]?.focus();
        e.preventDefault();
        break;
      case 'End':
        activeIdx = items.length - 1;
        itemRefs[activeIdx]?.focus();
        e.preventDefault();
        break;
      case 'Enter':
      case ' ':
        activateItem(activeIdx);
        e.preventDefault();
        break;
    }
  }
</script>

<svelte:window on:keydown={handleKeydown} />
```

### Risks & edge cases

- **Focus trap leakage**: if user Tabs inside the context menu, focus can escape to the canvas. Prevent by listening on `svelte:window` and ignoring tab when menu open, OR by trapping focus with a sentinel element before/after. Simpler: let Tab close the menu (common desktop pattern).
- **Screen reader support**: `role="menu"` + `role="menuitem"` is the W3C menubar pattern. Verify with VoiceOver if possible.
- **Double-focus ring**: if a component already has a `:focus` rule (without `-visible`), the new `:focus-visible` rule may layer on top. Replace legacy `:focus` selectors when found.
- **Outline on `border-radius` elements**: `outline-offset: 2px` works but can clip on rounded corners depending on browser. `border-radius` on the outline rule itself is a Chromium-specific hint; keep for forward compat.

### Out of scope

- Focus management for modal open/close (NewSessionModal already handles Escape; verify but don't rewrite).
- Tab order reordering.
- Skip links.
- High-contrast mode variants.
- Keyboard shortcuts beyond the context menu.

### Reference files

- `frontend/src/style.css` utility section — insertion point for global focus-ring
- `frontend/src/components/bmad/CanvasPane.svelte` — context menu implementation
- Existing keyboard handling in NewSessionModal / SpawnAgent — pattern source

Reference skills: `/simplify`, `/playwright-cli`.

## Acceptance Criteria

**AC-1: Global `:focus-visible` rule present**
- Given `frontend/src/style.css`
- When parsed
- Then a rule covering `button:focus-visible, input:focus-visible, select:focus-visible, textarea:focus-visible, [role="button"]:focus-visible, .agent-row:focus-visible, .repo-header:focus-visible` exists
- And it declares `outline: 2px solid var(--accent-green)` and `outline-offset: 2px`

**AC-2: Every actionable element in NotificationFeed shows a focus ring on Tab**
- Given a Playwright test that tabs through NotificationFeed
- When each interactive element is focused
- Then `getComputedStyle(el).outlineWidth` is at least `2px`
- And `getComputedStyle(el).outlineColor` matches `var(--accent-green)` in the current theme

**AC-3: Modal inputs show focus rings**
- Given NewSessionModal and SpawnAgent are open
- When each input/select/textarea/button is focused
- Then a visible outline is present

**AC-4: CanvasPane context menu closes on Escape**
- Given the BMAD canvas is visible
- When the user right-clicks to open the context menu
- And presses `Escape`
- Then the menu closes within one frame
- And focus returns to the element that had focus before the menu opened

**AC-5: Arrow keys cycle menu items with wrap**
- Given the context menu is open with N items
- When the user presses `ArrowDown` (N+1) times
- Then after N presses, the last item has focus
- And after N+1 presses, the first item has focus (wrap)
- And `ArrowUp` cycles in reverse

**AC-6: Enter activates selected item**
- Given the context menu is open and item 2 is focused
- When the user presses `Enter`
- Then the same action runs as clicking item 2
- And the menu closes

**AC-7: `role="menu"` and `role="menuitem"` set**
- Given CanvasPane.svelte
- When the context menu template is parsed
- Then the container has `role="menu"`
- And each item has `role="menuitem"` and `tabindex="-1"`

## BDD Test Scenarios

```gherkin
Feature: uiqa-08 focus states and keyboard nav

  Scenario: Global focus-visible rule present in style.css
    Given style.css
    When parsed
    Then a rule includes "button:focus-visible" and "outline: 2px solid var(--accent-green)"

  Scenario: Tab through NotificationFeed shows rings
    Given a Playwright test loading NotificationFeed with 2 repos and 3 agents each
    When the user presses Tab 10 times
    Then each focused element has a computed outlineWidth >= 2px

  Scenario: NewSessionModal inputs focusable with ring
    Given NewSessionModal open
    When Tab is pressed
    Then every focusable field shows a visible ring

  Scenario: Context menu closes on Escape
    Given BMAD CanvasPane loaded
    When the user right-clicks an empty canvas area
    And then presses Escape
    Then the menu is removed from the DOM
    And focus returns to the canvas

  Scenario: Arrow down cycles items
    Given the context menu is open with 4 items
    When the user presses ArrowDown once
    Then item index 1 has focus (item index 0 was default)
    When the user presses ArrowDown 4 more times
    Then item index 1 has focus again (wrapped around)

  Scenario: Enter activates selected
    Given the context menu is open and item at index 2 is focused
    When the user presses Enter
    Then the handler for item 2 is invoked
    And the menu is closed

  Scenario: Menu has ARIA roles
    Given CanvasPane.svelte
    When parsed
    Then role="menu" appears on the container
    And role="menuitem" appears on at least one item

  Scenario: Tab inside menu closes it
    Given the context menu is open
    When the user presses Tab
    Then the menu closes
```

## Tasks / Subtasks

- [ ] Task 1: Global focus-visible rule in style.css (AC-1)
  - [ ] Subtask 1a: Insert combined selector + rule after existing utility classes
  - [ ] Subtask 1b: Verify no existing `:focus` (without -visible) rules conflict

- [ ] Task 2: Add focus targets where tabindex needed (AC-2)
  - [ ] Subtask 2a: Audit NotificationFeed `.agent-row` — add `tabindex="0"` if it's a div
  - [ ] Subtask 2b: Same for `.repo-header` if clickable div
  - [ ] Subtask 2c: Replace click-only divs with `<button>` where semantically correct

- [ ] Task 3: Playwright tab-through tests (AC-2, AC-3)
  - [ ] Subtask 3a: Test NotificationFeed tab cycle
  - [ ] Subtask 3b: Test NewSessionModal field cycle
  - [ ] Subtask 3c: Test SpawnAgent field cycle

- [ ] Task 4: CanvasPane context menu keyboard handlers (AC-4, AC-5, AC-6)
  - [ ] Subtask 4a: Add svelte:window keydown listener gated on menuOpen
  - [ ] Subtask 4b: Implement Escape / ArrowUp / ArrowDown / Home / End / Enter / Space
  - [ ] Subtask 4c: Focus management via itemRefs[activeIdx]?.focus()
  - [ ] Subtask 4d: Return focus to previous element on close

- [ ] Task 5: ARIA roles on context menu (AC-7)
  - [ ] Subtask 5a: role="menu" on container
  - [ ] Subtask 5b: role="menuitem" and tabindex="-1" on each item

- [ ] Task 6: Playwright tests for menu nav (AC-4, AC-5, AC-6, AC-7)
  - [ ] Subtask 6a: Escape closes menu
  - [ ] Subtask 6b: Arrow keys cycle with wrap
  - [ ] Subtask 6c: Enter activates
  - [ ] Subtask 6d: Tab closes menu

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] Frontend build passes
- [ ] `/simplify` run on all modified files
- [ ] No new `#39ff14` introduced anywhere
- [ ] No orphaned token-scale values introduced in modified files
- [ ] Story status updated to `done`
