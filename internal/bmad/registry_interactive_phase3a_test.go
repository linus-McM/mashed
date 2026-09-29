// AC mapping for the Phase 3a analysis batch:
//   AC-1 → TestStoryRollout03_AC1_AnalysisProcessesAreIterative
//   AC-2 → TestStoryRollout03_AC2_RoundResponseSlot
//   AC-3 → TestStoryRollout03_AC3_GateBaselineOnly
//   AC-4 → TestStoryRollout03_AC4_OutputsAreFileMapped
//   AC-5 → TestStoryRollout03_AC5_SkipUnionIncludesPhase3a
package bmad

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// phase3aRolloutExpect: per-process expectations for the analysis batch.
type phase3aRolloutExpect struct {
	id           string
	artifactName string
}

var phase3aRolloutExpectations = []phase3aRolloutExpect{
	{id: "bmad-domain-research", artifactName: "domain-research"},
	{id: "bmad-market-research", artifactName: "market-research"},
	{id: "bmad-technical-research", artifactName: "tech-research"},
}

func TestStoryRollout03_AC1_AnalysisProcessesAreIterative(t *testing.T) {
	for _, e := range phase3aRolloutExpectations {
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

func TestStoryRollout03_AC2_RoundResponseSlot(t *testing.T) {
	for _, e := range phase3aRolloutExpectations {
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

func TestStoryRollout03_AC3_GateBaselineOnly(t *testing.T) {
	want := []string{"done", "wrap up", "complete"}
	for _, e := range phase3aRolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)

			require.NotNil(t, p.Gate, "AC-3: Gate must be populated for %q", e.id)
			assert.Equal(t, GateUserConfirm, p.Gate.Kind,
				"AC-3: Gate.Kind must be GateUserConfirm")
			assert.Equal(t, 30, p.Gate.MaxRounds,
				"AC-3: research processes use 30-round cap (Risk #6 natural cadence)")
			assert.Equal(t, want, p.Gate.AcceptTokens,
				"AC-3: AcceptTokens must equal baseline %v exactly (no domain extras)", want)
			assert.Equal(t, []string{"abort", "cancel"}, p.Gate.RejectTokens,
				"AC-3: RejectTokens must equal common reject set")
		})
	}
}

func TestStoryRollout03_AC4_OutputsAreFileMapped(t *testing.T) {
	for _, e := range phase3aRolloutExpectations {
		e := e
		t.Run(e.id, func(t *testing.T) {
			p, ok := ProcessByID(e.id)
			require.True(t, ok, "process %q must exist in registry", e.id)
			require.Len(t, p.OutputSpecs, 1,
				"AC-4: %q must declare exactly one OutputSpec", e.id)
			out := p.OutputSpecs[0]
			assert.Equal(t, OutputToFile, out.Target,
				"AC-4: research output must persist as file")
			assert.Equal(t, e.artifactName, out.ArtifactName,
				"AC-4: ArtifactName must match the mapped artifact constant")
		})
	}
}

func TestStoryRollout03_AC5_SkipUnionIncludesPhase3a(t *testing.T) {
	for _, e := range phase3aRolloutExpectations {
		assert.Truef(t, slices.Contains(skippedFromU0Goldens, e.id),
			"AC-5: skippedFromU0Goldens must include %q so TestU0_AC3 stays green for the upgraded entry", e.id)
	}
}
