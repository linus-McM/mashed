package eval

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	"mashed/internal/uiadapter"
)

// Scorecard holds every measured metric for one eval run (§4.5). Valid-JSON
// and validator-pass counts are stored as raw integers rather than pre-computed
// rates so the threshold-check error messages can cite "N of M" shape and the
// pretty-printer can format percentages itself.
type Scorecard struct {
	Model         string
	PromptVersion string

	Total           int
	ValidJSON       int
	ValidatorPassed int

	PerWidgetTP map[string]int
	PerWidgetFP map[string]int
	PerWidgetFN map[string]int

	URLsPreserved       int
	URLsDropped         int
	CodeBlocksPreserved int
	CodeBlocksDropped   int

	Latencies []time.Duration
}

// Thresholds per spec §4.5. validJSONThreshold is the "valid-JSON rate"
// target; validatorPassThreshold is the "validator-pass rate" target.
// p95Budget keeps the harness aligned with §4.7.3's 3s hard timeout so a
// green eval also means the real user-facing path is within SLA.
const (
	validJSONThreshold     = 0.90
	validatorPassThreshold = 0.80
	p95Budget              = 3 * time.Second
)

// Score runs every fixture through adapter.Translate, classifies the result,
// and returns a populated Scorecard. t is accepted (not a pure-func input)
// so callers can log diagnostics via t.Logf on non-fatal anomalies without
// exposing a second "logger" parameter.
func Score(t *testing.T, adapter uiadapter.Adapter, corpus []Fixture) *Scorecard {
	t.Helper()
	card := &Scorecard{
		PromptVersion: uiadapter.PromptVersion(),
		PerWidgetTP:   map[string]int{},
		PerWidgetFP:   map[string]int{},
		PerWidgetFN:   map[string]int{},
		Latencies:     make([]time.Duration, 0, len(corpus)),
	}

	for _, fx := range corpus {
		ast := adapter.Translate(context.Background(), fx.Raw, "eval")
		card.record(fx, ast)
	}
	return card
}

// record classifies one Translate result against its expected label set.
// Kept as a method on *Scorecard rather than a free function so the call-site
// in Score stays a single-line loop body.
func (s *Scorecard) record(fx Fixture, ast *uiadapter.UIAST) {
	s.Total++
	if ast == nil {
		return
	}

	if s.Model == "" {
		s.Model = inferModel(ast.GeneratedBy)
	}
	s.Latencies = append(s.Latencies, time.Duration(ast.Diagnostics.LatencyMs)*time.Millisecond)

	validJSON, validatorPassed := classifyGeneratedBy(ast.GeneratedBy)
	if validJSON {
		s.ValidJSON++
	}
	if validatorPassed {
		s.ValidatorPassed++
	}
	s.recordWidgets(fx.Expected.WidgetTypes, widgetTypesFrom(ast))
	s.recordPreservation(fx.Expected, ast)
}

func (s *Scorecard) recordWidgets(expected, produced []string) {
	expSet := toSet(expected)
	prodSet := toSet(produced)
	for w := range expSet {
		if _, ok := prodSet[w]; ok {
			s.PerWidgetTP[w]++
		} else {
			s.PerWidgetFN[w]++
		}
	}
	for w := range prodSet {
		if _, ok := expSet[w]; !ok {
			s.PerWidgetFP[w]++
		}
	}
}

func (s *Scorecard) recordPreservation(exp Expected, ast *uiadapter.UIAST) {
	haystack := renderedText(ast)
	for _, u := range exp.MustContainURLs {
		if strings.Contains(haystack, u) {
			s.URLsPreserved++
		} else {
			s.URLsDropped++
		}
	}
	for _, c := range exp.MustContainCodeBlocks {
		if strings.Contains(haystack, c) {
			s.CodeBlocksPreserved++
		} else {
			s.CodeBlocksDropped++
		}
	}
}

// ValidJSONRate is Total-guarded so an empty corpus does not panic.
func (s *Scorecard) ValidJSONRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.ValidJSON) / float64(s.Total)
}

func (s *Scorecard) ValidatorPassRate() float64 {
	if s.Total == 0 {
		return 0
	}
	return float64(s.ValidatorPassed) / float64(s.Total)
}

// P95Latency uses nearest-rank: sort ascending and return the ceil(0.95 * n)-th
// element (1-indexed). For uniform latencies this trivially matches the
// expected constant so the AC-2 test can assert "P95 == 1s" on n=10 1s
// samples without a percentile-interpolation tolerance.
func (s *Scorecard) P95Latency() time.Duration {
	if len(s.Latencies) == 0 {
		return 0
	}
	sorted := append([]time.Duration(nil), s.Latencies...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	rank := int(math.Ceil(0.95 * float64(len(sorted))))
	if rank < 1 {
		rank = 1
	}
	if rank > len(sorted) {
		rank = len(sorted)
	}
	return sorted[rank-1]
}

// PerWidgetPrecision returns TP / (TP + FP). Divide-by-zero (no TP or FP
// observed for this widget) returns 0 per AC-4; the test asserts that
// "0/0 → 0.00" instead of NaN so scorecards serialise cleanly to logs.
func (s *Scorecard) PerWidgetPrecision(widget string) float64 {
	tp := s.PerWidgetTP[widget]
	fp := s.PerWidgetFP[widget]
	if tp+fp == 0 {
		return 0
	}
	return float64(tp) / float64(tp+fp)
}

func (s *Scorecard) PerWidgetRecall(widget string) float64 {
	tp := s.PerWidgetTP[widget]
	fn := s.PerWidgetFN[widget]
	if tp+fn == 0 {
		return 0
	}
	return float64(tp) / float64(tp+fn)
}

// MeetsThresholds enforces §4.5: ≥ 90% valid-JSON, ≥ 80% validator-pass,
// and P95 latency ≤ 3s. Error messages deliberately cite the failing figure
// ("85.00%" / "4s") so a CI failure log shows exactly which gate closed
// without needing to inspect the scorecard struct.
func (s *Scorecard) MeetsThresholds() error {
	if r := s.ValidJSONRate(); r < validJSONThreshold {
		return fmt.Errorf("valid-JSON rate %.2f%% below 90%% threshold (§4.5)", 100*r)
	}
	if r := s.ValidatorPassRate(); r < validatorPassThreshold {
		return fmt.Errorf("validator-pass rate %.2f%% below 80%% threshold (§4.5)", 100*r)
	}
	if p95 := s.P95Latency(); p95 > p95Budget {
		return fmt.Errorf("P95 latency %s exceeds 3s budget (§4.7.3)", p95)
	}
	return nil
}

// PrettyPrint renders the scorecard as the §4.5-style block pasted into
// developer logs. Model + PromptVersion surface first so a swap shows up
// unambiguously; rate figures use one decimal so "90.0%" does not round
// over the threshold silently.
func (s *Scorecard) PrettyPrint() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Gemma UI AST eval — %s — prompt %s — %s\n",
		s.Model, s.PromptVersion, time.Now().UTC().Format("2006-01-02 15:04"))
	b.WriteString(strings.Repeat("─", 60))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "Total fixtures       : %d\n", s.Total)
	fmt.Fprintf(&b, "Valid JSON           : %d (%.1f%%) target ≥ 90%%\n",
		s.ValidJSON, 100*s.ValidJSONRate())
	fmt.Fprintf(&b, "Validator passed     : %d (%.1f%%) target ≥ 80%%\n",
		s.ValidatorPassed, 100*s.ValidatorPassRate())
	fmt.Fprintf(&b, "URLs preserved       : %d/%d\n",
		s.URLsPreserved, s.URLsPreserved+s.URLsDropped)
	fmt.Fprintf(&b, "Code blocks          : %d/%d\n",
		s.CodeBlocksPreserved, s.CodeBlocksPreserved+s.CodeBlocksDropped)
	fmt.Fprintf(&b, "P95 latency          : %s (budget ≤ 3s)\n", s.P95Latency())
	b.WriteByte('\n')
	b.WriteString("Per-widget (P / R):\n")
	for _, w := range sortedWidgetNames(s) {
		fmt.Fprintf(&b, "  %-10s: TP=%d FP=%d FN=%d (%.2f / %.2f)\n",
			w, s.PerWidgetTP[w], s.PerWidgetFP[w], s.PerWidgetFN[w],
			s.PerWidgetPrecision(w), s.PerWidgetRecall(w))
	}
	b.WriteString(strings.Repeat("─", 60))
	b.WriteByte('\n')
	return b.String()
}

// classifyGeneratedBy maps a UIAST.GeneratedBy tag onto (validJSON,
// validatorPassed). Rules:
//   - "ollama:*" — the adapter accepted the model's output end-to-end: JSON
//     parsed AND validator allowed it. Both true.
//   - "fallback:validation:malformed" — JSON parse failed: both false.
//   - "fallback:validation:*" (other) — JSON parsed but validator rejected
//     (required_dropped / oversize / marshal_error): valid JSON, validator
//     failed.
//   - "fallback:*" (any other) — transport / saturation / disabled: not a
//     model-output sample, treat as invalid JSON so these do not mask model
//     drift by inflating the denominator's numerator.
func classifyGeneratedBy(tag string) (validJSON, validatorPassed bool) {
	switch {
	case strings.HasPrefix(tag, "ollama:"):
		return true, true
	case tag == "fallback:validation:malformed":
		return false, false
	case strings.HasPrefix(tag, "fallback:validation:"):
		return true, false
	default:
		return false, false
	}
}

func inferModel(tag string) string {
	const prefix = "ollama:"
	if strings.HasPrefix(tag, prefix) {
		return strings.TrimPrefix(tag, prefix)
	}
	return ""
}

func widgetTypesFrom(ast *uiadapter.UIAST) []string {
	if ast == nil {
		return nil
	}
	var out []string
	for _, n := range ast.Nodes {
		if n.Widget != nil && n.Widget.Type != "" {
			out = append(out, n.Widget.Type)
		}
	}
	return out
}

func renderedText(ast *uiadapter.UIAST) string {
	if ast == nil {
		return ""
	}
	var b strings.Builder
	for _, n := range ast.Nodes {
		b.WriteString(n.Content)
		b.WriteByte('\n')
		b.WriteString(n.Heading)
		b.WriteByte('\n')
		b.WriteString(n.Prompt)
		b.WriteByte('\n')
		for _, bullet := range n.Bullets {
			b.WriteString(bullet)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func toSet(ss []string) map[string]struct{} {
	out := make(map[string]struct{}, len(ss))
	for _, s := range ss {
		if s == "" {
			continue
		}
		out[s] = struct{}{}
	}
	return out
}

func sortedWidgetNames(s *Scorecard) []string {
	seen := map[string]struct{}{}
	for w := range s.PerWidgetTP {
		seen[w] = struct{}{}
	}
	for w := range s.PerWidgetFP {
		seen[w] = struct{}{}
	}
	for w := range s.PerWidgetFN {
		seen[w] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for w := range seen {
		out = append(out, w)
	}
	sort.Strings(out)
	return out
}
