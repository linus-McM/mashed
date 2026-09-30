---
type: Module
title: .executeInteractiveNode
description: "Graphify community 278: internal/bmad/events.go, internal/bmad/executor.go, internal/bmad/executor_anyuseranswer_test.go, internal/bmad/executor_suspend_test.go, internal/bmad/gate.go, internal/bmad/p"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: events, resource: internal/bmad/events.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 38b90b66f8c11c26 }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_anyuseranswer_test, resource: internal/bmad/executor_anyuseranswer_test.go, last_modified: "2026-04-26T09:47:45+10:00", digest: c5b7c0a594991958 }
  - { id: executor_suspend_test, resource: internal/bmad/executor_suspend_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: b5f7b6a982bf834b }
  - { id: gate, resource: internal/bmad/gate.go, last_modified: "2026-04-21T20:21:32+10:00", digest: faeb23bd0495c1c9 }
  - { id: prompts, resource: internal/bmad/prompts.go, last_modified: "2026-04-27T13:46:23+10:00", digest: b76c227c938b8250 }
---

# Files
- `internal/bmad/events.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_anyuseranswer_test.go`
- `internal/bmad/executor_suspend_test.go`
- `internal/bmad/gate.go`
- `internal/bmad/prompts.go`

# Symbols
- events.go (internal/bmad/events.go:L1)
- abortedPayload() (internal/bmad/events.go:L108)
- sessionDeadPayload() (internal/bmad/events.go:L117)
- roundCompletePayload() (internal/bmad/events.go:L28)
- gateSatisfiedPayload() (internal/bmad/events.go:L38)
- roundLimitPayload() (internal/bmad/events.go:L48)
- sha256hex() (internal/bmad/events.go:L61)
- awaitingPayload() (internal/bmad/events.go:L69)
- awaitingDismissedPayload() (internal/bmad/events.go:L75)
- inputResolvedPayload() (internal/bmad/events.go:L87)
- invalidPayload() (internal/bmad/events.go:L98)
- .executeInteractiveNode() (internal/bmad/executor.go:L2382)
- .sendToSession() (internal/bmad/executor.go:L2609)
- .verifyOutputs() (internal/bmad/executor.go:L3017)
- TestGate_AnyUserAnswerMatches_LastRoundWindow() (internal/bmad/executor_anyuseranswer_test.go:L44)
- TestHashPendingPromptDeterminism() (internal/bmad/executor_suspend_test.go:L276)
- gate.go (internal/bmad/gate.go:L1)
- collectSubAnswersForSpec() (internal/bmad/gate.go:L115)
- findNodeProcessID() (internal/bmad/gate.go:L132)
- astStructuredInUse() (internal/bmad/gate.go:L152)
- flattenSubAnswers() (internal/bmad/gate.go:L169)
- .checkGate() (internal/bmad/gate.go:L19)
- Executor (internal/bmad/gate.go:L19)
- containsToken() (internal/bmad/gate.go:L76)
- anyUserAnswerMatches() (internal/bmad/gate.go:L95)
- hashPendingPrompt() (internal/bmad/prompts.go:L132)

# Depends on
- [Executor](/modules/executor.md)
- [.resolveInputs](/modules/resolveinputs.md)

# Inferred
- [NewExecutor](/modules/newexecutor.md)
- [ProcessByID](/modules/processbyid.md)
- [question_test.go](/modules/question-test-go.md)

# Features
- no feature plan names these files
