package bmad

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// CommandRunner executes a shell command and returns its output.
// The default implementation runs tmux; tests inject a mock.
type CommandRunner func(ctx context.Context, name string, args ...string) ([]byte, error)

// DefaultCommandRunner uses exec.CommandContext.
func DefaultCommandRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	return exec.CommandContext(ctx, name, args...).CombinedOutput()
}

// NodeStatusEvent is emitted on every node state change.
type NodeStatusEvent struct {
	ExecID     string             `json:"execId"`
	NodeID     string             `json:"nodeId"`
	Status     WorkflowNodeStatus `json:"status"`
	TmuxTarget string             `json:"tmuxTarget,omitempty"`
}

// ExecStatusEvent is emitted on overall execution state changes.
type ExecStatusEvent struct {
	ExecID string             `json:"execId"`
	Status WorkflowExecStatus `json:"status"`
}

// execState holds the mutable runtime state of a single execution.
type execState struct {
	exec   *WorkflowExecution
	cancel context.CancelFunc
	paused bool
	mu     sync.Mutex
}

// Executor manages workflow executions.
type Executor struct {
	storage      *Storage
	runCmd       CommandRunner
	emitEvent    func(string, interface{})
	executions   map[string]*execState
	pollInterval time.Duration
	mu           sync.RWMutex
}

// NewExecutor creates an Executor with the given storage and event emitter.
func NewExecutor(storage *Storage, emitEvent func(string, interface{})) *Executor {
	return &Executor{
		storage:      storage,
		runCmd:       DefaultCommandRunner,
		emitEvent:    emitEvent,
		executions:   make(map[string]*execState),
		pollInterval: 3 * time.Second,
	}
}

// SetCommandRunner replaces the command runner (for testing).
func (e *Executor) SetCommandRunner(runner CommandRunner) {
	e.runCmd = runner
}

// StartWorkflow loads a workflow, validates its DAG, creates an execution, and
// begins running nodes in topological order.
func (e *Executor) StartWorkflow(workflowID, repoPath, model string) (*WorkflowExecution, error) {
	wf, err := e.storage.LoadWorkflow(workflowID)
	if err != nil {
		return nil, err
	}

	tiers, err := topoSort(wf.Nodes, wf.Edges)
	if err != nil {
		return nil, err
	}

	// Build execution with copies of nodes, all pending.
	nodes := make([]WorkflowNode, len(wf.Nodes))
	for i, n := range wf.Nodes {
		nodes[i] = n
		nodes[i].Status = NodePending
	}

	execID := fmt.Sprintf("exec-%s-%d", workflowID, time.Now().UnixMilli())
	execution := &WorkflowExecution{
		ID:         execID,
		WorkflowID: workflowID,
		RepoPath:   repoPath,
		Status:     ExecRunning,
		Nodes:      nodes,
		StartedAt:  time.Now().UTC().Format(time.RFC3339),
	}

	ctx, cancel := context.WithCancel(context.Background())
	state := &execState{exec: execution, cancel: cancel}

	e.mu.Lock()
	e.executions[execID] = state
	e.mu.Unlock()

	e.emitEvent("bmad:execution:status", ExecStatusEvent{ExecID: execID, Status: ExecRunning})

	go e.run(ctx, state, tiers, repoPath, model)

	return execution, nil
}

// PauseWorkflow prevents new nodes from starting; running nodes finish.
func (e *Executor) PauseWorkflow(execID string) error {
	state, err := e.getState(execID)
	if err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.exec.Status != ExecRunning {
		return ErrExecNotRunning
	}
	state.paused = true
	state.exec.Status = ExecPaused
	e.emitEvent("bmad:execution:status", ExecStatusEvent{ExecID: execID, Status: ExecPaused})
	return nil
}

// ResumeWorkflow resumes a paused execution.
func (e *Executor) ResumeWorkflow(execID string) error {
	state, err := e.getState(execID)
	if err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()

	if state.exec.Status != ExecPaused {
		return ErrExecNotPaused
	}
	state.paused = false
	state.exec.Status = ExecRunning
	e.emitEvent("bmad:execution:status", ExecStatusEvent{ExecID: execID, Status: ExecRunning})
	return nil
}

// StopWorkflow cancels the execution context and marks it failed.
func (e *Executor) StopWorkflow(execID string) error {
	state, err := e.getState(execID)
	if err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()

	state.cancel()
	state.exec.Status = ExecFailed
	e.emitEvent("bmad:execution:status", ExecStatusEvent{ExecID: execID, Status: ExecFailed})
	return nil
}

// GetExecution returns a copy of the execution state.
func (e *Executor) GetExecution(execID string) (*WorkflowExecution, error) {
	state, err := e.getState(execID)
	if err != nil {
		return nil, err
	}
	state.mu.Lock()
	defer state.mu.Unlock()

	cp := *state.exec
	cp.Nodes = make([]WorkflowNode, len(state.exec.Nodes))
	copy(cp.Nodes, state.exec.Nodes)
	return &cp, nil
}

func (e *Executor) getState(execID string) (*execState, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	state, ok := e.executions[execID]
	if !ok {
		return nil, ErrExecNotFound
	}
	return state, nil
}

// run drives the execution through topological tiers.
func (e *Executor) run(ctx context.Context, state *execState, tiers [][]string, repoPath, model string) {
	nodeIndex := buildNodeIndex(state.exec.Nodes)

	for _, tier := range tiers {
		// Wait if paused.
		if !e.waitWhilePaused(ctx, state) {
			return // context canceled
		}

		var wg sync.WaitGroup
		for _, nodeID := range tier {
			// Check pause/cancel before spawning each node.
			if ctx.Err() != nil {
				return
			}
			state.mu.Lock()
			paused := state.paused
			state.mu.Unlock()
			if paused {
				// Re-enter the pause wait for remaining tiers.
				if !e.waitWhilePaused(ctx, state) {
					return
				}
			}

			wg.Add(1)
			go func(nID string) {
				defer wg.Done()
				e.executeNode(ctx, state, nodeIndex, nID, repoPath, model)
			}(nodeID)
		}
		wg.Wait()

		// Check for failures in this tier.
		state.mu.Lock()
		anyFailed := false
		for _, nID := range tier {
			if idx, ok := nodeIndex[nID]; ok && state.exec.Nodes[idx].Status == NodeFailed {
				anyFailed = true
				break
			}
		}
		if anyFailed && state.exec.Status == ExecRunning {
			state.paused = true
			state.exec.Status = ExecPaused
			e.emitEvent("bmad:execution:status", ExecStatusEvent{ExecID: state.exec.ID, Status: ExecPaused})
		}
		state.mu.Unlock()

		if anyFailed {
			if !e.waitWhilePaused(ctx, state) {
				return
			}
		}
	}

	// All tiers done — mark complete if still running.
	state.mu.Lock()
	if state.exec.Status == ExecRunning {
		state.exec.Status = ExecComplete
		e.emitEvent("bmad:execution:status", ExecStatusEvent{ExecID: state.exec.ID, Status: ExecComplete})
	}
	state.mu.Unlock()
}

// waitWhilePaused blocks until unpaused or context is canceled. Returns false if canceled.
func (e *Executor) waitWhilePaused(ctx context.Context, state *execState) bool {
	for {
		state.mu.Lock()
		paused := state.paused
		state.mu.Unlock()
		if !paused {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func (e *Executor) executeNode(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath, model string) {
	idx := nodeIndex[nodeID]

	// Mark running.
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeRunning
	state.exec.CurrentNode = nodeID
	state.mu.Unlock()
	e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

	// Look up the process definition (ProcessID is immutable, safe to read after init).
	processID := state.exec.Nodes[idx].ProcessID
	proc, ok := ProcessByID(processID)
	if !ok {
		e.failNode(state, idx, nodeID)
		return
	}

	// Snapshot nodes under lock for context string building.
	state.mu.Lock()
	nodesCopy := make([]WorkflowNode, len(state.exec.Nodes))
	copy(nodesCopy, state.exec.Nodes)
	state.mu.Unlock()

	// Build command.
	contextStr := buildContextString(proc, nodesCopy, nodeIndex)
	command := fmt.Sprintf(`claude --dangerously-skip-permissions --model %s "use %s%s"`, model, proc.SkillName, contextStr)

	sessionName := fmt.Sprintf("bmad-%s-%d", nodeID, time.Now().Unix())
	_, err := e.runCmd(ctx, "tmux", "new-session", "-d",
		"-s", sessionName,
		"-c", repoPath,
		command,
	)
	if err != nil {
		e.failNode(state, idx, nodeID)
		return
	}

	target := fmt.Sprintf("%s:0.0", sessionName)
	state.mu.Lock()
	state.exec.Nodes[idx].TmuxTarget = target
	state.mu.Unlock()
	e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning, TmuxTarget: target})

	// Poll for completion.
	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			e.failNode(state, idx, nodeID)
			return
		case <-ticker.C:
			out, err := e.runCmd(ctx, "tmux", "list-panes", "-t", target, "-F", "#{pane_dead}")
			if err != nil {
				// Session gone — treat as complete.
				e.completeNode(state, idx, nodeID)
				return
			}
			if strings.TrimSpace(string(out)) == "1" {
				e.completeNode(state, idx, nodeID)
				return
			}
		}
	}
}

func (e *Executor) completeNode(state *execState, idx int, nodeID string) {
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeComplete
	storyID := state.exec.Nodes[idx].StoryID
	repoPath := state.exec.RepoPath
	state.mu.Unlock()
	e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeComplete})

	// Auto-advance linked sprint story status on node completion.
	if storyID != "" && repoPath != "" {
		targetStatus := string(StoryInProgress)
		if err := UpdateStoryStatus(repoPath, storyID, targetStatus); err != nil {
			fmt.Printf("bmad: failed to update story %s status: %v\n", storyID, err)
		} else {
			e.emitEvent("bmad:sprint:updated", map[string]string{"storyId": storyID, "status": targetStatus})
		}
	}
}

func (e *Executor) failNode(state *execState, idx int, nodeID string) {
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeFailed
	state.mu.Unlock()
	e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeFailed})
}

func buildContextString(proc ProcessDef, nodes []WorkflowNode, nodeIndex map[string]int) string {
	if len(proc.Inputs) == 0 {
		return ""
	}

	// Find completed upstream nodes that produce artifacts matching our inputs.
	needed := make(map[string]bool)
	for _, input := range proc.Inputs {
		needed[input] = true
	}

	var parts []string
	for _, n := range nodes {
		if n.Status != NodeComplete {
			continue
		}
		upstream, ok := ProcessByID(n.ProcessID)
		if !ok {
			continue
		}
		for _, output := range upstream.Outputs {
			if needed[output] {
				parts = append(parts, fmt.Sprintf(" The upstream process '%s' produced '%s' -- use it as input.", upstream.Name, output))
			}
		}
	}
	return strings.Join(parts, "")
}

func buildNodeIndex(nodes []WorkflowNode) map[string]int {
	idx := make(map[string]int, len(nodes))
	for i, n := range nodes {
		idx[n.ID] = i
	}
	return idx
}

// topoSort performs Kahn's algorithm on the workflow graph.
// Returns tiers: groups of node IDs that can execute in parallel.
func topoSort(nodes []WorkflowNode, edges []WorkflowEdge) ([][]string, error) {
	// Build adjacency and in-degree.
	inDegree := make(map[string]int)
	outEdges := make(map[string][]string)
	for _, n := range nodes {
		inDegree[n.ID] = 0
	}
	for _, e := range edges {
		outEdges[e.Source] = append(outEdges[e.Source], e.Target)
		inDegree[e.Target]++
	}

	// Seed with zero in-degree nodes.
	var queue []string
	for _, n := range nodes {
		if inDegree[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}

	var tiers [][]string
	visited := 0
	for len(queue) > 0 {
		tier := queue
		queue = nil
		tiers = append(tiers, tier)
		visited += len(tier)
		for _, nID := range tier {
			for _, target := range outEdges[nID] {
				inDegree[target]--
				if inDegree[target] == 0 {
					queue = append(queue, target)
				}
			}
		}
	}

	if visited != len(nodes) {
		return nil, ErrCyclicWorkflow
	}
	return tiers, nil
}
