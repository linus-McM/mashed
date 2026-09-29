---
type: Module
title: scorecard_test.go
description: "Graphify community 25: internal/uiadapter/eval/corpus.go, internal/uiadapter/eval/corpus_test.go, internal/uiadapter/eval/scorecard.go, internal/uiadapter/eval/scorecard_test.go, internal/uiadapter/ev"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: corpus, resource: internal/uiadapter/eval/corpus.go, last_modified: "2026-04-22T14:13:03+10:00", digest: be80b9817e3d70f3 }
  - { id: corpus_test, resource: internal/uiadapter/eval/corpus_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 2df8e716d6d0fee2 }
  - { id: scorecard, resource: internal/uiadapter/eval/scorecard.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 478867f290bdf6a9 }
  - { id: scorecard_test, resource: internal/uiadapter/eval/scorecard_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: c10507f5ef6c8f7b }
  - { id: eval_test, resource: internal/uiadapter/eval_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 170e442f3eb11320 }
---

# Files
- `internal/uiadapter/eval/corpus.go`
- `internal/uiadapter/eval/corpus_test.go`
- `internal/uiadapter/eval/scorecard.go`
- `internal/uiadapter/eval/scorecard_test.go`
- `internal/uiadapter/eval_test.go`

# Symbols
- corpus.go (internal/uiadapter/eval/corpus.go:L1)
- Fixture (internal/uiadapter/eval/corpus.go:L21)
- Expected (internal/uiadapter/eval/corpus.go:L30)
- LoadCorpus() (internal/uiadapter/eval/corpus.go:L55)
- collectFixtureIDs() (internal/uiadapter/eval/corpus.go:L69)
- loadFixture() (internal/uiadapter/eval/corpus.go:L89)
- TestEval_Corpus_MinimumCount() (internal/uiadapter/eval/corpus_test.go:L13)
- scorecard.go (internal/uiadapter/eval/scorecard.go:L1)
- .recordPreservation() (internal/uiadapter/eval/scorecard.go:L112)
- .ValidJSONRate() (internal/uiadapter/eval/scorecard.go:L131)
- .ValidatorPassRate() (internal/uiadapter/eval/scorecard.go:L138)
- .P95Latency() (internal/uiadapter/eval/scorecard.go:L149)
- .PerWidgetPrecision() (internal/uiadapter/eval/scorecard.go:L168)
- .PerWidgetRecall() (internal/uiadapter/eval/scorecard.go:L177)
- Scorecard (internal/uiadapter/eval/scorecard.go:L19)
- .MeetsThresholds() (internal/uiadapter/eval/scorecard.go:L190)
- .PrettyPrint() (internal/uiadapter/eval/scorecard.go:L207)
- classifyGeneratedBy() (internal/uiadapter/eval/scorecard.go:L246)
- inferModel() (internal/uiadapter/eval/scorecard.go:L259)
- widgetTypesFrom() (internal/uiadapter/eval/scorecard.go:L267)
- renderedText() (internal/uiadapter/eval/scorecard.go:L280)
- toSet() (internal/uiadapter/eval/scorecard.go:L300)
- sortedWidgetNames() (internal/uiadapter/eval/scorecard.go:L311)
- Score() (internal/uiadapter/eval/scorecard.go:L53)
- .record() (internal/uiadapter/eval/scorecard.go:L73)
- .recordWidgets() (internal/uiadapter/eval/scorecard.go:L95)
- scorecard_test.go (internal/uiadapter/eval/scorecard_test.go:L1)
- TestEval_PerWidgetPrecisionRecall() (internal/uiadapter/eval/scorecard_test.go:L104)
- TestEval_Scorecard_PrettyPrint() (internal/uiadapter/eval/scorecard_test.go:L153)
- TestEval_MeetsThresholds_Table() (internal/uiadapter/eval/scorecard_test.go:L178)
- TestEval_Preservation_URLAndCodeBlockCounts() (internal/uiadapter/eval/scorecard_test.go:L249)
- TestEval_ScorecardRates_EmptyCorpus() (internal/uiadapter/eval/scorecard_test.go:L302)
- TestEval_PrettyPrint_IncludesPerWidgetLines() (internal/uiadapter/eval/scorecard_test.go:L313)
- newSynthFixture() (internal/uiadapter/eval/scorecard_test.go:L33)
- uniformLatencies() (internal/uiadapter/eval/scorecard_test.go:L331)
- astOllama() (internal/uiadapter/eval/scorecard_test.go:L44)
- astValidatorFailed() (internal/uiadapter/eval/scorecard_test.go:L53)
- astMalformed() (internal/uiadapter/eval/scorecard_test.go:L62)
- TestEval_Scorecard_Metrics() (internal/uiadapter/eval/scorecard_test.go:L72)
- TestEval_FullCorpus_MeetsThresholds() (internal/uiadapter/eval_test.go:L52)

# Depends on
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [NewDefault](/modules/newdefault.md)
- [prompt_test.go](/modules/prompt-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
