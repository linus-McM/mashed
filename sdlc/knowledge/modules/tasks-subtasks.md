---
type: Module
title: Tasks / Subtasks
description: "Graphify community 35: docs/SPECIFICATION.md, docs/stories/breadcrumbs-06-downstream-autofill.md, docs/stories/old_stories/bmad-05-wails-bindings.md, docs/stories/old_stories/bmad-06-workflow-view-can"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: breadcrumbs-06-downstream-autofill, resource: docs/stories/breadcrumbs-06-downstream-autofill.md, last_modified: "2026-04-14T16:10:56+10:00", digest: 9c851abd94d1be99 }
  - { id: bmad-05-wails-bindings, resource: docs/stories/old_stories/bmad-05-wails-bindings.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d868ae2882f2c29b }
  - { id: bmad-06-workflow-view-canvas, resource: docs/stories/old_stories/bmad-06-workflow-view-canvas.md, last_modified: "2026-04-08T10:23:03+10:00", digest: dd04c8998d8e89ef }
  - { id: bmad-07-custom-nodes-components, resource: docs/stories/old_stories/bmad-07-custom-nodes-components.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 0706808ed6acd792 }
  - { id: sprint2-03-repo-scoped-workflows, resource: docs/stories/old_stories/sprint2-03-repo-scoped-workflows.md, last_modified: "2026-04-08T17:10:27+10:00", digest: ebb314fdf3143516 }
  - { id: sprint2-04-repo-context-flow, resource: docs/stories/old_stories/sprint2-04-repo-context-flow.md, last_modified: "2026-04-08T17:10:27+10:00", digest: bee256727da15650 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/breadcrumbs-06-downstream-autofill.md`
- `docs/stories/old_stories/bmad-05-wails-bindings.md`
- `docs/stories/old_stories/bmad-06-workflow-view-canvas.md`
- `docs/stories/old_stories/bmad-07-custom-nodes-components.md`
- `docs/stories/old_stories/sprint2-03-repo-scoped-workflows.md`
- `docs/stories/old_stories/sprint2-04-repo-context-flow.md`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- BMAD Workflows (docs/SPECIFICATION.md:L934)
- Acceptance Criteria (docs/stories/breadcrumbs-06-downstream-autofill.md:L52)
- bmad-05-wails-bindings.md (docs/stories/old_stories/bmad-05-wails-bindings.md:L1)
- Wails Bindings: BMAD API Surface (docs/stories/old_stories/bmad-05-wails-bindings.md:L1)
- Technical Considerations (docs/stories/old_stories/bmad-05-wails-bindings.md:L119)
- Developer Notes (docs/stories/old_stories/bmad-05-wails-bindings.md:L12)
- Architecture (docs/stories/old_stories/bmad-05-wails-bindings.md:L14)
- Acceptance Criteria (docs/stories/old_stories/bmad-05-wails-bindings.md:L141)
- BDD Test Scenarios (docs/stories/old_stories/bmad-05-wails-bindings.md:L151)
- Scenario 1: Binding initialization (docs/stories/old_stories/bmad-05-wails-bindings.md:L153)
- Scenario 2: Process and template queries (docs/stories/old_stories/bmad-05-wails-bindings.md:L173)
- Scenario 3: Workflow CRUD via bindings (docs/stories/old_stories/bmad-05-wails-bindings.md:L196)
- Tasks / Subtasks (docs/stories/old_stories/bmad-05-wails-bindings.md:L216)
- Definition of Done (docs/stories/old_stories/bmad-05-wails-bindings.md:L236)
- Description (docs/stories/old_stories/bmad-05-wails-bindings.md:L8)
- CreateFromTemplate Logic (docs/stories/old_stories/bmad-05-wails-bindings.md:L85)
- Error Handling Pattern (docs/stories/old_stories/bmad-05-wails-bindings.md:L99)
- Acceptance Criteria (docs/stories/old_stories/bmad-06-workflow-view-canvas.md:L190)
- Risks & Edge Cases (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L155)
- Technical Considerations (docs/stories/old_stories/sprint2-03-repo-scoped-workflows.md:L59)
- Tasks / Subtasks (docs/stories/old_stories/sprint2-04-repo-context-flow.md:L157)
- ListBmadTemplates() (frontend/wailsjs/go/main/App.js:L197)
- ListBmadWorkflows() (frontend/wailsjs/go/main/App.js:L201)
- DeleteBmadAgent() (frontend/wailsjs/go/main/App.js:L21)
- DeleteBmadWorkflow() (frontend/wailsjs/go/main/App.js:L25)
- SaveBmadWorkflow() (frontend/wailsjs/go/main/App.js:L329)
- GetBmadModules() (frontend/wailsjs/go/main/App.js:L49)
- GetBmadProcesses() (frontend/wailsjs/go/main/App.js:L53)
- GetBmadProcessesByPhase() (frontend/wailsjs/go/main/App.js:L57)
- GetBmadWorkflow() (frontend/wailsjs/go/main/App.js:L61)

# Depends on
- [GetTerminalPort](/modules/getterminalport.md)
- [SpawnAgent](/modules/spawnagent.md)

# Inferred
- [10. Wails Bindings (Go → Svelte API)](/modules/10-wails-bindings-go-svelte-api.md)
- [CreateFromTemplate](/modules/createfromtemplate.md)
- [GetTerminalPort](/modules/getterminalport.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
