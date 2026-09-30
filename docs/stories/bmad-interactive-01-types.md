# bmad-interactive-01: Core types for interactive BMAD processes

**Status:** done
**Domain:** backend
**Size:** M
**Depends On:** none
**Priority:** P0-critical

## Story

As a BMAD executor author, I want typed declarations for interactive inputs, outputs, iteration gates, and pending prompts, so that every downstream story (routing, suspension, persistence, frontend widgets) can share a single canonical schema without guessing at field names or semantics.

## Description

Introduce the full type vocabulary for the interactive-process schema from §3 of `docs/bmad-interactive-process-schema.md` into `internal/bmad/types.go`. Additive only: every new field defaults to a zero value that preserves current autonomous behaviour. Ship JSON round-trip tests so downstream stories can rely on stable serialisation.

### Scope summary
- Add `InputSource`, `InputShape`, `InputSpec`, `OutputTarget`, `OutputSpec`, `InteractionMode`, `GateKind`, `IterationGate`, `PendingPrompt`, `NodeInputEntry`.
- Extend `WorkflowNodeStatus` with `NodeAwaitingInput`.
- Extend `ProcessDef` with `Mode`, `InputSpecs`, `OutputSpecs`, `Gate`.
- Extend `WorkflowExecution` with `NodeRounds`, `PendingPrompts`, `NodeInputs`, `NodeInputHistory`.
- JSON round-trip tests for every new type.

### Non-goals
- No executor changes (S2).
- No suspension primitive (S3).
- No registry edits (S7).
- `InputSpec` helper methods beyond struct definition (resolver lives in S3/S4).

## Developer Notes

### Files to modify
- `internal/bmad/types.go` — add all new constants, types, fields per §3.1-§3.5. Keep existing `ProcessDef.Inputs []string` / `Outputs []string` intact (legacy, §11.1).
- `internal/bmad/types_test.go` (new) — JSON marshal/unmarshal round-trip table tests.

### Type/symbol inventory (exact names)
- Constants: `InputFromFile`, `InputFromUpstream`, `InputFromUser`, `InputFromEnv`, `InputFromRegistry`; `ShapeFree`, `ShapeChoice`, `ShapeMultiChoice`, `ShapeApproval`, `ShapeFile`, `ShapeJSON`; `OutputToFile`, `OutputToMemory`, `OutputToBoth`; `InteractAutonomous`, `InteractGuided`, `InteractIterative`, `InteractParty`; `GateUserConfirm`, `GateArtifactExists`, `GateExpression`, `GateRoundLimit`; `NodeAwaitingInput` (added to existing block).
- Types: `InputSource string`, `InputShape string`, `OutputTarget string`, `InteractionMode string`, `GateKind string`, `InputSpec struct`, `OutputSpec struct`, `IterationGate struct`, `PendingPrompt struct`, `NodeInputEntry struct`.
- Extended `ProcessDef` fields (all `omitempty`): `Mode InteractionMode`, `InputSpecs []InputSpec`, `OutputSpecs []OutputSpec`, `Gate *IterationGate`.
- Extended `WorkflowExecution` fields (all `omitempty`): `NodeRounds map[string]int`, `PendingPrompts []PendingPrompt`, `NodeInputs map[string]map[string]string`, `NodeInputHistory map[string][]NodeInputEntry`.

### JSON tag contract
Every new struct tag set must match §3 exactly — downstream stories depend on the wire names (`inputSpecs`, `outputSpecs`, `pendingPrompts`, `nodeInputs`, `nodeInputHistory`, `nodeRounds`, `mode`, `gate`, `shape`, `prompt`, `promptId`, etc.). Do not invent alternative casings.

### Risks / gotchas
- `WorkflowNodeStatus` constant block sits next to existing statuses in `types.go` — add `NodeAwaitingInput` as the value `"awaiting_input"` (snake_case, matches spec §3.5) *between* `NodeRunning` and `NodeComplete` to preserve logical ordering but the JSON value is the load-bearing piece.
- `Gate *IterationGate` is a pointer to distinguish "no gate" from "zero gate". Other new slice/map fields rely on zero-value nil for the `omitempty` no-op path.
- `PendingPrompt.CreatedAt int64` is a Unix second, not milliseconds. Matches the §5.3 `time.Now().Unix()` reference.
- `PendingPrompt.PromptID` is a stable hash of `(NodeID | InputID | Round)`; the actual hash helper `hashPendingPrompt` ships in S3 — this story only declares the field.
- Do not touch `WorkflowNode` status field type — status values are stored as `WorkflowNodeStatus` strings; adding a new constant is automatically round-trip-safe.

### Reference files
- `internal/bmad/types.go` — existing file that gains additions (found via `.wolf/anatomy.md`).
- `docs/bmad-interactive-process-schema.md` §3 — canonical definitions.

## Acceptance Criteria

**AC-1: All new constants and types compile and carry the documented JSON tags**
- Given the `internal/bmad` package after this story's changes
- When `go build ./...` runs
- Then the build succeeds with no errors
- And `go vet ./...` reports zero issues
- And every new type declared in §3 is exported from `internal/bmad`

**AC-2: `NodeAwaitingInput` status round-trips through JSON**
- Given a `WorkflowNode` with `Status = NodeAwaitingInput`
- When the node is marshalled to JSON and unmarshalled back
- Then the resulting status equals `NodeAwaitingInput`
- And the JSON string contains the literal `"awaiting_input"`

**AC-3: `ProcessDef` with zero interactive fields is byte-identical to the legacy shape**
- Given a `ProcessDef` populated with only the pre-existing fields (ID, Name, Phase, AgentRole, SkillName, Description, Inputs, Outputs, ModuleID, Version)
- When it is marshalled to JSON
- Then the JSON output does not contain keys `mode`, `inputSpecs`, `outputSpecs`, or `gate`
- And unmarshalling it back yields an equal struct

**AC-4: `InputSpec` with every shape marshals and unmarshals losslessly**
- Given table-driven inputs covering `InputFromFile`, `InputFromUpstream`, `InputFromUser` (for each of the six `InputShape` values), `InputFromEnv`, `InputFromRegistry`
- When each is marshalled to JSON and unmarshalled back
- Then every field is preserved, including `Options`, `OptionsRef`, `Validation`, `MaxLength`, `HelpText`, `Default`, `Required`

**AC-5: `WorkflowExecution` round-trips with populated interactive state**
- Given a `WorkflowExecution` with non-empty `NodeRounds`, `PendingPrompts`, `NodeInputs`, and `NodeInputHistory`
- When it is marshalled and unmarshalled
- Then the resulting struct equals the original under `reflect.DeepEqual`

**AC-6: `IterationGate` absence is distinguishable from the zero value**
- Given a `ProcessDef` with `Gate: nil`
- When marshalled to JSON
- Then the JSON output has no `gate` key
- And given a `ProcessDef` with `Gate: &IterationGate{Kind: GateRoundLimit, MaxRounds: 0}`, marshalling yields a `gate` object in the JSON output

## BDD Test Scenarios

```gherkin
Feature: Interactive BMAD types

  Scenario: NodeAwaitingInput status serialises as awaiting_input
    Given a WorkflowNode with Status NodeAwaitingInput
    When the node is marshalled to JSON
    Then the JSON contains "status":"awaiting_input"
    And unmarshalling it reconstructs the same status

  Scenario: Legacy ProcessDef round-trip is unchanged
    Given a ProcessDef with only legacy fields populated
    When marshalled and unmarshalled
    Then the JSON output contains no mode, inputSpecs, outputSpecs, or gate keys
    And the round-trip value equals the original

  Scenario: InputSpec with OptionsRef preserves dynamic-lookup pointer
    Given an InputSpec{ID:"method", Source:InputFromUser, Shape:ShapeChoice, OptionsRef:"registry:methods.csv?random=5"}
    When it is marshalled to JSON
    Then the JSON has "optionsRef":"registry:methods.csv?random=5"
    And unmarshalling preserves both Options and OptionsRef

  Scenario: IterationGate nil vs populated round-trip
    Given a ProcessDef with Gate nil
    Then marshalling yields JSON without the "gate" key
    Given a ProcessDef with Gate &{Kind:GateUserConfirm, AcceptTokens:["done"], MaxRounds:30}
    Then marshalling yields JSON with a "gate" object whose kind is "userConfirm"

  Scenario: PendingPrompt preserves PromptID hash slot
    Given a PendingPrompt with PromptID "abc123"
    When it round-trips through JSON
    Then the unmarshalled PromptID equals "abc123"
    And the Round and CreatedAt fields are preserved

  Scenario: WorkflowExecution with populated interactive maps deep-equals after round-trip
    Given a WorkflowExecution with NodeRounds, PendingPrompts, NodeInputs, NodeInputHistory populated
    When marshalled and unmarshalled
    Then reflect.DeepEqual returns true between original and decoded
```

## Tasks / Subtasks

- [ ] Task 1: Declare input/output enums and specs (AC-1, AC-4)
  - [ ] Add `InputSource` and its five constants with JSON values `"file"|"upstream"|"user"|"env"|"registry"`
  - [ ] Add `InputShape` and its six constants (`free|choice|multi|approval|file|json`)
  - [ ] Add `InputSpec` struct with every field and tag from §3.1
  - [ ] Add `OutputTarget` and three constants; add `OutputSpec` struct from §3.2
- [ ] Task 2: Declare interaction mode + iteration gate (AC-1, AC-6)
  - [ ] Add `InteractionMode` enum with `autonomous|guided|iterative|party`
  - [ ] Add `GateKind` enum with `userConfirm|artifact|expression|rounds`
  - [ ] Add `IterationGate` struct from §3.3
- [ ] Task 3: Extend `ProcessDef` and `WorkflowNodeStatus` (AC-1, AC-2, AC-3)
  - [ ] Add `NodeAwaitingInput WorkflowNodeStatus = "awaiting_input"`
  - [ ] Add `Mode`, `InputSpecs`, `OutputSpecs`, `Gate` to `ProcessDef` with `omitempty`
  - [ ] Verify no existing registry.go entries break (they stay legacy)
- [ ] Task 4: Add `PendingPrompt`, `NodeInputEntry`, extend `WorkflowExecution` (AC-5)
  - [ ] Add `PendingPrompt` struct from §3.5
  - [ ] Add `NodeInputEntry` struct from §3.5
  - [ ] Add `NodeRounds`, `PendingPrompts`, `NodeInputs`, `NodeInputHistory` to `WorkflowExecution`
- [ ] Task 5: JSON round-trip table tests in `internal/bmad/types_test.go` (AC-2, AC-3, AC-4, AC-5, AC-6)
  - [ ] Test: `NodeAwaitingInput` constant serialisation
  - [ ] Test: legacy `ProcessDef` has no new keys in JSON
  - [ ] Test: `InputSpec` per-shape table
  - [ ] Test: `WorkflowExecution` deep-equal after round-trip
  - [ ] Test: `IterationGate` nil vs populated emits different JSON

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `internal/bmad/types.go` (new code)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on modified files; no CRITICAL/HIGH findings
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch
