# Story svelte-check-04b: View-level state — Editors (MonacoEditor / MarkdownEditor / CodeEditor) (Phase 4b)

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** svelte-check-03
**Status:** ready

## Description

Retype the three editor components (MonacoEditor 58, MarkdownEditor 21, CodeEditor 25 = 104 direct errors) plus their shared `editorUtils.ts` / `markdownEditorUtils.ts`. Refs hold third-party library instances (`Terminal`, `monaco.editor.IStandaloneCodeEditor`, Milkdown `Crepe`), so the core of this work is typing container refs against library surfaces. Expected total reduction: ~104 direct + ~30 downstream = ~134. Count drops from ~770 to ~640. Can run in parallel with 04a and 04c.

## Developer Notes

### Architecture
- Files:
  - `frontend/src/components/MonacoEditor.svelte` (58)
  - `frontend/src/components/MarkdownEditor.svelte` (21)
  - `frontend/src/components/CodeEditor.svelte` (25)
  - `frontend/src/components/editorUtils.ts` (already TS — verify types, fix any gaps)
  - `frontend/src/components/markdownEditorUtils.ts` (already TS — same)
- Ref typing:
  - Monaco: `let editor: monaco.editor.IStandaloneCodeEditor | null = null;`
  - Milkdown Crepe: `import type { Crepe } from '@milkdown/crepe'; let crepe: Crepe | null = null;`
  - Container: `let container: HTMLDivElement | null = null;`
- Monaco types ship with the `monaco-editor` package; import them directly: `import type * as monaco from 'monaco-editor'`.
- Milkdown Crepe types ship with `@milkdown/crepe`; import as above.
- Theme integration: `editorSettings.js` and `monacoTheme.js` are JS from Phase 1 territory. Both may already be migrated or may still need conversion — verify. If not migrated, add them to this story as a prerequisite (keeps editor types coherent).

### Technical Considerations
- **Monaco global vs module type.** Prefer `import type * as monaco from 'monaco-editor'` for types, not for values — values come via dynamic import to preserve the code-split bundle.
- **Milkdown WYSIWYG.** Per cerebrum 2026-04-08, Crepe is the chosen editor. Do not replace or simplify its API during this story. Only type the existing surface.
- **xterm (Terminal.svelte)** is out of scope here — handled in 04c.
- **Event handlers on editor instances** (`editor.onDidChangeModelContent(...)`) — typed by monaco's own declarations. Do not retype.
- **dispose() on unmount.** `onDestroy(() => { editor?.dispose(); crepe?.destroy(); })`. Already present — just make sure the null check survives retyping.
- **Shared util modules.** `editorUtils.ts` and `markdownEditorUtils.ts` already have `.ts` extensions. Open each and verify there are no implicit `any`s; fix any issues in-place.

### Risks & Edge Cases
- **Bundle-size regressions.** `import type` is compile-only — no bundle impact. A value `import` from `monaco-editor` would bloat the bundle. Keep value imports dynamic (`await import('monaco-editor')`).
- **Theme store circular dependency.** `editorSettings` might import theme and vice versa. Retype together if needed.
- **Monaco disposed-editor race.** If `editor.dispose()` fires before an async `setModelLanguage` call resolves, we get a runtime error. Null-guard is the existing mitigation — retyping must preserve it.
- **Milkdown ESM quirks.** Crepe's types occasionally lag its runtime; if an imported type is missing, declare a local `.d.ts` shim in `frontend/src/types/shims.d.ts` rather than using `any`.
- **R4 — PR sizing.** Likely fits in a single PR under 400 lines.

### Reference Files
- `frontend/src/components/MonacoEditor.svelte` (target)
- `frontend/src/components/MarkdownEditor.svelte` (target)
- `frontend/src/components/CodeEditor.svelte` (target)
- `frontend/src/components/editorUtils.ts` (shared helpers — verify types)
- `frontend/src/components/markdownEditorUtils.ts` (shared helpers — verify types)
- `frontend/src/components/EditorRouter.svelte` (dispatches between editors by file type; may need touch-ups)
- `frontend/src/components/DiffView.svelte` (may share monaco types — quick check)

### Skills to invoke
- `/simplify` — mandatory before commit.
- `/playwright-cli` — smoke monaco loads a TS file, markdown WYSIWYG edit persists, code editor saves.

## Acceptance Criteria

AC-1: Editor refs typed against library surfaces
- Given MonacoEditor, MarkdownEditor, CodeEditor
- When this story lands
- Then `editor`, `container`, `crepe` refs have explicit library-sourced types (`monaco.editor.IStandaloneCodeEditor | null`, `Crepe | null`, `HTMLDivElement | null`)
- And no `let editor = …` is inferred as `any` or `object`

AC-2: All props typed
- Given each editor component's `export let`s
- When this story lands
- Then every prop has an explicit type annotation
- And all emitted custom events carry typed `detail` payloads

AC-3: Shared util modules clean
- Given `editorUtils.ts` and `markdownEditorUtils.ts`
- When `just sveltecheck-file` is run on each
- Then each reports 0 errors

AC-4: Error count drops
- Given entering at ~770 (post-Phase 4a if sequential; otherwise baseline may differ when run in parallel — verify)
- When this story lands
- Then the sum of `sveltecheck-file` for the three editor components is ≤ 10 remaining errors
- And `just sveltecheck-ratchet` passes

AC-5: Runtime unchanged
- Given monaco opens a `.ts` file, markdown opens a `.md` file, code editor loads a diff
- When a user opens each editor type after this story
- Then all editing behaviours (syntax highlight, save, WYSIWYG toggle, diff scroll) work identically

## BDD Test Scenarios

### Scenario 1: Monaco ref typing
```gherkin
Feature: Monaco ref typed

  Scenario: editor disposed safely
    Given let editor: monaco.editor.IStandaloneCodeEditor | null = null
    When onDestroy fires
    Then editor?.dispose() compiles
    And svelte-check reports no error

  Scenario: dynamic import preserved
    Given the value import is "await import('monaco-editor')"
    When vite builds
    Then monaco-editor is in a separate chunk, not the main bundle
```

### Scenario 2: Crepe type
```gherkin
Feature: Milkdown Crepe ref typed

  Scenario: Crepe instance ref
    Given let crepe: Crepe | null = null
    When setup runs "crepe = new Crepe({ root: container, defaultValue: md })"
    Then svelte-check passes
    And crepe.destroy() compiles
```

### Scenario 3: Props
```gherkin
Feature: Typed editor props

  Scenario: MonacoEditor initialContent type
    Given "export let initialContent: string = ''"
    When a parent passes initialContent={42}
    Then svelte-check reports a type error

  Scenario: MarkdownEditor on:change event detail
    Given dispatch("change", { markdown: string })
    When a parent types "(e: CustomEvent<{ markdown: string }>)"
    Then e.detail.markdown is string (no implicit any)
```

### Scenario 4: Runtime smoke
```gherkin
Feature: Editor runtime preserved

  Scenario: Monaco opens TS file
    Given a repo with "main.ts"
    When the user opens main.ts in MonacoEditor
    Then TS syntax highlighting appears
    And Ctrl+S triggers save
    And the underlying Go save call succeeds

  Scenario: Markdown WYSIWYG round-trip
    Given a .md file
    When the user types "**bold**" in Crepe
    Then the rendered bold shows immediately
    And the saved file contains the literal "**bold**"
```

## Tasks / Subtasks

- [ ] Task 1: Verify prerequisite JS→TS (AC: 1)
  - [ ] Check if `editorSettings.js`, `monacoTheme.js`, `sprintColors.js`, `themes.js`, `themeInit.js`, `fileTree.js` are still `.js`
  - [ ] If present and untyped, convert them or call out as blockers before proceeding

- [ ] Task 2: Retype MonacoEditor (AC: 1, 2)
  - [ ] Add `import type * as monaco from 'monaco-editor'`
  - [ ] Type `editor`, `container`, any model listeners
  - [ ] Type all `export let` props (language, initialContent, readonly, theme)

- [ ] Task 3: Retype MarkdownEditor (AC: 1, 2)
  - [ ] Add `import type { Crepe } from '@milkdown/crepe'`
  - [ ] Type `crepe`, `container`, dispose in `onDestroy`
  - [ ] Type props (initialMarkdown, readonly, etc.)

- [ ] Task 4: Retype CodeEditor (AC: 1, 2)
  - [ ] Identify its underlying library (likely monaco or a prism-based one)
  - [ ] Apply the same ref-typing approach
  - [ ] Type props and emitted events

- [ ] Task 5: Verify shared utils (AC: 3)
  - [ ] `just sveltecheck-file src/components/editorUtils.ts` — 0 errors
  - [ ] `just sveltecheck-file src/components/markdownEditorUtils.ts` — 0 errors
  - [ ] Fix any remaining issues in-place

- [ ] Task 6: Verify reductions and behaviour (AC: 4, 5)
  - [ ] Each target file ≤ a handful of errors
  - [ ] `vitest run` — 696 tests pass
  - [ ] `vite build` clean; confirm monaco-editor is in its own chunk
  - [ ] `wails dev` smoke: open .ts / .md / diff — all render and save
  - [ ] Commit body records delta

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] `go build ./...` / `go vet ./...` / `go test ./... -race` all pass
- [ ] Target file error counts at or below targets
- [ ] `just sveltecheck-ratchet` passes
- [ ] `vitest run` — 696 tests pass
- [ ] `vite build` clean; monaco-editor remains code-split
- [ ] Manual wails dev smoke of all three editors
- [ ] `/simplify` run on every modified file
- [ ] No `any`, no `@ts-ignore`, no `@ts-nocheck`
- [ ] PR size <= 400 lines (R4)
- [ ] Commit body records: `sveltecheck-count: <prev> → <cur> (Δ -<n>)`
