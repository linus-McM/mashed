package uiadapter

import (
	"context"
	"errors"
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

	ast, tier, from, err := RunWithFallback(context.Background(), order, fn, minimal, plaintext)
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

	ast, tier, _, err := RunWithFallback(context.Background(), order, fn, minimal, plaintext)
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

	ast, tier, from, err := RunWithFallback(context.Background(), order, fn, minimal, plaintext)
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
	ast, tier, from, err := RunWithFallback(context.Background(), []string{"ollama"}, fn, nil, nil)
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
	_, _, _, err := RunWithFallback(ctx, []string{"any"}, fn, nil, func() *UIAST { return nil })
	require.ErrorIs(t, err, context.Canceled)
}
