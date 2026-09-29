---
type: Module
title: ModelInfo
description: "Graphify community 254: app_models.go, app_review_test.go, internal/domain/models.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app_models, resource: app_models.go, last_modified: "2026-04-10T11:36:27+10:00", digest: cf387264112e9ff8 }
  - { id: app_review_test, resource: app_review_test.go, last_modified: "2026-04-10T15:03:59+10:00", digest: cd3cd948419c464e }
  - { id: models, resource: internal/domain/models.go, last_modified: "2026-04-10T10:55:13+10:00", digest: 66a73fbeed36ede6 }
---

# Files
- `app_models.go`
- `app_review_test.go`
- `internal/domain/models.go`

# Symbols
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
- [claudeCommand](/modules/claudecommand.md)

# Features
- no feature plan names these files
