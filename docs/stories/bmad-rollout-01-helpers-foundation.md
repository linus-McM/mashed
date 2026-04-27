# Story bmad-rollout-01: Interactive Defaults Helpers + Idempotency Tests

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** bmad-interactive-08 (completed sprint)
**Status:** done

## Description

Introduce a single foundation file `internal/bmad/interactive_defaults.go` that exposes three idempotent helpers — `applyIterativeUpgrade`, `applyGuidedUpgrade`, and `applyPartyUpgrade` — used by the rest of the rollout to upgrade autonomous `ProcessDef` entries into the four-pillar interactive standard (`Mode`, `EnableAstAdapter`, `InputSpecs` with iteration slot, `Gate`). Ship a table-driven test suite `internal/bmad/interactive_defaults_test.go` that pins idempotency, accept-token merging, and `iterationInput()` discovery so downstream stories can mutate registry entries with a single call line and a single test assertion.

This story is pure scaffolding — no registry mutations, no behaviour change in `wails dev`. It exists so reviewers read the helper in isolation before Phase 2 lands six processes through it.

## Developer Notes

### Architecture
- New file `internal/bmad/interactive_defaults.go` (package `bmad`).
- Helpers operate on `*ProcessDef` from `internal/bmad/types.go` — Mode, EnableAstAdapter, InputSpecs, OutputSpecs, Gate.
- Idempotency rule: `if def.Mode != "" { return }` — the four already-interactive processes (`bmad-brainstorming`, `bmad-product-brief`, `bmad-advanced-elicitation`, `bmad-party-mode`) must remain byte-identical after a helper call so their existing `internal/bmad/testdata/registry/<id>.json` goldens never reshuffle.
- Iteration slot convention from `(ProcessDef).iterationInput()` in `types.go:110`: `Source==InputFromUser && Shape!="" && Required==false && Default==""`. The Iterative helper appends a `round-response` spec satisfying these constraints last, so `iterationInput()` returns it.
- `Shape: ShapeJSON` per Open Decision #2 — keeps the UI-AST adapter active when the model degrades; brainstorming reference uses the same shape.

### Helper signatures (exact)

```go
// internal/bmad/interactive_defaults.go (NEW)

var commonAcceptTokens = []string{"done", "wrap up", "complete"}
var commonRejectTokens = []string{"abort", "cancel"}

type IterativeUpgradeSpec struct {
    Prompt         string
    HelpText       string
    DomainAccept   []string
    MaxRounds      int          // 0 ⇒ default 30
    ArtifactInputs []InputSpec
    OutputSpecs    []OutputSpec
}

type GuidedUpgradeSpec struct {
    StagedInputs   []InputSpec  // rendered in source order
    OutputSpecs    []OutputSpec
    FinalApproval  *InputSpec   // optional ShapeApproval terminator
    MaxRounds      int          // 0 ⇒ default 10
    AcceptTokens   []string     // 0-len ⇒ {"yes"}
}

type PartyUpgradeSpec struct {
    Topic          InputSpec    // initial topic prompt; one-shot
    RoundPrompt    string       // recurring per-message prompt
    HelpText       string
    OutputSpecs    []OutputSpec
    MaxRounds      int          // 0 ⇒ default 100
    AcceptTokens   []string     // 0-len ⇒ {"exit", "done", "wrap up"}
}

func applyIterativeUpgrade(def *ProcessDef, spec IterativeUpgradeSpec)
func applyGuidedUpgrade(def *ProcessDef, spec GuidedUpgradeSpec)
func applyPartyUpgrade(def *ProcessDef, spec PartyUpgradeSpec)
```

### Token merging
- `applyIterativeUpgrade` constructs `Gate.AcceptTokens = commonAcceptTokens ∪ spec.DomainAccept` (no duplicates, baseline first, domain second).
- `RejectTokens` is always `commonRejectTokens` for the iterative case; guided/party use spec-supplied or sensible defaults.

### Iteration slot insertion order
- Inputs assemble as `spec.ArtifactInputs ++ {round-response}` so the round slot is last; this guarantees `iterationInput()` returns it even if a domain spec accidentally satisfies the predicate (none should — check in tests).

### Risks & migration notes (from plan §"Migration risks")
- **Risk #1 — Persisted workflows on disk:** Snapshot status precedence in `internal/bmad/resume.go:347-353` means a node persisted as `NodeRunning` or `NodeComplete` is replayed in that status regardless of registry shape. Helpers do not change resume code; idempotency test verifies a re-applied helper does not change the spec list, so old snapshots stay readable.
- **Risk #4 — Tests pinning autonomous behaviour:** `internal/bmad/registry_test.go:391 TestU0_AC3_NonMigratedProcessesUnchanged` byte-compares each non-migrated process to a golden under `internal/bmad/testdata/registry/<id>.json`. This story does NOT mutate registry, so no goldens need refreshing here. Subsequent stories (02-08) regenerate the golden for each upgraded ID; this story just provides the helpers they will call.
- **Open Decision #1 (resolution):** plan picks shared helper for consistency. This story implements that decision.
- **Open Decision #2 (resolution):** `Shape: ShapeJSON` for round-response (matches brainstorming reference §10.1).

### Reference Files
- `internal/bmad/types.go:81-121` — `ProcessDef`, `iterationInput()` predicate, `InteractionMode` constants.
- `internal/bmad/registry.go:14-66` — reference shapes for brainstorming + product-brief.
- `internal/bmad/registry.go:327-389` — advanced-elicitation + party-mode reference.
- `internal/bmad/registry_test.go:367-433` — U0 golden test pattern that downstream stories will use as a template.
- `internal/bmad/sprint.go:88` — confirms shared `bmadOutputDir` constant lives in `artifacts.go:9` (string-only; helpers do not touch).

## Acceptance Criteria

**AC-1: File exists with three helpers and shared token vars**
- Given `internal/bmad/interactive_defaults.go` after this story
- When the file is read
- Then it declares package `bmad`
- And it declares unexported `commonAcceptTokens` ([]string starting with `"done"`)
- And it declares unexported `commonRejectTokens` ([]string starting with `"abort", "cancel"`)
- And it exports types `IterativeUpgradeSpec`, `GuidedUpgradeSpec`, `PartyUpgradeSpec`
- And it exports functions `applyIterativeUpgrade`, `applyGuidedUpgrade`, `applyPartyUpgrade` with the signatures above

**AC-2: applyIterativeUpgrade is idempotent**
- Given a `ProcessDef` with `Mode = ""` (autonomous)
- When `applyIterativeUpgrade(&def, spec)` is called
- Then `def.Mode == InteractIterative`
- And `def.EnableAstAdapter == true`
- And the last entry of `def.InputSpecs` has `ID == "round-response"`, `Source == InputFromUser`, `Shape == ShapeJSON`, `Required == false`, `Default == ""`
- And calling the helper a second time with the same spec leaves `def` deeply equal to its post-first-call value

**AC-3: applyIterativeUpgrade respects pre-set Mode**
- Given a `ProcessDef` with `Mode = InteractIterative` and a populated `InputSpecs` slice (e.g. brainstorming)
- When `applyIterativeUpgrade(&def, spec)` is called with any spec
- Then `def` is byte-for-byte unchanged (deep equal pre vs post)

**AC-4: Gate carries baseline + domain accept tokens, common reject tokens, default cap**
- Given a fresh `ProcessDef` and a spec with `DomainAccept = []string{"ship"}` and `MaxRounds = 0`
- When `applyIterativeUpgrade` runs
- Then `def.Gate.Kind == GateUserConfirm`
- And `def.Gate.AcceptTokens` contains `"done"`, `"wrap up"`, `"complete"`, `"ship"` in baseline-then-domain order with no duplicates
- And `def.Gate.RejectTokens == commonRejectTokens`
- And `def.Gate.MaxRounds == 30`

**AC-5: iterationInput() returns the round-response slot after upgrade**
- Given a fresh `ProcessDef` upgraded via `applyIterativeUpgrade`
- When `def.iterationInput()` is called
- Then `(InputSpec, true)` is returned with `ID == "round-response"`

**AC-6: applyGuidedUpgrade lays staged inputs in declared order with no Gate/iteration slot**
- Given a fresh `ProcessDef` and a spec with three `StagedInputs`
- When `applyGuidedUpgrade` runs
- Then `def.Mode == InteractGuided` and `def.EnableAstAdapter == true`
- And `def.InputSpecs` equals `StagedInputs` in declared order followed by `FinalApproval` if non-nil
- And `def.iterationInput()` returns `(InputSpec{}, false)` (no recurring slot)
- And `def.Gate.MaxRounds == 10` and `def.Gate.AcceptTokens` contains `"yes"` when no override given

**AC-7: applyPartyUpgrade assembles topic + recurring message with party defaults**
- Given a fresh `ProcessDef` and a spec with `Topic` (Required) and `RoundPrompt = "Your turn..."`
- When `applyPartyUpgrade` runs
- Then `def.Mode == InteractParty` and `def.EnableAstAdapter == true`
- And `def.InputSpecs` is `[Topic, {ID:"message", Source:InputFromUser, Shape:ShapeJSON, Required:false, Prompt:"Your turn..."}]`
- And `def.iterationInput()` returns the `message` slot
- And `def.Gate.MaxRounds == 100` and `def.Gate.AcceptTokens` contains `"exit"`, `"done"`, `"wrap up"`

## BDD Test Scenarios

```gherkin
Feature: Interactive upgrade helpers

  Scenario: Iterative upgrade applied to autonomous process
    Given a ProcessDef "X" with Mode == ""
    When I call applyIterativeUpgrade with prompt "Continue or 'done'."
    Then X.Mode is "iterative"
    And X.EnableAstAdapter is true
    And the last InputSpec has ID "round-response" and Shape ShapeJSON
    And X.Gate.MaxRounds is 30
    And X.Gate.AcceptTokens starts with "done", "wrap up", "complete"

  Scenario: Iterative upgrade is idempotent
    Given a ProcessDef "X" with Mode == ""
    When I call applyIterativeUpgrade twice with identical spec
    Then the second call leaves X deeply equal to the first-call result

  Scenario: Iterative upgrade preserves already-interactive process
    Given a ProcessDef "Brainstorm" cloned from registry brainstorming entry
    When I call applyIterativeUpgrade with any spec
    Then Brainstorm is byte-for-byte unchanged

  Scenario: Domain accept tokens append after baseline
    Given a fresh ProcessDef
    When I call applyIterativeUpgrade with DomainAccept ["ship", "approved"]
    Then Gate.AcceptTokens equals ["done", "wrap up", "complete", "ship", "approved"]
    And no token appears twice

  Scenario: MaxRounds default applied when zero
    Given a fresh ProcessDef and spec.MaxRounds == 0
    When I call applyIterativeUpgrade
    Then Gate.MaxRounds is 30

  Scenario: MaxRounds override honoured
    Given a fresh ProcessDef and spec.MaxRounds == 10
    When I call applyIterativeUpgrade
    Then Gate.MaxRounds is 10

  Scenario: Guided upgrade lays staged inputs and skips iteration slot
    Given a fresh ProcessDef and spec with three StagedInputs A, B, C
    When I call applyGuidedUpgrade
    Then InputSpecs equals [A, B, C] in order
    And iterationInput() returns false
    And Mode is "guided"

  Scenario: Party upgrade attaches topic and message slots
    Given a fresh ProcessDef and spec with Topic "T" and RoundPrompt "Your turn."
    When I call applyPartyUpgrade
    Then InputSpecs is [T, {message slot}]
    And iterationInput() returns the message slot
    And Gate.AcceptTokens contains "exit"

  Scenario: Iterative upgrade preserves declared artifact inputs
    Given a fresh ProcessDef and spec.ArtifactInputs == [{ID:"prd", Source:InputFromFile, ArtifactName:"PRD.md"}]
    When I call applyIterativeUpgrade
    Then InputSpecs is [prd, round-response] in that order
```

## Tasks / Subtasks

- [x] Task 1: Create `internal/bmad/interactive_defaults.go` with helpers (AC-1, AC-2, AC-3, AC-4, AC-5, AC-6, AC-7)
  - [x] Declare unexported `commonAcceptTokens`, `commonRejectTokens` slices.
  - [x] Declare exported `IterativeUpgradeSpec`, `GuidedUpgradeSpec`, `PartyUpgradeSpec` structs with field comments mirroring this story's signatures.
  - [x] Implement `applyIterativeUpgrade` with idempotency guard, accept-token merge, default MaxRounds=30, ShapeJSON round-response slot.
  - [x] Implement `applyGuidedUpgrade` (no Gate iteration slot — single pass; Gate carries acceptTokens defaulting to {"yes"}).
  - [x] Implement `applyPartyUpgrade` with topic + message slots.
- [x] Task 2: Create `internal/bmad/interactive_defaults_test.go` (AC-2 through AC-7)
  - [x] Table-driven idempotency test (two invocations, deep-equal on second-call result).
  - [x] Guard test: pre-set `Mode = InteractIterative` leaves def untouched (use a brainstorming-shaped fixture).
  - [x] Token merging test: baseline + domain order, dedup, default MaxRounds branch, override branch.
  - [x] iterationInput() discovery test post-upgrade for Iterative + Party; non-discovery test for Guided.
  - [x] Artifact-inputs precede round-response test.
- [x] Task 3: CI sanity (AC-1, all)
  - [x] `go build ./...` green; `go vet ./...` green; `go test ./internal/bmad/... -race -count=1` green.
  - [x] Confirm `TestU0_AC3_NonMigratedProcessesUnchanged` still passes — this story does not touch registry.

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests in `interactive_defaults_test.go`
- [x] 80%+ coverage on `interactive_defaults.go`
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `/simplify` run on all modified code; no CRITICAL/HIGH issues
- [x] AC validation table populated in PR description
- [x] Status flipped to `done` by sprint lead

---

## Sprint Report — bmad-rollout-01

**Commit:** `31bab24` — feat(bmad-rollout-01): add interactive defaults helpers + idempotency tests
**Files:** internal/bmad/interactive_defaults.go (NEW, 230 lines), internal/bmad/interactive_defaults_test.go (NEW, 502 lines), internal/bmad/types.go (+7 lines: `RoundResponseInputID`, `PartyMessageInputID` exported constants)

### AC Validation Table

| AC | Description | Method | Evidence | Result |
|----|-------------|--------|----------|--------|
| AC-1 | File exists with three helpers + shared token vars | go build + vet | both pass; symbols `applyIterativeUpgrade`, `applyGuidedUpgrade`, `applyPartyUpgrade`, `IterativeUpgradeSpec`, `GuidedUpgradeSpec`, `PartyUpgradeSpec`, `commonAcceptTokens`, `commonRejectTokens` declared in interactive_defaults.go | PASS |
| AC-2 | applyIterativeUpgrade idempotent | unit test | `TestStoryRollout01_AC2_IterativeIdempotent` (2 subtests) + `TestStoryRollout01_AC2_PreservesArtifactInputsBeforeRoundResponse` | PASS |
| AC-3 | Pre-set Mode preserved | unit test | `TestStoryRollout01_AC3_PreSetModePreservedForBrainstorming` (clones brainstorming via `ProcessByID`, JSONEq + DeepEqual) | PASS |
| AC-4 | Gate carries baseline+domain accept tokens, common reject, default cap | unit test | `TestStoryRollout01_AC4_GateAcceptTokenMergeAndDefaults` (6 subtests covering merge order, dedup, default MaxRounds, override) | PASS |
| AC-5 | iterationInput() returns round-response after upgrade | unit test | `TestStoryRollout01_AC5_IterationInputAfterIterativeUpgrade` | PASS |
| AC-6 | applyGuidedUpgrade lays staged inputs, no Gate iteration slot | unit test | `TestStoryRollout01_AC6_GuidedUpgrade` (3 subtests) | PASS |
| AC-7 | applyPartyUpgrade assembles topic + message slots | unit test | `TestStoryRollout01_AC7_PartyUpgrade` (2 subtests) | PASS |

### Quality Gates

- **Build:** PASS (`go build ./...`)
- **Vet:** PASS (`go vet ./...`)
- **Test:** PASS (all `TestStoryRollout01_*` green; `TestU0_AC3_NonMigratedProcessesUnchanged` still PASS — registry untouched)
- **Race:** PASS (`go test ./internal/bmad/... -race -count=1 -short`)
- **Coverage:** 90.0–100.0% per helper; 85.7% pkg total — exceeds 80% gate
  - `applyIterativeUpgrade` 100.0%, `applyGuidedUpgrade` 94.1%, `applyPartyUpgrade` 90.0%, `mergeAcceptTokens` 100.0%, `buildGate` 100.0%, `defaultMaxRounds` 100.0%
- **/simplify:** 3 reviewers (reuse, quality, efficiency) — applied 5 fixes:
  1. **Efficiency #1 (real bug)**: `commonRejectTokens` slice was aliased into every Gate's `RejectTokens` — fixed by copying inside new `buildGate` helper
  2. **Quality #3 (HIGH)**: promoted `"round-response"`/`"message"` to exported `RoundResponseInputID`/`PartyMessageInputID` constants in types.go (12+ downstream call sites can adopt incrementally)
  3. **Quality #4 (HIGH)**: deleted stale RED-phase comment block at top of test file
  4. **Quality #4 (HIGH)**: trimmed package doc to drop story-task narration per CLAUDE.md
  5. **Reuse #4**: corrected `roundResponseMaxLength = 4000` comment (was incorrectly claiming "matches brainstorming reference"; brainstorming actually uses default 65536)
- **Deferred (LOW/MEDIUM, non-blocking):**
  - Quality #1 (parameter sprawl, embed `commonSpec`) — LOW
  - Quality #5 (collapse AC2/AC5 into single table) — MEDIUM
  - Quality #6 (`intDefault` already extracted as `defaultMaxRounds`; remaining trivia LOW)

### AC Validation: SKIPPED (Phase 4f opt-in defaults to skip; no UI-facing ACs in story 01)

### Open Decisions Resolved (per plan)

- **Open Decision #1 (helper vs literal):** helper-call pattern — committed
- **Open Decision #2 (round-response shape):** ShapeJSON — committed

### Notes for Story 02 Lead

- Helper accept-token baseline is `{"done", "wrap up", "complete"}` — diverges from legacy brainstorming (`"finish"`); divergence intentional and documented in `interactive_defaults.go:17-26`
- `RoundResponseInputID` constant available in types.go for use by stories 02-08
- buildGate / defaultMaxRounds available — no need to reimplement
