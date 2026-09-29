---
type: Module
title: log/slog.Logger
description: "Graphify community 20: app_uiadapter_claudecli.go, internal/bmad/registry_fs.go, internal/uiadapter/adapter.go, internal/uiadapter/encode.go, internal/uiadapter/encode_test.go, internal/uiadapter/fall"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app_uiadapter_claudecli, resource: app_uiadapter_claudecli.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 9c423a0a128af5b2 }
  - { id: registry_fs, resource: internal/bmad/registry_fs.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 144ec4b753400fe8 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: encode, resource: internal/uiadapter/encode.go, last_modified: "2026-04-26T11:30:52+10:00", digest: bf82e888f12311d6 }
  - { id: encode_test, resource: internal/uiadapter/encode_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: a0cd96e6b9f32bfb }
  - { id: fallback, resource: internal/uiadapter/fallback.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 187709266229767c }
  - { id: fallback_test, resource: internal/uiadapter/fallback_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 2fa70a931cc96a6b }
  - { id: fallback_tiers, resource: internal/uiadapter/fallback_tiers.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 88e0c867731816cd }
  - { id: fallback_tiers_test, resource: internal/uiadapter/fallback_tiers_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 73bb2c32262274bd }
  - { id: logging, resource: internal/uiadapter/logging.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f7d1f46bdb4ea215 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: logging_story5_sanitize_test, resource: internal/uiadapter/logging_story5_sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 49f3cb55ded4bd06 }
  - { id: prefix_cache, resource: internal/uiadapter/prefix_cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 8963dac6e4ff7cab }
  - { id: prefix_cache_test, resource: internal/uiadapter/prefix_cache_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 448153a81ce2894c }
  - { id: sampling, resource: internal/uiadapter/sampling.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 372d2a2aa9064491 }
  - { id: sampling_test, resource: internal/uiadapter/sampling_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 10d6cf4c605c03f7 }
  - { id: sanitize, resource: internal/uiadapter/sanitize.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 6cdb312fe6e18be6 }
  - { id: sanitize_adapter_test, resource: internal/uiadapter/sanitize_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: 2951e338741da597 }
  - { id: sanitize_test, resource: internal/uiadapter/sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: c90f1e17fbe3e70f }
  - { id: semaphore, resource: internal/uiadapter/semaphore.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a0783d029000dab4 }
  - { id: spotlight, resource: internal/uiadapter/spotlight.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 82f02eccfdc52095 }
  - { id: spotlight_test, resource: internal/uiadapter/spotlight_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: f9f81a13aa4a0344 }
  - { id: stages, resource: internal/uiadapter/stages.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 08f05eceb8fc5826 }
  - { id: stages_test, resource: internal/uiadapter/stages_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 74ccbf42bffba71e }
---

# Files
- `app_uiadapter_claudecli.go`
- `internal/bmad/registry_fs.go`
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/encode.go`
- `internal/uiadapter/encode_test.go`
- `internal/uiadapter/fallback.go`
- `internal/uiadapter/fallback_test.go`
- `internal/uiadapter/fallback_tiers.go`
- `internal/uiadapter/fallback_tiers_test.go`
- `internal/uiadapter/logging.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story5_sanitize_test.go`
- `internal/uiadapter/prefix_cache.go`
- `internal/uiadapter/prefix_cache_test.go`
- `internal/uiadapter/sampling.go`
- `internal/uiadapter/sampling_test.go`
- `internal/uiadapter/sanitize.go`
- `internal/uiadapter/sanitize_adapter_test.go`
- `internal/uiadapter/sanitize_test.go`
- `internal/uiadapter/semaphore.go`
- `internal/uiadapter/spotlight.go`
- `internal/uiadapter/spotlight_test.go`
- `internal/uiadapter/stages.go`
- `internal/uiadapter/stages_test.go`

# Symbols
- app_uiadapter_claudecli.go (app_uiadapter_claudecli.go:L1)
- truncReason() (app_uiadapter_claudecli.go:L105)
- splitHeadTail() (app_uiadapter_claudecli.go:L118)
- dumpRawIfRequested() (app_uiadapter_claudecli.go:L129)
- capModel() (app_uiadapter_claudecli.go:L150)
- claudeCLIAdapter (app_uiadapter_claudecli.go:L21)
- newClaudeCLIAdapter() (app_uiadapter_claudecli.go:L31)
- .Translate() (app_uiadapter_claudecli.go:L39)
- registry_fs.go (internal/bmad/registry_fs.go:L1)
- disabledAdapter (internal/uiadapter/adapter.go:L118)
- .Translate() (internal/uiadapter/adapter.go:L156)
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
- fallback_test.go (internal/uiadapter/fallback_test.go:L1)
- TestFallbackAST_LiteralShape() (internal/uiadapter/fallback_test.go:L14)
- TestFallbackAST_TurnSummaryTruncates() (internal/uiadapter/fallback_test.go:L38)
- TestFallbackAST_EmptyRawIsSafe() (internal/uiadapter/fallback_test.go:L51)
- fallback_tiers.go (internal/uiadapter/fallback_tiers.go:L1)
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
- TestStory2_AC2_FreeFunctionsAcceptNilLogger() (internal/uiadapter/logging_plumbing_test.go:L185)
- logging_story5_sanitize_test.go (internal/uiadapter/logging_story5_sanitize_test.go:L1)
- TestStory5_AC8_PerFileEmissionCoverage() (internal/uiadapter/logging_story5_sanitize_test.go:L102)
- TestStory5_AC8_SanitizeDisciplineHolds() (internal/uiadapter/logging_story5_sanitize_test.go:L138)
- TestStory5_AC9_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story5_sanitize_test.go:L173)
- runStory5PayloadShapingSurface() (internal/uiadapter/logging_story5_sanitize_test.go:L56)
- prefix_cache.go (internal/uiadapter/prefix_cache.go:L1)
- ClaudeSystemBlock() (internal/uiadapter/prefix_cache.go:L20)
- cacheControlType() (internal/uiadapter/prefix_cache.go:L53)
- OllamaKeepAliveEncoded() (internal/uiadapter/prefix_cache.go:L65)
- ClaudeSystemBlockJSON() (internal/uiadapter/prefix_cache.go:L86)
- prefix_cache_test.go (internal/uiadapter/prefix_cache_test.go:L1)
- TestClaudeSystemBlock_HasCacheControl() (internal/uiadapter/prefix_cache_test.go:L14)
- TestClaudeSystemBlock_OffTTLSkipsMarker() (internal/uiadapter/prefix_cache_test.go:L26)
- TestPrompt_StaticPrefix_ByteStable() (internal/uiadapter/prefix_cache_test.go:L39)
- TestOllamaKeepAliveEncoded() (internal/uiadapter/prefix_cache_test.go:L59)
- TestClaudeSystemBlock_EmptyPrefixNil() (internal/uiadapter/prefix_cache_test.go:L73)
- sampling.go (internal/uiadapter/sampling.go:L1)
- OllamaSamplingOptions() (internal/uiadapter/sampling.go:L25)
- ClaudeSamplingOptions() (internal/uiadapter/sampling.go:L52)
- TestOllamaSamplingOptions_DeterministicDefaults() (internal/uiadapter/sampling_test.go:L13)
- TestClaudeSamplingOptions_NoSeed() (internal/uiadapter/sampling_test.go:L24)
- TestSamplingOptions_ConfigurableThroughConfig() (internal/uiadapter/sampling_test.go:L33)
- sanitize.go (internal/uiadapter/sanitize.go:L1)
- sanitizeChrome() (internal/uiadapter/sanitize.go:L130)
- SanitizeCapture() (internal/uiadapter/sanitize.go:L29)
- stripOutsideFences() (internal/uiadapter/sanitize.go:L62)
- TestAdapter_Translate_SanitizeRunsBeforeChat() (internal/uiadapter/sanitize_adapter_test.go:L88)
- sanitize_test.go (internal/uiadapter/sanitize_test.go:L1)
- TestSanitize_StripsANSI_Golden() (internal/uiadapter/sanitize_test.go:L13)
- TestSanitize_PreservesFencedCodeBlocks() (internal/uiadapter/sanitize_test.go:L52)
- TestSanitize_Idempotent() (internal/uiadapter/sanitize_test.go:L70)
- TestSanitize_EmptyInput() (internal/uiadapter/sanitize_test.go:L80)
- TestSanitize_WhitespaceTrimmed() (internal/uiadapter/sanitize_test.go:L89)
- TestSanitize_LargeInput() (internal/uiadapter/sanitize_test.go:L98)
- semaphore.go (internal/uiadapter/semaphore.go:L1)
- spotlight.go (internal/uiadapter/spotlight.go:L1)
- Unspotlight() (internal/uiadapter/spotlight.go:L100)
- Spotlight() (internal/uiadapter/spotlight.go:L42)
- spotlight_test.go (internal/uiadapter/spotlight_test.go:L1)
- TestSpotlight_ReplacesWhitespace() (internal/uiadapter/spotlight_test.go:L13)
- TestStory5_AC2_UnspotlightRemoved() (internal/uiadapter/spotlight_test.go:L141)
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
- TestRunTwoStage_OllamaPolicyTwoCalls() (internal/uiadapter/stages_test.go:L86)

# Depends on
- [Config](/modules/config.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [go_pkg_testing](/modules/go-pkg-testing.md)
- [testing.T](/modules/testing-t.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- [allowlist_test.go](/modules/allowlist-test-go.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [repair.go](/modules/repair-go.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)
- [Validate](/modules/validate.md)

# Features
- no feature plan names these files
