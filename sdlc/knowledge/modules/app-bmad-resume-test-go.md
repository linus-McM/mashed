---
type: Module
title: app_bmad_resume_test.go
description: "Graphify community 392: app_bmad_question_test.go, app_bmad_respond_input_test.go, app_bmad_resume_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app_bmad_question_test, resource: app_bmad_question_test.go, last_modified: "2026-04-10T14:01:16+10:00", digest: 9386c29f8c1272f8 }
  - { id: app_bmad_respond_input_test, resource: app_bmad_respond_input_test.go, last_modified: "2026-04-20T13:50:13+10:00", digest: c31156ae45a58fd1 }
  - { id: app_bmad_resume_test, resource: app_bmad_resume_test.go, last_modified: "2026-04-20T14:41:10+10:00", digest: 3d0c6fc332298628 }
---

# Files
- `app_bmad_question_test.go`
- `app_bmad_respond_input_test.go`
- `app_bmad_resume_test.go`

# Symbols
- app_bmad_question_test.go (app_bmad_question_test.go:L1)
- newBmadTestApp() (app_bmad_question_test.go:L25)
- TestStory2_AC5_AppRespondToQuestion_Delegates() (app_bmad_question_test.go:L46)
- app_bmad_respond_input_test.go (app_bmad_respond_input_test.go:L1)
- TestAppRespondToInputIntegration() (app_bmad_respond_input_test.go:L101)
- TestAppRespondToInputProxies() (app_bmad_respond_input_test.go:L28)
- TestAppRespondToQuestionLegacyShimStillWorks() (app_bmad_respond_input_test.go:L80)
- app_bmad_resume_test.go (app_bmad_resume_test.go:L1)
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
- [testing.T](/modules/testing-t.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
