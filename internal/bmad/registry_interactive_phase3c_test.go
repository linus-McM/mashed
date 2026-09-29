// AC mapping for the Phase 3c implementation batch (single process):
//   AC-1 → TestStoryRollout05_AC1_QaProcessIsIterative
//   AC-2 → TestStoryRollout05_AC2_RoundResponseSlot
//   AC-3 → TestStoryRollout05_AC3_GateBaselineOnly
//   AC-4 → TestStoryRollout05_AC4_MemoryOnlyOutputs
//   AC-5 → TestStoryRollout05_AC5_SkipUnionIncludesPhase3c
package bmad

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const phase3cQaID = "bmad-qa-generate-e2e-tests"

func TestStoryRollout05_AC1_QaProcessIsIterative(t *testing.T) {
	p, ok := ProcessByID(phase3cQaID)
	require.True(t, ok, "process %q must exist in registry", phase3cQaID)
	assert.Equal(t, InteractIterative, p.Mode, "AC-1: Mode must be InteractIterative")
	assert.True(t, p.EnableAstAdapter, "AC-1: EnableAstAdapter must be true")
}

func TestStoryRollout05_AC2_RoundResponseSlot(t *testing.T) {
	p, ok := ProcessByID(phase3cQaID)
	require.True(t, ok)
	spec, found := p.iterationInput()
	require.True(t, found, "AC-2: iterationInput() must return a recurring slot")
	assert.Equal(t, RoundResponseInputID, spec.ID, "AC-2: iteration slot ID")
	assert.Equal(t, ShapeJSON, spec.Shape, "AC-2: iteration slot Shape")
}

func TestStoryRollout05_AC3_GateBaselineOnly(t *testing.T) {
	p, ok := ProcessByID(phase3cQaID)
	require.True(t, ok)
	require.NotNil(t, p.Gate, "AC-3: Gate must be populated")
	assert.Equal(t, GateUserConfirm, p.Gate.Kind, "AC-3: Gate.Kind")
	assert.Equal(t, 15, p.Gate.MaxRounds, "AC-3: 15-round cap")
	assert.Equal(t, []string{"done", "wrap up", "complete"}, p.Gate.AcceptTokens,
		"AC-3: AcceptTokens = baseline only (no domain extras)")
	assert.Equal(t, []string{"abort", "cancel"}, p.Gate.RejectTokens,
		"AC-3: RejectTokens common reject set")
}

func TestStoryRollout05_AC4_MemoryOnlyOutputs(t *testing.T) {
	p, ok := ProcessByID(phase3cQaID)
	require.True(t, ok)
	require.Len(t, p.OutputSpecs, 2, "AC-4: must declare 2 OutputSpecs (tests + any-doc)")
	for _, s := range p.OutputSpecs {
		assert.Equal(t, OutputToMemory, s.Target,
			"AC-4: %q output must be memory-only (unmapped artifact)", s.ID)
		assert.Empty(t, s.ArtifactName,
			"AC-4: %q ArtifactName must be empty for unmapped artifact", s.ID)
	}
}

func TestStoryRollout05_AC5_SkipUnionIncludesPhase3c(t *testing.T) {
	assert.Truef(t, slices.Contains(skippedFromU0Goldens, phase3cQaID),
		"AC-5: skippedFromU0Goldens must include %q", phase3cQaID)
}
