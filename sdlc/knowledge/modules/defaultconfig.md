---
type: Module
title: DefaultConfig
description: "Graphify community 6: internal/uiadapter/accountant.go, internal/uiadapter/accountant_test.go, internal/uiadapter/allowlist.go, internal/uiadapter/allowlist_test.go, internal/uiadapter/backend/backend"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: accountant, resource: internal/uiadapter/accountant.go, last_modified: "2026-04-23T11:29:34+10:00", digest: ae4389c89ff37e93 }
  - { id: accountant_test, resource: internal/uiadapter/accountant_test.go, last_modified: "2026-04-23T11:29:34+10:00", digest: 60c344b0cdff2ce8 }
  - { id: allowlist, resource: internal/uiadapter/allowlist.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ec96ab6a3976170d }
  - { id: allowlist_test, resource: internal/uiadapter/allowlist_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 865e6c967ac886df }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client_test, resource: internal/uiadapter/backend/claudeapi/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e3e3d9c28e77effe }
  - { id: stub, resource: internal/uiadapter/backend/claudeapi/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 0c95368a8fcd18d3 }
  - { id: stub, resource: internal/uiadapter/backend/claudecli/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 9d9685929a699a14 }
  - { id: lifecycle, resource: internal/uiadapter/backend/lifecycle.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e57a307fa43599dd }
  - { id: lifecycle_test, resource: internal/uiadapter/backend/lifecycle_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 7f78d92ee9e83203 }
  - { id: stub, resource: internal/uiadapter/backend/ollama/stub.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 82a69a451dac2ce8 }
  - { id: registry, resource: internal/uiadapter/backend/registry.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 34796db160eeb018 }
  - { id: registry_test, resource: internal/uiadapter/backend/registry_test.go, last_modified: "2026-04-23T11:09:52+10:00", digest: aa5a14fdab76f8f6 }
  - { id: router_test, resource: internal/uiadapter/backend/router_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 0a276fac34b2f9ab }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 60bb478e1146d7a5 }
  - { id: breaker, resource: internal/uiadapter/breaker.go, last_modified: "2026-04-26T10:14:41+10:00", digest: ce7fb713c254dda5 }
  - { id: breaker_test, resource: internal/uiadapter/breaker_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: c08c33460a000e4b }
  - { id: cache, resource: internal/uiadapter/cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 10b083d2f056ef95 }
  - { id: cache_test, resource: internal/uiadapter/cache_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: e8cf326b1dcc1d7a }
  - { id: config, resource: internal/uiadapter/config.go, last_modified: "2026-04-26T09:22:14+10:00", digest: d9832db7180bc48b }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: logging_story3_sanitize_test, resource: internal/uiadapter/logging_story3_sanitize_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a4de64483e766cb6 }
  - { id: prefix_cache_test, resource: internal/uiadapter/prefix_cache_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 448153a81ce2894c }
  - { id: sampling_test, resource: internal/uiadapter/sampling_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 10d6cf4c605c03f7 }
---

# Files
- `internal/uiadapter/accountant.go`
- `internal/uiadapter/accountant_test.go`
- `internal/uiadapter/allowlist.go`
- `internal/uiadapter/allowlist_test.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudeapi/client_test.go`
- `internal/uiadapter/backend/claudeapi/stub.go`
- `internal/uiadapter/backend/claudecli/stub.go`
- `internal/uiadapter/backend/lifecycle.go`
- `internal/uiadapter/backend/lifecycle_test.go`
- `internal/uiadapter/backend/ollama/stub.go`
- `internal/uiadapter/backend/registry.go`
- `internal/uiadapter/backend/registry_test.go`
- `internal/uiadapter/backend/router_test.go`
- `internal/uiadapter/backend/stubs.go`
- `internal/uiadapter/breaker.go`
- `internal/uiadapter/breaker_test.go`
- `internal/uiadapter/cache.go`
- `internal/uiadapter/cache_test.go`
- `internal/uiadapter/config.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story3_sanitize_test.go`
- `internal/uiadapter/prefix_cache_test.go`
- `internal/uiadapter/sampling_test.go`

# Symbols
- NewAccountant() (internal/uiadapter/accountant.go:L84)
- accountant_test.go (internal/uiadapter/accountant_test.go:L1)
- TestAccountant_TripsBeforeHard429() (internal/uiadapter/accountant_test.go:L27)
- TestAccountant_USDBudgetSoftLimit() (internal/uiadapter/accountant_test.go:L45)
- TestAccountant_Snapshot() (internal/uiadapter/accountant_test.go:L59)
- TestAccountant_SlidingWindow() (internal/uiadapter/accountant_test.go:L72)
- CheckModelAllowlist() (internal/uiadapter/allowlist.go:L40)
- allowlist_test.go (internal/uiadapter/allowlist_test.go:L1)
- TestStory5_AC6_AllowlistNilLoggerStillSafe() (internal/uiadapter/allowlist_test.go:L110)
- TestAllowlist_WarnsOnUnvetted() (internal/uiadapter/allowlist_test.go:L18)
- TestAllowlist_AllowsKnownModels() (internal/uiadapter/allowlist_test.go:L32)
- TestAllowlist_OverrideFlag() (internal/uiadapter/allowlist_test.go:L44)
- TestAllowlist_WarnsOnUnvettedClaude() (internal/uiadapter/allowlist_test.go:L57)
- Kind (internal/uiadapter/backend/backend.go:L17)
- apiKey (internal/uiadapter/backend/claudeapi/client.go:L34)
- .String() (internal/uiadapter/backend/claudeapi/client.go:L36)
- NewClient() (internal/uiadapter/backend/claudeapi/client.go:L41)
- claudeapi/client_test.go (internal/uiadapter/backend/claudeapi/client_test.go:L1)
- TestClaudeAPI_PromptCacheMarkers() (internal/uiadapter/backend/claudeapi/client_test.go:L111)
- TestClaudeAPI_AnthropicVersionHeader() (internal/uiadapter/backend/claudeapi/client_test.go:L122)
- TestClaudeAPI_StopReasonNotToolUseErrors() (internal/uiadapter/backend/claudeapi/client_test.go:L142)
- TestClaudeAPI_ToolUseRoundtrip() (internal/uiadapter/backend/claudeapi/client_test.go:L22)
- TestClaudeAPI_RetryAfter429() (internal/uiadapter/backend/claudeapi/client_test.go:L53)
- TestClaudeAPI_KeyNotLogged() (internal/uiadapter/backend/claudeapi/client_test.go:L89)
- TestClaudeAPI_WarmUpMissingKeyErrors() (internal/uiadapter/backend/claudeapi/client_test.go:L99)
- claudeapi/stub.go (internal/uiadapter/backend/claudeapi/stub.go:L1)
- init() (internal/uiadapter/backend/claudeapi/stub.go:L13)
- claudecli/stub.go (internal/uiadapter/backend/claudecli/stub.go:L1)
- init() (internal/uiadapter/backend/claudecli/stub.go:L12)
- NewLifecycle() (internal/uiadapter/backend/lifecycle.go:L29)
- lifecycle_test.go (internal/uiadapter/backend/lifecycle_test.go:L1)
- TestLifecycle_WarmUpConcurrentSafe() (internal/uiadapter/backend/lifecycle_test.go:L102)
- TestLifecycle_WarmUpAllInParallel() (internal/uiadapter/backend/lifecycle_test.go:L30)
- TestLifecycle_WarmUpRecordsResults() (internal/uiadapter/backend/lifecycle_test.go:L49)
- TestLifecycle_TickerDisableable() (internal/uiadapter/backend/lifecycle_test.go:L63)
- TestLifecycle_HealthTickerFeedsRouter() (internal/uiadapter/backend/lifecycle_test.go:L80)
- ollama/stub.go (internal/uiadapter/backend/ollama/stub.go:L1)
- init() (internal/uiadapter/backend/ollama/stub.go:L13)
- backend/registry.go (internal/uiadapter/backend/registry.go:L1)
- Constructor (internal/uiadapter/backend/registry.go:L13)
- Register() (internal/uiadapter/backend/registry.go:L24)
- From() (internal/uiadapter/backend/registry.go:L41)
- Available() (internal/uiadapter/backend/registry.go:L53)
- reset() (internal/uiadapter/backend/registry.go:L67)
- backend/registry_test.go (internal/uiadapter/backend/registry_test.go:L1)
- TestBackend_SingleShotUnsupportedIsSentinel() (internal/uiadapter/backend/registry_test.go:L119)
- TestBackend_Available() (internal/uiadapter/backend/registry_test.go:L132)
- TestBackend_FromUnknownName() (internal/uiadapter/backend/registry_test.go:L22)
- TestBackend_FromKnownName() (internal/uiadapter/backend/registry_test.go:L37)
- TestBackend_InterfaceStressConcurrent() (internal/uiadapter/backend/registry_test.go:L53)
- TestRouter_DefaultPolicyFallsBackToLocalWhenClaudeAbsent() (internal/uiadapter/backend/router_test.go:L59)
- TestRouter_ConfigurablePrivacyPatterns() (internal/uiadapter/backend/router_test.go:L92)
- NewStub() (internal/uiadapter/backend/stubs.go:L28)
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
- DefaultConfig() (internal/uiadapter/config.go:L10)
- logging_plumbing_test.go (internal/uiadapter/logging_plumbing_test.go:L1)
- TestStory2_AC3_NewDefaultScopesWithGroup() (internal/uiadapter/logging_plumbing_test.go:L300)
- keysOf() (internal/uiadapter/logging_plumbing_test.go:L343)
- TestStory2_AC3_NewDefaultNilParentSafe() (internal/uiadapter/logging_plumbing_test.go:L353)
- TestStory2_AC5_NilSafeLogger() (internal/uiadapter/logging_plumbing_test.go:L40)
- TestStory2_AC2_AC1_ConstructorsAcceptNilLogger() (internal/uiadapter/logging_plumbing_test.go:L98)
- TestStory3_AC8_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story3_sanitize_test.go:L159)
- prefix_cache_test.go (internal/uiadapter/prefix_cache_test.go:L1)
- TestClaudeSystemBlock_HasCacheControl() (internal/uiadapter/prefix_cache_test.go:L14)
- TestClaudeSystemBlock_OffTTLSkipsMarker() (internal/uiadapter/prefix_cache_test.go:L26)
- TestPrompt_StaticPrefix_ByteStable() (internal/uiadapter/prefix_cache_test.go:L39)
- TestOllamaKeepAliveEncoded() (internal/uiadapter/prefix_cache_test.go:L59)
- TestClaudeSystemBlock_EmptyPrefixNil() (internal/uiadapter/prefix_cache_test.go:L73)
- sampling_test.go (internal/uiadapter/sampling_test.go:L1)
- TestOllamaSamplingOptions_DeterministicDefaults() (internal/uiadapter/sampling_test.go:L13)
- TestClaudeSamplingOptions_NoSeed() (internal/uiadapter/sampling_test.go:L24)
- TestSamplingOptions_ConfigurableThroughConfig() (internal/uiadapter/sampling_test.go:L33)

# Depends on
- [Accountant](/modules/accountant.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [ProcessByID](/modules/processbyid.md)
- [question.go](/modules/question-go.md)
- [ResponseCache](/modules/responsecache.md)
- [Router](/modules/router.md)
- [StubBackend](/modules/stubbackend.md)
- [testing.T](/modules/testing-t.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)
- [logging_comprehensive_test.go](/modules/logging-comprehensive-test-go.md)
- [NewRepairer](/modules/newrepairer.md)
- [Router](/modules/router.md)
- [testLogBuffer](/modules/testlogbuffer.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Features
- no feature plan names these files
