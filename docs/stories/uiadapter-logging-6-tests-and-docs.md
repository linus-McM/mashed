# Story 6: End-to-end tests, ergonomics, and docs

**Priority:** P2-medium
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** uiadapter-logging-3-instrument-cache-and-network, uiadapter-logging-4-instrument-pipeline, uiadapter-logging-5-instrument-payload-shaping
**Status:** done

## Description

Cap the logging plan with end-to-end tests, developer ergonomics, and documentation. Verify the production logger's full surface (fan-out, level filter, file lifecycle, `WithAttrs`/`WithGroup` correctness), add an E2E `Translate` test that asserts ≥ 1 log entry per pipeline phase, prove the 10-message smoke run leaks no payload bytes, ship a `just trace` recipe, update `.gitignore` and README, and record the architectural decision in `.wolf/cerebrum.md` and `.wolf/anatomy.md`. After this story all six plan acceptance criteria are satisfied.

## Developer Notes

### Architecture

This story lands three deliverables:

#### 1. `internal/uiadapter/logging_test.go` — comprehensive logger tests

Already exists from Story 1 (smoke tests for `NewProductionLogger`). Extend it with:

| Test | Asserts |
|---|---|
| `TestProductionLogger_FanoutToBothSinks` | A single `Info` call lands one line in stdout AND one JSON line in the file |
| `TestProductionLogger_LevelFilter` | `Info`-level logger drops `Debug` records from both sinks |
| `TestProductionLogger_WithAttrs` | `logger.With("k","v").Info("msg")` carries `k=v` to both sinks |
| `TestProductionLogger_WithGroup` | `logger.WithGroup("g").Info("msg","k","v")` produces JSON with nested `"g":{"k":"v"}` |
| `TestProductionLogger_FileNameUTC` | The file name matches `uiadapter-YYYYMMDD.log` where the date is `time.Now().UTC()` precision |
| `TestProductionLogger_RaceSafe` | 100 goroutines logging concurrently; no `-race` failure |
| `TestProductionLogger_NilLoggerSafety` | `nilSafeLogger(nil)` returns a discard logger; calling its methods is a no-op |

#### 2. `internal/uiadapter/translate_e2e_test.go` — end-to-end coverage

New file. One test that:
1. Constructs a `defaultAdapter` with a Debug-level buffer-backed logger.
2. Stubs the HTTP client to return a known UIAST JSON.
3. Calls `Translate(ctx, "raw", "proc-1")`.
4. Decodes every JSON line from the buffer.
5. Asserts each of the following file-source identifiers shows up at least once via the `op` attr prefix:
   - `client.*`, `cache.*`, `breaker.*`, `semaphore.*` (Story 3)
   - `fastpath.*`, `repair.*` (only if repair triggered — make a parallel test with a stub that forces a repair), `stages.*`, `fallback.*`, `fallback.tier.*` (only if tier escalation triggered — same pattern)
   - `sanitize.*`, `spotlight.*`, `contextguard.*`, `validator.*`, `encode.*`, `sampling.*`, `schema.*` (Story 5)
6. Re-runs 10 times with random raw payloads and asserts NO record contains any payload byte (the §14 sanitize-discipline smoke).

#### 3. Docs and ergonomics

- **`justfile`** — add a `trace` recipe:
  ```
  trace:
      UIADAPTER_LOG_LEVEL=debug wails dev 2>&1 | tee logs/uiadapter-trace.log
  ```
  If `justfile` does not exist yet, create it. If it does, append the recipe.
- **`.gitignore`** — append `logs/` (verify Story 1 already did this; if so, no-op).
- **`README.md`** (project root) — append a short "Debug logging" section: env var, log location, format, sanitize promise (3–6 lines).
- **`.wolf/cerebrum.md`** — append a Decision Log entry dated `2026-04-26`:
  > Decision: uiadapter ships structured slog Debug logging with stdout text + daily JSON file fanout. Toggle: `UIADAPTER_LOG_LEVEL` env var. Default `info` so prod is unchanged. Sanitize discipline §14 enforced — no raw payloads, ever.
- **`.wolf/anatomy.md`** — register new files: `internal/uiadapter/logging.go`, `internal/uiadapter/logging_test.go`, `internal/uiadapter/translate_e2e_test.go`.

### Technical Considerations

- **E2E test isolation:** use `t.TempDir()` for the log directory. Use `httptest.NewServer` for the stubbed client.
- **Race safety:** the concurrent `TestProductionLogger_RaceSafe` test must pass `-race` cleanly. The fan-out handler is stateless; the file handle is the only shared mutable state and the stdlib `JSONHandler` already serialises writes via its internal mutex.
- **Sanitize-discipline check:** implement as a substring scan of every JSON record's stringified value space against a precomputed map of "forbidden substrings" (the random raw payload, the response body, any error message that the test induces). Fail loudly with the offending record.
- **No fixture sprawl:** reuse existing test fixtures from `internal/uiadapter/testdata/` where possible; add at most one new fixture if needed for the E2E case.
- **`just` not installed:** if the team doesn't standardise on `just`, the recipe lives in a top-level `justfile` regardless — anyone who has `just` benefits, and anyone without it can copy-paste the command from the file. Do not introduce a hard `just` dependency.
- **Wails binding:** none.

### Risks & Edge Cases

- **Flaky E2E timing:** never assert exact `latency_ms` values — only `>= 0` or `<= timeout * 2`. Time-based assertions belong outside the unit suite.
- **`time.Now().UTC()` filename test:** if the test runs at midnight UTC the filename may roll. Either freeze the clock via a `clock` interface (over-engineering for this story) or accept up to two valid names (`yesterday-or-today`). Pick the second; document the rationale in a comment.
- **Per-file emission proof gaps:** `repair.go` and `fallback_tiers.go` only emit when their conditions trigger. Use *separate* test cases that force those triggers; do not gate the full per-file assertion on a single happy-path Translate.
- **Coverage measurement:** `go test ./internal/uiadapter/... -coverprofile=cover.out` then `go tool cover -func=cover.out` should report ≥ 80% on every file modified by Stories 1, 3, 4, 5.
- **README drift:** keep the new section terse — it must not duplicate the spec.

### Reference Files

- All instrumented files from Stories 1–5.
- `/Users/linus/Development/mashed/internal/uiadapter/adapter_test.go` — pattern for constructing a test adapter.
- `/Users/linus/Development/mashed/internal/uiadapter/testdata/` — existing fixtures.
- `/Users/linus/Development/mashed/.wolf/cerebrum.md` — append-only decision log.
- `/Users/linus/Development/mashed/.wolf/anatomy.md` — file registry.
- `/Users/linus/Development/mashed/.gitignore` — append `logs/` if missing.

### Skills

- Invoke `/golang-testing` for the comprehensive `logging_test.go` and `translate_e2e_test.go` table-driven scaffolding.
- Invoke `/simplify` after the test files land — they're easy to over-complicate.

## Acceptance Criteria

AC-6.1: `logging_test.go` covers fan-out, levels, attrs, groups, race
- Given the augmented `logging_test.go`
- When `go test ./internal/uiadapter/...` runs
- Then every test listed in the table above passes (`Fanout`, `LevelFilter`, `WithAttrs`, `WithGroup`, `FileNameUTC`, `RaceSafe`, `NilLoggerSafety`)
- And `go test -race` passes for the package

AC-6.2: E2E `Translate` test asserts per-phase emission
- Given a `defaultAdapter` constructed with a Debug-level captured logger and a stubbed HTTP client
- When `Translate(ctx, "raw", "proc-1")` runs to completion
- Then the captured JSON buffer contains records with `op` attrs covering at minimum: `sanitize.*`, `spotlight.*`, `contextguard.*`, `client.*`, `cache.*`, `validator.*`, `encode.*`, `sampling.*`, `breaker.*` or `semaphore.*` (one of them must fire), `fastpath.*`
- And separate parallel sub-tests prove `repair.*` and `fallback.tier.*` emit when their triggers fire (forced via stubbed validator-fail and forced tier escalation)

AC-6.3: 10-message smoke leaks zero payload bytes
- Given 10 random raw payloads (length 50–5000 bytes, mixed ASCII + control chars)
- When each is fed through `Translate` with the production logger at Debug
- Then for every record in the captured JSON log, no value-bearing attribute contains any byte sequence from the corresponding raw payload OR the stubbed response body OR any synthetic error string
- And the assertion fails loudly with the offending `(payload-index, record)` pair if violated

AC-6.4: `UIADAPTER_LOG_LEVEL=debug wails dev` smoke (manual / CI)
- Given the env var set to `debug`
- When the app boots and a single Translate is exercised
- Then stdout shows interleaved human-readable text records
- And `./logs/uiadapter-YYYYMMDD.log` contains the same records as JSON
- And on graceful shutdown the file handle is closed (verified via `lsof -p $(pidof mashed)` showing no open handle, or via the existing closer-idempotency unit test)

AC-6.5: Default level remains `info` — production behaviour unchanged
- Given `UIADAPTER_LOG_LEVEL` is unset
- When the app boots and runs a Translate
- Then stdout contains zero Debug records
- And the log file contains zero Debug records (Info+ only)
- And the existing `logTelemetry` Info-level emissions in `adapter.go` are unchanged

AC-6.6: `just trace` recipe exists and is documented
- Given a developer runs `just --list`
- Then `trace` appears in the output with the documented description
- And running `just trace` sets `UIADAPTER_LOG_LEVEL=debug` and pipes both streams to a tee'd log

AC-6.7: Docs and registry updated
- Given a code review of the diff
- When the reviewer inspects `README.md`, `.wolf/cerebrum.md`, `.wolf/anatomy.md`
- Then the README contains a "Debug logging" section ≤ 10 lines covering env var, file location, sanitize promise
- And `.wolf/cerebrum.md` has a new Decision Log entry dated `2026-04-26` referencing this plan
- And `.wolf/anatomy.md` lists the three new files (`logging.go`, `logging_test.go`, `translate_e2e_test.go`)

AC-6.8: Coverage holds at 80% across all modified files
- Given the full uiadapter test run
- When `go tool cover -func=cover.out` reports per-file coverage
- Then every file modified by Stories 1, 3, 4, 5, 6 reports ≥ 80% line coverage
- And no file's coverage decreased from its pre-plan baseline

## BDD Test Scenarios

### Scenario 1: Comprehensive logger surface

```gherkin
Feature: Production logger fan-out, levels, attrs, groups

  Scenario: Single info call lands in both sinks
    Given a logger built with NewProductionLogger(Info, tempDir)
    When I call logger.Info("msg","k","v")
    Then stdout contains "msg" and "k=v"
    And the file contains exactly one JSON line with "msg":"msg","k":"v"

  Scenario: Level filter drops Debug
    Given a logger at level Info
    When I call logger.Debug("dropped") and logger.Info("kept")
    Then neither sink contains "dropped"
    And both sinks contain "kept"

  Scenario: WithAttrs propagates to both sinks
    Given a logger l = base.With("session","abc")
    When I call l.Info("event")
    Then both sinks contain session="abc" or "session":"abc"

  Scenario: WithGroup nests JSON cleanly
    Given a logger l = base.WithGroup("g")
    When I call l.Info("msg","k","v")
    Then the JSON sink contains "g":{"k":"v"}

  Scenario: Concurrent goroutines do not race
    Given a logger from NewProductionLogger
    When 100 goroutines each call logger.Info 100 times
    Then `go test -race` reports no data race
    And the file contains exactly 10000 JSON lines (each parseable)
```

### Scenario 2: End-to-end pipeline emission

```gherkin
Feature: Translate end-to-end produces logs from every phase

  Scenario: Happy-path Translate emits per-file records
    Given a defaultAdapter with a stubbed HTTP client returning a valid UIAST
    And a Debug-level buffer-backed logger
    When Translate(ctx, "raw", "proc-1") runs
    Then the captured JSON buffer contains records whose op attrs cover:
      sanitize.*, spotlight.*, contextguard.*, client.*, cache.*, validator.*, encode.*, sampling.*, fastpath.*
    And at least one of breaker.* or semaphore.* records appears

  Scenario: Repair-triggered Translate emits repair records
    Given a stub validator that fails the first attempt and passes the second
    When Translate runs through the repair loop
    Then repair.start, repair.attempt (n=1), repair.failure (n=1), repair.attempt (n=2), repair.success (n=2) records appear

  Scenario: Tier-escalation Translate emits tier records
    Given a stub primary tier failing with reason "transport"
    When Translate falls back to a secondary tier
    Then fallback.tier.start, fallback.tier.enter, fallback.tier.failure, fallback.tier.enter, fallback.tier.success records appear
```

### Scenario 3: 10-message zero-leak smoke

```gherkin
Feature: §14 sanitize discipline holds end-to-end

  Scenario: Random payloads leak no bytes
    Given 10 random raw strings (length 50..5000, mixed ASCII + control chars)
    And a stubbed HTTP client returning a unique synthetic response per request
    And a Debug-level captured logger
    When each raw is fed through Translate
    Then for every JSON record decoded:
      - no value contains any substring of length >= 8 from the corresponding raw payload
      - no value contains any substring of length >= 8 from the response body
      - no value contains any err.Error() text injected by the test
    And on violation, the test fails with the offending (payload-index, record) pair printed
```

### Scenario 4: Defaults preserve production behaviour

```gherkin
Feature: Default level is Info

  Scenario: Unset env var keeps production quiet
    Given UIADAPTER_LOG_LEVEL is unset in the environment
    When the App boots
    Then the production logger's effective level is Info
    And a Translate request emits zero Debug records to either sink
    And the existing adapter.go logTelemetry Info emission is unchanged
```

### Scenario 5: Ergonomics and docs

```gherkin
Feature: Developer ergonomics

  Scenario: just trace recipe exists
    Given the project's justfile
    When I run `just --list`
    Then "trace" appears in the recipe list
    And running `just trace` sets UIADAPTER_LOG_LEVEL=debug and tees output to logs/uiadapter-trace.log

  Scenario: README has Debug logging section
    Given the project README.md
    When I grep for "Debug logging" or "UIADAPTER_LOG_LEVEL"
    Then a section <= 10 lines describes the env var, log location, and sanitize promise

  Scenario: cerebrum and anatomy updated
    Given .wolf/cerebrum.md and .wolf/anatomy.md
    When I read the diff
    Then cerebrum.md has a Decision Log entry dated 2026-04-26 about uiadapter slog logging
    And anatomy.md lists logging.go, logging_test.go, translate_e2e_test.go with descriptions
```

## Tasks / Subtasks

- [ ] Task 1: Extend `logging_test.go` (AC: 6.1)
  - [ ] Subtask 1a: Add `TestProductionLogger_FanoutToBothSinks`, `TestProductionLogger_LevelFilter`.
  - [ ] Subtask 1b: Add `TestProductionLogger_WithAttrs`, `TestProductionLogger_WithGroup`.
  - [ ] Subtask 1c: Add `TestProductionLogger_FileNameUTC` accepting yesterday-or-today.
  - [ ] Subtask 1d: Add `TestProductionLogger_RaceSafe` (100 goroutines * 100 calls).
  - [ ] Subtask 1e: Add `TestNilSafeLogger` confirming discard semantics.

- [ ] Task 2: Create `internal/uiadapter/translate_e2e_test.go` (AC: 6.2, 6.3)
  - [ ] Subtask 2a: Build helper `e2eAdapter(t, stubResp string)` that returns a `defaultAdapter` with a captured logger.
  - [ ] Subtask 2b: Add `TestTranslate_E2E_HappyPath_AllPhasesLog` asserting the per-phase `op` set.
  - [ ] Subtask 2c: Add `TestTranslate_E2E_RepairTriggered` and `TestTranslate_E2E_TierEscalation`.
  - [ ] Subtask 2d: Add `TestTranslate_E2E_TenMessageNoLeak` with random payload generation and substring-scan assertion.

- [ ] Task 3: Add `just trace` recipe and ergonomics (AC: 6.6)
  - [ ] Subtask 3a: Create or edit top-level `justfile` with the `trace` recipe.
  - [ ] Subtask 3b: Verify `logs/` is in `.gitignore`.

- [ ] Task 4: Documentation updates (AC: 6.7)
  - [ ] Subtask 4a: Append "Debug logging" section to `README.md`.
  - [ ] Subtask 4b: Append Decision Log entry to `.wolf/cerebrum.md` (date `2026-04-26`).
  - [ ] Subtask 4c: Add `logging.go`, `logging_test.go`, `translate_e2e_test.go` entries to `.wolf/anatomy.md` with one-line descriptions and token estimates.

- [ ] Task 5: Coverage verification (AC: 6.8)
  - [ ] Subtask 5a: Run `go test ./internal/uiadapter/... -coverprofile=cover.out`.
  - [ ] Subtask 5b: Run `go tool cover -func=cover.out`; record per-file coverage in PR description.
  - [ ] Subtask 5c: If any file falls below 80%, add targeted tests until threshold is met.

- [ ] Task 6: Manual smoke verification (AC: 6.4, 6.5)
  - [ ] Subtask 6a: `UIADAPTER_LOG_LEVEL=debug wails dev`; visually confirm interleaved stdout + JSON file.
  - [ ] Subtask 6b: Unset the env var; confirm zero Debug records emit.
  - [ ] Subtask 6c: Trigger graceful app close; confirm log file is closed (or at least the closer ran — check via the existing idempotency test plus an `App.shutdown` invocation in a unit test).

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on every file modified by the logging plan (Stories 1, 3, 4, 5, 6)
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] All six plan acceptance criteria from `docs/plans/uiadapter-debug-logging.md` are satisfied
