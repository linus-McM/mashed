---
type: Module
title: InputResponseModal.svelte
description: "Graphify community 43: @milkdown/crepe, @milkdown/crepe/theme/classic-dark.css, @milkdown/plugin-listener, docs/stories/ui-ast-U7-decision-group.md, docs/stories/ui-ast-U8-sprint-report.md, frontend/s"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: crepe, resource: "@milkdown/crepe", last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: classic-dark, resource: "@milkdown/crepe/theme/classic-dark.css", last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: plugin-listener, resource: "@milkdown/plugin-listener", last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: ui-ast-U7-decision-group, resource: docs/stories/ui-ast-U7-decision-group.md, last_modified: "2026-04-22T11:51:34+10:00", digest: 2aab4f5ddd2f6bdb }
  - { id: ui-ast-U8-sprint-report, resource: docs/stories/ui-ast-U8-sprint-report.md, last_modified: "2026-04-22T12:58:08+10:00", digest: f3eb1e6e107bf539 }
  - { id: EditorRouter, resource: frontend/src/components/EditorRouter.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 86fac6035745e901 }
  - { id: ImageViewer, resource: frontend/src/components/ImageViewer.svelte, last_modified: "2026-04-22T19:23:56+10:00", digest: 4abd3a97e70be2b9 }
  - { id: MarkdownEditor, resource: frontend/src/components/MarkdownEditor.svelte, last_modified: "2026-09-30T00:51:37+10:00", digest: 68435a27812c07b0 }
  - { id: AstNode, resource: frontend/src/components/bmad/AstNode.svelte, last_modified: "2026-04-22T11:51:34+10:00", digest: ecea97f2b233cfbb }
  - { id: ComparisonTable, resource: frontend/src/components/bmad/ComparisonTable.svelte, last_modified: "2026-04-22T08:19:28+10:00", digest: 3ddc218b69a8b7db }
  - { id: DecisionGroup, resource: frontend/src/components/bmad/DecisionGroup.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 5d8416b6d4590e25 }
  - { id: DiagnosticsChip, resource: frontend/src/components/bmad/DiagnosticsChip.svelte, last_modified: "2026-04-22T20:27:37+10:00", digest: 132e44c9d91ec485 }
  - { id: HintBanner, resource: frontend/src/components/bmad/HintBanner.svelte, last_modified: "2026-04-22T08:19:28+10:00", digest: 80a8e54b196470fb }
  - { id: InputResponseModal, resource: frontend/src/components/bmad/InputResponseModal.svelte, last_modified: "2026-04-28T12:36:05+10:00", digest: 70963c4ec4be0aac }
  - { id: SummaryCard, resource: frontend/src/components/bmad/SummaryCard.svelte, last_modified: "2026-04-22T08:19:28+10:00", digest: cb815de294ae47a0 }
  - { id: TranscriptPane, resource: frontend/src/components/bmad/TranscriptPane.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 2cfde25c6a993123 }
  - { id: linkSanitiser.test, resource: frontend/src/components/bmad/__tests__/linkSanitiser.test.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: 4e352b3c3d0f7983 }
  - { id: ApprovalWidget, resource: frontend/src/components/bmad/inputWidgets/ApprovalWidget.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 0fc07df52f3669be }
  - { id: ChoiceWidget, resource: frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte, last_modified: "2026-04-28T12:36:05+10:00", digest: 627fa216a773bbcc }
  - { id: FreeTextWidget, resource: frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: e4781c45f272665c }
  - { id: JsonInputWidget, resource: frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: e0e8b47fab69e1f8 }
  - { id: linkSanitiser, resource: frontend/src/components/bmad/linkSanitiser.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: c487727219545c6a }
  - { id: errorMessage.test, resource: frontend/src/lib/errorMessage.test.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 8145a9da5a309eb2 }
  - { id: errorMessage, resource: frontend/src/lib/errorMessage.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 01d52ec3146aae0e }
  - { id: transcript, resource: frontend/src/types/transcript.ts, last_modified: "2026-04-22T20:18:54+10:00", digest: 3a79aee54e6f6bf3 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
  - { id: runtime, resource: frontend/wailsjs/runtime/runtime.js, last_modified: "2026-05-07T09:55:31+10:00", digest: e25fe86d3c590de7 }
  - { id: panzoom, resource: panzoom, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: store, resource: svelte/store, last_modified: "2026-09-29T15:22:13Z", digest: missing }
---

# Files
- `@milkdown/crepe`
- `@milkdown/crepe/theme/classic-dark.css`
- `@milkdown/plugin-listener`
- `docs/stories/ui-ast-U7-decision-group.md`
- `docs/stories/ui-ast-U8-sprint-report.md`
- `frontend/src/components/EditorRouter.svelte`
- `frontend/src/components/ImageViewer.svelte`
- `frontend/src/components/MarkdownEditor.svelte`
- `frontend/src/components/bmad/AstNode.svelte`
- `frontend/src/components/bmad/ComparisonTable.svelte`
- `frontend/src/components/bmad/DecisionGroup.svelte`
- `frontend/src/components/bmad/DiagnosticsChip.svelte`
- `frontend/src/components/bmad/HintBanner.svelte`
- `frontend/src/components/bmad/InputResponseModal.svelte`
- `frontend/src/components/bmad/SummaryCard.svelte`
- `frontend/src/components/bmad/TranscriptPane.svelte`
- `frontend/src/components/bmad/__tests__/linkSanitiser.test.ts`
- `frontend/src/components/bmad/inputWidgets/ApprovalWidget.svelte`
- `frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte`
- `frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte`
- `frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte`
- `frontend/src/components/bmad/linkSanitiser.ts`
- `frontend/src/lib/errorMessage.test.ts`
- `frontend/src/lib/errorMessage.ts`
- `frontend/src/types/transcript.ts`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/runtime/runtime.js`
- `panzoom`
- `svelte/store`

# Symbols
- @milkdown/crepe (@milkdown/crepe:)
- @milkdown/crepe/theme/classic-dark.css (@milkdown/crepe/theme/classic-dark.css:)
- @milkdown/plugin-listener (@milkdown/plugin-listener:)
- 4. Interaction model (docs/stories/ui-ast-U7-decision-group.md:L424)
- Follow-up debt (non-blocking, raised by reviewers) (docs/stories/ui-ast-U8-sprint-report.md:L95)
- EditorRouter.svelte (frontend/src/components/EditorRouter.svelte:L1)
- ImageViewer.svelte (frontend/src/components/ImageViewer.svelte:L1)
- MarkdownEditor.svelte (frontend/src/components/MarkdownEditor.svelte:L1)
- AstNode.svelte (frontend/src/components/bmad/AstNode.svelte:L1)
- ComparisonTable.svelte (frontend/src/components/bmad/ComparisonTable.svelte:L1)
- DecisionGroup.svelte (frontend/src/components/bmad/DecisionGroup.svelte:L1)
- onValue() (frontend/src/components/bmad/DecisionGroup.svelte:L46)
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
- SummaryCard.svelte (frontend/src/components/bmad/SummaryCard.svelte:L1)
- TranscriptPane.svelte (frontend/src/components/bmad/TranscriptPane.svelte:L1)
- linkSanitiser.test.ts (frontend/src/components/bmad/__tests__/linkSanitiser.test.ts:L1)
- ApprovalWidget.svelte (frontend/src/components/bmad/inputWidgets/ApprovalWidget.svelte:L1)
- ChoiceWidget.svelte (frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte:L1)
- selected (frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte:L105)
- submit() (frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte:L75)
- FreeTextWidget.svelte (frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte:L1)
- submit() (frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte:L31)
- onKeydown() (frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte:L37)
- JsonInputWidget.svelte (frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte:L1)
- submit() (frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte:L43)
- onKeydown() (frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte:L49)
- linkSanitiser.ts (frontend/src/components/bmad/linkSanitiser.ts:L1)
- REJECTED_SCHEMES (frontend/src/components/bmad/linkSanitiser.ts:L12)
- REJECTED_PROTOCOL_RE (frontend/src/components/bmad/linkSanitiser.ts:L13)
- sanitizeUrl() (frontend/src/components/bmad/linkSanitiser.ts:L15)
- UNSAFE_URL_IN_TEXT (frontend/src/components/bmad/linkSanitiser.ts:L31)
- scrubUnsafeUrls() (frontend/src/components/bmad/linkSanitiser.ts:L35)
- errorMessage.test.ts (frontend/src/lib/errorMessage.test.ts:L1)
- CustomError (frontend/src/lib/errorMessage.test.ts:L10)
- errorMessage.ts (frontend/src/lib/errorMessage.ts:L1)
- errorMessage() (frontend/src/lib/errorMessage.ts:L16)
- transcript.ts (frontend/src/types/transcript.ts:L1)
- TurnRole (frontend/src/types/transcript.ts:L13)
- normaliseInteractiveTurn() (frontend/src/types/transcript.ts:L34)
- GetInteractiveTranscript() (frontend/wailsjs/go/main/App.js:L85)
- EventsOn() (frontend/wailsjs/runtime/runtime.js:L43)
- panzoom (panzoom:)
- svelte/store (svelte/store:)

# Depends on
- [App.js](/modules/app-js.md)
- [EventsEmit](/modules/eventsemit.md)
- [imageViewerUtils.ts](/modules/imageviewerutils-ts.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [MarkdownEditor.test.ts](/modules/markdowneditor-test-ts.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [ref_node_path](/modules/ref-node-path.md)
- [runtime.js](/modules/runtime-js.md)
- [Story: meditor-01 — EditorRouter -- Extension-Based Editor Switching](/modules/story-meditor-01-editorrouter-extension-based-editor-switching.md)
- [svelte](/modules/svelte.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [uiAst.ts](/modules/uiast-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [svelte](/modules/svelte.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
