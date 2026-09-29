---
type: Module
title: MarkdownEditor.test.ts
description: "Graphify community 216: docs/stories/markdown-toolbar-08-e2e-verification.md, frontend/src/components/__tests__/MarkdownEditor.test.ts, frontend/src/components/markdownEditorUtils.ts"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: markdown-toolbar-08-e2e-verification, resource: docs/stories/markdown-toolbar-08-e2e-verification.md, last_modified: "2026-04-23T11:24:04+10:00", digest: e8063a9669ea959d }
  - { id: MarkdownEditor.test, resource: frontend/src/components/__tests__/MarkdownEditor.test.ts, last_modified: "2026-04-23T11:20:33+10:00", digest: ce3fb0e409088000 }
  - { id: markdownEditorUtils, resource: frontend/src/components/markdownEditorUtils.ts, last_modified: "2026-04-23T11:20:33+10:00", digest: 20c05d3fa910184f }
---

# Files
- `docs/stories/markdown-toolbar-08-e2e-verification.md`
- `frontend/src/components/__tests__/MarkdownEditor.test.ts`
- `frontend/src/components/markdownEditorUtils.ts`

# Symbols
- Outcome (docs/stories/markdown-toolbar-08-e2e-verification.md:L300)
- Static acceptance — PASS (docs/stories/markdown-toolbar-08-e2e-verification.md:L304)
- Strategy note — fallback shipped (docs/stories/markdown-toolbar-08-e2e-verification.md:L317)
- Moments A–G — live verification MANUAL (docs/stories/markdown-toolbar-08-e2e-verification.md:L326)
- Canonical screenshots (docs/stories/markdown-toolbar-08-e2e-verification.md:L340)
- Regressions filed (docs/stories/markdown-toolbar-08-e2e-verification.md:L344)
- "Feels right" summary (docs/stories/markdown-toolbar-08-e2e-verification.md:L348)
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
- [App.js](/modules/app-js.md)
- [markdownToolbarBuilder.test.ts](/modules/markdowntoolbarbuilder-test-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)
- [markdownToolbarBuilder.test.ts](/modules/markdowntoolbarbuilder-test-ts.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
