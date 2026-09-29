---
type: Module
title: GetConfig
description: "Graphify community 183: docs/plans/markdown-toolbar-settings.md, docs/stories/markdown-toolbar-01-backend-config.md, docs/stories/markdown-toolbar-07-app-hydration.md, frontend/src/lib/stores/markdown"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: markdown-toolbar-settings, resource: docs/plans/markdown-toolbar-settings.md, last_modified: "2026-04-23T10:53:10+10:00", digest: b031e8aa3523f352 }
  - { id: markdown-toolbar-01-backend-config, resource: docs/stories/markdown-toolbar-01-backend-config.md, last_modified: "2026-04-23T10:53:25+10:00", digest: 14619a436e67dd63 }
  - { id: markdown-toolbar-07-app-hydration, resource: docs/stories/markdown-toolbar-07-app-hydration.md, last_modified: "2026-04-23T11:09:52+10:00", digest: dcf0f3341a4bef15 }
  - { id: markdownMenuSettings, resource: frontend/src/lib/stores/markdownMenuSettings.ts, last_modified: "2026-04-23T11:02:33+10:00", digest: 5fd83b95073decb2 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/plans/markdown-toolbar-settings.md`
- `docs/stories/markdown-toolbar-01-backend-config.md`
- `docs/stories/markdown-toolbar-07-app-hydration.md`
- `frontend/src/lib/stores/markdownMenuSettings.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Resolved decisions (docs/plans/markdown-toolbar-settings.md:L420)
- markdown-toolbar-01-backend-config.md (docs/stories/markdown-toolbar-01-backend-config.md:L1)
- Story 01: Backend Config — MarkdownMenuSettings and Wails bindings (docs/stories/markdown-toolbar-01-backend-config.md:L1)
- Description (docs/stories/markdown-toolbar-01-backend-config.md:L10)
- BDD Test Scenarios (docs/stories/markdown-toolbar-01-backend-config.md:L101)
- Developer Notes (docs/stories/markdown-toolbar-01-backend-config.md:L14)
- Tasks / Subtasks (docs/stories/markdown-toolbar-01-backend-config.md:L145)
- Architecture (docs/stories/markdown-toolbar-01-backend-config.md:L16)
- Definition of Done (docs/stories/markdown-toolbar-01-backend-config.md:L166)
- Technical Considerations (docs/stories/markdown-toolbar-01-backend-config.md:L32)
- Wails Binding Regeneration (docs/stories/markdown-toolbar-01-backend-config.md:L37)
- Default Values (authoritative — DO NOT change without updating the plan) (docs/stories/markdown-toolbar-01-backend-config.md:L45)
- Risks & Edge Cases (docs/stories/markdown-toolbar-01-backend-config.md:L56)
- Acceptance Criteria (docs/stories/markdown-toolbar-01-backend-config.md:L66)
- markdown-toolbar-07-app-hydration.md (docs/stories/markdown-toolbar-07-app-hydration.md:L1)
- Story 07: App Hydration — initMarkdownMenuSettings on mount (docs/stories/markdown-toolbar-07-app-hydration.md:L1)
- Description (docs/stories/markdown-toolbar-07-app-hydration.md:L10)
- Tasks / Subtasks (docs/stories/markdown-toolbar-07-app-hydration.md:L124)
- Definition of Done (docs/stories/markdown-toolbar-07-app-hydration.md:L139)
- Developer Notes (docs/stories/markdown-toolbar-07-app-hydration.md:L14)
- Architecture (docs/stories/markdown-toolbar-07-app-hydration.md:L16)
- Where to place the call (docs/stories/markdown-toolbar-07-app-hydration.md:L31)
- Why cfg.markdownMenu and not GetMarkdownMenuSettings() (docs/stories/markdown-toolbar-07-app-hydration.md:L38)
- Behavior contract (docs/stories/markdown-toolbar-07-app-hydration.md:L45)
- Risks & Edge Cases (docs/stories/markdown-toolbar-07-app-hydration.md:L50)
- Reference Files (docs/stories/markdown-toolbar-07-app-hydration.md:L55)
- Acceptance Criteria (docs/stories/markdown-toolbar-07-app-hydration.md:L59)
- BDD Test Scenarios (docs/stories/markdown-toolbar-07-app-hydration.md:L90)
- initMarkdownMenuSettings() (frontend/src/lib/stores/markdownMenuSettings.ts:L30)
- DefaultMarkdownMenuSettings() (frontend/wailsjs/go/main/App.js:L17)
- SetMarkdownMenuSettings() (frontend/wailsjs/go/main/App.js:L373)
- GetConfig() (frontend/wailsjs/go/main/App.js:L65)
- GetMarkdownMenuSettings() (frontend/wailsjs/go/main/App.js:L89)

# Depends on
- [SetEditorSettings](/modules/seteditorsettings.md)

# Inferred
- [mashedConfig](/modules/mashedconfig.md)
- [SetEditorSettings](/modules/seteditorsettings.md)

# Features
- no feature plan names these files
