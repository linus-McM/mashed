# Story 1: Advice Mode Discovery & Parsing Backend

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Build the `internal/advice` package that discovers, parses, and merges advice mode `.md` files from three locations: bundled defaults (via `go:embed`), global user directory (`~/.mashed/advice/`), and local repo directory (`{repoPath}/.claude/advice/`). This is the foundation for the methodology advice system -- every downstream story depends on these types and the loading logic.

## Developer Notes

### Architecture

- **New package:** `internal/advice/`
- **New files:**
  - `internal/advice/types.go` -- `AdviceMode` struct
  - `internal/advice/loader.go` -- discovery, parsing, merging logic
  - `internal/advice/loader_test.go` -- tests
  - `internal/advice/defaults/` -- directory for bundled `.md` files (story review-05 populates this)

- **Types to create:**

```go
// AdviceMode represents a methodology advice perspective.
type AdviceMode struct {
    Name        string `json:"name" yaml:"name"`               // unique key, e.g. "clean-code"
    DisplayName string `json:"displayName" yaml:"displayName"` // human label, e.g. "Clean Code"
    Icon        string `json:"icon" yaml:"icon"`               // emoji or lucide icon name
    Order       int    `json:"order" yaml:"order"`             // sort order in dropdown
    Body        string `json:"-" yaml:"-"`                     // markdown body (not sent in list)
    Source      string `json:"source"`                         // "bundled" | "global" | "local"
    FilePath    string `json:"-"`                              // absolute path to the .md file
}
```

- **Functions to create in `loader.go`:**

```go
//go:embed defaults/*.md
var defaultAdviceFS embed.FS

// LoadAdviceModes discovers and merges advice files from all three sources.
// Priority: bundled < global < local (same name = replace).
func LoadAdviceModes(repoPath string) ([]AdviceMode, error)

// LoadAdviceBody reads the full markdown body for a given mode name.
func LoadAdviceBody(repoPath, modeName string) (string, error)

// parseAdviceFile parses a single .md file with YAML frontmatter.
func parseAdviceFile(path string, source string) (AdviceMode, error)

// parseAdviceFileFromFS parses from an embed.FS (for bundled defaults).
func parseAdviceFileFromFS(fs embed.FS, path string) (AdviceMode, error)

// globalAdviceDir returns ~/.mashed/advice/
func globalAdviceDir() (string, error)

// localAdviceDir returns {repoPath}/.claude/advice/
func localAdviceDir(repoPath string) string
```

### Data Flow

1. `LoadAdviceModes(repoPath)` scans all three directories
2. Bundled defaults are read from `defaultAdviceFS` via `go:embed`
3. Global overrides from `~/.mashed/advice/*.md`
4. Local additions from `{repoPath}/.claude/advice/*.md`
5. Merge by `Name` field: later sources replace earlier ones
6. Sort by `Order` field, then alphabetically by `DisplayName`
7. Return the merged, sorted slice

### Technical Considerations

- **YAML frontmatter parsing:** Use `gopkg.in/yaml.v3` (already in go.mod as indirect dep). Frontmatter is delimited by `---` lines. Split on first two `---` occurrences, YAML-parse the middle, rest is Body.
- **go:embed:** The `defaults/` directory will initially be empty (or contain a `.gitkeep`). Story review-05 populates it. The embed directive should use `defaults/*.md` pattern. If no files match at compile time, use `all:defaults` to avoid build errors on empty dir.
- **Error handling:** Non-fatal for individual file parse errors -- log and skip, return successfully parsed modes. Fatal only if all three directories fail to read.
- **Concurrency:** No goroutines needed -- this is a synchronous read-and-merge operation.
- **Home directory:** Use `os.UserHomeDir()` for the global path.

### Risks & Edge Cases

- Empty `defaults/` directory before story review-05 is done -- must not crash on zero bundled files
- Malformed YAML frontmatter -- skip file, log warning
- Missing `name` field in frontmatter -- skip file (name is the merge key)
- Duplicate `name` across global and local -- local wins (expected behavior)
- Non-`.md` files in advice directories -- ignore
- Directories that don't exist -- skip silently (common for global/local)

### Reference Files

- `internal/explain/explain.go` -- similar small standalone package pattern
- `internal/bmad/artifacts.go` -- uses `go:embed` (via sprint.go `bmadOutputDir`)
- `internal/bmad/types.go` -- struct definition patterns
- `go.mod` -- confirms `gopkg.in/yaml.v3` availability

## Acceptance Criteria

AC-1: Advice file parsing
- Given a `.md` file with YAML frontmatter containing `name`, `displayName`, `icon`, and `order` fields
- When `parseAdviceFile` is called with the file path
- Then it returns an `AdviceMode` with all frontmatter fields populated and `Body` containing the markdown below the closing `---`

AC-2: Three-source discovery and merge
- Given bundled defaults with modes "clean-code" and "solid", a global override for "clean-code", and a local addition "my-custom"
- When `LoadAdviceModes(repoPath)` is called
- Then it returns 3 modes: the global "clean-code" (overriding bundled), the bundled "solid", and the local "my-custom"
- And each mode has the correct `Source` field

AC-3: Sort order
- Given modes with different `Order` values (10, 20, 30) and two modes with the same `Order`
- When `LoadAdviceModes` returns the merged list
- Then modes are sorted by `Order` ascending, with ties broken by `DisplayName` alphabetically

AC-4: Graceful degradation on missing directories
- Given a repo path where `{repoPath}/.claude/advice/` does not exist and `~/.mashed/advice/` does not exist
- When `LoadAdviceModes(repoPath)` is called
- Then it returns only bundled defaults without error

AC-5: Malformed file handling
- Given an advice directory containing a file with invalid YAML frontmatter alongside a valid file
- When `LoadAdviceModes` scans that directory
- Then the valid file is included in the result and the malformed file is skipped without causing an error

AC-6: Body retrieval
- Given a loaded set of advice modes including mode "clean-code"
- When `LoadAdviceBody(repoPath, "clean-code")` is called
- Then it returns the full markdown body of the matching file (not just the frontmatter)

## BDD Test Scenarios

### Scenario 1: Parse a valid advice file

```gherkin
Feature: Advice file parsing

  Scenario: Parse valid frontmatter and body
    Given a file at "/tmp/test/strategic.md" with content:
      """
      ---
      name: strategic
      displayName: Strategic Advisor
      icon: compass
      order: 10
      ---
      You are a strategic code reviewer. Focus on architecture decisions...
      """
    When parseAdviceFile is called with path "/tmp/test/strategic.md" and source "global"
    Then the returned AdviceMode has name "strategic"
    And displayName "Strategic Advisor"
    And icon "compass"
    And order 10
    And source "global"
    And body starts with "You are a strategic"

  Scenario: Parse file with missing optional fields
    Given a file with only name in frontmatter
    When parseAdviceFile is called
    Then displayName defaults to name, icon defaults to empty, order defaults to 0

  Scenario: Parse file with no frontmatter delimiter
    Given a file with no "---" lines
    When parseAdviceFile is called
    Then it returns an error
```

### Scenario 2: Three-source merge

```gherkin
Feature: Advice mode discovery and merge

  Scenario: Local overrides global overrides bundled
    Given bundled defaults contain "clean-code" with displayName "Clean Code (Bundled)"
    And global advice contains "clean-code" with displayName "Clean Code (Custom)"
    And local advice contains "clean-code" with displayName "Clean Code (Project)"
    When LoadAdviceModes is called with the repo path
    Then only one "clean-code" mode is returned
    And its displayName is "Clean Code (Project)"
    And its source is "local"

  Scenario: Modes from all sources coexist
    Given bundled has "solid", global has "security-first", local has "my-custom"
    When LoadAdviceModes is called
    Then 3 modes are returned with sources "bundled", "global", "local" respectively

  Scenario: Empty repo path skips local
    Given repoPath is empty string
    When LoadAdviceModes is called
    Then only bundled and global modes are returned
```

### Scenario 3: Edge cases

```gherkin
Feature: Advice loading edge cases

  Scenario: Non-md files are ignored
    Given a directory containing "notes.txt" and "valid.md"
    When the directory is scanned for advice files
    Then only "valid.md" is parsed

  Scenario: Directory does not exist
    Given globalAdviceDir returns a path that does not exist
    When LoadAdviceModes is called
    Then no error is returned
    And bundled defaults are still loaded

  Scenario: File with name field missing
    Given an advice file with frontmatter but no "name" field
    When parseAdviceFile is called
    Then it returns an error indicating missing name
```

## Tasks / Subtasks

- [ ] Task 1: Define types (AC: AC-1)
  - [ ] Create `internal/advice/types.go` with `AdviceMode` struct
  - [ ] Ensure JSON and YAML struct tags are correct
  - [ ] Create `internal/advice/defaults/` directory with `.gitkeep`

- [ ] Task 2: Implement file parsing (AC: AC-1, AC-5)
  - [ ] Implement `parseAdviceFile(path, source)` with YAML frontmatter splitting
  - [ ] Implement `parseAdviceFileFromFS(fs, path)` for embedded files
  - [ ] Validate required `name` field, return error if missing
  - [ ] Handle malformed YAML gracefully (return error, don't panic)

- [ ] Task 3: Implement discovery and merge (AC: AC-2, AC-3, AC-4)
  - [ ] Implement `globalAdviceDir()` using `os.UserHomeDir()`
  - [ ] Implement `localAdviceDir(repoPath)` returning `{repoPath}/.claude/advice/`
  - [ ] Implement `LoadAdviceModes(repoPath)` with three-source scan and merge-by-name
  - [ ] Implement sort by `Order` then `DisplayName`

- [ ] Task 4: Implement body retrieval (AC: AC-6)
  - [ ] Implement `LoadAdviceBody(repoPath, modeName)` that finds the highest-priority file for the given name and returns its body

- [ ] Task 5: Write tests (AC: AC-1 through AC-6)
  - [ ] Table-driven tests for `parseAdviceFile` (valid, missing name, no frontmatter, malformed YAML)
  - [ ] Tests for `LoadAdviceModes` with `t.TempDir()` for global/local dirs
  - [ ] Tests for merge priority (local > global > bundled)
  - [ ] Tests for sort order
  - [ ] Tests for `LoadAdviceBody`
  - [ ] Tests for missing/nonexistent directories

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/advice/`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
