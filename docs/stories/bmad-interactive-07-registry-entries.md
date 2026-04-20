# bmad-interactive-07: Registry entries for the four reference interactive processes

**Status:** done
**Domain:** backend
**Size:** M
**Depends On:** bmad-interactive-01, bmad-interactive-03, bmad-interactive-04
**Priority:** P1-high

## Story

As a BMAD workflow author, I want the four canonical interactive processes (`bmad-brainstorming`, `bmad-product-brief`, `bmad-party-mode`, `bmad-advanced-elicitation`) to ship with typed `Mode`, `InputSpecs`, `OutputSpecs`, and `Gate` declarations, so that the canvas renders them correctly and the executor runs them through the interactive path without ad-hoc tmux heuristics.

## Description

Populate the four `ProcessDef` entries in `internal/bmad/registry.go` with the JSON shapes from §10 of the design spec (§10.1 through §10.4). Ship unit-style smoke tests that drive each process through the executor via a mock command runner — no real tmux, no real claude CLI. End-to-end tmux validation is explicitly deferred.

### Scope summary
- Update `bmad-brainstorming` → `InteractIterative` with `GateUserConfirm` (§10.1).
- Update `bmad-product-brief` → `InteractGuided` with `GateUserConfirm` (§10.2).
- Update `bmad-party-mode` → `InteractParty` with `GateUserConfirm` (§10.3).
- Update `bmad-advanced-elicitation` → `InteractIterative` with `GateUserConfirm` (§10.4).
- Add `brain-methods.csv` and `methods.csv` CSV fixtures under `internal/bmad/testdata/` for registry lookups — small (5-10 rows each) is fine; full catalogues live elsewhere.
- Extend `artifactPaths` map if any referenced artifact name (`brainstorm-notes`, `product-brief`, `retro-notes`, `elicitation-notes`) is not already mapped — most already are per `artifacts.go`.

### Non-goals
- No real tmux/claude end-to-end validation (deferred).
- No UI-level verification beyond what S6 covers (S6 runs Playwright against the same registry entries).
- No changes to the upstream BMAD skill markdown (`bmad-brainstorming.md` etc.).

## Developer Notes

### Files to modify
- `internal/bmad/registry.go` — update the four entries.
- `internal/bmad/testdata/brain-methods.csv` (new) — columns `technique_name,description`.
- `internal/bmad/testdata/methods.csv` (new) — columns `method_name,prompt`.
- `internal/bmad/artifacts.go` — confirm mappings for `brainstorm-notes`, `product-brief`, `retro-notes`, `elicitation-notes`. Add any missing entry.
- `internal/bmad/registry_interactive_test.go` (new) — smoke tests for all four processes.
- `internal/bmad/executor_interactive_smoke_test.go` (new) — executor-driven smoke tests with mock runner.

### Exact registry shapes
Use the JSON snippets in §10.1-§10.4 verbatim. Convert to Go struct literals:

```go
{
    ID: "bmad-brainstorming",
    Name: "Brainstorming",
    Phase: PhaseAnalysis,
    AgentRole: RoleAnalyst,
    SkillName: "bmad-brainstorming",
    // ... legacy fields (Inputs, Outputs) kept for back-compat
    Mode: InteractIterative,
    InputSpecs: []InputSpec{
        {ID:"topic",           Source:InputFromUser,     Shape:ShapeFree,   Required:true,  Prompt:"What topic do you want to brainstorm?", MaxLength:500},
        {ID:"approach",        Source:InputFromUser,     Shape:ShapeChoice, Required:true,  Prompt:"How should we pick techniques?", Options:[]string{"user-pick","ai-recommend","random","progressive"}},
        {ID:"technique",       Source:InputFromRegistry, OptionsRef:"registry:brain-methods.csv#technique_name"},
        {ID:"round-response",  Source:InputFromUser,     Shape:ShapeFree,   Prompt:"Add ideas, pivot, or type 'done' when satisfied.", HelpText:"Type 'done' to wrap up; 'skip' to move to the next technique."},
    },
    OutputSpecs: []OutputSpec{{ID:"brainstorm-notes", Target:OutputToFile, ArtifactName:"brainstorm-notes", Description:"Organised brainstorm session notes"}},
    Gate: &IterationGate{Kind:GateUserConfirm, MaxRounds:30, AcceptTokens:[]string{"done","wrap up","finish"}, RejectTokens:[]string{"abort","cancel"}},
},
```

Repeat for the other three per §10.2-§10.4. Preserve existing `Phase` / `AgentRole` / `SkillName` / `Description` / `ModuleID` / `Version` from the current entries — interactive fields are additive.

### Legacy back-compat
Keep the existing `Inputs []string` and `Outputs []string` fields populated with their prior values so any non-interactive callers (sprint status, legacy artifact verification) keep working during the rollout. Example: `bmad-brainstorming` current `Outputs: []string{"brainstorm-notes"}` stays.

### CSV fixtures
`internal/bmad/testdata/brain-methods.csv`:
```csv
technique_name,description
mindmap,Visual branching diagram of ideas
scamper,Substitute/Combine/Adapt/Modify/Put-to-use/Eliminate/Reverse
five-whys,Repeated 'why' drilldown
brainwriting,Silent written idea generation in rounds
lotus-blossom,Central idea plus 8 surrounding themes expanded
```

`internal/bmad/testdata/methods.csv`:
```csv
method_name,prompt
critique,What assumptions might be wrong here?
devils-advocate,Argue the opposite position
red-team,Attack this from an adversary's perspective
pre-mortem,Imagine it failed; what happened?
first-principles,Decompose to foundational truths and rebuild
inversion,What would guarantee the worst outcome?
steelman,Strongest possible version of the counter-argument
second-order,What are the downstream effects two steps out?
```

`registryLookup` (from S2/S3) resolves `registry:<file>#<column>` to the column values joined with newlines and `?random=N` to N random rows. Point the resolver at `internal/bmad/testdata/` for tests; production resolution can be stubbed or use the same path.

### Smoke tests (unit-style)
Each smoke test drives one process through `executeInteractiveNode` via a mock command runner:

```go
func TestSmoke_Brainstorming_ThreeRoundsDone(t *testing.T) {
    h := newInteractiveHarness(t)
    execID, nodeID := h.startSingleNode("bmad-brainstorming")

    h.expectAwaitingInput(nodeID, "topic")
    h.respond(execID, nodeID, "topic", "voice UX")

    h.expectAwaitingInput(nodeID, "approach")
    h.respond(execID, nodeID, "approach", "ai-recommend")

    for r := 1; r <= 2; r++ {
        h.waitForRoundComplete(nodeID, r)
        h.expectAwaitingInput(nodeID, "round-response")
        h.respond(execID, nodeID, "round-response", "keep going")
    }
    h.waitForRoundComplete(nodeID, 3)
    h.expectAwaitingInput(nodeID, "round-response")
    h.respond(execID, nodeID, "round-response", "done")

    h.assertGateSatisfied(nodeID, 3, "accept-token matched")
    h.assertNodeComplete(nodeID)
}
```

Parallel tests for `product-brief` (5 stages → approval), `party-mode` (3 messages → `exit`), `advanced-elicitation` (2 rounds → `x`).

The harness (`internal/bmad/testutil_interactive.go`, new or extended) plumbs:
- a mock `CommandRunner` that swallows `tmux new-session`/`send-keys` and returns success
- a mock `startSession` / `waitForIdle` / `captureRoundOutput` that advances round state deterministically
- event assertion helpers (`expectEvent`, `waitForEvent`)

### Risks / gotchas
- **Legacy `Inputs []string` vs new `InputSpecs`**: keep BOTH populated during the rollout. Delete legacy values ONLY when a follow-up migration confirms no reader depends on them.
- **`OptionsRef` CSV path**: for tests, point the resolver at `internal/bmad/testdata/`; for production, use an embedded `//go:embed` of the CSV or resolve relative to the BMAD skills directory. Recommended: `//go:embed testdata/*.csv` if the CSVs are small — ships with the binary, no filesystem scavenger hunts.
- **Party-mode optional output**: `"optional": true` on the transcript — `verifyOutputs` must NOT fail the node when the file is missing. Confirmed by the test.
- **Gate accept tokens are case-insensitive** (per §4 `containsToken`): tests that submit `"Done"` or `"DONE"` should still trigger gate satisfaction.
- **Elicitation's `method` OptionsRef `?random=5`**: every suspension re-rolls five options. The `PendingPrompt.Options` field carries the resolved-at-suspension-time options — test harness should assert at least 5 options on each re-prompt, but contents may differ.
- **Max round sanity**: the smoke test explicitly must NOT run `MaxRounds` of rounds (brainstorm 30, party 100). Drive 2-3 rounds then accept.

### Reference files
- `internal/bmad/registry.go` — target file (confirmed via use-repo-code skill).
- `internal/bmad/artifacts.go` — `artifactPaths` map to confirm artifact coverage.
- `docs/bmad-interactive-process-schema.md` §10 — canonical registry shapes.

## Acceptance Criteria

**AC-1: All four processes carry the declared interactive fields**
- Given the updated registry
- When each of the four IDs is looked up via `ProcessByID`
- Then `Mode`, `InputSpecs`, `OutputSpecs`, `Gate` match the §10 spec byte-for-byte
- And their legacy `Inputs`/`Outputs` slices are still populated (back-compat)

**AC-2: `iterationInput()` returns the expected per-round spec for each iterative/party process**
- Given `bmad-brainstorming`
- Then `iterationInput()` returns the `round-response` spec
- Given `bmad-party-mode`
- Then `iterationInput()` returns the `message` spec
- Given `bmad-advanced-elicitation`
- Then `iterationInput()` returns either `method` or `apply-changes` (spec-expected: the per-round recurring one)

**AC-3: Brainstorm smoke test completes via accept token**
- Given a single-node workflow using `bmad-brainstorming`
- When the harness responds to `topic`, `approach`, then `"keep going"` for 2 rounds and `"done"` on round 3
- Then `bmad:node:gate_satisfied` fires with round=3 and reason `"accept-token matched"`
- And the node completes
- And `verifyOutputs` succeeds (harness creates `_bmad-output/analysis-artifacts/brainstorm-notes.md`)

**AC-4: Product-brief guided smoke test completes via final approval**
- Given a single-node workflow using `bmad-product-brief`
- When the harness responds to `mode`, 5 stage-responses, then `final-approval = "yes"`
- Then the node completes
- And `bmad:node:gate_satisfied` fires
- And `product-brief` artifact verification succeeds

**AC-5: Party-mode smoke test exits on accept token with optional transcript**
- Given a single-node workflow using `bmad-party-mode`
- When the harness responds to `topic`, then 2 `message` answers, then `message = "exit"`
- Then the node completes
- And `verifyOutputs` does NOT fail even when the transcript file is absent (optional output)

**AC-6: Advanced-elicitation smoke test exits on `x` token**
- Given a single-node workflow using `bmad-advanced-elicitation`
- When the harness responds to `target-content` (via upstream mock), then `method` and `apply-changes` for 1 round, then `method = "x"`
- Then `bmad:node:gate_satisfied` fires with reason mentioning accept-token
- And the node completes

**AC-7: Registry entries resolve `OptionsRef` to non-empty option lists**
- Given the `registry:brain-methods.csv#technique_name` ref resolved via `registryLookup`
- Then the result is a non-empty string containing at least 3 technique names (from the fixture CSV)
- Given the `registry:methods.csv?random=5` ref resolved at suspension time
- Then `PendingPrompt.Options` contains exactly 5 entries from the fixture CSV

**AC-8: Reject token aborts brainstorm smoke test**
- Given a single-node `bmad-brainstorming` workflow
- When the harness responds `topic`, `approach`, then `"abort"` on the first iteration prompt
- Then `bmad:node:aborted` fires with reason `"rejected by user"`
- And the node transitions to `NodeFailed`

## BDD Test Scenarios

```gherkin
Feature: Registry entries for interactive processes

  Scenario: Brainstorming registry shape matches spec
    When ProcessByID("bmad-brainstorming") is queried
    Then the returned ProcessDef has Mode iterative
    And its InputSpecs include IDs [topic, approach, technique, round-response]
    And its Gate Kind is userConfirm with MaxRounds 30
    And its AcceptTokens include "done" and "finish"

  Scenario: Brainstorm smoke: done on round 3
    Given a harness running bmad-brainstorming through executeInteractiveNode
    When the harness responds topic "voice UX"
    And approach "ai-recommend"
    And round-response "keep going" on rounds 1 and 2
    And round-response "done" on round 3
    Then bmad:node:gate_satisfied fires at round 3
    And the node completes

  Scenario: Product brief smoke: final approval
    Given a harness running bmad-product-brief
    When the harness responds mode "guided"
    And five stage-responses
    And final-approval "yes"
    Then the node completes
    And the product-brief artifact verification passes

  Scenario: Party mode smoke: exit token
    Given a harness running bmad-party-mode
    When the harness responds topic "retro", then two messages, then message "exit"
    Then the node completes
    And verifyOutputs accepts the missing optional transcript

  Scenario: Advanced elicitation smoke: x token
    Given a harness running bmad-advanced-elicitation
    And an upstream node output wired to target-content
    When the harness responds method "critique", apply-changes "yes"
    And method "x" on round 2
    Then bmad:node:gate_satisfied fires at round 2

  Scenario: OptionsRef random=5 produces exactly 5 options
    Given an InputSpec with OptionsRef "registry:methods.csv?random=5"
    When registryLookup resolves it at suspension time
    Then the PendingPrompt Options slice has length 5
    And every entry comes from methods.csv

  Scenario: Reject token aborts the process
    Given a harness running bmad-brainstorming
    When the harness responds round-response "abort"
    Then bmad:node:aborted fires with reason "rejected by user"
    And the node status is failed
```

## Tasks / Subtasks

- [ ] Task 1: Update registry entries (AC-1, AC-2)
  - [ ] `bmad-brainstorming` per §10.1
  - [ ] `bmad-product-brief` per §10.2
  - [ ] `bmad-party-mode` per §10.3
  - [ ] `bmad-advanced-elicitation` per §10.4
  - [ ] Keep legacy `Inputs`/`Outputs` fields populated for back-compat
- [ ] Task 2: CSV fixtures + registryLookup wiring (AC-7)
  - [ ] Create `internal/bmad/testdata/brain-methods.csv`
  - [ ] Create `internal/bmad/testdata/methods.csv`
  - [ ] Add `//go:embed testdata/*.csv` and route `registryLookup` through the embedded FS
- [ ] Task 3: Confirm artifactPaths entries (AC-3, AC-4)
  - [ ] Verify `brainstorm-notes`, `product-brief`, `retro-notes`, `elicitation-notes` all resolvable
  - [ ] Add any missing
- [ ] Task 4: Interactive test harness (AC-3..AC-8)
  - [ ] `newInteractiveHarness(t)` helper in `testutil_interactive.go`
  - [ ] Mock CommandRunner; mock session/idle/capture primitives
  - [ ] `respond`, `expectAwaitingInput`, `waitForRoundComplete`, `assertGateSatisfied`, `assertNodeComplete`
- [ ] Task 5: Smoke tests per process (AC-3, AC-4, AC-5, AC-6, AC-8)
  - [ ] `TestSmoke_Brainstorming_ThreeRoundsDone`
  - [ ] `TestSmoke_Brainstorming_RejectAbort`
  - [ ] `TestSmoke_ProductBrief_GuidedApproval`
  - [ ] `TestSmoke_PartyMode_ExitToken`
  - [ ] `TestSmoke_AdvancedElicitation_XAccept`
- [ ] Task 6: Registry validation test (AC-1, AC-2)
  - [ ] Table test asserting struct shape matches §10 for all four entries
  - [ ] `iterationInput()` resolves expected spec

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on modified files; no CRITICAL/HIGH findings
- [ ] Legacy registry consumers (sprint status, artifact verification) still green
- [ ] Story status flipped to `done`
- [ ] Changes committed on branch
