package bmad

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// ansiEscape matches ANSI CSI and OSC escape sequences in a single pass.
// CSI: \x1b[[0-9;]*[a-zA-Z] (colors, cursor movement)
// OSC: \x1b][^\x07]*\x07 (operating system commands)
var ansiEscape = regexp.MustCompile(`\x1b(?:\[[0-9;]*[a-zA-Z]|\][^\x07]*\x07)`)

// questionPrefix matches lines starting with "? " (simple Claude CLI question format).
var questionPrefix = regexp.MustCompile(`^\?\s+(.+)`)

// borderedQuestionPrefix matches "? " inside a bordered box line (e.g., "│ ? What...").
var borderedQuestionPrefix = regexp.MustCompile(`^[│\|]\s*\?\s+(.+?)[\s│\|]*$`)

// borderedContinuation matches continuation lines inside a bordered box.
var borderedContinuation = regexp.MustCompile(`^[│\|]\s{3,}(.+?)[\s│\|]*$`)

// numberedOption matches lines like "1. Option text", "2. Another option".
var numberedOption = regexp.MustCompile(`^\d+\.\s+(.+)`)

// emojiQuestion matches lines starting with the red question mark emoji.
var emojiQuestion = regexp.MustCompile(`^❓\s+(.+)`)

// claudePromptRune is the Claude CLI's empty-input prompt chevron.
// It is NOT matched via a regex because Go's RE2 `\s` class only covers
// ASCII whitespace ([\t\n\f\r ]) and the Claude CLI actually renders a
// NO-BREAK SPACE (U+00A0) after the chevron, not a regular space. We
// use strings.TrimSpace (which trims via unicode.IsSpace, so it handles
// NBSP, ideographic space, and the rest of Unicode Zs correctly) and
// compare the trimmed line to this rune directly. See detectIdlePrompt
// for the call site and the regression test TestDetectIdlePrompt/NBSP.
const claudePromptRune = "❯"

// Event names for question detection/dismissal.
const (
	EventQuestion          = "bmad:node:question"
	EventQuestionDismissed = "bmad:node:question:dismissed"
	// EventIdle signals that a BMAD node's tmux pane is sitting at a
	// Claude CLI input prompt with no recent output activity — i.e. claude
	// is waiting for the user to type. Distinct from EventQuestion because
	// there is no structured question text to surface; the frontend shows
	// a "Waiting for input" style snackbar instead.
	EventIdle          = "bmad:node:idle"
	EventIdleDismissed = "bmad:node:idle:dismissed"
)

// QuestionEvent is emitted when Claude CLI asks the user a question.
type QuestionEvent struct {
	ExecID     string   `json:"execId"`
	NodeID     string   `json:"nodeId"`
	RepoPath   string   `json:"repoPath"`
	RepoName   string   `json:"repoName"`
	Question   string   `json:"question"`
	Options    []string `json:"options"`
	TmuxTarget string   `json:"tmuxTarget"`
	Timestamp  int64    `json:"timestamp"`
	QuestionID string   `json:"questionId"`
}

// IdleEvent is emitted when a BMAD node's tmux pane has reached the
// Claude CLI input prompt and its output has been stable for at least one
// poll cycle — signalling that claude has finished its current turn and
// is waiting on user input. The frontend surfaces this as a "Waiting for
// input" snackbar distinct from the structured question snackbar.
//
// Unlike QuestionEvent there is no question text or options — the frontend
// just prompts the user to open the terminal and type. A follow-up
// EventIdleDismissed (with map[string]string{execId, nodeId}) fires as
// soon as the pane emits new output or the node completes/fails.
type IdleEvent struct {
	ExecID     string `json:"execId"`
	NodeID     string `json:"nodeId"`
	RepoPath   string `json:"repoPath"`
	RepoName   string `json:"repoName"`
	TmuxTarget string `json:"tmuxTarget"`
	Timestamp  int64  `json:"timestamp"`
}

// stripANSI removes ANSI escape sequences (CSI and OSC) from a string in a single pass.
func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// detectQuestion scans the last 50 lines of output for a Claude CLI question prompt.
// It returns the question text, any numbered menu options, and whether a question was found.
// When multiple questions exist, the last one wins.
func detectQuestion(output string) (question string, options []string, found bool) {
	cleaned := stripANSI(output)

	lines := strings.Split(cleaned, "\n")

	// Take only the last 50 lines.
	if len(lines) > 50 {
		lines = lines[len(lines)-50:]
	}

	var lastQuestion string
	lastQuestionIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for simple "? " prefix.
		if m := questionPrefix.FindStringSubmatch(trimmed); m != nil {
			lastQuestion = strings.TrimSpace(m[1])
			lastQuestionIdx = i
			continue
		}

		// Check for emoji question mark prefix.
		if m := emojiQuestion.FindStringSubmatch(trimmed); m != nil {
			lastQuestion = strings.TrimSpace(m[1])
			lastQuestionIdx = i
			continue
		}

		// Check for bordered box question.
		if m := borderedQuestionPrefix.FindStringSubmatch(trimmed); m != nil {
			lastQuestion = strings.TrimSpace(m[1])
			lastQuestionIdx = i
			// Look ahead for continuation lines within the box.
			for j := i + 1; j < len(lines); j++ {
				contTrimmed := strings.TrimSpace(lines[j])
				if cm := borderedContinuation.FindStringSubmatch(contTrimmed); cm != nil {
					lastQuestion += " " + strings.TrimSpace(cm[1])
				} else {
					break
				}
			}
			continue
		}
	}

	if lastQuestion == "" {
		return "", nil, false
	}

	// Scan for numbered options after the last question line.
	// A nil slice is returned for freeform questions (no numbered options found).
	var opts []string
	for i := lastQuestionIdx + 1; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if m := numberedOption.FindStringSubmatch(trimmed); m != nil {
			opts = append(opts, strings.TrimSpace(m[1]))
		}
	}

	return lastQuestion, opts, true
}

// hashQuestion returns the SHA-256 hex digest of the question text.
func hashQuestion(q string) string {
	h := sha256.Sum256([]byte(q))
	return hex.EncodeToString(h[:])
}

// idlePromptScanTail bounds how many lines from the END of the captured
// output detectIdlePrompt considers. The Claude CLI renders its input
// prompt close to the bottom of the viewport, so anything earlier is
// guaranteed stale ANSI history that can be ignored.
const idlePromptScanTail = 30

// detectIdlePrompt returns true when the captured tmux output ends with a
// Claude CLI input prompt — specifically, when any of the last
// idlePromptScanTail lines (after ANSI stripping) is just a `❯` chevron
// with optional surrounding whitespace.
//
// This signal intentionally does NOT mean "a structured question is being
// asked" — detectQuestion handles that case. It only answers the weaker
// question "does the pane look like it is waiting for the user to type
// something". Pair it with an output-stability check in the caller
// (hash compared across two consecutive polls) so that transient prompt
// renders during a claude turn are not mistaken for idle.
//
// Whitespace handling: the Claude CLI renders the cursor row as the
// chevron followed by a NO-BREAK SPACE (U+00A0), NOT an ASCII space,
// and tmux capture-pane preserves that exact byte sequence. Go RE2's
// `\s` character class is ASCII-only, so any regex-based match of the
// form `^\s*❯\s*$` would miss every real pane. We sidestep the entire
// regex-vs-Unicode pitfall by using strings.TrimSpace — whose behaviour
// is defined in terms of unicode.IsSpace and therefore strips NBSP,
// ideographic space, and every other Zs category rune — and then
// comparing the trimmed line to the chevron rune directly.
func detectIdlePrompt(output string) bool {
	if output == "" {
		return false
	}
	cleaned := stripANSI(output)
	lines := strings.Split(cleaned, "\n")
	if len(lines) > idlePromptScanTail {
		lines = lines[len(lines)-idlePromptScanTail:]
	}
	for _, line := range lines {
		if strings.TrimSpace(line) == claudePromptRune {
			return true
		}
	}
	return false
}

// hashCapturedOutput returns a short hash of a tmux capture-pane payload,
// used by pollForIdle to detect "pane output is unchanged since the
// previous poll". Reuses SHA-256 (same as hashQuestion) for simplicity —
// collisions would cause a missed idle dismissal, not a security issue.
func hashCapturedOutput(output string) string {
	return hashQuestion(output)
}

// escapeTmuxLiteral sanitises an answer string for `tmux send-keys -l`.
//
// The BMAD executor invokes tmux directly via exec.Command (no shell), and
// `-l` (literal) mode sends bytes verbatim to the pane, so shell meta-
// characters such as `$`, `` ` ``, `;`, `|`, and backslashes are NOT
// interpreted and are preserved unchanged.
//
// However, C0 control bytes (including `\n`, `\r`, `\x1b`/ESC, and NUL)
// must be stripped: `\n`/`\r` would submit the Claude prompt prematurely
// before our explicit Enter dispatch, and ESC sequences could manipulate
// the Claude CLI's input state machine. Tab (`\t`) is preserved since
// it's a printable-adjacent character commonly used in answers.
func escapeTmuxLiteral(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		// Strip C0 controls (0x00-0x1f) except tab, plus DEL (0x7f).
		if r == '\t' || (r >= 0x20 && r != 0x7f) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// captureQuestionOutput captures the last 200 lines from a tmux pane for question scanning.
// This is a lighter capture than captureOutput (which uses -S -5000).
// Output is capped at maxCaptureBytes to prevent unbounded memory usage.
func (e *Executor) captureQuestionOutput(ctx context.Context, target string) (string, error) {
	out, err := e.runCmd(ctx, "tmux", "capture-pane", "-t", target, "-p", "-S", "-200")
	if err != nil {
		return "", err
	}
	s := string(out)
	if len(s) > maxCaptureBytes {
		s = s[len(s)-maxCaptureBytes:]
	}
	return s, nil
}
