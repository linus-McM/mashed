// AC mapping for the Phase 2 walking integration suite:
//   AC-1 → TestStoryRollout02_AC1_ProcessesAreIterativeWithAstAdapter
//   AC-2 → TestStoryRollout02_AC2_RoundResponseIterationSlot
//   AC-3 → TestStoryRollout02_AC3_GatePopulated
//   AC-4 → TestStoryRollout02_AC4_OutputSpecsReflectArtifactMapping
//   AC-5 → TestStoryRollout02_AC5_RolloutIDsDifferFromPreRolloutGoldens
//   AC-6 → exercised package-wide via existing executor_test.go fixtures
//
// BDD scenarios coverage overlaps with AC-* tests by design — story mandates
// BDD-named entries for backlog traceability.
package bmad

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// outputTuple is a (target, artifactName) pair used for multiset comparison
// of OutputSpecs. The OutputSpec.ID column is engineer discretion (story does
// not pin it) so AC-4 only asserts target + artifactName.
type outputTuple struct {
	Target       OutputTarget
	ArtifactName string
}

// rollout02Expect carries the expected interactive shape per process.
type rollout02Expect struct {
	id              string
	maxRounds       int
	domainAccept    []string      // appended after baseline {"done","wrap up","complete"}
	expectedOutputs []outputTuple // multiset compared via assert.ElementsMatch
}

// rollout02Expectations sources from the story §"Per-process specs" table.
// Used by AC-1..AC-4 sub-tests.
var rollout02Expectations = []rollout02Expect{
	{
		id:           "bmad-dev-story",
		maxRounds:    10,
		domainAccept: []string{"ship", "approved"},
		expectedOutputs: []outputTuple{
			{Target: OutputToMemory, ArtifactName: ""}, // code (unmapped → memory)
			{Target: OutputToMemory, ArtifactName: ""}, // tests (memory)
		},
	},
	{
		id:           "bmad-code-review",
		maxRounds:    10,
		domainAccept: []string{"ship", "approved"},
		expectedOutputs: []outputTuple{
			{Target: OutputToFile, ArtifactName: "review-report"},
			{Target: OutputToMemory, ArtifactName: ""}, // code
		},
	},
	{
		id:           "bmad-create-story",
		maxRounds:    15,
		domainAccept: nil, // baseline only — BDD: "no domain accept tokens beyond baseline"
		expectedOutputs: []outputTuple{
			{Target: OutputToMemory, ArtifactName: ""}, // story-*.md (treated as memory per AC-4)
		},
	},
	{
		id:           "bmad-validate-prd",
		maxRounds:    10,
		domainAccept: []string{"pass", "approved"},
		expectedOutputs: []outputTuple{
			{Target: OutputToFile, ArtifactName: "prd-validation"},
			{Target: OutputToMemory, ArtifactName: ""}, // PRD.md (memory per AC-4)
		},
	},
	{
		id:           "bmad-edit-prd",
		maxRounds:    15,
		domainAccept: []string{"apply", "ship"},
		expectedOutputs: []outputTuple{
			{Target: OutputToFile, ArtifactName: "PRD.md"},
		},
	},
	{
		id:           "bmad-quick-dev",
		maxRounds:    10,
		domainAccept: []string{"ship"},
		expectedOutputs: []outputTuple{
			{Target: OutputToMemory, ArtifactName: ""}, // code
		},
	},
}

// expectedAcceptTokens reproduces mergeAcceptTokens semantics from
// interactive_defaults.go: baseline {"done","wrap up","complete"} first, then
// domain tokens appended in declared order with duplicates dropped.
func expectedAcceptTokens(domain []string) []string {
	out := []string{"done", "wrap up", "complete"}
	seen := map[string]struct{}{
		"done": {}, "wrap up": {}, "complete": {},
	}
	for _, t := range domain {
		if _, dup := seen[t]; dup {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

// outputContains is a focused predicate used by the BDD scenarios that assert
// presence of a single (target, artifactName) tuple.
func outputContains(specs []OutputSpec, target OutputTarget, artifactName string) bool {
	for _, s := range specs {
		if s.Target == target && s.ArtifactName == artifactName {
			return true
		}
	}
	return false
}

// outputTuples extracts (target, artifactName) pairs for ElementsMatch.
func outputTuples(specs []OutputSpec) []outputTuple {
	out := make([]outputTuple, len(specs))
	for i, s := range specs {
		out[i] = outputTuple{Target: s.Target, ArtifactName: s.ArtifactName}
	}
	return out
}

// ─── AC-1 ─────────────────────────────────────────────────────────────────────
// Mode == InteractIterative AND EnableAstAdapter == true for every rollout ID.

// TestStoryRollout02_AC1_ProcessesAreIterativeWithAstAdapter exercises AC-1.
// BDD anchor: "dev-story is iterative with 10-round cap and ship token" (Mode +
// adapter half).
func TestStoryRollout02_AC1_ProcessesAreIterativeWithAstAdapter(t *testing.T) {
	for _, id := range phase2RolloutIDs {
		id := id
		t.Run(id, func(t *testing.T) {
			p, ok := ProcessByID(id)
			require.True(t, ok, "process %q must exist in registry", id)

			assert.Equal(t, InteractIterative, p.Mode,
				"AC-1: %q must be InteractIterative after registry_interactive_phase2.go init() runs (currently Mode=%q — RED expected)",
				id, p.Mode)
			assert.True(t, p.EnableAstAdapter,
				"AC-1: %q must opt into the AST adapter (EnableAstAdapter=true)", id)
		})
	}
}

// ─── AC-2 ─────────────────────────────────────────────────────────────────────
// iterationInput() returns (spec, true) with the round-response slot.

// TestStoryRollout02_AC2_RoundResponseIterationSlot exercises AC-2.
// BDD anchor: "Round-response is ShapeJSON".
func TestStoryRollout02_AC2_RoundResponseIterationSlot(t *testing.T) {
	for _, id := range phase2RolloutIDs {
		id := id
		t.Run(id, func(t *testing.T) {
			p, ok := ProcessByID(id)
			require.True(t, ok, "process %q must exist in registry", id)

			spec, found := p.iterationInput()
			require.True(t, found,
				"AC-2: iterationInput() must return a recurring slot for %q", id)
			assert.Equal(t, RoundResponseInputID, spec.ID,
				"AC-2: iteration slot ID must be %q", RoundResponseInputID)
			assert.Equal(t, ShapeJSON, spec.Shape,
				"AC-2 / Open Decision #2: iteration slot Shape must be ShapeJSON for %q", id)
			assert.Equal(t, InputFromUser, spec.Source,
				"AC-2: iteration slot Source must be InputFromUser")
			assert.False(t, spec.Required,
				"AC-2: iteration slot Required must be false (so iterationInput() resolves to it per §5.2)")
			assert.NotEmpty(t, spec.Prompt,
				"AC-2: iteration slot Prompt must be non-empty (story per-process specs table)")
		})
	}
}

// ─── AC-3 ─────────────────────────────────────────────────────────────────────
// Gate populated with GateUserConfirm + per-process MaxRounds + baseline +
// domain accept tokens + reject tokens {"abort","cancel"}.

// TestStoryRollout02_AC3_GatePopulated exercises AC-3.
// BDD anchors: "validate-prd accepts 'pass' and 'approved'", "quick-dev caps
// at 10 rounds", "create-story has no domain accept tokens beyond baseline".
func TestStoryRollout02_AC3_GatePopulated(t *testing.T) {
	for _, e := range rollout02Expectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)

			require.NotNil(t, p.Gate,
				"AC-3: Gate must be populated for %q", e.id)
			assert.Equal(t, GateUserConfirm, p.Gate.Kind,
				"AC-3: Gate.Kind must be GateUserConfirm")
			assert.Equal(t, e.maxRounds, p.Gate.MaxRounds,
				"AC-3: Gate.MaxRounds must equal the per-process override (story §Per-process specs)")
			assert.Equal(t, expectedAcceptTokens(e.domainAccept), p.Gate.AcceptTokens,
				"AC-3: AcceptTokens must equal baseline {'done','wrap up','complete'} + domain tokens (mergeAcceptTokens order, dedup preserved)")
			assert.Equal(t, []string{"abort", "cancel"}, p.Gate.RejectTokens,
				"AC-3: RejectTokens must equal common baseline {'abort','cancel'}")
		})
	}
}

// ─── AC-4 ─────────────────────────────────────────────────────────────────────
// OutputSpecs declare per-process artifact mapping (target, artifactName) as a
// multiset. ID is engineer discretion — not pinned.

// TestStoryRollout02_AC4_OutputSpecsReflectArtifactMapping exercises AC-4.
// BDD anchors: "code-review carries review-report file output", "edit-prd
// writes PRD.md as file output".
func TestStoryRollout02_AC4_OutputSpecsReflectArtifactMapping(t *testing.T) {
	for _, e := range rollout02Expectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)

			require.Len(t, p.OutputSpecs, len(e.expectedOutputs),
				"AC-4: %q must declare exactly %d OutputSpecs (got %d)",
				e.id, len(e.expectedOutputs), len(p.OutputSpecs))

			assert.ElementsMatch(t, e.expectedOutputs, outputTuples(p.OutputSpecs),
				"AC-4: OutputSpecs (target, artifactName) multiset for %q must match story per-process specs table", e.id)
		})
	}
}

// ─── AC-5 ─────────────────────────────────────────────────────────────────────
// Phase 2 upgrades alter the JSON shape of the six rollout IDs, so each must
// DIFFER from its committed pre-rollout golden at testdata/registry/<id>.json.
//
// Why this is the right RED indicator: the existing
// TestU0_AC3_NonMigratedProcessesUnchanged test (registry_test.go) asserts
// byte-equality for every non-migrated ID. Once the GREEN engineer applies
// the upgrades, the six entries drift from their goldens — TestU0_AC3 starts
// failing on them. To restore green the engineer MUST extend the skip
// predicate (per story §Risks #4: add `phase2RolloutIDs` slice and
// `slices.Contains(phase2RolloutIDs, p.ID)` to the skip clause). Both halves
// of AC-5 are therefore exercised by the combination of:
//   1. (this test, must-differ) — proves upgrades ran
//   2. (TestU0_AC3, must-equal-or-skip) — proves skip-list extended
//
// In RED, the registry entry equals the golden byte-for-byte → this test fails.
// In GREEN, the entry differs → this test passes; companion test asserts skip.

// TestStoryRollout02_AC5_RolloutIDsDifferFromPreRolloutGoldens exercises AC-5.
// BDD anchor: "U0 byte-equality holds for non-rolled-out processes" (the
// negative half: rolled-out processes MUST drift).
func TestStoryRollout02_AC5_RolloutIDsDifferFromPreRolloutGoldens(t *testing.T) {
	for _, id := range phase2RolloutIDs {
		id := id
		t.Run(id, func(t *testing.T) {
			p, ok := ProcessByID(id)
			require.True(t, ok, "process %q must exist in registry", id)

			got, err := json.Marshal(p)
			require.NoError(t, err, "marshal %q", id)

			goldenPath := filepath.Join("testdata", "registry", id+".json")
			raw, err := os.ReadFile(goldenPath)
			require.NoError(t, err,
				"AC-5: golden %q must exist (pre-rollout snapshot committed by U0)", goldenPath)
			want := bytes.TrimRight(raw, "\n")

			assert.Falsef(t, bytes.Equal(got, want),
				"AC-5: %q must differ from pre-rollout golden after registry_interactive_phase2.go init() runs — currently identical, meaning the helper-call upgrade did not mutate the entry. Engineer next step: extend phase2RolloutIDs in registry_test.go and add it to the skip predicate so TestU0_AC3 stays green.",
				id)
		})
	}
}

// TestStoryRollout02_AC5_Phase2RolloutIDsDeclaredInRegistryTest asserts that
// registry_test.go contains a `phase2RolloutIDs` slice declaration listing all
// six rollout IDs. This is a textual smoke test (not a compile-time reference)
// so it produces a clean assertion failure in RED rather than a compile error
// when the slice is missing.
//
// In RED: registry_test.go does not declare `phase2RolloutIDs` → assertion
// fails with a clear message naming the missing identifier.
// In GREEN: the engineer has added `phase2RolloutIDs = []string{...}` and
// extended the skip predicate → assertion passes.
//
// Companion: TestU0_AC3_NonMigratedProcessesUnchanged (registry_test.go) will
// only stay green if phase2RolloutIDs is also referenced by the skip clause —
// engineer cannot pass this test by adding the declaration alone.
func TestStoryRollout02_AC5_Phase2RolloutIDsDeclaredInRegistryTest(t *testing.T) {
	raw, err := os.ReadFile("registry_test.go")
	require.NoError(t, err, "registry_test.go must be readable from the package directory")
	src := string(raw)

	// Use strings.Contains + assert.Truef so failure messages stay concise —
	// assert.Contains dumps the entire haystack on failure, which is 100KB+ of
	// noise the engineer doesn't need.
	if !assert.Truef(t, strings.Contains(src, "phase2RolloutIDs"),
		"AC-5: registry_test.go must declare a `phase2RolloutIDs` slice (story §Risks #4: 'extend u0MigratedProcessIDs to include rollout phase 2 via a separate slice phase2RolloutIDs')") {
		return // downstream block check is meaningless without the declaration
	}

	// Tight check: extract the slice literal block and verify each rollout ID
	// is INSIDE it. Loose file-wide containment would pass for IDs that appear
	// elsewhere in registry_test.go (e.g. `bmad-dev-story` in a JSON round-trip
	// fixture).
	block, ok := extractPhase2RolloutIDsBlock(src)
	require.Truef(t, ok,
		"AC-5: phase2RolloutIDs declaration must use the `[]string{...}` slice-literal form (story §Risks #4 example)")

	for _, id := range phase2RolloutIDs {
		assert.Truef(t, strings.Contains(block, "\""+id+"\""),
			"AC-5: phase2RolloutIDs slice must contain %q", id)
	}

	// Engineer must wire the slice into the U0 skip union so TestU0_AC3 stays
	// green. The union now lives in `skippedFromU0Goldens` (composed from
	// u0MigratedProcessIDs + phase2RolloutIDs); subsequent rollout stories
	// append their own slices into the same union.
	assert.Truef(t, strings.Contains(src, "skippedFromU0Goldens"),
		"AC-5: registry_test.go must declare a `skippedFromU0Goldens` union slice that includes phase2RolloutIDs so TestU0_AC3_NonMigratedProcessesUnchanged stays green for the six upgraded IDs")
	assert.Truef(t, strings.Contains(src, "phase2RolloutIDs..."),
		"AC-5: skippedFromU0Goldens must spread phase2RolloutIDs into its union (look for `phase2RolloutIDs...` token)")
}

// extractPhase2RolloutIDsBlock returns the slice literal body for
// `phase2RolloutIDs = []string{ … }` — the text between the opening `{` and
// matching closing `}`. Brace nesting is tracked naively (no string-literal
// awareness) which is sufficient for a slice of string literals.
func extractPhase2RolloutIDsBlock(src string) (string, bool) {
	const sentinel = "phase2RolloutIDs = []string{"
	start := strings.Index(src, sentinel)
	if start < 0 {
		return "", false
	}
	open := start + len(sentinel)
	depth := 1
	for i := open; i < len(src); i++ {
		switch src[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return src[open:i], true
			}
		}
	}
	return "", false
}

// ─── BDD Scenarios ────────────────────────────────────────────────────────────
// One test per BDD scenario from the story §"BDD Test Scenarios" so the lead
// can map scenario → test name without ambiguity.

// BDD Scenario 1: "dev-story is iterative with 10-round cap and ship token".
func TestStoryRollout02_BDD_DevStoryIterativeWith10RoundCapAndShipToken(t *testing.T) {
	p, ok := ProcessByID("bmad-dev-story")
	require.True(t, ok)
	assert.Equal(t, InteractIterative, p.Mode, "Mode must be iterative")
	assert.True(t, p.EnableAstAdapter, "EnableAstAdapter must be true")
	require.NotNil(t, p.Gate, "Gate must be populated")
	assert.Equal(t, 10, p.Gate.MaxRounds, "Gate.MaxRounds must be 10")
	for _, want := range []string{"done", "wrap up", "complete", "ship", "approved"} {
		assert.Containsf(t, p.Gate.AcceptTokens, want,
			"AcceptTokens must contain %q", want)
	}
	spec, found := p.iterationInput()
	require.True(t, found, "iterationInput() must return a spec")
	assert.Equal(t, RoundResponseInputID, spec.ID,
		"iteration slot ID must be 'round-response'")
}

// BDD Scenario 2: "code-review carries review-report file output".
func TestStoryRollout02_BDD_CodeReviewCarriesReviewReportFileOutput(t *testing.T) {
	p, ok := ProcessByID("bmad-code-review")
	require.True(t, ok)
	assert.Truef(t,
		outputContains(p.OutputSpecs, OutputToFile, "review-report"),
		"code-review OutputSpecs must contain (target=file, artifactName=review-report). Got: %+v",
		p.OutputSpecs)
}

// BDD Scenario 3: "validate-prd accepts 'pass' and 'approved' tokens".
func TestStoryRollout02_BDD_ValidatePrdAcceptsPassAndApproved(t *testing.T) {
	p, ok := ProcessByID("bmad-validate-prd")
	require.True(t, ok)
	require.NotNil(t, p.Gate, "Gate must be populated")
	assert.Contains(t, p.Gate.AcceptTokens, "pass", "AcceptTokens must contain 'pass'")
	assert.Contains(t, p.Gate.AcceptTokens, "approved", "AcceptTokens must contain 'approved'")
}

// BDD Scenario 4: "edit-prd writes PRD.md as file output".
func TestStoryRollout02_BDD_EditPrdWritesPrdMdAsFileOutput(t *testing.T) {
	p, ok := ProcessByID("bmad-edit-prd")
	require.True(t, ok)
	assert.Truef(t,
		outputContains(p.OutputSpecs, OutputToFile, "PRD.md"),
		"edit-prd OutputSpecs must contain (target=file, artifactName=PRD.md). Got: %+v",
		p.OutputSpecs)
}

// BDD Scenario 5: "quick-dev caps at 10 rounds".
func TestStoryRollout02_BDD_QuickDevCapsAt10Rounds(t *testing.T) {
	p, ok := ProcessByID("bmad-quick-dev")
	require.True(t, ok)
	require.NotNil(t, p.Gate, "Gate must be populated")
	assert.Equal(t, 10, p.Gate.MaxRounds, "Gate.MaxRounds must be 10")
}

// BDD Scenario 6: "create-story has no domain accept tokens beyond baseline".
func TestStoryRollout02_BDD_CreateStoryHasOnlyBaselineAcceptTokens(t *testing.T) {
	p, ok := ProcessByID("bmad-create-story")
	require.True(t, ok)
	require.NotNil(t, p.Gate, "Gate must be populated")
	assert.Equal(t, []string{"done", "wrap up", "complete"}, p.Gate.AcceptTokens,
		"create-story AcceptTokens must equal baseline exactly (no domain extension)")
}

// BDD Scenario 7: "U0 byte-equality holds for non-rolled-out processes".
//
// Re-asserts the contract for the 24 entries that are NEITHER U0-migrated NOR
// phase-2 rollout. Sibling to TestU0_AC3_NonMigratedProcessesUnchanged
// (registry_test.go); declared here so the BDD scenario has a dedicated home
// in the phase 2 file.
//
// In RED state: passes (registry unchanged for those 24).
// In GREEN state: passes (engineer extends skip predicate; this test mirrors).
func TestStoryRollout02_BDD_U0ByteEqualityHoldsForNonRolledOutProcesses(t *testing.T) {
	for _, p := range AllProcesses() {
		if slices.Contains(skippedFromU0Goldens, p.ID) {
			continue
		}
		p := p
		t.Run(p.ID, func(t *testing.T) {
			got, err := json.Marshal(p)
			require.NoError(t, err)

			goldenPath := filepath.Join("testdata", "registry", p.ID+".json")
			raw, err := os.ReadFile(goldenPath)
			require.NoErrorf(t, err, "golden %q must exist", goldenPath)
			want := bytes.TrimRight(raw, "\n")

			assert.Truef(t, bytes.Equal(got, want),
				"BDD: non-rolled-out process %q must remain byte-identical to %s", p.ID, goldenPath)
		})
	}
}

// BDD Scenario 8: "Persisted execution.json from before rollout still resumes".
//
// The integration round-trip lives in executor_resume_test.go (per story
// §Risks #1: "Verified by existing executor_resume_test.go round-trip"). This
// test pins the high-level invariant inline so the BDD scenario maps to a
// dedicated test name in the phase 2 file: a node persisted with NodeRunning
// status survives a JSON round-trip with that status intact, so resume.go's
// "snapshot status precedence" branch (lines 347-353) replays running rather
// than re-emitting awaiting_input.
//
// The full executor-driven assertion (no awaiting_input event emitted on
// resume) belongs in the executor harness; duplicating that scaffolding here
// would obscure the BDD intent.
func TestStoryRollout02_BDD_PersistedExecutionStillResumes(t *testing.T) {
	exec := WorkflowExecution{
		ID:         "rollout02-bdd-resume",
		WorkflowID: "wf-bdd",
		RepoPath:   t.TempDir(),
		Status:     ExecRunning,
		Nodes: []WorkflowNode{
			{
				ID:        "node-1",
				ProcessID: "bmad-dev-story",
				Status:    NodeRunning, // pre-rollout snapshot status
			},
		},
	}

	data, err := json.Marshal(exec)
	require.NoError(t, err)

	var got WorkflowExecution
	require.NoError(t, json.Unmarshal(data, &got))

	require.Len(t, got.Nodes, 1, "snapshot must round-trip exactly one node")
	assert.Equal(t, NodeRunning, got.Nodes[0].Status,
		"BDD: NodeRunning status survives JSON round-trip — resume.go §347-353 replays running rather than re-emitting awaiting_input")
	assert.Equal(t, "bmad-dev-story", got.Nodes[0].ProcessID,
		"BDD: ProcessID stable across rollout (registry shape change does not rewrite snapshot)")
}

// BDD Scenario 9: "Round-response is ShapeJSON".
//
// Asserts the round-response slot on every rollout ID carries Shape ==
// ShapeJSON (Open Decision #2 resolution). Sibling to AC-2 — declared
// separately so the BDD scenario maps to a dedicated test name.
func TestStoryRollout02_BDD_RoundResponseIsShapeJSON(t *testing.T) {
	for _, id := range phase2RolloutIDs {
		id := id
		t.Run(id, func(t *testing.T) {
			p, ok := ProcessByID(id)
			require.True(t, ok)
			spec, found := p.iterationInput()
			require.True(t, found,
				"iterationInput() must return a spec for %q", id)
			assert.Equal(t, ShapeJSON, spec.Shape,
				"BDD: round-response Shape must be ShapeJSON for %q (Open Decision #2)", id)
		})
	}
}
