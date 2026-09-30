package bmad

import (
	"context"
	"errors"
	"fmt"
	"log"
)

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
		// round's capture so the modal can show what Claude just said. Pass
		// the tmux target so suspendForSpecWithPane runs the pane-activity
		// watchdog: if claude resumes producing output mid-suspension (e.g.
		// the user typed directly into the pane) the suspension aborts with
		// ErrAwaitingPaneActive and we re-run the idle-wait + suspend cycle
		// against the fresh capture instead of leaving the modal stuck on a
		// stale prompt.
		state.mu.Lock()
		lastOutput := ""
		if state.exec.NodeOutputs != nil {
			lastOutput = state.exec.NodeOutputs[lastRoundKey]
		}
		state.mu.Unlock()
		sErr := e.suspendForSpecWithPane(ctx, state, nodeIndex, nodeID, round+1, nextSpec, lastOutput, target)
		if errors.Is(sErr, ErrAwaitingPaneActive) {
			// Pane resumed activity. Skip the answer-injection block and
			// re-enter waitForIdleCompletion at the top of the round loop;
			// the next idle stop will re-suspend with the new capture.
			continue
		}
		if sErr != nil {
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
