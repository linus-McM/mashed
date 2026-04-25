package uiadapter

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOllamaFormatPayload_SchemaObject — AC-6.1. With LooseFormat=false,
// the payload is the per-kind JSON Schema object (has type=object).
func TestOllamaFormatPayload_SchemaObject(t *testing.T) {
	t.Parallel()
	for _, k := range []StageKind{StageKindYN, StageKindMenu, StageKindForm, StageKindText} {
		raw, err := OllamaFormatPayload(k, false, nil)
		require.NoError(t, err)
		var decoded map[string]any
		require.NoError(t, json.Unmarshal(raw, &decoded))
		assert.Equal(t, "object", decoded["type"], "per-kind schema %s is an object", k)
	}
}

// TestOllamaFormatPayload_LooseRollback — AC-6.2. LooseFormat=true falls
// back to the literal "json" string for emergency rollback.
func TestOllamaFormatPayload_LooseRollback(t *testing.T) {
	t.Parallel()
	raw, err := OllamaFormatPayload(StageKindYN, true, nil)
	require.NoError(t, err)
	assert.JSONEq(t, `"json"`, string(raw))
}

// TestClaudeToolInputSchema_SchemaObject — AC-6.3. input_schema is the
// same per-kind schema; byte-equal to the Ollama format payload.
func TestClaudeToolInputSchema_SchemaObject(t *testing.T) {
	t.Parallel()
	raw, err := ClaudeToolInputSchema(StageKindMenu, nil)
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.Equal(t, "object", decoded["type"])
}

// TestSchemas_IdenticalAcrossBackends — AC-6.4. The canonical per-kind
// schema is byte-equal whether served as Ollama format or Claude
// input_schema.
func TestSchemas_IdenticalAcrossBackends(t *testing.T) {
	t.Parallel()
	for _, k := range []StageKind{StageKindYN, StageKindMenu, StageKindForm, StageKindText} {
		ollama, err := OllamaFormatPayload(k, false, nil)
		require.NoError(t, err)
		claude, err := ClaudeToolInputSchema(k, nil)
		require.NoError(t, err)
		assert.True(t, bytes.Equal(ollama, claude),
			"%s schema must be byte-identical across Ollama format: and Claude input_schema", k)
	}
}

// TestClaudeToolName_StableShape — router uses this literal string to
// match tool_choice entries.
func TestClaudeToolName_StableShape(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "emit_uiast_menu", ClaudeToolName(StageKindMenu))
	assert.Equal(t, "emit_uiast_yn", ClaudeToolName(StageKindYN))
}
