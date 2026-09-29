---
type: Module
title: themeInit.js
description: "Graphify community 57: docs/SPECIFICATION.md, docs/stories/old_stories/edset-04-autoload-bundled-themes.md, docs/stories/old_stories/vsix-01-backend-zip-reading.md, frontend/src/lib/stores/theme.js, f"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: edset-04-autoload-bundled-themes, resource: docs/stories/old_stories/edset-04-autoload-bundled-themes.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 486391218b5c845a }
  - { id: vsix-01-backend-zip-reading, resource: docs/stories/old_stories/vsix-01-backend-zip-reading.md, last_modified: "2026-04-08T10:23:03+10:00", digest: b5c56b368a35836e }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: themeConverter, resource: frontend/src/lib/themeConverter.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: dccd00c61ec681ef }
  - { id: themeInit, resource: frontend/src/lib/themeInit.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 30161049c3a59f8f }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
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
- BDD Test Scenarios (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L150)
- Scenario 1: List themes from VSIX files (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L152)
- Scenario 2: Read theme file from VSIX (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L191)
- Scenario 3: Security and limits (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L230)
- Scenario 4: Backward compatibility (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L257)
- Tasks / Subtasks (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L270)
- Definition of Done (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L307)
- Technical Considerations (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L82)
- Description (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L9)
- Risks & Edge Cases (docs/stories/old_stories/vsix-01-backend-zip-reading.md:L91)
- registerSavedThemes() (frontend/src/lib/stores/theme.js:L57)
- unregisterImportedTheme() (frontend/src/lib/stores/theme.js:L62)
- validateConvertedTheme() (frontend/src/lib/themeConverter.ts:L308)
- themeInit.js (frontend/src/lib/themeInit.js:L1)
- convertedCache (frontend/src/lib/themeInit.js:L17)
- loadBundledThemes() (frontend/src/lib/themeInit.js:L193)
- loadSavedThemes() (frontend/src/lib/themeInit.js:L68)
- removeImportedTheme() (frontend/src/lib/themeInit.js:L86)
- GetSavedThemes() (frontend/wailsjs/go/main/App.js:L101)
- ListBundledThemes() (frontend/wailsjs/go/main/App.js:L209)
- ListVSCodiumThemes() (frontend/wailsjs/go/main/App.js:L249)
- ReadBundledThemeFile() (frontend/wailsjs/go/main/App.js:L277)
- ReadThemeFile() (frontend/wailsjs/go/main/App.js:L297)
- RemoveTheme() (frontend/wailsjs/go/main/App.js:L301)
- SaveTheme() (frontend/wailsjs/go/main/App.js:L337)

# Depends on
- [activateImportedTheme](/modules/activateimportedtheme.md)
- [App.js](/modules/app-js.md)
- [applyTheme](/modules/applytheme.md)
- [convertVSCodeTheme](/modules/convertvscodetheme.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [SetTheme](/modules/settheme.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- [activateImportedTheme](/modules/activateimportedtheme.md)
- [convertVSCodeTheme](/modules/convertvscodetheme.md)

# Features
- no feature plan names these files
