---
type: Module
title: mockTmuxSession
description: "Graphify community 342: internal/terminal/bridge.go, internal/terminal/bridge_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: bridge, resource: internal/terminal/bridge.go, last_modified: "2026-09-30T01:21:31+10:00", digest: 009ffd22cb3b9f12 }
  - { id: bridge_test, resource: internal/terminal/bridge_test.go, last_modified: "2026-09-30T01:21:31+10:00", digest: a14cd774ef1482a3 }
---

# Files
- `internal/terminal/bridge.go`
- `internal/terminal/bridge_test.go`

# Symbols
- TmuxSession (internal/terminal/bridge.go:L79)
- mockResizeCall (internal/terminal/bridge_test.go:L375)
- mockTmuxSession (internal/terminal/bridge_test.go:L384)
- .Read() (internal/terminal/bridge_test.go:L405)
- .SendInput() (internal/terminal/bridge_test.go:L417)
- .SendKey() (internal/terminal/bridge_test.go:L428)
- .Resize() (internal/terminal/bridge_test.go:L436)
- mockTmuxAttacher (internal/terminal/bridge_test.go:L459)
- .Attach() (internal/terminal/bridge_test.go:L476)

# Depends on
- [bridge_test.go](/modules/bridge-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
