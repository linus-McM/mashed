---
type: Module
title: InputResponseModal.svelte
description: "Graphify community 112: docs/stories/ui-ast-U7-decision-group.md, docs/stories/ui-ast-U8-sprint-report.md, frontend/src/components/bmad/InputResponseModal.svelte, frontend/src/components/bmad/Transcri"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:14Z" }
stale_after: "2026-10-13T12:23:14Z"
source_commit: ab1f2eb4ec65fc2b1fce15503c11f0e310758404
sources:
  - { id: ui-ast-U7-decision-group, resource: docs/stories/ui-ast-U7-decision-group.md, last_modified: "2026-04-22T11:51:34+10:00", digest: 2aab4f5ddd2f6bdb }
  - { id: ui-ast-U8-sprint-report, resource: docs/stories/ui-ast-U8-sprint-report.md, last_modified: "2026-04-22T12:58:08+10:00", digest: f3eb1e6e107bf539 }
  - { id: InputResponseModal, resource: frontend/src/components/bmad/InputResponseModal.svelte, last_modified: "2026-04-28T12:36:05+10:00", digest: 70963c4ec4be0aac }
  - { id: TranscriptPane, resource: frontend/src/components/bmad/TranscriptPane.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 2cfde25c6a993123 }
  - { id: errorMessage.test, resource: frontend/src/lib/errorMessage.test.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 8145a9da5a309eb2 }
  - { id: errorMessage, resource: frontend/src/lib/errorMessage.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 01d52ec3146aae0e }
  - { id: transcript, resource: frontend/src/types/transcript.ts, last_modified: "2026-04-22T20:18:54+10:00", digest: 3a79aee54e6f6bf3 }
  - { id: BranchModal, resource: frontend/src/views/BranchModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 6d3c58fc0fa1902c }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/stories/ui-ast-U7-decision-group.md`
- `docs/stories/ui-ast-U8-sprint-report.md`
- `frontend/src/components/bmad/InputResponseModal.svelte`
- `frontend/src/components/bmad/TranscriptPane.svelte`
- `frontend/src/lib/errorMessage.test.ts`
- `frontend/src/lib/errorMessage.ts`
- `frontend/src/types/transcript.ts`
- `frontend/src/views/BranchModal.svelte`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- 4. Interaction model (docs/stories/ui-ast-U7-decision-group.md:L424)
- Follow-up debt (non-blocking, raised by reviewers) (docs/stories/ui-ast-U8-sprint-report.md:L95)
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
- errorMessage.test.ts (frontend/src/lib/errorMessage.test.ts:L1)
- CustomError (frontend/src/lib/errorMessage.test.ts:L10)
- errorMessage() (frontend/src/lib/errorMessage.ts:L16)
- transcript.ts (frontend/src/types/transcript.ts:L1)
- TurnRole (frontend/src/types/transcript.ts:L13)
- normaliseInteractiveTurn() (frontend/src/types/transcript.ts:L34)
- create() (frontend/src/views/BranchModal.svelte:L66)
- cancel() (frontend/src/views/BranchModal.svelte:L79)
- handleKeydown() (frontend/src/views/BranchModal.svelte:L84)
- GitCreateBranch() (frontend/wailsjs/go/main/App.js:L137)
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
- [Repo health remediation](/features/repo-health-remediation.md)
