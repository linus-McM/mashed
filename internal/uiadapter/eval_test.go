//go:build ollama_eval

// Story ui-ast-U9: offline eval harness.
//
// Gated behind the `ollama_eval` build tag (AC-5) — `go test ./...` without
// the tag does not compile this file. Lives in package uiadapter_test (an
// external test package) so it can import mashed/internal/uiadapter/eval
// without the cycle that an in-package file would produce (the eval package
// itself depends on uiadapter for UIAST types).
package uiadapter_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"

	"mashed/internal/uiadapter"
	"mashed/internal/uiadapter/eval"
)

// TestEval_SkipWhenUnreachable — AC-6: call t.Skip when the adapter's first
// Translate returns fallback:unreachable, rather than failing the build.
// 127.0.0.1:1 is a reserved port guaranteed not to accept connections, so
// the assertion is deterministic regardless of whether the host machine
// happens to be running Ollama.
func TestEval_SkipWhenUnreachable(t *testing.T) {
	uiadapter.WithOllamaHost(t, "http://127.0.0.1:1")

	adapter := uiadapter.NewDefault(uiadapter.Config{
		Enabled:   true,
		Model:     "gemma3:4b",
		TimeoutMs: 200,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	ast := adapter.Translate(context.Background(), "synthetic raw turn", "proc")
	require.NotNil(t, ast)

	if ast.GeneratedBy == "fallback:unreachable" {
		t.Skip("ollama unreachable — install and pull gemma3:4b to run eval")
	}

	t.Fatalf("TEST PRECONDITION: expected fallback:unreachable from 127.0.0.1:1; got GeneratedBy=%q",
		ast.GeneratedBy)
}

// TestEval_FullCorpus_MeetsThresholds — AC-5/AC-6: runs every fixture through
// the default adapter against a real local Ollama. Skips cleanly when Ollama
// is unreachable so the build stays green on machines without it.
func TestEval_FullCorpus_MeetsThresholds(t *testing.T) {
	corpus := eval.LoadCorpus()
	require.NotEmpty(t, corpus, "LoadCorpus must return the committed testdata/eval fixtures")

	adapter := uiadapter.NewDefault(uiadapter.Config{
		Enabled:       true,
		Model:         "gemma3:4b",
		TimeoutMs:     5000,
		Deterministic: true,
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Probe before scoring the full corpus so an absent Ollama triggers a
	// single fast skip rather than N timed-out translates.
	probe := adapter.Translate(context.Background(), "probe", "probe")
	if probe != nil && probe.GeneratedBy == "fallback:unreachable" {
		t.Skip("ollama unreachable — install and pull gemma3:4b to run eval")
	}

	card := eval.Score(t, adapter, corpus)
	require.NotNil(t, card)

	t.Logf("scorecard:\n%s", card.PrettyPrint())
	if err := card.MeetsThresholds(); err != nil {
		t.Fatal(err)
	}
}
