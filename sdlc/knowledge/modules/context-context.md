---
type: Module
title: context.Context
description: "Graphify community 58: app.go, internal/bmad/cleanup.go, internal/bmad/executor.go, internal/bmad/question.go, internal/uiadapter/adapter.go, internal/uiadapter/backend/claudeapi/client.go, internal/u"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T15:22:13Z" }
stale_after: "2026-10-13T15:22:13Z"
source_commit: 918616f42268500be0e26faa9f92720109b96761
sources:
  - { id: app, resource: app.go, last_modified: "2026-09-30T01:21:31+10:00", digest: 774e87e5f070ece0 }
  - { id: cleanup, resource: internal/bmad/cleanup.go, last_modified: "2026-05-07T09:52:03+10:00", digest: 78b71ea7a045636f }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: question, resource: internal/bmad/question.go, last_modified: "2026-04-17T20:58:29+10:00", digest: 81034d253e6ceb12 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: client, resource: internal/uiadapter/client.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 07a2280966b51366 }
  - { id: semaphore, resource: internal/uiadapter/semaphore.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a0783d029000dab4 }
---

# Files
- `app.go`
- `internal/bmad/cleanup.go`
- `internal/bmad/executor.go`
- `internal/bmad/question.go`
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/client.go`
- `internal/uiadapter/semaphore.go`

# Symbols
- .shutdown() (app.go:L432)
- .CleanupStaleSessions() (internal/bmad/cleanup.go:L21)
- Executor (internal/bmad/cleanup.go:L21)
- isNoTmuxServer() (internal/bmad/cleanup.go:L61)
- .liveSessionNames() (internal/bmad/cleanup.go:L69)
- DefaultCommandRunner() (internal/bmad/executor.go:L34)
- .captureQuestionOutput() (internal/bmad/question.go:L294)
- Executor (internal/bmad/question.go:L294)
- classifyChatErr() (internal/uiadapter/adapter.go:L336)
- .WarmUp() (internal/uiadapter/backend/claudeapi/client.go:L80)
- .Health() (internal/uiadapter/backend/claudeapi/client.go:L89)
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
- [resume_ghost_test.go](/modules/resume-ghost-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
