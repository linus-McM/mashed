// Package uiadapter — production logging conduit (Story 1).
//
// NewProductionLogger returns an *slog.Logger that fans out every record to
// two sinks:
//
//   - stdout, via slog.NewTextHandler (human-readable),
//   - <logDir>/uiadapter-YYYYMMDD.log, via slog.NewJSONHandler (structured).
//
// Daily rollover is achieved by deriving the file name from time.Now().UTC()
// at construction time. There is no mid-process rotation.
//
// # Standard attribute names
//
// Subsequent stories MUST use these attribute keys when calling logger.With /
// logger.Info / logger.Debug etc., so log queries stay uniform across the
// adapter:
//
//   - op             — operation name, e.g. "translate", "cache.lookup"
//   - proc_id        — caller-supplied procedure correlation id
//   - model          — backend model identifier
//   - latency_ms     — wall-clock duration in milliseconds
//   - bytes_in       — request byte count
//   - bytes_out      — response byte count
//   - tier           — pipeline tier ("fast", "spotlight", "fallback", ...)
//   - reason         — short fallback / error category tag
//   - cache_key_hash — non-reversible hash of the cache key (never the raw key)
//
// This file ships only the conduit. Per-call instrumentation (and the §14
// sanitize discipline that goes with it) lands in later stories.
package uiadapter

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// NewProductionLogger constructs the production *slog.Logger. The returned
// io.Closer must be retained by the caller and invoked at shutdown — it is
// idempotent and safe to call after a fallback (in which case it is a no-op).
//
// The function never returns nil for the logger or closer: when logDir cannot
// be created or the daily file cannot be opened it degrades to a stdout-only
// logger and surfaces the underlying error so the caller can warn.
func NewProductionLogger(level slog.Level, logDir string) (*slog.Logger, io.Closer, error) {
	opts := &slog.HandlerOptions{Level: level}
	stdout := slog.NewTextHandler(os.Stdout, opts)

	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return slog.New(stdout), noopCloser{}, fmt.Errorf("uiadapter: create log dir %q: %w", logDir, err)
	}

	fileName := filepath.Join(logDir, "uiadapter-"+time.Now().UTC().Format("20060102")+".log")
	f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return slog.New(stdout), noopCloser{}, fmt.Errorf("uiadapter: open log file %q: %w", fileName, err)
	}

	jsonH := slog.NewJSONHandler(f, opts)
	fanout := fanoutHandler{stdout, jsonH}
	return slog.New(fanout), &fileCloser{f: f}, nil
}

// ParseLogLevel is the exported wrapper around parseSlogLevel for callers
// outside the package (e.g. boot wiring in app.go). The behaviour is
// identical: empty or unrecognized values fall through to slog.LevelInfo.
func ParseLogLevel(s string) slog.Level { return parseSlogLevel(s) }

// parseSlogLevel maps a case-insensitive level string to slog.Level. Empty or
// unrecognized values fall through to slog.LevelInfo so a typo in the
// UIADAPTER_LOG_LEVEL env var can never crash boot.
func parseSlogLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// fanoutHandler dispatches each slog.Record to every child handler. It is
// stateless and thus safe for concurrent use as long as every child is.
// stdlib TextHandler and JSONHandler are.
type fanoutHandler []slog.Handler

// Enabled returns true if ANY child handler is enabled at the requested
// level. This keeps the hot path cheap and lets callers skip allocation when
// neither sink wants the record.
func (h fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, child := range h {
		if child.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

// Handle dispatches to every child whose Enabled returns true. Each child
// receives its own clone of the record so any mutation a handler performs
// stays local.
func (h fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, child := range h {
		if !child.Enabled(ctx, r.Level) {
			continue
		}
		if err := child.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// WithAttrs returns a new fanout where every child has WithAttrs applied.
func (h fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make(fanoutHandler, len(h))
	for i, child := range h {
		out[i] = child.WithAttrs(attrs)
	}
	return out
}

// WithGroup returns a new fanout where every child has WithGroup applied.
func (h fanoutHandler) WithGroup(name string) slog.Handler {
	out := make(fanoutHandler, len(h))
	for i, child := range h {
		out[i] = child.WithGroup(name)
	}
	return out
}

// fileCloser closes an *os.File exactly once. The first Close attempt closes
// the file; subsequent calls return nil, even if the first call errored, so
// AC-1.2 is structurally guaranteed.
type fileCloser struct {
	f    *os.File
	once sync.Once
}

// Close implements io.Closer. It is safe to call concurrently and any number
// of times.
func (c *fileCloser) Close() error {
	c.once.Do(func() { _ = c.f.Close() })
	return nil
}

// noopCloser is the fallback closer returned when no log file was ever
// opened. Its Close always returns nil.
type noopCloser struct{}

// Close implements io.Closer.
func (noopCloser) Close() error { return nil }
