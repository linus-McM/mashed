---
type: Module
title: applyTheme
description: "Graphify community 61: docs/stories/old_stories/S03-config-persistence.md, docs/stories/old_stories/theme-03-store-refactor.md, frontend/src/lib/monacoTheme.js, frontend/src/lib/stores/theme.js, front"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: S03-config-persistence, resource: docs/stories/old_stories/S03-config-persistence.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d2003a3ed0a69d86 }
  - { id: theme-03-store-refactor, resource: docs/stories/old_stories/theme-03-store-refactor.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 093f8b49cf309f4f }
  - { id: monacoTheme, resource: frontend/src/lib/monacoTheme.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8ae058a6abdd0dfa }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
  - { id: plan, resource: sdlc/repo-health-remediation/plan.md, last_modified: "2026-09-29T12:23:00Z", digest: 630a4bebc7512078 }
---

# Files
- `docs/stories/old_stories/S03-config-persistence.md`
- `docs/stories/old_stories/theme-03-store-refactor.md`
- `frontend/src/lib/monacoTheme.js`
- `frontend/src/lib/stores/theme.js`
- `frontend/wailsjs/go/main/App.js`
- `sdlc/repo-health-remediation/plan.md`

# Symbols
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
- PR 2: Broken features and races (branch `repo-health/pr2-features`) (sdlc/repo-health-remediation/plan.md:L492)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [App.js](/modules/app-js.md)
- [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)
- [TakeScreenshot](/modules/takescreenshot.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
