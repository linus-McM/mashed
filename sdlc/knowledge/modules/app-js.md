---
type: Module
title: App.js
description: "Graphify community 4: docs/SPECIFICATION.md, frontend/src/App.svelte, frontend/src/components/EditorRouter.svelte, frontend/src/components/ImageViewer.svelte, frontend/src/components/MonacoEditor.svel"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: App, resource: frontend/src/App.svelte, last_modified: "2026-04-23T11:09:52+10:00", digest: 10db755a7a0e5abe }
  - { id: EditorRouter, resource: frontend/src/components/EditorRouter.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 86fac6035745e901 }
  - { id: ImageViewer, resource: frontend/src/components/ImageViewer.svelte, last_modified: "2026-04-22T19:23:56+10:00", digest: 4abd3a97e70be2b9 }
  - { id: MonacoEditor, resource: frontend/src/components/MonacoEditor.svelte, last_modified: "2026-04-22T18:59:31+10:00", digest: 7ea6153fc5c063fa }
  - { id: NewRepoModal, resource: frontend/src/components/NewRepoModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 88f99d701a9fb7ec }
  - { id: SparkLine, resource: frontend/src/components/SparkLine.svelte, last_modified: "2026-04-22T19:23:56+10:00", digest: 61bc0e7ea260e7dc }
  - { id: TitleBar, resource: frontend/src/components/TitleBar.svelte, last_modified: "2026-04-23T14:09:41+10:00", digest: 24aa88a7c4d75635 }
  - { id: GitPanel, resource: frontend/src/components/bmad/GitPanel.svelte, last_modified: "2026-04-22T19:45:14+10:00", digest: d269285a64cf37af }
  - { id: errorMessage, resource: frontend/src/lib/errorMessage.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 01d52ec3146aae0e }
  - { id: ptySize, resource: frontend/src/lib/ptySize.ts, last_modified: "2026-05-07T18:18:02+10:00", digest: 819faf79772ff83a }
  - { id: repoPalette, resource: frontend/src/lib/repoPalette.ts, last_modified: "2026-04-22T18:25:48+10:00", digest: 513bf749080da8bf }
  - { id: editorSettings, resource: frontend/src/lib/stores/editorSettings.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8f55b394c63663b4 }
  - { id: font, resource: frontend/src/lib/stores/font.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 06f56f59c2340c7a }
  - { id: AgentDetail, resource: frontend/src/views/AgentDetail.svelte, last_modified: "2026-05-07T18:18:02+10:00", digest: ae357b7360448f38 }
  - { id: BranchModal, resource: frontend/src/views/BranchModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 6d3c58fc0fa1902c }
  - { id: ForcePushModal, resource: frontend/src/views/ForcePushModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 6f35f823b43cd232 }
  - { id: MergeModal, resource: frontend/src/views/MergeModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: aad1d4fffa15e556 }
  - { id: NotificationFeed, resource: frontend/src/views/NotificationFeed.svelte, last_modified: "2026-05-07T18:18:02+10:00", digest: 06607e88326a1ff7 }
  - { id: Settings, resource: frontend/src/views/Settings.svelte, last_modified: "2026-04-23T11:09:52+10:00", digest: edc56fcaa6890012 }
  - { id: Setup, resource: frontend/src/views/Setup.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: aeb9be820e3be7e3 }
  - { id: SpawnAgent, resource: frontend/src/views/SpawnAgent.svelte, last_modified: "2026-05-07T18:18:02+10:00", digest: b9e9afd8fe710518 }
  - { id: SwitchBranchModal, resource: frontend/src/views/SwitchBranchModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 501e05cfd7841900 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `frontend/src/App.svelte`
- `frontend/src/components/EditorRouter.svelte`
- `frontend/src/components/ImageViewer.svelte`
- `frontend/src/components/MonacoEditor.svelte`
- `frontend/src/components/NewRepoModal.svelte`
- `frontend/src/components/SparkLine.svelte`
- `frontend/src/components/TitleBar.svelte`
- `frontend/src/components/bmad/GitPanel.svelte`
- `frontend/src/lib/errorMessage.ts`
- `frontend/src/lib/ptySize.ts`
- `frontend/src/lib/repoPalette.ts`
- `frontend/src/lib/stores/editorSettings.js`
- `frontend/src/lib/stores/font.js`
- `frontend/src/views/AgentDetail.svelte`
- `frontend/src/views/BranchModal.svelte`
- `frontend/src/views/ForcePushModal.svelte`
- `frontend/src/views/MergeModal.svelte`
- `frontend/src/views/NotificationFeed.svelte`
- `frontend/src/views/Settings.svelte`
- `frontend/src/views/Setup.svelte`
- `frontend/src/views/SpawnAgent.svelte`
- `frontend/src/views/SwitchBranchModal.svelte`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Font Discovery (docs/SPECIFICATION.md:L864)
- App.svelte (frontend/src/App.svelte:L1)
- if() (frontend/src/App.svelte:L344)
- EditorRouter.svelte (frontend/src/components/EditorRouter.svelte:L1)
- ImageViewer.svelte (frontend/src/components/ImageViewer.svelte:L1)
- MonacoEditor.svelte (frontend/src/components/MonacoEditor.svelte:L1)
- if() (frontend/src/components/MonacoEditor.svelte:L337)
- getWorker() (frontend/src/components/MonacoEditor.svelte:L477)
- NewRepoModal.svelte (frontend/src/components/NewRepoModal.svelte:L1)
- selected (frontend/src/components/NewRepoModal.svelte:L91)
- SparkLine.svelte (frontend/src/components/SparkLine.svelte:L1)
- TitleBar.svelte (frontend/src/components/TitleBar.svelte:L1)
- active (frontend/src/components/TitleBar.svelte:L156)
- GitPanel.svelte (frontend/src/components/bmad/GitPanel.svelte:L1)
- errorMessage.ts (frontend/src/lib/errorMessage.ts:L1)
- ptySize.ts (frontend/src/lib/ptySize.ts:L1)
- estimatePtySize() (frontend/src/lib/ptySize.ts:L17)
- repoPalette.ts (frontend/src/lib/repoPalette.ts:L1)
- REPO_BORDER_NONE (frontend/src/lib/repoPalette.ts:L18)
- REPO_BORDER_PALETTE (frontend/src/lib/repoPalette.ts:L28)
- editorSettings.js (frontend/src/lib/stores/editorSettings.js:L1)
- editorSettings (frontend/src/lib/stores/editorSettings.js:L23)
- defaults (frontend/src/lib/stores/editorSettings.js:L7)
- font.js (frontend/src/lib/stores/font.js:L1)
- applyFont() (frontend/src/lib/stores/font.js:L17)
- DEFAULT_MONO_FONT (frontend/src/lib/stores/font.js:L3)
- DEFAULT_FONT_SIZE (frontend/src/lib/stores/font.js:L4)
- registerLocalFonts() (frontend/src/lib/stores/font.js:L42)
- currentFontSize (frontend/src/lib/stores/font.js:L7)
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

# Depends on
- [13. Edge Cases and Failure Modes](/modules/13-edge-cases-and-failure-modes.md)
- [applyTheme](/modules/applytheme.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [Developer Notes](/modules/developer-notes-355.md)
- [DynamicUiSelector.svelte](/modules/dynamicuiselector-svelte.md)
- [FileTree.svelte](/modules/filetree-svelte.md)
- [GetConfig](/modules/getconfig.md)
- [imageViewerUtils.ts](/modules/imageviewerutils-ts.md)
- [InputResponseModal.svelte](/modules/inputresponsemodal-svelte.md)
- [ListOllamaModels](/modules/listollamamodels.md)
- [ListRepoSessions](/modules/listreposessions.md)
- [MarkdownEditor.test.ts](/modules/markdowneditor-test-ts.md)
- [markdownMenuSettings.ts](/modules/markdownmenusettings-ts.md)
- [mashedConfig](/modules/mashedconfig.md)
- [models.ts](/modules/models-ts.md)
- [NewSessionModal.svelte](/modules/newsessionmodal-svelte.md)
- [QuestionSnackbarStack.test.ts](/modules/questionsnackbarstack-test-ts.md)
- [ReadFileBase64](/modules/readfilebase64.md)
- [runtime.js](/modules/runtime-js.md)
- [SetEditorSettings](/modules/seteditorsettings.md)
- [SetTheme](/modules/settheme.md)
- [status.ts](/modules/status-ts.md)
- [Story 2: Code Review Summary & Advice Streaming Backend](/modules/story-2-code-review-summary-advice-streaming-backend.md)
- [Story: meditor-01 — EditorRouter -- Extension-Based Editor Switching](/modules/story-meditor-01-editorrouter-extension-based-editor-switching.md)
- [SummarisationModal.svelte](/modules/summarisationmodal-svelte.md)
- [svelte](/modules/svelte.md)
- [Tasks / Subtasks](/modules/tasks-subtasks.md)
- [Terminal.svelte](/modules/terminal-svelte.md)
- [theme.js](/modules/theme-js.md)
- [themeInit.js](/modules/themeinit-js.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [vitest](/modules/vitest.md)
- [WriteFile](/modules/writefile.md)

# Inferred
- [autoFill.test.ts](/modules/autofill-test-ts.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [Story 1: Native macOS Menu Bar Construction](/modules/story-1-native-macos-menu-bar-construction.md)
- [svelte](/modules/svelte.md)

# Features
- no feature plan names these files
