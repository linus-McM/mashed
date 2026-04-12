# Story uiqa-02: Eradicate `#39ff14` Neon Green

**Status:** done
**Size:** M
**Priority:** P1
**Domain:** frontend
**Depends on:** none

## Description

Replace every hardcoded instance of `#39ff14` (neon green) and `#00c4b3` (hardcoded teal) in `.svelte` files with their correct design system tokens (`var(--accent-green)` and `var(--accent-teal)`). This is the single biggest fix in the review — the neon green appears in 5 files and ~15 CSS rules plus 2 JS `statusColors` maps, and the wrong color breaks every non-default theme. This story is a mechanical find/replace grouped across five files to keep the theme-switching fix atomic (demoable by toggling themes and watching every running agent indicator re-color correctly).

## Developer Notes

### Files and exact line numbers (from review Section 4)

1. **`frontend/src/components/StatusBadge.svelte`**
   - Line 6 (JS `statusColors` map): `running: '#39ff14'` → `running: 'var(--accent-green)'`
   - Line 7 (JS `statusColors` map): `open: '#00c4b3'` → `open: 'var(--accent-teal)'`
   - This is the *highest visibility* fix — every running agent badge in the app uses this map.

2. **`frontend/src/views/NotificationFeed.svelte`** (1894 lines)
   - Line 1100 (JS `statusColors`): `running: '#39ff14'` → `running: 'var(--accent-green)'`
   - Line 1101 (JS `statusColors`): `open: '#00c4b3'` → `open: 'var(--accent-teal)'`
   - Lines 1700, 1706, 1713, 1719 (CSS `.action-hot` color): `#39ff14` → `var(--accent-green)`
   - Lines 1707, 1720 (CSS `.action-hot` border-color): `#39ff14` → `var(--accent-green)`
   - Line 1271 (CSS `.color-swatch.active`): `border-color: #fff` → `border-color: var(--text-primary)`

3. **`frontend/src/views/AgentDetail.svelte`** (~1320 lines)
   - Line 679 (`.back-btn` color): `#39ff14` → `var(--accent-green)`
   - Line 689 (`.back-btn` text-shadow): `rgba(57, 255, 20, ...)` — replace with `color-mix(in srgb, var(--accent-green) N%, transparent)` (keep the same alpha percentage; see Risks)
   - Line 1005 (`.git-hot` color): `#39ff14` → `var(--accent-green)`
   - Lines 1011, 1012 (`.git-hot` border/shadow): `#39ff14` → `var(--accent-green)`

4. **`frontend/src/components/RepoContextBar.svelte`**
   - Line 61: `#39ff14` → `var(--accent-green)` (single instance)

### Critical clarification: JS `statusColors` must use CSS variable strings

StatusBadge already uses `color-mix(in srgb, var(--status-color) 15%, transparent)` (review Section 2.6). That means the JS value IS pushed into a CSS custom property via `style=` binding. Storing the literal string `'var(--accent-green)'` works because it's resolved at paint time. Verify by:
1. Reading StatusBadge.svelte around lines 1–30 to confirm the CSS var is bound via `style=`.
2. If StatusBadge uses the raw hex in a non-CSS context (e.g., canvas draw), swap the whole value to `'var(--accent-green)'` anyway but add a comment and a fallback hex in JS. The review's recommendation is `'var(--accent-green)'`.

### Risks & edge cases

- **rgba derivation (AgentDetail:689)**: the existing `rgba(57, 255, 20, 0.X)` uses the *wrong* RGB — it's not even the design system's `#00e57a`. Don't port the wrong hex into a comment. Use `color-mix(in srgb, var(--accent-green) N%, transparent)` where `N` is `alpha * 100` rounded. For instance `rgba(57, 255, 20, 0.35)` becomes `color-mix(in srgb, var(--accent-green) 35%, transparent)`.
- **uiqa-05 overlap**: uiqa-05 does the broader rgba → color-mix sweep. This story only touches the `#39ff14`-adjacent rgbas at lines 685/689/1011/1012 because they're in the same rules as the hex colors. Leave other unrelated rgba instances for uiqa-05.
- **`borderPalette` (NotificationFeed.svelte:34-37)** is NOT part of this story — those 11 hex values are user-selectable data, not theme. The review explicitly marks them as acceptable. Do not touch.
- **Acceptable exceptions**: TitleBar.svelte traffic lights (lines 131-133) and Settings.svelte traffic lights (lines 165-167) are macOS chrome and must stay hex. Do not touch.

### Out of scope

- ExecutionBar cyan/orange (uiqa-03).
- Extracting a shared `.glow-btn` class (uiqa-04).
- The broader rgba → color-mix sweep (uiqa-05).
- Any entry animations (uiqa-06).

### Reference files

- `frontend/src/style.css` — verify `--accent-green: #00e57a` and `--accent-teal: #00c4b3` already exist (lines 10, 16 — they do).
- `frontend/src/components/StatusBadge.svelte` — existing `color-mix` usage is the pattern to validate.

Reference skills: `/simplify`, `/playwright-cli` for the computed-style verification AC.

## Acceptance Criteria

**AC-1: Zero `#39ff14` instances remain in `frontend/src/`**
- Given the modified codebase
- When `grep -r "#39ff14" frontend/src/` is run
- Then the exit code is 1 (no matches)
- And no `.svelte`, `.ts`, or `.js` file under `frontend/src/` contains `#39ff14` or `39ff14` (case-insensitive)

**AC-2: JS `statusColors` maps use CSS variable strings**
- Given `StatusBadge.svelte:6-7` and `NotificationFeed.svelte:1100-1101`
- When the JS is parsed
- Then `statusColors.running === 'var(--accent-green)'`
- And `statusColors.open === 'var(--accent-teal)'`

**AC-3: Running badge renders the design system green at runtime**
- Given a mounted `<StatusBadge status="running" />`
- When `getComputedStyle(badge).getPropertyValue('--status-color')` (or the binding target) is read
- Then the resolved color equals the value of `var(--accent-green)` from `:root` (i.e. `#00e57a` in the default theme)
- And NOT `#39ff14`

**AC-4: Theme switch propagates to every affected element**
- Given the default theme shows the running badge in `#00e57a`
- When a different theme is loaded that overrides `--accent-green: #ff00ff` (test-only theme)
- Then StatusBadge, NotificationFeed `.action-hot`, AgentDetail `.back-btn`/`.git-hot`, and RepoContextBar all render with the new magenta color without a page reload (after Svelte re-render)

**AC-5: `.color-swatch.active` border uses `--text-primary`**
- Given NotificationFeed line 1271
- When the CSS is parsed
- Then `border-color: var(--text-primary);` is present and `border-color: #fff` is absent

**AC-6: Acceptable exceptions preserved**
- Given `TitleBar.svelte:131-133` and `Settings.svelte:165-167`
- When grepped
- Then the three macOS traffic light hex values (`#ff5f57`, `#febc2e`, `#28c840`) are still present and unchanged

## BDD Test Scenarios

```gherkin
Feature: uiqa-02 eradicate neon green

  Scenario: No #39ff14 anywhere in frontend/src
    Given the modified codebase
    When the tree is grepped for "39ff14" case-insensitive
    Then zero matches are found in any file under frontend/src/

  Scenario: StatusBadge running resolves to --accent-green
    Given a Svelte component test mounting <StatusBadge status="running" />
    When it reads the computed --status-color CSS custom property
    Then the resolved value equals the value of --accent-green from :root

  Scenario: StatusBadge open resolves to --accent-teal
    Given <StatusBadge status="open" />
    Then the computed --status-color equals the value of --accent-teal

  Scenario: Theme override propagates to NotificationFeed hot action
    Given a test page that sets --accent-green to #ff00ff on :root
    And NotificationFeed is mounted with a "running" repo group
    When the .action-hot button is inspected
    Then its computed color is rgb(255, 0, 255)

  Scenario: AgentDetail .back-btn glow uses color-mix not raw rgba
    Given AgentDetail.svelte source
    When the CSS for .back-btn is parsed
    Then it contains "color-mix" and "var(--accent-green)"
    And it does NOT contain "rgba(57, 255, 20"
    And it does NOT contain "#39ff14"

  Scenario: Color swatch active border uses text-primary
    Given NotificationFeed.svelte source
    When searched for ".color-swatch.active"
    Then the following rule contains "border-color: var(--text-primary)"
    And does not contain "border-color: #fff"

  Scenario: macOS traffic lights preserved
    Given TitleBar.svelte and Settings.svelte
    When grepped for "#ff5f57", "#febc2e", "#28c840"
    Then each of the three appears at least once in each file
```

## Tasks / Subtasks

- [x] Task 1: Replace `#39ff14` and `#00c4b3` in StatusBadge.svelte (AC-1, AC-2, AC-3)
  - [x] Subtask 1a: Read lines 1–30 to confirm how `statusColors` values flow into CSS
  - [x] Subtask 1b: Change lines 6–7 to use `'var(--accent-green)'` and `'var(--accent-teal)'`

- [x] Task 2: Replace `#39ff14`, `#00c4b3`, and `#fff` in NotificationFeed.svelte (AC-1, AC-2, AC-5)
  - [x] Subtask 2a: Update JS statusColors at lines 1100–1101
  - [x] Subtask 2b: Update `.action-hot` color at lines 1700, 1706, 1713, 1719
  - [x] Subtask 2c: Update `.action-hot` border-color at lines 1707, 1720
  - [x] Subtask 2d: Update `.color-swatch.active` border at line 1271

- [x] Task 3: Replace `#39ff14` and adjacent rgbas in AgentDetail.svelte (AC-1, AC-4)
  - [x] Subtask 3a: `.back-btn` color at line 679
  - [x] Subtask 3b: `.back-btn` text-shadow at line 689 using `color-mix`
  - [x] Subtask 3c: `.git-hot` color at line 1005
  - [x] Subtask 3d: `.git-hot` border/shadow at lines 1011, 1012

- [x] Task 4: Replace `#39ff14` in RepoContextBar.svelte line 61 (AC-1) — actual path: `components/bmad/RepoContextBar.svelte`

- [x] Task 5: Write Svelte component tests (AC-3, AC-4, AC-5)
  - [x] Subtask 5a: Test asserting StatusBadge resolves via JS map chain (jsdom pivot — see notes)
  - [x] Subtask 5b: Test with an overridden `--accent-green` on `:root` confirming propagation to NotificationFeed `.action-hot` and AgentDetail `.back-btn`
  - [x] Subtask 5c: fs-based assertion that `#39ff14` is absent from any .svelte source under views/ or components/

- [x] Task 6: Verify acceptable exceptions still present (AC-6)
  - [x] Subtask 6a: Grep that TitleBar.svelte still contains `#ff5f57`, `#febc2e`, `#28c840`
  - [x] Subtask 6b: Same check for Settings.svelte

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ coverage on modified files (CSS test coverage via component tests)
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] Frontend build passes
- [x] `/simplify` run on all modified files
- [x] No new `#39ff14` introduced anywhere
- [x] No orphaned token-scale values introduced in modified files
- [x] Story status updated to `done`
