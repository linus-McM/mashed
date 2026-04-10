# Story 5: Bundled Default Advice Files

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** S
**Depends On:** review-01
**Status:** done

## Description

Create the 8 bundled default advice `.md` files in `internal/advice/defaults/` and update the `go:embed` directive to include them. These files provide out-of-the-box methodology perspectives (strategic, extreme-programming, clean-code, solid, domain-driven-design, security-first, performance, pragmatic) that users see immediately in the advice dropdown without any configuration.

## Developer Notes

### Architecture

- **New files (8):**
  - `internal/advice/defaults/strategic.md`
  - `internal/advice/defaults/extreme-programming.md`
  - `internal/advice/defaults/clean-code.md`
  - `internal/advice/defaults/solid.md`
  - `internal/advice/defaults/domain-driven-design.md`
  - `internal/advice/defaults/security-first.md`
  - `internal/advice/defaults/performance.md`
  - `internal/advice/defaults/pragmatic.md`

- **Modified file:** `internal/advice/loader.go` -- ensure `go:embed` directive picks up the new `.md` files. The `//go:embed defaults/*.md` directive from story review-01 should already handle this. Remove `.gitkeep` if present.

### File Format

Each file follows this template:

```markdown
---
name: {kebab-case-id}
displayName: {Human Readable Name}
icon: {lucide-icon-name}
order: {10-80, spaced by 10}
---

You are an expert code reviewer applying the {methodology} perspective.

## Focus Areas

- {key principle 1}
- {key principle 2}
- {key principle 3}

## Review Approach

{2-3 paragraphs of instruction for how Claude should review code from this perspective}

## What to Look For

- {specific pattern or smell to identify}
- {specific improvement to suggest}
- {specific anti-pattern to flag}
```

### Specific Modes

| name | displayName | icon | order | Focus |
|------|-------------|------|-------|-------|
| strategic | Strategic Advisor | compass | 10 | Architecture, long-term maintainability, tech debt |
| extreme-programming | Extreme Programming | zap | 20 | TDD, pair review, simplicity, refactoring |
| clean-code | Clean Code | sparkles | 30 | Naming, functions, comments, formatting (Robert C. Martin) |
| solid | SOLID Principles | blocks | 40 | SRP, OCP, LSP, ISP, DIP |
| domain-driven-design | Domain-Driven Design | layers | 50 | Bounded contexts, ubiquitous language, aggregates |
| security-first | Security First | shield | 60 | Input validation, auth, injection, data exposure |
| performance | Performance | gauge | 70 | Memory, CPU, I/O, caching, algorithmic complexity |
| pragmatic | Pragmatic Developer | wrench | 80 | Trade-offs, shipping velocity, YAGNI, good enough |

### Technical Considerations

- **Body content quality:** Each advice file body should be 200-400 words of substantive review instructions. These are system prompts for Claude -- they should be specific enough to produce meaningfully different reviews.
- **Icon names:** Must be valid `lucide-svelte` icon names. Verify against the lucide icon set.
- **go:embed:** After adding files, `go build ./...` must pass. The embed directive in `loader.go` (`//go:embed defaults/*.md`) matches `*.md` files in the `defaults/` subdirectory.
- **No runtime code changes:** This story only creates content files. The loading logic from review-01 handles everything.

### Risks & Edge Cases

- **Embed build failure:** If the `defaults/` directory is empty (no `.md` files yet), the `//go:embed defaults/*.md` pattern will cause a build error. Story review-01 should use `all:defaults` to handle this. If not, this story should fix the embed directive when adding files.
- **Icon name mismatch:** If an icon name doesn't exist in lucide, the frontend will show nothing. Verify icon names.
- **Frontmatter YAML errors:** Carefully validate YAML indentation. No tabs, consistent spacing.

### Reference Files

- `internal/advice/types.go` (from review-01) -- `AdviceMode` struct fields to match
- `internal/advice/loader.go` (from review-01) -- `go:embed` directive and parsing logic
- `DESIGN.md` -- tone and style guidance for the app

## Acceptance Criteria

AC-1: All 8 advice files exist
- Given the `internal/advice/defaults/` directory
- When listing its contents
- Then 8 `.md` files exist: strategic, extreme-programming, clean-code, solid, domain-driven-design, security-first, performance, pragmatic

AC-2: Each file has valid frontmatter
- Given any of the 8 bundled advice files
- When parsed by `parseAdviceFile`
- Then all 4 frontmatter fields (name, displayName, icon, order) are present and valid
- And the `name` field matches the filename (without `.md` extension)

AC-3: Each file has substantive body
- Given any of the 8 bundled advice files
- When parsed by `parseAdviceFile`
- Then the `Body` field is non-empty and contains at least 100 words of review instructions

AC-4: Files load via go:embed
- Given the `defaultAdviceFS` embed.FS in `loader.go`
- When `LoadAdviceModes("")` is called (no repo path, no global/local)
- Then 8 modes are returned with source "bundled"
- And they are sorted by order (10, 20, 30, 40, 50, 60, 70, 80)

AC-5: Build passes
- Given all 8 files are added to `internal/advice/defaults/`
- When `go build ./...` is run
- Then it succeeds without errors

## BDD Test Scenarios

### Scenario 1: Bundled defaults load correctly

```gherkin
Feature: Bundled advice defaults

  Scenario: All 8 defaults present
    Given the embedded defaults filesystem
    When LoadAdviceModes is called with empty repoPath
    Then 8 AdviceMode structs are returned
    And names are ["strategic", "extreme-programming", "clean-code", "solid", "domain-driven-design", "security-first", "performance", "pragmatic"]
    And each has a non-empty displayName, icon, and body

  Scenario: Order is correct
    Given the 8 bundled defaults are loaded
    Then they are sorted with orders [10, 20, 30, 40, 50, 60, 70, 80]
    And "strategic" is first and "pragmatic" is last
```

### Scenario 2: Individual file validity

```gherkin
Feature: Advice file format

  Scenario: Each file parses without error
    Given each of the 8 .md files in defaults/
    When parsed individually by parseAdviceFile
    Then none return an error
    And each has name matching filename
    And each has displayName that is non-empty
    And each has order that is a positive integer
```

## Tasks / Subtasks

- [ ] Task 1: Create advice content files (AC: AC-1, AC-2, AC-3)
  - [ ] Write `strategic.md` -- architecture, tech debt, scalability focus
  - [ ] Write `extreme-programming.md` -- TDD, simplicity, refactoring focus
  - [ ] Write `clean-code.md` -- naming, functions, formatting (Uncle Bob)
  - [ ] Write `solid.md` -- SRP, OCP, LSP, ISP, DIP principles
  - [ ] Write `domain-driven-design.md` -- bounded contexts, aggregates, ubiquitous language
  - [ ] Write `security-first.md` -- input validation, auth, injection prevention
  - [ ] Write `performance.md` -- memory, CPU, caching, algorithmic complexity
  - [ ] Write `pragmatic.md` -- trade-offs, shipping, YAGNI, good enough

- [ ] Task 2: Verify embed and build (AC: AC-4, AC-5)
  - [ ] Remove `.gitkeep` from `defaults/` if present
  - [ ] Verify `//go:embed defaults/*.md` or `all:defaults` directive works
  - [ ] Run `go build ./...` to confirm embed succeeds
  - [ ] Run existing `LoadAdviceModes` tests to verify 8 modes load

- [ ] Task 3: Add integration test (AC: AC-1 through AC-4)
  - [ ] Add test in `loader_test.go` that loads embedded defaults and asserts count = 8
  - [ ] Assert each mode has non-empty Body (>100 words)
  - [ ] Assert order values are unique and sequential by 10s
  - [ ] Assert names match filenames

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage maintained on `internal/advice/`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
