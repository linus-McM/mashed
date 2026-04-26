package uiadapter

import (
	"context"
	"log/slog"
	"regexp"
	"strings"
)

// sanitizeOp is the canonical op-attr value for every sanitize.* record.
// Centralised so emissions in this file cannot drift apart and §14
// reviewers can grep for a single string.
const sanitizeOp = "sanitize"

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
//
// Story 5: sanitize.start / sanitize.done emit metadata only — never any
// substring of raw. The §14 sanitize discipline reviewer must verify that
// no attr value carries payload bytes.
func SanitizeCapture(raw string, logger *slog.Logger) (sanitized string, deltaBytes int) {
	lg := nilSafeLogger(logger)
	ctx := context.Background()
	if lg.Enabled(ctx, slog.LevelDebug) {
		lg.LogAttrs(ctx, slog.LevelDebug, "sanitize.start",
			slog.String("op", sanitizeOp),
			slog.Int("bytes_in", len(raw)),
		)
	}
	var out string
	if raw != "" {
		out = strings.TrimSpace(stripOutsideFences(raw, lg))
	}
	delta := len(raw) - len(out)
	if lg.Enabled(ctx, slog.LevelDebug) {
		lg.LogAttrs(ctx, slog.LevelDebug, "sanitize.done",
			slog.String("op", sanitizeOp),
			slog.Int("bytes_in", len(raw)),
			slog.Int("bytes_out", len(out)),
			slog.Int("delta_bytes", delta),
		)
	}
	return out, delta
}

// stripOutsideFences walks raw splitting on triple-backtick fences, running
// sanitisation only on the prose regions. Preserving fences byte-for-byte is
// AC-1.3 — any nested regex pass here would eat legitimate backticks and
// brackets inside a code block.
//
// When logger is Debug-enabled and the cumulative chrome strip removes any
// bytes, a sanitize.fence_strip record is emitted carrying the byte count
// only — never the stripped content.
func stripOutsideFences(raw string, logger *slog.Logger) string {
	var b strings.Builder
	b.Grow(len(raw))

	i := 0
	inFence := false
	removed := 0
	for i < len(raw) {
		// Look for the next fence delimiter.
		idx := strings.Index(raw[i:], "```")
		if idx < 0 {
			// Remainder: sanitise if outside fence, verbatim if inside.
			segment := raw[i:]
			if inFence {
				b.WriteString(segment)
			} else {
				cleaned := sanitizeChrome(segment, logger)
				removed += len(segment) - len(cleaned)
				b.WriteString(cleaned)
			}
			break
		}
		// Segment up to the delimiter.
		segment := raw[i : i+idx]
		if inFence {
			b.WriteString(segment)
		} else {
			cleaned := sanitizeChrome(segment, logger)
			removed += len(segment) - len(cleaned)
			b.WriteString(cleaned)
		}
		// Write the fence delimiter verbatim and flip state.
		b.WriteString("```")
		inFence = !inFence
		i += idx + 3
	}
	if removed > 0 && logger.Enabled(context.Background(), slog.LevelDebug) {
		logger.LogAttrs(context.Background(), slog.LevelDebug, "sanitize.fence_strip",
			slog.String("op", sanitizeOp),
			slog.Int("removed_bytes", removed),
		)
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

// sanitizeChromeOp is the dotted op-attr value for sanitize.chrome_match
// records. The dotted form satisfies AC-5.8's per-file emission coverage —
// every prose-segment pass emits a chrome_match record, even when the
// match count is zero, so operators always have a sanitize.* breadcrumb.
const sanitizeChromeOp = "sanitize.chrome_match"

// sanitizeChrome removes ANSI CSI, OSC, and lone-ESC remnants in order.
// Run only on prose regions outside fences (stripOutsideFences). Emits a
// sanitize.chrome_match record with the count of patterns that matched —
// metadata only, never the matched content (§14 sanitize discipline).
// The emission fires unconditionally so the per-file emission set is
// reliably covered.
func sanitizeChrome(s string, logger *slog.Logger) string {
	matched := 0
	if locs := reOSC.FindAllStringIndex(s, -1); len(locs) > 0 {
		matched += len(locs)
		s = reOSC.ReplaceAllString(s, "")
	}
	if locs := reCSI.FindAllStringIndex(s, -1); len(locs) > 0 {
		matched += len(locs)
		s = reCSI.ReplaceAllString(s, "")
	}
	if locs := reLone.FindAllStringIndex(s, -1); len(locs) > 0 {
		matched += len(locs)
		s = reLone.ReplaceAllString(s, "")
	}
	if logger.Enabled(context.Background(), slog.LevelDebug) {
		logger.LogAttrs(context.Background(), slog.LevelDebug, "sanitize.chrome_match",
			slog.String("op", sanitizeChromeOp),
			slog.Int("patterns_matched_count", matched),
		)
	}
	return s
}
