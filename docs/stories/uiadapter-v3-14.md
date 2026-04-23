# uiadapter-v3-14: ClaudeAPIBackend

**Status:** ready
**Domain:** backend
**Size:** L
**Depends On:** B, C, v3-06, v3-07
**Priority:** P0-critical

## Story

As a Mashed user with an Anthropic API key, I want a direct `ClaudeAPIBackend` that uses `tool_use` for native structured output, `cache_control` for prompt caching, and the soft/hard rate-limit accountant, so that Translate reliability jumps to ≥0.995 parse rate at sub-400ms p50 latency on Haiku.

## Description

Plan §3 Phase 4 "Story 14 — `ClaudeAPIBackend`" (lines 561–599). Direct Anthropic API client via `net/http`, no third-party SDK dep. Endpoint `POST /v1/messages`. Headers pinned: `anthropic-version: 2023-06-01`. Structured output via `tools: [{name: "emit_uiast_<kind>", input_schema: ...}], tool_choice: {type: "tool", name: ...}`. Response has `stop_reason: "tool_use"` and a `tool_use` content block whose `input` is already a JSON object matching the schema — no intermediate parsing. Streaming is optional (`Config.Streaming=true`, default false, disabled in v3.0). Errors: 429 with `Retry-After` → wait-retry once; 529 → trip breaker; 400/500 → trip breaker. Secrets in sealed `type apiKey string` — redacted in every log.

## Developer Notes

- **Files (new):**
  - `internal/uiadapter/backend/claudeapi/backend.go` (LLMBackend impl + registry init)
  - `internal/uiadapter/backend/claudeapi/request.go` (request builder + cache_control)
  - `internal/uiadapter/backend/claudeapi/stream.go` (SSE parser, default-off)
  - `internal/uiadapter/backend/claudeapi/accountant.go` (shares Story 11b)
- **Types / API surface:**
  ```go
  type backend struct { http *http.Client; key apiKey; model string; caps Capabilities; accountant *Accountant }
  func New(cfg Config) (*backend, error)
  func init() { backend.Register("claude-api", New) }
  type apiKey string // no Stringer, redacted in logs
  ```
- **Request shape:** system block has `cache_control: {type:"ephemeral"}` (Story 7).
- **Risks:** §8 "Anthropic API contract changes" — version-pinned header; `httptest.Server` integration tests; monthly live-API smoke (manual).
- **Dependencies:** `net/http` stdlib only.
- **Use-repo-code directive:** use `use-repo-code` to confirm no existing Anthropic client already in-tree that this should extend vs. replace.

## Acceptance Criteria

AC-14.1: `TestClaudeAPI_ToolUseRoundtrip` — Translate through a stubbed HTTP server returns a golden UIAST matching `testdata/eval/*/expected.json`.

AC-14.2: `TestClaudeAPI_RetryAfter429` — 429 with `Retry-After: 2` is respected; second-attempt success returns a UIAST; second-attempt failure returns `ErrRateLimited`.

AC-14.3: `tool_use` block's `input` is used directly; no intermediate JSON parsing needed.

AC-14.4: `TestClaudeAPI_PromptCacheMarkers` — wire payload has `cache_control` markers; `usage.cache_read_input_tokens > 0` on the second call to an identical static prefix.

AC-14.5: `TestClaudeAPI_KeyNotLogged` — grep the captured slog output; no substring of the key.

## BDD Test Scenarios

```gherkin
Feature: ClaudeAPIBackend

  Scenario: AC-14.1 — tool_use roundtrip
    Given a stubbed Anthropic server returning a tool_use content block
    When Translate runs against a corpus entry
    Then the decoded UIAST matches the expected JSON byte-for-byte

  Scenario: AC-14.2 — 429 Retry-After honored
    Given the stub returns 429 with `Retry-After: 2` on the first call
    When Translate runs
    Then it waits ≥ 2s
    And the second call's success returns a UIAST
    And a second 429 returns ErrRateLimited

  Scenario: AC-14.3 — No intermediate parse
    Given a successful tool_use response
    When the backend decodes
    Then the `input` object is marshaled directly into UIAST
    And no `json.Unmarshal(string, ...)` bridging step is present

  Scenario: AC-14.4 — Prompt cache markers
    Given two back-to-back Translate calls with identical static prefix
    When the second call's response arrives
    Then the wire payload had `cache_control.type == "ephemeral"`
    And usage.cache_read_input_tokens > 0 on the second response

  Scenario: AC-14.5 — Key not logged
    Given ANTHROPIC_API_KEY=sk-secret-123
    When 100 Translate calls run and slog is captured
    Then the string "sk-secret-123" does not appear in any log line
```

## Tasks / Subtasks

- [ ] Task 1 — Implement backend struct + registry init (maps to AC-14.1)
- [ ] Task 2 — Request builder + `tool_use` encoding (maps to AC-14.1, AC-14.3)
- [ ] Task 3 — Cache-control markers (maps to AC-14.4)
- [ ] Task 4 — 429 Retry-After + 529 breaker handling (maps to AC-14.2)
- [ ] Task 5 — Key redaction (`type apiKey string`, no Stringer, scrub in logger) (maps to AC-14.5)
- [ ] Task 6 — Tests (maps to AC-14.1–14.5)

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
