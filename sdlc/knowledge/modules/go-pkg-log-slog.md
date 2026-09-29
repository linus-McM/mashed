---
type: Module
title: go_pkg_log_slog
description: "Graphify community 20: internal/uiadapter/allowlist.go, internal/uiadapter/allowlist_test.go, internal/uiadapter/fallback.go, internal/uiadapter/fallback_test.go, internal/uiadapter/fallback_tiers.go,"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:14Z" }
stale_after: "2026-10-13T12:23:14Z"
source_commit: ab1f2eb4ec65fc2b1fce15503c11f0e310758404
sources:
  - { id: allowlist, resource: internal/uiadapter/allowlist.go, last_modified: "2026-04-26T11:30:52+10:00", digest: ec96ab6a3976170d }
  - { id: allowlist_test, resource: internal/uiadapter/allowlist_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 865e6c967ac886df }
  - { id: fallback, resource: internal/uiadapter/fallback.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 187709266229767c }
  - { id: fallback_test, resource: internal/uiadapter/fallback_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 2fa70a931cc96a6b }
  - { id: fallback_tiers, resource: internal/uiadapter/fallback_tiers.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 88e0c867731816cd }
  - { id: fallback_tiers_test, resource: internal/uiadapter/fallback_tiers_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 73bb2c32262274bd }
  - { id: logging, resource: internal/uiadapter/logging.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f7d1f46bdb4ea215 }
  - { id: logging_plumbing_test, resource: internal/uiadapter/logging_plumbing_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 688be06cdc267d83 }
  - { id: prefix_cache, resource: internal/uiadapter/prefix_cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 8963dac6e4ff7cab }
  - { id: prefix_cache_test, resource: internal/uiadapter/prefix_cache_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 448153a81ce2894c }
  - { id: semaphore, resource: internal/uiadapter/semaphore.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a0783d029000dab4 }
  - { id: spotlight, resource: internal/uiadapter/spotlight.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 82f02eccfdc52095 }
  - { id: spotlight_test, resource: internal/uiadapter/spotlight_test.go, last_modified: "2026-04-26T11:30:52+10:00", digest: f9f81a13aa4a0344 }
---

# Files
- `internal/uiadapter/allowlist.go`
- `internal/uiadapter/allowlist_test.go`
- `internal/uiadapter/fallback.go`
- `internal/uiadapter/fallback_test.go`
- `internal/uiadapter/fallback_tiers.go`
- `internal/uiadapter/fallback_tiers_test.go`
- `internal/uiadapter/logging.go`
- `internal/uiadapter/logging_plumbing_test.go`
- `internal/uiadapter/prefix_cache.go`
- `internal/uiadapter/prefix_cache_test.go`
- `internal/uiadapter/semaphore.go`
- `internal/uiadapter/spotlight.go`
- `internal/uiadapter/spotlight_test.go`

# Symbols
- allowlist.go (internal/uiadapter/allowlist.go:L1)
- CheckModelAllowlist() (internal/uiadapter/allowlist.go:L40)
- joinAllowlist() (internal/uiadapter/allowlist.go:L78)
- allowlist_test.go (internal/uiadapter/allowlist_test.go:L1)
- TestStory5_AC6_AllowlistNilLoggerStillSafe() (internal/uiadapter/allowlist_test.go:L110)
- TestAllowlist_WarnsOnUnvetted() (internal/uiadapter/allowlist_test.go:L18)
- TestAllowlist_AllowsKnownModels() (internal/uiadapter/allowlist_test.go:L32)
- TestAllowlist_OverrideFlag() (internal/uiadapter/allowlist_test.go:L44)
- TestAllowlist_WarnsOnUnvettedClaude() (internal/uiadapter/allowlist_test.go:L57)
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
- TestStory2_AC5_NilSafeLogger() (internal/uiadapter/logging_plumbing_test.go:L40)
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

# Depends on
- [Config](/modules/config.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [testing.T](/modules/testing-t.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [repair.go](/modules/repair-go.md)
- [stages_test.go](/modules/stages-test-go.md)

# Features
- no feature plan names these files
