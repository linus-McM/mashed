---
type: Module
title: .resumeInteractiveNode
description: "Graphify community 350: internal/bmad/executor.go, internal/bmad/resume.go, internal/bmad/shell_quote_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: resume, resource: internal/bmad/resume.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 0a2ce1f02fac5f5b }
  - { id: shell_quote_test, resource: internal/bmad/shell_quote_test.go, last_modified: "2026-04-12T12:32:31+10:00", digest: a1f5b92816b5a5c6 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/resume.go`
- `internal/bmad/shell_quote_test.go`

# Symbols
- wrapInBashExec() (internal/bmad/executor.go:L1191)
- .RehydrateFromSnapshot() (internal/bmad/resume.go:L118)
- .RefreshPendingStructured() (internal/bmad/resume.go:L167)
- .paneAlive() (internal/bmad/resume.go:L18)
- Executor (internal/bmad/resume.go:L18)
- .rehydratePending() (internal/bmad/resume.go:L247)
- .resumeInteractiveNode() (internal/bmad/resume.go:L40)
- TestWrapInBashExec() (internal/bmad/shell_quote_test.go:L9)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [.executeInteractiveNode](/modules/executeinteractivenode.md)
- [ProcessByID](/modules/processbyid.md)
- [.resolveInputs](/modules/resolveinputs.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Features
- no feature plan names these files
