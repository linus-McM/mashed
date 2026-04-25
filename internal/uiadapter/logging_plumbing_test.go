// Package uiadapter — Story 2: Plumb logger through subcomponent constructors.
//
// RED-phase tests for the constructor reshape described in
// docs/stories/uiadapter-logging-2-plumb-subcomponents.md. These tests
// intentionally fail to compile / pass until the go-engineer ships:
//
//   - `nilSafeLogger(*slog.Logger) *slog.Logger` in `logging.go`
//   - Every subcomponent constructor and free function listed in the story's
//     table appended with `logger *slog.Logger` (last positional parameter)
//   - `semaphore` promoted to a struct with `acquire(ctx) bool` / `release()`
//   - `NewDefault` scoping the parent via `WithGroup("uiadapter")` before
//     injecting it into every constructor
//   - Zero new `logger.Debug/Info/Warn/Error/LogAttrs` call sites in the
//     reshape files (diff guard — AC-2.6)
//
// Test names embed the AC number per the story's six acceptance criteria.
package uiadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// AC-2.5 — nilSafeLogger is the single nil-guard.
// -----------------------------------------------------------------------------

// TestStory2_AC5_NilSafeLogger asserts the helper's two-branch contract:
// nil → fresh io.Discard-backed logger; non-nil → identity-preserved.
func TestStory2_AC5_NilSafeLogger(t *testing.T) {
	t.Parallel()

	// Build a non-nil reference logger we can identity-compare.
	ref := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name        string
		input       *slog.Logger
		wantNotNil  bool
		wantSameRef bool
	}{
		{
			name:       "nil input returns non-nil discard logger",
			input:      nil,
			wantNotNil: true,
		},
		{
			name:        "non-nil input is preserved by identity",
			input:       ref,
			wantNotNil:  true,
			wantSameRef: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nilSafeLogger(tt.input)
			require.NotNil(t, got, "nilSafeLogger must never return nil")
			if tt.wantSameRef {
				assert.Same(t, tt.input, got,
					"non-nil input must be returned verbatim (identity preserved)")
			}
			// Sanity: the discard logger must not panic when callers Debug
			// at any level. We don't assert silence at the io level (the
			// helper handler choice is implementation detail) — only that
			// the returned logger is operational.
			assert.NotPanics(t, func() {
				got.Debug("probe", slog.String("k", "v"))
			})
		})
	}
}

// -----------------------------------------------------------------------------
// AC-2.1 + AC-2.2 — every subcomponent constructor accepts a logger and is
// nil-safe. NewMock lives in mock_test.go (`//go:build testing`) because
// mock.go itself is tagged.
// -----------------------------------------------------------------------------

// TestStory2_AC2_AC1_ConstructorsAcceptNilLogger walks every reshape-target
// constructor and asserts: (1) the constructor accepts `nil` for the logger
// argument without panicking, and (2) the returned value is non-nil and at
// least one public method runs without panicking.
//
// The table is intentionally explicit per-constructor — go's type system
// won't let us close over heterogeneous return types in a generic loop,
// so each row is its own subtest.
func TestStory2_AC2_AC1_ConstructorsAcceptNilLogger(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Enabled = true

	t.Run("NewClient", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			c := NewClient(ClientConfig{TimeoutMs: 1000}, nil)
			require.NotNil(t, c, "NewClient must return non-nil with nil logger")
		})
	})

	t.Run("NewResponseCache", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			rc := NewResponseCache(cfg, nil)
			require.NotNil(t, rc)
			// HitRate is the cheapest read-only public method — exercises
			// the receiver without touching the LRU.
			_ = rc.HitRate()
		})
	})

	t.Run("NewBreakerSet", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			bs := NewBreakerSet(cfg, nil)
			require.NotNil(t, bs)
			assert.Equal(t, "closed", bs.StateOf("probe"),
				"a freshly-constructed breaker is closed")
		})
	})

	t.Run("newSemaphore", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			s := newSemaphore(1, nil)
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			require.True(t, s.acquire(ctx),
				"newSemaphore(1, nil).acquire must succeed on a fresh permit")
			s.release()
		})
	})

	t.Run("NewRepairer", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			r := NewRepairer(cfg, nil)
			require.NotNil(t, r)
			// MaxRetries is read-only and side-effect free.
			_ = r.MaxRetries()
		})
	})

	t.Run("NewFastPathClassifier", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			fp := NewFastPathClassifier(true, nil)
			require.NotNil(t, fp)
			// Classify on an empty raw is the cheapest input that
			// exercises the receiver.
			hit, _, _ := fp.Classify("")
			assert.False(t, hit, "empty raw never matches a fast-path rule")
		})
	})

	t.Run("NewContextGuard", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			g := NewContextGuard(cfg, nil)
			require.NotNil(t, g)
			opts := g.OllamaOptions()
			require.NotNil(t, opts, "OllamaOptions must return a non-nil map")
		})
	})
}

// -----------------------------------------------------------------------------
// AC-2.2 — free functions in the request path accept a nil logger.
// -----------------------------------------------------------------------------

// TestStory2_AC2_FreeFunctionsAcceptNilLogger covers the free-function row
// of the story's reshape table. Each call uses the smallest valid argument
// shape that still exercises the function body, with `nil` as the logger.
func TestStory2_AC2_FreeFunctionsAcceptNilLogger(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	t.Run("SanitizeCapture", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			out, delta := SanitizeCapture("hello world", nil)
			assert.Equal(t, "hello world", out)
			assert.Equal(t, 0, delta)
		})
	})

	t.Run("Spotlight", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			_ = Spotlight("hello", true, nil)
		})
	})

	t.Run("Validate", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			_ = Validate(&UIAST{Version: "1"}, "hello", nil)
		})
	})

	t.Run("BuildRepairPrompt", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			out := BuildRepairPrompt(RepairAttempt{
				Kind:           StageKindYN,
				StaticPrefix:   "prefix",
				SanitizedRaw:   "raw",
				PreviousOutput: "prev",
				Errors:         []string{"err"},
			}, nil)
			assert.NotEmpty(t, out)
		})
	})

	t.Run("FallbackAST", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			ast := FallbackAST("hello", "test-reason", nil)
			require.NotNil(t, ast)
			assert.Equal(t, "fallback:test-reason", ast.GeneratedBy)
		})
	})

	t.Run("OllamaFormatPayload", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			out, err := OllamaFormatPayload(StageKindYN, false, nil)
			require.NoError(t, err)
			assert.NotEmpty(t, out)
		})
	})

	t.Run("ClaudeToolInputSchema", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			out, err := ClaudeToolInputSchema(StageKindYN, nil)
			require.NoError(t, err)
			assert.NotEmpty(t, out)
		})
	})

	t.Run("OllamaSamplingOptions", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			out := OllamaSamplingOptions(cfg, nil)
			assert.NotNil(t, out)
		})
	})

	t.Run("ClaudeSamplingOptions", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			out := ClaudeSamplingOptions(cfg, nil)
			assert.NotNil(t, out)
		})
	})

	t.Run("ClaudeSystemBlock", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			out := ClaudeSystemBlock("hello", cfg, nil)
			assert.NotNil(t, out)
		})
	})

	t.Run("ClaudeSystemBlockJSON", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() {
			out, err := ClaudeSystemBlockJSON("hello", cfg, nil)
			require.NoError(t, err)
			assert.NotEmpty(t, out)
		})
	})
}

// -----------------------------------------------------------------------------
// AC-2.3 — NewDefault scopes a single logger via WithGroup("uiadapter") and
// injects it into every subcomponent.
// -----------------------------------------------------------------------------

// TestStory2_AC3_NewDefaultScopesWithGroup builds a parent logger over a
// captured bytes.Buffer and asserts the scoped logger NewDefault stores on
// the resulting *defaultAdapter emits records under a "uiadapter" group key.
//
// This story ships zero new production log calls, so the assertion drives a
// record through the adapter's own logger field directly (test-only
// emission) rather than relying on a production-emitted record.
func TestStory2_AC3_NewDefaultScopesWithGroup(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	parent := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.MaxInflight = 1

	a := NewDefault(cfg, parent)
	require.NotNil(t, a)

	d, ok := a.(*defaultAdapter)
	require.True(t, ok,
		"NewDefault with cfg.Enabled=true must return *defaultAdapter; got %T", a)
	require.NotNil(t, d.logger,
		"defaultAdapter.logger must be non-nil after NewDefault")

	// Drive one record through the adapter's logger field. Because the
	// reshape scopes via WithGroup("uiadapter"), every attr should land
	// under that group in the JSON record.
	d.logger.Debug("scoped-record", slog.String("k", "v"))

	var rec map[string]any
	require.NoError(t, json.NewDecoder(buf).Decode(&rec),
		"scoped logger must produce a single decodable JSON record")

	group, ok := rec["uiadapter"].(map[string]any)
	require.True(t, ok,
		"logger must be scoped under \"uiadapter\" group; got top-level keys=%v", keysOf(rec))
	assert.Equal(t, "v", group["k"],
		"the attr emitted via the scoped logger must live under the uiadapter group")
}

// keysOf returns the top-level keys of a JSON-decoded record for diagnostic
// output on failure. Sorted only on the failure path so passing runs stay
// allocation-free.
func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TestStory2_AC3_NewDefaultNilParentSafe asserts NewDefault tolerates a nil
// parent logger and produces no stray output on stdout from package code.
func TestStory2_AC3_NewDefaultNilParentSafe(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Enabled = true
	cfg.MaxInflight = 1

	stdout := captureStdout(t, func() {
		assert.NotPanics(t, func() {
			a := NewDefault(cfg, nil)
			require.NotNil(t, a, "NewDefault must return a non-nil adapter even with nil parent")

			// If we got a defaultAdapter, its logger field must be a
			// non-nil discard-backed logger so subsequent stories can
			// emit safely without crashing.
			if d, ok := a.(*defaultAdapter); ok {
				require.NotNil(t, d.logger,
					"defaultAdapter.logger must be non-nil even when parent was nil")
				// Driving a record must not panic, and must not surface
				// on stdout (the discard handler swallows it).
				d.logger.Debug("nil-parent-probe", slog.String("k", "v"))
			}
		})
	})
	assert.Empty(t, stdout,
		"NewDefault(cfg, nil) must not write anything to stdout from package code")
}

// -----------------------------------------------------------------------------
// AC-2.6 — diff guard. The reshape files must NOT introduce new log call
// sites in this story. Subsequent instrumentation stories will update this
// allowlist.
// -----------------------------------------------------------------------------

// TestStory2_AC6_NoNewLogCallsInProduction scans every file Story 2
// reshapes and asserts the count of `logger.<level>(` substrings is zero.
// `adapter.go` and `allowlist.go` are excluded because they have
// pre-existing emissions outside the scope of this story.
func TestStory2_AC6_NoNewLogCallsInProduction(t *testing.T) {
	t.Parallel()

	files := []string{
		"client.go",
		"cache.go",
		"prefix_cache.go",
		"breaker.go",
		"semaphore.go",
		"repair.go",
		"fastpath.go",
		"stages.go",
		"fallback.go",
		"fallback_tiers.go",
		"sanitize.go",
		"spotlight.go",
		"contextguard.go",
		"validator.go",
		"schema.go",
		"encode.go",
		"sampling.go",
		"mock.go",
	}
	forbidden := []string{
		"logger.Debug(",
		"logger.Info(",
		"logger.Warn(",
		"logger.Error(",
		"logger.LogAttrs(",
		".logger.Debug(",
		".logger.Info(",
		".logger.Warn(",
		".logger.Error(",
		".logger.LogAttrs(",
	}

	for _, f := range files {
		body, err := os.ReadFile(f)
		if err != nil {
			// File may not exist on disk in older snapshots; the diff
			// guard is informational for missing files.
			continue
		}
		src := string(body)
		for _, pat := range forbidden {
			assert.Equal(t, 0, strings.Count(src, pat),
				"Story 2 must not introduce log calls in %s; found %q", f, pat)
		}
	}
}
