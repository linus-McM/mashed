---
type: Module
title: context.Context
description: "Graphify community 58: app_git.go, internal/bmad/executor.go, internal/bmad/executor_adapter_test.go, internal/bmad/question.go, internal/uiadapter/adapter.go, internal/uiadapter/backend/backend.go, i"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_git, resource: app_git.go, last_modified: "2026-09-30T06:49:47+10:00", digest: 9681e22dfb0b07f4 }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_adapter_test, resource: internal/bmad/executor_adapter_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: f79102c9f400d28f }
  - { id: question, resource: internal/bmad/question.go, last_modified: "2026-04-17T20:58:29+10:00", digest: 81034d253e6ceb12 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: client_test, resource: internal/uiadapter/backend/claudecli/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 4342fc0ff1e40237 }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 60bb478e1146d7a5 }
  - { id: client, resource: internal/uiadapter/client.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 07a2280966b51366 }
  - { id: scorecard_test, resource: internal/uiadapter/eval/scorecard_test.go, last_modified: "2026-04-22T14:13:03+10:00", digest: c10507f5ef6c8f7b }
  - { id: semaphore, resource: internal/uiadapter/semaphore.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a0783d029000dab4 }
---

# Files
- `app_git.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_adapter_test.go`
- `internal/bmad/question.go`
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/claudecli/client_test.go`
- `internal/uiadapter/backend/stubs.go`
- `internal/uiadapter/client.go`
- `internal/uiadapter/eval/scorecard_test.go`
- `internal/uiadapter/semaphore.go`

# Symbols
- latestOpenPR() (app_git.go:L1052)
- DefaultCommandRunner() (internal/bmad/executor.go:L34)
- delayAdapter (internal/bmad/executor_adapter_test.go:L34)
- .Translate() (internal/bmad/executor_adapter_test.go:L39)
- .captureQuestionOutput() (internal/bmad/question.go:L294)
- Executor (internal/bmad/question.go:L294)
- classifyChatErr() (internal/uiadapter/adapter.go:L336)
- Capabilities (internal/uiadapter/backend/backend.go:L45)
- .GenerateSingleShot() (internal/uiadapter/backend/claudeapi/client.go:L139)
- Client (internal/uiadapter/backend/claudeapi/client.go:L23)
- .Name() (internal/uiadapter/backend/claudeapi/client.go:L62)
- .Capabilities() (internal/uiadapter/backend/claudeapi/client.go:L65)
- .WarmUp() (internal/uiadapter/backend/claudeapi/client.go:L80)
- .Health() (internal/uiadapter/backend/claudeapi/client.go:L89)
- .GenerateSingleShot() (internal/uiadapter/backend/claudecli/client.go:L116)
- .TranslateWithFullPrompt() (internal/uiadapter/backend/claudecli/client.go:L132)
- .runRaw() (internal/uiadapter/backend/claudecli/client.go:L152)
- Client (internal/uiadapter/backend/claudecli/client.go:L21)
- .runOneShot() (internal/uiadapter/backend/claudecli/client.go:L234)
- extractFencedJSON() (internal/uiadapter/backend/claudecli/client.go:L326)
- .Name() (internal/uiadapter/backend/claudecli/client.go:L41)
- .Capabilities() (internal/uiadapter/backend/claudecli/client.go:L44)
- .WarmUp() (internal/uiadapter/backend/claudecli/client.go:L58)
- .Health() (internal/uiadapter/backend/claudecli/client.go:L68)
- .Classify() (internal/uiadapter/backend/claudecli/client.go:L72)
- .Generate() (internal/uiadapter/backend/claudecli/client.go:L93)
- TestClaudeCLI_ParsingHandlesShapes() (internal/uiadapter/backend/claudecli/client_test.go:L27)
- StubBackend (internal/uiadapter/backend/stubs.go:L14)
- .Name() (internal/uiadapter/backend/stubs.go:L32)
- .Classify() (internal/uiadapter/backend/stubs.go:L34)
- .Generate() (internal/uiadapter/backend/stubs.go:L39)
- .GenerateSingleShot() (internal/uiadapter/backend/stubs.go:L44)
- .WarmUp() (internal/uiadapter/backend/stubs.go:L52)
- .Health() (internal/uiadapter/backend/stubs.go:L57)
- .Capabilities() (internal/uiadapter/backend/stubs.go:L62)
- .Calls() (internal/uiadapter/backend/stubs.go:L66)
- synthUIAST() (internal/uiadapter/backend/stubs.go:L77)
- .ListModels() (internal/uiadapter/client.go:L176)
- classifyTransportErr() (internal/uiadapter/client.go:L207)
- Client (internal/uiadapter/client.go:L45)
- .Chat() (internal/uiadapter/client.go:L87)
- .ChatDeterministic() (internal/uiadapter/client.go:L95)
- .chat() (internal/uiadapter/client.go:L99)
- sequentialMockAdapter (internal/uiadapter/eval/scorecard_test.go:L19)
- .Translate() (internal/uiadapter/eval/scorecard_test.go:L24)
- semaphore (internal/uiadapter/semaphore.go:L17)
- .acquire() (internal/uiadapter/semaphore.go:L36)
- .logWaitIfPositive() (internal/uiadapter/semaphore.go:L59)
- .logAcquired() (internal/uiadapter/semaphore.go:L71)
- .logCancelled() (internal/uiadapter/semaphore.go:L83)
- .release() (internal/uiadapter/semaphore.go:L94)

# Depends on
- [Accountant](/modules/accountant.md)
- [claudeapi/client.go](/modules/claudeapi-client-go.md)
- [DefaultConfig](/modules/defaultconfig.md)
- [log/slog.Logger](/modules/log-slog-logger.md)
- [testLogBuffer](/modules/testlogbuffer.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
