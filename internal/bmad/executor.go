package bmad

import (
	"context"
	"os/exec"
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

func (e *Executor) getState(execID string) (*execState, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	state, ok := e.executions[execID]
	if !ok {
		return nil, ErrExecNotFound
	}
	return state, nil
}

// ── Interactive process path (bmad-interactive-02) ────────────────────────────
