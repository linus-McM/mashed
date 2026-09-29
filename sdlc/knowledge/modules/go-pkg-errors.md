---
type: Module
title: go_pkg_errors
description: "Graphify community 53: app_uiadapter_v3.go, internal/uiadapter/backend/backend.go, internal/uiadapter/backend/claudeapi/stub.go, internal/uiadapter/backend/claudecli/client.go, internal/uiadapter/back"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app_uiadapter_v3, resource: app_uiadapter_v3.go, last_modified: "2026-04-23T11:43:31+10:00", digest: c6ede9f0c3335d2e }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: stub, resource: internal/uiadapter/backend/claudeapi/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 0c95368a8fcd18d3 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: stub, resource: internal/uiadapter/backend/claudecli/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 9d9685929a699a14 }
  - { id: lifecycle, resource: internal/uiadapter/backend/lifecycle.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e57a307fa43599dd }
  - { id: stub, resource: internal/uiadapter/backend/ollama/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 82a69a451dac2ce8 }
  - { id: registry, resource: internal/uiadapter/backend/registry.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 34796db160eeb018 }
  - { id: registry_test, resource: internal/uiadapter/backend/registry_test.go, last_modified: "2026-04-23T11:09:52+10:00", digest: aa5a14fdab76f8f6 }
  - { id: router, resource: internal/uiadapter/backend/router.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 456b1b3b77e3a38e }
  - { id: breaker, resource: internal/uiadapter/breaker.go, last_modified: "2026-04-26T10:14:41+10:00", digest: ce7fb713c254dda5 }
---

# Files
- `app_uiadapter_v3.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/stub.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/claudecli/stub.go`
- `internal/uiadapter/backend/lifecycle.go`
- `internal/uiadapter/backend/ollama/stub.go`
- `internal/uiadapter/backend/registry.go`
- `internal/uiadapter/backend/registry_test.go`
- `internal/uiadapter/backend/router.go`
- `internal/uiadapter/breaker.go`

# Symbols
- app_uiadapter_v3.go (app_uiadapter_v3.go:L1)
- backend.go (internal/uiadapter/backend/backend.go:L1)
- claudeapi/stub.go (internal/uiadapter/backend/claudeapi/stub.go:L1)
- init() (internal/uiadapter/backend/claudeapi/stub.go:L13)
- claudecli/client.go (internal/uiadapter/backend/claudecli/client.go:L1)
- init() (internal/uiadapter/backend/claudecli/client.go:L351)
- claudecli/stub.go (internal/uiadapter/backend/claudecli/stub.go:L1)
- init() (internal/uiadapter/backend/claudecli/stub.go:L12)
- lifecycle.go (internal/uiadapter/backend/lifecycle.go:L1)
- ollama/stub.go (internal/uiadapter/backend/ollama/stub.go:L1)
- init() (internal/uiadapter/backend/ollama/stub.go:L13)
- backend/registry.go (internal/uiadapter/backend/registry.go:L1)
- Constructor (internal/uiadapter/backend/registry.go:L13)
- Register() (internal/uiadapter/backend/registry.go:L24)
- From() (internal/uiadapter/backend/registry.go:L41)
- Available() (internal/uiadapter/backend/registry.go:L53)
- reset() (internal/uiadapter/backend/registry.go:L67)
- backend/registry_test.go (internal/uiadapter/backend/registry_test.go:L1)
- TestBackend_SingleShotUnsupportedIsSentinel() (internal/uiadapter/backend/registry_test.go:L119)
- TestBackend_Available() (internal/uiadapter/backend/registry_test.go:L132)
- TestBackend_FromUnknownName() (internal/uiadapter/backend/registry_test.go:L22)
- TestBackend_FromKnownName() (internal/uiadapter/backend/registry_test.go:L37)
- TestBackend_InterfaceStressConcurrent() (internal/uiadapter/backend/registry_test.go:L53)
- router.go (internal/uiadapter/backend/router.go:L1)
- breaker.go (internal/uiadapter/breaker.go:L1)

# Depends on
- [BreakerSet](/modules/breakerset.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [Router](/modules/router.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
