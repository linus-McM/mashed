# Story 5: Instrument payload shaping (sanitize, spotlight, contextguard, validator, schema, encode, sampling, mock, allowlist)

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** L
**Depends On:** uiadapter-logging-2-plumb-subcomponents
**Status:** done

## Description

Add structured `slog` debug-level instrumentation to the payload-shaping surface of the UI adapter: `sanitize.go`, `spotlight.go`, `contextguard.go`, `validator.go`, `schema.go`, `encode.go`, `sampling.go`, `mock.go`, `allowlist.go`. After this story, every uiadapter source file in the request path emits at least one debug-level log per request, satisfying plan AC-1's coverage criterion. Sanitize discipline (§14) is enforced — the *very files* that strip payloads must themselves emit only metadata about what they stripped.

This story is independent of Stories 3 and 4 and may run in parallel; all three depend on Story 2's plumbing. They touch disjoint files.

## Developer Notes

### Architecture

Instrument with `logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)` calls behind the `Enabled` guard.

#### `internal/uiadapter/sanitize.go`

| Site | Message | Required attrs |
|---|---|---|
| `SanitizeCapture` entry | `sanitize.start` | `op="sanitize"`, `bytes_in=len(raw)` |
| `SanitizeCapture` return | `sanitize.done` | `op`, `bytes_in`, `bytes_out`, `delta_bytes` |
| `stripOutsideFences` removed bytes | `sanitize.fence_strip` | `op`, `removed_bytes` |
| `sanitizeChrome` pattern matched | `sanitize.chrome_match` | `op`, `patterns_matched_count` |

`SanitizeCapture` already returns `deltaBytes` — emit that value verbatim. The new helpers may need a small refactor to count matches; do this minimally.

#### `internal/uiadapter/spotlight.go`

| Site | Message | Required attrs |
|---|---|---|
| `Spotlight` entry | `spotlight.start` | `op="spotlight"`, `enabled`, `bytes_in` |
| `Spotlight` produced markers | `spotlight.added` | `op`, `markers_added`, `bytes_out` |
| `Unspotlight` | `spotlight.removed` | `op`, `markers_removed`, `bytes_out` |
| Disabled short-circuit | `spotlight.disabled` | `op`, `bytes_in` |

#### `internal/uiadapter/contextguard.go`

| Site | Message | Required attrs |
|---|---|---|
| `ApplyOllama` entry | `contextguard.ollama.start` | `op="contextguard.ollama"`, `bytes_in` |
| `ApplyOllama` truncation | `contextguard.ollama.truncated` | `op`, `truncated=<bool>`, `original_tokens`, `fitted_tokens`, `reserve_tokens` |
| `ApplyClaude` entry | `contextguard.claude.start` | `op="contextguard.claude"`, `bytes_in`, `max_model_context` |
| `ApplyClaude` truncation | `contextguard.claude.truncated` | `op`, `truncated=<bool>`, `original_tokens`, `fitted_tokens` |
| `OllamaOptions` accessor | `contextguard.ollama.options` | `op` (one-line debug; called per-request) |

#### `internal/uiadapter/validator.go`

| Site | Message | Required attrs |
|---|---|---|
| `Validate` entry | `validator.start` | `op="validator"`, `node_count=len(ast.Nodes)`, `bytes_in=len(raw)` |
| `Validate` return | `validator.done` | `op`, `fail_count=len(reasons)`, `reasons=<sorted distinct enum list>` |
| Each rule firing (1..8) | `validator.rule.fail` | `op`, `rule_id=N`, `reason=<enum>` |

The reasons list in `validator.done` is bounded — log it as a `slog.Any("reasons", reasons)`. Each `reasons` element is an enum string, never user content.

#### `internal/uiadapter/schema.go`

| Site | Message | Required attrs |
|---|---|---|
| `WidgetNode.UnmarshalJSON` start | `schema.unmarshal.widget` | `op="schema.unmarshal"`, `bytes_in=len(b)` |
| Unmarshal failure | `schema.unmarshal.error` | `op`, `reason="parse"` (do NOT log err.Error()) |

#### `internal/uiadapter/encode.go`

| Site | Message | Required attrs |
|---|---|---|
| `OllamaFormatPayload` | `encode.ollama_format` | `op="encode.ollama_format"`, `kind=<StageKind>`, `loose=<bool>`, `bytes_out` |
| `ClaudeToolInputSchema` | `encode.claude_tool_schema` | `op="encode.claude_tool"`, `kind`, `bytes_out` |
| `ClaudeToolName` | `encode.claude_tool_name` | `op`, `kind`, `name` |
| `schemaBytes` selection | `encode.schema_select` | `op`, `kind`, `variant=<embedded-file-name>` |

#### `internal/uiadapter/sampling.go`

| Site | Message | Required attrs |
|---|---|---|
| `OllamaSamplingOptions` | `sampling.ollama` | `op="sampling.ollama"`, `seed`, `temperature` |
| `ClaudeSamplingOptions` | `sampling.claude` | `op="sampling.claude"`, `temperature` (Claude has no seed knob) |

#### `internal/uiadapter/mock.go`

| Site | Message | Required attrs |
|---|---|---|
| `NewMock` construction | `mock.init` | `op="mock.init"`, `enabled=true`, `fixture_name=<derived>` (derive from `fixed.GeneratedBy` or `"unnamed"`) |
| `MockAdapter.Translate` entry | `mock.translate` | `op="mock.translate"`, `proc_id`, `bytes_in=len(raw)` |

#### `internal/uiadapter/allowlist.go`

| Site | Message | Required attrs |
|---|---|---|
| `CheckModelAllowlist` vetted-OK path | `allowlist.ok` (Debug) | `op="allowlist.check"`, `model=cfg.Model` |
| Existing warn path on breach | `allowlist.breach` (Warn — already exists) | `op`, `model`, `allowlist_count` |

The existing `slog.Default()` call in `allowlist.go` is replaced by `cfg`'s injected logger (passed as the existing `logger *slog.Logger` arg — already in signature). No structural change beyond preferring the explicit logger over `slog.Default()` when both are non-nil.

### Technical Considerations

- **§14 sanitize discipline**: critical in this story because these files SEE the raw payload. Triple-check every emission for accidental inclusion of `raw`, `s`, `prompt`, etc. Reviewer must grep the diff for any attr value that comes from a `string` parameter holding model/user content.
- **`schema.go:WidgetNode.UnmarshalJSON`**: this runs inside `json.Unmarshal` and may be called many times in deeply-nested ASTs. Wrap emissions in `Enabled` guard religiously — this is the hottest path in the story.
- **`validator.go` per-rule emissions**: 8 rules * possible fires per request. Use `Enabled` guard. The aggregate `validator.done` carries the `reasons` slice so you can opt to make per-rule emission Trace-level (slog has no trace; treat per-rule as `slog.Level(-8)` if discrimination matters — otherwise Debug is fine).
- **`mock.go` fixture name**: `MockAdapter.fixed.GeneratedBy` is the natural identifier. If empty, fall back to `"unnamed"`.
- **`allowlist.go` migration off `slog.Default()`**: today `CheckModelAllowlist(cfg, logger)` falls back to `slog.Default()` when `logger == nil`. Replace the fallback with a call to `nilSafeLogger` (added in Story 1/2). This unifies behaviour with the rest of the package.
- **`schema.go:WidgetNode` has no logger field today** — it's a value type used during unmarshalling. Pass the logger explicitly via a closure or a context-scoped helper. **Practical approach:** skip per-call logging in `UnmarshalJSON` and instead instrument the *caller* (`adapter.go:Translate` already logs through `logTelemetry`). Mark this as a known exception in the story comment.

### Risks & Edge Cases

- **`UnmarshalJSON` cannot easily access the package logger.** The plan calls for it but the `json.Unmarshal` indirection makes it impractical. **Resolution:** instrument `schema.go` only at the function entry where a package-level helper can be invoked, OR document a deliberate skip and rely on the `adapter.go` caller's `validation:malformed` reason. The cleanest path: add a small package-level `var schemaLogger atomic.Pointer[slog.Logger]` set by `NewDefault` and read by `WidgetNode.UnmarshalJSON`. This is the *only* exception to "explicit logger param" in the package and must be documented in the story.
- **Validator log volume**: a malformed AST may trip 5+ rules. With per-rule emissions plus an aggregate, that's 6 records per request. Acceptable but verify with the volume test in AC-5.8.
- **`CheckModelAllowlist` is called once per backend init**, not per request — so the new `allowlist.ok` debug record is rare. No volume concern.
- **`sampling.go` log values are config knobs** — they look constant per-process, but a future story may make them dynamic. Keep the per-call emission; it's cheap and observable.
- **`encode.go` `schemaBytes` reads embedded files** — we may not have a stable "variant" identifier. Use the `kind.String()` value if no other identifier exists; mark this minor.

### Reference Files

- `/Users/dev/Development/mashed/internal/uiadapter/sanitize.go` — entry point for every Translate request.
- `/Users/dev/Development/mashed/internal/uiadapter/contextguard.go:40-99` — the two `Apply*` methods.
- `/Users/dev/Development/mashed/internal/uiadapter/validator.go:40-67` — `Validate` orchestrator and rule appliers.
- `/Users/dev/Development/mashed/internal/uiadapter/schema.go:83` — `WidgetNode.UnmarshalJSON`.
- `/Users/dev/Development/mashed/internal/uiadapter/allowlist.go:31` — already takes a logger; only file in plan that does.
- `/Users/dev/Development/mashed/internal/uiadapter/logging.go` (Story 1) — `nilSafeLogger`, attr name conventions.
- Story 3's `testLogBuffer` helper is reused.

### Skills

- Invoke `/simplify` after each file's instrumentation; this story has nine files so simplify is especially valuable for catching duplicated emission patterns that should be extracted.
- Invoke `/golang-testing` for table-driven tests over the validator rule matrix (8 rules x fire/no-fire).

## Acceptance Criteria

AC-5.1: `sanitize.go` emits delta and pattern metadata (no payload bytes)
- Given a Debug-level logger and `raw` containing ANSI escapes and code fences
- When `SanitizeCapture(raw, logger)` is called
- Then a `sanitize.start` and a `sanitize.done` record are emitted
- And the `done` record's `delta_bytes` equals `bytes_in - bytes_out`
- And no record contains any substring of `raw`

AC-5.2: `spotlight.go` emits added and removed counts
- Given `Spotlight(raw, true, logger)` is called on text with three terminal-prompt markers
- Then a `spotlight.start` and a `spotlight.added` record appear, with `markers_added=3`
- Given `Spotlight(raw, false, logger)`
- Then exactly one `spotlight.disabled` record appears
- And `Unspotlight(marked, logger)` (if exposed) emits `spotlight.removed`

AC-5.3: `contextguard.go` emits truncation outcomes
- Given an `ApplyOllama(raw)` call where `raw` exceeds NumCtx
- Then a `contextguard.ollama.truncated` record is emitted with `truncated=true`, `original_tokens`, `fitted_tokens`, `reserve_tokens`
- And under-budget input emits `truncated=false`
- And `ApplyClaude` produces analogous records via `contextguard.claude.*`

AC-5.4: `validator.go` emits aggregate and (optionally) per-rule failures
- Given a `Validate(ast, raw, logger)` call where rules 1, 4, and 8 fail
- Then a `validator.done` record is emitted with `fail_count=3` and `reasons` containing the three enum strings
- And per-rule `validator.rule.fail` records appear for rules 1, 4, 8 each with `rule_id` and `reason`
- And no record contains any substring of `raw` or any field from `ast` that holds user content

AC-5.5: `schema.go`, `encode.go`, `sampling.go`, `mock.go` emit construction-time metadata
- Given a Debug-level logger
- When `OllamaFormatPayload(StageKindWidget, false, logger)`, `ClaudeToolInputSchema(StageKindWidget, logger)`, `OllamaSamplingOptions(cfg, logger)`, `NewMock(fixed, logger)`, and `MockAdapter.Translate(ctx, raw, "proc-1")` are exercised
- Then each emits its respective record with the attrs listed in the table above
- And `mock.translate` carries `proc_id="proc-1"` and `bytes_in=len(raw)`

AC-5.6: `allowlist.go` `slog.Default()` removed; vetted-OK Debug record added
- Given the post-Story-2 signature `CheckModelAllowlist(cfg, logger)` is invoked with a non-nil logger and a vetted model
- Then a `allowlist.ok` Debug record is emitted with `op="allowlist.check"`, `model=cfg.Model`
- And no `slog.Default()` call remains anywhere in `allowlist.go`
- And nil-logger callers continue to work via `nilSafeLogger`

AC-5.7: `schema.go` UnmarshalJSON wires a package-level logger
- Given `NewDefault(cfg, logger)` is invoked with a Debug-level logger
- And the package-level `schemaLogger` is set during `NewDefault`
- When `json.Unmarshal` decodes a `WidgetNode`
- Then a `schema.unmarshal.widget` record is emitted with `bytes_in`
- And on parse failure a `schema.unmarshal.error` record is emitted with `reason="parse"` (no err text)
- AND when no logger has been registered (e.g. in a test using raw `json.Unmarshal`) emission silently no-ops with no panic

AC-5.8: Coverage and sanitize discipline across the suite
- Given a 100-iteration smoke test exercising every file instrumented in Stories 3, 4, and 5
- When the captured JSON log buffer is scanned
- Then for every request iteration, at least one record is emitted from every uiadapter source file in the request path
- And no record contains any byte of the raw payload, any err.Error() text, any prompt content, or any model output
- And every `reason` and `rule_id` value belongs to a closed enum

AC-5.9: Hot-path zero-allocation when Debug is off
- Given the package logger at level Info
- When the full Translate path runs 10,000 times
- Then `testing.AllocsPerRun` reports no extra allocations attributable to this story's instrumentation

## BDD Test Scenarios

### Scenario 1: Sanitize delta reporting

```gherkin
Feature: sanitize.go observability

  Scenario: ANSI strip reports delta_bytes only
    Given raw = "[31mhello[0m world" (16 bytes)
    And a Debug-level logger to bytes.Buffer
    When I call SanitizeCapture(raw, logger)
    Then a sanitize.start record is emitted with bytes_in=16
    And a sanitize.done record is emitted with bytes_out=11 and delta_bytes=5
    And no record contains the byte sequence "[" or "hello"
```

### Scenario 2: Validator aggregate reasons

```gherkin
Feature: validator.go aggregate emission

  Scenario: Multiple rule failures aggregate cleanly
    Given an AST that violates rules 1, 4, and 8
    And a Debug-level logger
    When Validate(ast, raw, logger) runs
    Then a validator.start record is emitted with node_count and bytes_in
    And per-rule records validator.rule.fail are emitted with rule_id ∈ {1,4,8}
    And a validator.done record is emitted with fail_count=3 and reasons containing the three enum strings
    And no record contains any field from the AST nodes (text bodies, prompts, etc.)
```

### Scenario 3: Context guard truncation

```gherkin
Feature: contextguard.go truncation telemetry

  Scenario: Ollama over-budget input is truncated
    Given a ContextGuard with NumCtx=1024
    And raw long enough to exceed 1024 tokens
    When ApplyOllama(raw) runs
    Then a contextguard.ollama.start record is emitted with bytes_in
    And a contextguard.ollama.truncated record is emitted with truncated=true, original_tokens > fitted_tokens, reserve_tokens >= 0
    And no record contains any substring of raw

  Scenario: Under-budget input passes through
    Given the same guard but a short raw
    When ApplyOllama(raw) runs
    Then the truncated record has truncated=false and original_tokens == fitted_tokens
```

### Scenario 4: Mock and allowlist

```gherkin
Feature: mock.go and allowlist.go observability

  Scenario: Mock translate emits proc_id and bytes_in
    Given NewMock(fixed, logger) where fixed.GeneratedBy="fixture:happy"
    And a Debug-level logger
    When the constructor runs
    Then a mock.init record is emitted with fixture_name="fixture:happy", enabled=true
    When MockAdapter.Translate(ctx, "raw input", "proc-99") runs
    Then a mock.translate record is emitted with proc_id="proc-99", bytes_in=9

  Scenario: Allowlist-OK path emits Debug record
    Given a vetted model in cfg.Model and a Debug-level logger
    When CheckModelAllowlist(cfg, logger) runs
    Then a single allowlist.ok record is emitted with model=cfg.Model
    And no slog.Default() call is made anywhere in allowlist.go (compile-time grep check)
```

### Scenario 5: Schema unmarshal logger

```gherkin
Feature: schema.go package-level logger plumbing

  Scenario: NewDefault registers schemaLogger and Unmarshal emits
    Given NewDefault(cfg, debugLogger) has run
    When json.Unmarshal decodes a payload containing a WidgetNode
    Then a schema.unmarshal.widget record is emitted with bytes_in equal to the WidgetNode's JSON byte length

  Scenario: No-logger fallback in raw test
    Given NewDefault has not been invoked (test calling json.Unmarshal directly)
    When a WidgetNode is unmarshalled
    Then no panic occurs and no record is emitted
```

### Scenario 6: Coverage and sanitize discipline

```gherkin
Feature: §14 discipline and per-file coverage

  Scenario: Every file in the request path emits at least one record
    Given the production-shaped logger at Debug level
    When a single Translate request flows through the full pipeline
    Then the captured JSON buffer contains at least one record sourced from each of:
      sanitize, spotlight, contextguard, validator, schema, encode, sampling
      AND from Stories 3 and 4: client, cache, breaker, semaphore, fastpath, repair, stages, fallback, fallback_tiers
    And no record contains any byte of the raw payload, body, or err.Error()
```

## Tasks / Subtasks

- [ ] Task 1: Instrument `sanitize.go` (AC: 5.1, 5.8, 5.9)
  - [ ] Subtask 1a: Emit `sanitize.start` / `sanitize.done` at function boundaries.
  - [ ] Subtask 1b: Emit `sanitize.fence_strip` and `sanitize.chrome_match` from helpers when a meaningful change occurs.

- [ ] Task 2: Instrument `spotlight.go` (AC: 5.2, 5.8, 5.9)
  - [ ] Subtask 2a: Count markers added in `Spotlight`; emit `spotlight.added` with the count.
  - [ ] Subtask 2b: Emit `spotlight.disabled` short-circuit; `spotlight.removed` in `Unspotlight`.

- [ ] Task 3: Instrument `contextguard.go` (AC: 5.3, 5.8, 5.9)
  - [ ] Subtask 3a: Add `start` and `truncated` records in `ApplyOllama` and `ApplyClaude`.
  - [ ] Subtask 3b: Add a tiny `OllamaOptions` Debug emission for visibility.

- [ ] Task 4: Instrument `validator.go` (AC: 5.4, 5.8, 5.9)
  - [ ] Subtask 4a: Emit `validator.start` and `validator.done` at function boundaries; emit `validator.rule.fail` from each `applyRule*` helper when reasons grow.
  - [ ] Subtask 4b: Verify the `reasons` slice serialises cleanly via `slog.Any`.

- [ ] Task 5: Instrument `schema.go` with package-level logger (AC: 5.7)
  - [ ] Subtask 5a: Add `var schemaLogger atomic.Pointer[slog.Logger]` to `schema.go`.
  - [ ] Subtask 5b: In `NewDefault` (`adapter.go`), call `schemaLogger.Store(scoped)` once after the scoped logger is built.
  - [ ] Subtask 5c: In `WidgetNode.UnmarshalJSON`, load the pointer; if non-nil and Debug-enabled, emit `schema.unmarshal.widget` and `schema.unmarshal.error` (no err text).
  - [ ] Subtask 5d: Document the deliberate exception to "explicit logger param" with a `// Exception:` comment block.

- [ ] Task 6: Instrument `encode.go` and `sampling.go` (AC: 5.5, 5.8, 5.9)
  - [ ] Subtask 6a: Add Debug emissions to `OllamaFormatPayload`, `ClaudeToolInputSchema`, `ClaudeToolName`, `schemaBytes`.
  - [ ] Subtask 6b: Add Debug emissions to `OllamaSamplingOptions` and `ClaudeSamplingOptions`.

- [ ] Task 7: Instrument `mock.go` (AC: 5.5)
  - [ ] Subtask 7a: Add `mock.init` to `NewMock`; derive `fixture_name` from `fixed.GeneratedBy` (or `"unnamed"`).
  - [ ] Subtask 7b: Add `mock.translate` to `MockAdapter.Translate`.

- [ ] Task 8: Tighten `allowlist.go` (AC: 5.6)
  - [ ] Subtask 8a: Replace any `slog.Default()` fallback with `nilSafeLogger`.
  - [ ] Subtask 8b: Add `allowlist.ok` Debug emission on the vetted-OK path.
  - [ ] Subtask 8c: Add a `golangci-lint`-style grep test asserting `slog.Default()` is gone from this file.

- [ ] Task 9: Tests (AC: 5.1–5.9)
  - [ ] Subtask 9a: Reuse `testLogBuffer`; augment each affected `*_test.go` with the BDD scenarios.
  - [ ] Subtask 9b: Add a coverage test that exercises a single Translate end-to-end and asserts the per-file emission set (AC-5.8). This anticipates Story 6's E2E test but lives at unit-test scope here.
  - [ ] Subtask 9c: Add the alloc-budget test parallel to Stories 3 and 4.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on the nine instrumented files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
