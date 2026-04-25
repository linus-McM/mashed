//go:build testing

// Story 2 RED-phase: NewMock lives behind `//go:build testing` (mock.go is
// tagged), so its nil-logger smoke test belongs in a parallel tagged file
// alongside the rest of the Story 2 plumbing tests.
package uiadapter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStory2_AC2_AC1_NewMockAcceptsNilLogger asserts NewMock's reshape
// (`NewMock(fixed *UIAST, logger *slog.Logger) Adapter`) accepts a nil
// logger and yields a working Adapter whose Translate path does not panic.
func TestStory2_AC2_AC1_NewMockAcceptsNilLogger(t *testing.T) {
	t.Parallel()

	fixed := &UIAST{
		Version:     "1",
		GeneratedBy: "mock",
		Nodes:       []UINode{{Type: "markdown", Content: "fixture"}},
	}

	assert.NotPanics(t, func() {
		a := NewMock(fixed, nil)
		require.NotNil(t, a, "NewMock must return a non-nil Adapter")
		got := a.Translate(context.Background(), "raw", "proc-1")
		require.Same(t, fixed, got,
			"NewMock(fixed, nil).Translate must return the supplied fixture verbatim")
	})
}
