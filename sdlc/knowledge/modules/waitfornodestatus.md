---
type: Module
title: waitForNodeStatus
description: "Graphify community 69: internal/bmad/executor_command_test.go, internal/bmad/executor_interactive_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: executor_command_test, resource: internal/bmad/executor_command_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: bc808e46119528f7 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
---

# Files
- `internal/bmad/executor_command_test.go`
- `internal/bmad/executor_interactive_test.go`

# Symbols
- tmuxCallRecord (internal/bmad/executor_command_test.go:L110)
- callTracker (internal/bmad/executor_command_test.go:L116)
- .record() (internal/bmad/executor_command_test.go:L121)
- .snapshot() (internal/bmad/executor_command_test.go:L127)
- idleCycleTrackingRunner() (internal/bmad/executor_command_test.go:L142)
- countingNewSessionRunner() (internal/bmad/executor_command_test.go:L25)
- TestStory1_AC3_ProcessNodeUnchanged() (internal/bmad/executor_command_test.go:L285)
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
- saveDownstreamWorkflow() (internal/bmad/executor_interactive_test.go:L94)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [newHarness](/modules/newharness.md)
- [Storage](/modules/storage.md)

# Inferred
- [newHarness](/modules/newharness.md)

# Features
- no feature plan names these files
