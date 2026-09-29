---
type: Module
title: skillgen_test.go
description: "Graphify community 305: internal/bmad/skillgen.go, internal/bmad/skillgen_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T07:07:25Z" }
stale_after: "2026-10-13T07:07:25Z"
source_commit: ""
sources:
  - { id: skillgen, resource: internal/bmad/skillgen.go, last_modified: "2026-09-29T07:07:25Z", digest: 925f4004ca59e699 }
  - { id: skillgen_test, resource: internal/bmad/skillgen_test.go, last_modified: "2026-09-29T07:07:25Z", digest: 421c997fe09f72e2 }
---

# Files
- `internal/bmad/skillgen.go`
- `internal/bmad/skillgen_test.go`

# Symbols
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

# Depends on
- [go_pkg_os](/modules/go-pkg-os.md)

# Inferred
- [bmad/registry_test.go](/modules/bmad-registry-test-go.md)

# Features
- no feature plan names these files
