# uiadapter-v3-03: ContextGuard

**Status:** done
**Landed:** 2026-04-23
**Domain:** backend
**Size:** M
**Depends On:** B, C
**Priority:** P0-critical

## Story

As an operator, I want per-backend context management (explicit `num_ctx` on Ollama with middle-elision truncation, honest refusal on Claude), so that long captures don't silently overflow the model's window and fall back deterministically.

## Description

Plan §3 Phase 1 "Story 3 — ContextGuard" (lines 316–329). §0.3 establishes the root cause: `gemma3:4b` has a 128K window but Ollama's default `num_ctx` is 2K–4K and silently truncates. Most "context" problems are misconfiguration. This story (a) sends `options.num_ctx: 8192` explicitly on every Ollama request, (b) truncates raw on Ollama path with middle-elision (`[... N lines elided ...]`), (c) refuses with `ContextGuard: refused` on the Claude path (Anthropic returns 400 on overflow; conversion to an explicit refusal lets the fallback chain handle it). Documents required env flags (`OLLAMA_FLASH_ATTENTION=1`, `OLLAMA_KV_CACHE_TYPE=q8_0`, `OLLAMA_NUM_PARALLEL=1`).

## Developer Notes

- **Files (new):** `internal/uiadapter/contextguard.go`, `internal/uiadapter/contextguard_test.go`.
- **Files (edited):** `internal/uiadapter/config.go` (`NumCtx` from Story B), `internal/uiadapter/backend/ollama/client.go` (send `options.num_ctx` on every chat request).
- **Types / API surface:** `func Guard(raw string, caps Capabilities, cfg Config) (guarded string, action string, err error)` where `action ∈ {"pass","truncate","refuse"}`.
- **Ollama branch:** approximate tokens as `len(raw)/4`; if estimate exceeds `num_ctx - 1024` reserve, keep first N lines and last M lines, replace middle with `[... N lines elided ...]`. Always preserve tail.
- **Claude branch:** no silent truncation. Return `ErrContextOverflow` from this module when estimate exceeds backend `MaxContextTokens - max_tokens`. Caller maps to refuse + warn log.
- **Risks:** mis-estimating tokens: ±20% sufficient (§3 Story 3). Guard is conservative (1024-token reserve).
- **Dependencies:** none new.
- **Use-repo-code directive:** use `use-repo-code` to read the current Ollama client (`internal/uiadapter/client.go` or equivalent) for the `chatRequest.Options` shape.

## Acceptance Criteria

AC-3.1: `TestContextGuard_TruncatesLongCapture_Ollama` — 20K-char input truncated to ≤8K tokens on Ollama path with sentinel in place.

AC-3.2: `num_ctx` present in wire payload on every Ollama call.

AC-3.3: `TestContextGuard_RefusesLongCapture_Claude` — 200K-token raw on the Claude path returns a `ContextGuard: refused` error, which the fallback chain (Story 11) handles.

AC-3.4: truncation events log at Warn, not Error.

## BDD Test Scenarios

```gherkin
Feature: ContextGuard

  Scenario: AC-3.1 — Ollama middle-elision truncation
    Given raw of ~20K chars (>= 5K approx tokens) and NumCtx=8192
    When Guard runs with Ollama capabilities
    Then action is "truncate"
    And the guarded string contains "[... N lines elided ...]"
    And the tail is preserved byte-for-byte

  Scenario: AC-3.2 — num_ctx wire payload
    Given an Ollama chat request
    When the payload is serialized
    Then the `options.num_ctx` field equals cfg.NumCtx

  Scenario: AC-3.3 — Claude refuses overflow
    Given raw of ~200K tokens and Claude capabilities
    When Guard runs
    Then err is ErrContextOverflow
    And the router fallback chain handles the refusal

  Scenario: AC-3.4 — Warn-level logging
    Given a truncation event
    When slog emits the log line
    Then the level is Warn, not Error
```

## Tasks / Subtasks

- [ ] Task 1 — Implement `Guard` (maps to AC-3.1, AC-3.3)
  - [ ] Subtask 1a — Token estimator `approxTokens(raw) = len(raw)/4`.
  - [ ] Subtask 1b — Ollama branch: middle-elision + sentinel.
  - [ ] Subtask 1c — Claude branch: return `ErrContextOverflow`.
- [ ] Task 2 — Ollama wire patch (maps to AC-3.2)
  - [ ] Subtask 2a — Set `chatRequest.Options.NumCtx = cfg.NumCtx` on every request.
- [ ] Task 3 — Tests `TestContextGuard_TruncatesLongCapture_Ollama`, `TestContextGuard_RefusesLongCapture_Claude` (maps to AC-3.1, AC-3.3)
- [ ] Task 4 — Warn-level slog on truncation (maps to AC-3.4)
- [ ] Task 5 — README / operator docs for env flags (supports AC-3.1, AC-3.2)

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
