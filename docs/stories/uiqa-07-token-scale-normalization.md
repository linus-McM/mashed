# Story uiqa-07: Token Scale Normalization

**Status:** ready
**Size:** M
**Priority:** P3
**Domain:** frontend
**Depends on:** none

## Description

Sweep orphaned font-size and spacing values across the frontend and snap them to the nearest design token. The review identified four font-size outliers (`10px`, `12px`, `15px`, `18px`) and three spacing outliers (`3px`, `5px`, `6px`) used in place of the canonical scale (`--text-label: 11px`, `--text-body: 13px`, `--sp-2xs: 2px`, `--sp-xs: 4px`, `--sp-sm: 8px`, `--sp-md: 12px`, etc.). Also adds `font-variant-numeric: tabular-nums` to every token-count and elapsed-time display so numbers stop jittering as they update. This is the typography polish that moves density-heavy views from "functional" to "Bloomberg-grade."

## Developer Notes

### Normalization rules (FIX-13, FIX-14, FIX-15)

**Font-size replacements:**

| Found | Replace with | Notes |
|-------|--------------|-------|
| `10px` | `var(--text-label)` (11px) | 1px shift upward; verify nothing breaks layout |
| `12px` | `var(--text-body)` (13px) or `var(--text-label)` (11px) | Pick `--text-label` for uppercase labels, `--text-body` for sentence text; judgment call per site |
| `15px` | `var(--text-data)` (14px) or `var(--text-section)` (16px) | Prefer `--text-data` unless section heading |
| `18px` | `var(--text-section)` (16px) or `var(--text-page)` (20px) | Prefer `--text-section` unless a clear page title |

**Spacing replacements:**

| Found | Replace with |
|-------|--------------|
| `3px` | `var(--sp-2xs)` (2px) — rounds down |
| `5px` | `var(--sp-xs)` (4px) — rounds down |
| `6px` | `var(--sp-xs)` (4px) or `var(--sp-sm)` (8px) — prefer `--sp-sm` if visual feel matters |

### Known hit list (from review)

The review explicitly calls out these sites; treat as the starting inventory but do not limit the story to them:

- `NotificationFeed.svelte:1399` `.sub-count` — `font-size: 10px` + `padding: 0 5px`
- `NotificationFeed.svelte:1598` `.new-session-btn` — `font-size: 12px`
- `AgentDetail.svelte:695` `.header-repo` — `font-size: 14px` (review AD-05; `14px` IS on the scale via `--text-data`, so this is a "use token instead of literal" cleanup, not a size change)
- `SpawnAgent.svelte:137` `h2` — `font-size: 18px` → `var(--text-section)`
- `SpawnAgent.svelte:152` `.field` — `margin-bottom: 20px` → `var(--sp-lg)` (16px) or `var(--sp-xl)` (24px); pick `--sp-lg` to reduce whitespace

### Tabular-nums for numeric displays (FIX-12)

Add `font-variant-numeric: tabular-nums;` to the following (find by reading each view):

- **NotificationFeed.svelte**: `.token-count`, `.elapsed`, `.sub-count`
- **AgentDetail.svelte**: `.token-progress` label, `.elapsed`, any header metric
- **StatusBadge.svelte** (if it shows numbers)
- **RepoContextBar.svelte** (if it displays count totals)

These are the displays that flicker when values increment. Tabular nums freeze the digit widths.

### Workflow

1. Full-tree grep for each outlier literal: `grep -rn "font-size: 10px" frontend/src/`. Repeat for 12px, 15px, 18px.
2. Same for `padding.*: *3px`, `padding.*: *5px`, `padding.*: *6px`, `margin.*: *3px` etc.
3. For each hit, read the surrounding rule, decide the target token, apply via Edit.
4. Grep for token/time display classes (`.token-count`, `.elapsed`, `.sub-count`, `.token-progress`) and add `font-variant-numeric: tabular-nums;`.
5. Build and run a "no orphaned values" linter assertion at the end of the story.

### Risks & edge cases

- **Layout reflow**: 10px → 11px or 18px → 16px can nudge baselines. Visual spot-check the affected views after each change.
- **Compound rules**: `padding: 3px 5px 3px 8px` becomes `padding: var(--sp-2xs) var(--sp-xs) var(--sp-2xs) var(--sp-sm)`. Preserve order.
- **Not every number is a token**: line-height `1.5`, opacity `0.5`, border-radius `50%` — leave alone. Only font-size, padding, margin, gap, top/left positioning literals matter.
- **False positives in grep**: `transform: translateY(-5px)` is NOT in scope — don't touch positioning offsets. Only box-model sizing.
- **`borderPalette` and user data**: do not touch.

### Out of scope

- Any colour changes (other stories).
- Typography weights, line-heights, letter-spacing.
- Animations, transitions.
- Layout grid changes.
- Adding new tokens to `style.css` — reuse what exists.

### Reference files

- `frontend/src/style.css` lines 26–39 — canonical token values
- `DESIGN.md` — token rationale (read-only)

Reference skills: `/simplify`, `/playwright-cli`.

## Acceptance Criteria

**AC-1: Font-size outliers eliminated from in-scope `.svelte` files**
- Given the modified frontend
- When `grep -rnE "font-size:\\s*(10|12|15|18)px" frontend/src/` is run
- Then zero matches exist in `.svelte` files inside `frontend/src/views/` and `frontend/src/components/`
- Exception: styled test fixtures or comments are allowed if explicitly marked

**AC-2: Spacing outliers eliminated**
- Given the modified frontend
- When `grep -rnE "(padding|margin|gap):\\s*.*(\\b3px\\b|\\b5px\\b|\\b6px\\b)" frontend/src/` is run against `.svelte` files in views/ and components/
- Then zero matches exist
- Exception: shorthand composites that include a token-compliant value alongside are still flagged and must be tokenized

**AC-3: `var(--text-data)` used for 14px references**
- Given `AgentDetail.svelte:695`
- When parsed
- Then `.header-repo` (or equivalent) uses `font-size: var(--text-data)` and not `font-size: 14px`

**AC-4: Numeric displays use tabular-nums**
- Given `.token-count`, `.elapsed`, `.sub-count`, `.token-progress` rules across NotificationFeed, AgentDetail, and any shared numeric display components
- When each rule is parsed
- Then it declares `font-variant-numeric: tabular-nums;`

**AC-5: No visual regression in known hot spots**
- Given before/after Playwright screenshots of NotificationFeed, AgentDetail, SpawnAgent
- When compared pixel-wise
- Then no element has shifted by more than 4 pixels in either axis
- And no text overflows its container
- And no element is cut off

**AC-6: Linter assertion blocks regressions**
- Given a small Node or Svelte compile-time script that greps modified files for orphaned values
- When run against `frontend/src/**/*.svelte`
- Then it exits 0
- And if a new orphaned value is introduced later, it exits nonzero

## BDD Test Scenarios

```gherkin
Feature: uiqa-07 token scale normalization

  Scenario: Zero 10px font-sizes in views and components
    Given modified frontend/src
    When grepped for "font-size: 10px"
    Then no matches exist in views/ or components/

  Scenario: Zero 12px font-sizes in views and components
    When grepped for "font-size: 12px"
    Then no matches exist

  Scenario: Zero 18px font-sizes
    When grepped for "font-size: 18px"
    Then no matches exist

  Scenario: AgentDetail header uses text-data token
    Given AgentDetail.svelte
    When the .header-repo rule is parsed
    Then it contains "font-size: var(--text-data)"
    And it does not contain "font-size: 14px"

  Scenario: Sub-count uses tabular-nums
    Given NotificationFeed.svelte
    When the .sub-count rule is parsed
    Then it contains "font-variant-numeric: tabular-nums"

  Scenario: Token count uses tabular-nums
    Given NotificationFeed.svelte
    When the .token-count rule is parsed
    Then it contains "font-variant-numeric: tabular-nums"

  Scenario: Linter passes on modified files
    Given the token-scale linter script
    When run on modified .svelte files
    Then it exits 0

  Scenario: No layout shift above 4px
    Given a Playwright harness mounting NotificationFeed pre- and post-change
    When element positions are compared
    Then no element has moved more than 4 pixels
```

## Tasks / Subtasks

- [ ] Task 1: Inventory orphaned font-sizes (AC-1, AC-3)
  - [ ] Subtask 1a: Grep for 10px, 12px, 15px, 18px font-sizes in views/ and components/
  - [ ] Subtask 1b: For each hit, read surrounding rule and pick target token
  - [ ] Subtask 1c: Apply edits

- [ ] Task 2: Inventory orphaned spacing (AC-2)
  - [ ] Subtask 2a: Grep for 3px, 5px, 6px in padding/margin/gap
  - [ ] Subtask 2b: Apply edits mapping to nearest --sp-* token

- [ ] Task 3: Add tabular-nums (AC-4)
  - [ ] Subtask 3a: Identify every numeric display class
  - [ ] Subtask 3b: Add `font-variant-numeric: tabular-nums;` to each rule

- [ ] Task 4: Visual regression checks (AC-5)
  - [ ] Subtask 4a: Playwright before/after screenshots
  - [ ] Subtask 4b: Pixel shift assertion

- [ ] Task 5: Linter script (AC-6)
  - [ ] Subtask 5a: Write a small Node script at `frontend/scripts/check-orphan-tokens.mjs`
  - [ ] Subtask 5b: Wire it into `package.json` or lefthook pre-commit if appropriate
  - [ ] Subtask 5c: Run against the full frontend; expect zero findings post-story

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
- [ ] Linter script passes (AC-6)
- [ ] Story status updated to `done`
