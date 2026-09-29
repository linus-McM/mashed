---
type: Module
title: testing.T
description: "Graphify community 0: app_terminal_registry_test.go, app_uiadapter_v3_test.go, internal/bmad/condition.go, internal/bmad/condition_test.go, internal/bmad/executor_interactive_smoke_test.go, internal/b"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: app_terminal_registry_test, resource: app_terminal_registry_test.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 2f85a0b04c1b8ba7 }
  - { id: app_uiadapter_v3_test, resource: app_uiadapter_v3_test.go, last_modified: "2026-04-23T11:43:31+10:00", digest: f965d914b8a9c4fa }
  - { id: condition, resource: internal/bmad/condition.go, last_modified: "2026-04-08T17:10:27+10:00", digest: 956ffd61999ceb65 }
  - { id: condition_test, resource: internal/bmad/condition_test.go, last_modified: "2026-04-08T17:10:27+10:00", digest: 0058fbc2d3f1868b }
  - { id: executor_interactive_smoke_test, resource: internal/bmad/executor_interactive_smoke_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 75f14883e4b4821b }
  - { id: interactive_types_test, resource: internal/bmad/interactive_types_test.go, last_modified: "2026-04-20T13:01:41+10:00", digest: 0282e550b6b93893 }
  - { id: testutil_interactive_test, resource: internal/bmad/testutil_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: f165a17bd64570bc }
  - { id: types_nodetype_test, resource: internal/bmad/types_nodetype_test.go, last_modified: "2026-04-08T17:10:27+10:00", digest: 63313f7f3a1e92c3 }
  - { id: types_test, resource: internal/bmad/types_test.go, last_modified: "2026-04-21T20:21:32+10:00", digest: f7fe93cfdff41ed5 }
  - { id: codegen_test, resource: internal/uiadapter/codegen_test.go, last_modified: "2026-04-23T11:04:07+10:00", digest: a4cb53187a0fa90c }
  - { id: schema_test, resource: internal/uiadapter/schema_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 706f4f8efc89d9d5 }
  - { id: schemas_test, resource: internal/uiadapter/schemas_test.go, last_modified: "2026-04-23T11:04:07+10:00", digest: e7499deb809d563b }
---

# Files
- `app_terminal_registry_test.go`
- `app_uiadapter_v3_test.go`
- `internal/bmad/condition.go`
- `internal/bmad/condition_test.go`
- `internal/bmad/executor_interactive_smoke_test.go`
- `internal/bmad/interactive_types_test.go`
- `internal/bmad/testutil_interactive_test.go`
- `internal/bmad/types_nodetype_test.go`
- `internal/bmad/types_test.go`
- `internal/uiadapter/codegen_test.go`
- `internal/uiadapter/schema_test.go`
- `internal/uiadapter/schemas_test.go`

# Symbols
- app_terminal_registry_test.go (app_terminal_registry_test.go:L1)
- testApp() (app_terminal_registry_test.go:L101)
- TestStory1_AC1_TerminalSessionJSONTags() (app_terminal_registry_test.go:L111)
- TestStory1_AC1_TerminalSessionRoundTrip() (app_terminal_registry_test.go:L137)
- TestStory1_AC2_FilterByRepo() (app_terminal_registry_test.go:L167)
- TestStory1_AC2_EmptyRepoPath() (app_terminal_registry_test.go:L187)
- TestStory1_AC2_SortBySpawnedAt() (app_terminal_registry_test.go:L199)
- TestStory1_AC3_PruneDeadSessions() (app_terminal_registry_test.go:L221)
- TestStory1_AC3_AllSessionsDead() (app_terminal_registry_test.go:L243)
- TestStory1_AC4_KillRemovesFromRegistry() (app_terminal_registry_test.go:L260)
- TestStory1_AC5_KillAlreadyDead() (app_terminal_registry_test.go:L282)
- TestStory1_AC5_KillNotFound() (app_terminal_registry_test.go:L297)
- TestStory1_RegisterOverwrite() (app_terminal_registry_test.go:L308)
- TestStory2_AC1_RecoverTerminalAndAgent() (app_terminal_registry_test.go:L335)
- TestStory2_AC2_FilterNonMatchingPrefixes() (app_terminal_registry_test.go:L358)
- TestStory2_AC3_TmuxNotRunning() (app_terminal_registry_test.go:L374)
- TestStory2_AC4_IdempotentRecovery() (app_terminal_registry_test.go:L385)
- TestStory2_ParsesRepoNameFromSessionName() (app_terminal_registry_test.go:L400)
- TestStory2_NoTimestampSuffix() (app_terminal_registry_test.go:L429)
- TestStory2_EmptyAndWhitespaceLines() (app_terminal_registry_test.go:L441)
- TestStory3_AC1_SpawnRegistersAgentSession() (app_terminal_registry_test.go:L454)
- TestStory3_AC2_SpawnRegistersTerminalSession() (app_terminal_registry_test.go:L481)
- TestStory3_AC3_KillAgentDeregisters() (app_terminal_registry_test.go:L505)
- newFakeManager() (app_terminal_registry_test.go:L51)
- TestStory3_AC4_FailedSpawnNoRegistration() (app_terminal_registry_test.go:L535)
- TestStory1_RegisterSessionConcurrent() (app_terminal_registry_test.go:L549)
- TestStory4_AC7_RecoverSessionsIsNoOp() (app_terminal_registry_test.go:L575)
- TestStory5_ResolveTmuxTarget() (app_terminal_registry_test.go:L589)
- app_uiadapter_v3_test.go (app_uiadapter_v3_test.go:L1)
- setupTempConfig() (app_uiadapter_v3_test.go:L15)
- TestApp_SetBackend_Validates() (app_uiadapter_v3_test.go:L24)
- TestApp_SetClaudeModel_Validates() (app_uiadapter_v3_test.go:L38)
- TestApp_SetRouterPolicy_Validates() (app_uiadapter_v3_test.go:L51)
- TestApp_ListBackendsAvailable_StableOrder() (app_uiadapter_v3_test.go:L64)
- TestApp_Setters_RoundtripToConfig() (app_uiadapter_v3_test.go:L74)
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
- interactive_types_test.go (internal/bmad/interactive_types_test.go:L1)
- TestIterationGateNilVsPopulated() (internal/bmad/interactive_types_test.go:L259)
- TestNodeAwaitingInputJSONRoundTrip() (internal/bmad/interactive_types_test.go:L29)
- TestWorkflowExecutionInteractiveState() (internal/bmad/interactive_types_test.go:L357)
- TestPendingPromptRoundTrip() (internal/bmad/interactive_types_test.go:L417)
- TestConstantJSONValues() (internal/bmad/interactive_types_test.go:L460)
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
- codegen_test.go (internal/uiadapter/codegen_test.go:L1)
- TestCodegen_AC_A1_GenDirectiveFileIsBuildTagFree() (internal/uiadapter/codegen_test.go:L108)
- TestCodegen_AC_A1_GeneratedFileExists() (internal/uiadapter/codegen_test.go:L124)
- TestCodegen_AC_A1_GeneratedFileDeclaresUIASTPackage() (internal/uiadapter/codegen_test.go:L135)
- TestCodegen_AC_A1_GeneratedPackageDirectoryExists() (internal/uiadapter/codegen_test.go:L149)
- TestCodegen_NoDrift() (internal/uiadapter/codegen_test.go:L179)
- ensureGoJsonschemaOnPATH() (internal/uiadapter/codegen_test.go:L40)
- TestCodegen_AC_A1_GenDirectiveFileExists() (internal/uiadapter/codegen_test.go:L84)
- TestCodegen_AC_A1_GenDirectiveFileDeclaresGoGenerate() (internal/uiadapter/codegen_test.go:L96)
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

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [go_pkg_strings](/modules/go-pkg-strings.md)
- [interactiveHarness](/modules/interactiveharness.md)
- [ManagedSession](/modules/managedsession.md)
- [ProcessDef](/modules/processdef.md)
- [TmuxPane](/modules/tmuxpane.md)

# Inferred
- [bmad/types.go](/modules/bmad-types-go.md)
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [loadConfig](/modules/loadconfig.md)
- [newHarness](/modules/newharness.md)

# Features
- no feature plan names these files
