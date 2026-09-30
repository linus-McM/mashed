// Package bmad_test contains JSON round-trip tests for interactive BMAD types.
// Story bmad-interactive-01: Core types for interactive BMAD processes
//
// RED Phase: These tests define expected behaviour for all new types and
// constants declared in §3 of docs/bmad-interactive-process-schema.md.
// They MUST fail to compile until the go-engineer adds the new symbols to
// internal/bmad/types.go.
package bmad_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/bmad"
)

// ---------------------------------------------------------------------------
// AC-2 / BDD: NodeAwaitingInput status serialises as awaiting_input
// ---------------------------------------------------------------------------

// TestNodeAwaitingInputJSONRoundTrip marshals a WorkflowNode whose Status is
// NodeAwaitingInput, asserts the wire value is the literal "awaiting_input",
// then unmarshals and asserts the status is preserved.
func TestNodeAwaitingInputJSONRoundTrip(t *testing.T) {
	node := bmad.WorkflowNode{
		ID:     "n1",
		Status: bmad.NodeAwaitingInput,
	}

	data, err := json.Marshal(node)
	require.NoError(t, err)

	assert.True(t, strings.Contains(string(data), `"status":"awaiting_input"`),
		"marshalled JSON must contain \"status\":\"awaiting_input\", got: %s", string(data))

	var got bmad.WorkflowNode
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, bmad.NodeAwaitingInput, got.Status)
}

// ---------------------------------------------------------------------------
// AC-3 / BDD: Legacy ProcessDef round-trip is unchanged
// ---------------------------------------------------------------------------

// TestProcessDefLegacyShape confirms that a ProcessDef populated with only
// the pre-existing fields does not emit mode, inputSpecs, outputSpecs, or
// gate in its JSON output, and that a round-trip preserves equality.
func TestProcessDefLegacyShape(t *testing.T) {
	original := bmad.ProcessDef{
		ID:          "brainstorm",
		Name:        "Brainstorm",
		Phase:       bmad.PhaseAnalysis,
		AgentRole:   bmad.RoleAnalyst,
		SkillName:   "brainstorm",
		Description: "Generate ideas",
		Inputs:      []string{"brief"},
		Outputs:     []string{"ideas"},
		ModuleID:    "core",
		Version:     "1.0.0",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	raw := string(data)
	assert.False(t, strings.Contains(raw, `"mode"`),
		"legacy ProcessDef must not emit \"mode\", got: %s", raw)
	assert.False(t, strings.Contains(raw, `"inputSpecs"`),
		"legacy ProcessDef must not emit \"inputSpecs\", got: %s", raw)
	assert.False(t, strings.Contains(raw, `"outputSpecs"`),
		"legacy ProcessDef must not emit \"outputSpecs\", got: %s", raw)
	assert.False(t, strings.Contains(raw, `"gate"`),
		"legacy ProcessDef must not emit \"gate\", got: %s", raw)

	var got bmad.ProcessDef
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, original, got)
}

// ---------------------------------------------------------------------------
// AC-4 / BDD: InputSpec per-shape table
// ---------------------------------------------------------------------------

// TestInputSpecShapes is a table-driven test covering every InputSource and
// every InputShape combination mandated by AC-4. Each row marshals, unmarshals,
// and asserts deep equality plus specific JSON key presence from §3.1.
func TestInputSpecShapes(t *testing.T) {
	tests := []struct {
		name         string
		input        bmad.InputSpec
		wantJSONKeys []string // substrings that must be present in the JSON
		wantNoKeys   []string // substrings that must NOT be present
	}{
		{
			name: "InputFromFile with ArtifactName",
			input: bmad.InputSpec{
				ID:           "brief",
				Source:       bmad.InputFromFile,
				Required:     true,
				ArtifactName: "product-brief",
			},
			wantJSONKeys: []string{`"source":"file"`, `"artifactName":"product-brief"`, `"required":true`},
			wantNoKeys:   []string{`"upstreamNodeId"`, `"shape"`, `"optionsRef"`},
		},
		{
			name: "InputFromUpstream with UpstreamNodeID",
			input: bmad.InputSpec{
				ID:             "priorOutput",
				Source:         bmad.InputFromUpstream,
				Required:       true,
				UpstreamNodeID: "node-analyze",
			},
			wantJSONKeys: []string{`"source":"upstream"`, `"upstreamNodeId":"node-analyze"`},
			wantNoKeys:   []string{`"artifactName"`, `"optionsRef"`},
		},
		{
			name: "InputFromUser ShapeFree with Validation MaxLength HelpText",
			input: bmad.InputSpec{
				ID:         "topic",
				Source:     bmad.InputFromUser,
				Shape:      bmad.ShapeFree,
				Required:   true,
				Prompt:     "Enter a topic",
				MaxLength:  500,
				HelpText:   "Be concise",
				Validation: `^.{1,500}$`,
			},
			wantJSONKeys: []string{
				`"source":"user"`,
				`"shape":"free"`,
				`"prompt":"Enter a topic"`,
				`"maxLength":500`,
				`"helpText":"Be concise"`,
				`"validation":"^.{1,500}$"`,
			},
		},
		{
			name: "InputFromUser ShapeChoice with static Options",
			input: bmad.InputSpec{
				ID:       "method",
				Source:   bmad.InputFromUser,
				Shape:    bmad.ShapeChoice,
				Required: true,
				Prompt:   "Pick a method",
				Options:  []string{"six-hats", "scamper", "biomimicry"},
			},
			wantJSONKeys: []string{`"source":"user"`, `"shape":"choice"`, `"options":`},
		},
		{
			name: "InputFromUser ShapeChoice with OptionsRef dynamic lookup",
			input: bmad.InputSpec{
				ID:         "method",
				Source:     bmad.InputFromUser,
				Shape:      bmad.ShapeChoice,
				Required:   true,
				Prompt:     "Pick a method",
				OptionsRef: "registry:methods.csv?random=5",
			},
			wantJSONKeys: []string{
				`"source":"user"`,
				`"shape":"choice"`,
				`"optionsRef":"registry:methods.csv?random=5"`,
			},
		},
		{
			name: "InputFromUser ShapeMultiChoice",
			input: bmad.InputSpec{
				ID:      "tags",
				Source:  bmad.InputFromUser,
				Shape:   bmad.ShapeMultiChoice,
				Options: []string{"a", "b", "c"},
			},
			wantJSONKeys: []string{`"source":"user"`, `"shape":"multi"`},
		},
		{
			name: "InputFromUser ShapeApproval",
			input: bmad.InputSpec{
				ID:     "approve",
				Source: bmad.InputFromUser,
				Shape:  bmad.ShapeApproval,
				Prompt: "Approve?",
			},
			wantJSONKeys: []string{`"source":"user"`, `"shape":"approval"`},
		},
		{
			name: "InputFromUser ShapeFile",
			input: bmad.InputSpec{
				ID:     "upload",
				Source: bmad.InputFromUser,
				Shape:  bmad.ShapeFile,
				Prompt: "Upload a file",
			},
			wantJSONKeys: []string{`"source":"user"`, `"shape":"file"`},
		},
		{
			name: "InputFromUser ShapeJSON with Default",
			input: bmad.InputSpec{
				ID:      "config",
				Source:  bmad.InputFromUser,
				Shape:   bmad.ShapeJSON,
				Default: `{"key":"value"}`,
			},
			wantJSONKeys: []string{`"source":"user"`, `"shape":"json"`},
		},
		{
			name: "InputFromEnv",
			input: bmad.InputSpec{
				ID:     "branch",
				Source: bmad.InputFromEnv,
			},
			wantJSONKeys: []string{`"source":"env"`},
			wantNoKeys:   []string{`"shape"`, `"artifactName"`, `"upstreamNodeId"`, `"optionsRef"`},
		},
		{
			name: "InputFromRegistry with OptionsRef",
			input: bmad.InputSpec{
				ID:         "technique",
				Source:     bmad.InputFromRegistry,
				OptionsRef: "registry:techniques.csv#technique_name",
			},
			wantJSONKeys: []string{`"source":"registry"`, `"optionsRef":"registry:techniques.csv#technique_name"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.input)
			require.NoError(t, err, "marshal must not error")

			raw := string(data)
			for _, key := range tt.wantJSONKeys {
				assert.True(t, strings.Contains(raw, key),
					"JSON must contain %q, got: %s", key, raw)
			}
			for _, key := range tt.wantNoKeys {
				assert.False(t, strings.Contains(raw, key),
					"JSON must NOT contain %q, got: %s", key, raw)
			}

			var got bmad.InputSpec
			require.NoError(t, json.Unmarshal(data, &got), "unmarshal must not error")
			assert.True(t, reflect.DeepEqual(tt.input, got),
				"round-trip must preserve all fields\nwant: %+v\n got: %+v", tt.input, got)
		})
	}
}

// ---------------------------------------------------------------------------
// AC-6 / BDD: IterationGate nil vs populated round-trip
// ---------------------------------------------------------------------------

// TestIterationGateNilVsPopulated asserts that Gate:nil emits no "gate" key,
// while a populated gate emits a "gate" object with the correct kind value.
func TestIterationGateNilVsPopulated(t *testing.T) {
	t.Run("nil gate omitted from JSON", func(t *testing.T) {
		pd := bmad.ProcessDef{
			ID:   "p1",
			Gate: nil,
		}
		data, err := json.Marshal(pd)
		require.NoError(t, err)
		assert.False(t, strings.Contains(string(data), `"gate"`),
			"ProcessDef{Gate:nil} must not emit \"gate\", got: %s", string(data))
	})

	t.Run("GateUserConfirm emits gate object with kind userConfirm", func(t *testing.T) {
		pd := bmad.ProcessDef{
			ID: "p2",
			Gate: &bmad.IterationGate{
				Kind:         bmad.GateUserConfirm,
				AcceptTokens: []string{"done"},
				MaxRounds:    30,
			},
		}
		data, err := json.Marshal(pd)
		require.NoError(t, err)

		raw := string(data)
		assert.True(t, strings.Contains(raw, `"gate":`),
			"must contain \"gate\":, got: %s", raw)
		assert.True(t, strings.Contains(raw, `"kind":"userConfirm"`),
			"gate must have kind \"userConfirm\", got: %s", raw)
		assert.True(t, strings.Contains(raw, `"maxRounds":30`),
			"gate must have maxRounds:30, got: %s", raw)
		assert.True(t, strings.Contains(raw, `"acceptTokens":`),
			"gate must have acceptTokens, got: %s", raw)
	})

	t.Run("GateRoundLimit with MaxRounds=0 still emits gate object", func(t *testing.T) {
		// AC-6: Gate: &IterationGate{Kind: GateRoundLimit, MaxRounds: 0} must still
		// marshal to a gate object (the pointer being non-nil is the discriminator).
		pd := bmad.ProcessDef{
			ID: "p3",
			Gate: &bmad.IterationGate{
				Kind:      bmad.GateRoundLimit,
				MaxRounds: 0,
			},
		}
		data, err := json.Marshal(pd)
		require.NoError(t, err)

		raw := string(data)
		assert.True(t, strings.Contains(raw, `"gate":`),
			"non-nil Gate with MaxRounds=0 must still emit \"gate\":, got: %s", raw)
		assert.True(t, strings.Contains(raw, `"kind":"rounds"`),
			"GateRoundLimit must serialise kind as \"rounds\", got: %s", raw)
	})

	t.Run("IterationGate full round-trip with all fields", func(t *testing.T) {
		original := bmad.IterationGate{
			Kind:         bmad.GateUserConfirm,
			MaxRounds:    30,
			AcceptTokens: []string{"done", "proceed"},
			RejectTokens: []string{"abort", "cancel"},
			CustomExpr:   "",
		}
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var got bmad.IterationGate
		require.NoError(t, json.Unmarshal(data, &got))
		assert.True(t, reflect.DeepEqual(original, got),
			"round-trip must satisfy reflect.DeepEqual\nwant: %+v\n got: %+v", original, got)
	})

	t.Run("GateExpression with CustomExpr", func(t *testing.T) {
		gate := &bmad.IterationGate{
			Kind:       bmad.GateExpression,
			CustomExpr: "len(outputs) > 0",
			MaxRounds:  10,
		}
		data, err := json.Marshal(gate)
		require.NoError(t, err)

		raw := string(data)
		assert.True(t, strings.Contains(raw, `"kind":"expression"`), "got: %s", raw)
		assert.True(t, strings.Contains(raw, `"customExpr":"len(outputs) \u003e 0"`), "got: %s", raw)

		var got bmad.IterationGate
		require.NoError(t, json.Unmarshal(data, &got))
		assert.Equal(t, *gate, got)
	})
}

// ---------------------------------------------------------------------------
// AC-5 / BDD: WorkflowExecution deep-equals after round-trip
// ---------------------------------------------------------------------------

// TestWorkflowExecutionInteractiveState constructs a WorkflowExecution with
// every new interactive field populated and asserts deep equality after a
// JSON round-trip.
func TestWorkflowExecutionInteractiveState(t *testing.T) {
	original := bmad.WorkflowExecution{
		ID:          "exec-1",
		WorkflowID:  "wf-brainstorm",
		RepoPath:    "/home/dev/project",
		Status:      bmad.ExecRunning,
		StartedAt:   "2024-01-01T00:00:00Z",
		CurrentNode: "n1",
		// New interactive fields
		NodeRounds: map[string]int{
			"n1": 3,
		},
		PendingPrompts: []bmad.PendingPrompt{
			{
				NodeID:    "n1",
				InputID:   "method",
				Prompt:    "Pick a brainstorming method",
				Shape:     bmad.ShapeChoice,
				Options:   []string{"six-hats", "scamper", "biomimicry"},
				Round:     3,
				CreatedAt: 1704067200,
				PromptID:  "abc123",
			},
		},
		NodeInputs: map[string]map[string]string{
			"n1": {"method": "six-hats"},
		},
		NodeInputHistory: map[string][]bmad.NodeInputEntry{
			"n1": {
				{
					InputID:   "method",
					Round:     1,
					Value:     "six-hats",
					Timestamp: 1704067200,
				},
			},
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	raw := string(data)
	assert.True(t, strings.Contains(raw, `"nodeRounds"`), "must contain nodeRounds key")
	assert.True(t, strings.Contains(raw, `"pendingPrompts"`), "must contain pendingPrompts key")
	assert.True(t, strings.Contains(raw, `"nodeInputs"`), "must contain nodeInputs key")
	assert.True(t, strings.Contains(raw, `"nodeInputHistory"`), "must contain nodeInputHistory key")

	var got bmad.WorkflowExecution
	require.NoError(t, json.Unmarshal(data, &got))
	assert.True(t, reflect.DeepEqual(original, got),
		"round-trip must satisfy reflect.DeepEqual\nwant: %+v\n got: %+v", original, got)
}

// ---------------------------------------------------------------------------
// AC-2 / BDD: PendingPrompt preserves PromptID hash slot
// ---------------------------------------------------------------------------

// TestPendingPromptRoundTrip asserts that PromptID, Round, and CreatedAt all
// survive a JSON round-trip with the exact wire key names from §3.5.
func TestPendingPromptRoundTrip(t *testing.T) {
	original := bmad.PendingPrompt{
		NodeID:    "n1",
		InputID:   "approval",
		Prompt:    "Do you approve?",
		Shape:     bmad.ShapeApproval,
		Round:     2,
		CreatedAt: 1704067200,
		PromptID:  "abc123",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	raw := string(data)
	assert.True(t, strings.Contains(raw, `"promptId":"abc123"`),
		"must contain \"promptId\":\"abc123\", got: %s", raw)
	assert.True(t, strings.Contains(raw, `"round":2`),
		"must contain \"round\":2, got: %s", raw)
	assert.True(t, strings.Contains(raw, `"createdAt":1704067200`),
		"must contain \"createdAt\":1704067200, got: %s", raw)
	assert.True(t, strings.Contains(raw, `"nodeId":"n1"`),
		"must contain \"nodeId\":\"n1\", got: %s", raw)
	assert.True(t, strings.Contains(raw, `"inputId":"approval"`),
		"must contain \"inputId\":\"approval\", got: %s", raw)
	assert.True(t, strings.Contains(raw, `"shape":"approval"`),
		"must contain \"shape\":\"approval\", got: %s", raw)

	var got bmad.PendingPrompt
	require.NoError(t, json.Unmarshal(data, &got))
	assert.Equal(t, original.PromptID, got.PromptID)
	assert.Equal(t, original.Round, got.Round)
	assert.Equal(t, original.CreatedAt, got.CreatedAt)
	assert.True(t, reflect.DeepEqual(original, got),
		"round-trip must satisfy reflect.DeepEqual\nwant: %+v\n got: %+v", original, got)
}

// ---------------------------------------------------------------------------
// Supplementary: constant wire-value contract
// ---------------------------------------------------------------------------

// TestConstantJSONValues asserts that every new constant carries exactly the
// JSON value mandated by §3. A rename of any constant value will surface here.
func TestConstantJSONValues(t *testing.T) {
	// InputSource constants (§3.1)
	assert.Equal(t, bmad.InputSource("file"), bmad.InputFromFile)
	assert.Equal(t, bmad.InputSource("upstream"), bmad.InputFromUpstream)
	assert.Equal(t, bmad.InputSource("user"), bmad.InputFromUser)
	assert.Equal(t, bmad.InputSource("env"), bmad.InputFromEnv)
	assert.Equal(t, bmad.InputSource("registry"), bmad.InputFromRegistry)

	// InputShape constants (§3.1)
	assert.Equal(t, bmad.InputShape("free"), bmad.ShapeFree)
	assert.Equal(t, bmad.InputShape("choice"), bmad.ShapeChoice)
	assert.Equal(t, bmad.InputShape("multi"), bmad.ShapeMultiChoice)
	assert.Equal(t, bmad.InputShape("approval"), bmad.ShapeApproval)
	assert.Equal(t, bmad.InputShape("file"), bmad.ShapeFile)
	assert.Equal(t, bmad.InputShape("json"), bmad.ShapeJSON)

	// OutputTarget constants (§3.2)
	assert.Equal(t, bmad.OutputTarget("file"), bmad.OutputToFile)
	assert.Equal(t, bmad.OutputTarget("memory"), bmad.OutputToMemory)
	assert.Equal(t, bmad.OutputTarget("both"), bmad.OutputToBoth)

	// InteractionMode constants (§3.3)
	assert.Equal(t, bmad.InteractionMode("autonomous"), bmad.InteractAutonomous)
	assert.Equal(t, bmad.InteractionMode("guided"), bmad.InteractGuided)
	assert.Equal(t, bmad.InteractionMode("iterative"), bmad.InteractIterative)
	assert.Equal(t, bmad.InteractionMode("party"), bmad.InteractParty)

	// GateKind constants (§3.3)
	assert.Equal(t, bmad.GateKind("userConfirm"), bmad.GateUserConfirm)
	assert.Equal(t, bmad.GateKind("artifact"), bmad.GateArtifactExists)
	assert.Equal(t, bmad.GateKind("expression"), bmad.GateExpression)
	assert.Equal(t, bmad.GateKind("rounds"), bmad.GateRoundLimit)

	// NodeAwaitingInput — snake_case per §3.5 (only snake_case status value)
	assert.Equal(t, bmad.WorkflowNodeStatus("awaiting_input"), bmad.NodeAwaitingInput)
}

// ---------------------------------------------------------------------------
// Supplementary: OutputSpec round-trip
// ---------------------------------------------------------------------------

// TestOutputSpecRoundTrip verifies OutputSpec marshals and unmarshals losslessly
// with the correct JSON tag names from §3.2.
func TestOutputSpecRoundTrip(t *testing.T) {
	tests := []struct {
		name         string
		spec         bmad.OutputSpec
		wantJSONKeys []string
	}{
		{
			name: "OutputToFile with ArtifactName and Description",
			spec: bmad.OutputSpec{
				ID:           "report",
				Target:       bmad.OutputToFile,
				ArtifactName: "sprint-report",
				Description:  "Sprint summary",
				Optional:     false,
			},
			wantJSONKeys: []string{
				`"id":"report"`,
				`"target":"file"`,
				`"artifactName":"sprint-report"`,
				`"description":"Sprint summary"`,
			},
		},
		{
			name: "OutputToMemory optional",
			spec: bmad.OutputSpec{
				ID:       "scratch",
				Target:   bmad.OutputToMemory,
				Optional: true,
			},
			wantJSONKeys: []string{`"target":"memory"`, `"optional":true`},
		},
		{
			name: "OutputToBoth",
			spec: bmad.OutputSpec{
				ID:           "artifact",
				Target:       bmad.OutputToBoth,
				ArtifactName: "design-doc",
			},
			wantJSONKeys: []string{`"target":"both"`, `"artifactName":"design-doc"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.spec)
			require.NoError(t, err)

			raw := string(data)
			for _, key := range tt.wantJSONKeys {
				assert.True(t, strings.Contains(raw, key),
					"JSON must contain %q, got: %s", key, raw)
			}

			var got bmad.OutputSpec
			require.NoError(t, json.Unmarshal(data, &got))
			assert.True(t, reflect.DeepEqual(tt.spec, got),
				"round-trip must preserve all fields\nwant: %+v\n got: %+v", tt.spec, got)
		})
	}
}

// ---------------------------------------------------------------------------
// Supplementary: ProcessDef interactive fields round-trip
// ---------------------------------------------------------------------------

// TestProcessDefInteractiveFieldsRoundTrip verifies that when Mode, InputSpecs,
// OutputSpecs, and Gate are set on a ProcessDef, they survive a round-trip.
func TestProcessDefInteractiveFieldsRoundTrip(t *testing.T) {
	original := bmad.ProcessDef{
		ID:        "brainstorm-interactive",
		Name:      "Brainstorm Interactive",
		Phase:     bmad.PhaseAnalysis,
		AgentRole: bmad.RoleAnalyst,
		Mode:      bmad.InteractIterative,
		InputSpecs: []bmad.InputSpec{
			{
				ID:         "method",
				Source:     bmad.InputFromUser,
				Shape:      bmad.ShapeChoice,
				Required:   true,
				Prompt:     "Pick a method",
				OptionsRef: "registry:methods.csv?random=5",
			},
		},
		OutputSpecs: []bmad.OutputSpec{
			{
				ID:           "ideas",
				Target:       bmad.OutputToFile,
				ArtifactName: "brainstorm-output",
			},
		},
		Gate: &bmad.IterationGate{
			Kind:         bmad.GateUserConfirm,
			AcceptTokens: []string{"done", "proceed"},
			RejectTokens: []string{"abort"},
			MaxRounds:    30,
		},
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	raw := string(data)
	assert.True(t, strings.Contains(raw, `"mode":"iterative"`), "must contain mode:iterative")
	assert.True(t, strings.Contains(raw, `"inputSpecs":`), "must contain inputSpecs")
	assert.True(t, strings.Contains(raw, `"outputSpecs":`), "must contain outputSpecs")
	assert.True(t, strings.Contains(raw, `"gate":`), "must contain gate")

	var got bmad.ProcessDef
	require.NoError(t, json.Unmarshal(data, &got))
	assert.True(t, reflect.DeepEqual(original, got),
		"round-trip must satisfy reflect.DeepEqual\nwant: %+v\n got: %+v", original, got)
}

// ---------------------------------------------------------------------------
// Supplementary: NodeInputEntry round-trip
// ---------------------------------------------------------------------------

// TestNodeInputEntryRoundTrip verifies NodeInputEntry marshals with the exact
// JSON tags declared in §3.5.
func TestNodeInputEntryRoundTrip(t *testing.T) {
	original := bmad.NodeInputEntry{
		InputID:   "method",
		Round:     1,
		Value:     "six-hats",
		Timestamp: 1704067200,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	raw := string(data)
	assert.True(t, strings.Contains(raw, `"inputId":"method"`), "must contain inputId key, got: %s", raw)
	assert.True(t, strings.Contains(raw, `"round":1`), "must contain round key, got: %s", raw)
	assert.True(t, strings.Contains(raw, `"value":"six-hats"`), "must contain value key, got: %s", raw)
	assert.True(t, strings.Contains(raw, `"timestamp":1704067200`), "must contain timestamp key, got: %s", raw)

	var got bmad.NodeInputEntry
	require.NoError(t, json.Unmarshal(data, &got))
	assert.True(t, reflect.DeepEqual(original, got),
		"round-trip must satisfy reflect.DeepEqual\nwant: %+v\n got: %+v", original, got)
}

// ---------------------------------------------------------------------------
// Supplementary: WorkflowExecution zero interactive fields emit no new keys
// ---------------------------------------------------------------------------

// TestWorkflowExecutionLegacyShape verifies that a WorkflowExecution with only
// pre-existing fields set does not emit the new interactive keys.
func TestWorkflowExecutionLegacyShape(t *testing.T) {
	exec := bmad.WorkflowExecution{
		ID:         "exec-legacy",
		WorkflowID: "wf-1",
		RepoPath:   "/repo",
		Status:     bmad.ExecComplete,
		StartedAt:  "2024-01-01T00:00:00Z",
	}

	data, err := json.Marshal(exec)
	require.NoError(t, err)

	raw := string(data)
	assert.False(t, strings.Contains(raw, `"nodeRounds"`),
		"zero-value WorkflowExecution must not emit \"nodeRounds\", got: %s", raw)
	assert.False(t, strings.Contains(raw, `"pendingPrompts"`),
		"zero-value WorkflowExecution must not emit \"pendingPrompts\", got: %s", raw)
	assert.False(t, strings.Contains(raw, `"nodeInputs"`),
		"zero-value WorkflowExecution must not emit \"nodeInputs\", got: %s", raw)
	assert.False(t, strings.Contains(raw, `"nodeInputHistory"`),
		"zero-value WorkflowExecution must not emit \"nodeInputHistory\", got: %s", raw)
}
