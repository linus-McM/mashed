package bmad

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── AC-4: ArtifactType constants ──

func TestAC4_ArtifactTypeConstants(t *testing.T) {
	tests := []struct {
		constant ArtifactType
		expected string
	}{
		{ArtifactMarkdown, "markdown"},
		{ArtifactYAML, "yaml"},
		{ArtifactDirectory, "directory"},
		{ArtifactCode, "code"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.constant), "ArtifactType constant mismatch")
		})
	}
}

// ── AC-4: ArtifactSpec JSON round-trip ──

func TestAC4_ArtifactSpec_JSONRoundTrip(t *testing.T) {
	t.Run("full spec round-trips correctly", func(t *testing.T) {
		spec := ArtifactSpec{
			Name:        "PRD.md",
			Type:        ArtifactMarkdown,
			Path:        "planning-artifacts/PRD.md",
			Description: "Product Requirements Document",
			Optional:    false,
		}
		data, err := json.Marshal(spec)
		require.NoError(t, err)

		var restored ArtifactSpec
		require.NoError(t, json.Unmarshal(data, &restored))
		assert.Equal(t, spec, restored)
	})

	t.Run("unmarshal from raw JSON with all fields", func(t *testing.T) {
		jsonStr := `{"name":"epics/","type":"directory","path":"solutioning-artifacts/epics/","description":"Epic definitions","optional":false}`
		var spec ArtifactSpec
		require.NoError(t, json.Unmarshal([]byte(jsonStr), &spec))

		assert.Equal(t, "epics/", spec.Name)
		assert.Equal(t, ArtifactDirectory, spec.Type)
		assert.Equal(t, "solutioning-artifacts/epics/", spec.Path)
		assert.Equal(t, "Epic definitions", spec.Description)
		assert.False(t, spec.Optional)
	})
}

func TestAC4_ArtifactSpec_ZeroValue(t *testing.T) {
	var spec ArtifactSpec
	data, err := json.Marshal(spec)
	require.NoError(t, err)

	var restored ArtifactSpec
	require.NoError(t, json.Unmarshal(data, &restored))
	assert.Equal(t, spec, restored)
}

// ── Story breadcrumbs-01, AC-4: Legacy WorkflowDef JSON round-trip ──
//
// RED Phase: verifies that loading a pre-story WorkflowDef and re-saving it
// does not inject reserved keys (inputPath, outputPath) into the JSON, and
// that a workflow that already carries inputPath preserves it faithfully.

func TestStory1_AC4_LegacyWorkflowRoundTrip(t *testing.T) {
	tests := []struct {
		name           string
		raw            string
		mustContain    []string
		mustNotContain []string
	}{
		{
			name: "legacy node without inputPath or outputPath",
			// Simulates a WorkflowDef saved before breadcrumbs-01 — Config has
			// unrelated keys only. Re-saving must not inject reserved keys.
			raw: `{
				"id": "legacy-wf-1",
				"name": "Legacy Workflow",
				"description": "",
				"nodes": [
					{
						"id": "n1",
						"processId": "util-file-loader",
						"label": "Input",
						"position": {"x": 100, "y": 100},
						"status": "pending",
						"config": {"processId": "util-file-loader", "label": "Input"},
						"tmuxTarget": ""
					}
				],
				"edges": [],
				"isTemplate": false,
				"createdAt": "2024-01-01T00:00:00Z"
			}`,
			mustNotContain: []string{`"inputPath"`, `"outputPath"`},
		},
		{
			name: "node with inputPath pre-set is preserved after round-trip",
			// Covers the shim behavior from breadcrumbs-05: a config that already
			// carries inputPath must survive marshal/unmarshal without data loss.
			raw: `{
				"id": "wf-with-paths",
				"name": "Workflow With Paths",
				"description": "",
				"nodes": [
					{
						"id": "n1",
						"processId": "util-file-loader",
						"label": "Input",
						"position": {"x": 100, "y": 100},
						"status": "pending",
						"config": {"inputPath": "/abs/input/file.md"},
						"tmuxTarget": ""
					}
				],
				"edges": [],
				"isTemplate": false,
				"createdAt": "2024-01-01T00:00:00Z"
			}`,
			mustContain:    []string{`"inputPath"`},
			mustNotContain: []string{`"outputPath"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var def WorkflowDef
			require.NoError(t, json.Unmarshal([]byte(tt.raw), &def), "initial unmarshal must succeed")

			remarshaled, err := json.Marshal(def)
			require.NoError(t, err, "re-marshal must succeed")

			remarshaledStr := string(remarshaled)

			for _, sub := range tt.mustContain {
				assert.Contains(t, remarshaledStr, sub, "re-marshaled JSON must contain %q", sub)
			}
			for _, sub := range tt.mustNotContain {
				assert.NotContains(t, remarshaledStr, sub, "re-marshaled JSON must not contain %q", sub)
			}

			// Second round-trip: re-unmarshal must be structurally equal.
			var def2 WorkflowDef
			require.NoError(t, json.Unmarshal(remarshaled, &def2), "second unmarshal must succeed")
			assert.Equal(t, def, def2, "double round-trip must produce structurally equal WorkflowDef")
		})
	}
}
