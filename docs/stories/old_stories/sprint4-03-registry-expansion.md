# Story 3: Registry Expansion, Module Population & New Templates

**Priority:** P1-high
**Domain:** backend
**Estimated Complexity:** M
**Depends On:** none
**Status:** done

## Description

Expand the BMAD process registry with 7 new processes (Party Mode, Quick Flow, Adversarial Review, Infrastructure & DevOps, Document Project, Web Orchestrator, Game Dev Studio), populate the 4 empty module definitions with their process memberships, add 2 new workflow templates, and audit all existing templates for consistent input/output chains. This makes the workflow builder immediately more useful by tripling the available processes and providing infrastructure/game-dev workflows.

## Developer Notes

### Architecture
- **Modified file:** `internal/bmad/registry.go`
  - Add 7 new `ProcessDef` entries to the `init()` registry slice
  - New processes span analysis, implementation, and support phases
  - New module IDs used: `"cis"` (Infrastructure & DevOps), `"bmgd"` (Game Dev Studio); remaining 5 are `"core"`
- **Modified file:** `internal/bmad/modules.go`
  - `GetModules()` currently returns 4 empty modules (bmb, tea, bmgd, cis). After adding new registry entries, the function should dynamically populate ALL modules from the registry using `ProcessesByModule()`, not just `"core"`.
  - Refactor: replace the hardcoded `coreIDs` logic with a loop over all module IDs that collects process IDs dynamically.
- **Modified file:** `internal/bmad/templates.go`
  - Add `rapidPrototype()` function: 2 nodes (Party Mode -> Code Review)
  - Add `fullInfra()` function: 3 nodes (Architecture -> Infrastructure & DevOps -> Code Review)
  - Update `BuiltinTemplates()` to return 8 templates (6 existing + 2 new)
- **Modified files:** `internal/bmad/registry_test.go`, `internal/bmad/modules_test.go`, `internal/bmad/templates_test.go`
- **Template audit (Phase 6c):** Verify all 8 templates have valid input/output chains -- every downstream node's required inputs must appear in at least one upstream node's outputs. Add a test that enforces this.

### Technical Considerations
- New process definitions (exact fields):
  - `bmad-party-mode`: phase=implementation, role=developer, inputs=[], outputs=["code"], module="core"
  - `bmad-quick-flow`: phase=analysis, role=developer, inputs=[], outputs=["code", "PRD.md"], module="core"
  - `bmad-adversarial-general`: phase=support, role=qa, inputs=["any-doc"], outputs=["adversarial-report"], module="core"
  - `bmad-infrastructure-devops`: phase=implementation, role=architect, inputs=["architecture.md"], outputs=["infra-config"], module="cis"
  - `bmad-document-project`: phase=support, role=tech-writer, inputs=["PRD.md", "architecture.md"], outputs=["project-docs"], module="core"
  - `bmad-web-orchestrator`: phase=implementation, role=developer, inputs=["architecture.md", "epics/"], outputs=["code"], module="core"
  - `bmad-game-dev-studio`: phase=implementation, role=developer, inputs=["PRD.md"], outputs=["code"], module="bmgd"
- New artifact names introduced: `"adversarial-report"`, `"infra-config"`, `"project-docs"`. Ensure these exist in `artifactPaths` from Story 1.
- Template ID conventions: `"tpl-rapid-prototype"`, `"tpl-full-infra"`
- The `chain()` helper in templates.go already handles edge creation for linear workflows.

### Risks & Edge Cases
- Module population must remain backward compatible -- `GetModules()` already returns the 5 module definitions, just with empty process lists for non-core.
- Template node IDs must be unique across ALL templates (they use prefixes like `tpl-rp-`, `tpl-fi-`).
- The `SkillName` for new processes must match what will be created in Story 4. Use the convention `bmad-{id-suffix}`.
- Template audit may reveal inconsistencies in existing templates (e.g., a node expects `"architecture.md"` but no upstream produces it). Log and fix any found.

### Reference Files
- `internal/bmad/registry.go` -- existing 25 processes, init() pattern
- `internal/bmad/modules.go` -- current `GetModules()` implementation
- `internal/bmad/templates.go` -- `node()`, `edge()`, `chain()` helpers, existing 6 templates
- `internal/bmad/templates_test.go` -- existing template tests

## Acceptance Criteria

AC-1: Seven new processes are in the registry
- Given the registry is loaded
- When `AllProcesses()` is called
- Then the result contains 32 processes (25 existing + 7 new)
- And each new process has a valid ID, Name, Phase, AgentRole, SkillName, ModuleID

AC-2: Modules are populated with correct process IDs
- Given the modules are loaded
- When `GetModules()` is called
- Then the `"cis"` module's Processes list contains `"bmad-infrastructure-devops"`
- And the `"bmgd"` module's Processes list contains `"bmad-game-dev-studio"`
- And the `"core"` module's Processes list contains all core processes (30 entries)

AC-3: Two new templates are available
- Given the templates are loaded
- When `BuiltinTemplates()` is called
- Then the result contains 8 templates
- And `"tpl-rapid-prototype"` has 2 nodes: Party Mode -> Code Review
- And `"tpl-full-infra"` has 3 nodes: Architecture -> Infrastructure & DevOps -> Code Review

AC-4: All templates have valid input/output chains
- Given all 8 built-in templates
- When each template's edge chain is validated against process definitions
- Then for every edge from node A to node B, at least one output of A's process appears in B's process inputs (or B's inputs are empty)

AC-5: New processes are retrievable by phase and module
- Given the expanded registry
- When `ProcessesByPhase(PhaseImplementation)` is called
- Then the result includes `"bmad-party-mode"`, `"bmad-infrastructure-devops"`, `"bmad-web-orchestrator"`, `"bmad-game-dev-studio"`
- And `ProcessesByModule("cis")` returns exactly `["bmad-infrastructure-devops"]`

## BDD Test Scenarios

### Scenario 1: Registry expansion

```gherkin
Feature: Expanded BMAD process registry

  Scenario: All new processes are registered
    Given the BMAD registry is initialized
    When AllProcesses() is called
    Then the count is 32
    And ProcessByID("bmad-party-mode") returns found=true
    And ProcessByID("bmad-quick-flow") returns found=true
    And ProcessByID("bmad-adversarial-general") returns found=true
    And ProcessByID("bmad-infrastructure-devops") returns found=true
    And ProcessByID("bmad-document-project") returns found=true
    And ProcessByID("bmad-web-orchestrator") returns found=true
    And ProcessByID("bmad-game-dev-studio") returns found=true

  Scenario: New processes have correct phases
    Given the BMAD registry
    When ProcessesByPhase("support") is called
    Then the result includes "bmad-adversarial-general" and "bmad-document-project"

  Scenario: New processes have correct modules
    Given the BMAD registry
    When ProcessesByModule("cis") is called
    Then the result includes "bmad-infrastructure-devops"
    And ProcessesByModule("bmgd") includes "bmad-game-dev-studio"
```

### Scenario 2: Module population

```gherkin
Feature: Populated BMAD modules

  Scenario: Non-core modules have process lists
    Given the modules are loaded via GetModules()
    When the "cis" module is inspected
    Then its Processes list is ["bmad-infrastructure-devops"]
    And the "bmgd" module Processes list is ["bmad-game-dev-studio"]

  Scenario: Core module includes new core processes
    Given the modules are loaded
    When the "core" module is inspected
    Then its Processes list includes "bmad-party-mode"
    And includes "bmad-quick-flow"
    And includes "bmad-adversarial-general"
    And includes "bmad-document-project"
    And includes "bmad-web-orchestrator"
```

### Scenario 3: New templates

```gherkin
Feature: New workflow templates

  Scenario: Rapid Prototype template structure
    Given BuiltinTemplates() is called
    When the "tpl-rapid-prototype" template is found
    Then it has 2 nodes
    And node 1 references process "bmad-party-mode"
    And node 2 references process "bmad-code-review"
    And there is 1 edge from node 1 to node 2

  Scenario: Full Infrastructure template structure
    Given BuiltinTemplates() is called
    When the "tpl-full-infra" template is found
    Then it has 3 nodes
    And node 1 references process "bmad-create-architecture"
    And node 2 references process "bmad-infrastructure-devops"
    And node 3 references process "bmad-code-review"
    And edges chain: node 1 -> node 2 -> node 3
```

### Scenario 4: Template input/output chain validation

```gherkin
Feature: Template chain integrity

  Scenario: All templates have valid artifact chains
    Given all 8 built-in templates
    When each template's edges are checked
    Then for every edge A->B, either B's process has empty inputs, or at least one of A's process outputs appears in B's process inputs
```

## Tasks / Subtasks

- [ ] Task 1: Add 7 new ProcessDef entries to registry.go (AC: AC-1, AC-5)
  - [ ] Subtask 1a: Add bmad-party-mode, bmad-quick-flow to appropriate phase sections
  - [ ] Subtask 1b: Add bmad-adversarial-general, bmad-document-project to support section
  - [ ] Subtask 1c: Add bmad-infrastructure-devops, bmad-web-orchestrator, bmad-game-dev-studio to implementation section
- [ ] Task 2: Refactor GetModules() to dynamically populate all modules (AC: AC-2)
  - [ ] Subtask 2a: Replace hardcoded coreIDs logic with loop over all known module IDs
  - [ ] Subtask 2b: Use ProcessesByModule() for each module to build Processes lists
- [ ] Task 3: Add 2 new templates and audit existing ones (AC: AC-3, AC-4)
  - [ ] Subtask 3a: Implement `rapidPrototype()` function (2 nodes, 1 edge)
  - [ ] Subtask 3b: Implement `fullInfra()` function (3 nodes, 2 edges)
  - [ ] Subtask 3c: Update `BuiltinTemplates()` to include both new templates
  - [ ] Subtask 3d: Audit all 8 templates for input/output chain consistency; fix any mismatches
- [ ] Task 4: Write tests (AC: AC-1, AC-2, AC-3, AC-4, AC-5)
  - [ ] Subtask 4a: Update registry_test.go -- count test (32 processes), lookup tests for new IDs, phase/module filter tests
  - [ ] Subtask 4b: Update modules_test.go -- verify non-empty Processes lists for cis, bmgd, core
  - [ ] Subtask 4c: Update templates_test.go -- add tests for new templates, add template chain validation test

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on modified files
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `/simplify` run on all modified code
- [ ] Code review: no CRITICAL/HIGH issues
