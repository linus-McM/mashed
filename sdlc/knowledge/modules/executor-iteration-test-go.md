---
type: Module
title: executor_iteration_test.go
description: "Graphify community 56: internal/bmad/executor_anyuseranswer_test.go, internal/bmad/executor_flatten_test.go, internal/bmad/executor_interactive_test.go, internal/bmad/executor_iteration_test.go, inter"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: executor_anyuseranswer_test, resource: internal/bmad/executor_anyuseranswer_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: c5b7c0a594991958 }
  - { id: executor_flatten_test, resource: internal/bmad/executor_flatten_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 620dfc16866d8fff }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_iteration_test, resource: internal/bmad/executor_iteration_test.go, last_modified: "2026-04-20T14:19:27+10:00", digest: 57d32c63707a4432 }
  - { id: executor_suspend_test, resource: internal/bmad/executor_suspend_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: b5f7b6a982bf834b }
---

# Files
- `internal/bmad/executor_anyuseranswer_test.go`
- `internal/bmad/executor_flatten_test.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_iteration_test.go`
- `internal/bmad/executor_suspend_test.go`

# Symbols
- TestPartyMode_JSONSubmission_RejectTokenInFreeWidget() (internal/bmad/executor_anyuseranswer_test.go:L159)
- eventHasReason() (internal/bmad/executor_anyuseranswer_test.go:L23)
- TestPartyMode_JSONSubmission_AcceptTokenOnApprovalWidget() (internal/bmad/executor_anyuseranswer_test.go:L253)
- TestPartyMode_LegacyShapeFreeUnchanged() (internal/bmad/executor_flatten_test.go:L162)
- snapshotNodeState() (internal/bmad/executor_flatten_test.go:L24)
- TestFlatten_CompositeAndBareKeys() (internal/bmad/executor_flatten_test.go:L46)
- saveInteractiveWorkflow() (internal/bmad/executor_interactive_test.go:L68)
- executor_iteration_test.go (internal/bmad/executor_iteration_test.go:L1)
- feedAnswers() (internal/bmad/executor_iteration_test.go:L119)
- roundCaptureRunner() (internal/bmad/executor_iteration_test.go:L187)
- nodeStatusFromExec() (internal/bmad/executor_iteration_test.go:L234)
- pollForTerminal() (internal/bmad/executor_iteration_test.go:L250)
- TestRoundLoopGateUserConfirmAcceptsOnThirdRound() (internal/bmad/executor_iteration_test.go:L268)
- roundCompleteEvents() (internal/bmad/executor_iteration_test.go:L32)
- TestRoundLoopMaxRoundsExit() (internal/bmad/executor_iteration_test.go:L360)
- gateSatisfiedEvents() (internal/bmad/executor_iteration_test.go:L38)
- roundLimitEvents() (internal/bmad/executor_iteration_test.go:L44)
- TestRoundLoopRejectTokenAborts() (internal/bmad/executor_iteration_test.go:L470)
- abortedEvents() (internal/bmad/executor_iteration_test.go:L49)
- TestRoundLoopSingleRoundGuided() (internal/bmad/executor_iteration_test.go:L557)
- buildIterationProcess() (internal/bmad/executor_iteration_test.go:L56)
- buildGuidedProcess() (internal/bmad/executor_iteration_test.go:L88)
- hookEvents() (internal/bmad/executor_suspend_test.go:L46)

# Depends on
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [NewMock](/modules/newmock.md)
- [question_test.go](/modules/question-test-go.md)

# Inferred
- [executor_respond_test.go](/modules/executor-respond-test-go.md)
- [newHarness](/modules/newharness.md)
- [registerTestProcess](/modules/registertestprocess.md)

# Features
- no feature plan names these files
