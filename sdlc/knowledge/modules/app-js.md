---
type: Module
title: App.js
description: "Graphify community 24: docs/SPECIFICATION.md, docs/stories/old_stories/bmad-05-wails-bindings.md, docs/stories/old_stories/bmad-07-custom-nodes-components.md, docs/stories/old_stories/bmad-08-executio"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: bmad-05-wails-bindings, resource: docs/stories/old_stories/bmad-05-wails-bindings.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d868ae2882f2c29b }
  - { id: bmad-07-custom-nodes-components, resource: docs/stories/old_stories/bmad-07-custom-nodes-components.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 0706808ed6acd792 }
  - { id: bmad-08-execution-integration, resource: docs/stories/old_stories/bmad-08-execution-integration.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 84e85833652d77c6 }
  - { id: bmad-09-modules-agents, resource: docs/stories/old_stories/bmad-09-modules-agents.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 4e44f1961fabcac4 }
  - { id: markdownMenuSettings.test, resource: frontend/src/lib/stores/markdownMenuSettings.test.ts, last_modified: "2026-04-23T11:02:33+10:00", digest: 440326bc7b010a88 }
  - { id: markdownMenuSettings, resource: frontend/src/lib/stores/markdownMenuSettings.ts, last_modified: "2026-04-23T11:02:33+10:00", digest: 5fd83b95073decb2 }
  - { id: AgentDetail, resource: frontend/src/views/AgentDetail.svelte, last_modified: "2026-05-07T18:18:02+10:00", digest: ae357b7360448f38 }
  - { id: SwitchBranchModal, resource: frontend/src/views/SwitchBranchModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 501e05cfd7841900 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/old_stories/bmad-05-wails-bindings.md`
- `docs/stories/old_stories/bmad-07-custom-nodes-components.md`
- `docs/stories/old_stories/bmad-08-execution-integration.md`
- `docs/stories/old_stories/bmad-09-modules-agents.md`
- `frontend/src/lib/stores/markdownMenuSettings.test.ts`
- `frontend/src/lib/stores/markdownMenuSettings.ts`
- `frontend/src/views/AgentDetail.svelte`
- `frontend/src/views/SwitchBranchModal.svelte`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- 10. Wails Bindings (Go → Svelte API) (docs/SPECIFICATION.md:L828)
- Font Discovery (docs/SPECIFICATION.md:L864)
- Agents, Sessions & Terminals (docs/SPECIFICATION.md:L872)
- Git Operations (docs/SPECIFICATION.md:L890)
- Files & Editor I/O (docs/SPECIFICATION.md:L911)
- Review, Explain & Advice (streaming) (docs/SPECIFICATION.md:L921)
- BMAD Workflows (docs/SPECIFICATION.md:L934)
- BMAD Agents (docs/SPECIFICATION.md:L962)
- Sprint Management (docs/SPECIFICATION.md:L969)
- Technical Considerations (docs/stories/old_stories/bmad-05-wails-bindings.md:L119)
- Developer Notes (docs/stories/old_stories/bmad-05-wails-bindings.md:L12)
- Risks & Edge Cases (docs/stories/old_stories/bmad-05-wails-bindings.md:L127)
- Reference Files (docs/stories/old_stories/bmad-05-wails-bindings.md:L133)
- Architecture (docs/stories/old_stories/bmad-05-wails-bindings.md:L14)
- Acceptance Criteria (docs/stories/old_stories/bmad-05-wails-bindings.md:L141)
- Tasks / Subtasks (docs/stories/old_stories/bmad-05-wails-bindings.md:L216)
- CreateFromTemplate Logic (docs/stories/old_stories/bmad-05-wails-bindings.md:L85)
- Error Handling Pattern (docs/stories/old_stories/bmad-05-wails-bindings.md:L99)
- NodeConfigPanel.svelte (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L108)
- Risks & Edge Cases (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L155)
- Terminal Access for Running Nodes (docs/stories/old_stories/bmad-08-execution-integration.md:L103)
- Developer Notes (docs/stories/old_stories/bmad-08-execution-integration.md:L12)
- Artifact Context Passing (Backend) (docs/stories/old_stories/bmad-08-execution-integration.md:L121)
- Event Data Shapes (docs/stories/old_stories/bmad-08-execution-integration.md:L130)
- Architecture (docs/stories/old_stories/bmad-08-execution-integration.md:L14)
- Technical Considerations (docs/stories/old_stories/bmad-08-execution-integration.md:L149)
- Risks & Edge Cases (docs/stories/old_stories/bmad-08-execution-integration.md:L157)
- Reference Files (docs/stories/old_stories/bmad-08-execution-integration.md:L164)
- Frontend Event Wiring (docs/stories/old_stories/bmad-08-execution-integration.md:L22)
- Acceptance Criteria (docs/stories/old_stories/bmad-09-modules-agents.md:L131)
- markdownMenuSettings.test.ts (frontend/src/lib/stores/markdownMenuSettings.test.ts:L1)
- mocks (frontend/src/lib/stores/markdownMenuSettings.test.ts:L18)
- DEFAULTS (frontend/src/lib/stores/markdownMenuSettings.test.ts:L22)
- markdownMenuSettings.ts (frontend/src/lib/stores/markdownMenuSettings.ts:L1)
- DEFAULTS (frontend/src/lib/stores/markdownMenuSettings.ts:L13)
- markdownMenuDirty (frontend/src/lib/stores/markdownMenuSettings.ts:L23)
- AgentDetail.svelte (frontend/src/views/AgentDetail.svelte:L1)
- active (frontend/src/views/AgentDetail.svelte:L457)
- if() (frontend/src/views/AgentDetail.svelte:L70)
- SwitchBranchModal.svelte (frontend/src/views/SwitchBranchModal.svelte:L1)
- switchBranch() (frontend/src/views/SwitchBranchModal.svelte:L42)
- cancel() (frontend/src/views/SwitchBranchModal.svelte:L55)
- handleKeydown() (frontend/src/views/SwitchBranchModal.svelte:L60)
- App.js (frontend/wailsjs/go/main/App.js:L1)
- GetScopedDiff() (frontend/wailsjs/go/main/App.js:L105)
- GetTerminalPort() (frontend/wailsjs/go/main/App.js:L117)
- GetWorktrees() (frontend/wailsjs/go/main/App.js:L121)
- GitCommit() (frontend/wailsjs/go/main/App.js:L125)
- GitCommitAndPush() (frontend/wailsjs/go/main/App.js:L129)
- GitCommitPushAndPR() (frontend/wailsjs/go/main/App.js:L133)
- GitForcePush() (frontend/wailsjs/go/main/App.js:L145)
- GitListBranches() (frontend/wailsjs/go/main/App.js:L149)
- GitMergeInto() (frontend/wailsjs/go/main/App.js:L153)
- GitPull() (frontend/wailsjs/go/main/App.js:L157)
- GitPush() (frontend/wailsjs/go/main/App.js:L161)
- GitSwitchBranch() (frontend/wailsjs/go/main/App.js:L165)
- IsExplainAvailable() (frontend/wailsjs/go/main/App.js:L169)
- ListAllAgents() (frontend/wailsjs/go/main/App.js:L185)
- ListBmadAgents() (frontend/wailsjs/go/main/App.js:L197)
- ListBmadTemplates() (frontend/wailsjs/go/main/App.js:L201)
- DeleteBmadAgent() (frontend/wailsjs/go/main/App.js:L21)
- ListLocalFonts() (frontend/wailsjs/go/main/App.js:L221)
- ListModels() (frontend/wailsjs/go/main/App.js:L225)
- ListNerdFonts() (frontend/wailsjs/go/main/App.js:L229)
- ListRepoFiles() (frontend/wailsjs/go/main/App.js:L241)
- DeleteBmadWorkflow() (frontend/wailsjs/go/main/App.js:L25)
- MarkRead() (frontend/wailsjs/go/main/App.js:L257)
- OpenFontsDir() (frontend/wailsjs/go/main/App.js:L261)
- RepoMtimes() (frontend/wailsjs/go/main/App.js:L309)
- RepoStatus() (frontend/wailsjs/go/main/App.js:L313)
- SaveBmadAgent() (frontend/wailsjs/go/main/App.js:L329)
- GetAgentLog() (frontend/wailsjs/go/main/App.js:L33)
- SaveMashedAssetFrontmatter() (frontend/wailsjs/go/main/App.js:L337)
- SetFontSize() (frontend/wailsjs/go/main/App.js:L369)
- SetSidebarWidth() (frontend/wailsjs/go/main/App.js:L397)
- StartBmadWorkflow() (frontend/wailsjs/go/main/App.js:L441)
- GetBmadExecution() (frontend/wailsjs/go/main/App.js:L45)
- WriteConsoleLog() (frontend/wailsjs/go/main/App.js:L469)
- GetBmadModules() (frontend/wailsjs/go/main/App.js:L49)
- GetBmadProcesses() (frontend/wailsjs/go/main/App.js:L53)
- GetBmadProcessesByPhase() (frontend/wailsjs/go/main/App.js:L57)
- GetBmadWorkflow() (frontend/wailsjs/go/main/App.js:L61)
- GetFontsDir() (frontend/wailsjs/go/main/App.js:L81)
- GetNotifications() (frontend/wailsjs/go/main/App.js:L97)

# Depends on
- [bmad-sprint-backlog.md](/modules/bmad-sprint-backlog-md.md)
- [BranchModal.svelte](/modules/branchmodal-svelte.md)
- [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [FileTree.svelte](/modules/filetree-svelte.md)
- [hydrate](/modules/hydrate.md)
- [InputResponseModal.svelte](/modules/inputresponsemodal-svelte.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [ListOllamaModels](/modules/listollamamodels.md)
- [ListRepoChoices](/modules/listrepochoices.md)
- [ListRepoSessions](/modules/listreposessions.md)
- [loadBundledThemes](/modules/loadbundledthemes.md)
- [markdownToolbarBuilder.test.ts](/modules/markdowntoolbarbuilder-test-ts.md)
- [models.ts](/modules/models-ts.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [runtime.js](/modules/runtime-js.md)
- [SetEditorSettings](/modules/seteditorsettings.md)
- [SetMarkdownMenuSettings](/modules/setmarkdownmenusettings.md)
- [SpawnAgent](/modules/spawnagent.md)
- [status.ts](/modules/status-ts.md)
- [Story 18 — Title-bar Dynamic UI model selector](/modules/story-18-title-bar-dynamic-ui-model-selector.md)
- [StreamAdvice](/modules/streamadvice.md)
- [SummarisationModal.svelte](/modules/summarisationmodal-svelte.md)
- [svelte](/modules/svelte.md)
- [theme.js](/modules/theme-js.md)
- [themeInit.js](/modules/themeinit-js.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [uiAdapterSettings.v3.test.ts](/modules/uiadaptersettings-v3-test-ts.md)
- [vitest](/modules/vitest.md)
- [WriteFile](/modules/writefile.md)

# Inferred
- [bmad-sprint-backlog.md](/modules/bmad-sprint-backlog-md.md)
- [ListRepoChoices](/modules/listrepochoices.md)
- [SpawnAgent](/modules/spawnagent.md)
- [StreamAdvice](/modules/streamadvice.md)
- [SummarisationModal.svelte](/modules/summarisationmodal-svelte.md)
- [svelte](/modules/svelte.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
