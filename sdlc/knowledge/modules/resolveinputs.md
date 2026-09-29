---
type: Module
title: .resolveInputs
description: "Graphify community 351: internal/bmad/executor.go, internal/bmad/executor_interactive_test.go, internal/bmad/executor_respond_test.go, internal/bmad/registry_interactive_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_respond_test, resource: internal/bmad/executor_respond_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 23fdebc82a94468b }
  - { id: registry_interactive_test, resource: internal/bmad/registry_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 2394d3c33e43a481 }
---

# Files
- `internal/bmad/executor.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_respond_test.go`
- `internal/bmad/registry_interactive_test.go`

# Symbols
- resolvedInputs (internal/bmad/executor.go:L2358)
- .resolveInputs() (internal/bmad/executor.go:L2649)
- firstDirectPredecessor() (internal/bmad/executor.go:L2749)
- envValue() (internal/bmad/executor.go:L2773)
- registryLookup() (internal/bmad/executor.go:L2787)
- loadRegistryCSV() (internal/bmad/executor.go:L2870)
- TestRegistryLookupRejectsNonRegistryScheme() (internal/bmad/executor_interactive_test.go:L660)
- TestRegistryLookupRejectsSchemes() (internal/bmad/executor_respond_test.go:L472)
- TestOptionsRefResolution() (internal/bmad/registry_interactive_test.go:L138)

# Depends on
- [Executor](/modules/executor.md)
- [executor_fileloader_test.go](/modules/executor-fileloader-test-go.md)

# Inferred
- [ProcessByID](/modules/processbyid.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Features
- no feature plan names these files
