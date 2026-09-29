---
type: Module
title: Bridge
description: "Graphify community 209: internal/terminal/bridge.go, internal/terminal/bridge_auth_test.go, internal/terminal/bridge_test.go, internal/terminal/manager.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: bridge, resource: internal/terminal/bridge.go, last_modified: "2026-09-30T01:21:31+10:00", digest: 009ffd22cb3b9f12 }
  - { id: bridge_auth_test, resource: internal/terminal/bridge_auth_test.go, last_modified: "2026-09-30T01:21:31+10:00", digest: e232c80fb7794ab7 }
  - { id: bridge_test, resource: internal/terminal/bridge_test.go, last_modified: "2026-09-30T01:21:31+10:00", digest: a14cd774ef1482a3 }
  - { id: manager, resource: internal/terminal/manager.go, last_modified: "2026-05-07T20:55:09+10:00", digest: 4c7107fb7905c1b1 }
---

# Files
- `internal/terminal/bridge.go`
- `internal/terminal/bridge_auth_test.go`
- `internal/terminal/bridge_test.go`
- `internal/terminal/manager.go`

# Symbols
- NewBridge() (internal/terminal/bridge.go:L115)
- newBridge() (internal/terminal/bridge.go:L120)
- .Token() (internal/terminal/bridge.go:L136)
- .Start() (internal/terminal/bridge.go:L142)
- .GetTerminalPort() (internal/terminal/bridge.go:L176)
- .Stop() (internal/terminal/bridge.go:L181)
- .shutdown() (internal/terminal/bridge.go:L187)
- .authorize() (internal/terminal/bridge.go:L196)
- .handleWS() (internal/terminal/bridge.go:L213)
- .proxyTmuxSession() (internal/terminal/bridge.go:L288)
- truncateForClose() (internal/terminal/bridge.go:L378)
- normalizeBMADPaneTarget() (internal/terminal/bridge.go:L396)
- TmuxAttacher (internal/terminal/bridge.go:L92)
- Bridge (internal/terminal/bridge.go:L99)
- TestBridge_TokenUniquePerInstance() (internal/terminal/bridge_auth_test.go:L106)
- TestNewBridge_RandFailureFailsClosed() (internal/terminal/bridge_auth_test.go:L149)
- TestBridge_StopIdempotent() (internal/terminal/bridge_test.go:L330)
- TestBridge_AC1_NewBridgeWithManager() (internal/terminal/bridge_test.go:L86)
- .Get() (internal/terminal/manager.go:L106)
- .Kill() (internal/terminal/manager.go:L117)
- .IsAlive() (internal/terminal/manager.go:L136)
- .List() (internal/terminal/manager.go:L147)
- .FindByPID() (internal/terminal/manager.go:L164)
- .Shutdown() (internal/terminal/manager.go:L180)
- SessionManager (internal/terminal/manager.go:L24)

# Depends on
- [ManagedSession](/modules/managedsession.md)
- [sync.Mutex](/modules/sync-mutex.md)

# Inferred
- [manager_test.go](/modules/manager-test-go.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
