// Package terminal — helpers for interacting with tmux safely.
//
// This file hosts the shared tmux string-escape helper and the package-level
// sentinel errors raised by the tmux adapter. It is intentionally a verbatim
// duplicate of internal/bmad/question.go's private escapeTmuxLiteral — until
// an internal/tmuxutil shared package is extracted, keeping the logic in two
// places is simpler than a circular import between bmad and terminal.
package terminal

import (
	"errors"
	"os/exec"
	"strings"
)

// Sentinel errors returned (wrapped) by the tmux adapter.
//
// Callers should use errors.Is to test for them:
//
//	if errors.Is(err, terminal.ErrPaneDead) { ... }
var (
	// ErrPaneDead means the tmux pane targeted by Attach is no longer
	// running (pane_dead == 1). Returned from Attach and surfaced as
	// io.EOF from subsequent Read calls once the pane-death watcher fires.
	ErrPaneDead = errors.New("terminal: pane is dead")

	// ErrTmuxUnavailable means the tmux binary could not be found on PATH
	// at the time of the check. Boot-time wiring uses IsTmuxAvailable to
	// decide whether a TmuxAdapter should be constructed at all.
	ErrTmuxUnavailable = errors.New("terminal: tmux not available")
)

// EscapeTmuxLiteral sanitises a string for `tmux send-keys -l`.
//
// This is a verbatim copy of internal/bmad/question.go's escapeTmuxLiteral.
// Intentional duplication until internal/tmuxutil is extracted as a shared
// package.
//
// The tmux adapter invokes tmux directly via exec.Command (no shell), and
// `-l` (literal) mode sends bytes verbatim to the pane, so shell meta-
// characters such as `$`, `` ` ``, `;`, `|`, and backslashes are NOT
// interpreted and are preserved unchanged.
//
// However, C0 control bytes (including `\n`, `\r`, `\x1b`/ESC, and NUL) must
// be stripped: `\n`/`\r` would submit the pane prompt prematurely, and ESC
// sequences could manipulate the CLI's input state machine. Tab (`\t`) is
// preserved. DEL (0x7f) is also stripped.
func EscapeTmuxLiteral(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r == '\t' || (r >= 0x20 && r != 0x7f) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// IsTmuxAvailable reports whether the tmux binary is on PATH.
//
// Boot-time wiring uses this to decide whether to create a TmuxAdapter at
// all — when tmux is not present, the adapter is never instantiated and
// callers fall back to the PTY-based SessionManager.
func IsTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}
