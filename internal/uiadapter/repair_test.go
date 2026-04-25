package uiadapter

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepair_RecoveryRate — AC-10.1. Given a bad-output corpus, recovery
// rate ≥ 0.6 with one retry (Ollama config). We simulate 10 attempts where
// the second call always produces a valid AST.
func TestRepair_RecoveryRate(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.RepairMaxRetries = 1
	r := NewRepairer(cfg, nil)

	successes := 0
	for i := 0; i < 10; i++ {
		bad := &UIAST{Version: ""} // empty version == validation fail
		good := &UIAST{Version: "1"}
		validate := func(ast *UIAST) []string {
			if ast.Version == "" {
				return []string{"version missing"}
			}
			return nil
		}
		generate := func(_ context.Context, _ string) (*UIAST, string, error) {
			return good, `{"version":"1"}`, nil
		}
		result, err := r.Run(context.Background(), bad, "{}", RepairAttempt{}, validate, generate)
		if err == nil && result != nil {
			successes++
		}
	}
	assert.GreaterOrEqual(t, successes, 6, "recovery rate ≥ 0.6")

	att, succ := r.Metrics()
	assert.EqualValues(t, 10, att)
	assert.EqualValues(t, 10, succ)
}

// TestRepair_NeverExceedsBudget — AC-10.2. repair_count never exceeds
// RepairMaxRetries.
func TestRepair_NeverExceedsBudget(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.RepairMaxRetries = 1
	r := NewRepairer(cfg, nil)

	bad := &UIAST{Version: ""}
	invokes := 0
	validate := func(*UIAST) []string { return []string{"still bad"} }
	generate := func(_ context.Context, _ string) (*UIAST, string, error) {
		invokes++
		return bad, `{"version":""}`, nil
	}
	_, err := r.Run(context.Background(), bad, "{}", RepairAttempt{}, validate, generate)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRepairExhausted)
	assert.Equal(t, 1, invokes, "exactly RepairMaxRetries generate calls")
}

// TestRepair_FastPathNoWastedCalls — AC-10.3. No repair fires when the
// initial AST passes validation.
func TestRepair_FastPathNoWastedCalls(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.RepairMaxRetries = 1
	r := NewRepairer(cfg, nil)

	good := &UIAST{Version: "1"}
	invokes := 0
	validate := func(*UIAST) []string { return nil }
	generate := func(_ context.Context, _ string) (*UIAST, string, error) {
		invokes++
		return nil, "", nil
	}
	result, err := r.Run(context.Background(), good, `{"version":"1"}`, RepairAttempt{}, validate, generate)
	require.NoError(t, err)
	assert.Same(t, good, result, "pointer passthrough when no repair needed")
	assert.Zero(t, invokes)
}

// TestRepair_Gated_PerBackend — AC-10.4. RepairMaxRetries=0 (Sonnet path)
// triggers no repair; caller falls back.
func TestRepair_Gated_PerBackend(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.RepairMaxRetries = 0
	r := NewRepairer(cfg, nil)

	bad := &UIAST{Version: ""}
	validate := func(*UIAST) []string { return []string{"still bad"} }
	generate := func(_ context.Context, _ string) (*UIAST, string, error) {
		t.Fatal("generate must not be called when RepairMaxRetries=0")
		return nil, "", nil
	}
	_, err := r.Run(context.Background(), bad, "{}", RepairAttempt{}, validate, generate)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRepairExhausted)
}

// TestBuildRepairPrompt_Shape — the repair prompt carries the previous
// output verbatim + each validation error on its own bullet line.
func TestBuildRepairPrompt_Shape(t *testing.T) {
	t.Parallel()
	prompt := BuildRepairPrompt(RepairAttempt{
		Kind:           StageKindMenu,
		StaticPrefix:   "PREFIX",
		SanitizedRaw:   "RAW",
		PreviousOutput: "BAD",
		Errors:         []string{"missing required field: prompt", "unknown field 'xyz'"},
	}, nil)
	assert.True(t, strings.HasPrefix(prompt, "PREFIX"))
	assert.Contains(t, prompt, "KIND: menu")
	assert.Contains(t, prompt, "PREVIOUS ATTEMPT (invalid):\nBAD")
	assert.Contains(t, prompt, "- missing required field: prompt")
	assert.Contains(t, prompt, "- unknown field 'xyz'")
}

// TestRepair_PropagatesGenerateError — downstream generate error flows up
// without counting against the repair budget.
func TestRepair_PropagatesGenerateError(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	cfg.RepairMaxRetries = 1
	r := NewRepairer(cfg, nil)

	sentinel := errors.New("network down")
	bad := &UIAST{Version: ""}
	validate := func(*UIAST) []string { return []string{"bad"} }
	generate := func(_ context.Context, _ string) (*UIAST, string, error) {
		return nil, "", sentinel
	}
	_, err := r.Run(context.Background(), bad, "{}", RepairAttempt{}, validate, generate)
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
}
