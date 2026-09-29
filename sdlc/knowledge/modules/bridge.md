---
type: Module
title: Bridge
description: "Graphify community 129: internal/terminal/bridge.go, internal/terminal/bridge_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: bridge, resource: internal/terminal/bridge.go, last_modified: "2026-04-11T19:45:53+10:00", digest: a42cec584372611b }
  - { id: bridge_test, resource: internal/terminal/bridge_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: b13529690aa8d966 }
---

# Files
- `internal/terminal/bridge.go`
- `internal/terminal/bridge_test.go`

# Symbols
- .Start() (internal/terminal/bridge.go:L107)
- .GetTerminalPort() (internal/terminal/bridge.go:L138)
- .Stop() (internal/terminal/bridge.go:L143)
- .shutdown() (internal/terminal/bridge.go:L149)
- .handleWS() (internal/terminal/bridge.go:L155)
- .proxyTmuxSession() (internal/terminal/bridge.go:L224)
- truncateForClose() (internal/terminal/bridge.go:L314)
- normalizeBMADPaneTarget() (internal/terminal/bridge.go:L332)
- TmuxAttacher (internal/terminal/bridge.go:L78)
- Bridge (internal/terminal/bridge.go:L85)
- NewBridge() (internal/terminal/bridge.go:L99)
- TestBridge_StopIdempotent() (internal/terminal/bridge_test.go:L327)
- TestBridge_AC1_NewBridgeWithManager() (internal/terminal/bridge_test.go:L86)

# Depends on
- [ManagedSession](/modules/managedsession.md)

# Inferred
- [manager_test.go](/modules/manager-test-go.md)

# Features
- no feature plan names these files
