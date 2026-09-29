---
type: Module
title: assets_test.go
description: "Graphify community 12: app_bmad.go, internal/bmad/assets.go, internal/bmad/assets_test.go, internal/bmad/assets_validate.go, internal/bmad/assets_validate_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T11:35:27Z" }
stale_after: "2026-10-13T11:35:27Z"
source_commit: a2484a9a3b200a7f406028196e6f86cc4cf9306f
sources:
  - { id: app_bmad, resource: app_bmad.go, last_modified: "2026-04-28T12:36:05+10:00", digest: fca7c61a439c150e }
  - { id: assets, resource: internal/bmad/assets.go, last_modified: "2026-04-12T16:58:02+10:00", digest: ae5984d95ad01e8c }
  - { id: assets_test, resource: internal/bmad/assets_test.go, last_modified: "2026-04-11T22:06:50+10:00", digest: 906af472f47c7866 }
  - { id: assets_validate, resource: internal/bmad/assets_validate.go, last_modified: "2026-04-12T16:58:02+10:00", digest: 3f68a2fbf7ac70cf }
  - { id: assets_validate_test, resource: internal/bmad/assets_validate_test.go, last_modified: "2026-04-12T16:58:02+10:00", digest: 0a7f087220505ed4 }
---

# Files
- `app_bmad.go`
- `internal/bmad/assets.go`
- `internal/bmad/assets_test.go`
- `internal/bmad/assets_validate.go`
- `internal/bmad/assets_validate_test.go`

# Symbols
- .ListAllMashedAssets() (app_bmad.go:L379)
- validateMashedGroup() (app_bmad.go:L423)
- MashedAssetRole (internal/bmad/assets.go:L107)
- GroupedMashedAssets (internal/bmad/assets.go:L119)
- parseMashedAsset() (internal/bmad/assets.go:L241)
- normaliseRole() (internal/bmad/assets.go:L299)
- applyRoleDefaults() (internal/bmad/assets.go:L314)
- firstBodyLine() (internal/bmad/assets.go:L346)
- nonNilStrings() (internal/bmad/assets.go:L361)
- LoadMashedAssetsFromDir() (internal/bmad/assets.go:L393)
- loadOneEntry() (internal/bmad/assets.go:L426)
- MashedAssetKind (internal/bmad/assets.go:L44)
- logMashedLoadWarning() (internal/bmad/assets.go:L458)
- MashedAssetSource (internal/bmad/assets.go:L55)
- MashedAssetInfo (internal/bmad/assets.go:L70)
- assets_test.go (internal/bmad/assets_test.go:L1)
- TestNormaliseRole() (internal/bmad/assets_test.go:L109)
- TestApplyRoleDefaults_Command() (internal/bmad/assets_test.go:L139)
- TestApplyRoleDefaults_Skill() (internal/bmad/assets_test.go:L147)
- TestApplyRoleDefaults_AuthorOverridesPreserved() (internal/bmad/assets_test.go:L154)
- TestFirstBodyLine() (internal/bmad/assets_test.go:L169)
- writeAsset() (internal/bmad/assets_test.go:L197)
- TestParseMashedAsset_MashedReady() (internal/bmad/assets_test.go:L205)
- TestParseMashedAsset_MissingMashedRole_SkipsSilently() (internal/bmad/assets_test.go:L235)
- TestParseMashedAsset_NoFrontmatter_SkipsSilently() (internal/bmad/assets_test.go:L249)
- TestParseMashedAsset_UnknownRole_ReturnsError() (internal/bmad/assets_test.go:L257)
- TestParseMashedAsset_MalformedYAML_ReturnsError() (internal/bmad/assets_test.go:L271)
- TestParseMashedAsset_DescriptionFallbackToFirstBodyLine() (internal/bmad/assets_test.go:L284)
- TestParseMashedAsset_SkillRoleDefaultsSessionPinned() (internal/bmad/assets_test.go:L301)
- TestLoadMashedAssetsFromDir_NonexistentDir_NotAnError() (internal/bmad/assets_test.go:L320)
- TestLoadMashedAssetsFromDir_EmptyDir() (internal/bmad/assets_test.go:L329)
- TestLoadMashedAssetsFromDir_EmptyDirString() (internal/bmad/assets_test.go:L336)
- TestLoadMashedAssetsFromDir_CommandLayout_MixedAssets() (internal/bmad/assets_test.go:L342)
- TestLoadMashedAssetsFromDir_SkillLayout_DirectoryPerSkill() (internal/bmad/assets_test.go:L371)
- TestLoadMashedAssetsFromDir_MalformedFile_SkippedNotFatal() (internal/bmad/assets_test.go:L396)
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
- [assets_write_test.go](/modules/assets-write-test-go.md)

# Inferred
- no INFERRED edges; treat any that appear as hints

# Features
- no feature plan names these files
