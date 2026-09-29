---
type: Module
title: ClaudeCodeProvider
description: "Graphify community 296: app.go, app_scan.go, internal/scanner/claude.go, internal/scanner/repos.go, internal/scanner/sessions.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_scan, resource: app_scan.go, last_modified: "2026-05-07T10:33:04+10:00", digest: e5b2c4798c9f1cad }
  - { id: claude, resource: internal/scanner/claude.go, last_modified: "2026-04-08T18:15:54+10:00", digest: 2c44ec40e17d728e }
  - { id: repos, resource: internal/scanner/repos.go, last_modified: "2026-05-07T10:33:04+10:00", digest: c8bd28e2f7bd59b8 }
  - { id: sessions, resource: internal/scanner/sessions.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 4878f58d15bdc696 }
---

# Files
- `app.go`
- `app_scan.go`
- `internal/scanner/claude.go`
- `internal/scanner/repos.go`
- `internal/scanner/sessions.go`

# Symbols
- sanitizeID() (app.go:L779)
- repoNameFromDir() (app.go:L794)
- App (app_scan.go:L21)
- .initScanning() (app_scan.go:L21)
- .scanLoop() (app_scan.go:L47)
- .resolveTmuxTarget() (app_scan.go:L66)
- .doScan() (app_scan.go:L77)
- ClaudeCodeProvider (internal/scanner/claude.go:L18)
- pidDirEntry (internal/scanner/claude.go:L32)
- NewClaudeCodeProvider() (internal/scanner/claude.go:L40)
- .SessionDir() (internal/scanner/claude.go:L70)
- .DevDir() (internal/scanner/claude.go:L76)
- .cachePidDir() (internal/scanner/claude.go:L80)
- .getCachedPidDir() (internal/scanner/claude.go:L86)
- NewRepoScanner() (internal/scanner/repos.go:L31)
- sessionParserState (internal/scanner/sessions.go:L17)

# Depends on
- [ModelInfo](/modules/modelinfo.md)
- [testing.T](/modules/testing-t.md)
- [time.Time](/modules/time-time.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
