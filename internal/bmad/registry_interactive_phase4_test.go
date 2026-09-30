// AC mapping for the Phase 4 guided batch:
//   AC-1 → TestStoryRollout07_AC1_GuidedProcessesAreGuided
//   AC-2 → TestStoryRollout07_AC2_NoIterationSlot
//   AC-3 → TestStoryRollout07_AC3_StagedInputsInDeclaredOrder
//   AC-4 → TestStoryRollout07_AC4_FinalApprovalAppended
//   AC-5 → TestStoryRollout07_AC5_OutputSpecsFileMapped
//   AC-6 → TestStoryRollout07_AC6_GateDefaults
//   AC-7 → TestStoryRollout07_AC7_SkipUnionIncludesPhase4
package bmad

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type phase4RolloutExpect struct {
	id            string
	stagedIDs     []string // declared order
	finalApprovID string
	artifactName  string
}

var phase4RolloutExpectations = []phase4RolloutExpect{
	{
		id:            "bmad-create-prd",
		stagedIDs:     []string{"scope", "audience", "timeline"},
		finalApprovID: "approve",
		artifactName:  "PRD.md",
	},
	{
		id:            "bmad-document-project",
		stagedIDs:     []string{"scope", "depth", "output-location"},
		finalApprovID: "approve",
		artifactName:  "project-docs",
	},
	{
		id:            "bmad-generate-project-context",
		stagedIDs:     []string{"scope", "sources"},
		finalApprovID: "approve",
		artifactName:  "project-context.md",
	},
}

func TestStoryRollout07_AC1_GuidedProcessesAreGuided(t *testing.T) {
	for _, e := range phase4RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)
			assert.Equal(t, InteractGuided, p.Mode, "AC-1: Mode must be InteractGuided")
			assert.True(t, p.EnableAstAdapter, "AC-1: EnableAstAdapter must be true")
		})
	}
}

func TestStoryRollout07_AC2_NoIterationSlot(t *testing.T) {
	for _, e := range phase4RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			_, found := p.iterationInput()
			assert.False(t, found,
				"AC-2: Guided processes must NOT expose a recurring iteration slot (helper appends none)")
		})
	}
}

func TestStoryRollout07_AC3_StagedInputsInDeclaredOrder(t *testing.T) {
	for _, e := range phase4RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			require.GreaterOrEqual(t, len(p.InputSpecs), len(e.stagedIDs),
				"AC-3: at least %d staged inputs", len(e.stagedIDs))
			for i, want := range e.stagedIDs {
				assert.Equal(t, want, p.InputSpecs[i].ID,
					"AC-3: staged input %d must have ID %q", i, want)
			}
		})
	}
}

func TestStoryRollout07_AC4_FinalApprovalAppended(t *testing.T) {
	for _, e := range phase4RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			require.Len(t, p.InputSpecs, len(e.stagedIDs)+1,
				"AC-4: InputSpecs must equal staged + 1 final approval")
			last := p.InputSpecs[len(p.InputSpecs)-1]
			assert.Equal(t, e.finalApprovID, last.ID,
				"AC-4: last InputSpec must be the FinalApproval")
			assert.Equal(t, ShapeApproval, last.Shape,
				"AC-4: FinalApproval Shape must be ShapeApproval")
			assert.True(t, last.Required, "AC-4: FinalApproval Required true")
		})
	}
}

func TestStoryRollout07_AC5_OutputSpecsFileMapped(t *testing.T) {
	for _, e := range phase4RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			require.Len(t, p.OutputSpecs, 1, "AC-5: exactly one OutputSpec")
			out := p.OutputSpecs[0]
			assert.Equal(t, OutputToFile, out.Target,
				"AC-5: OutputSpec must be file-mapped")
			assert.Equal(t, e.artifactName, out.ArtifactName,
				"AC-5: ArtifactName must match canonical mapping")
		})
	}
}

func TestStoryRollout07_AC6_GateDefaults(t *testing.T) {
	for _, e := range phase4RolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok)
			require.NotNil(t, p.Gate, "AC-6: Gate must be populated")
			assert.Equal(t, GateUserConfirm, p.Gate.Kind, "AC-6: Gate.Kind")
			assert.Equal(t, 10, p.Gate.MaxRounds, "AC-6: Guided default MaxRounds=10")
			assert.Equal(t, []string{"yes"}, p.Gate.AcceptTokens,
				"AC-6: Guided default AcceptTokens={'yes'} when spec leaves AcceptTokens nil")
		})
	}
}

func TestStoryRollout07_AC7_SkipUnionIncludesPhase4(t *testing.T) {
	for _, e := range phase4RolloutExpectations {
		assert.Truef(t, slices.Contains(skippedFromU0Goldens, e.id),
			"AC-7: skippedFromU0Goldens must include %q", e.id)
	}
}
