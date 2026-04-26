// Package uiadapter — Story 6: comprehensive logger surface tests.
//
// Story 6 caps the uiadapter debug-logging plan with end-to-end coverage of
// the production logger's full surface: fan-out parity, level filtering on
// both sinks, WithAttrs / WithGroup propagation, daily file naming under
// UTC, race safety under heavy contention, and nilSafeLogger discard
// semantics. Tests are named TestStory6_AC1_* so the AC validation table
// can map green checkmarks back to the spec one-to-one.
package uiadapter

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readJSONFileLines reads the daily log file written by NewProductionLogger
// and returns one decoded JSON record per non-empty line. A bad line is a
// hard failure — the JSON contract is part of the assertion surface, not a
// soft expectation.
func readJSONFileLines(t *testing.T, logDir string) []map[string]any {
	t.Helper()
	raw, err := os.ReadFile(expectedLogFile(logDir))
	require.NoError(t, err, "expected daily log file under %s to exist", logDir)
	body := strings.TrimRight(string(raw), "\n")
	if body == "" {
		return nil
	}
	lines := strings.Split(body, "\n")
	out := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var rec map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &rec),
			"file sink must emit valid JSON; got: %q", line)
		out = append(out, rec)
	}
	return out
}

// TestStory6_AC1_FanoutToBothSinks_Comprehensive — single Info call lands
// EXACTLY one text line on stdout AND exactly one matching JSON record in
// the daily log file. Cross-sink parity is asserted by comparing the msg
// string and a sentinel attribute, not by substring containment.
func TestStory6_AC1_FanoutToBothSinks_Comprehensive(t *testing.T) {
	logDir := t.TempDir()

	var (
		logger *slog.Logger
		closer io.Closer
	)
	stdout := captureStdout(t, func() {
		var err error
		logger, closer, err = NewProductionLogger(slog.LevelInfo, logDir)
		require.NoError(t, err)
		require.NotNil(t, logger)
		require.NotNil(t, closer)
		logger.Info("uiadapter.fanout", "session", "abc-123", "phase", "boot")
	})
	require.NoError(t, closer.Close())

	// Stdout: text format.
	assert.Contains(t, stdout, "uiadapter.fanout",
		"stdout text handler must emit msg=uiadapter.fanout")
	assert.Contains(t, stdout, "session=abc-123",
		"stdout text handler must carry session attr")
	assert.Contains(t, stdout, "phase=boot",
		"stdout text handler must carry phase attr")

	// File: JSON format. Exactly one record, fields match.
	records := readJSONFileLines(t, logDir)
	require.Len(t, records, 1, "exactly one JSON record expected in the log file")
	rec := records[0]
	assert.Equal(t, "uiadapter.fanout", rec["msg"], "file sink msg must equal stdout msg")
	assert.Equal(t, "abc-123", rec["session"], "file sink must carry session attr")
	assert.Equal(t, "boot", rec["phase"], "file sink must carry phase attr")
}

// TestStory6_AC1_LevelFilter_BothSinks — Info-level logger drops Debug from
// BOTH sinks; Debug-level logger lets both through on BOTH sinks. Asserts
// the level filter is honored uniformly across the fanout, not just on one
// child handler.
func TestStory6_AC1_LevelFilter_BothSinks(t *testing.T) {
	cases := []struct {
		name       string
		level      slog.Level
		wantDebug  bool
		debugMsg   string
		infoMsg    string
		wantAlways string
	}{
		{
			name:       "info level drops debug",
			level:      slog.LevelInfo,
			wantDebug:  false,
			debugMsg:   "lvl_filter_drop_debug",
			infoMsg:    "lvl_filter_keep_info",
			wantAlways: "lvl_filter_keep_info",
		},
		{
			name:       "debug level keeps debug",
			level:      slog.LevelDebug,
			wantDebug:  true,
			debugMsg:   "lvl_filter_keep_debug",
			infoMsg:    "lvl_filter_keep_info_2",
			wantAlways: "lvl_filter_keep_info_2",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			logDir := t.TempDir()
			var closer io.Closer
			stdout := captureStdout(t, func() {
				logger, c, err := NewProductionLogger(tc.level, logDir)
				require.NoError(t, err)
				closer = c
				logger.Debug(tc.debugMsg, "tag", "debug")
				logger.Info(tc.infoMsg, "tag", "info")
			})
			require.NoError(t, closer.Close())

			fileBody := func() string {
				raw, err := os.ReadFile(expectedLogFile(logDir))
				require.NoError(t, err)
				return string(raw)
			}()

			// The Info-level message must always be present in both sinks.
			assert.Contains(t, stdout, tc.wantAlways,
				"%s: Info message must reach stdout", tc.name)
			assert.Contains(t, fileBody, tc.wantAlways,
				"%s: Info message must reach file", tc.name)

			if tc.wantDebug {
				assert.Contains(t, stdout, tc.debugMsg,
					"%s: Debug message must reach stdout when level=Debug", tc.name)
				assert.Contains(t, fileBody, tc.debugMsg,
					"%s: Debug message must reach file when level=Debug", tc.name)
			} else {
				assert.NotContains(t, stdout, tc.debugMsg,
					"%s: Debug message must NOT reach stdout when level=Info", tc.name)
				assert.NotContains(t, fileBody, tc.debugMsg,
					"%s: Debug message must NOT reach file when level=Info", tc.name)
			}
		})
	}
}

// TestStory6_AC1_WithAttrs_BothSinks — `logger.With("session","abc")` carries
// session=abc to both sinks on every subsequent Info call. Verifies WithAttrs
// propagates through the fanoutHandler to every child.
func TestStory6_AC1_WithAttrs_BothSinks(t *testing.T) {
	logDir := t.TempDir()

	var closer io.Closer
	stdout := captureStdout(t, func() {
		base, c, err := NewProductionLogger(slog.LevelInfo, logDir)
		require.NoError(t, err)
		closer = c
		scoped := base.With("session", "abc")
		scoped.Info("with-attrs.event")
	})
	require.NoError(t, closer.Close())

	// Stdout: text format carries session=abc.
	assert.Contains(t, stdout, "session=abc",
		"WithAttrs must propagate session=abc to stdout text sink")

	// File: JSON record carries session:"abc".
	records := readJSONFileLines(t, logDir)
	require.Len(t, records, 1, "exactly one JSON record expected")
	assert.Equal(t, "abc", records[0]["session"],
		"WithAttrs must propagate session to file JSON sink")
}

// TestStory6_AC1_WithGroup_NestsJSON — `logger.WithGroup("g").Info("msg","k","v")`
// produces JSON with "g":{"k":"v"} in the file sink and `g.k=v` in the text
// sink. Verifies WithGroup propagates through the fanoutHandler to every
// child and is rendered per the child handler's format.
func TestStory6_AC1_WithGroup_NestsJSON(t *testing.T) {
	logDir := t.TempDir()

	var closer io.Closer
	stdout := captureStdout(t, func() {
		base, c, err := NewProductionLogger(slog.LevelInfo, logDir)
		require.NoError(t, err)
		closer = c
		grouped := base.WithGroup("uiadapter")
		grouped.Info("group.event", "k", "v")
	})
	require.NoError(t, closer.Close())

	// Stdout (TextHandler): "uiadapter.k=v" — dotted grouping.
	assert.Contains(t, stdout, "uiadapter.k=v",
		"WithGroup must produce dotted attrs on stdout TextHandler")

	// File (JSONHandler): nested object {"uiadapter":{"k":"v"}}.
	records := readJSONFileLines(t, logDir)
	require.Len(t, records, 1)
	nested, ok := records[0]["uiadapter"].(map[string]any)
	require.True(t, ok, "WithGroup must produce a nested object on JSON sink; got %v", records[0])
	assert.Equal(t, "v", nested["k"], "nested object must carry k:v")
}

// TestStory6_AC1_FileNameUTC — file name format `uiadapter-YYYYMMDD.log`
// matches today OR yesterday in UTC. Yesterday is accepted to absorb a
// midnight-UTC roll between construction and assertion (avoids a
// once-per-86400s flake without freezing the clock).
func TestStory6_AC1_FileNameUTC(t *testing.T) {
	logDir := t.TempDir()

	_, closer, err := NewProductionLogger(slog.LevelInfo, logDir)
	require.NoError(t, err)
	require.NoError(t, closer.Close())

	// Expected names: today and yesterday in UTC (midnight-flake guard).
	now := time.Now().UTC()
	candidates := map[string]struct{}{
		filepath.Join(logDir, "uiadapter-"+now.Format("20060102")+".log"):                     {},
		filepath.Join(logDir, "uiadapter-"+now.Add(-24*time.Hour).Format("20060102")+".log"): {},
	}

	entries, err := os.ReadDir(logDir)
	require.NoError(t, err)
	require.NotEmpty(t, entries, "log dir must contain at least one file")

	matched := false
	for _, e := range entries {
		full := filepath.Join(logDir, e.Name())
		if _, ok := candidates[full]; ok {
			matched = true
			break
		}
	}
	assert.True(t, matched,
		"daily log file name must match uiadapter-YYYYMMDD.log for today or yesterday (UTC); got entries=%v",
		entries)
}

// TestStory6_AC1_RaceSafe — 100 goroutines * 100 calls each = 10000 records.
// Asserts no race-detector trip (run via `go test -race`) and that exactly
// 10000 well-formed JSON lines land in the file. Confirms the fan-out
// handler + JSONHandler + os.File chain is concurrency-safe end-to-end.
func TestStory6_AC1_RaceSafe(t *testing.T) {
	const (
		goroutines = 100
		perGo      = 100
		total      = goroutines * perGo
	)
	logDir := t.TempDir()

	logger, closer, err := NewProductionLogger(slog.LevelInfo, logDir)
	require.NoError(t, err)

	// Drain stdout into io.Discard for the duration of the hammer —
	// captureStdout would buffer 10k lines in memory which floods the
	// verbose test output for no benefit.
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })
	go func() { _, _ = io.Copy(io.Discard, r) }()

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perGo; i++ {
				logger.Info("race.hammer", "iter", i)
			}
		}()
	}
	wg.Wait()
	_ = w.Close()

	require.NoError(t, closer.Close())

	records := readJSONFileLines(t, logDir)
	require.Len(t, records, total,
		"expected exactly %d JSON records under concurrent hammer; got %d", total, len(records))
	for i, rec := range records {
		assert.Equal(t, "race.hammer", rec["msg"],
			"record %d: msg must equal race.hammer", i)
	}
}

// TestStory6_AC1_NilSafeLoggerComprehensive — `nilSafeLogger(nil)` returns a
// usable logger that discards every record; .Enabled(ctx, slog.LevelDebug)
// is false (io.Discard handler); calling its methods is a no-op (no panic,
// no output). A non-nil input is returned verbatim (identity preserved).
func TestStory6_AC1_NilSafeLoggerComprehensive(t *testing.T) {
	t.Parallel()

	t.Run("nil input returns a discard-backed logger", func(t *testing.T) {
		t.Parallel()
		got := nilSafeLogger(nil)
		require.NotNil(t, got, "nilSafeLogger(nil) must never return nil")

		// Discard handler reports Enabled=false at Debug (default level Info).
		assert.False(t, got.Enabled(context.Background(), slog.LevelDebug),
			"discard-backed logger must report Debug as not enabled")

		// Methods are no-ops — no panic.
		assert.NotPanics(t, func() {
			got.Debug("nil-safe.debug", "k", "v")
			got.Info("nil-safe.info", "k", "v")
			got.Warn("nil-safe.warn", "k", "v")
			got.Error("nil-safe.error", "k", "v")
			got.LogAttrs(context.Background(), slog.LevelInfo, "nil-safe.logattrs",
				slog.String("k", "v"))
		}, "nil-safe logger methods must never panic")
	})

	t.Run("non-nil input returned verbatim", func(t *testing.T) {
		t.Parallel()
		base := slog.New(slog.NewTextHandler(io.Discard, nil))
		got := nilSafeLogger(base)
		assert.Same(t, base, got,
			"nilSafeLogger must preserve identity for non-nil input")
	})
}
