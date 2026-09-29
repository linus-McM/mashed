---
type: Module
title: context.Context
description: "Graphify community 91: app.go, internal/bmad/executor.go, internal/bmad/question.go, internal/uiadapter/adapter.go, internal/uiadapter/backend/backend.go, internal/uiadapter/backend/claudeapi/client.g"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: app, resource: app.go, last_modified: "2026-05-07T18:18:02+10:00", digest: 295875db4f0bdc1e }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: question, resource: internal/bmad/question.go, last_modified: "2026-04-17T20:58:29+10:00", digest: 81034d253e6ceb12 }
  - { id: adapter, resource: internal/uiadapter/adapter.go, last_modified: "2026-04-28T12:36:05+10:00", digest: aad905dbc8e94a13 }
  - { id: backend, resource: internal/uiadapter/backend/backend.go, last_modified: "2026-04-23T11:09:52+10:00", digest: d0eac5646b5c1ef3 }
  - { id: client, resource: internal/uiadapter/backend/claudeapi/client.go, last_modified: "2026-04-26T11:30:52+10:00", digest: 3e20643bf928e2f2 }
  - { id: client, resource: internal/uiadapter/backend/claudecli/client.go, last_modified: "2026-04-28T12:36:05+10:00", digest: 414128d39b8e1d31 }
  - { id: client_test, resource: internal/uiadapter/backend/claudecli/client_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 4342fc0ff1e40237 }
  - { id: lifecycle, resource: internal/uiadapter/backend/lifecycle.go, last_modified: "2026-04-23T11:34:52+10:00", digest: e57a307fa43599dd }
  - { id: lifecycle_test, resource: internal/uiadapter/backend/lifecycle_test.go, last_modified: "2026-04-23T11:34:52+10:00", digest: 7f78d92ee9e83203 }
  - { id: stubs, resource: internal/uiadapter/backend/stubs.go, last_modified: "2026-04-23T11:09:52+10:00", digest: 60bb478e1146d7a5 }
  - { id: client, resource: internal/uiadapter/client.go, last_modified: "2026-04-26T10:14:41+10:00", digest: 07a2280966b51366 }
  - { id: scorecard_v3, resource: internal/uiadapter/eval/scorecard_v3.go, last_modified: "2026-04-23T11:36:37+10:00", digest: 43f9103bcabd6620 }
  - { id: semaphore, resource: internal/uiadapter/semaphore.go, last_modified: "2026-04-26T10:14:41+10:00", digest: a0783d029000dab4 }
---

# Files
- `app.go`
- `internal/bmad/executor.go`
- `internal/bmad/question.go`
- `internal/uiadapter/adapter.go`
- `internal/uiadapter/backend/backend.go`
- `internal/uiadapter/backend/claudeapi/client.go`
- `internal/uiadapter/backend/claudecli/client.go`
- `internal/uiadapter/backend/claudecli/client_test.go`
- `internal/uiadapter/backend/lifecycle.go`
- `internal/uiadapter/backend/lifecycle_test.go`
- `internal/uiadapter/backend/stubs.go`
- `internal/uiadapter/client.go`
- `internal/uiadapter/eval/scorecard_v3.go`
- `internal/uiadapter/semaphore.go`

# Symbols
- .shutdown() (app.go:L374)
- DefaultCommandRunner() (internal/bmad/executor.go:L34)
- .captureQuestionOutput() (internal/bmad/question.go:L294)
- Executor (internal/bmad/question.go:L294)
- classifyChatErr() (internal/uiadapter/adapter.go:L336)
- Capabilities (internal/uiadapter/backend/backend.go:L45)
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
- TestClaudeCLI_ParsingHandlesShapes() (internal/uiadapter/backend/claudecli/client_test.go:L27)
- .WarmUpAll() (internal/uiadapter/backend/lifecycle.go:L42)
- slowBackend (internal/uiadapter/backend/lifecycle_test.go:L16)
- .WarmUp() (internal/uiadapter/backend/lifecycle_test.go:L22)
- StubBackend (internal/uiadapter/backend/stubs.go:L14)
- .Name() (internal/uiadapter/backend/stubs.go:L32)
- .Classify() (internal/uiadapter/backend/stubs.go:L34)
- .WarmUp() (internal/uiadapter/backend/stubs.go:L52)
- .Health() (internal/uiadapter/backend/stubs.go:L57)
- .Capabilities() (internal/uiadapter/backend/stubs.go:L62)
- .Calls() (internal/uiadapter/backend/stubs.go:L66)
- .ListModels() (internal/uiadapter/client.go:L176)
- classifyTransportErr() (internal/uiadapter/client.go:L207)
- Client (internal/uiadapter/client.go:L45)
- .Chat() (internal/uiadapter/client.go:L87)
- .ChatDeterministic() (internal/uiadapter/client.go:L95)
- .chat() (internal/uiadapter/client.go:L99)
- ShadowSampler (internal/uiadapter/eval/scorecard_v3.go:L262)
- .Counts() (internal/uiadapter/eval/scorecard_v3.go:L296)
- semaphore (internal/uiadapter/semaphore.go:L17)
- .acquire() (internal/uiadapter/semaphore.go:L36)
- .logWaitIfPositive() (internal/uiadapter/semaphore.go:L59)
- .logAcquired() (internal/uiadapter/semaphore.go:L71)
- .logCancelled() (internal/uiadapter/semaphore.go:L83)
- .release() (internal/uiadapter/semaphore.go:L94)

# Depends on
- [Config](/modules/config.md)
- [mashed/internal/uiadapter.UIAST](/modules/mashed-internal-uiadapter-uiast.md)
- [scorecard_v3_test.go](/modules/scorecard-v3-test-go.md)
- [stages_test.go](/modules/stages-test-go.md)
- [uiadapter/client_test.go](/modules/uiadapter-client-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
