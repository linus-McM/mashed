---
type: Module
title: executor_iteration_test.go
description: "Graphify community 64: internal/bmad/executor.go, internal/bmad/executor_anyuseranswer_test.go, internal/bmad/executor_flatten_test.go, internal/bmad/executor_interactive_test.go, internal/bmad/execut"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_anyuseranswer_test, resource: internal/bmad/executor_anyuseranswer_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: c5b7c0a594991958 }
  - { id: executor_flatten_test, resource: internal/bmad/executor_flatten_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 620dfc16866d8fff }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_iteration_test, resource: internal/bmad/executor_iteration_test.go, last_modified: "2026-04-20T14:19:27+10:00", digest: 57d32c63707a4432 }
  - { id: executor_suspend_test, resource: internal/bmad/executor_suspend_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: b5f7b6a982bf834b }
  - { id: types, resource: internal/bmad/types.go, last_modified: "2026-04-28T09:29:49+10:00", digest: 05a3e7d01f5f08c2 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_anyuseranswer_test.go`
- `internal/bmad/executor_flatten_test.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_iteration_test.go`
- `internal/bmad/executor_suspend_test.go`
- `internal/bmad/types.go`

# Symbols
- NodeStatusEvent (internal/bmad/executor.go:L39)
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
- WorkflowNodeStatus (internal/bmad/types.go:L124)

# Depends on
- [go_pkg_log_slog](/modules/go-pkg-log-slog.md)
- [NewExecutor](/modules/newexecutor.md)
- [question_test.go](/modules/question-test-go.md)
- [Storage](/modules/storage.md)

# Inferred
- [executor_respond_test.go](/modules/executor-respond-test-go.md)
- [newHarness](/modules/newharness.md)
- [registerTestProcess](/modules/registertestprocess.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
