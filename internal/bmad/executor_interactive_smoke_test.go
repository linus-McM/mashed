package bmad

// Executor-driven smoke tests for the four canonical interactive processes —
// bmad-brainstorming, bmad-product-brief, bmad-party-mode,
// bmad-advanced-elicitation. Each test exercises executeInteractiveNode via
// the mock CommandRunner defined in testutil_interactive_test.go. AC-3, AC-4,
// AC-5, AC-6, AC-8 of story bmad-interactive-07.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSmoke_Brainstorming_ThreeRoundsDone drives bmad-brainstorming to three
// rounds and accepts via the "done" token. AC-3.
func TestSmoke_Brainstorming_ThreeRoundsDone(t *testing.T) {
	// NOT parallel: testEventHook global is shared across tests.
	h := newInteractiveHarness(t)
	execID, nodeID := h.startSingleNode("bmad-brainstorming")

	// Seed the artifact BEFORE the accept token so verifyOutputs (which runs
	// immediately after the gate fires) finds the file on disk regardless of
	// scheduler timing.
	h.writeArtifact("brainstorm-notes", "## Brainstorm\n- idea one\n")

	// Pre-process required inputs.
	h.expectAwaitingInput(nodeID, "topic")
	h.respond(execID, nodeID, "topic", "voice UX")
	h.expectAwaitingInput(nodeID, "approach")
	h.respond(execID, nodeID, "approach", "ai-recommend")

	// Rounds 1 and 2: continue; round 3: done.
	for r := 1; r <= 2; r++ {
		h.waitForRoundComplete(nodeID, r)
		h.expectAwaitingInput(nodeID, "round-response")
		h.respond(execID, nodeID, "round-response", "keep going")
	}
	h.waitForRoundComplete(nodeID, 3)
	h.expectAwaitingInput(nodeID, "round-response")
	h.respond(execID, nodeID, "round-response", "done")

	h.assertGateSatisfied(nodeID, 3, "accept-token matched")
	h.assertNodeComplete(nodeID)
}

// TestSmoke_Brainstorming_RejectAbort drives bmad-brainstorming and exercises
// the reject-token path ("abort"). AC-8.
func TestSmoke_Brainstorming_RejectAbort(t *testing.T) {
	// NOT parallel: testEventHook global.
	h := newInteractiveHarness(t)
	execID, nodeID := h.startSingleNode("bmad-brainstorming")

	h.expectAwaitingInput(nodeID, "topic")
	h.respond(execID, nodeID, "topic", "voice UX")
	h.expectAwaitingInput(nodeID, "approach")
	h.respond(execID, nodeID, "approach", "ai-recommend")

	h.waitForRoundComplete(nodeID, 1)
	h.expectAwaitingInput(nodeID, "round-response")
	h.respond(execID, nodeID, "round-response", "abort")

	h.assertNodeFailed(nodeID)

	aborted := h.abortedEventsFor(nodeID)
	if assert.NotEmpty(t, aborted, "aborted event must fire") {
		assert.Equal(t, "rejected by user", aborted[0]["reason"])
	}
}

// TestSmoke_ProductBrief_GuidedApproval drives bmad-product-brief through
// mode + final-approval + multiple stage-responses ending with the accept
// token "yes". AC-4.
//
// Implementation note: `final-approval` is Required=true and therefore
// resolved pre-process in the current executor. Gate satisfaction in the
// round loop is driven by the last user answer, which means the "yes" accept
// token must appear as the final stage-response answer — not as the
// final-approval answer. The test design follows the actual executor
// semantics rather than the story's literal phrasing (see S7 developer
// guidance under AC-4).
func TestSmoke_ProductBrief_GuidedApproval(t *testing.T) {
	// NOT parallel: testEventHook global.
	h := newInteractiveHarness(t)
	execID, nodeID := h.startSingleNode("bmad-product-brief")

	// Seed artifact up-front to avoid racing verifyOutputs.
	h.writeArtifact("product-brief", "# Product Brief\n")

	// Pre-process required inputs (declaration order: mode, final-approval).
	h.expectAwaitingInput(nodeID, "mode")
	h.respond(execID, nodeID, "mode", "guided")
	h.expectAwaitingInput(nodeID, "final-approval")
	h.respond(execID, nodeID, "final-approval", "yes")

	// Five stage-response rounds — only the final one carries the accept token.
	for r := 1; r <= 5; r++ {
		h.waitForRoundComplete(nodeID, r)
		h.expectAwaitingInput(nodeID, "stage-response")
		answer := "stage " + string(rune('0'+r))
		if r == 5 {
			answer = "yes" // gate accept token
		}
		h.respond(execID, nodeID, "stage-response", answer)
	}

	h.assertGateSatisfied(nodeID, 5, "accept-token matched")
	h.assertNodeComplete(nodeID)
}

// TestSmoke_PartyMode_ExitToken drives bmad-party-mode through 3 messages
// ending with "exit", and verifies verifyOutputs does NOT fail even though
// the optional transcript artifact is absent. AC-5.
func TestSmoke_PartyMode_ExitToken(t *testing.T) {
	// NOT parallel: testEventHook global.
	h := newInteractiveHarness(t)
	execID, nodeID := h.startSingleNode("bmad-party-mode")

	// Pre-process: topic only.
	h.expectAwaitingInput(nodeID, "topic")
	h.respond(execID, nodeID, "topic", "retro")

	// Three message rounds — third is the exit token.
	answers := []string{"first thought", "second thought", "exit"}
	for r, ans := range answers {
		h.waitForRoundComplete(nodeID, r+1)
		h.expectAwaitingInput(nodeID, "message")
		h.respond(execID, nodeID, "message", ans)
	}

	// Intentionally do NOT writeArtifact — transcript is Optional=true.
	h.assertGateSatisfied(nodeID, 3, "accept-token matched")
	h.assertNodeComplete(nodeID)
}

// TestSmoke_AdvancedElicitation_XAccept drives bmad-advanced-elicitation
// through two rounds: method+apply-changes, then method="x". AC-6.
//
// Uses an InputSpec override to make target-content non-required because the
// single-node smoke harness has no upstream producer. The executor's other
// semantics are unchanged.
func TestSmoke_AdvancedElicitation_XAccept(t *testing.T) {
	// NOT parallel: testEventHook global.
	h := newInteractiveHarness(t)

	// Override target-content to Required=false so the single-node workflow
	// doesn't block waiting for an upstream output. method and apply-changes
	// keep their declared shapes.
	overrides := []InputSpec{
		{ID: "target-content", Source: InputFromUpstream, Required: false},
		{ID: "method", Source: InputFromUser, Shape: ShapeChoice, Prompt: "Pick a reasoning method:", OptionsRef: "registry:methods.csv?random=5"},
		{ID: "apply-changes", Source: InputFromUser, Shape: ShapeApproval, Required: true, Prompt: "Apply these changes?"},
	}
	execID, nodeID := h.startSingleNodeWithOverrides("bmad-advanced-elicitation", overrides)

	// Pre-process: apply-changes is required.
	h.expectAwaitingInput(nodeID, "apply-changes")
	h.respond(execID, nodeID, "apply-changes", "yes")

	// Round 1: method="critique" (a real CSV row, not an accept token).
	h.waitForRoundComplete(nodeID, 1)
	h.expectAwaitingInput(nodeID, "method")
	h.respond(execID, nodeID, "method", "critique")

	// elicitation-notes has Target=OutputToBoth — file must exist for the file
	// half; seed BEFORE the accept token so verifyOutputs finds it.
	h.writeArtifact("elicitation-notes", "## Elicitation\n")

	// Round 2: method="x" — accept token.
	h.waitForRoundComplete(nodeID, 2)
	h.expectAwaitingInput(nodeID, "method")
	h.respond(execID, nodeID, "method", "x")

	h.assertGateSatisfied(nodeID, 2, "accept-token matched")
	h.assertNodeComplete(nodeID)
}

