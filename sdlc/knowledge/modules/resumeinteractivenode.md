---
type: Module
title: .resumeInteractiveNode
description: "Graphify community 290: internal/bmad/executor.go, internal/bmad/executor_resume_test.go, internal/bmad/resume.go, internal/bmad/shell_quote_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_resume_test, resource: internal/bmad/executor_resume_test.go, last_modified: "2026-04-23T13:17:59+10:00", digest: 326e436b95533512 }
  - { id: resume, resource: internal/bmad/resume.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 0a2ce1f02fac5f5b }
  - { id: shell_quote_test, resource: internal/bmad/shell_quote_test.go, last_modified: "2026-04-12T12:32:31+10:00", digest: a1f5b92816b5a5c6 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_resume_test.go`
- `internal/bmad/resume.go`
- `internal/bmad/shell_quote_test.go`

# Symbols
- wrapInBashExec() (internal/bmad/executor.go:L1191)
- TestRenderRecap() (internal/bmad/executor_resume_test.go:L425)
- .RehydrateFromSnapshot() (internal/bmad/resume.go:L118)
- .paneAlive() (internal/bmad/resume.go:L18)
- Executor (internal/bmad/resume.go:L18)
- .rehydratePending() (internal/bmad/resume.go:L247)
- renderRecap() (internal/bmad/resume.go:L29)
- .resumeInteractiveNode() (internal/bmad/resume.go:L40)
- TestWrapInBashExec() (internal/bmad/shell_quote_test.go:L9)

# Depends on
- [ProcessByID](/modules/processbyid.md)

# Inferred
- [executor_fileloader_test.go](/modules/executor-fileloader-test-go.md)
- [ProcessByID](/modules/processbyid.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
