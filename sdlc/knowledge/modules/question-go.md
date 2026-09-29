---
type: Module
title: question.go
description: "Graphify community 267: internal/bmad/question.go, internal/explain/explain.go, internal/uiadapter/cache.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: question, resource: internal/bmad/question.go, last_modified: "2026-04-17T20:58:29+10:00", digest: 81034d253e6ceb12 }
  - { id: explain, resource: internal/explain/explain.go, last_modified: "2026-04-10T11:36:27+10:00", digest: c8a0f1b0e6aad414 }
  - { id: cache, resource: internal/uiadapter/cache.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 10b083d2f056ef95 }
---

# Files
- `internal/bmad/question.go`
- `internal/explain/explain.go`
- `internal/uiadapter/cache.go`

# Symbols
- question.go (internal/bmad/question.go:L1)
- QuestionEvent (internal/bmad/question.go:L55)
- IdleEvent (internal/bmad/question.go:L77)
- explain.go (internal/explain/explain.go:L1)
- .Explain() (internal/explain/explain.go:L29)
- cacheKey() (internal/explain/explain.go:L71)
- truncateHunk() (internal/explain/explain.go:L77)
- envWithoutAPIKey() (internal/explain/explain.go:L86)
- cache.go (internal/uiadapter/cache.go:L1)

# Depends on
- [App](/modules/app.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [Executor](/modules/executor.md)
- [question_test.go](/modules/question-test-go.md)
- [ResponseCache](/modules/responsecache.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
