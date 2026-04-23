package uiadapter

import (
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

// Spotlight replaces every whitespace rune with SpotlightMarker. The
// transform is lossless — Unspotlight recovers the original text by
// substituting marker→space everywhere. Because validator.contentPreserved
// runs on the *un-marked* sanitised text, callers pass the pre-Spotlight
// version to validation, not the spotlighted one.
//
// enabled=false bypasses the transform so eval harnesses can prove the
// defence is doing work (Story v3-08 AC-8.3).
func Spotlight(raw string, enabled bool) string {
	if !enabled || raw == "" {
		return raw
	}
	var b strings.Builder
	b.Grow(len(raw))
	for _, r := range raw {
		if unicode.IsSpace(r) {
			b.WriteRune(SpotlightMarker)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Unspotlight reverses Spotlight — every SpotlightMarker becomes a space.
// Whitespace that wasn't originally a single space is lost; Spotlight is
// designed for prompt-payload usage where the model receives the marked
// form and the reference copy stays separate for validation. Exposed so
// tests can assert round-trip safety.
func Unspotlight(marked string) string {
	if marked == "" {
		return marked
	}
	var b strings.Builder
	b.Grow(len(marked))
	for _, r := range marked {
		if r == SpotlightMarker {
			b.WriteRune(' ')
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
