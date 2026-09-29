---
type: Module
title: Router
description: "Graphify community 8: internal/uiadapter/backend/backend.go, internal/uiadapter/backend/lifecycle.go, internal/uiadapter/backend/lifecycle_test.go, internal/uiadapter/backend/router.go, internal/uiada"
resource: internal/uiadapter/backend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: lifecycle, resource: internal/uiadapter/backend/lifecycle.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e57a307fa43599dd }
  - { id: lifecycle_test, resource: internal/uiadapter/backend/lifecycle_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 7f78d92ee9e83203 }
  - { id: router, resource: internal/uiadapter/backend/router.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 456b1b3b77e3a38e }
  - { id: router_test, resource: internal/uiadapter/backend/router_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 0a276fac34b2f9ab }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 60bb478e1146d7a5 }
---

# Files
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/lifecycle.go`
- `internal/uiadapter/backend/lifecycle_test.go`
- `internal/uiadapter/backend/router.go`
- `internal/uiadapter/backend/router_test.go`
- `internal/uiadapter/backend/stubs.go`

# Symbols
- LLMBackend (internal/uiadapter/backend/backend.go:L31)
- .healthProbeAll() (internal/uiadapter/backend/lifecycle.go:L115)
- Lifecycle (internal/uiadapter/backend/lifecycle.go:L15)
- NewLifecycle() (internal/uiadapter/backend/lifecycle.go:L29)
- .WarmUpAll() (internal/uiadapter/backend/lifecycle.go:L42)
- .WarmUpState() (internal/uiadapter/backend/lifecycle.go:L69)
- .StartHealthTicker() (internal/uiadapter/backend/lifecycle.go:L86)
- TestLifecycle_WarmUpConcurrentSafe() (internal/uiadapter/backend/lifecycle_test.go:L102)
- TestLifecycle_WarmUpAllInParallel() (internal/uiadapter/backend/lifecycle_test.go:L30)
- TestLifecycle_WarmUpRecordsResults() (internal/uiadapter/backend/lifecycle_test.go:L49)
- TestLifecycle_TickerDisableable() (internal/uiadapter/backend/lifecycle_test.go:L63)
- TestLifecycle_HealthTickerFeedsRouter() (internal/uiadapter/backend/lifecycle_test.go:L80)
- .decideCostAware() (internal/uiadapter/backend/router.go:L116)
- .matchesPrivacy() (internal/uiadapter/backend/router.go:L126)
- defaultPrivacyPatterns() (internal/uiadapter/backend/router.go:L136)
- .Resolve() (internal/uiadapter/backend/router.go:L152)
- .Dispatch() (internal/uiadapter/backend/router.go:L162)
- RouterPolicy (internal/uiadapter/backend/router.go:L17)
- Router (internal/uiadapter/backend/router.go:L30)
- NewRouter() (internal/uiadapter/backend/router.go:L42)
- .SetHealth() (internal/uiadapter/backend/router.go:L58)
- Decision (internal/uiadapter/backend/router.go:L70)
- .Decide() (internal/uiadapter/backend/router.go:L79)
- router_test.go (internal/uiadapter/backend/router_test.go:L1)
- TestRouter_ResolveUnknownBackend() (internal/uiadapter/backend/router_test.go:L110)
- newTestRouter() (internal/uiadapter/backend/router_test.go:L14)
- TestRouter_CoversEveryPolicy() (internal/uiadapter/backend/router_test.go:L27)
- TestRouter_PrivacyStrictBlocksClaude() (internal/uiadapter/backend/router_test.go:L49)
- TestRouter_DefaultPolicyFallsBackToLocalWhenClaudeAbsent() (internal/uiadapter/backend/router_test.go:L59)
- TestRouter_CostAwareChoosesOllamaWhenHealthy() (internal/uiadapter/backend/router_test.go:L71)
- TestRouter_CostAwareFallsToClaudeWhenOllamaDown() (internal/uiadapter/backend/router_test.go:L82)
- TestRouter_ConfigurablePrivacyPatterns() (internal/uiadapter/backend/router_test.go:L92)
- NewStub() (internal/uiadapter/backend/stubs.go:L28)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
