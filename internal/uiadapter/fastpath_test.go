package uiadapter

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFastPath_CoversCommonCases — AC-2.1. Every trivially-structured
// shape in the v3.0 rule set is caught by a rule.
func TestFastPath_CoversCommonCases(t *testing.T) {
	t.Parallel()
	fp := NewFastPathClassifier(true, nil)

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
	fp := NewFastPathClassifier(false, nil)
	hit, rule, ast := fp.Classify("Proceed (y/n)?")
	assert.False(t, hit)
	assert.Empty(t, rule)
	assert.Nil(t, ast)
}

// TestFastPath_HitRateAndCounters — AC-2.2. Per-rule counters + hit rate.
func TestFastPath_HitRateAndCounters(t *testing.T) {
	t.Parallel()
	fp := NewFastPathClassifier(true, nil)
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
	fp := NewFastPathClassifier(true, nil)
	hit, rule, ast := fp.Classify("1) only option")
	assert.False(t, hit, "single numbered item must NOT hit")
	assert.Empty(t, rule)
	assert.Nil(t, ast)
}

// TestFastPath_NumberedMenuExtractsOptions — the generated UIAST has one
// WidgetOption per numbered line with numeric values preserved.
func TestFastPath_NumberedMenuExtractsOptions(t *testing.T) {
	t.Parallel()
	fp := NewFastPathClassifier(true, nil)
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
	fp := NewFastPathClassifier(true, nil)
	hit, rule, _ := fp.Classify("Is this correct (y/n)?")
	require.True(t, hit)
	assert.Equal(t, "yn-prompt", rule, "yn-prompt must outrank free-text-prompt")
}

// BenchmarkFastPath — AC-2.3 harness. Sub-1ms p95 is expected; the
// benchmark simply exercises 10k iterations so `go test -bench=. -benchtime=10000x`
// can verify the latency target empirically when needed.
func BenchmarkFastPath(b *testing.B) {
	fp := NewFastPathClassifier(true, nil)
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
	fp := NewFastPathClassifier(true, nil)
	narrative := strings.Repeat("This is long narrative prose without any structural cue. ", 10)
	hit, _, _ := fp.Classify(narrative)
	assert.False(t, hit)
}

// --- Story 4: uiadapter-logging-4-instrument-pipeline -----------------------
//
// AC-4.1: fastpath.go emits classify outcomes:
//   - fastpath.classify.start  (op=fastpath.classify, bytes_in, enabled)
//   - fastpath.classify.hit    (op, rule=<rule.Name>, latency_ms)
//   - fastpath.classify.skip   (op, latency_ms)
//   - fastpath.classify.disabled (op)  — sole emission when enabled=false.
//
// These RED-phase tests assert the JSON record contract. They will fail until
// fastpath.go is instrumented in Task 1.

// TestStory4_AC1_FastpathHit — Story 4, AC-4.1 (hit branch).
//
// A YN-style raw matches `yn-prompt`; the buffer must contain exactly one
// `fastpath.classify.start` and one `fastpath.classify.hit` and no skip /
// disabled record. The hit record must carry the matched rule name as `rule`.
func TestStory4_AC1_FastpathHit(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	fp := NewFastPathClassifier(true, logger)

	const raw = "Continue? [y/N]"
	hit, rule, ast := fp.Classify(raw)
	require.True(t, hit, "yn-prompt rule must match %q", raw)
	require.Equal(t, "yn-prompt", rule)
	require.NotNil(t, ast)

	records := decodeRecords(t, buf)
	starts := recordsByMsg(records, "fastpath.classify.start")
	hits := recordsByMsg(records, "fastpath.classify.hit")
	skips := recordsByMsg(records, "fastpath.classify.skip")
	disabled := recordsByMsg(records, "fastpath.classify.disabled")

	require.Len(t, starts, 1, "exactly one start record; got %v", records)
	require.Len(t, hits, 1, "exactly one hit record; got %v", records)
	assert.Empty(t, skips, "skip must not emit on a successful match")
	assert.Empty(t, disabled, "disabled must not emit when enabled=true")

	start := starts[0]
	assert.Equal(t, "fastpath.classify", start["op"])
	assert.Equal(t, true, start["enabled"], "enabled attr must mirror classifier flag")
	bytesIn, ok := start["bytes_in"].(float64)
	require.True(t, ok, "bytes_in must be numeric; got %T %v", start["bytes_in"], start["bytes_in"])
	assert.EqualValues(t, len(raw), bytesIn, "bytes_in must equal len(raw)")

	hitRec := hits[0]
	assert.Equal(t, "fastpath.classify", hitRec["op"])
	assert.Equal(t, "yn-prompt", hitRec["rule"], "hit record must carry the matched rule name")
	latencyMs, ok := hitRec["latency_ms"].(float64)
	require.True(t, ok, "hit record must carry numeric latency_ms; got %T", hitRec["latency_ms"])
	assert.GreaterOrEqual(t, latencyMs, 0.0)
}

// TestStory4_AC1_FastpathSkip — Story 4, AC-4.1 (skip branch).
//
// Prose with no rule match emits `fastpath.classify.start` then
// `fastpath.classify.skip` — never `hit` or `disabled`.
func TestStory4_AC1_FastpathSkip(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	fp := NewFastPathClassifier(true, logger)

	hit, _, _ := fp.Classify("just some prose")
	require.False(t, hit, "prose must not match any rule")

	records := decodeRecords(t, buf)
	require.Len(t, recordsByMsg(records, "fastpath.classify.start"), 1)
	require.Len(t, recordsByMsg(records, "fastpath.classify.skip"), 1)
	assert.Empty(t, recordsByMsg(records, "fastpath.classify.hit"))
	assert.Empty(t, recordsByMsg(records, "fastpath.classify.disabled"))

	skip := recordsByMsg(records, "fastpath.classify.skip")[0]
	assert.Equal(t, "fastpath.classify", skip["op"])
	latencyMs, ok := skip["latency_ms"].(float64)
	require.True(t, ok, "skip record must carry numeric latency_ms")
	assert.GreaterOrEqual(t, latencyMs, 0.0)
}

// TestStory4_AC1_FastpathDisabled — Story 4, AC-4.1 (disabled branch).
//
// enabled=false short-circuits with exactly one `fastpath.classify.disabled`
// record. No start / hit / skip records emit because the function exits before
// any rule check.
func TestStory4_AC1_FastpathDisabled(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	fp := NewFastPathClassifier(false, logger)

	hit, _, _ := fp.Classify("anything")
	require.False(t, hit, "disabled classifier never hits")

	records := decodeRecords(t, buf)
	disabled := recordsByMsg(records, "fastpath.classify.disabled")
	require.Len(t, disabled, 1, "exactly one disabled record; got %v", records)
	assert.Empty(t, recordsByMsg(records, "fastpath.classify.start"))
	assert.Empty(t, recordsByMsg(records, "fastpath.classify.hit"))
	assert.Empty(t, recordsByMsg(records, "fastpath.classify.skip"))

	assert.Equal(t, "fastpath.classify", disabled[0]["op"])
}
