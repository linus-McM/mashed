package uiadapter

import (
	"log/slog"
	"strings"
	"time"
)

const fallbackSummaryMaxLen = 120

// FallbackAST builds the deterministic single-markdown-node AST every failure
// path degrades to (spec §4.8). Version is always "1"; GeneratedBy is
// "fallback:" + reason; the raw capture is placed verbatim into a single
// markdown node so downstream consumers always have the original content.
// logger may be nil; nilSafeLogger normalises it so any future story can
// emit telemetry without an inline guard.
func FallbackAST(raw, reason string, logger *slog.Logger) *UIAST {
	_ = nilSafeLogger(logger)
	return &UIAST{
		Version:             "1",
		GeneratedBy:         "fallback:" + reason,
		GeneratedAt:         time.Now().Unix(),
		TurnSummary:         firstLine(raw, fallbackSummaryMaxLen),
		Nodes:               []UINode{{Type: markdownNodeType, Content: raw}},
		FallbackAnswerShape: "free",
	}
}

// firstLine returns the first newline-terminated segment of raw, capped at
// maxLen runes. Used for Diagnostics.TurnSummary per §4.8.
func firstLine(raw string, maxLen int) string {
	line := raw
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	if len(line) > maxLen {
		line = line[:maxLen]
	}
	return line
}
