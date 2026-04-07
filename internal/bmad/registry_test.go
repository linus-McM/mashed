package bmad

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllProcesses_CountAndFields(t *testing.T) {
	procs := AllProcesses()
	require.GreaterOrEqual(t, len(procs), 25, "registry must contain at least 25 processes")

	for _, p := range procs {
		assert.NotEmpty(t, p.ID, "ID must not be empty")
		assert.NotEmpty(t, p.Name, "Name must not be empty")
		assert.NotEmpty(t, string(p.Phase), "Phase must not be empty")
		assert.NotEmpty(t, string(p.AgentRole), "AgentRole must not be empty")
		assert.NotEmpty(t, p.SkillName, "SkillName must not be empty")
		assert.NotEmpty(t, p.Description, "Description must not be empty")
		assert.NotEmpty(t, p.ModuleID, "ModuleID must not be empty")
		assert.NotEmpty(t, p.Version, "Version must not be empty")
		assert.NotNil(t, p.Inputs, "Inputs must not be nil")
		assert.NotNil(t, p.Outputs, "Outputs must not be nil")
	}
}

func TestAllProcesses_UniqueIDs(t *testing.T) {
	procs := AllProcesses()
	seen := make(map[string]bool, len(procs))
	for _, p := range procs {
		assert.True(t, strings.HasPrefix(p.ID, "bmad-"), "ID %q must start with bmad-", p.ID)
		assert.False(t, seen[p.ID], "duplicate ID: %s", p.ID)
		seen[p.ID] = true
	}
	assert.Equal(t, len(procs), len(seen), "set size must equal slice length")
}

func TestProcessesByPhase(t *testing.T) {
	tests := []struct {
		phase BmadPhase
		count int
	}{
		{PhaseAnalysis, 5},
		{PhasePlanning, 4},
		{PhaseSolutioning, 4},
		{PhaseImplementation, 8},
		{PhaseSupport, 4},
	}

	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			got := ProcessesByPhase(tt.phase)
			assert.Len(t, got, tt.count)
			for _, p := range got {
				assert.Equal(t, tt.phase, p.Phase)
			}
		})
	}
}

func TestProcessesByPhase_UnionEqualsAll(t *testing.T) {
	phases := []BmadPhase{PhaseAnalysis, PhasePlanning, PhaseSolutioning, PhaseImplementation, PhaseSupport}
	var union []ProcessDef
	for _, ph := range phases {
		union = append(union, ProcessesByPhase(ph)...)
	}
	assert.Equal(t, len(AllProcesses()), len(union), "union of all phases must equal AllProcesses")
}

func TestProcessByID_Existing(t *testing.T) {
	p, ok := ProcessByID("bmad-create-prd")
	require.True(t, ok)
	assert.Equal(t, "Create PRD", p.Name)
	assert.Equal(t, PhasePlanning, p.Phase)
	assert.Equal(t, RolePM, p.AgentRole)
}

func TestProcessByID_Missing(t *testing.T) {
	_, ok := ProcessByID("nonexistent")
	assert.False(t, ok)
}

func TestProcessesByModule_Core(t *testing.T) {
	got := ProcessesByModule("core")
	assert.Equal(t, len(AllProcesses()), len(got), "all initial processes belong to core")
}

func TestProcessesByModule_Unknown(t *testing.T) {
	got := ProcessesByModule("nonexistent")
	assert.Empty(t, got)
}

func TestAllProcesses_ReturnsCopy(t *testing.T) {
	a := AllProcesses()
	b := AllProcesses()
	a[0].Name = "mutated"
	assert.NotEqual(t, a[0].Name, b[0].Name, "AllProcesses must return independent copies")
}

func TestProcessDef_JSONRoundTrip(t *testing.T) {
	original := ProcessDef{
		ID:          "bmad-create-architecture",
		Name:        "Create Architecture",
		Phase:       PhaseSolutioning,
		AgentRole:   RoleArchitect,
		SkillName:   "bmad-create-architecture",
		Description: "Design the system architecture based on PRD requirements.",
		Inputs:      []string{"PRD.md"},
		Outputs:     []string{"architecture.md"},
		ModuleID:    "core",
		Version:     "1.0.0",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded ProcessDef
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, decoded)
}

func TestWorkflowDef_JSONRoundTrip(t *testing.T) {
	original := WorkflowDef{
		ID:          "wf-1",
		Name:        "Test Workflow",
		Description: "A test workflow",
		Nodes: []WorkflowNode{
			{
				ID:        "n1",
				ProcessID: "bmad-brainstorming",
				Label:     "Brainstorm",
				Position:  Position{X: 100.5, Y: 200.75},
				Status:    NodePending,
				Config:    map[string]string{"key": "value"},
			},
			{
				ID:        "n2",
				ProcessID: "bmad-create-prd",
				Label:     "Create PRD",
				Position:  Position{X: 300, Y: 200.75},
				Status:    NodeRunning,
				Config:    map[string]string{},
			},
			{
				ID:        "n3",
				ProcessID: "bmad-create-architecture",
				Label:     "Architecture",
				Position:  Position{X: 500, Y: 200.75},
				Status:    NodeComplete,
				Config:    map[string]string{},
			},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", Source: "n1", Target: "n2"},
			{ID: "e2", Source: "n2", Target: "n3"},
		},
		IsTemplate: false,
		CreatedAt:  "2026-04-07T00:00:00Z",
		UpdatedAt:  "2026-04-07T00:00:00Z",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded WorkflowDef
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, decoded)
}

func TestWorkflowDef_OmitsEmptyTemplateID(t *testing.T) {
	wf := WorkflowDef{
		ID:         "wf-1",
		Name:       "Test",
		IsTemplate: false,
		TemplateID: "",
		Nodes:      []WorkflowNode{},
		Edges:      []WorkflowEdge{},
	}
	data, err := json.Marshal(wf)
	require.NoError(t, err)
	assert.NotContains(t, string(data), "templateId")
}

func TestWorkflowNode_JSONRoundTrip(t *testing.T) {
	original := WorkflowNode{
		ID:         "node-1",
		ProcessID:  "bmad-dev-story",
		Label:      "Dev Story",
		Position:   Position{X: 123.456, Y: 789.012},
		Status:     NodeFailed,
		Config:     map[string]string{"branch": "feature/test"},
		TmuxTarget: "session:0.1",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded WorkflowNode
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, decoded)
}
