---
type: Module
title: backend/registry_test.go
description: "Graphify community 395: internal/uiadapter/backend/registry.go, internal/uiadapter/backend/registry_test.go"
resource: internal/uiadapter/backend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: registry, resource: internal/uiadapter/backend/registry.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 34796db160eeb018 }
  - { id: registry_test, resource: internal/uiadapter/backend/registry_test.go, last_modified: "2026-04-23T11:09:52+10:00", digest: aa5a14fdab76f8f6 }
---

# Files
- `internal/uiadapter/backend/registry.go`
- `internal/uiadapter/backend/registry_test.go`

# Symbols
- From() (internal/uiadapter/backend/registry.go:L41)
- Available() (internal/uiadapter/backend/registry.go:L53)
- backend/registry_test.go (internal/uiadapter/backend/registry_test.go:L1)
- TestBackend_SingleShotUnsupportedIsSentinel() (internal/uiadapter/backend/registry_test.go:L119)
- TestBackend_Available() (internal/uiadapter/backend/registry_test.go:L132)
- TestBackend_FromUnknownName() (internal/uiadapter/backend/registry_test.go:L22)
- TestBackend_FromKnownName() (internal/uiadapter/backend/registry_test.go:L37)
- TestBackend_InterfaceStressConcurrent() (internal/uiadapter/backend/registry_test.go:L53)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [fastpath.go](/modules/fastpath-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
