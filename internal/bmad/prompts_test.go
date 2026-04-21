package bmad

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readFixture(t *testing.T, relPath string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", relPath))
	require.NoError(t, err, "read fixture %s", relPath)
	return data
}

// emitAwaitingInputForAll exercises the awaiting_input emit payload path for
// every PendingPrompt in exec. It mirrors what the executor's real emit path
// does on resume (prompts.go calls awaitingPayload(prompt)), but uses a no-op
// sink — the goal is to prove legacy snapshots don't panic when their prompts
// are threaded through the production payload helper.
func emitAwaitingInputForAll(exec *WorkflowExecution) {
	for _, p := range exec.PendingPrompts {
		// awaitingPayload is the real production helper. Passing a legacy
		// prompt through it is the smoke test: if any field access or type
		// assertion surprised the helper, we'd panic here.
		_ = awaitingPayload(p)
	}
}

// ── Story ui-ast-U4, AC-10: pre-U4 snapshot migration regression ──
//
// A WorkflowExecution snapshot captured BEFORE this story must load cleanly,
// leave every PendingPrompt.Structured empty, re-emit awaiting_input without
// panic, and survive a marshal → unmarshal round-trip under reflect.DeepEqual.
//
// Reference: docs/stories/ui-ast-U4-executor-wiring.md §5.1 lines 210–234.
func TestSnapshot_PreU4LoadRoundTrip(t *testing.T) {
	raw := readFixture(t, "snapshots/pre-u4/execution.json")

	var exec WorkflowExecution
	require.NoError(t, json.Unmarshal(raw, &exec),
		"legacy snapshot must decode under the post-U4 schema")

	t.Run("PendingPrompts load with empty Structured", func(t *testing.T) {
		require.NotEmpty(t, exec.PendingPrompts,
			"fixture must carry at least one PendingPrompt to be a meaningful probe")
		for _, p := range exec.PendingPrompts {
			assert.Empty(t, p.Structured,
				"Structured must be empty on legacy load (prompt=%s/%s)", p.NodeID, p.InputID)
		}
	})

	t.Run("NodeInputHistory legacy entries have empty Key", func(t *testing.T) {
		require.NotEmpty(t, exec.NodeInputHistory,
			"fixture must carry NodeInputHistory entries to probe legacy migration")
		for nodeID, entries := range exec.NodeInputHistory {
			require.NotEmpty(t, entries,
				"NodeInputHistory[%s] must have entries", nodeID)
			for i, ent := range entries {
				assert.Empty(t, ent.Key,
					"Key must be empty on legacy NodeInputHistory[%s][%d]", nodeID, i)
			}
		}
	})

	t.Run("awaiting_input re-emit does not panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			emitAwaitingInputForAll(&exec)
		}, "emitting awaiting_input for legacy prompts must not panic")
	})

	t.Run("struct round-trips via reflect.DeepEqual", func(t *testing.T) {
		again, err := json.Marshal(exec)
		require.NoError(t, err, "remarshal must succeed")

		var roundtrip WorkflowExecution
		require.NoError(t, json.Unmarshal(again, &roundtrip),
			"second unmarshal must succeed")

		assert.Equal(t, exec, roundtrip,
			"WorkflowExecution must round-trip byte-for-byte via JSON marshal/unmarshal")
	})
}
