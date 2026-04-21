//go:build testing

package uiadapter

import "context"

// MockAdapter returns a fixed AST on every Translate call. It exists only
// under -tags testing so executor tests (U4+) can inject deterministic
// fixtures without spinning up an HTTP stub.
type MockAdapter struct {
	Fixed *UIAST
}

// NewMock wraps fixed into an Adapter. A nil fixed causes Translate to
// degrade to a fallback:mock:nil AST rather than panic — mirrors the
// defaultAdapter contract that Translate never returns nil.
func NewMock(fixed *UIAST) Adapter {
	return &MockAdapter{Fixed: fixed}
}

// Translate returns the exact pointer supplied to NewMock, unless that
// pointer was nil in which case a fallback AST carrying the raw capture is
// returned instead.
func (m *MockAdapter) Translate(_ context.Context, raw, _ string) *UIAST {
	if m.Fixed == nil {
		return FallbackAST(raw, "mock:nil")
	}
	return m.Fixed
}
