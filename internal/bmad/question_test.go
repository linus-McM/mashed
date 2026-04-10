// Package bmad contains tests for question detection utilities.
// Story question-01: Backend Question Detection
//
// RED Phase: These tests define the expected behavior for stripANSI,
// detectQuestion, hashQuestion, and captureQuestionOutput. They MUST FAIL
// until the go-engineer implements question.go.

package bmad

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// questionTestWorkflow returns a single-node workflow for question detection tests.
func questionTestWorkflow(id string) WorkflowDef {
	return WorkflowDef{
		ID:   id,
		Name: "Question Test " + id,
		Nodes: []WorkflowNode{
			{
				ID:        "node-A",
				ProcessID: "bmad-brainstorming",
				Label:     "Brainstorm",
				Position:  Position{X: 0, Y: 0},
				Status:    NodePending,
				Config:    map[string]string{},
			},
		},
		Edges:     []WorkflowEdge{},
		CreatedAt: "2026-04-10T00:00:00Z",
		UpdatedAt: "2026-04-10T00:00:00Z",
	}
}

// ── AC-1: ANSI escape sequences are stripped from captured output ──

func TestStory1_AC1_StripANSI(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strip color codes",
			input: "\x1b[32mHello\x1b[0m \x1b[1;31mWorld\x1b[0m",
			want:  "Hello World",
		},
		{
			name:  "strip cursor movement",
			input: "\x1b[2J\x1b[H? What should I do?",
			want:  "? What should I do?",
		},
		{
			name:  "no-op on clean string",
			input: "Just plain text",
			want:  "Just plain text",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "mixed ANSI and text",
			input: "\x1b[1m\x1b[36m?\x1b[0m What directory should I create the files in?",
			want:  "? What directory should I create the files in?",
		},
		{
			name:  "bold and underline",
			input: "\x1b[1;4mBold Underline\x1b[0m normal",
			want:  "Bold Underline normal",
		},
		{
			name:  "256 color codes",
			input: "\x1b[38;5;82mGreen text\x1b[0m",
			want:  "Green text",
		},
		{
			name:  "multiline with ANSI on each line",
			input: "\x1b[32mline1\x1b[0m\n\x1b[33mline2\x1b[0m",
			want:  "line1\nline2",
		},
		{
			name:  "only ANSI codes no text",
			input: "\x1b[2J\x1b[H\x1b[0m",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripANSI(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ── AC-2: Freeform questions are detected from Claude CLI output ──

func TestStory1_AC2_DetectFreeformQuestion(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantFound   bool
		wantQ       string
		wantOptions []string
	}{
		{
			name: "simple ? prefix question",
			output: `Working on the implementation...

? What directory should I create the files in?

> `,
			wantFound:   true,
			wantQ:       "What directory should I create the files in?",
			wantOptions: nil,
		},
		{
			name: "bordered box question",
			output: `Analyzing code structure...

` + "\u256d\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u256e" + `
` + "\u2502" + ` ? What directory should I create the    ` + "\u2502" + `
` + "\u2502" + `   files in?                             ` + "\u2502" + `
` + "\u2570\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u2500\u256f" + `
`,
			wantFound:   true,
			wantQ:       "What directory should I create the files in?",
			wantOptions: nil,
		},
		{
			name: "question with ANSI color codes",
			output: "Working...\n\n\x1b[1m\x1b[36m?\x1b[0m What file should I modify?\n\n\x1b[36m>\x1b[0m ",
			wantFound:   true,
			wantQ:       "What file should I modify?",
			wantOptions: nil,
		},
		{
			name: "question with emoji prefix",
			output: "Processing...\n\n\u2753 Should I continue with the refactor?\n\n> ",
			wantFound:   true,
			wantQ:       "Should I continue with the refactor?",
			wantOptions: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, opts, found := detectQuestion(tt.output)

			assert.Equal(t, tt.wantFound, found, "found mismatch")
			if tt.wantFound {
				assert.Equal(t, tt.wantQ, q, "question text mismatch")
				assert.Empty(t, opts, "options should be empty for freeform")
			}
		})
	}
}

// ── AC-3: Menu-style questions with numbered options are detected ──

func TestStory1_AC3_DetectMenuQuestion(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		wantFound   bool
		wantQ       string
		wantOptions []string
	}{
		{
			name: "numbered menu options",
			output: `? How would you like to proceed?
1. Create new file
2. Modify existing file
3. Skip this step
`,
			wantFound: true,
			wantQ:     "How would you like to proceed?",
			wantOptions: []string{
				"Create new file",
				"Modify existing file",
				"Skip this step",
			},
		},
		{
			name: "menu with ANSI-colored options",
			output: "\x1b[1m\x1b[36m?\x1b[0m Select a model:\n\x1b[36m1.\x1b[0m Sonnet\n\x1b[36m2.\x1b[0m Opus\n\x1b[36m3.\x1b[0m Haiku\n",
			wantFound: true,
			wantQ:     "Select a model:",
			wantOptions: []string{
				"Sonnet",
				"Opus",
				"Haiku",
			},
		},
		{
			name: "two options only",
			output: `? Do you want to overwrite?
1. Yes
2. No
`,
			wantFound: true,
			wantQ:     "Do you want to overwrite?",
			wantOptions: []string{
				"Yes",
				"No",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, opts, found := detectQuestion(tt.output)

			require.True(t, found, "should detect menu question")
			assert.Equal(t, tt.wantQ, q, "question text mismatch")
			assert.Equal(t, tt.wantOptions, opts, "options mismatch")
		})
	}
}

// ── AC-7: Output without questions produces no false positives ──

func TestStory1_AC7_NoFalsePositives(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{
			name: "normal working output",
			output: `Reading file src/main.go...
Analyzing code structure...
Found 3 functions to modify.
Applying changes...`,
		},
		{
			name:   "empty input",
			output: "",
		},
		{
			name: "question mark in text but not a prompt",
			output: `The function returns nil if the value is not found.
Why? Because the map lookup fails gracefully.
This is expected behavior.`,
		},
		{
			name: "lines with numbered list but no question",
			output: `Changes made:
1. Updated the config parser
2. Added error handling
3. Fixed the test`,
		},
		{
			name: "output with > in code context not a prompt",
			output: `if x > 0 {
    return true
}
fmt.Println("done")`,
		},
		{
			name:   "only whitespace",
			output: "   \n\n  \t  \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, found := detectQuestion(tt.output)
			assert.False(t, found, "should not detect question in normal output")
		})
	}
}

// ── AC-2 edge case: Only last 50 lines are scanned ──

func TestStory1_AC2_LastFiftyLinesOnly(t *testing.T) {
	// Build 200 lines of padding, with a question at line 10 (far outside last 50).
	var lines []string
	for i := 0; i < 200; i++ {
		if i == 9 {
			lines = append(lines, "? What directory should I create the files in?")
		} else {
			lines = append(lines, fmt.Sprintf("Normal output line %d", i+1))
		}
	}
	output := strings.Join(lines, "\n")

	_, _, found := detectQuestion(output)
	assert.False(t, found, "question at line 10 of 200 should be outside the last-50-lines scan window")
}

func TestStory1_AC2_QuestionWithinLastFiftyLines(t *testing.T) {
	// Build 200 lines, with a question at line 180 (within last 50).
	var lines []string
	for i := 0; i < 200; i++ {
		if i == 179 {
			lines = append(lines, "? What branch should I use?")
		} else {
			lines = append(lines, fmt.Sprintf("Normal output line %d", i+1))
		}
	}
	output := strings.Join(lines, "\n")

	q, _, found := detectQuestion(output)
	assert.True(t, found, "question at line 180 of 200 should be within the last-50-lines scan window")
	assert.Equal(t, "What branch should I use?", q)
}

func TestStory1_AC2_MultipleQuestions_LastOneWins(t *testing.T) {
	output := `? First question about files?

Some output in between...

? Second question about branches?

> `
	q, _, found := detectQuestion(output)
	assert.True(t, found, "should detect a question")
	assert.Equal(t, "Second question about branches?", q, "should extract the last (most recent) question")
}

// ── AC-5: Deduplication via hash ──

func TestStory1_AC5_HashQuestion(t *testing.T) {
	t.Run("same input produces same hash", func(t *testing.T) {
		h1 := hashQuestion("What directory should I create the files in?")
		h2 := hashQuestion("What directory should I create the files in?")
		assert.Equal(t, h1, h2)
	})

	t.Run("different input produces different hash", func(t *testing.T) {
		h1 := hashQuestion("What directory should I create the files in?")
		h2 := hashQuestion("Which branch should I use?")
		assert.NotEqual(t, h1, h2)
	})

	t.Run("hash is non-empty", func(t *testing.T) {
		h := hashQuestion("anything")
		assert.NotEmpty(t, h)
	})

	t.Run("empty input still produces a hash", func(t *testing.T) {
		h := hashQuestion("")
		assert.NotEmpty(t, h, "even empty string should produce a deterministic hash")
	})
}

// ── AC-4: captureQuestionOutput uses tmux capture-pane ──

func TestStory1_AC4_CaptureQuestionOutput_Success(t *testing.T) {
	h := newHarness(t)

	// Mock CommandRunner that returns sample output for capture-pane.
	sampleOutput := "Working...\n? What file should I modify?\n> "
	h.executor.runCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "capture-pane" {
			return []byte(sampleOutput), nil
		}
		return []byte("ok"), nil
	}

	got, err := h.executor.captureQuestionOutput(context.Background(), "test-session:0.0")
	require.NoError(t, err)
	assert.Equal(t, sampleOutput, got)
}

func TestStory1_AC4_CaptureQuestionOutput_Error(t *testing.T) {
	h := newHarness(t)

	h.executor.runCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "capture-pane" {
			return nil, fmt.Errorf("session not found")
		}
		return []byte("ok"), nil
	}

	_, err := h.executor.captureQuestionOutput(context.Background(), "dead-session:0.0")
	assert.Error(t, err, "should return error when tmux capture fails")
}

func TestStory1_AC4_CaptureQuestionOutput_UsesShorterHistory(t *testing.T) {
	// Verify that captureQuestionOutput uses -S -200 (not -S -5000 like captureOutput).
	h := newHarness(t)

	var capturedArgs []string
	h.executor.runCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 && args[0] == "capture-pane" {
			capturedArgs = args
			return []byte("output"), nil
		}
		return []byte("ok"), nil
	}

	_, err := h.executor.captureQuestionOutput(context.Background(), "test:0.0")
	require.NoError(t, err)

	// Should use -S -200 for the lighter question capture.
	assert.Contains(t, capturedArgs, "-S", "should pass -S flag")
	sIdx := -1
	for i, a := range capturedArgs {
		if a == "-S" {
			sIdx = i
			break
		}
	}
	require.True(t, sIdx >= 0 && sIdx+1 < len(capturedArgs), "-S flag should have a value")
	assert.Equal(t, "-200", capturedArgs[sIdx+1], "should capture last 200 lines, not 5000")
}

// ── AC-6: Stale question notifications are dismissed on node completion ──
// These are integration-level tests using testHarness.

func TestStory1_AC6_DismissalOnComplete(t *testing.T) {
	h := newHarness(t)

	// Set up a mock that returns a question on first few polls, then pane_dead=1.
	pollCount := 0
	questionOutput := "Working...\n\n? What file should I modify?\n\n> "
	h.executor.runCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 {
			switch args[0] {
			case "new-session":
				return []byte("ok"), nil
			case "capture-pane":
				return []byte(questionOutput), nil
			case "list-panes":
				pollCount++
				// After enough polls for question detection, report pane_dead.
				if pollCount > 6 {
					return []byte("1\n"), nil
				}
				return []byte("0\n"), nil
			}
		}
		return []byte("ok"), nil
	}

	wf := questionTestWorkflow("wf-q-dismiss-complete")
	require.NoError(t, h.storage.SaveWorkflow(wf))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := h.executor.StartWorkflow(ctx, wf.ID, "/tmp/test-repo", "sonnet")
	require.NoError(t, err)

	// Wait for execution to finish.
	require.Eventually(t, func() bool {
		events := h.eventsByName("bmad:node:question:dismissed")
		return len(events) > 0
	}, 5*time.Second, 50*time.Millisecond, "should emit bmad:node:question:dismissed on node completion")

	// Verify the dismissed event has the required fields.
	dismissed := h.eventsByName("bmad:node:question:dismissed")
	require.NotEmpty(t, dismissed)
	data, ok := dismissed[0].data.(map[string]string)
	if ok {
		assert.NotEmpty(t, data["execId"], "dismissed event must include execId")
		assert.NotEmpty(t, data["nodeId"], "dismissed event must include nodeId")
	}
}

func TestStory1_AC6_DismissalOnFail(t *testing.T) {
	h := newHarness(t)

	// Simulate: question is detected, then context is cancelled (node fails).
	questionOutput := "Working...\n\n? What file should I modify?\n\n> "
	h.executor.runCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 {
			switch args[0] {
			case "new-session":
				return []byte("ok"), nil
			case "capture-pane":
				return []byte(questionOutput), nil
			case "list-panes":
				// Always return alive so context cancellation causes failure.
				return []byte("0\n"), nil
			}
		}
		return []byte("ok"), nil
	}

	wf := questionTestWorkflow("wf-q-dismiss-fail")
	require.NoError(t, h.storage.SaveWorkflow(wf))

	ctx, cancel := context.WithCancel(context.Background())

	_, err := h.executor.StartWorkflow(ctx, wf.ID, "/tmp/test-repo", "sonnet")
	require.NoError(t, err)

	// Wait a bit for question to be detected, then cancel to trigger failure.
	require.Eventually(t, func() bool {
		events := h.eventsByName("bmad:node:question")
		return len(events) > 0
	}, 5*time.Second, 50*time.Millisecond, "should emit bmad:node:question before cancellation")

	cancel()

	// After cancel, the node should fail and dismiss the question.
	require.Eventually(t, func() bool {
		events := h.eventsByName("bmad:node:question:dismissed")
		return len(events) > 0
	}, 5*time.Second, 50*time.Millisecond, "should emit bmad:node:question:dismissed on node failure")
}

// ── AC-5: Duplicate questions are not re-emitted ──

func TestStory1_AC5_DeduplicationSuppressesSameQuestion(t *testing.T) {
	h := newHarness(t)

	// The same question is returned on every capture-pane call.
	questionOutput := "Working...\n\n? What file should I modify?\n\n> "
	pollCount := 0
	h.executor.runCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 {
			switch args[0] {
			case "new-session":
				return []byte("ok"), nil
			case "capture-pane":
				return []byte(questionOutput), nil
			case "list-panes":
				pollCount++
				// Let it run for many polls then die.
				if pollCount > 20 {
					return []byte("1\n"), nil
				}
				return []byte("0\n"), nil
			}
		}
		return []byte("ok"), nil
	}

	wf := questionTestWorkflow("wf-q-test")
	require.NoError(t, h.storage.SaveWorkflow(wf))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := h.executor.StartWorkflow(ctx, wf.ID, "/tmp/test-repo", "sonnet")
	require.NoError(t, err)

	// Wait for execution to complete.
	require.Eventually(t, func() bool {
		events := h.eventsByName("bmad:node:question:dismissed")
		return len(events) > 0
	}, 5*time.Second, 50*time.Millisecond, "execution should complete")

	// Despite many polls with the same question, only ONE question event should be emitted.
	questionEvents := h.eventsByName("bmad:node:question")
	assert.Equal(t, 1, len(questionEvents), "same question should only be emitted once (deduplication)")
}

func TestStory1_AC5_NewQuestionReplacesOld(t *testing.T) {
	h := newHarness(t)

	// Use separate counters for each tmux subcommand so the throttle
	// (question poll every 3rd tick) doesn't collide with list-panes calls.
	var captureCount, listPanesCount int
	h.executor.runCmd = func(ctx context.Context, name string, args ...string) ([]byte, error) {
		if name == "tmux" && len(args) > 0 {
			switch args[0] {
			case "new-session":
				return []byte("ok"), nil
			case "capture-pane":
				captureCount++
				if captureCount == 1 {
					return []byte("? First question?\n\n> "), nil
				}
				return []byte("? Second question?\n\n> "), nil
			case "list-panes":
				listPanesCount++
				if listPanesCount > 30 {
					return []byte("1\n"), nil
				}
				return []byte("0\n"), nil
			}
		}
		return []byte("ok"), nil
	}

	wf := questionTestWorkflow("wf-q-test")
	require.NoError(t, h.storage.SaveWorkflow(wf))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := h.executor.StartWorkflow(ctx, wf.ID, "/tmp/test-repo", "sonnet")
	require.NoError(t, err)

	// Wait for execution to complete.
	require.Eventually(t, func() bool {
		events := h.eventsByName("bmad:node:question:dismissed")
		return len(events) > 0
	}, 5*time.Second, 50*time.Millisecond, "execution should complete")

	// Should have emitted 2 question events (first + second, different hashes).
	questionEvents := h.eventsByName("bmad:node:question")
	assert.GreaterOrEqual(t, len(questionEvents), 2, "different questions should each emit an event")
}

// ── QuestionEvent struct validation ──

func TestStory1_AC4_QuestionEventFields(t *testing.T) {
	// Verify the QuestionEvent struct has the expected fields and JSON tags.
	evt := QuestionEvent{
		ExecID:     "exec-123",
		NodeID:     "node-A",
		RepoPath:   "/tmp/test-repo",
		RepoName:   "test-repo",
		Question:   "What file?",
		Options:    []string{"a.go", "b.go"},
		TmuxTarget: "bmad-node-A-123:0.0",
		Timestamp:  1712700000000,
		QuestionID: "abc123hash",
	}

	assert.Equal(t, "exec-123", evt.ExecID)
	assert.Equal(t, "node-A", evt.NodeID)
	assert.Equal(t, "/tmp/test-repo", evt.RepoPath)
	assert.Equal(t, "test-repo", evt.RepoName)
	assert.Equal(t, "What file?", evt.Question)
	assert.Equal(t, []string{"a.go", "b.go"}, evt.Options)
	assert.Equal(t, "bmad-node-A-123:0.0", evt.TmuxTarget)
	assert.Equal(t, int64(1712700000000), evt.Timestamp)
	assert.Equal(t, "abc123hash", evt.QuestionID)
}

// ── Story 2 AC-3: escapeTmuxLiteral preserves answer semantics ──
//
// RED Phase: escapeTmuxLiteral does not yet exist in question.go. These tests
// define the contract for the go-engineer: the helper should sanitise input
// for `tmux send-keys -l -t {target} {input}` such that the original user
// intent is preserved and no shell metacharacter is interpreted by tmux.
//
// In tmux `-l` (literal) mode, bytes are sent verbatim to the pane, so
// backslashes, dollar signs, backticks and semicolons are NOT shell-
// interpreted — they should be preserved unchanged. Since our CommandRunner
// invokes tmux directly via exec (no shell), the expected behaviour is that
// escapeTmuxLiteral is effectively identity for all printable characters.

func TestStory2_AC3_EscapeTmuxLiteral(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "plain ASCII text", input: "src/main.go", want: "src/main.go"},
		{name: "menu option number", input: "1", want: "1"},
		{name: "sentence with spaces", input: "use the default directory", want: "use the default directory"},
		{name: "single quote in contraction", input: "it's fine", want: "it's fine"},
		{name: "multiple single quotes", input: "don't it's won't", want: "don't it's won't"},
		{name: "backslashes preserved literally", input: `path\to\file`, want: `path\to\file`},
		{name: "double backslashes preserved", input: `C:\\Users\\test`, want: `C:\\Users\\test`},
		{name: "double quotes preserved", input: `he said "hi"`, want: `he said "hi"`},
		{name: "mixed single and double quotes", input: `it's a "test"`, want: `it's a "test"`},
		{name: "dollar sign preserved (no shell interp in -l mode)", input: "$HOME/project", want: "$HOME/project"},
		{name: "backtick preserved (no shell interp in -l mode)", input: "`date`", want: "`date`"},
		{name: "semicolon preserved", input: "first; second", want: "first; second"},
		{name: "pipe preserved", input: "a | b", want: "a | b"},
		{name: "ampersand preserved", input: "foo && bar", want: "foo && bar"},
		{name: "redirect characters preserved", input: "cat > out.txt", want: "cat > out.txt"},
		{name: "parens preserved", input: "$(pwd)", want: "$(pwd)"},
		{name: "unicode preserved", input: "café résumé", want: "café résumé"},
		{name: "emoji preserved", input: "thumbs up 👍", want: "thumbs up 👍"},
		{name: "tab preserved", input: "col1\tcol2", want: "col1\tcol2"},
		{name: "shell injection attempt preserved literally", input: `it's a "test" with $vars`, want: `it's a "test" with $vars`},
		{name: "leading and trailing spaces preserved", input: "  padded  ", want: "  padded  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeTmuxLiteral(tt.input)
			assert.Equal(t, tt.want, got,
				"escapeTmuxLiteral(%q) must preserve input verbatim for tmux send-keys -l", tt.input)
		})
	}
}

// TestStory2_AC3_EscapeTmuxLiteral_StripsControlBytes verifies that C0 control
// bytes are stripped so newlines cannot prematurely submit the Claude prompt
// before the explicit Enter dispatch, and ESC sequences cannot manipulate the
// input state machine. Tab is preserved as a printable-adjacent character.
func TestStory2_AC3_EscapeTmuxLiteral_StripsControlBytes(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "newline stripped (premature submission risk)", input: "line1\nline2", want: "line1line2"},
		{name: "carriage return stripped", input: "abc\rdef", want: "abcdef"},
		{name: "CRLF stripped", input: "abc\r\ndef", want: "abcdef"},
		{name: "ESC stripped (terminal injection risk)", input: "before\x1b[31mred\x1b[0m", want: "before[31mred[0m"},
		{name: "NUL byte stripped", input: "abc\x00def", want: "abcdef"},
		{name: "BEL stripped", input: "abc\x07def", want: "abcdef"},
		{name: "backspace stripped", input: "abc\x08def", want: "abcdef"},
		{name: "DEL stripped", input: "abc\x7fdef", want: "abcdef"},
		{name: "tab preserved (whitelisted)", input: "a\tb", want: "a\tb"},
		{name: "only control bytes yields empty string", input: "\n\r\x00\x1b", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeTmuxLiteral(tt.input)
			assert.Equal(t, tt.want, got,
				"escapeTmuxLiteral(%q) must strip C0 controls except tab", tt.input)
		})
	}
}
