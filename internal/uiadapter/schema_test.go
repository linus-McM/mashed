package uiadapter

import (
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestU2_Schema_UIAST_RoundTrip asserts §3.1 envelope keys survive a JSON
// marshal/unmarshal round-trip and every expected JSON key appears verbatim.
func TestU2_Schema_UIAST_RoundTrip(t *testing.T) {
	t.Parallel()
	original := &UIAST{
		Version:             "1",
		GeneratedBy:         "ollama:gemma3:4b",
		GeneratedAt:         1776681692,
		TurnSummary:         "turn summary",
		Nodes:               []UINode{{Type: "markdown", Content: "hello"}},
		FallbackAnswerShape: "free",
		Diagnostics:         Diagnostics{InputBytes: 100, OutputBytes: 200, LatencyMs: 50},
	}
	blob, err := json.Marshal(original)
	require.NoError(t, err)

	for _, key := range []string{
		`"version"`, `"generated_by"`, `"generated_at"`, `"turn_summary"`,
		`"nodes"`, `"fallback_answer_shape"`, `"diagnostics"`,
	} {
		assert.Contains(t, string(blob), key, "envelope must include JSON key %s per §3.1", key)
	}

	var decoded UIAST
	require.NoError(t, json.Unmarshal(blob, &decoded))
	assert.Equal(t, *original, decoded)
}

// TestU2_Schema_Diagnostics_CancelReasonNotSerialized guards the §4.3 rule that
// Diagnostics.CancelReason is an in-memory-only signal (JSON tag "-").
func TestU2_Schema_Diagnostics_CancelReasonNotSerialized(t *testing.T) {
	t.Parallel()
	d := Diagnostics{CancelReason: "context canceled by test", InputBytes: 1}
	blob, err := json.Marshal(d)
	require.NoError(t, err)
	out := string(blob)
	assert.NotContains(t, out, "CancelReason")
	assert.NotContains(t, out, "cancel_reason")
	assert.NotContains(t, out, "context canceled by test",
		"CancelReason must have JSON tag \"-\"; value must never reach the wire")
}

// TestU2_Schema_WidgetNode_RoundTrip asserts §3.2 widget keys (including the
// two camelCase ones — repoRootRelative, maxLength).
func TestU2_Schema_WidgetNode_RoundTrip(t *testing.T) {
	t.Parallel()
	w := WidgetNode{
		Type:        "choice",
		Options:     []WidgetOption{{Value: "a", Label: "Alpha"}, {Value: "b"}},
		Default:     "a",
		Min:         1,
		Max:         3,
		YesLabel:    "yes",
		NoLabel:     "no",
		Placeholder: "type",
		MaxLength:   2000,
		Multiline:   true,
		Accept:      []string{".md", ".txt"},
		RepoRootRel: true,
		Schema:      json.RawMessage(`{"type":"object"}`),
	}
	blob, err := json.Marshal(w)
	require.NoError(t, err)
	out := string(blob)

	assert.Contains(t, out, `"repoRootRelative":true`)
	assert.Contains(t, out, `"maxLength":2000`)
	assert.Contains(t, out, `"yes_label":"yes"`)
	assert.Contains(t, out, `"no_label":"no"`)
	assert.Contains(t, out, `"accept":[".md",".txt"]`)

	var decoded WidgetNode
	require.NoError(t, json.Unmarshal(blob, &decoded))
	assert.Equal(t, "choice", decoded.Type)
	assert.Equal(t, []WidgetOption{{Value: "a", Label: "Alpha"}, {Value: "b"}}, decoded.Options)
	assert.True(t, decoded.RepoRootRel)
	assert.Equal(t, []string{".md", ".txt"}, decoded.Accept)
}

// TestU2_Schema_UINode_DecisionGroup_RoundTrip asserts every §3.2 decision_group
// key serialises under its exact JSON name.
func TestU2_Schema_UINode_DecisionGroup_RoundTrip(t *testing.T) {
	t.Parallel()
	n := UINode{
		Type:        "decision_group",
		Heading:     "Output sink",
		Prompt:      "where?",
		Help:        "either stdout or file",
		Required:    true,
		Widget:      &WidgetNode{Type: "choice", Options: []WidgetOption{{Value: "stdout"}}},
		ResponseKey: "output-sink",
	}
	blob, err := json.Marshal(n)
	require.NoError(t, err)
	out := string(blob)

	assert.Contains(t, out, `"type":"decision_group"`)
	assert.Contains(t, out, `"heading":"Output sink"`)
	assert.Contains(t, out, `"prompt":"where?"`)
	assert.Contains(t, out, `"help":"either stdout or file"`)
	assert.Contains(t, out, `"required":true`)
	assert.Contains(t, out, `"response_key":"output-sink"`)
	assert.Contains(t, out, `"widget":`)

	var decoded UINode
	require.NoError(t, json.Unmarshal(blob, &decoded))
	assert.Equal(t, n.ResponseKey, decoded.ResponseKey)
	require.NotNil(t, decoded.Widget)
	assert.Equal(t, "choice", decoded.Widget.Type)
}

// TestU2_Schema_UINode_Table_RoundTrip asserts columns/rows shape.
func TestU2_Schema_UINode_Table_RoundTrip(t *testing.T) {
	t.Parallel()
	n := UINode{
		Type:    "table",
		Heading: "Options compared",
		Columns: []string{"Option", "Pros"},
		Rows:    [][]string{{"stdout", "Simple"}, {"file", "Persistent"}},
	}
	blob, err := json.Marshal(n)
	require.NoError(t, err)
	out := string(blob)
	assert.Contains(t, out, `"columns":["Option","Pros"]`)
	assert.Contains(t, out, `"rows":[["stdout","Simple"],["file","Persistent"]]`)

	var decoded UINode
	require.NoError(t, json.Unmarshal(blob, &decoded))
	assert.Equal(t, n.Rows, decoded.Rows)
}

// TestU2_Schema_UINode_Summary_RoundTrip asserts bullets serialize.
func TestU2_Schema_UINode_Summary_RoundTrip(t *testing.T) {
	t.Parallel()
	n := UINode{
		Type:    "summary",
		Heading: "Context loaded",
		Bullets: []string{"File: brief.md", "2 epics"},
	}
	blob, err := json.Marshal(n)
	require.NoError(t, err)
	assert.Contains(t, string(blob), `"bullets":["File: brief.md","2 epics"]`)
}

// TestU2_Schema_UINode_Hint_RoundTrip asserts tone serializes.
func TestU2_Schema_UINode_Hint_RoundTrip(t *testing.T) {
	t.Parallel()
	n := UINode{Type: "hint", Tone: "warn", Content: "heads up"}
	blob, err := json.Marshal(n)
	require.NoError(t, err)
	assert.Contains(t, string(blob), `"tone":"warn"`)
}

// TestU2_Schema_UINode_Code_RoundTrip asserts lang + copyable serialize.
func TestU2_Schema_UINode_Code_RoundTrip(t *testing.T) {
	t.Parallel()
	n := UINode{Type: "code", Lang: "python", Content: "print(1)", Copyable: true}
	blob, err := json.Marshal(n)
	require.NoError(t, err)
	out := string(blob)
	assert.Contains(t, out, `"lang":"python"`)
	assert.Contains(t, out, `"copyable":true`)
}

// TestU2_Schema_UINode_OptionalFieldsOmitted asserts omitempty tags keep
// non-applicable fields out of the wire format.
func TestU2_Schema_UINode_OptionalFieldsOmitted(t *testing.T) {
	t.Parallel()
	n := UINode{Type: "markdown", Content: "hi"}
	blob, err := json.Marshal(n)
	require.NoError(t, err)
	out := string(blob)

	assert.Contains(t, out, `"type":"markdown"`)
	assert.Contains(t, out, `"content":"hi"`)
	for _, absent := range []string{`"widget"`, `"bullets"`, `"response_key"`, `"rows"`, `"columns"`, `"tone"`, `"lang"`} {
		assert.NotContains(t, out, absent, "markdown node must not serialise %s", absent)
	}
}

// TestStory5_AC7_WidgetNodeUnmarshalEmits — Story 5 AC-5.7.
// When the package-level schemaLogger has been registered (e.g. by NewDefault),
// WidgetNode.UnmarshalJSON emits a schema.unmarshal.widget record per call
// carrying bytes_in. Parse failures emit schema.unmarshal.error with
// reason="parse" — never the raw err.Error() text.
func TestStory5_AC7_WidgetNodeUnmarshalEmits(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	// Save and restore the package-level schemaLogger pointer so this test
	// composes cleanly with TestStory5_AC7_NoLoggerFallback (which expects
	// the pointer to be unset). Sequential — t.Parallel() omitted because
	// schemaLogger is a process-wide singleton.
	prev := schemaLogger.Swap(logger)
	t.Cleanup(func() { schemaLogger.Store(prev) })

	// Happy path — well-formed widget JSON.
	payload := []byte(`{"type":"choice","options":[{"value":"a"}]}`)
	var w WidgetNode
	require.NoError(t, json.Unmarshal(payload, &w))

	records := decodeRecords(t, buf)
	widgetRecs := recordsByMsg(records, "schema.unmarshal.widget")
	require.GreaterOrEqual(t, len(widgetRecs), 1,
		"at least one schema.unmarshal.widget record per UnmarshalJSON call")
	assert.Equal(t, "schema.unmarshal", widgetRecs[0]["op"])
	assert.EqualValues(t, len(payload), widgetRecs[0]["bytes_in"])

	// Reset the buffer for the error-path assertion so the prior record set
	// doesn't pollute the failure check.
	buf.Reset()

	// Error path — malformed JSON. The error attr must be reason="parse"; the
	// underlying err.Error() string must NOT appear in the buffer.
	//
	// NB: must be syntactically valid JSON so json.Unmarshal dispatches to
	// WidgetNode.UnmarshalJSON. encoding/json's checkValid short-circuits on
	// truncated input (e.g. missing close brace) and never invokes the user
	// UnmarshalJSON, so the unknown-field rejection inside the method is the
	// only reachable parse-error site.
	bad := []byte(`{"type":"choice","sneaky":1}`) // valid syntax + unknown field
	var bad2 WidgetNode
	err := json.Unmarshal(bad, &bad2)
	require.Error(t, err)

	records = decodeRecords(t, buf)
	errRecs := recordsByMsg(records, "schema.unmarshal.error")
	require.GreaterOrEqual(t, len(errRecs), 1, "schema.unmarshal.error must emit on parse failure")
	assert.Equal(t, "schema.unmarshal", errRecs[0]["op"])
	assert.Equal(t, "parse", errRecs[0]["reason"])
	// err.Error() text must not leak.
	body := buf.String()
	assert.NotContains(t, body, err.Error(),
		"schema.unmarshal.error must never carry err.Error() text")
}

// TestStory5_AC7_NoLoggerFallback — Story 5 AC-5.7.
// When the package-level schemaLogger has not been registered (e.g. a raw
// json.Unmarshal in a test that did not call NewDefault), UnmarshalJSON
// must not panic and must not emit any records.
func TestStory5_AC7_NoLoggerFallback(t *testing.T) {
	prev := schemaLogger.Swap(nil)
	t.Cleanup(func() { schemaLogger.Store(prev) })

	// Sanity: nothing currently registered.
	require.True(t, schemaLogger.Load() == nil, "schemaLogger must be nil for this case")

	payload := []byte(`{"type":"choice","options":[{"value":"a"}]}`)
	var w WidgetNode
	require.NotPanics(t, func() {
		_ = json.Unmarshal(payload, &w)
	}, "Unmarshal with no registered logger must not panic")
	// No way to capture records without a logger; the contract is that the
	// implementation no-ops silently. The non-panic assertion above is the
	// observable signal.
}

// TestU2_Schema_WidgetOption_RoundTrip covers the tiny shape explicitly.
func TestU2_Schema_WidgetOption_RoundTrip(t *testing.T) {
	t.Parallel()
	o := WidgetOption{Value: "stdout", Label: "Plain stdout"}
	blob, err := json.Marshal(o)
	require.NoError(t, err)
	assert.Equal(t, `{"value":"stdout","label":"Plain stdout"}`, string(blob))

	var decoded WidgetOption
	require.NoError(t, json.Unmarshal(blob, &decoded))
	assert.Equal(t, o, decoded)

	// Label is optional; bare-value option must omit the empty label key.
	bare, err := json.Marshal(WidgetOption{Value: "file"})
	require.NoError(t, err)
	assert.Equal(t, `{"value":"file"}`, string(bare))
}
