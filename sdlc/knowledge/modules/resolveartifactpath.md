---
type: Module
title: ResolveArtifactPath
description: "Graphify community 422: internal/bmad/artifacts.go, internal/bmad/artifacts_test.go, internal/bmad/executor.go, internal/bmad/executor_outputpaths_test.go, internal/bmad/executor_test.go, internal/bma"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 0fcf7b7f19ac0972 }
  - { id: artifacts_test, resource: internal/bmad/artifacts_test.go, last_modified: "2026-04-08T20:52:15+10:00", digest: d50e7198f73eb5fe }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: executor_test, resource: internal/bmad/executor_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 4449efba985daebc }
  - { id: testutil_interactive_test, resource: internal/bmad/testutil_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: f165a17bd64570bc }
---

# Files
- `internal/bmad/artifacts.go`
- `internal/bmad/artifacts_test.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/executor_test.go`
- `internal/bmad/testutil_interactive_test.go`

# Symbols
- ResolveArtifactPath() (internal/bmad/artifacts.go:L44)
- GetArtifactStatus() (internal/bmad/artifacts.go:L59)
- VerifyArtifacts() (internal/bmad/artifacts.go:L75)
- artifacts_test.go (internal/bmad/artifacts_test.go:L1)
- TestAC3_VerifyArtifacts_EdgeCases() (internal/bmad/artifacts_test.go:L110)
- TestAC3_VerifyArtifacts_NoBmadOutputDir() (internal/bmad/artifacts_test.go:L126)
- TestAC5_VerifyArtifacts_DirectoryArtifact() (internal/bmad/artifacts_test.go:L137)
- TestAC3_VerifyArtifacts_StatErrorNotNotExist() (internal/bmad/artifacts_test.go:L149)
- TestAC5_VerifyArtifacts_DirectoryMissing() (internal/bmad/artifacts_test.go:L168)
- TestAC1_RegistryCompleteness() (internal/bmad/artifacts_test.go:L18)
- TestAC2_ResolveArtifactPath() (internal/bmad/artifacts_test.go:L39)
- TestAC3_VerifyArtifacts_MixedFoundMissing() (internal/bmad/artifacts_test.go:L97)
- resolveOutputPaths() (internal/bmad/executor.go:L1660)
- TestAC2_UnmappedArtifacts_SkippedFromOutputPaths() (internal/bmad/executor_outputpaths_test.go:L83)
- TestGetArtifactStatus_Exists() (internal/bmad/executor_test.go:L2506)
- TestGetArtifactStatus_Missing() (internal/bmad/executor_test.go:L2517)
- TestGetArtifactStatus_UnmappedArtifact() (internal/bmad/executor_test.go:L2525)
- .writeArtifact() (internal/bmad/testutil_interactive_test.go:L262)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [bmad/registry_test.go](/modules/bmad-registry-test-go.md)

# Features
- no feature plan names these files
