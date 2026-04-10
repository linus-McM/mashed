# Story uiqa-06: Entry Animations — Repo Groups & Modals

**Status:** ready
**Size:** L
**Priority:** P2
**Domain:** frontend
**Depends on:** uiqa-01

## Description

Wire Svelte `transition:` directives into the three highest-impact motion deadzones the review identified: (1) NotificationFeed repo groups enter with staggered `fly` (40ms per group), (2) repo expand/collapse uses `slide`, and (3) modal entry/exit (NewSessionModal, SpawnAgent, BranchModal) uses `fade` (100ms). Every transition uses `var(--duration-medium)` (150ms) or `var(--duration-short)` (100ms) from `style.css` so timings stay consistent and theme-aware. This is the single biggest UX improvement per the review — Motion & Interaction is the weakest dimension (3.4/10) and this story lifts it to the high 5s.

## Developer Notes

### Files and exact changes

#### 1. NotificationFeed repo groups — `fly` with stagger (FIX-05)

**File:** `frontend/src/views/NotificationFeed.svelte`
**Location:** lines 707–1034 (the `{#each}` block rendering repo groups — verify by reading before editing)

- Import at top of `<script>`: `import { fly, slide } from 'svelte/transition';`
- On the repo group element inside the `{#each repos as repo, i}` block, add:
  ```svelte
  <div
    class="repo-group"
    transition:fly={{ y: -8, duration: 150, delay: i * 40 }}
  >
  ```
- `y: -8` is a subtle drop-in. `duration: 150` matches `--duration-medium`. `delay: i * 40` is the 40ms stagger from the review.
- **Do NOT** use `var(--duration-medium)` inside the Svelte prop — Svelte transitions take numeric ms, not CSS var strings. Use the literal `150`. Document this in a comment: `// 150ms matches --duration-medium; keep in sync`.

#### 2. NotificationFeed repo body collapse/expand — `slide` (FIX-06)

**File:** same
**Location:** inside the repo group where `{#if !collapsed}` (or equivalent) wraps the agent list — find this by reading the template around line 900.

- Wrap the agent list inside an `{#if ...}` block that already exists with:
  ```svelte
  {#if !repo.collapsed}
    <div class="agent-list" transition:slide={{ duration: 150 }}>
      ...
    </div>
  {/if}
  ```
- Slide direction is vertical by default (correct for accordion behavior).

#### 3. Modal fade entry/exit — `fade` 100ms (FIX-07, SM-03)

**Files:**
- `frontend/src/views/NewSessionModal.svelte`
- `frontend/src/views/SpawnAgent.svelte`
- `frontend/src/views/BranchModal.svelte` (if present; verify path)

For each:
- Import: `import { fade } from 'svelte/transition';`
- On the backdrop element: `transition:fade={{ duration: 100 }}`
- On the modal card/content element: `transition:fade={{ duration: 100, delay: 20 }}` — backdrop fades first, card follows fractionally behind for polish.

### Accessibility — `prefers-reduced-motion`

Wrap transitions so they respect `prefers-reduced-motion`. Svelte transitions don't auto-respect this, so add a utility:

```js
// at top of <script> in each file using transitions
const reducedMotion = typeof window !== 'undefined' &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches;
const motionDuration = reducedMotion ? 0 : 150;
```

Then use `motionDuration` in the transition props. For the stagger, set `delay: reducedMotion ? 0 : i * 40`.

### Risks & edge cases

- **Svelte `transition:` vs `in:`/`out:`**: `transition:` handles both directions. For a fly-in but no fly-out, use `in:fly` instead. The review asks for both directions for modals (fade in AND out), so `transition:fade` is correct.
- **Reactivity with `{#each}`**: Svelte's keyed each blocks play the transition when an item enters/exits. Confirm the each block uses `(repo.id)` key, not index, so out-of-order reordering doesn't re-animate the whole list.
- **List reordering drag**: NotificationFeed supports drag-to-reorder. If the transition plays on every reorder it looks glitchy. Use the existing flip/move animation (if any) via `animate:flip` and keep `transition:fly` only for genuine adds/removes. Test with reorder.
- **Modal teardown race**: when the modal is destroyed too quickly after open, fade-out can be cancelled. Svelte handles this natively — just don't set duration to 0.
- **Bundle size**: `svelte/transition` adds ~1 KB gzipped. Acceptable.

### Out of scope

- Running-pulse animation on agent rows (uiqa-10, Signature Moment 1).
- Model selector stagger inside NewSessionModal (uiqa-10, Signature Moment 4).
- Animating terminal output, commit panel, or git operations (uiqa-10, Signature Moment 5).
- Entry animations on AgentDetail, Settings, WorkflowBuilder (out of review scope for this story).

### Reference files

- `frontend/src/style.css` line 43–45 — duration tokens (source of truth for durations)
- Svelte docs for `fly`, `slide`, `fade` — `/simplify` and `/playwright-cli` for validation

Reference skills: `/simplify`, `/playwright-cli`.

## Acceptance Criteria

**AC-1: NotificationFeed imports `fly` and `slide` transitions**
- Given `NotificationFeed.svelte`
- When the `<script>` block is parsed
- Then it contains `import { fly, slide } from 'svelte/transition';` (or a combined import)

**AC-2: Repo group has `transition:fly` with staggered delay**
- Given the repo group element inside the `{#each}` block
- When parsed
- Then it has a `transition:fly={...}` directive
- And the delay expression references the loop index multiplied by 40
- And the duration is 150 (or derived from `motionDuration`)

**AC-3: Repo body uses `transition:slide` on expand/collapse**
- Given the collapsible agent list inside each repo group
- When parsed
- Then it has a `transition:slide={{ duration: ... }}` directive

**AC-4: Three modals import and apply `fade`**
- Given NewSessionModal, SpawnAgent, BranchModal (or the subset that exists)
- When parsed
- Then each imports `fade` from `svelte/transition`
- And the backdrop element has `transition:fade={{ duration: 100 }}`
- And the modal card element has `transition:fade={{ duration: 100, delay: 20 }}` (or `delay: 0` for reduced motion)

**AC-5: Reduced-motion honored**
- Given a test harness that forces `matchMedia('(prefers-reduced-motion: reduce)').matches === true`
- When the components mount
- Then transition duration is 0 (animations are effectively instant)

**AC-6: Repo groups enter with observable stagger**
- Given a Playwright test loading NotificationFeed with 3 repo groups
- When the first paint completes
- Then the second repo group's opacity reaches 1 at least 30ms after the first
- And the third reaches 1 at least 30ms after the second

**AC-7: Modal fade visible**
- Given a Playwright test opening NewSessionModal
- When the modal mounts
- Then there is at least one animation frame where the modal opacity is between 0 and 1 (non-instant appearance)

**AC-8: Drag reorder does not trigger `fly` on every item**
- Given a user drags a repo to reorder
- When the reorder completes
- Then no more than one repo group plays the `fly` transition (and only if genuinely added/removed)

## BDD Test Scenarios

```gherkin
Feature: uiqa-06 entry animations

  Scenario: NotificationFeed imports transition helpers
    Given NotificationFeed.svelte
    When the script block is parsed
    Then it imports fly and slide from svelte/transition

  Scenario: Repo group stagger directive present
    Given NotificationFeed.svelte
    When the repo each block is parsed
    Then the repo-group element has transition:fly
    And the delay expression multiplies the loop index by 40

  Scenario: Repo expand/collapse uses slide
    Given NotificationFeed.svelte
    When the collapsible agent list element is parsed
    Then it has transition:slide

  Scenario: NewSessionModal backdrop fades in
    Given a Playwright test
    When NewSessionModal is opened
    Then the backdrop opacity starts below 0.9 on first frame and reaches 1 within 120ms

  Scenario: Modal card fades slightly behind backdrop
    Given NewSessionModal is opened
    When first two animation frames are inspected
    Then the backdrop opacity is ahead of the card opacity by a measurable amount

  Scenario: Reduced motion disables duration
    Given prefers-reduced-motion is reduce
    When NotificationFeed mounts
    Then getComputedStyle on a newly-added repo group shows opacity 1 within 1 frame

  Scenario: Reorder does not re-fly existing groups
    Given 5 repo groups rendered
    When the user drags group 3 above group 1
    Then no repo group opacity drops below 1 during the reorder
```

## Tasks / Subtasks

- [ ] Task 1: Add fly + slide to NotificationFeed (AC-1, AC-2, AC-3, AC-6, AC-8)
  - [ ] Subtask 1a: Read lines 707–1034 to locate the repo each block and collapse/expand control
  - [ ] Subtask 1b: Import fly, slide from svelte/transition
  - [ ] Subtask 1c: Add reducedMotion helper
  - [ ] Subtask 1d: Add `transition:fly` with stagger on repo group
  - [ ] Subtask 1e: Add `transition:slide` on agent list collapse
  - [ ] Subtask 1f: Verify keyed each block uses stable id (not index)

- [ ] Task 2: Add fade to NewSessionModal (AC-4, AC-7)
  - [ ] Subtask 2a: Import fade
  - [ ] Subtask 2b: Backdrop `transition:fade`
  - [ ] Subtask 2c: Card `transition:fade` with 20ms delay

- [ ] Task 3: Add fade to SpawnAgent (AC-4)
  - [ ] Subtask 3a: Same pattern as Task 2

- [ ] Task 4: Add fade to BranchModal if file exists (AC-4)
  - [ ] Subtask 4a: Verify file exists; if not, log as "not applicable"
  - [ ] Subtask 4b: Apply pattern

- [ ] Task 5: Reduced motion respect (AC-5)
  - [ ] Subtask 5a: Add matchMedia check utility
  - [ ] Subtask 5b: Gate durations behind the reducedMotion flag

- [ ] Task 6: Playwright tests (AC-6, AC-7, AC-8)
  - [ ] Subtask 6a: Test asserting staggered opacity rise
  - [ ] Subtask 6b: Test asserting modal backdrop fade
  - [ ] Subtask 6c: Test asserting reorder does not re-fly

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
- [ ] `prefers-reduced-motion` manually verified
- [ ] Story status updated to `done`

## Design Brief

### Motion intent

Motion should feel like a **command center coming online**, not a consumer app. The repo groups don't "bounce" — they settle. Modals don't "zoom" — they resolve. Every curve favors deceleration on entry (content arrives and rests) and acceleration on exit (content gets out of the way). Amplitude is deliberately small: 8px of travel is enough to register motion peripherally without forcing the eye to track it. The user should barely notice the animations individually — they should only notice that the app feels alive.

### Duration tokens (literal ms values; Svelte props require numbers)

All three tiers map to tokens in `frontend/src/style.css` lines 43-45:

| Tier | Token | Literal | Use |
|------|-------|---------|-----|
| Instant feedback | `--duration-micro` | `50` | Modal card delay offset behind backdrop (`delay: 50` overrides the current `delay: 20` — 20ms is sub-frame and invisible) |
| UI acknowledgement | `--duration-short` | `100` | Modal backdrop fade in/out, modal card fade in/out |
| Entry choreography | `--duration-medium` | `150` | Repo group fly-in, repo body slide collapse/expand |

**Do not invent a longer token.** The design system (`DESIGN.md` line 54) caps motion at 150ms. If a transition feels too fast, the fix is tighter easing, not longer duration.

### Easing curves (CSS custom property mapping)

Svelte transitions accept an `easing` function from `svelte/easing`. Map the CSS tokens to Svelte imports:

| CSS token | Svelte import | Cubic-bezier | Direction |
|-----------|---------------|--------------|-----------|
| `--ease-enter` (ease-out) | `cubicOut` | `cubic-bezier(0.33, 1, 0.68, 1)` | Entry: fly-in, slide-open, fade-in |
| `--ease-exit` (ease-in) | `cubicIn` | `cubic-bezier(0.32, 0, 0.67, 0)` | Exit: slide-close, fade-out |
| `--ease-move` (ease-in-out) | `cubicInOut` | `cubic-bezier(0.65, 0, 0.35, 1)` | Flip/reorder animations only |

```js
import { fly, slide, fade } from 'svelte/transition';
import { cubicOut, cubicIn } from 'svelte/easing';
```

For `transition:` (bidirectional) on elements that need asymmetric in/out, split into `in:` and `out:`:

```svelte
<div in:fly={{ y: -8, duration: 150, easing: cubicOut, delay: i * 40 }}
     out:fade={{ duration: 100, easing: cubicIn }}>
```

### Stagger timing — repo group fly-in

- **Per-sibling delay:** `50ms` (midpoint of the 40-60ms band specified in DESIGN.md motion philosophy). Story text currently says 40ms — bump to 50ms so the cadence is perceptible at 60fps (each frame is ~16.6ms, so 50ms is exactly 3 frames between siblings).
- **Max stagger cap:** `400ms`. With 8 repos at 50ms each = 400ms total; beyond 8 repos, clamp so the last item never waits longer than 400ms.
- **Formula:** `delay: Math.min(i * 50, 400)`
- **Rationale:** users with 20+ repo groups (power users) should never wait 1+ seconds for the list to finish appearing. 400ms is the point where a user would start perceiving the list as "loading" rather than "assembling."

### Distance & direction

| Transition | Param | Value | Reason |
|------------|-------|-------|--------|
| Repo group fly-in | `y` | `-8` | Subtle drop from above — reads as "landing into place," not "rising from below" |
| Repo group fly-in | `opacity` | implicit `0` to `1` | Default Svelte behavior |
| Repo body slide | `axis` | `y` (default) | Vertical accordion |
| Repo body slide | `duration` | `150` | Matches fly-in so collapse/expand feels like the same material |
| Modal backdrop fade | no distance | opacity only | Backdrops never travel |
| Modal card fade | no distance | opacity only | Motion would feel like a notification, not a modal |

**Do NOT add `x` translation, scale, or rotation to any of these.** The design system is minimal-functional (DESIGN.md line 49); transforms beyond `y` translation introduce visual noise.

### Reduced-motion behavior

Under `prefers-reduced-motion: reduce`, all three transitions become effectively instant but **content still renders** — never hide elements from reduced-motion users.

```js
const reducedMotion = typeof window !== 'undefined' &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches;

const flyProps = reducedMotion
  ? { y: 0, duration: 0, delay: 0 }
  : { y: -8, duration: 150, delay: i * 50, easing: cubicOut };

const slideProps = reducedMotion
  ? { duration: 0 }
  : { duration: 150, easing: cubicOut };

const fadeProps = reducedMotion
  ? { duration: 0 }
  : { duration: 100, easing: cubicOut };
```

**Critical:** do NOT guard the `transition:` directive itself with an `{#if !reducedMotion}` wrapper — that would cause Svelte to tear down the DOM subtree on preference change. Always keep the directive; only zero out its parameters.

### Code snippet — ready to paste

```svelte
<script>
  import { fly, slide, fade } from 'svelte/transition';
  import { cubicOut, cubicIn } from 'svelte/easing';

  const reducedMotion = typeof window !== 'undefined' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  // Literal ms values mirror style.css --duration-* tokens. Keep in sync.
  const DUR_SHORT = reducedMotion ? 0 : 100;   // --duration-short
  const DUR_MEDIUM = reducedMotion ? 0 : 150;  // --duration-medium
  const STAGGER_MS = reducedMotion ? 0 : 50;
  const STAGGER_CAP = 400;
</script>

{#each repos as repo, i (repo.id)}
  <div
    class="repo-group"
    in:fly={{
      y: -8,
      duration: DUR_MEDIUM,
      delay: Math.min(i * STAGGER_MS, STAGGER_CAP),
      easing: cubicOut
    }}
    out:fade={{ duration: DUR_SHORT, easing: cubicIn }}
  >
    <!-- repo header -->

    {#if !repo.collapsed}
      <div
        class="agent-list"
        transition:slide={{ duration: DUR_MEDIUM, easing: cubicOut }}
      >
        <!-- agent rows -->
      </div>
    {/if}
  </div>
{/each}

<!-- Modal pattern -->
{#if modalOpen}
  <div
    class="modal-backdrop"
    transition:fade={{ duration: DUR_SHORT, easing: cubicOut }}
  >
    <div
      class="modal-card"
      transition:fade={{ duration: DUR_SHORT, delay: reducedMotion ? 0 : 50, easing: cubicOut }}
    >
      <!-- modal content -->
    </div>
  </div>
{/if}
```

**Note on backdrop-to-card delay:** the story originally specifies `delay: 20`, which is ~1.2 frames and imperceptible. Updated to `50` (exactly 3 frames), which creates a legible "backdrop arrives, then card resolves" choreography without feeling sluggish. This change is an intentional brief refinement; update AC-4 accordingly.
