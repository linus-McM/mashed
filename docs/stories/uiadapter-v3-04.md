# uiadapter-v3-04: ResponseCache + Singleflight

**Status:** ready
**Domain:** backend
**Size:** S
**Depends On:** B, C
**Priority:** P0-critical

## Story

As a Mashed user, I want identical Claude-Code turns cached and concurrent identical translates coalesced into one downstream call, so that repeated turns cost zero LLM calls and concurrent bursts don't fan out against a rate-limited provider.

## Description

Plan §3 Phase 1 "Story 4 — ResponseCache + Singleflight" (lines 331–345). LRU keyed on `sha256(backend_name + model + cacheVersion + sanitized_raw)` — backend in the key guarantees Ollama/Claude results never collide and switching backends invalidates stale entries. `cacheVersion` is a package-level constant bumped manually when prompts or schemas change (simplest correct invalidation). Singleflight via `golang.org/x/sync/singleflight` collapses concurrent identical calls into one. Optional semantic cache (nomic-embed-text via Ollama, cosine ≥0.97) is flag-gated, off by default, never used in eval mode (§8 risk "Semantic cache false positives").

## Developer Notes

- **Files (new):** `internal/uiadapter/cache.go`, `internal/uiadapter/cache_test.go`.
- **Types / API surface:**
  ```go
  type Cache interface {
      Get(key string) (*UIAST, bool)
      Put(key string, ast *UIAST)
  }
  func NewLRUCache(capacity int) Cache
  func CacheKey(backendName, model string, cacheVersion int, sanitizedRaw string) string
  const cacheVersion = 1
  ```
  Singleflight group lives on the `Adapter`.
- **Dependencies:** `github.com/hashicorp/golang-lru/v2`, `golang.org/x/sync/singleflight`.
- **Capacity:** `Config.CacheCapacity` (default 1024; 0 disables).
- **Semantic cache (optional):** `EnableSemanticCache=true` uses `nomic-embed-text`; cosine ≥0.97 returns stored AST. AC-4.4 guards parse-rate on the eval corpus.
- **Risks:** §8 "Semantic cache false positives" — off by default, conservative threshold.
- **Use-repo-code directive:** use `use-repo-code` to locate the current Translate entrypoint for cache wiring.

## Acceptance Criteria

AC-4.1: `TestCache_HashHit` — two identical Translate calls produce exactly one downstream call on the second path.

AC-4.2: `TestCache_SingleflightCoalesces` — 100 concurrent identical calls produce exactly one downstream call.

AC-4.3: `cache_hit_rate` and `singleflight_shared` attributes on every log line.

AC-4.4: semantic cache, when enabled, does not reduce parse-rate on the eval corpus.

AC-4.5: switching `Config.Backend` invalidates cache entries from the prior backend.

## BDD Test Scenarios

```gherkin
Feature: ResponseCache + Singleflight

  Scenario: AC-4.1 — Cache hit suppresses second downstream call
    Given a populated cache with one entry for raw=X
    When Translate(X) is called a second time
    Then the downstream backend is not invoked
    And the returned UIAST equals the cached one byte-for-byte

  Scenario: AC-4.2 — Singleflight coalesces concurrent identical calls
    Given 100 goroutines calling Translate(X) simultaneously with an empty cache
    When they all complete
    Then the downstream backend recorded exactly one call
    And all 100 goroutines received the same UIAST pointer (or deep-equal value)

  Scenario: AC-4.3 — Telemetry attributes
    Given any Translate call
    When slog emits the structured log line
    Then `cache_hit_rate` and `singleflight_shared` attributes are present

  Scenario: AC-4.4 — Semantic cache preserves parse-rate
    Given EnableSemanticCache=true
    When the eval corpus runs
    Then parse-rate does not drop versus EnableSemanticCache=false

  Scenario: AC-4.5 — Backend switch invalidates prior keys
    Given cached entries from Backend=ollama
    When Config.Backend flips to claude-api
    Then subsequent Translate(X) misses the cache and hits the new backend
```

## Tasks / Subtasks

- [ ] Task 1 — Implement LRU cache + key function (maps to AC-4.1, AC-4.5)
  - [ ] Subtask 1a — `NewLRUCache(capacity)` with `hashicorp/golang-lru/v2`.
  - [ ] Subtask 1b — `CacheKey` sha256 of `backend_name + model + cacheVersion + sanitizedRaw`.
- [ ] Task 2 — Wire singleflight (maps to AC-4.2)
  - [ ] Subtask 2a — `singleflight.Group` on Adapter.
  - [ ] Subtask 2b — DoChan keyed on the same hash.
- [ ] Task 3 — Tests (maps to AC-4.1, AC-4.2, AC-4.5)
- [ ] Task 4 — Slog attributes (maps to AC-4.3)
- [ ] Task 5 — Optional semantic cache (maps to AC-4.4)
  - [ ] Subtask 5a — Embed via Ollama `nomic-embed-text`.
  - [ ] Subtask 5b — Cosine ≥0.97 gate.
  - [ ] Subtask 5c — Disabled in eval mode.

## Definition of Done

- [ ] All ACs verified with PASS evidence
- [ ] 80%+ line coverage on modified Go files
- [ ] `/simplify` skill run
- [ ] `go build ./... && go vet ./...` clean
- [ ] `go test ./... -race -count=1 -short` clean
- [ ] Story file `Status: ready` → `done`; all `- [ ]` → `- [x]`

## Design Brief

_(backend story — no design brief)_
