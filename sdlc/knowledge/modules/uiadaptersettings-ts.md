---
type: Module
title: uiAdapterSettings.ts
description: "Graphify community 79: docs/plans/IMPLEMENTATION_PLAN_v3_final.md, docs/stories/uiadapter-v3-18.md, frontend/src/lib/stores/uiAdapterSettings.test.ts, frontend/src/lib/stores/uiAdapterSettings.ts, fro"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: IMPLEMENTATION_PLAN_v3_final, resource: docs/plans/IMPLEMENTATION_PLAN_v3_final.md, last_modified: "2026-04-23T11:03:15+10:00", digest: 9c675d157ef09d5b }
  - { id: uiadapter-v3-18, resource: docs/stories/uiadapter-v3-18.md, last_modified: "2026-04-23T11:44:08+10:00", digest: 7658ad0d7f8e1474 }
  - { id: uiAdapterSettings.test, resource: frontend/src/lib/stores/uiAdapterSettings.test.ts, last_modified: "2026-04-23T13:17:59+10:00", digest: 20ac6e8fba38b85d }
  - { id: uiAdapterSettings, resource: frontend/src/lib/stores/uiAdapterSettings.ts, last_modified: "2026-04-23T11:43:31+10:00", digest: 439345ecd230d960 }
  - { id: uiAdapterSettings.v3.test, resource: frontend/src/lib/stores/uiAdapterSettings.v3.test.ts, last_modified: "2026-04-23T11:43:31+10:00", digest: 1fa2de67e8d85161 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `docs/plans/IMPLEMENTATION_PLAN_v3_final.md`
- `docs/stories/uiadapter-v3-18.md`
- `frontend/src/lib/stores/uiAdapterSettings.test.ts`
- `frontend/src/lib/stores/uiAdapterSettings.ts`
- `frontend/src/lib/stores/uiAdapterSettings.v3.test.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- Story 18 — Title-bar Dynamic UI model selector (docs/plans/IMPLEMENTATION_PLAN_v3_final.md:L664)
- uiadapter-v3-18.md (docs/stories/uiadapter-v3-18.md:L1)
- uiadapter-v3-18: Title-bar Dynamic UI model selector (docs/stories/uiadapter-v3-18.md:L1)
- Story (docs/stories/uiadapter-v3-18.md:L10)
- Tasks / Subtasks (docs/stories/uiadapter-v3-18.md:L128)
- Description (docs/stories/uiadapter-v3-18.md:L14)
- Definition of Done (docs/stories/uiadapter-v3-18.md:L147)
- Developer Notes (docs/stories/uiadapter-v3-18.md:L30)
- Acceptance Criteria (docs/stories/uiadapter-v3-18.md:L45)
- BDD Test Scenarios (docs/stories/uiadapter-v3-18.md:L65)
- uiAdapterSettings.test.ts (frontend/src/lib/stores/uiAdapterSettings.test.ts:L1)
- mocks (frontend/src/lib/stores/uiAdapterSettings.test.ts:L42)
- resetStores() (frontend/src/lib/stores/uiAdapterSettings.test.ts:L44)
- uiAdapterSettings.ts (frontend/src/lib/stores/uiAdapterSettings.ts:L1)
- refreshModels() (frontend/src/lib/stores/uiAdapterSettings.ts:L114)
- setModel() (frontend/src/lib/stores/uiAdapterSettings.ts:L138)
- setOllamaEnabled() (frontend/src/lib/stores/uiAdapterSettings.ts:L149)
- setBackend() (frontend/src/lib/stores/uiAdapterSettings.ts:L162)
- setClaudeModel() (frontend/src/lib/stores/uiAdapterSettings.ts:L175)
- setCliModel() (frontend/src/lib/stores/uiAdapterSettings.ts:L188)
- setRouterPolicy() (frontend/src/lib/stores/uiAdapterSettings.ts:L201)
- setUntrustedExpanded() (frontend/src/lib/stores/uiAdapterSettings.ts:L214)
- uiAdapterEnabled (frontend/src/lib/stores/uiAdapterSettings.ts:L27)
- uiAdapterTimeoutMs (frontend/src/lib/stores/uiAdapterSettings.ts:L28)
- ollamaModel (frontend/src/lib/stores/uiAdapterSettings.ts:L29)
- ollamaEnabled (frontend/src/lib/stores/uiAdapterSettings.ts:L30)
- uiAdapterUntrustedExpanded (frontend/src/lib/stores/uiAdapterSettings.ts:L31)
- ollamaReachable (frontend/src/lib/stores/uiAdapterSettings.ts:L32)
- ollamaModels (frontend/src/lib/stores/uiAdapterSettings.ts:L33)
- backend (frontend/src/lib/stores/uiAdapterSettings.ts:L36)
- claudeModel (frontend/src/lib/stores/uiAdapterSettings.ts:L37)
- cliModel (frontend/src/lib/stores/uiAdapterSettings.ts:L38)
- routerPolicy (frontend/src/lib/stores/uiAdapterSettings.ts:L39)
- backendsAvailable (frontend/src/lib/stores/uiAdapterSettings.ts:L40)
- claudeModelList (frontend/src/lib/stores/uiAdapterSettings.ts:L41)
- routerPolicyList (frontend/src/lib/stores/uiAdapterSettings.ts:L42)
- claudeApiReachable (frontend/src/lib/stores/uiAdapterSettings.ts:L43)
- claudeCliReachable (frontend/src/lib/stores/uiAdapterSettings.ts:L44)
- BACKENDS_ENUM (frontend/src/lib/stores/uiAdapterSettings.ts:L46)
- ROUTER_POLICIES_ENUM (frontend/src/lib/stores/uiAdapterSettings.ts:L47)
- CLAUDE_MODEL_ENUM (frontend/src/lib/stores/uiAdapterSettings.ts:L55)
- validBackend() (frontend/src/lib/stores/uiAdapterSettings.ts:L60)
- validRouterPolicy() (frontend/src/lib/stores/uiAdapterSettings.ts:L63)
- validClaudeModel() (frontend/src/lib/stores/uiAdapterSettings.ts:L66)
- refreshProbeAndModels() (frontend/src/lib/stores/uiAdapterSettings.ts:L70)
- refreshBackendOptions() (frontend/src/lib/stores/uiAdapterSettings.ts:L81)
- hydrate() (frontend/src/lib/stores/uiAdapterSettings.ts:L92)
- uiAdapterSettings.v3.test.ts (frontend/src/lib/stores/uiAdapterSettings.v3.test.ts:L1)
- mockSet (frontend/src/lib/stores/uiAdapterSettings.v3.test.ts:L5)
- ListBackendsAvailable() (frontend/wailsjs/go/main/App.js:L189)
- ListClaudeModels() (frontend/wailsjs/go/main/App.js:L213)
- ListRouterPolicies() (frontend/wailsjs/go/main/App.js:L245)
- SetBackend() (frontend/wailsjs/go/main/App.js:L345)
- SetCLIModel() (frontend/wailsjs/go/main/App.js:L349)
- SetClaudeModel() (frontend/wailsjs/go/main/App.js:L353)
- SetOllamaEnabled() (frontend/wailsjs/go/main/App.js:L381)
- SetOllamaModel() (frontend/wailsjs/go/main/App.js:L385)
- SetRouterPolicy() (frontend/wailsjs/go/main/App.js:L389)
- SetUIAdapterUntrustedExpanded() (frontend/wailsjs/go/main/App.js:L409)

# Depends on
- [App.js](/modules/app-js.md)
- [GetConfig](/modules/getconfig.md)
- [ListOllamaModels](/modules/listollamamodels.md)
- [vitest](/modules/vitest.md)

# Inferred
- [GetConfig](/modules/getconfig.md)
- [ListOllamaModels](/modules/listollamamodels.md)

# Features
- no feature plan names these files
