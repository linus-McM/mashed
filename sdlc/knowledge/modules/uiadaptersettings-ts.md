---
type: Module
title: uiAdapterSettings.ts
description: "Graphify community 79: frontend/src/lib/stores/uiAdapterSettings.test.ts, frontend/src/lib/stores/uiAdapterSettings.ts, frontend/wailsjs/go/main/App.js"
resource: frontend
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: uiAdapterSettings.test, resource: frontend/src/lib/stores/uiAdapterSettings.test.ts, last_modified: "2026-04-23T13:17:59+10:00", digest: 20ac6e8fba38b85d }
  - { id: uiAdapterSettings, resource: frontend/src/lib/stores/uiAdapterSettings.ts, last_modified: "2026-04-23T11:43:31+10:00", digest: 439345ecd230d960 }
  - { id: App, resource: frontend/wailsjs/go/main/App.js, last_modified: "2026-05-07T18:18:02+10:00", digest: bdc3ffd8e98df7d8 }
---

# Files
- `frontend/src/lib/stores/uiAdapterSettings.test.ts`
- `frontend/src/lib/stores/uiAdapterSettings.ts`
- `frontend/wailsjs/go/main/App.js`

# Symbols
- uiAdapterSettings.test.ts (frontend/src/lib/stores/uiAdapterSettings.test.ts:L1)
- mocks (frontend/src/lib/stores/uiAdapterSettings.test.ts:L42)
- resetStores() (frontend/src/lib/stores/uiAdapterSettings.test.ts:L44)
- uiAdapterSettings.ts (frontend/src/lib/stores/uiAdapterSettings.ts:L1)
- refreshModels() (frontend/src/lib/stores/uiAdapterSettings.ts:L114)
- setOllamaEnabled() (frontend/src/lib/stores/uiAdapterSettings.ts:L149)
- setUntrustedExpanded() (frontend/src/lib/stores/uiAdapterSettings.ts:L214)
- uiAdapterEnabled (frontend/src/lib/stores/uiAdapterSettings.ts:L27)
- uiAdapterTimeoutMs (frontend/src/lib/stores/uiAdapterSettings.ts:L28)
- ollamaModel (frontend/src/lib/stores/uiAdapterSettings.ts:L29)
- ollamaEnabled (frontend/src/lib/stores/uiAdapterSettings.ts:L30)
- uiAdapterUntrustedExpanded (frontend/src/lib/stores/uiAdapterSettings.ts:L31)
- ollamaReachable (frontend/src/lib/stores/uiAdapterSettings.ts:L32)
- ollamaModels (frontend/src/lib/stores/uiAdapterSettings.ts:L33)
- cliModel (frontend/src/lib/stores/uiAdapterSettings.ts:L38)
- backendsAvailable (frontend/src/lib/stores/uiAdapterSettings.ts:L40)
- claudeModelList (frontend/src/lib/stores/uiAdapterSettings.ts:L41)
- routerPolicyList (frontend/src/lib/stores/uiAdapterSettings.ts:L42)
- claudeApiReachable (frontend/src/lib/stores/uiAdapterSettings.ts:L43)
- claudeCliReachable (frontend/src/lib/stores/uiAdapterSettings.ts:L44)
- BACKENDS_ENUM (frontend/src/lib/stores/uiAdapterSettings.ts:L46)
- ROUTER_POLICIES_ENUM (frontend/src/lib/stores/uiAdapterSettings.ts:L47)
- CLAUDE_MODEL_ENUM (frontend/src/lib/stores/uiAdapterSettings.ts:L55)
- refreshProbeAndModels() (frontend/src/lib/stores/uiAdapterSettings.ts:L70)
- SetOllamaEnabled() (frontend/wailsjs/go/main/App.js:L381)
- SetUIAdapterUntrustedExpanded() (frontend/wailsjs/go/main/App.js:L409)

# Depends on
- [App.js](/modules/app-js.md)
- [hydrate](/modules/hydrate.md)
- [interactiveInput.ts](/modules/interactiveinput-ts.md)
- [ListOllamaModels](/modules/listollamamodels.md)
- [Story 18 — Title-bar Dynamic UI model selector](/modules/story-18-title-bar-dynamic-ui-model-selector.md)
- [uiAdapterSettings.v3.test.ts](/modules/uiadaptersettings-v3-test-ts.md)
- [vitest](/modules/vitest.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
