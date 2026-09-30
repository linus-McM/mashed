---
type: Module
title: session_test.go
description: "Graphify community 237: internal/terminal/session.go, internal/terminal/session_test.go"
resource: internal/terminal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: session, resource: internal/terminal/session.go, last_modified: "2026-09-30T06:55:02+10:00", digest: 1a16af1bdd0b9f24 }
  - { id: session_test, resource: internal/terminal/session_test.go, last_modified: "2026-09-30T06:55:02+10:00", digest: 749d8be26cc054f3 }
---

# Files
- `internal/terminal/session.go`
- `internal/terminal/session_test.go`

# Symbols
- newScrollBuffer() (internal/terminal/session.go:L44)
- session_test.go (internal/terminal/session_test.go:L1)
- readWSTimeout() (internal/terminal/session_test.go:L103)
- TestScrollBuffer_AC1_WriteAndSnapshot() (internal/terminal/session_test.go:L114)
- TestScrollBuffer_AC1_LargeOverCapacity() (internal/terminal/session_test.go:L187)
- TestScrollBuffer_AC1_SnapshotIndependentCopy() (internal/terminal/session_test.go:L208)
- TestManagedSession_AC2_TwoClientsReceiveOutput() (internal/terminal/session_test.go:L235)
- TestManagedSession_LateClientAfterExitGetsScrollback() (internal/terminal/session_test.go:L259)
- TestManagedSession_AC3_NewClientGetsSnapshot() (internal/terminal/session_test.go:L290)
- TestManagedSession_AC4_KillTerminatesProcess() (internal/terminal/session_test.go:L312)
- TestManagedSession_AC4_DoubleKillNoPanic() (internal/terminal/session_test.go:L327)
- TestManagedSession_AC4_NaturalExitDetected() (internal/terminal/session_test.go:L340)
- TestManagedSession_AC5_FailedClientIsolation() (internal/terminal/session_test.go:L356)
- wsTestPair() (internal/terminal/session_test.go:L45)
- canPTYSpawn() (internal/terminal/session_test.go:L72)
- startTestSession() (internal/terminal/session_test.go:L85)

# Depends on
- [ManagedSession](/modules/managedsession.md)

# Inferred
- [ManagedSession](/modules/managedsession.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
