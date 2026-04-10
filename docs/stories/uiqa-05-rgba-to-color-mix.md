# Story uiqa-05: rgba() → color-mix() Migration

**Status:** ready
**Size:** L
**Priority:** P2
**Domain:** frontend
**Depends on:** uiqa-01, uiqa-02

## Description

Migrate the remaining ~40 hardcoded `rgba(R, G, B, alpha)` instances across the frontend to `color-mix(in srgb, var(--accent-X) N%, transparent)` so translucent accent colors follow the theme. This story is scoped to the accent families: neon green `rgba(57, 255, 20, ...)`, design green `rgba(0, 229, 122, ...)`, red `rgba(232, 69, 69, ...)`, and amber `rgba(240, 165, 0, ...)`. Black overlays (`rgba(0, 0, 0, ...)`) are left alone EXCEPT where they match the standardized modal backdrop, in which case they become `var(--overlay-backdrop)` from uiqa-01. After this story, `color-mix` is the app-wide idiom for translucency — matching the StatusBadge pattern the review calls out as correct.

## Developer Notes

### Scope — by review audit table (Section 4)

| Pattern | Approximate count | Replacement |
|---------|-------------------|-------------|
| `rgba(57, 255, 20, ...)` neon green | ~14 | `color-mix(in srgb, var(--accent-green) N%, transparent)` |
| `rgba(0, 229, 122, ...)` design green | ~12 | `color-mix(in srgb, var(--accent-green) N%, transparent)` |
| `rgba(232, 69, 69, ...)` red | ~10 | `color-mix(in srgb, var(--accent-red) N%, transparent)` |
| `rgba(240, 165, 0, ...)` amber | ~8 | `color-mix(in srgb, var(--accent-amber) N%, transparent)` |
| `rgba(0, 0, 0, 0.6)` modal backdrop | ~3–5 | `var(--overlay-backdrop)` |
| Other `rgba(0, 0, 0, ...)` overlays | ~10 | **leave unchanged** (theme-independent, review explicitly allows) |

### Alpha → percentage rule

`rgba(r, g, b, 0.35)` → `color-mix(in srgb, var(--accent-green) 35%, transparent)`.
- `0.05` → `5%`
- `0.08` → `8%`
- `0.1` → `10%`
- `0.15` → `15%`
- `0.2` → `20%`
- `0.35` → `35%`
- `0.6` → `60%`

### Workflow

1. `grep -rn "rgba(57, 255, 20" frontend/src/` — list every neon-green rgba. For each, read the surrounding rule to confirm it's on an element that should theme (not user data). Replace with `color-mix(in srgb, var(--accent-green) N%, transparent)`.
2. Repeat for `rgba(0, 229, 122`, `rgba(232, 69, 69`, `rgba(240, 165, 0`.
3. For `rgba(0, 0, 0, 0.6)`, only replace instances inside modal backdrops (`.modal-backdrop`, `.overlay`, `.scrim` classes) — use `var(--overlay-backdrop)`. Leave shadow-stack black rgbas (box-shadows, dropshadows) alone.

### Risks & edge cases

- **Already-migrated patterns**: StatusBadge.svelte already uses the `color-mix` idiom — do NOT re-edit. Same for any rule uiqa-02, uiqa-03, or uiqa-04 already fixed.
- **Compound shadows**: `box-shadow: 0 2px 4px rgba(0,0,0,0.3), 0 0 12px rgba(57,255,20,0.4);` — only the neon-green half changes. Preserve the black rgba.
- **Linear gradients**: `linear-gradient(180deg, rgba(0,229,122,0.1), transparent)` also qualifies — `color-mix` works inside gradients.
- **Percentage rounding**: `rgba(X,Y,Z, 0.125)` becomes `12.5%` — keep the fractional value, don't round to 13%.
- **Non-accent rgbas**: any `rgba()` that doesn't match one of the listed RGB triples stays unchanged (it's likely already off-palette; if a reviewer flags it, that's a follow-up story).

### Estimate: file count

From the review: "Over 80 instances of `rgba(R, G, B, alpha)`" — this story targets ~40 of them (the colored accents). Expect 5–10 files touched. Read each file's rgba rules before blanket-replacing.

### Out of scope

- Adding new tokens for accent families (all four accents already exist in `style.css`).
- Rewriting `box-shadow` stacks entirely — only the rgba inside them changes.
- Touching `borderPalette` in NotificationFeed (review Section 4 "Acceptable Exceptions" — user data).
- Any CSS rule the previous uiqa stories already touched.

### Reference files

- `frontend/src/components/StatusBadge.svelte` — reference for the correct `color-mix(in srgb, var(--status-color) 15%, transparent)` idiom.
- `frontend/src/style.css` — no edits, but confirm `--overlay-backdrop` (from uiqa-01) is present.

Reference skills: `/simplify`, `/playwright-cli`.

## Acceptance Criteria

**AC-1: All neon-green rgbas replaced**
- Given the modified codebase
- When `grep -rn "rgba(57" frontend/src/` runs
- Then the exit code is 1 (no matches remain)
- And a grep for `color-mix.*var(--accent-green)` returns more matches than before the story

**AC-2: All design-green rgbas replaced (except those inside StatusBadge and already-migrated files)**
- Given the modified codebase
- When `grep -rn "rgba(0, 229, 122" frontend/src/` runs
- Then the exit code is 1 (no matches remain)

**AC-3: All red rgbas (`rgba(232, 69, 69, ...)`) replaced**
- Given the modified codebase
- When grepped
- Then zero instances of `rgba(232, 69, 69` remain in `.svelte` files

**AC-4: All amber rgbas (`rgba(240, 165, 0, ...)`) replaced**
- Given the modified codebase
- When grepped
- Then zero instances of `rgba(240, 165, 0` remain in `.svelte` files

**AC-5: Modal backdrop uses `--overlay-backdrop`**
- Given the modal files in `frontend/src/views/` (NewSessionModal, SpawnAgent, BranchModal, and any others)
- When parsed
- Then each `.modal-backdrop` / `.overlay` / scrim rule uses `background: var(--overlay-backdrop);`
- And no modal backdrop uses literal `rgba(0, 0, 0, 0.6)`

**AC-6: Black box-shadow rgbas preserved**
- Given any rule with `box-shadow: 0 Npx Mpx rgba(0, 0, 0, 0.N)` not inside a modal backdrop
- When parsed
- Then these rgbas are unchanged (not replaced with color-mix)

**AC-7: Theme override propagates across migrated rules**
- Given a test overriding `--accent-red: #ff00ff` on `:root`
- When an element using `color-mix(in srgb, var(--accent-red) 15%, transparent)` is mounted
- Then its computed background contains magenta-derived values (not the old red)

## BDD Test Scenarios

```gherkin
Feature: uiqa-05 rgba to color-mix migration

  Scenario: Zero neon green rgbas remain
    Given the modified frontend/src
    When grepped for "rgba(57,\\s*255,\\s*20"
    Then no matches exist

  Scenario: Zero design green rgbas remain
    When grepped for "rgba(0,\\s*229,\\s*122"
    Then no matches exist outside of comments

  Scenario: Zero red rgbas remain
    When grepped for "rgba(232,\\s*69,\\s*69"
    Then no matches exist

  Scenario: Zero amber rgbas remain
    When grepped for "rgba(240,\\s*165,\\s*0"
    Then no matches exist

  Scenario: Modal backdrop uses token
    Given NewSessionModal.svelte
    When the backdrop rule is parsed
    Then its background is "var(--overlay-backdrop)"

  Scenario: Box-shadow black rgbas preserved
    Given a file that had "box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3)" before the story
    When the diff is inspected
    Then that specific line is unchanged

  Scenario: Theme override changes red translucent background
    Given a component using "color-mix(in srgb, var(--accent-red) 15%, transparent)"
    And a test setting --accent-red to #ff00ff on :root
    When the computed background is read
    Then it reflects a magenta-derived rgba, not the original red
```

## Tasks / Subtasks

- [ ] Task 1: Inventory rgba instances (AC-1, AC-2, AC-3, AC-4)
  - [ ] Subtask 1a: Grep for each of the four accent RGB triples and list (file, line) pairs
  - [ ] Subtask 1b: Note which lines were already touched by uiqa-02/03/04 and skip them
  - [ ] Subtask 1c: Note alpha percentages per instance

- [ ] Task 2: Replace neon green rgbas (AC-1)
  - [ ] Subtask 2a: For each (file, line) replace `rgba(57, 255, 20, N)` with `color-mix(in srgb, var(--accent-green) N*100%, transparent)`

- [ ] Task 3: Replace design green rgbas (AC-2)
  - [ ] Subtask 3a: For each (file, line) replace `rgba(0, 229, 122, N)` with `color-mix(in srgb, var(--accent-green) N*100%, transparent)`

- [ ] Task 4: Replace red and amber rgbas (AC-3, AC-4)
  - [ ] Subtask 4a: Replace `rgba(232, 69, 69, N)` with `color-mix(in srgb, var(--accent-red) N*100%, transparent)`
  - [ ] Subtask 4b: Replace `rgba(240, 165, 0, N)` with `color-mix(in srgb, var(--accent-amber) N*100%, transparent)`

- [ ] Task 5: Update modal backdrops (AC-5, AC-6)
  - [ ] Subtask 5a: Identify every modal backdrop / scrim rule
  - [ ] Subtask 5b: Replace with `var(--overlay-backdrop)`
  - [ ] Subtask 5c: Confirm non-backdrop black rgbas remain unchanged

- [ ] Task 6: Component theme tests (AC-7)
  - [ ] Subtask 6a: One test per accent family verifying theme override propagates through color-mix

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
- [ ] Story status updated to `done`
