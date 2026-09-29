---
type: Module
title: time.Time
description: "Graphify community 70: internal/uiadapter/adapter.go, internal/uiadapter/prompt.go, internal/uiadapter/prompt_test.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: prompt, resource: internal/uiadapter/prompt.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 07d009f8f445be65 }
  - { id: prompt_test, resource: internal/uiadapter/prompt_test.go, last_modified: "2026-04-27T10:45:20+10:00", digest: c0e8d16fee76a549 }
---

# Files
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/prompt.go`
- `internal/uiadapter/prompt_test.go`

# Symbols
- defaultAdapter (internal/uiadapter/adapter.go:L106)
- disabledAdapter (internal/uiadapter/adapter.go:L118)
- .Translate() (internal/uiadapter/adapter.go:L156)
- .Translate() (internal/uiadapter/adapter.go:L164)
- .chat() (internal/uiadapter/adapter.go:L210)
- .stampSuccessMetadata() (internal/uiadapter/adapter.go:L220)
- .emitFallback() (internal/uiadapter/adapter.go:L239)
- .logTelemetry() (internal/uiadapter/adapter.go:L255)
- .logTelemetryWithSanitize() (internal/uiadapter/adapter.go:L263)
- extractJSONObject() (internal/uiadapter/adapter.go:L293)
- prompt.go (internal/uiadapter/prompt.go:L1)
- SystemPrompt() (internal/uiadapter/prompt.go:L13)
- PromptVersion() (internal/uiadapter/prompt.go:L17)
- TestSystemPrompt_ContainsAllSections() (internal/uiadapter/prompt_test.go:L88)

# Depends on
- [context.Context](/modules/context-context.md)

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)
- [Validate](/modules/validate.md)

# Features
- no feature plan names these files
