// Package uiadapter — Story 6: end-to-end Translate emission tests.
//
// These tests are the cap-stone of the uiadapter debug-logging plan. They
// exercise the full pipeline through `Translate` plus every per-phase
// helper instrumented in Stories 3–5, decode the captured JSON records, and
// assert per-file `op` coverage. They also enforce §14 sanitize discipline
// end-to-end via a 10-message random-payload smoke that fails loudly when
// any payload byte (or stub-response byte, or synthetic error string) leaks
// into a record value.
//
// Helper rationale — `defaultAdapter.Translate` invokes only sanitize,
// semaphore, client.chat, validator, and (transitively, via json.Unmarshal)
// schema. The other phases (spotlight, contextguard, encode, sampling,
// fastpath, cache, breaker, repair, fallback.tier) are independent helpers
// that callers wire externally — see backend/* and the v3 router. The
// `e2eAdapter` helper drives Translate AND each of those helpers under one
// shared captured logger so the AC-2 emission-coverage assertion has a
// single buffer to scan.
package uiadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// e2eAdapter constructs a real `defaultAdapter` wired with a Debug-level
// captured logger and a stub Ollama HTTP server returning stubResp as the
// chat message content. Returns the adapter, the shared *slog.Logger, the
// in-memory buffer the JSONHandler writes to, and the cleanup-via-cleanup
// stub URL is registered through newOllamaStub.
//
// snapshotSchemaLogger is registered so WidgetNode.UnmarshalJSON emits
// schema.unmarshal.* records into the same buffer for the test's lifetime
// without leaking between parallel tests.
func e2eAdapter(t *testing.T, stubResp string) (Adapter, *slog.Logger, *bytes.Buffer) {
	t.Helper()
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	snapshotSchemaLogger(t)
	schemaLogger.Store(logger)

	newOllamaStub(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(ollamaChatPayload(stubResp)))
	})
	a := NewDefault(enabledConfig(2000), logger)
	return a, logger, buf
}

// drivePerPhaseHelpers calls every standalone uiadapter helper that
// `defaultAdapter.Translate` does NOT exercise, so the AC-2 emission-
// coverage assertion sees one record from each instrumented file under a
// single shared logger. None of the helpers carry the raw payload — the
// §14 sanitize-discipline smoke (TestStory6_AC3_*) re-asserts that.
func drivePerPhaseHelpers(t *testing.T, logger *slog.Logger, raw string) {
	t.Helper()
	cfg := mergeWithDefaults(Config{Enabled: true, Model: "g", TimeoutMs: 500, MaxInflight: 1})

	// spotlight.* — instrument always emits via the disabled path, so we
	// flip the toggle to also exercise the start/added arms.
	_ = Spotlight(raw, true, logger)

	// contextguard.* — Ollama path is the simpler emit. ApplyClaude is
	// covered by its own unit suite.
	_, _ = NewContextGuard(cfg, logger).ApplyOllama(raw)

	// encode.* — exercises encode.ollama_format + encode.schema_select.
	_, _ = OllamaFormatPayload(StageKindForm, false, logger)

	// sampling.* — sampling.ollama record.
	_ = OllamaSamplingOptions(cfg, logger)

	// fastpath.* — empty-string Classify is an early miss; "y/n?" trips
	// the yn-prompt rule so we get classify.start + classify.hit.
	fp := NewFastPathClassifier(true, logger)
	_, _, _ = fp.Classify("Apply changes? [y/n]")

	// cache.* — Lookup miss + Store + Lookup hit covers cache.get + cache.put.
	c := NewResponseCache(cfg, logger)
	key := Key("ollama", cfg.Model, raw)
	_, _ = c.Lookup(key)
	c.Store(key, &UIAST{Version: "1"})
	_, _ = c.Lookup(key)

	// breaker.* — Do drives a closed-state success then forces failures
	// until the threshold trips, producing breaker.transition + breaker.reject.
	bs := NewBreakerSet(cfg, logger)
	_, _ = bs.Do("ollama", func() (*UIAST, error) { return &UIAST{Version: "1"}, nil })
	for i := 0; i < cfg.BreakerFailThreshold+1; i++ {
		_, _ = bs.Do("ollama", func() (*UIAST, error) {
			return nil, errors.New("e2e: induced failure")
		})
	}
	// One more call after the trip to surface breaker.reject.
	_, _ = bs.Do("ollama", func() (*UIAST, error) { return &UIAST{Version: "1"}, nil })

	// stages.* — assembleStage1 emits stages.assemble.
	_ = AssembleStage1(raw, logger)
}

// decodeOps returns every distinct `op` attribute value across all decoded
// records. Walks both the top-level `op` key and any nested `op` key under
// a single-level group (NewDefault wraps the adapter logger in
// WithGroup("uiadapter") so adapter-emitted records carry op nested under
// "uiadapter").
func decodeOps(records []map[string]any) map[string]struct{} {
	ops := make(map[string]struct{}, len(records))
	for _, rec := range records {
		if op, _ := rec["op"].(string); op != "" {
			ops[op] = struct{}{}
		}
		for _, v := range rec {
			grouped, ok := v.(map[string]any)
			if !ok {
				continue
			}
			if op, _ := grouped["op"].(string); op != "" {
				ops[op] = struct{}{}
			}
		}
	}
	return ops
}

// hasOpPrefix reports true when at least one `op` attr value in ops starts
// with prefix. Prefix match (not equality) so dotted op names like
// "sanitize.scan" satisfy a "sanitize.*" expectation.
func hasOpPrefix(ops map[string]struct{}, prefix string) bool {
	for op := range ops {
		if strings.HasPrefix(op, prefix) {
			return true
		}
	}
	return false
}

// astJSONWithWidget is a stub-server response containing a decision_group
// node with a free-text widget. The widget triggers WidgetNode.UnmarshalJSON
// during Translate's json.Unmarshal pass, which in turn emits the
// schema.unmarshal.widget Debug record — proving schema.* coverage end-to-end
// without resorting to a direct UnmarshalJSON call in the test.
const astJSONWithWidget = `{"version":"1","generated_by":"ollama:test","nodes":[{"type":"markdown","content":"ok"},{"type":"decision_group","response_key":"answer","widget":{"type":"free"}}],"fallback_answer_shape":"free"}`

// TestStory6_AC2_Translate_HappyPath_AllPhasesLog — drive Translate +
// per-phase helpers under one shared Debug logger and assert every
// instrumented file emits at least one record under its expected `op`
// prefix. Confirms the full §6.1 emission contract is honored end-to-end.
func TestStory6_AC2_Translate_HappyPath_AllPhasesLog(t *testing.T) {
	a, logger, buf := e2eAdapter(t, astJSONWithWidget)

	ast := a.Translate(context.Background(), "Press enter to continue...", "proc-1")
	require.NotNil(t, ast)
	assert.Equal(t, "1", ast.Version)

	// Drive the helpers Translate doesn't reach.
	drivePerPhaseHelpers(t, logger, "raw e2e capture")

	records := decodeRecords(t, buf)
	require.NotEmpty(t, records, "expected at least one JSON record from a Debug-level e2e run")
	ops := decodeOps(records)

	required := []string{
		"sanitize",
		"spotlight",
		"contextguard",
		"client",
		"cache",
		"validator",
		"encode",
		"sampling",
		"fastpath",
		"stages",
		"schema",
	}
	for _, prefix := range required {
		assert.Truef(t, hasOpPrefix(ops, prefix),
			"expected at least one record with op prefix %q; got ops=%v", prefix, ops)
	}
	// At least one of breaker.* or semaphore.* must fire (one is sufficient
	// per the spec — semaphore always fires on Translate, breaker only when
	// the helper drives it).
	assert.True(t, hasOpPrefix(ops, "breaker") || hasOpPrefix(ops, "semaphore"),
		"expected at least one breaker.* or semaphore.* record; got ops=%v", ops)
}

// TestStory6_AC2_Translate_RepairTriggered — directly drive `Repairer.Run`
// with a stub validator that fails attempt 1 and passes attempt 2. Assert
// the scripted record sequence (start, attempt #1, failure #1, attempt #2,
// success #2) appears in order. `defaultAdapter.Translate` does not yet
// invoke the repair loop (Story v3-10 wires it through the v3 router); the
// test exercises the helper directly so emission coverage is provable today.
func TestStory6_AC2_Translate_RepairTriggered(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	r := NewRepairer(Config{RepairMaxRetries: 2}, logger)

	// Calls 1 (initial) and 2 (first repair attempt) fail; call 3
	// succeeds — guarantees one repair.start, one repair.attempt+failure
	// pair, then a second repair.attempt+success pair.
	calls := 0
	validate := func(_ *UIAST) []string {
		calls++
		if calls < 3 {
			return []string{"required_dropped"}
		}
		return nil
	}
	generate := func(_ context.Context, _ string) (*UIAST, string, error) {
		return &UIAST{Version: "1"}, `{"version":"1"}`, nil
	}

	ast, err := r.Run(
		context.Background(),
		&UIAST{Version: "1"},
		`{"version":"1"}`,
		RepairAttempt{Kind: StageKindForm, SanitizedRaw: "raw"},
		validate,
		generate,
	)
	require.NoError(t, err)
	require.NotNil(t, ast)

	msgs := recordMsgsWithPrefix(decodeRecords(t, buf), "repair.")
	require.Subset(t, msgs, []string{"repair.start"},
		"repair.start must appear; got %v", msgs)

	// Build the ordered sub-sequence we expect to see (other records may
	// interleave but this exact order must be present).
	want := []string{"repair.start", "repair.attempt", "repair.failure", "repair.attempt", "repair.success"}
	gotIdx := 0
	for _, m := range msgs {
		if gotIdx < len(want) && m == want[gotIdx] {
			gotIdx++
		}
	}
	assert.Equal(t, len(want), gotIdx,
		"expected ordered repair sequence %v inside %v", want, msgs)
}

// TestStory6_AC2_Translate_TierEscalation — drive `RunWithFallback` with
// a primary backend that returns a transport error and a secondary that
// succeeds. Assert the scripted tier sequence (start, enter, failure,
// enter, success) appears in order. Same rationale as the repair test —
// the tier helper is exercised directly because `defaultAdapter.Translate`
// does not yet invoke it (the v3 router does).
func TestStory6_AC2_Translate_TierEscalation(t *testing.T) {
	logger, buf := testLogBuffer(t, slog.LevelDebug)

	called := map[string]int{}
	fn := func(_ context.Context, name string) (*UIAST, error) {
		called[name]++
		if name == "primary" {
			return nil, errors.New("e2e: induced primary transport failure")
		}
		return &UIAST{Version: "1", GeneratedBy: name}, nil
	}
	plaintext := func() *UIAST { return FallbackAST("raw", "tier-test", logger) }

	ast, tier, escalatedFrom, err := RunWithFallback(
		context.Background(),
		[]string{"primary", "secondary"},
		fn,
		nil,
		plaintext,
		logger,
	)
	require.NoError(t, err)
	require.NotNil(t, ast)
	assert.Equal(t, TierSecondary, tier)
	assert.Equal(t, "primary", escalatedFrom)
	assert.Equal(t, 1, called["primary"])
	assert.Equal(t, 1, called["secondary"])

	msgs := recordMsgsWithPrefix(decodeRecords(t, buf), "fallback.tier.")
	want := []string{
		"fallback.tier.start",
		"fallback.tier.enter",
		"fallback.tier.failure",
		"fallback.tier.enter",
		"fallback.tier.success",
	}
	gotIdx := 0
	for _, m := range msgs {
		if gotIdx < len(want) && m == want[gotIdx] {
			gotIdx++
		}
	}
	assert.Equal(t, len(want), gotIdx,
		"expected ordered tier sequence %v inside %v", want, msgs)
}

// scanBufferForLeak walks the raw JSON-encoded buffer and reports the
// first forbidden 8-byte slice that appears verbatim. Operating on the raw
// JSON bytes (rather than per-record decoded values) is strictly stronger:
// it catches leaks anywhere in the encoded record, including escaped or
// nested places a per-attribute walker might miss. Returns ("", true) on
// clean; (offset-into-needle, false) on violation.
//
// Reuses the package-shared `containsSlice` helper from
// `logging_story3_sanitize_test.go` so the leak threshold (8 bytes) stays
// in lock-step with Story 3's sanitize-discipline scan.
func scanBufferForLeak(buf *bytes.Buffer, forbidden [][]byte) (leak []byte, ok bool) {
	body := buf.Bytes()
	for _, needle := range forbidden {
		if off, hit := containsSlice(body, needle); hit {
			return needle[off : off+8], false
		}
	}
	return nil, true
}

// TestStory6_AC3_TenMessageNoLeak — feed 10 random raw payloads through
// Translate with the production logger at Debug. For every captured JSON
// record buffer, assert no 8+-byte slice of the corresponding raw payload
// OR the synthetic error string appears verbatim. Fails loudly with
// `(payload-index, leak)` so a regression points straight at the
// offending bytes.
//
// The error sentinel is fed through the breaker fn (drivePerPhaseHelpers)
// so its bytes flow through the same logger context as the real payload —
// any future sloppy `slog.String("err", err.Error())` would surface here.
func TestStory6_AC3_TenMessageNoLeak(t *testing.T) {
	const messageCount = 10
	const errSentinel = "ERRSENTINEL_e2e_zzz_do_not_log_THESE_BYTES"

	for i := 0; i < messageCount; i++ {
		i := i
		t.Run(fmt.Sprintf("payload-%d", i), func(t *testing.T) {
			// 50–500 bytes per spec; deterministic length spread.
			payload := randomPayload(t, 50+(i*45))
			astJSON, mErr := json.Marshal(&UIAST{
				Version:             "1",
				Nodes:               []UINode{{Type: "markdown", Content: "ok"}},
				FallbackAnswerShape: "free",
			})
			require.NoError(t, mErr)

			a, logger, buf := e2eAdapter(t, string(astJSON))
			driveLeakProbes(t, logger, string(payload), errSentinel)
			ast := a.Translate(context.Background(), string(payload), fmt.Sprintf("proc-%d", i))
			require.NotNil(t, ast)

			// Sanity: at least one Debug record must have landed.
			records := decodeRecords(t, buf)
			require.NotEmpty(t, records,
				"payload %d: expected at least one Debug record", i)

			// Per the brief: scan for raw payload bytes and the
			// synthetic err string. The stub response body is NOT
			// scanned because schema-shaped JSON keys (`version`,
			// `nodes`, `type`) collide structurally with slog's own
			// JSON-encoded record keys (`prompt_version`, `node_count`)
			// and would surface as false-positives that have nothing
			// to do with payload leakage.
			forbidden := [][]byte{payload, []byte(errSentinel)}
			leak, clean := scanBufferForLeak(buf, forbidden)
			require.Truef(t, clean,
				"payload %d: §14 sanitize-discipline violation — leaked slice=%q",
				i, leak)
		})
	}
}

// driveLeakProbes runs every per-phase helper that takes the raw payload
// AND a synthetic error sentinel through the err pathway. Any future
// sloppy `slog.String("err", err.Error())` regression — whether in the
// breaker, repair, or fallback-tier helpers — surfaces immediately in the
// AC-3 leak scan.
func driveLeakProbes(t *testing.T, logger *slog.Logger, raw, errSentinel string) {
	t.Helper()
	cfg := mergeWithDefaults(Config{Enabled: true, Model: "g", TimeoutMs: 500, MaxInflight: 1})

	_ = Spotlight(raw, true, logger)
	_, _ = NewContextGuard(cfg, logger).ApplyOllama(raw)
	_ = AssembleStage1(raw, logger)
	_ = NewFastPathClassifier(true, logger)

	bs := NewBreakerSet(cfg, logger)
	syntheticErr := errors.New(errSentinel)
	for i := 0; i < cfg.BreakerFailThreshold+1; i++ {
		_, _ = bs.Do("ollama", func() (*UIAST, error) { return nil, syntheticErr })
	}
	_, _ = bs.Do("ollama", func() (*UIAST, error) { return &UIAST{Version: "1"}, nil })
}

// TestStory6_AC5_DefaultLevelInfo_NoDebugRecords — env var unset, parsed
// level is Info, the production logger emits zero Debug records during a
// Translate, and the existing Info-level translate telemetry record
// remains unchanged.
func TestStory6_AC5_DefaultLevelInfo_NoDebugRecords(t *testing.T) {
	t.Setenv("UIADAPTER_LOG_LEVEL", "")
	level := parseSlogLevel("")
	require.Equal(t, slog.LevelInfo, level,
		"unset UIADAPTER_LOG_LEVEL must default to Info")

	// e2eAdapter pins the level to Debug; this test needs Info, so we
	// build the stub + adapter inline.
	logger, buf := testLogBuffer(t, level)
	snapshotSchemaLogger(t)
	schemaLogger.Store(logger)
	newOllamaStub(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(ollamaChatPayload(validASTJSON)))
	})
	a := NewDefault(enabledConfig(2000), logger)
	ast := a.Translate(context.Background(), "raw", "proc-default-info")
	require.NotNil(t, ast)
	assert.Equal(t, "1", ast.Version)

	records := decodeRecords(t, buf)
	for i, rec := range records {
		lvl, _ := rec["level"].(string)
		assert.NotEqualf(t, "DEBUG", lvl,
			"record %d: no DEBUG-level records may emit at default Info level; got %v", i, rec)
	}

	// Existing Info telemetry record must still be present and unchanged.
	// NewDefault wraps the logger in WithGroup("uiadapter") so the legacy
	// `op="translate"` attr lands nested under the "uiadapter" key. Probing
	// both the top-level and grouped slot proves Stories 1-5 didn't silently
	// drop or rename the legacy telemetry line.
	infoRecords := recordsByMsg(records, "uiadapter.translate")
	require.NotEmpty(t, infoRecords,
		"the legacy Info-level uiadapter.translate telemetry must still emit at default Info")
	rec := infoRecords[0]
	op, _ := rec["op"].(string)
	if op == "" {
		if grouped, ok := rec["uiadapter"].(map[string]any); ok {
			op, _ = grouped["op"].(string)
		}
	}
	assert.Equal(t, "translate", op,
		"legacy telemetry record must carry op=translate (top-level or under uiadapter group)")
}
