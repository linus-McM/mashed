// Package backend is the LLM-dispatch boundary for the v3 uiadapter
// pipeline. One interface, one registry; every concrete backend lives in
// its own subpackage and is discovered via init-time registration, so the
// router, breaker, and cache stay backend-agnostic (Plan §2).
package backend

import (
	"context"
	"errors"

	"mashed/internal/uiadapter"
)

// Kind is the stage-1 classification. Two-stage pipelines emit one of
// these values from Classify; single-shot pipelines skip Classify and go
// directly through GenerateSingleShot.
type Kind string

const (
	KindYN   Kind = "yn"
	KindMenu Kind = "menu"
	KindForm Kind = "form"
	KindText Kind = "text"
)

// LLMBackend abstracts the translate-stage call(s) to a specific provider.
// Implementations MUST be safe for concurrent use. The interface is frozen
// at this story — later stories extend via narrower optional interfaces at
// the consumer rather than widening this contract (Plan §6.5 "Interfaces
// at the consumer").
type LLMBackend interface {
	Name() string
	Classify(ctx context.Context, raw string) (Kind, error)
	Generate(ctx context.Context, raw string, kind Kind) (*uiadapter.UIAST, error)
	GenerateSingleShot(ctx context.Context, raw string) (*uiadapter.UIAST, error)
	WarmUp(ctx context.Context) error
	Health(ctx context.Context) error
	Capabilities() Capabilities
}

// Capabilities is advisory metadata the router consults before dispatching.
// Provider + Model are strings so logging / shadow-mode scorecards can group
// rows without reflection. TokenCostUSDPerMil is zero for Ollama and feeds
// the cost accountant (Story v3-11b) for Claude backends.
type Capabilities struct {
	Provider            string  // "ollama" | "anthropic-api" | "claude-cli"
	Model               string
	MaxContextTokens    int
	SupportsSingleShot  bool    // true for Sonnet, false for gemma3:4b
	SupportsPromptCache bool    // true for Claude explicit, true for Ollama implicit
	SupportsSeed        bool    // true for Ollama, false for Claude
	TokenCostUSDPerMil  float64 // 0 for Ollama; used by the cost accountant
	IsLocal             bool    // true for Ollama; gates privacy-sensitive routing
}

// Sentinel errors backends surface. Consumers match via errors.Is.
var (
	ErrSingleShotUnsupported = errors.New("backend: single-shot not supported")
	ErrBackendUnreachable    = errors.New("backend: unreachable")
	ErrRateLimited           = errors.New("backend: rate limited")
	ErrBudgetExceeded        = errors.New("backend: cost budget exceeded")
	ErrUnknownBackend        = errors.New("backend: unknown name")
)
