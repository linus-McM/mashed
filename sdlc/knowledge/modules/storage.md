---
type: Module
title: Storage
description: "Graphify community 236: internal/bmad/executor_interactive_test.go, internal/bmad/storage.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: storage, resource: internal/bmad/storage.go, last_modified: "2026-09-30T01:08:11+10:00", digest: 39b6c398a492fdb8 }
---

# Files
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/storage.go`

# Symbols
- saveDownstreamWorkflow() (internal/bmad/executor_interactive_test.go:L94)
- .ListWorkflowsByRepo() (internal/bmad/storage.go:L108)
- .DeleteWorkflow() (internal/bmad/storage.go:L124)
- .SaveAgent() (internal/bmad/storage.go:L141)
- .ListAgents() (internal/bmad/storage.go:L151)
- .DeleteAgent() (internal/bmad/storage.go:L180)
- Storage (internal/bmad/storage.go:L19)
- atomicWriteJSON() (internal/bmad/storage.go:L227)
- validateID() (internal/bmad/storage.go:L38)
- .SaveWorkflow() (internal/bmad/storage.go:L46)
- .LoadWorkflow() (internal/bmad/storage.go:L56)
- .ListWorkflows() (internal/bmad/storage.go:L78)

# Depends on
- [WriteFileAtomic](/modules/writefileatomic.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
