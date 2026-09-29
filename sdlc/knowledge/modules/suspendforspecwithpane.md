---
type: Module
title: .suspendForSpecWithPane
description: "Graphify community 27: internal/bmad/executor.go, internal/bmad/executor_interactive_test.go, internal/bmad/executor_respond_test.go, internal/bmad/executor_suspend_test.go, internal/bmad/prompts.go,"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_respond_test, resource: internal/bmad/executor_respond_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 23fdebc82a94468b }
  - { id: executor_suspend_test, resource: internal/bmad/executor_suspend_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: b5f7b6a982bf834b }
  - { id: prompts, resource: internal/bmad/prompts.go, last_modified: "2026-04-27T13:46:23+10:00", digest: b76c227c938b8250 }
  - { id: registry_interactive_test, resource: internal/bmad/registry_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 2394d3c33e43a481 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_respond_test.go`
- `internal/bmad/executor_suspend_test.go`
- `internal/bmad/prompts.go`
- `internal/bmad/registry_interactive_test.go`

# Symbols
- registryLookup() (internal/bmad/executor.go:L2787)
- loadRegistryCSV() (internal/bmad/executor.go:L2870)
- TestRegistryLookupRejectsNonRegistryScheme() (internal/bmad/executor_interactive_test.go:L660)
- TestRegistryLookupRejectsSchemes() (internal/bmad/executor_respond_test.go:L472)
- TestUpsertAndRemovePromptHelpers() (internal/bmad/executor_suspend_test.go:L298)
- TestFindPendingPrompt() (internal/bmad/executor_suspend_test.go:L336)
- TestFindInputSpec() (internal/bmad/executor_suspend_test.go:L359)
- .waiter() (internal/bmad/prompts.go:L103)
- execState (internal/bmad/prompts.go:L103)
- .releaseWaiter() (internal/bmad/prompts.go:L120)
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
- .watchPaneForActivity() (internal/bmad/prompts.go:L503)
- .dismissAwaiting() (internal/bmad/prompts.go:L543)
- .RespondToInput() (internal/bmad/prompts.go:L564)
- TestOptionsRefResolution() (internal/bmad/registry_interactive_test.go:L138)

# Depends on
- [bmad/types.go](/modules/bmad-types-go.md)
- [.executeInteractiveNode](/modules/executeinteractivenode.md)
- [executor_fileloader_test.go](/modules/executor-fileloader-test-go.md)
- [prompts_test.go](/modules/prompts-test-go.md)

# Inferred
- [.executeInteractiveNode](/modules/executeinteractivenode.md)
- [Executor](/modules/executor.md)
- [ProcessByID](/modules/processbyid.md)
- [validate_test.go](/modules/validate-test-go.md)

# Features
- no feature plan names these files
