package bmad

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

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

	// Build execution with copies of nodes, all pending. Clear prior
	// OutputPaths/InputPaths per architectural decision #1b (stale paths
	// from a previous run must not leak into a fresh execution).
	nodes := make([]WorkflowNode, len(wf.Nodes))
	for i, n := range wf.Nodes {
		nodes[i] = n
		nodes[i].Status = NodePending
		nodes[i].OutputPaths = map[string]string{}
		nodes[i].InputPaths = map[string]string{}
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
		lastOutputHash:   make(map[string]string),
		idleEmitted:      make(map[string]bool),
	}

	e.mu.Lock()
	e.executions[execID] = state
	e.mu.Unlock()

	e.emit("bmad:execution:status", ExecStatusEvent{ExecID: execID, Status: ExecRunning})

	go e.runDynamic(ctx, state, repoPath, model)
	go e.monitorSessionLiveness(ctx, state)

	return execution, nil
}

// sessionDeadPollInterval is how often monitorSessionLiveness checks every
// live node's tmux session. 3s balances responsiveness against tmux IPC load.
const sessionDeadPollInterval = 3 * time.Second

// monitorSessionLiveness polls every tmux target owned by a running or
// awaiting_input node and emits EventSessionDead + clears TmuxTarget the
// first time the session is observed gone. A per-exec seen-set prevents
// re-emission. Exits on ctx.Done.
func (e *Executor) monitorSessionLiveness(ctx context.Context, state *execState) {
	ticker := time.NewTicker(sessionDeadPollInterval)
	defer ticker.Stop()
	dead := make(map[string]bool) // nodeID -> already reported

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			type probe struct{ nodeID, target string }
			var probes []probe
			state.mu.Lock()
			execID := state.exec.ID
			for _, n := range state.exec.Nodes {
				if n.TmuxTarget == "" || dead[n.ID] {
					continue
				}
				if n.Status != NodeRunning && n.Status != NodeAwaitingInput {
					continue
				}
				probes = append(probes, probe{nodeID: n.ID, target: n.TmuxTarget})
			}
			state.mu.Unlock()

			for _, p := range probes {
				// has-session accepts the session name (strip window/pane),
				// returns non-zero when the session is gone.
				session := p.target
				if i := strings.IndexByte(session, ':'); i >= 0 {
					session = session[:i]
				}
				probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				_, err := e.runCmd(probeCtx, "tmux", "has-session", "-t", session)
				cancel()
				if err == nil {
					continue // alive
				}
				dead[p.nodeID] = true
				// Clear TmuxTarget in state so restoreForRepo doesn't rehydrate
				// a stale target; persist so the change survives restart.
				state.mu.Lock()
				for i := range state.exec.Nodes {
					if state.exec.Nodes[i].ID == p.nodeID {
						state.exec.Nodes[i].TmuxTarget = ""
						break
					}
				}
				state.mu.Unlock()
				_ = e.persistSnapshot(state)
				e.emit(EventSessionDead, sessionDeadPayload(execID, p.nodeID, p.target))
			}
		}
	}
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
	e.emit("bmad:execution:status", ExecStatusEvent{ExecID: execID, Status: ExecPaused})
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
	e.emit("bmad:execution:status", ExecStatusEvent{ExecID: execID, Status: ExecRunning})
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
	e.emit("bmad:execution:status", ExecStatusEvent{ExecID: execID, Status: ExecFailed})
	return nil
}

// killWorkflowChainTails kills every unique tmux session held by any node in
// the workflow. Because chained command nodes share a session, bareSessionName
// dedupes to one kill-session call per chain tail. Errors are swallowed — an
// already-dead session is harmless.
func (e *Executor) killWorkflowChainTails(state *execState) {
	state.mu.Lock()
	seen := map[string]struct{}{}
	for _, node := range state.exec.Nodes {
		if node.TmuxTarget == "" {
			continue
		}
		seen[bareSessionName(node.TmuxTarget)] = struct{}{}
	}
	state.mu.Unlock()

	var wg sync.WaitGroup
	for session := range seen {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			killCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_, _ = e.runCmd(killCtx, "tmux", "kill-session", "-t", s)
		}(session)
	}
	wg.Wait()
}

// GetExecution returns a copy of the execution state.
func (e *Executor) GetExecution(execID string) (*WorkflowExecution, error) {
	state, err := e.getState(execID)
	if err != nil {
		return nil, err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	return cloneExecution(state.exec), nil
}

// InteractiveTurn is one half of an exchange during an interactive node —
// either Claude's captured pane output for a round, or the user's answer
// submitted via RespondToInput. The frontend transcript view orders these
// by (Round asc, Role="claude" before Role="user") so the modal reads like
// a chat log.
type InteractiveTurn struct {
	Round     int    `json:"round"`
	Role      string `json:"role"` // "claude" | "user"
	InputID   string `json:"inputId,omitempty"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// GetInteractiveTranscript returns the ordered turn-by-turn history for an
// interactive node: every captured Claude round output plus every user
// answer recorded in NodeInputHistory. Empty slice when the node has no
// activity yet. Used by the modal's transcript pane so users can re-read
// the full conversation while answering the current turn — BMAD turns can
// go deep, and the old single-line prompt hid everything.
func (e *Executor) GetInteractiveTranscript(execID, nodeID string) ([]InteractiveTurn, error) {
	state, err := e.getState(execID)
	if err != nil {
		return nil, err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.exec == nil {
		return nil, nil
	}
	// Collect Claude turns keyed by round from NodeOutputs["{nodeID}-round-N"].
	claudeByRound := map[int]string{}
	prefix := nodeID + "-round-"
	for key, val := range state.exec.NodeOutputs {
		if !strings.HasPrefix(key, prefix) {
			continue
		}
		n, cErr := strconv.Atoi(key[len(prefix):])
		if cErr != nil || n <= 0 {
			continue
		}
		claudeByRound[n] = val
	}
	// User answers from NodeInputHistory preserve round + timestamp + inputID.
	userEntries := state.exec.NodeInputHistory[nodeID]

	// Round cap: whichever the executor considers "current" — NodeRounds is
	// authoritative once the round loop has advanced; otherwise fall back
	// to the max round observed in either map.
	maxRound := 0
	if r, ok := state.exec.NodeRounds[nodeID]; ok {
		maxRound = r
	}
	for n := range claudeByRound {
		if n > maxRound {
			maxRound = n
		}
	}
	for _, entry := range userEntries {
		if entry.Round > maxRound {
			maxRound = entry.Round
		}
	}
	if maxRound == 0 {
		return []InteractiveTurn{}, nil
	}

	turns := make([]InteractiveTurn, 0, maxRound*2)
	for round := 1; round <= maxRound; round++ {
		if content, ok := claudeByRound[round]; ok && content != "" {
			turns = append(turns, InteractiveTurn{
				Round:   round,
				Role:    "claude",
				Content: content,
			})
		}
		for _, entry := range userEntries {
			if entry.Round != round {
				continue
			}
			turns = append(turns, InteractiveTurn{
				Round:     round,
				Role:      "user",
				InputID:   entry.InputID,
				Content:   entry.Value,
				Timestamp: entry.Timestamp,
			})
		}
	}
	return turns, nil
}

// GetCurrentExecution returns a deep copy of the most recently started
// NON-TERMINAL execution (ExecRunning or ExecPaused) whose RepoPath
// equals the given path, or (nil, nil) when no such execution exists.
// A nil return with a nil error means "no current execution" — it is NOT
// treated as a failure so frontend restore-on-mount callers can simply
// fall through to the draft / blank-canvas path.
//
// "Most recently started" is picked deterministically via StartedAt so
// that a sequence of runs for the same repo does not surface a stale
// earlier execution when multiple have entered non-terminal states
// (paused, running). Terminal executions (complete/failed) are ignored
// entirely — there is nothing live to restore from them.
func (e *Executor) GetCurrentExecution(repoPath string) (*WorkflowExecution, error) {
	if repoPath == "" {
		return nil, nil
	}

	// Snapshot the executions map under the exec-level RLock so the scan
	// does not pin the global mutex while touching per-state mutexes.
	e.mu.RLock()
	states := make([]*execState, 0, len(e.executions))
	for _, s := range e.executions {
		states = append(states, s)
	}
	e.mu.RUnlock()

	var best *WorkflowExecution
	var bestStart string // RFC3339 string compare suffices for ordering
	for _, s := range states {
		s.mu.Lock()
		if s.exec != nil &&
			s.exec.RepoPath == repoPath &&
			(s.exec.Status == ExecRunning || s.exec.Status == ExecPaused) {
			if best == nil || s.exec.StartedAt > bestStart {
				best = cloneExecution(s.exec)
				bestStart = s.exec.StartedAt
			}
		}
		s.mu.Unlock()
	}
	return best, nil
}

// cloneExecution returns a deep-copy of a WorkflowExecution so callers
// can read without holding the per-state mutex. Nodes and NodeOutputs are
// copied element-wise; nested maps/slices inside WorkflowNode (config)
// are shared by reference because they are treated as read-only once
// the executor has started a node.
func cloneExecution(src *WorkflowExecution) *WorkflowExecution {
	cp := *src
	cp.Nodes = make([]WorkflowNode, len(src.Nodes))
	copy(cp.Nodes, src.Nodes)
	if src.NodeOutputs != nil {
		cp.NodeOutputs = make(map[string]string, len(src.NodeOutputs))
		for k, v := range src.NodeOutputs {
			cp.NodeOutputs[k] = v
		}
	}
	return &cp
}

// maxAnswerBytes caps the size of a response written into a tmux pane.
// tmux send-keys -l handles long strings but extremely long input may be
// truncated or disrupted by terminal line editing, so we fail fast.
const maxAnswerBytes = 4096

// respondCmdTimeout bounds the tmux subprocess calls issued by RespondToQuestion
// so a hung tmux server cannot block the caller indefinitely.
const respondCmdTimeout = 5 * time.Second

// RespondToQuestionLegacy injects an answer into the Claude CLI tmux pane
// backing the given node. It verifies the pane is still alive, writes the
// literal answer via `tmux send-keys -l`, then dispatches Enter as a second
// call so tmux interprets the keystroke rather than sending the bytes "Enter".
//
// On success the node's cached question hash is cleared so the polling loop
// can re-detect any subsequent question. On failure (not found, dead pane,
// tmux error) the hash is left intact.
//
// This is the pre-schema §3 response path — retained as the fallback for
// autonomous nodes where claude CLI asks an unexpected question. Interactive
// processes use RespondToInput (§8.1) instead.
func (e *Executor) RespondToQuestionLegacy(execID, nodeID, answer string) error {
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
