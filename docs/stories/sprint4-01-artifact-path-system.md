# Story 1: Artifact Path Convention & Resolution System

**Priority:** P0-critical
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** ready

## Description

Create a convention-based artifact path resolution system that maps artifact names (e.g., `"PRD.md"`, `"brainstorm-notes"`) to canonical file paths under `_bmad-output/`. This foundational layer enables the executor to pass actual file paths to skills and verify artifacts exist on disk after processes complete. Without this, downstream processes have no way to locate upstream outputs.

## Developer Notes

### Architecture
- **New file:** `internal/bmad/artifacts.go` -- contains the `artifactPaths` map, `ResolveArtifactPath()`, and `VerifyArtifacts()`.
- **Modified file:** `internal/bmad/types.go` -- add `ArtifactType` constants and `ArtifactSpec` struct.
- **New file:** `internal/bmad/artifacts_test.go` -- comprehensive tests.
- The artifact path map is a package-level `var` (like `registry`), initialized at package load. It maps artifact names (matching `ProcessDef.Inputs`/`ProcessDef.Outputs` string values) to relative paths under `_bmad-output/`.
- No changes to `ProcessDef.Inputs`/`Outputs` -- they remain `[]string`. The `artifactPaths` map is the bridge layer.

### Technical Considerations
- `ResolveArtifactPath(name, repoPath string) string`:
  - Look up `name` in `artifactPaths`.
  - If found and non-empty, return `filepath.Join(repoPath, "_bmad-output", artifactPaths[name])`.
  - If found but empty string (like `"code"`, `"tests"`), return `""` (unmapped artifact).
  - If not found, return `""`.
- `VerifyArtifacts(repoPath string, outputNames []string) (found, missing []string)`:
  - For each output name, resolve the path via `ResolveArtifactPath`.
  - If path is empty (unmapped), skip it (do not count as missing).
  - If path ends with `/` (directory artifact), check with `os.Stat`.
  - Otherwise check file existence with `os.Stat`.
  - Return two slices: found artifact names and missing artifact names.
- Use `filepath.Join` for all path construction (cross-platform safety).
- No concurrency concerns -- these are pure functions operating on an immutable map.

### Risks & Edge Cases
- Artifact names containing wildcards (e.g., `"story-*.md"`) map to directories, not individual files. `ResolveArtifactPath` should return the directory path for these.
- The `_bmad-output/` directory may not exist yet (first run). `VerifyArtifacts` should handle this gracefully -- all artifacts are "missing" if the root dir does not exist.
- Empty `outputNames` slice should return two empty slices, not nil.

### Reference Files
- `internal/bmad/registry.go` -- existing artifact names in `ProcessDef.Inputs`/`Outputs`
- `internal/bmad/types.go` -- existing type patterns (BmadPhase, NodeType)
- `internal/bmad/registry_test.go` -- test pattern reference
- `internal/bmad/executor_test.go` -- `testHarness` pattern, `t.TempDir()` usage

## Acceptance Criteria

AC-1: Artifact path map covers all registry artifacts
- Given the artifact path map is defined in `artifacts.go`
- When compared against all `Inputs` and `Outputs` values across every `ProcessDef` in the registry
- Then every unique artifact name has an entry in the map (no unmapped names except intentionally empty ones like `"code"`, `"tests"`)

AC-2: ResolveArtifactPath returns correct paths
- Given a repo path of `/tmp/myrepo` and artifact name `"PRD.md"`
- When `ResolveArtifactPath("PRD.md", "/tmp/myrepo")` is called
- Then it returns `"/tmp/myrepo/_bmad-output/planning-artifacts/PRD.md"`
- And for unmapped artifacts like `"code"`, it returns `""`

AC-3: VerifyArtifacts detects found and missing artifacts
- Given a repo path with `_bmad-output/planning-artifacts/PRD.md` existing on disk
- When `VerifyArtifacts(repoPath, []string{"PRD.md", "architecture.md"})` is called
- Then `found` contains `"PRD.md"` and `missing` contains `"architecture.md"`
- And unmapped artifacts (e.g., `"code"`) are excluded from both slices

AC-4: ArtifactSpec and ArtifactType types are defined
- Given the `types.go` file
- When the `ArtifactType` and `ArtifactSpec` types are added
- Then `ArtifactType` has constants for `ArtifactMarkdown`, `ArtifactYAML`, `ArtifactDirectory`, `ArtifactCode`
- And `ArtifactSpec` has fields `Name`, `Type`, `Path`, `Description`, `Optional` with correct JSON tags

AC-5: Directory artifacts are verified correctly
- Given a repo path with `_bmad-output/solutioning-artifacts/epics/` directory existing on disk
- When `VerifyArtifacts(repoPath, []string{"epics/"})` is called
- Then `found` contains `"epics/"`

## BDD Test Scenarios

### Scenario 1: Resolve known artifact paths

```gherkin
Feature: Artifact path resolution

  Scenario: Resolve a known markdown artifact
    Given the artifact path map contains "PRD.md" -> "planning-artifacts/PRD.md"
    And a repo path of "/tmp/test-repo"
    When ResolveArtifactPath("PRD.md", "/tmp/test-repo") is called
    Then the result is "/tmp/test-repo/_bmad-output/planning-artifacts/PRD.md"

  Scenario: Resolve an unmapped artifact
    Given the artifact path map contains "code" -> ""
    And a repo path of "/tmp/test-repo"
    When ResolveArtifactPath("code", "/tmp/test-repo") is called
    Then the result is ""

  Scenario: Resolve an unknown artifact name
    Given the artifact path map does not contain "unknown-thing"
    When ResolveArtifactPath("unknown-thing", "/tmp/test-repo") is called
    Then the result is ""

  Scenario: Resolve a directory artifact
    Given the artifact path map contains "epics/" -> "solutioning-artifacts/epics/"
    And a repo path of "/tmp/test-repo"
    When ResolveArtifactPath("epics/", "/tmp/test-repo") is called
    Then the result is "/tmp/test-repo/_bmad-output/solutioning-artifacts/epics/"
```

### Scenario 2: Verify artifacts on disk

```gherkin
Feature: Artifact verification

  Scenario: Mixed found and missing artifacts
    Given a temp directory with "_bmad-output/planning-artifacts/PRD.md" created
    And "_bmad-output/solutioning-artifacts/architecture.md" does NOT exist
    When VerifyArtifacts(repoPath, ["PRD.md", "architecture.md"]) is called
    Then found is ["PRD.md"]
    And missing is ["architecture.md"]

  Scenario: Unmapped artifacts are excluded
    Given a temp directory
    When VerifyArtifacts(repoPath, ["code", "tests"]) is called
    Then found is empty
    And missing is empty

  Scenario: Empty output list
    Given a temp directory
    When VerifyArtifacts(repoPath, []) is called
    Then found is empty
    And missing is empty

  Scenario: _bmad-output directory does not exist
    Given a temp directory with no _bmad-output subdirectory
    When VerifyArtifacts(repoPath, ["PRD.md", "architecture.md"]) is called
    Then found is empty
    And missing is ["PRD.md", "architecture.md"]

  Scenario: Directory artifact exists
    Given a temp directory with "_bmad-output/solutioning-artifacts/epics/" directory created
    When VerifyArtifacts(repoPath, ["epics/"]) is called
    Then found is ["epics/"]
```

### Scenario 3: Artifact path map completeness

```gherkin
Feature: Artifact map completeness

  Scenario: All registry artifacts are mapped
    Given the full BMAD process registry
    When all unique Inputs and Outputs values are collected
    Then every artifact name exists as a key in artifactPaths
```

## Tasks / Subtasks

- [ ] Task 1: Add ArtifactType and ArtifactSpec to types.go (AC: AC-4)
  - [ ] Subtask 1a: Define `ArtifactType` string type with 4 constants (ArtifactMarkdown, ArtifactYAML, ArtifactDirectory, ArtifactCode)
  - [ ] Subtask 1b: Define `ArtifactSpec` struct with Name, Type, Path, Description, Optional fields and JSON tags
- [ ] Task 2: Create artifacts.go with path map and helpers (AC: AC-1, AC-2, AC-3, AC-5)
  - [ ] Subtask 2a: Define `artifactPaths` map with all 24 entries from the plan
  - [ ] Subtask 2b: Implement `ResolveArtifactPath(name, repoPath string) string`
  - [ ] Subtask 2c: Implement `VerifyArtifacts(repoPath string, outputNames []string) (found, missing []string)`
- [ ] Task 3: Create artifacts_test.go with comprehensive tests (AC: AC-1, AC-2, AC-3, AC-4, AC-5)
  - [ ] Subtask 3a: Table-driven tests for `ResolveArtifactPath` (known, unmapped, unknown, directory)
  - [ ] Subtask 3b: Tests for `VerifyArtifacts` using `t.TempDir()` (mixed, unmapped, empty, missing dir, directory artifact)
  - [ ] Subtask 3c: Registry completeness test -- iterate all ProcessDefs, assert every artifact name is in `artifactPaths`

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/bmad/artifacts.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
