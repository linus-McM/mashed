# Story svelte-check-01: JS stores → TypeScript (Phase 1)

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** svelte-check-00
**Status:** ready

## Description

Convert the four pure-logic JavaScript modules and their two companion test files to TypeScript with explicit types for their inputs and outputs. Because these modules are imported across many Svelte views, fixing their type surface cascades error reductions through the whole frontend. Expected error drop: ~130 direct + ~70 downstream = ~200, taking the count from 1520 to approximately 1320.

## Developer Notes

### Architecture
- Files to rename `.js` → `.ts` (in the same commit for each file pair):
  - `frontend/src/lib/stores/sessions.js` → `.ts` (18 errors)
  - `frontend/src/lib/workflowSerialisation.js` → `.ts` (48 errors)
  - `frontend/src/lib/themeConverter.js` → `.ts` (22 errors)
  - `frontend/src/main.js` → `.ts` (21 errors)
  - `frontend/src/lib/__tests__/workflowSerialisation.test.js` → `.ts` (34 errors)
  - `frontend/src/lib/themeConverter.test.js` → `.ts` (58 errors)
- Shared type declarations land in `frontend/src/types/` (next to `uiAst.ts`):
  - `frontend/src/types/session.ts` — `Session`, `SessionState`
  - `frontend/src/types/workflow.ts` — `Workflow`, `WorkflowNode`, `WorkflowEdge` (mirroring Go structs in `internal/bmad/types.go`)
  - `frontend/src/types/theme.ts` — `Theme`, `ThemeColors`, `MonacoThemeDef`
- Rely on Wails-generated `$lib/types/wails` re-exports from Story 00 where the domain types are Go-sourced (e.g. `Workflow`, `WorkflowNode`, `WorkflowEdge` re-export `main.Workflow` shapes when they already exist in `wailsjs/go/models.ts`).

### Technical Considerations
- **No `any`.** When input shape is unknown (e.g. deserialised JSON payload) type as `unknown` then narrow with `typeof` / `in` / `Array.isArray`.
- **Imports update:** after rename, every `.svelte` and `.ts` file importing these modules needs the import-path `.js` extension stripped if present, or updated to `.ts`. Most imports are extensionless and will resolve automatically.
- **Test file migrations:** the two `.test.js` files stay in the same directory and use the default vitest glob (`*.test.{js,ts}`). Confirmed vitest.config.js has no custom include array restricting to `.js` only (R2 cleared).
- **Theme shape:** `themeConverter` maps monaco-editor themes to app tokens; align the return type with the existing `monacoTheme.js` surface to avoid downstream break.
- **Workflow serialisation:** `serialise` / `deserialise` take `Workflow` and return a canonical string. Type the internal intermediate as `Record<string, unknown>` then narrow before assignment to `Workflow` fields.
- **main.ts:** bootstrap file — types should be minimal; mostly `mount` call plus Wails runtime event wiring. Event payloads typed from `$lib/types/wails`.

### Risks & Edge Cases
- **R3 — transient error bumps.** Once a store is strongly typed, previously-any-inferred call sites in downstream `.svelte` files will flag their own inconsistencies. Expected bump of up to +50 mid-phase before net drop of ~200. Run `just sveltecheck-count` after each file rename to track.
- **R1 — Wails types regen.** If `wails dev` regenerates `models.ts` during this story, re-import from the wails re-exports; do not duplicate Go-sourced types locally.
- **Test discovery:** verify `vitest run` after the two `.test.js` → `.ts` renames still discovers all tests (should be 696; no loss tolerated).
- **R4 — reviewer fatigue.** Split this story into up to 3 PRs if the diff exceeds 400 lines:
  - PR-A: `types/` new files + `themeConverter.ts` + test
  - PR-B: `sessions.ts` + `workflowSerialisation.ts` + test
  - PR-C: `main.ts` + any downstream import updates
- **Circular imports:** `sessions.ts` likely imports from `workflowSerialisation` or vice versa — rename both together to avoid `.js`/`.ts` duality confusing the resolver.

### Reference Files
- `frontend/src/lib/bmadSessionName.ts` — existing TS module in the same folder, use as a template for export style and JSDoc headers.
- `frontend/src/lib/stores/uiAdapterSettings.ts` — example of a typed Svelte store with `writable<T>`; follow this pattern for `sessions.ts`.
- `frontend/src/types/uiAst.ts` — shape of existing type modules; co-locate new ones.
- `frontend/wailsjs/go/models.ts` — authoritative Go→TS types; re-export through `$lib/types/wails` (from Story 00).

### Skills to invoke
- `/simplify` — mandatory before commit.
- `/wails` — when types need to line up with regenerated bindings.
- `/playwright-cli` — not needed for this story (no UI change).

## Acceptance Criteria

AC-1: Files renamed and compile
- Given the six `.js` files listed in Developer Notes
- When this story lands
- Then each has been renamed to `.ts` (or `.test.ts`)
- And each file has no `any` type (explicit or inferred-as-any) in its annotations
- And each file has no `// @ts-ignore` or `// @ts-nocheck`

AC-2: Typed exports
- Given the renamed modules
- When their public functions are inspected
- Then every exported function has a typed signature (parameters and return)
- And every exported `writable` store is typed with an explicit generic (`writable<Session[]>([])`)

AC-3: Domain types co-located
- Given the frontend `types/` directory
- When this story lands
- Then `types/session.ts`, `types/workflow.ts`, `types/theme.ts` exist
- And each re-exports or augments shapes available via `$lib/types/wails` where a Go source of truth exists

AC-4: Error count drops
- Given baseline is 1520
- When `just sveltecheck-count` is run after this story lands
- Then the count is <= 1320 (a ~200-error drop from cascading store fixes)
- And `just sveltecheck-ratchet` passes

AC-5: Tests still pass
- Given the two `.test.js` files were renamed to `.test.ts`
- When `vitest run` executes
- Then all 696 tests are discovered and pass
- And `vite build` succeeds

## BDD Test Scenarios

### Scenario 1: Pure-module rename
```gherkin
Feature: JS → TS conversion of pure modules

  Scenario: themeConverter rename preserves behaviour
    Given themeConverter.js exports "monacoToAppTheme"
    When themeConverter.ts replaces it
    Then "themeConverter.test.ts" passes all assertions unchanged
    And the exported function has signature "(theme: MonacoThemeDef) => Theme"

  Scenario: workflowSerialisation rename preserves behaviour
    Given a canonical workflow JSON fixture
    When serialise(fixture) → deserialise(...) → serialise(...) is round-tripped
    Then the final string equals the first string (idempotent)
    And the intermediate Workflow object satisfies the Workflow interface
```

### Scenario 2: Typed store
```gherkin
Feature: Typed sessions store

  Scenario: sessions.ts exposes typed writable
    Given "sessions.ts" exports "sessions: Writable<Session[]>"
    When a Svelte view subscribes "$sessions"
    Then the subscription value is typed as "Session[]" (no implicit any)
    And "sessions.update(arr => [...arr, newSession])" flags a type error when newSession is not a Session

  Scenario: Session shape matches Go
    Given "Session" type in "types/session.ts"
    When compared to "main.Session" in "wailsjs/go/models.ts"
    Then every field on the TS type matches a field on the Go type (no drift)
```

### Scenario 3: Error delta
```gherkin
Feature: Phase 1 error reduction

  Scenario: Count drops by at least 180
    Given baseline 1520
    When Phase 1 lands
    Then "just sveltecheck-count" is at most 1340
    And the ratchet passes

  Scenario: Test count preserved
    Given 696 vitest tests pass on baseline
    When Phase 1 lands
    Then "vitest run" reports 696 passing tests
    And no tests are skipped or lost in rename
```

### Scenario 4: Edge case — unknown payloads
```gherkin
Feature: Unknown narrowing in deserialisation

  Scenario: deserialise rejects malformed JSON
    Given deserialise(input: string) returns "Workflow | null"
    When called with "{}" (missing required fields)
    Then it returns null
    And never throws
    And never returns an "any"-shaped object
```

## Tasks / Subtasks

- [ ] Task 1: Create shared type modules (AC: 3)
  - [ ] Add `frontend/src/types/session.ts` — `Session`, `SessionState` (re-export from `$lib/types/wails` where possible)
  - [ ] Add `frontend/src/types/workflow.ts` — `Workflow`, `WorkflowNode`, `WorkflowEdge`
  - [ ] Add `frontend/src/types/theme.ts` — `Theme`, `MonacoThemeDef`

- [ ] Task 2: Convert `themeConverter` (AC: 1, 2, 5)
  - [ ] Rename `themeConverter.js` → `themeConverter.ts`, annotate signatures
  - [ ] Rename `themeConverter.test.js` → `themeConverter.test.ts`, update imports
  - [ ] Confirm `vitest run themeConverter` passes

- [ ] Task 3: Convert `workflowSerialisation` (AC: 1, 2, 5)
  - [ ] Rename `workflowSerialisation.js` → `workflowSerialisation.ts`
  - [ ] Type intermediate payloads as `unknown` + narrowing; reject malformed input explicitly
  - [ ] Rename `__tests__/workflowSerialisation.test.js` → `.test.ts`, retype fixture helpers

- [ ] Task 4: Convert `sessions` store (AC: 1, 2)
  - [ ] Rename `lib/stores/sessions.js` → `.ts` with `writable<Session[]>([])`
  - [ ] Update all Svelte views that subscribe — most will work without change, but verify `App.svelte`, `NotificationFeed.svelte`, `AgentDetail.svelte`

- [ ] Task 5: Convert `main.js` (AC: 1, 2)
  - [ ] Rename `main.js` → `main.ts`
  - [ ] Type Wails runtime event payloads via `$lib/types/wails`
  - [ ] Update `index.html` script src if it references `main.js` explicitly (expected: `src/main.ts`)

- [ ] Task 6: Verify cascade reduction (AC: 4)
  - [ ] Run `just sveltecheck-count` — confirm <= 1340
  - [ ] Run `just sveltecheck-by-file` — confirm the six renamed files each report 0 errors
  - [ ] Record delta in commit body: `sveltecheck-count: 1520 → <N> (Δ -<M>)`

- [ ] Task 7: Verify regressions absent (AC: 5)
  - [ ] `vitest run` — 696 tests pass
  - [ ] `vite build` — clean
  - [ ] `wails dev` smoke — app launches, sessions list renders, theme toggles work

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on modified files (existing tests suffice — do not add just-for-coverage tests)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `just sveltecheck-count` <= 1340 (target 1320)
- [ ] `just sveltecheck-ratchet` passes
- [ ] `vitest run` — 696 tests pass
- [ ] `vite build` clean
- [ ] `/simplify` run on every modified file
- [ ] Code review: no `any`, no `@ts-ignore`, no `@ts-nocheck`
- [ ] PR size <= 400 lines (split into multiple PRs if needed — R4)
- [ ] Commit body records: `sveltecheck-count: 1520 → <N> (Δ -<M>)`
