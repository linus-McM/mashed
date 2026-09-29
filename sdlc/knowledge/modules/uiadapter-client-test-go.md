---
type: Module
title: uiadapter/client_test.go
description: "Graphify community 31: internal/uiadapter/adapter.go, internal/uiadapter/adapter_test.go, internal/uiadapter/client.go, internal/uiadapter/client_test.go, internal/uiadapter/config_test.go, internal/u"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: adapter_test, resource: internal/uiadapter/adapter_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 14d9868bbf707930 }
  - { id: client, resource: internal/uiadapter/client.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 07a2280966b51366 }
  - { id: client_test, resource: internal/uiadapter/client_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: ae292f9785d96f43 }
  - { id: config_test, resource: internal/uiadapter/config_test.go, last_modified: "2026-04-23T11:43:31+10:00", digest: 748e88156e161a65 }
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
- `internal/uiadapter/config_test.go`
- `internal/uiadapter/log_test_helper_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story3_sanitize_test.go`
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/translate_e2e_test.go`

# Symbols
- NewDefault() (internal/uiadapter/adapter.go:L130)
- Adapter (internal/uiadapter/adapter.go:L19)
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
- TestStory3_AC1_ClientChatStartAndResponse_DebugRecords() (internal/uiadapter/client_test.go:L264)
- TestStory3_AC2_ClientTransportError() (internal/uiadapter/client_test.go:L322)
- newOllamaStub() (internal/uiadapter/client_test.go:L34)
- TestStory3_AC2_ClientHttpError() (internal/uiadapter/client_test.go:L369)
- closedSocketURL() (internal/uiadapter/client_test.go:L45)
- TestU1_AC1_Client_Chat_RequestBodyShape() (internal/uiadapter/client_test.go:L54)
- TestU1_AC2_Client_LocalhostPinned() (internal/uiadapter/client_test.go:L89)
- TestDefaultConfig_ZeroValueBootsAdapter() (internal/uiadapter/config_test.go:L59)
- snapshotSchemaLogger() (internal/uiadapter/log_test_helper_test.go:L99)
- TestStory2_AC3_NewDefaultScopesWithGroup() (internal/uiadapter/logging_plumbing_test.go:L300)
- keysOf() (internal/uiadapter/logging_plumbing_test.go:L343)
- randomPayload() (internal/uiadapter/logging_story3_sanitize_test.go:L34)
- TestStory3_AC7_SanitizeDisciplineAcrossFiles() (internal/uiadapter/logging_story3_sanitize_test.go:L73)
- TestAdapter_DisabledAdapterIgnoresSanitize() (internal/uiadapter/sanitize_adapter_test.go:L103)
- decodeOps() (internal/uiadapter/translate_e2e_test.go:L114)
- hasOpPrefix() (internal/uiadapter/translate_e2e_test.go:L136)
- TestStory6_AC2_Translate_HappyPath_AllPhasesLog() (internal/uiadapter/translate_e2e_test.go:L156)
- scanBufferForLeak() (internal/uiadapter/translate_e2e_test.go:L309)
- TestStory6_AC3_TenMessageNoLeak() (internal/uiadapter/translate_e2e_test.go:L329)
- TestStory6_AC5_DefaultLevelInfo_NoDebugRecords() (internal/uiadapter/translate_e2e_test.go:L397)
- e2eAdapter() (internal/uiadapter/translate_e2e_test.go:L45)

# Depends on
- [allowlist_test.go](/modules/allowlist-test-go.md)
- [Config](/modules/config.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)

# Inferred
- [allowlist_test.go](/modules/allowlist-test-go.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [logging_comprehensive_test.go](/modules/logging-comprehensive-test-go.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [testLogBuffer](/modules/testlogbuffer.md)
- [TestStory2_AC2_FreeFunctionsAcceptNilLogger](/modules/teststory2-ac2-freefunctionsacceptnillogger.md)

# Features
- no feature plan names these files
