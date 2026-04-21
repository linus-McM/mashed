//go:build testing

package uiadapter

import (
	"bytes"
	"context"
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
	got := NewMock(fixed).Translate(context.Background(), "", "")
	require.Same(t, fixed, got,
		"MockAdapter.Translate must return the exact pointer passed to NewMock")
}

// TestU2_AC8_MockAdapter_NilReturnsFallback — a nil fixture must produce a
// mock:nil fallback rather than panic.
func TestU2_AC8_MockAdapter_NilReturnsFallback(t *testing.T) {
	t.Parallel()
	got := NewMock(nil).Translate(context.Background(), "raw body", "")
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
