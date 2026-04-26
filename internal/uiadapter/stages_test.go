package uiadapter

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStageKind_ParseRoundtrip — AC-5.2 precondition. Classifier output
// parses to the enum; unknown values return ErrUnknownKind wrapped.
func TestStageKind_ParseRoundtrip(t *testing.T) {
	t.Parallel()
	for _, k := range []StageKind{StageKindYN, StageKindMenu, StageKindForm, StageKindText} {
		got, err := ParseStageKind(string(k), nil)
		require.NoError(t, err)
		assert.Equal(t, k, got)
	}
	got, err := ParseStageKind("  YN  ", nil)
	require.NoError(t, err)
	assert.Equal(t, StageKindYN, got, "trim + case-insensitive")

	_, err = ParseStageKind("not-a-kind", nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownKind)
}

// TestAssembleStage1_ByteStable — AC-5.4 precondition + Story v3-07 prefix
// cache precondition. Two assembly calls with the same raw produce the
// same byte output (deterministic, no timestamps or interpolation).
func TestAssembleStage1_ByteStable(t *testing.T) {
	t.Parallel()
	raw := "What colour?"
	a := AssembleStage1(raw, nil)
	b := AssembleStage1(raw, nil)
	assert.Equal(t, a, b)
	assert.True(t, strings.HasPrefix(a, "# Role"),
		"classify prompt header is the first thing the model sees")
	assert.Contains(t, a, "RAW CAPTURE:")
	assert.Contains(t, a, raw)
}

// TestAssembleStage2_ContainsKindDirective — stage-2 terminates with the
// chosen kind so the per-kind prompt + schema align.
func TestAssembleStage2_ContainsKindDirective(t *testing.T) {
	t.Parallel()
	out, err := AssembleStage2(StageKindMenu, "Pick one:\n1) a\n2) b", nil)
	require.NoError(t, err)
	assert.Contains(t, out, "KIND: menu")
	assert.Contains(t, out, "response_key")
	assert.Contains(t, out, "Pick one:")
}

// TestAssembleStage2_UnknownKindErrors — passing an invalid kind surfaces
// ErrUnknownKind so downstream code can fall back.
func TestAssembleStage2_UnknownKindErrors(t *testing.T) {
	t.Parallel()
	_, err := AssembleStage2("invalid", "anything", nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownKind)
}

// TestPromptsBudget — Plan §6.3 byte budgets. Each per-kind prompt ≤ 3 KiB.
func TestPromptsBudget(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"classify": stageClassifyPrompt,
		"yn":       stageGenerateYN,
		"menu":     stageGenerateMenu,
		"form":     stageGenerateForm,
		"text":     stageGenerateText,
	}
	for name, body := range cases {
		assert.LessOrEqual(t, len(body), 3*1024, "prompt %q exceeds §6.3 3 KiB budget", name)
	}
}

// TestRunTwoStage_OllamaPolicyTwoCalls — AC-5.5 Ollama path = 2 network
// calls (one classify + one generate). The router Story v3-16 turns this
// into a Backend.Classify + Backend.Generate dispatch.
func TestRunTwoStage_OllamaPolicyTwoCalls(t *testing.T) {
	t.Parallel()
	var classifyCalls, generateCalls int
	classify := func(_ context.Context, _ string) (StageKind, error) {
		classifyCalls++
		return StageKindMenu, nil
	}
	generate := func(_ context.Context, kind StageKind, prompt string) (*UIAST, error) {
		generateCalls++
		assert.Equal(t, StageKindMenu, kind)
		assert.Contains(t, prompt, "KIND: menu")
		return &UIAST{Version: "1", GeneratedBy: "stub"}, nil
	}
	ast, kind, err := RunTwoStage(context.Background(), "raw", classify, generate, nil)
	require.NoError(t, err)
	require.NotNil(t, ast)
	assert.Equal(t, StageKindMenu, kind)
	assert.Equal(t, 1, classifyCalls)
	assert.Equal(t, 1, generateCalls)
}

// TestRunTwoStage_ClassifyErrorAborts — classifier failure must not call
// generate. The router can then fall back to a different backend.
func TestRunTwoStage_ClassifyErrorAborts(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("classify-down")
	var generateCalls int
	_, _, err := RunTwoStage(context.Background(), "raw",
		func(context.Context, string) (StageKind, error) { return "", sentinel },
		func(context.Context, StageKind, string) (*UIAST, error) { generateCalls++; return nil, nil },
		nil)
	require.Error(t, err)
	assert.ErrorIs(t, err, sentinel)
	assert.Zero(t, generateCalls, "stage-2 must not fire when stage-1 errored")
}

// TestRunTwoStage_ContextCancelled — canceled context short-circuits the
// pipeline before any network call.
func TestRunTwoStage_ContextCancelled(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := RunTwoStage(ctx, "raw",
		func(context.Context, string) (StageKind, error) { return StageKindText, nil },
		func(context.Context, StageKind, string) (*UIAST, error) { return &UIAST{}, nil },
		nil)
	require.ErrorIs(t, err, context.Canceled)
}

// --- Story 4: uiadapter-logging-4-instrument-pipeline -----------------------
//
// AC-4.4: stages.go emits two-stage transitions
// AC-4.5: stages.go emits parse-error path
//
// Required emissions per Story 4 dev notes:
//   - stages.start              (op=stages.two, bytes_in)
//   - stages.classify.done      (op, kind, latency_ms)
//   - stages.generate.start     (op, kind)
//   - stages.final              (op, kind, latency_ms_total, bytes_out)
//   - stages.classify.parse_error (op, reason="parse")
//   - stages.assemble           (op, stage=1|2, bytes_out)
//   - stages.kind.invalid       (op, input_len)   — never the rejected string
//
// These RED-phase tests will fail to compile until ParseStageKind, AssembleStage1,
// and AssembleStage2 grow a *slog.Logger parameter, and until the parse-error
// branch in RunTwoStage detects classifyFn errors that wrap ErrUnknownKind.

// TestStory4_AC4_TwoStageHappyPath — Story 4, AC-4.4.
//
// Stub stage 1 returns kind=widget; stage 2 returns a UIAST. Expected
// emission sequence:
//
//	stages.start
//	stages.classify.done   (kind=widget, latency_ms ≥ 0)
//	stages.generate.start  (kind=widget)
//	stages.final           (kind=widget, latency_ms_total ≥ classify.latency_ms, bytes_out > 0)
//
// "widget" is not a valid StageKind in the current enum; the test uses
// StageKindForm in its place per the actual enum (form/menu/text/yn). The
// brief's "widget" is a placeholder for "the classified kind". Every
// kind-bearing record must carry the same kind value.
func TestStory4_AC4_TwoStageHappyPath(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	const raw = "Pick a colour"
	const wantKind = StageKindForm
	classify := func(_ context.Context, _ string) (StageKind, error) { return wantKind, nil }
	generate := func(_ context.Context, kind StageKind, _ string) (*UIAST, error) {
		assert.Equal(t, wantKind, kind)
		return &UIAST{Version: "1", GeneratedBy: "stub"}, nil
	}

	ast, kind, err := RunTwoStage(context.Background(), raw, classify, generate, logger)
	require.NoError(t, err)
	require.NotNil(t, ast)
	assert.Equal(t, wantKind, kind)

	records := decodeRecords(t, buf)
	wantOrder := []string{
		"stages.start",
		"stages.classify.done",
		"stages.generate.start",
		"stages.final",
	}
	require.Equal(t, wantOrder, recordMsgsWithPrefix(records, "stages."),
		"emissions must follow the happy-path script; got %v", records)

	start := recordsByMsg(records, "stages.start")[0]
	assert.Equal(t, "stages.two", start["op"])
	assert.EqualValues(t, len(raw), start["bytes_in"])

	classifyDone := recordsByMsg(records, "stages.classify.done")[0]
	assert.Equal(t, "stages.two", classifyDone["op"])
	assert.Equal(t, string(wantKind), classifyDone["kind"])
	classifyLat, ok := classifyDone["latency_ms"].(float64)
	require.True(t, ok, "classify.done must carry numeric latency_ms; got %T", classifyDone["latency_ms"])
	assert.GreaterOrEqual(t, classifyLat, 0.0)

	genStart := recordsByMsg(records, "stages.generate.start")[0]
	assert.Equal(t, "stages.two", genStart["op"])
	assert.Equal(t, string(wantKind), genStart["kind"])

	final := recordsByMsg(records, "stages.final")[0]
	assert.Equal(t, "stages.two", final["op"])
	assert.Equal(t, string(wantKind), final["kind"])
	totalLat, ok := final["latency_ms_total"].(float64)
	require.True(t, ok, "final must carry numeric latency_ms_total")
	assert.GreaterOrEqual(t, totalLat, classifyLat,
		"final latency_ms_total (%v) must include the stage-1 latency (%v)", totalLat, classifyLat)
	bytesOut, ok := final["bytes_out"].(float64)
	require.True(t, ok, "final must carry numeric bytes_out")
	assert.Greater(t, bytesOut, 0.0)
}

// TestStory4_AC5_StagesParseError — Story 4, AC-4.5.
//
// Stage 1 returns an error wrapping ErrUnknownKind to model a malformed-JSON /
// unparseable classifier response. The instrumented RunTwoStage must emit
// `stages.classify.parse_error` with `reason="parse"` and MUST NOT proceed to
// stage 2 (no `stages.generate.start` record).
func TestStory4_AC5_StagesParseError(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	parseErr := fmt.Errorf("%w: malformed json", ErrUnknownKind)
	classify := func(context.Context, string) (StageKind, error) { return "", parseErr }
	var generateCalls int
	generate := func(context.Context, StageKind, string) (*UIAST, error) {
		generateCalls++
		return nil, nil
	}

	_, _, err := RunTwoStage(context.Background(), "raw", classify, generate, logger)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownKind)
	assert.Zero(t, generateCalls, "stage 2 must not run after parse error")

	records := decodeRecords(t, buf)
	parseRecs := recordsByMsg(records, "stages.classify.parse_error")
	require.Len(t, parseRecs, 1, "exactly one parse_error record; got %v", records)
	assert.Equal(t, "stages.two", parseRecs[0]["op"])
	assert.Equal(t, "parse", parseRecs[0]["reason"])

	assert.Empty(t, recordsByMsg(records, "stages.generate.start"),
		"generate.start must not emit on parse failure")
	assert.Empty(t, recordsByMsg(records, "stages.final"),
		"final must not emit on parse failure")
}

// TestStory4_AC4_AssembleEmits — Story 4, AC-4.4 supporting site.
//
// AssembleStage1 and AssembleStage2 each emit a single `stages.assemble`
// record carrying the corresponding `stage` index and the produced prompt's
// length as `bytes_out`. The raw / kind contents themselves are NOT logged.
func TestStory4_AC4_AssembleEmits(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	stage1 := AssembleStage1("RAW", logger)
	stage2, err := AssembleStage2(StageKindMenu, "RAW", logger)
	require.NoError(t, err)

	records := decodeRecords(t, buf)
	asm := recordsByMsg(records, "stages.assemble")
	require.Len(t, asm, 2, "one assemble record per stage; got %v", records)

	byStage := map[float64]map[string]any{}
	for _, r := range asm {
		s, ok := r["stage"].(float64)
		require.True(t, ok, "stage attr must be numeric; got %T", r["stage"])
		byStage[s] = r
	}

	r1, ok := byStage[1]
	require.True(t, ok, "missing stage=1 record")
	assert.Equal(t, "stages.two", r1["op"])
	assert.EqualValues(t, len(stage1), r1["bytes_out"])

	r2, ok := byStage[2]
	require.True(t, ok, "missing stage=2 record")
	assert.Equal(t, "stages.two", r2["op"])
	assert.EqualValues(t, len(stage2), r2["bytes_out"])
}

// TestStory4_AC4_ParseStageKindRejected — Story 4, supporting site for the
// `stages.kind.invalid` emission. Per §14 sanitize discipline, the rejected
// string is NEVER logged verbatim; only its length is captured.
func TestStory4_AC4_ParseStageKindRejected(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	const rejected = "not-a-kind"
	_, err := ParseStageKind(rejected, logger)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownKind)

	records := decodeRecords(t, buf)
	rec := recordsByMsg(records, "stages.kind.invalid")
	require.Len(t, rec, 1, "exactly one stages.kind.invalid record; got %v", records)
	assert.Equal(t, "stages.two", rec[0]["op"])
	assert.EqualValues(t, len(rejected), rec[0]["input_len"],
		"input_len must equal len(rejected)=%d", len(rejected))

	// CRITICAL §14: no record may contain the rejected string verbatim.
	for _, r := range records {
		for k, v := range r {
			if s, ok := v.(string); ok {
				assert.NotContains(t, s, rejected,
					"record attr %q leaked the rejected string verbatim", k)
			}
		}
	}
}
