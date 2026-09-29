---
type: Module
title: executor_fileloader_test.go
description: "Graphify community 49: internal/bmad/executor.go, internal/bmad/executor_fileloader_test.go, internal/bmad/executor_multifileloader_test.go, internal/bmad/prompts.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_fileloader_test, resource: internal/bmad/executor_fileloader_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 5c7c8dfad64509c9 }
  - { id: executor_multifileloader_test, resource: internal/bmad/executor_multifileloader_test.go, last_modified: "2026-04-14T19:25:01+10:00", digest: 32a16dc3fc140e25 }
  - { id: prompts, resource: internal/bmad/prompts.go, last_modified: "2026-04-27T13:46:23+10:00", digest: b76c227c938b8250 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_fileloader_test.go`
- `internal/bmad/executor_multifileloader_test.go`
- `internal/bmad/prompts.go`

# Symbols
- buildInteractivePrompt() (internal/bmad/executor.go:L2919)
- executor_fileloader_test.go (internal/bmad/executor_fileloader_test.go:L1)
- TestFileLoader_PathTraversal_Rejected() (internal/bmad/executor_fileloader_test.go:L103)
- TestFileLoader_MissingFile_Fails() (internal/bmad/executor_fileloader_test.go:L124)
- TestFileLoader_EmptyFilePath_Fails() (internal/bmad/executor_fileloader_test.go:L139)
- TestFileLoader_Directory_Fails() (internal/bmad/executor_fileloader_test.go:L153)
- TestFileLoader_LargeFile_Truncated() (internal/bmad/executor_fileloader_test.go:L169)
- newStateWithUpstream() (internal/bmad/executor_fileloader_test.go:L192)
- TestBuildInteractivePrompt_IncludesUpstreamContext() (internal/bmad/executor_fileloader_test.go:L212)
- TestBuildInteractivePrompt_NoUpstream_SkipsBlock() (internal/bmad/executor_fileloader_test.go:L226)
- saveFileLoaderWorkflow() (internal/bmad/executor_fileloader_test.go:L23)
- TestBuildInteractivePrompt_NilState_Compatible() (internal/bmad/executor_fileloader_test.go:L235)
- TestBuildInteractivePrompt_TruncatesLargeUpstream() (internal/bmad/executor_fileloader_test.go:L246)
- TestExtractModalQuestion_Sentinel() (internal/bmad/executor_fileloader_test.go:L270)
- TestExtractModalQuestion_SentinelWins_OverTailFallback() (internal/bmad/executor_fileloader_test.go:L275)
- TestExtractModalQuestion_TailFallback_Truncates() (internal/bmad/executor_fileloader_test.go:L282)
- TestExtractModalQuestion_EmptyPassThrough() (internal/bmad/executor_fileloader_test.go:L289)
- TestExtractModalQuestion_UnclosedSentinel_FallsBackToTail() (internal/bmad/executor_fileloader_test.go:L294)
- TestFileLoader_HappyPath_PopulatesOutputsAndPaths() (internal/bmad/executor_fileloader_test.go:L44)
- TestFileLoader_RelativePath_ResolvedAgainstRepo() (internal/bmad/executor_fileloader_test.go:L88)
- TestMultiFileLoader_AC3_MissingFileTolerated() (internal/bmad/executor_multifileloader_test.go:L122)
- TestMultiFileLoader_AC4_DuplicateLabelsRejected() (internal/bmad/executor_multifileloader_test.go:L153)
- TestMultiFileLoader_TooManyEntries_Rejected() (internal/bmad/executor_multifileloader_test.go:L184)
- TestMultiFileLoader_RelativePath_ResolvedAgainstRepo() (internal/bmad/executor_multifileloader_test.go:L209)
- saveMultiFileWorkflow() (internal/bmad/executor_multifileloader_test.go:L25)
- waitForTerminal() (internal/bmad/executor_multifileloader_test.go:L52)
- TestMultiFileLoader_AC2_LabeledAndPositional_EmitsOutputPaths() (internal/bmad/executor_multifileloader_test.go:L84)
- extractModalQuestion() (internal/bmad/prompts.go:L75)

# Depends on
- [appendUpstreamContext](/modules/appendupstreamcontext.md)
- [Executor](/modules/executor.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [newHarness](/modules/newharness.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)

# Inferred
- [newHarness](/modules/newharness.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
