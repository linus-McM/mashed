---
type: Module
title: newBmadTestAppWithExec
description: "Graphify community 392: app_bmad_resume_test.go"
resource: .
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app_bmad_resume_test, resource: app_bmad_resume_test.go, last_modified: "2026-04-20T14:41:10+10:00", digest: 3d0c6fc332298628 }
---

# Files
- `app_bmad_resume_test.go`

# Symbols
- seedSnapshotOnDisk() (app_bmad_resume_test.go:L111)
- TestGetBmadCurrentExecutionReEmitsAwaitingInput() (app_bmad_resume_test.go:L131)
- TestGetBmadCurrentExecutionNoEmitForEmptyPending() (app_bmad_resume_test.go:L229)
- TestGetBmadCurrentExecutionTerminalExecutionReturnsNil() (app_bmad_resume_test.go:L269)
- capturedAppEvent (app_bmad_resume_test.go:L61)
- hookAppEmit() (app_bmad_resume_test.go:L68)
- newBmadTestAppWithExec() (app_bmad_resume_test.go:L94)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [Executor](/modules/executor.md)
- [NewExecutor](/modules/newexecutor.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
