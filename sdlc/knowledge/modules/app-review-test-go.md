---
type: Module
title: app_review_test.go
description: "Graphify community 161: app_models.go, app_review.go, app_review_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app_models, resource: app_models.go, last_modified: "2026-04-10T11:36:27+10:00", digest: cf387264112e9ff8 }
  - { id: app_review, resource: app_review.go, last_modified: "2026-05-07T18:18:02+10:00", digest: f186f322bd66e914 }
  - { id: app_review_test, resource: app_review_test.go, last_modified: "2026-04-10T15:03:59+10:00", digest: cd3cd948419c464e }
---

# Files
- `app_models.go`
- `app_review.go`
- `app_review_test.go`

# Symbols
- parseModelResponse() (app_models.go:L76)
- .SpawnRefactorPlan() (app_review.go:L326)
- refactorPlanFilename() (app_review.go:L374)
- slugifyPlanPath() (app_review.go:L410)
- truncateDiffLines() (app_review.go:L442)
- app_review_test.go (app_review_test.go:L1)
- TestReviewConcurrencyGuard_DifferentRepos() (app_review_test.go:L123)
- TestTruncateDiffLines() (app_review_test.go:L151)
- TestReviewSummary_StructConstruction() (app_review_test.go:L21)
- TestTruncateDiffLines_PreservesContent() (app_review_test.go:L212)
- TestFileSummarySystemPrompt() (app_review_test.go:L232)
- TestReviewConstants() (app_review_test.go:L242)
- TestParseModelResponse_ValidJSON() (app_review_test.go:L294)
- TestParseModelResponse_EmptyInput() (app_review_test.go:L307)
- TestParseModelResponse_InvalidJSON() (app_review_test.go:L312)
- TestParseModelResponse_JSONInCodeFence() (app_review_test.go:L317)
- TestParseModelResponse_JSONWithSurroundingProse() (app_review_test.go:L327)
- TestSpawnRefactorPlan_InputValidation() (app_review_test.go:L340)
- TestSpawnRefactorPlan_PlanPathFormat() (app_review_test.go:L371)
- TestRefactorPlanFilename() (app_review_test.go:L387)
- TestSlugifyPlanPath() (app_review_test.go:L478)
- TestFileSummary_BinaryFlag() (app_review_test.go:L74)
- TestReviewConcurrencyGuard() (app_review_test.go:L92)

# Depends on
- [ModelInfo](/modules/modelinfo.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
