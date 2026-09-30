package bmad

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// maxCaptureBytes is the maximum size of captured tmux output per node.
const maxCaptureBytes = 102400

// defaultProcessNodeTimeout is the maximum time executeProcessNode waits for
// a claude session to produce output and return to its idle prompt before
// declaring timeout (ErrIdleTimeoutNoStart).
const defaultProcessNodeTimeout = 30 * time.Minute

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

// wrapInBashExec wraps a command inside bash -c '...; exec bash' so the tmux
// pane survives after the inner command exits. Single quotes in the inner
// command are escaped using the portable '\” idiom.
func wrapInBashExec(innerCommand string) string {
	escaped := strings.ReplaceAll(innerCommand, `'`, `'\''`)
	return fmt.Sprintf(`bash -c '%s; exec bash'`, escaped)
}

// waitForIdleCompletion drives a three-stage state machine that detects when
// a claude session has finished producing output and returned to its idle
// prompt. The stages are:
//
//  1. PRIMING — captures the baseline pane hash; idle cannot fire.
//  2. WAITING_FOR_WORK — watches for the hash to change from baseline,
//     indicating claude has started producing output. If the deadline
//     expires before any change, returns ErrIdleTimeoutNoStart.
//  3. WATCHING_FOR_IDLE — waits for the hash to stabilise (same value
//     across two consecutive polls) AND detectIdlePrompt to return true.
//
// Pane death at any stage returns nil (legacy completion path).
// Context cancellation returns ctx.Err().
//
// isInteractive=true disables the hasRecentQuestion gate at stageWatchingForIdle:
// interactive processes (InteractParty/Guided/Iterative) treat a Claude-asked
// question as the CUE to return so the round loop can advance to
// suspendForSpec + UI-AST translation. Autonomous processes keep the gate so
// the legacy idle/question snackbar stays visible.
func (e *Executor) waitForIdleCompletion(ctx context.Context, state *execState, nodeID, target string, timeout time.Duration, isInteractive bool) error {
	type idleStage int
	const (
		stagePriming idleStage = iota
		stageWaitingForWork
		stageWatchingForIdle
	)

	stage := stagePriming
	var baselineHash string
	var lastStableHash string
	deadline := time.Now().Add(timeout)

	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	var questionPollCounter int

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Check pane liveness first — pane death is an implicit
			// completion signal at every stage.
			out, err := e.runCmd(ctx, "tmux", "list-panes", "-t", target, "-F", "#{pane_dead}")
			if err != nil {
				return nil // session gone — treat as completion
			}
			if strings.TrimSpace(string(out)) == "1" {
				return nil // pane dead — legacy completion
			}

			// Capture pane content for hashing and signal detection.
			captured, err := e.captureQuestionOutput(ctx, target)
			if err != nil {
				continue // capture failed — skip this tick
			}
			currentHash := hashCapturedOutput(captured)

			// Fire signal detection so snackbar events (idle/question)
			// continue working. Interactive processes own the snackbar via
			// PendingPrompt + EventAwaitingInput — skip legacy idle/question
			// emissions so the UI does not render two stacked cards per node.
			questionPollCounter++
			scanQuestion := questionPollCounter%questionScanStride == 0
			if !isInteractive {
				e.pollForIdle(state, nodeID, target, captured)
				if scanQuestion {
					e.pollForQuestionFromCapture(state, nodeID, target, captured)
				}
			}

			switch stage {
			case stagePriming:
				baselineHash = currentHash
				stage = stageWaitingForWork

			case stageWaitingForWork:
				if currentHash != baselineHash {
					lastStableHash = currentHash
					stage = stageWatchingForIdle
				} else if time.Now().After(deadline) {
					return ErrIdleTimeoutNoStart
				}

			case stageWatchingForIdle:
				if currentHash == lastStableHash && detectIdlePrompt(captured) {
					// Question gate — autonomous processes only.
					// Interactive processes treat a question as the cue to
					// return; the caller will translate the captured output
					// via the UI-AST adapter and emit a PendingPrompt.
					if !isInteractive && hasRecentQuestion(captured) {
						continue
					}
					return nil // stable + idle prompt → done
				}
				if currentHash != lastStableHash {
					lastStableHash = currentHash
				}
			}
		}
	}
}

// spawnCommandSession handles the tmux spawn sequence shared by process and
// command nodes: wraps the invocation string in bash -c for pane persistence,
// creates a tmux session, records TmuxTarget and StartedAt on the node, and
// emits the initial bmad:node:status with the target.
//
// invocation is the full inner command string (e.g.,
// `claude --dangerously-skip-permissions --model opus "use brainstorming"` for
// process nodes, or `claude --dangerously-skip-permissions --model opus "/simplify"`
// for command nodes).
func (e *Executor) spawnCommandSession(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath, invocation string) (string, error) {
	state.mu.Lock()
	idx, ok := nodeIndex[nodeID]
	label := ""
	if ok {
		label = state.exec.Nodes[idx].Label
	}
	state.mu.Unlock()

	command := wrapInBashExec(invocation)

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

	sessionName := BuildSessionName(repoPath, branch, label, nodeID, time.Now().UnixNano())

	_, err := e.runCmd(ctx, "tmux", "new-session", "-d",
		"-s", sessionName,
		"-c", repoPath,
		command,
	)
	if err != nil {
		return "", err
	}

	target := fmt.Sprintf("%s:0.0", sessionName)
	now := time.Now().UTC().Format(time.RFC3339)
	state.mu.Lock()
	if ok {
		state.exec.Nodes[idx].TmuxTarget = target
		state.exec.Nodes[idx].StartedAt = now
	}
	state.mu.Unlock()
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning, TmuxTarget: target})

	return target, nil
}

// resolveCommandSession scans incoming edges for a command node and returns
// a live parent tmux session target if one exists. This enables session reuse:
// a command node can inject its slash command into an already-running parent
// session rather than spawning a fresh one.
//
// Returns:
//   - (target, true, nil)  — a live parent session was found; caller injects into it
//   - ("", false, nil)     — no live parent; caller should spawn a new session
//   - ("", false, err)     — unexpected error during liveness check
//
// Multi-parent merge is a PUNT: when multiple live parents exist, the most
// recently started (by StartedAt RFC3339 string compare) is picked and a
// warning is logged. Real merge-node semantics are Phase 4+.
func (e *Executor) resolveCommandSession(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID string) (string, bool, error) {
	// Collect parent nodes via incoming edges (outEdges is keyed by source).
	state.mu.Lock()
	edges := make([]WorkflowEdge, 0)
	for _, edgeList := range state.outEdges {
		for _, edge := range edgeList {
			if edge.Target == nodeID {
				edges = append(edges, edge)
			}
		}
	}

	type parentInfo struct {
		target    string
		startedAt string
		nodeID    string
	}
	var candidates []parentInfo
	for _, edge := range edges {
		idx, ok := nodeIndex[edge.Source]
		if !ok {
			continue
		}
		n := state.exec.Nodes[idx]
		if n.TmuxTarget != "" {
			candidates = append(candidates, parentInfo{
				target:    n.TmuxTarget,
				startedAt: n.StartedAt,
				nodeID:    n.ID,
			})
		}
	}
	state.mu.Unlock()

	if len(candidates) == 0 {
		return "", false, nil
	}

	// Check pane liveness for each candidate.
	var live []parentInfo
	for _, c := range candidates {
		out, err := e.runCmd(ctx, "tmux", "list-panes", "-t", c.target, "-F", "#{pane_dead}")
		if err != nil {
			// Session gone entirely — treat as dead.
			continue
		}
		if strings.TrimSpace(string(out)) == "0" {
			live = append(live, c)
		}
		// "1" or any other output = dead pane, skip.
	}

	if len(live) == 0 {
		return "", false, nil
	}

	if len(live) == 1 {
		return live[0].target, true, nil
	}

	// Multiple live parents: pick most recently started (RFC3339 string compare).
	best := live[0]
	for _, c := range live[1:] {
		if c.startedAt > best.startedAt {
			best = c
		}
	}
	log.Printf("bmad: command node %s has %d live parents; reusing most recent (%s)", nodeID, len(live), best.nodeID)
	return best.target, true, nil
}

// injectSlashCommand sends a slash command into an existing tmux pane via
// `tmux send-keys -H`. The payload is `/<commandName>\n` encoded as hex bytes.
func (e *Executor) injectSlashCommand(ctx context.Context, target, commandName string) error {
	data := []byte("/" + commandName + "\n")
	args := make([]string, 0, 4+len(data))
	args = append(args, "send-keys", "-H", "-t", target)
	for _, b := range data {
		args = append(args, fmt.Sprintf("%02x", b))
	}
	_, err := e.runCmd(ctx, "tmux", args...)
	return err
}

// executeCommandNode runs a single command-type node: resolves or spawns a
// session, injects the slash command (if reusing), waits for idle completion,
// captures output, and marks the node complete.
func (e *Executor) executeCommandNode(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath, model string) {
	idx := nodeIndex[nodeID]

	// Mark running and read commandName in a single critical section.
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeRunning
	state.exec.CurrentNode = nodeID
	commandName := state.exec.Nodes[idx].Config["commandName"]
	state.mu.Unlock()
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

	if commandName == "" {
		log.Printf("bmad: command node %s missing commandName in config", nodeID)
		e.failNode(state, idx, nodeID)
		return
	}

	// Resolve an existing parent session.
	target, reused, err := e.resolveCommandSession(ctx, state, nodeIndex, nodeID)
	if err != nil {
		e.failNode(state, idx, nodeID)
		return
	}

	// No live parent → spawn a fresh session with the slash command as initial invocation.
	if !reused {
		innerCommand := fmt.Sprintf(`claude --dangerously-skip-permissions --model %s "/%s"`, model, commandName)
		target, err = e.spawnCommandSession(ctx, state, nodeIndex, nodeID, repoPath, innerCommand)
		if err != nil {
			e.failNode(state, idx, nodeID)
			return
		}
	}

	// Reused session → record target (spawnCommandSession already does this
	// for the spawn path) and inject the slash command.
	if reused {
		state.mu.Lock()
		state.exec.Nodes[idx].TmuxTarget = target
		state.mu.Unlock()
		e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning, TmuxTarget: target})

		if err := e.injectSlashCommand(ctx, target, commandName); err != nil {
			e.failNode(state, idx, nodeID)
			return
		}
	}

	// Wait for idle-prompt completion (or pane death as fallback).
	if err := e.waitForIdleCompletion(ctx, state, nodeID, target, defaultProcessNodeTimeout, false); err != nil {
		e.failNode(state, idx, nodeID)
		return
	}

	// Capture output (best-effort) regardless of completion path.
	captured, captureErr := e.captureOutput(ctx, target)
	if captureErr != nil {
		log.Printf("bmad: failed to capture output for command node %s: %v", nodeID, captureErr)
	}
	state.mu.Lock()
	state.exec.NodeOutputs[nodeID] = captured
	state.mu.Unlock()

	e.completeNode(state, idx, nodeID)
}

// executeProcessNode runs a single process-type node: looks up the process
// definition, builds the claude command, wraps it in bash -c for pane
// persistence, spawns a tmux session, and waits for idle-prompt completion.
func (e *Executor) executeProcessNode(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath, model string) {
	idx := nodeIndex[nodeID]

	// Mark running.
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeRunning
	state.exec.CurrentNode = nodeID
	state.mu.Unlock()
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

	// Look up the process definition (ProcessID is immutable, safe to read after init).
	processID := state.exec.Nodes[idx].ProcessID
	proc, ok := ProcessByID(processID)
	if !ok {
		e.failNode(state, idx, nodeID)
		return
	}

	// Snapshot nodes, outputs, and incoming edges under lock for context string building.
	state.mu.Lock()
	nodesCopy := make([]WorkflowNode, len(state.exec.Nodes))
	copy(nodesCopy, state.exec.Nodes)
	outputsCopy := make(map[string]string, len(state.exec.NodeOutputs))
	for k, v := range state.exec.NodeOutputs {
		outputsCopy[k] = v
	}
	// Collect edges targeting this node for edge-based context passing.
	var incomingEdges []WorkflowEdge
	for _, edgeList := range state.outEdges {
		for _, edge := range edgeList {
			if edge.Target == nodeID {
				incomingEdges = append(incomingEdges, edge)
			}
		}
	}
	state.mu.Unlock()

	// Build the inner claude invocation for this process node.
	contextStr := buildContextStringV3(proc, nodesCopy, nodeIndex, outputsCopy, repoPath, incomingEdges, nodeID)
	innerCommand := fmt.Sprintf(`claude --dangerously-skip-permissions --model %s "use %s%s"`, model, proc.SkillName, contextStr)

	// Spawn the tmux session via the shared helper.
	target, err := e.spawnCommandSession(ctx, state, nodeIndex, nodeID, repoPath, innerCommand)
	if err != nil {
		e.failNode(state, idx, nodeID)
		return
	}

	// Wait for idle-prompt completion (or pane death as fallback).
	waitErr := e.waitForIdleCompletion(ctx, state, nodeID, target, defaultProcessNodeTimeout, false)

	// Capture output (best-effort) regardless of completion path.
	captured, captureErr := e.captureOutput(ctx, target)
	if captureErr != nil {
		log.Printf("bmad: failed to capture output for node %s: %v", nodeID, captureErr)
	}
	state.mu.Lock()
	state.exec.NodeOutputs[nodeID] = captured
	state.mu.Unlock()

	if waitErr != nil {
		e.failNode(state, idx, nodeID)
		return
	}
	e.completeNode(state, idx, nodeID)
}

// questionScanStride controls how often the structured-question detector
// runs relative to the main poll ticker. Idle detection runs on every
// tick; question detection runs on every Nth tick to cap tmux subprocess
// overhead on long-running nodes.
const questionScanStride = 3
