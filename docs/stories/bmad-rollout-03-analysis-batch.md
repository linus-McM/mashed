# Story bmad-rollout-03: Phase 3a — Analysis Batch Iterative Upgrades

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** bmad-rollout-01, bmad-rollout-02
**Status:** done

## Description

Apply `applyIterativeUpgrade` to the three analysis-domain autonomous processes that the rollout plan classifies as iterative: `bmad-domain-research`, `bmad-market-research`, `bmad-technical-research`. (`bmad-generate-project-context` is handled in story 07 as a Guided upgrade per the plan's "Autonomous → upgrade to Guided" table.) Refresh U0 byte-equality coverage and add walking assertions identical in shape to story 02 so the helper-driven idiom stays uniform across phases.

## Developer Notes

### Architecture
- New file: `internal/bmad/registry_interactive_phase3a.go` running an `init()` that calls `applyIterativeUpgrade` three times via the same `processIndex(id)` helper introduced in story 02. Lexical filename ordering (`phase3a` follows `phase2`) keeps init ordering deterministic.
- New test file: `internal/bmad/registry_interactive_phase3a_test.go`.
- Extend the `phase2RolloutIDs` slice introduced in story 02's test refactor to a generalised `interactiveRolloutIDs` slice OR add a sibling `phase3aRolloutIDs` slice and union both in the U0 skip predicate. Recommended: rename the slice to `interactiveRolloutIDs` once and append per-phase from a single source of truth.

### Per-process specs (from plan §"Scope inventory" Analysis rows)

| ID | Prompt | Help text | DomainAccept | MaxRounds | OutputSpecs |
|---|---|---|---|---|---|
| bmad-domain-research | "Add findings or 'done'/'complete' when the domain notes are ready." | "Iterate on terminology, constraints; 'done' to close." | `[]` (baseline only) | 30 | `domain-research` (file) |
| bmad-market-research | "Add competitive insights or 'done'/'complete' when the market notes are ready." | "Iterate on competitors, positioning; 'done' to close." | `[]` (baseline only) | 30 | `market-research` (file) |
| bmad-technical-research | "Add technical findings or 'done'/'complete' when the tech notes are ready." | "Iterate on libraries, trade-offs; 'done' to close." | `[]` (baseline only) | 30 | `tech-research` (file) |

Notes:
- Plan §"Scope inventory" lists `done`, `complete` as suggested accept tokens for each. Both are in baseline `commonAcceptTokens` (`done`, `wrap up`, `complete`), so `DomainAccept` is empty.
- Per Risk #6, `MaxRounds = 30` is appropriate for research (long-form exploration is the natural cadence).
- `domain-research`, `market-research`, `tech-research` are all mapped artifacts — confirm in `internal/bmad/artifacts.go:artifactPaths` before submitting; if any is unmapped, fall back to `Target: OutputToMemory, ArtifactName: ""`.

### Risks & migration notes

- **Risk #1 (Persisted workflows)**: Same as story 02 — snapshot status precedence in `internal/bmad/resume.go:347-353` shields old executions.
- **Risk #4 (Tests pinning autonomous behaviour)**: Inventory pass:
  ```bash
  rg -l 'bmad-(domain|market|technical)-research' internal/bmad/*_test.go
  ```
  Lower-traffic IDs than story 02's batch — most likely zero hits. Audit and update if any test asserts `Mode == ""` on these IDs.
- **Risk #4 (golden refresh)**: Add the three IDs to the rollout skip slice in `registry_test.go`.
- **Risk #6 (Round limit)**: research processes intentionally keep the 30-round default. Plan §"Migration risks" #6 calls this out as the natural cadence.
- **Risk #7 (Adapter timeout amplification)**: Three more processes calling `adapter.Translate` per round adds Haiku traffic. Existing pre-cropping (commit `ada5b96`) keeps individual translations under budget; rate limits are the user's claude.ai quota.
- **Open Decision #4 (skill prompt alignment)**: Deferred to story 09. Phase 3a ships with existing skill copy.

### Reference Files
- `internal/bmad/registry.go:67-102` — current entries for the three research processes (autonomous, no Mode).
- `internal/bmad/artifacts.go:artifactPaths` — confirms which artifact names are mapped.
- `internal/bmad/registry_interactive_phase2.go` — story 02's helper-call template.
- `internal/bmad/registry_interactive_phase2_test.go` — walking-assertion template to mirror.

## Acceptance Criteria

**AC-1: Three analysis processes upgraded to InteractIterative**
- Given the registry after init
- When `ProcessByID("bmad-domain-research")`, `ProcessByID("bmad-market-research")`, `ProcessByID("bmad-technical-research")` are called
- Then each returns `Mode == InteractIterative` and `EnableAstAdapter == true`

**AC-2: Each declares a round-response iteration slot**
- Given each upgraded process
- When `iterationInput()` is called
- Then `(InputSpec{ID:"round-response", Shape:ShapeJSON, ...}, true)` is returned

**AC-3: Each declares a populated Gate with MaxRounds == 30 and baseline accept tokens**
- Given each upgraded process
- Then `Gate.Kind == GateUserConfirm`
- And `Gate.MaxRounds == 30`
- And `Gate.AcceptTokens` equals exactly `["done", "wrap up", "complete"]`
- And `Gate.RejectTokens` equals `["abort", "cancel"]`

**AC-4: OutputSpecs reflect file-mapped artifacts**
- Given each upgraded process
- When `OutputSpecs` is inspected
- Then a single entry with `Target == OutputToFile` and `ArtifactName` matching the artifact name (`domain-research`, `market-research`, or `tech-research`) is present

**AC-5: U0 byte-equality skip list contains all rollout IDs to date**
- Given `TestU0_AC3_NonMigratedProcessesUnchanged` runs
- When the rollout skip slice is consulted
- Then it includes the six phase-2 IDs and the three phase-3a IDs (nine total)
- And the test passes for the remaining ~17 unchanged processes

## BDD Test Scenarios

```gherkin
Feature: Phase 3a analysis-batch iterative upgrades

  Scenario: domain-research becomes iterative with 30-round cap
    When ProcessByID("bmad-domain-research") is called
    Then Mode is "iterative"
    And EnableAstAdapter is true
    And Gate.MaxRounds is 30
    And iterationInput() returns a spec with ID "round-response" and Shape ShapeJSON

  Scenario: market-research outputs file-mapped market-research artifact
    When ProcessByID("bmad-market-research") is called
    Then OutputSpecs has one entry with Target OutputToFile and ArtifactName "market-research"

  Scenario: technical-research carries baseline accept tokens only
    When ProcessByID("bmad-technical-research") is called
    Then Gate.AcceptTokens equals exactly ["done", "wrap up", "complete"]

  Scenario: U0 skip list extended to include phase 3a IDs
    When TestU0_AC3_NonMigratedProcessesUnchanged runs
    Then the skip predicate matches 9 rollout IDs (6 phase-2 + 3 phase-3a)
    And the test passes for the rest

  Scenario: Resume of a pre-rollout running domain-research node still works
    Given a workflows snapshot with a "bmad-domain-research" node in Status "complete"
    When the executor resumes
    Then the node remains "complete" with no awaiting_input event
```

## Tasks / Subtasks

- [x] Task 1: Create `internal/bmad/registry_interactive_phase3a.go` (AC-1, AC-2, AC-3, AC-4)
  - [x] `init()` calls `applyIterativeUpgrade(&registry[processIndex(id)], spec)` three times.
  - [x] Specs use baseline-only DomainAccept and MaxRounds: 30.
  - [x] OutputSpecs reference `OutputToFile` with the mapped artifact name.
- [x] Task 2: Update U0 skip slice (AC-5)
  - [x] Add the three IDs to the rollout skip list in `registry_test.go` (rename to `interactiveRolloutIDs` if appropriate, or extend `phase3aRolloutIDs` sibling slice).
- [x] Task 3: Walking integration test `registry_interactive_phase3a_test.go` (AC-1..4)
  - [x] Sub-test per ID asserting Mode/Gate/iterationInput/OutputSpecs.
- [x] Task 4: Audit prior tests for autonomous-shape pins (AC-6 implicit)
  - [x] `rg -l 'bmad-(domain|market|technical)-research' internal/bmad/*_test.go` and update any that read `Mode == ""`.
- [x] Task 5: Smoke verification (manual)
  - [x] `wails dev`, drop a `bmad-domain-research` node, confirm UI-AST modal opens on first idle and "done" closes the gate.

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ coverage on `registry_interactive_phase3a.go`
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `/simplify` run on all modified code; no CRITICAL/HIGH issues
- [x] AC validation table populated in PR description
- [x] Smoke run on at least one of the three research processes
- [x] Status flipped to `done` by sprint lead
