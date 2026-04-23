package bmad

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// paneAlive reports whether the tmux target still addresses a live pane.
// Implementation uses `tmux display-message -p '#{pane_id}'` — on success the
// non-empty pane ID is returned, on dead-pane or missing-session the command
// exits non-zero. An empty target returns false without invoking the runner.
func (e *Executor) paneAlive(target string) bool {
	if target == "" {
		return false
	}
	out, err := e.runCmd(context.Background(), "tmux", "display-message", "-t", target, "-p", "#{pane_id}")
	return err == nil && len(strings.TrimSpace(string(out))) > 0
}

// renderRecap formats a node's prior-round Q&A history as markdown, one line
// per history entry. Used by resumeInteractiveNode to rebuild conversational
// context when a pane is dead (§7.3). Empty or nil history yields "".
func renderRecap(_ ProcessDef, history []NodeInputEntry) string {
	var b strings.Builder
	for _, entry := range history {
		fmt.Fprintf(&b, "**Round %d — %s:** %s\n\n", entry.Round, entry.InputID, entry.Value)
	}
	return b.String()
}

// resumeInteractiveNode reconstructs a dead tmux session for a suspended
// interactive node (§7.3). The recap embeds prior-round Q&A so the agent can
// re-anchor context. When the existing pane is still alive this is a no-op.
func (e *Executor) resumeInteractiveNode(ctx context.Context, state *execState, nodeID string) error {
	nodeIndex := buildNodeIndex(state.exec.Nodes)
	idx, ok := nodeIndex[nodeID]
	if !ok {
		return fmt.Errorf("resumeInteractiveNode: unknown node %q: %w", nodeID, ErrExecNotFound)
	}

	state.mu.Lock()
	node := state.exec.Nodes[idx]
	repoPath := state.exec.RepoPath
	history := append([]NodeInputEntry(nil), state.exec.NodeInputHistory[nodeID]...)
	round := state.exec.NodeRounds[nodeID]
	state.mu.Unlock()

	if e.paneAlive(node.TmuxTarget) {
		return nil
	}

	// Prefer node-local InputSpecs (test override) before the registry.
	proc := ProcessDef{InputSpecs: node.InputSpecs}
	if len(proc.InputSpecs) == 0 {
		p, found := ProcessByID(node.ProcessID)
		if !found {
			return fmt.Errorf("resumeInteractiveNode: %w", ErrProcessNotFound)
		}
		proc = p
	} else {
		// Keep name/description when we have a registered process.
		if p, found := ProcessByID(node.ProcessID); found {
			proc.Name = p.Name
			proc.Description = p.Description
			proc.OutputSpecs = p.OutputSpecs
		}
	}

	resolved, _, err := e.resolveInputs(ctx, state, nodeID, round+1)
	if err != nil {
		return fmt.Errorf("resumeInteractiveNode: resolve inputs: %w", err)
	}

	recap := renderRecap(proc, history)
	prompt := buildInteractivePrompt(proc, resolved, state, nodeID) + "\n\n## Previous session recap\n" + recap

	// Short session name keeps the resume tmux argv concise; the
	// spawnCommandSession path uses a longer repo/branch-prefixed name
	// which is unnecessary here because the node already has identity via
	// nodeID in state.
	sessionName := fmt.Sprintf("resume-%s", nodeID)
	invocation := fmt.Sprintf(`claude --dangerously-skip-permissions %q`, prompt)
	command := wrapInBashExec(invocation)
	if _, err := e.runCmd(ctx, "tmux", "new-session", "-d",
		"-s", sessionName,
		"-c", repoPath,
		command,
	); err != nil {
		return fmt.Errorf("resumeInteractiveNode: spawn session: %w", err)
	}

	target := fmt.Sprintf("%s:0.0", sessionName)
	state.mu.Lock()
	state.exec.Nodes[idx].TmuxTarget = target
	state.mu.Unlock()
	return nil
}

// RehydrateFromSnapshot registers a disk-loaded WorkflowExecution back into
// the executor's in-memory map so subsequent calls to RespondToInput,
// GetExecution, StopWorkflow, etc. find it through getState(execID). Without
// this step, a restored execution has a live snackbar (re-emit in
// GetBmadCurrentExecution) but Send fails with ErrExecNotFound because the
// executor has no memory of the exec (§7.2 resume gap).
//
// Safe to call repeatedly: if the execID already lives in the map the
// existing state is returned unchanged. Spawns a rehydratePending pass so
// each PendingPrompt gains a waiter goroutine. Does NOT re-spawn the round
// loop — the user can respond to the outstanding prompt, but the
// subsequent round will not auto-advance until the executor gains a
// proper resume-from-round-N path (deferred).
func (e *Executor) RehydrateFromSnapshot(exec *WorkflowExecution) (*execState, error) {
	if exec == nil || exec.ID == "" {
		return nil, fmt.Errorf("RehydrateFromSnapshot: nil or missing exec ID")
	}
	e.mu.Lock()
	if existing, ok := e.executions[exec.ID]; ok {
		e.mu.Unlock()
		return existing, nil
	}
	// Build edge + in-degree maps from the snapshot so ready-set math and
	// upstream-context helpers work against the rehydrated state even
	// though no round loop is running.
	inDegree := make(map[string]int, len(exec.Nodes))
	outEdgesMap := map[string][]WorkflowEdge{}
	for _, n := range exec.Nodes {
		inDegree[n.ID] = 0
	}
	state := &execState{
		exec:             exec,
		cancel:           func() {}, // resume has no owning ctx; stop is a no-op
		inDegree:         inDegree,
		outEdges:         outEdgesMap,
		lastQuestionHash: map[string]string{},
		lastOutputHash:   map[string]string{},
		idleEmitted:      map[string]bool{},
	}
	e.executions[exec.ID] = state
	e.mu.Unlock()

	e.rehydratePending(state)
	return state, nil
}

// rehydratePending re-launches a waiter goroutine per PendingPrompt on restore
// so a subsequent RespondToInput can release the blocked responder exactly as
// it would during normal runtime (§7.1). The waiter channel is registered
// synchronously in the parent so tests observe its presence without racing
// the child goroutine's scheduling.
func (e *Executor) rehydratePending(state *execState) {
	state.mu.Lock()
	prompts := append([]PendingPrompt(nil), state.exec.PendingPrompts...)
	state.mu.Unlock()

	nodeIndex := buildNodeIndex(state.exec.Nodes)

	for _, p := range prompts {
		waitCh := state.waiter(p.NodeID, p.InputID)
		idx, ok := nodeIndex[p.NodeID]
		if !ok {
			continue
		}
		go func(p PendingPrompt, waitCh <-chan struct{}, idx int) {
			<-waitCh
			state.mu.Lock()
			state.exec.Nodes[idx].Status = NodeRunning
			state.exec.PendingPrompts = removePrompt(state.exec.PendingPrompts, p.NodeID, p.InputID)
			state.mu.Unlock()
			_ = e.persistSnapshot(state)
		}(p, waitCh, idx)
	}
}

// LoadExecutionFromDisk walks ~/.mashed/workflows/*/execution.json and returns
// the first non-terminal execution whose RepoPath matches. Terminal (complete
// or failed) executions are skipped per §7.2 — there is nothing live to
// restore. Executions whose tmux sessions have died since persistence are
// proactively marked failed on disk so the UI doesn't render a ghost prompt
// from an execution that can never resume.
//
// Returns (nil, nil) when no match exists; I/O errors propagate.
func LoadExecutionFromDisk(repoPath string) (*WorkflowExecution, error) {
	if repoPath == "" {
		return nil, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}
	root := filepath.Join(home, ".mashed", "workflows")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read workflows: %w", err)
	}

	var best *WorkflowExecution
	var bestStart string
	var bestPath string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		snapshotPath := filepath.Join(root, entry.Name(), "execution.json")
		raw, rerr := os.ReadFile(snapshotPath)
		if rerr != nil {
			continue
		}
		var exec WorkflowExecution
		if uErr := json.Unmarshal(raw, &exec); uErr != nil {
			continue
		}
		if exec.RepoPath != repoPath {
			continue
		}
		if exec.Status == ExecComplete || exec.Status == ExecFailed {
			continue
		}
		// Ghost filter: if the execution's tmux targets are all dead, the
		// prompt on disk can never be answered. Mark it failed and skip.
		if !executionTmuxAlive(&exec) {
			markExecutionFailedOnDisk(snapshotPath, &exec)
			continue
		}
		if best == nil || exec.StartedAt > bestStart {
			cp := exec
			best = &cp
			bestStart = exec.StartedAt
			bestPath = snapshotPath
		}
	}
	_ = bestPath // reserved for a future rehydrate-path enrichment
	return best, nil
}

// executionTmuxAlive returns true when at least one running or
// awaiting_input node on exec has a live tmux session. An execution with
// zero tmux-bearing nodes is considered alive (fresh unstarted) — the
// ghost case is specifically "had tmux, now dead".
func executionTmuxAlive(exec *WorkflowExecution) bool {
	hasTarget := false
	for _, n := range exec.Nodes {
		if n.Status != NodeRunning && n.Status != NodeAwaitingInput {
			continue
		}
		if n.TmuxTarget == "" {
			continue
		}
		hasTarget = true
		if tmuxSessionAlive(n.TmuxTarget) {
			return true
		}
	}
	// No tmux-bearing active node — nothing to ghost on. Preserve
	// backward compat: treat as alive so executions without TmuxTarget
	// (e.g. pure utility nodes) keep their current resume behaviour.
	return !hasTarget
}

// tmuxSessionAlive probes `tmux has-session -t <session>` with a 1s
// timeout. Non-zero exit → dead.
func tmuxSessionAlive(target string) bool {
	session := bareSessionName(target)
	if session == "" {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "tmux", "has-session", "-t", session)
	return cmd.Run() == nil
}

// markExecutionFailedOnDisk flips exec.Status = failed and rewrites the
// snapshot file. Errors are swallowed — the worst outcome is that the
// ghost persists for another load cycle, and the next call will retry.
func markExecutionFailedOnDisk(path string, exec *WorkflowExecution) {
	exec.Status = ExecFailed
	raw, err := json.MarshalIndent(exec, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, raw, 0o644)
}
