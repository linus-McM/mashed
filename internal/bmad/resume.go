package bmad

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
// restore. Returns (nil, nil) when no match exists; I/O errors propagate.
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
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		raw, rerr := os.ReadFile(filepath.Join(root, entry.Name(), "execution.json"))
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
		if best == nil || exec.StartedAt > bestStart {
			cp := exec
			best = &cp
			bestStart = exec.StartedAt
		}
	}
	return best, nil
}
