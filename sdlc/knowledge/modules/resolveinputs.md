---
type: Module
title: .resolveInputs
description: "Graphify community 351: internal/bmad/executor.go, internal/bmad/executor_fileloader_test.go, internal/bmad/executor_interactive_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_fileloader_test, resource: internal/bmad/executor_fileloader_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 5c7c8dfad64509c9 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_fileloader_test.go`
- `internal/bmad/executor_interactive_test.go`

# Symbols
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

# Depends on
- [Executor](/modules/executor.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Inferred
- [ProcessByID](/modules/processbyid.md)

# Features
- no feature plan names these files
