---
type: Module
title: themeInit.js
description: "Graphify community 103: docs/stories/old_stories/S03-config-persistence.md, docs/stories/old_stories/S04-titlebar-theme-popover.md, docs/stories/old_stories/S05-settings-view.md, docs/stories/old_stor"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: S03-config-persistence, resource: docs/stories/old_stories/S03-config-persistence.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d2003a3ed0a69d86 }
  - { id: S04-titlebar-theme-popover, resource: docs/stories/old_stories/S04-titlebar-theme-popover.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 997ae10d98fecd13 }
  - { id: S05-settings-view, resource: docs/stories/old_stories/S05-settings-view.md, last_modified: "2026-04-08T10:23:03+10:00", digest: dac9b7a63edb9e0e }
  - { id: theme-01-backend-scanner, resource: docs/stories/old_stories/theme-01-backend-scanner.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 1745d94d5246a6de }
  - { id: theme-04-settings-activation, resource: docs/stories/old_stories/theme-04-settings-activation.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 5604998bc5d1ad34 }
  - { id: theme-05-monaco-registration, resource: docs/stories/old_stories/theme-05-monaco-registration.md, last_modified: "2026-04-08T10:23:03+10:00", digest: dcb77471f481713a }
  - { id: vsix-02-frontend-vsix-path-handling, resource: docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 26f44fbbb6ff8bbb }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: themeInit, resource: frontend/src/lib/themeInit.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 30161049c3a59f8f }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
---

# Files
- `docs/stories/old_stories/S03-config-persistence.md`
- `docs/stories/old_stories/S04-titlebar-theme-popover.md`
- `docs/stories/old_stories/S05-settings-view.md`
- `docs/stories/old_stories/theme-01-backend-scanner.md`
- `docs/stories/old_stories/theme-04-settings-activation.md`
- `docs/stories/old_stories/theme-05-monaco-registration.md`
- `docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md`
- `frontend/src/lib/stores/theme.js`
- `frontend/src/lib/themeInit.js`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Technical Considerations (docs/stories/old_stories/S03-config-persistence.md:L111)
- Risks & Edge Cases (docs/stories/old_stories/S03-config-persistence.md:L119)
- Reference Files (docs/stories/old_stories/S03-config-persistence.md:L126)
- Developer Notes (docs/stories/old_stories/S03-config-persistence.md:L13)
- Architecture (docs/stories/old_stories/S03-config-persistence.md:L15)
- Go Changes (app.go) (docs/stories/old_stories/S03-config-persistence.md:L21)
- Tasks / Subtasks (docs/stories/old_stories/S03-config-persistence.md:L261)
- Go Changes (main.go) (docs/stories/old_stories/S03-config-persistence.md:L64)
- Frontend Changes (App.svelte) (docs/stories/old_stories/S03-config-persistence.md:L87)
- Tasks / Subtasks (docs/stories/old_stories/S04-titlebar-theme-popover.md:L262)
- Acceptance Criteria (docs/stories/old_stories/S05-settings-view.md:L166)
- Acceptance Criteria (docs/stories/old_stories/theme-01-backend-scanner.md:L156)
- Tasks / Subtasks (docs/stories/old_stories/theme-01-backend-scanner.md:L292)
- Technical Considerations (docs/stories/old_stories/theme-01-backend-scanner.md:L50)
- Tasks / Subtasks (docs/stories/old_stories/theme-04-settings-activation.md:L280)
- Tasks / Subtasks (docs/stories/old_stories/theme-05-monaco-registration.md:L255)
- vsix-02-frontend-vsix-path-handling.md (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L1)
- Story vsix-02: Update frontend path handling for VSIX-encoded theme paths (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L1)
- BDD Test Scenarios (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L129)
- Developer Notes (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L13)
- Scenario 1: VSIX path parsing (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L131)
- Architecture (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L15)
- Scenario 2: Settings.svelte integration (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L164)
- Tasks / Subtasks (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L181)
- Definition of Done (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L198)
- Technical Considerations (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L81)
- Risks & Edge Cases (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L88)
- Description (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L9)
- Reference Files (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L94)
- Acceptance Criteria (docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md:L99)
- applyTheme() (frontend/src/lib/stores/theme.js:L38)
- registerImportedTheme() (frontend/src/lib/stores/theme.js:L52)
- themeInit.js (frontend/src/lib/themeInit.js:L1)
- restoreImportedThemeFromConfig() (frontend/src/lib/themeInit.js:L167)
- convertedCache (frontend/src/lib/themeInit.js:L17)
- extractExtensionId() (frontend/src/lib/themeInit.js:L31)
- makeThemeId() (frontend/src/lib/themeInit.js:L52)
- activateImportedTheme() (frontend/src/lib/themeInit.js:L99)
- SetImportedTheme() (frontend/wailsjs/go/main/App.js:L373)
- SetTheme() (frontend/wailsjs/go/main/App.js:L401)
- SetVSCodiumExtPath() (frontend/wailsjs/go/main/App.js:L417)
- GetDevDir() (frontend/wailsjs/go/main/App.js:L73)

# Depends on
- [App.js](/modules/app-js.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [loadBundledThemes](/modules/loadbundledthemes.md)
- [theme.js](/modules/theme-js.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)

# Inferred
- [themeConverter.ts](/modules/themeconverter-ts.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
