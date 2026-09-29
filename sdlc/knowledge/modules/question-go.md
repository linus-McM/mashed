---
type: Module
title: question.go
description: "Graphify community 257: internal/bmad/question.go, internal/bmad/session_naming.go, internal/explain/explain.go, internal/uiadapter/cache.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: question, resource: internal/bmad/question.go, last_modified: "2026-09-29T07:07:25Z", digest: 81034d253e6ceb12 }
  - { id: session_naming, resource: internal/bmad/session_naming.go, last_modified: "2026-09-29T07:07:25Z", digest: acbdad6853f5eaed }
  - { id: explain, resource: internal/explain/explain.go, last_modified: "2026-09-29T07:07:25Z", digest: c8a0f1b0e6aad414 }
  - { id: cache, resource: internal/uiadapter/cache.go, last_modified: "2026-09-29T07:07:25Z", digest: 10b083d2f056ef95 }
---

# Files
- `internal/bmad/question.go`
- `internal/bmad/session_naming.go`
- `internal/explain/explain.go`
- `internal/uiadapter/cache.go`

# Symbols
- question.go (internal/bmad/question.go:L1)
- QuestionEvent (internal/bmad/question.go:L55)
- IdleEvent (internal/bmad/question.go:L77)
- session_naming.go (internal/bmad/session_naming.go:L1)
- explain.go (internal/explain/explain.go:L1)
- .Explain() (internal/explain/explain.go:L29)
- cacheKey() (internal/explain/explain.go:L71)
- truncateHunk() (internal/explain/explain.go:L77)
- envWithoutAPIKey() (internal/explain/explain.go:L86)
- cache.go (internal/uiadapter/cache.go:L1)

# Depends on
- [App](/modules/app.md)
- [context.Context](/modules/context-context.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [question_test.go](/modules/question-test-go.md)
- [ResponseCache](/modules/responsecache.md)
- [session_naming_test.go](/modules/session-naming-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
