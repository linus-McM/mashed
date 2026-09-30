---
type: Module
title: SetTheme
description: "Graphify community 333: docs/stories/old_stories/S03-config-persistence.md, docs/stories/old_stories/S05-settings-view.md, docs/stories/old_stories/theme-01-backend-scanner.md, frontend/wailsjs/go/mai"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: S03-config-persistence, resource: docs/stories/old_stories/S03-config-persistence.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d2003a3ed0a69d86 }
  - { id: S05-settings-view, resource: docs/stories/old_stories/S05-settings-view.md, last_modified: "2026-04-08T10:23:03+10:00", digest: dac9b7a63edb9e0e }
  - { id: theme-01-backend-scanner, resource: docs/stories/old_stories/theme-01-backend-scanner.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 1745d94d5246a6de }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
---

# Files
- `docs/stories/old_stories/S03-config-persistence.md`
- `docs/stories/old_stories/S05-settings-view.md`
- `docs/stories/old_stories/theme-01-backend-scanner.md`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Technical Considerations (docs/stories/old_stories/S03-config-persistence.md:L111)
- Risks & Edge Cases (docs/stories/old_stories/S03-config-persistence.md:L119)
- Reference Files (docs/stories/old_stories/S03-config-persistence.md:L126)
- Developer Notes (docs/stories/old_stories/S03-config-persistence.md:L13)
- Architecture (docs/stories/old_stories/S03-config-persistence.md:L15)
- Go Changes (app.go) (docs/stories/old_stories/S03-config-persistence.md:L21)
- Go Changes (main.go) (docs/stories/old_stories/S03-config-persistence.md:L64)
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
- SetImportedTheme() (frontend/wailsjs/go/main/App.js:L373)
- SetTheme() (frontend/wailsjs/go/main/App.js:L401)
- SetVSCodiumExtPath() (frontend/wailsjs/go/main/App.js:L417)

# Depends on
- [SpawnRefactorPlan](/modules/spawnrefactorplan.md)

# Inferred
- [applyTheme](/modules/applytheme.md)
- [themeInit.js](/modules/themeinit-js.md)
- [WriteFile](/modules/writefile.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
