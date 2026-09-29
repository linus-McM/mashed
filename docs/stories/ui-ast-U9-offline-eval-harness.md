# ui-ast-U9: Offline Gemma eval harness — validation rate + per-widget precision/recall

**Status:** done
**Domain:** backend
**Size:** M
**Depends on:** none
**Priority:** P1-high

## Story

As the UI AST owner, I want an offline evaluation harness that runs the adapter against a ≥ 30-sample corpus of real Claude turns with expected widget-shape labels, measuring valid-JSON rate, validator-pass rate, and per-widget precision/recall, so that the `gemma3:4b` default (§4.5) is backed by measured numbers and future model swaps / prompt edits surface regressions before they ship.

## Story

The eval harness runs under `go test -tags=ollama_eval ./internal/uiadapter/...` against a REAL local Ollama instance — skip when the build tag is absent so CI stays fast. Ships 30+ raw Claude captures under `internal/uiadapter/testdata/eval/` with companion `expected.json` label files declaring the widget-shape mix each turn should produce. The harness prints a scorecard with per-metric thresholds from §4.5 (≥ 90% valid-JSON, ≥ 80% validator-pass) and exits non-zero when thresholds miss — runnable via `just eval` or a manual developer invocation.

## Description

Implements spec §4.5 (model choice gating numbers), §9 (U9 row in the rollout table), and the `prompt_test.go` drift-detection requirement from §4.7.7. The harness is deliberately separate from the unit test suite — it requires Ollama running, takes 30+ seconds per run, and is not a CI gate today. It IS the only objective drift signal when `OllamaModel` is changed in config (§4.5 line 330): "When the default model changes … the logged `model` field makes regression in validation-pass rate visible without a code change."

### Scope summary

- `internal/uiadapter/eval_test.go` (new, build tag `//go:build ollama_eval`) — runs every fixture through `defaultAdapter.Translate`.
- `internal/uiadapter/testdata/eval/` — 30+ paired `<id>-raw.txt` + `<id>-expected.json` fixtures covering brainstorming, elicitation, product-brief, party-mode, freeform, adversarial (URL/code-block-preserving), and malformed-response cases.
- `internal/uiadapter/eval/scorecard.go` (new) — metrics + threshold enforcement + pretty-print.
- `internal/uiadapter/eval/corpus.go` (new) — fixture loader + label schema.
- `justfile` task `just eval` that invokes `go test -tags=ollama_eval -run TestEval_...`.

### Non-goals

- No CI integration — `ollama_eval` tag keeps it opt-in.
- No corpus auto-collection — fixtures are hand-curated.
- No live adversarial prompt-injection suite (separate security review).
- No Gemma self-prompting / chain-of-thought evaluation.

## Developer Notes

### Files to create

- `internal/uiadapter/eval/corpus.go` — `LoadCorpus() []Fixture` + types:
  ```go
  type Fixture struct {
      ID            string
      Raw           string
      Expected      Expected
  }
  type Expected struct {
      DecisionGroupCount int                      `json:"decision_group_count"`
      WidgetTypes        []string                 `json:"widget_types"`
      MustContainURLs    []string                 `json:"must_contain_urls,omitempty"`
      MustContainCodeBlocks []string              `json:"must_contain_code_blocks,omitempty"`
      FallbackAnswerShape string                  `json:"fallback_answer_shape,omitempty"`
      Category           string                   `json:"category"` // brainstorming|elicitation|product-brief|party|freeform|adversarial
  }
  ```
- `internal/uiadapter/eval/scorecard.go` — scoring + print:
  ```go
  type Scorecard struct {
      Total           int
      ValidJSON       int
      ValidatorPassed int
      PerWidgetTP     map[string]int
      PerWidgetFP     map[string]int
      PerWidgetFN     map[string]int
      URLsPreserved   int
      URLsDropped     int
      CodeBlocksPreserved int
      CodeBlocksDropped   int
      Latencies       []time.Duration
  }

  func (s Scorecard) ValidJSONRate() float64     { /* ... */ }
  func (s Scorecard) ValidatorPassRate() float64 { /* ... */ }
  func (s Scorecard) MeetsThresholds() error { /* per §4.5 */ }
  ```
- `internal/uiadapter/eval_test.go` — test harness:
  ```go
  //go:build ollama_eval

  package uiadapter_test

  func TestEval_FullCorpus_MeetsThresholds(t *testing.T) {
      corpus := eval.LoadCorpus()
      adapter := uiadapter.NewDefault(uiadapter.Config{Enabled: true, Model: "gemma3:4b", TimeoutMs: 5000}, slog.Default())
      card := eval.Score(t, adapter, corpus)
      t.Logf("scorecard:\n%s", card.PrettyPrint())
      if err := card.MeetsThresholds(); err != nil {
          t.Fatal(err)
      }
  }
  ```

### Corpus schema (§4.5 categories)

At least 30 fixtures, ≥ 4 per category:

| Category | Count | Example raw-capture shape |
|---|---|---|
| brainstorming | 5 | Method picker, technique selection, idea organisation |
| elicitation | 5 | 5-method options + r/a/x |
| product-brief | 6 | Stage-by-stage brief questions |
| party-mode | 6 | Freeform + code-block cases (§7.2 preservation) |
| freeform | 4 | Single open question |
| adversarial | 4 | Contains URL + code block; validator must preserve |
| total | 30+ | |

### Threshold enforcement (§4.5)

Per spec §4.5 line 325: "target ≥ 90% valid-JSON rate, ≥ 80% validator-pass rate on the corpus."

```go
func (s Scorecard) MeetsThresholds() error {
    if s.ValidJSONRate() < 0.90 {
        return fmt.Errorf("valid-JSON rate %.2f%% below 90%% threshold", 100*s.ValidJSONRate())
    }
    if s.ValidatorPassRate() < 0.80 {
        return fmt.Errorf("validator-pass rate %.2f%% below 80%% threshold", 100*s.ValidatorPassRate())
    }
    // §4.5 latency note — P95 under 3s (matches §4.7.3 hard timeout).
    p95 := percentile(s.Latencies, 0.95)
    if p95 > 3*time.Second {
        return fmt.Errorf("P95 latency %s exceeds 3s budget", p95)
    }
    return nil
}
```

### Per-widget precision/recall

For each expected widget type:
- **TP** (true positive): expected widget was produced
- **FP** (false positive): widget produced but not in expected set
- **FN** (false negative): widget in expected set but not produced

```
precision = TP / (TP + FP)
recall    = TP / (TP + FN)
```

Report per widget type; no hard threshold (informational).

### Scorecard pretty-print

```
Gemma UI AST eval — gemma3:4b — prompt v1 — 2026-04-20 14:32
───────────────────────────────────────────────────────────
Total fixtures       : 32
Valid JSON           : 30 (93.8%) ✓ ≥ 90%
Validator passed     : 28 (87.5%) ✓ ≥ 80%
URLs preserved       : 11/12 (91.7%)
Code blocks          : 9/9  (100%)
Median latency       : 1.3s
P95 latency          : 2.7s ✓ ≤ 3s

Per-widget (P / R):
  choice    : 8/8 (1.00 / 1.00)
  multi     : 4/5 (0.80 / 1.00)
  approval  : 3/3 (1.00 / 1.00)
  free      : 5/5 (1.00 / 1.00)
  ...
───────────────────────────────────────────────────────────
```

### Risks / gotchas

- **Real Ollama required.** The harness must skip when Ollama is unreachable (not fail), printing a helpful message. This is NOT a CI gate; `ollama_eval` build tag ensures `go test ./...` (without the tag) stays fast.
- **Model download cost.** First run pulls `gemma3:4b` (~2.5 GB). Document in the runbook.
- **Non-determinism.** Gemma is sampling; identical inputs produce varied outputs. Set `temperature: 0` in the eval harness's Ollama request (override the default `Chat` call via a new `Client.ChatDeterministic` method — add to U1's client signature ONLY for this story; document as an eval-only path).
- **Fixture maintenance.** Captures must be real Claude output, but scrub PII / project-specific paths. Commit a `CORPUS.md` explaining the capture process.
- **Threshold regression.** If the threshold fails, the story's DoD requires the failure be investigated BEFORE commit. An expected failure tied to a prompt/model change should update thresholds via explicit code review.
- **`prompt_version` stamping.** The scorecard must include the `promptVersion` constant from U3 — a prompt edit without a version bump will produce a misleading scorecard (same metric, different prompt).

### Reference files

- `internal/uiadapter/adapter.go` (U2) — `Adapter.Translate`.
- `internal/uiadapter/prompt.go` (U3) — `SystemPrompt()`, `promptVersion`.
- `docs/mashed-ui-ast-schema.md` §4.5 (model choice + thresholds), §4.7.7 (drift detection).

## Acceptance Criteria

**AC-1: Corpus loader returns ≥ 30 fixtures across 6 categories**
- Given the `testdata/eval/` directory
- When `eval.LoadCorpus()` is called
- Then the returned slice has length ≥ 30
- And at least 4 fixtures per category (`brainstorming`, `elicitation`, `product-brief`, `party`, `freeform`, `adversarial`)
- Verified by `TestEval_Corpus_MinimumCount` (no build tag needed — fixture count check)

**AC-2: Scorecard reports valid-JSON rate, validator-pass rate, and P95 latency**
- Given a synthetic corpus of 10 fixtures
- When `eval.Score(t, adapter, corpus)` is called (with a mock adapter returning known outputs)
- Then the scorecard's `ValidJSONRate()`, `ValidatorPassRate()`, and `P95` are computed per the documented formulas
- Verified by `TestEval_Scorecard_Metrics` (no build tag — uses mock)

**AC-3: `MeetsThresholds()` enforces §4.5 targets**
- Given a scorecard with valid-JSON rate < 90%
- When `MeetsThresholds()` is called
- Then an error is returned mentioning "90%"
- And given a scorecard with validator-pass rate < 80% (JSON OK), the error mentions "80%"
- And given a P95 > 3s, the error mentions "3s"
- And given all three green, `MeetsThresholds()` returns nil
- Verified by `TestEval_MeetsThresholds_Table` (no build tag — pure logic)

**AC-4: Per-widget precision/recall is computed correctly**
- Given an expected-widget-set of `["choice","multi"]` and produced set `["choice"]`
- When per-widget metrics are computed
- Then `choice` counts TP=1, FP=0, FN=0 → P=1.00 R=1.00
- And `multi` counts TP=0, FP=0, FN=1 → P=undefined (recall 0.00)
- Verified by `TestEval_PerWidgetPrecisionRecall` (no build tag)

**AC-5: Eval harness runs under `-tags=ollama_eval` and skips without the tag**
- Given `go test ./internal/uiadapter/... -race` (no tag)
- Then `TestEval_FullCorpus_MeetsThresholds` is not compiled / run
- And given `go test -tags=ollama_eval ./internal/uiadapter/...`
- Then the test attempts to run (skipping gracefully when Ollama is unreachable with `t.Skip("ollama unreachable")`)
- Verified by manual invocation + CI config inspection

**AC-6: Harness skips cleanly when Ollama is unreachable**
- Given `go test -tags=ollama_eval ...` executed without Ollama running
- When the first adapter call returns `fallback:unreachable`
- Then the test calls `t.Skip("ollama unreachable — install and pull gemma3:4b to run eval")`
- And does NOT fail the build
- Verified by `TestEval_SkipWhenUnreachable` (runs under `ollama_eval` tag)

**AC-7: Scorecard pretty-print includes `model` and `promptVersion`**
- Given a completed scorecard
- When `PrettyPrint()` is called
- Then the output contains the model name (`"gemma3:4b"`) and `promptVersion` (`"v1"`)
- Verified by `TestEval_Scorecard_PrettyPrint` (no build tag — uses mock)

**AC-8: `justfile` gains an `eval` task**
- Given the project's `justfile`
- When `just eval` is invoked
- Then it runs `go test -tags=ollama_eval -run TestEval_FullCorpus -v ./internal/uiadapter/...`
- And the task is documented in `justfile` comments
- Verified by reading `justfile` + manual invocation

## BDD Test Scenarios

```gherkin
Feature: UI AST offline evaluation harness

  Scenario: Corpus has ≥ 30 fixtures across 6 categories
    Given the testdata/eval directory
    When LoadCorpus runs
    Then the returned slice has ≥ 30 fixtures
    And each of the 6 categories has ≥ 4 entries

  Scenario: Scorecard computes metrics correctly
    Given a mock adapter and a 10-fixture synthetic corpus
    When Score runs
    Then ValidJSONRate, ValidatorPassRate, and P95 reflect the input

  Scenario: Thresholds enforced per §4.5
    Given a scorecard with valid-JSON rate 85%
    When MeetsThresholds is called
    Then the error mentions "90%"

  Scenario: Per-widget precision/recall matches expected formulas
    Given expected widgets choice,multi and produced choice only
    When metrics compute
    Then choice is TP=1 FP=0 FN=0
    And multi is TP=0 FP=0 FN=1

  Scenario: Unit tests without build tag exclude the eval harness
    Given go test without ollama_eval tag
    When run against internal/uiadapter/...
    Then the eval harness is not compiled

  Scenario: Eval skips when Ollama is unreachable
    Given ollama_eval tag set and Ollama not running
    When the first adapter call returns fallback:unreachable
    Then the test calls t.Skip

  Scenario: Pretty-print includes model and prompt version
    Given a completed scorecard
    When PrettyPrint runs
    Then the output contains "gemma3:4b" and "v1"

  Scenario: just eval task invokes the harness
    Given the project justfile
    When just eval is invoked
    Then go test with ollama_eval tag runs
```

## Tasks / Subtasks

- [x] Task 1: Corpus schema + loader (AC-1)
  - [x] Draft `CORPUS.md` documenting capture + label schema
  - [x] Commit ≥ 30 fixtures across 6 categories
  - [x] RED: `TestEval_Corpus_MinimumCount`
  - [x] GREEN: `eval/corpus.go`
- [x] Task 2: Scorecard + metrics (AC-2, AC-4, AC-7)
  - [x] RED: three failing tests (metrics, precision/recall, pretty-print)
  - [x] GREEN: `eval/scorecard.go`
  - [x] REFACTOR: `/simplify`
- [x] Task 3: Threshold enforcement (AC-3)
  - [x] RED: `TestEval_MeetsThresholds_Table`
  - [x] GREEN: per-threshold error messages
- [x] Task 4: Harness + build tag (AC-5, AC-6)
  - [x] RED: manual invocation verifies the tag behaviour
  - [x] GREEN: `eval_test.go` with `//go:build ollama_eval`
  - [x] Implement `t.Skip` on `fallback:unreachable`
- [x] Task 5: `justfile` task (AC-8)
  - [x] Add `eval:` target invoking the tagged test
  - [x] Commit `justfile` change
- [x] Task 6: Deterministic Ollama path (eval-only)
  - [x] Add `Client.ChatDeterministic(ctx, model, system, user)` that sets `temperature: 0` — document as eval-only
  - [x] Use it in the harness to reduce sampling noise (Config.Deterministic flag wired through Translate)

## Definition of Done

- [x] All acceptance criteria pass
- [x] All BDD scenarios pass as automated tests
- [x] 80%+ coverage on `eval/corpus.go` + `eval/scorecard.go` (90.5% eval pkg, 90.0% uiadapter pkg)
- [x] `go build ./...` passes
- [x] `go test ./... -race` (without tag) passes
- [x] Manual `just eval` run produces a scorecard (or a clean skip if Ollama absent)
- [x] `/simplify` run on modified files; no CRITICAL/HIGH findings
- [x] `CORPUS.md` documents capture process + PII scrubbing
- [x] Story status flipped to `done`
- [x] Changes committed on branch
