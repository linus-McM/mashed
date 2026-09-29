---
type: Module
title: Router
description: "Graphify community 82: internal/uiadapter/backend/backend.go, internal/uiadapter/backend/lifecycle.go, internal/uiadapter/backend/router.go, internal/uiadapter/backend/router_test.go"
resource: internal/uiadapter/backend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: lifecycle, resource: internal/uiadapter/backend/lifecycle.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e57a307fa43599dd }
  - { id: router, resource: internal/uiadapter/backend/router.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 456b1b3b77e3a38e }
  - { id: router_test, resource: internal/uiadapter/backend/router_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 0a276fac34b2f9ab }
---

# Files
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/lifecycle.go`
- `internal/uiadapter/backend/router.go`
- `internal/uiadapter/backend/router_test.go`

# Symbols
- LLMBackend (internal/uiadapter/backend/backend.go:L31)
- .healthProbeAll() (internal/uiadapter/backend/lifecycle.go:L115)
- Lifecycle (internal/uiadapter/backend/lifecycle.go:L15)
- .WarmUpAll() (internal/uiadapter/backend/lifecycle.go:L42)
- .WarmUpState() (internal/uiadapter/backend/lifecycle.go:L69)
- .StartHealthTicker() (internal/uiadapter/backend/lifecycle.go:L86)
- router.go (internal/uiadapter/backend/router.go:L1)
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
- TestRouter_CostAwareChoosesOllamaWhenHealthy() (internal/uiadapter/backend/router_test.go:L71)
- TestRouter_CostAwareFallsToClaudeWhenOllamaDown() (internal/uiadapter/backend/router_test.go:L82)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)

# Features
- no feature plan names these files
