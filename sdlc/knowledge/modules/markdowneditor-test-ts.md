---
type: Module
title: MarkdownEditor.test.ts
description: "Graphify community 216: docs/stories/markdown-toolbar-06-editor-wiring.md, docs/stories/markdown-toolbar-08-e2e-verification.md, frontend/src/components/__tests__/MarkdownEditor.test.ts, frontend/src/"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: markdown-toolbar-06-editor-wiring, resource: docs/stories/markdown-toolbar-06-editor-wiring.md, last_modified: "2026-04-23T11:20:33+10:00", digest: 125c03e40f91a85b }
  - { id: markdown-toolbar-08-e2e-verification, resource: docs/stories/markdown-toolbar-08-e2e-verification.md, last_modified: "2026-04-23T11:24:04+10:00", digest: e8063a9669ea959d }
  - { id: MarkdownEditor.test, resource: frontend/src/components/__tests__/MarkdownEditor.test.ts, last_modified: "2026-04-23T11:20:33+10:00", digest: ce3fb0e409088000 }
  - { id: markdownEditorUtils, resource: frontend/src/components/markdownEditorUtils.ts, last_modified: "2026-04-23T11:20:33+10:00", digest: 20c05d3fa910184f }
---

# Files
- `docs/stories/markdown-toolbar-06-editor-wiring.md`
- `docs/stories/markdown-toolbar-08-e2e-verification.md`
- `frontend/src/components/__tests__/MarkdownEditor.test.ts`
- `frontend/src/components/markdownEditorUtils.ts`

# Symbols
- Reference Files (docs/stories/markdown-toolbar-06-editor-wiring.md:L95)
- Moments A–G — live verification MANUAL (docs/stories/markdown-toolbar-08-e2e-verification.md:L326)
- MarkdownEditor.test.ts (frontend/src/components/__tests__/MarkdownEditor.test.ts:L1)
- ALL_ON (frontend/src/components/__tests__/MarkdownEditor.test.ts:L13)
- ALL_OFF (frontend/src/components/__tests__/MarkdownEditor.test.ts:L22)
- markdownEditorUtils.ts (frontend/src/components/markdownEditorUtils.ts:L1)
- TOOLBAR_KEYS (frontend/src/components/markdownEditorUtils.ts:L10)
- ToolbarKey (frontend/src/components/markdownEditorUtils.ts:L19)
- SaveStatus (frontend/src/components/markdownEditorUtils.ts:L3)
- applyToolbarAttributes() (frontend/src/components/markdownEditorUtils.ts:L33)
- computeToolbarApplyTarget() (frontend/src/components/markdownEditorUtils.ts:L56)
- DebouncedSave (frontend/src/components/markdownEditorUtils.ts:L72)
- .schedule() (frontend/src/components/markdownEditorUtils.ts:L73)
- .flush() (frontend/src/components/markdownEditorUtils.ts:L74)
- .cancel() (frontend/src/components/markdownEditorUtils.ts:L75)
- createDebouncedSave() (frontend/src/components/markdownEditorUtils.ts:L79)
- cancel() (frontend/src/components/markdownEditorUtils.ts:L86)

# Depends on
- [markdownToolbarBuilder.test.ts](/modules/markdowntoolbarbuilder-test-ts.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [vitest](/modules/vitest.md)

# Inferred
- [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
