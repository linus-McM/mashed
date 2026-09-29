# uiadapter-05: Extend the eval harness with parse-rate + widget-shape metrics

**Status:** ready
**Domain:** backend
**Size:** L
**Depends On:** uiadapter-01, uiadapter-02
**Priority:** P1-high

## Story

As a UIAdapter maintainer, I want an explicit, versioned scorecard with enforced thresholds (parse-rate ≥0.95, terminal-validation ≤0.02, widget-type accuracy ≥0.95) and a ≥15-entry corpus, so that prompt and model edits are gated on measurable quality and the Story 2 format-flag delta (loose vs. schema-constrained) is quantifiable.

## Description

`TestEval_FullCorpus_MeetsThresholds` exists in `eval_test.go` (line 52 per plan) but the thresholds themselves aren't defined in the uploaded file. This story adds `internal/uiadapter/eval/score.go` (create the package if needed), defines the `Scorecard` struct and `MeetsThresholds()` enforcement, commits a minimum-15-entry corpus under `testdata/eval/`, and documents the CI invocation gated by the `ollama_eval` build tag. Depends on stories 1 (new prompt) and 2 (schema-constrained format) so the deltas are meaningful.

### Scope summary
- `internal/uiadapter/eval/score.go` — `Scorecard` type, `MeetsThresholds() error`, `PrettyPrint(io.Writer)` method.
- `internal/uiadapter/eval_test.go` (update) — `TestEval_FullCorpus_MeetsThresholds` wires the Scorecard and enforces thresholds.
- `internal/uiadapter/testdata/eval/` — ≥15 corpus entries covering 5 categories × 3 variants each.
- Build-tag gating: all eval code + tests under `//go:build ollama_eval`.
- PR description records the scorecard delta between `LooseFormat=true` and `LooseFormat=false` (pins the Story 2 win).

### Non-goals
- Do not remove the `ollama_eval` gate — the eval takes minutes and needs a real Ollama.
- Do not enforce thresholds in CI without an Ollama runner.
- Do not add new widget kinds to the scorecard.
- Do not alter `prompt.md`, `client.go`, or sanitise logic.

## Developer Notes

### Files to modify / create
- `internal/uiadapter/eval/score.go` (new — may or may not exist; create the `eval` subpackage if absent).
- `internal/uiadapter/eval_test.go` — update `TestEval_FullCorpus_MeetsThresholds` to call the new scorecard.
- `internal/uiadapter/testdata/eval/<category>/<variant>/{raw.txt, expected.json}` — ≥15 entries.
- CI documentation (project-level docs or `eval/README.md` co-located with the harness) noting `go test -tags ollama_eval ./internal/uiadapter/... -run TestEval_FullCorpus_MeetsThresholds`.

### Type/symbol inventory (exact names)
- `eval.Scorecard` struct with fields:
  - `Entries []EntryResult` (per-corpus-entry results)
  - `EntryResult` fields: `Name string`, `ParseOK bool`, `ValidationTerminal bool`, `Untrusted bool`, `DecisionGroupsMatched int`, `WidgetTypesMatched map[string]int`, `OptionValuesMatched int`.
- `(*Scorecard) MeetsThresholds() error` — returns a wrapped sentinel error with details when any threshold is violated; nil when green.
- `(*Scorecard) PrettyPrint(w io.Writer)` — human-readable summary table.

### Threshold definitions (exact numbers — plan §1 Story 5)
- `parseRate := parseOKCount / totalEntries; parseRate >= 0.95`
- `terminalValidation := terminalCount / totalEntries; terminalValidation <= 0.02`
- `widgetTypeAccuracy := matchedWidgets / totalWidgets; widgetTypeAccuracy >= 0.95`

Use exported constants in `score.go` so thresholds are reviewable:
```go
const (
    MinParseRate            = 0.95
    MaxTerminalValidation   = 0.02
    MinWidgetTypeAccuracy   = 0.95
)
```

### Sentinel error for fail-closed behaviour
- `var ErrThresholdViolation = errors.New("eval: threshold violation")` in `score.go`.
- `MeetsThresholds()` wraps it with `%w` and includes the specific threshold and the actual value (e.g. `fmt.Errorf("%w: parse rate %.3f < %.3f", ErrThresholdViolation, got, MinParseRate)`).
- Test asserts `errors.Is(err, ErrThresholdViolation)`.

### Corpus layout (≥15 entries)
- 5 categories × 3 variants each = 15 entries minimum:
  - `brainstorming/v1`, `brainstorming/v2`, `brainstorming/v3`
  - `elicitation/v1..v3`
  - `product-brief/v1..v3`
  - `party-mode/v1..v3`
  - `freeform/v1..v3`
- Each entry: `raw.txt` (the captured tmux turn) + `expected.json` (the canonical UIAST the scorecard measures against).
- Keep fixture raws realistic — include ANSI escapes and CLI chrome so Story 3's sanitiser is exercised indirectly.

### Build-tag gating
- `//go:build ollama_eval` at the top of every new eval file.
- `eval_test.go`'s existing gate must be preserved.
- CI must NOT run this test without an Ollama runner; document this in the PR description and in an inline comment near the build tag.

### Telemetry + log discipline
- The scorecard is a test-layer construct; do not emit `slog` from inside the scorecard itself. Use `t.Logf` in the test when you want visibility during local runs.

### Risks / gotchas
- Eval runtime: minutes per full pass. Keep the test gated and note the runtime in `PrettyPrint`.
- Non-determinism: Ollama inference is non-deterministic even at `temperature=0`. Thresholds are statistical — do not assert per-entry equality in the scorecard; assert shape/type matches.
- Corpus authorship: expected.json must be produced by careful review, not by running the model against itself. The fixtures are the ground truth.
- Story 2 dependency: `LooseFormat=false` is the default after Story 2; the scorecard delta reproduction requires toggling the flag at the ClientConfig level before each run.
- Story 1 dependency: without the rewritten prompt, parse-rate may legitimately miss 0.95 — running this story before Story 1 lands will fail the gate for the wrong reason.

### Reference files
- `internal/uiadapter/eval_test.go` — existing gated test.
- `internal/uiadapter/testdata/prompts/` — example fixture shape to mimic for eval corpus.
- `docs/plans/IMPLEMENTATION_PLAN.md` §1 Story 5.

## Acceptance Criteria

**AC-5.1: `Scorecard.PrettyPrint` produces a human-readable summary**
- Given a populated `Scorecard`
- When `PrettyPrint(os.Stdout)` is called
- Then the output lists per-entry results (name, parseOK, validationTerminal, untrusted) AND aggregate rates (parse-rate, terminal-validation, widget-type accuracy)
- And the output fits on a single screen for a 15-entry corpus

**AC-5.2: `MeetsThresholds()` enforces thresholds and fails closed**
- Given a `Scorecard` with parse-rate 0.80 (below the 0.95 threshold)
- When `MeetsThresholds()` is called
- Then the returned error is non-nil
- And `errors.Is(err, ErrThresholdViolation) == true`
- And the error message names the violated threshold and the actual value

**AC-5.3: Corpus committed with ≥15 entries; scorecard green locally against `gemma3:4b`**
- Given the corpus under `testdata/eval/`
- When `go test -tags ollama_eval ./internal/uiadapter/... -run TestEval_FullCorpus_MeetsThresholds` runs on a dev machine with Ollama serving `gemma3:4b`
- Then the test passes
- And the corpus contains at least 15 `{raw.txt, expected.json}` pairs across the five categories

**AC-5.4: Scorecard delta (LooseFormat true vs. false) recorded in PR**
- Given Story 2's `ClientConfig.LooseFormat` flag
- When the scorecard runs twice (once with `LooseFormat=true`, once with `LooseFormat=false`)
- Then both scorecards are pasted into the PR description
- And the parse-rate delta is called out explicitly

**AC-5.5: Eval is gated by `ollama_eval` build tag**
- Given CI without an Ollama runner
- When `go test ./...` runs (without `-tags ollama_eval`)
- Then no eval tests are compiled or executed
- And `go build ./...` still passes

## BDD Test Scenarios

### Scenario 1: Threshold enforcement

```gherkin
Feature: Scorecard threshold gate

  Scenario: Below parse-rate threshold fails closed
    Given a Scorecard populated with 10 entries where 7 have ParseOK=true
    When MeetsThresholds is called
    Then the returned error wraps ErrThresholdViolation
    And the error message mentions "parse rate"

  Scenario: Green scorecard returns nil
    Given a Scorecard populated with 15 entries where 15 have ParseOK=true, 0 have ValidationTerminal=true, and widget-type accuracy is 1.0
    When MeetsThresholds is called
    Then the returned error is nil
```

### Scenario 2: PrettyPrint shape

```gherkin
Feature: Human-readable summary

  Scenario: PrettyPrint emits per-entry and aggregate rows
    Given a populated Scorecard
    When PrettyPrint writes to a buffer
    Then the buffer contains one row per entry
    And the buffer contains an "aggregate" row with parse-rate, terminal-validation, widget-type accuracy
```

### Scenario 3: Corpus coverage

```gherkin
Feature: Corpus categories and counts

  Scenario: Fifteen entries across five categories
    Given testdata/eval/ after this story lands
    When a helper walks the tree and counts {raw.txt, expected.json} pairs
    Then the total is at least 15
    And each of brainstorming, elicitation, product-brief, party-mode, freeform has at least 3 entries
```

### Scenario 4: Build-tag gating

```gherkin
Feature: Eval tests only compile under ollama_eval

  Scenario: Default build ignores the eval subpackage tests
    Given a checkout at this story's HEAD
    When go test ./... runs without the ollama_eval tag
    Then zero tests with names matching TestEval_ are executed
    And the run exits 0
```

## Tasks / Subtasks

- [ ] Task 1: Scaffold `eval` subpackage (AC: 5.1, 5.2, 5.5)
  - [ ] Create `internal/uiadapter/eval/score.go` with `//go:build ollama_eval`
  - [ ] Declare threshold constants and the `ErrThresholdViolation` sentinel
  - [ ] Implement `Scorecard`, `EntryResult`, `MeetsThresholds()`, `PrettyPrint(io.Writer)`
- [ ] Task 2: Write corpus (AC: 5.3)
  - [ ] Create five category directories under `internal/uiadapter/testdata/eval/`
  - [ ] Author 3 `{raw.txt, expected.json}` pairs per category
  - [ ] Include ANSI / CLI chrome in at least one raw per category so sanitiser coverage is exercised
- [ ] Task 3: Wire `TestEval_FullCorpus_MeetsThresholds` to the new Scorecard (AC: 5.1, 5.2, 5.3, 5.5)
  - [ ] Ensure the test file keeps `//go:build ollama_eval`
  - [ ] Iterate the corpus, populate Scorecard, assert `MeetsThresholds() == nil`
  - [ ] Call `PrettyPrint(os.Stderr)` with `t.Logf` for visibility
- [ ] Task 4: Test the threshold logic with a fabricated Scorecard (AC: 5.2)
  - [ ] Unit test `TestScorecard_MeetsThresholds` under `//go:build ollama_eval` or unguarded (prefer unguarded — pure logic)
  - [ ] Cover below-threshold, at-threshold, above-threshold cases
  - [ ] Assert `errors.Is(..., ErrThresholdViolation)`
- [ ] Task 5: Document the CI invocation and the Story 2 delta (AC: 5.4, 5.5)
  - [ ] Add a short `README.md` or inline-file comment with the invocation and runtime
  - [ ] Run the scorecard twice locally (LooseFormat true/false), capture results
  - [ ] Paste both scorecards into the PR description
- [ ] Task 6: Pre-flight and handoff (AC: all)
  - [ ] `go build ./...`, `go vet ./...`
  - [ ] `go test ./... -race -short` (untagged — eval stays dormant)
  - [ ] `go test -tags ollama_eval ./internal/uiadapter/... -run TestEval_FullCorpus_MeetsThresholds` on local dev (requires Ollama)
  - [ ] `/simplify` on `score.go`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests (threshold logic unguarded; corpus gated)
- [ ] 80%+ coverage on `score.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race -short` passes
- [ ] `go test -tags ollama_eval ./internal/uiadapter/... -run TestEval_FullCorpus_MeetsThresholds` green locally
- [ ] `/simplify` run on all modified code
- [ ] Pre-flight: ≥15 corpus entries, thresholds in constants, fail-closed on violation, build-tag gate intact
- [ ] AC Validation Table complete

### AC Validation Table (fill in PR description)

| AC | Test / Evidence | Status |
|----|-----------------|--------|
| AC-5.1 | `TestScorecard_PrettyPrint` (unguarded) | |
| AC-5.2 | `TestScorecard_MeetsThresholds` (unguarded, `errors.Is`) | |
| AC-5.3 | `TestEval_FullCorpus_MeetsThresholds` (`-tags ollama_eval`) | |
| AC-5.4 | Paired scorecards pasted into PR description | |
| AC-5.5 | `go test ./... -race -short` runs with zero `TestEval_*` compiled | |
