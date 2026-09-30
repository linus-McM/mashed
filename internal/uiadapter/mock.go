//go:build testing

package uiadapter

import (
	"context"
	"log/slog"
)

// op-attr constants for mock.* records.
const (
	mockInitOp      = "mock.init"
	mockTranslateOp = "mock.translate"
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
//
// Story 5: emits mock.init with op, enabled=true, fixture_name derived
// from fixed.GeneratedBy (or "unnamed" when empty / fixed is nil).
func NewMock(fixed *UIAST, logger *slog.Logger) Adapter {
	lg := nilSafeLogger(logger)
	if lg.Enabled(context.Background(), slog.LevelDebug) {
		lg.LogAttrs(context.Background(), slog.LevelDebug, "mock.init",
			slog.String("op", mockInitOp),
			slog.Bool("enabled", true),
			slog.String("fixture_name", fixtureName(fixed)),
		)
	}
	return &MockAdapter{Fixed: fixed, logger: lg}
}

// fixtureName derives the mock-init fixture identifier. Empty / nil GeneratedBy
// falls back to "unnamed" so the attr is always populated.
func fixtureName(fixed *UIAST) string {
	if fixed == nil || fixed.GeneratedBy == "" {
		return "unnamed"
	}
	return fixed.GeneratedBy
}

// Translate returns the exact pointer supplied to NewMock, unless that
// pointer was nil in which case a fallback AST carrying the raw capture is
// returned instead.
//
// Story 5: emits mock.translate with op, proc_id, bytes_in. The raw payload
// is never logged.
func (m *MockAdapter) Translate(_ context.Context, raw, procID string) *UIAST {
	if m.logger.Enabled(context.Background(), slog.LevelDebug) {
		m.logger.LogAttrs(context.Background(), slog.LevelDebug, "mock.translate",
			slog.String("op", mockTranslateOp),
			slog.String("proc_id", procID),
			slog.Int("bytes_in", len(raw)),
		)
	}
	if m.Fixed == nil {
		return FallbackAST(raw, "mock:nil", m.logger)
	}
	return m.Fixed
}
