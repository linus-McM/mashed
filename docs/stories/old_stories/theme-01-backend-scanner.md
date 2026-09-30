# Story 1: Go Backend -- Extension Scanner & Theme Reader

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Add Go backend functions to scan VSCodium extension directories for installed color themes and read their JSON files securely. This is the foundation for all theme import features -- the frontend cannot discover or load any external theme without these Wails-bound methods. The implementation incorporates all critical and high-priority security fixes from the adversarial review (path traversal, size check, JSONC stripping, mutex protection, tilde expansion, tmTheme filtering, and include resolution).

## Developer Notes

### Architecture

This story modifies `app.go` exclusively. No new packages are needed.

**New types to add (top of file, near `conductorConfig`):**

```go
type VSCodeThemeEntry struct {
    Label       string `json:"label"`
    ExtensionID string `json:"extensionId"`
    ThemePath   string `json:"themePath"`
    UITheme     string `json:"uiTheme"`
}
```

**Modified struct:**

```go
type conductorConfig struct {
    DevDir          string `json:"devDir"`
    Theme           string `json:"theme,omitempty"`
    VSCodiumExtPath string `json:"vscodiumExtPath,omitempty"`
    ImportedTheme   string `json:"importedTheme,omitempty"` // NEW
}
```

**New Wails-bound methods to add:**

1. `ListVSCodiumThemes() ([]VSCodeThemeEntry, error)` -- scans extension directory
2. `ReadThemeFile(themePath string) (string, error)` -- reads + JSONC-strips + resolves includes
3. `SetImportedTheme(themePath string) error` -- persists selected imported theme path

**Data flow:**
Frontend calls `ListVSCodiumThemes()` --> Go reads `package.json` files in extension dirs --> returns `[]VSCodeThemeEntry` to frontend. Frontend later calls `ReadThemeFile(path)` --> Go validates path security, checks size, reads file, strips JSONC comments, resolves `include` directives, returns clean JSON string.

### Technical Considerations

**Security (C-1 fix):** Use `filepath.EvalSymlinks` instead of `filepath.Abs` to resolve symlinks before the prefix check. Append `string(os.PathSeparator)` to both paths in the `strings.HasPrefix` check to prevent `/ext-other/` matching `/ext/`. Never discard errors from `os.UserHomeDir()` or `filepath.EvalSymlinks()`.

```go
absTheme, err := filepath.EvalSymlinks(themePath)
if err != nil {
    return "", fmt.Errorf("resolving theme path: %w", err)
}
absExt, err := filepath.EvalSymlinks(extDir)
if err != nil {
    return "", fmt.Errorf("resolving extensions path: %w", err)
}
if !strings.HasPrefix(absTheme+string(os.PathSeparator), absExt+string(os.PathSeparator)) {
    return "", fmt.Errorf("theme path outside extensions directory")
}
```

**Size check before read (C-4 fix):** Call `os.Stat` to check file size BEFORE `os.ReadFile` to prevent memory exhaustion.

```go
info, err := os.Stat(absTheme)
if err != nil {
    return "", fmt.Errorf("stat theme file: %w", err)
}
if info.Size() > 512*1024 {
    return "", fmt.Errorf("theme file too large (%d bytes, max 512KB)", info.Size())
}
```

**JSONC stripping (C-3 fix):** Strip `//` line comments and `/* */` block comments from the file contents in Go before returning the string. Use a simple state-machine or regex approach. Do NOT strip inside string literals.

A robust approach:
```go
func stripJSONC(data []byte) []byte {
    // Remove single-line comments: // to end of line (not inside strings)
    // Remove block comments: /* ... */ (not inside strings)
    // Remove trailing commas before } or ]
}
```

**Tilde expansion (H-3 fix):** Match `~/` prefix or exact `~`, not arbitrary `~user` paths. Always check `os.UserHomeDir()` error.

```go
if strings.HasPrefix(extDir, "~/") || extDir == "~" {
    home, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("resolving home directory: %w", err)
    }
    if extDir == "~" {
        extDir = home
    } else {
        extDir = filepath.Join(home, extDir[2:])
    }
}
```

**tmTheme filter (H-7 fix):** Skip theme entries where the path does not end in `.json` (case-insensitive).

```go
if !strings.HasSuffix(strings.ToLower(t.Path), ".json") {
    continue
}
```

**Include resolution (H-1 fix):** Merge `ReadThemeFileResolved` logic directly into `ReadThemeFile`. After reading and JSONC-stripping, check for an `"include"` field. If present, recursively read the included file (with a depth limit of 5) and merge base colors underneath the current theme's colors. Prepend base `tokenColors` before current ones.

```go
type rawTheme struct {
    Include     string                   `json:"include"`
    Name        string                   `json:"name"`
    Type        string                   `json:"type"`
    Colors      map[string]string        `json:"colors"`
    TokenColors []json.RawMessage        `json:"tokenColors"`
}
```

**Config mutex (H-2 fix):** All config load-modify-save operations (`SetTheme`, `SetVSCodiumExtPath`, `SetImportedTheme`, `SetDevDir`) must hold `a.mu` to prevent concurrent Wails calls from clobbering each other. The existing `a.mu` field on `App` is already available.

```go
func (a *App) SetImportedTheme(themePath string) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    cfg := loadConfig()
    cfg.ImportedTheme = themePath
    return saveConfig(cfg)
}
```

Apply the same pattern to `SetTheme`, `SetVSCodiumExtPath`, and `SetDevDir`.

### Risks & Edge Cases

- **Symlink loops:** `filepath.EvalSymlinks` will return an error for circular symlinks -- handle gracefully.
- **Missing extensions directory:** Return an empty slice (not an error) if the path is not configured; return an error with a clear message if configured but not accessible.
- **Corrupt package.json:** `json.Unmarshal` failure should skip the extension silently (already handled with `continue`).
- **Very large extension directories (L-4):** A user with 200+ extensions will trigger 200+ `ReadFile` calls for `package.json` files. This is acceptable for v1 but should be noted as a future caching opportunity.
- **JSONC edge cases:** Comments inside JSON string values should NOT be stripped. A naive regex will break themes with URLs containing `//`. The state-machine approach handles this correctly.
- **Include depth:** Cap at 5 levels to prevent infinite recursion. Log a warning if the cap is hit.

### Reference Files

- `/Users/dev/Development/mashed/app.go` -- all modifications go here. Study the existing Wails-bound methods pattern (lines 498-1250) and the `loadConfig`/`saveConfig` pattern (lines 56-78).
- `/Users/dev/Development/mashed/app.go:317` -- existing `a.mu.Lock()` usage pattern for the notification list.
- `/Users/dev/Development/mashed/app.go:44-78` -- existing `conductorConfig` struct and config helpers.

## Acceptance Criteria

AC-1: Extension directory scanning
- Given a configured `vscodiumExtPath` pointing to a directory with VSCodium extensions
- When `ListVSCodiumThemes()` is called
- Then it returns a sorted list of `VSCodeThemeEntry` with correct `Label`, `ExtensionID`, `ThemePath`, and `UITheme` fields
- And `.tmTheme` entries are excluded from results

AC-2: Secure theme file reading
- Given a valid theme path within the configured extensions directory
- When `ReadThemeFile(themePath)` is called
- Then it returns the file contents as a clean JSON string with JSONC comments stripped
- And symlinks are resolved before the security check
- And the path prefix check uses `os.PathSeparator` suffix

AC-3: Path traversal prevention
- Given a theme path containing `../` sequences, symlinks to external files, or a path outside the extensions directory
- When `ReadThemeFile(themePath)` is called
- Then it returns an error "theme path outside extensions directory"
- And no file content is leaked

AC-4: Size limit enforcement
- Given a theme file larger than 512KB
- When `ReadThemeFile(themePath)` is called
- Then it returns an error without reading the full file into memory
- And `os.Stat` is used before `os.ReadFile`

AC-5: Include resolution
- Given a theme file with `"include": "./base-theme.json"`
- When `ReadThemeFile(themePath)` is called
- Then the returned JSON contains merged colors (base colors overridden by current) and concatenated tokenColors (base prepended before current)
- And recursive includes are capped at depth 5

AC-6: Config persistence with mutex
- Given concurrent calls to `SetTheme`, `SetVSCodiumExtPath`, and `SetImportedTheme`
- When all three complete
- Then the config file contains all three updated fields (no clobbered writes)
- And the `a.mu` mutex is held during each load-modify-save cycle

AC-7: Tilde expansion
- Given `vscodiumExtPath` is set to `~/.vscode-oss/extensions`
- When `ListVSCodiumThemes()` is called
- Then `~` is correctly expanded to the user's home directory
- And `os.UserHomeDir()` errors are propagated (not silently discarded)

## BDD Test Scenarios

### Scenario 1: Happy path -- scan and read themes

```gherkin
Feature: VSCodium theme scanning

  Scenario: Scan extensions directory with multiple themes
    Given a temporary extensions directory with two extensions:
      | directory                           | theme_label | theme_path         | uiTheme  |
      | dracula-theme.theme-dracula-2.24.3  | Dracula     | ./theme/dracula.json | vs-dark |
      | github.github-vscode-theme-6.3.5    | GitHub Dark | ./themes/dark.json   | vs-dark |
    And each extension has a valid package.json with contributes.themes
    And each theme JSON file exists
    When ListVSCodiumThemes is called
    Then the result contains 2 entries sorted alphabetically by label
    And the first entry has Label "Dracula" and UITheme "vs-dark"

  Scenario: Read a theme file with JSONC comments
    Given a theme file containing:
      """
      {
        // This is a comment
        "name": "Test Theme",
        "colors": {
          "editor.background": "#282a36" /* inline comment */
        }
      }
      """
    When ReadThemeFile is called with the file's path
    Then the returned string parses as valid JSON
    And the parsed JSON has name "Test Theme"

  Scenario: Read a theme file with include directive
    Given a base theme file "base.json" with colors {"editor.background": "#000", "editor.foreground": "#fff"}
    And a child theme file "child.json" with include "./base.json" and colors {"editor.background": "#111"}
    When ReadThemeFile is called with the child theme's path
    Then the returned JSON has editor.background "#111" (child overrides)
    And the returned JSON has editor.foreground "#fff" (inherited from base)
```

### Scenario 2: Security -- path traversal

```gherkin
Feature: Theme file security

  Scenario: Reject path traversal via relative segments
    Given extensions directory is "/tmp/test-ext"
    When ReadThemeFile is called with "/tmp/test-ext/../../etc/passwd"
    Then an error is returned containing "outside extensions directory"

  Scenario: Reject path traversal via symlink
    Given extensions directory is "/tmp/test-ext"
    And "/tmp/test-ext/evil/theme.json" is a symlink to "/etc/passwd"
    When ReadThemeFile is called with "/tmp/test-ext/evil/theme.json"
    Then an error is returned containing "outside extensions directory"

  Scenario: Reject oversized files before reading
    Given a theme file of size 1MB exists at a valid path
    When ReadThemeFile is called
    Then an error is returned containing "too large"
    And memory allocation does not exceed 512KB for the file
```

### Scenario 3: Edge cases

```gherkin
Feature: Scanner edge cases

  Scenario: Skip tmTheme files
    Given an extension with a theme path ending in ".tmTheme"
    When ListVSCodiumThemes is called
    Then the tmTheme entry is not included in results

  Scenario: Handle unconfigured extensions path
    Given vscodiumExtPath is empty in config
    When ListVSCodiumThemes is called
    Then an error is returned containing "not configured"

  Scenario: Mutex prevents config corruption
    Given two goroutines calling SetTheme("dark") and SetImportedTheme("/path/theme.json") simultaneously
    When both calls complete
    Then the final config file contains both Theme "dark" and ImportedTheme "/path/theme.json"

  Scenario: Tilde expansion handles edge cases
    Given vscodiumExtPath is "~/.vscode-oss/extensions"
    When ListVSCodiumThemes resolves the path
    Then the path starts with the user's home directory
    And does not contain "~"
```

## Tasks / Subtasks

- [ ] Task 1: Add VSCodeThemeEntry type and update conductorConfig (AC: AC-1, AC-6)
  - [ ] Add `VSCodeThemeEntry` struct near existing types in `app.go`
  - [ ] Add `ImportedTheme string` field to `conductorConfig`
  - [ ] Add `SetImportedTheme` method with mutex protection
  - [ ] Add mutex to existing `SetTheme`, `SetVSCodiumExtPath`, `SetDevDir` methods

- [ ] Task 2: Implement ListVSCodiumThemes (AC: AC-1, AC-7)
  - [ ] Implement tilde expansion with proper error handling (`~/` or `~` only)
  - [ ] Scan extension directories and parse `package.json` for `contributes.themes`
  - [ ] Filter out non-`.json` theme paths (H-7 fix)
  - [ ] Sort results alphabetically by label
  - [ ] Return empty slice (not error) for unconfigured path vs error for inaccessible path

- [ ] Task 3: Implement JSONC stripping (AC: AC-2)
  - [ ] Write `stripJSONC(data []byte) []byte` function
  - [ ] Handle `//` line comments (not inside string literals)
  - [ ] Handle `/* */` block comments (not inside string literals)
  - [ ] Handle trailing commas before `}` or `]`
  - [ ] Write unit tests for JSONC stripping edge cases

- [ ] Task 4: Implement ReadThemeFile with security (AC: AC-2, AC-3, AC-4, AC-5)
  - [ ] Resolve symlinks with `filepath.EvalSymlinks` before path check
  - [ ] Add `os.PathSeparator` suffix to prefix check
  - [ ] Check file size with `os.Stat` before reading
  - [ ] Read file and strip JSONC comments
  - [ ] Parse for `include` directive and merge base theme (depth-limited to 5)
  - [ ] Re-serialize merged theme as JSON and return

- [ ] Task 5: Write tests (AC: all)
  - [ ] Unit tests for `stripJSONC` (comments, trailing commas, strings with //)
  - [ ] Unit tests for path security (traversal, symlinks, separator suffix)
  - [ ] Unit tests for size check ordering
  - [ ] Unit tests for include resolution and depth limit
  - [ ] Unit tests for tilde expansion
  - [ ] Integration test: scan a temp directory with mock extensions

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on new/modified code in `app.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] Code review: no CRITICAL/HIGH issues
