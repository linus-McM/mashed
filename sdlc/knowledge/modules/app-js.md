---
type: Module
title: App.js
description: "Graphify community 4: docs/SPECIFICATION.md, frontend/src/lib/stores/sessions.ts, frontend/src/types/session.ts, frontend/src/views/AgentDetail.svelte, frontend/src/views/SwitchBranchModal.svelte, fro"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:14Z" }
stale_after: "2026-10-13T12:23:14Z"
source_commit: ab1f2eb4ec65fc2b1fce15503c11f0e310758404
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: sessions, resource: frontend/src/lib/stores/sessions.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 0fa9462ed476c629 }
  - { id: session, resource: frontend/src/types/session.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 0b3fa196502eba06 }
  - { id: AgentDetail, resource: frontend/src/views/AgentDetail.svelte, last_modified: "2026-05-07T18:18:02+10:00", digest: ae357b7360448f38 }
  - { id: SwitchBranchModal, resource: frontend/src/views/SwitchBranchModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 501e05cfd7841900 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `frontend/src/lib/stores/sessions.ts`
- `frontend/src/types/session.ts`
- `frontend/src/views/AgentDetail.svelte`
- `frontend/src/views/SwitchBranchModal.svelte`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Font Discovery (docs/SPECIFICATION.md:L864)
- sessions.ts (frontend/src/lib/stores/sessions.ts:L1)
- repoSessions (frontend/src/lib/stores/sessions.ts:L11)
- removeSessionByName() (frontend/src/lib/stores/sessions.ts:L42)
- session.ts (frontend/src/types/session.ts:L1)
- DataFields (frontend/src/types/session.ts:L19)
- Session (frontend/src/types/session.ts:L40)
- AgentDetail.svelte (frontend/src/views/AgentDetail.svelte:L1)
- active (frontend/src/views/AgentDetail.svelte:L457)
- if() (frontend/src/views/AgentDetail.svelte:L70)
- SwitchBranchModal.svelte (frontend/src/views/SwitchBranchModal.svelte:L1)
- switchBranch() (frontend/src/views/SwitchBranchModal.svelte:L42)
- cancel() (frontend/src/views/SwitchBranchModal.svelte:L55)
- handleKeydown() (frontend/src/views/SwitchBranchModal.svelte:L60)
- App.js (frontend/wailsjs/go/main/App.js:L1)
- GetScopedDiff() (frontend/wailsjs/go/main/App.js:L105)
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
- SpawnPRReview() (frontend/wailsjs/go/main/App.js:L425)
- WriteConsoleLog() (frontend/wailsjs/go/main/App.js:L465)
- GetFontsDir() (frontend/wailsjs/go/main/App.js:L81)

# Depends on
- [10. Wails Bindings (Go → Svelte API)](/modules/10-wails-bindings-go-svelte-api.md)
- [applyTheme](/modules/applytheme.md)
- [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [CreateFromTemplate](/modules/createfromtemplate.md)
- [FileTree.svelte](/modules/filetree-svelte.md)
- [GetTerminalPort](/modules/getterminalport.md)
- [hydrate](/modules/hydrate.md)
- [InputResponseModal.svelte](/modules/inputresponsemodal-svelte.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [ListOllamaModels](/modules/listollamamodels.md)
- [ListRepoSessions](/modules/listreposessions.md)
- [markdownMenuSettings.ts](/modules/markdownmenusettings-ts.md)
- [mashedConfig](/modules/mashedconfig.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [runtime.js](/modules/runtime-js.md)
- [SetEditorSettings](/modules/seteditorsettings.md)
- [SetTheme](/modules/settheme.md)
- [SpawnAgent](/modules/spawnagent.md)
- [status.ts](/modules/status-ts.md)
- [Story 18 — Title-bar Dynamic UI model selector](/modules/story-18-title-bar-dynamic-ui-model-selector.md)
- [Story 2: Code Review Summary & Advice Streaming Backend](/modules/story-2-code-review-summary-advice-streaming-backend.md)
- [StreamScopedAdvice](/modules/streamscopedadvice.md)
- [svelte](/modules/svelte.md)
- [TakeScreenshot](/modules/takescreenshot.md)
- [Tasks / Subtasks](/modules/tasks-subtasks.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)
- [themeInit.js](/modules/themeinit-js.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [uiAdapterSettings.v3.test.ts](/modules/uiadaptersettings-v3-test-ts.md)
- [WriteFile](/modules/writefile.md)

# Inferred
- [svelte](/modules/svelte.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
