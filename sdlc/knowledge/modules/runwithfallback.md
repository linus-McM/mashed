---
type: Module
title: RunWithFallback
description: "Graphify community 181: internal/uiadapter/fallback_tiers.go, internal/uiadapter/fallback_tiers_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: fallback_tiers, resource: internal/uiadapter/fallback_tiers.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 88e0c867731816cd }
  - { id: fallback_tiers_test, resource: internal/uiadapter/fallback_tiers_test.go, last_modified: "2026-04-26T10:43:55+10:00", digest: 73bb2c32262274bd }
---

# Files
- `internal/uiadapter/fallback_tiers.go`
- `internal/uiadapter/fallback_tiers_test.go`

# Symbols
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

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)

# Features
- no feature plan names these files
