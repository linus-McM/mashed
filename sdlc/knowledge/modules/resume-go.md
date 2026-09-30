---
type: Module
title: resume.go
description: "Graphify community 289: internal/bmad/cleanup.go, internal/bmad/resume.go, internal/bmad/resume_ghost_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: cleanup, resource: internal/bmad/cleanup.go, last_modified: "2026-05-07T09:52:03+10:00", digest: 78b71ea7a045636f }
  - { id: resume, resource: internal/bmad/resume.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 0a2ce1f02fac5f5b }
  - { id: resume_ghost_test, resource: internal/bmad/resume_ghost_test.go, last_modified: "2026-04-23T14:17:26+10:00", digest: 81f5876fa72cade4 }
---

# Files
- `internal/bmad/cleanup.go`
- `internal/bmad/resume.go`
- `internal/bmad/resume_ghost_test.go`

# Symbols
- .CleanupStaleSessions() (internal/bmad/cleanup.go:L21)
- Executor (internal/bmad/cleanup.go:L21)
- isNoTmuxServer() (internal/bmad/cleanup.go:L61)
- .liveSessionNames() (internal/bmad/cleanup.go:L69)
- bareSessionName() (internal/bmad/cleanup.go:L94)
- resume.go (internal/bmad/resume.go:L1)
- LoadExecutionFromDisk() (internal/bmad/resume.go:L290)
- executionTmuxAlive() (internal/bmad/resume.go:L350)
- tmuxSessionAlive() (internal/bmad/resume.go:L372)
- markExecutionFailedOnDisk() (internal/bmad/resume.go:L386)
- TestLoadExecutionFromDisk_FiltersGhostTmuxSessions() (internal/bmad/resume_ghost_test.go:L17)
- TestLoadExecutionFromDisk_PreservesExecutionWithoutTmuxTargets() (internal/bmad/resume_ghost_test.go:L56)
- TestTmuxSessionAlive_DeadSessionReturnsFalse() (internal/bmad/resume_ghost_test.go:L82)

# Depends on
- [.resumeInteractiveNode](/modules/resumeinteractivenode.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
