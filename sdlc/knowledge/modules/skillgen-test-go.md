---
type: Module
title: skillgen_test.go
description: "Graphify community 397: internal/bmad/skillgen.go, internal/bmad/skillgen_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: skillgen, resource: internal/bmad/skillgen.go, last_modified: "2026-04-09T22:26:45+10:00", digest: 925f4004ca59e699 }
  - { id: skillgen_test, resource: internal/bmad/skillgen_test.go, last_modified: "2026-04-09T22:26:45+10:00", digest: 421c997fe09f72e2 }
---

# Files
- `internal/bmad/skillgen.go`
- `internal/bmad/skillgen_test.go`

# Symbols
- skillgen.go (internal/bmad/skillgen.go:L1)
- skillTemplateData (internal/bmad/skillgen.go:L11)
- artifactPathSpec (internal/bmad/skillgen.go:L21)
- GenerateSkillFiles() (internal/bmad/skillgen.go:L67)
- buildSkillData() (internal/bmad/skillgen.go:L97)
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

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [bmad/registry_test.go](/modules/bmad-registry-test-go.md)
- [ResolveArtifactPath](/modules/resolveartifactpath.md)

# Features
- no feature plan names these files
