# ui-ast-U6: `AstNode` dispatcher + node components + `pendingAst` derived store

**Status:** ready
**Domain:** frontend
**Size:** L
**Depends on:** ui-ast-U4
**Priority:** P0-critical

## Story

As a Mashed user looking at an interactive BMAD turn, I want the round's captured output to render as typed widgets (markdown prose, hint banners, summary cards, code blocks, comparison tables) instead of a single wall of text, so that the structure Claude produced is visible at a glance without reading the tmux pane.

## Description

Implements spec §6.1 (`pendingAst` derived store + wire-format `structured` field) and §6.2 (`AstNode` recursive dispatcher with six passive node components — `markdown`, `hint`, `summary`, `code`, `table`, plus the unknown-type graceful fallback). `decision_group` is NOT rendered in U6 — that's U7. U6 renders every AST as prose + banners + summaries, with `DecisionGroup` and the response-map collection deferred. This lets U6 ship behind the `UIAdapterEnabled=false` flag (default) and degrade cleanly: when `structured` is present but U7 is not yet deployed, the user sees the AST's passive nodes and the Layer-1 single-widget fallback underneath.

### Scope summary

- `frontend/src/stores/interactiveInput.ts` — extend `PendingPrompt` with `structured?: string`; add `pendingAst` derived store per spec §6.1.
- `frontend/src/components/bmad/AstNode.svelte` (new) — recursive dispatcher per §6.2.
- `frontend/src/components/bmad/MarkdownBlock.svelte` (new) — markdown renderer with the §7.2 link sanitisation.
- `frontend/src/components/bmad/HintBanner.svelte` (new) — tone-colored banner.
- `frontend/src/components/bmad/SummaryCard.svelte` (new) — bullet summary.
- `frontend/src/components/bmad/CodeBlock.svelte` (new) — syntax-highlighted block + copy button.
- `frontend/src/components/bmad/ComparisonTable.svelte` (new) — HTML table with zebra rows.
- `frontend/src/components/bmad/InputResponseModal.svelte` — subscribe to `pendingAst`; when non-null, render the `AstNode` tree above the Layer-1 fallback widget; when null, keep Layer-1 UX unchanged.
- `frontend/src/components/bmad/__tests__/*.test.ts` — per-component Vitest tests.
- `tests/ac/ui-ast-rendering.spec.ts` — Playwright ACs for the six node shapes.

### Non-goals

- No `DecisionGroup` component (U7).
- No response-map collection (U7).
- No "View raw" toggle (U8).
- No JSON vs collapse submit path (U7).
- No Monaco bundle changes.

## Developer Notes

### Files to create/modify

Per spec §6.2 directory convention, dispatcher + node components sit under `frontend/src/components/bmad/`:

- `frontend/src/stores/interactiveInput.ts` (modify) — add `structured` to `PendingPrompt`; add `pendingAst` derived store.
- `frontend/src/types/uiAst.ts` (new) — TypeScript mirror of the UIAST schema (version 1) from spec §3.
- `frontend/src/components/bmad/AstNode.svelte` (new).
- `frontend/src/components/bmad/MarkdownBlock.svelte` (new).
- `frontend/src/components/bmad/HintBanner.svelte` (new).
- `frontend/src/components/bmad/SummaryCard.svelte` (new).
- `frontend/src/components/bmad/CodeBlock.svelte` (new).
- `frontend/src/components/bmad/ComparisonTable.svelte` (new).
- `frontend/src/components/bmad/InputResponseModal.svelte` (modify — subscribe + render AST nodes above the Layer-1 widget).

### `pendingAst` derived store (spec §6.1 verbatim)

```ts
// frontend/src/stores/interactiveInput.ts
export interface PendingPrompt {
  // existing fields…
  lastOutput?: string;
  structured?: string;          // raw JSON string; absent on legacy snapshots (§5.1)
}

import { derived } from 'svelte/store';
import type { UIAST } from '../types/uiAst';

export const pendingAst = derived(pendingPrompt, ($p, set) => {
  if (!$p?.structured) { set(null); return; }
  try {
    const ast = JSON.parse($p.structured) as UIAST;
    if (ast?.version !== '1') { set(null); return; } // unknown version → fallback
    set(ast);
  } catch {
    set(null);                  // malformed → Layer-1 fallback
  }
});
```

### `AstNode.svelte` dispatcher (spec §6.2 verbatim)

```svelte
<script lang="ts">
  import type { UINode } from '../../types/uiAst';
  import MarkdownBlock from './MarkdownBlock.svelte';
  import HintBanner from './HintBanner.svelte';
  import SummaryCard from './SummaryCard.svelte';
  import CodeBlock from './CodeBlock.svelte';
  import ComparisonTable from './ComparisonTable.svelte';
  // DecisionGroup imported in U7

  export let node: UINode;
</script>

{#if node.type === 'markdown'}
  <MarkdownBlock content={node.content ?? ''} />
{:else if node.type === 'hint'}
  <HintBanner tone={node.tone ?? 'info'}>{node.content ?? ''}</HintBanner>
{:else if node.type === 'summary'}
  <SummaryCard heading={node.heading ?? ''} bullets={node.bullets ?? []} />
{:else if node.type === 'code'}
  <CodeBlock lang={node.lang ?? ''} content={node.content ?? ''} copyable={node.copyable ?? false} />
{:else if node.type === 'table'}
  <ComparisonTable heading={node.heading ?? ''} columns={node.columns ?? []} rows={node.rows ?? []} />
{:else if node.type === 'decision_group'}
  <!-- U7 renders DecisionGroup. Until then, graceful markdown fallback. -->
  <MarkdownBlock content={node.prompt ?? node.heading ?? JSON.stringify(node)} />
{:else}
  <!-- Unknown type → graceful markdown fallback -->
  <MarkdownBlock content={node.content ?? JSON.stringify(node)} />
{/if}
```

### §7.2 markdown link sanitisation (load-bearing security AC)

`MarkdownBlock.svelte` MUST:
1. Disable auto-linking in the markdown renderer (markdown-it `linkify: false` or remark equivalent).
2. On every `<a>` click — route through a confirm dialog showing the fully-resolved URL before opening via the system browser (use `BrowserOpenURL` from the Wails runtime).
3. At render time, strip `href` for any link whose URL uses a rejected scheme: `javascript:`, `data:`, `file:`. The link degrades to plain text.
4. Apply the same sanitisation to any URL-looking string emitted inside `HintBanner`, `SummaryCard`, and `ComparisonTable` content (belt-and-braces — the validator already flags missing URLs, but the frontend must still refuse to render rejected schemes).

Implementation sketch:

```svelte
<!-- MarkdownBlock.svelte -->
<script lang="ts">
  import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime.js';
  export let content: string;

  const REJECTED_SCHEMES = /^(javascript|data|file):/i;

  function sanitizeLink(href: string): string | null {
    try {
      const u = new URL(href, 'https://invalid/'); // dummy base for relative
      if (REJECTED_SCHEMES.test(u.protocol)) return null;
      return u.href;
    } catch { return null; }
  }

  function onLinkClick(e: MouseEvent, href: string) {
    e.preventDefault();
    const safe = sanitizeLink(href);
    if (!safe) return;
    if (confirm(`Open ${safe}?`)) {
      BrowserOpenURL(safe);
    }
  }
  // … rendered markdown passes each link through sanitizeLink and binds onLinkClick.
</script>
```

### Modal integration (Layer-1 preservation)

`InputResponseModal.svelte` subscribes to both `pendingPrompt` and `pendingAst`. When `pendingAst` is null (legacy snapshot or disabled adapter), render the existing Layer-1 UX unchanged — transcript + single widget keyed off `Spec.Shape`. When `pendingAst` is non-null, render the AST's nodes above the Layer-1 widget. The Layer-1 widget stays visible until U7 swaps it out based on `decision_group` presence.

### Design tokens (mirror `bmad-interactive-06` widget palette)

- `MarkdownBlock`: body font-stack, muted prose color.
- `HintBanner` tones (DESIGN.md canonical mapping, confirmed 2026-04-21): `info` (`var(--accent-info)` / `#3d9eff` blue border-left), `warn` (`var(--accent-amber)`), `error` (`var(--accent-red)`), `success` (`var(--accent-green)` / `#00e57a`, the brand accent). No tone reuses another tone's hue.
- `SummaryCard`: tinted panel with bullet list; collapsible.
- `CodeBlock`: `font-monospace` from editor-settings store; copy button uses existing Lucide `copy` icon.
- `ComparisonTable`: zebra rows, header row bold, border-thin.

### Risks / gotchas

- **Markdown renderer choice.** `@milkdown` / `markdown-it` / `remark` all work. Pick the lightest dependency that ships without a Monaco-style bundle balloon. The Crepe editor (already used in `MarkdownEditor.svelte`) would be overkill — prefer `markdown-it` with `linkify: false` and a custom `renderer.rules.link_open` hook for the sanitiser.
- **Unknown-type rendering is additive, not "view raw".** The fallback to markdown is the §3.3 rule-1 mirror on the frontend — U8 adds the "View raw" toggle on top.
- **Derived store `set(null)` semantics.** Svelte `derived` with the `(store, set)` signature requires calling `set` eventually — the initial run fires with `$p === undefined` so `set(null)` must execute in the guard clause.
- **Snackbar preview.** Spec §11 Q3 recommends using `turn_summary` on the snackbar — U6 does NOT wire this; follow-up polish. Document as a TODO in the modal.
- **XSS in markdown rendering.** `markdown-it` escapes HTML by default — DO NOT enable `html: true`. Audit the renderer config in the PR description.
- **Copy-to-clipboard on `CodeBlock`.** Use `navigator.clipboard.writeText`. No Wails binding needed. Per spec §7.2 point 4, this is already available via raw capture — no escalation.

### Reference files

- `frontend/src/components/bmad/inputWidgets/*.svelte` — existing Layer-1 widgets to leave untouched.
- `frontend/src/components/bmad/InputResponseModal.svelte` — existing modal (extend, don't replace).
- `docs/mashed-ui-ast-schema.md` §3 (schema), §6 (frontend integration), §7.2 (security).

## Acceptance Criteria

**AC-1: `PendingPrompt` wire format carries `structured?: string`**
- Given a `PendingPrompt` payload from a Wails event with `structured: "..."`
- When the TypeScript `PendingPrompt` type is used
- Then the `structured` field typechecks as `string | undefined`
- And absence in the payload leaves the field `undefined`
- Verified by `interactiveInput.test.ts` → `structured-field-optional`

**AC-2: `pendingAst` derived store parses valid v1 AST**
- Given `pendingPrompt.structured` is a valid v1 `UIAST` JSON string
- When `pendingAst` emits
- Then the emitted value is a parsed `UIAST` object with `version === "1"`
- Verified by `interactiveInput.test.ts` → `pending-ast-parses-valid`

**AC-3: `pendingAst` emits null on malformed JSON**
- Given `pendingPrompt.structured` is `"not-json"`
- When `pendingAst` emits
- Then the emitted value is `null`
- Verified by `interactiveInput.test.ts` → `pending-ast-null-on-malformed`

**AC-4: `pendingAst` emits null on unknown version**
- Given `pendingPrompt.structured` is `{"version":"2"}`
- When `pendingAst` emits
- Then the emitted value is `null`
- Verified by `interactiveInput.test.ts` → `pending-ast-null-on-unknown-version`

**AC-5: `AstNode` renders each of the six node shapes**
- Given a UIAST with one `markdown`, one `hint`, one `summary`, one `code`, one `table`, and one unknown-type node
- When `AstNode` is rendered against each
- Then the correct subcomponent renders for each
- And the unknown-type node falls back to `MarkdownBlock`
- Verified by component tests (`AstNode.test.ts`) + Playwright `ui-ast-rendering.spec.ts` → `six-node-shapes`

**AC-6: §7.2 security — `javascript:` href is stripped to plain text**
- Given a markdown node whose content contains `[click me](javascript:alert(1))`
- When `MarkdownBlock` renders it
- Then the rendered DOM contains the text "click me" with NO `<a>` tag (or an `<a>` with no `href`)
- And clicking the element does NOT invoke `alert`
- Verified by `MarkdownBlock.test.ts` → `javascript-href-stripped` and Playwright assertion

**AC-7: §7.2 security — `data:` and `file:` schemes are rejected identically**
- Given markdown content with `[x](data:text/html,...)` and `[y](file:///etc/passwd)`
- When rendered
- Then both links are stripped to plain text
- Verified by `MarkdownBlock.test.ts` → `data-and-file-schemes-rejected`

**AC-8: §7.2 security — valid https link routes through confirm dialog**
- Given markdown content with `[docs](https://example.com)`
- When the user clicks the rendered link
- Then a confirm dialog appears with the URL
- And on confirm, `BrowserOpenURL("https://example.com")` is invoked
- And on cancel, nothing is opened
- Verified by Playwright `ui-ast-rendering.spec.ts` → `https-link-confirm-dialog`

**AC-9: `InputResponseModal` renders AST above Layer-1 widget when `pendingAst != null`**
- Given `pendingAst` emits a valid UIAST
- When the modal is open
- Then the `AstNode` tree renders above the Layer-1 widget (the widget is still visible since U7 hasn't replaced it)
- And when `pendingAst` is null, the modal behaves exactly as pre-U6
- Verified by Playwright `ui-ast-rendering.spec.ts` → `modal-renders-ast-above-widget` and `modal-layer1-unchanged-when-null`

**AC-10: Unknown node type renders as markdown, not removed**
- Given a UIAST node with `type: "snarkfish", content: "hello"`
- When `AstNode` renders it
- Then a `MarkdownBlock` renders with content `"hello"`
- And no JavaScript error is thrown
- Verified by `AstNode.test.ts` → `unknown-type-markdown-fallback`

**AC-11: `CodeBlock` copy-to-clipboard copies verbatim content**
- Given a `code` node with `content: "rm -rf ~"` and `copyable: true`
- When the user clicks the copy button
- Then `navigator.clipboard.writeText` is called with `"rm -rf ~"`
- And the button briefly shows a "Copied" confirmation
- Verified by Playwright `ui-ast-rendering.spec.ts` → `code-copy-button-works`

## BDD Test Scenarios

```gherkin
Feature: AST dispatcher + passive node components

  Scenario: pendingAst parses a valid v1 AST
    Given pendingPrompt.structured is a valid v1 JSON
    When pendingAst emits
    Then the emitted value has version "1"

  Scenario: pendingAst emits null on malformed JSON
    Given pendingPrompt.structured is "not-json"
    When pendingAst emits
    Then the emitted value is null

  Scenario: AstNode renders markdown for a markdown node
    Given a node with type markdown and content "# Hi"
    When AstNode renders it
    Then a MarkdownBlock renders with content "# Hi"

  Scenario: AstNode renders a hint banner for a hint node
    Given a node with type hint and tone warn
    When AstNode renders it
    Then a HintBanner with tone warn is visible

  Scenario: Unknown type falls back to markdown
    Given a node with type snarkfish and content "hello"
    When AstNode renders it
    Then a MarkdownBlock with content "hello" is rendered
    And no error is thrown

  Scenario: javascript href is stripped
    Given a markdown node with "[x](javascript:alert(1))"
    When MarkdownBlock renders
    Then the rendered DOM has no a tag with javascript href

  Scenario: data and file schemes are rejected
    Given markdown links with data: and file: schemes
    When rendered
    Then both degrade to plain text

  Scenario: Valid https link routes through confirm
    Given a markdown link to https://example.com
    When the user clicks and confirms
    Then BrowserOpenURL is called with the URL

  Scenario: Modal renders AST above Layer-1 when pendingAst non-null
    Given pendingAst emits a valid AST
    When the modal opens
    Then the AST nodes render above the Layer-1 widget

  Scenario: Modal is Layer-1 unchanged when pendingAst is null
    Given pendingAst is null
    When the modal opens
    Then only the Layer-1 widget is visible

  Scenario: CodeBlock copy writes content verbatim
    Given a code node with copyable true
    When the user clicks the copy button
    Then navigator.clipboard.writeText is called with the content
```

## Tasks / Subtasks

- [ ] Task 1: TypeScript schema + `pendingAst` store (AC-1, AC-2, AC-3, AC-4)
  - [ ] RED: Vitest cases for the four store behaviours
  - [ ] GREEN: add `types/uiAst.ts` + extend `interactiveInput.ts`
- [ ] Task 2: Passive node components (AC-5, AC-10, AC-11)
  - [ ] RED: one failing component test per component
  - [ ] GREEN: write `MarkdownBlock`, `HintBanner`, `SummaryCard`, `CodeBlock`, `ComparisonTable`
  - [ ] RED: `AstNode.test.ts` exercising dispatcher switch
  - [ ] GREEN: write `AstNode.svelte`
- [ ] Task 3: §7.2 link sanitisation (AC-6, AC-7, AC-8)
  - [ ] RED: three failing tests (javascript, data, file schemes + https confirm)
  - [ ] GREEN: implement `sanitizeLink` + `onLinkClick` in `MarkdownBlock`
  - [ ] Disable markdown-it autolinking
- [ ] Task 4: Modal integration (AC-9)
  - [ ] RED: Playwright case for AST-above-Layer-1 + Layer-1-unchanged-when-null
  - [ ] GREEN: subscribe `InputResponseModal` to `pendingAst` and render conditionally
- [ ] Task 5: Playwright AC suite (AC-5, AC-6, AC-7, AC-8, AC-9, AC-11)
  - [ ] Write `tests/ac/ui-ast-rendering.spec.ts`
  - [ ] Invoke `/playwright-cli` to validate ACs visually
- [ ] Task 6: Refactor + docs
  - [ ] Run `/simplify` on every new Svelte + TS file
  - [ ] Add component doc-comments referencing spec §6.2

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on new Svelte components + store
- [ ] `npm run test` (Vitest) passes
- [ ] Playwright ACs pass via `/playwright-cli`
- [ ] Frontend build (`wails build` smoke test) passes
- [ ] `/simplify` run on every modified file; no CRITICAL/HIGH findings
- [ ] No `html: true` anywhere in the markdown config (XSS guard)
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch

## Design Brief

### 1. Layout composition

The AST node tree renders inside `InputResponseModal.svelte`'s existing `.modal-body` flex column, slotted **above** the existing `TranscriptPane` → `prompt-block` pair is NOT what we want. Correct stack (top-to-bottom) per story spec + spec §6.3:

1. `.modal-header` (existing, unchanged) — icon + title + round-pill + close.
2. `.modal-body` (existing flex-column, `gap: var(--sp-lg)`) — now contains, in order:
   - `TranscriptPane` (existing, prior-round history).
   - **NEW `ast-region`** — the AST nodes. Rendered ABOVE the `prompt-block` to honour "AST renders above existing transcript, BELOW modal subtitle". The `prompt-block` acts as the subtitle/turn prompt anchor; AST sits between the transcript and the prompt block? No — re-read story description: "AST renders ABOVE existing transcript, BELOW modal subtitle." The "existing transcript" in that sentence refers to the bottom Layer-1 widget context. Final order, confirmed:
     - `TranscriptPane` (prior rounds) — top.
     - `prompt-block` (current-turn subtitle / summary from backend) — acts as the modal subtitle anchor.
     - `ast-region` (NEW) — rendered directly under `prompt-block`.
     - Layer-1 widget (existing `FreeText/Choice/...` based on `Spec.Shape`) — bottom.
     - Help text (existing).
3. `.error-bar` (existing, unchanged).
4. `.modal-footer` with Cancel (existing).

- **`ast-region` grid:** vertical flex column, `gap: var(--sp-md)` (tighter than the body's `--sp-lg` so AST nodes feel like a grouped unit, separate from transcript/widget).
- **Node internal padding:** each AST sub-component owns its own padding; the region itself has `padding: 0` and lets children bleed to the full body width.
- **Separator between `prompt-block` and `ast-region`:** none. The green left-stripe of `prompt-block` naturally terminates the subtitle; a horizontal rule would fracture the rhythm. The `gap: var(--sp-lg)` inside `.modal-body` is the only separator.
- **Width:** AST nodes inherit `.modal-card` max-width (880px standard, 720px wide) — no new max-width. When the modal is in `editor` mode (JSON shape), the AST region still renders normally; the 480px min-height comes from `.modal-body.editor`.
- **Entry motion (staggered):** on modal mount, AST nodes animate in with `fly={{ y: 8, duration: var(--duration-medium), delay: i * 40ms, easing: cubicOut }}`. 40ms stagger is short enough to feel like one choreographed move (total < 250ms for a 6-node AST) and matches the existing modal's 200ms mount duration. Do NOT stagger if `prefers-reduced-motion` is set.

### 2. Typography plan

Every role maps to an **existing** token from `style.css`. Introducing a new `--text-*` or `--font-*` variable is prohibited.

| Role | Token | Font family | Weight | Line-height |
|------|-------|-------------|--------|-------------|
| Turn summary / modal subtitle (existing `prompt-block`) | `var(--text-data)` 14px | `var(--font-ui)` | 400 | 1.6 (existing) |
| `MarkdownBlock` body paragraphs | `var(--text-body)` 13px | `var(--font-ui)` | 400 | 1.55 |
| `MarkdownBlock` inline `code` | `var(--text-body)` 13px | `var(--font-code)` | 500 | 1.55 |
| `MarkdownBlock` `h1/h2/h3` | 16/14/13px (`--text-section`/`--text-data`/`--text-body`) | `var(--font-ui)` | 600 | 1.4 |
| `HintBanner` body | `var(--text-body)` 13px | `var(--font-ui)` | 500 | 1.5 |
| `HintBanner` tone label (icon alt-text) | `var(--text-label)` 11px uppercase | `var(--font-mono)` | 600, `letter-spacing: 0.05em` | 1 |
| `SummaryCard` heading | `var(--text-data)` 14px | `var(--font-ui)` | 600 | 1.3 |
| `SummaryCard` bullets | `var(--text-body)` 13px | `var(--font-ui)` | 400 | 1.5 |
| `CodeBlock` content | `var(--text-body)` 13px | `var(--font-code)` | 400 | 1.55, `font-variant-numeric: tabular-nums` |
| `CodeBlock` lang tag (top-right) | `var(--text-label)` 11px | `var(--font-mono)` | 500 uppercase, `letter-spacing: 0.05em` | 1 |
| `ComparisonTable` heading | `var(--text-data)` 14px | `var(--font-ui)` | 600 | 1.3 |
| `ComparisonTable` column headers | `var(--text-label)` 11px | `var(--font-mono)` | 600 uppercase, `letter-spacing: 0.05em` | 1.4 |
| `ComparisonTable` cells | `var(--text-body)` 13px | `var(--font-ui)` | 400, `tabular-nums` for numeric columns | 1.5 |
| Confirm-dialog URL preview | `var(--text-body)` 13px | `var(--font-code)` | 500 | 1.4, `word-break: break-all` |
| Confirm-dialog button labels | `var(--text-body)` 13px | `var(--font-ui)` | 600 | 1 |

**MarkdownBlock consistency rule:** the only other markdown rendering in the BMAD UI is the Crepe editor inside `MarkdownEditor.svelte`, which is an editor not a renderer. Since there is no precedent for read-only markdown in the app, `MarkdownBlock` ESTABLISHES the pattern — its prose rhythm (13px body, 1.55 line-height, `var(--font-ui)`) becomes the canonical markdown type scale. If a future component needs read-only markdown, it reuses `MarkdownBlock`.

### 3. Color strategy

- **Primary surface (AST region):** `var(--bg-surface)` inherited from the modal card. Nodes sit directly on this surface — no extra panel frame around the region.
- **Accent usage:** `var(--accent-green)` is reserved for the existing `prompt-block` left-stripe and modal focus rings. AST nodes DO NOT use the accent as a primary surface color — overuse would dilute the "alive" semantic.
- **Hint tone palette (load-bearing — U8 must reuse this exactly):**
  - `info` — background `color-mix(in srgb, var(--accent-blue) 10%, transparent)`, stripe & icon `var(--accent-blue)` (#3d9eff), body `var(--text-primary)`.
  - `warn` — background `color-mix(in srgb, var(--accent-amber) 10%, transparent)`, stripe & icon `var(--accent-amber)` (#f0a500), body `var(--text-primary)`. Matches U5's offline banner and U8's `Untrusted` chip.
  - `error` — background `color-mix(in srgb, var(--accent-red) 10%, transparent)`, stripe & icon `var(--accent-red)` (#e84545), body `var(--text-primary)`. Same alpha as the modal's `.error-bar` so errors feel consistent regardless of origin.
  - `success` — background `color-mix(in srgb, var(--accent-green) 10%, transparent)`, stripe & icon `var(--accent-green)` (#00e57a), body `var(--text-primary)`. Canonical per DESIGN.md + 2026-04-21 user confirmation: `info` is blue, `success` is green.
- **`SummaryCard`:** background `var(--bg-elevated)`, border 1px `var(--border-subtle)`, heading `var(--text-primary)`, bullets `var(--text-primary)`, bullet markers `var(--accent-green)` (the one accent moment per card — ties summaries to the "completed/ready" semantic).
- **`CodeBlock`:** background `var(--bg-deepest)` (darkest surface, matches Terminal pane), border 1px `var(--border-subtle)`, lang tag `var(--text-muted)`, copy button `var(--text-dim)` idle → `var(--accent-green)` on success ("Copied").
- **`ComparisonTable`:** header row background `var(--bg-elevated)`, header text `var(--text-dim)`, body rows alternate `transparent` and `color-mix(in srgb, var(--bg-elevated) 50%, transparent)` (subtle zebra — never more than 50% alpha or it becomes noisy at 13px density), border-top/bottom `1px solid var(--border-subtle)`, cell text `var(--text-primary)`.
- **Confirm-dialog for external links (§7.2):** background `var(--bg-elevated)` on top of the modal's existing backdrop — but since the confirm is a single-decision moment, a **native `window.confirm()`** is acceptable per the story's sketch and avoids nested-modal complexity. If a custom dialog is chosen instead, use `var(--bg-surface)` card + `var(--border-emphasis)` border, URL in a `var(--bg-deepest)` pre-block, buttons follow `ApprovalWidget` hierarchy ("Open" = green solid like `.btn-yes`, "Cancel" = outline like `.btn-no`).
- **Text contrast levels inside AST:**
  - Emphasised: `var(--text-primary)` for paragraph copy, summary headings, code, cells.
  - Default: `var(--text-dim)` for bullet marker labels, table column headers, lang tags.
  - Muted: `var(--text-muted)` for code lang tag, "copied" confirmation (when inactive).

### 4. Interaction model

- **Keyboard nav (spec §6.4):**
  - Tab order within AST region: `MarkdownBlock` links (if any) → `CodeBlock` copy button → `ComparisonTable` is not interactive (skip) → `HintBanner` is not interactive (skip) → `SummaryCard` collapse/expand chevron (if collapsible).
  - Since U6 has no `decision_group` yet, AST region's tab stops are sparse — the existing Layer-1 widget below gets focus naturally via `onMount` inside each widget. Do not alter that existing focus behaviour.
  - Escape closes the whole modal (existing handler at window-level — do not duplicate).
  - Enter on a focused markdown link opens the confirm dialog (same path as click).
- **Cmd+Enter:** in U6 there is nothing to submit via AST alone — the Layer-1 widget still owns submission. Do not bind Cmd+Enter at the modal level in U6 (U7 will add it).
- **Link confirm dialog:** `window.confirm()` is synchronous but acceptable for MVP. Text format: `Open ${safeURL}?` exactly. On confirm → `BrowserOpenURL(safeURL)`. On cancel → no-op. No telemetry, no styling (deliberately using OS chrome signals "this is a trust moment, pay attention"). A follow-up story can replace with a custom Svelte dialog.
- **Hover states:**
  - `MarkdownBlock` `<a>` links: `color: var(--accent-green)`, underline on hover via `text-decoration-thickness: 1.5px` + `text-underline-offset: 2px`. Idle: `color: var(--accent-green)`, no underline.
  - `SummaryCard` collapse chevron: `color: var(--text-dim)` → `var(--text-primary)` on hover.
  - `CodeBlock` copy button: `color: var(--text-dim)` → `var(--accent-green)` on hover; after successful copy, briefly `color: var(--accent-green)` with `background: color-mix(var(--accent-green) 15%)` for 1200ms then decays back via `var(--duration-short)` `var(--ease-exit)`.
  - `ComparisonTable` rows: no hover affordance (rows are not interactive).
- **Focus ring:** global `:focus-visible` rule from `style.css` (2px outline `var(--accent-green)`, offset 2px) applies to every interactive element automatically. Do not override.
- **Submit disabled state:** N/A for U6 (submission still runs through Layer-1 widget).
- **Collapse/expand animation (`SummaryCard` optional collapse, `CodeBlock` never collapses):** `slide` transition, `duration: var(--duration-medium)` (120-150ms per brief request — use the existing 150ms token), easing `var(--ease-move)` for open, `var(--ease-exit)` for close. Chevron rotates 90° via `transform: rotate(var(--angle))` over the same duration.
- **Animation discipline:** every transition uses one of three tokens — `var(--duration-micro)`, `var(--duration-short)`, `var(--duration-medium)`. No custom ms values inline.

### 5. Component specs

All measurements source CSS custom properties.

- **`ast-region` container:** `display: flex; flex-direction: column; gap: var(--sp-md); padding: 0; width: 100%;`. No border, no background (inherits modal).
- **`MarkdownBlock`:** `padding: 0; margin: 0; color: var(--text-primary);`. Paragraph spacing: `p + p { margin-top: var(--sp-sm); }`. Headings spacing: `h1/h2/h3 { margin-top: var(--sp-md); margin-bottom: var(--sp-xs); }` except first-child (`margin-top: 0`). Inline code: `padding: 0 var(--sp-2xs); background: var(--bg-elevated); border-radius: var(--radius-sm);`. Lists: `padding-left: var(--sp-lg); li + li { margin-top: var(--sp-2xs); }`.
- **`HintBanner` (matches `.prompt-block` + U5 offline-banner rhythm):**
  - `display: flex; align-items: flex-start; gap: var(--sp-sm);`.
  - `padding: var(--sp-sm) var(--sp-md);`.
  - `border-radius: 0 var(--radius-md) var(--radius-md) 0;` (flat on stripe side — mirrors `.prompt-block`).
  - `border-left: 3px solid <tone-color>;` where `<tone-color>` is the tone's accent (see §3).
  - `background: color-mix(in srgb, <tone-color> 10%, transparent);`.
  - Icon: 16px Lucide `Info` / `AlertTriangle` / `XCircle` / `CheckCircle`, `color: <tone-color>`.
- **`SummaryCard`:**
  - `padding: var(--sp-md);`.
  - `background: var(--bg-elevated);`.
  - `border: 1px solid var(--border-subtle);`.
  - `border-radius: var(--radius-md);`.
  - Heading-to-bullets gap: `var(--sp-sm)`.
  - Bullets list: `list-style: none; padding: 0;`. Each `li` prefixed with a `›` character in `var(--accent-green)` at `font-weight: 700` (same chevron idiom as `ChoiceWidget`'s `.option-chevron`).
- **`CodeBlock`:**
  - `padding: var(--sp-sm) var(--sp-md);`.
  - `background: var(--bg-deepest);` (darker than surface — code is "terminal-adjacent").
  - `border: 1px solid var(--border-subtle);`.
  - `border-radius: var(--radius-md);`.
  - `font-family: var(--font-code);` (JetBrains Mono).
  - `font-size: var(--text-body);` 13px (per DESIGN.md "Terminal pane: JetBrains Mono 13px").
  - `line-height: 1.55;`.
  - `overflow-x: auto; white-space: pre;` — horizontal scroll, never wrap.
  - Header strip: `display: flex; justify-content: space-between; align-items: center; margin-bottom: var(--sp-xs);`. Left: lang tag. Right: copy button (Lucide `Copy` 14px icon + "Copy" label at `var(--text-label)` 11px).
  - Copy success state: icon swaps to `Check` 14px in `var(--accent-green)`, label flips to "Copied".
- **`ComparisonTable`:**
  - Wrapper: `border: 1px solid var(--border-subtle); border-radius: var(--radius-md); overflow: hidden; overflow-x: auto;`.
  - Heading (if present, outside table): `padding: var(--sp-sm) var(--sp-md) var(--sp-xs); background: var(--bg-elevated);`.
  - Table: `width: 100%; border-collapse: collapse; font-size: var(--text-body);`.
  - `th`: `padding: var(--sp-xs) var(--sp-md); text-align: left; background: var(--bg-elevated); color: var(--text-dim); border-bottom: 1px solid var(--border-subtle); text-transform: uppercase; font-family: var(--font-mono); font-size: var(--text-label); font-weight: 600; letter-spacing: 0.05em;`.
  - `td`: `padding: var(--sp-xs) var(--sp-md); color: var(--text-primary); border-bottom: 1px solid color-mix(in srgb, var(--border-subtle) 50%, transparent);`.
  - Zebra row: `tbody tr:nth-child(even) td { background: color-mix(in srgb, var(--bg-elevated) 50%, transparent); }`.
  - Last-row `td { border-bottom: none; }`.
- **Confirm dialog (if custom, not native `window.confirm`):** width `var(--modal-width-min)` 360px, padding `var(--sp-lg)`, URL preview in a `<pre>` with `padding: var(--sp-sm) var(--sp-md); background: var(--bg-deepest); border-radius: var(--radius-sm); word-break: break-all;`. Button row `justify-content: flex-end; gap: var(--sp-md);` — "Cancel" left (outline, `ApprovalWidget.btn-no` style), "Open" right (solid green, `ApprovalWidget.btn-yes` style, `min-width: var(--btn-min-wide)` 96px).

### 6. Signature elements

- **Left-stripe motif as the repeated anchor.** The modal's `prompt-block` has a 3px green left-stripe. Every `HintBanner` mirrors that shape (3px stripe + tone-tinted background at 10% alpha + `border-radius: 0 var(--radius-md) var(--radius-md) 0`). One structural idea, four tonal variants. This is the strongest cross-component cue that we are inside a mashed workflow.
- **Staggered 40ms entry.** AST nodes flying in over ~240ms feels deliberate and structured — an AI slop UI would fade in everything at once. Reduced-motion users get instant render.
- **Mono-font tonal labels.** Hint labels, column headers, lang tags all use `var(--font-mono)` uppercase with `letter-spacing: 0.05em`. Same typographic cue as the Editor subsection titles (`.subsection-title`) in Settings — signals "machine-readable meta" vs human-readable body.
- **Zebra rhythm at 50% alpha.** `ComparisonTable` zebra rows use `color-mix(var(--bg-elevated) 50%)` instead of a hard color — at 13px density, hard zebra stripes look harsh; 50% alpha keeps density readable.
- **CodeBlock on deepest surface.** Setting `CodeBlock` background to `var(--bg-deepest)` instead of `var(--bg-surface)` (the modal's own background) creates a recessed-into-the-card effect that echoes the real Terminal pane treatment from `DESIGN.md`. Code "sits below" the surrounding prose — an intentional depth cue.
- **One-accent-per-card rule.** Each component gets exactly one `var(--accent-green)` moment: `SummaryCard` → bullet markers; `CodeBlock` → copy-success; `HintBanner success` → stripe; `MarkdownBlock` → links. Avoiding accent overuse is how the modal stays calm.

### Cross-story design coherence (binding rules)

- **`MarkdownBlock` prose type scale IS the BMAD UI markdown canon.** No existing read-only markdown renderer. Future BMAD markdown surfaces reuse this component or match its 13px / 1.55 / `var(--font-ui)` rhythm exactly.
- **`HintBanner` tone palette MUST match U5's offline banner (amber/warn) and U8's `DiagnosticsChip` (amber `Untrusted`, blue `N adapter notes`).** One palette, four tones, used across the whole UI AST surface.
- **Link confirm dialog (§7.2):** if implemented as a custom Svelte dialog, its button hierarchy reuses the `ApprovalWidget` green-solid + outline pattern exactly. Do not invent new button styles.
- **No new design tokens.** Every width, color, duration, radius must trace to an existing `style.css` custom property. If a value isn't covered, flag it in the PR — do not introduce a new `--ast-*` token.

