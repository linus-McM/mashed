---
type: Module
title: themeInit.js
description: "Graphify community 57: docs/SPECIFICATION.md, docs/stories/old_stories/edset-04-autoload-bundled-themes.md, docs/stories/old_stories/theme-04-settings-activation.md, docs/stories/old_stories/theme-05-"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: edset-04-autoload-bundled-themes, resource: docs/stories/old_stories/edset-04-autoload-bundled-themes.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 486391218b5c845a }
  - { id: theme-04-settings-activation, resource: docs/stories/old_stories/theme-04-settings-activation.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 5604998bc5d1ad34 }
  - { id: theme-05-monaco-registration, resource: docs/stories/old_stories/theme-05-monaco-registration.md, last_modified: "2026-04-08T10:23:03+10:00", digest: dcb77471f481713a }
  - { id: vsix-01-backend-zip-reading, resource: docs/stories/old_stories/vsix-01-backend-zip-reading.md, last_modified: "2026-04-08T10:23:03+10:00", digest: b5c56b368a35836e }
  - { id: vsix-02-frontend-vsix-path-handling, resource: docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 26f44fbbb6ff8bbb }
  - { id: vsix-sprint-backlog, resource: docs/stories/old_stories/vsix-sprint-backlog.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 2453bfb7bf7a10fb }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: themeConverter, resource: frontend/src/lib/themeConverter.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: dccd00c61ec681ef }
  - { id: themeInit, resource: frontend/src/lib/themeInit.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 30161049c3a59f8f }
  - { id: themes, resource: frontend/src/lib/themes.js, last_modified: "2026-04-07T10:03:32+10:00", digest: b15d9cf4598628ae }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/old_stories/edset-04-autoload-bundled-themes.md`
- `docs/stories/old_stories/theme-04-settings-activation.md`
- `docs/stories/old_stories/theme-05-monaco-registration.md`
- `docs/stories/old_stories/vsix-01-backend-zip-reading.md`
- `docs/stories/old_stories/vsix-02-frontend-vsix-path-handling.md`
- `docs/stories/old_stories/vsix-sprint-backlog.md`
- `frontend/src/lib/stores/theme.js`
- `frontend/src/lib/themeConverter.ts`
- `frontend/src/lib/themeInit.js`
- `frontend/src/lib/themes.js`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Theme Management (docs/SPECIFICATION.md:L852)
- edset-04-autoload-bundled-themes.md (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L1)
- Story 4: Auto-Load Bundled Themes from ./themes/ (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L1)
- BDD Test Scenarios (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L104)
- Scenario 1: Bundled theme discovery (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L106)
- Developer Notes (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L13)
- Scenario 2: Theme file reading (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L131)
- Scenario 3: Frontend auto-load (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L146)
- Architecture (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L15)
- Scenario 4: Refactor preserves behavior (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L171)
- Tasks / Subtasks (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L181)
- Definition of Done (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L206)
- Technical Considerations (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L41)
- Risks & Edge Cases (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L49)
- Reference Files (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L57)
- Acceptance Criteria (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L67)
- Description (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L9)
- theme-04-settings-activation.md (docs/stories/old_stories/theme-04-settings-activation.md:L1)
- Story 4: Settings UI -- Theme Scanning, Activation & Startup Restore (docs/stories/old_stories/theme-04-settings-activation.md:L1)
- Risks & Edge Cases (docs/stories/old_stories/theme-04-settings-activation.md:L122)
- Developer Notes (docs/stories/old_stories/theme-04-settings-activation.md:L13)
- Reference Files (docs/stories/old_stories/theme-04-settings-activation.md:L130)
- Acceptance Criteria (docs/stories/old_stories/theme-04-settings-activation.md:L138)
- Architecture (docs/stories/old_stories/theme-04-settings-activation.md:L15)
- Tasks / Subtasks (docs/stories/old_stories/theme-04-settings-activation.md:L280)
- Definition of Done (docs/stories/old_stories/theme-04-settings-activation.md:L319)
- Technical Considerations (docs/stories/old_stories/theme-04-settings-activation.md:L73)
- Description (docs/stories/old_stories/theme-04-settings-activation.md:L9)
- Tasks / Subtasks (docs/stories/old_stories/theme-05-monaco-registration.md:L255)
- vsix-01-backend-zip-reading.md (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L1)
- Story vsix-01: Read themes directly from .vsix zip archives (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L1)
- Reference Files (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L101)
- Acceptance Criteria (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L107)
- Developer Notes (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L13)
- Architecture (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L15)
- Tasks / Subtasks (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L270)
- Definition of Done (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L307)
- Technical Considerations (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L82)
- Description (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L9)
- Risks & Edge Cases (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L91)
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
- vsix-sprint-backlog.md (docs/stories/old_stories/vsix-sprint-backlog.md:L1)
- Sprint Backlog: Read themes directly from .vsix files (docs/stories/old_stories/vsix-sprint-backlog.md:L1)
- Summary (docs/stories/old_stories/vsix-sprint-backlog.md:L14)
- Sprint Backlog (docs/stories/old_stories/vsix-sprint-backlog.md:L3)
- theme.js (frontend/src/lib/stores/theme.js:L1)
- themes (frontend/src/lib/stores/theme.js:L18)
- builtInThemeIds (frontend/src/lib/stores/theme.js:L19)
- themeIds (frontend/src/lib/stores/theme.js:L28)
- currentThemeId (frontend/src/lib/stores/theme.js:L30)
- currentTheme (frontend/src/lib/stores/theme.js:L32)
- applyTheme() (frontend/src/lib/stores/theme.js:L38)
- registerImportedTheme() (frontend/src/lib/stores/theme.js:L52)
- registerSavedThemes() (frontend/src/lib/stores/theme.js:L57)
- unregisterImportedTheme() (frontend/src/lib/stores/theme.js:L62)
- validateConvertedTheme() (frontend/src/lib/themeConverter.ts:L308)
- themeInit.js (frontend/src/lib/themeInit.js:L1)
- restoreImportedThemeFromConfig() (frontend/src/lib/themeInit.js:L167)
- convertedCache (frontend/src/lib/themeInit.js:L17)
- loadBundledThemes() (frontend/src/lib/themeInit.js:L193)
- extractExtensionId() (frontend/src/lib/themeInit.js:L31)
- makeThemeId() (frontend/src/lib/themeInit.js:L52)
- loadSavedThemes() (frontend/src/lib/themeInit.js:L68)
- removeImportedTheme() (frontend/src/lib/themeInit.js:L86)
- activateImportedTheme() (frontend/src/lib/themeInit.js:L99)
- themes.js (frontend/src/lib/themes.js:L1)
- themeIds (frontend/src/lib/themes.js:L228)
- DEFAULT_THEME (frontend/src/lib/themes.js:L229)
- themes (frontend/src/lib/themes.js:L4)
- GetSavedThemes() (frontend/wailsjs/go/main/App.js:L101)
- ListBundledThemes() (frontend/wailsjs/go/main/App.js:L209)
- ListVSCodiumThemes() (frontend/wailsjs/go/main/App.js:L249)
- ReadBundledThemeFile() (frontend/wailsjs/go/main/App.js:L277)
- ReadThemeFile() (frontend/wailsjs/go/main/App.js:L297)
- RemoveTheme() (frontend/wailsjs/go/main/App.js:L301)
- SaveTheme() (frontend/wailsjs/go/main/App.js:L337)
- SetImportedTheme() (frontend/wailsjs/go/main/App.js:L369)

# Depends on
- [App.js](/modules/app-js.md)
- [frontend/package.json](/modules/frontend-package-json.md)
- [GetConfig](/modules/getconfig.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [svelte](/modules/svelte.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- [GetConfig](/modules/getconfig.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)

# Features
- no feature plan names these files
