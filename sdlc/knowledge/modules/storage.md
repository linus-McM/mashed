---
type: Module
title: Storage
description: "Graphify community 264: internal/bmad/executor_interactive_test.go, internal/bmad/storage.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: storage, resource: internal/bmad/storage.go, last_modified: "2026-04-10T12:50:28+10:00", digest: 360ebf80068d1480 }
---

# Files
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/storage.go`

# Symbols
- saveDownstreamWorkflow() (internal/bmad/executor_interactive_test.go:L94)
- .ListWorkflowsByRepo() (internal/bmad/storage.go:L106)
- .DeleteWorkflow() (internal/bmad/storage.go:L122)
- .SaveAgent() (internal/bmad/storage.go:L139)
- .ListAgents() (internal/bmad/storage.go:L149)
- Storage (internal/bmad/storage.go:L17)
- .DeleteAgent() (internal/bmad/storage.go:L178)
- atomicWriteJSON() (internal/bmad/storage.go:L225)
- validateID() (internal/bmad/storage.go:L36)
- .SaveWorkflow() (internal/bmad/storage.go:L44)
- .LoadWorkflow() (internal/bmad/storage.go:L54)
- .ListWorkflows() (internal/bmad/storage.go:L76)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
