---
type: Module
title: ResolveArtifactPath
description: "Graphify community 82: app_bmad.go, internal/bmad/artifacts.go, internal/bmad/artifacts_test.go, internal/bmad/executor.go, internal/bmad/executor_outputpaths_test.go, internal/bmad/executor_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T20:55:29Z" }
stale_after: "2026-10-13T20:55:29Z"
source_commit: c5568e08ac3c9ebbe5442e8f9b0ce4b6fe270ffd
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 0fcf7b7f19ac0972 }
  - { id: artifacts_test, resource: internal/bmad/artifacts_test.go, last_modified: "2026-04-08T20:52:15+10:00", digest: d50e7198f73eb5fe }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
---

# Files
- `app_bmad.go`
- `internal/bmad/artifacts.go`
- `internal/bmad/artifacts_test.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/executor_test.go`

# Symbols
- .GetArtifactStatus() (app_bmad.go:L463)
- ResolveArtifactPath() (internal/bmad/artifacts.go:L44)
- GetArtifactStatus() (internal/bmad/artifacts.go:L59)
- TestAC2_ResolveArtifactPath() (internal/bmad/artifacts_test.go:L39)
- resolveOutputPaths() (internal/bmad/executor.go:L1660)
- resolvedInputs (internal/bmad/executor.go:L2358)
- .resolveInputs() (internal/bmad/executor.go:L2649)
- firstDirectPredecessor() (internal/bmad/executor.go:L2749)
- envValue() (internal/bmad/executor.go:L2773)
- TestAC2_UnmappedArtifacts_SkippedFromOutputPaths() (internal/bmad/executor_outputpaths_test.go:L83)
- TestGetArtifactStatus_Exists() (internal/bmad/executor_test.go:L2506)
- TestGetArtifactStatus_Missing() (internal/bmad/executor_test.go:L2517)
- TestGetArtifactStatus_UnmappedArtifact() (internal/bmad/executor_test.go:L2525)

# Depends on
- [appendUpstreamContext](/modules/appendupstreamcontext.md)
- [Executor](/modules/executor.md)
- [registryLookup](/modules/registrylookup.md)

# Inferred
- [ProcessByID](/modules/processbyid.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
