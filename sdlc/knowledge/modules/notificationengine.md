---
type: Module
title: NotificationEngine
description: "Graphify community 187: app.go, internal/agent/engine.go, internal/agent/engine_test.go, internal/agent/tokensamples_test.go, internal/domain/types.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-30T06:52:18+10:00", digest: 19038d72636ae53b }
  - { id: engine, resource: internal/agent/engine.go, last_modified: "2026-04-11T17:08:28+10:00", digest: 1042190db571e4e7 }
  - { id: engine_test, resource: internal/agent/engine_test.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 745feda4505cf50c }
  - { id: tokensamples_test, resource: internal/agent/tokensamples_test.go, last_modified: "2026-04-11T17:08:28+10:00", digest: e73f3aa2bb785df4 }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
---

# Files
- `app.go`
- `internal/agent/engine.go`
- `internal/agent/engine_test.go`
- `internal/agent/tokensamples_test.go`
- `internal/domain/types.go`

# Symbols
- .GetNotifications() (app.go:L855)
- SortByPriority() (internal/agent/engine.go:L156)
- .RemoveAgent() (internal/agent/engine.go:L171)
- .ActiveAgentCount() (internal/agent/engine.go:L178)
- NotificationEngine (internal/agent/engine.go:L50)
- NewNotificationEngine() (internal/agent/engine.go:L60)
- .Events() (internal/agent/engine.go:L69)
- engine_test.go (internal/agent/engine_test.go:L1)
- drainEvents() (internal/agent/engine_test.go:L11)
- TestSortByPriority() (internal/agent/engine_test.go:L180)
- TestProcessAgentUpdate_EmptyID() (internal/agent/engine_test.go:L207)
- TestActiveAgentCount() (internal/agent/engine_test.go:L222)
- TestStateTransition_RunningToError() (internal/agent/engine_test.go:L28)
- TestDeduplication_SameStateTwice() (internal/agent/engine_test.go:L62)
- TestAgentRestart_CompletedStartedRunning() (internal/agent/engine_test.go:L94)
- TestProcessAgentUpdate_PropagatesTokenSamples() (internal/agent/tokensamples_test.go:L142)
- NotificationEvent (internal/domain/types.go:L168)

# Depends on
- [sessions.go](/modules/sessions-go.md)
- [time.Time](/modules/time-time.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
