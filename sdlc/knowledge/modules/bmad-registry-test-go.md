---
type: Module
title: bmad/registry_test.go
description: "Graphify community 62: internal/bmad/artifacts.go, internal/bmad/artifacts_test.go, internal/bmad/executor_multifileloader_test.go, internal/bmad/registry.go, internal/bmad/registry_interactive_phase2"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:16:15Z" }
stale_after: "2026-10-13T11:16:15Z"
source_commit: 412dea92db2fe0033e1ef1b18ae99e119ea06b8c
sources:
  - { id: artifacts, resource: internal/bmad/artifacts.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 0fcf7b7f19ac0972 }
  - { id: artifacts_test, resource: internal/bmad/artifacts_test.go, last_modified: "2026-04-08T20:52:15+10:00", digest: d50e7198f73eb5fe }
  - { id: executor_multifileloader_test, resource: internal/bmad/executor_multifileloader_test.go, last_modified: "2026-04-14T19:25:01+10:00", digest: 32a16dc3fc140e25 }
  - { id: registry, resource: internal/bmad/registry.go, last_modified: "2026-04-21T09:23:33+10:00", digest: df9f16ce4aa6d2e3 }
  - { id: registry_interactive_phase2_test, resource: internal/bmad/registry_interactive_phase2_test.go, last_modified: "2026-04-28T11:16:27+10:00", digest: 145cf97aa6278c93 }
  - { id: registry_test, resource: internal/bmad/registry_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: c231191d71cc2f84 }
---

# Files
- `internal/bmad/artifacts.go`
- `internal/bmad/artifacts_test.go`
- `internal/bmad/executor_multifileloader_test.go`
- `internal/bmad/registry.go`
- `internal/bmad/registry_interactive_phase2_test.go`
- `internal/bmad/registry_test.go`

# Symbols
- VerifyArtifacts() (internal/bmad/artifacts.go:L75)
- artifacts_test.go (internal/bmad/artifacts_test.go:L1)
- TestAC3_VerifyArtifacts_EdgeCases() (internal/bmad/artifacts_test.go:L110)
- TestAC3_VerifyArtifacts_NoBmadOutputDir() (internal/bmad/artifacts_test.go:L126)
- TestAC5_VerifyArtifacts_DirectoryArtifact() (internal/bmad/artifacts_test.go:L137)
- TestAC3_VerifyArtifacts_StatErrorNotNotExist() (internal/bmad/artifacts_test.go:L149)
- TestAC5_VerifyArtifacts_DirectoryMissing() (internal/bmad/artifacts_test.go:L168)
- TestAC1_RegistryCompleteness() (internal/bmad/artifacts_test.go:L18)
- TestAC3_VerifyArtifacts_MixedFoundMissing() (internal/bmad/artifacts_test.go:L97)
- TestMultiFileLoader_AC5_RegistrySurface() (internal/bmad/executor_multifileloader_test.go:L233)
- bmad/registry.go (internal/bmad/registry.go:L1)
- init() (internal/bmad/registry.go:L13)
- AllProcesses() (internal/bmad/registry.go:L492)
- ProcessesByPhase() (internal/bmad/registry.go:L499)
- ProcessesByModule() (internal/bmad/registry.go:L525)
- TestStoryRollout02_BDD_U0ByteEqualityHoldsForNonRolledOutProcesses() (internal/bmad/registry_interactive_phase2_test.go:L436)
- bmad/registry_test.go (internal/bmad/registry_test.go:L1)
- TestProcessesByModule_Unknown() (internal/bmad/registry_test.go:L103)
- TestAllProcesses_ReturnsCopy() (internal/bmad/registry_test.go:L108)
- TestProcessDef_JSONRoundTrip() (internal/bmad/registry_test.go:L115)
- TestWorkflowDef_JSONRoundTrip() (internal/bmad/registry_test.go:L137)
- TestWorkflowDef_OmitsEmptyTemplateID() (internal/bmad/registry_test.go:L185)
- TestAllProcesses_CountAndFields() (internal/bmad/registry_test.go:L19)
- TestAC1_RegistryExpansion_ProcessCount() (internal/bmad/registry_test.go:L201)
- TestAC1_RegistryExpansion_NewProcesses() (internal/bmad/registry_test.go:L206)
- TestAC5_ProcessesByPhase_NewImplementation() (internal/bmad/registry_test.go:L290)
- TestAC5_ProcessesByModule_CIS() (internal/bmad/registry_test.go:L308)
- TestAC5_ProcessesByModule_BMGD() (internal/bmad/registry_test.go:L318)
- TestWorkflowNode_JSONRoundTrip() (internal/bmad/registry_test.go:L328)
- TestAllProcesses_UniqueIDs() (internal/bmad/registry_test.go:L39)
- TestU0_AC2_MigratedProcesses_ShapeJSON() (internal/bmad/registry_test.go:L450)
- TestU0_AC3_NonMigratedProcessesUnchanged() (internal/bmad/registry_test.go:L474)
- TestProcessesByPhase() (internal/bmad/registry_test.go:L50)
- TestU0_AC5_MigratedProcessesEnableAstAdapter_Q6Resolved() (internal/bmad/registry_test.go:L504)
- TestProcessesByPhase_UnionEqualsAll() (internal/bmad/registry_test.go:L73)
- TestProcessByID_Existing() (internal/bmad/registry_test.go:L82)
- TestProcessByID_Missing() (internal/bmad/registry_test.go:L90)
- TestProcessesByModule_Core() (internal/bmad/registry_test.go:L95)

# Depends on
- [ProcessByID](/modules/processbyid.md)
- [ProcessDef](/modules/processdef.md)

# Inferred
- [ProcessByID](/modules/processbyid.md)

# Features
- no feature plan names these files
