package bmad

import (
	"encoding/json"
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
				ID: "n1",
				// bmad-domain-research is autonomous (Mode == "") so this fixture
				// still exercises the legacy non-interactive flow. S7 moved
				// bmad-brainstorming to InteractIterative.
				ProcessID: "bmad-domain-research",
				Label:     "Domain Research",
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

// ── Helper with RepoPath ──

func sampleWorkflowWithRepo(id, repoPath string) WorkflowDef {
	wf := sampleWorkflow(id)
	wf.RepoPath = repoPath
	return wf
}

// ── ListWorkflowsByRepo ──

func TestListWorkflowsByRepo(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(s *Storage)
		repoPath  string
		wantCount int
		wantIDs   []string
	}{
		{
			name: "filters by repo path correctly",
			setup: func(s *Storage) {
				require.NoError(t, s.SaveWorkflow(sampleWorkflowWithRepo("wf-a1", "/repo-a")))
				require.NoError(t, s.SaveWorkflow(sampleWorkflowWithRepo("wf-a2", "/repo-a")))
				require.NoError(t, s.SaveWorkflow(sampleWorkflowWithRepo("wf-b1", "/repo-b")))
			},
			repoPath:  "/repo-a",
			wantCount: 2,
			wantIDs:   []string{"wf-a1", "wf-a2"},
		},
		{
			name: "empty RepoPath workflows excluded from repo filter",
			setup: func(s *Storage) {
				require.NoError(t, s.SaveWorkflow(sampleWorkflowWithRepo("wf-r1", "/repo-a")))
				require.NoError(t, s.SaveWorkflow(sampleWorkflow("wf-nopath"))) // no RepoPath
			},
			repoPath:  "/repo-a",
			wantCount: 1,
			wantIDs:   []string{"wf-r1"},
		},
		{
			name: "trailing slash normalization",
			setup: func(s *Storage) {
				require.NoError(t, s.SaveWorkflow(sampleWorkflowWithRepo("wf-trail", "/repo-a/")))
				require.NoError(t, s.SaveWorkflow(sampleWorkflowWithRepo("wf-noslash", "/repo-a")))
			},
			repoPath:  "/repo-a/",
			wantCount: 2,
			wantIDs:   []string{"wf-noslash", "wf-trail"},
		},
		{
			name: "no matches returns empty slice",
			setup: func(s *Storage) {
				require.NoError(t, s.SaveWorkflow(sampleWorkflowWithRepo("wf-x", "/repo-x")))
			},
			repoPath:  "/repo-y",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestStorage(t)
			tt.setup(s)

			got, err := s.ListWorkflowsByRepo(tt.repoPath)
			require.NoError(t, err)
			assert.Len(t, got, tt.wantCount)

			if tt.wantIDs != nil {
				var gotIDs []string
				for _, wf := range got {
					gotIDs = append(gotIDs, wf.ID)
				}
				assert.ElementsMatch(t, tt.wantIDs, gotIDs)
			}
		})
	}
}

func TestListWorkflowsByRepo_ListWorkflowsStillReturnsAll(t *testing.T) {
	s := newTestStorage(t)
	require.NoError(t, s.SaveWorkflow(sampleWorkflowWithRepo("wf-a", "/repo-a")))
	require.NoError(t, s.SaveWorkflow(sampleWorkflowWithRepo("wf-b", "/repo-b")))
	require.NoError(t, s.SaveWorkflow(sampleWorkflow("wf-none"))) // no RepoPath

	all, err := s.ListWorkflows()
	require.NoError(t, err)
	assert.Len(t, all, 3, "ListWorkflows should return all workflows regardless of RepoPath")
}

// ── RepoPath Serialization ──

func TestWorkflowDef_RepoPathSerialization(t *testing.T) {
	t.Run("RepoPath persists through save/load cycle", func(t *testing.T) {
		s := newTestStorage(t)
		wf := sampleWorkflowWithRepo("wf-persist", "/my/project")
		require.NoError(t, s.SaveWorkflow(wf))

		loaded, err := s.LoadWorkflow("wf-persist")
		require.NoError(t, err)
		assert.Equal(t, "/my/project", loaded.RepoPath)
	})

	t.Run("backward compat: JSON without repoPath loads fine", func(t *testing.T) {
		s := newTestStorage(t)
		// Write JSON without repoPath field (simulating old data)
		oldJSON := `{
			"id": "wf-legacy",
			"name": "Legacy Workflow",
			"description": "No repoPath field",
			"nodes": [],
			"edges": [],
			"isTemplate": false,
			"createdAt": "2026-04-01T00:00:00Z",
			"updatedAt": "2026-04-01T00:00:00Z"
		}`
		legacyPath := filepath.Join(s.workflowDir, "wf-legacy.json")
		require.NoError(t, os.WriteFile(legacyPath, []byte(oldJSON), 0644))

		loaded, err := s.LoadWorkflow("wf-legacy")
		require.NoError(t, err)
		assert.Equal(t, "", loaded.RepoPath, "RepoPath should be empty string for legacy data")
	})

	t.Run("RepoPath omitted from JSON when empty", func(t *testing.T) {
		wf := sampleWorkflow("wf-empty-repo")
		data, err := json.Marshal(wf)
		require.NoError(t, err)
		assert.NotContains(t, string(data), "repoPath", "empty RepoPath should be omitted via omitempty")
	})

	t.Run("RepoPath present in JSON when set", func(t *testing.T) {
		wf := sampleWorkflowWithRepo("wf-with-repo", "/some/path")
		data, err := json.Marshal(wf)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"repoPath":"/some/path"`)
	})
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
