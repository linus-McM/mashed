# Story breadcrumbs-07: MultiFileLoader node — backend

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** breadcrumbs-05
**Status:** done

## Description

Phase 2b Task 2b.1 (plan lines 111-115). Introduce a new node type `multiFileLoader` that takes a list of `{label, path}` pairs and emits one output per path (synchronous, no tmux). Labels flow verbatim into downstream prompts; an empty label falls back to positional `file[0]`, `file[1]`, … per architectural decision #3 (plan line 192).

## Developer Notes

### Architecture

Three files to touch:

1. **`internal/bmad/types.go`** — add `NodeTypeMultiFileLoader NodeType = "multiFileLoader"` to the NodeType const block (corpus line 20800).
2. **`internal/bmad/executor.go`** — synchronous execution branch. Pattern mirrors existing `executeControlNode` (corpus line 25522) — no tmux, completes inline, writes to `NodeOutputs`.
3. **`internal/bmad/registry.go`** — register the node so sidebar surfaces it (corpus line 9082).

Config shape (plan line 192 + Task 2b.1):

```go
type MultiFileEntry struct {
    Label string `json:"label"`
    Path  string `json:"path"`
}
```

Serialize the list as JSON into `node.Config["entries"]` to avoid a schema migration. The executor parses it on run. Empty `Label` → positional `file[N]`.

### Technical Considerations

- **No tmux**: this node runs entirely in-process. Treat like a control node — mark running, read each file, populate `OutputPaths`, mark complete.
- **File reading is optional**: plan Task 2b.1 says "Reads each configured path, emits one output per path". "Reads" likely means "resolves + validates"; the content flows to downstream via the path, not by loading bytes. Validate each file exists (`os.Stat`); mark missing entries as skipped but do NOT fail the node (user may be pre-configuring future paths).
- **OutputPaths**: write each entry to `OutputPaths` using either `entry.Label` or the fallback `file[N]`. Downstream references resolve the same way as a normal process node via breadcrumbs-06.
- **Deterministic labeling**: two entries with identical labels is a user error — the second must fail validation with `ErrDuplicateMultiFileLabel`. Use sentinel + `errors.Is` per project's error pattern.

### Wails binding requirements

None. Executor runs it as part of the DAG. `bmad:node:artifacts` event fires like any other node.

### Risks & Edge Cases

- **Concurrent node writes**: same `state.mu` as breadcrumbs-05. Nothing new.
- **Very large lists**: cap at 64 entries (reasonable UX ceiling). Beyond that return `ErrMultiFileTooMany`.
- **Relative paths**: user may type a relative path; executor must resolve against `repoPath` before validating existence.
- **Registry entry**: the `ProcessDef` for MultiFileLoader should have empty `Inputs`, empty `SkillName`, and phase `PhaseUtilities` (corpus line 20764). Plan confirms utility grouping.

### Reference Files

- Plan: `docs/plans/node-path-breadcrumbs.md` lines 111-127, 192.
- Corpus: `internal/bmad/types.go` (line 20748), `internal/bmad/executor.go` (line 25304, especially `executeControlNode` at 25522), `internal/bmad/registry.go` (line 9082).
- Skills: `/golang-testing`, `/golang-error-handling`.

## Acceptance Criteria

AC-1: NodeType registered and parseable
- Given the NodeType constants block in `types.go`
- When a WorkflowDef with a `multiFileLoader` node is loaded
- Then `node.EffectiveType()` returns `NodeTypeMultiFileLoader`

AC-2: Synchronous execution emits one OutputPath per entry
- Given a MultiFileLoader node with entries `[{label:"brief", path:"/x/brief.md"}, {label:"", path:"/x/notes.md"}]`
- When the executor runs the node
- Then `OutputPaths = {"brief":"/x/brief.md", "file[1]":"/x/notes.md"}`
- And the node status transitions to `complete`

AC-3: Missing file on disk does not fail the node
- Given an entry whose path does not exist on disk
- When the executor runs the node
- Then the entry is skipped (no `OutputPaths` key)
- And `NodeArtifactEvent.Missing` lists the label
- And node status is still `complete`

AC-4: Duplicate labels rejected
- Given two entries with `label:"brief"`
- When the executor validates the config
- Then the node fails with a wrapped `ErrDuplicateMultiFileLabel`

AC-5: Registry surface
- Given the process registry
- When `ListProcesses` is called
- Then the MultiFileLoader entry appears with `Phase = PhaseUtilities`

## BDD Test Scenarios

```gherkin
Feature: MultiFileLoader node — backend

  Scenario: Labeled + positional mix
    Given a multiFileLoader with entries [{label:"brief", path:"/x/a.md"}, {path:"/x/b.md"}]
    And both files exist
    When the executor runs the node
    Then OutputPaths = {"brief":"/x/a.md", "file[1]":"/x/b.md"}

  Scenario: Missing file tolerated
    Given a multiFileLoader with entries [{label:"brief", path:"/x/missing.md"}]
    When the executor runs the node
    Then the node completes
    And OutputPaths has no "brief" key
    And NodeArtifactEvent.Missing contains "brief"

  Scenario: Duplicate label rejected
    Given entries with two labels both equal to "brief"
    When the executor validates config
    Then the node fails
    And the error is wrapped with ErrDuplicateMultiFileLabel

  Scenario: Relative path resolution
    Given an entry {path:"notes/brief.md"} and repoPath "/Users/x/proj"
    And /Users/x/proj/notes/brief.md exists
    When the executor runs the node
    Then OutputPaths["file[0]"] equals "/Users/x/proj/notes/brief.md"
```

## Tasks / Subtasks

- [x] Task 1: Declare `NodeTypeMultiFileLoader` (AC-1)
  - [x] NodeType const + MultiFileEntry struct in `types.go`
  - [x] `ErrDuplicateMultiFileLabel` + `ErrMultiFileTooMany` sentinels
- [x] Task 2: Synchronous executor branch (AC-2, AC-3, AC-4)
  - [x] `executeMultiFileLoader` mirrors control/transform node pattern (no tmux)
  - [x] JSON-parses `Config["entries"]`; validates dup labels + count ≤ 64
  - [x] Per entry: resolve relative to repoPath with `..`-escape containment; stat; populate OutputPaths or mark missing
  - [x] Emit NodeArtifactEvent with Paths + Missing
  - [x] Failure surfaces in `NodeOutputs[nodeID]` for frontend consumption
- [x] Task 3: Registry entry (AC-5)
  - [x] `util-multi-file-loader` ProcessDef under PhaseUtilities
- [x] Task 4: Tests (AC-1 through AC-4)
  - [x] `executor_multifileloader_test.go` — 7 tests (AC1-5 + too-many + relative path)

## Definition of Done

- [x] All acceptance criteria pass (7/7 tests)
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ code coverage (executeMultiFileLoader 93.8%, package 91.2%)
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `/simplify` run on all modified code
- [x] Code review: no CRITICAL/HIGH issues (coderabbit PASS; HIGH path-traversal advisory addressed with filepath.Rel containment)
