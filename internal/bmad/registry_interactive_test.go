package bmad

// Tests for the interactive registry entries added in story
// bmad-interactive-07. Covers AC-1 (shape), AC-2 (iterationInput), and AC-7
// (OptionsRef resolution via registryLookup).

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInteractiveRegistryShape asserts Mode / InputSpec IDs / OutputSpec
// target / Gate fields for the four reference interactive processes match §10.
// AC-1.
func TestInteractiveRegistryShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id               string
		mode             InteractionMode
		inputIDs         []string
		outputTarget     OutputTarget
		outputArtifact   string
		outputOptional   bool
		gateKind         GateKind
		gateMaxRounds    int
		gateAcceptTokens []string
	}{
		{
			id:               "bmad-brainstorming",
			mode:             InteractIterative,
			inputIDs:         []string{"topic", "approach", "technique", "round-response"},
			outputTarget:     OutputToFile,
			outputArtifact:   "brainstorm-notes",
			gateKind:         GateUserConfirm,
			gateMaxRounds:    30,
			gateAcceptTokens: []string{"done", "wrap up", "finish"},
		},
		{
			id:               "bmad-product-brief",
			mode:             InteractGuided,
			inputIDs:         []string{"mode", "existing-brief", "brainstorm-input", "stage-response", "final-approval"},
			outputTarget:     OutputToFile,
			outputArtifact:   "product-brief",
			gateKind:         GateUserConfirm,
			gateMaxRounds:    10,
			gateAcceptTokens: []string{"yes"},
		},
		{
			id:               "bmad-party-mode",
			mode:             InteractParty,
			inputIDs:         []string{"topic", "message"},
			outputTarget:     OutputToFile,
			outputArtifact:   "retro-notes",
			outputOptional:   true,
			gateKind:         GateUserConfirm,
			gateMaxRounds:    100,
			gateAcceptTokens: []string{"exit", "done", "wrap up"},
		},
		{
			id:               "bmad-advanced-elicitation",
			mode:             InteractIterative,
			inputIDs:         []string{"target-content", "method", "apply-changes"},
			outputTarget:     OutputToBoth,
			outputArtifact:   "elicitation-notes",
			gateKind:         GateUserConfirm,
			gateMaxRounds:    20,
			gateAcceptTokens: []string{"x", "proceed", "done"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			proc, ok := ProcessByID(tt.id)
			require.True(t, ok, "process %q must exist in registry", tt.id)

			assert.Equal(t, tt.mode, proc.Mode, "Mode mismatch")

			gotIDs := make([]string, 0, len(proc.InputSpecs))
			for _, s := range proc.InputSpecs {
				gotIDs = append(gotIDs, s.ID)
			}
			assert.Equal(t, tt.inputIDs, gotIDs, "InputSpec IDs mismatch")

			require.Len(t, proc.OutputSpecs, 1, "expected exactly one OutputSpec")
			assert.Equal(t, tt.outputTarget, proc.OutputSpecs[0].Target, "OutputSpec.Target mismatch")
			assert.Equal(t, tt.outputArtifact, proc.OutputSpecs[0].ArtifactName, "OutputSpec.ArtifactName mismatch")
			assert.Equal(t, tt.outputOptional, proc.OutputSpecs[0].Optional, "OutputSpec.Optional mismatch")

			require.NotNil(t, proc.Gate, "Gate must not be nil")
			assert.Equal(t, tt.gateKind, proc.Gate.Kind)
			assert.Equal(t, tt.gateMaxRounds, proc.Gate.MaxRounds)
			assert.Equal(t, tt.gateAcceptTokens, proc.Gate.AcceptTokens)

			// Legacy back-compat: Inputs / Outputs must still be non-nil.
			assert.NotNil(t, proc.Inputs, "legacy Inputs slice must remain populated")
			assert.NotNil(t, proc.Outputs, "legacy Outputs slice must remain populated")
			assert.NotEmpty(t, proc.Outputs, "legacy Outputs must not be empty")
		})
	}
}

// TestInteractiveRegistryIterationInput asserts iterationInput() returns the
// expected per-round spec for each iterative/party/guided process. AC-2.
func TestInteractiveRegistryIterationInput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id     string
		wantID string
	}{
		{"bmad-brainstorming", "round-response"},
		{"bmad-product-brief", "stage-response"},
		{"bmad-party-mode", "message"},
		{"bmad-advanced-elicitation", "method"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.id, func(t *testing.T) {
			t.Parallel()
			proc, ok := ProcessByID(tt.id)
			require.True(t, ok)
			spec, found := proc.iterationInput()
			require.True(t, found, "iterationInput must return a spec for %q", tt.id)
			assert.Equal(t, tt.wantID, spec.ID)
		})
	}
}

// TestOptionsRefResolution drives registryLookup for both the column-extract
// and random=N CSV queries that the four interactive processes rely on. AC-7.
func TestOptionsRefResolution(t *testing.T) {
	t.Parallel()

	t.Run("column extract returns at least 3 values", func(t *testing.T) {
		t.Parallel()
		out, err := registryLookup("registry:brain-methods.csv#technique_name")
		require.NoError(t, err)
		parts := strings.Split(out, "\n")
		assert.GreaterOrEqual(t, len(parts), 3,
			"column extract must return at least 3 technique names, got %d: %v", len(parts), parts)
		for _, p := range parts {
			assert.NotEmpty(t, p, "every extracted value must be non-empty")
		}
	})

	t.Run("random=5 returns exactly 5 distinct entries", func(t *testing.T) {
		t.Parallel()
		out, err := registryLookup("registry:methods.csv?random=5")
		require.NoError(t, err)
		parts := strings.Split(out, "\n")
		require.Len(t, parts, 5, "random=5 must return exactly 5 rows")
		seen := make(map[string]bool, 5)
		for _, p := range parts {
			assert.NotEmpty(t, p)
			seen[p] = true
		}
		assert.Len(t, seen, 5, "random=5 rows must be distinct")
	})

	t.Run("resolveOptions splits newline-joined result", func(t *testing.T) {
		t.Parallel()
		spec := InputSpec{
			OptionsRef: "registry:methods.csv?random=5",
		}
		opts := resolveOptions(spec, nil)
		assert.Len(t, opts, 5, "resolveOptions must yield 5 distinct options")
	})
}
