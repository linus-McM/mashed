# Story bmad-rollout-08: Phase 5 — Party Upgrades + applyPartyUpgrade Wiring

**Priority:** P3-low
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** bmad-rollout-01
**Status:** ready

## Description

Apply `applyPartyUpgrade` (helper from story 01) to the three autonomous processes the plan classifies as Party multi-agent collaboration: `bmad-retrospective`, `bmad-web-orchestrator`, `bmad-game-dev-studio`. These mirror the `bmad-party-mode` reference shape from §10.3 — a one-shot `topic` slot plus a recurring `message` slot, with a 100-round cap and `["exit", "done", "wrap up"]` accept tokens.

This is the lowest-priority phase per plan §"Phasing" — Party is more experimental and rarely-touched. It is independent of stories 02–06 and can run after story 01 only.

## Developer Notes

### Architecture
- New file: `internal/bmad/registry_interactive_phase5.go` running `init()` after registry init.
- Three `applyPartyUpgrade` calls via `processIndex(id)`.
- Append the three IDs to the rollout skip slice in `registry_test.go`.
- New test file: `internal/bmad/registry_interactive_phase5_test.go`.

### Per-process specs (from plan §"Autonomous → upgrade to Party")

**`bmad-retrospective`** — facilitator + participant agents.

```go
Topic: InputSpec{
    ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true,
    Prompt: "What do you want the retrospective to focus on?",
    HelpText: "Sprint scope, blockers to discuss, or 'general'.",
    MaxLength: 1000,
},
RoundPrompt: "Your turn. Type 'done' or 'exit' to end the retrospective.",
HelpText:    "[exit] to wrap up · [done] to close the round.",
OutputSpecs: []OutputSpec{
    {ID: "retro-notes", Target: OutputToFile, ArtifactName: "retro-notes", Optional: true},
},
```

**`bmad-web-orchestrator`** — orchestrator + worker agents.

```go
Topic: InputSpec{
    ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true,
    Prompt: "What feature should the web team build?",
    HelpText: "Reference architecture/epics; team will decompose.",
    MaxLength: 1500,
},
RoundPrompt: "Your turn. Type 'done' or 'exit' to end the session.",
HelpText:    "[exit] to wrap up · [done] to close the round.",
OutputSpecs: []OutputSpec{
    {ID: "code", Target: OutputToMemory, ArtifactName: ""},  // unmapped
},
```

**`bmad-game-dev-studio`** — director + asset team.

```go
Topic: InputSpec{
    ID: "topic", Source: InputFromUser, Shape: ShapeFree, Required: true,
    Prompt: "What game / feature is the studio building?",
    HelpText: "Reference the PRD; the studio will branch into specialists.",
    MaxLength: 1500,
},
RoundPrompt: "Your turn. Type 'done' or 'exit' to end the session.",
HelpText:    "[exit] to wrap up · [done] to close the round.",
OutputSpecs: []OutputSpec{
    {ID: "code", Target: OutputToMemory, ArtifactName: ""},  // unmapped
},
```

Notes:
- Helper sets `Mode = InteractParty`, `EnableAstAdapter = true`, default `Gate.MaxRounds = 100`, default `Gate.AcceptTokens = ["exit", "done", "wrap up"]`.
- Helper assembles `InputSpecs = [topic, message]` where the message slot is constructed by the helper (Source=InputFromUser, Shape=ShapeJSON, Required=false, Prompt=spec.RoundPrompt, HelpText=spec.HelpText) — `iterationInput()` must return the message slot post-helper.
- `code` is unmapped per `internal/bmad/artifacts.go:ResolveArtifactPath`; use `OutputToMemory`. `retro-notes` is mapped → `OutputToFile`. `Optional: true` matches party-mode reference §10.3.

### Risks & migration notes

- **Risk #1 (Persisted workflows)**: Snapshot status precedence shields old runs.
- **Risk #4 (Tests pinning autonomous behaviour)**:
  ```bash
  rg -l 'bmad-(retrospective|web-orchestrator|game-dev-studio)' internal/bmad/*_test.go
  ```
  Audit hits.
- **Risk #4 (golden refresh)**: Add the three IDs to rollout skip slice → 26 total IDs after Phase 5 (closing the rollout).
- **Risk #5 (Frontend modal copy)**: Plan calls for a Playwright smoke per batch. Drive `bmad-retrospective` through topic + one message + "exit" to verify.
- **Risk #6 (Round limit)**: 100 rounds is a generous safety ceiling for free-form party chat — matches `bmad-party-mode` reference.
- **Risk #7 (Adapter timeout amplification)**: Party processes are user-paced (one message per turn) — adapter call rate is low.
- **Open Decision #4 (skill prompts)**: Deferred to story 09. Party-mode skill markdown for these three may need touch-ups so claude's facilitator copy matches the modal vocabulary.

### Reference Files
- `internal/bmad/registry.go:368-389` — `bmad-party-mode` reference Party shape (§10.3).
- `internal/bmad/registry.go:289-300` — current `bmad-retrospective` (autonomous).
- `internal/bmad/registry.go:439-450` — current `bmad-web-orchestrator`.
- `internal/bmad/registry.go:451-462` — current `bmad-game-dev-studio`.
- `internal/bmad/artifacts.go:artifactPaths` — confirm `retro-notes` is mapped, `code` is unmapped.
- `internal/bmad/interactive_defaults.go` (story 01) — `applyPartyUpgrade` and `PartyUpgradeSpec`.

## Acceptance Criteria

**AC-1: Three processes upgraded to InteractParty**
- Given the registry after init
- When `ProcessByID(id)` is called for each of the three IDs
- Then `Mode == InteractParty` and `EnableAstAdapter == true`

**AC-2: InputSpecs is [topic, message] with helper-assembled message slot**
- Given each upgraded process
- When `InputSpecs` is inspected
- Then the first entry is the per-process Topic spec (Required=true, ShapeFree)
- And the second entry has `ID == "message"`, `Source == InputFromUser`, `Shape == ShapeJSON`, `Required == false`

**AC-3: Iteration slot is the message spec**
- Given each upgraded process
- When `iterationInput()` is called
- Then `(spec, true)` is returned with `ID == "message"`

**AC-4: Gate populated with party defaults**
- Given each upgraded process
- Then `Gate.Kind == GateUserConfirm`
- And `Gate.MaxRounds == 100`
- And `Gate.AcceptTokens` contains `"exit"`, `"done"`, `"wrap up"`

**AC-5: OutputSpecs reflect mapped vs unmapped artifacts**
- Given each upgraded process
- Then `bmad-retrospective` has one `OutputToFile` entry with `ArtifactName == "retro-notes"` and `Optional == true`
- And `bmad-web-orchestrator` and `bmad-game-dev-studio` each have one `OutputToMemory` entry with `ArtifactName == ""`

**AC-6: U0 skip list extended; non-rolled-out byte-equality intact**
- Given `TestU0_AC3_NonMigratedProcessesUnchanged` runs
- Then the rollout skip slice contains all 26 rollout IDs (20 iterative + 3 guided + 3 party)
- And the test passes

**AC-7: Playwright smoke for retrospective**
- Given `wails dev` running with `bmad-retrospective`
- When a node is dropped, the topic is provided, one message is exchanged, and "exit" is typed
- Then the gate satisfies and the node completes

## BDD Test Scenarios

```gherkin
Feature: Phase 5 party upgrades

  Scenario: retrospective is party with retro-notes file output
    When ProcessByID("bmad-retrospective") is called
    Then Mode is "party"
    And EnableAstAdapter is true
    And InputSpecs first entry has ID "topic"
    And iterationInput() returns a spec with ID "message"
    And OutputSpecs has one entry with ArtifactName "retro-notes" and Optional true

  Scenario: web-orchestrator carries unmapped code output
    When ProcessByID("bmad-web-orchestrator") is called
    Then OutputSpecs has one entry with Target OutputToMemory and ArtifactName ""

  Scenario: game-dev-studio carries unmapped code output
    When ProcessByID("bmad-game-dev-studio") is called
    Then OutputSpecs has one entry with Target OutputToMemory and ArtifactName ""

  Scenario: All Party processes default to 100-round cap and exit token
    When ProcessByID is called for each Party ID
    Then Gate.MaxRounds is 100
    And Gate.AcceptTokens contains "exit" and "done" and "wrap up"

  Scenario: U0 skip list contains all 26 rollout IDs
    When TestU0_AC3_NonMigratedProcessesUnchanged runs
    Then the skip predicate matches 26 rollout IDs (20 + 3 + 3)
    And the test passes for the remaining ~6 unchanged processes (sprint-planning, sprint-status, util-file-loader, util-multi-file-loader, plus the original four reference processes)

  Scenario: iteration slot survives helper invocation
    When ProcessByID("bmad-retrospective") is called
    Then iterationInput() returns a spec with Source InputFromUser and Shape ShapeJSON

  Scenario: Frontend smoke — retrospective party flow
    Given the Wails dev app running
    And a "bmad-retrospective" node on the canvas
    When the node executes and the topic modal opens
    And the user provides a topic
    Then the message modal opens for round 1
    When the user types "exit"
    Then the gate satisfies and the node turns complete
```

## Tasks / Subtasks

- [ ] Task 1: Create `internal/bmad/registry_interactive_phase5.go` (AC-1..5)
  - [ ] Three `applyPartyUpgrade` calls via `processIndex(id)` matching the per-process specs above.
  - [ ] Confirm OutputSpecs match mapped/unmapped table.
- [ ] Task 2: Extend rollout skip slice (AC-6)
  - [ ] Append the three Party IDs; total now 26.
- [ ] Task 3: Walking integration test `registry_interactive_phase5_test.go` (AC-1..5)
  - [ ] Sub-test per ID covering Mode/InputSpecs/iterationInput()/Gate/OutputSpecs.
- [ ] Task 4: Audit prior tests
  - [ ] `rg -l 'bmad-(retrospective|web-orchestrator|game-dev-studio)' internal/bmad/*_test.go`.
- [ ] Task 5: Playwright smoke for retrospective (AC-7)
  - [ ] Add a spec under `tests/ac/` driving the topic + message + exit Party flow.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `registry_interactive_phase5.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code; no CRITICAL/HIGH issues
- [ ] AC validation table populated in PR description
- [ ] Playwright smoke spec added for `bmad-retrospective`
- [ ] Status flipped to `done` by sprint lead
