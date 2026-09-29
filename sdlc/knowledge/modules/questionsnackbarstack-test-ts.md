---
type: Module
title: QuestionSnackbarStack.test.ts
description: "Graphify community 137: docs/stories/bmad-interactive-06-frontend-modal.md, docs/stories/old_stories/question-03-snackbar-stack.md, frontend/src/components/bmad/QuestionResponseModal.svelte, frontend/"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: bmad-interactive-06-frontend-modal, resource: docs/stories/bmad-interactive-06-frontend-modal.md, last_modified: "2026-04-20T14:57:51+10:00", digest: 1b80ffd59d82265b }
  - { id: question-03-snackbar-stack, resource: docs/stories/old_stories/question-03-snackbar-stack.md, last_modified: "2026-04-12T10:43:48+10:00", digest: c9e217081e4ca99f }
  - { id: QuestionResponseModal, resource: frontend/src/components/bmad/QuestionResponseModal.svelte, last_modified: "2026-04-22T17:53:53+10:00", digest: 5672933bac7b8d06 }
  - { id: QuestionResponseModal.test, resource: frontend/src/components/bmad/QuestionResponseModal.test.ts, last_modified: "2026-04-10T14:16:40+10:00", digest: a798f0e107780ec5 }
  - { id: QuestionSnackbarStack, resource: frontend/src/components/bmad/QuestionSnackbarStack.svelte, last_modified: "2026-04-20T14:57:51+10:00", digest: c485ed709e057516 }
  - { id: QuestionSnackbarStack.test, resource: frontend/src/components/bmad/QuestionSnackbarStack.test.ts, last_modified: "2026-04-11T19:45:53+10:00", digest: 2521496fc654bb79 }
  - { id: questionSnackbarUtils, resource: frontend/src/components/bmad/questionSnackbarUtils.ts, last_modified: "2026-04-11T19:45:53+10:00", digest: 7c6697429c616053 }
---

# Files
- `docs/stories/bmad-interactive-06-frontend-modal.md`
- `docs/stories/old_stories/question-03-snackbar-stack.md`
- `frontend/src/components/bmad/QuestionResponseModal.svelte`
- `frontend/src/components/bmad/QuestionResponseModal.test.ts`
- `frontend/src/components/bmad/QuestionSnackbarStack.svelte`
- `frontend/src/components/bmad/QuestionSnackbarStack.test.ts`
- `frontend/src/components/bmad/questionSnackbarUtils.ts`

# Symbols
- 6. Snackbar redesign (`NodeInputSnackbarStack.svelte`) (docs/stories/bmad-interactive-06-frontend-modal.md:L431)
- Tasks / Subtasks (docs/stories/old_stories/question-03-snackbar-stack.md:L219)
- QuestionResponseModal.svelte (frontend/src/components/bmad/QuestionResponseModal.svelte:L1)
- repoLabel() (frontend/src/components/bmad/QuestionResponseModal.svelte:L39)
- submit() (frontend/src/components/bmad/QuestionResponseModal.svelte:L45)
- handleSend() (frontend/src/components/bmad/QuestionResponseModal.svelte:L60)
- handleOption() (frontend/src/components/bmad/QuestionResponseModal.svelte:L66)
- handleOverlayClick() (frontend/src/components/bmad/QuestionResponseModal.svelte:L75)
- handleKeydown() (frontend/src/components/bmad/QuestionResponseModal.svelte:L80)
- QuestionResponseModal.test.ts (frontend/src/components/bmad/QuestionResponseModal.test.ts:L1)
- mount() (frontend/src/components/bmad/QuestionResponseModal.test.ts:L104)
- $all() (frontend/src/components/bmad/QuestionResponseModal.test.ts:L122)
- mockQuestion() (frontend/src/components/bmad/QuestionResponseModal.test.ts:L50)
- memoryStore (frontend/src/components/bmad/QuestionResponseModal.test.ts:L67)
- fakeStorage (frontend/src/components/bmad/QuestionResponseModal.test.ts:L68)
- MountResult (frontend/src/components/bmad/QuestionResponseModal.test.ts:L98)
- QuestionSnackbarStack.svelte (frontend/src/components/bmad/QuestionSnackbarStack.svelte:L1)
- questions (frontend/src/components/bmad/QuestionSnackbarStack.svelte:L10)
- QuestionSnackbarStack.test.ts (frontend/src/components/bmad/QuestionSnackbarStack.test.ts:L1)
- memoryStore (frontend/src/components/bmad/QuestionSnackbarStack.test.ts:L23)
- fakeStorage (frontend/src/components/bmad/QuestionSnackbarStack.test.ts:L24)
- queueOf() (frontend/src/components/bmad/QuestionSnackbarStack.test.ts:L384)
- makeEvent() (frontend/src/components/bmad/QuestionSnackbarStack.test.ts:L53)
- makeIdle() (frontend/src/components/bmad/QuestionSnackbarStack.test.ts:L69)
- asQuestion() (frontend/src/components/bmad/QuestionSnackbarStack.test.ts:L83)
- questionSnackbarUtils.ts (frontend/src/components/bmad/questionSnackbarUtils.ts:L1)
- MAX_VISIBLE (frontend/src/components/bmad/questionSnackbarUtils.ts:L10)
- IdleEventLike (frontend/src/components/bmad/questionSnackbarUtils.ts:L103)
- SnackbarEntry (frontend/src/components/bmad/questionSnackbarUtils.ts:L114)
- isQuestionEntry() (frontend/src/components/bmad/questionSnackbarUtils.ts:L121)
- REPO_COLORS_KEY (frontend/src/components/bmad/questionSnackbarUtils.ts:L13)
- upsertQuestion() (frontend/src/components/bmad/questionSnackbarUtils.ts:L134)
- upsertIdle() (frontend/src/components/bmad/questionSnackbarUtils.ts:L154)
- dismissQuestion() (frontend/src/components/bmad/questionSnackbarUtils.ts:L172)
- dismissIdle() (frontend/src/components/bmad/questionSnackbarUtils.ts:L185)
- partitionForDisplay() (frontend/src/components/bmad/questionSnackbarUtils.ts:L196)
- truncate() (frontend/src/components/bmad/questionSnackbarUtils.ts:L22)
- getBorderColor() (frontend/src/components/bmad/questionSnackbarUtils.ts:L40)
- timeAgo() (frontend/src/components/bmad/questionSnackbarUtils.ts:L64)
- DEFAULT_BORDER_COLOR (frontend/src/components/bmad/questionSnackbarUtils.ts:L7)
- QuestionEventLike (frontend/src/components/bmad/questionSnackbarUtils.ts:L81)

# Depends on
- [App.js](/modules/app-js.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [svelte](/modules/svelte.md)
- [vitest](/modules/vitest.md)

# Inferred
- [svelte](/modules/svelte.md)

# Features
- no feature plan names these files
