package bmad

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetModules_Count(t *testing.T) {
	modules := GetModules()
	assert.Len(t, modules, 5)
}

func TestGetModules_CoreHas25Processes(t *testing.T) {
	modules := GetModules()
	var core ModuleDef
	for _, m := range modules {
		if m.ID == "core" {
			core = m
			break
		}
	}
	require.NotEmpty(t, core.ID)
	assert.GreaterOrEqual(t, len(core.Processes), 25)
}

func TestGetModules_CoreProcessIDsExistInRegistry(t *testing.T) {
	modules := GetModules()
	for _, m := range modules {
		if m.ID == "core" {
			for _, pid := range m.Processes {
				_, ok := ProcessByID(pid)
				assert.True(t, ok, "core module references unknown process %q", pid)
			}
			return
		}
	}
	t.Fatal("core module not found")
}

func TestGetModules_UniqueIDs(t *testing.T) {
	seen := map[string]bool{}
	for _, m := range GetModules() {
		assert.False(t, seen[m.ID], "duplicate module ID: %s", m.ID)
		seen[m.ID] = true
	}
}

// ── Sprint 4 Story 3: Module Population Tests ──

func TestAC2_ModulesPopulated_CIS(t *testing.T) {
	modules := GetModules()
	var cis ModuleDef
	found := false
	for _, m := range modules {
		if m.ID == "cis" {
			cis = m
			found = true
			break
		}
	}
	require.True(t, found, "cis module must exist")
	assert.Contains(t, cis.Processes, "bmad-infrastructure-devops",
		"cis module must contain bmad-infrastructure-devops")
}

func TestAC2_ModulesPopulated_BMGD(t *testing.T) {
	modules := GetModules()
	var bmgd ModuleDef
	found := false
	for _, m := range modules {
		if m.ID == "bmgd" {
			bmgd = m
			found = true
			break
		}
	}
	require.True(t, found, "bmgd module must exist")
	assert.Contains(t, bmgd.Processes, "bmad-game-dev-studio",
		"bmgd module must contain bmad-game-dev-studio")
}

func TestAC2_ModulesPopulated_CoreIncludesNew(t *testing.T) {
	modules := GetModules()
	var core ModuleDef
	found := false
	for _, m := range modules {
		if m.ID == "core" {
			core = m
			found = true
			break
		}
	}
	require.True(t, found, "core module must exist")

	expectedNew := []string{
		"bmad-party-mode",
		"bmad-quick-flow",
		"bmad-adversarial-general",
		"bmad-document-project",
		"bmad-web-orchestrator",
	}
	for _, id := range expectedNew {
		assert.Contains(t, core.Processes, id,
			"core module must contain %q", id)
	}
}

func TestGetModules_UnpopulatedModulesEmpty(t *testing.T) {
	// bmb and tea have no processes registered; bmgd and cis now have processes.
	unpopulated := []string{"bmb", "tea"}
	modules := GetModules()
	byID := map[string]ModuleDef{}
	for _, m := range modules {
		byID[m.ID] = m
	}
	for _, id := range unpopulated {
		m, ok := byID[id]
		require.True(t, ok, "module %q not found", id)
		assert.Empty(t, m.Processes, "module %q should have empty processes", id)
	}
}
