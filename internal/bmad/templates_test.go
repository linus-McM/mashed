package bmad

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinTemplates_Count(t *testing.T) {
	templates := BuiltinTemplates()
	assert.Len(t, templates, 6)
}

func TestBuiltinTemplates_AllAreTemplates(t *testing.T) {
	for _, tpl := range BuiltinTemplates() {
		assert.True(t, tpl.IsTemplate, "template %q must have IsTemplate=true", tpl.Name)
	}
}

func TestBuiltinTemplates_HaveNodesAndEdges(t *testing.T) {
	for _, tpl := range BuiltinTemplates() {
		assert.GreaterOrEqual(t, len(tpl.Nodes), 2, "template %q must have at least 2 nodes", tpl.Name)
		assert.GreaterOrEqual(t, len(tpl.Edges), 1, "template %q must have at least 1 edge", tpl.Name)
	}
}

func TestBuiltinTemplates_NodeProcessIDsExistInRegistry(t *testing.T) {
	for _, tpl := range BuiltinTemplates() {
		for _, n := range tpl.Nodes {
			_, ok := ProcessByID(n.ProcessID)
			assert.True(t, ok, "template %q node %q references unknown process %q", tpl.Name, n.ID, n.ProcessID)
		}
	}
}

func TestBuiltinTemplates_UniqueIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, tpl := range BuiltinTemplates() {
		assert.False(t, seen[tpl.ID], "duplicate template ID: %s", tpl.ID)
		seen[tpl.ID] = true
	}
}

func TestBuiltinTemplates_UniqueNodeIDs(t *testing.T) {
	for _, tpl := range BuiltinTemplates() {
		seen := map[string]bool{}
		for _, n := range tpl.Nodes {
			assert.False(t, seen[n.ID], "template %q has duplicate node ID: %s", tpl.Name, n.ID)
			seen[n.ID] = true
		}
	}
}

func TestBuiltinTemplates_EdgeSourceTargetExist(t *testing.T) {
	for _, tpl := range BuiltinTemplates() {
		nodeIDs := map[string]bool{}
		for _, n := range tpl.Nodes {
			nodeIDs[n.ID] = true
		}
		for _, e := range tpl.Edges {
			assert.True(t, nodeIDs[e.Source], "template %q edge %q has unknown source %q", tpl.Name, e.ID, e.Source)
			assert.True(t, nodeIDs[e.Target], "template %q edge %q has unknown target %q", tpl.Name, e.ID, e.Target)
		}
	}
}

func TestBuiltinTemplates_AllNodesPending(t *testing.T) {
	for _, tpl := range BuiltinTemplates() {
		for _, n := range tpl.Nodes {
			assert.Equal(t, NodePending, n.Status, "template %q node %q must be pending", tpl.Name, n.ID)
		}
	}
}

func TestQuickSprint_Structure(t *testing.T) {
	tpl, ok := TemplateByName("Quick Sprint")
	require.True(t, ok)
	assert.Len(t, tpl.Nodes, 4)
	assert.Len(t, tpl.Edges, 3)

	expectedProcessIDs := []string{
		"bmad-create-story",
		"bmad-dev-story",
		"bmad-code-review",
		"bmad-qa-generate-e2e-tests",
	}
	for i, n := range tpl.Nodes {
		assert.Equal(t, expectedProcessIDs[i], n.ProcessID, "node %d", i)
	}
}

func TestFullProductLifecycle_Structure(t *testing.T) {
	tpl, ok := TemplateByName("Full Product Lifecycle")
	require.True(t, ok)
	assert.Len(t, tpl.Nodes, 13)
	assert.Len(t, tpl.Edges, 12)
}

func TestTemplateByName_Missing(t *testing.T) {
	_, ok := TemplateByName("nonexistent")
	assert.False(t, ok)
}
