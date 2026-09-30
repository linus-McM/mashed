package bmad

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// runDynamic drives execution using a dynamic ready-set algorithm.
// It replaces the static tier-based run() with dynamic in-degree tracking
// that supports condition branching and merge nodes.
func (e *Executor) runDynamic(ctx context.Context, state *execState, repoPath, model string) {
	defer func() {
		state.mu.Lock()
		status := state.exec.Status
		state.mu.Unlock()
		if status == ExecComplete || status == ExecFailed {
			e.killWorkflowChainTails(state)
		}
	}()

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
					// Route on ProcessDef.Mode: interactive modes take a
					// separate lifecycle with input resolution + output
					// verification; everything else falls through to the
					// legacy autonomous path byte-for-byte.
					state.mu.Lock()
					procID := state.exec.Nodes[idx].ProcessID
					state.mu.Unlock()
					// util-file-loader is a synchronous utility: read the
					// configured file into NodeOutputs/OutputPaths so
					// downstream nodes receive the contents through the
					// normal context-building path. No tmux, no claude.
					if procID == "util-file-loader" {
						e.executeFileLoader(ctx, state, nodeIndex, nID, repoPath)
					} else if proc, ok := ProcessByID(procID); ok {
						switch proc.Mode {
						case InteractGuided, InteractIterative, InteractParty:
							if testHookExecuteInteractiveNode != nil {
								testHookExecuteInteractiveNode(nID)
							}
							e.executeInteractiveNode(ctx, state, nodeIndex, nID, repoPath, model)
						default:
							if testHookExecuteNode != nil {
								testHookExecuteNode(nID)
							}
							e.executeProcessNode(ctx, state, nodeIndex, nID, repoPath, model)
						}
					} else {
						if testHookExecuteNode != nil {
							testHookExecuteNode(nID)
						}
						e.executeProcessNode(ctx, state, nodeIndex, nID, repoPath, model)
					}
				case NodeTypeCondition, NodeTypeMerge:
					e.executeControlNode(ctx, state, nodeIndex, nID, repoPath, model)
				case NodeTypeTransform:
					e.executeTransformNode(ctx, state, nodeIndex, nID)
				case NodeTypeLoop, NodeTypeLoopUntil:
					e.executeLoopNode(ctx, state, nodeIndex, nID, repoPath, model)
				case NodeTypeCommand:
					e.executeCommandNode(ctx, state, nodeIndex, nID, repoPath, model)
				case NodeTypeMultiFileLoader:
					e.executeMultiFileLoader(ctx, state, nodeIndex, nID, repoPath)
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
			e.emit("bmad:execution:status", ExecStatusEvent{ExecID: state.exec.ID, Status: ExecPaused})
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
			e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: n.ID, Status: NodeSkipped})
		}
	}

	if state.exec.Status == ExecRunning {
		state.exec.Status = ExecComplete
		e.emit("bmad:execution:status", ExecStatusEvent{ExecID: state.exec.ID, Status: ExecComplete})
	}
	state.mu.Unlock()
}

// activeOutEdges returns the outbound edges that should be activated after a node completes.
// For condition nodes, only edges matching the result ("true"/"false") are active.
// For loop nodes, only "loop-exit" edges are active (body edges are managed by executeLoopNode).
// For all other nodes, all outbound edges are active.
func (e *Executor) activeOutEdges(state *execState, nodeID, result string, effectiveType NodeType) []WorkflowEdge {
	// Diagnostic: activeOutEdges should only fire for completed nodes. Log
	// (but do not alter return value) if a caller violates that invariant.
	for _, n := range state.exec.Nodes {
		if n.ID == nodeID {
			if n.Status != NodeComplete {
				log.Printf("bmad: activeOutEdges called for node %s in status %s", nodeID, n.Status)
			}
			break
		}
	}
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
		e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

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
		e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

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
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

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
					e.executeProcessNode(ctx, state, nodeIndex, id, repoPath, model)
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
		e.emit("bmad:node:status", NodeStatusEvent{
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
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeSkipped})

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
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeSkipped})
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
