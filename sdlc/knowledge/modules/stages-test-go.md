---
type: Module
title: stages_test.go
description: "Graphify community 20: internal/uiadapter/backend/claudecli/client.go, internal/uiadapter/encode.go, internal/uiadapter/encode_test.go, internal/uiadapter/fallback_tiers.go, internal/uiadapter/fallbac"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-09-29T07:07:25Z", digest: 414128d39b8e1d31 }
  - { id: encode, resource: internal/uiadapter/encode.go, last_modified: "2026-09-29T07:07:25Z", digest: bf82e888f12311d6 }
  - { id: encode_test, resource: internal/uiadapter/encode_test.go, last_modified: "2026-09-29T07:07:25Z", digest: a0cd96e6b9f32bfb }
  - { id: fallback_tiers, resource: internal/uiadapter/fallback_tiers.go, last_modified: "2026-09-29T07:07:25Z", digest: 88e0c867731816cd }
  - { id: fallback_tiers_test, resource: internal/uiadapter/fallback_tiers_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 73bb2c32262274bd }
  - { id: logging_story4_sanitize_test, resource: internal/uiadapter/logging_story4_sanitize_test.go, last_modified: "2026-09-29T07:07:25Z", digest: b4229ff1bc6c01ad }
  - { id: stages, resource: internal/uiadapter/stages.go, last_modified: "2026-09-29T07:07:25Z", digest: 08f05eceb8fc5826 }
  - { id: stages_test, resource: internal/uiadapter/stages_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 74ccbf42bffba71e }
  - { id: translate_e2e_test, resource: internal/uiadapter/translate_e2e_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 45f6254d2ea3d0bc }
---

# Files
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/encode.go`
- `internal/uiadapter/encode_test.go`
- `internal/uiadapter/fallback_tiers.go`
- `internal/uiadapter/fallback_tiers_test.go`
- `internal/uiadapter/logging_story4_sanitize_test.go`
- `internal/uiadapter/stages.go`
- `internal/uiadapter/stages_test.go`
- `internal/uiadapter/translate_e2e_test.go`

# Symbols
- .GenerateSingleShot() (internal/uiadapter/backend/claudecli/client.go:L116)
- Client (internal/uiadapter/backend/claudecli/client.go:L21)
- .runOneShot() (internal/uiadapter/backend/claudecli/client.go:L234)
- .Name() (internal/uiadapter/backend/claudecli/client.go:L41)
- .Classify() (internal/uiadapter/backend/claudecli/client.go:L72)
- .Generate() (internal/uiadapter/backend/claudecli/client.go:L93)
- encode.go (internal/uiadapter/encode.go:L1)
- schemaBytes() (internal/uiadapter/encode.go:L112)
- OllamaFormatPayload() (internal/uiadapter/encode.go:L35)
- emitOllamaFormat() (internal/uiadapter/encode.go:L51)
- ClaudeToolInputSchema() (internal/uiadapter/encode.go:L71)
- encode_test.go (internal/uiadapter/encode_test.go:L1)
- TestOllamaFormatPayload_SchemaObject() (internal/uiadapter/encode_test.go:L15)
- TestOllamaFormatPayload_LooseRollback() (internal/uiadapter/encode_test.go:L28)
- TestClaudeToolInputSchema_SchemaObject() (internal/uiadapter/encode_test.go:L37)
- TestSchemas_IdenticalAcrossBackends() (internal/uiadapter/encode_test.go:L49)
- tierFailureReason() (internal/uiadapter/fallback_tiers.go:L19)
- FallbackTier (internal/uiadapter/fallback_tiers.go:L43)
- RunWithFallback() (internal/uiadapter/fallback_tiers.go:L57)
- fallback_tiers_test.go (internal/uiadapter/fallback_tiers_test.go:L1)
- TestFallback_TieredRecovery() (internal/uiadapter/fallback_tiers_test.go:L15)
- TestFallback_MinimalKindWhenAllBackendsFail() (internal/uiadapter/fallback_tiers_test.go:L37)
- TestFallback_PlaintextLastResort() (internal/uiadapter/fallback_tiers_test.go:L54)
- TestFallback_PrimarySuccessEscalatedFromEmpty() (internal/uiadapter/fallback_tiers_test.go:L69)
- TestFallback_ContextCancellation() (internal/uiadapter/fallback_tiers_test.go:L82)
- TestStory4_AC8_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story4_sanitize_test.go:L182)
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
- drivePerPhaseHelpers() (internal/uiadapter/translate_e2e_test.go:L63)

# Depends on
- [Client](/modules/client.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [go_pkg_fmt](/modules/go-pkg-fmt.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [StubBackend](/modules/stubbackend.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [NewFastPathClassifier](/modules/newfastpathclassifier.md)
- [NewRepairer](/modules/newrepairer.md)
- [nilSafeLogger](/modules/nilsafelogger.md)

# Features
- no feature plan names these files
