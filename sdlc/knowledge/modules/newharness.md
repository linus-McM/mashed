---
type: Module
title: newHarness
description: "Graphify community 35: internal/bmad/executor_command_test.go, internal/bmad/executor_interactive_test.go, internal/bmad/executor_outputpaths_test.go, internal/bmad/executor_test.go, internal/bmad/moc"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: executor_command_test, resource: internal/bmad/executor_command_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: bc808e46119528f7 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
  - { id: mock_helpers_test, resource: internal/bmad/mock_helpers_test.go, last_modified: "2026-04-12T16:42:48+10:00", digest: dd7b5edf17e36713 }
---

# Files
- `internal/bmad/executor_command_test.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/executor_test.go`
- `internal/bmad/mock_helpers_test.go`

# Symbols
- TestInjectSlashCommand_Argv() (internal/bmad/executor_command_test.go:L299)
- TestRoutingDispatchesInteractiveModesToExecuteInteractiveNode() (internal/bmad/executor_interactive_test.go:L178)
- TestAC3_MissingFile_SkippedFromOutputPaths() (internal/bmad/executor_outputpaths_test.go:L101)
- TestAC4_StartWorkflow_ClearsPriorOutputPaths() (internal/bmad/executor_outputpaths_test.go:L144)
- TestAC1_OutputPathsPopulated_OnNodeComplete() (internal/bmad/executor_outputpaths_test.go:L28)
- TestGetCurrentExecution_NoExecutions() (internal/bmad/executor_test.go:L1064)
- saveThreeNodeWorkflow() (internal/bmad/executor_test.go:L108)
- TestDynamicExecutor_SkippedStatus() (internal/bmad/executor_test.go:L1407)
- TestTransformNode_MissingSource() (internal/bmad/executor_test.go:L1670)
- TestStartWorkflow_Sequential() (internal/bmad/executor_test.go:L197)
- TestLoop_BodyFailure() (internal/bmad/executor_test.go:L2250)
- TestStartWorkflow_NodeFailure() (internal/bmad/executor_test.go:L226)
- TestCompleteNode_EmitsArtifactEvent_ProcessNode() (internal/bmad/executor_test.go:L2375)
- TestCompleteNode_ArtifactEvent_MissingArtifact() (internal/bmad/executor_test.go:L2422)
- TestCompleteNode_NoArtifactEvent_ControlNode() (internal/bmad/executor_test.go:L2460)
- TestStartWorkflow_CyclicWorkflow() (internal/bmad/executor_test.go:L248)
- TestLoopNode_EmptyItems() (internal/bmad/executor_test.go:L2679)
- TestStartWorkflow_ParallelBranches() (internal/bmad/executor_test.go:L272)
- TestStory2_RespondToQuestion_UnknownExec() (internal/bmad/executor_test.go:L2913)
- TestPauseAndResume() (internal/bmad/executor_test.go:L328)
- newHarness() (internal/bmad/executor_test.go:L34)
- TestStopWorkflow() (internal/bmad/executor_test.go:L373)
- TestPauseWorkflow_NotRunning() (internal/bmad/executor_test.go:L405)
- TestResumeWorkflow_NotPaused() (internal/bmad/executor_test.go:L423)
- TestGetExecution_NotFound() (internal/bmad/executor_test.go:L438)
- TestStartWorkflow_WorkflowNotFound() (internal/bmad/executor_test.go:L444)
- TestCompleteNode_WithStoryID_AdvancesStory() (internal/bmad/executor_test.go:L514)
- TestCompleteNode_WithoutStoryID_NoSprintEvent() (internal/bmad/executor_test.go:L566)
- .eventsByName() (internal/bmad/executor_test.go:L57)
- TestFailNode_WithStoryID_NoSprintUpdate() (internal/bmad/executor_test.go:L584)
- createSprintYAMLForExec() (internal/bmad/executor_test.go:L634)
- delayRunner() (internal/bmad/executor_test.go:L68)
- TestCaptureOutput_StoresOnCompletion() (internal/bmad/executor_test.go:L683)
- TestCaptureOutput_100KBCap() (internal/bmad/executor_test.go:L721)
- TestCaptureOutput_FailureNonFatal() (internal/bmad/executor_test.go:L764)
- TestGetExecution_CopiesNodeOutputs() (internal/bmad/executor_test.go:L954)
- TestGetExecution_ReturnsCopy() (internal/bmad/executor_test.go:L998)
- successRunner() (internal/bmad/mock_helpers_test.go:L35)
- failRunner() (internal/bmad/mock_helpers_test.go:L49)

# Depends on
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [executor_test.go](/modules/executor-test-go.md)
- [Storage](/modules/storage.md)
- [sync.Mutex](/modules/sync-mutex.md)

# Inferred
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [sprint_test.go](/modules/sprint-test-go.md)

# Features
- no feature plan names these files
