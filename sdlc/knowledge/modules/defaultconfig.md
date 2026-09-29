---
type: Module
title: DefaultConfig
description: "Graphify community 14: internal/uiadapter/backend/registry.go, internal/uiadapter/backend/registry_test.go, internal/uiadapter/breaker.go, internal/uiadapter/breaker_test.go, internal/uiadapter/cache."
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: registry, resource: internal/uiadapter/backend/registry.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 34796db160eeb018 }
  - { id: registry_test, resource: internal/uiadapter/backend/registry_test.go, last_modified: "2026-04-23T11:09:52+10:00", digest: aa5a14fdab76f8f6 }
  - { id: breaker, resource: internal/uiadapter/breaker.go, last_modified: "2026-04-26T10:14:41+10:00", digest: ce7fb713c254dda5 }
  - { id: breaker_test, resource: internal/uiadapter/breaker_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: c08c33460a000e4b }
  - { id: cache, resource: internal/uiadapter/cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 10b083d2f056ef95 }
  - { id: cache_test, resource: internal/uiadapter/cache_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: e8cf326b1dcc1d7a }
  - { id: config, resource: internal/uiadapter/config.go, last_modified: "2026-04-26T09:22:14+10:00", digest: d9832db7180bc48b }
  - { id: config_test, resource: internal/uiadapter/config_test.go, last_modified: "2026-04-23T11:43:31+10:00", digest: 748e88156e161a65 }
  - { id: contextguard, resource: internal/uiadapter/contextguard.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 417a5e0974dc194b }
  - { id: contextguard_test, resource: internal/uiadapter/contextguard_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 0cf07afee9915c31 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: logging_story3_sanitize_test, resource: internal/uiadapter/logging_story3_sanitize_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a4de64483e766cb6 }
  - { id: logging_story5_sanitize_test, resource: internal/uiadapter/logging_story5_sanitize_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 49f3cb55ded4bd06 }
  - { id: logging_test, resource: internal/uiadapter/logging_test.go, last_modified: "2026-04-26T09:22:14+10:00", digest: 744e5721d44f353b }
  - { id: translate_e2e_test, resource: internal/uiadapter/translate_e2e_test.go, last_modified: "2026-04-26T12:39:41+10:00", digest: 45f6254d2ea3d0bc }
---

# Files
- `internal/uiadapter/backend/registry.go`
- `internal/uiadapter/backend/registry_test.go`
- `internal/uiadapter/breaker.go`
- `internal/uiadapter/breaker_test.go`
- `internal/uiadapter/cache.go`
- `internal/uiadapter/cache_test.go`
- `internal/uiadapter/config.go`
- `internal/uiadapter/config_test.go`
- `internal/uiadapter/contextguard.go`
- `internal/uiadapter/contextguard_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story3_sanitize_test.go`
- `internal/uiadapter/logging_story5_sanitize_test.go`
- `internal/uiadapter/logging_test.go`
- `internal/uiadapter/translate_e2e_test.go`

# Symbols
- From() (internal/uiadapter/backend/registry.go:L41)
- TestBackend_SingleShotUnsupportedIsSentinel() (internal/uiadapter/backend/registry_test.go:L119)
- TestBackend_FromUnknownName() (internal/uiadapter/backend/registry_test.go:L22)
- TestBackend_FromKnownName() (internal/uiadapter/backend/registry_test.go:L37)
- TestBackend_InterfaceStressConcurrent() (internal/uiadapter/backend/registry_test.go:L53)
- NewBreakerSet() (internal/uiadapter/breaker.go:L49)
- breaker_test.go (internal/uiadapter/breaker_test.go:L1)
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
- TestConfig_PrivacyPatternsUsable() (internal/uiadapter/config_test.go:L113)
- TestDefaultConfig_Bootable_AllBackends() (internal/uiadapter/config_test.go:L20)
- TestDefaultConfig_MergeOntoZeroValued() (internal/uiadapter/config_test.go:L75)
- NewContextGuard() (internal/uiadapter/contextguard.go:L41)
- contextguard_test.go (internal/uiadapter/contextguard_test.go:L1)
- TestStory5_AC3_ApplyOllamaUnderBudget() (internal/uiadapter/contextguard_test.go:L135)
- TestContextGuard_TruncatesLongCapture_Ollama() (internal/uiadapter/contextguard_test.go:L15)
- TestStory5_AC3_ApplyClaudeTruncated() (internal/uiadapter/contextguard_test.go:L163)
- TestContextGuard_NoTruncationUnderBudget() (internal/uiadapter/contextguard_test.go:L30)
- TestContextGuard_OllamaOptions() (internal/uiadapter/contextguard_test.go:L42)
- TestContextGuard_RefusesLongCapture_Claude() (internal/uiadapter/contextguard_test.go:L53)
- TestContextGuard_AcceptsShortClaude() (internal/uiadapter/contextguard_test.go:L68)
- TestContextGuard_ZeroConfigDefaults() (internal/uiadapter/contextguard_test.go:L80)
- TestStory5_AC3_ApplyOllamaTruncated() (internal/uiadapter/contextguard_test.go:L94)
- TestStory2_AC2_AC1_ConstructorsAcceptNilLogger() (internal/uiadapter/logging_plumbing_test.go:L98)
- TestStory3_AC8_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story3_sanitize_test.go:L159)
- TestStory5_AC9_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story5_sanitize_test.go:L173)
- TestStory1_AC5_ConfigDefaults() (internal/uiadapter/logging_test.go:L260)
- driveLeakProbes() (internal/uiadapter/translate_e2e_test.go:L376)
- drivePerPhaseHelpers() (internal/uiadapter/translate_e2e_test.go:L63)

# Depends on
- [backend/registry.go](/modules/backend-registry-go.md)
- [BreakerSet](/modules/breakerset.md)
- [Config](/modules/config.md)
- [fastpath.go](/modules/fastpath-go.md)
- [ResponseCache](/modules/responsecache.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [NewFastPathClassifier](/modules/newfastpathclassifier.md)
- [NewRepairer](/modules/newrepairer.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [SanitizeCapture](/modules/sanitizecapture.md)
- [Spotlight](/modules/spotlight.md)
- [testLogBuffer](/modules/testlogbuffer.md)
- [TestStory2_AC2_FreeFunctionsAcceptNilLogger](/modules/teststory2-ac2-freefunctionsacceptnillogger.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Features
- no feature plan names these files
