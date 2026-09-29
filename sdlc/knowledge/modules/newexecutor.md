---
type: Module
title: NewExecutor
description: "Graphify community 23: internal/bmad/executor.go, internal/bmad/executor_gate_test.go, internal/bmad/executor_interactive_test.go, internal/bmad/executor_iteration_test.go, internal/bmad/executor_resu"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_gate_test, resource: internal/bmad/executor_gate_test.go, last_modified: "2026-04-20T14:19:27+10:00", digest: 9071d4934689a274 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_iteration_test, resource: internal/bmad/executor_iteration_test.go, last_modified: "2026-04-20T14:19:27+10:00", digest: 57d32c63707a4432 }
  - { id: executor_resume_test, resource: internal/bmad/executor_resume_test.go, last_modified: "2026-04-23T13:17:59+10:00", digest: 326e436b95533512 }
  - { id: executor_session_test, resource: internal/bmad/executor_session_test.go, last_modified: "2026-04-12T15:24:29+10:00", digest: e8e25bf75a4a8c21 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_gate_test.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_iteration_test.go`
- `internal/bmad/executor_resume_test.go`
- `internal/bmad/executor_session_test.go`

# Symbols
- NewExecutor() (internal/bmad/executor.go:L106)
- executor_gate_test.go (internal/bmad/executor_gate_test.go:L1)
- TestCheckGateRoundLimit() (internal/bmad/executor_gate_test.go:L115)
- TestCheckGateArtifactExists() (internal/bmad/executor_gate_test.go:L189)
- TestCheckGateUserConfirmAcceptToken() (internal/bmad/executor_gate_test.go:L26)
- TestCheckGateExpression() (internal/bmad/executor_gate_test.go:L289)
- TestCheckGateNil() (internal/bmad/executor_gate_test.go:L340)
- TestIterationInputHelper() (internal/bmad/executor_gate_test.go:L445)
- TestResolveInputsTable() (internal/bmad/executor_interactive_test.go:L312)
- TestVerifyOutputsRequiredFileMissing() (internal/bmad/executor_interactive_test.go:L502)
- TestVerifyOutputsMemoryTargetAlwaysOK() (internal/bmad/executor_interactive_test.go:L583)
- TestActiveOutEdgesLogsOnNonCompleteNode() (internal/bmad/executor_interactive_test.go:L613)
- TestSendToSessionTwoTmuxCalls() (internal/bmad/executor_iteration_test.go:L620)
- TestSendToSessionEscapesControlChars() (internal/bmad/executor_iteration_test.go:L698)
- TestFinalRoundMirrorsNodeOutputs() (internal/bmad/executor_iteration_test.go:L764)
- TestPersistSnapshotConcurrentCallsNoPanic() (internal/bmad/executor_resume_test.go:L203)
- TestPaneAliveTrueFalse() (internal/bmad/executor_resume_test.go:L239)
- TestResumeInteractiveNodeDeadPaneRecap() (internal/bmad/executor_resume_test.go:L310)
- TestPersistSnapshotWritesInteractiveFields() (internal/bmad/executor_resume_test.go:L32)
- TestRehydratePendingSpawnsWaiters() (internal/bmad/executor_resume_test.go:L487)
- TestStopWhileAwaitingRestoreProducesAbortedState() (internal/bmad/executor_resume_test.go:L577)
- TestSnapshotVersionField() (internal/bmad/executor_resume_test.go:L98)
- TestResolveCommandSession_ZeroLiveParents() (internal/bmad/executor_session_test.go:L141)
- TestResolveCommandSession_OneLiveParent() (internal/bmad/executor_session_test.go:L167)
- TestResolveCommandSession_StaleParent() (internal/bmad/executor_session_test.go:L197)
- TestResolveCommandSession_MultiParentPicksMostRecent() (internal/bmad/executor_session_test.go:L226)
- TestResolveCommandSession_SessionGoneError() (internal/bmad/executor_session_test.go:L262)
- newSessionState() (internal/bmad/executor_session_test.go:L42)

# Depends on
- [Executor](/modules/executor.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [gate.go](/modules/gate-go.md)

# Inferred
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [question_test.go](/modules/question-test-go.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
