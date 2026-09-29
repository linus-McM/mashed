---
type: Module
title: SpawnAgent
description: "Graphify community 60: docs/reports/pty-fork-exec-investigation.md, docs/stories/old_stories/bmad-04-executor.md, docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md, docs/stories/"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: pty-fork-exec-investigation, resource: docs/reports/pty-fork-exec-investigation.md, last_modified: "2026-04-09T15:37:16+10:00", digest: fd87bcc9b9b43b3e }
  - { id: bmad-04-executor, resource: docs/stories/old_stories/bmad-04-executor.md, last_modified: "2026-04-08T10:23:03+10:00", digest: 6eb7b5db583a7b89 }
  - { id: pty-04-app-integration-and-frontend-cleanup, resource: docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md, last_modified: "2026-04-12T10:43:48+10:00", digest: f0f69a09af0aa8d1 }
  - { id: ptyhelper-06-main-integration, resource: docs/stories/old_stories/ptyhelper-06-main-integration.md, last_modified: "2026-04-12T10:43:48+10:00", digest: c229fa5e3bd21db3 }
  - { id: sessions-03-spawn-registration-events, resource: docs/stories/old_stories/sessions-03-spawn-registration-events.md, last_modified: "2026-04-09T10:03:44+10:00", digest: d396b884ad9c18b7 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/reports/pty-fork-exec-investigation.md`
- `docs/stories/old_stories/bmad-04-executor.md`
- `docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md`
- `docs/stories/old_stories/ptyhelper-06-main-integration.md`
- `docs/stories/old_stories/sessions-03-spawn-registration-events.md`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Files Modified During Investigation (docs/reports/pty-fork-exec-investigation.md:L270)
- Reference Files (docs/stories/old_stories/bmad-04-executor.md:L112)
- Developer Notes (docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md:L19)
- Backend: App Struct & Spawn Integration (docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md:L21)
- Technical Considerations (docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md:L61)
- Risks & Edge Cases (docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md:L68)
- Reference Files (docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md:L74)
- Graceful Degradation (docs/stories/old_stories/ptyhelper-06-main-integration.md:L77)
- sessions-03-spawn-registration-events.md (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L1)
- Story 3: Register Sessions on Spawn & Emit Wails Events (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L1)
- BDD Test Scenarios (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L102)
- Scenario 1: Agent Spawn Registration (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L104)
- Scenario 2: Terminal Spawn Registration (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L118)
- Developer Notes (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L13)
- Scenario 3: Kill Deregisters (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L132)
- Scenario 4: Failed Spawn (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L145)
- Architecture (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L15)
- Tasks / Subtasks (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L158)
- Definition of Done (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L180)
- Changes to `spawnTmuxSession` (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L19)
- Caller Updates (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L38)
- Changes to `KillAgent` (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L44)
- Technical Considerations (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L57)
- Risks & Edge Cases (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L64)
- Reference Files (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L69)
- Acceptance Criteria (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L75)
- Description (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L9)
- SpawnAgent() (frontend/wailsjs/go/main/App.js:L417)
- SpawnAgentWithCommand() (frontend/wailsjs/go/main/App.js:L421)
- SpawnTerminal() (frontend/wailsjs/go/main/App.js:L433)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [ListRepoSessions](/modules/listreposessions.md)

# Features
- no feature plan names these files
