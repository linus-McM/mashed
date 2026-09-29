---
type: Module
title: ModelInfo
description: "Graphify community 262: app_git.go, app_models.go, app_review_test.go, internal/domain/models.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app_git, resource: app_git.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 6abf9d08d507b1db }
  - { id: app_models, resource: app_models.go, last_modified: "2026-04-10T11:36:27+10:00", digest: cf387264112e9ff8 }
  - { id: app_review_test, resource: app_review_test.go, last_modified: "2026-04-10T15:03:59+10:00", digest: cd3cd948419c464e }
  - { id: models, resource: internal/domain/models.go, last_modified: "2026-04-10T10:55:13+10:00", digest: 66a73fbeed36ede6 }
---

# Files
- `app_git.go`
- `app_models.go`
- `app_review_test.go`
- `internal/domain/models.go`

# Symbols
- .SpawnPRReview() (app_git.go:L876)
- App (app_models.go:L102)
- .initModelCache() (app_models.go:L102)
- .ListModels() (app_models.go:L109)
- fetchModelsFromCLI() (app_models.go:L32)
- modelsResponse (app_models.go:L70)
- TestModelRegistry_FallbackHasDefault() (app_review_test.go:L251)
- TestModelRegistry_AliasLookup() (app_review_test.go:L268)
- TestModelRegistry_UnknownAliasFallsBack() (app_review_test.go:L288)
- models.go (internal/domain/models.go:L1)
- FallbackModels() (internal/domain/models.go:L15)
- DefaultAlias() (internal/domain/models.go:L24)
- ModelByAlias() (internal/domain/models.go:L37)
- ModelInfo (internal/domain/models.go:L4)

# Depends on
- [app_review_test.go](/modules/app-review-test-go.md)

# Inferred
- [App](/modules/app-73.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
