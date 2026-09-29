---
type: Module
title: artifacts_test.go
description: "Graphify community 388: internal/bmad/artifacts.go, internal/bmad/artifacts_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T14:46:21Z" }
stale_after: "2026-10-13T14:46:21Z"
source_commit: 7c9b1d72863df713a8f6811089f72bc851d9337b
sources:
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 0fcf7b7f19ac0972 }
  - { id: artifacts_test, resource: internal/bmad/artifacts_test.go, last_modified: "2026-04-08T20:52:15+10:00", digest: d50e7198f73eb5fe }
---

# Files
- `internal/bmad/artifacts.go`
- `internal/bmad/artifacts_test.go`

# Symbols
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

# Depends on
- [ProcessByID](/modules/processbyid.md)

# Inferred
- [bmad/registry_test.go](/modules/bmad-registry-test-go.md)
- [ProcessByID](/modules/processbyid.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
