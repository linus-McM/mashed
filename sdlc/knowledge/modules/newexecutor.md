---
type: Module
title: NewExecutor
description: "Graphify community 36: internal/bmad/executor.go, internal/bmad/executor_gate_test.go, internal/bmad/executor_interactive_test.go, internal/bmad/executor_iteration_test.go, internal/bmad/executor_resu"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_gate_test, resource: internal/bmad/executor_gate_test.go, last_modified: "2026-04-20T14:19:27+10:00", digest: 9071d4934689a274 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_iteration_test, resource: internal/bmad/executor_iteration_test.go, last_modified: "2026-04-20T14:19:27+10:00", digest: 57d32c63707a4432 }
  - { id: executor_resume_test, resource: internal/bmad/executor_resume_test.go, last_modified: "2026-04-23T13:17:59+10:00", digest: 326e436b95533512 }
  - { id: executor_session_test, resource: internal/bmad/executor_session_test.go, last_modified: "2026-04-12T15:24:29+10:00", digest: e8e25bf75a4a8c21 }
  - { id: storage, resource: internal/bmad/storage.go, last_modified: "2026-04-10T12:50:28+10:00", digest: 360ebf80068d1480 }
  - { id: storage_test, resource: internal/bmad/storage_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: cbabd36eb06877ea }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_gate_test.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_iteration_test.go`
- `internal/bmad/executor_resume_test.go`
- `internal/bmad/executor_session_test.go`
- `internal/bmad/storage.go`
- `internal/bmad/storage_test.go`

# Symbols
- NewExecutor() (internal/bmad/executor.go:L106)
- executor_gate_test.go (internal/bmad/executor_gate_test.go:L1)
- TestCheckGateRoundLimit() (internal/bmad/executor_gate_test.go:L115)
- TestCheckGateArtifactExists() (internal/bmad/executor_gate_test.go:L189)
- TestCheckGateUserConfirmAcceptToken() (internal/bmad/executor_gate_test.go:L26)
- TestCheckGateExpression() (internal/bmad/executor_gate_test.go:L289)
- TestCheckGateNil() (internal/bmad/executor_gate_test.go:L340)
- TestContainsTokenCaseInsensitiveTrim() (internal/bmad/executor_gate_test.go:L358)
- TestIterationInputHelper() (internal/bmad/executor_gate_test.go:L445)
- TestResolveInputsTable() (internal/bmad/executor_interactive_test.go:L312)
- TestVerifyOutputsRequiredFileMissing() (internal/bmad/executor_interactive_test.go:L502)
- TestVerifyOutputsMemoryTargetAlwaysOK() (internal/bmad/executor_interactive_test.go:L583)
- TestActiveOutEdgesLogsOnNonCompleteNode() (internal/bmad/executor_interactive_test.go:L613)
- TestSendToSessionEscapesControlChars() (internal/bmad/executor_iteration_test.go:L698)
- TestFinalRoundMirrorsNodeOutputs() (internal/bmad/executor_iteration_test.go:L764)
- executor_resume_test.go (internal/bmad/executor_resume_test.go:L1)
- TestPersistSnapshotConcurrentCallsNoPanic() (internal/bmad/executor_resume_test.go:L203)
- TestPaneAliveTrueFalse() (internal/bmad/executor_resume_test.go:L239)
- TestResumeInteractiveNodeDeadPaneRecap() (internal/bmad/executor_resume_test.go:L310)
- TestPersistSnapshotWritesInteractiveFields() (internal/bmad/executor_resume_test.go:L32)
- TestRehydratePendingSpawnsWaiters() (internal/bmad/executor_resume_test.go:L487)
- TestStopWhileAwaitingRestoreProducesAbortedState() (internal/bmad/executor_resume_test.go:L577)
- TestRehydrateFromSnapshot_RegistersExecForRespond() (internal/bmad/executor_resume_test.go:L631)
- TestRehydrateFromSnapshot_IdempotentOnSecondCall() (internal/bmad/executor_resume_test.go:L709)
- TestGetInteractiveTranscript_OrdersClaudeThenUserPerRound() (internal/bmad/executor_resume_test.go:L733)
- TestGetInteractiveTranscript_EmptyNodeReturnsEmptySlice() (internal/bmad/executor_resume_test.go:L785)
- TestGetInteractiveTranscript_SkipsMissingClaudeRoundButKeepsUser() (internal/bmad/executor_resume_test.go:L809)
- TestGetInteractiveTranscript_ExecNotFound() (internal/bmad/executor_resume_test.go:L840)
- TestSnapshotVersionField() (internal/bmad/executor_resume_test.go:L98)
- TestResolveCommandSession_ZeroLiveParents() (internal/bmad/executor_session_test.go:L141)
- TestResolveCommandSession_OneLiveParent() (internal/bmad/executor_session_test.go:L167)
- TestResolveCommandSession_StaleParent() (internal/bmad/executor_session_test.go:L197)
- TestResolveCommandSession_MultiParentPicksMostRecent() (internal/bmad/executor_session_test.go:L226)
- TestResolveCommandSession_SessionGoneError() (internal/bmad/executor_session_test.go:L262)
- newSessionState() (internal/bmad/executor_session_test.go:L42)
- NewStorage() (internal/bmad/storage.go:L24)
- TestNewStorage_CreatesDirectories() (internal/bmad/storage_test.go:L221)

# Depends on
- [Executor](/modules/executor.md)
- [Storage](/modules/storage.md)

# Inferred
- [.executeInteractiveNode](/modules/executeinteractivenode.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)

# Features
- no feature plan names these files
