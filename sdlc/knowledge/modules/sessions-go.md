---
type: Module
title: sessions.go
description: "Graphify community 164: app_sessions.go, app_spawn.go, internal/agent/engine.go, internal/agent/engine_test.go, internal/domain/types.go, internal/scanner/claude.go, internal/scanner/sessions.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_sessions, resource: app_sessions.go, last_modified: "2026-09-30T06:49:47+10:00", digest: 3b583d6c0cbcb9a3 }
  - { id: app_spawn, resource: app_spawn.go, last_modified: "2026-09-30T06:49:47+10:00", digest: 58aa6b29e35c2f7b }
  - { id: engine, resource: internal/agent/engine.go, last_modified: "2026-04-11T17:08:28+10:00", digest: 1042190db571e4e7 }
  - { id: engine_test, resource: internal/agent/engine_test.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 745feda4505cf50c }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: claude, resource: internal/scanner/claude.go, last_modified: "2026-04-08T18:15:54+10:00", digest: 2c44ec40e17d728e }
  - { id: sessions, resource: internal/scanner/sessions.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 4878f58d15bdc696 }
---

# Files
- `app_sessions.go`
- `app_spawn.go`
- `internal/agent/engine.go`
- `internal/agent/engine_test.go`
- `internal/domain/types.go`
- `internal/scanner/claude.go`
- `internal/scanner/sessions.go`

# Symbols
- .findUnclaimed() (app_sessions.go:L124)
- .inferStatus() (app_sessions.go:L173)
- App (app_sessions.go:L18)
- .findSessionByID() (app_sessions.go:L43)
- .findLatestSession() (app_sessions.go:L56)
- .consumeEngineEvents() (app_sessions.go:L92)
- .GetAgentLog() (app_spawn.go:L108)
- .GetAgentStatus() (internal/agent/engine.go:L185)
- priorityFor() (internal/agent/engine.go:L197)
- statusToEventType() (internal/agent/engine.go:L218)
- buildSummary() (internal/agent/engine.go:L241)
- lastActivity() (internal/agent/engine.go:L266)
- .ProcessAgentUpdate() (internal/agent/engine.go:L75)
- TestPriorityOrdering() (internal/agent/engine_test.go:L145)
- domain/types.go (internal/domain/types.go:L1)
- AgentStatus (internal/domain/types.go:L10)
- SubAgentInfo (internal/domain/types.go:L106)
- EventType (internal/domain/types.go:L157)
- AgentProvider (internal/domain/types.go:L221)
- LogKind (internal/domain/types.go:L24)
- LogLine (internal/domain/types.go:L36)
- Agent (internal/domain/types.go:L43)
- DagEdge (internal/domain/types.go:L61)
- Workflow (internal/domain/types.go:L67)
- Repo (internal/domain/types.go:L77)
- SessionData (internal/domain/types.go:L91)
- ClaudeCodeProvider (internal/scanner/claude.go:L18)
- pidDirEntry (internal/scanner/claude.go:L32)
- NewClaudeCodeProvider() (internal/scanner/claude.go:L40)
- .SessionDir() (internal/scanner/claude.go:L70)
- .DevDir() (internal/scanner/claude.go:L76)
- .cachePidDir() (internal/scanner/claude.go:L80)
- .getCachedPidDir() (internal/scanner/claude.go:L86)
- sessions.go (internal/scanner/sessions.go:L1)
- fileInode() (internal/scanner/sessions.go:L167)
- sessionParserState (internal/scanner/sessions.go:L17)
- jsonlMessage (internal/scanner/sessions.go:L177)
- assistantMessage (internal/scanner/sessions.go:L182)
- tokenUsage (internal/scanner/sessions.go:L187)
- contentBlock (internal/scanner/sessions.go:L194)
- parseJSONLLine() (internal/scanner/sessions.go:L205)
- parseAssistantMessage() (internal/scanner/sessions.go:L223)
- parseUserMessage() (internal/scanner/sessions.go:L273)
- toolUseToLogLine() (internal/scanner/sessions.go:L309)
- extractToolDetail() (internal/scanner/sessions.go:L338)
- ClaudeCodeProvider (internal/scanner/sessions.go:L36)
- .ParseSession() (internal/scanner/sessions.go:L36)
- parseAgentToolInput() (internal/scanner/sessions.go:L387)
- sessionIDFromPath() (internal/scanner/sessions.go:L415)
- collapseNewlines() (internal/scanner/sessions.go:L419)
- truncate() (internal/scanner/sessions.go:L426)
- .ParseSessionIncremental() (internal/scanner/sessions.go:L89)

# Depends on
- [App](/modules/app-355.md)
- [diff.go](/modules/diff-go.md)
- [.handleFSEvent](/modules/handlefsevent.md)
- [NotificationEngine](/modules/notificationengine.md)
- [RepoScanner](/modules/reposcanner.md)
- [worktree.go](/modules/worktree-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
