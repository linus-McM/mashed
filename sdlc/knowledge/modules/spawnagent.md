---
type: Module
title: SpawnAgent
description: "Graphify community 69: docs/stories/old_stories/bmad-05-wails-bindings.md, docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md, docs/stories/old_stories/pty-04-app-spawn-integratio"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: bmad-05-wails-bindings, resource: docs/stories/old_stories/bmad-05-wails-bindings.md, last_modified: "2026-04-08T10:23:03+10:00", digest: d868ae2882f2c29b }
  - { id: pty-04-app-integration-and-frontend-cleanup, resource: docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md, last_modified: "2026-04-12T10:43:48+10:00", digest: f0f69a09af0aa8d1 }
  - { id: pty-04-app-spawn-integration, resource: docs/stories/old_stories/pty-04-app-spawn-integration.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 43787ada3d423bb2 }
  - { id: ptyhelper-06-main-integration, resource: docs/stories/old_stories/ptyhelper-06-main-integration.md, last_modified: "2026-04-12T10:43:48+10:00", digest: c229fa5e3bd21db3 }
  - { id: sessions-03-spawn-registration-events, resource: docs/stories/old_stories/sessions-03-spawn-registration-events.md, last_modified: "2026-04-09T10:03:44+10:00", digest: d396b884ad9c18b7 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
---

# Files
- `docs/stories/old_stories/bmad-05-wails-bindings.md`
- `docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md`
- `docs/stories/old_stories/pty-04-app-spawn-integration.md`
- `docs/stories/old_stories/ptyhelper-06-main-integration.md`
- `docs/stories/old_stories/sessions-03-spawn-registration-events.md`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Reference Files (docs/stories/old_stories/bmad-05-wails-bindings.md:L133)
- Backend: App Struct & Spawn Integration (docs/stories/old_stories/pty-04-app-integration-and-frontend-cleanup.md:L21)
- Architecture (docs/stories/old_stories/pty-04-app-spawn-integration.md:L15)
- Graceful Degradation (docs/stories/old_stories/ptyhelper-06-main-integration.md:L77)
- sessions-03-spawn-registration-events.md (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L1)
- Story 3: Register Sessions on Spawn & Emit Wails Events (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L1)
- Developer Notes (docs/stories/old_stories/sessions-03-spawn-registration-events.md:L13)
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
- SpawnAgent() (frontend/wailsjs/go/main/App.js:L421)
- SpawnAgentWithCommand() (frontend/wailsjs/go/main/App.js:L425)
- SpawnTerminal() (frontend/wailsjs/go/main/App.js:L437)
- GetNotifications() (frontend/wailsjs/go/main/App.js:L97)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [Developer Notes](/modules/developer-notes.md)
- [EventsEmit](/modules/eventsemit.md)
- [ListRepoSessions](/modules/listreposessions.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
