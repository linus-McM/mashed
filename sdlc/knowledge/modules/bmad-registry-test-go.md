---
type: Module
title: bmad/registry_test.go
description: "Graphify community 62: app_bmad.go, internal/bmad/executor_multifileloader_test.go, internal/bmad/registry.go, internal/bmad/registry_interactive_phase2_test.go, internal/bmad/registry_test.go, intern"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T10:00:58Z" }
stale_after: "2026-10-13T10:00:58Z"
source_commit: 54a45892a8f0a2af4b1dccb63269c628ab15b3cd
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: executor_multifileloader_test, resource: internal/bmad/executor_multifileloader_test.go, last_modified: "2026-04-14T19:25:01+10:00", digest: 32a16dc3fc140e25 }
  - { id: registry, resource: internal/bmad/registry.go, last_modified: "2026-04-21T09:23:33+10:00", digest: df9f16ce4aa6d2e3 }
  - { id: registry_interactive_phase2_test, resource: internal/bmad/registry_interactive_phase2_test.go, last_modified: "2026-04-28T11:16:27+10:00", digest: 145cf97aa6278c93 }
  - { id: registry_test, resource: internal/bmad/registry_test.go, last_modified: "2026-04-28T12:29:58+10:00", digest: c231191d71cc2f84 }
  - { id: skillgen, resource: internal/bmad/skillgen.go, last_modified: "2026-04-09T22:26:45+10:00", digest: 925f4004ca59e699 }
  - { id: skillgen_test, resource: internal/bmad/skillgen_test.go, last_modified: "2026-04-09T22:26:45+10:00", digest: 421c997fe09f72e2 }
  - { id: types, resource: internal/bmad/types.go, last_modified: "2026-04-28T09:29:49+10:00", digest: 05a3e7d01f5f08c2 }
---

# Files
- `app_bmad.go`
- `internal/bmad/executor_multifileloader_test.go`
- `internal/bmad/registry.go`
- `internal/bmad/registry_interactive_phase2_test.go`
- `internal/bmad/registry_test.go`
- `internal/bmad/skillgen.go`
- `internal/bmad/skillgen_test.go`
- `internal/bmad/types.go`

# Symbols
- .GetBmadProcessesByPhase() (app_bmad.go:L73)
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
- GenerateSkillFiles() (internal/bmad/skillgen.go:L67)
- skillgen_test.go (internal/bmad/skillgen_test.go:L1)
- TestGenerateSkillFiles_OutputPathsResolved() (internal/bmad/skillgen_test.go:L109)
- TestGenerateSkillFiles_NoInputsFallback() (internal/bmad/skillgen_test.go:L123)
- TestGenerateSkillFiles_CreatesAllDirectories() (internal/bmad/skillgen_test.go:L13)
- TestGenerateSkillFiles_UnmappedOutputFallback() (internal/bmad/skillgen_test.go:L137)
- TestGenerateSkillFiles_ErrorOnInvalidBaseDir() (internal/bmad/skillgen_test.go:L155)
- TestGenerateSkillFiles_ErrorOnReadOnlyDir() (internal/bmad/skillgen_test.go:L166)
- TestGenerateSkillFiles_MixedMappedAndUnmappedOutputs() (internal/bmad/skillgen_test.go:L179)
- TestGenerateSkillFiles_Idempotent() (internal/bmad/skillgen_test.go:L196)
- TestGenerateSkillFiles_DirectoryNamesMatchSkillNames() (internal/bmad/skillgen_test.go:L32)
- TestGenerateSkillFiles_FrontmatterCorrect() (internal/bmad/skillgen_test.go:L58)
- TestGenerateSkillFiles_InputPathsResolved() (internal/bmad/skillgen_test.go:L95)
- BmadPhase (internal/bmad/types.go:L57)

# Depends on
- [ProcessByID](/modules/processbyid.md)
- [ProcessDef](/modules/processdef.md)
- [skillgen.go](/modules/skillgen-go.md)

# Inferred
- [ProcessByID](/modules/processbyid.md)

# Features
- no feature plan names these files
