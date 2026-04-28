// AC mapping for the Phase 3d support batch:
//   AC-1 → TestStoryRollout06_AC1_SupportProcessesAreIterative
//   AC-2 → TestStoryRollout06_AC2_RoundResponseSlot
//   AC-3 → TestStoryRollout06_AC3_GatePerProcess
//   AC-4 → TestStoryRollout06_AC4_OutputSpecsMapping
//   AC-5 → TestStoryRollout06_AC5_SkipUnionIncludesPhase3d
package bmad

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type phase3dRolloutExpect struct {
	id           string
	maxRounds    int
	domainAccept []string
	outputs      []outputTuple
}

var phase3dRolloutExpectations = []phase3dRolloutExpect{
	{
		id: "bmad-editorial-review-prose", maxRounds: 15,
		domainAccept: []string{"approved"},
		outputs:      []outputTuple{{Target: OutputToFile, ArtifactName: "reviewed-doc"}},
	},
	{
		id: "bmad-editorial-review-structure", maxRounds: 15,
		domainAccept: []string{"approved"},
		outputs:      []outputTuple{{Target: OutputToFile, ArtifactName: "reviewed-doc"}},
	},
	{
		id: "bmad-review-edge-case-hunter", maxRounds: 15,
		domainAccept: []string{"approved"},
		outputs:      []outputTuple{{Target: OutputToFile, ArtifactName: "edge-case-report"}},
	},
	{
		id: "bmad-quick-flow", maxRounds: 10,
		domainAccept: []string{"ship"},
		outputs: []outputTuple{
			{Target: OutputToMemory, ArtifactName: ""},
			{Target: OutputToFile, ArtifactName: "PRD.md"},
		},
	},
	{
		id: "bmad-adversarial-general", maxRounds: 15,
		domainAccept: nil,
		outputs:      []outputTuple{{Target: OutputToFile, ArtifactName: "adversarial-report"}},
	},
	{
		id: "bmad-infrastructure-devops", maxRounds: 15,
		domainAccept: []string{"ship"},
		outputs:      []outputTuple{{Target: OutputToFile, ArtifactName: "infra-config"}},
	},
}

func TestStoryRollout06_AC1_SupportProcessesAreIterative(t *testing.T) {
	for _, e := range phase3dRolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)
			assert.Equal(t, InteractIterative, p.Mode, "AC-1: Mode")
			assert.True(t, p.EnableAstAdapter, "AC-1: EnableAstAdapter")
		})
	}
}

func TestStoryRollout06_AC2_RoundResponseSlot(t *testing.T) {
	for _, e := range phase3dRolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			spec, found := p.iterationInput()
			require.True(t, found, "AC-2: iterationInput()")
			assert.Equal(t, RoundResponseInputID, spec.ID, "AC-2: slot ID")
			assert.Equal(t, ShapeJSON, spec.Shape, "AC-2: slot Shape")
		})
	}
}

func TestStoryRollout06_AC3_GatePerProcess(t *testing.T) {
	for _, e := range phase3dRolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			require.NotNil(t, p.Gate, "AC-3: Gate")
			assert.Equal(t, GateUserConfirm, p.Gate.Kind, "AC-3: Gate.Kind")
			assert.Equal(t, e.maxRounds, p.Gate.MaxRounds, "AC-3: MaxRounds")
			want := append([]string{"done", "wrap up", "complete"}, e.domainAccept...)
			assert.Equal(t, want, p.Gate.AcceptTokens, "AC-3: AcceptTokens")
			assert.Equal(t, []string{"abort", "cancel"}, p.Gate.RejectTokens, "AC-3: RejectTokens")
		})
	}
}

func TestStoryRollout06_AC4_OutputSpecsMapping(t *testing.T) {
	for _, e := range phase3dRolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			require.Len(t, p.OutputSpecs, len(e.outputs),
				"AC-4: OutputSpecs count")
			got := make([]outputTuple, len(p.OutputSpecs))
			for i, s := range p.OutputSpecs {
				got[i] = outputTuple{Target: s.Target, ArtifactName: s.ArtifactName}
			}
			assert.ElementsMatch(t, e.outputs, got, "AC-4: OutputSpec multiset")
		})
	}
}

func TestStoryRollout06_AC5_SkipUnionIncludesPhase3d(t *testing.T) {
	for _, e := range phase3dRolloutExpectations {
		assert.Truef(t, slices.Contains(skippedFromU0Goldens, e.id),
			"AC-5: skippedFromU0Goldens must include %q", e.id)
	}
}
