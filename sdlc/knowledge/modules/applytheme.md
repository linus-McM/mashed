---
type: Module
title: applyTheme
description: "Graphify community 61: docs/feasibility-multi-editor.md, docs/stories/old_stories/S03-config-persistence.md, docs/stories/old_stories/theme-03-store-refactor.md, frontend/src/lib/monacoTheme.js, front"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: feasibility-multi-editor, resource: docs/feasibility-multi-editor.md, last_modified: "2026-04-09T21:06:37+10:00", digest: d1cc54cadf1483b3 }
  - { id: S03-config-persistence, resource: docs/stories/old_stories/S03-config-persistence.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d2003a3ed0a69d86 }
  - { id: theme-03-store-refactor, resource: docs/stories/old_stories/theme-03-store-refactor.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 093f8b49cf309f4f }
  - { id: monacoTheme, resource: frontend/src/lib/monacoTheme.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8ae058a6abdd0dfa }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/feasibility-multi-editor.md`
- `docs/stories/old_stories/S03-config-persistence.md`
- `docs/stories/old_stories/theme-03-store-refactor.md`
- `frontend/src/lib/monacoTheme.js`
- `frontend/src/lib/stores/theme.js`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Decision: **Milkdown Crepe** (`@milkdown/crepe`) (docs/feasibility-multi-editor.md:L101)
- Why Crepe (docs/feasibility-multi-editor.md:L105)
- Integration Pattern (docs/feasibility-multi-editor.md:L123)
- Theme Integration (docs/feasibility-multi-editor.md:L141)
- No Raw Source View Needed (docs/feasibility-multi-editor.md:L145)
- Getting Content for Auto-Save (docs/feasibility-multi-editor.md:L149)
- 3. Markdown Editor: Milkdown Crepe (DECIDED) (docs/feasibility-multi-editor.md:L99)
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
- Risks & Edge Cases (docs/stories/old_stories/theme-03-store-refactor.md:L96)
- defineImportedTheme() (frontend/src/lib/monacoTheme.js:L49)
- applyTheme() (frontend/src/lib/stores/theme.js:L38)
- registerImportedTheme() (frontend/src/lib/stores/theme.js:L52)
- GetDevDir() (frontend/wailsjs/go/main/App.js:L73)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [Story 02: Frontend Store — markdownMenuSettings](/modules/story-02-frontend-store-markdownmenusettings.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)
- [WriteFile](/modules/writefile.md)

# Features
- no feature plan names these files
