---
type: Module
title: vitest
description: "Graphify community 11: frontend/package.json, frontend/src/App.test.ts, frontend/src/components/bmad/InputResponseModal.test.ts, frontend/src/components/bmad/__tests__/CodeBlock.test.ts, frontend/src/"
resource: frontend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: package, resource: frontend/package.json, last_modified: "2026-04-22T16:28:45+10:00", digest: 5a2c6b5e36bbccae }
  - { id: App.test, resource: frontend/src/App.test.ts, last_modified: "2026-04-23T11:09:52+10:00", digest: 260d2fa32bf35b5d }
  - { id: InputResponseModal.test, resource: frontend/src/components/bmad/InputResponseModal.test.ts, last_modified: "2026-04-22T12:56:23+10:00", digest: dbef06f6be2ed227 }
  - { id: CodeBlock.test, resource: frontend/src/components/bmad/__tests__/CodeBlock.test.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: e491395879825947 }
  - { id: ApprovalWidget.test, resource: frontend/src/components/bmad/inputWidgets/ApprovalWidget.test.ts, last_modified: "2026-04-20T14:57:51+10:00", digest: dd83b6c89edd8eeb }
  - { id: ChoiceWidget.test, resource: frontend/src/components/bmad/inputWidgets/ChoiceWidget.test.ts, last_modified: "2026-04-20T14:57:51+10:00", digest: 3d4ad8158db13456 }
  - { id: FileInputWidget.test, resource: frontend/src/components/bmad/inputWidgets/FileInputWidget.test.ts, last_modified: "2026-04-20T14:57:51+10:00", digest: 85bc3c9b12efacd4 }
  - { id: FreeTextWidget.test, resource: frontend/src/components/bmad/inputWidgets/FreeTextWidget.test.ts, last_modified: "2026-04-20T14:57:51+10:00", digest: 5a6d885abc0c32fe }
  - { id: JsonInputWidget.test, resource: frontend/src/components/bmad/inputWidgets/JsonInputWidget.test.ts, last_modified: "2026-04-20T14:57:51+10:00", digest: be0fc7194d70532a }
  - { id: MultiChoiceWidget.test, resource: frontend/src/components/bmad/inputWidgets/MultiChoiceWidget.test.ts, last_modified: "2026-04-20T14:57:51+10:00", digest: 5d986cc8a7f48026 }
  - { id: markdownMenuSettings.test, resource: frontend/src/lib/stores/markdownMenuSettings.test.ts, last_modified: "2026-04-23T11:02:33+10:00", digest: 440326bc7b010a88 }
  - { id: markdownMenuSettings, resource: frontend/src/lib/stores/markdownMenuSettings.ts, last_modified: "2026-04-23T11:02:33+10:00", digest: 5fd83b95073decb2 }
  - { id: astResponses.test, resource: frontend/src/stores/__tests__/astResponses.test.ts, last_modified: "2026-04-22T11:51:34+10:00", digest: e2e9b800b1e40d74 }
  - { id: astResponses, resource: frontend/src/stores/astResponses.ts, last_modified: "2026-04-22T11:51:34+10:00", digest: cf393d4393fa3a1f }
  - { id: interactiveInput.test, resource: frontend/src/stores/interactiveInput.test.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: 0035d0dad2b7a72e }
  - { id: interactiveInput, resource: frontend/src/stores/interactiveInput.ts, last_modified: "2026-04-22T11:51:34+10:00", digest: 58e33ed1241fb1fd }
---

# Files
- `frontend/package.json`
- `frontend/src/App.test.ts`
- `frontend/src/components/bmad/InputResponseModal.test.ts`
- `frontend/src/components/bmad/__tests__/CodeBlock.test.ts`
- `frontend/src/components/bmad/inputWidgets/ApprovalWidget.test.ts`
- `frontend/src/components/bmad/inputWidgets/ChoiceWidget.test.ts`
- `frontend/src/components/bmad/inputWidgets/FileInputWidget.test.ts`
- `frontend/src/components/bmad/inputWidgets/FreeTextWidget.test.ts`
- `frontend/src/components/bmad/inputWidgets/JsonInputWidget.test.ts`
- `frontend/src/components/bmad/inputWidgets/MultiChoiceWidget.test.ts`
- `frontend/src/lib/stores/markdownMenuSettings.test.ts`
- `frontend/src/lib/stores/markdownMenuSettings.ts`
- `frontend/src/stores/__tests__/astResponses.test.ts`
- `frontend/src/stores/astResponses.ts`
- `frontend/src/stores/interactiveInput.test.ts`
- `frontend/src/stores/interactiveInput.ts`

# Symbols
- svelte (frontend/package.json:L19)
- vitest (frontend/package.json:L22)
- App.test.ts (frontend/src/App.test.ts:L1)
- DEFAULTS (frontend/src/App.test.ts:L24)
- APP_SVELTE_PATH (frontend/src/App.test.ts:L36)
- APP_SVELTE_SOURCE (frontend/src/App.test.ts:L40)
- InputResponseModal.test.ts (frontend/src/components/bmad/InputResponseModal.test.ts:L1)
- makePrompt() (frontend/src/components/bmad/InputResponseModal.test.ts:L34)
- astJson() (frontend/src/components/bmad/InputResponseModal.test.ts:L50)
- dgNode() (frontend/src/components/bmad/InputResponseModal.test.ts:L53)
- Mount (frontend/src/components/bmad/InputResponseModal.test.ts:L59)
- CodeBlock.test.ts (frontend/src/components/bmad/__tests__/CodeBlock.test.ts:L1)
- ApprovalWidget.test.ts (frontend/src/components/bmad/inputWidgets/ApprovalWidget.test.ts:L1)
- Mount (frontend/src/components/bmad/inputWidgets/ApprovalWidget.test.ts:L24)
- makePrompt() (frontend/src/components/bmad/inputWidgets/ApprovalWidget.test.ts:L9)
- ChoiceWidget.test.ts (frontend/src/components/bmad/inputWidgets/ChoiceWidget.test.ts:L1)
- Mount (frontend/src/components/bmad/inputWidgets/ChoiceWidget.test.ts:L24)
- makePrompt() (frontend/src/components/bmad/inputWidgets/ChoiceWidget.test.ts:L9)
- FileInputWidget.test.ts (frontend/src/components/bmad/inputWidgets/FileInputWidget.test.ts:L1)
- Mount (frontend/src/components/bmad/inputWidgets/FileInputWidget.test.ts:L24)
- makePrompt() (frontend/src/components/bmad/inputWidgets/FileInputWidget.test.ts:L9)
- FreeTextWidget.test.ts (frontend/src/components/bmad/inputWidgets/FreeTextWidget.test.ts:L1)
- Mount (frontend/src/components/bmad/inputWidgets/FreeTextWidget.test.ts:L25)
- makePrompt() (frontend/src/components/bmad/inputWidgets/FreeTextWidget.test.ts:L9)
- JsonInputWidget.test.ts (frontend/src/components/bmad/inputWidgets/JsonInputWidget.test.ts:L1)
- Mount (frontend/src/components/bmad/inputWidgets/JsonInputWidget.test.ts:L24)
- makePrompt() (frontend/src/components/bmad/inputWidgets/JsonInputWidget.test.ts:L9)
- MultiChoiceWidget.test.ts (frontend/src/components/bmad/inputWidgets/MultiChoiceWidget.test.ts:L1)
- Mount (frontend/src/components/bmad/inputWidgets/MultiChoiceWidget.test.ts:L24)
- makePrompt() (frontend/src/components/bmad/inputWidgets/MultiChoiceWidget.test.ts:L9)
- markdownMenuSettings.test.ts (frontend/src/lib/stores/markdownMenuSettings.test.ts:L1)
- mocks (frontend/src/lib/stores/markdownMenuSettings.test.ts:L18)
- DEFAULTS (frontend/src/lib/stores/markdownMenuSettings.test.ts:L22)
- markdownMenuSettings.ts (frontend/src/lib/stores/markdownMenuSettings.ts:L1)
- DEFAULTS (frontend/src/lib/stores/markdownMenuSettings.ts:L13)
- markdownMenuDirty (frontend/src/lib/stores/markdownMenuSettings.ts:L23)
- astResponses.test.ts (frontend/src/stores/__tests__/astResponses.test.ts:L1)
- astResponses.ts (frontend/src/stores/astResponses.ts:L1)
- makeAstResponses() (frontend/src/stores/astResponses.ts:L4)
- interactiveInput.test.ts (frontend/src/stores/interactiveInput.test.ts:L1)
- makePrompt() (frontend/src/stores/interactiveInput.test.ts:L32)
- interactiveInput.ts (frontend/src/stores/interactiveInput.ts:L1)
- updateRound() (frontend/src/stores/interactiveInput.ts:L107)
- openModal() (frontend/src/stores/interactiveInput.ts:L111)
- closeModal() (frontend/src/stores/interactiveInput.ts:L115)
- openModalForNode() (frontend/src/stores/interactiveInput.ts:L119)
- pushToast() (frontend/src/stores/interactiveInput.ts:L127)
- dismissNode() (frontend/src/stores/interactiveInput.ts:L133)
- resetInteractiveInput() (frontend/src/stores/interactiveInput.ts:L142)
- pendingPrompt (frontend/src/stores/interactiveInput.ts:L153)
- parseStructuredAst() (frontend/src/stores/interactiveInput.ts:L155)
- pendingAst (frontend/src/stores/interactiveInput.ts:L166)
- ToastKind (frontend/src/stores/interactiveInput.ts:L37)
- Toast (frontend/src/stores/interactiveInput.ts:L39)
- InteractiveInputState (frontend/src/stores/interactiveInput.ts:L45)
- initial (frontend/src/stores/interactiveInput.ts:L53)
- interactiveInput (frontend/src/stores/interactiveInput.ts:L61)
- validationKey() (frontend/src/stores/interactiveInput.ts:L63)
- upsertPrompt() (frontend/src/stores/interactiveInput.ts:L67)
- resolveInput() (frontend/src/stores/interactiveInput.ts:L76)
- setValidationError() (frontend/src/stores/interactiveInput.ts:L91)

# Depends on
- [App.js](/modules/app-js.md)
- [GetConfig](/modules/getconfig.md)
- [markdownToolbarBuilder.test.ts](/modules/markdowntoolbarbuilder-test-ts.md)
- [models.ts](/modules/models-ts.md)
- [mountSvelte.ts](/modules/mountsvelte-ts.md)
- [SetMarkdownMenuSettings](/modules/setmarkdownmenusettings.md)
- [svelte](/modules/svelte.md)
- [uiAst.ts](/modules/uiast-ts.md)

# Inferred
- [ui-ast-view-raw.spec.ts](/modules/ui-ast-view-raw-spec-ts.md)

# Features
- no feature plan names these files
