---
type: Module
title: context.Context
description: "Graphify community 91: app.go, app_uiadapter_claudecli.go, internal/bmad/executor.go, internal/bmad/question.go, internal/uiadapter/adapter.go, internal/uiadapter/backend/claudeapi/client.go, internal"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: app_uiadapter_claudecli, resource: app_uiadapter_claudecli.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 9c423a0a128af5b2 }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: question, resource: internal/bmad/question.go, last_modified: "2026-04-17T20:58:29+10:00", digest: 81034d253e6ceb12 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: client_test, resource: internal/uiadapter/backend/claudecli/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 4342fc0ff1e40237 }
  - { id: client, resource: internal/uiadapter/client.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 07a2280966b51366 }
  - { id: semaphore, resource: internal/uiadapter/semaphore.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a0783d029000dab4 }
---

# Files
- `app.go`
- `app_uiadapter_claudecli.go`
- `internal/bmad/executor.go`
- `internal/bmad/question.go`
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/claudecli/client_test.go`
- `internal/uiadapter/client.go`
- `internal/uiadapter/semaphore.go`

# Symbols
- .shutdown() (app.go:L374)
- truncReason() (app_uiadapter_claudecli.go:L105)
- splitHeadTail() (app_uiadapter_claudecli.go:L118)
- dumpRawIfRequested() (app_uiadapter_claudecli.go:L129)
- capModel() (app_uiadapter_claudecli.go:L150)
- claudeCLIAdapter (app_uiadapter_claudecli.go:L21)
- .Translate() (app_uiadapter_claudecli.go:L39)
- DefaultCommandRunner() (internal/bmad/executor.go:L34)
- .captureQuestionOutput() (internal/bmad/question.go:L294)
- Executor (internal/bmad/question.go:L294)
- classifyChatErr() (internal/uiadapter/adapter.go:L336)
- .WarmUp() (internal/uiadapter/backend/claudeapi/client.go:L80)
- .Health() (internal/uiadapter/backend/claudeapi/client.go:L89)
- .GenerateSingleShot() (internal/uiadapter/backend/claudecli/client.go:L116)
- .TranslateWithFullPrompt() (internal/uiadapter/backend/claudecli/client.go:L132)
- .runRaw() (internal/uiadapter/backend/claudecli/client.go:L152)
- Client (internal/uiadapter/backend/claudecli/client.go:L21)
- .runOneShot() (internal/uiadapter/backend/claudecli/client.go:L234)
- extractFencedJSON() (internal/uiadapter/backend/claudecli/client.go:L326)
- .Name() (internal/uiadapter/backend/claudecli/client.go:L41)
- .WarmUp() (internal/uiadapter/backend/claudecli/client.go:L58)
- .Health() (internal/uiadapter/backend/claudecli/client.go:L68)
- .Classify() (internal/uiadapter/backend/claudecli/client.go:L72)
- .Generate() (internal/uiadapter/backend/claudecli/client.go:L93)
- TestClaudeCLI_ParsingHandlesShapes() (internal/uiadapter/backend/claudecli/client_test.go:L27)
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
- [claudeapi/client.go](/modules/claudeapi-client-go.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [StubBackend](/modules/stubbackend.md)
- [time.Time](/modules/time-time.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
