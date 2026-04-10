package bmad

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Node output key helpers for loop iteration tracking.
func nodeIterKey(nodeID string) string { return nodeID + "_iter" }
func nodeItemKey(nodeID string) string { return nodeID + "_item" }

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
	Iteration  int                `json:"iteration,omitempty"`
}

// ExecStatusEvent is emitted on overall execution state changes.
type ExecStatusEvent struct {
	ExecID string             `json:"execId"`
	Status WorkflowExecStatus `json:"status"`
}

// execState holds the mutable runtime state of a single execution.
type execState struct {
	exec             *WorkflowExecution
	cancel           context.CancelFunc
	paused           bool
	mu               sync.Mutex
	inDegree         map[string]int            // current in-degree per node
	outEdges         map[string][]WorkflowEdge // source -> edges
	lastQuestionHash map[string]string         // nodeID -> last emitted question hash
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
// begins running nodes in topological order. The parent context allows callers
// to cancel the workflow externally.
func (e *Executor) StartWorkflow(parentCtx context.Context, workflowID, repoPath, model string) (*WorkflowExecution, error) {
	wf, err := e.storage.LoadWorkflow(workflowID)
	if err != nil {
		return nil, err
	}

	// topoSort is called only for cycle detection; the dynamic executor
	// computes readiness on the fly instead of using static tiers.
	if _, err := topoSort(wf.Nodes, wf.Edges); err != nil {
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
		ID:          execID,
		WorkflowID:  workflowID,
		RepoPath:    repoPath,
		Status:      ExecRunning,
		Nodes:       nodes,
		StartedAt:   time.Now().UTC().Format(time.RFC3339),
		NodeOutputs: make(map[string]string),
	}

	ctx, cancel := context.WithCancel(parentCtx)

	// Build in-degree and outEdges maps for the dynamic executor.
	inDegree := make(map[string]int, len(nodes))
	outEdgesMap := make(map[string][]WorkflowEdge, len(nodes))
	for _, n := range nodes {
		inDegree[n.ID] = 0
	}
	for _, edge := range wf.Edges {
		outEdgesMap[edge.Source] = append(outEdgesMap[edge.Source], edge)
		inDegree[edge.Target]++
	}

	state := &execState{
		exec:             execution,
		cancel:           cancel,
		inDegree:         inDegree,
		outEdges:         outEdgesMap,
		lastQuestionHash: make(map[string]string),
	}

	e.mu.Lock()
	e.executions[execID] = state
	e.mu.Unlock()

	e.emitEvent("bmad:execution:status", ExecStatusEvent{ExecID: execID, Status: ExecRunning})

	go e.runDynamic(ctx, state, repoPath, model)

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
	if state.exec.NodeOutputs != nil {
		cp.NodeOutputs = make(map[string]string, len(state.exec.NodeOutputs))
		for k, v := range state.exec.NodeOutputs {
			cp.NodeOutputs[k] = v
		}
	}
	return &cp, nil
}

// maxAnswerBytes caps the size of a response written into a tmux pane.
// tmux send-keys -l handles long strings but extremely long input may be
// truncated or disrupted by terminal line editing, so we fail fast.
const maxAnswerBytes = 4096

// respondCmdTimeout bounds the tmux subprocess calls issued by RespondToQuestion
// so a hung tmux server cannot block the caller indefinitely.
const respondCmdTimeout = 5 * time.Second

// RespondToQuestion injects an answer into the Claude CLI tmux pane backing
// the given node. It verifies the pane is still alive, writes the literal
// answer via `tmux send-keys -l`, then dispatches Enter as a second call so
// tmux interprets the keystroke rather than sending the bytes "Enter".
//
// On success the node's cached question hash is cleared so the polling loop
// can re-detect any subsequent question. On failure (not found, dead pane,
// tmux error) the hash is left intact.
func (e *Executor) RespondToQuestion(execID, nodeID, answer string) error {
	if len(answer) > maxAnswerBytes {
		return fmt.Errorf("bmad: answer is %d bytes (max %d): %w", len(answer), maxAnswerBytes, ErrAnswerTooLong)
	}

	state, err := e.getState(execID)
	if err != nil {
		return fmt.Errorf("bmad: exec %q: %w", execID, err)
	}

	// Read the node's tmux target under lock; do not hold the lock across
	// tmux subprocess calls which can block for hundreds of ms.
	state.mu.Lock()
	var target string
	var found bool
	for _, n := range state.exec.Nodes {
		if n.ID == nodeID {
			target = n.TmuxTarget
			found = true
			break
		}
	}
	state.mu.Unlock()

	if !found {
		return fmt.Errorf("bmad: node %q not found in exec %q: %w", nodeID, execID, ErrExecNotFound)
	}
	if target == "" {
		return fmt.Errorf("bmad: node %q has no tmux target: %w", nodeID, ErrExecNotRunning)
	}

	// Pane liveness check — its own timeout so a slow tmux server does not
	// consume the budget for the subsequent send-keys calls.
	paneCtx, paneCancel := context.WithTimeout(context.Background(), respondCmdTimeout)
	out, err := e.runCmd(paneCtx, "tmux", "list-panes", "-t", target, "-F", "#{pane_dead}")
	paneCancel()
	if err != nil {
		return fmt.Errorf("bmad: node %q pane unreachable: %w", nodeID, ErrExecNotRunning)
	}
	if strings.TrimSpace(string(out)) == "1" {
		return fmt.Errorf("bmad: node %q pane is dead, cannot respond: %w", nodeID, ErrExecNotRunning)
	}

	// Send the literal answer (empty string is allowed — still send Enter).
	literalCtx, literalCancel := context.WithTimeout(context.Background(), respondCmdTimeout)
	_, err = e.runCmd(literalCtx, "tmux", "send-keys", "-l", "-t", target, escapeTmuxLiteral(answer))
	literalCancel()
	if err != nil {
		return fmt.Errorf("bmad: send-keys -l failed for node %q: %w", nodeID, err)
	}

	// Send Enter as a separate call (no -l) so tmux treats it as a key.
	enterCtx, enterCancel := context.WithTimeout(context.Background(), respondCmdTimeout)
	_, err = e.runCmd(enterCtx, "tmux", "send-keys", "-t", target, "Enter")
	enterCancel()
	if err != nil {
		return fmt.Errorf("bmad: send-keys Enter failed for node %q: %w", nodeID, err)
	}

	// Clear the cached question hash so the polling loop can detect a new one.
	state.mu.Lock()
	state.lastQuestionHash[nodeID] = ""
	state.mu.Unlock()

	return nil
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

// runDynamic drives execution using a dynamic ready-set algorithm.
// It replaces the static tier-based run() with dynamic in-degree tracking
// that supports condition branching and merge nodes.
func (e *Executor) runDynamic(ctx context.Context, state *execState, repoPath, model string) {
	nodeIndex := buildNodeIndex(state.exec.Nodes)

	for {
		// Wait if paused.
		if !e.waitWhilePaused(ctx, state) {
			return // context canceled
		}

		// Find ready set: nodes where inDegree == 0 AND status == NodePending.
		state.mu.Lock()
		var ready []string
		for _, n := range state.exec.Nodes {
			if n.Status == NodePending && state.inDegree[n.ID] == 0 {
				ready = append(ready, n.ID)
			}
		}
		state.mu.Unlock()

		if len(ready) == 0 {
			break // done or stuck
		}

		// Execute all ready nodes in parallel.
		var wg sync.WaitGroup
		var completedMu sync.Mutex
		var completed []string

		for _, nodeID := range ready {
			if ctx.Err() != nil {
				return
			}
			state.mu.Lock()
			paused := state.paused
			state.mu.Unlock()
			if paused {
				if !e.waitWhilePaused(ctx, state) {
					return
				}
			}

			wg.Add(1)
			go func(nID string) {
				defer wg.Done()

				state.mu.Lock()
				idx := nodeIndex[nID]
				effectiveType := state.exec.Nodes[idx].EffectiveType()
				state.mu.Unlock()

				switch effectiveType {
				case NodeTypeProcess:
					e.executeNode(ctx, state, nodeIndex, nID, repoPath, model)
				case NodeTypeCondition, NodeTypeMerge:
					e.executeControlNode(ctx, state, nodeIndex, nID, repoPath, model)
				case NodeTypeTransform:
					e.executeTransformNode(ctx, state, nodeIndex, nID)
				case NodeTypeLoop, NodeTypeLoopUntil:
					e.executeLoopNode(ctx, state, nodeIndex, nID, repoPath, model)
				default:
					log.Printf("bmad: unknown node type %s, skipping: %s", effectiveType, nID)
					e.skipNode(state, nodeIndex[nID], nID)
				}

				completedMu.Lock()
				completed = append(completed, nID)
				completedMu.Unlock()
			}(nodeID)
		}
		wg.Wait()

		// Process completions: update in-degrees and skip inactive branches.
		state.mu.Lock()
		anyFailed := false
		for _, nID := range completed {
			idx := nodeIndex[nID]
			nodeStatus := state.exec.Nodes[idx].Status

			if nodeStatus == NodeFailed {
				anyFailed = true
				continue
			}
			if nodeStatus != NodeComplete {
				continue // skipped or other terminal state
			}

			// Determine which outbound edges are active.
			result := state.exec.NodeOutputs[nID]
			effectiveType := state.exec.Nodes[idx].EffectiveType()
			active := e.activeOutEdges(state, nID, result, effectiveType)

			// Build set of active target IDs.
			activeTargets := make(map[string]bool, len(active))
			for _, edge := range active {
				activeTargets[edge.Target] = true
			}

			// Decrement in-degree for active targets.
			for _, edge := range active {
				state.inDegree[edge.Target]--
			}

			// For inactive edges: skip those branches.
			for _, edge := range state.outEdges[nID] {
				if !activeTargets[edge.Target] {
					// This edge is inactive — skip the target branch.
					state.inDegree[edge.Target]--
					e.skipBranchLocked(state, nodeIndex, edge.Target)
				}
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

	// Mark remaining pending nodes as skipped.
	state.mu.Lock()
	for i, n := range state.exec.Nodes {
		if n.Status == NodePending {
			state.exec.Nodes[i].Status = NodeSkipped
			e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: n.ID, Status: NodeSkipped})
		}
	}

	if state.exec.Status == ExecRunning {
		state.exec.Status = ExecComplete
		e.emitEvent("bmad:execution:status", ExecStatusEvent{ExecID: state.exec.ID, Status: ExecComplete})
	}
	state.mu.Unlock()
}

// activeOutEdges returns the outbound edges that should be activated after a node completes.
// For condition nodes, only edges matching the result ("true"/"false") are active.
// For loop nodes, only "loop-exit" edges are active (body edges are managed by executeLoopNode).
// For all other nodes, all outbound edges are active.
func (e *Executor) activeOutEdges(state *execState, nodeID, result string, effectiveType NodeType) []WorkflowEdge {
	edges := state.outEdges[nodeID]
	if effectiveType == NodeTypeCondition {
		var active []WorkflowEdge
		for _, edge := range edges {
			if edge.SourceHandle == result {
				active = append(active, edge)
			}
		}
		return active
	}

	if effectiveType == NodeTypeLoop || effectiveType == NodeTypeLoopUntil {
		var active []WorkflowEdge
		for _, edge := range edges {
			if edge.SourceHandle == "loop-exit" {
				active = append(active, edge)
			}
		}
		return active
	}

	return edges
}

// executeControlNode handles condition and merge nodes without spawning tmux sessions.
func (e *Executor) executeControlNode(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath, model string) {
	idx := nodeIndex[nodeID]

	state.mu.Lock()
	node := state.exec.Nodes[idx]
	effectiveType := node.EffectiveType()
	state.mu.Unlock()

	switch effectiveType {
	case NodeTypeCondition:
		// Mark running.
		state.mu.Lock()
		state.exec.Nodes[idx].Status = NodeRunning
		state.exec.CurrentNode = nodeID
		state.mu.Unlock()
		e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

		// Parse condition from config.
		condJSON := node.Config["condition"]
		cond, err := ParseCondition(condJSON)
		if err != nil {
			log.Printf("bmad: invalid condition for node %s: %v", nodeID, err)
			e.failNode(state, idx, nodeID)
			return
		}

		// Evaluate the condition.
		state.mu.Lock()
		nodeOutputsCopy := make(map[string]string, len(state.exec.NodeOutputs))
		for k, v := range state.exec.NodeOutputs {
			nodeOutputsCopy[k] = v
		}
		state.mu.Unlock()

		result := cond.Evaluate(nodeOutputsCopy, repoPath)
		resultStr := "false"
		if result {
			resultStr = "true"
		}

		// Store result in NodeOutputs.
		state.mu.Lock()
		state.exec.NodeOutputs[nodeID] = resultStr
		state.mu.Unlock()

		e.completeNode(state, idx, nodeID)

	case NodeTypeMerge:
		// Merge is a no-op: it completes when in-degree reaches 0.
		state.mu.Lock()
		state.exec.Nodes[idx].Status = NodeRunning
		state.exec.CurrentNode = nodeID
		state.mu.Unlock()
		e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

		e.completeNode(state, idx, nodeID)
	}
}

// executeLoopNode handles loop and loopUntil nodes by repeatedly executing
// body nodes up to maxIterations times. For loopUntil, it checks a condition
// after each iteration and breaks early when the condition is met.
func (e *Executor) executeLoopNode(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath, model string) {
	idx := nodeIndex[nodeID]

	state.mu.Lock()
	node := state.exec.Nodes[idx]
	effectiveType := node.EffectiveType()
	state.exec.Nodes[idx].Status = NodeRunning
	state.exec.CurrentNode = nodeID
	state.mu.Unlock()
	e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

	// Parse maxIterations (default 10, cap at 100).
	maxIter := 10
	if v, ok := node.Config["maxIterations"]; ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			maxIter = n
		}
	}
	if maxIter > 100 {
		maxIter = 100
	}

	// Parse items array (JSON-encoded string list).
	var items []string
	if v, ok := node.Config["items"]; ok && v != "" {
		if err := json.Unmarshal([]byte(v), &items); err != nil {
			log.Printf("bmad: invalid items JSON for node %s: %v", nodeID, err)
			items = nil
		}
	}
	iterateItems := len(items) > 0
	if iterateItems && len(items) < maxIter {
		maxIter = len(items)
	}

	// Parse body node IDs.
	var bodyNodeIDs []string
	if v, ok := node.Config["loopBodyNodes"]; ok && v != "" {
		for _, id := range strings.Split(v, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				bodyNodeIDs = append(bodyNodeIDs, id)
			}
		}
	}

	// Empty body: complete with 0 iterations.
	if len(bodyNodeIDs) == 0 {
		state.mu.Lock()
		state.exec.NodeOutputs[nodeIterKey(nodeID)] = "0"
		state.mu.Unlock()
		e.completeNode(state, idx, nodeID)
		return
	}

	// For loopUntil: parse condition.
	var cond *Condition
	if effectiveType == NodeTypeLoopUntil {
		condJSON, ok := node.Config["condition"]
		if ok && condJSON != "" {
			var err error
			cond, err = ParseCondition(condJSON)
			if err != nil {
				log.Printf("bmad: invalid loop condition for node %s: %v", nodeID, err)
				e.failNode(state, idx, nodeID)
				return
			}
		}
	}

	// Snapshot initial in-degrees for body nodes (for resetting each iteration).
	initialInDegrees := make(map[string]int, len(bodyNodeIDs))
	state.mu.Lock()
	for _, bodyID := range bodyNodeIDs {
		initialInDegrees[bodyID] = state.inDegree[bodyID]
	}
	state.mu.Unlock()

	// Helper: check if a node ID is a body node.
	isBody := make(map[string]bool, len(bodyNodeIDs))
	for _, id := range bodyNodeIDs {
		isBody[id] = true
	}

	// Iterate.
	for iter := 1; iter <= maxIter; iter++ {
		if ctx.Err() != nil {
			e.failNode(state, idx, nodeID)
			return
		}

		// Reset body nodes to pending and restore in-degrees.
		state.mu.Lock()
		for _, bodyID := range bodyNodeIDs {
			bIdx := nodeIndex[bodyID]
			state.exec.Nodes[bIdx].Status = NodePending
			state.inDegree[bodyID] = initialInDegrees[bodyID]
		}
		// Ensure the first body node is ready.
		state.inDegree[bodyNodeIDs[0]] = 0
		// Set current item for this iteration.
		if iterateItems {
			state.exec.NodeOutputs[nodeItemKey(nodeID)] = items[iter-1]
		}
		state.mu.Unlock()

		// Mini ready-set loop for body nodes.
		bodyFailed := false
		for {
			if ctx.Err() != nil {
				e.failNode(state, idx, nodeID)
				return
			}

			state.mu.Lock()
			var bodyReady []string
			for _, bID := range bodyNodeIDs {
				bIdx := nodeIndex[bID]
				if state.exec.Nodes[bIdx].Status == NodePending && state.inDegree[bID] == 0 {
					bodyReady = append(bodyReady, bID)
				}
			}
			state.mu.Unlock()

			if len(bodyReady) == 0 {
				break
			}

			var wg sync.WaitGroup
			for _, bID := range bodyReady {
				wg.Add(1)
				go func(id string) {
					defer wg.Done()
					e.executeNode(ctx, state, nodeIndex, id, repoPath, model)
				}(bID)
			}
			wg.Wait()

			// Update in-degrees for body-internal edges and check for failures.
			state.mu.Lock()
			for _, bID := range bodyReady {
				bIdx := nodeIndex[bID]
				if state.exec.Nodes[bIdx].Status == NodeFailed {
					bodyFailed = true
					break
				}
				if state.exec.Nodes[bIdx].Status == NodeComplete {
					for _, edge := range state.outEdges[bID] {
						if isBody[edge.Target] {
							state.inDegree[edge.Target]--
						}
					}
				}
			}
			state.mu.Unlock()

			if bodyFailed {
				break
			}
		}

		if bodyFailed {
			e.failNode(state, idx, nodeID)
			return
		}

		// Store iteration count.
		state.mu.Lock()
		state.exec.NodeOutputs[nodeIterKey(nodeID)] = strconv.Itoa(iter)
		state.mu.Unlock()
		e.emitEvent("bmad:node:status", NodeStatusEvent{
			ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning, Iteration: iter,
		})

		// For loopUntil: evaluate condition.
		if effectiveType == NodeTypeLoopUntil && cond != nil {
			state.mu.Lock()
			outputsCopy := make(map[string]string, len(state.exec.NodeOutputs))
			for k, v := range state.exec.NodeOutputs {
				outputsCopy[k] = v
			}
			state.mu.Unlock()
			if cond.Evaluate(outputsCopy, repoPath) {
				break
			}
		}
	}

	// Store the full items array as node output for downstream reference.
	if iterateItems {
		state.mu.Lock()
		state.exec.NodeOutputs[nodeID] = node.Config["items"]
		state.mu.Unlock()
	}

	e.completeNode(state, idx, nodeID)
}

// skipBranchLocked recursively marks a node and its downstream nodes as skipped.
// MUST be called with state.mu held.
func (e *Executor) skipBranchLocked(state *execState, nodeIndex map[string]int, nodeID string) {
	idx, ok := nodeIndex[nodeID]
	if !ok {
		return
	}

	// Only skip if still pending and in-degree is 0 (no other active inbound edges).
	if state.exec.Nodes[idx].Status != NodePending {
		return
	}
	if state.inDegree[nodeID] > 0 {
		return // has other active inbound edges, don't skip
	}

	state.exec.Nodes[idx].Status = NodeSkipped
	e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeSkipped})

	// Recursively skip downstream nodes.
	for _, edge := range state.outEdges[nodeID] {
		state.inDegree[edge.Target]--
		e.skipBranchLocked(state, nodeIndex, edge.Target)
	}
}

// skipNode marks a single node as skipped (without recursion).
func (e *Executor) skipNode(state *execState, idx int, nodeID string) {
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeSkipped
	state.mu.Unlock()
	e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeSkipped})
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

// maxCaptureBytes is the maximum size of captured tmux output per node.
const maxCaptureBytes = 102400

// captureOutput captures the tmux pane scrollback for the given target.
// Output is capped at maxCaptureBytes; if larger, the beginning is truncated
// (keeping the tail which contains the final output).
func (e *Executor) captureOutput(ctx context.Context, target string) (string, error) {
	out, err := e.runCmd(ctx, "tmux", "capture-pane", "-t", target, "-p", "-S", "-5000", "-E", "-")
	if err != nil {
		return "", err
	}
	s := string(out)
	if len(s) > maxCaptureBytes {
		s = s[len(s)-maxCaptureBytes:]
	}
	return s, nil
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

	// Snapshot nodes and outputs under lock for context string building.
	state.mu.Lock()
	nodesCopy := make([]WorkflowNode, len(state.exec.Nodes))
	copy(nodesCopy, state.exec.Nodes)
	outputsCopy := make(map[string]string, len(state.exec.NodeOutputs))
	for k, v := range state.exec.NodeOutputs {
		outputsCopy[k] = v
	}
	state.mu.Unlock()

	// Build command.
	contextStr := buildContextStringV3(proc, nodesCopy, nodeIndex, outputsCopy, repoPath)
	command := fmt.Sprintf(`claude --dangerously-skip-permissions --model %s "use %s%s"`, model, proc.SkillName, contextStr)

	// Git errors here are non-fatal: the node continues with DetachedBranch
	// so the workflow still runs when the repo is detached or unreachable.
	// In true detached-HEAD state `git rev-parse --abbrev-ref HEAD` exits 0
	// and prints the literal "HEAD", so treat that case the same as an error.
	branch := DetachedBranch
	if out, err := e.runCmd(ctx, "git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		if trimmed := strings.TrimSpace(string(out)); trimmed != "" && trimmed != "HEAD" {
			branch = trimmed
		}
	}

	sessionName := BuildSessionName(repoPath, branch, nodesCopy[idx].Label, nodeID, time.Now().UnixNano())

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
	// questionPollCounter throttles question scanning to every 3rd tick
	// to reduce tmux subprocess overhead.
	var questionPollCounter int
	for {
		select {
		case <-ctx.Done():
			e.failNode(state, idx, nodeID)
			return
		case <-ticker.C:
			out, err := e.runCmd(ctx, "tmux", "list-panes", "-t", target, "-F", "#{pane_dead}")
			if err != nil {
				// Session gone — try to capture output (best-effort).
				captured, captureErr := e.captureOutput(ctx, target)
				if captureErr != nil {
					log.Printf("bmad: failed to capture output for node %s: %v", nodeID, captureErr)
				}
				state.mu.Lock()
				state.exec.NodeOutputs[nodeID] = captured
				state.mu.Unlock()
				e.completeNode(state, idx, nodeID)
				return
			}
			if strings.TrimSpace(string(out)) == "1" {
				// Capture output before completing (best-effort).
				captured, captureErr := e.captureOutput(ctx, target)
				if captureErr != nil {
					log.Printf("bmad: failed to capture output for node %s: %v", nodeID, captureErr)
				}
				state.mu.Lock()
				state.exec.NodeOutputs[nodeID] = captured
				state.mu.Unlock()
				e.completeNode(state, idx, nodeID)
				return
			}

			// Question detection: scan for questions every 3rd tick to reduce
			// tmux subprocess overhead on the hot path.
			questionPollCounter++
			if questionPollCounter%3 == 0 {
				e.pollForQuestion(ctx, state, nodeID, target)
			}
		}
	}
}

func (e *Executor) completeNode(state *execState, idx int, nodeID string) {
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeComplete
	storyID := state.exec.Nodes[idx].StoryID
	repoPath := state.exec.RepoPath
	node := state.exec.Nodes[idx]
	state.mu.Unlock()
	e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeComplete})

	// Dismiss any stale question notification.
	e.emitEvent(EventQuestionDismissed, map[string]string{
		"execId": state.exec.ID,
		"nodeId": nodeID,
	})

	// Auto-advance linked sprint story status on node completion.
	if storyID != "" && repoPath != "" {
		targetStatus := string(StoryInProgress)
		if err := UpdateStoryStatus(repoPath, storyID, targetStatus); err != nil {
			log.Printf("bmad: failed to update story %s status: %v", storyID, err)
		} else {
			e.emitEvent("bmad:sprint:updated", map[string]string{"storyId": storyID, "status": targetStatus})
		}
	}

	// Artifact verification for process nodes only.
	if node.EffectiveType() == NodeTypeProcess && node.ProcessID != "" {
		proc, ok := ProcessByID(node.ProcessID)
		if ok && len(proc.Outputs) > 0 {
			found, missing := VerifyArtifacts(repoPath, proc.Outputs)
			e.emitEvent("bmad:node:artifacts", NodeArtifactEvent{
				ExecID:  state.exec.ID,
				NodeID:  nodeID,
				Found:   found,
				Missing: missing,
			})
		}
	}
}

func (e *Executor) failNode(state *execState, idx int, nodeID string) {
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeFailed
	state.mu.Unlock()
	e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeFailed})

	// Dismiss any stale question notification.
	e.emitEvent(EventQuestionDismissed, map[string]string{
		"execId": state.exec.ID,
		"nodeId": nodeID,
	})
}

// pollForQuestion captures tmux output and checks for a Claude CLI question.
// If a new question is detected (different hash from last), it emits a bmad:node:question event.
func (e *Executor) pollForQuestion(ctx context.Context, state *execState, nodeID, target string) {
	captured, err := e.captureQuestionOutput(ctx, target)
	if err != nil {
		return // tmux capture failed — skip silently
	}

	question, options, found := detectQuestion(captured)
	if !found {
		return
	}

	qHash := hashQuestion(question)

	state.mu.Lock()
	lastHash := state.lastQuestionHash[nodeID]
	if qHash == lastHash {
		state.mu.Unlock()
		return // duplicate, already emitted
	}
	state.lastQuestionHash[nodeID] = qHash
	execID := state.exec.ID
	repoPath := state.exec.RepoPath
	state.mu.Unlock()

	e.emitEvent(EventQuestion, QuestionEvent{
		ExecID:     execID,
		NodeID:     nodeID,
		RepoPath:   repoPath,
		RepoName:   filepath.Base(repoPath),
		Question:   question,
		Options:    options,
		TmuxTarget: target,
		Timestamp:  time.Now().UnixMilli(),
		QuestionID: qHash,
	})
}

// executeTransformNode reads the output of a source node, applies an extraction
// (regex or line range), and stores the result. Transform nodes are synchronous.
func (e *Executor) executeTransformNode(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID string) {
	idx := nodeIndex[nodeID]

	// Mark running.
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeRunning
	state.exec.CurrentNode = nodeID
	node := state.exec.Nodes[idx]
	state.mu.Unlock()
	e.emitEvent("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

	// Read config.
	sourceNodeID := node.Config["sourceNode"]
	extractType := node.Config["extractType"]
	extractPattern := node.Config["extractPattern"]

	// Get source output.
	state.mu.Lock()
	sourceOutput := state.exec.NodeOutputs[sourceNodeID]
	state.mu.Unlock()

	// Apply extraction.
	var result string
	switch extractType {
	case "regex":
		result = extractRegex(sourceOutput, extractPattern)
	case "lines":
		result = extractLines(sourceOutput, extractPattern)
	default:
		result = sourceOutput // passthrough if unknown type
	}

	// Cap at 100KB.
	if len(result) > maxCaptureBytes {
		result = result[len(result)-maxCaptureBytes:]
	}

	// Store result.
	state.mu.Lock()
	state.exec.NodeOutputs[nodeID] = result
	state.mu.Unlock()

	e.completeNode(state, idx, nodeID)
}

// extractRegex applies a regex to input and returns the first capture group
// (or the full match if no groups). Returns "" on no match or invalid pattern.
func extractRegex(input, pattern string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	matches := re.FindStringSubmatch(input)
	if len(matches) == 0 {
		return ""
	}
	if len(matches) > 1 {
		return matches[1]
	}
	return matches[0]
}

// extractLines extracts lines from input by pattern: "2-4" (range), "-3" (last N), "5" (single).
// Line numbers are 1-indexed.
func extractLines(input, pattern string) string {
	lines := strings.Split(input, "\n")

	// Last N lines: "-3".
	if strings.HasPrefix(pattern, "-") {
		n, err := strconv.Atoi(pattern[1:])
		if err != nil || n <= 0 {
			return ""
		}
		if n > len(lines) {
			n = len(lines)
		}
		return strings.Join(lines[len(lines)-n:], "\n")
	}

	// Range or single: "2-4" or "5".
	parts := strings.SplitN(pattern, "-", 2)
	start, err := strconv.Atoi(parts[0])
	if err != nil || start < 1 {
		return ""
	}
	start-- // convert to 0-indexed

	end := start + 1
	if len(parts) == 2 {
		end, err = strconv.Atoi(parts[1])
		if err != nil {
			return ""
		}
	}

	// Clamp.
	if start >= len(lines) {
		return ""
	}
	if end > len(lines) {
		end = len(lines)
	}

	return strings.Join(lines[start:end], "\n")
}

// buildContextStringV3 builds the context string for a process node, including
// file-aware artifact matching (from upstream processes) and extracted transform data.
// When repoPath is non-empty, it resolves artifact paths on disk and provides
// file-path instructions. When repoPath is empty or the artifact is unmapped,
// it falls back to hint-style messages.
func buildContextStringV3(proc ProcessDef, nodes []WorkflowNode, nodeIndex map[string]int, nodeOutputs map[string]string, repoPath string) string {
	var parts []string

	// Artifact matching with file-path resolution.
	if len(proc.Inputs) > 0 {
		needed := make(map[string]bool)
		for _, input := range proc.Inputs {
			needed[input] = true
		}
		for _, n := range nodes {
			if n.Status != NodeComplete {
				continue
			}
			upstream, ok := ProcessByID(n.ProcessID)
			if !ok {
				continue
			}
			for _, output := range upstream.Outputs {
				if !needed[output] {
					continue
				}
				resolvedPath := ResolveArtifactPath(output, repoPath)
				if resolvedPath == "" {
					// Unmapped artifact — fall back to hint.
					parts = append(parts, fmt.Sprintf(" The upstream process '%s' produced '%s' -- use it as input.", upstream.Name, output))
					continue
				}
				_, err := os.Stat(resolvedPath)
				if err == nil {
					parts = append(parts, fmt.Sprintf(" Read the artifact '%s' from file '%s' and use it as input.", output, resolvedPath))
				} else {
					parts = append(parts, fmt.Sprintf(" The upstream process '%s' should have produced '%s' at '%s' but it was not found. Proceed with best effort.", upstream.Name, output, resolvedPath))
				}
			}
		}
	}

	// Include transform data from completed transform nodes.
	const maxTransformDataLen = 2000
	for _, n := range nodes {
		if n.Status != NodeComplete || n.EffectiveType() != NodeTypeTransform {
			continue
		}
		data := nodeOutputs[n.ID]
		if data == "" {
			continue
		}
		if len(data) > maxTransformDataLen {
			data = data[:maxTransformDataLen]
		}
		parts = append(parts, fmt.Sprintf(" The data transform '%s' extracted: %s", n.Label, data))
	}

	// Include current loop item if a loop with items is active.
	for _, n := range nodes {
		nt := n.EffectiveType()
		if nt != NodeTypeLoop && nt != NodeTypeLoopUntil {
			continue
		}
		item, hasItem := nodeOutputs[nodeItemKey(n.ID)]
		if !hasItem || item == "" {
			continue
		}
		iter := nodeOutputs[nodeIterKey(n.ID)]
		parts = append(parts, fmt.Sprintf(" Currently iterating: item=%q (iteration %s of loop '%s').", item, iter, n.Label))
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
