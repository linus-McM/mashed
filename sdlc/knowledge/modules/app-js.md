---
type: Module
title: App.js
description: "Graphify community 24: docs/SPECIFICATION.md, docs/stories/old_stories/bmad-07-custom-nodes-components.md, docs/stories/old_stories/bmad-08-execution-integration.md, frontend/src/lib/stores/sessions.t"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: bmad-07-custom-nodes-components, resource: docs/stories/old_stories/bmad-07-custom-nodes-components.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 0706808ed6acd792 }
  - { id: bmad-08-execution-integration, resource: docs/stories/old_stories/bmad-08-execution-integration.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 84e85833652d77c6 }
  - { id: sessions, resource: frontend/src/lib/stores/sessions.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 0fa9462ed476c629 }
  - { id: session, resource: frontend/src/types/session.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 0b3fa196502eba06 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/old_stories/bmad-07-custom-nodes-components.md`
- `docs/stories/old_stories/bmad-08-execution-integration.md`
- `frontend/src/lib/stores/sessions.ts`
- `frontend/src/types/session.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- 10. Wails Bindings (Go → Svelte API) (docs/SPECIFICATION.md:L828)
- Font Discovery (docs/SPECIFICATION.md:L864)
- Agents, Sessions & Terminals (docs/SPECIFICATION.md:L872)
- Git Operations (docs/SPECIFICATION.md:L890)
- Files & Editor I/O (docs/SPECIFICATION.md:L911)
- Review, Explain & Advice (streaming) (docs/SPECIFICATION.md:L921)
- BMAD Agents (docs/SPECIFICATION.md:L962)
- Sprint Management (docs/SPECIFICATION.md:L969)
- NodeConfigPanel.svelte (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L108)
- Terminal Access for Running Nodes (docs/stories/old_stories/bmad-08-execution-integration.md:L103)
- Reference Files (docs/stories/old_stories/bmad-08-execution-integration.md:L164)
- sessions.ts (frontend/src/lib/stores/sessions.ts:L1)
- repoSessions (frontend/src/lib/stores/sessions.ts:L11)
- removeSessionByName() (frontend/src/lib/stores/sessions.ts:L42)
- session.ts (frontend/src/types/session.ts:L1)
- DataFields (frontend/src/types/session.ts:L19)
- Session (frontend/src/types/session.ts:L40)
- App.js (frontend/wailsjs/go/main/App.js:L1)
- GetScopedDiff() (frontend/wailsjs/go/main/App.js:L105)
- GetTerminalPort() (frontend/wailsjs/go/main/App.js:L113)
- GetWorktrees() (frontend/wailsjs/go/main/App.js:L117)
- GitCommit() (frontend/wailsjs/go/main/App.js:L121)
- GitCommitAndPush() (frontend/wailsjs/go/main/App.js:L125)
- GitCommitPushAndPR() (frontend/wailsjs/go/main/App.js:L129)
- GitForcePush() (frontend/wailsjs/go/main/App.js:L141)
- GitListBranches() (frontend/wailsjs/go/main/App.js:L145)
- GitMergeInto() (frontend/wailsjs/go/main/App.js:L149)
- GitPull() (frontend/wailsjs/go/main/App.js:L153)
- GitPush() (frontend/wailsjs/go/main/App.js:L157)
- IsExplainAvailable() (frontend/wailsjs/go/main/App.js:L165)
- ListAllAgents() (frontend/wailsjs/go/main/App.js:L181)
- ListBmadAgents() (frontend/wailsjs/go/main/App.js:L193)
- ListLocalFonts() (frontend/wailsjs/go/main/App.js:L217)
- ListModels() (frontend/wailsjs/go/main/App.js:L221)
- ListNerdFonts() (frontend/wailsjs/go/main/App.js:L225)
- ListRepoFiles() (frontend/wailsjs/go/main/App.js:L237)
- MarkRead() (frontend/wailsjs/go/main/App.js:L253)
- OpenFontsDir() (frontend/wailsjs/go/main/App.js:L257)
- RepoMtimes() (frontend/wailsjs/go/main/App.js:L305)
- RepoStatus() (frontend/wailsjs/go/main/App.js:L309)
- GetAgentLog() (frontend/wailsjs/go/main/App.js:L33)
- SaveMashedAssetFrontmatter() (frontend/wailsjs/go/main/App.js:L333)
- SetFontSize() (frontend/wailsjs/go/main/App.js:L365)
- SetSidebarWidth() (frontend/wailsjs/go/main/App.js:L393)
- WriteConsoleLog() (frontend/wailsjs/go/main/App.js:L465)
- GetFontsDir() (frontend/wailsjs/go/main/App.js:L81)
- GetNotifications() (frontend/wailsjs/go/main/App.js:L97)

# Depends on
- [13. Edge Cases and Failure Modes](/modules/13-edge-cases-and-failure-modes.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [CreateFromTemplate](/modules/createfromtemplate.md)
- [GetConfig](/modules/getconfig.md)
- [hydrate](/modules/hydrate.md)
- [InputResponseModal.svelte](/modules/inputresponsemodal-svelte.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [ListOllamaModels](/modules/listollamamodels.md)
- [ListRepoSessions](/modules/listreposessions.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [SetEditorSettings](/modules/seteditorsettings.md)
- [SetMarkdownMenuSettings](/modules/setmarkdownmenusettings.md)
- [SpawnAgent](/modules/spawnagent.md)
- [Story 18 — Title-bar Dynamic UI model selector](/modules/story-18-title-bar-dynamic-ui-model-selector.md)
- [Story 2: Code Review Summary & Advice Streaming Backend](/modules/story-2-code-review-summary-advice-streaming-backend.md)
- [SummarisationModal.svelte](/modules/summarisationmodal-svelte.md)
- [svelte](/modules/svelte.md)
- [TakeScreenshot](/modules/takescreenshot.md)
- [Tasks / Subtasks](/modules/tasks-subtasks.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)
- [themeInit.js](/modules/themeinit-js.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [uiAdapterSettings.v3.test.ts](/modules/uiadaptersettings-v3-test-ts.md)
- [WriteFile](/modules/writefile.md)

# Inferred
- [SummarisationModal.svelte](/modules/summarisationmodal-svelte.md)
- [TakeScreenshot](/modules/takescreenshot.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
