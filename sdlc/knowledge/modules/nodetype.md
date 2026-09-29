---
type: Module
title: NodeType
description: "Graphify community 240: app_bmad.go, internal/bmad/executor.go, internal/bmad/types.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: types, resource: internal/bmad/types.go, last_modified: "2026-04-28T09:29:49+10:00", digest: 05a3e7d01f5f08c2 }
---

# Files
- `app_bmad.go`
- `internal/bmad/executor.go`
- `internal/bmad/types.go`

# Symbols
- .GetBmadExecution() (app_bmad.go:L205)
- .GetControlFlowNodes() (app_bmad.go:L468)
- ExecStatusEvent (internal/bmad/executor.go:L48)
- NodeType (internal/bmad/types.go:L232)
- WorkflowNode (internal/bmad/types.go:L256)
- .EffectiveType() (internal/bmad/types.go:L295)
- WorkflowExecStatus (internal/bmad/types.go:L332)
- WorkflowExecution (internal/bmad/types.go:L343)
- ControlFlowNodeDef (internal/bmad/types.go:L460)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
