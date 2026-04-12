# skills-cmd-03: Canvas load/save/restore symmetry + run-time fail verification

**Status:** ready
**Domain:** fullstack
**Size:** S
**Depends on:** skills-cmd-01, skills-cmd-02
**Phase:** 2

## Description

Close out Phase 2 by ensuring command nodes round-trip through `saveWorkflow`, `loadNodesEdges`, `snapshotCanvas`, and `restoreForRepo` without loss of metadata, and by adding an end-to-end negative test that proves a command node on the canvas fails with a clear error when the workflow is run (Phase 3 will later replace the failure with real execution). This is the integration story that makes Phase 2 shippable on its own: the canvas draws, the canvas saves, the canvas reloads, and attempting to run produces a legible error instead of a silent hang.

**This story is UI-facing — the failure-state visual must be reviewed by ui-architect.**

## Developer Notes

- **Files to create/modify:**
  - `frontend/src/views/WorkflowBuilder.svelte` — extend `loadNodesEdges` type-mapping ternary (≈ line 515) so `n.nodeType === 'command'` resolves to `type: 'command'`, not the `bmadProcess` fallback.
  - `frontend/src/views/WorkflowBuilder.svelte` — verify `saveWorkflow` already writes `nodeType: n.data.nodeType || ''` and that this carries `"command"` through untouched. No code change expected, but ADD an assertion test.
  - `frontend/src/views/WorkflowBuilder.svelte` — verify `snapshotCanvas` hashes `n.type` and therefore differentiates command vs process nodes. No code change expected — ADD a test.
  - `frontend/src/views/WorkflowBuilder.svelte` — verify `restoreForRepo` operates on node IDs (not types) and so does nothing type-specific. ADD a regression test.
  - `internal/bmad/executor_test.go` (or new file) — integration test: run a workflow with one command node, assert the node transitions to `failed` with the Phase 2 sentinel message.
- **Risks / gotchas:**
  - Plan §Phase 2 "Load/save/restore symmetry" lists four call sites. Only `loadNodesEdges` actively needs a code change; the other three are "no change needed" but the story MUST lock them with tests so a future refactor can't silently break them.
  - Do NOT touch the process-node code path. All changes are additive guards or extension of the type-mapping ternary.
  - The fail-fast message from skills-cmd-01 must appear in the frontend via the existing `bmad:node:status` event — confirm the frontend surfaces the failure reason (may already do so; write a Playwright assertion).
- **Prerequisites already in place:**
  - skills-cmd-01 added the backend dispatch that emits `NodeFailed` with a log line.
  - skills-cmd-02 added `CommandNode.svelte` and the drop handler.
  - `saveWorkflow` and `snapshotCanvas` already exist and function for process nodes.

## Acceptance Criteria

**AC-1: Save + reload preserves command nodes**
- Given a canvas containing one process node and one command node connected by an edge
- When the user clicks Save and then reloads the app
- Then both nodes reappear with their original `type` values (`bmadProcess` and `command`)
- And the command node's `config.commandName`, `commandPath`, `commandDescription` survive unchanged

**AC-2: Snapshot detects canvas drift for command nodes**
- Given a snapshot is taken of a canvas containing one command node
- When the user swaps the command node for a process node at the same ID and re-snapshots
- Then the two snapshot hashes differ

**AC-3: Running a workflow with a command node fails cleanly**
- Given a saved workflow with exactly one command node
- When the user clicks Run
- Then the node transitions to `failed` state visibly in the UI
- And the failure reason surfaced to the user references "command nodes not yet runnable" (or equivalent Phase 2 sentinel message)
- And the workflow runner transitions to `ExecFailed` (or equivalent terminal state) without hanging

**AC-4: `restoreForRepo` is type-agnostic**
- Given a canvas overlay restored for a repo with mixed process and command nodes
- When `restoreForRepo` runs
- Then the restored overlay operates on node IDs without filtering by `type`
- And both node types retain their positions

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Command node canvas round-trip

  Scenario: Mixed workflow survives save + reload
    Given a canvas with a process node "proc-A" and a command node "cmd-B" connected by an edge
    When the user saves the workflow
    And the user closes and reopens the app
    Then the loaded canvas contains "proc-A" rendered as ProcessNode
    And the loaded canvas contains "cmd-B" rendered as CommandNode
    And "cmd-B".data.config.commandName is unchanged

  Scenario: Snapshot differentiates node types
    Given a canvas with one command node at ID "n1"
    When snapshotCanvas is called and produces hash H1
    And the node at "n1" is replaced by a process node
    And snapshotCanvas is called again to produce hash H2
    Then H1 is not equal to H2

  Scenario: Running a command node fails with clear error
    Given a workflow containing one command node with commandName "simplify"
    When the user clicks Run
    Then the node transitions to "failed"
    And the UI displays a message containing "command nodes not yet runnable"
    And no tmux session was started
```

## Tasks / Subtasks

- [ ] Task 1 — Extend `loadNodesEdges` type mapping (AC-1)
  - [ ] Update the ternary so `command` maps to `type: 'command'`
  - [ ] Keep the default → `bmadProcess` fallback intact for legacy workflows
- [ ] Task 2 — Lock save + snapshot + restore with tests (AC-1, AC-2, AC-4)
  - [ ] Frontend unit test or Playwright: save a mixed canvas, reload, assert types and config keys
  - [ ] Frontend unit test: snapshotCanvas distinguishes process from command at the same ID
  - [ ] Frontend unit test: restoreForRepo preserves positions for both types
- [ ] Task 3 — Integration fail-test (AC-3)
  - [ ] Backend: Go integration test driving a one-command-node workflow through the executor, asserting the sentinel error path
  - [ ] Frontend: Playwright scenario clicking Run on a canvas with a command node, asserting the failed badge and the error message
- [ ] Task 4 — Phase 2 smoke checklist in commit message
  - [ ] Document the three verification steps from plan §Phase 2 "Verification before shipping Phase 2" as completed

## Definition of Done

- [ ] All ACs verified by an automated test (Go + Playwright; no "manually verified")
- [ ] Coverage ≥ 80% on modified files
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths or magic numbers added
- [ ] Existing tests still pass

## Design Brief

### Layout composition
No new surface introduced — this story hardens the canvas around existing components (`CommandNode` from skills-cmd-02, `ProcessNode`, `DeletableEdge`, `CanvasPane`). The design work is about visual differentiation and failure-state legibility on a shared canvas.

- **Canvas layering** unchanged: `Background` (dot pattern) → edges → nodes → `Controls` / `MiniMap`. Both node types sit at the same z-layer.
- **Edge rule zones** — connections between process→command, command→process, and command→command must look the same so a mixed workflow reads as one pipeline, not two. Use the existing `DeletableEdge` component unchanged; do not introduce a command-specific edge style.
- **Failure banner** — when `bmad:node:status` surfaces the Phase 2 sentinel message, show it as a toast-style strip anchored to the bottom of the canvas viewport (`position: absolute; bottom: var(--sp-lg); left: 50%; transform: translateX(-50%);`). This is a status channel, not a modal — it must not block interaction.

### Typography plan
- **Sentinel error toast**: `font-family: var(--font-ui)` (Geist), `font-size: var(--text-body)` (13px), `font-weight: 500`, `color: var(--text-primary)`.
- **Sentinel error label prefix** ("Command nodes not yet runnable"): same size but `color: var(--accent-red)`, `font-weight: 600`.
- **Node type differentiation** (inherited from skills-cmd-02): ProcessNode body stays `var(--font-mono)`, CommandNode header uses `var(--font-ui)`. This typographic split is the primary at-a-glance differentiator.

### Color strategy
- **Failure toast background**: `background: color-mix(in srgb, var(--accent-red) 12%, var(--bg-elevated));` — a red-tinted elevated surface, not a solid red wall. Matches the alpha grammar the rest of the app uses (`.btn-discard:hover` etc.).
- **Failure toast border**: `1px solid color-mix(in srgb, var(--accent-red) 40%, transparent)`.
- **Failed node border**: `var(--accent-red)` when status flips to `failed` — ProcessNode and CommandNode both already render this via their shared `.status-x` / `.failed-text` block.
- **Edge colors**: untouched — existing `DeletableEdge` stroke is `var(--border-emphasis)` with a hover state. A command→process edge is visually identical to a process→process edge on purpose: the pipeline is one pipeline.
- **Differentiation colors** (from skills-cmd-02): CommandNode stripe = `var(--accent-green)`; ProcessNode stripe = phase color from the `phaseColors` map. Same elevation, different accents.

### Interaction model
- **Connect rules**: `isValidConnection` allows any source→target pairing. No type-based rejection — the rendering, not the graph, signals which node is a command.
- **Failure toast dismissal**: auto-dismiss after `3000ms`, or `Escape` to dismiss immediately. Click does NOT dismiss (user should be able to copy the text).
- **Failed-node click**: selecting a failed node surfaces the same message in the sidebar detail panel (leverages existing detail flow).
- **State transition**: when a node transitions pending → failed, use a `transition: border-color var(--duration-medium) var(--ease-enter)` — no bounce or shake; this codebase is minimal-functional per `DESIGN.md`.
- **Toast entry**: `slide + fade from bottom, 150ms var(--ease-enter)` — mirrors `DESIGN.md`'s "notification entry" spec.

### Component specs
```
.canvas-failure-toast {
  position: absolute;
  bottom: var(--sp-lg);
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: var(--sp-sm);
  padding: var(--sp-sm) var(--sp-md);
  background: color-mix(in srgb, var(--accent-red) 12%, var(--bg-elevated));
  border: 1px solid color-mix(in srgb, var(--accent-red) 40%, transparent);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-family: var(--font-ui);
  font-size: var(--text-body);
  box-shadow: 0 8px 24px color-mix(in srgb, var(--bg-deepest) 70%, transparent);
  z-index: 50;
}
.canvas-failure-toast .label {
  color: var(--accent-red);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-size: var(--text-label);
}
```
Shadow is a hand-rolled value — no `--shadow-*` tokens exist in `style.css`. Token missing — use the literal above and add a TODO to promote to a `--shadow-lifted` token when a second modal needs it.

### Signature elements
- **Typographic differentiation between node types** (mono vs Geist) is the cohesive "mashed" signature on the canvas. No extra decoration needed — the two faces doing their jobs is the visual language.
- **Red-tinted toast at the bottom** rather than a modal dialog respects the "notification-first IDE" product thesis from `DESIGN.md`. The user stays in the canvas, sees the failure, continues editing.

### Distinguishing from existing patterns
- **vs NameWorkflowModal** (full overlay) — the failure toast is NON-blocking, bottom-anchored, auto-dismissing. Modals stop the user; toasts inform the user. Phase 2 failure is advisory (Phase 3 will replace with real execution), so a toast is the right weight.
- **vs existing `bmad:node:status` inline badge** — the badge on the node itself is persistent and compact (one glyph + label); the toast is transient and wordy. They complement: the node badge says "this failed", the toast says "and here's why".
