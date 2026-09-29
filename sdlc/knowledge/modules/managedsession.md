---
type: Module
title: ManagedSession
description: "Graphify community 48: app_terminal_registry_test.go, internal/terminal/helper/server.go, internal/terminal/manager_test.go, internal/terminal/session.go, internal/terminal/session_test.go, internal/t"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 2f85a0b04c1b8ba7 }
  - { id: server, resource: internal/terminal/helper/server.go, last_modified: "2026-04-09T21:03:43+10:00", digest: e54a6a9ea36ca59d }
  - { id: manager_test, resource: internal/terminal/manager_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 8a61b932bfb6bc29 }
  - { id: session, resource: internal/terminal/session.go, last_modified: "2026-04-09T18:42:25+10:00", digest: 6213e8a0ece661b9 }
  - { id: session_test, resource: internal/terminal/session_test.go, last_modified: "2026-04-09T18:42:25+10:00", digest: 64fc12b5f28e698a }
  - { id: stub, resource: internal/terminal/stub.go, last_modified: "2026-04-09T11:24:44+10:00", digest: 4e852c6c9f8e0735 }
---

# Files
- `app_terminal_registry_test.go`
- `internal/terminal/helper/server.go`
- `internal/terminal/manager_test.go`
- `internal/terminal/session.go`
- `internal/terminal/session_test.go`
- `internal/terminal/stub.go`

# Symbols
- fakeSessionManager (app_terminal_registry_test.go:L43)
- .Spawn() (app_terminal_registry_test.go:L62)
- .Kill() (app_terminal_registry_test.go:L69)
- .IsAlive() (app_terminal_registry_test.go:L80)
- .FindByPID() (app_terminal_registry_test.go:L86)
- .Shutdown() (app_terminal_registry_test.go:L93)
- helperSession (internal/terminal/helper/server.go:L21)
- TestSessionManager_ConcurrentInjectAndKill() (internal/terminal/manager_test.go:L301)
- ManagedSession (internal/terminal/session.go:L104)
- newManagedSession() (internal/terminal/session.go:L119)
- newRemoteManagedSession() (internal/terminal/session.go:L138)
- .readLoop() (internal/terminal/session.go:L152)
- .cleanupClients() (internal/terminal/session.go:L207)
- .AddClient() (internal/terminal/session.go:L218)
- .RemoveClient() (internal/terminal/session.go:L241)
- .clientReader() (internal/terminal/session.go:L252)
- .Kill() (internal/terminal/session.go:L280)
- .IsAlive() (internal/terminal/session.go:L296)
- .Wait() (internal/terminal/session.go:L306)
- .Name() (internal/terminal/session.go:L311)
- scrollBuffer (internal/terminal/session.go:L36)
- newScrollBuffer() (internal/terminal/session.go:L44)
- .Write() (internal/terminal/session.go:L52)
- .Snapshot() (internal/terminal/session.go:L78)
- TestScrollBuffer_AC1_WriteAndSnapshot() (internal/terminal/session_test.go:L114)
- TestScrollBuffer_AC1_LargeOverCapacity() (internal/terminal/session_test.go:L187)
- TestScrollBuffer_AC1_SnapshotIndependentCopy() (internal/terminal/session_test.go:L208)
- terminal/stub.go (internal/terminal/stub.go:L1)
- NewStubSession() (internal/terminal/stub.go:L4)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [manager_test.go](/modules/manager-test-go.md)

# Features
- no feature plan names these files
