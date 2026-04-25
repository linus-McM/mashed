//go:build testing

// Package bmad tests for the UI AST adapter wiring on the Executor.
// Story ui-ast-U4 Task 2 (AC-11).
//
// RED Phase: WithAdapter and Executor.adapter do not exist yet. Compile
// failure on this file is the intended RED signal — both tests here fail
// to build until T2-GREEN adds the option constructor and Executor field.
// TestWithAdapter_OptionAppliesAdapter remains behavior-RED through T3-GREEN
// (when suspendForSpec gains the §5.2 wiring that populates Structured).
//
// Runs only under `-tags testing` because MockAdapter lives behind that
// build tag (internal/uiadapter/mock.go).
package bmad

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/uiadapter"
)

// ── Test-local adapters (unit-local, not published to uiadapter) ───────────────

// delayAdapter returns fixed after blocking for delay, honouring ctx. On ctx
// cancellation it returns a canceled-fallback AST whose Diagnostics.CancelReason
// is populated from ctx.Err() — mirrors the defaultAdapter "canceled" path so
// §5.2's cancel-detection branch sees a realistic payload.
type delayAdapter struct {
	delay time.Duration
	fixed *uiadapter.UIAST
}

func (d *delayAdapter) Translate(ctx context.Context, raw, _ string) *uiadapter.UIAST {
	select {
	case <-time.After(d.delay):
		if d.fixed != nil {
			return d.fixed
		}
		return uiadapter.FallbackAST(raw, "mock:nil", nil)
	case <-ctx.Done():
		ast := uiadapter.FallbackAST(raw, "canceled", nil)
		ast.Diagnostics.CancelReason = ctx.Err().Error()
		return ast
	}
}

// newOversizeAST returns an AST whose JSON marshal exceeds sizeBytes. The blob
// lives in a single markdown node's Content, so the envelope is well-formed —
// bypassing the U2 Validator "oversize" reason still leaves a post-Translate
// blob that the T3-GREEN size guard must reject.
func newOversizeAST(sizeBytes int) *uiadapter.UIAST {
	return &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "mock:oversize",
		Nodes: []uiadapter.UINode{
			{Type: "markdown", Content: strings.Repeat("x", sizeBytes)},
		},
	}
}

// drainAfter cancels ctx and drains n entries from errCh, each with a 2s
// timeout fallback. Shared by every test that launches suspendForSpec in a
// goroutine so goroutine exit is guaranteed under t.Cleanup.
func drainAfter(cancel context.CancelFunc, errCh <-chan error, n int) func() {
	return func() {
		cancel()
		for i := 0; i < n; i++ {
			select {
			case <-errCh:
			case <-time.After(2 * time.Second):
			}
		}
	}
}

// ── AC-11: nil adapter short-circuits cleanly ─────────────────────────────────

// TestExecutor_NilAdapter_ShortCircuits verifies AC-11:
// Given an Executor constructed without WithAdapter on an EnableAstAdapter:true
// process, suspendForSpec must leave PendingPrompt.Structured == "" and must
// not panic — even when lastOutput is non-empty (which would otherwise trigger
// the adapter call per §5.2).
func TestExecutor_NilAdapter_ShortCircuits(t *testing.T) {
	// NOT parallel: uses testEventHook and testRegistry globals.
	const nodeID = "n1"
	const processID = "test-u4-nil-adapter"
	spec := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: true,
		Prompt:   "Enter a topic",
	}
	registerTestProcess(t, ProcessDef{
		ID:               processID,
		Name:             "U4 Nil Adapter",
		Mode:             InteractGuided,
		SkillName:        "test-skill",
		InputSpecs:       []InputSpec{spec},
		OutputSpecs:      []OutputSpec{{ID: "out1", Target: OutputToMemory}},
		EnableAstAdapter: true,
	})

	state, nodeIndex := newSuspendState(nodeID, processID, spec)
	state.exec.RepoPath = t.TempDir()

	// No WithAdapter — e.adapter stays nil. lastOutput is non-empty so that
	// §5.2's `lastOutput != ""` branch would fire if the adapter were set.
	e := NewExecutor(nil, func(name string, data interface{}) {})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	t.Cleanup(drainAfter(cancel, errCh, 1))
	go func() {
		// Nil-adapter panic would surface here and fail the test.
		errCh <- e.suspendForSpec(ctx, state, nodeIndex, nodeID, 2, spec, "Claude said something interesting")
	}()

	require.True(t,
		pollForStatus(state, nodeIndex, nodeID, NodeAwaitingInput, 500*time.Millisecond),
		"node must reach NodeAwaitingInput within 500ms")

	state.mu.Lock()
	prompts := make([]PendingPrompt, len(state.exec.PendingPrompts))
	copy(prompts, state.exec.PendingPrompts)
	state.mu.Unlock()

	require.Len(t, prompts, 1, "exactly one PendingPrompt expected")
	assert.Empty(t, prompts[0].Structured,
		"Structured must stay empty when Executor has no adapter (nil-adapter short-circuit)")
}

// ── AC-11 second half: WithAdapter option wires the adapter ──────────────────

// TestWithAdapter_OptionAppliesAdapter verifies the WithAdapter constructor
// option actually installs the adapter on the Executor — proven via public
// behavior: once the adapter is wired and suspendForSpec runs with a non-empty
// lastOutput, PendingPrompt.Structured must carry the marshaled AST.
//
// RED until T2-GREEN (WithAdapter defined) and T3-GREEN (suspendForSpec
// populates Structured). Post-T3-GREEN the MockAdapter's fixed AST surfaces
// as the PendingPrompt.Structured payload.
func TestWithAdapter_OptionAppliesAdapter(t *testing.T) {
	// NOT parallel: uses testRegistry globals.
	const nodeID = "n1"
	const processID = "test-u4-with-adapter"
	spec := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: true,
		Prompt:   "Enter a topic",
	}
	registerTestProcess(t, ProcessDef{
		ID:               processID,
		Name:             "U4 With Adapter",
		Mode:             InteractGuided,
		SkillName:        "test-skill",
		InputSpecs:       []InputSpec{spec},
		OutputSpecs:      []OutputSpec{{ID: "out1", Target: OutputToMemory}},
		EnableAstAdapter: true,
	})

	// Fixed AST with a recognisable GeneratedBy sentinel so the Structured
	// assertion is specific, not a mere non-empty check.
	fixed := &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "mock:u4-with-adapter",
		TurnSummary: "u4 with-adapter test",
		Nodes: []uiadapter.UINode{
			{Type: "markdown", Content: "hello"},
		},
	}
	mock := uiadapter.NewMock(fixed, nil)

	state, nodeIndex := newSuspendState(nodeID, processID, spec)
	state.exec.RepoPath = t.TempDir()

	e := NewExecutor(nil, func(name string, data interface{}) {}, WithAdapter(mock))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	t.Cleanup(drainAfter(cancel, errCh, 1))
	go func() {
		errCh <- e.suspendForSpec(ctx, state, nodeIndex, nodeID, 2, spec, "Claude replied with context")
	}()

	require.True(t,
		pollForStatus(state, nodeIndex, nodeID, NodeAwaitingInput, 500*time.Millisecond),
		"node must reach NodeAwaitingInput within 500ms")

	state.mu.Lock()
	prompts := make([]PendingPrompt, len(state.exec.PendingPrompts))
	copy(prompts, state.exec.PendingPrompts)
	state.mu.Unlock()

	require.Len(t, prompts, 1, "exactly one PendingPrompt expected")
	structured := prompts[0].Structured
	require.NotEmpty(t, structured,
		"Structured must be populated when WithAdapter installs a working adapter")
	assert.Contains(t, structured, "mock:u4-with-adapter",
		"Structured must contain the mock AST's GeneratedBy sentinel")
}

// ── Shared RED-test harness for Task 3 suspendForSpec wiring ──────────────────

// registerU4Process installs a ProcessDef for the AC-2..5 tests. Single spec,
// ShapeFree, EnableAstAdapter:true — the common shape every T3-RED case needs.
func registerU4Process(t *testing.T, processID string) InputSpec {
	t.Helper()
	spec := InputSpec{
		ID:       "topic",
		Source:   InputFromUser,
		Shape:    ShapeFree,
		Required: true,
		Prompt:   "Enter a topic",
	}
	registerTestProcess(t, ProcessDef{
		ID:               processID,
		Name:             processID,
		Mode:             InteractGuided,
		SkillName:        "test-skill",
		InputSpecs:       []InputSpec{spec},
		OutputSpecs:      []OutputSpec{{ID: "out1", Target: OutputToMemory}},
		EnableAstAdapter: true,
	})
	return spec
}

// ── AC-2: adapter Translate runs OUTSIDE state.mu ─────────────────────────────

// TestSuspendForSpec_TranslateOutsideLock verifies AC-2: two concurrent
// suspendForSpec calls on the SAME execState with a blocking adapter must
// BOTH reach NodeAwaitingInput within a budget tighter than 2× the adapter
// delay. Serial execution would imply Translate ran inside state.mu.
//
// Timing shape: delay=300ms, budget=450ms. Parallel translates finish in
// ~300-350ms (150ms headroom under the gate); serial would require ~600ms
// (150ms above the gate), so both false-positive and false-negative margins
// exceed typical CI scheduling noise.
func TestSuspendForSpec_TranslateOutsideLock(t *testing.T) {
	// NOT parallel: uses testRegistry globals.
	const processID = "test-u4-translate-outside-lock"
	const nodeA = "n-a"
	const nodeB = "n-b"
	spec := registerU4Process(t, processID)

	// Two-node execState sharing one state.mu.
	nodes := []WorkflowNode{
		{ID: nodeA, ProcessID: processID, Label: "A", Status: NodePending, NodeType: NodeTypeProcess, InputSpecs: []InputSpec{spec}},
		{ID: nodeB, ProcessID: processID, Label: "B", Status: NodePending, NodeType: NodeTypeProcess, InputSpecs: []InputSpec{spec}},
	}
	state, nodeIndex := newSessionState(nodes, nil)
	state.exec.RepoPath = t.TempDir()

	fixed := &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "mock:u4-outside-lock",
		Nodes:       []uiadapter.UINode{{Type: "markdown", Content: "ok"}},
	}
	const adapterDelay = 300 * time.Millisecond
	adapter := &delayAdapter{delay: adapterDelay, fixed: fixed}

	e := NewExecutor(nil, func(name string, data interface{}) {}, WithAdapter(adapter))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 2)
	t.Cleanup(drainAfter(cancel, errCh, 2))

	// Budget = 1.5 × adapterDelay. Preserves the 2× signal (serial would
	// take ~2 × 300 = 600ms) with ~150ms headroom on each side.
	const budget = 450 * time.Millisecond

	start := time.Now()
	for _, nodeID := range []string{nodeA, nodeB} {
		id := nodeID
		go func() {
			errCh <- e.suspendForSpec(ctx, state, nodeIndex, id, 2, spec, "last output "+id)
		}()
	}

	deadline := time.Now().Add(budget)
	bothReady := false
	for time.Now().Before(deadline) {
		state.mu.Lock()
		sa := state.exec.Nodes[nodeIndex[nodeA]].Status
		sb := state.exec.Nodes[nodeIndex[nodeB]].Status
		state.mu.Unlock()
		if sa == NodeAwaitingInput && sb == NodeAwaitingInput {
			bothReady = true
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	elapsed := time.Since(start)

	require.True(t, bothReady,
		"both nodes must reach NodeAwaitingInput within %v (elapsed=%v) — "+
			"serial would imply adapter held state.mu", budget, elapsed)
	assert.Less(t, elapsed, budget,
		"concurrent adapter calls must overlap (elapsed=%v); serial timing "+
			"(~%v) would indicate Translate ran inside state.mu",
		elapsed, 2*adapterDelay)

	// Both prompts must carry Structured — proves the adapter was consulted,
	// not a nil-adapter fast path.
	state.mu.Lock()
	prompts := make([]PendingPrompt, len(state.exec.PendingPrompts))
	copy(prompts, state.exec.PendingPrompts)
	state.mu.Unlock()

	require.Len(t, prompts, 2, "exactly two PendingPrompts expected")
	for _, p := range prompts {
		assert.NotEmpty(t, p.Structured,
			"Structured must be populated on prompt %s/%s — adapter must have been consulted",
			p.NodeID, p.InputID)
	}
}

// ── AC-3: ctx cancellation mid-Translate returns ctx.Err() ────────────────────

// TestSuspendForSpec_CancelDuringTranslate_Returns verifies AC-3: when ctx
// is canceled mid-Translate, suspendForSpec returns ctx.Err() and does NOT
// register a PendingPrompt (§5.2 post-Translate ctx.Err() guard).
func TestSuspendForSpec_CancelDuringTranslate_Returns(t *testing.T) {
	// NOT parallel: uses testRegistry globals.
	const nodeID = "n1"
	const processID = "test-u4-cancel-mid-translate"
	spec := registerU4Process(t, processID)

	state, nodeIndex := newSuspendState(nodeID, processID, spec)
	state.exec.RepoPath = t.TempDir()

	// 500ms translate window; we cancel 100ms in.
	adapter := &delayAdapter{delay: 500 * time.Millisecond}
	e := NewExecutor(nil, func(name string, data interface{}) {}, WithAdapter(adapter))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- e.suspendForSpec(ctx, state, nodeIndex, nodeID, 2, spec, "claude output mid-translate")
	}()

	// Give Translate ~100ms to enter its select before canceling. If the
	// goroutine is slow to enter Translate, cancel still fires before the
	// post-Translate ctx.Err() guard — either path exercises the §5.2 abort.
	time.Sleep(100 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		require.ErrorIs(t, err, context.Canceled,
			"suspendForSpec must return ctx.Err() when ctx is canceled mid-Translate, got %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("suspendForSpec did not return within 2s of cancellation")
	}

	// PendingPrompts must be untouched — §5.2's ctx.Err() guard aborts BEFORE
	// upsertPrompt runs when cancellation wins.
	state.mu.Lock()
	pending := len(state.exec.PendingPrompts)
	state.mu.Unlock()
	assert.Zero(t, pending,
		"PendingPrompts must stay empty when ctx is canceled mid-Translate (got %d)", pending)
}

// ── AC-4: Party-Mode first user-facing turn (round=2) gets Structured ─────────

// TestSuspendForSpec_PartyModeRound1UserTurn_HasStructured verifies AC-4:
// Party-Mode's first user-facing turn (round=2, lastOutput non-empty) must
// populate PendingPrompt.Structured, proving the `round > 1` gate in §5.2
// line 98 is NOT reintroduced.
func TestSuspendForSpec_PartyModeRound1UserTurn_HasStructured(t *testing.T) {
	// NOT parallel: uses testRegistry globals.
	const nodeID = "n1"
	const processID = "test-u4-party-mode-round2"
	spec := registerU4Process(t, processID)

	// Recognisable sentinel so Structured assertion is specific.
	fixed := &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "mock:party-mode-round2",
		TurnSummary: "first user-facing turn",
		Nodes: []uiadapter.UINode{
			{Type: "markdown", Content: "party mode first user turn"},
		},
	}
	mock := uiadapter.NewMock(fixed, nil)

	state, nodeIndex := newSuspendState(nodeID, processID, spec)
	state.exec.RepoPath = t.TempDir()

	e := NewExecutor(nil, func(name string, data interface{}) {}, WithAdapter(mock))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	t.Cleanup(drainAfter(cancel, errCh, 1))
	go func() {
		// round=2 = first user-facing turn in Party-Mode (autonomous was round 1).
		// lastOutput non-empty so §5.2's `lastOutput != ""` guard fires.
		errCh <- e.suspendForSpec(ctx, state, nodeIndex, nodeID, 2, spec,
			"Claude's first turn output that seeds the UI AST translation")
	}()

	require.True(t,
		pollForStatus(state, nodeIndex, nodeID, NodeAwaitingInput, 500*time.Millisecond),
		"node must reach NodeAwaitingInput within 500ms")

	state.mu.Lock()
	prompts := make([]PendingPrompt, len(state.exec.PendingPrompts))
	copy(prompts, state.exec.PendingPrompts)
	state.mu.Unlock()

	require.Len(t, prompts, 1, "exactly one PendingPrompt expected on round=2")
	assert.NotEmpty(t, prompts[0].Structured,
		"Structured must be populated on Party-Mode first user-facing turn (round=2) — "+
			"absence implies a `round > 1` gate was reintroduced")
	assert.Contains(t, prompts[0].Structured, "mock:party-mode-round2",
		"Structured must carry the mock AST's GeneratedBy sentinel")
}

// ── AC-5: oversize Structured blob dropped ────────────────────────────────────

// TestSuspendForSpec_OversizeBlobDropped verifies AC-5: when the adapter
// returns an AST whose JSON marshal exceeds §5.2's maxStructuredBytes (6 KiB),
// the executor-side size guard drops the blob — PendingPrompt.Structured
// stays empty. Paired with AC-4 to prove wiring + size cap hold together.
func TestSuspendForSpec_OversizeBlobDropped(t *testing.T) {
	// NOT parallel: uses testRegistry globals.
	const nodeID = "n1"
	const processID = "test-u4-oversize-dropped"
	spec := registerU4Process(t, processID)

	// 8 KiB of content — JSON marshal will exceed the §5.2 6 KiB cap comfortably.
	oversize := newOversizeAST(8 * 1024)
	mock := uiadapter.NewMock(oversize, nil)

	state, nodeIndex := newSuspendState(nodeID, processID, spec)
	state.exec.RepoPath = t.TempDir()

	e := NewExecutor(nil, func(name string, data interface{}) {}, WithAdapter(mock))

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	t.Cleanup(drainAfter(cancel, errCh, 1))
	go func() {
		errCh <- e.suspendForSpec(ctx, state, nodeIndex, nodeID, 2, spec,
			"lastOutput that produced an oversize AST")
	}()

	require.True(t,
		pollForStatus(state, nodeIndex, nodeID, NodeAwaitingInput, 500*time.Millisecond),
		"node must reach NodeAwaitingInput within 500ms")

	state.mu.Lock()
	prompts := make([]PendingPrompt, len(state.exec.PendingPrompts))
	copy(prompts, state.exec.PendingPrompts)
	state.mu.Unlock()

	require.Len(t, prompts, 1, "exactly one PendingPrompt expected")
	assert.Empty(t, prompts[0].Structured,
		"Structured must be dropped (empty) when the marshaled AST exceeds 6 KiB, got %d bytes",
		len(prompts[0].Structured))
}
