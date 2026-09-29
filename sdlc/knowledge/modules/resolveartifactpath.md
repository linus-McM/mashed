---
type: Module
title: ResolveArtifactPath
description: "Graphify community 422: internal/bmad/artifacts.go, internal/bmad/artifacts_test.go, internal/bmad/executor.go, internal/bmad/executor_interactive_test.go, internal/bmad/executor_outputpaths_test.go,"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:14Z" }
stale_after: "2026-10-13T12:23:14Z"
source_commit: ab1f2eb4ec65fc2b1fce15503c11f0e310758404
sources:
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 0fcf7b7f19ac0972 }
  - { id: artifacts_test, resource: internal/bmad/artifacts_test.go, last_modified: "2026-04-08T20:52:15+10:00", digest: d50e7198f73eb5fe }
  - { id: executor, resource: internal/bmad/executor.go, last_modified: "2026-04-27T10:45:20+10:00", digest: 244fdd7f469d5870 }
  - { id: executor_interactive_test, resource: internal/bmad/executor_interactive_test.go, last_modified: "2026-04-28T12:02:57+10:00", digest: 38021efc1dc00f72 }
  - { id: executor_outputpaths_test, resource: internal/bmad/executor_outputpaths_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: 8d577df9f438ee41 }
  - { id: executor_respond_test, resource: internal/bmad/executor_respond_test.go, last_modified: "2026-04-20T20:17:58+10:00", digest: 23fdebc82a94468b }
  - { id: registry_interactive_test, resource: internal/bmad/registry_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: 2394d3c33e43a481 }
  - { id: testutil_interactive_test, resource: internal/bmad/testutil_interactive_test.go, last_modified: "2026-04-20T15:36:58+10:00", digest: f165a17bd64570bc }
---

# Files
- `internal/bmad/artifacts.go`
- `internal/bmad/artifacts_test.go`
- `internal/bmad/executor.go`
- `internal/bmad/executor_interactive_test.go`
- `internal/bmad/executor_outputpaths_test.go`
- `internal/bmad/executor_respond_test.go`
- `internal/bmad/registry_interactive_test.go`
- `internal/bmad/testutil_interactive_test.go`

# Symbols
- ResolveArtifactPath() (internal/bmad/artifacts.go:L44)
- VerifyArtifacts() (internal/bmad/artifacts.go:L75)
- artifacts_test.go (internal/bmad/artifacts_test.go:L1)
- TestAC3_VerifyArtifacts_EdgeCases() (internal/bmad/artifacts_test.go:L110)
- TestAC3_VerifyArtifacts_NoBmadOutputDir() (internal/bmad/artifacts_test.go:L126)
- TestAC5_VerifyArtifacts_DirectoryArtifact() (internal/bmad/artifacts_test.go:L137)
- TestAC3_VerifyArtifacts_StatErrorNotNotExist() (internal/bmad/artifacts_test.go:L149)
- TestAC5_VerifyArtifacts_DirectoryMissing() (internal/bmad/artifacts_test.go:L168)
- TestAC2_ResolveArtifactPath() (internal/bmad/artifacts_test.go:L39)
- TestAC3_VerifyArtifacts_MixedFoundMissing() (internal/bmad/artifacts_test.go:L97)
- resolveOutputPaths() (internal/bmad/executor.go:L1660)
- resolvedInputs (internal/bmad/executor.go:L2358)
- .resolveInputs() (internal/bmad/executor.go:L2649)
- firstDirectPredecessor() (internal/bmad/executor.go:L2749)
- envValue() (internal/bmad/executor.go:L2773)
- registryLookup() (internal/bmad/executor.go:L2787)
- loadRegistryCSV() (internal/bmad/executor.go:L2870)
- TestRegistryLookupRejectsNonRegistryScheme() (internal/bmad/executor_interactive_test.go:L660)
- TestAC2_UnmappedArtifacts_SkippedFromOutputPaths() (internal/bmad/executor_outputpaths_test.go:L83)
- TestRegistryLookupRejectsSchemes() (internal/bmad/executor_respond_test.go:L472)
- TestOptionsRefResolution() (internal/bmad/registry_interactive_test.go:L138)
- .writeArtifact() (internal/bmad/testutil_interactive_test.go:L262)

# Depends on
- [appendUpstreamContext](/modules/appendupstreamcontext.md)
- [bmad/registry_test.go](/modules/bmad-registry-test-go.md)
- [Executor](/modules/executor.md)

# Inferred
- [ProcessByID](/modules/processbyid.md)
- [.suspendForSpecWithPane](/modules/suspendforspecwithpane.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
