---
type: Module
title: InputResponseModal.svelte
description: "Graphify community 43: docs/stories/ui-ast-U7-decision-group.md, docs/stories/ui-ast-U8-sprint-report.md, frontend/src/components/bmad/AstNode.svelte, frontend/src/components/bmad/DecisionGroup.svelte"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: ui-ast-U7-decision-group, resource: docs/stories/ui-ast-U7-decision-group.md, last_modified: "2026-04-22T11:51:34+10:00", digest: 2aab4f5ddd2f6bdb }
  - { id: ui-ast-U8-sprint-report, resource: docs/stories/ui-ast-U8-sprint-report.md, last_modified: "2026-04-22T12:58:08+10:00", digest: f3eb1e6e107bf539 }
  - { id: AstNode, resource: frontend/src/components/bmad/AstNode.svelte, last_modified: "2026-04-22T11:51:34+10:00", digest: ecea97f2b233cfbb }
  - { id: DecisionGroup, resource: frontend/src/components/bmad/DecisionGroup.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 5d8416b6d4590e25 }
  - { id: InputResponseModal, resource: frontend/src/components/bmad/InputResponseModal.svelte, last_modified: "2026-04-28T12:36:05+10:00", digest: 70963c4ec4be0aac }
  - { id: MultiFileLoaderNode, resource: frontend/src/components/bmad/MultiFileLoaderNode.svelte, last_modified: "2026-04-22T19:45:14+10:00", digest: 593fafc1b9a89f9f }
  - { id: TranscriptPane, resource: frontend/src/components/bmad/TranscriptPane.svelte, last_modified: "2026-04-22T20:18:54+10:00", digest: 2cfde25c6a993123 }
  - { id: ChoiceWidget, resource: frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte, last_modified: "2026-04-28T12:36:05+10:00", digest: 627fa216a773bbcc }
  - { id: FreeTextWidget, resource: frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: e4781c45f272665c }
  - { id: JsonInputWidget, resource: frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: e0e8b47fab69e1f8 }
  - { id: MultiChoiceWidget, resource: frontend/src/components/bmad/inputWidgets/MultiChoiceWidget.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: a5b9208bb3c31b33 }
  - { id: errorMessage.test, resource: frontend/src/lib/errorMessage.test.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 8145a9da5a309eb2 }
  - { id: errorMessage, resource: frontend/src/lib/errorMessage.ts, last_modified: "2026-04-22T17:53:53+10:00", digest: 01d52ec3146aae0e }
  - { id: interactiveInput, resource: frontend/src/stores/interactiveInput.ts, last_modified: "2026-04-22T11:51:34+10:00", digest: 58e33ed1241fb1fd }
  - { id: transcript, resource: frontend/src/types/transcript.ts, last_modified: "2026-04-22T20:18:54+10:00", digest: 3a79aee54e6f6bf3 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
  - { id: store, resource: svelte/store, last_modified: "2026-09-29T14:46:21Z", digest: missing }
---

# Files
- `docs/stories/ui-ast-U7-decision-group.md`
- `docs/stories/ui-ast-U8-sprint-report.md`
- `frontend/src/components/bmad/AstNode.svelte`
- `frontend/src/components/bmad/DecisionGroup.svelte`
- `frontend/src/components/bmad/InputResponseModal.svelte`
- `frontend/src/components/bmad/MultiFileLoaderNode.svelte`
- `frontend/src/components/bmad/TranscriptPane.svelte`
- `frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte`
- `frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte`
- `frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte`
- `frontend/src/components/bmad/inputWidgets/MultiChoiceWidget.svelte`
- `frontend/src/lib/errorMessage.test.ts`
- `frontend/src/lib/errorMessage.ts`
- `frontend/src/stores/interactiveInput.ts`
- `frontend/src/types/transcript.ts`
- `frontend/wailsjs/go/main/App.js`
- `svelte/store`

# Symbols
- Tasks / Subtasks (docs/stories/ui-ast-U7-decision-group.md:L329)
- 4. Interaction model (docs/stories/ui-ast-U7-decision-group.md:L424)
- Follow-up debt (non-blocking, raised by reviewers) (docs/stories/ui-ast-U8-sprint-report.md:L95)
- AstNode.svelte (frontend/src/components/bmad/AstNode.svelte:L1)
- DecisionGroup.svelte (frontend/src/components/bmad/DecisionGroup.svelte:L1)
- onValue() (frontend/src/components/bmad/DecisionGroup.svelte:L46)
- InputResponseModal.svelte (frontend/src/components/bmad/InputResponseModal.svelte:L1)
- onGroupActivate() (frontend/src/components/bmad/InputResponseModal.svelte:L141)
- onSend() (frontend/src/components/bmad/InputResponseModal.svelte:L146)
- loadTranscript() (frontend/src/components/bmad/InputResponseModal.svelte:L193)
- onWidgetSubmit() (frontend/src/components/bmad/InputResponseModal.svelte:L247)
- close() (frontend/src/components/bmad/InputResponseModal.svelte:L263)
- onKeydown() (frontend/src/components/bmad/InputResponseModal.svelte:L268)
- onOverlayClick() (frontend/src/components/bmad/InputResponseModal.svelte:L282)
- triggerShake() (frontend/src/components/bmad/InputResponseModal.svelte:L72)
- i() (frontend/src/components/bmad/MultiFileLoaderNode.svelte:L34)
- TranscriptPane.svelte (frontend/src/components/bmad/TranscriptPane.svelte:L1)
- ChoiceWidget.svelte (frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte:L1)
- selected (frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte:L105)
- submit() (frontend/src/components/bmad/inputWidgets/ChoiceWidget.svelte:L75)
- FreeTextWidget.svelte (frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte:L1)
- submit() (frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte:L31)
- onKeydown() (frontend/src/components/bmad/inputWidgets/FreeTextWidget.svelte:L37)
- JsonInputWidget.svelte (frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte:L1)
- submit() (frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte:L43)
- onKeydown() (frontend/src/components/bmad/inputWidgets/JsonInputWidget.svelte:L49)
- MultiChoiceWidget.svelte (frontend/src/components/bmad/inputWidgets/MultiChoiceWidget.svelte:L1)
- errorMessage.test.ts (frontend/src/lib/errorMessage.test.ts:L1)
- CustomError (frontend/src/lib/errorMessage.test.ts:L10)
- errorMessage() (frontend/src/lib/errorMessage.ts:L16)
- pushToast() (frontend/src/stores/interactiveInput.ts:L127)
- transcript.ts (frontend/src/types/transcript.ts:L1)
- TurnRole (frontend/src/types/transcript.ts:L13)
- normaliseInteractiveTurn() (frontend/src/types/transcript.ts:L34)
- GetInteractiveTranscript() (frontend/wailsjs/go/main/App.js:L85)
- svelte/store (svelte/store:)

# Depends on
- [App.js](/modules/app-js.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [runtime.js](/modules/runtime-js.md)
- [svelte](/modules/svelte.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [uiAst.ts](/modules/uiast-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
