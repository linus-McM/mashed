---
type: Module
title: interactiveHarness
description: "Graphify community 182: internal/bmad/testutil_interactive_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: testutil_interactive_test, resource: internal/bmad/testutil_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: f165a17bd64570bc }
---

# Files
- `internal/bmad/testutil_interactive_test.go`

# Symbols
- .makeRunner() (internal/bmad/testutil_interactive_test.go:L112)
- .respond() (internal/bmad/testutil_interactive_test.go:L140)
- .expectAwaitingInput() (internal/bmad/testutil_interactive_test.go:L148)
- .waitForRoundComplete() (internal/bmad/testutil_interactive_test.go:L170)
- .countRoundComplete() (internal/bmad/testutil_interactive_test.go:L184)
- .assertGateSatisfied() (internal/bmad/testutil_interactive_test.go:L203)
- .matchGateSatisfied() (internal/bmad/testutil_interactive_test.go:L216)
- .assertNodeComplete() (internal/bmad/testutil_interactive_test.go:L244)
- .assertNodeFailed() (internal/bmad/testutil_interactive_test.go:L253)
- .abortedEventsFor() (internal/bmad/testutil_interactive_test.go:L271)
- saveInteractiveWorkflowWithOverrides() (internal/bmad/testutil_interactive_test.go:L28)
- interactiveHarness (internal/bmad/testutil_interactive_test.go:L56)
- .startSingleNode() (internal/bmad/testutil_interactive_test.go:L83)
- .startSingleNodeWithOverrides() (internal/bmad/testutil_interactive_test.go:L92)

# Depends on
- [executor_command_test.go](/modules/executor-command-test-go.md)
- [newHarness](/modules/newharness.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)

# Inferred
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)

# Features
- no feature plan names these files
