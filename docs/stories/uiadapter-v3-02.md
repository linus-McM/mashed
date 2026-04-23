# uiadapter-v3-02: FastPathClassifier

**Status:** ready
**Domain:** backend
**Size:** M
**Depends On:** B, v3-01
**Priority:** P0-critical

## Story

As a Mashed user, I want common Claude-Code turns (Y/N prompts, numbered menus, "Press Enter", fenced confirm-diff, single-line free-text questions) to bypass the LLM entirely via a rule table, so that 40–70% of translates happen in <1ms p95 with zero failure surface and zero token cost on Claude backends.

## Description

Plan §3 Phase 1 "Story 2 — FastPathClassifier" (lines 283–314). Claude-Code output is repetitive; a short ordered regex table translates a large share of turns without calling any LLM. On a match the rule emits a UIAST tagged `generated_by="fastpath:<rule>"` and skips every downstream stage including cache. Story 13's shadow mode samples 10% of fast-path hits through the full pipeline to flag disagreement — this is the guard against incorrectly bypassing a richer widget (§8 risk).

## Developer Notes

- **Files (new):** `internal/uiadapter/fastpath.go`, `internal/uiadapter/fastpath_test.go`.
- **Types / API surface:**
  ```go
  type Rule struct {
      Name  string
      Match *regexp.Regexp
      Build func(raw string, m []string) *UIAST
  }
  func Classify(raw string) (*UIAST, string /*ruleName*/, bool /*hit*/)
  ```
- **Minimum rule set (v3.0):**
  - `yn-prompt` — `(?i)\b(y/n|\[Y/n\]|\[y/N\]|\(yes/no\))\b` → select, 2 options
  - `numbered-menu` — `^\s*\d+[\.\)]\s+\S` matched on ≥2 consecutive lines → select with numeric values
  - `press-enter` — `(?i)press (any key|enter) to (continue|exit)` → acknowledge button
  - `file-confirm` — `^\s*(apply|write|save) .*?\?` → confirm, 2 options
  - `free-text-prompt` — single line ending `?`, no list structure → textarea
- **Telemetry:** per-rule counter; emit `fast_path_hit_rate`, `fast_path_rule` on every slog line (§6.1).
- **Gated by `Config.EnableFastPath` (default true)** — Story B.
- **Risks:** §8 "Story 2 (fast-path) can incorrectly bypass the LLM" — mitigated by Story 13 shadow 10% of hits.
- **Dependencies:** `regexp` stdlib only.
- **Use-repo-code directive:** use `use-repo-code` to understand the current UIAST shape before authoring `Build` funcs.

## Acceptance Criteria

AC-2.1: every trivially-structured entry in the eval corpus (Story 13) is caught by a rule.

AC-2.2: `fast_path_hit_rate` and `fast_path_rule` attributes on every structured log line.

AC-2.3: fast-path match → UIAST in <1ms p95 (measured in a micro-benchmark).

AC-2.4: Story 13's shadow mode runs 10% of fast-path hits through the full LLM pipeline too and flags disagreement in the scorecard.

## BDD Test Scenarios

```gherkin
Feature: FastPathClassifier

  Scenario: AC-2.1 — Trivial corpus entries match a rule
    Given the eval corpus entries labelled "trivial" (yn, numbered-menu, press-enter, file-confirm, free-text)
    When Classify runs on each
    Then every entry returns hit=true with a non-empty ruleName
    And the emitted UIAST passes schema validation

  Scenario: AC-2.2 — Telemetry attributes present
    Given any translate call
    When slog output is captured
    Then `fast_path_hit_rate` and `fast_path_rule` (when hit) are attributes on the line

  Scenario: AC-2.3 — Sub-millisecond p95
    Given 10000 iterations of Classify on the fastpath corpus in BenchmarkFastPath
    When the benchmark runs
    Then the p95 per-call latency is under 1ms

  Scenario: AC-2.4 — Shadow flags disagreement (Story 13 hook)
    Given EnableFastPath=true and ShadowSampleRate=0.1
    When a fast-path-hit turn is sampled
    Then both the fast-path UIAST and the LLM-pipeline UIAST are written to `_bmad-output/shadow/`
    And disagreements appear in the scorecard delta column
```

## Tasks / Subtasks

- [ ] Task 1 — Define Rule and rule table (maps to AC-2.1, AC-2.3)
  - [ ] Subtask 1a — Compile regexes at init.
  - [ ] Subtask 1b — Author the five v3.0 rules + their Build funcs.
- [ ] Task 2 — Implement `Classify` (maps to AC-2.1)
  - [ ] Subtask 2a — First-match-wins iteration.
  - [ ] Subtask 2b — Emit `generated_by="fastpath:<rule>"` on the UIAST.
- [ ] Task 3 — Per-rule counter + slog attributes (maps to AC-2.2)
- [ ] Task 4 — `BenchmarkFastPath` (maps to AC-2.3)
- [ ] Task 5 — Shadow hook stub (maps to AC-2.4) — full wiring lives in Story 13

## Definition of Done

- [ ] All ACs verified with PASS evidence (test name or benchmark)
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
