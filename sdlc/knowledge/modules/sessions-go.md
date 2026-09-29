---
type: Module
title: sessions.go
description: "Graphify community 93: internal/domain/types.go, internal/scanner/sessions.go, internal/uiadapter/backend/claudeapi/client.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: sessions, resource: internal/scanner/sessions.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 4878f58d15bdc696 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
---

# Files
- `internal/domain/types.go`
- `internal/scanner/sessions.go`
- `internal/uiadapter/backend/claudeapi/client.go`

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

# Depends on
- [ClaudeCodeProvider](/modules/claudecodeprovider.md)
- [domain/types.go](/modules/domain-types-go.md)
- [SessionData](/modules/sessiondata.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
