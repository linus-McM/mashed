# bmad-interactive-04: Iteration gate + round loop

**Status:** done
**Domain:** backend
**Size:** L
**Depends On:** bmad-interactive-01, bmad-interactive-03
**Priority:** P0-critical

## Story

As a BMAD executor, I want iterative processes to loop through user-gated rounds with a declared termination criterion, so that brainstorming / elicitation / party-mode processes can run N rounds until the user is satisfied or a configured cap is reached.

## Description

Implement the iteration gate evaluator (`checkGate`) and the round loop in `executeInteractiveNode` (§5.2 step D). Gate kinds: `GateUserConfirm`, `GateArtifactExists`, `GateExpression`, `GateRoundLimit`. Every gate respects `MaxRounds` as a hard safety ceiling. Each round captures output under key `{nodeID}-round-{N}`, emits `bmad:node:round_complete`, checks the gate, and suspends on the next-round `InputSpec` if the gate is not satisfied.

Per-round user input is identified by `ProcessDef.iterationInput()`: the single `InputSpec` flagged as the recurring prompt (`Prompt != "" && Shape != ""` and not resolved in the pre-process pass). Spec §5.2.

### Scope summary
- `checkGate(state, gate, nodeID, round) (hit bool, reason string)` per §3.3 + §5.2 D3.
- Round loop in `executeInteractiveNode`: capture → gate check → round limit check → suspend for iteration input → inject answer → repeat.
- `iterationInput()` helper on `ProcessDef`.
- `sendToSession(ctx, state, nodeID, answer)` — injects the answer into the tmux pane via existing `escapeTmuxLiteral` + `send-keys -l`.
- Emit events: `bmad:node:round_complete`, `bmad:node:gate_satisfied`, `bmad:node:round_limit`, `bmad:node:aborted` (already in S3).
- `RejectTokens` support: if the user's answer matches a reject token, emit `bmad:node:aborted` and fail the node (§3.3).

### Non-goals
- No persistence/resume work beyond the S3 snapshot hook (S5).
- No frontend round counter (S6).
- No registry entries that exercise the gate (S7).

## Developer Notes

### Files to modify
- `internal/bmad/executor.go` — add `checkGate`, expand `executeInteractiveNode` step D, add `sendToSession`.
- `internal/bmad/types.go` — add `(*ProcessDef).iterationInput() (InputSpec, bool)` method.
- `internal/bmad/gate.go` (new) — `checkGate` + `GateExpression` evaluator (reuse condition/merge expression syntax — the existing code already has an evaluator for `NodeTypeCondition.CustomExpr`).
- `internal/bmad/events.go` — add `EventRoundComplete`, `EventGateSatisfied`, `EventRoundLimit`.
- Test files: `internal/bmad/executor_gate_test.go`, `internal/bmad/executor_iteration_test.go`.

### `iterationInput()` method
```go
func (p ProcessDef) iterationInput() (InputSpec, bool) {
    for _, s := range p.InputSpecs {
        if s.Source == InputFromUser && s.Prompt != "" && s.Shape != "" && !s.Required {
            return s, true
        }
    }
    // Fallback: any user-sourced spec with Prompt != "" that was not part of the one-shot pre-process set.
    // Convention (see §5.2): exactly one per iterative process.
    return InputSpec{}, false
}
```

Per §5.2 rationale: brainstorm's `round-response`, brief's `stage-response`, party's `message`, elicitation's `method`/`apply-changes` alternating. For the first cut implement the strict rule above; revisit after S7 smoke tests if registry entries don't cleanly match.

### Round loop (exact shape)
Replace S3's linear `executeInteractiveNode` body (after step C `startSession`) with:

```go
round := 1
for {
    if err := e.waitForIdle(ctx, state, nodeID); err != nil {
        e.failNode(state, idx, nodeID); return
    }

    roundKey := fmt.Sprintf("%s-round-%d", nodeID, round)
    output := e.captureRoundOutput(state, nodeID, roundKey)
    state.exec.NodeRounds[nodeID] = round
    e.persistSnapshot(state)
    e.emit(EventRoundComplete, roundCompletePayload(state.exec.ID, nodeID, round, roundKey))

    if hit, reason := e.checkGate(state, proc.Gate, nodeID, round); hit {
        e.emit(EventGateSatisfied, gateSatisfiedPayload(state.exec.ID, nodeID, round, reason))
        break
    }
    if proc.Gate != nil && proc.Gate.MaxRounds > 0 && round >= proc.Gate.MaxRounds {
        e.emit(EventRoundLimit, roundLimitPayload(state.exec.ID, nodeID, round))
        break
    }

    nextSpec, ok := proc.iterationInput()
    if !ok { break } // single-round process

    if err := e.suspendForSpec(ctx, state, nodeID, round+1, nextSpec); err != nil {
        e.failNode(state, idx, nodeID); return
    }
    answer := state.exec.NodeInputs[nodeID][nextSpec.ID]

    // Reject-token check
    if proc.Gate != nil && containsToken(proc.Gate.RejectTokens, answer) {
        e.emit(EventAborted, abortedPayload(state.exec.ID, nodeID, "rejected by user"))
        e.failNode(state, idx, nodeID); return
    }

    if err := e.sendToSession(ctx, state, nodeID, answer); err != nil {
        e.failNode(state, idx, nodeID); return
    }
    round++
}
```

### `checkGate` (exact shape)

```go
func (e *Executor) checkGate(state *execState, g *IterationGate, nodeID string, round int) (bool, string) {
    if g == nil { return true, "no gate; single round" }
    switch g.Kind {
    case GateUserConfirm:
        last := state.exec.lastUserAnswer(nodeID) // helper: last NodeInputHistory[nodeID].Value
        if containsToken(g.AcceptTokens, last) { return true, "accept-token matched" }
        return false, ""
    case GateArtifactExists:
        // Use OutputSpec[0].ArtifactName as the target.
        proc, _ := ProcessByID(state.nodes[state.nodeIndex[nodeID]].ProcessID)
        if len(proc.OutputSpecs) == 0 || proc.OutputSpecs[0].ArtifactName == "" {
            return false, ""
        }
        path := ResolveArtifactPath(proc.OutputSpecs[0].ArtifactName, state.repoPath)
        if path == "" { return false, "" }
        if _, err := os.Stat(path); err == nil { return true, "artifact present" }
        return false, ""
    case GateExpression:
        ok, err := evaluateExpr(g.CustomExpr, state.exec.NodeOutputs, nodeID, round)
        if err != nil { return false, "" }
        return ok, "expression matched"
    case GateRoundLimit:
        if g.MaxRounds > 0 && round >= g.MaxRounds { return true, "round limit reached" }
        return false, ""
    }
    return false, ""
}

func containsToken(tokens []string, answer string) bool {
    a := strings.TrimSpace(strings.ToLower(answer))
    for _, t := range tokens {
        if strings.ToLower(strings.TrimSpace(t)) == a { return true }
    }
    return false
}
```

Note: `GateRoundLimit` is a pure ceiling — for every other kind `MaxRounds` is evaluated *after* the kind-specific check, as a safety net, via the separate `if round >= MaxRounds` branch in the loop. This matches §3.3 rationale.

### `evaluateExpr`
Reuse the existing condition-node expression evaluator from `internal/bmad/executor.go` (the current `NodeTypeCondition` path already parses `CustomExpr` against `NodeOutputs`). Hook into whatever function handles that today — name it `evaluateConditionExpr` and extract to a shared helper if needed.

### `sendToSession`
```go
func (e *Executor) sendToSession(ctx context.Context, state *execState, nodeID, answer string) error {
    node := state.nodes[state.nodeIndex[nodeID]]
    escaped := escapeTmuxLiteral(answer)
    // send-keys -l for literal, then Enter
    if _, err := e.runner(ctx, "tmux", "send-keys", "-t", node.TmuxTarget, "-l", escaped); err != nil {
        return fmt.Errorf("send-keys literal: %w", err)
    }
    if _, err := e.runner(ctx, "tmux", "send-keys", "-t", node.TmuxTarget, "Enter"); err != nil {
        return fmt.Errorf("send-keys enter: %w", err)
    }
    return nil
}
```

Reuse `escapeTmuxLiteral` from `internal/bmad/question.go`. Matches existing `RespondToQuestionLegacy` behaviour exactly.

### Risks / gotchas
- **`NodeRounds` nil map**: allocate on first round if `state.exec.NodeRounds == nil`.
- **`NodeOutputs` key collision**: existing autonomous path writes `NodeOutputs[nodeID]`; iterative path writes `NodeOutputs[nodeID-round-N]`. Ensure downstream upstream resolution still finds the right key. Recommendation: also write to `NodeOutputs[nodeID]` on the final round (gate-satisfied or round-limit) so `resolveInputs` for `InputFromUpstream` keeps working without code changes.
- **Gate evaluation after the first round only**: the first round cannot be "accepted" — the user must answer before the gate can see anything. If `iterationInput()` returns false, the loop exits after one round (single-round process).
- **Reject tokens empty by default**: most registry entries won't have them; the token check is cheap and safe.
- **`MaxRounds == 0`** means "no cap" per §3.3 — the loop only exits via gate or iteration-input exhaustion.
- **`GateExpression` expression syntax**: document that it matches the condition-node syntax. If the existing evaluator expects `NodeOutputs[...] contains "foo"` style, the same applies here.

### Reference files
- `internal/bmad/executor.go` — existing `executeInteractiveNode` (from S2/S3), `executeControlNode` (for expression evaluator borrowing).
- `internal/bmad/question.go` — `escapeTmuxLiteral`.
- `internal/bmad/artifacts.go` — `ResolveArtifactPath` for `GateArtifactExists`.

## Acceptance Criteria

**AC-1: `GateUserConfirm` exits the loop when the last answer is an accept token**
- Given a `ProcessDef` with `Gate{Kind: GateUserConfirm, AcceptTokens: ["done", "finish"]}`
- When round 3 completes and the user's last answer is `"done"`
- Then `checkGate` returns `(true, "accept-token matched")`
- And the loop breaks
- And `bmad:node:gate_satisfied` fires

**AC-2: `GateRoundLimit` + `MaxRounds=3` exits after exactly 3 rounds**
- Given a `ProcessDef` with `Gate{Kind: GateRoundLimit, MaxRounds: 3}`
- When the loop runs without any accept token
- Then after round 3 the loop exits via the MaxRounds branch
- And `bmad:node:round_limit` fires with round=3
- And `state.exec.NodeRounds[nodeID] == 3`

**AC-3: `MaxRounds` applies as a safety ceiling to every gate kind**
- Given a `ProcessDef` with `Gate{Kind: GateUserConfirm, AcceptTokens: ["done"], MaxRounds: 2}`
- When the user never sends an accept token
- Then the loop exits at round 2 via `bmad:node:round_limit` (NOT `gate_satisfied`)
- And the node transitions to `NodeComplete`

**AC-4: `GateArtifactExists` exits when the declared artifact appears on disk**
- Given a `ProcessDef` with `Gate{Kind: GateArtifactExists}` and `OutputSpecs[0].ArtifactName = "product-brief"`
- And the file at `_bmad-output/analysis-artifacts/product-brief.md` does not exist at round 1
- But exists at round 2
- When `checkGate` runs each round
- Then round 1 returns `(false, "")`
- And round 2 returns `(true, "artifact present")`

**AC-5: `bmad:node:round_complete` fires once per round with the correct output key**
- Given an iterative node running 2 rounds before the gate
- When the loop completes
- Then `bmad:node:round_complete` has been emitted exactly 2 times
- And each emission's payload `outputKey` matches `{nodeID}-round-{N}`
- And `state.exec.NodeOutputs` contains both round keys

**AC-6: Reject token aborts the node**
- Given a `ProcessDef` with `Gate{RejectTokens: ["abort", "cancel"]}`
- When the user responds with `"cancel"` to the iteration input
- Then `bmad:node:aborted` fires with reason `"rejected by user"`
- And the node transitions to `NodeFailed`
- And no further rounds execute

**AC-7: `iterationInput()` returns false for single-round guided processes → loop exits after one round**
- Given a `ProcessDef` whose `InputSpecs` contain only required user inputs resolved pre-process
- When the loop runs round 1 and reaches `iterationInput()`
- Then the helper returns `ok=false`
- And the loop exits
- And the node completes

**AC-8: `sendToSession` injects the literal answer followed by Enter**
- Given a node with `TmuxTarget = "mashed_abc:0.1"` and answer `"hello world; rm -rf /"`
- When `sendToSession` runs
- Then the mock command runner observes two tmux calls
- And the first has args `["send-keys","-t","mashed_abc:0.1","-l","hello world; rm -rf /"]` (escaped per `escapeTmuxLiteral`)
- And the second has args `["send-keys","-t","mashed_abc:0.1","Enter"]`

## BDD Test Scenarios

```gherkin
Feature: Iteration gate and round loop

  Scenario: UserConfirm gate accepts on "done"
    Given a ProcessDef with Gate kind userConfirm acceptTokens [done, finish]
    And the user's last answer is "done"
    When checkGate runs at round 2
    Then it returns hit=true reason "accept-token matched"

  Scenario: Round limit exits after MaxRounds
    Given a ProcessDef with Gate kind rounds MaxRounds 3
    And the executor loops without any accept token
    When round 3 completes
    Then the loop breaks via MaxRounds
    And event bmad:node:round_limit fires with round 3

  Scenario: MaxRounds overrides UserConfirm when user never accepts
    Given a ProcessDef with Gate kind userConfirm acceptTokens [done] MaxRounds 2
    And the user answers "keep going" both rounds
    When the executor finishes round 2
    Then event bmad:node:round_limit fires (not gate_satisfied)
    And the node completes

  Scenario: ArtifactExists gate detects file creation between rounds
    Given a ProcessDef with Gate kind artifact
    And OutputSpecs[0].ArtifactName "product-brief"
    And the artifact file does not exist at round 1
    When round 2 captures output and the file has been created
    Then checkGate at round 2 returns true
    And event bmad:node:gate_satisfied fires

  Scenario: round_complete fires once per round with the round key
    Given an iterative node that runs 2 rounds
    When the loop exits
    Then bmad:node:round_complete has fired exactly 2 times
    And the payloads carry outputKey "{nodeID}-round-1" and "{nodeID}-round-2"

  Scenario: Reject token aborts the run
    Given a ProcessDef with Gate rejectTokens [abort, cancel]
    When the user answers "cancel"
    Then event bmad:node:aborted fires with reason "rejected by user"
    And the node status is failed

  Scenario: Guided single-round process exits after one round
    Given a ProcessDef with Mode guided and no iteration InputSpec
    When the loop runs
    Then after round 1 iterationInput returns false
    And the loop exits
    And the node completes

  Scenario: sendToSession writes literal then Enter
    Given a node with TmuxTarget "mashed_abc:0.1"
    When sendToSession is called with "hello; rm -rf /"
    Then the command runner observes send-keys -l with the escaped literal
    And a following send-keys Enter call
```

## Tasks / Subtasks

- [x] Task 1: `iterationInput()` helper (AC-7)
  - [ ] Add method on `ProcessDef` in `types.go`
  - [x] Unit test covering brainstorm, party, elicitation shapes (use registry-free fixtures)
- [x] Task 2: `checkGate` evaluator (AC-1, AC-2, AC-3, AC-4)
  - [ ] Implement all four `GateKind` branches in `internal/bmad/gate.go`
  - [ ] `containsToken` case-insensitive + trimmed
  - [ ] `evaluateExpr` wired to existing condition evaluator
  - [x] Table test `executor_gate_test.go` with fixtures per kind
- [x] Task 3: Round loop in `executeInteractiveNode` (AC-1, AC-2, AC-5, AC-6, AC-7) — RED tests written
  - [ ] Replace S3 linear body with the round loop
  - [ ] Emit `EventRoundComplete` / `EventGateSatisfied` / `EventRoundLimit` at the right seams
  - [ ] Handle reject-token abort path
  - [ ] Write final-round output to `NodeOutputs[nodeID]` in addition to round-keyed output
- [x] Task 4: `sendToSession` (AC-8) — RED tests written
  - [ ] Implement with `escapeTmuxLiteral` + two `send-keys` calls
  - [x] Unit test with mock command runner
- [ ] Task 5: Event payload builders (AC-5)
  - [ ] `roundCompletePayload`, `gateSatisfiedPayload`, `roundLimitPayload` in `events.go`
- [x] Task 6: Integration test `executor_iteration_test.go` (AC-1..AC-7) — RED tests written
  - [x] 3-round happy path with `GateUserConfirm` + accept on round 3
  - [x] `MaxRounds=2/3` reach-limit exit
  - [ ] `GateArtifactExists` with fs fixture (covered by unit test in gate file)
  - [x] Reject-token abort
  - [x] Single-round guided process

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `gate.go`, new code in `executor.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on modified files; no CRITICAL/HIGH findings
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch
