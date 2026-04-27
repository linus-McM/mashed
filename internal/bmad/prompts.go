package bmad

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// modalQuestionCap bounds the PendingPrompt.LastOutput payload. The event
// bus ships this string on every awaiting_input event + persists it to
// execution.json — 4 KiB balances readability (enough for a full Claude
// turn) against payload bloat. Callers truncate on the tail, not the head:
// the tail is the just-spoken question, the head is earlier context the
// user already saw.
const modalQuestionCap = 4 * 1024

// maxStructuredBytes caps the serialized UIAST blob attached to
// PendingPrompt.Structured (ui-ast-U4 §5.2, §3.3). Defence-in-depth against
// adapter misbehaviour: the uiadapter validator rejects >6 KiB AST, so this
// guard only fires on bugs. Blobs above this cap are dropped to "" so the
// frontend falls back to Layer 1.
const maxStructuredBytes = 6 * 1024

// extractModalQuestion pulls the best-available "question to show the user"
// from a raw tmux capture. Hybrid strategy:
//  1. If Claude authored a <MASHED_PROMPT>…</MASHED_PROMPT> sentinel (future
//     skill work), use the sentinel body — it's the formatted question.
//  2. Otherwise, fall back to the last ~4 KiB of the capture — noisy but
//     always present.
//
// Empty input returns "" so the modal falls back to renderPrompt()'s static
// spec.Prompt text.
func extractModalQuestion(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	const openTag = "<MASHED_PROMPT>"
	const closeTag = "</MASHED_PROMPT>"
	if o := strings.LastIndex(raw, openTag); o >= 0 {
		if c := strings.Index(raw[o+len(openTag):], closeTag); c >= 0 {
			body := strings.TrimSpace(raw[o+len(openTag) : o+len(openTag)+c])
			if body != "" {
				return body
			}
		}
	}
	if len(raw) > modalQuestionCap {
		raw = raw[len(raw)-modalQuestionCap:]
		// Trim a partial first line so the user sees clean output.
		if nl := strings.IndexByte(raw, '\n'); nl >= 0 && nl < 256 {
			raw = raw[nl+1:]
		}
	}
	return raw
}

// waiter returns (creating on first access) the release channel for a
// (nodeID, inputID) suspension. Subsequent callers receive the same channel
// so responders and suspenders share a rendezvous point.
func (s *execState) waiter(nodeID, inputID string) chan struct{} {
	key := nodeID + "/" + inputID
	s.waitersMu.Lock()
	defer s.waitersMu.Unlock()
	if s.waiters == nil {
		s.waiters = map[string]chan struct{}{}
	}
	if ch, ok := s.waiters[key]; ok {
		return ch
	}
	ch := make(chan struct{})
	s.waiters[key] = ch
	return ch
}

// releaseWaiter closes and removes the channel for (nodeID, inputID). Safe to
// call when no waiter exists (no-op).
func (s *execState) releaseWaiter(nodeID, inputID string) {
	key := nodeID + "/" + inputID
	s.waitersMu.Lock()
	defer s.waitersMu.Unlock()
	if ch, ok := s.waiters[key]; ok {
		close(ch)
		delete(s.waiters, key)
	}
}

// hashPendingPrompt produces the deterministic 16-char hex PromptID for a
// prompt identified by (nodeID, inputID, round).
func hashPendingPrompt(nodeID, inputID string, round int) string {
	return sha256hex(fmt.Sprintf("%s|%s|%d", nodeID, inputID, round), valueHashLen)
}

// upsertPrompt adds or replaces the PendingPrompt matching (NodeID, InputID)
// in place. Order-preserving: existing entries retain their position.
func upsertPrompt(prompts []PendingPrompt, p PendingPrompt) []PendingPrompt {
	for i := range prompts {
		if prompts[i].NodeID == p.NodeID && prompts[i].InputID == p.InputID {
			prompts[i] = p
			return prompts
		}
	}
	return append(prompts, p)
}

// removePrompt returns a new slice with the (nodeID, inputID) entry dropped.
// Unchanged-length result means the entry was not found.
func removePrompt(prompts []PendingPrompt, nodeID, inputID string) []PendingPrompt {
	out := make([]PendingPrompt, 0, len(prompts))
	for _, p := range prompts {
		if p.NodeID == nodeID && p.InputID == inputID {
			continue
		}
		out = append(out, p)
	}
	return out
}

// findPendingPrompt returns the entry matching (nodeID, inputID) and true,
// or the zero value and false.
func findPendingPrompt(prompts []PendingPrompt, nodeID, inputID string) (PendingPrompt, bool) {
	for _, p := range prompts {
		if p.NodeID == nodeID && p.InputID == inputID {
			return p, true
		}
	}
	return PendingPrompt{}, false
}

// findInputSpec looks up an InputSpec by ID within a ProcessDef.
func findInputSpec(proc ProcessDef, inputID string) (InputSpec, bool) {
	for _, s := range proc.InputSpecs {
		if s.ID == inputID {
			return s, true
		}
	}
	return InputSpec{}, false
}

// renderPrompt returns the user-visible prompt text for a spec. Current
// implementation returns spec.Prompt verbatim — template substitution will
// land when registry/context wiring is fleshed out (S7).
func renderPrompt(spec InputSpec, _ *execState, _ string) string {
	return spec.Prompt
}

// resolveOptions returns spec.Options when set, otherwise attempts a registry
// lookup for spec.OptionsRef. The registryLookup result is a newline-joined
// list of candidate values, so split into individual options for the UI.
// Unresolvable refs produce nil (the UI then renders a free-text input).
func resolveOptions(spec InputSpec, _ *execState) []string {
	if len(spec.Options) > 0 {
		return spec.Options
	}
	if spec.OptionsRef == "" {
		return nil
	}
	v, err := registryLookup(spec.OptionsRef)
	if err != nil || v == "" {
		return nil
	}
	lines := strings.Split(v, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// persistSnapshot writes state.exec to ~/.mashed/workflows/{execID}/execution.json
// atomically (tempfile + rename). Called on every PendingPrompts / NodeInputs
// transition so a crash leaves the latest suspension visible to restore paths.
//
// Concurrency: snapshotMu serialises the marshal→write→rename sequence per
// execution so concurrent writers cannot race on a shared tempfile name.
// Version=2 is stamped when any interactive field is non-zero (§16.5).
func (e *Executor) persistSnapshot(state *execState) error {
	state.snapshotMu.Lock()
	defer state.snapshotMu.Unlock()

	state.mu.Lock()
	if len(state.exec.PendingPrompts) > 0 ||
		len(state.exec.NodeInputs) > 0 ||
		len(state.exec.NodeInputHistory) > 0 ||
		len(state.exec.NodeRounds) > 0 {
		state.exec.Version = snapshotVersionInteractive
	}
	data, err := json.MarshalIndent(state.exec, "", "  ")
	execID := state.exec.ID
	state.mu.Unlock()
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}
	dir := filepath.Join(home, ".mashed", "workflows", execID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir snapshot: %w", err)
	}
	// Unique tempfile per call avoids POSIX rename collisions when the
	// snapshotMu guard is bypassed (e.g. by another process in the same
	// HOME). snapshotMu covers the intra-process case.
	tmp := filepath.Join(dir, fmt.Sprintf("execution.json.%d.%s.tmp", os.Getpid(), randHex8()))
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write snapshot: %w", err)
	}
	if err := os.Rename(tmp, filepath.Join(dir, "execution.json")); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename snapshot: %w", err)
	}
	return nil
}

// snapshotVersionInteractive is the schema version tag for snapshots that
// carry at least one interactive field (§16.5).
const snapshotVersionInteractive = 2

// randHex8 returns 8 random hex chars for tempfile uniqueness.
func randHex8() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// translateForPrompt runs the UI AST adapter (ui-ast-U4 §5.2) and returns
// the serialized AST plus an optional fallback-shape override.
//
// structured is "" when:
//   - the executor has no adapter installed (production with UIAdapterEnabled=false);
//   - lastOutput is empty (round-1 pre-claude suspension — no turn to translate);
//   - the adapter reported a cancellation via Diagnostics.CancelReason;
//   - marshal failed or the blob exceeds maxStructuredBytes (defence-in-depth).
//
// fallbackShape is non-empty only when the adapter degraded to its built-in
// fallback path (GeneratedBy="fallback:*"). Callers may use it to swap the
// suspended node's literal Shape (typically ShapeJSON for adapter-driven
// processes) for a sensible widget — without it, users see a raw JSON
// textarea every time the adapter times out, errors, or the model is missing.
//
// Must be called OUTSIDE state.mu and state.snapshotMu — Translate may block
// for the adapter timeout (~3 s) and holding either mutex would serialise all
// node progress. The brief lock here snapshots the node's ProcessID only.
func (e *Executor) translateForPrompt(ctx context.Context, state *execState, idx int, lastOutput string) (structured string, fallbackShape InputShape) {
	if e.adapter == nil || lastOutput == "" {
		return "", ""
	}
	state.mu.Lock()
	procID := state.exec.Nodes[idx].ProcessID
	state.mu.Unlock()

	ast := e.adapter.Translate(ctx, lastOutput, procID)
	if ast == nil || ast.Diagnostics.CancelReason != "" {
		return "", ""
	}
	blob, err := json.Marshal(ast)
	if err != nil || len(blob) > maxStructuredBytes {
		return "", ""
	}
	// Honor the AST's FallbackAnswerShape whenever it is supplied — not just
	// on `fallback:*` paths. A successful adapter run that emits only
	// markdown/info nodes (no decision_group) leaves the modal with no
	// rich widget; without this override the user lands back in a raw JSON
	// textarea even though the adapter explicitly hinted at a sensible
	// alternative shape. When decision_group nodes are present the modal
	// hides the Layer-1 widget anyway, so the override is a no-op.
	if ast.FallbackAnswerShape != "" {
		switch InputShape(ast.FallbackAnswerShape) {
		case ShapeFree, ShapeChoice, ShapeMultiChoice, ShapeApproval, ShapeFile, ShapeJSON:
			fallbackShape = InputShape(ast.FallbackAnswerShape)
		default:
			fallbackShape = ShapeFree
		}
	} else if strings.HasPrefix(ast.GeneratedBy, "fallback:") {
		// Defensive: a fallback path with no shape hint defaults to free.
		fallbackShape = ShapeFree
	} else {
		// Adapter succeeded but emitted no shape hint. Inspect the AST: if
		// it contains a decision_group node, the modal will render rich
		// widgets and we must NOT override the literal shape (the modal
		// hides Layer-1 anyway). Otherwise the user would land in a raw
		// JSON textarea for a process whose adapter produced no choices —
		// default to free-text so they can still respond.
		hasDecisionGroup := false
		for _, n := range ast.Nodes {
			if n.Type == "decision_group" {
				hasDecisionGroup = true
				break
			}
		}
		if !hasDecisionGroup {
			fallbackShape = ShapeFree
		}
	}
	return string(blob), fallbackShape
}

// suspendForSpec transitions a node into NodeAwaitingInput, emits the
// awaiting_input event, and blocks until either the matching waiter channel
// is closed (a response arrived) or ctx is canceled. On wake it restores
// NodeRunning, removes the PendingPrompt, persists a second snapshot, and
// emits input_resolved with the sha256-hashed value. Ctx cancel emits
// EventAborted with reason "workflow stopped" and returns ctx.Err().
//
// Use suspendForSpecWithPane when a tmux pane target is available so the
// suspension is also dismissed if the pane resumes activity (e.g. user typed
// directly into tmux). Pre-spawn suspends (declared inputs resolved before
// tmux session creation) call this wrapper with no pane watcher.
func (e *Executor) suspendForSpec(
	ctx context.Context,
	state *execState,
	nodeIndex map[string]int,
	nodeID string,
	round int,
	spec InputSpec,
	lastOutput string,
) error {
	return e.suspendForSpecWithPane(ctx, state, nodeIndex, nodeID, round, spec, lastOutput, "")
}

// suspendForSpecWithPane is suspendForSpec plus a pane-activity watchdog. When
// paneTarget is non-empty, a watcher goroutine polls `tmux capture-pane` at
// e.pollInterval. If the captured hash changes while the node is suspended in
// NodeAwaitingInput, the suspension is aborted: the node is demoted back to
// NodeRunning, the PendingPrompt is cleared, EventAwaitingDismissed fires, and
// ErrAwaitingPaneActive is returned so the caller can re-run the idle-wait +
// suspend cycle against the FRESH pane content.
func (e *Executor) suspendForSpecWithPane(
	ctx context.Context,
	state *execState,
	nodeIndex map[string]int,
	nodeID string,
	round int,
	spec InputSpec,
	lastOutput string,
	paneTarget string,
) error {
	idx, ok := nodeIndex[nodeID]
	if !ok {
		return fmt.Errorf("suspendForSpec: unknown node %q: %w", nodeID, ErrExecNotFound)
	}

	// ui-ast-U4 §5.2: translate the upstream turn into a UIAST BEFORE
	// state.mu/state.snapshotMu are acquired. Translate may block up to ~3 s;
	// holding either mutex here would stall every other node.
	structured, fallbackShape := e.translateForPrompt(ctx, state, idx, lastOutput)

	// §5.2: re-check ctx — Translate may have taken hundreds of ms and the
	// run could have been canceled in the meantime. Bail before mutating
	// PendingPrompts so a canceled run doesn't leave a phantom prompt.
	if err := ctx.Err(); err != nil {
		return err
	}

	// Adapter-driven processes (registry uses ShapeJSON to mean "let the
	// uiadapter pick a widget") degrade to a raw JSON textarea when the
	// adapter falls back. Swap to the AST's FallbackAnswerShape so the user
	// sees a sensible widget (typically free-text) instead of being asked to
	// type JSON into a multiple-choice prompt. Specs that explicitly want
	// raw JSON should not run through the adapter, so this override is safe.
	resolvedShape := spec.Shape
	if fallbackShape != "" && spec.Shape == ShapeJSON {
		resolvedShape = fallbackShape
	}

	prompt := PendingPrompt{
		NodeID:     nodeID,
		InputID:    spec.ID,
		Prompt:     renderPrompt(spec, state, nodeID),
		Shape:      resolvedShape,
		Options:    resolveOptions(spec, state),
		Round:      round,
		CreatedAt:  time.Now().Unix(),
		PromptID:   hashPendingPrompt(nodeID, spec.ID, round),
		LastOutput: extractModalQuestion(lastOutput),
		Structured: structured,
	}

	// Create the waiter BEFORE emitting so a fast responder that fires
	// in reaction to EventAwaitingInput finds a channel to close.
	waitCh := state.waiter(nodeID, spec.ID)

	state.mu.Lock()
	execID := state.exec.ID
	prompt.ExecID = execID
	state.exec.Nodes[idx].Status = NodeAwaitingInput
	state.exec.PendingPrompts = upsertPrompt(state.exec.PendingPrompts, prompt)
	state.mu.Unlock()

	if err := e.persistSnapshot(state); err != nil {
		return err
	}
	e.emit(EventAwaitingInput, awaitingPayload(prompt))

	// Pane-activity watchdog. Empty paneTarget disables the watcher so
	// pre-spawn suspends (no tmux session yet) keep their current behavior.
	paneActiveCh := make(chan struct{})
	watcherCtx, cancelWatcher := context.WithCancel(ctx)
	defer cancelWatcher()
	if paneTarget != "" {
		go e.watchPaneForActivity(watcherCtx, paneTarget, paneActiveCh)
	}

	select {
	case <-waitCh:
		// Response arrived; fall through to the wake-up block.
	case <-paneActiveCh:
		// Pane resumed activity (claude began producing output). Demote the
		// node back to NodeRunning, clear the stale prompt, and signal the
		// caller to redo the idle-wait + suspend cycle.
		return e.dismissAwaiting(state, idx, nodeID, spec.ID, round, "pane_active")
	case <-ctx.Done():
		e.emit(EventAborted, abortedPayload(execID, nodeID, "workflow stopped"))
		return ctx.Err()
	}

	// Read the resolved value and clear the prompt under a single lock so
	// the snapshot + emit see a consistent view. Emit EventInputResolved
	// BEFORE the NodeRunning transition: tests observe status via a poll
	// loop and then assert on captured events, so a status-last ordering
	// would race the emit against the poll's next tick.
	state.mu.Lock()
	state.exec.PendingPrompts = removePrompt(state.exec.PendingPrompts, nodeID, spec.ID)
	value := ""
	if state.exec.NodeInputs != nil {
		value = state.exec.NodeInputs[nodeID][spec.ID]
	}
	state.mu.Unlock()

	if err := e.persistSnapshot(state); err != nil {
		return err
	}
	e.emit(EventInputResolved, inputResolvedPayload(execID, nodeID, spec.ID, round, value))

	state.mu.Lock()
	state.exec.Nodes[idx].Status = NodeRunning
	state.mu.Unlock()
	return nil
}

// watchPaneForActivity polls the tmux pane and closes activeCh on the first
// hash change observed after the initial baseline capture. The watcher exits
// silently when ctx is canceled (e.g. the response arrived first or the
// suspend was aborted by ctx.Done) so the goroutine never leaks. Capture
// errors are non-fatal — they just skip the tick, mirroring waitForIdleCompletion.
func (e *Executor) watchPaneForActivity(ctx context.Context, paneTarget string, activeCh chan<- struct{}) {
	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	// Establish the baseline immediately. If the very first capture errors,
	// fall back to the empty hash; the watcher will pick up the real baseline
	// on the next successful capture.
	var baseline string
	if first, err := e.captureQuestionOutput(ctx, paneTarget); err == nil {
		baseline = hashCapturedOutput(first)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			captured, err := e.captureQuestionOutput(ctx, paneTarget)
			if err != nil {
				continue
			}
			cur := hashCapturedOutput(captured)
			if baseline == "" {
				baseline = cur
				continue
			}
			if cur != baseline {
				close(activeCh)
				return
			}
		}
	}
}

// dismissAwaiting demotes a node from NodeAwaitingInput back to NodeRunning,
// clears the matching PendingPrompt, releases the waiter so a stale
// RespondToInput cannot resolve the discarded prompt, persists the snapshot,
// emits EventAwaitingDismissed, and returns ErrAwaitingPaneActive so the
// round loop can re-enter waitForIdleCompletion + re-suspend with the latest
// pane capture.
func (e *Executor) dismissAwaiting(state *execState, idx int, nodeID, inputID string, round int, reason string) error {
	state.mu.Lock()
	execID := state.exec.ID
	state.exec.PendingPrompts = removePrompt(state.exec.PendingPrompts, nodeID, inputID)
	state.exec.Nodes[idx].Status = NodeRunning
	state.mu.Unlock()

	state.releaseWaiter(nodeID, inputID)

	if err := e.persistSnapshot(state); err != nil {
		return err
	}
	e.emit(EventAwaitingDismissed, awaitingDismissedPayload(execID, nodeID, inputID, round, reason))
	return ErrAwaitingPaneActive
}

// RespondToInput records a user-supplied answer for a suspended node input
// (§8.1). Validation failures emit EventInputInvalid and return a wrapped
// sentinel — the node stays in NodeAwaitingInput so the UI can retry.
// Success stores the value, appends to history, persists, and releases the
// waiter so suspendForSpec wakes up.
func (e *Executor) RespondToInput(execID, nodeID, inputID, value string) error {
	state, err := e.getState(execID)
	if err != nil {
		return fmt.Errorf("exec %q: %w", execID, err)
	}

	// Snapshot repo path + pending prompt + node process id in one critical
	// section so concurrent responders see a consistent view.
	state.mu.Lock()
	prompt, hasPrompt := findPendingPrompt(state.exec.PendingPrompts, nodeID, inputID)
	repoPath := state.exec.RepoPath
	var procID string
	var nodeInputSpecs []InputSpec
	for _, n := range state.exec.Nodes {
		if n.ID == nodeID {
			procID = n.ProcessID
			nodeInputSpecs = n.InputSpecs
			break
		}
	}
	state.mu.Unlock()

	if !hasPrompt {
		return fmt.Errorf("no pending prompt for %s/%s: %w", nodeID, inputID, ErrNoPendingPrompt)
	}

	// Prefer node-local InputSpecs (test overrides) before falling back to
	// the registry — same policy as resolveInputs.
	proc := ProcessDef{InputSpecs: nodeInputSpecs}
	if len(proc.InputSpecs) == 0 {
		if p, ok := ProcessByID(procID); ok {
			proc = p
		}
	}
	spec, specOK := findInputSpec(proc, inputID)
	if !specOK {
		return fmt.Errorf("input %q not declared on node %q: %w", inputID, nodeID, ErrUnknownInput)
	}

	if vErr := validateInput(spec, value); vErr != nil {
		e.emit(EventInputInvalid, invalidPayload(execID, nodeID, inputID, vErr.Error()))
		return vErr
	}

	stored := value
	if spec.Shape == ShapeFile {
		resolved, fErr := resolveFileInput(value, repoPath)
		if fErr != nil {
			e.emit(EventInputInvalid, invalidPayload(execID, nodeID, inputID, "path outside repository root"))
			return fErr
		}
		stored = resolved
	}

	state.mu.Lock()
	if state.exec.NodeInputs == nil {
		state.exec.NodeInputs = map[string]map[string]string{}
	}
	if state.exec.NodeInputs[nodeID] == nil {
		state.exec.NodeInputs[nodeID] = map[string]string{}
	}
	state.exec.NodeInputs[nodeID][inputID] = stored
	if state.exec.NodeInputHistory == nil {
		state.exec.NodeInputHistory = map[string][]NodeInputEntry{}
	}
	state.exec.NodeInputHistory[nodeID] = append(state.exec.NodeInputHistory[nodeID], NodeInputEntry{
		InputID:   inputID,
		Round:     prompt.Round,
		Value:     stored,
		Timestamp: time.Now().Unix(),
	})
	state.mu.Unlock()

	if err := e.persistSnapshot(state); err != nil {
		return err
	}

	state.releaseWaiter(nodeID, inputID)
	return nil
}
