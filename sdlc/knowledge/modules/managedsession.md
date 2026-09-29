---
type: Module
title: ManagedSession
description: "Graphify community 37: app_terminal_registry_test.go, internal/terminal/helper/server.go, internal/terminal/login_path.go, internal/terminal/manager.go, internal/terminal/session.go, internal/terminal"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 2f85a0b04c1b8ba7 }
  - { id: server, resource: internal/terminal/helper/server.go, last_modified: "2026-04-09T21:03:43+10:00", digest: e54a6a9ea36ca59d }
  - { id: login_path, resource: internal/terminal/login_path.go, last_modified: "2026-05-07T20:55:09+10:00", digest: 7ad84d121b34092b }
  - { id: manager, resource: internal/terminal/manager.go, last_modified: "2026-05-07T20:55:09+10:00", digest: 4c7107fb7905c1b1 }
  - { id: session, resource: internal/terminal/session.go, last_modified: "2026-04-09T18:42:25+10:00", digest: 6213e8a0ece661b9 }
  - { id: stub, resource: internal/terminal/stub.go, last_modified: "2026-04-09T11:24:44+10:00", digest: 4e852c6c9f8e0735 }
---

# Files
- `app_terminal_registry_test.go`
- `internal/terminal/helper/server.go`
- `internal/terminal/login_path.go`
- `internal/terminal/manager.go`
- `internal/terminal/session.go`
- `internal/terminal/stub.go`

# Symbols
- fakeSessionManager (app_terminal_registry_test.go:L43)
- .Spawn() (app_terminal_registry_test.go:L62)
- .Kill() (app_terminal_registry_test.go:L69)
- .IsAlive() (app_terminal_registry_test.go:L80)
- .FindByPID() (app_terminal_registry_test.go:L86)
- .Shutdown() (app_terminal_registry_test.go:L93)
- helperSession (internal/terminal/helper/server.go:L21)
- loginShellPATH() (internal/terminal/login_path.go:L22)
- resolveExecutable() (internal/terminal/login_path.go:L46)
- applyLoginPATH() (internal/terminal/login_path.go:L69)
- .Get() (internal/terminal/manager.go:L106)
- .Kill() (internal/terminal/manager.go:L117)
- .IsAlive() (internal/terminal/manager.go:L136)
- .List() (internal/terminal/manager.go:L147)
- .FindByPID() (internal/terminal/manager.go:L164)
- .Shutdown() (internal/terminal/manager.go:L180)
- SessionManager (internal/terminal/manager.go:L24)
- .Spawn() (internal/terminal/manager.go:L43)
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
- .Write() (internal/terminal/session.go:L52)
- .Snapshot() (internal/terminal/session.go:L78)
- terminal/stub.go (internal/terminal/stub.go:L1)
- NewStubSession() (internal/terminal/stub.go:L4)

# Depends on
- [Client](/modules/client.md)
- [session_test.go](/modules/session-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
