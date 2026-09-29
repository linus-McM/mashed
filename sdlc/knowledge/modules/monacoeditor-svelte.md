---
type: Module
title: MonacoEditor.svelte
description: "Graphify community 17: frontend/src/components/MonacoEditor.svelte, frontend/src/lib/monacoTheme.js, frontend/src/lib/stores/editorSettings.js, frontend/src/lib/stores/font.js, frontend/src/lib/stores"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: MonacoEditor, resource: frontend/src/components/MonacoEditor.svelte, last_modified: "2026-04-22T18:59:31+10:00", digest: 7ea6153fc5c063fa }
  - { id: monacoTheme, resource: frontend/src/lib/monacoTheme.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8ae058a6abdd0dfa }
  - { id: editorSettings, resource: frontend/src/lib/stores/editorSettings.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8f55b394c63663b4 }
  - { id: font, resource: frontend/src/lib/stores/font.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 06f56f59c2340c7a }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: monaco-editor, resource: monaco-editor, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: css.contribution, resource: monaco-editor/esm/vs/basic-languages/css/css.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: go.contribution, resource: monaco-editor/esm/vs/basic-languages/go/go.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: html.contribution, resource: monaco-editor/esm/vs/basic-languages/html/html.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: javascript.contribution, resource: monaco-editor/esm/vs/basic-languages/javascript/javascript.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: markdown.contribution, resource: monaco-editor/esm/vs/basic-languages/markdown/markdown.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: python.contribution, resource: monaco-editor/esm/vs/basic-languages/python/python.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: rust.contribution, resource: monaco-editor/esm/vs/basic-languages/rust/rust.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: shell.contribution, resource: monaco-editor/esm/vs/basic-languages/shell/shell.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: sql.contribution, resource: monaco-editor/esm/vs/basic-languages/sql/sql.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: typescript.contribution, resource: monaco-editor/esm/vs/basic-languages/typescript/typescript.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: xml.contribution, resource: monaco-editor/esm/vs/basic-languages/xml/xml.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: yaml.contribution, resource: monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: editor, resource: monaco-editor/esm/vs/editor/editor.api, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: editor.api, resource: monaco-editor/esm/vs/editor/editor.api.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
  - { id: monaco.contribution, resource: monaco-editor/esm/vs/language/json/monaco.contribution.js, last_modified: "2026-09-29T10:00:58Z", digest: missing }
---

# Files
- `frontend/src/components/MonacoEditor.svelte`
- `frontend/src/lib/monacoTheme.js`
- `frontend/src/lib/stores/editorSettings.js`
- `frontend/src/lib/stores/font.js`
- `frontend/src/lib/stores/theme.js`
- `monaco-editor`
- `monaco-editor/esm/vs/basic-languages/css/css.contribution.js`
- `monaco-editor/esm/vs/basic-languages/go/go.contribution.js`
- `monaco-editor/esm/vs/basic-languages/html/html.contribution.js`
- `monaco-editor/esm/vs/basic-languages/javascript/javascript.contribution.js`
- `monaco-editor/esm/vs/basic-languages/markdown/markdown.contribution.js`
- `monaco-editor/esm/vs/basic-languages/python/python.contribution.js`
- `monaco-editor/esm/vs/basic-languages/rust/rust.contribution.js`
- `monaco-editor/esm/vs/basic-languages/shell/shell.contribution.js`
- `monaco-editor/esm/vs/basic-languages/sql/sql.contribution.js`
- `monaco-editor/esm/vs/basic-languages/typescript/typescript.contribution.js`
- `monaco-editor/esm/vs/basic-languages/xml/xml.contribution.js`
- `monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution.js`
- `monaco-editor/esm/vs/editor/editor.api`
- `monaco-editor/esm/vs/editor/editor.api.js`
- `monaco-editor/esm/vs/language/json/monaco.contribution.js`

# Symbols
- MonacoEditor.svelte (frontend/src/components/MonacoEditor.svelte:L1)
- if() (frontend/src/components/MonacoEditor.svelte:L337)
- getWorker() (frontend/src/components/MonacoEditor.svelte:L477)
- monacoTheme.js (frontend/src/lib/monacoTheme.js:L1)
- toMonacoId() (frontend/src/lib/monacoTheme.js:L15)
- getEditorFont() (frontend/src/lib/monacoTheme.js:L54)
- editorSettings.js (frontend/src/lib/stores/editorSettings.js:L1)
- editorSettings (frontend/src/lib/stores/editorSettings.js:L23)
- defaults (frontend/src/lib/stores/editorSettings.js:L7)
- font.js (frontend/src/lib/stores/font.js:L1)
- applyFont() (frontend/src/lib/stores/font.js:L17)
- DEFAULT_MONO_FONT (frontend/src/lib/stores/font.js:L3)
- DEFAULT_FONT_SIZE (frontend/src/lib/stores/font.js:L4)
- registerLocalFonts() (frontend/src/lib/stores/font.js:L42)
- currentMonoFont (frontend/src/lib/stores/font.js:L6)
- currentFontSize (frontend/src/lib/stores/font.js:L7)
- allThemes (frontend/src/lib/stores/theme.js:L25)
- monaco-editor (monaco-editor:)
- monaco-editor/esm/vs/basic-languages/css/css.contribution.js (monaco-editor/esm/vs/basic-languages/css/css.contribution.js:)
- monaco-editor/esm/vs/basic-languages/go/go.contribution.js (monaco-editor/esm/vs/basic-languages/go/go.contribution.js:)
- monaco-editor/esm/vs/basic-languages/html/html.contribution.js (monaco-editor/esm/vs/basic-languages/html/html.contribution.js:)
- monaco-editor/esm/vs/basic-languages/javascript/javascript.contribution.js (monaco-editor/esm/vs/basic-languages/javascript/javascript.contribution.js:)
- monaco-editor/esm/vs/basic-languages/markdown/markdown.contribution.js (monaco-editor/esm/vs/basic-languages/markdown/markdown.contribution.js:)
- monaco-editor/esm/vs/basic-languages/python/python.contribution.js (monaco-editor/esm/vs/basic-languages/python/python.contribution.js:)
- monaco-editor/esm/vs/basic-languages/rust/rust.contribution.js (monaco-editor/esm/vs/basic-languages/rust/rust.contribution.js:)
- monaco-editor/esm/vs/basic-languages/shell/shell.contribution.js (monaco-editor/esm/vs/basic-languages/shell/shell.contribution.js:)
- monaco-editor/esm/vs/basic-languages/sql/sql.contribution.js (monaco-editor/esm/vs/basic-languages/sql/sql.contribution.js:)
- monaco-editor/esm/vs/basic-languages/typescript/typescript.contribution.js (monaco-editor/esm/vs/basic-languages/typescript/typescript.contribution.js:)
- monaco-editor/esm/vs/basic-languages/xml/xml.contribution.js (monaco-editor/esm/vs/basic-languages/xml/xml.contribution.js:)
- monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution.js (monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution.js:)
- monaco-editor/esm/vs/editor/editor.api (monaco-editor/esm/vs/editor/editor.api:)
- monaco-editor/esm/vs/editor/editor.api.js (monaco-editor/esm/vs/editor/editor.api.js:)
- monaco-editor/esm/vs/language/json/monaco.contribution.js (monaco-editor/esm/vs/language/json/monaco.contribution.js:)

# Depends on
- [App.js](/modules/app-js.md)
- [SetEditorSettings](/modules/seteditorsettings.md)
- [svelte](/modules/svelte.md)
- [themeInit.js](/modules/themeinit-js.md)
- [vitest](/modules/vitest.md)

# Inferred
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)

# Features
- no feature plan names these files
