---
type: Module
title: assets_validate_test.go
description: "Graphify community 112: internal/bmad/assets_validate.go, internal/bmad/assets_validate_test.go"
resource: internal/bmad
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: assets_validate, resource: internal/bmad/assets_validate.go, last_modified: "2026-04-12T16:58:02+10:00", digest: 3f68a2fbf7ac70cf }
  - { id: assets_validate_test, resource: internal/bmad/assets_validate_test.go, last_modified: "2026-04-12T16:58:02+10:00", digest: 0a7f087220505ed4 }
---

# Files
- `internal/bmad/assets_validate.go`
- `internal/bmad/assets_validate_test.go`

# Symbols
- assets_validate.go (internal/bmad/assets_validate.go:L1)
- isValidCompletion() (internal/bmad/assets_validate.go:L105)
- ValidationIssue (internal/bmad/assets_validate.go:L11)
- pathExistsInRepo() (internal/bmad/assets_validate.go:L119)
- ValidateMashedAsset() (internal/bmad/assets_validate.go:L35)
- assets_validate_test.go (internal/bmad/assets_validate_test.go:L1)
- TestValidateMashedAsset_ExistingInputPath_NoIssue() (internal/bmad/assets_validate_test.go:L101)
- TestValidateMashedAsset_GlobInputPath() (internal/bmad/assets_validate_test.go:L120)
- TestValidateMashedAsset_AC3_EmptyRepoPathSkipsPathChecks() (internal/bmad/assets_validate_test.go:L139)
- TestValidateMashedAsset_AC1_MissingDescription() (internal/bmad/assets_validate_test.go:L14)
- TestValidateMashedAsset_InvalidCompletion() (internal/bmad/assets_validate_test.go:L156)
- TestValidateMashedAsset_ValidCompletions() (internal/bmad/assets_validate_test.go:L172)
- TestValidateMashedAsset_CommandChainableNone() (internal/bmad/assets_validate_test.go:L192)
- TestValidateMashedAsset_FullyValid() (internal/bmad/assets_validate_test.go:L207)
- TestValidateMashedAsset_SkillRoleSkipsCompletionCheck() (internal/bmad/assets_validate_test.go:L221)
- TestValidateMashedAsset_MultipleIssues() (internal/bmad/assets_validate_test.go:L238)
- TestValidateMashedAsset_LoadedAsset_AttachesIssues() (internal/bmad/assets_validate_test.go:L255)
- TestValidateMashedAsset_LoadedAsset_NoIssues() (internal/bmad/assets_validate_test.go:L270)
- findIssue() (internal/bmad/assets_validate_test.go:L285)
- TestValidateMashedAsset_DescriptionTooShort() (internal/bmad/assets_validate_test.go:L31)
- TestValidateMashedAsset_DescriptionTooLong() (internal/bmad/assets_validate_test.go:L47)
- TestValidateMashedAsset_AC2_NonexistentInputPath() (internal/bmad/assets_validate_test.go:L67)
- TestValidateMashedAsset_NonexistentOutputPath() (internal/bmad/assets_validate_test.go:L84)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [assets_test.go](/modules/assets-test-go.md)

# Features
- no feature plan names these files
