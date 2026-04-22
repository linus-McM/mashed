# Story svelte-check-05: BMAD components retyping (Phase 5)

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** L
**Depends On:** svelte-check-03
**Status:** done
**Landed:** 112b7bd (2026-04-22)

## Description

Retype the seven BMAD canvas components (CanvasPane 58, NodeConfigPanel 45, ProcessSidebar 40, SprintPanel 25, GitPanel 24, NodeInputSnackbarStack 18, ArrayEditorModal 18 = 228 direct errors). These components back the core BMAD workflow UX — canvas rendering, node configuration drawer, sprint panel, git panel, snackbar stack. Can run in parallel with Phase 4 stories. Expected reduction: ~228 direct + minimal downstream (BMAD subsystem is self-contained). Count drops by ~220 (e.g. from ~420 to ~200).

## Developer Notes

### Architecture
- Files:
  - `frontend/src/components/bmad/CanvasPane.svelte` (58)
  - `frontend/src/components/bmad/NodeConfigPanel.svelte` (45)
  - `frontend/src/components/bmad/ProcessSidebar.svelte` (40)
  - `frontend/src/components/bmad/SprintPanel.svelte` (25)
  - `frontend/src/components/bmad/GitPanel.svelte` (24)
  - `frontend/src/components/bmad/NodeInputSnackbarStack.svelte` (18)
  - `frontend/src/components/bmad/ArrayEditorModal.svelte` (18)
- Shared BMAD types (all from `$lib/types/wails`):
  - `ProcessDef` — the catalogue entry
  - `WorkflowExecution` — live execution snapshot (NodeRounds, PendingPrompts, NodeInputs, NodeOutputs)
  - `PendingPrompt` — what InputResponseModal/Snackbar render (shape discriminator + options)
  - `NodeInputEntry` — a resolved input with valueHash
  - `InputSpec`, `OutputSpec`, `IterationGate` — from cerebrum §BMAD Interactive Processes
- xyflow types for CanvasPane — `Node<WorkflowNodeData>`, `Edge`, xyflow handles.
- NodeInputSnackbarStack merges legacy `QuestionEvent` with `PendingPrompt` keyed by `(nodeId, kind)` per cerebrum.

### Technical Considerations
- **NodeInputSnackbarStack already has interactive typing partially in place** (per cerebrum 2026-04-20, `interactiveInput.ts` store is the source of truth). Check whether residual errors are in the legacy merge logic — those are prime targets.
- **CanvasPane xyflow.** Same patterns as Phase 4a's WorkflowBuilder: `Node<WorkflowNodeData>`, `Edge`, `NodeChange[]`, `EdgeChange[]`. Reuse the types defined in Phase 4a.
- **NodeConfigPanel** — drawer with per-node-type config forms. Define a discriminated union for NodeType (`'process' | 'condition' | 'loop' | 'loopUntil' | 'merge' | 'transform' | 'command' | 'fileLoader' | 'multiFileLoader'`) and a `NodeConfig` union mapping NodeType → config shape. Cross-check with `internal/bmad/types.go`.
- **ProcessSidebar** lists `ProcessDef[]` — simple prop+list. Type the search/filter state. Watch for `categoryColors: Record<string, string>` — apply Phase 3 patterns, use tokens not hex.
- **SprintPanel** shows sprint-status.yaml parsed output. Define `SprintStatus`, `EpicEntry`, `StoryEntry` types in `frontend/src/types/sprint.ts` mirroring the Go YAML parser output.
- **GitPanel** wraps git status/branch/push operations — types from `$lib/types/wails` (git methods on `*App`).
- **ArrayEditorModal** — generic editor for string[] fields (e.g. `loopBodyNodes`). Type with `{ value: string[]; onChange: (next: string[]) => void }`.

### Risks & Edge Cases
- **R1 — Wails regen.** BMAD types evolved heavily through Sprints 3-7 and interactive stories. Re-run `wails dev` before starting this story to ensure the re-exports in `$lib/types/wails` reflect the latest Go types.
- **R3 — transient bumps.** Fixing BMAD types may unmask errors in BMAD tests (Phase 6). Do not fix tests here — keep scope to components.
- **Design tokens.** Per S6 cerebrum entry, modal/snackbar components must use `--modal-width-*`, `--duration-shake`, `--z-snackbar`. Do not replace tokens with literals during retyping.
- **NodeInputSnackbarStack legacy shim.** Per cerebrum, `QuestionSnackbarStack.svelte` is a re-export shim. If the shim's exports become a type-resolution issue, fix by making the shim forward types too (`export type * from './NodeInputSnackbarStack.svelte'`).
- **R4 — PR sizing.** 228 errors / 7 files — likely 2-3 PRs:
  - PR-A: CanvasPane + NodeConfigPanel (canvas core)
  - PR-B: ProcessSidebar + SprintPanel + GitPanel (sidebars)
  - PR-C: NodeInputSnackbarStack + ArrayEditorModal (modals/snackbar)

### Reference Files
- `frontend/src/components/bmad/InputResponseModal.svelte` — reference for fully-typed BMAD modal
- `frontend/src/components/bmad/ProcessNode.svelte` — already partially typed, shows xyflow node pattern
- `frontend/src/stores/interactiveInput.ts` — typed store (cerebrum), reuse types
- `internal/bmad/types.go` — all Go-side BMAD types (ProcessDef, InputSpec, OutputSpec, IterationGate, etc.)
- `.wolf/cerebrum.md` §BMAD Interactive Processes for the complete type landscape

### Skills to invoke
- `/simplify` — mandatory before commit.
- `/xyflow` — for CanvasPane node/edge typing.
- `/wails` — verify regenerated bindings match the re-exports used.
- `/playwright-cli` — smoke canvas render, node config drawer, sprint panel load, git panel push.

## Acceptance Criteria

AC-1: NodeType discriminated union defined
- Given the nine node types per `internal/bmad/types.go`
- When this story lands
- Then `frontend/src/types/bmad.ts` (or equivalent) exports a `NodeType` literal union
- And a `NodeConfig` discriminated union maps each NodeType to its config shape

AC-2: All BMAD component props typed
- Given the seven components
- When this story lands
- Then every `export let` has an explicit type annotation (from `$lib/types/wails` or local types)
- And no prop defaults to `any` or `{}`

AC-3: CanvasPane uses typed xyflow stores
- Given CanvasPane.svelte
- When this story lands
- Then `nodes`, `edges` stores are `Writable<Node<WorkflowNodeData>[]>` / `Writable<Edge[]>`
- And change handlers consume typed `NodeChange[]` / `EdgeChange[]`

AC-4: Tokens preserved
- Given the design tokens in use (`--modal-width-*`, `--z-snackbar`, etc.)
- When this story lands
- Then no hex colour literal is introduced in any of the seven files
- And token usage is preserved

AC-5: Error counts drop
- Given the entry counts per file (from sveltecheck-by-file)
- When this story lands
- Then each target file has ≤ 3 remaining errors
- And `just sveltecheck-count` drops by ≥ 200 from story entry
- And `just sveltecheck-ratchet` passes

## BDD Test Scenarios

### Scenario 1: NodeType discriminated union
```gherkin
Feature: NodeType discriminated union

  Scenario: Condition node config has condition field
    Given NodeConfig discriminated union on nodeType
    When a consumer narrows "if (node.data.nodeType === 'condition')"
    Then node.data.config.expression is typed string (no implicit any)
    And node.data.config.loopBodyNodes is a compile error

  Scenario: Transform node has extract fields
    Given nodeType === 'transform'
    When the discriminated branch is entered
    Then sourceNode, extractType, extractPattern are typed
    And invalid extractType (e.g. "xml") is a compile error
```

### Scenario 2: CanvasPane xyflow
```gherkin
Feature: CanvasPane typed stores

  Scenario: nodes store rejects arbitrary shape
    Given let nodes: Writable<Node<WorkflowNodeData>[]> = writable([])
    When nodes.update(arr => [...arr, { id: '1', data: { wrong: 1 } } as unknown as Node<WorkflowNodeData>])
    Then svelte-check flags the unsafe cast
    And the correctly shaped object is accepted

  Scenario: Edge sourceHandle typed
    Given Edge from @xyflow/svelte has sourceHandle?: string
    When a branch edge sets sourceHandle: "loop-exit"
    Then it compiles
```

### Scenario 3: Snackbar legacy+interactive merge
```gherkin
Feature: NodeInputSnackbarStack merges legacy + interactive

  Scenario: Pending prompt keyed by (nodeId, kind)
    Given a QuestionEvent and a PendingPrompt for the same nodeId
    When the stack renders
    Then deduplication is by (nodeId, kind) keys
    And both legacy questions and interactive prompts appear

  Scenario: Skip button only for optional prompts
    Given PendingPrompt with Required=false
    When rendered in the snackbar
    Then a "Skip" button appears
    And for Required=true, no Skip button appears
```

### Scenario 4: Error delta
```gherkin
Feature: Phase 5 reduction

  Scenario: Count drops by ≥ 200
    Given entering at the parallel-safe count (≥ 420)
    When Phase 5 lands
    Then "just sveltecheck-count" is at least 200 below entry
    And the ratchet passes
```

## Tasks / Subtasks

- [ ] Task 1: Define shared BMAD types (AC: 1)
  - [ ] `frontend/src/types/bmad.ts` with `NodeType`, `NodeConfig`, `NodeData = { processDef, nodeType, config }`
  - [ ] Mirror `internal/bmad/types.go` — any mismatch is a bug
  - [ ] Re-export shapes via `$lib/types/wails` where Go types already supply them

- [ ] Task 2: Retype CanvasPane (AC: 2, 3)
  - [ ] Use xyflow `Node`, `Edge` generic typing (reuse types from Phase 4a)
  - [ ] Type drag-drop, zoom, pan handlers

- [ ] Task 3: Retype NodeConfigPanel (AC: 2)
  - [ ] Discriminated union for per-node config forms
  - [ ] Type form input events and dispatchers

- [ ] Task 4: Retype ProcessSidebar, SprintPanel, GitPanel (AC: 2)
  - [ ] ProcessSidebar: `ProcessDef[]`, search/filter state
  - [ ] SprintPanel: define `SprintStatus`, `EpicEntry`, `StoryEntry` in `types/sprint.ts`; match Go YAML parser output
  - [ ] GitPanel: types from `$lib/types/wails` (git methods)

- [ ] Task 5: Retype NodeInputSnackbarStack + ArrayEditorModal (AC: 2)
  - [ ] Preserve legacy+interactive merge keying
  - [ ] ArrayEditorModal: `{ value: string[]; onChange: (next: string[]) => void }` props

- [ ] Task 6: Verify reductions and behaviour (AC: 5)
  - [ ] Each target file ≤ 3 residual errors
  - [ ] `vitest run` — 696 tests pass (BMAD test retyping happens in Phase 6)
  - [ ] `vite build` clean
  - [ ] `wails dev` smoke: open canvas, click a process node, open config drawer, view sprint panel, view git panel
  - [ ] Commit body records delta

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] `go build ./...` / `go vet ./...` / `go test ./... -race` pass
- [ ] Target files each have ≤ 3 residual errors
- [ ] `just sveltecheck-count` drops by ≥ 200 from entry
- [ ] `just sveltecheck-ratchet` passes
- [ ] `vitest run` — 696 tests pass
- [ ] `vite build` clean
- [ ] Manual wails dev smoke of full BMAD canvas workflow
- [ ] `/simplify` run on every modified file
- [ ] No `any`, no `@ts-ignore`, no `@ts-nocheck`
- [ ] PR size <= 400 lines (split into up to 3 PRs — R4)
- [ ] Commit body records: `sveltecheck-count: <prev> → <cur> (Δ -<n>)`
