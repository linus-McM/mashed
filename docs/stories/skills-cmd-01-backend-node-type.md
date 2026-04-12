# skills-cmd-01: Backend `NodeTypeCommand` + fail-fast dispatch

**Status:** ready
**Domain:** backend
**Size:** S
**Depends on:** none
**Phase:** 2

## Description

Extend the BMAD node-type enum with a new `NodeTypeCommand` variant and teach `executor.executeNode` to dispatch on it. Until Phase 3 lands, the command branch intentionally fails the node with a clear error so users who pull Phase 2 alone never see a silent hang on a canvas that contains command nodes. This story is the minimal backend hook Phase 2's frontend work plugs into.

Command metadata (name, path, description) reuses the existing `WorkflowNode.Config map[string]string` slot — no new fields on `WorkflowNode` and therefore no Wails type regeneration.

## Developer Notes

- **Files to create/modify:**
  - `internal/bmad/types.go` — add `NodeTypeCommand NodeType = "command"` alongside existing `NodeTypeProcess`, `NodeTypeCondition`, etc.
  - `internal/bmad/executor.go` — wrap current `executeNode` body in a `switch node.EffectiveType()` dispatch; add a `case NodeTypeCommand:` branch that calls `e.failNode(state, idx, nodeID)` with a log line `"bmad: command nodes not yet runnable (Phase 3)"`.
  - `internal/bmad/executor_test.go` (or new `executor_command_test.go`) — table test exercising the fail-fast path.
- **Types/symbols introduced:** `NodeTypeCommand` constant only. No new struct fields.
- **Config key conventions** (used by Phase 2 frontend and Phase 3 executor, locked here):
  - `config["commandName"]` — asset `name` field (e.g. `"simplify"`)
  - `config["commandPath"]` — absolute path to the SKILL.md / command markdown file
  - `config["commandDescription"]` — first non-blank description line, may be empty
- **Risks / gotchas:**
  - **Plan §Phase 2 "Backend changes"**: the fail-fast must be `explicit` — `NodeType` zero-value would otherwise fall through to the process-node path and silently spawn a claude pane. The `switch` must default to process only when `EffectiveType()` returns `NodeTypeProcess` (or empty, for legacy).
  - Do NOT touch `WorkflowDef` JSON serialization. The config map is already JSON-stable.
- **Prerequisites already in place:**
  - `NodeType` enum and `EffectiveType()` helper exist in `internal/bmad/types.go`.
  - `e.failNode(state, idx, nodeID)` already emits the correct failure event.
  - `WorkflowNode.Config map[string]string` is already serialised round-trip via existing save/load tests.

## Acceptance Criteria

**AC-1: `NodeTypeCommand` constant exists and round-trips through WorkflowDef JSON**
- Given a `WorkflowNode` with `NodeType: NodeTypeCommand` and config keys `commandName`, `commandPath`, `commandDescription`
- When the workflow is marshalled to JSON and unmarshalled back
- Then `EffectiveType()` returns `NodeTypeCommand`
- And all three config keys survive the round-trip

**AC-2: `executeNode` dispatches command nodes to the fail-fast branch**
- Given an execution state with a single command node
- When `executeNode` is invoked for that node
- Then the node transitions to `NodeFailed`
- And `failNode` is called exactly once
- And no tmux session is spawned

**AC-3: Process nodes continue to run unchanged**
- Given an execution state with a single process node (no command nodes anywhere in the graph)
- When `executeNode` is invoked
- Then the existing process-node code path runs exactly as it did before this story (same tmux argv, same completion polling)
- And existing `internal/bmad/executor_test.go` cases still pass without modification

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Command node dispatch

  Scenario: Command node fails fast with clear error
    Given a workflow containing one node with nodeType "command"
    And the node config has commandName "simplify" and commandPath "/tmp/simplify.md"
    When the executor calls executeNode on that node
    Then the node status is "failed"
    And no "tmux new-session" command was ever invoked
    And the log output contains "command nodes not yet runnable"

  Scenario: Command node round-trips through JSON
    Given a WorkflowDef with one command node carrying all three config keys
    When the workflow is marshalled to JSON and back
    Then the resulting node's EffectiveType returns NodeTypeCommand
    And config["commandName"] equals "simplify"
    And config["commandPath"] equals "/tmp/simplify.md"
    And config["commandDescription"] equals "Review recent changes"

  Scenario: Process node dispatch is unchanged
    Given a workflow containing one process node
    When the executor calls executeNode on that node
    Then the process-node spawn path runs exactly as before the command dispatch was added
    And the CommandRunner mock observes a "tmux new-session" invocation
```

## Tasks / Subtasks

- [ ] Task 1 — Add `NodeTypeCommand` constant (AC-1)
  - [ ] Declare `NodeTypeCommand NodeType = "command"` in `internal/bmad/types.go`
  - [ ] Confirm `EffectiveType()` returns the new constant for nodes whose `NodeType == "command"`
- [ ] Task 2 — Add fail-fast dispatch in `executeNode` (AC-2, AC-3)
  - [ ] Wrap current `executeNode` body so it inspects `EffectiveType()` under `state.mu`
  - [ ] Add `case NodeTypeCommand:` branch invoking `failNode` with a log line
  - [ ] Ensure the default branch runs the existing process-node code path untouched
- [ ] Task 3 — Tests (AC-1, AC-2, AC-3)
  - [ ] Round-trip test for `NodeTypeCommand` through `WorkflowDef` JSON
  - [ ] Fail-fast test using the `CommandRunner` mock (assert zero tmux invocations, node status transitions to failed)
  - [ ] Regression assertion that a baseline process-node test still passes unchanged

## Definition of Done

- [ ] All ACs verified by an automated test (Go table-driven; no "manually verified")
- [ ] Coverage ≥ 80% on modified files (`go test -coverprofile=cover.out ./internal/bmad/... && go tool cover -func=cover.out`)
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths or magic numbers added
- [ ] Existing tests still pass
