---
type: Module
title: router_test.go
description: "Graphify community 475: internal/uiadapter/backend/router_test.go"
resource: internal/uiadapter/backend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: router_test, resource: internal/uiadapter/backend/router_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 0a276fac34b2f9ab }
---

# Files
- `internal/uiadapter/backend/router_test.go`

# Symbols
- router_test.go (internal/uiadapter/backend/router_test.go:L1)
- TestRouter_ResolveUnknownBackend() (internal/uiadapter/backend/router_test.go:L110)
- newTestRouter() (internal/uiadapter/backend/router_test.go:L14)
- TestRouter_CoversEveryPolicy() (internal/uiadapter/backend/router_test.go:L27)
- TestRouter_PrivacyStrictBlocksClaude() (internal/uiadapter/backend/router_test.go:L49)
- TestRouter_CostAwareChoosesOllamaWhenHealthy() (internal/uiadapter/backend/router_test.go:L71)
- TestRouter_CostAwareFallsToClaudeWhenOllamaDown() (internal/uiadapter/backend/router_test.go:L82)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [NewStub](/modules/newstub.md)
- [Router](/modules/router.md)

# Inferred
- [NewStub](/modules/newstub.md)

# Features
- no feature plan names these files
