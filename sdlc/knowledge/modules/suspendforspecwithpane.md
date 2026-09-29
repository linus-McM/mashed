---
type: Module
title: .suspendForSpecWithPane
description: "Graphify community 27: internal/bmad/events.go, internal/bmad/executor_suspend_test.go, internal/bmad/prompts.go, internal/bmad/prompts_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:14Z" }
stale_after: "2026-10-13T12:23:14Z"
source_commit: ab1f2eb4ec65fc2b1fce15503c11f0e310758404
sources:
  - { id: events, resource: internal/bmad/events.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 38b90b66f8c11c26 }
  - { id: executor_suspend_test, resource: internal/bmad/executor_suspend_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: b5f7b6a982bf834b }
  - { id: prompts, resource: internal/bmad/prompts.go, last_modified: "2026-04-27T13:46:23+10:00", digest: b76c227c938b8250 }
  - { id: prompts_test, resource: internal/bmad/prompts_test.go, last_modified: "2026-04-27T13:46:23+10:00", digest: 6a4efc33b34610a9 }
---

# Files
- `internal/bmad/events.go`
- `internal/bmad/executor_suspend_test.go`
- `internal/bmad/prompts.go`
- `internal/bmad/prompts_test.go`

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
- TestHashPendingPromptDeterminism() (internal/bmad/executor_suspend_test.go:L276)
- TestUpsertAndRemovePromptHelpers() (internal/bmad/executor_suspend_test.go:L298)
- TestFindPendingPrompt() (internal/bmad/executor_suspend_test.go:L336)
- TestFindInputSpec() (internal/bmad/executor_suspend_test.go:L359)
- prompts.go (internal/bmad/prompts.go:L1)
- .waiter() (internal/bmad/prompts.go:L103)
- execState (internal/bmad/prompts.go:L103)
- .releaseWaiter() (internal/bmad/prompts.go:L120)
- hashPendingPrompt() (internal/bmad/prompts.go:L132)
- upsertPrompt() (internal/bmad/prompts.go:L138)
- removePrompt() (internal/bmad/prompts.go:L150)
- findPendingPrompt() (internal/bmad/prompts.go:L163)
- findInputSpec() (internal/bmad/prompts.go:L173)
- renderPrompt() (internal/bmad/prompts.go:L185)
- resolveOptions() (internal/bmad/prompts.go:L193)
- .persistSnapshot() (internal/bmad/prompts.go:L225)
- Executor (internal/bmad/prompts.go:L225)
- randHex8() (internal/bmad/prompts.go:L269)
- .translateForPrompt() (internal/bmad/prompts.go:L293)
- .suspendForSpec() (internal/bmad/prompts.go:L366)
- .suspendForSpecWithPane() (internal/bmad/prompts.go:L385)
- extractLastClaudeTurn() (internal/bmad/prompts.go:L50)
- .watchPaneForActivity() (internal/bmad/prompts.go:L503)
- .dismissAwaiting() (internal/bmad/prompts.go:L543)
- .RespondToInput() (internal/bmad/prompts.go:L564)
- TestExtractLastClaudeTurn_FallbackTailCrop() (internal/bmad/prompts_test.go:L105)
- TestExtractLastClaudeTurn_Empty() (internal/bmad/prompts_test.go:L117)
- TestExtractLastClaudeTurn_GlyphBoundary() (internal/bmad/prompts_test.go:L92)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [executor_fileloader_test.go](/modules/executor-fileloader-test-go.md)

# Inferred
- [Executor](/modules/executor.md)
- [ProcessByID](/modules/processbyid.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)
- [validate.go](/modules/validate-go.md)

# Features
- no feature plan names these files
