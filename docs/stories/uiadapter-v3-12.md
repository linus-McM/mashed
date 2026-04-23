# uiadapter-v3-12: Model allowlist

**Status:** ready
**Domain:** backend
**Size:** S
**Depends On:** B, C
**Priority:** P2-medium

## Story

As a Mashed operator, I want the adapter to warn when Model/ClaudeModelPrimary is set to an un-vetted model, so that configuration foot-guns (e.g. typo `"gemma4:latest"`, renamed `"claude-foo-9"`) get flagged at boot without breaking launch.

## Description

Plan §3 Phase 3 "Story 12 — Model allowlist" (lines 514–524). Supersedes stale story `uiadapter-04` (which only covered Ollama).

- **Ollama allowlist:** `gemma3:4b`, `gemma3:4b-it-qat`, `gemma3:12b`, `gemma3:1b`.
- **Claude allowlist:** `claude-haiku-4-5`, `claude-sonnet-4-6`, `claude-opus-4-6`.
- **Override:** `Config.AllowUnvettedModels=true`.
- **Behavior:** on construction, if the model isn't on the list and the override flag is false, log a Warn (don't fail — Go idiom prefers warn for config-layer).

## Developer Notes

- **Files (edited):** `internal/uiadapter/config.go`, `internal/uiadapter/backend/ollama/client.go`, `internal/uiadapter/backend/claudeapi/backend.go`, `internal/uiadapter/backend/claudecli/backend.go`.
- **Types / API surface:** package-level `OllamaAllowlist = []string{...}`, `ClaudeAllowlist = []string{...}`. `func IsAllowed(model string, allowlist []string) bool`.
- **Risks:** none material — allowlist is purely a diagnostic warn.
- **Dependencies:** none new.
- **Use-repo-code directive:** check `use-repo-code` for existing Ollama model allowlist from stale story `uiadapter-04` — do not import; this story supersedes.

## Acceptance Criteria

AC-12.1: `TestAllowlist_WarnsOnUnvetted` — `Model="gemma4:latest"` or `ClaudeModelPrimary="claude-foo-9"` with `AllowUnvettedModels=false` produces a Warn log.

AC-12.2: no Warn for allowlisted models.

AC-12.3: no Warn when `AllowUnvettedModels=true`.

## BDD Test Scenarios

```gherkin
Feature: Model allowlist

  Scenario: AC-12.1 — Warn on unvetted
    Given Model="gemma4:latest" and AllowUnvettedModels=false
    When the Ollama backend is constructed
    Then a Warn slog line includes the model name
    And construction succeeds

  Scenario: AC-12.1b — Warn on unvetted Claude
    Given ClaudeModelPrimary="claude-foo-9" and AllowUnvettedModels=false
    When the Claude API backend is constructed
    Then a Warn slog line includes the model name

  Scenario: AC-12.2 — No warn on allowlisted
    Given Model="gemma3:4b"
    When construction runs
    Then no Warn is emitted about the model

  Scenario: AC-12.3 — Override silences warn
    Given Model="gemma4:latest" and AllowUnvettedModels=true
    When construction runs
    Then no Warn is emitted about the model
```

## Tasks / Subtasks

- [ ] Task 1 — Define allowlists (maps to AC-12.1, AC-12.2)
- [ ] Task 2 — Add `IsAllowed` + call at constructor (maps to AC-12.1, AC-12.2, AC-12.3)
- [ ] Task 3 — Tests `TestAllowlist_WarnsOnUnvetted` + negatives

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
