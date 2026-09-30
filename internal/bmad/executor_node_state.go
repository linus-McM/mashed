package bmad

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

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
