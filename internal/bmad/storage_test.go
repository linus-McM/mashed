package bmad

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()
	s, err := NewStorage(t.TempDir())
	require.NoError(t, err)
	return s
}

func sampleWorkflow(id string) WorkflowDef {
	return WorkflowDef{
		ID:          id,
		Name:        "Test Workflow " + id,
		Description: "A test workflow",
		Nodes: []WorkflowNode{
			{
				ID:        "n1",
				ProcessID: "bmad-brainstorming",
				Label:     "Brainstorm",
				Position:  Position{X: 0, Y: 200},
				Status:    NodePending,
				Config:    map[string]string{},
			},
			{
				ID:        "n2",
				ProcessID: "bmad-create-prd",
				Label:     "Create PRD",
				Position:  Position{X: 250, Y: 200},
				Status:    NodePending,
				Config:    map[string]string{},
			},
		},
		Edges:     []WorkflowEdge{{ID: "e1", Source: "n1", Target: "n2"}},
		CreatedAt: "2026-04-07T00:00:00Z",
		UpdatedAt: "2026-04-07T00:00:00Z",
	}
}

func sampleAgent(id string) BmadAgentConfig {
	return BmadAgentConfig{
		ID:        id,
		Name:      "Test Agent " + id,
		Role:      RoleDeveloper,
		Persona:   "A helpful developer agent",
		Skills:    []string{"bmad-dev-story", "bmad-code-review"},
		Model:     "claude-opus-4-6",
		CreatedAt: "2026-04-07T00:00:00Z",
	}
}

// ── Workflow CRUD ──

func TestSaveAndLoadWorkflow(t *testing.T) {
	s := newTestStorage(t)
	wf := sampleWorkflow("test-wf-1")

	require.NoError(t, s.SaveWorkflow(wf))

	loaded, err := s.LoadWorkflow("test-wf-1")
	require.NoError(t, err)
	assert.Equal(t, wf, loaded)
}

func TestLoadWorkflow_NotFound(t *testing.T) {
	s := newTestStorage(t)
	_, err := s.LoadWorkflow("nonexistent")
	assert.True(t, errors.Is(err, ErrWorkflowNotFound))
}

func TestListWorkflows(t *testing.T) {
	s := newTestStorage(t)
	for _, id := range []string{"wf-a", "wf-b", "wf-c"} {
		require.NoError(t, s.SaveWorkflow(sampleWorkflow(id)))
	}

	list, err := s.ListWorkflows()
	require.NoError(t, err)
	assert.Len(t, list, 3)
}

func TestDeleteWorkflow(t *testing.T) {
	s := newTestStorage(t)
	require.NoError(t, s.SaveWorkflow(sampleWorkflow("to-delete")))

	require.NoError(t, s.DeleteWorkflow("to-delete"))

	_, err := s.LoadWorkflow("to-delete")
	assert.True(t, errors.Is(err, ErrWorkflowNotFound))
}

func TestDeleteWorkflow_NotFound(t *testing.T) {
	s := newTestStorage(t)
	err := s.DeleteWorkflow("nonexistent")
	assert.True(t, errors.Is(err, ErrWorkflowNotFound))
}

// ── Path traversal rejection ──

func TestSaveWorkflow_PathTraversal(t *testing.T) {
	s := newTestStorage(t)
	tests := []struct {
		name string
		id   string
	}{
		{"dot-dot-slash", "../../../etc/passwd"},
		{"slash", "foo/bar"},
		{"empty", ""},
		{"dot-dot", ".."},
		{"space", "has space"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := sampleWorkflow("valid")
			wf.ID = tt.id
			err := s.SaveWorkflow(wf)
			assert.True(t, errors.Is(err, ErrInvalidID), "ID %q should be rejected", tt.id)
		})
	}
}

func TestLoadWorkflow_InvalidID(t *testing.T) {
	s := newTestStorage(t)
	_, err := s.LoadWorkflow("../../../etc/passwd")
	assert.True(t, errors.Is(err, ErrInvalidID))
}

func TestDeleteWorkflow_InvalidID(t *testing.T) {
	s := newTestStorage(t)
	err := s.DeleteWorkflow("../hack")
	assert.True(t, errors.Is(err, ErrInvalidID))
}

// ── Malformed JSON resilience ──

func TestListWorkflows_SkipsMalformedJSON(t *testing.T) {
	s := newTestStorage(t)
	require.NoError(t, s.SaveWorkflow(sampleWorkflow("good-wf")))

	// Write a malformed JSON file directly
	badPath := filepath.Join(s.workflowDir, "bad.json")
	require.NoError(t, os.WriteFile(badPath, []byte("{invalid json!!!"), 0644))

	list, err := s.ListWorkflows()
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "good-wf", list[0].ID)
}

func TestListWorkflows_SkipsNonJSON(t *testing.T) {
	s := newTestStorage(t)
	require.NoError(t, s.SaveWorkflow(sampleWorkflow("good")))

	// Write a non-JSON file
	require.NoError(t, os.WriteFile(filepath.Join(s.workflowDir, "readme.txt"), []byte("hello"), 0644))

	list, err := s.ListWorkflows()
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

// ── Agent CRUD ──

func TestSaveAndListAgents(t *testing.T) {
	s := newTestStorage(t)
	ag := sampleAgent("custom-dev-1")
	require.NoError(t, s.SaveAgent(ag))

	agents, err := s.ListAgents()
	require.NoError(t, err)
	require.Len(t, agents, 1)
	assert.Equal(t, ag, agents[0])
}

func TestDeleteAgent(t *testing.T) {
	s := newTestStorage(t)
	require.NoError(t, s.SaveAgent(sampleAgent("to-delete")))

	require.NoError(t, s.DeleteAgent("to-delete"))

	agents, err := s.ListAgents()
	require.NoError(t, err)
	assert.Empty(t, agents)
}

func TestDeleteAgent_NotFound(t *testing.T) {
	s := newTestStorage(t)
	err := s.DeleteAgent("nonexistent")
	assert.True(t, errors.Is(err, ErrAgentNotFound))
}

func TestSaveAgent_InvalidID(t *testing.T) {
	s := newTestStorage(t)
	ag := sampleAgent("valid")
	ag.ID = "../hack"
	err := s.SaveAgent(ag)
	assert.True(t, errors.Is(err, ErrInvalidID))
}

func TestDeleteAgent_InvalidID(t *testing.T) {
	s := newTestStorage(t)
	err := s.DeleteAgent("../../etc")
	assert.True(t, errors.Is(err, ErrInvalidID))
}

// ── NewStorage creates directories ──

func TestNewStorage_CreatesDirectories(t *testing.T) {
	base := filepath.Join(t.TempDir(), "nested", "deep")
	s, err := NewStorage(base)
	require.NoError(t, err)

	info, err := os.Stat(s.workflowDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	info, err = os.Stat(s.agentDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// ── Overwrite existing workflow ──

func TestSaveWorkflow_Overwrite(t *testing.T) {
	s := newTestStorage(t)
	wf := sampleWorkflow("overwrite-me")
	require.NoError(t, s.SaveWorkflow(wf))

	wf.Name = "Updated Name"
	require.NoError(t, s.SaveWorkflow(wf))

	loaded, err := s.LoadWorkflow("overwrite-me")
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", loaded.Name)
}
