---
type: Module
title: "PR 1: Security (branch `repo-health/pr1-security`)"
description: "Graphify community 349: frontend/src/lib/stores/theme.js, frontend/src/lib/themeInit.js, frontend/src/lib/workflowSerialisation.ts, frontend/wailsjs/go/main/App.js, sdlc/repo-health-remediation/plan.m"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: theme, resource: frontend/src/lib/stores/theme.js, last_modified: "2026-04-22T19:23:56+10:00", digest: f7e61260263e6ac6 }
  - { id: themeInit, resource: frontend/src/lib/themeInit.js, last_modified: "2026-04-22T19:23:56+10:00", digest: 30161049c3a59f8f }
  - { id: workflowSerialisation, resource: frontend/src/lib/workflowSerialisation.ts, last_modified: "2026-04-28T12:36:05+10:00", digest: 36a9064bc7d2060a }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
  - { id: plan, resource: sdlc/repo-health-remediation/plan.md, last_modified: "2026-09-30T06:55:02+10:00", digest: 368f30486e86ea00 }
---

# Files
- `frontend/src/lib/stores/theme.js`
- `frontend/src/lib/themeInit.js`
- `frontend/src/lib/workflowSerialisation.ts`
- `frontend/wailsjs/go/main/App.js`
- `sdlc/repo-health-remediation/plan.md`

# Symbols
- unregisterImportedTheme() (frontend/src/lib/stores/theme.js:L62)
- removeImportedTheme() (frontend/src/lib/themeInit.js:L86)
- workflowNodesToCanvasNodes() (frontend/src/lib/workflowSerialisation.ts:L209)
- GetTerminalAuth() (frontend/wailsjs/go/main/App.js:L113)
- RemoveTheme() (frontend/wailsjs/go/main/App.js:L305)
- plan.md (sdlc/repo-health-remediation/plan.md:L1)
- Plan: Repo health remediation (sdlc/repo-health-remediation/plan.md:L1)
- Files that change (sdlc/repo-health-remediation/plan.md:L4)
- Order of work (sdlc/repo-health-remediation/plan.md:L416)
- PR 1: Security (branch `repo-health/pr1-security`) (sdlc/repo-health-remediation/plan.md:L428)
- PR 3: Build and test health (branch `repo-health/pr3-build`) (sdlc/repo-health-remediation/plan.md:L520)
- PR 4: Repo hygiene (branch `repo-health/pr4-hygiene`) (sdlc/repo-health-remediation/plan.md:L541)
- PR 5: Structure, no behaviour change (branch `repo-health/pr5-structure`) (sdlc/repo-health-remediation/plan.md:L565)
- Risks (sdlc/repo-health-remediation/plan.md:L630)
- Proof (sdlc/repo-health-remediation/plan.md:L669)

# Depends on
- [applyTheme](/modules/applytheme.md)
- [MonacoEditor.svelte](/modules/monacoeditor-svelte.md)
- [SpawnRefactorPlan](/modules/spawnrefactorplan.md)

# Inferred
- [Developer Notes](/modules/developer-notes.md)
- [SetTheme](/modules/settheme.md)
- [themeInit.js](/modules/themeinit-js.md)
- [WriteFile](/modules/writefile.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
