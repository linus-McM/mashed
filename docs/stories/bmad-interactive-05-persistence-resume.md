# bmad-interactive-05: Persistence extension + restore-on-mount re-emit

**Status:** ready
**Domain:** backend
**Size:** M
**Depends On:** bmad-interactive-03
**Priority:** P1-high

## Story

As a Mashed user who closed and reopened the app mid-workflow, I want the frontend to re-open the pending input prompt automatically after `GetBmadCurrentExecution` loads, so that interactive BMAD processes survive restart without losing in-flight human context.

## Description

Extend the snapshot lifecycle so that `PendingPrompts`, `NodeInputs`, `NodeInputHistory`, and `NodeRounds` are persisted on every relevant transition (node status change, prompt upsert/remove, NodeInputs update, round increment). On restore, `(*App).GetBmadCurrentExecution` re-emits `bmad:node:awaiting_input` for each pending prompt so the frontend snackbar re-appears. Dead-pane recovery reconstructs the tmux session from `NodeInputHistory` using a recap prompt (§7.3).

### Scope summary
- Snapshot cadence hooks per §7.1 (4 triggers).
- `(*App).GetBmadCurrentExecution` extended per §7.2 to re-emit awaiting events.
- Dead-pane recovery: `resumeInteractiveNode(ctx, state, nodeID)` per §7.3.
- `renderRecap(proc, history)` — markdown Q&A renderer used by recovery.
- Snapshot format versioning: add `"version": 2` to `execution.json` when interactive fields are present; readers handle absent version as v1 (§16.5).
- Edge cases §13.5 (dead pane during awaiting), §13.9 (paused while awaiting), §13.10 (stop while awaiting — already covered by S3 but re-verified at resume).

### Non-goals
- No new event types (§6 already complete from S3/S4).
- No frontend resume UI (S6).
- No migration code for old snapshots — they load as v1 without interactive fields and behave identically.

## Developer Notes

### Files to modify
- `internal/bmad/executor.go` — verify `persistSnapshot` writes new fields (already should thanks to S1 tags); add `resumeInteractiveNode`, `paneAlive`, `renderRecap`.
- `app_bmad.go` — extend `(*App).GetBmadCurrentExecution` to re-emit `bmad:node:awaiting_input` per pending prompt.
- `internal/bmad/executor_resume_test.go` (new) — snapshot/restore + dead-pane recovery.
- `app_bmad_resume_test.go` (new) — binding-level re-emit verification.

### Snapshot triggers (§7.1)
Every S3 call site already covers (1), (2), (3). Verify and add if missing:
1. Status transition in `setStatus` — already writes.
2. `PendingPrompts` mutation — S3's `suspendForSpec` writes on enter and exit.
3. `NodeInputs` update — S3's `RespondToInput` writes after appending.
4. Round increment — S4's loop must call `persistSnapshot` after `state.exec.NodeRounds[nodeID] = round`.

Wrap all snapshot writes in a 100 ms debouncer per §15.1 if multiple writes land within the window; otherwise keep synchronous. Acceptable single-writer sequencing: per `execID` serialise via `state.snapshotMu` and drop if another write is pending within 100 ms.

### Version field
Add `Version int json:"version,omitempty"` to `WorkflowExecution`. Writers set `Version = 2` when any interactive field is non-zero. Readers tolerate `Version == 0` (legacy) and proceed — absence of `PendingPrompts` means no pending prompts to re-emit.

### `GetBmadCurrentExecution` re-emit (§7.2)

```go
// app_bmad.go
func (a *App) GetBmadCurrentExecution(repoPath string) (*bmad.WorkflowExecution, error) {
    exec, err := a.bmadExecutor.CurrentExecution(repoPath)
    if err != nil || exec == nil { return exec, err }

    for _, p := range exec.PendingPrompts {
        runtime.EventsEmit(a.ctx, "bmad:node:awaiting_input", p)
    }
    return exec, nil
}
```

Do NOT re-emit `bmad:node:input_resolved` or `bmad:node:round_complete` — those are transitional events; snapshotting only the awaiting state is sufficient for UI restore.

### Dead-pane recovery (§7.3)

```go
func (e *Executor) resumeInteractiveNode(ctx context.Context, state *execState, nodeID string) error {
    idx := state.nodeIndex[nodeID]
    node := state.nodes[idx]

    if e.paneAlive(node.TmuxTarget) { return nil }

    proc, ok := ProcessByID(node.ProcessID)
    if !ok { return ErrProcessNotFound }

    recap := renderRecap(proc, state.exec.NodeInputHistory[nodeID])
    resolved, _, err := e.resolveInputs(ctx, state, nodeID, state.exec.NodeRounds[nodeID]+1)
    if err != nil { return err }

    prompt := buildInteractivePrompt(proc, resolved) + "\n\n## Previous session recap\n" + recap
    if err := e.startSession(state, nodeID, prompt, state.model); err != nil {
        return err
    }
    return nil
}
```

Hook `resumeInteractiveNode` into the executor's existing restart-on-mount path (wherever `CurrentExecution` re-registers a paused `execState`). Exact call site: right after the state is reconstructed from the snapshot, before `runDynamic` resumes processing.

### `paneAlive`
```go
func (e *Executor) paneAlive(target string) bool {
    if target == "" { return false }
    out, err := e.runner(context.Background(), "tmux", "display-message", "-t", target, "-p", "#{pane_id}")
    return err == nil && len(strings.TrimSpace(string(out))) > 0
}
```

### `renderRecap`
```go
func renderRecap(proc ProcessDef, history []NodeInputEntry) string {
    var b strings.Builder
    for _, e := range history {
        fmt.Fprintf(&b, "**Round %d — %s:** %s\n\n", e.Round, e.InputID, e.Value)
    }
    return b.String()
}
```

### Risks / gotchas
- **Re-emit ordering**: `GetBmadCurrentExecution` must emit AFTER returning the execution to avoid frontend races (the frontend subscribes on mount; emit before subscription = lost event). Use `time.AfterFunc(50ms, ...)` or emit on the next goroutine tick after returning. Simpler: emit inside a `go func()` launched from `GetBmadCurrentExecution` so the return happens first and the JS-side subscription has time to settle.
- **Pane-alive check in S3's suspension loop**: §13.5 says pane liveness should be checked every 5 s during `suspendForSpec`. For this story, add a `time.Ticker` inside `suspendForSpec`'s `select` block that fires a `paneAlive` check; if dead, emit `bmad:node:failed` (not aborted — user did not reject) and return `ErrPaneDead`. Keep this optional if it complicates S3 scope — call out in the PR description.
- **Snapshot race on resume**: restoring `PendingPrompts` then the suspended goroutine running again. On restore, do NOT re-enter `suspendForSpec` automatically — the pending prompt sits in snapshot, waiting for the user to call `RespondToInput`. The executor's normal `runDynamic` picks up running status and the goroutine that called `suspendForSpec` is gone; restore logic must re-spawn the goroutine waiting on the channel.

### Resume goroutine spawn
When `CurrentExecution` rebuilds state for an execution that has pending prompts, the executor must re-launch one goroutine per pending prompt that blocks on the per-waiter channel, re-emits `EventAwaitingInput`, and eventually calls `releaseWaiter` via `RespondToInput` exactly as S3. Implementation sketch:

```go
func (e *Executor) rehydratePending(state *execState) {
    for _, p := range state.exec.PendingPrompts {
        go func(p PendingPrompt) {
            idx := state.nodeIndex[p.NodeID]
            waitCh := state.waiter(p.NodeID, p.InputID)
            select {
            case <-waitCh:
                state.mu.Lock()
                state.nodes[idx].Status = NodeRunning
                state.exec.PendingPrompts = removePrompt(state.exec.PendingPrompts, p.NodeID, p.InputID)
                state.mu.Unlock()
                e.persistSnapshot(state)
                // Continue the node's run loop from where it stopped — requires re-invoking
                // executeInteractiveNode from the correct round. For S5, restart from current round.
            }
        }(p)
    }
}
```

Full "resume the node's goroutine from round N" logic is complex; simplest initial implementation: treat a restored pending prompt as "the node is parked; on answer, call `resumeInteractiveNode` which either attaches to a live pane or reconstructs a dead one with the recap". Gated behind the dead-pane check above.

### Reference files
- `internal/bmad/executor.go` — `execState`, `persistSnapshot`, `CurrentExecution`.
- `app_bmad.go` — `(*App).GetBmadCurrentExecution` docstring confirms it's the restore hook.
- `docs/bmad-interactive-process-schema.md` §7 and §13.5 for dead-pane semantics.

## Acceptance Criteria

**AC-1: Snapshot writes capture all four triggers**
- Given an interactive node executing
- When the node transitions `running → awaiting_input`
- Then the on-disk `execution.json` reflects `status=awaiting_input` and a `PendingPrompts` entry
- When the user calls `RespondToInput`
- Then the snapshot reflects the new `NodeInputs` map and `NodeInputHistory` append
- When the round loop increments `NodeRounds[nodeID]`
- Then the snapshot reflects the new round count

**AC-2: `GetBmadCurrentExecution` re-emits `bmad:node:awaiting_input` for every pending prompt**
- Given a snapshot with `PendingPrompts = [{NodeID: "n1", InputID: "topic", ...}]` on disk
- When `(*App).GetBmadCurrentExecution(repoPath)` is called
- Then the returned execution contains the pending prompt
- And within 500 ms of return, an `awaiting_input` event is emitted on the Wails bus
- And the event payload equals the `PendingPrompt` struct from the snapshot

**AC-3: Snapshot version field tags interactive executions as v2**
- Given a `WorkflowExecution` with at least one `PendingPrompt` or non-empty `NodeInputHistory`
- When the snapshot is written
- Then the JSON output contains `"version": 2`
- And given a snapshot with all interactive fields empty, the JSON either omits `version` or sets it to 0 (for byte-compat with legacy)

**AC-4: Legacy snapshots (v1, no interactive fields) load without error**
- Given an `execution.json` on disk with no `version` key, no `pendingPrompts`, no `nodeInputs`
- When `CurrentExecution` reads it
- Then the resulting `WorkflowExecution` has `Version == 0` and empty interactive fields
- And `GetBmadCurrentExecution` does not emit any `awaiting_input` events

**AC-5: Dead-pane recovery rebuilds the session with a recap**
- Given a restored execution whose node `n1` has `TmuxTarget = "dead_session:0"` (not alive)
- And `NodeInputHistory[n1]` contains 3 prior round entries
- When `resumeInteractiveNode(ctx, state, "n1")` runs
- Then `paneAlive` returns false
- And a new tmux session is started via `startSession`
- And the prompt passed to `startSession` contains `"## Previous session recap"` followed by the rendered Q&A

**AC-6: `renderRecap` formats history as markdown Q&A**
- Given a `[]NodeInputEntry` with entries `(Round=1, InputID="topic", Value="AI search")` and `(Round=2, InputID="round-response", Value="pivot to voice")`
- When `renderRecap` runs
- Then the output contains the literal lines `**Round 1 — topic:** AI search` and `**Round 2 — round-response:** pivot to voice`

**AC-7: Stop-while-awaiting produces a clean aborted state in the next restore**
- Given an execution with a pending prompt when `StopBmadWorkflow` is called
- When the app restarts and `CurrentExecution` loads the snapshot
- Then the execution status is `ExecFailed` (from the S3 ctx-cancelled path)
- And no `awaiting_input` event is re-emitted for the aborted prompt
- And `GetBmadCurrentExecution` returns nil (terminal executions are skipped by its existing filter)

## BDD Test Scenarios

```gherkin
Feature: Persistence and resume for interactive processes

  Scenario: Snapshot written on every interactive transition
    Given an interactive node running
    When it transitions to awaiting_input
    Then execution.json contains the pending prompt entry
    When RespondToInput is called with a valid answer
    Then execution.json reflects the new NodeInputs and NodeInputHistory
    When the round loop advances to round 2
    Then execution.json has NodeRounds[nodeID] == 2

  Scenario: Restore re-emits awaiting_input
    Given a snapshot on disk with one pending prompt
    When (*App).GetBmadCurrentExecution is called
    Then the returned execution has that pending prompt
    And within 500ms a bmad:node:awaiting_input event fires with matching payload

  Scenario: v2 version field on interactive snapshots
    Given a WorkflowExecution with a non-empty PendingPrompts slice
    When the snapshot is written
    Then the JSON contains "version":2

  Scenario: Legacy v1 snapshots load unchanged
    Given an execution.json with no version, no pendingPrompts, no nodeInputs
    When CurrentExecution reads it
    Then the struct loads with zero interactive fields
    And GetBmadCurrentExecution does not emit awaiting_input

  Scenario: Dead pane recovery rebuilds with recap
    Given a restored node whose TmuxTarget is dead
    And NodeInputHistory for that node has 2 round entries
    When resumeInteractiveNode runs
    Then paneAlive returns false
    And startSession is called with a prompt containing "## Previous session recap"

  Scenario: renderRecap formats entries as Q&A markdown
    Given a NodeInputHistory slice [{Round:1, InputID:"topic", Value:"AI"},{Round:2, InputID:"msg", Value:"pivot"}]
    When renderRecap runs
    Then the output contains "**Round 1 — topic:** AI"
    And the output contains "**Round 2 — msg:** pivot"

  Scenario: Stopped executions do not re-emit awaiting_input
    Given a snapshot of a failed execution whose PendingPrompts slice was populated before abort
    When the app restarts and loads state
    Then no awaiting_input event fires
    And GetBmadCurrentExecution returns nil for a terminal execution
```

## Tasks / Subtasks

- [ ] Task 1: Verify + extend snapshot triggers (AC-1)
  - [ ] Confirm `persistSnapshot` is called from `setStatus`, `suspendForSpec`, `RespondToInput`, round loop
  - [ ] Add any missing call sites flagged by the audit
  - [ ] Add 100 ms debouncer (optional; implement if contention becomes an issue)
- [ ] Task 2: Version field (AC-3, AC-4)
  - [ ] Add `Version int` to `WorkflowExecution`
  - [ ] Writer sets `Version = 2` when any interactive field is non-zero
  - [ ] Reader tolerates absence (`Version == 0`)
- [ ] Task 3: `(*App).GetBmadCurrentExecution` re-emit (AC-2, AC-7)
  - [ ] Loop `PendingPrompts` after returning; emit via `runtime.EventsEmit`
  - [ ] Use `go func()` so emit happens after return
  - [ ] Do not emit for terminal executions (the existing filter handles this)
- [ ] Task 4: `paneAlive` + `resumeInteractiveNode` (AC-5)
  - [ ] `paneAlive` via `tmux display-message`
  - [ ] `resumeInteractiveNode` reconstructs prompt + starts session when pane dead
  - [ ] Wire into the existing restore path
- [ ] Task 5: `renderRecap` (AC-6)
  - [ ] Markdown Q&A formatter
  - [ ] Unit test with multi-round fixture
- [ ] Task 6: `rehydratePending` goroutine spawn (AC-2, AC-5)
  - [ ] Launch one waiter goroutine per pending prompt on restore
  - [ ] Channel release flows identically to S3
- [ ] Task 7: Tests (all ACs)
  - [ ] `executor_resume_test.go` — snapshot round-trip, re-emit, dead-pane recap
  - [ ] `app_bmad_resume_test.go` — Wails binding re-emit timing

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `internal/bmad/executor.go` resume paths
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on modified files; no CRITICAL/HIGH findings
- [ ] Legacy snapshot fixtures still load cleanly
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch
