---
type: Module
title: domain/types.go
description: "Graphify community 164: app_spawn.go, internal/agent/engine.go, internal/agent/engine_test.go, internal/domain/types.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app_spawn, resource: app_spawn.go, last_modified: "2026-05-07T18:18:02+10:00", digest: cbe44d5e06097d50 }
  - { id: engine, resource: internal/agent/engine.go, last_modified: "2026-04-11T17:08:28+10:00", digest: 1042190db571e4e7 }
  - { id: engine_test, resource: internal/agent/engine_test.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 745feda4505cf50c }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
---

# Files
- `app_spawn.go`
- `internal/agent/engine.go`
- `internal/agent/engine_test.go`
- `internal/domain/types.go`

# Symbols
- .GetAgentLog() (app_spawn.go:L79)
- .GetAgentStatus() (internal/agent/engine.go:L185)
- priorityFor() (internal/agent/engine.go:L197)
- statusToEventType() (internal/agent/engine.go:L218)
- buildSummary() (internal/agent/engine.go:L241)
- lastActivity() (internal/agent/engine.go:L266)
- agentState (internal/agent/engine.go:L41)
- .ProcessAgentUpdate() (internal/agent/engine.go:L75)
- TestPriorityOrdering() (internal/agent/engine_test.go:L145)
- domain/types.go (internal/domain/types.go:L1)
- AgentStatus (internal/domain/types.go:L10)
- EventType (internal/domain/types.go:L157)
- AgentProvider (internal/domain/types.go:L221)
- LogKind (internal/domain/types.go:L24)
- LogLine (internal/domain/types.go:L36)
- Agent (internal/domain/types.go:L43)
- DagEdge (internal/domain/types.go:L61)
- Workflow (internal/domain/types.go:L67)
- Repo (internal/domain/types.go:L77)

# Depends on
- [App](/modules/app-355.md)
- [App](/modules/app-73.md)
- [AssetWatcher](/modules/assetwatcher.md)
- [NotificationEngine](/modules/notificationengine.md)
- [SessionData](/modules/sessiondata.md)
- [sessions.go](/modules/sessions-go.md)
- [time.Time](/modules/time-time.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
