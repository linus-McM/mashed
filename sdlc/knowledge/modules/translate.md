---
type: Module
title: .Translate
description: "Graphify community 450: internal/uiadapter/adapter.go, internal/uiadapter/logging.go"
resource: internal/uiadapter
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: logging, resource: internal/uiadapter/logging.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f7d1f46bdb4ea215 }
---

# Files
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/logging.go`

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
- .Handle() (internal/uiadapter/logging.go:L111)
- .WithAttrs() (internal/uiadapter/logging.go:L125)
- .WithGroup() (internal/uiadapter/logging.go:L134)
- fanoutHandler (internal/uiadapter/logging.go:L94)
- .Enabled() (internal/uiadapter/logging.go:L99)

# Depends on
- [context.Context](/modules/context-context.md)

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)
- [prompt_test.go](/modules/prompt-test-go.md)
- [SanitizeCapture](/modules/sanitizecapture.md)
- [testLogBuffer](/modules/testlogbuffer.md)

# Features
- no feature plan names these files
