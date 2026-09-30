---
type: Module
title: scorecard.go
description: "Graphify community 38: internal/uiadapter/eval/corpus.go, internal/uiadapter/eval/corpus_test.go, internal/uiadapter/eval/scorecard.go"
resource: internal/uiadapter/eval
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: corpus, resource: internal/uiadapter/eval/corpus.go, last_modified: "2026-04-22T14:13:03+10:00", digest: be80b9817e3d70f3 }
  - { id: corpus_test, resource: internal/uiadapter/eval/corpus_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 2df8e716d6d0fee2 }
  - { id: scorecard, resource: internal/uiadapter/eval/scorecard.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 478867f290bdf6a9 }
---

# Files
- `internal/uiadapter/eval/corpus.go`
- `internal/uiadapter/eval/corpus_test.go`
- `internal/uiadapter/eval/scorecard.go`

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
- .record() (internal/uiadapter/eval/scorecard.go:L73)
- .recordWidgets() (internal/uiadapter/eval/scorecard.go:L95)

# Depends on
- [scorecard_test.go](/modules/scorecard-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
