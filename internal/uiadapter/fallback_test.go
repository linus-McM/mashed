package uiadapter

import (
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
