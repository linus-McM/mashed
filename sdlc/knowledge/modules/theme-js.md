---
type: Module
title: theme.js
description: "Graphify community 143: frontend/src/lib/monacoTheme.js, frontend/src/lib/stores/theme.js, frontend/src/lib/themeInit.js, frontend/src/lib/themes.js, frontend/wailsjs/go/main/App.js, monaco-editor, mo"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: monacoTheme, resource: frontend/src/lib/monacoTheme.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8ae058a6abdd0dfa }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: themeInit, resource: frontend/src/lib/themeInit.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 30161049c3a59f8f }
  - { id: themes, resource: frontend/src/lib/themes.js, last_modified: "2026-04-07T10:03:32+10:00", digest: b15d9cf4598628ae }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
  - { id: monaco-editor, resource: monaco-editor, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: editor, resource: monaco-editor/esm/vs/editor/editor.api, last_modified: "2026-09-29T15:22:13Z", digest: missing }
---

# Files
- `frontend/src/lib/monacoTheme.js`
- `frontend/src/lib/stores/theme.js`
- `frontend/src/lib/themeInit.js`
- `frontend/src/lib/themes.js`
- `frontend/wailsjs/go/main/App.js`
- `monaco-editor`
- `monaco-editor/esm/vs/editor/editor.api`

# Symbols
- monacoTheme.js (frontend/src/lib/monacoTheme.js:L1)
- toMonacoId() (frontend/src/lib/monacoTheme.js:L15)
- getEditorFont() (frontend/src/lib/monacoTheme.js:L54)
- theme.js (frontend/src/lib/stores/theme.js:L1)
- themes (frontend/src/lib/stores/theme.js:L18)
- builtInThemeIds (frontend/src/lib/stores/theme.js:L19)
- allThemes (frontend/src/lib/stores/theme.js:L25)
- themeIds (frontend/src/lib/stores/theme.js:L28)
- currentThemeId (frontend/src/lib/stores/theme.js:L30)
- currentTheme (frontend/src/lib/stores/theme.js:L32)
- unregisterImportedTheme() (frontend/src/lib/stores/theme.js:L62)
- removeImportedTheme() (frontend/src/lib/themeInit.js:L86)
- themes.js (frontend/src/lib/themes.js:L1)
- themeIds (frontend/src/lib/themes.js:L228)
- DEFAULT_THEME (frontend/src/lib/themes.js:L229)
- themes (frontend/src/lib/themes.js:L4)
- RemoveTheme() (frontend/wailsjs/go/main/App.js:L305)
- monaco-editor (monaco-editor:)
- monaco-editor/esm/vs/editor/editor.api (monaco-editor/esm/vs/editor/editor.api:)

# Depends on
- [InputResponseModal.svelte](/modules/inputresponsemodal-svelte.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [loadBundledThemes](/modules/loadbundledthemes.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [themeInit.js](/modules/themeinit-js.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
