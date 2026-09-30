# Story vsix-01: Read themes directly from .vsix zip archives

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Replace the directory-based theme scanner with one that reads `.vsix` files (which are zip archives) directly from the configured extensions directory. This eliminates the need to extract `.vsix` files to disk before scanning for themes. `ListVSCodiumThemes` will open each `.vsix` as a zip, parse `extension/package.json`, and return theme entries with a composite path encoding (`/path/to.vsix::vsix::internal/path.json`). `ReadThemeFile` will dispatch to a new zip-based reader when it detects the `::vsix::` separator.

## Developer Notes

### Architecture

All changes are in **`/Users/dev/Development/mashed/theme_scanner.go`** (single file). No new packages needed.

**New constants and helpers to add (top of file, after imports):**

```go
const vsixSeparator = "::vsix::"

func isVSIXThemePath(p string) bool
// Returns true if p contains vsixSeparator.

func parseVSIXThemePath(p string) (vsixPath, internalPath string, ok bool)
// Splits on vsixSeparator. Returns the .vsix file path and the zip-internal path.

func makeVSIXThemePath(vsixPath, internalPath string) string
// Returns vsixPath + vsixSeparator + internalPath.

func readFileFromZip(zr *zip.ReadCloser, name string) ([]byte, error)
// Iterates zr.File, finds matching name, enforces 512KB size limit
// (check f.UncompressedSize64), reads into []byte.

func mergeThemes(child, base *rawTheme)
// Extracts the inline merge logic from readThemeFileWithDepth into a reusable function.
// - Base colors go underneath child colors (child wins on conflict)
// - Base tokenColors prepended before child tokenColors
// - Inherit name/type from base if child's are empty
```

**Modify `ListVSCodiumThemes`:**

Current code iterates `os.ReadDir` looking for subdirectories. Change to:
- Iterate `os.ReadDir` looking for files with `.vsix` extension (case-insensitive)
- For each `.vsix` file, open with `zip.OpenReader(filepath.Join(extDir, entry.Name()))`
- Find `extension/package.json` inside the zip using `readFileFromZip`
- Parse as `packageJSON`
- For each contributed theme, build `ThemePath` using `makeVSIXThemePath(absVsixPath, "extension/"+t.Path)`
- Set `ExtensionID` to the vsix filename minus the `.vsix` extension (use `strings.TrimSuffix(entry.Name(), ".vsix")`)
- Continue to skip non-`.json` theme paths (existing H-7 filter)
- Close the zip reader after processing each file
- Keep existing alphabetical sort

**Modify `ReadThemeFile`:**

Add dispatch at the top, before the existing `readThemeFileWithDepth` call:
```go
if isVSIXThemePath(themePath) {
    vsixPath, internalPath, _ := parseVSIXThemePath(themePath)
    return a.readThemeFromVSIX(vsixPath, internalPath, extDir, 0)
}
```

**New method `readThemeFromVSIX(vsixPath, internalPath, extDir string, depth int) (string, error)`:**

- Enforce depth limit (>5 returns error), same as existing
- Security: `filepath.EvalSymlinks` on vsixPath, verify it is inside extDir (same prefix check pattern as `readThemeFileWithDepth`)
- Open zip with `zip.OpenReader(vsixPath)`
- Read file via `readFileFromZip(zr, internalPath)` -- this enforces the 512KB limit
- `stripJSONC` on the data
- Unmarshal into `rawTheme`
- If `theme.Include != ""`: resolve include path using `path.Join(path.Dir(internalPath), theme.Include)` -- use `"path"` package (not `"filepath"`) since these are zip-internal forward-slash paths
- Recursively call `readThemeFromVSIX` with the resolved include path and `depth+1`
- Call `mergeThemes(child, base)` with the resolved base
- Clear `theme.Include`, marshal to JSON, return

**New imports to add:** `"archive/zip"`, `"io"`, `"path"` (the `"path"` package, not `"path/filepath"` -- zip entries use forward slashes)

### Technical Considerations

- **Zip entry matching:** `.vsix` files use forward slashes in entry names regardless of OS. Use `"path"` package for internal path manipulation, NOT `"filepath"`.
- **Resource cleanup:** Always `defer zr.Close()` after `zip.OpenReader`. In `ListVSCodiumThemes`, close each zip reader before moving to the next file to avoid holding many file handles open.
- **Size limit enforcement:** Check `f.UncompressedSize64` in `readFileFromZip` before reading. Do NOT rely on reading and checking afterwards -- a malicious zip could have a huge decompressed size (zip bomb).
- **Error handling:** Follow existing pattern -- `continue` on per-file errors in `ListVSCodiumThemes` (skip bad vsix files), return wrapped errors in `readThemeFromVSIX`.
- **`readFileFromZip` should use `io.ReadAll`** with the `f.Open()` reader. The size check happens before reading via `UncompressedSize64`.
- **Backward compatibility:** The existing `readThemeFileWithDepth` and filesystem-based code path must remain intact. `ReadThemeFile` dispatches based on `isVSIXThemePath`, so non-vsix paths continue working unchanged.

### Risks & Edge Cases

- **Zip entry not found:** `readFileFromZip` should return a clear error like `"file %q not found in archive"` when the entry doesn't exist.
- **Corrupted .vsix:** `zip.OpenReader` will return an error; skip in `ListVSCodiumThemes`, return error in `readThemeFromVSIX`.
- **No `extension/package.json`:** Skip the vsix silently in `ListVSCodiumThemes` (same pattern as current `continue` on missing `package.json`).
- **Include cycle within zip:** The depth limit (5) prevents infinite recursion, same as filesystem includes.
- **Non-.vsix files in the directory:** The scanner should skip non-`.vsix` entries (currently it skips non-directories; flip this to skip non-`.vsix` files).
- **Case sensitivity:** `.vsix` extension check should be case-insensitive (`strings.EqualFold` or `strings.ToLower`).
- **Path separator in vsixSeparator:** The `::vsix::` separator is chosen to never appear in valid filesystem or zip paths. `parseVSIXThemePath` should use `strings.SplitN(p, vsixSeparator, 2)` to handle edge cases.

### Reference Files

- `/Users/dev/Development/mashed/theme_scanner.go` -- all modifications go here
- `/Users/dev/Development/mashed/theme_scanner_test.go` -- existing test patterns to follow
- `/Users/dev/Development/mashed/app.go` lines 43-49 -- `VSCodeThemeEntry` struct (no changes needed)

## Acceptance Criteria

AC-1: VSIX theme listing
- Given the configured extensions directory contains `.vsix` files with valid `extension/package.json` inside
- When `ListVSCodiumThemes()` is called
- Then it returns a sorted list of `VSCodeThemeEntry` with `ThemePath` in the format `"/path/to/file.vsix::vsix::extension/themes/dark.json"` and `ExtensionID` set to the vsix filename minus `.vsix`

AC-2: VSIX theme file reading
- Given a theme path in the format `"/path/to/file.vsix::vsix::extension/themes/dark.json"`
- When `ReadThemeFile(themePath)` is called
- Then it opens the `.vsix` as a zip, reads the internal file, strips JSONC comments, and returns valid clean JSON

AC-3: Include resolution within VSIX
- Given a theme inside a `.vsix` that has an `"include": "./base.json"` directive pointing to another file in the same zip
- When `ReadThemeFile` is called on that theme
- Then includes are resolved recursively within the zip archive using forward-slash paths
- And the merge follows existing rules: base colors underneath, base tokenColors prepended, inherit name/type from base if not set

AC-4: Security -- path traversal prevention
- Given a crafted VSIX theme path where the vsix file path resolves outside the extensions directory
- When `ReadThemeFile` is called
- Then it returns an error containing "outside extensions directory"

AC-5: Size limit enforcement
- Given a `.vsix` file containing a theme file larger than 512KB (uncompressed)
- When `ReadThemeFile` is called for that theme
- Then it returns an error containing "too large"

AC-6: Graceful handling of bad VSIX files
- Given the extensions directory contains corrupted `.vsix` files or `.vsix` files without `extension/package.json`
- When `ListVSCodiumThemes()` is called
- Then it skips the bad files and returns themes from valid `.vsix` files without erroring

AC-7: Include depth limit within VSIX
- Given a `.vsix` file containing a chain of includes exceeding depth 5
- When `ReadThemeFile` is called
- Then it stops resolving at depth 5 and returns what it has without crashing

AC-8: Backward compatibility
- Given a theme path that does NOT contain `::vsix::` (a regular filesystem path)
- When `ReadThemeFile(themePath)` is called
- Then it uses the existing `readThemeFileWithDepth` code path unchanged

## BDD Test Scenarios

### Scenario 1: List themes from VSIX files

```gherkin
Feature: VSIX theme scanning

  Scenario: Discover themes from .vsix files in extensions directory
    Given the extensions directory contains "dracula.vsix" with extension/package.json contributing theme "Dracula" at "extension/themes/dracula.json"
    And the extensions directory contains "github.vsix" with extension/package.json contributing themes "GitHub Dark" at "extension/themes/dark.json" and "GitHub Light" at "extension/themes/light.json"
    When ListVSCodiumThemes is called
    Then it returns 3 themes sorted alphabetically: "Dracula", "GitHub Dark", "GitHub Light"
    And each ThemePath has the format "{vsixAbsPath}::vsix::extension/themes/{file}.json"
    And ExtensionID for dracula.vsix is "dracula"
    And ExtensionID for github.vsix is "github"

  Scenario: Skip corrupted VSIX files gracefully
    Given the extensions directory contains "good.vsix" with a valid theme
    And the extensions directory contains "corrupt.vsix" which is not a valid zip
    When ListVSCodiumThemes is called
    Then it returns 1 theme from "good.vsix"
    And no error is returned

  Scenario: Skip VSIX files without package.json
    Given the extensions directory contains "empty.vsix" which is a valid zip but has no extension/package.json
    And the extensions directory contains "valid.vsix" with a valid theme
    When ListVSCodiumThemes is called
    Then it returns themes only from "valid.vsix"

  Scenario: Filter out non-JSON theme paths from VSIX
    Given a VSIX file contributes a theme with path "extension/themes/old.tmTheme"
    And the same VSIX also contributes a theme with path "extension/themes/dark.json"
    When ListVSCodiumThemes is called
    Then only the .json theme is returned

  Scenario: Non-.vsix files in directory are ignored
    Given the extensions directory contains "notes.txt" and "theme.vsix"
    When ListVSCodiumThemes is called
    Then only "theme.vsix" is scanned
```

### Scenario 2: Read theme file from VSIX

```gherkin
Feature: VSIX theme file reading

  Scenario: Read a clean JSON theme from a VSIX
    Given "test.vsix" contains "extension/themes/dark.json" with content {"name": "Dark", "colors": {"bg": "#000"}}
    When ReadThemeFile is called with "test.vsix::vsix::extension/themes/dark.json"
    Then it returns valid JSON with name "Dark" and colors.bg "#000"

  Scenario: Strip JSONC comments from VSIX theme file
    Given "test.vsix" contains "extension/themes/theme.json" with JSONC content including // and /* */ comments
    When ReadThemeFile is called with the VSIX theme path
    Then it returns valid JSON with comments stripped

  Scenario: Resolve single-level include within VSIX
    Given "test.vsix" contains "extension/themes/child.json" with include "./base.json"
    And "test.vsix" contains "extension/themes/base.json" with colors {bg: "#000", fg: "#fff"}
    And child.json overrides colors {bg: "#111"} and adds tokenColor for "keyword"
    When ReadThemeFile is called for child.json
    Then colors.bg is "#111" (child override)
    And colors.fg is "#fff" (inherited from base)
    And tokenColors has base entries prepended before child entries

  Scenario: Multi-level include within VSIX (grandparent -> parent -> child)
    Given a VSIX with grandparent.json -> parent.json -> child.json include chain
    And grandparent has colors {a: "1", b: "2", c: "3"}
    And parent includes grandparent and overrides {b: "22"}
    And child includes parent and overrides {c: "333"}
    When ReadThemeFile is called for child.json
    Then colors are {a: "1", b: "22", c: "333"}

  Scenario: Include depth limit exceeded within VSIX
    Given a VSIX with a chain of 7 includes (level0 -> level1 -> ... -> level6)
    When ReadThemeFile is called for level0
    Then it resolves up to depth 5 and stops
    And no error is returned to the caller
```

### Scenario 3: Security and limits

```gherkin
Feature: VSIX security controls

  Scenario: VSIX path outside extensions directory is rejected
    Given the extensions directory is "/tmp/ext"
    And a ReadThemeFile call references "/tmp/other/evil.vsix::vsix::extension/theme.json"
    When ReadThemeFile is called
    Then it returns an error containing "outside extensions directory"

  Scenario: Symlinked VSIX outside extensions directory is rejected
    Given a symlink inside extensions directory points to a VSIX outside the directory
    When ReadThemeFile is called with the symlink-based VSIX path
    Then it returns an error containing "outside extensions directory"

  Scenario: Theme file exceeding 512KB inside VSIX is rejected
    Given a VSIX contains a theme file with UncompressedSize64 > 512KB
    When ReadThemeFile is called for that theme
    Then it returns an error containing "too large"

  Scenario: File not found inside VSIX
    Given a VSIX file exists but does not contain the requested internal path
    When ReadThemeFile is called with "test.vsix::vsix::extension/themes/missing.json"
    Then it returns an error containing "not found"
```

### Scenario 4: Backward compatibility

```gherkin
Feature: Backward compatibility

  Scenario: Regular filesystem theme path still works
    Given a theme file exists at "/tmp/ext/test-ext/theme.json" on the filesystem
    And the extensions directory is configured to "/tmp/ext"
    When ReadThemeFile is called with "/tmp/ext/test-ext/theme.json" (no ::vsix:: separator)
    Then it uses the existing readThemeFileWithDepth code path
    And returns valid JSON
```

## Tasks / Subtasks

- [ ] Task 1: Add VSIX path helpers and `readFileFromZip` (AC: AC-1, AC-2, AC-5)
  - [ ] Add `"archive/zip"`, `"io"`, `"path"` imports to `theme_scanner.go`
  - [ ] Implement `vsixSeparator` constant, `isVSIXThemePath`, `parseVSIXThemePath`, `makeVSIXThemePath`
  - [ ] Implement `readFileFromZip` with 512KB `UncompressedSize64` check
  - [ ] Extract `mergeThemes(child, base *rawTheme)` from the inline merge code in `readThemeFileWithDepth`
  - [ ] Update `readThemeFileWithDepth` to call `mergeThemes` instead of inline logic

- [ ] Task 2: Implement `readThemeFromVSIX` method (AC: AC-2, AC-3, AC-4, AC-5, AC-7)
  - [ ] Implement `readThemeFromVSIX(vsixPath, internalPath, extDir string, depth int) (string, error)`
  - [ ] Add security check: EvalSymlinks on vsixPath, prefix check against extDir
  - [ ] Add zip reading: open zip, read file, stripJSONC, unmarshal
  - [ ] Add include resolution using `path.Join(path.Dir(internalPath), theme.Include)` (forward-slash paths)
  - [ ] Merge with `mergeThemes`, clear Include field, marshal to JSON

- [ ] Task 3: Modify `ListVSCodiumThemes` to scan VSIX files (AC: AC-1, AC-6)
  - [ ] Change loop to filter for `.vsix` file extensions instead of directories
  - [ ] Open each vsix with `zip.OpenReader`, find `extension/package.json` via `readFileFromZip`
  - [ ] Parse `packageJSON`, build theme entries with `makeVSIXThemePath`
  - [ ] Set `ExtensionID` to filename minus `.vsix` extension
  - [ ] Maintain `continue` on error pattern for graceful skipping
  - [ ] Close zip reader after each file (not deferred in loop)

- [ ] Task 4: Modify `ReadThemeFile` dispatch (AC: AC-2, AC-8)
  - [ ] Add `isVSIXThemePath` check at top of `ReadThemeFile`
  - [ ] Dispatch to `readThemeFromVSIX` for vsix paths
  - [ ] Preserve existing `readThemeFileWithDepth` call for non-vsix paths

- [ ] Task 5: Write tests (AC: AC-1 through AC-8)
  - [ ] Add helper `createMockVSIX(t, dir, name string, files map[string][]byte)` to create test `.vsix` zip files
  - [ ] Test `ListVSCodiumThemes` with `.vsix` files (happy path, corrupt, missing package.json, .tmTheme filter)
  - [ ] Test `ReadThemeFile` with vsix paths (happy path, JSONC stripping, includes, multi-level includes)
  - [ ] Test security (path traversal, symlink, size limit, file not found in zip)
  - [ ] Test backward compatibility (non-vsix path still works)
  - [ ] Test include depth limit within VSIX

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified code in `theme_scanner.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
