# Story uiqa-03: Remaining Hardcoded Colors

**Status:** ready
**Size:** S
**Priority:** P1
**Domain:** frontend
**Depends on:** uiqa-01

## Description

Clean up the last hardcoded colors in the review's "High: Colors outside design system" table that aren't neon green: the macOS tab-close red in AgentDetail, the two orphaned indicator colors in Settings, the green text-shadow in Settings, and — most importantly — the two rogue colors `#22d3ee` and `#fb923c` in ExecutionBar that now have proper tokens (`--accent-cyan`, `--accent-orange`) thanks to uiqa-01. After this story lands, the only remaining hardcoded hex values in `frontend/src/` are the acceptable macOS chrome exceptions and the `borderPalette` user data array.

## Developer Notes

### Files and exact line numbers (from review Section 4)

1. **`frontend/src/views/AgentDetail.svelte`**
   - Line 821 (`.tab-close:hover` color): `#ff5f57` → `var(--accent-red)` (review AD-03). Note: this is NOT the macOS traffic light — it's a tab close button in the editor pane, so it must theme correctly.

2. **`frontend/src/views/Settings.svelte`**
   - Line 466 (`text-shadow`): `0 0 10px rgba(0, 229, 122, 0.6)` → `var(--glow-spread) color-mix(in srgb, var(--accent-green) 60%, transparent)` (review ST-04). Uses `--glow-spread` from uiqa-01.
   - Line 860 (`.import-indicator.dark` background): `#565670` → `var(--text-muted)` (review table: "Add to token set or use `var(--text-muted)`"). Pick `--text-muted` (`#2e3d4d`) since it's already in `style.css:19` and semantically matches "dark indicator."
   - Line 864 (`.import-indicator.light` background): `#c0c0d0` → `var(--text-dim)` (review table: "use `var(--text-dim)`"). Matches `style.css:18` (`#4a5a6a`) — `--text-dim` is the lighter of the two.

3. **`frontend/src/components/bmad/ExecutionBar.svelte`**
   - Line 163 (`.ctrl-btn.run` color): `#22d3ee` → `var(--accent-cyan)` (review WB-01)
   - Line 168 (`.ctrl-btn.run` shadow or border — verify at read time): `#22d3ee` or `rgba(34, 211, 238, ...)` → token-derived equivalent
   - Line 197 (`.ctrl-btn.stop` color): `#fb923c` → `var(--accent-orange)` (review WB-02)
   - Line 202 (`.ctrl-btn.stop` shadow or border): `#fb923c` or rgba → token-derived equivalent
   - Also check lines 160–210 for any secondary references to these hexes inside the same button rules.

4. **`frontend/src/components/bmad/ProcessNode.svelte`**
   - Line 126 (`.process-node` box-shadow): `rgba(0, 229, 122, 0.35)` → `color-mix(in srgb, var(--accent-green) 35%, transparent)` (review WB-03). Minor but keeps this story atomic for "all ProcessNode-adjacent rgba green".

5. **`frontend/src/views/WorkflowBuilder.svelte`**
   - Line 883: `rgba(248, 81, 73, 0.1)` and `rgba(248, 81, 73, 0.15)` → `color-mix(in srgb, var(--accent-red) 10%, transparent)` and `15%` respectively (review WB-05). These use a non-system red (`#f85149` vs design system `#e84545`); see Risks.

### Risks & edge cases

- **Red shade drift (WorkflowBuilder:883)**: `rgba(248, 81, 73, ...)` corresponds to `#f85149` (GitHub-style), not the design system's `--accent-red: #e84545`. Swapping will slightly shift the hue. This is the correct behaviour per the review — we want theme consistency, not visual pixel parity.
- **ExecutionBar verification**: the review cites exact lines but reading 160–210 first is mandatory to handle any rgba variants the review did not call out explicitly.
- **Dark/light indicator semantics**: the names `--text-muted` and `--text-dim` may not align with "dark/light indicator" visually. If a reviewer prefers adding new tokens `--indicator-dark` / `--indicator-light`, escalate — but the review's "Fix" column explicitly recommends the existing text tokens, so follow that.

### Out of scope

- Anything else in uiqa-02 (neon green). Assume uiqa-02 is complete.
- The broader rgba sweep across other files (uiqa-05).
- Shared glow-btn extraction (uiqa-04).

### Reference files

- `frontend/src/style.css` lines 10–19 — verify token values before mapping.
- `frontend/src/components/bmad/ExecutionBar.svelte` lines 150–210 — read before editing.

Reference skills: `/simplify`, `/playwright-cli`.

## Acceptance Criteria

**AC-1: ExecutionBar uses `--accent-cyan` and `--accent-orange`**
- Given `ExecutionBar.svelte`
- When the file is parsed
- Then `.ctrl-btn.run` rules reference `var(--accent-cyan)` and not `#22d3ee`
- And `.ctrl-btn.stop` rules reference `var(--accent-orange)` and not `#fb923c`
- And there are zero occurrences of `#22d3ee` or `#fb923c` in the file

**AC-2: AgentDetail tab-close hover uses `--accent-red`**
- Given `AgentDetail.svelte:821`
- When parsed
- Then `.tab-close:hover` declares `color: var(--accent-red);`
- And the file still contains the acceptable exception hex `#ff5f57` nowhere (this is not the macOS traffic light — it must be removed from this file)

**AC-3: Settings indicators and text-shadow use tokens**
- Given `Settings.svelte`
- When parsed
- Then `.import-indicator.dark` background references `var(--text-muted)`
- And `.import-indicator.light` background references `var(--text-dim)`
- And the line-466 text-shadow references `var(--glow-spread)` and `color-mix(...var(--accent-green)...)`
- And `Settings.svelte` contains zero occurrences of `#565670`, `#c0c0d0`, or `rgba(0, 229, 122`

**AC-4: ProcessNode and WorkflowBuilder use color-mix with tokens**
- Given `ProcessNode.svelte:126` and `WorkflowBuilder.svelte:883`
- When parsed
- Then each rgba-with-RGB reference flagged in Developer Notes is replaced with `color-mix(in srgb, var(--accent-...) N%, transparent)`
- And neither file contains `rgba(0, 229, 122` or `rgba(248, 81, 73`

**AC-5: Theme override propagates**
- Given a test that overrides `--accent-cyan: #00ffff` and `--accent-orange: #ffa500` on `:root`
- When ExecutionBar is mounted
- Then `.ctrl-btn.run` renders with `rgb(0, 255, 255)` and `.ctrl-btn.stop` renders with `rgb(255, 165, 0)`

**AC-6: Acceptable macOS chrome exceptions still present**
- Given `TitleBar.svelte:131-133` and `Settings.svelte:165-167`
- When grepped
- Then the three traffic light hex values remain

## BDD Test Scenarios

```gherkin
Feature: uiqa-03 remaining hardcoded colors

  Scenario: ExecutionBar run button uses accent-cyan
    Given ExecutionBar.svelte source
    When the .ctrl-btn.run block is parsed
    Then every color, border-color, and box-shadow uses var(--accent-cyan) or a color-mix referencing it
    And the literal "#22d3ee" does not appear in the file

  Scenario: ExecutionBar stop button uses accent-orange
    Given ExecutionBar.svelte source
    When the .ctrl-btn.stop block is parsed
    Then every color reference uses var(--accent-orange) or a color-mix referencing it
    And the literal "#fb923c" does not appear in the file

  Scenario: AgentDetail tab-close hover is themed
    Given AgentDetail.svelte source
    When searched for ".tab-close:hover"
    Then the rule contains "color: var(--accent-red)"
    And the literal "#ff5f57" does not appear in this file

  Scenario: Settings indicators use text tokens
    Given Settings.svelte source
    When searched for ".import-indicator.dark" and ".import-indicator.light"
    Then each rule references var(--text-muted) or var(--text-dim) respectively
    And the literal "#565670" and "#c0c0d0" do not appear

  Scenario: Settings text-shadow uses glow-spread
    Given Settings.svelte line 466 rule
    When parsed
    Then it contains "var(--glow-spread)"
    And it contains "color-mix" with "var(--accent-green)"
    And it does not contain "rgba(0, 229, 122"

  Scenario: Theme override changes ExecutionBar colors
    Given a test harness overriding --accent-cyan and --accent-orange
    When ExecutionBar is mounted
    Then .ctrl-btn.run and .ctrl-btn.stop computed colors reflect the override

  Scenario: TitleBar traffic lights preserved
    Given TitleBar.svelte
    When grepped for the three traffic light hex values
    Then each is still present
```

## Tasks / Subtasks

- [ ] Task 1: Replace cyan and orange in ExecutionBar (AC-1, AC-5)
  - [ ] Subtask 1a: Read ExecutionBar.svelte lines 150–210
  - [ ] Subtask 1b: Replace `#22d3ee` with `var(--accent-cyan)` at every occurrence
  - [ ] Subtask 1c: Replace `#fb923c` with `var(--accent-orange)` at every occurrence
  - [ ] Subtask 1d: Convert any `rgba(34, 211, 238, ...)` or `rgba(251, 146, 60, ...)` to `color-mix`

- [ ] Task 2: Replace tab-close hover in AgentDetail (AC-2)
  - [ ] Subtask 2a: Update line 821 to `color: var(--accent-red);`
  - [ ] Subtask 2b: Grep the rest of the file to ensure no other `#ff5f57` slipped in

- [ ] Task 3: Replace Settings indicators and text-shadow (AC-3)
  - [ ] Subtask 3a: Line 860 → `var(--text-muted)`
  - [ ] Subtask 3b: Line 864 → `var(--text-dim)`
  - [ ] Subtask 3c: Line 466 → `var(--glow-spread) color-mix(in srgb, var(--accent-green) 60%, transparent)`

- [ ] Task 4: Replace ProcessNode and WorkflowBuilder rgbas (AC-4)
  - [ ] Subtask 4a: ProcessNode line 126 box-shadow
  - [ ] Subtask 4b: WorkflowBuilder line 883 (two rgbas: 10% and 15%)

- [ ] Task 5: Write component tests verifying theme propagation (AC-5)
  - [ ] Subtask 5a: Test overriding `--accent-cyan` on ExecutionBar
  - [ ] Subtask 5b: Test overriding `--accent-orange` on ExecutionBar
  - [ ] Subtask 5c: Test overriding `--accent-red` on AgentDetail `.tab-close:hover`

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
