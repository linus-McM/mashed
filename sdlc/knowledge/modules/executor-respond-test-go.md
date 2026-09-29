---
type: Module
title: executor_respond_test.go
description: "Graphify community 92: internal/bmad/executor.go, internal/bmad/executor_interactive_test.go, internal/bmad/executor_respond_test.go, internal/bmad/registry_interactive_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_respond_test, resource: internal/bmad/executor_respond_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 23fdebc82a94468b }
  - { id: registry_interactive_test, resource: internal/bmad/registry_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 2394d3c33e43a481 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_respond_test.go`
- `internal/bmad/registry_interactive_test.go`

# Symbols
- registryLookup() (internal/bmad/executor.go:L2787)
- loadRegistryCSV() (internal/bmad/executor.go:L2870)
- TestRegistryLookupRejectsNonRegistryScheme() (internal/bmad/executor_interactive_test.go:L660)
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
- TestRegistryLookupRejectsSchemes() (internal/bmad/executor_respond_test.go:L472)
- .waitForAwaiting() (internal/bmad/executor_respond_test.go:L94)
- registry_interactive_test.go (internal/bmad/registry_interactive_test.go:L1)
- TestInteractiveRegistryIterationInput() (internal/bmad/registry_interactive_test.go:L110)
- TestOptionsRefResolution() (internal/bmad/registry_interactive_test.go:L138)
- TestInteractiveRegistryShape() (internal/bmad/registry_interactive_test.go:L18)

# Depends on
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)

# Inferred
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [ProcessByID](/modules/processbyid.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Features
- no feature plan names these files
