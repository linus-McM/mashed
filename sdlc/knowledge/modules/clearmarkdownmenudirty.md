---
type: Module
title: clearMarkdownMenuDirty
description: "Graphify community 81: docs/plans/markdown-toolbar-settings.md, docs/stories/markdown-toolbar-02-frontend-store.md, docs/stories/markdown-toolbar-05-markdown-editor-panel.md, docs/stories/markdown-too"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: markdown-toolbar-settings, resource: docs/plans/markdown-toolbar-settings.md, last_modified: "2026-04-23T10:53:10+10:00", digest: b031e8aa3523f352 }
  - { id: markdown-toolbar-02-frontend-store, resource: docs/stories/markdown-toolbar-02-frontend-store.md, last_modified: "2026-04-23T11:02:33+10:00", digest: fa89a5180f4ab303 }
  - { id: markdown-toolbar-05-markdown-editor-panel, resource: docs/stories/markdown-toolbar-05-markdown-editor-panel.md, last_modified: "2026-04-23T11:09:52+10:00", digest: e72128261212f783 }
  - { id: markdown-toolbar-06-editor-wiring, resource: docs/stories/markdown-toolbar-06-editor-wiring.md, last_modified: "2026-04-23T11:20:33+10:00", digest: 125c03e40f91a85b }
  - { id: markdown-toolbar-08-e2e-verification, resource: docs/stories/markdown-toolbar-08-e2e-verification.md, last_modified: "2026-04-23T11:24:04+10:00", digest: e8063a9669ea959d }
  - { id: markdownMenuSettings, resource: frontend/src/lib/stores/markdownMenuSettings.ts, last_modified: "2026-04-23T11:02:33+10:00", digest: 5fd83b95073decb2 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
  - { id: ui-ast-rendering.spec, resource: tests/ac/ui-ast-rendering.spec.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: dc25a90d2670ae4d }
---

# Files
- `docs/plans/markdown-toolbar-settings.md`
- `docs/stories/markdown-toolbar-02-frontend-store.md`
- `docs/stories/markdown-toolbar-05-markdown-editor-panel.md`
- `docs/stories/markdown-toolbar-06-editor-wiring.md`
- `docs/stories/markdown-toolbar-08-e2e-verification.md`
- `frontend/src/lib/stores/markdownMenuSettings.ts`
- `frontend/wailsjs/go/main/App.js`
- `tests/ac/ui-ast-rendering.spec.ts`

# Symbols
- Svelte — new test files (docs/plans/markdown-toolbar-settings.md:L390)
- markdown-toolbar-02-frontend-store.md (docs/stories/markdown-toolbar-02-frontend-store.md:L1)
- Story 02: Frontend Store — markdownMenuSettings (docs/stories/markdown-toolbar-02-frontend-store.md:L1)
- Description (docs/stories/markdown-toolbar-02-frontend-store.md:L10)
- Developer Notes (docs/stories/markdown-toolbar-02-frontend-store.md:L14)
- Tasks / Subtasks (docs/stories/markdown-toolbar-02-frontend-store.md:L140)
- Definition of Done (docs/stories/markdown-toolbar-02-frontend-store.md:L156)
- Architecture (docs/stories/markdown-toolbar-02-frontend-store.md:L16)
- Behavior contract (docs/stories/markdown-toolbar-02-frontend-store.md:L41)
- Technical Considerations (docs/stories/markdown-toolbar-02-frontend-store.md:L50)
- Risks & Edge Cases (docs/stories/markdown-toolbar-02-frontend-store.md:L54)
- Reference Files (docs/stories/markdown-toolbar-02-frontend-store.md:L59)
- Acceptance Criteria (docs/stories/markdown-toolbar-02-frontend-store.md:L63)
- BDD Test Scenarios (docs/stories/markdown-toolbar-02-frontend-store.md:L99)
- markdown-toolbar-05-markdown-editor-panel.md (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L1)
- Story 05: Markdown Editor Panel in Settings (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L1)
- Description (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L10)
- BDD Test Scenarios (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L137)
- Developer Notes (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L14)
- Architecture (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L16)
- Tasks / Subtasks (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L180)
- Definition of Done (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L198)
- Behavior contract (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L78)
- Copy (label text, final) (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L82)
- Styling (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L85)
- Risks & Edge Cases (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L88)
- Reference Files (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L94)
- Acceptance Criteria (docs/stories/markdown-toolbar-05-markdown-editor-panel.md:L98)
- Design Brief (docs/stories/markdown-toolbar-06-editor-wiring.md:L214)
- Intent (docs/stories/markdown-toolbar-06-editor-wiring.md:L216)
- 1. Layout composition (behavioural layout, not visual) (docs/stories/markdown-toolbar-06-editor-wiring.md:L221)
- 2. Typography plan (docs/stories/markdown-toolbar-06-editor-wiring.md:L247)
- 3. Color strategy (docs/stories/markdown-toolbar-06-editor-wiring.md:L250)
- 4. Interaction model — the "feels right" moments (docs/stories/markdown-toolbar-06-editor-wiring.md:L253)
- 5. Component specs (docs/stories/markdown-toolbar-06-editor-wiring.md:L269)
- 6. Signature elements — what makes this "unmistakably Mashed" behaviourally (docs/stories/markdown-toolbar-06-editor-wiring.md:L310)
- 7. Anti-patterns to avoid (docs/stories/markdown-toolbar-06-editor-wiring.md:L317)
- Behavior contract (docs/stories/markdown-toolbar-06-editor-wiring.md:L81)
- Moments A–G — live verification MANUAL (docs/stories/markdown-toolbar-08-e2e-verification.md:L326)
- updateMarkdownMenuItem() (frontend/src/lib/stores/markdownMenuSettings.ts:L62)
- clearMarkdownMenuDirty() (frontend/src/lib/stores/markdownMenuSettings.ts:L81)
- SetMarkdownMenuSettings() (frontend/wailsjs/go/main/App.js:L373)
- get() (tests/ac/ui-ast-rendering.spec.ts:L184)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [MarkdownEditor.test.ts](/modules/markdowneditor-test-ts.md)
- [markdownToolbarBuilder.test.ts](/modules/markdowntoolbarbuilder-test-ts.md)
- [SetEditorSettings](/modules/seteditorsettings.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
