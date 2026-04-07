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

func TestGetModules_ReservedModulesEmpty(t *testing.T) {
	reserved := []string{"bmb", "tea", "bmgd", "cis"}
	modules := GetModules()
	byID := map[string]ModuleDef{}
	for _, m := range modules {
		byID[m.ID] = m
	}
	for _, id := range reserved {
		m, ok := byID[id]
		require.True(t, ok, "module %q not found", id)
		assert.Empty(t, m.Processes, "reserved module %q should have empty processes", id)
	}
}
