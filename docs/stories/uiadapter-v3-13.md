# uiadapter-v3-13: Eval harness v2 + Shadow mode

**Status:** done
**Landed:** 2026-04-23
**Domain:** backend
**Size:** L
**Depends On:** v3-01, v3-02, v3-05, v3-14, v3-15, v3-16, v3-17
**Priority:** P0-critical

## Story

As a uiadapter maintainer, I want a scorecard with per-backend columns and per-backend thresholds plus 5%-sampled cross-backend shadow mode, so that every plan commitment (parse-rate ≥0.98 Ollama, ≥0.995 Sonnet, latency p50 ≤400ms Haiku, etc.) is proven or falsified automatically and regressions fail CI.

## Description

Plan §3 Phase 4 "Story 13 — Eval harness v2 + Shadow mode" (lines 528–559). Supersedes stale story `uiadapter-05` (eval scorecard).

Scorecard columns per corpus entry: `backend`, `model`, `single_shot`, `fast_path_hit`, `fast_path_rule`, `cache_hit`, `stage1_correct`, `stage2_parse_ok`, `stage2_validation_terminal`, `stage2_untrusted`, `repair_fired`, `repair_succeeded`, `breaker_tripped`, `escalated_from`, `latency_ms_cold`, `latency_ms_warm`, `tokens_in`, `tokens_out`, `cache_read_input_tokens`, `cache_creation_input_tokens`, `usd_cost_est`.

Per-backend thresholds in `MeetsThresholds()`:

| Backend | parse_rate | terminal_validation | p50_warm | p95_warm |
|---|---|---|---|---|
| Ollama | ≥0.98 | ≤0.02 | ≤600ms | ≤1500ms |
| Haiku (cached) | ≥0.99 | ≤0.005 | ≤400ms | ≤1000ms |
| Sonnet (single-shot, cached) | ≥0.995 | ≤0.002 | ≤800ms | ≤2000ms |
| Claude CLI | ≥0.95 | ≤0.02 | ≤2000ms | ≤5000ms |

Corpus: 30 entries, 6 per kind × 5 kinds (yn, menu, form, text, hard/mixed), each `{raw.txt, expected.json}`.

Shadow mode: 5% sample rate (configurable via `Config.ShadowSampleRate`). Each shadowed turn runs Stage 2 with both the current and a designated alternate backend; both ASTs plus a simple delta logged to `_bmad-output/shadow/<pair>.jsonl`. Never served to user. Cost-aware: shadow Claude calls budgeted against accountant.

Build tag: `ollama_eval` preserved — do not run without a local Ollama when the Ollama backend is enabled.

## Developer Notes

- **Files (new/edited):**
  - `internal/uiadapter/eval/score.go` — scorecard + MeetsThresholds.
  - `internal/uiadapter/eval/shadow.go` — shadow runner + delta logger.
  - `internal/uiadapter/eval_test.go` — orchestrator `TestEval_FullCorpus_MeetsThresholds_PerBackend`.
  - `internal/uiadapter/testdata/eval/<kind>/<id>/raw.txt`, `expected.json` — 30 entries.
- **Types / API surface:**
  ```go
  type Score struct { Backend string; Model string; ... /* all columns */ }
  type Thresholds struct { ParseRate, TerminalValidation, P50Warm, P95Warm float64 }
  func (s Scorecard) MeetsThresholds() error
  func Shadow(ctx, primary, alternate Backend, raw string) error
  ```
- **Dependencies:** none new.
- **Risks:** §8 "Shadow mode cost" — budget via accountant; make `ShadowSampleRate=0` honor literal skip.
- **Use-repo-code directive:** use `use-repo-code` to inspect the existing eval corpus from `ui-ast-U9-offline-eval-harness`.

## Acceptance Criteria

AC-13.1: scorecard pretty-prints all metrics, grouped by backend.

AC-13.2: thresholds fail-closed with a useful error message when violated.

AC-13.3: shadow JSONL is written, diffable offline, and includes both ASTs plus metadata.

AC-13.4: cross-backend scorecard prints a diff table (backend × metric).

## BDD Test Scenarios

```gherkin
Feature: Eval harness v2 + Shadow mode

  Scenario: AC-13.1 — Pretty-print grouped by backend
    Given the eval corpus runs on Ollama, Haiku, Sonnet, CLI
    When the scorecard prints
    Then the output has one section per backend
    And every column is present

  Scenario: AC-13.2 — Fail-closed thresholds
    Given a scorecard where Ollama parse_rate == 0.97
    When MeetsThresholds() is called
    Then it returns an error naming the backend, metric, threshold, actual

  Scenario: AC-13.3 — Shadow JSONL output
    Given ShadowSampleRate=1.0 for a single test run
    When Translate runs once on each corpus entry
    Then `_bmad-output/shadow/<pair>.jsonl` contains one row per entry
    And each row has both ASTs plus metadata (timestamp, kind, delta)

  Scenario: AC-13.4 — Cross-backend diff table
    Given scorecards from all four backends
    When the diff table prints
    Then rows are metrics and columns are backends
    And threshold violations are highlighted
```

## Tasks / Subtasks

- [ ] Task 1 — Scorecard data model (maps to AC-13.1, AC-13.4)
- [ ] Task 2 — Thresholds + fail-closed (maps to AC-13.2)
- [ ] Task 3 — Shadow runner + delta logger (maps to AC-13.3)
- [ ] Task 4 — 30-entry corpus under `testdata/eval/` (maps to AC-13.1, AC-13.2)
- [ ] Task 5 — `TestEval_FullCorpus_MeetsThresholds_PerBackend` (maps to AC-13.1–13.4)
- [ ] Task 6 — Shadow-mode fast-path sampler (maps to AC-2.4 coordination)
- [ ] Task 7 — Build-tag `ollama_eval` discipline

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
