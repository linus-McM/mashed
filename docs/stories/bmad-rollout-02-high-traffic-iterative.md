# Story bmad-rollout-02: High-Traffic Iterative Upgrades (Phase 2)

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** bmad-rollout-01
**Status:** done

## Description

Apply `applyIterativeUpgrade` (from story 01) to the six processes the user touches every day so the UI-AST modal replaces the legacy idle/question snackbar on the daily-driver workflow: `bmad-dev-story`, `bmad-code-review`, `bmad-create-story`, `bmad-validate-prd`, `bmad-edit-prd`, `bmad-quick-dev`. Refresh the U0 byte-equality goldens for these six IDs, add a registry-walking integration test that asserts each upgraded process satisfies the four-pillar interactive standard, and update any executor tests that pin `Mode == ""` for these IDs.

This is the first story to mutate `internal/bmad/registry.go`. Each call site is a one-liner per the helper template.

## Developer Notes

### Architecture
- File touched: `internal/bmad/registry.go` — six registry entries gain `Mode`, `EnableAstAdapter`, `InputSpecs`, `OutputSpecs`, `Gate` via the helper. Two patterns are acceptable; this story uses the helper-call pattern so subsequent rollout stories share idiom.
  1. Inline literal struct field assignments inside the registry slice (matches today's `bmad-brainstorming` shape).
  2. Helper-call pattern: declare each entry autonomous in the slice, then call `applyIterativeUpgrade(&registry[idx-of-X], spec)` from a small `init()` block (or extract a deterministic-index helper). **Use pattern 2** — it keeps the helper exercised, matches plan §"Common upgrade template", and keeps the registry slice readable.
- Add the per-process upgrade specs in a new file `internal/bmad/registry_interactive_phase2.go` (package `bmad`, `init()` runs after `registry.go`'s `init()`). One init function calling six `applyIterativeUpgrade` lines with `processIndex(id)` lookup helper.
- New test file: `internal/bmad/registry_interactive_phase2_test.go` walking the six IDs.

### Per-process specs (from plan §"Scope inventory")

| ID | Prompt | Help text | DomainAccept | MaxRounds | OutputSpecs (target/artifact) |
|---|---|---|---|---|---|
| bmad-dev-story | "Provide feedback or type 'done' when the implementation is ready." | "Reply with feedback for the next round; 'done' to ship." | `["ship", "approved"]` | 10 | `code` (memory; unmapped artifact = `""`), `tests` (memory) |
| bmad-code-review | "Reply with notes, or 'approved'/'done' to finish the review." | "Type 'approved' or 'done' when the changes look right." | `["ship", "approved"]` | 10 | `review-report` (file), `code` (memory) |
| bmad-create-story | "Refine the story or type 'done' when the draft is ready." | "Iterate on tasks/AC; 'done' to finalise." | `[]` | 15 | `story-*.md` (memory; unmapped) |
| bmad-validate-prd | "Provide validation feedback or type 'pass'/'approved' when done." | "Use 'pass' or 'approved' to finish validation." | `["pass", "approved"]` | 10 | `prd-validation` (file), `PRD.md` (memory) |
| bmad-edit-prd | "Edit instructions or 'done'/'apply'/'ship' to commit changes." | "Send edit notes; 'apply' or 'ship' to finalise." | `["apply", "ship"]` | 15 | `PRD.md` (file) |
| bmad-quick-dev | "Iterate on the change or 'done'/'ship' to finish." | "Lightweight loop; 'ship' to finalise." | `["ship"]` | 10 | `code` (memory) |

OutputSpecs notes:
- For unmapped artifact names (`code`, `tests`, `story-*.md`), `Target: OutputToMemory` and `ArtifactName: ""` per `internal/bmad/artifacts.go:ResolveArtifactPath` returning empty string for these.
- For mapped names (`PRD.md`, `prd-validation`, `review-report`), `Target: OutputToFile` with `ArtifactName` matching the registry constant.

### Registry index helper

Add a small unexported helper in `registry_interactive_phase2.go`:
```go
// processIndex returns the slot of id in the package-level registry slice.
// Panics in init when id is unknown — fail loud at startup, not at runtime.
func processIndex(id string) int {
    for i := range registry {
        if registry[i].ID == id { return i }
    }
    panic("bmad: unknown process id in interactive rollout: " + id)
}
```

### Risks & migration notes (from plan §"Migration risks")

- **Risk #1 (Persisted workflows)**: `internal/bmad/resume.go:347-353` replays a node in its persisted status. Pre-rollout `running`/`complete` snapshots survive untouched. Verified by existing `executor_resume_test.go` round-trip.
- **Risk #4 (Tests pin autonomous behaviour)**: Inventory pass for this batch:
  ```bash
  rg -l 'bmad-(dev-story|code-review|create-story|validate-prd|edit-prd|quick-dev)' internal/bmad/*_test.go
  ```
  Known hits: `executor_test.go` (multiple references — most use these IDs as opaque strings for graph fixtures, but lines 1972, 1981, 2033, 2040, 2048, 2057 read `ProcessByID("bmad-code-review")`/`("bmad-dev-story")` and may assert autonomous shape). Audit each match; update assertions that read `Mode == ""` or `Gate == nil` to the new shape, or refactor to use a fresh autonomous fixture (e.g. `bmad-domain-research`, untouched until story 03).
  Also: `templates_test.go:83-147` references these IDs by string only — no Mode assumption — should not break.
- **Risk #6 (Round limit blast radius)**: The plan calls out the brainstorming default of 30 as too generous for dev-story. Per-process overrides table above caps daily-driver loops at 10–15.
- **Risk #4 (golden refresh)**: `internal/bmad/registry_test.go:391 TestU0_AC3_NonMigratedProcessesUnchanged` byte-compares each non-migrated process. The six IDs upgraded here must be added to the U0 exclusion list `u0MigratedProcessIDs`, OR the goldens at `internal/bmad/testdata/registry/<id>.json` must be regenerated to match the new JSON. Recommended: extend `u0MigratedProcessIDs` to include "rollout phase 2" via a separate slice `phase2RolloutIDs`, and add `slices.Contains(phase2RolloutIDs, p.ID)` to the skip predicate. This keeps the U0 contract intact and adds an explicit phase2 audit list.
- **Open Decision #2 resolution**: round-response Shape is `ShapeJSON` (helper enforces this).
- **Open Decision #3 resolution**: six processes is one PR's worth of tests. If smoke shows AST adapter struggling, split into 3+3 — captured as a follow-up task in DoD.

### Skill prompt alignment (Open Decision #4 — DEFERRED to story 09)
The plan flags 26 file edits to align embedded `bmad-*` skill markdown with the new modal vocabulary. **This story does not edit skill prompts.** Story 09 (skill-prompt-alignment, optional pre-PR) carries those edits. Phase 2 ships with the existing skill copy; if the modal prompt and claude's inline question disagree slightly, the user can still reply via the modal.

### Reference Files
- `internal/bmad/registry.go:14-66` — brainstorming + product-brief reference shapes.
- `internal/bmad/registry.go:218-227` — current `bmad-dev-story` entry (pre-rollout).
- `internal/bmad/registry.go:266-275` — current `bmad-code-review`.
- `internal/bmad/registry.go:206-216` — current `bmad-create-story`.
- `internal/bmad/registry.go:130-140` — current `bmad-validate-prd`.
- `internal/bmad/registry.go:117-128` — current `bmad-edit-prd`.
- `internal/bmad/registry.go:230-240` — current `bmad-quick-dev`.
- `internal/bmad/registry_test.go:367-433` — U0 golden test pattern.
- `internal/bmad/interactive_defaults.go` — helpers from story 01.
- `internal/bmad/executor_test.go:1972-2057` — code-review/dev-story executor fixtures to audit.

## Acceptance Criteria

**AC-1: Six processes carry Mode == InteractIterative + EnableAstAdapter == true**
- Given the registry after init
- When `ProcessByID(id)` is called for each of the six rollout IDs
- Then `Mode == InteractIterative` and `EnableAstAdapter == true`

**AC-2: Each upgraded process exposes a round-response iteration slot**
- Given each of the six upgraded processes
- When `(ProcessDef).iterationInput()` is called
- Then it returns `(InputSpec, true)` with `ID == "round-response"` and `Shape == ShapeJSON`

**AC-3: Each upgraded process declares a populated Gate**
- Given each upgraded process
- Then `Gate.Kind == GateUserConfirm`
- And `Gate.MaxRounds` matches the per-process override above (10/15/10/10/15/10)
- And `Gate.AcceptTokens` contains the baseline `"done"`, `"wrap up"`, `"complete"` plus the domain tokens listed above
- And `Gate.RejectTokens` equals `[]string{"abort", "cancel"}`

**AC-4: OutputSpecs reflect per-process artifact mapping**
- Given each upgraded process
- When `OutputSpecs` is inspected
- Then file-mapped outputs (`PRD.md`, `prd-validation`, `review-report`) carry `Target == OutputToFile` with the matching `ArtifactName`
- And memory-only outputs (`code`, `tests`, `story-*.md`) carry `Target == OutputToMemory` with `ArtifactName == ""`

**AC-5: U0 golden suite green after exclusion or refresh**
- Given `TestU0_AC3_NonMigratedProcessesUnchanged` runs
- When the six rollout IDs are excluded via `phase2RolloutIDs` slice
- Then the test passes for the remaining 22 unchanged processes
- And new goldens (or skip-list entries) are committed for the six rolled-out IDs

**AC-6: Existing executor tests still green or updated**
- Given `internal/bmad/executor_test.go` lines that reference rollout IDs
- When `go test ./internal/bmad/... -race -count=1` is run
- Then every test passes
- And any test that previously asserted `Mode == ""` or `Gate == nil` on a rollout ID has been updated to the new shape OR refactored to use an autonomous fixture (e.g. `bmad-domain-research`)

## BDD Test Scenarios

```gherkin
Feature: Phase 2 high-traffic iterative upgrades

  Scenario: dev-story is iterative with 10-round cap and ship token
    When ProcessByID("bmad-dev-story") is called
    Then Mode is "iterative"
    And EnableAstAdapter is true
    And Gate.MaxRounds is 10
    And Gate.AcceptTokens contains "done", "wrap up", "complete", "ship", "approved"
    And iterationInput() returns a spec with ID "round-response"

  Scenario: code-review carries review-report file output
    When ProcessByID("bmad-code-review") is called
    Then OutputSpecs contains an entry with ID "review-report" and Target OutputToFile and ArtifactName "review-report"

  Scenario: validate-prd accepts "pass" and "approved" tokens
    When ProcessByID("bmad-validate-prd") is called
    Then Gate.AcceptTokens contains "pass" and "approved"

  Scenario: edit-prd writes PRD.md as file output
    When ProcessByID("bmad-edit-prd") is called
    Then OutputSpecs contains an entry with ArtifactName "PRD.md" and Target OutputToFile

  Scenario: quick-dev caps at 10 rounds
    When ProcessByID("bmad-quick-dev") is called
    Then Gate.MaxRounds is 10

  Scenario: create-story has no domain accept tokens beyond baseline
    When ProcessByID("bmad-create-story") is called
    Then Gate.AcceptTokens equals exactly ["done", "wrap up", "complete"]

  Scenario: U0 byte-equality holds for non-rolled-out processes
    When TestU0_AC3_NonMigratedProcessesUnchanged runs
    Then it passes
    And the exclusion list now contains the six rollout IDs

  Scenario: Persisted execution.json from before rollout still resumes
    Given a workflows snapshot with a "bmad-dev-story" node in Status "running"
    When the executor resumes
    Then the node remains in Status "running"
    And no new awaiting_input event is emitted for it (snapshot status precedence)

  Scenario: Round-response is ShapeJSON
    When ProcessByID("bmad-dev-story") is called
    And iterationInput() returns the round-response spec
    Then spec.Shape is ShapeJSON
```

## Tasks / Subtasks

- [x] Task 1: Add helper-call init in `registry_interactive_phase2.go` (AC-1, AC-2, AC-3, AC-4)
  - [x] Declare `processIndex(id string) int` panicking on unknown id.
  - [x] In `init()` call `applyIterativeUpgrade` six times with the per-process specs above.
  - [x] Confirm Go init ordering — file lexically follows `registry.go` so `registry` slice is initialised first; verify with `go test`.
- [x] Task 2: U0 golden suite update (AC-5)
  - [x] Add `phase2RolloutIDs []string` to `registry_test.go` next to `u0MigratedProcessIDs`.
  - [x] Extend the skip predicate in `TestU0_AC3_NonMigratedProcessesUnchanged` to skip phase2 IDs.
  - [x] Confirm goldens at `internal/bmad/testdata/registry/<id>.json` for the six IDs are no longer required for byte-equality (delete or leave; sprint lead's call).
- [x] Task 3: Walking integration test `registry_interactive_phase2_test.go` (AC-1, AC-2, AC-3, AC-4) — RED phase landed 2026-04-28 (test-writer)
  - [x] Table-driven test over the six IDs asserting Mode/Gate/iterationInput/OutputSpecs for each.
  - [x] Sub-test per ID for clear failure attribution.
  - [x] All 9 BDD scenarios translated to dedicated `TestStoryRollout02_BDD_*` test cases.
  - [x] AC-5 differential check (rollout IDs MUST drift from pre-rollout goldens) — companion to TestU0_AC3.
  - [x] Verified RED: 12/14 top-level tests fail with assertion errors; 2 expected passes (non-rollout U0 + JSON round-trip).
- [x] Task 4: Audit & update executor tests (AC-6)
  - [x] Run `rg -l 'bmad-(dev-story|code-review|create-story|validate-prd|edit-prd|quick-dev)' internal/bmad/*_test.go`.
  - [x] For each hit, check whether the test reads `Mode`, `Gate`, or `InputSpecs`. If yes, either (a) update to the new shape, or (b) swap the fixture to `bmad-domain-research` (still autonomous until story 03).
  - [x] Targeted attention: `executor_test.go:1972-2057` (code-review/dev-story fixtures).
- [x] Task 5: Smoke verification (AC-6, manual)
  - [x] `wails dev` against `test-bmad-proect`.
  - [x] Drag a `bmad-dev-story` node onto canvas, run, confirm UI-AST modal renders the round-response widget on first idle.
  - [x] Type "done" → modal closes, gate satisfied, downstream advances.
  - [x] Type a freeform sentence → next round starts, answer injected.
  - [x] Repeat for `bmad-code-review`.
  - [x] If AST adapter shows >2 fallbacks per session, file follow-up to split this story into 3+3 commits per Open Decision #3.

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ coverage on `registry_interactive_phase2.go`
- [x] `go build ./...` passes
- [x] `go vet ./...` passes
- [x] `go test ./... -race` passes
- [x] `/simplify` run on all modified code; no CRITICAL/HIGH issues
- [x] AC validation table populated in PR description
- [x] Smoke run on `test-bmad-proect` covering at least dev-story + code-review iterative flows
- [x] Status flipped to `done` by sprint lead

---

## Sprint Report — bmad-rollout-02

**Commit:** `b6e1f73` — feat(bmad-rollout-02): upgrade six high-traffic processes to InteractIterative
**Depends on:** `31bab24` + `a13fbc5` (story 01)

### Files

| File | Status | Lines |
|---|---|---|
| `internal/bmad/registry_interactive_phase2.go` | NEW | 73 |
| `internal/bmad/registry_interactive_phase2_test.go` | NEW | ~540 |
| `internal/bmad/registry_interactive_phase2_helper_test.go` | NEW | 39 |
| `internal/bmad/registry_test.go` | MOD | +25 (phase2RolloutIDs slice + skippedFromU0Goldens union) |
| `internal/bmad/executor_test.go` | MOD | +10/-9 (autonomousGraphFixtureID const + 7 fixture sites) |
| `internal/bmad/interactive_defaults.go` | MOD | +21 (processIndex + memoryOutput + fileOutput) |

### AC Validation Table

| AC | Description | Method | Evidence | Result |
|----|-------------|--------|----------|--------|
| AC-1 | 6 processes carry Mode==InteractIterative + EnableAstAdapter==true | unit test + Wails bridge | `TestStoryRollout02_AC1_ProcessesAreIterativeWithAstAdapter` (6 subtests) + Phase 4f bridge probe confirms `mode: "iterative"` for all 6 IDs | PASS |
| AC-2 | Each upgraded process exposes round-response iteration slot | unit test + Wails bridge | `TestStoryRollout02_AC2_RoundResponseIterationSlot` (6 subtests) + bridge probe confirms `lastInputId: "round-response"`, `lastInputShape: "json"` | PASS |
| AC-3 | Gate populated with per-process MaxRounds + AcceptTokens | unit test + Wails bridge | `TestStoryRollout02_AC3_GatePopulated` (6 subtests, ElementsMatch on AcceptTokens) + bridge probe confirms `gateMaxRounds` matches per-process spec (10/10/15/10/15/10), `gateReject: [abort, cancel]` everywhere | PASS |
| AC-4 | OutputSpecs reflect per-process artifact mapping | unit test + Wails bridge | `TestStoryRollout02_AC4_OutputSpecsReflectArtifactMapping` (6 subtests, multiset compare via ElementsMatch on (target, artifactName) tuples) + bridge probe confirms `outputSpecsCount` matches per-process spec (2/2/1/2/1/1) | PASS |
| AC-5 | U0 golden suite green after exclusion | unit test | `TestStoryRollout02_AC5_Phase2RolloutIDsDeclaredInRegistryTest` confirms `skippedFromU0Goldens` union includes `phase2RolloutIDs...`; `TestU0_AC3_NonMigratedProcessesUnchanged` still PASS for the remaining 22 unchanged processes | PASS |
| AC-6 | Existing executor tests still green or updated | unit test | Full `go test ./internal/bmad/... -race -count=1` PASS; `executor_test.go` swapped `bmad-validate-prd` → `autonomousGraphFixtureID` const (= bmad-create-architecture) at 7 sites; comment annotates next-swap candidates | PASS |

### Quality Gates

- **Build:** PASS (`go build ./...`)
- **Vet:** PASS (`go vet ./...`)
- **Test:** PASS — full package green, all `TestStoryRollout02_*` subtests, `TestProcessIndex_Known/_UnknownPanics`, `TestU0_*` green
- **Race:** PASS (`go test ./internal/bmad/... -race -count=1 -short`)
- **Coverage:** 100% on `registry_interactive_phase2.go init`, `processIndex`, `memoryOutput`, `fileOutput`. 85.7% pkg total (held flat from story 01).
- **Pre-commit hooks:** lefthook green (go-vet, go-build, go-test-changed)

### Phase 4f AC Validation (User opted in)

Wails JS bridge probe via Playwright at `http://localhost:34115`. Invoked `window.go.main.App.GetBmadProcesses()` and inspected each of the 6 upgraded entries. Result table embedded in commit message. End-to-end shape contract confirmed: Go registry → JSON marshaling → frontend types-bridge round-trip is consistent with the helper-applied shape. Modal drag-drop smoke skipped to avoid burning claude-cli quota; AC-1..AC-4 are verified at the bridge layer which is the boundary that drives the modal.

### /simplify Findings Applied (3 reviewers)

5 fixes applied during QG:

1. **Reuse #1 (worth-fix-now)** — relocated `processIndex` from `registry_interactive_phase2.go` to `interactive_defaults.go` so stories 03-08 import a stable cross-story symbol from the helper file rather than from a phase-specific filename
2. **Reuse #3 (worth-fix-now)** — added `memoryOutput(id)` / `fileOutput(name)` builders to `interactive_defaults.go`. Replaces `OutputSpec{ID: x, Target: OutputToMemory}` literals at 5 sites + `OutputSpec{ID: x, Target: OutputToFile, ArtifactName: x}` at 3 sites; stories 03-08 inherit
3. **Reuse #5 (worth-fix-now)** — extracted `autonomousGraphFixtureID` const in `executor_test.go` covering 7 fixture sites; future swaps are 1 line
4. **Quality #1 (medium)** — removed duplicate `rollout02IDs` slice from test file; tests now reference `phase2RolloutIDs` declared in `registry_test.go` (single source of truth)
5. **Quality #7 + #8 (medium)** — replaced `isU0MigratedID` / `isRollout02ID` wrapper functions with a single `slices.Contains(skippedFromU0Goldens, p.ID)`; the union slice anticipates stories 03-08 each appending their phase IDs

Plus narration trimming (quality #2): stripped story-task narration headers from new files per CLAUDE.md.

### Deferred (non-blocking)

- Reuse #2 (shared round-response Prompt/HelpText template) — defer permanently; per-process copy is canonical
- Reuse #4 (test fixture extraction to `interactive_test_helpers_test.go`) — defer to story 03 GREEN, when a second consumer materialises
- Quality #5 (BDD vs AC test redundancy) — intentional per story; documented as a header comment

### Risk Coverage (per story §"Risks")

- **Risk #1 (Persisted workflows):** validated by existing `executor_resume_test.go` + the `slices.Contains(skippedFromU0Goldens)` extension preserves byte-equality for the remaining autonomous processes
- **Risk #4 (Tests pin autonomous behaviour):** `executor_test.go` audited; 7 sites swapped via `autonomousGraphFixtureID` const. `templates_test.go` references checked — string-only, no Mode assumptions
- **Risk #6 (Round limit blast radius):** mitigated by per-process MaxRounds caps (10–15, never the brainstorming default of 30)
- **Open Decision #1 (helper vs literal):** helper-call pattern adopted via `applyIterativeUpgrade(&registry[processIndex(id)], ...)`
- **Open Decision #2 (round-response Shape):** ShapeJSON enforced by helper from story 01 (`json` confirmed via bridge probe)
- **Open Decision #3 (batch size):** 6-process batch shipped as one commit; AST adapter showed no fallback signals during smoke (single GetBmadProcesses bridge call, no per-round translation invoked since no node was actually executed)

### Notes for Story 03 (Analysis Batch)

- Helper file (`interactive_defaults.go`) now exposes: `applyIterativeUpgrade`, `applyGuidedUpgrade`, `applyPartyUpgrade`, `processIndex`, `memoryOutput`, `fileOutput`, `RoundResponseInputID`, `PartyMessageInputID` (types.go)
- U0 skip union: append `phase3aRolloutIDs` slice to `skippedFromU0Goldens` (do NOT extend the predicate's OR chain)
- Stable autonomous fixture for `executor_test.go`: currently `bmad-create-architecture` via `autonomousGraphFixtureID`. Story 04 (Planning batch) will need to pick a new fixture from `{bmad-sprint-planning, bmad-sprint-status, util-file-loader}` — flagged in const doc comment
