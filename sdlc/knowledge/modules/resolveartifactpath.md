---
type: Module
title: ResolveArtifactPath
description: "Graphify community 351: internal/bmad/artifacts.go, internal/bmad/artifacts_test.go, internal/bmad/executor.go, internal/bmad/executor_fileloader_test.go, internal/bmad/executor_interactive_test.go, i"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 0fcf7b7f19ac0972 }
  - { id: artifacts_test, resource: internal/bmad/artifacts_test.go, last_modified: "2026-04-08T20:52:15+10:00", digest: d50e7198f73eb5fe }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_fileloader_test, resource: internal/bmad/executor_fileloader_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 5c7c8dfad64509c9 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: testutil_interactive_test, resource: internal/bmad/testutil_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: f165a17bd64570bc }
---

# Files
- `internal/bmad/artifacts.go`
- `internal/bmad/artifacts_test.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_fileloader_test.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/testutil_interactive_test.go`

# Symbols
- ResolveArtifactPath() (internal/bmad/artifacts.go:L44)
- TestAC2_ResolveArtifactPath() (internal/bmad/artifacts_test.go:L39)
- resolveOutputPaths() (internal/bmad/executor.go:L1660)
- resolvedInputs (internal/bmad/executor.go:L2358)
- .resolveInputs() (internal/bmad/executor.go:L2649)
- firstDirectPredecessor() (internal/bmad/executor.go:L2749)
- truncate() (internal/bmad/executor.go:L2764)
- envValue() (internal/bmad/executor.go:L2773)
- buildInteractivePrompt() (internal/bmad/executor.go:L2919)
- appendUpstreamContext() (internal/bmad/executor.go:L2948)
- TestBuildInteractivePrompt_NoUpstream_SkipsBlock() (internal/bmad/executor_fileloader_test.go:L226)
- TestBuildInteractivePrompt_NilState_Compatible() (internal/bmad/executor_fileloader_test.go:L235)
- TestBuildInteractivePrompt_TruncatesLargeUpstream() (internal/bmad/executor_fileloader_test.go:L246)
- TestTruncateCap() (internal/bmad/executor_interactive_test.go:L752)
- TestAC2_UnmappedArtifacts_SkippedFromOutputPaths() (internal/bmad/executor_outputpaths_test.go:L83)
- .writeArtifact() (internal/bmad/testutil_interactive_test.go:L262)

# Depends on
- [Executor](/modules/executor.md)
- [executor_respond_test.go](/modules/executor-respond-test-go.md)

# Inferred
- [ProcessByID](/modules/processbyid.md)

# Features
- no feature plan names these files
