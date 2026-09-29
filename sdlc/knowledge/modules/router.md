---
type: Module
title: Router
description: "Graphify community 231: internal/uiadapter/backend/backend.go, internal/uiadapter/backend/lifecycle.go, internal/uiadapter/backend/router.go"
resource: internal/uiadapter/backend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-09-29T07:07:25Z", digest: d0eac5646b5c1ef3 }
  - { id: lifecycle, resource: internal/uiadapter/backend/lifecycle.go, last_modified: "2026-09-29T07:07:25Z", digest: e57a307fa43599dd }
  - { id: router, resource: internal/uiadapter/backend/router.go, last_modified: "2026-09-29T07:07:25Z", digest: 456b1b3b77e3a38e }
---

# Files
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/lifecycle.go`
- `internal/uiadapter/backend/router.go`

# Symbols
- LLMBackend (internal/uiadapter/backend/backend.go:L31)
- .healthProbeAll() (internal/uiadapter/backend/lifecycle.go:L115)
- Lifecycle (internal/uiadapter/backend/lifecycle.go:L15)
- .WarmUpAll() (internal/uiadapter/backend/lifecycle.go:L42)
- .WarmUpState() (internal/uiadapter/backend/lifecycle.go:L69)
- .StartHealthTicker() (internal/uiadapter/backend/lifecycle.go:L86)
- .decideCostAware() (internal/uiadapter/backend/router.go:L116)
- .matchesPrivacy() (internal/uiadapter/backend/router.go:L126)
- .Resolve() (internal/uiadapter/backend/router.go:L152)
- .Dispatch() (internal/uiadapter/backend/router.go:L162)
- RouterPolicy (internal/uiadapter/backend/router.go:L17)
- Router (internal/uiadapter/backend/router.go:L30)
- .SetHealth() (internal/uiadapter/backend/router.go:L58)
- Decision (internal/uiadapter/backend/router.go:L70)
- .Decide() (internal/uiadapter/backend/router.go:L79)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
