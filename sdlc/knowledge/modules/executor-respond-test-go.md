---
type: Module
title: executor_respond_test.go
description: "Graphify community 92: internal/bmad/executor_respond_test.go, internal/bmad/executor_suspend_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: executor_respond_test, resource: internal/bmad/executor_respond_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 23fdebc82a94468b }
  - { id: executor_suspend_test, resource: internal/bmad/executor_suspend_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: b5f7b6a982bf834b }
---

# Files
- `internal/bmad/executor_respond_test.go`
- `internal/bmad/executor_suspend_test.go`

# Symbols
- executor_respond_test.go (internal/bmad/executor_respond_test.go:L1)
- .nodeStatus() (internal/bmad/executor_respond_test.go:L101)
- .nodeInputs() (internal/bmad/executor_respond_test.go:L109)
- .nodeInputHistory() (internal/bmad/executor_respond_test.go:L124)
- .pendingPrompts() (internal/bmad/executor_respond_test.go:L137)
- TestRespondToInputHappyPath() (internal/bmad/executor_respond_test.go:L153)
- TestRespondToInputInvalidChoiceKeepsAwaiting() (internal/bmad/executor_respond_test.go:L219)
- TestRespondToInputShapeFilePathTraversalRejected() (internal/bmad/executor_respond_test.go:L267)
- suspendHarness (internal/bmad/executor_respond_test.go:L29)
- TestRespondToInputValueHashOnlyInEvent() (internal/bmad/executor_respond_test.go:L312)
- sha256hex16() (internal/bmad/executor_respond_test.go:L349)
- TestRespondToInputDoubleResponseSecondFails() (internal/bmad/executor_respond_test.go:L393)
- TestRespondToInputUnknownInputID() (internal/bmad/executor_respond_test.go:L424)
- setupSuspension() (internal/bmad/executor_respond_test.go:L45)
- TestRespondToInputExecNotFound() (internal/bmad/executor_respond_test.go:L457)
- .waitForAwaiting() (internal/bmad/executor_respond_test.go:L94)
- eventsNamed() (internal/bmad/executor_suspend_test.go:L67)

# Depends on
- [executor_adapter_test.go](/modules/executor-adapter-test-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Inferred
- [.executeInteractiveNode](/modules/executeinteractivenode.md)
- [executor_adapter_test.go](/modules/executor-adapter-test-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Features
- no feature plan names these files
