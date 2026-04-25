# Story 2: Plumb logger through uiadapter subcomponent constructors

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** L
**Depends On:** uiadapter-logging-1-infrastructure-and-boot
**Status:** ready

## Description

Refactor every uiadapter subcomponent constructor and free function in the request path to accept a `*slog.Logger` parameter. Each constructor must nil-guard so callers (especially existing tests) can continue to pass `nil` and get an `io.Discard`-backed logger. Wire `NewDefault` to scope a single logger via `slog.WithGroup("uiadapter")` and inject it into every subcomponent. **No actual log calls land in this story** — only plumbing — so the diff is mechanical and the test suite must stay green throughout.

## Developer Notes

### Architecture

This is a constructor-shape change across the package. The end-state every subsequent instrumentation story relies on:

| Subcomponent | File | Constructor / function | New shape |
|---|---|---|---|
| `Client` | `client.go:51` | `NewClient(cfg ClientConfig) *Client` | `NewClient(cfg ClientConfig, logger *slog.Logger) *Client` |
| `ResponseCache` | `cache.go:34` | `NewResponseCache(cfg Config) *ResponseCache` | `NewResponseCache(cfg Config, logger *slog.Logger) *ResponseCache` |
| `BreakerSet` | `breaker.go:28` | `NewBreakerSet(cfg Config) *BreakerSet` | `NewBreakerSet(cfg Config, logger *slog.Logger) *BreakerSet` |
| `semaphore` | `semaphore.go:12` | `newSemaphore(n int) semaphore` | `newSemaphore(n int, logger *slog.Logger) semaphore` (returns a struct now — see below) |
| `Repairer` | `repair.go:69` | `NewRepairer(cfg Config) *Repairer` | `NewRepairer(cfg Config, logger *slog.Logger) *Repairer` |
| `FastPathClassifier` | `fastpath.go:35` | `NewFastPathClassifier(enabled bool) *FastPathClassifier` | `NewFastPathClassifier(enabled bool, logger *slog.Logger) *FastPathClassifier` |
| `ContextGuard` | `contextguard.go:25` | `NewContextGuard(cfg Config) *ContextGuard` | `NewContextGuard(cfg Config, logger *slog.Logger) *ContextGuard` |
| `MockAdapter` | `mock.go:17` | `NewMock(fixed *UIAST) Adapter` | `NewMock(fixed *UIAST, logger *slog.Logger) Adapter` |

**Free functions** in the request path that need a `logger *slog.Logger` last positional parameter:

| File | Function |
|---|---|
| `sanitize.go:16` | `SanitizeCapture(raw string, logger *slog.Logger) (sanitized string, deltaBytes int)` |
| `spotlight.go:26` | `Spotlight(raw string, enabled bool, logger *slog.Logger) string` |
| `validator.go:40` | `Validate(ast *UIAST, raw string, logger *slog.Logger) []string` |
| `stages.go:124` | `RunTwoStage(...args..., logger *slog.Logger)` (append to existing signature) |
| `fallback_tiers.go:35` | `RunWithFallback(...args..., logger *slog.Logger)` |
| `prefix_cache.go:16,61` | `ClaudeSystemBlock(staticPrefix string, cfg Config, logger *slog.Logger)`, `ClaudeSystemBlockJSON(...)` |
| `encode.go:19,33` | `OllamaFormatPayload(kind StageKind, loose bool, logger *slog.Logger)`, `ClaudeToolInputSchema(kind StageKind, logger *slog.Logger)` |
| `sampling.go:10,24` | `OllamaSamplingOptions(cfg Config, logger *slog.Logger)`, `ClaudeSamplingOptions(cfg Config, logger *slog.Logger)` |
| `repair.go:39` | `BuildRepairPrompt(a RepairAttempt, logger *slog.Logger)` |
| `fallback.go:14` | `FallbackAST(raw, reason string, logger *slog.Logger)` |
| `allowlist.go:31` | already takes `logger *slog.Logger` — no change |

> **Decision required for `semaphore`**: today `semaphore` is a `chan struct{}` type alias. Adding a logger requires promoting it to a struct `type semaphore struct { ch chan struct{}; logger *slog.Logger }`. Update `acquire` / `release` accordingly. Keep callers compiling by ensuring all field reads continue to work.

**Single source of truth for logger plumbing**: `NewDefault` in `adapter.go:118` must:
1. Apply the existing nil-guard (already present).
2. Scope the logger: `scoped := logger.With(slog.Group("uiadapter"))` — this gives every subcomponent's records a `uiadapter.*` namespace.
3. Pass `scoped` (or a child `scoped.With("component", "<name>")`) to every constructor it instantiates.

The `defaultAdapter` struct already has a `logger` field. Add equivalent fields to each refactored subcomponent struct (e.g. `Client.logger`, `ResponseCache.logger`, `BreakerSet.logger`, `Repairer.logger`, `FastPathClassifier.logger`, `ContextGuard.logger`).

### Technical Considerations

- **Universal nil-guard helper:** add a private helper to `logging.go` (already created in Story 1):
  ```go
  func nilSafeLogger(l *slog.Logger) *slog.Logger {
      if l == nil {
          return slog.New(slog.NewTextHandler(io.Discard, nil))
      }
      return l
  }
  ```
  Every constructor calls this on entry and stores the non-nil result. `adapter.go:NewDefault` already inlines this; replace the inlined block with a call to the helper.
- **No call-site behaviour change**: this story must NOT add any `logger.Debug(...)` calls. The goal is to make the suite still green after the constructor reshape so subsequent stories can purely add log calls.
- **Test fallout**: every existing `*_test.go` that constructs a subcomponent must be updated to pass `nil`. Use `gofmt -r` or a sed pass with manual review. Examples to expect: `cache_test.go`, `breaker_test.go`, `client_test.go`, `repair_test.go`, `fastpath_test.go`, `contextguard_test.go`, `mock_test.go`, `sanitize_test.go`, `spotlight_test.go`, `validator_test.go`, `stages_test.go`, `fallback_tiers_test.go`, `encode_test.go`, `sampling_test.go`.
- **Wails binding:** none — this remains a pure-Go internal change.
- **`adapter.go` adjustments:** `NewDefault` already constructs `NewClient(ClientConfig{...})` and `newSemaphore(cfg.MaxInflight)`. Update those two call sites to pass the scoped logger. Other subcomponents are constructed elsewhere (router / accountant / etc. — handled in their own call sites).

### Risks & Edge Cases

- **`semaphore` type promotion** is the largest mechanical risk. `semaphore` currently used as `chan struct{}` in receivers (`acquire`, `release`). After promotion, the receiver methods read `s.ch` instead. Audit every call site (`grep -n "semaphore" internal/uiadapter`) and fix.
- **Test breakage**: aim for zero-behaviour-change. Run `go test ./internal/uiadapter/... -race` after each constructor-by-constructor commit. Keep PRs small (one constructor cluster per commit) so revert is cheap.
- **Forward compatibility**: subsequent stories will add fields to constructor configs (e.g. router accountant). The logger param is **always last** so future additions don't perturb arg order.
- **Backwards-compat shim**: do NOT keep wrapper `NewClientNoLogger(cfg)` shims. The package is internal — break callers, fix them in the same PR.
- **`allowlist.go` already takes a logger** — leave it untouched in this story; its existing nil handling (calls `slog.Default()` when nil) stays. We will tighten that in the instrumentation story.

### Reference Files

- `/Users/linus/Development/mashed/internal/uiadapter/adapter.go:118-136` — `NewDefault` template; this is the canonical "scope logger then inject" pattern to follow.
- `/Users/linus/Development/mashed/internal/uiadapter/adapter.go:119-121` — existing nil-guard inline; replace with `nilSafeLogger`.
- `/Users/linus/Development/mashed/internal/uiadapter/semaphore.go` — type promotion required; smallest file but highest blast radius.
- `/Users/linus/Development/mashed/internal/uiadapter/allowlist.go:31` — only existing constructor that already takes a logger; mirrors the desired final shape.

### Skills

- Invoke `/simplify` after the constructor reshape — this is exactly the kind of mechanical change where simplify catches accidental complexity (extra wrappers, unused fields).
- Invoke `/golang-testing` to confirm test files use the lowest-friction pattern: `nil` for the logger arg in unit tests; production code always uses the scoped logger.

## Acceptance Criteria

AC-2.1: Every subcomponent constructor accepts a `*slog.Logger`
- Given the package public/private constructor surface
- When a developer reads each of the 8 constructors listed in the table above
- Then each one's signature ends with `logger *slog.Logger` (last positional parameter)
- And every free function in the request path also has `logger *slog.Logger` as its last parameter

AC-2.2: `nil` logger is always safe
- Given a test caller passes `nil` for the logger argument
- When any constructor or free function listed above is invoked
- Then no panic occurs
- And the underlying logger writes to `io.Discard` (no stdout output, no file output)
- And `logger.Enabled(ctx, slog.LevelDebug)` returns `false` for the discard logger

AC-2.3: `NewDefault` scopes a single logger via `WithGroup("uiadapter")`
- Given `NewDefault(cfg, parentLogger)` is invoked with a non-nil parent
- When the resulting adapter is exercised
- Then every subcomponent it constructs receives a logger derived from `parentLogger.WithGroup("uiadapter")`
- And inspecting any log record emitted by a subcomponent in a later story will show the `uiadapter` group prefix

AC-2.4: All existing tests pass with `nil` loggers
- Given the existing `internal/uiadapter/*_test.go` suite
- When `go test ./internal/uiadapter/... -race` runs after the refactor
- Then every test passes
- And no test depends on a non-nil logger being injected

AC-2.5: `nilSafeLogger` helper is the single nil-guard
- Given any constructor or free function in the package
- When that function receives a `nil` logger
- Then it normalises the value through `nilSafeLogger` (or an exact equivalent)
- And the inline `if logger == nil { logger = slog.New(slog.NewTextHandler(io.Discard, nil)) }` block from `adapter.go:119` is replaced by the helper call

AC-2.6: No new log statements ship in this story
- Given a code review of the diff
- When the reviewer greps for new `logger.Debug`, `logger.Info`, `logger.Warn`, `logger.Error`, or `logger.LogAttrs` call sites
- Then zero new call sites exist (other than tests asserting nil-safety)
- And `internal/uiadapter` behaviour is byte-identical to the prior commit

## BDD Test Scenarios

### Scenario 1: Constructor nil-safety

```gherkin
Feature: Subcomponent constructors accept nil logger

  Scenario Outline: Constructor with nil logger does not panic
    Given a default Config
    When I call <Constructor> with logger=nil
    Then no panic occurs
    And the returned value is non-nil
    And calling its public methods with valid args succeeds

    Examples:
      | Constructor                |
      | NewClient                  |
      | NewResponseCache           |
      | NewBreakerSet              |
      | newSemaphore               |
      | NewRepairer                |
      | NewFastPathClassifier      |
      | NewContextGuard            |
      | NewMock                    |

  Scenario: Free functions accept nil logger
    Given a non-empty raw capture
    When I call SanitizeCapture(raw, nil), Spotlight(raw, true, nil), Validate(ast, raw, nil)
    Then each returns its documented value
    And no panic occurs
```

### Scenario 2: NewDefault wires the scoped logger

```gherkin
Feature: NewDefault propagates a uiadapter-scoped logger

  Scenario: Parent logger is scoped via WithGroup
    Given a parent *slog.Logger built over a captured bytes.Buffer
    And cfg.Enabled = true with default fields
    When I construct adapter := NewDefault(cfg, parent)
    And I call adapter.Translate(ctx, "raw", "proc-1")
    And the request path emits at least one log entry (via a stub debug call in this test only)
    Then the captured buffer contains records with the "uiadapter" group prefix

  Scenario: NewDefault with nil parent uses io.Discard
    Given a nil parent logger
    When I construct adapter := NewDefault(cfg, nil)
    Then no panic occurs
    And the adapter is operational
    And no output reaches stdout from package code
```

### Scenario 3: Existing test suite stays green

```gherkin
Feature: Refactor preserves behaviour

  Scenario: Full package suite passes with -race
    Given the refactor is complete
    When `go test ./internal/uiadapter/... -race` runs
    Then every test passes
    And no test was deleted or skipped to make this true

  Scenario: No new log calls land in production code
    Given a grep for `\.Debug\(|\.Info\(|\.Warn\(|\.Error\(|\.LogAttrs\(` over the diff
    When the reviewer compares the count of matches in production .go files before and after
    Then the difference is zero (excluding the existing allowlist.go and adapter.go logTelemetry sites)
```

## Tasks / Subtasks

- [ ] Task 1: Add `nilSafeLogger` helper (AC: 2.2, 2.5)
  - [ ] Subtask 1a: Add `nilSafeLogger(*slog.Logger) *slog.Logger` to `internal/uiadapter/logging.go`.
  - [ ] Subtask 1b: Replace the inline nil guard in `adapter.go:119-121` with a call to `nilSafeLogger`.

- [ ] Task 2: Promote `semaphore` to a struct (AC: 2.1, 2.4)
  - [ ] Subtask 2a: Edit `internal/uiadapter/semaphore.go` — define `type semaphore struct { ch chan struct{}; logger *slog.Logger }`.
  - [ ] Subtask 2b: Update `acquire` / `release` to read `s.ch`.
  - [ ] Subtask 2c: Update `newSemaphore` to take `(n int, logger *slog.Logger)` and call `nilSafeLogger`.
  - [ ] Subtask 2d: Update every caller (notably `adapter.go:132`).

- [ ] Task 3: Reshape struct constructors (AC: 2.1, 2.2, 2.5)
  - [ ] Subtask 3a: `NewClient` (`client.go:51`) — add `logger *slog.Logger` and `Client.logger` field.
  - [ ] Subtask 3b: `NewResponseCache` (`cache.go:34`) — add logger param + field.
  - [ ] Subtask 3c: `NewBreakerSet` (`breaker.go:28`) — add logger param + field.
  - [ ] Subtask 3d: `NewRepairer` (`repair.go:69`) — add logger param + field.
  - [ ] Subtask 3e: `NewFastPathClassifier` (`fastpath.go:35`) — add logger param + field.
  - [ ] Subtask 3f: `NewContextGuard` (`contextguard.go:25`) — add logger param + field.
  - [ ] Subtask 3g: `NewMock` (`mock.go:17`) — add logger param + field on `MockAdapter`.

- [ ] Task 4: Reshape free functions (AC: 2.1, 2.2)
  - [ ] Subtask 4a: `SanitizeCapture`, `Spotlight`, `Validate`, `RunTwoStage`, `RunWithFallback`, `BuildRepairPrompt`, `FallbackAST` — append `logger *slog.Logger` last; nil-guard via `nilSafeLogger`; do not yet emit.
  - [ ] Subtask 4b: `ClaudeSystemBlock`, `ClaudeSystemBlockJSON`, `OllamaFormatPayload`, `ClaudeToolInputSchema`, `OllamaSamplingOptions`, `ClaudeSamplingOptions` — same treatment.

- [ ] Task 5: Wire `NewDefault` to inject the scoped logger (AC: 2.3)
  - [ ] Subtask 5a: In `adapter.go:NewDefault`, after nil-guard, build `scoped := logger.WithGroup("uiadapter")`.
  - [ ] Subtask 5b: Pass `scoped` to `NewClient`, `newSemaphore`, and any other constructor invoked here.
  - [ ] Subtask 5c: Add a unit test in `adapter_test.go` that asserts the buffer-captured logger sees `uiadapter` group records (use a stub `t.Cleanup`-installed test handler).

- [ ] Task 6: Update every existing test to pass `nil` (AC: 2.4)
  - [ ] Subtask 6a: Audit each `internal/uiadapter/*_test.go` and update constructor / free-function calls.
  - [ ] Subtask 6b: Run `go test ./internal/uiadapter/... -race` and fix any compile or behavioural regressions.

- [ ] Task 7: Diff guard (AC: 2.6)
  - [ ] Subtask 7a: Document in the PR description that no new `Debug/Info/Warn/Error/LogAttrs` calls were added to production code (allowed exceptions: the new `nilSafeLogger`-using stubs and the test handler).

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/uiadapter/*.go` (no regression vs prior baseline)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
