---
type: Module
title: loadBundledThemes
description: "Graphify community 115: docs/SPECIFICATION.md, docs/stories/old_stories/edset-04-autoload-bundled-themes.md, frontend/src/lib/stores/theme.js, frontend/src/lib/themeConverter.ts, frontend/src/lib/them"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: edset-04-autoload-bundled-themes, resource: docs/stories/old_stories/edset-04-autoload-bundled-themes.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 486391218b5c845a }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: themeConverter, resource: frontend/src/lib/themeConverter.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: dccd00c61ec681ef }
  - { id: themeInit, resource: frontend/src/lib/themeInit.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 30161049c3a59f8f }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
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
- registerSavedThemes() (frontend/src/lib/stores/theme.js:L57)
- validateConvertedTheme() (frontend/src/lib/themeConverter.ts:L308)
- loadBundledThemes() (frontend/src/lib/themeInit.js:L193)
- loadSavedThemes() (frontend/src/lib/themeInit.js:L68)
- GetSavedThemes() (frontend/wailsjs/go/main/App.js:L101)
- ListBundledThemes() (frontend/wailsjs/go/main/App.js:L213)
- ReadBundledThemeFile() (frontend/wailsjs/go/main/App.js:L281)
- SaveTheme() (frontend/wailsjs/go/main/App.js:L341)

# Depends on
- [themeConverter.ts](/modules/themeconverter-ts.md)
- [themeInit.js](/modules/themeinit-js.md)

# Inferred
- [themeConverter.ts](/modules/themeconverter-ts.md)
- [themeInit.js](/modules/themeinit-js.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
