# uiadapter-v3-09: Deterministic sampling

**Status:** done
**Landed:** 2026-04-23
**Domain:** backend
**Size:** S
**Depends On:** B, C
**Priority:** P1-high

## Story

As a uiadapter maintainer, I want per-backend deterministic sampling defaults (seed=42 + temperature=0 on Ollama; temperature=0 on Claude because no seed is exposed), so that cache hit rates tighten, golden fixtures stay stable, and eval variance drops — with per-backend tolerances because Claude is "reduced" not "guaranteed" deterministic.

## Description

Plan §3 Phase 2 "Story 9 — Deterministic sampling" (lines 437–447). Ollama defaults: `temperature: 0, seed: 42, top_k: 1, top_p: 1.0`. Caveat — Ollama's determinism is imperfect (issues [#586](https://github.com/ollama/ollama/issues/586), [#1749](https://github.com/ollama/ollama/issues/1749), [#5321](https://github.com/ollama/ollama/issues/5321)); first-run drift is real. Claude defaults: `temperature: 0`. No seed available — determinism is not guaranteed across deployments, only reduced. AC-9.1 reflects the asymmetry: Ollama ≥0.95 byte-identical; Claude ≥0.80.

## Developer Notes

- **Files (edited):**
  - `internal/uiadapter/config.go` — `Temperature float32` (default 0), `Seed int64` (default 42).
  - `internal/uiadapter/backend/ollama/client.go` — send seed + temperature.
  - `internal/uiadapter/backend/claudeapi/client.go` — send temperature only.
- **Risks:** §8 "Determinism regression across backends" — per-backend thresholds are the guard.
- **Dependencies:** none new.
- **Use-repo-code directive:** use `use-repo-code` to locate current sampling parameters in the Ollama client.

## Acceptance Criteria

AC-9.1: `TestSampling_Determinism_PerBackend`:
- Ollama: two identical warm calls produce byte-identical UIAST ≥0.95 of the time.
- Claude: two identical warm calls produce byte-identical UIAST ≥0.80 of the time (softer target).

AC-9.2: every sampling parameter is configurable through `Config` so the eval harness can vary.

## BDD Test Scenarios

```gherkin
Feature: Deterministic sampling

  Scenario: AC-9.1 — Ollama byte-identical rate
    Given a warmed Ollama backend with seed=42, temperature=0
    When 100 pairs of identical Translate calls run
    Then ≥95 pairs return byte-identical UIAST

  Scenario: AC-9.1b — Claude softer rate
    Given a warmed Claude API backend with temperature=0
    When 100 pairs of identical Translate calls run
    Then ≥80 pairs return byte-identical UIAST

  Scenario: AC-9.2 — Eval-harness configurability
    Given cfg.Temperature=0.2 and cfg.Seed=7
    When the Ollama client sends a request
    Then the wire payload has `options.temperature=0.2` and `options.seed=7`
```

## Tasks / Subtasks

- [ ] Task 1 — Add Temperature/Seed to Config (maps to AC-9.2)
- [ ] Task 2 — Wire Ollama client (maps to AC-9.1, AC-9.2)
- [ ] Task 3 — Wire Claude API client (maps to AC-9.1, AC-9.2)
- [ ] Task 4 — Tests (maps to AC-9.1, AC-9.2)

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
