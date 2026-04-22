# Story svelte-check-00: Scaffolding — ratchet, hook, baseline, wails type re-exports

**Priority:** P0-critical
**Domain:** fullstack (tooling + frontend types)
**Estimated Complexity:** S
**Depends On:** none
**Status:** done
**Landed:** 1a54f7c (2026-04-22)

## Description

Establish the infrastructure that prevents svelte-check regressions while the 1520-error backlog is paid down. After this story: every commit runs svelte-check (ratcheted against a baseline), CI fails if the count goes up, and downstream stories can import Wails types under friendly names via `frontend/src/lib/types/wails.d.ts` rather than deep `wailsjs/go/main/App` imports. Establishes no-op baseline — error count is still ~1520 after this story merges.

## Developer Notes

### Architecture
- Baseline file: `docs/plans/svelte-check-baseline.md` — human-readable snapshot (1520 errors, date, how to regenerate).
- Ratchet logic: a `just sveltecheck-ratchet` recipe (new) reads the baseline number, runs `just sveltecheck-count`, and exits non-zero if `current > baseline`. Baseline is the allowed ceiling, not a floor.
- Wails type re-exports: new file `frontend/src/lib/types/wails.d.ts` that re-exports from `wailsjs/go/main/App.d.ts` and `wailsjs/go/models.ts` under friendly names (e.g. `export type { main.Agent as Agent } from '../../wailsjs/go/models'`).
- Lefthook already has `frontend-svelte-check` (verified in `lefthook.yml` lines ~40). Confirm it runs `npm run check` which is zero-error. Adjust to the ratchet recipe so pre-commit passes during migration (cannot go up from baseline) but flips to strict after Phase 7.

### Technical Considerations
- **Ratchet discipline:** `just sveltecheck-ratchet` parses `docs/plans/svelte-check-baseline.md` for the number. Use a fenced line like `baseline: 1520`. Keep parsing in pure shell (grep + awk) — no external deps.
- **CI step:** add `.github/workflows/svelte-check.yml` (or extend existing workflow) running `just sveltecheck-ratchet` on every PR. Non-blocking at first (`continue-on-error: true`) — gate flips in Phase 7 story.
- **Lefthook swap:** change `frontend-svelte-check` command from `npm run check` to `just sveltecheck-ratchet`. Without the swap, pre-commit blocks every single commit (1520 errors > 0).
- **Wails type re-exports:** Keep the file pure type re-exports (`export type ...`) — no runtime imports. This ensures the `.d.ts` compiles with `checkJs: true` and doesn't bundle.
- **R1 (Wails regen):** The re-export file must not redefine names. If `wails dev` regenerates `models.ts`, the re-exports resolve through. Document in the file header: "Do not hand-edit when bindings regenerate — types flow through."

### Risks & Edge Cases
- **R1:** Wails bindings regenerate on `wails dev`. The re-export file only re-exports, so regeneration flows through transparently. Test by deleting `wailsjs/` and re-running `wails dev`; the types must resolve.
- **Baseline drift:** If another branch merges before this story, baseline may shift. Recompute the count at merge time and update `svelte-check-baseline.md` if needed.
- **Pre-commit hang:** Running svelte-check on every staged file is slow (~30s). Ratchet runs full check once per commit — acceptable but document the timing.
- **CI non-blocking:** must use `continue-on-error: true` so the first PR landing the baseline doesn't fail; subsequent PRs inherit the ratchet.

### Reference Files
- `justfile` lines 67-84 — existing sveltecheck recipes (reuse the pattern).
- `lefthook.yml` — existing `frontend-svelte-check` hook to repoint.
- `frontend/wailsjs/go/main/App.d.ts`, `frontend/wailsjs/go/models.ts` — source of truth for re-exports.
- `frontend/jsconfig.json` — confirm `checkJs: true` setting expected by the plan.

### Skills to invoke
- `/wails` — confirm binding regeneration flow and file layout.
- `/simplify` — required before commit.

## Acceptance Criteria

AC-1: Baseline committed
- Given the current `dev` branch has 1520 svelte-check errors
- When this story lands
- Then `docs/plans/svelte-check-baseline.md` exists with `baseline: 1520` and a regeneration instruction
- And `just sveltecheck-count` still reports 1520 (no code fixes in this story)

AC-2: Ratchet recipe gates regressions
- Given `docs/plans/svelte-check-baseline.md` holds `baseline: 1520`
- When a developer introduces a new error making the count 1521
- Then `just sveltecheck-ratchet` exits non-zero
- And when the count is <= 1520, it exits 0

AC-3: Lefthook pre-commit uses the ratchet
- Given a developer has staged a frontend change
- When they run `git commit`
- Then lefthook invokes `just sveltecheck-ratchet` (not `npm run check`)
- And the commit succeeds when the count is at or below baseline
- And the commit blocks when the count rises above baseline

AC-4: CI runs the ratchet on every PR (non-blocking)
- Given a PR is opened
- When GitHub Actions runs
- Then a `svelte-check` job executes `just sveltecheck-ratchet`
- And the job runs with `continue-on-error: true` so the initial migration PRs can land

AC-5: Wails type re-exports compile
- Given `frontend/src/lib/types/wails.d.ts` exists
- When `just sveltecheck` runs
- Then the file produces no new errors
- And a downstream `.svelte` file can `import type { Agent } from '$lib/types/wails'` and receive the regenerated shape from `wailsjs/go/models.ts`

## BDD Test Scenarios

### Scenario 1: Baseline establishment
```gherkin
Feature: Svelte-check baseline ratchet

  Scenario: Baseline file records the starting count
    Given the dev branch is at commit sha X
    And "just sveltecheck-count" prints 1520
    When scaffolding lands
    Then "docs/plans/svelte-check-baseline.md" exists
    And it contains the line "baseline: 1520"
    And it references the commit sha or a date stamp

  Scenario: Ratchet allows unchanged count
    Given "docs/plans/svelte-check-baseline.md" holds "baseline: 1520"
    And "just sveltecheck-count" prints 1520
    When I run "just sveltecheck-ratchet"
    Then the exit code is 0

  Scenario: Ratchet blocks regression
    Given "docs/plans/svelte-check-baseline.md" holds "baseline: 1520"
    And a developer introduces a new error bumping the count to 1521
    When I run "just sveltecheck-ratchet"
    Then the exit code is non-zero
    And stderr includes the phrase "count rose above baseline"

  Scenario: Ratchet allows a decrease (happy path for migration stories)
    Given "docs/plans/svelte-check-baseline.md" holds "baseline: 1520"
    And a later story drops the count to 1320
    When I run "just sveltecheck-ratchet"
    Then the exit code is 0
```

### Scenario 2: Lefthook integration
```gherkin
Feature: Pre-commit ratchet enforcement

  Scenario: Commit succeeds when count is at baseline
    Given lefthook is installed and frontend-svelte-check points at "just sveltecheck-ratchet"
    And the error count is 1520
    When I stage a frontend change and run "git commit"
    Then the pre-commit hook passes

  Scenario: Commit fails when count rises
    Given the error count is 1520
    And I stage a change that introduces a new error
    When I run "git commit"
    Then the frontend-svelte-check hook fails
    And fail_text instructs me to fix types (not add @ts-ignore)
```

### Scenario 3: Wails re-exports
```gherkin
Feature: Wails type re-exports

  Scenario: Friendly names resolve
    Given "frontend/src/lib/types/wails.d.ts" re-exports Agent, Workflow, Session
    When a Svelte file imports "import type { Agent } from '$lib/types/wails'"
    Then svelte-check resolves the type without error
    And the type matches the regenerated "main.Agent" struct

  Scenario: Regeneration flows through
    Given a Go method is added to "app.go" and "wails dev" regenerates models.ts
    When svelte-check runs
    Then no edits to "wails.d.ts" are required
    And the re-exports still resolve
```

## Tasks / Subtasks

- [x] Task 1: Author baseline file (AC: 1)
  - [x] Create `docs/plans/svelte-check-baseline.md` with `baseline: 1520`, date, and regeneration instructions
  - [x] Verify `just sveltecheck-count` output matches the committed baseline

- [x] Task 2: Add `sveltecheck-ratchet` recipe (AC: 2)
  - [x] Add `sveltecheck-ratchet` target to `justfile` that reads baseline and compares against live count
  - [x] Exit 0 when `current <= baseline`, non-zero otherwise, with clear stderr message
  - [x] Add a regression unit test via a bash script under `frontend/scripts/` that simulates both outcomes

- [x] Task 3: Repoint lefthook to the ratchet (AC: 3)
  - [x] Change `lefthook.yml` `frontend-svelte-check.run` from `npm run check` to `just sveltecheck-ratchet`
  - [x] Update `fail_text` to instruct: "fix types — do NOT add `@ts-ignore` or `any`"
  - [x] Verify with a deliberately failing test commit (revert before landing)

- [x] Task 4: Add CI ratchet step (AC: 4)
  - [x] Create `.github/workflows/svelte-check.yml` (or add job to existing frontend workflow)
  - [x] Run `just sveltecheck-ratchet` on pull_request and push to dev
  - [x] Set `continue-on-error: true` (flip to false in Phase 7 story)

- [x] Task 5: Create Wails friendly-name re-exports (AC: 5)
  - [x] Add `frontend/src/lib/types/wails.d.ts` re-exporting `Agent`, `Workflow`, `Session`, `Notification`, `ProcessDef`, `PendingPrompt`, `WorkflowExecution` (and other top-level types) from `wailsjs/go/models`
  - [x] Add file header comment noting "auto-regenerated upstream — this file only re-exports"
  - [x] Confirm `just sveltecheck` count is unchanged after adding the file

- [x] Task 6: Record scaffolding delta (AC: 1)
  - [x] Commit body must include: "sveltecheck-count: 1520 → 1520 (Δ 0 — baseline)"

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests (bash harness for ratchet; tsc resolution check for Wails re-exports)
- [x] 80%+ coverage on new shell helpers (exercise both pass/fail paths)
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes (no Go changes, but must not regress)
- [x] `just sveltecheck-count` = 1520 (unchanged — this is a scaffolding-only story)
- [x] `vitest run` — all 696 tests pass
- [x] `vite build` succeeds
- [x] `/simplify` run on modified `justfile`, `lefthook.yml`, workflow YAML, and `wails.d.ts`
- [x] Code review: no CRITICAL/HIGH issues; no `@ts-ignore` / `@ts-nocheck` / `any` introduced
- [x] Commit body records: `sveltecheck-count: 1520 → 1520 (Δ 0 — baseline)`
