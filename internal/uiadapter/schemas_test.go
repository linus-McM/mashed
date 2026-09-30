// Package uiadapter — Story A (Schema-to-Go codegen) RED-phase tests.
//
// These tests validate the canonical JSON schema set under
// `internal/uiadapter/schemas/` (Story A Developer Notes → Files (new)):
//
//   - `schemas/uiast.json`       — full UIAST envelope schema.
//   - `schemas/generate_yn.json` — strict subset for yes/no widget nodes.
//   - `schemas/generate_menu.json` — strict subset for choice/menu widgets.
//   - `schemas/generate_form.json` — strict subset for multi-field form widgets.
//   - `schemas/generate_text.json` — strict subset for text-input widgets.
//
// Rule: per-kind schemas must be STRICT SUBSETS of uiast.json so a single
// source of truth remains (Plan §8 risk "Ollama and Claude schemas drift").
package uiadapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const schemasDir = "schemas"

// Envelope keys the full UIAST schema must describe (§3.1 of the UI AST spec,
// mirrored in internal/uiadapter/schema.go hand-written `UIAST` struct tags).
var uiastEnvelopeKeys = []string{
	"version",
	"generated_by",
	"generated_at",
	"turn_summary",
	"nodes",
	"fallback_answer_shape",
	"diagnostics",
}

// Per-kind schema files that must ship alongside the full schema.
var perKindSchemas = []string{
	"generate_yn.json",
	"generate_menu.json",
	"generate_form.json",
	"generate_text.json",
}

// collectSchemaPropertyKeys walks a decoded JSON schema recursively and
// returns the set of every `properties` key anywhere in the tree. This lets
// the subset tests compare per-kind schemas against the full UIAST property
// graph without modelling `$ref`/`oneOf`/`definitions` by hand.
func collectSchemaPropertyKeys(node any) map[string]struct{} {
	out := map[string]struct{}{}
	var walk func(v any)
	walk = func(v any) {
		switch t := v.(type) {
		case map[string]any:
			if props, ok := t["properties"].(map[string]any); ok {
				for k, v := range props {
					out[k] = struct{}{}
					walk(v)
				}
			}
			// Recurse through common JSON-schema composition keywords so
			// properties nested under `oneOf` / `anyOf` / `allOf` / `items`
			// / `definitions` / `$defs` are discovered too.
			for _, key := range []string{"oneOf", "anyOf", "allOf", "items",
				"definitions", "$defs", "additionalProperties", "patternProperties"} {
				if child, ok := t[key]; ok {
					walk(child)
				}
			}
		case []any:
			for _, child := range t {
				walk(child)
			}
		}
	}
	walk(node)
	return out
}

// readSchema loads and JSON-decodes a schema under schemas/, returning a
// generic map so tests can introspect `properties`, `required`, `$schema`, etc.
func readSchema(t *testing.T, name string) map[string]any {
	t.Helper()
	path := filepath.Join(schemasDir, name)
	b, err := os.ReadFile(path)
	require.NoErrorf(t, err, "Story A: %s must exist", path)
	var out map[string]any
	require.NoErrorf(t, json.Unmarshal(b, &out), "Story A: %s must be valid JSON", path)
	return out
}

// TestUIASTSchemas_ValidJSONSchema — task spec literal name: loads every
// `schemas/*.json` file and asserts each is a valid JSON Schema draft-07
// document (declares `$schema`, `type`, and `properties`). Satisfies the
// sign-off entry point `go test -run TestUIASTSchemas_ValidJSONSchema`.
// Must FAIL at RED because `internal/uiadapter/schemas/` does not exist.
func TestUIASTSchemas_ValidJSONSchema(t *testing.T) {
	t.Parallel()
	entries, err := os.ReadDir(schemasDir)
	require.NoErrorf(t, err,
		"Story A: %s/ directory must exist with the canonical schemas", schemasDir)

	var jsonFiles []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			jsonFiles = append(jsonFiles, e.Name())
		}
	}
	require.NotEmptyf(t, jsonFiles,
		"Story A: %s/ must contain at least one *.json schema "+
			"(expected uiast.json plus per-kind shards)", schemasDir)

	// Story A Developer Notes list five canonical filenames — the test must
	// see all of them to pass.
	seen := map[string]bool{}
	for _, f := range jsonFiles {
		seen[f] = true
	}
	for _, want := range append([]string{"uiast.json"}, perKindSchemas...) {
		assert.Truef(t, seen[want],
			"Story A: expected canonical schema %s/%s (missing)", schemasDir, want)
	}

	for _, f := range jsonFiles {
		f := f
		t.Run(f, func(t *testing.T) {
			t.Parallel()
			schema := readSchema(t, f)

			// draft-07 requirement: `$schema` URI.
			sRaw, ok := schema["$schema"]
			require.Truef(t, ok, "Story A: %s/%s must declare `$schema`", schemasDir, f)
			s, ok := sRaw.(string)
			require.Truef(t, ok, "Story A: %s/%s `$schema` must be a string", schemasDir, f)
			assert.Containsf(t, s, "draft-07",
				"Story A: %s/%s must reference JSON Schema draft-07 "+
					"(go-jsonschema targets draft-07 by default)", schemasDir, f)

			// draft-07 requirement: top-level `type`.
			_, hasType := schema["type"]
			assert.Truef(t, hasType,
				"Story A: %s/%s must declare top-level `type`", schemasDir, f)

			// draft-07 requirement for our use-case: top-level `properties`
			// (every schema in this set describes an object).
			_, hasProps := schema["properties"]
			assert.Truef(t, hasProps,
				"Story A: %s/%s must declare top-level `properties`", schemasDir, f)
		})
	}
}

// TestSchemas_AC_A1_DirectoryExists — Story A mandates the canonical
// schema directory at `internal/uiadapter/schemas/`.
func TestSchemas_AC_A1_DirectoryExists(t *testing.T) {
	t.Parallel()
	info, err := os.Stat(schemasDir)
	require.NoErrorf(t, err, "Story A: %s/ directory must exist", schemasDir)
	assert.True(t, info.IsDir(), "Story A: %s must be a directory", schemasDir)
}

// TestSchemas_AC_A1_UIASTFileExistsAndIsJSON — AC-A.1 canonical schema.
func TestSchemas_AC_A1_UIASTFileExistsAndIsJSON(t *testing.T) {
	t.Parallel()
	readSchema(t, "uiast.json")
}

// TestSchemas_AC_A1_UIASTDeclaresDraftSchema — JSON Schema best practice
// (and go-jsonschema requirement): declare `$schema`. Keeps the canonical
// schema self-describing and lets linters / LSPs validate it.
func TestSchemas_AC_A1_UIASTDeclaresDraftSchema(t *testing.T) {
	t.Parallel()
	schema := readSchema(t, "uiast.json")
	raw, ok := schema["$schema"]
	require.True(t, ok, "Story A: schemas/uiast.json must declare `$schema` "+
		"(required by go-jsonschema for deterministic generation)")
	s, ok := raw.(string)
	require.True(t, ok, "Story A: schemas/uiast.json `$schema` must be a string")
	assert.Contains(t, s, "json-schema.org",
		"Story A: schemas/uiast.json `$schema` should reference json-schema.org")
}

// TestSchemas_AC_A1_UIASTDescribesEnvelope — the canonical schema must
// describe every §3.1 envelope key. Drives AC-A.1 "generated types are the
// canonical UIAST" by ensuring codegen emits each envelope field.
func TestSchemas_AC_A1_UIASTDescribesEnvelope(t *testing.T) {
	t.Parallel()
	schema := readSchema(t, "uiast.json")
	propsAny, ok := schema["properties"]
	require.True(t, ok, "Story A: schemas/uiast.json must declare `properties`")
	props, ok := propsAny.(map[string]any)
	require.True(t, ok, "Story A: schemas/uiast.json `properties` must be an object")

	for _, key := range uiastEnvelopeKeys {
		_, present := props[key]
		assert.Truef(t, present,
			"Story A §3.1: schemas/uiast.json must describe envelope key %q", key)
	}
}

// TestSchemas_AC_A1_UIASTEnvelopeRequired — §3.1 envelope fields are
// non-optional (hand-written UIAST has no `omitempty` on them). The
// canonical schema must mark them `required` so codegen emits them as
// value types rather than pointers.
func TestSchemas_AC_A1_UIASTEnvelopeRequired(t *testing.T) {
	t.Parallel()
	schema := readSchema(t, "uiast.json")
	reqAny, ok := schema["required"]
	require.True(t, ok, "Story A: schemas/uiast.json must declare `required`")
	reqList, ok := reqAny.([]any)
	require.True(t, ok, "Story A: schemas/uiast.json `required` must be an array")

	reqSet := map[string]struct{}{}
	for _, v := range reqList {
		if s, ok := v.(string); ok {
			reqSet[s] = struct{}{}
		}
	}
	for _, key := range uiastEnvelopeKeys {
		_, present := reqSet[key]
		assert.Truef(t, present,
			"Story A §3.1: envelope key %q must appear in schemas/uiast.json `required`", key)
	}
}

// TestSchemas_AC_A1_UIASTRejectsAdditionalProperties — the envelope is a
// closed shape (§4.7.1 strict decoding applies to widgets but the generated
// type for the envelope must also be deterministic). Setting
// `additionalProperties: false` forces go-jsonschema to emit a closed struct.
func TestSchemas_AC_A1_UIASTRejectsAdditionalProperties(t *testing.T) {
	t.Parallel()
	schema := readSchema(t, "uiast.json")
	raw, ok := schema["additionalProperties"]
	require.True(t, ok,
		"Story A: schemas/uiast.json must declare `additionalProperties` to keep "+
			"the generated struct closed")
	b, isBool := raw.(bool)
	require.True(t, isBool, "Story A: schemas/uiast.json `additionalProperties` must be boolean")
	assert.False(t, b, "Story A: schemas/uiast.json must set `additionalProperties: false`")
}

// TestSchemas_AC_A1_PerKindShardsExist — every per-kind narrow schema the
// story lists must ship (Subtask 2b).
func TestSchemas_AC_A1_PerKindShardsExist(t *testing.T) {
	t.Parallel()
	for _, name := range perKindSchemas {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			readSchema(t, name)
		})
	}
}

// TestSchemas_AC_A1_PerKindDeclareDraftSchema — per-kind schemas must also
// carry `$schema` so go-jsonschema can process them without inference.
func TestSchemas_AC_A1_PerKindDeclareDraftSchema(t *testing.T) {
	t.Parallel()
	for _, name := range perKindSchemas {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			schema := readSchema(t, name)
			_, ok := schema["$schema"]
			assert.Truef(t, ok, "Story A: schemas/%s must declare `$schema`", name)
		})
	}
}

// TestSchemas_AC_A1_PerKindSubsetsOfUIAST — Subtask 2b: per-kind narrow
// schemas must be STRICT SUBSETS of `uiast.json`. Any property key that
// appears in a per-kind schema must also appear somewhere in the uiast
// property graph (envelope, nodes, or widget), otherwise the two schemas
// would drift (Plan §8 risk).
func TestSchemas_AC_A1_PerKindSubsetsOfUIAST(t *testing.T) {
	t.Parallel()
	full := readSchema(t, "uiast.json")
	allKeys := collectSchemaPropertyKeys(full)

	for _, name := range perKindSchemas {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			sub := readSchema(t, name)
			for k := range collectSchemaPropertyKeys(sub) {
				_, ok := allKeys[k]
				assert.Truef(t, ok,
					"Story A Subtask 2b: schemas/%s declares property %q which "+
						"is not present anywhere in schemas/uiast.json — per-kind "+
						"schemas must be strict subsets of the canonical schema",
					name, k)
			}
		})
	}
}

// TestSchemas_AC_A1_PerKindRejectsAdditionalProperties — per-kind schemas
// must also be closed shapes so narrow Ollama `format:` / Claude `input_schema`
// payloads don't accidentally accept drifted fields.
func TestSchemas_AC_A1_PerKindRejectsAdditionalProperties(t *testing.T) {
	t.Parallel()
	for _, name := range perKindSchemas {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			schema := readSchema(t, name)
			raw, ok := schema["additionalProperties"]
			require.Truef(t, ok, "Story A: schemas/%s must declare `additionalProperties`", name)
			b, isBool := raw.(bool)
			require.Truef(t, isBool, "Story A: schemas/%s `additionalProperties` must be boolean", name)
			assert.Falsef(t, b, "Story A: schemas/%s must set `additionalProperties: false`", name)
		})
	}
}
