---
type: Module
title: log/slog.Logger
description: "Graphify community 51: internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/backend/claudeapi/client_test.go, internal/uiadapter/encode.go, internal/uiadapter/encode_test.go, internal/ui"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client_test, resource: internal/uiadapter/backend/claudeapi/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e3e3d9c28e77effe }
  - { id: encode, resource: internal/uiadapter/encode.go, last_modified: "2026-04-26T11:30:52+10:00", digest: bf82e888f12311d6 }
  - { id: encode_test, resource: internal/uiadapter/encode_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: a0cd96e6b9f32bfb }
  - { id: logging, resource: internal/uiadapter/logging.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f7d1f46bdb4ea215 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: logging_story5_sanitize_test, resource: internal/uiadapter/logging_story5_sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 49f3cb55ded4bd06 }
  - { id: repair, resource: internal/uiadapter/repair.go, last_modified: "2026-04-26T10:43:55+10:00", digest: cc21726252779f85 }
  - { id: repair_test, resource: internal/uiadapter/repair_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 7d95d0c21f4ee4e3 }
  - { id: sampling, resource: internal/uiadapter/sampling.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 372d2a2aa9064491 }
  - { id: sampling_test, resource: internal/uiadapter/sampling_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 10d6cf4c605c03f7 }
  - { id: sanitize, resource: internal/uiadapter/sanitize.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 6cdb312fe6e18be6 }
  - { id: sanitize_adapter_test, resource: internal/uiadapter/sanitize_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 2951e338741da597 }
  - { id: sanitize_test, resource: internal/uiadapter/sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: c90f1e17fbe3e70f }
  - { id: stages, resource: internal/uiadapter/stages.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 08f05eceb8fc5826 }
  - { id: stages_test, resource: internal/uiadapter/stages_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 74ccbf42bffba71e }
  - { id: validator, resource: internal/uiadapter/validator.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ea853a4a6aadb5f2 }
  - { id: validator_test, resource: internal/uiadapter/validator_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ad90be2e31d6a4a0 }
---

# Files
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudeapi/client_test.go`
- `internal/uiadapter/encode.go`
- `internal/uiadapter/encode_test.go`
- `internal/uiadapter/logging.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story5_sanitize_test.go`
- `internal/uiadapter/repair.go`
- `internal/uiadapter/repair_test.go`
- `internal/uiadapter/sampling.go`
- `internal/uiadapter/sampling_test.go`
- `internal/uiadapter/sanitize.go`
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/sanitize_test.go`
- `internal/uiadapter/stages.go`
- `internal/uiadapter/stages_test.go`
- `internal/uiadapter/validator.go`
- `internal/uiadapter/validator_test.go`

# Symbols
- .Generate() (internal/uiadapter/backend/claudeapi/client.go:L114)
- requestBody (internal/uiadapter/backend/claudeapi/client.go:L147)
- classifyRequest() (internal/uiadapter/backend/claudeapi/client.go:L158)
- generateRequest() (internal/uiadapter/backend/claudeapi/client.go:L174)
- .Classify() (internal/uiadapter/backend/claudeapi/client.go:L93)
- TestClaudeAPI_PromptCacheMarkers() (internal/uiadapter/backend/claudeapi/client_test.go:L111)
- encode.go (internal/uiadapter/encode.go:L1)
- schemaBytes() (internal/uiadapter/encode.go:L112)
- OllamaFormatPayload() (internal/uiadapter/encode.go:L35)
- emitOllamaFormat() (internal/uiadapter/encode.go:L51)
- ClaudeToolInputSchema() (internal/uiadapter/encode.go:L71)
- ClaudeToolName() (internal/uiadapter/encode.go:L93)
- encode_test.go (internal/uiadapter/encode_test.go:L1)
- TestOllamaFormatPayload_SchemaObject() (internal/uiadapter/encode_test.go:L15)
- TestOllamaFormatPayload_LooseRollback() (internal/uiadapter/encode_test.go:L28)
- TestClaudeToolInputSchema_SchemaObject() (internal/uiadapter/encode_test.go:L37)
- TestSchemas_IdenticalAcrossBackends() (internal/uiadapter/encode_test.go:L49)
- TestClaudeToolName_StableShape() (internal/uiadapter/encode_test.go:L63)
- nilSafeLogger() (internal/uiadapter/logging.go:L169)
- TestStory2_AC2_FreeFunctionsAcceptNilLogger() (internal/uiadapter/logging_plumbing_test.go:L185)
- logging_story5_sanitize_test.go (internal/uiadapter/logging_story5_sanitize_test.go:L1)
- TestStory5_AC8_PerFileEmissionCoverage() (internal/uiadapter/logging_story5_sanitize_test.go:L102)
- TestStory5_AC8_SanitizeDisciplineHolds() (internal/uiadapter/logging_story5_sanitize_test.go:L138)
- TestStory5_AC9_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story5_sanitize_test.go:L173)
- runStory5PayloadShapingSurface() (internal/uiadapter/logging_story5_sanitize_test.go:L56)
- RepairAttempt (internal/uiadapter/repair.go:L39)
- BuildRepairPrompt() (internal/uiadapter/repair.go:L61)
- buildRepairPromptString() (internal/uiadapter/repair.go:L77)
- TestBuildRepairPrompt_Shape() (internal/uiadapter/repair_test.go:L111)
- sampling.go (internal/uiadapter/sampling.go:L1)
- OllamaSamplingOptions() (internal/uiadapter/sampling.go:L25)
- ClaudeSamplingOptions() (internal/uiadapter/sampling.go:L52)
- sampling_test.go (internal/uiadapter/sampling_test.go:L1)
- TestOllamaSamplingOptions_DeterministicDefaults() (internal/uiadapter/sampling_test.go:L13)
- TestClaudeSamplingOptions_NoSeed() (internal/uiadapter/sampling_test.go:L24)
- TestSamplingOptions_ConfigurableThroughConfig() (internal/uiadapter/sampling_test.go:L33)
- sanitizeChrome() (internal/uiadapter/sanitize.go:L130)
- SanitizeCapture() (internal/uiadapter/sanitize.go:L29)
- stripOutsideFences() (internal/uiadapter/sanitize.go:L62)
- TestAdapter_Translate_ANSIWrappedURL_NotUntrusted() (internal/uiadapter/sanitize_adapter_test.go:L21)
- sanitize_test.go (internal/uiadapter/sanitize_test.go:L1)
- TestSanitize_StripsANSI_Golden() (internal/uiadapter/sanitize_test.go:L13)
- TestSanitize_PreservesFencedCodeBlocks() (internal/uiadapter/sanitize_test.go:L52)
- TestSanitize_Idempotent() (internal/uiadapter/sanitize_test.go:L70)
- TestSanitize_EmptyInput() (internal/uiadapter/sanitize_test.go:L80)
- TestSanitize_WhitespaceTrimmed() (internal/uiadapter/sanitize_test.go:L89)
- TestSanitize_LargeInput() (internal/uiadapter/sanitize_test.go:L98)
- stages.go (internal/uiadapter/stages.go:L1)
- AssembleStage1() (internal/uiadapter/stages.go:L100)
- assembleStage1String() (internal/uiadapter/stages.go:L116)
- AssembleStage2() (internal/uiadapter/stages.go:L133)
- assembleStage2String() (internal/uiadapter/stages.go:L152)
- RunTwoStage() (internal/uiadapter/stages.go:L180)
- StageKind (internal/uiadapter/stages.go:L35)
- ParseStageKind() (internal/uiadapter/stages.go:L53)
- ClassifyPrompt() (internal/uiadapter/stages.go:L73)
- GeneratePromptFor() (internal/uiadapter/stages.go:L76)
- stages_test.go (internal/uiadapter/stages_test.go:L1)
- TestRunTwoStage_ClassifyErrorAborts() (internal/uiadapter/stages_test.go:L109)
- TestRunTwoStage_ContextCancelled() (internal/uiadapter/stages_test.go:L124)
- TestStageKind_ParseRoundtrip() (internal/uiadapter/stages_test.go:L17)
- TestAssembleStage1_ByteStable() (internal/uiadapter/stages_test.go:L36)
- TestAssembleStage2_ContainsKindDirective() (internal/uiadapter/stages_test.go:L50)
- TestAssembleStage2_UnknownKindErrors() (internal/uiadapter/stages_test.go:L61)
- TestPromptsBudget() (internal/uiadapter/stages_test.go:L69)
- TestRunTwoStage_OllamaPolicyTwoCalls() (internal/uiadapter/stages_test.go:L86)
- validator.go (internal/uiadapter/validator.go:L1)
- emitValidatorDone() (internal/uiadapter/validator.go:L107)
- emitRuleFail() (internal/uiadapter/validator.go:L123)
- dedupeAndSort() (internal/uiadapter/validator.go:L138)
- terminalReason() (internal/uiadapter/validator.go:L158)
- applyRule1() (internal/uiadapter/validator.go:L170)
- applyRules2And3() (internal/uiadapter/validator.go:L197)
- applyRule4() (internal/uiadapter/validator.go:L221)
- applyRule5() (internal/uiadapter/validator.go:L244)
- applyRule6() (internal/uiadapter/validator.go:L266)
- applyRule7() (internal/uiadapter/validator.go:L293)
- applyRule8() (internal/uiadapter/validator.go:L312)
- knownNodeType() (internal/uiadapter/validator.go:L327)
- contentPreserved() (internal/uiadapter/validator.go:L340)
- collectRenderedText() (internal/uiadapter/validator.go:L364)
- Validate() (internal/uiadapter/validator.go:L59)
- validator_test.go (internal/uiadapter/validator_test.go:L1)
- TestValidate_Rule5_LongKeyTruncated() (internal/uiadapter/validator_test.go:L111)
- TestValidate_Rule6_CountCapOptionalDrop() (internal/uiadapter/validator_test.go:L130)
- TestValidate_Rule1_UnknownTypeBecomesMarkdown() (internal/uiadapter/validator_test.go:L16)
- TestValidate_Rule6_CountCapRequiredDropFallback() (internal/uiadapter/validator_test.go:L160)
- TestValidate_Rule7_NodeCapRequiredDropFallback() (internal/uiadapter/validator_test.go:L179)
- TestValidate_Rule8_OversizeFullFallback() (internal/uiadapter/validator_test.go:L199)
- TestValidate_RuleOrdering_Deterministic() (internal/uiadapter/validator_test.go:L214)
- TestValidate_Security71_FileWidgetPreserved() (internal/uiadapter/validator_test.go:L247)
- TestContentPreservation_URLDropped_SetsUntrusted() (internal/uiadapter/validator_test.go:L278)
- TestContentPreservation_CodeBlockDropped_SetsUntrusted() (internal/uiadapter/validator_test.go:L293)
- TestContentPreservation_NumberedListToChoice_NoUntrusted() (internal/uiadapter/validator_test.go:L308)
- TestValidate_Rule2_EmptyOptionsDropsGroup() (internal/uiadapter/validator_test.go:L33)
- TestValidate_RuleOrdering_PerNodeDeclaration() (internal/uiadapter/validator_test.go:L334)
- TestValidate_WidgetUnknownField_Rejected() (internal/uiadapter/validator_test.go:L357)
- TestValidate_Rule3_NoWidgetDropsGroup() (internal/uiadapter/validator_test.go:L63)
- TestValidate_Rule4_DuplicateKeySuffix() (internal/uiadapter/validator_test.go:L76)

# Depends on
- [claudeapi/client.go](/modules/claudeapi-client-go.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [go_pkg_bytes](/modules/go-pkg-bytes.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Features
- no feature plan names these files
