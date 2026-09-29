---
type: Module
title: manager_test.go
description: "Graphify community 97: internal/terminal/manager.go, internal/terminal/manager_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: manager, resource: internal/terminal/manager.go, last_modified: "2026-05-07T20:55:09+10:00", digest: 4c7107fb7905c1b1 }
  - { id: manager_test, resource: internal/terminal/manager_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 8a61b932bfb6bc29 }
---

# Files
- `internal/terminal/manager.go`
- `internal/terminal/manager_test.go`

# Symbols
- NewSessionManager() (internal/terminal/manager.go:L32)
- manager_test.go (internal/terminal/manager_test.go:L1)
- TestRemoteSession_NilCmd() (internal/terminal/manager_test.go:L102)
- TestRemoteSession_ReadLoop() (internal/terminal/manager_test.go:L110)
- TestSessionManager_GetKillLifecycle() (internal/terminal/manager_test.go:L130)
- TestSessionManager_KillNonExistent() (internal/terminal/manager_test.go:L150)
- TestSessionManager_GetNonExistent() (internal/terminal/manager_test.go:L157)
- TestSessionManager_IsAliveNonExistent() (internal/terminal/manager_test.go:L164)
- TestSessionManager_DuplicateName() (internal/terminal/manager_test.go:L173)
- TestSessionManager_FindByPID() (internal/terminal/manager_test.go:L195)
- TestSessionManager_FindByPID_Unknown() (internal/terminal/manager_test.go:L207)
- TestSessionManager_FindByPID_EmptyManager() (internal/terminal/manager_test.go:L218)
- TestSessionManager_Shutdown() (internal/terminal/manager_test.go:L229)
- TestSessionManager_ShutdownIdempotent() (internal/terminal/manager_test.go:L253)
- TestSessionManager_List_All() (internal/terminal/manager_test.go:L265)
- TestSessionManager_List_WithFilter() (internal/terminal/manager_test.go:L277)
- TestSessionManager_List_Empty() (internal/terminal/manager_test.go:L291)
- TestSessionManager_ConcurrentInjectAndKill() (internal/terminal/manager_test.go:L301)
- injectRemoteSession() (internal/terminal/manager_test.go:L34)
- TestRemoteSession_KillNoCmdWait() (internal/terminal/manager_test.go:L352)
- TestSessionManager_NilClient_SpawnReturnsErrHelperNotRunning() (internal/terminal/manager_test.go:L57)
- TestSessionManager_NilClient_SpawnWithCommand() (internal/terminal/manager_test.go:L68)
- TestSessionManager_NilClient_ListReturnsEmpty() (internal/terminal/manager_test.go:L77)
- TestSessionManager_NilClient_IsAliveReturnsFalse() (internal/terminal/manager_test.go:L83)
- TestRemoteSession_IsAlive() (internal/terminal/manager_test.go:L92)

# Depends on
- [Client](/modules/client.md)
- [ManagedSession](/modules/managedsession.md)

# Inferred
- [ManagedSession](/modules/managedsession.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
