---
type: Module
title: InputResponseModal.svelte
description: "Graphify community 43: docs/stories/ui-ast-U7-decision-group.md, docs/stories/ui-ast-U8-sprint-report.md, frontend/src/components/bmad/AstNode.svelte, frontend/src/components/bmad/ComparisonTable.svel"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: ui-ast-U7-decision-group, resource: docs/stories/ui-ast-U7-decision-group.md, last_modified: "2026-04-22T11:51:34+10:00", digest: 2aab4f5ddd2f6bdb }
  - { id: ui-ast-U8-sprint-report, resource: docs/stories/ui-ast-U8-sprint-report.md, last_modified: "2026-04-22T12:58:08+10:00", digest: f3eb1e6e107bf539 }
  - { id: AstNode, resource: frontend/src/components/bmad/AstNode.svelte, last_modified: "2026-04-22T11:51:34+10:00", digest: ecea97f2b233cfbb }
  - { id: ComparisonTable, resource: frontend/src/components/bmad/ComparisonTable.svelte, last_modified: "2026-04-22T08:19:28+10:00", digest: 3ddc218b69a8b7db }
  - { id: DiagnosticsChip, resource: frontend/src/components/bmad/DiagnosticsChip.svelte, last_modified: "2026-04-22T20:27:37+10:00", digest: 132e44c9d91ec485 }
  - { id: HintBanner, resource: frontend/src/components/bmad/HintBanner.svelte, last_modified: "2026-04-22T08:19:28+10:00", digest: 80a8e54b196470fb }
  - { id: InputResponseModal, resource: frontend/src/components/bmad/InputResponseModal.svelte, last_modified: "2026-04-28T12:36:05+10:00", digest: 70963c4ec4be0aac }
  - { id: MarkdownBlock, resource: frontend/src/components/bmad/MarkdownBlock.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 3f0e52590dddfec2 }
  - { id: SummaryCard, resource: frontend/src/components/bmad/SummaryCard.svelte, last_modified: "2026-04-22T08:19:28+10:00", digest: cb815de294ae47a0 }
  - { id: TranscriptPane, resource: frontend/src/components/bmad/TranscriptPane.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 2cfde25c6a993123 }
  - { id: MarkdownBlock.test, resource: frontend/src/components/bmad/__tests__/MarkdownBlock.test.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: fffa845df450229f }
  - { id: linkSanitiser.test, resource: frontend/src/components/bmad/__tests__/linkSanitiser.test.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: 4e352b3c3d0f7983 }
  - { id: linkSanitiser, resource: frontend/src/components/bmad/linkSanitiser.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: c487727219545c6a }
  - { id: errorMessage.test, resource: frontend/src/lib/errorMessage.test.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 8145a9da5a309eb2 }
  - { id: errorMessage, resource: frontend/src/lib/errorMessage.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 01d52ec3146aae0e }
  - { id: transcript, resource: frontend/src/types/transcript.ts, last_modified: "2026-04-22T20:18:54+10:00", digest: 3a79aee54e6f6bf3 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
  - { id: markdown-it, resource: markdown-it, last_modified: "2026-09-29T20:55:29Z", digest: missing }
---

# Files
- `docs/stories/ui-ast-U7-decision-group.md`
- `docs/stories/ui-ast-U8-sprint-report.md`
- `frontend/src/components/bmad/AstNode.svelte`
- `frontend/src/components/bmad/ComparisonTable.svelte`
- `frontend/src/components/bmad/DiagnosticsChip.svelte`
- `frontend/src/components/bmad/HintBanner.svelte`
- `frontend/src/components/bmad/InputResponseModal.svelte`
- `frontend/src/components/bmad/MarkdownBlock.svelte`
- `frontend/src/components/bmad/SummaryCard.svelte`
- `frontend/src/components/bmad/TranscriptPane.svelte`
- `frontend/src/components/bmad/__tests__/MarkdownBlock.test.ts`
- `frontend/src/components/bmad/__tests__/linkSanitiser.test.ts`
- `frontend/src/components/bmad/linkSanitiser.ts`
- `frontend/src/lib/errorMessage.test.ts`
- `frontend/src/lib/errorMessage.ts`
- `frontend/src/types/transcript.ts`
- `frontend/wailsjs/go/main/App.js`
- `markdown-it`

# Symbols
- 4. Interaction model (docs/stories/ui-ast-U7-decision-group.md:L424)
- Follow-up debt (non-blocking, raised by reviewers) (docs/stories/ui-ast-U8-sprint-report.md:L95)
- AstNode.svelte (frontend/src/components/bmad/AstNode.svelte:L1)
- ComparisonTable.svelte (frontend/src/components/bmad/ComparisonTable.svelte:L1)
- DiagnosticsChip.svelte (frontend/src/components/bmad/DiagnosticsChip.svelte:L1)
- HintBanner.svelte (frontend/src/components/bmad/HintBanner.svelte:L1)
- InputResponseModal.svelte (frontend/src/components/bmad/InputResponseModal.svelte:L1)
- onGroupActivate() (frontend/src/components/bmad/InputResponseModal.svelte:L141)
- onSend() (frontend/src/components/bmad/InputResponseModal.svelte:L146)
- loadTranscript() (frontend/src/components/bmad/InputResponseModal.svelte:L193)
- onWidgetSubmit() (frontend/src/components/bmad/InputResponseModal.svelte:L247)
- close() (frontend/src/components/bmad/InputResponseModal.svelte:L263)
- onKeydown() (frontend/src/components/bmad/InputResponseModal.svelte:L268)
- onOverlayClick() (frontend/src/components/bmad/InputResponseModal.svelte:L282)
- triggerShake() (frontend/src/components/bmad/InputResponseModal.svelte:L72)
- MarkdownBlock.svelte (frontend/src/components/bmad/MarkdownBlock.svelte:L1)
- SummaryCard.svelte (frontend/src/components/bmad/SummaryCard.svelte:L1)
- TranscriptPane.svelte (frontend/src/components/bmad/TranscriptPane.svelte:L1)
- MarkdownBlock.test.ts (frontend/src/components/bmad/__tests__/MarkdownBlock.test.ts:L1)
- linkSanitiser.test.ts (frontend/src/components/bmad/__tests__/linkSanitiser.test.ts:L1)
- linkSanitiser.ts (frontend/src/components/bmad/linkSanitiser.ts:L1)
- REJECTED_SCHEMES (frontend/src/components/bmad/linkSanitiser.ts:L12)
- REJECTED_PROTOCOL_RE (frontend/src/components/bmad/linkSanitiser.ts:L13)
- sanitizeUrl() (frontend/src/components/bmad/linkSanitiser.ts:L15)
- UNSAFE_URL_IN_TEXT (frontend/src/components/bmad/linkSanitiser.ts:L31)
- scrubUnsafeUrls() (frontend/src/components/bmad/linkSanitiser.ts:L35)
- errorMessage.test.ts (frontend/src/lib/errorMessage.test.ts:L1)
- CustomError (frontend/src/lib/errorMessage.test.ts:L10)
- errorMessage() (frontend/src/lib/errorMessage.ts:L16)
- transcript.ts (frontend/src/types/transcript.ts:L1)
- TurnRole (frontend/src/types/transcript.ts:L13)
- normaliseInteractiveTurn() (frontend/src/types/transcript.ts:L34)
- GetInteractiveTranscript() (frontend/wailsjs/go/main/App.js:L85)
- markdown-it (markdown-it:)

# Depends on
- [App.js](/modules/app-js.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [runtime.js](/modules/runtime-js.md)
- [svelte](/modules/svelte.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [uiAst.ts](/modules/uiast-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- [EventsEmit](/modules/eventsemit.md)
- [svelte](/modules/svelte.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
