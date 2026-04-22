# Story svelte-check-04c: View-level state — Terminal + AgentDetail + Settings + SummarisationModal (Phase 4c)

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** svelte-check-03
**Status:** ready

## Description

Retype the remaining Phase 4 views: `AgentDetail.svelte` (81), `Terminal.svelte` (50), `Settings.svelte` (33), `SummarisationModal.svelte` (18) — 182 direct errors. Terminal holds an xterm `Terminal` ref plus Wails PTY events; AgentDetail is the core agent dashboard; Settings persists user preferences. Expected reduction: ~182 direct + ~40 downstream = ~220. Count drops from ~640 to ~420. Can run in parallel with 04a and 04b.

## Developer Notes

### Architecture
- Files:
  - `frontend/src/views/AgentDetail.svelte` (81)
  - `frontend/src/components/Terminal.svelte` (50)
  - `frontend/src/views/Settings.svelte` (33)
  - `frontend/src/views/SummarisationModal.svelte` (18)
- xterm refs:
  - `import type { Terminal as Xterm } from '@xterm/xterm'`
  - `import type { FitAddon } from '@xterm/addon-fit'`
  - `let term: Xterm | null = null; let fitAddon: FitAddon | null = null;`
- AgentDetail core types:
  - `Agent` (from `$lib/types/wails`)
  - `Notification[]` (from `$lib/types/wails`)
  - `TokenSamples: number[]` per cerebrum Notification entry (already a number array on the Go side; verify `main.Agent` has this field)
- Settings core types:
  - `EditorSettings`, `AdviceMode[]`, `ModelInfo[]` — all Go-sourced, re-export via `$lib/types/wails`
  - Per cerebrum ModelInfo section: `{ ID, Alias, DisplayName, ContextWindow, Tier, IsDefault }` is authoritative
- SummarisationModal: Wails streaming event payloads (`review:summary:progress`, `review:summary:done`) — define event payload interfaces.

### Technical Considerations
- **xterm PTY events.** Per cerebrum 2026-04-09, PTY architecture uses signed helper + no Setpgid; runtime event names follow `pty:<sessionId>:data|exit|resize`. Type the event payloads explicitly: `{ sessionId: string; chunk: string }` etc.
- **SparkLine in AgentDetail.** Per cerebrum entry on SparkLine, it takes ONLY `data: number[]` and `maxVal: number`. Do not pass width/height or add new props during retyping.
- **Color tokens.** Per cerebrum 2026-04-10, do NOT introduce `#39ff14` — use `var(--accent-green)`. The ban applies to AgentDetail, Settings, any neighbour.
- **Review streaming.** SummarisationModal likely subscribes to `review:summary:progress` (per cerebrum review backend). Event payload includes delta strings; type as `{ delta: string }` → accumulator is `string`.
- **Model registry.** Settings' model dropdown lists `ListModels()` output. Use `ModelInfo[]` from `$lib/types/wails`; don't hardcode model IDs (cerebrum explicit note — single source of truth).
- **R3.** Touching AgentDetail's per-agent notification map may reveal Record typing gaps skipped in Phase 3 — fix in place.

### Risks & Edge Cases
- **xterm dispose cycle.** `onDestroy(() => { term?.dispose(); })` — null-guard preserved after retyping.
- **Wails event unsubscribe.** `EventsOn` returns an unsubscribe function; store it as `let unsubscribe: (() => void) | null = null` and call in `onDestroy`.
- **Settings persistence shape.** The persisted JSON shape must match the `EditorSettings` interface; if the user has stale settings, narrow in the load path and fall back to defaults — never crash on malformed input.
- **SummarisationModal cancelation.** If the user closes the modal mid-stream, `EventsOff` must fire; preserve this during retyping.
- **R4 — PR sizing.** 4 files, likely 2 PRs:
  - PR-A: Terminal + xterm typing
  - PR-B: AgentDetail + Settings + SummarisationModal

### Reference Files
- `frontend/src/views/AgentDetail.svelte` (target)
- `frontend/src/components/Terminal.svelte` (target)
- `frontend/src/views/Settings.svelte` (target)
- `frontend/src/views/SummarisationModal.svelte` (target)
- `frontend/src/components/SparkLine.svelte` — props shape reminder (cerebrum)
- `internal/domain/models.go` — ModelInfo source of truth
- `internal/terminal/` or `internal/ptyhelper/` (per cerebrum PTY sprints) — event name conventions
- `app_review.go` — review streaming event names (per cerebrum)

### Skills to invoke
- `/simplify` — mandatory before commit.
- `/wails` — event subscription typing patterns.
- `/playwright-cli` — smoke Terminal input, AgentDetail detail view, Settings save, SummarisationModal stream.

## Acceptance Criteria

AC-1: xterm refs typed
- Given Terminal.svelte
- When this story lands
- Then `term`, `fitAddon`, `container` have explicit library-sourced types
- And PTY event payloads (`pty:<id>:data`, `:exit`, `:resize`) are typed with dedicated interfaces

AC-2: AgentDetail props + state typed
- Given AgentDetail.svelte
- When this story lands
- Then `Agent` props are typed; SparkLine usage still passes ONLY `data` and `maxVal`
- And no hex colour literal is introduced (cerebrum 2026-04-10)

AC-3: Settings uses ModelInfo registry
- Given Settings.svelte renders a model dropdown
- When this story lands
- Then the dropdown's options are typed `ModelInfo[]` sourced from `ListModels()`
- And no model ID / alias string is hardcoded

AC-4: SummarisationModal event payloads typed
- Given `review:summary:progress` and `review:summary:done`
- When this story lands
- Then each handler parameter is typed with a dedicated interface
- And the modal cleanly unsubscribes on close

AC-5: Error count drops
- Given entering at ~640 (or equivalent if parallel)
- When this story lands
- Then each target file reports ≤ a handful of errors
- And `just sveltecheck-count` drops by ≥ 180
- And `just sveltecheck-ratchet` passes

## BDD Test Scenarios

### Scenario 1: Terminal typing
```gherkin
Feature: Terminal xterm ref typing

  Scenario: Term ref typed
    Given let term: Xterm | null = null
    When term.write(chunk) is called
    Then svelte-check flags the null case and passes when null-checked
    And term.dispose() compiles

  Scenario: PTY data event payload typed
    Given EventsOn("pty:" + sid + ":data", (evt: PtyDataEvent) => term?.write(evt.chunk))
    When svelte-check runs
    Then evt.chunk is typed string
    And evt.wrong is a compile error
```

### Scenario 2: AgentDetail
```gherkin
Feature: AgentDetail agent typing

  Scenario: Agent prop is Agent type
    Given "export let agent: Agent"
    When parent passes agent={42}
    Then svelte-check reports a type error

  Scenario: SparkLine called with ONLY data and maxVal
    Given SparkLine usage in AgentDetail
    When retyping lands
    Then the only props passed are data (number[]) and maxVal (number)
    And no width/height/color prop is introduced

  Scenario: No hex colour literals introduced
    Given the cerebrum ban on "#39ff14"
    When Phase 4c lands
    Then "rg #[0-9a-fA-F]{6} frontend/src/views/AgentDetail.svelte" finds no new matches
```

### Scenario 3: Settings model dropdown
```gherkin
Feature: Settings model dropdown uses registry

  Scenario: Dropdown options typed ModelInfo
    Given "ListModels()" returns Promise<ModelInfo[]>
    When Settings renders the dropdown
    Then the options are ModelInfo instances with ID, Alias, DisplayName
    And no hardcoded model string appears in the template

  Scenario: Default model selected via IsDefault
    Given at least one ModelInfo has IsDefault=true
    When Settings opens
    Then that entry is the initial selection
```

### Scenario 4: Summarisation streaming
```gherkin
Feature: SummarisationModal streams typed progress

  Scenario: Progress accumulator typed
    Given progress handler "(evt: { delta: string }) => void"
    When 5 progress events fire with delta "abc", "def"...
    Then the accumulator is string-concatenated without any type errors

  Scenario: Unsubscribe on close
    Given EventsOn returned unsubscribe function u
    When the modal closes
    Then u() is invoked
    And no subsequent progress event updates state
```

## Tasks / Subtasks

- [ ] Task 1: Terminal.svelte retype (AC: 1)
  - [ ] Import type `Terminal as Xterm` from `@xterm/xterm`, `FitAddon` from `@xterm/addon-fit`
  - [ ] Type `term`, `fitAddon`, `container`
  - [ ] Define PTY event payload interfaces in `frontend/src/types/pty.ts`
  - [ ] Type all `EventsOn` handlers against those interfaces

- [ ] Task 2: AgentDetail.svelte retype (AC: 2)
  - [ ] Type the `agent: Agent` prop and any `notifications: Notification[]` state
  - [ ] Verify SparkLine usage: only `data` and `maxVal` passed
  - [ ] Ensure no hex colour literal appears

- [ ] Task 3: Settings.svelte retype (AC: 3)
  - [ ] Type model dropdown against `ModelInfo[]` from `$lib/types/wails`
  - [ ] Type `EditorSettings` state
  - [ ] Remove any hardcoded model ID strings

- [ ] Task 4: SummarisationModal.svelte retype (AC: 4)
  - [ ] Define `ReviewProgressEvent` and `ReviewDoneEvent` interfaces
  - [ ] Type `EventsOn` handlers and the unsubscribe pattern
  - [ ] Preserve cancel-on-close behaviour

- [ ] Task 5: Verify reductions and behaviour (AC: 5)
  - [ ] Each target file drops to ≤ a handful of errors
  - [ ] `vitest run` — 696 tests pass
  - [ ] `vite build` clean
  - [ ] `wails dev` smoke: open Terminal, open AgentDetail, toggle a setting and save, run summarisation
  - [ ] Commit body records delta

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] `go build ./...` / `go vet ./...` / `go test ./... -race` pass
- [ ] Each target file has ≤ 5 residual errors
- [ ] `just sveltecheck-count` drops by ≥ 180 from entry
- [ ] `just sveltecheck-ratchet` passes
- [ ] `vitest run` — 696 tests pass
- [ ] `vite build` clean
- [ ] Manual wails dev smoke across Terminal / AgentDetail / Settings / SummarisationModal
- [ ] `/simplify` run on every modified file
- [ ] No `any`, no `@ts-ignore`, no `@ts-nocheck`, no hex colour literal introduced
- [ ] PR size <= 400 lines (split into up to 2 PRs — R4)
- [ ] Commit body records: `sveltecheck-count: <prev> → <cur> (Δ -<n>)`
