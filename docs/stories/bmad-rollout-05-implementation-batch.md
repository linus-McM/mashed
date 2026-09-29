# Story bmad-rollout-05: Phase 3c — Implementation Batch Iterative Upgrades

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** bmad-rollout-01, bmad-rollout-02
**Status:** done

## Description

Apply `applyIterativeUpgrade` to the implementation-domain processes from the plan's "Autonomous → upgrade to Iterative" Implementation rows that did NOT ship in the high-traffic story 02. After story 02's six (`dev-story`, `code-review`, `create-story`, `validate-prd`, `edit-prd`, `quick-dev`), the residual Implementation row is `bmad-qa-generate-e2e-tests`. The plan also marks `bmad-infrastructure-devops` as Implementation by phase enum but classifies it under "Support" in the rollout table — that ID ships in story 06.

This batch is small (one process), but is its own commit per plan §"Phasing" review-locality discipline.

## Developer Notes

### Architecture
- New file: `internal/bmad/registry_interactive_phase3c.go` with one `applyIterativeUpgrade` call.
- New test file: `internal/bmad/registry_interactive_phase3c_test.go`.
- Append the ID to the rollout skip slice.

### Per-process spec (from plan §"Scope inventory" Implementation row)

| ID | Prompt | Help text | DomainAccept | MaxRounds | OutputSpecs |
|---|---|---|---|---|---|
| bmad-qa-generate-e2e-tests | "Iterate on the test suite or 'done'/'complete' when satisfied." | "Add cases, fix flakes; 'done' to finalise." | `[]` (baseline only) | 15 | `tests` (memory; unmapped), `any-doc` (memory; unmapped) |

Notes:
- Both outputs are unmapped per `internal/bmad/artifacts.go:ResolveArtifactPath` returning `""` for `tests`/`any-doc`. Use `OutputToMemory` with `ArtifactName: ""`.
- Story 02 already covered `bmad-dev-story`, `bmad-code-review`, `bmad-quick-dev`; do NOT redo them here. Add a guard in `processIndex`-using code or simply omit those IDs from this file.

### Risks & migration notes

- **Risk #1 (Persisted workflows)**: Snapshot status precedence shields old executions.
- **Risk #4 (Tests pinning autonomous behaviour)**:
  ```bash
  rg -l 'bmad-qa-generate-e2e-tests' internal/bmad/*_test.go
  ```
  Likely zero hits — this process is not part of the daily-driver path.
- **Risk #4 (golden refresh)**: Add the ID to the rollout skip slice.
- **Risk #6 (Round limit)**: 15 rounds — middle ground for test iteration.
- **Risk #7 (Adapter timeout amplification)**: Cumulative now ~14 newly-interactive processes. Pre-cropping in commit `ada5b96` keeps individual translations under budget.
- **Open Decision #4 (skill prompts)**: Deferred to story 09.

### Reference Files
- `internal/bmad/registry.go:278-288` — `bmad-qa-generate-e2e-tests` (current).
- `internal/bmad/artifacts.go:artifactPaths` — confirm `tests`/`any-doc` are unmapped.
- `internal/bmad/registry_interactive_phase2.go` — pattern reference.

## Acceptance Criteria

**AC-1: bmad-qa-generate-e2e-tests upgraded to InteractIterative**
- Given the registry after init
- When `ProcessByID("bmad-qa-generate-e2e-tests")` is called
- Then `Mode == InteractIterative` and `EnableAstAdapter == true`

**AC-2: Round-response iteration slot present**
- Given the upgraded process
- When `iterationInput()` is called
- Then `(InputSpec{ID:"round-response", Shape:ShapeJSON, ...}, true)` is returned

**AC-3: Gate populated with 15-round cap and baseline accept tokens**
- Given the upgraded process
- Then `Gate.Kind == GateUserConfirm`
- And `Gate.MaxRounds == 15`
- And `Gate.AcceptTokens` equals exactly `["done", "wrap up", "complete"]`
- And `Gate.RejectTokens` equals `["abort", "cancel"]`

**AC-4: OutputSpecs reflect unmapped artifacts as memory-only**
- Given the upgraded process
- Then `OutputSpecs` contains entries for `tests` and `any-doc` each with `Target == OutputToMemory` and `ArtifactName == ""`

**AC-5: U0 skip list extended; byte-equality holds for the rest**
- Given `TestU0_AC3_NonMigratedProcessesUnchanged` runs
- Then the rollout skip slice now includes 14 IDs (6 phase-2 + 3 phase-3a + 4 phase-3b + 1 phase-3c)
- And the test passes

## BDD Test Scenarios

```gherkin
Feature: Phase 3c implementation-batch iterative upgrade

  Scenario: qa-generate-e2e-tests becomes iterative with 15-round cap
    When ProcessByID("bmad-qa-generate-e2e-tests") is called
    Then Mode is "iterative"
    And EnableAstAdapter is true
    And Gate.MaxRounds is 15
    And iterationInput() returns a spec with ID "round-response" and Shape ShapeJSON

  Scenario: qa-generate-e2e-tests outputs are memory-only
    When ProcessByID("bmad-qa-generate-e2e-tests") is called
    Then each OutputSpec has Target OutputToMemory and ArtifactName ""

  Scenario: U0 skip list contains the rollout IDs to date
    When TestU0_AC3_NonMigratedProcessesUnchanged runs
    Then the skip predicate matches 14 rollout IDs

  Scenario: Resume of a pre-rollout qa-generate-e2e-tests node still works
    Given a workflows snapshot with a "bmad-qa-generate-e2e-tests" node in Status "complete"
    When the executor resumes
    Then the node remains "complete" with no awaiting_input event
```

## Tasks / Subtasks

- [x] Task 1: Create `internal/bmad/registry_interactive_phase3c.go` (AC-1..4)
  - [x] One `applyIterativeUpgrade` call for `bmad-qa-generate-e2e-tests` with baseline-only DomainAccept and MaxRounds: 15.
  - [x] OutputSpecs: two `OutputToMemory` entries (`tests`, `any-doc`).
- [x] Task 2: Extend rollout skip slice (AC-5)
  - [x] Append `bmad-qa-generate-e2e-tests`.
- [x] Task 3: Walking integration test `registry_interactive_phase3c_test.go` (AC-1..4)
  - [x] Single sub-test verifying Mode/Gate/iterationInput/OutputSpecs.
- [x] Task 4: Audit prior tests
  - [x] `rg -l 'bmad-qa-generate-e2e-tests' internal/bmad/*_test.go`.
- [x] Task 5: Smoke verification (manual)
  - [x] `wails dev`, drop a `bmad-qa-generate-e2e-tests` node, type "done", confirm gate satisfaction.

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ coverage on `registry_interactive_phase3c.go`
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `/simplify` run on all modified code; no CRITICAL/HIGH issues
- [x] AC validation table populated in PR description
- [x] Smoke run on the qa-generate-e2e-tests process
- [x] Status flipped to `done` by sprint lead
