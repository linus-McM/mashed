// AC mapping for the Phase 3b planning batch:
//   AC-1 → TestStoryRollout04_AC1_PlanningProcessesAreIterative
//   AC-2 → TestStoryRollout04_AC2_RoundResponseSlot
//   AC-3 → TestStoryRollout04_AC3_GatePerProcess
//   AC-4 → TestStoryRollout04_AC4_OutputSpecsMapping
//   AC-5 → TestStoryRollout04_AC5_SkipUnionIncludesPhase3b
package bmad

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type phase3bRolloutExpect struct {
	id           string
	maxRounds    int
	domainAccept []string // appended after baseline
	outputs      []outputTuple
}

var phase3bRolloutExpectations = []phase3bRolloutExpect{
	{
		id:           "bmad-create-ux-design",
		maxRounds:    15,
		domainAccept: []string{"approved", "ship"},
		outputs: []outputTuple{
			{Target: OutputToFile, ArtifactName: "ux-spec.md"},
			{Target: OutputToMemory, ArtifactName: ""},
		},
	},
	{
		id:           "bmad-create-architecture",
		maxRounds:    15,
		domainAccept: []string{"approved", "ship"},
		outputs: []outputTuple{
			{Target: OutputToFile, ArtifactName: "architecture.md"},
		},
	},
	{
		id:           "bmad-check-implementation-readiness",
		maxRounds:    10,
		domainAccept: []string{"ready"},
		outputs: []outputTuple{
			{Target: OutputToFile, ArtifactName: "readiness-report"},
			{Target: OutputToMemory, ArtifactName: ""},
			{Target: OutputToMemory, ArtifactName: ""},
		},
	},
	{
		id:           "bmad-create-epics-and-stories",
		maxRounds:    20,
		domainAccept: nil,
		outputs: []outputTuple{
			{Target: OutputToMemory, ArtifactName: ""},
		},
	},
}

func TestStoryRollout04_AC1_PlanningProcessesAreIterative(t *testing.T) {
	for _, e := range phase3bRolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)
			assert.Equal(t, InteractIterative, p.Mode,
				"AC-1: %q must be InteractIterative", e.id)
			assert.True(t, p.EnableAstAdapter,
				"AC-1: %q must opt into the AST adapter", e.id)
		})
	}
}

func TestStoryRollout04_AC2_RoundResponseSlot(t *testing.T) {
	for _, e := range phase3bRolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)
			spec, found := p.iterationInput()
			require.True(t, found,
				"AC-2: iterationInput() must return a recurring slot for %q", e.id)
			assert.Equal(t, RoundResponseInputID, spec.ID,
				"AC-2: iteration slot ID must be %q", RoundResponseInputID)
			assert.Equal(t, ShapeJSON, spec.Shape,
				"AC-2: iteration slot Shape must be ShapeJSON")
		})
	}
}

func TestStoryRollout04_AC3_GatePerProcess(t *testing.T) {
	for _, e := range phase3bRolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)
			require.NotNil(t, p.Gate, "AC-3: Gate must be populated for %q", e.id)
			assert.Equal(t, GateUserConfirm, p.Gate.Kind,
				"AC-3: Gate.Kind must be GateUserConfirm")
			assert.Equal(t, e.maxRounds, p.Gate.MaxRounds,
				"AC-3: Gate.MaxRounds must equal per-process override")
			want := append([]string{"done", "wrap up", "complete"}, e.domainAccept...)
			assert.Equal(t, want, p.Gate.AcceptTokens,
				"AC-3: AcceptTokens = baseline + domain")
			assert.Equal(t, []string{"abort", "cancel"}, p.Gate.RejectTokens,
				"AC-3: RejectTokens must equal common reject set")
		})
	}
}

func TestStoryRollout04_AC4_OutputSpecsMapping(t *testing.T) {
	for _, e := range phase3bRolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)
			require.Len(t, p.OutputSpecs, len(e.outputs),
				"AC-4: %q must declare %d OutputSpec(s)", e.id, len(e.outputs))
			got := make([]outputTuple, len(p.OutputSpecs))
			for i, s := range p.OutputSpecs {
				got[i] = outputTuple{Target: s.Target, ArtifactName: s.ArtifactName}
			}
			assert.ElementsMatch(t, e.outputs, got,
				"AC-4: OutputSpecs (target, artifactName) multiset mismatch")
		})
	}
}

func TestStoryRollout04_AC5_SkipUnionIncludesPhase3b(t *testing.T) {
	for _, e := range phase3bRolloutExpectations {
		assert.Truef(t, slices.Contains(skippedFromU0Goldens, e.id),
			"AC-5: skippedFromU0Goldens must include %q", e.id)
	}
}
