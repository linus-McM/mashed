// Package bmad — iteration gate evaluator.
// Story bmad-interactive-04: Iteration gate + round loop.
package bmad

import (
	"encoding/json"
	"os"
	"strings"
	"time"
)

// checkGate evaluates whether the iteration loop should exit. It returns
// (hit, reason): hit=true means the loop breaks; reason is a short human
// description for the bmad:node:gate_satisfied event payload.
//
// A nil gate means "single-round process" — the loop runs once and exits.
// For every gate kind, a non-matching evaluation returns (false, "") so the
// caller can fall through to the MaxRounds safety-ceiling check.
func (e *Executor) checkGate(state *execState, g *IterationGate, nodeID string, round int) (bool, string) {
	if g == nil {
		return true, "no gate; single round"
	}
	switch g.Kind {
	case GateUserConfirm:
		// ui-ast-U4 §5.3.2 / §5.3.4 AC-8 + AC-12: walk the last-round window
		// so an accept-token carried by any sub-answer (composite key) — not
		// just the tail entry — satisfies the gate. Legacy pre-U4 entries all
		// share Round==0 and collapse to a full-history walk.
		if anyUserAnswerMatches(state, nodeID, g.AcceptTokens) {
			return true, "accept-token matched"
		}
		return false, ""

	case GateArtifactExists:
		procID := findNodeProcessID(state, nodeID)
		proc, ok := ProcessByID(procID)
		if !ok || len(proc.OutputSpecs) == 0 || proc.OutputSpecs[0].ArtifactName == "" {
			return false, ""
		}
		state.mu.Lock()
		repoPath := state.exec.RepoPath
		state.mu.Unlock()
		path := ResolveArtifactPath(proc.OutputSpecs[0].ArtifactName, repoPath)
		if path == "" {
			return false, ""
		}
		if _, err := os.Stat(path); err == nil {
			return true, "artifact present"
		}
		return false, ""

	case GateExpression:
		// GateExpression evaluation is deferred: the existing condition
		// evaluator (condition.go) consumes JSON, not the free-form
		// CustomExpr string. Parity with condition nodes is tracked for
		// a follow-up story. Until wired, every expression — empty or
		// otherwise — returns (false, "") so the round loop falls
		// through to the MaxRounds safety ceiling.
		// TODO(bmad-interactive-04): wire CustomExpr through a shared
		// expression evaluator once the condition-node syntax is
		// formalised.
		return false, ""

	case GateRoundLimit:
		if g.MaxRounds > 0 && round >= g.MaxRounds {
			return true, "round limit reached"
		}
		return false, ""
	}
	return false, ""
}

// containsToken reports whether answer matches any token in tokens. Comparison
// is case-insensitive after trimming whitespace from both sides. Empty answer
// or empty tokens always return false.
func containsToken(tokens []string, answer string) bool {
	a := strings.TrimSpace(strings.ToLower(answer))
	if a == "" {
		return false
	}
	for _, t := range tokens {
		if strings.ToLower(strings.TrimSpace(t)) == a {
			return true
		}
	}
	return false
}

// anyUserAnswerMatches reports whether any user answer within the last-round
// window (entries whose Round equals the tail entry's Round) matches any of
// the supplied tokens. Legacy pre-U4 snapshots share Round==0 and therefore
// collapse to a full-history walk (ui-ast-U4 §5.3.4 AC-12).
//
// Acquires state.mu internally — callers MUST NOT hold it.
func anyUserAnswerMatches(state *execState, nodeID string, tokens []string) bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	hist := state.exec.NodeInputHistory[nodeID]
	if len(hist) == 0 {
		return false
	}
	lastRound := hist[len(hist)-1].Round
	for i := len(hist) - 1; i >= 0 && hist[i].Round == lastRound; i-- {
		if containsToken(tokens, hist[i].Value) {
			return true
		}
	}
	return false
}

// collectSubAnswersForSpec returns the Values of NodeInputs entries keyed
// under the composite form "<specID>:<sub>" (ui-ast-U4 §5.3.2). Order is
// map-iteration order; callers must not rely on it. Callers MUST hold
// state.mu.
func collectSubAnswersForSpec(state *execState, nodeID, specID string) []string {
	m := state.exec.NodeInputs[nodeID]
	if len(m) == 0 {
		return nil
	}
	prefix := specID + ":"
	out := make([]string, 0, len(m))
	for k, v := range m {
		if strings.HasPrefix(k, prefix) {
			out = append(out, v)
		}
	}
	return out
}

// findNodeProcessID returns the ProcessID of the node matching nodeID, or ""
// when the node is not present in state.exec.Nodes.
func findNodeProcessID(state *execState, nodeID string) string {
	state.mu.Lock()
	defer state.mu.Unlock()
	for _, n := range state.exec.Nodes {
		if n.ID == nodeID {
			return n.ProcessID
		}
	}
	return ""
}

// astStructuredInUse reports whether the node's process opts into the UI AST
// adapter (ui-ast-U4 §5.3.1). Gates flatten-on-receipt so non-migrated
// processes keep legacy bare-key semantics.
//
// Callers MUST hold state.mu — the lookup walks state.exec.Nodes directly.
// The stricter "most recent PendingPrompt had non-empty Structured" predicate
// from the spec narrative is unnecessary here: unmarshal failure in the
// flatten path already degrades gracefully to the legacy branch when the
// frontend submits plain text despite the process opting in.
func astStructuredInUse(state *execState, nodeID string) bool {
	for _, n := range state.exec.Nodes {
		if n.ID == nodeID {
			p, ok := ProcessByID(n.ProcessID)
			return ok && p.EnableAstAdapter
		}
	}
	return false
}

// flattenSubAnswers expands a JSON multi-decision answer into composite
// NodeInputs keys and replaces the bare NodeInputHistory entry just appended
// by RespondToInput with one entry per sub-answer (ui-ast-U4 §5.3.1).
//
// Callers MUST hold state.mu. Unmarshal failure is a no-op so the legacy
// bare-key path remains intact. The raw JSON blob stays under the bare
// specID so sendToSession + upstream readers keep a well-defined value.
func flattenSubAnswers(state *execState, nodeID, specID string, round int, rawAnswer string) {
	var decoded map[string]string
	if err := json.Unmarshal([]byte(rawAnswer), &decoded); err != nil || len(decoded) == 0 {
		return
	}
	if state.exec.NodeInputs == nil {
		state.exec.NodeInputs = map[string]map[string]string{}
	}
	if state.exec.NodeInputs[nodeID] == nil {
		state.exec.NodeInputs[nodeID] = map[string]string{}
	}
	for key, val := range decoded {
		state.exec.NodeInputs[nodeID][specID+":"+key] = val
	}
	state.exec.NodeInputs[nodeID][specID] = rawAnswer

	hist := state.exec.NodeInputHistory[nodeID]
	for i := len(hist) - 1; i >= 0; i-- {
		if hist[i].InputID == specID && hist[i].Round == round && hist[i].Key == "" {
			hist = append(hist[:i], hist[i+1:]...)
			break
		}
	}
	now := time.Now().Unix()
	for key, val := range decoded {
		hist = append(hist, NodeInputEntry{
			InputID:   specID,
			Round:     round,
			Value:     val,
			Timestamp: now,
			Key:       specID + ":" + key,
		})
	}
	if state.exec.NodeInputHistory == nil {
		state.exec.NodeInputHistory = map[string][]NodeInputEntry{}
	}
	state.exec.NodeInputHistory[nodeID] = hist
}
