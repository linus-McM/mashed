---
type: Module
title: uiadapter/client_test.go
description: "Graphify community 15: internal/uiadapter/adapter.go, internal/uiadapter/adapter_test.go, internal/uiadapter/client.go, internal/uiadapter/client_test.go, internal/uiadapter/log_test_helper_test.go, i"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: adapter_test, resource: internal/uiadapter/adapter_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 14d9868bbf707930 }
  - { id: client, resource: internal/uiadapter/client.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 07a2280966b51366 }
  - { id: client_test, resource: internal/uiadapter/client_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: ae292f9785d96f43 }
  - { id: log_test_helper_test, resource: internal/uiadapter/log_test_helper_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: e2a6968886cbeb48 }
  - { id: prompt, resource: internal/uiadapter/prompt.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 07d009f8f445be65 }
  - { id: prompt_test, resource: internal/uiadapter/prompt_test.go, last_modified: "2026-04-27T10:45:20+10:00", digest: c0e8d16fee76a549 }
  - { id: translate_e2e_test, resource: internal/uiadapter/translate_e2e_test.go, last_modified: "2026-04-26T12:39:41+10:00", digest: 45f6254d2ea3d0bc }
---

# Files
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/adapter_test.go`
- `internal/uiadapter/client.go`
- `internal/uiadapter/client_test.go`
- `internal/uiadapter/log_test_helper_test.go`
- `internal/uiadapter/prompt.go`
- `internal/uiadapter/prompt_test.go`
- `internal/uiadapter/translate_e2e_test.go`

# Symbols
- NewDefault() (internal/uiadapter/adapter.go:L130)
- Adapter (internal/uiadapter/adapter.go:L19)
- adapter_test.go (internal/uiadapter/adapter_test.go:L1)
- TestU2_AC2_Adapter_FallbackReasonTable() (internal/uiadapter/adapter_test.go:L147)
- TestU2_AC5_Adapter_Semaphore_BlockedCallSaturates() (internal/uiadapter/adapter_test.go:L278)
- ollamaChatPayload() (internal/uiadapter/adapter_test.go:L28)
- TestU2_AC7_Adapter_NetworkPinned_NoNonLocalhostReach() (internal/uiadapter/adapter_test.go:L291)
- TestU2_AC9_Adapter_DisabledProducesFallback() (internal/uiadapter/adapter_test.go:L317)
- TestU2_AC9_Adapter_DisabledNoHTTPCall() (internal/uiadapter/adapter_test.go:L333)
- TestAdapter_DeterministicRoutesThroughChatDeterministic() (internal/uiadapter/adapter_test.go:L353)
- enabledConfig() (internal/uiadapter/adapter_test.go:L37)
- TestU2_AC9_Adapter_UnreachableNotAutoDisabled() (internal/uiadapter/adapter_test.go:L393)
- newSaturatedAdapter() (internal/uiadapter/adapter_test.go:L49)
- TestU2_AC1_Adapter_Translate_NeverNil() (internal/uiadapter/adapter_test.go:L67)
- uiadapter/client.go (internal/uiadapter/client.go:L1)
- HTTPStatusError (internal/uiadapter/client.go:L29)
- .Error() (internal/uiadapter/client.go:L34)
- ClientConfig (internal/uiadapter/client.go:L41)
- NewClient() (internal/uiadapter/client.go:L54)
- chatMessage (internal/uiadapter/client.go:L61)
- chatRequest (internal/uiadapter/client.go:L66)
- chatResponse (internal/uiadapter/client.go:L74)
- tagsResponse (internal/uiadapter/client.go:L80)
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
- snapshotSchemaLogger() (internal/uiadapter/log_test_helper_test.go:L99)
- SystemPrompt() (internal/uiadapter/prompt.go:L13)
- prompt_test.go (internal/uiadapter/prompt_test.go:L1)
- TestPromptVersion_IsV2() (internal/uiadapter/prompt_test.go:L107)
- TestAdapter_SendsSystemPromptInRequest() (internal/uiadapter/prompt_test.go:L116)
- TestPrompt_Golden_Brainstorming() (internal/uiadapter/prompt_test.go:L139)
- TestPrompt_Golden_Elicitation() (internal/uiadapter/prompt_test.go:L159)
- TestPrompt_Golden_ProductBrief() (internal/uiadapter/prompt_test.go:L182)
- TestPrompt_Golden_PartyMode() (internal/uiadapter/prompt_test.go:L214)
- TestPrompt_Golden_PartyMode_CodeBlockPreserved() (internal/uiadapter/prompt_test.go:L232)
- TestPrompt_Golden_Freeform() (internal/uiadapter/prompt_test.go:L255)
- readFixture() (internal/uiadapter/prompt_test.go:L30)
- newStubbedOllama() (internal/uiadapter/prompt_test.go:L41)
- decisionGroups() (internal/uiadapter/prompt_test.go:L53)
- goldenAdapter() (internal/uiadapter/prompt_test.go:L66)
- runGolden() (internal/uiadapter/prompt_test.go:L77)
- TestSystemPrompt_ContainsAllSections() (internal/uiadapter/prompt_test.go:L88)
- TestStory6_AC5_DefaultLevelInfo_NoDebugRecords() (internal/uiadapter/translate_e2e_test.go:L397)
- e2eAdapter() (internal/uiadapter/translate_e2e_test.go:L45)

# Depends on
- [context.Context](/modules/context-context.md)
- [log/slog.Logger](/modules/log-slog-logger.md)

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)
- [logging_comprehensive_test.go](/modules/logging-comprehensive-test-go.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Features
- no feature plan names these files
