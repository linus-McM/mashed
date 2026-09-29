---
type: Module
title: newHarness
description: "Graphify community 1: internal/bmad/executor_outputpaths_test.go, internal/bmad/executor_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
---

# Files
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/executor_test.go`

# Symbols
- TestAC4_StartWorkflow_ClearsPriorOutputPaths() (internal/bmad/executor_outputpaths_test.go:L144)
- executor_test.go (internal/bmad/executor_test.go:L1)
- seedExecState() (internal/bmad/executor_test.go:L1029)
- TestGetCurrentExecution_EmptyRepoPath() (internal/bmad/executor_test.go:L1055)
- TestGetCurrentExecution_NoExecutions() (internal/bmad/executor_test.go:L1064)
- TestGetCurrentExecution_NoMatchingRepo() (internal/bmad/executor_test.go:L1072)
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
- TestTransformNode_RegexExtraction() (internal/bmad/executor_test.go:L1579)
- TestTransformNode_LinesExtraction() (internal/bmad/executor_test.go:L1631)
- TestTransformNode_Passthrough() (internal/bmad/executor_test.go:L1706)
- loopRunner() (internal/bmad/executor_test.go:L2068)
- TestLoopNode_FixedCount() (internal/bmad/executor_test.go:L2092)
- TestLoopUntil_ConditionMet() (internal/bmad/executor_test.go:L2152)
- TestLoopUntil_MaxIterations() (internal/bmad/executor_test.go:L2201)
- TestLoop_BodyFailure() (internal/bmad/executor_test.go:L2250)
- TestLoop_EmptyBody() (internal/bmad/executor_test.go:L2322)
- TestStartWorkflow_CyclicWorkflow() (internal/bmad/executor_test.go:L248)
- TestLoopNode_IteratesOverItems() (internal/bmad/executor_test.go:L2535)
- TestLoopNode_ItemsCappedByMaxIterations() (internal/bmad/executor_test.go:L2590)
- TestLoopNode_InvalidItemsJSON() (internal/bmad/executor_test.go:L2634)
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
- newHarness() (internal/bmad/executor_test.go:L34)
- TestGetExecution_NotFound() (internal/bmad/executor_test.go:L438)
- TestStartWorkflow_WorkflowNotFound() (internal/bmad/executor_test.go:L444)
- TestWorkflowNode_StoryID_Serialization() (internal/bmad/executor_test.go:L454)
- TestWorkflowNode_StoryID_BackwardCompat() (internal/bmad/executor_test.go:L503)
- TestCaptureOutput_ParallelNodes() (internal/bmad/executor_test.go:L802)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [extractLines](/modules/extractlines.md)
- [ParseSessionName](/modules/parsesessionname.md)
- [ProcessByID](/modules/processbyid.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)
- [successRunner](/modules/successrunner.md)
- [topoSort](/modules/toposort.md)

# Inferred
- [NewExecutor](/modules/newexecutor.md)
- [ParseSessionName](/modules/parsesessionname.md)
- [successRunner](/modules/successrunner.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
