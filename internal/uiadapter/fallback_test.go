package uiadapter

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFallbackAST_LiteralShape asserts the §4.8 fallback construction verbatim.
func TestFallbackAST_LiteralShape(t *testing.T) {
	t.Parallel()
	raw := "first line\nsecond line\nthird"
	before := time.Now().Unix()
	ast := FallbackAST(raw, "reason", nil)
	after := time.Now().Unix()

	require.NotNil(t, ast, "FallbackAST must never return nil")
	assert.Equal(t, "1", ast.Version)
	assert.Equal(t, "fallback:reason", ast.GeneratedBy)
	assert.GreaterOrEqual(t, ast.GeneratedAt, before)
	assert.LessOrEqual(t, ast.GeneratedAt, after)
	assert.Equal(t, "first line", ast.TurnSummary,
		"TurnSummary must be firstLine(raw, 120) — first newline terminates")
	assert.Equal(t, "free", ast.FallbackAnswerShape)

	require.Len(t, ast.Nodes, 1, "fallback must be exactly one markdown node")
	assert.Equal(t, "markdown", ast.Nodes[0].Type)
	assert.Equal(t, raw, ast.Nodes[0].Content,
		"the sole node must carry the raw capture verbatim")
}

// TestFallbackAST_TurnSummaryTruncates — §4.8 helper: firstLine(raw, 120)
// truncates long single-line input to 120 chars.
func TestFallbackAST_TurnSummaryTruncates(t *testing.T) {
	t.Parallel()
	raw := strings.Repeat("a", 200)
	ast := FallbackAST(raw, "r", nil)
	require.NotNil(t, ast)
	assert.LessOrEqual(t, len(ast.TurnSummary), 120,
		"TurnSummary must be bounded at 120 chars per §4.8 firstLine helper")
	assert.Equal(t, raw, ast.Nodes[0].Content,
		"raw content is NOT truncated — only TurnSummary is")
}

// TestFallbackAST_EmptyRawIsSafe — an empty capture yields an empty-but-valid
// fallback AST. No panic, no nil.
func TestFallbackAST_EmptyRawIsSafe(t *testing.T) {
	t.Parallel()
	ast := FallbackAST("", "disabled", nil)
	require.NotNil(t, ast)
	assert.Equal(t, "1", ast.Version)
	assert.Equal(t, "fallback:disabled", ast.GeneratedBy)
	require.Len(t, ast.Nodes, 1)
	assert.Equal(t, "markdown", ast.Nodes[0].Type)
	assert.Equal(t, "", ast.Nodes[0].Content)
}

// --- Story 4: uiadapter-logging-4-instrument-pipeline -----------------------
//
// AC-4.6: fallback.go emits build + truncate records.
//
// Required emissions per Story 4 dev notes:
//   - fallback.ast.build    (op=fallback.ast, reason, bytes_in)
//   - fallback.ast.truncate (op, truncated, bytes_kept)   — when firstLine
//                                                           clipped the input.

// TestStory4_AC6_FallbackASTBuild — Story 4, AC-4.6 (build site).
//
// FallbackAST emits exactly one `fallback.ast.build` record per call carrying
// the input reason verbatim (validator-reason enum) and the raw payload's byte
// length. The raw bytes themselves are NEVER logged (sanitize discipline §14).
func TestStory4_AC6_FallbackASTBuild(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	const raw = "some raw"
	const reason = "validation:oversize"
	ast := FallbackAST(raw, reason, logger)
	require.NotNil(t, ast)

	records := decodeRecords(t, buf)
	build := recordsByMsg(records, "fallback.ast.build")
	require.Len(t, build, 1, "exactly one build record; got %v", records)

	rec := build[0]
	assert.Equal(t, "fallback.ast", rec["op"])
	assert.Equal(t, reason, rec["reason"], "reason attr must equal the input reason verbatim")
	assert.EqualValues(t, len(raw), rec["bytes_in"])

	// §14: no record may carry the raw payload as an attr value.
	for _, r := range records {
		for k, v := range r {
			if s, ok := v.(string); ok && s == raw {
				t.Errorf("record attr %q leaked the raw payload verbatim", k)
			}
		}
	}
}

// TestStory4_AC6_FallbackASTTruncate — Story 4, AC-4.6 (truncate site).
//
// A multi-line raw whose first line exceeds the §4.8 fallbackSummaryMaxLen
// (120 bytes) must trigger the `firstLine` truncate emission with
// truncated=true and bytes_kept=120.
func TestStory4_AC6_FallbackASTTruncate(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	// First line is 200 chars (> 120); FallbackAST → firstLine truncates to 120.
	raw := strings.Repeat("a", 200) + "\nsecond line"
	ast := FallbackAST(raw, "oversize", logger)
	require.NotNil(t, ast)
	assert.LessOrEqual(t, len(ast.TurnSummary), 120)

	records := decodeRecords(t, buf)
	trunc := recordsByMsg(records, "fallback.ast.truncate")
	require.Len(t, trunc, 1, "exactly one truncate record; got %v", records)

	rec := trunc[0]
	assert.Equal(t, "fallback.ast", rec["op"])
	truncated, ok := rec["truncated"].(bool)
	require.True(t, ok, "truncated attr must be bool; got %T", rec["truncated"])
	assert.True(t, truncated, "truncated must be true when firstLine clipped")
	bytesKept, ok := rec["bytes_kept"].(float64)
	require.True(t, ok, "bytes_kept must be numeric; got %T", rec["bytes_kept"])
	assert.EqualValues(t, 120, bytesKept,
		"bytes_kept must equal fallbackSummaryMaxLen=120 when truncated")
}
