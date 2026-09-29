---
type: Module
title: DefaultConfig
description: "Graphify community 14: app_uiadapter_claudecli.go, internal/uiadapter/adapter.go, internal/uiadapter/allowlist.go, internal/uiadapter/allowlist_test.go, internal/uiadapter/backend/claudeapi/client.go,"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: app_uiadapter_claudecli, resource: app_uiadapter_claudecli.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 9c423a0a128af5b2 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: allowlist, resource: internal/uiadapter/allowlist.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ec96ab6a3976170d }
  - { id: allowlist_test, resource: internal/uiadapter/allowlist_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 865e6c967ac886df }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client_test, resource: internal/uiadapter/backend/claudeapi/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e3e3d9c28e77effe }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: client_test, resource: internal/uiadapter/backend/claudecli/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 4342fc0ff1e40237 }
  - { id: breaker, resource: internal/uiadapter/breaker.go, last_modified: "2026-04-26T10:14:41+10:00", digest: ce7fb713c254dda5 }
  - { id: breaker_test, resource: internal/uiadapter/breaker_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: c08c33460a000e4b }
  - { id: cache, resource: internal/uiadapter/cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 10b083d2f056ef95 }
  - { id: cache_test, resource: internal/uiadapter/cache_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: e8cf326b1dcc1d7a }
  - { id: config, resource: internal/uiadapter/config.go, last_modified: "2026-04-26T09:22:14+10:00", digest: d9832db7180bc48b }
  - { id: config_test, resource: internal/uiadapter/config_test.go, last_modified: "2026-04-23T11:43:31+10:00", digest: 748e88156e161a65 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: logging_story3_sanitize_test, resource: internal/uiadapter/logging_story3_sanitize_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a4de64483e766cb6 }
  - { id: logging_story4_sanitize_test, resource: internal/uiadapter/logging_story4_sanitize_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: b4229ff1bc6c01ad }
  - { id: logging_test, resource: internal/uiadapter/logging_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: 744e5721d44f353b }
  - { id: prefix_cache, resource: internal/uiadapter/prefix_cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 8963dac6e4ff7cab }
  - { id: prefix_cache_test, resource: internal/uiadapter/prefix_cache_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 448153a81ce2894c }
  - { id: repair, resource: internal/uiadapter/repair.go, last_modified: "2026-04-26T10:43:55+10:00", digest: cc21726252779f85 }
  - { id: repair_test, resource: internal/uiadapter/repair_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 7d95d0c21f4ee4e3 }
  - { id: sampling, resource: internal/uiadapter/sampling.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 372d2a2aa9064491 }
  - { id: sampling_test, resource: internal/uiadapter/sampling_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 10d6cf4c605c03f7 }
  - { id: translate_e2e_test, resource: internal/uiadapter/translate_e2e_test.go, last_modified: "2026-04-26T12:39:41+10:00", digest: 45f6254d2ea3d0bc }
---

# Files
- `app_uiadapter_claudecli.go`
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/allowlist.go`
- `internal/uiadapter/allowlist_test.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudeapi/client_test.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/claudecli/client_test.go`
- `internal/uiadapter/breaker.go`
- `internal/uiadapter/breaker_test.go`
- `internal/uiadapter/cache.go`
- `internal/uiadapter/cache_test.go`
- `internal/uiadapter/config.go`
- `internal/uiadapter/config_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story3_sanitize_test.go`
- `internal/uiadapter/logging_story4_sanitize_test.go`
- `internal/uiadapter/logging_test.go`
- `internal/uiadapter/prefix_cache.go`
- `internal/uiadapter/prefix_cache_test.go`
- `internal/uiadapter/repair.go`
- `internal/uiadapter/repair_test.go`
- `internal/uiadapter/sampling.go`
- `internal/uiadapter/sampling_test.go`
- `internal/uiadapter/translate_e2e_test.go`

# Symbols
- newClaudeCLIAdapter() (app_uiadapter_claudecli.go:L31)
- Config (internal/uiadapter/adapter.go:L32)
- CheckModelAllowlist() (internal/uiadapter/allowlist.go:L40)
- allowlist_test.go (internal/uiadapter/allowlist_test.go:L1)
- TestStory5_AC6_AllowlistNilLoggerStillSafe() (internal/uiadapter/allowlist_test.go:L110)
- TestAllowlist_WarnsOnUnvetted() (internal/uiadapter/allowlist_test.go:L18)
- TestAllowlist_AllowsKnownModels() (internal/uiadapter/allowlist_test.go:L32)
- TestAllowlist_OverrideFlag() (internal/uiadapter/allowlist_test.go:L44)
- TestAllowlist_WarnsOnUnvettedClaude() (internal/uiadapter/allowlist_test.go:L57)
- TestStory5_AC6_AllowlistDefaultRemoved() (internal/uiadapter/allowlist_test.go:L93)
- requestBody (internal/uiadapter/backend/claudeapi/client.go:L147)
- classifyRequest() (internal/uiadapter/backend/claudeapi/client.go:L158)
- generateRequest() (internal/uiadapter/backend/claudeapi/client.go:L174)
- TestClaudeAPI_PromptCacheMarkers() (internal/uiadapter/backend/claudeapi/client_test.go:L111)
- NewClient() (internal/uiadapter/backend/claudecli/client.go:L32)
- TestClaudeCLI_VersionCheck() (internal/uiadapter/backend/claudecli/client_test.go:L15)
- TestClaudeCLI_Capabilities() (internal/uiadapter/backend/claudecli/client_test.go:L56)
- NewBreakerSet() (internal/uiadapter/breaker.go:L49)
- TestBreaker_TripsAfterThree() (internal/uiadapter/breaker_test.go:L16)
- TestBreaker_PerBackendIsolation() (internal/uiadapter/breaker_test.go:L43)
- TestBreaker_StateOnIdleIsClosed() (internal/uiadapter/breaker_test.go:L65)
- TestBreaker_ConcurrentConstruction() (internal/uiadapter/breaker_test.go:L73)
- NewResponseCache() (internal/uiadapter/cache.go:L46)
- Key() (internal/uiadapter/cache.go:L80)
- cache_test.go (internal/uiadapter/cache_test.go:L1)
- TestCache_HitRate() (internal/uiadapter/cache_test.go:L119)
- TestCache_SingleflightPropagatesError() (internal/uiadapter/cache_test.go:L139)
- TestCache_KeyDeterministic() (internal/uiadapter/cache_test.go:L151)
- TestCache_HashHit() (internal/uiadapter/cache_test.go:L17)
- TestCache_SingleflightCoalesces() (internal/uiadapter/cache_test.go:L38)
- TestCache_HashKeyIncludesBackendAndModel() (internal/uiadapter/cache_test.go:L71)
- TestCache_DisabledWhenZeroCapacity() (internal/uiadapter/cache_test.go:L83)
- config.go (internal/uiadapter/config.go:L1)
- DefaultConfig() (internal/uiadapter/config.go:L10)
- mergeWithDefaults() (internal/uiadapter/config.go:L61)
- configFieldSet() (internal/uiadapter/config.go:L98)
- TestConfig_PrivacyPatternsUsable() (internal/uiadapter/config_test.go:L113)
- TestDefaultConfig_Bootable_AllBackends() (internal/uiadapter/config_test.go:L20)
- TestDefaultConfig_MergeOntoZeroValued() (internal/uiadapter/config_test.go:L75)
- TestConfig_HasEveryPlanField() (internal/uiadapter/config_test.go:L87)
- TestStory2_AC2_FreeFunctionsAcceptNilLogger() (internal/uiadapter/logging_plumbing_test.go:L185)
- TestStory2_AC2_AC1_ConstructorsAcceptNilLogger() (internal/uiadapter/logging_plumbing_test.go:L98)
- TestStory3_AC8_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story3_sanitize_test.go:L159)
- randomPayload() (internal/uiadapter/logging_story3_sanitize_test.go:L34)
- containsSlice() (internal/uiadapter/logging_story3_sanitize_test.go:L55)
- TestStory3_AC7_SanitizeDisciplineAcrossFiles() (internal/uiadapter/logging_story3_sanitize_test.go:L73)
- TestStory4_AC8_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story4_sanitize_test.go:L182)
- TestStory4_AC7_SanitizeDisciplineAcrossPipeline() (internal/uiadapter/logging_story4_sanitize_test.go:L90)
- TestStory1_AC5_ConfigDefaults() (internal/uiadapter/logging_test.go:L260)
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
- TestStory3_AC6_PrefixCacheBuildEvents() (internal/uiadapter/prefix_cache_test.go:L89)
- NewRepairer() (internal/uiadapter/repair.go:L109)
- TestRepair_PropagatesGenerateError() (internal/uiadapter/repair_test.go:L129)
- TestRepair_RecoveryRate() (internal/uiadapter/repair_test.go:L17)
- TestRepair_NeverExceedsBudget() (internal/uiadapter/repair_test.go:L50)
- TestRepair_FastPathNoWastedCalls() (internal/uiadapter/repair_test.go:L71)
- TestRepair_Gated_PerBackend() (internal/uiadapter/repair_test.go:L92)
- sampling.go (internal/uiadapter/sampling.go:L1)
- OllamaSamplingOptions() (internal/uiadapter/sampling.go:L25)
- ClaudeSamplingOptions() (internal/uiadapter/sampling.go:L52)
- TestOllamaSamplingOptions_DeterministicDefaults() (internal/uiadapter/sampling_test.go:L13)
- TestClaudeSamplingOptions_NoSeed() (internal/uiadapter/sampling_test.go:L24)
- TestSamplingOptions_ConfigurableThroughConfig() (internal/uiadapter/sampling_test.go:L33)
- scanBufferForLeak() (internal/uiadapter/translate_e2e_test.go:L309)
- TestStory6_AC3_TenMessageNoLeak() (internal/uiadapter/translate_e2e_test.go:L329)
- driveLeakProbes() (internal/uiadapter/translate_e2e_test.go:L376)
- drivePerPhaseHelpers() (internal/uiadapter/translate_e2e_test.go:L63)

# Depends on
- [BreakerSet](/modules/breakerset.md)
- [BuildRepairPrompt](/modules/buildrepairprompt.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [NewDefault](/modules/newdefault.md)
- [ProcessByID](/modules/processbyid.md)
- [ResponseCache](/modules/responsecache.md)
- [SanitizeCapture](/modules/sanitizecapture.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [BuildRepairPrompt](/modules/buildrepairprompt.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [NewContextGuard](/modules/newcontextguard.md)
- [NewDefault](/modules/newdefault.md)
- [NewFastPathClassifier](/modules/newfastpathclassifier.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [RunWithFallback](/modules/runwithfallback.md)
- [SanitizeCapture](/modules/sanitizecapture.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Features
- no feature plan names these files
