package uiadapter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFastPath_CoversCommonCases — AC-2.1. Every trivially-structured
// shape in the v3.0 rule set is caught by a rule.
func TestFastPath_CoversCommonCases(t *testing.T) {
	t.Parallel()
	fp := NewFastPathClassifier(true)

	cases := []struct {
		name     string
		raw      string
		wantRule string
	}{
		{"yn-bare", "Proceed (y/n)?", "yn-prompt"},
		{"yn-bracketed", "Apply? [Y/n]", "yn-prompt"},
		{"press-enter", "Press Enter to continue", "press-enter"},
		{"file-confirm", "  apply patch file.go?", "file-confirm"},
		{"numbered-menu-3", "Choose:\n1) first\n2) second\n3) third", "numbered-menu"},
		{"free-text", "What is your name?", "free-text-prompt"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hit, rule, ast := fp.Classify(tc.raw)
			require.True(t, hit, "expected fast-path hit for %q", tc.raw)
			assert.Equal(t, tc.wantRule, rule)
			require.NotNil(t, ast)
			assert.Equal(t, "fastpath:"+tc.wantRule, ast.GeneratedBy)
			assert.Equal(t, "1", ast.Version)
			require.NotEmpty(t, ast.Nodes)
		})
	}
}

// TestFastPath_DisabledReturnsMiss — Config.EnableFastPath=false short-
// circuits the classifier to miss unconditionally (Story B gating).
func TestFastPath_DisabledReturnsMiss(t *testing.T) {
	t.Parallel()
	fp := NewFastPathClassifier(false)
	hit, rule, ast := fp.Classify("Proceed (y/n)?")
	assert.False(t, hit)
	assert.Empty(t, rule)
	assert.Nil(t, ast)
}

// TestFastPath_HitRateAndCounters — AC-2.2. Per-rule counters + hit rate.
func TestFastPath_HitRateAndCounters(t *testing.T) {
	t.Parallel()
	fp := NewFastPathClassifier(true)
	_, _, _ = fp.Classify("Proceed (y/n)?")
	_, _, _ = fp.Classify("What is your name?")
	_, _, _ = fp.Classify("some narrative with no rule match at all")
	_, _, _ = fp.Classify("Proceed (y/n)?")

	assert.InDelta(t, 0.75, fp.HitRate(), 0.001, "3 hits out of 4 calls")

	hits := fp.HitsPerRule()
	assert.EqualValues(t, 2, hits["yn-prompt"])
	assert.EqualValues(t, 1, hits["free-text-prompt"])
	assert.EqualValues(t, 0, hits["numbered-menu"])
}

// TestFastPath_NumberedMenuRequiresTwoItems — single-item numbered prompts
// fall through to the LLM pipeline (conservative default; §8 risk mitigation).
func TestFastPath_NumberedMenuRequiresTwoItems(t *testing.T) {
	t.Parallel()
	fp := NewFastPathClassifier(true)
	hit, rule, ast := fp.Classify("1) only option")
	assert.False(t, hit, "single numbered item must NOT hit")
	assert.Empty(t, rule)
	assert.Nil(t, ast)
}

// TestFastPath_NumberedMenuExtractsOptions — the generated UIAST has one
// WidgetOption per numbered line with numeric values preserved.
func TestFastPath_NumberedMenuExtractsOptions(t *testing.T) {
	t.Parallel()
	fp := NewFastPathClassifier(true)
	raw := "Pick one:\n1. apple\n2. banana\n3. cherry"
	hit, _, ast := fp.Classify(raw)
	require.True(t, hit)
	require.NotNil(t, ast)
	require.Len(t, ast.Nodes, 1)
	require.NotNil(t, ast.Nodes[0].Widget)
	options := ast.Nodes[0].Widget.Options
	require.Len(t, options, 3)
	assert.Equal(t, "1", options[0].Value)
	assert.Equal(t, "apple", options[0].Label)
	assert.Equal(t, "cherry", options[2].Label)
}

// TestFastPath_FirstMatchWins — yn-prompt precedes free-text-prompt in the
// rule table so a raw that matches both rules returns yn-prompt.
func TestFastPath_FirstMatchWins(t *testing.T) {
	t.Parallel()
	fp := NewFastPathClassifier(true)
	hit, rule, _ := fp.Classify("Is this correct (y/n)?")
	require.True(t, hit)
	assert.Equal(t, "yn-prompt", rule, "yn-prompt must outrank free-text-prompt")
}

// BenchmarkFastPath — AC-2.3 harness. Sub-1ms p95 is expected; the
// benchmark simply exercises 10k iterations so `go test -bench=. -benchtime=10000x`
// can verify the latency target empirically when needed.
func BenchmarkFastPath(b *testing.B) {
	fp := NewFastPathClassifier(true)
	inputs := []string{
		"Proceed (y/n)?",
		"1) first\n2) second",
		"Press Enter to continue",
		"What is your favourite colour?",
		"narrative that does not match any rule",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = fp.Classify(inputs[i%len(inputs)])
	}
}

// TestFastPath_NoRuleMatches — narrative / multi-paragraph turns fall
// through to the LLM pipeline.
func TestFastPath_NoRuleMatches(t *testing.T) {
	t.Parallel()
	fp := NewFastPathClassifier(true)
	narrative := strings.Repeat("This is long narrative prose without any structural cue. ", 10)
	hit, _, _ := fp.Classify(narrative)
	assert.False(t, hit)
}
