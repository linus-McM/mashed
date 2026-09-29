---
type: Module
title: ManagedSession
description: "Graphify community 37: app_terminal_registry_test.go, internal/terminal/helper/server.go, internal/terminal/manager.go, internal/terminal/session.go, internal/terminal/stub.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-09-30T06:45:21+10:00", digest: 2844ed2cca730a69 }
  - { id: server, resource: internal/terminal/helper/server.go, last_modified: "2026-04-09T21:03:43+10:00", digest: e54a6a9ea36ca59d }
  - { id: manager, resource: internal/terminal/manager.go, last_modified: "2026-09-30T06:41:13+10:00", digest: ef641b063b6161d6 }
  - { id: session, resource: internal/terminal/session.go, last_modified: "2026-09-30T06:55:02+10:00", digest: 1a16af1bdd0b9f24 }
  - { id: stub, resource: internal/terminal/stub.go, last_modified: "2026-04-09T11:24:44+10:00", digest: 4e852c6c9f8e0735 }
---

# Files
- `app_terminal_registry_test.go`
- `internal/terminal/helper/server.go`
- `internal/terminal/manager.go`
- `internal/terminal/session.go`
- `internal/terminal/stub.go`

# Symbols
- .FindByPID() (app_terminal_registry_test.go:L102)
- .Shutdown() (app_terminal_registry_test.go:L109)
- fakeSessionManager (app_terminal_registry_test.go:L43)
- .SpawnArgv() (app_terminal_registry_test.go:L52)
- .spawnedArgvs() (app_terminal_registry_test.go:L61)
- .Spawn() (app_terminal_registry_test.go:L78)
- .Kill() (app_terminal_registry_test.go:L85)
- .IsAlive() (app_terminal_registry_test.go:L96)
- helperSession (internal/terminal/helper/server.go:L21)
- .Get() (internal/terminal/manager.go:L129)
- .Kill() (internal/terminal/manager.go:L140)
- .IsAlive() (internal/terminal/manager.go:L159)
- .List() (internal/terminal/manager.go:L170)
- .FindByPID() (internal/terminal/manager.go:L187)
- .Shutdown() (internal/terminal/manager.go:L203)
- SessionManager (internal/terminal/manager.go:L24)
- .Spawn() (internal/terminal/manager.go:L62)
- .SpawnArgv() (internal/terminal/manager.go:L79)
- ManagedSession (internal/terminal/session.go:L104)
- newManagedSession() (internal/terminal/session.go:L119)
- newRemoteManagedSession() (internal/terminal/session.go:L138)
- .readLoop() (internal/terminal/session.go:L152)
- .cleanupClients() (internal/terminal/session.go:L207)
- .AddClient() (internal/terminal/session.go:L218)
- .RemoveClient() (internal/terminal/session.go:L251)
- .clientReader() (internal/terminal/session.go:L262)
- .Kill() (internal/terminal/session.go:L290)
- .IsAlive() (internal/terminal/session.go:L306)
- .Wait() (internal/terminal/session.go:L316)
- .Name() (internal/terminal/session.go:L321)
- scrollBuffer (internal/terminal/session.go:L36)
- .Write() (internal/terminal/session.go:L52)
- .Snapshot() (internal/terminal/session.go:L78)
- terminal/stub.go (internal/terminal/stub.go:L1)
- NewStubSession() (internal/terminal/stub.go:L4)

# Depends on
- [session_test.go](/modules/session-test-go.md)
- [SpawnRequest](/modules/spawnrequest.md)

# Inferred
- [SpawnRequest](/modules/spawnrequest.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
