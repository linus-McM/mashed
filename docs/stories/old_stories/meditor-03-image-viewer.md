# Story 3: ImageViewer Component with Panzoom

**Priority:** P0-critical
**Domain:** frontend
**Estimated Complexity:** M
**Depends On:** Story 1 (meditor-01), Story 2 (meditor-02)
**Status:** done

## Description

Create an `ImageViewer.svelte` component that displays image files with zoom and pan capabilities using the `panzoom` library (~3 kB). The component loads image data via the `ReadFileBase64` Wails binding and renders it in a full-height container. Replace the placeholder in EditorRouter with the real ImageViewer.

## Developer Notes

### Architecture
- **New file:** `frontend/src/components/ImageViewer.svelte`
- **Modified file:** `frontend/src/components/EditorRouter.svelte` -- replace image placeholder with real ImageViewer import
- **New dependency:** `panzoom` (~3 kB gzip) -- `npm install panzoom`
- Props: `filePath` (string), `repoPath` (string)
- Data flow: `filePath` changes -> call `ReadFileBase64(fullPath)` -> set `<img src={dataUri}>` -> attach panzoom

### Technical Considerations
- **Panzoom as Svelte action:** Use the `use:zoomable` pattern from the feasibility doc. The action attaches panzoom to the `<img>` element and returns a `destroy` callback for cleanup.
- **Full path construction:** `fullPath = repoPath + '/' + filePath` (matching MonacoEditor pattern)
- **Reactive file loading:** Use `$: loadImage(filePath, repoPath)` to reload when the file changes
- **Loading state:** Show a loading spinner or skeleton while `ReadFileBase64` is in flight
- **Error state:** If `ReadFileBase64` fails, show error message with file path
- **Fit-to-container:** Initial zoom should fit the image within the container. Panzoom's `initialZoom` or manual calculation based on container vs image dimensions.
- **Theme integration:** Container background should use `var(--bg-deepest)` for contrast. No other theming needed (image content is the theme).

### Risks & Edge Cases
- **Large images:** Base64 encoding doubles the size. A 5 MB image becomes a ~6.7 MB data URI string in the DOM. The 10 MB backend cap should prevent issues.
- **SVG rendering:** SVGs render as images (not editable source). This is intentional for Phase 1; SVG dual-mode is Phase 3.
- **Broken images:** If the file is deleted while viewing, show a "file not found" state
- **Animated GIFs:** Will animate naturally in `<img>` tag -- no special handling needed
- **Component destroy:** Panzoom must be disposed on unmount to prevent memory leaks

### Reference Files
- `frontend/src/components/MonacoEditor.svelte` -- pattern for file loading, error states, Wails binding calls
- `docs/feasibility-multi-editor.md` -- Section 4 (panzoom recommendation and code pattern)
- `frontend/src/components/EditorRouter.svelte` -- where ImageViewer gets mounted

## Acceptance Criteria

AC-1: Image display
- Given a PNG file exists in the repo
- When the user selects it in the FileTree
- Then the ImageViewer displays the image using data from `ReadFileBase64`
- And the image is visible within the editor pane area

AC-2: Zoom and pan
- Given an image is displayed in the ImageViewer
- When the user scrolls the mouse wheel over the image
- Then the image zooms in or out
- And when the user clicks and drags
- Then the image pans in the drag direction

AC-3: File switching
- Given the ImageViewer is displaying `logo.png`
- When the user selects `banner.jpg` from the FileTree
- Then the previous panzoom instance is destroyed
- And the new image loads and displays correctly

AC-4: Loading state
- Given a large image file is selected
- When `ReadFileBase64` is in flight
- Then the ImageViewer shows a loading indicator
- And the loading indicator disappears when the image loads

AC-5: Error handling
- Given a file path that does not exist or exceeds the size limit
- When the ImageViewer attempts to load it
- Then an error message is displayed in the viewer area
- And the error includes the file name

AC-6: Supported formats
- Given image files of types PNG, JPEG, GIF, SVG, and WebP
- When each is selected in the FileTree
- Then each displays correctly in the ImageViewer

## BDD Test Scenarios

### Scenario 1: Image loading and display
```gherkin
Feature: ImageViewer displays images from filesystem

  Scenario: Display a PNG image
    Given a PNG file "assets/logo.png" exists in the repo
    And EditorRouter routes the file to ImageViewer
    When ImageViewer mounts
    Then it calls ReadFileBase64 with the full path
    And renders an <img> element with the returned data URI as src

  Scenario: Display a JPEG image
    Given a JPEG file "photos/team.jpg" exists in the repo
    When ImageViewer loads the file
    Then the <img> src starts with "data:image/jpeg;base64,"

  Scenario: Display an SVG image
    Given an SVG file "icons/arrow.svg" exists in the repo
    When ImageViewer loads the file
    Then the <img> src starts with "data:image/svg+xml;base64,"
    And the SVG renders as an image (not editable source)
```

### Scenario 2: Zoom and pan interaction
```gherkin
Feature: ImageViewer zoom and pan

  Scenario: Zoom in with scroll wheel
    Given an image is displayed in the ImageViewer
    When the user scrolls the mouse wheel up
    Then the panzoom instance zooms in
    And the image appears larger

  Scenario: Pan by dragging
    Given an image is zoomed in beyond the container
    When the user clicks and drags on the image
    Then the visible area of the image shifts in the drag direction
```

### Scenario 3: Error and edge cases
```gherkin
Feature: ImageViewer error handling

  Scenario: File not found
    Given a filePath points to a non-existent image
    When ImageViewer attempts to load it
    Then an error message is shown containing the file name
    And no broken image icon is displayed

  Scenario: File switch cleanup
    Given ImageViewer is displaying "a.png" with an active panzoom instance
    When the filePath changes to "b.png"
    Then the panzoom instance for "a.png" is disposed
    And a new panzoom instance is created for "b.png"

  Scenario: Loading indicator
    Given ImageViewer is mounted with a new filePath
    When ReadFileBase64 has not yet returned
    Then a loading indicator is visible
    When ReadFileBase64 returns successfully
    Then the loading indicator is hidden and the image is shown
```

## Tasks / Subtasks

- [ ] Task 1: Install panzoom dependency (AC: 2)
  - [ ] Subtask 1a: Run `npm install panzoom` in the frontend directory
  - [ ] Subtask 1b: Verify the package is added to `package.json`

- [ ] Task 2: Create ImageViewer.svelte (AC: 1, 2, 4, 5, 6)
  - [ ] Subtask 2a: Implement component skeleton with props (`filePath`, `repoPath`), loading/error/display states
  - [ ] Subtask 2b: Add reactive image loading: `$: loadImage(filePath, repoPath)` calling `ReadFileBase64`
  - [ ] Subtask 2c: Implement `zoomable` Svelte action wrapping panzoom with `destroy` cleanup
  - [ ] Subtask 2d: Style container with `var(--bg-deepest)` background, centered image, `object-contain` fit
  - [ ] Subtask 2e: Add filename display in a header bar (matching editor header pattern)

- [ ] Task 3: Wire into EditorRouter (AC: 1, 3, 6)
  - [ ] Subtask 3a: Replace image placeholder import with real `ImageViewer` import in EditorRouter.svelte
  - [ ] Subtask 3b: Pass `filePath` and `repoPath` props to ImageViewer

- [ ] Task 4: Lifecycle and cleanup tests (AC: 3)
  - [ ] Subtask 4a: Verify panzoom dispose is called on component destroy
  - [ ] Subtask 4b: Verify panzoom dispose is called on filePath change before new instance

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `panzoom` added to package.json
- [ ] Image files render correctly for PNG, JPEG, GIF, SVG, WebP
- [ ] Zoom and pan work smoothly
- [ ] No memory leaks on file switching (panzoom disposed)
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
