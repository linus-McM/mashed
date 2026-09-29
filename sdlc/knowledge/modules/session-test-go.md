---
type: Module
title: session_test.go
description: "Graphify community 237: internal/terminal/bridge_test.go, internal/terminal/session.go, internal/terminal/session_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: bridge_test, resource: internal/terminal/bridge_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: b13529690aa8d966 }
  - { id: session, resource: internal/terminal/session.go, last_modified: "2026-04-09T18:42:25+10:00", digest: 6213e8a0ece661b9 }
  - { id: session_test, resource: internal/terminal/session_test.go, last_modified: "2026-04-09T18:42:25+10:00", digest: 64fc12b5f28e698a }
---

# Files
- `internal/terminal/bridge_test.go`
- `internal/terminal/session.go`
- `internal/terminal/session_test.go`

# Symbols
- TestBridge_AC2_PTYSessionTakesPrecedence() (internal/terminal/bridge_test.go:L545)
- newScrollBuffer() (internal/terminal/session.go:L44)
- session_test.go (internal/terminal/session_test.go:L1)
- readWSTimeout() (internal/terminal/session_test.go:L103)
- TestScrollBuffer_AC1_WriteAndSnapshot() (internal/terminal/session_test.go:L114)
- TestScrollBuffer_AC1_LargeOverCapacity() (internal/terminal/session_test.go:L187)
- TestScrollBuffer_AC1_SnapshotIndependentCopy() (internal/terminal/session_test.go:L208)
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
- [bridge_test.go](/modules/bridge-test-go.md)
- [ManagedSession](/modules/managedsession.md)

# Inferred
- [ManagedSession](/modules/managedsession.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
