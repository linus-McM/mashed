---
type: Module
title: newHarness
description: "Graphify community 1: internal/bmad/artifacts.go, internal/bmad/executor.go, internal/bmad/executor_outputpaths_test.go, internal/bmad/executor_test.go, internal/bmad/mock_helpers_test.go, internal/bm"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 0fcf7b7f19ac0972 }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
  - { id: mock_helpers_test, resource: internal/bmad/mock_helpers_test.go, last_modified: "2026-04-12T16:42:48+10:00", digest: dd7b5edf17e36713 }
  - { id: wait_idle_test, resource: internal/bmad/wait_idle_test.go, last_modified: "2026-04-23T13:05:14+10:00", digest: 1dda9adef91d1698 }
---

# Files
- `internal/bmad/artifacts.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/executor_test.go`
- `internal/bmad/mock_helpers_test.go`
- `internal/bmad/wait_idle_test.go`

# Symbols
- GetArtifactStatus() (internal/bmad/artifacts.go:L59)
- .SetCommandRunner() (internal/bmad/executor.go:L123)
- extractRegex() (internal/bmad/executor.go:L2118)
- extractLines() (internal/bmad/executor.go:L2135)
- topoSort() (internal/bmad/executor.go:L2312)
- CommandRunner (internal/bmad/executor.go:L31)
- TestAC3_MissingFile_SkippedFromOutputPaths() (internal/bmad/executor_outputpaths_test.go:L101)
- TestAC4_StartWorkflow_ClearsPriorOutputPaths() (internal/bmad/executor_outputpaths_test.go:L144)
- TestAC5_ConcurrentCompletion_NoRace() (internal/bmad/executor_outputpaths_test.go:L188)
- cloneStrMap() (internal/bmad/executor_outputpaths_test.go:L238)
- TestAC1_OutputPathsPopulated_OnNodeComplete() (internal/bmad/executor_outputpaths_test.go:L28)
- executor_test.go (internal/bmad/executor_test.go:L1)
- seedExecState() (internal/bmad/executor_test.go:L1029)
- TestGetCurrentExecution_EmptyRepoPath() (internal/bmad/executor_test.go:L1055)
- TestGetCurrentExecution_NoExecutions() (internal/bmad/executor_test.go:L1064)
- TestGetCurrentExecution_NoMatchingRepo() (internal/bmad/executor_test.go:L1072)
- saveThreeNodeWorkflow() (internal/bmad/executor_test.go:L108)
- TestGetCurrentExecution_SingleRunningMatch() (internal/bmad/executor_test.go:L1082)
- TestGetCurrentExecution_PausedCountsAsNonTerminal() (internal/bmad/executor_test.go:L1093)
- TestGetCurrentExecution_IgnoresTerminalExecutions() (internal/bmad/executor_test.go:L1103)
- TestGetCurrentExecution_PicksLatestByStartedAt() (internal/bmad/executor_test.go:L1114)
- TestGetCurrentExecution_ReturnsDeepCopy() (internal/bmad/executor_test.go:L1127)
- conditionRunner() (internal/bmad/executor_test.go:L1153)
- sessionLabelFromArgs() (internal/bmad/executor_test.go:L1175)
- TestDynamicExecutor_ConditionBranching_TrueBranch() (internal/bmad/executor_test.go:L1187)
- TestDynamicExecutor_ConditionBranching_FalseBranch() (internal/bmad/executor_test.go:L1244)
- TestDynamicExecutor_MergeAfterCondition() (internal/bmad/executor_test.go:L1299)
- TestTopoSort_Sequential() (internal/bmad/executor_test.go:L133)
- TestDynamicExecutor_AllBranchesSkipped() (internal/bmad/executor_test.go:L1357)
- TestDynamicExecutor_SkippedStatus() (internal/bmad/executor_test.go:L1407)
- TestExtractRegex_WithCaptureGroup() (internal/bmad/executor_test.go:L1462)
- TestTopoSort_Parallel() (internal/bmad/executor_test.go:L149)
- TestExtractRegex_WithoutCaptureGroup() (internal/bmad/executor_test.go:L1490)
- TestExtractRegex_NoMatch() (internal/bmad/executor_test.go:L1518)
- TestExtractRegex_InvalidRegex() (internal/bmad/executor_test.go:L1523)
- TestExtractLines_Range() (internal/bmad/executor_test.go:L1530)
- TestExtractLines_LastN() (internal/bmad/executor_test.go:L1536)
- TestExtractLines_Single() (internal/bmad/executor_test.go:L1542)
- TestExtractLines_OutOfRange() (internal/bmad/executor_test.go:L1548)
- TestExtractLines_InvalidPattern() (internal/bmad/executor_test.go:L1554)
- TestExtractLines_StartBeyondLength() (internal/bmad/executor_test.go:L1571)
- TestTransformNode_RegexExtraction() (internal/bmad/executor_test.go:L1579)
- TestTransformNode_LinesExtraction() (internal/bmad/executor_test.go:L1631)
- TestTopoSort_Cycle() (internal/bmad/executor_test.go:L164)
- TestTransformNode_MissingSource() (internal/bmad/executor_test.go:L1670)
- TestTransformNode_Passthrough() (internal/bmad/executor_test.go:L1706)
- TestTopoSort_Diamond() (internal/bmad/executor_test.go:L174)
- TestStartWorkflow_Sequential() (internal/bmad/executor_test.go:L197)
- loopRunner() (internal/bmad/executor_test.go:L2068)
- TestLoopNode_FixedCount() (internal/bmad/executor_test.go:L2092)
- TestLoopUntil_ConditionMet() (internal/bmad/executor_test.go:L2152)
- eventRecord (internal/bmad/executor_test.go:L22)
- TestLoopUntil_MaxIterations() (internal/bmad/executor_test.go:L2201)
- TestLoop_BodyFailure() (internal/bmad/executor_test.go:L2250)
- TestStartWorkflow_NodeFailure() (internal/bmad/executor_test.go:L226)
- TestLoop_EmptyBody() (internal/bmad/executor_test.go:L2322)
- TestCompleteNode_EmitsArtifactEvent_ProcessNode() (internal/bmad/executor_test.go:L2375)
- TestCompleteNode_ArtifactEvent_MissingArtifact() (internal/bmad/executor_test.go:L2422)
- TestCompleteNode_NoArtifactEvent_ControlNode() (internal/bmad/executor_test.go:L2460)
- TestStartWorkflow_CyclicWorkflow() (internal/bmad/executor_test.go:L248)
- TestGetArtifactStatus_Exists() (internal/bmad/executor_test.go:L2506)
- TestGetArtifactStatus_Missing() (internal/bmad/executor_test.go:L2517)
- TestGetArtifactStatus_UnmappedArtifact() (internal/bmad/executor_test.go:L2525)
- TestLoopNode_IteratesOverItems() (internal/bmad/executor_test.go:L2535)
- TestLoopNode_ItemsCappedByMaxIterations() (internal/bmad/executor_test.go:L2590)
- TestLoopNode_InvalidItemsJSON() (internal/bmad/executor_test.go:L2634)
- TestLoopNode_EmptyItems() (internal/bmad/executor_test.go:L2679)
- testHarness (internal/bmad/executor_test.go:L27)
- TestStartWorkflow_ParallelBranches() (internal/bmad/executor_test.go:L272)
- responseRunner() (internal/bmad/executor_test.go:L2737)
- findCall() (internal/bmad/executor_test.go:L2757)
- seedResponseState() (internal/bmad/executor_test.go:L2780)
- TestStory2_AC1_RespondToQuestion_Success() (internal/bmad/executor_test.go:L2805)
- TestStory2_AC1_RespondToQuestion_MenuOption() (internal/bmad/executor_test.go:L2832)
- TestStory2_AC2_RespondToQuestion_DeadPane() (internal/bmad/executor_test.go:L2848)
- TestStory2_AC2_RespondToQuestion_PaneCheckFails() (internal/bmad/executor_test.go:L2866)
- TestStory2_AC4_RespondToQuestion_ClearsHash() (internal/bmad/executor_test.go:L2883)
- TestStory2_AC4_RespondToQuestion_HashPreservedOnFailure() (internal/bmad/executor_test.go:L2898)
- TestStory2_RespondToQuestion_UnknownExec() (internal/bmad/executor_test.go:L2913)
- TestStory2_RespondToQuestion_UnknownNode() (internal/bmad/executor_test.go:L2921)
- TestStory2_RespondToQuestion_LongAnswer() (internal/bmad/executor_test.go:L2933)
- TestStory2_RespondToQuestion_NoTmuxTarget() (internal/bmad/executor_test.go:L2953)
- TestStory2_RespondToQuestion_SendKeysLiteralFails() (internal/bmad/executor_test.go:L2970)
- TestStory2_RespondToQuestion_SendKeysEnterFails() (internal/bmad/executor_test.go:L2989)
- TestStory2_RespondToQuestion_EmptyAnswer() (internal/bmad/executor_test.go:L3014)
- TestPauseAndResume() (internal/bmad/executor_test.go:L328)
- newHarness() (internal/bmad/executor_test.go:L34)
- TestStopWorkflow() (internal/bmad/executor_test.go:L373)
- TestPauseWorkflow_NotRunning() (internal/bmad/executor_test.go:L405)
- TestResumeWorkflow_NotPaused() (internal/bmad/executor_test.go:L423)
- TestGetExecution_NotFound() (internal/bmad/executor_test.go:L438)
- TestStartWorkflow_WorkflowNotFound() (internal/bmad/executor_test.go:L444)
- TestWorkflowNode_StoryID_Serialization() (internal/bmad/executor_test.go:L454)
- .getEvents() (internal/bmad/executor_test.go:L49)
- TestWorkflowNode_StoryID_BackwardCompat() (internal/bmad/executor_test.go:L503)
- TestCompleteNode_WithStoryID_AdvancesStory() (internal/bmad/executor_test.go:L514)
- TestCompleteNode_WithoutStoryID_NoSprintEvent() (internal/bmad/executor_test.go:L566)
- .eventsByName() (internal/bmad/executor_test.go:L57)
- TestFailNode_WithStoryID_NoSprintUpdate() (internal/bmad/executor_test.go:L584)
- createSprintYAMLForExec() (internal/bmad/executor_test.go:L634)
- captureRunner() (internal/bmad/executor_test.go:L663)
- delayRunner() (internal/bmad/executor_test.go:L68)
- TestCaptureOutput_StoresOnCompletion() (internal/bmad/executor_test.go:L683)
- TestCaptureOutput_100KBCap() (internal/bmad/executor_test.go:L721)
- TestCaptureOutput_FailureNonFatal() (internal/bmad/executor_test.go:L764)
- TestCaptureOutput_ParallelNodes() (internal/bmad/executor_test.go:L802)
- TestGetExecution_CopiesNodeOutputs() (internal/bmad/executor_test.go:L954)
- TestGetExecution_ReturnsCopy() (internal/bmad/executor_test.go:L998)
- makeIdleOutput() (internal/bmad/mock_helpers_test.go:L117)
- successRunner() (internal/bmad/mock_helpers_test.go:L35)
- failRunner() (internal/bmad/mock_helpers_test.go:L49)
- idleMockRunner() (internal/bmad/mock_helpers_test.go:L86)
- wait_idle_test.go (internal/bmad/wait_idle_test.go:L1)
- TestWaitForIdleCompletion_CtxCancel() (internal/bmad/wait_idle_test.go:L119)
- newWaitIdleState() (internal/bmad/wait_idle_test.go:L17)
- TestWaitForIdleCompletion_PaneDeathSessionGone() (internal/bmad/wait_idle_test.go:L184)
- TestWaitForIdleCompletion_IdlePromptRequiresStableHash() (internal/bmad/wait_idle_test.go:L212)
- TestWaitForIdleCompletion_HappyPath() (internal/bmad/wait_idle_test.go:L37)
- TestWaitForIdleCompletion_NoWork_Timeout() (internal/bmad/wait_idle_test.go:L65)
- TestWaitForIdleCompletion_PaneDeath() (internal/bmad/wait_idle_test.go:L81)

# Depends on
- [cleanup_test.go](/modules/cleanup-test-go.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [NodeType](/modules/nodetype.md)
- [ProcessByID](/modules/processbyid.md)
- [session_naming_test.go](/modules/session-naming-test-go.md)

# Inferred
- [NewExecutor](/modules/newexecutor.md)
- [session_naming_test.go](/modules/session-naming-test-go.md)
- [sprint_test.go](/modules/sprint-test-go.md)
- [testing.T](/modules/testing-t.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
