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
