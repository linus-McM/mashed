# Story breadcrumbs-01: Discovery + node-path data model

**Priority:** P0-critical
**Domain:** fullstack
**Estimated Complexity:** S
**Depends On:** none
**Status:** done

## Description

Phase 1 Task 1.0 discovery plus the supporting data-model shim so the UI has something concrete to bind to in breadcrumbs-02. We (a) pin down which Svelte components render the File Loader / process nodes today, and (b) introduce a typed `inputPath` / `outputPath` field path inside `WorkflowNode` config that the UI will read. Without this, every later story has a dangling assumption about where `node.data.config.inputPath` lives.

## Developer Notes

### Architecture

Per plan docs/plans/node-path-breadcrumbs.md lines 30-56 (Phase 1), the breadcrumb row reads `node.data.config.inputPath` / `outputPath`. The current canonical shape in `internal/bmad/types.go` (see corpus line 20816) is:

```go
Config     map[string]string  `json:"config"`
```

— a flat string map, not a struct. Rather than a large refactor here, scope this story to:
1. **Discovery**: grep `frontend/src/components/bmad/` for `File Loader` and confirm it renders through `ProcessNode.svelte` (registry.go:9234 already tags File Loader as a utility process). Plan line 34 lists `ProcessNode.svelte`, `CommandNode.svelte`, or an undiscovered type as candidates — confirm which.
2. **Data-model shim**: add well-known keys `inputPath` and `outputPath` to the existing `Config map[string]string`. No struct change yet — Phase 2 (breadcrumbs-05) introduces `OutputPaths` per architectural decision #1 (plan line 190).
3. Expose a tiny frontend helper `getNodePath(node, dir: 'in'|'out'): string` that reads `node.data.config.inputPath|outputPath` and returns `''` when unset. This is the single read-point the UI consumes from breadcrumbs-02 onward.

### Technical Considerations

- Do NOT change `WorkflowNode.Config`'s type yet — keep backward compatibility with saved workflows (plan line 190 rule: `json:",omitempty"` so old saves still parse).
- Document the reserved keys (`inputPath`, `outputPath`) inline in `internal/bmad/types.go` near the Config field so future authors don't collide.
- The helper returns an empty string (not `undefined`) so downstream `.svelte` templates can render the `—` dim dash unconditionally.

### Risks & Edge Cases

- A `CommandNode.svelte` mentioned in the plan may not exist in tree (corpus shows no such file). If discovery confirms absence, record it in the story output — story breadcrumbs-03 will drop the CommandNode task.
- File Loader may be a pseudo process registered in `internal/bmad/registry.go` (corpus line 9082) without its own component — it renders through `ProcessNode.svelte`. Verify.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 30-62 (Phase 1 intent), 190 (decision #1).
- Corpus: `internal/bmad/types.go` (WorkflowNode.Config), `internal/bmad/registry.go` (File Loader process registration), `frontend/src/components/bmad/ProcessNode.svelte` (likely renderer).
- Skill: `/wails` for binding patterns; `/simplify` on the helper.

### Discovery Findings

- **File Loader renderer:** `frontend/src/components/bmad/ProcessNode.svelte` — the generic process node, keyed only by xyflow node `type: 'bmadProcess'` via the `nodeTypes` map in `frontend/src/views/WorkflowBuilder.svelte` (`nodeTypes = { bmadProcess: ProcessNode, condition, loop, loopUntil, transform, merge }`). ProcessNode.svelte itself has **NO** conditional branch on `data.processId === 'util-file-loader'`; File Loader renders with the same template as every other process.
- **Distinct File Loader component file?** no — no dedicated `FileLoaderNode.svelte` or equivalent exists in the tree.
- **CommandNode.svelte exists?** no — zero matches for `CommandNode` in the corpus (`files.md`). Confirmed absent.
- **`util-file-loader` special-case location:** only `frontend/src/components/bmad/NodeConfigPanel.svelte` (right-sidebar config panel) branches on `isFileLoader = node?.data?.processId === 'util-file-loader'` to show a file picker + preview; canvas rendering itself is generic.
- **File Loader process registration:** `internal/bmad/registry.go` line 406 (`ID: "util-file-loader"`, `Name: "File Loader"`, `Phase: PhaseUtilities`, `Outputs: []string{"file-path"}`) inside the `// ── Utilities ──` section at line 404. No `NodeTypeFileLoader` constant exists — processes are looked up by string ID via `ProcessByID`.
- **Consequence for downstream stories:** breadcrumbs-03 should target `ProcessNode.svelte` (adding a conditional breadcrumb render guarded by `data.processId === 'util-file-loader'`, or a generic breadcrumb that reads `config.inputPath`/`outputPath` for every process). The "CommandNode.svelte" task in the plan is obsolete — drop it. Any "distinct File Loader renderer" task is likewise obsolete; work flows through the shared ProcessNode renderer plus NodeConfigPanel's existing `isFileLoader` branch.

## Acceptance Criteria

AC-1: Discovery report captured
- Given the canvas uses multiple node components
- When the dev lists every renderer of a File Loader node on the canvas
- Then the story records the exact file path of the File Loader renderer
- And records whether `CommandNode.svelte` exists (true/false) so downstream stories can skip or target it

AC-2: Reserved config keys documented
- Given `WorkflowNode.Config map[string]string` is the current storage
- When a reader inspects `internal/bmad/types.go`
- Then inline comments declare `inputPath` and `outputPath` as reserved keys with documented semantics (resolved absolute paths, empty string = unresolved)

AC-3: Frontend helper returns safe defaults
- Given a node whose `data.config.inputPath` is `undefined`, `null`, or missing
- When the UI calls `getNodePath(node, 'in')`
- Then it returns the empty string `''`
- And `getNodePath(node, 'out')` behaves symmetrically for `outputPath`

AC-4: No breakage of saved workflows
- Given an existing saved `WorkflowDef` JSON without the new keys
- When the app loads and re-saves it
- Then the JSON round-trips with no spurious `inputPath`/`outputPath` entries and no errors

## BDD Test Scenarios

```gherkin
Feature: Node path data model

  Scenario: getNodePath returns empty when key absent
    Given a WorkflowNode whose data.config is an empty object
    When getNodePath(node, 'in') is called
    Then it returns ''
    And getNodePath(node, 'out') returns ''

  Scenario: getNodePath reads populated keys
    Given a WorkflowNode whose data.config.inputPath is "/Users/x/_bmad-output/planning-artifacts/PRD.md"
    When getNodePath(node, 'in') is called
    Then it returns "/Users/x/_bmad-output/planning-artifacts/PRD.md"

  Scenario: Legacy workflow round-trip
    Given a WorkflowDef JSON saved before this story
    When the app loads and re-saves it via SaveBmadWorkflow
    Then the re-saved JSON contains no new reserved keys
    And re-loading the workflow succeeds without error
```

## Tasks / Subtasks

- [x] Task 1: Discovery of File Loader renderer (AC-1)
  - [x] Grep `frontend/src/components/bmad/` for the string "File Loader"
  - [x] Grep for `file-path` symbolic artifact handling
  - [x] Record findings at top of this story's Developer Notes (append, don't edit ACs)
- [x] Task 2: Document reserved config keys in `internal/bmad/types.go` (AC-2)
  - [x] Add doc comment above `Config` field listing `inputPath`, `outputPath`
  - [x] State that values are absolute filesystem paths and empty means unresolved
- [x] Task 3: Frontend helper `getNodePath` (AC-3)
  - [x] Add `frontend/src/lib/bmad/nodePath.ts` exporting `getNodePath(node, dir)`
  - [x] Handle missing `data`, missing `config`, and non-string values — all return `''`
  - [x] Unit test covering the three null/missing cases + the happy path
- [x] Task 4: Legacy compatibility check (AC-4)
  - [x] Add a Go round-trip test in `internal/bmad/types_test.go` (or nearest existing test file) covering WorkflowDef without the new keys

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ code coverage on new/modified files (bmad package: 91%; nodePath.ts all branches)
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `/simplify` run on all modified code
- [x] Code review: no CRITICAL/HIGH issues (lead self-review — doc-only + 15-line helper)
