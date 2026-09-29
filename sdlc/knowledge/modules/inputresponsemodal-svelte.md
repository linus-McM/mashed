---
type: Module
title: InputResponseModal.svelte
description: "Graphify community 53: docs/stories/ui-ast-U7-decision-group.md, docs/stories/ui-ast-U8-sprint-report.md, frontend/src/components/bmad/DiagnosticsChip.svelte, frontend/src/components/bmad/InputRespons"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: ui-ast-U7-decision-group, resource: docs/stories/ui-ast-U7-decision-group.md, last_modified: "2026-09-29T07:07:25Z", digest: 2aab4f5ddd2f6bdb }
  - { id: ui-ast-U8-sprint-report, resource: docs/stories/ui-ast-U8-sprint-report.md, last_modified: "2026-09-29T07:07:25Z", digest: f3eb1e6e107bf539 }
  - { id: DiagnosticsChip, resource: frontend/src/components/bmad/DiagnosticsChip.svelte, last_modified: "2026-09-29T07:07:25Z", digest: 132e44c9d91ec485 }
  - { id: InputResponseModal, resource: frontend/src/components/bmad/InputResponseModal.svelte, last_modified: "2026-09-29T07:07:25Z", digest: 70963c4ec4be0aac }
  - { id: TranscriptPane, resource: frontend/src/components/bmad/TranscriptPane.svelte, last_modified: "2026-09-29T07:07:25Z", digest: 2cfde25c6a993123 }
  - { id: DiagnosticsChip.test, resource: frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts, last_modified: "2026-09-29T07:07:25Z", digest: 5a1b2aad12debba0 }
  - { id: errorMessage.test, resource: frontend/src/lib/errorMessage.test.ts, last_modified: "2026-09-29T07:07:25Z", digest: 8145a9da5a309eb2 }
  - { id: errorMessage, resource: frontend/src/lib/errorMessage.ts, last_modified: "2026-09-29T07:07:25Z", digest: 01d52ec3146aae0e }
  - { id: transcript, resource: frontend/src/types/transcript.ts, last_modified: "2026-09-29T07:07:25Z", digest: 3a79aee54e6f6bf3 }
  - { id: uiAst, resource: frontend/src/types/uiAst.ts, last_modified: "2026-09-29T07:07:25Z", digest: 67c7b5fc9513892e }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-29T07:07:25Z", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/stories/ui-ast-U7-decision-group.md`
- `docs/stories/ui-ast-U8-sprint-report.md`
- `frontend/src/components/bmad/DiagnosticsChip.svelte`
- `frontend/src/components/bmad/InputResponseModal.svelte`
- `frontend/src/components/bmad/TranscriptPane.svelte`
- `frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts`
- `frontend/src/lib/errorMessage.test.ts`
- `frontend/src/lib/errorMessage.ts`
- `frontend/src/types/transcript.ts`
- `frontend/src/types/uiAst.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Tasks / Subtasks (docs/stories/ui-ast-U7-decision-group.md:L329)
- 4. Interaction model (docs/stories/ui-ast-U7-decision-group.md:L424)
- Follow-up debt (non-blocking, raised by reviewers) (docs/stories/ui-ast-U8-sprint-report.md:L95)
- DiagnosticsChip.svelte (frontend/src/components/bmad/DiagnosticsChip.svelte:L1)
- InputResponseModal.svelte (frontend/src/components/bmad/InputResponseModal.svelte:L1)
- onGroupActivate() (frontend/src/components/bmad/InputResponseModal.svelte:L141)
- onSend() (frontend/src/components/bmad/InputResponseModal.svelte:L146)
- loadTranscript() (frontend/src/components/bmad/InputResponseModal.svelte:L193)
- onWidgetSubmit() (frontend/src/components/bmad/InputResponseModal.svelte:L247)
- close() (frontend/src/components/bmad/InputResponseModal.svelte:L263)
- onKeydown() (frontend/src/components/bmad/InputResponseModal.svelte:L268)
- onOverlayClick() (frontend/src/components/bmad/InputResponseModal.svelte:L282)
- triggerShake() (frontend/src/components/bmad/InputResponseModal.svelte:L72)
- TranscriptPane.svelte (frontend/src/components/bmad/TranscriptPane.svelte:L1)
- DiagnosticsChip.test.ts (frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts:L1)
- mount (frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts:L12)
- render() (frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts:L13)
- findUntrusted() (frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts:L18)
- findNotesChip() (frontend/src/components/bmad/__tests__/DiagnosticsChip.test.ts:L23)
- errorMessage.test.ts (frontend/src/lib/errorMessage.test.ts:L1)
- CustomError (frontend/src/lib/errorMessage.test.ts:L10)
- errorMessage() (frontend/src/lib/errorMessage.ts:L16)
- transcript.ts (frontend/src/types/transcript.ts:L1)
- TurnRole (frontend/src/types/transcript.ts:L13)
- normaliseInteractiveTurn() (frontend/src/types/transcript.ts:L34)
- uiAst.ts (frontend/src/types/uiAst.ts:L1)
- UINodeType (frontend/src/types/uiAst.ts:L18)
- HintTone (frontend/src/types/uiAst.ts:L20)
- KnownWidgetType (frontend/src/types/uiAst.ts:L22)
- WidgetType (frontend/src/types/uiAst.ts:L23)
- UINode (frontend/src/types/uiAst.ts:L41)
- KnownUINodeType (frontend/src/types/uiAst.ts:L8)
- GetInteractiveTranscript() (frontend/wailsjs/go/main/App.js:L85)

# Depends on
- [App.js](/modules/app-js.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [runtime.js](/modules/runtime-js.md)
- [svelte](/modules/svelte.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- [svelte](/modules/svelte.md)

# Features
- no feature plan names these files
