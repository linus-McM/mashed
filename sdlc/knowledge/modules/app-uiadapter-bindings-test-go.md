---
type: Module
title: app_uiadapter_bindings_test.go
description: "Graphify community 255: app_uiadapter.go, app_uiadapter_bindings_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: app_uiadapter, resource: app_uiadapter.go, last_modified: "2026-09-29T07:07:25Z", digest: eebf36a959a4ef0e }
  - { id: app_uiadapter_bindings_test, resource: app_uiadapter_bindings_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 189277263c74d9ce }
---

# Files
- `app_uiadapter.go`
- `app_uiadapter_bindings_test.go`

# Symbols
- swapOllamaBaseURLForTest() (app_uiadapter.go:L50)
- app_uiadapter_bindings_test.go (app_uiadapter_bindings_test.go:L1)
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

# Depends on
- [loadConfig](/modules/loadconfig.md)
- [setupTestConfig](/modules/setuptestconfig.md)

# Inferred
- [loadConfig](/modules/loadconfig.md)
- [setupTestConfig](/modules/setuptestconfig.md)

# Features
- no feature plan names these files
