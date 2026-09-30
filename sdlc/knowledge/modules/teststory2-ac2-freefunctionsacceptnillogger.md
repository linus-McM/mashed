---
type: Module
title: TestStory2_AC2_FreeFunctionsAcceptNilLogger
description: "Graphify community 112: internal/uiadapter/logging_plumbing_test.go, internal/uiadapter/prefix_cache.go, internal/uiadapter/prefix_cache_test.go, internal/uiadapter/sampling.go, internal/uiadapter/sam"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: prefix_cache, resource: internal/uiadapter/prefix_cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 8963dac6e4ff7cab }
  - { id: prefix_cache_test, resource: internal/uiadapter/prefix_cache_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 448153a81ce2894c }
  - { id: sampling, resource: internal/uiadapter/sampling.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 372d2a2aa9064491 }
  - { id: sampling_test, resource: internal/uiadapter/sampling_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 10d6cf4c605c03f7 }
---

# Files
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/prefix_cache.go`
- `internal/uiadapter/prefix_cache_test.go`
- `internal/uiadapter/sampling.go`
- `internal/uiadapter/sampling_test.go`

# Symbols
- TestStory2_AC2_FreeFunctionsAcceptNilLogger() (internal/uiadapter/logging_plumbing_test.go:L185)
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

# Depends on
- [Config](/modules/config.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [NewRepairer](/modules/newrepairer.md)
- [nilSafeLogger](/modules/nilsafelogger.md)
- [SanitizeCapture](/modules/sanitizecapture.md)
- [Spotlight](/modules/spotlight.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Features
- no feature plan names these files
