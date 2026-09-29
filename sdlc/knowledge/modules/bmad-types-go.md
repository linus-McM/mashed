---
type: Module
title: bmad/types.go
description: "Graphify community 48: app_bmad.go, internal/bmad/storage.go, internal/bmad/types.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: storage, resource: internal/bmad/storage.go, last_modified: "2026-09-30T01:08:11+10:00", digest: 39b6c398a492fdb8 }
  - { id: types, resource: internal/bmad/types.go, last_modified: "2026-04-28T09:29:49+10:00", digest: 05a3e7d01f5f08c2 }
---

# Files
- `app_bmad.go`
- `internal/bmad/storage.go`
- `internal/bmad/types.go`

# Symbols
- .ListBmadAgents() (app_bmad.go:L308)
- .SaveBmadAgent() (app_bmad.go:L316)
- .ListAllAgents() (app_bmad.go:L333)
- ListClaudeAgents() (internal/bmad/storage.go:L199)
- bmad/types.go (internal/bmad/types.go:L1)
- InputSource (internal/bmad/types.go:L138)
- InputShape (internal/bmad/types.go:L149)
- InputSpec (internal/bmad/types.go:L161)
- OutputTarget (internal/bmad/types.go:L178)
- OutputSpec (internal/bmad/types.go:L187)
- GateKind (internal/bmad/types.go:L213)
- IterationGate (internal/bmad/types.go:L223)
- MultiFileEntry (internal/bmad/types.go:L250)
- Position (internal/bmad/types.go:L303)
- PendingPrompt (internal/bmad/types.go:L364)
- NodeInputEntry (internal/bmad/types.go:L392)
- BmadAgentConfig (internal/bmad/types.go:L404)
- AgentInfo (internal/bmad/types.go:L415)
- GroupedAgents (internal/bmad/types.go:L421)
- ArtifactType (internal/bmad/types.go:L428)
- ArtifactSpec (internal/bmad/types.go:L438)
- NodeArtifactEvent (internal/bmad/types.go:L448)
- BmadAgentRole (internal/bmad/types.go:L69)

# Depends on
- [App](/modules/app.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [NodeType](/modules/nodetype.md)
- [ProcessDef](/modules/processdef.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
