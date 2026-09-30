# Story svelte-check-04a: View-level state — WorkflowBuilder (Phase 4a)

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** L
**Depends On:** svelte-check-03
**Status:** done
**Landed:** 13276ce (2026-04-22)

## Description

Retype the single largest view in the codebase: `WorkflowBuilder.svelte` (162 errors). Every `export let prop` gains a JSDoc `@type` or explicit annotation; every local `let` inferred as `any` gets a real type; every canvas ref (`@xyflow/svelte` nodes/edges stores) is typed against the library surface. Expected reduction: ~162 direct + ~40 downstream = ~200. Count drops from ~930 to ~730.

## Developer Notes

### Architecture
- Sole file: `frontend/src/views/WorkflowBuilder.svelte` (162 errors at Phase 4 entry).
- This is the canonical BMAD workflow canvas — subscribes to `bmad:*` events per cerebrum (interactiveInput store routing).
- Core typed shapes needed:
  - `Workflow`, `WorkflowNode`, `WorkflowEdge` — re-export from `$lib/types/wails` (Story 01)
  - `PendingPrompt`, `WorkflowExecution` — re-export from `$lib/types/wails`
  - `@xyflow/svelte` types: `Node<WorkflowNodeData>`, `Edge`, `XYPosition`, `NodeChange`, `EdgeChange` — import from `@xyflow/svelte` directly
- Event payload types:
  - `bmad:node:awaiting_input` → `{ execId: string; nodeId: string; prompt: PendingPrompt }`
  - `bmad:node:input_resolved` → `{ execId: string; nodeId: string; valueHash: string }`
  - `bmad:node:round_complete` / `:gate_satisfied` / `:round_limit` / `:aborted` / `:input_invalid`
- Reactive statements (`$:`) inherit types; no action once underlying vars are typed.

### Technical Considerations
- **xyflow types are the load-bearing dependency.** `@xyflow/svelte` ships its own `.d.ts`. Import `Node`, `Edge`, `NodeChange`, `EdgeChange` directly. The nodes/edges writable stores should be typed `Writable<Node<WorkflowNodeData>[]>`.
- **Node data payload.** Define `type WorkflowNodeData = { processDef: ProcessDef; nodeType: NodeType; ... }` matching the Go-side node data. Cross-check `internal/bmad/types.go`.
- **Drag & drop handlers.** `on:drop={(e: DragEvent) => ...}` — extract `e.dataTransfer` carefully; it's `DataTransfer | null`. Narrow or early-return.
- **Panzoom integration.** `panzoom` is typed via `@types/...` — if not, declare a local module in `frontend/src/types/shims.d.ts`.
- **Cross-pollination with BMAD subsystem state.** `NodeRounds`, `PendingPrompts`, `NodeInputs` per cerebrum §BMAD Interactive Processes — reuse those types from the Wails re-exports, do not redeclare.
- **R3 — transient bumps.** WorkflowBuilder ripple exposes downstream issues in ProcessNode.svelte, CanvasPane.svelte, NodeConfigPanel.svelte. Those are covered in Phase 5 story; expect a small +30 bump mid-PR before net -162.

### Risks & Edge Cases
- **R1 — Wails regen.** If `wails dev` regenerates models during this story, double-check `WorkflowExecution` and `PendingPrompt` fields match the imports.
- **R4 — reviewer fatigue.** 162 errors in one file could mean >300 lines of diff. Split into up to 2 PRs if needed:
  - PR-A: props + local `let` typing (no behaviour change)
  - PR-B: event subscriptions + handler typing
- **xyflow version pinning.** `@xyflow/svelte@0.1.39` per package.json. Lock API — do not bump.
- **Drop targets.** HTML5 DragEvent's `dataTransfer.types` is readonly DOMStringList — iterating requires spread or for-loop.
- **Tokens.** Per S6 cerebrum entry, all modal/canvas tokens (`--z-canvas`, `--duration-shake`) must continue to be referenced from CSS — do not introduce magic numbers.

### Reference Files
- `frontend/src/views/WorkflowBuilder.svelte` (target; read to catalogue `any` sites)
- `frontend/src/components/bmad/InputResponseModal.svelte` — existing fully-typed modal, copy its pattern.
- `frontend/src/components/bmad/NodeInputSnackbarStack.svelte` — interactive snackbar, already typed.
- `frontend/src/stores/interactiveInput.ts` — event store types (cerebrum).
- `internal/bmad/types.go` — backend source of truth for ProcessDef, WorkflowExecution, PendingPrompt.

### Skills to invoke
- `/simplify` — mandatory before commit.
- `/xyflow` — xyflow Svelte canvas patterns, node/edge typing, handle orientation.
- `/wails` — event subscription typing (`EventsOn` generic signature).
- `/playwright-cli` — smoke the workflow canvas (drag a node, run a workflow, answer a prompt).

## Acceptance Criteria

AC-1: All props typed
- Given WorkflowBuilder's `<script lang="ts">` (or JSDoc blocks on `<script>`)
- When this story lands
- Then every `export let` has an explicit type annotation or JSDoc `@type`
- And no prop is inferred as `any` or `{}`

AC-2: xyflow stores typed
- Given the nodes and edges `writable` stores
- When this story lands
- Then each is typed as `Writable<Node<WorkflowNodeData>[]>` / `Writable<Edge[]>`
- And node/edge change handlers consume typed `NodeChange[]` / `EdgeChange[]`

AC-3: Event subscriptions typed
- Given `EventsOn('bmad:node:awaiting_input', handler)` and peers
- When this story lands
- Then the handler parameter is typed against the event payload interface
- And every payload field used (execId, nodeId, prompt, valueHash) has a real type

AC-4: WorkflowBuilder error count drops to ≤ 10
- Given the file entered with 162 errors
- When this story lands
- Then `just sveltecheck-file src/views/WorkflowBuilder.svelte` reports ≤ 10 remaining errors
- And all remaining errors are documented in the PR description with a follow-up story reference

AC-5: Runtime unchanged
- Given the canvas and workflow execution flows (drag node, run, answer prompt, resume)
- When a user runs a workflow after this story
- Then all behaviours (execution, node status animation, snackbar, modal) work identically to before
- And vitest+playwright regression suites pass

## BDD Test Scenarios

### Scenario 1: Node data typing
```gherkin
Feature: xyflow node data shape

  Scenario: Node<WorkflowNodeData> rejects arbitrary data
    Given the nodes store is typed Writable<Node<WorkflowNodeData>[]>
    When a caller adds a node with data: { wrong: 1 }
    Then svelte-check reports a type error
    And data: { processDef, nodeType } passes

  Scenario: Edge handles typed
    Given SourceHandle and TargetHandle are optional strings
    When an edge has sourceHandle: "loop-exit"
    Then it matches the Edge type without error
```

### Scenario 2: Event subscription
```gherkin
Feature: Typed BMAD event subscription

  Scenario: awaiting_input payload typed
    Given EventsOn("bmad:node:awaiting_input", (evt) => ...)
    When the handler reads evt.prompt.shape
    Then the type is Shape ("free"|"choice"|"multi"|"approval"|"file"|"json")
    And evt.prompt.wrong is a compile error

  Scenario: input_resolved payload carries valueHash only
    Given EventsOn("bmad:node:input_resolved", (evt) => ...)
    When the handler attempts to read evt.value
    Then svelte-check reports a type error
    And evt.valueHash is accessible as string
```

### Scenario 3: Drag and drop
```gherkin
Feature: Typed drag-drop handlers

  Scenario: DataTransfer null handled
    Given on:drop={(e: DragEvent) => dropHandler(e)}
    When the handler accesses e.dataTransfer
    Then svelte-check flags e.dataTransfer as "DataTransfer | null"
    And the handler must narrow before using .getData

  Scenario: Dropped node creates canvas node
    Given a user drags a ProcessDef from the sidebar onto the canvas
    When the drop fires
    Then a new node appears at the cursor position
    And the node.data matches WorkflowNodeData
```

### Scenario 4: Error delta
```gherkin
Feature: Phase 4a reduction

  Scenario: WorkflowBuilder drops to ≤10 errors
    Given file baseline 162
    When Phase 4a lands
    Then "just sveltecheck-file src/views/WorkflowBuilder.svelte" reports ≤10
    And the ratchet passes

  Scenario: Overall count drops
    Given entering at ~930
    When Phase 4a lands
    Then total count is ≤ 770
```

## Tasks / Subtasks

- [ ] Task 1: Inventory the untyped surface (AC: 1, 2)
  - [ ] Run `just sveltecheck-file src/views/WorkflowBuilder.svelte`, save the error list
  - [ ] Identify the distinct `let` / `export let` / store decls
  - [ ] Draft the list of required types and where they live

- [ ] Task 2: Define/refine shared types (AC: 1, 2, 3)
  - [ ] In `frontend/src/types/workflow.ts` or a new `types/canvas.ts`: `WorkflowNodeData`, `CanvasNode = Node<WorkflowNodeData>`
  - [ ] Re-export BMAD types via `$lib/types/wails`
  - [ ] Add BMAD event payload interfaces (one per event)

- [ ] Task 3: Retype props and local state (AC: 1, 2)
  - [ ] Convert `<script>` → `<script lang="ts">` if not already
  - [ ] Annotate each `export let` with an explicit type
  - [ ] Annotate each internal `let` with a precise type

- [ ] Task 4: Retype event handlers and subscriptions (AC: 3)
  - [ ] Apply event-payload interfaces to every `EventsOn` handler
  - [ ] Type every DOM handler (drag, drop, click, keydown)
  - [ ] Ensure no `any` appears in handler signatures

- [ ] Task 5: Verify reductions and behaviour (AC: 4, 5)
  - [ ] `just sveltecheck-file src/views/WorkflowBuilder.svelte` ≤ 10
  - [ ] `vitest run` — 696 tests pass
  - [ ] `vite build` clean
  - [ ] `wails dev` smoke: drag node onto canvas, start workflow, answer a prompt, verify idle completion
  - [ ] Commit body records delta: `sveltecheck-count: <prev> → <cur> (Δ -<n>)`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage preserved on any modified helper types (no regression)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `just sveltecheck-file src/views/WorkflowBuilder.svelte` ≤ 10
- [ ] `just sveltecheck-count` ≤ 770
- [ ] `just sveltecheck-ratchet` passes
- [ ] `vitest run` — 696 tests pass
- [ ] `vite build` clean
- [ ] Manual wails dev smoke: canvas drag, run, answer prompt — all pass
- [ ] `/simplify` run on WorkflowBuilder.svelte and new type files
- [ ] Code review: no `any`, no `@ts-ignore`, no `@ts-nocheck`
- [ ] PR size <= 400 lines (split into up to 2 PRs — R4)
- [ ] Commit body records: `sveltecheck-count: <prev> → <cur> (Δ -<n>)`
