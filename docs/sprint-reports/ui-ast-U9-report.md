# Sprint Report — ui-ast-U9: Offline Gemma eval harness

**Date:** 2026-04-22
**Story:** `docs/stories/ui-ast-U9-offline-eval-harness.md`
**Status:** done
**Commit:** `e9d3592 feat(ui-ast/U9): offline Gemma eval harness`

## Team

| Role | Agent | Phase |
|------|-------|-------|
| test-writer | go-svelte-test | RED — authored 7 failing tests across 4 files |
| go-engineer | golang-pro | GREEN — corpus + scorecard + client + 30 fixtures + justfile |
| fix-engineer | golang-pro | F-1 wiring — Config.Deterministic + branch in Translate |
| reviewer | coderabbit:code-reviewer (async) | QG-3 code review |
| security-check | general-purpose (async) | QG-3 security review |
| perf-auditor | golang-pro (async) | QG-3 perf review |
| spec-reviewer | spec-adversarial-reviewer (async) | QG-4 AC verification |

## Files

**New**
- `internal/uiadapter/eval/corpus.go` — Fixture/Expected types + LoadCorpus via //go:embed
- `internal/uiadapter/eval/corpus_test.go`
- `internal/uiadapter/eval/scorecard.go` — Score, MeetsThresholds, PerWidgetPrecision/Recall, PrettyPrint
- `internal/uiadapter/eval/scorecard_test.go`
- `internal/uiadapter/eval/testdata/` — 30 fixtures × 2 files + CORPUS.md
- `internal/uiadapter/eval_test.go` — `//go:build ollama_eval` harness
- `internal/uiadapter/testhelper_eval.go` — exported WithOllamaHost under tag (breaks import cycle)

**Modified**
- `internal/uiadapter/adapter.go` — Config.Deterministic + chat() branch
- `internal/uiadapter/adapter_test.go` — TestAdapter_DeterministicRoutesThroughChatDeterministic
- `internal/uiadapter/client.go` — ChatDeterministic + shared chat() helper
- `internal/uiadapter/client_test.go` — ChatDeterministic tests
- `internal/uiadapter/prompt.go` — public PromptVersion() getter
- `justfile` — `eval` recipe

## AC Validation

| AC | Evidence | Result |
|----|----------|--------|
| AC-1 | `TestEval_Corpus_MinimumCount` — 30 fixtures, 6 categories ≥4 each | PASS |
| AC-2 | `TestEval_Scorecard_Metrics` — mock adapter, rates + P95 verified | PASS |
| AC-3 | `TestEval_MeetsThresholds_Table` — 4 subcases (90% / 80% / 3s / green) | PASS |
| AC-4 | `TestEval_PerWidgetPrecisionRecall` — TP/FP/FN + divide-by-zero guard | PASS |
| AC-5 | `go vet -tags=ollama_eval ./internal/uiadapter/...` clean; default `go test` excludes harness | PASS |
| AC-6 | `TestEval_SkipWhenUnreachable` — SKIP on `fallback:unreachable` | PASS |
| AC-7 | `TestEval_Scorecard_PrettyPrint` — "gemma3:4b" + "v1" present | PASS |
| AC-8 | `justfile` `eval:` recipe runs `-tags=ollama_eval -run TestEval_FullCorpus` | PASS |

## Quality Gates

| Gate | Result |
|------|--------|
| Build (`go build ./...`) | PASS |
| Vet (`go vet ./...` + `go vet -tags=ollama_eval`) | PASS |
| Tests (`go test ./internal/uiadapter/... -count=1`) | PASS |
| Race (`go test -race -short -count=1`) | PASS |
| Coverage | uiadapter 90.0%, eval 90.5% (≥80% gate met) |
| /simplify | Run per agent self-report; no CRITICAL/HIGH |
| Code review | MEDIUM findings (nil-ast defensive gap, time.Now in PrettyPrint) — accepted as cosmetic |
| Security review | M1 (loopback hardening for WithOllamaHost) — noted, not applied; no CRITICAL/HIGH |
| Perf review | No CRITICAL/HIGH; map-churn in recordWidgets LOW, corpus-scale-dependent |
| Spec review | F-1 HIGH (ChatDeterministic dead code) → FIX task #12 → wired via Config.Deterministic |
| AC Validation (Phase 4f) | SKIPPED — unit-test coverage sufficient (Phase 4f is opt-in, no UI ACs) |

## Fix Cycle

- **F-1 (HIGH, spec-reviewer):** `ChatDeterministic` existed but `defaultAdapter.Translate` still called `Chat`. Fix: added `Config.Deterministic bool`, `defaultAdapter.chat()` selector, branched on flag; `eval_test.go` harness now passes `Deterministic: true`. Coverage 100% on new `chat()` dispatch; 2 new subtests verify request body options.temperature presence/absence.

## Deviations from story

- None material. Spec's §Risks deterministic path was wired via `Config.Deterministic` rather than a separate constructor — simpler API surface, same spec semantics.

## Follow-ups (not blocking)

- Low: `sortedWidgetNames` in `scorecard.go` has 3 redundant range-over-map blocks; a `mergeMapKeys` helper would deduplicate.
- Low: `testhelper_eval.go:WithOllamaHost` could validate that the URL points to loopback as defense-in-depth against future library misuse.
- Low: `PrettyPrint` uses `time.Now()` — snapshot tests would require injected clock. Not an issue for Contains-based assertions currently in place.
