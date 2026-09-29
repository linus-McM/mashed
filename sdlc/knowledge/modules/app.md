---
type: Module
title: App
description: "Graphify community 72: app_bmad.go, internal/bmad/executor.go, internal/bmad/storage.go, internal/bmad/types.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: storage, resource: internal/bmad/storage.go, last_modified: "2026-04-10T12:50:28+10:00", digest: 360ebf80068d1480 }
  - { id: types, resource: internal/bmad/types.go, last_modified: "2026-04-28T09:29:49+10:00", digest: 05a3e7d01f5f08c2 }
---

# Files
- `app_bmad.go`
- `internal/bmad/executor.go`
- `internal/bmad/storage.go`
- `internal/bmad/types.go`

# Symbols
- .GetSprintStatus() (app_bmad.go:L136)
- .UpdateStoryStatus() (app_bmad.go:L141)
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
- .ListAllAgents() (app_bmad.go:L333)
- .ListBmadWorkflows() (app_bmad.go:L34)
- .GetBmadWorkflow() (app_bmad.go:L42)
- .SaveMashedAssetFrontmatter() (app_bmad.go:L436)
- .GetNodeOutput() (app_bmad.go:L450)
- .GetArtifactStatus() (app_bmad.go:L463)
- .SaveBmadWorkflow() (app_bmad.go:L50)
- .DeleteBmadWorkflow() (app_bmad.go:L58)
- .ListBmadWorkflowsByRepo() (app_bmad.go:L78)
- .ListBmadTemplates() (app_bmad.go:L88)
- .CreateFromTemplate() (app_bmad.go:L93)
- InteractiveTurn (internal/bmad/executor.go:L380)
- ListClaudeAgents() (internal/bmad/storage.go:L197)
- WorkflowEdge (internal/bmad/types.go:L309)
- WorkflowDef (internal/bmad/types.go:L318)

# Depends on
- [assets_test.go](/modules/assets-test-go.md)
- [bmad/registry_test.go](/modules/bmad-registry-test-go.md)
- [bmad/types.go](/modules/bmad-types-go.md)
- [BuiltinTemplates](/modules/builtintemplates.md)
- [GetModules](/modules/getmodules.md)
- [LoadExecutionFromDisk](/modules/loadexecutionfromdisk.md)
- [ProcessByID](/modules/processbyid.md)
- [ProcessDef](/modules/processdef.md)
- [sprint.go](/modules/sprint-go.md)
- [sprint_test.go](/modules/sprint-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
