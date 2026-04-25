// Package uiadapter — Story 1: UI Adapter logging infrastructure and boot wiring.
//
// RED-phase tests for `internal/uiadapter/logging.go` (which does not exist yet)
// and the new `Config.LogLevel` / `Config.LogDir` fields.
//
// These tests are intentionally failing (compile error) until the go-engineer
// ships:
//   - `NewProductionLogger(level slog.Level, logDir string) (*slog.Logger, io.Closer, error)`
//   - `parseSlogLevel(s string) slog.Level`
//   - `Config.LogLevel string` (default "info")
//   - `Config.LogDir   string` (default "./logs")
//
// Test names reference the AC numbers in
// `docs/stories/uiadapter-logging-1-infrastructure-and-boot.md`.
package uiadapter

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// captureStdout swaps os.Stdout for a pipe for the duration of fn() and returns
// everything written. The pipe is drained on a goroutine so a blocking writer
// in fn() can never deadlock the test.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()

	fn()
	_ = w.Close()
	return <-done
}

// expectedLogFile returns the file path NewProductionLogger should create today
// in UTC: `<logDir>/uiadapter-YYYYMMDD.log`.
func expectedLogFile(logDir string) string {
	return filepath.Join(logDir, "uiadapter-"+time.Now().UTC().Format("20060102")+".log")
}

// TestStory1_AC1_NewProductionLogger_FanoutToBothSinks asserts that the
// returned logger fans Debug records out to BOTH stdout (text) and the daily
// JSON log file under logDir.
//
// Story 1, AC-1.1: New `NewProductionLogger` returns stdout+file fan-out logger.
func TestStory1_AC1_NewProductionLogger_FanoutToBothSinks(t *testing.T) {
	logDir := t.TempDir()

	var (
		logger *slog.Logger
		closer io.Closer
		err    error
	)

	stdout := captureStdout(t, func() {
		logger, closer, err = NewProductionLogger(slog.LevelDebug, logDir)
		require.NoError(t, err, "construction must succeed when logDir is writable")
		require.NotNil(t, logger, "logger must be non-nil")
		require.NotNil(t, closer, "closer must be non-nil")

		logger.Debug("boot", "k", "v")
	})

	// Flush the file handle before reading. Idempotent close per AC-1.2.
	require.NoError(t, closer.Close(), "first Close must succeed")

	// --- stdout sink ---
	assert.Contains(t, stdout, "boot", "stdout text handler must emit msg=boot")
	assert.Contains(t, stdout, "k=v", "stdout text handler must emit attr k=v")

	// --- file sink ---
	logPath := expectedLogFile(logDir)
	raw, readErr := os.ReadFile(logPath)
	require.NoError(t, readErr, "expected log file %s to exist", logPath)

	lines := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
	require.Len(t, lines, 1, "exactly one JSON line expected in the log file")

	var rec map[string]any
	require.NoError(t, json.Unmarshal([]byte(lines[0]), &rec),
		"file sink must emit valid JSON; got: %q", lines[0])
	assert.Equal(t, "boot", rec["msg"], `JSON record must carry "msg":"boot"`)
	assert.Equal(t, "v", rec["k"], `JSON record must carry "k":"v"`)
}

// TestStory1_AC1_LevelFilter asserts that level=Info drops Debug records on
// BOTH sinks and lets Info records through on BOTH sinks.
//
// Story 1, AC-1.1 / Scenario 1 second clause.
func TestStory1_AC1_LevelFilter(t *testing.T) {
	logDir := t.TempDir()

	var closer io.Closer
	stdout := captureStdout(t, func() {
		logger, c, err := NewProductionLogger(slog.LevelInfo, logDir)
		require.NoError(t, err)
		require.NotNil(t, logger)
		closer = c

		logger.Debug("filtered-debug", "tag", "drop-me")
		logger.Info("kept-info", "tag", "keep-me")
	})
	require.NoError(t, closer.Close())

	// Stdout: debug suppressed, info present.
	assert.NotContains(t, stdout, "filtered-debug",
		"Debug records must be filtered out of stdout when level=Info")
	assert.Contains(t, stdout, "kept-info",
		"Info records must reach stdout when level=Info")

	// File: same expectations.
	raw, err := os.ReadFile(expectedLogFile(logDir))
	require.NoError(t, err)
	body := string(raw)
	assert.NotContains(t, body, "filtered-debug",
		"Debug records must be filtered out of file when level=Info")
	assert.Contains(t, body, "kept-info",
		"Info records must reach file when level=Info")
}

// TestStory1_AC2_CloserIdempotent asserts double-close is safe and subsequent
// log writes do not panic even after the file handle is closed.
//
// Story 1, AC-1.2: Closer is idempotent.
func TestStory1_AC2_CloserIdempotent(t *testing.T) {
	logDir := t.TempDir()

	var (
		logger *slog.Logger
		closer io.Closer
	)
	_ = captureStdout(t, func() {
		var err error
		logger, closer, err = NewProductionLogger(slog.LevelInfo, logDir)
		require.NoError(t, err)
		require.NotNil(t, logger)
		require.NotNil(t, closer)
	})

	assert.NoError(t, closer.Close(), "first Close must return nil")
	assert.NoError(t, closer.Close(), "second Close must return nil (idempotent)")

	assert.NotPanics(t, func() {
		logger.Info("post-close-write", "phase", "after-close")
	}, "logging after Close must not panic; file output may silently drop")
}

// TestStory1_AC3_StdoutOnlyFallback asserts that an unwritable logDir produces
// a non-nil logger that still writes to stdout, a non-nil no-op closer, and a
// non-nil error so callers can surface the warning.
//
// Story 1, AC-1.3: Fallback to stdout-only when log dir is unwritable.
func TestStory1_AC3_StdoutOnlyFallback(t *testing.T) {
	// /dev/null is a character device — MkdirAll of a child path fails on
	// macOS and Linux because /dev/null is not a directory.
	badDir := "/dev/null/cannot/exist"

	var (
		logger *slog.Logger
		closer io.Closer
		err    error
	)
	stdout := captureStdout(t, func() {
		logger, closer, err = NewProductionLogger(slog.LevelInfo, badDir)
		if logger != nil {
			logger.Info("smoke", "phase", "fallback")
		}
	})

	require.NotNil(t, logger, "logger must be non-nil even when log dir is unwritable")
	require.NotNil(t, closer, "closer must be non-nil (no-op closer is acceptable)")
	require.Error(t, err, "error must surface so caller can log a warning")

	assert.NoError(t, closer.Close(), "no-op closer Close must return nil")
	assert.Contains(t, stdout, "smoke", "stdout sink must continue to receive Info records")

	// No file should have been created under the bad path.
	_, statErr := os.Stat(expectedLogFile(badDir))
	assert.True(t, os.IsNotExist(statErr) || statErr != nil,
		"no log file should be created when logDir is unwritable; got stat err=%v", statErr)
}

// TestStory1_AC4_ParseSlogLevel covers parseSlogLevel — package-private helper
// that maps env-var strings to slog.Level. Empty / garbage inputs default to
// Info so a typo can never crash boot.
//
// Story 1, AC-1.4: `UIADAPTER_LOG_LEVEL` env var controls boot level.
func TestStory1_AC4_ParseSlogLevel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		want  slog.Level
	}{
		{name: "lowercase debug", input: "debug", want: slog.LevelDebug},
		{name: "uppercase DEBUG", input: "DEBUG", want: slog.LevelDebug},
		{name: "lowercase info", input: "info", want: slog.LevelInfo},
		{name: "lowercase warn", input: "warn", want: slog.LevelWarn},
		{name: "lowercase error", input: "error", want: slog.LevelError},
		{name: "empty string defaults to Info", input: "", want: slog.LevelInfo},
		{name: "garbage defaults to Info", input: "garbage", want: slog.LevelInfo},
		{name: "mixed case Warn", input: "Warn", want: slog.LevelWarn},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := parseSlogLevel(tc.input)
			assert.Equal(t, tc.want, got,
				"parseSlogLevel(%q) = %v, want %v", tc.input, got, tc.want)
		})
	}
}

// TestStory1_AC4_EnvVarBootLevel asserts the env-var → level pipeline used at
// boot in app.go: parseSlogLevel(os.Getenv("UIADAPTER_LOG_LEVEL")).
//
// Story 1, AC-1.4: Boot wiring honours UIADAPTER_LOG_LEVEL.
func TestStory1_AC4_EnvVarBootLevel(t *testing.T) {
	t.Setenv("UIADAPTER_LOG_LEVEL", "debug")
	assert.Equal(t, slog.LevelDebug, parseSlogLevel(os.Getenv("UIADAPTER_LOG_LEVEL")),
		"UIADAPTER_LOG_LEVEL=debug must yield slog.LevelDebug")

	t.Setenv("UIADAPTER_LOG_LEVEL", "")
	assert.Equal(t, slog.LevelInfo, parseSlogLevel(os.Getenv("UIADAPTER_LOG_LEVEL")),
		"unset UIADAPTER_LOG_LEVEL must default to slog.LevelInfo")

	t.Setenv("UIADAPTER_LOG_LEVEL", "not-a-level")
	assert.Equal(t, slog.LevelInfo, parseSlogLevel(os.Getenv("UIADAPTER_LOG_LEVEL")),
		"unparseable UIADAPTER_LOG_LEVEL must default to slog.LevelInfo without panicking")
}

// TestStory1_AC5_ConfigDefaults asserts the new Config.LogLevel and
// Config.LogDir fields are backfilled to "info" / "./logs" by mergeWithDefaults
// and that explicit caller values are preserved verbatim.
//
// Story 1, AC-1.5: Config exposes LogLevel and LogDir with defaults.
func TestStory1_AC5_ConfigDefaults(t *testing.T) {
	t.Parallel()

	t.Run("zero value backfilled to documented defaults", func(t *testing.T) {
		t.Parallel()
		merged := mergeWithDefaults(Config{})
		assert.Equal(t, "info", merged.LogLevel,
			"zero-value LogLevel must default to \"info\"")
		assert.Equal(t, "./logs", merged.LogDir,
			"zero-value LogDir must default to \"./logs\"")
	})

	t.Run("explicit values preserved", func(t *testing.T) {
		t.Parallel()
		merged := mergeWithDefaults(Config{
			LogLevel: "debug",
			LogDir:   "/var/log/mashed",
		})
		assert.Equal(t, "debug", merged.LogLevel,
			"explicit LogLevel must win over the default")
		assert.Equal(t, "/var/log/mashed", merged.LogDir,
			"explicit LogDir must win over the default")
	})

	t.Run("DefaultConfig also exposes the documented defaults", func(t *testing.T) {
		t.Parallel()
		def := DefaultConfig()
		assert.Equal(t, "info", def.LogLevel,
			"DefaultConfig().LogLevel must be \"info\"")
		assert.Equal(t, "./logs", def.LogDir,
			"DefaultConfig().LogDir must be \"./logs\"")
	})
}
