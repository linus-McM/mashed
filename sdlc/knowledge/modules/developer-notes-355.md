---
type: Module
title: Developer Notes
description: "Graphify community 355: docs/SPECIFICATION.md, docs/stories/old_stories/bmad-05-wails-bindings.md, docs/stories/old_stories/bmad-08-execution-integration.md, docs/stories/old_stories/sprint2-04-repo-c"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: SPECIFICATION, resource: docs/SPECIFICATION.md, last_modified: "2026-04-12T09:58:35+10:00", digest: 2150575fba9a1ee2 }
  - { id: bmad-05-wails-bindings, resource: docs/stories/old_stories/bmad-05-wails-bindings.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d868ae2882f2c29b }
  - { id: bmad-08-execution-integration, resource: docs/stories/old_stories/bmad-08-execution-integration.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 84e85833652d77c6 }
  - { id: sprint2-04-repo-context-flow, resource: docs/stories/old_stories/sprint2-04-repo-context-flow.md, last_modified: "2026-04-08T17:10:27+10:00", digest: bee256727da15650 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/SPECIFICATION.md`
- `docs/stories/old_stories/bmad-05-wails-bindings.md`
- `docs/stories/old_stories/bmad-08-execution-integration.md`
- `docs/stories/old_stories/sprint2-04-repo-context-flow.md`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- 7. Terminal Bridge (docs/SPECIFICATION.md:L620)
- Two-process architecture (docs/SPECIFICATION.md:L622)
- Helper Protocol (`internal/terminal/helper/protocol.go`) (docs/SPECIFICATION.md:L639)
- WebSocket Bridge (`internal/terminal/bridge.go`) (docs/SPECIFICATION.md:L643)
- Session Registry (`app_terminal_registry.go`) (docs/SPECIFICATION.md:L647)
- tmux Pane Discovery (`internal/terminal/panes.go`) (docs/SPECIFICATION.md:L651)
- Agent Spawning (`app_spawn.go`) (docs/SPECIFICATION.md:L659)
- Risks & Edge Cases (docs/stories/old_stories/bmad-05-wails-bindings.md:L127)
- Execution Integration: Live Status, Terminal Access & Artifact Passing (docs/stories/old_stories/bmad-08-execution-integration.md:L1)
- Terminal Access for Running Nodes (docs/stories/old_stories/bmad-08-execution-integration.md:L103)
- Developer Notes (docs/stories/old_stories/bmad-08-execution-integration.md:L12)
- Artifact Context Passing (Backend) (docs/stories/old_stories/bmad-08-execution-integration.md:L121)
- Event Data Shapes (docs/stories/old_stories/bmad-08-execution-integration.md:L130)
- Architecture (docs/stories/old_stories/bmad-08-execution-integration.md:L14)
- Technical Considerations (docs/stories/old_stories/bmad-08-execution-integration.md:L149)
- Risks & Edge Cases (docs/stories/old_stories/bmad-08-execution-integration.md:L157)
- Reference Files (docs/stories/old_stories/bmad-08-execution-integration.md:L164)
- Acceptance Criteria (docs/stories/old_stories/bmad-08-execution-integration.md:L173)
- BDD Test Scenarios (docs/stories/old_stories/bmad-08-execution-integration.md:L183)
- Scenario 1: End-to-end execution (docs/stories/old_stories/bmad-08-execution-integration.md:L185)
- Scenario 2: Terminal access (docs/stories/old_stories/bmad-08-execution-integration.md:L214)
- Frontend Event Wiring (docs/stories/old_stories/bmad-08-execution-integration.md:L22)
- Scenario 3: Error handling (docs/stories/old_stories/bmad-08-execution-integration.md:L227)
- Definition of Done (docs/stories/old_stories/bmad-08-execution-integration.md:L271)
- Description (docs/stories/old_stories/bmad-08-execution-integration.md:L8)
- sprint2-04-repo-context-flow.md (docs/stories/old_stories/sprint2-04-repo-context-flow.md:L1)
- Story 4: Repo Context Flow (App.svelte to WorkflowBuilder) (docs/stories/old_stories/sprint2-04-repo-context-flow.md:L1)
- BDD Test Scenarios (docs/stories/old_stories/sprint2-04-repo-context-flow.md:L118)
- Scenario 1: Repo selection flow (docs/stories/old_stories/sprint2-04-repo-context-flow.md:L120)
- Definition of Done (docs/stories/old_stories/sprint2-04-repo-context-flow.md:L178)
- Acceptance Criteria (docs/stories/old_stories/sprint2-04-repo-context-flow.md:L87)
- Description (docs/stories/old_stories/sprint2-04-repo-context-flow.md:L9)
- GetTerminalPort() (frontend/wailsjs/go/main/App.js:L113)
- StartBmadWorkflow() (frontend/wailsjs/go/main/App.js:L437)
- GetBmadExecution() (frontend/wailsjs/go/main/App.js:L45)

# Depends on
- [13. Edge Cases and Failure Modes](/modules/13-edge-cases-and-failure-modes.md)
- [Tasks / Subtasks](/modules/tasks-subtasks.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
