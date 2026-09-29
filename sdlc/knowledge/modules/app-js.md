---
type: Module
title: App.js
description: "Graphify community 4: docs/SPECIFICATION.md, frontend/src/App.svelte, frontend/src/components/NewRepoModal.svelte, frontend/src/components/SparkLine.svelte, frontend/src/components/bmad/GitPanel.svelt"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: App, resource: frontend/src/App.svelte, last_modified: "2026-04-23T11:09:52+10:00", digest: 10db755a7a0e5abe }
  - { id: NewRepoModal, resource: frontend/src/components/NewRepoModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 88f99d701a9fb7ec }
  - { id: SparkLine, resource: frontend/src/components/SparkLine.svelte, last_modified: "2026-04-22T19:23:56+10:00", digest: 61bc0e7ea260e7dc }
  - { id: GitPanel, resource: frontend/src/components/bmad/GitPanel.svelte, last_modified: "2026-04-22T19:45:14+10:00", digest: d269285a64cf37af }
  - { id: NodeConfigPanel, resource: frontend/src/components/bmad/NodeConfigPanel.svelte, last_modified: "2026-04-28T12:36:05+10:00", digest: 4f3162377da28e00 }
  - { id: errorMessage, resource: frontend/src/lib/errorMessage.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 01d52ec3146aae0e }
  - { id: ptySize, resource: frontend/src/lib/ptySize.ts, last_modified: "2026-05-07T18:18:02+10:00", digest: 819faf79772ff83a }
  - { id: repoPalette, resource: frontend/src/lib/repoPalette.ts, last_modified: "2026-04-22T18:25:48+10:00", digest: 513bf749080da8bf }
  - { id: editorSettings, resource: frontend/src/lib/stores/editorSettings.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8f55b394c63663b4 }
  - { id: AgentDetail, resource: frontend/src/views/AgentDetail.svelte, last_modified: "2026-05-07T18:18:02+10:00", digest: ae357b7360448f38 }
  - { id: BranchModal, resource: frontend/src/views/BranchModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 6d3c58fc0fa1902c }
  - { id: ForcePushModal, resource: frontend/src/views/ForcePushModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 6f35f823b43cd232 }
  - { id: MergeModal, resource: frontend/src/views/MergeModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: aad1d4fffa15e556 }
  - { id: NotificationFeed, resource: frontend/src/views/NotificationFeed.svelte, last_modified: "2026-05-07T18:18:02+10:00", digest: 06607e88326a1ff7 }
  - { id: Settings, resource: frontend/src/views/Settings.svelte, last_modified: "2026-04-23T11:09:52+10:00", digest: edc56fcaa6890012 }
  - { id: Setup, resource: frontend/src/views/Setup.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: aeb9be820e3be7e3 }
  - { id: SpawnAgent, resource: frontend/src/views/SpawnAgent.svelte, last_modified: "2026-05-07T18:18:02+10:00", digest: b9e9afd8fe710518 }
  - { id: SummarisationModal, resource: frontend/src/views/SummarisationModal.svelte, last_modified: "2026-04-22T20:32:12+10:00", digest: 1b621514b92c549f }
  - { id: SwitchBranchModal, resource: frontend/src/views/SwitchBranchModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 501e05cfd7841900 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
  - { id: runtime, resource: frontend/wailsjs/runtime/runtime.js, last_modified: "2026-05-07T09:55:31+10:00", digest: e25fe86d3c590de7 }
  - { id: easing, resource: svelte/easing, last_modified: "2026-09-29T11:35:20Z", digest: missing }
---

# Files
- `docs/SPECIFICATION.md`
- `frontend/src/App.svelte`
- `frontend/src/components/NewRepoModal.svelte`
- `frontend/src/components/SparkLine.svelte`
- `frontend/src/components/bmad/GitPanel.svelte`
- `frontend/src/components/bmad/NodeConfigPanel.svelte`
- `frontend/src/lib/errorMessage.ts`
- `frontend/src/lib/ptySize.ts`
- `frontend/src/lib/repoPalette.ts`
- `frontend/src/lib/stores/editorSettings.js`
- `frontend/src/views/AgentDetail.svelte`
- `frontend/src/views/BranchModal.svelte`
- `frontend/src/views/ForcePushModal.svelte`
- `frontend/src/views/MergeModal.svelte`
- `frontend/src/views/NotificationFeed.svelte`
- `frontend/src/views/Settings.svelte`
- `frontend/src/views/Setup.svelte`
- `frontend/src/views/SpawnAgent.svelte`
- `frontend/src/views/SummarisationModal.svelte`
- `frontend/src/views/SwitchBranchModal.svelte`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/runtime/runtime.js`
- `svelte/easing`

# Symbols
- Font Discovery (docs/SPECIFICATION.md:L864)
- App.svelte (frontend/src/App.svelte:L1)
- if() (frontend/src/App.svelte:L344)
- NewRepoModal.svelte (frontend/src/components/NewRepoModal.svelte:L1)
- selected (frontend/src/components/NewRepoModal.svelte:L91)
- SparkLine.svelte (frontend/src/components/SparkLine.svelte:L1)
- GitPanel.svelte (frontend/src/components/bmad/GitPanel.svelte:L1)
- onMouseMove() (frontend/src/components/bmad/NodeConfigPanel.svelte:L39)
- onMouseUp() (frontend/src/components/bmad/NodeConfigPanel.svelte:L43)
- errorMessage.ts (frontend/src/lib/errorMessage.ts:L1)
- ptySize.ts (frontend/src/lib/ptySize.ts:L1)
- estimatePtySize() (frontend/src/lib/ptySize.ts:L17)
- repoPalette.ts (frontend/src/lib/repoPalette.ts:L1)
- REPO_BORDER_NONE (frontend/src/lib/repoPalette.ts:L18)
- REPO_BORDER_PALETTE (frontend/src/lib/repoPalette.ts:L28)
- editorSettings.js (frontend/src/lib/stores/editorSettings.js:L1)
- editorSettings (frontend/src/lib/stores/editorSettings.js:L23)
- defaults (frontend/src/lib/stores/editorSettings.js:L7)
- AgentDetail.svelte (frontend/src/views/AgentDetail.svelte:L1)
- active (frontend/src/views/AgentDetail.svelte:L457)
- if() (frontend/src/views/AgentDetail.svelte:L70)
- BranchModal.svelte (frontend/src/views/BranchModal.svelte:L1)
- repoPath (frontend/src/views/BranchModal.svelte:L22)
- repoBranch (frontend/src/views/BranchModal.svelte:L23)
- prefixes (frontend/src/views/BranchModal.svelte:L25)
- sanitize() (frontend/src/views/BranchModal.svelte:L45)
- handleInput() (frontend/src/views/BranchModal.svelte:L58)
- create() (frontend/src/views/BranchModal.svelte:L66)
- cancel() (frontend/src/views/BranchModal.svelte:L79)
- handleKeydown() (frontend/src/views/BranchModal.svelte:L84)
- ForcePushModal.svelte (frontend/src/views/ForcePushModal.svelte:L1)
- cancel() (frontend/src/views/ForcePushModal.svelte:L28)
- handleKeydown() (frontend/src/views/ForcePushModal.svelte:L33)
- MergeModal.svelte (frontend/src/views/MergeModal.svelte:L1)
- NotificationFeed.svelte (frontend/src/views/NotificationFeed.svelte:L1)
- if() (frontend/src/views/NotificationFeed.svelte:L377)
- Settings.svelte (frontend/src/views/Settings.svelte:L1)
- Setup.svelte (frontend/src/views/Setup.svelte:L1)
- SpawnAgent.svelte (frontend/src/views/SpawnAgent.svelte:L1)
- selected (frontend/src/views/SpawnAgent.svelte:L91)
- SummarisationModal.svelte (frontend/src/views/SummarisationModal.svelte:L1)
- close() (frontend/src/views/SummarisationModal.svelte:L127)
- handleKeydown() (frontend/src/views/SummarisationModal.svelte:L131)
- openFile() (frontend/src/views/SummarisationModal.svelte:L135)
- toggleFile() (frontend/src/views/SummarisationModal.svelte:L139)
- handleCardKey() (frontend/src/views/SummarisationModal.svelte:L156)
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
- GitCreateBranch() (frontend/wailsjs/go/main/App.js:L137)
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
- WriteConsoleLog() (frontend/wailsjs/go/main/App.js:L465)
- GetFontsDir() (frontend/wailsjs/go/main/App.js:L81)
- EventsOn() (frontend/wailsjs/runtime/runtime.js:L43)
- svelte/easing (svelte/easing:)

# Depends on
- [10. Wails Bindings (Go → Svelte API)](/modules/10-wails-bindings-go-svelte-api.md)
- [applyTheme](/modules/applytheme.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [FileTree.svelte](/modules/filetree-svelte.md)
- [hydrate](/modules/hydrate.md)
- [KillAgent](/modules/killagent.md)
- [ListOllamaModels](/modules/listollamamodels.md)
- [ListRepoSessions](/modules/listreposessions.md)
- [MarkdownEditor.svelte](/modules/markdowneditor-svelte.md)
- [markdownMenuSettings.ts](/modules/markdownmenusettings-ts.md)
- [mashedConfig](/modules/mashedconfig.md)
- [models.ts](/modules/models-ts.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [NewSessionModal.svelte](/modules/newsessionmodal-svelte.md)
- [QuestionSnackbarStack.test.ts](/modules/questionsnackbarstack-test-ts.md)
- [ReadFileBase64](/modules/readfilebase64.md)
- [runtime.js](/modules/runtime-js.md)
- [SetEditorSettings](/modules/seteditorsettings.md)
- [SetMarkdownMenuSettings](/modules/setmarkdownmenusettings.md)
- [SetTheme](/modules/settheme.md)
- [status.ts](/modules/status-ts.md)
- [Story 18 — Title-bar Dynamic UI model selector](/modules/story-18-title-bar-dynamic-ui-model-selector.md)
- [Story 2: Code Review Summary & Advice Streaming Backend](/modules/story-2-code-review-summary-advice-streaming-backend.md)
- [StreamAdvice](/modules/streamadvice.md)
- [StreamScopedAdvice](/modules/streamscopedadvice.md)
- [svelte](/modules/svelte.md)
- [TakeScreenshot](/modules/takescreenshot.md)
- [Tasks / Subtasks](/modules/tasks-subtasks.md)
- [themeInit.js](/modules/themeinit-js.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [uiAdapterSettings.v3.test.ts](/modules/uiadaptersettings-v3-test-ts.md)
- [vitest](/modules/vitest.md)
- [WriteFile](/modules/writefile.md)

# Inferred
- [autoFill.test.ts](/modules/autofill-test-ts.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [StreamScopedAdvice](/modules/streamscopedadvice.md)
- [svelte](/modules/svelte.md)

# Features
- no feature plan names these files
