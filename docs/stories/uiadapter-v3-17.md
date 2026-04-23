# uiadapter-v3-17: Per-backend WarmUp and Health

**Status:** done
**Landed:** 2026-04-23
**Domain:** backend
**Size:** M
**Depends On:** C, v3-14, v3-15
**Priority:** P0-critical

## Story

As a Mashed user, I want all enabled backends warmed up in parallel at adapter boot (non-blocking) and health-pinged every 30s, so that the first Translate call doesn't pay cold-start cost and the router can route around unreachable backends without waiting for a timeout.

## Description

Plan §3 Phase 4 "Story 17 — Per-backend WarmUp and Health" (lines 653–662). WarmUp: adapter constructor kicks off `WarmUp` for each enabled backend in parallel goroutines, each with `Config.WarmUpTimeoutMs` timeout. Errors recorded but non-fatal. `sync.Once` per backend so repeated ticker calls don't re-warm. Health: periodic ticker (default 30s, disable via `Config.DisableHealthTicker`) calls `Health` per backend. Results feed the router's live-backends set and the breaker's half-open probe.

## Developer Notes

- **Files (new):**
  - `internal/uiadapter/backend/lifecycle.go`
  - `internal/uiadapter/backend/lifecycle_test.go`
- **Types / API surface:**
  ```go
  type Lifecycle struct { backends []LLMBackend; warmed map[string]*sync.Once; health atomic.Value /* map[string]bool */ }
  func (l *Lifecycle) Start(ctx context.Context)
  func (l *Lifecycle) Stop()
  func (l *Lifecycle) IsHealthy(name string) bool
  ```
- **Risks:** goroutine leak on adapter shutdown — `Stop()` cancels the ticker ctx; tests assert no goroutines leak (`goleak`).
- **Dependencies:** optional `go.uber.org/goleak` for test-only leak checks.
- **Use-repo-code directive:** use `use-repo-code` to inspect existing reachability stores (`ollamaReachable`) that this story's Health ticker will feed via Wails events in Story 18.

## Acceptance Criteria

AC-17.1: `TestAdapter_WarmupDoesNotBlock` — adapter constructor returns in <50ms even when all backends are unreachable (extends Story 7 AC).

AC-17.2: `TestLifecycle_TickerDisableable` — with `DisableHealthTicker=true`, no Health calls fire in tests.

AC-17.3: `TestLifecycle_WarmUpAllInParallel` — three backends' WarmUp all run in parallel; total elapsed time ≈ `max(WarmUp durations)`, not the sum.

## BDD Test Scenarios

```gherkin
Feature: Per-backend WarmUp and Health

  Scenario: AC-17.1 — Non-blocking construction with all unreachable
    Given all three backends return errors from WarmUp
    When NewDefault(cfg) runs
    Then it returns in under 50ms

  Scenario: AC-17.2 — Ticker disable
    Given DisableHealthTicker=true
    When the lifecycle runs for 5s
    Then zero Health calls are recorded

  Scenario: AC-17.3 — Parallel WarmUp
    Given three backends with WarmUp durations 1s, 2s, 3s
    When WarmUp is kicked off by the constructor
    Then total elapsed time is ≈ 3s (±200ms), not 6s

  Scenario: No goroutine leak
    Given a lifecycle starts and Stop() is called
    When the test ends
    Then goleak reports no leaked goroutines
```

## Tasks / Subtasks

- [ ] Task 1 — Parallel WarmUp with per-backend timeout + sync.Once (maps to AC-17.1, AC-17.3)
- [ ] Task 2 — Health ticker + atomic map (maps to AC-17.2)
- [ ] Task 3 — Disable-ticker flag (maps to AC-17.2)
- [ ] Task 4 — Router integration (feed `IsHealthy`) — coordinate with Story 16
- [ ] Task 5 — Tests (maps to AC-17.1, AC-17.2, AC-17.3) + goleak

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
