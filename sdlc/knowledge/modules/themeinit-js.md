---
type: Module
title: themeInit.js
description: "Graphify community 115: docs/SPECIFICATION.md, docs/stories/old_stories/edset-04-autoload-bundled-themes.md, docs/stories/old_stories/vsix-01-backend-zip-reading.md, frontend/src/lib/stores/theme.js,"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: edset-04-autoload-bundled-themes, resource: docs/stories/old_stories/edset-04-autoload-bundled-themes.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 486391218b5c845a }
  - { id: vsix-01-backend-zip-reading, resource: docs/stories/old_stories/vsix-01-backend-zip-reading.md, last_modified: "2026-04-08T10:23:03+10:00", digest: b5c56b368a35836e }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: themeConverter, resource: frontend/src/lib/themeConverter.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: dccd00c61ec681ef }
  - { id: themeInit, resource: frontend/src/lib/themeInit.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 30161049c3a59f8f }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/old_stories/edset-04-autoload-bundled-themes.md`
- `docs/stories/old_stories/vsix-01-backend-zip-reading.md`
- `frontend/src/lib/stores/theme.js`
- `frontend/src/lib/themeConverter.ts`
- `frontend/src/lib/themeInit.js`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Theme Management (docs/SPECIFICATION.md:L852)
- edset-04-autoload-bundled-themes.md (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L1)
- Story 4: Auto-Load Bundled Themes from ./themes/ (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L1)
- Developer Notes (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L13)
- Architecture (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L15)
- Tasks / Subtasks (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L181)
- Definition of Done (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L206)
- Technical Considerations (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L41)
- Risks & Edge Cases (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L49)
- Reference Files (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L57)
- Acceptance Criteria (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L67)
- Description (docs/stories/old_stories/edset-04-autoload-bundled-themes.md:L9)
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
- registerSavedThemes() (frontend/src/lib/stores/theme.js:L57)
- validateConvertedTheme() (frontend/src/lib/themeConverter.ts:L308)
- themeInit.js (frontend/src/lib/themeInit.js:L1)
- convertedCache (frontend/src/lib/themeInit.js:L17)
- loadBundledThemes() (frontend/src/lib/themeInit.js:L193)
- loadSavedThemes() (frontend/src/lib/themeInit.js:L68)
- GetSavedThemes() (frontend/wailsjs/go/main/App.js:L101)
- ListBundledThemes() (frontend/wailsjs/go/main/App.js:L213)
- ListVSCodiumThemes() (frontend/wailsjs/go/main/App.js:L253)
- ReadBundledThemeFile() (frontend/wailsjs/go/main/App.js:L281)
- ReadThemeFile() (frontend/wailsjs/go/main/App.js:L301)
- SaveTheme() (frontend/wailsjs/go/main/App.js:L341)

# Depends on
- [activateImportedTheme](/modules/activateimportedtheme.md)
- [App.js](/modules/app-js.md)
- [applyTheme](/modules/applytheme.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [PR 1: Security (branch `repo-health/pr1-security`)](/modules/pr-1-security-branch-repo-health-pr1-security.md)
- [SetTheme](/modules/settheme.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)

# Inferred
- [activateImportedTheme](/modules/activateimportedtheme.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
