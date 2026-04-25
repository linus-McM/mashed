package uiadapter

import (
	"log/slog"
	"regexp"
	"strings"
)

// SanitizeCapture strips ANSI/CLI chrome from a raw Claude-Code turn so
// validator.contentPreserved's byte-match against URLs/code doesn't false-
// positive on escape sequences. Fenced triple-backtick blocks are preserved
// byte-for-byte — sanitisation applies only outside them (Plan §3 Story 1,
// AC-1.1 / AC-1.3).
//
// deltaBytes is len(raw) - len(sanitized) and feeds the `sanitize_delta_bytes`
// slog attribute (AC-1.4, §6.1 telemetry contract). logger may be nil;
// nilSafeLogger normalises it so any future story can emit telemetry without
// an inline guard.
func SanitizeCapture(raw string, logger *slog.Logger) (sanitized string, deltaBytes int) {
	_ = nilSafeLogger(logger)
	if raw == "" {
		return "", 0
	}
	out := stripOutsideFences(raw)
	out = strings.TrimSpace(out)
	return out, len(raw) - len(out)
}

// stripOutsideFences walks raw splitting on triple-backtick fences, running
// sanitisation only on the prose regions. Preserving fences byte-for-byte is
// AC-1.3 — any nested regex pass here would eat legitimate backticks and
// brackets inside a code block.
func stripOutsideFences(raw string) string {
	var b strings.Builder
	b.Grow(len(raw))

	i := 0
	inFence := false
	for i < len(raw) {
		// Look for the next fence delimiter.
		idx := strings.Index(raw[i:], "```")
		if idx < 0 {
			// Remainder: sanitise if outside fence, verbatim if inside.
			segment := raw[i:]
			if inFence {
				b.WriteString(segment)
			} else {
				b.WriteString(sanitizeChrome(segment))
			}
			break
		}
		// Segment up to the delimiter.
		segment := raw[i : i+idx]
		if inFence {
			b.WriteString(segment)
		} else {
			b.WriteString(sanitizeChrome(segment))
		}
		// Write the fence delimiter verbatim and flip state.
		b.WriteString("```")
		inFence = !inFence
		i += idx + 3
	}
	return b.String()
}

var (
	// ANSI CSI: ESC '[' params final-byte. Final byte is any ASCII letter
	// (case-insensitive). Covers colour codes, cursor moves, clear lines.
	reCSI = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)
	// ANSI OSC: ESC ']' payload BEL. Minimal non-greedy. OSC-8 hyperlinks
	// use this shape.
	reOSC = regexp.MustCompile(`\x1b\].*?\x07`)
	// Lone escape-sequence remnants (e.g. truncated captures).
	reLone = regexp.MustCompile(`\x1b[^\[\]]`)
)

// sanitizeChrome removes ANSI CSI, OSC, and lone-ESC remnants in order.
// Run only on prose regions outside fences (stripOutsideFences).
func sanitizeChrome(s string) string {
	s = reOSC.ReplaceAllString(s, "")
	s = reCSI.ReplaceAllString(s, "")
	s = reLone.ReplaceAllString(s, "")
	return s
}
