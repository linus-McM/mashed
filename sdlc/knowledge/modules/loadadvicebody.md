---
type: Module
title: LoadAdviceBody
description: "Graphify community 466: app_review.go, internal/advice/loader.go, internal/advice/loader_test.go, internal/advice/types.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app_review, resource: app_review.go, last_modified: "2026-09-30T01:02:27+10:00", digest: 948be09757145e9c }
  - { id: loader, resource: internal/advice/loader.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 8dbc36587dafa6a5 }
  - { id: loader_test, resource: internal/advice/loader_test.go, last_modified: "2026-04-10T10:10:32+10:00", digest: 99b099c62c602a3e }
  - { id: types, resource: internal/advice/types.go, last_modified: "2026-04-10T10:10:32+10:00", digest: 91b5fd3d9c0081b2 }
---

# Files
- `app_review.go`
- `internal/advice/loader.go`
- `internal/advice/loader_test.go`
- `internal/advice/types.go`

# Symbols
- App (app_review.go:L105)
- .ListAdviceModes() (app_review.go:L105)
- .StreamAdvice() (app_review.go:L243)
- .SpawnRefactorPlan() (app_review.go:L338)
- LoadAdviceBody() (internal/advice/loader.go:L127)
- LoadAdviceModes() (internal/advice/loader.go:L28)
- globalAdviceDir() (internal/advice/loader.go:L304)
- localAdviceDir() (internal/advice/loader.go:L313)
- TestLoadAdviceModes_EmptyRepoPathSkipsLocal() (internal/advice/loader_test.go:L327)
- TestLoadAdviceBody_PublicAPI() (internal/advice/loader_test.go:L622)
- TestGlobalAdviceDir() (internal/advice/loader_test.go:L645)
- TestLocalAdviceDir() (internal/advice/loader_test.go:L652)
- advice/types.go (internal/advice/types.go:L1)
- AdviceMode (internal/advice/types.go:L4)

# Depends on
- [app_review_test.go](/modules/app-review-test-go.md)
- [loader_test.go](/modules/loader-test-go.md)
- [ModelInfo](/modules/modelinfo.md)
- [time.Duration](/modules/time-duration.md)

# Inferred
- [App](/modules/app-73.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
