# bmad-interactive-03: Suspension primitive + RespondToInput Wails binding

**Status:** ready
**Domain:** fullstack
**Size:** L
**Depends On:** bmad-interactive-01, bmad-interactive-02
**Priority:** P0-critical

## Story

As a Mashed user running an interactive BMAD process, I want the executor to visibly pause on a declared user input, emit a typed event the frontend can render as a widget, and accept my validated answer through a new Wails binding, so that interactive processes become the single source of truth for suspension instead of the heuristic tmux question/idle detection.

## Description

Implement the suspension backbone: `suspendForSpec`, per-waiter channels, the `NodeAwaitingInput` status transition, `PendingPrompts` upsert/remove with snapshot persistence, the typed event contract (`bmad:node:awaiting_input`, `bmad:node:input_resolved`, `bmad:node:input_invalid`, `bmad:node:aborted`), the `(*Executor).RespondToInput` method and its Wails binding on `*App`. Validation and security checks (§8.4, §14.2, §14.3, §14.4) all land here.

This is the backbone for S4 (gate/loop), S5 (resume), and S6 (frontend). Must ship with complete unit + integration tests.

### Scope summary
- `suspendForSpec(ctx, state, nodeID, round, spec)` per §5.3.
- Per-waiter channels on `execState`: `waiters map[string]chan struct{}` keyed by `nodeID + "/" + inputID`.
- `upsertPrompt`, `removePrompt`, `findPendingPrompt`, `hashPendingPrompt`, `renderPrompt`, `resolveOptions` helpers.
- Snapshot persistence hook: `e.persistSnapshot(state)` must write `PendingPrompts` / `NodeInputs` / `NodeInputHistory` to `~/.mashed/workflows/{id}/execution.json`.
- `(*Executor).RespondToInput(execID, nodeID, inputID, value)` per §8.1 with validation per §8.4.
- `(*App).RespondToInput` Wails binding per §8.2.
- Legacy `RespondToQuestion` renamed/wrapped to `RespondToQuestionLegacy` (kept for autonomous fallback).
- Event emission helpers: `awaitingPayload`, `inputResolvedPayload`, `invalidPayload`.
- `executeInteractiveNode` step B replaces "fail on missing" from S2 with `suspendForSpec` loop.
- Security: `ShapeFile` path-traversal rejection (§14.2), `bmad:node:input_resolved` emits SHA-256 `valueHash` only (§14.3), `registryLookup` rejects non-`registry:` refs (§14.4).

### Non-goals
- No iteration gate or round loop (S4).
- No resume-from-snapshot logic (S5 handles `GetBmadCurrentExecution` re-emit + dead-pane recovery).
- No frontend widgets (S6).
- No registry entries activated (S7).

## Developer Notes

### Files to modify
- `internal/bmad/executor.go` — add suspension primitive, waiter map, `RespondToInput`, snapshot hook upgrades.
- `internal/bmad/types.go` — add `ErrStalePrompt`, `ErrInvalidInput`, `ErrUnknownInput`, `ErrPathOutsideRepo` sentinels next to `ErrProcessNotFound`.
- `internal/bmad/validate.go` (new) — `validateInput(spec, value)` per §8.4, `resolveFileInput(value, repoRoot)` per §14.2.
- `internal/bmad/events.go` (new or existing) — event name constants and payload builders.
- `app_bmad.go` — add `RespondToInput` binding, rename legacy to `RespondToQuestionLegacy` (keep `RespondToQuestion` as a forwarding shim).
- Test files: `internal/bmad/executor_suspend_test.go`, `internal/bmad/validate_test.go`, `app_bmad_respond_input_test.go`.

### Execution state extension
Add to `execState` (defined in `internal/bmad/executor.go` ~25357):

```go
waiters  map[string]chan struct{}  // key: nodeID + "/" + inputID
waitersMu sync.Mutex                // protects waiters map
```

Helpers on `*execState`:
- `waiter(nodeID, inputID string) chan struct{}` — lazy-create + return.
- `releaseWaiter(nodeID, inputID string)` — close-and-delete.

### Suspension primitive (exact shape)

```go
func (e *Executor) suspendForSpec(
    ctx context.Context, state *execState,
    nodeID string, round int, spec InputSpec,
) error {
    prompt := PendingPrompt{
        NodeID:    nodeID,
        InputID:   spec.ID,
        Prompt:    renderPrompt(spec, state, nodeID),
        Shape:     spec.Shape,
        Options:   resolveOptions(spec, state),
        Round:     round,
        CreatedAt: time.Now().Unix(),
        PromptID:  hashPendingPrompt(nodeID, spec.ID, round),
    }

    state.mu.Lock()
    idx := state.nodeIndex[nodeID]
    state.nodes[idx].Status = NodeAwaitingInput
    state.exec.PendingPrompts = upsertPrompt(state.exec.PendingPrompts, prompt)
    state.mu.Unlock()

    if err := e.persistSnapshot(state); err != nil { return err }
    e.emit(EventAwaitingInput, awaitingPayload(prompt))

    waitCh := state.waiter(nodeID, spec.ID)
    select {
    case <-waitCh: // normal release
    case <-ctx.Done():
        e.emit(EventAborted, abortedPayload(state.exec.ID, nodeID, "workflow stopped"))
        return ctx.Err()
    }

    state.mu.Lock()
    state.nodes[idx].Status = NodeRunning
    state.exec.PendingPrompts = removePrompt(state.exec.PendingPrompts, nodeID, spec.ID)
    state.mu.Unlock()
    if err := e.persistSnapshot(state); err != nil { return err }

    value := state.exec.NodeInputs[nodeID][spec.ID]
    e.emit(EventInputResolved, inputResolvedPayload(state.exec.ID, nodeID, spec.ID, round, value))
    return nil
}
```

### PromptID hash

```go
func hashPendingPrompt(nodeID, inputID string, round int) string {
    h := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%d", nodeID, inputID, round)))
    return hex.EncodeToString(h[:])[:16]
}
```

### RespondToInput (executor)

Follows §8.1 exactly. Key checks in order:
1. `getState(execID)` → 404-equivalent if not found (`ErrExecNotFound`).
2. `findPendingPrompt(pending, nodeID, inputID)` → `ErrNoPendingPrompt` if absent.
3. `findInputSpec(proc, inputID)` → `ErrUnknownInput`.
4. `validateInput(spec, value)` → emit `EventInputInvalid`, return wrapped error.
5. For `ShapeFile`, additionally call `resolveFileInput(value, state.repoPath)` — wraps the path-traversal check. Store the resolved absolute path as the value.
6. Append to `NodeInputs[nodeID][inputID]` + `NodeInputHistory[nodeID]`.
7. `persistSnapshot`.
8. `releaseWaiter(nodeID, inputID)` — signals the suspended goroutine.

Stale-prompt check: if caller passes a stale `PromptID` (in a future overloaded signature), return `ErrStalePrompt`. For this story, `RespondToInput` accepts `(execID, nodeID, inputID, value)` only — staleness is enforced by the pending-prompt lookup (the prompt is removed when the round advances). Edge case §13.2 handling.

### Wails binding

```go
// app_bmad.go
func (a *App) RespondToInput(execID, nodeID, inputID, value string) error {
    if a.bmadExecutor == nil { return bmad.ErrExecNotInitialized }
    return a.bmadExecutor.RespondToInput(execID, nodeID, inputID, value)
}
```

Legacy shim:
```go
func (a *App) RespondToQuestion(execID, nodeID, answer string) error {
    if a.bmadExecutor == nil { return bmad.ErrExecNotInitialized }
    return a.bmadExecutor.RespondToQuestionLegacy(execID, nodeID, answer)
}
```

Rename the current executor `RespondToQuestion` to `RespondToQuestionLegacy`. Callers inside `internal/bmad/*_test.go` that exercised it stay unchanged — they now address the legacy entry point.

### Event contract (per §6)

Constants (add to `events.go`):
```go
const (
    EventAwaitingInput = "bmad:node:awaiting_input"
    EventInputResolved = "bmad:node:input_resolved"
    EventInputInvalid  = "bmad:node:input_invalid"
    EventAborted       = "bmad:node:aborted"
)
```

Payload for `EventInputResolved` per §14.3: `{execId, nodeId, inputId, round, valueHash}` where `valueHash = sha256(value)[:16]`. The raw value is NEVER in the event bus.

### Validation (§8.4)

`validateInput(spec, value)` implements the switch exactly as written. Add these extras:
- `ShapeJSON` default `MaxLength` of 64 KiB (65536) when `spec.MaxLength == 0` (§13.4).
- `ShapeFile` uses `resolveFileInput` (absolute-path rejection) first, then regex/MaxLength.

### Security (§14)

- `resolveFileInput(value, repoRoot string) (string, error)`:
  ```go
  abs, err := filepath.Abs(value)
  if err != nil { return "", err }
  clean := filepath.Clean(abs)
  prefix := filepath.Clean(repoRoot) + string(filepath.Separator)
  if !strings.HasPrefix(clean, prefix) {
      return "", ErrPathOutsideRepo
  }
  return clean, nil
  ```
  Wrap with `%w` for error traversal.
- `registryLookup(ref)` (already started in S2): reject any ref whose prefix is not `registry:` — return `ErrInvalidRegistryRef`.
- `bmad:node:input_resolved` payload only contains `valueHash`, never raw value.

### Snapshot persistence

Existing snapshot write lives in the executor (S5 will trace the exact call site and add re-emit logic). For S3, `persistSnapshot(state *execState)` must:
1. Lock `state.mu.RLock()` to read.
2. Marshal `state.exec` to JSON (the new fields `PendingPrompts`, `NodeInputs`, `NodeInputHistory`, `NodeRounds` now flow through thanks to S1).
3. Write atomically to `~/.mashed/workflows/{execID}/execution.json`.

If the existing executor already has `persistSnapshot`, just ensure the new fields are marshalled — the `json` struct tags from S1 handle this automatically. If absent, introduce it.

### Risks / gotchas
- **Channel ordering**: `releaseWaiter` must be called AFTER writing `NodeInputs` so the suspended goroutine reads the value correctly when it wakes. The release races with the goroutine's status-restore block in `suspendForSpec` — acquire `state.mu` first in the wake-up code.
- **Double-respond**: if `RespondToInput` is called twice for the same `(nodeID, inputID)` while the suspension is active, the second call must return `ErrNoPendingPrompt` (prompt already removed). The first release drains the channel.
- **Context cancellation during suspend**: emit `EventAborted` with reason `"workflow stopped"` before returning `ctx.Err()`. The caller (`executeInteractiveNode`) then calls `failNode`.
- **PII in snapshot**: §14.3 acknowledges raw values live in `~/.mashed/` — do NOT redact the snapshot file.
- **`ShapeApproval` value canonicalisation**: accept `"yes"` / `"no"` as written. Do NOT accept `"true"` / `"false"` — frontend must send the exact strings.

### Reference files
- `internal/bmad/executor.go` — existing `setStatus`, `failNode`, `completeNode`, `emit`, `runDynamic`.
- `internal/bmad/question.go` — legacy `RespondToQuestion` site, `escapeTmuxLiteral` (still used by S4 when injecting round answers).
- `app_bmad.go` — existing `RespondToQuestion` binding for the shim rename.
- `app_bmad_question_test.go` — template for binding-level tests (the new `RespondToInput` binding mirrors this shape).

## Acceptance Criteria

**AC-1: Missing required user input causes the node to enter `NodeAwaitingInput` with a `PendingPrompt`**
- Given an interactive `ProcessDef` with one required `InputFromUser` spec (ID `"topic"`, Shape `ShapeFree`)
- When `executeInteractiveNode` runs and the user has not yet answered
- Then the node's status becomes `NodeAwaitingInput`
- And `state.exec.PendingPrompts` contains exactly one entry for `(nodeID, "topic")`
- And an event `bmad:node:awaiting_input` is emitted with the `PendingPrompt` payload
- And the snapshot file reflects the `PendingPrompts` entry

**AC-2: `RespondToInput` with a valid answer releases the suspension and resumes the node**
- Given a node suspended on input `"topic"` with `Shape = ShapeFree`
- When `RespondToInput(execID, nodeID, "topic", "Building a brainstorm feature")` is called
- Then validation succeeds
- Then `state.exec.NodeInputs[nodeID]["topic"] == "Building a brainstorm feature"`
- And `state.exec.NodeInputHistory[nodeID]` gains a `NodeInputEntry` with `Round=1`
- And the node status flips back to `NodeRunning`
- And `PendingPrompts` no longer contains the entry
- And `bmad:node:input_resolved` fires with a `valueHash` (SHA-256 of the value, no raw text)

**AC-3: `RespondToInput` with invalid input stays in awaiting state and emits `bmad:node:input_invalid`**
- Given a node suspended on an `InputFromUser` spec with `Shape = ShapeChoice, Options = ["a","b","c"]`
- When `RespondToInput(execID, nodeID, inputID, "z")` is called
- Then the function returns an error wrapping `ErrInvalidInput`
- And `bmad:node:input_invalid` fires with reason mentioning the allowed options
- And the node status is still `NodeAwaitingInput`
- And `PendingPrompts` still contains the entry

**AC-4: `ShapeFile` rejects paths outside the repo root**
- Given a suspended spec with `Shape = ShapeFile` and `state.repoPath = "/Users/x/repo"`
- When `RespondToInput(execID, nodeID, inputID, "/etc/passwd")` is called
- Then the function returns an error wrapping `ErrPathOutsideRepo`
- And `bmad:node:input_invalid` fires with reason `"path outside repository root"`
- And the node remains `NodeAwaitingInput`

**AC-5: `bmad:node:input_resolved` payload contains valueHash, never raw value**
- Given any successful `RespondToInput` call with value `"secret-token-abc"`
- When the resolved event is captured from the event bus
- Then the payload has a `valueHash` field equal to the first 16 hex chars of `sha256("secret-token-abc")`
- And the payload has NO `value` field
- And the payload has NO field containing the literal string `"secret-token-abc"`

**AC-6: `registryLookup` rejects non-`registry:` OptionsRef schemes**
- Given an `InputSpec` with `OptionsRef = "file:///etc/passwd"`
- When `registryLookup(spec.OptionsRef)` is invoked
- Then it returns an error wrapping `ErrInvalidRegistryRef`
- And no file-system read occurs

**AC-7: Context cancellation during suspension aborts the node cleanly**
- Given a node suspended on a required user input
- When the execution context is cancelled (e.g. `StopBmadWorkflow`)
- Then `suspendForSpec` returns `context.Canceled`
- And `bmad:node:aborted` is emitted with reason `"workflow stopped"`
- And the surrounding `executeInteractiveNode` calls `failNode`

**AC-8: Wails binding `(*App).RespondToInput` round-trips to the executor**
- Given an `App` with a live `bmadExecutor` and a node suspended on input `"approval"` (`ShapeApproval`)
- When `(*App).RespondToInput(execID, nodeID, "approval", "yes")` is invoked
- Then the executor's `RespondToInput` is called with matching arguments
- And the binding returns nil
- Given the same setup but `bmadExecutor` is nil
- Then the binding returns `bmad.ErrExecNotInitialized`

## BDD Test Scenarios

```gherkin
Feature: Suspension primitive for user inputs

  Scenario: Required user input suspends the node
    Given an interactive ProcessDef with one required InputFromUser spec "topic" shape free
    And no NodeInputs for the node
    When executeInteractiveNode runs
    Then the node status becomes awaiting_input within 500ms
    And PendingPrompts contains an entry for the node with inputId "topic"
    And event bmad:node:awaiting_input fires once

  Scenario: Valid response releases the suspension
    Given a node suspended on input "topic" shape free
    When RespondToInput(execID, nodeID, "topic", "my idea") is called
    Then the call returns nil
    And NodeInputs[nodeID]["topic"] equals "my idea"
    And NodeInputHistory[nodeID] contains an entry with round 1
    And the node status returns to running
    And PendingPrompts no longer contains the entry
    And event bmad:node:input_resolved fires with a valueHash

  Scenario: Invalid choice stays awaiting and emits input_invalid
    Given a node suspended on input "approach" shape choice options [user-pick, ai-recommend]
    When RespondToInput with value "random" is called
    Then the call returns an error
    And event bmad:node:input_invalid fires
    And the node status is still awaiting_input
    And PendingPrompts is unchanged

  Scenario: Path traversal on ShapeFile is rejected
    Given a node suspended on input "file" shape file
    And the execution repoPath is /Users/me/repo
    When RespondToInput with value "/etc/passwd" is called
    Then the call returns ErrPathOutsideRepo
    And event bmad:node:input_invalid fires with reason mentioning "path outside repository root"

  Scenario: Resolved event carries valueHash, not raw value
    Given a node suspended on input "topic" shape free
    When RespondToInput with value "secret-value-xyz" is called
    And the bmad:node:input_resolved event is captured
    Then the event payload has a valueHash string
    And the event payload does not contain "secret-value-xyz"

  Scenario: OptionsRef with non-registry scheme is rejected
    Given an InputSpec with OptionsRef "file:///etc/passwd"
    When the executor calls registryLookup on it
    Then an error wrapping ErrInvalidRegistryRef is returned
    And no file read is performed

  Scenario: Stop workflow while awaiting aborts the node
    Given a node suspended on a required user input
    When the execution context is cancelled
    Then suspendForSpec returns context.Canceled
    And event bmad:node:aborted fires with reason "workflow stopped"
    And the node status becomes failed

  Scenario: Wails RespondToInput binding proxies to the executor
    Given an App with a live bmadExecutor and a suspended node
    When the Wails binding RespondToInput is called
    Then the executor RespondToInput receives the same arguments
    And the binding returns nil
```

## Tasks / Subtasks

- [ ] Task 1: Execution-state extensions (AC-1, AC-2)
  - [ ] Add `waiters map[string]chan struct{}` and `waitersMu sync.Mutex` to `execState`
  - [ ] Implement `(*execState).waiter(nodeID, inputID)` and `releaseWaiter`
- [ ] Task 2: Prompt helpers + hash (AC-1, AC-2)
  - [ ] `upsertPrompt`, `removePrompt`, `findPendingPrompt`
  - [ ] `hashPendingPrompt(nodeID, inputID, round) string`
  - [ ] `renderPrompt(spec, state, nodeID) string` (simple template substitution for `{{...}}`)
  - [ ] `resolveOptions(spec, state) []string` — returns `spec.Options` or `registryLookup` result
- [ ] Task 3: Sentinels + validation (AC-3, AC-4, AC-6)
  - [ ] Add `ErrStalePrompt`, `ErrInvalidInput`, `ErrUnknownInput`, `ErrPathOutsideRepo`, `ErrInvalidRegistryRef`, `ErrNoPendingPrompt` to `internal/bmad/types.go`
  - [ ] Implement `validateInput(spec, value)` per §8.4 including `ShapeJSON` 64 KiB default
  - [ ] Implement `resolveFileInput(value, repoRoot)` per §14.2
  - [ ] Harden `registryLookup` to reject non-`registry:` schemes
- [ ] Task 4: Event constants + payload builders (AC-1, AC-2, AC-3, AC-5, AC-7)
  - [ ] Define `EventAwaitingInput`, `EventInputResolved`, `EventInputInvalid`, `EventAborted`
  - [ ] `awaitingPayload(prompt)` returns the `PendingPrompt` struct
  - [ ] `inputResolvedPayload(execID, nodeID, inputID, round, value)` returns payload with `valueHash = sha256(value)[:16]`, never raw
  - [ ] `invalidPayload(nodeID, inputID, reason)` and `abortedPayload(execID, nodeID, reason)`
- [ ] Task 5: `suspendForSpec` primitive (AC-1, AC-7)
  - [ ] Implement per §5.3 with ctx cancellation branch emitting `EventAborted`
  - [ ] Ensure `persistSnapshot` fires on both transitions (awaiting → running)
- [ ] Task 6: `(*Executor).RespondToInput` (AC-2, AC-3, AC-4, AC-5)
  - [ ] Find pending prompt; return `ErrNoPendingPrompt` if missing
  - [ ] Validate; emit `EventInputInvalid` on failure
  - [ ] `ShapeFile` additionally runs `resolveFileInput`
  - [ ] Store in `NodeInputs` + `NodeInputHistory`; persist snapshot
  - [ ] Call `releaseWaiter`
- [ ] Task 7: Wails binding + legacy rename (AC-8)
  - [ ] `(*App).RespondToInput` in `app_bmad.go`
  - [ ] Rename executor `RespondToQuestion` → `RespondToQuestionLegacy`
  - [ ] Keep `(*App).RespondToQuestion` as a shim forwarding to `RespondToQuestionLegacy`
- [ ] Task 8: Update `executeInteractiveNode` step B (AC-1)
  - [ ] Replace the S2 "fail on missing" with a loop: for each missing user spec, call `suspendForSpec`, then re-run `resolveInputs`
  - [ ] After suspension resolves, include user answers in `resolvedInputs`
- [ ] Task 9: Tests (all ACs)
  - [ ] `executor_suspend_test.go` — suspend happy-path, ctx cancellation, double-respond, hash stability
  - [ ] `validate_test.go` — table-driven coverage of every Shape
  - [ ] `app_bmad_respond_input_test.go` — Wails binding happy and nil-executor paths
  - [ ] Event-hash assertion test (AC-5) — verifies no raw value leaks to event bus
  - [ ] Registry-ref injection test (AC-6) — rejects `file:`, `http:`, `mcp:`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on every file modified or created
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes (races in `waiters` map are common — verify with `-race`)
- [ ] `/simplify` run on all modified files; no CRITICAL/HIGH findings
- [ ] No raw user value ever appears in `bmad:node:input_resolved` payloads (manual log inspection in one test)
- [ ] Legacy `RespondToQuestion` tests still green
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch
