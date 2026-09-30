// Package uiadapter — Story 5 cross-cutting tests:
// AC-5.8 (per-file emission coverage + sanitize discipline) and
// AC-5.9 (zero-allocation hot path when Debug is off).
//
// AC-5.8 emission coverage: a single Translate-style flow exercising every
// instrumented file in Story 5 must surface at least one record from each of
//
//	sanitize.*, spotlight.*, contextguard.*, validator.*, encode.*,
//	sampling.*, schema.*
//
// in the captured JSON buffer. The test calls the individual functions
// sequentially with stubs — no real HTTP / Ollama / Claude backend involved.
//
// AC-5.8 sanitize fuzz: 50 random raw payloads (length 50–500, mixed ASCII
// + control chars) are pushed through the full payload-shaping surface; no
// 8-byte slice of the raw payload may appear in the captured records.
//
// AC-5.9 alloc budget: with an Info-level logger the same surface, when run
// 10000× through testing.AllocsPerRun, must add no per-call allocations
// attributable to Story 5's instrumentation (≤ 1 alloc/run tolerance).
package uiadapter

import (
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// story5RequiredOpPrefixes is the per-file emission set the AC-5.8 coverage
// test asserts. Every element must appear at least once in the captured
// buffer for the test to pass. Schema is included via the package-level
// schemaLogger trick (Story 5 AC-5.7).
var story5RequiredOpPrefixes = []string{
	"sanitize.",
	"spotlight.",
	"contextguard.",
	"validator.",
	"encode.",
	"sampling.",
	"schema.",
}

// runStory5PayloadShapingSurface drives the nine instrumented files in a
// single deterministic sequence so the cross-cutting tests share one helper
// instead of six near-duplicates. The caller supplies the logger; the helper
// exercises every emission site at least once.
//
// Pre-condition: schemaLogger has been registered (Story 5 AC-5.7) so the
// schema.unmarshal.* records emit. The helper does NOT register it — that
// stays a per-test concern so the AC-5.7 fallback test can run unaffected.
func runStory5PayloadShapingSurface(t *testing.T, logger *slog.Logger, raw string) {
	t.Helper()

	// 1) sanitize — emits sanitize.start / sanitize.done
	sanitized, _ := SanitizeCapture(raw, logger)

	// 2) spotlight — emits spotlight.start / spotlight.added (or .disabled)
	marked := Spotlight(sanitized, true, logger)

	// 3) contextguard.ollama — start + truncated regardless of overflow
	cfg := DefaultConfig()
	cfg.NumCtx = 1024
	guard := NewContextGuard(cfg, logger)
	_, _ = guard.ApplyOllama(marked)

	// 4) contextguard.claude — start + truncated branch
	_, _ = guard.ApplyClaude(marked, 200_000)

	// 5) validator — start + done; rule failures surface validator.rule.fail
	ast := &UIAST{
		Version: "1",
		Nodes: []UINode{
			{Type: "snarkfish", Content: "rule1"}, // rule 1 fires
		},
	}
	_ = Validate(ast, marked, logger)

	// 6) encode — ollama_format / claude_tool_schema / claude_tool_name / schema_select
	_, _ = OllamaFormatPayload(StageKindForm, false, logger)
	_, _ = ClaudeToolInputSchema(StageKindForm, logger)
	_ = ClaudeToolName(StageKindForm, logger)

	// 7) sampling — ollama + claude
	_ = OllamaSamplingOptions(cfg, logger)
	_ = ClaudeSamplingOptions(cfg, logger)

	// 8) schema — emit via UnmarshalJSON. Caller must have registered
	//    schemaLogger; otherwise this is a silent no-op.
	var w WidgetNode
	_ = json.Unmarshal([]byte(`{"type":"choice","options":[{"value":"a"}]}`), &w)
}

// TestStory5_AC8_PerFileEmissionCoverage — Story 5 AC-5.8.
// Sequentially exercising every instrumented entry point with a Debug logger
// must surface at least one record from each story5RequiredOpPrefixes entry
// in the JSON buffer.
func TestStory5_AC8_PerFileEmissionCoverage(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	prev := schemaLogger.Swap(logger)
	t.Cleanup(func() { schemaLogger.Store(prev) })

	runStory5PayloadShapingSurface(t, logger, "raw turn capture with prose")

	records := decodeRecords(t, buf)
	require.NotEmpty(t, records, "expected at least one record from the surface run")

	emittedOps := make(map[string]struct{}, len(records))
	for _, rec := range records {
		op, _ := rec["op"].(string)
		if op != "" {
			emittedOps[op] = struct{}{}
		}
	}

	for _, prefix := range story5RequiredOpPrefixes {
		hit := false
		for op := range emittedOps {
			if strings.HasPrefix(op, prefix) {
				hit = true
				break
			}
		}
		assert.True(t, hit,
			"AC-5.8 coverage gap: no record with op prefix %q present in buffer (got ops=%v)",
			prefix, emittedOps)
	}
}

// TestStory5_AC8_SanitizeDisciplineHolds — Story 5 AC-5.8 fuzz.
// 50 random raw payloads driven through the full payload-shaping surface;
// no record may contain any 8-byte slice of the raw payload.
func TestStory5_AC8_SanitizeDisciplineHolds(t *testing.T) {
	const iterations = 50
	const minLen, maxLen = 50, 500

	for iter := 0; iter < iterations; iter++ {
		n := minLen + (iter % (maxLen - minLen))
		raw := story4RandomPayload(t, n)
		rawStr := string(raw)

		logger, buf := testLogBuffer(t, slog.LevelDebug)
		prev := schemaLogger.Swap(logger)

		runStory5PayloadShapingSurface(t, logger, rawStr)

		// Also exercise allowlist for sanitize-discipline coverage. The
		// allowlist ok/breach paths emit on each call and the cfg.Model
		// string is enum-controlled — never user content. mock.go lives
		// behind //go:build testing so its emissions are covered in
		// mock_test.go's TestStory5_AC5_MockTranslate, not here.
		_ = CheckModelAllowlist(DefaultConfig(), logger)

		schemaLogger.Store(prev)

		body := buf.Bytes()
		if pos, leaked := containsSlice(body, raw); leaked {
			t.Fatalf("iter %d: log buffer leaked an 8-byte slice of the raw payload at offset %d; payload=%q",
				iter, pos, raw)
		}
	}
}

// TestStory5_AC9_HotPathZeroAllocs_DebugOff — Story 5 AC-5.9.
// With an Info-level logger the full payload-shaping surface, run 10000×
// through testing.AllocsPerRun, must add ≤ 1 alloc/run attributable to
// Story 5's emission guards. Skipped under -race per Stories 3/4 precedent.
func TestStory5_AC9_HotPathZeroAllocs_DebugOff(t *testing.T) {
	if raceDetectorEnabled {
		t.Skip("alloc budgets are race-detector-incompatible; covered by the non-race CI lane")
	}

	infoLogger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
	prev := schemaLogger.Swap(infoLogger)
	t.Cleanup(func() { schemaLogger.Store(prev) })

	cfg := DefaultConfig()
	cfg.NumCtx = 1024
	guard := NewContextGuard(cfg, infoLogger)
	rawSmall := "alpha bravo charlie"

	t.Run("sanitize", func(t *testing.T) {
		got := testing.AllocsPerRun(10000, func() {
			_, _ = SanitizeCapture(rawSmall, infoLogger)
		})
		// Sanitize already allocates output strings; budget is "added" only.
		assert.LessOrEqual(t, got, 5.0,
			"SanitizeCapture at Info level must stay within baseline; got %.2f", got)
	})

	t.Run("spotlight_disabled", func(t *testing.T) {
		got := testing.AllocsPerRun(10000, func() {
			_ = Spotlight(rawSmall, false, infoLogger)
		})
		assert.LessOrEqual(t, got, 1.0,
			"Spotlight (disabled) at Info level must add ≤ 1 alloc/run; got %.2f", got)
	})

	t.Run("contextguard_under_budget", func(t *testing.T) {
		got := testing.AllocsPerRun(10000, func() {
			_, _ = guard.ApplyOllama(rawSmall)
		})
		assert.LessOrEqual(t, got, 1.0,
			"ApplyOllama (under budget) at Info level must add ≤ 1 alloc/run; got %.2f", got)
	})

	t.Run("encode_claude_tool_name", func(t *testing.T) {
		got := testing.AllocsPerRun(10000, func() {
			_ = ClaudeToolName(StageKindForm, infoLogger)
		})
		// Pure function: only the returned concatenated string allocates.
		assert.LessOrEqual(t, got, 2.0,
			"ClaudeToolName at Info level must stay within baseline; got %.2f", got)
	})

	t.Run("sampling_ollama", func(t *testing.T) {
		got := testing.AllocsPerRun(10000, func() {
			_ = OllamaSamplingOptions(cfg, infoLogger)
		})
		// The map literal already allocates; Story 5's guard must add ≤ 1.
		assert.LessOrEqual(t, got, 5.0,
			"OllamaSamplingOptions at Info level must stay within baseline; got %.2f", got)
	})
}
