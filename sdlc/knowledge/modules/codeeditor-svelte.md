---
type: Module
title: CodeEditor.svelte
description: "Graphify community 248: frontend/src/components/CodeEditor.svelte, frontend/wailsjs/go/main/App.js"
resource: frontend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:14Z" }
stale_after: "2026-10-13T12:23:14Z"
source_commit: ab1f2eb4ec65fc2b1fce15503c11f0e310758404
sources:
  - { id: CodeEditor, resource: frontend/src/components/CodeEditor.svelte, last_modified: "2026-04-22T20:32:12+10:00", digest: 1bb474a9cd8f0465 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `frontend/src/components/CodeEditor.svelte`
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
- ExplainDiffHunk() (frontend/wailsjs/go/main/App.js:L29)

# Depends on
- [10. Wails Bindings (Go → Svelte API)](/modules/10-wails-bindings-go-svelte-api.md)
- [App.js](/modules/app-js.md)
- [InputResponseModal.svelte](/modules/inputresponsemodal-svelte.md)
- [svelte](/modules/svelte.md)
- [WriteFile](/modules/writefile.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
