# Story bmad-rollout-06: Phase 3d — Support Batch Iterative Upgrades

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** bmad-rollout-01, bmad-rollout-02
**Status:** ready

## Description

Apply `applyIterativeUpgrade` to the six support-domain autonomous processes the rollout plan classifies as iterative: `bmad-editorial-review-prose`, `bmad-editorial-review-structure`, `bmad-review-edge-case-hunter`, `bmad-quick-flow`, `bmad-adversarial-general`, `bmad-infrastructure-devops`. This closes the Iterative phase of the rollout (20 total Iterative upgrades = 6 phase-2 + 3 phase-3a + 4 phase-3b + 1 phase-3c + 6 phase-3d).

## Developer Notes

### Architecture
- New file: `internal/bmad/registry_interactive_phase3d.go` with six `applyIterativeUpgrade` calls.
- New test file: `internal/bmad/registry_interactive_phase3d_test.go`.
- Append the six IDs to the rollout skip slice.

### Per-process specs (from plan §"Scope inventory" Support rows)

| ID | Prompt | Help text | DomainAccept | MaxRounds | OutputSpecs |
|---|---|---|---|---|---|
| bmad-editorial-review-prose | "Add prose feedback or 'approved'/'done' when satisfied." | "Iterate on tone/clarity; 'approved' to ship." | `["approved"]` | 15 | `reviewed-doc` (file) |
| bmad-editorial-review-structure | "Add structural feedback or 'approved'/'done' when satisfied." | "Iterate on organisation/sections; 'approved' to ship." | `["approved"]` | 15 | `reviewed-doc` (file) |
| bmad-review-edge-case-hunter | "Add edge cases or 'done'/'approved' when the report is complete." | "Iterate on contradictions/missing reqs; 'done' to close." | `["approved"]` | 15 | `edge-case-report` (file) |
| bmad-quick-flow | "Iterate on the flow or 'done'/'ship' when ready." | "Lightweight loop; 'ship' to finalise." | `["ship"]` | 10 | `code` (memory; unmapped), `PRD.md` (file) |
| bmad-adversarial-general | "Add adversarial findings or 'done'/'complete' when the review is closed." | "Iterate on weaknesses/blind spots; 'done' to close." | `[]` (baseline only) | 15 | `adversarial-report` (file) |
| bmad-infrastructure-devops | "Iterate on infra/CI or 'done'/'ship' when ready." | "Discuss pipeline/deploy; 'ship' to finalise." | `["ship"]` | 15 | `infra-config` (file) |

Notes:
- `reviewed-doc`, `edge-case-report`, `PRD.md`, `adversarial-report`, `infra-config` are mapped artifacts per `internal/bmad/artifacts.go:artifactPaths`.
- `code` is unmapped → `OutputToMemory, ArtifactName: ""`.
- Confirm artifact mappings before locking the OutputSpec target column.

### Risks & migration notes

- **Risk #1 (Persisted workflows)**: Snapshot status precedence shields old executions.
- **Risk #4 (Tests pinning autonomous behaviour)**:
  ```bash
  rg -l 'bmad-(editorial-review-prose|editorial-review-structure|review-edge-case-hunter|quick-flow|adversarial-general|infrastructure-devops)' internal/bmad/*_test.go
  ```
  Most are likely opaque ID references in graph fixtures. Audit each.
- **Risk #4 (golden refresh)**: Add the six IDs to the rollout skip slice — 20 total entries after this story (closing the Iterative phase).
- **Risk #5 (Frontend modal copy)**: This is the broadest single batch. Plan §"Migration risks" #5 calls for a Playwright smoke run hitting one node from each batch — apply that to at least one Support process here (e.g. `bmad-editorial-review-prose`).
- **Risk #6 (Round limit)**: 10–15 rounds per process per the cadence hint.
- **Risk #7 (Adapter timeout amplification)**: Cumulative now 20 newly-interactive processes. Pre-cropping (commit `ada5b96`) handles per-translation budget.
- **Open Decision #4 (skill prompts)**: Deferred to story 09.

### Reference Files
- `internal/bmad/registry.go:303-313` — `bmad-editorial-review-prose`.
- `internal/bmad/registry.go:316-325` — `bmad-editorial-review-structure`.
- `internal/bmad/registry.go:355-365` — `bmad-review-edge-case-hunter`.
- `internal/bmad/registry.go:392-402` — `bmad-quick-flow`.
- `internal/bmad/registry.go:404-414` — `bmad-adversarial-general`.
- `internal/bmad/registry.go:415-426` — `bmad-infrastructure-devops`.
- `internal/bmad/registry_interactive_phase2.go` — pattern reference.

## Acceptance Criteria

**AC-1: Six support processes upgraded to InteractIterative**
- Given the registry after init
- When `ProcessByID(id)` is called for each of the six IDs
- Then `Mode == InteractIterative` and `EnableAstAdapter == true`

**AC-2: Each exposes a ShapeJSON round-response iteration slot**
- Given each upgraded process
- When `iterationInput()` is called
- Then it returns `(spec, true)` with `ID == "round-response"` and `Shape == ShapeJSON`

**AC-3: Per-process Gate cap and accept-token sets match the table**
- Given each upgraded process
- Then `Gate.Kind == GateUserConfirm`
- And `Gate.MaxRounds` matches the table (15/15/15/10/15/15)
- And `Gate.AcceptTokens` contains the baseline + per-process domain tokens
- And `Gate.RejectTokens` equals `["abort", "cancel"]`

**AC-4: OutputSpecs reflect mapped vs unmapped artifacts**
- Given each upgraded process
- Then mapped artifacts (`reviewed-doc`, `edge-case-report`, `PRD.md`, `adversarial-report`, `infra-config`) carry `Target == OutputToFile`
- And `code` (unmapped) carries `Target == OutputToMemory` with `ArtifactName == ""`

**AC-5: U0 skip list extended to 20 entries; non-rolled-out byte-equality intact**
- Given `TestU0_AC3_NonMigratedProcessesUnchanged` runs
- Then the rollout skip slice contains 20 IDs (closing the Iterative phase)
- And the test passes

**AC-6: Playwright smoke for at least one Support process**
- Given `wails dev` running with the editorial-review-prose process
- When a workflow node is dropped, executed, and "approved" is typed in the modal
- Then the gate satisfies and the node transitions to complete

## BDD Test Scenarios

```gherkin
Feature: Phase 3d support-batch iterative upgrades

  Scenario: editorial-review-prose accepts "approved" and writes reviewed-doc
    When ProcessByID("bmad-editorial-review-prose") is called
    Then Mode is "iterative"
    And Gate.AcceptTokens contains "approved"
    And OutputSpecs has one entry with Target OutputToFile and ArtifactName "reviewed-doc"

  Scenario: editorial-review-structure parallels prose with same shape
    When ProcessByID("bmad-editorial-review-structure") is called
    Then Mode is "iterative"
    And Gate.AcceptTokens contains "approved"
    And Gate.MaxRounds is 15

  Scenario: review-edge-case-hunter writes edge-case-report
    When ProcessByID("bmad-review-edge-case-hunter") is called
    Then OutputSpecs has one entry with ArtifactName "edge-case-report" and Target OutputToFile

  Scenario: quick-flow caps at 10 rounds with "ship" token and PRD.md output
    When ProcessByID("bmad-quick-flow") is called
    Then Gate.MaxRounds is 10
    And Gate.AcceptTokens contains "ship"
    And OutputSpecs has one entry with ArtifactName "PRD.md" and Target OutputToFile

  Scenario: adversarial-general carries baseline tokens only
    When ProcessByID("bmad-adversarial-general") is called
    Then Gate.AcceptTokens equals exactly ["done", "wrap up", "complete"]

  Scenario: infrastructure-devops accepts "ship" token
    When ProcessByID("bmad-infrastructure-devops") is called
    Then Gate.AcceptTokens contains "ship"
    And OutputSpecs has one entry with ArtifactName "infra-config"

  Scenario: U0 skip list contains all 20 rollout IDs
    When TestU0_AC3_NonMigratedProcessesUnchanged runs
    Then the skip predicate matches 20 rollout IDs (6 + 3 + 4 + 1 + 6)

  Scenario: Frontend modal smoke for editorial-review-prose
    Given the Wails dev app running
    And a "bmad-editorial-review-prose" node on the canvas
    When the node executes and reaches the first idle
    Then the InputResponseModal renders the round-response widget
    When the user types "approved"
    Then the gate satisfies and the node turns complete
```

## Tasks / Subtasks

- [ ] Task 1: Create `internal/bmad/registry_interactive_phase3d.go` (AC-1..4)
  - [ ] Six `applyIterativeUpgrade` calls via `processIndex(id)` matching the table.
  - [ ] OutputSpecs assembled per the mapped/unmapped column.
- [ ] Task 2: Extend rollout skip slice (AC-5)
  - [ ] Append the six IDs; total now 20.
- [ ] Task 3: Walking integration test `registry_interactive_phase3d_test.go` (AC-1..4)
  - [ ] Sub-test per ID covering Mode/Gate/iterationInput/OutputSpecs.
- [ ] Task 4: Audit dependent tests
  - [ ] `rg` for the six IDs in `*_test.go`; update or refactor.
- [ ] Task 5: Playwright smoke (AC-6)
  - [ ] Add or extend a spec under `tests/ac/` that drives `bmad-editorial-review-prose` through a one-round-then-"approved" flow with the InputResponseModal.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `registry_interactive_phase3d.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code; no CRITICAL/HIGH issues
- [ ] AC validation table populated in PR description
- [ ] Playwright smoke spec added for at least one Support process
- [ ] Status flipped to `done` by sprint lead
