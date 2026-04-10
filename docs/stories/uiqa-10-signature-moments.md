# Story uiqa-10: Signature Moments — Running Pulse + Status Bar Ambient + Terminal Commit Panel

**Status:** ready
**Size:** L
**Priority:** P3
**Domain:** frontend
**Depends on:** uiqa-01, uiqa-02, uiqa-04, uiqa-06, uiqa-09

## Description

Implement three of the five "signature moments" from review Section 6 — the enhancements that inject mashed personality into views that are currently functional but generic. Running agent rows get a subtle left-border pulse (Moment 1), the bottom status bar becomes an "ambient signal" with an aggregate sparkline and pulsing indicator (Moment 3), and the commit streaming panel leans into its terminal aesthetic with a monospace prompt character, active-step accent color, and dimmed completed steps (Moment 5). Moments 2 (agent sparklines) and 4 (modal stagger) are covered by uiqa-09 and uiqa-06 respectively and are not re-implemented here.

**This is an enhancement story, not a bug fix.** Each moment can be split into its own story if the sprint lead prefers finer granularity; the three are grouped because they share the "make mashed feel like mashed" theme and all reference the same tokens.

## Developer Notes

### Moment 1: Running Pulse on Agent Rows (review Section 6, Moment 1)

**File:** `frontend/src/views/NotificationFeed.svelte`

The ProcessNode already has `node-pulse` — extend the same pattern to NotificationFeed's `.agent-row`.

Add to `style.css`:
```css
@keyframes agent-running-pulse {
  0%, 100% {
    box-shadow: inset 3px 0 0 color-mix(in srgb, var(--accent-green) 60%, transparent);
  }
  50% {
    box-shadow: inset 3px 0 0 color-mix(in srgb, var(--accent-green) 100%, transparent);
  }
}
.agent-row.is-running {
  animation: agent-running-pulse 2s ease-in-out infinite;
}
```

In the template, add `class:is-running={agent.status === 'running'}` to the `.agent-row` element.

**Reduced motion**: wrap with `@media (prefers-reduced-motion: no-preference) { ... }`.

### Moment 3: Status Bar as Ambient Signal (review Section 6, Moment 3)

**File:** `frontend/src/views/NotificationFeed.svelte` (the bottom status bar) OR a dedicated `StatusBar.svelte` if it exists — grep for "total agents" or "status-bar" to locate.

- Derive an aggregate token history: sum of `tokenSamples` across all running agents per tick. This requires a small derived store or a Svelte `$:` reactive statement:
  ```svelte
  $: aggregateSamples = computeAggregateSamples(agents); // aligns timestamps, sums
  ```
- Render a small SparkLine in the status bar: `<SparkLine data={aggregateSamples} width={64} height={12} color="var(--accent-teal)" />`.
- Add a pulsing dot indicator when any agent is running:
  ```svelte
  {#if anyRunning}
    <span class="status-pulse" aria-label="agents active"></span>
  {/if}
  ```
  ```css
  .status-pulse {
    display: inline-block;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--accent-green);
    box-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-green) 80%, transparent);
    animation: status-pulse 1.5s ease-in-out infinite;
  }
  @keyframes status-pulse {
    0%, 100% { opacity: 0.6; transform: scale(1); }
    50% { opacity: 1; transform: scale(1.15); }
  }
  ```

### Moment 5: Commit Panel Terminal Aesthetic (review Section 6, Moment 5)

**File:** Find via grep — likely `frontend/src/views/NotificationFeed.svelte` (commit streaming inline) or a dedicated `CommitPanel.svelte`. The review says "commit streaming panel already has a terminal feel with step-by-step output."

Changes:
1. Prepend each step with a monospace prompt character `$ ` (or `❯ ` if Unicode is acceptable — simpler: `$ `).
2. Active step uses `color: var(--accent-green);` and `font-weight: 500`.
3. Completed steps dim to `color: var(--text-dim);`.
4. Use `var(--font-mono)` / `var(--font-terminal)` on the step list container.
5. Add a trailing blinking cursor block on the active step:
   ```css
   .commit-step.is-active::after {
     content: '▋';
     color: var(--accent-green);
     animation: cursor-blink 1s steps(2) infinite;
     margin-left: 4px;
   }
   @keyframes cursor-blink {
     0%, 49% { opacity: 1; }
     50%, 100% { opacity: 0; }
   }
   ```

### Risks & edge cases

- **Performance**: infinite CSS animations on many elements (100+ agents) can cost CPU. The running-pulse is cheap (box-shadow only), but measure frame time on a 50-agent test fixture.
- **Aggregate sparkline alignment**: agent token histories have different timestamps. Simpler: sum the latest N=20 values position-wise, ignoring timing misalignment. Document this simplification.
- **Reduced motion**: all three moments must respect `prefers-reduced-motion`. Use the `@media (prefers-reduced-motion: no-preference)` wrapper so animations simply don't play.
- **Existing commit panel structure**: must read the current template before assuming classes. If steps don't have per-step elements, the entire story shifts to "refactor commit panel to emit per-step elements first."
- **Accent color reliance**: all three moments depend on `var(--accent-green)` — works because uiqa-02 guarantees the token is used correctly.

### Splitting guidance

If this story is sized too large (likely — each moment is ~2 hours of work including tests), split into:
- `uiqa-10a` — Running Pulse (smallest, pure CSS)
- `uiqa-10b` — Status Bar Ambient (needs derived aggregate)
- `uiqa-10c` — Commit Panel Terminal Aesthetic (needs template audit)

Label dependencies accordingly. For this backlog, it is left as a single story with a note that the sprint lead may split at intake.

### Out of scope

- Moment 2 (agent row sparklines) — uiqa-09.
- Moment 4 (modal stagger) — uiqa-06 (modal fades); a future polish story may add per-field stagger inside modals.
- Any new backend events or sampling intervals beyond what uiqa-09 provides.

### Reference files

- `frontend/src/components/bmad/ProcessNode.svelte` — reference for `node-pulse` keyframe pattern
- `frontend/src/components/SparkLine.svelte` — reusable sparkline
- `frontend/src/views/NotificationFeed.svelte` — status bar + commit panel location (verify via read)
- `frontend/src/style.css` — keyframes insertion point

Reference skills: `/simplify`, `/playwright-cli`.

## Acceptance Criteria

**AC-1: `.agent-row.is-running` pulses**
- Given an agent in `running` status
- When NotificationFeed renders
- Then the `.agent-row` has class `is-running`
- And a keyframe animation `agent-running-pulse` is declared in `style.css`
- And `getComputedStyle(row).animationName` equals `agent-running-pulse`

**AC-2: Pulse disabled under reduced motion**
- Given `prefers-reduced-motion: reduce`
- When the running agent row is mounted
- Then `getComputedStyle(row).animationName === 'none'` (or the animation does not play)

**AC-3: Status bar shows aggregate SparkLine when any agent has samples**
- Given at least one agent has `tokenSamples.length >= 2`
- When NotificationFeed renders
- Then the status bar contains a `<SparkLine>` element
- And it receives a non-empty `data` array

**AC-4: Status bar pulse dot appears when any agent is running**
- Given at least one agent in `running` status
- When the status bar renders
- Then a `.status-pulse` element is present
- And it has `aria-label="agents active"` (or equivalent)
- And it is absent when no agents are running

**AC-5: Commit panel uses mono font and prompt character**
- Given the commit panel is visible with at least one step
- When parsed
- Then each `.commit-step` element starts with a visible `$ ` prefix (via `::before` pseudo or template text)
- And the container uses `font-family: var(--font-mono)` or `var(--font-terminal)`

**AC-6: Active commit step is highlighted, completed steps dimmed**
- Given three commit steps: step 1 completed, step 2 active, step 3 pending
- When the panel renders
- Then step 1 has `color: var(--text-dim)` (computed)
- And step 2 has `color: var(--accent-green)` (computed)
- And step 2 has the blinking cursor pseudo-element present

**AC-7: All three animations respect `prefers-reduced-motion`**
- Given reduced motion is active
- When each animated element mounts
- Then the computed `animationName` is `none` OR the animation is contained in a `@media (prefers-reduced-motion: no-preference)` block that does not match

## BDD Test Scenarios

```gherkin
Feature: uiqa-10 signature moments

  Scenario: Running agent row pulses
    Given a test mounting NotificationFeed with one agent in running status
    When the agent row is inspected
    Then it has class "is-running"
    And computed animationName is "agent-running-pulse"

  Scenario: Pulse skipped under reduced motion
    Given prefers-reduced-motion is "reduce"
    When a running agent row mounts
    Then computed animationName is "none"

  Scenario: Status bar aggregate sparkline
    Given 3 agents each with 5 token samples
    When the status bar renders
    Then a SparkLine is present
    And its data array length is >= 1

  Scenario: Status bar pulse dot on running
    Given at least one running agent
    When status bar renders
    Then a .status-pulse element exists
    And it has a non-empty aria-label

  Scenario: Status bar pulse dot absent when idle
    Given all agents are idle
    When status bar renders
    Then no .status-pulse element exists

  Scenario: Commit step mono font and prompt
    Given a commit panel with step "Running tests"
    When rendered
    Then the step shows "$ Running tests"
    And the container font-family is mono

  Scenario: Active step green
    Given commit steps [done, active, pending]
    When rendered
    Then the active step computed color matches var(--accent-green)
    And the done step computed color matches var(--text-dim)

  Scenario: Active step has blinking cursor
    Given the active commit step
    When its ::after pseudo-element is inspected
    Then it contains the cursor block character
    And has a blink animation

  Scenario: All animations wrapped in no-preference media query
    Given style.css
    When parsed
    Then each of agent-running-pulse, status-pulse, cursor-blink keyframe rules
    Or their .class animation declarations
    Are inside a @media (prefers-reduced-motion: no-preference) block
```

## Tasks / Subtasks

- [ ] Task 1: Moment 1 — Running pulse (AC-1, AC-2, AC-7)
  - [ ] Subtask 1a: Add `agent-running-pulse` keyframe to style.css inside a reduced-motion media wrapper
  - [ ] Subtask 1b: Add `class:is-running` binding to `.agent-row` in NotificationFeed
  - [ ] Subtask 1c: Test pulse on + off under reduced motion

- [ ] Task 2: Moment 3 — Status bar ambient (AC-3, AC-4, AC-7)
  - [ ] Subtask 2a: Locate status bar template and read it
  - [ ] Subtask 2b: Add derived `aggregateSamples` reactive computation
  - [ ] Subtask 2c: Render SparkLine with teal accent
  - [ ] Subtask 2d: Add `.status-pulse` element conditional on `anyRunning`
  - [ ] Subtask 2e: Add status-pulse keyframe inside reduced-motion wrapper

- [ ] Task 3: Moment 5 — Commit panel terminal aesthetic (AC-5, AC-6, AC-7)
  - [ ] Subtask 3a: Locate commit panel template and read it
  - [ ] Subtask 3b: Apply mono font to container
  - [ ] Subtask 3c: Add `$ ` prompt prefix via `::before` or template
  - [ ] Subtask 3d: Style active / completed / pending states
  - [ ] Subtask 3e: Add blinking cursor pseudo-element + keyframe

- [ ] Task 4: Frontend component tests (AC-1 through AC-7)
  - [ ] Subtask 4a: Running pulse tests
  - [ ] Subtask 4b: Status bar sparkline + pulse tests
  - [ ] Subtask 4c: Commit panel rendering tests
  - [ ] Subtask 4d: Reduced motion assertions for each

- [ ] Task 5: Perf sanity check
  - [ ] Subtask 5a: Mount 50-agent fixture with all running; verify frame time remains under 16ms on dev hardware

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
- [ ] Reduced motion manually verified
- [ ] Story status updated to `done`

## Design Brief

### Overall philosophy

The three signature moments exist to **make mashed feel inhabited** — the difference between a dashboard that shows data and a command center that feels like it's alive. Every motion here is ambient: the user does not consciously watch any of them. They should feel the room breathe.

Token discipline: all three moments reference only existing tokens in `frontend/src/style.css` plus the uiqa-01 additions (`--accent-cyan`, `--accent-orange`, `--overlay-backdrop`, `--glow-spread`). No new tokens invented.

---

### Sub-brief A: Running pulse on agent rows

#### Which element pulses

**Left border only.** Specifically: `box-shadow: inset 3px 0 0 <color>`. Not full-row box-shadow (perf cost on 20+ rows), not background color (conflicts with hover/selected), not the status dot (too small to read peripherally). The left border is already the row's status accent per DESIGN.md line 57 — pulsing it extends existing language rather than introducing a new idiom.

#### Animation specification

```css
@media (prefers-reduced-motion: no-preference) {
  @keyframes agent-running-pulse {
    0%, 100% {
      box-shadow: inset 3px 0 0 color-mix(in srgb, var(--accent-green) 55%, transparent);
    }
    50% {
      box-shadow: inset 3px 0 0 color-mix(in srgb, var(--accent-green) 100%, transparent);
    }
  }
  .agent-row.is-running {
    animation: agent-running-pulse 2000ms var(--ease-move) infinite;
  }
}

/* Static fallback — always applied, pulse only overlays when motion allowed */
.agent-row.is-running {
  box-shadow: inset 3px 0 0 var(--accent-green);
}
```

| Field | Value | Reason |
|-------|-------|--------|
| Keyframe name | `agent-running-pulse` | Matches story spec |
| Duration | `2000ms` | Cardiac-rest cadence — reads as breathing, not blinking |
| Timing function | `var(--ease-move)` (ease-in-out) | Symmetric wave, not a flash |
| Iteration | `infinite` | |
| Low opacity | `55%` | Just above "is this broken?" threshold |
| High opacity | `100%` | Full accent at peak |

#### Color mechanism

`color-mix(in srgb, var(--accent-green) 55%, transparent)` is the token-compliant way to vary opacity. Do NOT use `rgba(0, 229, 122, 0.55)` — hardcoded, drifts if `--accent-green` changes.

#### Reduced-motion behavior

The `@media` wrapper excludes both the keyframe AND the animation declaration. The static fallback `box-shadow: inset 3px 0 0 var(--accent-green)` outside the media query remains — reduced-motion users see a **solid green left border** indicating "running," just without the pulse. Never remove the status signal, only the motion.

---

### Sub-brief B: Status bar ambient sparkline

#### Which metric drives it

**Aggregate token rate** — position-wise sum of `tokenSamples` arrays across all running agents. This is the "Bloomberg ticker" feel: one line showing cumulative fleet productivity. Rejected alternatives: concurrent agent count (stepped/staircase), CPU (not project-relevant), notifications per minute (sparse/flat).

#### Size budget

| Property | Value |
|----------|-------|
| Height | `12px` (SparkLine text-ascender height at 12px font) |
| Width | `~130px` (20 chars × ~6.5px at 12px mono) |
| Font | inherited `var(--font-mono)` from SparkLine |
| Color | `var(--accent-teal)` (default — teal is ambient-data per DESIGN.md; green is reserved for "alive/running" pulse) |

SparkLine is the existing text-block component (verified in `frontend/src/components/SparkLine.svelte`). It does not accept `width`/`height` props — size is glyph-driven.

#### Update rate

**1 Hz cadence**, driven naturally by the existing `agents` reactive prop. Because uiqa-09's backend throttles token samples (50-token or 1s delta), the sparkline updates at ~1Hz without any additional timer. Do NOT add a `setInterval` — it causes redundant renders on idle.

```svelte
$: aggregateSamples = agents
  .filter(a => a.status === 'running' && a.tokenSamples?.length)
  .reduce((acc, a) => {
    a.tokenSamples.forEach((v, i) => { acc[i] = (acc[i] ?? 0) + v; });
    return acc;
  }, []);
```

#### Visual weight

Must not compete with primary status data (agent count, total tokens, elapsed):

- Opacity `0.7` on wrapper → reads as secondary
- `margin-left: var(--sp-md)` (12px) separation from preceding elements
- No border, no background — floats in status bar surface
- No hover state (ambient = non-interactive)

#### Pulse dot specification

```css
.status-pulse {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent-green);
  box-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-green) 80%, transparent);
  margin-right: var(--sp-sm);
  vertical-align: middle;
}

@media (prefers-reduced-motion: no-preference) {
  @keyframes status-pulse {
    0%, 100% { opacity: 0.6; transform: scale(1); }
    50%      { opacity: 1;   transform: scale(1.15); }
  }
  .status-pulse {
    animation: status-pulse 1500ms var(--ease-move) infinite;
  }
}
```

Duration `1500ms` — faster than the 2000ms agent-row pulse. One dot can afford faster cadence than many rows; it feels "active" without visual noise. Reduced-motion fallback: static green dot with glow, no animation.

---

### Sub-brief C: Terminal-style commit panel

#### Typography

Every element in the commit panel uses `var(--font-mono)`:

```css
.commit-panel,
.commit-panel * {
  font-family: var(--font-mono);
}
```

Font-size `var(--text-body)` (13px). Do NOT use `var(--font-terminal)` — that's JetBrains Mono, reserved per DESIGN.md for actual terminal panes. Commit panel is "terminal aesthetic" (Geist Mono), not "terminal content."

#### Prompt character

**Use `$ `** — dollar sign + single space. Not `❯`, not `›`:

- `$` is the universal shell prompt; instantly readable
- Pure ASCII — zero Unicode rendering concerns
- Two-char width gives a predictable alignment grid

Implemented via `::before` pseudo-element so it's decoration (screen readers skip it):

```css
.commit-step::before {
  content: '$ ';
  color: var(--text-muted);
  font-family: var(--font-mono);
}
.commit-step.is-done::before   { color: var(--text-dim); }
.commit-step.is-active::before { color: var(--accent-green); }
```

Same character across states — only color differs so vertical alignment stays rock-solid.

#### State colors

| State | Color | Weight | Reason |
|-------|-------|--------|--------|
| `.is-done` | `var(--text-dim)` (#4a5a6a) | `400` | Past events recede into history |
| `.is-active` | `var(--accent-green)` (#00e57a) | `500` | Current action gets brand accent — attention lives here |
| `.is-pending` | `var(--text-muted)` (#2e3d4d) | `400` | Below dim — "not yet real" |

These map exactly to style.css lines 17-19. The three-tier hierarchy (muted < dim < primary) already encodes the narrative; the accent hits only the one active line.

#### Line-height and vertical rhythm

```css
.commit-panel {
  background: var(--bg-deepest);   /* #07080a — deeper than surface */
  padding: var(--sp-md);            /* 12px */
  font-family: var(--font-mono);
  font-size: var(--text-body);      /* 13px */
  line-height: 1.5;                 /* 19.5px — standard terminal rhythm */
  border-radius: 0;                 /* DESIGN.md: none for terminal panes */
}

.commit-step {
  display: block;
  padding: 0;                       /* rhythm via line-height, not padding */
  white-space: pre;                 /* preserve mono alignment */
}
```

- Background `var(--bg-deepest)` reinforces the "this is a terminal" metaphor (deeper than surrounding surfaces)
- `border-radius: 0` per DESIGN.md "none for terminal panes"
- `line-height: 1.5` matches Zed/Warp/VT100 rhythm

#### Blinking cursor on active step

```css
.commit-step.is-active::after {
  content: '▋';
  color: var(--accent-green);
  margin-left: var(--sp-xs);
  vertical-align: baseline;
}

@media (prefers-reduced-motion: no-preference) {
  @keyframes cursor-blink {
    0%, 49%   { opacity: 1; }
    50%, 100% { opacity: 0; }
  }
  .commit-step.is-active::after {
    animation: cursor-blink 1000ms steps(2, end) infinite;
  }
}
```

- Character `▋` (U+258B, left five-eighths block) — heavier than `|`, reads as a real cursor
- Duration `1000ms` — industry-standard terminal blink rate (VT100, modern shells)
- Timing `steps(2, end)` — hard on/off; fading cursors look like animation bugs
- Reduced-motion fallback: cursor stays statically visible so users still know which step is active

#### Scanline overlay — DEFERRED

Story mentions "optional subtle scanline overlay." **Do not implement.** Rationale: scanlines need a `repeating-linear-gradient` full-screen overlay (GPU cost per paint), and the monospace + prompt + color hierarchy already delivers the terminal feel. Scanlines are diminishing returns and risk feeling like a costume. If a future polish story adds them, gate behind a user preference toggle.

---

### Cross-moment coordination

The three moments share a rhythm: **row pulse 2000ms, status dot pulse 1500ms, cursor blink 1000ms**. These are deliberately non-harmonic (not 2:1 ratios) so they never visually sync. Syncing would create a "disco" effect — the eye catches the coincidence and the illusion of ambient life collapses. Non-harmonic means the three animations drift relative to each other indefinitely, which is exactly how real systems feel.
