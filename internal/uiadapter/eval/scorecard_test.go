package eval

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/uiadapter"
)

// sequentialMockAdapter returns pre-staged UIASTs one per Translate call in
// slice order. The uiadapter.MockAdapter in mock.go is unsuitable here: it is
// behind //go:build testing (these tests do not set that tag) and returns a
// fixed AST, but the metrics test needs a distinct response per fixture.
type sequentialMockAdapter struct {
	responses []*uiadapter.UIAST
	idx       int
}

func (m *sequentialMockAdapter) Translate(_ context.Context, _, _ string) *uiadapter.UIAST {
	if m.idx >= len(m.responses) {
		panic("sequentialMockAdapter: Translate called more times than responses were staged")
	}
	r := m.responses[m.idx]
	m.idx++
	return r
}

func newSynthFixture(id string, widgets []string) Fixture {
	return Fixture{
		ID:  id,
		Raw: "synth-" + id,
		Expected: Expected{
			Category:    "freeform",
			WidgetTypes: widgets,
		},
	}
}

func astOllama(latencyMs int) *uiadapter.UIAST {
	return &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "ollama:gemma3:4b",
		Nodes:       []uiadapter.UINode{{Type: "markdown", Content: "ok"}},
		Diagnostics: uiadapter.Diagnostics{LatencyMs: latencyMs},
	}
}

func astValidatorFailed(latencyMs int) *uiadapter.UIAST {
	return &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "fallback:validation:required_dropped",
		Nodes:       []uiadapter.UINode{{Type: "markdown", Content: "raw"}},
		Diagnostics: uiadapter.Diagnostics{LatencyMs: latencyMs},
	}
}

func astMalformed(latencyMs int) *uiadapter.UIAST {
	return &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "fallback:validation:malformed",
		Nodes:       []uiadapter.UINode{{Type: "markdown", Content: "raw"}},
		Diagnostics: uiadapter.Diagnostics{LatencyMs: latencyMs},
	}
}

// TestEval_Scorecard_Metrics — AC-2: ValidJSONRate, ValidatorPassRate, P95.
func TestEval_Scorecard_Metrics(t *testing.T) {
	t.Parallel()

	// Uniform 1s latencies so P95 equals 1s under any percentile formula.
	responses := []*uiadapter.UIAST{
		astOllama(1000), astOllama(1000), astOllama(1000), astOllama(1000),
		astOllama(1000), astOllama(1000), astOllama(1000), astOllama(1000),
		astValidatorFailed(1000),
		astMalformed(1000),
	}

	corpus := make([]Fixture, len(responses))
	for i := range corpus {
		corpus[i] = newSynthFixture(fmt.Sprintf("f%02d", i), nil)
	}

	adapter := &sequentialMockAdapter{responses: responses}
	card := Score(t, adapter, corpus)
	require.NotNil(t, card, "Score must return a non-nil scorecard")

	assert.Equal(t, 10, card.Total)
	assert.Equal(t, 9, card.ValidJSON, "8 ollama + 1 validator-failed = 9 valid JSON")
	assert.Equal(t, 8, card.ValidatorPassed)

	assert.InDelta(t, 0.9, card.ValidJSONRate(), 0.001)
	assert.InDelta(t, 0.8, card.ValidatorPassRate(), 0.001)
	assert.Equal(t, time.Second, card.P95Latency(),
		"uniform 1s latencies → P95 = 1s under any percentile formula")
}

// TestEval_PerWidgetPrecisionRecall — AC-4: expected ["choice","multi"] vs
// produced ["choice"] → choice TP=1/FP=0/FN=0; multi TP=0/FP=0/FN=1.
func TestEval_PerWidgetPrecisionRecall(t *testing.T) {
	t.Parallel()

	corpus := []Fixture{
		{
			ID:  "pr1",
			Raw: "synth",
			Expected: Expected{
				Category:    "freeform",
				WidgetTypes: []string{"choice", "multi"},
			},
		},
	}

	produced := &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "ollama:gemma3:4b",
		Nodes: []uiadapter.UINode{
			{
				Type:   "decision_group",
				Widget: &uiadapter.WidgetNode{Type: "choice"},
			},
		},
		Diagnostics: uiadapter.Diagnostics{LatencyMs: 100},
	}

	adapter := &sequentialMockAdapter{responses: []*uiadapter.UIAST{produced}}
	card := Score(t, adapter, corpus)
	require.NotNil(t, card)

	assert.Equal(t, 1, card.PerWidgetTP["choice"])
	assert.Equal(t, 0, card.PerWidgetFP["choice"])
	assert.Equal(t, 0, card.PerWidgetFN["choice"])

	assert.Equal(t, 0, card.PerWidgetTP["multi"])
	assert.Equal(t, 0, card.PerWidgetFP["multi"])
	assert.Equal(t, 1, card.PerWidgetFN["multi"])

	assert.InDelta(t, 1.0, card.PerWidgetPrecision("choice"), 0.001)
	assert.InDelta(t, 1.0, card.PerWidgetRecall("choice"), 0.001)

	// multi: TP=0, FP=0 → P = 0/0 must be divide-by-zero-guarded to 0.00.
	assert.InDelta(t, 0.0, card.PerWidgetPrecision("multi"), 0.001)
	assert.InDelta(t, 0.0, card.PerWidgetRecall("multi"), 0.001)
}

// TestEval_Scorecard_PrettyPrint — AC-7: pretty-print includes model,
// promptVersion, rate figures, and P95 so drift is visible in logs
// (§4.5 line 330).
func TestEval_Scorecard_PrettyPrint(t *testing.T) {
	t.Parallel()

	card := &Scorecard{
		Model:           "gemma3:4b",
		PromptVersion:   "v1",
		Total:           10,
		ValidJSON:       10,
		ValidatorPassed: 10,
		Latencies:       []time.Duration{time.Second},
	}

	out := card.PrettyPrint()
	assert.Contains(t, out, "gemma3:4b",
		"AC-7: model name must surface so regression in validation-pass rate is visible on model swap")
	assert.Contains(t, out, "v1",
		"AC-7: promptVersion must surface so prompt-edit drift shows without a code change")
	assert.Contains(t, out, "100",
		"pretty-print must cite the actual valid-JSON / validator-pass rate numbers")
	assert.Contains(t, out, "P95",
		"pretty-print must label the P95 latency line")
}

// TestEval_MeetsThresholds_Table — AC-3: §4.5 thresholds enforced
// (≥ 90% valid-JSON, ≥ 80% validator-pass, P95 ≤ 3s).
func TestEval_MeetsThresholds_Table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		scorecard *Scorecard
		wantErr   string // "" = expect nil
	}{
		{
			name: "low_valid_json_85pct",
			scorecard: &Scorecard{
				Total:           100,
				ValidJSON:       85,
				ValidatorPassed: 85,
				Latencies:       uniformLatencies(10, time.Second),
			},
			wantErr: "90%",
		},
		{
			name: "low_validator_pass_75pct_json_ok",
			scorecard: &Scorecard{
				Total:           100,
				ValidJSON:       100,
				ValidatorPassed: 75,
				Latencies:       uniformLatencies(10, time.Second),
			},
			wantErr: "80%",
		},
		{
			name: "p95_above_3s_budget",
			scorecard: &Scorecard{
				Total:           10,
				ValidJSON:       10,
				ValidatorPassed: 10,
				Latencies:       uniformLatencies(10, 4*time.Second),
			},
			wantErr: "3s",
		},
		{
			name: "all_green",
			scorecard: &Scorecard{
				Total:           100,
				ValidJSON:       100,
				ValidatorPassed: 100,
				Latencies:       uniformLatencies(10, time.Second),
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.scorecard.MeetsThresholds()
			if tt.wantErr == "" {
				assert.NoError(t, err, "all-green scorecard must pass MeetsThresholds")
				return
			}
			require.Error(t, err, "MeetsThresholds must return an error when a target is missed")
			assert.Contains(t, err.Error(), tt.wantErr,
				"error message must cite the failing threshold (%q)", tt.wantErr)
		})
	}
}

// TestEval_Preservation_URLAndCodeBlockCounts — §7.2: a raw URL / fenced
// code block present in a fixture's expected set must be counted preserved
// when the produced AST's rendered text contains the verbatim string, and
// counted dropped otherwise. Exercises the recordPreservation path which
// the base metric tests do not hit (their fixtures omit MustContain*).
func TestEval_Preservation_URLAndCodeBlockCounts(t *testing.T) {
	t.Parallel()

	corpus := []Fixture{
		{
			ID:  "preserved",
			Raw: "synth",
			Expected: Expected{
				Category:              "adversarial",
				MustContainURLs:       []string{"https://example.com"},
				MustContainCodeBlocks: []string{"```go\nfmt.Println()\n```"},
			},
		},
		{
			ID:  "dropped",
			Raw: "synth",
			Expected: Expected{
				Category:              "adversarial",
				MustContainURLs:       []string{"https://dropped.example"},
				MustContainCodeBlocks: []string{"```py\nmissing\n```"},
			},
		},
	}

	preservedAST := &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "ollama:gemma3:4b",
		Nodes: []uiadapter.UINode{
			{Type: "markdown", Content: "See https://example.com\n```go\nfmt.Println()\n```"},
		},
		Diagnostics: uiadapter.Diagnostics{LatencyMs: 100},
	}
	droppedAST := &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "ollama:gemma3:4b",
		Nodes: []uiadapter.UINode{
			{Type: "markdown", Content: "nothing relevant"},
		},
		Diagnostics: uiadapter.Diagnostics{LatencyMs: 100},
	}

	adapter := &sequentialMockAdapter{responses: []*uiadapter.UIAST{preservedAST, droppedAST}}
	card := Score(t, adapter, corpus)
	require.NotNil(t, card)

	assert.Equal(t, 1, card.URLsPreserved)
	assert.Equal(t, 1, card.URLsDropped)
	assert.Equal(t, 1, card.CodeBlocksPreserved)
	assert.Equal(t, 1, card.CodeBlocksDropped)
}

// TestEval_ScorecardRates_EmptyCorpus — empty-corpus guard: rates must be
// zero-valued (not NaN) so a skipped harness still serialises cleanly.
func TestEval_ScorecardRates_EmptyCorpus(t *testing.T) {
	t.Parallel()
	var s Scorecard
	assert.Equal(t, 0.0, s.ValidJSONRate())
	assert.Equal(t, 0.0, s.ValidatorPassRate())
	assert.Equal(t, time.Duration(0), s.P95Latency())
}

// TestEval_PrettyPrint_IncludesPerWidgetLines — ensures the per-widget
// precision/recall block renders one line per known widget in sorted order
// so drift across widget types surfaces in logs.
func TestEval_PrettyPrint_IncludesPerWidgetLines(t *testing.T) {
	t.Parallel()
	card := &Scorecard{
		Model:         "gemma3:4b",
		PromptVersion: "v1",
		Total:         1,
		ValidJSON:     1,
		PerWidgetTP:   map[string]int{"choice": 1},
		PerWidgetFP:   map[string]int{"multi": 1},
		PerWidgetFN:   map[string]int{"free": 1},
		Latencies:     []time.Duration{500 * time.Millisecond},
	}
	out := card.PrettyPrint()
	assert.Contains(t, out, "choice")
	assert.Contains(t, out, "multi")
	assert.Contains(t, out, "free")
}

func uniformLatencies(n int, d time.Duration) []time.Duration {
	out := make([]time.Duration, n)
	for i := range out {
		out[i] = d
	}
	return out
}
