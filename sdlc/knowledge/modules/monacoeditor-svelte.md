---
type: Module
title: MonacoEditor.svelte
description: "Graphify community 16: @xterm/addon-fit, @xterm/xterm, @xterm/xterm/css/xterm.css, docs/playwright_cli_US_validate/pty-06-frontend-cleanup-report.md, frontend/package.json, frontend/src/App.svelte, fr"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: addon-fit, resource: "@xterm/addon-fit", last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: xterm, resource: "@xterm/xterm", last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: xterm, resource: "@xterm/xterm/css/xterm.css", last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: pty-06-frontend-cleanup-report, resource: docs/playwright_cli_US_validate/pty-06-frontend-cleanup-report.md, last_modified: "2026-04-09T21:07:51+10:00", digest: f5bcc9033b3d735b }
  - { id: package, resource: frontend/package.json, last_modified: "2026-04-22T16:28:45+10:00", digest: 5a2c6b5e36bbccae }
  - { id: App, resource: frontend/src/App.svelte, last_modified: "2026-04-23T11:09:52+10:00", digest: 10db755a7a0e5abe }
  - { id: MonacoEditor, resource: frontend/src/components/MonacoEditor.svelte, last_modified: "2026-09-30T00:51:37+10:00", digest: 7739c288d141a53a }
  - { id: Terminal, resource: frontend/src/components/Terminal.svelte, last_modified: "2026-09-30T01:21:31+10:00", digest: ffd229a35749a5e9 }
  - { id: TitleBar, resource: frontend/src/components/TitleBar.svelte, last_modified: "2026-04-23T14:09:41+10:00", digest: 24aa88a7c4d75635 }
  - { id: QuestionSnackbarStack, resource: frontend/src/components/bmad/QuestionSnackbarStack.svelte, last_modified: "2026-04-20T14:57:51+10:00", digest: c485ed709e057516 }
  - { id: monacoTheme, resource: frontend/src/lib/monacoTheme.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8ae058a6abdd0dfa }
  - { id: editorSettings, resource: frontend/src/lib/stores/editorSettings.js, last_modified: "2026-04-22T18:59:31+10:00", digest: 8f55b394c63663b4 }
  - { id: font, resource: frontend/src/lib/stores/font.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 06f56f59c2340c7a }
  - { id: markdownMenuSettings.test, resource: frontend/src/lib/stores/markdownMenuSettings.test.ts, last_modified: "2026-04-23T11:02:33+10:00", digest: 440326bc7b010a88 }
  - { id: markdownMenuSettings, resource: frontend/src/lib/stores/markdownMenuSettings.ts, last_modified: "2026-04-23T11:02:33+10:00", digest: 5fd83b95073decb2 }
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: themes, resource: frontend/src/lib/themes.js, last_modified: "2026-04-07T10:03:32+10:00", digest: b15d9cf4598628ae }
  - { id: pty, resource: frontend/src/types/pty.ts, last_modified: "2026-04-22T19:23:56+10:00", digest: 44628db4a836ad08 }
  - { id: Settings, resource: frontend/src/views/Settings.svelte, last_modified: "2026-04-23T11:09:52+10:00", digest: edc56fcaa6890012 }
  - { id: runtime, resource: frontend/wailsjs/runtime/runtime.js, last_modified: "2026-05-07T09:55:31+10:00", digest: e25fe86d3c590de7 }
  - { id: monaco-editor, resource: monaco-editor, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: css.contribution, resource: monaco-editor/esm/vs/basic-languages/css/css.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: go.contribution, resource: monaco-editor/esm/vs/basic-languages/go/go.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: html.contribution, resource: monaco-editor/esm/vs/basic-languages/html/html.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: javascript.contribution, resource: monaco-editor/esm/vs/basic-languages/javascript/javascript.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: markdown.contribution, resource: monaco-editor/esm/vs/basic-languages/markdown/markdown.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: python.contribution, resource: monaco-editor/esm/vs/basic-languages/python/python.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: rust.contribution, resource: monaco-editor/esm/vs/basic-languages/rust/rust.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: shell.contribution, resource: monaco-editor/esm/vs/basic-languages/shell/shell.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: sql.contribution, resource: monaco-editor/esm/vs/basic-languages/sql/sql.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: typescript.contribution, resource: monaco-editor/esm/vs/basic-languages/typescript/typescript.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: xml.contribution, resource: monaco-editor/esm/vs/basic-languages/xml/xml.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: yaml.contribution, resource: monaco-editor/esm/vs/basic-languages/yaml/yaml.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: editor, resource: monaco-editor/esm/vs/editor/editor.api, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: editor.api, resource: monaco-editor/esm/vs/editor/editor.api.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
  - { id: monaco.contribution, resource: monaco-editor/esm/vs/language/json/monaco.contribution.js, last_modified: "2026-09-29T20:55:29Z", digest: missing }
---

# Files
- `@xterm/addon-fit`
- `@xterm/xterm`
- `@xterm/xterm/css/xterm.css`
- `docs/playwright_cli_US_validate/pty-06-frontend-cleanup-report.md`
- `frontend/package.json`
- `frontend/src/App.svelte`
- `frontend/src/components/MonacoEditor.svelte`
- `frontend/src/components/Terminal.svelte`
- `frontend/src/components/TitleBar.svelte`
- `frontend/src/components/bmad/QuestionSnackbarStack.svelte`
- `frontend/src/lib/monacoTheme.js`
- `frontend/src/lib/stores/editorSettings.js`
- `frontend/src/lib/stores/font.js`
- `frontend/src/lib/stores/markdownMenuSettings.test.ts`
- `frontend/src/lib/stores/markdownMenuSettings.ts`
- `frontend/src/lib/stores/theme.js`
- `frontend/src/lib/themes.js`
- `frontend/src/types/pty.ts`
- `frontend/src/views/Settings.svelte`
- `frontend/wailsjs/runtime/runtime.js`
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
- @xterm/addon-fit (@xterm/addon-fit:)
- @xterm/xterm (@xterm/xterm:)
- @xterm/xterm/css/xterm.css (@xterm/xterm/css/xterm.css:)
- AC-5: Clipboard integration unchanged — BLOCKED (docs/playwright_cli_US_validate/pty-06-frontend-cleanup-report.md:L68)
- @xterm/addon-fit (frontend/package.json:L28)
- @xterm/xterm (frontend/package.json:L29)
- App.svelte (frontend/src/App.svelte:L1)
- if() (frontend/src/App.svelte:L344)
- MonacoEditor.svelte (frontend/src/components/MonacoEditor.svelte:L1)
- if() (frontend/src/components/MonacoEditor.svelte:L340)
- getWorker() (frontend/src/components/MonacoEditor.svelte:L480)
- Terminal.svelte (frontend/src/components/Terminal.svelte:L1)
- copyText() (frontend/src/components/Terminal.svelte:L153)
- sendResize() (frontend/src/components/Terminal.svelte:L243)
- localWs (frontend/src/components/Terminal.svelte:L260)
- TitleBar.svelte (frontend/src/components/TitleBar.svelte:L1)
- active (frontend/src/components/TitleBar.svelte:L156)
- QuestionSnackbarStack.svelte (frontend/src/components/bmad/QuestionSnackbarStack.svelte:L1)
- questions (frontend/src/components/bmad/QuestionSnackbarStack.svelte:L10)
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
- markdownMenuSettings.test.ts (frontend/src/lib/stores/markdownMenuSettings.test.ts:L1)
- mocks (frontend/src/lib/stores/markdownMenuSettings.test.ts:L18)
- DEFAULTS (frontend/src/lib/stores/markdownMenuSettings.test.ts:L22)
- markdownMenuSettings.ts (frontend/src/lib/stores/markdownMenuSettings.ts:L1)
- DEFAULTS (frontend/src/lib/stores/markdownMenuSettings.ts:L13)
- markdownMenuDirty (frontend/src/lib/stores/markdownMenuSettings.ts:L23)
- theme.js (frontend/src/lib/stores/theme.js:L1)
- themes (frontend/src/lib/stores/theme.js:L18)
- builtInThemeIds (frontend/src/lib/stores/theme.js:L19)
- allThemes (frontend/src/lib/stores/theme.js:L25)
- themeIds (frontend/src/lib/stores/theme.js:L28)
- currentThemeId (frontend/src/lib/stores/theme.js:L30)
- currentTheme (frontend/src/lib/stores/theme.js:L32)
- themes.js (frontend/src/lib/themes.js:L1)
- themeIds (frontend/src/lib/themes.js:L228)
- DEFAULT_THEME (frontend/src/lib/themes.js:L229)
- themes (frontend/src/lib/themes.js:L4)
- pty.ts (frontend/src/types/pty.ts:L1)
- PtyDataEvent (frontend/src/types/pty.ts:L10)
- PtyExitEvent (frontend/src/types/pty.ts:L16)
- PtyResizeEvent (frontend/src/types/pty.ts:L22)
- ScreenshotInjectEvent (frontend/src/types/pty.ts:L32)
- LogLineKind (frontend/src/types/pty.ts:L38)
- AgentLogLine (frontend/src/types/pty.ts:L41)
- PtyResizeFrame (frontend/src/types/pty.ts:L51)
- Settings.svelte (frontend/src/views/Settings.svelte:L1)
- ClipboardGetText() (frontend/wailsjs/runtime/runtime.js:L200)
- ClipboardSetText() (frontend/wailsjs/runtime/runtime.js:L204)
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
- [applyTheme](/modules/applytheme.md)
- [bmadEvents.ts](/modules/bmadevents-ts.md)
- [clearMarkdownMenuDirty](/modules/clearmarkdownmenudirty.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [DynamicUiSelector.svelte](/modules/dynamicuiselector-svelte.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [markdownToolbarBuilder.test.ts](/modules/markdowntoolbarbuilder-test-ts.md)
- [models.ts](/modules/models-ts.md)
- [NewSessionModal.svelte](/modules/newsessionmodal-svelte.md)
- [PR 1: Security (branch `repo-health/pr1-security`)](/modules/pr-1-security-branch-repo-health-pr1-security.md)
- [QuestionSnackbarStack.test.ts](/modules/questionsnackbarstack-test-ts.md)
- [ref_node_fs](/modules/ref-node-fs.md)
- [runtime.js](/modules/runtime-js.md)
- [SetEditorSettings](/modules/seteditorsettings.md)
- [SetMarkdownMenuSettings](/modules/setmarkdownmenusettings.md)
- [svelte](/modules/svelte.md)
- [Terminal.auth.test.ts](/modules/terminal-auth-test-ts.md)
- [themeInit.js](/modules/themeinit-js.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- [autoFill.test.ts](/modules/autofill-test-ts.md)
- [CodeEditor.svelte](/modules/codeeditor-svelte.md)
- [EventsEmit](/modules/eventsemit.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
