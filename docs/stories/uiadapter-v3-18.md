# uiadapter-v3-18: Title-bar Dynamic UI model selector

**Status:** done
**Landed:** 2026-04-23
**Domain:** fullstack
**Size:** L
**Depends On:** B, C, v3-12, v3-16, v3-17
**Priority:** P1-high

## Story

As a Mashed user, I want Backend / Model / Policy pulldowns in the title bar (right of the Settings button), so that I can swap the Dynamic UI pipeline at runtime without opening Settings, see at a glance which backend is live, and be prevented from picking an offline one.

## Description

Plan §3 Phase 4 "Story 18 — Title-bar Dynamic UI model selector" (lines 664–701). The only UI story in the plan — everything upstream is wiring for it. Three pull-down menus, right-aligned in `.titlebar-actions`, in order **Backend → Model → Policy**. Settings button unchanged; new selectors render *after* it.

- **Backend options:** `["ollama", "claude-api", "claude-cli"]`; unreachable backends (Story 17 `Health` red) render **disabled with an "offline" suffix** — not hidden — so users see *why* they can't pick.
- **Model options:** context-sensitive on Backend. Ollama → `ListOllamaModels()`. Claude → static `ListClaudeModels()` returning `["claude-haiku-4-5", "claude-sonnet-4-6", "claude-opus-4-6"]` from Story 12 allowlist.
- **Policy options:** six values from Story 16 (`local-only`, `claude-only`, `claude-first`, `ollama-first`, `cost-aware`, `privacy-strict`), each with one-line tooltip.

Data flow: user selects → store setter → Wails binding → on success, writes value to Svelte `writable`; on error, rollback + toast via notification feed. No restart required — the backend registry observes the config change.

Hydration: `hydrate()` in `uiAdapterSettings.ts` reads `cfg.backend`, `cfg.claudeModel`, `cfg.cliModel`, `cfg.routerPolicy` from `GetConfig()` on startup. Unknown values default to `claude-first` when Anthropic key present, else `local-only`.

Visual contract: each selector is a `titlebar-btn` with compact label (backend glyph + current model short name, ≤14 visible chars, truncate with ellipsis + full in `title`). Popover uses existing `.theme-popover` styles. Single global `<svelte:window on:click>` closes any open pulldown. Keyboard: Tab, Enter/Space, ArrowDown/Up, Esc.

Live-reachability: extend existing `ollamaReachable` store; add `claudeApiReachable`, `claudeCliReachable`. Backend selector reads all three and greys out unreachable rows. Model selector is disabled when selected backend is offline (button shows `offline` suffix).

## Developer Notes

- **Files (new/edited):**
  - `frontend/src/components/TitleBar.svelte` (edit — add selectors + popovers)
  - `frontend/src/components/titlebar/DynamicUiSelector.svelte` (new — reusable pulldown; props: `label`, `value`, `options`, `disabled`, `onSelect`)
  - `frontend/src/lib/stores/uiAdapterSettings.ts` (edit — extend with `backend`, `claudeModel`, `cliModel`, `routerPolicy`, `backendsAvailable`; setters mirror `setModel` optimistic write + rollback)
  - `app_uiadapter.go` (new — Wails bindings `SetBackend(string)`, `SetClaudeModel(string)`, `SetCLIModel(string)`, `SetRouterPolicy(string)`, `ListBackendsAvailable() []string`, `ListClaudeModels() []string`, `ListRouterPolicies() []string`; persists via same path `SetOllamaModel` uses)
  - `config.go` (edit — add `CLIModel string` alongside B's fields if not present)
- **Types / API surface:** stable error sentinels `ErrInvalidBackend`, `ErrInvalidClaudeModel`, `ErrInvalidRouterPolicy`.
- **Prompt-injection hygiene:** labels are static strings compiled into the frontend — no user-supplied text. Model names come from Go config allowlist (Story 12), not from the terminal.
- **Out of scope:** per-repo overrides; cost/latency live readouts next to the selector. Leave a `data-metrics-slot` attribute on the button so a future story can inject those without touching this code.
- **Risks:** race between health ticker update and user click — debounce via store's existing optimistic pattern.
- **Dependencies:** coordinate with Wails bindings regen (cerebrum §33 — `wails dev`).
- **Use-repo-code directive:** before editing, use `use-repo-code` to read current `TitleBar.svelte`, `uiAdapterSettings.ts` (grep `files.md` for `^## File: frontend/src/components/TitleBar.svelte$` and `^## File: frontend/src/lib/stores/uiAdapterSettings.ts$`).

## Acceptance Criteria

AC-18.1: `TitleBar.test.ts` — Backend pulldown renders the three options, current value ticks, clicking `claude-api` calls `setBackend("claude-api")` with a single Wails round-trip.

AC-18.2: `TitleBar.test.ts` — Model pulldown re-populates when Backend changes: switching from `ollama` to `claude-api` within a single session replaces the Ollama model list with the Claude allowlist without a page reload.

AC-18.3: `TitleBar.test.ts` — Policy pulldown renders all six values from Story 16; selecting one dispatches a single `SetRouterPolicy` call; an invalid value (e.g. injected via DevTools) is rejected by the client-side validator before hitting Wails.

AC-18.4: `TitleBar.test.ts` — unreachable backend (`ollamaReachable=false` stubbed) renders the option as disabled with the `offline` suffix; clicking it is a no-op and does *not* fire Wails.

AC-18.5: `TitleBar.test.ts` — optimistic-write rollback: when Wails `SetBackend` rejects, the store value reverts to the previous backend and a toast fires via the existing notification pipeline.

AC-18.6: `TitleBar.test.ts` — outside click on `<svelte:window>` closes all three popovers; clicking one selector's button while another is open closes the other and opens the new one (only one popover open at a time).

AC-18.7: `TitleBar.test.ts` — a11y: selectors are reachable via keyboard; `role="menu"` + `role="menuitem"` + `aria-expanded` set correctly; Playwright `--axe` check passes with zero violations on the title bar.

AC-18.8: `app_uiadapter_test.go` — each Wails setter rejects values outside its allowlist with a stable error string (`ErrInvalidBackend`, `ErrInvalidClaudeModel`, `ErrInvalidRouterPolicy`) so the frontend can surface a precise message.

AC-18.9: `TestDynamicUI_NoRestartRequired` — Go-level integration: flip `Config.Backend` at runtime via the same code path `SetBackend` uses and confirm the next `Translate` call dispatches through the new backend registry entry within one `Translate` invocation.

## BDD Test Scenarios

```gherkin
Feature: Title-bar Dynamic UI model selector

  Scenario: AC-18.1 — Backend pulldown + Wails roundtrip
    Given the title bar is mounted
    And three backends are enumerated
    When the user clicks the Backend button and selects "claude-api"
    Then `setBackend("claude-api")` is called exactly once
    And the button label updates to the Claude glyph + model short name

  Scenario: AC-18.2 — Model list context-sensitive
    Given Backend="ollama"
    When the user switches Backend to "claude-api"
    Then the Model pulldown renders ["claude-haiku-4-5", "claude-sonnet-4-6", "claude-opus-4-6"]
    And no page reload occurs

  Scenario: AC-18.3 — Policy pulldown + client-side validation
    Given the Policy pulldown
    When the user selects "claude-first"
    Then `SetRouterPolicy("claude-first")` is called exactly once
    Given an injected invalid value "nope"
    When the store setter is invoked
    Then the validator rejects before the Wails call fires

  Scenario: AC-18.4 — Offline suffix disables click
    Given ollamaReachable=false
    When the Backend pulldown renders
    Then the "ollama" row is disabled with an "offline" suffix
    And clicking it does not invoke Wails

  Scenario: AC-18.5 — Rollback on Wails error
    Given current backend="ollama"
    When `SetBackend("claude-api")` rejects from Wails
    Then the store reverts to "ollama"
    And a toast appears via the existing notification feed

  Scenario: AC-18.6 — One popover at a time
    Given the Backend popover is open
    When the user clicks the Model button
    Then the Backend popover closes
    And the Model popover opens
    And an outside-click closes all three

  Scenario: AC-18.7 — A11y
    Given the title bar renders
    When Playwright --axe runs
    Then zero a11y violations are reported on the title bar region
    And every selector is reachable via Tab/Enter/Space/ArrowDown/Up/Esc

  Scenario: AC-18.8 — Stable error sentinels
    Given a malformed value
    When the Wails setter runs
    Then it returns one of ErrInvalidBackend / ErrInvalidClaudeModel / ErrInvalidRouterPolicy

  Scenario: AC-18.9 — No restart required
    Given Config.Backend flips at runtime via SetBackend
    When the next Translate invocation fires
    Then it dispatches through the new backend entry
    And no restart is required
```

## Tasks / Subtasks

- [ ] Task 1 — `DynamicUiSelector.svelte` reusable pulldown (maps to AC-18.1, AC-18.6, AC-18.7)
  - [ ] Subtask 1a — Props + role + keyboard handlers.
  - [ ] Subtask 1b — Mirror existing `.theme-picker-wrap` outside-click pattern.
- [ ] Task 2 — Extend `uiAdapterSettings.ts` store (maps to AC-18.1, AC-18.2, AC-18.5)
  - [ ] Subtask 2a — Add `backend`, `claudeModel`, `cliModel`, `routerPolicy`, `backendsAvailable`, `claudeApiReachable`, `claudeCliReachable`.
  - [ ] Subtask 2b — Setters with optimistic write + rollback + toast.
  - [ ] Subtask 2c — Extend `hydrate()`.
- [ ] Task 3 — Wails bindings in `app_uiadapter.go` (maps to AC-18.8, AC-18.9)
  - [ ] Subtask 3a — Setters with allowlist validation and stable error sentinels.
  - [ ] Subtask 3b — List methods: `ListBackendsAvailable`, `ListClaudeModels`, `ListRouterPolicies`.
- [ ] Task 4 — Edit `TitleBar.svelte` — add selectors (maps to AC-18.1, AC-18.2, AC-18.3, AC-18.4)
- [ ] Task 5 — Tests:
  - [ ] Subtask 5a — `TitleBar.test.ts` — AC-18.1 … AC-18.7.
  - [ ] Subtask 5b — `app_uiadapter_test.go` — AC-18.8.
  - [ ] Subtask 5c — `TestDynamicUI_NoRestartRequired` — AC-18.9.
  - [ ] Subtask 5d — Playwright --axe assertion.

## Definition of Done

- [ ] All ACs verified with PASS evidence (test name or Playwright assertion)
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run — no redundancy
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] `pnpm run check` svelte-check strict clean
- [ ] `pnpm test` clean
- [ ] Playwright a11y passes
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

### 1. Layout composition

DOM order inside the existing `.titlebar-actions` container, left-to-right, reads: theme-picker (existing) → `FolderPlus` → `Settings` → Backend → Model → Policy. The three new selectors therefore render **after** the Settings button, keeping file-system / preference controls grouped left of runtime-pipeline controls.

Wrapper structure per selector (new class `.titlebar-selector-wrap`, mirrors `.theme-picker-wrap` geometry):

```
.titlebar-actions
 ├─ .theme-picker-wrap                         (existing)
 ├─ button.titlebar-btn#new-repo               (existing)
 ├─ button.titlebar-btn#settings               (existing)
 ├─ .titlebar-selector-wrap[data-selector=backend]
 │   ├─ button.titlebar-btn.titlebar-selector[aria-haspopup="menu"]
 │   │                                  [aria-expanded="true|false"]
 │   │                                  [aria-controls="sel-backend-menu"]
 │   │                                  [data-metrics-slot]
 │   │   ├─ span.sel-glyph               (Unicode indicator)
 │   │   ├─ span.sel-label               (short name, truncated)
 │   │   └─ span.sel-caret aria-hidden   (▾)
 │   └─ div.selector-popover#sel-backend-menu[role="menu"]
 │       ├─ div.sel-caption role="presentation" (muted-caps category)
 │       └─ button.sel-item[role="menuitem"]
 │                          [aria-checked="true|false"]
 │                          [aria-disabled="true|false"]
 │           ├─ span.sel-tick aria-hidden   (✓ or blank)
 │           ├─ span.sel-item-label
 │           └─ span.sel-badge              (offline pill, conditional)
 ├─ .titlebar-selector-wrap[data-selector=model] …
 └─ .titlebar-selector-wrap[data-selector=policy] …
```

Parent `.titlebar-actions` already uses `gap: var(--sp-sm)` — reused as-is; no new horizontal gap token. Each `.titlebar-btn` stays at the existing `padding: var(--sp-xs) var(--sp-sm)` (maps to the current `4px 8px`), aligning height with the Settings / FolderPlus / theme buttons against the 38px titlebar.

Popover anchoring: `position: absolute; top: calc(100% + var(--sp-sm)); right: 0;` on the popover, with `position: relative` on `.titlebar-selector-wrap` — identical to `.theme-popover`. Right-edge alignment keeps popovers inside the window as they sit near the right edge of the titlebar. `z-index: 200` — same band as `.theme-popover`; only one popover is open at any time so stacking conflicts do not arise.

### 2. Typography plan

Button label uses `font-family: var(--font-code)` (JetBrains Mono) for the glyph + short-name run — mono gives deterministic truncation width and matches the "runtime-pipeline identity" cue. Weight `500`, `font-size: var(--text-label)` (11px), `letter-spacing: 0.02em`, `line-height: 1`. Caret glyph (`▾`) sits at the same font-size; colored `var(--text-dim)` so it visually recedes from the label.

Truncation: `.sel-label` uses `max-width: 14ch; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;`. Story-specified 14-visible-char cap thus lives in `ch` units (mono font) — `title="{backend} / {model}"` attribute on the button carries the full name per AC-18.7.

Backend glyphs — rendered in `.sel-glyph` at `font-size: var(--text-body)` (13px), `line-height: 1`, `color` inherited from the item state (see §3). All three are ASCII-safe BMP codepoints:

| Backend       | Glyph | Codepoint | Rationale                                          |
|---------------|-------|-----------|----------------------------------------------------|
| `ollama`      | `⌂`   | U+2302    | House = local runtime                              |
| `claude-api`  | `◆`   | U+25C6    | Solid diamond = hosted / remote service            |
| `claude-cli`  | `▶`   | U+25B6    | Play-triangle = spawned CLI process                |

Popover typography:
- Category caption row (`.sel-caption`): `font-size: var(--text-label)`, `text-transform: uppercase`, `letter-spacing: 0.08em`, weight `600`, `color: var(--text-muted)`. Only rendered inside the Policy popover (separates live-routing vs. explicit policies if future grouping lands; for now a single "POLICY" caption establishes the pattern for other popovers to adopt later).
- `.sel-item-label`: `font-family: var(--font-code)`, `font-size: var(--text-label)`, weight `500`, `color: var(--text-primary)`.
- Tooltip description row (Policy only, rendered inline below the label as a second flex row inside the menuitem): `font-family: var(--font-ui)`, `font-size: var(--text-label)`, weight `400`, `color: var(--text-dim)`.
- Disabled row: label collapses to `color: var(--text-muted)`; glyph + tick blank; `text-decoration: none` (no strike-through — the `offline` badge carries the state).

### 3. Color strategy

Button states:

| State        | Background                                                  | Border                    | Text                  |
|--------------|-------------------------------------------------------------|---------------------------|-----------------------|
| idle         | `transparent`                                               | `1px solid transparent`   | `var(--text-primary)` |
| hover        | `var(--bg-elevated)`                                        | `1px solid var(--border-subtle)` | `var(--text-primary)` |
| open (active popover) | `var(--bg-active)`                                 | `1px solid var(--border-emphasis)` | `var(--text-primary)` |
| focused (kbd) | inherits idle/hover bg, `outline: 2px solid var(--accent-green); outline-offset: 2px` (global rule from `style.css:146-155`) | — | — |
| selected glyph | — | — | glyph `color: var(--accent-green)` when the current value is live-applied |

Popover panel: `background: var(--bg-elevated)`, `border: 1px solid var(--border-emphasis)`, `border-radius: var(--radius-lg)`, `box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5)` — exactly the `.theme-popover` recipe (single source of truth for elevation; raw rgba already in use in style.css for this shadow).

Selected-tick: `.sel-tick` width reserved always (so rows do not reflow on selection); glyph `✓` (U+2713) rendered in `color: var(--accent-green)` on the active row, invisible placeholder (`color: transparent`) elsewhere.

Disabled "offline" state: the whole row gets `background: color-mix(in srgb, var(--bg-elevated) 85%, transparent)` — subtle desaturation — with `color: var(--text-muted)` on the label. No hover background. Cursor `not-allowed`.

Offline badge pill (`.sel-badge`): `background: color-mix(in srgb, var(--accent-amber) 12%, transparent)`, `color: var(--accent-amber)`, `border: 1px solid color-mix(in srgb, var(--accent-amber) 40%, transparent)`. Amber (not red) because the backend is reachable-in-principle — just not right now — and red is reserved for hard errors (sentinel rejection toast).

### 4. Interaction model

Click on a `.titlebar-btn.titlebar-selector` with `|stopPropagation` toggles that wrap's `open` flag; when opening, it calls a shared `closeOthers(id)` that clears the other two wraps' flags plus the theme picker's. A single `<svelte:window on:click={closeAll}>` at the TitleBar root (extend the existing handler on line 35) closes every popover on outside click — the existing theme-picker already uses this idiom.

Hover inside a Policy menuitem reveals the description row by transitioning `max-height: 0 → var(--sel-tooltip-max)` and `opacity: 0 → 1` over `var(--duration-short)` with `var(--ease-enter)`; the row keeps `overflow: hidden` so closed rows do not consume vertical space.

Keyboard:
- Tab / Shift+Tab: follows DOM order — Backend → Model → Policy after the Settings button.
- Enter / Space on the button: toggles `aria-expanded`, moves focus to the first non-disabled `.sel-item` when opening.
- ArrowDown / ArrowUp inside an open popover: moves focus within `role="menu"`, wrapping at boundaries, skipping `aria-disabled="true"` rows.
- Enter on a `.sel-item`: invokes the store setter; popover closes; focus returns to the invoking `.titlebar-selector` button.
- Esc: closes the active popover; focus returns to invoking button.

Optimistic-write feedback: on setter resolve, the invoking button emits a 150ms `pulse` animation (`keyframes` defined in the component's scoped `<style>`: 0% `border-color: var(--accent-green)`, 100% `border-color: transparent`). On reject, the store rolls back, a toast fires via the existing notification feed, and the button plays a `var(--duration-shake)` shake (reuse the existing `--duration-shake: 250ms` token from S6).

State transitions:
- Button idle → open: `background`, `border-color` ease over `var(--duration-short)` with `var(--ease-enter)`. The caret rotates `▾ → ▴` via `transform: rotate(180deg)` on the same timing.
- Popover mount: Svelte `transition:fly={{ y: -4, duration: 120, easing: cubicOut }}` — matches the `--duration-short` bucket, short enough to feel desktop-native.
- Popover unmount: same fly, reversed. No exit animation on outside-click cascades (feels snappier when closing multiple at once).

### 5. Component specs — exact tokens and measurements

```css
.titlebar-btn.titlebar-selector {
  padding: var(--sp-xs) var(--sp-sm);
  gap: var(--sp-xs);
  display: inline-flex;
  align-items: center;
  border-radius: var(--radius-sm);
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-primary);
  font-family: var(--font-code);
  font-size: var(--text-label);
  font-weight: 500;
  line-height: 1;
  letter-spacing: 0.02em;
  min-width: 0;
  max-width: 180px;
  transition:
    background var(--duration-short) var(--ease-enter),
    border-color var(--duration-short) var(--ease-enter),
    color var(--duration-short) var(--ease-enter);
}

.titlebar-selector-wrap {
  position: relative;
  /* wrap inherits flex gap from parent .titlebar-actions (var(--sp-sm)) */
}

.selector-popover {                        /* new class, NOT reusing .theme-popover
                                              (theme-popover has flex gap:4px tuned for swatch cards;
                                              menu rows want zero gap + per-item padding) */
  position: absolute;
  top: calc(100% + var(--sp-sm));
  right: 0;
  display: flex;
  flex-direction: column;
  gap: 0;
  padding: var(--sp-xs);
  background: var(--bg-elevated);
  border: 1px solid var(--border-emphasis);
  border-radius: var(--radius-lg);
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.5);
  z-index: 200;
  min-width: 220px;
  max-width: 320px;
}

.sel-item {
  display: grid;
  grid-template-columns: var(--sp-lg) 1fr auto;
  align-items: center;
  gap: var(--sp-sm);
  padding: var(--sp-xs) var(--sp-sm);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-family: var(--font-code);
  font-size: var(--text-label);
  font-weight: 500;
  line-height: 1.4;
  cursor: pointer;
  transition:
    background var(--duration-short) var(--ease-enter),
    border-color var(--duration-short) var(--ease-enter);
}
.sel-item:hover,
.sel-item:focus-visible {
  background: var(--bg-active);
  border-color: var(--border-subtle);
}
.sel-item[aria-checked="true"] {
  border-color: var(--accent-green);
  background: var(--bg-active);
}

.sel-item[aria-disabled="true"] {
  background: color-mix(in srgb, var(--bg-elevated) 85%, transparent);
  color: var(--text-muted);
  cursor: not-allowed;
}
.sel-item[aria-disabled="true"]:hover {
  background: color-mix(in srgb, var(--bg-elevated) 85%, transparent);
  border-color: transparent;
}

.sel-badge {                               /* offline pill */
  padding: 0 var(--sp-xs);
  height: var(--chip-height);              /* reuse U8 chip-height token */
  display: inline-flex;
  align-items: center;
  border-radius: var(--radius-sm);
  background: color-mix(in srgb, var(--accent-amber) 12%, transparent);
  color: var(--accent-amber);
  border: 1px solid color-mix(in srgb, var(--accent-amber) 40%, transparent);
  font-family: var(--font-code);
  font-size: var(--text-label);
  font-weight: 500;
  letter-spacing: 0.05em;
  text-transform: lowercase;
}
```

**New tokens declared:** one, to be added to `frontend/src/style.css` `:root{}` block adjacent to the other layout tokens (around line 67):

- `--sel-tooltip-max: 48px;` — max-height target for the Policy row tooltip expand animation. 48px comfortably fits two wrapped lines at `var(--text-label)` with 1.4 line-height.

All other values derive from existing tokens. Raw `1px` borders match existing `style.css` convention (lines 8-9, 93, 182, 203 already use `1px solid` as the primitive unit).

### 6. Signature elements

Three chosen, each concretely present in this control:

1. **Monospace glyph prefix pattern** — every selector button leads with a deliberate BMP Unicode glyph (`⌂ ◆ ▶`) in `var(--font-code)`; the menu items inside continue the mono treatment. The combination reads as a "runtime pipeline address bar," not a generic dropdown, and doubles as an at-a-glance backend indicator when the button is rendered at 14 visible chars.

2. **Very short transition timings (60–120ms)** — the popover fly uses 120ms, state transitions hit `var(--duration-short)` (100ms), the optimistic-write pulse is 150ms. No transition exceeds 150ms — this is the crispness cue that separates a native desktop selector from a web dropdown that feels "cloudy."

3. **Inline muted-caps category caption inside the popover** — a single `.sel-caption` row above the list (`POLICY`, `BACKEND`, `MODEL`) in `var(--text-muted)` uppercase with `letter-spacing: 0.08em`. Cheap to render, carries the terminal-aesthetic signature already seen on the `.commit-step` prompt glyph, and ensures the three popovers feel like one family even when opened independently.
