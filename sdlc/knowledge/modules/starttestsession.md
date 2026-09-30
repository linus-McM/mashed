---
type: Module
title: startTestSession
description: "Graphify community 251: internal/terminal/session_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: session_test, resource: internal/terminal/session_test.go, last_modified: "2026-04-09T18:42:25+10:00", digest: 64fc12b5f28e698a }
---

# Files
- `internal/terminal/session_test.go`

# Symbols
- readWSTimeout() (internal/terminal/session_test.go:L103)
- TestManagedSession_AC2_TwoClientsReceiveOutput() (internal/terminal/session_test.go:L235)
- TestManagedSession_AC3_NewClientGetsSnapshot() (internal/terminal/session_test.go:L261)
- TestManagedSession_AC4_KillTerminatesProcess() (internal/terminal/session_test.go:L283)
- TestManagedSession_AC4_DoubleKillNoPanic() (internal/terminal/session_test.go:L298)
- TestManagedSession_AC4_NaturalExitDetected() (internal/terminal/session_test.go:L311)
- TestManagedSession_AC5_FailedClientIsolation() (internal/terminal/session_test.go:L327)
- wsTestPair() (internal/terminal/session_test.go:L45)
- canPTYSpawn() (internal/terminal/session_test.go:L72)
- startTestSession() (internal/terminal/session_test.go:L85)

# Depends on
- [ManagedSession](/modules/managedsession.md)

# Inferred
- [ManagedSession](/modules/managedsession.md)

# Features
- no feature plan names these files
