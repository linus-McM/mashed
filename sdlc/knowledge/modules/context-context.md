---
type: Module
title: context.Context
description: "Graphify community 91: internal/bmad/executor.go, internal/bmad/question.go, internal/uiadapter/adapter.go, internal/uiadapter/backend/claudecli/client.go, internal/uiadapter/client.go, internal/uiada"
resource: internal
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: question, resource: internal/bmad/question.go, last_modified: "2026-04-17T20:58:29+10:00", digest: 81034d253e6ceb12 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: client, resource: internal/uiadapter/client.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 07a2280966b51366 }
  - { id: semaphore, resource: internal/uiadapter/semaphore.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a0783d029000dab4 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/question.go`
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/client.go`
- `internal/uiadapter/semaphore.go`

# Symbols
- DefaultCommandRunner() (internal/bmad/executor.go:L34)
- .captureQuestionOutput() (internal/bmad/question.go:L294)
- Executor (internal/bmad/question.go:L294)
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
- classifyChatErr() (internal/uiadapter/adapter.go:L336)
- .WarmUp() (internal/uiadapter/backend/claudecli/client.go:L58)
- .Health() (internal/uiadapter/backend/claudecli/client.go:L68)
- .ListModels() (internal/uiadapter/client.go:L176)
- classifyTransportErr() (internal/uiadapter/client.go:L207)
- Client (internal/uiadapter/client.go:L45)
- .Chat() (internal/uiadapter/client.go:L87)
- .ChatDeterministic() (internal/uiadapter/client.go:L95)
- .chat() (internal/uiadapter/client.go:L99)
- semaphore (internal/uiadapter/semaphore.go:L17)
- .acquire() (internal/uiadapter/semaphore.go:L36)
- .logWaitIfPositive() (internal/uiadapter/semaphore.go:L59)
- .logAcquired() (internal/uiadapter/semaphore.go:L71)
- .logCancelled() (internal/uiadapter/semaphore.go:L83)
- .release() (internal/uiadapter/semaphore.go:L94)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [log/slog.Logger](/modules/log-slog-logger.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [prompt_test.go](/modules/prompt-test-go.md)
- [SanitizeCapture](/modules/sanitizecapture.md)

# Features
- no feature plan names these files
