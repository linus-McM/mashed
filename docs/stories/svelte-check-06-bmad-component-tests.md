# Story svelte-check-06: BMAD component tests (.test.js → .test.ts) (Phase 6)

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** svelte-check-05
**Status:** ready

## Description

Migrate every `.test.js` file under `frontend/src/components/bmad/__tests__/` to `.test.ts`, retyping the mocks against the regenerated `wailsjs/go/main/App.d.ts`. Top single-file counts are `ProcessNode.status.test.js` (52), `assetWatcher.test.js` (30), `MultiFileLoaderNode.test.js` (29), `ProcessNode.multiInput.test.js` (24), `CommandNode.breadcrumb.test.js` (24), `CommandNode.status.test.js` (22). Expected reduction ~180, from ~200 to ~30.

## Developer Notes

### Architecture
- Test files to migrate (approximate counts from plan):
  - `components/bmad/__tests__/ProcessNode.status.test.js` (52)
  - `components/bmad/__tests__/assetWatcher.test.js` (30)
  - `components/bmad/__tests__/MultiFileLoaderNode.test.js` (29)
  - `components/bmad/__tests__/ProcessNode.multiInput.test.js` (24)
  - `components/bmad/__tests__/CommandNode.breadcrumb.test.js` (24)
  - `components/bmad/__tests__/CommandNode.status.test.js` (22)
  - Plus any other `.test.js` files in that directory (ProcessSidebar.colors.test.js, SkillEditorModal.colors.test.js, CommandNode.colors.test.js, CanvasFailureToast.colors.test.js, CanvasPane.drop.test.js are also present per the directory listing)
- Shared typed mount helper: `frontend/src/components/bmad/__tests__/mountSvelte.ts` (already TypeScript). Extend it with typed mock factories.
- Wails mocks: the generated `App.d.ts` is the authoritative mock target. Define `createAppMock(): typeof import('../../../wailsjs/go/main/App')` helper in `mountSvelte.ts`.

### Technical Considerations
- **R2 — vitest glob verified.** `frontend/vitest.config.js` uses default globs (`*.test.{js,ts,jsx,tsx}` — vitest default). Files continue to be discovered after `.js` → `.ts` rename. Confirm by running `vitest run` after the first file migration.
- **Rename atomically.** Git-mv one file at a time and commit per file (or per cluster). Never half-rename.
- **Mock factory pattern.** Every test calls `vi.mock('../../../wailsjs/go/main/App', () => createAppMock({ StartBmadRun: vi.fn() }))`. The factory returns a fully typed object; individual overrides are type-checked against `App.d.ts`.
- **Asset watcher tests.** `assetWatcher.test.js` likely tests fsnotify-driven reload (per cerebrum Sprint 4 file-aware context). It probably stubs the `asset:*` events. Type those event payloads.
- **ProcessNode.status** exercises the 7-state status machine. The status enum is domain-sourced from `internal/domain/agent.go` — reuse `StatusToken` from Phase 3's `types/status.ts`.
- **No `any` in mocks.** Use `vi.fn<[arg: Arg], Ret>()` with generic signatures to preserve argument and return typing.

### Risks & Edge Cases
- **R2 — vitest glob.** If the glob somehow doesn't catch `.test.ts`, fix `vitest.config.js` BEFORE renaming any file. Verify by adding one trivial `.test.ts` file with a `.skip()` test and checking it's picked up.
- **R1 — Wails regen.** The mock factory is generated from `App.d.ts`. If bindings regenerate mid-phase, regenerate the mock factory interface too.
- **Test count preservation.** The 696 total must not change. Every renamed file must have all its assertions pass. If a hidden `any` in the test masked a stale assertion, fix the assertion — never widen with `as any`.
- **Coverage floor.** This story touches test files; the 80% coverage rule applies to new production code, not to test files. BMAD component production code was retyped in Phase 5 — tests still must exercise it.
- **R4 — PR sizing.** ~180 errors across 6-11 files likely fits in 1-2 PRs. Plan split if > 400 lines:
  - PR-A: ProcessNode.*.test.ts + CommandNode.*.test.ts (status/multiInput/breadcrumb/colors)
  - PR-B: assetWatcher + MultiFileLoaderNode + remaining

### Reference Files
- `frontend/src/components/bmad/__tests__/mountSvelte.ts` (typed helper — extend, don't duplicate)
- `frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts` (fully-typed example from U8)
- `frontend/src/components/bmad/__tests__/ProcessNode.awaiting.test.ts` (interactive, typed)
- `frontend/wailsjs/go/main/App.d.ts` (mock target authority)
- `.wolf/cerebrum.md` BMAD executor + interactive sections (status machine, node types)

### Skills to invoke
- `/simplify` — mandatory before commit.
- `/playwright-cli` — not needed for unit tests.

## Acceptance Criteria

AC-1: All BMAD test files are `.test.ts`
- Given the `frontend/src/components/bmad/__tests__/` directory
- When this story lands
- Then `ls *.test.js` in that directory returns nothing (all renamed)
- And each renamed file has zero svelte-check errors

AC-2: Typed mock factory exists
- Given `mountSvelte.ts`
- When this story lands
- Then it exports a typed `createAppMock()` (or similar) producing a `typeof` `wailsjs/go/main/App` object
- And the factory accepts overrides typed against `App.d.ts`

AC-3: Vitest glob verified and tests discovered
- Given the renamed files
- When `vitest run` executes
- Then all previously-discovered tests (696) are discovered
- And the pass count is identical to pre-story
- And no test is silently skipped

AC-4: Error count drops
- Given entering at ~200
- When this story lands
- Then `just sveltecheck-count` is ≤ 30 (approx Phase 7 entry)
- And `just sveltecheck-ratchet` passes

AC-5: No `any` or type silencing in tests
- Given every renamed test file
- When inspected
- Then no file contains `any`, `@ts-ignore`, `@ts-nocheck`, or `as any`
- And every mock function uses `vi.fn<...>()` generics

## BDD Test Scenarios

### Scenario 1: Vitest glob safety
```gherkin
Feature: Vitest discovers .test.ts

  Scenario: Trivial .test.ts is discovered
    Given a file "pilot.test.ts" with "it.skip('noop', () => {})"
    When vitest run is executed
    Then the pilot test appears in the reporter output as skipped
    And the total discovered test count is baseline + 1

  Scenario: Glob covers both extensions during migration
    Given .test.js and .test.ts files coexist
    When vitest run is executed
    Then both are discovered
    And no test is duplicated
```

### Scenario 2: Typed mock factory
```gherkin
Feature: Typed createAppMock

  Scenario: Override type-checked
    Given createAppMock({ StartBmadRun: vi.fn() })
    When a caller passes StartBmadRun: vi.fn<[wrong: number], void>()
    Then svelte-check reports a type error because StartBmadRun signature mismatches
    And the correct signature from App.d.ts passes

  Scenario: Unreferenced methods default to vi.fn()
    Given createAppMock({}) with no overrides
    When a component calls App.ListSessions()
    Then the mock returns a default value (empty array or resolved undefined)
    And the return type is Promise<Session[]>
```

### Scenario 3: Test behaviour preserved
```gherkin
Feature: Test assertions unchanged

  Scenario: ProcessNode.status test still asserts correct colours
    Given ProcessNode.status.test.js had 10 assertions
    When renamed to .test.ts and run
    Then all 10 assertions pass
    And no assertion is skipped, disabled, or converted to a softer form

  Scenario: assetWatcher event payloads typed
    Given assetWatcher.test.js mocked "asset:changed" events as { type: 'x', path: 'y' }
    When renamed to .test.ts with typed event payload
    Then the mock conforms to AssetChangedEvent interface
    And the test still passes
```

### Scenario 4: Error delta
```gherkin
Feature: Phase 6 reduction

  Scenario: Count drops to ≤ 30
    Given entering at ~200
    When Phase 6 lands
    Then "just sveltecheck-count" ≤ 30
    And the ratchet passes
```

## Tasks / Subtasks

- [ ] Task 1: Verify R2 — vitest glob safety (AC: 3)
  - [ ] Confirm `frontend/vitest.config.js` uses default glob (include `**/*.test.{js,ts}`)
  - [ ] Add a pilot `.test.ts` as a smoke and confirm discovery, then remove before merge

- [ ] Task 2: Extend mountSvelte with typed mock factory (AC: 2)
  - [ ] Add `createAppMock(overrides?: Partial<typeof App>): typeof App`
  - [ ] Ensure it forwards the generic signatures from `App.d.ts`
  - [ ] Export typed shared mocks for common event payloads

- [ ] Task 3: Migrate ProcessNode tests (AC: 1, 5)
  - [ ] Rename `ProcessNode.status.test.js` → `.test.ts`, retype mocks
  - [ ] Rename `ProcessNode.multiInput.test.js` → `.test.ts`
  - [ ] Confirm all assertions pass

- [ ] Task 4: Migrate CommandNode tests (AC: 1, 5)
  - [ ] Rename `CommandNode.status.test.js`, `CommandNode.breadcrumb.test.js`, `CommandNode.colors.test.js` → `.test.ts`
  - [ ] Retype mocks

- [ ] Task 5: Migrate remaining tests (AC: 1, 5)
  - [ ] `assetWatcher.test.js`, `MultiFileLoaderNode.test.js`, `ProcessSidebar.colors.test.js`, `SkillEditorModal.colors.test.js`, `CanvasFailureToast.colors.test.js`, `CanvasPane.drop.test.js`
  - [ ] Retype mocks against `App.d.ts`
  - [ ] Preserve all assertion behaviour

- [ ] Task 6: Verify reduction (AC: 4)
  - [ ] `just sveltecheck-count` ≤ 30
  - [ ] `vitest run` — all 696 tests pass
  - [ ] `vite build` clean
  - [ ] Commit body records delta

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] `go build ./...` / `go vet ./...` / `go test ./... -race` pass
- [ ] `just sveltecheck-count` ≤ 30
- [ ] `just sveltecheck-ratchet` passes
- [ ] `vitest run` — all 696 tests pass, no skipped tests
- [ ] `vite build` clean
- [ ] `ls frontend/src/components/bmad/__tests__/*.test.js` returns nothing
- [ ] `/simplify` run on modified files
- [ ] No `any`, `@ts-ignore`, `@ts-nocheck`, or `as any` in any renamed file
- [ ] PR size <= 400 lines (split into up to 2 PRs — R4)
- [ ] Commit body records: `sveltecheck-count: <prev> → <cur> (Δ -<n>)`
