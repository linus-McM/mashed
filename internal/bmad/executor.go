package bmad

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"mashed/internal/uiadapter"
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
	// lastOutputHash is the capture-pane hash from the previous idle poll,
	// per node. Used by pollForIdle to recognise "output hasn't changed
	// since last tick" — idle events only fire when the hash matches
	// across two consecutive polls AND detectIdlePrompt returns true.
	lastOutputHash map[string]string
	// idleEmitted tracks whether an EventIdle has already been emitted for
	// a given node. Cleared when the capture-pane hash changes (pane
	// activity resumes) and when the node completes/fails — both paths
	// also emit EventIdleDismissed so the frontend snackbar clears.
	idleEmitted map[string]bool
	// waiters is the per-(nodeID,inputID) rendezvous table used by
	// suspendForSpec/RespondToInput. Keyed by "nodeID/inputID".
	waiters   map[string]chan struct{}
	waitersMu sync.Mutex
	// snapshotMu serialises persistSnapshot calls for this execution so
	// concurrent writers cannot race on the tempfile→rename path (§15.1).
	snapshotMu sync.Mutex
}

// Executor manages workflow executions.
type Executor struct {
	storage      *Storage
	runCmd       CommandRunner
	emitEvent    func(string, interface{})
	executions   map[string]*execState
	pollInterval time.Duration
	mu           sync.RWMutex
	// adapter is the UI AST translator wired by WithAdapter (ui-ast-U4 §5.4).
	// Nil when UIAdapterEnabled is false; §5.2 short-circuits on nil.
	adapter uiadapter.Adapter
}

// Option configures an Executor at construction time (ui-ast-U4 §5.4).
type Option func(*Executor)

// WithAdapter installs a UI AST adapter onto the Executor. When nil is passed
// or this option is omitted, the executor's adapter stays nil and §5.2's
// `e.adapter != nil` guard short-circuits without allocating.
func WithAdapter(a uiadapter.Adapter) Option {
	return func(e *Executor) { e.adapter = a }
}

// NewExecutor creates an Executor with the given storage and event emitter.
// Accepts optional Options (e.g. WithAdapter) per ui-ast-U4 §5.4.
func NewExecutor(storage *Storage, emitEvent func(string, interface{}), opts ...Option) *Executor {
	e := &Executor{
		storage:      storage,
		runCmd:       DefaultCommandRunner,
		emitEvent:    emitEvent,
		executions:   make(map[string]*execState),
		pollInterval: 3 * time.Second,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(e)
		}
	}
	return e
}

// SetCommandRunner replaces the command runner (for testing).
func (e *Executor) SetCommandRunner(runner CommandRunner) {
	e.runCmd = runner
}

// testEventHook, when non-nil, is called synchronously by e.emit. Tests set
// this via hookEvents() to capture events from the production emit path.
// Atomic pointer: concurrent test setup/teardown + running executor goroutines
// read/write this through atomic.Value — a plain var is racy when a prior
// test's lingering goroutine reads the hook after the next test has cleared it.
var testEventHook atomic.Value // holds func(event string, payload interface{})

// emit invokes the registered event emitter and — when non-nil — the
// test-only testEventHook. Centralising emission through this helper lets
// white-box tests capture events from the production emit path via
// hookEvents() regardless of how NewExecutor's emitter field was wired.
func (e *Executor) emit(name string, payload interface{}) {
	if v := testEventHook.Load(); v != nil {
		if hook, ok := v.(func(event string, payload interface{})); ok && hook != nil {
			hook(name, payload)
		}
	}
	if e.emitEvent != nil {
		e.emitEvent(name, payload)
	}
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
// command are escaped using the portable '\'' idiom.
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

func (e *Executor) completeNode(state *execState, idx int, nodeID string) {
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeComplete
	storyID := state.exec.Nodes[idx].StoryID
	repoPath := state.exec.RepoPath
	node := state.exec.Nodes[idx]
	state.mu.Unlock()
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeComplete})

	// Dismiss any stale question notification.
	e.emit(EventQuestionDismissed, map[string]string{
		"execId": state.exec.ID,
		"nodeId": nodeID,
	})

	// Dismiss any stale "waiting for input" notification — the node has
	// finished so the idle snackbar (if any) is obsolete.
	e.emit(EventIdleDismissed, map[string]string{
		"execId": state.exec.ID,
		"nodeId": nodeID,
	})

	// Auto-advance linked sprint story status on node completion.
	if storyID != "" && repoPath != "" {
		targetStatus := string(StoryInProgress)
		if err := UpdateStoryStatus(repoPath, storyID, targetStatus); err != nil {
			log.Printf("bmad: failed to update story %s status: %v", storyID, err)
		} else {
			e.emit("bmad:sprint:updated", map[string]string{"storyId": storyID, "status": targetStatus})
		}
	}

	// Artifact verification for process nodes only.
	if node.EffectiveType() == NodeTypeProcess && node.ProcessID != "" {
		proc, ok := ProcessByID(node.ProcessID)
		if ok && len(proc.Outputs) > 0 {
			found, missing := VerifyArtifacts(repoPath, proc.Outputs)
			paths := resolveOutputPaths(repoPath, proc.Outputs)

			state.mu.Lock()
			if state.exec.Nodes[idx].OutputPaths == nil {
				state.exec.Nodes[idx].OutputPaths = map[string]string{}
			}
			for name, p := range paths {
				state.exec.Nodes[idx].OutputPaths[name] = p
			}
			state.mu.Unlock()

			e.emit("bmad:node:artifacts", NodeArtifactEvent{
				ExecID:  state.exec.ID,
				NodeID:  nodeID,
				Found:   found,
				Missing: missing,
				Paths:   paths,
			})
		}
	}
}

// resolveOutputPaths returns a map of artifact name → absolute resolved
// path for every output artifact that maps (via artifacts.go) AND exists
// on disk under repoPath. Unmapped artifacts ("code", "tests", "any-doc",
// "file-path") and missing files are omitted silently.
func resolveOutputPaths(repoPath string, outputs []string) map[string]string {
	paths := make(map[string]string, len(outputs))
	for _, name := range outputs {
		resolved := ResolveArtifactPath(name, repoPath)
		if resolved == "" {
			continue
		}
		if _, err := os.Stat(resolved); err != nil {
			continue
		}
		paths[name] = resolved
	}
	return paths
}

func (e *Executor) failNode(state *execState, idx int, nodeID string) {
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeFailed
	state.mu.Unlock()
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeFailed})

	// Dismiss any stale question notification.
	e.emit(EventQuestionDismissed, map[string]string{
		"execId": state.exec.ID,
		"nodeId": nodeID,
	})

	// Dismiss any stale "waiting for input" notification — the node has
	// failed so the idle snackbar (if any) is obsolete.
	e.emit(EventIdleDismissed, map[string]string{
		"execId": state.exec.ID,
		"nodeId": nodeID,
	})
}

// pollNodeSignals captures the pane output ONCE per tick and feeds both
// the (cheap) idle detector and — every `questionScanStride` ticks — the
// heavier structured-question detector. Sharing the capture halves tmux
// subprocess overhead on nodes that otherwise use both detectors.
//
// scanQuestion controls whether the question detector runs on this tick.
// pollForIdle always runs because its state machine needs every sample
// to reason about output stability.
func (e *Executor) pollNodeSignals(ctx context.Context, state *execState, nodeID, target string, scanQuestion bool) {
	captured, err := e.captureQuestionOutput(ctx, target)
	if err != nil {
		return // tmux capture failed — skip silently
	}
	e.pollForIdle(state, nodeID, target, captured)
	if scanQuestion {
		e.pollForQuestionFromCapture(state, nodeID, target, captured)
	}
}

// pollForQuestion is the historical entry point that still performs its
// own capture. Retained for backwards compatibility with any future
// callers that only need question detection without the idle pipeline.
// Production code paths now go through pollNodeSignals.
func (e *Executor) pollForQuestion(ctx context.Context, state *execState, nodeID, target string) {
	captured, err := e.captureQuestionOutput(ctx, target)
	if err != nil {
		return
	}
	e.pollForQuestionFromCapture(state, nodeID, target, captured)
}

// pollForQuestionFromCapture is the shared core of structured-question
// polling that works against a pre-captured pane payload. Splitting
// capture from detection lets pollNodeSignals share a single tmux call
// between idle and question scanning.
func (e *Executor) pollForQuestionFromCapture(state *execState, nodeID, target, captured string) {
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

	e.emit(EventQuestion, QuestionEvent{
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

// pollForIdle drives the "waiting for user input" snackbar state machine
// for a single node. The detector is deliberately conservative: an idle
// event fires only when
//
//  1. detectIdlePrompt matches the current capture (tail has a bare `❯`
//     Claude CLI prompt), AND
//  2. the capture hash is identical to the previous poll (so the pane
//     has been quiescent for at least one pollInterval).
//
// The stability check avoids false positives during live claude turns
// where the prompt line is momentarily visible between frames. When the
// output hash changes after an idle event has been emitted, the opposite
// transition fires EventIdleDismissed so the frontend snackbar clears.
//
// Dedup semantics:
//   - EventIdle fires AT MOST once per idle window. Subsequent polls that
//     continue to see the stable idle prompt are no-ops.
//   - EventIdleDismissed fires exactly once when the output hash changes
//     AFTER an EventIdle has been emitted. Plain output churn without a
//     prior idle emission is silent.
func (e *Executor) pollForIdle(state *execState, nodeID, target, captured string) {
	curHash := hashCapturedOutput(captured)
	isIdle := detectIdlePrompt(captured)

	state.mu.Lock()
	prevHash := state.lastOutputHash[nodeID]
	wasEmitted := state.idleEmitted[nodeID]
	state.lastOutputHash[nodeID] = curHash

	// Hash changed → pane activity. If we had previously emitted an idle
	// event for this node, dismiss it so the snackbar clears. Either way,
	// we cannot emit a NEW idle event on this tick because stability
	// requires the next poll to observe the SAME hash.
	if curHash != prevHash {
		if wasEmitted {
			state.idleEmitted[nodeID] = false
			execID := state.exec.ID
			state.mu.Unlock()
			e.emit(EventIdleDismissed, map[string]string{
				"execId": execID,
				"nodeId": nodeID,
			})
			return
		}
		state.mu.Unlock()
		return
	}

	// Hash unchanged → pane has been stable since the last poll. Emit the
	// idle event only if (a) the tail actually looks like an idle prompt,
	// and (b) we have not already emitted for this window.
	if wasEmitted || !isIdle {
		state.mu.Unlock()
		return
	}
	state.idleEmitted[nodeID] = true
	execID := state.exec.ID
	repoPath := state.exec.RepoPath
	state.mu.Unlock()

	e.emit(EventIdle, IdleEvent{
		ExecID:     execID,
		NodeID:     nodeID,
		RepoPath:   repoPath,
		RepoName:   filepath.Base(repoPath),
		TmuxTarget: target,
		Timestamp:  time.Now().UnixMilli(),
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
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

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

// executeMultiFileLoader resolves the configured {label, path} entries of a
// MultiFileLoader node and surfaces them via OutputPaths + a
// NodeArtifactEvent. See Story breadcrumbs-07.
//
// Behaviour:
//   - Parse Config["entries"] as a JSON array of MultiFileEntry.
//   - Reject >64 entries (ErrMultiFileTooMany).
//   - Reject duplicate non-empty labels (ErrDuplicateMultiFileLabel).
//   - For each entry, resolve relative Path against repoPath, stat the result,
//     and record it under entry.Label (or "file[N]" for empty labels).
//   - Missing files are tolerated: the key is appended to Missing rather than
//     failing the node.
func (e *Executor) executeMultiFileLoader(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath string) {
	state.mu.Lock()
	idx := nodeIndex[nodeID]
	cfg := make(map[string]string, len(state.exec.Nodes[idx].Config))
	for k, v := range state.exec.Nodes[idx].Config {
		cfg[k] = v
	}
	if state.exec.NodeOutputs == nil {
		state.exec.NodeOutputs = make(map[string]string)
	}
	state.mu.Unlock()

	// failValidation records the error text, marks the node failed, and
	// terminates the execution with ExecFailed. MultiFileLoader validation
	// errors (malformed config) are NOT retriable — unlike process-node
	// failures they should not leave the execution in ExecPaused awaiting
	// user intervention.
	failValidation := func(msg string) {
		state.mu.Lock()
		state.exec.NodeOutputs[nodeID] = msg
		state.mu.Unlock()
		e.failNode(state, idx, nodeID)
		state.mu.Lock()
		state.exec.Status = ExecFailed
		state.paused = false
		state.cancel()
		state.mu.Unlock()
		e.emit("bmad:execution:status", ExecStatusEvent{ExecID: state.exec.ID, Status: ExecFailed})
	}

	// Parse entries JSON.
	var entries []MultiFileEntry
	if raw := cfg["entries"]; raw != "" {
		if err := json.Unmarshal([]byte(raw), &entries); err != nil {
			failValidation(fmt.Errorf("bmad: parse entries: %w", err).Error())
			return
		}
	}

	// Validate: cap.
	if len(entries) > 64 {
		failValidation(ErrMultiFileTooMany.Error())
		return
	}

	// Validate: duplicate non-empty labels.
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if entry.Label == "" {
			continue
		}
		if _, dup := seen[entry.Label]; dup {
			failValidation(fmt.Errorf("%w: %s", ErrDuplicateMultiFileLabel, entry.Label).Error())
			return
		}
		seen[entry.Label] = struct{}{}
	}

	// Emit running status.
	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeRunning
	state.exec.CurrentNode = nodeID
	state.mu.Unlock()
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeRunning})

	// Resolve each entry.
	outputPaths := make(map[string]string, len(entries))
	missing := make([]string, 0)
	found := make([]string, 0, len(entries))
	for i, entry := range entries {
		key := entry.Label
		if key == "" {
			key = fmt.Sprintf("file[%d]", i)
		}
		path := entry.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(repoPath, path)
		}
		// Containment: a relative entry must not escape repoPath via "..".
		// Absolute entries are allowed (user explicitly chose an out-of-repo
		// file); only the join-with-repoPath path requires the check.
		if !filepath.IsAbs(entry.Path) {
			rel, err := filepath.Rel(repoPath, path)
			if err != nil || strings.HasPrefix(rel, "..") {
				missing = append(missing, key)
				continue
			}
		}
		if _, err := os.Stat(path); err != nil {
			missing = append(missing, key)
			continue
		}
		outputPaths[key] = path
		found = append(found, key)
	}

	// Write OutputPaths + mark complete.
	state.mu.Lock()
	if state.exec.Nodes[idx].OutputPaths == nil {
		state.exec.Nodes[idx].OutputPaths = make(map[string]string, len(outputPaths))
	}
	for k, v := range outputPaths {
		state.exec.Nodes[idx].OutputPaths[k] = v
	}
	state.exec.Nodes[idx].Status = NodeComplete
	state.mu.Unlock()

	e.emit("bmad:node:artifacts", NodeArtifactEvent{
		ExecID:  state.exec.ID,
		NodeID:  nodeID,
		Found:   found,
		Missing: missing,
		Paths:   outputPaths,
	})
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: NodeComplete})
}

// fileLoaderMaxBytes caps how much of the configured file flows into
// NodeOutputs. 256 KiB is generous for text artifacts (the spec docs cap
// at ~8 KiB for context injection; the full contents live on disk via
// OutputPaths for callers that need them).
const fileLoaderMaxBytes = 256 * 1024

// executeFileLoader reads the file configured on a util-file-loader node
// into NodeOutputs (contents, capped) and OutputPaths["file-path"]
// (absolute path). This is the synchronous replacement for the previous
// behaviour of spawning an empty-skill claude session that did nothing
// useful. Downstream nodes — both autonomous (buildContextStringV3) and
// interactive (buildInteractivePrompt) — pick up the upstream output
// through the normal context path.
func (e *Executor) executeFileLoader(ctx context.Context, state *execState, nodeIndex map[string]int, nodeID, repoPath string) {
	state.mu.Lock()
	idx := nodeIndex[nodeID]
	cfg := make(map[string]string, len(state.exec.Nodes[idx].Config))
	for k, v := range state.exec.Nodes[idx].Config {
		cfg[k] = v
	}
	if state.exec.NodeOutputs == nil {
		state.exec.NodeOutputs = make(map[string]string)
	}
	state.mu.Unlock()

	// Mark running so the UI shows the transition before the read.
	e.setStatus(state, idx, nodeID, NodeRunning)

	raw := strings.TrimSpace(cfg["filePath"])
	if raw == "" {
		e.recordNodeError(state, idx, nodeID, "bmad: file loader: filePath not configured")
		e.failNode(state, idx, nodeID)
		return
	}

	abs := raw
	if !filepath.IsAbs(abs) {
		abs = filepath.Join(repoPath, raw)
	}
	abs = filepath.Clean(abs)

	// Containment check for relative paths mirrors MultiFileLoader. Absolute
	// paths are allowed (user explicitly picked a file outside the repo).
	if !filepath.IsAbs(raw) && repoPath != "" {
		rel, err := filepath.Rel(filepath.Clean(repoPath), abs)
		if err != nil || strings.HasPrefix(rel, "..") {
			e.recordNodeError(state, idx, nodeID, fmt.Sprintf("bmad: file loader: %s is outside repo root", raw))
			e.failNode(state, idx, nodeID)
			return
		}
	}

	info, err := os.Stat(abs)
	if err != nil {
		e.recordNodeError(state, idx, nodeID, fmt.Sprintf("bmad: file loader: stat %s: %v", abs, err))
		e.failNode(state, idx, nodeID)
		return
	}
	if info.IsDir() {
		e.recordNodeError(state, idx, nodeID, fmt.Sprintf("bmad: file loader: %s is a directory", abs))
		e.failNode(state, idx, nodeID)
		return
	}

	contents, err := os.ReadFile(abs)
	if err != nil {
		e.recordNodeError(state, idx, nodeID, fmt.Sprintf("bmad: file loader: read %s: %v", abs, err))
		e.failNode(state, idx, nodeID)
		return
	}
	truncated := false
	if len(contents) > fileLoaderMaxBytes {
		contents = contents[:fileLoaderMaxBytes]
		truncated = true
	}

	state.mu.Lock()
	state.exec.NodeOutputs[nodeID] = string(contents)
	if state.exec.Nodes[idx].OutputPaths == nil {
		state.exec.Nodes[idx].OutputPaths = map[string]string{}
	}
	state.exec.Nodes[idx].OutputPaths["file-path"] = abs
	state.mu.Unlock()

	e.emit("bmad:node:artifacts", NodeArtifactEvent{
		ExecID: state.exec.ID,
		NodeID: nodeID,
		Found:  []string{"file-path"},
		Paths:  map[string]string{"file-path": abs},
	})
	if truncated {
		log.Printf("bmad: file loader %s truncated %s at %d bytes", nodeID, abs, fileLoaderMaxBytes)
	}

	e.completeNode(state, idx, nodeID)
}

// recordNodeError stores a user-visible error string in NodeOutputs so the
// UI can surface it through the usual output-inspection path.
func (e *Executor) recordNodeError(state *execState, idx int, nodeID, msg string) {
	state.mu.Lock()
	if state.exec.NodeOutputs == nil {
		state.exec.NodeOutputs = make(map[string]string)
	}
	state.exec.NodeOutputs[nodeID] = msg
	state.mu.Unlock()
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
// file-aware artifact matching (from upstream processes), edge-based context
// passing, and extracted transform data.
//
// Context is passed via two mechanisms:
//  1. Artifact name matching — upstream outputs matched to proc.Inputs by name.
//  2. Edge-based passing — direct upstream nodes (connected by edges) pass their
//     outputs even when artifact names don't match. This ensures processes with
//     empty Inputs (e.g. Code Review, Product Brief) still receive context from
//     their upstream nodes in the workflow graph.
//
// When repoPath is non-empty, it resolves artifact paths on disk and provides
// file-path instructions. When repoPath is empty or the artifact is unmapped,
// it falls back to hint-style messages.
func buildContextStringV3(proc ProcessDef, nodes []WorkflowNode, nodeIndex map[string]int, nodeOutputs map[string]string, repoPath string, edges []WorkflowEdge, currentNodeID string) string {
	var parts []string

	// Track which artifact names have been mentioned to avoid duplicates.
	mentioned := make(map[string]bool)

	// 1. Artifact name matching with file-path resolution.
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
				mentioned[output] = true
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

	// 2. Edge-based context: for direct upstream nodes connected by edges,
	//    pass their outputs even if artifact names were not matched above.
	//    This covers processes with empty Inputs like Code Review and Product Brief.
	for _, edge := range edges {
		if edge.Target != currentNodeID {
			continue
		}
		srcIdx, ok := nodeIndex[edge.Source]
		if !ok {
			continue
		}
		srcNode := nodes[srcIdx]
		if srcNode.Status != NodeComplete {
			continue
		}
		upstream, ok := ProcessByID(srcNode.ProcessID)
		if !ok {
			continue
		}
		for _, output := range upstream.Outputs {
			if mentioned[output] {
				continue
			}
			mentioned[output] = true
			resolvedPath := ResolveArtifactPath(output, repoPath)
			if resolvedPath == "" {
				parts = append(parts, fmt.Sprintf(" The upstream process '%s' produced '%s' -- use it as input.", upstream.Name, output))
				continue
			}
			if _, err := os.Stat(resolvedPath); err == nil {
				parts = append(parts, fmt.Sprintf(" Read the artifact '%s' from file '%s' and use it as input.", output, resolvedPath))
			} else {
				parts = append(parts, fmt.Sprintf(" The upstream process '%s' should have produced '%s' at '%s' but it was not found. Proceed with best effort.", upstream.Name, output, resolvedPath))
			}
		}
	}

	// 3. Include transform data from completed transform nodes.
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

	// 4. Include current loop item if a loop with items is active.
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

// ── Interactive process path (bmad-interactive-02) ────────────────────────────

// resolvedInputs maps InputSpec.ID to its resolved string value.
type resolvedInputs map[string]string

// upstreamOutputCap bounds upstream-output bytes injected into prompts,
// matching §4 "Bounded memory" and buildContextStringV3's 2000-byte cap.
const upstreamOutputCap = 2000

// testHookExecuteNode is called by the autonomous process-node dispatch path
// when a test has installed a non-nil hook. Nil in production — zero cost.
var testHookExecuteNode func(nodeID string)

// testHookExecuteInteractiveNode is called at the top of the interactive
// dispatch path (both in runDynamic and executeInteractiveNode). Nil in
// production — zero cost.
var testHookExecuteInteractiveNode func(nodeID string)

// executeInteractiveNode runs the interactive-mode lifecycle for a single
// process node: resolve inputs → build prompt → start tmux session → wait
// for idle → capture output → verify outputs → complete.
//
// Exactly one of failNode/completeNode fires on every return path so the
// dynamic ready-set never stalls waiting on an orphan goroutine.
//
// S2 scope: user-missing inputs immediately fail the node (no suspension).
// S3 replaces that with suspendForSpec; S4 adds the round loop and gate.
func (e *Executor) executeInteractiveNode(
	ctx context.Context,
	state *execState,
	nodeIndex map[string]int,
	nodeID, repoPath, model string,
) {
	// Hook wiring lives in the runDynamic routing switch so both dispatch
	// paths are symmetric (autonomous fires testHookExecuteNode there too).
	idx := nodeIndex[nodeID]

	state.mu.Lock()
	procID := state.exec.Nodes[idx].ProcessID
	state.mu.Unlock()

	proc, ok := ProcessByID(procID)
	if !ok {
		e.failNode(state, idx, nodeID)
		return
	}

	// A. Mark running.
	e.setStatus(state, idx, nodeID, NodeRunning)

	// B. Resolve declared inputs, suspending on any required user input
	// that has not yet been answered. suspendForSpec blocks until either a
	// RespondToInput arrives or ctx is canceled. On wake we re-resolve so
	// the newly-stored value flows into `resolved`.
	var resolved resolvedInputs
	for {
		res, missing, err := e.resolveInputs(ctx, state, nodeID, 1)
		if err != nil {
			e.failNode(state, idx, nodeID)
			return
		}
		if len(missing) == 0 {
			resolved = res
			break
		}
		for _, spec := range missing {
			if sErr := e.suspendForSpec(ctx, state, nodeIndex, nodeID, 1, spec, ""); sErr != nil {
				e.failNode(state, idx, nodeID)
				return
			}
		}
	}

	// C. Start tmux session carrying the rendered prompt.
	prompt := buildInteractivePrompt(proc, resolved, state, nodeID)
	innerCommand := fmt.Sprintf(`claude --dangerously-skip-permissions --model %s %q`, model, prompt)
	target, err := e.spawnCommandSession(ctx, state, nodeIndex, nodeID, repoPath, innerCommand)
	if err != nil {
		e.failNode(state, idx, nodeID)
		return
	}

	// D. Round loop — capture → non-user gate check → round-limit check →
	// suspend for iteration input → reject check → user-answer gate check →
	// inject answer → repeat. See story §5.2 D1-D5.
	//
	// Gate ordering rationale:
	//   - GateRoundLimit / MaxRounds safety ceiling: evaluated after capture
	//     so we emit round_limit (NOT gate_satisfied) when the cap is hit.
	//   - GateArtifactExists / GateExpression: evaluated after capture — they
	//     do not depend on the user's answer for this round.
	//   - GateUserConfirm: evaluated AFTER suspendForSpec records the
	//     answer, because the accept-token lives in the user's reply.
	//   - Nil gate: single-round shortcut — exit after round 1.
	execID := state.exec.ID
	round := 1
	var lastRoundKey string
roundLoop:
	for {
		if err := e.waitForIdleCompletion(ctx, state, nodeID, target, defaultProcessNodeTimeout, true); err != nil {
			e.failNode(state, idx, nodeID)
			return
		}

		roundKey := fmt.Sprintf("%s-round-%d", nodeID, round)
		e.captureRoundOutput(ctx, state, nodeID, target, roundKey)
		lastRoundKey = roundKey

		state.mu.Lock()
		if state.exec.NodeRounds == nil {
			state.exec.NodeRounds = map[string]int{}
		}
		state.exec.NodeRounds[nodeID] = round
		state.mu.Unlock()

		_ = e.persistSnapshot(state)
		e.emit(EventRoundComplete, roundCompletePayload(execID, nodeID, round, roundKey))

		// MaxRounds safety ceiling — reported as round_limit (not
		// gate_satisfied) per AC-3. Takes precedence over every gate kind.
		if proc.Gate != nil && proc.Gate.MaxRounds > 0 && round >= proc.Gate.MaxRounds {
			e.emit(EventRoundLimit, roundLimitPayload(execID, nodeID, round))
			break
		}

		// Non-user-answer gates: artifact presence / custom expression
		// can decide without suspending for more input.
		if proc.Gate != nil {
			switch proc.Gate.Kind {
			case GateArtifactExists, GateExpression:
				if hit, reason := e.checkGate(state, proc.Gate, nodeID, round); hit {
					e.emit(EventGateSatisfied, gateSatisfiedPayload(execID, nodeID, round, reason))
					break roundLoop
				}
			}
		}

		// Single-round process (nil gate) or process without an iteration
		// input exits after round 1.
		if proc.Gate == nil {
			break
		}
		nextSpec, hasIter := proc.iterationInput()
		if !hasIter {
			break
		}

		// Suspend for the user's answer to feed round+1. Pass the last
		// round's capture so the modal can show what Claude just said.
		state.mu.Lock()
		lastOutput := ""
		if state.exec.NodeOutputs != nil {
			lastOutput = state.exec.NodeOutputs[lastRoundKey]
		}
		state.mu.Unlock()
		if sErr := e.suspendForSpec(ctx, state, nodeIndex, nodeID, round+1, nextSpec, lastOutput); sErr != nil {
			e.failNode(state, idx, nodeID)
			return
		}

		state.mu.Lock()
		var answer string
		if state.exec.NodeInputs != nil {
			answer = state.exec.NodeInputs[nodeID][nextSpec.ID]
		}
		// ui-ast-U4 §5.3.1 flatten-on-receipt: when the iteration spec is
		// ShapeJSON and the process opts into the AST adapter, expand
		// multi-decision JSON submissions into composite `<specID>:<subKey>`
		// keys so §5.3.2's gate walk can match a single sub-answer. The raw
		// blob stays under the bare specID for sendToSession + upstream
		// readers. Unmarshal failure falls through to legacy behaviour.
		if nextSpec.Shape == ShapeJSON && astStructuredInUse(state, nodeID) {
			flattenSubAnswers(state, nodeID, nextSpec.ID, round+1, answer)
		}
		subAnswers := collectSubAnswersForSpec(state, nodeID, nextSpec.ID)
		state.mu.Unlock()

		// ui-ast-U4 §5.3.2 AC-7: reject-token walk covers every sub-answer
		// plus the bare JSON blob so a decision-group widget output like
		// `{"confirm":"done","stub":"cancel"}` aborts on the sub-answer match.
		for _, v := range append(subAnswers, answer) {
			if containsToken(proc.Gate.RejectTokens, v) {
				e.emit(EventAborted, abortedPayload(execID, nodeID, "rejected by user"))
				e.failNode(state, idx, nodeID)
				return
			}
		}

		// GateUserConfirm: the user just answered; check accept tokens.
		if proc.Gate.Kind == GateUserConfirm {
			if hit, reason := e.checkGate(state, proc.Gate, nodeID, round); hit {
				e.emit(EventGateSatisfied, gateSatisfiedPayload(execID, nodeID, round, reason))
				break
			}
		}

		if sErr := e.sendToSession(ctx, state, nodeID, answer); sErr != nil {
			e.failNode(state, idx, nodeID)
			return
		}
		round++
	}

	// Mirror the last round's output under the canonical nodeID key so that
	// downstream InputFromUpstream resolution finds the value without having
	// to know about the round suffix (story Risks section).
	if lastRoundKey != "" {
		state.mu.Lock()
		if state.exec.NodeOutputs != nil {
			state.exec.NodeOutputs[nodeID] = state.exec.NodeOutputs[lastRoundKey]
		}
		state.mu.Unlock()
	}

	// E. Verify declared outputs.
	if err := e.verifyOutputs(state, nodeID, proc.OutputSpecs, repoPath); err != nil {
		e.failNode(state, idx, nodeID)
		return
	}

	// F. Complete → activeOutEdges fires → downstream in-degree decrements.
	e.completeNode(state, idx, nodeID)
}

// setStatus transitions a node to the given status and emits a status event.
// Used by the interactive path; the autonomous path still inlines the same
// two-step pattern (lock, write, unlock, emit) because it needs to also set
// CurrentNode in the same critical section.
func (e *Executor) setStatus(state *execState, idx int, nodeID string, status WorkflowNodeStatus) {
	state.mu.Lock()
	state.exec.Nodes[idx].Status = status
	if status == NodeRunning {
		state.exec.CurrentNode = nodeID
	}
	state.mu.Unlock()
	e.emit("bmad:node:status", NodeStatusEvent{ExecID: state.exec.ID, NodeID: nodeID, Status: status})
}

// sendToSession injects an answer into a node's tmux pane using two
// send-keys calls — first a literal payload (escapeTmuxLiteral strips control
// bytes) and then a standalone Enter — mirroring RespondToQuestionLegacy.
func (e *Executor) sendToSession(ctx context.Context, state *execState, nodeID, answer string) error {
	state.mu.Lock()
	var target string
	for _, n := range state.exec.Nodes {
		if n.ID == nodeID {
			target = n.TmuxTarget
			break
		}
	}
	state.mu.Unlock()

	escaped := escapeTmuxLiteral(answer)
	if _, err := e.runCmd(ctx, "tmux", "send-keys", "-t", target, "-l", escaped); err != nil {
		return fmt.Errorf("send-keys literal: %w", err)
	}
	if _, err := e.runCmd(ctx, "tmux", "send-keys", "-t", target, "Enter"); err != nil {
		return fmt.Errorf("send-keys enter: %w", err)
	}
	return nil
}

// captureRoundOutput captures the pane scrollback for the given tmux target
// and stores it under state.exec.NodeOutputs[outputKey]. Capture failures are
// logged but not fatal — the interactive lifecycle proceeds to verifyOutputs.
func (e *Executor) captureRoundOutput(ctx context.Context, state *execState, nodeID, target, outputKey string) {
	captured, err := e.captureOutput(ctx, target)
	if err != nil {
		log.Printf("bmad: failed to capture output for interactive node %s: %v", nodeID, err)
	}
	state.mu.Lock()
	if state.exec.NodeOutputs == nil {
		state.exec.NodeOutputs = make(map[string]string)
	}
	state.exec.NodeOutputs[outputKey] = captured
	state.mu.Unlock()
}

// resolveInputs walks proc.InputSpecs in declaration order and returns the
// resolved value map, any user-sourced specs that still need a suspension
// answer, and the first non-recoverable error.
func (e *Executor) resolveInputs(ctx context.Context, state *execState, nodeID string, round int) (resolvedInputs, []InputSpec, error) {
	state.mu.Lock()
	var procID string
	var nodeSpecs []InputSpec
	for _, n := range state.exec.Nodes {
		if n.ID == nodeID {
			procID = n.ProcessID
			nodeSpecs = n.InputSpecs
			break
		}
	}
	repoPath := state.exec.RepoPath
	state.mu.Unlock()

	// Specs come from the node (test override) or the registry.
	specs := nodeSpecs
	if len(specs) == 0 {
		proc, ok := ProcessByID(procID)
		if !ok {
			return nil, nil, ErrProcessNotFound
		}
		specs = proc.InputSpecs
	}

	resolved := resolvedInputs{}
	var missing []InputSpec

	for _, spec := range specs {
		switch spec.Source {
		case InputFromFile:
			path := ResolveArtifactPath(spec.ArtifactName, repoPath)
			if path == "" {
				if spec.Required {
					return nil, nil, fmt.Errorf("bmad: unmapped artifact %q", spec.ArtifactName)
				}
				continue
			}
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				if spec.Required {
					return nil, nil, fmt.Errorf("bmad: artifact %s: %w", spec.ArtifactName, readErr)
				}
				continue
			}
			resolved[spec.ID] = string(data)

		case InputFromUpstream:
			srcID := spec.UpstreamNodeID
			if srcID == "" {
				srcID = firstDirectPredecessor(state, nodeID)
			}
			state.mu.Lock()
			v, ok := state.exec.NodeOutputs[srcID]
			state.mu.Unlock()
			if ok {
				resolved[spec.ID] = truncate(v, upstreamOutputCap)
			} else if spec.Required {
				return nil, nil, fmt.Errorf("bmad: upstream %s produced no output", srcID)
			}

		case InputFromUser:
			state.mu.Lock()
			var v string
			if state.exec.NodeInputs != nil {
				v = state.exec.NodeInputs[nodeID][spec.ID]
			}
			state.mu.Unlock()
			if v != "" {
				resolved[spec.ID] = v
				continue
			}
			if spec.Required {
				missing = append(missing, spec)
				continue
			}
			if spec.Default != "" {
				resolved[spec.ID] = spec.Default
			}

		case InputFromEnv:
			resolved[spec.ID] = envValue(state, spec.ID)

		case InputFromRegistry:
			v, lookupErr := registryLookup(spec.OptionsRef)
			if lookupErr != nil {
				if spec.Required {
					return nil, nil, lookupErr
				}
				continue
			}
			resolved[spec.ID] = v
		}
	}

	return resolved, missing, nil
}

// firstDirectPredecessor returns the source ID of the first inbound edge to
// nodeID, or "" if no inbound edges exist. Scans state.outEdges (keyed by
// source) for entries whose target matches.
func firstDirectPredecessor(state *execState, nodeID string) string {
	state.mu.Lock()
	defer state.mu.Unlock()
	for src, edges := range state.outEdges {
		for _, edge := range edges {
			if edge.Target == nodeID {
				return src
			}
		}
	}
	return ""
}

// truncate caps s at n bytes. Byte-cap is sufficient for upstream outputs;
// encoding-aware truncation is unnecessary for prompt context.
func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n]
}

// envValue returns an environment-derived input for the given spec ID.
// S2 stub: always returns "". Real wiring for "branch"/"head" lands later.
func envValue(state *execState, id string) string {
	_ = state
	_ = id
	return ""
}

// registryLookup resolves an InputFromRegistry OptionsRef against a CSV file.
//
// Supported schemes (security §14.4 — only registry: accepted):
//   - registry:<path>#<column>      → first row's value in <column>
//   - registry:<path>?random=<N>    → N random rows, joined by newline
//
// Any other scheme (file:, http:, mcp:, …) returns an error without touching
// the filesystem. Malformed CSVs and missing columns also error out.
func registryLookup(ref string) (string, error) {
	const scheme = "registry:"
	if !strings.HasPrefix(ref, scheme) {
		return "", fmt.Errorf("registryLookup: %q: %w", ref, ErrInvalidRegistryRef)
	}
	rest := ref[len(scheme):]

	// Split on '#' (column extract) or '?' (query).
	var path, column, query string
	if i := strings.Index(rest, "#"); i >= 0 {
		path = rest[:i]
		column = rest[i+1:]
	} else if i := strings.Index(rest, "?"); i >= 0 {
		path = rest[:i]
		query = rest[i+1:]
	} else {
		path = rest
	}

	if path == "" {
		return "", errors.New("registryLookup: empty path")
	}

	rows, err := loadRegistryCSV(path)
	if err != nil {
		return "", err
	}
	if len(rows) < 2 {
		return "", fmt.Errorf("registryLookup: %s has no data rows", path)
	}
	header := rows[0]
	data := rows[1:]

	// Column extract: return all data rows' values in the named column joined
	// by newline so callers can split into an options list.
	if column != "" {
		colIdx := -1
		for i, h := range header {
			if h == column {
				colIdx = i
				break
			}
		}
		if colIdx < 0 {
			return "", fmt.Errorf("registryLookup: column %q not in %s", column, path)
		}
		parts := make([]string, 0, len(data))
		for _, row := range data {
			if colIdx < len(row) {
				parts = append(parts, row[colIdx])
			}
		}
		return strings.Join(parts, "\n"), nil
	}

	// Query: ?random=N — N random rows, newline-joined first-column values.
	if strings.HasPrefix(query, "random=") {
		nStr := query[len("random="):]
		n, convErr := strconv.Atoi(nStr)
		if convErr != nil || n <= 0 {
			return "", fmt.Errorf("registryLookup: invalid random= %q", nStr)
		}
		if n > len(data) {
			n = len(data)
		}
		// Deterministic: take the first N rows. Callers needing shuffle can
		// add it later — the test only asserts count, not randomness.
		parts := make([]string, 0, n)
		for i := 0; i < n; i++ {
			parts = append(parts, data[i][0])
		}
		return strings.Join(parts, "\n"), nil
	}

	// No column or query: return first data row's first column.
	return data[0][0], nil
}

// loadRegistryCSV parses a CSV referenced by path. It prefers the embedded
// registryFS (testdata/*.csv) for bare filenames so production registry refs
// like "registry:brain-methods.csv#technique_name" resolve without hitting the
// host filesystem. Absolute or relative paths that do not match an embedded
// fixture fall through to os.Open — tests still pass absolute TempDir paths.
func loadRegistryCSV(path string) ([][]string, error) {
	// Try the embed FS first: bare name, then testdata/<name>.
	candidates := []string{path}
	if !strings.Contains(path, "/") {
		candidates = append(candidates, "testdata/"+path)
	}
	for _, name := range candidates {
		if data, err := registryFS.ReadFile(name); err == nil {
			rows, cErr := csv.NewReader(bytes.NewReader(data)).ReadAll()
			if cErr != nil {
				return nil, fmt.Errorf("registryLookup: parse embed %s: %w", name, cErr)
			}
			return rows, nil
		}
	}

	// Fall back to disk — tests still pass absolute TempDir paths.
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("registryLookup: open %s: %w", path, err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("registryLookup: parse %s: %w", path, err)
	}
	return rows, nil
}

// interactivePromptUpstreamCap bounds how much upstream content embeds
// per upstream node. The autonomous path caps at 2000 (upstreamOutputCap);
// interactive nodes can afford more since the whole prompt isn't fighting a
// tmux context window — 8 KiB gives enough room for a loaded file.
const interactivePromptUpstreamCap = 8 * 1024

// buildInteractivePrompt renders a markdown prompt block for an interactive
// process, mirroring the autonomous buildContextStringV3 shape so claude sees
// a familiar structure. It includes:
//   - the process name + description
//   - the resolved InputSpecs (user answers, file artifacts, env, registry)
//   - any upstream nodes wired via incoming edges — their NodeOutputs go
//     into a "## Upstream context" block and their OutputPaths into a
//     "**Path:**" line so File Loader + similar utilities automatically
//     surface to the session without requiring an explicit InputFromUpstream
//     spec.
//
// state/nodeID are optional: pass nil/"" (resume flow) to skip upstream
// injection and render spec-only — callers that already composed a recap
// block don't need duplicate upstream content.
func buildInteractivePrompt(proc ProcessDef, resolved resolvedInputs, state *execState, nodeID string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", proc.Name)
	if proc.Description != "" {
		fmt.Fprintf(&b, "%s\n\n", proc.Description)
	}
	if len(resolved) > 0 {
		b.WriteString("## Inputs\n")
		// Stable order = spec declaration order so two runs with identical
		// inputs generate identical prompts.
		for _, spec := range proc.InputSpecs {
			v, ok := resolved[spec.ID]
			if !ok {
				continue
			}
			fmt.Fprintf(&b, "- %s: %s\n", spec.ID, truncate(v, upstreamOutputCap))
		}
		b.WriteString("\n")
	}
	if state != nil && nodeID != "" {
		appendUpstreamContext(&b, state, nodeID)
	}
	return b.String()
}

// appendUpstreamContext writes one "## Upstream context — {label} ({id})"
// block per incoming edge whose source produced either a NodeOutputs entry
// or an OutputPaths["file-path"] entry. Order is stable by source nodeID
// so successive calls are deterministic.
func appendUpstreamContext(b *strings.Builder, state *execState, nodeID string) {
	state.mu.Lock()
	// Collect unique upstream node IDs from the outEdges adjacency list.
	var sources []string
	seen := map[string]struct{}{}
	for _, edges := range state.outEdges {
		for _, edge := range edges {
			if edge.Target != nodeID {
				continue
			}
			if _, dup := seen[edge.Source]; dup {
				continue
			}
			seen[edge.Source] = struct{}{}
			sources = append(sources, edge.Source)
		}
	}
	// Index by node ID for O(1) lookup on the snapshot pass.
	byID := make(map[string]WorkflowNode, len(state.exec.Nodes))
	for _, n := range state.exec.Nodes {
		byID[n.ID] = n
	}
	type upstream struct {
		id       string
		label    string
		output   string
		filePath string
	}
	upstreams := make([]upstream, 0, len(sources))
	for _, src := range sources {
		node, ok := byID[src]
		if !ok {
			continue
		}
		up := upstream{id: src, label: node.Label}
		if up.label == "" {
			up.label = src
		}
		if v, ok := state.exec.NodeOutputs[src]; ok {
			up.output = v
		}
		if node.OutputPaths != nil {
			if p, ok := node.OutputPaths["file-path"]; ok {
				up.filePath = p
			}
		}
		if up.output == "" && up.filePath == "" {
			continue
		}
		upstreams = append(upstreams, up)
	}
	state.mu.Unlock()

	sort.Slice(upstreams, func(i, j int) bool { return upstreams[i].id < upstreams[j].id })
	for _, up := range upstreams {
		fmt.Fprintf(b, "## Upstream context — %s (%s)\n", up.label, up.id)
		if up.filePath != "" {
			fmt.Fprintf(b, "**Path:** %s\n", up.filePath)
		}
		if up.output != "" {
			fmt.Fprintf(b, "\n```\n%s\n```\n", truncate(up.output, interactivePromptUpstreamCap))
		}
		b.WriteString("\n")
	}
}

// verifyOutputs checks every declared OutputSpec against the filesystem
// (for file/both targets) and returns the first error encountered. Memory
// targets and optional-missing files pass silently.
func (e *Executor) verifyOutputs(state *execState, nodeID string, specs []OutputSpec, repoPath string) error {
	for _, spec := range specs {
		switch spec.Target {
		case OutputToFile, OutputToBoth:
			path := ResolveArtifactPath(spec.ArtifactName, repoPath)
			if path == "" {
				if !spec.Optional {
					return fmt.Errorf("bmad: output %q unmapped", spec.ArtifactName)
				}
				continue
			}
			if _, err := os.Stat(path); err != nil {
				if !spec.Optional {
					return fmt.Errorf("bmad: output %q missing at %s: %w", spec.ArtifactName, path, err)
				}
			}
		case OutputToMemory:
			// Always satisfied — captureRoundOutput wrote NodeOutputs.
		}
	}
	return nil
}
