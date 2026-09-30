---
type: Module
title: ListOllamaModels
description: "Graphify community 255: docs/plans/IMPLEMENTATION_PLAN_v3_final.md, docs/stories/ui-ast-U5-settings-ui.md, docs/stories/uiadapter-v3-18.md, frontend/src/lib/stores/uiAdapterSettings.ts, frontend/wails"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: IMPLEMENTATION_PLAN_v3_final, resource: docs/plans/IMPLEMENTATION_PLAN_v3_final.md, last_modified: "2026-04-23T11:03:15+10:00", digest: 9c675d157ef09d5b }
  - { id: ui-ast-U5-settings-ui, resource: docs/stories/ui-ast-U5-settings-ui.md, last_modified: "2026-04-21T21:07:33+10:00", digest: 52cf1571c97b6ab5 }
  - { id: uiadapter-v3-18, resource: docs/stories/uiadapter-v3-18.md, last_modified: "2026-04-23T11:44:08+10:00", digest: 7658ad0d7f8e1474 }
  - { id: uiAdapterSettings, resource: frontend/src/lib/stores/uiAdapterSettings.ts, last_modified: "2026-04-23T11:43:31+10:00", digest: 439345ecd230d960 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-09-30T01:21:31+10:00", digest: 5d5b17a1b5164973 }
---

# Files
- `docs/plans/IMPLEMENTATION_PLAN_v3_final.md`
- `docs/stories/ui-ast-U5-settings-ui.md`
- `docs/stories/uiadapter-v3-18.md`
- `frontend/src/lib/stores/uiAdapterSettings.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Story 18 — Title-bar Dynamic UI model selector (docs/plans/IMPLEMENTATION_PLAN_v3_final.md:L664)
- ui-ast-U5-settings-ui.md (docs/stories/ui-ast-U5-settings-ui.md:L1)
- ui-ast-U5: Settings UI — UIAdapterEnabled toggle + timeout + OllamaModel picker (docs/stories/ui-ast-U5-settings-ui.md:L1)
- Story (docs/stories/ui-ast-U5-settings-ui.md:L10)
- Description (docs/stories/ui-ast-U5-settings-ui.md:L14)
- Acceptance Criteria (docs/stories/ui-ast-U5-settings-ui.md:L164)
- Scope summary (docs/stories/ui-ast-U5-settings-ui.md:L18)
- BDD Test Scenarios (docs/stories/ui-ast-U5-settings-ui.md:L256)
- Non-goals (docs/stories/ui-ast-U5-settings-ui.md:L27)
- Tasks / Subtasks (docs/stories/ui-ast-U5-settings-ui.md:L296)
- Definition of Done (docs/stories/ui-ast-U5-settings-ui.md:L317)
- Developer Notes (docs/stories/uiadapter-v3-18.md:L30)
- setEnabled() (frontend/src/lib/stores/uiAdapterSettings.ts:L118)
- setTimeoutMs() (frontend/src/lib/stores/uiAdapterSettings.ts:L128)
- setModel() (frontend/src/lib/stores/uiAdapterSettings.ts:L138)
- ListOllamaModels() (frontend/wailsjs/go/main/App.js:L233)
- SetOllamaModel() (frontend/wailsjs/go/main/App.js:L389)
- SetUIAdapterEnabled() (frontend/wailsjs/go/main/App.js:L405)
- SetUIAdapterTimeoutMs() (frontend/wailsjs/go/main/App.js:L409)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [hydrate](/modules/hydrate.md)
- [uiAdapterSettings.ts](/modules/uiadaptersettings-ts.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
