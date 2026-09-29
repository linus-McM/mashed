---
type: Module
title: App
description: "Graphify community 76: app_bmad.go, internal/bmad/executor.go, internal/bmad/types.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: types, resource: internal/bmad/types.go, last_modified: "2026-04-28T09:29:49+10:00", digest: 05a3e7d01f5f08c2 }
---

# Files
- `app_bmad.go`
- `internal/bmad/executor.go`
- `internal/bmad/types.go`

# Symbols
- .GetSprintStatus() (app_bmad.go:L136)
- .StartBmadWorkflow() (app_bmad.go:L148)
- .PauseBmadWorkflow() (app_bmad.go:L160)
- .ResumeBmadWorkflow() (app_bmad.go:L168)
- .StopBmadWorkflow() (app_bmad.go:L176)
- .RespondToQuestion() (app_bmad.go:L186)
- .RespondToInput() (app_bmad.go:L197)
- .GetInteractiveTranscript() (app_bmad.go:L217)
- App (app_bmad.go:L23)
- .emitEvent() (app_bmad.go:L23)
- .GetBmadCurrentExecution() (app_bmad.go:L235)
- .DeleteBmadAgent() (app_bmad.go:L324)
- .ListBmadWorkflows() (app_bmad.go:L34)
- .GetBmadWorkflow() (app_bmad.go:L42)
- .SaveMashedAssetFrontmatter() (app_bmad.go:L436)
- .GetNodeOutput() (app_bmad.go:L450)
- .SaveBmadWorkflow() (app_bmad.go:L50)
- .DeleteBmadWorkflow() (app_bmad.go:L58)
- .ListBmadWorkflowsByRepo() (app_bmad.go:L78)
- .ListBmadTemplates() (app_bmad.go:L88)
- .CreateFromTemplate() (app_bmad.go:L93)
- InteractiveTurn (internal/bmad/executor.go:L380)
- WorkflowEdge (internal/bmad/types.go:L309)
- WorkflowDef (internal/bmad/types.go:L318)

# Depends on
- [assets_test.go](/modules/assets-test-go.md)
- [bmad/types.go](/modules/bmad-types-go.md)
- [BmadAgentConfig](/modules/bmadagentconfig.md)
- [BuiltinTemplates](/modules/builtintemplates.md)
- [GetModules](/modules/getmodules.md)
- [LoadExecutionFromDisk](/modules/loadexecutionfromdisk.md)
- [newHarness](/modules/newharness.md)
- [ProcessDef](/modules/processdef.md)
- [sprint.go](/modules/sprint-go.md)
- [sprint_test.go](/modules/sprint-test-go.md)
- [WorkflowExecution](/modules/workflowexecution.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
