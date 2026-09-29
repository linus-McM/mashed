---
type: Module
title: question.go
description: "Graphify community 327: internal/bmad/question.go, internal/explain/explain.go, internal/uiadapter/breaker.go, internal/uiadapter/cache.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: question, resource: internal/bmad/question.go, last_modified: "2026-04-17T20:58:29+10:00", digest: 81034d253e6ceb12 }
  - { id: explain, resource: internal/explain/explain.go, last_modified: "2026-04-10T11:36:27+10:00", digest: c8a0f1b0e6aad414 }
  - { id: breaker, resource: internal/uiadapter/breaker.go, last_modified: "2026-04-26T10:14:41+10:00", digest: ce7fb713c254dda5 }
  - { id: cache, resource: internal/uiadapter/cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 10b083d2f056ef95 }
---

# Files
- `internal/bmad/question.go`
- `internal/explain/explain.go`
- `internal/uiadapter/breaker.go`
- `internal/uiadapter/cache.go`

# Symbols
- question.go (internal/bmad/question.go:L1)
- QuestionEvent (internal/bmad/question.go:L55)
- IdleEvent (internal/bmad/question.go:L77)
- explain.go (internal/explain/explain.go:L1)
- Explainer (internal/explain/explain.go:L15)
- New() (internal/explain/explain.go:L21)
- .Explain() (internal/explain/explain.go:L29)
- cacheKey() (internal/explain/explain.go:L71)
- truncateHunk() (internal/explain/explain.go:L77)
- envWithoutAPIKey() (internal/explain/explain.go:L86)
- .StateOf() (internal/uiadapter/breaker.go:L118)
- .Do() (internal/uiadapter/breaker.go:L134)
- breakerStateName() (internal/uiadapter/breaker.go:L16)
- BreakerSet (internal/uiadapter/breaker.go:L38)
- .For() (internal/uiadapter/breaker.go:L62)
- cache.go (internal/uiadapter/cache.go:L1)
- .Store() (internal/uiadapter/cache.go:L111)
- .logCacheGet() (internal/uiadapter/cache.go:L127)
- .HitRate() (internal/uiadapter/cache.go:L144)
- .DoShared() (internal/uiadapter/cache.go:L162)
- hashKey() (internal/uiadapter/cache.go:L17)
- .Metrics() (internal/uiadapter/cache.go:L174)
- ResponseCache (internal/uiadapter/cache.go:L35)
- .onEvicted() (internal/uiadapter/cache.go:L66)
- .Lookup() (internal/uiadapter/cache.go:L94)

# Depends on
- [Config](/modules/config.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [Executor](/modules/executor.md)
- [question_test.go](/modules/question-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
