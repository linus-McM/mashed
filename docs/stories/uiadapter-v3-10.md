# uiadapter-v3-10: Self-repair loop (bounded, gated)

**Status:** ready
**Domain:** backend
**Size:** M
**Depends On:** v3-05, v3-06
**Priority:** P1-high

## Story

As a uiadapter author, I want a bounded validator-error-feedback repair retry that fires only on semantic failure (wrong enum, slugification drift, missing required field) and is gated per-backend (1 on Ollama/Haiku, 0 on Sonnet), so that cheap recoveries happen in-line and expensive ones escalate to backend fallback instead.

## Description

Plan §3 Phase 3 "Story 10 — Self-repair loop (bounded, gated)" (lines 451–474). Schema-constrained decoding prevents structural failure; semantic failures still happen. One retry with validator-error feedback catches most. Retry cost is not justified on Sonnet — fall back to another backend or minimal-kind UIAST instead. Repair never fires when Stage 2 passes (no wasted calls — AC-10.3).

Repair prompt shape:
```
[static_prefix]
RAW CAPTURE: <original sanitized raw>
KIND: <classifier kind>
PREVIOUS ATTEMPT (invalid):
<previous bad output>
VALIDATION ERRORS:
<jsonschema error list, one per line>

Emit a corrected UIAST that addresses each error. Do not repeat the previous errors.
```

## Developer Notes

- **Files (new):** `internal/uiadapter/repair.go`, `internal/uiadapter/repair_test.go`.
- **Types / API surface:** `func Repair(ctx, backend, static string, raw string, kind Kind, previous *UIAST, errors []string) (*UIAST, error)`.
- **Per-backend retry budget:**
  - Ollama / Haiku: `RepairMaxRetries=1`. Second failure → fallback tier.
  - Sonnet: `RepairMaxRetries=0`. Retry cost not justified.
- **Telemetry:** `repair_count` per-translate metric (§6.1). Never exceeds `RepairMaxRetries`.
- **Risks:** wasted calls on success path — AC-10.3 enforces zero-fire on Stage-2 pass.
- **Dependencies:** none new.
- **Use-repo-code directive:** use `use-repo-code` to locate existing validator error-list shape.

## Acceptance Criteria

AC-10.1: `TestRepair_RecoveryRate` — on a synthetic bad-output corpus, recovery rate ≥0.6 with one retry (Ollama).

AC-10.2: `repair_count` per-translate metric; never exceeds `RepairMaxRetries`.

AC-10.3: repair never fires when stage-2 passes (no wasted calls).

AC-10.4: `TestRepair_Gated_PerBackend` — Sonnet translate with bad output triggers backend fallback, not repair.

## BDD Test Scenarios

```gherkin
Feature: Self-repair loop

  Scenario: AC-10.1 — Recovery ≥0.6 on Ollama
    Given a 50-entry synthetic bad-output corpus
    When Repair runs with RepairMaxRetries=1 on Ollama
    Then ≥30 entries recover to a valid UIAST

  Scenario: AC-10.2 — Retry budget respected
    Given RepairMaxRetries=1 and a permanently-broken generator
    When Translate runs on a bad Stage-2
    Then repair_count on the slog line is exactly 1

  Scenario: AC-10.3 — No fire on success
    Given a clean Stage-2 result
    When Translate runs
    Then repair_count is 0

  Scenario: AC-10.4 — Sonnet gated out
    Given Backend=claude-sonnet and a bad Stage-2 output
    When Translate runs
    Then repair is not invoked
    And the router escalates to the next backend
```

## Tasks / Subtasks

- [ ] Task 1 — Implement `Repair` (maps to AC-10.1, AC-10.3)
  - [ ] Subtask 1a — Build repair prompt per shape above.
  - [ ] Subtask 1b — Single retry per call site.
- [ ] Task 2 — Per-backend gate (maps to AC-10.2, AC-10.4)
  - [ ] Subtask 2a — Read `Config.RepairMaxRetries` resolved per backend.
- [ ] Task 3 — Tests (maps to AC-10.1–10.4)
  - [ ] Subtask 3a — Synthetic bad-output corpus under `testdata/repair/`.
- [ ] Task 4 — Telemetry: `repair_count` slog attr (maps to AC-10.2)

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
