---
type: Module
title: mockTmuxSession
description: "Graphify community 341: internal/terminal/bridge.go, internal/terminal/bridge_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: bridge, resource: internal/terminal/bridge.go, last_modified: "2026-04-11T19:45:53+10:00", digest: a42cec584372611b }
  - { id: bridge_test, resource: internal/terminal/bridge_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: b13529690aa8d966 }
---

# Files
- `internal/terminal/bridge.go`
- `internal/terminal/bridge_test.go`

# Symbols
- TmuxSession (internal/terminal/bridge.go:L65)
- mockResizeCall (internal/terminal/bridge_test.go:L372)
- mockTmuxSession (internal/terminal/bridge_test.go:L381)
- .Read() (internal/terminal/bridge_test.go:L402)
- .SendInput() (internal/terminal/bridge_test.go:L414)
- .SendKey() (internal/terminal/bridge_test.go:L425)
- .Resize() (internal/terminal/bridge_test.go:L433)
- mockTmuxAttacher (internal/terminal/bridge_test.go:L456)
- .Attach() (internal/terminal/bridge_test.go:L473)

# Depends on
- [startBridgeWithMockAdapter](/modules/startbridgewithmockadapter.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
