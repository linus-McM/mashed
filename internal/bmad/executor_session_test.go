package bmad

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── Session-reuse test helpers ──

// capturedArgv records tmux new-session argv for regression testing.
type capturedArgv struct {
	mu   sync.Mutex
	args [][]string
}

func (c *capturedArgv) record(args []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]string, len(args))
	copy(cp, args)
	c.args = append(c.args, cp)
}

func (c *capturedArgv) last() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.args) == 0 {
		return nil
	}
	return c.args[len(c.args)-1]
}

// newSessionState creates a minimal execState with edges and nodes for testing
// resolveCommandSession. Does not need a real workflow or storage.
func newSessionState(nodes []WorkflowNode, edges []WorkflowEdge) (*execState, map[string]int) {
	outEdges := make(map[string][]WorkflowEdge, len(nodes))
	inDegree := make(map[string]int, len(nodes))
	for _, n := range nodes {
		inDegree[n.ID] = 0
	}
	for _, edge := range edges {
		outEdges[edge.Source] = append(outEdges[edge.Source], edge)
		inDegree[edge.Target]++
	}

	exec := &WorkflowExecution{
		ID:          "exec-test",
		WorkflowID:  "wf-test",
		Status:      ExecRunning,
		Nodes:       nodes,
		StartedAt:   time.Now().UTC().Format(time.RFC3339),
		NodeOutputs: make(map[string]string),
	}

	state := &execState{
		exec:             exec,
		inDegree:         inDegree,
		outEdges:         outEdges,
		lastQuestionHash: make(map[string]string),
		lastOutputHash:   make(map[string]string),
		idleEmitted:      make(map[string]bool),
	}

	nodeIndex := make(map[string]int, len(nodes))
	for i, n := range nodes {
		nodeIndex[n.ID] = i
	}

	return state, nodeIndex
}

// ── AC-1: spawnCommandSession regression — same argv as pre-refactor ──

func TestSpawnCommandSession_RegressionFromRefactor(t *testing.T) {
	var captured capturedArgv

	runner := func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "new-session" {
			captured.record(args)
		}
		if name == "git" {
			return []byte("main\n"), nil
		}
		return []byte("ok"), nil
	}

	h := newHarness(t)
	h.executor.SetCommandRunner(runner)

	// Build a minimal state with a process node.
	nodes := []WorkflowNode{
		{ID: "proc-1", ProcessID: "bmad-brainstorming", Label: "Brainstorm", Position: Position{X: 0, Y: 0}, Status: NodeRunning, Config: map[string]string{}},
	}
	state, nodeIndex := newSessionState(nodes, nil)

	ctx := context.Background()
	invocation := `claude --dangerously-skip-permissions --model opus "use brainstorming"`

	target, err := h.executor.spawnCommandSession(ctx, state, nodeIndex, "proc-1", "/tmp/test-repo", invocation)
	require.NoError(t, err, "spawnCommandSession must succeed")
	require.NotEmpty(t, target, "target must be returned")

	// Verify the argv structure matches expected pattern.
	argv := captured.last()
	require.NotNil(t, argv, "new-session must have been called")

	// Expected: new-session -d -s <session-name> -c <repoPath> bash -c '...; exec bash'
	assert.Equal(t, "new-session", argv[0])
	assert.Equal(t, "-d", argv[1])
	assert.Equal(t, "-s", argv[2])
	// argv[3] is the session name — dynamic, just verify non-empty
	assert.NotEmpty(t, argv[3], "session name must be present")
	assert.Equal(t, "-c", argv[4])
	assert.Equal(t, "/tmp/test-repo", argv[5])

	// The final arg is the bash-wrapped command.
	bashCmd := argv[6]
	assert.True(t, strings.HasPrefix(bashCmd, "bash -c '"), "command must be bash -c wrapped")
	assert.True(t, strings.HasSuffix(bashCmd, "; exec bash'"), "command must end with exec bash")
	assert.Contains(t, bashCmd, invocation, "inner invocation must be preserved")

	// Verify TmuxTarget was set on the node.
	state.mu.Lock()
	assert.NotEmpty(t, state.exec.Nodes[0].TmuxTarget, "TmuxTarget must be recorded")
	assert.NotEmpty(t, state.exec.Nodes[0].StartedAt, "StartedAt must be recorded")
	state.mu.Unlock()

	// Target must be session:0.0
	assert.True(t, strings.HasSuffix(target, ":0.0"), "target must end with :0.0")
}

// ── AC-3: resolveCommandSession — zero live parents ──

func TestResolveCommandSession_ZeroLiveParents(t *testing.T) {
	// Parent has empty TmuxTarget → no candidate → reused=false.
	nodes := []WorkflowNode{
		{ID: "parent-1", Label: "Parent", Status: NodeComplete, Config: map[string]string{}, TmuxTarget: ""},
		{ID: "cmd-1", Label: "Command", Status: NodePending, Config: map[string]string{}, NodeType: NodeTypeCommand},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "parent-1", Target: "cmd-1"},
	}

	state, nodeIndex := newSessionState(nodes, edges)
	e := NewExecutor(nil, func(string, interface{}) {})
	e.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		return nil, fmt.Errorf("should not be called")
	})

	ctx := context.Background()
	target, reused, err := e.resolveCommandSession(ctx, state, nodeIndex, "cmd-1")

	assert.NoError(t, err)
	assert.False(t, reused, "must not reuse when parent has empty TmuxTarget")
	assert.Empty(t, target, "target must be empty when no live parent")
}

// ── AC-4: resolveCommandSession — one live parent ──

func TestResolveCommandSession_OneLiveParent(t *testing.T) {
	parentTarget := "bmad-parent:0.0"
	nodes := []WorkflowNode{
		{ID: "parent-1", Label: "Parent", Status: NodeComplete, Config: map[string]string{}, TmuxTarget: parentTarget, StartedAt: "2026-04-12T10:00:00Z"},
		{ID: "cmd-1", Label: "Command", Status: NodePending, Config: map[string]string{}, NodeType: NodeTypeCommand},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "parent-1", Target: "cmd-1"},
	}

	state, nodeIndex := newSessionState(nodes, edges)
	e := NewExecutor(nil, func(string, interface{}) {})
	e.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		// list-panes returns alive
		if name == "tmux" && len(args) > 0 && args[0] == "list-panes" {
			return []byte("0\n"), nil
		}
		return []byte("ok"), nil
	})

	ctx := context.Background()
	target, reused, err := e.resolveCommandSession(ctx, state, nodeIndex, "cmd-1")

	assert.NoError(t, err)
	assert.True(t, reused, "must reuse live parent session")
	assert.Equal(t, parentTarget, target, "target must match parent's TmuxTarget")
}

// ── AC-5: resolveCommandSession — stale parent (dead pane) ──

func TestResolveCommandSession_StaleParent(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "parent-1", Label: "Parent", Status: NodeComplete, Config: map[string]string{}, TmuxTarget: "bmad-stale:0.0", StartedAt: "2026-04-12T10:00:00Z"},
		{ID: "cmd-1", Label: "Command", Status: NodePending, Config: map[string]string{}, NodeType: NodeTypeCommand},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "parent-1", Target: "cmd-1"},
	}

	state, nodeIndex := newSessionState(nodes, edges)
	e := NewExecutor(nil, func(string, interface{}) {})
	e.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		// list-panes returns dead
		if name == "tmux" && len(args) > 0 && args[0] == "list-panes" {
			return []byte("1\n"), nil
		}
		return []byte("ok"), nil
	})

	ctx := context.Background()
	target, reused, err := e.resolveCommandSession(ctx, state, nodeIndex, "cmd-1")

	assert.NoError(t, err)
	assert.False(t, reused, "must not reuse dead pane")
	assert.Empty(t, target, "target must be empty when pane is dead")
}

// ── AC-6: resolveCommandSession — multi-parent picks most recent ──

func TestResolveCommandSession_MultiParentPicksMostRecent(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "parent-a", Label: "ParentA", Status: NodeComplete, Config: map[string]string{}, TmuxTarget: "bmad-a:0.0", StartedAt: "2026-04-12T10:00:00Z"},
		{ID: "parent-b", Label: "ParentB", Status: NodeComplete, Config: map[string]string{}, TmuxTarget: "bmad-b:0.0", StartedAt: "2026-04-12T10:05:00Z"},
		{ID: "cmd-1", Label: "Command", Status: NodePending, Config: map[string]string{}, NodeType: NodeTypeCommand},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "parent-a", Target: "cmd-1"},
		{ID: "e2", Source: "parent-b", Target: "cmd-1"},
	}

	state, nodeIndex := newSessionState(nodes, edges)

	var logOutput strings.Builder
	e := NewExecutor(nil, func(string, interface{}) {})
	e.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "list-panes" {
			return []byte("0\n"), nil // both alive
		}
		return []byte("ok"), nil
	})

	ctx := context.Background()
	target, reused, err := e.resolveCommandSession(ctx, state, nodeIndex, "cmd-1")

	assert.NoError(t, err)
	assert.True(t, reused, "must reuse most recently started parent")
	assert.Equal(t, "bmad-b:0.0", target, "must pick parent-b (later StartedAt)")

	// Verify log output (log.Printf goes to stderr, capture is complex —
	// we test the logic picks correctly and trust the log.Printf call).
	_ = logOutput
}

// ── AC-5 variant: session gone entirely (error from tmux) ──

func TestResolveCommandSession_SessionGoneError(t *testing.T) {
	nodes := []WorkflowNode{
		{ID: "parent-1", Label: "Parent", Status: NodeComplete, Config: map[string]string{}, TmuxTarget: "bmad-gone:0.0", StartedAt: "2026-04-12T10:00:00Z"},
		{ID: "cmd-1", Label: "Command", Status: NodePending, Config: map[string]string{}, NodeType: NodeTypeCommand},
	}
	edges := []WorkflowEdge{
		{ID: "e1", Source: "parent-1", Target: "cmd-1"},
	}

	state, nodeIndex := newSessionState(nodes, edges)
	e := NewExecutor(nil, func(string, interface{}) {})
	e.SetCommandRunner(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		// tmux errors when the session doesn't exist
		if name == "tmux" && len(args) > 0 && args[0] == "list-panes" {
			return nil, fmt.Errorf("session not found: bmad-gone")
		}
		return []byte("ok"), nil
	})

	ctx := context.Background()
	target, reused, err := e.resolveCommandSession(ctx, state, nodeIndex, "cmd-1")

	assert.NoError(t, err, "session gone should not return error — just no reuse")
	assert.False(t, reused)
	assert.Empty(t, target)
}
