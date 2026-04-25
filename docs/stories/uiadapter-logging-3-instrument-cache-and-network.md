# Story 3: Instrument cache and network layer (client, cache, prefix_cache, breaker, semaphore)

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** uiadapter-logging-2-plumb-subcomponents
**Status:** ready

## Description

Add structured `slog` debug-level instrumentation to the cache and network surface of the UI adapter: `client.go`, `cache.go`, `prefix_cache.go`, `breaker.go`, `semaphore.go`. Every Debug call respects the §14 sanitize discipline (lengths, hashes, durations, IDs only — never raw payload bytes). Hot-path call sites use the `Enabled` guard pattern to keep zero-allocation behaviour when debug is off. After this story, a `UIADAPTER_LOG_LEVEL=debug` run shows every HTTP request, every cache hit/miss/eviction, every breaker state change, and every semaphore acquire wait.

## Developer Notes

### Architecture

Instrument these five files with `logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)` calls. Every site must follow the **hot-path guard**:

```go
if c.logger.Enabled(ctx, slog.LevelDebug) {
    c.logger.LogAttrs(ctx, slog.LevelDebug, "client.chat.start",
        slog.String("op", "client.chat"),
        slog.String("model", model),
        slog.Int("bytes_in", len(raw)),
    )
}
```

#### `internal/uiadapter/client.go`

| Site | Message | Required attrs |
|---|---|---|
| Entry of `Chat` / `ChatDeterministic` | `client.chat.start` | `op="client.chat"`, `model`, `bytes_in=len(raw)` |
| Successful response | `client.chat.response` | `op`, `model`, `latency_ms`, `bytes_out=len(body)` |
| Transport error | `client.chat.transport_error` | `op`, `model`, `latency_ms`, `reason=classifyChatErr(err)` (do NOT log `err.Error()` — only the classifier enum) |
| Non-2xx HTTP status | `client.chat.http_error` | `op`, `model`, `status_code`, `latency_ms` |

`Client.logger` was added in Story 2; if not present, the constructor must be re-checked.

#### `internal/uiadapter/cache.go`

| Site | Message | Required attrs |
|---|---|---|
| Get hit | `cache.hit` | `op="cache.get"`, `cache_key_hash` (8-char SHA-256 prefix) |
| Get miss | `cache.miss` | `op="cache.get"`, `cache_key_hash` |
| Put new entry | `cache.put` | `op="cache.put"`, `cache_key_hash`, `size_bytes=len(value)` |
| Eviction (LRU) | `cache.evict` | `op="cache.evict"`, `cache_key_hash`, `reason="lru"` |

`cache_key_hash` is computed as `hex.EncodeToString(sha256.Sum256([]byte(key))[:4])` (8 hex chars). Add a small private helper `hashKey(key string) string` next to the cache type.

#### `internal/uiadapter/prefix_cache.go`

| Site | Message | Required attrs |
|---|---|---|
| `ClaudeSystemBlock` entry | `prefix_cache.build` | `op="prefix_cache.claude"`, `prefix_len=len(staticPrefix)`, `ttl=cacheControlType(ttl)` |
| `ClaudeSystemBlockJSON` marshal success | `prefix_cache.build.success` | `op`, `bytes_out=len(out)` |
| `OllamaKeepAliveEncoded` | `prefix_cache.ollama_keep_alive` | `op="prefix_cache.ollama"`, `keep_alive=cfg.KeepAlive` |

These are free functions — they receive `logger *slog.Logger` from Story 2's plumbing.

#### `internal/uiadapter/breaker.go`

| Site | Message | Required attrs |
|---|---|---|
| State transition (closed -> open) | `breaker.transition` | `op="breaker.transition"`, `from="closed"`, `to="open"`, `backend=<key>` |
| State transition (open -> half_open) | `breaker.transition` | `from="open"`, `to="half_open"`, `backend` |
| State transition (half_open -> closed) | `breaker.transition` | `from="half_open"`, `to="closed"`, `backend` |
| Reject (breaker open at request time) | `breaker.reject` | `op="breaker.reject"`, `backend`, `state="open"` |

`BreakerSet.logger` was added in Story 2.

#### `internal/uiadapter/semaphore.go`

| Site | Message | Required attrs |
|---|---|---|
| Acquire wait > 0 | `semaphore.wait` | `op="semaphore.acquire"`, `wait_ms`, `in_flight` |
| Acquire success | `semaphore.acquired` | `op`, `in_flight` |
| Acquire ctx-cancelled | `semaphore.cancelled` | `op`, `wait_ms` |

`semaphore.in_flight` should be derivable from the underlying channel: `cap(s.ch) - len(s.ch)`. Compute lazily inside the `Enabled` guard.

### Technical Considerations

- **Sanitize discipline (§14)** — never log: `raw`, `body`, `prompt`, `response.text`, model output, error.Error() text, cache values. Only: lengths, counts, durations, IDs, enums, hash prefixes, booleans.
- **Hash helper:** add `hashKey(s string) string` in `cache.go` — 8-char SHA-256 hex prefix. Reuse the same helper in `prefix_cache.go` if it ever needs to log a key.
- **Latency measurement:** use `time.Since(start).Milliseconds()` and emit as `slog.Int64("latency_ms", ms)`. Capture `start` immediately before the operation begins.
- **Error classification:** `client.chat.transport_error` uses `classifyChatErr(err)` (already exists in `adapter.go`). It returns enums like `"timeout"`, `"saturated"`, `"transport"`. Never the raw `err.Error()`.
- **Context plumbing:** every Debug call site must pass `ctx` so child handlers can attach trace IDs in the future. If a function doesn't have a `ctx` (rare), pass `context.Background()`.
- **Breaker concurrency:** state transitions happen under the breaker's existing mutex. Emit the log INSIDE the lock so the `from`/`to` pair is consistent (the cost is minimal — debug is filtered when off).
- **Semaphore allocation:** `cap(ch)-len(ch)` is racy without locking but only used inside `Enabled`-guarded debug emission; an off-by-one is acceptable observability noise.

### Risks & Edge Cases

- **Hot-path overhead:** Translate is called per-Claude-turn. The `Enabled` guard is mandatory on every site — without it, attribute-construction allocations occur even when the level is filtered out.
- **`classifyChatErr` location:** lives in `adapter.go`. If `client.go` cannot see it (package internal — it can), expose a small package-private helper or move/duplicate the function. Prefer reuse to avoid drift.
- **Breaker state names:** stick to the literal strings `"closed" | "open" | "half_open"` in `from`/`to` attrs — these become enum strings in BI dashboards and must match the existing breaker telemetry from `adapter.go:logTelemetry`.
- **No log spam under load:** every site must be Debug. Anything users care about under normal operation already lives in `adapter.go:logTelemetry`. Keep this story strictly Debug.
- **Tests at info level:** existing tests do not assert on log content. Story 6 adds the asserting tests. Each site here must remain testable — keep the function bodies short and instrument only at well-defined entry/exit/branch points.
- **No string interpolation in messages:** the `msg` string is a stable identifier (e.g. `"client.chat.start"`). Variables go in attrs. Greppability matters.

### Reference Files

- `/Users/linus/Development/mashed/internal/uiadapter/adapter.go:200-260` — existing `logTelemetry` / `logTelemetryWithSanitize` shows the canonical sanitize discipline. Mirror its attribute style.
- `/Users/linus/Development/mashed/internal/uiadapter/allowlist.go:31` — existing `slog.Default()` usage gives the template for warn-level emission.
- `/Users/linus/Development/mashed/internal/uiadapter/logging.go` (created in Story 1) — contains `nilSafeLogger`; standard attribute names in the doc comment.
- `/Users/linus/Development/mashed/internal/uiadapter/client.go` — primary file; ~80 LOC.
- `/Users/linus/Development/mashed/internal/uiadapter/cache.go` — second largest target; LRU eviction is the eviction site.
- `/Users/linus/Development/mashed/internal/uiadapter/breaker.go` — state machine; transition is the only branch worth instrumenting.
- `/Users/linus/Development/mashed/internal/uiadapter/semaphore.go` — small file but every Translate goes through it.

### Skills

- Invoke `/simplify` after each file's instrumentation.
- Invoke `/golang-testing` to write the table-driven log-assertion helper used in this story's tests (a `bytes.Buffer` + `slog.NewJSONHandler` + `json.Decoder` loop).

## Acceptance Criteria

AC-3.1: `client.go` emits debug logs at request boundaries
- Given a `Client` constructed with a captured-buffer logger at level Debug
- When `client.Chat(ctx, model, system, raw)` runs to a successful response
- Then the buffer contains a `client.chat.start` record with `op`, `model`, `bytes_in` attrs
- And the buffer contains a `client.chat.response` record with `op`, `model`, `latency_ms`, `bytes_out` attrs
- And neither record contains the strings `raw`, `system`, or any substring of the body

AC-3.2: `client.go` emits debug log on transport error and HTTP error
- Given a `Client` pointed at a closed listener
- When `Chat` is called
- Then the buffer contains a `client.chat.transport_error` record with attribute `reason="transport"` (or the appropriate `classifyChatErr` enum)
- And no attribute named `error` containing the raw `err.Error()` is present
- Given a `Client` pointed at a server returning HTTP 500
- When `Chat` is called
- Then the buffer contains a `client.chat.http_error` record with `status_code=500`

AC-3.3: `cache.go` emits hit, miss, put, and evict events
- Given a `ResponseCache` with capacity 2 and Debug logger
- When the cache is exercised: put A, put B, get A (hit), put C (evicts B), get B (miss)
- Then the buffer contains records: `cache.put` x3, `cache.hit` x1, `cache.miss` x1, `cache.evict` x1 with `reason="lru"`
- And every record carries an 8-character `cache_key_hash` attribute

AC-3.4: `breaker.go` emits transition and reject events
- Given a `BreakerSet` with `BreakerFailThreshold=2` and Debug logger
- When two failures are recorded for backend `"ollama"`
- Then the buffer contains a `breaker.transition` record with `from="closed"`, `to="open"`, `backend="ollama"`
- When a third request arrives while the breaker is open
- Then the buffer contains a `breaker.reject` record with `backend="ollama"`, `state="open"`

AC-3.5: `semaphore.go` emits wait and acquired events
- Given a `semaphore` of capacity 1 with a Debug logger and a permit already held
- When a second goroutine calls `acquire(ctx)` and the permit is released after 5ms
- Then the buffer contains a `semaphore.wait` record with `wait_ms >= 5`
- And the buffer contains a `semaphore.acquired` record after the wait
- Given the permit is held and the `ctx` is cancelled before release
- When the second goroutine's `acquire` returns false
- Then the buffer contains a `semaphore.cancelled` record

AC-3.6: `prefix_cache.go` emits build events
- Given a Debug-level logger
- When `ClaudeSystemBlock("hello", cfg, logger)` and `ClaudeSystemBlockJSON("hello", cfg, logger)` are called
- Then the buffer contains a `prefix_cache.build` record with `prefix_len=5` and a `ttl` attr
- And the JSON variant additionally emits `prefix_cache.build.success` with `bytes_out`

AC-3.7: Sanitize discipline holds across all five files
- Given any 100-iteration smoke test exercising every instrumented site with random raw payloads
- When the captured JSON log buffer is scanned
- Then no log record contains any byte of the raw payload, the response body, or any error message text
- And every value-bearing attr is one of: integer, float, boolean, enum string, hash prefix, or duration

AC-3.8: Hot-path guard prevents allocation when Debug is off
- Given a logger at level Info (Debug filtered)
- When the request path runs 10,000 iterations through the instrumented surface
- Then `testing.AllocsPerRun` reports zero additional heap allocations attributable to the new instrumentation
- And the captured buffer is empty (no Debug records leaked)

## BDD Test Scenarios

### Scenario 1: Client request lifecycle

```gherkin
Feature: client.go debug instrumentation

  Scenario: Successful chat round-trip
    Given a Client pointed at a stub HTTP server that returns 200 with body "{}"
    And a Debug-level logger writing to bytes.Buffer
    When I call Client.Chat(ctx, "gemma3:4b", "system", "raw input")
    Then the buffer contains exactly one record with msg="client.chat.start"
    And that record has op="client.chat", model="gemma3:4b", bytes_in=9
    And the buffer contains exactly one record with msg="client.chat.response"
    And that record has bytes_out=2 and a non-zero latency_ms
    And no record contains the substring "raw input" or "system"

  Scenario: Transport failure
    Given a Client pointed at 127.0.0.1:1 (closed)
    And a Debug-level logger
    When I call Chat
    Then a record with msg="client.chat.transport_error" is emitted
    And its reason attribute is one of "transport", "timeout"
    And no attribute carries the raw err.Error() text

  Scenario: HTTP 500
    Given a Client pointed at a stub returning 500
    When I call Chat
    Then a record with msg="client.chat.http_error" and status_code=500 is emitted
```

### Scenario 2: Cache eviction trail

```gherkin
Feature: cache.go LRU eviction telemetry

  Scenario: Capacity-2 cache evicts the oldest
    Given a ResponseCache with capacity 2 and Debug logger
    When I Put("k1","v1"), Put("k2","v2"), Get("k1"), Put("k3","v3")
    Then the buffer JSON-decodes to records in order:
      | msg          |
      | cache.put    |
      | cache.put    |
      | cache.hit    |
      | cache.evict  |
      | cache.put    |
    And the cache.evict record has reason="lru" and cache_key_hash matching hashKey("k2")

  Scenario: Miss on absent key
    Given an empty Debug-level cache
    When I Get("missing")
    Then a single cache.miss record is emitted with cache_key_hash matching hashKey("missing")
```

### Scenario 3: Breaker and semaphore concurrency

```gherkin
Feature: breaker and semaphore observability

  Scenario: Breaker opens after threshold failures
    Given a BreakerSet with FailThreshold=2 for backend "ollama"
    And a Debug-level logger
    When I record 2 consecutive failures via the breaker API
    Then a breaker.transition record with from="closed",to="open",backend="ollama" is emitted
    When I attempt a request while open
    Then a breaker.reject record with state="open" is emitted

  Scenario: Semaphore acquisition waits and succeeds
    Given a semaphore of capacity 1 with the permit held
    And a Debug-level logger
    When goroutine G calls acquire(ctx) and the permit is released after >= 5ms
    Then a semaphore.wait record with wait_ms >= 5 is emitted
    And a semaphore.acquired record follows it

  Scenario: Semaphore acquisition cancels via context
    Given a semaphore of capacity 1 with permit held
    When goroutine G calls acquire with a ctx that cancels after 2ms
    Then a semaphore.cancelled record with wait_ms ~= 2 is emitted
    And acquire returns false
```

### Scenario 4: Hot-path zero-allocation

```gherkin
Feature: Hot-path guard prevents alloc when level filtered

  Scenario: Info-level logger emits nothing and allocates nothing
    Given a logger at slog.LevelInfo
    When I run testing.AllocsPerRun(10000) over a closure that exercises Chat, cache.Get, and breaker checks
    Then the average allocations-per-run is unchanged from the pre-instrumentation baseline
    And the captured buffer is empty
```

## Tasks / Subtasks

- [ ] Task 1: Add `hashKey` helper and standard attr scaffolding (AC: 3.3, 3.6, 3.7)
  - [ ] Subtask 1a: Add `hashKey(string) string` to `cache.go` (8-char hex SHA-256 prefix).
  - [ ] Subtask 1b: Add `// Standard attrs: ...` doc comment cross-referencing `logging.go`.

- [ ] Task 2: Instrument `client.go` (AC: 3.1, 3.2, 3.7, 3.8)
  - [ ] Subtask 2a: Add `client.chat.start` and `client.chat.response` Debug calls in `Chat` and `ChatDeterministic`.
  - [ ] Subtask 2b: Add `client.chat.transport_error` and `client.chat.http_error` calls along the error paths; ensure `reason` enum, never raw err text.
  - [ ] Subtask 2c: Wrap every site in the `Enabled` guard.

- [ ] Task 3: Instrument `cache.go` (AC: 3.3, 3.7, 3.8)
  - [ ] Subtask 3a: Emit `cache.hit` / `cache.miss` in Get.
  - [ ] Subtask 3b: Emit `cache.put` in Put; emit `cache.evict` with `reason="lru"` at the LRU eviction call site.
  - [ ] Subtask 3c: Confirm every emission uses `cache_key_hash` (never raw key).

- [ ] Task 4: Instrument `prefix_cache.go` (AC: 3.6, 3.7)
  - [ ] Subtask 4a: Emit `prefix_cache.build` in `ClaudeSystemBlock` with `prefix_len`, `ttl`.
  - [ ] Subtask 4b: Emit `prefix_cache.build.success` in `ClaudeSystemBlockJSON` with `bytes_out`.
  - [ ] Subtask 4c: Emit `prefix_cache.ollama_keep_alive` in `OllamaKeepAliveEncoded`.

- [ ] Task 5: Instrument `breaker.go` (AC: 3.4, 3.7)
  - [ ] Subtask 5a: At every state transition, emit `breaker.transition` with `from`, `to`, `backend`.
  - [ ] Subtask 5b: At reject path, emit `breaker.reject` with `backend`, `state="open"`.
  - [ ] Subtask 5c: Emit inside the lock so transitions are consistent.

- [ ] Task 6: Instrument `semaphore.go` (AC: 3.5, 3.7, 3.8)
  - [ ] Subtask 6a: In `acquire`, capture `start := time.Now()`; on `<-ctx.Done()` emit `semaphore.cancelled`; on permit acquisition, emit `semaphore.wait` (only if `wait_ms > 0`) and `semaphore.acquired`.
  - [ ] Subtask 6b: Use `cap(s.ch) - len(s.ch)` for `in_flight`.

- [ ] Task 7: Tests (AC: 3.1–3.8)
  - [ ] Subtask 7a: Add a `testLogBuffer(t)` helper to `logging_test.go` (or a shared test helper file) that returns a buffer and a JSON-handler logger.
  - [ ] Subtask 7b: Augment `client_test.go`, `cache_test.go`, `breaker_test.go` (no semaphore_test.go exists — create one), and `prefix_cache_test.go` with assertions matching the AC tables.
  - [ ] Subtask 7c: Add an alloc-budget test: `BenchmarkTranslate_DebugOff_Allocs` or `TestZeroAllocsWhenInfo` using `testing.AllocsPerRun`.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on the five instrumented files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
