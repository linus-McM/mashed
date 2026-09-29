---
type: Module
title: newHarness
description: "Graphify community 1: app_bmad.go, internal/bmad/artifacts.go, internal/bmad/cleanup_test.go, internal/bmad/executor.go, internal/bmad/executor_cleanup_test.go, internal/bmad/executor_command_test.go,"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 0fcf7b7f19ac0972 }
  - { id: cleanup_test, resource: internal/bmad/cleanup_test.go, last_modified: "2026-05-07T09:52:03+10:00", digest: 322fffc24b5dcaad }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_cleanup_test, resource: internal/bmad/executor_cleanup_test.go, last_modified: "2026-04-12T15:59:00+10:00", digest: b58becad5281bc85 }
  - { id: executor_command_test, resource: internal/bmad/executor_command_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: bc808e46119528f7 }
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: executor_session_test, resource: internal/bmad/executor_session_test.go, last_modified: "2026-04-12T15:24:29+10:00", digest: e8e25bf75a4a8c21 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
  - { id: mock_helpers_test, resource: internal/bmad/mock_helpers_test.go, last_modified: "2026-04-12T16:42:48+10:00", digest: dd7b5edf17e36713 }
---

# Files
- `app_bmad.go`
- `internal/bmad/artifacts.go`
- `internal/bmad/cleanup_test.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_cleanup_test.go`
- `internal/bmad/executor_command_test.go`
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/executor_session_test.go`
- `internal/bmad/executor_test.go`
- `internal/bmad/mock_helpers_test.go`

# Symbols
- .GetArtifactStatus() (app_bmad.go:L463)
- GetArtifactStatus() (internal/bmad/artifacts.go:L59)
- cleanup_test.go (internal/bmad/cleanup_test.go:L1)
- TestExecutor_CleanupStaleSessions_AC2_PreservesTrackedSessions() (internal/bmad/cleanup_test.go:L111)
- TestExecutor_CleanupStaleSessions_AC3_SwallowsNoServerRunning() (internal/bmad/cleanup_test.go:L142)
- TestExecutor_CleanupStaleSessions_AC3_SwallowsErrorConnectingTo() (internal/bmad/cleanup_test.go:L166)
- TestExecutor_CleanupStaleSessions_AC4_ContinuesAfterIndividualKillFailure() (internal/bmad/cleanup_test.go:L191)
- cleanupRunner() (internal/bmad/cleanup_test.go:L25)
- killSessionTargets() (internal/bmad/cleanup_test.go:L61)
- TestExecutor_CleanupStaleSessions_AC1_KillsOrphanedBmadSessions() (internal/bmad/cleanup_test.go:L80)
- extractRegex() (internal/bmad/executor.go:L2118)
- extractLines() (internal/bmad/executor.go:L2135)
- TestKillWorkflowChainTails_NoTargets() (internal/bmad/executor_cleanup_test.go:L146)
- TestKillWorkflowChainTails_Dedup() (internal/bmad/executor_cleanup_test.go:L21)
- TestKillWorkflowChainTails_OnComplete() (internal/bmad/executor_cleanup_test.go:L55)
- TestKillWorkflowChainTails_OnFailed() (internal/bmad/executor_cleanup_test.go:L91)
- TestInjectSlashCommand_Argv() (internal/bmad/executor_command_test.go:L299)
- TestAC3_MissingFile_SkippedFromOutputPaths() (internal/bmad/executor_outputpaths_test.go:L101)
- TestAC4_StartWorkflow_ClearsPriorOutputPaths() (internal/bmad/executor_outputpaths_test.go:L144)
- TestAC1_OutputPathsPopulated_OnNodeComplete() (internal/bmad/executor_outputpaths_test.go:L28)
- capturedArgv (internal/bmad/executor_session_test.go:L18)
- .record() (internal/bmad/executor_session_test.go:L23)
- .last() (internal/bmad/executor_session_test.go:L31)
- TestSpawnCommandSession_RegressionFromRefactor() (internal/bmad/executor_session_test.go:L81)
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
- TestDynamicExecutor_AllBranchesSkipped() (internal/bmad/executor_test.go:L1357)
- TestDynamicExecutor_SkippedStatus() (internal/bmad/executor_test.go:L1407)
- TestExtractRegex_WithCaptureGroup() (internal/bmad/executor_test.go:L1462)
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
- TestTransformNode_MissingSource() (internal/bmad/executor_test.go:L1670)
- TestTransformNode_Passthrough() (internal/bmad/executor_test.go:L1706)
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
- cmdCall (internal/bmad/executor_test.go:L2729)
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
- successRunner() (internal/bmad/mock_helpers_test.go:L35)
- failRunner() (internal/bmad/mock_helpers_test.go:L49)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [ProcessByID](/modules/processbyid.md)
- [session_naming_test.go](/modules/session-naming-test-go.md)
- [Storage](/modules/storage.md)

# Inferred
- [NewExecutor](/modules/newexecutor.md)
- [session_naming_test.go](/modules/session-naming-test-go.md)
- [sprint_test.go](/modules/sprint-test-go.md)
- [storage_test.go](/modules/storage-test-go.md)

# Features
- no feature plan names these files
