# uiadapter-v3-16: Backend Router + Policy

**Status:** ready
**Domain:** backend
**Size:** L
**Depends On:** C, v3-11, v3-14, v3-15
**Priority:** P0-critical

## Story

As a Mashed user, I want a router that picks a backend per request based on a policy (`claude-first`, `local-only`, `privacy-strict`, etc.) and orchestrates cross-backend fallback, so that one config knob swaps my reliability/cost/privacy profile without touching call sites.

## Description

Plan §3 Phase 4 "Story 16 — Backend Router + Policy" (lines 629–651). Policies:

| Policy | Primary | Escalation | Fallback order |
|---|---|---|---|
| `local-only` | Ollama | — | plaintext |
| `claude-only` | Claude Haiku | Claude Sonnet on repair-fail | plaintext |
| `claude-first` | Claude Haiku | Claude Sonnet | Ollama → plaintext |
| `ollama-first` | Ollama | Claude Haiku | plaintext |
| `cost-aware` | Ollama when healthy; else Claude | Sonnet on validation-terminal | plaintext |
| `privacy-strict` | Ollama only; refuse Claude if raw matches `Config.PrivacyPatterns` | — | plaintext |

Escalation triggers: primary validation-terminal OR breaker open OR (when available) low-confidence. Concurrency: router cooperates with singleflight (Story 4) — a second concurrent identical request does not independently escalate. Default policy: `claude-first` on hosts with an Anthropic key, else `local-only`.

Privacy patterns (defaults): AWS access-key (`AKIA[0-9A-Z]{16}`), GitHub token (`ghp_[A-Za-z0-9]{36,}`), generic `sk-[A-Za-z0-9]{20,}`, email addresses, generic JWT shape.

## Developer Notes

- **Files (new):**
  - `internal/uiadapter/backend/router.go`
  - `internal/uiadapter/backend/router_test.go`
- **Types / API surface:**
  ```go
  type Router struct { policy string; primaries map[string]LLMBackend; fallbackOrder []string; privacyPatterns []*regexp.Regexp }
  func NewRouter(cfg Config) (*Router, error)
  func (r *Router) Translate(ctx, raw string) (*UIAST, RouterDecision, error)
  type RouterDecision struct { Primary, Fallback, Escalation string }
  ```
- **Telemetry:** `router_decision={"primary":..., "fallback":..., "escalation":...}` JSON attribute on every slog line.
- **Risks:** policy matrix breaks on config defaults drift — AC-16.1 is a 6×4 matrix test.
- **Dependencies:** `errors.Join` for multi-backend failure (§6.5).
- **Use-repo-code directive:** use `use-repo-code` to audit any `Config.PrivacyPatterns` defaults already present.

## Acceptance Criteria

AC-16.1: `TestRouter_CoversEveryPolicy` — matrix test over all 6 policies × {all healthy, Ollama down, Claude rate-limited, all down}.

AC-16.2: `TestRouter_PrivacyStrictBlocksClaude` — synthetic AWS-key-shaped raw triggers privacy-strict to use Ollama only.

AC-16.3: `router_decision={"primary":..., "fallback":..., "escalation":...}` attribute on every translate log line.

## BDD Test Scenarios

```gherkin
Feature: Backend Router + Policy

  Scenario: AC-16.1 — Policy × health matrix
    Given one of 6 policies × 4 health states (24 cells)
    When the router selects a backend
    Then every cell returns the plan-documented primary + fallback
    And no cell panics or returns nil without error

  Scenario: AC-16.2 — privacy-strict blocks Claude
    Given RouterPolicy=privacy-strict
    And raw contains "AKIAABCDEFGHIJKLMNOP"
    When the router picks a backend
    Then the Claude backend is not invoked
    And Ollama is the primary
    And the decision log records the blocked intent

  Scenario: AC-16.3 — router_decision telemetry
    Given any Translate call
    When slog emits
    Then the `router_decision` JSON attribute is present with primary/fallback/escalation
```

## Tasks / Subtasks

- [ ] Task 1 — Router skeleton + policy enum (maps to AC-16.1)
- [ ] Task 2 — Primary selection per policy + health state (maps to AC-16.1)
- [ ] Task 3 — Privacy-strict regex gate (maps to AC-16.2)
- [ ] Task 4 — Escalation triggers (validation-terminal, breaker-open) (maps to AC-16.1)
- [ ] Task 5 — `router_decision` JSON telemetry (maps to AC-16.3)
- [ ] Task 6 — Singleflight-compatible dispatch (coordinate with Story 4)
- [ ] Task 7 — Tests (maps to AC-16.1–16.3)

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
