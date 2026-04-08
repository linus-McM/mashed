# Module Versioning & Custom Agent Management

**Story ID**: bmad-09
**Status**: ready
**Priority**: P2
**Depends On**: bmad-05, bmad-07

## Description

Add the module versioning system and custom agent creation UI. Modules group BMAD processes by package (core, BMB, TEA, BMGD, CIS) with version tracking for future upgradability. Custom agents let users create personalized BMAD agent personas with specific model preferences, custom system prompts, and skill restrictions, then assign them to workflow nodes. This is a polish/extensibility story that enhances the workflow builder but is not required for core functionality.

## Developer Notes

### Architecture

**New files:**
- `internal/bmad/modules.go` -- Module definitions and version tracking
- `internal/bmad/modules_test.go` -- Module tests
- `frontend/src/components/bmad/AgentConfigModal.svelte` -- Modal for creating/editing custom agents

**Modified files:**
- `frontend/src/views/WorkflowBuilder.svelte` -- Add agent management button, wire AgentConfigModal
- `frontend/src/components/bmad/ProcessSidebar.svelte` -- Show module badges on process items
- `frontend/src/components/bmad/NodeConfigPanel.svelte` -- Populate agent dropdown from ListBmadAgents

### modules.go

```go
package bmad

type ModuleDef struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    Version     string   `json:"version"`
    Processes   []string `json:"processes"`  // process IDs in this module
    UpgradePath string   `json:"upgradePath,omitempty"`
}

func GetModules() []ModuleDef
```

Initial modules (all version "1.0.0"):
- `core` -- All 25 processes from the registry
- `bmb` -- Empty (reserved for BMAD Module Builder processes)
- `tea` -- Empty (reserved for Testing & Engineering Automation)
- `bmgd` -- Empty (reserved for BMAD Game Design)
- `cis` -- Empty (reserved for CI/CD & Infrastructure)

`GetModules` returns all defined modules. For the initial implementation, only `core` has processes. The others are placeholders that show the extensibility model.

### AgentConfigModal.svelte

Modal dialog for creating/editing custom BMAD agents:

```
+----------------------------------------+
| Create Custom Agent            [Close] |
|                                        |
| Name: [________________]              |
| Role: [dropdown: analyst/pm/dev/...]   |
| Model: [dropdown: opus/sonnet/haiku]   |
|                                        |
| Persona (system prompt additions):     |
| [multiline textarea]                   |
|                                        |
| Allowed Skills:                        |
| [x] bmad-brainstorming                |
| [x] bmad-create-prd                   |
| [ ] bmad-dev-story                    |
| ...                                    |
|                                        |
| [Save Agent]  [Delete]                |
+----------------------------------------+
```

Props:
- `agent` -- BmadAgentConfig to edit (null for new)
- `processes` -- all processes (for skill checkboxes)

Events:
- `save` with the agent config
- `delete` with agent ID
- `close`

On save: call `SaveBmadAgent(agent)` then refresh the agents list.

### Process Sidebar Module Badges

Each process item in the sidebar shows a small badge indicating its module:
```svelte
<span class="module-badge">{process.moduleId}</span>
```

Style with a subtle background color per module.

### NodeConfigPanel Agent Dropdown

The agent dropdown in NodeConfigPanel (from bmad-07) should now be populated:
```svelte
<select bind:value={nodeConfig.agentId}>
  <option value="">Default (no custom agent)</option>
  {#each agents as agent}
    <option value={agent.id}>{agent.name} ({agent.role})</option>
  {/each}
</select>
```

When an agent is assigned to a node, the executor uses that agent's model and persona when constructing the Claude CLI command.

### Technical Considerations

- **Module versioning** is forward-looking infrastructure. For MVP, all modules are "1.0.0" and there is no upgrade check. The `CheckForUpgrades` function can return `("", false)` for now.
- **Agent skill restriction**: When an agent has a non-empty `Skills` slice, the executor should verify the node's process is in the agent's skill list before execution. If not, return an error.
- **Agent persona injection**: When an agent is assigned to a node, the executor modifies the command: `claude --dangerously-skip-permissions --model {agent.model} --system-prompt "{agent.persona}" "use {skillName}"`. Verify the exact Claude CLI flag for custom system prompts.
- **Storage**: Agents are persisted at `~/.mashed/bmad-agents/{id}.json` (handled by Storage from bmad-03).

### Risks & Edge Cases

- **Agent with no skills**: If `Skills` is empty, the agent can execute any process. This is the default.
- **Deleted agent still referenced by node**: If a user deletes a custom agent that is assigned to workflow nodes, those nodes should fall back to default execution (no custom agent).
- **Module with no processes**: The reserved modules (bmb, tea, etc.) have empty process lists. The sidebar should not show empty sections for modules with no processes.
- **Persona length**: The custom persona text could be very long. Truncate display in the modal but send the full text to the executor.

### Reference Files

- `frontend/src/views/SpawnAgent.svelte` -- Modal pattern with form inputs and dropdowns
- `internal/bmad/registry.go` (from bmad-02) -- ProcessDef has moduleId field
- `internal/bmad/storage.go` (from bmad-03) -- SaveAgent/ListAgents/DeleteAgent
- `app.go` -- SaveBmadAgent/ListBmadAgents/DeleteBmadAgent bindings (from bmad-05)

## Acceptance Criteria

- [ ] AC1: Given `GetModules()` is called, When the result is inspected, Then it returns at least 5 modules, with the `core` module containing all 25 process IDs.
- [ ] AC2: Given the AgentConfigModal is opened, When the user fills in name, role, model, persona, and selects skills, Then clicking Save calls `SaveBmadAgent` and the agent is persisted.
- [ ] AC3: Given a custom agent exists, When the NodeConfigPanel is opened for a node, Then the agent appears in the agent dropdown and can be assigned to the node.
- [ ] AC4: Given a node has a custom agent assigned with model "claude-opus-4-6", When the workflow executes, Then the executor uses "claude-opus-4-6" for that node's Claude CLI command.
- [ ] AC5: Given the ProcessSidebar, When processes load, Then each process shows a small module badge (e.g., "core").
- [ ] AC6: Given a custom agent is deleted, When a node that referenced that agent executes, Then it falls back to default execution (no custom persona, default model).

## BDD Test Scenarios

### Scenario 1: Module definitions

```gherkin
Feature: BMAD Modules

  Scenario: GetModules returns all modules
    Given the bmad package is imported
    When GetModules() is called
    Then at least 5 modules are returned
    And the "core" module has 25 process IDs
    And all process IDs reference valid registry entries

  Scenario: Reserved modules have no processes
    Given GetModules() is called
    When the "bmb" module is inspected
    Then its Processes slice is empty
    And its Version is "1.0.0"
```

### Scenario 2: Custom agent CRUD

```gherkin
Feature: Custom BMAD Agents

  Scenario: Create custom agent
    Given the AgentConfigModal is open
    And the user enters name "Fast Developer", role "developer", model "claude-sonnet-4-20250514"
    And selects skills "bmad-dev-story", "bmad-code-review"
    When Save is clicked
    Then SaveBmadAgent is called with the agent data
    And the agent appears in ListBmadAgents

  Scenario: Delete custom agent
    Given a custom agent "Fast Developer" exists
    When Delete is clicked in the AgentConfigModal
    Then DeleteBmadAgent is called
    And the agent no longer appears in ListBmadAgents

  Scenario: Assign agent to node
    Given a custom agent "Fast Developer" exists
    And a node "Dev Story" is selected on the canvas
    When the user selects "Fast Developer" in the NodeConfigPanel agent dropdown
    Then the node's config is updated with the agent ID
```

### Scenario 3: Agent in execution

```gherkin
Feature: Agent-Aware Execution

  Scenario: Node uses assigned agent's model
    Given node "Dev Story" has agent "Fast Developer" assigned with model "claude-sonnet-4-20250514"
    When the executor spawns this node
    Then the Claude CLI command uses --model claude-sonnet-4-20250514

  Scenario: Deleted agent falls back to default
    Given node "Dev Story" has agent "Deleted Agent" assigned
    And "Deleted Agent" has been deleted from storage
    When the executor spawns this node
    Then the default model is used
    And no persona is injected
```

## Tasks / Subtasks

- [ ] Task 1: Implement modules.go (AC: AC1)
  - [ ] Subtask 1a: Define ModuleDef struct
  - [ ] Subtask 1b: Define the 5 module entries (core with all IDs, others empty)
  - [ ] Subtask 1c: Implement GetModules()
  - [ ] Subtask 1d: Write tests verifying core module cross-references with registry
- [ ] Task 2: Build AgentConfigModal.svelte (AC: AC2)
  - [ ] Subtask 2a: Create modal with name, role, model, persona inputs
  - [ ] Subtask 2b: Add skill checkboxes populated from GetBmadProcesses
  - [ ] Subtask 2c: Wire Save to SaveBmadAgent and Delete to DeleteBmadAgent
  - [ ] Subtask 2d: Add modal toggle in WorkflowBuilder (agent management button)
- [ ] Task 3: Populate NodeConfigPanel agent dropdown (AC: AC3)
  - [ ] Subtask 3a: Load agents via ListBmadAgents on panel open
  - [ ] Subtask 3b: Bind selected agent to node config
- [ ] Task 4: Add module badges to ProcessSidebar (AC: AC5)
  - [ ] Subtask 4a: Add moduleId badge to each process item
  - [ ] Subtask 4b: Style badges with subtle background colors
- [ ] Task 5: Handle agent-aware execution (AC: AC4, AC6)
  - [ ] Subtask 5a: Modify executor to check for assigned agent on node
  - [ ] Subtask 5b: Use agent model and persona in CLI command construction
  - [ ] Subtask 5c: Handle deleted-agent fallback gracefully

## Definition of Done

- [ ] All acceptance criteria pass
- [ ] All BDD scenarios pass as automated tests
- [ ] 80%+ code coverage on `modules.go`
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes
- [ ] `go test ./... -race` passes
- [ ] `cd frontend && npm run build` passes
- [ ] `wails build` passes
- [ ] /simplify run on all new code
