# ui-ast-U4: Wire `adapter.Translate` into `suspendForSpec`; extend `PendingPrompt.Structured`; JSON flatten + gate/reject changes; snapshot-migration regression

**Status:** done
**Domain:** backend
**Size:** L
**Depends on:** ui-ast-U0, ui-ast-U2
**Priority:** P0-critical

## Story

As the BMAD executor, I want the UI AST adapter wired into `suspendForSpec` behind the `UIAdapterEnabled` feature flag, `PendingPrompt.Structured` populated for opted-in processes, the flatten-on-receipt map path implemented inside the round-loop, and the gate/reject-token checks upgraded to scan every sub-answer, so that multi-decision AST submissions flow through `RespondToInput` correctly while every non-opted-in process and every pre-U4 snapshot round-trips byte-identically.

## Description

Implements spec §5.1 (`PendingPrompt.Structured` field), §5.2 (`suspendForSpec` wiring — the adapter call happens OUTSIDE any lock, before the state mutex is acquired), §5.3.1 (flatten-on-receipt), §5.3.2 (gate + reject-token walk over every sub-answer), §5.3.3 (three named regression tests), §5.4 (`Executor.adapter` field + `WithAdapter` option). Ships the snapshot-migration regression test promised in §5.1 (line 410) — a committed pre-U4 fixture snapshot must load cleanly, re-emit `awaiting_input` without panic, and render Layer-1 UX unchanged.

### Scope summary

- `internal/bmad/types.go` — add `PendingPrompt.Structured string `json:"structured,omitempty"``.
- `internal/bmad/types.go` — add `NodeInputEntry.Round int` + `NodeInputEntry.Key string` (§5.3.1, §5.3.2).
- `internal/bmad/executor.go` — new field `adapter uiadapter.Adapter` on `Executor`; new option `WithAdapter(a)`.
- `internal/bmad/executor.go` — `suspendForSpec` adapter-call wiring per §5.2 (OUTSIDE lock, `ctx.Err()` re-check, `maxStructuredBytes` guard).
- `internal/bmad/executor.go` — flatten-on-receipt pseudocode (§5.3.1) inside the just-woke-up section immediately after `suspendForSpec` returns at `executor.go:2415`.
- `internal/bmad/gate.go` — `anyUserAnswerMatches(state, nodeID, tokens)` helper; update `checkGate` to use it.
- `app.go` — construct `uiadapter.NewDefault(cfg, logger)` and pass via `WithAdapter` when `UIAdapterEnabled == true`.
- `internal/bmad/executor_test.go` — add three named tests from §5.3.3 + the snapshot-migration regression.

### Non-goals

- No adapter package changes (U2, U3 shipped).
- No frontend (U6, U7).
- No "View raw" toggle (U8).
- No settings UI (U5).

## Developer Notes

### Files to modify

- `internal/bmad/types.go` — 3 additive fields.
- `internal/bmad/executor.go` — adapter wiring + flatten-on-receipt.
- `internal/bmad/gate.go` — `anyUserAnswerMatches`.
- `app.go` — adapter construction + `WithAdapter` call.
- `internal/bmad/executor_interactive_test.go` (extend) — regression coverage.
- `internal/bmad/testdata/snapshots/pre-u4/execution.json` (new fixture) — snapshot from `main` committed BEFORE this story lands; load at test-time.

### `PendingPrompt.Structured` field (spec §5.1, lines 397–410)

```go
// internal/bmad/types.go
type PendingPrompt struct {
    // existing fields unchanged…
    LastOutput  string `json:"lastOutput,omitempty"`

    // NEW (§5.1): serialized UIAST JSON string. Empty when adapter disabled
    // or when a process opts out via Shape. Frontend decodes on receive.
    Structured  string `json:"structured,omitempty"`
}
```

Serialized as a STRING (spec §5.1 line 408) — not a nested JSON object — so the existing snapshot round-trip in S5 needs zero changes.

### §5.2 wiring sequence (verbatim from spec lines 413–450)

Today's signature:
```go
e.suspendForSpec(ctx, state, nodeIndex, nodeID, round+1, nextSpec, lastOutput)
```

No signature change. Adapter call is OUTSIDE any lock:

```go
// 1. Adapter call — OUTSIDE any lock. Skipped when there's no upstream turn
//    (round-1 pre-claude suspension has lastOutput == "").
var structured string
if e.adapter != nil && lastOutput != "" {
    ast := e.adapter.Translate(ctx, lastOutput, proc.ID) // never nil; §4.3
    if ast.Diagnostics.CancelReason == "" {              // ctx still live
        if blob, err := json.Marshal(ast); err == nil && len(blob) <= maxStructuredBytes {
            structured = string(blob)
        }
    }
}

// 2. Re-check ctx — Translate may have taken ~3 s; the run could have been canceled.
if err := ctx.Err(); err != nil {
    return err
}

// 3. Build PendingPrompt + acquire state.mu for the atomic upsert + persist + emit.
prompt := PendingPrompt{
    // existing fields…
    LastOutput: extractModalQuestion(lastOutput),
    Structured: structured,
}
// ... existing suspend path unchanged (state.mu + snapshotMu sequence).
```

**The `round > 1` guard is removed** (spec §5.2 line 448). `lastOutput != ""` is sufficient — round-1 pre-claude suspensions have no captured output, round-2+ iteration suspensions always do. Also enables the adapter for Party Mode's first user-facing turn.

`maxStructuredBytes = 6 * 1024` — matches §3.3 (defence-in-depth).

### §5.3.1 flatten-on-receipt pseudocode (verbatim from spec lines 460–478)

Inside the just-woke-up section immediately after `suspendForSpec` returns at `executor.go:2415`, inside the existing `state.mu` critical section:

```go
// pseudocode — actual PR wires this inside the existing state.mu critical section.
if nextSpec.Shape == ShapeJSON && astStructuredInUse(state, nodeID) {
    var decoded map[string]string
    if err := json.Unmarshal([]byte(rawAnswer), &decoded); err != nil {
        // Treat as plain string — legacy behaviour.
        state.exec.NodeInputs[nodeID][nextSpec.ID] = rawAnswer
    } else {
        for key, val := range decoded {
            composite := nextSpec.ID + ":" + key
            state.exec.NodeInputs[nodeID][composite] = val
        }
        // Also store the raw blob under the bare spec ID so existing code
        // paths (upstream context builder, sendToSession) keep a well-defined
        // value to read.
        state.exec.NodeInputs[nodeID][nextSpec.ID] = rawAnswer
    }
}
```

History (`NodeInputHistory[nodeID]`) receives ONE entry per sub-answer, each tagged with its composite key in the new `NodeInputEntry.Key` field. `NodeInputEntry.Round int` also added so `lastRound` walks (§5.3.2 gate) work.

`astStructuredInUse(state, nodeID)` = helper that returns true when the process has `EnableAstAdapter == true` AND the most recent `PendingPrompt` for this `(execID, nodeID)` had a non-empty `Structured` field.

### §5.3.2 gate + reject-token changes (verbatim spec lines 483–534)

Before:
```go
answer = state.exec.NodeInputs[nodeID][nextSpec.ID]
if containsToken(proc.Gate.RejectTokens, answer) { ... }
// checkGate walks AcceptTokens vs lastUserAnswer()
```

After:
```go
bareAnswer := state.exec.NodeInputs[nodeID][nextSpec.ID]
subAnswers := collectSubAnswersForSpec(state, nodeID, nextSpec.ID) // walks "<nextSpec.ID>:" prefixed keys

for _, v := range append(subAnswers, bareAnswer) {
    if containsToken(proc.Gate.RejectTokens, v) {
        e.emit(EventAborted, abortedPayload(execID, nodeID, "rejected by user"))
        e.failNode(state, idx, nodeID)
        return
    }
}

// sendToSession still gets the bare blob (JSON) — Claude sees the raw submission.
if sErr := e.sendToSession(ctx, state, nodeID, bareAnswer); sErr != nil { ... }
```

And `checkGate`'s `lastUserAnswer` path is swapped for `anyUserAnswerMatches`:

```go
// gate.go
func anyUserAnswerMatches(state *execState, nodeID string, tokens []string) bool {
    state.mu.Lock()
    defer state.mu.Unlock()
    hist := state.exec.NodeInputHistory[nodeID]
    if len(hist) == 0 { return false }
    lastRound := hist[len(hist)-1].Round
    for i := len(hist) - 1; i >= 0 && hist[i].Round == lastRound; i-- {
        if containsToken(tokens, hist[i].Value) {
            return true
        }
    }
    return false
}
```

Swap `containsToken(gate.AcceptTokens, lastUserAnswer(...))` for `anyUserAnswerMatches(state, nodeID, gate.AcceptTokens)`. Legacy entries with `Round == 0` fall back to "most recent entry only" behaviour (the walk terminates at the first entry since all pre-U4 entries share `Round == 0`).

### §5.4 Executor lifecycle

```go
type Executor struct {
    // existing fields…
    adapter uiadapter.Adapter  // interface; nil when disabled
}

func WithAdapter(a uiadapter.Adapter) Option { ... }
```

`app.go`:
```go
var opts []bmad.Option
if cfg.UIAdapterEnabled {
    adapter := uiadapter.NewDefault(uiadapter.Config{
        Enabled: true, Model: cfg.OllamaModel, TimeoutMs: cfg.UIAdapterTimeoutMs, MaxInflight: 1,
    }, slog.Default())
    opts = append(opts, bmad.WithAdapter(adapter))
}
executor := bmad.NewExecutor(..., opts...)
```

When `UIAdapterEnabled == false`, `adapter == nil` and §5.2's `e.adapter != nil` check short-circuits.

### §5.3.3 regression coverage (verbatim test names)

- `TestPartyMode_JSONSubmission_AcceptTokenOnApprovalWidget` — AST with one `decision_group{widget.type=approval, response_key="confirm"}`; user submits `{"confirm":"done"}`; gate `AcceptTokens=["done"]` fires `EventGateSatisfied`.
- `TestPartyMode_JSONSubmission_RejectTokenInFreeWidget` — multi-decision AST, one sub-answer is `"cancel"`, `RejectTokens=["cancel"]` aborts.
- `TestPartyMode_LegacyShapeFreeUnchanged` — iteration spec still `ShapeFree`; `Structured=""`; round loop behaves exactly as pre-U4 (regression guard against U4 leaking into non-migrated processes).

Use `uiadapter.NewMock(fixedAST)` from U2's `//go:build testing` mock — wire tests with `-tags testing`.

### §5.1 snapshot-migration regression (spec lines 408–410)

Commit a fixture at `internal/bmad/testdata/snapshots/pre-u4/execution.json` captured from `main` BEFORE this story. Test:

```go
// prompts_test.go (new or extended)
func TestSnapshot_PreU4LoadRoundTrip(t *testing.T) {
    raw := readFixture(t, "snapshots/pre-u4/execution.json")
    var exec WorkflowExecution
    require.NoError(t, json.Unmarshal(raw, &exec))
    // Assertions per spec §5.1:
    for _, p := range exec.PendingPrompts {
        require.Empty(t, p.Structured, "Structured must be empty on legacy load")
    }
    // Re-emit awaiting_input without panic (via a test harness).
    require.NotPanics(t, func() {
        emitAwaitingInputForAll(&exec)
    })
    // All prior fields round-trip.
    again, _ := json.Marshal(exec)
    var roundtrip WorkflowExecution
    require.NoError(t, json.Unmarshal(again, &roundtrip))
    require.Equal(t, exec, roundtrip)
}
```

### Risks / gotchas

- **Lock ordering.** `Translate` can block up to 3 s. If it ran under `state.mu`, every other node would stall. Spec §5.2 line 419: "must not hold `state.mu` or `state.snapshotMu` for the duration." Violating this deadlocks the executor — the AC below codifies it via a deadlock-probe test.
- **`ctx.Err()` re-check.** Between `Translate` and the state mutation, a StopWorkflow could land. Without the re-check the prompt lands in a canceled run — benign but noisy in logs. Test: `TestSuspendForSpec_CancelDuringTranslate_Returns`.
- **`maxStructuredBytes` belt-and-braces.** The validator caps at 6 KiB, but this guard is defence-in-depth against bugs. Do NOT remove it.
- **`NodeInputHistory` growth.** Flatten-on-receipt now writes N entries per turn instead of 1. Ensure snapshot size doesn't balloon — the spec §8 performance table budgets "+16 KiB worst-case per round" (6 KiB AST + expanded history). A long-running (50 rounds, 8 sub-answers) session accumulates ~400 entries; the existing persistence path handles this.
- **`Round` field on legacy entries.** Default `0` is intentional — `anyUserAnswerMatches` walks back to round 0 and stops (since all legacy entries share round 0, the walk becomes "all entries since the start"). For a session that mixed pre-U4 and post-U4 entries (upgrade mid-session), this is a small behavioural delta — document in the story commit.
- **`app.go` wiring.** The `uiadapter` import is new — ensure `internal/uiadapter` doesn't accidentally pull `internal/bmad` (no import cycle; U2 ships without any bmad import).
- **Executor test `-tags testing`.** `MockAdapter` lives under `//go:build testing`. Running `go test ./internal/bmad/... -tags testing -race` is the new gate; `justfile` / CI must know.

### Reference files

- `internal/bmad/executor.go` — `suspendForSpec` (existing), post-return round-loop section (line ~2415).
- `internal/bmad/gate.go` — `checkGate`, `containsToken`, `lastUserAnswer`.
- `internal/uiadapter/adapter.go` — `Adapter.Translate` (U2).
- `internal/uiadapter/mock.go` — `MockAdapter` (U2 `//go:build testing`).
- `docs/mashed-ui-ast-schema.md` §5 (backend integration), §5.3.3 (named tests).

## Acceptance Criteria

**AC-1: `PendingPrompt.Structured` field round-trips**
- Given a `PendingPrompt` with a 4 KiB `Structured` JSON string
- When marshalled to JSON and unmarshalled
- Then the `Structured` field survives byte-for-byte
- And a `PendingPrompt` with `Structured: ""` emits no `structured` key (omitempty)
- Verified by `TestPendingPrompt_Structured_RoundTrip`

**AC-2: §5.2 adapter call occurs outside state locks (deadlock-probe)**
- Given an adapter whose `Translate` blocks for 500ms
- When two concurrent `suspendForSpec` calls race for `state.mu`
- Then neither blocks the other on `state.mu` while `Translate` is running
- And the test completes in < 700ms (not 1000ms, which would indicate serial translate)
- Verified by `TestSuspendForSpec_TranslateOutsideLock`

**AC-3: `ctx.Err()` re-check after `Translate` returns early**
- Given a context canceled 100ms into a 500ms `Translate`
- When `suspendForSpec` resumes after `Translate`
- Then it returns the context error without mutating `PendingPrompts`
- Verified by `TestSuspendForSpec_CancelDuringTranslate_Returns`

**AC-4: `round > 1` guard is absent — Party-Mode round-1 user turn gets an AST**
- Given a Party-Mode workflow whose first user turn is `suspendForSpec(round=2, lastOutput="...")`
- When the adapter is enabled and the process has `EnableAstAdapter: true`
- Then `PendingPrompt.Structured` is populated on that turn
- Verified by `TestSuspendForSpec_PartyModeRound1UserTurn_HasStructured`

**AC-5: `maxStructuredBytes` drops oversize blobs defensively**
- Given a test adapter returning an AST that marshals to 8 KiB (validator bypassed via mock)
- When `suspendForSpec` runs
- Then `PendingPrompt.Structured == ""` (dropped)
- And the modal falls back to Layer 1
- Verified by `TestSuspendForSpec_OversizeBlobDropped`

**AC-6: §5.3.1 flatten-on-receipt populates composite + bare keys**
- Given an iteration with `Shape: ShapeJSON` and a JSON answer `{"confirm":"done", "stub":"pytest"}`
- When the round-loop resumes post-`RespondToInput`
- Then `NodeInputs[nodeID]["specID:confirm"] == "done"`
- And `NodeInputs[nodeID]["specID:stub"] == "pytest"`
- And `NodeInputs[nodeID]["specID"] == <raw JSON>`
- Verified by `TestFlatten_CompositeAndBareKeys`

**AC-7: §5.3.2 reject-token matches a sub-answer and aborts**
- Given a JSON answer `{"confirm":"done", "stub":"cancel"}` and `RejectTokens: ["cancel"]`
- When the round-loop resumes
- Then `EventAborted` fires
- And `failNode` is called
- Verified by `TestPartyMode_JSONSubmission_RejectTokenInFreeWidget` (from §5.3.3)

**AC-8: §5.3.2 accept-token over any sub-answer**
- Given a JSON answer `{"confirm":"done"}` and `AcceptTokens: ["done"]`
- When `checkGate` runs via `anyUserAnswerMatches`
- Then `EventGateSatisfied` fires
- Verified by `TestPartyMode_JSONSubmission_AcceptTokenOnApprovalWidget` (from §5.3.3)

**AC-9: Legacy `ShapeFree` processes behave byte-identically to pre-U4**
- Given a process with `Mode: InteractIterative` and iteration `Shape: ShapeFree` (NOT migrated by U0)
- When the round-loop runs from start to gate-satisfied
- Then every `PendingPrompt.Structured == ""`
- And `NodeInputs` does NOT contain composite keys
- And the emitted events match the pre-U4 event trace exactly
- Verified by `TestPartyMode_LegacyShapeFreeUnchanged` (from §5.3.3)

**AC-10: §5.1 snapshot-migration regression**
- Given the committed `testdata/snapshots/pre-u4/execution.json` fixture
- When loaded via `json.Unmarshal(raw, &WorkflowExecution{})`
- Then every `PendingPrompt.Structured == ""`
- And all prior fields round-trip under `reflect.DeepEqual`
- And `awaiting_input` re-emits without panic
- Verified by `TestSnapshot_PreU4LoadRoundTrip`

**AC-11: `WithAdapter` option + `adapter == nil` short-circuit**
- Given an `Executor` constructed without `WithAdapter`
- When `suspendForSpec` runs on an `EnableAstAdapter: true` process
- Then `PendingPrompt.Structured == ""` (no adapter call attempted)
- And no panic, no error
- Verified by `TestExecutor_NilAdapter_ShortCircuits`

**AC-12: `anyUserAnswerMatches` walks the last-round window only**
- Given a `NodeInputHistory` with entries spanning rounds 1 and 2
- When `anyUserAnswerMatches` is called
- Then only round-2 entries are scanned
- And round-1 entries are ignored (prevents stale `AcceptToken` matches across rounds)
- Verified by `TestGate_AnyUserAnswerMatches_LastRoundWindow`

## BDD Test Scenarios

```gherkin
Feature: Executor wiring for UI AST adapter

  Scenario: PendingPrompt.Structured round-trips as a JSON string
    Given a PendingPrompt with a 4 KiB Structured field
    When marshalled and unmarshalled
    Then the Structured field is byte-identical

  Scenario: Translate is called outside state.mu
    Given an adapter that blocks for 500ms in Translate
    When two concurrent suspendForSpec calls race
    Then both complete within 700ms

  Scenario: Context cancel during Translate short-circuits
    Given a context canceled mid-Translate
    When suspendForSpec resumes
    Then it returns the context error and does not mutate state

  Scenario: Party Mode round-1 user turn receives an AST
    Given Party Mode with EnableAstAdapter true
    When the first user-facing turn runs (round 2 in executor terms)
    Then PendingPrompt.Structured is non-empty

  Scenario: Oversize Structured blob is dropped to empty
    Given an adapter returning an AST that marshals to 8 KiB
    When suspendForSpec runs
    Then PendingPrompt.Structured is empty

  Scenario: Flatten-on-receipt populates composite keys
    Given a ShapeJSON answer {"confirm":"done","stub":"pytest"}
    When the round-loop resumes
    Then NodeInputs has entries for specID:confirm, specID:stub, and bare specID

  Scenario: Reject-token match in sub-answer aborts node
    Given RejectTokens ["cancel"] and a sub-answer "cancel"
    When the round-loop runs
    Then EventAborted fires and the node fails

  Scenario: Accept-token match in any sub-answer satisfies gate
    Given AcceptTokens ["done"] and sub-answer "done" in {"confirm":"done"}
    When checkGate runs
    Then EventGateSatisfied fires

  Scenario: Legacy ShapeFree process is byte-identical to pre-U4
    Given a process not opted into EnableAstAdapter
    When a round completes
    Then Structured stays empty and events match the pre-U4 trace

  Scenario: Pre-U4 snapshot loads without panic
    Given the testdata/snapshots/pre-u4/execution.json fixture
    When unmarshalled
    Then all PendingPrompt.Structured are empty
    And re-emitting awaiting_input does not panic

  Scenario: Nil adapter short-circuits cleanly
    Given an Executor without WithAdapter
    When suspendForSpec runs on an opted-in process
    Then no panic and Structured stays empty

  Scenario: Gate walks only the last round's entries
    Given NodeInputHistory spanning rounds 1 and 2
    When anyUserAnswerMatches runs
    Then only round-2 values are scanned
```

## Tasks / Subtasks

- [x] Task 1: Extend types (AC-1, AC-12)
  - [x] RED: `TestPendingPrompt_Structured_RoundTrip`, `TestNodeInputEntry_RoundAndKey_RoundTrip`
  - [x] GREEN: add `Structured`, `NodeInputEntry.Round`, `NodeInputEntry.Key`
  - [x] REFACTOR: `/simplify`
- [x] Task 2: `Executor.adapter` + `WithAdapter` (AC-11)
  - [x] RED: `TestExecutor_NilAdapter_ShortCircuits`
  - [x] GREEN: add field + option
  - [x] REFACTOR: `/simplify`
- [x] Task 3: §5.2 suspendForSpec wiring (AC-2, AC-3, AC-4, AC-5)
  - [x] RED: four failing tests — outside-lock, cancel-during-translate, party-mode round-1, oversize drop
  - [x] GREEN: insert the §5.2 sequence (adapter call outside lock, ctx re-check, maxStructuredBytes guard)
  - [x] REFACTOR: `/simplify`
- [x] Task 4: §5.3.1 flatten-on-receipt (AC-6, AC-9)
  - [x] RED: `TestFlatten_CompositeAndBareKeys` + `TestPartyMode_LegacyShapeFreeUnchanged`
  - [x] GREEN: wire the pseudocode inside the post-suspend critical section
  - [x] REFACTOR: `/simplify`
- [x] Task 5: §5.3.2 gate + reject-token updates (AC-7, AC-8, AC-12)
  - [x] RED: `TestPartyMode_JSONSubmission_RejectTokenInFreeWidget` + `TestPartyMode_JSONSubmission_AcceptTokenOnApprovalWidget` + `TestGate_AnyUserAnswerMatches_LastRoundWindow`
  - [x] GREEN: add `anyUserAnswerMatches`, swap reject-token walk, update `checkGate`
  - [x] REFACTOR: `/simplify`
- [x] Task 6: Snapshot-migration regression (AC-10)
  - [x] Capture fixture from `main` → `internal/bmad/testdata/snapshots/pre-u4/execution.json`
  - [x] RED: `TestSnapshot_PreU4LoadRoundTrip`
  - [x] GREEN: confirm passes unchanged (additive schema)
- [x] Task 7: `app.go` wiring
  - [x] Add `uiadapter` import; construct adapter when `UIAdapterEnabled == true`; pass via `WithAdapter`
  - [x] No new Wails bindings
- [x] Task 8: justfile / CI gate
  - [x] Ensure `go test -tags testing ./internal/bmad/... -race` runs in CI

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ coverage on modified code in `executor.go`, `gate.go`, `types.go`, `app.go`
- [x] `go build ./...` passes (no `-tags testing`)
- [x] `go test -tags testing ./... -race` passes
- [x] `go vet ./...` passes
- [x] `/simplify` run on every modified Go file; no CRITICAL/HIGH findings
- [x] Pre-U4 snapshot fixture committed
- [x] Story status flipped to `done`
- [x] Changes committed on branch
