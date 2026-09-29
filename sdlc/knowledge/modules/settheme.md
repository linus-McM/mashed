---
type: Module
title: SetTheme
description: "Graphify community 74: docs/stories/old_stories/S03-config-persistence.md, docs/stories/old_stories/S05-settings-view.md, docs/stories/old_stories/theme-01-backend-scanner.md, frontend/wailsjs/go/main"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: S03-config-persistence, resource: docs/stories/old_stories/S03-config-persistence.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d2003a3ed0a69d86 }
  - { id: S05-settings-view, resource: docs/stories/old_stories/S05-settings-view.md, last_modified: "2026-04-08T10:23:03+10:00", digest: dac9b7a63edb9e0e }
  - { id: theme-01-backend-scanner, resource: docs/stories/old_stories/theme-01-backend-scanner.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 1745d94d5246a6de }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
  - { id: intent, resource: sdlc/repo-health-remediation/intent.md, last_modified: "2026-09-29T21:16:08+10:00", digest: 476d74d6341cb18a }
---

# Files
- `docs/stories/old_stories/S03-config-persistence.md`
- `docs/stories/old_stories/S05-settings-view.md`
- `docs/stories/old_stories/theme-01-backend-scanner.md`
- `frontend/wailsjs/go/main/App.js`
- `sdlc/repo-health-remediation/intent.md`

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
- SetDevDir() (frontend/wailsjs/go/main/App.js:L357)
- SetImportedTheme() (frontend/wailsjs/go/main/App.js:L369)
- SetTheme() (frontend/wailsjs/go/main/App.js:L397)
- SetVSCodiumExtPath() (frontend/wailsjs/go/main/App.js:L413)
- intent.md (sdlc/repo-health-remediation/intent.md:L1)
- Intent: Repo health remediation (sdlc/repo-health-remediation/intent.md:L1)
- Proposed outcome (sdlc/repo-health-remediation/intent.md:L19)
- Affected users and systems (sdlc/repo-health-remediation/intent.md:L32)
- Problem (sdlc/repo-health-remediation/intent.md:L4)
- Constraints (sdlc/repo-health-remediation/intent.md:L47)
- Open questions (sdlc/repo-health-remediation/intent.md:L58)

# Depends on
- [applyTheme](/modules/applytheme.md)

# Inferred
- [applyTheme](/modules/applytheme.md)
- [GetConfig](/modules/getconfig.md)
- [ReadFileBase64](/modules/readfilebase64.md)
- [themeInit.js](/modules/themeinit-js.md)
- [WriteFile](/modules/writefile.md)

# Features
- no feature plan names these files
