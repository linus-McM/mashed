---
type: Module
title: time.Duration
description: "Graphify community 90: app_review.go, internal/uiadapter/eval/scorecard.go, internal/uiadapter/eval/scorecard_v3.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: app_review, resource: app_review.go, last_modified: "2026-05-07T18:18:02+10:00", digest: f186f322bd66e914 }
  - { id: scorecard, resource: internal/uiadapter/eval/scorecard.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 478867f290bdf6a9 }
  - { id: scorecard_v3, resource: internal/uiadapter/eval/scorecard_v3.go, last_modified: "2026-04-23T11:36:37+10:00", digest: 43f9103bcabd6620 }
---

# Files
- `app_review.go`
- `internal/uiadapter/eval/scorecard.go`
- `internal/uiadapter/eval/scorecard_v3.go`

# Symbols
- .StreamCodeReviewSummary() (app_review.go:L112)
- runClaudePrompt() (app_review.go:L459)
- isReviewableFile() (app_review.go:L84)
- .ValidJSONRate() (internal/uiadapter/eval/scorecard.go:L131)
- .ValidatorPassRate() (internal/uiadapter/eval/scorecard.go:L138)
- .P95Latency() (internal/uiadapter/eval/scorecard.go:L149)
- .PerWidgetPrecision() (internal/uiadapter/eval/scorecard.go:L168)
- .PerWidgetRecall() (internal/uiadapter/eval/scorecard.go:L177)
- Scorecard (internal/uiadapter/eval/scorecard.go:L19)
- .MeetsThresholds() (internal/uiadapter/eval/scorecard.go:L190)
- .PrettyPrint() (internal/uiadapter/eval/scorecard.go:L207)
- toSet() (internal/uiadapter/eval/scorecard.go:L300)
- sortedWidgetNames() (internal/uiadapter/eval/scorecard.go:L311)
- .recordWidgets() (internal/uiadapter/eval/scorecard.go:L95)
- BackendThresholds (internal/uiadapter/eval/scorecard_v3.go:L20)

# Depends on
- [app_review_test.go](/modules/app-review-test-go.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [ScopedDiff](/modules/scopeddiff.md)

# Inferred
- [App](/modules/app-63.md)

# Features
- no feature plan names these files
