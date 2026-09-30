# uiadapter-03: Strip ANSI + CLI chrome upstream of `adapter.Translate`

**Status:** ready
**Domain:** backend
**Size:** M
**Depends On:** none
**Priority:** P0-critical

## Story

As a UIAdapter maintainer, I want captured tmux turns sanitised of ANSI CSI/OSC sequences and common CLI chrome before `adapter.Translate` sees them, so that `validator.contentPreserved`'s byte-for-byte URL/code-block containment checks stop failing on cosmetic differences and the adapter stops stamping `Diagnostics.Untrusted=true` for clean ASTs.

## Description

`validator.go:239` runs `strings.Contains` on URLs extracted via the regex `https?://[^\s"'<>\])}]+`, which greedily scoops trailing ANSI reset codes. The model strips ANSI in its output, so the containment check fails and every capture with a wrapped URL gets `Untrusted=true`. Fixing this at the source (sanitise before `Translate`) is cleaner than loosening the validator. Runs in parallel with stories 1, 2, 4.

### Scope summary
- New file `internal/uiadapter/sanitize.go` with `func SanitizeCapture(raw string) string`.
- New file `internal/uiadapter/sanitize_test.go` with golden table tests.
- Single call-site change in `internal/executor.go` near the existing `translateForPrompt` call (plan references "around line 2437").
- Additive telemetry attribute `sanitize_delta_bytes` on the existing `uiadapter.translate` slog line.
- New end-to-end test `TestAdapter_Translate_ANSIWrappedURL_NotUntrusted` in the existing adapter test file.

### Non-goals
- Do not loosen `validator.contentPreserved` logic. The fix is at the sanitizer boundary only.
- Do not rewrite the URL regex inside the validator.
- Do not add new widget kinds or schema changes.
- Do not sanitise inside fenced code blocks beyond pass-through — a test asserts fences round-trip unchanged.

## Developer Notes

### Files to modify / create
- `internal/uiadapter/sanitize.go` (new) — `func SanitizeCapture(raw string) string`.
- `internal/uiadapter/sanitize_test.go` (new) — golden table tests.
- `internal/uiadapter/adapter.go` — add the `sanitize_delta_bytes` attribute to the `uiadapter.translate` slog line (preserve the single-line invariant from plan §3).
- `internal/uiadapter/adapter_test.go` — add `TestAdapter_Translate_ANSIWrappedURL_NotUntrusted`.
- `internal/executor.go` — single call-site change: `raw = uiadapter.SanitizeCapture(raw)` immediately before `adapter.Translate(...)` (near the existing `translateForPrompt` call around line 2437).

### Type/symbol inventory (exact names)
- `uiadapter.SanitizeCapture(raw string) string` — package-exported function.
- Telemetry attribute: `sanitize_delta_bytes` (int, on the existing `uiadapter.translate` slog line).

### Regex discipline (keep narrow)
Plan §1 Story 3 risk note: overly-aggressive sanitisation could eat legitimate backticks or brackets in code blocks. Keep regexes narrow:
- CSI: `\x1b\[[0-9;?]*[a-zA-Z]`
- OSC (terminated by BEL): `\x1b\].*?\x07`
- Cursor-position and screen-clear variants (`\x1b\[2K`, `\x1b\[H`) are covered by the CSI pattern.
- Common tmux status-bar fragments and box glyphs — enumerate only ones known to appear in captures; do NOT generalise.
- Leading/trailing whitespace trim via `strings.TrimSpace` at the end.

### Required tests (golden table in `sanitize_test.go`)
- `TestSanitize_StripsANSI_Golden` with four scenarios:
  1. Bare ANSI color: `"\x1b[32mhello\x1b[0m"` → `"hello"`.
  2. Cursor-position: `"\x1b[2Ktext"` → `"text"`.
  3. OSC-8 hyperlink: `"\x1b]8;;https://x\x07link\x1b]8;;\x07"` → `"link"` (URL preserved if also visible elsewhere; the OSC sequence itself is stripped).
  4. Pathological mixed: ANSI + tmux status + box glyphs + spinner frames, all stripped.
- `TestSanitize_PreservesFencedCodeBlocks` — a raw with ```` ```go\nfunc…\n``` ```` round-trips unchanged (backticks, brackets, whitespace inside the fence preserved).

### End-to-end test
- `TestAdapter_Translate_ANSIWrappedURL_NotUntrusted` — feeds a raw capture with an ANSI-wrapped URL (e.g. `"visit \x1b[34mhttps://example.com\x1b[0m for docs"`) through `adapter.Translate` with a stubbed transport that returns a minimal valid UIAST containing `https://example.com`. Asserts `ast.Diagnostics.Untrusted == false`.

### Telemetry (single-line invariant)
Plan §3: "Preserve the single structured `uiadapter.translate` log line per Translate call. New attributes are additive only." Add `sanitize_delta_bytes` as a new `slog.Int` attribute on that same line — do NOT emit a second log line.

### Risks / gotchas
- The call-site in `executor.go` is near line 2437 (plan reference). Search for `translateForPrompt(` to confirm the exact current line — the file evolves. There is one insertion point.
- Sanitise must be idempotent: `SanitizeCapture(SanitizeCapture(x)) == SanitizeCapture(x)`. Add a test.
- Empty input: `SanitizeCapture("") == ""`. Edge case — include in the table.
- Unicode box glyphs (U+2500..U+257F) — decide whether to strip or preserve. Default: preserve (they can appear in intentional output). Only strip specific known-chrome sequences.
- `sanitize_delta_bytes` can be `0` on clean input — log it anyway to keep telemetry shape consistent across calls.

### Reference files
- `internal/uiadapter/validator.go` line 239 — the `contentPreserved` check this story works around.
- `internal/uiadapter/adapter.go` — existing `uiadapter.translate` slog line (single-line invariant).
- `internal/executor.go` — the single call-site; search for `translateForPrompt`.
- `docs/plans/IMPLEMENTATION_PLAN.md` §1 Story 3.

## Acceptance Criteria

**AC-3.1: `TestSanitize_StripsANSI_Golden` passes four scenarios**
- Given the four scenarios (bare ANSI, cursor-position, OSC-8, pathological mixed)
- When `go test ./internal/uiadapter/... -run TestSanitize_StripsANSI_Golden` runs
- Then each expected output matches the actual sanitiser output byte-for-byte

**AC-3.2: Fenced code blocks round-trip unchanged**
- Given a raw input containing a triple-backtick fenced code block with backticks, brackets, and whitespace inside
- When `SanitizeCapture` runs on it
- Then the fence and its contents are byte-identical to the input

**AC-3.3: End-to-end — ANSI-wrapped URL no longer triggers Untrusted**
- Given a raw capture containing `"visit \x1b[34mhttps://example.com\x1b[0m for docs"`
- And a stubbed transport returning a minimal valid UIAST that mentions `https://example.com`
- When `adapter.Translate` runs (sanitise is called upstream via executor, but the test exercises the adapter path with a pre-sanitised raw to exercise the validator)
- Then `ast.Diagnostics.Untrusted == false`

**AC-3.4: Telemetry line includes `sanitize_delta_bytes`**
- Given a raw input where sanitise removes N bytes
- When `adapter.Translate` runs and emits the `uiadapter.translate` slog line
- Then that single line includes an integer attribute `sanitize_delta_bytes == N`
- And no additional log lines are emitted per Translate call

**AC-3.5: Executor wires sanitise into the translate path**
- Given the updated `internal/executor.go`
- When the code path around the existing `translateForPrompt` call is inspected
- Then `uiadapter.SanitizeCapture(raw)` is invoked immediately before `adapter.Translate`
- And there is exactly one such call in the executor

**AC-3.6: Idempotence**
- Given any input `raw`
- When `SanitizeCapture` is called twice in succession
- Then `SanitizeCapture(SanitizeCapture(raw)) == SanitizeCapture(raw)`

## BDD Test Scenarios

### Scenario 1: Golden sanitise cases

```gherkin
Feature: CLI chrome sanitiser

  Scenario: Bare ANSI color is stripped
    Given raw equal to "\x1b[32mhello\x1b[0m"
    When SanitizeCapture is called
    Then the result equals "hello"

  Scenario: OSC-8 hyperlink envelope is stripped, visible text survives
    Given raw equal to "\x1b]8;;https://x\x07link\x1b]8;;\x07"
    When SanitizeCapture is called
    Then the result contains "link"
    And the result does not contain the byte "\x1b"

  Scenario: Fenced code block round-trips unchanged
    Given raw equal to "```go\nfunc x() {}\n```"
    When SanitizeCapture is called
    Then the result equals the input byte-for-byte
```

### Scenario 2: End-to-end Untrusted gate

```gherkin
Feature: Validator no longer flags ANSI-wrapped URLs as Untrusted

  Scenario: ANSI-wrapped URL survives the contentPreserved check
    Given a raw capture "visit \x1b[34mhttps://example.com\x1b[0m for docs"
    And a stubbed adapter transport returning a UIAST containing https://example.com
    When adapter.Translate runs on the sanitised raw
    Then the resulting AST has Diagnostics.Untrusted equal to false
```

### Scenario 3: Telemetry shape

```gherkin
Feature: Additive telemetry

  Scenario: sanitize_delta_bytes present on the translate log line
    Given a raw input where SanitizeCapture removes exactly 12 bytes
    When adapter.Translate runs with a slog handler that captures records
    Then exactly one record named "uiadapter.translate" is emitted
    And that record has an integer attribute sanitize_delta_bytes equal to 12
```

### Scenario 4: Executor call-site

```gherkin
Feature: Executor wires sanitise into the translate path

  Scenario: Single call-site upstream of Translate
    Given internal/executor.go after this story
    When the translateForPrompt call-site is inspected
    Then SanitizeCapture is invoked exactly once immediately before adapter.Translate
```

## Tasks / Subtasks

- [ ] Task 1: Implement `SanitizeCapture` (AC: 3.1, 3.2, 3.6)
  - [ ] Create `internal/uiadapter/sanitize.go` with narrow CSI + OSC regex
  - [ ] Add `strings.TrimSpace` at the end
  - [ ] Ensure idempotence (simple: run the pipeline twice in a unit assertion)
- [ ] Task 2: Write `sanitize_test.go` golden table (AC: 3.1, 3.2, 3.6)
  - [ ] Cover bare ANSI, cursor-position, OSC-8, pathological mixed
  - [ ] Cover fenced-code-block round-trip
  - [ ] Cover idempotence
- [ ] Task 3: Add `sanitize_delta_bytes` telemetry (AC: 3.4)
  - [ ] Locate the existing `uiadapter.translate` slog line in `adapter.go`
  - [ ] Add `slog.Int("sanitize_delta_bytes", pre-post)` as an additive attribute
  - [ ] Preserve the single-line invariant
- [ ] Task 4: Wire the sanitiser in `executor.go` (AC: 3.5)
  - [ ] Search for `translateForPrompt(` in `internal/executor.go`
  - [ ] Insert `raw = uiadapter.SanitizeCapture(raw)` immediately before `adapter.Translate`
  - [ ] Confirm exactly one call-site
- [ ] Task 5: Add end-to-end `TestAdapter_Translate_ANSIWrappedURL_NotUntrusted` (AC: 3.3)
  - [ ] Stub the transport to return a minimal valid UIAST referencing the URL
  - [ ] Assert `Diagnostics.Untrusted == false`
- [ ] Task 6: Pre-flight and handoff (AC: all)
  - [ ] `go build ./...`, `go vet ./...`
  - [ ] `go test ./internal/uiadapter/... -race -short`
  - [ ] `go test ./... -race -short` (executor call-site sanity)
  - [ ] `/simplify` on `sanitize.go` and the executor call-site change

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ coverage on `sanitize.go` and the executor diff
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race -short` passes (uiadapter + executor)
- [ ] `/simplify` run on all modified code
- [ ] Pre-flight: single call-site in executor, single slog line, idempotence verified
- [ ] AC Validation Table complete

### AC Validation Table (fill in PR description)

| AC | Test | Status |
|----|------|--------|
| AC-3.1 | `TestSanitize_StripsANSI_Golden` | |
| AC-3.2 | `TestSanitize_PreservesFencedCodeBlocks` | |
| AC-3.3 | `TestAdapter_Translate_ANSIWrappedURL_NotUntrusted` | |
| AC-3.4 | Slog capture asserts `sanitize_delta_bytes` attribute | |
| AC-3.5 | Manual diff inspection of `internal/executor.go` | |
| AC-3.6 | Idempotence assertion in `sanitize_test.go` | |
