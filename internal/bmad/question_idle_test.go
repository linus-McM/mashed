package bmad

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetectIdlePrompt(t *testing.T) {
	// Real Party Mode capture excerpt — note the `❯` on its own line between
	// two dashed separators, followed by the Claude CLI status bar.
	partyModeIdle := "" +
		"  What would you like to discuss with the team today, Linus? 🌊\n" +
		"\n" +
		"  (Say *exit, goodbye, end party, or quit anytime to wrap up.)\n" +
		"\n" +
		"✻ Sautéed for 50s\n" +
		"\n" +
		"────────────────────────────────────────────────────────────────\n" +
		"❯ \n" +
		"────────────────────────────────────────────────────────────────\n" +
		"   cwd: /Users/dev/Development/surfseer  ⎇ main\n" +
		"   Model: Opus 4.6                         Total: 68.5 t/s\n" +
		"  ⏵⏵ bypass permissions on (shift+tab to cycle)\n"

	// Regression: tmux capture-pane of a real Claude CLI run showed the
	// prompt rendered as "❯\u00A0" — the chevron followed by a NO-BREAK
	// SPACE (U+00A0), NOT an ASCII space. The original regex-based
	// detector used Go's `\s` class which is ASCII-only and silently
	// failed to match, which is why the idle snackbar never fired in
	// practice. Keep this exact byte sequence here so any future
	// refactor that reintroduces an ASCII-only whitespace assumption
	// fails loudly at test time.
	nbspIdle := "claude is thinking...\n" +
		"\n" +
		"────────────\n" +
		"❯\u00a0\n" + // U+00A0 NO-BREAK SPACE
		"────────────\n" +
		"   cwd: /tmp\n"

	// Same layout but the user has started typing — the prompt line now has
	// trailing content so detectIdlePrompt must NOT match.
	partyModeTyping := "" +
		"────────────────────────────────────────────────────────────────\n" +
		"❯ winston what do you think about\n" +
		"────────────────────────────────────────────────────────────────\n"

	// A structured question (detectQuestion's territory) with a bare prompt
	// below — the idle detector should still match on the prompt because
	// its job is pane-level idleness, not question-vs-idle discrimination.
	// The CALLER (executor) decides which signal wins when both fire.
	structuredPlusPrompt := "" +
		"? Would you like to proceed?\n" +
		"1. Yes\n" +
		"2. No\n" +
		"\n" +
		"❯ \n"

	// ANSI-wrapped idle prompt — the detector must strip escapes before
	// applying the regex so the prompt is still recognised.
	ansiIdle := "line1\n\x1b[2m\x1b[0m\x1b[38;5;242m❯ \x1b[0m\nline3\n"

	// Prompt exists but is buried far above the idleScanTail window —
	// detector should only look at the last N lines. Anything older is
	// ignored because the pane has scrolled past it.
	prefix := strings.Repeat("filler line\n", idlePromptScanTail+5)
	promptBuried := "❯\n" + prefix

	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "empty", in: "", want: false},
		{name: "no prompt anywhere", in: "hello world\nfoo bar\n", want: false},
		{name: "bare prompt on last line", in: "hello\n❯\n", want: true},
		{name: "prompt with trailing space", in: "hello\n❯ \n", want: true},
		{name: "prompt with leading whitespace", in: "hello\n  ❯  \n", want: true},
		{name: "prompt followed by NBSP (real Claude CLI)", in: nbspIdle, want: true},
		{name: "prompt surrounded by mixed ASCII + NBSP", in: "x\n \u00a0 ❯\u00a0\u00a0 \ny\n", want: true},
		{name: "real party-mode idle capture", in: partyModeIdle, want: true},
		{name: "user typing after prompt — NOT idle", in: partyModeTyping, want: false},
		{name: "structured question + prompt → idle detector still fires", in: structuredPlusPrompt, want: true},
		{name: "ANSI-wrapped prompt", in: ansiIdle, want: true},
		{name: "prompt older than scan tail", in: promptBuried, want: false},
		{name: "mid-line ❯ is not an idle prompt", in: "  ❯ surfseer wants a ❯ chevron mid-line\n", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, detectIdlePrompt(tt.in),
				"detectIdlePrompt(%q)", tt.in)
		})
	}
}

func TestHashCapturedOutput_StableAndDifferentiates(t *testing.T) {
	a := hashCapturedOutput("hello\n❯\n")
	aCopy := hashCapturedOutput("hello\n❯\n")
	b := hashCapturedOutput("hello\n❯ typed\n")

	assert.Equal(t, a, aCopy,
		"hashCapturedOutput is deterministic for identical input")
	assert.NotEqual(t, a, b,
		"hashCapturedOutput differentiates on any byte change")
	assert.NotEmpty(t, a,
		"hashCapturedOutput never returns empty string")
}

// newIdleTestState builds a minimal execState suitable for driving
// pollForIdle directly, bypassing the full workflow runner. The
// repoPath/execID are fixed because the tests only assert on event
// emission count and payload shape, not on workflow integration.
func newIdleTestState() *execState {
	return &execState{
		exec: &WorkflowExecution{
			ID:       "exec-idle-test",
			RepoPath: "/tmp/idle-test",
		},
		lastQuestionHash: map[string]string{},
		lastOutputHash:   map[string]string{},
		idleEmitted:      map[string]bool{},
	}
}

func TestPollForIdle_EmitsAfterStableIdleFrame(t *testing.T) {
	var idleEvents int
	var dismissedEvents int
	e := &Executor{emitEvent: func(name string, _ interface{}) {
		switch name {
		case EventIdle:
			idleEvents++
		case EventIdleDismissed:
			dismissedEvents++
		}
	}}
	state := newIdleTestState()

	// First poll: idle prompt present but no prior capture → only seeds
	// lastOutputHash, no event. Stability requires TWO consecutive matching
	// captures.
	e.pollForIdle(state, "node-A", "t1:0.0", "hello\n❯\n")
	assert.Equal(t, 0, idleEvents,
		"first poll must NOT emit — stability needs two matching captures")

	// Second poll with the SAME output → stable AND idle prompt present
	// → emit EventIdle exactly once.
	e.pollForIdle(state, "node-A", "t1:0.0", "hello\n❯\n")
	assert.Equal(t, 1, idleEvents,
		"second matching poll must emit EventIdle")
	assert.Equal(t, 0, dismissedEvents)

	// Third poll, still the same idle frame → dedupe, no new event.
	e.pollForIdle(state, "node-A", "t1:0.0", "hello\n❯\n")
	assert.Equal(t, 1, idleEvents,
		"subsequent identical polls must dedupe")
}

func TestPollForIdle_DismissesOnOutputChange(t *testing.T) {
	var idleEvents int
	var dismissedEvents int
	e := &Executor{emitEvent: func(name string, _ interface{}) {
		switch name {
		case EventIdle:
			idleEvents++
		case EventIdleDismissed:
			dismissedEvents++
		}
	}}
	state := newIdleTestState()

	// Drive the state machine to "emitted".
	e.pollForIdle(state, "node-A", "t1:0.0", "hello\n❯\n")
	e.pollForIdle(state, "node-A", "t1:0.0", "hello\n❯\n")
	require.Equal(t, 1, idleEvents)

	// Output changes — user started typing, or claude emitted more output.
	// pollForIdle must emit EventIdleDismissed exactly once and clear the
	// idleEmitted flag so a future idle window can emit again.
	e.pollForIdle(state, "node-A", "t1:0.0", "hello\n❯ winston\n")
	assert.Equal(t, 1, dismissedEvents,
		"hash change after emit must fire EventIdleDismissed")

	// Subsequent change with no prior emit → silent.
	e.pollForIdle(state, "node-A", "t1:0.0", "hello\n❯ winston, what\n")
	assert.Equal(t, 1, dismissedEvents,
		"further hash changes without a prior emit must stay silent")

	// Stability re-establishes on the new frame — eventually able to emit
	// a new idle event.
	e.pollForIdle(state, "node-A", "t1:0.0", "hello\n❯\n")
	e.pollForIdle(state, "node-A", "t1:0.0", "hello\n❯\n")
	assert.Equal(t, 2, idleEvents,
		"after dismissal, a new idle window can emit again")
}

func TestPollForIdle_DoesNotEmitWhenNotIdle(t *testing.T) {
	var idleEvents int
	e := &Executor{emitEvent: func(name string, _ interface{}) {
		if name == EventIdle {
			idleEvents++
		}
	}}
	state := newIdleTestState()

	// Stable but not idle (no `❯` line) → must never emit.
	for i := 0; i < 5; i++ {
		e.pollForIdle(state, "node-A", "t1:0.0", "claude is thinking...\n")
	}
	assert.Equal(t, 0, idleEvents,
		"stable non-idle output must never emit EventIdle")
}

func TestPollForIdle_EmitsIdleEventPayload(t *testing.T) {
	var captured IdleEvent
	e := &Executor{emitEvent: func(name string, payload interface{}) {
		if name == EventIdle {
			captured = payload.(IdleEvent)
		}
	}}
	state := newIdleTestState()

	e.pollForIdle(state, "node-A", "bmad-foo-main-party-mode-abcdef01:0.0", "hello\n❯\n")
	e.pollForIdle(state, "node-A", "bmad-foo-main-party-mode-abcdef01:0.0", "hello\n❯\n")

	assert.Equal(t, "exec-idle-test", captured.ExecID)
	assert.Equal(t, "node-A", captured.NodeID)
	assert.Equal(t, "/tmp/idle-test", captured.RepoPath)
	assert.Equal(t, "idle-test", captured.RepoName,
		"RepoName must be filepath.Base(RepoPath)")
	assert.Equal(t, "bmad-foo-main-party-mode-abcdef01:0.0", captured.TmuxTarget)
	assert.NotZero(t, captured.Timestamp,
		"Timestamp must be set to the emission time")
}

func TestHasRecentQuestion(t *testing.T) {
	tests := []struct {
		name     string
		captured string
		want     bool
	}{
		{
			name:     "structured_question_prefix",
			captured: "some output\n? Do you want to continue?\n❯\n",
			want:     true,
		},
		{
			name:     "emoji_question",
			captured: "output line\n❓ Which option do you prefer?\n❯\n",
			want:     true,
		},
		{
			name:     "bordered_question",
			captured: "output\n│ ? What should I build? │\n❯\n",
			want:     true,
		},
		{
			name:     "natural_language_question",
			captured: "Agent Roster\n───────\nWhat do you want to discuss?\n───────\n❯\n",
			want:     true,
		},
		{
			name:     "natural_language_question_with_context",
			captured: "Welcome! The team is ready.\nWhat would you like to build today?\n❯\n",
			want:     true,
		},
		{
			name:     "no_question_just_idle",
			captured: "Task completed successfully.\nAll files written.\n❯\n",
			want:     false,
		},
		{
			name:     "short_fragment_not_question",
			captured: "Done.\nOK?\n❯\n",
			want:     false, // "OK?" is only 3 chars, below minQuestionLength
		},
		{
			name:     "empty_output",
			captured: "",
			want:     false,
		},
		{
			name:     "question_too_far_from_end",
			captured: "What do you want?\n" + strings.Repeat("line\n", 20) + "❯\n",
			want:     false, // question is beyond questionGateScanTail
		},
		{
			name:     "question_within_scan_window",
			captured: strings.Repeat("line\n", 5) + "What do you want to discuss?\n❯\n",
			want:     true,
		},
		{
			name:     "ansi_stripped_before_check",
			captured: "\x1b[32mWhat do you want to build?\x1b[0m\n❯\n",
			want:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasRecentQuestion(tt.captured)
			assert.Equal(t, tt.want, got)
		})
	}
}
