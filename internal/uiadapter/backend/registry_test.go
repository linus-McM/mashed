package backend_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mashed/internal/uiadapter"
	"mashed/internal/uiadapter/backend"
	_ "mashed/internal/uiadapter/backend/claudeapi"
	_ "mashed/internal/uiadapter/backend/claudecli"
	_ "mashed/internal/uiadapter/backend/ollama"
)

// TestBackend_FromUnknownName — Story C AC-C.2. Unknown backend name
// returns an error wrapping ErrUnknownBackend whose message lists the
// sorted available set.
func TestBackend_FromUnknownName(t *testing.T) {
	t.Parallel()
	cfg := uiadapter.DefaultConfig()
	_, err := backend.From("nonexistent", cfg)
	require.Error(t, err, "From must error on unknown backend name")
	assert.ErrorIs(t, err, backend.ErrUnknownBackend, "must wrap sentinel")
	msg := err.Error()
	for _, want := range []string{"ollama", "claude-api", "claude-cli"} {
		assert.Contains(t, msg, want,
			"error must list available backend %q so the caller can correct the name", want)
	}
}

// TestBackend_FromKnownName — positive path for AC-C.2. Each registered
// backend resolves to a non-nil LLMBackend.
func TestBackend_FromKnownName(t *testing.T) {
	t.Parallel()
	cfg := uiadapter.DefaultConfig()
	for _, name := range []string{"ollama", "claude-api", "claude-cli"} {
		b, err := backend.From(name, cfg)
		require.NoErrorf(t, err, "From(%q) unexpected error", name)
		require.NotNilf(t, b, "From(%q) returned nil backend", name)
		assert.Equal(t, name, b.Name())
		caps := b.Capabilities()
		assert.NotEmpty(t, caps.Provider, "Capabilities.Provider must be set")
	}
}

// TestBackend_InterfaceStressConcurrent — Story C AC-C.1. 100 goroutines ×
// 1000 ops each, no data race under `go test -race`, and per-backend
// counters match the total call count.
func TestBackend_InterfaceStressConcurrent(t *testing.T) {
	t.Parallel()
	const (
		goroutines = 100
		opsPerG    = 1000
	)
	cfg := uiadapter.DefaultConfig()

	ollama, err := backend.From("ollama", cfg)
	require.NoError(t, err)
	claudeAPI, err := backend.From("claude-api", cfg)
	require.NoError(t, err)
	claudeCLI, err := backend.From("claude-cli", cfg)
	require.NoError(t, err)

	backends := []backend.LLMBackend{ollama, claudeAPI, claudeCLI}
	ctx := context.Background()
	raw := "stress-test raw capture"

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			b := backends[idx%len(backends)]
			for j := 0; j < opsPerG; j++ {
				switch j % 5 {
				case 0:
					_, _ = b.Classify(ctx, raw)
				case 1:
					_, _ = b.Generate(ctx, raw, backend.KindText)
				case 2:
					_, _ = b.GenerateSingleShot(ctx, raw)
				case 3:
					_ = b.WarmUp(ctx)
				case 4:
					_ = b.Health(ctx)
				}
			}
		}(i)
	}
	wg.Wait()

	// Counter parity: across the three backends, each method sees exactly
	// (goroutines * opsPerG / 5) = 20_000 calls total.
	const expectedPerMethod = goroutines * opsPerG / 5
	var tClassify, tGen, tSingle, tWarm, tHealth int64
	for _, b := range backends {
		sb, ok := b.(*backend.StubBackend)
		require.True(t, ok, "stubs implement StubBackend for counter access")
		cl, g, ss, w, h := sb.Calls()
		tClassify += cl
		tGen += g
		tSingle += ss
		tWarm += w
		tHealth += h
	}
	assert.EqualValues(t, expectedPerMethod, tClassify, "total Classify calls across backends")
	assert.EqualValues(t, expectedPerMethod, tGen, "total Generate calls")
	assert.EqualValues(t, expectedPerMethod, tSingle, "total GenerateSingleShot calls")
	assert.EqualValues(t, expectedPerMethod, tWarm, "total WarmUp calls")
	assert.EqualValues(t, expectedPerMethod, tHealth, "total Health calls")
}

// TestBackend_SingleShotUnsupportedIsSentinel — the gemma3 path must surface
// ErrSingleShotUnsupported so the router knows to fall back to two-stage.
func TestBackend_SingleShotUnsupportedIsSentinel(t *testing.T) {
	t.Parallel()
	cfg := uiadapter.DefaultConfig()
	ollama, err := backend.From("ollama", cfg)
	require.NoError(t, err)
	_, err = ollama.GenerateSingleShot(context.Background(), "anything")
	require.Error(t, err, "Ollama stub exposes SupportsSingleShot=false")
	assert.True(t, errors.Is(err, backend.ErrSingleShotUnsupported),
		"must wrap ErrSingleShotUnsupported so router Story v3-16 can branch")
}

// TestBackend_Available — returns sorted names; used by Story v3-18 title
// bar selector to populate the pull-down.
func TestBackend_Available(t *testing.T) {
	t.Parallel()
	names := backend.Available()
	// claude-api < claude-cli < ollama lexicographic.
	assert.Equal(t, []string{"claude-api", "claude-cli", "ollama"}, names)
}
