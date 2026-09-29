---
type: Module
title: NewStub
description: "Graphify community 328: internal/uiadapter/backend/lifecycle.go, internal/uiadapter/backend/lifecycle_test.go, internal/uiadapter/backend/router.go, internal/uiadapter/backend/router_test.go, internal"
resource: internal/uiadapter/backend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: lifecycle, resource: internal/uiadapter/backend/lifecycle.go, last_modified: "2026-09-29T07:07:25Z", digest: e57a307fa43599dd }
  - { id: lifecycle_test, resource: internal/uiadapter/backend/lifecycle_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 7f78d92ee9e83203 }
  - { id: router, resource: internal/uiadapter/backend/router.go, last_modified: "2026-09-29T07:07:25Z", digest: 456b1b3b77e3a38e }
  - { id: router_test, resource: internal/uiadapter/backend/router_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 0a276fac34b2f9ab }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-09-29T07:07:25Z", digest: 60bb478e1146d7a5 }
---

# Files
- `internal/uiadapter/backend/lifecycle.go`
- `internal/uiadapter/backend/lifecycle_test.go`
- `internal/uiadapter/backend/router.go`
- `internal/uiadapter/backend/router_test.go`
- `internal/uiadapter/backend/stubs.go`

# Symbols
- NewLifecycle() (internal/uiadapter/backend/lifecycle.go:L29)
- TestLifecycle_WarmUpConcurrentSafe() (internal/uiadapter/backend/lifecycle_test.go:L102)
- TestLifecycle_WarmUpAllInParallel() (internal/uiadapter/backend/lifecycle_test.go:L30)
- TestLifecycle_WarmUpRecordsResults() (internal/uiadapter/backend/lifecycle_test.go:L49)
- TestLifecycle_TickerDisableable() (internal/uiadapter/backend/lifecycle_test.go:L63)
- TestLifecycle_HealthTickerFeedsRouter() (internal/uiadapter/backend/lifecycle_test.go:L80)
- defaultPrivacyPatterns() (internal/uiadapter/backend/router.go:L136)
- NewRouter() (internal/uiadapter/backend/router.go:L42)
- TestRouter_DefaultPolicyFallsBackToLocalWhenClaudeAbsent() (internal/uiadapter/backend/router_test.go:L59)
- TestRouter_ConfigurablePrivacyPatterns() (internal/uiadapter/backend/router_test.go:L92)
- NewStub() (internal/uiadapter/backend/stubs.go:L28)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [Router](/modules/router.md)
- [StubBackend](/modules/stubbackend.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
