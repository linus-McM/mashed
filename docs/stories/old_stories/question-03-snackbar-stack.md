# Story 3: Global Question Snackbar Stack

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** Story 1 (question-01-backend-detection)
**Status:** done

## Description

Add a global snackbar notification stack that surfaces BMAD process questions to the user regardless of which view they are on. The stack appears in the top-right corner, shows one card per pending question (repo name, truncated question text, timestamp), and clicking a snackbar navigates the user to the workflows view for that repo. The stack is mounted in `App.svelte` and driven by `bmad:node:question` / `bmad:node:question:dismissed` Wails events.

## Developer Notes

### Architecture

- **New file:** `frontend/src/components/bmad/QuestionSnackbarStack.svelte` -- the global overlay component.
- **Modified file:** `frontend/src/App.svelte` -- mount `QuestionSnackbarStack`, add event listeners for `bmad:node:question` and `bmad:node:question:dismissed`, manage `questionQueue` array state, handle snackbar click navigation.

### Component Design: `QuestionSnackbarStack.svelte`

```svelte
<script>
  export let questions = [];  // Array of QuestionEvent objects
  import { createEventDispatcher } from 'svelte';
  import { fly } from 'svelte/transition';
  import { MessageCircleQuestion } from 'lucide-svelte';
  const dispatch = createEventDispatcher();
  // dispatch('navigate', { repoPath, question })
  // dispatch('dismiss', { questionId })
</script>
```

- **Positioning:** `position: fixed; top: 16px; right: 16px; z-index: 300;` -- above all other UI but below modals (which use z-index 400+).
- **Each snackbar card:**
  - Left border: 3px solid, color from `localStorage.getItem('mashed:repoBorderColors')` parsed as JSON, keyed by `question.repoName`. Default to `#1e2530` if not found.
  - Background: `var(--surface-2)`
  - Text: `var(--text-primary)` for repo name, `var(--text-secondary)` for truncated question
  - Question text truncated to 80 characters with ellipsis
  - Timestamp shown as relative time (e.g., "2m ago") -- use simple helper, no external dependency
  - Icon: `MessageCircleQuestion` from lucide-svelte (size 16)
  - Cursor: pointer; on click dispatches `navigate` event
- **Transitions:** Svelte `fly` transition from right (`x: 300, duration: 300`) on each card.
- **Stacking:** Vertical flex column with `8px` gap. Newest on top (array sorted by timestamp descending for display, or just prepend new items).
- **Max visible:** Show at most 5 snackbars. If more, show a "+N more" indicator at the bottom.

### App.svelte Integration

```javascript
let questionQueue = [];

EventsOn('bmad:node:question', (event) => {
  // Replace existing question for same nodeId (hash changed = new question)
  questionQueue = questionQueue.filter(q => q.nodeId !== event.nodeId);
  questionQueue = [...questionQueue, event];
});

EventsOn('bmad:node:question:dismissed', (event) => {
  questionQueue = questionQueue.filter(q => q.nodeId !== event.nodeId);
});
```

On snackbar `navigate` event:
1. Set `builderRepoPath = event.detail.repoPath`
2. Set `currentView = 'workflows'`
3. Set `pendingQuestion = event.detail.question` (new variable for Story 4)

Mount `QuestionSnackbarStack` outside the view conditionals (always visible):
```svelte
<QuestionSnackbarStack
  questions={questionQueue}
  on:navigate={handleQuestionNavigate}
/>
```

### Technical Considerations

- **localStorage access:** `localStorage.getItem('mashed:repoBorderColors')` may return null or malformed JSON. Wrap in try/catch, default to `#1e2530`.
- **Relative time helper:** Simple inline function: `function timeAgo(unixMillis) { ... }` supporting "just now", "Nm ago", "Nh ago". No external library.
- **Reactivity:** `questionQueue` is a Svelte reactive array; reassignment triggers UI updates. Use spread operator for immutable updates.
- **No Wails binding calls in this story:** The snackbar only displays data from events and dispatches navigation intents. The actual response is handled in Story 4.

### Risks & Edge Cases

- Snackbar could overlap with other fixed-position elements. The z-index 300 should be above the canvas toolbar but below modals.
- If many questions arrive simultaneously (e.g., 10+ nodes asking questions), the stack could overflow. The max-5 limit with "+N more" handles this.
- `repoName` extraction: `QuestionEvent.repoName` is set by the backend (last path component of `repoPath`). If empty, fall back to "Unknown Repo".

### Reference Files

- `frontend/src/App.svelte` -- existing `EventsOn` usage pattern (lines 73-129)
- `frontend/src/components/bmad/AgentConfigModal.svelte` -- modal/overlay styling patterns
- `frontend/src/components/bmad/ExecutionBar.svelte` -- status display patterns
- `DESIGN.md` -- design system tokens (`--surface-2`, `--text-primary`, `--text-secondary`, `--accent-green`)

## Acceptance Criteria

AC-1: Snackbar appears when a question event is received
- Given the user is on any view (feed, detail, settings, workflows)
- When a `bmad:node:question` event is received from the backend
- Then a snackbar card appears in the top-right corner showing the repo name and truncated question text

AC-2: Snackbar disappears when a dismissal event is received
- Given a snackbar is visible for node "node-A"
- When a `bmad:node:question:dismissed` event is received with `nodeId: "node-A"`
- Then the snackbar for "node-A" is removed from the stack with a fly-out animation

AC-3: Clicking a snackbar navigates to the workflows view for that repo
- Given a snackbar is visible for a question from repo at `/Users/dev/my-project`
- When the user clicks the snackbar
- Then `currentView` changes to `'workflows'`
- And `builderRepoPath` is set to `/Users/dev/my-project`

AC-4: Snackbar shows correct visual styling
- Given a question event with `repoName: "my-project"`
- When the snackbar renders
- Then the left border color matches the repo's saved border color from localStorage
- And background uses `var(--surface-2)`
- And question text is truncated at 80 characters with ellipsis

AC-5: New question from same node replaces the old snackbar
- Given a snackbar is visible for node "node-A" with question "First question?"
- When a new `bmad:node:question` event arrives for "node-A" with question "Second question?"
- Then only one snackbar for "node-A" is shown
- And it displays "Second question?"

AC-6: Maximum of 5 snackbars are shown with overflow indicator
- Given 7 pending question events
- When the snackbar stack renders
- Then 5 snackbar cards are visible
- And a "+2 more" indicator is shown

## BDD Test Scenarios

### Scenario 1: Event-Driven Snackbar Lifecycle

```gherkin
Feature: Question snackbar stack

  Scenario: Single question event creates snackbar
    Given the questionQueue is empty
    When a bmad:node:question event is received with nodeId "n1" and question "What file?"
    Then questionQueue has 1 entry
    And the snackbar stack renders 1 card with text containing "What file?"

  Scenario: Dismissed event removes snackbar
    Given questionQueue contains an entry for nodeId "n1"
    When a bmad:node:question:dismissed event is received for nodeId "n1"
    Then questionQueue is empty
    And the snackbar stack renders 0 cards

  Scenario: Multiple questions from different nodes
    Given questionQueue contains entries for nodeId "n1" and "n2"
    When the snackbar stack renders
    Then 2 snackbar cards are visible
    And each shows its respective question text
```

### Scenario 2: Navigation on Click

```gherkin
Feature: Snackbar click navigation

  Scenario: Click navigates to workflow view
    Given a snackbar for repoPath "/Users/dev/project-a" is visible
    And currentView is "feed"
    When the user clicks the snackbar
    Then currentView becomes "workflows"
    And builderRepoPath is "/Users/dev/project-a"

  Scenario: Click from settings view also works
    Given a snackbar is visible
    And currentView is "settings"
    When the user clicks the snackbar
    Then currentView becomes "workflows"
```

### Scenario 3: Visual Styling

```gherkin
Feature: Snackbar visual consistency

  Scenario: Repo border color from localStorage
    Given localStorage mashed:repoBorderColors contains {"my-repo": "#ff6b6b"}
    And a question event has repoName "my-repo"
    When the snackbar renders
    Then the left border color is "#ff6b6b"

  Scenario: Fallback color when repo not in localStorage
    Given localStorage mashed:repoBorderColors does not contain "unknown-repo"
    And a question event has repoName "unknown-repo"
    When the snackbar renders
    Then the left border color is "#1e2530"

  Scenario: Long question text is truncated
    Given a question event with question text of 120 characters
    When the snackbar renders
    Then the displayed text is 80 characters followed by an ellipsis
```

### Scenario 4: Overflow Handling

```gherkin
Feature: Snackbar overflow

  Scenario: More than 5 questions shows indicator
    Given 7 question events in questionQueue
    When the snackbar stack renders
    Then 5 snackbar cards are visible
    And a text element shows "+2 more"

  Scenario: Exactly 5 questions shows no indicator
    Given 5 question events in questionQueue
    When the snackbar stack renders
    Then 5 snackbar cards are visible
    And no overflow indicator is shown
```

## Tasks / Subtasks

- [ ] Task 1: Create `QuestionSnackbarStack.svelte` component (AC: AC-1, AC-4, AC-6)
  - [ ] Subtask 1a: Create component with `questions` prop and `navigate`/`dismiss` event dispatchers
  - [ ] Subtask 1b: Implement snackbar card layout with left border color, repo name, truncated question, timestamp
  - [ ] Subtask 1c: Add `fly` transition (from right, duration 300ms)
  - [ ] Subtask 1d: Add max-5 display limit with "+N more" overflow indicator
  - [ ] Subtask 1e: Implement `timeAgo()` helper and `truncate()` helper

- [ ] Task 2: Integrate into `App.svelte` event system (AC: AC-1, AC-2, AC-5)
  - [ ] Subtask 2a: Add `questionQueue` reactive variable
  - [ ] Subtask 2b: Add `EventsOn('bmad:node:question', ...)` handler with node-dedup logic
  - [ ] Subtask 2c: Add `EventsOn('bmad:node:question:dismissed', ...)` handler
  - [ ] Subtask 2d: Mount `QuestionSnackbarStack` outside view conditionals, pass `questionQueue`

- [ ] Task 3: Implement snackbar click navigation (AC: AC-3)
  - [ ] Subtask 3a: Add `handleQuestionNavigate` function that sets `builderRepoPath` and `currentView`
  - [ ] Subtask 3b: Add `pendingQuestion` variable (used by Story 4) and set it on navigate

- [ ] Task 4: Style and design system compliance (AC: AC-4)
  - [ ] Subtask 4a: Use `var(--surface-2)`, `var(--text-primary)`, `var(--text-secondary)` tokens
  - [ ] Subtask 4b: Implement localStorage repo color lookup with fallback
  - [ ] Subtask 4c: Verify z-index 300 does not conflict with existing overlays

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes (no backend changes but verify nothing broke)
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues

## Design Brief

### 1. Layout Composition

The snackbar stack is a **fixed-position overlay** in the top-right corner of the viewport, independent of all view layouts. It sits above normal content but below modals.

```
position: fixed
top: var(--sp-lg)          /* 16px */
right: var(--sp-lg)        /* 16px */
z-index: 200               /* above TitleBar z:200, below modals z:100+overlay */
```

**z-index rationale**: The story specifies z-index 300, but the existing codebase uses z:100 for modals and z:200 for TitleBar. To sit above TitleBar but below modals, use **z-index: 250**. This avoids conflicting with the TitleBar (200) and stays below the modal overlay layer. If modals are raised to 400+ per Story 4, then 300 is also safe -- use **300** per the story spec for forward-compatibility.

Each snackbar card is a horizontal row:

```
[4px accent stripe] [icon 16px] [  repo name / question / time  ] 
                                 ^--- flex-grow text block
```

Stack container: `display: flex; flex-direction: column; gap: var(--sp-sm); width: 320px;`

Individual card internal layout:
- Outer: `display: flex; align-items: stretch;` (stretch so the left border stripe fills full height)
- Left accent stripe: `width: 3px; flex-shrink: 0; border-radius: var(--radius-sm) 0 0 var(--radius-sm);` with dynamic `background-color` from localStorage repo color
- Content area: `display: flex; flex-direction: column; padding: var(--sp-sm) var(--sp-md); gap: var(--sp-2xs); flex: 1; min-width: 0;`
- Top row (repo + time): `display: flex; justify-content: space-between; align-items: baseline;`
- Bottom row: truncated question text

### 2. Typography Plan

| Element | Font | Size | Weight | Color | Extra |
|---------|------|------|--------|-------|-------|
| Repo name | `var(--font-ui)` | `var(--text-body)` (13px) | 600 | `var(--text-primary)` | `letter-spacing: -0.01em` |
| Question text | `var(--font-ui)` | `var(--text-body)` (13px) | 400 | `var(--text-dim)` | Truncate at 80 chars + ellipsis. `white-space: nowrap; overflow: hidden; text-overflow: ellipsis;` |
| Timestamp | `var(--font-mono)` | `var(--text-label)` (11px) | 400 | `var(--text-muted)` | `font-variant-numeric: tabular-nums;` |
| Overflow indicator "+N more" | `var(--font-mono)` | `var(--text-label)` (11px) | 500 | `var(--text-dim)` | Centered, `letter-spacing: 0.02em` |

All text: `line-height: 1.4` (tighter than body default 1.5 for compact notification density).

### 3. Color Strategy

| Surface | Token | Value |
|---------|-------|-------|
| Card background | `var(--bg-elevated)` | `#12151a` |
| Card border | `1px solid var(--border-subtle)` | `#1e2530` |
| Card hover background | `var(--bg-active)` | `#181c23` |
| Left accent stripe | Dynamic from `localStorage('mashed:repoBorderColors')` | per-repo color |
| Left accent fallback | `var(--border-subtle)` | `#1e2530` |
| Icon color | `var(--accent-amber)` | `#f0a500` -- amber signals "needs attention" |
| Overflow indicator bg | `var(--bg-surface)` | `#0d0f12` |

**Note on story's `var(--surface-2)` reference**: This token does NOT exist in `style.css`. The actual equivalent is `var(--bg-elevated)` (`#12151a`). All implementation must use the real token names from the design system.

**No raw hex or rgba()**: Use `color-mix(in srgb, var(--token) N%, transparent)` for any opacity needs, following the StatusBadge pattern and the Do-Not-Repeat rule.

Card box-shadow: `0 2px 8px color-mix(in srgb, var(--bg-deepest) 60%, transparent)` -- subtle depth without a heavy drop-shadow.

### 4. Interaction Model

**Primary action**: Click anywhere on the snackbar card to navigate to the workflows view for that repo.

**Hover behavior**:
- Card: `background` transitions from `var(--bg-elevated)` to `var(--bg-active)` over `var(--duration-short)` with `var(--ease-enter)`.
- Card: `border-color` transitions from `var(--border-subtle)` to `var(--border-emphasis)`.
- Cursor: `pointer`.

**Keyboard navigation**: Not required for V1 (snackbars are transient notifications, not a primary navigation surface). However, each card should have `role="button"` and `tabindex="0"` for accessibility, with `on:keydown` handling Enter to trigger navigation.

**State transitions**:
- **Entry**: New snackbar flies in from the right (`fly={{ x: 300, duration: 300, easing: cubicOut }}`). The story specifies 300ms which is longer than the design system's max 150ms -- use **200ms** as a compromise that feels snappy but visible. Easing: `cubicOut` (aggressive deceleration for "arriving" feel).
- **Exit** (dismissed): `fly={{ x: 300, duration: 200, easing: cubicIn }}` -- exits back to the right.
- **Replace** (same nodeId, new question): The old card exits, new card enters. Svelte's keyed `{#each}` with `fly` transition handles this automatically when keyed by `nodeId`.

**Dismiss button**: Not specified in the story (dismiss only via backend event). No X button on individual snackbars -- they clear when the process no longer needs input.

### 5. Component Specs

```css
/* QuestionSnackbarStack.svelte */
.snackbar-stack {
  position: fixed;
  top: var(--sp-lg);
  right: var(--sp-lg);
  z-index: 300;
  display: flex;
  flex-direction: column;
  gap: var(--sp-sm);
  width: 320px;
  pointer-events: none;        /* allow clicks to pass through gaps */
}

.snackbar-card {
  pointer-events: auto;        /* re-enable clicks on cards */
  display: flex;
  align-items: stretch;
  background: var(--bg-elevated);
  border: 1px solid var(--border-subtle);
  border-radius: var(--radius-md);
  cursor: pointer;
  overflow: hidden;
  transition:
    background var(--duration-short) var(--ease-enter),
    border-color var(--duration-short) var(--ease-enter);
}

.snackbar-card:hover {
  background: var(--bg-active);
  border-color: var(--border-emphasis);
}

.snackbar-card:active {
  transform: scale(0.98);
}

.accent-stripe {
  width: 3px;
  flex-shrink: 0;
  border-radius: var(--radius-sm) 0 0 var(--radius-sm);
  /* background-color set dynamically via style:background-color */
}

.card-content {
  display: flex;
  flex-direction: column;
  padding: var(--sp-sm) var(--sp-md);
  gap: var(--sp-2xs);
  flex: 1;
  min-width: 0;
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--sp-sm);
}

.repo-name {
  font-weight: 600;
  font-size: var(--text-body);
  color: var(--text-primary);
  line-height: 1.4;
  letter-spacing: -0.01em;
}

.question-icon {
  color: var(--accent-amber);
  flex-shrink: 0;
}

.timestamp {
  font-family: var(--font-mono);
  font-size: var(--text-label);
  color: var(--text-muted);
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
  flex-shrink: 0;
}

.question-text {
  font-size: var(--text-body);
  color: var(--text-dim);
  line-height: 1.4;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.overflow-indicator {
  font-family: var(--font-mono);
  font-size: var(--text-label);
  font-weight: 500;
  color: var(--text-dim);
  text-align: center;
  padding: var(--sp-xs) 0;
  letter-spacing: 0.02em;
  pointer-events: auto;
}
```

### 6. Motion Design

| Moment | Transition | Duration | Easing | Direction |
|--------|-----------|----------|--------|-----------|
| Card entry | `fly` | 200ms | `cubicOut` | `x: 300` (from right) |
| Card exit (dismissed) | `fly` | 200ms | `cubicIn` | `x: 300` (to right) |
| Card hover bg | CSS `transition` | `var(--duration-short)` (100ms) | `var(--ease-enter)` | n/a |
| Card press | CSS `transform` | 50ms | `ease-out` | `scale(0.98)` |
| Stack reorder | Svelte `animate:flip` | 200ms | `cubicOut` | positional |

**Stagger**: No stagger on entry. Snackbars arrive one at a time from events, so each enters independently. If multiple arrive within 100ms (burst), they will naturally stagger due to event processing.

**animate:flip**: When a card is dismissed from the middle of the stack, remaining cards should smoothly reflow. Use Svelte's `animate:flip={{ duration: 200 }}` on each `{#each}` item.

### 7. Signature Elements

- **Amber question icon**: `MessageCircleQuestion` in `var(--accent-amber)` immediately signals "human input needed" -- distinct from green (success) or red (error). This is the "attention" color in the Mashed palette.
- **Repo accent stripe**: The 3px left border in each repo's unique color creates instant visual grouping -- a user managing 5 repos can identify which project needs them without reading text.
- **Compact density**: 320px wide, tight `var(--sp-sm)` padding. These are glanceable notifications, not content cards. The information hierarchy is: color stripe (which repo) > amber icon (needs input) > repo name (confirmation) > question text (detail).
- **Fly-from-right**: Notifications enter from the edge of the screen, reinforcing the "incoming message" metaphor. The rightward exit on dismiss completes the gesture -- the notification "leaves" the same way it came.
- **No dismiss button**: Snackbars clear only when the question is answered or the process moves on. This is intentional -- the question genuinely requires attention, and it should not be easy to accidentally dismiss it. The user's action is to click and respond (Story 4), not to swipe away.
