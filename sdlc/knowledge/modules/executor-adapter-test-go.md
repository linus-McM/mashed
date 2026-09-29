---
type: Module
title: executor_adapter_test.go
description: "Graphify community 37: internal/bmad/executor.go, internal/bmad/executor_adapter_test.go, internal/bmad/executor_respond_test.go, internal/bmad/executor_suspend_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_adapter_test, resource: internal/bmad/executor_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f79102c9f400d28f }
  - { id: executor_respond_test, resource: internal/bmad/executor_respond_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 23fdebc82a94468b }
  - { id: executor_suspend_test, resource: internal/bmad/executor_suspend_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: b5f7b6a982bf834b }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_adapter_test.go`
- `internal/bmad/executor_respond_test.go`
- `internal/bmad/executor_suspend_test.go`

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
- TestRespondToInputConcurrencyNoPanic() (internal/bmad/executor_respond_test.go:L509)
- TestSuspendForSpecEntersAwaitingInput() (internal/bmad/executor_suspend_test.go:L110)
- TestSuspendForSpecCtxCancellationAborts() (internal/bmad/executor_suspend_test.go:L190)
- containsSubstr() (internal/bmad/executor_suspend_test.go:L259)
- newSuspendState() (internal/bmad/executor_suspend_test.go:L29)
- pollForStatus() (internal/bmad/executor_suspend_test.go:L87)

# Depends on
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [executor_respond_test.go](/modules/executor-respond-test-go.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [NewMock](/modules/newmock.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [NewExecutor](/modules/newexecutor.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
