# ui-ast-U8: "View raw" fallback toggle + diagnostics surface

**Status:** done
**Domain:** frontend
**Size:** S
**Depends on:** ui-ast-U6
**Priority:** P2-medium

## Story

As a Mashed user who sees a structured UI AST but suspects the adapter got something wrong, I want a "View raw" toggle that reveals the original Claude capture and a small diagnostics panel showing any `fallback_reasons` plus the `untrusted` flag, so that I can verify the translation is faithful and fall back to reading the raw output without leaving the modal.

## Description

Implements spec §6.3's "[ View raw ▾ ]" affordance and the §7.2 `diagnostics.untrusted` surface. Adds a collapsible panel at the bottom of `InputResponseModal.svelte` (above the Send button) that reveals `pendingPrompt.lastOutput` verbatim as preformatted text. Shows the diagnostic chip "Untrusted" when `ast.diagnostics.untrusted == true`, linking to the collapse chip that shows `fallback_reasons` when non-empty. Honours §11 Q7 UX preference: collapsed-by-default, with the expanded-by-default behaviour switchable via `mashedConfig.UIAdapterUntrustedExpanded` (shipped in U5).

### Scope summary

- `frontend/src/components/bmad/RawViewToggle.svelte` (new) — collapsible `<details>`-like panel with `lastOutput` as preformatted text.
- `frontend/src/components/bmad/DiagnosticsChip.svelte` (new) — two-chip display (Untrusted / N fallback_reasons) with a tooltip listing reasons verbatim.
- `frontend/src/components/bmad/InputResponseModal.svelte` — mount both above the Send button, conditional on `pendingAst` non-null + a diagnostics field being present.
- Playwright AC for toggle behaviour + config-driven default.

### Non-goals

- No changes to the validator (§3.3 — U2 shipped).
- No changes to the "Untrusted" detection logic (§4.7.4 / §7.2 — U2 shipped).
- No per-reason user education tooltips beyond the raw reason string.
- No logging of the toggle being used — purely UX.

## Developer Notes

### Files to create/modify

- `frontend/src/components/bmad/RawViewToggle.svelte` (new).
- `frontend/src/components/bmad/DiagnosticsChip.svelte` (new).
- `frontend/src/components/bmad/InputResponseModal.svelte` (modify — mount the two components).
- `frontend/src/components/bmad/__tests__/RawViewToggle.test.ts` (new).
- `frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts` (new).
- `tests/ac/ui-ast-view-raw.spec.ts` (new Playwright).

### `RawViewToggle.svelte`

```svelte
<script lang="ts">
  export let raw: string;
  export let expandedByDefault = false;  // from U5 UIAdapterUntrustedExpanded
  export let triggered = false;          // true when diagnostics flagged

  let open = expandedByDefault && triggered;
</script>

{#if raw}
  <details class="raw-view" bind:open>
    <summary>
      <span class="chevron" aria-hidden>{open ? '▾' : '▸'}</span>
      View raw
    </summary>
    <pre class="raw-content">{raw}</pre>
  </details>
{/if}
```

### `DiagnosticsChip.svelte`

```svelte
<script lang="ts">
  import type { Diagnostics } from '../../types/uiAst';
  export let diagnostics: Diagnostics | null;
</script>

{#if diagnostics?.untrusted || (diagnostics?.fallback_reasons?.length ?? 0) > 0}
  <div class="diagnostics" role="status">
    {#if diagnostics.untrusted}
      <span class="chip warn" title="The adapter may have dropped or altered content. See raw below.">
        Untrusted
      </span>
    {/if}
    {#if (diagnostics.fallback_reasons?.length ?? 0) > 0}
      <span class="chip info" title={diagnostics.fallback_reasons.join(', ')}>
        {diagnostics.fallback_reasons.length} adapter {diagnostics.fallback_reasons.length === 1 ? 'note' : 'notes'}
      </span>
    {/if}
  </div>
{/if}
```

### Modal integration

Mount above the Send button, under the AST node tree:

```svelte
<DiagnosticsChip diagnostics={$pendingAst?.diagnostics ?? null} />
<RawViewToggle
  raw={$pendingPrompt?.lastOutput ?? ''}
  triggered={$pendingAst?.diagnostics?.untrusted ?? false}
  expandedByDefault={uiAdapterUntrustedExpanded}
/>
```

`uiAdapterUntrustedExpanded` comes from the U5 `mashedConfig` field via the settings store (`uiAdapterSettings.ts`).

### §11 Q7 — resolved 2026-04-21 (user confirmed spec recommendation)

**Decision:** collapsed-by-default + tinted banner. U5 persists `UIAdapterUntrustedExpanded` (default `false`). U8 honours it:

- Config `false` (default): panel is collapsed even when `untrusted: true`. Tinted amber banner still renders to surface the `untrusted` signal.
- Config `true`: panel is expanded when `untrusted: true` on mount (instant, no animation per the design brief's "Q7 expanded-by-default motion" note).
- User clicks the chevron: override for the current modal session (no persistence — each new suspension resets).

### Risks / gotchas

- **`<details>` semantics.** Using native `<details>` gives free keyboard + a11y; do NOT re-roll this with custom JS.
- **Raw content escaping.** Rendered as `<pre>{raw}</pre>` — Svelte escapes by default. Verify no `{@html ...}` anywhere.
- **Missing `lastOutput`.** Legacy snapshots (pre-Layer 1) may have empty `lastOutput`. If empty, do NOT render the toggle at all. AC codifies.
- **Diagnostics absence.** When `pendingAst.diagnostics` is entirely absent or has no flags, `DiagnosticsChip` renders nothing. Both chips are conditional.
- **`fallback_reasons` list length.** Spec §3.3 allows accumulation — reasonable cap to keep the tooltip readable is 8 reasons. Show a "+N more" tail. Test the boundary.
- **Color tokens.** Chips use `warn` (amber) for Untrusted and `info` (neon green) for notes — mirror the `HintBanner` palette from U6.

### Reference files

- `frontend/src/components/bmad/InputResponseModal.svelte` — U6+U7 scaffolding to extend.
- `frontend/src/components/bmad/HintBanner.svelte` — palette precedent.
- `frontend/src/lib/stores/uiAdapterSettings.ts` (U5) — `uiAdapterUntrustedExpanded`.
- `docs/mashed-ui-ast-schema.md` §6.3 (View raw), §7.2 (untrusted), §11 Q7 (default).

## Acceptance Criteria

**AC-1: `RawViewToggle` collapses by default**
- Given `raw` is non-empty and `triggered` is false
- When the component renders
- Then the `<details>` element has no `open` attribute
- And the raw content is not visible
- Verified by `RawViewToggle.test.ts` → `collapses-by-default`

**AC-2: `RawViewToggle` opens when user clicks the summary**
- Given the collapsed toggle
- When the user clicks "View raw"
- Then the raw content becomes visible
- Verified by Playwright `ui-ast-view-raw.spec.ts` → `toggle-opens-on-click`

**AC-3: `RawViewToggle` is NOT rendered when `raw` is empty**
- Given `raw` is an empty string
- When the component renders
- Then the DOM contains no `<details>` element for the raw view
- Verified by `RawViewToggle.test.ts` → `hidden-when-empty`

**AC-4: `expandedByDefault` + `triggered` opens the panel on mount**
- Given `expandedByDefault = true` and `triggered = true`
- When the component mounts
- Then `<details open>` is present
- And given either flag is false, the panel is collapsed
- Verified by `RawViewToggle.test.ts` → `expanded-when-both-flags-true`

**AC-5: `DiagnosticsChip` shows "Untrusted" chip when flagged**
- Given `diagnostics.untrusted == true`
- When the component renders
- Then a chip with text "Untrusted" is visible
- And its tooltip (`title`) contains explanatory text
- Verified by `DiagnosticsChip.test.ts` → `untrusted-chip-visible`

**AC-6: `DiagnosticsChip` shows note count for `fallback_reasons`**
- Given `fallback_reasons = ["empty_options", "dup_key"]`
- When rendered
- Then a chip with text `"2 adapter notes"` is visible
- And its `title` attribute contains both reason strings
- And given one reason, the chip reads `"1 adapter note"` (singular)
- Verified by `DiagnosticsChip.test.ts` → `note-count-pluralisation`

**AC-7: `DiagnosticsChip` renders nothing when no flags**
- Given `diagnostics.untrusted == false` and `fallback_reasons == []`
- When rendered
- Then no chip is visible
- Verified by `DiagnosticsChip.test.ts` → `empty-diagnostics-silent`

**AC-8: §11 Q7 resolved — config `UIAdapterUntrustedExpanded: false` (default) keeps panel collapsed even on untrusted**
- Given the Settings config has `uiAdapterUntrustedExpanded: false`
- And the AST has `diagnostics.untrusted == true`
- When the modal opens
- Then `RawViewToggle` renders collapsed
- Verified by Playwright `ui-ast-view-raw.spec.ts` → `q7-default-collapsed`

**AC-9: §11 Q7 resolved — config `UIAdapterUntrustedExpanded: true` (user override) + `untrusted: true` opens the panel on mount**
- Given Settings has `uiAdapterUntrustedExpanded: true`
- And the AST has `diagnostics.untrusted == true`
- When the modal opens
- Then `RawViewToggle` renders with the panel expanded
- Verified by Playwright `ui-ast-view-raw.spec.ts` → `q7-expand-when-flagged`

**AC-10: Modal mounts both components only when adapter state is non-null**
- Given `pendingAst` is `null` (Layer-1 fallback)
- When the modal renders
- Then neither `DiagnosticsChip` nor `RawViewToggle` are in the DOM
- And the modal behaves exactly as pre-U8
- Verified by `InputResponseModal.test.ts` → `u8-components-absent-when-pendingAst-null`

**AC-11: Raw content is rendered safely (no HTML injection)**
- Given `raw` contains `"<script>alert(1)</script>"`
- When rendered
- Then the script tag appears as literal text inside `<pre>`
- And no `alert` fires
- Verified by `RawViewToggle.test.ts` → `raw-content-escaped` + Playwright assertion

## BDD Test Scenarios

```gherkin
Feature: View raw toggle + diagnostics surface

  Scenario: Raw panel collapsed by default
    Given raw is non-empty and triggered is false
    When the component renders
    Then the details element is not open

  Scenario: User clicking summary opens the panel
    Given the collapsed toggle
    When the user clicks "View raw"
    Then the raw content is visible

  Scenario: Raw panel is hidden when raw is empty
    Given raw is empty
    When rendered
    Then no details element exists

  Scenario: Both flags true open the panel on mount
    Given expandedByDefault true and triggered true
    When mounted
    Then the panel is open

  Scenario: Untrusted chip shown when diagnostics flagged
    Given diagnostics.untrusted true
    When rendered
    Then a chip with text "Untrusted" is visible

  Scenario: Adapter-note count pluralises correctly
    Given fallback_reasons with one entry
    When rendered
    Then the chip reads "1 adapter note"

  Scenario: Silent when no flags
    Given diagnostics empty
    When rendered
    Then no chip is visible

  Scenario: Q7 default config keeps panel collapsed even on untrusted
    Given uiAdapterUntrustedExpanded false and untrusted true
    When the modal opens
    Then the panel is collapsed

  Scenario: Q7 expand-by-default opens panel on untrusted
    Given uiAdapterUntrustedExpanded true and untrusted true
    When the modal opens
    Then the panel is expanded

  Scenario: Components absent when pendingAst is null
    Given pendingAst null
    When the modal renders
    Then neither RawViewToggle nor DiagnosticsChip are in the DOM

  Scenario: Raw content is HTML-escaped
    Given raw contains "<script>alert(1)</script>"
    When rendered
    Then the script tag is literal text inside pre
```

## Tasks / Subtasks

- [x] Task 1: `RawViewToggle` component (AC-1…AC-4, AC-11)
  - [x] RED: Vitest cases for collapse/open, empty-hide, both-flags, HTML escape
  - [x] GREEN: write `RawViewToggle.svelte`
- [x] Task 2: `DiagnosticsChip` component (AC-5, AC-6, AC-7)
  - [x] RED: Vitest cases for untrusted chip, note-count pluralisation, silent
  - [x] GREEN: write `DiagnosticsChip.svelte`
- [x] Task 3: Modal integration (AC-10)
  - [x] Mount both components conditionally on `pendingAst != null`
  - [x] RED: `InputResponseModal.test.ts` → pendingAst-null case
  - [x] GREEN: wire conditional mount
- [x] Task 4: Q7 config hook (AC-8, AC-9)
  - [x] Subscribe to `uiAdapterUntrustedExpanded` from `uiAdapterSettings` store
  - [x] Pass as `expandedByDefault` prop
  - [x] Playwright cases for both values
- [x] Task 5: Playwright AC suite
  - [x] Write `tests/ac/ui-ast-view-raw.spec.ts`
  - [x] Invoke `/playwright-cli` to validate

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ coverage on new Svelte components
- [x] Vitest passes
- [x] Playwright ACs pass via `/playwright-cli`
- [x] Frontend build passes
- [x] `/simplify` run on every modified file; no CRITICAL/HIGH findings
- [x] No `{@html ...}` in the raw rendering path (XSS guard)
- [x] Story status flipped to `done`
- [x] Changes committed on branch

## Design Brief

### 1. Layout composition

Both `DiagnosticsChip` and `RawViewToggle` mount inside `InputResponseModal.svelte`'s `.modal-body`, **below** the AST region / Layer-1 widget and **above** the footer. Stack (top-to-bottom), extending U6+U7:

1. `.modal-header` (unchanged).
2. `TranscriptPane` (unchanged).
3. `prompt-block` (unchanged).
4. (U7) Collapse banner, conditional.
5. `ast-region` (U6 + U7 `DecisionGroup` cards inside).
6. (U7) Layer-1 fallback widget, only when zero `decision_group` nodes.
7. **NEW `DiagnosticsChip` row** — flush-left, one-line strip of chips.
8. **NEW `RawViewToggle` block** — collapsible `<details>`, full width.
9. `.error-bar` (unchanged).
10. `.modal-footer` (unchanged / modified by U7).

- **`DiagnosticsChip` row:** `display: flex; align-items: center; gap: var(--sp-xs);`. Sits as a 24px-tall row with vertical padding `var(--sp-2xs)`. When there are no flags, the entire row does not render (no empty wrapper, no collapsed space).
- **`RawViewToggle` card:** native `<details>` element, full modal-body width. Summary row is always visible (when `raw` is non-empty); content below reveals on open. Summary + content together read like a single collapsible panel.
- **Rhythm:** both components inherit the `.modal-body` flex `gap: var(--sp-lg)` so they sit naturally `var(--sp-lg)` below the AST region. No extra margins — the parent gap does the work.
- **Banner coexistence rule:** when both `DiagnosticsChip` flags (untrusted + N notes) fire, chips render side-by-side in the same flex row. Never stack chips.

### 2. Typography plan

No new type roles beyond what U6 established. All sizes map to existing tokens.

| Role | Token | Font family | Weight |
|------|-------|-------------|--------|
| `DiagnosticsChip` "Untrusted" label | `var(--text-label)` 11px | `var(--font-mono)` | 600 uppercase, `letter-spacing: 0.05em` |
| `DiagnosticsChip` "N adapter notes" label | `var(--text-label)` 11px | `var(--font-mono)` | 500, `letter-spacing: 0.04em` (like `.round-pill`) |
| `RawViewToggle` summary ("View raw") | `var(--text-body)` 13px | `var(--font-mono)` | 500, `color: var(--text-dim)` |
| `RawViewToggle` chevron | `var(--text-body)` 13px | `var(--font-mono)` | 500, `color: var(--text-muted)` |
| `RawViewToggle` raw content `<pre>` | `var(--text-body)` 13px | `var(--font-code)` (JetBrains Mono) | 400, `line-height: 1.5` |

The chip label uses `var(--font-mono)` uppercase, matching the `HintBanner` tone-label convention from U6 — chips are meta-information about the AST, so they type as meta-information.

### 3. Color strategy

**Load-bearing rule: `DiagnosticsChip` and the `Untrusted` banner MUST reuse U6 `HintBanner` tones exactly. Do not invent new tints.**

- **`Untrusted` chip (warn tone):**
  - Background `color-mix(in srgb, var(--accent-amber) 15%, transparent)` (same 15% as `.round-pill` — chips are brighter than HintBanner bodies because they are compact and need to read at a glance).
  - Text `var(--accent-amber)`.
  - Border `1px solid color-mix(in srgb, var(--accent-amber) 30%, transparent)` (matches the `.glow-btn` border alpha formula).
  - Tooltip on hover, no explicit border-glow. The chip should read as **concern, not alarm** — amber + 15% alpha + no shadow = "heads up" not "ERROR".
- **`N adapter notes` chip (info tone):**
  - Background `color-mix(in srgb, var(--accent-blue) 12%, transparent)`.
  - Text `var(--accent-blue)` (#3d9eff).
  - Border `1px solid color-mix(in srgb, var(--accent-blue) 25%, transparent)`.
  - Tooltip (`title` attribute) enumerates reasons verbatim.
- **`RawViewToggle` summary (collapsed state):**
  - Background `transparent`.
  - Border `1px dashed var(--border-subtle)` — the dashed border signals "fallback / raw / diagnostic" without making the panel look broken. A solid border would feel like a peer-level card; dashed deliberately demotes it.
  - Text `var(--text-dim)`.
- **`RawViewToggle` summary (open state):**
  - Border becomes `1px solid var(--border-emphasis)` — closed = dashed hint, open = firm present.
  - Text `var(--text-primary)`.
- **`RawViewToggle` raw content `<pre>`:**
  - Background `var(--bg-deepest)` (same as `CodeBlock` from U6 — the deepest surface signals "this is raw capture").
  - Text `var(--text-primary)`.
  - Border: none (the summary's border wraps the whole `<details>`).
  - Maximum height `240px` with `overflow-y: auto` — long captures must not push the Send button off-screen.
- **Untrusted banner tint discipline — explicit rule:** the tint is amber at 10-15% alpha. It is NEVER red-tinted. Prompt-injection events will be common during dogfooding; red would read as "the system failed" and breed fatigue. Amber reads as "worth a look before trusting" — the calibrated tone for this UX moment.
- **Text contrast:**
  - Emphasised: open-state summary, raw content.
  - Default: collapsed-state summary, chip labels (the tint provides the emphasis).
  - Muted: chevron glyph.

### 4. Interaction model

- **Keyboard nav:**
  - `RawViewToggle` uses native `<details>`/`<summary>` — Space/Enter on the focused summary toggles open/close. Free a11y and keyboard support via the browser primitive.
  - Tab order position: after the last AST region interaction → `DiagnosticsChip` is NOT focusable (chips are read-only status, the tooltip is mouse+keyboard via native `title` attr, no explicit tabindex) → `RawViewToggle` summary → (existing) footer buttons. This keeps the expected keyboard reading: fill the form, check diagnostics, optionally open raw, then submit.
  - Escape closes the modal (existing handler). Escape on an open `<details>` does NOT close just the details — the modal is the atomic unit.
- **Hover states:**
  - `DiagnosticsChip`: no hover color change (status chips are not buttons). Tooltip fires via native `title` attribute delay.
  - `RawViewToggle` summary (collapsed): hover → `color: var(--text-primary)`, `border-color: var(--border-emphasis)`. Cursor `pointer`.
  - `RawViewToggle` summary (open): hover → no change (already emphasised).
- **Focus ring:** `<summary>` picks up the global `:focus-visible` rule from `style.css` (2px `var(--accent-green)` outline) automatically — do NOT override. Chips are not focusable, no ring.
- **Collapse/expand animation:** native `<details>` is instant by default (no animation). The story brief requests ~120ms with an easing token. Implementation options:
  - **Preferred:** JS-controlled transition — listen to the `toggle` event, animate `max-height` on the content `<pre>` via `transition: max-height var(--duration-medium) var(--ease-move)` (150ms = closest existing token to 120ms). The `<details>` attribute toggles immediately; visual content slides after.
  - **Alternative (if JS cost unjustified):** leave native instant-toggle. Acceptable because `<details>` semantics are more valuable than a 120ms slide.
  - Chevron rotation: `transform: rotate(90deg)` over `var(--duration-medium)` with `var(--ease-move)` — purely cosmetic, compatible with either approach.
- **Q7 expanded-by-default motion:** when `expandedByDefault && triggered` opens the panel on mount, skip the animation (instant-open). Animating in a pre-open panel looks like a bug. The `open` attribute is rendered directly in SSR/initial tick, not toggled.
- **Reduced-motion:** if `prefers-reduced-motion`, skip max-height transition and chevron rotation entirely. Rely on native instant toggle.

### 5. Component specs

- **`DiagnosticsChip` row wrapper:** `display: flex; align-items: center; gap: var(--sp-xs); padding: 0;`. `role="status"` per story spec.
- **Chip (both variants share shape — tone differs):**
  - `padding: 0 var(--sp-xs);` (0 vertical since font-size 11px + line-height controls height).
  - `border-radius: var(--radius-sm);` (2px — same as `.round-pill`).
  - `height: 18px; line-height: 18px;` — compact pill.
  - `border-width: 1px; border-style: solid;` — tone supplies the color.
  - `font-variant-numeric: tabular-nums;` on the notes chip so digit widths don't jitter between "1 note" and "12 notes".
  - `white-space: nowrap;`.
  - `display: inline-flex; align-items: center; gap: var(--sp-2xs);` — allows a tiny leading icon (Lucide `ShieldAlert` 11px for `Untrusted`, `Info` 11px for `notes`) without breaking the pill's vertical rhythm.
- **`RawViewToggle <details>`:**
  - `border: 1px dashed var(--border-subtle);` (collapsed) / `border: 1px solid var(--border-emphasis);` (open via `details[open]` selector).
  - `border-radius: var(--radius-md);` (4px).
  - `background: transparent;` (collapsed) / `background: var(--bg-surface);` (open).
  - `overflow: hidden;` so children respect the border-radius.
- **`RawViewToggle <summary>`:**
  - `padding: var(--sp-xs) var(--sp-sm);`.
  - `display: flex; align-items: center; gap: var(--sp-xs);`.
  - `cursor: pointer;`.
  - `user-select: none;`.
  - `list-style: none;` + `::-webkit-details-marker { display: none; }` to replace the default arrow with the custom chevron.
- **Chevron glyph:** 12x12 area containing `▸` / `▾` character (per story sketch), `color: var(--text-muted)`, `transition: transform var(--duration-medium) var(--ease-move);`.
- **`RawViewToggle <pre>`:**
  - `padding: var(--sp-sm) var(--sp-md);`.
  - `background: var(--bg-deepest);`.
  - `border-top: 1px solid var(--border-subtle);` (delimits summary from content when open).
  - `font-family: var(--font-code);`.
  - `font-size: var(--text-body);`.
  - `line-height: 1.5;`.
  - `white-space: pre;`.
  - `overflow: auto;`.
  - `max-height: 240px;` (large enough to see meaningful capture, small enough to keep Send button in view on default modal height).
- **Fallback-reasons overflow cap:** the `title` attribute joins reasons with `, ` (per story) — cap display at 8; if more, append "… +N more" to the tooltip string. Chip label still shows the total count numerically.

### 6. Signature elements

- **Dashed border on collapsed raw panel.** No other component in the app uses a dashed border. That is deliberate: dashed = fallback / diagnostic / optional, firm solid = present / required / actionable. A user can tell at a glance that `RawViewToggle` is a tool, not a decision.
- **Amber-not-red Untrusted chip.** The most important tone-calibration decision in the whole UI AST feature. Red would signal "broken", amber signals "verify". Prompt injection attempts are expected during dogfooding; they should not feel like failures.
- **Chips on a line, then a panel below.** The horizontal chip row + vertical panel block is a two-beat rhythm that matches the modal's existing structure: short metadata (round-pill row) → long content (prompt-block + widget). Diagnostics echoes that pattern one level down.
- **`<details>` native semantics.** Using the browser primitive instead of a Svelte-controlled dropdown is a discipline signature: mashed ships simple, keyboard-accessible, screen-reader-native UI where it can. Over-engineering the collapse would betray the "minimal decoration" DESIGN.md directive.
- **Mono-font chip labels.** Every piece of metadata in the modal (round-pill, chip labels, lang tags, column headers) uses `var(--font-mono)` uppercase with `letter-spacing`. This is the signature that tells the user "this is machine status, not prose".
- **Raw content on deepest surface.** `<pre>` sitting on `var(--bg-deepest)` — same trick as `CodeBlock` — makes the raw capture feel like a terminal pane nested inside the modal. Echoes the command-center aesthetic.

### Cross-story design coherence (binding rules)

- **`Untrusted` chip tint = U6 `HintBanner warn` tint, 15% alpha.** If U6's warn color shifts, `DiagnosticsChip` follows. No second amber token.
- **`N adapter notes` chip tint = U6 `HintBanner info` tint, 12% alpha.** Info is blue, not green — confirm U6 implements `info` as blue per DESIGN.md (not green as the story text incorrectly suggests).
- **`RawViewToggle <pre>` background = U6 `CodeBlock` background (`var(--bg-deepest)`).** Two read-only code surfaces — same surface color.
- **Collapse transition token = `var(--duration-medium)` (150ms) + `var(--ease-move)`.** Match any future "slide open" animations (e.g., U5's Custom… input reveal).
- **`uiAdapterUntrustedExpanded` config label (U5) must match the chip's `title` tooltip sense exactly.** Settings checkbox copy: "Expand 'View raw' by default when a translation is flagged as untrusted" → tooltip on chip: "The adapter may have dropped or altered content. See raw below." → `RawViewToggle` summary text: "View raw". Three touchpoints, one mental model.
- **No new tokens.** Every dimension, color, duration traces to `style.css`. If the dashed-border-vs-solid-border pattern needs variants elsewhere, extract a utility class before introducing a new token.

