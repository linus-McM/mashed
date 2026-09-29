---
type: Module
title: GetConfig
description: "Graphify community 74: docs/stories/markdown-toolbar-07-app-hydration.md, docs/stories/old_stories/S03-config-persistence.md, docs/stories/old_stories/S04-titlebar-theme-popover.md, docs/stories/old_s"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: markdown-toolbar-07-app-hydration, resource: docs/stories/markdown-toolbar-07-app-hydration.md, last_modified: "2026-04-23T11:09:52+10:00", digest: dcf0f3341a4bef15 }
  - { id: S03-config-persistence, resource: docs/stories/old_stories/S03-config-persistence.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d2003a3ed0a69d86 }
  - { id: S04-titlebar-theme-popover, resource: docs/stories/old_stories/S04-titlebar-theme-popover.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 997ae10d98fecd13 }
  - { id: S05-settings-view, resource: docs/stories/old_stories/S05-settings-view.md, last_modified: "2026-04-08T10:23:03+10:00", digest: dac9b7a63edb9e0e }
  - { id: theme-01-backend-scanner, resource: docs/stories/old_stories/theme-01-backend-scanner.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 1745d94d5246a6de }
  - { id: markdownMenuSettings, resource: frontend/src/lib/stores/markdownMenuSettings.ts, last_modified: "2026-04-23T11:02:33+10:00", digest: 5fd83b95073decb2 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/stories/markdown-toolbar-07-app-hydration.md`
- `docs/stories/old_stories/S03-config-persistence.md`
- `docs/stories/old_stories/S04-titlebar-theme-popover.md`
- `docs/stories/old_stories/S05-settings-view.md`
- `docs/stories/old_stories/theme-01-backend-scanner.md`
- `frontend/src/lib/stores/markdownMenuSettings.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
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
- S03-config-persistence.md (docs/stories/old_stories/S03-config-persistence.md:L1)
- Story 3: Config Persistence + Edit Menu (docs/stories/old_stories/S03-config-persistence.md:L1)
- Technical Considerations (docs/stories/old_stories/S03-config-persistence.md:L111)
- Risks & Edge Cases (docs/stories/old_stories/S03-config-persistence.md:L119)
- Reference Files (docs/stories/old_stories/S03-config-persistence.md:L126)
- Developer Notes (docs/stories/old_stories/S03-config-persistence.md:L13)
- Acceptance Criteria (docs/stories/old_stories/S03-config-persistence.md:L133)
- Architecture (docs/stories/old_stories/S03-config-persistence.md:L15)
- BDD Test Scenarios (docs/stories/old_stories/S03-config-persistence.md:L176)
- Scenario 1: Config Struct Serialization (docs/stories/old_stories/S03-config-persistence.md:L178)
- Scenario 2: Wails-Bound Methods (docs/stories/old_stories/S03-config-persistence.md:L201)
- Go Changes (app.go) (docs/stories/old_stories/S03-config-persistence.md:L21)
- Scenario 3: Edit Menu (docs/stories/old_stories/S03-config-persistence.md:L224)
- Scenario 4: Theme Restoration on Startup (docs/stories/old_stories/S03-config-persistence.md:L242)
- Tasks / Subtasks (docs/stories/old_stories/S03-config-persistence.md:L261)
- Definition of Done (docs/stories/old_stories/S03-config-persistence.md:L286)
- Go Changes (main.go) (docs/stories/old_stories/S03-config-persistence.md:L64)
- Frontend Changes (App.svelte) (docs/stories/old_stories/S03-config-persistence.md:L87)
- Description (docs/stories/old_stories/S03-config-persistence.md:L9)
- Tasks / Subtasks (docs/stories/old_stories/S04-titlebar-theme-popover.md:L262)
- Acceptance Criteria (docs/stories/old_stories/S05-settings-view.md:L166)
- theme-01-backend-scanner.md (docs/stories/old_stories/theme-01-backend-scanner.md:L1)
- Story 1: Go Backend -- Extension Scanner & Theme Reader (docs/stories/old_stories/theme-01-backend-scanner.md:L1)
- Developer Notes (docs/stories/old_stories/theme-01-backend-scanner.md:L13)
- Risks & Edge Cases (docs/stories/old_stories/theme-01-backend-scanner.md:L141)
- Architecture (docs/stories/old_stories/theme-01-backend-scanner.md:L15)
- Reference Files (docs/stories/old_stories/theme-01-backend-scanner.md:L150)
- Acceptance Criteria (docs/stories/old_stories/theme-01-backend-scanner.md:L156)
- BDD Test Scenarios (docs/stories/old_stories/theme-01-backend-scanner.md:L201)
- Scenario 1: Happy path -- scan and read themes (docs/stories/old_stories/theme-01-backend-scanner.md:L203)
- Scenario 2: Security -- path traversal (docs/stories/old_stories/theme-01-backend-scanner.md:L242)
- Scenario 3: Edge cases (docs/stories/old_stories/theme-01-backend-scanner.md:L265)
- Tasks / Subtasks (docs/stories/old_stories/theme-01-backend-scanner.md:L292)
- Definition of Done (docs/stories/old_stories/theme-01-backend-scanner.md:L330)
- Technical Considerations (docs/stories/old_stories/theme-01-backend-scanner.md:L50)
- Description (docs/stories/old_stories/theme-01-backend-scanner.md:L9)
- initMarkdownMenuSettings() (frontend/src/lib/stores/markdownMenuSettings.ts:L30)
- SetDevDir() (frontend/wailsjs/go/main/App.js:L357)
- SetTheme() (frontend/wailsjs/go/main/App.js:L397)
- SetVSCodiumExtPath() (frontend/wailsjs/go/main/App.js:L413)
- GetConfig() (frontend/wailsjs/go/main/App.js:L65)
- GetDevDir() (frontend/wailsjs/go/main/App.js:L73)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [SetEditorSettings](/modules/seteditorsettings.md)
- [SetMarkdownMenuSettings](/modules/setmarkdownmenusettings.md)
- [Story svelte-check-01: JS stores → TypeScript (Phase 1)](/modules/story-svelte-check-01-js-stores-typescript-phase-1.md)
- [themeInit.js](/modules/themeinit-js.md)
- [WriteFile](/modules/writefile.md)

# Features
- no feature plan names these files
