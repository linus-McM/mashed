---
type: Module
title: sessions.go
description: "Graphify community 93: app_sessions.go, internal/domain/types.go, internal/scanner/sessions.go, internal/uiadapter/backend/claudeapi/client.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app_sessions, resource: app_sessions.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 7b79f8178152c1fb }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: sessions, resource: internal/scanner/sessions.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 4878f58d15bdc696 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
---

# Files
- `app_sessions.go`
- `internal/domain/types.go`
- `internal/scanner/sessions.go`
- `internal/uiadapter/backend/claudeapi/client.go`

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
- sessions.go (internal/scanner/sessions.go:L1)
- fileInode() (internal/scanner/sessions.go:L167)
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
- responseBody (internal/uiadapter/backend/claudeapi/client.go:L195)

# Depends on
- [ClaudeCodeProvider](/modules/claudecodeprovider.md)
- [domain/types.go](/modules/domain-types-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
