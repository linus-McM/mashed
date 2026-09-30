package uiadapter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAccountant_CostMatchesBilling — AC-11b.2. Cost accumulator matches
// the pricing table within ±1%. Haiku: $0.80 in / $4 out per MTok.
func TestAccountant_CostMatchesBilling(t *testing.T) {
	t.Parallel()
	u := Usage{
		Model:        "claude-haiku-4-5",
		InputTokens:  1_000_000,
		OutputTokens: 500_000,
	}
	got := CostUSD(u)
	want := 0.80 + 2.00 // $0.80 input + $2.00 output
	assert.InDelta(t, want, got, want*0.01)
}

// TestAccountant_TripsBeforeHard429 — AC-11b.1. At soft limit the
// accountant returns ErrAccountantLimit proactively.
func TestAccountant_TripsBeforeHard429(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.RPMSoftLimit = 10
	a := NewAccountant(cfg)

	// Nine requests — still under the 85% soft line (≈8).
	// At the 8th record the accountant crosses soft; 9th checkPrecall trips.
	for i := 0; i < 8; i++ {
		require.NoError(t, a.CheckPrecall(100))
		a.Record(Usage{Model: "claude-haiku-4-5", InputTokens: 10, OutputTokens: 10})
	}
	err := a.CheckPrecall(100)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAccountantLimit)
}

// TestAccountant_USDBudgetSoftLimit — AC-11b.1 USD branch.
func TestAccountant_USDBudgetSoftLimit(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.UsdBudgetPerSession = 1.00
	a := NewAccountant(cfg)

	// One Haiku call at 1M/500K = $2.80 — instantly exceeds 85% of $1.
	a.Record(Usage{Model: "claude-haiku-4-5", InputTokens: 1_000_000, OutputTokens: 500_000})
	err := a.CheckPrecall(0)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrAccountantLimit)
}

// TestAccountant_Snapshot — metrics attribution surface for §6.1 slog.
func TestAccountant_Snapshot(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	a := NewAccountant(cfg)
	a.Record(Usage{Model: "claude-haiku-4-5", InputTokens: 100, OutputTokens: 200})
	a.Record(Usage{Model: "claude-haiku-4-5", InputTokens: 100, OutputTokens: 200})
	reqs, tok, spend := a.Snapshot()
	assert.Equal(t, 2, reqs)
	assert.Equal(t, 600, tok)
	assert.Greater(t, spend, 0.0)
}

// TestAccountant_SlidingWindow — entries older than 60s are dropped.
func TestAccountant_SlidingWindow(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.RPMSoftLimit = 100
	a := NewAccountant(cfg)

	// Inject a fake clock.
	base := time.Now()
	a.now = func() time.Time { return base }
	a.Record(Usage{Model: "claude-haiku-4-5", InputTokens: 1, OutputTokens: 1})

	// Advance 61s — the record should expire.
	a.now = func() time.Time { return base.Add(61 * time.Second) }
	reqs, _, _ := a.Snapshot()
	assert.Zero(t, reqs, "records older than 60s are dropped")
}

// TestAccountant_UnknownModelIsFree — non-Claude models don't spend.
func TestAccountant_UnknownModelIsFree(t *testing.T) {
	t.Parallel()
	assert.Zero(t, CostUSD(Usage{Model: "gemma3:4b", InputTokens: 1000, OutputTokens: 1000}))
}
