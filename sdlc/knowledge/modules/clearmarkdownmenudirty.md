---
type: Module
title: clearMarkdownMenuDirty
description: "Graphify community 81: docs/plans/markdown-toolbar-settings.md, docs/stories/markdown-toolbar-02-frontend-store.md, docs/stories/markdown-toolbar-05-markdown-editor-panel.md, docs/stories/markdown-too"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: markdown-toolbar-settings, resource: docs/plans/markdown-toolbar-settings.md, last_modified: "2026-04-23T10:53:10+10:00", digest: b031e8aa3523f352 }
  - { id: markdown-toolbar-02-frontend-store, resource: docs/stories/markdown-toolbar-02-frontend-store.md, last_modified: "2026-04-23T11:02:33+10:00", digest: fa89a5180f4ab303 }
  - { id: markdown-toolbar-05-markdown-editor-panel, resource: docs/stories/markdown-toolbar-05-markdown-editor-panel.md, last_modified: "2026-04-23T11:09:52+10:00", digest: e72128261212f783 }
  - { id: markdown-toolbar-06-editor-wiring, resource: docs/stories/markdown-toolbar-06-editor-wiring.md, last_modified: "2026-04-23T11:20:33+10:00", digest: 125c03e40f91a85b }
  - { id: markdownMenuSettings, resource: frontend/src/lib/stores/markdownMenuSettings.ts, last_modified: "2026-04-23T11:02:33+10:00", digest: 5fd83b95073decb2 }
  - { id: ui-ast-rendering.spec, resource: tests/ac/ui-ast-rendering.spec.ts, last_modified: "2026-04-22T08:19:28+10:00", digest: dc25a90d2670ae4d }
---

# Files
- `docs/plans/markdown-toolbar-settings.md`
- `docs/stories/markdown-toolbar-02-frontend-store.md`
- `docs/stories/markdown-toolbar-05-markdown-editor-panel.md`
- `docs/stories/markdown-toolbar-06-editor-wiring.md`
- `frontend/src/lib/stores/markdownMenuSettings.ts`
- `tests/ac/ui-ast-rendering.spec.ts`

# Symbols
- Tests (docs/plans/markdown-toolbar-settings.md:L382)
- Go — `app_test.go` (docs/plans/markdown-toolbar-settings.md:L384)
- Svelte — new test files (docs/plans/markdown-toolbar-settings.md:L390)
- markdown-toolbar-02-frontend-store.md (docs/stories/markdown-toolbar-02-frontend-store.md:L1)
- Story 02: Frontend Store — markdownMenuSettings (docs/stories/markdown-toolbar-02-frontend-store.md:L1)
- Description (docs/stories/markdown-toolbar-02-frontend-store.md:L10)
- Tasks / Subtasks (docs/stories/markdown-toolbar-02-frontend-store.md:L140)
- Definition of Done (docs/stories/markdown-toolbar-02-frontend-store.md:L156)
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
- markdown-toolbar-06-editor-wiring.md (docs/stories/markdown-toolbar-06-editor-wiring.md:L1)
- Story 06: MarkdownEditor Wiring — featureConfigs + deferred re-init (docs/stories/markdown-toolbar-06-editor-wiring.md:L1)
- Description (docs/stories/markdown-toolbar-06-editor-wiring.md:L10)
- Acceptance Criteria (docs/stories/markdown-toolbar-06-editor-wiring.md:L100)
- BDD Test Scenarios (docs/stories/markdown-toolbar-06-editor-wiring.md:L139)
- Developer Notes (docs/stories/markdown-toolbar-06-editor-wiring.md:L14)
- Architecture (docs/stories/markdown-toolbar-06-editor-wiring.md:L16)
- Tasks / Subtasks (docs/stories/markdown-toolbar-06-editor-wiring.md:L183)
- Definition of Done (docs/stories/markdown-toolbar-06-editor-wiring.md:L204)
- Design Brief (docs/stories/markdown-toolbar-06-editor-wiring.md:L214)
- Intent (docs/stories/markdown-toolbar-06-editor-wiring.md:L216)
- 1. Layout composition (behavioural layout, not visual) (docs/stories/markdown-toolbar-06-editor-wiring.md:L221)
- 2. Typography plan (docs/stories/markdown-toolbar-06-editor-wiring.md:L247)
- 3. Color strategy (docs/stories/markdown-toolbar-06-editor-wiring.md:L250)
- 4. Interaction model — the "feels right" moments (docs/stories/markdown-toolbar-06-editor-wiring.md:L253)
- 5. Component specs (docs/stories/markdown-toolbar-06-editor-wiring.md:L269)
- 6. Signature elements — what makes this "unmistakably Mashed" behaviourally (docs/stories/markdown-toolbar-06-editor-wiring.md:L310)
- 7. Anti-patterns to avoid (docs/stories/markdown-toolbar-06-editor-wiring.md:L317)
- The saver.flush() placement (resolved plan decision) (docs/stories/markdown-toolbar-06-editor-wiring.md:L57)
- Reactivity semantics (docs/stories/markdown-toolbar-06-editor-wiring.md:L70)
- Why JSON stringify (docs/stories/markdown-toolbar-06-editor-wiring.md:L78)
- Behavior contract (docs/stories/markdown-toolbar-06-editor-wiring.md:L81)
- Risks & Edge Cases (docs/stories/markdown-toolbar-06-editor-wiring.md:L87)
- updateMarkdownMenuItem() (frontend/src/lib/stores/markdownMenuSettings.ts:L62)
- clearMarkdownMenuDirty() (frontend/src/lib/stores/markdownMenuSettings.ts:L81)
- get() (tests/ac/ui-ast-rendering.spec.ts:L184)

# Depends on
- [MarkdownEditor.test.ts](/modules/markdowneditor-test-ts.md)
- [SetEditorSettings](/modules/seteditorsettings.md)

# Inferred
- [markdownMenuSettings.ts](/modules/markdownmenusettings-ts.md)
- [SetEditorSettings](/modules/seteditorsettings.md)

# Features
- no feature plan names these files
