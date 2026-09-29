---
type: Module
title: MonacoEditor.svelte
description: "Graphify community 16: @xterm/addon-fit, @xterm/xterm, @xterm/xterm/css/xterm.css, frontend/package.json, frontend/src/components/MonacoEditor.svelte, frontend/src/components/Terminal.svelte, frontend"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: addon-fit, resource: "@xterm/addon-fit", last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: xterm, resource: "@xterm/xterm", last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: xterm, resource: "@xterm/xterm/css/xterm.css", last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: package, resource: frontend/package.json, last_modified: "2026-04-22T16:28:45+10:00", digest: 5a2c6b5e36bbccae }
  - { id: MonacoEditor, resource: frontend/src/components/MonacoEditor.svelte, last_modified: "2026-09-30T00:51:37+10:00", digest: 7739c288d141a53a }
  - { id: Terminal, resource: frontend/src/components/Terminal.svelte, last_modified: "2026-09-30T01:21:31+10:00", digest: ffd229a35749a5e9 }
  - { id: editorSettings, resource: frontend/src/lib/stores/editorSettings.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8f55b394c63663b4 }
  - { id: font, resource: frontend/src/lib/stores/font.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 06f56f59c2340c7a }
  - { id: pty, resource: frontend/src/types/pty.ts, last_modified: "2026-04-22T19:23:56+10:00", digest: 44628db4a836ad08 }
  - { id: css.contribution, resource: monaco-editor/esm/vs/basic-languages/css/css.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: go.contribution, resource: monaco-editor/esm/vs/basic-languages/go/go.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: html.contribution, resource: monaco-editor/esm/vs/basic-languages/html/html.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: javascript.contribution, resource: monaco-editor/esm/vs/basic-languages/javascript/javascript.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: markdown.contribution, resource: monaco-editor/esm/vs/basic-languages/markdown/markdown.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: python.contribution, resource: monaco-editor/esm/vs/basic-languages/python/python.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: rust.contribution, resource: monaco-editor/esm/vs/basic-languages/rust/rust.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: shell.contribution, resource: monaco-editor/esm/vs/basic-languages/shell/shell.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: sql.contribution, resource: monaco-editor/esm/vs/basic-languages/sql/sql.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: typescript.contribution, resource: monaco-editor/esm/vs/basic-languages/typescript/typescript.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: xml.contribution, resource: monaco-editor/esm/vs/basic-languages/xml/xml.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: yaml.contribution, resource: monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: editor.api, resource: monaco-editor/esm/vs/editor/editor.api.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
  - { id: monaco.contribution, resource: monaco-editor/esm/vs/language/json/monaco.contribution.js, last_modified: "2026-09-29T15:22:13Z", digest: missing }
---

# Files
- `@xterm/addon-fit`
- `@xterm/xterm`
- `@xterm/xterm/css/xterm.css`
- `frontend/package.json`
- `frontend/src/components/MonacoEditor.svelte`
- `frontend/src/components/Terminal.svelte`
- `frontend/src/lib/stores/editorSettings.js`
- `frontend/src/lib/stores/font.js`
- `frontend/src/types/pty.ts`
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
- `monaco-editor/esm/vs/editor/editor.api.js`
- `monaco-editor/esm/vs/language/json/monaco.contribution.js`

# Symbols
- @xterm/addon-fit (@xterm/addon-fit:)
- @xterm/xterm (@xterm/xterm:)
- @xterm/xterm/css/xterm.css (@xterm/xterm/css/xterm.css:)
- @xterm/addon-fit (frontend/package.json:L28)
- @xterm/xterm (frontend/package.json:L29)
- MonacoEditor.svelte (frontend/src/components/MonacoEditor.svelte:L1)
- if() (frontend/src/components/MonacoEditor.svelte:L340)
- getWorker() (frontend/src/components/MonacoEditor.svelte:L480)
- Terminal.svelte (frontend/src/components/Terminal.svelte:L1)
- sendResize() (frontend/src/components/Terminal.svelte:L243)
- localWs (frontend/src/components/Terminal.svelte:L260)
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
- pty.ts (frontend/src/types/pty.ts:L1)
- PtyDataEvent (frontend/src/types/pty.ts:L10)
- PtyExitEvent (frontend/src/types/pty.ts:L16)
- PtyResizeEvent (frontend/src/types/pty.ts:L22)
- ScreenshotInjectEvent (frontend/src/types/pty.ts:L32)
- LogLineKind (frontend/src/types/pty.ts:L38)
- AgentLogLine (frontend/src/types/pty.ts:L41)
- PtyResizeFrame (frontend/src/types/pty.ts:L51)
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
- monaco-editor/esm/vs/editor/editor.api.js (monaco-editor/esm/vs/editor/editor.api.js:)
- monaco-editor/esm/vs/language/json/monaco.contribution.js (monaco-editor/esm/vs/language/json/monaco.contribution.js:)

# Depends on
- [App.js](/modules/app-js.md)
- [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)
- [InputResponseModal.svelte](/modules/inputresponsemodal-svelte.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [ref_node_path](/modules/ref-node-path.md)
- [runtime.js](/modules/runtime-js.md)
- [SetEditorSettings](/modules/seteditorsettings.md)
- [Story: pty-06 — Frontend Terminal and Session Cleanup](/modules/story-pty-06-frontend-terminal-and-session-cleanup.md)
- [svelte](/modules/svelte.md)
- [Terminal.auth.test.ts](/modules/terminal-auth-test-ts.md)
- [theme.js](/modules/theme-js.md)

# Inferred
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
