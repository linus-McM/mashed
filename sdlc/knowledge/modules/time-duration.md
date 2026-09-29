---
type: Module
title: time.Duration
description: "Graphify community 310: app_review.go, internal/bmad/executor_adapter_test.go, internal/uiadapter/backend/lifecycle_test.go, internal/uiadapter/eval/scorecard.go, internal/uiadapter/eval/scorecard_v3."
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app_review, resource: app_review.go, last_modified: "2026-09-30T01:02:27+10:00", digest: 948be09757145e9c }
  - { id: executor_adapter_test, resource: internal/bmad/executor_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f79102c9f400d28f }
  - { id: lifecycle_test, resource: internal/uiadapter/backend/lifecycle_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 7f78d92ee9e83203 }
  - { id: scorecard, resource: internal/uiadapter/eval/scorecard.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 478867f290bdf6a9 }
  - { id: scorecard_v3, resource: internal/uiadapter/eval/scorecard_v3.go, last_modified: "2026-04-23T11:36:37+10:00", digest: 43f9103bcabd6620 }
---

# Files
- `app_review.go`
- `internal/bmad/executor_adapter_test.go`
- `internal/uiadapter/backend/lifecycle_test.go`
- `internal/uiadapter/eval/scorecard.go`
- `internal/uiadapter/eval/scorecard_v3.go`

# Symbols
- .StreamCodeReviewSummary() (app_review.go:L112)
- runClaudePrompt() (app_review.go:L471)
- isReviewableFile() (app_review.go:L84)
- delayAdapter (internal/bmad/executor_adapter_test.go:L34)
- .Translate() (internal/bmad/executor_adapter_test.go:L39)
- slowBackend (internal/uiadapter/backend/lifecycle_test.go:L16)
- .WarmUp() (internal/uiadapter/backend/lifecycle_test.go:L22)
- .ValidJSONRate() (internal/uiadapter/eval/scorecard.go:L131)
- .ValidatorPassRate() (internal/uiadapter/eval/scorecard.go:L138)
- .P95Latency() (internal/uiadapter/eval/scorecard.go:L149)
- .PerWidgetPrecision() (internal/uiadapter/eval/scorecard.go:L168)
- .PerWidgetRecall() (internal/uiadapter/eval/scorecard.go:L177)
- Scorecard (internal/uiadapter/eval/scorecard.go:L19)
- .MeetsThresholds() (internal/uiadapter/eval/scorecard.go:L190)
- .PrettyPrint() (internal/uiadapter/eval/scorecard.go:L207)
- sortedWidgetNames() (internal/uiadapter/eval/scorecard.go:L311)
- BackendThresholds (internal/uiadapter/eval/scorecard_v3.go:L20)

# Depends on
- [App](/modules/app-73.md)
- [app_review_test.go](/modules/app-review-test-go.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)

# Inferred
- [App](/modules/app-73.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
