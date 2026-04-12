# exec-04: `killWorkflowChainTails` terminal-state cleanup

**Status:** ready
**Domain:** backend
**Size:** S
**Depends on:** exec-03
**Phase:** 3

## Description

Add a workflow-terminal-state finalizer that kills every unique tmux session held by any node in the workflow. Because chained command nodes share a target, this naturally dedupes to one `tmux kill-session` per chain tail. Without this, long-lived panes created by the session-reuse path accumulate until `CleanupStaleSessions` at the next app startup reaps them — unacceptable for a user who runs workflows back-to-back.

`CleanupStaleSessions` stays as the safety net for crashes; this story adds the explicit happy-path kill.

## Developer Notes

- **Files to create/modify:**
  - `internal/bmad/executor.go`:
    - Add `killWorkflowChainTails(ctx context.Context, state *execState)` per plan §Phase 3 "Piece 5" full implementation
    - Wire the finalizer into the workflow runner wherever `ExecComplete` or `ExecFailed` is set — run it before the runner goroutine returns
  - `internal/bmad/cleanup.go`:
    - Confirm `bareSessionName(target)` helper already exists; if missing, add it (strips pane suffix like `:0.0`)
  - `internal/bmad/executor_test.go` (or new `executor_cleanup_test.go`):
    - Integration test: at workflow-complete, assert `tmux kill-session` is called once per unique session across all nodes
    - Integration test: at workflow-fail mid-execution, assert the kill still runs on the running node's target
    - Regression test: `CleanupStaleSessions` still behaves identically when run standalone
- **Full implementation (from plan §Phase 3 "Piece 5"):**
  ```go
  func (e *Executor) killWorkflowChainTails(ctx context.Context, state *execState) {
      state.mu.Lock()
      seen := map[string]struct{}{}
      for _, node := range state.exec.Nodes {
          if node.TmuxTarget == "" {
              continue
          }
          seen[bareSessionName(node.TmuxTarget)] = struct{}{}
      }
      state.mu.Unlock()

      for session := range seen {
          killCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
          _, _ = e.runCmd(killCtx, "tmux", "kill-session", "-t", session)
          cancel()
      }
  }
  ```
- **Why dedup by bare session name:** command nodes in a chain share one session (e.g. `bmad-abc:0.0`, `bmad-abc:0.0`), so `bareSessionName` collapses both to `bmad-abc` and `kill-session -t bmad-abc` runs once.
- **Why `context.Background()` with a 2s timeout:** the parent ctx may already be cancelled (if the workflow was killed by the user), which would block the cleanup tmux call. Use a fresh ctx scoped to 2s so cleanup always runs.
- **Risks / gotchas:**
  - Plan §Phase 3 "Piece 5" calls this out explicitly: "Run it before the runner's goroutine returns." Identify the exact terminal-state transition in the runner — likely `state.exec.Status = ExecComplete` / `ExecFailed`. Hook the finalizer directly there, not as a deferred function elsewhere.
  - Do NOT remove or modify `CleanupStaleSessions` — it remains the safety net for crashes.
  - Errors from `kill-session` are swallowed (already-dead session is fine). Log at debug level only.
  - Cleanup must run for BOTH `ExecComplete` and `ExecFailed`. Missing the failed case means a failing workflow leaks its session.
  - Dedup via `bareSessionName` is load-bearing — if `TmuxTarget` values differ only by pane suffix (`:0.0` vs `:0.1`), the raw strings would double-kill the same session.
- **Prerequisites already in place:**
  - exec-03 ships `executeCommandNode` which sets `node.TmuxTarget` for each command node in a chain.
  - `bareSessionName` helper in `internal/bmad/cleanup.go`.
  - `CleanupStaleSessions(ctx) error` as the existing startup safety net.
  - `ExecComplete` / `ExecFailed` terminal states are already defined.

## Acceptance Criteria

**AC-1: `killWorkflowChainTails` dedupes by bare session name**
- Given a workflow with three nodes whose `TmuxTarget` values are `"bmad-abc:0.0"`, `"bmad-abc:0.0"`, and `"bmad-def:0.0"`
- When `killWorkflowChainTails` runs
- Then exactly two `tmux kill-session -t ...` invocations occur
- And the targets are `bmad-abc` and `bmad-def` (bare session names)

**AC-2: Finalizer runs on `ExecComplete`**
- Given a workflow that completes successfully with one process node and one command node sharing a session
- When the runner transitions to `ExecComplete`
- Then `tmux kill-session` is called exactly once (chained nodes share the session)
- And this happens BEFORE the runner goroutine returns

**AC-3: Finalizer runs on `ExecFailed`**
- Given a workflow that fails mid-execution with a live tmux session
- When the runner transitions to `ExecFailed`
- Then `tmux kill-session` is called on the live session
- And the workflow's failure state is preserved (cleanup doesn't rewrite status)

**AC-4: Nodes without `TmuxTarget` are skipped**
- Given a workflow containing a transform node (synchronous, no tmux) and a failing command node that never reached the resolve step
- When `killWorkflowChainTails` runs
- Then only nodes with non-empty `TmuxTarget` contribute to the kill set
- And zero kill-session invocations occur if no node ever acquired a target

**AC-5: `CleanupStaleSessions` regression**
- Given the existing `CleanupStaleSessions` tests
- When they run after this story
- Then every test passes unchanged
- And `CleanupStaleSessions` still kills stale sessions at app startup

## BDD Test Scenarios (Gherkin)

```gherkin
Feature: Workflow terminal-state chain-tail cleanup

  Scenario: Dedup by bare session name
    Given a workflow with nodes targeting "bmad-abc:0.0", "bmad-abc:0.0", "bmad-def:0.0"
    When killWorkflowChainTails runs
    Then tmux kill-session is called for "bmad-abc" exactly once
    And tmux kill-session is called for "bmad-def" exactly once

  Scenario: Finalizer fires on completion
    Given a 3-node workflow that completes successfully
    When the runner reaches ExecComplete
    Then killWorkflowChainTails is invoked before the runner goroutine returns
    And the chain's tmux session is killed

  Scenario: Finalizer fires on failure
    Given a 3-node workflow that fails on the second node
    When the runner reaches ExecFailed
    Then killWorkflowChainTails is invoked
    And the running tmux session is killed
    And the workflow's final status remains ExecFailed

  Scenario: Nodes without targets are skipped
    Given a workflow with a transform node (no tmux) and a command node that failed before resolve
    When killWorkflowChainTails runs
    Then zero kill-session invocations occur

  Scenario: CleanupStaleSessions still works at startup
    Given a stale bmad session left over from a previous app run
    When CleanupStaleSessions runs at startup
    Then the stale session is killed
    And no regression from this story's changes
```

## Tasks / Subtasks

- [ ] Task 1 — Implement `killWorkflowChainTails` (AC-1, AC-4)
  - [ ] Build the dedup set via `bareSessionName` under `state.mu`
  - [ ] Iterate and kill with `context.Background()` + 2s timeout per call
  - [ ] Log swallowed errors at debug level
- [ ] Task 2 — Wire into terminal-state transitions (AC-2, AC-3)
  - [ ] Identify where `ExecComplete` is set in the workflow runner; call the finalizer before the goroutine returns
  - [ ] Same for `ExecFailed`
- [ ] Task 3 — Tests (AC-1 through AC-5)
  - [ ] Unit test for dedup
  - [ ] Integration test for ExecComplete path
  - [ ] Integration test for ExecFailed path
  - [ ] Negative test for no-target workflows
  - [ ] Regression run of existing `CleanupStaleSessions` tests

## Definition of Done

- [ ] All ACs verified by an automated test
- [ ] Coverage ≥ 80% on modified files
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -short` clean
- [ ] `/simplify` run before sign-off
- [ ] No hardcoded paths or magic numbers added
- [ ] Existing tests still pass
