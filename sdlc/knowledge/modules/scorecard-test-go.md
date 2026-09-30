---
type: Module
title: scorecard_test.go
description: "Graphify community 191: internal/uiadapter/eval/scorecard.go, internal/uiadapter/eval/scorecard_test.go, internal/uiadapter/eval_test.go, internal/uiadapter/prompt.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: scorecard, resource: internal/uiadapter/eval/scorecard.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 478867f290bdf6a9 }
  - { id: scorecard_test, resource: internal/uiadapter/eval/scorecard_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: c10507f5ef6c8f7b }
  - { id: eval_test, resource: internal/uiadapter/eval_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 170e442f3eb11320 }
  - { id: prompt, resource: internal/uiadapter/prompt.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 07d009f8f445be65 }
---

# Files
- `internal/uiadapter/eval/scorecard.go`
- `internal/uiadapter/eval/scorecard_test.go`
- `internal/uiadapter/eval_test.go`
- `internal/uiadapter/prompt.go`

# Symbols
- Score() (internal/uiadapter/eval/scorecard.go:L53)
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
- prompt.go (internal/uiadapter/prompt.go:L1)
- PromptVersion() (internal/uiadapter/prompt.go:L17)

# Depends on
- [context.Context](/modules/context-context.md)
- [scorecard.go](/modules/scorecard-go.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
