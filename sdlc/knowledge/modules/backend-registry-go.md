---
type: Module
title: backend/registry.go
description: "Graphify community 53: internal/uiadapter/backend/claudeapi/stub.go, internal/uiadapter/backend/claudecli/stub.go, internal/uiadapter/backend/ollama/stub.go, internal/uiadapter/backend/registry.go, in"
resource: internal/uiadapter/backend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: stub, resource: internal/uiadapter/backend/claudeapi/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 0c95368a8fcd18d3 }
  - { id: stub, resource: internal/uiadapter/backend/claudecli/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 9d9685929a699a14 }
  - { id: stub, resource: internal/uiadapter/backend/ollama/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 82a69a451dac2ce8 }
  - { id: registry, resource: internal/uiadapter/backend/registry.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 34796db160eeb018 }
  - { id: registry_test, resource: internal/uiadapter/backend/registry_test.go, last_modified: "2026-04-23T11:09:52+10:00", digest: aa5a14fdab76f8f6 }
---

# Files
- `internal/uiadapter/backend/claudeapi/stub.go`
- `internal/uiadapter/backend/claudecli/stub.go`
- `internal/uiadapter/backend/ollama/stub.go`
- `internal/uiadapter/backend/registry.go`
- `internal/uiadapter/backend/registry_test.go`

# Symbols
- claudeapi/stub.go (internal/uiadapter/backend/claudeapi/stub.go:L1)
- init() (internal/uiadapter/backend/claudeapi/stub.go:L13)
- claudecli/stub.go (internal/uiadapter/backend/claudecli/stub.go:L1)
- init() (internal/uiadapter/backend/claudecli/stub.go:L12)
- ollama/stub.go (internal/uiadapter/backend/ollama/stub.go:L1)
- init() (internal/uiadapter/backend/ollama/stub.go:L13)
- backend/registry.go (internal/uiadapter/backend/registry.go:L1)
- Constructor (internal/uiadapter/backend/registry.go:L13)
- Register() (internal/uiadapter/backend/registry.go:L24)
- Available() (internal/uiadapter/backend/registry.go:L53)
- reset() (internal/uiadapter/backend/registry.go:L67)
- TestBackend_Available() (internal/uiadapter/backend/registry_test.go:L132)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [fastpath.go](/modules/fastpath-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
