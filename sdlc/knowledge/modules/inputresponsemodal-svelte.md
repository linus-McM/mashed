---
type: Module
title: InputResponseModal.svelte
description: "Graphify community 159: docs/stories/bmad-interactive-02-executor-routing.md, docs/stories/ui-ast-U7-decision-group.md, docs/stories/ui-ast-U8-sprint-report.md, frontend/src/components/bmad/InputRespo"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: bmad-interactive-02-executor-routing, resource: docs/stories/bmad-interactive-02-executor-routing.md, last_modified: "2026-04-20T13:25:48+10:00", digest: 4f2046e0e690ce95 }
  - { id: ui-ast-U7-decision-group, resource: docs/stories/ui-ast-U7-decision-group.md, last_modified: "2026-04-22T11:51:34+10:00", digest: 2aab4f5ddd2f6bdb }
  - { id: ui-ast-U8-sprint-report, resource: docs/stories/ui-ast-U8-sprint-report.md, last_modified: "2026-04-22T12:58:08+10:00", digest: f3eb1e6e107bf539 }
  - { id: InputResponseModal, resource: frontend/src/components/bmad/InputResponseModal.svelte, last_modified: "2026-04-28T12:36:05+10:00", digest: 70963c4ec4be0aac }
  - { id: TranscriptPane, resource: frontend/src/components/bmad/TranscriptPane.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 2cfde25c6a993123 }
  - { id: errorMessage.test, resource: frontend/src/lib/errorMessage.test.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 8145a9da5a309eb2 }
  - { id: errorMessage, resource: frontend/src/lib/errorMessage.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 01d52ec3146aae0e }
  - { id: transcript, resource: frontend/src/types/transcript.ts, last_modified: "2026-04-22T20:18:54+10:00", digest: 3a79aee54e6f6bf3 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/stories/bmad-interactive-02-executor-routing.md`
- `docs/stories/ui-ast-U7-decision-group.md`
- `docs/stories/ui-ast-U8-sprint-report.md`
- `frontend/src/components/bmad/InputResponseModal.svelte`
- `frontend/src/components/bmad/TranscriptPane.svelte`
- `frontend/src/lib/errorMessage.test.ts`
- `frontend/src/lib/errorMessage.ts`
- `frontend/src/types/transcript.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Acceptance Criteria (docs/stories/bmad-interactive-02-executor-routing.md:L131)
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
- user (frontend/src/components/bmad/TranscriptPane.svelte:L61)
- errorMessage.test.ts (frontend/src/lib/errorMessage.test.ts:L1)
- CustomError (frontend/src/lib/errorMessage.test.ts:L10)
- errorMessage() (frontend/src/lib/errorMessage.ts:L16)
- transcript.ts (frontend/src/types/transcript.ts:L1)
- TurnRole (frontend/src/types/transcript.ts:L13)
- normaliseInteractiveTurn() (frontend/src/types/transcript.ts:L34)
- GetInteractiveTranscript() (frontend/wailsjs/go/main/App.js:L85)

# Depends on
- [App.js](/modules/app-js.md)
- [mountSvelte.ts](/modules/mountsvelte-ts.md)
- [runtime.js](/modules/runtime-js.md)
- [Terminal.svelte](/modules/terminal-svelte.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [uiAst.ts](/modules/uiast-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- [svelte](/modules/svelte.md)

# Features
- no feature plan names these files
