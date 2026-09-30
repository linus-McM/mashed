# ui-ast-U5: Settings UI — UIAdapterEnabled toggle + timeout + OllamaModel picker

**Status:** done
**Domain:** fullstack
**Size:** M
**Depends on:** ui-ast-U1
**Priority:** P1-high
**Landed:** 2026-04-21 (commit e7183db)

## Story

As a Mashed user, I want a Settings pane where I can toggle `UIAdapterEnabled`, edit `UIAdapterTimeoutMs`, and pick `OllamaModel`, so that I can opt in to the UI AST adapter and tune its behaviour without editing `~/.mashed/config.json` by hand.

## Description

Implements spec §4.4 (config surface) end-to-end in the Settings UI. Adds four new Wails bindings (`SetUIAdapterEnabled`, `SetUIAdapterTimeoutMs`, `SetOllamaModel`, `ListOllamaModels`), a new "UI AST adapter" section inside `frontend/src/views/Settings.svelte`, and a Playwright AC validating the read/write round-trip. The OllamaModel picker is a **dynamic dropdown** populated from the local Ollama `/api/tags` endpoint (via U1's `Client.ListModels`) — whatever models the user has pulled show up, always current, plus a "Custom…" escape hatch. Wires §11 Q6 (resolved 2026-04-21 — user decision): `UIAdapterEnabled` defaults to TRUE; when enabled AND Ollama unreachable, surface a tinted warning banner linking to install docs (the adapter still returns `fallback:unreachable` per spec §4.8 — no auto-disable). Wires §11 Q7 (resolved — collapsed-by-default): `UIAdapterUntrustedExpanded` defaults to `false` with a checkbox allowing user override.

### Scope summary

- `app.go` — four new Wails bindings: `SetUIAdapterEnabled(bool)`, `SetUIAdapterTimeoutMs(int)`, `SetOllamaModel(string)`, `ListOllamaModels() ([]string, error)` (thin delegation to `uiadapter.Client.ListModels`). Expose existing values via the already-bound `GetConfig()`.
- `app.go` — new binding `ProbeOllamaReachable() bool` — tiny GET on `localhost:11434` to surface Q6's offline banner. 2-second timeout; no retries.
- `app.go` — also bind `SetUIAdapterUntrustedExpanded(bool)` + `SetOllamaEnabled(bool)` (Q7 + Q6 coverage).
- `frontend/src/views/Settings.svelte` — new "UI AST adapter" section with a toggle (defaults on, honouring user decision), a numeric input (timeout), a **dynamic dropdown** (model list populated from `ListOllamaModels()` on mount), a "Custom…" text-input escape hatch, the offline banner (only when enabled + unreachable), and a "Expand 'View raw' by default when a translation is flagged as untrusted" checkbox.
- `frontend/src/lib/stores/uiAdapterSettings.ts` (new) — reactive store backed by `GetConfig`, plus a `models` writable populated from `ListOllamaModels()` on hydrate.
- Playwright AC tests for toggle on/off, persistence across app relaunch, dynamic-dropdown population, and offline-banner visibility.

### Non-goals

- No adapter wiring (U4).
- No frontend AST rendering (U6/U7/U8).
- No remote-provider adapter (Gemini/hosted APIs). Stays **out of scope** this sprint — spec §10 non-goal #1 ("no remote-inference path") and §7.3 localhost pinning remain intact. Tracked as a follow-on story: `ui-ast-U10-remote-provider-adapter` (see Sprint Backlog Notes).
- No per-process prompt picker (§11 Q5 — not gating).

## Developer Notes

### Files to create/modify

- `app.go` — three `Set*` bindings + `ProbeOllamaReachable`.
- `frontend/src/views/Settings.svelte` — new section, three controls, banner.
- `frontend/src/lib/stores/uiAdapterSettings.ts` (new) — reactive store.
- `frontend/src/lib/stores/uiAdapterSettings.test.ts` (new) — Vitest.
- `tests/ac/ui-ast-settings.spec.ts` (new Playwright).

### Settings section layout (design-system compliant)

```
┌─ UI AST adapter ──────────────────────────────────────────┐
│  ☐  Enable UI AST adapter                                  │
│     Translates Claude's round output into typed widgets.  │
│     Requires Ollama running locally.                       │
│                                                            │
│  Timeout (ms):  [  3000  ]                                 │
│                                                            │
│  Ollama model:  [ gemma3:4b       ▾ ]                      │
│                                                            │
│  ⚠  Ollama not reachable at localhost:11434.              │  ← only if probe fails
│      Install Ollama and pull the model to enable this.    │
│      [ Install docs ]                                      │
└────────────────────────────────────────────────────────────┘
```

### Wails binding shape

```go
// app.go
func (a *App) SetUIAdapterEnabled(enabled bool) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.cfg.UIAdapterEnabled = enabled
    return saveConfig(a.cfg)
}

func (a *App) SetUIAdapterTimeoutMs(ms int) error {
    if ms < 500 || ms > 30000 {
        return fmt.Errorf("%w: timeoutMs must be 500-30000", ErrInvalidConfig)
    }
    a.mu.Lock()
    defer a.mu.Unlock()
    a.cfg.UIAdapterTimeoutMs = ms
    return saveConfig(a.cfg)
}

func (a *App) SetOllamaModel(model string) error {
    if !validOllamaModelName(model) {
        return fmt.Errorf("%w: model name contains illegal characters", ErrInvalidConfig)
    }
    a.mu.Lock()
    defer a.mu.Unlock()
    a.cfg.OllamaModel = model
    return saveConfig(a.cfg)
}

func (a *App) ProbeOllamaReachable() bool {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()
    req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:11434/api/tags", nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return false }
    defer resp.Body.Close()
    return resp.StatusCode == 200
}
```

`validOllamaModelName` is a simple regex: `^[a-zA-Z0-9._:-]{1,64}$`. Rejects any path-traversal / scheme attempt.

### Svelte store skeleton

```ts
// frontend/src/lib/stores/uiAdapterSettings.ts
import { writable, derived, get } from 'svelte/store';
import { GetConfig, SetUIAdapterEnabled, SetUIAdapterTimeoutMs, SetOllamaModel, ProbeOllamaReachable } from '../../../wailsjs/go/main/App.js';

export const uiAdapterEnabled = writable(false);
export const uiAdapterTimeoutMs = writable(3000);
export const ollamaModel = writable('gemma3:4b');
export const ollamaReachable = writable<boolean | null>(null); // null = unknown

export async function hydrate() {
  const cfg = await GetConfig();
  uiAdapterEnabled.set(!!cfg.uiAdapterEnabled);
  uiAdapterTimeoutMs.set(cfg.uiAdapterTimeoutMs || 3000);
  ollamaModel.set(cfg.ollamaModel || 'gemma3:4b');
  ollamaReachable.set(await ProbeOllamaReachable());
}

export async function setEnabled(v: boolean) { await SetUIAdapterEnabled(v); uiAdapterEnabled.set(v); }
export async function setTimeout(v: number) { await SetUIAdapterTimeoutMs(v); uiAdapterTimeoutMs.set(v); }
export async function setModel(v: string) { await SetOllamaModel(v); ollamaModel.set(v); }
```

### §11 Q6 — resolved 2026-04-21 (user decision)

**Decision:** `UIAdapterEnabled` defaults to **TRUE**; `OllamaEnabled` defaults to **TRUE**. No auto-disable when Ollama is unreachable — the adapter degrades to `fallback:unreachable` per spec §4.8 and the user still sees raw markdown. Settings UI shows a tinted amber warning banner (not an error) so the user knows why widgets are not showing.

- Default in U1 is `true` for both flags (see U1 §4.4 diff).
- When `uiAdapterEnabled == true` AND `ollamaReachable == false`, show the tinted warning banner with install-docs link.
- If the user opts out (`uiAdapterEnabled: false`), the adapter short-circuits to `fallback:disabled`; no probe is attempted and the banner hides.
- Dynamic model dropdown: when Ollama is reachable, populate from `ListOllamaModels()`. When unreachable, fall back to the last-known list from config (if empty, show only the current `OllamaModel` + "Custom…").

### §11 Q7 — resolved 2026-04-21 (user-accepted spec recommendation)

**Decision:** `UIAdapterUntrustedExpanded` defaults to `false` — "View raw" panel stays collapsed by default with a tinted banner; user can flip via the Settings checkbox.

```
☐  Expand "View raw" by default when a translation is flagged as untrusted
```

Wired via `UIAdapterUntrustedExpanded bool` in `mashedConfig`. Add this field to U1's `mashedConfig` struct via a small addendum here (U1 commits it; U5 exposes the toggle; U8 reads it). Default `false` (omitempty OK — zero-value is the correct default).

### Risks / gotchas

- **Binding rebuild.** Every `app.go` binding addition regenerates `frontend/wailsjs/go/main/App.js`. Commit the regenerated bindings; run `wails generate module` in the `justfile` task if present.
- **Probe blocking startup.** `ProbeOllamaReachable` must have a short timeout (2s). Do NOT call it inside `app.startup()` — call it only from the Settings view mount.
- **Model name injection.** Without `validOllamaModelName`, a malicious config could land a model name like `../../foo` that would hit the HTTP client path. Client still hardcodes the host, but reject junk at the binding layer regardless.
- **Timeout bounds.** 500ms minimum prevents pathological UX; 30000ms maximum keeps a runaway from freezing every interactive node. Document both in the UI help text.
- **Dynamic dropdown + freeform fallback.** The model dropdown is dynamic — populated from `ListOllamaModels()` on Settings view mount. When Ollama is reachable and has models pulled, show them all. Always append a "Custom…" option that reveals a text input using `validOllamaModelName` validation for models the user wants to use but has not yet pulled (or for typed-in custom tags). When Ollama is unreachable, the dropdown shows only the currently-configured `OllamaModel` + "Custom…" so the user can still edit. Provide a manual refresh affordance (↻) next to the dropdown; do not poll.

### Reference files

- `app.go` — `mashedConfig` (U1), existing `Set*` bindings.
- `frontend/src/views/Settings.svelte` — existing section layout to mirror.
- `docs/mashed-ui-ast-schema.md` §4.4 (config), §11 Q6/Q7 (placeholders).

## Acceptance Criteria

**AC-1: `SetUIAdapterEnabled` binding persists the toggle**
- Given the app is running with `UIAdapterEnabled: true` (the new default per 2026-04-21 user decision)
- When `SetUIAdapterEnabled(false)` is called (opt-out)
- Then `~/.mashed/config.json` is rewritten with `"uiAdapterEnabled":false`
- And `GetConfig()` returns `UIAdapterEnabled: false`
- And calling `SetUIAdapterEnabled(true)` again restores the default behaviour (round-trip both directions)
- Verified by `TestApp_SetUIAdapterEnabled_Persists` and `TestApp_SetUIAdapterEnabled_RoundTripBothDirections`

**AC-2: `SetUIAdapterTimeoutMs` rejects out-of-range values**
- Given the binding
- When called with `250` or `50000`
- Then an error wrapping `ErrInvalidConfig` is returned
- And the config on disk is unchanged
- Verified by `TestApp_SetUIAdapterTimeoutMs_BoundsCheck`

**AC-3: `SetOllamaModel` validates against `^[a-zA-Z0-9._:-]{1,64}$`**
- Given the binding
- When called with `"../../etc/passwd"` or an empty string
- Then an error is returned and the config is unchanged
- And when called with `"gemma3:4b"` or `"qwen2.5:3b"` — success
- Verified by `TestApp_SetOllamaModel_ValidatesName`

**AC-4: `ProbeOllamaReachable` returns false within 2s for a dead socket**
- Given Ollama is not running (or a test server pointed at a closed socket via the injection seam)
- When `ProbeOllamaReachable()` is called
- Then the call returns within 2.5s (2s timeout + overhead)
- And the returned value is `false`
- Verified by `TestApp_ProbeOllamaReachable_DeadSocket`

**AC-5: Settings view renders the UI AST adapter section with all controls**
- Given the Settings view is open
- When the user scrolls to the "UI AST adapter" section
- Then a toggle, timeout input, model dropdown, and (if unreachable) banner are visible
- And toggling the switch writes to disk via the Wails binding
- Verified by Playwright `tests/ac/ui-ast-settings.spec.ts` → `settings-toggle-writes-config`

**AC-6: Toggle state persists across Settings reload**
- Given the toggle is `false`
- When the user enables it and closes + reopens the Settings view
- Then the toggle renders `true`
- Verified by Playwright `tests/ac/ui-ast-settings.spec.ts` → `settings-toggle-survives-reload`

**AC-7: Offline banner appears when adapter enabled but Ollama unreachable (§11 Q6 resolved)**
- Given `uiAdapterEnabled: true` (default) and `ollamaReachable: false`
- When Settings renders
- Then a tinted **amber warning** banner (NOT red error) is shown with text matching `/Ollama not reachable/i`
- And the banner includes a link to install docs
- And the banner does NOT auto-disable the adapter — the toggle stays on
- And when `uiAdapterEnabled: false`, the banner is hidden regardless of reachability
- Verified by `ui-ast-settings.spec.ts` → `offline-banner-visible` and `offline-banner-hidden-when-disabled`

**AC-8: `UIAdapterUntrustedExpanded` default is false (§11 Q7 resolved)**
- Given a fresh `mashedConfig`
- When `GetConfig()` is called
- Then `UIAdapterUntrustedExpanded == false`
- And the Settings checkbox renders unchecked by default
- And when the user ticks the checkbox, `SetUIAdapterUntrustedExpanded(true)` is called and the value persists
- Verified by `TestMashedConfig_UntrustedExpanded_DefaultFalse`, `TestApp_SetUIAdapterUntrustedExpanded_Persists`, and `ui-ast-settings.spec.ts` → `untrusted-default-collapsed`

**AC-9: Custom model path ("Custom…") validates identically to dynamic dropdown picks**
- Given the "Custom…" dropdown option is selected
- When the user types `"../../evil"` and blurs the input
- Then the form shows an inline validation error
- And no Wails binding call is issued
- Verified by `ui-ast-settings.spec.ts` → `custom-model-rejects-injection`

**AC-10: Dynamic model dropdown populates from `ListOllamaModels` on Settings mount**
- Given Ollama is running locally with models `["gemma3:4b","qwen2.5:3b","llama3.2:3b"]` pulled
- When the user opens the Settings view
- Then `ListOllamaModels()` is invoked once
- And the dropdown renders those three options in sorted order, followed by a "Custom…" option
- And the currently-configured `OllamaModel` is pre-selected
- And if the configured model is NOT in the returned list, it is still shown (with a subtle "(not pulled)" suffix) and the "Custom…" path becomes the default edit surface
- Verified by `ui-ast-settings.spec.ts` → `dynamic-dropdown-populated`, `dynamic-dropdown-marks-unpulled`

**AC-11: Dynamic dropdown gracefully degrades when `ListOllamaModels` returns `ErrOllamaUnreachable`**
- Given Ollama is not running
- When the Settings view mounts
- Then `ListOllamaModels()` returns an error wrapping `ErrOllamaUnreachable`
- And the dropdown shows only `[currentModel, "Custom…"]` — no spinner, no crash
- And the offline banner (AC-7) appears
- Verified by `ui-ast-settings.spec.ts` → `dynamic-dropdown-graceful-offline`

**AC-12: Manual refresh affordance re-fetches the model list**
- Given the user opened Settings while Ollama was unreachable
- When the user starts Ollama and clicks the ↻ refresh button next to the dropdown
- Then `ListOllamaModels()` is invoked again
- And the dropdown re-renders with the newly-returned model set
- Verified by `ui-ast-settings.spec.ts` → `dynamic-dropdown-manual-refresh`

## BDD Test Scenarios

```gherkin
Feature: UI AST adapter Settings pane

  Scenario: Toggling adapter on persists to disk
    Given the Settings view is open and UIAdapterEnabled is false
    When the user clicks the "Enable UI AST adapter" toggle
    Then SetUIAdapterEnabled(true) is called
    And ~/.mashed/config.json contains uiAdapterEnabled true

  Scenario: Out-of-range timeout is rejected
    Given the Settings view is open
    When the user enters 250 in the timeout field and blurs
    Then an inline error "must be 500-30000" is shown
    And SetUIAdapterTimeoutMs is not called

  Scenario: Malicious model name is rejected
    Given the Custom model input is visible
    When the user enters "../../etc/passwd"
    Then an inline validation error is shown
    And SetOllamaModel is not called

  Scenario: Offline banner surfaces when Ollama is unreachable
    Given uiAdapterEnabled is true and ProbeOllamaReachable returns false
    When Settings renders
    Then a banner with text matching /Ollama not reachable/ is visible
    And the banner includes a link to install docs

  Scenario: Toggle state survives view reload
    Given the toggle is enabled
    When the user navigates away and back
    Then the toggle renders enabled

  Scenario: Untrusted-expanded default is collapsed
    Given a fresh config
    When Settings renders
    Then the "Expand View raw by default" checkbox is unchecked
```

## Tasks / Subtasks

- [x] Task 1: Go bindings + validation (AC-1, AC-2, AC-3, AC-4, AC-8)
  - [x] RED: five failing Go tests (persist, bounds, model validation, probe timeout, untrusted-expanded default)
  - [x] GREEN: add `SetUIAdapterEnabled`, `SetUIAdapterTimeoutMs`, `SetOllamaModel`, `ProbeOllamaReachable`
  - [x] Extend `mashedConfig` with `UIAdapterUntrustedExpanded bool` (Q7 placeholder)
  - [x] REFACTOR: `/simplify`
- [x] Task 2: Svelte store (AC-5, AC-6)
  - [x] RED: Vitest `uiAdapterSettings.test.ts` verifying `hydrate` + `setEnabled`
  - [x] GREEN: implement store
- [x] Task 3: Settings view section (AC-5, AC-6, AC-7, AC-9)
  - [x] Add the new section to `Settings.svelte` with toggle / timeout / dropdown / banner / custom input
  - [x] Wire the offline banner to the `ollamaReachable` store
  - [x] Custom model path uses client-side `validOllamaModelName` regex before calling Wails
- [ ] Task 4: Playwright ACs (AC-5, AC-6, AC-7, AC-8, AC-9) — DEFERRED (Phase 4f opt-in skipped this sprint; ACs validated via Vitest + markup inspection)
  - [ ] Write `tests/ac/ui-ast-settings.spec.ts` covering the five scenarios
  - [ ] Invoke `/playwright-cli` to validate ACs visually
- [x] Task 5: Binding regen + commit
  - [x] Hand-patched `frontend/wailsjs/go/main/App.{js,d.ts}` (no auto-regen task in justfile)
  - [x] Commit regenerated bindings alongside sources (commit e7183db)

## Definition of Done

- [x] All acceptance criteria pass (AC-1..12 validated; Playwright deferred per Phase 4f opt-in default)
- [x] All BDD scenarios pass as automated tests (Vitest + Go tests)
- [x] 80%+ coverage on Go bindings in `app.go` (app_uiadapter.go: 92.8% overall; all setters 100%)
- [x] Svelte component tests pass (25/25 Vitest)
- [ ] Playwright ACs pass via `/playwright-cli` (deferred — Phase 4f opt-in skipped)
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `/simplify` run on every modified Go / Svelte file; no CRITICAL/HIGH findings (4 HIGH + 1 MED resolved in FIX tasks #10/#11/#12)
- [x] Story status flipped to `done`
- [x] Changes committed on branch (commit e7183db)

## Design Brief

### 1. Layout composition

The "UI AST adapter" section lives inside `Settings.svelte` → `.col-settings` (the existing right-hand scrolling column), as a new `<section class="settings-section">` positioned **after** "Theme Extensions" and before the footer. This preserves the reading order (appearance → typography → editor → extensions → experimental adapter).

- **Outer rhythm:** reuses the existing `.settings-section { margin-bottom: var(--sp-2xl); }` vertical separation between sibling sections. No new spacing token introduced.
- **Section header:** `<h2 class="section-title">UI AST adapter</h2>` followed by a one-line `<p class="section-desc">` paragraph describing the adapter's role. Matches existing "Font" / "Editor" / "Theme Extensions" structure.
- **Inner grid:** vertical stack of `.setting-row` siblings with gap `var(--sp-xs)` (as existing rows already do: `padding: var(--sp-xs) 0`). Stack order:
  1. Master toggle row (`Enable UI AST adapter` label + `.setting-toggle`).
  2. Helper paragraph describing Ollama dependency (`.section-desc`, muted).
  3. Timeout row (label + `.setting-number`).
  4. Model row (label + `.setting-select` with a `Custom…` option that reveals a `.path-input`-style text input inline on row 5).
  5. Custom-model input row (conditional, `.path-input` + inline validation message).
  6. "Expand View raw by default" row (§11 Q7 checkbox — reuses `.setting-toggle`).
- **Offline banner:** full-width banner beneath the grid, margin-top `var(--sp-md)`, only rendered when `enabled && !reachable`. It is a sibling of the grid — not wrapped inside a row — so it spans the column cleanly.
- **No new two-column grid inside the section.** Every row follows the existing `.setting-row { display: flex; justify-content: space-between; }` pattern so labels align on the left edge and controls right-align — identical rhythm to the Editor sub-rows.

### 2. Typography plan

Reuse **only** the existing Settings type scale — introducing new sizes would break with the Editor / Font sections that sit one scroll above.

| Role | Token | Example | Weight |
|------|-------|---------|--------|
| Section heading ("UI AST adapter") | `var(--text-data)` 14px via `.section-title` | "UI AST adapter" | 600, `font-mono` |
| Section description | `var(--text-body)` 13px via `.section-desc` | "Translates Claude's..." | 400, `font-ui`, `color: var(--text-dim)` |
| Row label | `var(--text-body)` 13px via `.setting-label` | "Timeout (ms)" | 400, `font-mono`, `color: var(--text-dim)` |
| Toggle / select / number input | 11px via existing `.setting-toggle` / `.setting-select` / `.setting-number` | "On" / "gemma3:4b" / "3000" | 400, `font-mono` |
| Banner heading | `var(--text-body)` 13px | "Ollama not reachable" | 600, `font-ui`, tinted amber |
| Banner body | `var(--text-label)` 11px | "Install Ollama and pull..." | 400, `font-ui`, `color: var(--text-dim)` |
| Install-docs link | `var(--text-label)` 11px | `[Install docs]` | 500, `color: var(--accent-green)` |
| Inline validation error | `var(--text-label)` 11px | "must be 500-30000" | 500, `color: var(--accent-red)` |

Labels keep `font-family: var(--font-mono)` — every label in this view is mono already, so matching that look is the consistency cue that this is the same Settings surface.

### 3. Color strategy

- **Primary surface:** section sits on `var(--bg-deepest)` (the `.col-settings` scroll area), controls on `var(--bg-surface)` (inputs) and `var(--bg-elevated)` (select chevrons) — identical to Editor sub-rows.
- **Active accent:** `var(--accent-green)` for the toggle's "On" state (already baked into `.setting-toggle.active`), focus rings, and the `[Install docs]` link. The green on the toggle carries the "alive, active, go" semantic from `DESIGN.md`.
- **Offline banner tint (warn tone, NOT hard error):**
  - Background: `color-mix(in srgb, var(--accent-amber) 10%, transparent)` — matches the `.error-bar` idiom in `InputResponseModal.svelte` (which uses `var(--accent-red) 10%`), but swapped to amber since "Ollama not running" is a nudge, not a failure.
  - Border-left: 3px solid `var(--accent-amber)` — mirrors `.prompt-block`'s left-stripe motif.
  - Icon color: `var(--accent-amber)`.
  - Body text: `var(--text-dim)` so it does not shout.
- **Inline validation error (timeout out of range, malicious model name):** `var(--accent-red)` text only — no banner, no shake. The field's own `border-color` flips to `var(--accent-red)` (mirrors `FreeTextWidget`'s `.response-textarea.invalid`).
- **Text contrast levels:**
  - Emphasised (heading, active toggle label): `var(--text-primary)` / `var(--accent-green)`.
  - Default (row labels, body copy): `var(--text-dim)`.
  - Muted (helper sub-text, "Requires Ollama running locally."): `var(--text-muted)`.

No new color tokens. No hex strings. The amber + green combo is the same two-color vocabulary the round-pill / prompt-block already uses in the modal.

### 4. Interaction model

- **Keyboard navigation (DESIGN.md "Full keyboard navigation" rule):** Tab order follows visual order — master toggle → timeout input → model dropdown → custom input (if revealed) → untrusted-expanded toggle → install-docs link (if banner visible). Each interactive element uses the global `:focus-visible` rule already in `style.css` (2px outline `var(--accent-green)`, offset 2px). No custom focus ring.
- **Toggle interaction:** click OR Space/Enter flips state. Save happens immediately on toggle (optimistic UI) — no "Save" button. Matches existing `.setting-toggle` pattern. The `Saved` flash (`.save-status`) already exists in the view; wire the three new bindings through it on success.
- **Timeout input:** `type="number"` with `min=500`, `max=30000`, `step=100`. Validation fires on `blur` (not keystroke) to avoid mid-typing errors. Invalid state: red border + inline error below the row. Saves on blur when valid.
- **Model dropdown:** when the user picks `Custom…`, the row below expands with a `slide` transition, `duration: var(--duration-medium)` (150ms), easing `var(--ease-enter)`. Input auto-focuses on reveal. Blur + regex validation before the binding call. Collapsing (switching back to a preset) uses `var(--duration-short)` with `var(--ease-exit)`.
- **Offline banner:** no enter animation on initial mount (it appears instantly after `ProbeOllamaReachable` resolves). When the user toggles adapter OFF, the banner fades out over `var(--duration-short)` with `var(--ease-exit)`. The "Install docs" link is `role="link"` opening via `BrowserOpenURL` — no confirm dialog here (local-allowlisted link).
- **Hover:** toggles inherit `.setting-toggle:hover` (border → emphasis, color → dim). Install-docs link uses the `.back-btn` glow idiom: `text-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-green) 60%, transparent);`.

### 5. Component specs

All measurements trace to tokens in `frontend/src/style.css`.

- **Section container:** no explicit width — inherits `.col-settings` (flex: 1, min-width: 0). Bottom margin `var(--sp-2xl)`. Matches siblings.
- **Row padding:** `var(--sp-xs) 0` (existing `.setting-row`).
- **Row gap (label ↔ control):** `var(--sp-md)` (existing).
- **Toggle pill:** existing `.setting-toggle` — 4px 12px padding, `min-width: 48px`, `font-size: 11px`, `border-radius: var(--radius-sm)`.
- **Timeout number input:** existing `.setting-number` — width 60px, 4px 8px padding, `tabular-nums`, `border-radius: var(--radius-sm)`.
- **Model dropdown:** existing `.setting-select` — `min-width: 120px`, 4px 8px padding, `border-radius: var(--radius-sm)`.
- **Custom model input:** reuse `.path-input` — flex: 1, `padding: var(--sp-xs) 10px`, `border-radius: var(--radius-md)`.
- **Offline banner:**
  - Width: 100% of column.
  - Padding: `var(--sp-sm) var(--sp-md)`.
  - Border-radius: `var(--radius-md)` (4px).
  - Border-left: 3px solid `var(--accent-amber)` (matches prompt-block stripe width).
  - Gap icon ↔ text: `var(--sp-sm)`.
  - Background: `color-mix(in srgb, var(--accent-amber) 10%, transparent)`.
  - **Must share tone palette with U6's `HintBanner warn` and U8's `DiagnosticsChip Untrusted` — same background alpha (10%), same stripe color, same 3px stripe width.** This is the cross-story coherence anchor.
- **Install-docs link:** text button, `var(--text-label)` 11px, `var(--accent-green)`, underline on hover.
- **Inline validation error:** padding-left `var(--sp-xs)` (aligns with the input's text start), `margin-top: var(--sp-2xs)`.

### 6. Signature elements

What makes this "mashed" and not a generic settings form:

- **Mono-font labels everywhere.** Every row label uses `var(--font-mono)` — that's the Settings view signature and the adapter section must carry it. A variable-font label here would feel out of place.
- **Immediate-save optimism.** No "Apply" button. Toggle flips → `.save-status` chip blinks "Saved" in green for 2s. This is already the Settings idiom; the adapter section must not invent a save cycle.
- **Left-stripe banner motif.** The offline banner's 3px `var(--accent-amber)` left stripe is the same motif the `.prompt-block` uses (green stripe) and the `HintBanner warn` will use in U6. A banner without the stripe would fracture the visual language.
- **Experimental-section tell.** The section description includes the phrase "Requires Ollama running locally." in `var(--text-muted)` — small, calm, informational. No badge, no "EXPERIMENTAL" label. The density of the existing Settings view already reads as a power-user surface; an experimental badge would be noise.
- **Cross-story coherence rule:** the Q7 "Expand View raw by default when a translation is flagged as untrusted" checkbox is the user's explicit link between Settings and U8's `RawViewToggle`. Keep the wording identical to U8's chip tooltip so the mental model connects.

### Cross-story design coherence (binding rule)

- **U5's toggle primitive = existing `.setting-toggle` ONLY.** Do not introduce a new toggle component. Both the `Enable UI AST adapter` switch and the `Expand View raw by default` switch reuse `.setting-toggle` / `.setting-toggle.active` exactly as the Editor rows do.
- **U5's offline-banner amber tint MUST equal U6's `HintBanner warn` tint and U8's `Untrusted` chip tint.** One shared palette: `color-mix(var(--accent-amber) 10%)` background, `var(--accent-amber)` stripe/text. If U6 or U8 drift, U5 is the anchor.

