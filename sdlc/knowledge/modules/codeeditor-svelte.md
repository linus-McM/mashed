---
type: Module
title: CodeEditor.svelte
description: "Graphify community 248: docs/stories/ui-ast-U7-decision-group.md, docs/stories/ui-ast-U8-sprint-report.md, frontend/src/components/CodeEditor.svelte, frontend/src/components/bmad/InputResponseModal.sv"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: ui-ast-U7-decision-group, resource: docs/stories/ui-ast-U7-decision-group.md, last_modified: "2026-04-22T11:51:34+10:00", digest: 2aab4f5ddd2f6bdb }
  - { id: ui-ast-U8-sprint-report, resource: docs/stories/ui-ast-U8-sprint-report.md, last_modified: "2026-04-22T12:58:08+10:00", digest: f3eb1e6e107bf539 }
  - { id: CodeEditor, resource: frontend/src/components/CodeEditor.svelte, last_modified: "2026-04-22T20:32:12+10:00", digest: 1bb474a9cd8f0465 }
  - { id: InputResponseModal, resource: frontend/src/components/bmad/InputResponseModal.svelte, last_modified: "2026-04-28T12:36:05+10:00", digest: 70963c4ec4be0aac }
  - { id: errorMessage.test, resource: frontend/src/lib/errorMessage.test.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 8145a9da5a309eb2 }
  - { id: errorMessage, resource: frontend/src/lib/errorMessage.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 01d52ec3146aae0e }
  - { id: transcript, resource: frontend/src/types/transcript.ts, last_modified: "2026-04-22T20:18:54+10:00", digest: 3a79aee54e6f6bf3 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/stories/ui-ast-U7-decision-group.md`
- `docs/stories/ui-ast-U8-sprint-report.md`
- `frontend/src/components/CodeEditor.svelte`
- `frontend/src/components/bmad/InputResponseModal.svelte`
- `frontend/src/lib/errorMessage.test.ts`
- `frontend/src/lib/errorMessage.ts`
- `frontend/src/types/transcript.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- 4. Interaction model (docs/stories/ui-ast-U7-decision-group.md:L424)
- Follow-up debt (non-blocking, raised by reviewers) (docs/stories/ui-ast-U8-sprint-report.md:L95)
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
- onSend() (frontend/src/components/bmad/InputResponseModal.svelte:L146)
- loadTranscript() (frontend/src/components/bmad/InputResponseModal.svelte:L193)
- onWidgetSubmit() (frontend/src/components/bmad/InputResponseModal.svelte:L247)
- triggerShake() (frontend/src/components/bmad/InputResponseModal.svelte:L72)
- errorMessage.test.ts (frontend/src/lib/errorMessage.test.ts:L1)
- CustomError (frontend/src/lib/errorMessage.test.ts:L10)
- errorMessage() (frontend/src/lib/errorMessage.ts:L16)
- normaliseInteractiveTurn() (frontend/src/types/transcript.ts:L34)
- ExplainDiffHunk() (frontend/wailsjs/go/main/App.js:L29)
- GetInteractiveTranscript() (frontend/wailsjs/go/main/App.js:L85)

# Depends on
- [10. Wails Bindings (Go → Svelte API)](/modules/10-wails-bindings-go-svelte-api.md)
- [App.js](/modules/app-js.md)
- [StreamAdvice](/modules/streamadvice.md)
- [vitest](/modules/vitest.md)
- [WriteFile](/modules/writefile.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
