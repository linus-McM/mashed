# Story uiqa-09: Render SparkLine in NotificationFeed

**Status:** ready
**Size:** M
**Priority:** P2
**Domain:** fullstack
**Depends on:** uiqa-01

## Description

Wire the existing `SparkLine` component (imported but never rendered — review NF-08 / FIX-18) into each agent row in `NotificationFeed.svelte` to show token-consumption history over time. This is the "Bloomberg in a dev tool" signature moment the review calls out in Section 6 (Moment 2) and one of the three executive-summary "actions that would most improve the app." The story is fullstack because token history isn't currently persisted per agent — the backend must expose a rolling window of samples (e.g., last 20 token-count snapshots per session) and emit them to the frontend.

## Developer Notes

### Backend work

**File:** `internal/sessions/session.go` (or wherever the session state struct lives — read `.wolf/anatomy.md` or grep for `TokenCount` to find it)

- Add a field:
  ```go
  // TokenSamples is a rolling window of recent token counts for sparkline rendering.
  // Newest sample is at the end. Capped at MaxTokenSamples.
  TokenSamples []int `json:"tokenSamples"`
  ```
- Add a constant:
  ```go
  const MaxTokenSamples = 20
  ```
- When the token count is updated (search for the existing `TokenCount` update site), append the new value to `TokenSamples` and trim to `MaxTokenSamples` from the front:
  ```go
  s.TokenSamples = append(s.TokenSamples, newCount)
  if len(s.TokenSamples) > MaxTokenSamples {
      s.TokenSamples = s.TokenSamples[len(s.TokenSamples)-MaxTokenSamples:]
  }
  ```
- Wails binding: any existing method that returns the session struct (`ListAgents`, `GetSession`, etc.) automatically gets the new field. If a DTO mapping exists, update it.

**Concurrency:** if `TokenCount` is updated under a mutex, `TokenSamples` must be under the same lock. Do not introduce a separate mutex.

### Frontend work

**File:** `frontend/src/views/NotificationFeed.svelte`

- SparkLine is already imported at line 14 (review NF-08). Verify the import path and props API by reading `frontend/src/components/SparkLine.svelte` first.
- Inside the agent row template (find by grepping for `token-count` or `elapsed`), add:
  ```svelte
  {#if agent.tokenSamples?.length > 1}
    <SparkLine data={agent.tokenSamples} width={48} height={14} />
  {/if}
  ```
- Position the sparkline between the sub-count and token-count in the row's horizontal layout. Confirm vertical alignment matches the row's typographic baseline.
- The `{#if > 1}` guard avoids rendering for brand-new sessions with a single data point.

### Alternative fallback (if backend work is blocked)

If engineering decides backend sampling is out of scope, the alternative is **removal** (FIX-18 from the review): delete the unused import. **Do NOT do both.** The preferred outcome is rendering with real data.

### Risks & edge cases

- **Sample granularity**: token counts update frequently. If every update pushes a sample, 20 samples might span <1 minute. Decide on a minimum delta (e.g., only push a sample if the count changed by >= 100 tokens OR 1 second has passed since last sample) — put this decision in the backend update site as a helper `func (s *Session) maybeAppendTokenSample(newCount int)`.
- **Persistence**: samples are in-memory only. On restart, the array is empty and the sparkline hides. Document this.
- **JSON size**: 20 ints per session is negligible (~80 bytes).
- **SparkLine API**: verify it accepts `data: number[]` and `width`/`height` props. If it accepts `values` or something else, adapt.
- **Empty state**: `agent.tokenSamples === undefined` on first load from old state — guard with `?.length` (covered).
- **Reactivity**: Svelte reactivity on array mutation — the backend emits full session objects, not patches, so Svelte's default reactive `=` assignment on the repo/agent list will trigger re-render.

### Out of scope

- Sparkline hover tooltip showing exact values.
- Color-coding the sparkline by status (uiqa-10 may add this as a signature enhancement).
- Aggregate status-bar sparkline (uiqa-10, Signature Moment 3).
- Persisting samples across restarts.

### Reference files

- `frontend/src/components/SparkLine.svelte` — existing component (review Section 2.6 confirms it uses `var(--accent-teal)` and `var(--font-mono)`)
- `frontend/src/views/NotificationFeed.svelte:14` — import site
- `internal/sessions/` — backend session state (confirm exact path)
- `app.go` — Wails bindings (if DTO mapping exists here)

Reference skills: `/wails`, `/golang-testing`, `/golang-error-handling`, `/simplify`, `/playwright-cli`.

## Acceptance Criteria

**AC-1: `TokenSamples` field exists on session state**
- Given the session state struct (e.g., `Session` in `internal/sessions/`)
- When the type is inspected
- Then it has a field `TokenSamples []int` with a json tag `tokenSamples`

**AC-2: Token updates append to samples with cap**
- Given a session with `TokenSamples = []int{100, 200, 300}`
- When the token count is updated 25 more times (each with a distinct new value)
- Then `len(TokenSamples) == MaxTokenSamples` (20)
- And the oldest values have been dropped (no `100` at index 0)
- And the newest values are at the end

**AC-3: Sample appending is thread-safe**
- Given the session mutex protects `TokenCount`
- When the test runs `go test -race` with concurrent token updates
- Then no race is detected

**AC-4: SparkLine renders in agent rows when samples exist**
- Given an agent with `tokenSamples.length >= 2`
- When NotificationFeed renders the agent row
- Then an `<svg>` (or the SparkLine root element) is present inside the row
- And the SparkLine receives the samples array as its data prop

**AC-5: SparkLine hidden for empty or single-sample agents**
- Given an agent with `tokenSamples === undefined` or `tokenSamples.length < 2`
- When NotificationFeed renders
- Then no SparkLine is rendered in that row
- And no console errors occur

**AC-6: Sample emits through Wails binding**
- Given the frontend subscribes to the existing session events / ListAgents binding
- When the backend appends a new sample
- Then the frontend receives an agent DTO whose `tokenSamples` array matches the backend state

**AC-7: No dead import remains**
- Given `NotificationFeed.svelte`
- When parsed
- Then `import SparkLine from ...` is still present
- And the `<SparkLine` tag appears at least once in the template

## BDD Test Scenarios

```gherkin
Feature: uiqa-09 render SparkLine

  Scenario: TokenSamples field exists
    Given the Session struct definition
    When reflected or parsed
    Then it has a field "TokenSamples" of type []int with json tag "tokenSamples"

  Scenario: Capped rolling window
    Given a session with TokenSamples initialized empty
    When maybeAppendTokenSample is called 25 times with increasing values
    Then len(TokenSamples) == 20
    And TokenSamples[19] is the most recent value
    And TokenSamples[0] is the 6th-appended value (first 5 dropped)

  Scenario: Minimum delta throttle
    Given a session whose last sample is 1000
    When maybeAppendTokenSample is called with 1010 immediately
    Then TokenSamples length does not change (delta below threshold)

  Scenario: Race-free concurrent updates
    Given 100 goroutines calling maybeAppendTokenSample in parallel
    When go test -race runs
    Then no race condition is reported

  Scenario: SparkLine renders when samples present
    Given an agent prop with tokenSamples [100, 150, 200, 210, 300]
    When <NotificationFeed agents={[agent]} /> is mounted
    Then a <SparkLine> element is present inside the agent row
    And its data prop equals the tokenSamples array

  Scenario: SparkLine hidden for sparse data
    Given an agent prop with tokenSamples undefined
    When mounted
    Then no SparkLine is rendered
    And no runtime errors occur

  Scenario: Wails binding returns samples
    Given a session updated with 5 token samples
    When ListAgents (or equivalent) is called from the frontend
    Then the returned DTO's tokenSamples field equals the backend slice

  Scenario: Import still used
    Given NotificationFeed.svelte
    When searched for "<SparkLine"
    Then at least one occurrence is found
```

## Tasks / Subtasks

- [ ] Task 1: Add TokenSamples field and constant (AC-1)
  - [ ] Subtask 1a: Locate Session struct via `.wolf/anatomy.md` or grep
  - [ ] Subtask 1b: Add field with json tag
  - [ ] Subtask 1c: Add MaxTokenSamples constant

- [ ] Task 2: Implement maybeAppendTokenSample helper (AC-2, AC-3)
  - [ ] Subtask 2a: Decide minimum delta threshold (suggest: abs delta >= 50 tokens OR >= 1s since last)
  - [ ] Subtask 2b: Append + trim under existing mutex
  - [ ] Subtask 2c: Wire into existing TokenCount update site

- [ ] Task 3: Go tests (AC-2, AC-3)
  - [ ] Subtask 3a: Table-driven cap test
  - [ ] Subtask 3b: Delta throttle test
  - [ ] Subtask 3c: Race test with 100 concurrent writers

- [ ] Task 4: Verify Wails DTO propagation (AC-6)
  - [ ] Subtask 4a: Check ListAgents / related bindings serialize the new field
  - [ ] Subtask 4b: Add a Go test asserting JSON marshaling includes tokenSamples

- [ ] Task 5: Render SparkLine in agent row (AC-4, AC-5, AC-7)
  - [ ] Subtask 5a: Read SparkLine.svelte props
  - [ ] Subtask 5b: Add conditional render in NotificationFeed agent row
  - [ ] Subtask 5c: Visual check in wails dev

- [ ] Task 6: Frontend component test (AC-4, AC-5)
  - [ ] Subtask 6a: Mount NotificationFeed with a fake agent that has samples; assert SparkLine present
  - [ ] Subtask 6b: Mount with undefined samples; assert no SparkLine

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on modified Go files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] Frontend build passes
- [ ] `/simplify` run on all modified code
- [ ] No new `#39ff14` introduced anywhere
- [ ] No orphaned token-scale values introduced in modified files
- [ ] Story status updated to `done`

## Design Brief

### Critical discovery: SparkLine is text, not SVG

Before sizing, note that `frontend/src/components/SparkLine.svelte` is a **unicode block-character text component**, not an SVG renderer. Its full API is:

```js
export let data = [];        // number[]
export let maxVal = 0;       // 0 = auto-scale
```

It renders `<span class="sparkline">▁▂▃▄▅▆▇█...</span>`. It does **not** accept `width`, `height`, `stroke`, or `color` props. The story's original `width={48} height={14}` call signature will silently no-op. This brief corrects the API mismatch.

### Dimensions (text-metric, not pixel)

Because SparkLine is glyph-based, sizing is controlled by `font-size`, `letter-spacing`, and character count:

| Property | Value | Source |
|----------|-------|--------|
| Font | `var(--font-mono)` | SparkLine.svelte line 21 (already set) |
| Font-size | `12px` | SparkLine.svelte line 22 (already set) |
| Letter-spacing | `-0.5px` | SparkLine.svelte line 24 (already set — tightens glyphs so they look like a continuous waveform, not discrete boxes) |
| Sample count (width) | `20 chars max` | `MaxTokenSamples = 20` from backend — each char is ~6.5px wide at 12px mono, so effective width is `~130px max`, `~65px` for 10 samples |
| Row height impact | `14px` | Matches `--text-data` line-height of 12px text — zero row-height disruption |

**Do NOT override `width` or `height` attributes on the SparkLine tag.** If the engineer wants a shorter rendering, clip the data array: `<SparkLine data={agent.tokenSamples.slice(-10)} />`.

### Data source (backend field confirmation)

The story specifies the new `TokenSamples []int` field on `Session` in `internal/sessions/` with `MaxTokenSamples = 20`. Sample appending must go through the `maybeAppendTokenSample` helper that throttles:

- **Minimum delta:** `50` tokens (absolute value of new - last)
- **OR minimum interval:** `1s` since last sample
- **Rationale:** token counts update on every streaming chunk (potentially 10/sec). Without throttling, the 20-sample window collapses to a ~2-second span and the sparkline becomes meaningless. With throttle, 20 samples span 20+ seconds of meaningful history.

**Fallback for insufficient history:** `tokenSamples.length < 2` → **hide** the sparkline entirely (per story AC-5). Do NOT render a flat 2-point line placeholder — empty space reads as "no data yet" faster than a flat line reads as "data is flat."

### Color strategy

SparkLine's current color is hardcoded to `var(--accent-teal)` in its `<style>` block. This brief keeps it that way for **running** agents but adds a status-aware override via the parent:

| Agent status | Color | Mechanism |
|--------------|-------|-----------|
| `running` | `var(--accent-teal)` | Default (no override) |
| `queued` / `paused` / `stopped` | `var(--text-dim)` | Parent wraps in `<span class="sparkline-dim">` |
| Peak spike highlight | `var(--accent-green)` | **Out of scope for uiqa-09** — a future polish story may color the highest-value glyph. Do not implement here. |

```svelte
{#if agent.tokenSamples?.length >= 2}
  <span class="sparkline-wrap" class:dimmed={agent.status !== 'running'}>
    <SparkLine data={agent.tokenSamples} />
  </span>
{/if}

<style>
  .sparkline-wrap.dimmed :global(.sparkline) {
    color: var(--text-dim);
  }
</style>
```

Using `:global()` is the Svelte-idiomatic way to reach into a child component's scoped class.

### Placement in agent row

Per the story, placement must not displace existing elements: **model, status, sub-count, summary, tokens, elapsed, kill button**. The row's horizontal layout (grep `NotificationFeed.svelte` for `.agent-row` flexbox) already has a spot where the sparkline fits naturally:

**Position: immediately before the token count, after the elapsed time.**

```
[status-dot] [model] [sub-count] [summary.................] [SPARKLINE] [tokens] [elapsed] [kill]
                                                              ^--- new
```

- **Spacing:** `margin-right: var(--sp-sm)` (8px gap between sparkline and token count)
- **Spacing before:** `margin-left: var(--sp-sm)` (8px gap from summary)
- **Alignment:** `align-self: center` on the wrapper so the glyph baseline sits on the row's vertical center (text-baseline alignment looks misaligned with the other numeric elements)
- **Flex-shrink:** `flex-shrink: 0` so long summaries don't compress the sparkline into illegibility

Why before tokens and not after? The sparkline is the **history** of the number shown next to it — grouping them visually ("here's the graph, here's the current value") reads as one data cell, not two separate concerns.

### Hover/focus state — intentionally deferred

**uiqa-09 does NOT implement hover tooltips.** The story's "Out of scope" section explicitly defers this. Design rationale:

1. The 20-sample window at 12px is too small for precise point tooltips.
2. Tooltip infrastructure is not yet in the codebase; adding it for one component bloats the story.
3. The current-value token count is already visible adjacent to the sparkline.

**However,** add a basic `title` attribute for screen readers and long-hover discovery:

```svelte
<span
  class="sparkline-wrap"
  title="Token history: {agent.tokenSamples[0]} → {agent.tokenSamples.at(-1)} over last {agent.tokenSamples.length} samples"
>
  <SparkLine data={agent.tokenSamples} />
</span>
```

This is zero-infrastructure and satisfies a11y without introducing a tooltip system.

### Performance budget

- **Target:** `20 agents × 20 samples = 400 glyph computations per render`. Negligible — glyph rendering is a single string-map operation.
- **Hard ceiling:** `50 agents simultaneously running`. At 50×20=1000 glyph computations per render at ~60fps = 60k ops/sec, well under any perf concern.
- **Re-render discipline:** SparkLine's reactive block re-runs the `render(data, maxVal)` function on every `data` prop change. Because backend emits full session objects (not patches), every token update causes a full re-render. This is acceptable at current scale but note: if agent count ever exceeds 100, migrate to a derived store that only re-runs affected agents.
- **No SVG draw calls:** because SparkLine is text, there is zero GPU cost. Safe to scale.

### Out of scope confirmation

- No SVG version of SparkLine (would be a separate component/story).
- No hover tooltip (deferred per story).
- No peak-value color accent (deferred per story).
- No aggregate sparkline in status bar (uiqa-10 Sub-brief B).
