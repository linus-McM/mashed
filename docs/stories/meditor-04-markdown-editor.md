# Story 4: MarkdownEditor with Milkdown Crepe WYSIWYG

**Priority:** P1-high
**Domain:** frontend
**Estimated Complexity:** L
**Depends On:** Story 1 (meditor-01)
**Status:** ready

## Description

Create a `MarkdownEditor.svelte` component using Milkdown Crepe for full WYSIWYG markdown editing. Users edit rendered markdown directly (like Notion) -- no raw source view, no split pane. The editor loads markdown via `ReadFile`, provides auto-save with 800ms debounce via `WriteFile`, and integrates with the mashed design system through a CSS variable bridge. Replace the placeholder in EditorRouter with the real MarkdownEditor.

## Developer Notes

### Architecture
- **New files:**
  - `frontend/src/components/MarkdownEditor.svelte` -- main component
  - `frontend/src/styles/crepe-mashed.css` -- CSS variable bridge (Crepe vars -> mashed design tokens)
- **Modified file:** `frontend/src/components/EditorRouter.svelte` -- replace markdown placeholder with real import
- **New dependency:** `@milkdown/crepe` -- `npm install @milkdown/crepe`
- Props: `filePath` (string), `repoPath` (string), `editable` (boolean)
- Data flow: mount -> `ReadFile(fullPath)` -> `new Crepe({ root, defaultValue })` -> `crepe.create()` -> user edits -> debounce -> `crepe.getMarkdown()` -> `WriteFile(fullPath, content)`

### Technical Considerations
- **Lazy loading:** Dynamic `import()` for Crepe, identical to Monaco pattern:
  ```javascript
  const { Crepe } = await import('@milkdown/crepe');
  await import('@milkdown/crepe/theme/common/style.css');
  ```
- **Theme bridge:** `crepe-mashed.css` maps Crepe's CSS variables to mashed design tokens. Theme changes propagate automatically via CSS variable cascade -- no imperative theme API needed (unlike Monaco).
  - Key mappings: `--crepe-color-background` -> `var(--bg-deepest)`, `--crepe-color-on-surface` -> `var(--text-primary)`, `--crepe-color-primary` -> `var(--accent-green)`, `--crepe-font-code` -> `var(--font-mono)`
- **Auto-save (800ms debounce):** Match Monaco's pattern. On Crepe content change, debounce 800ms, then call `crepe.getMarkdown()` and `WriteFile`. Use a `setTimeout`/`clearTimeout` pattern.
- **Cmd+S save:** Register a DOM `keydown` listener for `Cmd+S` / `Ctrl+S` that triggers immediate save (bypass debounce). Prevent default browser behavior.
- **Editable toggle:** Use `crepe.setReadonly(!editable)` when `editable` prop changes.
- **File switching:** On `filePath` change, destroy the old Crepe instance (`crepe.destroy()`), read new file content, create new Crepe instance. This matches MonacoEditor's `destroyEditor()` pattern.
- **Save status indicator:** Show "Saving...", "Saved", or "Modified" status in a header bar, matching MonacoEditor's `saveStatus` pattern.
- **Content change detection:** Crepe fires ProseMirror transactions. Listen to the editor's state changes to detect modifications.

### Risks & Edge Cases
- **Crepe API stability:** Milkdown Crepe is actively maintained but the API may have minor differences from docs. Verify `crepe.getMarkdown()`, `crepe.setReadonly()`, and `crepe.destroy()` signatures after install.
- **Large markdown files:** Files over 100 KB may cause ProseMirror performance issues. Consider showing a warning or falling back to Monaco for very large files.
- **Frontmatter:** YAML frontmatter in markdown files (common in docs) should be preserved. Crepe/Milkdown may strip or mangle it. Test with frontmatter-heavy files.
- **Concurrent save + file switch:** If a debounced save is pending when the user switches files, it must complete with the OLD content before the new file loads. Flush pending save on file switch.
- **Module cache:** Like Monaco, the Crepe module import should be cached at module scope so subsequent `.md` file opens don't re-download the module.

### Reference Files
- `frontend/src/components/MonacoEditor.svelte` -- auto-save pattern (800ms debounce), save status, lazy loading, lifecycle management
- `frontend/src/lib/monacoTheme.js` -- theme integration pattern (for contrast: Crepe uses CSS, not this imperative approach)
- `docs/feasibility-multi-editor.md` -- Sections 3, 5.2, 5.3
- `DESIGN.md` -- mashed design tokens for the CSS variable bridge

## Acceptance Criteria

AC-1: WYSIWYG rendering
- Given a markdown file with headings, bold, links, code blocks, and lists
- When the MarkdownEditor loads the file
- Then all markdown elements render as styled WYSIWYG content
- And no raw markdown syntax is visible

AC-2: WYSIWYG editing
- Given the MarkdownEditor is showing a markdown file with `editable=true`
- When the user types text, adds headings, or creates lists
- Then the changes appear immediately as rendered WYSIWYG content
- And `crepe.getMarkdown()` returns valid markdown reflecting the edits

AC-3: Auto-save with debounce
- Given the user has made edits in the MarkdownEditor
- When 800ms passes without further edits
- Then `WriteFile` is called with the current markdown content
- And the save status shows "Saved"

AC-4: Cmd+S manual save
- Given the user has unsaved edits in the MarkdownEditor
- When the user presses Cmd+S (or Ctrl+S)
- Then `WriteFile` is called immediately (debounce bypassed)
- And the save status shows "Saved"

AC-5: Read-only mode
- Given the MarkdownEditor is mounted with `editable=false`
- When the user attempts to type or edit
- Then no changes are possible
- And the content displays as read-only rendered markdown

AC-6: Theme integration
- Given the mashed app is using a dark theme with `--bg-deepest: #07080a`
- When the MarkdownEditor is displayed
- Then the Crepe editor uses the mashed design tokens for colors, fonts, and spacing
- And theme changes propagate automatically without page reload

AC-7: File switching cleanup
- Given the MarkdownEditor is editing `README.md` with a pending auto-save
- When the user selects `CHANGELOG.md`
- Then the pending save for `README.md` completes first
- And the old Crepe instance is destroyed
- And a new Crepe instance loads `CHANGELOG.md` content

## BDD Test Scenarios

### Scenario 1: WYSIWYG rendering and editing
```gherkin
Feature: MarkdownEditor WYSIWYG

  Scenario: Render markdown as WYSIWYG
    Given a file "README.md" contains "# Hello\n\nSome **bold** text"
    When MarkdownEditor loads the file
    Then an h1 element with text "Hello" is rendered
    And a strong element with text "bold" is rendered
    And no raw "#" or "**" characters are visible

  Scenario: Edit in WYSIWYG mode
    Given MarkdownEditor is showing "README.md" with editable=true
    When the user types "World" after "Hello"
    Then the heading shows "HelloWorld"
    And crepe.getMarkdown() contains "# HelloWorld"
```

### Scenario 2: Auto-save behavior
```gherkin
Feature: MarkdownEditor auto-save

  Scenario: Debounced auto-save after 800ms
    Given the user types "new text" in the editor
    And 800ms passes without further edits
    When the debounce timer fires
    Then WriteFile is called with the full markdown content
    And the save status shows "Saved"

  Scenario: Debounce resets on continued typing
    Given the user types "a" in the editor
    And 400ms later types "b"
    When 800ms passes from the last keystroke
    Then WriteFile is called once (not twice)

  Scenario: Cmd+S bypasses debounce
    Given the user has unsaved edits
    When the user presses Cmd+S
    Then WriteFile is called immediately
    And any pending debounce timer is cancelled

  Scenario: Pending save flushes on file switch
    Given the user has unsaved edits to "README.md"
    When the filePath changes to "CHANGELOG.md"
    Then WriteFile is called for "README.md" before the switch
    And the Crepe instance is destroyed and recreated for the new file
```

### Scenario 3: Read-only and error handling
```gherkin
Feature: MarkdownEditor read-only and errors

  Scenario: Read-only mode
    Given MarkdownEditor is mounted with editable=false
    When the user tries to type
    Then the document content does not change

  Scenario: File read error
    Given ReadFile fails for the given path
    When MarkdownEditor attempts to load
    Then an error message is displayed with the file name
    And no Crepe editor is created
```

## Tasks / Subtasks

- [ ] Task 1: Install Milkdown Crepe (AC: 1)
  - [ ] Subtask 1a: Run `npm install @milkdown/crepe` in the frontend directory
  - [ ] Subtask 1b: Verify the package is added to `package.json`

- [ ] Task 2: Create crepe-mashed.css theme bridge (AC: 6)
  - [ ] Subtask 2a: Create `frontend/src/styles/crepe-mashed.css` with Crepe CSS variable overrides
  - [ ] Subtask 2b: Map `--crepe-color-*` vars to mashed `--bg-*`, `--text-*`, `--accent-*` tokens
  - [ ] Subtask 2c: Map `--crepe-font-*` vars to mashed `--font-mono` and `--font-ui` tokens
  - [ ] Subtask 2d: Add element-level styles for `.crepe .heading`, `.crepe .code-inline`, `.crepe .link`, `.crepe .blockquote`

- [ ] Task 3: Create MarkdownEditor.svelte (AC: 1, 2, 3, 4, 5, 7)
  - [ ] Subtask 3a: Implement component skeleton with props, loading/error/display states
  - [ ] Subtask 3b: Add lazy-loaded Crepe import with module-scope caching
  - [ ] Subtask 3c: Implement file loading via `ReadFile` and Crepe initialization (`new Crepe({ root, defaultValue })`)
  - [ ] Subtask 3d: Add auto-save with 800ms debounce using `crepe.getMarkdown()` -> `WriteFile`
  - [ ] Subtask 3e: Add Cmd+S handler with immediate save (bypass debounce)
  - [ ] Subtask 3f: Add reactive `editable` toggle via `crepe.setReadonly()`
  - [ ] Subtask 3g: Add file switch handling: flush pending save, `crepe.destroy()`, recreate for new file
  - [ ] Subtask 3h: Add save status indicator ("Saving...", "Saved", "Modified") in header bar

- [ ] Task 4: Wire into EditorRouter (AC: 1)
  - [ ] Subtask 4a: Replace markdown placeholder with real `MarkdownEditor` import in EditorRouter.svelte
  - [ ] Subtask 4b: Pass `filePath`, `repoPath`, and `editable` props to MarkdownEditor

- [ ] Task 5: Crepe API verification and edge case testing (AC: 1, 2, 7)
  - [ ] Subtask 5a: Verify `crepe.getMarkdown()`, `crepe.setReadonly()`, `crepe.destroy()` work as expected
  - [ ] Subtask 5b: Test with frontmatter-containing markdown files
  - [ ] Subtask 5c: Test file switching does not leak Crepe/ProseMirror instances

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `@milkdown/crepe` added to package.json
- [ ] WYSIWYG editing works for headings, bold, italic, links, code blocks, lists
- [ ] Auto-save fires after 800ms debounce
- [ ] Cmd+S triggers immediate save
- [ ] Theme integration uses CSS variable bridge (no imperative theme API)
- [ ] No memory leaks on file switching (Crepe destroyed, new instance created)
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
