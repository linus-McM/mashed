// Package eval — Story v3-13. Extends the ui-ast-U9 scorecard with
// per-backend thresholds + cross-backend shadow mode per plan §3 Story 13.
package eval

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// BackendThresholds encodes the per-backend SLOs from plan §3 Story 13.
// | Backend              | parse_rate | terminal_validation | p50_warm | p95_warm |
// | Ollama               | ≥ 0.98     | ≤ 0.02              | ≤ 600ms  | ≤ 1500ms |
// | Haiku (cached)       | ≥ 0.99     | ≤ 0.005             | ≤ 400ms  | ≤ 1000ms |
// | Sonnet (single-shot) | ≥ 0.995    | ≤ 0.002             | ≤ 800ms  | ≤ 2000ms |
// | Claude CLI           | ≥ 0.95     | ≤ 0.02              | ≤ 2000ms | ≤ 5000ms |
type BackendThresholds struct {
	MinParseRate          float64
	MaxTerminalValidation float64
	MaxP50Warm            time.Duration
	MaxP95Warm            time.Duration
}

// PerBackendThresholds maps backend names to their v3.0 SLOs.
var PerBackendThresholds = map[string]BackendThresholds{
	"ollama": {
		MinParseRate: 0.98, MaxTerminalValidation: 0.02,
		MaxP50Warm: 600 * time.Millisecond, MaxP95Warm: 1500 * time.Millisecond,
	},
	"claude-haiku-4-5": {
		MinParseRate: 0.99, MaxTerminalValidation: 0.005,
		MaxP50Warm: 400 * time.Millisecond, MaxP95Warm: 1000 * time.Millisecond,
	},
	"claude-sonnet-4-6": {
		MinParseRate: 0.995, MaxTerminalValidation: 0.002,
		MaxP50Warm: 800 * time.Millisecond, MaxP95Warm: 2000 * time.Millisecond,
	},
	"claude-cli": {
		MinParseRate: 0.95, MaxTerminalValidation: 0.02,
		MaxP50Warm: 2000 * time.Millisecond, MaxP95Warm: 5000 * time.Millisecond,
	},
}

// Row is a single scorecard entry per plan §3 Story 13 column set.
type Row struct {
	Backend                  string
	Model                    string
	SingleShot               bool
	FastPathHit              bool
	FastPathRule             string
	CacheHit                 bool
	Stage1Correct            bool
	Stage2ParseOK            bool
	Stage2ValidationTerminal bool
	Stage2Untrusted          bool
	RepairFired              bool
	RepairSucceeded          bool
	BreakerTripped           bool
	EscalatedFrom            string
	LatencyMsCold            int
	LatencyMsWarm            int
	TokensIn                 int
	TokensOut                int
	CacheReadInputTokens     int
	CacheCreationInputTokens int
	USDCostEst               float64
}

// ScorecardV3 aggregates Rows and per-backend threshold enforcement.
type ScorecardV3 struct {
	mu   sync.Mutex
	rows []Row
}

// NewScorecardV3 constructs an empty scorecard.
func NewScorecardV3() *ScorecardV3 { return &ScorecardV3{} }

// Add records a row; safe for concurrent calls.
func (s *ScorecardV3) Add(r Row) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rows = append(s.rows, r)
}

// Rows returns a copy of the underlying rows.
func (s *ScorecardV3) Rows() []Row {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Row, len(s.rows))
	copy(out, s.rows)
	return out
}

// Aggregate computes the metric bundle per backend+model: parse rate,
// terminal-validation rate, p50/p95 warm latency.
type Aggregate struct {
	Backend                 string
	Model                   string
	Rows                    int
	ParseRate               float64
	TerminalValidation      float64
	P50Warm                 time.Duration
	P95Warm                 time.Duration
	FastPathHitRate         float64
	CacheHitRate            float64
}

// AggregateByBackend collapses rows by (backend, model) and returns a
// sorted slice for stable scorecard output.
func (s *ScorecardV3) AggregateByBackend() []Aggregate {
	s.mu.Lock()
	defer s.mu.Unlock()
	groups := map[string]*Aggregate{}
	latencies := map[string][]int{}
	for _, r := range s.rows {
		key := r.Backend + "|" + r.Model
		agg, ok := groups[key]
		if !ok {
			agg = &Aggregate{Backend: r.Backend, Model: r.Model}
			groups[key] = agg
		}
		agg.Rows++
		if r.Stage2ParseOK {
			agg.ParseRate++
		}
		if r.Stage2ValidationTerminal {
			agg.TerminalValidation++
		}
		if r.FastPathHit {
			agg.FastPathHitRate++
		}
		if r.CacheHit {
			agg.CacheHitRate++
		}
		if r.LatencyMsWarm > 0 {
			latencies[key] = append(latencies[key], r.LatencyMsWarm)
		}
	}
	for k, agg := range groups {
		if agg.Rows == 0 {
			continue
		}
		agg.ParseRate /= float64(agg.Rows)
		agg.TerminalValidation /= float64(agg.Rows)
		agg.FastPathHitRate /= float64(agg.Rows)
		agg.CacheHitRate /= float64(agg.Rows)
		lat := latencies[k]
		sort.Ints(lat)
		if len(lat) > 0 {
			agg.P50Warm = time.Duration(lat[len(lat)/2]) * time.Millisecond
			agg.P95Warm = time.Duration(lat[percentileIdx(len(lat), 0.95)]) * time.Millisecond
		}
	}
	var out []Aggregate
	for _, v := range groups {
		out = append(out, *v)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Backend != out[j].Backend {
			return out[i].Backend < out[j].Backend
		}
		return out[i].Model < out[j].Model
	})
	return out
}

func percentileIdx(n int, p float64) int {
	idx := int(float64(n-1) * p)
	if idx < 0 {
		return 0
	}
	if idx >= n {
		return n - 1
	}
	return idx
}

// MeetsThresholds asserts every aggregate row against its backend's
// thresholds. Returns a combined error naming each violation. AC-13.2.
func (s *ScorecardV3) MeetsThresholds() error {
	var violations []string
	for _, agg := range s.AggregateByBackend() {
		// Choose thresholds by Model when backend is claude-api (haiku vs sonnet).
		key := agg.Backend
		if key == "claude-api" {
			key = agg.Model
		}
		th, ok := PerBackendThresholds[key]
		if !ok {
			continue
		}
		if agg.ParseRate < th.MinParseRate {
			violations = append(violations, fmt.Sprintf("%s/%s parse_rate %.3f < %.3f", agg.Backend, agg.Model, agg.ParseRate, th.MinParseRate))
		}
		if agg.TerminalValidation > th.MaxTerminalValidation {
			violations = append(violations, fmt.Sprintf("%s/%s terminal_validation %.3f > %.3f", agg.Backend, agg.Model, agg.TerminalValidation, th.MaxTerminalValidation))
		}
		if th.MaxP50Warm > 0 && agg.P50Warm > th.MaxP50Warm {
			violations = append(violations, fmt.Sprintf("%s/%s p50_warm %v > %v", agg.Backend, agg.Model, agg.P50Warm, th.MaxP50Warm))
		}
		if th.MaxP95Warm > 0 && agg.P95Warm > th.MaxP95Warm {
			violations = append(violations, fmt.Sprintf("%s/%s p95_warm %v > %v", agg.Backend, agg.Model, agg.P95Warm, th.MaxP95Warm))
		}
	}
	if len(violations) == 0 {
		return nil
	}
	return fmt.Errorf("scorecard-v3 threshold violations:\n  - %s", strings.Join(violations, "\n  - "))
}

// PrettyPrint renders a grouped-by-backend ASCII table. AC-13.1.
func (s *ScorecardV3) PrettyPrint() string {
	var b strings.Builder
	b.WriteString("Backend | Model | Rows | Parse | Term | P50 | P95 | FastPath | Cache\n")
	b.WriteString(strings.Repeat("-", 90) + "\n")
	for _, agg := range s.AggregateByBackend() {
		fmt.Fprintf(&b, "%s | %s | %d | %.3f | %.3f | %v | %v | %.2f | %.2f\n",
			agg.Backend, agg.Model, agg.Rows,
			agg.ParseRate, agg.TerminalValidation,
			agg.P50Warm, agg.P95Warm,
			agg.FastPathHitRate, agg.CacheHitRate)
	}
	return b.String()
}

// CrossBackendDelta renders the backend × metric delta table. AC-13.4.
func (s *ScorecardV3) CrossBackendDelta() string {
	aggs := s.AggregateByBackend()
	var b strings.Builder
	b.WriteString("Metric | " + joinBackends(aggs) + "\n")
	for _, metric := range []string{"ParseRate", "TerminalValidation", "P50Warm"} {
		b.WriteString(metric)
		for _, agg := range aggs {
			switch metric {
			case "ParseRate":
				fmt.Fprintf(&b, " | %.3f", agg.ParseRate)
			case "TerminalValidation":
				fmt.Fprintf(&b, " | %.3f", agg.TerminalValidation)
			case "P50Warm":
				fmt.Fprintf(&b, " | %v", agg.P50Warm)
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func joinBackends(aggs []Aggregate) string {
	names := make([]string, len(aggs))
	for i, a := range aggs {
		names[i] = a.Backend + "/" + a.Model
	}
	return strings.Join(names, " | ")
}

// ShadowSampler implements plan §3 Story 13 "5% sampled shadow mode".
// Concurrency-safe; counter-based so tests can deterministically verify
// a sample rate by setting it to 1.0 or 0.0.
type ShadowSampler struct {
	rate    float64
	calls   atomic.Int64
	sampled atomic.Int64
}

// NewShadowSampler constructs a sampler with rate in [0.0, 1.0]. Outside
// that range the sampler is a no-op.
func NewShadowSampler(rate float64) *ShadowSampler { return &ShadowSampler{rate: rate} }

// ShouldSample returns true with probability `rate`. Deterministic via a
// modulo rather than RNG so tests stay flake-free.
func (s *ShadowSampler) ShouldSample() bool {
	n := s.calls.Add(1)
	if s.rate <= 0 {
		return false
	}
	if s.rate >= 1.0 {
		s.sampled.Add(1)
		return true
	}
	// Deterministic: every ceil(1/rate)-th call samples.
	step := int64(1.0 / s.rate)
	if step < 1 {
		step = 1
	}
	if n%step == 0 {
		s.sampled.Add(1)
		return true
	}
	return false
}

// Counts returns (total, sampled) snapshot for telemetry / tests.
func (s *ShadowSampler) Counts() (int64, int64) {
	return s.calls.Load(), s.sampled.Load()
}
