# Story bmad-rollout-04: Phase 3b — Planning Batch Iterative Upgrades

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** bmad-rollout-01, bmad-rollout-02
**Status:** ready

## Description

Apply `applyIterativeUpgrade` to four planning-domain autonomous processes that the rollout plan classifies as iterative: `bmad-create-ux-design`, `bmad-create-architecture`, `bmad-check-implementation-readiness`, `bmad-create-epics-and-stories`. Note: `bmad-edit-prd` and `bmad-validate-prd` (planning, but high-traffic) shipped in story 02; `bmad-create-prd` is Guided and ships in story 07.

## Developer Notes

### Architecture
- New file: `internal/bmad/registry_interactive_phase3b.go` with an `init()` calling `applyIterativeUpgrade` four times via `processIndex(id)`.
- New test file: `internal/bmad/registry_interactive_phase3b_test.go`.
- Append the four IDs to the rollout skip slice in `registry_test.go` (introduced in stories 02–03).

### Per-process specs (from plan §"Scope inventory" Planning rows)

| ID | Prompt | Help text | DomainAccept | MaxRounds | OutputSpecs |
|---|---|---|---|---|---|
| bmad-create-ux-design | "Refine the UX or 'done'/'approved'/'ship' to finalise." | "Iterate on flows/wireframes; 'approved' to ship the spec." | `["approved", "ship"]` | 15 | `ux-spec.md` (file), `PRD.md` (memory) |
| bmad-create-architecture | "Iterate on architecture or 'done'/'approved'/'ship' to finalise." | "Discuss components, trade-offs; 'approved' or 'ship' to seal." | `["approved", "ship"]` | 15 | `architecture.md` (file) |
| bmad-check-implementation-readiness | "Provide readiness feedback or 'ready'/'done' when complete." | "Confirm artifacts coherent; 'ready' to proceed." | `["ready"]` | 10 | `readiness-report` (file), `architecture.md` (memory), `any-doc` (memory; unmapped) |
| bmad-create-epics-and-stories | "Refine epics/stories or 'done'/'complete' when ready." | "Iterate on epic split, story carving; 'done' to finalise." | `[]` (baseline only) | 20 | `epics/` (memory; unmapped) |

Notes:
- `ux-spec.md`, `architecture.md`, `readiness-report` map to artifact paths via `internal/bmad/artifacts.go:artifactPaths`. `epics/`, `any-doc` are unmapped per `ResolveArtifactPath` returning `""` — use `OutputToMemory`.
- `MaxRounds: 10` for readiness aligns with Risk #6 (short, focused review).
- `MaxRounds: 20` for epics/stories balances larger surface vs cap (Risk #6).

### Risks & migration notes

- **Risk #1 (Persisted workflows)**: Snapshot status precedence in `internal/bmad/resume.go:347-353` shields old executions.
- **Risk #4 (Tests pinning autonomous behaviour)**:
  ```bash
  rg -l 'bmad-(create-ux-design|create-architecture|check-implementation-readiness|create-epics-and-stories)' internal/bmad/*_test.go
  ```
  Most matches will be opaque ID references in graph fixtures. Audit each for `Mode`/`Gate` assertions.
- **Risk #4 (golden refresh)**: Add the four IDs to the rollout skip slice.
- **Risk #6 (Round limit)**: per-process caps assigned per the plan's hint table.
- **Open Decision #4 (skill prompts)**: Deferred to story 09.

### Reference Files
- `internal/bmad/registry.go:142-152` — `bmad-create-ux-design` (current).
- `internal/bmad/registry.go:155-166` — `bmad-create-architecture`.
- `internal/bmad/registry.go:167-178` — `bmad-check-implementation-readiness`.
- `internal/bmad/registry.go:179-190` — `bmad-create-epics-and-stories`.
- `internal/bmad/artifacts.go:artifactPaths` — confirm artifact mappings.
- `internal/bmad/registry_interactive_phase2.go` — pattern reference.

## Acceptance Criteria

**AC-1: Four planning processes upgraded to InteractIterative**
- Given the registry after init
- When `ProcessByID(id)` is called for each of the four IDs
- Then `Mode == InteractIterative` and `EnableAstAdapter == true`

**AC-2: Each exposes a ShapeJSON round-response iteration slot**
- Given each upgraded process
- When `iterationInput()` is called
- Then it returns `(spec, true)` with `ID == "round-response"` and `Shape == ShapeJSON`

**AC-3: Per-process Gate cap and accept-token sets match the table**
- Given each upgraded process
- Then `Gate.Kind == GateUserConfirm`
- And `Gate.MaxRounds` matches the table (15/15/10/20)
- And `Gate.AcceptTokens` contains the baseline `"done", "wrap up", "complete"` plus the per-process domain tokens
- And `Gate.RejectTokens` equals `["abort", "cancel"]`

**AC-4: OutputSpecs reflect mapped vs unmapped artifacts**
- Given each upgraded process
- Then `ux-spec.md`, `architecture.md`, `readiness-report` carry `Target == OutputToFile` with matching `ArtifactName`
- And `epics/`, `any-doc`, `PRD.md` (when listed as memory) carry `Target == OutputToMemory` with empty `ArtifactName`

**AC-5: U0 skip list extended; non-rolled-out byte-equality intact**
- Given `TestU0_AC3_NonMigratedProcessesUnchanged` runs
- Then the rollout skip slice includes the four phase-3b IDs (in addition to phase-2 and phase-3a entries)
- And the test passes

## BDD Test Scenarios

```gherkin
Feature: Phase 3b planning-batch iterative upgrades

  Scenario: create-ux-design caps at 15 rounds with approved/ship tokens
    When ProcessByID("bmad-create-ux-design") is called
    Then Mode is "iterative"
    And Gate.MaxRounds is 15
    And Gate.AcceptTokens contains "approved" and "ship"
    And OutputSpecs has one OutputToFile entry with ArtifactName "ux-spec.md"

  Scenario: create-architecture writes architecture.md as file output
    When ProcessByID("bmad-create-architecture") is called
    Then OutputSpecs has one entry with Target OutputToFile and ArtifactName "architecture.md"

  Scenario: check-implementation-readiness uses "ready" accept token and 10-round cap
    When ProcessByID("bmad-check-implementation-readiness") is called
    Then Gate.AcceptTokens contains "ready"
    And Gate.MaxRounds is 10

  Scenario: create-epics-and-stories caps at 20 rounds with baseline tokens only
    When ProcessByID("bmad-create-epics-and-stories") is called
    Then Gate.MaxRounds is 20
    And Gate.AcceptTokens equals exactly ["done", "wrap up", "complete"]

  Scenario: U0 skip list contains all phase-2 + phase-3a + phase-3b IDs
    When TestU0_AC3_NonMigratedProcessesUnchanged runs
    Then the skip predicate matches 13 rollout IDs (6 + 3 + 4)

  Scenario: round-response slot is ShapeJSON across the batch
    When iterationInput() is called for each of the four IDs
    Then each returned spec has Shape ShapeJSON
```

## Tasks / Subtasks

- [ ] Task 1: Create `internal/bmad/registry_interactive_phase3b.go` (AC-1..4)
  - [ ] Four `applyIterativeUpgrade` calls via `processIndex(id)`.
  - [ ] OutputSpecs assembled per the table (mapped → file, unmapped → memory).
- [ ] Task 2: Extend rollout skip slice (AC-5)
  - [ ] Append the four IDs.
- [ ] Task 3: Walking integration test `registry_interactive_phase3b_test.go` (AC-1..4)
  - [ ] Sub-test per ID covering the four assertions.
- [ ] Task 4: Audit dependent tests
  - [ ] `rg` for the four IDs in `*_test.go`; update or refactor.
- [ ] Task 5: Smoke verification (manual)
  - [ ] `wails dev`, drop `bmad-create-architecture` onto canvas, drive a single round to "approved", confirm gate satisfaction.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `registry_interactive_phase3b.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code; no CRITICAL/HIGH issues
- [ ] AC validation table populated in PR description
- [ ] Smoke run on at least one of the four planning processes
- [ ] Status flipped to `done` by sprint lead
