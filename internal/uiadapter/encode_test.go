package uiadapter

import (
	"bytes"
	"encoding/json"
	"log/slog"
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
	assert.Equal(t, "emit_uiast_menu", ClaudeToolName(StageKindMenu, nil))
	assert.Equal(t, "emit_uiast_yn", ClaudeToolName(StageKindYN, nil))
}

// TestStory5_AC5_EncodeOllamaFormat — Story 5 AC-5.5.
// OllamaFormatPayload emits an encode.ollama_format record carrying op, kind,
// loose, bytes_out — never the schema bytes themselves.
func TestStory5_AC5_EncodeOllamaFormat(t *testing.T) {
	t.Parallel()
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	out, err := OllamaFormatPayload(StageKindForm, false, logger)
	require.NoError(t, err)
	require.NotEmpty(t, out)

	records := decodeRecords(t, buf)
	formatRecs := recordsByMsg(records, "encode.ollama_format")
	require.GreaterOrEqual(t, len(formatRecs), 1)
	assert.Equal(t, "encode.ollama_format", formatRecs[0]["op"])
	assert.Equal(t, string(StageKindForm), formatRecs[0]["kind"])
	assert.EqualValues(t, false, formatRecs[0]["loose"])
	bytesOut, ok := formatRecs[0]["bytes_out"].(float64)
	require.True(t, ok)
	assert.Greater(t, bytesOut, 0.0)
}

// TestStory5_AC5_EncodeClaudeToolSchema — Story 5 AC-5.5.
// ClaudeToolInputSchema emits encode.claude_tool_schema with kind + bytes_out.
func TestStory5_AC5_EncodeClaudeToolSchema(t *testing.T) {
	t.Parallel()
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	out, err := ClaudeToolInputSchema(StageKindMenu, logger)
	require.NoError(t, err)
	require.NotEmpty(t, out)

	records := decodeRecords(t, buf)
	schemaRecs := recordsByMsg(records, "encode.claude_tool_schema")
	require.GreaterOrEqual(t, len(schemaRecs), 1)
	assert.Equal(t, "encode.claude_tool", schemaRecs[0]["op"])
	assert.Equal(t, string(StageKindMenu), schemaRecs[0]["kind"])
	bytesOut, ok := schemaRecs[0]["bytes_out"].(float64)
	require.True(t, ok)
	assert.Greater(t, bytesOut, 0.0)
}

// TestStory5_AC5_EncodeClaudeToolName — Story 5 AC-5.5.
// ClaudeToolName takes a logger param (signature reshape) and emits
// encode.claude_tool_name with op, kind, name attrs.
func TestStory5_AC5_EncodeClaudeToolName(t *testing.T) {
	t.Parallel()
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	name := ClaudeToolName(StageKindYN, logger)
	require.Equal(t, "emit_uiast_yn", name)

	records := decodeRecords(t, buf)
	nameRecs := recordsByMsg(records, "encode.claude_tool_name")
	require.GreaterOrEqual(t, len(nameRecs), 1)
	assert.Equal(t, "encode.claude_tool_name", nameRecs[0]["op"])
	assert.Equal(t, string(StageKindYN), nameRecs[0]["kind"])
	assert.Equal(t, name, nameRecs[0]["name"])
}

// TestStory5_AC5_EncodeSchemaSelect — Story 5 AC-5.5.
// schemaBytes emits encode.schema_select alongside the wrapper emission.
// We exercise it indirectly via OllamaFormatPayload.
func TestStory5_AC5_EncodeSchemaSelect(t *testing.T) {
	t.Parallel()
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	_, err := OllamaFormatPayload(StageKindText, false, logger)
	require.NoError(t, err)

	records := decodeRecords(t, buf)
	selectRecs := recordsByMsg(records, "encode.schema_select")
	require.GreaterOrEqual(t, len(selectRecs), 1,
		"schemaBytes must emit encode.schema_select once per call")
	assert.Equal(t, "encode.schema_select", selectRecs[0]["op"])
	assert.Equal(t, string(StageKindText), selectRecs[0]["kind"])
	// variant attribute exists; precise value is impl-defined (embedded file
	// name). We only assert it's a non-empty string.
	variant, ok := selectRecs[0]["variant"].(string)
	require.True(t, ok, "variant must be a string")
	assert.NotEmpty(t, variant)
}
