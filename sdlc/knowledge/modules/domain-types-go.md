---
type: Module
title: domain/types.go
description: "Graphify community 75: internal/agent/engine.go, internal/agent/engine_test.go, internal/domain/types.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: engine, resource: internal/agent/engine.go, last_modified: "2026-04-11T17:08:28+10:00", digest: 1042190db571e4e7 }
  - { id: engine_test, resource: internal/agent/engine_test.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 745feda4505cf50c }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
---

# Files
- `internal/agent/engine.go`
- `internal/agent/engine_test.go`
- `internal/domain/types.go`

# Symbols
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
- [App](/modules/app-349.md)
- [AssetWatcher](/modules/assetwatcher.md)
- [diff.go](/modules/diff-go.md)
- [NotificationEngine](/modules/notificationengine.md)
- [RepoScanner](/modules/reposcanner.md)
- [SessionData](/modules/sessiondata.md)
- [sessions.go](/modules/sessions-go.md)
- [worktree.go](/modules/worktree-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
