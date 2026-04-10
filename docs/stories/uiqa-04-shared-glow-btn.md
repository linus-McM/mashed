# Story uiqa-04: Shared `.glow-btn` Class

**Status:** ready
**Size:** M
**Priority:** P2
**Domain:** frontend
**Depends on:** uiqa-01, uiqa-02

## Description

Extract the duplicated "hot action button" pattern (currently `.action-hot` in NotificationFeed and `.git-hot` in AgentDetail — identical glow + green border + hover escalation, copy-pasted) into a single shared CSS class `.glow-btn` in `frontend/src/style.css`, using `var(--accent-green)` and `var(--glow-spread)` from uiqa-01. Component files then apply `class="glow-btn"` (or extend via `&.glow-btn`). This is the fix the review's executive summary calls out as one of the "three actions that would most improve the app" — it eliminates copy-paste drift and centralizes the mashed "neon hot action" identity into one place.

## Developer Notes

### Current state (review Section 3.1, Theme 2)

Both rules are identical except for class name:

- `NotificationFeed.svelte` lines 1700–1722 (`.action-hot`, `.action-review.action-hot`)
- `AgentDetail.svelte` lines 1004–1014 (`.git-hot`)

After uiqa-02 lands, both will reference `var(--accent-green)` instead of `#39ff14`. This story consolidates the rules.

### Target: add `.glow-btn` to `frontend/src/style.css`

Place it in the "Utility classes" section (after line 94). Suggested shape:

```css
/* Hot action button — shared neon glow */
.glow-btn {
  color: var(--accent-green);
  border: 1px solid var(--accent-green);
  background: color-mix(in srgb, var(--accent-green) 8%, transparent);
  text-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-green) 70%, transparent);
  box-shadow:
    var(--glow-spread) color-mix(in srgb, var(--accent-green) 35%, transparent),
    inset 0 0 0 1px color-mix(in srgb, var(--accent-green) 20%, transparent);
  transition:
    background var(--duration-short) var(--ease-enter),
    box-shadow var(--duration-short) var(--ease-enter);
}
.glow-btn:hover {
  background: color-mix(in srgb, var(--accent-green) 15%, transparent);
  box-shadow:
    var(--glow-spread) color-mix(in srgb, var(--accent-green) 55%, transparent),
    inset 0 0 0 1px color-mix(in srgb, var(--accent-green) 30%, transparent);
}
.glow-btn:focus-visible {
  outline: 2px solid var(--accent-green);
  outline-offset: 2px;
}
```

**Verify exact shadow/alpha values by reading the existing `.action-hot` rule first**, then mirror (don't invent) — the percentages above are approximations of the review's description. If the existing rule uses different alphas, preserve those to avoid a visual regression, then document the chosen values in the CSS comment.

### Call sites to update

- `frontend/src/views/NotificationFeed.svelte`
  - Delete the `.action-hot` rule at lines 1700–1722.
  - Find the element using `class="action-hot"` (likely in the new-session / review button template) and change `action-hot` to `glow-btn`. Keep any other classes already on the element.
  - If `.action-review.action-hot` has extra styles, move the extra bits into a scoped `.action-review.glow-btn` rule or a modifier class.

- `frontend/src/views/AgentDetail.svelte`
  - Delete the `.git-hot` rule at lines 1004–1014.
  - Update the element template to use `class="glow-btn"`.

### Risks & edge cases

- **Svelte scoped styles**: `.glow-btn` declared in `style.css` is global. The class name in component templates doesn't need `:global()` to apply. Verify by visually confirming after build.
- **Class collision**: `.glow-btn` is a new class name — grep the frontend to confirm it's unused before adding.
- **Selector specificity**: if the existing rules had `.action-hot.something-else` compound selectors for state variants, those must be converted to `.glow-btn.something-else`.
- **Visual regression**: this is a pure refactor — before/after screenshots should be pixel-identical (within anti-aliasing tolerance) if alphas are preserved correctly.

### Out of scope

- Any color changes beyond uiqa-02 (those should already be done).
- Adding `.glow-btn--red`, `.glow-btn--amber` variants — propose as a follow-up if needed.
- Converting other non-hot buttons to this pattern.

### Reference files

- `frontend/src/style.css` line 89–94 — utility classes section (insertion point)
- `frontend/src/views/NotificationFeed.svelte` lines 1700–1722 — source pattern
- `frontend/src/views/AgentDetail.svelte` lines 1004–1014 — duplicate pattern

Reference skills: `/simplify`, `/playwright-cli` for visual regression check.

## Acceptance Criteria

**AC-1: `.glow-btn` exists in `style.css`**
- Given `frontend/src/style.css`
- When parsed
- Then a rule `.glow-btn { ... }` exists
- And it uses `var(--accent-green)` for color and border
- And it references `var(--glow-spread)` in text-shadow or box-shadow
- And a `.glow-btn:hover` rule exists
- And a `.glow-btn:focus-visible` rule exists

**AC-2: `.action-hot` rule removed from NotificationFeed**
- Given `NotificationFeed.svelte`
- When parsed
- Then no CSS rule selector `.action-hot` exists in the `<style>` block
- And at least one element in the template uses `class="glow-btn"` (or a binding that includes `glow-btn`)

**AC-3: `.git-hot` rule removed from AgentDetail**
- Given `AgentDetail.svelte`
- When parsed
- Then no CSS rule selector `.git-hot` exists in the `<style>` block
- And at least one element in the template uses `class="glow-btn"` (or a binding that includes `glow-btn`)

**AC-4: Rendered hot button matches the accent**
- Given a mounted NotificationFeed with a hot action button visible
- When `getComputedStyle(button).color` is read
- Then it resolves to the computed value of `var(--accent-green)` in the current theme
- And `getComputedStyle(button).boxShadow` contains a non-empty value (glow present)

**AC-5: Theme override propagates to `.glow-btn`**
- Given a test overriding `--accent-green: #ff00ff` on `:root`
- When a `.glow-btn` is mounted
- Then its computed color is `rgb(255, 0, 255)`
- And its border-color is `rgb(255, 0, 255)`

**AC-6: No regression in total hot-button count**
- Given the pre-refactor count of elements that had `action-hot` or `git-hot` classes
- When the post-refactor template is rendered
- Then the count of elements with `glow-btn` class equals the pre-refactor count (same buttons, new class)

## BDD Test Scenarios

```gherkin
Feature: uiqa-04 shared glow-btn

  Scenario: .glow-btn declared in style.css
    Given style.css
    When parsed
    Then a rule matching "^\.glow-btn\s*\{" exists
    And the rule body references "var(--accent-green)" and "var(--glow-spread)"

  Scenario: .glow-btn:hover and :focus-visible defined
    Given style.css
    When parsed
    Then rules for ".glow-btn:hover" and ".glow-btn:focus-visible" exist

  Scenario: .action-hot fully removed from NotificationFeed
    Given NotificationFeed.svelte
    When searched for "\.action-hot"
    Then no matches exist in the <style> block
    And the template contains at least one element with class "glow-btn"

  Scenario: .git-hot fully removed from AgentDetail
    Given AgentDetail.svelte
    When searched for "\.git-hot"
    Then no matches exist in the <style> block
    And the template contains at least one element with class "glow-btn"

  Scenario: Computed color matches accent-green
    Given a test harness mounting <button class="glow-btn">Spawn</button>
    When the computed color is read
    Then it equals the resolved value of var(--accent-green)

  Scenario: Theme override propagates
    Given a test that sets --accent-green to #ff00ff on :root
    When a glow-btn element is mounted
    Then its computed color is rgb(255, 0, 255)
    And its computed border-color is rgb(255, 0, 255)

  Scenario: Focus ring visible on keyboard focus
    Given a mounted glow-btn
    When the button is focused via keyboard (Tab)
    Then getComputedStyle reports a non-zero outline-width
```

## Tasks / Subtasks

- [ ] Task 1: Add `.glow-btn` to style.css (AC-1)
  - [ ] Subtask 1a: Read existing `.action-hot` rule at NotificationFeed 1700–1722 to capture exact alpha values
  - [ ] Subtask 1b: Insert `.glow-btn`, `.glow-btn:hover`, `.glow-btn:focus-visible` after line 94 in `style.css`
  - [ ] Subtask 1c: Grep to confirm `.glow-btn` is not already defined elsewhere

- [ ] Task 2: Replace `.action-hot` in NotificationFeed (AC-2, AC-6)
  - [ ] Subtask 2a: Count occurrences of `class="action-hot"` and `action-hot` in `class:` bindings before editing
  - [ ] Subtask 2b: Replace template class references with `glow-btn`
  - [ ] Subtask 2c: Delete `.action-hot` CSS rule
  - [ ] Subtask 2d: Port any `.action-review.action-hot` compound rule to `.action-review.glow-btn`

- [ ] Task 3: Replace `.git-hot` in AgentDetail (AC-3, AC-6)
  - [ ] Subtask 3a: Count occurrences before editing
  - [ ] Subtask 3b: Replace template class references with `glow-btn`
  - [ ] Subtask 3c: Delete `.git-hot` CSS rule

- [ ] Task 4: Write component tests (AC-4, AC-5)
  - [ ] Subtask 4a: Minimal test mounting `<button class="glow-btn">` and asserting computed color
  - [ ] Subtask 4b: Theme-override test for `--accent-green`
  - [ ] Subtask 4c: Focus-ring test

- [ ] Task 5: Visual regression check via Playwright CLI
  - [ ] Subtask 5a: Capture before/after screenshots of NotificationFeed and AgentDetail
  - [ ] Subtask 5b: Diff the hot buttons specifically

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
- [ ] Hot button count pre/post refactor matches
- [ ] Story status updated to `done`
