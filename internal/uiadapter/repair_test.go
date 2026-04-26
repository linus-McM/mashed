package uiadapter

import (
	"context"
	"errors"
	"log/slog"
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

// --- Story 4: uiadapter-logging-4-instrument-pipeline -----------------------
//
// AC-4.2: repair.go emits attempt lifecycle (exhausted path)
// AC-4.3: repair.go emits success on first try
//
// Required emissions per Story 4 dev notes:
//   - repair.start            (op=repair.run, max_retries, bytes_in)
//   - repair.attempt          (op, attempt=N, reason=<previous validator reason>)
//   - repair.success          (op, attempt=N, latency_ms)
//   - repair.failure          (op, attempt=N, reason=<new validator reason>)
//   - repair.exhausted        (op, attempts_total, final_reason)
//   - repair.prompt.build     (op=repair.prompt, bytes_out)
//
// The validator reasons logged are short closed-enum strings ("oversize",
// "marshal_error", "required_dropped", …) per §14 sanitize discipline.

// TestStory4_AC2_RepairExhausted — Story 4, AC-4.2 (exhausted path).
//
// RepairMaxRetries=2 with a stub that always fails validation with reason
// "oversize". Expected emission sequence:
//
//	repair.start
//	repair.attempt   (n=1, reason=oversize)
//	repair.failure   (n=1, reason=oversize)
//	repair.attempt   (n=2, reason=oversize)
//	repair.failure   (n=2, reason=oversize)
//	repair.exhausted (attempts_total=2, final_reason=oversize)
//
// No repair.success record may appear.
func TestStory4_AC2_RepairExhausted(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	cfg := DefaultConfig()
	cfg.RepairMaxRetries = 2
	r := NewRepairer(cfg, logger)

	bad := &UIAST{Version: ""}
	validate := func(*UIAST) []string { return []string{"oversize"} }
	generate := func(_ context.Context, _ string) (*UIAST, string, error) {
		return bad, "{}", nil
	}

	_, err := r.Run(context.Background(), bad, "{}", RepairAttempt{}, validate, generate)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrRepairExhausted)

	records := decodeRecords(t, buf)
	msgs := recordMsgsWithPrefix(records, "repair.")

	wantOrder := []string{
		"repair.start",
		"repair.attempt",
		"repair.failure",
		"repair.attempt",
		"repair.failure",
		"repair.exhausted",
	}
	require.Equal(t, wantOrder, msgs, "emissions must follow the exhausted-path script")
	assert.Empty(t, recordsByMsg(records, "repair.success"))

	start := recordsByMsg(records, "repair.start")[0]
	assert.Equal(t, "repair.run", start["op"])
	assert.EqualValues(t, 2, start["max_retries"])
	assert.EqualValues(t, len("{}"), start["bytes_in"])

	attempts := recordsByMsg(records, "repair.attempt")
	require.Len(t, attempts, 2)
	for i, a := range attempts {
		assert.Equal(t, "repair.run", a["op"])
		assert.EqualValues(t, i+1, a["attempt"], "attempt counter must be 1-indexed")
		assert.Equal(t, "oversize", a["reason"], "reason must be the previous validator enum")
	}

	failures := recordsByMsg(records, "repair.failure")
	require.Len(t, failures, 2)
	for i, f := range failures {
		assert.EqualValues(t, i+1, f["attempt"])
		assert.Equal(t, "oversize", f["reason"])
	}

	ex := recordsByMsg(records, "repair.exhausted")[0]
	assert.Equal(t, "repair.run", ex["op"])
	assert.EqualValues(t, 2, ex["attempts_total"])
	assert.Equal(t, "oversize", ex["final_reason"])
}

// TestStory4_AC3_RepairSuccessOnFirstTry — Story 4, AC-4.3.
//
// RepairMaxRetries=2 with a stub that succeeds on the first call. Expected
// sequence: repair.start, repair.attempt (n=1), repair.success (n=1, latency_ms ≥ 0).
// No repair.failure or repair.exhausted may appear.
func TestStory4_AC3_RepairSuccessOnFirstTry(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	cfg := DefaultConfig()
	cfg.RepairMaxRetries = 2
	r := NewRepairer(cfg, logger)

	bad := &UIAST{Version: ""}
	good := &UIAST{Version: "1"}
	validate := func(ast *UIAST) []string {
		if ast.Version == "" {
			return []string{"oversize"}
		}
		return nil
	}
	generate := func(_ context.Context, _ string) (*UIAST, string, error) {
		return good, `{"version":"1"}`, nil
	}

	got, err := r.Run(context.Background(), bad, "{}", RepairAttempt{}, validate, generate)
	require.NoError(t, err)
	assert.Same(t, good, got)

	records := decodeRecords(t, buf)
	msgs := recordMsgsWithPrefix(records, "repair.")
	wantOrder := []string{"repair.start", "repair.attempt", "repair.success"}
	require.Equal(t, wantOrder, msgs, "first-try success script")

	assert.Empty(t, recordsByMsg(records, "repair.failure"))
	assert.Empty(t, recordsByMsg(records, "repair.exhausted"))

	success := recordsByMsg(records, "repair.success")[0]
	assert.Equal(t, "repair.run", success["op"])
	assert.EqualValues(t, 1, success["attempt"])
	latencyMs, ok := success["latency_ms"].(float64)
	require.True(t, ok, "success record must carry numeric latency_ms")
	assert.GreaterOrEqual(t, latencyMs, 0.0)
}

// TestStory4_AC2_BuildRepairPromptEmits — Story 4, AC-4.2 supporting site.
//
// BuildRepairPrompt emits a single `repair.prompt.build` record carrying the
// produced prompt's length as `bytes_out`. The prompt content itself is NEVER
// logged (sanitize discipline §14).
func TestStory4_AC2_BuildRepairPromptEmits(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	prompt := BuildRepairPrompt(RepairAttempt{
		Kind:           StageKindMenu,
		StaticPrefix:   "PREFIX",
		SanitizedRaw:   "RAW",
		PreviousOutput: "BAD",
		Errors:         []string{"oversize"},
	}, logger)
	require.True(t, strings.HasPrefix(prompt, "PREFIX"))

	records := decodeRecords(t, buf)
	emits := recordsByMsg(records, "repair.prompt.build")
	require.Len(t, emits, 1, "expected exactly one repair.prompt.build record; got %v", records)

	rec := emits[0]
	assert.Equal(t, "repair.prompt", rec["op"])
	bytesOut, ok := rec["bytes_out"].(float64)
	require.True(t, ok, "bytes_out must be numeric")
	assert.EqualValues(t, len(prompt), bytesOut)

	// Sanitize discipline: no record may contain the prompt body verbatim.
	for _, r := range records {
		for k, v := range r {
			if s, ok := v.(string); ok {
				assert.NotContains(t, s, prompt,
					"record attr %q leaked the full prompt body", k)
			}
		}
	}
}
