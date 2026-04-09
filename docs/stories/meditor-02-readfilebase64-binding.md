# Story 2: ReadFileBase64 Wails Binding for Image Data

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** none
**Status:** ready

## Description

Add a `ReadFileBase64` Go method on the App struct that reads a binary file and returns its content as a base64-encoded data URI string. This binding is required by the ImageViewer component to display images from the filesystem through the Wails webview, which cannot access local file paths directly.

## Developer Notes

### Architecture
- **Modified file:** `app_git.go` -- add `ReadFileBase64` adjacent to existing `ReadFile` (line 660)
- The method lives alongside the existing file I/O methods: `ReadFile`, `WriteFile`, `ReadFileAtHead`, `ReadFileDiff`
- Returns a complete data URI: `data:image/png;base64,<encoded>` so the frontend can use it directly as an `<img src>` value
- MIME type detection based on file extension (not content sniffing)

### Technical Considerations
- **File size cap:** Images can be large. Cap at 10 MB (vs 1 MB for ReadFile text). Return error for oversized files.
- **MIME mapping:** Map extensions to MIME types: `png->image/png`, `jpg/jpeg->image/jpeg`, `gif->image/gif`, `svg->image/svg+xml`, `webp->image/webp`, `bmp->image/bmp`, `ico->image/x-icon`
- **Error handling:** Use `fmt.Errorf` with `%w` wrapping, matching existing patterns in `app_git.go`
- **Binary safety:** `encoding/base64.StdEncoding.EncodeToString(data)` handles any binary content
- **SVG special case:** SVGs are text-based XML. `ReadFileBase64` should still work for them (base64-encoded SVG in a data URI is valid), but the EditorRouter could also use `ReadFile` for SVG source view later.

### Risks & Edge Cases
- Symlinks: `os.ReadFile` follows symlinks, which is the desired behavior
- Non-image files: If called with a `.go` file, it will still base64-encode it -- the frontend is responsible for only calling this for image files
- Empty files: Return `data:image/png;base64,` (empty base64) -- frontend should handle gracefully
- Unknown extensions: Default to `application/octet-stream`

### Reference Files
- `app_git.go:660-670` -- existing `ReadFile` method (pattern to follow)
- `app_git.go:652-657` -- existing `WriteFile` method
- Wails auto-generates bindings in `frontend/wailsjs/go/main/App.js` after `wails dev` or `wails generate module`

## Acceptance Criteria

AC-1: Base64 encoding
- Given a PNG image file exists at `/path/to/image.png`
- When `ReadFileBase64("/path/to/image.png")` is called
- Then it returns a string starting with `data:image/png;base64,`
- And the remainder is valid base64 encoding of the file contents

AC-2: MIME type detection
- Given image files with various extensions (.png, .jpg, .jpeg, .gif, .svg, .webp)
- When `ReadFileBase64` is called for each
- Then the returned data URI contains the correct MIME type for each extension

AC-3: File size limit
- Given an image file larger than 10 MB
- When `ReadFileBase64` is called
- Then it returns an error containing "file too large"
- And no base64 data is returned

AC-4: Error handling for missing files
- Given a file path that does not exist
- When `ReadFileBase64` is called
- Then it returns an error wrapping the underlying os error
- And the error message includes the file path

AC-5: Empty path handling
- Given an empty string as the file path
- When `ReadFileBase64` is called
- Then it returns an error "empty file path"

## BDD Test Scenarios

### Scenario 1: Successful image reading
```gherkin
Feature: ReadFileBase64 binary file encoding

  Scenario: Read a PNG file
    Given a file "test.png" exists in a temp directory with known binary content
    When ReadFileBase64 is called with the full path
    Then the result starts with "data:image/png;base64,"
    And decoding the base64 portion yields the original file content

  Scenario: Read a JPEG file
    Given a file "photo.jpg" exists in a temp directory
    When ReadFileBase64 is called with the full path
    Then the result starts with "data:image/jpeg;base64,"

  Scenario: Read an SVG file
    Given a file "icon.svg" exists in a temp directory with XML content
    When ReadFileBase64 is called with the full path
    Then the result starts with "data:image/svg+xml;base64,"
```

### Scenario 2: Error cases
```gherkin
Feature: ReadFileBase64 error handling

  Scenario: File does not exist
    Given no file exists at "/tmp/nonexistent.png"
    When ReadFileBase64 is called with "/tmp/nonexistent.png"
    Then an error is returned containing "nonexistent.png"

  Scenario: File exceeds size limit
    Given a file "huge.png" exists with size 11 MB
    When ReadFileBase64 is called
    Then an error is returned containing "file too large"

  Scenario: Empty path
    Given an empty string as the path argument
    When ReadFileBase64 is called with ""
    Then an error is returned containing "empty file path"

  Scenario: Unknown extension
    Given a file "data.raw" exists in a temp directory
    When ReadFileBase64 is called with the full path
    Then the result starts with "data:application/octet-stream;base64,"
```

## Tasks / Subtasks

- [ ] Task 1: Implement ReadFileBase64 method (AC: 1, 2, 3, 4, 5)
  - [ ] Subtask 1a: Add `mimeForExt(ext string) string` helper function with extension-to-MIME mapping
  - [ ] Subtask 1b: Implement `func (a *App) ReadFileBase64(path string) (string, error)` in `app_git.go`
  - [ ] Subtask 1c: Add empty path guard, file size check (10 MB cap), and proper error wrapping

- [ ] Task 2: Write Go tests (AC: 1, 2, 3, 4, 5)
  - [ ] Subtask 2a: Table-driven test for MIME type mapping across all supported extensions
  - [ ] Subtask 2b: Test successful base64 encoding with known content (verify round-trip)
  - [ ] Subtask 2c: Test error cases: missing file, oversized file, empty path
  - [ ] Subtask 2d: Test unknown extension defaults to `application/octet-stream`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
