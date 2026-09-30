# uiadapter-v3-11b: Cost + rate-limit accountant

**Status:** done
**Landed:** 2026-04-23
**Domain:** backend
**Size:** M
**Depends On:** v3-11, v3-14 (partial — accountant lives in claudeapi pkg)
**Priority:** P1-high

## Story

As a Mashed user, I want a Claude-specific accountant that tracks RPM / TPM / USD and preemptively trips the breaker at 85% of soft limits, so that a naive translate loop can't blow through a monthly budget in minutes and 429s are respected before they happen.

## Description

Plan §3 Phase 3 "Story 11b — Cost + rate-limit accountant" (lines 497–512). New failure mode unique to Claude. Sliding-window counters per minute for RPM and TPM. Running `usd_spent_session` total. Each Claude response's `usage.input_tokens`, `usage.output_tokens`, `cache_creation_input_tokens`, `cache_read_input_tokens` feed the accountant. Pricing table hardcoded for v3.0 (Haiku, Sonnet); externalizable later. Soft limits at 85% of `RPMSoftLimit`, `TPMSoftLimit`, `UsdBudgetPerSession` trigger a Warn log. Hard limit or 429/529 trips the Claude breaker open proactively with exponential backoff (respecting `Retry-After`).

## Developer Notes

- **Files (new):** `internal/uiadapter/backend/claudeapi/accountant.go`, `internal/uiadapter/backend/claudeapi/accountant_test.go`.
- **Types / API surface:**
  ```go
  type Accountant struct { RPMLimit, TPMLimit int; UsdBudget float64; ... }
  func (a *Accountant) Observe(resp Usage) error   // returns ErrBudgetExceeded / ErrRateLimited
  func (a *Accountant) AllowCall() error           // pre-call gate
  ```
- **Pricing table:** per-model USD-per-million tokens, checked into source at release time (Haiku, Sonnet, Opus).
- **Retry-After:** parse header on 429; wait then retry once; failure surfaces `ErrRateLimited`.
- **Coordination with Story 11:** accountant trips the Claude breaker when limits cross; breaker uses existing `Call` wrapper.
- **Telemetry:** `tokens_in`, `tokens_out`, `cache_read_input_tokens`, `cache_creation_input_tokens`, `usd_cost_est` on every Claude Translate line (§6.1).
- **Risks:** §8 "Rate-limit surprise costs" and "Shadow mode cost" — accountant is the mitigation for both.
- **Dependencies:** none new (sliding window via ring buffer).
- **Use-repo-code directive:** use `use-repo-code` to confirm no existing cost-tracking scaffolding to avoid double-bookkeeping.

## Acceptance Criteria

AC-11b.1: `TestAccountant_TripsBeforeHard429` — simulate 100 calls just under the soft limit; breaker stays closed. At soft limit, breaker opens preemptively.

AC-11b.2: `TestAccountant_CostMatchesBilling` — cost accumulator matches Anthropic's billing reference calculation within ±1% on the corpus.

AC-11b.3: `tokens_in`, `tokens_out`, `cache_read_input_tokens`, `cache_creation_input_tokens`, `usd_cost_est` attributes on every translate log line when Claude is the backend.

AC-11b.4: a 429 with `Retry-After: 2` is respected on retry; failure after retry surfaces `ErrRateLimited`.

## BDD Test Scenarios

```gherkin
Feature: Cost + rate-limit accountant

  Scenario: AC-11b.1 — Preemptive trip
    Given RPMSoftLimit=60 and 50 observed calls in 60s
    When the 51st call is made
    Then breaker remains closed (<85%)
    Given 51 observed calls
    When the 52nd call is made
    Then breaker opens preemptively at 85% threshold

  Scenario: AC-11b.2 — Cost parity ±1%
    Given a corpus of Claude responses with known token counts
    When the accountant computes cost
    Then the total is within ±1% of the billing reference

  Scenario: AC-11b.3 — Telemetry fields
    Given a Claude API translate
    When slog emits
    Then `tokens_in`, `tokens_out`, `cache_read_input_tokens`, `cache_creation_input_tokens`, `usd_cost_est` are present

  Scenario: AC-11b.4 — Retry-After honored
    Given a 429 response with Retry-After: 2
    When the accountant retries
    Then the retry waits ≥ 2 seconds
    And on second failure returns ErrRateLimited
```

## Tasks / Subtasks

- [ ] Task 1 — Sliding-window counters (maps to AC-11b.1)
- [ ] Task 2 — Pricing table + cost math (maps to AC-11b.2, AC-11b.3)
- [ ] Task 3 — Preemptive breaker trip (maps to AC-11b.1)
- [ ] Task 4 — Retry-After handling (maps to AC-11b.4)
- [ ] Task 5 — Telemetry attrs (maps to AC-11b.3)
- [ ] Task 6 — Tests (maps to AC-11b.1–11b.4)

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
