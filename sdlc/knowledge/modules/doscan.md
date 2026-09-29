---
type: Module
title: .doScan
description: "Graphify community 32: app.go, app_scan.go, internal/scanner/repos.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-30T01:21:31+10:00", digest: 774e87e5f070ece0 }
  - { id: app_scan, resource: app_scan.go, last_modified: "2026-05-07T10:33:04+10:00", digest: e5b2c4798c9f1cad }
  - { id: repos, resource: internal/scanner/repos.go, last_modified: "2026-05-07T10:33:04+10:00", digest: c8bd28e2f7bd59b8 }
---

# Files
- `app.go`
- `app_scan.go`
- `internal/scanner/repos.go`

# Symbols
- sanitizeID() (app.go:L846)
- repoNameFromDir() (app.go:L861)
- App (app_scan.go:L21)
- .initScanning() (app_scan.go:L21)
- .scanLoop() (app_scan.go:L47)
- .resolveTmuxTarget() (app_scan.go:L66)
- .doScan() (app_scan.go:L77)
- NewRepoScanner() (internal/scanner/repos.go:L31)

# Depends on
- [App](/modules/app-109.md)
- [ModelInfo](/modules/modelinfo.md)
- [time.Time](/modules/time-time.md)
- [tokensamples_test.go](/modules/tokensamples-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
