---
type: Module
title: mashed/internal/uiadapter.UIAST
description: "Graphify community 38: internal/bmad/executor_adapter_test.go, internal/uiadapter/backend/stubs.go, internal/uiadapter/eval/corpus.go, internal/uiadapter/eval/corpus_test.go, internal/uiadapter/eval/s"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: executor_adapter_test, resource: internal/bmad/executor_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f79102c9f400d28f }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 60bb478e1146d7a5 }
  - { id: corpus, resource: internal/uiadapter/eval/corpus.go, last_modified: "2026-04-22T14:13:03+10:00", digest: be80b9817e3d70f3 }
  - { id: corpus_test, resource: internal/uiadapter/eval/corpus_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 2df8e716d6d0fee2 }
  - { id: scorecard, resource: internal/uiadapter/eval/scorecard.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 478867f290bdf6a9 }
  - { id: scorecard_test, resource: internal/uiadapter/eval/scorecard_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: c10507f5ef6c8f7b }
  - { id: eval_test, resource: internal/uiadapter/eval_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 170e442f3eb11320 }
  - { id: prompt, resource: internal/uiadapter/prompt.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 07d009f8f445be65 }
---

# Files
- `internal/bmad/executor_adapter_test.go`
- `internal/uiadapter/backend/stubs.go`
- `internal/uiadapter/eval/corpus.go`
- `internal/uiadapter/eval/corpus_test.go`
- `internal/uiadapter/eval/scorecard.go`
- `internal/uiadapter/eval/scorecard_test.go`
- `internal/uiadapter/eval_test.go`
- `internal/uiadapter/prompt.go`

# Symbols
- delayAdapter (internal/bmad/executor_adapter_test.go:L34)
- .Translate() (internal/bmad/executor_adapter_test.go:L39)
- .Generate() (internal/uiadapter/backend/stubs.go:L39)
- .GenerateSingleShot() (internal/uiadapter/backend/stubs.go:L44)
- synthUIAST() (internal/uiadapter/backend/stubs.go:L77)
- Fixture (internal/uiadapter/eval/corpus.go:L21)
- Expected (internal/uiadapter/eval/corpus.go:L30)
- LoadCorpus() (internal/uiadapter/eval/corpus.go:L55)
- collectFixtureIDs() (internal/uiadapter/eval/corpus.go:L69)
- loadFixture() (internal/uiadapter/eval/corpus.go:L89)
- TestEval_Corpus_MinimumCount() (internal/uiadapter/eval/corpus_test.go:L13)
- .recordPreservation() (internal/uiadapter/eval/scorecard.go:L112)
- classifyGeneratedBy() (internal/uiadapter/eval/scorecard.go:L246)
- inferModel() (internal/uiadapter/eval/scorecard.go:L259)
- widgetTypesFrom() (internal/uiadapter/eval/scorecard.go:L267)
- renderedText() (internal/uiadapter/eval/scorecard.go:L280)
- Score() (internal/uiadapter/eval/scorecard.go:L53)
- .record() (internal/uiadapter/eval/scorecard.go:L73)
- scorecard_test.go (internal/uiadapter/eval/scorecard_test.go:L1)
- TestEval_PerWidgetPrecisionRecall() (internal/uiadapter/eval/scorecard_test.go:L104)
- TestEval_Scorecard_PrettyPrint() (internal/uiadapter/eval/scorecard_test.go:L153)
- TestEval_MeetsThresholds_Table() (internal/uiadapter/eval/scorecard_test.go:L178)
- sequentialMockAdapter (internal/uiadapter/eval/scorecard_test.go:L19)
- .Translate() (internal/uiadapter/eval/scorecard_test.go:L24)
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
- prompt.go (internal/uiadapter/prompt.go:L1)
- PromptVersion() (internal/uiadapter/prompt.go:L17)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [go_pkg_log_slog](/modules/go-pkg-log-slog.md)
- [time.Duration](/modules/time-duration.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
