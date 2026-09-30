package bmad

import (
	"crypto/sha256"
	"encoding/hex"
)

// Event names for interactive suspension / response lifecycle (schema §6).
const (
	EventAwaitingInput     = "bmad:node:awaiting_input"
	EventAwaitingDismissed = "bmad:node:awaiting_dismissed"
	EventInputResolved     = "bmad:node:input_resolved"
	EventInputInvalid      = "bmad:node:input_invalid"
	EventAborted           = "bmad:node:aborted"

	// Iteration loop events (schema §6, story bmad-interactive-04).
	EventRoundComplete = "bmad:node:round_complete"
	EventGateSatisfied = "bmad:node:gate_satisfied"
	EventRoundLimit    = "bmad:node:round_limit"

	// EventSessionDead fires when the per-exec liveness poller observes that
	// a node's tmux pane has died while the node is still marked running or
	// awaiting_input. Frontend clears the tmuxTarget + shows an "ended" badge.
	EventSessionDead = "bmad:node:session_dead"
)

// roundCompletePayload returns the payload for EventRoundComplete.
func roundCompletePayload(execID, nodeID string, round int, outputKey string) map[string]any {
	return map[string]any{
		"execId":    execID,
		"nodeId":    nodeID,
		"round":     round,
		"outputKey": outputKey,
	}
}

// gateSatisfiedPayload returns the payload for EventGateSatisfied.
func gateSatisfiedPayload(execID, nodeID string, round int, reason string) map[string]any {
	return map[string]any{
		"execId": execID,
		"nodeId": nodeID,
		"round":  round,
		"reason": reason,
	}
}

// roundLimitPayload returns the payload for EventRoundLimit.
func roundLimitPayload(execID, nodeID string, round int) map[string]any {
	return map[string]any{
		"execId": execID,
		"nodeId": nodeID,
		"round":  round,
	}
}

// valueHashLen is the number of hex chars kept from the SHA-256 digest for
// event payloads (§14.3). 16 hex chars = 64 bits — collision-safe and compact.
const valueHashLen = 16

// sha256hex returns the first N hex chars of SHA-256(s).
func sha256hex(s string, n int) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])[:n]
}

// awaitingPayload returns the payload for EventAwaitingInput. The PendingPrompt
// is returned verbatim — no hashing needed because the prompt text is not
// sensitive user data.
func awaitingPayload(p PendingPrompt) PendingPrompt { return p }

// awaitingDismissedPayload returns the payload for EventAwaitingDismissed. The
// node has been demoted from NodeAwaitingInput back to NodeRunning because the
// tmux pane resumed activity. Frontend clears the modal + restores the running
// badge.
func awaitingDismissedPayload(execID, nodeID, inputID string, round int, reason string) map[string]any {
	return map[string]any{
		"execId":  execID,
		"nodeId":  nodeID,
		"inputId": inputID,
		"round":   round,
		"reason":  reason,
	}
}

// inputResolvedPayload returns the payload for EventInputResolved. Per §14.3
// the raw value NEVER appears on the event bus — only a short SHA-256 hash.
func inputResolvedPayload(execID, nodeID, inputID string, round int, value string) map[string]any {
	return map[string]any{
		"execId":    execID,
		"nodeId":    nodeID,
		"inputId":   inputID,
		"round":     round,
		"valueHash": sha256hex(value, valueHashLen),
	}
}

// invalidPayload returns the payload for EventInputInvalid.
func invalidPayload(execID, nodeID, inputID, reason string) map[string]any {
	return map[string]any{
		"execId":  execID,
		"nodeId":  nodeID,
		"inputId": inputID,
		"reason":  reason,
	}
}

// abortedPayload returns the payload for EventAborted.
func abortedPayload(execID, nodeID, reason string) map[string]any {
	return map[string]any{
		"execId": execID,
		"nodeId": nodeID,
		"reason": reason,
	}
}

// sessionDeadPayload returns the payload for EventSessionDead.
func sessionDeadPayload(execID, nodeID, tmuxTarget string) map[string]any {
	return map[string]any{
		"execId":     execID,
		"nodeId":     nodeID,
		"tmuxTarget": tmuxTarget,
	}
}
