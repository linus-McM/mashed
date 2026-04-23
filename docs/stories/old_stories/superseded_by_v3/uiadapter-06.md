# uiadapter-06: Timeout + keep_alive tuning and warm-up probe

**Status:** ready
**Domain:** backend
**Size:** M
**Depends On:** uiadapter-01, uiadapter-02, uiadapter-03
**Priority:** P2-medium

## Story

As a UIAdapter maintainer, I want a 5-second default timeout, a 30-minute Ollama `keep_alive`, a non-blocking warm-up probe at `NewDefault` time, and split cold/warm latency telemetry, so that cold-start inference (2–4s on a 4B model) no longer falls through to the fallback under a too-tight 2.9s budget, and operators can distinguish a cold first turn from warm steady-state latency.

## Description

Cold-load on `gemma3:4b` is 2–4 seconds; warm inference is 300–800 ms. The current default timeout (2.9s) guarantees cold-start misses that degrade to fallback unnecessarily. Raise the default to 5000 ms (matching `eval_test.go:60`), ask Ollama to keep the model resident for 30 minutes via `options.keep_alive`, fire a non-blocking warm-up probe at construction, and add a `first_turn bool` attribute to the existing `uiadapter.translate` slog line so cold-vs-warm latency can be measured separately. Runs last — its latency target depends on stories 1–3 reducing validation churn first.

### Scope summary
- `Config.TimeoutMs` default: `2900` → `5000`.
- `chatRequest.Options["keep_alive"] = "30m"` in the outgoing JSON body.
- New warm-up probe: `NewDefault` dispatches `Chat(ctx, model, "", "warm")` in a goroutine with a 10-second timeout; errors are swallowed (best-effort).
- Split telemetry: add `first_turn bool` attribute on the `uiadapter.translate` slog line; keep `latency_ms` and add it so cold-path calls are distinguishable from warm-path ones.
- New tests: `TestAdapter_WarmupDoesNotBlockConstruction`, `TestAdapter_RequestIncludesKeepAlive`.

### Non-goals
- Do not change the Ollama client retry logic.
- Do not add a request-level timeout override field (use `Config.TimeoutMs`).
- Do not change the widget schema or prompt.
- Do not change the adapter pattern (long-lived — confirmed in `adapter.go:66` per plan §1 Story 6).

## Developer Notes

### Files to modify
- `internal/uiadapter/client.go` — `chatRequest.Options` gains `keep_alive`; consider introducing a small `chatOptions` struct if helpful.
- `internal/uiadapter/adapter.go` — raise default `Config.TimeoutMs` to 5000; add warm-up goroutine at the end of `NewDefault`; thread a `first_turn` flag via adapter state (e.g. `sync.Once` or atomic bool); add `first_turn` to the slog line.
- `internal/uiadapter/adapter_test.go` — new tests for warm-up non-blocking behaviour and `keep_alive` in captured request.

### Type/symbol inventory (exact names)
- `Config.TimeoutMs int` — default changes from 2900 to 5000 (confirm current default value and update; declare as an exported constant `DefaultTimeoutMs = 5000` if one exists).
- `chatRequest.Options map[string]any` (or typed struct) — add `keep_alive` key with string value `"30m"`.
- Warm-up goroutine — fire-and-forget; use `context.WithTimeout(ctx, 10*time.Second)`.
- `Adapter.firstTurn *atomic.Bool` (or equivalent) — flipped from true→false on the first successful `Translate` invocation.
- Telemetry: `slog.Bool("first_turn", isFirst)` added to the existing `uiadapter.translate` line (single-line invariant preserved).

### Warm-up non-blocking contract
- Plan §1 Story 6 AC-6.2: `NewDefault` returns in <50 ms even if Ollama is unreachable.
- Implementation: goroutine launched via `go func() { ... }()` with an internally-bounded 10-second context. No `sync.WaitGroup.Wait()` in `NewDefault`. Errors are logged at `Debug` only (or swallowed).
- Test strategy: use a stub HTTP transport that hangs indefinitely; assert `NewDefault` returns within 50 ms.

### `first_turn` flag semantics
- True on the first `Translate` call of an adapter instance, false on every subsequent call.
- Thread-safe: use `sync/atomic.Bool` or `sync.Once` + a read-flag.
- The warm-up probe itself does NOT set `first_turn=false` — only user-visible `Translate` calls do.
- Latency telemetry: keep the existing `latency_ms` attribute; splitting is via `first_turn`, not a new attribute name.

### Default-value migration
- Where is `Config.TimeoutMs` defaulted? Grep for `2900` under `internal/uiadapter/`. Plan references `eval_test.go:60` uses 5000 already — align the default so production behaviour matches eval conditions.
- If a `DefaultTimeoutMs` constant exists, update it. If the default is inline in `NewDefault`, extract it to a constant in this story for reviewability.

### Request body shape (for AC-6.3 test)
- After this story, every chat request's JSON body must include:
  ```json
  {
    ...,
    "options": {
      ...,
      "keep_alive": "30m"
    }
  }
  ```
- Other existing `options` keys (temperature, num_predict, etc.) must be preserved — plan §3 "additive only".

### Risks / gotchas
- Warm-up adds one free Ollama call per process start. Plan §1 Story 6 risk note: fine for tmux use-case (one adapter per tmux session), confirmed long-lived adapter pattern at `adapter.go:66`.
- If Ollama is unreachable, the warm-up goroutine will fail silently after its 10-second timeout. Do NOT retry; do NOT block construction. Log at Debug only.
- `keep_alive` as a string `"30m"` — Ollama accepts duration strings and numeric seconds. String is more readable and survives JSON-roundtrip cleanly.
- The `first_turn` atomic must be an instance field, not package-global — multiple adapters per process must each track their own first turn.
- Story 3's `sanitize_delta_bytes` attribute must not be removed when adding `first_turn`; both coexist on the same slog line.

### Telemetry (single-line invariant)
Plan §3: "Preserve the single structured `uiadapter.translate` log line per Translate call. New attributes are additive only."
- Existing attributes remain (after Story 3 lands: includes `sanitize_delta_bytes`).
- Add `first_turn` (bool). Do NOT emit a second log line for cold vs. warm.

### Reference files
- `internal/uiadapter/adapter.go` — `NewDefault`, `Config`, `Translate`, the existing single slog line, confirmed long-lived adapter at line 66.
- `internal/uiadapter/client.go` — `ClientConfig`, `chatRequest`, `Client.chat`.
- `internal/uiadapter/eval_test.go:60` — the reference 5000 ms timeout.
- `docs/plans/IMPLEMENTATION_PLAN.md` §1 Story 6.

## Acceptance Criteria

**AC-6.1: Warm eval latency median <1.5 s**
- Given Story 1/2/3 shipped and this story's warm-up probe active
- When `go test -tags ollama_eval ./internal/uiadapter/... -run TestEval_FullCorpus_MeetsThresholds` runs
- Then the median `latency_ms` across the corpus is <1500
- Note: statistical target — reproduce locally; document the median in the PR description.

**AC-6.2: `NewDefault` returns in <50 ms even when Ollama is unreachable**
- Given a stubbed HTTP transport that hangs (or resolves a loopback to a port with no listener)
- When `adapter.NewDefault(cfg)` is called
- Then the call returns within 50 ms
- And no goroutine leaks after the warm-up's 10-second timeout

**AC-6.3: `keep_alive` appears in every captured chat request body**
- Given a `Client` dispatching a request to a stubbed transport
- When the request body is unmarshalled
- Then `body.options.keep_alive == "30m"`
- And other pre-existing `options` keys remain intact

**AC-6.4: Default `Config.TimeoutMs` is 5000**
- Given `adapter.NewDefault(Config{})` (zero value)
- When the resulting adapter's timeout is inspected
- Then it equals 5000 ms

**AC-6.5: `first_turn` bool attribute on the translate slog line**
- Given a fresh adapter instance and a slog test handler
- When `Translate` is invoked twice in succession (both successful)
- Then the first `uiadapter.translate` record has `first_turn == true`
- And the second record has `first_turn == false`
- And no additional log lines per Translate call are emitted

## BDD Test Scenarios

### Scenario 1: Non-blocking warm-up

```gherkin
Feature: Warm-up probe does not block construction

  Scenario: Unreachable Ollama does not stall NewDefault
    Given an HTTP transport that blocks indefinitely
    When adapter.NewDefault is called with a 10ms stopwatch
    Then the call returns within 50 milliseconds
    And no warm-up error is propagated to the caller

  Scenario: Successful warm-up is silent on the happy path
    Given a stubbed transport returning 200 OK for a "warm" payload
    When NewDefault runs
    Then NewDefault returns within 50 ms
    And a Debug-level warm-up success is optionally logged
```

### Scenario 2: keep_alive on the wire

```gherkin
Feature: keep_alive request option

  Scenario: Every chat request carries keep_alive
    Given a default Client and a stubbed transport capturing the body
    When Client.chat is invoked
    Then body.options.keep_alive equals "30m"

  Scenario: Pre-existing options keys are preserved
    Given a Client configured with custom temperature
    When Client.chat is invoked
    Then body.options.temperature is preserved
    And body.options.keep_alive equals "30m"
```

### Scenario 3: Timeout default

```gherkin
Feature: 5-second default timeout

  Scenario: Zero-value Config defaults to 5000ms
    Given adapter.NewDefault called with Config{}
    When the adapter's configured timeout is inspected
    Then it equals 5000 milliseconds
```

### Scenario 4: first_turn telemetry

```gherkin
Feature: Cold-vs-warm telemetry via first_turn attribute

  Scenario: First Translate emits first_turn=true, subsequent are false
    Given a fresh adapter and a slog test handler
    When Translate is called twice back-to-back
    Then the first uiadapter.translate record has first_turn=true
    And the second has first_turn=false
    And exactly two records are emitted
```

## Tasks / Subtasks

- [ ] Task 1: Raise default timeout (AC: 6.4)
  - [ ] Locate the current default (grep for `2900` under `internal/uiadapter/`)
  - [ ] Update to 5000 (prefer an exported `DefaultTimeoutMs` constant)
  - [ ] Update any test that asserts the old default
- [ ] Task 2: Add `keep_alive` to every chat request (AC: 6.3)
  - [ ] Extend `chatRequest.Options` to include `keep_alive: "30m"`
  - [ ] Preserve all pre-existing options keys
  - [ ] Add `TestAdapter_RequestIncludesKeepAlive` asserting the captured body
- [ ] Task 3: Add non-blocking warm-up probe (AC: 6.2)
  - [ ] Launch a goroutine at the end of `NewDefault`
  - [ ] Use `context.WithTimeout(ctx, 10*time.Second)`
  - [ ] Swallow errors (or log at Debug only)
  - [ ] Add `TestAdapter_WarmupDoesNotBlockConstruction` using a hanging stub transport
- [ ] Task 4: Thread `first_turn` flag through Translate (AC: 6.5)
  - [ ] Add `firstTurn atomic.Bool` to the adapter struct
  - [ ] On `Translate` entry, read-and-clear with `CompareAndSwap(true, false)`
  - [ ] Add `slog.Bool("first_turn", isFirst)` to the existing `uiadapter.translate` line
  - [ ] Preserve single-line invariant (do NOT emit a second line)
- [ ] Task 5: Document warm latency expectation (AC: 6.1)
  - [ ] Run the eval harness once locally after this story lands
  - [ ] Paste the median latency into the PR description
- [ ] Task 6: Pre-flight and handoff (AC: all)
  - [ ] `go build ./...`, `go vet ./...`
  - [ ] `go test ./internal/uiadapter/... -race -short`
  - [ ] `/simplify` on `adapter.go` and `client.go`
  - [ ] Goroutine-leak check: the 10-second warm-up goroutine must exit by the end of any test that creates an adapter

## Definition of Done

- [ ] All acceptance criteria pass (AC-6.1 captured in PR description)
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on modified lines of `adapter.go` and `client.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./internal/uiadapter/... -race -short` passes
- [ ] `/simplify` run on all modified code
- [ ] Pre-flight: warm-up non-blocking (<50 ms), keep_alive on wire, default timeout 5000, first_turn attribute present
- [ ] AC Validation Table complete

### AC Validation Table (fill in PR description)

| AC | Test / Evidence | Status |
|----|-----------------|--------|
| AC-6.1 | Median latency from eval harness in PR description | |
| AC-6.2 | `TestAdapter_WarmupDoesNotBlockConstruction` | |
| AC-6.3 | `TestAdapter_RequestIncludesKeepAlive` | |
| AC-6.4 | Unit assertion on default `Config.TimeoutMs` | |
| AC-6.5 | Slog capture on paired `Translate` calls | |
