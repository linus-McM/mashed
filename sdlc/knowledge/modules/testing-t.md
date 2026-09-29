---
type: Module
title: testing.T
description: "Graphify community 0: app_bmad_question_test.go, app_bmad_respond_input_test.go, app_shutdown_test.go, internal/agent/tokensamples.go, internal/agent/tokensamples_test.go, internal/bmad/condition.go,"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: app_bmad_question_test, resource: app_bmad_question_test.go, last_modified: "2026-04-10T14:01:16+10:00", digest: 9386c29f8c1272f8 }
  - { id: app_bmad_respond_input_test, resource: app_bmad_respond_input_test.go, last_modified: "2026-04-20T13:50:13+10:00", digest: c31156ae45a58fd1 }
  - { id: app_shutdown_test, resource: app_shutdown_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: bb2b11e788c7ac0b }
  - { id: tokensamples, resource: internal/agent/tokensamples.go, last_modified: "2026-04-11T17:08:28+10:00", digest: 281db4a1d2ab32a0 }
  - { id: tokensamples_test, resource: internal/agent/tokensamples_test.go, last_modified: "2026-04-11T17:08:28+10:00", digest: e73f3aa2bb785df4 }
  - { id: condition, resource: internal/bmad/condition.go, last_modified: "2026-04-08T17:10:27+10:00", digest: 956ffd61999ceb65 }
  - { id: condition_test, resource: internal/bmad/condition_test.go, last_modified: "2026-04-08T17:10:27+10:00", digest: 0058fbc2d3f1868b }
  - { id: executor_interactive_smoke_test, resource: internal/bmad/executor_interactive_smoke_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 75f14883e4b4821b }
  - { id: executor_multifileloader_test, resource: internal/bmad/executor_multifileloader_test.go, last_modified: "2026-04-14T19:25:01+10:00", digest: 32a16dc3fc140e25 }
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: interactive_types_test, resource: internal/bmad/interactive_types_test.go, last_modified: "2026-04-20T13:01:41+10:00", digest: 0282e550b6b93893 }
  - { id: testutil_interactive_test, resource: internal/bmad/testutil_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: f165a17bd64570bc }
  - { id: types_nodetype_test, resource: internal/bmad/types_nodetype_test.go, last_modified: "2026-04-08T17:10:27+10:00", digest: 63313f7f3a1e92c3 }
  - { id: types_test, resource: internal/bmad/types_test.go, last_modified: "2026-04-21T20:21:32+10:00", digest: f7fe93cfdff41ed5 }
  - { id: sessions_test, resource: internal/scanner/sessions_test.go, last_modified: "2026-04-07T10:03:32+10:00", digest: a97c8a08b2345a0c }
  - { id: adapter_test, resource: internal/uiadapter/adapter_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 14d9868bbf707930 }
  - { id: allowlist_test, resource: internal/uiadapter/allowlist_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 865e6c967ac886df }
  - { id: logging_handler_test, resource: internal/uiadapter/logging_handler_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: 998f6bb512a268ee }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: sanitize_adapter_test, resource: internal/uiadapter/sanitize_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 2951e338741da597 }
  - { id: schema_test, resource: internal/uiadapter/schema_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 706f4f8efc89d9d5 }
  - { id: schemas_test, resource: internal/uiadapter/schemas_test.go, last_modified: "2026-04-23T11:04:07+10:00", digest: e7499deb809d563b }
  - { id: stages_test, resource: internal/uiadapter/stages_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 74ccbf42bffba71e }
  - { id: readfilebase64_test, resource: readfilebase64_test.go, last_modified: "2026-04-09T11:58:53+10:00", digest: 82165929a5efaa1e }
  - { id: screenshot_fullstack_test, resource: screenshot_fullstack_test.go, last_modified: "2026-04-21T10:01:10+10:00", digest: da72543a3716b145 }
---

# Files
- `app_bmad_question_test.go`
- `app_bmad_respond_input_test.go`
- `app_shutdown_test.go`
- `internal/agent/tokensamples.go`
- `internal/agent/tokensamples_test.go`
- `internal/bmad/condition.go`
- `internal/bmad/condition_test.go`
- `internal/bmad/executor_interactive_smoke_test.go`
- `internal/bmad/executor_multifileloader_test.go`
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/interactive_types_test.go`
- `internal/bmad/testutil_interactive_test.go`
- `internal/bmad/types_nodetype_test.go`
- `internal/bmad/types_test.go`
- `internal/scanner/sessions_test.go`
- `internal/uiadapter/adapter_test.go`
- `internal/uiadapter/allowlist_test.go`
- `internal/uiadapter/logging_handler_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/schema_test.go`
- `internal/uiadapter/schemas_test.go`
- `internal/uiadapter/stages_test.go`
- `readfilebase64_test.go`
- `screenshot_fullstack_test.go`

# Symbols
- TestStory2_AC6_AppRespondToQuestion_NilExecutor() (app_bmad_question_test.go:L59)
- TestAppRespondToInputNilExecutor() (app_bmad_respond_input_test.go:L44)
- TestAppRespondToInputSignature() (app_bmad_respond_input_test.go:L62)
- TestStory1_AC6_AppShutdownDrainsHooks() (app_shutdown_test.go:L28)
- tokensamples.go (internal/agent/tokensamples.go:L1)
- MaybeAppendTokenSample() (internal/agent/tokensamples.go:L24)
- tokensamples_test.go (internal/agent/tokensamples_test.go:L1)
- TestNotificationEventJSON_IncludesTokenSamples() (internal/agent/tokensamples_test.go:L112)
- TestNotificationEventJSON_OmitEmptyTokenSamples() (internal/agent/tokensamples_test.go:L129)
- TestMaybeAppendTokenSample_CapEnforced() (internal/agent/tokensamples_test.go:L15)
- TestMaybeAppendTokenSample_FirstSampleAlwaysAppends() (internal/agent/tokensamples_test.go:L34)
- TestMaybeAppendTokenSample_DeltaThrottle() (internal/agent/tokensamples_test.go:L41)
- TestMaybeAppendTokenSample_DeltaBypassedForDecreases() (internal/agent/tokensamples_test.go:L61)
- TestMaybeAppendTokenSample_ThreadSafeRace() (internal/agent/tokensamples_test.go:L70)
- TestAgentJSON_IncludesTokenSamples() (internal/agent/tokensamples_test.go:L97)
- ParseCondition() (internal/bmad/condition.go:L91)
- condition_test.go (internal/bmad/condition_test.go:L1)
- TestCondition_Evaluate_ExitCode() (internal/bmad/condition_test.go:L101)
- TestCondition_Evaluate_Contains() (internal/bmad/condition_test.go:L12)
- TestCondition_Evaluate_FileExists() (internal/bmad/condition_test.go:L133)
- TestCondition_Evaluate_Always() (internal/bmad/condition_test.go:L166)
- TestCondition_Evaluate_MissingSourceNode() (internal/bmad/condition_test.go:L187)
- TestCondition_Evaluate_UnknownType() (internal/bmad/condition_test.go:L224)
- TestParseCondition_Valid() (internal/bmad/condition_test.go:L234)
- TestParseCondition_Invalid() (internal/bmad/condition_test.go:L290)
- TestCondition_Evaluate_NilOutputs() (internal/bmad/condition_test.go:L311)
- TestCondition_Evaluate_NotContains() (internal/bmad/condition_test.go:L43)
- TestCondition_Evaluate_Regex() (internal/bmad/condition_test.go:L70)
- executor_interactive_smoke_test.go (internal/bmad/executor_interactive_smoke_test.go:L1)
- TestSmoke_PartyMode_ExitToken() (internal/bmad/executor_interactive_smoke_test.go:L114)
- TestSmoke_AdvancedElicitation_XAccept() (internal/bmad/executor_interactive_smoke_test.go:L142)
- TestSmoke_Brainstorming_ThreeRoundsDone() (internal/bmad/executor_interactive_smoke_test.go:L17)
- TestSmoke_Brainstorming_RejectAbort() (internal/bmad/executor_interactive_smoke_test.go:L49)
- TestSmoke_ProductBrief_GuidedApproval() (internal/bmad/executor_interactive_smoke_test.go:L82)
- TestMultiFileLoader_AC1_NodeTypeRoundTrip() (internal/bmad/executor_multifileloader_test.go:L64)
- TestAC6_LegacyJSON_RoundTrip_NoStaleOutputPaths() (internal/bmad/executor_outputpaths_test.go:L247)
- TestNodeArtifactEvent_PathsField_Serializes() (internal/bmad/executor_outputpaths_test.go:L264)
- interactive_types_test.go (internal/bmad/interactive_types_test.go:L1)
- TestIterationGateNilVsPopulated() (internal/bmad/interactive_types_test.go:L259)
- TestNodeAwaitingInputJSONRoundTrip() (internal/bmad/interactive_types_test.go:L29)
- TestWorkflowExecutionInteractiveState() (internal/bmad/interactive_types_test.go:L357)
- TestPendingPromptRoundTrip() (internal/bmad/interactive_types_test.go:L417)
- TestOutputSpecRoundTrip() (internal/bmad/interactive_types_test.go:L503)
- TestProcessDefLegacyShape() (internal/bmad/interactive_types_test.go:L53)
- TestProcessDefInteractiveFieldsRoundTrip() (internal/bmad/interactive_types_test.go:L570)
- TestNodeInputEntryRoundTrip() (internal/bmad/interactive_types_test.go:L623)
- TestWorkflowExecutionLegacyShape() (internal/bmad/interactive_types_test.go:L652)
- TestInputSpecShapes() (internal/bmad/interactive_types_test.go:L92)
- newInteractiveHarness() (internal/bmad/testutil_interactive_test.go:L70)
- types_nodetype_test.go (internal/bmad/types_nodetype_test.go:L1)
- TestNodeTypeConstants() (internal/bmad/types_nodetype_test.go:L11)
- TestWorkflowExecution_NodeOutputs_JSONRoundTrip() (internal/bmad/types_nodetype_test.go:L126)
- TestWorkflowNode_EffectiveType() (internal/bmad/types_nodetype_test.go:L28)
- TestWorkflowNode_NodeType_JSONRoundTrip() (internal/bmad/types_nodetype_test.go:L50)
- TestWorkflowEdge_Handles_JSONRoundTrip() (internal/bmad/types_nodetype_test.go:L93)
- types_test.go (internal/bmad/types_test.go:L1)
- TestAC4_ArtifactTypeConstants() (internal/bmad/types_test.go:L13)
- TestU0_AC1_ProcessDef_EnableAstAdapter_RoundTrip() (internal/bmad/types_test.go:L169)
- TestPendingPrompt_Structured_RoundTrip() (internal/bmad/types_test.go:L216)
- TestNodeInputEntry_RoundAndKey_RoundTrip() (internal/bmad/types_test.go:L279)
- TestAC4_ArtifactSpec_JSONRoundTrip() (internal/bmad/types_test.go:L32)
- TestAC4_ArtifactSpec_ZeroValue() (internal/bmad/types_test.go:L62)
- TestStory1_AC4_LegacyWorkflowRoundTrip() (internal/bmad/types_test.go:L78)
- sessions_test.go (internal/scanner/sessions_test.go:L1)
- TestParseSession_SubAgentDetection() (internal/scanner/sessions_test.go:L102)
- writeTempJSONL() (internal/scanner/sessions_test.go:L12)
- TestParseSession_MalformedJSONLSkipped() (internal/scanner/sessions_test.go:L137)
- TestParseSession_EmptyFile() (internal/scanner/sessions_test.go:L160)
- TestParseSession_UnknownToolName() (internal/scanner/sessions_test.go:L180)
- newTestProvider() (internal/scanner/sessions_test.go:L23)
- TestParseSession_AssistantTokenAccumulation() (internal/scanner/sessions_test.go:L32)
- TestParseSession_ToolUseLogKinds() (internal/scanner/sessions_test.go:L62)
- TestU2_AC7_Adapter_NetworkPinned_NoNonLocalhostReach() (internal/uiadapter/adapter_test.go:L291)
- TestStory5_AC6_AllowlistDefaultRemoved() (internal/uiadapter/allowlist_test.go:L93)
- TestNoopCloser_AlwaysNil() (internal/uiadapter/logging_handler_test.go:L118)
- TestFanoutHandler_WithAttrs() (internal/uiadapter/logging_handler_test.go:L39)
- TestFanoutHandler_WithGroup() (internal/uiadapter/logging_handler_test.go:L60)
- TestFanoutHandler_EnabledShortCircuit() (internal/uiadapter/logging_handler_test.go:L83)
- TestFileCloser_ConcurrentDoubleClose() (internal/uiadapter/logging_handler_test.go:L99)
- TestStory2_AC6_NoNewLogCallsInProduction() (internal/uiadapter/logging_plumbing_test.go:L389)
- TestAdapter_Translate_LogsSanitizeDelta() (internal/uiadapter/sanitize_adapter_test.go:L50)
- TestAdapter_Translate_SanitizeZeroDeltaStillLogged() (internal/uiadapter/sanitize_adapter_test.go:L74)
- schema_test.go (internal/uiadapter/schema_test.go:L1)
- TestU2_Schema_UINode_Table_RoundTrip() (internal/uiadapter/schema_test.go:L124)
- TestU2_Schema_UIAST_RoundTrip() (internal/uiadapter/schema_test.go:L14)
- TestU2_Schema_UINode_Summary_RoundTrip() (internal/uiadapter/schema_test.go:L144)
- TestU2_Schema_UINode_Hint_RoundTrip() (internal/uiadapter/schema_test.go:L157)
- TestU2_Schema_UINode_Code_RoundTrip() (internal/uiadapter/schema_test.go:L166)
- TestU2_Schema_UINode_OptionalFieldsOmitted() (internal/uiadapter/schema_test.go:L178)
- TestStory5_AC7_NoLoggerFallback() (internal/uiadapter/schema_test.go:L251)
- TestU2_Schema_WidgetOption_RoundTrip() (internal/uiadapter/schema_test.go:L269)
- TestU2_Schema_Diagnostics_CancelReasonNotSerialized() (internal/uiadapter/schema_test.go:L42)
- TestU2_Schema_WidgetNode_RoundTrip() (internal/uiadapter/schema_test.go:L56)
- TestU2_Schema_UINode_DecisionGroup_RoundTrip() (internal/uiadapter/schema_test.go:L93)
- schemas_test.go (internal/uiadapter/schemas_test.go:L1)
- TestUIASTSchemas_ValidJSONSchema() (internal/uiadapter/schemas_test.go:L101)
- TestSchemas_AC_A1_DirectoryExists() (internal/uiadapter/schemas_test.go:L159)
- TestSchemas_AC_A1_UIASTFileExistsAndIsJSON() (internal/uiadapter/schemas_test.go:L167)
- TestSchemas_AC_A1_UIASTDeclaresDraftSchema() (internal/uiadapter/schemas_test.go:L175)
- TestSchemas_AC_A1_UIASTDescribesEnvelope() (internal/uiadapter/schemas_test.go:L190)
- TestSchemas_AC_A1_UIASTEnvelopeRequired() (internal/uiadapter/schemas_test.go:L209)
- TestSchemas_AC_A1_UIASTRejectsAdditionalProperties() (internal/uiadapter/schemas_test.go:L234)
- TestSchemas_AC_A1_PerKindShardsExist() (internal/uiadapter/schemas_test.go:L248)
- TestSchemas_AC_A1_PerKindDeclareDraftSchema() (internal/uiadapter/schemas_test.go:L261)
- TestSchemas_AC_A1_PerKindSubsetsOfUIAST() (internal/uiadapter/schemas_test.go:L279)
- TestSchemas_AC_A1_PerKindRejectsAdditionalProperties() (internal/uiadapter/schemas_test.go:L304)
- collectSchemaPropertyKeys() (internal/uiadapter/schemas_test.go:L53)
- readSchema() (internal/uiadapter/schemas_test.go:L86)
- TestPromptsBudget() (internal/uiadapter/stages_test.go:L69)
- TestReadFileBase64_AC4_MissingFile() (readfilebase64_test.go:L102)
- TestReadFileBase64_AC5_EmptyPath() (readfilebase64_test.go:L116)
- TestReadFileBase64_AC1_Base64Encoding() (readfilebase64_test.go:L15)
- TestReadFileBase64_AC3_FileSizeLimit() (readfilebase64_test.go:L86)
- TestTakeScreenshot_AC5_EmptyRepoPath() (screenshot_fullstack_test.go:L135)
- TestSetActiveContext_AC3_SetAndRead() (screenshot_fullstack_test.go:L17)
- TestSetActiveContext_AC3_ConcurrentSafety() (screenshot_fullstack_test.go:L27)
- TestSetActiveContext_AC4_Clear() (screenshot_fullstack_test.go:L47)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [Condition](/modules/condition.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [interactiveHarness](/modules/interactiveharness.md)
- [NotificationEngine](/modules/notificationengine.md)

# Inferred
- [bmad/types.go](/modules/bmad-types-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [newHarness](/modules/newharness.md)

# Features
- no feature plan names these files
