---
type: Module
title: nilSafeLogger
description: "Graphify community 50: internal/uiadapter/encode.go, internal/uiadapter/encode_test.go, internal/uiadapter/logging.go, internal/uiadapter/logging_comprehensive_test.go, internal/uiadapter/logging_plum"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: encode, resource: internal/uiadapter/encode.go, last_modified: "2026-09-29T07:07:25Z", digest: bf82e888f12311d6 }
  - { id: encode_test, resource: internal/uiadapter/encode_test.go, last_modified: "2026-09-29T07:07:25Z", digest: a0cd96e6b9f32bfb }
  - { id: logging, resource: internal/uiadapter/logging.go, last_modified: "2026-09-29T07:07:25Z", digest: f7d1f46bdb4ea215 }
  - { id: logging_comprehensive_test, resource: internal/uiadapter/logging_comprehensive_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 446816f0c9fa2be3 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 688be06cdc267d83 }
  - { id: logging_story5_sanitize_test, resource: internal/uiadapter/logging_story5_sanitize_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 49f3cb55ded4bd06 }
  - { id: prefix_cache, resource: internal/uiadapter/prefix_cache.go, last_modified: "2026-09-29T07:07:25Z", digest: 8963dac6e4ff7cab }
  - { id: prefix_cache_test, resource: internal/uiadapter/prefix_cache_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 448153a81ce2894c }
  - { id: sampling, resource: internal/uiadapter/sampling.go, last_modified: "2026-09-29T07:07:25Z", digest: 372d2a2aa9064491 }
  - { id: sampling_test, resource: internal/uiadapter/sampling_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 10d6cf4c605c03f7 }
  - { id: spotlight, resource: internal/uiadapter/spotlight.go, last_modified: "2026-09-29T07:07:25Z", digest: 82f02eccfdc52095 }
  - { id: spotlight_test, resource: internal/uiadapter/spotlight_test.go, last_modified: "2026-09-29T07:07:25Z", digest: f9f81a13aa4a0344 }
---

# Files
- `internal/uiadapter/encode.go`
- `internal/uiadapter/encode_test.go`
- `internal/uiadapter/logging.go`
- `internal/uiadapter/logging_comprehensive_test.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/logging_story5_sanitize_test.go`
- `internal/uiadapter/prefix_cache.go`
- `internal/uiadapter/prefix_cache_test.go`
- `internal/uiadapter/sampling.go`
- `internal/uiadapter/sampling_test.go`
- `internal/uiadapter/spotlight.go`
- `internal/uiadapter/spotlight_test.go`

# Symbols
- ClaudeToolName() (internal/uiadapter/encode.go:L93)
- TestClaudeToolName_StableShape() (internal/uiadapter/encode_test.go:L63)
- nilSafeLogger() (internal/uiadapter/logging.go:L169)
- TestStory6_AC1_NilSafeLoggerComprehensive() (internal/uiadapter/logging_comprehensive_test.go:L305)
- TestStory2_AC2_FreeFunctionsAcceptNilLogger() (internal/uiadapter/logging_plumbing_test.go:L185)
- TestStory2_AC5_NilSafeLogger() (internal/uiadapter/logging_plumbing_test.go:L40)
- TestStory5_AC9_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story5_sanitize_test.go:L173)
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
- sampling.go (internal/uiadapter/sampling.go:L1)
- OllamaSamplingOptions() (internal/uiadapter/sampling.go:L25)
- ClaudeSamplingOptions() (internal/uiadapter/sampling.go:L52)
- sampling_test.go (internal/uiadapter/sampling_test.go:L1)
- TestOllamaSamplingOptions_DeterministicDefaults() (internal/uiadapter/sampling_test.go:L13)
- TestClaudeSamplingOptions_NoSeed() (internal/uiadapter/sampling_test.go:L24)
- TestSamplingOptions_ConfigurableThroughConfig() (internal/uiadapter/sampling_test.go:L33)
- Unspotlight() (internal/uiadapter/spotlight.go:L100)
- Spotlight() (internal/uiadapter/spotlight.go:L42)
- spotlight_test.go (internal/uiadapter/spotlight_test.go:L1)
- TestSpotlight_ReplacesWhitespace() (internal/uiadapter/spotlight_test.go:L13)
- TestSpotlight_DisabledBypass() (internal/uiadapter/spotlight_test.go:L21)
- TestSpotlight_RoundTripLossless() (internal/uiadapter/spotlight_test.go:L29)
- TestSpotlight_InjectionCorpus() (internal/uiadapter/spotlight_test.go:L39)
- TestSpotlight_EmptyAndUnicode() (internal/uiadapter/spotlight_test.go:L65)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [stages_test.go](/modules/stages-test-go.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [FallbackAST](/modules/fallbackast.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [NewRepairer](/modules/newrepairer.md)
- [stages_test.go](/modules/stages-test-go.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Features
- no feature plan names these files
