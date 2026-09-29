---
type: Module
title: sessions.go
description: "Graphify community 93: internal/domain/types.go, internal/scanner/sessions.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/schema.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: sessions, resource: internal/scanner/sessions.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 4878f58d15bdc696 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: schema, resource: internal/uiadapter/schema.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 31cb03cd04da1f49 }
---

# Files
- `internal/domain/types.go`
- `internal/scanner/sessions.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/schema.go`

# Symbols
- SubAgentInfo (internal/domain/types.go:L106)
- sessions.go (internal/scanner/sessions.go:L1)
- jsonlMessage (internal/scanner/sessions.go:L177)
- assistantMessage (internal/scanner/sessions.go:L182)
- tokenUsage (internal/scanner/sessions.go:L187)
- contentBlock (internal/scanner/sessions.go:L194)
- parseJSONLLine() (internal/scanner/sessions.go:L205)
- parseAssistantMessage() (internal/scanner/sessions.go:L223)
- parseUserMessage() (internal/scanner/sessions.go:L273)
- toolUseToLogLine() (internal/scanner/sessions.go:L309)
- extractToolDetail() (internal/scanner/sessions.go:L338)
- parseAgentToolInput() (internal/scanner/sessions.go:L387)
- collapseNewlines() (internal/scanner/sessions.go:L419)
- truncate() (internal/scanner/sessions.go:L426)
- responseBody (internal/uiadapter/backend/claudeapi/client.go:L195)
- .UnmarshalJSON() (internal/uiadapter/schema.go:L112)
- WidgetNode (internal/uiadapter/schema.go:L58)

# Depends on
- [ClaudeCodeProvider](/modules/claudecodeprovider.md)
- [domain/types.go](/modules/domain-types-go.md)
- [SessionData](/modules/sessiondata.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
