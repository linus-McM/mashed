---
type: Module
title: Executor
description: "Graphify community 17: internal/bmad/executor.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
---

# Files
- `internal/bmad/executor.go`

# Symbols
- .skipBranchLocked() (internal/bmad/executor.go:L1116)
- .skipNode() (internal/bmad/executor.go:L1141)
- .waitWhilePaused() (internal/bmad/executor.go:L1149)
- .captureOutput() (internal/bmad/executor.go:L1176)
- .waitForIdleCompletion() (internal/bmad/executor.go:L1215)
- .spawnCommandSession() (internal/bmad/executor.go:L1309)
- .resolveCommandSession() (internal/bmad/executor.go:L1368)
- .emit() (internal/bmad/executor.go:L138)
- .injectSlashCommand() (internal/bmad/executor.go:L1441)
- .executeCommandNode() (internal/bmad/executor.go:L1455)
- .StartWorkflow() (internal/bmad/executor.go:L152)
- .executeProcessNode() (internal/bmad/executor.go:L1524)
- .completeNode() (internal/bmad/executor.go:L1597)
- .failNode() (internal/bmad/executor.go:L1675)
- .pollNodeSignals() (internal/bmad/executor.go:L1703)
- .pollForQuestion() (internal/bmad/executor.go:L1718)
- .pollForQuestionFromCapture() (internal/bmad/executor.go:L1730)
- .pollForIdle() (internal/bmad/executor.go:L1782)
- .executeTransformNode() (internal/bmad/executor.go:L1834)
- .executeMultiFileLoader() (internal/bmad/executor.go:L1891)
- .executeFileLoader() (internal/bmad/executor.go:L2021)
- .recordNodeError() (internal/bmad/executor.go:L2107)
- .monitorSessionLiveness() (internal/bmad/executor.go:L229)
- .executeInteractiveNode() (internal/bmad/executor.go:L2382)
- .setStatus() (internal/bmad/executor.go:L2596)
- .sendToSession() (internal/bmad/executor.go:L2609)
- .captureRoundOutput() (internal/bmad/executor.go:L2633)
- .PauseWorkflow() (internal/bmad/executor.go:L286)
- .verifyOutputs() (internal/bmad/executor.go:L3017)
- .ResumeWorkflow() (internal/bmad/executor.go:L304)
- .StopWorkflow() (internal/bmad/executor.go:L322)
- .killWorkflowChainTails() (internal/bmad/executor.go:L340)
- .GetExecution() (internal/bmad/executor.go:L365)
- .GetInteractiveTranscript() (internal/bmad/executor.go:L394)
- .GetCurrentExecution() (internal/bmad/executor.go:L478)
- cloneExecution() (internal/bmad/executor.go:L514)
- execState (internal/bmad/executor.go:L54)
- .RespondToQuestionLegacy() (internal/bmad/executor.go:L548)
- .getState() (internal/bmad/executor.go:L615)
- .runDynamic() (internal/bmad/executor.go:L628)
- .activeOutEdges() (internal/bmad/executor.go:L818)
- Executor (internal/bmad/executor.go:L82)
- .executeControlNode() (internal/bmad/executor.go:L854)
- .executeLoopNode() (internal/bmad/executor.go:L916)

# Depends on
- [App](/modules/app.md)
- [bmad/types.go](/modules/bmad-types-go.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [executor_fileloader_test.go](/modules/executor-fileloader-test-go.md)
- [extractLines](/modules/extractlines.md)
- [ProcessByID](/modules/processbyid.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)
- [.resumeInteractiveNode](/modules/resumeinteractivenode.md)
- [Storage](/modules/storage.md)
- [topoSort](/modules/toposort.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)
- [wait_idle_test.go](/modules/wait-idle-test-go.md)

# Inferred
- [gate.go](/modules/gate-go.md)
- [go_pkg_testing](/modules/go-pkg-testing.md)
- [ProcessByID](/modules/processbyid.md)
- [question_test.go](/modules/question-test-go.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)
- [resume.go](/modules/resume-go.md)
- [session_naming_test.go](/modules/session-naming-test-go.md)
- [sprint_test.go](/modules/sprint-test-go.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)
- [testing.T](/modules/testing-t.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
