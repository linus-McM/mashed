package uiadapter

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFallback_TieredRecovery — AC-11.5. Primary fails, secondary succeeds,
// result carries escalated_from=primary.
func TestFallback_TieredRecovery(t *testing.T) {
	t.Parallel()
	order := []string{"claude-api", "ollama"}
	fn := func(_ context.Context, name string) (*UIAST, error) {
		if name == "claude-api" {
			return nil, errors.New("claude down")
		}
		return &UIAST{Version: "1", GeneratedBy: name}, nil
	}
	minimal := func(context.Context) (*UIAST, error) { return nil, errors.New("no kind") }
	plaintext := func() *UIAST { return &UIAST{GeneratedBy: "fallback:plaintext"} }

	ast, tier, from, err := RunWithFallback(context.Background(), order, fn, minimal, plaintext, nil)
	require.NoError(t, err)
	require.NotNil(t, ast)
	assert.Equal(t, TierSecondary, tier)
	assert.Equal(t, "claude-api", from, "escalated_from names the failing primary")
	assert.Equal(t, "ollama", ast.GeneratedBy)
}

// TestFallback_MinimalKindWhenAllBackendsFail — every registered backend
// fails; minimal-kind rescues the call.
func TestFallback_MinimalKindWhenAllBackendsFail(t *testing.T) {
	t.Parallel()
	order := []string{"claude-api", "ollama"}
	fn := func(context.Context, string) (*UIAST, error) { return nil, errors.New("x") }
	minimal := func(context.Context) (*UIAST, error) {
		return &UIAST{GeneratedBy: "minimal-kind"}, nil
	}
	plaintext := func() *UIAST { return &UIAST{GeneratedBy: "fallback:plaintext"} }

	ast, tier, _, err := RunWithFallback(context.Background(), order, fn, minimal, plaintext, nil)
	require.NoError(t, err)
	require.NotNil(t, ast)
	assert.Equal(t, TierMinimal, tier)
	assert.Equal(t, "minimal-kind", ast.GeneratedBy)
}

// TestFallback_PlaintextLastResort — minimal-kind also fails → plaintext.
func TestFallback_PlaintextLastResort(t *testing.T) {
	t.Parallel()
	order := []string{"claude-api"}
	fn := func(context.Context, string) (*UIAST, error) { return nil, errors.New("x") }
	minimal := func(context.Context) (*UIAST, error) { return nil, errors.New("no") }
	plaintext := func() *UIAST { return &UIAST{GeneratedBy: "fallback:plaintext"} }

	ast, tier, from, err := RunWithFallback(context.Background(), order, fn, minimal, plaintext, nil)
	require.Error(t, err, "joined errors surface all causes")
	require.NotNil(t, ast)
	assert.Equal(t, TierPlaintext, tier)
	assert.Equal(t, "claude-api", from)
}

// TestFallback_PrimarySuccessEscalatedFromEmpty — happy path; no escalation.
func TestFallback_PrimarySuccessEscalatedFromEmpty(t *testing.T) {
	t.Parallel()
	fn := func(_ context.Context, name string) (*UIAST, error) {
		return &UIAST{GeneratedBy: name}, nil
	}
	ast, tier, from, err := RunWithFallback(context.Background(), []string{"ollama"}, fn, nil, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, ast)
	assert.Equal(t, TierPrimary, tier)
	assert.Empty(t, from)
}

// TestFallback_ContextCancellation — canceled ctx aborts before dispatch.
func TestFallback_ContextCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	fn := func(context.Context, string) (*UIAST, error) {
		t.Fatal("must not be invoked on canceled ctx")
		return nil, nil
	}
	_, _, _, err := RunWithFallback(ctx, []string{"any"}, fn, nil, func() *UIAST { return nil }, nil)
	require.ErrorIs(t, err, context.Canceled)
}

// --- Story 4: uiadapter-logging-4-instrument-pipeline -----------------------
//
// AC-4.6: fallback_tiers.go emits the tier-escalation flow.
//
// Required emissions per Story 4 dev notes:
//   - fallback.tier.start    (op=fallback.tier, tiers_count)
//   - fallback.tier.enter    (op, tier, attempt_index)
//   - fallback.tier.success  (op, tier, latency_ms)
//   - fallback.tier.failure  (op, tier, reason)
//   - fallback.tier.exhausted (op, tiers_tried)
//
// Per §14, tier names are closed-enum strings ("primary", "secondary",
// "minimal-kind", "plaintext"); reasons logged here are short-enum tokens
// drawn from the fallback failure-reason set ("transport", "validate",
// "breaker_open", …) — NEVER err.Error() text.

// TestStory4_AC6_TierThirdSuccess — Story 4, AC-4.6 (escalation success).
//
// Three tiers; first two fail with reason "transport"; third succeeds.
// Expected emission sequence:
//
//	fallback.tier.start    (tiers_count=3)
//	fallback.tier.enter    (idx=0)
//	fallback.tier.failure  (reason=transport)
//	fallback.tier.enter    (idx=1)
//	fallback.tier.failure  (reason=transport)
//	fallback.tier.enter    (idx=2)
//	fallback.tier.success  (latency_ms ≥ 0)
//
// `fallback.tier.exhausted` MUST NOT appear because tier 2 succeeded.
func TestStory4_AC6_TierThirdSuccess(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	// Backend names are enum-ish IDs and may be logged (§14 explicitly permits).
	order := []string{"t0", "t1", "t2"}
	fn := func(_ context.Context, name string) (*UIAST, error) {
		if name == "t2" {
			return &UIAST{Version: "1", GeneratedBy: name}, nil
		}
		// The reason must be drawn from the closed enum; the implementer is
		// expected to map a generic transport error to reason="transport".
		return nil, errors.New("transport")
	}
	plaintext := func() *UIAST { return &UIAST{GeneratedBy: "fallback:plaintext"} }

	ast, _, _, err := RunWithFallback(context.Background(), order, fn, nil, plaintext, logger)
	require.NoError(t, err)
	require.NotNil(t, ast)

	records := decodeRecords(t, buf)
	wantOrder := []string{
		"fallback.tier.start",
		"fallback.tier.enter",
		"fallback.tier.failure",
		"fallback.tier.enter",
		"fallback.tier.failure",
		"fallback.tier.enter",
		"fallback.tier.success",
	}
	require.Equal(t, wantOrder, recordMsgsWithPrefix(records, "fallback.tier."),
		"tier emissions must follow the 2-fail-then-success script; got %v", records)
	assert.Empty(t, recordsByMsg(records, "fallback.tier.exhausted"),
		"exhausted must not emit when a tier succeeded")

	start := recordsByMsg(records, "fallback.tier.start")[0]
	assert.Equal(t, "fallback.tier", start["op"])
	assert.EqualValues(t, len(order), start["tiers_count"])

	enters := recordsByMsg(records, "fallback.tier.enter")
	require.Len(t, enters, 3)
	for i, e := range enters {
		assert.Equal(t, "fallback.tier", e["op"])
		assert.EqualValues(t, i, e["attempt_index"], "attempt_index must be 0-indexed")
		assert.NotEmpty(t, e["tier"], "tier attr must be non-empty")
	}

	failures := recordsByMsg(records, "fallback.tier.failure")
	require.Len(t, failures, 2)
	for _, f := range failures {
		assert.Equal(t, "transport", f["reason"], "reason must be a closed-enum token")
	}

	success := recordsByMsg(records, "fallback.tier.success")[0]
	assert.Equal(t, "fallback.tier", success["op"])
	latencyMs, ok := success["latency_ms"].(float64)
	require.True(t, ok, "success must carry numeric latency_ms")
	assert.GreaterOrEqual(t, latencyMs, 0.0)
}

// TestStory4_AC6_TierAllExhausted — Story 4, AC-4.6 (all tiers exhausted).
//
// Three tiers all fail; no minimal-kind. Last record must be
// `fallback.tier.exhausted` with `tiers_tried=3`.
func TestStory4_AC6_TierAllExhausted(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	order := []string{"t0", "t1", "t2"}
	fn := func(context.Context, string) (*UIAST, error) {
		return nil, errors.New("transport")
	}
	plaintext := func() *UIAST { return &UIAST{GeneratedBy: "fallback:plaintext"} }

	_, _, _, err := RunWithFallback(context.Background(), order, fn, nil, plaintext, logger)
	require.Error(t, err, "joined errors surface from all upstream failures")

	records := decodeRecords(t, buf)
	tier := recordMsgsWithPrefix(records, "fallback.tier.")
	require.NotEmpty(t, tier)
	assert.Equal(t, "fallback.tier.exhausted", tier[len(tier)-1],
		"last tier emission must be the exhausted record; got %v", tier)

	ex := recordsByMsg(records, "fallback.tier.exhausted")
	require.Len(t, ex, 1)
	assert.Equal(t, "fallback.tier", ex[0]["op"])
	assert.EqualValues(t, len(order), ex[0]["tiers_tried"])

	// §14: no err.Error() text should leak in any record.
	for _, r := range records {
		for k, v := range r {
			if s, ok := v.(string); ok {
				assert.NotEqual(t, "transport: : " /*joined error fragment*/, s,
					"record attr %q leaked joined err.Error() text", k)
			}
		}
	}
}
