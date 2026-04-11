package bmad

// fixture_verify_test.go — empirical regression check against a real
// tmux capture-pane dump recorded from a live Claude CLI brainstorming
// session. The fixture lives outside the repo (in /tmp) so this test
// is gated on its presence and is skipped on machines that do not have
// it handy. It exists because the unit-test hand-written fixture would
// otherwise have to mirror every subtle byte difference the Claude CLI
// emits (NBSP, trailing wide-character padding, ANSI sequences from the
// status bar, etc.), and reproducing those by hand is exactly how the
// NBSP-space bug slipped past the earlier tests.
//
// Run with:
//
//	go test ./internal/bmad/... -run TestDetectIdlePrompt_RealFixture -v
//
// To refresh the fixture after any Claude CLI UI change:
//
//	tmux capture-pane -t <session>:0.0 -p -S -200 > /tmp/pane-capture.txt

import (
	"os"
	"testing"
)

func TestDetectIdlePrompt_RealFixture(t *testing.T) {
	const fixture = "/tmp/pane-capture.txt"
	data, err := os.ReadFile(fixture)
	if err != nil {
		t.Skipf("fixture %s not present — skipping; regenerate with tmux capture-pane", fixture)
	}
	if !detectIdlePrompt(string(data)) {
		// Emit a small hex tail to the log so the next person to hit this
		// can see exactly what the CLI is rendering. Capture-pane output
		// is padded to the terminal width, so the tail hints at whether
		// a new whitespace variant has been introduced.
		n := len(data)
		start := n - 200
		if start < 0 {
			start = 0
		}
		t.Logf("fixture tail (last %d bytes): %q", n-start, data[start:])
		t.Fatal("detectIdlePrompt returned false for a real Claude CLI idle pane — did whitespace handling regress?")
	}
}
