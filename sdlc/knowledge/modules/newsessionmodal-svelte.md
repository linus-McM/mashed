---
type: Module
title: NewSessionModal.svelte
description: "Graphify community 258: frontend/src/__tests__/status-token.test.ts, frontend/src/components/SparkLine.svelte, frontend/src/components/StatusBadge.svelte, frontend/src/lib/ptySize.ts, frontend/src/lib"
resource: frontend/src
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: status-token.test, resource: frontend/src/__tests__/status-token.test.ts, last_modified: "2026-04-22T18:25:48+10:00", digest: 1a2f08d41d9861de }
  - { id: SparkLine, resource: frontend/src/components/SparkLine.svelte, last_modified: "2026-04-22T19:23:56+10:00", digest: 61bc0e7ea260e7dc }
  - { id: StatusBadge, resource: frontend/src/components/StatusBadge.svelte, last_modified: "2026-04-22T18:25:48+10:00", digest: adf6af8794b8416c }
  - { id: ptySize, resource: frontend/src/lib/ptySize.ts, last_modified: "2026-05-07T18:18:02+10:00", digest: 819faf79772ff83a }
  - { id: repoPalette, resource: frontend/src/lib/repoPalette.ts, last_modified: "2026-04-22T18:25:48+10:00", digest: 513bf749080da8bf }
  - { id: status, resource: frontend/src/types/status.ts, last_modified: "2026-04-22T18:25:48+10:00", digest: 396d5990d2eddad1 }
  - { id: NewSessionModal, resource: frontend/src/views/NewSessionModal.svelte, last_modified: "2026-05-07T11:13:58+10:00", digest: 800ee3ff24898e95 }
  - { id: NotificationFeed, resource: frontend/src/views/NotificationFeed.svelte, last_modified: "2026-05-07T18:18:02+10:00", digest: 06607e88326a1ff7 }
---

# Files
- `frontend/src/__tests__/status-token.test.ts`
- `frontend/src/components/SparkLine.svelte`
- `frontend/src/components/StatusBadge.svelte`
- `frontend/src/lib/ptySize.ts`
- `frontend/src/lib/repoPalette.ts`
- `frontend/src/types/status.ts`
- `frontend/src/views/NewSessionModal.svelte`
- `frontend/src/views/NotificationFeed.svelte`

# Symbols
- status-token.test.ts (frontend/src/__tests__/status-token.test.ts:L1)
- SparkLine.svelte (frontend/src/components/SparkLine.svelte:L1)
- StatusBadge.svelte (frontend/src/components/StatusBadge.svelte:L1)
- ptySize.ts (frontend/src/lib/ptySize.ts:L1)
- estimatePtySize() (frontend/src/lib/ptySize.ts:L17)
- repoPalette.ts (frontend/src/lib/repoPalette.ts:L1)
- REPO_BORDER_NONE (frontend/src/lib/repoPalette.ts:L18)
- REPO_BORDER_PALETTE (frontend/src/lib/repoPalette.ts:L28)
- status.ts (frontend/src/types/status.ts:L1)
- AgentStatusToken (frontend/src/types/status.ts:L31)
- EventTypeToken (frontend/src/types/status.ts:L46)
- UITerminalToken (frontend/src/types/status.ts:L54)
- StatusToken (frontend/src/types/status.ts:L60)
- STATUS_TOKENS (frontend/src/types/status.ts:L69)
- isStatusToken() (frontend/src/types/status.ts:L93)
- NewSessionModal.svelte (frontend/src/views/NewSessionModal.svelte:L1)
- buildCommand() (frontend/src/views/NewSessionModal.svelte:L122)
- spawn() (frontend/src/views/NewSessionModal.svelte:L188)
- cancel() (frontend/src/views/NewSessionModal.svelte:L192)
- handleKeydown() (frontend/src/views/NewSessionModal.svelte:L197)
- repoPath (frontend/src/views/NewSessionModal.svelte:L26)
- repoName (frontend/src/views/NewSessionModal.svelte:L27)
- toggles (frontend/src/views/NewSessionModal.svelte:L37)
- selected (frontend/src/views/NewSessionModal.svelte:L435)
- conditionalEnabled (frontend/src/views/NewSessionModal.svelte:L44)
- conditionalValues (frontend/src/views/NewSessionModal.svelte:L46)
- textValues (frontend/src/views/NewSessionModal.svelte:L54)
- openSections (frontend/src/views/NewSessionModal.svelte:L61)
- toggleSection() (frontend/src/views/NewSessionModal.svelte:L70)
- NotificationFeed.svelte (frontend/src/views/NotificationFeed.svelte:L1)
- if() (frontend/src/views/NotificationFeed.svelte:L377)

# Depends on
- [App.js](/modules/app-js.md)
- [bmadEvents.ts](/modules/bmadevents-ts.md)
- [runtime.js](/modules/runtime-js.md)
- [svelte](/modules/svelte.md)
- [vitest](/modules/vitest.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
