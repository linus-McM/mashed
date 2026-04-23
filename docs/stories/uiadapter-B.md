# uiadapter-B: Unified Config surface

**Status:** done
**Domain:** backend
**Size:** S
**Depends On:** —
**Priority:** P0-critical
**Landed:** 2026-04-23

## Story

As a uiadapter author, I want every knob introduced by the v3 plan to live in a single `Config` struct with a working `DefaultConfig()` factory, so that all downstream stories read from one dependency-injected surface instead of package-level constants.

## Description

Plan §3 Phase 0 "Story B — Unified `Config` surface" (lines 188–252). The struct consolidates backend selection, Ollama tuning, Claude API settings, Claude CLI settings, cost/rate accountant fields, sampling, timeouts, runtime behavior (cache capacity, fast-path enable, shadow-sample rate, breaker thresholds, privacy regex patterns, streaming). §6.5 codifies the "Dependency injection via `Config`. No package-level globals." rule — this story is where it becomes law.

The existing `internal/uiadapter/` package already has a thin `Config` from `ui-ast-U2-adapter-package` (via `use-repo-code`). This story supersedes it with the full v3 shape. Zero-valued input must flow through `DefaultConfig()` to a bootable adapter with no panics — AC-B.1.

## Developer Notes

- **Files (edited):** `internal/uiadapter/config.go`, `internal/uiadapter/adapter.go` (constructor wiring).
- **Types / API surface:** `Config` struct with all fields from plan §3 Story B (Backend, RouterPolicy, FallbackOrder, OllamaEndpoint, Model, AllowUnvettedModels, NumCtx, KeepAlive, LooseFormat, AnthropicAPIKeyEnv, ClaudeModelPrimary, ClaudeModelHard, ClaudeMaxTokens, AnthropicVersion, PromptCacheTTL, ClaudeCLIBinary, ClaudeCLIExtraFlags, UsdBudgetPerSession, RPMSoftLimit, TPMSoftLimit, Temperature, Seed, TimeoutMs, WarmUpTimeoutMs, CacheCapacity, EnableSemanticCache, EnableFastPath, EnableSpotlighting, RepairMaxRetries, BreakerFailThreshold, BreakerResetMs, ShadowSampleRate, PrivacyPatterns, DisableHealthTicker, Streaming). `DefaultConfig() Config` returns the documented defaults.
- **Risks:** zero-value field additions in later stories may silently regress boot — the `TestDefaultConfig_Bootable_AllBackends` test is the guard (§9 Deliverables).
- **Dependencies:** `regexp` (for `PrivacyPatterns`). No new third-party deps.
- **Use-repo-code directive:** Inspect the existing `internal/uiadapter/config.go`, `adapter.go`, `NewDefault` signature via `use-repo-code` (`.claude/skills/use-repo-code/references/summary.md` → grep `files.md` for those paths).

## Acceptance Criteria

AC-B.1: zero-valued `Config` passed through `DefaultConfig` boots a working adapter — no panics, no missing fields.

AC-B.2: every subsequent story reads its knobs from `Config`, not from package constants.

## BDD Test Scenarios

```gherkin
Feature: Unified Config surface

  Scenario: AC-B.1 — Default config boots an adapter
    Given the caller passes Config{} to DefaultConfig()
    And feeds the resulting struct to NewDefault(cfg)
    When the adapter is constructed
    Then no panic fires
    And every backend-relevant field has a sane zero value
    And `Translate` returns a valid UIAST for a trivial raw capture

  Scenario: AC-B.2 — Downstream stories consume Config (contract test)
    Given the adapter package compiles
    When `go vet ./internal/uiadapter/...` inspects the tree
    Then grep reveals zero references to unexported package-level knobs that bypass Config
    And all timeouts, endpoints, and feature flags are sourced from the Config passed to the constructor
```

## Tasks / Subtasks

- [ ] Task 1 — Define Config struct and DefaultConfig (maps to AC-B.1)
  - [ ] Subtask 1a — Port the full field list from plan §3 Story B into `config.go`.
  - [ ] Subtask 1b — Implement `DefaultConfig()` returning the documented defaults.
  - [ ] Subtask 1c — Sanitize zero-valued fields inside the constructor (e.g. empty Model → "gemma3:4b").
- [ ] Task 2 — Wire constructor (maps to AC-B.1)
  - [ ] Subtask 2a — Update `NewDefault` to accept `Config` and apply defaults via `DefaultConfig`.
  - [ ] Subtask 2b — Ensure existing callers in `executor.go` continue to compile (keep a no-arg overload or provide migration stub).
- [ ] Task 3 — Write `TestDefaultConfig_Bootable_AllBackends` (maps to AC-B.1)
- [ ] Task 4 — Enforce no-global-constants policy (maps to AC-B.2)
  - [ ] Subtask 4a — Grep-based lint (or staticcheck rule) rejects `const` backend URLs, timeouts, model names in package scope.

## Definition of Done

- [ ] All ACs verified with PASS evidence (test name or CI check)
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run — no redundancy
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
