# Story 4: Instrument the request pipeline (fastpath, repair, stages, fallback, fallback_tiers)

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** uiadapter-logging-2-plumb-subcomponents
**Status:** done

## Description

Add structured `slog` debug-level instrumentation along the UI adapter request pipeline: `fastpath.go`, `repair.go`, `stages.go`, `fallback.go`, `fallback_tiers.go`. After this story, a Debug-level run shows whether the fast-path matched, every repair attempt with success/failure, every two-stage transition (classify -> generate), every fallback tier entry with its reason. Sanitize discipline (§14) is enforced — no raw payload bytes ever surface in any record.

This story is **independent** of Story 3 — it can run in parallel — but both depend on Story 2's plumbing. They touch disjoint files.

## Developer Notes

### Architecture

Instrument with `logger.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)` calls behind the `Enabled` guard.

#### `internal/uiadapter/fastpath.go`

| Site | Message | Required attrs |
|---|---|---|
| `Classify` entry | `fastpath.classify.start` | `op="fastpath.classify"`, `bytes_in=len(raw)`, `enabled=f.enabled` |
| Rule hit | `fastpath.classify.hit` | `op`, `rule=<rule.name>`, `latency_ms` |
| No rule matched (skip) | `fastpath.classify.skip` | `op`, `latency_ms` |
| Disabled short-circuit | `fastpath.classify.disabled` | `op` |

`FastPathClassifier.logger` from Story 2.

#### `internal/uiadapter/repair.go`

| Site | Message | Required attrs |
|---|---|---|
| `Repairer.Run` entry | `repair.start` | `op="repair.run"`, `max_retries`, `bytes_in` |
| Each attempt N | `repair.attempt` | `op`, `attempt=N`, `reason=<previous validator reason>` |
| Attempt success | `repair.success` | `op`, `attempt=N`, `latency_ms` |
| Attempt failure (will retry) | `repair.failure` | `op`, `attempt=N`, `reason=<new validator reason>` |
| Exhausted retries (final fail) | `repair.exhausted` | `op`, `attempts_total`, `final_reason` |
| `BuildRepairPrompt` | `repair.prompt.build` | `op="repair.prompt"`, `bytes_out=len(prompt)` |

#### `internal/uiadapter/stages.go`

| Site | Message | Required attrs |
|---|---|---|
| `RunTwoStage` entry | `stages.start` | `op="stages.two"`, `bytes_in` |
| Stage 1 (classify) done | `stages.classify.done` | `op`, `kind=<StageKind>`, `latency_ms` |
| Stage 2 (generate) start | `stages.generate.start` | `op`, `kind` |
| Stage 2 done (final) | `stages.final` | `op`, `kind`, `latency_ms_total`, `bytes_out` |
| Stage 1 parse failure | `stages.classify.parse_error` | `op`, `reason="parse"` |
| `AssembleStage1` / `AssembleStage2` | `stages.assemble` | `op`, `stage=1|2`, `bytes_out=len(prompt)` |
| `ParseStageKind` rejected input | `stages.kind.invalid` | `op`, `input_len` (do NOT log the rejected string verbatim) |

#### `internal/uiadapter/fallback.go`

| Site | Message | Required attrs |
|---|---|---|
| `FallbackAST` entry | `fallback.ast.build` | `op="fallback.ast"`, `reason=<input reason>`, `bytes_in=len(raw)` |
| `firstLine` truncation | `fallback.ast.truncate` | `op`, `truncated=<bool>`, `bytes_kept` |

#### `internal/uiadapter/fallback_tiers.go`

| Site | Message | Required attrs |
|---|---|---|
| `RunWithFallback` entry | `fallback.tier.start` | `op="fallback.tier"`, `tiers_count` |
| Tier entered | `fallback.tier.enter` | `op`, `tier=<FallbackTier>`, `attempt_index` |
| Tier success | `fallback.tier.success` | `op`, `tier`, `latency_ms` |
| Tier failure (advancing) | `fallback.tier.failure` | `op`, `tier`, `reason` |
| All tiers exhausted | `fallback.tier.exhausted` | `op`, `tiers_tried` |

### Technical Considerations

- **§14 sanitize discipline strictly enforced:** never log `raw`, prompt content, model output, full validator-reason strings that contain user data, or `err.Error()`. Validator reasons in this codebase are short enums (`required_dropped`, `oversize`, `marshal_error`, etc.) — those ARE safe to log. If unsure, log only the *first* reason and only by enum name.
- **`reason` enum hygiene:** in `repair.go` the input reason comes from the validator. The validator's reasons are bounded enums — log them verbatim. Anywhere else, prefer a closed set.
- **Two-stage timing:** in `stages.go`, capture `start` once at entry; measure `stage1Done := time.Now()` after stage 1; final emission uses `latency_ms_total = stage1+stage2` and a separate `stages.classify.done` carries `latency_ms = stage1Done - start`.
- **Hot-path guard:** every site wrapped in `if logger.Enabled(ctx, slog.LevelDebug)`.
- **Free function signatures:** Story 2 added `logger *slog.Logger` as the last param. Use it.
- **Avoid duplicate emissions:** `repair.attempt` is emitted once per attempt; `repair.failure` is the *post*-attempt outcome. Do not double-emit.
- **`FallbackAST` is called from many sites** — every fallback path goes through it. Keep the logging minimal (one record at entry, optional truncate record). Otherwise the log volume blows up.

### Risks & Edge Cases

- **Excessive emission in tier loops:** `RunWithFallback` may iterate ~3 tiers. That's 1 entry + N enter + N success-or-failure + 1 exhausted at most ≈ 7 records per request. Acceptable for Debug.
- **`stages.go:RunTwoStage` is a long function** — 50+ LOC. Place log calls only at the four boundaries listed in the table; do not pepper internal helpers.
- **Repair concurrency:** `Repairer.Run` is single-threaded per request. No locking concerns.
- **`fastpath.HitsPerRule` map** — not a request-path call, do NOT instrument it. Leave metric reporting to existing telemetry.
- **`ParseStageKind` rejected input:** the rejected `s` could be model output. Log only `input_len` and `op`, never the string. AC-4.7 enforces this.
- **Test isolation:** existing tests in `repair_test.go`, `stages_test.go`, `fallback_tiers_test.go` will need to be updated to pass a Debug-level test logger when asserting log content; default `nil` keeps current tests untouched.

### Reference Files

- `/Users/linus/Development/mashed/internal/uiadapter/fastpath.go:35-115` — full classifier surface.
- `/Users/linus/Development/mashed/internal/uiadapter/repair.go:81-122` — `Repairer.Run` retry loop.
- `/Users/linus/Development/mashed/internal/uiadapter/stages.go:124-...` — `RunTwoStage`.
- `/Users/linus/Development/mashed/internal/uiadapter/fallback.go:14-32` — small file, two functions.
- `/Users/linus/Development/mashed/internal/uiadapter/fallback_tiers.go:35-...` — `RunWithFallback`.
- `/Users/linus/Development/mashed/internal/uiadapter/logging.go` (Story 1) — standard attr names.
- Story 3's `testLogBuffer` helper is reused here.

### Skills

- Invoke `/simplify` after each file's instrumentation.
- Invoke `/golang-testing` for table-driven assertions on the multi-tier fallback flow.

## Acceptance Criteria

AC-4.1: `fastpath.go` emits classify outcomes
- Given a `FastPathClassifier(true, debugLogger)`
- When `Classify("y/n: continue?")` matches the YN rule
- Then the buffer contains `fastpath.classify.start` and `fastpath.classify.hit` records
- And the hit record's `rule` attr equals the matched rule name
- Given the same classifier on input that matches no rule
- Then a `fastpath.classify.skip` record is emitted instead of `hit`
- Given `NewFastPathClassifier(false, debugLogger)`
- When `Classify` is called
- Then a `fastpath.classify.disabled` record is the only fastpath emission

AC-4.2: `repair.go` emits attempt lifecycle
- Given a `Repairer` with `RepairMaxRetries=2` and a stub run that fails twice then succeeds — but max is 2 so the third call is never made
- When `Repairer.Run` executes
- Then exactly one `repair.start` record is emitted
- And exactly two `repair.attempt` records are emitted (numbered 1 and 2)
- And exactly one `repair.exhausted` record is emitted with `attempts_total=2`
- And no `repair.success` record is emitted

AC-4.3: `repair.go` emits success on first try
- Given `RepairMaxRetries=2` and a stub run that succeeds on attempt 1
- When `Repairer.Run` executes
- Then `repair.start`, `repair.attempt` (n=1), `repair.success` (n=1) records are emitted in that order
- And no `repair.failure` or `repair.exhausted` record is emitted

AC-4.4: `stages.go` emits two-stage transitions
- Given a `RunTwoStage` invocation that successfully classifies as `widget` and generates output
- When the function returns
- Then the buffer contains records in order:
  | msg                  |
  | stages.start         |
  | stages.classify.done |
  | stages.generate.start|
  | stages.final         |
- And the final record's `latency_ms_total` is >= the classify record's `latency_ms`
- And the `kind` attr equals `"widget"` on every applicable record

AC-4.5: `stages.go` emits parse error path
- Given Stage 1 returns malformed JSON
- When `RunTwoStage` runs
- Then a `stages.classify.parse_error` record with `reason="parse"` is emitted
- And no `stages.generate.start` record follows

AC-4.6: `fallback.go` and `fallback_tiers.go` emit reason and tier flow
- Given `FallbackAST("some raw", "validation:oversize", logger)` is called
- Then a `fallback.ast.build` record with `reason="validation:oversize"` is emitted
- Given `RunWithFallback` cycles through three tiers, the first two failing and the third succeeding
- Then in order: `fallback.tier.start`, `fallback.tier.enter` (idx=0), `fallback.tier.failure`, `fallback.tier.enter` (idx=1), `fallback.tier.failure`, `fallback.tier.enter` (idx=2), `fallback.tier.success` records appear
- And `fallback.tier.exhausted` is NOT emitted (because tier 3 succeeded)

AC-4.7: Sanitize discipline holds
- Given a 50-iteration smoke test feeding random raw payloads through the pipeline
- When the captured JSON log buffer is scanned
- Then no log record contains the raw payload, the rejected `ParseStageKind` input string, or any `err.Error()` content
- And every `reason` value belongs to a known enum (validator-reason, classify-error enum, or tier name)

AC-4.8: Hot-path zero-allocation when Debug is off
- Given the package logger at level Info
- When `RunTwoStage`, `Repairer.Run`, `FastPathClassifier.Classify`, and `RunWithFallback` are exercised 10,000 times
- Then `testing.AllocsPerRun` reports no extra allocations attributable to this story's instrumentation

## BDD Test Scenarios

### Scenario 1: Fast path classification telemetry

```gherkin
Feature: fastpath.go classify telemetry

  Scenario: Match emits hit
    Given a FastPathClassifier(enabled=true) with Debug logger
    When I call Classify("Continue? [y/N]")
    Then a fastpath.classify.start record is emitted with bytes_in=16, enabled=true
    And a fastpath.classify.hit record is emitted with rule="yn"
    And the records have non-zero latency_ms

  Scenario: Skip when nothing matches
    Given a FastPathClassifier(enabled=true) with Debug logger
    When I call Classify("just some prose")
    Then a fastpath.classify.skip record is emitted

  Scenario: Disabled short-circuit
    Given a FastPathClassifier(enabled=false) with Debug logger
    When I call Classify("anything")
    Then exactly one record is emitted: fastpath.classify.disabled
```

### Scenario 2: Repair retry trail

```gherkin
Feature: repair.go retry telemetry

  Scenario: Exhaust retries
    Given a Repairer with maxRetries=2 and a stub that always fails with validator reason "oversize"
    When Repairer.Run is invoked
    Then records appear in order:
      | msg              | attempt | reason   |
      | repair.start     |         |          |
      | repair.attempt   | 1       | oversize |
      | repair.failure   | 1       | oversize |
      | repair.attempt   | 2       | oversize |
      | repair.failure   | 2       | oversize |
      | repair.exhausted |         |          |
    And the repair.exhausted record has attempts_total=2 and final_reason="oversize"

  Scenario: Succeed on first try
    Given a Repairer with maxRetries=2 and a stub that succeeds on attempt 1
    When Repairer.Run is invoked
    Then records repair.start, repair.attempt (n=1), repair.success (n=1) are emitted in order
    And no repair.failure or repair.exhausted record appears
```

### Scenario 3: Two-stage execution trail

```gherkin
Feature: stages.go RunTwoStage telemetry

  Scenario: Happy path widget classification
    Given a RunTwoStage invocation that classifies "widget" then generates a UIAST
    When the function returns
    Then records appear in order: stages.start, stages.classify.done, stages.generate.start, stages.final
    And every record's kind attr equals "widget"
    And stages.final.latency_ms_total >= stages.classify.done.latency_ms

  Scenario: Stage 1 parse error
    Given a stub stage 1 returning invalid JSON
    When RunTwoStage runs
    Then a stages.classify.parse_error record is emitted with reason="parse"
    And no stages.generate.start record follows
```

### Scenario 4: Fallback tier flow

```gherkin
Feature: fallback_tiers.go tier escalation

  Scenario: Third-tier success
    Given a RunWithFallback configured with three tiers
    And the first two tiers fail with reason "transport"
    And the third tier succeeds
    When the function runs
    Then records appear in order:
      | msg                    | tier        |
      | fallback.tier.start    |             |
      | fallback.tier.enter    | <tier-0>    |
      | fallback.tier.failure  | <tier-0>    |
      | fallback.tier.enter    | <tier-1>    |
      | fallback.tier.failure  | <tier-1>    |
      | fallback.tier.enter    | <tier-2>    |
      | fallback.tier.success  | <tier-2>    |
    And no fallback.tier.exhausted record appears

  Scenario: All tiers exhausted
    Given the same setup but tier 3 also fails
    When the function runs
    Then a fallback.tier.exhausted record is the last entry with tiers_tried=3
```

### Scenario 5: Sanitize discipline

```gherkin
Feature: §14 sanitize discipline

  Scenario: 50 random payloads leak nothing
    Given 50 random raw strings of varying lengths
    And a Debug-level logger writing to a single bytes.Buffer
    When each payload is fed through fastpath.Classify, Repairer.Run, and RunTwoStage stubs
    Then for every JSON record decoded from the buffer:
      - no value field contains any byte sequence from the raw payload
      - no value field contains any err.Error() text
      - every "reason" attribute belongs to a closed enum
```

## Tasks / Subtasks

- [ ] Task 1: Instrument `fastpath.go` (AC: 4.1, 4.7, 4.8)
  - [ ] Subtask 1a: Emit `fastpath.classify.start` at `Classify` entry behind `Enabled` guard.
  - [ ] Subtask 1b: Emit `fastpath.classify.hit` on rule match (with `rule` attr); `fastpath.classify.skip` on no match.
  - [ ] Subtask 1c: Emit `fastpath.classify.disabled` when `f.enabled == false`.

- [ ] Task 2: Instrument `repair.go` (AC: 4.2, 4.3, 4.7, 4.8)
  - [ ] Subtask 2a: Emit `repair.start` at `Run` entry.
  - [ ] Subtask 2b: Inside the retry loop, emit `repair.attempt` on each iteration before the attempt; `repair.success` or `repair.failure` after.
  - [ ] Subtask 2c: Emit `repair.exhausted` after the loop if no success.
  - [ ] Subtask 2d: Emit `repair.prompt.build` in `BuildRepairPrompt`.

- [ ] Task 3: Instrument `stages.go` (AC: 4.4, 4.5, 4.7, 4.8)
  - [ ] Subtask 3a: Emit `stages.start` at `RunTwoStage` entry.
  - [ ] Subtask 3b: Emit `stages.classify.done` after Stage 1; `stages.generate.start` before Stage 2; `stages.final` at return.
  - [ ] Subtask 3c: Emit `stages.classify.parse_error` on stage-1 parse failure.
  - [ ] Subtask 3d: Emit `stages.assemble` in `AssembleStage1` and `AssembleStage2` with `stage` and `bytes_out`.
  - [ ] Subtask 3e: Emit `stages.kind.invalid` in `ParseStageKind` on rejection (`input_len` only).

- [ ] Task 4: Instrument `fallback.go` (AC: 4.6, 4.7)
  - [ ] Subtask 4a: Emit `fallback.ast.build` in `FallbackAST` with `reason`, `bytes_in`.
  - [ ] Subtask 4b: Emit `fallback.ast.truncate` in `firstLine` when truncation occurs.

- [ ] Task 5: Instrument `fallback_tiers.go` (AC: 4.6, 4.7, 4.8)
  - [ ] Subtask 5a: Emit `fallback.tier.start` at `RunWithFallback` entry with `tiers_count`.
  - [ ] Subtask 5b: Emit `fallback.tier.enter` at each iteration with `tier`, `attempt_index`.
  - [ ] Subtask 5c: Emit `fallback.tier.success` or `fallback.tier.failure` per attempt; `fallback.tier.exhausted` after the loop if no success.

- [x] Task 6: Tests (AC: 4.1–4.8) — RED phase
  - [x] Subtask 6a: Reuse `testLogBuffer(t)` helper from Story 3.
  - [x] Subtask 6b: Augment `fastpath_test.go`, `repair_test.go`, `stages_test.go`, `fallback_test.go`, `fallback_tiers_test.go` with assertion blocks for each AC.
  - [x] Subtask 6c: Add a sanitize-discipline fuzz-style test (50 random payloads) in a new `logging_story4_sanitize_test.go`.
  - [x] Subtask 6d: Add an alloc-budget test parallel to Story 3's, covering this story's surfaces.

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on the five instrumented files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
