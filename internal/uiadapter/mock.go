//go:build testing

package uiadapter

import (
	"context"
	"log/slog"
)

// MockAdapter returns a fixed AST on every Translate call. It exists only
// under -tags testing so executor tests (U4+) can inject deterministic
// fixtures without spinning up an HTTP stub.
type MockAdapter struct {
	Fixed  *UIAST
	logger *slog.Logger
}

// NewMock wraps fixed into an Adapter. A nil fixed causes Translate to
// degrade to a fallback:mock:nil AST rather than panic — mirrors the
// defaultAdapter contract that Translate never returns nil. logger may be
// nil; nilSafeLogger normalises it so the field is always usable.
func NewMock(fixed *UIAST, logger *slog.Logger) Adapter {
	return &MockAdapter{Fixed: fixed, logger: nilSafeLogger(logger)}
}

// Translate returns the exact pointer supplied to NewMock, unless that
// pointer was nil in which case a fallback AST carrying the raw capture is
// returned instead.
func (m *MockAdapter) Translate(_ context.Context, raw, _ string) *UIAST {
	if m.Fixed == nil {
		return FallbackAST(raw, "mock:nil", m.logger)
	}
	return m.Fixed
}
