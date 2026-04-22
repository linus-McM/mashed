# Story svelte-check-07: Tail — namespace fix + residuals + strict ratchet flip (Phase 7)

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** svelte-check-06
**Status:** ready

## Description

Close out the remaining ≤ 50 errors introduced or left behind by Phases 0-6, fix the one-off `Turn` namespace issue (P6) in `TranscriptPane`, and flip the ratchet from non-blocking to strict: CI fails on any svelte-check error, lefthook blocks any commit that reintroduces one. After this story: `just sveltecheck` exits 0, `just sveltecheck-count` prints 0, and all further type regressions are blocked at commit time.

## Developer Notes

### Architecture
- Targeted fixes:
  - **P6 namespace fix**: The single Svelte namespace issue — `Namespace '…Svelte' has no exported member 'Turn'`. Move the `Turn` type from wherever it's re-exported off a `.svelte` file to a dedicated `.d.ts` or `.ts` file (e.g. `frontend/src/types/transcript.ts`) and import from there. Affected file: likely `TranscriptPane.svelte` or neighbour referencing `Turn`.
  - **P7 template syntax**: `Left side of comma operator is unused` (~10 errors) — fix in-template (usually a stray `,` in a Svelte expression or a `{#each ... as item, i}` vs `{#each ... as item}`).
  - **Residuals**: whatever's left after Phases 0-6 land. Expect one-offs spread across views.
- Ratchet flip:
  - `docs/plans/svelte-check-baseline.md` → `baseline: 0`
  - `.github/workflows/svelte-check.yml` → `continue-on-error: false`
  - `lefthook.yml` `frontend-svelte-check.run` → `just sveltecheck` (strict zero-error check, not ratchet)
  - Remove the `just sveltecheck-ratchet` recipe (ratchet only mattered during migration)

### Technical Considerations
- **Single source of truth for Turn.** Per plan P6, the `Turn` type is re-exported from a `.svelte` file which doesn't expose its types via the namespace. Move the type definition to `frontend/src/types/transcript.ts` and import from there. Verify all imports of `Turn` across the codebase after the move.
- **Template syntax errors.** These are genuine Svelte parser issues. Inspect each site and fix — do NOT suppress.
- **Ratchet removal.** After the flip, the ratchet recipe is dead code. Remove it to avoid confusion. Update the regeneration instructions in `docs/plans/svelte-check-baseline.md` to note the migration is complete.
- **No `any` creep.** This is the last gate. If a residual requires `as unknown as T`, accept it. If it requires `as any`, fix the type model.
- **Vite + vitest still green.** This story must keep `vitest run` at 696 and `vite build` clean.

### Risks & Edge Cases
- **R1 — Wails regen.** One last `wails dev` to confirm bindings are fresh. Re-run the full suite.
- **R3 — hidden residuals.** A previously-masked `any` might surface in this phase. Budget 30% slack for unexpected one-offs.
- **Ratchet flip ordering.** If the flip happens before count = 0, CI starts failing. Enforce: run `just sveltecheck-count` first, confirm it prints `0`, THEN flip the ratchet in the same PR.
- **R4 — PR sizing.** This story should fit in a single small PR.
- **Pre-existing template syntax issues.** The `Left side of comma operator` errors may hide genuine bugs (e.g. `let a = b, c = d` where one was expected to be a reactive assignment). Smoke the component runtime after fixing.

### Reference Files
- `frontend/src/components/bmad/TranscriptPane.svelte` (if it exists) or equivalent — source of P6 issue
- `docs/plans/svelte-check-baseline.md` (from Story 00) — update to `baseline: 0`
- `.github/workflows/svelte-check.yml` (from Story 00) — flip to strict
- `lefthook.yml` (from Story 00) — swap to `just sveltecheck`
- `justfile` — remove the `sveltecheck-ratchet` recipe after the flip

### Skills to invoke
- `/simplify` — mandatory before commit.
- `/wails` — final regen sanity check.
- `/playwright-cli` — smoke TranscriptPane if it exists.

## Acceptance Criteria

AC-1: Namespace P6 fix
- Given the `Turn` type currently imported from a `.svelte` namespace
- When this story lands
- Then `Turn` lives in `frontend/src/types/transcript.ts` (or equivalent dedicated `.ts`/`.d.ts`)
- And every import of `Turn` references the dedicated module
- And no `Namespace '…Svelte' has no exported member 'Turn'` error remains

AC-2: Residuals cleared
- Given the remaining ≤ 50 errors after Phase 6
- When this story lands
- Then `just sveltecheck` exits 0
- And `just sveltecheck-count` prints `0`
- And no `@ts-ignore` / `@ts-nocheck` / `any` was added as a shortcut

AC-3: Baseline flipped to zero
- Given `docs/plans/svelte-check-baseline.md`
- When this story lands
- Then the file holds `baseline: 0`
- And a note records the migration-complete date

AC-4: CI flipped to strict
- Given `.github/workflows/svelte-check.yml`
- When this story lands
- Then the job runs `just sveltecheck` (not the ratchet)
- And `continue-on-error` is false
- And a PR that introduces a single svelte-check error fails CI

AC-5: Lefthook flipped to strict
- Given `lefthook.yml`
- When this story lands
- Then `frontend-svelte-check` runs `just sveltecheck` (not the ratchet)
- And any staged change introducing an error blocks the commit
- And the `sveltecheck-ratchet` recipe is removed from `justfile`

## BDD Test Scenarios

### Scenario 1: Turn namespace fix
```gherkin
Feature: Turn type moved out of Svelte namespace

  Scenario: Direct import works
    Given "types/transcript.ts" exports "Turn"
    When TranscriptPane.svelte imports "import type { Turn } from '$types/transcript'"
    Then svelte-check resolves without namespace error

  Scenario: Old namespace import rejected
    Given any remaining "import type { Turn } from './TranscriptPane.svelte'"
    When svelte-check runs
    Then no such import remains in the codebase
```

### Scenario 2: Zero errors
```gherkin
Feature: Migration complete

  Scenario: sveltecheck exits 0
    Given Phase 7 lands
    When "just sveltecheck" runs
    Then exit code is 0
    And no output lines start with "ERROR"

  Scenario: sveltecheck-count is 0
    Given Phase 7 lands
    When "just sveltecheck-count" runs
    Then output is exactly "0"
```

### Scenario 3: Strict ratchet
```gherkin
Feature: Strict svelte-check gate

  Scenario: CI fails on new error
    Given a PR introduces one deliberate svelte-check error
    When GitHub Actions runs the "svelte-check" job
    Then the job fails
    And the PR is blocked from merging

  Scenario: Pre-commit blocks new error
    Given a staged change introduces one svelte-check error
    When "git commit" is run
    Then lefthook reports the error via "just sveltecheck"
    And the commit is aborted

  Scenario: Ratchet recipe removed
    Given the migration is complete
    When "just -l" is invoked
    Then "sveltecheck-ratchet" is NOT listed
    And "sveltecheck" IS listed
```

### Scenario 4: No cheats
```gherkin
Feature: Clean code only

  Scenario: No @ts-ignore introduced
    Given the full final codebase
    When "rg '@ts-(ignore|nocheck)' frontend/src/" is run
    Then no matches are found

  Scenario: No lingering any
    Given the full final codebase
    When "rg ': any\b' frontend/src/" is run
    Then no production-code matches remain (excluding monaco-editor type shims or documented exceptions)
```

## Tasks / Subtasks

- [ ] Task 1: Fix P6 namespace — Turn (AC: 1)
  - [ ] Create `frontend/src/types/transcript.ts` exporting `Turn` (and any neighbour types)
  - [ ] Update all imports across the codebase to the new path
  - [ ] Remove the type re-export from the `.svelte` file
  - [ ] Verify with `just sveltecheck-file` on TranscriptPane

- [ ] Task 2: Fix remaining P7 comma-operator syntax errors (AC: 2)
  - [ ] Locate every `Left side of comma operator` error
  - [ ] Fix the underlying template expression (likely `{#each arr as item, i}` or stray comma in reactive assignment)
  - [ ] Smoke the affected component at runtime

- [ ] Task 3: Sweep residual errors (AC: 2)
  - [ ] Run `just sveltecheck-by-file`; for each file with > 0 errors, fix in-place
  - [ ] No suppressions — fix the type model

- [ ] Task 4: Flip baseline + CI + lefthook to strict (AC: 3, 4, 5)
  - [ ] Update `docs/plans/svelte-check-baseline.md` to `baseline: 0` + migration-complete note
  - [ ] Update `.github/workflows/svelte-check.yml` → run `just sveltecheck`, `continue-on-error: false`
  - [ ] Update `lefthook.yml` → `frontend-svelte-check.run: just sveltecheck`
  - [ ] Remove the `sveltecheck-ratchet` recipe from `justfile`

- [ ] Task 5: Final verification (AC: 2)
  - [ ] `just sveltecheck` exits 0
  - [ ] `just sveltecheck-count` prints `0`
  - [ ] `vitest run` — 696 tests pass
  - [ ] `vite build` clean
  - [ ] `wails dev` smoke — app launches end-to-end
  - [ ] Commit body records: `sveltecheck-count: <prev> → 0 (Δ -<n>) — MIGRATION COMPLETE`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] `go build ./...` / `go vet ./...` / `go test ./... -race` pass
- [ ] `just sveltecheck` exits 0
- [ ] `just sveltecheck-count` prints `0`
- [ ] `vitest run` — 696 tests pass
- [ ] `vite build` clean
- [ ] `rg '@ts-(ignore|nocheck)' frontend/src/` — 0 matches
- [ ] CI strict mode verified with a deliberate failure commit (reverted before merge)
- [ ] Pre-commit strict mode verified locally
- [ ] `/simplify` run on every modified file
- [ ] No `any`, no `@ts-ignore`, no `@ts-nocheck` introduced
- [ ] PR size <= 400 lines
- [ ] Commit body records: `sveltecheck-count: <prev> → 0 (Δ -<n>) — MIGRATION COMPLETE`
