---
type: Module
title: resume_ghost_test.go
description: "Graphify community 233: internal/bmad/cleanup.go, internal/bmad/resume.go, internal/bmad/resume_ghost_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
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
- bareSessionName() (internal/bmad/cleanup.go:L94)
- LoadExecutionFromDisk() (internal/bmad/resume.go:L290)
- executionTmuxAlive() (internal/bmad/resume.go:L350)
- tmuxSessionAlive() (internal/bmad/resume.go:L372)
- markExecutionFailedOnDisk() (internal/bmad/resume.go:L386)
- resume_ghost_test.go (internal/bmad/resume_ghost_test.go:L1)
- TestLoadExecutionFromDisk_FiltersGhostTmuxSessions() (internal/bmad/resume_ghost_test.go:L17)
- TestLoadExecutionFromDisk_PreservesExecutionWithoutTmuxTargets() (internal/bmad/resume_ghost_test.go:L56)
- TestTmuxSessionAlive_DeadSessionReturnsFalse() (internal/bmad/resume_ghost_test.go:L82)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
