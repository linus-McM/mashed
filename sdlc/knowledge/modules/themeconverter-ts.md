---
type: Module
title: themeConverter.ts
description: "Graphify community 34: docs/stories/old_stories/theme-02-converter.md, docs/stories/svelte-check-01-js-stores-to-ts.md, docs/stories/svelte-check-complete-report.md, frontend/src/lib/themeConverter.te"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:14Z" }
stale_after: "2026-10-13T12:23:14Z"
source_commit: ab1f2eb4ec65fc2b1fce15503c11f0e310758404
sources:
  - { id: theme-02-converter, resource: docs/stories/old_stories/theme-02-converter.md, last_modified: "2026-04-08T10:23:03+10:00", digest: f358caa9c259f69b }
  - { id: svelte-check-01-js-stores-to-ts, resource: docs/stories/svelte-check-01-js-stores-to-ts.md, last_modified: "2026-04-22T17:06:36+10:00", digest: 3cdcd1064aabe570 }
  - { id: svelte-check-complete-report, resource: docs/stories/svelte-check-complete-report.md, last_modified: "2026-04-22T20:20:54+10:00", digest: 19f8a1ad90aa5249 }
  - { id: themeConverter.test, resource: frontend/src/lib/themeConverter.test.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 687a4060b9b73676 }
  - { id: themeConverter, resource: frontend/src/lib/themeConverter.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: dccd00c61ec681ef }
  - { id: session, resource: frontend/src/types/session.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 0b3fa196502eba06 }
  - { id: theme, resource: frontend/src/types/theme.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 0c70f7df33f92e72 }
---

# Files
- `docs/stories/old_stories/theme-02-converter.md`
- `docs/stories/svelte-check-01-js-stores-to-ts.md`
- `docs/stories/svelte-check-complete-report.md`
- `frontend/src/lib/themeConverter.test.ts`
- `frontend/src/lib/themeConverter.ts`
- `frontend/src/types/session.ts`
- `frontend/src/types/theme.ts`

# Symbols
- theme-02-converter.md (docs/stories/old_stories/theme-02-converter.md:L1)
- Story 2: Frontend Theme Converter (B2 Hand-Rolled) (docs/stories/old_stories/theme-02-converter.md:L1)
- Reference Files (docs/stories/old_stories/theme-02-converter.md:L107)
- Acceptance Criteria (docs/stories/old_stories/theme-02-converter.md:L113)
- Developer Notes (docs/stories/old_stories/theme-02-converter.md:L13)
- Architecture (docs/stories/old_stories/theme-02-converter.md:L15)
- Tasks / Subtasks (docs/stories/old_stories/theme-02-converter.md:L259)
- Technical Considerations (docs/stories/old_stories/theme-02-converter.md:L29)
- Definition of Done (docs/stories/old_stories/theme-02-converter.md:L298)
- Description (docs/stories/old_stories/theme-02-converter.md:L9)
- Risks & Edge Cases (docs/stories/old_stories/theme-02-converter.md:L99)
- Tasks / Subtasks (docs/stories/svelte-check-01-js-stores-to-ts.md:L156)
- Architecture (docs/stories/svelte-check-01-js-stores-to-ts.md:L16)
- Deferred follow-ups (logged during migration) (docs/stories/svelte-check-complete-report.md:L67)
- themeConverter.test.ts (frontend/src/lib/themeConverter.test.ts:L1)
- DRACULA_THEME (frontend/src/lib/themeConverter.test.ts:L14)
- LIGHT_THEME (frontend/src/lib/themeConverter.test.ts:L72)
- HC_BLACK_THEME (frontend/src/lib/themeConverter.test.ts:L86)
- HC_LIGHT_THEME (frontend/src/lib/themeConverter.test.ts:L97)
- themeConverter.ts (frontend/src/lib/themeConverter.ts:L1)
- RawColors (frontend/src/lib/themeConverter.ts:L107)
- colorString() (frontend/src/lib/themeConverter.ts:L110)
- mapVSCodeColorsToCSSVars() (frontend/src/lib/themeConverter.ts:L119)
- convertTokenColors() (frontend/src/lib/themeConverter.ts:L177)
- ScopeMapping (frontend/src/lib/themeConverter.ts:L21)
- deriveXtermTheme() (frontend/src/lib/themeConverter.ts:L228)
- getMonacoBase() (frontend/src/lib/themeConverter.ts:L253)
- SCOPE_TO_TOKEN (frontend/src/lib/themeConverter.ts:L26)
- convertVSCodeTheme() (frontend/src/lib/themeConverter.ts:L271)
- BG_VARS (frontend/src/lib/themeConverter.ts:L42)
- REQUIRED_CSS_VARS (frontend/src/lib/themeConverter.ts:L50)
- normalizeHex() (frontend/src/lib/themeConverter.ts:L66)
- stripAlpha() (frontend/src/lib/themeConverter.ts:L83)
- dimColor() (frontend/src/lib/themeConverter.ts:L92)
- SessionState (frontend/src/types/session.ts:L48)
- theme.ts (frontend/src/types/theme.ts:L1)
- ThemeColors (frontend/src/types/theme.ts:L108)
- Theme (frontend/src/types/theme.ts:L114)
- VSCodeTokenColor (frontend/src/types/theme.ts:L18)
- VSCodeTheme (frontend/src/types/theme.ts:L32)
- MonacoBase (frontend/src/types/theme.ts:L45)
- MonacoTokenRule (frontend/src/types/theme.ts:L48)
- MonacoThemeDef (frontend/src/types/theme.ts:L56)
- XtermTheme (frontend/src/types/theme.ts:L68)
- ThemeColorKey (frontend/src/types/theme.ts:L89)

# Depends on
- [themeInit.js](/modules/themeinit-js.md)
- [vitest](/modules/vitest.md)

# Inferred
- [workflowSerialisation.ts](/modules/workflowserialisation-ts.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
