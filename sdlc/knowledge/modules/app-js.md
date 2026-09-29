---
type: Module
title: App.js
description: "Graphify community 4: docs/SPECIFICATION.md, docs/stories/old_stories/bmad-05-wails-bindings.md, docs/stories/old_stories/bmad-07-custom-nodes-components.md, docs/stories/old_stories/bmad-08-execution"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: bmad-05-wails-bindings, resource: docs/stories/old_stories/bmad-05-wails-bindings.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d868ae2882f2c29b }
  - { id: bmad-07-custom-nodes-components, resource: docs/stories/old_stories/bmad-07-custom-nodes-components.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 0706808ed6acd792 }
  - { id: bmad-08-execution-integration, resource: docs/stories/old_stories/bmad-08-execution-integration.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 84e85833652d77c6 }
  - { id: ui-ast-U8-sprint-report, resource: docs/stories/ui-ast-U8-sprint-report.md, last_modified: "2026-04-22T12:58:08+10:00", digest: f3eb1e6e107bf539 }
  - { id: InputResponseModal, resource: frontend/src/components/bmad/InputResponseModal.svelte, last_modified: "2026-04-28T12:36:05+10:00", digest: 70963c4ec4be0aac }
  - { id: sessions, resource: frontend/src/lib/stores/sessions.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 0fa9462ed476c629 }
  - { id: session, resource: frontend/src/types/session.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 0b3fa196502eba06 }
  - { id: transcript, resource: frontend/src/types/transcript.ts, last_modified: "2026-04-22T20:18:54+10:00", digest: 3a79aee54e6f6bf3 }
  - { id: AgentDetail, resource: frontend/src/views/AgentDetail.svelte, last_modified: "2026-05-07T18:18:02+10:00", digest: ae357b7360448f38 }
  - { id: SwitchBranchModal, resource: frontend/src/views/SwitchBranchModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 501e05cfd7841900 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/old_stories/bmad-05-wails-bindings.md`
- `docs/stories/old_stories/bmad-07-custom-nodes-components.md`
- `docs/stories/old_stories/bmad-08-execution-integration.md`
- `docs/stories/ui-ast-U8-sprint-report.md`
- `frontend/src/components/bmad/InputResponseModal.svelte`
- `frontend/src/lib/stores/sessions.ts`
- `frontend/src/types/session.ts`
- `frontend/src/types/transcript.ts`
- `frontend/src/views/AgentDetail.svelte`
- `frontend/src/views/SwitchBranchModal.svelte`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- 10. Wails Bindings (Go → Svelte API) (docs/SPECIFICATION.md:L828)
- Font Discovery (docs/SPECIFICATION.md:L864)
- Agents, Sessions & Terminals (docs/SPECIFICATION.md:L872)
- Files & Editor I/O (docs/SPECIFICATION.md:L911)
- Review, Explain & Advice (streaming) (docs/SPECIFICATION.md:L921)
- BMAD Agents (docs/SPECIFICATION.md:L962)
- Sprint Management (docs/SPECIFICATION.md:L969)
- Reference Files (docs/stories/old_stories/bmad-05-wails-bindings.md:L133)
- NodeConfigPanel.svelte (docs/stories/old_stories/bmad-07-custom-nodes-components.md:L108)
- Terminal Access for Running Nodes (docs/stories/old_stories/bmad-08-execution-integration.md:L103)
- Reference Files (docs/stories/old_stories/bmad-08-execution-integration.md:L164)
- Follow-up debt (non-blocking, raised by reviewers) (docs/stories/ui-ast-U8-sprint-report.md:L95)
- loadTranscript() (frontend/src/components/bmad/InputResponseModal.svelte:L193)
- sessions.ts (frontend/src/lib/stores/sessions.ts:L1)
- repoSessions (frontend/src/lib/stores/sessions.ts:L11)
- removeSessionByName() (frontend/src/lib/stores/sessions.ts:L42)
- session.ts (frontend/src/types/session.ts:L1)
- DataFields (frontend/src/types/session.ts:L19)
- Session (frontend/src/types/session.ts:L40)
- transcript.ts (frontend/src/types/transcript.ts:L1)
- TurnRole (frontend/src/types/transcript.ts:L13)
- normaliseInteractiveTurn() (frontend/src/types/transcript.ts:L34)
- AgentDetail.svelte (frontend/src/views/AgentDetail.svelte:L1)
- active (frontend/src/views/AgentDetail.svelte:L457)
- if() (frontend/src/views/AgentDetail.svelte:L70)
- SwitchBranchModal.svelte (frontend/src/views/SwitchBranchModal.svelte:L1)
- switchBranch() (frontend/src/views/SwitchBranchModal.svelte:L42)
- cancel() (frontend/src/views/SwitchBranchModal.svelte:L55)
- handleKeydown() (frontend/src/views/SwitchBranchModal.svelte:L60)
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
- GitSwitchBranch() (frontend/wailsjs/go/main/App.js:L161)
- IsExplainAvailable() (frontend/wailsjs/go/main/App.js:L165)
- ListAllAgents() (frontend/wailsjs/go/main/App.js:L181)
- ListBmadAgents() (frontend/wailsjs/go/main/App.js:L193)
- ListLocalFonts() (frontend/wailsjs/go/main/App.js:L217)
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
- GetInteractiveTranscript() (frontend/wailsjs/go/main/App.js:L85)
- GetNotifications() (frontend/wailsjs/go/main/App.js:L97)

# Depends on
- [13. Edge Cases and Failure Modes](/modules/13-edge-cases-and-failure-modes.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [FileTree.svelte](/modules/filetree-svelte.md)
- [frontend/package.json](/modules/frontend-package-json.md)
- [GetConfig](/modules/getconfig.md)
- [hydrate](/modules/hydrate.md)
- [ListOllamaModels](/modules/listollamamodels.md)
- [ListRepoSessions](/modules/listreposessions.md)
- [runtime.js](/modules/runtime-js.md)
- [SetEditorSettings](/modules/seteditorsettings.md)
- [SetMarkdownMenuSettings](/modules/setmarkdownmenusettings.md)
- [SpawnAgent](/modules/spawnagent.md)
- [SpawnRefactorPlan](/modules/spawnrefactorplan.md)
- [status.ts](/modules/status-ts.md)
- [Story 18 — Title-bar Dynamic UI model selector](/modules/story-18-title-bar-dynamic-ui-model-selector.md)
- [Story svelte-check-01: JS stores → TypeScript (Phase 1)](/modules/story-svelte-check-01-js-stores-typescript-phase-1.md)
- [svelte](/modules/svelte.md)
- [Tasks / Subtasks](/modules/tasks-subtasks.md)
- [themeInit.js](/modules/themeinit-js.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [uiAdapterSettings.v3.test.ts](/modules/uiadaptersettings-v3-test-ts.md)
- [vitest](/modules/vitest.md)
- [WriteFile](/modules/writefile.md)

# Inferred
- [GetConfig](/modules/getconfig.md)
- [SpawnAgent](/modules/spawnagent.md)
- [SpawnRefactorPlan](/modules/spawnrefactorplan.md)
- [svelte](/modules/svelte.md)

# Features
- no feature plan names these files
