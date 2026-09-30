package bmad

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractRegex_WithCaptureGroup(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pattern string
		want    string
	}{
		{
			name:    "version capture group",
			input:   "Released version: 1.2.3",
			pattern: `version: (\S+)`,
			want:    "1.2.3",
		},
		{
			name:    "parenthesized group",
			input:   "error code: (42)",
			pattern: `code: \((\d+)\)`,
			want:    "42",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRegex(tt.input, tt.pattern)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExtractRegex_WithoutCaptureGroup(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pattern string
		want    string
	}{
		{
			name:    "semver match",
			input:   "Version 2.5.1",
			pattern: `\d+\.\d+\.\d+`,
			want:    "2.5.1",
		},
		{
			name:    "word match",
			input:   "hello world",
			pattern: `\w+`,
			want:    "hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRegex(tt.input, tt.pattern)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestExtractRegex_NoMatch(t *testing.T) {
	result := extractRegex("nothing here", "NOTFOUND")
	assert.Equal(t, "", result)
}

func TestExtractRegex_InvalidRegex(t *testing.T) {
	result := extractRegex("some text", "[invalid")
	assert.Equal(t, "", result)
}

func TestExtractLines_Range(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5"
	result := extractLines(input, "2-4")
	assert.Equal(t, "line2\nline3\nline4", result)
}

func TestExtractLines_LastN(t *testing.T) {
	input := "line1\nline2\nline3\nline4\nline5"
	result := extractLines(input, "-3")
	assert.Equal(t, "line3\nline4\nline5", result)
}

func TestExtractLines_Single(t *testing.T) {
	input := "line1\nline2\nline3"
	result := extractLines(input, "1")
	assert.Equal(t, "line1", result)
}

func TestExtractLines_OutOfRange(t *testing.T) {
	input := "line1\nline2"
	result := extractLines(input, "1-100")
	assert.Equal(t, "line1\nline2", result, "should clamp to available lines")
}

func TestExtractLines_InvalidPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
	}{
		{name: "non-numeric", pattern: "abc"},
		{name: "zero start", pattern: "0"},
		{name: "negative start", pattern: "-0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLines("line1\nline2", tt.pattern)
			assert.Equal(t, "", result)
		})
	}
}

func TestExtractLines_StartBeyondLength(t *testing.T) {
	input := "line1\nline2"
	result := extractLines(input, "10")
	assert.Equal(t, "", result, "start beyond line count should return empty")
}

func TestTransformNode_RegexExtraction(t *testing.T) {
	// Workflow: A(process) -> T(transform, regex) -> B(process).
	// A produces "version: 3.4.5". T extracts "3.4.5". B should complete.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"process-a": "build version: 3.4.5 deployed",
	}))

	wf := WorkflowDef{
		ID:   "wf-transform-regex",
		Name: "Transform Regex",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "T", NodeType: NodeTypeTransform, Label: "Extract Version", Config: map[string]string{
				"sourceNode":     "A",
				"extractType":    "regex",
				"extractPattern": `version: (\S+)`,
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
			{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Process B", NodeType: NodeTypeProcess,
				Position: Position{X: 500, Y: 0}, Status: NodePending, Config: map[string]string{}},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "T"},
			{ID: "e2", Source: "T", Target: "B"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-transform-regex", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)

	// Verify transform extracted the version.
	assert.Equal(t, "3.4.5", ex.NodeOutputs["T"])

	// All nodes should complete.
	for _, n := range ex.Nodes {
		assert.Equal(t, NodeComplete, n.Status, "node %s should be complete", n.ID)
	}
}

func TestTransformNode_LinesExtraction(t *testing.T) {
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"process-a": "header\nline2\nline3\nline4\nfooter",
	}))

	wf := WorkflowDef{
		ID:   "wf-transform-lines",
		Name: "Transform Lines",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "T", NodeType: NodeTypeTransform, Label: "Extract Lines", Config: map[string]string{
				"sourceNode":     "A",
				"extractType":    "lines",
				"extractPattern": "2-4",
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "T"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-transform-lines", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, "line2\nline3\nline4", ex.NodeOutputs["T"])
}

func TestTransformNode_MissingSource(t *testing.T) {
	// Transform with nonexistent sourceNode -> empty output, completes.
	h := newHarness(t)
	h.executor.SetCommandRunner(successRunner())

	wf := WorkflowDef{
		ID:   "wf-transform-nosrc",
		Name: "Transform Missing Source",
		Nodes: []WorkflowNode{
			{ID: "T", NodeType: NodeTypeTransform, Label: "Orphan Transform", Config: map[string]string{
				"sourceNode":     "nonexistent",
				"extractType":    "regex",
				"extractPattern": `(\d+)`,
			}, Position: Position{X: 0, Y: 0}, Status: NodePending},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-transform-nosrc", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, ExecComplete, ex.Status)
	assert.Equal(t, NodeComplete, ex.Nodes[0].Status)
	assert.Equal(t, "", ex.NodeOutputs["T"], "missing source should produce empty output")
}

func TestTransformNode_Passthrough(t *testing.T) {
	// Unknown extractType should passthrough source output.
	h := newHarness(t)
	h.executor.SetCommandRunner(conditionRunner(map[string]string{
		"process-a": "raw output data",
	}))

	wf := WorkflowDef{
		ID:   "wf-transform-pass",
		Name: "Transform Passthrough",
		Nodes: []WorkflowNode{
			{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", NodeType: NodeTypeProcess,
				Position: Position{X: 0, Y: 0}, Status: NodePending, Config: map[string]string{}},
			{ID: "T", NodeType: NodeTypeTransform, Label: "Passthrough", Config: map[string]string{
				"sourceNode":     "A",
				"extractType":    "unknown",
				"extractPattern": "",
			}, Position: Position{X: 250, Y: 0}, Status: NodePending},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "A", Target: "T"},
		},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
	require.NoError(t, h.storage.SaveWorkflow(wf))

	exec, err := h.executor.StartWorkflow(context.Background(), "wf-transform-pass", "/tmp/repo", "sonnet")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		ex, _ := h.executor.GetExecution(exec.ID)
		return ex != nil && ex.Status == ExecComplete
	}, 5*time.Second, 50*time.Millisecond)

	ex, err := h.executor.GetExecution(exec.ID)
	require.NoError(t, err)
	assert.Equal(t, "raw output data", ex.NodeOutputs["T"], "unknown extractType should passthrough")
}

func TestBuildContextStringV3_IncludesTransformData(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: autonomousProcessFixtureID, Label: "Process A", Status: NodeComplete, Config: map[string]string{}},
		{ID: "T", NodeType: NodeTypeTransform, Label: "Version Extract", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "3.4.5",
	}

	proc, _ := ProcessByID("bmad-create-prd")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", nil, "B")
	assert.Contains(t, result, "Version Extract")
	assert.Contains(t, result, "3.4.5")
}

func TestBuildContextStringV3_TruncatesLongData(t *testing.T) {
	longData := strings.Repeat("X", 3000)
	nodes := []WorkflowNode{
		{ID: "T", NodeType: NodeTypeTransform, Label: "Big Transform", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": longData,
	}

	proc, _ := ProcessByID("bmad-domain-research")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", nil, "B")
	assert.Contains(t, result, "Big Transform")
	// The data portion should be capped at 2000 chars.
	assert.LessOrEqual(t, len(result), 2100, "result should not contain full 3000-char data")
	assert.NotContains(t, result, longData, "full long data should be truncated")
}

func TestBuildContextStringV3_IncludesArtifactMatching(t *testing.T) {
	// bmad-create-prd has Inputs: ["product-brief"] and bmad-product-brief has Outputs: ["product-brief"].
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-product-brief", Label: "Product Brief", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Create PRD", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-create-prd")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", nil, "B")
	assert.Contains(t, result, "upstream process")
	assert.Contains(t, result, "product-brief")
}

func TestBuildContextStringV3_EmptyTransformData(t *testing.T) {
	// Transform node with empty output should be excluded.
	nodes := []WorkflowNode{
		{ID: "T", NodeType: NodeTypeTransform, Label: "Empty Transform", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "",
	}

	proc, _ := ProcessByID("bmad-domain-research")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", nil, "B")
	assert.NotContains(t, result, "Empty Transform", "empty transform data should not appear")
}

func TestBuildContextStringV3_SkipsNonCompleteTransforms(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "T", NodeType: NodeTypeTransform, Label: "Pending Transform", Status: NodePending, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Process B", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "some data",
	}

	proc, _ := ProcessByID("bmad-domain-research")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", nil, "B")
	assert.NotContains(t, result, "Pending Transform", "non-complete transform should not appear")
}

func TestBuildContextStringV3_FilePathResolution(t *testing.T) {
	tests := []struct {
		name        string
		setupFiles  bool   // whether to create the artifact file on disk
		procID      string // downstream process
		upstreamID  string // upstream process
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:       "file_exists",
			setupFiles: true,
			procID:     "bmad-create-architecture", // Inputs: ["PRD.md"] — concrete contract, not autonomous-fixture
			upstreamID: "bmad-create-prd",          // Outputs: ["PRD.md"]
			wantContain: []string{
				"Read the artifact 'PRD.md' from file",
				"planning-artifacts/PRD.md",
			},
			wantAbsent: []string{
				"was not found",
			},
		},
		{
			name:       "file_missing",
			setupFiles: false,
			procID:     "bmad-create-architecture", // Inputs: ["PRD.md"] — concrete contract
			upstreamID: "bmad-create-prd",          // Outputs: ["PRD.md"]
			wantContain: []string{
				"should have produced 'PRD.md'",
				"but it was not found",
			},
			wantAbsent: []string{
				"Read the artifact",
			},
		},
		{
			name:       "unmapped_artifact",
			setupFiles: false,
			procID:     "bmad-domain-research", // Inputs: [] (empty)
			upstreamID: "bmad-domain-research", // Outputs: ["brainstorm-notes"]
			// brainstorming has no inputs, so no artifact matching at all
			wantContain: []string{},
			wantAbsent:  []string{"upstream process"},
		},
		{
			name:        "no_inputs",
			setupFiles:  false,
			procID:      "bmad-domain-research", // Inputs: []
			upstreamID:  "",
			wantContain: []string{},
			wantAbsent:  []string{"upstream", "artifact"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.setupFiles {
				// Create the artifact file.
				proc, ok := ProcessByID(tt.upstreamID)
				require.True(t, ok)
				for _, output := range proc.Outputs {
					p := ResolveArtifactPath(output, tmpDir)
					if p == "" {
						continue
					}
					require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
					require.NoError(t, os.WriteFile(p, []byte("test content"), 0o644))
				}
			}

			var nodes []WorkflowNode
			if tt.upstreamID != "" {
				nodes = append(nodes, WorkflowNode{
					ID:        "upstream",
					ProcessID: tt.upstreamID,
					Label:     "Upstream",
					Status:    NodeComplete,
					Config:    map[string]string{},
				})
			}
			nodes = append(nodes, WorkflowNode{
				ID:        "downstream",
				ProcessID: tt.procID,
				Label:     "Downstream",
				Status:    NodePending,
				Config:    map[string]string{},
			})

			nodeIndex := buildNodeIndex(nodes)
			nodeOutputs := map[string]string{}

			proc, ok := ProcessByID(tt.procID)
			require.True(t, ok)

			result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, tmpDir, nil, "downstream")

			for _, want := range tt.wantContain {
				assert.Contains(t, result, want, "expected result to contain %q", want)
			}
			for _, absent := range tt.wantAbsent {
				assert.NotContains(t, result, absent, "expected result NOT to contain %q", absent)
			}
		})
	}
}

func TestBuildContextStringV3_TransformDataPreservedWithRepoPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the product-brief artifact file so we get file-path message.
	prdPath := ResolveArtifactPath("product-brief", tmpDir)
	require.NoError(t, os.MkdirAll(filepath.Dir(prdPath), 0o755))
	require.NoError(t, os.WriteFile(prdPath, []byte("brief content"), 0o644))

	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-product-brief", Label: "Product Brief", Status: NodeComplete, Config: map[string]string{}},
		{ID: "T", NodeType: NodeTypeTransform, Label: "Version Extract", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Create PRD", Status: NodePending, Config: map[string]string{}},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{
		"T": "3.4.5",
	}

	proc, _ := ProcessByID("bmad-create-prd")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, tmpDir, nil, "B")

	// Should contain both file-path info AND transform data.
	assert.Contains(t, result, "Read the artifact 'product-brief' from file")
	assert.Contains(t, result, "Version Extract")
	assert.Contains(t, result, "3.4.5")
}

func TestBuildContextStringV3_EdgeBasedContext_EmptyInputs(t *testing.T) {
	// Code Review has Inputs: [] but is edge-connected to Dev Story (Outputs: ["code", "tests"]).
	// Edge-based context should pass upstream outputs even without artifact name matching.
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-dev-story", Label: "Develop Story", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-code-review", Label: "Code Review", Status: NodePending, Config: map[string]string{}},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-code-review")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", edges, "B")
	assert.Contains(t, result, "Develop Story")
	assert.Contains(t, result, "code")
	assert.Contains(t, result, "tests")
}

func TestBuildContextStringV3_EdgeBasedContext_WithFileResolution(t *testing.T) {
	// bmad-domain-research outputs "domain-research"; edge-based context
	// mentions both the upstream label and the artifact name.
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-domain-research", Label: "Domain Research", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-product-brief", Label: "Product Brief", Status: NodePending, Config: map[string]string{}},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-product-brief")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", edges, "B")
	assert.Contains(t, result, "Domain Research")
	assert.Contains(t, result, "domain-research")
}

func TestBuildContextStringV3_EdgeBasedContext_NoDuplicates(t *testing.T) {
	// When artifact name matching already covers an output, edge-based context
	// should NOT duplicate it.
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-product-brief", Label: "Product Brief", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: autonomousProcessFixtureID, Label: "Create PRD", Status: NodePending, Config: map[string]string{}},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-create-prd") // Inputs: ["product-brief"]
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", edges, "B")
	// The artifact should be mentioned by exactly one context message, not duplicated
	// by both artifact matching and edge-based context. Count the message prefix
	// (not the raw substring, which also appears in file paths).
	assert.Equal(t, 1, strings.Count(result, "'product-brief'"), "product-brief should be referenced in exactly one context message")
}

func TestBuildContextStringV3_EdgeBasedContext_NonConnectedNodeIgnored(t *testing.T) {
	// Node C is complete but NOT edge-connected to B. Its outputs should NOT
	// appear in edge-based context (only artifact name matching can pick them up).
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-brainstorming", Label: "Brainstorming", Status: NodeComplete, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-code-review", Label: "Code Review", Status: NodePending, Config: map[string]string{}},
	}
	// No edge connecting A to B.
	edges := []WorkflowEdge{}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-code-review")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", edges, "B")
	assert.Empty(t, result, "no context without edges or matching inputs")
}

func TestBuildContextStringV3_EdgeBasedContext_SkipsNonCompleteUpstream(t *testing.T) {
	// Upstream node is connected by edge but not yet complete.
	nodes := []WorkflowNode{
		{ID: "A", ProcessID: "bmad-dev-story", Label: "Develop Story", Status: NodeRunning, Config: map[string]string{}},
		{ID: "B", ProcessID: "bmad-code-review", Label: "Code Review", Status: NodePending, Config: map[string]string{}},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "A", Target: "B"},
	}
	nodeIndex := buildNodeIndex(nodes)
	nodeOutputs := map[string]string{}

	proc, _ := ProcessByID("bmad-code-review")
	result := buildContextStringV3(proc, nodes, nodeIndex, nodeOutputs, "", edges, "B")
	assert.Empty(t, result, "running upstream should not produce context")
}
