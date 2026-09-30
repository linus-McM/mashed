---
type: Module
title: refactorPlanFilename
description: "Graphify community 466: app_review.go, app_review_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app_review, resource: app_review.go, last_modified: "2026-05-07T18:18:02+10:00", digest: f186f322bd66e914 }
  - { id: app_review_test, resource: app_review_test.go, last_modified: "2026-04-10T15:03:59+10:00", digest: cd3cd948419c464e }
---

# Files
- `app_review.go`
- `app_review_test.go`

# Symbols
- .SpawnRefactorPlan() (app_review.go:L326)
- refactorPlanFilename() (app_review.go:L374)
- slugifyPlanPath() (app_review.go:L410)
- TestRefactorPlanFilename() (app_review_test.go:L387)
- TestSlugifyPlanPath() (app_review_test.go:L478)

# Depends on
- [ModelInfo](/modules/modelinfo.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
