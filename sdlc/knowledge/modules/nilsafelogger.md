---
type: Module
title: nilSafeLogger
description: "Graphify community 20: internal/bmad/registry_fs.go, internal/uiadapter/backend/claudeapi/client.go, internal/uiadapter/encode.go, internal/uiadapter/encode_test.go, internal/uiadapter/fallback_tiers."
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: registry_fs, resource: internal/bmad/registry_fs.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 144ec4b753400fe8 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: encode, resource: internal/uiadapter/encode.go, last_modified: "2026-04-26T11:30:52+10:00", digest: bf82e888f12311d6 }
  - { id: encode_test, resource: internal/uiadapter/encode_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: a0cd96e6b9f32bfb }
  - { id: fallback_tiers, resource: internal/uiadapter/fallback_tiers.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 88e0c867731816cd }
  - { id: fallback_tiers_test, resource: internal/uiadapter/fallback_tiers_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 73bb2c32262274bd }
  - { id: logging, resource: internal/uiadapter/logging.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f7d1f46bdb4ea215 }
  - { id: logging_comprehensive_test, resource: internal/uiadapter/logging_comprehensive_test.go, last_modified: "2026-04-26T12:39:41+10:00", digest: 446816f0c9fa2be3 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: logging_story4_sanitize_test, resource: internal/uiadapter/logging_story4_sanitize_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: b4229ff1bc6c01ad }
  - { id: logging_story5_sanitize_test, resource: internal/uiadapter/logging_story5_sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 49f3cb55ded4bd06 }
  - { id: semaphore, resource: internal/uiadapter/semaphore.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a0783d029000dab4 }
  - { id: stages, resource: internal/uiadapter/stages.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 08f05eceb8fc5826 }
  - { id: stages_test, resource: internal/uiadapter/stages_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 74ccbf42bffba71e }
---

# Files
- `internal/bmad/registry_fs.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/encode.go`
- `internal/uiadapter/encode_test.go`
- `internal/uiadapter/fallback_tiers.go`
- `internal/uiadapter/fallback_tiers_test.go`
- `internal/uiadapter/logging.go`
- `internal/uiadapter/logging_comprehensive_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story4_sanitize_test.go`
- `internal/uiadapter/logging_story5_sanitize_test.go`
- `internal/uiadapter/semaphore.go`
- `internal/uiadapter/stages.go`
- `internal/uiadapter/stages_test.go`

# Symbols
- registry_fs.go (internal/bmad/registry_fs.go:L1)
- .Generate() (internal/uiadapter/backend/claudeapi/client.go:L114)
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
- tierFailureReason() (internal/uiadapter/fallback_tiers.go:L19)
- FallbackTier (internal/uiadapter/fallback_tiers.go:L43)
- RunWithFallback() (internal/uiadapter/fallback_tiers.go:L57)
- fallback_tiers_test.go (internal/uiadapter/fallback_tiers_test.go:L1)
- TestFallback_TieredRecovery() (internal/uiadapter/fallback_tiers_test.go:L15)
- TestFallback_MinimalKindWhenAllBackendsFail() (internal/uiadapter/fallback_tiers_test.go:L37)
- TestFallback_PlaintextLastResort() (internal/uiadapter/fallback_tiers_test.go:L54)
- TestFallback_PrimarySuccessEscalatedFromEmpty() (internal/uiadapter/fallback_tiers_test.go:L69)
- TestFallback_ContextCancellation() (internal/uiadapter/fallback_tiers_test.go:L82)
- nilSafeLogger() (internal/uiadapter/logging.go:L169)
- TestStory6_AC1_NilSafeLoggerComprehensive() (internal/uiadapter/logging_comprehensive_test.go:L305)
- TestStory2_AC5_NilSafeLogger() (internal/uiadapter/logging_plumbing_test.go:L40)
- TestStory4_AC8_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story4_sanitize_test.go:L182)
- runStory5PayloadShapingSurface() (internal/uiadapter/logging_story5_sanitize_test.go:L56)
- newSemaphore() (internal/uiadapter/semaphore.go:L25)
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

# Depends on
- [Config](/modules/config.md)
- [context.Context](/modules/context-context.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [NewFastPathClassifier](/modules/newfastpathclassifier.md)
- [NewRepairer](/modules/newrepairer.md)
- [SanitizeCapture](/modules/sanitizecapture.md)
- [Spotlight](/modules/spotlight.md)
- [TestStory2_AC2_FreeFunctionsAcceptNilLogger](/modules/teststory2-ac2-freefunctionsacceptnillogger.md)

# Features
- no feature plan names these files
