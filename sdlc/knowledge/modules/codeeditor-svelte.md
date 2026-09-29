---
type: Module
title: CodeEditor.svelte
description: "Graphify community 248: @milkdown/crepe, @milkdown/crepe/theme/classic-dark.css, @milkdown/plugin-listener, frontend/src/components/CodeEditor.svelte, frontend/src/components/EditorRouter.svelte, fron"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:08Z" }
stale_after: "2026-10-13T11:16:08Z"
source_commit: 85e7124b5a77a2f29bd9db818781712ea7a1049b
sources:
  - { id: crepe, resource: "@milkdown/crepe", last_modified: "2026-09-29T11:16:08Z", digest: missing }
  - { id: classic-dark, resource: "@milkdown/crepe/theme/classic-dark.css", last_modified: "2026-09-29T11:16:08Z", digest: missing }
  - { id: plugin-listener, resource: "@milkdown/plugin-listener", last_modified: "2026-09-29T11:16:08Z", digest: missing }
  - { id: CodeEditor, resource: frontend/src/components/CodeEditor.svelte, last_modified: "2026-04-22T20:32:12+10:00", digest: 1bb474a9cd8f0465 }
  - { id: EditorRouter, resource: frontend/src/components/EditorRouter.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 86fac6035745e901 }
  - { id: ImageViewer, resource: frontend/src/components/ImageViewer.svelte, last_modified: "2026-04-22T19:23:56+10:00", digest: 4abd3a97e70be2b9 }
  - { id: MarkdownEditor, resource: frontend/src/components/MarkdownEditor.svelte, last_modified: "2026-04-23T11:20:33+10:00", digest: 39b9b94efde879aa }
  - { id: ImageViewer.test, resource: frontend/src/components/__tests__/ImageViewer.test.ts, last_modified: "2026-04-09T12:11:12+10:00", digest: 53d820a8d332d63b }
  - { id: imageViewerUtils, resource: frontend/src/components/imageViewerUtils.ts, last_modified: "2026-04-09T12:11:12+10:00", digest: 2bec4b44eedc3260 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
  - { id: panzoom, resource: panzoom, last_modified: "2026-09-29T11:16:08Z", digest: missing }
---

# Files
- `@milkdown/crepe`
- `@milkdown/crepe/theme/classic-dark.css`
- `@milkdown/plugin-listener`
- `frontend/src/components/CodeEditor.svelte`
- `frontend/src/components/EditorRouter.svelte`
- `frontend/src/components/ImageViewer.svelte`
- `frontend/src/components/MarkdownEditor.svelte`
- `frontend/src/components/__tests__/ImageViewer.test.ts`
- `frontend/src/components/imageViewerUtils.ts`
- `frontend/wailsjs/go/main/App.js`
- `panzoom`

# Symbols
- @milkdown/crepe (@milkdown/crepe:)
- @milkdown/crepe/theme/classic-dark.css (@milkdown/crepe/theme/classic-dark.css:)
- @milkdown/plugin-listener (@milkdown/plugin-listener:)
- CodeEditor.svelte (frontend/src/components/CodeEditor.svelte:L1)
- handleKeydown() (frontend/src/components/CodeEditor.svelte:L117)
- getLanguage() (frontend/src/components/CodeEditor.svelte:L140)
- isDiffLine() (frontend/src/components/CodeEditor.svelte:L156)
- parseHunks() (frontend/src/components/CodeEditor.svelte:L170)
- getHunkForLine() (frontend/src/components/CodeEditor.svelte:L193)
- handleDiffLineEnter() (frontend/src/components/CodeEditor.svelte:L201)
- handleDiffLineLeave() (frontend/src/components/CodeEditor.svelte:L238)
- dismissTooltip() (frontend/src/components/CodeEditor.svelte:L242)
- loadFile() (frontend/src/components/CodeEditor.svelte:L52)
- startEditing() (frontend/src/components/CodeEditor.svelte:L80)
- handleInput() (frontend/src/components/CodeEditor.svelte:L87)
- scheduleSave() (frontend/src/components/CodeEditor.svelte:L92)
- doSave() (frontend/src/components/CodeEditor.svelte:L98)
- EditorRouter.svelte (frontend/src/components/EditorRouter.svelte:L1)
- ImageViewer.svelte (frontend/src/components/ImageViewer.svelte:L1)
- MarkdownEditor.svelte (frontend/src/components/MarkdownEditor.svelte:L1)
- ImageViewer.test.ts (frontend/src/components/__tests__/ImageViewer.test.ts:L1)
- ImageViewerState (frontend/src/components/imageViewerUtils.ts:L1)
- imageViewerUtils.ts (frontend/src/components/imageViewerUtils.ts:L1)
- isDataUri() (frontend/src/components/imageViewerUtils.ts:L12)
- buildImagePath() (frontend/src/components/imageViewerUtils.ts:L6)
- ExplainDiffHunk() (frontend/wailsjs/go/main/App.js:L29)
- panzoom (panzoom:)

# Depends on
- [App.js](/modules/app-js.md)
- [MarkdownEditor.test.ts](/modules/markdowneditor-test-ts.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [SpawnRefactorPlan](/modules/spawnrefactorplan.md)
- [Story: meditor-01 — EditorRouter -- Extension-Based Editor Switching](/modules/story-meditor-01-editorrouter-extension-based-editor-switching.md)
- [svelte](/modules/svelte.md)
- [vitest](/modules/vitest.md)
- [WriteFile](/modules/writefile.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
