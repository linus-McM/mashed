# bmad-interactive-06: Frontend modal + snackbar + shape widgets

**Status:** ready
**Domain:** frontend
**Size:** L
**Depends On:** bmad-interactive-03 (event contract)
**Priority:** P0-critical
**Phase 0d design brief required:** YES — spawn ui-architect before any Svelte code lands. The five widget shapes (free / choice / multi / approval / file / json) need coordinated visual language + motion with the existing BMAD canvas palette (ProcessNode amber badge, QuestionSnackbarStack). Do not improvise; ship the brief first.

## Story

As a Mashed user answering a BMAD interactive prompt, I want a modal whose widget matches the input shape (textarea / radio / checkboxes / approval / file picker / JSON editor) and a snackbar that announces new prompts with a round indicator, so that I can answer without reading tmux and without hunting through the canvas for the right node.

## Description

Replace `QuestionResponseModal.svelte` with `InputResponseModal.svelte` — the new modal switches widget per `PendingPrompt.Shape`. Rename `QuestionSnackbarStack.svelte` to `NodeInputSnackbarStack.svelte` and add a round indicator + `Skip` button for optional prompts. Wire `WorkflowBuilder.svelte` to the new event contract from S3/S4 (`bmad:node:awaiting_input`, `bmad:node:input_resolved`, `bmad:node:round_complete`, `bmad:node:gate_satisfied`, `bmad:node:round_limit`, `bmad:node:aborted`, `bmad:node:input_invalid`). Update `ProcessNode.svelte` to render the `awaiting_input` badge (amber, pulsing, icon `message-circle-question`) and the `"N/M rounds"` subtitle.

### Scope summary
- New component `InputResponseModal.svelte` with five shape widgets (§9.2).
- Rename `QuestionSnackbarStack.svelte` → `NodeInputSnackbarStack.svelte` (§9.3); keep legacy export alias for one release to avoid breaking `App.svelte` imports if the rename lands separately.
- `ProcessNode.svelte`: `awaiting_input` status badge + round counter subtitle (§9.1).
- `WorkflowBuilder.svelte`: event subscriptions per §9.4.
- `RespondToInput` Wails binding call from the modal; legacy `RespondToQuestion` stays for autonomous fallback.
- Validation error handling: `bmad:node:input_invalid` → inline shake animation, modal stays open (§9.2).
- Playwright AC tests per shape (§12.4).

### Non-goals
- No Monaco bundle upgrade (use existing editor setup if present; fall back to `<textarea>` with JSON validation if Monaco not already wired).
- No registry-driven dynamic options UI changes beyond accepting `PendingPrompt.Options` as-is.
- No resume UI beyond reacting to the re-emitted `awaiting_input` event (S5 supplies it).

## Developer Notes

### Files to create/modify
- `frontend/src/components/bmad/InputResponseModal.svelte` (new).
- `frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte` (new).
- `frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte` (new).
- `frontend/src/components/bmad/inputWidgets/MultiChoiceWidget.svelte` (new).
- `frontend/src/components/bmad/inputWidgets/ApprovalWidget.svelte` (new).
- `frontend/src/components/bmad/inputWidgets/FileInputWidget.svelte` (new).
- `frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte` (new).
- `frontend/src/components/bmad/NodeInputSnackbarStack.svelte` (renamed from `QuestionSnackbarStack.svelte`; keep `QuestionSnackbarStack.svelte` as a re-export shim for one release).
- `frontend/src/components/bmad/ProcessNode.svelte` (modify — add awaiting status + round counter).
- `frontend/src/views/WorkflowBuilder.svelte` (modify — new event subscriptions).
- Test files:
  - `frontend/src/components/bmad/InputResponseModal.test.ts` (Vitest).
  - `frontend/src/components/bmad/inputWidgets/*.test.ts` (one per widget).
  - `tests/ac/bmad-input-free.spec.ts`, `bmad-input-choice.spec.ts`, `bmad-input-multi.spec.ts`, `bmad-input-approval.spec.ts`, `bmad-input-file.spec.ts`, `bmad-input-json.spec.ts` (Playwright).

### Event contract the UI subscribes to (from S3/S4)
| Event | Action |
|---|---|
| `bmad:node:awaiting_input` | Push snackbar for node; queue up if multiple. Modal opens if user clicks snackbar or node card. |
| `bmad:node:input_resolved` | Dismiss the snackbar + close the modal if open for this `(nodeId, inputId)`. |
| `bmad:node:input_invalid` | Inline error on modal with shake animation; modal stays open. |
| `bmad:node:round_complete` | Update round counter subtitle on `ProcessNode`. |
| `bmad:node:gate_satisfied` / `bmad:node:round_limit` | Transitional flash from `awaiting_input` badge → `complete` badge on the node. |
| `bmad:node:aborted` | Dismiss snackbar + modal; flash node red → failed badge. |

### Wails binding call
```svelte
import { RespondToInput } from '../../wailsjs/go/main/App.js';
...
await RespondToInput(prompt.execId, prompt.nodeId, prompt.inputId, value);
```

Distinguish interactive (`RespondToInput`) from legacy (`RespondToQuestion`) based on node status: `NodeAwaitingInput` → interactive; `NodeRunning` with legacy `bmad:node:question` fired → legacy.

### Widget spec details (§9.2)
| Shape | Widget | Submit |
|---|---|---|
| `free` | `<textarea>` with `maxlength` from `spec.MaxLength` | Cmd+Enter or `Submit` button |
| `choice` | `<div role="radiogroup">` per option | `Enter` after selection |
| `multi` | `<div>` of `<input type="checkbox">` | `Submit` when ≥ 1 checked |
| `approval` | Two buttons `Yes` / `No`, sends `"yes"` / `"no"` literal | Click |
| `file` | Drop zone + text input; sends absolute or repo-relative path | Enter |
| `json` | Monaco mini-editor if available, else `<textarea>`; validates via `JSON.parse` before send | `Submit` |

### Design tokens
Defer visual spec to the ui-architect brief; for backend-shipped scaffolding:
- Use existing `--color-amber-*` token for pulsing awaiting badge.
- Reuse `QuestionSnackbarStack`'s existing layout for the snackbar carousel.
- Round indicator: small monospace subtitle, e.g. `3 / 30`, same typography scale as existing node sub-line.

### Shake animation
CSS class `input-error-shake` with a 250 ms keyframe translation. Trigger by toggling a `validationError` store value in response to `bmad:node:input_invalid`.

### Risks / gotchas
- **File widget + backend path validation**: the `resolveFileInput` security check in S3 is server-side. The frontend need NOT restrict the path widget to repo-relative paths at input time — backend rejects out-of-bounds paths and emits `input_invalid`. Show the returned reason verbatim.
- **JSON widget**: if Monaco is not already in the bundle, do not add it here. Use a monospace `<textarea>` with inline parse-error tooltip via try/catch on `JSON.parse`.
- **Legacy snackbar coexistence**: during the rollout window, both `QuestionSnackbarStack` (autonomous) and `NodeInputSnackbarStack` (interactive) may fire. Merge them into a single visual stack — otherwise the user sees two snackbars for two different code paths. Simplest approach: keep one component, dispatch on `question.type === "interactive" | "legacy"`.
- **Event payload shapes**: for `bmad:node:awaiting_input`, the payload is the full `PendingPrompt` struct (NodeID, InputID, Prompt, Shape, Options, Round, CreatedAt, PromptID). `bmad:node:input_resolved` carries `{execId, nodeId, inputId, round, valueHash}` — hash only, no raw value (§14.3). Do NOT try to display the resolved value.
- **Stale prompts**: if a `bmad:node:round_limit` arrives while the modal is open for the matching node, close the modal with a toast `"That prompt already expired."` (§13.2).
- **Screen reader**: the snackbar uses `role="status" aria-live="polite"` (existing pattern). The modal trap focus via existing `Modal.svelte` shell — do not reinvent.

### Reference files
- `frontend/src/components/bmad/QuestionResponseModal.svelte` — existing modal pattern with validation + RespondToQuestion binding.
- `frontend/src/components/bmad/QuestionSnackbarStack.svelte` — existing snackbar carousel.
- `frontend/src/components/bmad/ProcessNode.svelte` — status rendering; adds `awaiting_input` case.
- `frontend/src/views/WorkflowBuilder.svelte` — event subscription wiring; shows existing question-event pattern to extend.
- `.wolf/reframe-frameworks.md` is unrelated — frontend is Svelte, do not migrate.
- Use the `/xyflow` skill for any canvas/ProcessNode rework.
- Use the `/playwright-cli` skill for the AC specs.

## Acceptance Criteria

**AC-1: Snackbar appears on `bmad:node:awaiting_input` and carries round indicator**
- Given the frontend subscribed to `bmad:node:awaiting_input`
- When an event fires with payload `{NodeID:"n1", InputID:"topic", Shape:"free", Round:3, Prompt:"What next?", Options:[]}`
- Then a snackbar appears within 200 ms
- And its subtitle shows `Round 3`
- And it carries a `Respond` button
- And (because `Required` was true in the event payload) no `Skip` button is shown

**AC-2: Modal renders correct widget per shape**
- Given an awaiting prompt with `Shape = "choice"` and `Options = ["a","b","c"]`
- When the user opens the modal
- Then a radiogroup is rendered with three options
- And `Enter` after selecting `"b"` calls `RespondToInput(execId, nodeId, inputId, "b")`
- Given a prompt with `Shape = "approval"`
- Then two buttons labelled `Yes` and `No` render
- And clicking `Yes` sends `"yes"` via `RespondToInput`
- Given a prompt with `Shape = "multi"` and `Options = ["x","y","z"]`
- Then three checkboxes render
- And `Submit` sends a comma-joined value `"x,z"` for the user's selection

**AC-3: `bmad:node:input_invalid` triggers shake + keeps modal open**
- Given the modal is open for a `choice` prompt with `Options = ["a","b"]`
- When the user submits `"z"` and the backend emits `bmad:node:input_invalid` with reason `"value must be one of [a b]"`
- Then the modal does NOT close
- And an inline error containing the reason renders
- And the modal root gains the `input-error-shake` class for 250 ms
- And the shake class is removed afterwards

**AC-4: `bmad:node:input_resolved` closes the snackbar + modal**
- Given an open modal and visible snackbar for `(n1, topic)`
- When `bmad:node:input_resolved` fires with `{nodeId:"n1", inputId:"topic"}`
- Then the snackbar dismisses
- And the modal closes
- And the resolved event's `valueHash` is never rendered to the DOM

**AC-5: `ProcessNode` renders awaiting badge + round counter**
- Given a node whose status is `awaiting_input` and whose execution's `NodeRounds[nodeID] = 3` with a process whose `Gate.MaxRounds = 30`
- When `ProcessNode.svelte` renders
- Then an amber pulsing badge with icon `message-circle-question` is visible
- And a subtitle containing `3 / 30` is visible
- And clicking the node opens the modal

**AC-6: `round_complete` updates the counter without closing the modal**
- Given the modal open and current round indicator `2 / 30`
- When `bmad:node:round_complete` fires with round=3
- Then the `ProcessNode` subtitle updates to `3 / 30`
- And the modal stays in its current state (if still open)

**AC-7: `gate_satisfied` and `round_limit` transition the node to complete visually**
- Given a node in `awaiting_input` state
- When `bmad:node:gate_satisfied` fires for that node
- Then the badge transitions from amber pulsing to the existing `complete` badge within 400 ms
- And the snackbar dismisses
- Given a different node in `awaiting_input`
- When `bmad:node:round_limit` fires
- Then the same visual transition happens
- And a toast `"Reached round limit — process ended"` appears briefly

**AC-8: `NodeInputSnackbarStack` is aria-live polite and keyboard-reachable**
- Given a visible snackbar
- Then its root element has `role="status"` and `aria-live="polite"`
- And `Tab` focus reaches the `Respond` button in one step from the document
- And pressing `Enter` on the focused button opens the modal

## BDD Test Scenarios

```gherkin
Feature: Interactive input modal and snackbar

  Scenario: Awaiting event shows snackbar with round indicator
    Given WorkflowBuilder subscribed to bmad events
    When bmad:node:awaiting_input fires with Round 3 and Shape "free"
    Then a snackbar appears within 200ms
    And the snackbar subtitle contains "Round 3"

  Scenario: Choice shape renders radiogroup
    Given a PendingPrompt with Shape "choice" and Options ["a","b","c"]
    When the user opens the modal
    Then a radiogroup with three options is visible
    When the user selects "b" and presses Enter
    Then RespondToInput is invoked with value "b"

  Scenario: Approval shape sends yes or no literal
    Given a PendingPrompt with Shape "approval"
    When the user clicks the "Yes" button
    Then RespondToInput is invoked with value "yes"
    When the user clicks "No" in a fresh prompt
    Then RespondToInput is invoked with value "no"

  Scenario: Multi shape sends comma-joined list
    Given a PendingPrompt with Shape "multi" and Options ["x","y","z"]
    When the user checks "x" and "z" and clicks Submit
    Then RespondToInput is invoked with value "x,z"

  Scenario: File shape sends literal path and shows backend error
    Given a PendingPrompt with Shape "file"
    When the user enters "/etc/passwd" and presses Enter
    And bmad:node:input_invalid fires with reason "path outside repository root"
    Then the modal stays open
    And the inline error text shows "path outside repository root"
    And the modal root gains class input-error-shake for 250ms

  Scenario: JSON shape rejects invalid JSON before send
    Given a PendingPrompt with Shape "json"
    When the user enters "{not-json}" and clicks Submit
    Then RespondToInput is NOT invoked
    And an inline parse error renders

  Scenario: Input resolved closes modal and snackbar
    Given an open modal and visible snackbar for (n1, topic)
    When bmad:node:input_resolved fires for (n1, topic)
    Then the snackbar dismisses
    And the modal closes
    And no valueHash is rendered in the DOM

  Scenario: Round complete updates ProcessNode subtitle
    Given a ProcessNode whose subtitle reads "2 / 30"
    When bmad:node:round_complete fires with round 3
    Then the subtitle updates to "3 / 30"

  Scenario: Gate satisfied flashes node to complete
    Given a ProcessNode in awaiting_input
    When bmad:node:gate_satisfied fires for that node
    Then the badge transitions to the complete style within 400ms
    And the snackbar dismisses

  Scenario: Snackbar is keyboard reachable and aria-live
    Given a visible snackbar
    Then the snackbar root has role "status" and aria-live "polite"
    When the user presses Tab from the document root
    Then the Respond button receives focus within one Tab press
```

## Tasks / Subtasks

- [ ] Task 0: Wait for Phase 0d design brief (blocking)
  - [ ] ui-architect produces a design brief for modal + snackbar + awaiting badge
  - [ ] Brief includes token usage (amber pulse, focus ring), motion timings, typography scale, empty states
  - [ ] Do not begin Svelte work until brief is approved
- [ ] Task 1: Event wiring in `WorkflowBuilder.svelte` (AC-1, AC-4, AC-6, AC-7)
  - [ ] Subscribe to all six new events via `EventsOn` from wails runtime
  - [ ] Route events to a single `InteractiveStore` keeping `{pendingPrompts, rounds, lastToast}`
  - [ ] Wire existing stores so `ProcessNode` re-renders on round changes
- [ ] Task 2: Snackbar rename + round indicator (AC-1, AC-8)
  - [ ] Rename `QuestionSnackbarStack.svelte` → `NodeInputSnackbarStack.svelte`
  - [ ] Add round subtitle when payload carries `Round > 0`
  - [ ] Add `Skip` button only when `Required == false` (passed through from payload)
  - [ ] Keep legacy question stack interop (merge rendered list by `(nodeId, kind)` key)
- [ ] Task 3: `InputResponseModal.svelte` shell (AC-2, AC-3, AC-4)
  - [ ] Dispatch on `Shape` to the right widget component
  - [ ] Subscribe to `input_invalid` for the currently-open prompt; render inline error with shake
  - [ ] Call `RespondToInput` on widget submit
- [ ] Task 4: Shape widgets (AC-2)
  - [ ] `FreeTextWidget`, `ChoiceWidget`, `MultiChoiceWidget`, `ApprovalWidget`, `FileInputWidget`, `JsonInputWidget`
  - [ ] Each has its own vitest with at least 3 cases (render, valid submit, keyboard access)
- [ ] Task 5: `ProcessNode.svelte` awaiting badge + round counter (AC-5, AC-6, AC-7)
  - [ ] Render amber pulsing badge for `awaiting_input`
  - [ ] Render `{round} / {maxRounds}` subtitle when round > 0
  - [ ] Transition badge to `complete` on gate_satisfied / round_limit via CSS class
- [ ] Task 6: Playwright AC tests (AC-1..AC-7)
  - [ ] One spec per shape in `tests/ac/bmad-input-*.spec.ts`
  - [ ] Smoke spec for snackbar + modal end-to-end
  - [ ] Use `/playwright-cli` skill for setup
- [ ] Task 7: Vitest coverage for shared store + helpers
  - [ ] `InteractiveStore` behaviour
  - [ ] Event→snackbar mapping

## Definition of Done

- [ ] ui-architect design brief approved and referenced in the PR
- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests (Vitest + Playwright)
- [ ] 80%+ coverage on new components under `frontend/src/components/bmad/`
- [ ] `go build ./...` passes (no backend impact expected)
- [ ] `npm run test` / frontend test suite passes
- [ ] Playwright AC specs green
- [ ] `/simplify` run on all modified Svelte/TS files; no CRITICAL/HIGH findings
- [ ] No `valueHash` rendered anywhere in DOM
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch

## Design Brief

> Author: ui-architect (Phase 0d). Source palette: `frontend/src/style.css`. Reference aesthetic: `DESIGN.md` ("Calm command center. Linear meets Bloomberg Terminal"). All values MUST resolve to existing CSS custom properties or the **PROPOSED** tokens listed at the bottom. No magic numbers.

### 1. Layout composition

`InputResponseModal.svelte` shell (inherits the `QuestionResponseModal` geometry deliberately — the existing shell is strong; only the body changes).

- **Overlay**: `position: fixed; inset: 0`, background `var(--overlay-backdrop)`, z-index `400`, padding `var(--sp-2xl)` so the card never kisses the window edge on resize.
- **Card dimensions**:
  - `max-width: 560px` for `free / choice / multi / approval`
  - `max-width: 720px` for `file / json` (PROPOSED `--modal-width-wide: 720px`) — JSON and file picker need horizontal room
  - `width: 100%`; `min-width: 360px` (PROPOSED `--modal-width-min: 360px`)
  - `max-height: 80vh`; on `json` shape `min-height: 480px` (PROPOSED `--modal-height-editor: 480px`)
  - `background: var(--bg-surface)`, `border: 1px solid var(--border-emphasis)`, `border-radius: var(--radius-lg)`
  - `box-shadow: 0 16px 48px color-mix(in srgb, var(--bg-deepest) 80%, transparent)`
- **Header**: `padding: var(--sp-md) var(--sp-lg)`, border-bottom `1px solid var(--border-subtle)`. Contains: amber `MessageCircleQuestion` icon, title `"Input from <repo>"`, **round pill** (see §6), close button. Header is `flex-shrink: 0`.
- **Body**: `padding: var(--sp-lg)`, `gap: var(--sp-lg)` between prompt block + widget + helpText. `overflow-y: auto; flex: 1`.
- **Footer**: `padding: var(--sp-md) var(--sp-lg)`, border-top `1px solid var(--border-subtle)`, `flex-shrink: 0`. Right-aligned `Cancel` + primary submit. For `approval` shape the footer is hidden (buttons are the widget).
- **Inline error bar** (above footer): `padding: var(--sp-sm) var(--sp-lg)`, `background: color-mix(in srgb, var(--accent-red) 10%, transparent)`, `border-top: 1px solid color-mix(in srgb, var(--accent-red) 30%, transparent)`.
- **Snackbar stack**: unchanged position — `position: fixed; top: var(--sp-lg); right: var(--sp-lg); z-index: 300; width: 320px`. Modal (z-400) sits over snackbar (z-300) over canvas (default). Toast layer (`round_limit` / `aborted` / `stale prompt`) uses PROPOSED `--z-toast: 500` to sit above the modal.

### 2. Typography plan

| Slot | Font | Size | Weight | Notes |
|---|---|---|---|---|
| Modal title | `--font-ui` | `var(--text-body)` (13px) | 600 | repo-name span inherits `--font-mono` + `--accent-green` per existing pattern |
| Prompt body (`.prompt-block`) | `--font-ui` | `var(--text-data)` (14px) | 400 | `line-height: 1.6`, `white-space: pre-wrap` |
| `helpText` subtitle | `--font-ui` | `var(--text-label)` (11px) | 400 | `color: var(--text-dim)`, italic, shown below widget |
| Option label (choice/multi) | `--font-ui` | `var(--text-body)` | 500 | |
| Option number prefix | `--font-mono` | `var(--text-label)` | 400 | `color: var(--text-dim)`, `width: 20px` (tabular) |
| Validation error | `--font-ui` | `var(--text-body)` | 500 | `color: var(--accent-red)` |
| Round pill | `--font-mono` | `var(--text-label)` | 600 | `font-variant-numeric: tabular-nums`, uppercase letter-spacing `0.04em` |
| Keyboard hint (`Cmd+Enter`) | `--font-mono` | `var(--text-label)` | 400 | `color: var(--text-muted)`, right-aligned |
| JSON editor content | JetBrains Mono via PROPOSED `--font-code: 'JetBrains Mono', monospace` (already referenced in `DESIGN.md` but not yet a token) | 13px | 400 | monospace required for JSON shape |

### 3. Color strategy

| Element | Token |
|---|---|
| `awaiting_input` badge background | `color-mix(in srgb, var(--accent-amber) 15%, transparent)` |
| `awaiting_input` badge border/fg | `var(--accent-amber)` |
| Amber pulse mid-key color | `color-mix(in srgb, var(--accent-amber) 100%, transparent)` → `35%` at valley (see §7) |
| Amber glow | `var(--glow-spread) color-mix(in srgb, var(--accent-amber) 60%, transparent)` |
| Primary submit button | `background: var(--accent-green)`, `color: var(--bg-deepest)` (keeps the "alive / go" semantic) |
| Disabled submit | `background: var(--accent-green-dim)`, `color: var(--text-muted)` |
| Destructive (reject token) | `border: 1px solid color-mix(in srgb, var(--accent-red) 40%, transparent)`, text `var(--accent-red)` |
| Approval **Yes** | matches primary submit (green) |
| Approval **No** | ghost-button: transparent bg, `border: 1px solid var(--border-emphasis)`, text `var(--text-dim)` (not red — "no" is not destructive, just a vote) |
| Reject-token button (if `Gate.RejectTokens` supplied via snackbar) | destructive red variant |
| Inline validation error text | `var(--accent-red)` |
| Validation error bar bg | `color-mix(in srgb, var(--accent-red) 10%, transparent)` |
| Checkbox / radio selected bg | `color-mix(in srgb, var(--accent-green) 15%, transparent)` |
| Checkbox / radio selected border | `var(--accent-green)` |
| Checkbox check glyph | `var(--accent-green)` |
| Monaco theme | `vs-dark` with overrides: `editor.background` = `--bg-deepest`, `editor.foreground` = `--text-primary`, `editor.lineHighlightBackground` = `--bg-elevated`, `editorGutter.background` = `--bg-surface`, `editorError.foreground` = `--accent-red`. Fallback `<textarea>`: `background: var(--bg-deepest)`, `font-family: var(--font-code)`. |

### 4. Interaction model

Focus trap is inherited from existing `QuestionResponseModal` pattern (implement via a `modalEl.querySelectorAll('button,[href],input,textarea,select,[tabindex]:not([tabindex="-1"])')` helper — same pattern already in repo).

| Shape | Primary submit | Autofocus | Tab order | Esc behavior |
|---|---|---|---|---|
| `free` | `Cmd/Ctrl+Enter` OR `Submit` button | `<textarea>` | textarea → submit → cancel → close | Close modal (cosmetic only — `NodeAwaitingInput` persists per §13.6) |
| `choice` | `Enter` after selection; or click option | first option `<button role="radio">` | options (arrow keys cycle, j/k alt) → cancel → close | Close modal; prompt persists |
| `multi` | `Submit` button (disabled until ≥1 checked) | first checkbox | checkboxes → submit → cancel → close | Close modal; persists |
| `approval` | Click `Yes` / `No` | `Yes` button | Yes → No → close | Close modal; persists |
| `file` | `Enter` on path input, or drop file, or click Submit | path `<input>` | drop-zone → path input → submit → cancel | Close modal; persists |
| `json` | `Cmd/Ctrl+Enter` OR `Submit` | Monaco editor (or textarea fallback) | editor → submit → cancel | Close modal; persists |

**Esc MUST NOT resolve `NodeAwaitingInput`** — it only closes the modal UI. The snackbar remains to re-invite.

**Shake on invalid** (triggered by `bmad:node:input_invalid`):
- class `.input-error-shake` applied to `.modal-card`
- duration: `250ms` (PROPOSED `--duration-shake: 250ms`)
- easing: `var(--ease-move)` (`ease-in-out`)
- amplitude: `±6px` horizontal translation over 4 peaks
- keyframes: `0% { translate: 0 } 20% { translate: -6px } 40% { translate: 6px } 60% { translate: -4px } 80% { translate: 4px } 100% { translate: 0 }`
- class removed after animation via `animationend` listener. Wrapped in `@media (prefers-reduced-motion: no-preference)` — reduced-motion users see the red error bar without shake.

### 5. Component specs (per shape)

Shared widget chrome: each widget is a Svelte component under `frontend/src/components/bmad/inputWidgets/`. Each receives `{ prompt: PendingPrompt, disabled: boolean }` and emits `submit` with the string value the backend expects, or `cancel`.

#### 5.1 `FreeTextWidget.svelte`
- DOM: `<label><textarea/></label><span class="keyboard-hint">Cmd+Enter to send</span>`
- Textarea: `min-height: 72px; max-height: 200px; resize: vertical; padding: var(--sp-sm) var(--sp-md); border-radius: var(--radius-md); border: 1px solid var(--border-subtle); background: var(--bg-deepest); font: var(--font-ui) / 1.5; font-size: var(--text-body)`
- States: `:focus { border-color: var(--accent-green) }`; `[aria-invalid=true] { border-color: var(--accent-red) }`; `:disabled { opacity: 0.6 }`
- ARIA: `aria-label={prompt.prompt}`, `aria-required={spec.required}`, `aria-invalid={!!validationError}`, `maxlength={spec.maxLength || 5000}`
- Counter (if `MaxLength`): bottom-right monospace, `color: var(--text-muted)`, turns `--accent-amber` at 90%, `--accent-red` at 100%

#### 5.2 `ChoiceWidget.svelte`
- DOM: `<div role="radiogroup" aria-label={prompt.prompt} aria-required="true">` containing `<button role="radio" aria-checked={selected === i}>` per option
- Button: `display: flex; padding: var(--sp-sm) var(--sp-md); border-radius: var(--radius-md); gap: var(--sp-md)`
- Default: `background: var(--bg-elevated); border: 1px solid var(--border-subtle); color: var(--text-primary)`
- Hover: `background: var(--bg-active); border-color: var(--border-emphasis)`
- Focus-visible: `outline: 2px solid var(--accent-green); outline-offset: -1px`
- **Selected**: `background: color-mix(in srgb, var(--accent-green) 15%, transparent); border-color: var(--accent-green)`; left-edge chevron glyph `›` in `--accent-green` at `var(--sp-xs)` from the option-num (signature element — echoes node card phase-bar accent)
- Disabled: `opacity: 0.6; cursor: not-allowed`
- Keyboard: Arrow / j/k cycle (same helper as existing modal); `Enter` submits

#### 5.3 `MultiChoiceWidget.svelte`
- DOM: `<div role="group" aria-label={prompt.prompt} aria-required={spec.required}>` with native `<input type="checkbox" id={...}>` + `<label for={...}>`
- Row spacing: `gap: var(--sp-sm)` between rows
- Checkbox: visually replaced — hide native with `opacity: 0; position: absolute`. Render `::before` on label as `12px × 12px` square, `border: 1px solid var(--border-emphasis); border-radius: var(--radius-sm); background: var(--bg-deepest)`
- Checked state: `background: color-mix(in srgb, var(--accent-green) 15%, transparent); border-color: var(--accent-green)`; checkmark `Check` lucide icon at 10px in `--accent-green`
- Submit enabled when ≥1 checked; sends comma-joined `"x,z"` per AC-2
- `aria-invalid` on the `role="group"` when invalid

#### 5.4 `ApprovalWidget.svelte`
- DOM: `<div class="approval-row">` with two buttons `Yes` / `No`
- Layout: `display: flex; gap: var(--sp-md); justify-content: flex-end` inside body. Footer hidden.
- Yes button: primary-submit styling (see §3)
- No button: ghost — `background: transparent; border: 1px solid var(--border-emphasis); color: var(--text-dim); padding: var(--sp-xs) var(--sp-xl); border-radius: var(--radius-md)`; hover `color: var(--text-primary); background: var(--bg-elevated)`
- Both buttons `min-width: 96px` (PROPOSED `--btn-min-wide: 96px`) for visual balance
- Sends literal `"yes"` / `"no"` per AC-2 and §8.4

#### 5.5 `FileInputWidget.svelte`
- DOM:
  ```
  <div class="drop-zone" role="button" tabindex="0">
    <Upload size=20 />
    <span>Drop a file or click to choose</span>
    <span class="path-hint">Paths outside the repo root will be rejected</span>
  </div>
  <input type="text" class="path-input" placeholder="/absolute/or/repo-relative/path" />
  ```
- Drop zone: `padding: var(--sp-xl) var(--sp-lg); border: 1px dashed var(--border-emphasis); border-radius: var(--radius-md); background: var(--bg-deepest); text-align: center; gap: var(--sp-sm); flex-direction: column; display: flex; align-items: center`
- Drop zone states:
  - hover / focus: `border-color: var(--accent-green); background: color-mix(in srgb, var(--accent-green) 5%, transparent)`
  - drag-over: `border-style: solid; border-color: var(--accent-green); background: color-mix(in srgb, var(--accent-green) 10%, transparent)`
  - invalid after backend rejection: `border-color: var(--accent-red)`
- Path input: same styling as FreeText textarea but `height` auto / single line
- Repo-root constraint hint: `<span class="path-hint">` in `var(--font-mono)` / `var(--text-label)` / `var(--text-muted)`. Backend rejection reason renders verbatim in the inline error bar (per §13.9 of schema).
- Frontend does NOT preempt path validation — backend is source of truth (per story risks/gotchas and §14.2).

#### 5.6 `JsonInputWidget.svelte`
- Primary: Monaco editor container, `height: var(--modal-height-editor)` (480px PROPOSED), `border: 1px solid var(--border-subtle); border-radius: var(--radius-md); overflow: hidden`
- Fallback (Monaco absent): `<textarea>` with `font-family: var(--font-code); font-size: 13px; min-height: 320px; tab-size: 2; white-space: pre; overflow: auto`
- Gutter: Monaco renders markers from a local `JSON.parse` try/catch; error squiggle on the offending line in `--accent-red`
- Below editor: `<div class="parse-status">` — valid state shows `✓ valid JSON` in `--accent-green`, invalid shows parse error in `--accent-red`, mono font, `--text-label`
- Submit disabled while parse-invalid (matches story: "rejects invalid JSON before send")

### 6. Snackbar redesign (`NodeInputSnackbarStack.svelte`)

Rename of `QuestionSnackbarStack.svelte`. Keep `QuestionSnackbarStack.svelte` as a one-line re-export shim for one release:

```svelte
<!-- QuestionSnackbarStack.svelte (legacy alias) -->
<script>import Stack from './NodeInputSnackbarStack.svelte';</script>
<Stack {...$$props} on:navigate on:dismiss />
```

Card structure:

```
[accent stripe] [icon] [repo-name] [round pill]  [timestamp]
                [prompt preview (1 line, truncated)]
                [Respond]  [Skip?]
```

- **Round pill**: `background: color-mix(in srgb, var(--accent-amber) 15%, transparent); color: var(--accent-amber); padding: 0 var(--sp-xs); border-radius: var(--radius-sm); font-family: var(--font-mono); font-size: var(--text-label); font-weight: 600; font-variant-numeric: tabular-nums`. Text: `"3 / 30"` (spaces around slash) when `Round > 0 && Gate.MaxRounds > 0`; `"Round 3"` when `Round > 0 && MaxRounds == 0`; hidden otherwise.
- **Prompt preview**: truncate at 80 chars with ellipsis, single line, `font-size: var(--text-body); color: var(--text-primary)`
- **Respond** button: `background: color-mix(in srgb, var(--accent-amber) 15%, transparent); color: var(--accent-amber); border: 1px solid color-mix(in srgb, var(--accent-amber) 40%, transparent); padding: var(--sp-2xs) var(--sp-sm); border-radius: var(--radius-sm); font-size: var(--text-label); font-weight: 600`. Hover: border `--accent-amber` solid, subtle glow.
- **Skip** button (only when `Required == false`): ghost — `background: transparent; color: var(--text-dim); border: 1px solid var(--border-subtle); padding: var(--sp-2xs) var(--sp-sm); border-radius: var(--radius-sm); font-size: var(--text-label)`. Calls `RespondToInput` with empty string (backend treats as "skip" when `!Required`).
- **Stack behavior**: existing `partitionForDisplay` semantics (max visible, overflow indicator). Legacy + interactive prompts merge into a single stack keyed by `(nodeId, kind)` per story's legacy-coexistence risk.
- **Dismissal**: X button or click-outside is **cosmetic only**. `NodeAwaitingInput` persists (§13.6). Re-mount when the user clicks the ProcessNode. On `bmad:node:input_resolved` / `gate_satisfied` / `round_limit` / `aborted` the card auto-dismisses with `fly` transition (x=320).
- ARIA: `role="status" aria-live="polite"` on the stack root (unchanged). Each card `role="button" tabindex="0"`. Respond button inside card is in the natural Tab order (AC-8 guarantees one-Tab reachability).

### 7. ProcessNode badge (`awaiting_input`)

Extend `ProcessNode.svelte` status row with a new branch:

```svelte
{:else if status === 'awaiting_input'}
  <span class="awaiting-badge" aria-label="Awaiting input">
    <MessageCircleQuestion size={11} />
  </span>
  <span class="status-text awaiting-text">awaiting</span>
```

**Badge CSS**:

```css
.awaiting-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  border-radius: 50%;
  background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
  color: var(--accent-amber);
  flex-shrink: 0;
}
.awaiting-text { color: var(--accent-amber); }
@media (prefers-reduced-motion: no-preference) {
  .process-node.awaiting .awaiting-badge {
    animation: awaiting-pulse 1800ms var(--ease-move) infinite;
  }
  .process-node.awaiting {
    animation: awaiting-node-pulse 1800ms var(--ease-move) infinite;
  }
}
@keyframes awaiting-pulse {
  0%, 100% {
    box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent-amber) 0%, transparent);
    background: color-mix(in srgb, var(--accent-amber) 15%, transparent);
  }
  50% {
    box-shadow: var(--glow-spread) color-mix(in srgb, var(--accent-amber) 60%, transparent);
    background: color-mix(in srgb, var(--accent-amber) 30%, transparent);
  }
}
@keyframes awaiting-node-pulse {
  0%, 100% { box-shadow: 0 0 0 0 color-mix(in srgb, var(--accent-amber) 0%, transparent); }
  50% { box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-amber) 25%, transparent); }
}
```

Non-harmonic 1800ms cadence prevents sync with existing 2000ms `agent-running-pulse`.

**Round counter subtitle**: new `.round-counter` row above `.status-row`, visible when `data.nodeRound > 0`:

```css
.round-counter {
  font-family: var(--font-mono);
  font-size: 9px;
  font-variant-numeric: tabular-nums;
  color: var(--accent-amber);
  letter-spacing: 0.04em;
  padding-top: var(--sp-2xs);
  margin-top: var(--sp-2xs);
  border-top: 1px solid var(--border-subtle);
}
```

Text: `"3 / 30"` (monospace tabular). Gap from badge `var(--sp-xs)` when on the same line on wide nodes.

**Gate-satisfied transition** (AC-7): on `bmad:node:gate_satisfied` the node gets `.gate-flash` for 400ms — a single keyframe that cross-fades the amber badge to the green check via opacity + a `background: color-mix(... green ... 30%, transparent)` mid-point. Concretely:

```css
@keyframes gate-flash {
  0%   { background: color-mix(in srgb, var(--accent-amber) 30%, transparent); }
  50%  { background: color-mix(in srgb, var(--accent-green)  40%, transparent); box-shadow: 0 0 0 3px color-mix(in srgb, var(--accent-green) 35%, transparent); }
  100% { background: transparent; }
}
.process-node.gate-flash { animation: gate-flash 400ms var(--ease-exit) 1; }
```

After 400ms the status prop has flipped to `complete` and the existing check glyph renders.

### 8. Signature elements (unmistakably Mashed)

1. **Chevron echo on selected radio option**: the `›` glyph in `--accent-green` mirrors the `.phase-bar` accent on the canvas node card — creates continuity between canvas and modal. No other dev-tool modal does this.
2. **Terminal-style prompt ID footer** (debug-only, shown when `localStorage.bmadDebug === '1'`): a single line at the bottom of the modal body rendered in `font-family: var(--font-code); color: var(--text-muted); font-size: var(--text-label)` with content `$ prompt-id: <hash8> · round <n>`. Echoes the `.commit-step::before` `$ ` prompt glyph in `style.css`. Makes Mashed feel like a terminal that grew a face.
3. **Amber-to-green gate-satisfied pulse**: the 400ms cross-fade described in §7 is the visual reward. Color transitions that traverse the palette (not just fade out) are rare in dev tools and make the "gate closed" moment feel earned.
4. **Round pill monospace with tight tabular nums**: `"3 / 30"` in Geist Mono with spaces around the slash — a Bloomberg-terminal tell. Most modals would write `"Round 3 of 30"`; we compress to numerical glance-value.

### 9. Motion catalogue

| Event | Transition | Duration | Easing | Notes |
|---|---|---|---|---|
| Modal enter | `fly` y=20 on `.modal-card`; `fade` on `.modal-overlay` | 200ms (card), `var(--duration-medium)` (overlay) | `cubicOut` / `var(--ease-enter)` | matches existing `QuestionResponseModal` |
| Modal exit | `fly` y=20 out | 150ms | `cubicIn` / `var(--ease-exit)` | |
| Snackbar slide-in | `fly` x=320 | 200ms | `cubicOut` | right-edge entry |
| Snackbar slide-out | `fly` x=320 | 200ms | `cubicIn` | |
| Snackbar reflow | `flip` | 200ms | default | when prompts resolve mid-stack |
| Badge pulse (`awaiting_input`) | `@keyframes awaiting-pulse` | 1800ms infinite | `var(--ease-move)` | 1800ms ≠ 1500ms (status-dot) ≠ 2000ms (agent-row) |
| Node halo pulse | `@keyframes awaiting-node-pulse` | 1800ms infinite | `var(--ease-move)` | amber 3px halo at 50% |
| Invalid shake | `@keyframes input-error-shake` | `var(--duration-shake)` 250ms once | `var(--ease-move)` | ±6px horizontal |
| Gate-satisfied flash | `@keyframes gate-flash` | `var(--duration-flash)` 400ms once | `var(--ease-exit)` | amber → green cross-fade |
| Round-complete tick | color transition on `.round-counter` amber → green 70% → amber | `var(--duration-tick)` 600ms | `var(--ease-move)` | fires on `bmad:node:round_complete`; number changes at 50% |
| Error-bar reveal | Svelte `slide` | `var(--duration-medium)` | default | same as existing |
| Submit-button press | `transform: scale(0.97)` | `var(--duration-micro)` | `var(--ease-enter)` | matches existing `.btn-send` |
| `round_limit` / `aborted` toast | `fly` y=-20 in, auto-dismiss 3s | 200ms | `cubicOut` | uses PROPOSED `--z-toast: 500` |

All motion wrapped in `@media (prefers-reduced-motion: no-preference)` where the effect is non-functional (pulse, shake, flash, tick). Enter/exit transitions stay because they convey structure, not decoration.

### 10. Proposed new tokens (confirm before implementation)

Add to `:root` in `frontend/src/style.css`:

```css
/* Modal sizing */
--modal-width-min: 360px;       /* PROPOSED */
--modal-width-wide: 720px;      /* PROPOSED — file/json shapes */
--modal-height-editor: 480px;   /* PROPOSED — json shape Monaco container */

/* Motion */
--duration-shake: 250ms;        /* PROPOSED — input-error-shake */
--duration-tick: 600ms;         /* PROPOSED — round-counter increment */
--duration-flash: 400ms;        /* PROPOSED — gate-satisfied cross-fade */

/* Layering */
--z-canvas: 1;                  /* PROPOSED — documents existing default */
--z-snackbar: 300;              /* PROPOSED — codifies existing magic number */
--z-modal: 400;                 /* PROPOSED — codifies existing magic number */
--z-toast: 500;                 /* PROPOSED — new toast layer above modal */

/* Buttons */
--btn-min-wide: 96px;           /* PROPOSED — approval yes/no balance */

/* Typography */
--font-code: 'JetBrains Mono', 'Geist Mono', ui-monospace, monospace; /* PROPOSED — referenced in DESIGN.md but no token exists */
```

Rationale: each value currently appears as a magic number in at least one component. Codifying before S6 code lands prevents further drift.
