package uiadapter

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
)

// Plan §3 Story 6 — structured output dispatch. One source of truth for
// the per-kind schemas (authored in Story A, embedded here so the Ollama
// `format:` payload and Claude `input_schema` carry byte-identical JSON).

//go:embed schemas/generate_yn.json schemas/generate_menu.json schemas/generate_form.json schemas/generate_text.json
var perKindSchemasFS embed.FS

// op-attr constants for every encode.* record. Centralised so the four
// emission sites can't drift apart.
const (
	encodeOllamaFormatOp   = "encode.ollama_format"
	encodeClaudeToolOp     = "encode.claude_tool"
	encodeClaudeToolNameOp = "encode.claude_tool_name"
	encodeSchemaSelectOp   = "encode.schema_select"
)

// OllamaFormatPayload returns the JSON Schema object for format:<kind>
// on an Ollama chat request (Story v3-06 AC-6.1). When Config.LooseFormat
// is true the payload is the literal "json" string (AC-6.2 rollback path).
// logger may be nil; nilSafeLogger normalises it so any future story can
// emit telemetry without an inline guard.
//
// Story 5: emits encode.ollama_format with kind, loose, bytes_out — never
// the schema bytes themselves.
func OllamaFormatPayload(kind StageKind, loose bool, logger *slog.Logger) (json.RawMessage, error) {
	lg := nilSafeLogger(logger)
	if loose {
		out := json.RawMessage(`"json"`)
		emitOllamaFormat(lg, kind, true, len(out))
		return out, nil
	}
	data, err := schemaBytes(kind, lg)
	if err != nil {
		return nil, err
	}
	emitOllamaFormat(lg, kind, false, len(data))
	return data, nil
}

// emitOllamaFormat writes the encode.ollama_format Debug record.
func emitOllamaFormat(logger *slog.Logger, kind StageKind, loose bool, bytesOut int) {
	ctx := context.Background()
	if !logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	logger.LogAttrs(ctx, slog.LevelDebug, "encode.ollama_format",
		slog.String("op", encodeOllamaFormatOp),
		slog.String("kind", string(kind)),
		slog.Bool("loose", loose),
		slog.Int("bytes_out", bytesOut),
	)
}

// ClaudeToolInputSchema returns the Anthropic `tools[0].input_schema`
// bytes for tool_use dispatch (AC-6.3). The per-kind schema file doubles
// as the tool schema — one source of truth. logger may be nil;
// nilSafeLogger normalises it so any future story can emit telemetry
// without an inline guard.
//
// Story 5: emits encode.claude_tool_schema with kind + bytes_out.
func ClaudeToolInputSchema(kind StageKind, logger *slog.Logger) (json.RawMessage, error) {
	lg := nilSafeLogger(logger)
	data, err := schemaBytes(kind, lg)
	if err != nil {
		return nil, err
	}
	if lg.Enabled(context.Background(), slog.LevelDebug) {
		lg.LogAttrs(context.Background(), slog.LevelDebug, "encode.claude_tool_schema",
			slog.String("op", encodeClaudeToolOp),
			slog.String("kind", string(kind)),
			slog.Int("bytes_out", len(data)),
		)
	}
	return data, nil
}

// ClaudeToolName is the tool name used on the Anthropic Messages API
// request. Router Story v3-16 passes it via `tool_choice: {type:"tool",
// name:"emit_uiast_<kind>"}`.
//
// Story 5: signature reshaped to take an explicit *slog.Logger; emits
// encode.claude_tool_name with op, kind, name attrs.
func ClaudeToolName(kind StageKind, logger *slog.Logger) string {
	lg := nilSafeLogger(logger)
	name := "emit_uiast_" + string(kind)
	if lg.Enabled(context.Background(), slog.LevelDebug) {
		lg.LogAttrs(context.Background(), slog.LevelDebug, "encode.claude_tool_name",
			slog.String("op", encodeClaudeToolNameOp),
			slog.String("kind", string(kind)),
			slog.String("name", name),
		)
	}
	return name
}

// schemaBytes loads the per-kind embedded schema. Centralised so
// TestSchemas_IdenticalAcrossBackends can assert byte-equality across
// every transport (AC-6.4).
//
// Story 5: emits encode.schema_select with op, kind, variant (the embedded
// file's basename — used when the per-kind schema is selected).
func schemaBytes(kind StageKind, logger *slog.Logger) (json.RawMessage, error) {
	variant := "generate_" + string(kind) + ".json"
	path := "schemas/" + variant
	b, err := perKindSchemasFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("uiadapter: read per-kind schema %s: %w", path, err)
	}
	if logger.Enabled(context.Background(), slog.LevelDebug) {
		logger.LogAttrs(context.Background(), slog.LevelDebug, "encode.schema_select",
			slog.String("op", encodeSchemaSelectOp),
			slog.String("kind", string(kind)),
			slog.String("variant", variant),
		)
	}
	return json.RawMessage(b), nil
}
