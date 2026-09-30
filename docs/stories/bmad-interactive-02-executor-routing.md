# bmad-interactive-02: Executor routing for interactive nodes

**Status:** done
**Domain:** backend
**Size:** M
**Depends On:** bmad-interactive-01
**Priority:** P0-critical

## Story

As the BMAD executor, I want to route nodes to an interactive execution path based on their `ProcessDef.Mode`, so that interactive processes can own their own lifecycle without disturbing autonomous nodes.

## Description

Add the routing switch in `runDynamic` and a stub `executeInteractiveNode` in `internal/bmad/executor.go` per §5.1 of the design spec. In this story the interactive path is a no-op skeleton: resolve declared inputs (file/upstream/env/registry — no user suspension yet), build the prompt, start the tmux session via existing helpers, wait for idle, capture output, and call `completeNode`. User-sourced specs and gates are handled in S3 and S4.

### Scope summary
- Branch `runDynamic`'s per-ready-node dispatch on `ProcessDef.Mode`.
- Add `executeInteractiveNode(ctx, state, nodeIndex, nodeID, repoPath, model)` with the steps A/B/C/E/F from §5.2 but without user-suspension (step D omitted — gate/rounds land in S4).
- Reuse existing `startSession`, `waitRoundStable`-equivalent (idle polling), `captureRoundOutput` primitives — name them consistently with spec §5.2 (`buildInteractivePrompt`, `verifyOutputs`).
- Assertion log in `activeOutEdges` per §5.4.
- All paths with empty `InputSpecs` and `Mode == ""` fall through to existing `executeNode` — byte-for-byte.

### Non-goals
- No `suspendForSpec`, `RespondToInput`, or `PendingPrompt` emission (S3).
- No gate evaluation or round loop (S4).
- No snapshot persistence changes (S5).
- No registry entries use the new mode yet (S7).

## Developer Notes

### Files to modify
- `internal/bmad/executor.go` — add the switch in `runDynamic`, add `executeInteractiveNode`, add `buildInteractivePrompt`, add `verifyOutputs`, add defensive log in `activeOutEdges`.
- `internal/bmad/executor_interactive_test.go` (new) — table tests for routing + no-op happy path.

### Routing site
Per skill lookup, `runDynamic` is at `internal/bmad/executor.go` line ~25481. The per-ready-node dispatch currently goes straight to `go e.executeNode(...)`. Insert:

```go
proc, _ := ProcessByID(state.nodes[idx].ProcessID)
switch proc.Mode {
case InteractGuided, InteractIterative, InteractParty:
    go e.executeInteractiveNode(ctx, state, nodeIndex, nodeID, repoPath, model)
default:
    go e.executeNode(ctx, state, nodeIndex, nodeID, repoPath, model)
}
```

Control flow nodes (condition/merge/loop/loopUntil) must still route through `executeControlNode` / `executeLoopNode` — the Mode switch sits *inside* the existing NodeType branch that already dispatches process nodes.

### `executeInteractiveNode` skeleton (this story)
Mirror §5.2 steps A, B (without user suspension), C, E, F:

```go
func (e *Executor) executeInteractiveNode(
    ctx context.Context, state *execState,
    nodeIndex map[string]int, nodeID, repoPath, model string,
) {
    idx := nodeIndex[nodeID]
    proc, _ := ProcessByID(state.nodes[idx].ProcessID)

    e.setStatus(state, idx, NodeRunning)

    resolved, missing, err := e.resolveInputs(ctx, state, nodeID, 1)
    if err != nil { e.failNode(state, idx, nodeID); return }
    if len(missing) > 0 {
        // S3 will suspend; for S2 treat user-missing as a fail.
        e.failNode(state, idx, nodeID); return
    }

    prompt := buildInteractivePrompt(proc, resolved)
    if err := e.startSession(state, nodeID, prompt, model); err != nil {
        e.failNode(state, idx, nodeID); return
    }

    if err := e.waitForIdle(ctx, state, nodeID); err != nil {
        e.failNode(state, idx, nodeID); return
    }

    e.captureRoundOutput(state, nodeID, nodeID) // single-round stub key

    if err := e.verifyOutputs(state, nodeID, proc.OutputSpecs, repoPath); err != nil {
        e.failNode(state, idx, nodeID); return
    }
    e.completeNode(state, idx, nodeID)
}
```

### `resolveInputs` (partial)
This story ships `resolveInputs` for the four non-user sources (`InputFromFile`, `InputFromUpstream`, `InputFromEnv`, `InputFromRegistry`) per §4. `InputFromUser` returns the spec in the `missing` list; S3 turns that into a suspension. Keep the function signature from §4:

```go
func (e *Executor) resolveInputs(ctx context.Context, state *execState, nodeID string, round int) (resolvedInputs, []InputSpec, error)
```

Supporting helpers to add:
- `firstDirectPredecessor(state, nodeID) string`
- `truncate(s string, n int) string`
- `envValue(state, id string) string` — stub returning `""` except for `"branch"`, `"head"` which can be wired later.
- `registryLookup(ref string) (string, error)` — supports `registry:<file>#<column>` and `registry:<file>?random=N` (CSV read). Rejects any `ref` not starting with `registry:` (security §14.4 — fully enforced in S3).

### `buildInteractivePrompt`
Renders the existing BMAD system prompt layout, interpolating `resolvedInputs` values by spec.ID into a markdown block. Keep the layout identical to the autonomous path's context-building so claude is unsurprised — S7 registry entries rely on this format.

### `verifyOutputs`
Loops over `OutputSpec`. For `OutputToFile` or `OutputToBoth`, call `ResolveArtifactPath(artifactName, repoPath)` and check `os.Stat`. If missing and `!spec.Optional`, return an error. Memory-only outputs are always satisfied — the captured `NodeOutputs[nodeID]` covers them.

### `activeOutEdges` assertion
Per §5.4, insert a `log.Printf` when called on a non-complete node:

```go
if status := state.nodes[state.nodeIndex[nodeID]].Status; status != NodeComplete {
    log.Printf("bmad: activeOutEdges called for node %s in status %s", nodeID, status)
}
```

This is diagnostic only — do not change return semantics.

### Risks / gotchas
- **Dead code fear**: S2 lands without any registry entry that uses `Mode != ""` — routing is exercised only by tests. That is intentional (§11.2 staged rollout); S7 flips the registry.
- **Goroutine leak**: `executeInteractiveNode` must always call exactly one of `failNode` or `completeNode` before returning. Missing that path causes ready-set stall.
- **NodeOutputs write**: existing `captureRoundOutput` expects a key — single-round stub uses `nodeID` directly to remain compatible with the autonomous `NodeOutputs[nodeID]` convention. S4 switches to `nodeID-round-N`.
- **Tmux pane naming**: reuse `BuildSessionName` (already in executor.go) — do not invent new session naming.
- `ProcessByID` returns `(ProcessDef, bool)`; treat `false` as `failNode` — never continue with a zero-value `ProcessDef`.

### Reference files
- `internal/bmad/executor.go` — `runDynamic`, `executeNode`, `completeNode`, `failNode`, `setStatus`, `activeOutEdges`.
- `internal/bmad/artifacts.go` — `ResolveArtifactPath`, `artifactPaths` map.
- Existing registry entries in `internal/bmad/registry.go` — do not touch.

## Acceptance Criteria

**AC-1: Routing switch dispatches autonomous nodes unchanged**
- Given a workflow containing only process nodes whose `ProcessDef.Mode == ""`
- When the executor advances the ready-set
- Then every node goes through `executeNode` (the existing path)
- And zero tmux invocations differ from the pre-S2 baseline (verified by mock command runner capture)

**AC-2: Routing switch dispatches `InteractGuided`/`InteractIterative`/`InteractParty` to `executeInteractiveNode`**
- Given a test-only `ProcessDef` with `Mode = InteractGuided` and empty `InputSpecs`
- When the executor picks up the node from the ready-set
- Then `executeInteractiveNode` is entered (verified via instrumentation counter or log tap)
- And `executeNode` is not entered for that node

**AC-3: `executeInteractiveNode` happy path completes a node with no user inputs**
- Given a `ProcessDef` with `Mode = InteractGuided`, no user-sourced `InputSpecs`, one `OutputSpec` with `Target = OutputToMemory`
- When `executeInteractiveNode` runs end-to-end against the mock command runner
- Then the node transitions `pending → running → complete`
- And `NodeOutputs[nodeID]` is populated from the captured round output
- And downstream nodes begin execution

**AC-4: `resolveInputs` resolves non-user sources and reports missing user specs**
- Given an `InputSpec` table with one `file` (present), one `upstream` (present via `NodeOutputs`), one `user` (required)
- When `resolveInputs` runs
- Then the returned `resolved` map contains the file and upstream values
- And the returned `missing` slice contains the single user spec
- And no error is returned

**AC-5: `verifyOutputs` fails when a required file output is absent**
- Given a `ProcessDef` with `OutputSpec{Target: OutputToFile, ArtifactName: "product-brief", Optional: false}` and the file does not exist
- When `verifyOutputs` runs
- Then it returns a non-nil error
- And the surrounding `executeInteractiveNode` calls `failNode`
- And `NodeStatus == NodeFailed`

**AC-6: `activeOutEdges` defensive log fires when invoked on non-complete status**
- Given a node whose status is `NodeRunning`
- When `activeOutEdges` is called on it (via a test harness)
- Then a log line containing `"activeOutEdges called for node"` is emitted
- And the function's return value matches the pre-S2 behaviour (no semantic change)

## BDD Test Scenarios

```gherkin
Feature: Executor routing for interactive nodes

  Scenario: Autonomous node routes through executeNode unchanged
    Given a workflow with a single process node whose Mode is ""
    When runDynamic picks the node from the ready-set
    Then executeNode is invoked exactly once
    And executeInteractiveNode is never invoked

  Scenario: Interactive-mode node routes through executeInteractiveNode
    Given a workflow with a single process node whose Mode is "iterative"
    When runDynamic picks the node from the ready-set
    Then executeInteractiveNode is invoked exactly once
    And executeNode is never invoked for that node

  Scenario: executeInteractiveNode with no user specs completes normally
    Given a ProcessDef with Mode "guided" and no InputFromUser specs
    And an upstream node output of "hello" resolvable by an InputFromUpstream spec
    When executeInteractiveNode runs
    Then the node transitions running -> complete
    And NodeOutputs for the node is non-empty

  Scenario: Required file output missing triggers failNode
    Given a ProcessDef OutputSpec target=file, artifactName "product-brief", optional=false
    And the file does not exist under _bmad-output/analysis-artifacts/product-brief.md
    When executeInteractiveNode reaches verifyOutputs
    Then failNode is called
    And the final node status is failed

  Scenario: resolveInputs collects missing user specs without error
    Given an InputSpec slice with one file, one upstream, one required user
    And the file and upstream are present
    When resolveInputs runs
    Then err is nil
    And the missing slice has exactly one entry with Source InputFromUser

  Scenario: activeOutEdges on non-complete node logs diagnostic line
    Given a node with status NodeRunning
    When activeOutEdges is invoked against that node
    Then the log output contains "activeOutEdges called for node"
```

## Tasks / Subtasks

- [ ] Task 1: Add `resolveInputs` covering file/upstream/env/registry sources (AC-4)
  - [ ] Implement `resolveInputs(ctx, state, nodeID, round)` per §4
  - [ ] Implement helpers: `firstDirectPredecessor`, `truncate`, `envValue`, `registryLookup`
  - [ ] `registryLookup` enforces the `registry:` scheme prefix (reject others)
  - [ ] User-sourced required specs append to `missing`, not resolved
- [ ] Task 2: Add `executeInteractiveNode` skeleton (AC-2, AC-3, AC-5)
  - [ ] Wire status A→B→C→E→F; fail-node on any error, complete-node on success
  - [ ] Treat `len(missing) > 0` as a fail for this story; S3 replaces with suspend
  - [ ] Call `startSession`, `waitForIdle`, `captureRoundOutput` using existing helpers
- [ ] Task 3: Add `buildInteractivePrompt` + `verifyOutputs` (AC-3, AC-5)
  - [ ] `buildInteractivePrompt(proc, resolved)` renders a markdown block with spec.ID → value interpolation
  - [ ] `verifyOutputs(state, nodeID, specs, repoPath)` checks `ResolveArtifactPath` + `os.Stat` for file targets
- [ ] Task 4: Wire routing switch in `runDynamic` (AC-1, AC-2)
  - [ ] Insert the `Mode` switch inside the process-node dispatch branch
  - [ ] Ensure control/loop dispatch is not affected
- [ ] Task 5: Add defensive log in `activeOutEdges` (AC-6)
  - [ ] Log when called on non-`NodeComplete` status
  - [ ] Do not alter return value
- [x] Task 6: Tests in `internal/bmad/executor_interactive_test.go` (all ACs)
  - [x] Table test for routing dispatch
  - [x] Happy-path interactive run with mock runner
  - [x] `resolveInputs` table: file present/missing, upstream present/missing, user missing, env, registry
  - [x] `verifyOutputs` failure case
  - [x] `activeOutEdges` log assertion (use `log.SetOutput` to a buffer in the test)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on new functions in `internal/bmad/executor.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on modified files; no CRITICAL/HIGH findings
- [ ] Existing autonomous workflow tests pass unchanged
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch
