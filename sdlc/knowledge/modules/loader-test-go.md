---
type: Module
title: loader_test.go
description: "Graphify community 15: app_review_scoped.go, app_review_scoped_test.go, internal/advice/loader.go, internal/advice/loader_test.go"
resource: ""
tags: [module, graphify]
status: draft
generated: { by: sdlc/0.8.1, at: "2026-09-29T12:23:00Z" }
stale_after: "2026-10-13T12:23:00Z"
source_commit: c111e9108518da0d6e68dc3fa508835faaebfd96
sources:
  - { id: app_review_scoped, resource: app_review_scoped.go, last_modified: "2026-04-10T13:49:09+10:00", digest: 83c692114f39a0c9 }
  - { id: app_review_scoped_test, resource: app_review_scoped_test.go, last_modified: "2026-04-10T13:49:09+10:00", digest: 495a403ad8cace57 }
  - { id: loader, resource: internal/advice/loader.go, last_modified: "2026-04-12T20:21:26+10:00", digest: 8dbc36587dafa6a5 }
  - { id: loader_test, resource: internal/advice/loader_test.go, last_modified: "2026-04-10T10:10:32+10:00", digest: 99b099c62c602a3e }
---

# Files
- `app_review_scoped.go`
- `app_review_scoped_test.go`
- `internal/advice/loader.go`
- `internal/advice/loader_test.go`

# Symbols
- scopedAdviceEvent() (app_review_scoped.go:L108)
- App (app_review_scoped.go:L121)
- .StreamScopedAdvice() (app_review_scoped.go:L121)
- buildScopedDiff() (app_review_scoped.go:L24)
- containedPath() (app_review_scoped.go:L68)
- listTrackedFiles() (app_review_scoped.go:L82)
- assembleScopedPayload() (app_review_scoped.go:L99)
- app_review_scoped_test.go (app_review_scoped_test.go:L1)
- TestBuildScopedDiff_MixedTrackedUntracked() (app_review_scoped_test.go:L121)
- TestBuildScopedDiff_AC4_EmptyInput() (app_review_scoped_test.go:L147)
- TestBuildScopedDiff_AllFilesUnchanged() (app_review_scoped_test.go:L178)
- TestBuildScopedDiff_RejectsPathTraversal() (app_review_scoped_test.go:L197)
- initTestGitRepo() (app_review_scoped_test.go:L20)
- TestBuildScopedDiff_SeparatorBetweenFiles() (app_review_scoped_test.go:L235)
- TestAssembleScopedPayload_AC3_AdditionalContextPrepended() (app_review_scoped_test.go:L253)
- TestAssembleScopedPayload_EmptyContext() (app_review_scoped_test.go:L310)
- TestScopedAdviceEvent_AC5_MatchesStreamAdviceShape() (app_review_scoped_test.go:L331)
- gitRun() (app_review_scoped_test.go:L34)
- commitFile() (app_review_scoped_test.go:L43)
- modifyTrackedFile() (app_review_scoped_test.go:L52)
- createUntrackedFile() (app_review_scoped_test.go:L59)
- TestBuildScopedDiff_AC1_OnlySelectedFiles() (app_review_scoped_test.go:L69)
- TestBuildScopedDiff_AC2_UntrackedFallback() (app_review_scoped_test.go:L99)
- loader.go (internal/advice/loader.go:L1)
- loadModesFromDir() (internal/advice/loader.go:L103)
- LoadAdviceBody() (internal/advice/loader.go:L127)
- loadAdviceBodyFromDirs() (internal/advice/loader.go:L142)
- adviceFrontmatter (internal/advice/loader.go:L18)
- findModeInDir() (internal/advice/loader.go:L184)
- parseAdviceFile() (internal/advice/loader.go:L207)
- parseAdviceFileFromFS() (internal/advice/loader.go:L222)
- parseAdviceContent() (internal/advice/loader.go:L237)
- splitFrontmatter() (internal/advice/loader.go:L269)
- LoadAdviceModes() (internal/advice/loader.go:L28)
- globalAdviceDir() (internal/advice/loader.go:L304)
- localAdviceDir() (internal/advice/loader.go:L313)
- loadAdviceModesFromDirs() (internal/advice/loader.go:L44)
- loader_test.go (internal/advice/loader_test.go:L1)
- writeAdviceFile() (internal/advice/loader_test.go:L14)
- TestLoadAdviceModes_MergePriority() (internal/advice/loader_test.go:L167)
- TestLoadAdviceModes_SortOrderThenDisplayName() (internal/advice/loader_test.go:L234)
- TestParseAdviceFile() (internal/advice/loader_test.go:L24)
- TestLoadAdviceModes_MissingDirectories() (internal/advice/loader_test.go:L269)
- TestLoadAdviceModes_NonMdFilesIgnored() (internal/advice/loader_test.go:L276)
- TestLoadAdviceModes_MalformedFilesSkipped() (internal/advice/loader_test.go:L298)
- TestLoadAdviceModes_EmptyRepoPathSkipsLocal() (internal/advice/loader_test.go:L327)
- TestLoadAdviceBody() (internal/advice/loader_test.go:L335)
- TestParseAdviceFileFromFS() (internal/advice/loader_test.go:L380)
- TestParseAdviceFile_NonexistentFile() (internal/advice/loader_test.go:L396)
- TestLoadAdviceBody_GlobalOnly() (internal/advice/loader_test.go:L402)
- TestBundledDefaults_AllFilesValid() (internal/advice/loader_test.go:L419)
- TestBundledDefaults_BodyLoadable() (internal/advice/loader_test.go:L471)
- TestBundledDefaults_LocalOverridesBundled() (internal/advice/loader_test.go:L490)
- TestLoadAdviceBody_BundledFallback() (internal/advice/loader_test.go:L524)
- TestLoadAdviceModes_BundledWithEmbeddedFS() (internal/advice/loader_test.go:L531)
- TestLoadAdviceModes_AllThreeSources() (internal/advice/loader_test.go:L549)
- TestFindModeInDir_NotFound() (internal/advice/loader_test.go:L588)
- TestFindModeInDir_NonexistentDir() (internal/advice/loader_test.go:L603)
- TestSplitFrontmatter_EmptyBody() (internal/advice/loader_test.go:L608)
- TestLoadModesFromDir_NonexistentDir() (internal/advice/loader_test.go:L616)
- TestLoadAdviceBody_PublicAPI() (internal/advice/loader_test.go:L622)
- TestSplitFrontmatter_CRLFLineEndings() (internal/advice/loader_test.go:L629)
- TestSplitFrontmatter_OpeningCRLF() (internal/advice/loader_test.go:L637)
- TestGlobalAdviceDir() (internal/advice/loader_test.go:L645)
- TestLocalAdviceDir() (internal/advice/loader_test.go:L652)

# Depends on
- no EXTRACTED edges to other modules

# Inferred
- [App](/modules/app-63.md)

# Features
- [Repo health remediation](/features/repo-health-remediation.md)
