---
type: Module
title: themeInit.js
description: "Graphify community 57: docs/SPECIFICATION.md, docs/stories/old_stories/edset-04-autoload-bundled-themes.md, frontend/src/lib/stores/theme.js, frontend/src/lib/themeConverter.ts, frontend/src/lib/theme"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: edset-04-autoload-bundled-themes, resource: docs/stories/old_stories/edset-04-autoload-bundled-themes.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 486391218b5c845a }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: themeConverter, resource: frontend/src/lib/themeConverter.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: dccd00c61ec681ef }
  - { id: themeInit, resource: frontend/src/lib/themeInit.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 30161049c3a59f8f }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/old_stories/edset-04-autoload-bundled-themes.md`
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
- registerImportedTheme() (frontend/src/lib/stores/theme.js:L52)
- registerSavedThemes() (frontend/src/lib/stores/theme.js:L57)
- convertVSCodeTheme() (frontend/src/lib/themeConverter.ts:L271)
- validateConvertedTheme() (frontend/src/lib/themeConverter.ts:L308)
- themeInit.js (frontend/src/lib/themeInit.js:L1)
- convertedCache (frontend/src/lib/themeInit.js:L17)
- loadBundledThemes() (frontend/src/lib/themeInit.js:L193)
- loadSavedThemes() (frontend/src/lib/themeInit.js:L68)
- activateImportedTheme() (frontend/src/lib/themeInit.js:L99)
- GetSavedThemes() (frontend/wailsjs/go/main/App.js:L101)
- ListBundledThemes() (frontend/wailsjs/go/main/App.js:L209)
- ReadBundledThemeFile() (frontend/wailsjs/go/main/App.js:L277)
- SaveTheme() (frontend/wailsjs/go/main/App.js:L337)

# Depends on
- [App.js](/modules/app-js.md)
- [applyTheme](/modules/applytheme.md)
- [font.js](/modules/font-js.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [makeThemeId](/modules/makethemeid.md)
- [SetTheme](/modules/settheme.md)
- [Story 2: Frontend Theme Converter (B2 Hand-Rolled)](/modules/story-2-frontend-theme-converter-b2-hand-rolled.md)
- [theme.js](/modules/theme-js.md)
- [themeConverter.ts](/modules/themeconverter-ts.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
