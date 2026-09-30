---
type: Module
title: SpawnRequest
description: "Graphify community 282: internal/terminal/helper/protocol.go, internal/terminal/login_path.go, internal/terminal/manager.go, internal/terminal/manager_argv_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: protocol, resource: internal/terminal/helper/protocol.go, last_modified: "2026-04-09T16:46:50+10:00", digest: 07f801b52b8e41da }
  - { id: login_path, resource: internal/terminal/login_path.go, last_modified: "2026-05-07T20:55:09+10:00", digest: 7ad84d121b34092b }
  - { id: manager, resource: internal/terminal/manager.go, last_modified: "2026-09-30T06:41:13+10:00", digest: ef641b063b6161d6 }
  - { id: manager_argv_test, resource: internal/terminal/manager_argv_test.go, last_modified: "2026-09-30T06:41:13+10:00", digest: 8de0ab57649ae98b }
---

# Files
- `internal/terminal/helper/protocol.go`
- `internal/terminal/login_path.go`
- `internal/terminal/manager.go`
- `internal/terminal/manager_argv_test.go`

# Symbols
- SpawnRequest (internal/terminal/helper/protocol.go:L35)
- loginShellPATH() (internal/terminal/login_path.go:L22)
- resolveExecutable() (internal/terminal/login_path.go:L46)
- applyLoginPATH() (internal/terminal/login_path.go:L69)
- helperSpawner (internal/terminal/manager.go:L32)
- newSessionManagerWith() (internal/terminal/manager.go:L48)
- fakeSpawner (internal/terminal/manager_argv_test.go:L18)
- .Spawn() (internal/terminal/manager_argv_test.go:L23)
- .Close() (internal/terminal/manager_argv_test.go:L35)
- .last() (internal/terminal/manager_argv_test.go:L37)
- TestSpawnArgv_PassesArgsVerbatim() (internal/terminal/manager_argv_test.go:L46)
- TestSpawnArgv_RejectsEmpty() (internal/terminal/manager_argv_test.go:L60)
- TestSpawn_StringFormUnchanged() (internal/terminal/manager_argv_test.go:L70)

# Depends on
- [ManagedSession](/modules/managedsession.md)
- [server_test.go](/modules/server-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
