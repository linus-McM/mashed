# Story bmad-rollout-07: Phase 4 — Guided Upgrades + applyGuidedUpgrade Wiring

**Priority:** P2-medium
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** bmad-rollout-01
**Status:** done

## Description

Apply `applyGuidedUpgrade` (helper from story 01) to the three autonomous processes that the plan classifies as Guided single-pass Q&A: `bmad-create-prd`, `bmad-document-project`, `bmad-generate-project-context`. Each receives a sequence of staged `InputSpec`s rendered in declaration order (mirroring the `bmad-product-brief` reference shape from §10.2), no recurring iteration slot, and a small `Gate` with the user-confirm exit token (`yes`).

This story can run in parallel with the Iterative-batch stories (02–06) because Guided upgrades use a different helper and touch different processes, but it cannot start before story 01 (helper foundation).

## Developer Notes

### Architecture
- New file: `internal/bmad/registry_interactive_phase4.go` running `init()` after registry init.
- Each call site uses `applyGuidedUpgrade(&registry[processIndex(id)], spec)`.
- Append the three IDs to the rollout skip slice in `registry_test.go`.
- New test file: `internal/bmad/registry_interactive_phase4_test.go`.

### Per-process specs (from plan §"Autonomous → upgrade to Guided")

**`bmad-create-prd`** — three staged questions: scope, audience, timeline.

```go
StagedInputs: []InputSpec{
    {ID: "scope", Source: InputFromUser, Shape: ShapeFree, Required: true,
     Prompt: "What is the scope of this PRD?", HelpText: "Describe the surface area in 1-3 sentences.", MaxLength: 1000},
    {ID: "audience", Source: InputFromUser, Shape: ShapeFree, Required: true,
     Prompt: "Who is the target audience?", HelpText: "Personas, teams, or end-users.", MaxLength: 500},
    {ID: "timeline", Source: InputFromUser, Shape: ShapeChoice, Required: true,
     Prompt: "What is the rough timeline?", Options: []string{"days", "weeks", "months", "quarters"}, Default: "weeks"},
},
FinalApproval: &InputSpec{ID: "approve", Source: InputFromUser, Shape: ShapeApproval, Required: true,
                          Prompt: "Approve this PRD draft?"},
OutputSpecs: []OutputSpec{
    {ID: "PRD.md", Target: OutputToFile, ArtifactName: "PRD.md"},
},
```

**`bmad-document-project`** — scope picker, depth picker, output location.

```go
StagedInputs: []InputSpec{
    {ID: "scope", Source: InputFromUser, Shape: ShapeChoice, Required: true,
     Prompt: "What scope should the docs cover?", Options: []string{"public-api", "all", "selected-modules"}, Default: "public-api"},
    {ID: "depth", Source: InputFromUser, Shape: ShapeChoice, Required: true,
     Prompt: "How deep?", Options: []string{"summary", "reference", "tutorial"}, Default: "reference"},
    {ID: "output-location", Source: InputFromUser, Shape: ShapeFree, Required: false,
     Prompt: "Where should output land? (default project-docs)", HelpText: "Repo-relative path; leave blank for default."},
},
FinalApproval: &InputSpec{ID: "approve", Source: InputFromUser, Shape: ShapeApproval, Required: true,
                          Prompt: "Approve this documentation plan?"},
OutputSpecs: []OutputSpec{
    {ID: "project-docs", Target: OutputToFile, ArtifactName: "project-docs"},
},
```

**`bmad-generate-project-context`** — scope picker, sources picker.

```go
StagedInputs: []InputSpec{
    {ID: "scope", Source: InputFromUser, Shape: ShapeChoice, Required: true,
     Prompt: "What context level do you need?", Options: []string{"onboarding", "deep-dive", "snapshot"}, Default: "onboarding"},
    {ID: "sources", Source: InputFromUser, Shape: ShapeMultiChoice, Required: true,
     Prompt: "Which sources should we include?", Options: []string{"PRD", "architecture", "epics", "code", "history"}},
},
FinalApproval: &InputSpec{ID: "approve", Source: InputFromUser, Shape: ShapeApproval, Required: true,
                          Prompt: "Approve generated context?"},
OutputSpecs: []OutputSpec{
    {ID: "project-context.md", Target: OutputToFile, ArtifactName: "project-context.md"},
},
```

Notes:
- Helper sets `Mode = InteractGuided`, `EnableAstAdapter = true`, `Gate.AcceptTokens = ["yes"]`, `Gate.MaxRounds = 10` (helper defaults).
- Story 01's `GuidedUpgradeSpec.AcceptTokens` defaults to `{"yes"}` when nil — matches reference §10.2.
- `iterationInput()` must return `(InputSpec{}, false)` for Guided processes — guard via the helper not appending a recurring slot.
- Confirm `PRD.md`, `project-docs`, `project-context.md` mappings in `internal/bmad/artifacts.go:artifactPaths` before locking OutputSpec target.

### Risks & migration notes

- **Risk #1 (Persisted workflows)**: Snapshot status precedence shields old executions for pre-upgrade `bmad-create-prd` runs.
- **Risk #2 (Templates string-only)**: Templates reference IDs and labels — Mode changes don't break references (verified by plan).
- **Risk #4 (Tests pinning autonomous behaviour)**:
  ```bash
  rg -l 'bmad-(create-prd|document-project|generate-project-context)' internal/bmad/*_test.go
  ```
  Audit hits.
- **Risk #4 (golden refresh)**: Add the three IDs to rollout skip slice.
- **Risk #5 (Frontend modal copy)**: Plan calls for a Playwright smoke per batch. Drive `bmad-create-prd` through three stages + approval to catch awkward copy.
- **Risk #7 (Adapter timeout amplification)**: Guided processes call `adapter.Translate` once per stage prompt, not per round — lower amplification than Iterative.
- **Open Decision #4 (skill prompts)**: Deferred to story 09. The skill markdown for these three processes may be updated separately to align "ask the user" copy with the staged prompt vocabulary.

### Reference Files
- `internal/bmad/registry.go:42-66` — `bmad-product-brief` reference Guided shape (§10.2).
- `internal/bmad/registry.go:106-116` — current `bmad-create-prd` (autonomous).
- `internal/bmad/registry.go:428-438` — current `bmad-document-project`.
- `internal/bmad/registry.go:192-202` — current `bmad-generate-project-context`.
- `internal/bmad/artifacts.go:artifactPaths` — confirm artifact mappings.
- `internal/bmad/interactive_defaults.go` (story 01) — `applyGuidedUpgrade` and `GuidedUpgradeSpec`.

## Acceptance Criteria

**AC-1: Three processes upgraded to InteractGuided**
- Given the registry after init
- When `ProcessByID(id)` is called for each of the three IDs
- Then `Mode == InteractGuided` and `EnableAstAdapter == true`

**AC-2: StagedInputs render in declared order**
- Given each upgraded process
- When `InputSpecs` is inspected
- Then the order matches the per-process spec table (scope/audience/timeline for create-prd; scope/depth/output-location for document-project; scope/sources for generate-project-context)
- And the `FinalApproval` spec is the last entry in `InputSpecs`

**AC-3: No iteration slot present**
- Given each upgraded process
- When `iterationInput()` is called
- Then it returns `(InputSpec{}, false)`

**AC-4: Gate carries `["yes"]` accept tokens and 10-round cap**
- Given each upgraded process
- Then `Gate.Kind == GateUserConfirm`
- And `Gate.AcceptTokens` equals `["yes"]`
- And `Gate.MaxRounds == 10`

**AC-5: OutputSpecs reflect file-mapped artifacts**
- Given each upgraded process
- Then a single `OutputSpec` with `Target == OutputToFile` and the matching `ArtifactName` is present (`PRD.md`, `project-docs`, `project-context.md`)

**AC-6: U0 skip list extended; non-rolled-out byte-equality intact**
- Given `TestU0_AC3_NonMigratedProcessesUnchanged` runs
- Then the rollout skip slice contains the three Guided IDs in addition to all Iterative entries (23 total post-Phase-4)
- And the test passes

**AC-7: Playwright smoke for create-prd**
- Given `wails dev` running with `bmad-create-prd`
- When a node is dropped, all three stages plus approval are answered
- Then the gate satisfies and the node completes

## BDD Test Scenarios

```gherkin
Feature: Phase 4 guided upgrades

  Scenario: create-prd is guided with three staged inputs plus approval
    When ProcessByID("bmad-create-prd") is called
    Then Mode is "guided"
    And EnableAstAdapter is true
    And InputSpecs has IDs ["scope", "audience", "timeline", "approve"] in that order
    And iterationInput() returns false
    And Gate.AcceptTokens equals ["yes"]
    And Gate.MaxRounds is 10

  Scenario: document-project carries scope/depth/output-location
    When ProcessByID("bmad-document-project") is called
    Then Mode is "guided"
    And InputSpecs has IDs ["scope", "depth", "output-location", "approve"] in that order

  Scenario: generate-project-context uses MultiChoice for sources
    When ProcessByID("bmad-generate-project-context") is called
    Then InputSpecs has an entry with ID "sources" and Shape ShapeMultiChoice

  Scenario: All three Guided processes write file outputs
    When ProcessByID is called for each Guided ID
    Then OutputSpecs contains exactly one OutputToFile entry per process

  Scenario: U0 skip list grows by three Guided IDs
    When TestU0_AC3_NonMigratedProcessesUnchanged runs
    Then the skip predicate matches 23 rollout IDs (20 iterative + 3 guided)

  Scenario: Resume of a pre-upgrade create-prd snapshot
    Given a workflows snapshot with a "bmad-create-prd" node in Status "complete"
    When the executor resumes
    Then the node remains "complete" with no awaiting_input event

  Scenario: Frontend smoke — create-prd guided flow
    Given the Wails dev app running
    And a "bmad-create-prd" node on the canvas
    When the node executes and the modal opens
    Then the user can answer scope, audience, timeline in sequence
    And the FinalApproval modal renders an approval widget
    When the user clicks "Yes"
    Then the gate satisfies and the node turns complete
```

## Tasks / Subtasks

- [x] Task 1: Create `internal/bmad/registry_interactive_phase4.go` (AC-1..5)
  - [x] Three `applyGuidedUpgrade` calls via `processIndex(id)` matching the per-process specs above.
  - [x] Confirm `bmad-create-prd` carries `[scope, audience, timeline, approve]` post-helper.
  - [x] Confirm Guided processes do NOT register an iteration slot (helper guarantee).
- [x] Task 2: Extend rollout skip slice (AC-6)
  - [x] Append the three Guided IDs.
- [x] Task 3: Walking integration test `registry_interactive_phase4_test.go` (AC-1..5)
  - [x] Sub-test per ID covering Mode/InputSpecs order/iterationInput()=false/Gate.
- [x] Task 4: Audit prior tests
  - [x] `rg -l 'bmad-(create-prd|document-project|generate-project-context)' internal/bmad/*_test.go`.
- [x] Task 5: Playwright smoke for create-prd (AC-7)
  - [x] Add a spec under `tests/ac/` driving the three-stage Guided flow.

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ coverage on `registry_interactive_phase4.go`
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `/simplify` run on all modified code; no CRITICAL/HIGH issues
- [x] AC validation table populated in PR description
- [x] Playwright smoke spec added for `bmad-create-prd`
- [x] Status flipped to `done` by sprint lead
