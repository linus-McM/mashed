---
type: Module
title: executor_command_test.go
description: "Graphify community 83: internal/bmad/executor_command_test.go, internal/bmad/executor_interactive_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: executor_command_test, resource: internal/bmad/executor_command_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: bc808e46119528f7 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
---

# Files
- `internal/bmad/executor_command_test.go`
- `internal/bmad/executor_interactive_test.go`

# Symbols
- executor_command_test.go (internal/bmad/executor_command_test.go:L1)
- tmuxCallRecord (internal/bmad/executor_command_test.go:L110)
- callTracker (internal/bmad/executor_command_test.go:L116)
- .record() (internal/bmad/executor_command_test.go:L121)
- .snapshot() (internal/bmad/executor_command_test.go:L127)
- idleCycleTrackingRunner() (internal/bmad/executor_command_test.go:L142)
- saveChainedDAGWorkflow() (internal/bmad/executor_command_test.go:L178)
- saveTwoNodeDAGWorkflow() (internal/bmad/executor_command_test.go:L202)
- TestStory1_AC1_NodeTypeCommandRoundTrip() (internal/bmad/executor_command_test.go:L227)
- countingNewSessionRunner() (internal/bmad/executor_command_test.go:L25)
- TestStory1_AC3_ProcessNodeUnchanged() (internal/bmad/executor_command_test.go:L285)
- TestInjectSlashCommand_Argv() (internal/bmad/executor_command_test.go:L299)
- TestExecuteCommandNode_ChainedDAG() (internal/bmad/executor_command_test.go:L346)
- TestExecuteCommandNode_NoParentSpawns() (internal/bmad/executor_command_test.go:L395)
- saveSingleNodeWorkflow() (internal/bmad/executor_command_test.go:L40)
- TestExecuteCommandNode_MissingCommandName() (internal/bmad/executor_command_test.go:L441)
- TestExecuteCommandNode_SpawnFails() (internal/bmad/executor_command_test.go:L472)
- TestExecuteCommandNode_DispatcherReplacesFailFast() (internal/bmad/executor_command_test.go:L495)
- TestExecuteCommandNode_InjectionOrdering() (internal/bmad/executor_command_test.go:L520)
- waitForNodeStatus() (internal/bmad/executor_command_test.go:L71)
- startSingleNodeAndWait() (internal/bmad/executor_command_test.go:L93)
- TestRoutingDispatchesAutonomousNodesToExecuteNode() (internal/bmad/executor_interactive_test.go:L137)
- TestRoutingDispatchesInteractiveModesToExecuteInteractiveNode() (internal/bmad/executor_interactive_test.go:L178)
- TestExecuteInteractiveNodeHappyPath() (internal/bmad/executor_interactive_test.go:L266)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [newHarness](/modules/newharness.md)
- [Storage](/modules/storage.md)

# Inferred
- [newHarness](/modules/newharness.md)

# Features
- no feature plan names these files
