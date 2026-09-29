---
type: Module
title: SessionData
description: "Graphify community 99: app_sessions.go, internal/domain/types.go, internal/scanner/sessions.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app_sessions, resource: app_sessions.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 7b79f8178152c1fb }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: sessions, resource: internal/scanner/sessions.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 4878f58d15bdc696 }
---

# Files
- `app_sessions.go`
- `internal/domain/types.go`
- `internal/scanner/sessions.go`

# Symbols
- .findUnclaimed() (app_sessions.go:L116)
- App (app_sessions.go:L15)
- .watchSessions() (app_sessions.go:L15)
- .inferStatus() (app_sessions.go:L165)
- .findSessionByID() (app_sessions.go:L37)
- .findLatestSession() (app_sessions.go:L50)
- .consumeEngineEvents() (app_sessions.go:L86)
- SubAgentInfo (internal/domain/types.go:L106)
- SessionData (internal/domain/types.go:L91)
- fileInode() (internal/scanner/sessions.go:L167)
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
- [domain/types.go](/modules/domain-types-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
