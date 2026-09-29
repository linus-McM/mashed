---
type: Module
title: bmad/types.go
description: "Graphify community 65: app_bmad.go, internal/bmad/executor.go, internal/bmad/types.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
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
- .ListBmadAgents() (app_bmad.go:L308)
- .SaveBmadAgent() (app_bmad.go:L316)
- .GetControlFlowNodes() (app_bmad.go:L468)
- ExecStatusEvent (internal/bmad/executor.go:L48)
- bmad/types.go (internal/bmad/types.go:L1)
- InputSource (internal/bmad/types.go:L138)
- InputShape (internal/bmad/types.go:L149)
- InputSpec (internal/bmad/types.go:L161)
- OutputTarget (internal/bmad/types.go:L178)
- OutputSpec (internal/bmad/types.go:L187)
- GateKind (internal/bmad/types.go:L213)
- IterationGate (internal/bmad/types.go:L223)
- NodeType (internal/bmad/types.go:L232)
- MultiFileEntry (internal/bmad/types.go:L250)
- WorkflowNode (internal/bmad/types.go:L256)
- .EffectiveType() (internal/bmad/types.go:L295)
- Position (internal/bmad/types.go:L303)
- WorkflowExecStatus (internal/bmad/types.go:L332)
- WorkflowExecution (internal/bmad/types.go:L343)
- PendingPrompt (internal/bmad/types.go:L364)
- NodeInputEntry (internal/bmad/types.go:L392)
- BmadAgentConfig (internal/bmad/types.go:L404)
- AgentInfo (internal/bmad/types.go:L415)
- GroupedAgents (internal/bmad/types.go:L421)
- ArtifactType (internal/bmad/types.go:L428)
- ArtifactSpec (internal/bmad/types.go:L438)
- NodeArtifactEvent (internal/bmad/types.go:L448)
- ControlFlowNodeDef (internal/bmad/types.go:L460)
- BmadAgentRole (internal/bmad/types.go:L69)

# Depends on
- [App](/modules/app-76.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [ProcessDef](/modules/processdef.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
