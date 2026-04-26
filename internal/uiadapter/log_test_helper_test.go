// Package uiadapter — Story 3 shared test helpers.
//
// `testLogBuffer`, `decodeRecords`, and `recordsByMsg` are the canonical
// triple every Story 3 test uses to drive a *slog.Logger over an in-memory
// bytes.Buffer and assert on the JSON records that come out. The file lives
// behind `_test.go` so it never compiles into the production binary.
//
// Centralised here so each instrumented file's test (client, cache,
// breaker, semaphore, prefix_cache, plus the cross-cutting sanitize and
// alloc tests) can share the same decoder loop without copy-paste drift.
package uiadapter

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// testLogBuffer returns an *slog.Logger backed by a JSON handler writing to a
// bytes.Buffer at the supplied level. Tests use the buffer to decode the
// emitted records and assert on attribute shape / sanitize discipline.
//
// The handler clones records before emitting, so the buffer is safe to read
// after the logger is dropped. Buffer is heap-allocated so the caller can
// pass it across goroutine boundaries (see semaphore tests).
func testLogBuffer(t *testing.T, level slog.Level) (*slog.Logger, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	h := slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: level})
	return slog.New(h), buf
}

// decodeRecords parses every newline-delimited JSON record in buf and returns
// them as ordered maps. Empty trailing lines are tolerated (slog.JSONHandler
// always terminates each record with a newline).
//
// A bad line is treated as a hard test failure — the JSON contract is part
// of the assertion surface, not a soft expectation.
func decodeRecords(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	body := strings.TrimRight(buf.String(), "\n")
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
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("decodeRecords: bad JSON line %q: %v", line, err)
		}
		out = append(out, rec)
	}
	return out
}

// recordsByMsg returns every record whose `msg` field equals the supplied
// string. Used for "exactly one record with msg=X" assertions in Story 3.
func recordsByMsg(records []map[string]any, msg string) []map[string]any {
	out := make([]map[string]any, 0, len(records))
	for _, r := range records {
		if s, _ := r["msg"].(string); s == msg {
			out = append(out, r)
		}
	}
	return out
}
