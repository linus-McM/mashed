package bmad

import (
	"crypto/sha256"
	"encoding/hex"
)

// Event names for interactive suspension / response lifecycle (schema §6).
const (
	EventAwaitingInput = "bmad:node:awaiting_input"
	EventInputResolved = "bmad:node:input_resolved"
	EventInputInvalid  = "bmad:node:input_invalid"
	EventAborted       = "bmad:node:aborted"
)

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
