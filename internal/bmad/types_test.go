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

// ── Story ui-ast-U0, AC-1: ProcessDef.EnableAstAdapter round-trip ──
//
// RED Phase: EnableAstAdapter does not exist on ProcessDef yet. These tests
// MUST fail until the go-engineer adds the field (GREEN phase). Compile
// failure in this package blocks every bmad test — that is the intended RED
// signal.
func TestU0_AC1_ProcessDef_EnableAstAdapter_RoundTrip(t *testing.T) {
	base := ProcessDef{
		ID:          "bmad-test-u0",
		Name:        "Test",
		Phase:       PhaseAnalysis,
		AgentRole:   RoleAnalyst,
		SkillName:   "bmad-test-u0",
		Description: "desc",
		Inputs:      []string{},
		Outputs:     []string{},
		ModuleID:    "core",
		Version:     "1.0.0",
	}

	t.Run("EnableAstAdapter=true marshals with key and round-trips", func(t *testing.T) {
		p := base
		p.EnableAstAdapter = true

		data, err := json.Marshal(p)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"enableAstAdapter":true`,
			"JSON must include enableAstAdapter:true when the flag is set")

		var decoded ProcessDef
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.True(t, decoded.EnableAstAdapter,
			"decoded EnableAstAdapter must round-trip to true")
	})

	t.Run("EnableAstAdapter=false (zero) omits JSON key", func(t *testing.T) {
		p := base
		data, err := json.Marshal(p)
		require.NoError(t, err)
		assert.NotContains(t, string(data), "enableAstAdapter",
			"JSON must omit enableAstAdapter key when the flag is zero (omitempty)")
	})
}

// ── Story ui-ast-U4, AC-1: PendingPrompt.Structured round-trip ──
//
// RED Phase: Structured field does not exist on PendingPrompt yet. Compile
// failure in this package is the intended RED signal until T1-GREEN adds
// the field with `json:"structured,omitempty"`.
//
// Spec references:
//   - docs/stories/ui-ast-U4-executor-wiring.md §5.1 (lines 48–58, 410)
//   - AC-1: 4 KiB Structured survives byte-for-byte; empty omits JSON key.
func TestPendingPrompt_Structured_RoundTrip(t *testing.T) {
	base := PendingPrompt{
		NodeID:    "node-1",
		InputID:   "spec-confirm",
		Prompt:    "Confirm?",
		Shape:     ShapeJSON,
		Options:   []string{"done", "skip"},
		Round:     2,
		CreatedAt: 1700000000,
		PromptID:  "p-1",
	}

	t.Run("4 KiB Structured survives JSON round-trip byte-for-byte", func(t *testing.T) {
		// Build a ~4 KiB JSON-ish blob. The adapter emits a UIAST JSON string;
		// for the round-trip test the content is opaque — only identity matters.
		payload := make([]byte, 4096)
		for i := range payload {
			payload[i] = byte('a' + (i % 26))
		}
		structured := string(payload)

		p := base
		p.Structured = structured

		data, err := json.Marshal(p)
		require.NoError(t, err, "marshal must succeed")

		var decoded PendingPrompt
		require.NoError(t, json.Unmarshal(data, &decoded), "unmarshal must succeed")

		assert.Equal(t, structured, decoded.Structured,
			"Structured field must survive marshal/unmarshal byte-for-byte")
	})

	t.Run("non-empty Structured emits the structured JSON key", func(t *testing.T) {
		p := base
		p.Structured = `{"type":"unknown"}`

		data, err := json.Marshal(p)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"structured":`,
			"JSON must include structured key when Structured is non-empty")
	})

	t.Run("empty Structured omits the JSON key (omitempty)", func(t *testing.T) {
		p := base
		data, err := json.Marshal(p)
		require.NoError(t, err)
		assert.NotContains(t, string(data), `"structured":`,
			"JSON must omit structured key when Structured is empty (omitempty)")
	})
}

// ── Story ui-ast-U4, AC-12 setup: NodeInputEntry Round + Key round-trip ──
//
// RED Phase: NodeInputEntry.Key does not exist yet. Compile failure is the
// RED signal until T1-GREEN adds `Key string` (with omitempty) and ensures
// `Round int` carries `omitempty` so pre-U4 snapshots don't bloat with
// "round":0 "key":"" noise.
//
// Spec references:
//   - docs/stories/ui-ast-U4-executor-wiring.md §5.3.1 (lines 104–126)
//   - §5.3.2 gate walk relies on Round; §5.3.1 flatten writes per-sub-answer Key.
func TestNodeInputEntry_RoundAndKey_RoundTrip(t *testing.T) {
	t.Run("populated Round and Key round-trip via JSON", func(t *testing.T) {
		e := NodeInputEntry{
			InputID:   "spec-confirm",
			Round:     2,
			Value:     "done",
			Timestamp: 1700000000,
			Key:       "spec-confirm:confirm",
		}
		data, err := json.Marshal(e)
		require.NoError(t, err, "marshal must succeed")

		assert.Contains(t, string(data), `"round":2`,
			"JSON must include round when Round is non-zero")
		assert.Contains(t, string(data), `"key":"spec-confirm:confirm"`,
			"JSON must include key when Key is non-empty")

		var decoded NodeInputEntry
		require.NoError(t, json.Unmarshal(data, &decoded), "unmarshal must succeed")
		assert.Equal(t, e, decoded,
			"NodeInputEntry must round-trip via JSON")
	})

	t.Run("zero Round omits the round JSON key (omitempty)", func(t *testing.T) {
		e := NodeInputEntry{
			InputID: "spec-confirm",
			Value:   "done",
		}
		data, err := json.Marshal(e)
		require.NoError(t, err)
		assert.NotContains(t, string(data), `"round"`,
			"JSON must omit round key when Round is zero (omitempty)")
	})

	t.Run("empty Key omits the key JSON field (omitempty)", func(t *testing.T) {
		e := NodeInputEntry{
			InputID: "spec-confirm",
			Round:   1,
			Value:   "done",
		}
		data, err := json.Marshal(e)
		require.NoError(t, err)
		assert.NotContains(t, string(data), `"key"`,
			"JSON must omit key field when Key is empty (omitempty)")
	})

	t.Run("legacy entries without Key or Round decode cleanly", func(t *testing.T) {
		raw := `{"inputId":"spec-confirm","value":"done","timestamp":1700000000}`
		var e NodeInputEntry
		require.NoError(t, json.Unmarshal([]byte(raw), &e))
		assert.Equal(t, "spec-confirm", e.InputID)
		assert.Equal(t, "done", e.Value)
		assert.Zero(t, e.Round, "legacy absence of round must decode to zero")
		assert.Empty(t, e.Key, "legacy absence of key must decode to empty")
	})
}
