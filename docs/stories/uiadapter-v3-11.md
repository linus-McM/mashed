# uiadapter-v3-11: CircuitBreaker + Tiered Fallback

**Status:** done
**Landed:** 2026-04-23
**Domain:** backend
**Size:** M
**Depends On:** C, v3-10
**Priority:** P0-critical

## Story

As a Mashed user, I want per-backend circuit breakers plus a richest-to-simplest fallback chain (primary → secondary backend → minimal-kind UIAST → plaintext widget), so that a single backend outage never blocks the UI and recovery time from all-backends-down is <5ms.

## Description

Plan §3 Phase 3 "Story 11 — CircuitBreaker + Tiered Fallback" (lines 476–495). One breaker instance per registered backend via `sony/gobreaker`. Thresholds (via `Config`):
- Open after 3 consecutive failures OR 50% failure rate over a rolling 10s window.
- Half-open probe every 30s using a cached known-good prompt.
- `DEGRADED` sub-state for schema-valid-but-empty outputs (still "success" from the transport's view, but never served to the user).

Tiered fallback (richest → simplest):
1. Primary backend, validated (and repaired if enabled for that backend).
2. Secondary backend per `Config.FallbackOrder`; runs the validate+repair chain again. Common chain: `claude-api → ollama → plaintext`.
3. Minimal-kind UIAST: classifier succeeded but Stage 2 exhausted → emit default widget for that kind.
4. Plaintext widget — last resort, always available.

## Developer Notes

- **Files (new):** `internal/uiadapter/breaker.go`.
- **Files (edited):** `internal/uiadapter/fallback.go` — extend existing plaintext fallback with tiers 1–3.
- **Types / API surface:**
  ```go
  type Breaker interface { Call(ctx, func() (any, error)) (any, error); State() string }
  func NewBreaker(name string, cfg Config) Breaker
  ```
- **Dependencies:** `github.com/sony/gobreaker`.
- **Thresholds:** `Config.BreakerFailThreshold` (default 3), `Config.BreakerResetMs` (default 30000).
- **Risks:** flapping breakers — use rolling window + half-open probe.
- **Use-repo-code directive:** use `use-repo-code` to locate existing `fallback.go` behavior before extending.

## Acceptance Criteria

AC-11.1: `TestBreaker_TripsAfterThree` — after 3 simulated 500s from a backend, next Translate returns via fallback chain in <5ms.

AC-11.2: `TestBreaker_RecoversOnProbe` — breaker returns to closed after one successful half-open probe.

AC-11.3: `breaker_state` (`closed|half_open|open`) attribute per backend on every translate log line.

AC-11.4: `TestBreaker_PerBackendIsolation` — tripping the Claude breaker does not affect the Ollama breaker.

AC-11.5: `TestFallback_TieredRecovery` — primary fail → secondary success returns a valid UIAST with `escalated_from=<primary>` on the log line.

## BDD Test Scenarios

```gherkin
Feature: CircuitBreaker + Tiered Fallback

  Scenario: AC-11.1 — Breaker trips after 3 failures
    Given a backend returning 500 three times
    When a fourth Translate is called
    Then it returns via the fallback chain in under 5ms
    And breaker_state == "open"

  Scenario: AC-11.2 — Probe recovers
    Given an open breaker
    And a successful half-open probe
    When Translate runs next
    Then breaker_state == "closed"

  Scenario: AC-11.3 — Telemetry
    Given any Translate call
    When slog emits
    Then the attribute `breaker_state` per backend is present

  Scenario: AC-11.4 — Per-backend isolation
    Given Claude breaker opens
    When Ollama Translate runs
    Then Ollama's breaker is still closed
    And the Ollama call succeeds without fallback

  Scenario: AC-11.5 — Tiered recovery escalation
    Given primary=claude-api returns 500
    And secondary=ollama is healthy
    When Translate runs
    Then the response is a valid UIAST
    And `escalated_from="claude-api"` appears on the slog line
```

## Tasks / Subtasks

- [ ] Task 1 — Wrap each backend with a breaker (maps to AC-11.1, AC-11.4)
  - [ ] Subtask 1a — One `gobreaker.CircuitBreaker` per registered backend.
  - [ ] Subtask 1b — Rolling 10s window configuration.
- [ ] Task 2 — Half-open probe (maps to AC-11.2)
- [ ] Task 3 — Extend `fallback.go` to 4 tiers (maps to AC-11.5)
  - [ ] Subtask 3a — Secondary backend.
  - [ ] Subtask 3b — Minimal-kind UIAST per Kind.
  - [ ] Subtask 3c — Plaintext last resort.
- [ ] Task 4 — Telemetry `breaker_state`, `escalated_from` (maps to AC-11.3, AC-11.5)
- [ ] Task 5 — Tests — `TestBreaker_TripsAfterThree`, `TestBreaker_RecoversOnProbe`, `TestBreaker_PerBackendIsolation`, `TestFallback_TieredRecovery`

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
