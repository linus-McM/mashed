# BMAD Interactive Rollout — Plan

## Goal

Bring every `bmad-*` process in `internal/bmad/registry.go` up to the
interactive standard so the UI-AST modal replaces the legacy idle/question
snackbar uniformly across the workflow canvas. Today only four processes
are wired (`bmad-brainstorming`, `bmad-product-brief`,
`bmad-advanced-elicitation`, `bmad-party-mode`); the remaining ~28 fall
through to `executeProcessNode` and surface only the legacy snackbar.

A process meets the standard when it carries:

- `Mode` ∈ {`InteractIterative`, `InteractGuided`, `InteractParty`}
- `EnableAstAdapter: true`
- At least one `InputSpec` with `Source: InputFromUser` so
  `proc.iterationInput()` returns a spec for the round loop
- A `Gate` with `Kind: GateUserConfirm` (or `GateArtifactExists` /
  `GateExpression` where artifact-driven completion is the natural exit),
  populated `MaxRounds`, and a `AcceptTokens` set the user can type to
  finish

## Scope inventory

Current registry pass (`registry.go`):

**Already interactive — no work needed:**

| ID | Mode | Notes |
|---|---|---|
| bmad-brainstorming | Iterative | reference shape (§10.1) |
| bmad-product-brief | Guided | three staged questions |
| bmad-advanced-elicitation | Iterative | per-round method picker |
| bmad-party-mode | Party | multi-agent free-form |

**Autonomous → upgrade to Iterative (round-loop with user confirm):**

| ID | Domain | Suggested AcceptTokens |
|---|---|---|
| bmad-domain-research | analysis | `done`, `wrap up`, `complete` |
| bmad-market-research | analysis | `done`, `complete` |
| bmad-technical-research | analysis | `done`, `complete` |
| bmad-edit-prd | planning | `done`, `apply`, `ship` |
| bmad-validate-prd | planning | `pass`, `done`, `approved` |
| bmad-create-ux-design | planning | `done`, `approved`, `ship` |
| bmad-create-architecture | planning | `done`, `approved`, `ship` |
| bmad-check-implementation-readiness | planning | `ready`, `done` |
| bmad-create-epics-and-stories | planning | `done`, `complete` |
| bmad-create-story | planning | `done`, `complete` |
| bmad-dev-story | implementation | `done`, `ship`, `complete` |
| bmad-quick-dev | implementation | `done`, `ship` |
| bmad-code-review | implementation | `approved`, `done`, `ship` |
| bmad-qa-generate-e2e-tests | implementation | `done`, `complete` |
| bmad-editorial-review-prose | support | `approved`, `done` |
| bmad-editorial-review-structure | support | `approved`, `done` |
| bmad-review-edge-case-hunter | support | `done`, `approved` |
| bmad-quick-flow | support | `done`, `ship` |
| bmad-adversarial-general | support | `done`, `complete` |
| bmad-infrastructure-devops | support | `done`, `ship` |

**Autonomous → upgrade to Guided (staged single-pass Q&A):**

| ID | Domain | Notes |
|---|---|---|
| bmad-create-prd | planning | three staged questions: scope, audience, timeline |
| bmad-document-project | support | scope picker, depth picker, output location |
| bmad-generate-project-context | analysis | scope picker, sources picker |

**Autonomous → upgrade to Party (multi-agent collab):**

| ID | Domain | Notes |
|---|---|---|
| bmad-retrospective | support | facilitator + participant agents |
| bmad-web-orchestrator | support | orchestrator + worker agents |
| bmad-game-dev-studio | support | director + asset team |

**Out of scope — leave autonomous:**

| ID | Reason |
|---|---|
| bmad-sprint-planning | reads sprint-status.yaml; no per-round user iteration |
| bmad-sprint-status | reads sprint-status.yaml; pure file generation |
| util-file-loader | utility node, no tmux session |
| util-multi-file-loader | utility node, no tmux session |

Net work: **20 Iterative upgrades + 3 Guided upgrades + 3 Party upgrades = 26 process definitions touched**.

## Common upgrade template

To minimise churn and keep accept-token discipline consistent, a shared
helper builds the iterative upgrade. The Gate accept set is widened with
the canonical exit verbs every interactive process honours (`done`,
`abort`/`cancel` always present); domain-specific tokens are appended.

```go
// internal/bmad/interactive_defaults.go (new)

// commonAcceptTokens are the exit verbs every interactive process honours.
// Domain-specific tokens (per-process) are appended to this baseline so
// users can always escape with "done"/"abort" regardless of the process.
var commonAcceptTokens = []string{"done", "wrap up", "complete"}

var commonRejectTokens = []string{"abort", "cancel"}

// IterativeUpgradeSpec describes the per-process knobs for the standard
// "round-loop with user-confirm gate" upgrade. domainAccept is appended
// to commonAcceptTokens; iteration prompt + help text come from the
// domain so the modal copy reads naturally per process.
type IterativeUpgradeSpec struct {
    Prompt          string
    HelpText        string
    DomainAccept    []string
    MaxRounds       int            // defaults to 30 when zero
    ArtifactInputs  []InputSpec    // file-based prerequisites (optional)
    OutputSpecs     []OutputSpec
}

// applyIterativeUpgrade mutates def to satisfy the interactive standard.
// Idempotent: passing a process that already declares Mode != "" leaves
// it untouched so the four reference processes are not reshuffled.
func applyIterativeUpgrade(def *ProcessDef, spec IterativeUpgradeSpec) {
    if def.Mode != "" {
        return
    }
    def.Mode = InteractIterative
    def.EnableAstAdapter = true

    // Preserve declared artifact inputs first; iteration input goes last
    // so iterationInput() picks the round-response spec.
    inputs := append([]InputSpec(nil), spec.ArtifactInputs...)
    inputs = append(inputs, InputSpec{
        ID:        "round-response",
        Source:    InputFromUser,
        Shape:     ShapeJSON,
        Prompt:    spec.Prompt,
        HelpText:  spec.HelpText,
        MaxLength: 4000,
    })
    def.InputSpecs = inputs
    def.OutputSpecs = spec.OutputSpecs

    accept := append([]string(nil), commonAcceptTokens...)
    accept = append(accept, spec.DomainAccept...)
    maxRounds := spec.MaxRounds
    if maxRounds == 0 {
        maxRounds = 30
    }
    def.Gate = &IterationGate{
        Kind:         GateUserConfirm,
        MaxRounds:    maxRounds,
        AcceptTokens: accept,
        RejectTokens: commonRejectTokens,
    }
}
```

Per-process call site (registry.go) becomes a one-liner per upgrade:

```go
applyIterativeUpgrade(&processes[idx-of-dev-story], IterativeUpgradeSpec{
    Prompt:       "Provide feedback or type 'done' when the implementation is ready.",
    HelpText:     "Reply with feedback for the next round; 'done' to ship.",
    DomainAccept: []string{"ship", "approved"},
    OutputSpecs:  []OutputSpec{{ID: "code", Target: OutputToBoth, ArtifactName: "code"}},
})
```

The Guided and Party upgrades follow the same pattern with separate
`applyGuidedUpgrade` / `applyPartyUpgrade` helpers; Guided takes a
`StagedInputs` list, Party takes an `AgentRoster`.

## Phasing

Five phases, each shippable independently with its own commit. Each phase
ends with `go test ./internal/bmad/ ./internal/uiadapter/` + a
`wails dev` smoke check.

### Phase 1 — Helper + tests (foundation)

- New `internal/bmad/interactive_defaults.go` with the three
  `apply*Upgrade` helpers
- Table-driven test exercising idempotency: applying twice equals
  applying once; pre-set Mode is preserved
- Test covering iterationInput() returns the round-response spec, Gate
  has accept/reject tokens, EnableAstAdapter=true

No registry mutations yet — pure scaffolding so reviewers can read the
helper in isolation.

### Phase 2 — High-traffic Iterative upgrades

Six processes the user touches every day:
- bmad-dev-story
- bmad-code-review
- bmad-create-story
- bmad-validate-prd
- bmad-edit-prd
- bmad-quick-dev

Each gets its own helper-call line. New tests in `registry_test.go`
walk these IDs and assert Mode/Gate/InputSpecs are populated.

Smoke: run a story-development workflow end-to-end against
`test-bmad-proect`. Confirm:
1. Modal fires on each node's first idle
2. UI-AST renders a round-response widget (free shape via fallback)
3. Typing `done` flips the gate and advances the workflow
4. Typing feedback advances to round+1 with the answer injected

### Phase 3 — Remaining Iterative upgrades

The other 14 Iterative upgrades. Same pattern, batched by phase
(Analysis → Planning → Implementation → Support) for review locality.

### Phase 4 — Guided upgrades

- bmad-create-prd
- bmad-document-project
- bmad-generate-project-context

`applyGuidedUpgrade` walks a `StagedInputs []InputSpec` slice — each
becomes a separate `decision_group` rendered in source order (mirrors
bmad-product-brief's three-staged shape). No Gate, single pass.

### Phase 5 — Party upgrades

- bmad-retrospective
- bmad-web-orchestrator
- bmad-game-dev-studio

`applyPartyUpgrade` provides the `AgentRoster` pattern. Lower priority
because Party is more experimental and the user touches it rarely.

## Migration risks

1. **Persisted workflows on disk** — `~/.mashed/workflows/*/execution.json`
   snapshots from before the upgrade have `Status: running` /
   `Status: complete` for nodes whose process now expects an iteration
   input. The Resume path handles this safely: snapshot status takes
   precedence over registry shape (`resume.go` re-emits awaiting_input
   ONLY for nodes that were `awaiting_input` at suspend time). Verified
   by `executor_resume_test.go` round-trip tests.

2. **Templates in `templates.go`** — every template references process
   IDs by string. Mode changes do NOT break references; the templates
   only carry IDs and labels, no shape assumptions.

3. **Skill prompts** — the embedded BMAD skill markdown sometimes
   instructs claude to "ask the user for X". When a process becomes
   iterative the skill prompt and the round-response widget should
   match the same vocabulary so the modal `prompt` and claude's
   inline question agree. **Action**: read each skill markdown during
   the per-process change and align the prompt copy.

4. **Tests that pin autonomous behavior** — grep for `bmad-dev-story`
   / `bmad-code-review` etc. in `executor_test.go`; any test that
   asserts `Mode == ""` or `Gate == nil` on these IDs needs updating
   to the new shape. Inventory pass:
   ```bash
   rg -l 'bmad-(dev-story|code-review|create-story|validate-prd|edit-prd|quick-dev)' internal/bmad/*_test.go
   ```

5. **Frontend modal copy** — the round-response widget renders
   `prompt` + `helpText` directly. Each phase's test plan includes
   a Playwright smoke run hitting one node from each batch to catch
   typos / awkward copy.

6. **Round limit blast radius** — `MaxRounds: 30` is the brainstorming
   default. For dev-story / code-review a runaway loop can burn 30
   claude-cli invocations × ~30 s each = 15 minutes of wall clock.
   Consider per-process `MaxRounds` overrides (e.g. dev-story = 10,
   research = 30) so the safety ceiling matches the natural cadence.

7. **Adapter timeout amplification** — every interactive node now
   calls `adapter.Translate` on every round (autonomous nodes never
   did). 26 newly-interactive processes × N rounds × 60 s timeout is
   a lot of Haiku traffic. Pre-cropping (the prior commit `ada5b96`
   addresses this) keeps individual translations under budget; rate
   limits are the user's claude.ai quota, not ours.

## Test strategy

- **Unit**: `interactive_defaults_test.go` — table-driven idempotency,
  default propagation, accept-token merging.
- **Integration**: `registry_interactive_test.go` — walks every process
  ID, asserts `Mode != ""` for the rolled-out IDs, asserts
  `iterationInput().ID == "round-response"` where the helper was used,
  asserts Gate populated when expected.
- **End-to-end**: `executor_interactive_e2e_test.go` — drives a
  three-node workflow (file loader → dev-story → code-review) through
  one round each with a stub UI adapter, asserts both nodes emit
  `awaiting_input` and gate-satisfy on `"done"`.
- **Frontend smoke**: Playwright spec opens a workflow with each newly
  interactive node, types `done` in the modal, expects the next node to
  start.

## Open decisions

1. **Single shared helper vs per-process explicit literal?** The plan
   above uses a helper for consistency. Alternative: write each upgrade
   as a literal struct field assignment so the registry stays
   self-documenting and reviewers can read the Gate/InputSpecs inline.
   Trade-off: helper is ~30 lines saved per process but adds a hop when
   debugging.

2. **`round-response` shape: `ShapeJSON` vs `ShapeFree`?** Brainstorming
   uses `ShapeJSON` so the adapter's UI-AST shapes the widget. Setting
   `ShapeFree` would skip the adapter altogether (raw textbox). The plan
   uses `ShapeJSON` so adapter-driven processes keep the structured
   modal even when the model degrades.

3. **Phase 2 batch size?** Six processes is one PR's worth of tests. If
   the smoke run shows the AST adapter struggling under the new traffic
   (timeouts, fallbacks), Phase 2 should split into two commits of three
   each.

4. **Skill prompt alignment** — touching the embedded markdown for each
   process is ~26 file edits. Worth a separate prep PR before Phase 2 so
   the helper rollout doesn't bundle skill copy edits with registry
   shape changes.
