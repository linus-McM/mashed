---
type: Module
title: 10. Wails Bindings (Go → Svelte API)
description: "Graphify community 165: docs/SPECIFICATION.md, docs/stories/bmad-interactive-05-persistence-resume.md, docs/stories/old_stories/bmad-05-wails-bindings.md, docs/stories/old_stories/bmad-07-custom-nodes"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: bmad-interactive-05-persistence-resume, resource: docs/stories/bmad-interactive-05-persistence-resume.md, last_modified: "2026-04-20T14:41:10+10:00", digest: 74c5e957d012b474 }
  - { id: bmad-05-wails-bindings, resource: docs/stories/old_stories/bmad-05-wails-bindings.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d868ae2882f2c29b }
  - { id: bmad-07-custom-nodes-components, resource: docs/stories/old_stories/bmad-07-custom-nodes-components.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 0706808ed6acd792 }
  - { id: bmad-08-execution-integration, resource: docs/stories/old_stories/bmad-08-execution-integration.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 84e85833652d77c6 }
  - { id: bmad-sprint-backlog, resource: docs/stories/old_stories/bmad-sprint-backlog.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 2dd27762cbfcdc8d }
  - { id: DynamicUiSelector, resource: frontend/src/components/titlebar/DynamicUiSelector.svelte, last_modified: "2026-04-23T13:17:59+10:00", digest: 5665ff6cccbc963c }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/bmad-interactive-05-persistence-resume.md`
- `docs/stories/old_stories/bmad-05-wails-bindings.md`
- `docs/stories/old_stories/bmad-07-custom-nodes-components.md`
- `docs/stories/old_stories/bmad-08-execution-integration.md`
- `docs/stories/old_stories/bmad-sprint-backlog.md`
- `frontend/src/components/titlebar/DynamicUiSelector.svelte`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- 10. Wails Bindings (Go → Svelte API) (docs/SPECIFICATION.md:L828)
- Git Operations (docs/SPECIFICATION.md:L890)
- Files & Editor I/O (docs/SPECIFICATION.md:L911)
- Review, Explain & Advice (streaming) (docs/SPECIFICATION.md:L921)
- BMAD Agents (docs/SPECIFICATION.md:L962)
- Sprint Management (docs/SPECIFICATION.md:L969)
- Risks / gotchas (docs/stories/bmad-interactive-05-persistence-resume.md:L113)
- bmad-05-wails-bindings.md (docs/stories/old_stories/bmad-05-wails-bindings.md:L1)
- bmad-07-custom-nodes-components.md (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L1)
- Frontend: Custom ProcessNode & Supporting Components (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L1)
- NodeConfigPanel.svelte (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L108)
- Developer Notes (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L12)
- TemplatePicker.svelte (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L131)
- Architecture (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L14)
- Technical Considerations (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L148)
- Reference Files (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L161)
- Acceptance Criteria (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L169)
- BDD Test Scenarios (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L179)
- Scenario 1: ProcessNode rendering (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L181)
- Scenario 2: ExecutionBar controls (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L208)
- Scenario 3: Node configuration (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L227)
- Tasks / Subtasks (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L245)
- ProcessNode.svelte (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L25)
- Definition of Done (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L275)
- Description (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L8)
- ExecutionBar.svelte (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L80)
- bmad-08-execution-integration.md (docs/stories/old_stories/bmad-08-execution-integration.md:L1)
- bmad-sprint-backlog.md (docs/stories/old_stories/bmad-sprint-backlog.md:L1)
- Sprint Backlog: BMAD Workflow Builder (docs/stories/old_stories/bmad-sprint-backlog.md:L1)
- Dependency Graph (docs/stories/old_stories/bmad-sprint-backlog.md:L17)
- Stories (docs/stories/old_stories/bmad-sprint-backlog.md:L3)
- Parallel Execution Opportunities (docs/stories/old_stories/bmad-sprint-backlog.md:L35)
- Recommended Sprint Order (docs/stories/old_stories/bmad-sprint-backlog.md:L60)
- Summary (docs/stories/old_stories/bmad-sprint-backlog.md:L71)
- DynamicUiSelector.svelte (frontend/src/components/titlebar/DynamicUiSelector.svelte:L1)
- close() (frontend/src/components/titlebar/DynamicUiSelector.svelte:L13)
- handleKey() (frontend/src/components/titlebar/DynamicUiSelector.svelte:L23)
- select() (frontend/src/components/titlebar/DynamicUiSelector.svelte:L33)
- truncate() (frontend/src/components/titlebar/DynamicUiSelector.svelte:L41)
- active (frontend/src/components/titlebar/DynamicUiSelector.svelte:L72)
- disabled (frontend/src/components/titlebar/DynamicUiSelector.svelte:L73)
- IsExplainAvailable() (frontend/wailsjs/go/main/App.js:L165)
- ListBmadAgents() (frontend/wailsjs/go/main/App.js:L193)

# Depends on
- [App.js](/modules/app-js.md)
- [SetEditorSettings](/modules/seteditorsettings.md)
- [Tasks / Subtasks](/modules/tasks-subtasks.md)
- [themeInit.js](/modules/themeinit-js.md)
- [ui-ast-U8: "View raw" fallback toggle + diagnostics surface](/modules/ui-ast-u8-view-raw-fallback-toggle-diagnostics-surface.md)

# Inferred
- [StreamAdvice](/modules/streamadvice.md)

# Features
- no feature plan names these files
