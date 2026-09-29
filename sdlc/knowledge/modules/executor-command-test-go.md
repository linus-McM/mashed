---
type: Module
title: executor_command_test.go
description: "Graphify community 69: internal/bmad/executor.go, internal/bmad/executor_command_test.go, internal/bmad/executor_interactive_test.go, internal/bmad/interactive_types_test.go, internal/bmad/storage.go,"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_command_test, resource: internal/bmad/executor_command_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: bc808e46119528f7 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: interactive_types_test, resource: internal/bmad/interactive_types_test.go, last_modified: "2026-04-20T13:01:41+10:00", digest: 0282e550b6b93893 }
  - { id: storage, resource: internal/bmad/storage.go, last_modified: "2026-04-10T12:50:28+10:00", digest: 360ebf80068d1480 }
  - { id: types, resource: internal/bmad/types.go, last_modified: "2026-04-28T09:29:49+10:00", digest: 05a3e7d01f5f08c2 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_command_test.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/interactive_types_test.go`
- `internal/bmad/storage.go`
- `internal/bmad/types.go`

# Symbols
- NodeStatusEvent (internal/bmad/executor.go:L39)
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
- saveDownstreamWorkflow() (internal/bmad/executor_interactive_test.go:L94)
- TestConstantJSONValues() (internal/bmad/interactive_types_test.go:L460)
- .ListWorkflowsByRepo() (internal/bmad/storage.go:L106)
- .DeleteWorkflow() (internal/bmad/storage.go:L122)
- .SaveAgent() (internal/bmad/storage.go:L139)
- .ListAgents() (internal/bmad/storage.go:L149)
- Storage (internal/bmad/storage.go:L17)
- .DeleteAgent() (internal/bmad/storage.go:L178)
- atomicWriteJSON() (internal/bmad/storage.go:L225)
- validateID() (internal/bmad/storage.go:L36)
- .SaveWorkflow() (internal/bmad/storage.go:L44)
- .LoadWorkflow() (internal/bmad/storage.go:L54)
- .ListWorkflows() (internal/bmad/storage.go:L76)
- WorkflowNodeStatus (internal/bmad/types.go:L124)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [newHarness](/modules/newharness.md)
- [ProcessDef](/modules/processdef.md)

# Inferred
- [newHarness](/modules/newharness.md)

# Features
- no feature plan names these files
