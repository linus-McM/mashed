---
type: Module
title: executor_iteration_test.go
description: "Graphify community 56: internal/bmad/executor.go, internal/bmad/executor_adapter_test.go, internal/bmad/executor_anyuseranswer_test.go, internal/bmad/executor_flatten_test.go, internal/bmad/executor_i"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_adapter_test, resource: internal/bmad/executor_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f79102c9f400d28f }
  - { id: executor_anyuseranswer_test, resource: internal/bmad/executor_anyuseranswer_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: c5b7c0a594991958 }
  - { id: executor_flatten_test, resource: internal/bmad/executor_flatten_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 620dfc16866d8fff }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_iteration_test, resource: internal/bmad/executor_iteration_test.go, last_modified: "2026-04-20T14:19:27+10:00", digest: 57d32c63707a4432 }
  - { id: executor_respond_test, resource: internal/bmad/executor_respond_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 23fdebc82a94468b }
  - { id: executor_suspend_test, resource: internal/bmad/executor_suspend_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: b5f7b6a982bf834b }
  - { id: logging_plumbing_mock_test, resource: internal/uiadapter/logging_plumbing_mock_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: bce0a1e9685e2603 }
  - { id: mock, resource: internal/uiadapter/mock.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ee77ed81915c20b0 }
  - { id: mock_test, resource: internal/uiadapter/mock_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: e6b7941070a1bd3a }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_adapter_test.go`
- `internal/bmad/executor_anyuseranswer_test.go`
- `internal/bmad/executor_flatten_test.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_iteration_test.go`
- `internal/bmad/executor_respond_test.go`
- `internal/bmad/executor_suspend_test.go`
- `internal/uiadapter/logging_plumbing_mock_test.go`
- `internal/uiadapter/mock.go`
- `internal/uiadapter/mock_test.go`

# Symbols
- WithAdapter() (internal/bmad/executor.go:L100)
- Option (internal/bmad/executor.go:L95)
- executor_adapter_test.go (internal/bmad/executor_adapter_test.go:L1)
- TestWithAdapter_OptionAppliesAdapter() (internal/bmad/executor_adapter_test.go:L149)
- registerU4Process() (internal/bmad/executor_adapter_test.go:L215)
- TestSuspendForSpec_TranslateOutsideLock() (internal/bmad/executor_adapter_test.go:L247)
- TestSuspendForSpec_CancelDuringTranslate_Returns() (internal/bmad/executor_adapter_test.go:L331)
- TestSuspendForSpec_PartyModeRound1UserTurn_HasStructured() (internal/bmad/executor_adapter_test.go:L379)
- TestSuspendForSpec_OversizeBlobDropped() (internal/bmad/executor_adapter_test.go:L434)
- newOversizeAST() (internal/bmad/executor_adapter_test.go:L57)
- drainAfter() (internal/bmad/executor_adapter_test.go:L70)
- TestExecutor_NilAdapter_ShortCircuits() (internal/bmad/executor_adapter_test.go:L89)
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
- TestRespondToInputNoPendingPrompt() (internal/bmad/executor_respond_test.go:L358)
- TestRespondToInputConcurrencyNoPanic() (internal/bmad/executor_respond_test.go:L509)
- executor_suspend_test.go (internal/bmad/executor_suspend_test.go:L1)
- TestSuspendForSpecEntersAwaitingInput() (internal/bmad/executor_suspend_test.go:L110)
- TestSuspendForSpecCtxCancellationAborts() (internal/bmad/executor_suspend_test.go:L190)
- containsSubstr() (internal/bmad/executor_suspend_test.go:L259)
- newSuspendState() (internal/bmad/executor_suspend_test.go:L29)
- hookEvents() (internal/bmad/executor_suspend_test.go:L46)
- eventsNamed() (internal/bmad/executor_suspend_test.go:L67)
- registerTestProcess() (internal/bmad/executor_suspend_test.go:L79)
- pollForStatus() (internal/bmad/executor_suspend_test.go:L87)
- TestStory2_AC2_AC1_NewMockAcceptsNilLogger() (internal/uiadapter/logging_plumbing_mock_test.go:L19)
- mock.go (internal/uiadapter/mock.go:L1)
- MockAdapter (internal/uiadapter/mock.go:L19)
- NewMock() (internal/uiadapter/mock.go:L31)
- fixtureName() (internal/uiadapter/mock.go:L45)
- .Translate() (internal/uiadapter/mock.go:L58)
- TestU2_AC8_MockAdapter_FixedReturn() (internal/uiadapter/mock_test.go:L19)
- TestU2_AC8_MockAdapter_NilReturnsFallback() (internal/uiadapter/mock_test.go:L33)

# Depends on
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [NewExecutor](/modules/newexecutor.md)
- [Storage](/modules/storage.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [NewExecutor](/modules/newexecutor.md)
- [newHarness](/modules/newharness.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Features
- no feature plan names these files
