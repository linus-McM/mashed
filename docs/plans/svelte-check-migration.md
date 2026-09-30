# Plan — Svelte-check error migration (1520 → 0)

**Baseline:** `just sveltecheck-count` reports **1520 errors** across **~30 files** on branch `dev` as of 2026-04-22.

**Goal:** Clean `just sveltecheck` (exit 0, zero errors) without regressing runtime behaviour or the 696 vitest tests.

**Non-goal:** TypeScript strict mode. This plan fixes the existing `checkJs: true` + default svelte-check contract — not `noImplicitAny: true` across the whole bundle.

## Why now

- Silent type drift — every new prop, event, or handler adds untyped surface area.
- False green from `npm run build` / `vitest` — neither runs the Svelte template typechecker.
- The U8/U9 frontend work exposed the gap: when `DiagnosticsChip` gained a new prop, the caller's `generatedBy` field was previously untyped so the regression would not have surfaced.

## Error taxonomy

The 1520 errors collapse to 7 repeat patterns. Fixing the pattern unblocks large blocks at once.

| # | Pattern | Count (approx) | Typical fix |
|---|---------|----------------|-------------|
| P1 | `Parameter 'e' implicitly has an 'any' type` | 89 | Annotate `(e: unknown)` in catch; `(e: MouseEvent)` / `(e: CustomEvent<T>)` in handlers |
| P2 | `Variable 'x' implicitly has type 'any[]'` | ~200 | Declare `/** @type {Foo[]} */` JSDoc or `let x: Foo[] = []` in TS |
| P3 | `Element implicitly has 'any' type … index type '{}'` | ~80 | Declare map as `Record<string, Value>` or `Map<string, Value>` |
| P4 | `Property 'message' does not exist on type '{}'` | 23 | `catch (e) { const msg = e instanceof Error ? e.message : String(e) }` |
| P5 | `Property 'data' does not exist on type 'object'` | ~30 | Replace `object` with `Record<string, unknown>` or proper interface |
| P6 | `Namespace '…Svelte' has no exported member 'Turn'` | 1 | Import types from dedicated `.d.ts` / `.ts`, not from `.svelte` |
| P7 | `Left side of comma operator is unused` | ~10 | Syntax error in template (separate Svelte template issue) |

## Phase breakdown

Each phase is independently mergeable. Phase N+1 depends only on Phase N's files not regressing — no architectural coupling.

### Phase 0 — Scaffolding (½ day)

Outcomes:
- Add `just sveltecheck` to lefthook pre-commit (error-only, scoped to staged files) so regressions cannot land.
- Add a CI step running `just sveltecheck` on every PR (non-blocking initially so we can count-down).
- Add a baseline counter committed as `docs/plans/svelte-check-baseline.md` so each phase shows its delta.
- Introduce `frontend/src/lib/types/wails.d.ts` re-exporting the Wails-generated types under friendly names, to avoid deep `wailsjs/go/main/App.js` imports in handler signatures.

**Gate:** after Phase 0, `just sveltecheck-count` still reports ~1520 (baseline established, no fixes yet).

### Phase 1 — JS stores → TypeScript (1 day, ~130 errors)

Files (all under `frontend/src/lib/` or `frontend/src/`):
- `lib/stores/sessions.js` (18)
- `lib/workflowSerialisation.js` (48)
- `lib/themeConverter.js` (22)
- `main.js` (21)
- `lib/__tests__/workflowSerialisation.test.js` (34)
- `lib/themeConverter.test.js` (58)

Approach:
1. Rename `.js` → `.ts`, update imports.
2. Declare explicit input/output types (most functions are pure — `Session`, `Workflow`, `Theme`, etc.).
3. Add missing interfaces to `frontend/src/types/` (co-locate with `uiAst.ts`).
4. Migrate the two `.test.js` files in the same commit as the store rename so tests typecheck against the new surface.

**Expected reduction:** ~200 errors (incl. downstream files that import these stores and were flagged via the chain).

### Phase 2 — Catch + event-handler hygiene (½ day, ~110 errors)

Files: any Svelte component with `catch (e) { … e.message }` or `(e) => …` event handlers. Top hits: `Setup.svelte`, `BranchModal.svelte`, `SwitchBranchModal.svelte`, `MergeModal.svelte`, `ForcePushModal.svelte`.

Approach:
1. Add a single helper `frontend/src/lib/errorMessage.ts` exporting `errorMessage(e: unknown): string` with the Error/String fallback.
2. Swap every `catch (e) { … e.message }` for `errorMessage(e)`.
3. Type every DOM handler as `(e: MouseEvent)` / `(e: KeyboardEvent)` / etc. based on the usage.
4. CustomEvent dispatches: add `/** @type {import('svelte').EventDispatcher<{submit: {value: string}}>} */` JSDoc on `createEventDispatcher`.

**Expected reduction:** ~110 errors directly, plus ~50 downstream.

### Phase 3 — Record-shaped state ({} → Record) (1 day, ~280 errors)

Files: `NewSessionModal.svelte` (30), `NotificationFeed.svelte` (164), `App.svelte` (27), `StatusBadge.svelte` (2), plus neighbours.

Approach per file:
1. Find every `{}`-typed variable used as a lookup table.
2. Define the record shape: `type StatusToken = 'running' | 'open' | ...; let colors: Record<StatusToken, string> = {...}`.
3. Narrow the index expression (`status as StatusToken`) when the key comes from an external source.
4. For runtime-keyed state (e.g. `repoOrder[name] = color`), type as `Record<string, Color>`.

**Expected reduction:** ~280 errors. Biggest single-file wins: `NotificationFeed` (164) and `WorkflowBuilder` (partial, see Phase 4).

### Phase 4 — View-level untyped state (2 days, ~500 errors)

Files: `WorkflowBuilder.svelte` (162), `AgentDetail.svelte` (81), `Settings.svelte` (33), `SummarisationModal.svelte` (18), `Terminal.svelte` (50), `MonacoEditor.svelte` (58), `MarkdownEditor.svelte` (21), `CodeEditor.svelte` (25).

Approach:
1. Walk each file: annotate every `export let prop` with a JSDoc `@type`.
2. Type every local `let` that today is inferred as `any`.
3. For refs (`let container`, `let term`), type against the library surface (`let container: HTMLDivElement | null = null`, `let term: Terminal | null = null`).
4. Svelte reactive statements (`$:`) inherit types from the referenced vars — no action once the underlying `let` is typed.

**Expected reduction:** ~500 errors. This is the largest phase by volume but mostly mechanical.

### Phase 5 — BMAD components (1 day, ~220 errors)

Files: `components/bmad/CanvasPane.svelte` (58), `NodeConfigPanel.svelte` (45), `ProcessSidebar.svelte` (40), `SprintPanel.svelte` (25), `GitPanel.svelte` (24), `NodeInputSnackbarStack.svelte` (18), `ArrayEditorModal.svelte` (18).

Approach: identical to Phase 4. Isolated from Phase 4 so BMAD subsystem churn doesn't block workflow-view fixes.

### Phase 6 — BMAD component tests (1 day, ~180 errors)

Files: every `src/components/bmad/__tests__/*.test.js`. Top: `ProcessNode.status.test.js` (52), `assetWatcher.test.js` (30), `MultiFileLoaderNode.test.js` (29), `ProcessNode.multiInput.test.js` (24), `CommandNode.breadcrumb.test.js` (24), `CommandNode.status.test.js` (22).

Approach:
1. Rename `.test.js` → `.test.ts` one file at a time.
2. Type all mocks: the generated `wailsjs/go/main/App.js` has type declarations — mock against them, not `{}`.
3. The test helpers in `components/bmad/__tests__/mountSvelte.ts` are already TypeScript; re-export typed mock factories from there.

**Expected reduction:** ~180 errors.

### Phase 7 — Tail + namespace fix (½ day, <50 errors)

Whatever remains: one-offs, the TranscriptPane `Turn` import (P6), and any new errors introduced by Phase 1-6 merges.

**Gate:** `just sveltecheck-count` prints `0`.

## Sequencing

```
Phase 0 (scaffolding) ──► 1 (JS→TS stores) ──► 2 (catch+events) ──► 3 (records)
                                                                       │
                    ┌──────────────────────────────────────────────────┤
                    ▼                                                  ▼
Phase 4 (views) ──► Phase 5 (BMAD components) ──► Phase 6 (BMAD tests) ──► Phase 7 (tail)
```

Phases 1→2→3 must be serial (each reduces type surface the next depends on). Phases 4 and 5 can run in parallel by two engineers. Phase 6 depends on 5.

## Budget

| Phase | Days | Error target |
|-------|------|--------------|
| 0 | 0.5 | baseline |
| 1 | 1 | ~1320 |
| 2 | 0.5 | ~1210 |
| 3 | 1 | ~930 |
| 4 | 2 | ~430 |
| 5 | 1 | ~210 |
| 6 | 1 | ~30 |
| 7 | 0.5 | **0** |

**Total: 7.5 engineer-days** (single engineer) or **~5 calendar days** with Phases 4/5 parallelised.

## Guardrails

- Every phase keeps the 696 vitest tests green and the vite production build clean.
- Each phase ends with a single atomic commit that drops the `sveltecheck-count` number and includes the delta in the commit message.
- No `// @ts-ignore`, no `// @ts-nocheck` — if a fix requires silencing, that's a signal the type model is wrong.
- No `any` as a fix for `any`. `unknown` + narrowing instead.
- CI runs `just sveltecheck` on every PR with a ratchet: the count cannot go up from the baseline committed in `docs/plans/svelte-check-baseline.md`. After Phase 7 the ratchet flips to `--threshold error --fail-on-warnings=false` failing the build on non-zero.

## Risks

- **R1 — Wails binding types drift.** Auto-generated `wailsjs/go/main/App.d.ts` regenerates when Go methods change; phase work pinned against old types breaks on next `wails dev`. Mitigation: re-run `wails dev` at the start of each phase; treat regenerated types as the source of truth.
- **R2 — Test-file migrations break test discovery.** Vitest's glob is configured for `*.test.{js,ts}` (verify before Phase 6). If not, extend it first.
- **R3 — Hidden `any` propagation.** A single `any` in a store can mask dozens of dependent errors; fixing it sometimes *increases* the count in the short term as previously-any-inferred sites get real types flagged. Expect transient bumps of ±50 inside a phase.
- **R4 — Reviewer fatigue.** 1520-line reviews are unreadable. Each phase ships as ≤5 PRs of ≤400 lines each, one per logical group.

## Decision to make before starting

- **Single-engineer vs. sprint team.** Sprint team (scrum-master → go-svelte-test + ui-engineer) parallelises Phases 4/5 and produces per-phase sprint reports. Single-engineer is lower coordination overhead but linear calendar time.
- **Strict TypeScript for new code.** Recommend setting `strict: true` in a `tsconfig.app.json` referenced from `jsconfig.json`'s `extends`, scoped to new `.ts` files only — prevents reintroducing `any` while we clean up.
