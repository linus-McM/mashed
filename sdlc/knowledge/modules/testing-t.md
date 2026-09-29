---
type: Module
title: testing.T
description: "Graphify community 0: app_bmad_question_test.go, app_bmad_respond_input_test.go, internal/bmad/condition.go, internal/bmad/condition_test.go, internal/bmad/executor_interactive_smoke_test.go, internal"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: app_bmad_question_test, resource: app_bmad_question_test.go, last_modified: "2026-04-10T14:01:16+10:00", digest: 9386c29f8c1272f8 }
  - { id: app_bmad_respond_input_test, resource: app_bmad_respond_input_test.go, last_modified: "2026-04-20T13:50:13+10:00", digest: c31156ae45a58fd1 }
  - { id: condition, resource: internal/bmad/condition.go, last_modified: "2026-04-08T17:10:27+10:00", digest: 956ffd61999ceb65 }
  - { id: condition_test, resource: internal/bmad/condition_test.go, last_modified: "2026-04-08T17:10:27+10:00", digest: 0058fbc2d3f1868b }
  - { id: executor_interactive_smoke_test, resource: internal/bmad/executor_interactive_smoke_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 75f14883e4b4821b }
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: interactive_types_test, resource: internal/bmad/interactive_types_test.go, last_modified: "2026-04-20T13:01:41+10:00", digest: 0282e550b6b93893 }
  - { id: storage_test, resource: internal/bmad/storage_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: cbabd36eb06877ea }
  - { id: testutil_interactive_test, resource: internal/bmad/testutil_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: f165a17bd64570bc }
  - { id: types_nodetype_test, resource: internal/bmad/types_nodetype_test.go, last_modified: "2026-04-08T17:10:27+10:00", digest: 63313f7f3a1e92c3 }
  - { id: types_test, resource: internal/bmad/types_test.go, last_modified: "2026-04-21T20:21:32+10:00", digest: f7fe93cfdff41ed5 }
  - { id: adapter_test, resource: internal/uiadapter/adapter_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 14d9868bbf707930 }
  - { id: eval_test, resource: internal/uiadapter/eval_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: 170e442f3eb11320 }
  - { id: logging_handler_test, resource: internal/uiadapter/logging_handler_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: 998f6bb512a268ee }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: mock_test, resource: internal/uiadapter/mock_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: e6b7941070a1bd3a }
  - { id: sanitize_adapter_test, resource: internal/uiadapter/sanitize_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 2951e338741da597 }
  - { id: schema_test, resource: internal/uiadapter/schema_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 706f4f8efc89d9d5 }
  - { id: testhelper_eval, resource: internal/uiadapter/testhelper_eval.go, last_modified: "2026-04-22T14:13:03+10:00", digest: d5eeb86f7290c4bf }
---

# Files
- `app_bmad_question_test.go`
- `app_bmad_respond_input_test.go`
- `internal/bmad/condition.go`
- `internal/bmad/condition_test.go`
- `internal/bmad/executor_interactive_smoke_test.go`
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/interactive_types_test.go`
- `internal/bmad/storage_test.go`
- `internal/bmad/testutil_interactive_test.go`
- `internal/bmad/types_nodetype_test.go`
- `internal/bmad/types_test.go`
- `internal/uiadapter/adapter_test.go`
- `internal/uiadapter/eval_test.go`
- `internal/uiadapter/logging_handler_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/mock_test.go`
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/schema_test.go`
- `internal/uiadapter/testhelper_eval.go`

# Symbols
- newBmadTestApp() (app_bmad_question_test.go:L25)
- TestStory2_AC5_AppRespondToQuestion_Delegates() (app_bmad_question_test.go:L46)
- app_bmad_respond_input_test.go (app_bmad_respond_input_test.go:L1)
- TestAppRespondToInputIntegration() (app_bmad_respond_input_test.go:L101)
- TestAppRespondToInputProxies() (app_bmad_respond_input_test.go:L28)
- TestAppRespondToInputNilExecutor() (app_bmad_respond_input_test.go:L44)
- TestAppRespondToInputSignature() (app_bmad_respond_input_test.go:L62)
- TestAppRespondToQuestionLegacyShimStillWorks() (app_bmad_respond_input_test.go:L80)
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
- storage_test.go (internal/bmad/storage_test.go:L1)
- TestDeleteWorkflow_NotFound() (internal/bmad/storage_test.go:L105)
- TestSaveWorkflow_PathTraversal() (internal/bmad/storage_test.go:L113)
- TestLoadWorkflow_InvalidID() (internal/bmad/storage_test.go:L135)
- newTestStorage() (internal/bmad/storage_test.go:L14)
- TestDeleteWorkflow_InvalidID() (internal/bmad/storage_test.go:L141)
- TestListWorkflows_SkipsMalformedJSON() (internal/bmad/storage_test.go:L149)
- TestListWorkflows_SkipsNonJSON() (internal/bmad/storage_test.go:L163)
- TestSaveAndListAgents() (internal/bmad/storage_test.go:L177)
- TestDeleteAgent() (internal/bmad/storage_test.go:L188)
- TestDeleteAgent_NotFound() (internal/bmad/storage_test.go:L199)
- TestSaveAgent_InvalidID() (internal/bmad/storage_test.go:L205)
- sampleWorkflow() (internal/bmad/storage_test.go:L21)
- TestDeleteAgent_InvalidID() (internal/bmad/storage_test.go:L213)
- TestNewStorage_CreatesDirectories() (internal/bmad/storage_test.go:L221)
- sampleWorkflowWithRepo() (internal/bmad/storage_test.go:L239)
- TestListWorkflowsByRepo() (internal/bmad/storage_test.go:L247)
- TestListWorkflowsByRepo_ListWorkflowsStillReturnsAll() (internal/bmad/storage_test.go:L316)
- TestWorkflowDef_RepoPathSerialization() (internal/bmad/storage_test.go:L329)
- TestSaveWorkflow_Overwrite() (internal/bmad/storage_test.go:L378)
- sampleAgent() (internal/bmad/storage_test.go:L53)
- TestSaveAndLoadWorkflow() (internal/bmad/storage_test.go:L67)
- TestLoadWorkflow_NotFound() (internal/bmad/storage_test.go:L78)
- TestListWorkflows() (internal/bmad/storage_test.go:L84)
- TestDeleteWorkflow() (internal/bmad/storage_test.go:L95)
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
- TestU2_AC7_Adapter_NetworkPinned_NoNonLocalhostReach() (internal/uiadapter/adapter_test.go:L291)
- TestEval_SkipWhenUnreachable() (internal/uiadapter/eval_test.go:L29)
- TestNoopCloser_AlwaysNil() (internal/uiadapter/logging_handler_test.go:L118)
- TestFanoutHandler_WithAttrs() (internal/uiadapter/logging_handler_test.go:L39)
- TestFanoutHandler_WithGroup() (internal/uiadapter/logging_handler_test.go:L60)
- TestFanoutHandler_EnabledShortCircuit() (internal/uiadapter/logging_handler_test.go:L83)
- TestFileCloser_ConcurrentDoubleClose() (internal/uiadapter/logging_handler_test.go:L99)
- TestStory2_AC6_NoNewLogCallsInProduction() (internal/uiadapter/logging_plumbing_test.go:L389)
- runGoList() (internal/uiadapter/mock_test.go:L121)
- TestU2_AC8_MockAdapter_NotInProductionBuild() (internal/uiadapter/mock_test.go:L46)
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
- testhelper_eval.go (internal/uiadapter/testhelper_eval.go:L1)
- WithOllamaHost() (internal/uiadapter/testhelper_eval.go:L13)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [go_pkg_strings](/modules/go-pkg-strings.md)
- [NewExecutor](/modules/newexecutor.md)
- [Storage](/modules/storage.md)
- [sync.Mutex](/modules/sync-mutex.md)
- [testLogBuffer](/modules/testlogbuffer.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [bmad/types.go](/modules/bmad-types-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [NewExecutor](/modules/newexecutor.md)
- [newHarness](/modules/newharness.md)

# Features
- no feature plan names these files
