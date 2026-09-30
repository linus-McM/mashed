---
type: Module
title: executor_test.go
description: "Graphify community 1: app_bmad.go, internal/bmad/artifacts.go, internal/bmad/executor.go, internal/bmad/executor_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 0fcf7b7f19ac0972 }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
---

# Files
- `app_bmad.go`
- `internal/bmad/artifacts.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_test.go`

# Symbols
- .GetArtifactStatus() (app_bmad.go:L463)
- GetArtifactStatus() (internal/bmad/artifacts.go:L59)
- extractRegex() (internal/bmad/executor.go:L2118)
- extractLines() (internal/bmad/executor.go:L2135)
- topoSort() (internal/bmad/executor.go:L2312)
- executor_test.go (internal/bmad/executor_test.go:L1)
- seedExecState() (internal/bmad/executor_test.go:L1029)
- TestGetCurrentExecution_EmptyRepoPath() (internal/bmad/executor_test.go:L1055)
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
- TestTopoSort_Sequential() (internal/bmad/executor_test.go:L133)
- TestDynamicExecutor_AllBranchesSkipped() (internal/bmad/executor_test.go:L1357)
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
- TestTransformNode_Passthrough() (internal/bmad/executor_test.go:L1706)
- TestTopoSort_Diamond() (internal/bmad/executor_test.go:L174)
- loopRunner() (internal/bmad/executor_test.go:L2068)
- TestLoopNode_FixedCount() (internal/bmad/executor_test.go:L2092)
- TestLoopUntil_ConditionMet() (internal/bmad/executor_test.go:L2152)
- TestLoopUntil_MaxIterations() (internal/bmad/executor_test.go:L2201)
- TestLoop_EmptyBody() (internal/bmad/executor_test.go:L2322)
- TestGetArtifactStatus_Exists() (internal/bmad/executor_test.go:L2506)
- TestGetArtifactStatus_Missing() (internal/bmad/executor_test.go:L2517)
- TestGetArtifactStatus_UnmappedArtifact() (internal/bmad/executor_test.go:L2525)
- TestLoopNode_IteratesOverItems() (internal/bmad/executor_test.go:L2535)
- TestLoopNode_ItemsCappedByMaxIterations() (internal/bmad/executor_test.go:L2590)
- TestLoopNode_InvalidItemsJSON() (internal/bmad/executor_test.go:L2634)
- TestWorkflowNode_StoryID_Serialization() (internal/bmad/executor_test.go:L454)
- TestWorkflowNode_StoryID_BackwardCompat() (internal/bmad/executor_test.go:L503)
- captureRunner() (internal/bmad/executor_test.go:L663)
- TestCaptureOutput_ParallelNodes() (internal/bmad/executor_test.go:L802)
- runExecuteNodeSessionCase() (internal/bmad/executor_test.go:L852)
- TestExecuteNode_AC7_UsesDescriptiveName() (internal/bmad/executor_test.go:L904)
- TestExecuteNode_AC8_BranchLookupFailureFallsBackToDetached() (internal/bmad/executor_test.go:L922)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [cleanup_test.go](/modules/cleanup-test-go.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [newHarness](/modules/newharness.md)
- [ProcessByID](/modules/processbyid.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)
- [seedResponseState](/modules/seedresponsestate.md)
- [sync.Mutex](/modules/sync-mutex.md)

# Inferred
- [newHarness](/modules/newharness.md)
- [session_naming_test.go](/modules/session-naming-test-go.md)

# Features
- no feature plan names these files
