---
type: Module
title: time.Time
description: "Graphify community 123: internal/agent/engine.go, internal/uiadapter/adapter.go"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: engine, resource: internal/agent/engine.go, last_modified: "2026-04-11T17:08:28+10:00", digest: 1042190db571e4e7 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
---

# Files
- `internal/agent/engine.go`
- `internal/uiadapter/adapter.go`

# Symbols
- agentState (internal/agent/engine.go:L41)
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

# Depends on
- [context.Context](/modules/context-context.md)
- [sessions.go](/modules/sessions-go.md)

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)
- [testLogBuffer](/modules/testlogbuffer.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Features
- no feature plan names these files
