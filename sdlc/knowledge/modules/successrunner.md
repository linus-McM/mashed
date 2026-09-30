---
type: Module
title: successRunner
description: "Graphify community 18: internal/bmad/executor_outputpaths_test.go, internal/bmad/executor_test.go, internal/bmad/mock_helpers_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
  - { id: mock_helpers_test, resource: internal/bmad/mock_helpers_test.go, last_modified: "2026-04-12T16:42:48+10:00", digest: dd7b5edf17e36713 }
---

# Files
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/executor_test.go`
- `internal/bmad/mock_helpers_test.go`

# Symbols
- TestAC3_MissingFile_SkippedFromOutputPaths() (internal/bmad/executor_outputpaths_test.go:L101)
- TestAC1_OutputPathsPopulated_OnNodeComplete() (internal/bmad/executor_outputpaths_test.go:L28)
- saveThreeNodeWorkflow() (internal/bmad/executor_test.go:L108)
- TestTransformNode_MissingSource() (internal/bmad/executor_test.go:L1670)
- TestStartWorkflow_Sequential() (internal/bmad/executor_test.go:L197)
- eventRecord (internal/bmad/executor_test.go:L22)
- TestStartWorkflow_NodeFailure() (internal/bmad/executor_test.go:L226)
- TestCompleteNode_EmitsArtifactEvent_ProcessNode() (internal/bmad/executor_test.go:L2375)
- TestCompleteNode_ArtifactEvent_MissingArtifact() (internal/bmad/executor_test.go:L2422)
- TestCompleteNode_NoArtifactEvent_ControlNode() (internal/bmad/executor_test.go:L2460)
- TestLoopNode_EmptyItems() (internal/bmad/executor_test.go:L2679)
- testHarness (internal/bmad/executor_test.go:L27)
- TestPauseAndResume() (internal/bmad/executor_test.go:L328)
- TestStopWorkflow() (internal/bmad/executor_test.go:L373)
- TestPauseWorkflow_NotRunning() (internal/bmad/executor_test.go:L405)
- TestResumeWorkflow_NotPaused() (internal/bmad/executor_test.go:L423)
- .getEvents() (internal/bmad/executor_test.go:L49)
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
- TestGetExecution_CopiesNodeOutputs() (internal/bmad/executor_test.go:L954)
- TestGetExecution_ReturnsCopy() (internal/bmad/executor_test.go:L998)
- successRunner() (internal/bmad/mock_helpers_test.go:L35)
- failRunner() (internal/bmad/mock_helpers_test.go:L49)

# Depends on
- [newHarness](/modules/newharness.md)
- [Storage](/modules/storage.md)

# Inferred
- [newHarness](/modules/newharness.md)
- [sprint_test.go](/modules/sprint-test-go.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
