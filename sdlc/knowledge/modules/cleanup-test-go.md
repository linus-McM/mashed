---
type: Module
title: cleanup_test.go
description: "Graphify community 235: internal/bmad/cleanup_test.go, internal/bmad/executor_cleanup_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: cleanup_test, resource: internal/bmad/cleanup_test.go, last_modified: "2026-05-07T09:52:03+10:00", digest: 322fffc24b5dcaad }
  - { id: executor_cleanup_test, resource: internal/bmad/executor_cleanup_test.go, last_modified: "2026-04-12T15:59:00+10:00", digest: b58becad5281bc85 }
---

# Files
- `internal/bmad/cleanup_test.go`
- `internal/bmad/executor_cleanup_test.go`

# Symbols
- cleanup_test.go (internal/bmad/cleanup_test.go:L1)
- TestExecutor_CleanupStaleSessions_AC2_PreservesTrackedSessions() (internal/bmad/cleanup_test.go:L111)
- TestExecutor_CleanupStaleSessions_AC3_SwallowsNoServerRunning() (internal/bmad/cleanup_test.go:L142)
- TestExecutor_CleanupStaleSessions_AC3_SwallowsErrorConnectingTo() (internal/bmad/cleanup_test.go:L166)
- TestExecutor_CleanupStaleSessions_AC4_ContinuesAfterIndividualKillFailure() (internal/bmad/cleanup_test.go:L191)
- cleanupRunner() (internal/bmad/cleanup_test.go:L25)
- killSessionTargets() (internal/bmad/cleanup_test.go:L61)
- TestExecutor_CleanupStaleSessions_AC1_KillsOrphanedBmadSessions() (internal/bmad/cleanup_test.go:L80)
- executor_cleanup_test.go (internal/bmad/executor_cleanup_test.go:L1)
- TestKillWorkflowChainTails_NoTargets() (internal/bmad/executor_cleanup_test.go:L146)
- TestKillWorkflowChainTails_Dedup() (internal/bmad/executor_cleanup_test.go:L21)
- TestKillWorkflowChainTails_OnComplete() (internal/bmad/executor_cleanup_test.go:L55)
- TestKillWorkflowChainTails_OnFailed() (internal/bmad/executor_cleanup_test.go:L91)

# Depends on
- [newHarness](/modules/newharness.md)

# Inferred
- [newHarness](/modules/newharness.md)
- [storage_test.go](/modules/storage-test-go.md)

# Features
- no feature plan names these files
