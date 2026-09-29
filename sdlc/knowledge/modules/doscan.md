---
type: Module
title: .doScan
description: "Graphify community 452: app.go, app_scan.go, app_spawn.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_scan, resource: app_scan.go, last_modified: "2026-05-07T10:33:04+10:00", digest: e5b2c4798c9f1cad }
  - { id: app_spawn, resource: app_spawn.go, last_modified: "2026-05-07T18:18:02+10:00", digest: cbe44d5e06097d50 }
---

# Files
- `app.go`
- `app_scan.go`
- `app_spawn.go`

# Symbols
- sanitizeID() (app.go:L779)
- repoNameFromDir() (app.go:L794)
- App (app_scan.go:L21)
- .initScanning() (app_scan.go:L21)
- .scanLoop() (app_scan.go:L47)
- .resolveTmuxTarget() (app_scan.go:L66)
- .doScan() (app_scan.go:L77)
- App (app_spawn.go:L19)
- .spawnSession() (app_spawn.go:L19)
- .SpawnAgent() (app_spawn.go:L49)
- .SpawnAgentWithCommand() (app_spawn.go:L62)
- .SpawnTerminal() (app_spawn.go:L71)
- .KillAgent() (app_spawn.go:L99)

# Depends on
- [App](/modules/app-349.md)
- [ClaudeCodeProvider](/modules/claudecodeprovider.md)
- [domain/types.go](/modules/domain-types-go.md)
- [ModelInfo](/modules/modelinfo.md)
- [RepoScanner](/modules/reposcanner.md)
- [tokensamples_test.go](/modules/tokensamples-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
