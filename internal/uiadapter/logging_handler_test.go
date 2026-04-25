// Supplementary coverage for logging.go internals that aren't exercised by
// the AC-numbered tests in logging_test.go. These tests cover:
//
//   - ParseLogLevel exported wrapper (Story 1 boot-wiring helper used by app.go).
//   - fanoutHandler WithAttrs / WithGroup methods (slog.Handler contract surface).
//   - fileCloser idempotency under concurrent Close().
//
// They are intentionally separate from logging_test.go (which is RED-phase
// AC evidence) so the AC suite can be re-run independently.
package uiadapter

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseLogLevel_ExportedWrapper asserts the exported wrapper used by the
// app.go boot wiring delegates to parseSlogLevel verbatim.
func TestParseLogLevel_ExportedWrapper(t *testing.T) {
	t.Parallel()
	assert.Equal(t, slog.LevelDebug, ParseLogLevel("debug"))
	assert.Equal(t, slog.LevelInfo, ParseLogLevel(""))
	assert.Equal(t, slog.LevelInfo, ParseLogLevel("nonsense"))
	assert.Equal(t, slog.LevelWarn, ParseLogLevel("WARN"))
}

// TestFanoutHandler_WithAttrs asserts attrs added via With propagate to every
// child handler. The fanout returned must be a fresh slice, not a mutation
// of the receiver.
func TestFanoutHandler_WithAttrs(t *testing.T) {
	t.Parallel()
	var bufA, bufB bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	fanout := fanoutHandler{
		slog.NewJSONHandler(&bufA, opts),
		slog.NewJSONHandler(&bufB, opts),
	}

	scoped := slog.New(fanout).With("scope", "unit")
	scoped.Info("with-attrs")

	for _, raw := range []string{bufA.String(), bufB.String()} {
		var rec map[string]any
		require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(raw)), &rec))
		assert.Equal(t, "unit", rec["scope"], "WithAttrs must propagate to every child")
		assert.Equal(t, "with-attrs", rec["msg"])
	}
}

// TestFanoutHandler_WithGroup asserts groups propagate to every child.
func TestFanoutHandler_WithGroup(t *testing.T) {
	t.Parallel()
	var bufA, bufB bytes.Buffer
	opts := &slog.HandlerOptions{Level: slog.LevelInfo}
	fanout := fanoutHandler{
		slog.NewJSONHandler(&bufA, opts),
		slog.NewJSONHandler(&bufB, opts),
	}

	grouped := slog.New(fanout).WithGroup("uiadapter")
	grouped.Info("grouped", "k", "v")

	for _, raw := range []string{bufA.String(), bufB.String()} {
		var rec map[string]any
		require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(raw)), &rec))
		nested, ok := rec["uiadapter"].(map[string]any)
		require.True(t, ok, "WithGroup must produce a nested object on every child; got %v", rec)
		assert.Equal(t, "v", nested["k"])
	}
}

// TestFanoutHandler_EnabledShortCircuit asserts Enabled returns true if any
// child is enabled and false only when every child rejects the level.
func TestFanoutHandler_EnabledShortCircuit(t *testing.T) {
	t.Parallel()
	infoOnly := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo})
	errorOnly := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError})
	fanout := fanoutHandler{infoOnly, errorOnly}

	assert.False(t, fanout.Enabled(t.Context(), slog.LevelDebug),
		"neither child enabled at Debug — fanout must report disabled")
	assert.True(t, fanout.Enabled(t.Context(), slog.LevelInfo),
		"infoOnly child enabled at Info — fanout must report enabled")
	assert.True(t, fanout.Enabled(t.Context(), slog.LevelError),
		"both children enabled at Error — fanout must report enabled")
}

// TestFileCloser_ConcurrentDoubleClose hammers Close from multiple goroutines
// to demonstrate sync.Once-backed idempotency under contention.
func TestFileCloser_ConcurrentDoubleClose(t *testing.T) {
	t.Parallel()
	tmp, err := os.CreateTemp(t.TempDir(), "fc-*.log")
	require.NoError(t, err)

	c := &fileCloser{f: tmp}
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(t, c.Close())
		}()
	}
	wg.Wait()
}

// TestNoopCloser_AlwaysNil locks in the contract that the fallback closer
// returns nil, regardless of how many times it is invoked.
func TestNoopCloser_AlwaysNil(t *testing.T) {
	t.Parallel()
	var c io.Closer = noopCloser{}
	assert.NoError(t, c.Close())
	assert.NoError(t, c.Close())
}
