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

// Event names for question detection/dismissal.
const (
	EventQuestion          = "bmad:node:question"
	EventQuestionDismissed = "bmad:node:question:dismissed"
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
