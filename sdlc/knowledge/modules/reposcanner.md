---
type: Module
title: RepoScanner
description: "Graphify community 32: app.go, app_scan.go, app_sessions.go, internal/domain/types.go, internal/scanner/repos.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-30T06:52:18+10:00", digest: 19038d72636ae53b }
  - { id: app_scan, resource: app_scan.go, last_modified: "2026-09-30T06:49:47+10:00", digest: 3db9f1956543f87b }
  - { id: app_sessions, resource: app_sessions.go, last_modified: "2026-09-30T06:49:47+10:00", digest: 3b583d6c0cbcb9a3 }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: repos, resource: internal/scanner/repos.go, last_modified: "2026-05-07T10:33:04+10:00", digest: c8bd28e2f7bd59b8 }
---

# Files
- `app.go`
- `app_scan.go`
- `app_sessions.go`
- `internal/domain/types.go`
- `internal/scanner/repos.go`

# Symbols
- sanitizeID() (app.go:L904)
- .resolveTmuxTarget() (app_scan.go:L110)
- .doScan() (app_scan.go:L121)
- scanState (app_scan.go:L20)
- App (app_scan.go:L27)
- .scanSnapshot() (app_scan.go:L27)
- .initScanning() (app_scan.go:L38)
- .stopScanning() (app_scan.go:L73)
- .stopScanningLocked() (app_scan.go:L79)
- .scanLoop() (app_scan.go:L88)
- .watchSessions() (app_sessions.go:L18)
- AgentSession (internal/domain/types.go:L137)
- RepoInfo (internal/domain/types.go:L146)
- .fetchGitInfo() (internal/scanner/repos.go:L122)
- RepoScanner (internal/scanner/repos.go:L17)
- .InvalidateCache() (internal/scanner/repos.go:L173)
- repoCacheEntry (internal/scanner/repos.go:L25)
- NewRepoScanner() (internal/scanner/repos.go:L31)
- .ScanRepos() (internal/scanner/repos.go:L48)
- .getRepoInfo() (internal/scanner/repos.go:L96)

# Depends on
- [ModelInfo](/modules/modelinfo.md)
- [sessions.go](/modules/sessions-go.md)
- [testing.T](/modules/testing-t.md)

# Inferred
- [App](/modules/app-355.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
