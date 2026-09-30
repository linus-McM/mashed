---
type: Module
title: AssetWatcher
description: "Graphify community 286: internal/bmad/asset_watcher.go, internal/domain/types.go, internal/scanner/watcher.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: asset_watcher, resource: internal/bmad/asset_watcher.go, last_modified: "2026-04-12T16:58:02+10:00", digest: 25909dc6b81603ba }
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: watcher, resource: internal/scanner/watcher.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 8aeb13ebec4f12a0 }
---

# Files
- `internal/bmad/asset_watcher.go`
- `internal/domain/types.go`
- `internal/scanner/watcher.go`

# Symbols
- .loop() (internal/bmad/asset_watcher.go:L108)
- AssetWatcher (internal/bmad/asset_watcher.go:L26)
- .Start() (internal/bmad/asset_watcher.go:L48)
- .Stop() (internal/bmad/asset_watcher.go:L71)
- .addRecursive() (internal/bmad/asset_watcher.go:L82)
- SessionEvent (internal/domain/types.go:L243)
- .repoPathFromSessionDir() (internal/scanner/watcher.go:L116)
- ClaudeCodeProvider (internal/scanner/watcher.go:L19)
- .WatchSessions() (internal/scanner/watcher.go:L19)
- .watchLoop() (internal/scanner/watcher.go:L48)
- .handleFSEvent() (internal/scanner/watcher.go:L82)

# Depends on
- [asset_watcher_test.go](/modules/asset-watcher-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
