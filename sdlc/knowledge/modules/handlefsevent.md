---
type: Module
title: .handleFSEvent
description: "Graphify community 286: internal/domain/types.go, internal/scanner/watcher.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: types, resource: internal/domain/types.go, last_modified: "2026-04-11T17:08:28+10:00", digest: a6b057db991af9b4 }
  - { id: watcher, resource: internal/scanner/watcher.go, last_modified: "2026-04-07T10:03:32+10:00", digest: 8aeb13ebec4f12a0 }
---

# Files
- `internal/domain/types.go`
- `internal/scanner/watcher.go`

# Symbols
- SessionEvent (internal/domain/types.go:L243)
- .repoPathFromSessionDir() (internal/scanner/watcher.go:L116)
- ClaudeCodeProvider (internal/scanner/watcher.go:L19)
- .WatchSessions() (internal/scanner/watcher.go:L19)
- .watchLoop() (internal/scanner/watcher.go:L48)
- .handleFSEvent() (internal/scanner/watcher.go:L82)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
