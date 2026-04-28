// Helper-level coverage for processIndex (interactive_defaults.go). The init()
// happy path is exercised by package load; this file pins the panic path so
// the "unknown process id" guard cannot regress silently.
package bmad

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessIndex_Known(t *testing.T) {
	for _, id := range phase2RolloutIDs {
		id := id
		t.Run(id, func(t *testing.T) {
			idx := processIndex(id)
			require.GreaterOrEqual(t, idx, 0, "processIndex must return non-negative slot")
			require.Less(t, idx, len(registry),
				"processIndex must return slot inside registry slice")
			assert.Equal(t, id, registry[idx].ID,
				"processIndex(%q) must point at the matching ProcessDef", id)
		})
	}
}

func TestProcessIndex_UnknownPanics(t *testing.T) {
	defer func() {
		r := recover()
		require.NotNil(t, r, "processIndex must panic on unknown id")
		msg, ok := r.(string)
		require.True(t, ok, "panic value must be a string, got %T", r)
		assert.True(t,
			strings.HasPrefix(msg, "bmad: unknown process id in interactive rollout: "),
			"panic message must name the missing id (got %q)", msg)
		assert.Contains(t, msg, "bmad-does-not-exist",
			"panic message must include the typo")
	}()

	_ = processIndex("bmad-does-not-exist")
	t.Fatal("expected processIndex to panic")
}
