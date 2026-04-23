# uiadapter-v3-07: Prompt prefix caching

**Status:** done
**Landed:** 2026-04-23
**Domain:** backend
**Size:** M
**Depends On:** B, C, v3-05
**Priority:** P0-critical

## Story

As a uiadapter author, I want prompt prefixes cached per-backend — byte-stable prefix on Ollama for KV reuse, explicit `cache_control` markers on Claude API, `--resume <session>` on Claude CLI — so that warm p50 Stage-2 latency drops under 600ms on Ollama, cache-read tokens show up on the second Claude call, and cold-start subprocess cost on CLI is paid once per session.

## Description

Plan §3 Phase 2 "Story 7 — Prompt prefix caching" (lines 410–422). Ollama: llama.cpp auto-reuses KV cache when byte-exact prefix matches across calls; `options.keep_alive = "30m"` (or `OLLAMA_KEEP_ALIVE=-1` in dev); fire-and-forget warm-up at `NewDefault` with 10s timeout. Claude API: explicit `cache_control: {type:"ephemeral"}` markers on static system blocks; TTL via `Config.PromptCacheTTL` (`5m`/`1h`/`off`); `usage.cache_creation_input_tokens` and `usage.cache_read_input_tokens` logged. Claude CLI: extract session ID from first call's `init` event and pass `--resume <id>` on subsequent calls; one session per backend instance; rotate on error. §8 risk "Story 7 (static prefix) is order-sensitive" — any non-byte-identical variation destroys cache.

Supersedes stale story `uiadapter-06` (partial — timeout/keep_alive/warmup).

## Developer Notes

- **Files (edited):**
  - `internal/uiadapter/backend/ollama/client.go` — `keep_alive`, warm-up goroutine.
  - `internal/uiadapter/backend/claudeapi/request.go` — `cache_control` markers on system blocks.
  - `internal/uiadapter/backend/claudecli/session.go` — `--resume <session_id>` tracking.
- **Types / API surface:** `Backend.WarmUp(ctx) error` implemented per backend; `Adapter.NewDefault` spawns WarmUp goroutines non-blocking.
- **Byte-stable static prefix:** share with Story 5 — every call assembled as `[static_prefix][delimiter][dynamic_suffix]`. `TestPrompt_StaticPrefix_ByteStable` hashes prefix across 1000 synthetic calls.
- **Risks:** §8 order-sensitive prefix; session rotation needed on CLI error.
- **Dependencies:** Claude API docs pinned at `anthropic-version: 2023-06-01`.
- **Use-repo-code directive:** inspect current Ollama client's `Chat` signature and executor's adapter construction timing via `use-repo-code`.

## Acceptance Criteria

AC-7.1: (Ollama) p50 warm Stage-2 latency <600ms.

AC-7.2: (Ollama) cold-start p95 <4s after fresh `ollama serve`.

AC-7.3: `TestAdapter_WarmupDoesNotBlockConstruction` — `NewDefault` returns <50ms even if no backends are reachable.

AC-7.4: `TestPrompt_StaticPrefix_ByteStable` — sha256 of prefix is identical across 1000 synthetic Translate calls. Runs against all three backends.

AC-7.5: (Claude API) cached-call latency p50 <400ms for Haiku and <800ms for Sonnet after warm-up (second call onward).

AC-7.6: (Claude API) cache hit rate on the static prefix ≥0.95 after the first call; `usage.cache_read_input_tokens` > 0 on the second call to an identical prefix.

## BDD Test Scenarios

```gherkin
Feature: Prompt prefix caching

  Scenario: AC-7.1 — Ollama warm p50
    Given a warmed-up Ollama backend
    When 100 Stage-2 calls run back to back
    Then p50 is under 600ms

  Scenario: AC-7.2 — Cold-start p95
    Given a freshly-started `ollama serve`
    When the first 20 Stage-2 calls run
    Then p95 is under 4s

  Scenario: AC-7.3 — Non-blocking construction
    Given no backend is reachable
    When NewDefault(cfg) runs
    Then it returns in under 50ms
    And WarmUp continues in a goroutine

  Scenario: AC-7.4 — Byte-stable prefix hash
    Given 1000 synthetic Translate calls per backend
    When the static prefix is sha256-hashed each call
    Then every hash is identical

  Scenario: AC-7.5 — Claude API cached p50
    Given a warmed Haiku/Sonnet backend (second call onward)
    When 50 cached Stage-2 calls run
    Then Haiku p50 < 400ms and Sonnet p50 < 800ms

  Scenario: AC-7.6 — cache_read_input_tokens on second call
    Given an identical static prefix across two Claude API calls
    When the second call completes
    Then usage.cache_read_input_tokens > 0
    And cache hit rate over 100 calls ≥ 0.95
```

## Tasks / Subtasks

- [ ] Task 1 — Ollama keep_alive + warm-up (maps to AC-7.1, AC-7.2, AC-7.3)
  - [ ] Subtask 1a — Send `options.keep_alive = cfg.KeepAlive`.
  - [ ] Subtask 1b — Goroutine warm-up in NewDefault with 10s timeout.
- [ ] Task 2 — Claude API cache_control (maps to AC-7.5, AC-7.6)
  - [ ] Subtask 2a — Mark static system blocks with `cache_control: {type:"ephemeral"}`.
  - [ ] Subtask 2b — Log `usage.cache_*_input_tokens`.
- [ ] Task 3 — Claude CLI session reuse (maps to AC-7.5)
  - [ ] Subtask 3a — Parse `init` event for session_id.
  - [ ] Subtask 3b — Append `--resume <session_id>` on subsequent calls.
  - [ ] Subtask 3c — Rotate on error.
- [ ] Task 4 — Static prefix byte-stable enforcement (maps to AC-7.4)
- [ ] Task 5 — Tests — see §9 Deliverables

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
