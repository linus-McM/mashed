package bmad

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ----------------------------------------------------------------------------
// Shared fixtures
// ----------------------------------------------------------------------------

// freshAutonomousDef returns a minimal ProcessDef with Mode=="" so the
// helpers' upgrade path runs (the idempotency guard `if def.Mode != ""` does
// not short-circuit).
func freshAutonomousDef() ProcessDef {
	return ProcessDef{
		ID:          "test-fresh",
		Name:        "Test Fresh",
		Phase:       PhaseAnalysis,
		AgentRole:   RoleAnalyst,
		SkillName:   "test-fresh",
		Description: "Fresh autonomous fixture for upgrade-helper tests.",
		Inputs:      []string{},
		Outputs:     []string{},
		ModuleID:    "core",
		Version:     "1.0.0",
	}
}

// defaultIterativeSpec returns a spec with a non-empty Prompt (required so the
// resulting round-response slot satisfies iterationInput()) and no domain
// extras / overrides. Tests that exercise specific fields override after the
// call (or use a struct literal directly).
func defaultIterativeSpec() IterativeUpgradeSpec {
	return IterativeUpgradeSpec{
		Prompt:   "Continue iterating, or type 'done' to wrap up.",
		HelpText: "Type 'done' to finish.",
	}
}

// ----------------------------------------------------------------------------
// AC-2: applyIterativeUpgrade is idempotent.
// BDD scenarios covered:
//   - "Iterative upgrade applied to autonomous process"
//   - "Iterative upgrade is idempotent"
//   - "Iterative upgrade preserves declared artifact inputs"
// ----------------------------------------------------------------------------

func TestStoryRollout01_AC2_IterativeIdempotent(t *testing.T) {
	tests := []struct {
		name string
		spec IterativeUpgradeSpec
	}{
		{
			name: "minimal_spec",
			spec: defaultIterativeSpec(),
		},
		{
			name: "spec_with_domain_accept_and_artifact_inputs",
			spec: IterativeUpgradeSpec{
				Prompt:       "Continue or 'done'.",
				HelpText:     "Type 'done' to wrap up.",
				DomainAccept: []string{"ship"},
				MaxRounds:    15,
				ArtifactInputs: []InputSpec{
					{ID: "prd", Source: InputFromFile, ArtifactName: "PRD.md", Required: true},
				},
				OutputSpecs: []OutputSpec{
					{ID: "result", Target: OutputToFile, ArtifactName: "result"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := freshAutonomousDef()

			// First upgrade — should mutate the autonomous def into the
			// iterative shape per AC-2.
			applyIterativeUpgrade(&def, tt.spec)

			require.Equal(t, InteractIterative, def.Mode,
				"AC-2: Mode must be set to InteractIterative")
			require.True(t, def.EnableAstAdapter,
				"AC-2: EnableAstAdapter must be true after upgrade")
			require.NotEmpty(t, def.InputSpecs,
				"AC-2: InputSpecs must be populated with at least the round-response slot")

			last := def.InputSpecs[len(def.InputSpecs)-1]
			assert.Equal(t, RoundResponseInputID, last.ID,
				"AC-2: last InputSpec must be the round-response slot")
			assert.Equal(t, InputFromUser, last.Source,
				"AC-2: round-response Source must be InputFromUser")
			assert.Equal(t, ShapeJSON, last.Shape,
				"AC-2: round-response Shape must be ShapeJSON (Open Decision #2)")
			assert.False(t, last.Required,
				"AC-2: round-response Required must be false")
			assert.Equal(t, "", last.Default,
				"AC-2: round-response Default must be empty string")
			assert.NotEmpty(t, last.Prompt,
				"AC-2: round-response Prompt must be non-empty so iterationInput() can discover it")

			// Snapshot post-first-call state via JSON so a subsequent in-place
			// mutation by the second call still surfaces (slice elements share
			// backing arrays otherwise).
			afterFirst, err := json.Marshal(def)
			require.NoError(t, err)

			// Second upgrade with the same spec — must be idempotent.
			applyIterativeUpgrade(&def, tt.spec)

			afterSecond, err := json.Marshal(def)
			require.NoError(t, err)

			require.JSONEq(t, string(afterFirst), string(afterSecond),
				"AC-2 idempotency (BDD: 'Iterative upgrade is idempotent'): second call must leave def deeply equal to the first-call result")
		})
	}
}

func TestStoryRollout01_AC2_PreservesArtifactInputsBeforeRoundResponse(t *testing.T) {
	// BDD scenario: "Iterative upgrade preserves declared artifact inputs"
	def := freshAutonomousDef()
	spec := IterativeUpgradeSpec{
		Prompt: "Continue iterating.",
		ArtifactInputs: []InputSpec{
			{ID: "prd", Source: InputFromFile, ArtifactName: "PRD.md", Required: true},
		},
	}

	applyIterativeUpgrade(&def, spec)

	require.Len(t, def.InputSpecs, 2,
		"AC-2/BDD: ArtifactInputs ++ {round-response} == 2 slots")
	assert.Equal(t, "prd", def.InputSpecs[0].ID,
		"AC-2/BDD: artifact input must precede round-response")
	assert.Equal(t, RoundResponseInputID, def.InputSpecs[1].ID,
		"AC-2/BDD: round-response must be last so iterationInput() resolves to it")
}

// ----------------------------------------------------------------------------
// AC-3: applyIterativeUpgrade respects pre-set Mode.
// BDD scenario: "Iterative upgrade preserves already-interactive process"
//
// The brainstorming reference is cloned from registry.go's actual entry via
// ProcessByID so a future registry edit propagates here automatically rather
// than silently desyncing a hand-rolled fixture.
// ----------------------------------------------------------------------------

func TestStoryRollout01_AC3_PreSetModePreservedForBrainstorming(t *testing.T) {
	def, ok := ProcessByID("bmad-brainstorming")
	require.True(t, ok, "registry must contain bmad-brainstorming")
	require.Equal(t, InteractIterative, def.Mode,
		"sanity: brainstorming registry entry must already be Iterative — if this fails, the rollout precondition is broken")

	// Snapshot via JSON BEFORE the call so we detect in-place mutation of the
	// shared slice/pointer backing storage. JSONEq below compares the parsed
	// JSON values, which is the "byte-for-byte unchanged" check the story spec
	// (AC-3) calls out.
	preJSON, err := json.Marshal(def)
	require.NoError(t, err)

	// Any spec — the helper must short-circuit on `def.Mode != ""`.
	spec := IterativeUpgradeSpec{
		Prompt:       "ANYTHING — this prompt MUST NOT be applied.",
		HelpText:     "Helper must short-circuit.",
		DomainAccept: []string{"ship", "approved"},
		MaxRounds:    999,
		ArtifactInputs: []InputSpec{
			{ID: "should-not-appear", Source: InputFromFile, ArtifactName: "ghost"},
		},
		OutputSpecs: []OutputSpec{
			{ID: "should-not-appear", Target: OutputToFile, ArtifactName: "ghost"},
		},
	}

	applyIterativeUpgrade(&def, spec)

	postJSON, err := json.Marshal(def)
	require.NoError(t, err)

	require.JSONEq(t, string(preJSON), string(postJSON),
		"AC-3: brainstorming def must be byte-for-byte unchanged when Mode is already set")

	// Defence-in-depth: re-fetch a fresh copy from the registry and DeepEqual
	// against our (untouched) def. Catches the case where the helper somehow
	// mutated def in a way that round-trips through JSON identically.
	freshFromRegistry, _ := ProcessByID("bmad-brainstorming")
	require.True(t, reflect.DeepEqual(freshFromRegistry, def),
		"AC-3: def must remain reflect.DeepEqual to the registry entry after the no-op upgrade call")
}

// ----------------------------------------------------------------------------
// AC-4: Gate carries baseline + domain accept tokens, common reject tokens,
// default cap.
// BDD scenarios covered:
//   - "Domain accept tokens append after baseline"
//   - "MaxRounds default applied when zero"
//   - "MaxRounds override honoured"
// ----------------------------------------------------------------------------

func TestStoryRollout01_AC4_GateAcceptTokenMergeAndDefaults(t *testing.T) {
	tests := []struct {
		name              string
		domainAccept      []string
		maxRounds         int
		wantAcceptTokens  []string
		wantMaxRounds     int
		dedupExplain      string
	}{
		{
			name:             "default_max_rounds_no_domain_accept",
			domainAccept:     nil,
			maxRounds:        0,
			wantAcceptTokens: []string{"done", "wrap up", "complete"},
			wantMaxRounds:    30,
			dedupExplain:     "no domain tokens — accept slice equals commonAcceptTokens",
		},
		{
			name:             "domain_appended_after_baseline",
			domainAccept:     []string{"ship", "approved"},
			maxRounds:        0,
			wantAcceptTokens: []string{"done", "wrap up", "complete", "ship", "approved"},
			wantMaxRounds:    30,
			dedupExplain:     "BDD: 'Domain accept tokens append after baseline' — baseline first, domain second, in declared order",
		},
		{
			name:             "single_domain_token_ship",
			domainAccept:     []string{"ship"},
			maxRounds:        0,
			wantAcceptTokens: []string{"done", "wrap up", "complete", "ship"},
			wantMaxRounds:    30,
			dedupExplain:     "AC-4 happy path",
		},
		{
			name:             "domain_dedup_against_baseline",
			domainAccept:     []string{"done", "ship"},
			maxRounds:        0,
			wantAcceptTokens: []string{"done", "wrap up", "complete", "ship"},
			wantMaxRounds:    30,
			dedupExplain:     "BDD: 'no token appears twice' — duplicates against the baseline are dropped",
		},
		{
			name:             "domain_internal_dedup",
			domainAccept:     []string{"ship", "ship", "approved"},
			maxRounds:        0,
			wantAcceptTokens: []string{"done", "wrap up", "complete", "ship", "approved"},
			wantMaxRounds:    30,
			dedupExplain:     "BDD: 'no token appears twice' — duplicates within the domain slice are dropped",
		},
		{
			name:             "max_rounds_override_honoured",
			domainAccept:     nil,
			maxRounds:        10,
			wantAcceptTokens: []string{"done", "wrap up", "complete"},
			wantMaxRounds:    10,
			dedupExplain:     "BDD: 'MaxRounds override honoured' — non-zero spec value wins over default 30",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := freshAutonomousDef()
			spec := IterativeUpgradeSpec{
				Prompt:       "Iterate then 'done'.",
				DomainAccept: tt.domainAccept,
				MaxRounds:    tt.maxRounds,
			}

			applyIterativeUpgrade(&def, spec)

			require.NotNil(t, def.Gate, "AC-4: Gate must be populated by applyIterativeUpgrade")
			assert.Equal(t, GateUserConfirm, def.Gate.Kind,
				"AC-4: Gate.Kind must be GateUserConfirm for iterative upgrades")
			assert.Equal(t, tt.wantAcceptTokens, def.Gate.AcceptTokens,
				"AC-4: AcceptTokens must equal expected baseline+domain merge — %s", tt.dedupExplain)
			assert.Equal(t, commonRejectTokens, def.Gate.RejectTokens,
				"AC-4: RejectTokens must equal the shared commonRejectTokens var (iterative path)")
			assert.Equal(t, tt.wantMaxRounds, def.Gate.MaxRounds,
				"AC-4: MaxRounds default branch — %s", tt.dedupExplain)
		})
	}
}

// ----------------------------------------------------------------------------
// AC-5: iterationInput() returns the round-response slot after upgrade.
// (No standalone BDD scenario — implicit in AC-2/AC-7 scenarios; this test
//  isolates the predicate so the lead can map AC-5 to a single test name.)
// ----------------------------------------------------------------------------

func TestStoryRollout01_AC5_IterationInputAfterIterativeUpgrade(t *testing.T) {
	def := freshAutonomousDef()
	spec := IterativeUpgradeSpec{
		Prompt: "Continue or 'done'.",
		ArtifactInputs: []InputSpec{
			// Required + Default-having + has-Shape: must NOT match the
			// iteration predicate so we prove the helper's last-slot ordering
			// (not coincidental matching) is what makes iterationInput()
			// resolve to round-response.
			{ID: "scope", Source: InputFromUser, Shape: ShapeFree, Required: true, Prompt: "Scope?"},
		},
	}

	applyIterativeUpgrade(&def, spec)

	got, ok := def.iterationInput()
	require.True(t, ok,
		"AC-5: iterationInput() must return (slot, true) after applyIterativeUpgrade")
	assert.Equal(t, RoundResponseInputID, got.ID,
		"AC-5: iterationInput() must resolve to the round-response slot — not the artifact input")
}

// ----------------------------------------------------------------------------
// AC-6: applyGuidedUpgrade lays staged inputs in declared order with no
// iteration slot.
// BDD scenario: "Guided upgrade lays staged inputs and skips iteration slot"
// ----------------------------------------------------------------------------

func TestStoryRollout01_AC6_GuidedUpgrade(t *testing.T) {
	stagedA := InputSpec{ID: "stage-a", Source: InputFromUser, Shape: ShapeFree, Required: true, Prompt: "Stage A?"}
	stagedB := InputSpec{ID: "stage-b", Source: InputFromUser, Shape: ShapeChoice, Required: true, Prompt: "Stage B?", Options: []string{"x", "y"}}
	stagedC := InputSpec{ID: "stage-c", Source: InputFromFile, ArtifactName: "stage-c-input"}
	finalApproval := InputSpec{ID: "final-approval", Source: InputFromUser, Shape: ShapeApproval, Required: true, Prompt: "Approve?"}

	tests := []struct {
		name           string
		spec           GuidedUpgradeSpec
		wantInputIDs   []string
		wantMaxRounds  int
		wantAcceptTok  []string
	}{
		{
			name: "three_staged_inputs_no_final_approval_no_overrides",
			spec: GuidedUpgradeSpec{
				StagedInputs: []InputSpec{stagedA, stagedB, stagedC},
			},
			wantInputIDs:  []string{"stage-a", "stage-b", "stage-c"},
			wantMaxRounds: 10,
			wantAcceptTok: []string{"yes"},
		},
		{
			name: "three_staged_inputs_with_final_approval",
			spec: GuidedUpgradeSpec{
				StagedInputs:  []InputSpec{stagedA, stagedB, stagedC},
				FinalApproval: &finalApproval,
			},
			wantInputIDs:  []string{"stage-a", "stage-b", "stage-c", "final-approval"},
			wantMaxRounds: 10,
			wantAcceptTok: []string{"yes"},
		},
		{
			name: "accept_tokens_override_honoured",
			spec: GuidedUpgradeSpec{
				StagedInputs: []InputSpec{stagedA},
				AcceptTokens: []string{"approve", "ship"},
				MaxRounds:    7,
			},
			wantInputIDs:  []string{"stage-a"},
			wantMaxRounds: 7,
			wantAcceptTok: []string{"approve", "ship"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := freshAutonomousDef()

			applyGuidedUpgrade(&def, tt.spec)

			require.Equal(t, InteractGuided, def.Mode,
				"AC-6: Mode must be InteractGuided")
			require.True(t, def.EnableAstAdapter,
				"AC-6: EnableAstAdapter must be true")

			// InputSpecs equals StagedInputs in declared order, plus
			// FinalApproval if non-nil. Compare IDs in order — if the helper
			// inserts a recurring slot, the length check below catches it.
			require.Len(t, def.InputSpecs, len(tt.wantInputIDs),
				"AC-6: InputSpecs length must equal StagedInputs (+ FinalApproval if non-nil)")
			gotIDs := make([]string, len(def.InputSpecs))
			for i, s := range def.InputSpecs {
				gotIDs[i] = s.ID
			}
			assert.Equal(t, tt.wantInputIDs, gotIDs,
				"AC-6: InputSpecs must appear in declared order")

			// No iteration slot — guided is single-pass.
			_, hasIter := def.iterationInput()
			assert.False(t, hasIter,
				"AC-6 / BDD: 'iterationInput() returns false' — guided upgrade must not add a recurring slot")

			require.NotNil(t, def.Gate, "AC-6: Gate must be populated")
			assert.Equal(t, tt.wantMaxRounds, def.Gate.MaxRounds,
				"AC-6: MaxRounds default 10 / spec override")
			assert.Equal(t, tt.wantAcceptTok, def.Gate.AcceptTokens,
				"AC-6: AcceptTokens default {'yes'} / spec override")
		})
	}
}

// ----------------------------------------------------------------------------
// AC-7: applyPartyUpgrade assembles topic + recurring message with party
// defaults.
// BDD scenario: "Party upgrade attaches topic and message slots"
// ----------------------------------------------------------------------------

func TestStoryRollout01_AC7_PartyUpgrade(t *testing.T) {
	topic := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: true, // Required==true so Topic does NOT match iterationInput().
		Prompt:   "What's the discussion topic?",
	}

	tests := []struct {
		name             string
		spec             PartyUpgradeSpec
		wantMaxRounds    int
		wantAcceptTokens []string
	}{
		{
			name: "default_party_spec",
			spec: PartyUpgradeSpec{
				Topic:       topic,
				RoundPrompt: "Your turn...",
				HelpText:    "Type 'exit' when done.",
			},
			wantMaxRounds:    100,
			wantAcceptTokens: []string{"exit", "done", "wrap up"},
		},
		{
			name: "max_rounds_and_accept_tokens_overrides",
			spec: PartyUpgradeSpec{
				Topic:        topic,
				RoundPrompt:  "Speak.",
				MaxRounds:    50,
				AcceptTokens: []string{"bye", "fin"},
			},
			wantMaxRounds:    50,
			wantAcceptTokens: []string{"bye", "fin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def := freshAutonomousDef()

			applyPartyUpgrade(&def, tt.spec)

			require.Equal(t, InteractParty, def.Mode,
				"AC-7: Mode must be InteractParty")
			require.True(t, def.EnableAstAdapter,
				"AC-7: EnableAstAdapter must be true")

			require.Len(t, def.InputSpecs, 2,
				"AC-7: InputSpecs must be exactly [Topic, message]")
			assert.Equal(t, tt.spec.Topic, def.InputSpecs[0],
				"AC-7: first slot must be the spec.Topic verbatim")

			msg := def.InputSpecs[1]
			assert.Equal(t, PartyMessageInputID, msg.ID,
				"AC-7: second slot must have ID 'message'")
			assert.Equal(t, InputFromUser, msg.Source,
				"AC-7: message Source must be InputFromUser")
			assert.Equal(t, ShapeJSON, msg.Shape,
				"AC-7: message Shape must be ShapeJSON")
			assert.False(t, msg.Required,
				"AC-7: message Required must be false")
			assert.Equal(t, tt.spec.RoundPrompt, msg.Prompt,
				"AC-7: message Prompt must equal spec.RoundPrompt")

			// iterationInput() must resolve to the message slot, not Topic.
			got, ok := def.iterationInput()
			require.True(t, ok, "AC-7: iterationInput() must return (slot, true) after party upgrade")
			assert.Equal(t, PartyMessageInputID, got.ID,
				"AC-7 / BDD: 'iterationInput() returns the message slot' — Topic must NOT match the predicate")

			require.NotNil(t, def.Gate, "AC-7: Gate must be populated")
			assert.Equal(t, tt.wantMaxRounds, def.Gate.MaxRounds,
				"AC-7: MaxRounds default 100 / spec override")
			assert.Equal(t, tt.wantAcceptTokens, def.Gate.AcceptTokens,
				"AC-7: AcceptTokens default {'exit','done','wrap up'} / spec override; 'exit' presence is the BDD assertion")
		})
	}
}
