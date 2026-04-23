# uiadapter-C: LLMBackend interface + registry

**Status:** done
**Domain:** backend
**Size:** M
**Depends On:** B
**Priority:** P0-critical
**Landed:** 2026-04-23

## Story

As a uiadapter architect, I want a single `LLMBackend` interface plus a `backend.From(name, cfg)` registry, so that every backend (Ollama / Claude API / Claude CLI) is interchangeable behind one boundary and the router, breaker, and cache can remain backend-agnostic.

## Description

Plan §2 "The `LLMBackend` interface" (lines 117–168) and §3 Phase 0 "Story C — `LLMBackend` interface + registry" (lines 253–261). Introduces the interface (`Name`, `Classify`, `Generate`, `GenerateSingleShot`, `WarmUp`, `Health`, `Capabilities`), the `Capabilities` struct, the sentinel errors (`ErrSingleShotUnsupported`, `ErrBackendUnreachable`, `ErrRateLimited`, `ErrBudgetExceeded`), and the registry. Each concrete implementation lives in its own subpackage (`backend/ollama`, `backend/claudeapi`, `backend/claudecli`) and calls `init() → backend.Register(name, ctor)` to avoid import cycles.

Consequences downstream stories rely on (plan §2 lines 163–168):
- Router (Story 16) reads `Capabilities()` to decide single-shot vs. two-stage.
- Breaker (Story 11) wraps each backend independently — Claude outage leaves Ollama closed.
- Cache (Story 4) keys on `sha256(backend_name + model + raw)` so outputs never collide.

## Developer Notes

- **Files (new):**
  - `internal/uiadapter/backend/backend.go` (interface + Capabilities + sentinels)
  - `internal/uiadapter/backend/registry.go` (`Register`, `From`, available-names listing)
  - stub subpackages `internal/uiadapter/backend/ollama/`, `.../claudeapi/`, `.../claudecli/` (just enough to satisfy the interface and the concurrency stress test; real implementations are Stories 14, 15, and Ollama-specific touches)
- **Types / API surface:**
  ```go
  type Kind string
  const (KindYN, KindMenu, KindForm, KindText Kind = "yn", "menu", "form", "text")
  type LLMBackend interface { Name() string; Classify(ctx, raw) (Kind, error); Generate(ctx, raw, kind) (*UIAST, error); GenerateSingleShot(ctx, raw) (*UIAST, error); WarmUp(ctx) error; Health(ctx) error; Capabilities() Capabilities }
  type Capabilities struct { Provider, Model string; MaxContextTokens int; SupportsSingleShot, SupportsPromptCache, SupportsSeed bool; TokenCostUSDPerMil float64; IsLocal bool }
  var (ErrSingleShotUnsupported, ErrBackendUnreachable, ErrRateLimited, ErrBudgetExceeded error)
  ```
- **Risks:** interface churn breaks later stories. Freeze shape in this story; later stories extend via new optional methods on narrower interfaces at the consumer (§6.5 "Interfaces at the consumer").
- **Dependencies:** none new.
- **Use-repo-code directive:** use `use-repo-code` to read the current Ollama client (`internal/uiadapter/client.go` if present) when shaping the Ollama stub.

## Acceptance Criteria

AC-C.1: three stub implementations pass the concurrency stress test (`TestBackend_InterfaceStressConcurrent`) — N=100 goroutines, 1000 ops each, no data races.

AC-C.2: `backend.From` surfaces a useful error for unknown names.

## BDD Test Scenarios

```gherkin
Feature: LLMBackend interface + registry

  Scenario: AC-C.1 — Concurrency stress
    Given three registered stub backends ("ollama", "claude-api", "claude-cli")
    And 100 goroutines each performing 1000 Classify/Generate/Health calls
    When the run finishes under `go test -race`
    Then no data race is reported
    And every backend's counters match the total call count

  Scenario: AC-C.2 — Unknown name surfaces available options
    Given Config{Backend:"nonexistent"}
    When backend.From("nonexistent", cfg) is called
    Then an error is returned whose message lists the available names ("ollama", "claude-api", "claude-cli")
    And the error wraps a stable sentinel or typed error the caller can match
```

## Tasks / Subtasks

- [ ] Task 1 — Define interface and sentinels (maps to AC-C.1, AC-C.2)
  - [ ] Subtask 1a — Write `backend/backend.go` with interface + Capabilities + errors.
  - [ ] Subtask 1b — Define `Kind` enum.
- [ ] Task 2 — Registry (maps to AC-C.2)
  - [ ] Subtask 2a — `Register(name, func(Config) (LLMBackend, error))` with mutex-guarded map.
  - [ ] Subtask 2b — `From(name, cfg)` with unknown-name error listing sorted keys.
- [ ] Task 3 — Stub implementations (maps to AC-C.1)
  - [ ] Subtask 3a — `backend/ollama/stub.go`, `backend/claudeapi/stub.go`, `backend/claudecli/stub.go`, each registering via `init()`.
  - [ ] Subtask 3b — Stubs return synthetic UIASTs + zero-valued Capabilities appropriate to the backend.
- [ ] Task 4 — Write `TestBackend_InterfaceStressConcurrent` (maps to AC-C.1)

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
