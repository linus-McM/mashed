package eval

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScorecardV3_AggregateByBackend — AC-13.1 precondition. Per-(backend,
// model) aggregation computes parse rate + latency percentiles.
func TestScorecardV3_AggregateByBackend(t *testing.T) {
	t.Parallel()
	s := NewScorecardV3()
	s.Add(Row{Backend: "ollama", Model: "gemma3:4b", Stage2ParseOK: true, LatencyMsWarm: 200})
	s.Add(Row{Backend: "ollama", Model: "gemma3:4b", Stage2ParseOK: true, LatencyMsWarm: 400})
	s.Add(Row{Backend: "ollama", Model: "gemma3:4b", Stage2ParseOK: false, LatencyMsWarm: 600})
	aggs := s.AggregateByBackend()
	require.Len(t, aggs, 1)
	assert.Equal(t, "ollama", aggs[0].Backend)
	assert.InDelta(t, 2.0/3.0, aggs[0].ParseRate, 0.01)
	assert.Equal(t, 400*time.Millisecond, aggs[0].P50Warm)
}

// TestScorecardV3_MeetsThresholds — AC-13.2. Violating the Ollama SLO
// surfaces a named violation.
func TestScorecardV3_MeetsThresholds(t *testing.T) {
	t.Parallel()
	s := NewScorecardV3()
	// Make parse rate deliberately low (20 rows, only 10 pass = 0.5 < 0.98).
	for i := 0; i < 10; i++ {
		s.Add(Row{Backend: "ollama", Model: "gemma3:4b", Stage2ParseOK: true, LatencyMsWarm: 200})
	}
	for i := 0; i < 10; i++ {
		s.Add(Row{Backend: "ollama", Model: "gemma3:4b", Stage2ParseOK: false, LatencyMsWarm: 200})
	}
	err := s.MeetsThresholds()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parse_rate")
	assert.Contains(t, err.Error(), "ollama")
}

// TestScorecardV3_MeetsThresholds_Pass — healthy row set satisfies SLOs.
func TestScorecardV3_MeetsThresholds_Pass(t *testing.T) {
	t.Parallel()
	s := NewScorecardV3()
	for i := 0; i < 100; i++ {
		s.Add(Row{Backend: "ollama", Model: "gemma3:4b", Stage2ParseOK: true, LatencyMsWarm: 300})
	}
	require.NoError(t, s.MeetsThresholds())
}

// TestScorecardV3_PrettyPrint_GroupsByBackend — AC-13.1. Output carries
// one row per (backend, model) with every metric column.
func TestScorecardV3_PrettyPrint_GroupsByBackend(t *testing.T) {
	t.Parallel()
	s := NewScorecardV3()
	s.Add(Row{Backend: "ollama", Model: "gemma3:4b", Stage2ParseOK: true, LatencyMsWarm: 100})
	s.Add(Row{Backend: "claude-api", Model: "claude-haiku-4-5", Stage2ParseOK: true, LatencyMsWarm: 200})
	out := s.PrettyPrint()
	assert.Contains(t, out, "ollama")
	assert.Contains(t, out, "claude-api")
	assert.Contains(t, out, "gemma3:4b")
	assert.Contains(t, out, "claude-haiku-4-5")
}

// TestScorecardV3_CrossBackendDelta — AC-13.4. Diff table prints one
// column per backend.
func TestScorecardV3_CrossBackendDelta(t *testing.T) {
	t.Parallel()
	s := NewScorecardV3()
	s.Add(Row{Backend: "ollama", Model: "gemma3:4b", Stage2ParseOK: true})
	s.Add(Row{Backend: "claude-api", Model: "claude-haiku-4-5", Stage2ParseOK: true})
	delta := s.CrossBackendDelta()
	assert.Contains(t, delta, "ParseRate")
	assert.True(t, strings.Count(delta, "|") >= 2)
}

// TestShadowSampler_Rate — sample rate respected deterministically.
func TestShadowSampler_Rate(t *testing.T) {
	t.Parallel()
	s := NewShadowSampler(0.10) // every 10th
	sampled := 0
	for i := 0; i < 100; i++ {
		if s.ShouldSample() {
			sampled++
		}
	}
	assert.Equal(t, 10, sampled, "0.10 rate over 100 calls = 10 samples")

	total, sampCount := s.Counts()
	assert.EqualValues(t, 100, total)
	assert.EqualValues(t, 10, sampCount)
}

// TestShadowSampler_Extremes — rate 0 never samples; rate 1 always samples.
func TestShadowSampler_Extremes(t *testing.T) {
	t.Parallel()
	zero := NewShadowSampler(0)
	for i := 0; i < 10; i++ {
		assert.False(t, zero.ShouldSample())
	}
	one := NewShadowSampler(1.0)
	for i := 0; i < 10; i++ {
		assert.True(t, one.ShouldSample())
	}
}

// TestScorecardV3_ClaudeAPIModelSpecificThreshold — the model column,
// not backend, picks SLO for claude-api (Haiku vs Sonnet).
func TestScorecardV3_ClaudeAPIModelSpecificThreshold(t *testing.T) {
	t.Parallel()
	s := NewScorecardV3()
	// Sonnet SLO: parse_rate ≥ 0.995. 100 rows at 0.98 violates.
	for i := 0; i < 98; i++ {
		s.Add(Row{Backend: "claude-api", Model: "claude-sonnet-4-6", Stage2ParseOK: true})
	}
	for i := 0; i < 2; i++ {
		s.Add(Row{Backend: "claude-api", Model: "claude-sonnet-4-6", Stage2ParseOK: false})
	}
	err := s.MeetsThresholds()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claude-sonnet-4-6")
}
