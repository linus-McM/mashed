// Package bmad contains tests for the iteration gate evaluator and helpers.
// Story bmad-interactive-04: Iteration gate + round loop.
//
// RED Phase: These tests define expected behaviour for checkGate,
// containsToken, and iterationInput. They MUST fail until the go-engineer
// implements the feature (GREEN phase).
//
// Tests use the white-box package (package bmad) to access unexported helpers.
package bmad

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── AC-1: GateUserConfirm ─────────────────────────────────────────────────────

// TestCheckGateUserConfirmAcceptToken (AC-1)
// Given Gate{Kind: GateUserConfirm, AcceptTokens: ["done","finish"]}.
// Table rows cover: exact match, case-insensitive match, trimmed whitespace match,
// non-matching answer, and nil gate (single-round shortcut).
func TestCheckGateUserConfirmAcceptToken(t *testing.T) {
	t.Parallel()

	const nodeID = "n1"

	tests := []struct {
		name       string
		lastAnswer string
		tokens     []string
		wantHit    bool
		wantReason string
	}{
		{
			name:       "exact match returns true with accept-token reason",
			lastAnswer: "done",
			tokens:     []string{"done", "finish"},
			wantHit:    true,
			wantReason: "accept-token matched",
		},
		{
			name:       "DONE upper-case is case-insensitive match",
			lastAnswer: "DONE",
			tokens:     []string{"done", "finish"},
			wantHit:    true,
			wantReason: "accept-token matched",
		},
		{
			name:       "whitespace-padded finish is trimmed and matched",
			lastAnswer: "  finish  ",
			tokens:     []string{"done", "finish"},
			wantHit:    true,
			wantReason: "accept-token matched",
		},
		{
			name:       "non-matching answer returns false with empty reason",
			lastAnswer: "keep going",
			tokens:     []string{"done", "finish"},
			wantHit:    false,
			wantReason: "",
		},
		{
			name:       "empty answer with tokens returns false",
			lastAnswer: "",
			tokens:     []string{"done", "finish"},
			wantHit:    false,
			wantReason: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state, _ := newSessionState([]WorkflowNode{
				{
					ID:        nodeID,
					ProcessID: "test-gate-userconfirm",
					Status:    NodeRunning,
					NodeType:  NodeTypeProcess,
				},
			}, nil)

			// Seed NodeInputHistory with the last answer.
			state.exec.NodeInputHistory = map[string][]NodeInputEntry{
				nodeID: {
					{InputID: "round-response", Round: 1, Value: tt.lastAnswer},
				},
			}

			gate := &IterationGate{
				Kind:         GateUserConfirm,
				AcceptTokens: tt.tokens,
			}

			// checkGate does not exist yet — RED phase.
			e := NewExecutor(nil, func(string, interface{}) {})
			hit, reason := e.checkGate(state, gate, nodeID, 1)

			assert.Equal(t, tt.wantHit, hit, "hit mismatch")
			assert.Equal(t, tt.wantReason, reason, "reason mismatch")
		})
	}
}

// ── AC-2: GateRoundLimit ──────────────────────────────────────────────────────

// TestCheckGateRoundLimit (AC-2)
// GateRoundLimit exits when round >= MaxRounds.
func TestCheckGateRoundLimit(t *testing.T) {
	t.Parallel()

	const nodeID = "n1"

	tests := []struct {
		name      string
		maxRounds int
		round     int
		wantHit   bool
		wantReason string
	}{
		{
			name:      "round 1 of 3 is not hit",
			maxRounds: 3,
			round:     1,
			wantHit:   false,
		},
		{
			name:       "round 3 of 3 is hit — round limit reached",
			maxRounds:  3,
			round:      3,
			wantHit:    true,
			wantReason: "round limit reached",
		},
		{
			name:      "round 2 of 3 is not hit",
			maxRounds: 3,
			round:     2,
			wantHit:   false,
		},
		{
			name:      "MaxRounds=0 means no cap — round 0 not hit",
			maxRounds: 0,
			round:     0,
			wantHit:   false,
		},
		{
			name:      "MaxRounds=0 means no cap — round 99 not hit",
			maxRounds: 0,
			round:     99,
			wantHit:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state, _ := newSessionState([]WorkflowNode{
				{ID: nodeID, ProcessID: "test-gate-rounds", Status: NodeRunning, NodeType: NodeTypeProcess},
			}, nil)

			gate := &IterationGate{
				Kind:      GateRoundLimit,
				MaxRounds: tt.maxRounds,
			}

			e := NewExecutor(nil, func(string, interface{}) {})
			hit, reason := e.checkGate(state, gate, nodeID, tt.round)

			assert.Equal(t, tt.wantHit, hit, "hit mismatch for round=%d maxRounds=%d", tt.round, tt.maxRounds)
			if tt.wantHit {
				assert.Equal(t, tt.wantReason, reason)
			}
		})
	}
}

// ── AC-4: GateArtifactExists ──────────────────────────────────────────────────

// TestCheckGateArtifactExists (AC-4)
// Gate exits when the declared artifact file appears on disk.
func TestCheckGateArtifactExists(t *testing.T) {
	// NOT parallel: uses testRegistry global to register a process.
	const nodeID = "n1"
	const processID = "test-gate-artifact-exists"

	artifactName := "product-brief" // maps to analysis-artifacts/product-brief.md

	repoPath := t.TempDir()

	registerTestProcess(t, ProcessDef{
		ID:        processID,
		Name:      "Test Gate Artifact",
		Mode:      InteractIterative,
		SkillName: "test-skill",
		InputSpecs: []InputSpec{},
		OutputSpecs: []OutputSpec{
			{ID: "out1", Target: OutputToFile, ArtifactName: artifactName},
		},
		Gate: &IterationGate{Kind: GateArtifactExists},
	})

	state, _ := newSessionState([]WorkflowNode{
		{ID: nodeID, ProcessID: processID, Status: NodeRunning, NodeType: NodeTypeProcess},
	}, nil)
	state.exec.RepoPath = repoPath

	gate := &IterationGate{Kind: GateArtifactExists}

	e := NewExecutor(nil, func(string, interface{}) {})

	// Round 1: file absent → false.
	hit1, reason1 := e.checkGate(state, gate, nodeID, 1)
	assert.False(t, hit1, "file absent: gate must return false")
	assert.Empty(t, reason1, "file absent: reason must be empty")

	// Create the file at the expected path.
	artifactDir := filepath.Join(repoPath, "_bmad-output", "analysis-artifacts")
	require.NoError(t, os.MkdirAll(artifactDir, 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(artifactDir, "product-brief.md"),
		[]byte("# Product Brief\nContent."),
		0o644,
	))

	// Round 2: file present → true "artifact present".
	hit2, reason2 := e.checkGate(state, gate, nodeID, 2)
	assert.True(t, hit2, "file present: gate must return true")
	assert.Equal(t, "artifact present", reason2)

	// Missing OutputSpecs → false.
	registerTestProcess(t, ProcessDef{
		ID:          processID + "-noout",
		Name:        "Test Gate Artifact No Out",
		Mode:        InteractIterative,
		SkillName:   "test-skill",
		InputSpecs:  []InputSpec{},
		OutputSpecs: []OutputSpec{}, // no OutputSpecs
		Gate:        &IterationGate{Kind: GateArtifactExists},
	})

	state2, _ := newSessionState([]WorkflowNode{
		{ID: nodeID, ProcessID: processID + "-noout", Status: NodeRunning, NodeType: NodeTypeProcess},
	}, nil)
	state2.exec.RepoPath = repoPath

	hit3, reason3 := e.checkGate(state2, gate, nodeID, 1)
	assert.False(t, hit3, "no OutputSpecs: gate must return false")
	assert.Empty(t, reason3)

	// Empty ArtifactName → false.
	registerTestProcess(t, ProcessDef{
		ID:          processID + "-noartname",
		Name:        "Test Gate Artifact No ArtName",
		Mode:        InteractIterative,
		SkillName:   "test-skill",
		InputSpecs:  []InputSpec{},
		OutputSpecs: []OutputSpec{{ID: "out1", Target: OutputToFile, ArtifactName: ""}},
		Gate:        &IterationGate{Kind: GateArtifactExists},
	})

	state3, _ := newSessionState([]WorkflowNode{
		{ID: nodeID, ProcessID: processID + "-noartname", Status: NodeRunning, NodeType: NodeTypeProcess},
	}, nil)
	state3.exec.RepoPath = repoPath

	hit4, reason4 := e.checkGate(state3, gate, nodeID, 1)
	assert.False(t, hit4, "empty ArtifactName: gate must return false")
	assert.Empty(t, reason4)
}

// ── GateExpression (basic) ────────────────────────────────────────────────────

// TestCheckGateExpression
// Minimal validation of the GateExpression path.
// If the evaluator signature is unclear from GREEN code, these rows test
// the safe-failure paths (empty/malformed CustomExpr → false).
//
// NOTE: Full expression evaluation will be validated once the GREEN engineer
// exposes evaluateExpr with a documented signature. These tests drive
// discovery of that interface.
func TestCheckGateExpression(t *testing.T) {
	t.Parallel()

	const nodeID = "n1"

	tests := []struct {
		name       string
		customExpr string
		nodeOutputs map[string]string
		wantHit    bool
	}{
		{
			name:        "empty CustomExpr returns false (no expression to evaluate)",
			customExpr:  "",
			nodeOutputs: map[string]string{nodeID: "some output"},
			wantHit:     false,
		},
		{
			name:        "malformed expression returns false (evaluator error)",
			customExpr:  "!@#$%^&*() invalid",
			nodeOutputs: map[string]string{nodeID: "some output"},
			wantHit:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			state, _ := newSessionState([]WorkflowNode{
				{ID: nodeID, ProcessID: "test-gate-expr", Status: NodeRunning, NodeType: NodeTypeProcess},
			}, nil)
			state.exec.NodeOutputs = tt.nodeOutputs

			gate := &IterationGate{
				Kind:       GateExpression,
				CustomExpr: tt.customExpr,
			}

			e := NewExecutor(nil, func(string, interface{}) {})
			hit, _ := e.checkGate(state, gate, nodeID, 1)
			assert.Equal(t, tt.wantHit, hit)
		})
	}
}

// ── nil gate ─────────────────────────────────────────────────────────────────

// TestCheckGateNil (AC-7 companion)
// Nil gate means single-round process: checkGate returns (true, "no gate; single round").
func TestCheckGateNil(t *testing.T) {
	t.Parallel()

	const nodeID = "n1"
	state, _ := newSessionState([]WorkflowNode{
		{ID: nodeID, ProcessID: "test-gate-nil", Status: NodeRunning, NodeType: NodeTypeProcess},
	}, nil)

	e := NewExecutor(nil, func(string, interface{}) {})
	hit, reason := e.checkGate(state, nil, nodeID, 1)
	assert.True(t, hit, "nil gate must return hit=true")
	assert.Equal(t, "no gate; single round", reason)
}

// ── containsToken ─────────────────────────────────────────────────────────────

// TestContainsTokenCaseInsensitiveTrim
// containsToken must compare case-insensitively after trimming whitespace from both sides.
func TestContainsTokenCaseInsensitiveTrim(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		answer string
		tokens []string
		want   bool
	}{
		{
			name:   "DONE upper-case matches lower-case token",
			answer: "DONE",
			tokens: []string{"done"},
			want:   true,
		},
		{
			name:   "padded finish matches finish token",
			answer: "  finish  ",
			tokens: []string{"finish"},
			want:   true,
		},
		{
			name:   "non-matching answer returns false",
			answer: "foo",
			tokens: []string{"bar"},
			want:   false,
		},
		{
			name:   "empty tokens slice always returns false",
			answer: "done",
			tokens: []string{},
			want:   false,
		},
		{
			name:   "nil tokens returns false",
			answer: "done",
			tokens: nil,
			want:   false,
		},
		{
			name:   "empty answer with non-empty tokens returns false",
			answer: "",
			tokens: []string{"done"},
			want:   false,
		},
		{
			name:   "mixed-case token matches mixed-case answer",
			answer: "Finish",
			tokens: []string{"FINISH"},
			want:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// containsToken does not exist yet — RED phase.
			got := containsToken(tt.tokens, tt.answer)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ── AC-7: iterationInput helper ───────────────────────────────────────────────

// TestIterationInputHelper (AC-7)
// iterationInput returns the single InputSpec that represents the recurring
// per-round user prompt: Source=InputFromUser, Prompt!="", Shape!="", Required=false.
//
// ProcessDef A — brainstorm-like: topic (required user), approach (required choice),
//
//	technique (registry), round-response (user free optional Prompt!="" Shape!="").
//	iterationInput returns round-response.
//
// ProcessDef B — product-brief-like: mode (required choice), stage-response
//
//	(optional user free Prompt!="" Shape!=""), final-approval (required approval).
//	iterationInput returns stage-response.
//
// ProcessDef C — all required user inputs, none match iteration pattern.
//
//	iterationInput returns (InputSpec{}, false).
//
// ProcessDef D — party-like: topic (required), message (optional user free Prompt!="").
//
//	iterationInput returns message.
func TestIterationInputHelper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		proc      ProcessDef
		wantID    string
		wantFound bool
	}{
		{
			name: "ProcessDef A brainstorm-like: returns round-response",
			proc: ProcessDef{
				ID:   "test-iter-a",
				Mode: InteractIterative,
				InputSpecs: []InputSpec{
					{
						ID:       "topic",
						Source:   InputFromUser,
						Shape:    ShapeFree,
						Required: true,
						Prompt:   "What is the topic?",
					},
					{
						ID:       "approach",
						Source:   InputFromUser,
						Shape:    ShapeChoice,
						Required: true,
						Prompt:   "Which approach?",
						Options:  []string{"SCAMPER", "TRIZ"},
					},
					{
						ID:         "technique",
						Source:     InputFromRegistry,
						OptionsRef: "registry:/some/path.csv#technique",
					},
					{
						ID:       "round-response",
						Source:   InputFromUser,
						Shape:    ShapeFree,
						Required: false,
						Prompt:   "Your response for this round:",
					},
				},
			},
			wantID:    "round-response",
			wantFound: true,
		},
		{
			name: "ProcessDef B product-brief-like guided: returns stage-response",
			proc: ProcessDef{
				ID:   "test-iter-b",
				Mode: InteractGuided,
				InputSpecs: []InputSpec{
					{
						ID:       "mode",
						Source:   InputFromUser,
						Shape:    ShapeChoice,
						Required: true,
						Prompt:   "Select mode:",
						Options:  []string{"guided", "free"},
					},
					{
						ID:       "stage-response",
						Source:   InputFromUser,
						Shape:    ShapeFree,
						Required: false,
						Prompt:   "Your input for this stage:",
					},
					{
						ID:       "final-approval",
						Source:   InputFromUser,
						Shape:    ShapeApproval,
						Required: true,
						Prompt:   "Approve the output?",
					},
				},
			},
			wantID:    "stage-response",
			wantFound: true,
		},
		{
			name: "ProcessDef C all required user inputs: returns false",
			proc: ProcessDef{
				ID:   "test-iter-c",
				Mode: InteractGuided,
				InputSpecs: []InputSpec{
					{
						ID:       "topic",
						Source:   InputFromUser,
						Shape:    ShapeFree,
						Required: true,
						Prompt:   "Topic?",
					},
					{
						ID:       "style",
						Source:   InputFromUser,
						Shape:    ShapeChoice,
						Required: true,
						Prompt:   "Style?",
					},
				},
			},
			wantID:    "",
			wantFound: false,
		},
		{
			name: "ProcessDef D party-like: returns message",
			proc: ProcessDef{
				ID:   "test-iter-d",
				Mode: InteractParty,
				InputSpecs: []InputSpec{
					{
						ID:       "topic",
						Source:   InputFromUser,
						Shape:    ShapeFree,
						Required: true,
						Prompt:   "Party topic?",
					},
					{
						ID:       "message",
						Source:   InputFromUser,
						Shape:    ShapeFree,
						Required: false,
						Prompt:   "Your message:",
					},
				},
			},
			wantID:    "message",
			wantFound: true,
		},
		{
			name: "ProcessDef with optional user spec but empty Prompt: not matched",
			proc: ProcessDef{
				ID:   "test-iter-e",
				Mode: InteractIterative,
				InputSpecs: []InputSpec{
					{
						ID:       "optional-no-prompt",
						Source:   InputFromUser,
						Shape:    ShapeFree,
						Required: false,
						Prompt:   "", // empty Prompt — does not match iteration spec
					},
				},
			},
			wantID:    "",
			wantFound: false,
		},
		{
			name: "ProcessDef with optional user spec but empty Shape: not matched",
			proc: ProcessDef{
				ID:   "test-iter-f",
				Mode: InteractIterative,
				InputSpecs: []InputSpec{
					{
						ID:       "optional-no-shape",
						Source:   InputFromUser,
						Shape:    "", // empty Shape — does not match iteration spec
						Required: false,
						Prompt:   "Your message:",
					},
				},
			},
			wantID:    "",
			wantFound: false,
		},
		{
			name: "empty InputSpecs: returns false",
			proc: ProcessDef{
				ID:         "test-iter-g",
				Mode:       InteractIterative,
				InputSpecs: []InputSpec{},
			},
			wantID:    "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// iterationInput does not exist yet — RED phase.
			spec, found := tt.proc.iterationInput()
			assert.Equal(t, tt.wantFound, found, "found mismatch")
			if tt.wantFound {
				require.True(t, found, "expected iterationInput to return true")
				assert.Equal(t, tt.wantID, spec.ID, "spec.ID mismatch")
			} else {
				assert.Equal(t, InputSpec{}, spec, "not-found must return zero InputSpec")
			}
		})
	}
}
