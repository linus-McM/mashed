---
type: Module
title: sync.Mutex
description: "Graphify community 182: internal/bmad/executor_session_test.go, internal/bmad/executor_test.go, internal/bmad/testutil_interactive_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:20Z" }
stale_after: "2026-10-13T11:35:20Z"
source_commit: 4ff58d4a9c9200fdb96164858cc43b268e21e005
sources:
  - { id: executor_session_test, resource: internal/bmad/executor_session_test.go, last_modified: "2026-04-12T15:24:29+10:00", digest: e8e25bf75a4a8c21 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
  - { id: testutil_interactive_test, resource: internal/bmad/testutil_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: f165a17bd64570bc }
---

# Files
- `internal/bmad/executor_session_test.go`
- `internal/bmad/executor_test.go`
- `internal/bmad/testutil_interactive_test.go`

# Symbols
- capturedArgv (internal/bmad/executor_session_test.go:L18)
- .record() (internal/bmad/executor_session_test.go:L23)
- .last() (internal/bmad/executor_session_test.go:L31)
- TestSpawnCommandSession_RegressionFromRefactor() (internal/bmad/executor_session_test.go:L81)
- eventRecord (internal/bmad/executor_test.go:L22)
- testHarness (internal/bmad/executor_test.go:L27)
- .getEvents() (internal/bmad/executor_test.go:L49)
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
- [NewExecutor](/modules/newexecutor.md)
- [newHarness](/modules/newharness.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)
- [Storage](/modules/storage.md)

# Inferred
- [executor_iteration_test.go](/modules/executor-iteration-test-go.md)
- [newHarness](/modules/newharness.md)

# Features
- no feature plan names these files
