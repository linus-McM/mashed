---
type: Module
title: ClaudeCodeProvider
description: "Graphify community 267: internal/explain/explain.go, internal/scanner/claude.go, internal/scanner/sessions.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: explain, resource: internal/explain/explain.go, last_modified: "2026-04-10T11:36:27+10:00", digest: c8a0f1b0e6aad414 }
  - { id: claude, resource: internal/scanner/claude.go, last_modified: "2026-04-08T18:15:54+10:00", digest: 2c44ec40e17d728e }
  - { id: sessions, resource: internal/scanner/sessions.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 4878f58d15bdc696 }
---

# Files
- `internal/explain/explain.go`
- `internal/scanner/claude.go`
- `internal/scanner/sessions.go`

# Symbols
- Explainer (internal/explain/explain.go:L15)
- New() (internal/explain/explain.go:L21)
- ClaudeCodeProvider (internal/scanner/claude.go:L18)
- pidDirEntry (internal/scanner/claude.go:L32)
- NewClaudeCodeProvider() (internal/scanner/claude.go:L40)
- .SessionDir() (internal/scanner/claude.go:L70)
- .DevDir() (internal/scanner/claude.go:L76)
- .cachePidDir() (internal/scanner/claude.go:L80)
- .getCachedPidDir() (internal/scanner/claude.go:L86)
- sessionParserState (internal/scanner/sessions.go:L17)

# Depends on
- [go_pkg_strings](/modules/go-pkg-strings.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
