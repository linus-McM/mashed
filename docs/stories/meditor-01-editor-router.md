# Story 1: EditorRouter -- Extension-Based Editor Switching

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** S
**Depends On:** none
**Status:** done

## Description

Create an `EditorRouter.svelte` component that inspects a file's extension and mounts the correct viewer component (MonacoEditor for code, ImageViewer for images, MarkdownEditor for markdown). Replace the direct `<MonacoEditor>` usage in `AgentDetail.svelte` with `<EditorRouter>`. This is the foundation component that all subsequent multi-editor stories build on.

## Developer Notes

### Architecture
- **New file:** `frontend/src/components/EditorRouter.svelte`
- **Modified file:** `frontend/src/views/AgentDetail.svelte` (lines 9, 547-555)
- EditorRouter receives the same props as MonacoEditor: `filePath`, `repoPath`, `mode`, `editable`
- Extension mapping function `getEditorType(ext)` returns `'markdown'` | `'image'` | `'code'`
  - Markdown extensions: `['md', 'mdx', 'markdown']`
  - Image extensions: `['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'bmp', 'ico']`
  - Everything else: `'code'` (falls through to MonacoEditor)
- Use Svelte `{#if}` blocks for conditional rendering -- Svelte automatically destroys the previous component and mounts the new one when `editorType` changes

### Technical Considerations
- **Phase 1 stubs:** Since ImageViewer and MarkdownEditor don't exist yet, create minimal placeholder components that display the file path and a "coming soon" message. This lets the router be tested immediately.
- The `mode` prop is only relevant for MonacoEditor (source/diff). EditorRouter passes it through to Monaco but not to other editors.
- The `editable` prop applies to Monaco and MarkdownEditor but not ImageViewer (always read-only).
- Svelte's `{#if}` lifecycle guarantees that switching file types destroys the old editor cleanly.

### Risks & Edge Cases
- Files with no extension should fall through to `'code'` (Monaco handles them fine)
- Files with uppercase extensions (`.MD`, `.PNG`) must be handled -- lowercase before matching
- Dot-only files (`.gitignore`, `.env`) have no meaningful extension -- should route to code
- `filePath` could be null/undefined if no file is selected -- guard with early return

### Reference Files
- `frontend/src/components/MonacoEditor.svelte` -- existing editor component (props interface)
- `frontend/src/views/AgentDetail.svelte` -- integration point at lines 547-555
- `docs/feasibility-multi-editor.md` -- Section 2.1 (EditorRouter design)

## Acceptance Criteria

AC-1: Extension routing logic
- Given a file with extension `.md`, `.mdx`, or `.markdown`
- When the EditorRouter receives the filePath
- Then it sets `editorType` to `'markdown'`
- And mounts the MarkdownEditor component (or placeholder)

AC-2: Image routing
- Given a file with extension `.png`, `.jpg`, `.jpeg`, `.gif`, `.svg`, `.webp`, `.bmp`, or `.ico`
- When the EditorRouter receives the filePath
- Then it sets `editorType` to `'image'`
- And mounts the ImageViewer component (or placeholder)

AC-3: Code fallback routing
- Given a file with any other extension (`.ts`, `.go`, `.json`, etc.) or no extension
- When the EditorRouter receives the filePath
- Then it sets `editorType` to `'code'`
- And mounts the MonacoEditor component with all original props (`filePath`, `repoPath`, `mode`, `editable`)

AC-4: AgentDetail integration
- Given the AgentDetail view is displaying a selected file
- When the user selects a file from the FileTree
- Then `<EditorRouter>` renders instead of `<MonacoEditor>`
- And all existing code editing functionality (auto-save, diff, read-only) works identically

AC-5: Case-insensitive extension matching
- Given a file named `README.MD` or `photo.PNG`
- When the EditorRouter processes the extension
- Then it correctly routes to markdown or image editor respectively

## BDD Test Scenarios

### Scenario 1: Extension-based routing
```gherkin
Feature: EditorRouter extension routing

  Scenario: Route markdown files to markdown editor
    Given the EditorRouter component is mounted
    And filePath is "/repo/docs/README.md"
    When the component evaluates the file extension
    Then editorType is "markdown"

  Scenario: Route image files to image viewer
    Given the EditorRouter component is mounted
    And filePath is "/repo/assets/logo.png"
    When the component evaluates the file extension
    Then editorType is "image"

  Scenario: Route code files to Monaco editor
    Given the EditorRouter component is mounted
    And filePath is "/repo/src/main.ts"
    When the component evaluates the file extension
    Then editorType is "code"

  Scenario: Handle files with no extension
    Given the EditorRouter component is mounted
    And filePath is "/repo/.gitignore"
    When the component evaluates the file extension
    Then editorType is "code"

  Scenario: Handle uppercase extensions
    Given the EditorRouter component is mounted
    And filePath is "/repo/CHANGELOG.MD"
    When the component evaluates the file extension
    Then editorType is "markdown"
```

### Scenario 2: Component lifecycle
```gherkin
Feature: EditorRouter component switching

  Scenario: Switch from code to markdown file
    Given the EditorRouter is displaying a MonacoEditor for "main.ts"
    When the user selects "README.md" from the file tree
    Then MonacoEditor is destroyed
    And MarkdownEditor (or placeholder) is mounted

  Scenario: Switch between two code files
    Given the EditorRouter is displaying MonacoEditor for "app.ts"
    When the user selects "utils.ts" from the file tree
    Then MonacoEditor receives the new filePath prop
    And MonacoEditor handles the file switch internally
```

## Tasks / Subtasks

- [ ] Task 1: Create EditorRouter.svelte (AC: 1, 2, 3, 5)
  - [ ] Subtask 1a: Implement `getEditorType(ext)` function with extension-to-editor mapping
  - [ ] Subtask 1b: Add reactive `$: ext` and `$: editorType` declarations
  - [ ] Subtask 1c: Add `{#if}` blocks for each editor type (markdown placeholder, image placeholder, MonacoEditor)
  - [ ] Subtask 1d: Create minimal placeholder components: `MarkdownEditorPlaceholder` and `ImageViewerPlaceholder` (inline or separate files)

- [ ] Task 2: Integrate into AgentDetail.svelte (AC: 4)
  - [ ] Subtask 2a: Replace `import MonacoEditor` with `import EditorRouter` at line 9
  - [ ] Subtask 2b: Replace `<MonacoEditor ... />` with `<EditorRouter ... />` at lines 549-554
  - [ ] Subtask 2c: Verify all props are passed through correctly

- [ ] Task 3: Unit tests for getEditorType (AC: 1, 2, 3, 5)
  - [ ] Subtask 3a: Test all markdown extensions including case variants
  - [ ] Subtask 3b: Test all image extensions including case variants
  - [ ] Subtask 3c: Test code fallback for various extensions and edge cases (no extension, dot-only files)

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `getEditorType` function has comprehensive unit tests
- [ ] Existing MonacoEditor functionality is not regressed (code files still work identically)
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
