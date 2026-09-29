---
type: Module
title: cache_test.go
description: "Graphify community 14: internal/uiadapter/cache.go, internal/uiadapter/cache_test.go, internal/uiadapter/logging_story3_sanitize_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: cache, resource: internal/uiadapter/cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 10b083d2f056ef95 }
  - { id: cache_test, resource: internal/uiadapter/cache_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: e8cf326b1dcc1d7a }
  - { id: logging_story3_sanitize_test, resource: internal/uiadapter/logging_story3_sanitize_test.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a4de64483e766cb6 }
---

# Files
- `internal/uiadapter/cache.go`
- `internal/uiadapter/cache_test.go`
- `internal/uiadapter/logging_story3_sanitize_test.go`

# Symbols
- cache.go (internal/uiadapter/cache.go:L1)
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
- TestStory3_AC8_HotPathZeroAllocs_DebugOff() (internal/uiadapter/logging_story3_sanitize_test.go:L159)

# Depends on
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [ResponseCache](/modules/responsecache.md)

# Inferred
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)

# Features
- no feature plan names these files
