---
type: Module
title: executor_respond_test.go
description: "Graphify community 100: internal/bmad/executor_respond_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: executor_respond_test, resource: internal/bmad/executor_respond_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 23fdebc82a94468b }
---

# Files
- `internal/bmad/executor_respond_test.go`

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

# Depends on
- [executor_adapter_test.go](/modules/executor-adapter-test-go.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [registryLookup](/modules/registrylookup.md)

# Inferred
- [executor_adapter_test.go](/modules/executor-adapter-test-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Features
- no feature plan names these files
