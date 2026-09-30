---
type: Module
title: CodeEditor.svelte
description: "Graphify community 257: frontend/src/components/CodeEditor.svelte, frontend/src/components/NewRepoModal.svelte, frontend/wailsjs/go/main/App.js"
resource: frontend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: CodeEditor, resource: frontend/src/components/CodeEditor.svelte, last_modified: "2026-04-22T20:32:12+10:00", digest: 1bb474a9cd8f0465 }
  - { id: NewRepoModal, resource: frontend/src/components/NewRepoModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 88f99d701a9fb7ec }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
---

# Files
- `frontend/src/components/CodeEditor.svelte`
- `frontend/src/components/NewRepoModal.svelte`
- `frontend/wailsjs/go/main/App.js`

# Symbols
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
- NewRepoModal.svelte (frontend/src/components/NewRepoModal.svelte:L1)
- selected (frontend/src/components/NewRepoModal.svelte:L91)
- ExplainDiffHunk() (frontend/wailsjs/go/main/App.js:L29)

# Depends on
- [App.js](/modules/app-js.md)
- [InputResponseModal.svelte](/modules/inputresponsemodal-svelte.md)
- [runtime.js](/modules/runtime-js.md)
- [StreamAdvice](/modules/streamadvice.md)
- [WriteFile](/modules/writefile.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
