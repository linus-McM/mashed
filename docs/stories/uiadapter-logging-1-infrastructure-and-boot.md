# Story 1: UI Adapter logging infrastructure and boot wiring

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Introduce the foundational logging infrastructure for `internal/uiadapter`: a new `logging.go` file that exposes `NewProductionLogger(level, logDir) (*slog.Logger, io.Closer, error)`, a fan-out `slog.Handler` that emits human-readable text to stdout AND structured JSON to `./logs/uiadapter-YYYYMMDD.log`, plus the boot wiring in `app.go` and `main.go` so the production logger replaces `slog.Default()` and is closed cleanly on shutdown. No behaviour or instrumentation changes ship in this story — only infrastructure and the first wired entry point. This is the foundation every subsequent logging story builds on.

## Developer Notes

### Architecture

- **New file:** `internal/uiadapter/logging.go` (~60 LOC).
  - Export: `func NewProductionLogger(level slog.Level, logDir string) (*slog.Logger, io.Closer, error)`.
  - Internal: a `fanoutHandler` implementing `slog.Handler` (`Enabled`, `Handle`, `WithAttrs`, `WithGroup`) that dispatches each record to a stdout `slog.NewTextHandler` AND a file-backed `slog.NewJSONHandler`.
  - The returned `io.Closer` closes the underlying log file. Closing must be idempotent.
  - Daily rollover is achieved by deriving the file name `uiadapter-YYYYMMDD.log` from `time.Now().UTC()` at construction. No mid-process rotation.
  - Default level fallback: if the caller passes an unparseable env value, default to `slog.LevelInfo`.
- **Edited file:** `internal/uiadapter/config.go`.
  - Add two fields to `Config` (after `DisableHealthTicker`):
    - `LogLevel string` — `"debug" | "info" | "warn" | "error"` (default `"info"`).
    - `LogDir string` — default `"./logs"`.
  - Extend `DefaultConfig()` with these defaults.
  - `mergeWithDefaults` already handles strings via reflection — no extra change required, but verify both fields are picked up by the existing reflection pass.
- **Edited file:** `app.go` around line 283 (where `uiadapter.NewDefault(...)` is called).
  - Build the production logger before constructing the adapter.
  - Read `UIADAPTER_LOG_LEVEL` env var (`os.Getenv`); parse it through a small helper `parseSlogLevel(s string) slog.Level` (kept in `logging.go`). Empty / unparseable -> `slog.LevelInfo`.
  - Pass the logger to `uiadapter.NewDefault`.
  - Register the returned `io.Closer` in a new `App.shutdownHooks []func() error` slice (append it). Drain & call all hooks inside the existing `(*App).shutdown` function. Errors logged with `log.Printf` only — never block shutdown.
- **Edited file:** `main.go`.
  - Before Wails `Run()`, ensure log dir exists: `os.MkdirAll("./logs", 0o755)`. Failure to create is non-fatal: log `log.Printf("uiadapter: log dir create failed: %v", err)` and continue (the production logger will gracefully degrade to stdout-only when the file cannot be opened — see Risks).
- **Edited file:** `.gitignore` — append `logs/`.

### Technical Considerations

- **Hot-path:** `fanoutHandler.Enabled` must short-circuit using the most permissive child handler (i.e. enabled if ANY child is enabled at the requested level). This keeps `Enabled` cheap and lets debug-level callers skip allocation when neither sink wants the record.
- **Concurrency:** `slog.Handler` implementations must be safe for concurrent use. The stdlib `TextHandler` and `JSONHandler` already are; `fanoutHandler` is a stateless dispatcher so it inherits that safety. No additional mutex required.
- **Closer semantics:** the `io.Closer` returned by `NewProductionLogger` only needs to close the file handle. If file open failed at construction time, return a no-op `io.Closer` (so the caller can still defer close unconditionally) and a logger that fans out only to stdout.
- **Standard attribute names** — adopt these globally and document them in `logging.go` doc comment so subsequent stories use them: `op`, `proc_id`, `model`, `latency_ms`, `bytes_in`, `bytes_out`, `tier`, `reason`, `cache_key_hash`.
- **Wails binding:** none. This story is pure Go.

### Risks & Edge Cases

- **Log dir not writable** (e.g. read-only working directory): `NewProductionLogger` must return a stdout-only logger and a no-op closer rather than failing the boot. The error from `os.OpenFile` should be reported via the returned `error` but the logger and closer must be non-nil so the caller can log the error then proceed.
- **Daily rollover at midnight:** out of scope — the file name is computed once at boot. A long-running session keeps writing to the boot-day file. Acceptable for this story.
- **`./logs` relative path:** macOS Wails apps may run with the working dir set to the bundle, not the project root. Use `cfg.LogDir` (defaults to `./logs`) verbatim — do not call `os.Getwd`. The user can override via env / settings later.
- **Secret leakage:** the production logger MUST NOT add a custom `ReplaceAttr` that strips fields — that is the job of the per-call instrumentation in later stories. This story just provides the conduit.
- **`UIADAPTER_LOG_LEVEL` typo:** silently fall back to `info`; do not crash boot.
- **Test reproducibility:** `time.Now().UTC()` for the date stamp is acceptable; injecting a clock is over-engineering for daily-precision filenames.

### Reference Files

- `/Users/dev/Development/mashed/internal/uiadapter/adapter.go:118` — `NewDefault(cfg Config, logger *slog.Logger) Adapter` already accepts the logger and nil-guards via `slog.New(slog.NewTextHandler(io.Discard, nil))`. Boot wiring in this story replaces the `slog.Default()` call site at `app.go:283`.
- `/Users/dev/Development/mashed/internal/uiadapter/config.go:10` — `DefaultConfig()` and `mergeWithDefaults` (reflection-based) — pattern to follow for adding `LogLevel` / `LogDir`.
- `/Users/dev/Development/mashed/app.go:283` — exact line to edit for boot wiring.
- `/Users/dev/Development/mashed/app.go:317-331` — `(*App).shutdown` — append hook drain here.
- `/Users/dev/Development/mashed/main.go:194` — `OnShutdown: app.shutdown` — verify hook drain runs.
- `/Users/dev/Development/mashed/internal/uiadapter/allowlist.go` — existing `slog.Default()` usage; do NOT modify in this story.

### Skills

- Invoke `/golang-error-handling` for the `NewProductionLogger` error sentinel pattern (e.g. `ErrLogDirUnwritable`) before writing the constructor.
- Invoke `/golang-testing` for the `logging_test.go` table-driven scaffolding (the actual logging tests land in Story 6, but this story should ship a smoke test for `NewProductionLogger` with stdout-only fallback).
- Invoke `/simplify` after implementation as required by Definition of Done.

## Acceptance Criteria

AC-1.1: New `NewProductionLogger` returns stdout+file fan-out logger
- Given the adapter package
- When `NewProductionLogger(slog.LevelDebug, t.TempDir())` is called
- Then it returns a non-nil `*slog.Logger`, a non-nil `io.Closer`, and a nil error
- And calling `logger.Debug("hello", "k", "v")` writes one human-readable line to stdout AND one JSON line to `tempDir/uiadapter-YYYYMMDD.log`

AC-1.2: Closer closes the file handle and is idempotent
- Given a logger built by `NewProductionLogger`
- When the returned closer's `Close()` is invoked twice
- Then the first call closes the file, the second call returns `nil` (no panic, no error)
- And subsequent log writes do not panic — they may silently drop file output

AC-1.3: Fallback to stdout-only when log dir is unwritable
- Given `NewProductionLogger(slog.LevelInfo, "/nonexistent/readonly/path")` is invoked on a path that cannot be created
- When the constructor returns
- Then the returned `*slog.Logger` is non-nil and writes to stdout only
- And the returned `io.Closer` is non-nil and `Close()` returns nil
- And the returned `error` is non-nil so the caller can surface the warning

AC-1.4: `UIADAPTER_LOG_LEVEL` env var controls boot level
- Given `UIADAPTER_LOG_LEVEL=debug` is set in the environment when `app.go` boots
- When the production logger is constructed
- Then the logger's effective level is `slog.LevelDebug`
- And unset / unparseable env values fall back to `slog.LevelInfo` without erroring

AC-1.5: Config exposes `LogLevel` and `LogDir`
- Given a zero-value `Config{}`
- When `mergeWithDefaults(cfg)` runs
- Then `cfg.LogLevel == "info"` and `cfg.LogDir == "./logs"`
- And explicit caller values are preserved verbatim

AC-1.6: Shutdown closes the log file cleanly
- Given the app boots with the production logger
- When `(*App).shutdown` is called by Wails
- Then every closer registered in `App.shutdownHooks` is invoked exactly once
- And shutdown does not block on closer errors (errors are `log.Printf`'d only)

## BDD Test Scenarios

### Scenario 1: Production logger fan-out

```gherkin
Feature: NewProductionLogger fan-out

  Scenario: Debug log lands in both sinks
    Given a fresh tempDir for log output
    When I call NewProductionLogger with level=Debug and logDir=tempDir
    Then the returned logger is non-nil
    And the returned closer is non-nil
    And the returned error is nil
    When I call logger.Debug with msg="boot" and attrs k=v
    Then stdout contains the substring "boot"
    And the file tempDir/uiadapter-<UTC-YYYYMMDD>.log contains exactly one JSON line
    And that JSON line parses successfully and contains "msg":"boot","k":"v"

  Scenario: Info filters out Debug when level=Info
    Given a logger built with level=Info
    When I call logger.Debug("filtered")
    Then neither stdout nor the file contain "filtered"
    When I call logger.Info("kept")
    Then both stdout and the file contain "kept"
```

### Scenario 2: Closer idempotency and stdout fallback

```gherkin
Feature: Closer behaviour and graceful degradation

  Scenario: Double close is safe
    Given a logger and closer from NewProductionLogger
    When I call closer.Close()
    Then it returns nil
    When I call closer.Close() a second time
    Then it returns nil
    And no panic occurs

  Scenario: Unwritable log dir degrades to stdout
    Given a path "/this/cannot/be/created/even/with/mkdir-p" guaranteed unwritable
    When I call NewProductionLogger with that path
    Then the returned logger is non-nil
    And the returned closer is non-nil
    And the returned error is non-nil and identifies the path issue
    When I call logger.Info("smoke")
    Then stdout contains "smoke"
    And no file is created
```

### Scenario 3: Boot wiring honours env var and shutdown hook

```gherkin
Feature: app.go boot wiring

  Scenario: UIADAPTER_LOG_LEVEL=debug is honoured
    Given UIADAPTER_LOG_LEVEL=debug is set in the environment
    When the App boots and constructs the uiadapter
    Then the logger passed to uiadapter.NewDefault has effective level Debug

  Scenario: Unset env defaults to Info
    Given UIADAPTER_LOG_LEVEL is unset
    When the App boots
    Then the logger passed to uiadapter.NewDefault has effective level Info

  Scenario: Shutdown drains hooks
    Given the App has booted and the production logger's closer is registered
    When App.shutdown is called
    Then every registered shutdown hook is invoked exactly once
    And the log file handle is closed
    And shutdown returns even if a hook returns an error
```

## Tasks / Subtasks

- [ ] Task 1: Create `internal/uiadapter/logging.go` (AC: 1.1, 1.2, 1.3, 1.4)
  - [ ] Subtask 1a: Define `fanoutHandler` struct with a `[]slog.Handler` field; implement `Enabled`, `Handle`, `WithAttrs`, `WithGroup` (each method delegates to every child handler).
  - [ ] Subtask 1b: Implement `NewProductionLogger(level slog.Level, logDir string) (*slog.Logger, io.Closer, error)`. On `os.MkdirAll` or `os.OpenFile` failure return a stdout-only logger plus the wrapped error.
  - [ ] Subtask 1c: Implement `parseSlogLevel(s string) slog.Level` accepting `"debug"`, `"info"`, `"warn"`, `"error"` case-insensitively; default `slog.LevelInfo`.
  - [ ] Subtask 1d: Implement an idempotent `closeOnce` wrapper over `*os.File` so AC-1.2 is structurally guaranteed.
  - [ ] Subtask 1e: Add a top-of-file doc comment listing the standard attribute names (`op`, `proc_id`, `model`, `latency_ms`, `bytes_in`, `bytes_out`, `tier`, `reason`, `cache_key_hash`).

- [ ] Task 2: Extend `Config` with `LogLevel` and `LogDir` (AC: 1.5)
  - [ ] Subtask 2a: Edit `internal/uiadapter/config.go` — add fields after `DisableHealthTicker`.
  - [ ] Subtask 2b: Add defaults in `DefaultConfig()`.
  - [ ] Subtask 2c: Verify `mergeWithDefaults` already handles them (string fields backfill via reflection); add a regression test.

- [ ] Task 3: Wire boot in `app.go` and `main.go` (AC: 1.4, 1.6)
  - [ ] Subtask 3a: Edit `main.go` — call `os.MkdirAll("./logs", 0o755)` before `wails.Run`; non-fatal on error.
  - [ ] Subtask 3b: Add `shutdownHooks []func() error` field to `App` struct.
  - [ ] Subtask 3c: At `app.go:283`, replace `slog.Default()` with the result of `uiadapter.NewProductionLogger(parseSlogLevel(os.Getenv("UIADAPTER_LOG_LEVEL")), "./logs")`. Append the closer (wrapped as `func() error`) to `a.shutdownHooks`.
  - [ ] Subtask 3d: In `(*App).shutdown`, iterate `shutdownHooks` and invoke each, `log.Printf` errors.

- [ ] Task 4: Smoke tests for `NewProductionLogger` (AC: 1.1, 1.2, 1.3)
  - [ ] Subtask 4a: New file `internal/uiadapter/logging_test.go` with cases for fan-out, level filter, idempotent close, unwritable dir fallback. Use `t.TempDir()`.
  - [ ] Subtask 4b: Add `TestParseSlogLevel` table test covering all valid inputs + empty + garbage.
  - [ ] Subtask 4c: Add a regression test asserting `mergeWithDefaults(Config{}).LogLevel == "info"` and `LogDir == "./logs"`.

- [ ] Task 5: `.gitignore` housekeeping (AC: 1.1)
  - [ ] Subtask 5a: Append `logs/` to `.gitignore`.
  - [ ] Subtask 5b: Verify with `git status --ignored` that `./logs/uiadapter-*.log` is ignored.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/uiadapter/logging.go` and the modified `config.go` block
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
