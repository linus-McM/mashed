package bmad

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const execVerifyTestdataDir = "testdata/exec_verify"

// Markers the Claude CLI renders while processing a slash command.
// The exact label varies between CLI releases, so any match proves
// the command was dispatched (not left in the input buffer).
var processingMarkers = []string{"Baking", "thinking", "Thinking", "Pondering", "Working"}

// Shells that may take over a tmux pane after claude exits under a
// `bash -c 'claude ...; exec bash'` wrapper.
var allowedSurvivalShells = []string{"bash", "zsh", "sh"}

func fixturePath(name string) string {
	return filepath.Join(execVerifyTestdataDir, name)
}

func TestVerifySlashInjectionFixture(t *testing.T) {
	data, err := os.ReadFile(fixturePath("verify1_slash_injection.txt"))
	require.NoError(t, err)

	content := string(data)
	require.Contains(t, content, "/simplify",
		"fixture must contain echo of slash command — claude did not recognise the injection")

	var foundMarker string
	for _, m := range processingMarkers {
		if strings.Contains(content, m) {
			foundMarker = m
			break
		}
	}
	require.NotEmpty(t, foundMarker,
		"fixture must show claude entered a processing state (none of %v found)", processingMarkers)
}

// buildSendKeysHexArgv mirrors the argv shape produced by
// internal/terminal/tmux_adapter.go's (*TmuxAttachment).SendInput.
// Duplicated (not imported) because SendInput is a method on a live
// attachment — there is no exported pure-function helper, and pulling
// internal/terminal into an offline fixture test would add a runtime
// dependency for no gain. The real contract is locked in by
// TestTmuxAttachment_AC3_SendInputUsesSendKeysHex.
func buildSendKeysHexArgv(target string, data []byte) []string {
	argv := make([]string, 0, 4+len(data))
	argv = append(argv, "send-keys", "-H", "-t", target)
	for _, b := range data {
		argv = append(argv, fmt.Sprintf("%02x", b))
	}
	return argv
}

func TestVerifySendInputArgv(t *testing.T) {
	const target = "verify-1:0.0"
	wantHeader := []string{"send-keys", "-H", "-t", target}

	// verify1 was recorded with CR (0x0d); LF (0x0a) is locked as an
	// alternative so a future refactor that flips terminators fails loudly.
	cases := []struct {
		name    string
		input   []byte
		wantHex []string
	}{
		{
			name:    "slash_simplify_CR",
			input:   []byte("/simplify\r"),
			wantHex: []string{"2f", "73", "69", "6d", "70", "6c", "69", "66", "79", "0d"},
		},
		{
			name:    "slash_simplify_LF",
			input:   []byte("/simplify\n"),
			wantHex: []string{"2f", "73", "69", "6d", "70", "6c", "69", "66", "79", "0a"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			argv := buildSendKeysHexArgv(target, tc.input)
			require.Equal(t, wantHeader, argv[:4])
			require.Equal(t, tc.wantHex, argv[4:])
		})
	}
}

var captureBlockRE = regexp.MustCompile(`=== CAPTURE_\d+ ===\n([\s\S]*?)=== END_CAPTURE_\d+ ===\n?`)

// parseCaptureFixture splits a multi-capture fixture on the
// `=== CAPTURE_N === / === END_CAPTURE_N ===` markers the recording
// script writes between consecutive `tmux capture-pane -p` invocations.
func parseCaptureFixture(s string) []string {
	matches := captureBlockRE.FindAllStringSubmatch(s, -1)
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}
	return out
}

func TestVerifyIdleStabilityFixture(t *testing.T) {
	data, err := os.ReadFile(fixturePath("verify2_idle_stable_pane.txt"))
	require.NoError(t, err)

	captures := parseCaptureFixture(string(data))
	require.Len(t, captures, 5, "expected 5 captures in the idle-stability fixture")

	baseline := hashCapturedOutput(captures[0])
	for i := 1; i < len(captures); i++ {
		require.Equal(t, baseline, hashCapturedOutput(captures[i]),
			"capture %d hash differs from capture 1 — pane is not stable at idle", i+1)
	}

	for i, c := range captures {
		require.True(t, detectIdlePrompt(c),
			"capture %d must be recognised as an idle prompt by detectIdlePrompt", i+1)
	}
}

func TestVerifySessionSurvivalFixture(t *testing.T) {
	data, err := os.ReadFile(fixturePath("verify3_session_survives_exit_panes.txt"))
	require.NoError(t, err)

	line := strings.TrimSpace(string(data))
	parts := strings.Fields(line)
	require.Len(t, parts, 2,
		"expected 'pane_dead pane_current_command' on one line, got %q", line)

	require.Equal(t, "0", parts[0],
		"pane_dead must be 0 (pane still alive after claude exit)")
	require.Contains(t, allowedSurvivalShells, parts[1],
		"pane_current_command must be a shell in %v, got %q", allowedSurvivalShells, parts[1])
}
