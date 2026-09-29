---
type: Module
title: app_uiadapter_bindings_test.go
description: "Graphify community 185: app.go, app_uiadapter.go, app_uiadapter_bindings_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_uiadapter, resource: app_uiadapter.go, last_modified: "2026-04-21T21:06:39+10:00", digest: eebf36a959a4ef0e }
  - { id: app_uiadapter_bindings_test, resource: app_uiadapter_bindings_test.go, last_modified: "2026-04-21T21:06:39+10:00", digest: 189277263c74d9ce }
---

# Files
- `app.go`
- `app_uiadapter.go`
- `app_uiadapter_bindings_test.go`

# Symbols
- defaultConfig() (app.go:L203)
- swapOllamaBaseURLForTest() (app_uiadapter.go:L50)
- app_uiadapter_bindings_test.go (app_uiadapter_bindings_test.go:L1)
- TestU5_AC3_App_SetOllamaModel_ValidatesName() (app_uiadapter_bindings_test.go:L111)
- TestU5_AC4_App_ProbeOllamaReachable_DeadSocket() (app_uiadapter_bindings_test.go:L153)
- TestU5_AC4_App_ProbeOllamaReachable_LiveServer() (app_uiadapter_bindings_test.go:L166)
- TestU5_AC4_App_ProbeOllamaReachable_Non2xx() (app_uiadapter_bindings_test.go:L180)
- TestU5_AC8_MashedConfig_UntrustedExpanded_DefaultFalse() (app_uiadapter_bindings_test.go:L194)
- TestU5_AC8_MashedConfig_UntrustedExpanded_ExplicitTrue_Honored() (app_uiadapter_bindings_test.go:L202)
- withBindingsBaseURL() (app_uiadapter_bindings_test.go:L23)
- TestU5_App_ListOllamaModels_ReturnsSortedList() (app_uiadapter_bindings_test.go:L241)
- TestU5_App_ListOllamaModels_EmptyOK() (app_uiadapter_bindings_test.go:L258)
- TestU5_App_ListOllamaModels_WrapsErrOllamaUnreachable() (app_uiadapter_bindings_test.go:L272)
- TestU5_ErrInvalidConfig_IsDistinctSentinel() (app_uiadapter_bindings_test.go:L284)
- closedSocketURL() (app_uiadapter_bindings_test.go:L30)
- TestU5_AC1_App_SetUIAdapterEnabled_RoundTripBothDirections() (app_uiadapter_bindings_test.go:L56)
- TestU5_AC2_App_SetUIAdapterTimeoutMs_BoundsCheck() (app_uiadapter_bindings_test.go:L71)

# Depends on
- [App](/modules/app.md)
- [markdown_menu_test.go](/modules/markdown-menu-test-go.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)
- [setupTestConfig](/modules/setuptestconfig.md)

# Features
- no feature plan names these files
