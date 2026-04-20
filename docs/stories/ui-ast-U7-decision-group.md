# ui-ast-U7: `DecisionGroup` + response-map collection + conditional JSON vs collapse submit path

**Status:** ready
**Domain:** frontend
**Size:** L
**Depends on:** ui-ast-U6
**Priority:** P0-critical

## Story

As a Mashed user answering a multi-decision Claude turn, I want each `decision_group` in the AST to render as a labelled inline widget (choice / multi / approval / free / file / json), collect responses into a keyed map, and submit through `RespondToInput` — as a JSON blob when the iteration spec is `ShapeJSON` or collapsed to a single string per §3.4 when it isn't — so that every decision I make survives the round-trip to Claude without me having to type a freeform summary.

## Description

Implements spec §3.4 (response payload shape + collapse rule) and §6.3 (modal layout that embeds AST nodes + collects responses + ships through `RespondToInput`). Ships the `DecisionGroup.svelte` component that dispatches to the six widgets already in `frontend/src/components/bmad/inputWidgets/` (from `bmad-interactive-06`). Wires the response-map reactive store and the Send-all-responses button. Honours the collapse rule: when the iteration `Spec.Shape != ShapeJSON` and the AST has multiple `decision_group` nodes, surface the banner "This process accepts a single answer — pick one decision to submit" and enable only one widget.

### Scope summary

- `frontend/src/components/bmad/DecisionGroup.svelte` (new) — labels + prompt + help + widget dispatch.
- `frontend/src/stores/astResponses.ts` (new) — per-modal `writable<Record<string, string>>` for response collection.
- `frontend/src/components/bmad/InputResponseModal.svelte` — Send-all-responses button; §3.4 submit path (JSON vs collapse); `Spec.Shape` wiring to decide submit format.
- `frontend/src/components/bmad/AstNode.svelte` — route `decision_group` to the new component (remove the U6 markdown fallback).
- Playwright ACs for JSON submission + collapse behaviour + per-widget rendering.

### Non-goals

- No "View raw" toggle (U8).
- No new widget components — reuse from `bmad-interactive-06`.
- No backend changes (U0 + U4 shipped).

## Developer Notes

### Files to create/modify

- `frontend/src/components/bmad/DecisionGroup.svelte` (new).
- `frontend/src/stores/astResponses.ts` (new).
- `frontend/src/components/bmad/InputResponseModal.svelte` (modify — Send button + submit path).
- `frontend/src/components/bmad/AstNode.svelte` (modify — route `decision_group` to `DecisionGroup`, not markdown).
- `frontend/src/components/bmad/__tests__/DecisionGroup.test.ts` (new).
- `frontend/src/components/bmad/__tests__/InputResponseModal.test.ts` (modify).
- `tests/ac/ui-ast-decision-group.spec.ts` (new Playwright).

### `DecisionGroup.svelte` (spec §6.2 child dispatch)

```svelte
<script lang="ts">
  import type { UINode, WidgetNode } from '../../types/uiAst';
  import ChoiceWidget from './inputWidgets/ChoiceWidget.svelte';
  import MultiChoiceWidget from './inputWidgets/MultiChoiceWidget.svelte';
  import ApprovalWidget from './inputWidgets/ApprovalWidget.svelte';
  import FreeTextWidget from './inputWidgets/FreeTextWidget.svelte';
  import FileInputWidget from './inputWidgets/FileInputWidget.svelte';
  import JsonInputWidget from './inputWidgets/JsonInputWidget.svelte';
  import type { Writable } from 'svelte/store';

  export let node: UINode; // type === 'decision_group'
  export let responses: Writable<Record<string, string>>;
  export let disabled = false;  // collapse-rule: when another group owns the turn

  const widget: WidgetNode = node.widget!;

  function onValue(v: string) {
    responses.update(r => ({ ...r, [node.response_key!]: v }));
  }
</script>

<section aria-labelledby={`dg-${node.response_key}`}>
  <h3 id={`dg-${node.response_key}`}>{node.heading ?? ''}</h3>
  {#if node.prompt}<p class="prompt">{node.prompt}</p>{/if}
  {#if node.help}<p class="help">{node.help}</p>{/if}

  {#if widget.type === 'choice'}
    <ChoiceWidget options={widget.options ?? []} default={widget.default ?? ''} {disabled} on:change={e => onValue(e.detail)} />
  {:else if widget.type === 'multi'}
    <MultiChoiceWidget options={widget.options ?? []} min={widget.min ?? 0} max={widget.max ?? 0} {disabled} on:change={e => onValue(e.detail)} />
  {:else if widget.type === 'approval'}
    <ApprovalWidget yesLabel={widget.yes_label ?? 'Yes'} noLabel={widget.no_label ?? 'No'} default={widget.default ?? ''} {disabled} on:change={e => onValue(e.detail)} />
  {:else if widget.type === 'free'}
    <FreeTextWidget placeholder={widget.placeholder ?? ''} maxLength={widget.maxLength ?? 2000} multiline={widget.multiline ?? true} {disabled} on:change={e => onValue(e.detail)} />
  {:else if widget.type === 'file'}
    <FileInputWidget accept={widget.accept ?? []} repoRootRelative={widget.repoRootRelative ?? true} {disabled} on:change={e => onValue(e.detail)} />
  {:else if widget.type === 'json'}
    <JsonInputWidget schema={widget.schema ?? null} {disabled} on:change={e => onValue(e.detail)} />
  {/if}
</section>
```

### `astResponses` store

```ts
// frontend/src/stores/astResponses.ts
import { writable } from 'svelte/store';

export function makeAstResponses() {
  return writable<Record<string, string>>({});
}
```

Constructed per-modal-instance so responses reset between suspensions.

### §3.4 submit path — spec verbatim (lines 206–224)

```ts
// InputResponseModal.svelte
async function onSend() {
  const r = get(responses);
  const ast = get(pendingAst);
  const shape = pendingPromptValue?.spec?.shape; // from PendingPrompt.Spec.Shape

  if (shape === 'json') {
    // ShapeJSON: ship the full map (lossless; executor flattens per §5.3).
    await RespondToInput(execId, nodeId, inputId, JSON.stringify(r));
    return;
  }

  // Non-JSON iteration specs — collapse rule.
  const groups = decisionGroups(ast);
  if (groups.length === 0) {
    // No decision_group — use fallback_answer_shape widget value from the
    // bottom-of-modal fallback input. Existing Layer-1 behaviour.
    await RespondToInput(execId, nodeId, inputId, fallbackValue);
    return;
  }
  if (groups.length === 1) {
    const key = groups[0].response_key!;
    await RespondToInput(execId, nodeId, inputId, r[key] ?? '');
    return;
  }
  // Multiple decision_groups + non-JSON shape: the UI enforced "pick one".
  const active = groups.find(g => g.required) ?? groups[0];
  const key = active.response_key!;
  // diagnostics.collapsed flag — surfaced in the snackbar by a separate store event.
  await RespondToInput(execId, nodeId, inputId, r[key] ?? '');
}
```

### Collapse-rule banner (spec §3.4 rule 2 line 216)

When `shape != 'json'` AND `groups.length > 1`:

```
┌ ⚠ This process accepts a single answer ─────────┐
│ Pick one decision to submit. Others are hidden. │
└──────────────────────────────────────────────────┘
```

Only one `DecisionGroup` has `disabled = false`; the first `required: true` group (else first group) is the active one. User clicks any group to make it active (the modal swaps which `disabled` flag is set).

### Modal layout (spec §6.3 lines 622–640)

```
┌ Input from {procName} — Round N ─────────────────────────┐
│  [transcript: Layer-1 renders prior rounds here]          │
│  ════════════════════════════════════════════════════════ │
│  Current turn:                                            │
│  {turn_summary}                                           │
│                                                           │
│  <AstNode/> for each node in ast.nodes                    │
│    - markdown prose (U6)                                  │
│    - hint banners (U6)                                    │
│    - decision_group with inline widget (U7)               │
│    - …                                                    │
│                                                           │
│  [ View raw ▾ ] (only if diagnostics.untrusted — U8)      │
│                                                           │
│  [ Send all responses ]  [ Cancel ]                       │
└───────────────────────────────────────────────────────────┘
```

`Send` is disabled until every `required: true` decision_group has a value.

### Keyboard + a11y (spec §6.4 lines 644–649)

- Tab order: transcript → each `decision_group` widget in order → Send → Cancel.
- `Cmd+Enter` submits the whole form regardless of focus.
- Each `DecisionGroup` has `aria-labelledby` linking heading → widget.
- Passive AST content (markdown / hint / summary / code / table from U6) is `role="region"` with an `aria-label` drawn from the node's `heading` or `turn_summary`.

### Risks / gotchas

- **`Spec.Shape` vs AST `fallback_answer_shape`.** `PendingPrompt.Spec.Shape` is the BACKEND's declared shape (from U0's registry). `ast.fallback_answer_shape` is the ADAPTER's guess. Submission format follows `Spec.Shape` ALWAYS — never `fallback_answer_shape`. Spec §3.4 line 206: "the backend is the single source of truth". Easy to confuse; ACs codify.
- **Required-group validation.** The Send button's enabled state depends on `responses` having a value for every `required: true` group. For `multi` widgets, "value" means at least one checkbox — delegate to the widget's existing validation.
- **Collapse-rule selection UX.** Spec doesn't specify HOW the user picks which group becomes active — recommendation: click any group card to activate it; the others grey out with an "inactive" tint. Inactive groups' widgets are `disabled`.
- **`diagnostics.collapsed` flag.** Ship a snackbar event when the collapse path is taken so the user sees "decision was collapsed" feedback. Emit via the existing `NodeInputSnackbarStack` with a new `tone: "collapsed"` variant.
- **Widget change events.** Existing widgets from `bmad-interactive-06` dispatch `change` events with a detail payload. Ensure the payload shape matches what `DecisionGroup` expects (single string). If the multi widget dispatches an array, join with `,` or JSON.stringify — document in the widget-integration table.
- **File widget path resolution.** `FileInputWidget` returns a path string; backend still runs `resolveFileInput` at submit time (§7.1) — frontend does NOT pre-resolve absolute paths.

### Reference files

- `frontend/src/components/bmad/inputWidgets/*.svelte` — six widgets from `bmad-interactive-06`.
- `frontend/src/components/bmad/AstNode.svelte` — U6 dispatcher.
- `frontend/src/components/bmad/InputResponseModal.svelte` — U6 modal scaffolding.
- `docs/mashed-ui-ast-schema.md` §3.4 (response payload), §6.3 (modal layout), §6.4 (a11y).

## Acceptance Criteria

**AC-1: `DecisionGroup` dispatches each of the six widget types**
- Given a `decision_group` with `widget.type` in `{choice, multi, approval, free, file, json}`
- When `DecisionGroup` renders
- Then the correct widget from `inputWidgets/` is mounted
- Verified by `DecisionGroup.test.ts` → `dispatches-six-widget-types`

**AC-2: `responses` store accumulates one entry per `decision_group`**
- Given a modal with three `decision_group` nodes (keys `a`, `b`, `c`)
- When the user changes each widget's value
- Then `responses` emits `{a: "...", b: "...", c: "..."}`
- Verified by `InputResponseModal.test.ts` → `responses-accumulate`

**AC-3: §3.4 — `ShapeJSON` submits the full map as JSON**
- Given `Spec.Shape == 'json'` and three populated decision_groups
- When the user clicks Send
- Then `RespondToInput(execId, nodeId, inputId, '{"a":"x","b":"y","c":"z"}')` is called
- Verified by Playwright `ui-ast-decision-group.spec.ts` → `json-submit-full-map`

**AC-4: §3.4 — single decision_group on non-JSON shape submits plain string**
- Given `Spec.Shape == 'free'` and one decision_group with value `"option-1"`
- When the user clicks Send
- Then `RespondToInput(execId, nodeId, inputId, 'option-1')` is called (plain string)
- Verified by Playwright `ui-ast-decision-group.spec.ts` → `single-group-plain-string-submit`

**AC-5: §3.4 — multiple decision_groups on non-JSON shape trigger collapse banner**
- Given `Spec.Shape == 'free'` and three decision_groups
- When the modal opens
- Then a banner matching `/single answer/i` is visible
- And only one widget is enabled (the first `required:true`, else first)
- And the other widgets are `disabled`
- Verified by Playwright `ui-ast-decision-group.spec.ts` → `collapse-banner-visible`

**AC-6: Collapse-rule — user can switch which group is active**
- Given three decision_groups under the collapse banner
- When the user clicks a different group card
- Then that group's widget becomes enabled
- And the previously-active widget becomes disabled
- Verified by Playwright `ui-ast-decision-group.spec.ts` → `collapse-user-switches-active`

**AC-7: Send button disabled until required groups have values**
- Given a modal with two decision_groups, one `required: true`
- When the required group has no value
- Then the Send button is disabled
- And once the required group has a value, Send becomes enabled
- Verified by `InputResponseModal.test.ts` → `send-disabled-until-required-filled`

**AC-8: Cmd+Enter submits the whole form**
- Given focus anywhere inside the modal with all required groups filled
- When the user presses Cmd+Enter (or Ctrl+Enter on non-macOS)
- Then `RespondToInput` is called with the same payload as clicking Send
- Verified by Playwright `ui-ast-decision-group.spec.ts` → `cmd-enter-submits`

**AC-9: Zero decision_groups falls back to Layer-1 widget**
- Given an AST with zero decision_group nodes and `fallback_answer_shape: "free"`
- When the modal renders
- Then the Layer-1 `FreeTextWidget` is visible below the AST nodes
- And Send uses the fallback widget's value as the single-string submission
- Verified by Playwright `ui-ast-decision-group.spec.ts` → `zero-groups-fallback-layer1`

**AC-10: `aria-labelledby` wires heading → widget for each group**
- Given a decision_group with heading "Output sink"
- When rendered
- Then the widget's `aria-labelledby` attribute points to the heading's `id`
- And the heading renders with a matching `id`
- Verified by `DecisionGroup.test.ts` → `aria-labelledby-wired`

**AC-11: §7.1 security — file widget path string is sent verbatim to backend**
- Given a decision_group with `widget.type: "file"`
- When the user selects `../../etc/passwd` and clicks Send (on a `ShapeFile` iteration)
- Then `RespondToInput` receives the path string verbatim (no pre-resolution)
- And the backend's `resolveFileInput` path-traversal rejection fires — asserted indirectly by the absence of any frontend URL-normalisation
- Verified by `InputResponseModal.test.ts` → `file-widget-passthrough`

## BDD Test Scenarios

```gherkin
Feature: DecisionGroup + response-map collection + §3.4 submit path

  Scenario: DecisionGroup dispatches to choice widget
    Given a decision_group with widget type choice
    When DecisionGroup renders
    Then a ChoiceWidget is mounted

  Scenario: Three decision_groups accumulate responses
    Given three decision_groups with keys a, b, c
    When the user fills all three
    Then responses emits {a:..., b:..., c:...}

  Scenario: ShapeJSON submits full map
    Given Spec.Shape json and three filled groups
    When Send is clicked
    Then RespondToInput receives a JSON map string

  Scenario: Single group on non-JSON shape submits plain string
    Given Spec.Shape free and one decision_group with value "x"
    When Send is clicked
    Then RespondToInput receives "x"

  Scenario: Multiple groups on non-JSON shape show collapse banner
    Given Spec.Shape free and three decision_groups
    When the modal opens
    Then a banner matching "single answer" is visible
    And only one widget is enabled

  Scenario: User switches active group under collapse rule
    Given the collapse banner is visible
    When the user clicks a different group card
    Then that group's widget becomes enabled and the previous disables

  Scenario: Send disabled until required groups filled
    Given a required decision_group with no value
    When the modal is open
    Then the Send button is disabled

  Scenario: Cmd+Enter submits the form
    Given all required fields filled
    When the user presses Cmd+Enter
    Then RespondToInput is called with the expected payload

  Scenario: Zero decision_groups uses Layer-1 fallback
    Given an AST with no decision_group nodes
    When the modal opens
    Then the Layer-1 widget is visible
    And Send submits the fallback widget value

  Scenario: File widget path is sent verbatim to backend
    Given a file-widget decision_group with path "../../etc/passwd"
    When Send is clicked on a ShapeFile iteration
    Then RespondToInput receives the literal string without URL normalisation
```

## Tasks / Subtasks

- [ ] Task 1: `DecisionGroup` component (AC-1, AC-10)
  - [ ] RED: Vitest for six-widget dispatch + aria-labelledby
  - [ ] GREEN: write `DecisionGroup.svelte`
- [ ] Task 2: `astResponses` store + modal integration (AC-2, AC-7)
  - [ ] RED: Vitest for accumulate + Send-disabled gate
  - [ ] GREEN: wire store + Send-button state
- [ ] Task 3: §3.4 submit path (AC-3, AC-4, AC-5, AC-6, AC-9)
  - [ ] RED: five Playwright cases
  - [ ] GREEN: implement `onSend` with the JSON / collapse / fallback branches
  - [ ] Implement collapse banner + active-group switching UI
- [ ] Task 4: Keyboard + a11y (AC-8, AC-10)
  - [ ] Add Cmd+Enter handler
  - [ ] Playwright assertion
- [ ] Task 5: Security passthrough (AC-11)
  - [ ] Vitest for file-widget literal passthrough
- [ ] Task 6: Integrate `DecisionGroup` into `AstNode.svelte` (AC-1)
  - [ ] Replace U6's markdown fallback for `decision_group` with the new component
- [ ] Task 7: Playwright AC suite
  - [ ] Write `tests/ac/ui-ast-decision-group.spec.ts`
  - [ ] Invoke `/playwright-cli` to validate ACs visually

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on new Svelte components + store
- [ ] Vitest passes
- [ ] Playwright ACs pass via `/playwright-cli`
- [ ] Frontend build passes
- [ ] `/simplify` run on every modified file; no CRITICAL/HIGH findings
- [ ] Cmd+Enter + Tab order verified manually on macOS
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch

## Design Brief

### 1. Layout composition

`DecisionGroup` replaces U6's markdown-fallback for `decision_group` nodes inside the existing `ast-region` (from U6). The modal's `.modal-body` stack becomes:

1. `.modal-header` (unchanged).
2. `TranscriptPane` (unchanged).
3. `prompt-block` — current-turn subtitle (unchanged).
4. **Collapse banner** (NEW, conditional — only when `shape != 'json' && groups.length > 1`).
5. `ast-region` — AST nodes, now including `DecisionGroup` components inline with passive nodes.
6. Layer-1 widget (visible ONLY when `groups.length === 0`; in §3.4's `ast.length > 0` path it is removed entirely).
7. Help text (unchanged).
8. `.error-bar` (unchanged).
9. **Footer actions** (MODIFIED) — `Send all responses` primary + `Cancel` secondary.

- **`DecisionGroup` card within `ast-region`:** each group renders as a distinct `<section>` card with the **same structural rhythm as the widget shells inside `inputWidgets/`** — padding, border, radius, inner gap identical to a widget row. A `DecisionGroup` is effectively a widget-with-a-heading.
- **Heading → prompt → help → widget stack inside each card:**
  - Heading (`<h3>`): margin 0, `font-size: var(--text-data)`, `font-weight: 600`.
  - Prompt (if present): `var(--text-body)`, `color: var(--text-primary)`, `margin-top: var(--sp-2xs)`.
  - Help (if present): `var(--text-label)`, `color: var(--text-dim)`, `font-style: italic`, `margin-top: var(--sp-2xs)`.
  - Gap between last-text-block and widget: `var(--sp-md)`.
- **Region-level `gap` between cards:** `var(--sp-md)` — identical to U6's `ast-region` gap so mixed passive + decision nodes read as one list.
- **Collapse banner:** renders BETWEEN `prompt-block` and `ast-region` when active. Full-width banner, uses the `HintBanner warn` tone from U6. Contains short title ("This process accepts a single answer") + one-line instruction ("Pick one decision to submit. Others are hidden.").
- **Active-group visual state under collapse:** all cards render, but inactive cards apply `opacity: 0.45` + `pointer-events: none` on the widget (card still clickable to activate), inactive widget internals appear `disabled`. Active card has full opacity + accent-green border (see §3).
- **Footer change:** the existing single-Cancel footer becomes `[Cancel] [Send all responses]` layout, `justify-content: flex-end; gap: var(--sp-md);`. Send is the primary action so it sits right-most.

### 2. Typography plan

Inherits U6's scale. New roles only:

| Role | Token | Font family | Weight |
|------|-------|-------------|--------|
| `DecisionGroup` heading | `var(--text-data)` 14px | `var(--font-ui)` | 600 |
| `DecisionGroup` prompt | `var(--text-body)` 13px | `var(--font-ui)` | 400, `line-height: 1.5` |
| `DecisionGroup` help text | `var(--text-label)` 11px | `var(--font-ui)` | 400 italic, `color: var(--text-dim)` |
| Collapse banner title | `var(--text-body)` 13px | `var(--font-ui)` | 600, `color: var(--accent-amber)` |
| Collapse banner body | `var(--text-label)` 11px | `var(--font-ui)` | 400, `color: var(--text-dim)` |
| "Required" inline badge | `var(--text-label)` 11px | `var(--font-mono)` | 600 uppercase, `letter-spacing: 0.05em`, `color: var(--accent-green)` |
| `Send all responses` button label | `var(--text-body)` 13px | `var(--font-ui)` | 600 |
| Keyboard-hint footnote ("⌘+Enter to send") | `var(--text-label)` 11px | `var(--font-mono)` | 400, `color: var(--text-muted)` |

The heading → prompt → help hierarchy (14/13/11 px) is the SAME three-tier cascade `FreeTextWidget` already uses for its label → textarea → footer-hint (13/textarea/11) — `DecisionGroup` is effectively one level up that cascade (a heading over a widget shell).

### 3. Color strategy

- **Primary card surface:** `var(--bg-elevated)` — same as `SummaryCard` from U6 and same as `.option-btn` idle state from `ChoiceWidget`. A `DecisionGroup` card should look like "a widget wrapped in a header", not a new surface class.
- **Default card border:** `1px solid var(--border-subtle)`.
- **Active state (when this group owns the turn under collapse rule OR when it is selected as the focused question):** border flips to `var(--accent-green)`, background unchanged. Same idiom as `.option-btn.selected` → `.theme-list-btn.active` pattern already in the app. No glow, no box-shadow — just the border color swap.
- **Inactive state (collapse rule, other groups):** border stays subtle, card `opacity: 0.45`, cursor `pointer` on the card shell (clicking activates it), widget internals receive the widgets' own `disabled` state (opacity 0.6 from existing widget CSS — compounds to ~0.27 which is correct: these are visibly but not confusingly out-of-service).
- **Collapse banner:** reuse U6's `HintBanner warn` tone EXACTLY — background `color-mix(in srgb, var(--accent-amber) 10%, transparent)`, 3px left-stripe `var(--accent-amber)`, icon `var(--accent-amber)`, body `var(--text-primary)` for title and `var(--text-dim)` for instruction. No new banner class.
- **"Required" badge:** a small pill next to the heading. `background: color-mix(in srgb, var(--accent-green) 15%, transparent)`, `color: var(--accent-green)`, `padding: 0 var(--sp-xs)`, `border-radius: var(--radius-sm)`, `font-size: var(--text-label)`. Mirrors the `.round-pill` idiom already used in `InputResponseModal`.
- **`Send all responses` button:** primary action, `background: var(--accent-green)`, `color: var(--bg-deepest)`, no border. Exactly matches `FreeTextWidget.btn-submit` — same padding, radius, font, hover, active, disabled states. Disabled state: `background: var(--accent-green-dim)`, `color: var(--text-muted)`, cursor `not-allowed`.
- **Cancel button:** unchanged from existing `.btn-cancel` in the modal.
- **Text contrast:**
  - Emphasised: heading `var(--text-primary)`, active-group widget label `var(--text-primary)`.
  - Default: prompt copy `var(--text-primary)`, help `var(--text-dim)`.
  - Muted: keyboard hint `var(--text-muted)`, inactive-group help `var(--text-muted)`.

### 4. Interaction model

- **Keyboard nav (spec §6.4):**
  - Tab order: `TranscriptPane` internals → first `DecisionGroup`'s widget → second `DecisionGroup`'s widget → … → `Cancel` button → `Send all responses` button. The Send button is LAST in tab order but first in visual priority — this keeps users moving forward through the form before reaching submit.
  - Inside each widget, the widget's existing focus behavior applies (e.g., `ChoiceWidget` radios, `FreeTextWidget` textarea auto-focus).
  - Under collapse rule, inactive cards are OUT of tab order (`tabindex="-1"` on the card + `disabled` on widget controls).
  - `Escape` closes the whole modal (existing window-level handler).
- **Cmd+Enter (or Ctrl+Enter on non-mac):** anywhere inside the modal triggers `onSend` — equivalent to clicking `Send all responses`. Only fires if the Send button's `disabled` state is false (i.e., every `required: true` group has a value). Shake the modal via the existing `.input-error-shake` keyframe if invoked while required fields are empty — reuse the existing mechanism.
- **Collapse-rule group switch:** clicking any inactive card's card-shell (NOT its disabled widget internals) activates it. Transition: border-color and opacity interpolate over `var(--duration-short)` (100ms) using `var(--ease-move)`. No content repaint — the swap is purely styling.
- **Send button enabled state (reactive):** `disabled = required_keys.some(k => !responses[k])`. For `multi` widgets, the response value is the joined string from the widget; empty string counts as unfilled. For `approval`, no empty state exists (Yes/No always emits); for `free`, `value.trim().length > 0` per existing widget logic.
- **Hover states:**
  - Active card: no hover change (already highlighted).
  - Inactive card: `opacity: 0.45` → `opacity: 0.7` on hover, cursor `pointer`, `transition: opacity var(--duration-short) var(--ease-enter)`.
  - `Send all responses`: inherits `FreeTextWidget.btn-submit` hover (`filter: brightness(1.1)`).
  - `Cancel`: inherits existing `.btn-cancel` hover.
- **Focus ring:** global `:focus-visible` rule (2px `var(--accent-green)`, offset 2px). `Send all responses` uses `outline-offset: 2px` so the ring sits outside the button's solid green background (inside the deepest bg frames it correctly).
- **Submit flow:** on Send, the button briefly disables itself (`sending = true` following the existing `InputResponseModal.onWidgetSubmit` pattern). On backend error, the `.input-error-shake` keyframe fires via the existing `triggerShake()` helper.
- **Snackbar emit on collapse path:** when `onSend` takes the collapse branch (`shape != 'json' && groups.length > 1`), emit a "decision collapsed" snackbar via the existing `NodeInputSnackbarStack` with `tone: 'info'` (DO NOT invent a new `collapsed` tone — the story text suggests one but the existing tone palette is enough; map to `info` with explicit message).

### 5. Component specs

All measurements source tokens. `DecisionGroup` card rhythm ALIGNS to widget shells:

- **`DecisionGroup` card:**
  - `padding: var(--sp-md);` — matches `SummaryCard`.
  - `background: var(--bg-elevated);`.
  - `border: 1px solid var(--border-subtle);` (→ `var(--accent-green)` when active).
  - `border-radius: var(--radius-md);` (4px).
  - `display: flex; flex-direction: column; gap: var(--sp-xs);` for heading/prompt/help, then `margin-top: var(--sp-md)` on widget wrapper.
  - Heading row: `display: flex; align-items: center; gap: var(--sp-sm);` so the "Required" pill sits inline with the heading.
- **Widget shell inside card:** no extra wrapper padding — the widget already owns its own internal spacing (`ChoiceWidget.gap: var(--sp-sm)`, `FreeTextWidget.gap: var(--sp-md)`, etc.). Setting a consistent rhythm: `DecisionGroup` card padding + widget's own internal gap is the spacing contract.
- **Collapse banner:** identical dims to U6's `HintBanner warn` (see U6 §5) — do not diverge.
- **`Send all responses` button:**
  - `min-width: var(--btn-min-wide);` (96px — matches `ApprovalWidget`).
  - `padding: var(--sp-xs) var(--sp-xl);` (same as `FreeTextWidget.btn-submit`).
  - `border-radius: var(--radius-md);`.
  - `gap: var(--sp-xs);` (icon + label).
  - Uses a `Send` Lucide icon (same as `FreeTextWidget.btn-submit`) at 13px.
- **Footer:** `padding: var(--sp-md) var(--sp-lg);` (unchanged from existing `.modal-footer`), `gap: var(--sp-md);` between buttons, `justify-content: flex-end;`.
- **Keyboard hint row:** optional `<div class="footer-hint">⌘+Enter to send</div>` absolute-positioned on the left of the footer (`left: var(--sp-lg)`). Mirrors the `FreeTextWidget.keyboard-hint` idiom so users learn the same convention from Layer-1 to Layer-2.
- **Required badge pill:** `padding: 0 var(--sp-xs); border-radius: var(--radius-sm); font-size: var(--text-label);`.
- **Inactive card transition:** `transition: opacity var(--duration-short) var(--ease-move), border-color var(--duration-short) var(--ease-move);`.

### 6. Signature elements

- **Cards-as-widgets.** Every `DecisionGroup` looks like an `inputWidgets/` shell with a headline prepended. Same padding, same radius, same border, same bg. A user who has completed one interactive prompt already knows how to complete a `DecisionGroup` — the mental model compounds.
- **Green-border active state.** Under collapse rule, "which question is live" is communicated solely by the border color swap from `var(--border-subtle)` to `var(--accent-green)` — no scale, no box-shadow, no glow. The restraint is the signature: a single-pixel color change does all the work.
- **Opacity-only inactive treatment.** Inactive cards fade to 0.45 rather than grey-out text. Preserves the type hierarchy so the user can still READ the inactive question — they just can't answer it. This is the Bloomberg-Terminal density trade-off in action.
- **`Send all responses` = same button as `FreeText`.** Users should not notice that submission is structurally different between Layer-1 and Layer-2 — the button is identical in every pixel. Continuity over novelty.
- **Collapse banner uses the already-seen `HintBanner warn`.** The user has already parsed this banner shape in the passive AST nodes (U6) — reusing it when the UI has to say "you can only pick one" inherits all that parse cost for free.
- **⌘+Enter hint lives under the Send button.** Reinforces the keyboard-first ethos from `DESIGN.md` and matches `FreeTextWidget`'s in-widget footer placement. The same shortcut works whether the widget is Layer-1 or Layer-2.

### Cross-story design coherence (binding rules)

- **`DecisionGroup` card padding / radius / border MUST equal `inputWidgets/`-widget internal shell rhythm.** If `ChoiceWidget` adjusts its padding, `DecisionGroup` adjusts in lockstep. The two are a system.
- **Collapse banner reuses U6 `HintBanner warn` tone unchanged.** If U6 renames a token, U7 inherits. Do NOT hard-code amber values.
- **`Send all responses` button = `FreeTextWidget.btn-submit` rules verbatim.** Copy, don't re-invent.
- **"Required" pill = `.round-pill` idiom from `InputResponseModal`, color-swapped to green.** Reuse the alpha-tint pill pattern, swap `accent-amber` → `accent-green`.
- **Snackbar tone on collapse = `info`, not a new `collapsed` variant.** The existing `NodeInputSnackbarStack` tone palette is sufficient.
- **No new tokens, ever.** Every measurement traces to `style.css`. If a new need emerges (e.g., an "active-card green ring width"), it is `border-color` swap only — no extra ring dimension token.

