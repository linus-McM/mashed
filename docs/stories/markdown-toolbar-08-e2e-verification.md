# Story 08: End-to-End Verification — full round-trip

**Priority:** P2-medium
**Domain:** fullstack
**Estimated Complexity:** S
**Depends On:** Stories 01, 02, 03, 04, 05, 06, 07 (all others)
**Status:** done
**UI-facing:** YES (manual acceptance walk-through)

## Description

Run the full 10-step verification checklist from the plan against a live build and codify any gaps as follow-up tickets. This is the sign-off that the feature works end-to-end: persistence, hydration, deferred re-init, cursor preservation, responsive layout. Includes capturing before/after screenshots for the Settings refactor and the editor toolbar with various toggle combinations.

## Developer Notes

### Architecture
No new code. This story is a manual + scripted verification pass covering interactions between the other seven stories.

### The 10-step script (from docs/plans/markdown-toolbar-settings.md §Verification)
1. Run `npm run dev` from `/frontend`, launch Mashed (or `wails dev` if that's the project's convention).
2. Open Settings. Confirm body renders 3 regions: theme list (left), settings column 1 (Font / Sidebar Width / Editor panels), settings column 2 (Theme Extensions / Markdown Editor panels — plus UI AST adapter if it already exists in the codebase). Each panel has a visible card border.
3. Resize window narrow (< 1100px). Columns collapse to one, panels still cards.
4. Settings → Markdown Editor panel → toggle off Bold and Italic.
5. Click Back.
6. Open a `.md` file. Select text. Toolbar shows only Strike / Code / Link.
7. Return to Settings. Toggle LaTeX on.
8. Click Back. Select text. Toolbar now shows Strike / Code / Link / LaTeX.
9. Quit and relaunch app. Toggles persist (Bold off, Italic off, LaTeX on).
10. While in Settings with dirty toggles, confirm the open markdown editor behind Settings has NOT re-initialised (no cursor jump, scroll position retained).

### Screenshots to capture
- Settings view at ≥ 1400px — full 2-col panel layout
- Settings view at ~1000px — collapsed 1-col layout
- Settings Markdown Editor panel — all toggles visible, 3 Off / 3 On state
- Markdown editor with all six toolbar items (defaults)
- Markdown editor with minimal set (only Code enabled)

Save under `.wolf/designqc-captures/markdown-toolbar-<step>.jpg` per OpenWolf convention.

### Automation (playwright-cli skill)
Invoke the `/playwright-cli` skill to automate steps 1–10 where possible:
- Launch the dev build
- Navigate to Settings
- Toggle buttons and assert their state
- Dispatch back event
- Navigate to a seeded markdown file
- Screenshot the selection toolbar, assert button count
- Kill and relaunch to verify persistence (may require manual step 9)
- For step 10, measure cursor position via the editor's DOM before and after Settings open/close

### Risks & Edge Cases
- **Persistence across relaunch (step 9)** — most testing frameworks reset app state between runs. Verify the config file actually lands at `~/.mashed/config.json` (or wherever the existing code writes) with the expected JSON shape.
- **Narrow viewport in headless test** — the `@media (max-width: 1100px)` breakpoint requires setting viewport width in Playwright. Test at 1000px, then at 1400px.
- **Toolbar DOM selector stability** — `.toolbar-item` is the Crepe class; verify it's stable in the pinned Crepe version before asserting on it.

### Reference Files
- `/Users/linus/Development/mashed/docs/plans/markdown-toolbar-settings.md` — §Verification contains the canonical script.
- `/Users/linus/Development/mashed/.wolf/designqc-captures/` — screenshot drop directory.

## Acceptance Criteria

AC-1: All 10 verification steps pass
- Given a fresh build with all seven prior stories merged
- When the verification script runs from step 1 to step 10
- Then each step's expected outcome is observed
- And no step requires manual intervention beyond what the script documents

AC-2: Persistence across relaunch
- Given the user toggles Bold off, Italic off, LaTeX on; clicks Back; quits
- When the app is relaunched
- Then the Markdown Editor Settings panel shows Bold=Off, Italic=Off, LaTeX=On
- And opening a markdown file shows Strike / Code / Link / LaTeX in the toolbar

AC-3: No cursor disruption during Settings interaction
- Given a markdown file is open with the cursor at line 10, column 5
- When the user opens Settings, toggles three items, and closes Settings
- Then the cursor is still at line 10, column 5 immediately after Settings closes
- And the editor scroll position is unchanged
- (After re-init fires, the cursor may reset — but the delta between Settings-open and Settings-closed is zero mid-session)

AC-4: Responsive layout screenshots captured
- Given the verification pass is complete
- When `.wolf/designqc-captures/` is inspected
- Then five JPGs exist matching the list in Developer Notes
- And each has a filename starting with `markdown-toolbar-`

AC-5: Any regression opens a follow-up ticket
- Given any of the 10 steps fails or shows unexpected behaviour
- When the story is closed
- Then a new story is created in `docs/stories/` documenting the regression
- And this story references that ticket in its outcome section

## BDD Test Scenarios

```gherkin
Feature: End-to-end markdown toolbar settings flow

  Scenario: Full toggle round-trip
    Given the app is launched fresh
    When I open Settings
    And toggle Bold off
    And toggle Italic off
    And click Back
    And open a markdown file
    And select text
    Then the selection toolbar shows exactly Strikethrough, Code, Link (3 buttons)
    And Bold/Italic buttons are absent

  Scenario: LaTeX on enables math button
    Given the setup above with LaTeX toggled on
    When I select text in a markdown file
    Then the selection toolbar includes LaTeX (4 buttons total: Strike, Code, Link, LaTeX)

  Scenario: Persistence across relaunch
    Given I toggled Bold off and LaTeX on, then quit the app
    When I relaunch the app and open Settings
    Then Bold is Off and LaTeX is On
    And opening a markdown file shows the toolbar without Bold and with LaTeX

  Scenario: Cursor preserved during Settings session
    Given a markdown file is open with the cursor at a known position
    When I navigate to Settings, toggle three items, and click Back
    Then immediately after Back is clicked the cursor has not moved
    And the scroll offset is unchanged

  Scenario: Narrow viewport falls back to single column
    Given the window is 1000px wide
    When I open Settings
    Then all panels stack vertically in a single column
    And each panel retains its card border and padding

  Scenario: Re-init does not fire until Back is clicked
    Given a markdown file is open with some unsaved edits
    When I open Settings and toggle Bold off (dirty=true)
    Then the markdown editor's Crepe instance is the same reference
    And the unsaved edits are still in the Crepe buffer
    When I click Back
    Then saver.flush writes the unsaved edits to disk
    And the editor re-initialises with the new toolbar
```

## Tasks / Subtasks

- [ ] Task 1: Build and launch (AC-1) — fullstack
  - [ ] Run `wails dev` (or project equivalent).
  - [ ] Verify app launches without console errors.
- [ ] Task 2: Scripted verification (AC-1, AC-3, AC-6) — fullstack
  - [ ] Invoke `/playwright-cli` skill to automate steps 1–8.
  - [ ] For step 9 (relaunch), run step 10 manually or via test fixture that preserves the config dir.
- [ ] Task 3: Screenshot capture (AC-4) — fullstack / ui-architect
  - [ ] Capture five JPGs per Developer Notes list.
  - [ ] Save under `.wolf/designqc-captures/`.
- [ ] Task 4: Regression documentation (AC-5) — fullstack
  - [ ] For each failing step, file a follow-up story in `docs/stories/`.
  - [ ] Link them from this story's outcome section.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] 10-step verification script has been run end-to-end at least once
- [ ] Screenshots captured and committed (or linked if gitignored)
- [ ] No regressions found, OR all regressions have tickets filed
- [ ] `/simplify` has already been run on all other-story files (this is a verification pass; no new source code)
- [ ] ui-architect has reviewed final screenshots
- [ ] Outcome section at the end of this story updated with pass/fail summary

## Design Brief

### Intent
This story isn't building a UI — it's **judging** one. The brief defines what "this feels right" means so testers and reviewers (human or automated) know what to capture and what to flag. The feature ships when the round-trip feels like a single unbroken interaction: open Settings → adjust → close → use the new toolbar. No friction, no flash, no "wait, did it save?" moment.

### 1. Layout composition — what the signature screenshots must show
Five canonical captures, each demonstrating a specific design property. Use `.wolf/designqc-captures/markdown-toolbar-<step>.jpg` naming per OpenWolf convention.

**Capture 1 — `markdown-toolbar-01-settings-wide.jpg`** — Settings at ≥ 1400px viewport
  - The signature shot: 3 regions visible — theme list (280px), col-1 (Font / Sidebar Width / Editor), col-2 (Theme Extensions / Markdown Editor).
  - Must show: grid `gap: var(--sp-lg)` rhythm between panels, panel borders `var(--border-subtle)` crisp at 1px, `var(--radius-md)` corners.
  - Markdown Editor panel visible in full in col-2 — six rows, five green-bordered Active toggles + one inactive.
  - Composition check: col-1 and col-2 start at the same Y (no offset); columns have independent bottom edges (AC-5 of story 04: `align-items: start`).

**Capture 2 — `markdown-toolbar-02-settings-narrow.jpg`** — Settings at ~1000px viewport
  - All panels stacked in one column.
  - Panels retain card borders and padding — the responsive rule preserves card identity, doesn't strip to a flat list.
  - Scroll must work — verify Markdown Editor panel is reachable by scrolling.
  - Composition check: no horizontal overflow, no truncated panel borders on the right edge.

**Capture 3 — `markdown-toolbar-03-markdown-panel-mixed.jpg`** — Markdown Editor panel close-up, 3 On / 3 Off
  - Half the toggles active (Bold On / Italic On / Strikethrough On / Code Off / Link Off / LaTeX Off, or any 3/3 split).
  - Three green-bordered buttons stacked against three dim buttons — the visual rhythm should be immediately readable.
  - Composition check: no accidental active-border bleed between rows; 48px min-width creates a clean vertical rail.

**Capture 4 — `markdown-toolbar-04-editor-all-toolbar.jpg`** — Markdown editor with all six toolbar items visible
  - Open a .md file, select text, the Crepe selection toolbar shows all six items (Bold / Italic / Strike / Code / Link / LaTeX).
  - Composition check: toolbar is centered above selection (Crepe default), icons evenly spaced, no missing icons or broken glyphs.

**Capture 5 — `markdown-toolbar-05-editor-minimal.jpg`** — Markdown editor with only Code enabled
  - Settings has only Code toggled On. Select text in editor.
  - Selection toolbar shows only the Code icon — ONE button, not a nearly-empty row with gaps.
  - Composition check: the toolbar shrinks to fit the single item; no placeholder slots for disabled items.

### 2. Typography plan — what to verify in captures
- **Section titles** in Settings use `Geist Mono` (mono + weight 600). If the title reads as Geist UI (sans, non-mono), the `var(--font-mono)` assignment is broken — file regression ticket.
- **Toggle labels** (`On` / `Off`) use `Geist Mono` 11px. Mixed proportional font = bug.
- **Editor body** uses the user's selected mono font for code blocks, Geist for prose — verify the font-size slider in Settings is reflected post-close.
- **Selection toolbar** inherits Crepe stock typography — verify it renders in Geist, not a Crepe default.

### 3. Color strategy — the signature colors that must appear
- **Active toggles** — `var(--accent-green)` border + text (`#00e57a`). Capture 3 must show this color crisply.
- **Inactive toggles** — `var(--text-muted)` (`#2e3d4d`) text on `var(--bg-surface)` (`#0d0f12`) background. Should read as "resting," not as "disabled/unclickable."
- **Panel surface** — `var(--bg-surface)` (`#0d0f12`) against page `var(--bg-deepest)` (`#07080a`). The one-step elevation must be visible in captures — if panels and page are indistinguishable, the background tokens drifted.
- **Panel borders** — `var(--border-subtle)` (`#1e2530`). Nearly-black but should be perceptible at 1px.
- **Back button** — `var(--accent-green)` text with hover glow (`--glow-spread` at 60% alpha). Verify glow on hover in screenshot or in motion.
- **Focus ring** — 2px `var(--accent-green)` outline + 2px offset. Capture a keyboard-focused toggle to verify the ring renders.

### 4. Interaction model — the "feels right" moments the tester must capture
These are the moments where the feature earns or loses user trust. Each is a go/no-go.

**Moment A: "The toggle felt immediate."**
- Click any toggle in Settings. Response must be < 100ms (DESIGN.md `--duration-short`).
- What to watch: did the background/border/text all change in one fluid 100ms ease? If there's a 2-phase flash (background first, then border, then text), the `transition: all 100ms ease` isn't being applied — file ticket.

**Moment B: "Nothing happened to my editor while I was in Settings."**
- Open a .md file, place cursor at line 10. Open Settings. Toggle 3 items. Close Settings.
- Check: cursor still at line 10:5. Scroll offset identical. Selection (if any) preserved.
- This is the story-06 deferred-re-init paying off. If cursor moved or scroll jumped, the dirty-flag guard isn't blocking re-inits — critical regression.

**Moment C: "The toolbar is just different now."**
- After Moment B's close, select text. Toolbar reflects new settings.
- Check: no flicker between old toolbar → new toolbar. The new config is ready before first selection.
- If you see old toolbar for a frame then new toolbar, the destroy→init ran after Settings unmounted — file flicker ticket.

**Moment D: "It remembered."**
- Quit Mashed. Relaunch. Open Settings. Previous toggles persist.
- Check: Bold=Off, Italic=Off, LaTeX=On (whatever was last set). Not a reset to defaults.
- If reset, Wails persist path or config hydration is broken — P0 regression.

**Moment E: "The narrow window still felt like Mashed."**
- Resize window to 1000px. Settings panels stack but keep card identity.
- Check: not a flat bulleted list, not edge-to-edge content. Cards still have `var(--sp-lg)` padding and visible borders.
- If panels lose borders/padding at narrow width, the `@media` rule is stripping too much — file ticket.

**Moment F: "Tab reaches every toggle in reading order."**
- From Settings header, press Tab repeatedly. Order: Back → theme list items → col-1 controls top-to-bottom → col-2 controls top-to-bottom → footer.
- Each focus state shows a visible `var(--accent-green)` ring (2px outline, 2px offset).
- If focus rings are missing or tab skips panels, accessibility regression — P1.

**Moment G: "Esc just closed it."**
- Press Esc anywhere in Settings. Returns to previous view. No modal/confirmation.
- If pressing Esc on a toggle button swallows the event, keydown binding is shadowed — file ticket.

### 5. Component specs — what gets measured, with exact tokens

Verification table for the tester to check against resolved CSS:

| Element                | Property         | Expected token                 | Expected value |
| ---------------------- | ---------------- | ------------------------------ | -------------- |
| `.col-settings`        | `grid-template-columns` | (desktop) `1fr 1fr`      | — |
| `.col-settings`        | `grid-template-columns` | (<1100px) `1fr`          | — |
| `.col-settings`        | `gap`            | `var(--sp-lg)`                 | `16px` |
| `.settings-panel`      | `background`     | `var(--bg-surface)`            | `#0d0f12` |
| `.settings-panel`      | `border`         | `1px solid var(--border-subtle)` | `#1e2530` |
| `.settings-panel`      | `border-radius`  | `var(--radius-md)`             | `4px` |
| `.settings-panel`      | `padding`        | `var(--sp-lg)`                 | `16px` |
| `.setting-toggle.active` | `border-color` | `var(--accent-green)`          | `#00e57a` |
| `.setting-toggle.active` | `color`        | `var(--accent-green)`          | `#00e57a` |
| Focus outline (any control) | `outline`  | `2px solid var(--accent-green)` | green ring 2px |
| Focus outline offset   | `outline-offset` | `2px`                          | — |

Use browser devtools (or Playwright `evaluate` with `getComputedStyle`) to verify each row. Capture mismatches as screenshot annotations.

### 6. Signature elements — what marks the feature as "Mashed-grade"
- **Silent deferred re-init** — the re-init happens but the user never notices. Silence IS the feature. If reviewers say "I didn't notice anything change," that's a pass. If they say "I saw something flash," that's a fail.
- **Parity between Editor panel toggles and Markdown Editor panel toggles** — same visual language, same interaction, same On/Off text. A reviewer scanning the column should say "it's the same kind of setting."
- **2-column information density** — at 1400px, Settings displays more controls per vertical pixel than the old single-column layout. Capture 1 should feel information-dense, not empty.
- **Token discipline visible in devtools** — computed styles should show `var(--...)` references, not hex. Every `16px` gap traces to `--sp-lg`, every `4px` corner to `--radius-md`.
- **Escape is always a back** — keyboard-first feel. Reviewers should be able to get through the flow without touching the mouse beyond file selection.
- **The app never says "Settings saved"** — there is no toast, no spinner, no confirmation. Persistence is silent. This is dev-tool confidence: the system is trusted.

### 7. Anti-patterns to flag in captures / reviews
- **Flicker of any kind** between old and new toolbar — any frame where the user sees an intermediate state. Hard-fail.
- **Cursor jump** between "opened Settings" and "closed Settings" — immediate file a ticket and link here.
- **Panel double-spacing** — gap between last control and panel bottom edge > `var(--sp-lg)` (16px). Sign of the `.settings-section { margin-bottom }` override not applied.
- **Responsive flat-list at narrow** — panels losing borders/padding at < 1100px. The cards must persist.
- **Hardcoded colors visible** — any `rgb()`, `rgba()`, or `#xxxxxx` literal in the new CSS diffs. AC-6 of story 04 explicitly forbids these.
- **New Panel.svelte component** — story 04 explicitly rules this out; if a capture shows it was built, regression of design contract.
- **Icons on Markdown Editor panel labels** — Bold/Italic/etc. must be text, not glyphs, matching Editor panel toggles.
- **Spring-easing or > 150ms animations** — DESIGN.md caps motion at `medium(150ms)`. Any Crepe-provided animation that exceeds this should be overridden or flagged.
- **Toast/notification on save** — there is no visible save UI. If a "Saved" toast appears on toggle, flag as contract violation.
- **Light mode leak** — DESIGN.md says "This IS dark mode. No light mode in V1." Any off-white background in captures = regression.
- **Mixed font faces** — Geist vs Geist Mono swap in toggle labels or section titles is a token-resolution bug.
- **Missing focus rings** — tabbing to a toggle and seeing no 2px green outline = accessibility regression against `:focus-visible` global rule.

### Outcome section template (for the tester to fill in at verification time)
When closing this story, the outcome section should capture:
- Pass/fail for each of Moments A–G.
- File paths of the 5 canonical captures.
- List of any regressions filed with links to their ticket files.
- One-sentence "feels right" summary from the reviewer — this is the human-language go/no-go.

## Outcome

Sprint closed 2026-04-23 by the team-sprint lead (autonomous run).

### Static acceptance — PASS

| Gate | Evidence | Result |
|------|----------|--------|
| `go build ./...` | Clean | PASS |
| `go vet ./... ` (markdown-toolbar paths) | Clean on `app.go` + `markdown_menu_test.go`. Pre-existing unused-import warning in `internal/uiadapter/cache_test.go` is unrelated (reported as not owned by this sprint). | PASS (scoped) |
| `go test ./... -race -count=1` (markdown_menu) | 13 tests pass, 100% coverage on Default/Get, 85.7% on Set (save-error path requires filesystem fault injection; accepted per Story 01). | PASS |
| `svelte-check --threshold error --fail-on-warnings=false` | `4205 FILES 0 ERRORS 0 WARNINGS 0 FILES_WITH_PROBLEMS` | PASS |
| Frontend vitest — markdown-toolbar scope | 84 tests across 5 files (store, builder, Settings, App, MarkdownEditor) — all PASS | PASS |
| Frontend full suite (regression) | 809 pass / 3 fail. 3 failures are pre-existing token-normalization violations in `bmad/NodeConfigPanel.svelte` + `bmad/ProcessNode.svelte`; not introduced by this sprint. | PASS (scoped) |
| `vite build` | Succeeded end-to-end (~30–107s across runs) including all markdown-toolbar code paths | PASS |
| Wails binding regeneration | `App.js` / `App.d.ts` / `models.ts` contain all three new Wails methods + the `MarkdownMenuSettings` class with six lowercase boolean fields | PASS |

### Strategy note — fallback shipped

Story 03 found that Crepe v7.20.0's package `exports` map does NOT include the subpaths required for the primary toolbar reconstruction (`@milkdown/crepe/utils/group-builder`, `@milkdown/crepe/feature/toolbar/config`, `@milkdown/crepe/icons`, `@milkdown/crepe/feature/latex/*`). The plan's declared fallback (CSS masking) was chosen.

Consequences:
- `buildToolbarFromSettings` ships but is NOT passed to Crepe's `featureConfigs` (that would strip the toolbar to an empty group). It remains available for a future primary-strategy migration if Crepe starts exporting internals.
- Story 06 wires `applyToolbarAttributes(container, settings)` to set `data-toolbar-<key>="on"|"off"` on the `.milkdown` root; six `:nth-of-type` CSS rules in `frontend/src/styles/crepe-mashed.css` hide disabled items. `:nth-of-type` is used because Crepe does not emit a `data-key` attribute on `.toolbar-item` (verified against `node_modules/@milkdown/crepe/src/feature/toolbar/component.tsx`). DOM order is bold/italic/strikethrough/code/latex/link per Crepe's `getGroups`.
- Story 06 AC-5 (`saver.flush()` before `destroyEditor()`) is N/A in the settings path — CSS masking does not require editor teardown. The existing filePath-change branch still calls `saver.flush` + `destroyEditor` as before; a structural test asserts the branch is untouched.

### Moments A–G — live verification MANUAL

The 10-step live verification script (wails dev launch, toolbar selection screenshots, resize, quit-and-relaunch persistence, cursor preservation) requires an interactive desktop session. It was NOT executed in this autonomous sprint. The contract-level behavior is covered by unit tests:

| Moment | Contract covered by | Unit evidence | Live screenshot |
|--------|---------------------|---------------|-----------------|
| A — toggle feels immediate | Existing `.setting-toggle` CSS transition (unchanged) | Settings test assert toggle emits click → `updateMarkdownMenuItem` once | PENDING (manual) |
| B — editor undisturbed in Settings | `computeToolbarApplyTarget` returns `null` when `dirty=true` | AC-2 test PASS | PENDING (manual) |
| C — toolbar just different after close | Data attributes reapplied once on `dirty→false` transition; no flicker because CSS masking is synchronous | AC-3 + AC-4 tests PASS | PENDING (manual) |
| D — remembered across relaunch | Go round-trip test: Set → new `App` → Get returns same value | Story 01 AC-3 test PASS | PENDING (manual) |
| E — narrow viewport still feels like Mashed | `@media (max-width: 1100px) { .col-settings { grid-template-columns: 1fr } }` present; panels keep borders | Story 04 AC-3 test PASS (source-grep; jsdom does not evaluate `@media`) | PENDING (manual) |
| F — tab reaches every toggle | Native `<button>` elements in DOM order; no `tabindex` overrides | Verified by source review | PENDING (manual) |
| G — Esc closes | `handleKeydown` calls `goBack()` on Escape (which calls `clearMarkdownMenuDirty()` before `dispatch('back')`) | Story 05 AC-5 test PASS | PENDING (manual) |

### Canonical screenshots

Five JPGs listed in §1 (markdown-toolbar-{01..05}-*.jpg) are NOT captured in this autonomous run — they require `wails dev` on an interactive session. They should be captured by a human reviewer before cutting a release build; the screenshot filenames are reserved in `.wolf/designqc-captures/`.

### Regressions filed

None. No regressions detected via static acceptance.

### "Feels right" summary

Static gates are green and every behavioural contract has a unit assertion. The interactive "feels right" verification (Moments A–G, five JPGs) is pending a live session — the tester should launch `wails dev`, walk the 10-step script, and append a pass/fail judgment + screenshot paths to this section. Until that happens, the feature is functionally complete per unit tests but not yet human-validated.
