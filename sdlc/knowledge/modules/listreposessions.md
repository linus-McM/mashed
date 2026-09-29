---
type: Module
title: ListRepoSessions
description: "Graphify community 61: docs/stories/old_stories/sessions-04-svelte-sessions-store.md, docs/stories/old_stories/sessions-05-session-tab-bar.md, docs/stories/old_stories/sessions-backlog.md, frontend/sr"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: sessions-04-svelte-sessions-store, resource: docs/stories/old_stories/sessions-04-svelte-sessions-store.md, last_modified: "2026-04-09T10:03:44+10:00", digest: b0dcba8f3a115e70 }
  - { id: sessions-05-session-tab-bar, resource: docs/stories/old_stories/sessions-05-session-tab-bar.md, last_modified: "2026-04-09T10:03:44+10:00", digest: 73a79624d6dddc2a }
  - { id: sessions-backlog, resource: docs/stories/old_stories/sessions-backlog.md, last_modified: "2026-04-12T10:43:48+10:00", digest: 033f2f719354ebab }
  - { id: sessions, resource: frontend/src/lib/stores/sessions.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 0fa9462ed476c629 }
  - { id: session, resource: frontend/src/types/session.ts, last_modified: "2026-04-22T17:06:36+10:00", digest: 0b3fa196502eba06 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
---

# Files
- `docs/stories/old_stories/sessions-04-svelte-sessions-store.md`
- `docs/stories/old_stories/sessions-05-session-tab-bar.md`
- `docs/stories/old_stories/sessions-backlog.md`
- `frontend/src/lib/stores/sessions.ts`
- `frontend/src/types/session.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- sessions-04-svelte-sessions-store.md (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L1)
- Story 4: Svelte Sessions Store & Event Subscription (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L1)
- Developer Notes (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L13)
- Architecture (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L15)
- Tasks / Subtasks (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L196)
- Definition of Done (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L214)
- Store API (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L23)
- App.svelte Event Wiring (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L55)
- Technical Considerations (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L79)
- Risks & Edge Cases (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L85)
- Description (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L9)
- Reference Files (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L91)
- Acceptance Criteria (docs/stories/old_stories/sessions-04-svelte-sessions-store.md:L98)
- Styling (docs/stories/old_stories/sessions-05-session-tab-bar.md:L107)
- Technical Considerations (docs/stories/old_stories/sessions-05-session-tab-bar.md:L118)
- Risks & Edge Cases (docs/stories/old_stories/sessions-05-session-tab-bar.md:L124)
- Developer Notes (docs/stories/old_stories/sessions-05-session-tab-bar.md:L13)
- Reference Files (docs/stories/old_stories/sessions-05-session-tab-bar.md:L131)
- Architecture (docs/stories/old_stories/sessions-05-session-tab-bar.md:L15)
- AgentDetail.svelte Changes (docs/stories/old_stories/sessions-05-session-tab-bar.md:L21)
- Tasks / Subtasks (docs/stories/old_stories/sessions-05-session-tab-bar.md:L253)
- NotificationFeed.svelte Changes (docs/stories/old_stories/sessions-05-session-tab-bar.md:L87)
- Files Created/Modified (docs/stories/old_stories/sessions-backlog.md:L34)
- sessions.ts (frontend/src/lib/stores/sessions.ts:L1)
- repoSessions (frontend/src/lib/stores/sessions.ts:L11)
- refreshSessions() (frontend/src/lib/stores/sessions.ts:L17)
- addSession() (frontend/src/lib/stores/sessions.ts:L25)
- removeSession() (frontend/src/lib/stores/sessions.ts:L34)
- removeSessionByName() (frontend/src/lib/stores/sessions.ts:L42)
- session.ts (frontend/src/types/session.ts:L1)
- DataFields (frontend/src/types/session.ts:L19)
- Session (frontend/src/types/session.ts:L40)
- SessionState (frontend/src/types/session.ts:L48)
- ListRepoSessions() (frontend/wailsjs/go/main/App.js:L245)

# Depends on
- [App.js](/modules/app-js.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)

# Inferred
- [SpawnAgent](/modules/spawnagent.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
