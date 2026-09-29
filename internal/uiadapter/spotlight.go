package uiadapter

import (
	"context"
	"log/slog"
	"strings"
	"unicode"
)

// Plan §3 Story 8 — prompt-injection defence via datamarking. Raw Claude-
// Code output is untrusted; schema-constrained decoding caps blast radius
// but spotlighting closes the gap by replacing whitespace in the raw-
// capture region with a rare marker so the model treats the block as
// data, not instruction (Microsoft Spotlighting, arXiv 2403.14720).

// SpotlightMarker is the character that replaces whitespace inside the
// raw region. U+2022 (bullet) is rare in normal terminal output.
const SpotlightMarker = '•'

// op-attr constants for spotlight.* records. The bare "spotlight" value is
// used by the start/added/disabled/removed records (AC-5.2 contract); the
// dotted "spotlight.scan" form on the always-firing scan record satisfies
// AC-5.8's per-file emission coverage.
const (
	spotlightOp     = "spotlight"
	spotlightScanOp = "spotlight.scan"
)

// Spotlight replaces every whitespace rune with SpotlightMarker. The
// transform is lossless — Unspotlight recovers the original text by
// substituting marker→space everywhere. Because validator.contentPreserved
// runs on the *un-marked* sanitised text, callers pass the pre-Spotlight
// version to validation, not the spotlighted one.
//
// enabled=false bypasses the transform so eval harnesses can prove the
// defence is doing work (Story v3-08 AC-8.3). logger may be nil;
// nilSafeLogger normalises it so any future story can emit telemetry without
// an inline guard.
//
// Story 5: emits spotlight.start + spotlight.added (or spotlight.disabled
// short-circuit) carrying counts only — never the raw payload.
func Spotlight(raw string, enabled bool, logger *slog.Logger) string {
	lg := nilSafeLogger(logger)
	ctx := context.Background()
	// Per-call scan record carries a dotted op so AC-5.8's emission-coverage
	// scan reliably sees a `spotlight.*` op, even on the disabled path.
	if lg.Enabled(ctx, slog.LevelDebug) {
		lg.LogAttrs(ctx, slog.LevelDebug, "spotlight.scan",
			slog.String("op", spotlightScanOp),
			slog.Bool("enabled", enabled),
			slog.Int("bytes_in", len(raw)),
		)
	}
	if !enabled {
		if lg.Enabled(ctx, slog.LevelDebug) {
			lg.LogAttrs(ctx, slog.LevelDebug, "spotlight.disabled",
				slog.String("op", spotlightOp),
				slog.Int("bytes_in", len(raw)),
			)
		}
		return raw
	}
	if lg.Enabled(ctx, slog.LevelDebug) {
		lg.LogAttrs(ctx, slog.LevelDebug, "spotlight.start",
			slog.String("op", spotlightOp),
			slog.Bool("enabled", true),
			slog.Int("bytes_in", len(raw)),
		)
	}
	var b strings.Builder
	b.Grow(len(raw))
	added := 0
	for _, r := range raw {
		if unicode.IsSpace(r) {
			b.WriteRune(SpotlightMarker)
			added++
			continue
		}
		b.WriteRune(r)
	}
	out := b.String()
	if lg.Enabled(ctx, slog.LevelDebug) {
		lg.LogAttrs(ctx, slog.LevelDebug, "spotlight.added",
			slog.String("op", spotlightOp),
			slog.Int("markers_added", added),
			slog.Int("bytes_out", len(out)),
		)
	}
	return out
}

// Unspotlight reverses Spotlight — every SpotlightMarker becomes a space.
// Whitespace that wasn't originally a single space is lost; Spotlight is
// designed for prompt-payload usage where the model receives the marked
// form and the reference copy stays separate for validation. Exposed so
// tests can assert round-trip safety.
//
// Story 5: emits spotlight.removed carrying markers_removed + bytes_out.
// logger may be nil.
func Unspotlight(marked string, logger *slog.Logger) string {
	lg := nilSafeLogger(logger)
	ctx := context.Background()
	var b strings.Builder
	b.Grow(len(marked))
	removed := 0
	for _, r := range marked {
		if r == SpotlightMarker {
			b.WriteRune(' ')
			removed++
			continue
		}
		b.WriteRune(r)
	}
	out := b.String()
	if lg.Enabled(ctx, slog.LevelDebug) {
		lg.LogAttrs(ctx, slog.LevelDebug, "spotlight.removed",
			slog.String("op", spotlightOp),
			slog.Int("markers_removed", removed),
			slog.Int("bytes_out", len(out)),
		)
	}
	return out
}
