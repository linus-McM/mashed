---
type: Module
title: theme.js
description: "Graphify community 17: docs/stories/old_stories/theme-03-store-refactor.md, frontend/src/lib/monacoTheme.js, frontend/src/lib/stores/font.js, frontend/src/lib/stores/theme.js, frontend/src/lib/themeIn"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: theme-03-store-refactor, resource: docs/stories/old_stories/theme-03-store-refactor.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 093f8b49cf309f4f }
  - { id: monacoTheme, resource: frontend/src/lib/monacoTheme.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8ae058a6abdd0dfa }
  - { id: font, resource: frontend/src/lib/stores/font.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 06f56f59c2340c7a }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: themeInit, resource: frontend/src/lib/themeInit.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 30161049c3a59f8f }
  - { id: themes, resource: frontend/src/lib/themes.js, last_modified: "2026-04-07T10:03:32+10:00", digest: b15d9cf4598628ae }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/stories/old_stories/theme-03-store-refactor.md`
- `frontend/src/lib/monacoTheme.js`
- `frontend/src/lib/stores/font.js`
- `frontend/src/lib/stores/theme.js`
- `frontend/src/lib/themeInit.js`
- `frontend/src/lib/themes.js`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Risks & Edge Cases (docs/stories/old_stories/theme-03-store-refactor.md:L96)
- monacoTheme.js (frontend/src/lib/monacoTheme.js:L1)
- toMonacoId() (frontend/src/lib/monacoTheme.js:L15)
- getEditorFont() (frontend/src/lib/monacoTheme.js:L54)
- currentMonoFont (frontend/src/lib/stores/font.js:L6)
- theme.js (frontend/src/lib/stores/theme.js:L1)
- themes (frontend/src/lib/stores/theme.js:L18)
- builtInThemeIds (frontend/src/lib/stores/theme.js:L19)
- allThemes (frontend/src/lib/stores/theme.js:L25)
- themeIds (frontend/src/lib/stores/theme.js:L28)
- currentThemeId (frontend/src/lib/stores/theme.js:L30)
- currentTheme (frontend/src/lib/stores/theme.js:L32)
- registerImportedTheme() (frontend/src/lib/stores/theme.js:L52)
- unregisterImportedTheme() (frontend/src/lib/stores/theme.js:L62)
- removeImportedTheme() (frontend/src/lib/themeInit.js:L86)
- themes.js (frontend/src/lib/themes.js:L1)
- themeIds (frontend/src/lib/themes.js:L228)
- DEFAULT_THEME (frontend/src/lib/themes.js:L229)
- themes (frontend/src/lib/themes.js:L4)
- RemoveTheme() (frontend/wailsjs/go/main/App.js:L301)

# Depends on
- [App.js](/modules/app-js.md)
- [applyTheme](/modules/applytheme.md)
- [themeInit.js](/modules/themeinit-js.md)
- [vitest](/modules/vitest.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
