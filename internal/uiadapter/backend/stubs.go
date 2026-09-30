package backend

import (
	"context"
	"sync/atomic"

	"mashed/internal/uiadapter"
)

// StubBackend is a test-friendly LLMBackend used by the interface
// concurrency stress test (AC-C.1) and as a placeholder until Stories
// v3-14/v3-15 ship real Claude backends. Every method increments an
// atomic counter so tests can verify call counts without locks.
type StubBackend struct {
	name         string
	capabilities Capabilities

	classifyCalls   atomic.Int64
	generateCalls   atomic.Int64
	singleShotCalls atomic.Int64
	warmUpCalls     atomic.Int64
	healthCalls     atomic.Int64
}

// NewStub constructs a named stub backend. Exported so each provider
// subpackage can wrap its own Stub with provider-specific Capabilities
// before a real implementation lands.
func NewStub(name string, caps Capabilities) *StubBackend {
	return &StubBackend{name: name, capabilities: caps}
}

func (s *StubBackend) Name() string { return s.name }

func (s *StubBackend) Classify(_ context.Context, _ string) (Kind, error) {
	s.classifyCalls.Add(1)
	return KindText, nil
}

func (s *StubBackend) Generate(_ context.Context, raw string, kind Kind) (*uiadapter.UIAST, error) {
	s.generateCalls.Add(1)
	return synthUIAST(s.name, raw, string(kind)), nil
}

func (s *StubBackend) GenerateSingleShot(_ context.Context, raw string) (*uiadapter.UIAST, error) {
	s.singleShotCalls.Add(1)
	if !s.capabilities.SupportsSingleShot {
		return nil, ErrSingleShotUnsupported
	}
	return synthUIAST(s.name, raw, "single-shot"), nil
}

func (s *StubBackend) WarmUp(_ context.Context) error {
	s.warmUpCalls.Add(1)
	return nil
}

func (s *StubBackend) Health(_ context.Context) error {
	s.healthCalls.Add(1)
	return nil
}

func (s *StubBackend) Capabilities() Capabilities { return s.capabilities }

// Calls returns a snapshot tuple (classify, generate, singleShot, warmup,
// health). Exposed for AC-C.1's counter parity assertion.
func (s *StubBackend) Calls() (int64, int64, int64, int64, int64) {
	return s.classifyCalls.Load(),
		s.generateCalls.Load(),
		s.singleShotCalls.Load(),
		s.warmUpCalls.Load(),
		s.healthCalls.Load()
}

// synthUIAST produces a minimal UIAST envelope stamped with the backend's
// name. Used by both Generate and GenerateSingleShot stubs so the
// call-routing contract stays visible in test output.
func synthUIAST(backendName, raw, label string) *uiadapter.UIAST {
	return &uiadapter.UIAST{
		Version:     "1",
		GeneratedBy: "stub:" + backendName + ":" + label,
		TurnSummary: label,
		Nodes: []uiadapter.UINode{{
			Type:    "markdown",
			Content: raw,
		}},
		FallbackAnswerShape: "free",
	}
}
