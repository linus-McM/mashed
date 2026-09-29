---
type: Module
title: log/slog.Logger
description: "Graphify community 20: app_uiadapter_claudecli.go, internal/bmad/registry_fs.go, internal/uiadapter/adapter.go, internal/uiadapter/backend/claudecli/client.go, internal/uiadapter/backend/claudecli/cli"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: app_uiadapter_claudecli, resource: app_uiadapter_claudecli.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 9c423a0a128af5b2 }
  - { id: registry_fs, resource: internal/bmad/registry_fs.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 144ec4b753400fe8 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: client_test, resource: internal/uiadapter/backend/claudecli/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 4342fc0ff1e40237 }
  - { id: config, resource: internal/uiadapter/config.go, last_modified: "2026-04-26T09:22:14+10:00", digest: d9832db7180bc48b }
  - { id: config_test, resource: internal/uiadapter/config_test.go, last_modified: "2026-04-23T11:43:31+10:00", digest: 748e88156e161a65 }
  - { id: contextguard, resource: internal/uiadapter/contextguard.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 417a5e0974dc194b }
  - { id: encode, resource: internal/uiadapter/encode.go, last_modified: "2026-04-26T11:30:52+10:00", digest: bf82e888f12311d6 }
  - { id: encode_test, resource: internal/uiadapter/encode_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: a0cd96e6b9f32bfb }
  - { id: fallback, resource: internal/uiadapter/fallback.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 187709266229767c }
  - { id: fallback_test, resource: internal/uiadapter/fallback_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 2fa70a931cc96a6b }
  - { id: logging, resource: internal/uiadapter/logging.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f7d1f46bdb4ea215 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: logging_story5_sanitize_test, resource: internal/uiadapter/logging_story5_sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 49f3cb55ded4bd06 }
  - { id: logging_test, resource: internal/uiadapter/logging_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: 744e5721d44f353b }
  - { id: prefix_cache, resource: internal/uiadapter/prefix_cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 8963dac6e4ff7cab }
  - { id: sampling, resource: internal/uiadapter/sampling.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 372d2a2aa9064491 }
  - { id: sanitize, resource: internal/uiadapter/sanitize.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 6cdb312fe6e18be6 }
  - { id: sanitize_test, resource: internal/uiadapter/sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: c90f1e17fbe3e70f }
  - { id: semaphore, resource: internal/uiadapter/semaphore.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a0783d029000dab4 }
  - { id: spotlight, resource: internal/uiadapter/spotlight.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 82f02eccfdc52095 }
  - { id: spotlight_test, resource: internal/uiadapter/spotlight_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: f9f81a13aa4a0344 }
  - { id: stages, resource: internal/uiadapter/stages.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 08f05eceb8fc5826 }
  - { id: stages_test, resource: internal/uiadapter/stages_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 74ccbf42bffba71e }
  - { id: translate_e2e_test, resource: internal/uiadapter/translate_e2e_test.go, last_modified: "2026-04-26T12:39:41+10:00", digest: 45f6254d2ea3d0bc }
---

# Files
- `app_uiadapter_claudecli.go`
- `internal/bmad/registry_fs.go`
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/claudecli/client_test.go`
- `internal/uiadapter/config.go`
- `internal/uiadapter/config_test.go`
- `internal/uiadapter/contextguard.go`
- `internal/uiadapter/encode.go`
- `internal/uiadapter/encode_test.go`
- `internal/uiadapter/fallback.go`
- `internal/uiadapter/fallback_test.go`
- `internal/uiadapter/logging.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story5_sanitize_test.go`
- `internal/uiadapter/logging_test.go`
- `internal/uiadapter/prefix_cache.go`
- `internal/uiadapter/sampling.go`
- `internal/uiadapter/sanitize.go`
- `internal/uiadapter/sanitize_test.go`
- `internal/uiadapter/semaphore.go`
- `internal/uiadapter/spotlight.go`
- `internal/uiadapter/spotlight_test.go`
- `internal/uiadapter/stages.go`
- `internal/uiadapter/stages_test.go`
- `internal/uiadapter/translate_e2e_test.go`

# Symbols
- newClaudeCLIAdapter() (app_uiadapter_claudecli.go:L31)
- registry_fs.go (internal/bmad/registry_fs.go:L1)
- Config (internal/uiadapter/adapter.go:L32)
- NewClient() (internal/uiadapter/backend/claudecli/client.go:L32)
- TestClaudeCLI_VersionCheck() (internal/uiadapter/backend/claudecli/client_test.go:L15)
- TestClaudeCLI_Capabilities() (internal/uiadapter/backend/claudecli/client_test.go:L56)
- config.go (internal/uiadapter/config.go:L1)
- mergeWithDefaults() (internal/uiadapter/config.go:L61)
- configFieldSet() (internal/uiadapter/config.go:L98)
- TestDefaultConfig_MergeOntoZeroValued() (internal/uiadapter/config_test.go:L75)
- TestConfig_HasEveryPlanField() (internal/uiadapter/config_test.go:L87)
- truncateToBudget() (internal/uiadapter/contextguard.go:L105)
- .ApplyClaude() (internal/uiadapter/contextguard.go:L135)
- .OllamaOptions() (internal/uiadapter/contextguard.go:L176)
- ContextGuard (internal/uiadapter/contextguard.go:L20)
- .ApplyOllama() (internal/uiadapter/contextguard.go:L55)
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
- fallback.go (internal/uiadapter/fallback.go:L1)
- FallbackAST() (internal/uiadapter/fallback.go:L23)
- firstLine() (internal/uiadapter/fallback.go:L47)
- TestFallbackAST_LiteralShape() (internal/uiadapter/fallback_test.go:L14)
- TestFallbackAST_TurnSummaryTruncates() (internal/uiadapter/fallback_test.go:L38)
- TestFallbackAST_EmptyRawIsSafe() (internal/uiadapter/fallback_test.go:L51)
- nilSafeLogger() (internal/uiadapter/logging.go:L169)
- TestStory2_AC2_FreeFunctionsAcceptNilLogger() (internal/uiadapter/logging_plumbing_test.go:L185)
- runStory5PayloadShapingSurface() (internal/uiadapter/logging_story5_sanitize_test.go:L56)
- TestStory1_AC5_ConfigDefaults() (internal/uiadapter/logging_test.go:L260)
- prefix_cache.go (internal/uiadapter/prefix_cache.go:L1)
- ClaudeSystemBlock() (internal/uiadapter/prefix_cache.go:L20)
- cacheControlType() (internal/uiadapter/prefix_cache.go:L53)
- OllamaKeepAliveEncoded() (internal/uiadapter/prefix_cache.go:L65)
- ClaudeSystemBlockJSON() (internal/uiadapter/prefix_cache.go:L86)
- sampling.go (internal/uiadapter/sampling.go:L1)
- OllamaSamplingOptions() (internal/uiadapter/sampling.go:L25)
- ClaudeSamplingOptions() (internal/uiadapter/sampling.go:L52)
- sanitize.go (internal/uiadapter/sanitize.go:L1)
- sanitizeChrome() (internal/uiadapter/sanitize.go:L130)
- SanitizeCapture() (internal/uiadapter/sanitize.go:L29)
- stripOutsideFences() (internal/uiadapter/sanitize.go:L62)
- sanitize_test.go (internal/uiadapter/sanitize_test.go:L1)
- TestSanitize_StripsANSI_Golden() (internal/uiadapter/sanitize_test.go:L13)
- TestSanitize_PreservesFencedCodeBlocks() (internal/uiadapter/sanitize_test.go:L52)
- TestSanitize_Idempotent() (internal/uiadapter/sanitize_test.go:L70)
- TestSanitize_EmptyInput() (internal/uiadapter/sanitize_test.go:L80)
- TestSanitize_WhitespaceTrimmed() (internal/uiadapter/sanitize_test.go:L89)
- TestSanitize_LargeInput() (internal/uiadapter/sanitize_test.go:L98)
- newSemaphore() (internal/uiadapter/semaphore.go:L25)
- spotlight.go (internal/uiadapter/spotlight.go:L1)
- Unspotlight() (internal/uiadapter/spotlight.go:L100)
- Spotlight() (internal/uiadapter/spotlight.go:L42)
- spotlight_test.go (internal/uiadapter/spotlight_test.go:L1)
- TestSpotlight_ReplacesWhitespace() (internal/uiadapter/spotlight_test.go:L13)
- TestSpotlight_DisabledBypass() (internal/uiadapter/spotlight_test.go:L21)
- TestSpotlight_RoundTripLossless() (internal/uiadapter/spotlight_test.go:L29)
- TestSpotlight_InjectionCorpus() (internal/uiadapter/spotlight_test.go:L39)
- TestSpotlight_EmptyAndUnicode() (internal/uiadapter/spotlight_test.go:L65)
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
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [cache_test.go](/modules/cache-test-go.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [repair.go](/modules/repair-go.md)
- [Validate](/modules/validate.md)

# Features
- no feature plan names these files
