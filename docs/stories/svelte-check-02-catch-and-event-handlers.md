# Story svelte-check-02: Catch + event-handler hygiene (Phase 2)

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** svelte-check-01
**Status:** done
**Landed:** 2026-04-22

## Description

Introduce a single `errorMessage(e: unknown): string` helper and replace every `catch (e) { ... e.message }` across the Svelte views with it, while typing DOM event handlers (`(e: MouseEvent)`, `(e: KeyboardEvent)`, `(e: CustomEvent<T>)`) and adding typed `createEventDispatcher` generics. Eliminates 110 direct errors (P1 + P4 patterns) plus ~50 downstream, taking the count from ~1320 to ~1160.

## Developer Notes

### Architecture
- New helper: `frontend/src/lib/errorMessage.ts`:
  ```ts
  export function errorMessage(e: unknown): string {
    if (e instanceof Error) return e.message;
    if (typeof e === 'string') return e;
    try { return JSON.stringify(e); } catch { return String(e); }
  }
  ```
- Callers: every `.svelte` file with `catch (e) { … e.message }` — top hits in the plan are `Setup.svelte`, `BranchModal.svelte`, `SwitchBranchModal.svelte`, `MergeModal.svelte`, `ForcePushModal.svelte`, but `just sveltecheck-by-file` will show the full list live.
- DOM handler typing: annotate every inline arrow handler parameter. Typical signatures:
  - `on:click={(e: MouseEvent) => ...}`
  - `on:keydown={(e: KeyboardEvent) => ...}`
  - `on:submit|preventDefault={(e: SubmitEvent) => ...}`
- `createEventDispatcher` generics: declare the event map up front —
  ```ts
  const dispatch = createEventDispatcher<{ submit: { value: string }; cancel: void }>();
  ```
- CustomEvent consumers: `on:submit={(e: CustomEvent<{ value: string }>) => e.detail.value}` — the detail shape must match the dispatcher's generic.

### Technical Considerations
- **Single helper, not per-file `try/catch` boilerplate.** The helper is the single write-once utility; its behaviour is tested once with unit tests.
- **`unknown` > `any`.** The catch variable in TS 4.4+ is `unknown` by default. Never widen to `any`.
- **Event type precision.** `MouseEvent` is too broad for handlers that need `.currentTarget`. When needed: `(e: MouseEvent & { currentTarget: HTMLButtonElement })` — or destructure at call site.
- **Template event handlers** that forward (`on:click={handleClick}`) don't need inline types IF `handleClick` is typed in the `<script>` block. Prefer named functions over inline arrows for non-trivial handlers.
- **Preventing regressions in SparkLine-style components:** per cerebrum, SparkLine accepts only `data: number[]` and `maxVal: number`. Don't accidentally add new props while retyping.

### Risks & Edge Cases
- **R3 — cascade.** Fixing catch-variable typing unmasks previously-hidden errors on `.message` accesses where the error flows into a function parameter. Run `just sveltecheck-count` after each file for an early signal.
- **Event type mismatch.** A handler typed as `MouseEvent` but bound to `on:keydown` will compile but misfire. Use the exact handler / event type pair — consult MDN if unsure.
- **PointerEvent vs MouseEvent.** Modern browsers fire pointer events; keep `MouseEvent` unless the component specifically uses pointer features (e.g. `pointerType`).
- **CustomEvent detail = undefined.** When dispatching with no detail, the generic should be `void`, not `{}`. This is why `cancel: void` is shown above.
- **R4 — PR size.** 100+ files get touched in this phase. Split into at most 3 PRs of ≤ 400 lines each: (a) helper + unit tests, (b) views/*.svelte catch migrations, (c) components/**/*.svelte handler typing.

### Reference Files
- Any existing `.svelte` component with a typed `<script lang="ts">` block — e.g. `frontend/src/components/bmad/DiagnosticsChip.svelte` after U8 — as a reference for event dispatcher generic placement.
- `frontend/src/components/bmad/InputResponseModal.svelte` — example of mature typed handlers and dispatcher.
- MDN `GlobalEventHandlersEventMap` — canonical DOM handler→event mapping.

### Skills to invoke
- `/simplify` — mandatory before commit.
- `/golang-error-handling` — reference for `unknown` + narrow pattern (analogous to Go `errors.As`).
- `/playwright-cli` — light smoke of Setup.svelte, BranchModal submit flow after changes.

## Acceptance Criteria

AC-1: Helper created and unit-tested
- Given a new `frontend/src/lib/errorMessage.ts`
- When called with `new Error('boom')`
- Then it returns `"boom"`
- And when called with the string `"plain"` it returns `"plain"`
- And when called with `null` it returns a non-empty string (never throws)
- And a companion `errorMessage.test.ts` covers all three branches

AC-2: All `.message` accesses routed through helper
- Given svelte-check baseline has ~110 P1/P4 errors
- When this story lands
- Then `rg "catch \(e\).*e\.message" frontend/src` returns zero matches
- And every catch block in a `.svelte` file calls `errorMessage(e)` (or equivalent helper)

AC-3: DOM handlers typed
- Given views previously had `(e) => ...` inline handlers
- When this story lands
- Then every inline arrow in a template has an annotated event parameter (e.g. `(e: MouseEvent)`)
- And named handler functions in `<script lang="ts">` blocks have typed parameters

AC-4: Dispatcher generics applied
- Given Svelte files with `createEventDispatcher()`
- When this story lands
- Then every `createEventDispatcher` call provides an explicit event-map generic
- And `CustomEvent` consumers type `e.detail` to match the dispatcher's map

AC-5: Error count drops
- Given baseline ~1320 entering this story
- When it lands
- Then `just sveltecheck-count` reports <= 1210 (target 1160)
- And `just sveltecheck-ratchet` passes

## BDD Test Scenarios

### Scenario 1: errorMessage helper
```gherkin
Feature: errorMessage unknown narrowing

  Scenario: Error instance returns message
    When errorMessage(new Error("fail"))
    Then the result is "fail"

  Scenario: String input returned as-is
    When errorMessage("already-a-string")
    Then the result is "already-a-string"

  Scenario: Null handled without throwing
    When errorMessage(null)
    Then the result is a string
    And no exception is thrown

  Scenario: Object input serialised
    When errorMessage({ code: "E42", detail: "x" })
    Then the result contains "E42"
    And no exception is thrown

  Scenario: Circular reference falls back to String()
    Given a circular object o where o.self = o
    When errorMessage(o) is called
    Then no exception is thrown
    And the result is a non-empty string
```

### Scenario 2: Handler retyping
```gherkin
Feature: DOM event handler typing

  Scenario: MouseEvent handler preserves existing behaviour
    Given BranchModal.svelte has a Submit button with on:click={save}
    When svelte-check runs after retyping
    Then no errors report on that template line
    And the runtime behaviour is unchanged (clicking Submit still submits)

  Scenario: KeyboardEvent handler preserves behaviour
    Given SwitchBranchModal has on:keydown={handleKeydown}
    When handleKeydown is typed "(e: KeyboardEvent) => void"
    Then svelte-check passes
    And Enter key still submits, Escape still closes
```

### Scenario 3: Dispatcher generics
```gherkin
Feature: Typed createEventDispatcher

  Scenario: Dispatcher generic enforces payload shape
    Given createEventDispatcher<{ submit: { value: string } }>()
    When a caller invokes dispatch('submit', { wrong: 1 })
    Then svelte-check reports a type error at the dispatch call
    And the correct call dispatch('submit', { value: 'ok' }) passes

  Scenario: CustomEvent consumer detail is typed
    Given a parent listens on:submit={(e: CustomEvent<{ value: string }>) => ...}
    When e.detail.value is read
    Then the type is string (no implicit any)
    And e.detail.wrong is a compile error
```

### Scenario 4: Error delta
```gherkin
Feature: Phase 2 reduction

  Scenario: Count drops by at least 110
    Given Phase 1 leaves count at ~1320
    When Phase 2 lands
    Then "just sveltecheck-count" is at most 1210
    And the ratchet passes
```

## Tasks / Subtasks

- [x] Task 1: Author errorMessage helper (AC: 1)
  - [x] Create `frontend/src/lib/errorMessage.ts` with Error / string / JSON fallback
  - [x] Create `frontend/src/lib/errorMessage.test.ts` covering Error, string, null, object, circular

- [x] Task 2: Migrate catch blocks (AC: 2)
  - [x] List files via `rg "catch \(e\).*\.message" frontend/src --files-with-matches`
  - [x] In each, import `errorMessage` and replace `e.message` with `errorMessage(e)`
  - [x] Record per-file sveltecheck delta as files are migrated

- [x] Task 3: Type DOM handlers (AC: 3)
  - [x] For every inline arrow in a Svelte template, annotate parameter (`(e: MouseEvent)` / `(e: KeyboardEvent)` etc.)
  - [x] Lift non-trivial handlers out to typed named functions in `<script lang="ts">`
  - [x] Prefer `SubmitEvent` for form submit, `FocusEvent` for focus/blur, `InputEvent` for input

- [x] Task 4: Apply dispatcher generics (AC: 4)
  - [x] For each `createEventDispatcher()` call, define an event-map generic above the call
  - [x] Update all consuming parents to type `CustomEvent` detail to match
  - [x] Prefer `void` over `{}` for detail-less events

- [x] Task 5: Verify reduction (AC: 5)
  - [x] Run `just sveltecheck-count` — confirm <= 1210
  - [x] Run `vitest run` — 696 tests pass
  - [x] Run `vite build` — clean
  - [x] Record commit body delta: `sveltecheck-count: <prev> → <cur> (Δ -<n>)`

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ coverage on `errorMessage.ts` (all four branches — Error, string, object, null)
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `just sveltecheck-count` <= 1210
- [x] `just sveltecheck-ratchet` passes
- [x] `vitest run` — 696 tests pass
- [x] `vite build` clean
- [x] `rg "catch \(e\).*e\.message" frontend/src` → zero matches
- [x] `/simplify` run on every modified file
- [x] Code review: no `any`, no `@ts-ignore`, no `@ts-nocheck`
- [x] PR size <= 400 lines (split into up to 3 PRs — R4)
- [x] Commit body records: `sveltecheck-count: <prev> → <cur> (Δ -<n>)`
