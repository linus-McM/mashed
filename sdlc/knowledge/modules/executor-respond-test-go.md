---
type: Module
title: executor_respond_test.go
description: "Graphify community 92: internal/bmad/executor_respond_test.go, internal/bmad/executor_suspend_test.go, internal/bmad/prompts.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: executor_respond_test, resource: internal/bmad/executor_respond_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 23fdebc82a94468b }
  - { id: executor_suspend_test, resource: internal/bmad/executor_suspend_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: b5f7b6a982bf834b }
  - { id: prompts, resource: internal/bmad/prompts.go, last_modified: "2026-04-27T13:46:23+10:00", digest: b76c227c938b8250 }
---

# Files
- `internal/bmad/executor_respond_test.go`
- `internal/bmad/executor_suspend_test.go`
- `internal/bmad/prompts.go`

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
- TestHashPendingPromptDeterminism() (internal/bmad/executor_suspend_test.go:L276)
- eventsNamed() (internal/bmad/executor_suspend_test.go:L67)
- hashPendingPrompt() (internal/bmad/prompts.go:L132)

# Depends on
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [go_pkg_strings](/modules/go-pkg-strings.md)
- [registerTestProcess](/modules/registertestprocess.md)

# Inferred
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [registerTestProcess](/modules/registertestprocess.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Features
- no feature plan names these files
