package uiadapter

import (
	"context"
	"errors"
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
		got, err := ParseStageKind(string(k))
		require.NoError(t, err)
		assert.Equal(t, k, got)
	}
	got, err := ParseStageKind("  YN  ")
	require.NoError(t, err)
	assert.Equal(t, StageKindYN, got, "trim + case-insensitive")

	_, err = ParseStageKind("not-a-kind")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownKind)
}

// TestAssembleStage1_ByteStable — AC-5.4 precondition + Story v3-07 prefix
// cache precondition. Two assembly calls with the same raw produce the
// same byte output (deterministic, no timestamps or interpolation).
func TestAssembleStage1_ByteStable(t *testing.T) {
	t.Parallel()
	raw := "What colour?"
	a := AssembleStage1(raw)
	b := AssembleStage1(raw)
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
	out, err := AssembleStage2(StageKindMenu, "Pick one:\n1) a\n2) b")
	require.NoError(t, err)
	assert.Contains(t, out, "KIND: menu")
	assert.Contains(t, out, "response_key")
	assert.Contains(t, out, "Pick one:")
}

// TestAssembleStage2_UnknownKindErrors — passing an invalid kind surfaces
// ErrUnknownKind so downstream code can fall back.
func TestAssembleStage2_UnknownKindErrors(t *testing.T) {
	t.Parallel()
	_, err := AssembleStage2("invalid", "anything")
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
