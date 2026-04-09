# Story 4: Programmatic Skill Definition Generation

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** sprint4-01-artifact-path-system, sprint4-03-registry-expansion
**Status:** done

## Description

Create a Go test helper that programmatically generates SKILL.md files for all 32 BMAD processes from the registry and artifact path map. Each generated skill file tells Claude exactly which files to read and write during execution. This is the critical missing piece that prevents any BMAD workflow from actually running -- without skill definitions, `claude use bmad-brainstorming` fails immediately.

## Developer Notes

### Architecture
- **New file:** `internal/bmad/skillgen.go` -- contains `GenerateSkillFiles(baseDir string) error` and the skill template logic.
- **New file:** `internal/bmad/skillgen_test.go` -- tests that generate skills to a temp dir and verify structure.
- **Generated output:** 32+ directories at `{baseDir}/bmad-{name}/SKILL.md`.
  - When called from a test with `baseDir = projectRoot + "/.claude/skills"`, it generates the actual skill files.
  - The test itself also validates the generated content structure.
- The generation function iterates `AllProcesses()`, and for each process:
  1. Creates directory `{baseDir}/{proc.SkillName}/`
  2. Generates `SKILL.md` from a Go template using `text/template`
  3. Resolves input/output artifact paths via `ResolveArtifactPath` to embed exact paths in the skill

### Technical Considerations
- **Skill SKILL.md template** (Go `text/template`):
  ```
  ---
  name: {{.SkillName}}
  description: {{.Description}}
  ---

  # {{.Name}}

  ## What You Do
  {{.Description}}

  ## Inputs
  {{range .InputSpecs}}- Read `_bmad-output/{{.Path}}` for {{.Name}}
  {{end}}{{if not .InputSpecs}}- No file inputs required
  {{end}}

  ## Outputs
  {{range .OutputSpecs}}- Write `_bmad-output/{{.Path}}` — {{.Name}}
  {{end}}{{if not .OutputSpecs}}- No file outputs (operates on working directory)
  {{end}}

  ## Instructions
  1. Read any input artifacts listed above.
  2. Perform the task described in "What You Do."
  3. Write all output artifacts to the exact paths listed above.
  4. All artifacts go under `_bmad-output/` in the repository root.

  ## Output Convention
  Write all artifacts to `_bmad-output/` using the exact paths above.
  ```
- For each input/output string, use `ResolveArtifactPath(name, "")` to get the relative path portion. If empty (unmapped like `"code"`), use a special note: `"Operates on the repository codebase directly"`.
- Create a helper struct `skillTemplateData` to hold the template variables.
- Use `os.MkdirAll` for directory creation, `os.WriteFile` for SKILL.md.
- The function should be idempotent -- running it again overwrites existing files.

### Risks & Edge Cases
- Processes with empty inputs (like `bmad-brainstorming`) should generate a "No file inputs" section, not an empty list.
- Processes with unmapped outputs (like `"code"`) should note they operate on the repo directly.
- The `text/template` must handle special characters in descriptions (no HTML escaping needed for markdown).
- Directory permissions: use 0755 for dirs, 0644 for files.
- The test that generates actual skills should use a build tag or `TestGenerate` prefix so it can be run intentionally.

### Reference Files
- `internal/bmad/artifacts.go` -- `ResolveArtifactPath`, `artifactPaths` (from Story 1)
- `internal/bmad/registry.go` -- `AllProcesses()` (expanded by Story 3)
- `.claude/skills/desloppify/SKILL.md` -- example of existing skill file format
- `.claude/skills/skill-validator/SKILL.md` -- another example

## Acceptance Criteria

AC-1: GenerateSkillFiles creates a directory and SKILL.md for every registry process
- Given the full registry of 32 processes
- When `GenerateSkillFiles(tempDir)` is called
- Then 32 directories are created under `tempDir`, each named after the process's `SkillName`
- And each directory contains a `SKILL.md` file

AC-2: Generated SKILL.md contains correct frontmatter
- Given a generated SKILL.md for `bmad-create-prd`
- When the file is read
- Then it contains YAML frontmatter with `name: bmad-create-prd` and `description:` matching the process description

AC-3: Generated SKILL.md includes resolved input artifact paths
- Given a generated SKILL.md for `bmad-create-architecture` (inputs: `["PRD.md"]`)
- When the Inputs section is read
- Then it contains `_bmad-output/planning-artifacts/PRD.md`

AC-4: Generated SKILL.md includes resolved output artifact paths
- Given a generated SKILL.md for `bmad-create-prd` (outputs: `["PRD.md"]`)
- When the Outputs section is read
- Then it contains `_bmad-output/planning-artifacts/PRD.md`

AC-5: Unmapped artifacts have appropriate fallback text
- Given a generated SKILL.md for `bmad-dev-story` (outputs: `["code", "tests"]`)
- When the Outputs section is read
- Then it contains text indicating the process operates on the repository codebase directly

AC-6: Function is idempotent
- Given `GenerateSkillFiles(tempDir)` has already been called
- When called again with the same `tempDir`
- Then it succeeds without error
- And all SKILL.md files are present and valid

## BDD Test Scenarios

### Scenario 1: Generate all skill files

```gherkin
Feature: Programmatic skill generation

  Scenario: Generate skills for all 32 processes
    Given a temporary base directory
    When GenerateSkillFiles(tempDir) is called
    Then 32 subdirectories exist under tempDir
    And each subdirectory contains a SKILL.md file
    And no error is returned

  Scenario: Skill directory names match SkillName
    Given GenerateSkillFiles has been called
    When the directory names are listed
    Then each matches a SkillName from AllProcesses()
    And there are no extra or missing directories
```

### Scenario 2: SKILL.md content structure

```gherkin
Feature: Skill file content

  Scenario: Frontmatter is correct
    Given GenerateSkillFiles has been called
    When the SKILL.md for "bmad-brainstorming" is read
    Then line 1 is "---"
    And it contains "name: bmad-brainstorming"
    And it contains "description:"
    And the frontmatter closes with "---"

  Scenario: Input paths are resolved
    Given GenerateSkillFiles has been called
    When the SKILL.md for "bmad-create-architecture" is read
    Then the Inputs section contains "_bmad-output/planning-artifacts/PRD.md"

  Scenario: Process with no inputs has fallback text
    Given GenerateSkillFiles has been called
    When the SKILL.md for "bmad-brainstorming" is read
    Then the Inputs section contains "No file inputs"
```

### Scenario 3: Edge cases

```gherkin
Feature: Skill generation edge cases

  Scenario: Unmapped outputs use fallback
    Given GenerateSkillFiles has been called
    When the SKILL.md for "bmad-quick-dev" is read
    Then the Outputs section references direct codebase operation
    And does NOT contain "_bmad-output/" for the "code" artifact

  Scenario: Idempotent regeneration
    Given GenerateSkillFiles has been called once
    When GenerateSkillFiles is called again with the same directory
    Then no error is returned
    And all 32 SKILL.md files exist
```

## Tasks / Subtasks

- [ ] Task 1: Implement GenerateSkillFiles in skillgen.go (AC: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6)
  - [ ] Subtask 1a: Define `skillTemplateData` struct and the SKILL.md Go template
  - [ ] Subtask 1b: Implement `GenerateSkillFiles(baseDir string) error` that iterates AllProcesses
  - [ ] Subtask 1c: Resolve input/output artifact paths and handle unmapped artifacts
  - [ ] Subtask 1d: Use `os.MkdirAll` and `os.WriteFile` for file creation
- [ ] Task 2: Write tests in skillgen_test.go (AC: AC-1, AC-2, AC-3, AC-4, AC-5, AC-6)
  - [ ] Subtask 2a: Test that GenerateSkillFiles creates correct number of directories
  - [ ] Subtask 2b: Test SKILL.md frontmatter parsing for several processes
  - [ ] Subtask 2c: Test input/output path resolution in generated content
  - [ ] Subtask 2d: Test unmapped artifact fallback text
  - [ ] Subtask 2e: Test idempotency (call twice, verify no error)
- [ ] Task 3: Generate actual skill files (AC: AC-1)
  - [ ] Subtask 3a: Add a `TestGenerateActualSkills` test that writes to `.claude/skills/` (skipped by default, run manually)
  - [ ] Subtask 3b: Run the test to generate all 32 SKILL.md files

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `internal/bmad/skillgen.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
- [ ] 32 SKILL.md files generated in `.claude/skills/bmad-*/`
