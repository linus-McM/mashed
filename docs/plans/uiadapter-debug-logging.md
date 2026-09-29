# UI Adapter Debug Logging — Implementation Plan

**Status:** Draft
**Author:** Claude (Opus 4.7)
**Date:** 2026-04-26
**Owner:** TBD
**Estimate:** ~8h

## Goal

Add structured `slog` debug logging across every `internal/uiadapter/*.go` source file to enable end-to-end tracing of UI AST translation requests. Logs fan out to **stdout** (human-readable) and `./logs/uiadapter-YYYYMMDD.log` (JSON, machine-grep). Toggleable via `UIADAPTER_LOG_LEVEL` env var; default `info` so production verbosity is unchanged.

## Non-Goals

- No log rotation library (daily file rollover only — size-based via `lumberjack` is future work).
- No OpenTelemetry / cross-goroutine trace IDs. Propagate existing `proc_id` only.
- No structured logging in `*_test.go` files beyond what is already required by tests.

## Current State

| File | Logging |
|------|---------|
| `adapter.go` | Uses `*slog.Logger` (`logTelemetry`, `logTelemetryWithSanitize`) |
| `allowlist.go` | Uses `slog.Default()` for warn-level allowlist breaches |
| All other 23 source files | **Zero logging** |

`NewDefault(cfg, logger)` accepts a logger; `app.go:283` passes `slog.Default()`. Subcomponents (`Client`, `Cache`, `Breaker`, `FastPath`, `Repair`, `Stages`, `Fallback`, `Sanitize`, `Spotlight`, `ContextGuard`, `Validator`, `Schema`, `Encode`, `Sampling`, `Semaphore`, `PrefixCache`, `Mock`) receive **no logger** today.

## Logging Discipline

Hard rules to prevent payload leakage (§14 sanitize discipline):

- **Never** log `raw`, `body`, `prompt`, `response.text`, model output strings, or any user-supplied content.
- Log lengths (`len()`), SHA-256 hash prefixes (8 hex chars), counts, durations, IDs, enums, booleans only.
- Use `slog.LogAttrs(ctx, level, msg, attrs...)` in hot paths (zero-allocation).
- Standard attribute names across all files:
  - `op` (string) — operation name (`translate`, `cache.get`, `client.request`)
  - `proc_id` (string) — propagated from `Translate(ctx, raw, procID)`
  - `model` (string)
  - `latency_ms` (int)
  - `bytes_in` / `bytes_out` (int)
  - `tier` (string) — `fastpath` / `ollama` / `claude-haiku` / `claude-sonnet` / `fallback`
  - `reason` (string) — short tag, no free-form text
  - `cache_key_hash` (string, 8 hex)

## Phase 1 — Logger Infrastructure

**New file: `internal/uiadapter/logging.go`** (~60 LOC)

```go
package uiadapter

import (
    "context"
    "io"
    "log/slog"
    "os"
    "path/filepath"
    "time"
)

func NewProductionLogger(level slog.Level, logDir string) (*slog.Logger, io.Closer, error) {
    if err := os.MkdirAll(logDir, 0o755); err != nil {
        return nil, nil, err
    }
    fileName := filepath.Join(logDir, "uiadapter-"+time.Now().Format("20060102")+".log")
    f, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
    if err != nil {
        return nil, nil, err
    }
    addSource := level <= slog.LevelDebug
    stdoutH := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level, AddSource: addSource})
    fileH := slog.NewJSONHandler(f, &slog.HandlerOptions{Level: level, AddSource: addSource})
    return slog.New(fanout{stdoutH, fileH}), f, nil
}

// fanout dispatches each record to every wrapped handler.
type fanout []slog.Handler

func (f fanout) Enabled(ctx context.Context, l slog.Level) bool {
    for _, h := range f {
        if h.Enabled(ctx, l) {
            return true
        }
    }
    return false
}
func (f fanout) Handle(ctx context.Context, r slog.Record) error {
    for _, h := range f {
        if h.Enabled(ctx, r.Level) {
            _ = h.Handle(ctx, r.Clone())
        }
    }
    return nil
}
func (f fanout) WithAttrs(a []slog.Attr) slog.Handler {
    out := make(fanout, len(f))
    for i, h := range f {
        out[i] = h.WithAttrs(a)
    }
    return out
}
func (f fanout) WithGroup(g string) slog.Handler {
    out := make(fanout, len(f))
    for i, h := range f {
        out[i] = h.WithGroup(g)
    }
    return out
}
```

**Update `internal/uiadapter/config.go`** — add fields:

```go
LogLevel string // "debug" | "info" | "warn" | "error"; default "info"
LogDir   string // default "./logs"
```

Backfill in `mergeWithDefaults`.

## Phase 2 — Wire Boot

**`app.go:283`** — replace `slog.Default()` with the production logger:

```go
level := parseLogLevel(os.Getenv("UIADAPTER_LOG_LEVEL"))
logger, closer, err := uiadapter.NewProductionLogger(level, cfg.LogDir)
if err != nil {
    log.Printf("uiadapter logger init failed, falling back to stderr only: %v", err)
    logger = slog.Default()
}
if closer != nil {
    a.shutdownHooks = append(a.shutdownHooks, closer.Close)
}
bmadOpts = append(bmadOpts, bmad.WithAdapter(uiadapter.NewDefault(uiadapter.Config{...}, logger)))
```

**`main.go`** — best-effort `os.MkdirAll("./logs", 0o755)` early at boot; failure is non-fatal (warn + stdout-only).

**`.gitignore`** — add `logs/`.

## Phase 3 — Plumb Logger to Subcomponents

Right now `Client`, `Cache`, `Breaker`, etc. are constructed inside `NewDefault` without a logger. Add `logger *slog.Logger` field to each struct and update constructors:

| Constructor | Old signature | New signature |
|-------------|---------------|---------------|
| `NewClient` | `(ClientConfig)` | `(ClientConfig, *slog.Logger)` |
| `newCache` | `(cap int)` | `(cap int, logger *slog.Logger)` |
| `newBreaker` | `(cfg)` | `(cfg, logger)` |
| `newSemaphore` | `(n int)` | `(n int, logger)` |
| `newPrefixCache` | `(cap)` | `(cap, logger)` |

Free functions (`runFastPath`, `repair`, `runTwoStage`, `applyFallbackTiers`, `Sanitize`, `Spotlight`, `ContextGuard.Apply*`, `validate`, `encode`) take `logger *slog.Logger` as an **explicit param** — context-value logging is an anti-pattern.

`NewDefault` injects the same logger (with a `slog.WithGroup("uiadapter")` namespace) into every subcomponent. All constructors guard `nil` → `slog.New(slog.NewTextHandler(io.Discard, nil))`, matching the existing pattern at `adapter.go:119`.

## Phase 4 — Instrument Each File

Add `Debug` calls at function entry/exit and key branches. Per-file targets:

| File | Log points |
|------|-----------|
| `client.go` | request start (`model`, `endpoint`, `num_ctx`, `bytes_in`); response (`status`, `latency_ms`, `bytes_out`); transport error (reason class only) |
| `cache.go` | hit/miss (`cache_key_hash`, `size`); eviction (`reason`, `evicted_count`) |
| `prefix_cache.go` | prefix hit (`depth`, `bytes_saved`); miss |
| `breaker.go` | state transition (`from`, `to`, `fail_count`); reject (`reason`) |
| `fastpath.go` | trigger (`pattern_id`, `latency_saved_ms`); skip (`reason`) |
| `repair.go` | repair attempt N (`attempt`, `prev_reason`); success/failure |
| `stages.go` | stage 1 done (`tier`, `latency_ms`); stage 2 start (`reason`); final (`tier`) |
| `fallback.go`, `fallback_tiers.go` | tier entered (`from`, `to`, `reason`) |
| `sanitize.go` | `delta_bytes`, `patterns_matched_count` (no content) |
| `spotlight.go` | spotlight added (`capture_size_bytes`) |
| `contextguard.go` | `truncated` bool, `original_tokens`, `fitted_tokens`, `reserve_tokens` |
| `validator.go` | validation reasons (already tagged via existing pattern), `fail_count` |
| `schema.go`, `encode.go`, `sampling.go` | schema variant chosen (`variant`), encoding path, `seed` |
| `semaphore.go` | acquire wait (`wait_ms` only if > threshold), `in_flight` |
| `mock.go` | enabled, `fixture_name` |
| `allowlist.go` | already warns — add debug for "vetted ok" path |
| `repair.go`, `fallback.go` | every retry: `attempt`, `prev_tier`, `next_tier`, `reason` |

**Hot-path guard:** in `client.go` and `cache.go`, gate expensive attribute computation:

```go
if a.logger.Enabled(ctx, slog.LevelDebug) {
    a.logger.LogAttrs(ctx, slog.LevelDebug, "client.request", attrs...)
}
```

## Phase 5 — Tests

**New: `internal/uiadapter/logging_test.go`**

- Asserts log file is created in temp dir, JSON parses, level filter works (debug vs info).
- Asserts fanout dispatches to all handlers.
- Asserts handler chain still respects `WithAttrs` / `WithGroup`.

**Existing tests** — pass `nil` logger → `io.Discard` fallback in every new constructor (mirror `adapter.go:119`). No behavioural change.

**E2E test** — run `Translate` with a debug logger pointed at `bytes.Buffer`, assert ≥1 log entry per phase: `client.request`, `cache.miss`, `validator.ok` (or equivalent), `translate.return`.

**Race-safety** — tests under `-race` must not write to a shared file; always use `t.TempDir()`.

## Phase 6 — Ergonomics & Docs

- **`justfile` recipe** — `just trace` → `UIADAPTER_LOG_LEVEL=debug wails dev`
- **`.gitignore`** — `logs/`
- **`README` snippet** — how to tail the JSON log:
  ```sh
  tail -f logs/uiadapter-*.log | jq 'select(.op=="client.request")'
  ```
- **`.wolf/cerebrum.md`** — Decision Log entry: dual-sink slog (`fanout` over stdout text + file JSON), per-component logger plumbing, no context-value logging.
- **`.wolf/anatomy.md`** — register `internal/uiadapter/logging.go` and `logging_test.go`.

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Debug logging slows hot path | `logger.Enabled(ctx, slog.LevelDebug)` guard before attr building |
| Payload leakage via misuse | Code review checklist + lint rule (future): forbid `slog.String("body", ...)` etc. |
| File handle leak | `closer` returned from `NewProductionLogger`, deferred in app shutdown |
| Disk fill in long sessions | Daily rollover; document cleanup recipe in README; future lumberjack |
| Test flake on shared log file | Always use `t.TempDir()`; never share across tests |

## Order of Operations

1. **Phase 1 + 2** — infra + boot. Green build, no behavioural change.
2. **Phase 3** — plumbing refactor. All tests still green (nil-logger fallback covers it).
3. **Phase 4** — instrument file-by-file in small PRs (one file or cluster per PR for blast-radius control).
4. **Phase 5 + 6** — tests, docs, ergonomics.

## Acceptance Criteria

- [ ] `UIADAPTER_LOG_LEVEL=debug wails dev` produces interleaved stdout text logs and `./logs/uiadapter-YYYYMMDD.log` JSON.
- [ ] Every uiadapter source file emits at least one debug-level log per request (for files in the request path).
- [ ] All existing tests pass with `nil` logger.
- [ ] No raw payload bytes appear in any log entry across a 10-message smoke test (grep `logs/*.log` for known prompt fragments).
- [ ] Default level remains `info`; production behaviour unchanged.
- [ ] Shutdown closes the log file cleanly (no `file already closed` errors on app quit).
