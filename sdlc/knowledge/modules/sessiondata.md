---
type: Module
title: SessionData
description: "Graphify community 99: app_sessions.go, internal/domain/types.go, internal/scanner/sessions.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
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
- SessionData (internal/domain/types.go:L91)
- fileInode() (internal/scanner/sessions.go:L167)
- ClaudeCodeProvider (internal/scanner/sessions.go:L36)
- .ParseSession() (internal/scanner/sessions.go:L36)
- sessionIDFromPath() (internal/scanner/sessions.go:L415)
- .ParseSessionIncremental() (internal/scanner/sessions.go:L89)

# Depends on
- [domain/types.go](/modules/domain-types-go.md)
- [sessions.go](/modules/sessions-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
