package bmad

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuiltinTemplates_Count(t *testing.T) {
	templates := BuiltinTemplates()
	assert.Len(t, templates, 8)
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

// ── Sprint 4 Story 3: New Template & Chain Validation Tests ──

func TestAC3_BuiltinTemplateCount(t *testing.T) {
	templates := BuiltinTemplates()
	assert.Len(t, templates, 8, "BuiltinTemplates must return 8 templates (6 existing + 2 new)")
}

func TestAC3_RapidPrototypeTemplate(t *testing.T) {
	templates := BuiltinTemplates()
	var tpl WorkflowDef
	found := false
	for _, t := range templates {
		if t.ID == "tpl-rapid-prototype" {
			tpl = t
			found = true
			break
		}
	}
	require.True(t, found, "tpl-rapid-prototype template must exist")
	assert.True(t, tpl.IsTemplate, "must be a template")

	require.Len(t, tpl.Nodes, 2, "rapid prototype must have 2 nodes")
	assert.Equal(t, "bmad-party-mode", tpl.Nodes[0].ProcessID, "node 1 must be party-mode")
	assert.Equal(t, "bmad-code-review", tpl.Nodes[1].ProcessID, "node 2 must be code-review")

	require.Len(t, tpl.Edges, 1, "rapid prototype must have 1 edge")
	assert.Equal(t, tpl.Nodes[0].ID, tpl.Edges[0].Source, "edge source must be node 1")
	assert.Equal(t, tpl.Nodes[1].ID, tpl.Edges[0].Target, "edge target must be node 2")
}

func TestAC3_FullInfraTemplate(t *testing.T) {
	templates := BuiltinTemplates()
	var tpl WorkflowDef
	found := false
	for _, t := range templates {
		if t.ID == "tpl-full-infra" {
			tpl = t
			found = true
			break
		}
	}
	require.True(t, found, "tpl-full-infra template must exist")
	assert.True(t, tpl.IsTemplate, "must be a template")

	require.Len(t, tpl.Nodes, 3, "full infra must have 3 nodes")
	assert.Equal(t, "bmad-create-architecture", tpl.Nodes[0].ProcessID, "node 1 must be create-architecture")
	assert.Equal(t, "bmad-infrastructure-devops", tpl.Nodes[1].ProcessID, "node 2 must be infrastructure-devops")
	assert.Equal(t, "bmad-code-review", tpl.Nodes[2].ProcessID, "node 3 must be code-review")

	require.Len(t, tpl.Edges, 2, "full infra must have 2 edges chained")
	assert.Equal(t, tpl.Nodes[0].ID, tpl.Edges[0].Source, "edge 1 source")
	assert.Equal(t, tpl.Nodes[1].ID, tpl.Edges[0].Target, "edge 1 target")
	assert.Equal(t, tpl.Nodes[1].ID, tpl.Edges[1].Source, "edge 2 source")
	assert.Equal(t, tpl.Nodes[2].ID, tpl.Edges[1].Target, "edge 2 target")
}

func TestAC4_TemplateChainValidation(t *testing.T) {
	templates := BuiltinTemplates()
	require.Len(t, templates, 8, "must have 8 templates for this test to be valid")

	for _, tpl := range templates {
		t.Run(tpl.ID, func(t *testing.T) {
			// Build node ID -> ProcessDef lookup for this template.
			nodeProcess := make(map[string]ProcessDef)
			for _, n := range tpl.Nodes {
				p, ok := ProcessByID(n.ProcessID)
				require.True(t, ok, "template %q node %q references unknown process %q",
					tpl.ID, n.ID, n.ProcessID)
				nodeProcess[n.ID] = p
			}

			// For every edge A->B, verify B's inputs are empty or overlap with A's outputs.
			for _, e := range tpl.Edges {
				srcProc := nodeProcess[e.Source]
				dstProc := nodeProcess[e.Target]

				if len(dstProc.Inputs) == 0 {
					continue // B accepts anything (no required inputs).
				}

				srcOutputs := make(map[string]bool, len(srcProc.Outputs))
				for _, out := range srcProc.Outputs {
					srcOutputs[out] = true
				}

				hasOverlap := false
				for _, in := range dstProc.Inputs {
					if srcOutputs[in] {
						hasOverlap = true
						break
					}
				}
				assert.True(t, hasOverlap,
					"template %q edge %q: process %q outputs %v do not satisfy process %q inputs %v",
					tpl.ID, e.ID, srcProc.ID, srcProc.Outputs, dstProc.ID, dstProc.Inputs)
			}
		})
	}
}

func TestTemplateByName_Missing(t *testing.T) {
	_, ok := TemplateByName("nonexistent")
	assert.False(t, ok)
}
