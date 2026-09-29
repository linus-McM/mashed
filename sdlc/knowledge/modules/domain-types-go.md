---
type: Module
title: domain/types.go
description: "Graphify community 75: app.go, internal/agent/engine.go, internal/agent/engine_test.go, internal/domain/types.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: engine, resource: internal/agent/engine.go, last_modified: "2026-04-11T17:08:28+10:00", digest: 1042190db571e4e7 }
  - { id: engine_test, resource: internal/agent/engine_test.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 745feda4505cf50c }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
---

# Files
- `app.go`
- `internal/agent/engine.go`
- `internal/agent/engine_test.go`
- `internal/domain/types.go`

# Symbols
- .GetNotifications() (app.go:L743)
- .RemoveAgent() (internal/agent/engine.go:L171)
- .ActiveAgentCount() (internal/agent/engine.go:L178)
- .GetAgentStatus() (internal/agent/engine.go:L185)
- priorityFor() (internal/agent/engine.go:L197)
- statusToEventType() (internal/agent/engine.go:L218)
- buildSummary() (internal/agent/engine.go:L241)
- lastActivity() (internal/agent/engine.go:L266)
- agentState (internal/agent/engine.go:L41)
- NotificationEngine (internal/agent/engine.go:L50)
- .Events() (internal/agent/engine.go:L69)
- .ProcessAgentUpdate() (internal/agent/engine.go:L75)
- TestPriorityOrdering() (internal/agent/engine_test.go:L145)
- domain/types.go (internal/domain/types.go:L1)
- AgentStatus (internal/domain/types.go:L10)
- SessionType (internal/domain/types.go:L117)
- TerminalSession (internal/domain/types.go:L125)
- EventType (internal/domain/types.go:L157)
- NotificationEvent (internal/domain/types.go:L168)
- AgentProvider (internal/domain/types.go:L221)
- LogKind (internal/domain/types.go:L24)
- LogLine (internal/domain/types.go:L36)
- Agent (internal/domain/types.go:L43)
- DagEdge (internal/domain/types.go:L61)
- Workflow (internal/domain/types.go:L67)
- Repo (internal/domain/types.go:L77)

# Depends on
- [AssetWatcher](/modules/assetwatcher.md)
- [diff.go](/modules/diff-go.md)
- [RepoScanner](/modules/reposcanner.md)
- [sessions.go](/modules/sessions-go.md)
- [worktree.go](/modules/worktree-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
