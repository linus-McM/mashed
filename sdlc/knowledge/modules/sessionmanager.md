---
type: Module
title: SessionManager
description: "Graphify community 322: internal/terminal/login_path.go, internal/terminal/manager.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: login_path, resource: internal/terminal/login_path.go, last_modified: "2026-05-07T20:55:09+10:00", digest: 7ad84d121b34092b }
  - { id: manager, resource: internal/terminal/manager.go, last_modified: "2026-05-07T20:55:09+10:00", digest: 4c7107fb7905c1b1 }
---

# Files
- `internal/terminal/login_path.go`
- `internal/terminal/manager.go`

# Symbols
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

# Depends on
- [ManagedSession](/modules/managedsession.md)
- [server_test.go](/modules/server-test-go.md)

# Inferred
- [ManagedSession](/modules/managedsession.md)

# Features
- no feature plan names these files
