---
type: Module
title: applyTheme
description: "Graphify community 61: docs/stories/markdown-toolbar-02-frontend-store.md, docs/stories/markdown-toolbar-06-editor-wiring.md, docs/stories/old_stories/S03-config-persistence.md, docs/stories/old_stori"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: markdown-toolbar-02-frontend-store, resource: docs/stories/markdown-toolbar-02-frontend-store.md, last_modified: "2026-04-23T11:02:33+10:00", digest: fa89a5180f4ab303 }
  - { id: markdown-toolbar-06-editor-wiring, resource: docs/stories/markdown-toolbar-06-editor-wiring.md, last_modified: "2026-04-23T11:20:33+10:00", digest: 125c03e40f91a85b }
  - { id: S03-config-persistence, resource: docs/stories/old_stories/S03-config-persistence.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d2003a3ed0a69d86 }
  - { id: theme-03-store-refactor, resource: docs/stories/old_stories/theme-03-store-refactor.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 093f8b49cf309f4f }
  - { id: monacoTheme, resource: frontend/src/lib/monacoTheme.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8ae058a6abdd0dfa }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
  - { id: ui-ast-rendering.spec, resource: tests/ac/ui-ast-rendering.spec.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: dc25a90d2670ae4d }
---

# Files
- `docs/stories/markdown-toolbar-02-frontend-store.md`
- `docs/stories/markdown-toolbar-06-editor-wiring.md`
- `docs/stories/old_stories/S03-config-persistence.md`
- `docs/stories/old_stories/theme-03-store-refactor.md`
- `frontend/src/lib/monacoTheme.js`
- `frontend/src/lib/stores/theme.js`
- `frontend/wailsjs/go/main/App.js`
- `tests/ac/ui-ast-rendering.spec.ts`

# Symbols
- Acceptance Criteria (docs/stories/markdown-toolbar-02-frontend-store.md:L63)
- 5. Component specs (docs/stories/markdown-toolbar-06-editor-wiring.md:L269)
- Tasks / Subtasks (docs/stories/old_stories/S03-config-persistence.md:L261)
- Frontend Changes (App.svelte) (docs/stories/old_stories/S03-config-persistence.md:L87)
- theme-03-store-refactor.md (docs/stories/old_stories/theme-03-store-refactor.md:L1)
- Story 3: Theme Store Refactor & Consumer Updates (docs/stories/old_stories/theme-03-store-refactor.md:L1)
- Reference Files (docs/stories/old_stories/theme-03-store-refactor.md:L103)
- Acceptance Criteria (docs/stories/old_stories/theme-03-store-refactor.md:L114)
- Developer Notes (docs/stories/old_stories/theme-03-store-refactor.md:L13)
- Architecture (docs/stories/old_stories/theme-03-store-refactor.md:L15)
- BDD Test Scenarios (docs/stories/old_stories/theme-03-store-refactor.md:L161)
- Scenario 1: Store operations (docs/stories/old_stories/theme-03-store-refactor.md:L163)
- Scenario 2: TitleBar consumer update (docs/stories/old_stories/theme-03-store-refactor.md:L189)
- Scenario 3: Settings consumer update (docs/stories/old_stories/theme-03-store-refactor.md:L207)
- Scenario 4: Backward compatibility (docs/stories/old_stories/theme-03-store-refactor.md:L225)
- Tasks / Subtasks (docs/stories/old_stories/theme-03-store-refactor.md:L242)
- Definition of Done (docs/stories/old_stories/theme-03-store-refactor.md:L277)
- Technical Considerations (docs/stories/old_stories/theme-03-store-refactor.md:L72)
- Description (docs/stories/old_stories/theme-03-store-refactor.md:L9)
- defineImportedTheme() (frontend/src/lib/monacoTheme.js:L49)
- applyTheme() (frontend/src/lib/stores/theme.js:L38)
- GetDevDir() (frontend/wailsjs/go/main/App.js:L73)
- ui-ast-rendering.spec.ts (tests/ac/ui-ast-rendering.spec.ts:L1)
- get() (tests/ac/ui-ast-rendering.spec.ts:L184)
- set() (tests/ac/ui-ast-rendering.spec.ts:L187)
- StructuredPromptOpts (tests/ac/ui-ast-rendering.spec.ts:L21)
- emitAwaitingInput() (tests/ac/ui-ast-rendering.spec.ts:L29)
- openModal() (tests/ac/ui-ast-rendering.spec.ts:L53)
- ensureHelpers() (tests/ac/ui-ast-rendering.spec.ts:L58)

# Depends on
- [@playwright/test](/modules/playwright-test.md)
- [theme.js](/modules/theme-js.md)

# Inferred
- [GetConfig](/modules/getconfig.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)

# Features
- no feature plan names these files
