package uiadapter

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

const fallbackSummaryMaxLen = 120

// fallbackAstOp is the canonical `op` attribute value for fallback-AST log
// emissions (Story 4 §14 sanitize discipline — closed enum).
const fallbackAstOp = "fallback.ast"

// FallbackAST builds the deterministic single-markdown-node AST every failure
// path degrades to (spec §4.8). Version is always "1"; GeneratedBy is
// "fallback:" + reason; the raw capture is placed verbatim into a single
// markdown node so downstream consumers always have the original content.
// logger may be nil; an emitted `fallback.ast.build` record carries the
// reason verbatim (closed-enum) and the raw payload's byte length only —
// never the raw bytes themselves (Story 4 §14 sanitize discipline).
func FallbackAST(raw, reason string, logger *slog.Logger) *UIAST {
	log := nilSafeLogger(logger)
	ctx := context.Background()
	debug := log.Enabled(ctx, slog.LevelDebug)
	if debug {
		log.LogAttrs(ctx, slog.LevelDebug, "fallback.ast.build",
			slog.String("op", fallbackAstOp),
			slog.String("reason", reason),
			slog.Int("bytes_in", len(raw)))
	}
	return &UIAST{
		Version:             "1",
		GeneratedBy:         "fallback:" + reason,
		GeneratedAt:         time.Now().Unix(),
		TurnSummary:         firstLine(raw, fallbackSummaryMaxLen, log),
		Nodes:               []UINode{{Type: markdownNodeType, Content: raw}},
		FallbackAnswerShape: "free",
	}
}

// firstLine returns the first newline-terminated segment of raw, capped at
// maxLen bytes. Used for Diagnostics.TurnSummary per §4.8. When the line is
// truncated, a `fallback.ast.truncate` debug record is emitted carrying
// only the byte count kept (Story 4 §14 sanitize discipline).
func firstLine(raw string, maxLen int, logger *slog.Logger) string {
	line := raw
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = line[:i]
	}
	if len(line) > maxLen {
		line = line[:maxLen]
		ctx := context.Background()
		if logger.Enabled(ctx, slog.LevelDebug) {
			logger.LogAttrs(ctx, slog.LevelDebug, "fallback.ast.truncate",
				slog.String("op", fallbackAstOp),
				slog.Bool("truncated", true),
				slog.Int("bytes_kept", len(line)))
		}
	}
	return line
}
