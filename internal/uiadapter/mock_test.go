//go:build testing

package uiadapter

import (
	"bytes"
	"context"
	"log/slog"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestU2_AC8_MockAdapter_FixedReturn — the mock returns the exact pointer
// passed to NewMock; callers can assert on the fixture directly.
func TestU2_AC8_MockAdapter_FixedReturn(t *testing.T) {
	t.Parallel()
	fixed := &UIAST{
		Version:     "1",
		GeneratedBy: "mock",
		Nodes:       []UINode{{Type: "markdown", Content: "fixture"}},
	}
	got := NewMock(fixed, nil).Translate(context.Background(), "", "")
	require.Same(t, fixed, got,
		"MockAdapter.Translate must return the exact pointer passed to NewMock")
}

// TestU2_AC8_MockAdapter_NilReturnsFallback — a nil fixture must produce a
// mock:nil fallback rather than panic.
func TestU2_AC8_MockAdapter_NilReturnsFallback(t *testing.T) {
	t.Parallel()
	got := NewMock(nil, nil).Translate(context.Background(), "raw body", "")
	require.NotNil(t, got)
	assert.Equal(t, "1", got.Version)
	assert.Equal(t, "fallback:mock:nil", got.GeneratedBy)
	require.Len(t, got.Nodes, 1)
	assert.Equal(t, "raw body", got.Nodes[0].Content)
}

// TestU2_AC8_MockAdapter_NotInProductionBuild — mock.go must only compile
// under -tags testing. Compare `go list -f '{{.GoFiles}}'` with and without
// the tag; mock.go must appear only in the tagged build.
func TestU2_AC8_MockAdapter_NotInProductionBuild(t *testing.T) {
	t.Parallel()

	prod := runGoList(t, "")
	assert.NotContains(t, prod, "mock.go",
		"mock.go must NOT compile into the production build (no -tags testing)")
	assert.NotContains(t, prod, "MockAdapter",
		"MockAdapter symbol must NOT leak into production sources")

	tagged := runGoList(t, "testing")
	assert.Contains(t, tagged, "mock.go",
		"mock.go MUST compile when -tags testing is active")
}

// TestStory5_AC5_MockInit — Story 5 AC-5.5.
// NewMock emits a mock.init record carrying op, enabled=true, fixture_name
// derived from fixed.GeneratedBy. An empty GeneratedBy falls back to
// "unnamed".
func TestStory5_AC5_MockInit(t *testing.T) {
	t.Parallel()

	t.Run("named fixture", func(t *testing.T) {
		t.Parallel()
		logger, buf := testLogBuffer(t, slog.LevelDebug)
		fixed := &UIAST{GeneratedBy: "fixture:happy"}

		_ = NewMock(fixed, logger)

		records := decodeRecords(t, buf)
		recs := recordsByMsg(records, "mock.init")
		require.GreaterOrEqual(t, len(recs), 1)
		assert.Equal(t, "mock.init", recs[0]["op"])
		assert.EqualValues(t, true, recs[0]["enabled"])
		assert.Equal(t, "fixture:happy", recs[0]["fixture_name"])
	})

	t.Run("unnamed fallback", func(t *testing.T) {
		t.Parallel()
		logger, buf := testLogBuffer(t, slog.LevelDebug)
		fixed := &UIAST{GeneratedBy: ""}

		_ = NewMock(fixed, logger)

		records := decodeRecords(t, buf)
		recs := recordsByMsg(records, "mock.init")
		require.GreaterOrEqual(t, len(recs), 1)
		assert.Equal(t, "unnamed", recs[0]["fixture_name"],
			"empty GeneratedBy must fall back to \"unnamed\"")
	})
}

// TestStory5_AC5_MockTranslate — Story 5 AC-5.5.
// MockAdapter.Translate emits mock.translate with op, proc_id, bytes_in.
func TestStory5_AC5_MockTranslate(t *testing.T) {
	t.Parallel()
	logger, buf := testLogBuffer(t, slog.LevelDebug)
	fixed := &UIAST{Version: "1", GeneratedBy: "fixture:happy"}
	adapter := NewMock(fixed, logger)

	// Reset the buffer so init records don't leak into the translate assertion.
	buf.Reset()

	_ = adapter.Translate(context.Background(), "raw input", "proc-99")

	records := decodeRecords(t, buf)
	recs := recordsByMsg(records, "mock.translate")
	require.GreaterOrEqual(t, len(recs), 1)
	assert.Equal(t, "mock.translate", recs[0]["op"])
	assert.Equal(t, "proc-99", recs[0]["proc_id"])
	assert.EqualValues(t, len("raw input"), recs[0]["bytes_in"])

	// §14: payload string must not leak.
	assert.NotContains(t, buf.String(), "raw input")
}

func runGoList(t *testing.T, tag string) string {
	t.Helper()
	args := []string{"list", "-f", "{{.GoFiles}}"}
	if tag != "" {
		args = append(args, "-tags", tag)
	}
	args = append(args, ".")

	var out, errb bytes.Buffer
	cmd := exec.Command("go", args...)
	cmd.Stdout = &out
	cmd.Stderr = &errb
	require.NoError(t, cmd.Run(),
		"go list %v failed: %s", args, errb.String())
	return strings.TrimSpace(out.String())
}
