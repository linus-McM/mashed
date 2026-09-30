// AC mapping for the Phase 5 party batch:
//   AC-1 → TestStoryRollout08_AC1_PartyProcessesAreParty
//   AC-2 → TestStoryRollout08_AC2_TopicAndMessageInputs
//   AC-3 → TestStoryRollout08_AC3_IterationInputIsMessage
//   AC-4 → TestStoryRollout08_AC4_GateDefaults
//   AC-5 → TestStoryRollout08_AC5_OutputSpecsMapping
//   AC-6 → TestStoryRollout08_AC6_OptionalRetroNotes
//   AC-7 → TestStoryRollout08_AC7_SkipUnionIncludesPhase5
package bmad

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type phase5RolloutExpect struct {
	id            string
	topicID       string
	messagePrompt string
	output        outputTuple
	outputOpt     bool
}

var phase5RolloutExpectations = []phase5RolloutExpect{
	{
		id:            "bmad-retrospective",
		topicID:       "topic",
		messagePrompt: "Your turn. Type 'done' or 'exit' to end the retrospective.",
		output:        outputTuple{Target: OutputToFile, ArtifactName: "retro-notes"},
		outputOpt:     true,
	},
	{
		id:            "bmad-web-orchestrator",
		topicID:       "topic",
		messagePrompt: "Your turn. Type 'done' or 'exit' to end the session.",
		output:        outputTuple{Target: OutputToMemory, ArtifactName: ""},
	},
	{
		id:            "bmad-game-dev-studio",
		topicID:       "topic",
		messagePrompt: "Your turn. Type 'done' or 'exit' to end the session.",
		output:        outputTuple{Target: OutputToMemory, ArtifactName: ""},
	},
}

func TestStoryRollout08_AC1_PartyProcessesAreParty(t *testing.T) {
	for _, e := range phase5RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)
			assert.Equal(t, InteractParty, p.Mode, "AC-1: Mode must be InteractParty")
			assert.True(t, p.EnableAstAdapter, "AC-1: EnableAstAdapter must be true")
		})
	}
}

func TestStoryRollout08_AC2_TopicAndMessageInputs(t *testing.T) {
	for _, e := range phase5RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			require.Len(t, p.InputSpecs, 2, "AC-2: Party InputSpecs must be [topic, message]")
			assert.Equal(t, e.topicID, p.InputSpecs[0].ID, "AC-2: first slot is topic")
			assert.True(t, p.InputSpecs[0].Required, "AC-2: topic Required must be true")

			msg := p.InputSpecs[1]
			assert.Equal(t, PartyMessageInputID, msg.ID, "AC-2: second slot is message")
			assert.Equal(t, InputFromUser, msg.Source, "AC-2: message Source")
			assert.Equal(t, ShapeJSON, msg.Shape, "AC-2: message Shape")
			assert.False(t, msg.Required, "AC-2: message Required false (so iterationInput resolves to it)")
			assert.Equal(t, e.messagePrompt, msg.Prompt, "AC-2: message Prompt verbatim from spec")
		})
	}
}

func TestStoryRollout08_AC3_IterationInputIsMessage(t *testing.T) {
	for _, e := range phase5RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			spec, found := p.iterationInput()
			require.True(t, found, "AC-3: iterationInput() must return the message slot")
			assert.Equal(t, PartyMessageInputID, spec.ID,
				"AC-3: iteration slot must be the recurring message slot")
		})
	}
}

func TestStoryRollout08_AC4_GateDefaults(t *testing.T) {
	for _, e := range phase5RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			require.NotNil(t, p.Gate, "AC-4: Gate must be populated")
			assert.Equal(t, GateUserConfirm, p.Gate.Kind, "AC-4: Gate.Kind")
			assert.Equal(t, 100, p.Gate.MaxRounds, "AC-4: Party default MaxRounds=100")
			assert.Equal(t, []string{"exit", "done", "wrap up"}, p.Gate.AcceptTokens,
				"AC-4: Party default AcceptTokens")
		})
	}
}

func TestStoryRollout08_AC5_OutputSpecsMapping(t *testing.T) {
	for _, e := range phase5RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			require.Len(t, p.OutputSpecs, 1, "AC-5: exactly one OutputSpec")
			out := p.OutputSpecs[0]
			assert.Equal(t, e.output.Target, out.Target, "AC-5: Target")
			assert.Equal(t, e.output.ArtifactName, out.ArtifactName, "AC-5: ArtifactName")
		})
	}
}

func TestStoryRollout08_AC6_OptionalRetroNotes(t *testing.T) {
	p, ok := ProcessByID("bmad-retrospective")
	require.True(t, ok)
	require.Len(t, p.OutputSpecs, 1)
	assert.True(t, p.OutputSpecs[0].Optional,
		"AC-6: retro-notes OutputSpec must be Optional (matches party-mode reference §10.3)")
}

func TestStoryRollout08_AC7_SkipUnionIncludesPhase5(t *testing.T) {
	for _, e := range phase5RolloutExpectations {
		assert.Truef(t, slices.Contains(skippedFromU0Goldens, e.id),
			"AC-7: skippedFromU0Goldens must include %q", e.id)
	}
}
