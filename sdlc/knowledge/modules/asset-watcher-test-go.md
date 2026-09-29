---
type: Module
title: asset_watcher_test.go
description: "Graphify community 283: internal/bmad/asset_watcher.go, internal/bmad/asset_watcher_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: asset_watcher, resource: internal/bmad/asset_watcher.go, last_modified: "2026-04-12T16:58:02+10:00", digest: 25909dc6b81603ba }
  - { id: asset_watcher_test, resource: internal/bmad/asset_watcher_test.go, last_modified: "2026-04-12T16:58:02+10:00", digest: 4e1bcbed1fab098b }
---

# Files
- `internal/bmad/asset_watcher.go`
- `internal/bmad/asset_watcher_test.go`

# Symbols
- isMarkdown() (internal/bmad/asset_watcher.go:L102)
- AssetWatchRoots() (internal/bmad/asset_watcher.go:L166)
- NewAssetWatcher() (internal/bmad/asset_watcher.go:L38)
- asset_watcher_test.go (internal/bmad/asset_watcher_test.go:L1)
- TestAssetWatcher_AC4_NonexistentRootSilentlySkipped() (internal/bmad/asset_watcher_test.go:L122)
- TestAssetWatcher_AC5_CleansUpOnContextCancel() (internal/bmad/asset_watcher_test.go:L150)
- TestAssetWatcher_AC6_NonMarkdownFilesIgnored() (internal/bmad/asset_watcher_test.go:L182)
- TestAssetWatcher_AC1_SingleWriteDebouncedEvent() (internal/bmad/asset_watcher_test.go:L19)
- TestAssetWatcher_DoubleStartReturnsError() (internal/bmad/asset_watcher_test.go:L212)
- TestAssetWatchRoots() (internal/bmad/asset_watcher_test.go:L227)
- TestIsMarkdown() (internal/bmad/asset_watcher_test.go:L246)
- TestAssetWatcher_AC2_BurstCoalescesToOneEvent() (internal/bmad/asset_watcher_test.go:L55)
- TestAssetWatcher_AC3_NewSubdirectoryWatchedDynamically() (internal/bmad/asset_watcher_test.go:L87)

# Depends on
- [App](/modules/app-109.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
