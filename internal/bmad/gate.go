// Package bmad — iteration gate evaluator.
// Story bmad-interactive-04: Iteration gate + round loop.
package bmad

import (
	"os"
	"strings"
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
		last := lastUserAnswer(state, nodeID)
		if containsToken(g.AcceptTokens, last) {
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

// lastUserAnswer returns the most recent NodeInputHistory entry's Value for
// the given node, or "" when no history exists.
func lastUserAnswer(state *execState, nodeID string) string {
	state.mu.Lock()
	defer state.mu.Unlock()
	hist := state.exec.NodeInputHistory[nodeID]
	if len(hist) == 0 {
		return ""
	}
	return hist[len(hist)-1].Value
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
