---
type: Module
title: uiadapter/client_test.go
description: "Graphify community 31: internal/uiadapter/adapter.go, internal/uiadapter/adapter_test.go, internal/uiadapter/client.go, internal/uiadapter/client_test.go, internal/uiadapter/eval_test.go, internal/uia"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: adapter_test, resource: internal/uiadapter/adapter_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 14d9868bbf707930 }
  - { id: client, resource: internal/uiadapter/client.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 07a2280966b51366 }
  - { id: client_test, resource: internal/uiadapter/client_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: ae292f9785d96f43 }
  - { id: eval_test, resource: internal/uiadapter/eval_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 170e442f3eb11320 }
  - { id: log_test_helper_test, resource: internal/uiadapter/log_test_helper_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: e2a6968886cbeb48 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: logging_story3_sanitize_test, resource: internal/uiadapter/logging_story3_sanitize_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a4de64483e766cb6 }
  - { id: sanitize_adapter_test, resource: internal/uiadapter/sanitize_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 2951e338741da597 }
  - { id: translate_e2e_test, resource: internal/uiadapter/translate_e2e_test.go, last_modified: "2026-04-26T12:39:41+10:00", digest: 45f6254d2ea3d0bc }
---

# Files
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/adapter_test.go`
- `internal/uiadapter/client.go`
- `internal/uiadapter/client_test.go`
- `internal/uiadapter/eval_test.go`
- `internal/uiadapter/log_test_helper_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story3_sanitize_test.go`
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/translate_e2e_test.go`

# Symbols
- NewDefault() (internal/uiadapter/adapter.go:L130)
- Adapter (internal/uiadapter/adapter.go:L19)
- adapter_test.go (internal/uiadapter/adapter_test.go:L1)
- TestU2_AC2_Adapter_FallbackReasonTable() (internal/uiadapter/adapter_test.go:L147)
- TestU2_AC5_Adapter_Semaphore_BlockedCallSaturates() (internal/uiadapter/adapter_test.go:L278)
- ollamaChatPayload() (internal/uiadapter/adapter_test.go:L28)
- TestU2_AC9_Adapter_DisabledProducesFallback() (internal/uiadapter/adapter_test.go:L317)
- TestU2_AC9_Adapter_DisabledNoHTTPCall() (internal/uiadapter/adapter_test.go:L333)
- TestAdapter_DeterministicRoutesThroughChatDeterministic() (internal/uiadapter/adapter_test.go:L353)
- enabledConfig() (internal/uiadapter/adapter_test.go:L37)
- TestU2_AC9_Adapter_UnreachableNotAutoDisabled() (internal/uiadapter/adapter_test.go:L393)
- newSaturatedAdapter() (internal/uiadapter/adapter_test.go:L49)
- TestU2_AC1_Adapter_Translate_NeverNil() (internal/uiadapter/adapter_test.go:L67)
- ClientConfig (internal/uiadapter/client.go:L41)
- NewClient() (internal/uiadapter/client.go:L54)
- uiadapter/client_test.go (internal/uiadapter/client_test.go:L1)
- TestU1_AC3_Client_Chat_ContextDeadlineBeatsClientTimeout() (internal/uiadapter/client_test.go:L104)
- TestU1_AC5_Client_Chat_UnreachableSurfaceError() (internal/uiadapter/client_test.go:L122)
- TestU1_AC6_Client_Chat_ExtractsMessageContent() (internal/uiadapter/client_test.go:L132)
- TestU1_AC7_Client_ListModels_ParsesAndSorts() (internal/uiadapter/client_test.go:L144)
- TestU1_AC8_Client_ListModels_UnreachableError() (internal/uiadapter/client_test.go:L167)
- TestU1_AC9_Client_ListModels_EmptyOK() (internal/uiadapter/client_test.go:L178)
- TestU2_ClientHTTPStatusError_ErrorString() (internal/uiadapter/client_test.go:L189)
- TestClient_ChatDeterministic_SetsTemperature() (internal/uiadapter/client_test.go:L200)
- TestClient_ChatDeterministic_ReturnsContent() (internal/uiadapter/client_test.go:L237)
- withOllamaHost() (internal/uiadapter/client_test.go:L24)
- TestStory3_AC2_ClientTransportError() (internal/uiadapter/client_test.go:L322)
- newOllamaStub() (internal/uiadapter/client_test.go:L34)
- closedSocketURL() (internal/uiadapter/client_test.go:L45)
- TestU1_AC1_Client_Chat_RequestBodyShape() (internal/uiadapter/client_test.go:L54)
- TestU1_AC2_Client_LocalhostPinned() (internal/uiadapter/client_test.go:L89)
- TestEval_FullCorpus_MeetsThresholds() (internal/uiadapter/eval_test.go:L52)
- snapshotSchemaLogger() (internal/uiadapter/log_test_helper_test.go:L99)
- logging_plumbing_test.go (internal/uiadapter/logging_plumbing_test.go:L1)
- TestStory2_AC3_NewDefaultScopesWithGroup() (internal/uiadapter/logging_plumbing_test.go:L300)
- keysOf() (internal/uiadapter/logging_plumbing_test.go:L343)
- TestStory2_AC3_NewDefaultNilParentSafe() (internal/uiadapter/logging_plumbing_test.go:L353)
- TestStory2_AC5_NilSafeLogger() (internal/uiadapter/logging_plumbing_test.go:L40)
- logging_story3_sanitize_test.go (internal/uiadapter/logging_story3_sanitize_test.go:L1)
- captureWriter (internal/uiadapter/logging_story3_sanitize_test.go:L247)
- .Write() (internal/uiadapter/logging_story3_sanitize_test.go:L252)
- randomPayload() (internal/uiadapter/logging_story3_sanitize_test.go:L34)
- containsSlice() (internal/uiadapter/logging_story3_sanitize_test.go:L55)
- TestStory3_AC7_SanitizeDisciplineAcrossFiles() (internal/uiadapter/logging_story3_sanitize_test.go:L73)
- TestAdapter_DisabledAdapterIgnoresSanitize() (internal/uiadapter/sanitize_adapter_test.go:L103)
- TestStory6_AC5_DefaultLevelInfo_NoDebugRecords() (internal/uiadapter/translate_e2e_test.go:L397)
- e2eAdapter() (internal/uiadapter/translate_e2e_test.go:L45)

# Depends on
- [cache_test.go](/modules/cache-test-go.md)
- [Config](/modules/config.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [testing.T](/modules/testing-t.md)

# Inferred
- [cache_test.go](/modules/cache-test-go.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [logging_comprehensive_test.go](/modules/logging-comprehensive-test-go.md)

# Features
- no feature plan names these files
