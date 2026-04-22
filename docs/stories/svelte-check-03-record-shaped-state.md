# Story svelte-check-03: Record-shaped state — `{}` → `Record<K, V>` (Phase 3)

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** svelte-check-02
**Status:** ready

## Description

Replace every `{}`-typed lookup table and accumulator across the high-volume view components with precise `Record<K, V>` (or `Map<K, V>`) shapes. NotificationFeed.svelte alone is 164 errors — by far the biggest single-file win in the migration. After this story, ~280 errors (P2 + P3 patterns) are eliminated, dropping the count from ~1210 to ~930.

## Developer Notes

### Architecture
- Target files (error counts in parentheses — current state; confirm with `just sveltecheck-by-file`):
  - `frontend/src/views/NotificationFeed.svelte` (164)
  - `frontend/src/views/NewSessionModal.svelte` (30)
  - `frontend/src/App.svelte` (27)
  - `frontend/src/components/StatusBadge.svelte` (2)
  - Any neighbour files surfaced by `sveltecheck-by-file` that share the P2/P3 pattern
- Pattern 1 — closed set of keys (use string literal union):
  ```ts
  type StatusToken = 'running' | 'open' | 'blocked' | 'done' | 'waiting';
  const colors: Record<StatusToken, string> = {
    running: 'var(--accent-green)',
    open: 'var(--accent-amber)',
    blocked: 'var(--accent-red)',
    done: 'var(--accent-teal)',
    waiting: 'var(--muted-foreground)',
  };
  const color = colors[status as StatusToken];
  ```
- Pattern 2 — runtime-keyed accumulator (use broad string key):
  ```ts
  let repoOrder: Record<string, number> = {};
  repoOrder[name] = index;
  ```
- Pattern 3 — when order matters or keys may be non-string, use `Map<string, V>` or `Map<number, V>`.

### Technical Considerations
- **Narrow at the boundary, not at every access.** Cast `(status as StatusToken)` where the string enters from an external source (Wails event payload, DOM data attribute). Don't pepper casts across the component.
- **StatusToken must match backend.** Cross-check against `internal/domain/agent.go` status constants. Do not invent tokens — if `'unknown'` isn't emitted by the backend, don't include it.
- **Color tokens must use design system.** Per cerebrum 2026-04-10, never hardcode `#39ff14`. Always use `var(--accent-green)` etc. This story must NOT introduce hex literals — use tokens.
- **NotificationFeed is the dominant file.** It has 164 errors — typically many `{}`-typed running tallies (per-repo counts, per-severity counts, per-agent history). Map each to a `Record<string, number>` or `Record<NotificationSeverity, ...>`.
- **Window type augmentations** — if `window.wails` or custom globals are typed as `{}`, add a declaration to `frontend/src/app.d.ts` (create if missing) with `declare global { interface Window { … } }`.

### Risks & Edge Cases
- **R3 — transient bumps.** Typing a Record correctly sometimes uncovers missing-key accesses that were silently returning `undefined`. The correct fix is to use `| undefined` on the value type and handle the missing case, not to widen the key type back to `string`.
- **Runtime-keyed maps can leak memory** if never pruned. Not a svelte-check concern but flag any unbounded `Record<string, T>` for a follow-up — cerebrum notes the token-sample pruning pattern in `app_scan.go`.
- **`Record<K, V>` vs `Partial<Record<K, V>>`** — if not every key is guaranteed, wrap in `Partial`. Example: a lookup built lazily should be `Partial<Record<StatusToken, string>>` so TypeScript flags missing-key access.
- **Keyof narrowing bug** (TS4.x): `const key = 'x' as StatusToken; colors[key]` returns `string` (the value type), not `string | undefined`, because `Record` assumes exhaustive. Use `Partial<Record<...>>` for the lazy case.
- **R4 — PR sizing.** NotificationFeed alone could push the diff above 400 lines. Plan for 2 PRs:
  - PR-A: NotificationFeed + StatusBadge (status tokens family)
  - PR-B: NewSessionModal + App.svelte + neighbours

### Reference Files
- `frontend/src/components/StatusBadge.svelte` — already uses tokens correctly; type the internal color map after its pattern.
- `frontend/src/lib/stores/uiAdapterSettings.ts` — example of a narrow value type from Story 01.
- `internal/domain/agent.go` — authoritative agent status tokens (cross-check before defining `StatusToken`).
- `.wolf/cerebrum.md` entry dated 2026-04-10 — `#39ff14` ban and token enforcement.

### Skills to invoke
- `/simplify` — mandatory before commit.
- `/playwright-cli` — smoke NotificationFeed rendering (real Wails event → feed update still works).

## Acceptance Criteria

AC-1: Literal-union status types defined
- Given existing `{}`-typed status color maps
- When this story lands
- Then `frontend/src/types/status.ts` (or equivalent) exports a `StatusToken` union
- And the union matches the backend's emitted status strings from `internal/domain/agent.go`

AC-2: Record-shaped state explicit
- Given every identified `{}`-typed lookup table in the target files
- When this story lands
- Then each is annotated as `Record<K, V>` or `Partial<Record<K, V>>` or `Map<K, V>`
- And no `let x: {} = {}` or untyped `let x = {}` remains in the target files

AC-3: No hardcoded colors introduced
- Given the Record value types reference theme tokens
- When this story lands
- Then `rg "#(39ff14|[0-9a-fA-F]{6})" frontend/src/views/ frontend/src/components/` finds no new matches versus baseline
- And all colour values are `var(--…)` tokens

AC-4: NotificationFeed drops by >=150 errors
- Given NotificationFeed has 164 errors on entry
- When this story lands
- Then `just sveltecheck-file src/views/NotificationFeed.svelte` reports <= 14 errors
- And the runtime behaviour (feed rendering, per-repo grouping) is unchanged

AC-5: Overall count drops
- Given baseline entering this story is ~1210
- When it lands
- Then `just sveltecheck-count` reports <= 930
- And `just sveltecheck-ratchet` passes

## BDD Test Scenarios

### Scenario 1: Literal-union enforcement
```gherkin
Feature: Typed status colour lookup

  Scenario: Invalid status key caught at compile time
    Given "Record<StatusToken, string>" where StatusToken = 'running' | 'open' | 'done'
    When a caller writes colors["pending"]
    Then svelte-check reports a type error

  Scenario: External string narrowed at boundary
    Given event.payload.status is typed as string
    When the caller writes colors[event.payload.status as StatusToken]
    Then svelte-check passes
    And the narrowing is confined to one spot

  Scenario: Partial Record flags missing keys
    Given "Partial<Record<StatusToken, string>>"
    When the caller reads colors["running"]
    Then the returned type is "string | undefined"
    And callers must handle undefined explicitly
```

### Scenario 2: Runtime-keyed accumulator
```gherkin
Feature: Record-shaped accumulator

  Scenario: repoOrder lookup typed
    Given "let repoOrder: Record<string, number> = {}"
    When entries are assigned "repoOrder[repo] = i"
    Then svelte-check passes
    And reads "const n = repoOrder[name]" type n as "number | undefined" (when marked Partial) or "number"
```

### Scenario 3: No colour regressions
```gherkin
Feature: Design token enforcement preserved

  Scenario: No #39ff14 introduced
    Given the cerebrum 2026-04-10 ban on "#39ff14"
    When Phase 3 lands
    Then "rg #39ff14 frontend/src/" finds no new matches vs baseline

  Scenario: All new Record values use tokens
    Given every new Record<K, string> for colours
    When inspected
    Then every value starts with "var(--" or references a token constant
```

### Scenario 4: NotificationFeed runtime preserved
```gherkin
Feature: NotificationFeed behaviour unchanged

  Scenario: Notification grouping still works
    Given a sequence of 5 notifications across 2 repos
    When NotificationFeed renders after retyping
    Then 2 repo groups are visible
    And each group lists its notifications in the same order as before
    And no console errors appear
```

## Tasks / Subtasks

- [ ] Task 1: Define shared status types (AC: 1)
  - [ ] Create `frontend/src/types/status.ts` with `StatusToken` literal union
  - [ ] Cross-check tokens against `internal/domain/agent.go`
  - [ ] Export `isStatusToken(s: unknown): s is StatusToken` narrowing helper

- [ ] Task 2: Retype NotificationFeed (AC: 2, 3, 4)
  - [ ] Walk every `let … = {}` and annotate `Record<K, V>`
  - [ ] Preserve runtime behaviour (check with existing vitest NotificationFeed tests if any, else manual smoke)
  - [ ] Cap the file at <= 14 remaining errors

- [ ] Task 3: Retype NewSessionModal, App.svelte, StatusBadge (AC: 2, 3)
  - [ ] Annotate each `{}`-typed lookup with `Record<K, V>` or `Partial<Record<K, V>>`
  - [ ] Ensure no hex colour literals are introduced

- [ ] Task 4: Sweep neighbour files flagged by `sveltecheck-by-file` (AC: 2)
  - [ ] Identify any other files above threshold after the primary four land
  - [ ] Apply the same Record pattern
  - [ ] Confirm remaining P2/P3 errors match plan target

- [ ] Task 5: Verify reduction (AC: 5)
  - [ ] `just sveltecheck-count` <= 930
  - [ ] `vitest run` — 696 tests pass
  - [ ] `vite build` clean
  - [ ] `wails dev` smoke — NotificationFeed renders live notifications end-to-end

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on new helpers in `types/status.ts` (narrowing function)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `just sveltecheck-count` <= 930
- [ ] `just sveltecheck-ratchet` passes
- [ ] `vitest run` — 696 tests pass
- [ ] `vite build` clean
- [ ] NotificationFeed runtime smoke (manual OR playwright) confirms no regression
- [ ] `/simplify` run on every modified file
- [ ] Code review: no `any`, no `@ts-ignore`, no `@ts-nocheck`, no hex colour literals
- [ ] PR size <= 400 lines (plan for 2 PRs — R4)
- [ ] Commit body records: `sveltecheck-count: <prev> → <cur> (Δ -<n>)`
