package uiadapter

import (
	"embed"
	"encoding/json"
	"fmt"
)

// Plan §3 Story 6 — structured output dispatch. One source of truth for
// the per-kind schemas (authored in Story A, embedded here so the Ollama
// `format:` payload and Claude `input_schema` carry byte-identical JSON).

//go:embed schemas/generate_yn.json schemas/generate_menu.json schemas/generate_form.json schemas/generate_text.json
var perKindSchemasFS embed.FS

// OllamaFormatPayload returns the JSON Schema object for format:<kind>
// on an Ollama chat request (Story v3-06 AC-6.1). When Config.LooseFormat
// is true the payload is the literal "json" string (AC-6.2 rollback path).
func OllamaFormatPayload(kind StageKind, loose bool) (json.RawMessage, error) {
	if loose {
		return json.RawMessage(`"json"`), nil
	}
	data, err := schemaBytes(kind)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ClaudeToolInputSchema returns the Anthropic `tools[0].input_schema`
// bytes for tool_use dispatch (AC-6.3). The per-kind schema file doubles
// as the tool schema — one source of truth.
func ClaudeToolInputSchema(kind StageKind) (json.RawMessage, error) {
	return schemaBytes(kind)
}

// ClaudeToolName is the tool name used on the Anthropic Messages API
// request. Router Story v3-16 passes it via `tool_choice: {type:"tool",
// name:"emit_uiast_<kind>"}`.
func ClaudeToolName(kind StageKind) string {
	return "emit_uiast_" + string(kind)
}

// schemaBytes loads the per-kind embedded schema. Centralised so
// TestSchemas_IdenticalAcrossBackends can assert byte-equality across
// every transport (AC-6.4).
func schemaBytes(kind StageKind) (json.RawMessage, error) {
	path := "schemas/generate_" + string(kind) + ".json"
	b, err := perKindSchemasFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("uiadapter: read per-kind schema %s: %w", path, err)
	}
	return json.RawMessage(b), nil
}
